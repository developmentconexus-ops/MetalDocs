package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// fakeReleaseRecorder stands in for the ADR 0085 terminal-approval seam
// (approval fact + async-freeze pin + coordinator evaluation). Substituting it
// here keeps the release substrate's SQL out of these driver-level decision
// tests; the real recorder is exercised by the release integration suite.
type fakeReleaseRecorder struct {
	calls int
	err   error
	last  TerminalApprovalInput
}

func (f *fakeReleaseRecorder) RecordTerminalApproval(_ context.Context, _ db.Tx, in TerminalApprovalInput) (string, error) {
	f.calls++
	f.last = in
	if f.err != nil {
		return "", f.err
	}
	return "gen-" + in.InstanceID, nil
}

type freezeDecisionConn struct {
	stageSignoffs      []signoffRow
	authzGranted       bool
	authzSet           bool
	areaCode           string
	actorID            string
	tenantID           string
	unresolvedComments int
	assertedCaps       string

	documentStatus string
	pendingStatus  *string
	committed      bool
	rolledBack     bool
}

type freezeDecisionStmt struct {
	conn  *freezeDecisionConn
	query string
}

type freezeDecisionNoopResult struct{}
type freezeDecisionEmptyRows struct{}
type freezeDecisionSingleValueRows struct {
	value any
	done  bool
}

func (freezeDecisionNoopResult) LastInsertId() (int64, error) { return 0, nil }
func (freezeDecisionNoopResult) RowsAffected() (int64, error) { return 1, nil }
func (freezeDecisionEmptyRows) Columns() []string             { return nil }
func (freezeDecisionEmptyRows) Close() error                  { return nil }
func (freezeDecisionEmptyRows) Next([]driver.Value) error     { return io.EOF }
func (r *freezeDecisionSingleValueRows) Columns() []string    { return []string{"v"} }
func (r *freezeDecisionSingleValueRows) Close() error         { return nil }
func (r *freezeDecisionSingleValueRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

func (s *freezeDecisionStmt) Close() error  { return nil }
func (s *freezeDecisionStmt) NumInput() int { return -1 }
func (s *freezeDecisionStmt) Exec(args []driver.Value) (driver.Result, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "set_config('metaldocs.asserted_caps'") && len(args) > 0 {
		if raw, ok := args[0].(string); ok {
			s.conn.assertedCaps = raw
		}
	}
	if strings.Contains(q, "update documents") {
		if !strings.Contains(s.conn.assertedCaps, `"cap":"document.edit"`) {
			return nil, errors.New("ErrCapabilityNotAsserted: none of {document.edit} present in asserted_caps on documents")
		}
		if strings.Contains(q, "set status") && strings.Contains(q, "'approved'") {
			status := "approved"
			s.conn.pendingStatus = &status
		}
		if strings.Contains(q, "set status") && strings.Contains(q, "'draft'") {
			status := "draft"
			s.conn.pendingStatus = &status
		}
	}
	return freezeDecisionNoopResult{}, nil
}
func (s *freezeDecisionStmt) Query(_ []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "form_data_json") && strings.Contains(q, "from documents") {
		return &freezeDecisionSingleValueRows{value: []byte(`{"title":"Doc"}`)}, nil
	}
	if strings.Contains(q, "content_hash_at_submit") && strings.Contains(q, "from documents") {
		return &freezeDecisionSingleValueRows{value: validContentHash}, nil
	}
	if strings.Contains(q, "from documents") {
		return &docAreaRows{snapshot: s.conn.areaCode}, nil
	}
	if strings.Contains(q, "from document_comments") {
		return &freezeDecisionSingleValueRows{value: s.conn.unresolvedComments}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "iam_user_roles") {
		return &freezeDecisionSingleValueRows{value: false}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "role_capabilities") {
		return &freezeDecisionSingleValueRows{value: s.conn.authzGranted}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.asserted_caps'") {
		if s.conn.assertedCaps == "" {
			return &freezeDecisionSingleValueRows{value: nil}, nil
		}
		return &freezeDecisionSingleValueRows{value: s.conn.assertedCaps}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.tenant_id'") {
		return &freezeDecisionSingleValueRows{value: s.conn.tenantID}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.actor_id'") {
		return &freezeDecisionSingleValueRows{value: s.conn.actorID}, nil
	}
	if strings.Contains(q, "approval_signoffs") && isStageQuery(s.query) {
		return &signoffRows{rows: s.conn.stageSignoffs}, nil
	}
	return freezeDecisionEmptyRows{}, nil
}

func (c *freezeDecisionConn) Prepare(query string) (driver.Stmt, error) {
	return &freezeDecisionStmt{conn: c, query: query}, nil
}
func (c *freezeDecisionConn) Close() error              { return nil }
func (c *freezeDecisionConn) Begin() (driver.Tx, error) { return c, nil }
func (c *freezeDecisionConn) Commit() error {
	c.committed = true
	if c.pendingStatus != nil {
		c.documentStatus = *c.pendingStatus
	}
	c.pendingStatus = nil
	return nil
}
func (c *freezeDecisionConn) Rollback() error {
	c.rolledBack = true
	c.pendingStatus = nil
	return nil
}

type freezeDecisionDriver struct{ conn *freezeDecisionConn }

func (d *freezeDecisionDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

var freezeDecisionDBCounter int

func newFreezeDecisionTestDB(t *testing.T, conn *freezeDecisionConn) *sql.DB {
	t.Helper()
	if conn.areaCode == "" {
		conn.areaCode = "QA"
	}
	if conn.actorID == "" {
		conn.actorID = "approver-1"
	}
	if conn.tenantID == "" {
		conn.tenantID = tenant.DevTenantID
	}
	if !conn.authzSet {
		conn.authzGranted = true
	}
	if conn.documentStatus == "" {
		conn.documentStatus = "under_review"
	}
	freezeDecisionDBCounter++
	name := fmt.Sprintf("decision_freeze_test_%d", freezeDecisionDBCounter)
	sql.Register(name, &freezeDecisionDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open decision freeze test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestRecordSignoff_PinError_RollsBackTransaction(t *testing.T) {
	const (
		instanceID = "inst-pin-b"
		stageID    = "stage-pin-b"
		actorID    = "approver-1"
		authorID   = "author-1"
	)
	signedAt := time.Date(2026, 4, 23, 12, 10, 0, 0, time.UTC)
	repo := &fakeDecisionRepo{
		instance:         buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID}),
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "sig-b", WasReplay: false},
	}
	pin := &fakeReleaseRecorder{err: errors.New("pin failed")}
	conn := &freezeDecisionConn{
		actorID: actorID,
		stageSignoffs: []signoffRow{{
			id:                 "sig-b",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "approve",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		}},
	}
	db := newFreezeDecisionTestDB(t, conn)
	svc := (&DecisionService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: signedAt},
	}).WithReleaseRecorder(pin)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	})
	if err == nil {
		t.Fatal("expected pin error")
	}
	if conn.committed {
		t.Fatal("transaction should not commit on pin error")
	}
	if !conn.rolledBack {
		t.Fatal("transaction should roll back on pin error")
	}
	if conn.documentStatus != "under_review" {
		t.Fatalf("document status should stay under_review, got %q", conn.documentStatus)
	}
}

// TestRecordSignoff_PDFDispatchError_IsBestEffort was deleted: it existed solely
// to cover the deprecated post-commit pdfDispatcher.Dispatch path, which has been
// removed. The document-approve path Pins via the async-freeze seam (ADR 0015);
// MaterializeJobRunner is the sole pdf producer. The old synchronous in-tx
// pdf-dispatch seam on DecisionService was structurally dead and removed in QR-C.

func TestRecordSignoff_WasReplay_DoesNotPin(t *testing.T) {
	const (
		instanceID = "inst-pin-d"
		stageID    = "stage-pin-d"
		actorID    = "approver-1"
		authorID   = "author-1"
	)
	signedAt := time.Date(2026, 4, 23, 12, 30, 0, 0, time.UTC)
	repo := &fakeDecisionRepo{
		instance:         buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID}),
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "sig-d", WasReplay: true},
	}
	pin := &fakeReleaseRecorder{}
	conn := &freezeDecisionConn{actorID: actorID}
	db := newFreezeDecisionTestDB(t, conn)
	svc := (&DecisionService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: signedAt},
	}).WithReleaseRecorder(pin)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:        "tenant-1",
		InstanceID:      instanceID,
		StageInstanceID: stageID,
		ActorUserID:     actorID,
		Decision:        "approve",
		ContentFormData: map[string]any{"title": "Doc", "_content_hash": validContentHash},
	})
	if err != nil {
		t.Fatalf("RecordSignoff() error = %v", err)
	}
	if pin.calls != 0 {
		t.Fatalf("Pin must not run on replay, got %d call(s)", pin.calls)
	}
}

// TestRecordSignoff_UnresolvedComments_NoLongerBlocksApprove is F5's
// regression test mirroring decision_service_test.go's: the
// s.repo.HasUnresolvedComments gate that used to roll back final-approve here
// is removed (W10). freezeDecisionConn.unresolvedComments > 0 no longer
// blocks — decision_service.go doesn't consult it at all anymore.
func TestRecordSignoff_UnresolvedComments_NoLongerBlocksApprove(t *testing.T) {
	const (
		instanceID = "inst-freeze-comments"
		stageID    = "stage-freeze-comments"
		actorID    = "approver-1"
		authorID   = "author-1"
	)
	signedAt := time.Date(2026, 4, 23, 12, 40, 0, 0, time.UTC)
	repo := &fakeDecisionRepo{
		instance:         buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID}),
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "sig-comments", WasReplay: false},
	}
	pin := &fakeReleaseRecorder{}
	conn := &freezeDecisionConn{
		actorID:            actorID,
		unresolvedComments: 1,
		stageSignoffs: []signoffRow{{
			id:                 "sig-comments",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "approve",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		}},
	}
	db := newFreezeDecisionTestDB(t, conn)
	svc := (&DecisionService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: signedAt},
	}).WithReleaseRecorder(pin)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
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
	if !conn.committed {
		t.Fatal("transaction should commit — comments no longer block approval")
	}
	if conn.rolledBack {
		t.Fatal("transaction should not roll back")
	}
	if conn.documentStatus != "approved" {
		t.Fatalf("document status should be approved, got %q", conn.documentStatus)
	}
	if pin.calls != 1 {
		t.Fatalf("pin should run once approval completes, got %d calls", pin.calls)
	}
}

func TestRecordSignoff_PinInvoker_CallsPin(t *testing.T) {
	const (
		instanceID = "inst-pin-1"
		stageID    = "stage-pin-1"
		actorID    = "approver-1"
		authorID   = "author-1"
	)
	signedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	repo := &fakeDecisionRepo{
		instance:         buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID}),
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "sig-pin-1", WasReplay: false},
	}
	pin := &fakeReleaseRecorder{}
	conn := &freezeDecisionConn{
		actorID: actorID,
		stageSignoffs: []signoffRow{{
			id:                 "sig-pin-1",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "approve",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		}},
	}
	db := newFreezeDecisionTestDB(t, conn)
	svc := (&DecisionService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: signedAt},
	}).WithReleaseRecorder(pin)

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
		t.Fatalf("RecordSignoff() error = %v", err)
	}
	if !result.InstanceApproved || conn.documentStatus != "approved" {
		t.Fatalf("expected approved document, status=%q result=%+v", conn.documentStatus, result)
	}
	if pin.calls != 1 {
		t.Fatalf("Pin should be called once, got %d", pin.calls)
	}
}

func TestRecordSignoff_RejectPath_AssertsDocumentEditBeforeDocumentWrite(t *testing.T) {
	const (
		instanceID = "inst-freeze-reject"
		stageID    = "stage-freeze-reject"
		actorID    = "approver-1"
		authorID   = "author-1"
	)
	signedAt := time.Date(2026, 4, 23, 12, 50, 0, 0, time.UTC)
	repo := &fakeDecisionRepo{
		instance:         buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID}),
		insertSignoffRes: infrastructure.SignoffInsertResult{ID: "sig-reject", WasReplay: false},
	}
	conn := &freezeDecisionConn{
		actorID: actorID,
		stageSignoffs: []signoffRow{{
			id:                 "sig-reject",
			approvalInstanceID: instanceID,
			stageInstanceID:    stageID,
			actorUserID:        actorID,
			actorTenantID:      "tenant-1",
			decision:           "reject",
			comment:            "needs changes",
			signedAt:           signedAt,
			signatureMethod:    "password",
			signaturePayload:   []byte(`{}`),
			contentHash:        validContentHash,
		}},
	}
	db := newFreezeDecisionTestDB(t, conn)
	svc := (&DecisionService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: signedAt},
	}).WithReleaseRecorder(&fakeReleaseRecorder{})

	result, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "reject",
		Comment:          "needs changes",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"hash": "abc"},
		ContentFormData:  map[string]any{"title": "Doc", "_content_hash": validContentHash},
	})
	if err != nil {
		t.Fatalf("RecordSignoff() error = %v", err)
	}
	if !result.InstanceRejected || conn.documentStatus != "draft" {
		t.Fatalf("expected rejected document reset to draft, status=%q result=%+v", conn.documentStatus, result)
	}
}
