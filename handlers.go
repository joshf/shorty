package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	code, err := generateCode(6)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate code")
		return
	}
	err = s.store.Save(Link{Code: code, URL: req.URL, CreatedAt: time.Now()})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save link")
		return
	}
	fmt.Println("saved", code, req.URL)

	writeJSON(w, http.StatusCreated, map[string]string{
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
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	// header has to be set before WriteHeader or it gets ignored
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
