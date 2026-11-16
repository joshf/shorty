package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenAndRedirect(t *testing.T) {
	s := &Server{store: NewMemoryStore(), baseURL: "http://test"}
	h := s.routes()

	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url":"https://go.dev"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("shorten: got status %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	req = httptest.NewRequest("GET", "/"+resp.Code, nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("redirect: got status %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "https://go.dev" {
		t.Errorf("redirect: got Location %q, want %q", loc, "https://go.dev")
	}
}
