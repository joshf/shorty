package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	baseURL := flag.String("base-url", "http://localhost:8080", "public base URL used in short links")
	flag.Parse()

	srv := &Server{
		store:   NewMemoryStore(),
		baseURL: *baseURL,
	}

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, logRequests(srv.routes())))
}

// logRequests is middleware that logs every request and how long it took
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
