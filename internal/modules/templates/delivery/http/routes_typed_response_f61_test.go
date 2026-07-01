package http_test

// Feature F6.1 — typed-shape contract tests for the templates lifecycle 200 responses.
// Each subtest decodes the handler body with json.Decoder + DisallowUnknownFields() into
// the codegen-generated *Response type. Asserts the typed field round-trips non-zero —
// locking the handlers to the OpenAPI declaration so future drift fails at this gate.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/domain"
)

func decodeStrict[T any](t *testing.T, body []byte) T {
	t.Helper()
	var out T
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode strict: %v\nbody: %s", err, string(body))
	}
	return out
}

func TestLifecycle_TypedResponseShape(t *testing.T) {
	const (
		tplID    = "11111111-1111-1111-1111-111111111111"
		tenantID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		verID    = "22222222-2222-4222-8222-222222222222"
	)
	reviewerRole := "reviewer"

	t.Run("submit", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID}
		repo.versions[verID] = &domain.TemplateVersion{
			ID: verID, TemplateID: tplID, VersionNumber: 1,
			Status: domain.VersionStatusDraft, ContentHash: "deadbeef", AuthorID: "author-1",
		}
		repo.approvalConfigs[tplID] = &domain.ApprovalConfig{
			TemplateID: tplID, ReviewerRole: &reviewerRole, ApproverRole: "approver",
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/submit", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.TemplateVersionEnvelope](t, rr.Body.Bytes())
		if string(out.Data.Version.Status) != string(domain.VersionStatusUnderReview) {
			t.Fatalf("status=%q want=%q", out.Data.Version.Status, domain.VersionStatusUnderReview)
		}
	})

	t.Run("review", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID}
		repo.versions[verID] = &domain.TemplateVersion{
			ID: verID, TemplateID: tplID, VersionNumber: 1,
			Status:              domain.VersionStatusUnderReview,
			AuthorID:            "author-1",
			PendingReviewerRole: &reviewerRole,
			PendingApproverRole: "approver",
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		raw, _ := json.Marshal(map[string]any{"accept": true})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/review", bytes.NewReader(raw))
		withHeaders(req)
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "reviewer-1", []iamdomain.Role{}))
		withActorRoles(req, "reviewer")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.TemplateVersionEnvelope](t, rr.Body.Bytes())
		if out.Data.Version.Id.String() == "" {
			t.Fatal("version.id empty")
		}
	})

	t.Run("approve_accept_publish", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID}
		repo.versions[verID] = &domain.TemplateVersion{
			ID: verID, TemplateID: tplID, VersionNumber: 1,
			Status:              domain.VersionStatusApproved,
			ContentHash:         "deadbeef", // F-T2: publish path now gates on a committed docx
			AuthorID:            "author-1",
			PendingReviewerRole: &reviewerRole,
			PendingApproverRole: "approver",
			ReviewerID:          strPtr("reviewer-1"),
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		raw, _ := json.Marshal(map[string]any{"accept": true})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/approve", bytes.NewReader(raw))
		withHeaders(req)
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "approver-1", []iamdomain.Role{}))
		withActorRoles(req, "approver")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.ApproveTemplateVersionResponse](t, rr.Body.Bytes())
		if string(out.Data.Version.Status) != string(domain.VersionStatusPublished) {
			t.Fatalf("status=%q want=published", out.Data.Version.Status)
		}
		// M1·T2: auto next-draft spawn removed; next_draft field is gone from the schema.
		// The strict decode above (DisallowUnknownFields) already guards against next_draft leaking back.
	})

	t.Run("approve_reject", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID}
		repo.versions[verID] = &domain.TemplateVersion{
			ID: verID, TemplateID: tplID, VersionNumber: 1,
			Status:              domain.VersionStatusApproved,
			ContentHash:         "deadbeef", // F-T2: publish path now gates on a committed docx
			AuthorID:            "author-1",
			PendingReviewerRole: &reviewerRole,
			PendingApproverRole: "approver",
			ReviewerID:          strPtr("reviewer-1"),
		}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		raw, _ := json.Marshal(map[string]any{"accept": false, "reason": "needs work"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/versions/1/approve", bytes.NewReader(raw))
		withHeaders(req)
		req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "approver-1", []iamdomain.Role{}))
		withActorRoles(req, "approver")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.ApproveTemplateVersionResponse](t, rr.Body.Bytes())
		if string(out.Data.Version.Status) != string(domain.VersionStatusDraft) {
			t.Fatalf("status=%q want=draft on reject", out.Data.Version.Status)
		}
		// M1·T2: next_draft field removed from schema; strict decode guards against leaks.
	})

	t.Run("archive", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+tplID+"/archive", nil)
		withHeaders(req)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.ArchiveTemplateResponse](t, rr.Body.Bytes())
		if out.Data.Template.ArchivedAt == nil || out.Data.Template.ArchivedAt.IsZero() {
			t.Fatal("template.archived_at not set")
		}
	})

	t.Run("upsert_approval_config", func(t *testing.T) {
		repo := newFakeRepo()
		repo.templates[tplID] = &domain.Template{ID: tplID, TenantID: tenantID, CreatedBy: "user-a"}
		mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

		raw, _ := json.Marshal(map[string]any{
			"reviewer_role": "reviewer",
			"approver_role": "approver",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+tplID+"/approval-config", bytes.NewReader(raw))
		withHeaders(req)
		withActorRoles(req, "editor")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		out := decodeStrict[templatesapi.UpsertTemplateApprovalConfigResponse](t, rr.Body.Bytes())
		if out.Data.ApprovalConfig.ApproverRole != "approver" {
			t.Fatalf("approver_role=%q want=approver", out.Data.ApprovalConfig.ApproverRole)
		}
		if out.Data.ApprovalConfig.ReviewerRole == nil || *out.Data.ApprovalConfig.ReviewerRole != "reviewer" {
			t.Fatalf("reviewer_role=%v want=reviewer", out.Data.ApprovalConfig.ReviewerRole)
		}
		if out.Data.ApprovalConfig.TemplateId.String() != tplID {
			t.Fatalf("template_id=%q want=%q", out.Data.ApprovalConfig.TemplateId.String(), tplID)
		}
	})
}
