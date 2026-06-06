package v2documents

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	searchdomain "metaldocs/internal/modules/search/domain"
)

// readerCols are the columns ListDocuments projects, in order.
var readerCols = []string{
	"id", "name", "status", "profile_code_snapshot", "family_code",
	"process_area_code_snapshot", "department_code", "created_by",
	"code", "sequence_num", "effective_from", "effective_to", "created_at",
}

func TestListDocumentsFiltersByTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(readerCols).AddRow(
		"doc-1", "Manual", "ACTIVE", "profile-a", "family-a",
		"quality", "qa", "user-1",
		"QMS-001", 7, nil, nil,
		time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	)
	// The query must enforce per-document visibility against the unified grant
	// model — assert the predicate is present, then assert the actor is bound.
	mock.ExpectQuery("controlled_document_area_grants[\\s\\S]*controlled_document_user_grants[\\s\\S]*LIMIT \\$11 OFFSET \\$12").
		WithArgs("tenant-1", "", "", "", "", "", "", "", nil, nil, 20, 0, "user-1").
		WillReturnRows(rows)

	got, err := NewReader(db).ListDocuments(context.Background(), searchdomain.Query{TenantID: "tenant-1", ActorUserID: "user-1"}, 20, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("documents = %d, want 1", len(got))
	}
	if got[0].DocumentFamily != "family-a" {
		t.Fatalf("family = %q, want family-a", got[0].DocumentFamily)
	}
	if got[0].Department != "qa" {
		t.Fatalf("department = %q, want qa", got[0].Department)
	}
	if got[0].DocumentCode != "QMS-001" || got[0].DocumentSequence != 7 {
		t.Fatalf("document identity = (%q,%d), want (QMS-001,7)", got[0].DocumentCode, got[0].DocumentSequence)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListDocumentsBindsActorForVisibility(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(readerCols).AddRow(
		"doc-2", "Instruction", "ACTIVE", "profile-b", "family-b",
		"quality", "", "user-2",
		"QMS-002", 8, nil, nil,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
	)
	mock.ExpectQuery("controlled_document_area_grants[\\s\\S]*controlled_document_user_grants[\\s\\S]*LIMIT \\$11 OFFSET \\$12").
		WithArgs("tenant-1", "", "", "", "", "", "", "", nil, nil, 20, 0, "user-9").
		WillReturnRows(rows)

	got, err := NewReader(db).ListDocuments(context.Background(), searchdomain.Query{TenantID: "tenant-1", ActorUserID: "user-9"}, 20, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("documents = %d, want 1", len(got))
	}
	if got[0].Department != "" {
		t.Fatalf("department = %q, want empty when unset", got[0].Department)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
