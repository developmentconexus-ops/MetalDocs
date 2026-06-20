package v2documents

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	searchdomain "metaldocs/internal/modules/search/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// TestListDocuments_EscapesLikeWildcards guards B6 / CWE-943: the $2 text
// argument must be LIKE-escaped and the SQL predicate must carry ESCAPE so
// Postgres interprets the backslash escapes correctly.
func TestListDocuments_EscapesLikeWildcards(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(readerCols)

	// The regexp must require that the ESCAPE token follows the LIKE predicate.
	// Pipe | and $ are regex metacharacters — escape them with \\.
	// We match a minimal substring that proves the ESCAPE clause is present.
	mock.ExpectQuery(`LIKE '%' \|\| \$2 \|\| '%' ESCAPE`).
		WithArgs(
			"tenant-1",  // $1 tenant_id
			"a\\%b\\_c", // $2 escaped+lowercased: LikeEscape("a%b_c") = "a\%b\_c"
			"",          // $3 status
			"",          // $4 profile_code
			"",          // $5 family
			"",          // $6 process_area
			"",          // $7 department
			"",          // $8 owner
			nil,         // $9 expiry_before
			nil,         // $10 expiry_after
			20,          // $11 limit
			0,           // $12 offset
			"user-1",    // $13 actor
			sqlmock.AnyArg(), // $14 family-filter codes (pq.Array, empty here)
		).
		WillReturnRows(rows)

	_, err = NewReader(db, taxonomydomain.NoopFamilyCodeResolver{}).ListDocuments(
		context.Background(),
		searchdomain.Query{TenantID: "tenant-1", Text: "a%b_c", ActorUserID: "user-1"},
		20, 0,
	)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
