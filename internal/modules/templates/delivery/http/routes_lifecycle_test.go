package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
)

func strPtr(v string) *string { return &v }

func withActorRoles(req *http.Request, roles string) {
	parts := strings.Split(roles, ",")
	ctxRoles := make([]iamdomain.Role, 0, len(parts))
	for _, role := range parts {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		ctxRoles = append(ctxRoles, iamdomain.Role(role))
	}
	*req = *req.WithContext(iamdomain.WithAuthContext(req.Context(), iamdomain.UserIDFromContext(req.Context()), ctxRoles))
}

func TestSubmitForReview_Happy(t *testing.T) {
	repo := newFakeRepo()
	reviewerRole := "reviewer"
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    "11111111-1111-1111-1111-111111111111",
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "deadbeef",
		AuthorID:      "author-1",
	}
	repo.approvalConfigs["11111111-1111-1111-1111-111111111111"] = &domain.ApprovalConfig{
		TemplateID:   "11111111-1111-1111-1111-111111111111",
		ReviewerRole: &reviewerRole,
		ApproverRole: "approver",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/submit", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Data struct {
			Version struct {
				Status string `json:"status"`
			} `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.Version.Status != string(domain.VersionStatusInReview) {
		t.Fatalf("expected status=in_review, got %q", out.Data.Version.Status)
	}
}

func TestSubmitForReview_UsesTemplateSubmitCapability(t *testing.T) {
	repo := newFakeRepo()
	reviewerRole := "reviewer"
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    "11111111-1111-1111-1111-111111111111",
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "deadbeef",
		AuthorID:      "author-1",
	}
	repo.approvalConfigs["11111111-1111-1111-1111-111111111111"] = &domain.ApprovalConfig{
		TemplateID:   "11111111-1111-1111-1111-111111111111",
		ReviewerRole: &reviewerRole,
		ApproverRole: "approver",
	}

	var gotAction string
	mux := newMux(t, func(_ *http.Request, _, _, action string) error {
		gotAction = action
		return nil
	}, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/submit", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if gotAction != "template.submit" {
		t.Fatalf("authz action = %q, want template.submit", gotAction)
	}
}

func TestSubmitForReview_NonDraft(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    "11111111-1111-1111-1111-111111111111",
		VersionNumber: 1,
		Status:        domain.VersionStatusInReview,
	}
	repo.approvalConfigs["11111111-1111-1111-1111-111111111111"] = &domain.ApprovalConfig{TemplateID: "11111111-1111-1111-1111-111111111111", ApproverRole: "approver"}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/submit", nil)
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
	if out.Code != "invalid_state_transition" {
		t.Fatalf("expected error.code=invalid_state_transition, got %q", out.Code)
	}
}

func TestSubmitForReview_NoUpload(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    "11111111-1111-1111-1111-111111111111",
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "", // no autosave committed
	}
	repo.approvalConfigs["11111111-1111-1111-1111-111111111111"] = &domain.ApprovalConfig{TemplateID: "11111111-1111-1111-1111-111111111111", ApproverRole: "approver"}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/submit", nil)
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
	if out.Code != "upload_missing" {
		t.Fatalf("expected error.code=upload_missing, got %q", out.Code)
	}
}

func TestSubmitForReview_SystemOwnedTemplateImmutable(t *testing.T) {
	repo := newFakeRepo()
	templateID := "00000000-0000-0000-0000-000000000101"
	repo.templates[templateID] = &domain.Template{
		ID:          templateID,
		TenantID:    "tenant-a",
		SystemOwned: true,
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    templateID,
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "deadbeef",
	}
	repo.approvalConfigs[templateID] = &domain.ApprovalConfig{TemplateID: templateID, ApproverRole: "approver"}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID+"/versions/1/submit", nil)
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
	if out.Code != "SYSTEM_TEMPLATE_IMMUTABLE" {
		t.Fatalf("expected error.code=SYSTEM_TEMPLATE_IMMUTABLE, got %q", out.Code)
	}
}

func TestReview_Accept_Happy(t *testing.T) {
	repo := newFakeRepo()
	reviewerRole := "reviewer"
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          "11111111-1111-1111-1111-111111111111",
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	raw, _ := json.Marshal(map[string]any{"accept": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/review", bytes.NewReader(raw))
	withHeaders(req)
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "reviewer-1", []iamdomain.Role{}))
	withActorRoles(req, "reviewer")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Data struct {
			Version struct {
				Status string `json:"status"`
			} `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.Version.Status != string(domain.VersionStatusApproved) {
		t.Fatalf("expected status=approved, got %q", out.Data.Version.Status)
	}
}

func TestApprove_Accept_Happy(t *testing.T) {
	repo := newFakeRepo()
	reviewerRole := "reviewer"
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          "11111111-1111-1111-1111-111111111111",
		VersionNumber:       1,
		Status:              domain.VersionStatusApproved,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		PendingApproverRole: "approver",
		ReviewerID:          strPtr("reviewer-1"),
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	raw, _ := json.Marshal(map[string]any{"accept": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/approve", bytes.NewReader(raw))
	withHeaders(req)
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "approver-1", []iamdomain.Role{}))
	withActorRoles(req, "approver")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Data struct {
			Version struct {
				Status string `json:"status"`
			} `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.Version.Status != string(domain.VersionStatusPublished) {
		t.Fatalf("expected status=published, got %q", out.Data.Version.Status)
	}
}

func TestArchiveTemplate_Happy(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a"}
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
	repo.templates[templateID] = &domain.Template{ID: templateID, TenantID: "tenant-a", SystemOwned: true}
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
	if out.Code != "SYSTEM_TEMPLATE_IMMUTABLE" {
		t.Fatalf("expected error.code=SYSTEM_TEMPLATE_IMMUTABLE, got %q", out.Code)
	}
}

// TestPublishTemplateVersion_ForbiddenRoleRFC9457 verifies that POST /publish
// rejects an actor who holds the template.publish capability but lacks the
// version's PendingApproverRole binding, returning RFC 9457 problem+json with
// `code: "forbidden_role"` and HTTP 403. Closes residual T-004 contract gap.
func TestPublishTemplateVersion_ForbiddenRoleRFC9457(t *testing.T) {
	repo := newFakeRepo()
	templateID := "11111111-1111-1111-1111-111111111111"
	repo.templates[templateID] = &domain.Template{ID: templateID, TenantID: "tenant-a", LatestVersion: 1}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          templateID,
		VersionNumber:       1,
		Status:              domain.VersionStatusDraft,
		DocxStorageKey:      "templates/" + templateID + "/versions/1.docx",
		ContentHash:         "hash_ok",
		AuthorID:            "author-1",
		PendingApproverRole: "approver",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	body, _ := json.Marshal(map[string]any{
		"docx_key":   "templates/" + templateID + "/versions/1.docx",
		"schema_key": "templates/" + templateID + "/versions/1.schema.json",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID+"/versions/1/publish", bytes.NewReader(body))
	withHeaders(req)
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "publisher-1", []iamdomain.Role{}))
	withActorRoles(req, "editor") // capability holder, NOT the required approver role
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
	if out.Code != "forbidden_role" {
		t.Fatalf("expected error.code=forbidden_role, got %q (body=%s)", out.Code, rr.Body.String())
	}

	stored := repo.versions["ver-1"]
	if stored.Status != domain.VersionStatusDraft {
		t.Fatalf("expected version status unchanged (draft), got %q", stored.Status)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditPublishForbiddenRole {
		t.Fatalf("expected one publish_forbidden_role audit event, got %+v", repo.audit)
	}
}

func TestUpsertApprovalConfig_Happy(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-a", CreatedBy: "user-a"}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	raw, _ := json.Marshal(map[string]any{
		"reviewer_role": "reviewer",
		"approver_role": "approver",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/11111111-1111-1111-1111-111111111111/approval-config", bytes.NewReader(raw))
	withHeaders(req)
	withActorRoles(req, "editor")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Data struct {
			ApprovalConfig struct {
				ApproverRole string `json:"approver_role"`
			} `json:"approval_config"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.ApprovalConfig.ApproverRole != "approver" {
		t.Fatalf("expected data.approval_config.approver_role=approver, got %q", out.Data.ApprovalConfig.ApproverRole)
	}
}
