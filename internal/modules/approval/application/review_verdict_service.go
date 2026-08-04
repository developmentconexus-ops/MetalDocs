package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// ErrVerdictWrongStageKind mirrors domain.ErrVerdictWrongStageKind for callers
// that only import the application package.
var ErrVerdictWrongStageKind = domain.ErrVerdictWrongStageKind

// ReviewVerdictService records `ready`/`request_changes` verdicts against
// review-kind stages (F4, design spec.md §2). Mirrors DecisionService.RecordSignoff's
// structure exactly: off-tx actor-name lookup, in-tx FOR UPDATE load, OCC,
// area-grade authz, eligibility, SoD, idempotent insert, quorum evaluation
// (ready) or instance-collapsing transition (request_changes).
type ReviewVerdictService struct {
	repo              infrastructure.ApprovalRepository
	emitter           EventEmitter
	clock             Clock
	cdRead            controlleddocumentsdomain.CDFieldReader
	lifecycleEnqueuer docsdomain.LifecycleEventEnqueuer
	// releaseRecorder is the SAME terminal-approval seam DecisionService uses.
	// Before ADR 0085 this route reached terminal approval without ever
	// pinning the document, so a document approved via a review verdict never
	// materialized and could never be published (F-QA4-14). The two routes now
	// share one terminal-approval implementation, injected from one call site.
	releaseRecorder TerminalApprovalReleaseRecorder
}

// WithReleaseRecorder wires the ADR 0085 terminal-approval seam (pin +
// approval fact + coordinator evaluation). Mandatory for the terminal-approval
// path — see the releaseRecorder field comment (F-QA4-14).
func (s *ReviewVerdictService) WithReleaseRecorder(recorder TerminalApprovalReleaseRecorder) *ReviewVerdictService {
	s.releaseRecorder = recorder
	return s
}

// WithCDFieldReader wires the controlleddocuments read-port used to resolve a
// document's controlled-document area in the area-grade authz check.
func (s *ReviewVerdictService) WithCDFieldReader(r controlleddocumentsdomain.CDFieldReader) *ReviewVerdictService {
	s.cdRead = r
	return s
}

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer. Note:
// request_changes is a non-terminal transition (changes_requested, not
// approved/rejected) so no lifecycle event is enqueued for it — mirrors the
// spec's Non-goals; only kept here for interface parity with the other
// services in case a future terminal path is added.
func (s *ReviewVerdictService) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *ReviewVerdictService {
	s.lifecycleEnqueuer = e
	return s
}

// ReviewVerdictRequest carries all inputs for RecordVerdict.
type ReviewVerdictRequest struct {
	TenantID                string
	InstanceID              string
	StageInstanceID         string
	ActorUserID             string
	Verdict                 domain.Verdict
	Comment                 string
	ExpectedRevisionVersion int
}

// ReviewVerdictResult is returned by RecordVerdict.
type ReviewVerdictResult struct {
	// VerdictID is the approval_review_verdicts row id persisted by this call.
	// On the DB-level idempotent replay branch (ON CONFLICT) it carries the
	// ORIGINAL row id, so a retrying caller always sees the same identifier
	// (F-QA4-7 class, mirrors SignoffResult.SignoffID).
	VerdictID        string
	StageCompleted   bool
	InstanceApproved bool // true when all stages complete (ready path only)
	ChangesRequested bool // true when the instance collapsed to changes_requested
	// FastForwardEligible (R5, unit 2.3 G3) is true iff this verdict completed
	// the review stage (quorum satisfied) AND the actor is eligible on the
	// now-active approval-kind stage. A hint only — never an error path.
	FastForwardEligible bool
	// NextStageID is the now-active stage's id, set only when
	// FastForwardEligible is true.
	NextStageID *string
}

// RecordVerdict records a ready/request_changes verdict for the given
// review-kind stage instance.
func (s *ReviewVerdictService) RecordVerdict(ctx context.Context, runner db.TxRunner, req ReviewVerdictRequest) (ReviewVerdictResult, error) {
	// H-PRE-1: resolve the actor display-name snapshot OFF the lock-holding tx.
	actorDisplayName, err := s.repo.LoadActorDisplayName(ctx, req.TenantID, req.ActorUserID)
	if err != nil {
		return ReviewVerdictResult{}, fmt.Errorf("recordVerdict: lookup actor display name: %w", err)
	}

	var result ReviewVerdictResult
	var eligibilityEvent *GovernanceEvent
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		var txErr error
		result, eligibilityEvent, txErr = s.recordVerdictInTx(ctx, tx, req, actorDisplayName)
		return txErr
	})

	if eligibilityEvent != nil {
		if emitErr := s.emitEligibilityRejection(ctx, runner, req.TenantID, req.ActorUserID, *eligibilityEvent); emitErr != nil {
			return ReviewVerdictResult{}, fmt.Errorf("recordVerdict: emit eligibility rejection: %w", emitErr)
		}
		return ReviewVerdictResult{}, err
	}
	if err != nil {
		return ReviewVerdictResult{}, err
	}
	return result, nil
}

// recordVerdictInTx is the tx-scoped core of RecordVerdict, extracted so a
// fast-forward flow (unit 2.3 G3) can run it inside a transaction it shares
// with DecisionService.recordSignoffInTx. It does not own commit/rollback and
// must not call runner itself. actorDisplayName is the off-tx preload from
// RecordVerdict. The eligibility-rejection event (when non-nil) must be
// emitted by the caller AFTER the tx this ran in has closed — mirrors
// RecordVerdict's original post-tx emitEligibilityRejection call.
func (s *ReviewVerdictService) recordVerdictInTx(ctx context.Context, tx *sql.Tx, req ReviewVerdictRequest, actorDisplayName string) (ReviewVerdictResult, *GovernanceEvent, error) {
	ctx = authz.WithCapCache(ctx)
	var result ReviewVerdictResult

	instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: %w", infrastructure.ErrNoActiveInstance)
		}
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: load instance: %w", err)
	}
	if instance == nil {
		return ReviewVerdictResult{}, nil, infrastructure.ErrNoActiveInstance
	}
	if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
		return ReviewVerdictResult{}, nil, infrastructure.ErrStaleRevision
	}

	// Only in_progress instances accept new verdicts (changes_requested and
	// the three original terminal statuses all reject).
	if instance.Status != domain.InstanceInProgress {
		return ReviewVerdictResult{}, nil, infrastructure.ErrInstanceCompleted
	}

	areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, instance.DocumentID)
	if err != nil {
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: load document area: %w", err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapApprovalReview), areaCode); err != nil {
		return ReviewVerdictResult{}, nil, err
	}

	activeStage := instance.Active()
	if activeStage == nil {
		return ReviewVerdictResult{}, nil, domain.ErrNoActiveStage
	}
	if req.StageInstanceID == "" || activeStage.ID != req.StageInstanceID {
		return ReviewVerdictResult{}, nil, infrastructure.ErrStageNotActive
	}

	// Verdicts are valid on review stages (both values) and on approval
	// stages only as request_changes (R3/G2: approval powers include the
	// power to converse/return without signing; `ready` on an approval stage
	// would bypass the e-signature ceremony). Checked here first for a clean
	// error before eligibility/SoD work; NewVerdict re-enforces authoritatively.
	switch activeStage.Kind {
	case domain.StageKindReview:
	case domain.StageKindApproval:
		if req.Verdict == domain.VerdictReady {
			return ReviewVerdictResult{}, nil, domain.ErrVerdictReadyOnApprovalStage
		}
	default:
		return ReviewVerdictResult{}, nil, domain.ErrVerdictWrongStageKind
	}

	// Eligibility, widened by any active delegation (F9/ADR 0077) — mirrors
	// RecordSignoff's composition exactly: same domain.ResolveEligibleIdentity
	// helper, same underlying domain.CheckEligibility predicate, no parallel
	// rule.
	delegations, err := s.repo.LoadActiveDelegationsFor(ctx, tx, req.TenantID, req.ActorUserID, s.clock.Now())
	if err != nil {
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: load active delegations: %w", err)
	}
	onBehalfOf, err := domain.ResolveEligibleIdentity(req.ActorUserID, activeStage.EligibleActorIDs, delegations)
	if err != nil {
		event := GovernanceEvent{
			TenantID:     req.TenantID,
			EventType:    EventTypeSignoffRejected,
			ActorUserID:  req.ActorUserID,
			ResourceType: "approval_instance",
			ResourceID:   req.InstanceID,
			Reason:       "not_eligible",
			OccurredAt:   s.clock.Now(),
		}
		return ReviewVerdictResult{}, &event, err
	}

	// SoD: author cannot verdict their own submission (spec.md: "SoD blocks
	// self-verdict"); a delegate cannot verdict on behalf of a delegator who
	// is the author (F9/ADR 0077). This calls the SAME domain.CheckSoD
	// predicate RecordSignoff calls (F7 unification) — priorSignoffs is nil
	// because the cross-stage-reuse clause needs a []Signoff-shaped
	// prior-record source that doesn't apply to review verdicts;
	// InsertVerdict's idempotent-replay/conflict distinction below already
	// handles actor-already-recorded-a-verdict-on-this-stage (mirrors
	// InsertSignoff).
	if err := domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, onBehalfOf, nil); err != nil {
		return ReviewVerdictResult{}, nil, err
	}

	now := s.clock.Now()
	verdict, err := domain.NewVerdict(domain.VerdictParams{
		ID:                       uuid.New().String(),
		ApprovalInstanceID:       req.InstanceID,
		StageInstanceID:          activeStage.ID,
		StageKind:                activeStage.Kind,
		ActorUserID:              req.ActorUserID,
		ActorTenantID:            req.TenantID,
		Verdict:                  req.Verdict,
		Comment:                  req.Comment,
		VerdictAt:                now,
		ActorDisplayNameSnapshot: actorDisplayName,
		OnBehalfOfUserID:         onBehalfOf,
	})
	if err != nil {
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: build verdict: %w", err)
	}

	insertResult, err := s.repo.InsertVerdict(ctx, tx, *verdict)
	if err != nil {
		if errors.Is(err, infrastructure.ErrActorAlreadySigned) {
			return ReviewVerdictResult{}, nil, err
		}
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: insert verdict: %w", err)
	}
	// Carry the persisted approval_review_verdicts id out of the tx. On the ON
	// CONFLICT branch InsertVerdict returns the ORIGINAL row's id, so a retrying
	// caller sees the same identifier instead of an empty string.
	result.VerdictID = insertResult.ID
	if insertResult.WasReplay {
		// Idempotent replay: the stage is not advanced again, so the result
		// carries only VerdictID.
		return result, nil, nil
	}

	switch req.Verdict {
	case domain.VerdictReady:
		allStageVerdicts, err := s.repo.LoadStageVerdicts(ctx, tx, req.TenantID, activeStage.ID)
		if err != nil {
			return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: load stage verdicts: %w", err)
		}
		approvals := verdictsAsApprovals(allStageVerdicts)

		currentEligible := activeStage.EligibleActorIDs
		if activeStage.OnEligibilityDriftSnapshot != domain.DriftKeepSnapshot {
			currentEligible, err = resolveCurrentEligibleForDrift(ctx, tx, s.repo, s.cdRead, req.TenantID, *activeStage, instance.Subject)
			if err != nil {
				return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: resolve current eligible actors: %w", err)
			}
		}
		drift := domain.ApplyEligibilityDrift(*activeStage, currentEligible)
		effectiveDenominator := drift.EffectiveDenominator
		outcome := drift.ForcedOutcome
		if outcome == domain.QuorumPending {
			outcome = domain.EvaluateQuorum(*activeStage, approvals, nil, effectiveDenominator)
			if outcome == domain.QuorumError {
				return ReviewVerdictResult{}, nil, domain.ErrEmptyEligiblePool
			}
		}

		if outcome == domain.QuorumApprovedStage {
			// Capture the completing stage's kind BEFORE AdvanceStage mutates
			// the in-memory stage slice (F5, plan.md task 5) — needed below to
			// decide whether this transition crosses the freeze boundary.
			completingStageKind := activeStage.Kind

			if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageCompleted, domain.StageActive); err != nil {
				return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: complete stage: %w", err)
			}
			result.StageCompleted = true

			if err := instance.AdvanceStage(); err != nil {
				return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: advance stage: %w", err)
			}

			// Freeze boundary (F5, design spec.md §2.2 "W2 core"): the last
			// review-kind stage completing — whether the instance is now fully
			// approved (no more stages) or has advanced into an approval-kind
			// stage — crosses the point past which the document content is
			// immutable. This must run BEFORE the (now-removed, W10)
			// unresolved-comments gate below, since freeze is the sole
			// remaining enforcement of that concern.
			nextActive := instance.Active()
			crossesFreezeBoundary := completingStageKind == domain.StageKindReview &&
				(nextActive == nil || nextActive.Kind == domain.StageKindApproval)
			if crossesFreezeBoundary {
				if err := executeFreeze(ctx, tx, s.repo, req.TenantID, instance); err != nil {
					return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: freeze: %w", err)
				}
			}

			if instance.Status == domain.InstanceApproved {
				// F-QA4-14 / ADR 0085: this route is identical to the signoff
				// and ADR 0087 auto-approve routes — same instance CAS, same
				// document.edit assertion, same Pin, same approval fact, same
				// coordinator evaluation, same lifecycle event, same tx.
				if err := completeDocumentTerminalApproval(ctx, tx, documentTerminalApprovalPorts{
					repo:              s.repo,
					releaseRecorder:   s.releaseRecorder,
					lifecycleEnqueuer: s.lifecycleEnqueuer,
					serviceName:       "review verdict service",
				}, documentTerminalApprovalInput{
					TenantID:          req.TenantID,
					InstanceID:        req.InstanceID,
					DocumentID:        instance.DocumentID,
					AreaCode:          areaCode,
					RevisionVersion:   instance.RevisionVersion,
					FrozenContentHash: derefString(instance.FrozenContentHash),
					FinalApproverID:   req.ActorUserID,
					SubmittedBy:       instance.SubmittedBy,
					FromStatus:        domain.InstanceInProgress,
					Now:               now,
				}); err != nil {
					return ReviewVerdictResult{}, nil, err
				}
				result.InstanceApproved = true
			} else {
				nextStage := instance.Active()
				if nextStage != nil {
					if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, nextStage.ID, domain.StageActive, domain.StagePending); err != nil {
						return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: activate next stage: %w", err)
					}

					// R5 (unit 2.3 G3) fast-forward eligibility probe: only
					// offered when the now-active stage is approval-kind and
					// the actor is eligible on it. A failed eligibility
					// resolve here is a hint miss, not an error — the
					// verdict itself already succeeded above.
					if nextStage.Kind == domain.StageKindApproval {
						if _, err := domain.ResolveEligibleIdentity(req.ActorUserID, nextStage.EligibleActorIDs, delegations); err == nil {
							result.FastForwardEligible = true
							nextStageID := nextStage.ID
							result.NextStageID = &nextStageID
						}
					}
				}
			}
		}

	case domain.VerdictRequestChanges:
		// No quorum needed — a single request_changes verdict collapses the
		// instance immediately (spec.md §2 consumer contract). The reviewer's
		// comment is already durably recorded on the approval_review_verdicts
		// row inserted above; it is NOT also written to
		// approval_instances.cancel_reason — that column is reserved for
		// actual cancellations (CancelInstance), and reusing it here would
		// silently overload an audit field's meaning with an unrelated
		// concern (duplicate data, misleading field on a changes_requested
		// instance).
		if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
			domain.InstanceChangesRequested, domain.InstanceInProgress, &now); err != nil {
			return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: set changes_requested: %w", err)
		}
		result.ChangesRequested = true

		// SET LOCAL cancel GUC authorises the under_review -> draft edge in
		// the document-transition trigger (same gate used by reject/cancel).
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.cancel_in_progress', $1, true)`,
			instance.ID,
		); err != nil {
			return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: set cancel GUC: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
			return ReviewVerdictResult{}, nil, err
		}
		if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusUnderReview, docsdomain.DocStatusDraft); err != nil {
			return ReviewVerdictResult{}, nil, err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE documents
			   SET status           = 'draft',
			       revision_version = revision_version + 1
			 WHERE id        = $1
			   AND tenant_id = $2
			   AND status    = 'under_review'`,
			instance.DocumentID, req.TenantID,
		)
		if err != nil {
			return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: revert document to draft: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: revert document rows affected: %w", err)
		}
		if rows == 0 {
			return ReviewVerdictResult{}, nil, infrastructure.ErrStaleRevision
		}
	}

	// Emit governance event.
	eventType := EventTypeReviewVerdictRecorded
	if req.Verdict == domain.VerdictRequestChanges {
		eventType = EventTypeReviewChangesRequested
	}
	payloadMap := map[string]any{
		"instance_id":       req.InstanceID,
		"stage_instance_id": activeStage.ID,
		"verdict":           req.Verdict,
		"on_behalf_of":      onBehalfOf,
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: marshal event payload: %w", err)
	}
	event := GovernanceEvent{
		TenantID:     req.TenantID,
		EventType:    eventType,
		ActorUserID:  req.ActorUserID,
		ResourceType: "approval_instance",
		ResourceID:   req.InstanceID,
		Reason:       req.Comment,
		PayloadJSON:  json.RawMessage(payloadBytes),
		OccurredAt:   now,
	}
	if err := s.emitter.Emit(ctx, tx, event); err != nil {
		return ReviewVerdictResult{}, nil, fmt.Errorf("recordVerdict: emit event: %w", err)
	}

	return result, nil, nil
}

func (s *ReviewVerdictService) emitEligibilityRejection(ctx context.Context, runner db.TxRunner, tenantID, actorID string, event GovernanceEvent) error {
	return runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)
		if err := s.emitter.Emit(ctx, tx, event); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}
		return nil
	})
}

// zeroContentHash is a placeholder 64-lowercase-hex value satisfying
// NewSignoff's ContentHash format validation. verdictsAsApprovals only needs
// the ActorUserID field for EvaluateQuorum's eligibility filtering; verdicts
// carry no content-hash concept of their own (spec.md Non-goals — no
// e-signature semantics), so this is a format-only filler, never persisted.
const zeroContentHash = "0000000000000000000000000000000000000000000000000000000000000000"

// verdictsAsApprovals adapts ready-verdict rows into the []domain.Signoff-shaped
// approvals bucket EvaluateQuorum expects — reusing EvaluateQuorum's counting
// function exactly (spec.md Non-goals: no widened storage, just fn reuse).
// Only ActorUserID is read by EvaluateQuorum's filtering, so a minimal Signoff
// is constructed per verdict.
func verdictsAsApprovals(verdicts []domain.ReviewVerdict) []domain.Signoff {
	var approvals []domain.Signoff
	for _, v := range verdicts {
		if v.Verdict() != domain.VerdictReady {
			continue
		}
		s, err := domain.NewSignoff(domain.SignoffParams{
			ID:                 v.ID(),
			ApprovalInstanceID: v.ApprovalInstanceID(),
			StageInstanceID:    v.StageInstanceID(),
			ActorUserID:        v.ActorUserID(),
			ActorTenantID:      v.ActorTenantID(),
			Decision:           domain.DecisionApprove,
			SignedAt:           v.VerdictAt(),
			SignatureMethod:    "review_verdict",
			ContentHash:        zeroContentHash,
		})
		if err != nil {
			// Should never happen — all required fields are populated above.
			continue
		}
		approvals = append(approvals, *s)
	}
	return approvals
}

// newReviewVerdictService constructs a ReviewVerdictService.
func newReviewVerdictService(repo infrastructure.ApprovalRepository, emitter EventEmitter, clock Clock) *ReviewVerdictService {
	return &ReviewVerdictService{repo: repo, emitter: emitter, clock: clock}
}
