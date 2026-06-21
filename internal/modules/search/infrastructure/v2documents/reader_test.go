package v2documents

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	searchdomain "metaldocs/internal/modules/search/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// readerCols are the columns ListDocuments projects, in order. Column 5 is the raw
// profile-code join key (COALESCE(d.profile_code_snapshot, cd.profile_code)); the
// document's family_code is resolved from this key by the FamilyCodeResolver port
// after the scan (ADR 0038), not projected by SQL.
var readerCols = []string{
	"id", "name", "status", "profile_code_snapshot", "profile_key",
	"process_area_code_snapshot", "department_code", "created_by",
	"code", "sequence_num", "effective_from", "effective_to", "created_at",
}

// stubFamilyResolver is an in-test FamilyCodeResolver: it maps a fixed set of
// profile-code keys to families for the projection path and never matches a family
// filter (these unit tests exercise the unfiltered list path).
type stubFamilyResolver struct{ m map[string]string }

func (s stubFamilyResolver) ResolveFamilyCodes(_ context.Context, _ string, codes []taxonomydomain.ProfileCode) (map[taxonomydomain.ProfileCode]taxonomydomain.FamilyCode, error) {
	out := make(map[taxonomydomain.ProfileCode]taxonomydomain.FamilyCode, len(codes))
	for _, c := range codes {
		if f, ok := s.m[string(c)]; ok {
			out[c] = taxonomydomain.FamilyCode(f)
		}
	}
	return out, nil
}

func (stubFamilyResolver) ProfileCodesForFamily(_ context.Context, _ string, _ taxonomydomain.FamilyCode) ([]taxonomydomain.ProfileCode, error) {
	return []taxonomydomain.ProfileCode{}, nil
}

func TestListDocumentsFiltersByTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(readerCols).AddRow(
		"doc-1", "Manual", "ACTIVE", "profile-a", "profile-a",
		"quality", "qa", "user-1",
		"QMS-001", 7, nil, nil,
		time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	)
	// The query must enforce per-document visibility through controlleddocuments'
	// PUBLISHED contract (v_cd_grantee, ADR-0039 D3a / M4/F4.3) — assert the grantee
	// EXISTS leg is present, then assert the actor is bound at $13.
	mock.ExpectQuery("v_cd_grantee[\\s\\S]*LIMIT \\$11 OFFSET \\$12").
		WithArgs("tenant-1", "", "", "", "", "", "", "", nil, nil, 20, 0, "user-1", sqlmock.AnyArg()).
		WillReturnRows(rows)

	resolver := stubFamilyResolver{m: map[string]string{"profile-a": "family-a"}}
	got, err := NewReader(db, resolver).ListDocuments(context.Background(), searchdomain.Query{TenantID: "tenant-1", ActorUserID: "user-1"}, 20, 0)
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
		"doc-2", "Instruction", "ACTIVE", "profile-b", "profile-b",
		"quality", "", "user-2",
		"QMS-002", 8, nil, nil,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
	)
	mock.ExpectQuery("v_cd_grantee[\\s\\S]*LIMIT \\$11 OFFSET \\$12").
		WithArgs("tenant-1", "", "", "", "", "", "", "", nil, nil, 20, 0, "user-9", sqlmock.AnyArg()).
		WillReturnRows(rows)

	resolver := stubFamilyResolver{m: map[string]string{"profile-b": "family-b"}}
	got, err := NewReader(db, resolver).ListDocuments(context.Background(), searchdomain.Query{TenantID: "tenant-1", ActorUserID: "user-9"}, 20, 0)
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
