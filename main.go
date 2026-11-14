package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// maps arent safe to use from multiple goroutines and every request
// runs in its own goroutine, so lock around them
var (
	urls = map[string]string{}
	mu   sync.Mutex
)

type shortenRequest struct {
	URL string `json:"url"`
}

func shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", 400)
		return
	}

	code, err := generateCode(6)
	if err != nil {
		http.Error(w, "could not generate code", 500)
		return
	}
	mu.Lock()
	urls[code] = req.URL
	mu.Unlock()
	fmt.Println("saved", code, req.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code":      code,
		"short_url": "http://localhost:8080/" + code,
	})
}

func redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	fmt.Println("looking up", code)

	mu.Lock()
	url, ok := urls[code]
	mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, 302)
}

func main() {
	mux := http.NewServeMux()
	// go 1.22 lets you put the method and wildcards in the pattern
	mux.HandleFunc("POST /shorten", shorten)
	mux.HandleFunc("GET /{code}", redirect)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
