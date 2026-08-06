package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

// taxonomyRequest builds a tenant-scoped GET against the taxonomy router.
func taxonomyRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	return req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
}

// TestListProfiles_AnnotatesHasActiveRoute is deliverable E's contract: the
// admin profile list carries has_active_route so a profile that cannot receive
// documents (no active approval route) is visibly flagged instead of failing
// only at creation time with a 409.
//
// ADR 0086 adds the second, INDEPENDENT flag has_active_template_route. The
// fixture deliberately crosses the two sets (po: document only, it: template
// only) so a handler that resolved both flags from one set would fail.
func TestListProfiles_AnnotatesHasActiveRoute(t *testing.T) {
	svc := fakeProfileServiceWithItems{
		items: []domain.DocumentProfile{
			{Code: "po", TenantID: "test-tenant", FamilyCode: "quality", Name: "Procedimento"},
			{Code: "it", TenantID: "test-tenant", FamilyCode: "quality", Name: "Instrucao"},
		},
		routeReady:         map[string]struct{}{"po": {}},
		templateRouteReady: map[string]struct{}{"it": {}},
	}
	handler := &Handler{profiles: svc}
	mux := http.NewServeMux()
	handler.Mount(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, taxonomyRequest(t, "/api/v1/taxonomy/profiles"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []struct {
			Code                   string `json:"code"`
			HasActiveRoute         *bool  `json:"has_active_route"`
			HasActiveTemplateRoute *bool  `json:"has_active_template_route"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items = %d, want 2 (body=%s)", len(resp.Items), rec.Body.String())
	}
	got := map[string]bool{}
	gotTemplate := map[string]bool{}
	for _, item := range resp.Items {
		if item.HasActiveRoute == nil {
			t.Fatalf("item %q missing required has_active_route", item.Code)
		}
		if item.HasActiveTemplateRoute == nil {
			t.Fatalf("item %q missing required has_active_template_route", item.Code)
		}
		got[item.Code] = *item.HasActiveRoute
		gotTemplate[item.Code] = *item.HasActiveTemplateRoute
	}
	if !got["po"] {
		t.Error("po has an active document route, want has_active_route=true")
	}
	if got["it"] {
		t.Error("it has no active document route, want has_active_route=false")
	}
	if !gotTemplate["it"] {
		t.Error("it has an active template route, want has_active_template_route=true")
	}
	if gotTemplate["po"] {
		t.Error("po has no active template route, want has_active_template_route=false")
	}
}

// TestListProfiles_RouteReadinessFailureIsNotDefaultedToFalse keeps the badge
// honest: a failed readiness read must 500, never silently render every
// profile as route-less (a fabricated integrity signal — no-fallback rule).
func TestListProfiles_RouteReadinessFailureIsNotDefaultedToFalse(t *testing.T) {
	svc := fakeProfileServiceWithItems{
		items:         []domain.DocumentProfile{{Code: "po", TenantID: "test-tenant", FamilyCode: "quality", Name: "Procedimento"}},
		routeReadyErr: errors.New("approval read failed"),
	}
	handler := &Handler{profiles: svc}
	mux := http.NewServeMux()
	handler.Mount(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, taxonomyRequest(t, "/api/v1/taxonomy/profiles"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body = %s", rec.Code, rec.Body.String())
	}
}

// TestGetProfile_AnnotatesHasActiveRoute covers the single-profile read path,
// which resolves readiness through the same bulk reader.
func TestGetProfile_AnnotatesHasActiveRoute(t *testing.T) {
	svc := fakeProfileServiceWithItems{
		items:              []domain.DocumentProfile{{Code: "po", TenantID: "test-tenant", FamilyCode: "quality", Name: "Procedimento"}},
		routeReady:         map[string]struct{}{"po": {}},
		templateRouteReady: map[string]struct{}{},
	}
	handler := &Handler{profiles: svc}
	mux := http.NewServeMux()
	handler.Mount(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, taxonomyRequest(t, "/api/v1/taxonomy/profiles/po"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var item struct {
		HasActiveRoute         *bool `json:"has_active_route"`
		HasActiveTemplateRoute *bool `json:"has_active_template_route"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	if item.HasActiveRoute == nil || !*item.HasActiveRoute {
		t.Fatalf("has_active_route = %v, want true (body=%s)", item.HasActiveRoute, rec.Body.String())
	}
	// The two flags are independent: po has a document route only.
	if item.HasActiveTemplateRoute == nil || *item.HasActiveTemplateRoute {
		t.Fatalf("has_active_template_route = %v, want false (body=%s)", item.HasActiveTemplateRoute, rec.Body.String())
	}
}

// TestGetProfile_AnnotatesHasActiveTemplateRoute is the ADR 0086 mirror: a
// profile with a template route but no document route reports
// has_active_template_route=true and has_active_route=false.
func TestGetProfile_AnnotatesHasActiveTemplateRoute(t *testing.T) {
	svc := fakeProfileServiceWithItems{
		items:              []domain.DocumentProfile{{Code: "it", TenantID: "test-tenant", FamilyCode: "quality", Name: "Instrucao"}},
		routeReady:         map[string]struct{}{},
		templateRouteReady: map[string]struct{}{"it": {}},
	}
	handler := &Handler{profiles: svc}
	mux := http.NewServeMux()
	handler.Mount(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, taxonomyRequest(t, "/api/v1/taxonomy/profiles/it"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var item struct {
		HasActiveRoute         *bool `json:"has_active_route"`
		HasActiveTemplateRoute *bool `json:"has_active_template_route"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	if item.HasActiveTemplateRoute == nil || !*item.HasActiveTemplateRoute {
		t.Fatalf("has_active_template_route = %v, want true (body=%s)", item.HasActiveTemplateRoute, rec.Body.String())
	}
	if item.HasActiveRoute == nil || *item.HasActiveRoute {
		t.Fatalf("has_active_route = %v, want false (body=%s)", item.HasActiveRoute, rec.Body.String())
	}
}
