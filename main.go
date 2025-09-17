package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	store = make(map[string]string)
	mu    sync.Mutex
)

var resultTemplate *template.Template

func generateShortURL() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	original := r.FormValue("url")
	if original == "" {
		http.Error(w, "no url provided", http.StatusBadRequest)
		return
	}

	mu.Lock()
	short := generateShortURL()
	store[short] = original
	mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		ShortURL string
		FullURL  string
	}{
		ShortURL: short,
		FullURL:  "http://localhost:8080/" + short,
	}

	err := resultTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Page display error", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]

	if code == "" {
		http.ServeFile(w, r, "index.html")
		return
	}

	mu.Lock()
	original, ok := store[code]
	mu.Unlock()

	if !ok {
		http.ServeFile(w, r, "index.html")
		return
	}

	http.Redirect(w, r, original, http.StatusFound)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if _, err := os.Stat("index.html"); os.IsNotExist(err) {
		log.Fatal("index.html not found in current directory")
	}

	var err error
	resultTemplate, err = template.ParseFiles("result.html")
	if err != nil {
		log.Fatal("Download error result.html: ", err)
	}

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
