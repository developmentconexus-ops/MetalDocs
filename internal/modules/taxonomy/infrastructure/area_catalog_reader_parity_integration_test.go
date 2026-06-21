//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"

	"metaldocs/tests/integration/testdb"
)

// F2.3 / B7 + B8 parity gate (D6: parity-before-delete). Proves the taxonomy
// AreaCatalogReader port returns the EXACT result the deleted raw reads produced:
//
//	B7 (in-tx, documents create): SELECT name FROM metaldocs.document_process_areas
//	    WHERE tenant_id=$1::uuid AND code=$2   (sql.ErrNoRows-tolerant)
//	B8 (off-tx, iam pre-invite):  SELECT EXISTS(SELECT 1 FROM metaldocs.document_process_areas
//	    WHERE tenant_id=$1::uuid AND code=$2)
//
// The raw baselines below are verbatim the SQL that lived at the B7/B8 call sites
// before the port replaced them.

func rawAreaName(t *testing.T, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, tenantID, code string) (string, bool) {
	t.Helper()
	var name sql.NullString
	err := q.QueryRowContext(context.Background(),
		`SELECT name FROM metaldocs.document_process_areas WHERE tenant_id=$1::uuid AND code=$2`,
		tenantID, code,
	).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false
		}
		t.Fatalf("raw area name: %v", err)
	}
	return name.String, true
}

func rawAreaExists(t *testing.T, db *sql.DB, tenantID, code string) bool {
	t.Helper()
	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.document_process_areas
  WHERE tenant_id = $1::uuid
    AND code = $2
)`, tenantID, code).Scan(&exists); err != nil {
		t.Fatalf("raw area exists: %v", err)
	}
	return exists
}

func TestAreaCatalogReader_AreaNameParityWithRaw(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	port := NewAreaCatalogReaderPG()
	tax := testdb.NewTaxonomy(t, db)
	absentCode := "absent-" + uuid.NewString()[:8]

	// B7 runs INSIDE the create tx — exercise the port with a *sql.Tx executor.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	t.Run("present_area", func(t *testing.T) {
		rawName, rawFound := rawAreaName(t, tx, tax.TenantID, tax.ProcessAreaCode)
		gotName, gotFound, err := port.AreaName(ctx, tx, tax.TenantID, tax.ProcessAreaCode)
		if err != nil {
			t.Fatalf("port AreaName: %v", err)
		}
		if rawFound != gotFound || rawName != gotName {
			t.Fatalf("present_area: raw=(%q,%v) port=(%q,%v)", rawName, rawFound, gotName, gotFound)
		}
		if !gotFound || gotName != tax.ProcessAreaCode { // factory sets name = code
			t.Fatalf("present_area: expected found name=%q, got (%q,%v)", tax.ProcessAreaCode, gotName, gotFound)
		}
	})

	t.Run("absent_area", func(t *testing.T) {
		rawName, rawFound := rawAreaName(t, tx, tax.TenantID, absentCode)
		gotName, gotFound, err := port.AreaName(ctx, tx, tax.TenantID, absentCode)
		if err != nil {
			t.Fatalf("port AreaName: %v", err)
		}
		if rawFound != gotFound || rawName != gotName {
			t.Fatalf("absent_area: raw=(%q,%v) port=(%q,%v)", rawName, rawFound, gotName, gotFound)
		}
		if gotFound {
			t.Fatalf("absent_area: expected found=false, got (%q,%v)", gotName, gotFound)
		}
	})
}

func TestAreaCatalogReader_AreaExistsParityWithRaw(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	port := NewAreaCatalogReaderPG()
	tax := testdb.NewTaxonomy(t, db)
	otherTenant := uuid.NewString()
	absentCode := "absent-" + uuid.NewString()[:8]

	cases := []struct {
		name     string
		tenantID string
		code     string
	}{
		{"present", tax.TenantID, tax.ProcessAreaCode},
		{"absent_code", tax.TenantID, absentCode},
		{"wrong_tenant", otherTenant, tax.ProcessAreaCode},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := rawAreaExists(t, db, tc.tenantID, tc.code)
			got, err := port.AreaExists(ctx, db, tc.tenantID, tc.code)
			if err != nil {
				t.Fatalf("port AreaExists: %v", err)
			}
			if want != got {
				t.Fatalf("%s: raw=%v port=%v", tc.name, want, got)
			}
		})
	}
}
