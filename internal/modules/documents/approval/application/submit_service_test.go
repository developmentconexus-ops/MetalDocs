package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// Ensure fakeSubmitRepo satisfies ApprovalRepository at compile time.
var _ infrastructure.ApprovalRepository = (*fakeSubmitRepo)(nil)

// ---------------------------------------------------------------------------
// Fake repo — only implements the methods called by SubmitRevisionForReview.
// ---------------------------------------------------------------------------

type fakeSubmitRepo struct {
	insertInstanceErr       error
	insertStageInstancesErr error
	lastInstance            domain.Instance
	// Embed a no-op for the full interface; all other methods panic if called.
	infrastructure.ApprovalRepository
}

func (r *fakeSubmitRepo) InsertInstance(_ context.Context, _ db.Tx, inst domain.Instance) error {
	r.lastInstance = inst
	return r.insertInstanceErr
}

func (r *fakeSubmitRepo) InsertStageInstances(_ context.Context, _ db.Tx, _ []domain.StageInstance) error {
	return r.insertStageInstancesErr
}

// LoadRoute and ResolveEligibleActors are called by SubmitRevisionForReview;
// they use the tx so the in-memory test driver serves their queries.

func (r *fakeSubmitRepo) LoadRoute(ctx context.Context, tx db.Tx, tenantID, routeID string) (domain.Route, error) {
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
		SELECT ars.stage_order, ars.name, ars.required_role, ars.required_capability,
		       ars.area_code, ars.quorum, ars.quorum_m, ars.on_eligibility_drift
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
	defer rows.Close()
	for rows.Next() {
		var stage domain.Stage
		var quorumM sql.NullInt32
		if err := rows.Scan(
			&stage.Order, &stage.Name, &stage.RequiredRole, &stage.RequiredCapability,
			&stage.AreaCode, &stage.Quorum, &quorumM, &stage.OnEligibilityDrift,
		); err != nil {
			return domain.Route{}, err
		}
		if quorumM.Valid {
			v := int(quorumM.Int32)
			stage.QuorumM = &v
		}
		route.Stages = append(route.Stages, stage)
	}
	if err := rows.Err(); err != nil {
		return domain.Route{}, err
	}
	return route, nil
}

// LoadGovernedRevisionNumber is called by SubmitRevisionForReview (T8b) to
// derive the governed documents.revision_number in-tx instead of trusting a
// client-supplied value. Delegates to the tx so the fake driver's
// governedRevisionNumberRows serves it.
func (r *fakeSubmitRepo) LoadGovernedRevisionNumber(ctx context.Context, tx db.Tx, tenantID, documentID string) (int, error) {
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

func (r *fakeSubmitRepo) ResolveEligibleActors(ctx context.Context, tx db.Tx, tenantID, areaCode, requiredRole string) ([]string, error) {
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

// The remaining 4 new interface methods are not called by SubmitRevisionForReview.
// The embedded infrastructure.ApprovalRepository is nil, so these must be explicitly
// implemented to avoid a nil-pointer panic if the compiler checks the interface.

func (r *fakeSubmitRepo) LoadPriorSignoffs(_ context.Context, _ db.Tx, _, _, _ string) ([]domain.Signoff, error) {
	panic("fakeSubmitRepo.LoadPriorSignoffs: not expected to be called in submit tests")
}

func (r *fakeSubmitRepo) LoadStageSignoffs(_ context.Context, _ db.Tx, _, _ string) ([]domain.Signoff, error) {
	panic("fakeSubmitRepo.LoadStageSignoffs: not expected to be called in submit tests")
}

func (r *fakeSubmitRepo) HasUnresolvedComments(_ context.Context, _ db.Tx, _, _ string) (bool, error) {
	panic("fakeSubmitRepo.HasUnresolvedComments: not expected to be called in submit tests")
}

// HasUnresolvedInstanceComments (F5) IS called by SubmitRevisionForReview for
// approval-only routes (executeFreeze at submit time) — unlike the
// document-wide HasUnresolvedComments above, this one is a real, expected
// call site. Always reports clean so existing submit tests exercise the
// freeze happy path without needing to set up comment fixtures.
func (r *fakeSubmitRepo) HasUnresolvedInstanceComments(_ context.Context, _ db.Tx, _, _ string, _ time.Time) (bool, error) {
	return false, nil
}

// PinFrozenHash (F5) is the freeze CAS write; always reports a won pin for
// these submit-path unit tests (no concurrent-freeze scenario in this file —
// see freeze_test.go for that unit coverage).
func (r *fakeSubmitRepo) PinFrozenHash(_ context.Context, _ db.Tx, _, _, _ string) (bool, error) {
	return true, nil
}

func (r *fakeSubmitRepo) LoadFrozenContentHash(_ context.Context, _ db.Tx, _, _ string) (string, error) {
	panic("fakeSubmitRepo.LoadFrozenContentHash: not expected to be called in submit tests")
}

func (r *fakeSubmitRepo) InsertDelegation(_ context.Context, _ db.Tx, _ domain.Delegation) error {
	panic("fakeSubmitRepo.InsertDelegation: not expected to be called in submit tests")
}

func (r *fakeSubmitRepo) DeleteDelegation(_ context.Context, _ db.Tx, _, _, _ string, _ bool) (bool, error) {
	panic("fakeSubmitRepo.DeleteDelegation: not expected to be called in submit tests")
}

func (r *fakeSubmitRepo) LoadActiveDelegationsFor(_ context.Context, _ db.Tx, _, _ string, _ time.Time) ([]domain.Delegation, error) {
	panic("fakeSubmitRepo.LoadActiveDelegationsFor: not expected to be called in submit tests")
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
	authzGranted           bool
	areaCode               string
	actorID                string
	tenantID               string
	revisionTitle          string
	reasonForChange        sql.NullString
	reasonCategory         sql.NullString
	governedRevisionNumber int // documents.revision_number the fake DB row reports (T8b)
}

// governedRevisionNumberRows is the fake-driver result for
// ApprovalRepository.LoadGovernedRevisionNumber's SELECT revision_number FROM
// documents ... query (T8b) — one column, one row.
type governedRevisionNumberRows struct {
	value int64
	done  bool
}

func (r *governedRevisionNumberRows) Columns() []string { return []string{"revision_number"} }
func (r *governedRevisionNumberRows) Close() error      { return nil }
func (r *governedRevisionNumberRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
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
	if strings.Contains(q, "select revision_number") && strings.Contains(q, "from documents") {
		// Governed revision_number read (T8b): SubmitRevisionForReview derives
		// this in-tx via ApprovalRepository.LoadGovernedRevisionNumber instead of
		// trusting a client-supplied value.
		return &governedRevisionNumberRows{value: int64(s.conn.governedRevisionNumber)}, nil
	}
	if strings.Contains(q, "from documents") {
		// Ported area resolver (M2/F2.1) selects two columns
		// (process_area_code_snapshot, controlled_document_id). A non-NULL
		// snapshot wins the COALESCE, so returning areaCode as the snapshot with
		// a NULL CD link reproduces the prior single-column "QA" result.
		return &docAreaRows{snapshot: s.conn.areaCode}, nil
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
	if strings.Contains(q, "user_process_areas") {
		// W6 (F2): submit-time eligible-pool check requires a non-empty
		// resolution for these pre-existing fixtures, which never seeded
		// eligibility rows because it was unenforced before this feature.
		return &submitSingleValueRows{value: "user-eligible-1"}, nil
	}
	return submitEmptyRows{}, nil
}

func (s *submitTestStmt) ExecContext(_ context.Context, args []driver.NamedValue) (driver.Result, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "update documents") && len(args) >= 4 {
		if title, ok := args[3].Value.(string); ok {
			s.conn.revisionTitle = title
		}
		// args[4]/args[5] = reason_for_change/reason_category ($5/$6). A nil
		// driver.Value (bound via nullableString) means the field was unset and
		// persists as SQL NULL; a non-nil value is the trimmed string.
		if len(args) >= 6 {
			if v := args[4].Value; v != nil {
				if reason, ok := v.(string); ok {
					s.conn.reasonForChange = sql.NullString{String: reason, Valid: true}
				}
			} else {
				s.conn.reasonForChange = sql.NullString{}
			}
			if v := args[5].Value; v != nil {
				if category, ok := v.(string); ok {
					s.conn.reasonCategory = sql.NullString{String: category, Valid: true}
				}
			} else {
				s.conn.reasonCategory = sql.NullString{}
			}
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
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		ContentFormData: map[string]any{"title": "My Doc", "revision": 1},
		RevisionVersion: 1,
	}

	result, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	if result.InstanceID == "" {
		t.Error("SubmitResult.InstanceID must not be empty")
	}
	// The client idempotency key must be threaded verbatim onto the instance so
	// the UNIQUE(document_id, idempotency_key) constraint is a real backstop
	// (F-D4) — never a server-clock derivation.
	if repo.lastInstance.IdempotencyKey != req.IdempotencyKey {
		t.Errorf("instance idempotency key = %q; want %q (client key threaded)", repo.lastInstance.IdempotencyKey, req.IdempotencyKey)
	}
	if len(emitter.Events) != 1 {
		t.Errorf("expected 1 governance event; got %d", len(emitter.Events))
	}
	if emitter.Events[0].EventType != "approval_submitted" {
		t.Errorf("event type = %q; want %q", emitter.Events[0].EventType, "approval_submitted")
	}
}

func TestSubmitRevisionForReview_RequiresIdempotencyKey(t *testing.T) {
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
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
		// IdempotencyKey deliberately empty.
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrIdempotencyKeyRequired) {
		t.Fatalf("error = %v, want ErrIdempotencyKeyRequired", err)
	}
	// Fail-closed before any instance is inserted or event emitted.
	if repo.lastInstance.ID != "" {
		t.Error("no instance should be inserted when the idempotency key is missing")
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted; got %d", len(emitter.Events))
	}
}

func TestSubmitRevisionForReview_DefaultsRevisionTitleForFirstGovernedRevision(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 0

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
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
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "   ",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
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
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1 // deliberately != RevisionVersion (7) below
	formData := map[string]any{"title": "My Doc"}

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Atualizacao",
		ReasonForChange: "Content hash binding regression coverage",
		ContentFormData: formData,
		RevisionVersion: 7,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	want, err := ComputeContentHash(ContentHashInput{
		TenantID:       req.TenantID,
		DocumentID:     req.DocumentID,
		RevisionNumber: conn.governedRevisionNumber,
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
		t.Fatalf("content hash must not bind to technical RevisionVersion when it differs from governed revision_number")
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
	// End anchor: next top-level func after SubmitRevisionForReview.
	// loadRoute was relocated to the repository layer (H-5.1); use the next
	// "\nfunc " boundary so the tripwire is robust regardless of what follows.
	rest := body[start+len("func (s *SubmitService) SubmitRevisionForReview"):]
	end := strings.Index(rest, "\nfunc ")
	var submit string
	if end == -1 {
		// SubmitRevisionForReview is the last func in the file — use the whole tail.
		submit = body[start:]
	} else {
		submit = body[start : start+len("func (s *SubmitService) SubmitRevisionForReview")+end]
	}

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
		insertInstanceErr: infrastructure.ErrDuplicateSubmission,
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
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		ContentFormData: map[string]any{"title": "My Doc", "revision": 1},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected ErrDuplicateSubmission; got nil")
	}
	if !errors.Is(err, infrastructure.ErrDuplicateSubmission) {
		t.Errorf("expected errors.Is(err, infrastructure.ErrDuplicateSubmission); got %v", err)
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
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
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

// ---------------------------------------------------------------------------
// F6.3 — structured reason-for-change (validation-contract.md §5)
// ---------------------------------------------------------------------------

// TestSubmitPersistsReason proves a REV>=1 submit with reason_for_change (+
// reason_category) persists BOTH on the SAME draft->under_review UPDATE to
// public.documents that already sets revision_title — not a second write, and
// NOT stored into revision_title (contract §5.1/§5.2).
func TestSubmitPersistsReason(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Atualizacao de conteudo",
		ReasonForChange: "Corrected a dimensional tolerance per customer complaint #482",
		ReasonCategory:  "corrective",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 3,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	if !conn.reasonForChange.Valid || conn.reasonForChange.String != req.ReasonForChange {
		t.Fatalf("persisted reason_for_change = %+v, want %q", conn.reasonForChange, req.ReasonForChange)
	}
	if !conn.reasonCategory.Valid || conn.reasonCategory.String != req.ReasonCategory {
		t.Fatalf("persisted reason_category = %+v, want %q", conn.reasonCategory, req.ReasonCategory)
	}
	// The revision_title UPDATE target must be untouched by the reason value —
	// proves no code path aliases reason_for_change into revision_title.
	if conn.revisionTitle != req.RevisionTitle {
		t.Fatalf("revision_title = %q, want %q (must not be overwritten by reason_for_change)", conn.revisionTitle, req.RevisionTitle)
	}
}

// TestSubmitPersistsReason_OptionalCategoryOmitted proves reason_category is
// genuinely optional: submitting with only reason_for_change persists a NULL
// reason_category (not an empty-string sentinel).
func TestSubmitPersistsReason_OptionalCategoryOmitted(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Atualizacao",
		ReasonForChange: "Administrative renumbering only",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 2,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	if !conn.reasonForChange.Valid || conn.reasonForChange.String != req.ReasonForChange {
		t.Fatalf("persisted reason_for_change = %+v, want %q", conn.reasonForChange, req.ReasonForChange)
	}
	if conn.reasonCategory.Valid {
		t.Fatalf("persisted reason_category = %+v, want SQL NULL (unset, not empty string)", conn.reasonCategory)
	}
}

// TestSubmitReasonOnAuditTrail proves the submit emits exactly ONE
// approval_submitted governance event whose payload carries reason_for_change
// (+ reason_category), inside the same business tx (contract §5.2).
func TestSubmitReasonOnAuditTrail(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Revisao regulatoria",
		ReasonForChange: "Updated per new ISO 9001:2026 clause 8.5.6",
		ReasonCategory:  "regulatory",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	if len(emitter.Events) != 1 {
		t.Fatalf("emitted %d governance events, want exactly 1", len(emitter.Events))
	}
	ev := emitter.Events[0]
	if ev.EventType != "approval_submitted" {
		t.Fatalf("event type = %q, want %q", ev.EventType, "approval_submitted")
	}
	var payload map[string]any
	if err := json.Unmarshal(ev.PayloadJSON, &payload); err != nil {
		t.Fatalf("unmarshal event payload: %v", err)
	}
	if payload["reason_for_change"] != req.ReasonForChange {
		t.Fatalf("payload.reason_for_change = %v, want %q", payload["reason_for_change"], req.ReasonForChange)
	}
	if payload["reason_category"] != req.ReasonCategory {
		t.Fatalf("payload.reason_category = %v, want %q", payload["reason_category"], req.ReasonCategory)
	}
}

// TestSubmitReasonRequiredRev1 proves a REV>=1 submit without reason_for_change
// is rejected with ErrReasonForChangeRequired (mapped to 422 problem+json at the
// http layer; see errors_test.go), and no governance event is emitted.
func TestSubmitReasonRequiredRev1(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Revisao sem motivo",
		ReasonForChange: "   ", // whitespace-only counts as empty
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrReasonForChangeRequired) {
		t.Fatalf("error = %v, want ErrReasonForChangeRequired", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted when reason_for_change is missing; got %d", len(emitter.Events))
	}
}

// TestSubmitReasonOptionalAtRev0 proves REV 0 (initial creation) does NOT
// require reason_for_change, mirroring the RevisionTitle REV-0 default
// convention (contract §5.3).
func TestSubmitReasonOptionalAtRev0(t *testing.T) {
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
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error at REV 0: %v", err)
	}
	if conn.reasonForChange.Valid {
		t.Fatalf("persisted reason_for_change = %+v, want SQL NULL at REV 0 with no reason supplied", conn.reasonForChange)
	}
}

// TestSubmitReasonRequiredRev1_DerivedFromDocumentRow_NoClientRevisionNumber
// proves the REV>=1 reason_for_change gate fires from the document row's OWN
// governed revision_number — read in-tx via
// ApprovalRepository.LoadGovernedRevisionNumber — even though the request
// never supplies a client-controlled revision number (T8b root-cause fix:
// SubmitRequest no longer has a RevisionNumber field at all; the prior bug let
// every live HTTP submit default to 0 and silently skip this gate for
// revision-of-published-document submits with a real revision_number >= 1).
func TestSubmitReasonRequiredRev1_DerivedFromDocumentRow_NoClientRevisionNumber(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1 // the document row is a revision of a published doc

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Revisao sem motivo",
		ReasonForChange: "", // omitted, as a real client payload would never carry revision_number
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
		// No RevisionNumber field: the value must come from the documents row.
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrReasonForChangeRequired) {
		t.Fatalf("error = %v, want ErrReasonForChangeRequired (gate must fire from the derived DB revision_number, not a client-supplied value)", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted when reason_for_change is missing; got %d", len(emitter.Events))
	}
}

// TestSubmitReasonOptionalAtRev0_DerivedFromDocumentRow proves REV 0
// (documents.revision_number = 0, initial creation) still allows an empty
// reason_for_change — no regression from deriving the value in-tx instead of
// trusting a client-supplied field.
func TestSubmitReasonOptionalAtRev0_DerivedFromDocumentRow(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 0

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error at derived REV 0: %v", err)
	}
	if conn.reasonForChange.Valid {
		t.Fatalf("persisted reason_for_change = %+v, want SQL NULL at REV 0 with no reason supplied", conn.reasonForChange)
	}
}

// TestSubmitRevisionTitleRequiredRev1_DerivedFromDocumentRow proves the
// sibling REV>=1 revision_title gate also fires from the derived governed
// revision_number (the latent gap T8b also closes).
func TestSubmitRevisionTitleRequiredRev1_DerivedFromDocumentRow(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 1

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "   ", // whitespace-only counts as empty
		ReasonForChange: "Some reason so only the title gate is exercised",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrRevisionTitleRequired) {
		t.Fatalf("error = %v, want ErrRevisionTitleRequired", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted when revision title is missing; got %d", len(emitter.Events))
	}
}

// TestSubmitContentHash_UsesDerivedGovernedRevisionNumber proves the derived
// revision_number reaches ContentHashInput.RevisionNumber (the full-correct
// fix: one governed value feeds both the normalizers AND the content hash,
// rather than the content hash silently continuing to hash with 0).
func TestSubmitContentHash_UsesDerivedGovernedRevisionNumber(t *testing.T) {
	repo := &fakeSubmitRepo{}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SubmitService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSubmitTestDBWithConn(t, true)
	conn.governedRevisionNumber = 5
	formData := map[string]any{"title": "My Doc"}

	req := SubmitRequest{
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		RouteID:         "route-uuid-1",
		SubmittedBy:     "user-1",
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Atualizacao",
		ReasonForChange: "Content hash must bind to the DB-derived revision_number",
		ContentFormData: formData,
		RevisionVersion: 7,
	}

	if _, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	want, err := ComputeContentHash(ContentHashInput{
		TenantID:       req.TenantID,
		DocumentID:     req.DocumentID,
		RevisionNumber: 5, // conn.governedRevisionNumber, NOT req.RevisionVersion (7) NOR 0
		FormData:       formData,
	})
	if err != nil {
		t.Fatalf("ComputeContentHash: %v", err)
	}
	if repo.lastInstance.ContentHashAtSubmit != want {
		t.Fatalf("content hash must bind to the derived governed revision_number (5): got %s want %s", repo.lastInstance.ContentHashAtSubmit, want)
	}
}

// TestSubmitReasonCategoryInvalidRejected proves an out-of-enum reason_category
// is rejected at the friendly first-line (app) layer, mirroring the DB CHECK
// ck_documents_reason_category (contract §5.1; DB-level proof lives in
// document_review_columns_integration_test.go, T1).
func TestSubmitReasonCategoryInvalidRejected(t *testing.T) {
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
		IdempotencyKey:  "33333333-3333-3333-3333-333333333333",
		RevisionTitle:   "Revisao",
		ReasonForChange: "Some reason",
		ReasonCategory:  "not_a_real_category",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 1,
	}

	_, err := svc.SubmitRevisionForReview(context.Background(), newTxRunner(db), req)
	if !errors.Is(err, ErrReasonCategoryInvalid) {
		t.Fatalf("error = %v, want ErrReasonCategoryInvalid", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted when reason_category is invalid; got %d", len(emitter.Events))
	}
}
