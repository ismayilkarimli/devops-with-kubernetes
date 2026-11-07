package main

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

func pongServiceURL() string {
	url := os.Getenv("PONG_SERVICE_URL")
	if url == "" {
		url = "http://ping-pong-app-svc:2346/pings"
	}
	return url
}

func fetchPongs() string {
	resp, err := http.Get(pongServiceURL())
	if err != nil {
		return "0"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "0"
	}

	if n := strings.TrimSpace(string(body)); n != "" {
		return n
	}
	return "0"
}

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
		fileContent, err := os.ReadFile("/etc/config/information.txt")
		if err != nil {
			fileContent = []byte("could not read file")
		}
		message := os.Getenv("MESSAGE")
		if message == "" {
			message = "no message"
		}

		timestamp := time.Now().Format(time.RFC3339)
		uuid := generateRandomString(14)
		html := fmt.Sprintf("file content: %s\nenv variable: MESSAGE=%s\n%s : %s.\nPing / Pongs: %s",
			strings.TrimSpace(string(fileContent)), message, timestamp, uuid, fetchPongs())
		fmt.Fprint(w, html)
	})

	fmt.Println("Log app server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
