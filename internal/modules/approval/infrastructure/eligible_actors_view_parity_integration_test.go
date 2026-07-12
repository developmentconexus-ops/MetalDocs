//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"sort"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// F3.3 approval membership-read parity gate (M3 Category C, C3; ADR-0039 D3a, H-PRE-1).
//
// ResolveEligibleActors (:1133) resolves eligible approvers from iam's membership
// inside the caller's approval tx. The pre-F3.3 SQL read metaldocs.user_process_areas
// with an INTERVAL predicate (effective_from <= now() AND (effective_to IS NULL OR
// effective_to > now())). M3/F3.3 repoints it to the published view
// metaldocs.v_active_user_areas (which encodes effective_to IS NULL, ADR 0037 D1) and
// drops the temporal predicate. Under ADR-0037 Model A the interval form and
// effective_to IS NULL select the IDENTICAL set (effective_from is never future;
// effective_to > now() is the empty set), so this preserves behavior AND retires the
// Model-B interval leak.
//
// This test proves ZERO authz drift: repo.ResolveEligibleActors on a real *sql.Tx
// (post-repoint = view form) returns EXACTLY what a VERBATIM inline copy of the deleted
// interval SQL returns on the same tx. The revoked + wrong-role rows are the
// discriminators. H-PRE-1: both run as plain non-recording tx.QueryContext SELECTs.

func seedFixedArea(t *testing.T, db *sql.DB, tenantID, code, name string) {
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

// rawEligible runs the VERBATIM pre-F3.3 interval SQL on the caller's tx. The
// post-repoint repo.ResolveEligibleActors must equal this for the same arguments.
func rawEligible(t *testing.T, tx *sql.Tx, tenantID, areaCode, requiredRole string) []string {
	t.Helper()
	rows, err := tx.QueryContext(context.Background(),
		`SELECT user_id
		   FROM metaldocs.user_process_areas
		  WHERE tenant_id = $1::uuid
		    AND area_code = $2
		    AND role      = $3
		    AND effective_from <= now()
		    AND (effective_to IS NULL OR effective_to > now())`,
		tenantID, areaCode, requiredRole,
	)
	if err != nil {
		t.Fatalf("rawEligible: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("rawEligible scan: %v", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rawEligible rows: %v", err)
	}
	return out
}

func sortedCopy(in []string) []string {
	cp := append([]string(nil), in...)
	sort.Strings(cp)
	return cp
}

func TestResolveEligibleActors_ViewParityWithRaw(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	const areaQ, areaS = "quality", "safety"
	seedFixedArea(t, db, tenantID, areaQ, "Quality")
	seedFixedArea(t, db, tenantID, areaS, "Safety")

	eligibleA := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	eligibleB := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	wrongRole := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	revoked := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	wrongArea := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	// Wrong-tenant eligible row. document_process_areas PK is global `code`
	// (baseline 0001:2375), so the other tenant needs a distinct area code; the
	// tenant_id filter — not the area — is what must exclude this row.
	otherTenant := testdb.NewTenant(t, db)
	const areaQOther = "quality_ot"
	seedFixedArea(t, db, otherTenant.ID, areaQOther, "Quality (other tenant)")
	wrongTenant := testdb.NewUser(t, db, testdb.WithTenant(otherTenant.ID)).ID

	now := time.Now().UTC()
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		ins := func(user, tenant, area, role string, from time.Time) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO public.user_process_areas
				   (user_id, tenant_id, area_code, role, effective_from)
				 VALUES ($1,$2::uuid,$3,$4,$5)`,
				user, tenant, area, role, from,
			)
			return err
		}
		if err := ins(eligibleA, tenantID, areaQ, "qms_admin", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		if err := ins(eligibleB, tenantID, areaQ, "qms_admin", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		if err := ins(wrongRole, tenantID, areaQ, "approver", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		if err := ins(wrongArea, tenantID, areaS, "qms_admin", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		if err := ins(wrongTenant, otherTenant.ID, areaQOther, "qms_admin", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		// REVOKED qms_admin on quality (past effective_to, revoked_by set) — must be excluded.
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
			 VALUES ($1,$2::uuid,$3,'qms_admin',$4,$5,$1,$1)`,
			revoked, tenantID, areaQ, now.Add(-3*time.Hour), now.Add(-1*time.Hour),
		)
		return err
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got := sortedCopy(mustResolve(t, repo, ctx, tx, tenantID, areaQ, "qms_admin"))
	want := sortedCopy(rawEligible(t, tx, tenantID, areaQ, "qms_admin"))

	if len(got) != len(want) {
		t.Fatalf("eligible-actor drift: view=%v raw=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("eligible-actor drift: view=%v raw=%v", got, want)
		}
	}

	// Discriminators on the view-form result.
	set := map[string]bool{}
	for _, id := range got {
		set[id] = true
	}
	if !set[eligibleA] || !set[eligibleB] {
		t.Fatalf("both active qms_admin members must be eligible, got %v", got)
	}
	for name, id := range map[string]string{
		"revoked": revoked, "wrongRole": wrongRole, "wrongArea": wrongArea, "wrongTenant": wrongTenant,
	} {
		if set[id] {
			t.Fatalf("%s member must NOT be eligible, got %v", name, got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 eligible actors, got %d: %v", len(got), got)
	}
}

func mustResolve(t *testing.T, repo ApprovalRepository, ctx context.Context, tx *sql.Tx, tenantID, area, role string) []string {
	t.Helper()
	ids, err := repo.ResolveEligibleActors(ctx, tx, tenantID, area, role)
	if err != nil {
		t.Fatalf("ResolveEligibleActors: %v", err)
	}
	return ids
}
