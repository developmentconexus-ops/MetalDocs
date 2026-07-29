//go:build integration
// +build integration

// ADR 0085 Stage C — the executable backfill for legacy approved documents.
// Drives the REAL per-document core (scripts/release-backfill/backfill) against
// a real Postgres instance, with a real River client and the real dispatch
// outbox enqueuer, and proves:
//
//   - a pinned, approved, legacy document gets its generation (correct 7-column
//     identity, approval fact stamped, final approver and submitter recorded),
//     its freeze lineage re-pinned, its materialize dispatch tagged with that
//     generation, and its evaluation enqueued — all in ONE transaction, with the
//     documents tripwire never firing and the approver's signed freeze
//     (status, values_hash, revision_version) never altered
//   - replay is a true no-op: same generation, no duplicate outbox row, no
//     duplicate evaluation job
//   - dry-run writes nothing at all
//   - preflight fails closed: an unpinned document and an instance without a
//     frozen_content_hash are both refused, and refusal leaves no partial state
//   - FreezeService.RepairMaterialization itself refuses a never-frozen
//     document, so the fail-closed rule holds even if a future caller skips
//     preflight
//   - repair is the sanctioned re-pin path for legacy NULL-pin documents
//     (migration 0313): it writes frozen_revision_id from the CURRENT revision
//     in the repair tx, supersedes a stale pin, and never touches values_hash
//   - -repair-only (Options.RepairOnly) is the operator-ratified invalid-artifact
//     expurgo: a PUBLISHED, frozen document is re-pinned and re-dispatched with
//     NO approval fact, NO release generation and NO evaluation; its dry-run
//     writes nothing; a never-frozen document, one with no current revision and
//     one with no approved approval instance are all refused; and the default
//     pass still rejects published documents in preflight
//   - the expurgo purges the STALE TERMINAL staging rows that would otherwise
//     swallow it — in BOTH metaldocs.materialize_dispatch_outbox and
//     metaldocs.pdf_dispatch_outbox — while an in-flight (pending/processing)
//     row on either side refuses the whole repair, and dry-run only reports the
//     purge plan
//   - it purges the DELIVERY layer under the same rule: terminal
//     (published/dead-lettered) metaldocs.outbox_events rows on the two dispatch
//     idempotency keys are removed so the next publish can land, an unpublished
//     (still claimable) event refuses the whole repair, and the purge is KEYED —
//     foreign documents and other key shapes are untouched
//
// NOT covered here, by design: the concurrent check-then-insert race between two
// simultaneous repair invocations. The integration harness cannot drive two
// competing transactions deterministically without heavy scaffolding, and the
// race is closed STRUCTURALLY instead — requireMaterializeEnqueued re-reads the
// dedupe key inside the same transaction after the enqueue, so a swallowed
// insert rolls the repair back rather than reporting success (see its doc
// comment in scripts/release-backfill/backfill/backfill.go).
//
// go test -tags=integration ./tests/integration/approval/... -run TestReleaseBackfill
package approval_test

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	docsapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/iam/authz"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/scripts/release-backfill/backfill"
	"metaldocs/tests/integration/testdb"
)

// backfillValuesHash is the 32-byte pin the fixture writes into
// documents.values_hash (documents_values_hash_len enforces exactly 32).
const backfillValuesHash = "abababababababababababababababababababababababababababababababab"

type backfillFixture struct {
	tenantID    string
	authorID    string
	approverID  string
	documentID  string
	revisionID  string
	instanceID  string
	stageID     string
	revisionVer int
	database    *sql.DB
	deps        backfill.Deps
}

// seedBackfillFixture builds exactly what the legacy path left behind: an
// APPROVED, PINNED document with an approved approval instance and a real
// approve signoff — and NO release generation, NO materialize dispatch, NO
// evaluation. That absence is the condition the backfill exists to repair.
//
// pinned controls whether values_hash / values_frozen_at are written; frozenHash
// controls the instance pin. Both are knobs so the preflight tests can seed a
// document that is legitimately unrepairable.
func seedBackfillFixture(t *testing.T, database *sql.DB, pinned bool, frozenHash string) backfillFixture {
	t.Helper()
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"))
	approver := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Approver"))

	doc := testdb.NewDocument(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("approved"),
		testdb.WithSubmitReadySnapshots(),
	)
	revisionID := seedReleaseRevision(t, database, doc.TenantID, doc.ID, author.ID)

	// current_revision_id is the generation identity's revision leg; values_hash
	// + values_frozen_at are the freeze pin RepairMaterialization requires.
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		var err error
		if pinned {
			_, err = tx.ExecContext(ctx, `
				UPDATE public.documents
				   SET current_revision_id = $1::uuid,
				       values_hash         = decode($2, 'hex'),
				       values_frozen_at    = now()
				 WHERE id = $3::uuid`,
				revisionID, backfillValuesHash, doc.ID)
		} else {
			_, err = tx.ExecContext(ctx, `
				UPDATE public.documents
				   SET current_revision_id = $1::uuid,
				       values_hash         = NULL,
				       values_frozen_at    = NULL
				 WHERE id = $2::uuid`,
				revisionID, doc.ID)
		}
		return err
	})

	route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))
	instance := testdb.NewApprovalInstance(t, database,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("approved"),
		testdb.WithFrozenContentHash(frozenHash),
	)

	var stageID string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
		VALUES ($1::uuid, 1, 'Legacy Stage', 'approver', 'document.signoff', 'qa',
		        'any_1_of', 'keep_snapshot', $2::jsonb, 'completed', 'approval')
		RETURNING id::text`,
		instance.ID, `["`+approver.ID+`"]`,
	).Scan(&stageID); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}

	// The approve signoff the backfill reads the final approver from.
	testdb.SeedWithCaps(t, database, `[{"cap":"document.signoff"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO public.approval_signoffs
			  (approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
			   actor_display_name_snapshot, decision, signed_at, signature_method,
			   signature_payload, content_hash)
			VALUES ($1::uuid, $2::uuid, $3, $4::uuid, 'Approver', 'approve', now(),
			        'password', '{}', $5)`,
			instance.ID, stageID, approver.ID, tenant.ID, releaseContentHash)
		return err
	})

	var revisionVer int
	if err := database.QueryRowContext(ctx,
		`SELECT revision_version FROM public.documents WHERE id = $1::uuid`, doc.ID,
	).Scan(&revisionVer); err != nil {
		t.Fatalf("read revision_version: %v", err)
	}

	// Real River client + real Wire: the test drives the production object
	// graph, not a hand-assembled stand-in, so the outbox and the job insert
	// are the ones the tool will actually perform.
	bundle, err := riverjobs.NewClientBundle(database, riverjobs.Config{SkipUnknownJobCheck: true}, nil)
	if err != nil {
		t.Fatalf("new river client bundle: %v", err)
	}
	deps, err := backfill.Wire(database, bundle.Client)
	if err != nil {
		t.Fatalf("wire backfill: %v", err)
	}

	return backfillFixture{
		tenantID:    tenant.ID,
		authorID:    author.ID,
		approverID:  approver.ID,
		documentID:  doc.ID,
		revisionID:  revisionID,
		instanceID:  instance.ID,
		stageID:     stageID,
		revisionVer: revisionVer,
		database:    database,
		deps:        deps,
	}
}

func countRows(t *testing.T, database *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := database.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	return n
}

func (fx backfillFixture) generationCount(t *testing.T) int {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM public.release_generations WHERE document_id = $1::uuid`, fx.documentID)
}

func (fx backfillFixture) outboxCount(t *testing.T) int {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM metaldocs.materialize_dispatch_outbox WHERE revision_id = $1::uuid`, fx.documentID)
}

func (fx backfillFixture) pdfOutboxCount(t *testing.T) int {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM metaldocs.pdf_dispatch_outbox WHERE revision_id = $1::uuid`, fx.documentID)
}

// seedStaleMaterializeDispatch plants the row shape the LIVE expurgo target
// carries: a generation-less materialize dispatch left behind by the render that
// produced the blank artifacts. status is a knob so the tests can pin both the
// terminal (purged) and the in-flight (refused) verdicts.
func (fx backfillFixture) seedStaleMaterializeDispatch(t *testing.T, status string) string {
	t.Helper()
	var id string
	if err := fx.database.QueryRowContext(context.Background(), `
		INSERT INTO metaldocs.materialize_dispatch_outbox
		  (tenant_id, revision_id, values_hash, release_generation_id, status, dispatched_at)
		VALUES ($1::uuid, $2::uuid, decode($3, 'hex'), NULL, $4,
		        CASE WHEN $4 = 'dispatched' THEN now() ELSE NULL END)
		RETURNING id::text`,
		fx.tenantID, fx.documentID, backfillValuesHash, status,
	).Scan(&id); err != nil {
		t.Fatalf("seed stale materialize dispatch (%s): %v", status, err)
	}
	return id
}

// seedStalePdfDispatch plants the SECOND stale row of the live target. It is the
// non-obvious half of the swallow: the worker's downstream EnqueuePDFTx runs on
// the same (tenant, revision, generation) key, so this row would leave a
// repaired frozen.docx paired with the original blank final.pdf.
func (fx backfillFixture) seedStalePdfDispatch(t *testing.T, status string) string {
	t.Helper()
	var id string
	if err := fx.database.QueryRowContext(context.Background(), `
		INSERT INTO metaldocs.pdf_dispatch_outbox
		  (tenant_id, revision_id, frozen_docx_hash, final_docx_s3_key, release_generation_id, status, dispatched_at)
		VALUES ($1::uuid, $2::uuid, decode($3, 'hex'), 'legacy/blank/frozen.docx', NULL, $4,
		        CASE WHEN $4 = 'dispatched' THEN now() ELSE NULL END)
		RETURNING id::text`,
		fx.tenantID, fx.documentID, backfillValuesHash, status,
	).Scan(&id); err != nil {
		t.Fatalf("seed stale pdf dispatch (%s): %v", status, err)
	}
	return id
}

// dispatchRowExists answers whether a specific staging row survived.
func (fx backfillFixture) dispatchRowExists(t *testing.T, table, id string) bool {
	t.Helper()
	var query string
	switch table {
	case "materialize":
		query = `SELECT count(*) FROM metaldocs.materialize_dispatch_outbox WHERE id = $1::uuid`
	case "pdf":
		query = `SELECT count(*) FROM metaldocs.pdf_dispatch_outbox WHERE id = $1::uuid`
	default:
		t.Fatalf("dispatchRowExists: unknown table %q", table)
	}
	return countRows(t, fx.database, query, id) == 1
}

// materializeDispatch returns the single materialize row's (id, status).
func (fx backfillFixture) materializeDispatch(t *testing.T) (string, string) {
	t.Helper()
	var id, status string
	if err := fx.database.QueryRowContext(context.Background(), `
		SELECT id::text, status
		  FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		fx.tenantID, fx.documentID,
	).Scan(&id, &status); err != nil {
		t.Fatalf("load materialize dispatch row: %v", err)
	}
	return id, status
}

// markDispatched moves a staging row to the terminal state the real consumer
// leaves behind (staging_outbox.go MarkDispatched).
func (fx backfillFixture) markDispatched(t *testing.T, id string) {
	t.Helper()
	if _, err := fx.database.ExecContext(context.Background(), `
		UPDATE metaldocs.materialize_dispatch_outbox
		   SET status = 'dispatched', dispatched_at = now()
		 WHERE id = $1::uuid`, id); err != nil {
		t.Fatalf("mark dispatched: %v", err)
	}
}

// deliveryKey rebuilds one of the two delivery idempotency keys the dispatch
// jobs publish on, in the generation-less shape a legacy document uses:
// "<prefix>:<tenant>:<document>" (dispatchjobs/workers.go:46-52 — the revision
// leg carries documents.id). Spelled out literally here ON PURPOSE: if the
// production key function drifts, this test must fail rather than follow it.
func (fx backfillFixture) deliveryKey(prefix string) string {
	return prefix + ":" + fx.tenantID + ":" + fx.documentID
}

// deliveryKeyForGeneration is the SAME key one release generation later: the
// generation id is appended, which is the whole point of ADR 0085's widening —
// two approvals of one revision are two different artifact sets and must not
// dedupe into each other (workers.go:38-52). A repair on a document that HAS a
// generation head must speak this shape, not the legacy one.
func (fx backfillFixture) deliveryKeyForGeneration(prefix, generationID string) string {
	return fx.deliveryKey(prefix) + ":" + generationID
}

// seedDeliveryEvent plants a metaldocs.outbox_events row holding one of this
// document's generation-less delivery keys — the layer-2 swallow.
func (fx backfillFixture) seedDeliveryEvent(t *testing.T, prefix, state string) string {
	t.Helper()
	return fx.seedDeliveryEventKeyed(t, fx.deliveryKey(prefix), prefix, state)
}

// seedDeliveryEventKeyed is seedDeliveryEvent with an explicit key, so tests can
// plant events on the generation-aware shape (or on a deliberately foreign one).
// state selects the row's disposition: "published" and "dead-lettered" are
// terminal (purgeable); "unpublished" is still deliverable (must refuse).
func (fx backfillFixture) seedDeliveryEventKeyed(t *testing.T, key, prefix, state string) string {
	t.Helper()
	eventID := uuid.NewString()
	var publishedAt, deadLetteredAt sql.NullTime
	switch state {
	case "published":
		publishedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	case "dead-lettered":
		deadLetteredAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	case "unpublished":
		// both NULL: the row the consumer can still deliver
		// (platform/messaging/outbox/postgres/consumer.go:41-46).
	default:
		t.Fatalf("seedDeliveryEvent: unknown state %q", state)
	}
	eventType := "docx_materialize"
	if prefix == "docgen_v2_pdf" {
		eventType = "docgen_v2_pdf"
	}
	payload := `{"tenant_id":"` + fx.tenantID + `","revision_id":"` + fx.documentID + `"}`
	if _, err := fx.database.ExecContext(context.Background(), `
		INSERT INTO metaldocs.outbox_events (
		  event_id, event_type, aggregate_type, aggregate_id, occurred_at, version,
		  idempotency_key, producer, trace_id, payload, published_at, dead_lettered_at
		)
		VALUES ($1, $2, 'document_revision', $3, now(), 1, $4, 'release-backfill-test',
		        'trace-'||$1, $5::jsonb, $6, $7)`,
		eventID, eventType, fx.documentID, key, payload, publishedAt, deadLetteredAt,
	); err != nil {
		t.Fatalf("seed delivery event (%s/%s): %v", key, state, err)
	}
	return eventID
}

func (fx backfillFixture) deliveryEventExists(t *testing.T, eventID string) bool {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM metaldocs.outbox_events WHERE event_id = $1`, eventID) == 1
}

// deliveryEventCount counts every event still holding one of this document's
// two delivery keys — the number that must be 0 for a re-dispatch to land.
func (fx backfillFixture) deliveryEventCount(t *testing.T) int {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM metaldocs.outbox_events WHERE idempotency_key IN ($1, $2)`,
		fx.deliveryKey("materialize_fanout"), fx.deliveryKey("docgen_v2_pdf"))
}

// frozenRevisionPin returns documents.frozen_revision_id as text, "" when NULL
// (the state every pre-0313 freeze is in).
func (fx backfillFixture) frozenRevisionPin(t *testing.T) string {
	t.Helper()
	var pin sql.NullString
	if err := fx.database.QueryRowContext(context.Background(),
		`SELECT frozen_revision_id::text FROM public.documents WHERE id = $1::uuid`, fx.documentID,
	).Scan(&pin); err != nil {
		t.Fatalf("read frozen_revision_id: %v", err)
	}
	return pin.String
}

// publish moves the seeded document into the legacy PUBLISHED shape: an actual
// effective_from is mandatory (ck_documents_published_effective_from, migration
// 0310). This is the state the repair-only expurgo exists for — a released
// document whose frozen artifacts are blank.
func (fx backfillFixture) publish(t *testing.T) {
	t.Helper()
	testdb.SeedWithCaps(t, fx.database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
			UPDATE public.documents
			   SET status         = 'published',
			       effective_from = now()
			 WHERE id = $1::uuid`, fx.documentID)
		return err
	})
}

// freezePin returns the (values_hash hex, values_frozen_at) pair the approver
// signed. Repair must carry both over byte-for-byte.
func (fx backfillFixture) freezePin(t *testing.T) (string, sql.NullTime) {
	t.Helper()
	var (
		valuesHash []byte
		frozenAt   sql.NullTime
	)
	if err := fx.database.QueryRowContext(context.Background(),
		`SELECT values_hash, values_frozen_at FROM public.documents WHERE id = $1::uuid`, fx.documentID,
	).Scan(&valuesHash, &frozenAt); err != nil {
		t.Fatalf("read freeze pin: %v", err)
	}
	return strings.ToLower(hex.EncodeToString(valuesHash)), frozenAt
}

// assertApprovalHistoryUntouched proves the repair changed nothing the approver
// signed or the approval kernel owns: no release generation was created, the
// instance keeps its status and its freeze pin, and the signoff still stands.
func (fx backfillFixture) assertApprovalHistoryUntouched(t *testing.T) {
	t.Helper()
	if n := fx.generationCount(t); n != 0 {
		t.Fatalf("repair-only created %d release_generations rows; it must record no approval fact", n)
	}
	var (
		status     string
		frozenHash sql.NullString
	)
	if err := fx.database.QueryRowContext(context.Background(),
		`SELECT status, frozen_content_hash FROM public.approval_instances WHERE id = $1::uuid`, fx.instanceID,
	).Scan(&status, &frozenHash); err != nil {
		t.Fatalf("reload approval instance: %v", err)
	}
	if status != "approved" || frozenHash.String != releaseContentHash {
		t.Fatalf("approval instance mutated: status=%s frozen_content_hash=%q", status, frozenHash.String)
	}
	if n := countRows(t, fx.database,
		`SELECT count(*) FROM public.approval_signoffs WHERE approval_instance_id = $1::uuid`, fx.instanceID); n != 1 {
		t.Fatalf("approval signoffs = %d, want the original 1", n)
	}
}

func (fx backfillFixture) evaluationJobCount(t *testing.T) int {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM river_job WHERE kind = 'release_evaluate' AND args->>'document_id' = $1`, fx.documentID)
}

func TestReleaseBackfill_LegacyApprovedDocument_RecordsFactAndArmsPipeline(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if res.Err != nil {
		t.Fatalf("backfill: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomeBackfilled {
		t.Fatalf("outcome = %s, want %s", res.Outcome, backfill.OutcomeBackfilled)
	}
	if res.GenerationID == "" {
		t.Fatal("backfill returned no generation id")
	}

	// ── the generation carries the exact 7-column identity + the approval fact
	var (
		tenantID, subjectKind, documentID, instanceID, revisionID, frozenHash string
		revisionVersion                                                       int
		approvalFactAt                                                        sql.NullTime
		finalApprover, submittedBy                                            sql.NullString
		artifactFactAt                                                        sql.NullTime
		releasedAt                                                            sql.NullTime
	)
	if err := database.QueryRowContext(ctx, `
		SELECT tenant_id::text, subject_kind, document_id::text, approval_instance_id::text,
		       revision_id::text, revision_version, frozen_content_hash,
		       approval_fact_at, final_approver_id, submitted_by, artifact_fact_at, released_at
		  FROM public.release_generations WHERE id = $1::uuid`, res.GenerationID,
	).Scan(&tenantID, &subjectKind, &documentID, &instanceID, &revisionID, &revisionVersion,
		&frozenHash, &approvalFactAt, &finalApprover, &submittedBy, &artifactFactAt, &releasedAt); err != nil {
		t.Fatalf("load generation: %v", err)
	}

	if tenantID != fx.tenantID || documentID != fx.documentID || instanceID != fx.instanceID || revisionID != fx.revisionID {
		t.Fatalf("generation identity mismatch: tenant=%s document=%s instance=%s revision=%s",
			tenantID, documentID, instanceID, revisionID)
	}
	if subjectKind != "document" {
		t.Fatalf("subject_kind = %q, want document", subjectKind)
	}
	if revisionVersion != fx.revisionVer {
		t.Fatalf("revision_version = %d, want %d", revisionVersion, fx.revisionVer)
	}
	if frozenHash != releaseContentHash {
		t.Fatalf("frozen_content_hash = %q, want the instance pin", frozenHash)
	}
	if !approvalFactAt.Valid {
		t.Fatal("approval_fact_at not stamped")
	}
	if finalApprover.String != fx.approverID {
		t.Fatalf("final_approver_id = %q, want %q", finalApprover.String, fx.approverID)
	}
	if submittedBy.String != fx.authorID {
		t.Fatalf("submitted_by = %q, want %q", submittedBy.String, fx.authorID)
	}
	// No legacy artifact was fast-forwarded: readiness must be earned by a real
	// render, so the generation is honestly incomplete and unreleased.
	if artifactFactAt.Valid || releasedAt.Valid {
		t.Fatalf("generation must not be artifact-ready or released yet: artifact=%v released=%v",
			artifactFactAt.Valid, releasedAt.Valid)
	}

	// ── materialization dispatched, tagged with THIS generation
	var outboxGen sql.NullString
	var outboxHash []byte
	if err := database.QueryRowContext(ctx, `
		SELECT release_generation_id::text, values_hash
		  FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		fx.tenantID, fx.documentID,
	).Scan(&outboxGen, &outboxHash); err != nil {
		t.Fatalf("load materialize outbox row: %v", err)
	}
	if outboxGen.String != res.GenerationID {
		t.Fatalf("outbox release_generation_id = %q, want %q", outboxGen.String, res.GenerationID)
	}
	if got := strings.ToLower(hex.EncodeToString(outboxHash)); got != backfillValuesHash {
		t.Fatalf("outbox values_hash = %s, want the document's pinned values_hash %s", got, backfillValuesHash)
	}

	// ── evaluation armed
	if n := fx.evaluationJobCount(t); n != 1 {
		t.Fatalf("release_evaluate jobs = %d, want 1", n)
	}

	// ── the freeze lineage is now pinned to the document's current revision:
	// the legacy row was frozen before migration 0313 and carried no pin, and
	// repair records the true one instead of leaving materialization dead.
	if pin := fx.frozenRevisionPin(t); pin != fx.revisionID {
		t.Fatalf("frozen_revision_id = %q, want the current revision %q", pin, fx.revisionID)
	}

	// ── nothing the approver signed was rewritten: the tripwire had nothing to
	// fire on, and status / values_hash / revision_version are untouched.
	var status string
	var valuesHash []byte
	var version int
	if err := database.QueryRowContext(ctx,
		`SELECT status, values_hash, revision_version FROM public.documents WHERE id = $1::uuid`, fx.documentID,
	).Scan(&status, &valuesHash, &version); err != nil {
		t.Fatalf("reload document: %v", err)
	}
	if status != "approved" || version != fx.revisionVer {
		t.Fatalf("document mutated: status=%s revision_version=%d", status, version)
	}
	if strings.ToLower(hex.EncodeToString(valuesHash)) != backfillValuesHash {
		t.Fatalf("values_hash changed: %s", hex.EncodeToString(valuesHash))
	}
}

func TestReleaseBackfill_Replay_IsIdempotent(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	first := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if first.Err != nil || first.Outcome != backfill.OutcomeBackfilled {
		t.Fatalf("first run: outcome=%s err=%v", first.Outcome, first.Err)
	}

	second := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if second.Err != nil {
		t.Fatalf("replay: unexpected error: %v", second.Err)
	}
	if second.Outcome != backfill.OutcomeAlreadyBackfilled {
		t.Fatalf("replay outcome = %s, want %s", second.Outcome, backfill.OutcomeAlreadyBackfilled)
	}
	if second.GenerationID != first.GenerationID {
		t.Fatalf("replay generation = %s, want %s", second.GenerationID, first.GenerationID)
	}

	if n := fx.generationCount(t); n != 1 {
		t.Fatalf("release_generations rows = %d, want 1", n)
	}
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want 1", n)
	}
	if n := fx.evaluationJobCount(t); n != 1 {
		t.Fatalf("release_evaluate jobs = %d, want 1", n)
	}
}

func TestReleaseBackfill_DryRun_WritesNothing(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{DryRun: true})
	if res.Err != nil {
		t.Fatalf("dry-run: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomePlanned {
		t.Fatalf("dry-run outcome = %s, want %s", res.Outcome, backfill.OutcomePlanned)
	}
	if res.Detail == "" {
		t.Fatal("dry-run reported no plan")
	}

	if n := fx.generationCount(t); n != 0 {
		t.Fatalf("dry-run wrote %d release_generations rows", n)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("dry-run wrote %d materialize outbox rows", n)
	}
	if n := fx.evaluationJobCount(t); n != 0 {
		t.Fatalf("dry-run enqueued %d release_evaluate jobs", n)
	}
}

func TestReleaseBackfill_UnpinnedDocument_Aborts(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, false, releaseContentHash)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
		t.Fatalf("unpinned document must abort, got outcome=%s err=%v", res.Outcome, res.Err)
	}
	if !strings.Contains(res.Err.Error(), "not pinned") {
		t.Fatalf("error should name the pin: %v", res.Err)
	}
	if n := fx.generationCount(t); n != 0 {
		t.Fatalf("aborted run wrote %d generations", n)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("aborted run wrote %d outbox rows", n)
	}
}

func TestReleaseBackfill_InstanceWithoutFrozenHash_Aborts(t *testing.T) {
	database, _ := testdb.Open(t)
	// WithFrozenContentHash("") forces the column NULL — a legacy instance that
	// never crossed the freeze boundary.
	fx := seedBackfillFixture(t, database, true, "")
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
		t.Fatalf("instance without frozen hash must abort, got outcome=%s err=%v", res.Outcome, res.Err)
	}
	if !strings.Contains(res.Err.Error(), "frozen_content_hash") {
		t.Fatalf("error should name the missing pin: %v", res.Err)
	}
	if n := fx.generationCount(t); n != 0 {
		t.Fatalf("aborted run wrote %d generations", n)
	}
}

// runRepair drives FreezeService.RepairMaterialization exactly as the tool
// does — same runner, same tenant seed, same system bypass — so the tests
// exercise the production path and not a hand-rolled transaction.
func (fx backfillFixture) runRepair(t *testing.T, generationID string) error {
	t.Helper()
	ctx := authz.WithBackgroundBypass(context.Background())
	return fx.deps.Runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(ctx, tx, fx.tenantID); err != nil {
			return err
		}
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return err
		}
		return fx.deps.Freeze.RepairMaterialization(ctx, tx, fx.tenantID, fx.documentID, generationID)
	})
}

// TestReleaseBackfill_RepairMaterialization_RefusesNeverFrozen pins the
// fail-closed rule to the freeze service itself, not merely to the tool's
// preflight: a future caller that skips preflight still cannot dispatch a render
// for a document that has no approved snapshot. Repair re-pins lineage, but it
// never invents the freeze itself.
func TestReleaseBackfill_RepairMaterialization_RefusesNeverFrozen(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, false, releaseContentHash)

	err := fx.runRepair(t, "")
	if err == nil {
		t.Fatal("RepairMaterialization accepted a never-frozen document")
	}
	if !strings.Contains(err.Error(), "was never frozen") {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("refused repair still wrote %d outbox rows", n)
	}
	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("refused repair pinned %q", pin)
	}

	// Guard the type assertion the tool depends on: Deps.Freeze is the real
	// FreezeService, not a stand-in.
	var _ *docsapp.FreezeService = fx.deps.Freeze
}

// TestReleaseBackfill_RepairMaterialization_PinsLegacyNullPinDocument is the
// §5.3 unblock: a document frozen BEFORE migration 0313 carries values_hash but
// no frozen_revision_id, so Materialize fails closed on it forever. Repair —
// the operator-only path — takes a FRESH pin from the current revision in the
// repair transaction and only then enqueues, which is what makes those legacy
// documents renderable again without fabricating history.
func TestReleaseBackfill_RepairMaterialization_PinsLegacyNullPinDocument(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)

	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("fixture must start unpinned (pre-0313 legacy shape), got %q", pin)
	}

	if err := fx.runRepair(t, ""); err != nil {
		t.Fatalf("repair of a legacy NULL-pin document failed: %v", err)
	}

	if pin := fx.frozenRevisionPin(t); pin != fx.revisionID {
		t.Fatalf("frozen_revision_id = %q, want the current revision %q", pin, fx.revisionID)
	}
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want 1", n)
	}

	// The signed freeze is untouched: repair re-pins lineage, it does not
	// recompute the hash the approver put their name on.
	var valuesHash []byte
	if err := database.QueryRowContext(context.Background(),
		`SELECT values_hash FROM public.documents WHERE id = $1::uuid`, fx.documentID,
	).Scan(&valuesHash); err != nil {
		t.Fatalf("reload values_hash: %v", err)
	}
	if strings.ToLower(hex.EncodeToString(valuesHash)) != backfillValuesHash {
		t.Fatalf("repair rewrote values_hash: %s", hex.EncodeToString(valuesHash))
	}
}

// TestReleaseBackfill_RepairMaterialization_SupersedesStalePin covers the other
// half of the ruling: choosing to repair IS the decision to re-freeze, so a
// document that already carries a pin gets re-pinned to the current revision
// rather than having the old pin defended.
func TestReleaseBackfill_RepairMaterialization_SupersedesStalePin(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	// A second revision that is NOT the head — the stale pin. It reuses the
	// fixture's editor session (idx_one_active_session_per_doc allows only one
	// active session per document) and needs its own content_hash
	// (document_revisions_document_id_content_hash_key).
	var stale string
	if err := database.QueryRowContext(ctx, `
		INSERT INTO public.document_revisions
			(document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		SELECT document_id, id, session_id, '', $2, '{}'
		  FROM public.document_revisions WHERE id = $1::uuid
		RETURNING id::text`,
		fx.revisionID, strings.Repeat("b", 64),
	).Scan(&stale); err != nil {
		t.Fatalf("seed stale revision: %v", err)
	}
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET frozen_revision_id = $1::uuid WHERE id = $2::uuid`,
			stale, fx.documentID)
		return err
	})

	if err := fx.runRepair(t, ""); err != nil {
		t.Fatalf("repair of an already-pinned document failed: %v", err)
	}

	pin := fx.frozenRevisionPin(t)
	if pin == stale {
		t.Fatal("repair kept the stale pin instead of re-pinning")
	}
	if pin != fx.revisionID {
		t.Fatalf("frozen_revision_id = %q, want the current revision %q", pin, fx.revisionID)
	}
}

// ── repair-only mode (operator-ratified invalid-artifact expurgo) ────────────
//
// A document frozen BEFORE the frozen-body fix carries blank frozen.docx /
// final.pdf artifacts. Artifact keys are deterministic per revision id, so
// re-materializing from the SAME current revision with a fresh lineage pin
// replaces those bytes in place. This is integrity restoration of invalid
// artifacts, NOT history mutation: no approval fact is recorded, no release
// generation is created, and the values_hash / values_frozen_at the approver
// signed are carried over unchanged.

// TestReleaseBackfill_RepairOnly_PublishedDocument_RematerializesInPlace is the
// ratified case: a PUBLISHED, frozen document — which the default pass refuses
// outright — is re-pinned and re-dispatched, and nothing the approval kernel
// owns moves.
func TestReleaseBackfill_RepairOnly_PublishedDocument_RematerializesInPlace(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	hashBefore, frozenAtBefore := fx.freezePin(t)

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Err != nil {
		t.Fatalf("repair-only: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("outcome = %s, want %s", res.Outcome, backfill.OutcomeRepaired)
	}

	// ── the freeze lineage is pinned to the document's current revision, which
	// is what makes the fresh render reproduce the approved body.
	if pin := fx.frozenRevisionPin(t); pin != fx.revisionID {
		t.Fatalf("frozen_revision_id = %q, want the current revision %q", pin, fx.revisionID)
	}

	// ── the signed freeze is byte-for-byte what it was.
	hashAfter, frozenAtAfter := fx.freezePin(t)
	if hashAfter != hashBefore || hashAfter != backfillValuesHash {
		t.Fatalf("values_hash changed: %s -> %s", hashBefore, hashAfter)
	}
	if !frozenAtAfter.Valid || !frozenAtAfter.Time.Equal(frozenAtBefore.Time) {
		t.Fatalf("values_frozen_at changed: %v -> %v", frozenAtBefore, frozenAtAfter)
	}

	// ── the document itself was not re-published, demoted or re-versioned.
	var status string
	var version int
	if err := database.QueryRowContext(ctx,
		`SELECT status, revision_version FROM public.documents WHERE id = $1::uuid`, fx.documentID,
	).Scan(&status, &version); err != nil {
		t.Fatalf("reload document: %v", err)
	}
	if status != "published" || version != fx.revisionVer {
		t.Fatalf("document mutated: status=%s revision_version=%d", status, version)
	}

	// ── materialization dispatched, carrying the signed values_hash and NO
	// generation (this legacy document has none — repair never invents one).
	var outboxGen sql.NullString
	var outboxHash []byte
	if err := database.QueryRowContext(ctx, `
		SELECT release_generation_id::text, values_hash
		  FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		fx.tenantID, fx.documentID,
	).Scan(&outboxGen, &outboxHash); err != nil {
		t.Fatalf("load materialize outbox row: %v", err)
	}
	if outboxGen.Valid {
		t.Fatalf("outbox release_generation_id = %q, want NULL for a generation-less legacy document", outboxGen.String)
	}
	if got := strings.ToLower(hex.EncodeToString(outboxHash)); got != backfillValuesHash {
		t.Fatalf("outbox values_hash = %s, want the document's pinned values_hash %s", got, backfillValuesHash)
	}

	// ── approval history untouched, and no release evaluation armed: a
	// published document evaluates to a no-op, so the tool does not enqueue one
	// (see runRepairOnly's ruling).
	fx.assertApprovalHistoryUntouched(t)
	if n := fx.evaluationJobCount(t); n != 0 {
		t.Fatalf("repair-only enqueued %d release_evaluate jobs, want 0", n)
	}
}

// TestReleaseBackfill_RepairOnly_DryRun_WritesNothing pins the dry-run
// guarantee to the repair path too: the plan is printed, the pin is not moved
// and no dispatch is enqueued.
func TestReleaseBackfill_RepairOnly_DryRun_WritesNothing(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true, DryRun: true})
	if res.Err != nil {
		t.Fatalf("repair-only dry-run: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomePlanned {
		t.Fatalf("dry-run outcome = %s, want %s", res.Outcome, backfill.OutcomePlanned)
	}
	if res.Detail == "" {
		t.Fatal("dry-run reported no plan")
	}

	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("dry-run wrote frozen_revision_id = %q", pin)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("dry-run wrote %d materialize outbox rows", n)
	}
	if n := fx.generationCount(t); n != 0 {
		t.Fatalf("dry-run wrote %d release_generations rows", n)
	}
}

// TestReleaseBackfill_RepairOnly_NeverFrozenDocument_Aborts is the fail-closed
// half: repair re-pins lineage, it never invents a freeze, so a document with
// no values_hash / values_frozen_at is refused before anything is written.
func TestReleaseBackfill_RepairOnly_NeverFrozenDocument_Aborts(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, false, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
		t.Fatalf("never-frozen document must abort, got outcome=%s err=%v", res.Outcome, res.Err)
	}
	if !strings.Contains(res.Err.Error(), "not frozen") {
		t.Fatalf("error should name the freeze: %v", res.Err)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("aborted repair wrote %d outbox rows", n)
	}
	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("aborted repair pinned %q", pin)
	}
}

// TestReleaseBackfill_PublishedDocument_DefaultModeStillRejects guards the
// unchanged behaviour of the default pass: -repair-only is the ONLY way a
// published document is accepted, so the Stage C backfill can never
// accidentally re-record an approval fact for one.
func TestReleaseBackfill_PublishedDocument_DefaultModeStillRejects(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
		t.Fatalf("default mode must reject a published document, got outcome=%s err=%v", res.Outcome, res.Err)
	}
	if !strings.Contains(res.Err.Error(), "is published, not approved") {
		t.Fatalf("error should name the status: %v", res.Err)
	}
	if n := fx.generationCount(t); n != 0 {
		t.Fatalf("rejected run wrote %d generations", n)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("rejected run wrote %d outbox rows", n)
	}
}

// TestReleaseBackfill_RepairOnly_StaleTerminalDispatchRows_PurgedAndRedispatched
// is the LIVE target's shape: the document still carries both stale dispatched
// staging rows from the render that produced the blank artifacts. Both outboxes
// dedupe on (tenant_id, revision_id, COALESCE(release_generation_id, nil-uuid))
// with ON CONFLICT DO NOTHING (migration 0310 §5), and a skipped insert is
// reported as a nil error — so leaving either row in place makes the repair a
// successful-looking no-op (materialize side) or a half-repair with a stale
// blank final.pdf (pdf side). The purge is the expurgo.
func TestReleaseBackfill_RepairOnly_StaleTerminalDispatchRows_PurgedAndRedispatched(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	staleMaterialize := fx.seedStaleMaterializeDispatch(t, "dispatched")
	stalePDF := fx.seedStalePdfDispatch(t, "dispatched")

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Err != nil {
		t.Fatalf("repair-only over stale rows: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("outcome = %s, want %s", res.Outcome, backfill.OutcomeRepaired)
	}

	// ── both stale rows are gone, and the pdf side is left EMPTY so the worker's
	// own downstream EnqueuePDFTx can land.
	if fx.dispatchRowExists(t, "materialize", staleMaterialize) {
		t.Fatalf("stale materialize row %s survived the purge", staleMaterialize)
	}
	if fx.dispatchRowExists(t, "pdf", stalePDF) {
		t.Fatalf("stale pdf row %s survived the purge", stalePDF)
	}
	if n := fx.pdfOutboxCount(t); n != 0 {
		t.Fatalf("pdf outbox rows = %d, want 0 (repair enqueues no pdf; the worker does)", n)
	}

	// ── exactly one FRESH pending materialize row replaced it.
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want the single fresh row", n)
	}
	freshID, freshStatus := fx.materializeDispatch(t)
	if freshID == staleMaterialize {
		t.Fatal("materialize row was not replaced: the repair reused the stale row's id")
	}
	if freshStatus != "pending" {
		t.Fatalf("fresh materialize row status = %q, want pending", freshStatus)
	}

	// ── the purge is operator-auditable: the outcome line names both ids.
	if !strings.Contains(res.Detail, staleMaterialize) || !strings.Contains(res.Detail, stalePDF) {
		t.Fatalf("outcome detail must report the purged row ids, got %q", res.Detail)
	}

	// ── and none of this touched what the approver signed.
	if pin := fx.frozenRevisionPin(t); pin != fx.revisionID {
		t.Fatalf("frozen_revision_id = %q, want the current revision %q", pin, fx.revisionID)
	}
	fx.assertApprovalHistoryUntouched(t)
}

// TestReleaseBackfill_RepairOnly_InFlightDispatchRow_Refuses is the fail-closed
// half of the purge: a pending/processing row is live work whose River job may
// still fire, so deleting it would orphan the job and repairing around it would
// be swallowed. The tool refuses the whole repair and names the row.
func TestReleaseBackfill_RepairOnly_InFlightDispatchRow_Refuses(t *testing.T) {
	cases := []struct {
		name   string
		table  string
		status string
	}{
		{name: "materialize_pending", table: "materialize", status: "pending"},
		{name: "materialize_processing", table: "materialize", status: "processing"},
		{name: "pdf_pending", table: "pdf", status: "pending"},
		{name: "pdf_processing", table: "pdf", status: "processing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			database, _ := testdb.Open(t)
			fx := seedBackfillFixture(t, database, true, releaseContentHash)
			fx.publish(t)
			ctx := context.Background()

			var rowID string
			if tc.table == "materialize" {
				rowID = fx.seedStaleMaterializeDispatch(t, tc.status)
			} else {
				rowID = fx.seedStalePdfDispatch(t, tc.status)
			}

			res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
			if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
				t.Fatalf("in-flight %s row must abort the repair, got outcome=%s err=%v", tc.table, res.Outcome, res.Err)
			}
			if !strings.Contains(res.Err.Error(), "IN-FLIGHT") || !strings.Contains(res.Err.Error(), rowID) {
				t.Fatalf("error should name the in-flight row %s: %v", rowID, res.Err)
			}

			// The refusal is total: the row stands and nothing was written.
			if !fx.dispatchRowExists(t, tc.table, rowID) {
				t.Fatalf("refused repair deleted the in-flight %s row %s", tc.table, rowID)
			}
			if pin := fx.frozenRevisionPin(t); pin != "" {
				t.Fatalf("refused repair pinned %q", pin)
			}
			if tc.table == "pdf" && fx.outboxCount(t) != 0 {
				t.Fatal("refused repair enqueued a materialize row anyway")
			}
		})
	}
}

// TestReleaseBackfill_RepairOnly_DryRun_ReportsPurgePlan proves the dry-run is
// honest about the destructive half too: it names the rows it WOULD purge and
// deletes none of them.
func TestReleaseBackfill_RepairOnly_DryRun_ReportsPurgePlan(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	staleMaterialize := fx.seedStaleMaterializeDispatch(t, "dispatched")
	stalePDF := fx.seedStalePdfDispatch(t, "failed")

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true, DryRun: true})
	if res.Err != nil {
		t.Fatalf("repair-only dry-run: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomePlanned {
		t.Fatalf("outcome = %s, want %s", res.Outcome, backfill.OutcomePlanned)
	}
	if !strings.Contains(res.Detail, "would purge") ||
		!strings.Contains(res.Detail, staleMaterialize) ||
		!strings.Contains(res.Detail, stalePDF) {
		t.Fatalf("dry-run plan must name both rows it would purge, got %q", res.Detail)
	}

	if !fx.dispatchRowExists(t, "materialize", staleMaterialize) || !fx.dispatchRowExists(t, "pdf", stalePDF) {
		t.Fatal("dry-run deleted a staging row")
	}
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want the untouched stale row only", n)
	}
	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("dry-run wrote frozen_revision_id = %q", pin)
	}
}

// TestReleaseBackfill_RepairOnly_Replay pins the honest replay semantics, which
// follow from the purge rule rather than being chosen for convenience:
//
//   - immediately after a repair the fresh row is PENDING — live work — so a
//     second repair refuses rather than orphaning its River job;
//   - once that dispatch is terminal, re-repairing is purge-and-redo: repeated
//     integrity restoration of the same document is idempotent by construction
//     (same revision, same values_hash, deterministic artifact keys), so there
//     is nothing to protect against.
//
// The replay leg carries BOTH tables, not just the materialize one: by the time
// a repair is replayed the worker has normally landed its own downstream pdf
// dispatch, so purge-and-redo has to clear the pdf row too or the second repair
// re-creates the half-repair the purge exists to prevent.
func TestReleaseBackfill_RepairOnly_Replay(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	first := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if first.Err != nil || first.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("first repair: outcome=%s err=%v", first.Outcome, first.Err)
	}
	firstRow, firstStatus := fx.materializeDispatch(t)
	if firstStatus != "pending" {
		t.Fatalf("first repair left status %q, want pending", firstStatus)
	}

	// ── replay while the dispatch is still in flight: refused.
	second := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if second.Outcome != backfill.OutcomeFailed || second.Err == nil {
		t.Fatalf("replay over a pending dispatch must abort, got outcome=%s err=%v", second.Outcome, second.Err)
	}
	if !strings.Contains(second.Err.Error(), "IN-FLIGHT") {
		t.Fatalf("error should name the in-flight dispatch: %v", second.Err)
	}
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want the single row from the first repair", n)
	}

	// ── the worker consumed that dispatch and landed its own pdf dispatch, which
	// then dispatched too: the real state a replayed repair meets.
	fx.markDispatched(t, firstRow)
	workerPDFRow := fx.seedStalePdfDispatch(t, "dispatched")

	// ── re-repair purges BOTH terminal rows and redoes the materialization.
	third := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if third.Err != nil || third.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("re-repair over dispatched rows: outcome=%s err=%v", third.Outcome, third.Err)
	}
	if !strings.Contains(third.Detail, firstRow) || !strings.Contains(third.Detail, workerPDFRow) {
		t.Fatalf("re-repair must report both purged rows (%s, %s), got %q", firstRow, workerPDFRow, third.Detail)
	}
	if fx.dispatchRowExists(t, "pdf", workerPDFRow) {
		t.Fatalf("re-repair left the worker's pdf row %s in place; the next pdf enqueue would be swallowed", workerPDFRow)
	}
	if n := fx.pdfOutboxCount(t); n != 0 {
		t.Fatalf("pdf outbox rows = %d after purge-and-redo, want 0", n)
	}
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want exactly one after purge-and-redo", n)
	}
	thirdRow, thirdStatus := fx.materializeDispatch(t)
	if thirdRow == firstRow || thirdStatus != "pending" {
		t.Fatalf("re-repair did not replace the dispatch: id=%s status=%s", thirdRow, thirdStatus)
	}
	fx.assertApprovalHistoryUntouched(t)
}

// TestReleaseBackfill_RepairOnly_NoApprovedInstance_Refuses proves "post-approval"
// is proved by the DURABLE approval record, not by the mutable status column: a
// document sitting in a post-approval status with no approved approval_instance
// is refused, because repairing its artifacts would render content nobody signed.
func TestReleaseBackfill_RepairOnly_NoApprovedInstance_Refuses(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	if _, err := database.ExecContext(ctx,
		`UPDATE public.approval_instances SET status = 'cancelled' WHERE id = $1::uuid`, fx.instanceID,
	); err != nil {
		t.Fatalf("cancel approval instance: %v", err)
	}

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
		t.Fatalf("document without an approved instance must abort, got outcome=%s err=%v", res.Outcome, res.Err)
	}
	if !strings.Contains(res.Err.Error(), "never completed approval") {
		t.Fatalf("error should name the missing approval: %v", res.Err)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("aborted repair wrote %d outbox rows", n)
	}
	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("aborted repair pinned %q", pin)
	}
}

// TestReleaseBackfill_RepairOnly_NoCurrentRevision_Refuses: the repair re-pins
// from the CURRENT revision, so a document without one has nothing to pin. It is
// refused in preflight rather than reaching RepairMaterialization.
func TestReleaseBackfill_RepairOnly_NoCurrentRevision_Refuses(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET current_revision_id = NULL WHERE id = $1::uuid`, fx.documentID)
		return err
	})

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
		t.Fatalf("document without a current revision must abort, got outcome=%s err=%v", res.Outcome, res.Err)
	}
	if !strings.Contains(res.Err.Error(), "no current_revision_id") {
		t.Fatalf("error should name the missing revision: %v", res.Err)
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("aborted repair wrote %d outbox rows", n)
	}
	if pin := fx.frozenRevisionPin(t); pin != "" {
		t.Fatalf("aborted repair pinned %q", pin)
	}
}

// ── layer 2: the DELIVERY ledger ────────────────────────────────────────────
//
// metaldocs.outbox_events is UNIQUE on idempotency_key and the publisher is
// ON CONFLICT DO NOTHING returning nil (publisher.go:35, :43-57), so a leftover
// event on one of the repair's two keys makes the River dispatch job publish
// NOTHING and still succeed — it marks the staging row dispatched and the
// operator sees a completed job with no work behind it. Not a theory: it is what
// a live run of this tool did before the delivery purge existed.

// TestReleaseBackfill_RepairOnly_StaleDeliveryEvents_Purged mirrors the LIVE
// state of the expurgo target: a dispatched staging row plus both stale delivery
// events still holding the legacy generation-less keys. One published, one
// dead-lettered — both terminal, both purgeable.
func TestReleaseBackfill_RepairOnly_StaleDeliveryEvents_Purged(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	staleStaging := fx.seedStaleMaterializeDispatch(t, "dispatched")
	materializeEvent := fx.seedDeliveryEvent(t, "materialize_fanout", "published")
	pdfEvent := fx.seedDeliveryEvent(t, "docgen_v2_pdf", "dead-lettered")

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Err != nil {
		t.Fatalf("repair-only over stale delivery events: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("outcome = %s, want %s", res.Outcome, backfill.OutcomeRepaired)
	}

	// ── both delivery keys are free, so the next publish can actually land.
	if fx.deliveryEventExists(t, materializeEvent) {
		t.Fatalf("published materialize event %s survived; the re-dispatch would publish nothing", materializeEvent)
	}
	if fx.deliveryEventExists(t, pdfEvent) {
		t.Fatalf("dead-lettered pdf event %s survived; the pdf leg would publish nothing", pdfEvent)
	}
	if n := fx.deliveryEventCount(t); n != 0 {
		t.Fatalf("events still holding this document's delivery keys = %d, want 0", n)
	}

	// ── the staging layer behaved exactly as before: purged and re-enqueued.
	if fx.dispatchRowExists(t, "materialize", staleStaging) {
		t.Fatalf("stale staging row %s survived the purge", staleStaging)
	}
	if n := fx.outboxCount(t); n != 1 {
		t.Fatalf("materialize outbox rows = %d, want the single fresh row", n)
	}
	freshID, freshStatus := fx.materializeDispatch(t)
	if freshID == staleStaging || freshStatus != "pending" {
		t.Fatalf("staging row not replaced: id=%s status=%s", freshID, freshStatus)
	}

	// ── the purge is auditable: every purged id is on the outcome line.
	for _, id := range []string{staleStaging, materializeEvent, pdfEvent} {
		if !strings.Contains(res.Detail, id) {
			t.Fatalf("outcome detail must report purged id %s, got %q", id, res.Detail)
		}
	}

	fx.assertApprovalHistoryUntouched(t)
}

// TestReleaseBackfill_RepairOnly_InFlightDeliveryEvent_Refuses is the fail-closed
// half of the delivery purge: an unpublished, non-dead-lettered event is still
// claimable by the outbox consumer (consumer.go:41-44), so deleting it would
// drop a live message. The repair refuses entirely instead.
func TestReleaseBackfill_RepairOnly_InFlightDeliveryEvent_Refuses(t *testing.T) {
	for _, prefix := range []string{"materialize_fanout", "docgen_v2_pdf"} {
		t.Run(prefix, func(t *testing.T) {
			database, _ := testdb.Open(t)
			fx := seedBackfillFixture(t, database, true, releaseContentHash)
			fx.publish(t)
			ctx := context.Background()

			liveEvent := fx.seedDeliveryEvent(t, prefix, "unpublished")

			res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
			if res.Outcome != backfill.OutcomeFailed || res.Err == nil {
				t.Fatalf("in-flight delivery event must abort the repair, got outcome=%s err=%v", res.Outcome, res.Err)
			}
			if !strings.Contains(res.Err.Error(), "IN-FLIGHT delivery event") ||
				!strings.Contains(res.Err.Error(), liveEvent) ||
				!strings.Contains(res.Err.Error(), fx.deliveryKey(prefix)) {
				t.Fatalf("error should name the live event and its key: %v", res.Err)
			}

			// Total refusal: the event stands and nothing was written.
			if !fx.deliveryEventExists(t, liveEvent) {
				t.Fatalf("refused repair deleted the in-flight event %s", liveEvent)
			}
			if n := fx.outboxCount(t); n != 0 {
				t.Fatalf("refused repair wrote %d staging rows", n)
			}
			if pin := fx.frozenRevisionPin(t); pin != "" {
				t.Fatalf("refused repair pinned %q", pin)
			}
		})
	}
}

// TestReleaseBackfill_RepairOnly_DryRun_ReportsDeliveryEventPurgePlan proves the
// dry-run is honest about the delivery layer too.
func TestReleaseBackfill_RepairOnly_DryRun_ReportsDeliveryEventPurgePlan(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	materializeEvent := fx.seedDeliveryEvent(t, "materialize_fanout", "published")
	pdfEvent := fx.seedDeliveryEvent(t, "docgen_v2_pdf", "published")

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true, DryRun: true})
	if res.Err != nil {
		t.Fatalf("repair-only dry-run: unexpected error: %v", res.Err)
	}
	if res.Outcome != backfill.OutcomePlanned {
		t.Fatalf("outcome = %s, want %s", res.Outcome, backfill.OutcomePlanned)
	}
	if !strings.Contains(res.Detail, "would purge") ||
		!strings.Contains(res.Detail, materializeEvent) ||
		!strings.Contains(res.Detail, pdfEvent) {
		t.Fatalf("dry-run plan must name both delivery events it would purge, got %q", res.Detail)
	}

	if !fx.deliveryEventExists(t, materializeEvent) || !fx.deliveryEventExists(t, pdfEvent) {
		t.Fatal("dry-run deleted a delivery event")
	}
	if n := fx.outboxCount(t); n != 0 {
		t.Fatalf("dry-run wrote %d staging rows", n)
	}
}

// TestReleaseBackfill_RepairOnly_GenerationAwareDeliveryKeys_Purged is the key-
// PARITY test: a document that HAS a release generation head must have its
// delivery layer cleared on the GENERATION-SUFFIXED keys, because that is what
// its dispatch jobs will publish on (dispatchjobs/workers.go:46-52). Repair
// computes those keys itself; if it ever computed the legacy generation-less
// shape instead, this document's real keys would stay held and the re-dispatch
// would publish nothing — the exact failure the delivery purge exists to end.
//
// The generation is created by the REAL Stage C backfill rather than hand-
// inserted, so the identity under test is the production one.
func TestReleaseBackfill_RepairOnly_GenerationAwareDeliveryKeys_Purged(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	// ── give the document a real generation head + its generation-tagged staging
	// row, then let the worker "consume" that dispatch.
	first := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{})
	if first.Err != nil || first.Outcome != backfill.OutcomeBackfilled {
		t.Fatalf("stage C backfill: outcome=%s err=%v", first.Outcome, first.Err)
	}
	generationID := first.GenerationID
	stagingRow, _ := fx.materializeDispatch(t)
	fx.markDispatched(t, stagingRow)

	// ── the two events the dispatch jobs published for THIS generation.
	genMaterializeEvent := fx.seedDeliveryEventKeyed(t,
		fx.deliveryKeyForGeneration("materialize_fanout", generationID), "materialize_fanout", "published")
	genPDFEvent := fx.seedDeliveryEventKeyed(t,
		fx.deliveryKeyForGeneration("docgen_v2_pdf", generationID), "docgen_v2_pdf", "published")

	// ── and the legacy generation-LESS events for the same document, left over
	// from a pre-ADR-0085 render. They are NOT on this repair's keys, so a keyed
	// purge must leave them alone; deleting them would prove the tool is
	// sweeping by document instead of by key.
	legacyMaterializeEvent := fx.seedDeliveryEvent(t, "materialize_fanout", "published")
	legacyPDFEvent := fx.seedDeliveryEvent(t, "docgen_v2_pdf", "published")

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Err != nil || res.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("repair-only on a generation-aware document: outcome=%s err=%v", res.Outcome, res.Err)
	}
	if res.GenerationID != generationID {
		t.Fatalf("repair reported generation %q, want the head %q", res.GenerationID, generationID)
	}

	// ── the generation-suffixed keys are free.
	if fx.deliveryEventExists(t, genMaterializeEvent) {
		t.Fatalf("generation-aware materialize event %s survived; the re-dispatch would publish nothing", genMaterializeEvent)
	}
	if fx.deliveryEventExists(t, genPDFEvent) {
		t.Fatalf("generation-aware pdf event %s survived; the pdf leg would publish nothing", genPDFEvent)
	}
	for _, id := range []string{genMaterializeEvent, genPDFEvent} {
		if !strings.Contains(res.Detail, id) {
			t.Fatalf("outcome detail must report purged event %s, got %q", id, res.Detail)
		}
	}

	// ── the legacy keys are untouched: the purge is keyed, not swept.
	if !fx.deliveryEventExists(t, legacyMaterializeEvent) || !fx.deliveryEventExists(t, legacyPDFEvent) {
		t.Fatal("purge deleted the generation-LESS events; it is keying on the document, not on the generation-aware key")
	}

	// ── and the staging layer stayed generation-tagged through the repair.
	freshRow, freshStatus := fx.materializeDispatch(t)
	if freshRow == stagingRow || freshStatus != "pending" {
		t.Fatalf("staging row not replaced: id=%s status=%s", freshRow, freshStatus)
	}
	var outboxGen sql.NullString
	if err := database.QueryRowContext(ctx,
		`SELECT release_generation_id::text FROM metaldocs.materialize_dispatch_outbox WHERE id = $1::uuid`, freshRow,
	).Scan(&outboxGen); err != nil {
		t.Fatalf("load fresh staging row: %v", err)
	}
	if outboxGen.String != generationID {
		t.Fatalf("fresh staging row release_generation_id = %q, want the head %q", outboxGen.String, generationID)
	}

	// ── repair recorded no NEW approval fact: the generation count is still the
	// one the Stage C pass created.
	if n := fx.generationCount(t); n != 1 {
		t.Fatalf("release_generations rows = %d, want the single Stage C generation", n)
	}
}

// TestReleaseBackfill_RepairOnly_ForeignDeliveryEvents_Untouched pins the blast
// radius of the delivery purge: it is KEYED, not swept. An event belonging to
// another document — and an event on a different key SHAPE for this same
// document, such as a generation-aware one — are both left alone.
func TestReleaseBackfill_RepairOnly_ForeignDeliveryEvents_Untouched(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	fx.publish(t)
	ctx := context.Background()

	mine := fx.seedDeliveryEvent(t, "materialize_fanout", "published")

	otherDocument := uuid.NewString()
	foreignKeys := []string{
		"materialize_fanout:" + fx.tenantID + ":" + otherDocument,
		"materialize_fanout:" + fx.tenantID + ":" + fx.documentID + ":" + uuid.NewString(),
	}
	foreignIDs := make([]string, 0, len(foreignKeys))
	for _, key := range foreignKeys {
		eventID := uuid.NewString()
		if _, err := database.ExecContext(ctx, `
			INSERT INTO metaldocs.outbox_events (
			  event_id, event_type, aggregate_type, aggregate_id, occurred_at, version,
			  idempotency_key, producer, trace_id, payload, published_at
			)
			VALUES ($1, 'docx_materialize', 'document_revision', $2, now(), 1, $3,
			        'release-backfill-test', 'trace-'||$1, '{}'::jsonb, now())`,
			eventID, otherDocument, key,
		); err != nil {
			t.Fatalf("seed foreign delivery event: %v", err)
		}
		foreignIDs = append(foreignIDs, eventID)
	}

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, backfill.Options{RepairOnly: true})
	if res.Err != nil || res.Outcome != backfill.OutcomeRepaired {
		t.Fatalf("repair-only: outcome=%s err=%v", res.Outcome, res.Err)
	}

	if fx.deliveryEventExists(t, mine) {
		t.Fatalf("this document's stale event %s survived", mine)
	}
	for i, id := range foreignIDs {
		if !fx.deliveryEventExists(t, id) {
			t.Fatalf("purge deleted a foreign event on key %q (event %s)", foreignKeys[i], id)
		}
	}
}
