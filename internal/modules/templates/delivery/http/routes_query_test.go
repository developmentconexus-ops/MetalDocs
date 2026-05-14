package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/templates/domain"
)

func TestGetSystemBlankTemplate_RequiresTemplateViewAuthz(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["00000000-0000-0000-0000-000000000101"] = &domain.Template{
		ID:            "00000000-0000-0000-0000-000000000101",
		TenantID:      "ffffffff-ffff-ffff-ffff-ffffffffffff",
		Name:          "Blank",
		LatestVersion: 1,
	}
	repo.versions["blank-ver-1"] = &domain.TemplateVersion{
		ID:            "blank-ver-1",
		TemplateID:    "00000000-0000-0000-0000-000000000101",
		VersionNumber: 1,
	}

	var gotTenant, gotArea, gotAction string
	authz := func(_ *http.Request, tenantID, area, action string) error {
		gotTenant = tenantID
		gotArea = area
		gotAction = action
		return nil
	}

	mux := newMux(t, authz, repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/system/blank", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if gotTenant != "tenant-a" || gotArea != "*" || gotAction != "template.view" {
		t.Fatalf("unexpected authz call: tenant=%q area=%q action=%q", gotTenant, gotArea, gotAction)
	}
}

