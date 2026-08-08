package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}

	return string(result)
}

func main() {
	const filePath = "/usr/src/app/logs/data.txt"

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("opened file: %s", filePath)
	defer f.Close()

	for {
		s := generateRandomString(14)

		if _, err := fmt.Fprintln(f, s); err != nil {
			f.Close()
			fmt.Fprintf(os.Stderr, "error writing to the file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("wrote: %s\n", s)
		time.Sleep(5 * time.Second)
	}
}
