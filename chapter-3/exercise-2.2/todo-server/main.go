package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type todo struct {
	Text string
	Done bool
}

var (
	mu    sync.Mutex
	todos = []todo{
		{Text: "Learn Kubernetes basics", Done: true},
		{Text: "Set up the persistent volume", Done: false},
		{Text: "Build the todo app", Done: false},
	}
)

func todoAppURL() string {
	url := os.Getenv("TODO_APP_URL")
	if url == "" {
		url = "http://todo-app-svc:2345"
	}
	return url
}

func listTodos(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()
	log.Printf("returning %d todos", len(todos))

	var body strings.Builder
	for _, t := range todos {
		mark := "&#9745;"
		class := "done"
		if !t.Done {
			mark = "&#9744;"
			class = ""
		}
		fmt.Fprintf(&body, `<li class="%s">%s <span>%s</span></li>`, class, mark, t.Text)
	}
	log.Printf("todos response ready: %d bytes", body.Len())
	fmt.Fprint(w, body.String())
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("todo")
	mu.Lock()
	if text != "" {
		todos = append(todos, todo{Text: text})
		log.Printf("saved todo: %q (total %d todos)", text, len(todos))
	} else {
		log.Printf("ignored empty todo submission")
	}
	mu.Unlock()
	log.Printf("returning response after saving todo")
	http.Redirect(w, r, todoAppURL()+"/", http.StatusSeeOther)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	addr := ":" + port

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
