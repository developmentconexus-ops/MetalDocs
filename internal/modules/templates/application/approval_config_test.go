package application_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

func TestUpsertApprovalConfig_Happy_NeverPublished_AsAuthor(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1"}
	repo.templates[tpl.ID] = tpl
	reviewerRole := "reviewer"

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "author-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"editor"},
		ReviewerRole: &reviewerRole,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("UpsertApprovalConfig returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil config")
	}
	if got.TemplateID != tpl.ID {
		t.Fatalf("expected TemplateID %q, got %q", tpl.ID, got.TemplateID)
	}
	if got.ReviewerRole == nil || *got.ReviewerRole != reviewerRole {
		t.Fatalf("expected ReviewerRole %q, got %v", reviewerRole, got.ReviewerRole)
	}
	if got.ApproverRole != "approver" {
		t.Fatalf("expected ApproverRole %q, got %q", "approver", got.ApproverRole)
	}
	stored := repo.approvalConfigs[tpl.ID]
	if stored == nil || stored.ApproverRole != "approver" {
		t.Fatalf("expected stored approval config with approver role %q", "approver")
	}
	if len(repo.audit) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(repo.audit))
	}
	if repo.audit[0].Action != domain.AuditApprovalConfigUpdated {
		t.Fatalf("expected audit action %q, got %q", domain.AuditApprovalConfigUpdated, repo.audit[0].Action)
	}
	if repo.audit[0].Details["approver_role"] != "approver" {
		t.Fatalf("expected approver_role %q, got %v", "approver", repo.audit[0].Details["approver_role"])
	}
	reviewerDetail, ok := repo.audit[0].Details["reviewer_role"].(*string)
	if !ok || reviewerDetail == nil || *reviewerDetail != reviewerRole {
		t.Fatalf("expected reviewer_role detail %q, got %v", reviewerRole, repo.audit[0].Details["reviewer_role"])
	}
}

func TestUpsertApprovalConfig_Happy_NeverPublished_AsAdmin(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1"}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "admin-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"admin"},
		ReviewerRole: nil,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("UpsertApprovalConfig returned error: %v", err)
	}
	if got.ReviewerRole != nil {
		t.Fatalf("expected nil ReviewerRole, got %v", got.ReviewerRole)
	}
}

func TestUpsertApprovalConfig_Happy_EverPublished_AsAdmin(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1", PublishedVersionID: strPtr("ver-1")}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "admin-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"admin"},
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("UpsertApprovalConfig returned error: %v", err)
	}
}

func TestUpsertApprovalConfig_Forbidden_NeverPublished_NonAuthor_NonAdmin(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1"}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "user-2",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"editor"},
		ApproverRole: "approver",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpsertApprovalConfig_Forbidden_EverPublished_NonAdmin(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1", PublishedVersionID: strPtr("ver-1")}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "author-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"editor"},
		ApproverRole: "approver",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpsertApprovalConfig_Archived(t *testing.T) {
	repo := newFakeRepo()
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1", ArchivedAt: &archivedAt}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "admin-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"admin"},
		ApproverRole: "approver",
	})
	if !errors.Is(err, domain.ErrArchived) {
		t.Fatalf("expected ErrArchived, got %v", err)
	}
}

func TestUpsertApprovalConfig_EmptyApproverRole(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1"}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "author-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"editor"},
		ApproverRole: "",
	})
	if !errors.Is(err, domain.ErrInvalidApprovalConfig) {
		t.Fatalf("expected ErrInvalidApprovalConfig, got %v", err)
	}
}

func TestUpsertApprovalConfig_WithDB_UsesTxAndAuthz(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	tpl := &domain.Template{ID: "tpl-1", TenantID: "tenant-a", CreatedBy: "author-1"}
	repo.templates[tpl.ID] = tpl
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithDB(db)

	mock.ExpectBegin()
	expectTemplateEditAuthzConfig(mock, "author-1", "tenant-a")
	mock.ExpectCommit()

	_, err = svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "author-1",
		TemplateID:   tpl.ID,
		ActorRoles:   []string{"editor"},
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("UpsertApprovalConfig returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func expectTemplateEditAuthzConfig(mock sqlmock.Sqlmock, actorID, tenantID string) {
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.tenant_id', $1, true)")).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.actor_id', $1, true)")).
		WithArgs(actorID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(actorID))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(tenantID))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles
   WHERE user_id   = $1
     AND tenant_id = $2::uuid
     AND role_code = 'system_admin'
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(`[{"area":"tenant","cap":"template.edit"}]`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}
