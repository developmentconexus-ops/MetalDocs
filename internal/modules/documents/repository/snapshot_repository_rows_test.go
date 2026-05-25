package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/documents/domain"
)

func TestSnapshotRepository_WriteMethodsNoRowsReturnError(t *testing.T) {
	tests := []struct {
		name string
		run  func(context.Context, *SnapshotRepository) error
	}{
		{
			name: "WriteSnapshot",
			run: func(ctx context.Context, repo *SnapshotRepository) error {
				return repo.WriteSnapshot(ctx, "tenant-1", "doc-1", domain.TemplateSnapshot{
					PlaceholderSchemaJSON: []byte(`{}`),
					CompositionJSON:       []byte(`{}`),
					BodyDocxS3Key:         "snapshots/doc.docx",
				})
			},
		},
		{
			name: "WriteFreeze",
			run: func(ctx context.Context, repo *SnapshotRepository) error {
				return repo.WriteFreeze(ctx, "tenant-1", "doc-1", []byte("hash"), time.Now())
			},
		},
		{
			name: "WriteFinalDocx",
			run: func(ctx context.Context, repo *SnapshotRepository) error {
				return repo.WriteFinalDocx(ctx, "tenant-1", "doc-1", "final/doc.docx", []byte("hash"))
			},
		},
		{
			name: "WritePDF",
			run: func(ctx context.Context, repo *SnapshotRepository) error {
				return repo.WritePDF(ctx, "tenant-1", "doc-1", "final/doc.pdf", []byte("hash"), time.Now())
			},
		},
		{
			name: "AppendReconstruction",
			run: func(ctx context.Context, repo *SnapshotRepository) error {
				return repo.AppendReconstruction(ctx, "tenant-1", "doc-1", []byte(`[{}]`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			mock.ExpectExec(`UPDATE documents`).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err = tt.run(context.Background(), NewSnapshotRepository(db))
			if err == nil || !strings.Contains(err.Error(), "not found or already in target state") {
				t.Fatalf("%s error = %v, want not found/target state", tt.name, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
