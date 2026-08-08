package main

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
)

const (
	filePath = "/usr/src/app/logs/data.txt"
	port     = ":8080"
)

func fileHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "could not read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Escape content and wrap each line in a <p> for readable HTML output
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString("<p>")
		sb.WriteString(html.EscapeString(line))
		sb.WriteString("</p>\n")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>File Contents</title>
  <style>
    body { font-family: monospace; padding: 2rem; background: #f5f5f5; }
    h1   { font-size: 1.2rem; color: #333; }
    p    { margin: 0.2rem 0; color: #222; }
  </style>
</head>
<body>
  <h1>Contents of %s</h1>
%s
</body>
</html>`, filePath, sb.String())
}

func main() {
	http.HandleFunc("/", fileHandler)
	fmt.Printf("server listening on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
