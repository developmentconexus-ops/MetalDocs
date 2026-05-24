package v2documents

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListDocumentsFiltersByTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("LIMIT $2")).
		WithArgs("tenant-1", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"name",
			"status",
			"profile_code_snapshot",
			"process_area_code_snapshot",
			"created_by",
			"code",
			"sequence_num",
			"created_at",
		}))

	_, err = NewReader(db).ListDocuments(context.Background(), "tenant-1", 20)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
