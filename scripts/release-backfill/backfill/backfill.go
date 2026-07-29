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
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

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

// RunDocument backfills exactly one document. It never panics on bad input and
// never leaks a partial write: every failure rolls the whole transaction back
// and is reported on the returned Result, so the caller can continue with the
// next document in the allowlist.
func RunDocument(ctx context.Context, deps Deps, documentID string, dryRun bool) Result {
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

		pre, err := preflight(ctx, tx, tenantID, documentID)
		if err != nil {
			return err
		}

		if pre.existingGenerationID != "" {
			res.Outcome = OutcomeAlreadyBackfilled
			res.GenerationID = pre.existingGenerationID
			res.Detail = "already backfilled"
			if dryRun {
				return errDryRunRollback
			}
			return nil
		}

		if dryRun {
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
