//go:build integration
// +build integration

// Package approval_test — G1 (ROADMAP unit 2.1, per-profile signature policy)
// DB last-line validation. Confirms migration 0295 installs the bidirectional
// DEFERRABLE INITIALLY DEFERRED constraint triggers that bind a document
// profile's governance_class to the shape of its ACTIVE approval route:
//
//	controlado -> the active route MUST contain >=1 approval-kind stage
//	simples    -> a review-only route is permitted
//	livre      -> an active route is REQUIRED and MUST carry zero stages
//	              (ADR 0087, migration 0316 — supersedes the pre-0087 rule
//	              "no approval route is permitted at all")
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

// TestGovernancePolicy_Livre_StagedRoute_RejectedAtCommit asserts a livre
// profile rejects an active route that carries ANY stage — approval-kind here.
// The route and its stage are written in one tx, so the deferred trigger fires
// once against the full committed stage set.
func TestGovernancePolicy_Livre_StagedRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"approval"})
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Livre_ReviewOnlyStage_RejectedAtCommit asserts the livre
// arm counts stage ROWS, not approval-kind stages: a review-only stage is just
// as forbidden. "A livre route that reviews is not livre."
func TestGovernancePolicy_Livre_ReviewOnlyStage_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"review"})
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Livre_StageAddedToExistingRoute_RejectedAtCommit isolates
// the STAGE-write direction: the livre zero-stage route is already committed and
// untouched, and a later tx inserts a stage into it. That tx writes only
// approval_route_stages, so the stage trigger is the sole guard here.
func TestGovernancePolicy_Livre_StageAddedToExistingRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil); err != nil {
		t.Fatalf("seed livre zero-stage route: %v", err)
	}
	routeID := activeRouteID(t, db, tenant.ID, tax.ProfileCode)

	assertPolicyViolation(t, addStageToRoute(t, db, routeID, "approval"), "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Livre_IntraTxStageAddThenDelete_Accepted is the mirror of
// the controlado intra-tx downgrade case: a tx that passes through an INVALID
// intermediate (livre route WITH a stage) but commits a valid final state (stage
// deleted) is accepted, because the DEFERRABLE INITIALLY DEFERRED trigger judges
// only the final committed stage set.
func TestGovernancePolicy_Livre_IntraTxStageAddThenDelete_Accepted(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

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
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM public.approval_route_stages WHERE route_id = $1::uuid`,
		routeID,
	); err != nil {
		t.Fatalf("delete approval stage: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("livre route whose final committed shape is stageless must commit: %v", err)
	}
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

// TestGovernancePolicy_Simples_ZeroStageRoute_RejectedAtCommit is the ADR 0087
// §4 exclusivity guard, direction A' (route-row write): a stageless active
// route AUTO-APPROVES on submit, so that shape may exist ONLY under livre.
// simples had no stage-count floor at the DB line before migration 0316 — the
// app-side structural check was the only thing standing between a governed
// profile and silent auto-approval, and it is not authoritative
// (RouteAdminService.resolvePolicy yields "" whenever the policy reader is
// unwired, and Validate("") is structural-only by design).
func TestGovernancePolicy_Simples_ZeroStageRoute_RejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("simples"))

	err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil)
	assertPolicyViolation(t, err, "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_Simples_ZeroStageRoute_ActivationRejectedAtCommit is the
// same guard through the OTHER write direction: the route is seeded inactive
// (unconstrained) and later activated, touching approval_routes only — the
// stage trigger never fires, so only the route-row trigger can catch it.
func TestGovernancePolicy_Simples_ZeroStageRoute_ActivationRejectedAtCommit(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("simples"))

	routeID := seedInactiveRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil)
	assertPolicyViolation(t, activateRoute(t, db, routeID), "ErrRouteViolatesProfilePolicy")
}

// TestGovernancePolicy_ReclassifyLivreToSimples_ZeroStage_Rejected closes the
// third direction: reclassification. A livre profile's legal zero-stage route
// must not become a simples profile's illegal one by editing the class.
func TestGovernancePolicy_ReclassifyLivreToSimples_ZeroStage_Rejected(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))

	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil); err != nil {
		t.Fatalf("seed livre zero-stage route: %v", err)
	}
	assertPolicyViolation(t, reclassify(t, db, tenant.ID, tax.ProfileCode, "simples"), "ErrClassChangeRouteConflict")
}

// TestGovernancePolicy_Livre_ZeroStageRoute_Accepted is the ADR 0087 positive
// case and the exact inversion of the pre-0087 rule: a livre profile's active
// route with zero stages is the CONFIGURED shape, and must commit cleanly. It is
// what makes livre documents creatable at all under universal config-first.
func TestGovernancePolicy_Livre_ZeroStageRoute_Accepted(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil); err != nil {
		t.Fatalf("livre zero-stage active route must commit: %v", err)
	}
}

// TestGovernancePolicy_ReclassifyControladoToLivre_Conflict_Rejected asserts
// direction B under the new livre rule: a controlado profile with an active
// approval-stage route cannot be reclassified to livre, because the surviving
// route would carry stages a livre profile forbids.
func TestGovernancePolicy_ReclassifyControladoToLivre_Conflict_Rejected(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("controlado"))

	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, []string{"approval"}); err != nil {
		t.Fatalf("seed controlado approval route: %v", err)
	}

	err := reclassify(t, db, tenant.ID, tax.ProfileCode, "livre")
	assertPolicyViolation(t, err, "ErrClassChangeRouteConflict")
}

// TestGovernancePolicy_ReclassifyLivreToControlado_Conflict_Rejected asserts the
// other direction: a livre profile's zero-stage active route is invalid under
// controlado (no approval-kind stage), so the reclassification is rejected.
func TestGovernancePolicy_ReclassifyLivreToControlado_Conflict_Rejected(t *testing.T) {
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))
	setGovernanceClass(t, db, tenant.ID, tax.ProfileCode, "livre")

	if err := commitRouteWithStages(t, db, tenant.ID, tax.ProfileCode, owner.ID, nil); err != nil {
		t.Fatalf("seed livre zero-stage route: %v", err)
	}

	err := reclassify(t, db, tenant.ID, tax.ProfileCode, "controlado")
	assertPolicyViolation(t, err, "ErrClassChangeRouteConflict")
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

// activeRouteID returns the id of the profile's single active route. Fixture
// lookup, not an assertion — it fatals if the route is missing.
func activeRouteID(t *testing.T, db *sql.DB, tenantID, profileCode string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(context.Background(),
		`SELECT id::text FROM public.approval_routes
		  WHERE tenant_id = $1::uuid AND profile_code = $2 AND active`,
		tenantID, profileCode,
	).Scan(&id); err != nil {
		t.Fatalf("lookup active route: %v", err)
	}
	return id
}

// addStageToRoute inserts a single stage into an already-committed route in its
// own tx and returns the COMMIT error. Touches ONLY approval_route_stages, so
// the stage trigger is the sole guard.
func addStageToRoute(t *testing.T, db *sql.DB, routeID, kind string) error {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin add-stage tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO public.approval_route_stages
		  (route_id, stage_order, name, required_capability,
		   quorum, on_eligibility_drift, stage_kind)
		VALUES ($1::uuid, 1, 'Stage', 'document.review',
		        'any_1_of', 'reduce_quorum', $2)`,
		routeID, kind,
	); err != nil {
		t.Fatalf("insert stage (%s): %v", kind, err)
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
