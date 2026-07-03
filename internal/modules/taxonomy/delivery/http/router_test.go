package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRegisterRoutes_AllRoutesRegistered pins every operation declared for the
// taxonomy tag in api/openapi/v1/openapi.yaml to a live mux registration —
// mirrors internal/modules/documents/approval/http/router_test.go. A spec
// route with no ServerInterface method fails to compile (routes_generated.go
// asserts `var _ taxonomyapi.ServerInterface = (*Handler)(nil)`); this test
// closes the remaining gap where a route compiles but is never reachable
// (e.g. dropped from RegisterRoutes, wrong BaseURL). A 404 here means the
// path is unreachable at runtime even though the spec declares it.
func TestRegisterRoutes_AllRoutesRegistered(t *testing.T) {
	mux := http.NewServeMux()
	h := NewHandler(fakeProfileService{}, fakeAreaService{}, fakeFamilyService{})
	h.RegisterRoutes(mux)

	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/taxonomy/profiles"},
		{method: http.MethodPost, path: "/api/v1/taxonomy/profiles"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/profiles/p-1"},
		{method: http.MethodPatch, path: "/api/v1/taxonomy/profiles/p-1"},
		{method: http.MethodDelete, path: "/api/v1/taxonomy/profiles/p-1"},
		{method: http.MethodPut, path: "/api/v1/taxonomy/profiles/p-1/default-template"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/areas"},
		{method: http.MethodPost, path: "/api/v1/taxonomy/areas"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/areas/a-1"},
		{method: http.MethodPut, path: "/api/v1/taxonomy/areas/a-1"},
		{method: http.MethodDelete, path: "/api/v1/taxonomy/areas/a-1"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/families"},
		{method: http.MethodPost, path: "/api/v1/taxonomy/families"},
		{method: http.MethodGet, path: "/api/v1/taxonomy/families/f-1"},
		{method: http.MethodPatch, path: "/api/v1/taxonomy/families/f-1"},
		{method: http.MethodDelete, path: "/api/v1/taxonomy/families/f-1"},
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
