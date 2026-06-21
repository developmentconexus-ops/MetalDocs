package infrastructure

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// ptr is a local helper (not re-declared if already present in the package's
// test files — check: the existing repository_test.go does not define one).
func ptr(s string) *string { return &s }

// cdListCols matches the SELECT projection in List.
var cdListCols = []string{
	"id", "tenant_id", "profile_code", "process_area_code", "department_code",
	"code", "sequence_num", "title", "owner_user_id", "override_template_version_id",
	"visibility_scope", "status", "created_at", "updated_at",
}

// TestList_EscapesLikeWildcards guards B6 / CWE-943: the Query bound argument
// must be LIKE-escaped and the SQL predicate must carry ESCAPE on both ILIKEs.
func TestList_EscapesLikeWildcards(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresControlledDocumentRepository(db, documentsdomain.NoopActiveInstanceReader{})

	// With only Query set, args order is:
	//   $1 = tenantID
	//   $2 = "%"+LikeEscape("a%b")+"%" = "%a\%b%"
	//   $3 = limit+1 probe (ClampLimit(0)=20 → 21)
	mock.ExpectQuery(`code ILIKE \$\d+ ESCAPE`).
		WithArgs(
			"tenant-1", // $1
			"%a\\%b%",  // $2 — LikeEscape("a%b") = "a\%b", wrapped in %…%
			21,         // $3 — limit+1 probe (ClampLimit(0)=20 → 21)
		).
		WillReturnRows(sqlmock.NewRows(cdListCols)) // zero rows → hydrateVisibilityGrants no-ops

	filter := controlleddocumentsdomain.CDFilter{Query: ptr("a%b")}
	items, hasMore, err := repo.List(context.Background(), "tenant-1", filter)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %d, want 0", len(items))
	}
	if hasMore {
		t.Fatal("hasMore = true, want false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
