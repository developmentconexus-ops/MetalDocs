//go:build integration
// +build integration

package stuck_instance_watchdog

// M5 F5.2 T6 — P1 proof. After the River job-consolidation migration removed
// the custom lease scheduler + its advisory lock (ADR 0067), this test proves
// the extracted run(...) body still auto-cancels a genuinely stuck approval
// instance identically to the pre-migration behavior, against real Postgres
// via the canonical testdb factory (ADR 0034) — no sqlmock.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

func TestIntegration_Watchdog_P1_AutoCancelEquivalence(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := authz.WithBackgroundBypass(context.Background())

	tenant := testdb.NewTenant(t, db)

	// Stuck instance: in_progress, submitted 8 days ago, active stage snapshot
	// says auto_cancel — the watchdog must cancel it.
	stuckDoc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("under_review"))
	stuckRoute := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenant.ID))
	stuckInstance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(stuckDoc),
		testdb.WithRoute(stuckRoute),
		testdb.WithStatus("in_progress"),
	)
	backdateSubmittedAt(t, db, stuckInstance.ID, 8*24*time.Hour)
	seedActiveStageSnapshot(t, db, stuckInstance.ID, "auto_cancel")

	// Non-stuck instance: submitted just now, same drift policy — must survive
	// completely untouched (proves the 7-day threshold is honored).
	freshDoc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("under_review"))
	freshRoute := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenant.ID))
	freshInstance := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(freshDoc),
		testdb.WithRoute(freshRoute),
		testdb.WithStatus("in_progress"),
	)
	seedActiveStageSnapshot(t, db, freshInstance.ID, "auto_cancel")

	repo := repository.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	emitter := application.NewSQLEmitter()
	services := application.NewServices(repo, emitter, application.RealClock{}, controlleddocumentsdomain.NoopCDFieldReader{})

	runner := platformdb.NewTxRunner(db)

	if err := run(ctx, db, runner, services.Cancel, emitter); err != nil {
		t.Fatalf("watchdog run: %v", err)
	}

	assertInstanceStatus(t, db, stuckInstance.ID, "cancelled")
	assertDocumentStatus(t, db, stuckDoc.ID, "draft")

	assertInstanceStatus(t, db, freshInstance.ID, "in_progress")
	assertDocumentStatus(t, db, freshDoc.ID, "under_review")
}

// TestIntegration_Watchdog_P1_AlertOnlyEquivalence proves the alert-only
// (non auto_cancel) drift policy path emits a governance event and leaves the
// instance untouched — the companion branch of the same run(...) body.
func TestIntegration_Watchdog_P1_AlertOnlyEquivalence(t *testing.T) {
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
	seedActiveStageSnapshot(t, db, instance.ID, "reduce_quorum")

	repo := repository.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	emitter := application.NewSQLEmitter()
	services := application.NewServices(repo, emitter, application.RealClock{}, controlleddocumentsdomain.NoopCDFieldReader{})

	runner := platformdb.NewTxRunner(db)

	if err := run(ctx, db, runner, services.Cancel, emitter); err != nil {
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
// column listStuckInstances joins on to decide auto_cancel vs alert-only.
func seedActiveStageSnapshot(t *testing.T, db *sql.DB, instanceID, driftPolicy string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO approval_stage_instances
		  (id, approval_instance_id, stage_order, name_snapshot,
		   required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		   quorum_snapshot, on_eligibility_drift_snapshot, eligible_actor_ids, status, opened_at)
		VALUES (gen_random_uuid(), $1::uuid, 1, 'Stage 1',
		        'reviewer', 'doc.signoff', 'QA',
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
		`SELECT status FROM documents WHERE id = $1::uuid`, docID,
	).Scan(&got); err != nil {
		t.Fatalf("assertDocumentStatus: %v", err)
	}
	if got != want {
		t.Fatalf("document %s status = %q, want %q", docID, got, want)
	}
}
