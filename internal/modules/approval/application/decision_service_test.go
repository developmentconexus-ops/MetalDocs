package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// Ensure fakeDecisionRepo satisfies ApprovalRepository at compile time.
var _ infrastructure.ApprovalRepository = (*fakeDecisionRepo)(nil)

// ---------------------------------------------------------------------------
// Fake repo — only the methods called by RecordSignoff.
// ---------------------------------------------------------------------------

type fakeDecisionRepo struct {
	// Embed no-op to satisfy interface; listed methods are real overrides.
	infrastructure.ApprovalRepository

	instance                   *domain.Instance
	loadInstanceErr            error
	insertSignoffRes           infrastructure.SignoffInsertResult
	insertSignoffErr           error
	insertedSignoff            *domain.Signoff
	updateStageErr             error
	updateInstanceErr          error
	instanceStatusTo           domain.InstanceStatus
	instanceStatusFrom         domain.InstanceStatus
	actorDisplayName           string
	loadActorDisplayNameCalled bool
	loadActorDisplayNameTenant string
	loadActorDisplayNameUser   string

	insertedVerdict      *domain.ReviewVerdict
	insertVerdictRes     infrastructure.VerdictInsertResult
	insertVerdictErr     error
	stageVerdicts        []domain.ReviewVerdict
	loadStageVerdictsErr error
}

func (r *fakeDecisionRepo) LoadInstance(_ context.Context, _ db.Tx, _, _ string) (*domain.Instance, error) {
	return r.instance, r.loadInstanceErr
}

func (r *fakeDecisionRepo) InsertSignoff(_ context.Context, _ db.Tx, s domain.Signoff) (infrastructure.SignoffInsertResult, error) {
	r.insertedSignoff = &s
	return r.insertSignoffRes, r.insertSignoffErr
}

func (r *fakeDecisionRepo) InsertVerdict(_ context.Context, _ db.Tx, v domain.ReviewVerdict) (infrastructure.VerdictInsertResult, error) {
	r.insertedVerdict = &v
	return r.insertVerdictRes, r.insertVerdictErr
}

func (r *fakeDecisionRepo) LoadStageVerdicts(_ context.Context, _ db.Tx, _, _ string) ([]domain.ReviewVerdict, error) {
	return r.stageVerdicts, r.loadStageVerdictsErr
}

func (r *fakeDecisionRepo) UpdateStageStatus(_ context.Context, _ db.Tx, _, _ string, _, _ domain.StageStatus) error {
	return r.updateStageErr
}

func (r *fakeDecisionRepo) UpdateInstanceStatus(_ context.Context, _ db.Tx, _, _ string, to domain.InstanceStatus, from domain.InstanceStatus, _ *time.Time) error {
	r.instanceStatusTo = to
	r.instanceStatusFrom = from
	return r.updateInstanceErr
}

// The 6 new interface methods delegate to the tx so the in-memory decisionTestConn
// driver handles their queries — preserving existing test logic without changes.

func (r *fakeDecisionRepo) LoadPriorSignoffs(ctx context.Context, tx db.Tx, tenantID, instanceID, activeStageID string) ([]domain.Signoff, error) {
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
	defer rows.Close()
	return scanDecisionSignoffRows(rows)
}

func (r *fakeDecisionRepo) LoadStageSignoffs(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.Signoff, error) {
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
	defer rows.Close()
	return scanDecisionSignoffRows(rows)
}

func scanDecisionSignoffRows(rows *sql.Rows) ([]domain.Signoff, error) {
	var signoffs []domain.Signoff
	for rows.Next() {
		var (
			id, instanceID, stageID, actorUserID, actorTenantID string
			decision, comment, signatureMethod, contentHash     string
			signedAt                                            time.Time
			sigPayload                                          []byte
		)
		if err := rows.Scan(&id, &instanceID, &stageID, &actorUserID, &actorTenantID,
			&decision, &comment, &signedAt, &signatureMethod, &sigPayload, &contentHash); err != nil {
			return nil, err
		}
		s, err := domain.NewSignoff(domain.SignoffParams{
			ID:                 id,
			ApprovalInstanceID: instanceID,
			StageInstanceID:    stageID,
			ActorUserID:        actorUserID,
			ActorTenantID:      actorTenantID,
			Decision:           domain.Decision(decision),
			Comment:            comment,
			SignedAt:           signedAt,
			SignatureMethod:    signatureMethod,
			SignaturePayload:   json.RawMessage(sigPayload),
			ContentHash:        contentHash,
		})
		if err != nil {
			return nil, fmt.Errorf("scan signoff %s: %w", id, err)
		}
		signoffs = append(signoffs, *s)
	}
	return signoffs, rows.Err()
}

func (r *fakeDecisionRepo) HasUnresolvedComments(ctx context.Context, tx db.Tx, tenantID, documentID string) (bool, error) {
	var unresolvedCount int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM document_comments
		 WHERE tenant_id = $1
		   AND document_id = $2
		   AND resolved_at IS NULL`,
		tenantID, documentID,
	).Scan(&unresolvedCount)
	if err != nil {
		return false, err
	}
	return unresolvedCount > 0, nil
}

// LoadFrozenContentHash (F6) reads directly from the already-loaded in-memory
// instance's FrozenContentHash field — the fake has no separate approval_instances
// table to query, and the instance passed to RecordSignoff IS the row this method
// would read in production. No-fallback: NULL FrozenContentHash always returns
// ErrNoActiveContentHash, never a substitute value.
func (r *fakeDecisionRepo) LoadFrozenContentHash(_ context.Context, _ db.Tx, _, _ string) (string, error) {
	if r.instance == nil || r.instance.FrozenContentHash == nil {
		return "", infrastructure.ErrNoActiveContentHash
	}
	return *r.instance.FrozenContentHash, nil
}

func (r *fakeDecisionRepo) LoadActorDisplayName(_ context.Context, tenantID, userID string) (string, error) {
	r.loadActorDisplayNameCalled = true
	r.loadActorDisplayNameTenant = tenantID
	r.loadActorDisplayNameUser = userID
	return r.actorDisplayName, nil
}

func (r *fakeDecisionRepo) ResolveEligibleActors(ctx context.Context, tx db.Tx, tenantID, areaCode, requiredRole string) ([]string, error) {
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
	defer rows.Close()
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

func (r *fakeDecisionRepo) LoadRoute(_ context.Context, _ db.Tx, _, _ string) (domain.Route, error) {
	panic("fakeDecisionRepo.LoadRoute: not expected to be called in decision tests")
}

func (r *fakeDecisionRepo) InsertDelegation(_ context.Context, _ db.Tx, _ domain.Delegation) error {
	return nil
}

func (r *fakeDecisionRepo) DeleteDelegation(_ context.Context, _ db.Tx, _, _, _ string, _ bool) (bool, error) {
	return false, nil
}

// LoadActiveDelegationsFor returns no delegations (F9 default fake behavior):
// RecordSignoff now calls this on every invocation to resolve
// domain.ResolveEligibleIdentity's widened eligibility input. Existing
// decision tests exercise direct-actor eligibility only, so an empty result
// preserves their prior behavior unchanged.
func (r *fakeDecisionRepo) LoadActiveDelegationsFor(_ context.Context, _ db.Tx, _, _ string, _ time.Time) ([]domain.Delegation, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Minimal in-memory SQL driver for DecisionService tests.
//
// RecordSignoff calls db.BeginTx and then tx.QueryContext for:
//   1. loadPriorSignoffs   → approval_signoffs WHERE stage_instance_id != active
//   2. loadStageSignoffs   → approval_signoffs WHERE stage_instance_id = active
//
// decisionTestConn is configurable: stageSignoffs contains the rows that
// loadStageSignoffs should return (simulating the just-inserted signoff being
// visible within the same transaction).
// ---------------------------------------------------------------------------

// signoffRow holds the raw column values for one approval_signoffs row.
type signoffRow struct {
	id                 string
	approvalInstanceID string
	stageInstanceID    string
	actorUserID        string
	actorTenantID      string
	decision           string
	comment            string
	signedAt           time.Time
	signatureMethod    string
	signaturePayload   []byte
	contentHash        string
}

type decisionTestConn struct {
	stageSignoffs       []signoffRow // rows returned by loadStageSignoffs
	authzGranted        bool
	authzSet            bool
	areaCode            string
	formDataJSON        string
	contentHashAtSubmit string
	actorID             string
	tenantID            string
	unresolvedComments  int
	execQueries         []string // SQL passed to Exec calls, for assertion
}

type decisionNoopResult struct{}
type decisionEmptyRows struct{}
type decisionSingleValueRows struct {
	value any
	done  bool
}

func (decisionNoopResult) LastInsertId() (int64, error) { return 0, nil }
func (decisionNoopResult) RowsAffected() (int64, error) { return 1, nil }
func (decisionEmptyRows) Columns() []string             { return nil }
func (decisionEmptyRows) Close() error                  { return nil }
func (decisionEmptyRows) Next([]driver.Value) error     { return io.EOF }
func (r *decisionSingleValueRows) Columns() []string    { return []string{"v"} }
func (r *decisionSingleValueRows) Close() error         { return nil }
func (r *decisionSingleValueRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

// signoffRows returns configured signoff rows.
type signoffRows struct {
	rows []signoffRow
	idx  int
}

func (r *signoffRows) Columns() []string {
	return []string{
		"id", "approval_instance_id", "stage_instance_id",
		"actor_user_id", "actor_tenant_id", "decision",
		"comment", "signed_at", "signature_method", "signature_payload", "content_hash",
	}
}
func (r *signoffRows) Close() error { return nil }
func (r *signoffRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	dest[0] = row.id
	dest[1] = row.approvalInstanceID
	dest[2] = row.stageInstanceID
	dest[3] = row.actorUserID
	dest[4] = row.actorTenantID
	dest[5] = row.decision
	dest[6] = row.comment
	dest[7] = row.signedAt
	dest[8] = row.signatureMethod
	dest[9] = row.signaturePayload
	dest[10] = row.contentHash
	return nil
}

type decisionTestStmt struct {
	conn  *decisionTestConn
	query string
}

func (s *decisionTestStmt) Close() error  { return nil }
func (s *decisionTestStmt) NumInput() int { return -1 }
func (s *decisionTestStmt) Exec(_ []driver.Value) (driver.Result, error) {
	s.conn.execQueries = append(s.conn.execQueries, s.query)
	return decisionNoopResult{}, nil
}
func (s *decisionTestStmt) Query(_ []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "form_data_json") && strings.Contains(q, "from documents") {
		return &decisionSingleValueRows{value: []byte(s.conn.formDataJSON)}, nil
	}
	if strings.Contains(q, "content_hash_at_submit") && strings.Contains(q, "from documents") {
		hash := s.conn.contentHashAtSubmit
		if hash == "" {
			hash = validContentHash
		}
		return &decisionSingleValueRows{value: hash}, nil
	}
	if strings.Contains(q, "from documents") {
		return &docAreaRows{snapshot: s.conn.areaCode}, nil
	}
	if strings.Contains(q, "from document_comments") {
		return &decisionSingleValueRows{value: s.conn.unresolvedComments}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "iam_user_roles") {
		return &decisionSingleValueRows{value: false}, nil // non-admin
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "role_capabilities") {
		return &decisionSingleValueRows{value: s.conn.authzGranted}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.asserted_caps'") {
		return &decisionSingleValueRows{value: nil}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.tenant_id'") {
		return &decisionSingleValueRows{value: s.conn.tenantID}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.actor_id'") {
		return &decisionSingleValueRows{value: s.conn.actorID}, nil
	}
	// loadStageSignoffs queries WHERE stage_instance_id = $1 (no "!=")
	// loadPriorSignoffs queries WHERE stage_instance_id != $2
	// Both hit "approval_signoffs" table.
	// We return configured stageSignoffs for the stage query and empty for prior.
	if strings.Contains(q, "approval_signoffs") && isStageQuery(s.query) {
		return &signoffRows{rows: s.conn.stageSignoffs}, nil
	}
	return decisionEmptyRows{}, nil
}

// isStageQuery returns true when the query is for a single stage (no "!=" exclusion).
func isStageQuery(q string) bool {
	// loadStageSignoffs: "WHERE stage_instance_id = $1" — no "!="
	// loadPriorSignoffs: "WHERE ... AND stage_instance_id != $2"
	for i := 0; i < len(q)-1; i++ {
		if q[i] == '!' && q[i+1] == '=' {
			return false
		}
	}
	return true
}

func (c *decisionTestConn) Prepare(query string) (driver.Stmt, error) {
	return &decisionTestStmt{conn: c, query: query}, nil
}
func (c *decisionTestConn) Close() error              { return nil }
func (c *decisionTestConn) Begin() (driver.Tx, error) { return c, nil }
func (c *decisionTestConn) Commit() error             { return nil }
func (c *decisionTestConn) Rollback() error           { return nil }

type decisionTestDriver struct{ conn *decisionTestConn }

func (d *decisionTestDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

var decisionDBCounter int

func newDecisionTestDB(t *testing.T, conn *decisionTestConn) *sql.DB {
	t.Helper()
	if conn.areaCode == "" {
		conn.areaCode = "QA"
	}
	if conn.formDataJSON == "" {
		conn.formDataJSON = `{"title":"Doc"}`
	}
	if conn.actorID == "" {
		conn.actorID = "user-1"
	}
	if conn.tenantID == "" {
		conn.tenantID = tenant.DevTenantID
	}
	if !conn.authzSet {
		conn.authzGranted = true
	}
	decisionDBCounter++
	name := fmt.Sprintf("decision_test_%d", decisionDBCounter)
	sql.Register(name, &decisionTestDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open decision test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildSingleStageInstance returns an Instance with one active stage using
// any_1_of quorum and the given eligible actors. FrozenContentHash is set to
// validContentHash (F5/F6): by the time an approval-kind stage is active and
// signoff is possible, the instance must already be frozen, so every fixture
// that exercises a real signoff path needs a non-nil pin — tests that
// specifically want the fail-closed NULL-pin path override this field.
func buildSingleStageInstance(instanceID, stageID, authorUserID string, eligible []string) *domain.Instance {
	now := time.Now().UTC()
	frozenHash := validContentHash
	return &domain.Instance{
		ID:                  instanceID,
		TenantID:            "tenant-1",
		DocumentID:          "doc-1",
		Subject:             domain.NewDocumentSubject("doc-1"),
		RouteID:             "route-1",
		Status:              domain.InstanceInProgress,
		SubmittedBy:         authorUserID,
		SubmittedAt:         now,
		RevisionVersion:     1,
		ContentHashAtSubmit: validContentHash,
		FrozenContentHash:   &frozenHash,
		Stages: []domain.StageInstance{
			{
				ID:                         stageID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 1,
				NameSnapshot:               "QA Review",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				EligibleActorIDs:           eligible,
				Status:                     domain.StageActive,
				OpenedAt:                   &now,
			},
		},
	}
}

// buildTwoApproverInstance returns an Instance with one active stage using
// all_of quorum and two eligible actors.
func buildTwoApproverInstance(instanceID, stageID, authorUserID string, eligible []string) *domain.Instance {
	inst := buildSingleStageInstance(instanceID, stageID, authorUserID, eligible)
	inst.Stages[0].QuorumSnapshot = domain.QuorumAllOf
	return inst
}

// buildReviewThenStageInstance returns a two-stage instance: an active
// any_1_of review-kind stage (reviewStageID) followed by a pending stage of
// the given kind (nextStageID). Used by S1's fast-forward eligibility tests
// (R5, unit 2.3 G3) to exercise the QuorumApprovedStage -> AdvanceStage ->
// next-active-stage eligibility probe.
func buildReviewThenStageInstance(instanceID, reviewStageID, nextStageID, authorUserID string, reviewEligible, nextEligible []string, nextKind domain.StageKind) *domain.Instance {
	now := time.Now().UTC()
	frozenHash := validContentHash
	return &domain.Instance{
		ID:                  instanceID,
		TenantID:            "tenant-1",
		DocumentID:          "doc-1",
		Subject:             domain.NewDocumentSubject("doc-1"),
		RouteID:             "route-1",
		Status:              domain.InstanceInProgress,
		SubmittedBy:         authorUserID,
		SubmittedAt:         now,
		RevisionVersion:     1,
		ContentHashAtSubmit: validContentHash,
		FrozenContentHash:   &frozenHash,
		Stages: []domain.StageInstance{
			{
				ID:                         reviewStageID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 1,
				NameSnapshot:               "Review",
				QuorumSnapshot:             domain.QuorumAny1Of,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				Kind:                       domain.StageKindReview,
				EligibleActorIDs:           reviewEligible,
				Status:                     domain.StageActive,
				OpenedAt:                   &now,
			},
			{
				ID:                         nextStageID,
				ApprovalInstanceID:         instanceID,
				StageOrder:                 2,
				NameSnapshot:               "Next",
				QuorumSnapshot:             domain.QuorumAllOf,
				OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
				Kind:                       nextKind,
				EligibleActorIDs:           nextEligible,
				Status:                     domain.StagePending,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// validContentHash is a 64-char lowercase hex string used in test signoff rows.
const validContentHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// validContentHashPtr is an addressable copy of validContentHash (F6): several
// fixtures across this package need a *string for domain.Instance.FrozenContentHash,
// and Go cannot take the address of a string constant directly.
var validContentHashPtr = validContentHash

// TestRecordSignoff_ApprovePath_QuorumMet: single approver quorum (any_1_of).
// Expect StageCompleted=true, InstanceApproved=true (only 1 stage).
func TestRecordSignoff_ApprovePath_QuorumMet(t *testing.T) {
	const (
		instanceID = "inst-1"
		stageID    = "stage-1"
		actorID    = "approver-1"
		authorID   = "author-1"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})

	// Simulate the just-inserted signoff being visible in loadStageSignoffs.
	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	stageSignoffs := []signoffRow{
		{
			id:                 "signoff-1",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "approve",
			comment:            "LGTM",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		},
	}

	conn := &decisionTestConn{
		stageSignoffs: stageSignoffs,
		authzGranted:  true,
		areaCode:      "QA",
		actorID:       actorID,
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-1", WasReplay: false},
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: signedAt}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: clock, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		Comment:          "LGTM",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}

	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if !result.StageCompleted {
		t.Error("expected StageCompleted=true")
	}
	if !result.InstanceApproved {
		t.Error("expected InstanceApproved=true (single-stage instance)")
	}
	if result.InstanceRejected {
		t.Error("expected InstanceRejected=false")
	}
	// F-QA4-7: the approval_signoffs id InsertSignoff returned must be threaded
	// out on SignoffResult so the HTTP layer can put it on the wire.
	if result.SignoffID != "signoff-1" {
		t.Errorf("SignoffID = %q; want %q", result.SignoffID, "signoff-1")
	}
	if len(emitter.Events) != 1 {
		t.Errorf("expected 1 governance event; got %d", len(emitter.Events))
	}
	if emitter.Events[0].EventType != "signoff_recorded" {
		t.Errorf("event type = %q; want %q", emitter.Events[0].EventType, "signoff_recorded")
	}
}

// TestRecordSignoff_DBLevelReplayCarriesOriginalSignoffID pins F-QA4-7 on the
// ON CONFLICT branch: when InsertSignoff detects an existing matching row it
// returns WasReplay=true plus the ORIGINAL id, and RecordSignoff must surface
// that id rather than the pre-fix empty SignoffResult{}. The neutral
// stage/instance flags stay false — the stage must not advance twice.
func TestRecordSignoff_DBLevelReplayCarriesOriginalSignoffID(t *testing.T) {
	const (
		instanceID = "inst-replay"
		stageID    = "stage-replay"
		actorID    = "approver-1"
		authorID   = "author-1"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)

	conn := &decisionTestConn{authzGranted: true, areaCode: "QA", actorID: actorID}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-original", WasReplay: true},
	}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		Comment:          "LGTM",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	})
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if result.SignoffID != "signoff-original" {
		t.Errorf("SignoffID = %q; want %q (original row id on replay)", result.SignoffID, "signoff-original")
	}
	if result.StageCompleted || result.InstanceApproved || result.InstanceRejected {
		t.Errorf("replay must not re-advance the stage/instance; got %+v", result)
	}
}

// TestRecordSignoff_ApprovePath_QuorumNotYetMet: all_of quorum with two eligible actors.
// First signoff (only one of two) should leave StageCompleted=false.
func TestRecordSignoff_ApprovePath_QuorumNotYetMet(t *testing.T) {
	const (
		instanceID = "inst-2"
		stageID    = "stage-2"
		actorID    = "approver-1"
		authorID   = "author-2"
	)

	eligible := []string{"approver-1", "approver-2"}
	inst := buildTwoApproverInstance(instanceID, stageID, authorID, eligible)

	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	// Only one signoff visible — quorum all_of needs 2.
	stageSignoffs := []signoffRow{
		{
			id:                 "signoff-2",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "approve",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		},
	}

	conn := &decisionTestConn{
		stageSignoffs: stageSignoffs,
		authzGranted:  true,
		areaCode:      "QA",
		actorID:       actorID,
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-2", WasReplay: false},
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: signedAt}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: clock, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     actorID,
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}

	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if result.StageCompleted {
		t.Error("expected StageCompleted=false — quorum not yet met (1 of 2 approved)")
	}
	if result.InstanceApproved {
		t.Error("expected InstanceApproved=false")
	}
}

// TestRecordSignoff_ContentHashEchoesInstanceSubmitHash: the persisted signoff content_hash
// must equal instance.ContentHashAtSubmit (the value submit stored and that the FE echoes back).
// Signoff does not recompute over current document form data — that diverges from submit
// canonicalization and breaks the FE content-pin contract.
func TestRecordSignoff_ContentHashEchoesInstanceSubmitHash(t *testing.T) {
	const (
		instanceID = "inst-db-hash"
		stageID    = "stage-db-hash"
		actorID    = "approver-db-hash"
		authorID   = "author-db-hash"
	)

	inst := buildTwoApproverInstance(instanceID, stageID, authorID, []string{actorID, "approver-2"})
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{
		repo:       repo,
		emitter:    &MemoryEmitter{},
		clock:      fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		pinInvoker: &fakePinInvoker{},
	}
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignaturePayload: map[string]any{},
		ContentFormData:  map[string]any{"_content_hash": inst.ContentHashAtSubmit},
	})
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if repo.insertedSignoff == nil {
		t.Fatal("expected inserted signoff")
	}
	if got := repo.insertedSignoff.ContentHash(); got != inst.ContentHashAtSubmit {
		t.Errorf("ContentHash() = %q; want instance.ContentHashAtSubmit %q", got, inst.ContentHashAtSubmit)
	}
}

// TestRecordSignoff_ApproveSetsSignatureMeaningApproval asserts an approve
// decision persists signature_meaning = "approval" (F7, W8, 21 CFR
// 11.50(a)(3) — the signed record must state the meaning of the signature).
func TestRecordSignoff_ApproveSetsSignatureMeaningApproval(t *testing.T) {
	const (
		instanceID = "inst-sigmeaning-approve"
		stageID    = "stage-sigmeaning-approve"
		actorID    = "approver-sigmeaning-approve"
		authorID   = "author-sigmeaning-approve"
	)

	inst := buildTwoApproverInstance(instanceID, stageID, authorID, []string{actorID, "approver-2"})
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{
		repo:       repo,
		emitter:    &MemoryEmitter{},
		clock:      fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		pinInvoker: &fakePinInvoker{},
	}
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignaturePayload: map[string]any{},
		ContentFormData:  map[string]any{"_content_hash": inst.ContentHashAtSubmit},
	})
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if repo.insertedSignoff == nil {
		t.Fatal("expected inserted signoff")
	}
	if got := repo.insertedSignoff.SignatureMeaning(); got != "approval" {
		t.Errorf("SignatureMeaning() = %q; want %q", got, "approval")
	}
}

func TestRecordSignoff_ThreadsActorDisplayNameFromRepo(t *testing.T) {
	const (
		instanceID = "inst-dn"
		stageID    = "stage-dn"
		actorID    = "approver-dn"
		authorID   = "author-dn"
	)

	inst := buildTwoApproverInstance(instanceID, stageID, authorID, []string{actorID, "approver-2"})
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst, actorDisplayName: "Alice Approver"}
	svc := &DecisionService{
		repo:       repo,
		emitter:    &MemoryEmitter{},
		clock:      fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		pinInvoker: &fakePinInvoker{},
	}
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignaturePayload: map[string]any{},
		ContentFormData:  map[string]any{"_content_hash": inst.ContentHashAtSubmit},
	})
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if repo.insertedSignoff == nil {
		t.Fatal("expected inserted signoff")
	}
	if got := repo.insertedSignoff.ActorDisplayNameSnapshot(); got != "Alice Approver" {
		t.Errorf("ActorDisplayNameSnapshot() = %q; want %q (must be threaded from repo.LoadActorDisplayName)", got, "Alice Approver")
	}
	if !repo.loadActorDisplayNameCalled {
		t.Error("expected LoadActorDisplayName to be called; it was not")
	}
	if repo.loadActorDisplayNameTenant != "tenant-1" {
		t.Errorf("LoadActorDisplayName tenantID = %q; want %q", repo.loadActorDisplayNameTenant, "tenant-1")
	}
	if repo.loadActorDisplayNameUser != actorID {
		t.Errorf("LoadActorDisplayName userID = %q; want %q", repo.loadActorDisplayNameUser, actorID)
	}
}

func TestRecordSignoff_ContentHashMismatchFailsBeforePersisting(t *testing.T) {
	const (
		instanceID = "inst-hash-mismatch"
		stageID    = "stage-hash-mismatch"
		actorID    = "approver-hash-mismatch"
		authorID   = "author-hash-mismatch"
	)

	inst := buildTwoApproverInstance(instanceID, stageID, authorID, []string{actorID, "approver-2"})
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{
		repo:       repo,
		emitter:    &MemoryEmitter{},
		clock:      fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		pinInvoker: &fakePinInvoker{},
	}
	db := newDecisionTestDB(t, conn)

	// Client echoes a hash that does NOT match instance.ContentHashAtSubmit.
	const drifted = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignaturePayload: map[string]any{},
		ContentFormData:  map[string]any{"_content_hash": drifted},
	})
	if !errors.Is(err, ErrContentHashMismatch) {
		t.Fatalf("expected ErrContentHashMismatch, got %v", err)
	}
	if repo.insertedSignoff != nil {
		t.Fatal("signoff must not be persisted on client/server content hash mismatch")
	}
}

// TestRecordSignoff_NullFrozenHash_FailsClosed (F6, spec §11 no-fallback
// principle): an instance whose FrozenContentHash is nil (structurally
// impossible via the normal API path per F5 stage-kind gating, but
// constructible directly for this test) must fail closed with
// ErrContentHashMismatch — it must NEVER fall through to any other hash
// source (e.g. ContentHashAtSubmit or a head document_revisions hash), even
// though ContentHashAtSubmit is set and the client echoes a hash that would
// have matched it under the old (now-deleted) COALESCE behavior.
func TestRecordSignoff_NullFrozenHash_FailsClosed(t *testing.T) {
	const (
		instanceID = "inst-null-frozen-hash"
		stageID    = "stage-null-frozen-hash"
		actorID    = "approver-null-frozen-hash"
		authorID   = "author-null-frozen-hash"
	)

	inst := buildTwoApproverInstance(instanceID, stageID, authorID, []string{actorID, "approver-2"})
	inst.FrozenContentHash = nil // never frozen — the fail-closed case under test.

	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	svc := &DecisionService{
		repo:       repo,
		emitter:    &MemoryEmitter{},
		clock:      fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		pinInvoker: &fakePinInvoker{},
	}
	db := newDecisionTestDB(t, conn)

	// Client echoes exactly ContentHashAtSubmit — which would have matched the
	// OLD (deleted) COALESCE-over-documents-table behavior. It must NOT match now.
	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignaturePayload: map[string]any{},
		ContentFormData:  map[string]any{"_content_hash": inst.ContentHashAtSubmit},
	})
	if !errors.Is(err, ErrContentHashMismatch) {
		t.Fatalf("expected ErrContentHashMismatch (fail-closed on nil FrozenContentHash), got %v", err)
	}
	if repo.insertedSignoff != nil {
		t.Fatal("signoff must not be persisted when the frozen pin is absent")
	}
}

// TestRecordSignoff_RejectPath: single actor rejects on any_1_of stage.
// Expect InstanceRejected=true, StageCompleted=false (stage is rejected_here, not completed),
// and a single "signoff_recorded" governance event emitted.
func TestRecordSignoff_RejectPath(t *testing.T) {
	const (
		instanceID = "inst-reject-1"
		stageID    = "stage-reject-1"
		actorID    = "approver-r1"
		authorID   = "author-r1"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})

	signedAt := time.Date(2026, 4, 22, 14, 0, 0, 0, time.UTC)
	// Simulate the reject signoff visible in loadStageSignoffs after insert.
	stageSignoffs := []signoffRow{
		{
			id:                 "signoff-r1",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "reject",
			comment:            "Not acceptable",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		},
	}

	conn := &decisionTestConn{
		stageSignoffs: stageSignoffs,
		authzGranted:  true,
		areaCode:      "QA",
		actorID:       actorID,
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-r1", WasReplay: false},
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: signedAt}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: clock, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "reject",
		Comment:          "Not acceptable",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "def"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}

	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if result.StageCompleted {
		t.Error("expected StageCompleted=false — rejected stage is rejected_here, not completed")
	}
	if result.InstanceApproved {
		t.Error("expected InstanceApproved=false")
	}
	if !result.InstanceRejected {
		t.Error("expected InstanceRejected=true")
	}
	if len(emitter.Events) != 1 {
		t.Errorf("expected 1 governance event; got %d", len(emitter.Events))
	}
	if emitter.Events[0].EventType != "signoff_recorded" {
		t.Errorf("event type = %q; want %q", emitter.Events[0].EventType, "signoff_recorded")
	}
	cancelGUCSet := false
	docDrafted := false
	docRejected := false
	for _, q := range conn.execQueries {
		ql := strings.ToLower(q)
		if strings.Contains(ql, "set_config('metaldocs.cancel_in_progress'") {
			cancelGUCSet = true
		}
		if strings.Contains(ql, "update documents") && strings.Contains(ql, "'draft'") {
			docDrafted = true
		}
		if strings.Contains(ql, "update documents") && strings.Contains(ql, "'rejected'") {
			docRejected = true
		}
	}
	if !cancelGUCSet {
		t.Error("expected set_config('metaldocs.cancel_in_progress', ...) on rejection path")
	}
	if !docDrafted {
		t.Error("expected UPDATE documents SET status='draft' on rejection path")
	}
	if docRejected {
		t.Error("did not expect UPDATE documents SET status='rejected' on rejection path")
	}
	if repo.instanceStatusTo != domain.InstanceRejected {
		t.Errorf("approval instance status = %q; want %q", repo.instanceStatusTo, domain.InstanceRejected)
	}
	if repo.instanceStatusFrom != domain.InstanceInProgress {
		t.Errorf("approval instance previous status = %q; want %q", repo.instanceStatusFrom, domain.InstanceInProgress)
	}
}

// TestRecordSignoff_RejectSetsSignatureMeaningRejection asserts a reject
// decision persists signature_meaning = "rejection", NOT the domain default
// "approval" (F7, W8, 21 CFR 11.50(a)(3) — the pre-F7 defect: RecordSignoff
// never set SignatureMeaning, so NewSignoff's empty-field default silently
// stamped every reject signoff with "approval", a false attestation).
func TestRecordSignoff_RejectSetsSignatureMeaningRejection(t *testing.T) {
	const (
		instanceID = "inst-sigmeaning-reject"
		stageID    = "stage-sigmeaning-reject"
		actorID    = "approver-sigmeaning-reject"
		authorID   = "author-sigmeaning-reject"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})

	signedAt := time.Date(2026, 4, 22, 14, 0, 0, 0, time.UTC)
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-sigmeaning-reject", WasReplay: false},
	}
	svc := &DecisionService{repo: repo, emitter: &MemoryEmitter{}, clock: fixedClock{t: signedAt}, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "reject",
		Comment:          "Not acceptable",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "def"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	})
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if repo.insertedSignoff == nil {
		t.Fatal("expected inserted signoff")
	}
	if got := repo.insertedSignoff.SignatureMeaning(); got != "rejection" {
		t.Errorf("SignatureMeaning() = %q; want %q (must not default to approval on reject)", got, "rejection")
	}
}

// TestRecordSignoff_SoDViolation: actor == document author → ErrAuthorCannotSign.
func TestRecordSignoff_SoDViolation(t *testing.T) {
	const (
		instanceID = "inst-3"
		stageID    = "stage-3"
		authorID   = "author-and-actor" // same person!
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{authorID})

	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      authorID,
	} // SoD check fires before signoff SQL reads
	repo := &fakeDecisionRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: clock, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     authorID, // same as SubmittedBy
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected error for SoD violation; got nil")
	}
	if !errors.Is(err, domain.ErrAuthorCannotSign) {
		t.Errorf("expected domain.ErrAuthorCannotSign; got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on SoD violation; got %d", len(emitter.Events))
	}
}

// TestRecordSignoff_RejectsNonEligibleActor: actor not in EligibleActorIDs → ErrActorNotEligible
// and a signoff.rejected governance event with Reason="not_eligible".
func TestRecordSignoff_RejectsNonEligibleActor(t *testing.T) {
	const (
		instanceID = "inst-elig-1"
		stageID    = "stage-elig-1"
		actorID    = "bob" // NOT in eligible list
		authorID   = "alice"
	)

	// eligible list contains only "alice" — bob is the author but also not eligible as signoff actor
	// Use a separate author so SoD doesn't fire first. bob is non-eligible.
	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{"alice"})

	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: clock, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     actorID,
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected ErrActorNotEligible; got nil")
	}
	if !errors.Is(err, domain.ErrActorNotEligible) {
		t.Fatalf("expected domain.ErrActorNotEligible; got %v", err)
	}

	// Expect one audit governance event with kind=signoff.rejected, reason=not_eligible.
	if len(emitter.Events) != 1 {
		t.Fatalf("expected 1 audit event; got %d", len(emitter.Events))
	}
	ev := emitter.Events[0]
	if ev.EventType != "signoff.rejected" {
		t.Errorf("event type = %q; want %q", ev.EventType, "signoff.rejected")
	}
	if ev.Reason != "not_eligible" {
		t.Errorf("event reason = %q; want %q", ev.Reason, "not_eligible")
	}
	if ev.ActorUserID != actorID {
		t.Errorf("event actor = %q; want %q", ev.ActorUserID, actorID)
	}
}

func TestRecordSignoff_CapabilityDenied(t *testing.T) {
	const (
		instanceID = "inst-authz-1"
		stageID    = "stage-authz-1"
		actorID    = "approver-authz-1"
		authorID   = "author-authz-1"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})

	conn := &decisionTestConn{
		authzGranted: false,
		authzSet:     true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}
	svc := &DecisionService{repo: repo, emitter: emitter, clock: clock, pinInvoker: &fakePinInvoker{}}
	db := newDecisionTestDB(t, conn)

	req := SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     actorID,
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	}

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected ErrCapabilityDenied; got nil")
	}
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("expected ErrCapabilityDenied; got %v", err)
	}
	if denied.Capability != "document.signoff" {
		t.Errorf("capability = %q; want %q", denied.Capability, "document.signoff")
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on denied capability; got %d", len(emitter.Events))
	}
}

// TestRecordSignoff_FinalApprove_NoLongerGatedByUnresolvedComments is F5's
// regression test for the W10 removal: decision_service.go used to block
// final-approve on ANY unresolved document-wide comment
// (ErrApprovalBlockedByUnresolvedComments). That gate is now gone —
// freeze (fired earlier via ReviewVerdictService's stage-advance path, or at
// submit for approval-only routes) is the sole enforcement of this concern,
// scoped to the instance's own review window rather than the whole document's
// history. This test proves RecordSignoff completes the instance even with
// unresolvedComments > 0 on the fake conn — the old test asserted the
// opposite (blocked) behavior; that assertion is gone along with the gate.
func TestRecordSignoff_FinalApprove_NoLongerGatedByUnresolvedComments(t *testing.T) {
	const (
		instanceID = "inst-comments-1"
		stageID    = "stage-comments-1"
		actorID    = "approver-comments-1"
		authorID   = "author-comments-1"
	)

	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	signedAt := time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC)
	stageSignoffs := []signoffRow{
		{
			id:                 "signoff-comments-1",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "approve",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		},
	}

	conn := &decisionTestConn{
		stageSignoffs:      stageSignoffs,
		authzGranted:       true,
		areaCode:           "QA",
		actorID:            actorID,
		unresolvedComments: 2,
	}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "signoff-comments-1", WasReplay: false},
	}
	emitter := &MemoryEmitter{}
	pin := &fakePinInvoker{}
	svc := &DecisionService{
		repo:       repo,
		emitter:    emitter,
		clock:      fixedClock{t: signedAt},
		pinInvoker: pin,
	}
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	})
	if err != nil {
		t.Fatalf("RecordSignoff() error = %v, want nil (comments gate removed from decision_service.go)", err)
	}
	if !result.InstanceApproved {
		t.Fatalf("expected instance approved, got %+v", result)
	}
	if pin.calls != 1 {
		t.Fatalf("pin should still run once approval completes, got %d calls", pin.calls)
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("expected 1 governance event on successful approval, got %d", len(emitter.Events))
	}
}
