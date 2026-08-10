package httpresponse_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/httpresponse"
)

// The former TestWriteError_ProblemJSON is gone with httpresponse.WriteError:
// error emission has exactly one entry point now (problem.Respond), and its
// envelope is covered by internal/platform/problem's own tests. What remains
// httpresponse's to prove is the piece it still owns — the Allow header on a
// 405 — and that the envelope it delegates to is still problem+json.
func TestWriteMethodNotAllowed_ProblemJSONWithAllowHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/sessions", nil)

	httpresponse.WriteMethodNotAllowed(rr, req, "GET")

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	if got := rr.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow = %q, want %q", got, "GET")
	}
	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/problem+json")
	}
	var body struct {
		Code   string `json:"code"`
		Status int    `json:"status"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != "request.method_not_allowed" || body.Status != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected body: %+v", body)
	}
}
