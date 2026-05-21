package approvalhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRoutes_AllRoutesRegistered(t *testing.T) {
	mux := http.NewServeMux()
	h := &Handler{}
	h.RegisterRoutes(mux)

	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/documents/doc-1/submit"},
		{method: http.MethodPost, path: "/api/v1/approval/instances/i-1/stages/s-1/signoffs"},
		{method: http.MethodPost, path: "/api/v1/documents/doc-1/publish"},
		{method: http.MethodPost, path: "/api/v1/documents/doc-1/schedule-publish"},
		{method: http.MethodPost, path: "/api/v1/documents/doc-1/supersede"},
		{method: http.MethodPost, path: "/api/v1/documents/doc-1/obsolete"},
		{method: http.MethodPost, path: "/api/v1/approval/instances/i-1/cancel"},
		{method: http.MethodGet, path: "/api/v1/approval/instances/i-1"},
		{method: http.MethodGet, path: "/api/v1/approval/inbox"},
		{method: http.MethodPost, path: "/api/v1/approval/routes"},
		{method: http.MethodPut, path: "/api/v1/approval/routes/r-1"},
		{method: http.MethodDelete, path: "/api/v1/approval/routes/r-1"},
		{method: http.MethodGet, path: "/api/v1/approval/routes"},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("route %s %s not registered (got 404)", rt.method, rt.path)
		}
	}
}

// TestRegisterRoutes_WrapperValidationError exercises the generated wrapper's
// ErrorHandlerFunc path: a malformed UUID path parameter must produce an
// RFC 9457 application/problem+json response, not plain-text http.Error.
func TestRegisterRoutes_WrapperValidationError(t *testing.T) {
	mux := http.NewServeMux()
	(&Handler{}).RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/approval/instances/not-a-uuid", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/problem+json") {
		t.Fatalf("Content-Type: got %q, want application/problem+json", ct)
	}

	var body struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Code   string `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (raw: %s)", err, w.Body.String())
	}
	if body.Status != http.StatusBadRequest {
		t.Errorf("body.status: got %d, want 400", body.Status)
	}
	if body.Code != "validation.param_format" {
		t.Errorf("body.code: got %q, want validation.param_format", body.Code)
	}
	if body.Title == "" {
		t.Error("body.title: empty")
	}
}
