package application

// review_verdict_service_test.go — application-layer unit tests for
// ReviewVerdictService.RecordVerdict, reusing the shared fake harness defined
// in decision_service_test.go (fakeDecisionRepo, decisionTestConn,
// newDecisionTestDB, buildSingleStageInstance, newTxRunner). No real DB.

import (
	"context"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
)

// TestRecordVerdict_ReadyOnApprovalStageRejected proves slice 2's service
// pre-check: a `ready` verdict targeting an approval-kind active stage is
// rejected with domain.ErrVerdictReadyOnApprovalStage BEFORE eligibility/SoD
// work and before any DB mutation (InsertVerdict / instance-status write) —
// mirroring NewVerdict's domain-level guard (slice 1) at the service boundary.
//
// Isolation: the acting actor is deliberately NOT in the stage's eligible pool.
// The service flow is authz.Require -> stage-kind guard -> eligibility -> SoD ->
// NewVerdict. If the service-level guard (review_verdict_service.go:130) were
// deleted, execution would reach ResolveEligibleIdentity and fail with an
// eligibility error instead. Asserting the stage-kind sentinel with an
// ineligible actor therefore proves THIS guard fires first — the assertion would
// not hold on the domain backstop alone (that only re-raises inside NewVerdict,
// past the eligibility gate this actor cannot clear).
func TestRecordVerdict_ReadyOnApprovalStageRejected(t *testing.T) {
	const (
		instanceID = "inst-verdict-approval-kind"
		stageID    = "stage-verdict-approval-kind"
		actorID    = "approver-verdict-1"
		authorID   = "author-verdict-1"
	)

	// Eligible pool excludes the acting actor (see isolation note above).
	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{"some-other-approver"})
	inst.Stages[0].Kind = domain.StageKindApproval // active stage is approval-kind

	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordVerdict(context.Background(), newTxRunner(db), ReviewVerdictRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     actorID,
		Verdict:         domain.VerdictReady,
	})

	if !errors.Is(err, domain.ErrVerdictReadyOnApprovalStage) {
		t.Fatalf("RecordVerdict() error = %v; want domain.ErrVerdictReadyOnApprovalStage", err)
	}
	if result != (ReviewVerdictResult{}) {
		t.Errorf("RecordVerdict() result = %+v; want zero value (no mutation before the stage-kind guard fires)", result)
	}
	if repo.insertedSignoff != nil {
		t.Error("no signoff/verdict insert should have been attempted before the stage-kind guard")
	}
	if repo.instanceStatusTo != "" || repo.instanceStatusFrom != "" {
		t.Errorf("no instance-status mutation should have been attempted; got to=%q from=%q", repo.instanceStatusTo, repo.instanceStatusFrom)
	}
}

// readyVerdictFor builds a minimal `ready` domain.ReviewVerdict for the given
// stage/actor, used to populate fakeDecisionRepo.stageVerdicts so
// LoadStageVerdicts reports the just-inserted verdict visible within the same
// tx (mirrors decision_service_test.go's stageSignoffs fixture pattern).
func readyVerdictFor(id, instanceID, stageID, actorID string, at time.Time) domain.ReviewVerdict {
	v, err := domain.NewVerdict(domain.VerdictParams{
		ID:                       id,
		ApprovalInstanceID:       instanceID,
		StageInstanceID:          stageID,
		StageKind:                domain.StageKindReview,
		ActorUserID:              actorID,
		ActorTenantID:            "tenant-1",
		Verdict:                  domain.VerdictReady,
		VerdictAt:                at,
		ActorDisplayNameSnapshot: "Reviewer",
	})
	if err != nil {
		panic(err)
	}
	return *v
}

// TestRecordVerdict_FastForwardEligible_ActorEligibleOnNextApprovalStage
// (R5, unit 2.3 G3, case i): the verdict completes the review stage (quorum
// satisfied) and the actor is in the now-active approval stage's eligible
// pool ⇒ FastForwardEligible=true and NextStageID set to the approval
// stage's id.
func TestRecordVerdict_FastForwardEligible_ActorEligibleOnNextApprovalStage(t *testing.T) {
	const (
		instanceID      = "inst-ff-1"
		reviewStageID   = "stage-ff-review-1"
		approvalStageID = "stage-ff-approval-1"
		actorID         = "reviewer-ff-1"
		authorID        = "author-ff-1"
	)

	inst := buildReviewThenStageInstance(instanceID, reviewStageID, approvalStageID, authorID,
		[]string{actorID}, []string{actorID}, domain.StageKindApproval)

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:      inst,
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-ff-1", instanceID, reviewStageID, actorID, signedAt)},
	}
	svc := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordVerdict(context.Background(), newTxRunner(db), ReviewVerdictRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: reviewStageID,
		ActorUserID:     actorID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict: unexpected error: %v", err)
	}
	if !result.StageCompleted {
		t.Error("expected StageCompleted=true")
	}
	if !result.FastForwardEligible {
		t.Error("expected FastForwardEligible=true — actor is eligible on the now-active approval stage")
	}
	if result.NextStageID == nil || *result.NextStageID != approvalStageID {
		t.Errorf("NextStageID = %v; want %q", result.NextStageID, approvalStageID)
	}
}

// TestRecordVerdict_FastForwardIneligible_ActorNotOnNextApprovalStage
// (case ii): actor not in the next stage's eligible pool ⇒ false.
func TestRecordVerdict_FastForwardIneligible_ActorNotOnNextApprovalStage(t *testing.T) {
	const (
		instanceID      = "inst-ff-2"
		reviewStageID   = "stage-ff-review-2"
		approvalStageID = "stage-ff-approval-2"
		actorID         = "reviewer-ff-2"
		authorID        = "author-ff-2"
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
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-ff-2", instanceID, reviewStageID, actorID, signedAt)},
	}
	svc := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordVerdict(context.Background(), newTxRunner(db), ReviewVerdictRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: reviewStageID,
		ActorUserID:     actorID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict: unexpected error: %v", err)
	}
	if !result.StageCompleted {
		t.Error("expected StageCompleted=true")
	}
	if result.FastForwardEligible {
		t.Error("expected FastForwardEligible=false — actor is not eligible on the now-active approval stage")
	}
	if result.NextStageID != nil {
		t.Errorf("NextStageID = %v; want nil", result.NextStageID)
	}
}

// TestRecordVerdict_FastForwardIneligible_NextStageIsReviewKind (case iii):
// the now-active stage is review-kind, not approval-kind ⇒ false regardless
// of actor eligibility (R5 only offers fast-forward onto an approval stage).
func TestRecordVerdict_FastForwardIneligible_NextStageIsReviewKind(t *testing.T) {
	const (
		instanceID = "inst-ff-3"
		stage1ID   = "stage-ff-review-3a"
		stage2ID   = "stage-ff-review-3b"
		actorID    = "reviewer-ff-3"
		authorID   = "author-ff-3"
	)

	inst := buildReviewThenStageInstance(instanceID, stage1ID, stage2ID, authorID,
		[]string{actorID}, []string{actorID}, domain.StageKindReview)

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:      inst,
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-ff-3", instanceID, stage1ID, actorID, signedAt)},
	}
	svc := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordVerdict(context.Background(), newTxRunner(db), ReviewVerdictRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stage1ID,
		ActorUserID:     actorID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict: unexpected error: %v", err)
	}
	if !result.StageCompleted {
		t.Error("expected StageCompleted=true")
	}
	if result.FastForwardEligible {
		t.Error("expected FastForwardEligible=false — the now-active stage is review-kind, not approval-kind")
	}
	if result.NextStageID != nil {
		t.Errorf("NextStageID = %v; want nil", result.NextStageID)
	}
}

// TestRecordVerdict_FastForwardIneligible_QuorumNotYetMet (case iv): the
// verdict does NOT complete the review stage (quorum still pending) ⇒ false.
func TestRecordVerdict_FastForwardIneligible_QuorumNotYetMet(t *testing.T) {
	const (
		instanceID      = "inst-ff-4"
		reviewStageID   = "stage-ff-review-4"
		approvalStageID = "stage-ff-approval-4"
		actorID         = "reviewer-ff-4a"
		authorID        = "author-ff-4"
	)

	// all_of quorum, two eligible reviewers — only one verdict recorded so
	// quorum stays pending.
	inst := buildReviewThenStageInstance(instanceID, reviewStageID, approvalStageID, authorID,
		[]string{actorID, "reviewer-ff-4b"}, []string{actorID}, domain.StageKindApproval)
	inst.Stages[0].QuorumSnapshot = domain.QuorumAllOf

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:      inst,
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-ff-4", instanceID, reviewStageID, actorID, signedAt)},
	}
	svc := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordVerdict(context.Background(), newTxRunner(db), ReviewVerdictRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: reviewStageID,
		ActorUserID:     actorID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict: unexpected error: %v", err)
	}
	if result.StageCompleted {
		t.Error("expected StageCompleted=false — quorum not yet met (1 of 2 reviewers)")
	}
	if result.FastForwardEligible {
		t.Error("expected FastForwardEligible=false — the review stage has not completed")
	}
	if result.NextStageID != nil {
		t.Errorf("NextStageID = %v; want nil", result.NextStageID)
	}
}

// TestRecordVerdict_FastForwardIneligible_InstanceFullyApproved (case v): a
// single-stage (review-kind) instance completing leaves no next active stage
// ⇒ FastForwardEligible=false (InstanceApproved=true instead).
func TestRecordVerdict_FastForwardIneligible_InstanceFullyApproved(t *testing.T) {
	const (
		instanceID = "inst-ff-5"
		stageID    = "stage-ff-review-5"
		actorID    = "reviewer-ff-5"
		authorID   = "author-ff-5"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	inst.Stages[0].Kind = domain.StageKindReview

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:      inst,
		stageVerdicts: []domain.ReviewVerdict{readyVerdictFor("verdict-ff-5", instanceID, stageID, actorID, signedAt)},
	}
	svc := &ReviewVerdictService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordVerdict(context.Background(), newTxRunner(db), ReviewVerdictRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     actorID,
		Verdict:         domain.VerdictReady,
	})
	if err != nil {
		t.Fatalf("RecordVerdict: unexpected error: %v", err)
	}
	if !result.InstanceApproved {
		t.Error("expected InstanceApproved=true — single-stage instance")
	}
	if result.FastForwardEligible {
		t.Error("expected FastForwardEligible=false — no next stage exists")
	}
	if result.NextStageID != nil {
		t.Errorf("NextStageID = %v; want nil", result.NextStageID)
	}
}
