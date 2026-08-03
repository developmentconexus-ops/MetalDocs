//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

// TestMigration0308_DocumentProfilesTenantPK proves both directions of the
// 0308 tenant-scoped composite PK migration (unit 4.5) against a real,
// per-test-isolated database:
//
//  1. Positive path (bootstrap proof): the curated bootstrap this database
//     was cloned from already applied baseline + every forward migration
//     through 0308, so before this test touches anything: the 4 dead
//     profile-adjacent tables must be gone, document_profiles' PK must be
//     exactly (tenant_id, code), ux_document_profiles_tenant_code must be
//     gone, the 3 composite FKs must exist and reference
//     document_profiles(tenant_id, code), and the '0308' ledger row must be
//     present exactly once.
//  2. Behavioral proof: two rows sharing the same code but different
//     tenant_id must both insert cleanly (tenant-scoped uniqueness); the
//     same tenant_id+code pair must collide with 23505.
//
// The former replay tests in this file (TestMigration0308_ReRunIsCleanNoOp and
// TestMigration0308_MalformedPrestateAborts, both of which re-executed the
// committed migration file bytes) were deleted with the 2026-07-29 fold: 0308
// is folded into db/baseline/0001_current_schema.sql and archived under
// archive/migrations/post-baseline-2026-07-fold/, so it is never applied to any
// database again and its replay/abort behaviour guards nothing.
func TestMigration0308_DocumentProfilesTenantPK(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	// (1a) The 4 dead tables must be gone.
	for _, tbl := range []string{
		"metaldocs.document_profile_governance",
		"metaldocs.document_profile_schema_versions",
		"metaldocs.document_sequences",
		"metaldocs.document_profile_template_defaults",
	} {
		if got := regclass(t, ctx, db, tbl); got.Valid {
			t.Fatalf("expected %s to be dropped by migration 0308, but it exists as %q", tbl, got.String)
		}
	}

	// (1b) document_profiles PK must be exactly (tenant_id, code).
	assertPrimaryKeyColumns(t, ctx, db, "metaldocs.document_profiles", []string{"tenant_id", "code"})

	// (1c) The redundant unique index must be gone.
	if got := regclassIndex(t, ctx, db, "metaldocs.ux_document_profiles_tenant_code"); got.Valid {
		t.Fatalf("expected ux_document_profiles_tenant_code to be dropped by migration 0308, but it exists as %q", got.String)
	}

	// (1d) The 3 composite FKs must exist and reference document_profiles(tenant_id, code).
	for _, fk := range []struct {
		name  string
		table string
	}{
		{"approval_routes_document_profile_fk", "public.approval_routes"},
		{"cd_sequence_counters_tenant_id_profile_code_fkey", "public.cd_sequence_counters"},
		{"controlled_documents_tenant_id_profile_code_fkey", "public.controlled_documents"},
	} {
		assertForeignKeyTargetsDocumentProfiles(t, ctx, db, fk.name, fk.table)
	}

	// (1e) Ledger row present exactly once from bootstrap.
	if n := ledgerCount(t, ctx, db, "0308"); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0308' after bootstrap, got %d", n)
	}

	// (2) Behavioral: same code, different tenant_id -> both succeed;
	// same tenant_id + code -> 23505.
	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	const profileCode = "shared-pk-probe"
	const areaCode = "shared-pk-probe-area"
	testdb.SeedGovernedTaxonomy(t, db, tenantA.ID, profileCode, areaCode)
	testdb.SeedGovernedTaxonomy(t, db, tenantB.ID, profileCode, areaCode)

	// SeedGovernedTaxonomy is itself idempotent (ON CONFLICT DO NOTHING), so
	// re-running it for both tenants above already proves the composite PK
	// admits the same code under two different tenants. Prove the negative
	// (same tenant_id + code collides) directly against the tripwire-guarded
	// table using the same authorized-write path the fixture uses.
	var dupErr error
	err := withTaxonomyCapability(ctx, t, db, tenantA.ID, func(tx *sql.Tx) error {
		_, e := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_profiles
			    (code, tenant_id, family_code, name, review_interval_days, alias)
			 VALUES ($1, $2::uuid, 'test-family', $1, 365, $1)`,
			profileCode, tenantA.ID,
		)
		dupErr = e
		return e
	})
	if err == nil {
		t.Fatal("expected duplicate (tenant_id, code) insert to fail with 23505, but it succeeded")
	}
	if !hasSQLState(dupErr, "23505") {
		t.Fatalf("expected SQLSTATE 23505 for duplicate (tenant_id, code), got: %v", dupErr)
	}
}

func regclass(t *testing.T, ctx context.Context, db *sql.DB, qualifiedName string) sql.NullString {
	t.Helper()
	var got sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass($1)::text`, qualifiedName).Scan(&got); err != nil {
		t.Fatalf("to_regclass(%s): %v", qualifiedName, err)
	}
	return got
}

func regclassIndex(t *testing.T, ctx context.Context, db *sql.DB, qualifiedName string) sql.NullString {
	t.Helper()
	return regclass(t, ctx, db, qualifiedName)
}

func ledgerCount(t *testing.T, ctx context.Context, db *sql.DB, version string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM public.schema_migrations WHERE version = $1`, version).Scan(&n); err != nil {
		t.Fatalf("count schema_migrations rows for %q: %v", version, err)
	}
	return n
}

// assertPrimaryKeyColumns asserts the given table's primary key is backed by
// exactly the given ordered column list.
func assertPrimaryKeyColumns(t *testing.T, ctx context.Context, db *sql.DB, qualifiedTable string, wantCols []string) {
	t.Helper()
	rows, err := db.QueryContext(ctx, `
		SELECT a.attname
		  FROM pg_index i
		  JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		 WHERE i.indrelid = $1::regclass
		   AND i.indisprimary
		 ORDER BY array_position(i.indkey, a.attnum)`, qualifiedTable)
	if err != nil {
		t.Fatalf("query primary key columns for %s: %v", qualifiedTable, err)
	}
	defer rows.Close()

	var gotCols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			t.Fatalf("scan pk column for %s: %v", qualifiedTable, err)
		}
		gotCols = append(gotCols, col)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate pk columns for %s: %v", qualifiedTable, err)
	}

	if len(gotCols) != len(wantCols) {
		t.Fatalf("%s primary key columns = %v, want %v", qualifiedTable, gotCols, wantCols)
	}
	for i, c := range wantCols {
		if gotCols[i] != c {
			t.Fatalf("%s primary key columns = %v, want %v", qualifiedTable, gotCols, wantCols)
		}
	}
}

// assertForeignKeyTargetsDocumentProfiles asserts the named FK constraint on
// the given table exists and references metaldocs.document_profiles(tenant_id, code).
func assertForeignKeyTargetsDocumentProfiles(t *testing.T, ctx context.Context, db *sql.DB, constraintName, qualifiedTable string) {
	t.Helper()
	var confrelid string
	var confCols string
	err := db.QueryRowContext(ctx, `
		SELECT c.confrelid::regclass::text,
		       (SELECT string_agg(a.attname, ',' ORDER BY array_position(c.confkey, a.attnum))
		          FROM pg_attribute a
		         WHERE a.attrelid = c.confrelid AND a.attnum = ANY(c.confkey))
		  FROM pg_constraint c
		 WHERE c.conname = $1
		   AND c.conrelid = $2::regclass
		   AND c.contype = 'f'`, constraintName, qualifiedTable).Scan(&confrelid, &confCols)
	if err != nil {
		t.Fatalf("query FK constraint %s on %s: %v", constraintName, qualifiedTable, err)
	}
	if confrelid != "metaldocs.document_profiles" && confrelid != "document_profiles" {
		t.Fatalf("FK %s on %s references %q, want metaldocs.document_profiles", constraintName, qualifiedTable, confrelid)
	}
	if confCols != "tenant_id,code" {
		t.Fatalf("FK %s on %s references columns %q, want \"tenant_id,code\"", constraintName, qualifiedTable, confCols)
	}
}

// withTaxonomyCapability runs fn inside a transaction that has asserted the
// taxonomy.manage capability transaction-locally via the production
// authz.SeedTxTenant helper, matching the tripwire (trg_require_cap_asserted)
// that guards metaldocs.document_profiles writes — see
// testdb.SeedGovernedTaxonomy (fixtures.go) for the same idiom.
func withTaxonomyCapability(ctx context.Context, t *testing.T, db *sql.DB, tenantID string, fn func(tx *sql.Tx) error) error {
	t.Helper()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	testdb.SetCapsOnTx(t, tx, `[{"cap":"taxonomy.manage"}]`)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func hasSQLState(err error, state string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == state
	}
	return strings.Contains(err.Error(), state)
}
