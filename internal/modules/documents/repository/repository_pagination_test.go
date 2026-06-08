package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/platform/pagination"
)

func TestListDocumentsPaginated_StatusFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	r := New(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "template_version_id", "name", "status", "form_data_json",
		"current_revision_id", "active_session_id", "archived_at",
		"created_at", "updated_at", "created_by", "controlled_document_id", "code",
		"profile_code_snapshot", "process_area_code_snapshot", "revision_version", "revision_number",
	}).
		AddRow("doc-1", "tenant-1", "tpl-1", "Doc B", "under_review", []byte(`{}`), "", "", nil, now, now, "user-1", nil, "", "prof", "area", int64(1), int64(0)).
		AddRow("doc-2", "tenant-1", "tpl-1", "Doc C", "under_review", []byte(`{}`), "", "", nil, now, now, "user-2", nil, "", "prof", "area", int64(1), int64(0))

	// First page (no cursor): filter args + limit+1 probe (=11).
	mock.ExpectQuery("status = ANY").
		WithArgs("tenant-1", sqlmock.AnyArg(), 11).
		WillReturnRows(rows)

	got, hasMore, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{
		Status:   []string{"under_review"},
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated: %v", err)
	}
	if len(got) != 2 || hasMore {
		t.Fatalf("want 2 rows hasMore=false, got %d hasMore=%v", len(got), hasMore)
	}
	for _, d := range got {
		if d.Status != "under_review" {
			t.Fatalf("unexpected status %q", d.Status)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListDocumentsPaginated_CursorKeyset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	r := New(db)
	now := time.Now()

	makeRows := func(n int) *sqlmock.Rows {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "template_version_id", "name", "status", "form_data_json",
			"current_revision_id", "active_session_id", "archived_at",
			"created_at", "updated_at", "created_by", "controlled_document_id", "code",
			"profile_code_snapshot", "process_area_code_snapshot", "revision_version", "revision_number",
		})
		for i := 0; i < n; i++ {
			rows.AddRow("doc-id", "tenant-1", "tpl-1", "Doc", "draft", []byte(`{}`), "", "", nil, now, now, "user-1", nil, "", "prof", "area", int64(1), int64(0))
		}
		return rows
	}

	// First page: no cursor, ORDER BY (updated_at, id) DESC, LIMIT $2 (limit+1=11).
	// 11 rows returned → trimmed to 10 + hasMore=true.
	mock.ExpectQuery("ORDER BY updated_at DESC, id DESC").WithArgs("tenant-1", 11).WillReturnRows(makeRows(11))
	page1, hasMore1, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{PageSize: 10})
	if err != nil {
		t.Fatalf("page1 err: %v", err)
	}
	if len(page1) != 10 || !hasMore1 {
		t.Fatalf("page1: want 10 rows hasMore=true, got %d hasMore=%v", len(page1), hasMore1)
	}

	// Second page: cursor binds the keyset anchor (updated_at, id) < ($2, $3),
	// then LIMIT $4. 5 rows → no probe row → hasMore=false (last page).
	cursor := pagination.EncodeCursor("2026-01-01T00:00:00Z", "doc-x")
	mock.ExpectQuery(`\(updated_at, id\) < \(\$2::timestamptz, \$3\)`).
		WithArgs("tenant-1", "2026-01-01T00:00:00Z", "doc-x", 11).
		WillReturnRows(makeRows(5))
	page2, hasMore2, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{PageSize: 10, Cursor: cursor})
	if err != nil {
		t.Fatalf("page2 err: %v", err)
	}
	if len(page2) != 5 || hasMore2 {
		t.Fatalf("page2: want 5 rows hasMore=false, got %d hasMore=%v", len(page2), hasMore2)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListDocumentsPaginated_InvalidCursor(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	r := New(db)
	if _, _, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{Cursor: "!!!bad"}); err != pagination.ErrInvalidCursor {
		t.Fatalf("want ErrInvalidCursor, got %v", err)
	}
}

func TestCountDocuments_RespectsFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	r := New(db)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM documents WHERE`).
		WithArgs("tenant-1", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	got, err := r.CountDocuments(context.Background(), "tenant-1", ListOptions{Status: []string{"draft"}})
	if err != nil {
		t.Fatalf("CountDocuments: %v", err)
	}
	if got != 2 {
		t.Fatalf("want count 2, got %d", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
