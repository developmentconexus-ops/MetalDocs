//go:build integration
// +build integration

// Package approval_test — ADR 0085 release coordinator (Stage A). Exercises
// approval-driven publication end-to-end against a real Postgres instance
// (testdb per-test clone):
//
//   - the SIGNOFF route produces the approval fact; artifact facts complete the
//     set; one Evaluate releases the document with an ACTUAL effective_from
//   - the REVIEW-VERDICT route does exactly the same thing (F-QA4-14
//     regression: before ADR 0085 this route never pinned and never released)
//   - a future planned_effective_from schedules and arms a timer; firing the
//     timer publishes with the plan PRESERVED and the actual = evaluation time
//   - two concurrent evaluations produce exactly one winner (source CAS)
//   - a supersession target that moved rolls the WHOLE release back and records
//     a visible supersede_conflict hold
//   - a release that supersedes the published head emits a governance event for
//     the target's OWN transition, not only for the published source
//   - a plan-named CROSS-controlled-document target's lifecycle event carries
//     ITS controlled document, not the released source's
//   - a missing artifact holds on `materializing` and releases once it lands
//   - the DB itself refuses a second published head per controlled document, so
//     no writer — present or future — can break the supersession invariant
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsapp "metaldocs/internal/modules/documents/application"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// releaseContentHash is the frozen pin the fixture's approval instance carries.
// The signoff route's OCC guard compares the client echo against exactly this
// value, and it is the generation identity's frozen_content_hash.
const releaseContentHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// ── test doubles ────────────────────────────────────────────────────────────

// capturingReleaseEnqueuer records every coordinator evaluation the fact
// producers and the coordinator itself enqueue, instead of inserting a River
// job. The River leg is a thin InsertTx wrapper (approval/jobs); what these
// tests need to assert is WHICH generation was enqueued and WHEN it was armed
// for — which is exactly what the port carries.
type capturingReleaseEnqueuer struct {
	mu   sync.Mutex
	keys []domain.ReleaseGenerationKey
	runs []time.Time
}

func (c *capturingReleaseEnqueuer) EnqueueReleaseEvaluationTx(_ context.Context, _ db.Tx, key domain.ReleaseGenerationKey, runAt time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys = append(c.keys, key)
	c.runs = append(c.runs, runAt)
	return nil
}

func (c *capturingReleaseEnqueuer) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.keys)
}

func (c *capturingReleaseEnqueuer) last() (domain.ReleaseGenerationKey, time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.keys) == 0 {
		return domain.ReleaseGenerationKey{}, time.Time{}, false
	}
	return c.keys[len(c.keys)-1], c.runs[len(c.runs)-1], true
}

// noopPin stands in for the documents-module FreezeService. The pin's own
// behaviour (placeholder resolution, materialize dispatch) is covered by the
// documents module's suites; here it must simply not be the reason a terminal
// approval fails, so the release path itself is what is under test. It still
// records the call count — "did this route pin at all" IS the F-QA4-14
// assertion.
type noopPin struct {
	mu    sync.Mutex
	calls int
}

func (p *noopPin) Pin(_ context.Context, _ db.Tx, _, _ string, _ docsapp.ApproverContext, _ string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return nil
}

func (p *noopPin) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// capturingLifecycleEnqueuer records the notification-fanout rows the release
// transaction enqueues. The River leg is a thin InsertTx wrapper; what these
// tests assert is the ADDRESSING of each event — notification fanout resolves
// its reader audience from ControlledDocumentID, so an event carrying the wrong
// controlled document notifies the wrong readers.
type capturingLifecycleEnqueuer struct {
	mu     sync.Mutex
	events []docsdomain.LifecycleEventArgs
}

func (c *capturingLifecycleEnqueuer) EnqueueLifecycleEventTx(_ context.Context, _ db.Tx, args docsdomain.LifecycleEventArgs) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, args)
	return nil
}

// find returns the single lifecycle event for (resourceID, eventType).
func (c *capturingLifecycleEnqueuer) find(resourceID, eventType string) (docsdomain.LifecycleEventArgs, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, ev := range c.events {
		if ev.ResourceID == resourceID && ev.EventType == eventType {
			return ev, true
		}
	}
	return docsdomain.LifecycleEventArgs{}, false
}

// fixedReviewInterval is a stub ProfileReviewIntervalReader. The real adapter
// performs an authz-recording taxonomy read, which the coordinator is required
// to run OFF-lock (H-PRE-1); this stub keeps the review-cycle rule assertable
// without seeding a profile cadence.
type fixedReviewInterval struct{ days int }

func (f fixedReviewInterval) ReviewIntervalDays(_ context.Context, _, _ string) (int, error) {
	return f.days, nil
}

// ── fixture ─────────────────────────────────────────────────────────────────

type releaseFixture struct {
	tenantID     string
	authorID     string
	approverID   string
	documentID   string
	controlledID string
	instanceID   string
	stageID      string
	revisionID   string
	runner       db.TxRunner
	database     *sql.DB
	decision     *application.DecisionService
	verdict      *application.ReviewVerdictService
	coordinator  *application.ReleaseCoordinator
	enqueuer     *capturingReleaseEnqueuer
	lifecycle    *capturingLifecycleEnqueuer
	pin          *noopPin
}

func (fx releaseFixture) ctxFor(actorID string) context.Context {
	return testdb.AuthzCtx(fx.tenantID, actorID)
}

// seedReleaseFixture builds a document at under_review with ONE active stage of
// the requested kind, its instance already frozen, a real current_revision_id
// (the generation identity keys on it), and the area grants both routes to
// terminal approval need.
func seedReleaseFixture(t *testing.T, database *sql.DB, stageKind domain.StageKind) releaseFixture {
	t.Helper()
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"))
	approver := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Approver"))

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("under_review"),
		testdb.WithSubmitReadySnapshots(),
	)

	revisionID := seedReleaseRevision(t, database, doc.TenantID, doc.ID, author.ID)

	// process_area_code_snapshot resolves the authz area without a
	// controlled-document read-port. profile_code_snapshot is what the
	// coordinator's off-lock preflight resolves the review cadence from.
	// current_revision_id is what RecordApprovalFactTx keys the generation on
	// (it fails closed without one).
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE public.documents d
			   SET process_area_code_snapshot = 'qa',
			       profile_code_snapshot      = cd.profile_code,
			       current_revision_id        = $1::uuid
			  FROM public.controlled_documents cd
			 WHERE cd.id = d.controlled_document_id
			   AND d.id  = $2::uuid`,
			revisionID, doc.ID)
		return err
	})

	route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))
	instance := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("in_progress"),
		// A real instance that has crossed the freeze boundary carries this pin;
		// the generation identity is meaningless without it (fail-closed).
		testdb.WithFrozenContentHash(releaseContentHash),
	)

	testdb.SeedWithCaps(t, database, `[{"cap":"membership.manage"},{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ('qa', $1::uuid, 'QA') ON CONFLICT (tenant_id, code) DO NOTHING`,
			tenant.ID,
		); err != nil {
			return err
		}
		// 'approver' carries document.signoff (the approval-kind stage gate),
		// approval.review (the review-kind stage gate) AND document.edit (the
		// terminal document transition) — one role satisfies every tier-2 check
		// on both routes, and ux_user_process_areas_one_active permits only one
		// active row per (user, tenant, area) anyway.
		_, err := tx.ExecContext(ctx, `INSERT INTO metaldocs.user_process_areas
			(user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			VALUES ($1,$2,'qa','approver',now(),NULL,$1)`, approver.ID, tenant.ID)
		return err
	})

	requiredCap := "document.signoff"
	requiredRole := "approver"
	if stageKind == domain.StageKindReview {
		requiredCap = "approval.review"
		requiredRole = "reviewer"
	}
	var stageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
		VALUES ($1::uuid, 1, 'Release Stage', $2, $3, 'qa',
		        'any_1_of', 'keep_snapshot', $4::jsonb, 'active', $5)
		RETURNING id::text`,
		instance.ID, requiredRole, requiredCap, `["`+approver.ID+`"]`, string(stageKind),
	).Scan(&stageID); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}

	repo := infrastructure.NewPostgresApprovalRepository(database, iampostgres.NewUserDisplayNameRepository(database))
	runner := db.NewTxRunner(database)
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)

	enqueuer := &capturingReleaseEnqueuer{}
	pin := &noopPin{}
	// ONE recorder into BOTH terminal-approval routes — the production wiring
	// shape (Services.WithReleaseRecorder), so this fixture cannot accidentally
	// reproduce the asymmetry F-QA4-14 was.
	services.WithReleaseRecorder(application.NewReleaseFactRecorder(pin, enqueuer))

	lifecycle := &capturingLifecycleEnqueuer{}
	coordinator := application.NewReleaseCoordinator(repo, application.NewSQLEmitter(), application.RealClock{}).
		WithEvaluationEnqueuer(enqueuer).
		WithLifecycleEnqueuer(lifecycle)

	return releaseFixture{
		tenantID:     tenant.ID,
		authorID:     author.ID,
		approverID:   approver.ID,
		documentID:   doc.ID,
		controlledID: doc.ControlledDocumentID,
		instanceID:   instance.ID,
		stageID:      stageID,
		revisionID:   revisionID,
		runner:       runner,
		database:     database,
		decision:     services.Decision,
		verdict:      services.ReviewVerdict,
		coordinator:  coordinator,
		enqueuer:     enqueuer,
		lifecycle:    lifecycle,
		pin:          pin,
	}
}

// seedReleaseRevision inserts the editor_sessions + document_revisions pair the
// documents.current_revision_id FK requires (neither table carries a tripwire —
// mirrors testdb.InsertDraftDocument's raw-insert idiom).
func seedReleaseRevision(t *testing.T, database *sql.DB, tenantID, docID, userID string) string {
	t.Helper()
	ctx := context.Background()
	var sessID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.editor_sessions
			(tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		VALUES ($1::uuid, $2::uuid, $3, now() + interval '1 hour',
		        '00000000-0000-0000-0000-000000000000', 'active')
		RETURNING id::text`,
		tenantID, docID, userID,
	).Scan(&sessID); err != nil {
		t.Fatalf("seedReleaseRevision: insert editor_sessions: %v", err)
	}
	var revID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.document_revisions
			(document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		VALUES ($1::uuid, NULL, $2::uuid, '', $3, '{}')
		RETURNING id::text`,
		docID, sessID, releaseContentHash,
	).Scan(&revID); err != nil {
		t.Fatalf("seedReleaseRevision: insert document_revisions: %v", err)
	}
	return revID
}

// ── assertions / drivers ────────────────────────────────────────────────────

// generationState is the release_generations row as the tests observe it. It
// carries the full identity so every Evaluate call is driven by the generation
// the fact producer actually wrote — never by the document's CURRENT
// revision_version, which the release itself increments.
type generationState struct {
	id                 string
	tenantID           string
	documentID         string
	approvalInstanceID string
	revisionID         string
	revisionVersion    int
	frozenContentHash  string
	approvalFactAt     sql.NullTime
	artifactFactAt     sql.NullTime
	holdReason         sql.NullString
	releasedAt         sql.NullTime
}

func (g generationState) key() domain.ReleaseGenerationKey {
	return domain.ReleaseGenerationKey{
		TenantID:           g.tenantID,
		SubjectKind:        domain.SubjectKindDocument,
		DocumentID:         g.documentID,
		ApprovalInstanceID: g.approvalInstanceID,
		RevisionID:         g.revisionID,
		RevisionVersion:    g.revisionVersion,
		FrozenContentHash:  g.frozenContentHash,
	}
}

func loadGeneration(t *testing.T, database *sql.DB, documentID string) generationState {
	t.Helper()
	var g generationState
	if err := database.QueryRowContext(context.Background(), `
		SELECT id::text, tenant_id::text, document_id::text, approval_instance_id::text,
		       revision_id::text, revision_version, frozen_content_hash,
		       approval_fact_at, artifact_fact_at, hold_reason, released_at
		  FROM public.release_generations
		 WHERE document_id = $1::uuid
		 ORDER BY generation_seq DESC
		 LIMIT 1`, documentID,
	).Scan(&g.id, &g.tenantID, &g.documentID, &g.approvalInstanceID,
		&g.revisionID, &g.revisionVersion, &g.frozenContentHash,
		&g.approvalFactAt, &g.artifactFactAt, &g.holdReason, &g.releasedAt); err != nil {
		t.Fatalf("load release generation for %s: %v", documentID, err)
	}
	return g
}

type documentRelease struct {
	status      string
	effective   sql.NullTime
	planned     sql.NullTime
	effectiveTo sql.NullTime
	reviewDue   sql.NullTime
	superseded  sql.NullString
	revisionVer int
}

func loadDocumentRelease(t *testing.T, database *sql.DB, documentID string) documentRelease {
	t.Helper()
	var d documentRelease
	if err := database.QueryRowContext(context.Background(), `
		SELECT status, effective_from, planned_effective_from, effective_to,
		       review_due_at, superseded_document_id::text, revision_version
		  FROM public.documents WHERE id = $1::uuid`, documentID,
	).Scan(&d.status, &d.effective, &d.planned, &d.effectiveTo, &d.reviewDue, &d.superseded, &d.revisionVer); err != nil {
		t.Fatalf("load document %s: %v", documentID, err)
	}
	return d
}

// recordArtifacts writes the artifact-fact halves for a generation the way the
// render workers do: each in its own transaction, the PDF completing the set.
func (fx releaseFixture) recordArtifacts(t *testing.T, generationID string, kinds ...application.ArtifactKind) {
	t.Helper()
	for _, kind := range kinds {
		kind := kind
		if err := fx.runner.Do(context.Background(), func(tx *sql.Tx) error {
			_, err := application.RecordArtifactFactTx(context.Background(), tx, fx.enqueuer, application.ArtifactFactInput{
				TenantID:            fx.tenantID,
				ReleaseGenerationID: generationID,
				Kind:                kind,
				S3Key:               "tenants/" + fx.tenantID + "/" + fx.documentID + "/" + string(kind),
			})
			return err
		}); err != nil {
			t.Fatalf("record %s artifact fact: %v", kind, err)
		}
	}
}

// approveViaSignoff drives the signoff route to terminal approval.
func (fx releaseFixture) approveViaSignoff(t *testing.T) application.SignoffResult {
	t.Helper()
	res, err := fx.decision.RecordSignoff(fx.ctxFor(fx.approverID), fx.runner, application.SignoffRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.stageID,
		ActorUserID:      fx.approverID,
		Decision:         domain.DecisionApprove,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
		ContentFormData:  map[string]any{"_content_hash": releaseContentHash},
	})
	if err != nil {
		t.Fatalf("RecordSignoff(approve): %v", err)
	}
	if !res.InstanceApproved {
		t.Fatalf("expected instance approved, got %+v", res)
	}
	return res
}

// setDocumentPlan writes the plan date directly. The plan-authoring endpoint is
// Stage B's submission-plan contract; Stage A's coordinator only READS the
// plan, so the tests set it the way a plan would leave it.
func (fx releaseFixture) setDocumentPlan(t *testing.T, planned time.Time) {
	t.Helper()
	testdb.SeedWithCaps(t, fx.database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`UPDATE public.documents SET planned_effective_from = $1 WHERE id = $2::uuid`,
			planned, fx.documentID)
		return err
	})
}

// ── tests ───────────────────────────────────────────────────────────────────

// TestRelease_SignoffRoute_EndToEnd is the ADR 0085 happy path: terminal
// approval records the approval fact and enqueues an evaluation in the SAME
// transaction; the artifact facts complete the set; one Evaluate publishes with
// an ACTUAL effective_from.
func TestRelease_SignoffRoute_EndToEnd(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	fx.approveViaSignoff(t)

	if got := fx.pin.callCount(); got != 1 {
		t.Fatalf("pin calls = %d, want 1 (terminal approval must pin)", got)
	}
	gen := loadGeneration(t, database, fx.documentID)
	if !gen.approvalFactAt.Valid {
		t.Fatal("approval fact not recorded in the terminal-approval transaction")
	}
	if gen.artifactFactAt.Valid {
		t.Fatal("artifact fact stamped before any artifact was written")
	}
	key, runAt, ok := fx.enqueuer.last()
	if !ok {
		t.Fatal("terminal approval did not enqueue a coordinator evaluation")
	}
	if !runAt.IsZero() {
		t.Fatalf("post-approval evaluation runAt = %v, want zero (evaluate now)", runAt)
	}
	if key.DocumentID != fx.documentID || key.RevisionID != fx.revisionID || key.FrozenContentHash != releaseContentHash {
		t.Fatalf("enqueued key %+v does not match the generation identity", key)
	}

	// Approved but not materialized: the predicate must hold, not release.
	if got := documentStatus(t, database, fx.documentID); got != "approved" {
		t.Fatalf("document status = %q, want approved", got)
	}
	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (pre-artifact): %v", err)
	}
	if eval.Action != domain.ReleaseActionHold || eval.Hold != domain.HoldMaterializing {
		t.Fatalf("pre-artifact evaluation = %+v, want hold/materializing", eval)
	}

	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)
	if !gen.artifactFactAt.Valid {
		t.Fatal("artifact fact not stamped once the full DOCX+PDF set existed")
	}

	eval, err = fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (release): %v", err)
	}
	if eval.Action != domain.ReleaseActionRelease {
		t.Fatalf("evaluation = %+v, want release", eval)
	}

	doc := loadDocumentRelease(t, database, fx.documentID)
	if doc.status != "published" {
		t.Fatalf("document status = %q, want published", doc.status)
	}
	if !doc.effective.Valid {
		t.Fatal("published document has no effective_from — this is exactly F-QA4-13")
	}
	if doc.planned.Valid {
		t.Fatalf("planned_effective_from = %v, want NULL (no plan was declared)", doc.planned.Time)
	}
	gen = loadGeneration(t, database, fx.documentID)
	if !gen.releasedAt.Valid {
		t.Fatal("release generation not stamped released_at")
	}
	if gen.holdReason.Valid {
		t.Fatalf("released generation still carries hold %q", gen.holdReason.String)
	}

	// The release governance event is attributed to the system principal, with
	// the causal identities in the payload.
	var actor, payload string
	if err := database.QueryRowContext(context.Background(), `
		SELECT actor_user_id, payload_json::text
		  FROM public.governance_events
		 WHERE resource_id = $1 AND event_type = 'document_published'
		 ORDER BY created_at DESC LIMIT 1`, fx.documentID,
	).Scan(&actor, &payload); err != nil {
		t.Fatalf("load release governance event: %v", err)
	}
	if actor != application.ReleaseSystemPrincipal {
		t.Fatalf("release event actor = %q, want %q", actor, application.ReleaseSystemPrincipal)
	}
	for _, want := range []string{fx.approverID, gen.id, fx.revisionID} {
		if !strings.Contains(payload, want) {
			t.Fatalf("release event payload missing causal identity %q: %s", want, payload)
		}
	}

	// Idempotency: a redelivered evaluation must be a no-op, not a re-release.
	eval, err = fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (replay): %v", err)
	}
	if eval.Action != domain.ReleaseActionNoop {
		t.Fatalf("replayed evaluation = %+v, want noop", eval)
	}
}

// TestRelease_ReviewVerdictRoute_PinsAndReleases is the F-QA4-14 regression.
// Before ADR 0085 a document approved through a review verdict never called
// Pin, never materialized, and could never be published. It must now be
// indistinguishable from the signoff route.
func TestRelease_ReviewVerdictRoute_PinsAndReleases(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindReview)

	res, err := fx.verdict.RecordVerdict(fx.ctxFor(fx.approverID), fx.runner, application.ReviewVerdictRequest{
		TenantID:        fx.tenantID,
		InstanceID:      fx.instanceID,
		StageInstanceID: fx.stageID,
		ActorUserID:     fx.approverID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict(ready): %v", err)
	}
	if !res.InstanceApproved {
		t.Fatalf("expected instance approved (single review stage), got %+v", res)
	}

	// F-QA4-14: the pin is the thing that never happened on this route.
	if got := fx.pin.callCount(); got != 1 {
		t.Fatalf("pin calls = %d, want 1 — the review-verdict route must pin exactly like signoff", got)
	}
	gen := loadGeneration(t, database, fx.documentID)
	if !gen.approvalFactAt.Valid {
		t.Fatal("review-verdict terminal approval recorded no approval fact")
	}
	if _, _, ok := fx.enqueuer.last(); !ok {
		t.Fatal("review-verdict terminal approval enqueued no coordinator evaluation")
	}

	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)
	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if eval.Action != domain.ReleaseActionRelease {
		t.Fatalf("evaluation = %+v, want release", eval)
	}
	doc := loadDocumentRelease(t, database, fx.documentID)
	if doc.status != "published" || !doc.effective.Valid {
		t.Fatalf("document = %+v, want published with effective_from set", doc)
	}
}

// TestRelease_FuturePlannedDate_SchedulesThenTimerPublishes: the plan is
// immutable and the actual is the evaluation instant. The timer's payload is
// the same generation key, and firing it is just another Evaluate.
func TestRelease_FuturePlannedDate_SchedulesThenTimerPublishes(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	planned := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Millisecond)
	fx.setDocumentPlan(t, planned)

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)

	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (schedule): %v", err)
	}
	if eval.Action != domain.ReleaseActionSchedule || eval.Hold != domain.HoldAwaitingEffectiveDate {
		t.Fatalf("evaluation = %+v, want schedule/awaiting_effective_date", eval)
	}
	if eval.ScheduledFor == nil || !eval.ScheduledFor.Equal(planned) {
		t.Fatalf("armed timer = %v, want %v", eval.ScheduledFor, planned)
	}
	key, runAt, _ := fx.enqueuer.last()
	if !runAt.Equal(planned) {
		t.Fatalf("timer runAt = %v, want the planned date %v", runAt, planned)
	}
	if key.DocumentID != fx.documentID {
		t.Fatalf("timer payload key %+v is not this generation", key)
	}

	doc := loadDocumentRelease(t, database, fx.documentID)
	if doc.status != "scheduled" {
		t.Fatalf("document status = %q, want scheduled", doc.status)
	}
	if doc.effective.Valid {
		t.Fatal("scheduled document must not carry an ACTUAL effective_from")
	}
	if !doc.planned.Valid || !doc.planned.Time.Equal(planned) {
		t.Fatalf("planned_effective_from = %v, want %v", doc.planned, planned)
	}

	// Fire the timer: the plan is now in the past.
	past := time.Now().UTC().Add(-time.Hour).Truncate(time.Millisecond)
	fx.setDocumentPlan(t, past)

	before := time.Now().UTC()
	eval, err = fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (timer fired): %v", err)
	}
	if eval.Action != domain.ReleaseActionRelease {
		t.Fatalf("evaluation = %+v, want release", eval)
	}

	doc = loadDocumentRelease(t, database, fx.documentID)
	if doc.status != "published" {
		t.Fatalf("document status = %q, want published", doc.status)
	}
	// The plan is immutable; the actual is when the release really happened.
	if !doc.planned.Valid || !doc.planned.Time.Equal(past) {
		t.Fatalf("planned_effective_from mutated: %v, want %v", doc.planned, past)
	}
	if !doc.effective.Valid {
		t.Fatal("published document has no effective_from")
	}
	if doc.effective.Time.Before(before) {
		t.Fatalf("effective_from %v predates the evaluation (%v) — it must be the ACTUAL release instant, not the plan",
			doc.effective.Time, before)
	}
}

// TestRelease_ConcurrentEvaluations_ExactlyOneWinner: two evaluations of the
// same ready generation race. The source CAS admits one; the loser reports a
// no-op, never an error and never a second release.
func TestRelease_ConcurrentEvaluations_ExactlyOneWinner(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)
	key := gen.key()

	var wg sync.WaitGroup
	results := make([]application.ReleaseEvaluation, 2)
	errs := make([]error, 2)
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = fx.coordinator.Evaluate(context.Background(), fx.runner, key)
		}(i)
	}
	close(start)
	wg.Wait()

	releases := 0
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("evaluation %d errored: %v", i, errs[i])
		}
		if results[i].Action == domain.ReleaseActionRelease {
			releases++
		}
	}
	if releases != 1 {
		t.Fatalf("release winners = %d, want exactly 1 (results: %+v)", releases, results)
	}

	if got := documentStatus(t, database, fx.documentID); got != "published" {
		t.Fatalf("document status = %q, want published", got)
	}
	// Exactly one release event: the CAS is what makes the emit exactly-once.
	var events int
	if err := database.QueryRowContext(context.Background(), `
		SELECT count(*) FROM public.governance_events
		 WHERE resource_id = $1 AND event_type = 'document_published'`, fx.documentID,
	).Scan(&events); err != nil {
		t.Fatalf("count release events: %v", err)
	}
	if events != 1 {
		t.Fatalf("document_published events = %d, want exactly 1", events)
	}
}

// TestRelease_SupersedeTargetConflict_RollsBackAndHolds: the plan names a
// supersession target that is no longer the published head. The WHOLE release
// transaction rolls back — no partial release — and the conflict is recorded as
// a VISIBLE hold rather than lost as an error.
func TestRelease_SupersedeTargetConflict_RollsBackAndHolds(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	// The target the plan names has already left `published` (a concurrent
	// cutover superseded it), which is exactly what MarkSuperseded refuses.
	// revision_number 1 is only there to satisfy ux_documents_cd_revision (one
	// document per (controlled_document, revision_number)) — the conflict path
	// never reads it.
	staleTarget := testdb.NewDocument(t, database,
		testdb.WithTenant(fx.tenantID),
		testdb.WithOwner(fx.authorID),
		testdb.WithControlledDoc(testdb.ControlledDoc{
			ID: fx.controlledID, TenantID: fx.tenantID, OwnerUserID: fx.authorID,
		}),
		testdb.WithStatus("superseded"),
		testdb.WithRevisionNumber(1),
	)
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`UPDATE public.documents SET superseded_document_id = $1::uuid WHERE id = $2::uuid`,
			staleTarget.ID, fx.documentID)
		return err
	})

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)

	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if eval.Action != domain.ReleaseActionHold || eval.Hold != domain.HoldSupersedeConflict {
		t.Fatalf("evaluation = %+v, want hold/supersede_conflict", eval)
	}

	// Nothing partially applied.
	doc := loadDocumentRelease(t, database, fx.documentID)
	if doc.status != "approved" {
		t.Fatalf("document status = %q, want approved (release must have rolled back whole)", doc.status)
	}
	if doc.effective.Valid {
		t.Fatal("rolled-back release still wrote effective_from")
	}
	gen = loadGeneration(t, database, fx.documentID)
	if !gen.holdReason.Valid || gen.holdReason.String != string(domain.HoldSupersedeConflict) {
		t.Fatalf("generation hold = %v, want supersede_conflict", gen.holdReason)
	}
	if gen.releasedAt.Valid {
		t.Fatal("conflicted generation was stamped released_at")
	}
}

// TestRelease_Supersession_EmitsGovernanceEventPerTarget: ADR 0085 requires
// exactly one governance event AND one lifecycle event per affected transition.
// The published source is one transition; EACH superseded target is another and
// gets its own governance event — the source's published event names the
// targets but is not the audit record for them.
func TestRelease_Supersession_EmitsGovernanceEventPerTarget(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	// The controlled document's current published head: the implicit
	// same-document supersession target the coordinator discovers by itself.
	// revision_number 1 satisfies ux_documents_cd_revision (one document per
	// (controlled_document, revision_number)).
	priorHead := testdb.NewDocument(t, database,
		testdb.WithTenant(fx.tenantID),
		testdb.WithOwner(fx.authorID),
		testdb.WithControlledDoc(testdb.ControlledDoc{
			ID: fx.controlledID, TenantID: fx.tenantID, OwnerUserID: fx.authorID,
		}),
		testdb.WithStatus("published"),
		testdb.WithRevisionNumber(1),
	)

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)

	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if eval.Action != domain.ReleaseActionRelease {
		t.Fatalf("evaluation = %+v, want release", eval)
	}
	if len(eval.SupersededDocumentIDs) != 1 || eval.SupersededDocumentIDs[0] != priorHead.ID {
		t.Fatalf("superseded = %v, want [%s]", eval.SupersededDocumentIDs, priorHead.ID)
	}
	if got := documentStatus(t, database, priorHead.ID); got != "superseded" {
		t.Fatalf("prior head status = %q, want superseded", got)
	}

	// Exactly one governance event for the target's own transition, attributed
	// to the system principal and carrying the release identity.
	var events int
	if err := database.QueryRowContext(context.Background(), `
		SELECT count(*) FROM public.governance_events
		 WHERE resource_id = $1 AND event_type = 'document_superseded'`, priorHead.ID,
	).Scan(&events); err != nil {
		t.Fatalf("count superseded governance events: %v", err)
	}
	if events != 1 {
		t.Fatalf("document_superseded events for %s = %d, want exactly 1", priorHead.ID, events)
	}

	var actor, payload string
	if err := database.QueryRowContext(context.Background(), `
		SELECT actor_user_id, payload_json::text
		  FROM public.governance_events
		 WHERE resource_id = $1 AND event_type = 'document_superseded'
		 ORDER BY created_at DESC LIMIT 1`, priorHead.ID,
	).Scan(&actor, &payload); err != nil {
		t.Fatalf("load superseded governance event: %v", err)
	}
	if actor != application.ReleaseSystemPrincipal {
		t.Fatalf("superseded event actor = %q, want %q", actor, application.ReleaseSystemPrincipal)
	}
	// Release generation, the superseding document, and the causal approver.
	for _, want := range []string{gen.id, fx.documentID, fx.approverID} {
		if !strings.Contains(payload, want) {
			t.Fatalf("superseded event payload missing %q: %s", want, payload)
		}
	}
}

// TestRelease_CrossControlledDocumentSupersede_AddressesTargetsOwnCD: a
// publication plan may name a supersession target belonging to a DIFFERENT
// controlled document (one controlled document retired in favour of another).
// The target's lifecycle event must be addressed with the TARGET's controlled
// document — notification fanout resolves the reader audience from it, so
// inheriting the source's would notify the wrong readers and leave the retired
// document's own readers uninformed.
func TestRelease_CrossControlledDocumentSupersede_AddressesTargetsOwnCD(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	// CD-2: a published document under its own controlled document.
	otherCD := testdb.NewControlledDoc(t, database,
		testdb.WithTenant(fx.tenantID),
		testdb.WithOwner(fx.authorID))
	crossTarget := testdb.NewDocument(t, database,
		testdb.WithTenant(fx.tenantID),
		testdb.WithOwner(fx.authorID),
		testdb.WithControlledDoc(otherCD),
		testdb.WithStatus("published"),
		testdb.WithRevisionNumber(1))

	// The plan names it: superseded_document_id is the plan's cross-document
	// target, carried on the source row (Stage A reads the plan; the
	// plan-authoring endpoint is Stage B).
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`UPDATE public.documents SET superseded_document_id = $1::uuid WHERE id = $2::uuid`,
			crossTarget.ID, fx.documentID)
		return err
	})

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)

	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if eval.Action != domain.ReleaseActionRelease {
		t.Fatalf("evaluation = %+v, want release", eval)
	}
	if len(eval.SupersededDocumentIDs) != 1 || eval.SupersededDocumentIDs[0] != crossTarget.ID {
		t.Fatalf("superseded = %v, want [%s]", eval.SupersededDocumentIDs, crossTarget.ID)
	}
	if got := documentStatus(t, database, crossTarget.ID); got != "superseded" {
		t.Fatalf("cross-CD target status = %q, want superseded", got)
	}

	published, ok := fx.lifecycle.find(fx.documentID, docsdomain.EventTypeDocumentPublished)
	if !ok {
		t.Fatal("release enqueued no published lifecycle event for the source")
	}
	if published.ControlledDocumentID != fx.controlledID {
		t.Fatalf("published lifecycle event controlled document = %q, want the source's %q",
			published.ControlledDocumentID, fx.controlledID)
	}

	superseded, ok := fx.lifecycle.find(crossTarget.ID, docsdomain.EventTypeDocumentSuperseded)
	if !ok {
		t.Fatal("release enqueued no superseded lifecycle event for the cross-CD target")
	}
	if superseded.ControlledDocumentID != otherCD.ID {
		t.Fatalf("superseded lifecycle event controlled document = %q, want the TARGET's own %q (the source's is %q)",
			superseded.ControlledDocumentID, otherCD.ID, fx.controlledID)
	}
}

// TestRelease_SecondPublishedHeadRefusedByDB pins ADR 0085's supersession-head
// invariant to the DATABASE, with no application code in the way at all.
//
// Before Stage B this was driven through PublishService.PublishApproved, which
// published an approved document WITHOUT superseding the controlled document's
// existing head. That service is gone — publication is now the release
// coordinator's decision — but the guarantee it accidentally probed is the
// point and outlives it: no amount of application code, present or future, can
// create a second published row for one controlled document, because
// ux_documents_published_head (migration 0310) refuses it, and MapPgError turns
// the resulting 23505 into a conflict rather than a naked 500.
//
// Driving it by raw SQL is deliberate: an assertion about a DB invariant should
// not be mediated by whichever writer happens to exist this quarter.
func TestRelease_SecondPublishedHeadRefusedByDB(t *testing.T) {
	database, _ := testdb.Open(t)
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	admin := testdb.NewUser(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithRole("system_admin"))

	head := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(admin.ID),
		testdb.WithStatus("published"),
		testdb.WithRevisionNumber(1))

	// A second revision of the SAME controlled document, approved and about to
	// claim the head slot without demoting the incumbent.
	challenger := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(admin.ID),
		testdb.WithControlledDoc(testdb.ControlledDoc{
			ID: head.ControlledDocumentID, TenantID: tenant.ID, OwnerUserID: admin.ID,
		}),
		testdb.WithStatus("approved"),
		testdb.WithRevisionNumber(2))

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tenant.ID, admin.ID); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)

	_, err = tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET status           = 'published',
		       effective_from   = now(),
		       revision_version = revision_version + 1
		 WHERE id        = $1::uuid
		   AND tenant_id = $2::uuid`,
		challenger.ID, tenant.ID)
	if err == nil {
		t.Fatal("a SECOND published head was accepted — ux_documents_published_head is not enforcing the ADR 0085 supersession-head invariant")
	}
	mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
	if !errors.Is(mapped, infrastructure.ErrStaleRevision) {
		t.Fatalf("mapped error = %v, want infrastructure.ErrStaleRevision (409 conflict), not an unmapped database error", mapped)
	}
	// The unique-index violation aborted the transaction; roll it back so the
	// assertions below read committed state.
	_ = tx.Rollback()

	var publishedRows int
	if err := database.QueryRowContext(ctx, `
		SELECT count(*) FROM public.documents
		 WHERE tenant_id = $1::uuid AND controlled_document_id = $2::uuid AND status = 'published'`,
		tenant.ID, head.ControlledDocumentID,
	).Scan(&publishedRows); err != nil {
		t.Fatalf("count published rows: %v", err)
	}
	if publishedRows != 1 {
		t.Fatalf("published rows for the controlled document = %d, want exactly 1", publishedRows)
	}
	if got := documentStatus(t, database, challenger.ID); got != "approved" {
		t.Fatalf("challenger status = %q, want approved (the whole tx must roll back)", got)
	}
	if got := documentStatus(t, database, head.ID); got != "published" {
		t.Fatalf("existing head status = %q, want published (untouched)", got)
	}
}

// TestRelease_ArtifactMissing_HoldsMaterializingThenReleases: half an artifact
// set is not readiness. The hold is visible and clears by itself when the
// second half lands.
func TestRelease_ArtifactMissing_HoldsMaterializingThenReleases(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)

	// Only the DOCX half.
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx)
	gen = loadGeneration(t, database, fx.documentID)
	if gen.artifactFactAt.Valid {
		t.Fatal("artifact fact stamped on a half set — artifact-ready(G) requires DOCX AND PDF")
	}

	eval, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (half set): %v", err)
	}
	if eval.Action != domain.ReleaseActionHold || eval.Hold != domain.HoldMaterializing {
		t.Fatalf("evaluation = %+v, want hold/materializing", eval)
	}
	gen = loadGeneration(t, database, fx.documentID)
	if !gen.holdReason.Valid || gen.holdReason.String != string(domain.HoldMaterializing) {
		t.Fatalf("generation hold = %v, want materializing", gen.holdReason)
	}
	if got := documentStatus(t, database, fx.documentID); got != "approved" {
		t.Fatalf("document status = %q, want approved (still holding)", got)
	}

	// PDF lands: the completing write enqueues the evaluation itself.
	beforeEnqueues := fx.enqueuer.count()
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalPDF)
	if got := fx.enqueuer.count() - beforeEnqueues; got != 1 {
		t.Fatalf("completing artifact write enqueued %d evaluations, want exactly 1", got)
	}

	gen = loadGeneration(t, database, fx.documentID)
	eval, err = fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key())
	if err != nil {
		t.Fatalf("Evaluate (full set): %v", err)
	}
	if eval.Action != domain.ReleaseActionRelease {
		t.Fatalf("evaluation = %+v, want release", eval)
	}
	gen = loadGeneration(t, database, fx.documentID)
	if gen.holdReason.Valid {
		t.Fatalf("released generation still holds %q", gen.holdReason.String)
	}
}

// TestRelease_ReviewCycle_RecomputedFromActualReleaseTime pins ADR 0085's
// review-cycle rule to real rows: with no planned review date, the profile
// interval derives review_due_at from the ACTUAL release instant, which is what
// keeps the M6 periodic-review surfacer able to see the document (F-QA4-13).
func TestRelease_ReviewCycle_RecomputedFromActualReleaseTime(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedReleaseFixture(t, database, domain.StageKindApproval)
	fx.coordinator.WithProfileReviewIntervalReader(fixedReviewInterval{days: 30})

	fx.approveViaSignoff(t)
	gen := loadGeneration(t, database, fx.documentID)
	fx.recordArtifacts(t, gen.id, application.ArtifactFinalDocx, application.ArtifactFinalPDF)
	gen = loadGeneration(t, database, fx.documentID)

	if _, err := fx.coordinator.Evaluate(context.Background(), fx.runner, gen.key()); err != nil {
		t.Fatalf("Evaluate: %v", err)
	}

	doc := loadDocumentRelease(t, database, fx.documentID)
	if !doc.effective.Valid {
		t.Fatal("published document has no effective_from")
	}
	if !doc.reviewDue.Valid {
		t.Fatal("review_due_at not derived from the profile interval")
	}
	want := doc.effective.Time.AddDate(0, 0, 30)
	if diff := doc.reviewDue.Time.Sub(want); diff > time.Second || diff < -time.Second {
		t.Fatalf("review_due_at = %v, want ~%v (effective_from + 30d)", doc.reviewDue.Time, want)
	}
}
