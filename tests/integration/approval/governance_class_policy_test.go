//go:build integration
// +build integration

// Package approval_test — G1 (ROADMAP unit 2.1, per-profile signature policy)
// DB last-line validation. Confirms migration 0295 installs the bidirectional
// DEFERRABLE INITIALLY DEFERRED constraint triggers that bind a document
// profile's governance_class to the shape of its ACTIVE approval route:
//
//	controlado -> the active route MUST contain >=1 approval-kind stage
//	simples    -> a review-only route is permitted
//	livre      -> no approval route is permitted at all
//
// Direction A (approval_route_stages write) is validated at COMMIT against the
// full committed stage set; direction B (document_profiles.governance_class
// reclassification) re-validates every active route against the new class.
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/tests/integration/testdb"
)

// TestGovernancePolicy_Controlado_ReviewOnlyRoute_RejectedAtCommit asserts a
// controlado profile whose active route carries only a review-kind stage is
// rejected at COMMIT with the ErrRouteViolatesProfilePolicy: P0001 token.
func TestGovernancePolicy_Controlado_ReviewOnlyRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"review"})
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Controlado_ApprovalStage_Accepted asserts a controlado
// profile whose active route contains an approval-kind stage commits cleanly.
func TestGovernancePolicy_Controlado_ApprovalStage_Accepted(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"review", "approval"})
	if err != nil {
		t.Fatalf("controlado route with an approval stage must commit: %v", err)
	}
}

// TestGovernancePolicy_Livre_AnyRoute_RejectedAtCommit asserts a livre profile
// rejects the very existence of an active route, even one with an approval
// stage.
func TestGovernancePolicy_Livre_AnyRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"approval"})
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Simples_ReviewOnlyRoute_Accepted asserts a simples
// profile accepts a review-only route (no approval stage required).
func TestGovernancePolicy_Simples_ReviewOnlyRoute_Accepted(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "simples")

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"review"})
	if err != nil {
		t.Fatalf("simples review-only route must commit: %v", err)
	}
}

// TestGovernancePolicy_ReclassifySimplesToControlado_Conflict_Rejected asserts
// direction B: a simples profile with an active review-only route, reclassified
// to controlado, is rejected at COMMIT with the ErrClassChangeRouteConflict:
// token (the app maps this to a friendly 409).
func TestGovernancePolicy_ReclassifySimplesToControlado_Conflict_Rejected(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "simples")

	// A review-only route is valid under simples.
	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"review"}); err != nil {
		t.Fatalf("seed simples review-only route: %v", err)
	}

	// Reclassifying to controlado now conflicts with that active route.
	err := reclassify(t, db, tenant.ID, tax.ProfileCode, "controlado")
	assertPolicyViolation(t, err, "ErrClassChangeRouteConflict")
}

// TestGovernancePolicy_ReclassifyControladoToSimples_Compatible_Accepted
// asserts direction B allows a widening reclassification (controlado ->
// simples) even while an approval-stage route is active.
func TestGovernancePolicy_ReclassifyControladoToSimples_Compatible_Accepted(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"approval"}); err != nil {
		t.Fatalf("seed controlado approval route: %v", err)
	}

	if err := reclassify(t, db, tenant.ID, tax.ProfileCode, "simples"); err != nil {
		t.Fatalf("controlado->simples with an approval route must commit: %v", err)
	}
}

// TestGovernancePolicy_DeferredTrigger_IntraTxDowngrade_RejectedAtCommit proves
// the DEFERRABLE INITIALLY DEFERRED trigger closes the ordering/race hole that
// makes app-side submit-time enforcement unnecessary (G1 option-1 ratification):
// a transaction that passes through a VALID intermediate state (controlado route
// WITH an approval stage) but COMMITs an invalid final state (approval stage
// removed, leaving review-only) is still rejected — the trigger evaluates the
// final committed stage set at COMMIT, not the intermediate. No sequence of
// row writes can leave an active controlado route without a signature stage.
func TestGovernancePolicy_DeferredTrigger_IntraTxDowngrade_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	routeID := uuid.NewString()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.approval_routes (id, tenant_id, name, profile_code, active, created_by)
		 VALUES ($1::uuid, $2::uuid, 'Route', $3, true, $4)`,
		routeID, tenant.ID, tax.ProfileCode, owner.ID,
	); err != nil {
		t.Fatalf("insert route: %v", err)
	}
	// Intermediate VALID state: controlado route with an approval-kind stage.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO public.approval_route_stages
		  (route_id, stage_order, name, required_capability,
		   quorum, on_eligibility_drift, stage_kind)
		VALUES ($1::uuid, 1, 'Signoff', 'document.signoff',
		        'any_1_of', 'reduce_quorum', 'approval')`,
		routeID,
	); err != nil {
		t.Fatalf("insert approval stage: %v", err)
	}
	// Downgrade before COMMIT: remove the approval stage → final state is
	// review-only (here: no stages at all), which controlado forbids.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM public.approval_route_stages WHERE route_id = $1::uuid`,
		routeID,
	); err != nil {
		t.Fatalf("delete approval stage: %v", err)
	}

	assertPolicyViolation(t, tx.Commit(), "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Controlado_ZeroStageRoute_RejectedAtCommit proves the
// approval_routes row trigger closes the zero-stage hole: a controlado active
// route with NO stage rows writes nothing to approval_route_stages, so the
// stage trigger never fires — only the route-row trigger catches it.
func TestGovernancePolicy_Controlado_ZeroStageRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil)
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Livre_ZeroStageRoute_RejectedAtCommit proves a livre
// active route with no stages (the cheapest way to smuggle a forbidden route)
// is still rejected — only the route-row trigger sees it.
func TestGovernancePolicy_Livre_ZeroStageRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil)
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_ActivateReviewOnlyRoute_RejectedAtCommit proves the
// activate-later hole is closed: a controlado route seeded INACTIVE with only a
// review stage (valid while inactive) is rejected when a later, stages-untouched
// UPDATE flips active=true. That UPDATE writes only approval_routes, so the
// stage trigger cannot fire — the route-row trigger is the sole guard.
func TestGovernancePolicy_ActivateReviewOnlyRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	// Seed an INACTIVE review-only route (valid: only active routes are enforced).
	routeID := seedInactiveRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"review"})

	// Flip active=true in a standalone UPDATE (touches only approval_routes).
	err := activateRoute(t, db, routeID)
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// commitRouteWithStages inserts one active approval route plus a stage per
// stageKinds entry inside a SINGLE transaction and returns the COMMIT error, so
// the deferred trigger fires once against the full committed stage set (mirrors
// the real route-admin multi-statement write). approval_routes /
// approval_route_stages carry no tripwire, so no caps assertion is needed.
func commitRouteWithStages(t *testing.T, db *sql.DB, tenantID, profileCode, owner string, stageKinds []string) error {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin route tx: %v", err)
	}
	defer tx.Rollback()

	routeID := uuid.NewString()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.approval_routes (id, tenant_id, name, profile_code, active, created_by)
		 VALUES ($1::uuid, $2::uuid, 'Route', $3, true, $4)`,
		routeID, tenantID, profileCode, owner,
	); err != nil {
		t.Fatalf("insert approval_routes: %v", err)
	}

	for i, kind := range stageKinds {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.approval_route_stages
			  (route_id, stage_order, name, required_capability,
			   quorum, on_eligibility_drift, stage_kind)
			VALUES ($1::uuid, $2, $3, 'document.review',
			        'any_1_of', 'reduce_quorum', $4)`,
			routeID, i+1, "Stage "+kind, kind,
		); err != nil {
			t.Fatalf("insert approval_route_stages (%s): %v", kind, err)
		}
	}

	return tx.Commit()
}

// seedInactiveRouteWithStages inserts one INACTIVE route plus its stages in a
// single committed tx and returns the route id. Because the route is inactive,
// neither trigger raises (only active routes are enforced), so this commits
// cleanly regardless of the stage shape — it is fixture setup, not an assertion.
func seedInactiveRouteWithStages(t *testing.T, db *sql.DB, tenantID, profileCode, owner string, stageKinds []string) string {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin seed tx: %v", err)
	}
	defer tx.Rollback()

	routeID := uuid.NewString()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.approval_routes (id, tenant_id, name, profile_code, active, created_by)
		 VALUES ($1::uuid, $2::uuid, 'Route', $3, false, $4)`,
		routeID, tenantID, profileCode, owner,
	); err != nil {
		t.Fatalf("insert inactive route: %v", err)
	}
	for i, kind := range stageKinds {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.approval_route_stages
			  (route_id, stage_order, name, required_capability,
			   quorum, on_eligibility_drift, stage_kind)
			VALUES ($1::uuid, $2, $3, 'document.review',
			        'any_1_of', 'reduce_quorum', $4)`,
			routeID, i+1, "Stage "+kind, kind,
		); err != nil {
			t.Fatalf("insert stage (%s): %v", kind, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit inactive route seed: %v", err)
	}
	return routeID
}

// activateRoute flips approval_routes.active true in its own tx and returns the
// COMMIT error, so the route-row trigger is what surfaces. Touches ONLY
// approval_routes (no stage write), so the stage trigger cannot fire.
func activateRoute(t *testing.T, db *sql.DB, routeID string) error {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin activate tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE public.approval_routes SET active = true WHERE id = $1::uuid`,
		routeID,
	); err != nil {
		t.Fatalf("update active: %v", err)
	}
	return tx.Commit()
}

// setGovernanceClass reclassifies a freshly-seeded profile that has NO active
// routes yet, so the direction-B guard trivially passes. It fatals on error
// (fixture setup, not an assertion). document_profiles carries the taxonomy
// tripwire, so the UPDATE runs with taxonomy.manage asserted tx-locally.
func setGovernanceClass(t *testing.T, db *sql.DB, tenantID, code, class string) {
	t.Helper()
	if err := reclassify(t, db, tenantID, code, class); err != nil {
		t.Fatalf("setGovernanceClass(%s): %v", class, err)
	}
}

// reclassify UPDATEs document_profiles.governance_class inside its own tx (with
// taxonomy.manage asserted tx-locally) and returns the COMMIT error, so the
// deferred reclassification guard is what surfaces.
func reclassify(t *testing.T, db *sql.DB, tenantID, code, class string) error {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin reclassify tx: %v", err)
	}
	defer tx.Rollback()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"taxonomy.manage"}]`)
	if _, err := tx.ExecContext(ctx,
		`UPDATE metaldocs.document_profiles SET governance_class = $1
		  WHERE tenant_id = $2::uuid AND code = $3`,
		class, tenantID, code,
	); err != nil {
		t.Fatalf("update governance_class: %v", err)
	}

	return tx.Commit()
}

// assertPolicyViolation fails unless err is a P0001 exception whose message
// starts with the given token prefix.
func assertPolicyViolation(t *testing.T, err error, token string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s P0001 exception, got nil error", token)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code != "P0001" {
			t.Fatalf("expected SQLSTATE P0001, got %s: %v", pgErr.Code, err)
		}
		if !strings.HasPrefix(pgErr.Message, token) {
			t.Fatalf("expected message prefix %q, got %q", token, pgErr.Message)
		}
		return
	}
	if !strings.Contains(err.Error(), token) {
		t.Fatalf("expected error mentioning %s, got: %v", token, err)
	}
}
