package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	mu       sync.Mutex
	fallback int
	db       *sql.DB
)

func dsnFromEnv() string {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://postgres:postgres@ping-db-svc:5432/pingdb?sslmode=disable"
	}
	return dsn
}

func initDB() {
	candidates := []string{dsnFromEnv()}

	var lastErr error
	for _, dsn := range candidates {
		tmp, err := sql.Open("postgres", dsn)
		if err != nil {
			lastErr = err
			continue
		}
		tmp.SetMaxOpenConns(5)
		// retry for up to ~30s, DB pod may still be starting
		var pingErr error
		for i := 0; i < 30; i++ {
			pingErr = tmp.Ping()
			if pingErr == nil {
				break
			}
			log.Printf("waiting for postgres (%s): %v", dsn, pingErr)
			time.Sleep(2 * time.Second)
		}
		if pingErr != nil {
			lastErr = pingErr
			tmp.Close()
			continue
		}
		// ensure table exists
		if _, err := tmp.Exec(`CREATE TABLE IF NOT EXISTS pings (id INT PRIMARY KEY, count INT NOT NULL)`); err != nil {
			lastErr = err
			tmp.Close()
			continue
		}
		if _, err := tmp.Exec(`INSERT INTO pings (id, count) VALUES (1, 0) ON CONFLICT (id) DO NOTHING`); err != nil {
			lastErr = err
			tmp.Close()
			continue
		}
		// load current count into fallback for seamless failover
		var n int
		if err := tmp.QueryRow(`SELECT count FROM pings WHERE id = 1`).Scan(&n); err == nil {
			mu.Lock()
			fallback = n
			mu.Unlock()
		}
		db = tmp
		log.Printf("connected to postgres at %s", dsn)
		return
	}
	log.Printf("postgres unavailable (%v), using in-memory counter", lastErr)
}

func incrementAndGet() int {
	if db != nil {
		var n int
		err := db.QueryRow(`UPDATE pings SET count = count + 1 WHERE id = 1 RETURNING count`).Scan(&n)
		if err == nil {
			mu.Lock()
			fallback = n
			mu.Unlock()
			return n
		}
		log.Printf("db increment failed: %v, falling back", err)
	}
	mu.Lock()
	defer mu.Unlock()
	fallback++
	return fallback
}

func currentCount() int {
	if db != nil {
		var n int
		if err := db.QueryRow(`SELECT count FROM pings WHERE id = 1`).Scan(&n); err == nil {
			mu.Lock()
			fallback = n
			mu.Unlock()
			return n
		}
	}
	mu.Lock()
	defer mu.Unlock()
	return fallback
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	initDB()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		n := incrementAndGet()
		fmt.Fprintf(w, "<h1>pong %d</h1>", n)
	})

	http.HandleFunc("/pings", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, currentCount())
	})

	log.Printf("Ping app server started in port %s (DATABASE_URL=%s)", addr, os.Getenv("DATABASE_URL"))
	log.Fatal(http.ListenAndServe(addr, nil))
}
