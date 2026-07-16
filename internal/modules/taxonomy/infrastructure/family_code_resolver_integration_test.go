//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

// seedFamily ensures a document_families row exists for code (FK parent for
// document_profiles). The INSERT requires taxonomy.manage tripwire.
func seedFamily(t *testing.T, db *sql.DB, code string) {
	t.Helper()
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.document_families (code, name)
			 VALUES ($1, $1) ON CONFLICT (code) DO NOTHING`,
			code,
		)
		return err
	})
}

// seedProfile inserts a document_profiles row. archivedAt = true sets archived_at = now().
// Requires taxonomy.manage tripwire.
func seedProfile(t *testing.T, db *sql.DB, code, tenantID, familyCode string, archived bool) {
	t.Helper()
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		if archived {
			_, err := tx.ExecContext(context.Background(),
				`INSERT INTO metaldocs.document_profiles
				   (code, tenant_id, family_code, name, review_interval_days, alias, archived_at)
				 VALUES ($1, $2::uuid, $3, $1, 365, $1, now())
				 ON CONFLICT (tenant_id, code) DO NOTHING`,
				code, tenantID, familyCode,
			)
			return err
		}
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.document_profiles
			   (code, tenant_id, family_code, name, review_interval_days, alias)
			 VALUES ($1, $2::uuid, $3, $1, 365, $1)
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			code, tenantID, familyCode,
		)
		return err
	})
}

// uniqueCode returns a valid profile code from a test-local counter + tenant suffix
// satisfying ^[a-z][a-z0-9_-]{1,63}$.
func uniqueCode(prefix, suffix string, n int) string {
	return fmt.Sprintf("%s%s-%02d", prefix, suffix[:4], n)
}

// TestFamilyCodeResolverRepository_TenantAndSentinelOwnershipResolve: a code owned
// by the queried tenant and a (distinct) code owned by the sentinel both resolve in
// one batch — proving the `tenant_id IN (tenant, sentinel)` arm. Distinct codes
// here just isolate that arm; TestFamilyCodeResolverRepository_TenantPrecedenceOverSentinel
// below covers the same-code tie-break directly.
func TestFamilyCodeResolverRepository_TenantAndSentinelOwnershipResolve(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	shortID := strings.ReplaceAll(tenantID, "-", "")[:6]

	codeTenant := domain.ProfileCode("ptenant-" + shortID)
	codeGlobal := domain.ProfileCode("pglobal-" + shortID)
	famTenant := "fam-tenant-" + shortID
	famGlobal := "fam-global-" + shortID

	seedFamily(t, db, famTenant)
	seedFamily(t, db, famGlobal)
	seedProfile(t, db, string(codeTenant), tenantID, famTenant, false)         // tenant-owned
	seedProfile(t, db, string(codeGlobal), sentinelTenantID, famGlobal, false) // sentinel-owned

	repo := NewFamilyCodeResolverRepository(db)
	got, err := repo.ResolveFamilyCodes(ctx, tenantID, []domain.ProfileCode{codeTenant, codeGlobal})
	if err != nil {
		t.Fatalf("ResolveFamilyCodes: %v", err)
	}
	if got[codeTenant] != domain.FamilyCode(famTenant) {
		t.Fatalf("tenant-owned: got family %q, want %q", got[codeTenant], famTenant)
	}
	if got[codeGlobal] != domain.FamilyCode(famGlobal) {
		t.Fatalf("sentinel-owned: got family %q, want %q", got[codeGlobal], famGlobal)
	}
}

// TestFamilyCodeResolverRepository_CompositeTenantCodePrimaryKey pins the schema
// invariant that makes the resolver's tenant-vs-sentinel ORDER BY tie-break live
// behavior: the primary key of metaldocs.document_profiles is (tenant_id, code)
// (migration 0308), so the same code can be defined under two tenants (or a
// tenant and the sentinel) simultaneously. See
// TestFamilyCodeResolverRepository_TenantPrecedenceOverSentinel for the tie-break
// itself. (ADR 0038.)
func TestFamilyCodeResolverRepository_CompositeTenantCodePrimaryKey(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, `
SELECT a.attname
FROM pg_index i
JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY (i.indkey)
WHERE i.indrelid = 'metaldocs.document_profiles'::regclass
  AND i.indisprimary
ORDER BY a.attname`)
	if err != nil {
		t.Fatalf("query primary key columns: %v", err)
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatalf("scan: %v", err)
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	want := []string{"code", "tenant_id"}
	if len(cols) != len(want) || cols[0] != want[0] || cols[1] != want[1] {
		t.Fatalf("document_profiles primary key = %v, want %v (composite PK; resolver precedence tie-break is live behavior)", cols, want)
	}
}

// TestFamilyCodeResolverRepository_TenantPrecedenceOverSentinel: the same code
// defined under both the queried tenant and the sentinel resolves to the
// tenant-owned family — proving the ORDER BY tenant-over-sentinel tie-break,
// which the composite (tenant_id, code) PK makes reachable.
func TestFamilyCodeResolverRepository_TenantPrecedenceOverSentinel(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	shortID := strings.ReplaceAll(tenantID, "-", "")[:6]

	sharedCode := domain.ProfileCode("pshared-" + shortID)
	famTenant := "fam-shtenant-" + shortID
	famSentinel := "fam-shsentinel-" + shortID

	seedFamily(t, db, famTenant)
	seedFamily(t, db, famSentinel)
	seedProfile(t, db, string(sharedCode), sentinelTenantID, famSentinel, false) // sentinel-owned
	seedProfile(t, db, string(sharedCode), tenantID, famTenant, false)           // tenant-owned, same code

	repo := NewFamilyCodeResolverRepository(db)
	got, err := repo.ResolveFamilyCodes(ctx, tenantID, []domain.ProfileCode{sharedCode})
	if err != nil {
		t.Fatalf("ResolveFamilyCodes: %v", err)
	}
	if got[sharedCode] != domain.FamilyCode(famTenant) {
		t.Fatalf("tenant precedence: got family %q, want tenant-owned %q", got[sharedCode], famTenant)
	}
}

// TestFamilyCodeResolverRepository_SentinelFallback: code only under sentinel →
// still resolves for any tenant; code under neither → absent.
func TestFamilyCodeResolverRepository_SentinelFallback(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	shortID := strings.ReplaceAll(tenantID, "-", "")[:6]

	codeGlobal := domain.ProfileCode("pglobal-" + shortID)
	codeAbsent := domain.ProfileCode("pabsent-" + shortID)
	famGlobal := "fam-sentfl-" + shortID

	seedFamily(t, db, famGlobal)
	// Only sentinel row.
	seedProfile(t, db, string(codeGlobal), sentinelTenantID, famGlobal, false)

	repo := NewFamilyCodeResolverRepository(db)

	// codeGlobal via sentinel → should resolve.
	got, err := repo.ResolveFamilyCodes(ctx, tenantID, []domain.ProfileCode{codeGlobal, codeAbsent})
	if err != nil {
		t.Fatalf("ResolveFamilyCodes: %v", err)
	}
	if got[codeGlobal] != domain.FamilyCode(famGlobal) {
		t.Fatalf("sentinel fallback: got %q, want %q", got[codeGlobal], famGlobal)
	}
	if _, ok := got[codeAbsent]; ok {
		t.Fatalf("absent code must not appear in map, got %q", got[codeAbsent])
	}
	if got == nil {
		t.Fatal("returned map must be non-nil")
	}
}

// TestFamilyCodeResolverRepository_ArchivedIncluded: an archived profile still resolves.
func TestFamilyCodeResolverRepository_ArchivedIncluded(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	shortID := strings.ReplaceAll(tenantID, "-", "")[:6]

	code := domain.ProfileCode("parch-" + shortID)
	fam := "fam-arch-" + shortID

	seedFamily(t, db, fam)
	// Seed as archived.
	seedProfile(t, db, string(code), tenantID, fam, true)

	repo := NewFamilyCodeResolverRepository(db)
	got, err := repo.ResolveFamilyCodes(ctx, tenantID, []domain.ProfileCode{code})
	if err != nil {
		t.Fatalf("ResolveFamilyCodes: %v", err)
	}
	if got[code] != domain.FamilyCode(fam) {
		t.Fatalf("archived profile: got family %q, want %q", got[code], fam)
	}
}

// TestFamilyCodeResolverRepository_ExactCaseCode: lowercase code seeded; querying
// an uppercased variant must return absent (exact-case match).
func TestFamilyCodeResolverRepository_ExactCaseCode(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	shortID := strings.ReplaceAll(tenantID, "-", "")[:6]

	// Must satisfy ^[a-z][a-z0-9_-]{1,63}$.
	lowerCode := "pcase-" + shortID
	upperQuery := domain.ProfileCode(strings.ToUpper(lowerCode))
	fam := "fam-case-" + shortID

	seedFamily(t, db, fam)
	seedProfile(t, db, lowerCode, tenantID, fam, false)

	repo := NewFamilyCodeResolverRepository(db)
	got, err := repo.ResolveFamilyCodes(ctx, tenantID, []domain.ProfileCode{upperQuery})
	if err != nil {
		t.Fatalf("ResolveFamilyCodes: %v", err)
	}
	if _, ok := got[upperQuery]; ok {
		t.Fatalf("exact-case: upper-cased query should be absent from map, got %q", got[upperQuery])
	}
}

// TestFamilyCodeResolverRepository_ProfileCodesForFamily: case-insensitive family
// filter + sentinel precedence + empty-family guard + ordering.
func TestFamilyCodeResolverRepository_ProfileCodesForFamily(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	shortID := strings.ReplaceAll(tenantID, "-", "")[:6]

	// Two profiles in same family, one in another.
	fam := "fam-pfor-" + shortID
	famOther := "fam-other-" + shortID

	// Profile codes: must be ordered lexicographically for assertion.
	codeA := "pfor-" + shortID + "-aa"
	codeB := "pfor-" + shortID + "-bb"
	codeOther := "pother-" + shortID

	// codeA, codeB → fam; codeOther → famOther. Distinct codes here (not required by
	// the schema, just this test's shape); a sentinel-owned member also counts
	// toward the family.
	seedFamily(t, db, fam)
	seedFamily(t, db, famOther)
	seedProfile(t, db, codeA, tenantID, fam, false)
	seedProfile(t, db, codeB, sentinelTenantID, fam, false) // sentinel-owned member of fam
	seedProfile(t, db, codeOther, tenantID, famOther, false)

	repo := NewFamilyCodeResolverRepository(db)

	// Query with upper-cased family (case-insensitive match expected).
	upper := domain.FamilyCode(strings.ToUpper(fam))
	codes, err := repo.ProfileCodesForFamily(ctx, tenantID, upper)
	if err != nil {
		t.Fatalf("ProfileCodesForFamily: %v", err)
	}

	// Should return codeA and codeB only (codeOther in famOther); tenant wins for codeA.
	want := []domain.ProfileCode{
		domain.ProfileCode(codeA),
		domain.ProfileCode(codeB),
	}
	if len(codes) != len(want) {
		t.Fatalf("ProfileCodesForFamily: got %v (len %d), want %v (len %d)", codes, len(codes), want, len(want))
	}
	for i, w := range want {
		if codes[i] != w {
			t.Fatalf("ProfileCodesForFamily[%d]: got %q, want %q", i, codes[i], w)
		}
	}

	// Empty family → empty non-nil slice.
	empty, err := repo.ProfileCodesForFamily(ctx, tenantID, "")
	if err != nil {
		t.Fatalf("ProfileCodesForFamily(empty): %v", err)
	}
	if empty == nil {
		t.Fatal("empty family must return non-nil slice")
	}
	if len(empty) != 0 {
		t.Fatalf("empty family must return empty slice, got %v", empty)
	}
}
