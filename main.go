package main

import (
	"fmt"
	"net/http"
)

func main() {
	srv := &Server{
		store:   NewMemoryStore(),
		baseURL: "http://localhost:8080",
	}

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", srv.routes())
}
