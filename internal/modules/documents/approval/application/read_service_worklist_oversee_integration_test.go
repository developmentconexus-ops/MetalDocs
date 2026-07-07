//go:build integration

// F8 (milestone 2b, approval-kernel-backend) — ListWorklist scope=oversee
// integration proof against real Postgres (ADR 0034 testdb harness). Proves
// the authz gate genuinely runs (sqlmock cannot exercise authz.Require's
// system_admin / role_capabilities queries — see
// read_service_test.go's TestListWorklist_Oversee_QueryShapeDropsEligibilityPredicate,
// which only pins the SQL shape):
//
//  1. an actor with NO CapApprovalOversee grant is denied (authz.ErrCapDenied,
//     403) when Oversee=true, even if they hold no other relevant grant.
//  2. an actor with a real qms_admin role grant (-> approval.oversee per
//     db/reference-data/0001_product_reference_data.sql) succeeds and sees an
//     instance NOT in their own eligible_actor_ids pool — proving the
//     eligibility predicate is genuinely dropped (TRUE), not merely bypassed
//     client-side.
package application

import (
	"context"
	"database/sql"
	"testing"

	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// seedActiveStageForOversightFixture seeds an in-progress instance with one
// active stage whose eligible_actor_ids is a DIFFERENT actor than the
// oversight test's caller, so a positive result can only come from the
// Oversee=true eligibility-predicate bypass, never an incidental pool match.
func seedActiveStageForOversightFixture(t *testing.T, database *sql.DB, tenantID string) (instanceID string) {
	t.Helper()
	poolActor := testdb.NewUser(t, database, testdb.WithTenant(tenantID))
	inst := testdb.NewApprovalInstance(t, database, testdb.WithTenant(tenantID), testdb.WithStatus("in_progress"))

	eligible := `["` + poolActor.ID + `"]`
	testdb.SeedWithCaps(t, database, `[{"cap":"document.submit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
			INSERT INTO public.approval_stage_instances
			  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
			   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
			   eligible_actor_ids, status, stage_kind)
			VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', 'QA', 'any_1_of', 'keep_snapshot',
			        $2::jsonb, 'active', 'approval')`,
			inst.ID, eligible,
		)
		return err
	})
	return inst.ID
}

func TestListWorklist_Oversee_NoGrant_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
	// Seed an unrelated in-progress instance so a false-positive "empty result"
	// pass can't masquerade as a denial.
	testdb.NewApprovalInstance(t, db, testdb.WithStatus("in_progress"))

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	_, _, err := svc.ListWorklist(ctxWithIdentity(ten.ID, actor.ID), runner, ten.ID, actor.ID, "", InboxFilter{Oversee: true}, 25, 0)
	if err == nil {
		t.Fatal("ListWorklist(scope=oversee) for actor with no CapApprovalOversee grant = nil error, want authz.ErrCapDenied")
	}
}

func TestListWorklist_Oversee_WithGrant_SeesInstanceOutsideOwnEligiblePool(t *testing.T) {
	db, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, db)

	// The oversight actor is NOT the submitter and NOT in the instance's
	// eligible_actor_ids — the only way ListWorklist can return this row for
	// them is the Oversee=true eligibility-predicate bypass.
	instanceID := seedActiveStageForOversightFixture(t, db, ten.ID)

	overseer := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
	otherArea := testdb.NewTaxonomy(t, db, testdb.WithTenant(ten.ID)).ProcessAreaCode
	grantRoleInArea(t, db, ten.ID, overseer.ID, otherArea, "qms_admin")

	svc := newReadServiceForIntegration(t, db)
	runner := platformdb.NewTxRunner(db)

	items, _, err := svc.ListWorklist(ctxWithIdentity(ten.ID, overseer.ID), runner, ten.ID, overseer.ID, "", InboxFilter{Oversee: true}, 25, 0)
	if err != nil {
		t.Fatalf("ListWorklist(scope=oversee) for CapApprovalOversee holder = %v, want nil", err)
	}
	found := false
	for _, it := range items {
		if it.InstanceID == instanceID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListWorklist(scope=oversee) items = %#v, want to include instance %q (oversight must see instances outside the actor's own eligible pool)", items, instanceID)
	}
}
