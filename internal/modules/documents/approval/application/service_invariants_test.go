package application

// service_invariants_test.go — invariant-guarding edge cases for the approval
// application services, pruned from the former coverage_boost_test.go (TST-10).
// Each test pins a behavior NOT covered by the per-service test files:
//
//   - SQLEmitter executes an INSERT into governance_events (audit-trail write).
//   - Float ban enforced at every service boundary: submit form data (top-level
//     and nested via content hash), signoff signature payload, and []any-nested
//     form data in ComputeContentHash (content_hash_test.go covers map nesting).
//   - RecordSignoff state legality: completed instance, non-active stage,
//     missing stage instance id, instance with no active stage.
//   - Signoff idempotency: duplicate actor rejected; idempotent replay returns
//     a neutral result without error.
//   - Multi-stage progression: completing stage 1 reports StageCompleted while
//     the instance is not yet approved (stage 2 pending).
//   - OCC stale-revision conflict on PublishApproved and SchedulePublish.
//   - Not-found → infrastructure.ErrNoActiveInstance mapping for signoff, publish,
//     and obsolete entrypoints.
//   - SchedulePublish requires an approved instance.
//
// All tests use the shared fakes defined in the sibling *_test.go files
// (decision_service_test.go, submit_service_test.go, publish_service_test.go,
// obsolete_service_test.go, recording_driver_test.go).

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
)

// ============================================================
// SQLEmitter — governance event audit-trail insert
// ============================================================

var sqlEmitterDBCounter int

func newSQLEmitterTestDB(t *testing.T, conn *recordingConn) *sql.DB {
	t.Helper()
	sqlEmitterDBCounter++
	name := fmt.Sprintf("sql_emitter_test_%d", sqlEmitterDBCounter)
	sql.Register(name, &recordingDriverInstance{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open sql emitter test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSQLEmitter_Emit_ExecutesInsert(t *testing.T) {
	conn := &recordingConn{}
	db := newSQLEmitterTestDB(t, conn)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	e := NewSQLEmitter()
	ev := GovernanceEvent{
		TenantID:     "t1",
		EventType:    "doc_submitted",
		ActorUserID:  "u1",
		ResourceType: "document",
		ResourceID:   "doc-1",
		Reason:       "test",
		PayloadJSON:  json.RawMessage(`{"x":1}`),
	}
	if err := e.Emit(context.Background(), tx, ev); err != nil {
		t.Fatalf("Emit: unexpected error: %v", err)
	}

	// Check that an INSERT statement was executed.
	found := false
	for _, q := range conn.executed {
		if strings.Contains(strings.ToLower(q), "insert") &&
			strings.Contains(strings.ToLower(q), "governance_events") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected INSERT into governance_events; executed: %v", conn.executed)
	}

	_ = tx.Rollback()
}

// ============================================================
// Float ban — enforced at the service boundaries
// ============================================================

func TestComputeContentHash_FloatInNestedArray(t *testing.T) {
	_, err := ComputeContentHash(ContentHashInput{
		TenantID:       "t",
		DocumentID:     "d",
		RevisionNumber: 1,
		FormData:       map[string]any{"nums": []any{float64(1.5)}},
	})
	if !errors.Is(err, ErrFloatInFormData) {
		t.Errorf("want ErrFloatInFormData; got %v", err)
	}
}

func TestSubmitRevisionForReview_FloatPayloadRejected(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Now()}
	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db := newSubmitTestDB(t)

	req := SubmitRequest{
		TenantID:        "t",
		DocumentID:      "d",
		RouteID:         "r",
		SubmittedBy:     "u",
		ContentFormData: map[string]any{"bad": float64(1.5)},
		RevisionVersion: 1,
	}
	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrFloatInFormData) {
		t.Errorf("want ErrFloatInFormData; got %v", err)
	}
}

// Float inside nested map → ErrFloatInFormData wraps to content hash error.
// ValidateEventPayload only checks top-level keys; ComputeContentHash calls
// validateNoFloats which walks nested.
func TestSubmitRevisionForReview_ContentHashError(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Now()}
	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db := newSubmitTestDB(t)

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		ContentFormData: map[string]any{"nested": map[string]any{"val": float64(1.5)}},
		RevisionVersion: 1,
	}
	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected error for float in nested form data; got nil")
	}
	if !errors.Is(err, ErrFloatInFormData) {
		t.Errorf("want ErrFloatInFormData in chain; got %v", err)
	}
}

func TestRecordSignoff_FloatInPayload(t *testing.T) {
	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:         "t",
		InstanceID:       "inst",
		StageInstanceID:  "stage",
		ActorUserID:      "actor",
		Decision:         "approve",
		SignaturePayload: map[string]any{"bad": float64(1.5)},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrFloatInFormData) {
		t.Errorf("want ErrFloatInFormData; got %v", err)
	}
}

// ============================================================
// RecordSignoff — state-transition legality
// ============================================================

func TestRecordSignoff_LoadInstanceNotFound(t *testing.T) {
	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{loadInstanceErr: sql.ErrNoRows}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "t",
		InstanceID:      "inst",
		StageInstanceID: "stage",
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, infrastructure.ErrNoActiveInstance) {
		t.Errorf("want ErrNoActiveInstance; got %v", err)
	}
}

func TestRecordSignoff_InstanceAlreadyCompleted(t *testing.T) {
	inst := buildSingleStageInstance("inst-done", "stage-done", "author", []string{"actor"})
	inst.Status = domain.InstanceApproved // terminal

	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "t",
		InstanceID:      "inst-done",
		StageInstanceID: "stage-done",
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, infrastructure.ErrInstanceCompleted) {
		t.Errorf("want ErrInstanceCompleted; got %v", err)
	}
}

func TestRecordSignoff_StageNotActive(t *testing.T) {
	inst := buildSingleStageInstance("inst-stale", "stage-1", "author", []string{"actor"})
	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "t",
		InstanceID:      "inst-stale",
		StageInstanceID: "wrong-stage-id", // doesn't match active stage
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, infrastructure.ErrStageNotActive) {
		t.Errorf("want ErrStageNotActive; got %v", err)
	}
}

func TestRecordSignoff_RequiresStageInstanceID(t *testing.T) {
	inst := buildSingleStageInstance("inst-missing-stage", "stage-required", "author", []string{"actor"})
	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:        "t",
		InstanceID:      "inst-missing-stage",
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"_content_hash": validContentHash},
	})
	if !errors.Is(err, infrastructure.ErrStageNotActive) {
		t.Errorf("want ErrStageNotActive for missing stage instance id; got %v", err)
	}
}

func buildAllCompletedInstance(instanceID, stageID string) *domain.Instance {
	now := time.Now().UTC()
	return &domain.Instance{
		ID:              instanceID,
		TenantID:        "tenant-1",
		DocumentID:      "doc-1",
		RouteID:         "route-1",
		Status:          domain.InstanceInProgress,
		SubmittedBy:     "author",
		SubmittedAt:     now,
		RevisionVersion: 1,
		Stages: []domain.StageInstance{
			{
				ID:         stageID,
				StageOrder: 1,
				Status:     domain.StageCompleted, // completed, not active
			},
		},
	}
}

func TestRecordSignoff_NoActiveStage(t *testing.T) {
	inst := buildAllCompletedInstance("inst-no-active", "stage-completed")
	// Instance status is InProgress but no stage is Active.
	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      "inst-no-active",
		StageInstanceID: "", // empty → won't hit the stageNotActive check
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, domain.ErrNoActiveStage) {
		t.Errorf("want domain.ErrNoActiveStage; got %v", err)
	}
}

// ============================================================
// RecordSignoff — idempotency
// ============================================================

func TestRecordSignoff_ActorAlreadySigned(t *testing.T) {
	inst := buildSingleStageInstance("inst-dup", "stage-dup", "author", []string{"actor"})

	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffErr: infrastructure.ErrActorAlreadySigned,
	}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "t",
		InstanceID:      "inst-dup",
		StageInstanceID: "stage-dup",
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, infrastructure.ErrActorAlreadySigned) {
		t.Errorf("want ErrActorAlreadySigned; got %v", err)
	}
}

func TestRecordSignoff_IdempotentReplay(t *testing.T) {
	inst := buildSingleStageInstance("inst-replay", "stage-replay", "author", []string{"actor"})
	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)

	conn := &decisionTestConn{}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-replay", WasReplay: true},
	}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "t",
		InstanceID:      "inst-replay",
		StageInstanceID: "stage-replay",
		ActorUserID:     "actor",
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("unexpected error on replay: %v", err)
	}
	// Replay returns neutral result.
	if result.StageCompleted || result.InstanceApproved || result.InstanceRejected {
		t.Error("replay should return neutral SignoffResult")
	}
}

// ============================================================
// RecordSignoff — multi-stage progression
// ============================================================

// Two-stage instance: advancing stage activates next stage.
func buildTwoStageInstance(instanceID, stage1ID, stage2ID, authorUserID string, eligible []string) *domain.Instance {
	now := time.Now().UTC()
	return &domain.Instance{
		ID:                  instanceID,
		TenantID:            "tenant-1",
		DocumentID:          "doc-1",
		RouteID:             "route-1",
		Status:              domain.InstanceInProgress,
		SubmittedBy:         authorUserID,
		SubmittedAt:         now,
		RevisionVersion:     1,
		ContentHashAtSubmit: validContentHash,
		Stages: []domain.StageInstance{
			{
				ID:                         stage1ID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 1,
				NameSnapshot:               "Stage 1",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				EligibleActorIDs:           eligible,
				Status:                     domain.StageActive,
				OpenedAt:                   &now,
			},
			{
				ID:                         stage2ID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 2,
				NameSnapshot:               "Stage 2",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				EligibleActorIDs:           eligible,
				Status:                     domain.StagePending,
			},
		},
	}
}

func TestRecordSignoff_ActivateNextStage(t *testing.T) {
	const (
		instanceID = "inst-two-stage"
		stage1ID   = "stage-one"
		stage2ID   = "stage-two"
		actorID    = "actor-ts"
	)
	inst := buildTwoStageInstance(instanceID, stage1ID, stage2ID, "author", []string{actorID})
	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)

	stageSignoffs := []signoffRow{{
		id:                 "sig-ts-1",
		approvalInstanceID: instanceID,
		stageInstanceID:    stage1ID,
		actorUserID:        actorID,
		actorTenantID:      "tenant-1",
		decision:           "approve",
		signedAt:           signedAt,
		signaturePayload:   []byte(`{}`),
		contentHash:        validContentHash,
	}}

	conn := &decisionTestConn{stageSignoffs: stageSignoffs}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "sig-ts-1", WasReplay: false},
	}
	emitter := &MemoryEmitter{}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stage1ID,
		ActorUserID:     actorID,
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}
	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.StageCompleted {
		t.Error("expected StageCompleted=true")
	}
	// Second stage still pending → instance NOT approved yet.
	if result.InstanceApproved {
		t.Error("expected InstanceApproved=false (second stage pending)")
	}
}

// ============================================================
// PublishApproved / SchedulePublish — OCC and state legality
// ============================================================

func TestPublishApproved_LoadInstanceNotFound(t *testing.T) {
	repo := &fakePublishRepo{loadErr: sql.ErrNoRows}
	svc := &PublishService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}}
	db := newPublishTestDB(t, 1)

	_, err := svc.PublishApproved(context.Background(), newTxRunner(db), PublishRequest{
		TenantID:    "t",
		InstanceID:  "inst",
		PublishedBy: "u",
	})
	if !errors.Is(err, infrastructure.ErrNoActiveInstance) {
		t.Errorf("want ErrNoActiveInstance; got %v", err)
	}
}

func TestPublishApproved_OCC_StaleRevision(t *testing.T) {
	inst := &domain.Instance{
		ID:              "inst-stale-pub",
		TenantID:        "t",
		DocumentID:      "doc-stale-pub",
		Status:          domain.InstanceApproved,
		RevisionVersion: 3,
	}
	repo := &fakePublishRepo{instance: inst}
	svc := &PublishService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}}
	// rowsAffected=0 → OCC conflict.
	db := newPublishTestDB(t, 0)

	_, err := svc.PublishApproved(context.Background(), newTxRunner(db), PublishRequest{
		TenantID:    "t",
		InstanceID:  "inst-stale-pub",
		PublishedBy: "u",
	})
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
		t.Errorf("want ErrStaleRevision; got %v", err)
	}
}

func TestSchedulePublish_InstanceNotApproved(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)

	inst := &domain.Instance{
		ID:       "inst-sched-na",
		TenantID: "t",
		Status:   domain.InstanceInProgress, // not approved
	}
	repo := &fakePublishRepo{instance: inst}
	svc := &PublishService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: now}}
	db := newPublishTestDB(t, 1)

	_, err := svc.SchedulePublish(context.Background(), newTxRunner(db), SchedulePublishRequest{
		TenantID:      "t",
		InstanceID:    "inst-sched-na",
		EffectiveDate: future,
		ScheduledBy:   "u",
	})
	if !errors.Is(err, ErrInstanceNotApproved) {
		t.Errorf("want ErrInstanceNotApproved; got %v", err)
	}
}

func TestSchedulePublish_OCC_StaleRevision(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)

	inst := &domain.Instance{
		ID:         "inst-sched-stale",
		TenantID:   "t",
		DocumentID: "doc-sched-stale",
		Status:     domain.InstanceApproved,
	}
	repo := &fakePublishRepo{instance: inst}
	svc := &PublishService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: now}}
	// rowsAffected=0 → OCC conflict.
	db := newPublishTestDB(t, 0)

	_, err := svc.SchedulePublish(context.Background(), newTxRunner(db), SchedulePublishRequest{
		TenantID:      "t",
		InstanceID:    "inst-sched-stale",
		EffectiveDate: future,
		ScheduledBy:   "u",
	})
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
		t.Errorf("want ErrStaleRevision; got %v", err)
	}
}

// ============================================================
// MarkObsolete — not-found mapping
// ============================================================

func TestMarkObsolete_DocumentNotFound(t *testing.T) {
	conn := &obsoleteTestConn{notFound: true}
	db := newObsoleteTestDB(t, conn)
	svc := &ObsoleteService{emitter: &MemoryEmitter{}, clock: fixedClock{t: time.Now()}}

	_, err := svc.MarkObsolete(context.Background(), newTxRunner(db), MarkObsoleteRequest{
		TenantID:        "t",
		DocumentID:      "doc-nf",
		MarkedBy:        "u",
		RevisionVersion: 1,
		Reason:          "test",
	})
	if !errors.Is(err, infrastructure.ErrNoActiveInstance) {
		t.Errorf("want ErrNoActiveInstance for not-found doc; got %v", err)
	}
}
