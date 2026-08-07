package http_test

// Feature F6.2 — typed-shape contract tests for the templates query 200 responses.
// Each subtest decodes the handler body with json.Decoder + DisallowUnknownFields() into
// the codegen-generated *Response type. Asserts the typed field round-trips non-zero —
// locking the handlers to the OpenAPI declaration so future drift fails at this gate.

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/domain"
)

func TestQuery_TypedResponseShape(t *testing.T) {
	const (
		tplID    = "11111111-1111-1111-1111-111111111111"
		tenantID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		verID    = "22222222-2222-4222-8222-222222222222"
		actorID  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	)

	t.Run("list_templates", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{
			ID:        tplID,
			TenantID:  tenantID,
			Name:      "My Template",
			Key:       "my-template",
			CreatedBy: "user-a",
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.ListTemplatesResponse](t, rr.Body.Bytes())
		if out.Data.Templates == nil {
			t.Fatal("data.templates nil")
		}
		if len(out.Data.Templates) == 0 {
			t.Fatal("data.templates empty; expected at least 1")
		}
		if out.Meta.Limit == 0 {
			t.Fatal("meta.limit zero")
		}
	})

	t.Run("get_system_blank", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates["00000000-0000-0000-0000-000000000101"] = &domain.Template{
			ID:            "00000000-0000-0000-0000-000000000101",
			TenantID:      "ffffffff-ffff-ffff-ffff-ffffffffffff",
			Name:          "Blank",
			Key:           "blank",
			CreatedBy:     "system",
			LatestVersion: 1,
		}
		repo.versions[verID] = &domain.TemplateVersion{
			ID:            verID,
			TemplateID:    "00000000-0000-0000-0000-000000000101",
			VersionNumber: 1,
			Status:        domain.VersionStatusDraft,
			ContentHash:   "deadbeef",
			AuthorID:      "system",
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/system/blank", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.SystemBlankTemplateResponse](t, rr.Body.Bytes())
		if out.Name == "" {
			t.Fatal("name empty")
		}
		if out.TemplateId.String() == "" {
			t.Fatal("template_id empty")
		}
		if out.TemplateVersionId.String() == "" {
			t.Fatal("template_version_id empty")
		}
	})

	t.Run("get_template", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{
			ID:            tplID,
			TenantID:      tenantID,
			Name:          "My Template",
			Key:           "my-template",
			CreatedBy:     "user-a",
			LatestVersion: 1,
		}
		repo.versions[verID] = &domain.TemplateVersion{
			ID:            verID,
			TemplateID:    tplID,
			VersionNumber: 1,
			Status:        domain.VersionStatusDraft,
			ContentHash:   "deadbeef",
			AuthorID:      "user-a",
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+tplID, nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.GetTemplateResponse](t, rr.Body.Bytes())
		if out.Data.Template.Id.String() == "" {
			t.Fatal("data.template.id empty")
		}
		if out.Data.LatestVersion.Id.String() == "" {
			t.Fatal("data.latest_version.id empty")
		}
	})

	t.Run("get_docx_url", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{
			ID:            tplID,
			TenantID:      tenantID,
			Name:          "My Template",
			Key:           "my-template",
			CreatedBy:     "user-a",
			LatestVersion: 1,
		}
		repo.versions[verID] = &domain.TemplateVersion{
			ID:             verID,
			TemplateID:     tplID,
			VersionNumber:  1,
			Status:         domain.VersionStatusPublished,
			ContentHash:    "deadbeef",
			AuthorID:       "user-a",
			DocxStorageKey: "tenants/" + tenantID + "/templates/" + tplID + "/v1/template.docx",
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+tplID+"/versions/1/docx-url", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.GetTemplateDocxUrlResponse](t, rr.Body.Bytes())
		if out.Data.Url == "" {
			t.Fatal("data.url empty")
		}
	})

	t.Run("list_audit", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{
			ID:            tplID,
			TenantID:      tenantID,
			Name:          "My Template",
			Key:           "my-template",
			CreatedBy:     "user-a",
			LatestVersion: 1,
		}
		auditEvent, err := domain.NewAuditEvent(
			tenantID, tplID, actorID, nil,
			domain.AuditCreated, nil,
			time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
		)
		if err != nil {
			t.Fatalf("NewAuditEvent: %v", err)
		}
		repo.audit = append(repo.audit, &auditEvent)
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+tplID+"/audit", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.ListTemplateAuditResponse](t, rr.Body.Bytes())
		if out.Data.Audit == nil {
			t.Fatal("data.audit nil")
		}
		if len(out.Data.Audit) == 0 {
			t.Fatal("data.audit empty; expected 1 event")
		}
		if out.Data.Audit[0].Action == "" {
			t.Fatal("audit[0].action empty")
		}
	})
}
