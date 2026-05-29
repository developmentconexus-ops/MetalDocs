package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/documents/repository"
)

// TestListDocumentsPaginated_ScansSnapshotAndRevisionColumns guards the F12 fix:
// the list query must SELECT and Scan profile_code_snapshot, process_area_code_snapshot,
// revision_version, and revision_number so they reach the API response.
// The query is also exercised with NULL snapshot values to confirm *string scan safety.
func TestListDocumentsPaginated_ScansSnapshotAndRevisionColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.New(db)

	cols := []string{
		"id", "tenant_id", "template_version_id", "name", "status", "form_data_json",
		"current_revision_id", "active_session_id", "archived_at",
		"created_at", "updated_at", "created_by", "controlled_document_id", "code",
		"profile_code_snapshot", "process_area_code_snapshot",
		"revision_version", "revision_number",
	}

	areaCode := "rh"
	profileCode := "operational"

	rows := sqlmock.NewRows(cols).
		// Row 1: snapshots populated, governed revision 1.
		AddRow(
			"doc-1", "tenant-1", "tpl-1", "Doc One", "draft", []byte("{}"),
			"", "", nil,
			time.Unix(0, 0), time.Unix(0, 0), "user-1", nil, "PO-RH-001",
			profileCode, areaCode,
			int64(1), int64(1),
		).
		// Row 2: snapshots NULL (must scan into *string without panicking) and rev 0.
		AddRow(
			"doc-2", "tenant-1", "tpl-1", "Doc Two", "draft", []byte("{}"),
			"", "", nil,
			time.Unix(0, 0), time.Unix(0, 0), "user-1", nil, "PO-RH-002",
			nil, nil,
			int64(0), int64(0),
		)

	// Match SELECT shape and ORDER BY. Anchor on the four columns whose absence caused F12.
	mock.ExpectQuery(regexp.QuoteMeta(`profile_code_snapshot, process_area_code_snapshot`)).
		WillReturnRows(rows)

	got, err := repo.ListDocumentsPaginated(context.Background(), "tenant-1", repository.ListOptions{Page: 1, PageSize: 20})
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
