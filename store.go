package main

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("link not found")
	ErrExists   = errors.New("code already exists")
)

type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Hits      int       `json:"hits"`
	CreatedAt time.Time `json:"created_at"`
}

// Store is an interface so I can swap in a real database later
type Store interface {
	Save(link Link) error
	Get(code string) (Link, error)
	IncrementHits(code string) error
}

type MemoryStore struct {
	mu    sync.RWMutex // lets reads happen at the same time
	links map[string]Link
}

func NewMemoryStore() *MemoryStore {
	// have to make() the map, writing to a nil map panics
	return &MemoryStore{links: make(map[string]Link)}
}

func (s *MemoryStore) Save(link Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.links[link.Code]; ok {
		return ErrExists
	}
	s.links[link.Code] = link
	return nil
}

func (s *MemoryStore) Get(code string) (Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return link, nil
}

func (s *MemoryStore) IncrementHits(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, ok := s.links[code]
	if !ok {
		return ErrNotFound
	}
	// link is a copy, not a reference into the map, so it has to be put back
	link.Hits++
	s.links[code] = link
	return nil
}
