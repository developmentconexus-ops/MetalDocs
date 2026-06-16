package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/tenant"
)

// templateVersionDeclaredKeys is the flat snake_case key set declared by the OpenAPI
// component schema VersionDTO (reused by F1.2 / ADR 0035 — no new schema name). Used by
// V2, V3a, V3b shape assertions.
var templateVersionDeclaredKeys = map[string]struct{}{
	"id":                    {},
	"template_id":           {},
	"version_number":        {},
	"revision_number":       {},
	"status":                {},
	"docx_storage_key":      {},
	"content_hash":          {},
	"metadata_schema":       {},
	"placeholder_schema":    {},
	"author_id":             {},
	"pending_reviewer_role": {},
	"pending_approver_role": {},
	"reviewer_id":           {},
	"approver_id":           {},
	"submitted_at":          {},
	"reviewed_at":           {},
	"approved_at":           {},
	"published_at":          {},
	"obsoleted_at":          {},
	"lock_version":          {},
	"created_at":            {},
}

func assertFlatTemplateVersionShape(t *testing.T, raw map[string]json.RawMessage) {
	t.Helper()
	if _, ok := raw["data"]; ok {
		t.Fatalf("forbidden top-level key %q — F1.2/ADR-0035 drops {data:{...}} envelope on this endpoint", "data")
	}
	for k := range raw {
		if _, ok := templateVersionDeclaredKeys[k]; !ok {
			t.Fatalf("unexpected key %q in TemplateVersion response (extra/PascalCase): keys=%v", k, mapKeysT(raw))
		}
	}
	// Required keys (mirror OpenAPI VersionDTO required list).
	for _, required := range []string{
		"id", "template_id", "version_number", "status", "author_id", "created_at",
	} {
		if _, ok := raw[required]; !ok {
			t.Fatalf("missing required key %q in TemplateVersion response: keys=%v", required, mapKeysT(raw))
		}
	}
}

func mapKeysT(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestCreateNextVersion_TypedResponseShape — V2 (F1.2/A4).
// POST /api/v1/templates/{id}/versions returns 201 + flat TemplateVersion body (no {data:{...}} envelope).
func TestCreateNextVersion_TypedResponseShape(t *testing.T) {
	repo := newFakeRepo()
	const tplID = "11111111-1111-1111-1111-111111111111"
	repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Key: "k", Name: "n", LatestVersion: 1}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     tplID,
		VersionNumber:  1,
		Status:         domain.VersionStatusPublished,
		DocxStorageKey: "templates/" + tplID + "/versions/1.docx",
		ContentHash:    "hash_v1",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions", bytes.NewReader([]byte(`{}`)))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v (raw=%s)", err, rr.Body.String())
	}
	assertFlatTemplateVersionShape(t, raw)
}

// TestPresignAutosave_TypedResponseShape — V3 (F1.2/A5).
// POST /api/v1/templates/{id}/versions/{n}/autosave/presign returns 200 + flat
// TemplatePresignAutosaveResponse (no {data:{...}} envelope).
func TestPresignAutosave_TypedResponseShape(t *testing.T) {
	repo := newFakeRepo()
	const tplID = "11111111-1111-1111-1111-111111111111"
	repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     tplID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/" + tplID + "/versions/1.docx",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/autosave/presign", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 (F1.2/ADR-0035: presign endpoints return 200), got %d body=%s", rr.Code, rr.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v (raw=%s)", err, rr.Body.String())
	}
	if _, ok := raw["data"]; ok {
		t.Fatalf("forbidden top-level key %q — F1.2/ADR-0035 drops {data:{...}} envelope", "data")
	}
	declared := map[string]struct{}{
		"upload_url": {}, "storage_key": {}, "expires_at": {},
	}
	for k := range raw {
		if _, ok := declared[k]; !ok {
			t.Fatalf("unexpected key %q in presign response: keys=%v", k, mapKeysT(raw))
		}
	}
	for k := range declared {
		if _, ok := raw[k]; !ok {
			t.Fatalf("missing required key %q in presign response: keys=%v", k, mapKeysT(raw))
		}
	}
}

// TestCommitAutosave_TypedResponseShape — V3a (F1.2/P2-light).
// POST /api/v1/templates/{id}/versions/{n}/autosave/commit returns 200 + flat TemplateVersion
// (no {data:{version:...}} envelope).
func TestCommitAutosave_TypedResponseShape(t *testing.T) {
	repo := newFakeRepo()
	const tplID = "11111111-1111-1111-1111-111111111111"
	repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     tplID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/" + tplID + "/versions/1.docx",
		ContentHash:    "hash_abc",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	body, _ := json.Marshal(map[string]any{"expected_content_hash": "hash_abc"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/autosave/commit", bytes.NewReader(body))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v (raw=%s)", err, rr.Body.String())
	}
	assertFlatTemplateVersionShape(t, raw)
}

// TestGetTemplateVersion_TypedResponseShape — V3b (F1.2/P2-light).
// GET /api/v1/templates/{id}/versions/{n} returns 200 + flat TemplateVersion
// (no {data:{version:...}} envelope).
func TestGetTemplateVersion_TypedResponseShape(t *testing.T) {
	repo := newFakeRepo()
	const tplID = "11111111-1111-1111-1111-111111111111"
	repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     tplID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/" + tplID + "/versions/1.docx",
		ContentHash:    "hash_v1",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+tplID+"/versions/1", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v (raw=%s)", err, rr.Body.String())
	}
	assertFlatTemplateVersionShape(t, raw)
}

// TestCreateTemplate_TypedResponseShape — V1/V2 (F1.3/A3).
// POST /api/v1/templates (createTemplate) must return 201 + {data:{template:TemplateDTO,version:VersionDTO}}
// with NO undeclared top-level "id" or "version_id" (H-D fix).
// Uses a UUID-format tenant so toAPITemplateDTO can parse TenantId (the mapper calls uuid.Parse).
func TestCreateTemplate_TypedResponseShape(t *testing.T) {
	repo := newFakeRepo()
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	const tenantUUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	body := createBody("shape-f13-key")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
	ctx := tenant.WithTenantID(req.Context(), tenantUUID)
	ctx = iamdomain.WithAuthContext(ctx, "user-shape", []iamdomain.Role{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw: %v (body=%s)", err, rr.Body.String())
	}

	// V1: undeclared top-level fields must be absent.
	if _, ok := raw["id"]; ok {
		t.Error("top-level 'id' must not be present (A3/H-D: field not declared in CreateTemplateResponse)")
	}
	if _, ok := raw["version_id"]; ok {
		t.Error("top-level 'version_id' must not be present (A3/H-D: field not declared in CreateTemplateResponse)")
	}

	// V2: declared envelope structure must be present.
	dataRaw, ok := raw["data"]
	if !ok {
		t.Fatal("missing top-level 'data' key (declared by CreateTemplateResponse)")
	}

	var dataObj struct {
		Template json.RawMessage `json:"template"`
		Version  json.RawMessage `json:"version"`
	}
	if err := json.Unmarshal(dataRaw, &dataObj); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if dataObj.Template == nil {
		t.Fatal("data.template must be present")
	}
	if dataObj.Version == nil {
		t.Fatal("data.version must be present")
	}

	// Smoke fields: data.template must carry id and tenant_id (TemplateDTO required).
	var tpl struct {
		Id       string `json:"id"`
		TenantId string `json:"tenant_id"`
		Key      string `json:"key"`
	}
	if err := json.Unmarshal(dataObj.Template, &tpl); err != nil {
		t.Fatalf("decode data.template: %v", err)
	}
	if tpl.Id == "" {
		t.Error("data.template.id must be non-empty (TemplateDTO required field)")
	}
	if tpl.TenantId != tenantUUID {
		t.Errorf("data.template.tenant_id = %q; want %q", tpl.TenantId, tenantUUID)
	}
	if tpl.Key != "shape-f13-key" {
		t.Errorf("data.template.key = %q; want %q", tpl.Key, "shape-f13-key")
	}

	// Smoke fields: data.version must carry version_number (VersionDTO required).
	var ver struct {
		VersionNumber int `json:"version_number"`
	}
	if err := json.Unmarshal(dataObj.Version, &ver); err != nil {
		t.Fatalf("decode data.version: %v", err)
	}
	if ver.VersionNumber != 1 {
		t.Errorf("data.version.version_number = %d; want 1", ver.VersionNumber)
	}
}
