package application

// fast_forward_service_test.go — application-layer unit tests for
// FastForwardService.RecordFastForward (R5, unit 2.3 G3, slice S3), reusing
// the shared fake harness defined in decision_service_test.go
// (fakeDecisionRepo, decisionTestConn, newDecisionTestDB, newTxRunner) and
// review_verdict_service_test.go (readyVerdictFor, buildReviewThenStageInstance).
// No real DB.

import (
	"context"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
)

// TestRecordFastForward_HappyPath_RecordsBothVerdictAndSignoff (R5 case i):
// a review stage completes (quorum satisfied) and the actor is eligible on
// the now-active approval stage — RecordFastForward records the `ready`
// verdict AND the approve signoff in one call, no error.
func TestRecordFastForward_HappyPath_RecordsBothVerdictAndSignoff(t *testing.T) {
	const (
		instanceID      = "inst-fastfwd-1"
		reviewStageID   = "stage-fastfwd-review-1"
		approvalStageID = "stage-fastfwd-approval-1"
		actorID         = "actor-fastfwd-1"
		authorID        = "author-fastfwd-1"
	)

	inst := buildReviewThenStageInstance(instanceID, reviewStageID, approvalStageID, authorID,
		[]string{actorID}, []string{actorID}, domain.StageKindApproval)

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
		stageSignoffs: []signoffRow{
			{
				id:                 "signoff-fastfwd-1",
				approvalInstanceID: instanceID,
				stageInstanceID:    approvalStageID,
				actorUserID:        actorID,
				actorTenantID:      "tenant-1",
				decision:           "approve",
				signedAt:           signedAt,
				signatureMethod:    "password",
				signaturePayload:   []byte(`{}`),
				contentHash:        validContentHash,
			},
		},
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		stageVerdicts:    []domain.ReviewVerdict{readyVerdictFor("verdict-fastfwd-1", instanceID, reviewStageID, actorID, signedAt)},
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-fastfwd-1", WasReplay: false},
	}
	verdicts := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	decisions := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	svc := newFastForwardService(verdicts, decisions)
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordFastForward(context.Background(), newTxRunner(db), FastForwardRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  reviewStageID,
		ActorUserID:      actorID,
		Comment:          "lgtm, fast-forwarding",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentHash:      validContentHash,
	})
	if err != nil {
		t.Fatalf("RecordFastForward: unexpected error: %v", err)
	}
	if !result.Verdict.StageCompleted {
		t.Error("expected Verdict.StageCompleted=true")
	}
	if !result.Verdict.FastForwardEligible {
		t.Error("expected Verdict.FastForwardEligible=true")
	}
	if !result.Signoff.StageCompleted {
		t.Error("expected Signoff.StageCompleted=true")
	}
	if repo.insertedVerdict == nil {
		t.Error("expected InsertVerdict to have been called")
	}
	if repo.insertedSignoff == nil {
		t.Error("expected InsertSignoff to have been called")
	}
}

// TestRecordFastForward_QuorumPending_StageNotCompleted (case ii): all_of
// quorum with two reviewers, only one verdict recorded — the review stage
// does not complete, so RecordFastForward returns
// domain.ErrFastForwardStageNotCompleted and never attempts the signoff leg.
func TestRecordFastForward_QuorumPending_StageNotCompleted(t *testing.T) {
	const (
		instanceID      = "inst-fastfwd-2"
		reviewStageID   = "stage-fastfwd-review-2"
		approvalStageID = "stage-fastfwd-approval-2"
		actorID         = "actor-fastfwd-2a"
		authorID        = "author-fastfwd-2"
	)

	inst := buildReviewThenStageInstance(instanceID, reviewStageID, approvalStageID, authorID,
		[]string{actorID, "actor-fastfwd-2b"}, []string{actorID}, domain.StageKindApproval)
	inst.Stages[0].QuorumSnapshot = domain.QuorumAllOf

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:      inst,
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-fastfwd-2", instanceID, reviewStageID, actorID, signedAt)},
	}
	verdicts := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	decisions := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	svc := newFastForwardService(verdicts, decisions)
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordFastForward(context.Background(), newTxRunner(db), FastForwardRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  reviewStageID,
		ActorUserID:      actorID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentHash:      validContentHash,
	})
	if !errors.Is(err, domain.ErrFastForwardStageNotCompleted) {
		t.Fatalf("RecordFastForward() error = %v; want domain.ErrFastForwardStageNotCompleted", err)
	}
	if repo.insertedSignoff != nil {
		t.Error("InsertSignoff must not be called when the review stage did not complete")
	}
}

// TestRecordFastForward_ActorNotEligibleOnNextStage_NotEligible (case iii):
// the review stage completes but the actor is not in the next approval
// stage's eligible pool — RecordFastForward returns
// domain.ErrFastForwardNotEligible and never attempts the signoff leg.
func TestRecordFastForward_ActorNotEligibleOnNextStage_NotEligible(t *testing.T) {
	const (
		instanceID      = "inst-fastfwd-3"
		reviewStageID   = "stage-fastfwd-review-3"
		approvalStageID = "stage-fastfwd-approval-3"
		actorID         = "actor-fastfwd-3"
		authorID        = "author-fastfwd-3"
	)

	inst := buildReviewThenStageInstance(instanceID, reviewStageID, approvalStageID, authorID,
		[]string{actorID}, []string{"someone-else"}, domain.StageKindApproval)

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:      inst,
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-fastfwd-3", instanceID, reviewStageID, actorID, signedAt)},
	}
	verdicts := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	decisions := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	svc := newFastForwardService(verdicts, decisions)
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordFastForward(context.Background(), newTxRunner(db), FastForwardRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  reviewStageID,
		ActorUserID:      actorID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentHash:      validContentHash,
	})
	if !errors.Is(err, domain.ErrFastForwardNotEligible) {
		t.Fatalf("RecordFastForward() error = %v; want domain.ErrFastForwardNotEligible", err)
	}
	if repo.insertedSignoff != nil {
		t.Error("InsertSignoff must not be called when the actor is not eligible on the next stage")
	}
}

// TestRecordFastForward_ReadyOnApprovalStage_ErrorPropagated (case iv, G2
// regression): the active stage targeted by the request is already
// approval-kind — the verdict core's stage-kind guard rejects `ready` with
// domain.ErrVerdictReadyOnApprovalStage before any mutation, and
// RecordFastForward propagates that error unchanged.
func TestRecordFastForward_ReadyOnApprovalStage_ErrorPropagated(t *testing.T) {
	const (
		instanceID = "inst-fastfwd-4"
		stageID    = "stage-fastfwd-approval-4"
		actorID    = "actor-fastfwd-4"
		authorID   = "author-fastfwd-4"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	inst.Stages[0].Kind = domain.StageKindApproval // active stage is approval-kind

	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	verdicts := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}}
	decisions := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	svc := newFastForwardService(verdicts, decisions)
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordFastForward(context.Background(), newTxRunner(db), FastForwardRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentHash:      validContentHash,
	})
	if !errors.Is(err, domain.ErrVerdictReadyOnApprovalStage) {
		t.Fatalf("RecordFastForward() error = %v; want domain.ErrVerdictReadyOnApprovalStage", err)
	}
	if repo.insertedSignoff != nil {
		t.Error("InsertSignoff must not be called when the verdict leg is rejected")
	}
}

// TestRecordFastForward_SignoffLegFails_ErrorPropagated (case v): the verdict
// leg succeeds and completes the review stage with the actor eligible on the
// next approval stage, but the signoff leg fails (simulated via
// insertSignoffErr on the fake repo) — RecordFastForward propagates the
// signoff error. Atomic rollback of both writes is exercised at the
// integration level in S4; this test only asserts error propagation.
func TestRecordFastForward_SignoffLegFails_ErrorPropagated(t *testing.T) {
	const (
		instanceID      = "inst-fastfwd-5"
		reviewStageID   = "stage-fastfwd-review-5"
		approvalStageID = "stage-fastfwd-approval-5"
		actorID         = "actor-fastfwd-5"
		authorID        = "author-fastfwd-5"
	)

	inst := buildReviewThenStageInstance(instanceID, reviewStageID, approvalStageID, authorID,
		[]string{actorID}, []string{actorID}, domain.StageKindApproval)

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	wantErr := errors.New("simulated signoff insert failure")
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		stageVerdicts:    []domain.ReviewVerdict{readyVerdictFor("verdict-fastfwd-5", instanceID, reviewStageID, actorID, signedAt)},
		insertSignoffErr: wantErr,
	}
	verdicts := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	decisions := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	svc := newFastForwardService(verdicts, decisions)
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordFastForward(context.Background(), newTxRunner(db), FastForwardRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  reviewStageID,
		ActorUserID:      actorID,
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentHash:      validContentHash,
	})
	if err == nil {
		t.Fatal("expected error from signoff leg; got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("RecordFastForward() error = %v; want wrapped %v", err, wantErr)
	}
	if repo.insertedVerdict == nil {
		t.Error("expected the verdict leg to have run (and its insert attempted) before the signoff leg failed")
	}
}
