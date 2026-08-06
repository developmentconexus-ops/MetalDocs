package httprouter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/httprouter"
)

func TestRecorderRecordsAndDelegates(t *testing.T) {
	inner := http.NewServeMux()
	rec := httprouter.NewRecorder(inner)

	rec.HandleFunc("GET /a", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	rec.Handle("POST /b", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(205) }))

	got := rec.Patterns()
	if len(got) != 2 || got[0] != "GET /a" || got[1] != "POST /b" {
		t.Fatalf("Patterns() = %v", got)
	}

	// Delegation is what makes this a production recorder rather than a test
	// double: the real mux really serves.
	rr := httptest.NewRecorder()
	inner.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/a", nil))
	if rr.Code != 204 {
		t.Fatalf("inner mux did not receive the registration: %d", rr.Code)
	}
}

func TestRecorderPanicsOnNilInner(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic: a nil muxer is a wiring bug, not a branch")
		}
	}()
	httprouter.NewRecorder(nil)
}
