package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}
