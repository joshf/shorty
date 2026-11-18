package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
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
	mux.HandleFunc("GET /{code}/stats", s.handleStats)
	return mux
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	// 1 MiB is way more than any url needs
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	u, err := url.ParseRequestURI(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeError(w, http.StatusBadRequest, "url must be an absolute http(s) URL")
		return
	}

	// codes are random so they could collide, try a few times
	for range 5 {
		code, err := generateCode(6)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate code")
			return
		}

		err = s.store.Save(Link{Code: code, URL: u.String(), CreatedAt: time.Now()})
		if errors.Is(err, ErrExists) {
			continue
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not save link")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{
			"code":      code,
			"short_url": s.baseURL + "/" + code,
		})
		return
	}

	writeError(w, http.StatusInternalServerError, "could not find a free code")
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	link, err := s.store.Get(code)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}

	// dont fail the redirect just because the counter didnt work
	if err := s.store.IncrementHits(code); err != nil {
		log.Printf("increment hits for %s: %v", code, err)
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	link, err := s.store.Get(r.PathValue("code"))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, link)
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
