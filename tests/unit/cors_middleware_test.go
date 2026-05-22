package unit

import (
  "net/http"
  "net/http/httptest"
  "testing"

  "metaldocs/internal/platform/config"
  "metaldocs/internal/platform/security"
)

func TestCORSPreflightAllowedOrigin(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"http://127.0.0.1:4173"},
    AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "X-Trace-Id"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodOptions, "/api/v1/documents", nil)
  req.Header.Set("Origin", "http://127.0.0.1:4173")
  req.Header.Set("Access-Control-Request-Method", http.MethodGet)
  rec := httptest.NewRecorder()

  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    t.Fatalf("preflight should not reach next handler")
  })).ServeHTTP(rec, req)

  if rec.Code != http.StatusNoContent {
    t.Fatalf("expected 204, got %d", rec.Code)
  }
  if rec.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:4173" {
    t.Fatalf("unexpected allow origin header: %s", rec.Header().Get("Access-Control-Allow-Origin"))
  }
}

func TestCORSDeniesUnknownOriginPreflight(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"http://127.0.0.1:4173"},
    AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "X-Trace-Id"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodOptions, "/api/v1/documents", nil)
  req.Header.Set("Origin", "http://malicious.local")
  req.Header.Set("Access-Control-Request-Method", http.MethodGet)
  rec := httptest.NewRecorder()

  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    t.Fatalf("denied preflight should not reach next handler")
  })).ServeHTTP(rec, req)

  if rec.Code != http.StatusForbidden {
    t.Fatalf("expected 403, got %d", rec.Code)
  }
}

func TestCORSAddsHeadersToSimpleRequest(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"http://127.0.0.1:4173"},
    AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "X-Trace-Id"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
  req.Header.Set("Origin", "http://127.0.0.1:4173")
  rec := httptest.NewRecorder()

  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
  })).ServeHTTP(rec, req)

  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
  if rec.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:4173" {
    t.Fatalf("unexpected allow origin header: %s", rec.Header().Get("Access-Control-Allow-Origin"))
  }
}

func TestCORSDeniesUnknownOriginGET(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"http://127.0.0.1:4173"},
    AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
  req.Header.Set("Origin", "http://malicious.local")
  rec := httptest.NewRecorder()

  reached := false
  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    reached = true
  })).ServeHTTP(rec, req)

  if reached {
    t.Fatalf("disallowed GET origin must not reach handler")
  }
  if rec.Code != http.StatusForbidden {
    t.Fatalf("expected 403, got %d", rec.Code)
  }
  if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
    t.Fatalf("expected application/problem+json, got %s", ct)
  }
}

func TestCORSDeniesUnknownOriginPOSTBeforeHandler(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"http://127.0.0.1:4173"},
    AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", nil)
  req.Header.Set("Origin", "http://malicious.local")
  rec := httptest.NewRecorder()

  reached := false
  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    reached = true
    w.WriteHeader(http.StatusCreated)
  })).ServeHTTP(rec, req)

  if reached {
    t.Fatalf("disallowed POST origin reached handler — server-side side effects would commit")
  }
  if rec.Code != http.StatusForbidden {
    t.Fatalf("expected 403, got %d", rec.Code)
  }
}

func TestCORSAllowlistCaseInsensitive(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"https://App.Example.Com"},
    AllowedMethods: []string{"GET", "POST", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
  req.Header.Set("Origin", "https://app.example.com")
  rec := httptest.NewRecorder()

  reached := false
  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    reached = true
    w.WriteHeader(http.StatusOK)
  })).ServeHTTP(rec, req)

  if !reached {
    t.Fatalf("case-mismatched allowlist must still match per RFC 6454")
  }
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
  if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
    t.Fatalf("expected normalized allow-origin, got %s", got)
  }
}

func TestCORSPassesThroughNoOriginHeader(t *testing.T) {
  middleware := security.NewCORS(config.CORSConfig{
    Enabled:        true,
    AllowedOrigins: []string{"http://127.0.0.1:4173"},
    AllowedMethods: []string{"GET", "POST", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type"},
    MaxAgeSeconds:  300,
  })

  req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
  rec := httptest.NewRecorder()

  reached := false
  middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    reached = true
    w.WriteHeader(http.StatusOK)
  })).ServeHTTP(rec, req)

  if !reached {
    t.Fatalf("no-Origin request must pass through")
  }
  if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
  }
  if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
    t.Fatalf("expected no CORS header for no-Origin request, got %s", got)
  }
}
