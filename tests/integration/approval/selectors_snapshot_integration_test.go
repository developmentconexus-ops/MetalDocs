//go:build integration
// +build integration

// Package approval_test — unit 3.2 slice 6a (M4 ActorSelector, Option C′
// drift-source cutover, ADDITIVE step). Real-Postgres coverage for the two DB
// behaviors slice 6a introduces on approval_stage_instances.selectors_snapshot
// (docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md §8b):
//
//  1. Round-trip: a mixed selectors snapshot (named_user + role_in_fixed_area +
//     role_in_document_area) written by InsertStageInstances survives a
//     LoadInstance read byte-for-byte as a slice of domain.ActorSelector —
//     proving the marshal (json tags: kind/user_id/role/area_code), the new
//     bound INSERT column, and BOTH scan sites' unmarshal agree. submit_choice
//     is intentionally absent: the submit path materializes it into named_user
//     BEFORE it ever reaches the instance, so a persisted snapshot never
//     carries a submit_choice selector.
//  2. Backfill parity: migration 0304's verbatim backfill UPDATE, run against a
//     bare in-flight row seeded at selectors_snapshot='[]' with flat scalars,
//     synthesizes exactly one role_in_fixed_area(required_role_snapshot,
//     area_code_snapshot) selector — the mapping that keeps the rewired drift
//     path resolving the same set the pre-cutover flat path did. This is the
//     deploy-time SQL, exercised directly (mirrors slice 2's reviewer pattern
//     of running the migration's own backfill statement verbatim).
//
// Targets the repository layer directly (InsertStageInstances / LoadInstance),
// the same seams submit_service.go calls in production, without the full
// SubmitService stack — mirrors sla_due_at_integration_test.go.
package approval_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

func selectorsSnapshotSeedInstance(t *testing.T, database *sql.DB) (tenantID, instanceID, authorID string) {
	t.Helper()
	tenant := testdb.NewTenant(t, database)
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"))
	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)
	route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))
	instance := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
	)
	return tenant.ID, instance.ID, author.ID
}

// TestInsertStageInstances_SelectorsSnapshotRoundTrip proves a mixed
// selectors snapshot written by InsertStageInstances reloads intact through
// LoadInstance — the persist (marshal + bound column) and both scan sites'
// unmarshal round-trip the discriminated union without loss or key drift.
func TestInsertStageInstances_SelectorsSnapshotRoundTrip(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	tenantID, instanceID, authorID := selectorsSnapshotSeedInstance(t, database)

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	runner := db.NewTxRunner(database)

	want := []domain.ActorSelector{
		{Kind: domain.SelectorNamedUser, UserID: authorID},
		{Kind: domain.SelectorRoleInFixedArea, Role: "qms_admin", AreaCode: "quality"},
		{Kind: domain.SelectorRoleInDocumentArea, Role: "approver"},
	}
	stageID := uuid.NewString()

	err := runner.Do(ctx, func(tx *sql.Tx) error {
		return repo.InsertStageInstances(ctx, tx, []domain.StageInstance{
			{
				ID:                         stageID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 1,
				NameSnapshot:               "Stage 1",
				RequiredRoleSnapshot:       "qms_admin",
				RequiredCapabilitySnapshot: "document.approve",
				AreaCodeSnapshot:           "quality",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftReduceQuorum,
				Kind:                       domain.StageKindApproval,
				EligibleActorIDs:           []string{authorID},
				Status:                     domain.StageActive,
				SelectorsSnapshot:          want,
			},
		})
	})
	if err != nil {
		t.Fatalf("InsertStageInstances: %v", err)
	}

	var loaded *domain.Instance
	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		var lerr error
		loaded, lerr = repo.LoadInstance(ctx, tx, tenantID, instanceID)
		return lerr
	}); err != nil {
		t.Fatalf("LoadInstance: %v", err)
	}
	if loaded == nil || len(loaded.Stages) != 1 {
		t.Fatalf("LoadInstance: got %d stages, want 1", len(loaded.Stages))
	}

	got := loaded.Stages[0].SelectorsSnapshot
	assertSelectorsEqual(t, got, want)
}

// TestMigration0304Backfill_SynthesizesFlatSelector runs migration 0304's
// verbatim backfill UPDATE against a bare in-flight row (selectors_snapshot
// '[]', flat scalars set) and proves it synthesizes exactly one
// role_in_fixed_area selector matching those scalars — the parity mapping the
// rewired drift path relies on for pre-cutover instances. The WHERE guard also
// makes it idempotent, so a second run is a no-op.
func TestMigration0304Backfill_SynthesizesFlatSelector(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	_, instanceID, authorID := selectorsSnapshotSeedInstance(t, database)

	// Seed a bare stage row exactly as a pre-6a instance would carry it: flat
	// scalars set, selectors_snapshot left at the column default '[]'.
	var stageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
		   eligible_actor_ids, status, stage_kind, required_capability_snapshot)
		VALUES ($1::uuid, 1, 'Stage 1', 'qms_admin', 'quality', 'any_1_of', 'reduce_quorum',
		        $2::jsonb, 'active', 'approval', 'document.approve')
		RETURNING id::text`,
		instanceID, `["`+authorID+`"]`,
	).Scan(&stageID); err != nil {
		t.Fatalf("seed bare stage: %v", err)
	}

	// Precondition: default is the empty array.
	if raw := selectorsSnapshotRaw(t, database, stageID); raw != "[]" {
		t.Fatalf("precondition: selectors_snapshot = %q, want []", raw)
	}

	// Migration 0304's backfill statement, verbatim.
	const backfill = `
		UPDATE public.approval_stage_instances
		   SET selectors_snapshot = jsonb_build_array(
		         jsonb_build_object(
		           'kind', 'role_in_fixed_area',
		           'role', required_role_snapshot,
		           'area_code', area_code_snapshot))
		 WHERE selectors_snapshot = '[]'::jsonb
		   AND required_role_snapshot <> ''
		   AND area_code_snapshot <> ''`
	if _, err := database.ExecContext(ctx, backfill); err != nil {
		t.Fatalf("run backfill: %v", err)
	}

	assertSelectorsEqual(t, decodeSelectors(t, database, stageID), []domain.ActorSelector{
		{Kind: domain.SelectorRoleInFixedArea, Role: "qms_admin", AreaCode: "quality"},
	})

	// Idempotency: a second run touches nothing (row is no longer '[]').
	if _, err := database.ExecContext(ctx, backfill); err != nil {
		t.Fatalf("run backfill (2nd): %v", err)
	}
	assertSelectorsEqual(t, decodeSelectors(t, database, stageID), []domain.ActorSelector{
		{Kind: domain.SelectorRoleInFixedArea, Role: "qms_admin", AreaCode: "quality"},
	})
}

func selectorsSnapshotRaw(t *testing.T, database *sql.DB, stageID string) string {
	t.Helper()
	var raw string
	if err := database.QueryRowContext(context.Background(),
		`SELECT selectors_snapshot::text FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&raw); err != nil {
		t.Fatalf("query selectors_snapshot for stage %s: %v", stageID, err)
	}
	return raw
}

func decodeSelectors(t *testing.T, database *sql.DB, stageID string) []domain.ActorSelector {
	t.Helper()
	var raw []byte
	if err := database.QueryRowContext(context.Background(),
		`SELECT selectors_snapshot FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&raw); err != nil {
		t.Fatalf("query selectors_snapshot for stage %s: %v", stageID, err)
	}
	var out []domain.ActorSelector
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal selectors_snapshot %q: %v", string(raw), err)
	}
	return out
}

func assertSelectorsEqual(t *testing.T, got, want []domain.ActorSelector) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("selectors length: got %d %+v, want %d %+v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("selector[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
