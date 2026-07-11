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

	"metaldocs/internal/modules/documents/approval/domain"
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
