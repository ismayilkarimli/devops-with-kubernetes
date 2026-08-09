package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	imageURL   = "https://picsum.photos/128"
	maxAge     = 10 * time.Minute
	waitWindow = 5 * time.Second
	imageFile  = "current.jpg"
	stampFile  = "current.txt"
)

var cacheMu sync.Mutex

func cacheDir() string {
	if d := os.Getenv("CACHE_DIR"); d != "" {
		return d
	}
	return "/usr/src/app/data"
}

func loadCache() ([]byte, time.Time, bool) {
	dir := cacheDir()
	data, err := os.ReadFile(filepath.Join(dir, imageFile))
	if err != nil {
		return nil, time.Time{}, false
	}
	raw, err := os.ReadFile(filepath.Join(dir, stampFile))
	if err != nil {
		return nil, time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, time.Time{}, false
	}
	return data, ts, true
}

func storeCache(data []byte, ts time.Time) error {
	dir := cacheDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, imageFile), data, 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, stampFile), []byte(ts.Format(time.RFC3339)), 0644)
}

func clearCache() error {
	dir := cacheDir()
	if err := os.Remove(filepath.Join(dir, imageFile)); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(filepath.Join(dir, stampFile)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func fetchImage() ([]byte, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func refreshCache() ([]byte, bool) {
	log.Println("fetching image from picsum:", imageURL)
	data, err := fetchImage()
	if err != nil {
		log.Printf("fetch failed: %v", err)
		return nil, false
	}
	now := time.Now()
	if err := storeCache(data, now); err != nil {
		log.Printf("store failed: %v", err)
		return nil, false
	}
	log.Printf("cached new image with timestamp %s", now.Format(time.RFC3339))
	return data, true
}

func cachedImage() ([]byte, string) {
	deadline := time.Now().Add(waitWindow)

	for time.Now().Before(deadline) {
		cacheMu.Lock()
		cached, ts, ok := loadCache()
		if ok && time.Since(ts) >= maxAge {
			log.Printf("evicting stale image (served at %s)", ts.Format(time.RFC3339))
			_ = clearCache()
			cached, _ = refreshCache()
		}
		if cached == nil {
			cached, _ = refreshCache()
		}
		cacheMu.Unlock()

		if cached != nil {
			log.Printf("serving cached image to request")
			return cached, base64.StdEncoding.EncodeToString(cached)
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("no image available for request")
	return nil, ""
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	cacheMu.Lock()
	if _, _, ok := loadCache(); !ok {
		refreshCache()
	}
	cacheMu.Unlock()

	http.HandleFunc("/", homeHandler)
	log.Println("Server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type todo struct {
	Text string
	Done bool
}

var seedTodos = []todo{
	{Text: "Learn Kubernetes basics", Done: true},
	{Text: "Set up the persistent volume", Done: false},
	{Text: "Build the todo app", Done: false},
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var body strings.Builder
	for _, t := range seedTodos {
		mark := "&#9745;"
		class := "done"
		if !t.Done {
			mark = "&#9744;"
			class = ""
		}
		fmt.Fprintf(&body, `<li class="%s">%s <span>%s</span></li>`, class, mark, t.Text)
	}

	img, encoded := cachedImage()
	var imgHTML string
	if img != nil {
		imgHTML = fmt.Sprintf(`<img src="data:image/jpeg;base64,%s" alt="random picture" width="128" height="128">`, encoded)
	} else {
		imgHTML = `<p>😢 The picture could not be fetched right now. Please try again later.</p>`
	}

	fmt.Fprintf(w, homeTpl, imgHTML, body.String())
}

const homeTpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Random Picture &amp; Todos</title>
</head>
<body style="background:#1b1b2f;color:#e8e8f0;font-family:system-ui,-apple-system,Segoe UI,sans-serif;margin:0;padding:0;min-height:100vh">
	<main style="max-width:560px;margin:0 auto;padding:64px 24px">
		<header style="text-align:center;margin-bottom:32px">
			<p style="margin:0 0 8px;color:#8b8ba8;font-size:13px;letter-spacing:.08em;text-transform:uppercase">Chapter 2 · Exercise 1.13</p>
			<h1 style="margin:0 0 20px;font-size:34px;font-weight:700;text-wrap:balance">Random Picture</h1>
			%s
			<p style="margin:12px 0 0;color:#8b8ba8;font-size:13px">A fresh picture, cached for 10 minutes.</p>
		</header>

		<section>
			<h2 style="margin:0 0 16px;font-size:22px;font-weight:600">Todos</h2>
			<form method="post" action="/" style="display:flex;gap:12px;margin-bottom:28px">
				<input type="text" name="todo" maxlength="140" required
					placeholder="What needs doing?"
					aria-label="New todo (max 140 characters)"
					style="flex:1;padding:14px 16px;font-size:15px;border:1px solid #3a3a55;border-radius:10px;background:#22223a;color:#e8e8f0;outline:none">
				<button type="submit"
					style="padding:14px 20px;font-size:15px;font-weight:600;border:none;border-radius:10px;background:#e0527b;color:#fff;cursor:pointer">Add todo</button>
			</form>
			<ul style="list-style:none;margin:0;padding:0">
				%s
			</ul>
		</section>
	</main>
</body>
</html>`


