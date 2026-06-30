package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/domain"
)

// TestPresignTemplateDocxUploadUrl_TypedResponseShape — F5.3 H-D pin.
// POST /api/v1/templates/{id}/versions/{n}/upload/docx returns 200 + flat
// {url, storage_key} matching templatesapi.PresignTemplateDocxUploadUrl200JSONResponse.
// No envelope, no PascalCase, no other keys.
func TestPresignTemplateDocxUploadUrl_TypedResponseShape(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/docx-upload-url", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	assertF53PresignShape(t, rr.Body.Bytes())
}

// TestPresignTemplateSchemaUploadUrl_TypedResponseShape — F5.3 H-D pin (schema sibling).
func TestPresignTemplateSchemaUploadUrl_TypedResponseShape(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/schema-upload-url", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	assertF53PresignShape(t, rr.Body.Bytes())
}

func assertF53PresignShape(t *testing.T, body []byte) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decode body: %v (raw=%s)", err, string(body))
	}
	if _, ok := raw["data"]; ok {
		t.Fatalf("forbidden top-level key %q — presign upload-url returns flat body", "data")
	}
	declared := map[string]struct{}{"url": {}, "storage_key": {}}
	for k := range raw {
		if _, ok := declared[k]; !ok {
			t.Fatalf("unexpected key %q in presign upload-url response: keys=%v (must match PresignTemplateDocxUploadUrl200JSONResponse)", k, mapKeysT(raw))
		}
	}
	for k := range declared {
		if _, ok := raw[k]; !ok {
			t.Fatalf("missing required key %q in presign upload-url response: keys=%v", k, mapKeysT(raw))
		}
	}
}

// TestPublishTemplateVersion200JSONResponse_DeclaredKeysOnly — F5.3 H-D pin (publish).
// Pins the strict-server generated type's JSON shape: exactly the 1 field declared
// at openapi.yaml (published_version_id only — next_draft_* removed in M1·T2).
// `published_version_number` and `next_draft_*` must NOT appear.
//
// Compile-time guarantee: the handler at routes_generated.go writes this
// strict-server struct directly, so leaked fields cannot be re-introduced
// without first amending openapi.yaml + re-generating api.gen.go.
func TestPublishTemplateVersion200JSONResponse_DeclaredKeysOnly(t *testing.T) {
	resp := templatesapi.PublishTemplateVersion200JSONResponse{
		PublishedVersionId: "11111111-1111-1111-1111-111111111111",
	}
	encoded, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal PublishTemplateVersion200JSONResponse: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("decode encoded body: %v (raw=%s)", err, string(encoded))
	}
	declared := map[string]struct{}{
		"published_version_id": {},
	}
	for k := range raw {
		if _, ok := declared[k]; !ok {
			t.Fatalf("unexpected key %q in publish 200 body: keys=%v (must match openapi.yaml)", k, mapKeysT(raw))
		}
	}
	for k := range declared {
		if _, ok := raw[k]; !ok {
			t.Fatalf("missing required key %q in publish 200 body: keys=%v", k, mapKeysT(raw))
		}
	}
	if _, leaked := raw["published_version_number"]; leaked {
		t.Fatalf("undeclared key %q leaked into publish 200 body — F5.3/M1.F1.3 regression", "published_version_number")
	}
	if _, leaked := raw["next_draft_id"]; leaked {
		t.Fatalf("removed key %q must not appear in publish 200 body — M1.T2 regression", "next_draft_id")
	}
	if _, leaked := raw["next_draft_version_num"]; leaked {
		t.Fatalf("removed key %q must not appear in publish 200 body — M1.T2 regression", "next_draft_version_num")
	}
}
