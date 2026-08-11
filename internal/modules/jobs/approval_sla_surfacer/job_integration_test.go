//go:build integration
// +build integration

package approval_sla_surfacer

// F8 (milestone 2b, approval-kernel-backend) — integration proof for the
// River periodic approval SLA surfacer's core run() body against real
// Postgres via the canonical testdb factory (ADR 0034). Mirrors
// document_review_surfacer/job_integration_test.go's structure and proves
// the equivalent contract for the DISTINCT per-stage due_at mechanism:
//
//  1. cross-tenant coverage: an overdue active stage in tenant A AND one in
//     tenant B both get sla_surfaced_at set by a single run() call; a
//     not-yet-due active stage in tenant A is left untouched.
//  2. tenant-seed isolation: MarkStagesOverdueSurfaced scoped to tenant A
//     only does not touch tenant B's overdue stage.
//  3. idempotency: running run() twice surfaces an overdue stage ONCE.
//  4. alert-only (ADR 0068): a run() call never mutates the stage's status
//     or due_at — only sla_surfaced_at changes.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// recordingNotifier is a test-local double for
// approvaldomain.ApprovalNotificationEnqueuer that records every call instead
// of enqueuing a River job, so these tests assert on what the surfacer
// PROMISED to notify without needing a live River worker/queue. Mirrors
// application/submit_notification_integration_test.go's recordingNotifier in
// the approval application package — that copy is unexported and lives in a
// different package, so this is a second, correct-Go copy, not duplication to
// avoid (per the task brief).
type recordingNotifier struct {
	sent []approvaldomain.ApprovalNotificationArgs
}

func (r *recordingNotifier) EnqueueApprovalNotificationTx(_ context.Context, _ platformdb.Tx, args approvaldomain.ApprovalNotificationArgs) error {
	r.sent = append(r.sent, args)
	return nil
}

// assertSameUsers fails the test unless got and want contain exactly the same
// user ids, order-insensitive. Not exported/shared: mirrors
// submit_notification_integration_test.go's copy in a different package —
// no importable test-only helper package exists for this, so each package
// defines its own (per the task brief's resolved ambiguity).
func assertSameUsers(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("recipients = %v, want %v", got, want)
	}
	gotSet := make(map[string]int, len(got))
	for _, id := range got {
		gotSet[id]++
	}
	wantSet := make(map[string]int, len(want))
	for _, id := range want {
		wantSet[id]++
	}
	for id, n := range wantSet {
		if gotSet[id] != n {
			t.Fatalf("recipients = %v, want %v", got, want)
		}
	}
	for id, n := range gotSet {
		if wantSet[id] != n {
			t.Fatalf("recipients = %v, want %v", got, want)
		}
	}
}

// seedOverdueStageFixture seeds tenant/user/document/route/instance plus one
// active stage instance with the given due_at (nil = no SLA configured, i.e.
// due_at NULL) and sla_surfaced_at left NULL. Mirrors
// review_verdict_integration_test.go's seedReviewVerdictFixture raw-SQL
// FK-chain seeding, extended with due_at (no factory builder exists for a
// stage_instance with a specific due_at).
func seedOverdueStageFixture(t *testing.T, db *sql.DB, tenantID string, dueAt *time.Time) (stageID, instanceID string) {
	t.Helper()
	ctx := context.Background()

	author := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	doc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenantID))
	instance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
	)

	eligible := `["` + author.ID + `"]`
	var id string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
		   eligible_actor_ids, status, stage_kind, due_at, required_capability_snapshot)
		VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', 'QA', 'any_1_of', 'keep_snapshot',
		        $2::jsonb, 'active', 'review', $3, 'document.review')
		RETURNING id::text`,
		instance.ID, eligible, nullableTime(dueAt),
	).Scan(&id); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}
	return id, instance.ID
}

// seedOverdueStageFixtureForActor is seedOverdueStageFixture's twin, except
// the stage's eligible_actor_ids snapshot is exactly {eligibleActorID} —
// a caller-supplied user, not the fixture's own auto-created author — so the
// Task 7 notification tests can assert the surfacer's recipient list against
// a user they hold a reference to. No factory builder exists for this
// (testdb.WithEligibleActors was checked and does not exist in this repo).
func seedOverdueStageFixtureForActor(t *testing.T, db *sql.DB, tenantID string, dueAt *time.Time, eligibleActorID string) (stageID, instanceID string) {
	t.Helper()
	ctx := context.Background()

	author := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	doc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenantID))
	instance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
	)

	eligible := `["` + eligibleActorID + `"]`
	var id string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
		   eligible_actor_ids, status, stage_kind, due_at, required_capability_snapshot)
		VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', 'QA', 'any_1_of', 'keep_snapshot',
		        $2::jsonb, 'active', 'review', $3, 'document.review')
		RETURNING id::text`,
		instance.ID, eligible, nullableTime(dueAt),
	).Scan(&id); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}
	return id, instance.ID
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func readSLASurfacedAt(t *testing.T, db *sql.DB, stageID string) *time.Time {
	t.Helper()
	var ts sql.NullTime
	if err := db.QueryRowContext(context.Background(),
		`SELECT sla_surfaced_at FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&ts); err != nil {
		t.Fatalf("readSLASurfacedAt: %v", err)
	}
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

func readStageStatusAndDueAt(t *testing.T, db *sql.DB, stageID string) (status string, dueAt *time.Time) {
	t.Helper()
	var due sql.NullTime
	if err := db.QueryRowContext(context.Background(),
		`SELECT status, due_at FROM public.approval_stage_instances WHERE id = $1::uuid`, stageID,
	).Scan(&status, &due); err != nil {
		t.Fatalf("readStageStatusAndDueAt: %v", err)
	}
	if due.Valid {
		dueAt = &due.Time
	}
	return status, dueAt
}

// TestIntegration_SLASurfacer_FullTick_IteratesAllTenants proves one run()
// tick surfaces overdue active stages in BOTH tenant A and tenant B, while a
// not-yet-due active stage in tenant A is left untouched.
func TestIntegration_SLASurfacer_FullTick_IteratesAllTenants(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	overdue := now.Add(-1 * time.Hour)
	future := now.Add(30 * 24 * time.Hour)

	overdueStageA, _ := seedOverdueStageFixture(t, db, tntA.ID, &overdue)
	notDueStageA, _ := seedOverdueStageFixture(t, db, tntA.ID, &future)
	overdueStageB, _ := seedOverdueStageFixture(t, db, tntB.ID, &overdue)

	reader := approvalrepo.NewSLAOverdueReaderPG(db)
	writer := approvalrepo.NewSLASurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, platformdb.NewTxRunner(db), reader, writer, approvaldomain.NoopApprovalNotificationEnqueuer{}, now); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := readSLASurfacedAt(t, db, overdueStageA); got == nil || !got.Equal(now) {
		t.Fatalf("tenant A overdue stage sla_surfaced_at = %v, want %v", got, now)
	}
	if got := readSLASurfacedAt(t, db, overdueStageB); got == nil || !got.Equal(now) {
		t.Fatalf("tenant B overdue stage sla_surfaced_at = %v, want %v (full tick must iterate every tenant)", got, now)
	}
	if got := readSLASurfacedAt(t, db, notDueStageA); got != nil {
		t.Fatalf("tenant A not-yet-due stage sla_surfaced_at = %v, want nil (untouched)", got)
	}
}

// TestIntegration_SLASurfacer_Writer_TenantSeed_DoesNotSurfaceOtherTenant is
// the isolation proof: seed the tx as tenant A ONLY, call
// MarkStagesOverdueSurfaced(ctx, tx, tntA.ID, now) directly, and assert
// tenant A's overdue stage is surfaced while tenant B's is not touched.
// Isolation is via the explicit tenant_id predicate (join to
// approval_instances), not solely RLS (dev/CI DB role is superuser +
// BYPASSRLS) — mirrors document_review_surfacer's twin test rationale.
func TestIntegration_SLASurfacer_Writer_TenantSeed_DoesNotSurfaceOtherTenant(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	overdue := now.Add(-1 * time.Hour)

	stageA, _ := seedOverdueStageFixture(t, db, tntA.ID, &overdue)
	stageB, _ := seedOverdueStageFixture(t, db, tntB.ID, &overdue)

	writer := approvalrepo.NewSLASurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		t.Fatalf("BypassSystem: %v", err)
	}
	if err := authz.SeedTxTenant(ctx, tx, tntA.ID); err != nil {
		t.Fatalf("SeedTxTenant(A): %v", err)
	}

	if _, err := writer.MarkStagesOverdueSurfaced(ctx, tx, tntA.ID, now); err != nil {
		t.Fatalf("MarkStagesOverdueSurfaced: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if got := readSLASurfacedAt(t, db, stageA); got == nil || !got.Equal(now) {
		t.Fatalf("tenant A overdue stage sla_surfaced_at = %v, want %v (seeded tenant must be surfaced)", got, now)
	}
	if got := readSLASurfacedAt(t, db, stageB); got != nil {
		t.Fatalf("tenant B overdue stage sla_surfaced_at = %v, want nil (must not surface outside seeded tenant A)", got)
	}
}

// TestIntegration_SLASurfacer_Idempotent_SecondRunNoOp proves running run()
// twice surfaces an overdue stage ONCE: the second run flips 0 rows and
// sla_surfaced_at is unchanged.
func TestIntegration_SLASurfacer_Idempotent_SecondRunNoOp(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	overdue := now.Add(-1 * time.Hour)

	stageID, _ := seedOverdueStageFixture(t, db, tnt.ID, &overdue)

	reader := approvalrepo.NewSLAOverdueReaderPG(db)
	writer := approvalrepo.NewSLASurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, platformdb.NewTxRunner(db), reader, writer, approvaldomain.NoopApprovalNotificationEnqueuer{}, now); err != nil {
		t.Fatalf("first run: %v", err)
	}
	firstSurfacedAt := readSLASurfacedAt(t, db, stageID)
	if firstSurfacedAt == nil || !firstSurfacedAt.Equal(now) {
		t.Fatalf("after first run sla_surfaced_at = %v, want %v", firstSurfacedAt, now)
	}

	secondNow := now.Add(10 * time.Minute)
	if err := run(ctx, platformdb.NewTxRunner(db), reader, writer, approvaldomain.NoopApprovalNotificationEnqueuer{}, secondNow); err != nil {
		t.Fatalf("second run: %v", err)
	}
	secondSurfacedAt := readSLASurfacedAt(t, db, stageID)
	if secondSurfacedAt == nil || !secondSurfacedAt.Equal(*firstSurfacedAt) {
		t.Fatalf("after second run sla_surfaced_at = %v, want unchanged %v (idempotency guard must flip 0 rows)", secondSurfacedAt, firstSurfacedAt)
	}
}

// TestIntegration_SLASurfacer_AlertOnly_DoesNotMutateStatusOrDueAt proves
// ADR 0068 alert-only compliance directly: after a run() call surfaces an
// overdue stage, the stage's status and due_at are UNCHANGED from before the
// run — only sla_surfaced_at is written.
func TestIntegration_SLASurfacer_AlertOnly_DoesNotMutateStatusOrDueAt(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	overdue := now.Add(-1 * time.Hour)

	stageID, _ := seedOverdueStageFixture(t, db, tnt.ID, &overdue)
	statusBefore, dueBefore := readStageStatusAndDueAt(t, db, stageID)

	reader := approvalrepo.NewSLAOverdueReaderPG(db)
	writer := approvalrepo.NewSLASurfaceWriterPG(db)
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, platformdb.NewTxRunner(db), reader, writer, approvaldomain.NoopApprovalNotificationEnqueuer{}, now); err != nil {
		t.Fatalf("run: %v", err)
	}

	statusAfter, dueAfter := readStageStatusAndDueAt(t, db, stageID)
	if statusAfter != statusBefore {
		t.Fatalf("status changed by surfacer run: before=%q after=%q; want unchanged (alert-only, ADR 0068)", statusBefore, statusAfter)
	}
	if dueBefore == nil || dueAfter == nil || !dueBefore.Equal(*dueAfter) {
		t.Fatalf("due_at changed by surfacer run: before=%v after=%v; want unchanged (alert-only, ADR 0068)", dueBefore, dueAfter)
	}
}

// TestSurfacer_OverdueStage_EnqueuesApprovalOverdue proves Task 7's whole
// point: surfacing an overdue stage must also tell the people holding it.
// Alert-only (ADR 0068) still means alert — it constrains what the job
// MUTATES (only sla_surfaced_at), not whether it communicates.
func TestSurfacer_OverdueStage_EnqueuesApprovalOverdue(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	approver := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID))
	now := time.Now().UTC().Truncate(time.Second)
	overdue := now.Add(-1 * time.Hour)

	seedOverdueStageFixtureForActor(t, db, tnt.ID, &overdue, approver.ID)

	reader := approvalrepo.NewSLAOverdueReaderPG(db)
	writer := approvalrepo.NewSLASurfaceWriterPG(db)
	notifier := &recordingNotifier{}
	ctx := authz.WithBackgroundBypass(context.Background())

	if err := run(ctx, platformdb.NewTxRunner(db), reader, writer, notifier, now); err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(notifier.sent) != 1 {
		t.Fatalf("enqueued %d events, want 1", len(notifier.sent))
	}
	got := notifier.sent[0]
	if got.EventType != approvaldomain.EventTypeApprovalOverdue {
		t.Fatalf("event type = %q, want %q", got.EventType, approvaldomain.EventTypeApprovalOverdue)
	}
	if got.ActorUserID != "" {
		t.Fatalf("actor = %q, want empty — an SLA tick has no human cause", got.ActorUserID)
	}
	assertSameUsers(t, got.RecipientUserIDs, approver.ID)
}

// TestSurfacer_TwoTicksSameDeadline_EnqueuesOnce proves two ticks against the
// same deadline surface once. sla_surfaced_at IS the guard — no second
// "already reminded" flag is invented, because two flags that must agree are
// a bug with a date on it.
func TestSurfacer_TwoTicksSameDeadline_EnqueuesOnce(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	approver := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID))
	now := time.Now().UTC().Truncate(time.Second)
	overdue := now.Add(-1 * time.Hour)

	seedOverdueStageFixtureForActor(t, db, tnt.ID, &overdue, approver.ID)

	reader := approvalrepo.NewSLAOverdueReaderPG(db)
	writer := approvalrepo.NewSLASurfaceWriterPG(db)
	notifier := &recordingNotifier{}
	ctx := authz.WithBackgroundBypass(context.Background())

	for i := 0; i < 2; i++ {
		if err := run(ctx, platformdb.NewTxRunner(db), reader, writer, notifier, now); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}

	if len(notifier.sent) != 1 {
		t.Fatalf("enqueued %d events, want 1 — the second tick has nothing new to surface", len(notifier.sent))
	}
}
