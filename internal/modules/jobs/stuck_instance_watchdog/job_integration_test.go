//go:build integration
// +build integration

package stuck_instance_watchdog

// M5 F5.8 T2 — honest alert-only proof (ADR 0068). The orphaned timeout-action
// branch was removed in T1: the watchdog is alert-only for every stuck (in_progress,
// submitted_at < now()-7d) approval instance, regardless of drift policy.
// These tests prove that against real Postgres via the canonical testdb
// factory (ADR 0034) — no sqlmock, no schema-impossible fixture values.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

// TestIntegration_Watchdog_P1_AlertOnly proves a genuinely stuck approval
// instance (in_progress, submitted 8 days ago) gets exactly one
// approval.instance.stuck_alert governance event emitted and is otherwise
// left completely untouched — instance stays in_progress, document stays
// under_review. A fresh (<7d) instance with the same drift policy is proven
// untouched, confirming the 7-day threshold is honored.
func TestIntegration_Watchdog_P1_AlertOnly(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := authz.WithBackgroundBypass(context.Background())

	tenant := testdb.NewTenant(t, db)

	// Stuck instance: in_progress, submitted 8 days ago — the watchdog must
	// alert on it, not touch its status.
	stuckDoc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("under_review"))
	stuckRoute := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenant.ID))
	stuckInstance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(stuckDoc),
		testdb.WithRoute(stuckRoute),
		testdb.WithStatus("in_progress"),
	)
	backdateSubmittedAt(t, db, stuckInstance.ID, 8*24*time.Hour)
	seedActiveStageSnapshot(t, db, stuckInstance.ID, "reduce_quorum")

	// Non-stuck instance: submitted just now, same drift policy — must
	// survive completely untouched (proves the 7-day threshold is honored).
	freshDoc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("under_review"))
	freshRoute := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenant.ID))
	freshInstance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(freshDoc),
		testdb.WithRoute(freshRoute),
		testdb.WithStatus("in_progress"),
	)
	seedActiveStageSnapshot(t, db, freshInstance.ID, "reduce_quorum")

	emitter := application.NewSQLEmitter()

	if err := run(ctx, db, emitter); err != nil {
		t.Fatalf("watchdog run: %v", err)
	}

	assertInstanceStatus(t, db, stuckInstance.ID, "in_progress")
	assertDocumentStatus(t, db, stuckDoc.ID, "under_review")

	assertInstanceStatus(t, db, freshInstance.ID, "in_progress")
	assertDocumentStatus(t, db, freshDoc.ID, "under_review")

	var stuckAlertCount int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM governance_events WHERE resource_id = $1 AND event_type = 'approval.instance.stuck_alert'`,
		stuckInstance.ID,
	).Scan(&stuckAlertCount); err != nil {
		t.Fatalf("count stuck alerts (stuck instance): %v", err)
	}
	if stuckAlertCount != 1 {
		t.Fatalf("stuck alert count for stuck instance = %d, want 1", stuckAlertCount)
	}

	var freshAlertCount int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM governance_events WHERE resource_id = $1 AND event_type = 'approval.instance.stuck_alert'`,
		freshInstance.ID,
	).Scan(&freshAlertCount); err != nil {
		t.Fatalf("count stuck alerts (fresh instance): %v", err)
	}
	if freshAlertCount != 0 {
		t.Fatalf("stuck alert count for fresh instance = %d, want 0", freshAlertCount)
	}
}

// TestIntegration_Watchdog_P1_AlertOnly_AnyDriftPolicy proves the alert-only
// behavior holds for a different valid drift policy value (keep_snapshot) —
// there is no branch in run() that special-cases any drift_policy string
// (ADR 0068 removed the last such orphaned timeout-action branch).
func TestIntegration_Watchdog_P1_AlertOnly_AnyDriftPolicy(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := authz.WithBackgroundBypass(context.Background())

	tenant := testdb.NewTenant(t, db)
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("under_review"))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenant.ID))
	instance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithStatus("in_progress"),
	)
	backdateSubmittedAt(t, db, instance.ID, 8*24*time.Hour)
	seedActiveStageSnapshot(t, db, instance.ID, "keep_snapshot")

	emitter := application.NewSQLEmitter()

	if err := run(ctx, db, emitter); err != nil {
		t.Fatalf("watchdog run: %v", err)
	}

	assertInstanceStatus(t, db, instance.ID, "in_progress")
	assertDocumentStatus(t, db, doc.ID, "under_review")

	var alertCount int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM governance_events WHERE resource_id = $1 AND event_type = 'approval.instance.stuck_alert'`,
		instance.ID,
	).Scan(&alertCount); err != nil {
		t.Fatalf("count stuck alerts: %v", err)
	}
	if alertCount != 1 {
		t.Fatalf("stuck alert count = %d, want 1", alertCount)
	}
}

// --- fixture / assertion helpers ---------------------------------------------

// backdateSubmittedAt rewrites the instance's submitted_at to age in the past,
// simulating a stuck instance without waiting real wall-clock time. No
// tripwire on approval_instances UPDATE for this column.
func backdateSubmittedAt(t *testing.T, db *sql.DB, instanceID string, age time.Duration) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`UPDATE approval_instances SET submitted_at = now() - $2::interval WHERE id = $1::uuid`,
		instanceID, age.String(),
	); err != nil {
		t.Fatalf("backdateSubmittedAt: %v", err)
	}
}

// seedActiveStageSnapshot inserts an active approval_stage_instances row for
// instanceID carrying the given on_eligibility_drift_snapshot policy — the
// column listStuckInstances joins on and surfaces in the stuck-alert payload.
// Valid values per the CHECK constraint: reduce_quorum, fail_stage,
// keep_snapshot.
func seedActiveStageSnapshot(t *testing.T, db *sql.DB, instanceID, driftPolicy string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO approval_stage_instances
		  (id, approval_instance_id, stage_order, name_snapshot,
		   required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		   quorum_snapshot, on_eligibility_drift_snapshot, eligible_actor_ids, status, opened_at)
		VALUES (gen_random_uuid(), $1::uuid, 1, 'Stage 1',
		        'reviewer', 'document.signoff', 'QA',
		        'any_1_of', $2, '[]'::jsonb, 'active', now())`,
		instanceID, driftPolicy,
	); err != nil {
		t.Fatalf("seedActiveStageSnapshot: %v", err)
	}
}

func assertInstanceStatus(t *testing.T, db *sql.DB, instanceID, want string) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(context.Background(),
		`SELECT status FROM approval_instances WHERE id = $1::uuid`, instanceID,
	).Scan(&got); err != nil {
		t.Fatalf("assertInstanceStatus: %v", err)
	}
	if got != want {
		t.Fatalf("instance %s status = %q, want %q", instanceID, got, want)
	}
}

func assertDocumentStatus(t *testing.T, db *sql.DB, docID, want string) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(context.Background(),
		`SELECT status FROM public.documents WHERE id = $1::uuid`, docID,
	).Scan(&got); err != nil {
		t.Fatalf("assertDocumentStatus: %v", err)
	}
	if got != want {
		t.Fatalf("document %s status = %q, want %q", docID, got, want)
	}
}
