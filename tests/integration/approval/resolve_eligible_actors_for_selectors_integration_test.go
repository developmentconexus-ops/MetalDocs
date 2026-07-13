//go:build integration
// +build integration

// Package approval_test — unit 3.2 slice 3 (M4 ActorSelector, B3 resolver)
// real-Postgres parity + union coverage for
// postgresApprovalRepository.ResolveEligibleActorsForSelectors
// (docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md §4).
//
//  1. Parity: a stage whose selectors are exactly one
//     role_in_fixed_area(role, area) resolves to the IDENTICAL set as the
//     legacy flat ResolveEligibleActors(area, role) — the backfilled-route
//     invariant slice 3 must hold before migration 0304 can drop the flat
//     columns.
//  2. Union: two role_in_fixed_area selectors (different roles, overlapping
//     user) resolve to the deduped union of their individual pools.
package approval_test

import (
	"context"
	"database/sql"
	"sort"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// seedProcessArea inserts a public.document_process_areas row (PK is the
// global `code`, so distinct codes are needed across tenants — mirrors
// eligible_actors_view_parity_integration_test.go's seedFixedArea).
func seedProcessArea(t *testing.T, db *sql.DB, tenantID, code, name string) {
	t.Helper()
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ($1, $2::uuid, $3)
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			code, tenantID, name,
		)
		return err
	})
}

// grantActiveArea inserts an active (no effective_to) public.user_process_areas
// row, granting userID role in area for tenantID.
func grantActiveArea(t *testing.T, db *sql.DB, tenantID, userID, area, role string) {
	t.Helper()
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1,$2::uuid,$3,$4,$5)`,
			userID, tenantID, area, role, time.Now().UTC().Add(-time.Hour),
		)
		return err
	})
}

func sortedIDs(in []string) []string {
	cp := append([]string(nil), in...)
	sort.Strings(cp)
	return cp
}

// TestResolveEligibleActorsForSelectors_ParityWithFlatResolver proves that a
// single role_in_fixed_area selector resolves to the identical pool the
// legacy flat ResolveEligibleActors(area, role) resolves to — the invariant
// slice 6 / migration 0304 requires before the flat route-config columns can
// be dropped.
func TestResolveEligibleActorsForSelectors_ParityWithFlatResolver(t *testing.T) {
	ctx := context.Background()
	database, _ := testdb.Open(t)
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})

	tenantID := testdb.NewTenant(t, database).ID
	const area = "quality"
	seedProcessArea(t, database, tenantID, area, "Quality")

	inPool1 := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	inPool2 := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	outsideRole := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID

	grantActiveArea(t, database, tenantID, inPool1, area, "qms_admin")
	grantActiveArea(t, database, tenantID, inPool2, area, "qms_admin")
	// Outside the pool: wrong role in the same area.
	grantActiveArea(t, database, tenantID, outsideRole, area, "approver")

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	selectors := []domain.ActorSelector{
		{Kind: domain.SelectorRoleInFixedArea, Role: "qms_admin", AreaCode: area},
	}
	got, err := repo.ResolveEligibleActorsForSelectors(ctx, tx, tenantID, selectors, "")
	if err != nil {
		t.Fatalf("ResolveEligibleActorsForSelectors: %v", err)
	}
	want, err := repo.ResolveEligibleActors(ctx, tx, tenantID, area, "qms_admin")
	if err != nil {
		t.Fatalf("ResolveEligibleActors: %v", err)
	}

	got, want = sortedIDs(got), sortedIDs(want)
	if len(got) != 2 {
		t.Fatalf("got = %v, want exactly 2 eligible actors", got)
	}
	if len(got) != len(want) {
		t.Fatalf("selector-union resolver drift: selectors=%v flat=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("selector-union resolver drift: selectors=%v flat=%v", got, want)
		}
	}
}

// TestResolveEligibleActorsForSelectors_UnionAcrossSelectors proves that two
// role_in_fixed_area selectors (different roles, overlapping user) resolve to
// the deduped union of their individual pools.
func TestResolveEligibleActorsForSelectors_UnionAcrossSelectors(t *testing.T) {
	ctx := context.Background()
	database, _ := testdb.Open(t)
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})

	tenantID := testdb.NewTenant(t, database).ID
	const areaQ = "quality"
	const areaS = "safety"
	seedProcessArea(t, database, tenantID, areaQ, "Quality")
	seedProcessArea(t, database, tenantID, areaS, "Safety")

	qualityOnly := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	safetyOnly := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	// overlap holds BOTH roles — appears in both selectors' pools, must be
	// deduped to a single entry in the union.
	overlap := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID

	grantActiveArea(t, database, tenantID, qualityOnly, areaQ, "qms_admin")
	grantActiveArea(t, database, tenantID, overlap, areaQ, "qms_admin")
	grantActiveArea(t, database, tenantID, safetyOnly, areaS, "area_admin")
	grantActiveArea(t, database, tenantID, overlap, areaS, "area_admin")

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	selectors := []domain.ActorSelector{
		{Kind: domain.SelectorRoleInFixedArea, Role: "qms_admin", AreaCode: areaQ},
		{Kind: domain.SelectorRoleInFixedArea, Role: "area_admin", AreaCode: areaS},
	}
	got, err := repo.ResolveEligibleActorsForSelectors(ctx, tx, tenantID, selectors, "")
	if err != nil {
		t.Fatalf("ResolveEligibleActorsForSelectors: %v", err)
	}
	got = sortedIDs(got)
	want := sortedIDs([]string{qualityOnly, safetyOnly, overlap})

	if len(got) != 3 {
		t.Fatalf("union must dedup overlap to a single entry: got %v, want 3 distinct ids", got)
	}
	if len(got) != len(want) {
		t.Fatalf("got = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got = %v, want %v", got, want)
		}
	}
}
