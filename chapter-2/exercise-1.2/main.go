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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello")
	})

	fmt.Println("Server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
