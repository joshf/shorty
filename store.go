package main

import (
	"sync"
	"time"
)

type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// maps arent safe to use from multiple goroutines and every request
// runs in its own goroutine, so lock around them
var (
	links = map[string]Link{}
	mu    sync.Mutex
)

func saveLink(l Link) {
	mu.Lock()
	defer mu.Unlock()
	links[l.Code] = l
}

func getLink(code string) (Link, bool) {
	mu.Lock()
	defer mu.Unlock()
	l, ok := links[code]
	return l, ok
}
