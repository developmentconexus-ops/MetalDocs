package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

// ReleaseSystemPrincipal is the stable system identity the release
// coordinator's governance events are attributed to (ADR 0085: "the release
// event's actor is the system principal; the causal identities — approval
// instance, final approver, submitter — are payload").
//
// It is NOT a user id and never resolves to one: attributing an automatic
// release to the last human who touched the document would be a false audit
// claim.
const ReleaseSystemPrincipal = "system:release-coordinator"

// releaseEventIDNamespace derives deterministic lifecycle event ids so a
// retried release cannot fan out a duplicate notification. Fixed UUID; never
// regenerate it.
var releaseEventIDNamespace = uuid.MustParse("6f0f5b6e-2a41-4d9a-9b4b-6d0e6a0f0851")

// ProfileReviewIntervalReader is the approval module's narrow read port onto
// the taxonomy module's review cadence. Implemented by an approval
// infrastructure adapter over the taxonomy profile repository — the approval
// module never queries metaldocs.document_profiles itself.
//
// IMPORTANT (H-PRE-1): implementations run their OWN short transaction and
// perform an authz-recording read, so the coordinator MUST call this BEFORE
// opening the lock-holding release transaction. Never move it inside.
type ProfileReviewIntervalReader interface {
	ReviewIntervalDays(ctx context.Context, tenantID, profileCode string) (int, error)
}

// ReleaseCoordinator is the single writer of the released state (ADR 0085).
//
// Publication is not an endpoint: it is the coordinator's reaction to durable
// facts. Every path — signoff, review verdict, artifact completion, effective
// date timer, reconciliation — funnels into the same idempotent Evaluate, so
// there is exactly one place that decides whether a document is released and
// exactly one transaction shape that releases it.
type ReleaseCoordinator struct {
	repo               infrastructure.ApprovalRepository
	emitter            EventEmitter
	clock              Clock
	evaluationEnqueuer ReleaseEvaluationEnqueuer
	lifecycleEnqueuer  docsdomain.LifecycleEventEnqueuer
	reviewIntervals    ProfileReviewIntervalReader
}

// NewReleaseCoordinator constructs the coordinator with its mandatory
// collaborators. Optional ports are wired via the With* builders.
func NewReleaseCoordinator(repo infrastructure.ApprovalRepository, emitter EventEmitter, clock Clock) *ReleaseCoordinator {
	return &ReleaseCoordinator{repo: repo, emitter: emitter, clock: clock}
}

// WithEvaluationEnqueuer wires the timer/self-enqueue port.
func (c *ReleaseCoordinator) WithEvaluationEnqueuer(e ReleaseEvaluationEnqueuer) *ReleaseCoordinator {
	c.evaluationEnqueuer = e
	return c
}

// WithLifecycleEnqueuer wires the notification-fanout port.
func (c *ReleaseCoordinator) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *ReleaseCoordinator {
	c.lifecycleEnqueuer = e
	return c
}

// WithProfileReviewIntervalReader wires the review-cadence read port.
func (c *ReleaseCoordinator) WithProfileReviewIntervalReader(r ProfileReviewIntervalReader) *ReleaseCoordinator {
	c.reviewIntervals = r
	return c
}

// ReleaseEvaluation is what one Evaluate call concluded — the observable
// outcome the projection and the tests assert on.
type ReleaseEvaluation struct {
	DocumentID string
	Action     domain.ReleaseAction
	Hold       domain.ReleaseHoldReason
	// EffectiveFrom is the ACTUAL release instant, set only when the release
	// transaction won.
	EffectiveFrom *time.Time
	// SupersededDocumentIDs lists targets this release moved to superseded.
	SupersededDocumentIDs []string
	// ScheduledFor is the armed timer instant on a schedule outcome.
	ScheduledFor *time.Time
}

// errSupersedeConflict is the in-tx sentinel that forces the whole release
// transaction to roll back (no partial release), after which the hold is
// recorded in a fresh transaction.
var errSupersedeConflict = errors.New("approval: supersession target moved")

// preflight is the small off-lock snapshot used to resolve the review cadence
// before any lock is taken.
type releasePreflight struct {
	DocumentID   string
	ProfileCode  string
	Found        bool
	AlreadyDone  bool
	ReviewPlan   *time.Time
	ReviewDays   int
	ControlledID string
}

// Evaluate runs the ADR 0085 release predicate for one generation and, when it
// holds, performs the release transaction. It is idempotent and safe to
// deliver any number of times, from any trigger.
func (c *ReleaseCoordinator) Evaluate(ctx context.Context, runner db.TxRunner, key domain.ReleaseGenerationKey) (ReleaseEvaluation, error) {
	if key.SubjectKind != "" && key.SubjectKind != domain.SubjectKindDocument {
		// The coordinator is document-only by ADR 0085. Template approval has
		// no artifact set and no effective window; refusing loudly beats
		// silently doing nothing.
		return ReleaseEvaluation{}, fmt.Errorf("release coordinator: subject kind %q is not releasable", key.SubjectKind)
	}
	ctx = authz.WithBackgroundBypass(ctx)

	// Phase 1 — off-lock preflight. Resolving the review cadence requires an
	// authz-recording taxonomy read (H-PRE-1), so it must complete before the
	// release transaction takes a single row lock.
	pre, err := c.preflight(ctx, runner, key)
	if err != nil {
		return ReleaseEvaluation{}, err
	}
	if !pre.Found {
		return ReleaseEvaluation{}, ErrReleaseGenerationNotFound
	}

	// Phase 2 — the release transaction.
	eval, err := c.evaluateTx(ctx, runner, key, pre)
	if errors.Is(err, errSupersedeConflict) {
		// The whole release rolled back. Record the hold in its own
		// transaction so the conflict is a VISIBLE state, not a lost error.
		if holdErr := c.recordHold(ctx, runner, key, domain.HoldSupersedeConflict, "supersession target is no longer the published head"); holdErr != nil {
			return ReleaseEvaluation{}, fmt.Errorf("release coordinator: record supersede conflict hold: %w", holdErr)
		}
		return ReleaseEvaluation{
			DocumentID: pre.DocumentID,
			Action:     domain.ReleaseActionHold,
			Hold:       domain.HoldSupersedeConflict,
		}, nil
	}
	if IsLostReleaseCAS(err) {
		// Another evaluation performed the complete release; ours rolled back.
		// That is the exactly-one-winner property working, not a failure.
		return ReleaseEvaluation{DocumentID: pre.DocumentID, Action: domain.ReleaseActionNoop}, nil
	}
	if err != nil {
		return ReleaseEvaluation{}, err
	}
	return eval, nil
}

func (c *ReleaseCoordinator) preflight(ctx context.Context, runner db.TxRunner, key domain.ReleaseGenerationKey) (releasePreflight, error) {
	var pre releasePreflight
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(ctx, tx, key.TenantID); err != nil {
			return fmt.Errorf("release preflight: seed tenant: %w", err)
		}
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return fmt.Errorf("release preflight: bypass: %w", err)
		}
		var releasedAt sql.NullTime
		var profileCode, controlledID sql.NullString
		var reviewPlan sql.NullTime
		err := tx.QueryRowContext(ctx, `
			SELECT g.released_at, d.profile_code_snapshot, d.controlled_document_id::text, d.review_due_at
			  FROM release_generations g
			  JOIN documents d
			    ON d.id = g.document_id AND d.tenant_id = g.tenant_id
			 WHERE g.tenant_id            = $1::uuid
			   AND g.document_id          = $2::uuid
			   AND g.approval_instance_id = $3::uuid
			   AND g.revision_id          = $4::uuid
			   AND g.revision_version     = $5
			   AND g.frozen_content_hash  = $6`,
			key.TenantID, key.DocumentID, key.ApprovalInstanceID,
			key.RevisionID, key.RevisionVersion, key.FrozenContentHash,
		).Scan(&releasedAt, &profileCode, &controlledID, &reviewPlan)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("release preflight: load generation: %w", err)
		}
		pre.Found = true
		pre.DocumentID = key.DocumentID
		pre.AlreadyDone = releasedAt.Valid
		pre.ProfileCode = profileCode.String
		pre.ControlledID = controlledID.String
		if reviewPlan.Valid {
			t := reviewPlan.Time
			pre.ReviewPlan = &t
		}
		return nil
	})
	if err != nil {
		return releasePreflight{}, err
	}
	if !pre.Found || pre.AlreadyDone {
		return pre, nil
	}
	// H-PRE-1 safe: this is outside every transaction above and below.
	if c.reviewIntervals != nil && pre.ProfileCode != "" {
		days, err := c.reviewIntervals.ReviewIntervalDays(ctx, key.TenantID, pre.ProfileCode)
		if err != nil {
			return releasePreflight{}, fmt.Errorf("release preflight: review interval for profile %q: %w", pre.ProfileCode, err)
		}
		pre.ReviewDays = days
	}
	return pre, nil
}

// releaseDecisionContext is the per-evaluation snapshot the decision table and
// its handlers share — everything that does NOT change when the source document
// row is re-read under its lock.
type releaseDecisionContext struct {
	gen  releaseGenerationRow
	pre  releasePreflight
	head bool
	now  time.Time
	// targets is the supersession set: discovered lock-free, deduplicated and
	// sorted by document id, then locked in that order before any write.
	targets []string
	// targetControlledIDs maps each LOCKED document id to that row's own
	// controlled_document_id ("" when it has none). A plan-named target can
	// belong to a different controlled document than the source, so its
	// lifecycle event must be addressed with ITS controlled document, not the
	// source's — retaining the locked state is what makes that knowable.
	targetControlledIDs map[string]string
	out                 *ReleaseEvaluation
}

// decide applies the pure ADR 0085 predicate to one source-document snapshot.
func (rc releaseDecisionContext) decide(doc documentReleaseState) domain.ReleaseDecision {
	return domain.EvaluateRelease(domain.ReleaseInputs{
		Released:             rc.gen.ReleasedAt.Valid,
		IsFreezeHead:         rc.head,
		ApprovalFact:         rc.gen.ApprovalFactAt.Valid,
		ArtifactFact:         rc.gen.ArtifactFactAt.Valid,
		DocumentStatus:       doc.Status,
		PlannedEffectiveFrom: doc.PlannedFrom,
		EffectiveTo:          doc.EffectiveTo,
	}, rc.now)
}

// evaluateTx opens the release transaction, takes the generation lock and runs
// the ADR 0085 decision table over an UNLOCKED source-document read.
func (c *ReleaseCoordinator) evaluateTx(ctx context.Context, runner db.TxRunner, key domain.ReleaseGenerationKey, pre releasePreflight) (ReleaseEvaluation, error) {
	var out ReleaseEvaluation
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(ctx, tx, key.TenantID); err != nil {
			return fmt.Errorf("release: seed tenant: %w", err)
		}
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return fmt.Errorf("release: bypass: %w", err)
		}

		gen, err := loadReleaseGenerationByKeyTx(ctx, tx, key)
		if err != nil {
			return err
		}
		head, err := isFreezeHeadTx(ctx, tx, gen)
		if err != nil {
			return err
		}

		// Lock order across the whole release is: generation (taken by the load
		// above, ALWAYS first) → the full document set (source + supersession
		// targets) locked one row at a time in sorted id order.
		//
		// That is why this read is UNLOCKED. The targets are only knowable after
		// reading the source row, so locking the source here would put it ahead
		// of a lower-sorting target and reintroduce exactly the AB-BA deadlock
		// the sort exists to prevent. The source row is re-read under its lock
		// and the decision re-run before anything is written.
		doc, err := readDocumentForRelease(ctx, tx, key.TenantID, key.DocumentID)
		if err != nil {
			return err
		}

		rc := releaseDecisionContext{gen: gen, pre: pre, head: head, now: c.clock.Now().UTC(), out: &out}
		return c.applyDecision(ctx, tx, rc, doc, rc.decide(doc), false)
	})
	if err != nil {
		return ReleaseEvaluation{}, err
	}
	return out, nil
}

// applyDecision executes one outcome of the ADR 0085 decision table. The branch
// count IS that table — each case's body is extracted to a small,
// outcome-named helper (scheduleReleaseTx, lockAndRedecide/releaseTx) so the
// dispatch itself stays a flat, readable switch; the case list is never
// split or turned into a lookup map, which would hide which outcome each
// conjunct of the table produces.
//
// locked reports whether the document lock set is already held: the release
// branch takes the locks, re-decides under them and re-enters here exactly once
// with the fresh decision.
func (c *ReleaseCoordinator) applyDecision(ctx context.Context, tx *sql.Tx, rc releaseDecisionContext, doc documentReleaseState, decision domain.ReleaseDecision, locked bool) error {
	gen := rc.gen
	*rc.out = ReleaseEvaluation{DocumentID: gen.DocumentID, Action: decision.Action, Hold: decision.Hold}

	switch decision.Action {
	case domain.ReleaseActionNoop:
		return touchEvaluatedTx(ctx, tx, gen.ID, "")

	case domain.ReleaseActionHold:
		return touchEvaluatedTx(ctx, tx, gen.ID, decision.Hold)

	case domain.ReleaseActionSchedule:
		return c.scheduleReleaseTx(ctx, tx, rc, doc, decision)

	case domain.ReleaseActionRelease:
		if !locked {
			return c.lockAndRedecide(ctx, tx, rc, doc)
		}
		return c.releaseTx(ctx, tx, rc, doc)
	}
	return fmt.Errorf("release: unhandled action %q", decision.Action)
}

// scheduleReleaseTx handles the ADR 0085 ReleaseActionSchedule outcome: CAS
// the document approved -> scheduled (a lost CAS falls back to the noop
// outcome), records the evaluation, and arms the effective-date timer.
func (c *ReleaseCoordinator) scheduleReleaseTx(ctx context.Context, tx *sql.Tx, rc releaseDecisionContext, doc documentReleaseState, decision domain.ReleaseDecision) error {
	gen := rc.gen
	if doc.Status == string(docsdomain.DocStatusApproved) {
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusApproved, docsdomain.DocStatusScheduled); err != nil {
			return err
		}
		// No document pre-lock: this guarded UPDATE is itself the row lock
		// AND the CAS. Zero rows means a concurrent evaluation won.
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status              = 'scheduled',
			       effective_from      = NULL,
			       revision_version    = revision_version + 1,
			       schedule_generation = schedule_generation + 1
			 WHERE id = $1::uuid AND tenant_id = $2::uuid
			   AND status = 'approved' AND revision_version = $3`,
			gen.DocumentID, gen.TenantID, doc.RevisionVersion)
		if err != nil {
			return fmt.Errorf("release: transition to scheduled: %w", err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("release: scheduled rows affected: %w", err)
		}
		if affected == 0 {
			// Lost the CAS to a concurrent evaluation.
			rc.out.Action = domain.ReleaseActionNoop
			rc.out.Hold = ""
			return touchEvaluatedTx(ctx, tx, gen.ID, "")
		}
	}
	if err := touchEvaluatedTx(ctx, tx, gen.ID, decision.Hold); err != nil {
		return err
	}
	// Arm the effective-date timer. Firing just re-runs this same
	// evaluation, so a duplicate timer is harmless.
	if c.evaluationEnqueuer != nil && doc.PlannedFrom != nil {
		if err := c.evaluationEnqueuer.EnqueueReleaseEvaluationTx(ctx, tx, gen.Key(), *doc.PlannedFrom); err != nil {
			return fmt.Errorf("release: arm effective-date timer: %w", err)
		}
		rc.out.ScheduledFor = doc.PlannedFrom
	}
	return nil
}

// lockAndRedecide is the ADR 0085 deterministic lock step: discover the
// supersession targets lock-free, then lock the FULL set (source ∪ targets) in
// sorted id order, one `FOR UPDATE` statement per id — a single
// `id = ANY(...) ORDER BY id FOR UPDATE` gives no guarantee about the order
// Postgres actually acquires the locks in.
//
// Values may have moved between the unlocked read and the lock, so the source
// row is re-read under its own lock and the decision re-run; anything other
// than a release falls back to that outcome's normal handling.
func (c *ReleaseCoordinator) lockAndRedecide(ctx context.Context, tx *sql.Tx, rc releaseDecisionContext, doc documentReleaseState) error {
	gen := rc.gen
	targets, err := c.discoverSupersedeTargets(ctx, tx, rc, doc)
	if err != nil {
		return err
	}
	rc.targets = targets

	lockSet := append(append(make([]string, 0, len(targets)+1), gen.DocumentID), targets...)
	sort.Strings(lockSet)

	locked := doc
	rc.targetControlledIDs = make(map[string]string, len(lockSet))
	previous := "" // never equal to a real document id, so i==0 always locks
	for _, id := range lockSet {
		if id == previous {
			continue // sorted, so duplicates are adjacent
		}
		previous = id
		state, err := lockDocumentForRelease(ctx, tx, gen.TenantID, id)
		if err != nil {
			return err
		}
		// Retain the locked state instead of discarding it: each row's own
		// controlled document is what its lifecycle event must be addressed
		// with (see emitReleaseEvents).
		rc.targetControlledIDs[id] = state.ControlledID
		if id == gen.DocumentID {
			locked = state
		}
	}

	return c.applyDecision(ctx, tx, rc, locked, rc.decide(locked), true)
}

// discoverSupersedeTargets resolves the supersession set WITHOUT taking any
// lock: the plan's cross-document target (carried on the source row) plus the
// controlled document's current published head. Sorting by document id gives
// every concurrent release the SAME lock order, which is what makes the
// multi-row supersede deadlock-free.
func (c *ReleaseCoordinator) discoverSupersedeTargets(ctx context.Context, tx *sql.Tx, rc releaseDecisionContext, doc documentReleaseState) ([]string, error) {
	targets := map[string]struct{}{}
	if doc.SupersedeTarget != "" && doc.SupersedeTarget != rc.gen.DocumentID {
		targets[doc.SupersedeTarget] = struct{}{}
	}
	if rc.pre.ControlledID != "" {
		currentHead, err := c.repo.LoadCurrentPublishedHeadNoLock(ctx, tx, rc.gen.TenantID, rc.pre.ControlledID)
		if err != nil {
			return nil, fmt.Errorf("release: load current published head: %w", err)
		}
		if currentHead != "" && currentHead != rc.gen.DocumentID {
			targets[currentHead] = struct{}{}
		}
	}
	ordered := make([]string, 0, len(targets))
	for id := range targets {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	return ordered, nil
}

// releaseTx is the ADR 0085 release transaction: deterministic supersession,
// source CAS, actual effective_from, review-cycle recompute, facts, events.
// Every row it writes was already locked in sorted id order by lockAndRedecide.
// Any failure rolls the WHOLE thing back — there is no partial release.
func (c *ReleaseCoordinator) releaseTx(ctx context.Context, tx *sql.Tx, rc releaseDecisionContext, doc documentReleaseState) error {
	gen := rc.gen
	tenantID := gen.TenantID

	// --- supersession, in the order the rows were locked --------------------
	// The rows are held; MarkSuperseded's status = 'published' predicate stays
	// the guard that a target is still supersedable.
	for _, targetID := range rc.targets {
		if err := c.repo.MarkSuperseded(ctx, tx, tenantID, targetID); err != nil {
			if errors.Is(err, infrastructure.ErrScheduledSupersedeConflict) {
				return errSupersedeConflict
			}
			return fmt.Errorf("release: supersede %s: %w", targetID, err)
		}
	}

	// --- source CAS: approved|scheduled → published ------------------------
	from := docsdomain.DocumentStatus(doc.Status)
	if err := docsdomain.CanTransitionDocumentStatus(from, docsdomain.DocStatusPublished); err != nil {
		return err
	}
	reviewDue := domain.ResolveActualReviewDueAt(rc.pre.ReviewPlan, rc.now, rc.pre.ReviewDays)
	res, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET status                 = 'published',
		       effective_from         = $1,
		       review_due_at          = $2,
		       superseded_document_id = NULLIF($3, '')::uuid,
		       revision_version       = revision_version + 1
		 WHERE id        = $4::uuid
		   AND tenant_id = $5::uuid
		   AND status    IN ('approved','scheduled')
		   AND revision_version = $6`,
		rc.now, nullableTime(reviewDue), firstSupersedeTarget(rc.targets),
		gen.DocumentID, tenantID, doc.RevisionVersion,
	)
	if err != nil {
		return fmt.Errorf("release: publish document: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("release: publish rows affected: %w", err)
	}
	if affected == 0 {
		// Exactly-one-winner: another evaluation released this generation
		// between our read and our write. Roll back the supersede work by
		// returning the no-op through a rollback-free path is NOT possible
		// here (the supersede UPDATEs are in this tx), so fail the tx: the
		// winner already did the whole job.
		return errLostReleaseCAS
	}

	// --- durable release fact ---------------------------------------------
	if _, err := tx.ExecContext(ctx, `
		UPDATE release_generations
		   SET released_at       = $1,
		       hold_reason       = NULL,
		       hold_detail       = NULL,
		       last_evaluated_at = $1,
		       updated_at        = now()
		 WHERE id = $2::uuid`,
		rc.now, gen.ID,
	); err != nil {
		return fmt.Errorf("release: stamp released_at: %w", err)
	}

	// --- events, exactly once by construction ------------------------------
	if err := c.emitReleaseEvents(ctx, tx, rc); err != nil {
		return err
	}

	effectiveFrom := rc.now
	rc.out.Action = domain.ReleaseActionRelease
	rc.out.Hold = ""
	rc.out.EffectiveFrom = &effectiveFrom
	rc.out.SupersededDocumentIDs = rc.targets
	return nil
}

// errLostReleaseCAS is returned when the source CAS affects zero rows. It
// rolls the transaction back (discarding any supersede work) — the concurrent
// winner already performed the complete release.
var errLostReleaseCAS = errors.New("approval: release CAS lost to a concurrent evaluation")

// IsLostReleaseCAS reports whether err is the benign "another evaluation won"
// outcome. Callers (the River worker) treat it as success.
func IsLostReleaseCAS(err error) bool { return errors.Is(err, errLostReleaseCAS) }

func firstSupersedeTarget(ordered []string) string {
	if len(ordered) == 0 {
		return ""
	}
	return ordered[0]
}

// emitReleaseEvents writes ADR 0085's "exactly one governance event and one
// lifecycle event per affected transition": the published source is one
// transition, and EACH superseded target is another.
func (c *ReleaseCoordinator) emitReleaseEvents(ctx context.Context, tx *sql.Tx, rc releaseDecisionContext) error {
	gen, superseded, now := rc.gen, rc.targets, rc.now

	// Causal identities: every release event's ACTOR is the system principal,
	// but the humans who caused it are recorded in the payload (ADR 0085).
	causal := map[string]any{}
	if gen.FinalApproverID.Valid {
		causal["final_approver_id"] = gen.FinalApproverID.String
	}
	if gen.SubmittedBy.Valid {
		causal["submitted_by"] = gen.SubmittedBy.String
	}

	payload := map[string]any{
		"release_generation_id": gen.ID,
		"approval_instance_id":  gen.ApprovalInstanceID,
		"revision_id":           gen.RevisionID,
		"revision_version":      gen.RevisionVersion,
		"frozen_content_hash":   gen.FrozenContentHash,
		"effective_from":        now.Format(time.RFC3339),
		"trigger":               "release_coordinator",
	}
	for k, v := range causal {
		payload[k] = v
	}
	if len(superseded) > 0 {
		payload["superseded_document_ids"] = superseded
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("release: marshal event payload: %w", err)
	}
	if err := c.emitter.Emit(ctx, tx, GovernanceEvent{
		TenantID:     gen.TenantID,
		EventType:    EventTypeDocumentPublished,
		ActorUserID:  ReleaseSystemPrincipal,
		ResourceType: "document",
		ResourceID:   gen.DocumentID,
		Reason:       "release coordinator: approval-driven publication",
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}); err != nil {
		return fmt.Errorf("release: emit governance event: %w", err)
	}

	// Each supersession is a governed transition of ITS OWN document, so it
	// gets its own governance event — the source's published event names the
	// targets but is not the audit record for them.
	for _, targetID := range superseded {
		targetPayload := map[string]any{
			"release_generation_id":   gen.ID,
			"approval_instance_id":    gen.ApprovalInstanceID,
			"superseding_document_id": gen.DocumentID,
			"superseded_document_id":  targetID,
			"effective_from":          now.Format(time.RFC3339),
			"trigger":                 "release_coordinator",
		}
		for k, v := range causal {
			targetPayload[k] = v
		}
		targetBytes, err := json.Marshal(targetPayload)
		if err != nil {
			return fmt.Errorf("release: marshal superseded event payload for %s: %w", targetID, err)
		}
		if err := c.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     gen.TenantID,
			EventType:    EventTypeDocumentSuperseded,
			ActorUserID:  ReleaseSystemPrincipal,
			ResourceType: "document",
			ResourceID:   targetID,
			Reason:       "release coordinator: superseded by approval-driven publication",
			PayloadJSON:  json.RawMessage(targetBytes),
			OccurredAt:   now,
		}); err != nil {
			return fmt.Errorf("release: emit superseded governance event for %s: %w", targetID, err)
		}
	}

	if c.lifecycleEnqueuer == nil {
		return nil
	}
	if err := c.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, docsdomain.LifecycleEventArgs{
		EventID:              releaseEventID(gen, docsdomain.EventTypeDocumentPublished, gen.DocumentID),
		TenantID:             gen.TenantID,
		EventType:            docsdomain.EventTypeDocumentPublished,
		ResourceType:         "document",
		ResourceID:           gen.DocumentID,
		ControlledDocumentID: rc.pre.ControlledID,
		OccurredAt:           now,
	}); err != nil {
		return fmt.Errorf("release: enqueue published lifecycle event: %w", err)
	}
	for _, targetID := range superseded {
		// The TARGET's own controlled document, captured when its row was
		// locked — a plan-named cross-document target belongs to a different
		// controlled document than the source, and the fanout addresses
		// readers by controlled document. "" when the target has none.
		if err := c.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, docsdomain.LifecycleEventArgs{
			EventID:              releaseEventID(gen, docsdomain.EventTypeDocumentSuperseded, targetID),
			TenantID:             gen.TenantID,
			EventType:            docsdomain.EventTypeDocumentSuperseded,
			ResourceType:         "document",
			ResourceID:           targetID,
			ControlledDocumentID: rc.targetControlledIDs[targetID],
			OccurredAt:           now,
		}); err != nil {
			return fmt.Errorf("release: enqueue superseded lifecycle event for %s: %w", targetID, err)
		}
	}
	return nil
}

// releaseEventID derives a deterministic event id from (generation,
// transition, affected document) so a retried release re-uses the same
// idempotency key instead of fanning out a duplicate notification.
func releaseEventID(gen releaseGenerationRow, transition, affectedDocumentID string) string {
	return uuid.NewSHA1(releaseEventIDNamespace,
		[]byte("release:"+gen.ID+":"+transition+":"+affectedDocumentID)).String()
}

// documentReleaseState is the release-relevant slice of a documents row, read
// either unlocked (the decision pre-pass) or under the row lock.
//
// ControlledID is the row's OWN controlled document. It is read per row (not
// inherited from the source) because a plan-named supersession target may live
// under a DIFFERENT controlled document than the document being released — its
// lifecycle event has to carry its own.
type documentReleaseState struct {
	Status          string
	PlannedFrom     *time.Time
	EffectiveTo     *time.Time
	SupersedeTarget string
	ControlledID    string
	RevisionVersion int
}

const documentReleaseColumns = `status, planned_effective_from, effective_to,
	       superseded_document_id::text, controlled_document_id::text, revision_version`

func scanDocumentReleaseState(row interface{ Scan(...any) error }) (documentReleaseState, error) {
	var st documentReleaseState
	var planned, expiry sql.NullTime
	var target, controlled sql.NullString
	if err := row.Scan(&st.Status, &planned, &expiry, &target, &controlled, &st.RevisionVersion); err != nil {
		return documentReleaseState{}, err
	}
	st.ControlledID = controlled.String
	if planned.Valid {
		t := planned.Time
		st.PlannedFrom = &t
	}
	if expiry.Valid {
		t := expiry.Time
		st.EffectiveTo = &t
	}
	st.SupersedeTarget = target.String
	return st, nil
}

// readDocumentForRelease reads the plan fields the predicate needs WITHOUT
// locking. The coordinator decides on this snapshot, then locks the full
// document set in sorted id order and re-decides (see evaluateTx).
func readDocumentForRelease(ctx context.Context, tx db.Tx, tenantID, documentID string) (documentReleaseState, error) {
	st, err := scanDocumentReleaseState(tx.QueryRowContext(ctx, `
		SELECT `+documentReleaseColumns+`
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		documentID, tenantID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return documentReleaseState{}, fmt.Errorf("release: document %s not found", documentID)
	}
	if err != nil {
		return documentReleaseState{}, fmt.Errorf("release: read document: %w", err)
	}
	return st, nil
}

// lockDocumentForRelease takes ONE document row lock. Callers must invoke it in
// ascending document-id order across the whole lock set.
func lockDocumentForRelease(ctx context.Context, tx db.Tx, tenantID, documentID string) (documentReleaseState, error) {
	st, err := scanDocumentReleaseState(tx.QueryRowContext(ctx, `
		SELECT `+documentReleaseColumns+`
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid
		 FOR UPDATE`,
		documentID, tenantID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return documentReleaseState{}, fmt.Errorf("release: document %s not found", documentID)
	}
	if err != nil {
		return documentReleaseState{}, fmt.Errorf("release: lock document: %w", err)
	}
	return st, nil
}

// touchEvaluatedTx records the evaluation outcome on the generation. An empty
// reason clears the hold.
func touchEvaluatedTx(ctx context.Context, tx db.Tx, generationID string, reason domain.ReleaseHoldReason) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE release_generations
		   SET hold_reason       = NULLIF($1, ''),
		       last_evaluated_at = now(),
		       updated_at        = now()
		 WHERE id = $2::uuid`,
		string(reason), generationID,
	)
	if err != nil {
		return fmt.Errorf("release: record evaluation outcome: %w", err)
	}
	return nil
}

// recordHold writes a hold in its own transaction. Used after a rollback, when
// the reason for the rollback would otherwise be lost.
func (c *ReleaseCoordinator) recordHold(ctx context.Context, runner db.TxRunner, key domain.ReleaseGenerationKey, reason domain.ReleaseHoldReason, detail string) error {
	return runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(ctx, tx, key.TenantID); err != nil {
			return fmt.Errorf("release hold: seed tenant: %w", err)
		}
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return fmt.Errorf("release hold: bypass: %w", err)
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE release_generations
			   SET hold_reason       = $1,
			       hold_detail       = NULLIF($2, ''),
			       last_evaluated_at = now(),
			       updated_at        = now()
			 WHERE tenant_id            = $3::uuid
			   AND document_id          = $4::uuid
			   AND approval_instance_id = $5::uuid
			   AND revision_id          = $6::uuid
			   AND revision_version     = $7
			   AND frozen_content_hash  = $8`,
			string(reason), detail, key.TenantID, key.DocumentID,
			key.ApprovalInstanceID, key.RevisionID, key.RevisionVersion, key.FrozenContentHash,
		)
		if err != nil {
			return fmt.Errorf("release hold: update generation: %w", err)
		}
		return nil
	})
}
