package v2documents

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	searchdomain "metaldocs/internal/modules/search/domain"
)

func TestListDocumentsFiltersByTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id",
		"name",
		"status",
		"profile_code_snapshot",
		"family_code",
		"process_area_code_snapshot",
		"subject_code",
		"business_unit",
		"department_code",
		"classification",
		"tags",
		"created_by",
		"code",
		"sequence_num",
		"effective_from",
		"effective_to",
		"created_at",
	}).AddRow(
		"doc-1",
		"Manual",
		"ACTIVE",
		"profile-a",
		"family-a",
		"quality",
		"deviation",
		"ops",
		"qa",
		"INTERNAL",
		`["tag-a","tag-b"]`,
		"user-1",
		"QMS-001",
		7,
		nil,
		nil,
		time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	)
	mock.ExpectQuery(regexp.QuoteMeta("LIMIT $15 OFFSET $16")).
		WithArgs("tenant-1", "", "", "", "", "", "", "", "", "", "", "", nil, nil, 20, 0).
		WillReturnRows(rows)

	got, err := NewReader(db).ListDocuments(context.Background(), searchdomain.Query{TenantID: "tenant-1"}, 20, 0)
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
	if got[0].BusinessUnit != "ops" {
		t.Fatalf("business unit = %q, want ops", got[0].BusinessUnit)
	}
	if got[0].Subject != "deviation" {
		t.Fatalf("subject = %q, want deviation", got[0].Subject)
	}
	if got[0].Classification != searchdomain.Classification("INTERNAL") {
		t.Fatalf("classification = %q, want INTERNAL", got[0].Classification)
	}
	if len(got[0].Tags) != 2 || got[0].Tags[0] != "tag-a" || got[0].Tags[1] != "tag-b" {
		t.Fatalf("tags = %#v, want [tag-a tag-b]", got[0].Tags)
	}
	if got[0].DocumentCode != "QMS-001" || got[0].DocumentSequence != 7 {
		t.Fatalf("document identity = (%q,%d), want (QMS-001,7)", got[0].DocumentCode, got[0].DocumentSequence)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestListDocumentsLeavesBusinessUnitEmptyWhenUnset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id",
		"name",
		"status",
		"profile_code_snapshot",
		"family_code",
		"process_area_code_snapshot",
		"subject_code",
		"business_unit",
		"department_code",
		"classification",
		"tags",
		"created_by",
		"code",
		"sequence_num",
		"effective_from",
		"effective_to",
		"created_at",
	}).AddRow(
		"doc-2",
		"Instruction",
		"ACTIVE",
		"profile-b",
		"family-b",
		"quality",
		"",
		"",
		"qa",
		"",
		`[]`,
		"user-2",
		"QMS-002",
		8,
		nil,
		nil,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
	)
	mock.ExpectQuery(regexp.QuoteMeta("LIMIT $15 OFFSET $16")).
		WithArgs("tenant-1", "", "", "", "", "", "", "", "", "", "", "", nil, nil, 20, 0).
		WillReturnRows(rows)

	got, err := NewReader(db).ListDocuments(context.Background(), searchdomain.Query{TenantID: "tenant-1"}, 20, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("documents = %d, want 1", len(got))
	}
	if got[0].BusinessUnit != "" {
		t.Fatalf("business unit = %q, want empty when unset", got[0].BusinessUnit)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
