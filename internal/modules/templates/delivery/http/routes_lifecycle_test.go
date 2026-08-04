package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/templates/domain"
)

func TestArchiveTemplate_Happy(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/archive", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Data struct {
			Template struct {
				ArchivedAt *string `json:"archived_at"`
			} `json:"template"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.Template.ArchivedAt == nil || *out.Data.Template.ArchivedAt == "" {
		t.Fatal("expected data.template.archived_at to be set")
	}
}

func TestArchiveTemplate_SystemOwnedTemplateImmutable(t *testing.T) {
	repo := newFakeRepo()
	templateID := "00000000-0000-0000-0000-000000000101"
	repo.templates[templateID] = &domain.Template{ID: templateID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", SystemOwned: true}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID+"/archive", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "state.system_template_immutable" {
		t.Fatalf("expected error.code=SYSTEM_TEMPLATE_IMMUTABLE, got %q", out.Code)
	}
}

// TestPublishTemplateVersion_SelfPublishForbiddenRFC9457 verifies that POST
// /publish rejects an actor attempting to publish a version they authored,
// returning RFC 9457 problem+json with `code: "permission.iso_segregation_violation"`
// and HTTP 403. This is the surviving identity-based SoD gate (CheckSegregation).
//
// Formerly TestPublishTemplateVersion_ForbiddenRoleRFC9457, which asserted the
// role-binding tier-2 gate (RoleBindingFor(Published) + containsRole). ADR
// 0082 unit 3.1a slice S2 deletes that gate — a capability holder with no
// matching role now publishes successfully (see
// TestPublishTemplateVersion_NoRoles_CapabilityAlone in lifecycle_test.go for
// the application-layer equivalent). No other HTTP-level test in this module
// asserted the identity-SoD 403 shape, so this test is rewritten rather than
// deleted to close that coverage gap.
func TestPublishTemplateVersion_SelfPublishForbiddenRFC9457(t *testing.T) {
	repo := newFakeRepo()
	templateID := "11111111-1111-1111-1111-111111111111"
	repo.templates[templateID] = &domain.Template{ID: templateID, TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", LatestVersion: 1}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     templateID,
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/" + templateID + "/versions/1.docx",
		ContentHash:    "hash_ok",
		AuthorID:       "user-a", // matches withHeaders' actor id below -> self-publish
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID+"/versions/1/publish", nil)
	withHeaders(req) // sets actor "user-a" -- the version's own AuthorID
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "permission.iso_segregation_violation" {
		t.Fatalf("expected error.code=ISO_SEGREGATION_VIOLATION, got %q (body=%s)", out.Code, rr.Body.String())
	}

	stored := repo.versions["22222222-2222-4222-8222-222222222222"]
	if stored.Status != domain.VersionStatusDraft {
		t.Fatalf("expected version status unchanged (draft), got %q", stored.Status)
	}
	if len(repo.audit) != 0 {
		t.Fatalf("expected no audit events on a segregation-blocked publish, got %+v", repo.audit)
	}
}
