package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	srv := &Server{
		store:   NewMemoryStore(),
		baseURL: "http://localhost:8080",
	}

	log.Printf("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logRequests(srv.routes())))
}

// logRequests is middleware that logs every request and how long it took
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
