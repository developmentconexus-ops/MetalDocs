package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/documents/repository"
)

// TestListDocumentsPaginated_ScansSnapshotAndRevisionColumns guards the F12 fix:
// the list query must SELECT and Scan profile_code_snapshot, process_area_code_snapshot,
// revision_version, and revision_number so they reach the API response.
// The query is also exercised with NULL snapshot values to confirm *string scan safety.
func TestListDocumentsPaginated_ScansSnapshotAndRevisionColumns(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	cols := []string{
		"id", "tenant_id", "template_version_id", "name", "status", "form_data_json",
		"current_revision_id", "active_session_id", "archived_at",
		"created_at", "updated_at", "created_by", "controlled_document_id", "code",
		"profile_code_snapshot", "process_area_code_snapshot",
		"revision_version", "revision_number",
		"effective_from", "effective_to", "review_due_at", "last_reviewed_at", "review_surfaced_at",
		"total_count",
	}

	areaCode := "rh"
	profileCode := "operational"
	reviewDueAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(cols).
		// Row 1: snapshots populated, governed revision 1, review_due_at set
		// (M6 F6.2 T6 — the other four review/expiry columns left NULL to
		// also prove nil-safe scan for the still-unset fields).
		AddRow(
			"doc-1", "tenant-1", "tpl-1", "Doc One", "draft", []byte("{}"),
			"", "", nil,
			time.Unix(0, 0), time.Unix(0, 0), "user-1", nil, "PO-RH-001",
			profileCode, areaCode,
			int64(1), int64(1),
			nil, nil, reviewDueAt, nil, nil,
			int64(2),
		).
		// Row 2: snapshots NULL (must scan into *string without panicking) and rev 0.
		// All five review/expiry columns NULL (legacy row, no cycle set).
		AddRow(
			"doc-2", "tenant-1", "tpl-1", "Doc Two", "draft", []byte("{}"),
			"", "", nil,
			time.Unix(0, 0), time.Unix(0, 0), "user-1", nil, "PO-RH-002",
			nil, nil,
			int64(0), int64(0),
			nil, nil, nil, nil, nil,
			int64(2),
		)

	// Match SELECT shape and ORDER BY. Anchor on the four columns whose absence caused F12.
	mock.ExpectQuery(regexp.QuoteMeta(`profile_code_snapshot, process_area_code_snapshot`)).
		WillReturnRows(rows)

	got, _, _, err := repo.ListDocumentsPaginated(context.Background(), "tenant-1", repository.ListOptions{PageSize: 20})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}

	d1 := got[0]
	if d1.ProfileCodeSnapshot == nil || *d1.ProfileCodeSnapshot != profileCode {
		t.Errorf("row 1 ProfileCodeSnapshot = %v, want %q", d1.ProfileCodeSnapshot, profileCode)
	}
	if d1.ProcessAreaCodeSnapshot == nil || *d1.ProcessAreaCodeSnapshot != areaCode {
		t.Errorf("row 1 ProcessAreaCodeSnapshot = %v, want %q", d1.ProcessAreaCodeSnapshot, areaCode)
	}
	if d1.RevisionVersion != 1 {
		t.Errorf("row 1 RevisionVersion = %d, want 1", d1.RevisionVersion)
	}
	if d1.RevisionNumber != 1 {
		t.Errorf("row 1 RevisionNumber = %d, want 1", d1.RevisionNumber)
	}

	if d1.ReviewDueAt == nil || !d1.ReviewDueAt.Equal(reviewDueAt) {
		t.Errorf("row 1 ReviewDueAt = %v, want %v", d1.ReviewDueAt, reviewDueAt)
	}
	if d1.EffectiveFrom != nil {
		t.Errorf("row 1 EffectiveFrom = %v, want nil", *d1.EffectiveFrom)
	}
	if d1.EffectiveTo != nil {
		t.Errorf("row 1 EffectiveTo = %v, want nil", *d1.EffectiveTo)
	}
	if d1.LastReviewedAt != nil {
		t.Errorf("row 1 LastReviewedAt = %v, want nil", *d1.LastReviewedAt)
	}
	if d1.ReviewSurfacedAt != nil {
		t.Errorf("row 1 ReviewSurfacedAt = %v, want nil", *d1.ReviewSurfacedAt)
	}

	d2 := got[1]
	if d2.ProfileCodeSnapshot != nil {
		t.Errorf("row 2 ProfileCodeSnapshot = %v, want nil", *d2.ProfileCodeSnapshot)
	}
	if d2.ProcessAreaCodeSnapshot != nil {
		t.Errorf("row 2 ProcessAreaCodeSnapshot = %v, want nil", *d2.ProcessAreaCodeSnapshot)
	}
	if d2.RevisionVersion != 0 {
		t.Errorf("row 2 RevisionVersion = %d, want 0", d2.RevisionVersion)
	}
	if d2.RevisionNumber != 0 {
		t.Errorf("row 2 RevisionNumber = %d, want 0", d2.RevisionNumber)
	}
	if d2.EffectiveFrom != nil || d2.EffectiveTo != nil || d2.ReviewDueAt != nil || d2.LastReviewedAt != nil || d2.ReviewSurfacedAt != nil {
		t.Errorf("row 2 review/expiry fields = (%v,%v,%v,%v,%v), want all nil (legacy row, no cycle set)",
			d2.EffectiveFrom, d2.EffectiveTo, d2.ReviewDueAt, d2.LastReviewedAt, d2.ReviewSurfacedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
