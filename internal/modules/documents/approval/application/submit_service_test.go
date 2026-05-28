package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// Fake repo — only implements the methods called by SubmitRevisionForReview.
// ---------------------------------------------------------------------------

type fakeSubmitRepo struct {
	insertInstanceErr       error
	insertStageInstancesErr error
	lastInstance            domain.Instance
	// Embed a no-op for the full interface; all other methods panic if called.
	repository.ApprovalRepository
}

func (r *fakeSubmitRepo) InsertInstance(_ context.Context, _ *sql.Tx, inst domain.Instance) error {
	r.lastInstance = inst
	return r.insertInstanceErr
}

func (r *fakeSubmitRepo) InsertStageInstances(_ context.Context, _ *sql.Tx, _ []domain.StageInstance) error {
	return r.insertStageInstancesErr
}

// ---------------------------------------------------------------------------
// Fixed clock for deterministic idempotency keys.
// ---------------------------------------------------------------------------

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// ---------------------------------------------------------------------------
// Minimal in-memory SQL driver that serves route + stage rows.
// ---------------------------------------------------------------------------
// The driver must handle:
//   1. BEGIN             — from db.BeginTx
//   2. approval_routes   SELECT → returns one route row
//   3. approval_route_stages SELECT → returns one stage row
//   4. COMMIT/ROLLBACK   — tx lifecycle

type submitTestConn struct {
	name          string // driver instance name, unused but kept for debugging
	authzGranted  bool
	areaCode      string
	actorID       string
	tenantID      string
	revisionTitle string
}

type submitNoopResult struct{}

func (submitNoopResult) LastInsertId() (int64, error) { return 0, nil }
func (submitNoopResult) RowsAffected() (int64, error) { return 1, nil }

type submitEmptyRows struct{}

func (submitEmptyRows) Columns() []string         { return nil }
func (submitEmptyRows) Close() error              { return nil }
func (submitEmptyRows) Next([]driver.Value) error { return io.EOF }

type submitSingleValueRows struct {
	value any
	done  bool
}

func (r *submitSingleValueRows) Columns() []string { return []string{"v"} }
func (r *submitSingleValueRows) Close() error      { return nil }
func (r *submitSingleValueRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

// routeRow returns a single-column-set row representing approval_routes.
// Columns: id, tenant_id, profile_code, version
type routeRows struct {
	done bool
}

func (r *routeRows) Columns() []string { return []string{"id", "tenant_id", "profile_code", "version"} }
func (r *routeRows) Close() error      { return nil }
func (r *routeRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = "route-uuid-1"
	dest[1] = "tenant-uuid-1"
	dest[2] = "ISO_9001"
	dest[3] = int64(1)
	return nil
}

// stageRows returns one stage row representing approval_route_stages.
// Columns: stage_order, name, required_role, required_capability,
//
//	area_code, quorum, quorum_m, on_eligibility_drift
type stageRows struct {
	done bool
}

func (r *stageRows) Columns() []string {
	return []string{
		"stage_order", "name", "required_role", "required_capability",
		"area_code", "quorum", "quorum_m", "on_eligibility_drift",
	}
}
func (r *stageRows) Close() error { return nil }
func (r *stageRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = int64(1)
	dest[1] = "QA Review"
	dest[2] = "quality_approver"
	dest[3] = "doc.approve"
	dest[4] = "QA"
	dest[5] = "any_1_of"
	dest[6] = nil // quorum_m NULL
	dest[7] = "keep_snapshot"
	return nil
}

type submitTestStmt struct {
	conn  *submitTestConn
	query string
}

func (s *submitTestStmt) Close() error  { return nil }
func (s *submitTestStmt) NumInput() int { return -1 }

func (s *submitTestStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return submitNoopResult{}, nil
}

func (s *submitTestStmt) Query(_ []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "from documents") {
		return &submitSingleValueRows{value: s.conn.areaCode}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "iam_user_roles") {
		return &submitSingleValueRows{value: false}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "role_capabilities") {
		return &submitSingleValueRows{value: s.conn.authzGranted}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.asserted_caps'") {
		return &submitSingleValueRows{value: nil}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.tenant_id'") {
		return &submitSingleValueRows{value: s.conn.tenantID}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.actor_id'") {
		return &submitSingleValueRows{value: s.conn.actorID}, nil
	}
	if strings.Contains(q, "approval_routes") && strings.Contains(q, "where") && !strings.Contains(q, "approval_route_stages") {
		return &routeRows{}, nil
	}
	if strings.Contains(q, "approval_route_stages") {
		return &stageRows{}, nil
	}
	return submitEmptyRows{}, nil
}

func (s *submitTestStmt) ExecContext(_ context.Context, args []driver.NamedValue) (driver.Result, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "update documents") && len(args) >= 4 {
		if title, ok := args[3].Value.(string); ok {
			s.conn.revisionTitle = title
		}
	}
	return submitNoopResult{}, nil
}

var _ driver.StmtExecContext = (*submitTestStmt)(nil)

func (c *submitTestConn) Prepare(query string) (driver.Stmt, error) {
	return &submitTestStmt{conn: c, query: query}, nil
}
func (c *submitTestConn) Close() error              { return nil }
func (c *submitTestConn) Begin() (driver.Tx, error) { return c, nil }
func (c *submitTestConn) Commit() error             { return nil }
func (c *submitTestConn) Rollback() error           { return nil }

type submitTestDriver struct{ conn *submitTestConn }

func (d *submitTestDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

// newSubmitTestDB registers a unique driver per test and returns a *sql.DB.
func newSubmitTestDBWithConn(t *testing.T, authzGranted ...bool) (*sql.DB, *submitTestConn) {
	t.Helper()
	granted := true
	if len(authzGranted) > 0 {
		granted = authzGranted[0]
	}
	conn := &submitTestConn{
		authzGranted: granted,
		areaCode:     "QA",
		actorID:      "user-1",
		tenantID:     tenant.DevTenantID,
	}
	name := fmt.Sprintf("submit_test_%p", conn)
	sql.Register(name, &submitTestDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open submit test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, conn
}

func newSubmitTestDB(t *testing.T, authzGranted ...bool) *sql.DB {
	t.Helper()
	db, _ := newSubmitTestDBWithConn(t, authzGranted...)
	return db
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestSubmitRevisionForReview_HappyPath(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db := newSubmitTestDB(t, true)

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		ContentFormData: map[string]any{"title": "My Doc", "revision": 1},
		RevisionVersion: 1,
	}

	result, err := svc.SubmitRevisionForReview(context.Background(), db, req)
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	if result.InstanceID == "" {
		t.Error("SubmitResult.InstanceID must not be empty")
	}
	if len(emitter.Events) != 1 {
		t.Errorf("expected 1 governance event; got %d", len(emitter.Events))
	}
	if emitter.Events[0].EventType != "approval_submitted" {
		t.Errorf("event type = %q; want %q", emitter.Events[0].EventType, "approval_submitted")
	}
}

func TestSubmitRevisionForReview_DefaultsRevisionTitleForFirstGovernedRevision(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		RevisionTitle:   "",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
		RevisionNumber:  0,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), db, req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	if conn.revisionTitle != defaultInitialRevisionTitle {
		t.Fatalf("revision title = %q, want %q", conn.revisionTitle, defaultInitialRevisionTitle)
	}
}

func TestSubmitRevisionForReview_RequiresRevisionTitleAfterFirstGovernedRevision(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, _ := newSubmitTestDBWithConn(t, true)

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		RevisionTitle:   "   ",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
		RevisionNumber:  1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), db, req)
	if !errors.Is(err, ErrRevisionTitleRequired) {
		t.Fatalf("error = %v, want ErrRevisionTitleRequired", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted when revision title is missing; got %d", len(emitter.Events))
	}
}

func TestSubmitRevisionForReview_ContentHashUsesGovernedRevisionNumber(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, _ := newSubmitTestDBWithConn(t, true)
	formData := map[string]any{"title": "My Doc"}

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		RevisionTitle:   "Atualizacao",
		ContentFormData: formData,
		RevisionVersion: 7,
		RevisionNumber:  1,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), db, req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	want, err := ComputeContentHash(ContentHashInput{
		TenantID:       req.TenantID,
		DocumentID:     req.DocumentID,
		RevisionNumber: req.RevisionNumber,
		FormData:       formData,
	})
	if err != nil {
		t.Fatalf("ComputeContentHash: %v", err)
	}
	if repo.lastInstance.ContentHashAtSubmit != want {
		t.Fatalf("content hash revision binding mismatch: got %s want %s", repo.lastInstance.ContentHashAtSubmit, want)
	}
	staleVersionHash, err := ComputeContentHash(ContentHashInput{
		TenantID:       req.TenantID,
		DocumentID:     req.DocumentID,
		RevisionNumber: req.RevisionVersion,
		FormData:       formData,
	})
	if err != nil {
		t.Fatalf("ComputeContentHash stale version: %v", err)
	}
	if repo.lastInstance.ContentHashAtSubmit == staleVersionHash {
		t.Fatalf("content hash must not bind to technical RevisionVersion when it differs from governed RevisionNumber")
	}
}

func TestSubmitRevisionForReview_AssertsDocumentEditBeforeDocumentsUpdate(t *testing.T) {
	src, err := os.ReadFile("submit_service.go")
	if err != nil {
		t.Fatalf("read submit_service.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (s *SubmitService) SubmitRevisionForReview")
	if start == -1 {
		t.Fatal("SubmitRevisionForReview not found")
	}
	end := strings.Index(body[start:], "func (s *SubmitService) loadRoute")
	if end == -1 {
		t.Fatal("loadRoute not found")
	}
	submit := body[start : start+end]

	submitRequire := strings.Index(submit, "CapDocumentSubmit")
	if submitRequire == -1 {
		t.Fatal("SubmitRevisionForReview must assert document.submit")
	}
	editRequire := strings.Index(submit, "CapDocumentEdit")
	if editRequire == -1 {
		t.Fatal("SubmitRevisionForReview must assert document.edit before updating documents")
	}
	updateDocuments := strings.Index(submit, "UPDATE documents")
	if updateDocuments == -1 {
		t.Fatal("SubmitRevisionForReview documents update not found")
	}
	if editRequire < submitRequire {
		t.Fatal("SubmitRevisionForReview must assert document.submit before document.edit")
	}
	if editRequire > updateDocuments {
		t.Fatal("SubmitRevisionForReview asserts document.edit after documents update; tripwire requires it before UPDATE")
	}
}

func TestSubmitRevisionForReview_DuplicateSubmission(t *testing.T) {
	repo := &fakeSubmitRepo{
		insertInstanceErr: repository.ErrDuplicateSubmission,
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db := newSubmitTestDB(t, true)

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		ContentFormData: map[string]any{"title": "My Doc", "revision": 1},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), db, req)
	if err == nil {
		t.Fatal("expected ErrDuplicateSubmission; got nil")
	}
	if !errors.Is(err, repository.ErrDuplicateSubmission) {
		t.Errorf("expected errors.Is(err, repository.ErrDuplicateSubmission); got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on duplicate; got %d", len(emitter.Events))
	}
}

func TestSubmitRevisionForReview_CapabilityDenied(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db := newSubmitTestDB(t, false)

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), db, req)
	if err == nil {
		t.Fatal("expected ErrCapabilityDenied; got nil")
	}
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("expected ErrCapabilityDenied; got %v", err)
	}
	if denied.Capability != "document.submit" {
		t.Errorf("capability = %q; want %q", denied.Capability, "document.submit")
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on denied capability; got %d", len(emitter.Events))
	}
}
