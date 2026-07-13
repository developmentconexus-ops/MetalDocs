//go:build integration
// +build integration

// Package approval_test — F8 (milestone 2b, approval-kernel-backend) SLA
// due_at wiring. Proves end-to-end, against a real Postgres testdb (ADR 0034
// harness), that the repository's activation-time due_at computation reads
// the per-stage-instance due_in_days_snapshot (NOT approval_route_stages),
// per approval-remediation-design.md §4/W4:
//
//  1. InsertStageInstances' INSERT ... CASE WHEN status='active' expression
//     (postgres_approval_repository.go) sets due_at ~ now()+due_in_days_snapshot
//     for a stage inserted directly as active, and leaves due_at NULL for a
//     pending stage even when it carries a non-NULL due_in_days_snapshot.
//  2. UpdateStageStatus's activation UPDATE ... CASE WHEN $1='active'
//     expression recomputes due_at from THAT stage's OWN
//     due_in_days_snapshot at the moment it transitions pending->active —
//     here NULL, so due_at stays NULL after activation (no-fallback
//     principle, spec.md §11: no substitute value silently applied).
//
// This targets the repository layer directly (InsertStageInstances /
// UpdateStageStatus), the same seams submit_service.go and
// review_verdict_service.go call in production, without going through the
// full SubmitService/ReviewVerdictService stack — avoiding the unrelated
// v_active_user_areas eligibility-pool dependency that a real
// SubmitRevisionForReview call would require to seed.
package approval_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

func sladueAtStageDueAt(t *testing.T, database *sql.DB, stageID string) *time.Time {
	t.Helper()
	var due sql.NullTime
	if err := database.QueryRowContext(context.Background(),
		`SELECT due_at FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&due); err != nil {
		t.Fatalf("query due_at for stage %s: %v", stageID, err)
	}
	if !due.Valid {
		return nil
	}
	return &due.Time
}

// TestInsertStageInstances_ActiveStageGetsDueAt_PendingStageStaysNull proves
// InsertStageInstances' due_at CASE expression: an active stage with
// due_in_days_snapshot=3 gets due_at~now()+3d; a pending stage with
// due_in_days_snapshot=5 (non-NULL) still gets due_at=NULL because it is not
// yet active — due_at is set only at activation, never at mere insert time
// for a not-yet-active stage.
func TestInsertStageInstances_ActiveStageGetsDueAt_PendingStageStaysNull(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

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

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	runner := db.NewTxRunner(database)

	three := 3
	five := 5
	stage1ID := uuid.NewString()
	stage2ID := uuid.NewString()

	err := runner.Do(ctx, func(tx *sql.Tx) error {
		return repo.InsertStageInstances(ctx, tx, []domain.StageInstance{
			{
				ID:                         stage1ID,
				ApprovalInstanceID:         instance.ID,
				StageOrder:                 1,
				NameSnapshot:               "Stage 1",
				RequiredRoleSnapshot:       "reviewer",
				RequiredCapabilitySnapshot: "document.review",
				AreaCodeSnapshot:           "QA",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				Kind:                       domain.StageKindReview,
				EligibleActorIDs:           []string{author.ID},
				Status:                     domain.StageActive,
				DueInDaysSnapshot:          &three,
			},
			{
				ID:                         stage2ID,
				ApprovalInstanceID:         instance.ID,
				StageOrder:                 2,
				NameSnapshot:               "Stage 2",
				RequiredRoleSnapshot:       "reviewer",
				RequiredCapabilitySnapshot: "document.review",
				AreaCodeSnapshot:           "QA",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				Kind:                       domain.StageKindReview,
				EligibleActorIDs:           []string{author.ID},
				Status:                     domain.StagePending,
				DueInDaysSnapshot:          &five,
			},
		})
	})
	if err != nil {
		t.Fatalf("InsertStageInstances: %v", err)
	}

	due1 := sladueAtStageDueAt(t, database, stage1ID)
	if due1 == nil {
		t.Fatalf("stage 1 (active, due_in_days_snapshot=3) due_at = NULL; want ~now()+3d")
	}
	wantMin := time.Now().UTC().Add(3*24*time.Hour - 5*time.Minute)
	wantMax := time.Now().UTC().Add(3*24*time.Hour + 5*time.Minute)
	if due1.Before(wantMin) || due1.After(wantMax) {
		t.Errorf("stage 1 due_at = %v; want within +-5min of now()+3d (%v..%v)", due1, wantMin, wantMax)
	}

	due2 := sladueAtStageDueAt(t, database, stage2ID)
	if due2 != nil {
		t.Errorf("stage 2 (pending, due_in_days_snapshot=5) due_at = %v; want NULL (not active yet)", *due2)
	}
}

// TestUpdateStageStatus_ActivationRecomputesDueAt_FromOwnSnapshot proves the
// UpdateStageStatus activation path: a stage transitioning pending->active
// gets due_at computed from ITS OWN due_in_days_snapshot at that moment. A
// NULL snapshot yields NULL due_at (no fallback substituted) even though the
// stage is now active — mirrors the review_verdict_service.go "activate next
// stage" call this feature wired the CASE expression into.
func TestUpdateStageStatus_ActivationRecomputesDueAt_FromOwnSnapshot(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

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

	// Seed stage 2 directly as pending with due_in_days_snapshot=NULL — the
	// exact shape SubmitService produces for a stage whose route config had
	// due_in_days=NULL (see submit_service.go's per-stage DueInDaysSnapshot
	// wiring).
	var stage2ID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
		   eligible_actor_ids, status, stage_kind, due_in_days_snapshot, required_capability_snapshot)
		VALUES ($1::uuid, 2, 'Stage 2', 'reviewer', 'QA', 'any_1_of', 'keep_snapshot',
		        $2::jsonb, 'pending', 'review', NULL, 'document.review')
		RETURNING id::text`,
		instance.ID, `["`+author.ID+`"]`,
	).Scan(&stage2ID); err != nil {
		t.Fatalf("seed stage 2 (pending): %v", err)
	}

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	runner := db.NewTxRunner(database)

	err := runner.Do(ctx, func(tx *sql.Tx) error {
		return repo.UpdateStageStatus(ctx, tx, tenant.ID, stage2ID, domain.StageActive, domain.StagePending)
	})
	if err != nil {
		t.Fatalf("UpdateStageStatus(activate stage 2): %v", err)
	}

	due2 := sladueAtStageDueAt(t, database, stage2ID)
	if due2 != nil {
		t.Errorf("stage 2 due_at after activation = %v; want NULL (due_in_days_snapshot=NULL, no-fallback principle)", *due2)
	}
}

// TestUpdateStageStatus_ActivationSetsDueAt_FromNonNullSnapshot is the
// positive counterpart: a stage with due_in_days_snapshot=7 that transitions
// pending->active gets due_at~now()+7d, computed at activation time (not at
// insert time, proving the UPDATE path — not just the INSERT path — performs
// this computation).
func TestUpdateStageStatus_ActivationSetsDueAt_FromNonNullSnapshot(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

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

	var stageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
		   eligible_actor_ids, status, stage_kind, due_in_days_snapshot, required_capability_snapshot)
		VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', 'QA', 'any_1_of', 'keep_snapshot',
		        $2::jsonb, 'pending', 'review', 7, 'document.review')
		RETURNING id::text`,
		instance.ID, `["`+author.ID+`"]`,
	).Scan(&stageID); err != nil {
		t.Fatalf("seed stage 1 (pending, due_in_days_snapshot=7): %v", err)
	}

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	runner := db.NewTxRunner(database)

	err := runner.Do(ctx, func(tx *sql.Tx) error {
		return repo.UpdateStageStatus(ctx, tx, tenant.ID, stageID, domain.StageActive, domain.StagePending)
	})
	if err != nil {
		t.Fatalf("UpdateStageStatus(activate): %v", err)
	}

	due := sladueAtStageDueAt(t, database, stageID)
	if due == nil {
		t.Fatalf("due_at after activation = NULL; want ~now()+7d (due_in_days_snapshot=7)")
	}
	wantMin := time.Now().UTC().Add(7*24*time.Hour - 5*time.Minute)
	wantMax := time.Now().UTC().Add(7*24*time.Hour + 5*time.Minute)
	if due.Before(wantMin) || due.After(wantMax) {
		t.Errorf("due_at = %v; want within +-5min of now()+7d (%v..%v)", due, wantMin, wantMax)
	}
}
