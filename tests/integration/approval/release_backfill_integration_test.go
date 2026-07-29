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
//
// go test -tags=integration ./tests/integration/approval/... -run TestReleaseBackfill
package approval_test

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strings"
	"testing"

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

func (fx backfillFixture) evaluationJobCount(t *testing.T) int {
	t.Helper()
	return countRows(t, fx.database,
		`SELECT count(*) FROM river_job WHERE kind = 'release_evaluate' AND args->>'document_id' = $1`, fx.documentID)
}

func TestReleaseBackfill_LegacyApprovedDocument_RecordsFactAndArmsPipeline(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedBackfillFixture(t, database, true, releaseContentHash)
	ctx := context.Background()

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, false)
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

	first := backfill.RunDocument(ctx, fx.deps, fx.documentID, false)
	if first.Err != nil || first.Outcome != backfill.OutcomeBackfilled {
		t.Fatalf("first run: outcome=%s err=%v", first.Outcome, first.Err)
	}

	second := backfill.RunDocument(ctx, fx.deps, fx.documentID, false)
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

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, true)
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

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, false)
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

	res := backfill.RunDocument(ctx, fx.deps, fx.documentID, false)
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
