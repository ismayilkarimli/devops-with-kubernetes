package main

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var healthy = true

func liveHandler(w http.ResponseWriter, r *http.Request) {
	if !healthy {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func breakAppHandler(w http.ResponseWriter, r *http.Request) {
	healthy = false
	fmt.Fprintln(w, "Backend marked unhealthy. Its liveness probe will now fail.")
}

func dsnFromEnv() string {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://postgres:postgres@todo-app-db-svc:5432/tododb?sslmode=disable"
	}
	return dsn
}

func todoAppURL() string {
	url := os.Getenv("TODO_APP_URL")
	if url == "" {
		url = "http://todo-app-svc:2345"
	}
	return url
}

func initDB() {
	dsn := dsnFromEnv()
	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("sql.Open failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("waiting for postgres: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("postgres unavailable: %v", err)
	}
	db.SetMaxOpenConns(5)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		content VARCHAR(140) NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		log.Fatalf("create table failed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM todos`).Scan(&count); err != nil {
		log.Fatalf("count query failed: %v", err)
	}
	if count == 0 {
		seed := []struct {
			text string
			done bool
		}{
			{"Learn Kubernetes basics", true},
			{"Set up the persistent volume", false},
			{"Build the todo app", false},
		}
		for _, s := range seed {
			if _, err := db.Exec(`INSERT INTO todos (content, done) VALUES ($1, $2)`, s.text, s.done); err != nil {
				log.Fatalf("seed insert failed: %v", err)
			}
		}
		log.Printf("seeded %d todos", len(seed))
	}
	log.Printf("connected to postgres at %s", dsn)
}

func listTodos(w http.ResponseWriter) {
	rows, err := db.Query(`SELECT id, content, done FROM todos ORDER BY id`)
	if err != nil {
		log.Printf("query failed: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var body strings.Builder
	n := 0
	for rows.Next() {
		var id int
		var text string
		var done bool
		if err := rows.Scan(&id, &text, &done); err != nil {
			log.Printf("scan failed: %v", err)
			continue
		}
		n++
		escaped := html.EscapeString(text)
		if done {
			fmt.Fprintf(&body, `<li class="todo" data-id="%d"><span class="todo-text done">%s</span><strong class="todo-status">Done</strong></li>`, id, escaped)
		} else {
			fmt.Fprintf(&body, `<li class="todo" data-id="%d"><span class="todo-text">%s</span><input type="checkbox" class="todo-check" aria-label="Mark %s as done"></li>`, id, escaped, escaped)
		}
	}
	log.Printf("returning %d todos", n)
	fmt.Fprint(w, body.String())
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.FormValue("todo"))
	log.Printf("todo POST received: %q (len=%d) from %s", text, len(text), r.RemoteAddr)
	if text == "" {
		log.Printf("ignored empty todo submission from %s", r.RemoteAddr)
	} else if len(text) > 140 {
		log.Printf("rejected todo over 140 chars: %q len=%d from %s", text, len(text), r.RemoteAddr)
		http.Error(w, "todo too long (max 140)", http.StatusBadRequest)
		return
	} else {
		if _, err := db.Exec(`INSERT INTO todos (content) VALUES ($1)`, text); err != nil {
			log.Printf("insert failed for %q: %v", text, err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		log.Printf("saved todo: %q from %s (total will increase)", text, r.RemoteAddr)
	}
	http.Redirect(w, r, todoAppURL()+"/", http.StatusSeeOther)
}

func markDone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/todos/"))
	if err != nil {
		http.Error(w, "invalid todo id", http.StatusBadRequest)
		return
	}
	res, err := db.Exec(`UPDATE todos SET done = TRUE WHERE id = $1`, id)
	if err != nil {
		log.Printf("update failed for todo %d: %v", id, err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	log.Printf("marked todo %d as done", id)
	w.WriteHeader(http.StatusNoContent)
}

func checkDB() error {
	conn, err := sql.Open("postgres", dsnFromEnv())
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var result int
	return conn.QueryRowContext(ctx, "SELECT 1").Scan(&result)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--check-db" {
		if err := checkDB(); err != nil {
			log.Printf("database readiness check failed: %v", err)
			os.Exit(1)
		}
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	initDB()

	http.HandleFunc("/live", liveHandler)
	http.HandleFunc("/break", breakAppHandler)

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			listTodos(w)
		case http.MethodPost:
			addTodo(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		markDone(w, r)
	})

	fmt.Println("Todo server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
