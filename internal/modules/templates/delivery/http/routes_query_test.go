package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	templatesapi "metaldocs/internal/modules/templates/api"
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
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:            "22222222-2222-4222-8222-222222222222",
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
	if gotTenant != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" || gotArea != "*" || gotAction != "template.view" {
		t.Fatalf("unexpected authz call: tenant=%q area=%q action=%q", gotTenant, gotArea, gotAction)
	}
}

func TestListTemplates_LimitOver200Rejected(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{ID: "tpl-1", TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Name: "Template"}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates?limit=201", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "request.invalid" {
		t.Fatalf("expected error.code=VALIDATION_ERROR, got %q", out.Code)
	}
}

// TestListTemplates_PublishedTrueReturnsOnlyPublished proves the
// published=true query param maps through to ListFilter.PublishedOnly and
// excludes templates with no published version, while omitting the param
// keeps the management view (every status).
func TestListTemplates_PublishedTrueReturnsOnlyPublished(t *testing.T) {
	publishedTplID := "11111111-1111-4111-8111-111111111111"
	draftTplID := "33333333-3333-4333-8333-333333333333"
	publishedVerID := "22222222-2222-4222-8222-222222222222"
	repo := newFakeRepo()
	repo.templates[publishedTplID] = &domain.Template{
		ID: publishedTplID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Name: "Published Template", PublishedVersionID: &publishedVerID,
	}
	repo.templates[draftTplID] = &domain.Template{
		ID: draftTplID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Name: "Draft Template",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates?published=true", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp templatesapi.ListTemplatesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(resp.Data.Templates) != 1 || resp.Data.Templates[0].Name != "Published Template" {
		t.Fatalf("expected only the published template, got %+v", resp.Data.Templates)
	}

	// Omitted published param keeps the management view (both statuses).
	reqAll := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	withHeaders(reqAll)
	rrAll := httptest.NewRecorder()
	mux.ServeHTTP(rrAll, reqAll)
	if rrAll.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rrAll.Code, rrAll.Body.String())
	}
	var respAll templatesapi.ListTemplatesResponse
	if err := json.Unmarshal(rrAll.Body.Bytes(), &respAll); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(respAll.Data.Templates) != 2 {
		t.Fatalf("expected both templates in management view, got %+v", respAll.Data.Templates)
	}
}
