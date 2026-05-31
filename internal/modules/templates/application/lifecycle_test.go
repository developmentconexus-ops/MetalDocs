package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

func TestSubmitForReview_Happy(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	version := &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "deadbeef",
		AuthorID:      "author-1",
	}
	reviewerRole := "reviewer"
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	repo.approvalConfigs[template.ID] = &domain.ApprovalConfig{
		TemplateID:   template.ID,
		ReviewerRole: &reviewerRole,
		ApproverRole: "approver",
	}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.SubmitForReview(context.Background(), application.SubmitForReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "author-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("SubmitForReview returned error: %v", err)
	}
	if got.Status != domain.VersionStatusInReview {
		t.Fatalf("expected status %q, got %q", domain.VersionStatusInReview, got.Status)
	}
	if got.SubmittedAt == nil {
		t.Fatal("expected SubmittedAt to be set")
	}
	if got.PendingReviewerRole == nil || *got.PendingReviewerRole != reviewerRole {
		t.Fatalf("expected pending reviewer role %q, got %v", reviewerRole, got.PendingReviewerRole)
	}
	if got.PendingApproverRole != "approver" {
		t.Fatalf("expected pending approver role %q, got %q", "approver", got.PendingApproverRole)
	}
	if len(repo.audit) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(repo.audit))
	}
	if repo.audit[0].Action != domain.AuditSubmitted {
		t.Fatalf("expected audit action %q, got %q", domain.AuditSubmitted, repo.audit[0].Action)
	}
	if repo.audit[0].Details["approver_role"] != "approver" {
		t.Fatalf("expected approver_role detail %q, got %v", "approver", repo.audit[0].Details["approver_role"])
	}
	if repo.audit[0].Details["reviewer_role"] != &reviewerRole {
		t.Fatalf("expected reviewer_role detail to be configured pointer")
	}
}

func TestSubmitForReview_NonDraft(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	version := &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
		Status:        domain.VersionStatusInReview,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	repo.approvalConfigs[template.ID] = &domain.ApprovalConfig{
		TemplateID:   template.ID,
		ApproverRole: "approver",
	}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.SubmitForReview(context.Background(), application.SubmitForReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "author-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestSubmitForReview_NoUpload(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	version := &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	repo.approvalConfigs[template.ID] = &domain.ApprovalConfig{
		TemplateID:   template.ID,
		ApproverRole: "approver",
	}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.SubmitForReview(context.Background(), application.SubmitForReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "author-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
	})
	if !errors.Is(err, domain.ErrUploadMissing) {
		t.Fatalf("expected ErrUploadMissing, got %v", err)
	}
}

func TestReview_Accept(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	submittedAt := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		SubmittedAt:         &submittedAt,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.Review(context.Background(), application.ReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "reviewer-1",
		ActorRoles:    []string{"reviewer"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if got.Status != domain.VersionStatusApproved {
		t.Fatalf("expected status %q, got %q", domain.VersionStatusApproved, got.Status)
	}
	if got.ReviewerID == nil || *got.ReviewerID != "reviewer-1" {
		t.Fatalf("expected reviewer id %q, got %v", "reviewer-1", got.ReviewerID)
	}
	if got.ReviewedAt == nil {
		t.Fatal("expected ReviewedAt to be set")
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditReviewed {
		t.Fatalf("expected one %q audit event, got %v", domain.AuditReviewed, repo.audit)
	}
}

func TestReview_Reject(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	submittedAt := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		SubmittedAt:         &submittedAt,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.Review(context.Background(), application.ReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "reviewer-1",
		ActorRoles:    []string{"reviewer"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        false,
		Reason:        "changes requested",
	})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if got.Status != domain.VersionStatusDraft {
		t.Fatalf("expected status %q, got %q", domain.VersionStatusDraft, got.Status)
	}
	if got.SubmittedAt != nil {
		t.Fatalf("expected SubmittedAt to be cleared, got %v", got.SubmittedAt)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditRejected {
		t.Fatalf("expected one %q audit event, got %v", domain.AuditRejected, repo.audit)
	}
	if repo.audit[0].Details["reason"] != "changes requested" {
		t.Fatalf("expected reason detail %q, got %v", "changes requested", repo.audit[0].Details["reason"])
	}
	if repo.audit[0].Details["stage"] != "reviewer" {
		t.Fatalf("expected stage detail %q, got %v", "reviewer", repo.audit[0].Details["stage"])
	}
}

func TestReview_WrongRole(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.Review(context.Background(), application.ReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "reviewer-1",
		ActorRoles:    []string{"not-reviewer"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if !errors.Is(err, domain.ErrForbiddenRole) {
		t.Fatalf("expected ErrForbiddenRole, got %v", err)
	}
}

func TestReview_SegregationViolation(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.Review(context.Background(), application.ReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "author-1",
		ActorRoles:    []string{"reviewer"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if !errors.Is(err, domain.ErrISOSegregationViolation) {
		t.Fatalf("expected ErrISOSegregationViolation, got %v", err)
	}
}

func TestReview_NoReviewerStage(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	version := &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    template.ID,
		VersionNumber: 1,
		Status:        domain.VersionStatusInReview,
		AuthorID:      "author-1",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.Review(context.Background(), application.ReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "reviewer-1",
		ActorRoles:    []string{"reviewer"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestApprove_Accept_WithReviewer(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{
		ID:                 "tpl-1",
		TenantID:           "tenant-a",
		PublishedVersionID: strPtr("ver-old"),
	}
	reviewerRole := "reviewer"
	oldPublished := &domain.TemplateVersion{
		ID:            "ver-old",
		TemplateID:    template.ID,
		VersionNumber: 1,
		Status:        domain.VersionStatusPublished,
		AuthorID:      "author-0",
	}
	version := &domain.TemplateVersion{
		ID:                  "ver-2",
		TemplateID:          template.ID,
		VersionNumber:       2,
		Status:              domain.VersionStatusApproved,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		PendingApproverRole: "approver",
		ReviewerID:          strPtr("reviewer-1"),
	}
	repo.templates[template.ID] = template
	repo.versions[oldPublished.ID] = oldPublished
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	res, err := svc.Approve(context.Background(), application.ApproveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		ActorRoles:    []string{"approver"},
		TemplateID:    template.ID,
		VersionNumber: 2,
		Accept:        true,
	})
	if err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}
	got := res.Version
	if got.Status != domain.VersionStatusPublished {
		t.Fatalf("expected status %q, got %q", domain.VersionStatusPublished, got.Status)
	}
	if got.ApproverID == nil || *got.ApproverID != "approver-1" {
		t.Fatalf("expected approver id %q, got %v", "approver-1", got.ApproverID)
	}
	if got.ApprovedAt == nil || got.PublishedAt == nil {
		t.Fatal("expected ApprovedAt and PublishedAt to be set")
	}
	if oldPublished.ObsoletedAt == nil {
		t.Fatal("expected previously published version to be obsoleted")
	}
	if template.PublishedVersionID == nil || *template.PublishedVersionID != version.ID {
		t.Fatalf("expected PublishedVersionID %q, got %v", version.ID, template.PublishedVersionID)
	}
	if template.PublishedVersionNumber == nil || *template.PublishedVersionNumber != version.VersionNumber {
		t.Fatalf("expected PublishedVersionNumber %d, got %v", version.VersionNumber, template.PublishedVersionNumber)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditPublished {
		t.Fatalf("expected one %q audit event, got %v", domain.AuditPublished, repo.audit)
	}
	if res.NextDraft == nil {
		t.Fatal("expected NextDraft to be populated on approve-publish")
	}
	if res.NextDraft.VersionNumber != version.VersionNumber+1 {
		t.Fatalf("expected NextDraft.VersionNumber %d, got %d", version.VersionNumber+1, res.NextDraft.VersionNumber)
	}
	if res.NextDraft.Status != domain.VersionStatusDraft {
		t.Fatalf("expected NextDraft.Status %q, got %q", domain.VersionStatusDraft, res.NextDraft.Status)
	}
	if template.LatestVersion != res.NextDraft.VersionNumber {
		t.Fatalf("expected template.LatestVersion %d, got %d", res.NextDraft.VersionNumber, template.LatestVersion)
	}
}

func TestApprove_Accept_NoReviewer(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: nil,
		PendingApproverRole: "approver",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	res, err := svc.Approve(context.Background(), application.ApproveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		ActorRoles:    []string{"approver"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}
	got := res.Version
	if got.Status != domain.VersionStatusPublished {
		t.Fatalf("expected status %q, got %q", domain.VersionStatusPublished, got.Status)
	}
	if template.PublishedVersionID == nil || *template.PublishedVersionID != version.ID {
		t.Fatalf("expected PublishedVersionID %q, got %v", version.ID, template.PublishedVersionID)
	}
	if template.PublishedVersionNumber == nil || *template.PublishedVersionNumber != version.VersionNumber {
		t.Fatalf("expected PublishedVersionNumber %d, got %v", version.VersionNumber, template.PublishedVersionNumber)
	}
	if res.NextDraft == nil {
		t.Fatal("expected NextDraft to be populated on approve-publish (no-reviewer path)")
	}
	if res.NextDraft.VersionNumber != version.VersionNumber+1 {
		t.Fatalf("expected NextDraft.VersionNumber %d, got %d", version.VersionNumber+1, res.NextDraft.VersionNumber)
	}
	if res.NextDraft.Status != domain.VersionStatusDraft {
		t.Fatalf("expected NextDraft.Status %q, got %q", domain.VersionStatusDraft, res.NextDraft.Status)
	}
	if template.LatestVersion != res.NextDraft.VersionNumber {
		t.Fatalf("expected template.LatestVersion %d, got %d", res.NextDraft.VersionNumber, template.LatestVersion)
	}
}

func TestApprove_Reject(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	submittedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	reviewedAt := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	approvedAt := time.Date(2026, 4, 20, 11, 30, 0, 0, time.UTC)
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusApproved,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		PendingApproverRole: "approver",
		ReviewerID:          strPtr("reviewer-1"),
		SubmittedAt:         &submittedAt,
		ReviewedAt:          &reviewedAt,
		ApprovedAt:          &approvedAt,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	res, err := svc.Approve(context.Background(), application.ApproveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		ActorRoles:    []string{"approver"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        false,
		Reason:        "missing legal section",
	})
	if err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}
	got := res.Version
	if res.NextDraft != nil {
		t.Fatalf("expected NextDraft nil on reject, got %v", res.NextDraft)
	}
	if got.Status != domain.VersionStatusDraft {
		t.Fatalf("expected status %q, got %q", domain.VersionStatusDraft, got.Status)
	}
	if got.SubmittedAt != nil || got.ReviewedAt != nil || got.ApprovedAt != nil {
		t.Fatalf("expected SubmittedAt, ReviewedAt and ApprovedAt to be cleared, got submitted=%v reviewed=%v approved=%v", got.SubmittedAt, got.ReviewedAt, got.ApprovedAt)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditRejected {
		t.Fatalf("expected one %q audit event, got %v", domain.AuditRejected, repo.audit)
	}
	if repo.audit[0].Details["reason"] != "missing legal section" {
		t.Fatalf("expected reason detail %q, got %v", "missing legal section", repo.audit[0].Details["reason"])
	}
	if repo.audit[0].Details["stage"] != "approver" {
		t.Fatalf("expected stage detail %q, got %v", "approver", repo.audit[0].Details["stage"])
	}
}

func TestApprove_WrongRole(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingApproverRole: "approver",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.Approve(context.Background(), application.ApproveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		ActorRoles:    []string{"not-approver"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if !errors.Is(err, domain.ErrForbiddenRole) {
		t.Fatalf("expected ErrForbiddenRole, got %v", err)
	}
}

func TestApprove_SegregationViolation(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusApproved,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		PendingApproverRole: "approver",
		ReviewerID:          strPtr("reviewer-1"),
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.Approve(context.Background(), application.ApproveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "reviewer-1",
		ActorRoles:    []string{"approver"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        true,
	})
	if !errors.Is(err, domain.ErrISOSegregationViolation) {
		t.Fatalf("expected ErrISOSegregationViolation, got %v", err)
	}
}

func TestArchiveTemplate_Happy(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[template.ID] = template

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.ArchiveTemplate(context.Background(), application.ArchiveCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-1",
		TemplateID:  template.ID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate returned error: %v", err)
	}
	if got.ArchivedAt == nil {
		t.Fatal("expected ArchivedAt to be set")
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditArchived {
		t.Fatalf("expected one %q audit event, got %v", domain.AuditArchived, repo.audit)
	}
	if repo.audit[0].VersionID != nil {
		t.Fatalf("expected nil version id in archive audit, got %v", repo.audit[0].VersionID)
	}
}

func TestArchiveTemplate_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	archivedAt := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	template := &domain.Template{
		ID:         "tpl-1",
		TenantID:   "tenant-a",
		ArchivedAt: &archivedAt,
	}
	repo.templates[template.ID] = template

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.ArchiveTemplate(context.Background(), application.ArchiveCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-1",
		TemplateID:  template.ID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate returned error: %v", err)
	}
	if got.ArchivedAt == nil || !got.ArchivedAt.Equal(archivedAt) {
		t.Fatalf("expected ArchivedAt to remain %v, got %v", archivedAt, got.ArchivedAt)
	}
	if len(repo.audit) != 0 {
		t.Fatalf("expected no audit events for idempotent archive, got %d", len(repo.audit))
	}
}

func TestReview_UsesTemplateReviewCapability(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithDB(db)
	mock.ExpectBegin()
	expectTemplateAuthz(mock, "reviewer-1", "tenant-a", "template.review")
	mock.ExpectCommit()

	_, err = svc.Review(context.Background(), application.ReviewCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "reviewer-1",
		ActorRoles:    []string{"reviewer"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        false,
		Reason:        "changes requested",
	})
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestArchiveTemplate_UsesTemplateArchiveCapability(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.templates[template.ID] = template

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithDB(db)
	mock.ExpectBegin()
	expectTemplateAuthz(mock, "user-1", "tenant-a", "template.archive")
	mock.ExpectCommit()

	_, err = svc.ArchiveTemplate(context.Background(), application.ArchiveCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-1",
		TemplateID:  template.ID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestApproveRejectReturnsStaleLockVersionWhenVersionMoved(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	reviewerRole := "reviewer"
	submittedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	reviewedAt := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	approvedAt := time.Date(2026, 4, 20, 11, 30, 0, 0, time.UTC)
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          template.ID,
		VersionNumber:       1,
		Status:              domain.VersionStatusApproved,
		AuthorID:            "author-1",
		PendingReviewerRole: &reviewerRole,
		PendingApproverRole: "approver",
		ReviewerID:          strPtr("reviewer-1"),
		SubmittedAt:         &submittedAt,
		ReviewedAt:          &reviewedAt,
		ApprovedAt:          &approvedAt,
		LockVersion:         0,
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	repo.lockVersions[version.ID] = 1

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.Approve(context.Background(), application.ApproveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		ActorRoles:    []string{"approver"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		Accept:        false,
		Reason:        "stale review",
	})
	if !errors.Is(err, domain.ErrStaleLockVersion) {
		t.Fatalf("expected ErrStaleLockVersion, got %v", err)
	}
}

func TestPublishTemplateVersion_RollbackOnNextDraftFailure(t *testing.T) {
	repo := newFakeRepo()
	repo.failCreateVersion = true
	template := &domain.Template{
		ID:            "tpl-1",
		TenantID:      "tenant-a",
		LatestVersion: 1,
	}
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          "tpl-1",
		VersionNumber:       1,
		Status:              domain.VersionStatusDraft,
		DocxStorageKey:      "templates/tpl-1/versions/1.docx",
		ContentHash:         "hash_ok",
		AuthorID:            "author-1",
		PendingApproverRole: "approver",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.PublishTemplateVersion(context.Background(), application.PublishTemplateVersionCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "approver-1",
		ActorRoles:    []string{"approver"},
		TemplateID:    template.ID,
		VersionNumber: 1,
		DocxKey:       "templates/tpl-1/versions/1.docx",
		SchemaKey:     "templates/tpl-1/versions/1.schema.json",
	})
	if err == nil {
		t.Fatalf("expected publish to fail when creating next draft fails")
	}
	if !errors.Is(err, domain.ErrTransactionRequired) {
		t.Fatalf("expected ErrTransactionRequired, got %v", err)
	}
	if version.Status != domain.VersionStatusDraft {
		t.Fatalf("expected version status to remain draft, got %q", version.Status)
	}
}
