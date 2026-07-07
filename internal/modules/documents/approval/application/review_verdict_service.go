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
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
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
	StageCompleted   bool
	InstanceApproved bool // true when all stages complete (ready path only)
	ChangesRequested bool // true when the instance collapsed to changes_requested
}

// RecordVerdict records a ready/request_changes verdict for the given
// review-kind stage instance.
func (s *ReviewVerdictService) RecordVerdict(ctx context.Context, runner db.TxRunner, req ReviewVerdictRequest) (ReviewVerdictResult, error) {
	var result ReviewVerdictResult
	var eligibilityEvent *GovernanceEvent

	// H-PRE-1: resolve the actor display-name snapshot OFF the lock-holding tx.
	actorDisplayName, err := s.repo.LoadActorDisplayName(ctx, req.TenantID, req.ActorUserID)
	if err != nil {
		return ReviewVerdictResult{}, fmt.Errorf("recordVerdict: lookup actor display name: %w", err)
	}

	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		instance, err := s.repo.LoadInstance(ctx, tx, req.TenantID, req.InstanceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("recordVerdict: %w", infrastructure.ErrNoActiveInstance)
			}
			return fmt.Errorf("recordVerdict: load instance: %w", err)
		}
		if instance == nil {
			return infrastructure.ErrNoActiveInstance
		}
		if req.ExpectedRevisionVersion > 0 && req.ExpectedRevisionVersion != instance.RevisionVersion {
			return infrastructure.ErrStaleRevision
		}

		// Only in_progress instances accept new verdicts (changes_requested and
		// the three original terminal statuses all reject).
		if instance.Status != domain.InstanceInProgress {
			return infrastructure.ErrInstanceCompleted
		}

		areaCode, _, err := docapp.LoadDocumentAreaCode(ctx, tx, s.cdRead, req.TenantID, instance.DocumentID)
		if err != nil {
			return fmt.Errorf("recordVerdict: load document area: %w", err)
		}
		if err := authz.Require(ctx, tx, string(iamdomain.CapApprovalReview), areaCode); err != nil {
			return err
		}

		activeStage := instance.Active()
		if activeStage == nil {
			return domain.ErrNoActiveStage
		}
		if req.StageInstanceID == "" || activeStage.ID != req.StageInstanceID {
			return infrastructure.ErrStageNotActive
		}

		// Verdicts are only valid against review-kind stages; approval-kind
		// stages use signoffs. NewVerdict enforces this too, but checking here
		// first lets us return a clean error before any eligibility/SoD work.
		if activeStage.Kind != domain.StageKindReview {
			return domain.ErrVerdictWrongStageKind
		}

		if err := domain.CheckEligibility(req.ActorUserID, activeStage.EligibleActorIDs); err != nil {
			event := GovernanceEvent{
				TenantID:     req.TenantID,
				EventType:    EventTypeSignoffRejected,
				ActorUserID:  req.ActorUserID,
				ResourceType: "approval_instance",
				ResourceID:   req.InstanceID,
				Reason:       "not_eligible",
				OccurredAt:   s.clock.Now(),
			}
			eligibilityEvent = &event
			return err
		}

		// SoD: author cannot verdict their own submission (spec.md: "SoD blocks
		// self-verdict"). Actor-already-recorded-a-verdict-on-this-stage is
		// handled by InsertVerdict's idempotent-replay/conflict distinction
		// below (mirrors InsertSignoff — a second table-shape-specific
		// duplicate helper here would be redundant, not more correct).
		if instance.SubmittedBy == req.ActorUserID {
			return domain.ErrAuthorCannotSign
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
		})
		if err != nil {
			return fmt.Errorf("recordVerdict: build verdict: %w", err)
		}

		insertResult, err := s.repo.InsertVerdict(ctx, tx, *verdict)
		if err != nil {
			if errors.Is(err, infrastructure.ErrActorAlreadySigned) {
				return err
			}
			return fmt.Errorf("recordVerdict: insert verdict: %w", err)
		}
		if insertResult.WasReplay {
			result = ReviewVerdictResult{}
			return nil
		}

		switch req.Verdict {
		case domain.VerdictReady:
			allStageVerdicts, err := s.repo.LoadStageVerdicts(ctx, tx, req.TenantID, activeStage.ID)
			if err != nil {
				return fmt.Errorf("recordVerdict: load stage verdicts: %w", err)
			}
			approvals := verdictsAsApprovals(allStageVerdicts)

			currentEligible := activeStage.EligibleActorIDs
			if activeStage.OnEligibilityDriftSnapshot != domain.DriftKeepSnapshot {
				currentEligible, err = s.repo.ResolveEligibleActors(ctx, tx, req.TenantID, activeStage.AreaCodeSnapshot, activeStage.RequiredRoleSnapshot)
				if err != nil {
					return fmt.Errorf("recordVerdict: resolve current eligible actors: %w", err)
				}
			}
			drift := domain.ApplyEligibilityDrift(*activeStage, currentEligible)
			effectiveDenominator := drift.EffectiveDenominator
			outcome := drift.ForcedOutcome
			if outcome == domain.QuorumPending {
				outcome = domain.EvaluateQuorum(*activeStage, approvals, nil, effectiveDenominator)
				if outcome == domain.QuorumError {
					return domain.ErrEmptyEligiblePool
				}
			}

			if outcome == domain.QuorumApprovedStage {
				// Capture the completing stage's kind BEFORE AdvanceStage mutates
				// the in-memory stage slice (F5, plan.md task 5) — needed below to
				// decide whether this transition crosses the freeze boundary.
				completingStageKind := activeStage.Kind

				if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, activeStage.ID, domain.StageCompleted, domain.StageActive); err != nil {
					return fmt.Errorf("recordVerdict: complete stage: %w", err)
				}
				result.StageCompleted = true

				if err := instance.AdvanceStage(); err != nil {
					return fmt.Errorf("recordVerdict: advance stage: %w", err)
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
						return fmt.Errorf("recordVerdict: freeze: %w", err)
					}
				}

				if instance.Status == domain.InstanceApproved {
					if err := s.repo.UpdateInstanceStatus(ctx, tx, req.TenantID, req.InstanceID,
						domain.InstanceApproved, domain.InstanceInProgress, &now); err != nil {
						return fmt.Errorf("recordVerdict: complete instance: %w", err)
					}
					if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
						return err
					}
					if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusUnderReview, docsdomain.DocStatusApproved); err != nil {
						return err
					}
					res, err := tx.ExecContext(ctx, `
						UPDATE documents
						   SET status           = 'approved',
						       revision_version = revision_version + 1
						 WHERE id        = $1
						   AND tenant_id = $2
						   AND status    = 'under_review'`,
						instance.DocumentID, req.TenantID,
					)
					if err != nil {
						return fmt.Errorf("recordVerdict: approve document: %w", err)
					}
					rows, err := res.RowsAffected()
					if err != nil {
						return fmt.Errorf("recordVerdict: approve document rows affected: %w", err)
					}
					if rows == 0 {
						return infrastructure.ErrStaleRevision
					}
					result.InstanceApproved = true

					if s.lifecycleEnqueuer != nil {
						largs := docsdomain.LifecycleEventArgs{
							EventID:      uuid.NewString(),
							TenantID:     req.TenantID,
							EventType:    docsdomain.EventTypeDocumentApproved,
							ResourceType: "approval_instance",
							ResourceID:   req.InstanceID,
							SubmittedBy:  instance.SubmittedBy,
							OccurredAt:   now,
						}
						if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
							return fmt.Errorf("recordVerdict: enqueue lifecycle event: %w", err)
						}
					}
				} else {
					nextStage := instance.Active()
					if nextStage != nil {
						if err := s.repo.UpdateStageStatus(ctx, tx, req.TenantID, nextStage.ID, domain.StageActive, domain.StagePending); err != nil {
							return fmt.Errorf("recordVerdict: activate next stage: %w", err)
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
				return fmt.Errorf("recordVerdict: set changes_requested: %w", err)
			}
			result.ChangesRequested = true

			// SET LOCAL cancel GUC authorises the under_review -> draft edge in
			// the document-transition trigger (same gate used by reject/cancel).
			if _, err := tx.ExecContext(ctx,
				`SELECT set_config('metaldocs.cancel_in_progress', $1, true)`,
				instance.ID,
			); err != nil {
				return fmt.Errorf("recordVerdict: set cancel GUC: %w", err)
			}
			if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err != nil {
				return err
			}
			if err := docsdomain.CanTransitionDocumentStatus(docsdomain.DocStatusUnderReview, docsdomain.DocStatusDraft); err != nil {
				return err
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
				return fmt.Errorf("recordVerdict: revert document to draft: %w", err)
			}
			rows, err := res.RowsAffected()
			if err != nil {
				return fmt.Errorf("recordVerdict: revert document rows affected: %w", err)
			}
			if rows == 0 {
				return infrastructure.ErrStaleRevision
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
		}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return fmt.Errorf("recordVerdict: marshal event payload: %w", err)
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
			return fmt.Errorf("recordVerdict: emit event: %w", err)
		}

		return nil
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
