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
	if got.UploadURL != "https://presigned/put/templates/tpl-1/versions/3.docx" {
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
		Status:         domain.VersionStatusInReview,
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

func TestPresignTemplateUpload_IgnoresCallerStorageKey(t *testing.T) {
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
		StorageKey:    "templates/other-tenant/versions/1.docx",
	})
	if err != nil {
		t.Fatalf("PresignTemplateUpload returned error: %v", err)
	}
	if got.StorageKey != "templates/tpl-1/versions/1.docx" {
		t.Fatalf("expected server-derived storage key, got %q", got.StorageKey)
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
	presigner := &fakePresigner{HeadResult: "hash_abc"}
	svc := application.New(repo, presigner, fakeClock{}, &fakeUUID{})

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
	if presigner.DeleteCalled != 0 {
		t.Fatalf("expected DeleteCalled 0, got %d", presigner.DeleteCalled)
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
	svc := application.New(repo, &fakePresigner{HeadResult: "hash_abc"}, fakeClock{}, &fakeUUID{}).WithDB(db)

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
	presigner := &fakePresigner{HeadResult: "hash_actual"}
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
	if presigner.DeleteCalled != 1 {
		t.Fatalf("expected DeleteCalled 1, got %d", presigner.DeleteCalled)
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
	presigner := &fakePresigner{HeadErr: domain.ErrUploadMissing}
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
	if presigner.DeleteCalled != 0 {
		t.Fatalf("expected DeleteCalled 0, got %d", presigner.DeleteCalled)
	}
}

func TestSaveTemplateDraft_StaleLockVersion(t *testing.T) {
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
	repo.lockVersions["ver-1"] = 2
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})

	err := svc.SaveTemplateDraft(context.Background(), application.SaveTemplateDraftCmd{
		TenantID:            "tenant-a",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       1,
		ExpectedLockVersion: 1,
		DocxStorageKey:      "templates/tpl-1/versions/1.docx",
		SchemaStorageKey:    "templates/tpl-1/versions/1.schema.json",
		DocxContentHash:     "hash_new",
		SchemaContentHash:   "schema_hash",
	})
	if !errors.Is(err, domain.ErrStaleLockVersion) {
		t.Fatalf("expected ErrStaleLockVersion, got %v", err)
	}
}

func TestSaveTemplateDraft_WithDBSetsTemplateEditAuthz(t *testing.T) {
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
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/tpl-1/versions/1.docx",
	}
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).WithDB(db)

	mock.ExpectBegin()
	expectTemplateEditAuthz(mock, "user-a", "11111111-1111-1111-1111-111111111111")
	mock.ExpectCommit()

	err = svc.SaveTemplateDraft(context.Background(), application.SaveTemplateDraftCmd{
		TenantID:            "11111111-1111-1111-1111-111111111111",
		ActorUserID:         "user-a",
		TemplateID:          "tpl-1",
		VersionNumber:       1,
		ExpectedLockVersion: 0,
		DocxStorageKey:      "templates/tpl-1/versions/1.docx",
		SchemaStorageKey:    "templates/tpl-1/versions/1.schema.json",
		DocxContentHash:     "hash_new",
		SchemaContentHash:   "schema_hash",
	})
	if err != nil {
		t.Fatalf("SaveTemplateDraft returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func expectTemplateEditAuthz(mock sqlmock.Sqlmock, actorID, tenantID string) {
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
