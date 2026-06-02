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

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type routeAdminRows struct {
	cols   []string
	values []driver.Value
	done   bool
}

func (r *routeAdminRows) Columns() []string { return r.cols }
func (r *routeAdminRows) Close() error      { return nil }
func (r *routeAdminRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	for i, v := range r.values {
		dest[i] = v
	}
	return nil
}

type routeAdminEmptyRows struct{ cols []string }

func (r routeAdminEmptyRows) Columns() []string         { return r.cols }
func (r routeAdminEmptyRows) Close() error              { return nil }
func (r routeAdminEmptyRows) Next([]driver.Value) error { return io.EOF }

type routeAdminResult struct{ rowsAffected int64 }

func (r routeAdminResult) LastInsertId() (int64, error) { return 0, nil }
func (r routeAdminResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type routeAdminStmt struct {
	conn  *routeAdminConn
	query string
}

func (s *routeAdminStmt) Close() error  { return nil }
func (s *routeAdminStmt) NumInput() int { return -1 }

func (s *routeAdminStmt) Exec(args []driver.Value) (driver.Result, error) {
	lower := strings.ToLower(s.query)

	if strings.Contains(lower, "set_config('metaldocs.tenant_id'") && len(args) > 0 {
		if v, ok := args[0].(string); ok {
			s.conn.capturedTenantGUC = v
		}
	}
	if strings.Contains(lower, "set_config('metaldocs.actor_id'") && len(args) > 0 {
		if v, ok := args[0].(string); ok {
			s.conn.capturedActorGUC = v
		}
	}
	if strings.Contains(lower, "update approval_routes") && strings.Contains(lower, "set active = false") {
		if s.conn.deactivateErr != nil {
			return nil, s.conn.deactivateErr
		}
		if s.conn.deactivateNoRows {
			return routeAdminResult{rowsAffected: 0}, nil
		}
		return routeAdminResult{rowsAffected: 1}, nil
	}
	if strings.Contains(lower, "insert into approval_route_stages") {
		s.conn.stageInsertExecCount++
		s.conn.stageInsertArgCount = len(args)
		return routeAdminResult{rowsAffected: int64(len(args) / 9)}, nil
	}
	return routeAdminResult{rowsAffected: 1}, nil
}

func (s *routeAdminStmt) Query(args []driver.Value) (driver.Rows, error) {
	lower := strings.ToLower(s.query)

	if strings.Contains(lower, "set_config('metaldocs.tenant_id'") && len(args) > 0 {
		if v, ok := args[0].(string); ok {
			s.conn.capturedTenantGUC = v
		}
	}
	if strings.Contains(lower, "set_config('metaldocs.actor_id'") && len(args) > 0 {
		if v, ok := args[0].(string); ok {
			s.conn.capturedActorGUC = v
		}
	}

	if strings.Contains(lower, "select exists") && strings.Contains(lower, "iam_user_roles") {
		return &routeAdminRows{cols: []string{"exists"}, values: []driver.Value{false}}, nil
	}
	if strings.Contains(lower, "select exists") && strings.Contains(lower, "role_capabilities") {
		return &routeAdminRows{
			cols:   []string{"exists"},
			values: []driver.Value{s.conn.authzGranted},
		}, nil
	}
	if strings.Contains(lower, "current_setting('metaldocs.asserted_caps'") {
		return &routeAdminRows{cols: []string{"v"}, values: []driver.Value{nil}}, nil
	}
	if strings.Contains(lower, "current_setting('metaldocs.tenant_id'") {
		return &routeAdminRows{cols: []string{"v"}, values: []driver.Value{s.conn.tenantID}}, nil
	}
	if strings.Contains(lower, "current_setting('metaldocs.actor_id'") {
		return &routeAdminRows{cols: []string{"v"}, values: []driver.Value{s.conn.actorID}}, nil
	}
	if strings.Contains(lower, "set_config") {
		return &routeAdminRows{cols: []string{"v"}, values: []driver.Value{"ok"}}, nil
	}
	if strings.Contains(lower, "insert into approval_routes") && strings.Contains(lower, "returning id") {
		return &routeAdminRows{
			cols:   []string{"id"},
			values: []driver.Value{s.conn.createdRouteID},
		}, nil
	}
	if strings.Contains(lower, "from approval_routes") && strings.Contains(lower, "for update") {
		if !s.conn.routeExists {
			return routeAdminEmptyRows{cols: []string{"id", "version", "active"}}, nil
		}
		return &routeAdminRows{
			cols:   []string{"id", "version", "active"},
			values: []driver.Value{s.conn.lockedRouteID, int64(s.conn.lockedRouteVersion), s.conn.lockedRouteActive},
		}, nil
	}
	if strings.Contains(lower, "update approval_routes") && strings.Contains(lower, "returning version") {
		if s.conn.updateErr != nil {
			return nil, s.conn.updateErr
		}
		if s.conn.updateNoRows {
			return routeAdminEmptyRows{cols: []string{"version"}}, nil
		}
		return &routeAdminRows{
			cols:   []string{"version"},
			values: []driver.Value{int64(s.conn.newVersion)},
		}, nil
	}
	return routeAdminEmptyRows{cols: []string{"v"}}, nil
}

type routeAdminConn struct {
	authzGranted        bool
	actorID             string
	tenantID            string
	createdRouteID      string
	lockedRouteID       string
	lockedRouteVersion  int
	lockedRouteActive   bool
	lockedRouteStateSet bool
	routeExists         bool
	newVersion          int
	updateErr           error
	deactivateErr       error
	updateNoRows        bool
	deactivateNoRows    bool
	capturedTenantGUC   string
	capturedActorGUC    string
	stageInsertExecCount int
	stageInsertArgCount  int
}

func (c *routeAdminConn) Prepare(query string) (driver.Stmt, error) {
	return &routeAdminStmt{conn: c, query: query}, nil
}
func (c *routeAdminConn) Close() error              { return nil }
func (c *routeAdminConn) Begin() (driver.Tx, error) { return c, nil }
func (c *routeAdminConn) Commit() error             { return nil }
func (c *routeAdminConn) Rollback() error           { return nil }

type routeAdminDriver struct{ conn *routeAdminConn }

func (d *routeAdminDriver) Open(_ string) (driver.Conn, error) { return d.conn, nil }

func newRouteAdminTestDB(t *testing.T, conn *routeAdminConn) *sql.DB {
	t.Helper()
	if conn.actorID == "" {
		conn.actorID = "user-1"
	}
	if conn.tenantID == "" {
		conn.tenantID = tenant.DevTenantID
	}
	if conn.createdRouteID == "" {
		conn.createdRouteID = "route-1"
	}
	if conn.lockedRouteID == "" {
		conn.lockedRouteID = "route-1"
	}
	if conn.lockedRouteVersion == 0 {
		conn.lockedRouteVersion = 2
	}
	if !conn.lockedRouteStateSet {
		conn.lockedRouteActive = true
	}
	if conn.newVersion == 0 {
		conn.newVersion = 2
	}
	name := fmt.Sprintf("route_admin_test_%p", conn)
	sql.Register(name, &routeAdminDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open route admin test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func validRouteStages() []domain.Stage {
	return []domain.Stage{
		{
			Order:              1,
			Name:               "quality",
			RequiredRole:       "qa_reviewer",
			RequiredCapability: "workflow.sign",
			AreaCode:           "tenant",
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
		},
		{
			Order:              2,
			Name:               "approval",
			RequiredRole:       "qa_manager",
			RequiredCapability: "workflow.sign",
			AreaCode:           "tenant",
			Quorum:             domain.QuorumAllOf,
			OnEligibilityDrift: domain.DriftFailStage,
		},
	}
}

func TestRouteAdminCreate_HappyPath(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
	}

	out, err := svc.Create(context.Background(), db, CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if out.RouteID != "route-1" {
		t.Errorf("RouteID = %q; want %q", out.RouteID, "route-1")
	}
	if len(emitter.Events) != 1 || emitter.Events[0].EventType != "route.config.created" {
		t.Errorf("expected 1 route.config.created event; got %v", emitter.Events)
	}
}

func TestRouteAdminCreate_CapDenied(t *testing.T) {
	conn := &routeAdminConn{authzGranted: false}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), db, CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      validRouteStages(),
	})
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Errorf("expected ErrCapabilityDenied; got %v", err)
	}
}

func TestRouteAdminCreate_InvalidRoute(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), db, CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      nil,
	})
	if err == nil || !strings.Contains(err.Error(), "route must have at least one stage") {
		t.Fatalf("expected validation error; got %v", err)
	}
	if len(emitter.Events) != 0 {
		t.Fatalf("expected no events on validation failure; got %d", len(emitter.Events))
	}
}

func TestRouteAdminUpdate_HappyPath(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted: true,
		routeExists:  true,
		newVersion:   4,
	}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
	}

	out, err := svc.Update(context.Background(), db, UpdateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		Name:            "PO Route v2",
		ActorUserID:     "user-1",
		ExpectedVersion: 3,
		Stages:          validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Update: unexpected error: %v", err)
	}
	if out.RouteID != "route-1" {
		t.Errorf("RouteID = %q; want %q", out.RouteID, "route-1")
	}
	if out.NewVersion != 4 {
		t.Errorf("NewVersion = %d; want %d", out.NewVersion, 4)
	}
	if len(emitter.Events) != 1 || emitter.Events[0].EventType != "route.config.updated" {
		t.Errorf("expected 1 route.config.updated event; got %v", emitter.Events)
	}
}

func TestRouteAdminUpdate_RouteInUse(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted: true,
		routeExists:  true,
		updateErr: &pgconn.PgError{
			Code:    "P0001",
			Message: "ErrRouteInUse: route xyz is referenced by one or more approval instances and cannot be modified",
		},
	}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Update(context.Background(), db, UpdateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		Name:            "PO Route v2",
		ActorUserID:     "user-1",
		ExpectedVersion: 3,
		Stages:          validRouteStages(),
	})
	if !errors.Is(err, repository.ErrRouteInUse) {
		t.Fatalf("expected ErrRouteInUse; got %v", err)
	}
}

func TestRouteAdminDeactivate_HappyPath(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted: true,
		routeExists:  true,
	}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
	}

	out, err := svc.Deactivate(context.Background(), db, DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		Reason:          "obsolete policy",
		ExpectedVersion: 2,
	})
	if err != nil {
		t.Fatalf("Deactivate: unexpected error: %v", err)
	}
	if out.RouteID != "route-1" {
		t.Errorf("RouteID = %q; want %q", out.RouteID, "route-1")
	}
	if len(emitter.Events) != 1 || emitter.Events[0].EventType != "route.config.deactivated" {
		t.Errorf("expected 1 route.config.deactivated event; got %v", emitter.Events)
	}
}

func TestRouteAdminUpdate_StaleVersion(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted: true,
		routeExists:  true,
		updateNoRows: true,
	}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Update(context.Background(), db, UpdateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		Name:            "PO Route v2",
		ActorUserID:     "user-1",
		ExpectedVersion: 9,
		Stages:          validRouteStages(),
	})
	if !errors.Is(err, repository.ErrStaleRevision) {
		t.Fatalf("expected ErrStaleRevision; got %v", err)
	}
}

func TestRouteAdminDeactivate_StaleVersion(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted:     true,
		routeExists:      true,
		deactivateNoRows: true,
	}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Deactivate(context.Background(), db, DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		Reason:          "stale check",
		ExpectedVersion: 5,
	})
	if !errors.Is(err, repository.ErrStaleRevision) {
		t.Fatalf("expected ErrStaleRevision; got %v", err)
	}
}

func TestRouteAdminDeactivate_AlreadyInactive(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted:        true,
		routeExists:         true,
		lockedRouteVersion:  5,
		lockedRouteActive:   false,
		lockedRouteStateSet: true,
	}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Deactivate(context.Background(), db, DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		Reason:          "already inactive check",
		ExpectedVersion: 5,
	})
	if !errors.Is(err, ErrRouteAlreadyInactive) {
		t.Fatalf("expected ErrRouteAlreadyInactive; got %v", err)
	}
}

// --- PR-2 hardening: idempotency, reason, single-tx list ---

type memoryRouteAdminIdempStore struct {
	create     map[string]*memoryRouteAdminSlot
	update     map[string]*memoryRouteAdminSlot
	deactivate map[string]*memoryRouteAdminSlot
}

type memoryRouteAdminSlot struct {
	hash    string
	replay  *RouteAdminReplay
	pending bool
}

type memoryRouteAdminCommitter struct {
	slot *memoryRouteAdminSlot
}

func (c *memoryRouteAdminCommitter) Complete(routeID string, newVersion *int) error {
	c.slot.pending = false
	c.slot.replay = &RouteAdminReplay{RouteID: routeID, NewVersion: newVersion}
	return nil
}

func (c *memoryRouteAdminCommitter) Fail(_ error) error {
	c.slot.pending = false
	c.slot.replay = nil
	return nil
}

func newMemoryRouteAdminIdempStore() *memoryRouteAdminIdempStore {
	return &memoryRouteAdminIdempStore{
		create:     map[string]*memoryRouteAdminSlot{},
		update:     map[string]*memoryRouteAdminSlot{},
		deactivate: map[string]*memoryRouteAdminSlot{},
	}
}

func memoryIdempKey(tenantID, actorID, key string) string {
	return tenantID + "|" + actorID + "|" + key
}

func (m *memoryRouteAdminIdempStore) begin(bucket map[string]*memoryRouteAdminSlot, tenantID, actorID, key, hash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error) {
	k := memoryIdempKey(tenantID, actorID, key)
	if slot, ok := bucket[k]; ok {
		if slot.hash != hash {
			return nil, nil, idempotency.ErrConflict
		}
		if slot.pending {
			return nil, nil, fmt.Errorf("idempotency: in_flight orphan (key in use)")
		}
		if slot.replay != nil {
			cp := *slot.replay
			return nil, &cp, nil
		}
		return &memoryRouteAdminCommitter{slot: slot}, nil, nil
	}
	slot := &memoryRouteAdminSlot{hash: hash, pending: true}
	bucket[k] = slot
	return &memoryRouteAdminCommitter{slot: slot}, nil, nil
}

func (m *memoryRouteAdminIdempStore) BeginCreateReplay(_ context.Context, tenantID, actorID, key, hash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error) {
	return m.begin(m.create, tenantID, actorID, key, hash)
}

func (m *memoryRouteAdminIdempStore) BeginUpdateReplay(_ context.Context, tenantID, actorID, key, hash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error) {
	return m.begin(m.update, tenantID, actorID, key, hash)
}

func (m *memoryRouteAdminIdempStore) BeginDeactivateReplay(_ context.Context, tenantID, actorID, key, hash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error) {
	return m.begin(m.deactivate, tenantID, actorID, key, hash)
}

func TestRouteAdminCreate_ReplayReturnsPriorResponse(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, createdRouteID: "route-replay-1"}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	emitter := &MemoryEmitter{}
	svc := (&RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Now()},
	}).WithIdempStore(store)

	in := CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		IdempotencyKey: "idem-create-1",
		Stages:         validRouteStages(),
	}

	first, err := svc.Create(context.Background(), db, in)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if first.RouteID != "route-replay-1" {
		t.Fatalf("first RouteID = %q", first.RouteID)
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("first Create: events=%d want 1", len(emitter.Events))
	}

	second, err := svc.Create(context.Background(), db, in)
	if err != nil {
		t.Fatalf("replay Create: %v", err)
	}
	if second.RouteID != first.RouteID {
		t.Fatalf("replay RouteID = %q; want %q", second.RouteID, first.RouteID)
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("replay must not re-run handler: events=%d want 1", len(emitter.Events))
	}
}

func TestRouteAdminCreate_IdempotencyKeyConflict(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, createdRouteID: "route-replay-2"}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}).WithIdempStore(store)

	if _, err := svc.Create(context.Background(), db, CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		IdempotencyKey: "idem-conflict",
		Stages:         validRouteStages(),
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err := svc.Create(context.Background(), db, CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route — different body",
		ActorUserID:    "user-1",
		IdempotencyKey: "idem-conflict",
		Stages:         validRouteStages(),
	})
	if !errors.Is(err, idempotency.ErrConflict) {
		t.Fatalf("expected idempotency.ErrConflict; got %v", err)
	}
}

func TestRouteAdminDeactivate_RejectsEmptyReason(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, routeExists: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Deactivate(context.Background(), db, DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		Reason:          "   ",
		ExpectedVersion: 2,
	})
	if !errors.Is(err, ErrRouteDeactivateReasonRequired) {
		t.Fatalf("expected ErrRouteDeactivateReasonRequired; got %v", err)
	}
}

func TestRouteAdminDeactivate_ReasonInGovernancePayload(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, routeExists: true}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	if _, err := svc.Deactivate(context.Background(), db, DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		Reason:          "policy retired by quality manager",
		ExpectedVersion: 2,
	}); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("events=%d want 1", len(emitter.Events))
	}
	ev := emitter.Events[0]
	if ev.EventType != "route.config.deactivated" {
		t.Fatalf("event_type = %q", ev.EventType)
	}
	var payload map[string]any
	if err := json.Unmarshal(ev.PayloadJSON, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["reason"] != "policy retired by quality manager" {
		t.Fatalf("payload.reason = %v; want %q", payload["reason"], "policy retired by quality manager")
	}
}

type stubRouteListRepo struct {
	repository.ApprovalRepository
	called   bool
	tenant   string
	routes   []repository.Route
	returnTx *sql.Tx
}

func (s *stubRouteListRepo) ListRoutesTx(_ context.Context, tx *sql.Tx, tenantID string) ([]repository.Route, error) {
	s.called = true
	s.tenant = tenantID
	s.returnTx = tx
	return s.routes, nil
}

func TestRouteAdminList_RunsUnderTenantGUC(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	repo := &stubRouteListRepo{routes: []repository.Route{{ID: "r1", TenantID: "tenant-list"}}}
	svc := &RouteAdminService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	out, err := svc.List(context.Background(), db, "tenant-list", "actor-list")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !repo.called {
		t.Fatalf("ListRoutesTx not called")
	}
	if repo.tenant != "tenant-list" {
		t.Fatalf("ListRoutesTx tenant = %q; want tenant-list", repo.tenant)
	}
	if conn.capturedTenantGUC != "tenant-list" {
		t.Fatalf("tenant GUC = %q; want tenant-list (single-tx authz+read)", conn.capturedTenantGUC)
	}
	if conn.capturedActorGUC != "actor-list" {
		t.Fatalf("actor GUC = %q; want actor-list", conn.capturedActorGUC)
	}
	if len(out.Routes) != 1 || out.Routes[0].ID != "r1" {
		t.Fatalf("unexpected routes: %+v", out.Routes)
	}
}

func TestRouteAdminCreate_BatchesStageInsert(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	stages := []domain.Stage{
		{Order: 1, Name: "s1", RequiredRole: "r1", RequiredCapability: "workflow.sign", AreaCode: "tenant", Quorum: domain.QuorumAny1Of, OnEligibilityDrift: domain.DriftReduceQuorum},
		{Order: 2, Name: "s2", RequiredRole: "r2", RequiredCapability: "workflow.sign", AreaCode: "tenant", Quorum: domain.QuorumAllOf, OnEligibilityDrift: domain.DriftFailStage},
		{Order: 3, Name: "s3", RequiredRole: "r3", RequiredCapability: "workflow.sign", AreaCode: "tenant", Quorum: domain.QuorumAllOf, OnEligibilityDrift: domain.DriftFailStage},
	}
	if _, err := svc.Create(context.Background(), db, CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO",
		ActorUserID: "user-1",
		Stages:      stages,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if conn.stageInsertExecCount != 1 {
		t.Fatalf("stage insert exec count = %d; want 1 (batched)", conn.stageInsertExecCount)
	}
	if conn.stageInsertArgCount != 27 {
		t.Fatalf("stage insert arg count = %d; want 27 (3 stages * 9 cols)", conn.stageInsertArgCount)
	}
}

