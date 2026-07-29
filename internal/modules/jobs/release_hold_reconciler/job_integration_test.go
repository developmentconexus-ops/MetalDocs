//go:build integration
// +build integration

package release_hold_reconciler

// ADR 0085 Stage C W2 — the reconciliation sweep, proved against real Postgres
// via the canonical testdb factory (ADR 0034). No sqlmock: the whole value of
// this job is WHICH rows its predicate selects, which is a database question.
//
// The contract under test:
//   - a generation stuck in readiness hold past the threshold alerts exactly
//     once per tick, and NOTHING about it is mutated (alert-only, ADR 0068);
//   - a never-evaluated generation alerts with an honest "never_evaluated"
//     reason rather than a blank field;
//   - a released generation, a freshly-evaluated one, a just-created one, a
//     scheduled one whose timer is still legitimately armed, and a superseded
//     (non-head) one all stay silent — the four ways this sweep could turn into
//     permanent noise;
//   - a scheduled generation whose plan date has PASSED does alert (dead timer);
//   - the sweep is cross-tenant and per-tenant-seeded: two tenants' stuck
//     generations are both reported, each under its own tenant_id.

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/application"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

// reconcilerContentHash is a schema-valid 64-hex frozen content hash
// (ck_release_generations_frozen_hash).
const reconcilerContentHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

// genSpec describes one seeded public.release_generations row. Ages are
// expressed as durations in the past so the fixture never depends on wall-clock
// timing.
type genSpec struct {
	documentID string
	// createdAgo is how long ago the generation was created.
	createdAgo time.Duration
	// lastEvaluatedAgo is how long ago the coordinator last evaluated it; nil
	// means never evaluated (last_evaluated_at IS NULL).
	lastEvaluatedAgo *time.Duration
	holdReason       string
	holdDetail       string
	// released marks the generation released (which requires both facts, per
	// ck_release_generations_released_implies_facts).
	released bool
}

func ago(d time.Duration) *time.Duration { return &d }

func TestIntegration_ReleaseHoldReconciler_AlertsOnlyStuckGenerations(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := authz.WithBackgroundBypass(context.Background())

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)

	approvedDoc := func(tenantID string) string {
		return testdb.NewDocument(t, db,
			testdb.WithTenant(tenantID),
			testdb.WithStatus("approved"),
		).ID
	}

	// --- tenant A: one of each shape -----------------------------------------

	// (1) STUCK: unreleased, old, last evaluated well past the threshold.
	stuckID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID:       approvedDoc(tenantA.ID),
		createdAgo:       2 * time.Hour,
		lastEvaluatedAgo: ago(90 * time.Minute),
		holdReason:       "materializing",
		holdDetail:       "final pdf never landed",
	})

	// (2) STUCK, never evaluated — the lost-fact / dead-lettered-consumer case.
	neverEvaluatedID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID: approvedDoc(tenantA.ID),
		createdAgo: 2 * time.Hour,
	})

	// (3) STUCK: scheduled document whose plan date has PASSED — dead timer.
	deadTimerDoc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenantA.ID),
		testdb.WithStatus("scheduled"),
		testdb.WithPlannedEffectiveFrom(time.Now().UTC().Add(-1*time.Hour)),
	)
	deadTimerID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID:       deadTimerDoc.ID,
		createdAgo:       3 * time.Hour,
		lastEvaluatedAgo: ago(2 * time.Hour),
		holdReason:       "awaiting_effective_date",
	})

	// (4) RELEASED: the release happened; never an alert.
	releasedID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID:       approvedDoc(tenantA.ID),
		createdAgo:       2 * time.Hour,
		lastEvaluatedAgo: ago(2 * time.Hour),
		released:         true,
	})

	// (5) FRESH HOLD: evaluated inside the threshold — in flight, not stuck.
	freshID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID:       approvedDoc(tenantA.ID),
		createdAgo:       2 * time.Hour,
		lastEvaluatedAgo: ago(1 * time.Minute),
		holdReason:       "materializing",
	})

	// (6) JUST CREATED: never evaluated, but younger than the threshold.
	newbornID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID: approvedDoc(tenantA.ID),
		createdAgo: 1 * time.Minute,
	})

	// (7) ARMED TIMER: scheduled with a plan date still in the future. Old and
	// unevaluated by the letter of the sweep, but genuinely waiting — the
	// single largest false-positive source if the predicate ignored the plan.
	armedDoc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenantA.ID),
		testdb.WithStatus("scheduled"),
		testdb.WithPlannedEffectiveFrom(time.Now().UTC().Add(7*24*time.Hour)),
	)
	armedID := seedGeneration(t, db, tenantA.ID, genSpec{
		documentID:       armedDoc.ID,
		createdAgo:       3 * time.Hour,
		lastEvaluatedAgo: ago(2 * time.Hour),
		holdReason:       "awaiting_effective_date",
	})

	// --- tenant B: one stuck generation, to prove the cross-tenant sweep -----
	stuckBID := seedGeneration(t, db, tenantB.ID, genSpec{
		documentID:       approvedDoc(tenantB.ID),
		createdAgo:       2 * time.Hour,
		lastEvaluatedAgo: ago(45 * time.Minute),
		holdReason:       "failed",
	})

	before := snapshotGeneration(t, db, stuckID)

	if err := run(ctx, db, approvalrepo.NewReleaseHoldReaderPG(db), application.NewSQLEmitter(), time.Now().UTC()); err != nil {
		t.Fatalf("reconciler run: %v", err)
	}

	assertAlertCount(t, db, stuckID, 1)
	assertAlertCount(t, db, neverEvaluatedID, 1)
	assertAlertCount(t, db, deadTimerID, 1)
	assertAlertCount(t, db, stuckBID, 1)

	assertAlertCount(t, db, releasedID, 0)
	assertAlertCount(t, db, freshID, 0)
	assertAlertCount(t, db, newbornID, 0)
	assertAlertCount(t, db, armedID, 0)

	// Alert-only (ADR 0068): the sweep must not have touched the row it alerted
	// on — not the hold, not the evaluation stamp, not the released state.
	after := snapshotGeneration(t, db, stuckID)
	if after != before {
		t.Fatalf("stuck generation mutated by an alert-only sweep:\n before = %+v\n after  = %+v", before, after)
	}

	// Each alert is attributed to the system principal, carries the honest
	// reason and lands under its own tenant.
	stuckEvent := loadAlert(t, db, stuckID)
	if stuckEvent.actorUserID != SystemActor {
		t.Fatalf("alert actor = %q, want %q", stuckEvent.actorUserID, SystemActor)
	}
	if stuckEvent.tenantID != tenantA.ID {
		t.Fatalf("alert tenant = %q, want %q", stuckEvent.tenantID, tenantA.ID)
	}
	if stuckEvent.payload["hold_reason"] != "materializing" {
		t.Fatalf("alert hold_reason = %v, want materializing", stuckEvent.payload["hold_reason"])
	}
	if stuckEvent.payload["hold_detail"] != "final pdf never landed" {
		t.Fatalf("alert hold_detail = %v, want the seeded detail", stuckEvent.payload["hold_detail"])
	}

	neverEvent := loadAlert(t, db, neverEvaluatedID)
	if neverEvent.payload["hold_reason"] != "never_evaluated" {
		t.Fatalf("never-evaluated alert hold_reason = %v, want never_evaluated", neverEvent.payload["hold_reason"])
	}
	if neverEvent.payload["last_evaluated_at"] != "" {
		t.Fatalf("never-evaluated alert last_evaluated_at = %v, want empty (no substitute value)", neverEvent.payload["last_evaluated_at"])
	}

	tenantBEvent := loadAlert(t, db, stuckBID)
	if tenantBEvent.tenantID != tenantB.ID {
		t.Fatalf("tenant B alert tenant = %q, want %q", tenantBEvent.tenantID, tenantB.ID)
	}
}

// TestIntegration_ReleaseHoldReconciler_SupersededGenerationSilent proves the
// freeze-head conjunct: when a document was returned to draft and re-approved,
// the OLD generation can never release (ADR 0085 EvaluateRelease no-ops on a
// non-head), so alerting on it forever would be permanent noise. Only the head
// generation is reported.
func TestIntegration_ReleaseHoldReconciler_SupersededGenerationSilent(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := authz.WithBackgroundBypass(context.Background())

	tenant := testdb.NewTenant(t, db)
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("approved"))

	// Both generations are stuck-shaped; only the newer one is the freeze head.
	oldGenID := seedGeneration(t, db, tenant.ID, genSpec{
		documentID:       doc.ID,
		createdAgo:       4 * time.Hour,
		lastEvaluatedAgo: ago(3 * time.Hour),
		holdReason:       "materializing",
	})
	headGenID := seedGeneration(t, db, tenant.ID, genSpec{
		documentID:       doc.ID,
		createdAgo:       2 * time.Hour,
		lastEvaluatedAgo: ago(90 * time.Minute),
		holdReason:       "materializing",
	})

	if err := run(ctx, db, approvalrepo.NewReleaseHoldReaderPG(db), application.NewSQLEmitter(), time.Now().UTC()); err != nil {
		t.Fatalf("reconciler run: %v", err)
	}

	assertAlertCount(t, db, oldGenID, 0)
	assertAlertCount(t, db, headGenID, 1)
}

// TestIntegration_ReleaseHoldReconciler_NonReleasableDocumentSilent proves the
// document-status conjunct: a generation whose document left the releasable
// status set (here: back to draft) is not waiting for anything, so it must not
// be alerted on.
func TestIntegration_ReleaseHoldReconciler_NonReleasableDocumentSilent(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := authz.WithBackgroundBypass(context.Background())

	tenant := testdb.NewTenant(t, db)
	draftDoc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("draft"))

	draftGenID := seedGeneration(t, db, tenant.ID, genSpec{
		documentID:       draftDoc.ID,
		createdAgo:       2 * time.Hour,
		lastEvaluatedAgo: ago(90 * time.Minute),
		holdReason:       "materializing",
	})

	if err := run(ctx, db, approvalrepo.NewReleaseHoldReaderPG(db), application.NewSQLEmitter(), time.Now().UTC()); err != nil {
		t.Fatalf("reconciler run: %v", err)
	}

	assertAlertCount(t, db, draftGenID, 0)
}

// --- fixture / assertion helpers ---------------------------------------------

// seedGeneration inserts one public.release_generations row with the ages the
// spec describes. Raw insert by design: the table carries no tripwire
// (migration 0310: "written only by the coordinator"), and the point of these
// tests is to place rows in states the happy path cannot reach on demand.
func seedGeneration(t *testing.T, db *sql.DB, tenantID string, spec genSpec) string {
	t.Helper()

	var lastEvaluated any
	if spec.lastEvaluatedAgo != nil {
		lastEvaluated = time.Now().UTC().Add(-*spec.lastEvaluatedAgo)
	}

	var approvalFactAt, artifactFactAt, releasedAt any
	var docxKey, pdfKey any
	if spec.released {
		// ck_release_generations_released_implies_facts +
		// ck_release_generations_artifact_fact_full_set.
		now := time.Now().UTC()
		approvalFactAt = now.Add(-spec.createdAgo)
		artifactFactAt = now.Add(-spec.createdAgo)
		releasedAt = now.Add(-spec.createdAgo)
		docxKey = "tenants/" + tenantID + "/final.docx"
		pdfKey = "tenants/" + tenantID + "/final.pdf"
	}

	var id string
	if err := db.QueryRowContext(context.Background(), `
		INSERT INTO public.release_generations
			(tenant_id, subject_kind, document_id, approval_instance_id, revision_id,
			 revision_version, frozen_content_hash,
			 approval_fact_at, final_docx_s3_key, final_pdf_s3_key, artifact_fact_at,
			 hold_reason, hold_detail, released_at, last_evaluated_at, created_at)
		VALUES ($1::uuid, 'document', $2::uuid, gen_random_uuid(), gen_random_uuid(),
		        1, $3,
		        $4::timestamptz, $5, $6, $7::timestamptz,
		        NULLIF($8, ''), NULLIF($9, ''), $10::timestamptz, $11::timestamptz,
		        now() - $12::interval)
		RETURNING id::text`,
		tenantID, spec.documentID, reconcilerContentHash,
		approvalFactAt, docxKey, pdfKey, artifactFactAt,
		spec.holdReason, spec.holdDetail, releasedAt, lastEvaluated,
		spec.createdAgo.String(),
	).Scan(&id); err != nil {
		t.Fatalf("seedGeneration: %v", err)
	}
	return id
}

// generationSnapshot is the mutable coordinator state the sweep must never
// touch.
type generationSnapshot struct {
	holdReason    sql.NullString
	holdDetail    sql.NullString
	releasedAt    sql.NullTime
	lastEvaluated sql.NullTime
	updatedAt     time.Time
}

func snapshotGeneration(t *testing.T, db *sql.DB, generationID string) generationSnapshot {
	t.Helper()
	var s generationSnapshot
	if err := db.QueryRowContext(context.Background(), `
		SELECT hold_reason, hold_detail, released_at, last_evaluated_at, updated_at
		  FROM public.release_generations WHERE id = $1::uuid`, generationID,
	).Scan(&s.holdReason, &s.holdDetail, &s.releasedAt, &s.lastEvaluated, &s.updatedAt); err != nil {
		t.Fatalf("snapshotGeneration: %v", err)
	}
	return s
}

func assertAlertCount(t *testing.T, db *sql.DB, generationID string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.governance_events
		  WHERE resource_id = $1 AND event_type = $2`,
		generationID, AlertEventType,
	).Scan(&got); err != nil {
		t.Fatalf("assertAlertCount: %v", err)
	}
	if got != want {
		t.Fatalf("stuck-hold alert count for generation %s = %d, want %d", generationID, got, want)
	}
}

type alertEvent struct {
	tenantID     string
	actorUserID  string
	resourceType string
	payload      map[string]any
}

func loadAlert(t *testing.T, db *sql.DB, generationID string) alertEvent {
	t.Helper()
	var ev alertEvent
	var raw []byte
	if err := db.QueryRowContext(context.Background(), `
		SELECT tenant_id::text, actor_user_id, resource_type, payload_json
		  FROM public.governance_events
		 WHERE resource_id = $1 AND event_type = $2`,
		generationID, AlertEventType,
	).Scan(&ev.tenantID, &ev.actorUserID, &ev.resourceType, &raw); err != nil {
		t.Fatalf("loadAlert: %v", err)
	}
	if err := json.Unmarshal(raw, &ev.payload); err != nil {
		t.Fatalf("loadAlert: unmarshal payload: %v", err)
	}
	if ev.resourceType != "release_generation" {
		t.Fatalf("alert resource_type = %q, want release_generation", ev.resourceType)
	}
	return ev
}
