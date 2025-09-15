package main

import (
	"fmt"
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

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Short URL: <a href='/%s'>http://localhost:8080/%s</a><br><a href='/'>Back to form</a>", short, short)
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

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server started at :8080")
	fmt.Println("Make sure index.html is in the same directory as the executable")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
