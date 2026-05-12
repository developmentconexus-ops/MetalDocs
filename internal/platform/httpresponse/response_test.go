package httpresponse_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/httpresponse"
)

func TestWriteError_ProblemJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	httpresponse.WriteError(rr, http.StatusNotFound, "NOT_FOUND", "resource not found")

	if got := rr.Code; got != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", got, http.StatusNotFound)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want %q", ct, "application/problem+json")
	}
	var body struct {
		Code   string `json:"code"`
		Title  string `json:"title"`
		Status int    `json:"status"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != "NOT_FOUND" {
		t.Fatalf("code = %q, want NOT_FOUND", body.Code)
	}
	if body.Status != http.StatusNotFound {
		t.Fatalf("status field = %d, want %d", body.Status, http.StatusNotFound)
	}
}
