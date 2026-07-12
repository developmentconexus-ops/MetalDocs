//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// P3.S2b-3b-i: template approval route stages carry no process area and are
// created with stage AreaCode = 'tenant' — the reserved area-blind sentinel
// (baseline CHECK area_code_not_tenant on metaldocs.document_process_areas
// forbids a real area ever being named 'tenant', so user_process_areas.area_code
// — FK'd to document_process_areas(tenant_id, code) — can never literally hold
// 'tenant' either). ResolveEligibleActors must therefore treat areaCode='tenant'
// as "role held in ANY area for this tenant", mirroring the authz predicate at
// internal/modules/iam/authz/authz.go:155 ($2 = 'tenant' OR upa.area_code = $2).
//
// This test proves:
//  1. area-blind mode ('tenant' sentinel) finds a user who holds the role in a
//     real area (there is no other way to represent "tenant-wide" given the FK)
//  2. area-scoped mode (a real area code) is unchanged: only members of that
//     exact area are returned, and a member of a different area is excluded.
func TestResolveEligibleActors_AreaBlindForTemplateStages(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	const areaQ, areaX = "quality", "safety"
	seedFixedArea(t, db, tenantID, areaQ, "Quality")
	seedFixedArea(t, db, tenantID, areaX, "Safety")

	// tenantWideApprover holds the role only in area "quality" — under the
	// area-blind ('tenant') predicate this must still resolve, because no row
	// can ever carry area_code='tenant' literally (FK to document_process_areas,
	// which forbids code='tenant' via area_code_not_tenant). Area-blind means
	// "any area", not "a literal tenant row".
	tenantWideApprover := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	// scopedOnlyInX holds the role only in area "safety" — must be excluded from
	// an area-scoped query for "quality", proving the real-area path is unchanged.
	scopedOnlyInX := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	now := time.Now().UTC()
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		ins := func(user, area, role string) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO public.user_process_areas
				   (user_id, tenant_id, area_code, role, effective_from)
				 VALUES ($1,$2::uuid,$3,$4,$5)`,
				user, tenantID, area, role, now.Add(-time.Hour),
			)
			return err
		}
		if err := ins(tenantWideApprover, areaQ, "approver"); err != nil {
			return err
		}
		return ins(scopedOnlyInX, areaX, "approver")
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// 1. Area-blind (template sentinel): must find the tenant-wide approver
	// regardless of which real area their grant lives in.
	blind, err := repo.ResolveEligibleActors(ctx, tx, tenantID, "tenant", "approver")
	if err != nil {
		t.Fatalf("ResolveEligibleActors(tenant): %v", err)
	}
	blindSet := map[string]bool{}
	for _, id := range blind {
		blindSet[id] = true
	}
	if !blindSet[tenantWideApprover] {
		t.Fatalf("area-blind query must find approver granted in a real area, got %v", blind)
	}
	if !blindSet[scopedOnlyInX] {
		t.Fatalf("area-blind query must find ALL approvers regardless of area, got %v", blind)
	}

	// 2. Area-scoped (document path, real area code): unchanged behavior —
	// only members of "quality" are returned; the "safety"-only member is excluded.
	scoped, err := repo.ResolveEligibleActors(ctx, tx, tenantID, areaQ, "approver")
	if err != nil {
		t.Fatalf("ResolveEligibleActors(quality): %v", err)
	}
	if len(scoped) != 1 || scoped[0] != tenantWideApprover {
		t.Fatalf("area-scoped query must return only quality-area members, got %v", scoped)
	}
}
