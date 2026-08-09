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
	imageURL   = "https://picsum.photos/1200"
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

func handler(w http.ResponseWriter, r *http.Request) {
	deadline := time.Now().Add(waitWindow)

	var data []byte
	for time.Now().Before(deadline) {
		cacheMu.Lock()
		cached, ts, ok := loadCache()
		if ok {
			if time.Since(ts) >= maxAge {
				log.Printf("evicting stale image (served at %s)", ts.Format(time.RFC3339))
				_ = clearCache()
				if fresh, fetched := refreshCache(); fetched {
					cached = fresh
				} else {
					cached = nil
				}
			}
			data = cached
		}
		cacheMu.Unlock()

		if data != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if data == nil {
		log.Printf("no image available for request from %s", r.RemoteAddr)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<h1>😢 Sorry, the picture could not be fetched in time. Please try again later.</h1>`)
		return
	}

	log.Printf("serving cached image to request from %s", r.RemoteAddr)
	encoded := base64.StdEncoding.EncodeToString(data)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, htmlTpl, encoded)
}

const htmlTpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Random Picture</title>
</head>
<body style="background:#1b1b2f;color:#eee;font-family:system-ui,sans-serif;text-align:center;margin:0;padding:40px">
	<h1>Random Picture</h1>
	<img src="data:image/jpeg;base64,%s" alt="random picture" style="max-width:90vw;border-radius:12px;box-shadow:0 8px 24px rgba(0,0,0,.5)">
	<p><small>Refresh to see if a new picture has arrived.</small></p>
</body>
</html>`

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

	http.HandleFunc("/", handler)
	log.Println("Server started in port", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
