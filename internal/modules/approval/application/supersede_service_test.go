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

	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// Fake repository for supersede tests.
//
// PublishSuperseding re-reads the prior document's revision_version under the
// row lock and CASes on THAT value (never on req.PriorRevisionVersion, which is
// pre-lock and possibly stale). The service therefore requires a wired
// repository — it fails closed without one — so every test wires this fake.
// The embedded interface supplies the rest of ApprovalRepository; any method
// this path unexpectedly starts calling panics with a nil-interface
// dereference, which is the fail-loud behaviour we want here.
// ---------------------------------------------------------------------------

type supersedeFakeRepo struct {
	infrastructure.ApprovalRepository

	revisionVersion int
	err             error

	calls         int
	gotDocumentID string
	gotTenantID   string
}

func (r *supersedeFakeRepo) GetDocumentRevisionVersion(_ context.Context, _ platformdb.Tx, documentID, tenantID string) (int, error) {
	r.calls++
	r.gotDocumentID = documentID
	r.gotTenantID = tenantID
	if r.err != nil {
		return 0, r.err
	}
	return r.revisionVersion, nil
}

// ---------------------------------------------------------------------------
// Fake driver for supersede tests.
//
// Two UPDATE statements are issued per call to PublishSuperseding, in this
// order (ADR 0085: the incumbent head must be demoted BEFORE the new document
// takes the published slot, or ux_documents_published_head is transiently
// violated inside the transaction):
//   1. UPDATE documents SET status='superseded' (prior doc OCC)
//   2. UPDATE documents SET status='published'  (new doc OCC)
//
// supersedeTestConn dispatches on the statement text (not on call order, so the
// tests genuinely observe the order rather than assuming it), returns the
// caller-configured rowsAffected for each, and records the position each UPDATE
// occupied so the happy path can assert prior-before-new.
// Non-UPDATE statements (BEGIN, COMMIT, governance event INSERT) always succeed.
// ---------------------------------------------------------------------------

type supersedeTestResult struct{ rowsAffected int64 }

func (r supersedeTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r supersedeTestResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type supersedeEmptyRows struct{}

func (supersedeEmptyRows) Columns() []string         { return nil }
func (supersedeEmptyRows) Close() error              { return nil }
func (supersedeEmptyRows) Next([]driver.Value) error { return io.EOF }

type supersedeSingleValueRows struct {
	value any
	done  bool
}

func (r *supersedeSingleValueRows) Columns() []string { return []string{"v"} }
func (r *supersedeSingleValueRows) Close() error      { return nil }
func (r *supersedeSingleValueRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

type supersedeTestStmt struct {
	conn  *supersedeTestConn
	query string
}

func (s *supersedeTestStmt) Close() error  { return nil }
func (s *supersedeTestStmt) NumInput() int { return -1 }

func (s *supersedeTestStmt) Exec(args []driver.Value) (driver.Result, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "set_config('metaldocs.asserted_caps'") {
		if len(args) > 0 {
			if raw, ok := args[0].(string); ok {
				s.conn.setAssertedCaps(raw)
			}
		}
		return supersedeTestResult{rowsAffected: 1}, nil
	}
	if strings.Contains(q, "update documents") && !s.conn.hasAssertedCap("document.edit") {
		return nil, fmt.Errorf("ErrCapabilityNotAsserted: document.edit required on documents")
	}
	if !strings.Contains(q, "update") {
		// Non-UPDATE (governance_events INSERT, etc.) always succeed with 1.
		return supersedeTestResult{rowsAffected: 1}, nil
	}
	// Track which UPDATE call this is, dispatching on the statement itself.
	// 'superseded' appears only in the prior-document UPDATE's SET clause.
	s.conn.updateCount++
	if strings.Contains(q, "'superseded'") {
		s.conn.priorUpdateOrder = s.conn.updateCount
		s.conn.priorUpdateArgs = append([]driver.Value(nil), args...)
		return supersedeTestResult{rowsAffected: s.conn.priorDocRowsAffected}, nil
	}
	s.conn.newUpdateOrder = s.conn.updateCount
	s.conn.newUpdateArgs = append([]driver.Value(nil), args...)
	return supersedeTestResult{rowsAffected: s.conn.newDocRowsAffected}, nil
}

func (s *supersedeTestStmt) Query(args []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)
	// ADR 0085 lock-order parity: lockDocumentRowsInIDOrder locks one document
	// row per statement, in sorted id order, and reads back the id it locked.
	if strings.Contains(q, "for update") && strings.Contains(q, "from documents") {
		var lockedID any
		if len(args) > 0 {
			lockedID = args[0]
		}
		return &supersedeSingleValueRows{value: lockedID}, nil
	}
	if strings.Contains(q, "from documents") {
		// Ported area resolver (M2/F2.1): two columns; non-NULL snapshot wins the
		// COALESCE, so areaCode-as-snapshot with a NULL CD link reproduces prior.
		return &docAreaRows{snapshot: s.conn.areaCode}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "iam_user_roles") {
		return &supersedeSingleValueRows{value: false}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "role_capabilities") {
		return &supersedeSingleValueRows{value: s.conn.authzGranted}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.asserted_caps'") {
		return &supersedeSingleValueRows{value: nil}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.tenant_id'") {
		return &supersedeSingleValueRows{value: s.conn.tenantID}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.actor_id'") {
		return &supersedeSingleValueRows{value: s.conn.actorID}, nil
	}
	return supersedeEmptyRows{}, nil
}

type supersedeTestConn struct {
	newDocRowsAffected   int64
	priorDocRowsAffected int64
	updateCount          int            // incremented on each UPDATE exec
	priorUpdateOrder     int            // 1-based position of the prior-doc UPDATE
	newUpdateOrder       int            // 1-based position of the new-doc UPDATE
	priorUpdateArgs      []driver.Value // bind args of the prior-doc UPDATE ($1 id, $2 tenant, $3 revision_version)
	newUpdateArgs        []driver.Value // bind args of the new-doc UPDATE ($1 id, $2 tenant, $3 revision_version)
	authzGranted         bool
	areaCode             string
	actorID              string
	tenantID             string
	assertedCaps         []string
}

func (c *supersedeTestConn) setAssertedCaps(raw string) {
	if strings.TrimSpace(raw) == "" {
		c.assertedCaps = nil
		return
	}
	var asserted []map[string]string
	if err := json.Unmarshal([]byte(raw), &asserted); err != nil {
		c.assertedCaps = nil
		return
	}
	caps := make([]string, 0, len(asserted))
	for _, item := range asserted {
		capability := strings.TrimSpace(item["cap"])
		if capability != "" {
			caps = append(caps, capability)
		}
	}
	c.assertedCaps = caps
}

func (c *supersedeTestConn) hasAssertedCap(capability string) bool {
	for _, asserted := range c.assertedCaps {
		if asserted == capability {
			return true
		}
	}
	return false
}

// supersedeOCCArg returns the revision_version bind ($3) captured from an
// UPDATE's args — i.e. the exact value the OCC/CAS predicate compared against.
func supersedeOCCArg(t *testing.T, args []driver.Value, label string) int64 {
	t.Helper()
	if len(args) < 3 {
		t.Fatalf("%s UPDATE: expected at least 3 bind args, got %d (%v)", label, len(args), args)
	}
	v, ok := args[2].(int64)
	if !ok {
		t.Fatalf("%s UPDATE: revision_version bind %#v is not int64", label, args[2])
	}
	return v
}

func (c *supersedeTestConn) Prepare(query string) (driver.Stmt, error) {
	return &supersedeTestStmt{conn: c, query: query}, nil
}

func (c *supersedeTestConn) Close() error              { return nil }
func (c *supersedeTestConn) Begin() (driver.Tx, error) { return c, nil }
func (c *supersedeTestConn) Commit() error             { return nil }
func (c *supersedeTestConn) Rollback() error           { return nil }

type supersedeTestDriver struct{ conn *supersedeTestConn }

func (d *supersedeTestDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

// newSupersedeTestDB registers a unique driver and returns the *sql.DB plus the
// backing conn (so tests can assert observed statement order). newRowsAffected
// controls the new-document UPDATE; priorRowsAffected the prior-document one —
// by statement, independent of which runs first.
func newSupersedeTestDB(t *testing.T, newRowsAffected, priorRowsAffected int64, authzGranted ...bool) (*sql.DB, *supersedeTestConn) {
	t.Helper()
	granted := true
	if len(authzGranted) > 0 {
		granted = authzGranted[0]
	}
	conn := &supersedeTestConn{
		newDocRowsAffected:   newRowsAffected,
		priorDocRowsAffected: priorRowsAffected,
		authzGranted:         granted,
		areaCode:             "QA",
		actorID:              "user-1",
		tenantID:             tenant.DevTenantID,
	}
	name := fmt.Sprintf("supersede_test_%p", conn)
	sql.Register(name, &supersedeTestDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open supersede test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, conn
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPublishSuperseding_HappyPath(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	repo := &supersedeFakeRepo{revisionVersion: 7}
	svc := &SupersedeService{repo: repo, emitter: emitter, clock: clock}
	// Both UPDATE statements match one row each.
	db, conn := newSupersedeTestDB(t, 1, 1, true)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-uuid-1",
		PriorDocumentID:      "doc-prior-uuid-1",
		SupersededBy:         "user-1",
		NewRevisionVersion:   4,
		PriorRevisionVersion: 7,
	}

	result, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req)
	if err != nil {
		t.Fatalf("PublishSuperseding: unexpected error: %v", err)
	}
	if result.NewDocumentStatus != "published" {
		t.Errorf("NewDocumentStatus = %q; want %q", result.NewDocumentStatus, "published")
	}
	if result.PriorDocumentStatus != "superseded" {
		t.Errorf("PriorDocumentStatus = %q; want %q", result.PriorDocumentStatus, "superseded")
	}

	if len(emitter.Events) != 1 {
		t.Fatalf("expected 1 governance event; got %d", len(emitter.Events))
	}
	ev := emitter.Events[0]
	if ev.EventType != "document_superseded" {
		t.Errorf("event type = %q; want %q", ev.EventType, "document_superseded")
	}
	if ev.ResourceID != "doc-new-uuid-1" {
		t.Errorf("event resource_id = %q; want %q", ev.ResourceID, "doc-new-uuid-1")
	}
	if ev.ActorUserID != "user-1" {
		t.Errorf("event actor_user_id = %q; want %q", ev.ActorUserID, "user-1")
	}
	if ev.TenantID != "tenant-uuid-1" {
		t.Errorf("event tenant_id = %q; want %q", ev.TenantID, "tenant-uuid-1")
	}

	// ADR 0085 / migration 0310 ux_documents_published_head: the incumbent head
	// must be demoted BEFORE the new document takes the published slot, or a
	// same-controlled-document supersession violates the partial unique index
	// mid-transaction.
	if conn.priorUpdateOrder != 1 || conn.newUpdateOrder != 2 {
		t.Errorf("update order = prior:%d new:%d; want prior:1 new:2 (prior demoted first)",
			conn.priorUpdateOrder, conn.newUpdateOrder)
	}

	// The prior document's OCC value must come from the under-lock re-read, not
	// from the request; the new document's still comes from the request (If-Match).
	if repo.calls != 1 {
		t.Errorf("GetDocumentRevisionVersion called %d times; want exactly 1 (under-lock re-read)", repo.calls)
	}
	if repo.gotDocumentID != "doc-prior-uuid-1" || repo.gotTenantID != "tenant-uuid-1" {
		t.Errorf("under-lock re-read scoped to (doc:%q tenant:%q); want (doc:%q tenant:%q)",
			repo.gotDocumentID, repo.gotTenantID, "doc-prior-uuid-1", "tenant-uuid-1")
	}
	if got := supersedeOCCArg(t, conn.priorUpdateArgs, "prior"); got != 7 {
		t.Errorf("prior-document OCC arg = %d; want 7 (re-read value)", got)
	}
	if got := supersedeOCCArg(t, conn.newUpdateArgs, "new"); got != 4 {
		t.Errorf("new-document OCC arg = %d; want 4 (req.NewRevisionVersion)", got)
	}
}

// TestPublishSuperseding_PriorOCCUsesUnderLockReRead is the regression pin for
// the OCC guarantee the row lock exists to provide: the prior document's CAS
// predicate must bind the revision_version read from the row AFTER the lock,
// never the caller-supplied (pre-lock, possibly stale) req.PriorRevisionVersion.
// The two values are deliberately different here, so a regression that reads the
// request field again makes the assertion fail rather than pass by coincidence.
func TestPublishSuperseding_PriorOCCUsesUnderLockReRead(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	const underLockVersion = 42
	repo := &supersedeFakeRepo{revisionVersion: underLockVersion}
	svc := &SupersedeService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSupersedeTestDB(t, 1, 1, true)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-occ-1",
		PriorDocumentID:      "doc-prior-occ-1",
		SupersededBy:         "user-1",
		NewRevisionVersion:   4,
		PriorRevisionVersion: 7, // stale: MUST NOT reach the UPDATE
	}

	if _, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("PublishSuperseding: unexpected error: %v", err)
	}

	if repo.calls != 1 {
		t.Fatalf("GetDocumentRevisionVersion called %d times; want exactly 1", repo.calls)
	}
	got := supersedeOCCArg(t, conn.priorUpdateArgs, "prior")
	if got == int64(req.PriorRevisionVersion) {
		t.Fatalf("prior-document OCC arg = %d — the stale request value was used; the under-lock re-read (%d) must win",
			got, underLockVersion)
	}
	if got != underLockVersion {
		t.Errorf("prior-document OCC arg = %d; want %d (under-lock re-read)", got, underLockVersion)
	}
}

// TestPublishSuperseding_ReReadFailurePropagates pins fail-closed behaviour: if
// the under-lock re-read errors there is no substitute OCC value, so the write
// must not happen and no governance event may be emitted (no-fallback principle).
func TestPublishSuperseding_ReReadFailurePropagates(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	repo := &supersedeFakeRepo{err: infrastructure.ErrStaleRevision}
	svc := &SupersedeService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newSupersedeTestDB(t, 1, 1, true)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-occ-2",
		PriorDocumentID:      "doc-prior-occ-2",
		SupersededBy:         "user-1",
		NewRevisionVersion:   4,
		PriorRevisionVersion: 7,
	}

	_, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected the under-lock re-read failure to propagate; got nil")
	}
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
		t.Errorf("expected errors.Is(err, ErrStaleRevision); got %v", err)
	}
	if conn.updateCount != 0 {
		t.Errorf("%d UPDATE(s) ran after the re-read failed; none may run", conn.updateCount)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted when the re-read fails; got %d", len(emitter.Events))
	}
}

// TestPublishSuperseding_NilRepoFailsClosed pins the fail-closed guard: with no
// repository the under-lock re-read is impossible, and falling back to the
// request's pre-lock value would silently weaken the OCC guarantee. Refusing is
// the only correct answer.
func TestPublishSuperseding_NilRepoFailsClosed(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SupersedeService{emitter: emitter, clock: clock} // repo deliberately nil
	db, conn := newSupersedeTestDB(t, 1, 1, true)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-nilrepo-1",
		PriorDocumentID:      "doc-prior-nilrepo-1",
		SupersededBy:         "user-1",
		NewRevisionVersion:   4,
		PriorRevisionVersion: 7,
	}

	_, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected PublishSuperseding to fail closed without a repository; got nil")
	}
	if !strings.Contains(err.Error(), "repository not wired") {
		t.Errorf("error = %v; want a 'repository not wired' fail-closed error", err)
	}
	if conn.updateCount != 0 {
		t.Errorf("%d UPDATE(s) ran without a wired repository; none may run", conn.updateCount)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted; got %d", len(emitter.Events))
	}
}

func TestPublishSuperseding_OCC_NewConflict(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SupersedeService{repo: &supersedeFakeRepo{revisionVersion: 5}, emitter: emitter, clock: clock}
	// The new-doc UPDATE (now second) returns 0 — OCC conflict.
	db, _ := newSupersedeTestDB(t, 0, 1, true)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-uuid-2",
		PriorDocumentID:      "doc-prior-uuid-2",
		SupersededBy:         "user-1",
		NewRevisionVersion:   3,
		PriorRevisionVersion: 5,
	}

	_, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected ErrStaleRevision; got nil")
	}
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
		t.Errorf("expected errors.Is(err, ErrStaleRevision); got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on OCC conflict; got %d", len(emitter.Events))
	}
}

func TestPublishSuperseding_OCC_PriorConflict(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

	svc := &SupersedeService{repo: &supersedeFakeRepo{revisionVersion: 9}, emitter: emitter, clock: clock}
	// The prior-doc UPDATE (now first) returns 0 — OCC conflict; the new-doc
	// UPDATE must never run.
	db, conn := newSupersedeTestDB(t, 1, 0, true)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-uuid-3",
		PriorDocumentID:      "doc-prior-uuid-3",
		SupersededBy:         "user-1",
		NewRevisionVersion:   2,
		PriorRevisionVersion: 9,
	}

	_, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected ErrStaleRevision; got nil")
	}
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
		t.Errorf("expected errors.Is(err, ErrStaleRevision); got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on OCC conflict; got %d", len(emitter.Events))
	}
	if conn.newUpdateOrder != 0 {
		t.Errorf("new-document UPDATE ran (position %d) after the prior-head demotion lost its CAS; it must not", conn.newUpdateOrder)
	}
}

// TestPublishSuperseding_CapabilityAssertPairsBeforeGatedWrite pins the
// tripwire pairing invariant (APP-04 / T-006) for the cutover/supersede path.
// supersedeTestStmt.Exec (line ~78) rejects any "UPDATE documents" unless
// metaldocs.asserted_caps already recorded "document.edit" — PublishSuperseding
// asserts CapDocumentSupersede then CapDocumentEdit via authz.Require before
// either OCC UPDATE runs, so this only passes because that call order holds
// in supersede_service.go. Mirrors
// TestCancelInstance_CapabilityAssertPairsBeforeGatedWrite in
// cancel_service_test.go.
func TestPublishSuperseding_CapabilityAssertPairsBeforeGatedWrite(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}
	svc := &SupersedeService{repo: &supersedeFakeRepo{revisionVersion: 1}, emitter: emitter, clock: clock}

	conn := &supersedeTestConn{
		newDocRowsAffected:   1,
		priorDocRowsAffected: 1,
		authzGranted:         true,
		areaCode:             "QA",
		actorID:              "user-1",
		tenantID:             tenant.DevTenantID,
	}
	name := fmt.Sprintf("supersede_pairing_test_%p", conn)
	sql.Register(name, &supersedeTestDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open supersede pairing test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-pairing-1",
		PriorDocumentID:      "doc-prior-pairing-1",
		SupersededBy:         "user-1",
		NewRevisionVersion:   1,
		PriorRevisionVersion: 1,
	}

	if _, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req); err != nil {
		t.Fatalf("PublishSuperseding: unexpected error (pairing broken?): %v", err)
	}
	if !conn.hasAssertedCap("document.edit") {
		t.Error("expected document.edit to be recorded in metaldocs.asserted_caps by end of tx")
	}
	if !conn.hasAssertedCap("document.supersede") {
		t.Error("expected document.supersede to be recorded in metaldocs.asserted_caps by end of tx")
	}
}

func TestPublishSuperseding_CapabilityDenied(t *testing.T) {
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}
	svc := &SupersedeService{repo: &supersedeFakeRepo{revisionVersion: 1}, emitter: emitter, clock: clock}
	db, _ := newSupersedeTestDB(t, 1, 1, false)

	req := SupersedeRequest{
		TenantID:             "tenant-uuid-1",
		NewDocumentID:        "doc-new-authz-1",
		PriorDocumentID:      "doc-prior-authz-1",
		SupersededBy:         "user-1",
		NewRevisionVersion:   1,
		PriorRevisionVersion: 1,
	}

	_, err := svc.PublishSuperseding(context.Background(), newTxRunner(db), req)
	if err == nil {
		t.Fatal("expected ErrCapabilityDenied; got nil")
	}
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("expected ErrCapabilityDenied; got %v", err)
	}
	if denied.Capability != "document.supersede" {
		t.Errorf("capability = %q; want %q", denied.Capability, "document.supersede")
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on denied capability; got %d", len(emitter.Events))
	}
}
