package application_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	iamauthz "metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

// TestUpsertApprovalConfig_ManageHolder_PublishedTemplate verifies a holder of
// template.manage can edit the approval config of a published template (case a).
func TestUpsertApprovalConfig_ManageHolder_PublishedTemplate(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:                 "tpl-1",
		TenantID:           "tenant-a",
		CreatedBy:          "author-1",
		PublishedVersionID: strPtr("ver-1"),
	}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).
		WithRunner(newTxRunner(newPermissiveMockDB(t)))

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "qms-1",
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("template.manage holder on published template: expected nil error, got %v", err)
	}
}

// TestUpsertApprovalConfig_NonHolder_PublishedTemplate verifies that an actor
// without template.manage gets ErrCapDenied on a published template (case b).
func TestUpsertApprovalConfig_NonHolder_PublishedTemplate(t *testing.T) {
	actorID, tenantID := "editor-1", "tenant-a"
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:                 "tpl-1",
		TenantID:           tenantID,
		CreatedBy:          "author-1",
		PublishedVersionID: strPtr("ver-1"),
	}
	repo.templates[tpl.ID] = tpl
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).
		WithRunner(newTxRunner(sqlDB))

	mock.ExpectBegin()
	// SeedTxIdentity
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	// MustActorID GUC
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(actorID))
	// MustTenantID GUC
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(tenantID))
	// system_admin bypass → false
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// template.edit cap grant → true (so the base gate passes)
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// appendAssertedCap for template.edit
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(""))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	// MustActorID GUC (second Require call)
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(actorID))
	// MustTenantID GUC
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(tenantID))
	// system_admin bypass → false (not admin)
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// template.manage cap grant → false (not granted)
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectRollback()

	_, gotErr := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     tenantID,
		ActorUserID:  actorID,
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})

	var denied iamauthz.ErrCapDenied
	if !errors.As(gotErr, &denied) {
		t.Fatalf("expected ErrCapDenied for non-template.manage holder on published template, got %T: %v", gotErr, gotErr)
	}
	if denied.Capability != string(iamdomain.CapTemplateManage) {
		t.Fatalf("expected ErrCapDenied for cap %q, got %q", iamdomain.CapTemplateManage, denied.Capability)
	}
}

// TestUpsertApprovalConfig_Creator_UnpublishedTemplate verifies the creator
// of an unpublished template can configure it without template.manage (case c).
// The permissive DB bypasses via system_admin → creator shortcut is not tested
// via SQL denial; domain-ownership means requireManage=false → no manage check.
func TestUpsertApprovalConfig_Creator_UnpublishedTemplate(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:        "tpl-1",
		TenantID:  "tenant-a",
		CreatedBy: "author-1",
	}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).
		WithRunner(newTxRunner(newPermissiveMockDB(t)))

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "author-1",
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("creator on unpublished template: expected nil error, got %v", err)
	}
}

// TestUpsertApprovalConfig_NonCreatorNonHolder_UnpublishedTemplate verifies that
// a non-creator without template.manage is denied on an unpublished template (case d).
func TestUpsertApprovalConfig_NonCreatorNonHolder_UnpublishedTemplate(t *testing.T) {
	actorID, tenantID := "editor-1", "tenant-a"
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:        "tpl-1",
		TenantID:  tenantID,
		CreatedBy: "author-1", // actor is NOT the creator
	}
	repo.templates[tpl.ID] = tpl
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).
		WithRunner(newTxRunner(sqlDB))

	mock.ExpectBegin()
	// SeedTxIdentity
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	// MustActorID / MustTenantID / system_admin / edit grant / asserted_caps (first Require)
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(actorID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(tenantID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false)) // not admin
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))  // has template.edit
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(""))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	// Second Require (template.manage) — actor does NOT have it
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(actorID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(tenantID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false)) // not admin
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false)) // not granted
	mock.ExpectRollback()

	_, gotErr := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     tenantID,
		ActorUserID:  actorID,
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})

	var denied iamauthz.ErrCapDenied
	if !errors.As(gotErr, &denied) {
		t.Fatalf("expected ErrCapDenied for non-creator/non-holder on unpublished template, got %T: %v", gotErr, gotErr)
	}
	if denied.Capability != string(iamdomain.CapTemplateManage) {
		t.Fatalf("expected ErrCapDenied for cap %q, got %q", iamdomain.CapTemplateManage, denied.Capability)
	}
}

// --- Retained edge-case tests (precondition guards, unchanged semantics) ---

func TestUpsertApprovalConfig_Archived(t *testing.T) {
	repo := newFakeRepo()
	archivedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	tpl := &domain.Template{
		ID:         "tpl-1",
		TenantID:   "tenant-a",
		CreatedBy:  "author-1",
		ArchivedAt: &archivedAt,
	}
	repo.templates[tpl.ID] = tpl

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "admin-1",
		TemplateID:   tpl.ID,
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
		ApproverRole: "",
	})
	if !errors.Is(err, domain.ErrInvalidApprovalConfig) {
		t.Fatalf("expected ErrInvalidApprovalConfig, got %v", err)
	}
}

// TestUpsertApprovalConfig_WithDB_UsesTxAndBothAuthzChecks verifies that the
// in-tx sequence seeds identity, checks template.edit, then template.manage
// (when requireManage=true — published template). Uses permissive mock because
// the two-Require sequence (with system_admin bypass for each) has ordering
// complexity not worth asserting in a unit test; correctness of individual Require
// calls is covered by the authz package's own tests.
func TestUpsertApprovalConfig_WithDB_UsesTxAndBothAuthzChecks(t *testing.T) {
	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:                 "tpl-1",
		TenantID:           "tenant-a",
		CreatedBy:          "author-1",
		PublishedVersionID: strPtr("ver-1"),
	}
	repo.templates[tpl.ID] = tpl
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).
		WithRunner(newTxRunner(newPermissiveMockDB(t)))

	got, err := svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "qms-1",
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("UpsertApprovalConfig returned error: %v", err)
	}
	if got == nil || got.TemplateID != tpl.ID {
		t.Fatalf("unexpected config result: %+v", got)
	}
}

// TestUpsertApprovalConfig_WithDB_CreatorPath_UsesTxAndSingleAuthzCheck verifies
// the creator-on-unpublished path: only template.edit is required in-tx (no
// template.manage call); strict mock captures the exact SQL.
func TestUpsertApprovalConfig_WithDB_CreatorPath_UsesTxAndSingleAuthzCheck(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:        "tpl-1",
		TenantID:  "tenant-a",
		CreatedBy: "author-1",
	}
	repo.templates[tpl.ID] = tpl
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(db))

	mock.ExpectBegin()
	expectTemplateEditAuthzConfig(mock, "author-1", "tenant-a")
	mock.ExpectCommit()

	_, err = svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     "tenant-a",
		ActorUserID:  "author-1",
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("UpsertApprovalConfig returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestUpsertApprovalConfig_WithDB_ManageHolderPath_AssertsBothRequires is the
// strict-mock counterpart to the permissive ManageHolder happy-path test: it
// asserts the published-template path runs BOTH in-tx authz.Require calls
// (template.edit THEN template.manage, both granted) in order, then commits.
// A permissive mock would still pass if the second Require were accidentally
// deleted; here ExpectationsWereMet fails if the template.manage check is skipped
// (its query/exec expectations would remain unmet) — closing that coverage gap
// (F-CD7 stage-2 review Finding 6).
func TestUpsertApprovalConfig_WithDB_ManageHolderPath_AssertsBothRequires(t *testing.T) {
	actorID, tenantID := "qms-1", "tenant-a"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	tpl := &domain.Template{
		ID:                 "tpl-1",
		TenantID:           tenantID,
		CreatedBy:          "author-1", // actor is NOT the creator → manage required even if unpublished; published also forces it
		PublishedVersionID: strPtr("ver-1"),
	}
	repo.templates[tpl.ID] = tpl
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(db))

	mock.ExpectBegin()
	// SeedTxIdentity
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	// Require(template.edit) — explicit grant (non-admin) path, granted.
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(actorID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(tenantID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false)) // not admin
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))  // has template.edit
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(""))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	// Require(template.manage) — second check; MUST run for published templates, granted.
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(actorID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(tenantID))
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false)) // not admin
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))  // has template.manage
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(`[{"area":"tenant","cap":"template.edit"}]`))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err = svc.UpsertApprovalConfig(context.Background(), application.UpsertApprovalConfigCmd{
		TenantID:     tenantID,
		ActorUserID:  actorID,
		TemplateID:   tpl.ID,
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("manage holder on published template: expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations (both authz.Require must fire): %v", err)
	}
}

// expectTemplateEditAuthzConfig sets up strict sqlmock expectations for the
// SeedTxIdentity + Require(template.edit) in-tx sequence (system_admin bypass path).
// Used by tests that verify the creator-on-unpublished path where only one
// authz.Require call (template.edit) is made.
func expectTemplateEditAuthzConfig(mock sqlmock.Sqlmock, actorID, tenantID string) {
	mock.ExpectExec(regexp.QuoteMeta(`
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`)).
		WithArgs(tenantID, actorID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(actorID))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(tenantID))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(`[{"area":"tenant","cap":"template.edit"}]`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}
