package application

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/platform/db"
)

// FastForwardService implements the "Aprovar já" gesture (R5, unit 2.3 G3):
// one caller action records a `ready` review verdict AND an approve signoff
// as two separate ledger writes in ONE transaction. It composes the two
// existing services rather than duplicating their logic — recordVerdictInTx
// and recordSignoffInTx are the same tx-scoped cores RecordVerdict and
// RecordSignoff call individually, extracted in S2 for exactly this reuse.
type FastForwardService struct {
	verdicts  *ReviewVerdictService
	decisions *DecisionService
}

// newFastForwardService constructs a FastForwardService composing the given
// review-verdict and decision services. Both must share the same underlying
// infrastructure.ApprovalRepository — the same repo row locked in the verdict
// leg's LoadInstance is the row the signoff leg then re-derives its next
// stage against, inside the same transaction.
func newFastForwardService(verdicts *ReviewVerdictService, decisions *DecisionService) *FastForwardService {
	return &FastForwardService{verdicts: verdicts, decisions: decisions}
}

// FastForwardRequest carries all inputs for RecordFastForward. Field names
// and types mirror ReviewVerdictRequest / SignoffRequest so the mapping into
// each leg's request is mechanical.
type FastForwardRequest struct {
	TenantID                string
	InstanceID              string
	StageInstanceID         string
	ActorUserID             string
	Comment                 string
	ExpectedRevisionVersion int
	SignatureMethod         string
	SignaturePayload        map[string]any
	ContentHash             string
}

// FastForwardResult carries the outcome of both ledger writes.
type FastForwardResult struct {
	Verdict ReviewVerdictResult
	Signoff SignoffResult
}

// RecordFastForward records a `ready` verdict against the active review
// stage identified by req.StageInstanceID and, iff that verdict completes
// the stage (quorum satisfied) onto a now-active approval-kind stage the
// actor is eligible on, an approve signoff against that next stage — both in
// one transaction. Fails closed: any error from either leg rolls back both
// writes, so a partial fast-forward (verdict recorded but no signoff, or
// vice versa) never persists.
func (s *FastForwardService) RecordFastForward(ctx context.Context, runner db.TxRunner, req FastForwardRequest) (FastForwardResult, error) {
	// H-PRE-1: ONE off-tx actor display-name preload, shared by both legs —
	// mirrors RecordVerdict/RecordSignoff's own off-tx preload, done once here
	// instead of twice since both legs need the identical value.
	actorDisplayName, err := s.verdicts.repo.LoadActorDisplayName(ctx, req.TenantID, req.ActorUserID)
	if err != nil {
		return FastForwardResult{}, fmt.Errorf("recordFastForward: lookup actor display name: %w", err)
	}

	var result FastForwardResult
	var eligibilityEvent *GovernanceEvent
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		verdictReq := ReviewVerdictRequest{
			TenantID:                req.TenantID,
			InstanceID:              req.InstanceID,
			StageInstanceID:         req.StageInstanceID,
			ActorUserID:             req.ActorUserID,
			Verdict:                 domain.VerdictReady,
			Comment:                 req.Comment,
			ExpectedRevisionVersion: req.ExpectedRevisionVersion,
		}
		verdictResult, verdictEligibilityEvent, txErr := s.verdicts.recordVerdictInTx(ctx, tx, verdictReq, actorDisplayName)
		if verdictEligibilityEvent != nil {
			eligibilityEvent = verdictEligibilityEvent
		}
		if txErr != nil {
			return txErr
		}
		result.Verdict = verdictResult

		if !verdictResult.StageCompleted {
			return domain.ErrFastForwardStageNotCompleted
		}
		if verdictResult.InstanceApproved || !verdictResult.FastForwardEligible || verdictResult.NextStageID == nil {
			return domain.ErrFastForwardNotEligible
		}

		// OCC: the verdict leg above already OCC-checked instance.RevisionVersion
		// against req.ExpectedRevisionVersion and does not itself mutate
		// RevisionVersion (only the terminal instance-approved / instance
		// reverted-to-draft branches touch the documents row's
		// revision_version — neither fires on the mid-route stage-advance
		// path this method requires, verdictResult.InstanceApproved == false
		// having just been checked above). So the signoff leg's OCC check
		// receives the SAME ExpectedRevisionVersion unchanged, not 0/skipped.
		signoffReq := SignoffRequest{
			TenantID:                req.TenantID,
			InstanceID:              req.InstanceID,
			StageInstanceID:         *verdictResult.NextStageID,
			ActorUserID:             req.ActorUserID,
			Decision:                domain.DecisionApprove,
			Comment:                 req.Comment,
			SignatureMethod:         req.SignatureMethod,
			SignaturePayload:        req.SignaturePayload,
			ContentFormData:         map[string]any{"_content_hash": req.ContentHash},
			ExpectedRevisionVersion: req.ExpectedRevisionVersion,
		}
		signoffResult, signoffEligibilityEvent, txErr := s.decisions.recordSignoffInTx(ctx, tx, signoffReq, actorDisplayName)
		if signoffEligibilityEvent != nil {
			eligibilityEvent = signoffEligibilityEvent
		}
		if txErr != nil {
			return txErr
		}
		result.Signoff = signoffResult
		return nil
	})

	if eligibilityEvent != nil {
		if emitErr := s.verdicts.emitEligibilityRejection(ctx, runner, req.TenantID, req.ActorUserID, *eligibilityEvent); emitErr != nil {
			return FastForwardResult{}, fmt.Errorf("recordFastForward: emit eligibility rejection: %w", emitErr)
		}
		return FastForwardResult{}, err
	}
	if err != nil {
		return FastForwardResult{}, err
	}
	return result, nil
}
