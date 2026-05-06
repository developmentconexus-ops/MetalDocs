package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
	}).
		AddRow("doc-1", "tenant-1", "tpl-1", "Doc B", "under_review", []byte(`{}`), "", "", nil, now, now, "user-1", nil, "").
		AddRow("doc-2", "tenant-1", "tpl-1", "Doc C", "under_review", []byte(`{}`), "", "", nil, now, now, "user-2", nil, "")

	mock.ExpectQuery("status = ANY").
		WithArgs("tenant-1", sqlmock.AnyArg(), 10, 0).
		WillReturnRows(rows)

	got, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{
		Status:   []string{"under_review"},
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListDocumentsPaginated: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 rows, got %d", len(got))
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

func TestListDocumentsPaginated_PageOffset(t *testing.T) {
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
		})
		for i := 0; i < n; i++ {
			rows.AddRow("doc-id", "tenant-1", "tpl-1", "Doc", "draft", []byte(`{}`), "", "", nil, now, now, "user-1", nil, "")
		}
		return rows
	}

	mock.ExpectQuery("LIMIT \\$2 OFFSET \\$3").WithArgs("tenant-1", 10, 0).WillReturnRows(makeRows(10))
	mock.ExpectQuery("LIMIT \\$2 OFFSET \\$3").WithArgs("tenant-1", 10, 10).WillReturnRows(makeRows(10))
	mock.ExpectQuery("LIMIT \\$2 OFFSET \\$3").WithArgs("tenant-1", 10, 20).WillReturnRows(makeRows(5))

	page1, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("page1 err: %v", err)
	}
	page2, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("page2 err: %v", err)
	}
	page3, err := r.ListDocumentsPaginated(context.Background(), "tenant-1", ListOptions{Page: 3, PageSize: 10})
	if err != nil {
		t.Fatalf("page3 err: %v", err)
	}

	if len(page1) != 10 || len(page2) != 10 || len(page3) != 5 {
		t.Fatalf("want page sizes 10,10,5; got %d,%d,%d", len(page1), len(page2), len(page3))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
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
