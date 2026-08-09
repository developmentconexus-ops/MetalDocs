package application

// phase5_integration_test.go — cross-service scenario tests for Spec 2 Phase 5.
//
// These are NOT database integration tests: no real Postgres is required.
// They wire Submit and Decision together using the same fake-driver pattern as
// the per-service tests in this package, then chain real service method calls
// to prove end-to-end service wiring works.
//
// Two scenarios:
//   1. FullApprovalToTerminal — Submit → approve signoff → terminal approval
//      handed to the ADR 0085 release seam. There is no publish step left to
//      chain: publication is the release coordinator's reaction to durable
//      facts, so recording the approval fact IS the end of the human chain.
//   2. RejectThenResubmit     — Submit → reject signoff → Submit again (new instance)

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// Phase 5 combined fake repo
//
// Satisfies infrastructure.ApprovalRepository for the three scenarios.
// Each service only calls the methods it needs; all others forward to the
// embedded no-op and would panic if unexpectedly called.
// ---------------------------------------------------------------------------

type phase5Repo struct {
	infrastructure.ApprovalRepository // no-op embed

	// Submit methods
	insertInstanceErr       error
	insertStageInstancesErr error

	// Decision methods
	instance          *domain.Instance
	loadInstanceErr   error
	insertSignoffRes  infrastructure.SignoffInsertResult
	insertSignoffErr  error
	updateStageErr    error
	updateInstanceErr error
}

func (r *phase5Repo) InsertInstance(_ context.Context, _ db.Tx, _ domain.Instance) error {
	return r.insertInstanceErr
}

func (r *phase5Repo) InsertStageInstances(_ context.Context, _ db.Tx, _ []domain.StageInstance) error {
	return r.insertStageInstancesErr
}

func (r *phase5Repo) LoadInstance(_ context.Context, _ db.Tx, _, _ string) (*domain.Instance, error) {
	return r.instance, r.loadInstanceErr
}

func (r *phase5Repo) InsertSignoff(_ context.Context, _ db.Tx, _ domain.Signoff) (infrastructure.SignoffInsertResult, error) {
	return r.insertSignoffRes, r.insertSignoffErr
}

func (r *phase5Repo) UpdateStageStatus(_ context.Context, _ db.Tx, _, _ string, _, _ domain.StageStatus) error {
	return r.updateStageErr
}

func (r *phase5Repo) UpdateInstanceStatus(_ context.Context, _ db.Tx, _, _ string, _ domain.InstanceStatus, _ domain.InstanceStatus, _ *time.Time) error {
	return r.updateInstanceErr
}

// The 6 new interface methods for phase5Repo. LoadRoute and ResolveEligibleActors
// are called during Submit; the others during Decision. All delegate to tx so
// the phase5Conn driver handles the queries.

// LoadGovernedRevisionNumber is called by SubmitRevisionForReview (T8b).
// Delegates to tx so phase5Conn's governedRevisionNumberRows serves it (fixed
// at 0 here — these scenarios only exercise the REV-0 default-title path).
func (r *phase5Repo) LoadGovernedRevisionNumber(ctx context.Context, tx db.Tx, tenantID, documentID string) (int, error) {
	var n int64
	err := tx.QueryRowContext(ctx,
		`SELECT revision_number FROM documents WHERE id = $1 AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&n)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *phase5Repo) LoadRoute(ctx context.Context, tx db.Tx, tenantID, routeID string) (domain.Route, error) {
	var route domain.Route
	err := tx.QueryRowContext(ctx, `
		SELECT id, tenant_id, profile_code, version
		FROM approval_routes
		WHERE id = $1 AND tenant_id = $2 AND active = TRUE`,
		routeID, tenantID,
	).Scan(&route.ID, &route.TenantID, &route.ProfileCode, &route.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Route{}, fmt.Errorf("route %s not found for tenant %s", routeID, tenantID)
		}
		return domain.Route{}, err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT ars.stage_order, ars.name, ars.required_capability,
		       ars.quorum, ars.quorum_m, ars.on_eligibility_drift
		  FROM approval_route_stages ars
		  JOIN approval_routes ar
		    ON ar.id = ars.route_id
		   AND ar.tenant_id = $2
		 WHERE ars.route_id = $1
		 ORDER BY ars.stage_order ASC`,
		routeID, tenantID,
	)
	if err != nil {
		return domain.Route{}, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var stage domain.Stage
		var quorumM sql.NullInt32
		if err := rows.Scan(
			&stage.Order, &stage.Name, &stage.RequiredCapability,
			&stage.Quorum, &quorumM, &stage.OnEligibilityDrift,
		); err != nil {
			return domain.Route{}, err
		}
		if quorumM.Valid {
			v := int(quorumM.Int32)
			stage.QuorumM = &v
		}
		// Selectors is the sole source of truth post-slice-6b (see
		// submit_service_test.go's fakeSubmitRepo.LoadRoute for the same
		// hardcoded-selector pattern this fake repo mirrors).
		stage.Selectors = []domain.ActorSelector{{Kind: domain.SelectorRoleInFixedArea, Role: "quality_approver", AreaCode: "QA"}}
		route.Stages = append(route.Stages, stage)
	}
	if err := rows.Err(); err != nil {
		return domain.Route{}, err
	}
	return route, nil
}

func (r *phase5Repo) ResolveEligibleActors(ctx context.Context, tx db.Tx, tenantID, areaCode, requiredRole string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT user_id
		   FROM metaldocs.user_process_areas
		  WHERE tenant_id = $1::uuid
		    AND area_code = $2
		    AND role      = $3
		    AND effective_from <= now()
		    AND (effective_to IS NULL OR effective_to > now())`,
		tenantID, areaCode, requiredRole,
	)
	if err != nil {
		return []string{}, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return []string{}, err
		}
		ids = append(ids, uid)
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, rows.Err()
}

// ResolveEligibleActorsForSelectors is called by SubmitRevisionForReview (M4,
// unit 3.2, slice 3) via stage.Selectors. These scenarios only exercise a
// single role_in_fixed_area selector — delegate to ResolveEligibleActors for
// that kind.
func (r *phase5Repo) ResolveEligibleActorsForSelectors(ctx context.Context, tx db.Tx, tenantID string, selectors []domain.ActorSelector, subjectArea string) ([]string, error) {
	seen := make(map[string]struct{})
	var ids []string
	for _, sel := range selectors {
		var area string
		switch sel.Kind {
		case domain.SelectorRoleInFixedArea:
			area = sel.AreaCode
		case domain.SelectorRoleInDocumentArea:
			area = subjectArea
		default:
			continue
		}
		got, err := r.ResolveEligibleActors(ctx, tx, tenantID, area, sel.Role)
		if err != nil {
			return []string{}, err
		}
		for _, id := range got {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}

func (r *phase5Repo) LoadPriorSignoffs(ctx context.Context, tx db.Tx, tenantID, instanceID, activeStageID string) ([]domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_instance_id,
		       actor_user_id, actor_tenant_id, decision,
		       comment, signed_at, signature_method, signature_payload, content_hash
		FROM approval_signoffs
		WHERE approval_instance_id = $1
		  AND stage_instance_id != $2
		  AND actor_tenant_id = $3
		ORDER BY signed_at ASC`,
		instanceID, activeStageID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDecisionSignoffRows(rows)
}

func (r *phase5Repo) LoadStageSignoffs(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id,
		       s.actor_user_id, s.actor_tenant_id, s.decision,
		       s.comment, s.signed_at, s.signature_method, s.signature_payload, s.content_hash
		  FROM approval_signoffs s
		  JOIN approval_stage_instances asi
		    ON asi.id = s.stage_instance_id
		  JOIN approval_instances ai
		    ON ai.id = asi.approval_instance_id
		   AND ai.id = s.approval_instance_id
		 WHERE s.stage_instance_id = $1
		   AND ai.tenant_id = $2::uuid
		   AND s.actor_tenant_id = ai.tenant_id
		 ORDER BY s.signed_at ASC`,
		stageInstanceID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDecisionSignoffRows(rows)
}

func (r *phase5Repo) HasUnresolvedComments(ctx context.Context, tx db.Tx, tenantID, documentID string) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM document_comments
		 WHERE tenant_id = $1
		   AND document_id = $2
		   AND resolved_at IS NULL`,
		tenantID, documentID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasUnresolvedInstanceComments (F5) mirrors HasUnresolvedComments above —
// this fake's document_comments table has no created_at scoping, so the
// `since` scoping just isn't exercised by the fake driver; the query still
// runs the same shape so tests targeting the freeze call sites can drive it.
func (r *phase5Repo) HasUnresolvedInstanceComments(ctx context.Context, tx db.Tx, tenantID, documentID string, since time.Time) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM document_comments
		 WHERE tenant_id = $1
		   AND document_id = $2
		   AND resolved_at IS NULL`,
		tenantID, documentID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// PinFrozenHash (F5) is a no-op success for this cross-service wiring test —
// the freeze boundary firing successfully (not its persisted row shape) is
// what these scenarios need to prove Submit/Decision/Publish still wire
// together end-to-end.
func (r *phase5Repo) PinFrozenHash(_ context.Context, _ db.Tx, _, _, _ string) (bool, error) {
	return true, nil
}

// LoadFrozenContentHash (F6) reads directly from the passed-in instance's
// FrozenContentHash field, mirroring fakeDecisionRepo's approach — this
// cross-service wiring test cares about Submit/Decision/Publish wiring
// together, not repository query shapes. No-fallback: nil FrozenContentHash
// always returns ErrNoActiveContentHash.
func (r *phase5Repo) LoadFrozenContentHash(_ context.Context, _ db.Tx, _, _ string) (string, error) {
	if r.instance == nil || r.instance.FrozenContentHash == nil {
		return "", infrastructure.ErrNoActiveContentHash
	}
	return *r.instance.FrozenContentHash, nil
}

func (r *phase5Repo) LoadActorDisplayName(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (r *phase5Repo) InsertDelegation(_ context.Context, _ db.Tx, _ domain.Delegation) error {
	return nil
}

func (r *phase5Repo) DeleteDelegation(_ context.Context, _ db.Tx, _, _, _ string, _ bool) (bool, error) {
	return false, nil
}

// LoadActiveDelegationsFor returns no delegations — the F9 default. The
// decision service now calls this on every RecordSignoff to resolve
// domain.ResolveEligibleIdentity's widened eligibility input; this scenario
// exercises direct-actor eligibility only.
func (r *phase5Repo) LoadActiveDelegationsFor(_ context.Context, _ db.Tx, _, _ string, _ time.Time) ([]domain.Delegation, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Phase 5 combined fake SQL driver
//
// Handles all four service shapes in a single conn:
//   • Submit:    route + stage queries (SELECT)
//   • Decision:  signoff queries (SELECT with/without "!=")
//   • Publish:   UPDATE documents
//   • Scheduler: UPDATE documents (with configurable rowsAffected)
//
// updateResults is consumed left-to-right; one value per UPDATE Exec call.
// stageSignoffs feeds the decision service's loadStageSignoffs query.
// ---------------------------------------------------------------------------

type phase5Conn struct {
	stageSignoffs      []signoffRow // reuse decisionTestConn's signoffRow type
	updateResults      []int64
	unresolvedComments int
	updateIdx          int32 // atomic
}

func (c *phase5Conn) nextRowsAffected() int64 {
	idx := atomic.AddInt32(&c.updateIdx, 1) - 1
	if int(idx) >= len(c.updateResults) {
		return 1 // default: success
	}
	return c.updateResults[idx]
}

type phase5NoopResult struct{ ra int64 }

func (r phase5NoopResult) LastInsertId() (int64, error) { return 0, nil }
func (r phase5NoopResult) RowsAffected() (int64, error) { return r.ra, nil }

type phase5EmptyRows struct{}

func (phase5EmptyRows) Columns() []string         { return nil }
func (phase5EmptyRows) Close() error              { return nil }
func (phase5EmptyRows) Next([]driver.Value) error { return io.EOF }

type phase5Stmt struct {
	conn  *phase5Conn
	query string
}

func (s *phase5Stmt) Close() error  { return nil }
func (s *phase5Stmt) NumInput() int { return -1 }

func (s *phase5Stmt) Exec(_ []driver.Value) (driver.Result, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "update") {
		return phase5NoopResult{ra: s.conn.nextRowsAffected()}, nil
	}
	return phase5NoopResult{ra: 1}, nil
}

func (s *phase5Stmt) Query(_ []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)

	if strings.Contains(q, "form_data_json") && strings.Contains(q, "from documents") {
		return &submitSingleValueRows{value: []byte(`{"title":"Doc"}`)}, nil
	}
	if strings.Contains(q, "content_hash_at_submit") && strings.Contains(q, "from documents") {
		return &submitSingleValueRows{value: validContentHash}, nil
	}
	if strings.Contains(q, "select revision_number") && strings.Contains(q, "from documents") {
		// Governed revision_number read (T8b) via ApprovalRepository.LoadGovernedRevisionNumber.
		return &governedRevisionNumberRows{value: 0}, nil
	}
	if strings.Contains(q, "from documents") {
		return &docAreaRows{snapshot: "QA"}, nil
	}
	if strings.Contains(q, "from document_comments") {
		return &submitSingleValueRows{value: s.conn.unresolvedComments}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "iam_user_roles") {
		return &submitSingleValueRows{value: false}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "role_capabilities") {
		return &submitSingleValueRows{value: true}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.asserted_caps'") {
		return &submitSingleValueRows{value: nil}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.tenant_id'") {
		return &submitSingleValueRows{value: tenant.DevTenantID}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.actor_id'") {
		return &submitSingleValueRows{value: "user-1"}, nil
	}

	// Submit: approval_routes SELECT
	if strings.Contains(q, "approval_routes") && strings.Contains(q, "where") && !strings.Contains(q, "approval_route_stages") {
		return &routeRows{}, nil // reuse submit_service_test.go's routeRows
	}
	// Submit: approval_route_stages SELECT
	if strings.Contains(q, "approval_route_stages") {
		return &stageRows{}, nil // reuse submit_service_test.go's stageRows
	}
	// Submit (W6): eligible-actor pool resolution must be non-empty for these
	// pre-existing fixtures, which never seeded eligibility rows because it
	// was unenforced before F2.
	if strings.Contains(q, "user_process_areas") {
		return &submitSingleValueRows{value: "user-eligible-1"}, nil
	}
	// Decision: approval_signoffs SELECT
	if strings.Contains(q, "approval_signoffs") {
		if isStageQuery(s.query) {
			return &signoffRows{rows: s.conn.stageSignoffs}, nil
		}
		return phase5EmptyRows{}, nil // prior-signoffs query (empty)
	}

	return phase5EmptyRows{}, nil
}

func (c *phase5Conn) Prepare(query string) (driver.Stmt, error) {
	return &phase5Stmt{conn: c, query: query}, nil
}

func (c *phase5Conn) Close() error              { return nil }
func (c *phase5Conn) Begin() (driver.Tx, error) { return c, nil }

// BeginTx honours non-default isolation levels used by service transactions.
func (c *phase5Conn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return c, nil
}

func (c *phase5Conn) Commit() error   { return nil }
func (c *phase5Conn) Rollback() error { return nil }

type phase5Driver struct{ conn *phase5Conn }

func (d *phase5Driver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

// newPhase5DB registers a unique fake driver and opens a *sql.DB against it.
func newPhase5DB(t *testing.T, conn *phase5Conn) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("phase5_test_%p", conn)
	sql.Register(name, &phase5Driver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open phase5 test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// ---------------------------------------------------------------------------
// Scenario 1: FullApprovalToTerminal
//
// Flow: Submit → RecordSignoff (approve, quorum met) → terminal approval
// recorded on the release seam.
//
// After RecordSignoff we expect InstanceApproved=true and exactly one
// RecordTerminalApproval call carrying this instance — that hand-off is the
// ADR 0085 replacement for the retired PublishApproved step. Whether the
// document then reaches published is the release coordinator's decision, made
// against durable facts, and is covered by the release integration suite.
//
// Each service call gets its own repo/DB pair because RecordSignoff requires
// instance.Status == InProgress. The shared emitter accumulates events across
// both phases.
// ---------------------------------------------------------------------------

func TestPhase5_FullApprovalToTerminal(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)

	const (
		tenantID   = "tenant-p5-1"
		documentID = "doc-p5-1"
		routeID    = "route-uuid-1" // matches routeRows fixture
		actorID    = "approver-p5-1"
		authorID   = "author-p5-1"
		instanceID = "inst-p5-1"
		stageID    = "stage-p5-1"
	)

	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	// --- Step 1: Submit ---
	// Submit only needs InsertInstance + InsertStageInstances from the repo.
	submitRepo := &phase5Repo{}
	submitConn := &phase5Conn{}
	submitDB := newPhase5DB(t, submitConn)

	submitSvc := &SubmitService{repo: submitRepo, emitter: emitter, clock: clock}

	submitReq := SubmitRequest{
		TenantID:        tenantID,
		DocumentID:      documentID,
		RouteID:         routeID,
		SubmittedBy:     authorID,
		ContentFormData: map[string]any{"title": "P5 Doc", "_content_hash": validContentHash},
		RevisionVersion: 1,
		IdempotencyKey:  "44444444-4444-4444-4444-444444444401",
	}
	submitResult, err := submitSvc.SubmitRevisionForReview(ctx, newTxRunner(submitDB), submitReq)
	if err != nil {
		t.Fatalf("Submit: unexpected error: %v", err)
	}
	if submitResult.InstanceID == "" {
		t.Fatal("Submit: InstanceID must not be empty")
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("after Submit: want 1 event; got %d", len(emitter.Events))
	}
	if emitter.Events[0].EventType != "approval_submitted" {
		t.Errorf("event[0].EventType = %q; want approval_submitted", emitter.Events[0].EventType)
	}

	// --- Step 2: RecordSignoff (approve, quorum met) ---
	// Instance must be InProgress for RecordSignoff to accept it.
	// FrozenContentHash set (F5/F6): by the time signoff is possible the
	// instance must already be frozen — RecordSignoff now reads only this pin.
	p5FrozenHash1 := validContentHash
	inProgressInstance := &domain.Instance{
		ID:                  instanceID,
		TenantID:            tenantID,
		DocumentID:          documentID,
		Subject:             domain.NewDocumentSubject(documentID),
		RouteID:             routeID,
		Status:              domain.InstanceInProgress,
		SubmittedBy:         authorID,
		SubmittedAt:         now,
		RevisionVersion:     1,
		ContentHashAtSubmit: validContentHash,
		FrozenContentHash:   &p5FrozenHash1,
		Stages: []domain.StageInstance{
			{
				ID:                         stageID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 1,
				NameSnapshot:               "QA Review",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				EligibleActorIDs:           []string{actorID},
				Status:                     domain.StageActive,
				OpenedAt:                   &now,
			},
		},
	}

	// signoff row visible in loadStageSignoffs — quorum met immediately (any_1_of).
	decisionStageSignoffs := []signoffRow{
		{
			id:                 "signoff-p5-1",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      tenantID,
			decision:           "approve",
			comment:            "LGTM",
			signedAt:           now,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		},
	}

	decisionRepo := &phase5Repo{
		instance:         inProgressInstance,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-p5-1", WasReplay: false},
	}
	decisionConn := &phase5Conn{stageSignoffs: decisionStageSignoffs}
	decisionDB := newPhase5DB(t, decisionConn)
	releaseRecorder := &fakeReleaseRecorder{}
	decisionSvc := &DecisionService{repo: decisionRepo, emitter: emitter, clock: clock, releaseRecorder: releaseRecorder}

	signoffReq := SignoffRequest{
		TenantID:         tenantID,
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		Comment:          "LGTM",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentFormData:  map[string]any{"title": "P5 Doc", "_content_hash": validContentHash},
	}
	signoffResult, err := decisionSvc.RecordSignoff(ctx, newTxRunner(decisionDB), signoffReq)
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if !signoffResult.InstanceApproved {
		t.Error("after signoff: want InstanceApproved=true")
	}
	if signoffResult.InstanceRejected {
		t.Error("after signoff: want InstanceRejected=false")
	}
	if len(emitter.Events) != 2 {
		t.Fatalf("after RecordSignoff: want 2 events; got %d", len(emitter.Events))
	}
	if emitter.Events[1].EventType != "signoff_recorded" {
		t.Errorf("event[1].EventType = %q; want signoff_recorded", emitter.Events[1].EventType)
	}

	// --- Step 3: terminal approval handed to the release seam ---
	// ADR 0085 stage B: this replaces the retired PublishApproved step. The
	// human chain ends by recording the approval fact; the release coordinator
	// decides publication asynchronously against durable facts.
	if releaseRecorder.calls != 1 {
		t.Fatalf("RecordTerminalApproval calls = %d; want 1", releaseRecorder.calls)
	}
	if releaseRecorder.last.InstanceID != instanceID {
		t.Errorf("release seam InstanceID = %q; want %q", releaseRecorder.last.InstanceID, instanceID)
	}
	// Total events across the human chain: submit + signoff = 2. There is no
	// third "document_published" event here — that one is emitted by the
	// release transaction, not by anything a user invokes.
	if len(emitter.Events) != 2 {
		t.Fatalf("after terminal approval: want 2 events; got %d", len(emitter.Events))
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: RejectThenResubmit
//
// Flow: Submit → RecordSignoff (reject) → verify InstanceRejected=true
//       → Submit again → verify new InstanceID returned
//
// Two separate Submit calls each get their own DB (distinct fake instances)
// to isolate idempotency keys between submissions.
// ---------------------------------------------------------------------------

func TestPhase5_RejectThenResubmit(t *testing.T) {
	ctx := context.Background()

	// Two distinct clock values so the two submissions produce different idempotency keys.
	clockAtSubmit1 := fixedClock{t: time.Date(2026, 4, 22, 11, 0, 0, 0, time.UTC)}
	clockAtSignoff := clockAtSubmit1
	clockAtSubmit2 := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	const (
		tenantID   = "tenant-p5-rej"
		documentID = "doc-p5-rej"
		routeID    = "route-uuid-1"
		actorID    = "approver-p5-rej"
		authorID   = "author-p5-rej"
		instanceID = "inst-p5-rej"
		stageID    = "stage-p5-rej"
	)

	// Instance in in-progress state for the signoff.
	// FrozenContentHash set (F5/F6): by the time signoff is possible the
	// instance must already be frozen — RecordSignoff now reads only this pin.
	p5FrozenHash2 := validContentHash
	inProgressInstance := &domain.Instance{
		ID:                  instanceID,
		TenantID:            tenantID,
		DocumentID:          documentID,
		Subject:             domain.NewDocumentSubject(documentID),
		RouteID:             routeID,
		Status:              domain.InstanceInProgress,
		SubmittedBy:         authorID,
		SubmittedAt:         clockAtSubmit1.t,
		RevisionVersion:     1,
		ContentHashAtSubmit: validContentHash,
		FrozenContentHash:   &p5FrozenHash2,
		Stages: []domain.StageInstance{
			{
				ID:                         stageID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 1,
				NameSnapshot:               "QA Review",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				EligibleActorIDs:           []string{actorID},
				Status:                     domain.StageActive,
				OpenedAt:                   &clockAtSubmit1.t,
			},
		},
	}

	// Reject signoff row returned by loadStageSignoffs.
	rejectSignoffRows := []signoffRow{
		{
			id:                 "signoff-rej-1",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      tenantID,
			decision:           "reject",
			comment:            "Not ready",
			signedAt:           clockAtSignoff.t,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		},
	}

	// --- Phase A: Submit #1 + reject signoff ---
	repo := &phase5Repo{
		instance:         inProgressInstance,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-rej-1", WasReplay: false},
	}
	emitter := &MemoryEmitter{}

	// Submit #1 DB (route + stage queries only).
	conn1 := &phase5Conn{}
	db1 := newPhase5DB(t, conn1)

	submitSvc1 := &SubmitService{repo: repo, emitter: emitter, clock: clockAtSubmit1}

	submitReq1 := SubmitRequest{
		TenantID:        tenantID,
		DocumentID:      documentID,
		RouteID:         routeID,
		SubmittedBy:     authorID,
		ContentFormData: map[string]any{"title": "Reject Doc v1"},
		RevisionVersion: 1,
		IdempotencyKey:  "44444444-4444-4444-4444-444444444402",
	}
	submitResult1, err := submitSvc1.SubmitRevisionForReview(ctx, newTxRunner(db1), submitReq1)
	if err != nil {
		t.Fatalf("Submit #1: unexpected error: %v", err)
	}
	if submitResult1.InstanceID == "" {
		t.Fatal("Submit #1: InstanceID must not be empty")
	}

	// Decision DB — needs approval_signoffs rows.
	connDecision := &phase5Conn{stageSignoffs: rejectSignoffRows}
	dbDecision := newPhase5DB(t, connDecision)

	decisionSvc := &DecisionService{repo: repo, emitter: emitter, clock: clockAtSignoff, releaseRecorder: &fakeReleaseRecorder{}}

	signoffReq := SignoffRequest{
		TenantID:         tenantID,
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "reject",
		Comment:          "Not ready",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "rej"},
		ContentFormData:  map[string]any{"title": "Reject Doc v1", "_content_hash": validContentHash},
	}
	signoffResult, err := decisionSvc.RecordSignoff(ctx, newTxRunner(dbDecision), signoffReq)
	if err != nil {
		t.Fatalf("RecordSignoff (reject): unexpected error: %v", err)
	}
	if !signoffResult.InstanceRejected {
		t.Error("after reject signoff: want InstanceRejected=true")
	}
	if signoffResult.InstanceApproved {
		t.Error("after reject signoff: want InstanceApproved=false")
	}

	// Verify governance events so far: 1 submit + 1 signoff.
	if len(emitter.Events) != 2 {
		t.Fatalf("after reject signoff: want 2 events; got %d", len(emitter.Events))
	}
	if emitter.Events[1].EventType != "signoff_recorded" {
		t.Errorf("event[1].EventType = %q; want signoff_recorded", emitter.Events[1].EventType)
	}

	// --- Phase B: Submit #2 (resubmission after rejection) ---
	conn2 := &phase5Conn{}
	db2 := newPhase5DB(t, conn2)

	// Fresh submit service with a later clock to guarantee a distinct idempotency key.
	submitSvc2 := &SubmitService{repo: repo, emitter: emitter, clock: clockAtSubmit2}

	submitReq2 := SubmitRequest{
		TenantID:        tenantID,
		DocumentID:      documentID,
		RouteID:         routeID,
		SubmittedBy:     authorID,
		ContentFormData: map[string]any{"title": "Reject Doc v2"},
		RevisionVersion: 2, // bumped revision
		IdempotencyKey:  "44444444-4444-4444-4444-444444444403",
	}
	submitResult2, err := submitSvc2.SubmitRevisionForReview(ctx, newTxRunner(db2), submitReq2)
	if err != nil {
		t.Fatalf("Submit #2: unexpected error: %v", err)
	}
	if submitResult2.InstanceID == "" {
		t.Fatal("Submit #2: InstanceID must not be empty")
	}

	// The two submissions must produce distinct instance IDs (both are UUIDs generated
	// inside the service; verifying they are non-empty and different is the meaningful
	// assertion — the exact UUIDs are non-deterministic).
	if submitResult1.InstanceID == submitResult2.InstanceID {
		t.Errorf("Submit #1 and #2 must produce distinct InstanceIDs; both returned %q", submitResult1.InstanceID)
	}

	// Total events: 1 (submit1) + 1 (signoff) + 1 (submit2) = 3.
	if len(emitter.Events) != 3 {
		t.Fatalf("after resubmit: want 3 events; got %d", len(emitter.Events))
	}
	if emitter.Events[2].EventType != "approval_submitted" {
		t.Errorf("event[2].EventType = %q; want approval_submitted", emitter.Events[2].EventType)
	}
}
