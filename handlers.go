package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Server holds what the handlers need instead of using globals
type Server struct {
	store   Store
	baseURL string
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	// go 1.22 lets you put the method and wildcards in the pattern
	mux.HandleFunc("POST /shorten", s.handleShorten)
	mux.HandleFunc("GET /{code}", s.handleRedirect)
	return mux
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
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
	err = s.store.Save(Link{Code: code, URL: req.URL, CreatedAt: time.Now()})
	if err != nil {
		http.Error(w, "could not save link", 500)
		return
	}
	fmt.Println("saved", code, req.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code":      code,
		"short_url": s.baseURL + "/" + code,
	})
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	fmt.Println("looking up", code)

	link, err := s.store.Get(code)
	if err == ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "lookup failed", 500)
		return
	}
	http.Redirect(w, r, link.URL, 302)
}
