package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

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
	rows, err := db.Query(`SELECT content, done FROM todos ORDER BY id`)
	if err != nil {
		log.Printf("query failed: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var body strings.Builder
	n := 0
	for rows.Next() {
		var text string
		var done bool
		if err := rows.Scan(&text, &done); err != nil {
			log.Printf("scan failed: %v", err)
			continue
		}
		n++
		mark := "&#9745;"
		if !done {
			mark = "&#9744;"
		}
		fmt.Fprintf(&body, `<li>%s <span>%s</span></li>`, mark, text)
	}
	log.Printf("returning %d todos", n)
	fmt.Fprint(w, body.String())
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.FormValue("todo"))
	if text == "" {
		log.Printf("ignored empty todo submission")
	} else if len(text) > 140 {
		log.Printf("rejected todo over 140 chars: %q", text)
		http.Error(w, "todo too long (max 140)", http.StatusBadRequest)
		return
	} else {
		if _, err := db.Exec(`INSERT INTO todos (content) VALUES ($1)`, text); err != nil {
			log.Printf("insert failed: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		log.Printf("saved todo: %q", text)
	}
	http.Redirect(w, r, todoAppURL()+"/", http.StatusSeeOther)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	initDB()

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

	fmt.Println("Todo server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
