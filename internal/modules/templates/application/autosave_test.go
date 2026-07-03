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
	"metaldocs/internal/platform/objectstore"
)

func TestPresignAutosave_Happy(t *testing.T) {
	repo := newFakeRepo()
	template := &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	version := &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  3,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/3.docx",
	}
	repo.templates[template.ID] = template
	repo.versions[version.ID] = version

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	got, err := svc.PresignAutosave(context.Background(), application.PresignAutosaveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "user-a",
		TemplateID:    "tpl-1",
		VersionNumber: 3,
	})
	if err != nil {
		t.Fatalf("PresignAutosave returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.UploadURL != "https://example/put" {
		t.Fatalf("unexpected upload url: %q", got.UploadURL)
	}
	if got.StorageKey != "templates/tpl-1/versions/3.docx" {
		t.Fatalf("unexpected storage key: %q", got.StorageKey)
	}
	wantExpiresAt := time.Date(2026, 4, 20, 12, 10, 0, 0, time.UTC)
	if !got.ExpiresAt.Equal(wantExpiresAt) {
		t.Fatalf("expected expiresAt %s, got %s", wantExpiresAt, got.ExpiresAt)
	}
}

func TestPresignAutosave_NonDraft(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  1,
		Status:         domain.VersionStatusUnderReview,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
	}

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	_, err := svc.PresignAutosave(context.Background(), application.PresignAutosaveCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "user-a",
		TemplateID:    "tpl-1",
		VersionNumber: 1,
	})
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestPresignTemplateUpload_UsesDBStorageKey(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
	}
	presigner := &fakePresigner{}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{})

	got, err := svc.PresignTemplateUpload(context.Background(), application.PresignTemplateUploadCmd{
		TenantID:      "tenant-a",
		ActorUserID:   "user-a",
		TemplateID:    "tpl-1",
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PresignTemplateUpload returned error: %v", err)
	}
	if got.StorageKey != "templates/tpl-1/versions/1.docx" {
		t.Fatalf("expected DB-sourced storage key, got %q", got.StorageKey)
	}
	if len(presigner.PutKeys) != 1 || presigner.PutKeys[0] != "templates/tpl-1/versions/1.docx" {
		t.Fatalf("unexpected presign keys: %v", presigner.PutKeys)
	}
}

func TestCommitAutosave_Happy(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  7,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/7.docx",
	}
	presigner := &fakePresigner{}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	got, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       7,
		ExpectedContentHash: "hash_abc",
	})
	if err != nil {
		t.Fatalf("CommitAutosave returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil version")
	}
	if got.ContentHash != "hash_abc" {
		t.Fatalf("expected content hash hash_abc, got %q", got.ContentHash)
	}
	if len(repo.audit) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(repo.audit))
	}
	audit := repo.audit[0]
	if audit.Action != domain.AuditSaved {
		t.Fatalf("expected action %q, got %q", domain.AuditSaved, audit.Action)
	}
	detailHash, ok := audit.Details["content_hash"]
	if !ok || detailHash != "hash_abc" {
		t.Fatalf("expected details content_hash=hash_abc, got %v", audit.Details)
	}
}

func TestCommitAutosave_WithDBSetsTemplateEditAuthz(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "11111111-1111-1111-1111-111111111111",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  7,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/7.docx",
	}
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(db))

	mock.ExpectBegin()
	expectTemplateEditAuthz(mock, "user-a", "11111111-1111-1111-1111-111111111111")
	mock.ExpectCommit()

	_, err = svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "11111111-1111-1111-1111-111111111111",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       7,
		ExpectedContentHash: "hash_abc",
	})
	if err != nil {
		t.Fatalf("CommitAutosave returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestCommitAutosave_LoserGetsConcurrentTransition pins APP-06/STO-01: the
// unlocked GetVersion read at the top of CommitAutosave races a concurrent
// writer that lands first and bumps lock_version. The repository's CAS UPDATE
// (UpdateVersionTx -> updateVersion, WHERE lock_version = $N) then affects 0
// rows and returns domain.ErrStaleLockVersion. CommitAutosave must remap that
// to domain.ErrConcurrentTransition (matching every other lifecycle
// transition's F-T3 remap, and the commitTemplateAutosave OpenAPI contract,
// which documents 409 and has no 412 response) instead of silently losing the
// race or leaking the raw 412-mapped error.
func TestCommitAutosave_LoserGetsConcurrentTransition(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  9,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/9.docx",
		LockVersion:    0,
	}
	// Simulate a concurrent commit (or other version write) that landed first
	// and bumped lock_version after our unlocked GetVersion read captured 0.
	repo.lockVersions["ver-1"] = 1

	presigner := &fakePresigner{}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       9,
		ExpectedContentHash: "hash_abc",
	})
	if !errors.Is(err, domain.ErrConcurrentTransition) {
		t.Fatalf("expected ErrConcurrentTransition (409-class), got %v", err)
	}
	// The loser must NOT have appended a "saved" audit event — the write never
	// committed, so there is no state change to attest to.
	if len(repo.audit) != 0 {
		t.Fatalf("expected no audit events for the losing racer, got %v", repo.audit)
	}
}

// TestCommitAutosave_WinnerPathUnchanged is the control for the above: with no
// concurrent writer (lock_version untouched between GetVersion and the CAS
// UPDATE), CommitAutosave must still succeed exactly as before, incrementing
// lock_version and appending exactly one "saved" audit event.
func TestCommitAutosave_WinnerPathUnchanged(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  9,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/9.docx",
		LockVersion:    3,
	}
	repo.lockVersions["ver-1"] = 3

	presigner := &fakePresigner{}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(newPermissiveMockDB(t)))

	got, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       9,
		ExpectedContentHash: "hash_abc",
	})
	if err != nil {
		t.Fatalf("CommitAutosave returned error: %v", err)
	}
	if got.LockVersion != 4 {
		t.Fatalf("expected lock_version bumped to 4, got %d", got.LockVersion)
	}
	if len(repo.audit) != 1 || repo.audit[0].Action != domain.AuditSaved {
		t.Fatalf("expected exactly 1 saved audit event, got %v", repo.audit)
	}
}

// TestCommitAutosave_WithDB_LoserCommitEnvelopeRollsBack exercises the same
// loser race against a real sqlmock *sql.DB (not the fakeRepo lock table) to
// pin the tx envelope: the CAS UPDATE affects 0 rows, the repo's existence
// recheck finds the row still exists (so ErrStaleLockVersion, not
// ErrNotFound), and the service must surface ErrConcurrentTransition while the
// runner rolls back (no ExpectCommit — Do() rolls back on a non-nil closure
// error).
func TestCommitAutosave_WithDB_LoserCommitEnvelopeRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "11111111-1111-1111-1111-111111111111",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  7,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/7.docx",
		LockVersion:    0,
	}
	// Simulate a concurrent winner ahead of us in the fake lock table too, so
	// the fakeRepo.UpdateVersion (invoked via UpdateVersionTx) returns
	// ErrStaleLockVersion just like the real CAS UPDATE would on a 0-row affect.
	repo.lockVersions["ver-1"] = 1

	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(db))

	mock.ExpectBegin()
	expectTemplateEditAuthz(mock, "user-a", "11111111-1111-1111-1111-111111111111")
	mock.ExpectRollback()

	_, err = svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "11111111-1111-1111-1111-111111111111",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       7,
		ExpectedContentHash: "hash_abc",
	})
	if !errors.Is(err, domain.ErrConcurrentTransition) {
		t.Fatalf("expected ErrConcurrentTransition (409-class), got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestCommitAutosave_HashMismatch(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  2,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/2.docx",
	}
	presigner := &fakePresigner{confirmErr: objectstore.ErrHashMismatch}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{})

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       2,
		ExpectedContentHash: "hash_expected",
	})
	if !errors.Is(err, domain.ErrContentHashMismatch) {
		t.Fatalf("expected ErrContentHashMismatch, got %v", err)
	}
}

func TestCommitAutosave_HashMismatchMapsToDomainError(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{ID: "tpl-1", TenantID: "tenant-a"}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  2,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/2.docx",
	}
	// Confirm handles mismatch internally (quiet delete); port maps kernel error to domain error.
	presigner := &fakePresigner{confirmErr: objectstore.ErrHashMismatch}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{})

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       2,
		ExpectedContentHash: "hash_expected",
	})
	if !errors.Is(err, domain.ErrContentHashMismatch) {
		t.Fatalf("expected ErrContentHashMismatch, got %v", err)
	}
}

func TestCommitAutosave_UploadMissing(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  4,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/4.docx",
	}
	presigner := &fakePresigner{confirmErr: objectstore.ErrObjectMissing}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{})

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       4,
		ExpectedContentHash: "hash_abc",
	})
	if !errors.Is(err, domain.ErrUploadMissing) {
		t.Fatalf("expected ErrUploadMissing, got %v", err)
	}
}

func TestCommitAutosave_UploadTooLarge(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["tpl-1"] = &domain.Template{
		ID:       "tpl-1",
		TenantID: "tenant-a",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:             "ver-1",
		TemplateID:     "tpl-1",
		VersionNumber:  5,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/5.docx",
	}
	presigner := &fakePresigner{confirmErr: objectstore.ErrObjectTooLarge}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{})

	_, err := svc.CommitAutosave(context.Background(), application.CommitAutosaveCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       5,
		ExpectedContentHash: "hash_abc",
	})
	if !errors.Is(err, domain.ErrUploadTooLarge) {
		t.Fatalf("expected ErrUploadTooLarge, got %v", err)
	}
}

func expectTemplateEditAuthz(mock sqlmock.Sqlmock, actorID, tenantID string) {
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
