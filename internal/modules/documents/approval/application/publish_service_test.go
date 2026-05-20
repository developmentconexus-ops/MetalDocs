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

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// Fake repo for publish tests — only LoadInstance is called.
// ---------------------------------------------------------------------------

type fakePublishRepo struct {
	instance                      *domain.Instance
	loadErr                       error
	currentPublishedHeadForDoc    string
	loadCurrentPublishedHeadErr   error
	loadCurrentPublishedHeadCalls int
	loadCurrentPublishedHeadDocID string
	validateSupersedeCalls        int
	validateSupersedeTenantID     string
	validateSupersedeDocumentID   string
	validateSupersedeTargetID     string
	validateSupersedeErr          error
	repository.ApprovalRepository // no-op embed for unused methods
}

func (r *fakePublishRepo) LoadInstance(_ context.Context, _ *sql.Tx, _, _ string) (*domain.Instance, error) {
	return r.instance, r.loadErr
}

func (r *fakePublishRepo) LoadCurrentPublishedHeadForDocument(_ context.Context, _ *sql.Tx, tenantID, documentID string) (string, error) {
	r.loadCurrentPublishedHeadCalls++
	r.validateSupersedeTenantID = tenantID
	r.loadCurrentPublishedHeadDocID = documentID
	return r.currentPublishedHeadForDoc, r.loadCurrentPublishedHeadErr
}

func (r *fakePublishRepo) ValidateScheduledSupersedeTarget(_ context.Context, _ *sql.Tx, tenantID, documentID, supersededDocumentID string) error {
	r.validateSupersedeCalls++
	r.validateSupersedeTenantID = tenantID
	r.validateSupersedeDocumentID = documentID
	r.validateSupersedeTargetID = supersededDocumentID
	return r.validateSupersedeErr
}

// ---------------------------------------------------------------------------
// Fake driver for publish tests.
// The driver handles:
//   1. BEGIN             — from db.BeginTx
//   2. UPDATE documents  — OCC transition (exec, returns rowsAffected)
//   3. COMMIT/ROLLBACK
// ---------------------------------------------------------------------------

type publishTestResult struct {
	rowsAffected int64
}

func (r publishTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r publishTestResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type publishEmptyRows struct{}

func (publishEmptyRows) Columns() []string         { return nil }
func (publishEmptyRows) Close() error              { return nil }
func (publishEmptyRows) Next([]driver.Value) error { return io.EOF }

type publishSingleValueRows struct {
	value any
	done  bool
}

func (r *publishSingleValueRows) Columns() []string { return []string{"v"} }
func (r *publishSingleValueRows) Close() error      { return nil }
func (r *publishSingleValueRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

type publishNoRows struct{}

func (publishNoRows) Columns() []string         { return []string{"v"} }
func (publishNoRows) Close() error              { return nil }
func (publishNoRows) Next([]driver.Value) error { return io.EOF }

type publishTestStmt struct {
	query        string
	rowsAffected int64 // injected by the conn
	conn         *publishTestConn
}

func (s *publishTestStmt) Close() error  { return nil }
func (s *publishTestStmt) NumInput() int { return -1 }

func (s *publishTestStmt) Exec(args []driver.Value) (driver.Result, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "set_config('metaldocs.asserted_caps'") {
		if len(args) > 0 {
			raw, _ := args[0].(string)
			s.conn.setAssertedCaps(raw)
		}
		return publishTestResult{rowsAffected: 1}, nil
	}
	if strings.Contains(q, "set_config('metaldocs.tenant_id'") {
		if len(args) > 0 {
			if tenantID, ok := args[0].(string); ok {
				s.conn.tenantID = tenantID
			}
		}
		return publishTestResult{rowsAffected: 1}, nil
	}
	if strings.Contains(q, "set_config('metaldocs.actor_id'") {
		if len(args) > 0 {
			if actorID, ok := args[0].(string); ok {
				s.conn.actorID = actorID
			}
		}
		return publishTestResult{rowsAffected: 1}, nil
	}
	if strings.Contains(q, "update documents") {
		s.conn.lastUpdateArgs = append([]driver.Value(nil), args...)
		if !s.conn.hasAssertedCap("document.edit") {
			return nil, fmt.Errorf("ErrCapabilityNotAsserted: document.edit required on documents")
		}
	}
	return publishTestResult{rowsAffected: s.rowsAffected}, nil
}

func (s *publishTestStmt) Query(args []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "update documents") {
		s.conn.lastUpdateArgs = append([]driver.Value(nil), args...)
		if !s.conn.hasAssertedCap("document.edit") {
			return nil, fmt.Errorf("ErrCapabilityNotAsserted: document.edit required on documents")
		}
		if s.conn.rowsAffected == 0 {
			return publishNoRows{}, nil
		}
		return &publishSingleValueRows{value: s.conn.scheduleGeneration}, nil
	}
	if strings.Contains(q, "from documents") {
		return &publishSingleValueRows{value: s.conn.areaCode}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "iam_user_roles") {
		return &publishSingleValueRows{value: false}, nil
	}
	if strings.Contains(q, "select exists") && strings.Contains(q, "role_capabilities") {
		return &publishSingleValueRows{value: s.conn.authzGranted}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.asserted_caps'") {
		return &publishSingleValueRows{value: s.conn.assertedCapsRawValue()}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.tenant_id'") {
		return &publishSingleValueRows{value: s.conn.tenantID}, nil
	}
	if strings.Contains(q, "current_setting('metaldocs.actor_id'") {
		return &publishSingleValueRows{value: s.conn.actorID}, nil
	}
	return publishEmptyRows{}, nil
}

type publishTestConn struct {
	// rowsAffected controls what the UPDATE returns.
	rowsAffected       int64
	authzGranted       bool
	areaCode           string
	actorID            string
	tenantID           string
	assertedCaps       []string
	lastUpdateArgs     []driver.Value
	scheduleGeneration int64
}

func (c *publishTestConn) setAssertedCaps(raw string) {
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

func (c *publishTestConn) assertedCapsRawValue() any {
	if len(c.assertedCaps) == 0 {
		return nil
	}
	items := make([]map[string]string, 0, len(c.assertedCaps))
	for _, capability := range c.assertedCaps {
		items = append(items, map[string]string{
			"cap":  capability,
			"area": c.areaCode,
		})
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return nil
	}
	return string(raw)
}

func (c *publishTestConn) hasAssertedCap(capability string) bool {
	for _, asserted := range c.assertedCaps {
		if asserted == capability {
			return true
		}
	}
	return false
}

func (c *publishTestConn) Prepare(query string) (driver.Stmt, error) {
	ra := c.rowsAffected
	if !strings.Contains(strings.ToLower(query), "update") {
		// Non-UPDATE statements (governance_events INSERT, etc.) always succeed with 1.
		ra = 1
	}
	return &publishTestStmt{query: query, rowsAffected: ra, conn: c}, nil
}

func (c *publishTestConn) Close() error              { return nil }
func (c *publishTestConn) Begin() (driver.Tx, error) { return c, nil }
func (c *publishTestConn) Commit() error             { return nil }
func (c *publishTestConn) Rollback() error           { return nil }

type publishTestDriver struct{ conn *publishTestConn }

func (d *publishTestDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

func newPublishTestHarness(t *testing.T, rowsAffected int64, authzGranted ...bool) (*sql.DB, *publishTestConn) {
	t.Helper()
	granted := true
	if len(authzGranted) > 0 {
		granted = authzGranted[0]
	}
	conn := &publishTestConn{
		rowsAffected:       rowsAffected,
		authzGranted:       granted,
		areaCode:           "QA",
		actorID:            "user-1",
		tenantID:           tenant.DevTenantID,
		scheduleGeneration: 1,
	}
	name := fmt.Sprintf("publish_test_%p", conn)
	sql.Register(name, &publishTestDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open publish test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, conn
}

// newPublishTestDB registers a unique driver per test and returns a *sql.DB.
func newPublishTestDB(t *testing.T, rowsAffected int64, authzGranted ...bool) *sql.DB {
	t.Helper()
	db, _ := newPublishTestHarness(t, rowsAffected, authzGranted...)
	return db
}

type fakeScheduledPublishEnqueuer struct {
	inputs []ScheduledPublishJobInput
	err    error
}

func (f *fakeScheduledPublishEnqueuer) EnqueueScheduledPublishTx(_ context.Context, _ *sql.Tx, input ScheduledPublishJobInput) error {
	f.inputs = append(f.inputs, input)
	return f.err
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPublishApproved_HappyPath(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	inst := &domain.Instance{
		ID:              "inst-uuid-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 3,
	}

	repo := &fakePublishRepo{
		instance:                   inst,
		currentPublishedHeadForDoc: "doc-published-1",
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	enqueuer := &fakeScheduledPublishEnqueuer{}
	svc := (&PublishService{repo: repo, emitter: emitter, clock: clock}).WithScheduledPublishEnqueuer(enqueuer)
	// rowsAffected=1 → UPDATE matched one document row.
	db := newPublishTestDB(t, 1, true)

	req := PublishRequest{
		TenantID:    "tenant-uuid-1",
		InstanceID:  "inst-uuid-1",
		PublishedBy: "user-1",
	}

	result, err := svc.PublishApproved(context.Background(), db, req)
	if err != nil {
		t.Fatalf("PublishApproved: unexpected error: %v", err)
	}
	if result.DocumentID != "doc-uuid-1" {
		t.Errorf("result.DocumentID = %q; want %q", result.DocumentID, "doc-uuid-1")
	}
	if result.NewStatus != "published" {
		t.Errorf("result.NewStatus = %q; want %q", result.NewStatus, "published")
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("expected 1 governance event; got %d", len(emitter.Events))
	}
	ev := emitter.Events[0]
	if ev.EventType != "document_published" {
		t.Errorf("event type = %q; want %q", ev.EventType, "document_published")
	}
	if ev.ResourceID != "doc-uuid-1" {
		t.Errorf("event resource_id = %q; want %q", ev.ResourceID, "doc-uuid-1")
	}
	if ev.ActorUserID != "user-1" {
		t.Errorf("event actor_user_id = %q; want %q", ev.ActorUserID, "user-1")
	}
}

func TestPublishApproved_NotApprovedInstance(t *testing.T) {
	tests := []struct {
		name   string
		status domain.InstanceStatus
	}{
		{"in_progress", domain.InstanceInProgress},
		{"rejected", domain.InstanceRejected},
		{"cancelled", domain.InstanceCancelled},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inst := &domain.Instance{
				ID:         "inst-uuid-2",
				TenantID:   "tenant-uuid-1",
				DocumentID: "doc-uuid-2",
				Status:     tc.status,
			}

			repo := &fakePublishRepo{instance: inst}
			emitter := &MemoryEmitter{}
			clock := fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)}

			svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
			// rowsAffected irrelevant — should never reach UPDATE.
			db := newPublishTestDB(t, 0, true)

			req := PublishRequest{
				TenantID:    "tenant-uuid-1",
				InstanceID:  "inst-uuid-2",
				PublishedBy: "user-1",
			}

			_, err := svc.PublishApproved(context.Background(), db, req)
			if err == nil {
				t.Fatal("expected ErrInstanceNotApproved; got nil")
			}
			if !errors.Is(err, ErrInstanceNotApproved) {
				t.Errorf("expected errors.Is(err, ErrInstanceNotApproved); got %v", err)
			}
			if len(emitter.Events) != 0 {
				t.Errorf("no governance event should be emitted; got %d", len(emitter.Events))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SchedulePublish tests
// ---------------------------------------------------------------------------

func TestSchedulePublish_HappyPath(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour) // strictly after now

	inst := &domain.Instance{
		ID:              "inst-sched-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-sched-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 5,
	}

	repo := &fakePublishRepo{
		instance:                   inst,
		currentPublishedHeadForDoc: "doc-published-1",
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}
	enqueuer := &fakeScheduledPublishEnqueuer{}

	svc := (&PublishService{repo: repo, emitter: emitter, clock: clock}).WithScheduledPublishEnqueuer(enqueuer)
	// rowsAffected=1 → UPDATE matched one document row.
	db := newPublishTestDB(t, 1, true)

	req := SchedulePublishRequest{
		TenantID:      "tenant-uuid-1",
		InstanceID:    "inst-sched-1",
		EffectiveDate: future,
		ScheduledBy:   "user-sched-1",
	}

	result, err := svc.SchedulePublish(context.Background(), db, req)
	if err != nil {
		t.Fatalf("SchedulePublish: unexpected error: %v", err)
	}
	if result.DocumentID != "doc-sched-1" {
		t.Errorf("result.DocumentID = %q; want %q", result.DocumentID, "doc-sched-1")
	}
	if !result.EffectiveDate.Equal(future) {
		t.Errorf("result.EffectiveDate = %v; want %v", result.EffectiveDate, future)
	}
	if result.ScheduleGeneration != 1 {
		t.Errorf("result.ScheduleGeneration = %d; want 1", result.ScheduleGeneration)
	}
	if len(enqueuer.inputs) != 1 {
		t.Fatalf("enqueue calls = %d; want 1", len(enqueuer.inputs))
	}
	if enqueuer.inputs[0].DocumentID != "doc-sched-1" {
		t.Fatalf("enqueue document id = %q; want doc-sched-1", enqueuer.inputs[0].DocumentID)
	}
	if enqueuer.inputs[0].ExpectedRevisionVersion != 6 {
		t.Fatalf("enqueue expected revision = %d; want 6", enqueuer.inputs[0].ExpectedRevisionVersion)
	}
	if enqueuer.inputs[0].ScheduleGeneration != 1 {
		t.Fatalf("enqueue schedule generation = %d; want 1", enqueuer.inputs[0].ScheduleGeneration)
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("expected 1 governance event; got %d", len(emitter.Events))
	}
	ev := emitter.Events[0]
	if ev.EventType != "publish_scheduled" {
		t.Errorf("event type = %q; want %q", ev.EventType, "publish_scheduled")
	}
	if ev.ResourceID != "doc-sched-1" {
		t.Errorf("event resource_id = %q; want %q", ev.ResourceID, "doc-sched-1")
	}
	if ev.ActorUserID != "user-sched-1" {
		t.Errorf("event actor_user_id = %q; want %q", ev.ActorUserID, "user-sched-1")
	}
}

func TestSchedulePublish_PersistsSupersededDocumentID(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)

	inst := &domain.Instance{
		ID:              "inst-sched-supersede-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-sched-supersede-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 2,
	}

	repo := &fakePublishRepo{
		instance:                   inst,
		currentPublishedHeadForDoc: "doc-published-1",
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	db, conn := newPublishTestHarness(t, 1, true)

	_, err := svc.SchedulePublish(context.Background(), db, SchedulePublishRequest{
		TenantID:             "tenant-uuid-1",
		InstanceID:           "inst-sched-supersede-1",
		ExpectedRevision:     2,
		EffectiveDate:        future,
		ScheduledBy:          "user-sched-1",
		SupersededDocumentID: "doc-published-1",
	})
	if err != nil {
		t.Fatalf("SchedulePublish: unexpected error: %v", err)
	}
	if len(conn.lastUpdateArgs) != 5 {
		t.Fatalf("update args = %d, want 5", len(conn.lastUpdateArgs))
	}
	if got := conn.lastUpdateArgs[1]; got != "doc-published-1" {
		t.Fatalf("superseded document arg = %v, want doc-published-1", got)
	}
	if repo.loadCurrentPublishedHeadCalls != 1 {
		t.Fatalf("LoadCurrentPublishedHeadForDocument calls = %d, want 1", repo.loadCurrentPublishedHeadCalls)
	}
}

func TestSchedulePublish_RejectsSelfSupersede(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)

	inst := &domain.Instance{
		ID:              "inst-sched-self-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-sched-self-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 2,
	}

	repo := &fakePublishRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	db := newPublishTestDB(t, 1, true)

	_, err := svc.SchedulePublish(context.Background(), db, SchedulePublishRequest{
		TenantID:             "tenant-uuid-1",
		InstanceID:           "inst-sched-self-1",
		ExpectedRevision:     2,
		EffectiveDate:        future,
		ScheduledBy:          "user-sched-1",
		SupersededDocumentID: "doc-sched-self-1",
	})
	if err == nil {
		t.Fatal("expected self-supersede validation error; got nil")
	}
	if !errors.Is(err, repository.ErrInvalidScheduledSupersedeTarget) {
		t.Fatalf("error = %v, want ErrInvalidScheduledSupersedeTarget", err)
	}
	if repo.validateSupersedeCalls != 0 {
		t.Fatalf("validate calls = %d, want 0 for self-supersede", repo.validateSupersedeCalls)
	}
}

func TestSchedulePublish_ValidatesSupersedeTarget(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	expectedErr := repository.ErrInvalidScheduledSupersedeTarget

	inst := &domain.Instance{
		ID:              "inst-sched-validate-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-sched-validate-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 2,
	}

	repo := &fakePublishRepo{
		instance:                   inst,
		currentPublishedHeadForDoc: "doc-published-current",
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	db := newPublishTestDB(t, 1, true)

	_, err := svc.SchedulePublish(context.Background(), db, SchedulePublishRequest{
		TenantID:             "tenant-uuid-1",
		InstanceID:           "inst-sched-validate-1",
		ExpectedRevision:     2,
		EffectiveDate:        future,
		ScheduledBy:          "user-sched-1",
		SupersededDocumentID: "doc-published-1",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
	if repo.validateSupersedeCalls != 0 {
		t.Fatalf("validate calls = %d, want 0", repo.validateSupersedeCalls)
	}
	if repo.validateSupersedeTenantID != "tenant-uuid-1" {
		t.Fatalf("tenant id = %q, want tenant-uuid-1", repo.validateSupersedeTenantID)
	}
	if repo.loadCurrentPublishedHeadDocID != "doc-sched-validate-1" {
		t.Fatalf("document id = %q, want doc-sched-validate-1", repo.loadCurrentPublishedHeadDocID)
	}
}

func TestSchedulePublish_PastDate(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Second) // strictly before now

	repo := &fakePublishRepo{} // LoadInstance must never be called
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	// rowsAffected irrelevant — guard fires before any tx is opened.
	db := newPublishTestDB(t, 0, true)

	req := SchedulePublishRequest{
		TenantID:      "tenant-uuid-1",
		InstanceID:    "inst-sched-2",
		EffectiveDate: past,
		ScheduledBy:   "user-sched-2",
	}

	_, err := svc.SchedulePublish(context.Background(), db, req)
	if err == nil {
		t.Fatal("expected ErrEffectiveDateInPast; got nil")
	}
	if !errors.Is(err, ErrEffectiveDateInPast) {
		t.Errorf("expected errors.Is(err, ErrEffectiveDateInPast); got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted; got %d", len(emitter.Events))
	}
}

func TestSchedulePublish_RejectsMismatchedRevisionVersion(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)

	inst := &domain.Instance{
		ID:              "inst-sched-mismatch-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-sched-mismatch-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 7,
	}

	repo := &fakePublishRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	// rowsAffected would be ignored because explicit precondition fails before UPDATE.
	db := newPublishTestDB(t, 1, true)

	req := SchedulePublishRequest{
		TenantID:         "tenant-uuid-1",
		InstanceID:       "inst-sched-mismatch-1",
		ExpectedRevision: 6,
		EffectiveDate:    future,
		ScheduledBy:      "user-sched-mismatch-1",
	}

	_, err := svc.SchedulePublish(context.Background(), db, req)
	if err == nil {
		t.Fatal("expected ErrStaleRevision; got nil")
	}
	if !errors.Is(err, repository.ErrStaleRevision) {
		t.Fatalf("expected ErrStaleRevision; got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted; got %d", len(emitter.Events))
	}
}

func TestSchedulePublish_AllowsMissingExpectedRevision(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)

	inst := &domain.Instance{
		ID:              "inst-sched-no-precondition-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-sched-no-precondition-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 4,
	}

	repo := &fakePublishRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}

	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	db := newPublishTestDB(t, 1, true)

	req := SchedulePublishRequest{
		TenantID:      "tenant-uuid-1",
		InstanceID:    "inst-sched-no-precondition-1",
		EffectiveDate: future,
		ScheduledBy:   "user-sched-no-precondition-1",
	}

	result, err := svc.SchedulePublish(context.Background(), db, req)
	if err != nil {
		t.Fatalf("SchedulePublish: unexpected error: %v", err)
	}
	if result.DocumentID != "doc-sched-no-precondition-1" {
		t.Errorf("result.DocumentID = %q; want %q", result.DocumentID, "doc-sched-no-precondition-1")
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("expected 1 governance event; got %d", len(emitter.Events))
	}
}

func TestSchedulePublish_AutoPersistsCurrentPublishedHeadWhenClientOmitsTarget(t *testing.T) {
	now := time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC)
	inst := &domain.Instance{
		ID:              "inst-uuid-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-uuid-target",
		Status:          domain.InstanceApproved,
		RevisionVersion: 9,
	}

	repo := &fakePublishRepo{
		instance:                   inst,
		currentPublishedHeadForDoc: "doc-uuid-head",
	}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}
	enqueuer := &fakeScheduledPublishEnqueuer{}
	svc := (&PublishService{repo: repo, emitter: emitter, clock: clock}).WithScheduledPublishEnqueuer(enqueuer)
	db, conn := newPublishTestHarness(t, 1, true)
	conn.scheduleGeneration = 4

	_, err := svc.SchedulePublish(context.Background(), db, SchedulePublishRequest{
		TenantID:         "tenant-uuid-1",
		InstanceID:       "inst-uuid-1",
		ExpectedRevision: 9,
		EffectiveDate:    now.Add(30 * time.Minute),
		ScheduledBy:      "user-1",
	})
	if err != nil {
		t.Fatalf("SchedulePublish: unexpected error: %v", err)
	}
	if got := conn.lastUpdateArgs[1]; got != "doc-uuid-head" {
		t.Fatalf("superseded document arg = %v, want doc-uuid-head", got)
	}
}

func TestPublishApproved_CapabilityDenied(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	inst := &domain.Instance{
		ID:              "inst-authz-1",
		TenantID:        "tenant-uuid-1",
		DocumentID:      "doc-authz-1",
		Status:          domain.InstanceApproved,
		RevisionVersion: 3,
	}

	repo := &fakePublishRepo{instance: inst}
	emitter := &MemoryEmitter{}
	clock := fixedClock{t: now}
	svc := &PublishService{repo: repo, emitter: emitter, clock: clock}
	db := newPublishTestDB(t, 1, false)

	req := PublishRequest{
		TenantID:    "tenant-uuid-1",
		InstanceID:  "inst-authz-1",
		PublishedBy: "user-1",
	}

	_, err := svc.PublishApproved(context.Background(), db, req)
	if err == nil {
		t.Fatal("expected ErrCapabilityDenied; got nil")
	}
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("expected ErrCapabilityDenied; got %v", err)
	}
	if denied.Capability != "doc.publish" {
		t.Errorf("capability = %q; want %q", denied.Capability, "doc.publish")
	}
	if len(emitter.Events) != 0 {
		t.Errorf("no governance event should be emitted on denied capability; got %d", len(emitter.Events))
	}
}
