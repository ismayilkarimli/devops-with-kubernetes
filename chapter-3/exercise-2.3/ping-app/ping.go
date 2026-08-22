package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	mu    sync.Mutex
	count int
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		n := count
		mu.Unlock()
		html := fmt.Sprintf("<h1>pong %d</h1>", n)
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/pings", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		n := count
		mu.Unlock()
		fmt.Fprint(w, n)
	})

	fmt.Println("Ping app server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
