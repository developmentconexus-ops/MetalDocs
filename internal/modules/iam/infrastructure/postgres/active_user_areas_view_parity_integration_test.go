//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"metaldocs/tests/integration/testdb"
)

// F3.1 view-vs-base parity gate (D6: the producer proves itself before any
// consumer repoints). metaldocs.v_active_user_areas MUST return EXACTLY the
// active-now projection of public.user_process_areas — the Model-A predicate
// effective_to IS NULL (ADR 0037 D1; ADR-0039 D3a/D4). Columns: the four the
// three M3 consumers consume — (tenant_id, user_id, area_code, role). role is
// required by approval ResolveEligibleActors (C3).
//
// The seeded set exercises the authz-drift guards: an active row, a SECOND
// active row (different role, same user+area) for multi-role, a REVOKED row
// (past effective_to — must be EXCLUDED), and a WRONG-TENANT active row (must
// not leak). Parity = set-equality, not a count.

type upaRow struct {
	tenantID string
	userID   string
	areaCode string
	role     string
}

func seedActiveAndRevoked(t *testing.T, db *sql.DB) (tenantID, userID, area1, area2 string) {
	t.Helper()
	ctx := context.Background()

	tn := testdb.NewTenant(t, db)
	tenantID = tn.ID
	user := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	userID = user.ID
	// Two governed areas — the partial unique index ux_user_process_areas_one_active
	// allows at most ONE active row per (user, tenant, area), so multi-role coverage
	// uses two areas (area1 qms_admin, area2 approver), not two roles on one area.
	tax1 := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	tax2 := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	area1 = tax1.ProcessAreaCode
	area2 = tax2.ProcessAreaCode

	now := time.Now().UTC()

	// user_process_areas carries the membership.manage tripwire — assert it tx-locally.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		// (a) active row, area1, role qms_admin
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'qms_admin', $4)`,
			userID, tenantID, area1, now.Add(-2*time.Hour),
		); err != nil {
			return err
		}
		// (b) active row, area2, role approver -> multi-area / multi-role coverage
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'approver', $4)`,
			userID, tenantID, area2, now.Add(-2*time.Hour),
		); err != nil {
			return err
		}
		// (c) REVOKED row on area1 (past effective_to > effective_from, revoked_by set) ->
		//     MUST be excluded from the active view. effective_to NOT NULL, so it is not
		//     covered by the active partial unique index and coexists with (a).
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
			 VALUES ($1, $2::uuid, $3, 'editor', $4, $5, $1, $1)`,
			userID, tenantID, area1, now.Add(-3*time.Hour), now.Add(-1*time.Hour),
		); err != nil {
			return err
		}
		return nil
	})
	return tenantID, userID, area1, area2
}

func queryRows(t *testing.T, db *sql.DB, query string, args ...any) []upaRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("query: %v\nsql: %s", err, query)
	}
	defer rows.Close()
	var out []upaRow
	for rows.Next() {
		var r upaRow
		if err := rows.Scan(&r.tenantID, &r.userID, &r.areaCode, &r.role); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return out
}

func TestActiveUserAreasView_ParityWithBaseActiveNow(t *testing.T) {
	db, _ := testdb.Open(t)
	tenantID, _, area1, area2 := seedActiveAndRevoked(t, db)

	// Wrong-tenant active row — must NOT appear when we scope to tenantID.
	otherTenant := testdb.NewTenant(t, db)
	otherUser := testdb.NewUser(t, db, testdb.WithTenant(otherTenant.ID))
	otherTax := testdb.NewTaxonomy(t, db, testdb.WithTenant(otherTenant.ID))
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'qms_admin', now())`,
			otherUser.ID, otherTenant.ID, otherTax.ProcessAreaCode,
		)
		return err
	})

	const viewQ = `SELECT tenant_id, user_id, area_code, role
	                 FROM metaldocs.v_active_user_areas
	                WHERE tenant_id = $1::uuid
	                ORDER BY area_code, role, user_id`
	const baseQ = `SELECT tenant_id, user_id, area_code, role
	                 FROM public.user_process_areas
	                WHERE tenant_id = $1::uuid AND effective_to IS NULL
	                ORDER BY area_code, role, user_id`

	got := queryRows(t, db, viewQ, tenantID)
	want := queryRows(t, db, baseQ, tenantID)

	if len(got) != len(want) {
		t.Fatalf("row count: view=%d base-active=%d\nview=%+v\nbase=%+v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("row %d: view=%+v base=%+v", i, got[i], want[i])
		}
	}

	// Explicit authz-drift assertions on the scoped set.
	byArea := map[string]string{}
	for _, r := range got {
		byArea[r.areaCode] = r.role
	}
	if byArea[area1] != "qms_admin" {
		t.Fatalf("expected active qms_admin on area1 %q, got %+v", area1, byArea)
	}
	if byArea[area2] != "approver" {
		t.Fatalf("expected active approver on area2 %q, got %+v", area2, byArea)
	}
	for _, r := range got {
		if r.role == "editor" {
			t.Fatalf("revoked role 'editor' must be EXCLUDED from the active view, got %+v", got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 active rows for tenant (area1/qms_admin + area2/approver), got %d: %+v", len(got), got)
	}
}
