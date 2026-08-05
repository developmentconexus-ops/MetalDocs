//go:build integration
// +build integration

package application

// Task 9 (approval accountability loop) — real-DB integration proof for
// SLAExtensionService.Extend (sla_extension_service.go).
//
// Follows mark_reviewed_service_integration_test.go's shape (service-layer
// call, real testdb.Open, real authz.Require via a genuine user_process_areas
// grant) and job_integration_test.go's raw-SQL stage-instance fixture
// (approval_sla_surfacer/job_integration_test.go's seedOverdueStageFixture* —
// no factory builder exists for a stage instance with a specific due_at, and
// that helper is unexported in its own package, so this is a second,
// deliberately local copy per the task brief's resolved ambiguity: no shared
// test helper package is introduced).
//
// Proves:
//  1. TestExtendSLA_MovesDueAtForward_RecordsGovernanceEvent — a forward
//     extension updates due_at and writes exactly one governance_events row
//     (event_type = approval_sla_extended) carrying previous/new due_at and
//     reason.
//  2. TestExtendSLA_BackwardsDueAt_Rejected — a due_at at or before the
//     current due_at is rejected (ErrSLAExtensionNotForward), no write.
//  3. TestExtendSLA_BlankReason_Rejected — a whitespace-only reason is
//     rejected (ErrSLAExtensionReasonRequired) before the transaction opens.
//  4. TestExtendSLA_ActorWithoutCapability_Forbidden — an actor without
//     approval.sla_extend is denied by the real tier-2 authz.Require gate
//     (authz.ErrCapDenied), not a DB tripwire leak.
//  5. TestExtendSLA_ReArmsOverdueEligibility — the interaction with the SLA
//     surfacer's marker (Task 7): extending an already-overdue stage forward
//     (while still landing in the past) flips the surfacer's real
//     ListOverdueStages predicate from "not eligible" (already surfaced for
//     the old due_at) to "eligible again" (stale sla_surfaced_at is now
//     before the new due_at) — proving the re-arm is structural, not
//     something Extend has to implement by hand.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	approvalinfra "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// extendSLACtxWithIdentity mirrors markReviewedCtxWithIdentity: seeds the
// tenant/actor identity TxRunner.Do's chokepoint auto-seed expects.
func extendSLACtxWithIdentity(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, actorID)
	return ctx
}

// grantSLAExtendCapability seeds a user_process_areas membership under role
// qms_admin, which curated role_capabilities maps to approval.sla_extend
// (db/reference-data/0001_product_reference_data.sql:302). CapApprovalSLAExtend
// is ScopeArea (capability_scope.go:51), so areaCode must match the stage's
// area_code_snapshot for the grant to apply.
func grantSLAExtendCapability(t *testing.T, dbc *sql.DB, tenantID, actorID, areaCode string) {
	t.Helper()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, 'qms_admin', now() - interval '1 hour')`,
			actorID, tenantID, areaCode,
		)
		return err
	})
}

// seedActiveStageWithDueAt seeds tenant/user/document/route/instance plus one
// ACTIVE stage instance with the given due_at and sla_surfaced_at (either may
// be nil), stamped with areaCode as its area_code_snapshot. Mirrors
// approval_sla_surfacer/job_integration_test.go's seedOverdueStageFixtureForActor
// raw-SQL FK-chain seeding; that helper is package-private to a different
// package, so this is a local, non-shared copy (per the task brief's
// resolved ambiguity — no shared harness scaffolding is added to
// tests/integration/testdb). areaCode must be a real process area on
// tenantID's taxonomy (user_process_areas_tenant_id_area_code_fkey) — callers
// pass testdb.NewTaxonomy(...).ProcessAreaCode, matching the value they grant
// the capability against via grantSLAExtendCapability.
func seedActiveStageWithDueAt(t *testing.T, dbc *sql.DB, tenantID, areaCode string, dueAt, slaSurfacedAt *time.Time, eligibleActorID string) (stageID, instanceID string) {
	t.Helper()
	ctx := context.Background()

	author := testdb.NewUser(t, dbc, testdb.WithTenant(tenantID))
	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(tenantID))
	instance := testdb.NewApprovalInstance(t, dbc,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
	)

	eligible := `["` + eligibleActorID + `"]`
	var id string
	if err := dbc.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
		   eligible_actor_ids, status, stage_kind, due_at, sla_surfaced_at,
		   required_capability_snapshot)
		VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', $2, 'any_1_of', 'keep_snapshot',
		        $3::jsonb, 'active', 'review', $4, $5, 'document.review')
		RETURNING id::text`,
		instance.ID, areaCode, eligible,
		nullableTimeArg(dueAt), nullableTimeArg(slaSurfacedAt),
	).Scan(&id); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}
	return id, instance.ID
}

func nullableTimeArg(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func readStageDueAt(t *testing.T, dbc *sql.DB, stageID string) *time.Time {
	t.Helper()
	var due sql.NullTime
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT due_at FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&due); err != nil {
		t.Fatalf("readStageDueAt: %v", err)
	}
	if !due.Valid {
		return nil
	}
	return &due.Time
}

func newSLAExtensionTestService() *SLAExtensionService {
	return NewSLAExtensionService(NewSQLEmitter(), RealClock{})
}

// TestExtendSLA_MovesDueAtForward_RecordsGovernanceEvent proves the happy
// path: due_at is postponed and exactly one governance_events row records the
// exception (previous_due_at, new_due_at, reason).
func TestExtendSLA_MovesDueAtForward_RecordsGovernanceEvent(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantSLAExtendCapability(t, dbc, tnt.ID, actor.ID, tax.ProcessAreaCode)

	now := time.Now().UTC().Truncate(time.Second)
	oldDue := now.Add(2 * 24 * time.Hour)
	newDue := now.Add(5 * 24 * time.Hour)

	_, instanceID := seedActiveStageWithDueAt(t, dbc, tnt.ID, tax.ProcessAreaCode, &oldDue, nil, actor.ID)

	svc := newSLAExtensionTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := extendSLACtxWithIdentity(tnt.ID, actor.ID)

	err := svc.Extend(callCtx, runner, ExtendRequest{
		TenantID:   tnt.ID,
		InstanceID: instanceID,
		ActorID:    actor.ID,
		NewDueAt:   newDue,
		Reason:     "vendor shipment delayed one week",
	})
	if err != nil {
		t.Fatalf("Extend: unexpected error: %v", err)
	}

	rows, err := dbc.QueryContext(context.Background(),
		`SELECT payload_json, reason FROM public.governance_events
		  WHERE tenant_id = $1::uuid AND resource_id = $2 AND event_type = 'approval_sla_extended'`,
		tnt.ID, instanceID,
	)
	if err != nil {
		t.Fatalf("query governance_events: %v", err)
	}
	defer rows.Close()

	var payloads []map[string]any
	var reasons []string
	for rows.Next() {
		var raw []byte
		var reason string
		if err := rows.Scan(&raw, &reason); err != nil {
			t.Fatalf("scan governance_events: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("decode payload_json: %v", err)
		}
		payloads = append(payloads, payload)
		reasons = append(reasons, reason)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("governance_events approval_sla_extended rows = %d, want exactly 1", len(payloads))
	}
	if reasons[0] != "vendor shipment delayed one week" {
		t.Fatalf("reason = %q, want %q", reasons[0], "vendor shipment delayed one week")
	}
	if payloads[0]["new_due_at"] == nil {
		t.Fatal("payload.new_due_at missing")
	}
	if payloads[0]["previous_due_at"] == nil {
		t.Fatal("payload.previous_due_at missing")
	}
}

// TestExtendSLA_BackwardsDueAt_Rejected proves a non-forward due_at is
// rejected and leaves due_at unchanged — no lost update, no silent no-op.
func TestExtendSLA_BackwardsDueAt_Rejected(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantSLAExtendCapability(t, dbc, tnt.ID, actor.ID, tax.ProcessAreaCode)

	now := time.Now().UTC().Truncate(time.Second)
	oldDue := now.Add(5 * 24 * time.Hour)
	backwards := now.Add(2 * 24 * time.Hour)

	stageID, instanceID := seedActiveStageWithDueAt(t, dbc, tnt.ID, tax.ProcessAreaCode, &oldDue, nil, actor.ID)

	svc := newSLAExtensionTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := extendSLACtxWithIdentity(tnt.ID, actor.ID)

	err := svc.Extend(callCtx, runner, ExtendRequest{
		TenantID:   tnt.ID,
		InstanceID: instanceID,
		ActorID:    actor.ID,
		NewDueAt:   backwards,
		Reason:     "should be rejected",
	})
	if !errors.Is(err, ErrSLAExtensionNotForward) {
		t.Fatalf("Extend with backwards due_at = %v, want ErrSLAExtensionNotForward", err)
	}

	got := readStageDueAt(t, dbc, stageID)
	if got == nil || !got.Equal(oldDue) {
		t.Fatalf("due_at after rejected backwards extension = %v, want unchanged %v", got, oldDue)
	}
}

// TestExtendSLA_BlankReason_Rejected proves a whitespace-only reason fails
// closed before the transaction opens (a governance exception with no reason
// is not a governance record).
func TestExtendSLA_BlankReason_Rejected(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantSLAExtendCapability(t, dbc, tnt.ID, actor.ID, tax.ProcessAreaCode)

	now := time.Now().UTC().Truncate(time.Second)
	oldDue := now.Add(2 * 24 * time.Hour)
	newDue := now.Add(5 * 24 * time.Hour)

	stageID, instanceID := seedActiveStageWithDueAt(t, dbc, tnt.ID, tax.ProcessAreaCode, &oldDue, nil, actor.ID)

	svc := newSLAExtensionTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := extendSLACtxWithIdentity(tnt.ID, actor.ID)

	err := svc.Extend(callCtx, runner, ExtendRequest{
		TenantID:   tnt.ID,
		InstanceID: instanceID,
		ActorID:    actor.ID,
		NewDueAt:   newDue,
		Reason:     "   ",
	})
	if !errors.Is(err, ErrSLAExtensionReasonRequired) {
		t.Fatalf("Extend with blank reason = %v, want ErrSLAExtensionReasonRequired", err)
	}

	got := readStageDueAt(t, dbc, stageID)
	if got == nil || !got.Equal(oldDue) {
		t.Fatalf("due_at after rejected blank-reason extension = %v, want unchanged %v", got, oldDue)
	}
}

// TestExtendSLA_ActorWithoutCapability_Forbidden proves the real tier-2
// authz.Require gate: the SAME request denied without the approval.sla_extend
// grant, then succeeds once granted — mirrors
// TestMarkReviewed_RequiresDocumentReviewCapability's (a) withheld / (b)
// granted shape.
func TestExtendSLA_ActorWithoutCapability_Forbidden(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	// Actor with no role/grant at all (not system_admin, no user_process_areas row).
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))

	now := time.Now().UTC().Truncate(time.Second)
	oldDue := now.Add(2 * 24 * time.Hour)
	newDue := now.Add(5 * 24 * time.Hour)

	stageID, instanceID := seedActiveStageWithDueAt(t, dbc, tnt.ID, tax.ProcessAreaCode, &oldDue, nil, actor.ID)

	svc := newSLAExtensionTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := extendSLACtxWithIdentity(tnt.ID, actor.ID)

	req := ExtendRequest{
		TenantID:   tnt.ID,
		InstanceID: instanceID,
		ActorID:    actor.ID,
		NewDueAt:   newDue,
		Reason:     "should be denied first",
	}

	// (a) withheld -> denied.
	err := svc.Extend(callCtx, runner, req)
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("Extend without approval.sla_extend grant = %v, want authz.ErrCapDenied", err)
	}
	if got := readStageDueAt(t, dbc, stageID); got == nil || !got.Equal(oldDue) {
		t.Fatalf("due_at after denied extension = %v, want unchanged %v", got, oldDue)
	}

	// (b) granted -> succeeds.
	grantSLAExtendCapability(t, dbc, tnt.ID, actor.ID, tax.ProcessAreaCode)
	if err := svc.Extend(callCtx, runner, req); err != nil {
		t.Fatalf("Extend with approval.sla_extend grant = %v, want success", err)
	}
	if got := readStageDueAt(t, dbc, stageID); got == nil || !got.Equal(newDue) {
		t.Fatalf("due_at after granted extension = %v, want %v", got, newDue)
	}
}

// TestExtendSLA_ReArmsOverdueEligibility proves the Task 7 interaction called
// out in the brief: an already-overdue stage that was already surfaced (so
// the surfacer's real ListOverdueStages predicate currently excludes it) is
// re-armed by a forward extension that still lands in the past — because the
// stale sla_surfaced_at becomes less than the new, later due_at. Extend does
// not touch sla_surfaced_at at all; the re-arm is a structural consequence of
// the shared predicate, proved here against the REAL reader
// (infrastructure.SLAOverdueReaderPG), not a re-implementation of its SQL.
func TestExtendSLA_ReArmsOverdueEligibility(t *testing.T) {
	dbc, _ := testdb.Open(t)
	dbc.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, dbc)
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantSLAExtendCapability(t, dbc, tnt.ID, actor.ID, tax.ProcessAreaCode)

	now := time.Now().UTC().Truncate(time.Second)
	oldDue := now.Add(-2 * time.Hour)     // overdue
	surfacedAt := now.Add(-1 * time.Hour) // already surfaced for oldDue (surfacedAt > oldDue)
	newDue := now.Add(-30 * time.Minute)  // forward of oldDue, but STILL overdue

	stageID, instanceID := seedActiveStageWithDueAt(t, dbc, tnt.ID, tax.ProcessAreaCode, &oldDue, &surfacedAt, actor.ID)

	reader := approvalinfra.NewSLAOverdueReaderPG(dbc)
	isEligible := func() bool {
		t.Helper()
		ctx := authz.WithBackgroundBypass(context.Background())
		tx, err := dbc.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()
		if err := authz.BypassSystem(ctx, tx); err != nil {
			t.Fatalf("BypassSystem: %v", err)
		}
		if err := authz.SeedTxTenant(ctx, tx, tnt.ID); err != nil {
			t.Fatalf("SeedTxTenant: %v", err)
		}
		views, err := reader.ListOverdueStages(ctx, tx, tnt.ID, time.Now().UTC(), 25)
		if err != nil {
			t.Fatalf("ListOverdueStages: %v", err)
		}
		for _, v := range views {
			if v.StageInstanceID == stageID {
				return true
			}
		}
		return false
	}

	if isEligible() {
		t.Fatalf("stage %s eligible before extension, want NOT eligible (already surfaced for the current due_at)", stageID)
	}

	svc := newSLAExtensionTestService()
	runner := db.NewTxRunner(dbc)
	callCtx := extendSLACtxWithIdentity(tnt.ID, actor.ID)

	if err := svc.Extend(callCtx, runner, ExtendRequest{
		TenantID:   tnt.ID,
		InstanceID: instanceID,
		ActorID:    actor.ID,
		NewDueAt:   newDue,
		Reason:     "extend the overdue deadline by 90 minutes",
	}); err != nil {
		t.Fatalf("Extend: unexpected error: %v", err)
	}

	if !isEligible() {
		t.Fatalf("stage %s NOT eligible after extension, want eligible (stale sla_surfaced_at is now before the new due_at)", stageID)
	}

	// sla_surfaced_at itself must be untouched by Extend — the marker is the
	// surfacer's, not the extension's, to write.
	got := readSLASurfacedAtLocal(t, dbc, stageID)
	if got == nil || !got.Equal(surfacedAt) {
		t.Fatalf("sla_surfaced_at after Extend = %v, want unchanged %v (Extend must never write this column)", got, surfacedAt)
	}
}

func readSLASurfacedAtLocal(t *testing.T, dbc *sql.DB, stageID string) *time.Time {
	t.Helper()
	var ts sql.NullTime
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT sla_surfaced_at FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&ts); err != nil {
		t.Fatalf("readSLASurfacedAtLocal: %v", err)
	}
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}
