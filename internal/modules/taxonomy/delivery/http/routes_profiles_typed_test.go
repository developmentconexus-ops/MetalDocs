package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeProfileServiceWithItems struct {
	items []domain.DocumentProfile
	// routeReady is the set of profile codes reported as having an active
	// approval route (has_active_route); nil means none.
	routeReady map[string]struct{}
	// routeReadyErr simulates an approval-readiness read failure; the handler
	// must 500 rather than default has_active_route to false.
	routeReadyErr error
}

func (f fakeProfileServiceWithItems) RouteReadySubjects(_ context.Context, _ string) (map[string]struct{}, error) {
	if f.routeReadyErr != nil {
		return nil, f.routeReadyErr
	}
	if f.routeReady == nil {
		return map[string]struct{}{}, nil
	}
	return f.routeReady, nil
}

func (f fakeProfileServiceWithItems) List(_ context.Context, _ string, _ bool) ([]domain.DocumentProfile, error) {
	return f.items, nil
}

func (f fakeProfileServiceWithItems) Get(_ context.Context, _ string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	for i, p := range f.items {
		if p.Code == code {
			return &f.items[i], nil
		}
	}
	return nil, domain.ErrProfileNotFound
}

func (f fakeProfileServiceWithItems) Create(_ context.Context, _ *domain.DocumentProfile) error {
	return nil
}

func (f fakeProfileServiceWithItems) Update(_ context.Context, _ *domain.DocumentProfile) error {
	return nil
}

func (f fakeProfileServiceWithItems) Reclassify(_ context.Context, _ string, _ domain.ProfileCode, _ domain.GovernanceClass, _ string) error {
	return nil
}

func (f fakeProfileServiceWithItems) SetDefaultTemplate(_ context.Context, _ string, _ domain.ProfileCode, _, _ string) error {
	return nil
}

func (f fakeProfileServiceWithItems) Archive(_ context.Context, _ string, _ domain.ProfileCode, _ string) error {
	return nil
}

func ptrStr(s string) *string { return &s }

// TestListProfiles_DropsDomainFields asserts that listProfiles response
// does NOT emit domain-only fields (tenant_id, owner_user_id, editable_by_role).
// RED before fix: current handler emits raw domain.DocumentProfile which includes these fields.
func TestListProfiles_DropsDomainFields(t *testing.T) {
	svc := fakeProfileServiceWithItems{items: []domain.DocumentProfile{
		{
			Code:           "ISO-9001",
			TenantID:       "tenant-x",
			FamilyCode:     "quality",
			Name:           "ISO 9001",
			Description:    "Quality management",
			Alias:          "iso",
			OwnerUserID:    ptrStr("user-123"),
			EditableByRole: "editor",
		},
	}}
	handler := &Handler{profiles: svc}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/profiles", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected at least one item")
	}
	item := resp.Items[0]
	for _, forbidden := range []string{"tenant_id", "owner_user_id", "editable_by_role", "TenantID", "OwnerUserID", "EditableByRole"} {
		if _, ok := item[forbidden]; ok {
			t.Errorf("response item must not contain domain-only field %q (H-D class: raw domain.DocumentProfile emitted)", forbidden)
		}
	}
	for _, required := range []string{"code", "name", "family_code"} {
		if _, ok := item[required]; !ok {
			t.Errorf("response item missing required field %q", required)
		}
	}
}

// TestGetProfile_DropsDomainFields asserts that getProfile response
// does NOT emit domain-only fields (tenant_id, owner_user_id, editable_by_role).
// RED before fix: current handler emits raw domain.DocumentProfile which includes these fields.
func TestGetProfile_DropsDomainFields(t *testing.T) {
	svc := fakeProfileServiceWithItems{items: []domain.DocumentProfile{
		{
			Code:           "ISO-9001",
			TenantID:       "tenant-x",
			FamilyCode:     "quality",
			Name:           "ISO 9001",
			Description:    "Quality management",
			Alias:          "iso",
			OwnerUserID:    ptrStr("user-123"),
			EditableByRole: "editor",
		},
	}}
	handler := &Handler{profiles: svc}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/profiles/ISO-9001", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var item map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rec.Body.String())
	}
	for _, forbidden := range []string{"tenant_id", "owner_user_id", "editable_by_role", "TenantID", "OwnerUserID", "EditableByRole"} {
		if _, ok := item[forbidden]; ok {
			t.Errorf("response must not contain domain-only field %q (H-D class: raw domain.DocumentProfile emitted)", forbidden)
		}
	}
	for _, required := range []string{"code", "name", "family_code"} {
		if _, ok := item[required]; !ok {
			t.Errorf("response missing required field %q", required)
		}
	}
}
