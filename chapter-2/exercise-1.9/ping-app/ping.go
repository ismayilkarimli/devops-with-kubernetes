package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	counter := 0

	http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
		html := fmt.Sprintf("<h1>pong %s</h1>", counter)
		fmt.Fprint(w, html)
		counter++
	})

	fmt.Println("Ping app server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
