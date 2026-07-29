// Package backfill implements the ADR 0085 Stage C in-flight disposition: a
// one-shot, idempotent repair pass that gives a legacy APPROVED document the
// release prerequisites the release coordinator needs, without re-approving it
// and without trusting anything the legacy path left behind.
//
// Per document, in ONE transaction (production parity with
// approval/application.recordTerminalApprovalRelease):
//
//  1. tx-local security setup — SeedTxTenant + BypassSystem, exactly as the
//     release_evaluate background job does. No BYPASSRLS role, no superuser.
//  2. preflight, fail-closed — the document must be approved, controlled and
//     PINNED; there must be exactly one approved approval_instance carrying a
//     frozen_content_hash; that instance must have at least one approve
//     signoff. Any miss aborts THIS document (the caller continues to the
//     next); nothing partial is written.
//  3. RecordApprovalFactTx — synthesizes the missing approval fact.
//  4. FreezeService.RepairMaterialization — re-enqueues materialization
//     against the new generation.
//  5. EnqueueReleaseEvaluationTx — arms the coordinator.
//
// Deliberately absent: fastForwardExistingArtifacts. Legacy final_docx /
// final_pdf pointers are untrusted — column presence is never readiness
// (ADR 0085) — so no legacy artifact is promoted into the new generation.
// Fresh artifacts are rendered from the frozen values snapshot by the
// production worker, which lands both artifact facts and re-enqueues
// evaluation; until then the generation honestly holds on `materializing`.
//
// Replay is a no-op at every write: the generation identity upserts, the
// staging outbox dedupe is generation-aware, and an already-backfilled
// document is detected in preflight and skipped.
//
// # Repair-only mode
//
// Options.RepairOnly is the SECOND, operator-ratified pass: integrity
// restoration of INVALID artifacts, not history mutation. A document that was
// frozen before the frozen-body fix carries a blank frozen.docx / final.pdf;
// re-materializing it from the SAME current revision with a fresh lineage pin
// replaces those bytes in place (artifact keys are deterministic per revision
// id) while the approver's signed freeze — values_hash, values_frozen_at, the
// approval instance, its signoffs and any release generation — is left
// untouched. Only step 4 (RepairMaterialization) runs; steps 3 and 5 do not.
//
// # The swallow stack
//
// Re-dispatching a document that was already rendered once is not one dedupe
// hop but THREE, and each of them reports its skip as success. All three were
// found the hard way — the third only after a live run printed `repaired`,
// completed its River job, and rendered nothing. Per leg:
//
//  1. STAGING dedupe — metaldocs.{materialize,pdf}_dispatch_outbox unique on
//     (tenant_id, revision_id, COALESCE(release_generation_id, nil-uuid)),
//     NOT partial, so it covers every status (migration 0310 §5). The INSERT is
//     ON CONFLICT DO NOTHING and a skip returns an empty id with a nil error
//     (render/fanout/staging_outbox.go:93, :133), which the enqueuer reads as
//     "nothing to dispatch" — no River job at all
//     (render/fanout/dispatchjobs/enqueuer.go:110-113).
//  2. DELIVERY dedupe — metaldocs.outbox_events is UNIQUE on idempotency_key
//     (baseline 0001_current_schema.sql:2535-2536) and the publisher is also
//     ON CONFLICT DO NOTHING returning nil
//     (platform/messaging/outbox/postgres/publisher.go:35, :43-57). The key is
//     built by dispatchjobs.dispatchIdempotencyKey (workers.go:46-52), and for a
//     GENERATION-LESS render it is the historical revision-only shape —
//     "materialize_fanout:<tenant>:<document>" — deliberately preserved verbatim
//     by ADR 0085 so nothing already shipped changed. A stale published event on
//     that key makes the River job succeed while publishing nothing: run()
//     treats the nil error as a successful publish and marks the staging row
//     dispatched (workers.go:104-117).
//  3. The PDF leg repeats BOTH of the above. The worker's own downstream
//     EnqueuePDFTx runs on the same staging key
//     (platform/worker/materialize_job_runner.go:146), and its dispatch job
//     publishes on "docgen_v2_pdf:<tenant>:<document>" — so a repair that
//     clears only the materialize side lands a fresh frozen.docx next to the
//     original blank final.pdf.
//
// Repair-only clears layers 1 and 2 on both legs, and refuses when anything on
// either layer is still in flight.
//
// Literal write set of one repair transaction, stated at SQL level so nobody
// has to infer it:
//
//   - DELETE of the TERMINAL (dispatched/failed) staging rows that match this
//     repair's dedupe key, in metaldocs.materialize_dispatch_outbox AND
//     metaldocs.pdf_dispatch_outbox. Those rows are the ops provenance of the
//     invalid artifacts, not audit facts — purging them IS the expurgo. See
//     purgeStaleDispatch for why both tables and why terminal-only.
//   - DELETE of the TERMINAL (published or dead-lettered) metaldocs.outbox_events
//     rows on this repair's two delivery idempotency keys. Same expurgo, one
//     layer down: outbox_events is the DELIVERY ledger of the very dispatch the
//     staging purge already covers. It is NOT an audit table — audit facts live
//     in metaldocs.audit_events with their own hash chain, and this tool never
//     reads or writes them. See classifyStaleDeliveryEvents.
//   - UPDATE public.documents SET values_hash, values_frozen_at,
//     frozen_revision_id (snapshot_repository.go:166-175, one statement). Only
//     frozen_revision_id CHANGES; values_hash and values_frozen_at are
//     re-assigned byte- and timestamp-identical, read back out of the same row
//     earlier in the same tx (freeze_service.go:337-353). The invariant holds —
//     the approver's signed freeze survives the repair — but the UPDATE does
//     name those columns, so do not read this as "untouched at SQL level".
//   - INSERT of one pending metaldocs.materialize_dispatch_outbox row, plus the
//     paired River dispatch job the enqueuer inserts in the same tx when (and
//     only when) that outbox INSERT returns an id
//     (render/fanout/dispatchjobs/enqueuer.go:105-123).
//
// Nothing else is written: no approval fact, no release generation, no signoff,
// no status change, no artifact pointer. The fresh artifacts are produced later
// by the production worker.
package backfill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldom "metaldocs/internal/modules/approval/domain"
	docapp "metaldocs/internal/modules/documents/application"
	docdom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

// Options selects which of the two operator-ratified passes RunDocument
// performs and whether it is allowed to write.
type Options struct {
	// RepairOnly runs ONLY FreezeService.RepairMaterialization: a frozen,
	// post-approval document is re-pinned to its current revision and its
	// materialization re-dispatched, so the worker renders fresh artifacts over
	// the invalid ones. No approval fact is recorded and no release evaluation
	// is enqueued (see runRepairOnly for the evaluation ruling and its
	// evidence). Default false = the ADR 0085 Stage C backfill.
	RepairOnly bool
	// DryRun plans only and writes nothing. Enforced by construction in both
	// modes: every dry-run path ends in errDryRunRollback.
	DryRun bool
}

// Deps is the minimal object graph the per-document core needs. Wire builds
// the production one; tests may build it from their own pool.
type Deps struct {
	Runner   db.TxRunner
	Freeze   *docapp.FreezeService
	Evaluate approvalapp.ReleaseEvaluationEnqueuer
}

func (d Deps) validate() error {
	if d.Runner == nil {
		return errors.New("backfill: tx runner not configured")
	}
	if d.Freeze == nil {
		return errors.New("backfill: freeze service not configured")
	}
	if d.Evaluate == nil {
		return errors.New("backfill: release evaluation enqueuer not configured")
	}
	return nil
}

// Outcome is the per-document verdict.
type Outcome string

// Outcome values.
const (
	// OutcomeBackfilled means the generation was created and the pipeline armed.
	OutcomeBackfilled Outcome = "backfilled"
	// OutcomeAlreadyBackfilled means a generation for this exact identity
	// already exists; replaying wrote nothing.
	OutcomeAlreadyBackfilled Outcome = "already-backfilled"
	// OutcomeRepaired is the repair-only verdict: the freeze lineage was
	// re-pinned and materialization re-dispatched; no approval fact and no
	// release generation were written.
	OutcomeRepaired Outcome = "repaired"
	// OutcomePlanned is the dry-run verdict: preflight passed, nothing written.
	OutcomePlanned Outcome = "planned"
	// OutcomeFailed means preflight (or a write) rejected this document. The
	// transaction rolled back; other documents are unaffected.
	OutcomeFailed Outcome = "failed"
)

// Result is the per-document report line.
type Result struct {
	DocumentID   string
	Outcome      Outcome
	GenerationID string
	Detail       string
	Err          error
}

// String renders one report line. It never prints connection details or any
// other secret — only ids, the verdict, and the error text.
func (r Result) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "doc=%s outcome=%s", r.DocumentID, r.Outcome)
	if r.GenerationID != "" {
		fmt.Fprintf(&b, " generation=%s", r.GenerationID)
	}
	if r.Detail != "" {
		fmt.Fprintf(&b, " %s", r.Detail)
	}
	if r.Err != nil {
		fmt.Fprintf(&b, " error=%q", r.Err.Error())
	}
	return b.String()
}

// errDryRunRollback aborts the transaction after a successful dry-run
// preflight. Dry-run therefore cannot write by construction rather than by
// discipline: every dry-run path ends in a rollback.
var errDryRunRollback = errors.New("backfill: dry-run rollback")

// nilGenerationUUID is the sentinel the generation-aware staging-outbox dedupe
// indexes COALESCE a NULL release_generation_id to (migration 0310 §5,
// render/fanout/staging_outbox.go:19). Legacy, generation-less rows live in
// this slot, so any query that has to speak the dedupe key must use it too.
const nilGenerationUUID = "00000000-0000-0000-0000-000000000000"

// RunDocument backfills exactly one document. It never panics on bad input and
// never leaks a partial write: every failure rolls the whole transaction back
// and is reported on the returned Result, so the caller can continue with the
// next document in the allowlist.
func RunDocument(ctx context.Context, deps Deps, documentID string, opts Options) Result {
	res := Result{DocumentID: documentID, Outcome: OutcomeFailed}

	if err := deps.validate(); err != nil {
		res.Err = err
		return res
	}
	if _, err := uuid.Parse(strings.TrimSpace(documentID)); err != nil {
		res.Err = fmt.Errorf("backfill: %q is not a document uuid", documentID)
		return res
	}
	documentID = strings.TrimSpace(documentID)

	// Background-bypass marker, set at this tool's composition root exactly as
	// the release_evaluate worker sets it at ITS root. authz.BypassSystem fails
	// closed without it, so the tx-local tripwire bypass cannot be obtained by
	// any request-scoped path.
	ctx = authz.WithBackgroundBypass(ctx)

	tenantID, err := lookupTenant(ctx, deps.Runner, documentID)
	if err != nil {
		res.Err = err
		return res
	}

	txErr := deps.Runner.Do(ctx, func(tx *sql.Tx) error {
		// Same tx-local setup the release_evaluate job performs: tenant GUC
		// (RLS backstop) then the scheduler bypass GUC (documents tripwire).
		// Both are SET LOCAL — they die with the transaction.
		if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
			return fmt.Errorf("backfill: seed tenant: %w", err)
		}
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return fmt.Errorf("backfill: bypass: %w", err)
		}

		if opts.RepairOnly {
			return runRepairOnly(ctx, tx, deps, tenantID, documentID, opts.DryRun, &res)
		}

		pre, err := preflight(ctx, tx, tenantID, documentID)
		if err != nil {
			return err
		}

		if pre.existingGenerationID != "" {
			res.Outcome = OutcomeAlreadyBackfilled
			res.GenerationID = pre.existingGenerationID
			res.Detail = "already backfilled"
			if opts.DryRun {
				return errDryRunRollback
			}
			return nil
		}

		if opts.DryRun {
			res.Outcome = OutcomePlanned
			res.Detail = pre.plan()
			return errDryRunRollback
		}

		genID, genKey, err := approvalapp.RecordApprovalFactTx(ctx, tx, approvalapp.ApprovalFactInput{
			TenantID:           tenantID,
			DocumentID:         documentID,
			ApprovalInstanceID: pre.instanceID,
			RevisionVersion:    pre.revisionVersion,
			FrozenContentHash:  pre.frozenContentHash,
			FinalApproverID:    pre.finalApproverID,
			SubmittedBy:        pre.submittedBy,
		})
		if err != nil {
			return fmt.Errorf("backfill: record approval fact: %w", err)
		}

		// Pin-equivalent repair. Not Pin: the freeze is already final and
		// re-running it would recompute a values_hash the approver never signed.
		// It DOES re-pin frozen_revision_id from the current revision in this
		// same tx (migration 0313 header): that fresh, true pin is what makes
		// legacy pre-0313 documents materializable again.
		if err := deps.Freeze.RepairMaterialization(ctx, tx, tenantID, documentID, genID); err != nil {
			return fmt.Errorf("backfill: repair materialization: %w", err)
		}

		// Zero runAt = "evaluate now" (the port's documented contract). The
		// generation holds on `materializing` until the worker lands both
		// artifact facts, which re-enqueues evaluation and releases.
		if err := deps.Evaluate.EnqueueReleaseEvaluationTx(ctx, tx, genKey, time.Time{}); err != nil {
			return fmt.Errorf("backfill: enqueue release evaluation: %w", err)
		}

		res.Outcome = OutcomeBackfilled
		res.GenerationID = genID
		return nil
	})

	switch {
	case errors.Is(txErr, errDryRunRollback):
		// Expected: the dry-run transaction was rolled back on purpose.
		return res
	case txErr != nil:
		res.Outcome = OutcomeFailed
		res.GenerationID = ""
		res.Err = txErr
		return res
	default:
		return res
	}
}

// runRepairOnly is the operator-ratified expurgo pass, executed inside the
// caller's already-seeded, already-bypassed transaction.
//
// Sequence, all in the caller's transaction: repairPreflight (which classifies
// BOTH the stale staging rows and the stale delivery events, refusing outright
// if anything on either layer is still in flight) → purgeStaleDispatch →
// purgeStaleDeliveryEvents → RepairMaterialization → requireMaterializeEnqueued.
// The two purges run before the re-dispatch because their whole purpose is to
// free the keys that dispatch is about to write on.
//
// The only production seam it drives is FreezeService.RepairMaterialization,
// which re-pins documents.frozen_revision_id from the CURRENT revision in this
// same tx, carries values_hash / values_frozen_at over unchanged (re-assigned
// identical — see the package doc's literal write set), and enqueues
// materialization. The worker then verifies the fetched bytes against the
// pinned revision's document_revisions.content_hash before writing any artifact
// (freeze_service.go Materialize), so a repair can never put a signature on
// substituted content.
//
// Deliberately absent — RecordApprovalFactTx. Repair restores invalid
// artifacts; it does not re-approve, and it must not manufacture or re-stamp an
// approval fact for a document whose approval history is already correct.
//
// Deliberately absent — EnqueueReleaseEvaluationTx. Two independent reasons,
// both fail-closed rather than optimistic:
//
//   - For a PUBLISHED document the evaluation is a provable no-op that
//     converges nothing: EvaluateRelease returns ReleaseActionNoop because
//     `published` is not in the releasable status set
//     (approval/domain/release.go:124-127 + :165-167), and the coordinator's
//     only write on that branch is touchEvaluatedTx — last_evaluated_at plus a
//     hold_reason clear (approval/application/release_coordinator.go:319-320).
//     It never demotes the document and never inserts a generation row (only
//     RecordApprovalFactTx does that, release_facts.go:158-174), so including
//     it would be SAFE — merely pointless.
//   - Repair-only holds no generation key to enqueue with. A legacy
//     generation-less document has no release_generations row at all, and a
//     synthesized key resolves to ErrReleaseGenerationNotFound
//     (release_facts.go:26, release_coordinator.go:140), which the worker logs
//     and skips (approval/jobs/release_evaluate_job.go:56-60). Enqueuing a job
//     that is known to find nothing is noise, not convergence.
//
// Convergence for the case that actually needs it stays owned by the production
// path: when the repaired render lands both artifacts, RecordArtifactFactTx
// enqueues the evaluation itself, in the artifact-persistence transaction
// (release_facts.go:377-387).
func runRepairOnly(ctx context.Context, tx db.Tx, deps Deps, tenantID, documentID string, dryRun bool, res *Result) error {
	pre, err := repairPreflight(ctx, tx, tenantID, documentID)
	if err != nil {
		return err
	}

	res.GenerationID = pre.generationID

	if dryRun {
		res.Outcome = OutcomePlanned
		res.Detail = pre.plan()
		return errDryRunRollback
	}

	if err := purgeStaleDispatch(ctx, tx, pre.staleDispatch); err != nil {
		return err
	}

	if err := purgeStaleDeliveryEvents(ctx, tx, pre.staleEvents); err != nil {
		return err
	}

	if err := deps.Freeze.RepairMaterialization(ctx, tx, tenantID, documentID, pre.generationID); err != nil {
		return fmt.Errorf("backfill: repair materialization: %w", err)
	}

	outboxID, err := requireMaterializeEnqueued(ctx, tx, tenantID, documentID, pre.generationID)
	if err != nil {
		return err
	}

	res.Outcome = OutcomeRepaired
	res.Detail = fmt.Sprintf("re-pinned frozen_revision_id=%s, %s, %s, re-dispatched materialization (outbox=%s status=pending, document status=%s); approval facts untouched",
		pre.revisionID,
		describePurge(pre.staleDispatch, "purged"),
		describeEventPurge(pre.staleEvents, "purged"),
		outboxID, pre.status)
	return nil
}

// repairPreflightData is what the repair-only preflight proved.
type repairPreflightData struct {
	status       string
	revisionID   string
	generationID string
	// staleDispatch are the TERMINAL staging rows on this repair's dedupe key
	// that the write path will purge. Empty is the common case. A non-terminal
	// row never reaches here — preflight refuses instead.
	staleDispatch []dispatchRow
	// staleEvents are the TERMINAL delivery-ledger rows on this repair's two
	// idempotency keys, purged by the same write path and under the same rule.
	staleEvents []deliveryEventRow
}

func (p repairPreflightData) plan() string {
	generation := p.generationID
	if generation == "" {
		generation = "none (legacy, generation-less)"
	}
	return fmt.Sprintf("would re-pin frozen_revision_id from current revision %s, %s, %s, and re-dispatch materialization for a %s document (generation=%s); no approval fact, no release evaluation",
		p.revisionID,
		describePurge(p.staleDispatch, "would purge"),
		describeEventPurge(p.staleEvents, "would purge"),
		p.status, generation)
}

// repairOnlyStatuses is the CLOSED post-approval status set repair-only
// accepts. `approved` and `published` are the two states in which the
// document's artifact set is the governing one and the approval history is
// already terminal, so re-rendering it restores integrity instead of changing
// history. Every other status — including `scheduled`, `superseded`, `obsolete`
// and anything pre-terminal — is refused: no operator ruling covers them, and
// guessing would be exactly the fallback this tool refuses to have.
var repairOnlyStatuses = map[string]struct{}{
	string(docdom.DocStatusApproved):  {},
	string(docdom.DocStatusPublished): {},
}

// repairPreflight proves every precondition the repair write path assumes.
// Fail-closed throughout: nothing is inferred, defaulted or substituted.
func repairPreflight(ctx context.Context, tx db.Tx, tenantID, documentID string) (repairPreflightData, error) {
	var (
		pre        repairPreflightData
		revisionID sql.NullString
		valuesHash []byte
		frozenAt   *time.Time
	)
	err := tx.QueryRowContext(ctx, `
		SELECT status, current_revision_id::text, values_hash, values_frozen_at
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		documentID, tenantID,
	).Scan(&pre.status, &revisionID, &valuesHash, &frozenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return repairPreflightData{}, fmt.Errorf("repair preflight: document %s not found", documentID)
	}
	if err != nil {
		return repairPreflightData{}, fmt.Errorf("repair preflight: load document: %w", err)
	}
	if _, ok := repairOnlyStatuses[pre.status]; !ok {
		return repairPreflightData{}, fmt.Errorf("repair preflight: document %s is %s; repair-only accepts only approved or published documents", documentID, pre.status)
	}
	if !revisionID.Valid || revisionID.String == "" {
		return repairPreflightData{}, fmt.Errorf("repair preflight: document %s has no current_revision_id to re-pin", documentID)
	}
	pre.revisionID = revisionID.String
	// Frozen means BOTH columns, the same rule the backfill preflight and
	// RepairMaterialization itself apply. Repair re-pins lineage; it never
	// invents the freeze.
	if frozenAt == nil || len(valuesHash) == 0 {
		return repairPreflightData{}, fmt.Errorf("repair preflight: document %s is not frozen (values_hash/values_frozen_at)", documentID)
	}

	// Post-approval is proved by the DURABLE approval record, not by the
	// mutable status column alone: an approved approval_instance is the fact
	// that the document crossed the approval boundary. No uniqueness is
	// required here (unlike the backfill preflight) because repair derives no
	// generation identity from it — it only needs the existence proof.
	if err := requireApprovedInstance(ctx, tx, tenantID, documentID); err != nil {
		return repairPreflightData{}, err
	}

	pre.generationID, err = loadHeadGeneration(ctx, tx, tenantID, documentID)
	if err != nil {
		return repairPreflightData{}, err
	}

	pre.staleDispatch, err = classifyStaleDispatch(ctx, tx, tenantID, documentID, pre.generationID)
	if err != nil {
		return repairPreflightData{}, err
	}

	pre.staleEvents, err = classifyStaleDeliveryEvents(ctx, tx, documentID,
		repairDeliveryKeys(tenantID, documentID, pre.generationID))
	if err != nil {
		return repairPreflightData{}, err
	}

	return pre, nil
}

// requireApprovedInstance fails closed when the document has no approved
// approval instance: repairing artifacts for a document that never completed
// approval would render content nobody signed.
func requireApprovedInstance(ctx context.Context, tx db.Tx, tenantID, documentID string) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM approval_instances
		     WHERE tenant_id = $1::uuid AND document_id = $2::uuid AND status = 'approved'
		)`,
		tenantID, documentID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("repair preflight: load approved instances: %w", err)
	}
	if !exists {
		return fmt.Errorf("repair preflight: document %s has no approved approval instance; it never completed approval", documentID)
	}
	return nil
}

// loadHeadGeneration returns the document's CURRENT freeze-head generation id,
// or "" when the document has none (the legacy, pre-ADR-0085 shape).
//
// Threading the true id matters twice: the worker's RecordArtifactFactTx records
// the repaired artifacts against that generation (and refuses any generation
// that is not the head, release_facts.go:314-321), and the materialize outbox
// dedupe key is generation-aware (migration 0310 §5). Passing "" while a
// generation exists would be a lie in both places.
func loadHeadGeneration(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `
		SELECT id::text
		  FROM release_generations
		 WHERE tenant_id = $1::uuid AND document_id = $2::uuid
		 ORDER BY generation_seq DESC
		 LIMIT 1`,
		tenantID, documentID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("repair preflight: load head release generation: %w", err)
	}
	return id, nil
}

// dispatchRow is one staging-outbox row that collides with this repair's dedupe
// key. The generation-aware unique index (migration 0310 §5) makes at most one
// such row exist per table, so a repair sees at most two.
type dispatchRow struct {
	table  string
	id     string
	status string
}

// terminalDispatchStatuses are the two statuses that mean the row's work is
// OVER: the dispatch either succeeded ('dispatched') or was dead-lettered
// ('failed') — staging_outbox.go:151-199. 'pending' and 'processing', the other
// two members of the CHECK constraint (baseline 0001_current_schema.sql:1256,
// :1372), mean work is in flight and are never purged.
var terminalDispatchStatuses = map[string]struct{}{
	"dispatched": {},
	"failed":     {},
}

// classifyStaleDispatch resolves what stands between this repair and a real
// re-render, and refuses rather than guessing.
//
// Both staging outboxes dedupe on
// (tenant_id, revision_id, COALESCE(release_generation_id, nil-uuid)) with
// ON CONFLICT DO NOTHING (migration 0310 §5 — the unique indexes are NOT
// partial, they cover every status; render/fanout/staging_outbox.go:93 and
// :133), and a skipped insert is reported as an empty id with a NIL error,
// which the dispatch enqueuer reads as "nothing to dispatch"
// (render/fanout/dispatchjobs/enqueuer.go:110-113). A colliding row therefore
// turns an enqueue into a successful-looking no-op.
//
// Both tables matter, and the pdf one is the non-obvious half. Even a
// successful materialize enqueue only gets the frozen.docx re-rendered: the
// worker's own downstream EnqueuePDFTx runs on the SAME key — same tenant, same
// revision_id (= document id), same generation id off the job payload
// (platform/worker/materialize_job_runner.go:146) — so a stale
// pdf_dispatch_outbox row swallows it too, leaving a repaired frozen.docx
// paired with the original blank final.pdf. Purging only the materialize side
// would produce exactly that half-repair.
//
// Verdict per row: terminal → purge it in the write path (operator-ratified
// expurgo: these are ops staging rows in the metaldocs schema recording the
// dispatch of the INVALID artifacts, not audit facts, and their retention purge
// — PurgeDispatched, staging_outbox.go:257-264 — would delete them anyway);
// anything else → refuse the whole repair, loudly and with the row named,
// because a pending/processing row is live work whose River job may still fire
// and whose removal would orphan it.
//
// revision_id carries documents.id here (the known column misnomer), which is
// what RepairMaterialization enqueues with.
func classifyStaleDispatch(ctx context.Context, tx db.Tx, tenantID, documentID, generationID string) ([]dispatchRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT 'metaldocs.materialize_dispatch_outbox' AS tbl, id::text, status
		  FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid
		   AND revision_id = $2::uuid
		   AND COALESCE(release_generation_id, $4::uuid) = COALESCE(NULLIF($3, '')::uuid, $4::uuid)
		 UNION ALL
		SELECT 'metaldocs.pdf_dispatch_outbox' AS tbl, id::text, status
		  FROM metaldocs.pdf_dispatch_outbox
		 WHERE tenant_id = $1::uuid
		   AND revision_id = $2::uuid
		   AND COALESCE(release_generation_id, $4::uuid) = COALESCE(NULLIF($3, '')::uuid, $4::uuid)`,
		tenantID, documentID, generationID, nilGenerationUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("repair preflight: load staging dispatch rows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stale []dispatchRow
	for rows.Next() {
		var r dispatchRow
		if err := rows.Scan(&r.table, &r.id, &r.status); err != nil {
			return nil, fmt.Errorf("repair preflight: scan staging dispatch row: %w", err)
		}
		if _, ok := terminalDispatchStatuses[r.status]; !ok {
			return nil, fmt.Errorf("repair preflight: document %s has an IN-FLIGHT dispatch row on this generation key (%s id=%s status=%s); repair would either be swallowed by it or orphan its River job — wait for that dispatch to finish, then re-run",
				documentID, r.table, r.id, r.status)
		}
		stale = append(stale, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repair preflight: iterate staging dispatch rows: %w", err)
	}
	return stale, nil
}

// purgeStaleDispatch deletes the terminal staging rows classifyStaleDispatch
// approved, inside the repair transaction — so a later failure (including the
// requireMaterializeEnqueued check) rolls the purge back with everything else.
//
// Deletion is by PRIMARY KEY alone, one statement per row, against a CONSTANT
// table name chosen by a closed switch: no dynamic SQL, and no way for a scanned
// value to address a table this tool is not allowed to touch. No tenant
// predicate is repeated here because the id came from classifyStaleDispatch's
// tenant-keyed read earlier in this same transaction, and the tenant GUC seeded
// at the top of the tx keeps the RLS policy engaged for the DELETE regardless
// (baseline 0001_current_schema.sql:4320, :4334).
func purgeStaleDispatch(ctx context.Context, tx db.Tx, stale []dispatchRow) error {
	for _, r := range stale {
		var stmt string
		switch r.table {
		case "metaldocs.materialize_dispatch_outbox":
			stmt = `DELETE FROM metaldocs.materialize_dispatch_outbox WHERE id = $1::uuid`
		case "metaldocs.pdf_dispatch_outbox":
			stmt = `DELETE FROM metaldocs.pdf_dispatch_outbox WHERE id = $1::uuid`
		default:
			return fmt.Errorf("repair purge: refusing unknown staging table %q", r.table)
		}
		res, err := tx.ExecContext(ctx, stmt, r.id)
		if err != nil {
			return fmt.Errorf("repair purge: delete %s id=%s: %w", r.table, r.id, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("repair purge: rows affected for %s id=%s: %w", r.table, r.id, err)
		}
		if n != 1 {
			return fmt.Errorf("repair purge: expected to delete exactly 1 row from %s id=%s, deleted %d", r.table, r.id, n)
		}
	}
	return nil
}

// deliveryEventRow is one metaldocs.outbox_events row on a delivery idempotency
// key this repair needs free. The table is UNIQUE on idempotency_key, so at most
// one row exists per key and a repair sees at most two.
type deliveryEventRow struct {
	eventID        string
	idempotencyKey string
	published      bool
	deadLettered   bool
}

// terminal reports whether the event's delivery is OVER: published or
// dead-lettered.
//
// Its negation is a deliberately CONSERVATIVE SUPERSET of the currently
// claimable set, not its exact complement. The consumer claims on
// published_at IS NULL AND dead_lettered_at IS NULL AND
// (next_attempt_at IS NULL OR next_attempt_at <= NOW())
// (platform/messaging/outbox/postgres/consumer.go:41-46), so an unpublished row
// with a FUTURE next_attempt_at is backing off — not claimable this second, but
// re-claimable the moment its timer expires. This treats it as in-flight and
// refuses, which is the correct call: the row is alive, and deleting it would
// drop a message that was about to be retried. The partial index
// idx_outbox_claimable (baseline 0001_current_schema.sql:3078) indexes only the
// first two predicates, which is why they alone are the durable "is this event
// finished" question.
func (e deliveryEventRow) terminal() bool { return e.published || e.deadLettered }

func (e deliveryEventRow) state() string {
	switch {
	case e.deadLettered:
		return "dead-lettered"
	case e.published:
		return "published"
	default:
		return "unpublished"
	}
}

// repairDeliveryKeys builds the two delivery idempotency keys this repair's
// dispatch jobs will publish on, in the SAME shape as
// dispatchjobs.dispatchIdempotencyKey (render/fanout/dispatchjobs/workers.go:46-52):
// "<prefix>:<tenant>:<revision>", with ":<generation>" appended only when the
// render is generation-aware. The revision leg carries documents.id (the known
// column misnomer that runs through the whole pin/fanout/worker chain), which is
// what the dispatch args are built from.
//
// The prefixes are the two the workers use verbatim: "materialize_fanout"
// (workers.go:88) and "docgen_v2_pdf" (workers.go:66).
//
// This is a DELIBERATE DUPLICATION of a production key function across a module
// boundary, and it is the tool's weakest seam: it cannot import
// dispatchIdempotencyKey (unexported) and there is no lint pinning the two
// together. If workers.go:46-52 ever changes shape — a new prefix, a new
// component, a different separator — this function MUST be changed with it, or
// repair will purge keys nobody publishes on and silently stop clearing the
// delivery layer. Anyone touching that function should treat this comment as the
// callback.
func repairDeliveryKeys(tenantID, documentID, generationID string) []string {
	build := func(prefix string) string {
		key := prefix + ":" + tenantID + ":" + documentID
		if generationID != "" {
			key += ":" + generationID
		}
		return key
	}
	return []string{build("materialize_fanout"), build("docgen_v2_pdf")}
}

// classifyStaleDeliveryEvents is layer 2 of the swallow stack (see the package
// doc). metaldocs.outbox_events is UNIQUE on idempotency_key and the publisher
// is ON CONFLICT DO NOTHING returning nil
// (platform/messaging/outbox/postgres/publisher.go:35, :43-57), so a leftover
// event on one of this repair's keys makes the dispatch job publish NOTHING and
// still report success — it then marks the staging row dispatched
// (dispatchjobs/workers.go:104-117) and the operator sees a completed River job
// with no work behind it. That is exactly how the live run of this tool rendered
// nothing while printing `repaired`.
//
// Verdict per row, same discipline as the staging classifier: terminal
// (published or dead-lettered) → purge; anything else → refuse the whole repair,
// because deleting an event the consumer may still deliver — now, or once its
// backoff expires; see deliveryEventRow.terminal — would drop a live message
// rather than free a dead key.
//
// Purging here is expurgo, not audit mutation. outbox_events is the DELIVERY
// ledger of the same invalid dispatch the staging purge already covers; the
// audit chain lives in metaldocs.audit_events, a different table this tool never
// touches. Worth knowing when judging the cost: unlike the staging rows, which
// retention eventually deletes (staging_outbox.go:257-264), NOTHING purges
// outbox_events today — no retention job addresses it — so a stale published
// event holds its key forever and every future repair on that key is swallowed
// until someone removes it.
func classifyStaleDeliveryEvents(ctx context.Context, tx db.Tx, documentID string, keys []string) ([]deliveryEventRow, error) {
	if len(keys) != 2 {
		return nil, fmt.Errorf("repair preflight: expected exactly 2 delivery idempotency keys, got %d", len(keys))
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT event_id, idempotency_key,
		       (published_at IS NOT NULL)     AS published,
		       (dead_lettered_at IS NOT NULL) AS dead_lettered
		  FROM metaldocs.outbox_events
		 WHERE idempotency_key IN ($1, $2)`,
		keys[0], keys[1],
	)
	if err != nil {
		return nil, fmt.Errorf("repair preflight: load delivery outbox events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stale []deliveryEventRow
	for rows.Next() {
		var e deliveryEventRow
		if err := rows.Scan(&e.eventID, &e.idempotencyKey, &e.published, &e.deadLettered); err != nil {
			return nil, fmt.Errorf("repair preflight: scan delivery outbox event: %w", err)
		}
		if !e.terminal() {
			return nil, fmt.Errorf("repair preflight: document %s has an IN-FLIGHT delivery event on idempotency_key %q (event_id=%s, neither published nor dead-lettered, so the consumer can still deliver it — now or after its backoff); repair would drop a live message — wait for that delivery to finish, then re-run",
				documentID, e.idempotencyKey, e.eventID)
		}
		stale = append(stale, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repair preflight: iterate delivery outbox events: %w", err)
	}
	return stale, nil
}

// purgeStaleDeliveryEvents frees the delivery keys, inside the repair
// transaction, by PRIMARY KEY (outbox_events_pkey on event_id, baseline
// 0001_current_schema.sql:2543-2544). No tenant predicate is possible or needed:
// outbox_events carries NO tenant_id column at all (baseline :1335-1352) — the
// tenant lives inside the idempotency key and the payload — and the ids came
// from classifyStaleDeliveryEvents' key-scoped read earlier in this same tx.
func purgeStaleDeliveryEvents(ctx context.Context, tx db.Tx, stale []deliveryEventRow) error {
	for _, e := range stale {
		res, err := tx.ExecContext(ctx,
			`DELETE FROM metaldocs.outbox_events WHERE event_id = $1`, e.eventID)
		if err != nil {
			return fmt.Errorf("repair purge: delete outbox event %s (key %q): %w", e.eventID, e.idempotencyKey, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("repair purge: rows affected for outbox event %s: %w", e.eventID, err)
		}
		if n != 1 {
			return fmt.Errorf("repair purge: expected to delete exactly 1 outbox event %s, deleted %d", e.eventID, n)
		}
	}
	return nil
}

// requireMaterializeEnqueued is the read-your-own-write proof that the repair
// actually dispatched something, and it closes the check-then-insert race that
// classifyStaleDispatch alone cannot.
//
// classifyStaleDispatch reads the dedupe key; RepairMaterialization inserts on
// it later. Two concurrent repair invocations can both pass the read with no
// pre-existing row, after which one INSERT wins and the other hits ON CONFLICT
// DO NOTHING — empty id, nil error, no River job
// (render/fanout/dispatchjobs/enqueuer.go:110-113) — and would print "repaired"
// having dispatched nothing. (The purge narrows this window for the
// pre-existing-row case, since DELETE takes a row lock, but cannot close the
// no-row case: there is nothing to lock.)
//
// Re-reading the key inside the same transaction is deterministic: the row this
// SELECT sees is either the one this tx just inserted or a committed one from a
// competing repair. Both outcomes are honest — a pending row exists and WILL
// render — and absence is a hard failure that rolls the whole repair back. The
// status='pending' predicate matters: it is the enqueue state
// (staging_outbox.go:91-95), so a row left over in any other state cannot be
// mistaken for this repair's dispatch.
//
// LIMIT, stated plainly: this proves layer 1 ONLY — that the staging row and its
// River job exist. It cannot prove layer 2. The delivery publish happens later,
// in the River job's own transaction, long after this one commits
// (dispatchjobs/workers.go:104-117), so no read in this tx can observe it. What
// makes that later publish land is the delivery purge, not a check:
// purgeStaleDeliveryEvents removes the rows holding those idempotency keys, so
// the publisher's ON CONFLICT DO NOTHING has nothing to conflict with and the
// INSERT actually writes. The verification available here is the strongest one
// available here; the layer below is made safe by construction instead.
//
// Also worth stating: the pdf leg is unverifiable from this tool in principle.
// Its staging row and its event are both created by the WORKER, after a
// successful render. Repair frees both of its keys and stops there.
func requireMaterializeEnqueued(ctx context.Context, tx db.Tx, tenantID, documentID, generationID string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `
		SELECT id::text
		  FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid
		   AND revision_id = $2::uuid
		   AND COALESCE(release_generation_id, $4::uuid) = COALESCE(NULLIF($3, '')::uuid, $4::uuid)
		   AND status = 'pending'`,
		tenantID, documentID, generationID, nilGenerationUUID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("repair verify: document %s has no pending materialize dispatch row after RepairMaterialization — the enqueue was swallowed by the generation-aware dedupe (ON CONFLICT DO NOTHING) and nothing will render; rolling back", documentID)
	}
	if err != nil {
		return "", fmt.Errorf("repair verify: load materialize dispatch row: %w", err)
	}
	return id, nil
}

// describePurge renders the purge set for the operator-facing outcome line.
// Row ids are printed so the expurgo is auditable after the fact.
func describePurge(stale []dispatchRow, verb string) string {
	if len(stale) == 0 {
		return "no stale dispatch rows to purge"
	}
	parts := make([]string, 0, len(stale))
	for _, r := range stale {
		parts = append(parts, fmt.Sprintf("%s(id=%s status=%s)", strings.TrimPrefix(r.table, "metaldocs."), r.id, r.status))
	}
	return fmt.Sprintf("%s %d stale dispatch row(s): %s", verb, len(stale), strings.Join(parts, ", "))
}

// describeEventPurge renders the delivery-ledger purge set for the outcome line,
// naming each event id and the key it was holding.
func describeEventPurge(stale []deliveryEventRow, verb string) string {
	if len(stale) == 0 {
		return "no stale delivery events to purge"
	}
	parts := make([]string, 0, len(stale))
	for _, e := range stale {
		parts = append(parts, fmt.Sprintf("outbox_events(event_id=%s key=%s state=%s)", e.eventID, e.idempotencyKey, e.state()))
	}
	return fmt.Sprintf("%s %d stale delivery event(s): %s", verb, len(stale), strings.Join(parts, ", "))
}

// lookupTenant resolves the owning tenant of documentID.
//
// It runs in its own short read-only transaction, BEFORE the work transaction,
// because SeedTxTenant needs the tenant id it is about to seed. The read is
// legitimate unseeded: the tenant RLS policies are NULL-GUC permissive, which
// is the same posture every cross-tenant system sweep relies on. Keeping it
// outside the work transaction also keeps the work transaction seeded from its
// very first statement.
func lookupTenant(ctx context.Context, runner db.TxRunner, documentID string) (string, error) {
	var tenantID string
	err := runner.DoReadOnly(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			SELECT tenant_id::text FROM documents WHERE id = $1::uuid`,
			documentID,
		).Scan(&tenantID)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("backfill: document %s not found", documentID)
	}
	if err != nil {
		return "", fmt.Errorf("backfill: resolve tenant: %w", err)
	}
	return tenantID, nil
}

// preflightData is what preflight proved and what the write path consumes.
type preflightData struct {
	status               string
	controlledDocumentID string
	revisionVersion      int
	instanceID           string
	frozenContentHash    string
	submittedBy          string
	finalApproverID      string
	existingGenerationID string
}

func (p preflightData) plan() string {
	return fmt.Sprintf("would record approval fact (instance=%s rev_version=%d hash=%s approver=%s), repair materialization, enqueue release evaluation",
		p.instanceID, p.revisionVersion, p.frozenContentHash, p.finalApproverID)
}

// preflight proves every precondition the write path assumes. Every check
// fails closed: nothing is inferred, defaulted, or substituted.
func preflight(ctx context.Context, tx db.Tx, tenantID, documentID string) (preflightData, error) {
	var (
		pre          preflightData
		controlledID sql.NullString
		revisionID   sql.NullString
		valuesHash   []byte
		frozenAt     *time.Time
	)
	err := tx.QueryRowContext(ctx, `
		SELECT status, controlled_document_id::text, current_revision_id::text,
		       revision_version, values_hash, values_frozen_at
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		documentID, tenantID,
	).Scan(&pre.status, &controlledID, &revisionID, &pre.revisionVersion, &valuesHash, &frozenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return preflightData{}, fmt.Errorf("preflight: document %s not found", documentID)
	}
	if err != nil {
		return preflightData{}, fmt.Errorf("preflight: load document: %w", err)
	}
	if pre.status != "approved" {
		return preflightData{}, fmt.Errorf("preflight: document %s is %s, not approved", documentID, pre.status)
	}
	if !controlledID.Valid || controlledID.String == "" {
		return preflightData{}, fmt.Errorf("preflight: document %s has no controlled_document_id", documentID)
	}
	pre.controlledDocumentID = controlledID.String
	if !revisionID.Valid || revisionID.String == "" {
		return preflightData{}, fmt.Errorf("preflight: document %s has no current_revision_id", documentID)
	}
	// Pinned means BOTH columns. A half-pinned document is a corrupt freeze, not
	// a repairable one — no fallback, no re-freeze.
	if frozenAt == nil || len(valuesHash) == 0 {
		return preflightData{}, fmt.Errorf("preflight: document %s is not pinned (values_hash/values_frozen_at)", documentID)
	}

	instanceID, frozenHash, submittedBy, err := loadApprovedInstance(ctx, tx, tenantID, documentID)
	if err != nil {
		return preflightData{}, err
	}
	pre.instanceID, pre.frozenContentHash, pre.submittedBy = instanceID, frozenHash, submittedBy

	pre.finalApproverID, err = loadFinalApprover(ctx, tx, tenantID, instanceID)
	if err != nil {
		return preflightData{}, err
	}

	pre.existingGenerationID, err = loadExistingGeneration(ctx, tx, tenantID, documentID, instanceID, revisionID.String, pre.revisionVersion, pre.frozenContentHash)
	if err != nil {
		return preflightData{}, err
	}

	return pre, nil
}

// loadApprovedInstance requires EXACTLY ONE approved instance. Two approved
// instances on one document means the legacy history is ambiguous about which
// approval the release represents; the tool refuses to guess.
func loadApprovedInstance(ctx context.Context, tx db.Tx, tenantID, documentID string) (instanceID, frozenHash, submittedBy string, err error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id::text, frozen_content_hash, submitted_by
		  FROM approval_instances
		 WHERE tenant_id = $1::uuid AND document_id = $2::uuid AND status = 'approved'
		 ORDER BY id`,
		tenantID, documentID,
	)
	if err != nil {
		return "", "", "", fmt.Errorf("preflight: load approved instances: %w", err)
	}
	defer rows.Close()

	var found int
	var hash, submitter sql.NullString
	for rows.Next() {
		found++
		if found > 1 {
			return "", "", "", fmt.Errorf("preflight: document %s has %d+ approved approval instances; refusing to guess", documentID, found)
		}
		if err := rows.Scan(&instanceID, &hash, &submitter); err != nil {
			return "", "", "", fmt.Errorf("preflight: scan approved instance: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return "", "", "", fmt.Errorf("preflight: iterate approved instances: %w", err)
	}
	if found == 0 {
		return "", "", "", fmt.Errorf("preflight: document %s has no approved approval instance", documentID)
	}
	if !hash.Valid || hash.String == "" {
		return "", "", "", fmt.Errorf("preflight: approval instance %s has no frozen_content_hash", instanceID)
	}
	return instanceID, hash.String, submitter.String, nil
}

// loadFinalApprover returns the actor of the LATEST approve signoff on the
// instance — the approver whose decision made it terminal. Fails closed when
// the instance is approved but carries no approve signoff at all: that is a
// broken history, and inventing an approver would forge the release record.
func loadFinalApprover(ctx context.Context, tx db.Tx, tenantID, instanceID string) (string, error) {
	var actor sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT actor_user_id
		  FROM approval_signoffs
		 WHERE approval_instance_id = $1::uuid
		   AND actor_tenant_id = $2::uuid
		   AND decision = 'approve'
		 ORDER BY signed_at DESC, id DESC
		 LIMIT 1`,
		instanceID, tenantID,
	).Scan(&actor)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("preflight: approval instance %s has no approve signoff", instanceID)
	}
	if err != nil {
		return "", fmt.Errorf("preflight: load final approver: %w", err)
	}
	if !actor.Valid || actor.String == "" {
		return "", fmt.Errorf("preflight: approval instance %s has an approve signoff with no actor", instanceID)
	}
	return actor.String, nil
}

// loadExistingGeneration looks the 7-column generation identity up verbatim, so
// "already backfilled" means the identical generation, not merely some
// generation on the same document.
func loadExistingGeneration(ctx context.Context, tx db.Tx, tenantID, documentID, instanceID, revisionID string, revisionVersion int, frozenHash string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `
		SELECT id::text
		  FROM release_generations
		 WHERE tenant_id = $1::uuid
		   AND subject_kind = $2
		   AND document_id = $3::uuid
		   AND approval_instance_id = $4::uuid
		   AND revision_id = $5::uuid
		   AND revision_version = $6
		   AND frozen_content_hash = $7`,
		tenantID, string(approvaldom.SubjectKindDocument), documentID, instanceID, revisionID, revisionVersion, frozenHash,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("preflight: load existing generation: %w", err)
	}
	return id, nil
}
