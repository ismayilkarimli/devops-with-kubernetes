package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)

// generateRandomString picks random characters from a given charset
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}

	return string(result)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now().Format(time.RFC3339)
		uuid := generateRandomString(14)
		html := fmt.Sprintf("<h1>%s : %s</h1>", timestamp, uuid)
		fmt.Fprint(w, html)
	})

	fmt.Println("Log app server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
