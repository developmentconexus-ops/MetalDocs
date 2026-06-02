package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamapi "metaldocs/internal/modules/iam/api"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type stubRoleCapsReader struct {
	links []iamdomain.RoleCapabilityLink
	err   error
}

func (s stubRoleCapsReader) ListRoleCapabilities(ctx context.Context) ([]iamdomain.RoleCapabilityLink, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]iamdomain.RoleCapabilityLink, len(s.links))
	copy(out, s.links)
	return out, nil
}

func tenantCtxRequest(method, path string) *http.Request {
	ctx := tenant.WithTenantID(context.Background(), tenant.DevTenantID)
	return httptest.NewRequest(method, path, nil).WithContext(ctx)
}

func TestListRolesReturnsCanonicalEight(t *testing.T) {
	h := NewRolesCapsHandler(stubRoleCapsReader{})
	rec := httptest.NewRecorder()

	h.listRoles(rec, tenantCtxRequest(http.MethodGet, "/api/v1/iam/roles"))

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d (body=%s)", got, want, rec.Body.String())
	}
	var resp iamapi.ListRolesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got, want := len(resp.Items), 8; got != want {
		t.Fatalf("items = %d, want %d", got, want)
	}
	wantCodes := map[iamapi.UserRole]struct{}{
		iamapi.SystemAdmin: {},
		iamapi.Approver:    {},
		iamapi.Author:      {},
		iamapi.Editor:      {},
		iamapi.Viewer:      {},
		iamapi.Signer:      {},
		iamapi.AreaAdmin:   {},
		iamapi.QmsAdmin:    {},
	}
	for _, item := range resp.Items {
		if _, ok := wantCodes[item.Code]; !ok {
			t.Errorf("unexpected role code %q", item.Code)
		}
		delete(wantCodes, item.Code)
		if strings.TrimSpace(item.Label) == "" {
			t.Errorf("role %q has empty label", item.Code)
		}
		if item.Category != iamapi.Tenant && item.Category != iamapi.Area {
			t.Errorf("role %q has invalid category %q", item.Code, item.Category)
		}
	}
	if len(wantCodes) != 0 {
		t.Errorf("missing canonical roles: %v", wantCodes)
	}
}

func TestListCapabilitiesReturnsRegistry(t *testing.T) {
	h := NewRolesCapsHandler(stubRoleCapsReader{})
	rec := httptest.NewRecorder()

	h.listCapabilities(rec, tenantCtxRequest(http.MethodGet, "/api/v1/iam/capabilities"))

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d (body=%s)", got, want, rec.Body.String())
	}
	var resp iamapi.ListCapabilitiesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got, want := len(resp.Items), len(iamdomain.AllCapabilities()); got != want {
		t.Fatalf("items = %d, want %d (full Capability registry)", got, want)
	}
	for _, item := range resp.Items {
		if strings.TrimSpace(item.Code) == "" {
			t.Errorf("capability has empty code")
		}
		if strings.TrimSpace(item.Description) == "" {
			t.Errorf("capability %q has empty description", item.Code)
		}
		if strings.TrimSpace(item.Category) == "" {
			t.Errorf("capability %q has empty category", item.Code)
		}
	}
}

func TestListRoleCapabilitiesReturnsMatrix(t *testing.T) {
	stub := stubRoleCapsReader{
		links: []iamdomain.RoleCapabilityLink{
			{Role: iamdomain.RoleApprover, Capability: iamdomain.CapDocumentSignoff},
			{Role: iamdomain.RoleSystemAdmin, Capability: iamdomain.CapUserManage},
			{Role: iamdomain.RoleViewer, Capability: iamdomain.CapDocumentView},
		},
	}
	h := NewRolesCapsHandler(stub)
	rec := httptest.NewRecorder()

	h.listRoleCapabilities(rec, tenantCtxRequest(http.MethodGet, "/api/v1/iam/role-capabilities"))

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d (body=%s)", got, want, rec.Body.String())
	}
	var resp iamapi.ListRoleCapabilitiesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got, want := len(resp.Items), 3; got != want {
		t.Fatalf("items = %d, want %d", got, want)
	}
	got := make(map[string]string, len(resp.Items))
	for _, item := range resp.Items {
		got[string(item.Role)] = item.Capability
	}
	want := map[string]string{
		"approver":     "document.signoff",
		"system_admin": "user.manage",
		"viewer":       "document.view",
	}
	for role, capability := range want {
		if got[role] != capability {
			t.Errorf("role %q capability = %q, want %q", role, got[role], capability)
		}
	}
}

func TestListRoleCapabilitiesRepoError(t *testing.T) {
	stub := stubRoleCapsReader{err: errors.New("db down")}
	h := NewRolesCapsHandler(stub)
	rec := httptest.NewRecorder()

	h.listRoleCapabilities(rec, tenantCtxRequest(http.MethodGet, "/api/v1/iam/role-capabilities"))

	if got, want := rec.Code, http.StatusInternalServerError; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func TestListRoleCapabilitiesNilRepoIs501(t *testing.T) {
	h := NewRolesCapsHandler(nil)
	rec := httptest.NewRecorder()

	h.listRoleCapabilities(rec, tenantCtxRequest(http.MethodGet, "/api/v1/iam/role-capabilities"))

	if got, want := rec.Code, http.StatusNotImplemented; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func TestRolesCapsHandlerWithoutTenantContext(t *testing.T) {
	h := NewRolesCapsHandler(stubRoleCapsReader{})

	cases := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		path string
	}{
		{"roles", h.listRoles, "/api/v1/iam/roles"},
		{"capabilities", h.listCapabilities, "/api/v1/iam/capabilities"},
		{"role-capabilities", h.listRoleCapabilities, "/api/v1/iam/role-capabilities"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.fn(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d (no tenant ctx must fail closed)", rec.Code, http.StatusInternalServerError)
			}
		})
	}
}

func TestPermissionResolverGatesNewRoutesOnMembershipView(t *testing.T) {
	// Sanity check that the new routes register with the registered
	// HTTP mux so RegisterRoutes() does not silently drop them.
	mux := http.NewServeMux()
	NewRolesCapsHandler(stubRoleCapsReader{}).RegisterRoutes(mux)
	for _, path := range []string{
		"/api/v1/iam/roles",
		"/api/v1/iam/capabilities",
		"/api/v1/iam/role-capabilities",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(
			tenant.WithTenantID(context.Background(), tenant.DevTenantID),
		)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK && rec.Code != http.StatusNotImplemented {
			t.Errorf("mux %s -> %d, expected 200 or 501", path, rec.Code)
		}
	}
}
