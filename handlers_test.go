package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() http.Handler {
	s := &Server{store: NewMemoryStore(), baseURL: "http://test"}
	return s.routes()
}

func TestShortenAndRedirect(t *testing.T) {
	h := newTestServer()

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

func TestShortenRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"not json", `hello`},
		{"empty url", `{"url":""}`},
		{"relative url", `{"url":"/just/a/path"}`},
		{"wrong scheme", `{"url":"ftp://example.com"}`},
	}

	h := newTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/shorten", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}
