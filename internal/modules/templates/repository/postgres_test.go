package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"metaldocs/internal/modules/templates/domain"
)

func TestUpdateVersionIncrementsLockVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := New(db)
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          "tpl-1",
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		MetadataSchema:      domain.MetadataSchema{},
		PlaceholderSchema:   []domain.Placeholder{},
		AuthorID:            "author-1",
		PendingApproverRole: "approver",
		LockVersion:         4,
		CreatedAt:           time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE templates_template_version")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateVersion(context.Background(), "tenant-a", version); err != nil {
		t.Fatalf("UpdateVersion: %v", err)
	}
	if version.LockVersion != 5 {
		t.Fatalf("lock version = %d, want 5", version.LockVersion)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUpdateVersionReturnsStaleLockVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := New(db)
	version := &domain.TemplateVersion{
		ID:                  "ver-1",
		TemplateID:          "tpl-1",
		VersionNumber:       1,
		Status:              domain.VersionStatusInReview,
		MetadataSchema:      domain.MetadataSchema{},
		PlaceholderSchema:   []domain.Placeholder{},
		AuthorID:            "author-1",
		PendingApproverRole: "approver",
		LockVersion:         2,
		CreatedAt:           time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE templates_template_version")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS (")).
		WithArgs("ver-1", "tenant-a").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if err := repo.UpdateVersion(context.Background(), "tenant-a", version); err != domain.ErrStaleLockVersion {
		t.Fatalf("error = %v, want ErrStaleLockVersion", err)
	}
	if version.LockVersion != 2 {
		t.Fatalf("lock version = %d, want unchanged 2", version.LockVersion)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestGetVersionMatchesUUIDTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := New(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT
	v.id::text, v.template_id::text, v.version_number, v.status, v.docx_storage_key, v.content_hash,
	v.metadata_schema, v.placeholder_schema, v.author_id,
	v.pending_reviewer_role, v.pending_approver_role, v.reviewer_id, v.approver_id,
	v.submitted_at, v.reviewed_at, v.approved_at, v.published_at, v.obsoleted_at, v.lock_version, v.created_at
FROM templates_template_version v
JOIN templates_template t ON t.id = v.template_id
WHERE v.template_id = $1 AND v.version_number = $2 AND t.tenant_id = $3::uuid`)).
		WithArgs("tpl-1", 1, "tenant-a").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "template_id", "version_number", "status", "docx_storage_key", "content_hash",
			"metadata_schema", "placeholder_schema", "author_id",
			"pending_reviewer_role", "pending_approver_role", "reviewer_id", "approver_id",
			"submitted_at", "reviewed_at", "approved_at", "published_at", "obsoleted_at", "lock_version", "created_at",
		}).AddRow(
			"ver-1", "tpl-1", 1, "published", "system/templates/blank.docx", "hash-1",
			`{}`, `[]`, "admin",
			nil, "system", nil, nil,
			nil, nil, nil, time.Date(2026, 5, 27, 19, 0, 0, 0, time.UTC), nil, 0, time.Date(2026, 5, 27, 19, 0, 0, 0, time.UTC),
		))

	version, err := repo.GetVersion(context.Background(), "tenant-a", "tpl-1", 1)
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if version.ID != "ver-1" {
		t.Fatalf("version.ID = %q, want ver-1", version.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
