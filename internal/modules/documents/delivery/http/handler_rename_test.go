package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRenameDocument_TypedResponseShape asserts F1.2/A2 contract:
// PATCH /api/v1/documents/{id} returns 200 OK with EMPTY body (no JSON envelope, no raw
// domain.Document leak). OpenAPI declares the response as 200 no-schema; consumer FE adapter
// is Promise<void>.
func TestRenameDocument_TypedResponseShape(t *testing.T) {
	svc := &fakeSvc{}
	mux := newMux(t, svc)

	body, _ := json.Marshal(map[string]any{"name": "renamed"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/22222222-2222-4222-8222-222222222222", bytes.NewReader(body))
	req.SetPathValue("id", "22222222-2222-4222-8222-222222222222")
	withAuthHeaders(req, "editor")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rr.Code, rr.Body.String())
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %d bytes: %q", rr.Body.Len(), rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "" {
		t.Fatalf("expected no Content-Type header on empty body, got %q", ct)
	}
	if svc.renameName != "renamed" {
		t.Fatalf("expected service rename called with name=%q, got %q", "renamed", svc.renameName)
	}
}
