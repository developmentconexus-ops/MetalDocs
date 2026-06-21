package repository

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// TestLoadDocumentArea_FailClosed locks the ADR 0022 Phase 7 security property:
// a missing document row yields ("", nil) — NOT an error and NOT a "tenant"
// fallback — so the empty area string fails closed at authz.Require for
// non-system actors (the SQL predicate ($2='tenant' OR area_code=$2) cannot match
// an empty area). A present row returns its real area.
func TestLoadDocumentArea_FailClosed(t *testing.T) {
	const q = "SELECT process_area_code_snapshot FROM documents WHERE tenant_id=$1 AND id=$2"

	t.Run("missing row -> empty, no error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("tenant-1", "missing").WillReturnError(sql.ErrNoRows)
		tx, _ := db.Begin()
		area, err := loadDocumentArea(context.Background(), tx, "tenant-1", "missing")
		if err != nil || area != "" {
			t.Fatalf("missing row: want (\"\", nil), got (%q, %v)", area, err)
		}
	})

	t.Run("present row -> real area", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("tenant-1", "doc-1").
			WillReturnRows(sqlmock.NewRows([]string{"process_area_code_snapshot"}).AddRow("qms"))
		tx, _ := db.Begin()
		area, err := loadDocumentArea(context.Background(), tx, "tenant-1", "doc-1")
		if err != nil || area != "qms" {
			t.Fatalf("present row: want (\"qms\", nil), got (%q, %v)", area, err)
		}
	})
}

func TestMarkArchived_StampsTimestampWithoutStatusChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{})

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.MarkArchived(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("MarkArchived: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkArchived_NoRowsReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{})

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = r.MarkArchived(context.Background(), "tenant-1", "doc-1", "user-1")
	if err == nil || !strings.Contains(err.Error(), "not found or already in target state") {
		t.Fatalf("MarkArchived error = %v, want not found/target state", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchive_ClearsTimestamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{})

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.Unarchive(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("Unarchive: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchive_NoRowsReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{})

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = r.Unarchive(context.Background(), "tenant-1", "doc-1", "user-1")
	if err == nil || !strings.Contains(err.Error(), "not found or already in target state") {
		t.Fatalf("Unarchive error = %v, want not found/target state", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectDocumentEditAuthz(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	mock.ExpectExec(`(?s)SELECT\s+set_config\('metaldocs\.tenant_id', \$1, true\),\s*set_config\('metaldocs\.actor_id', \$2, true\)`).
		WithArgs("tenant-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// ADR 0022 Phase 7: document.edit loads the document's process area before
	// the tier-2 check so authz is area-scoped, not tenant-wide.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT process_area_code_snapshot FROM documents WHERE tenant_id=$1 AND id=$2")).
		WithArgs("tenant-1", "doc-1").
		WillReturnRows(sqlmock.NewRows([]string{"process_area_code_snapshot"}).AddRow("area-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("user-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("tenant-1"))
	mock.ExpectQuery(`SELECT EXISTS`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps'")).
		WillReturnResult(sqlmock.NewResult(0, 0))
}
