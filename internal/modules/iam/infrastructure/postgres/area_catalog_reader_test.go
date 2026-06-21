package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
)

func TestProcessAreaCatalog_AreaCodeExists(t *testing.T) {
	tests := []struct {
		name     string
		exists   bool
		queryErr error
		want     bool
		wantErr  bool
	}{
		{name: "area present", exists: true, want: true},
		{name: "area absent", exists: false, want: false},
		{name: "query error", queryErr: errors.New("boom"), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			expect := mock.ExpectQuery("FROM metaldocs.document_process_areas").
				WithArgs("tenant-1", "welding")
			if tc.queryErr != nil {
				expect.WillReturnError(tc.queryErr)
			} else {
				expect.WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(tc.exists))
			}

			got, err := NewProcessAreaCatalog(db, taxonomyinfra.NewAreaCatalogReaderPG()).AreaCodeExists(context.Background(), "tenant-1", "welding")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("AreaCodeExists: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
