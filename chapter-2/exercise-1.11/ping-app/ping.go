package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const pongsFile = "/usr/src/app/logs/pongs.txt"

func readCount() int {
	data, err := os.ReadFile(pongsFile)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}
	return n
}

func writeCount(n int) error {
	return os.WriteFile(pongsFile, []byte(strconv.Itoa(n)), 0644)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
		count := readCount() + 1
		writeCount(count)
		html := fmt.Sprintf("<h1>pong %s</h1>", count)
		fmt.Fprint(w, html)
	})

	fmt.Println("Ping app server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
