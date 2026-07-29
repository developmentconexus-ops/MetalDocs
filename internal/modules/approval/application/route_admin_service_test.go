package application

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
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
	copy(dest, r.values)
	return nil
}

type routeAdminEmptyRows struct{ cols []string }

func (r routeAdminEmptyRows) Columns() []string         { return r.cols }
func (r routeAdminEmptyRows) Close() error              { return nil }
func (r routeAdminEmptyRows) Next([]driver.Value) error { return io.EOF }

type routeAdminMultiRows struct {
	cols   []string
	values [][]driver.Value
	idx    int
}

func (r *routeAdminMultiRows) Columns() []string { return r.cols }
func (r *routeAdminMultiRows) Close() error      { return nil }
func (r *routeAdminMultiRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

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
	if strings.Contains(lower, "set_config('metaldocs.actor_id'") && len(args) > 1 {
		if v, ok := args[1].(string); ok {
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
	if strings.Contains(lower, "delete from approval_route_stages") {
		s.conn.stageDeleteExecCount++
		return routeAdminResult{rowsAffected: 1}, nil
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
	if strings.Contains(lower, "set_config('metaldocs.actor_id'") && len(args) > 1 {
		if v, ok := args[1].(string); ok {
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
		if len(args) >= 6 {
			if v, ok := args[4].(string); ok {
				s.conn.capturedSubjectKind = v
			}
			if v, ok := args[5].(string); ok {
				s.conn.capturedSubjectKey = v
			}
		}
		if s.conn.insertRouteErr != nil {
			return nil, s.conn.insertRouteErr
		}
		return &routeAdminRows{
			cols:   []string{"id"},
			values: []driver.Value{s.conn.createdRouteID},
		}, nil
	}
	// F-E4-2: Create now reads the persisted route back in-tx
	// (scanRouteProjectionTx) so the 201 body carries the real projection
	// instead of a two-field stub. Serve that read-back here; the `in_use`
	// EXISTS subquery on approval_instances makes the query unmistakable.
	if strings.Contains(lower, "from approval_routes") && strings.Contains(lower, "as in_use") {
		if s.conn.projectionMissing {
			return routeAdminEmptyRows{cols: []string{"id", "name", "version", "active", "created_at", "profile_code", "in_use"}}, nil
		}
		var profileCode driver.Value
		if s.conn.projectionProfileCode != nil {
			profileCode = *s.conn.projectionProfileCode
		}
		return &routeAdminRows{
			cols: []string{"id", "name", "version", "active", "created_at", "profile_code", "in_use"},
			values: []driver.Value{
				s.conn.createdRouteID,
				s.conn.projectionName,
				int64(s.conn.projectionVersion),
				s.conn.projectionActive,
				s.conn.projectionCreatedAt,
				profileCode,
				s.conn.projectionInUse,
			},
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
	if strings.Contains(lower, "from approval_route_stage_selectors") {
		stages := s.conn.stageLoadStages
		rows := &routeAdminMultiRows{
			cols:   []string{"route_stage_id", "kind", "user_id", "role", "area_code"},
			values: make([][]driver.Value, 0),
		}
		for _, st := range stages {
			stageID := fmt.Sprintf("stage-%d", st.Order)
			for _, sel := range st.Selectors {
				var userID, role, areaCode driver.Value
				if sel.UserID != "" {
					userID = sel.UserID
				}
				if sel.Role != "" {
					role = sel.Role
				}
				if sel.AreaCode != "" {
					areaCode = sel.AreaCode
				}
				rows.values = append(rows.values, []driver.Value{stageID, string(sel.Kind), userID, role, areaCode})
			}
		}
		return rows, nil
	}
	if strings.Contains(lower, "from approval_route_stages") && strings.Contains(lower, "select") {
		stages := s.conn.stageLoadStages
		rows := &routeAdminMultiRows{
			cols:   []string{"id", "stage_order", "name", "required_capability", "quorum", "quorum_m", "on_eligibility_drift", "stage_kind", "due_in_days"},
			values: make([][]driver.Value, 0, len(stages)),
		}
		for _, st := range stages {
			var qm driver.Value
			if st.QuorumM != nil {
				qm = int64(*st.QuorumM)
			} else {
				qm = nil
			}
			var dueInDays driver.Value
			if st.DueInDays != nil {
				dueInDays = int64(*st.DueInDays)
			} else {
				dueInDays = nil
			}
			kind := st.Kind
			if kind == "" {
				kind = domain.StageKindApproval
			}
			stageID := fmt.Sprintf("stage-%d", st.Order)
			rows.values = append(rows.values, []driver.Value{stageID, int64(st.Order), st.Name, st.RequiredCapability, string(st.Quorum), qm, string(st.OnEligibilityDrift), string(kind), dueInDays})
		}
		return rows, nil
	}
	return routeAdminEmptyRows{cols: []string{"v"}}, nil
}

type routeAdminConn struct {
	authzGranted         bool
	actorID              string
	tenantID             string
	createdRouteID       string
	lockedRouteID        string
	lockedRouteVersion   int
	lockedRouteActive    bool
	lockedRouteStateSet  bool
	routeExists          bool
	newVersion           int
	updateErr            error
	deactivateErr        error
	updateNoRows         bool
	deactivateNoRows     bool
	capturedTenantGUC    string
	capturedActorGUC     string
	stageInsertExecCount int
	stageInsertArgCount  int
	stageLoadStages      []domain.Stage
	stageDeleteExecCount int
	insertRouteErr       error
	capturedSubjectKind  string
	capturedSubjectKey   string

	// F-E4-2 create read-back projection (scanRouteProjectionTx).
	projectionName        string
	projectionVersion     int
	projectionActive      bool
	projectionActiveSet   bool
	projectionCreatedAt   time.Time
	projectionProfileCode *string
	projectionInUse       bool
	projectionMissing     bool
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
	// F-E4-2 read-back defaults: a freshly created route is active at
	// version 1 with no instances referencing it yet.
	if conn.projectionName == "" {
		conn.projectionName = "Test Route"
	}
	if conn.projectionVersion == 0 {
		conn.projectionVersion = 1
	}
	if !conn.projectionActiveSet {
		conn.projectionActive = true
	}
	if conn.projectionCreatedAt.IsZero() {
		conn.projectionCreatedAt = time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC)
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
			RequiredCapability: "document.signoff",
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorRoleInFixedArea, Role: "qa_reviewer", AreaCode: "tenant"},
			},
		},
		{
			Order:              2,
			Name:               "approval",
			RequiredCapability: "document.signoff",
			Quorum:             domain.QuorumAllOf,
			OnEligibilityDrift: domain.DriftFailStage,
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorRoleInFixedArea, Role: "qa_manager", AreaCode: "tenant"},
			},
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

	out, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
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

// TestRouteAdminCreate_ReturnsPersistedProjection is the F-E4-2 regression
// pin at the service layer: Create must return the projection it read back
// from the committed row (name/version/active/in_use/created_at/profile_code/
// stages), not a RouteID-only stub. The delivery layer serializes exactly this
// struct, which is why the live create response was all zero values while the
// DB row was fully persisted.
func TestRouteAdminCreate_ReturnsPersistedProjection(t *testing.T) {
	profileCode := "po"
	createdAt := time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC)
	conn := &routeAdminConn{
		authzGranted:          true,
		projectionName:        "PO Route",
		projectionVersion:     1,
		projectionActive:      true,
		projectionActiveSet:   true,
		projectionCreatedAt:   createdAt,
		projectionProfileCode: &profileCode,
		stageLoadStages:       validRouteStages(),
	}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: createdAt},
	}

	out, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if out.Name != "PO Route" {
		t.Errorf("Name = %q, want %q (was \"\" before F-E4-2)", out.Name, "PO Route")
	}
	if out.Version != 1 {
		t.Errorf("Version = %d, want 1 (was 0 before F-E4-2)", out.Version)
	}
	if !out.Active {
		t.Errorf("Active = false, want true (was false before F-E4-2)")
	}
	if out.InUse {
		t.Errorf("InUse = true, want false (no approval_instances reference a fresh route)")
	}
	if !out.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v (was the zero time before F-E4-2)", out.CreatedAt, createdAt)
	}
	if out.ProfileCode == nil || *out.ProfileCode != "po" {
		t.Errorf("ProfileCode = %v, want \"po\"", out.ProfileCode)
	}
	if len(out.Stages) != 2 {
		t.Fatalf("Stages len = %d, want 2 (was nil before F-E4-2)", len(out.Stages))
	}
	if out.Stages[0].Name != "quality" || out.Stages[1].Name != "approval" {
		t.Errorf("stage names = %q/%q, want quality/approval", out.Stages[0].Name, out.Stages[1].Name)
	}
}

// TestRouteAdminCreate_TemplateSubjectProjectionHasProfileCode pins the
// template arm of the same read-back: under ADR 0086 a template route is
// profile-keyed, so profile_code is NOT NULL and equals subject_key, and the
// projection must carry it (never a nil pointer / JSON null).
func TestRouteAdminCreate_TemplateSubjectProjectionHasProfileCode(t *testing.T) {
	tmplProfile := "po"
	conn := &routeAdminConn{
		authzGranted:          true,
		projectionName:        "Template Route",
		projectionProfileCode: &tmplProfile,
		stageLoadStages:       validRouteStages(),
	}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	out, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "Template Route",
		ActorUserID: "user-1",
		SubjectKind: "template",
		SubjectKey:  "po",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if out.ProfileCode == nil || *out.ProfileCode != "po" {
		t.Errorf("ProfileCode = %v, want \"po\" for a template route", out.ProfileCode)
	}
	if out.Name != "Template Route" {
		t.Errorf("Name = %q, want %q", out.Name, "Template Route")
	}
	if len(out.Stages) != 2 {
		t.Errorf("Stages len = %d, want 2", len(out.Stages))
	}
}

// TestRouteAdminCreate_ReplayReturnsPersistedProjection pins that the
// idempotent-replay branch is NOT a degraded path: the replay envelope stores
// only the route id, so the service must re-read the same projection rather
// than returning a RouteID-only stub on the second call.
func TestRouteAdminCreate_ReplayReturnsPersistedProjection(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted:    true,
		createdRouteID:  "route-replay-projection",
		projectionName:  "PO Route",
		stageLoadStages: validRouteStages(),
	}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}).WithIdempStore(store)

	in := CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		IdempotencyKey: "idem-create-projection",
		Stages:         validRouteStages(),
	}

	first, err := svc.Create(context.Background(), newTxRunner(db), in)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	second, err := svc.Create(context.Background(), newTxRunner(db), in)
	if err != nil {
		t.Fatalf("replay Create: %v", err)
	}
	if second.RouteID != first.RouteID {
		t.Fatalf("replay RouteID = %q, want %q", second.RouteID, first.RouteID)
	}
	if second.Name != first.Name || second.Version != first.Version || second.Active != first.Active {
		t.Fatalf("replay projection = %+v, want the same projection as the first call %+v", second, first)
	}
	if len(second.Stages) != len(first.Stages) {
		t.Fatalf("replay stages len = %d, want %d", len(second.Stages), len(first.Stages))
	}
}

func TestRouteAdminCreate_CapDenied(t *testing.T) {
	conn := &routeAdminConn{authzGranted: false}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
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

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
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

// TestRouteAdminCreate_ProfileFKViolation verifies that a 23503 violation of the
// profile FK constraint maps to ErrRouteProfileUnknown (→ 422).
func TestRouteAdminCreate_ProfileFKViolation(t *testing.T) {
	profileFKErr := &pgconn.PgError{
		Code:           "23503",
		ConstraintName: routeProfileFKConstraint,
	}
	conn := &routeAdminConn{authzGranted: true, insertRouteErr: profileFKErr}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "unknown-profile",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      validRouteStages(),
	})
	if !errors.Is(err, ErrRouteProfileUnknown) {
		t.Errorf("expected ErrRouteProfileUnknown; got %v", err)
	}
}

// TestRouteAdminCreate_OtherFKViolation verifies that a 23503 violation of a
// different FK constraint does NOT map to ErrRouteProfileUnknown and instead
// falls through to a wrapped error (preserving diagnostic context for a 500).
func TestRouteAdminCreate_OtherFKViolation(t *testing.T) {
	otherFKErr := &pgconn.PgError{
		Code:           "23503",
		ConstraintName: "some_other_fk_constraint",
	}
	conn := &routeAdminConn{authzGranted: true, insertRouteErr: otherFKErr}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      validRouteStages(),
	})
	if errors.Is(err, ErrRouteProfileUnknown) {
		t.Errorf("unexpected ErrRouteProfileUnknown for unrelated FK violation")
	}
	if err == nil {
		t.Errorf("expected an error for FK violation; got nil")
	}
	// The wrapped error must still carry FK violation context.
	if !errors.Is(err, infrastructure.ErrFKViolation) {
		t.Errorf("expected wrapped ErrFKViolation; got %v", err)
	}
}

// TestRouteAdminCreate_SubjectDefaultsToDocument verifies P2.S3's byte-equal
// default: when CreateRouteInput carries no SubjectKind/SubjectKey, the
// persisted subject must be (document, profile_code) — identical to
// pre-P2.S3 behavior.
func TestRouteAdminCreate_SubjectDefaultsToDocument(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if conn.capturedSubjectKind != string(domain.SubjectKindDocument) {
		t.Errorf("subject_kind = %q, want %q", conn.capturedSubjectKind, domain.SubjectKindDocument)
	}
	if conn.capturedSubjectKey != "po" {
		t.Errorf("subject_key = %q, want %q", conn.capturedSubjectKey, "po")
	}
}

// TestRouteAdminCreate_ExplicitDocumentSubject verifies that an explicit
// subject_kind=document + subject_key is accepted and stored as given, as
// long as the explicit key equals profile_code. Per rail R1 (M3 P3.S2b-2
// regression fix: ErrDocumentSubjectKeyMismatch), a document route's
// subject_key is the backward-compat ALIAS for profile_code — a divergent
// explicit key is now rejected; see
// TestRouteAdminCreate_DocumentSubjectKeyMismatch for that case.
func TestRouteAdminCreate_ExplicitDocumentSubject(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		SubjectKind: string(domain.SubjectKindDocument),
		SubjectKey:  "po",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if conn.capturedSubjectKind != string(domain.SubjectKindDocument) {
		t.Errorf("subject_kind = %q, want %q", conn.capturedSubjectKind, domain.SubjectKindDocument)
	}
	if conn.capturedSubjectKey != "po" {
		t.Errorf("subject_key = %q, want %q", conn.capturedSubjectKey, "po")
	}
}

// TestRouteAdminCreate_SubjectKindTemplate_ProfileKeyed verifies that
// subject_kind=template is persisted keyed by the profile it governs (ADR
// 0086): an absent subject_key defaults to ProfileCode, exactly like the
// document arm.
func TestRouteAdminCreate_SubjectKindTemplate_ProfileKeyed(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		SubjectKind: string(domain.SubjectKindTemplate),
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	if conn.capturedSubjectKind != string(domain.SubjectKindTemplate) {
		t.Errorf("subject_kind = %q, want %q", conn.capturedSubjectKind, domain.SubjectKindTemplate)
	}
	if conn.capturedSubjectKey != "po" {
		t.Errorf("subject_key = %q, want %q", conn.capturedSubjectKey, "po")
	}
}

// TestRouteAdminCreate_TemplateSubjectKeyMismatch pins the ADR 0086 twin of the
// document rule: a template route is looked up by doc_type_code, so a
// subject_key diverging from profile_code would be unfindable at template
// submit time and is rejected up front.
func TestRouteAdminCreate_TemplateSubjectKeyMismatch(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		SubjectKind: string(domain.SubjectKindTemplate),
		SubjectKey:  "tmpl-1",
		Stages:      validRouteStages(),
	})
	if !errors.Is(err, ErrTemplateSubjectKeyMismatch) {
		t.Fatalf("expected ErrTemplateSubjectKeyMismatch; got %v", err)
	}
}

// TestRouteAdminCreate_InvalidSubjectKind verifies an unknown subject_kind is
// rejected via domain.Subject.Validate (ErrInvalidSubjectKind), never
// silently accepted or defaulted.
func TestRouteAdminCreate_InvalidSubjectKind(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "po",
		Name:        "PO Route",
		ActorUserID: "user-1",
		SubjectKind: "bogus",
		SubjectKey:  "x",
		Stages:      validRouteStages(),
	})
	if !errors.Is(err, domain.ErrInvalidSubjectKind) {
		t.Fatalf("expected ErrInvalidSubjectKind; got %v", err)
	}
}

// TestRouteAdminCreate_DocumentSubjectKeyMismatch verifies that an explicit
// subject_kind=document with a subject_key that diverges from profile_code
// is rejected with ErrDocumentSubjectKeyMismatch (M3 P3.S2b-2 regression
// fix). Per rail R1, profile_code is the backward-compat alias for
// (document, profile_code) — LoadActiveRouteIDByProfile delegates to
// LoadActiveRouteIDBySubject(tenantID, "document", profileCode), so a
// document route whose subject_key differs from profile_code would be
// unfindable by that alias lookup and would regress document submission to
// ErrApprovalRouteMissing.
func TestRouteAdminCreate_DocumentSubjectKeyMismatch(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:    "tenant-1",
		ProfileCode: "ops",
		Name:        "Ops Route",
		ActorUserID: "user-1",
		SubjectKind: string(domain.SubjectKindDocument),
		SubjectKey:  "custom-key",
		Stages:      validRouteStages(),
	})
	if !errors.Is(err, ErrDocumentSubjectKeyMismatch) {
		t.Fatalf("expected ErrDocumentSubjectKeyMismatch; got %v", err)
	}
}

func TestRouteAdminUpdate_HappyPath(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted:       true,
		routeExists:        true,
		newVersion:         4,
		lockedRouteVersion: 3,
	}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
	}

	out, err := svc.Update(context.Background(), newTxRunner(db), UpdateRouteInput{
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

// TestRouteAdminUpdate_RouteInUse verifies F2's transparent supersede: an
// in-use route's Update() no longer surfaces ErrRouteInUse to the caller —
// it creates a new versioned row and supersedes the old one in the same tx.
func TestRouteAdminUpdate_RouteInUse(t *testing.T) {
	conn := &routeAdminConn{
		authzGranted:       true,
		routeExists:        true,
		lockedRouteVersion: 3,
		createdRouteID:     "route-2",
		updateErr: &pgconn.PgError{
			Code:    "P0001",
			Message: "ErrRouteInUse: route xyz is referenced by one or more approval instances and cannot be modified",
		},
	}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Now()},
	}

	out, err := svc.Update(context.Background(), newTxRunner(db), UpdateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		Name:            "PO Route v2",
		ActorUserID:     "user-1",
		ExpectedVersion: 3,
		Stages:          validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Update: unexpected error on in-use supersede: %v", err)
	}
	if out.RouteID != "route-2" {
		t.Errorf("RouteID = %q; want new superseding row id %q", out.RouteID, "route-2")
	}
	if out.NewVersion != 4 {
		t.Errorf("NewVersion = %d; want %d", out.NewVersion, 4)
	}
	if len(emitter.Events) != 1 || emitter.Events[0].EventType != "route.config.updated" {
		t.Errorf("expected 1 route.config.updated event; got %v", emitter.Events)
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

	out, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
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

	_, err := svc.Update(context.Background(), newTxRunner(db), UpdateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		Name:            "PO Route v2",
		ActorUserID:     "user-1",
		ExpectedVersion: 9,
		Stages:          validRouteStages(),
	})
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
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

	_, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		Reason:          "stale check",
		ExpectedVersion: 5,
	})
	if !errors.Is(err, infrastructure.ErrStaleRevision) {
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

	_, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
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
	failErr error
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
	if c.slot.failErr != nil {
		return c.slot.failErr
	}
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

	first, err := svc.Create(context.Background(), newTxRunner(db), in)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if first.RouteID != "route-replay-1" {
		t.Fatalf("first RouteID = %q", first.RouteID)
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("first Create: events=%d want 1", len(emitter.Events))
	}

	second, err := svc.Create(context.Background(), newTxRunner(db), in)
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

	if _, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		IdempotencyKey: "idem-conflict",
		Stages:         validRouteStages(),
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
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

// TestRouteAdminCreate_IdempotencyKeyConflict_DifferingSubjectKind is the ADR
// 0086 successor of QR-A finding A. Both kinds are now profile-keyed, so a
// document route and a template route on the SAME profile differ ONLY in
// subject_kind — and the payload hash must fold that in. If it did not, two
// creates sharing an Idempotency-Key + profile + name + stages would hash
// identically and the second caller would silently be handed the first
// caller's route of the WRONG kind instead of an idempotency conflict.
func TestRouteAdminCreate_IdempotencyKeyConflict_DifferingSubjectKind(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, createdRouteID: "route-tmpl-1"}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}).WithIdempStore(store)

	if _, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		SubjectKind:    string(domain.SubjectKindDocument),
		IdempotencyKey: "idem-kind-conflict",
		Stages:         validRouteStages(),
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		SubjectKind:    string(domain.SubjectKindTemplate), // only field that differs
		IdempotencyKey: "idem-kind-conflict",
		Stages:         validRouteStages(),
	})
	if !errors.Is(err, idempotency.ErrConflict) {
		t.Fatalf("expected idempotency.ErrConflict (different subject_kind must not replay); got %v", err)
	}
}

// TestRouteAdminCreate_IdempotencyKeyConflict_TemplateWhitespaceSubjectKey
// pins the codex gate round-3 CRITICAL finding, re-expressed for ADR 0086: the
// payload hash must not TrimSpace the subject key, because the resolved key is
// persisted EXACTLY as given (resolveCreateRouteSubject does no trim). Since a
// template route's key IS its profile code, the whitespace axis now travels in
// via ProfileCode: "po" and " po " are distinct persisted subjects, so a second
// create reusing the Idempotency-Key with only whitespace added must surface
// idempotency.ErrConflict rather than replay the first route.
func TestRouteAdminCreate_IdempotencyKeyConflict_TemplateWhitespaceSubjectKey(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, createdRouteID: "route-tmpl-ws-1"}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}).WithIdempStore(store)

	if _, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "tmpl-a",
		Name:           "Template Route",
		ActorUserID:    "user-1",
		SubjectKind:    string(domain.SubjectKindTemplate),
		IdempotencyKey: "idem-tmpl-ws-conflict",
		Stages:         validRouteStages(),
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    " tmpl-a ", // only field that differs from the first call — whitespace only
		Name:           "Template Route",
		ActorUserID:    "user-1",
		SubjectKind:    string(domain.SubjectKindTemplate),
		IdempotencyKey: "idem-tmpl-ws-conflict",
		Stages:         validRouteStages(),
	})
	if !errors.Is(err, idempotency.ErrConflict) {
		t.Fatalf("expected idempotency.ErrConflict (whitespace-differing subject key must not replay); got %v", err)
	}
}

// TestRouteAdminCreate_ReplayReturnsPriorResponse_DocumentSubject verifies
// the QR-A finding A fix does not regress the legacy document-route replay
// path: a byte-identical repeat of a document create (same body twice, same
// Idempotency-Key) still replays the cached result rather than re-running
// the handler, even though the hash now also folds in the resolved subject
// (kind=document, key=profile_code for this branch — see
// resolveCreateRouteSubject).
func TestRouteAdminCreate_ReplayReturnsPriorResponse_DocumentSubject(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, createdRouteID: "route-doc-replay-1"}
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
		IdempotencyKey: "idem-doc-replay-1",
		Stages:         validRouteStages(),
	}

	first, err := svc.Create(context.Background(), newTxRunner(db), in)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if first.RouteID != "route-doc-replay-1" {
		t.Fatalf("first RouteID = %q", first.RouteID)
	}

	second, err := svc.Create(context.Background(), newTxRunner(db), in)
	if err != nil {
		t.Fatalf("second Create (replay): %v", err)
	}
	if second.RouteID != "route-doc-replay-1" {
		t.Fatalf("second RouteID = %q, want replay of %q", second.RouteID, "route-doc-replay-1")
	}
	if len(emitter.Events) != 1 {
		t.Fatalf("replay must not re-run handler: events=%d want 1", len(emitter.Events))
	}
}

func TestRouteAdminDeactivate_RejectsEmptyReason(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, routeExists: true}
	db := newRouteAdminTestDB(t, conn)

	svc := &RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	_, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
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

	if _, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
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
	infrastructure.ApprovalRepository
	called   bool
	tenant   string
	routes   []infrastructure.Route
	returnTx db.Tx
}

func (s *stubRouteListRepo) ListRoutesTx(_ context.Context, tx db.Tx, tenantID string) ([]infrastructure.Route, error) {
	s.called = true
	s.tenant = tenantID
	s.returnTx = tx
	return s.routes, nil
}

func TestRouteAdminList_RunsUnderTenantGUC(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	repo := &stubRouteListRepo{routes: []infrastructure.Route{{ID: "r1", TenantID: "tenant-list"}}}
	svc := &RouteAdminService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Now()},
	}

	out, err := svc.List(authzCtx("tenant-list", "actor-list"), newTxRunner(db), "tenant-list", "actor-list")
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
		{Order: 1, Name: "s1", RequiredCapability: "document.signoff", Quorum: domain.QuorumAny1Of, OnEligibilityDrift: domain.DriftReduceQuorum, Selectors: []domain.ActorSelector{{Kind: domain.SelectorRoleInFixedArea, Role: "r1", AreaCode: "tenant"}}},
		{Order: 2, Name: "s2", RequiredCapability: "document.signoff", Quorum: domain.QuorumAllOf, OnEligibilityDrift: domain.DriftFailStage, Selectors: []domain.ActorSelector{{Kind: domain.SelectorRoleInFixedArea, Role: "r2", AreaCode: "tenant"}}},
		{Order: 3, Name: "s3", RequiredCapability: "document.signoff", Quorum: domain.QuorumAllOf, OnEligibilityDrift: domain.DriftFailStage, Selectors: []domain.ActorSelector{{Kind: domain.SelectorRoleInFixedArea, Role: "r3", AreaCode: "tenant"}}},
	}
	if _, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
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

func TestRouteAdminUpdate_NoOpStages_SkipsDelete(t *testing.T) {
	stages := validRouteStages()
	conn := &routeAdminConn{
		authzGranted:    true,
		routeExists:     true,
		newVersion:      3,
		stageLoadStages: stages,
	}
	db := newRouteAdminTestDB(t, conn)

	emitter := &MemoryEmitter{}
	svc := &RouteAdminService{
		emitter: emitter,
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	out, err := svc.Update(context.Background(), newTxRunner(db), UpdateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		Name:            "Same Name Different Update",
		ActorUserID:     "user-1",
		ExpectedVersion: 2,
		Stages:          stages,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if out.NewVersion != 3 {
		t.Fatalf("NewVersion = %d; want 3 (version still bumps on no-op stage)", out.NewVersion)
	}
	if conn.stageDeleteExecCount != 0 {
		t.Fatalf("stage DELETE exec count = %d; want 0 (no-op skip)", conn.stageDeleteExecCount)
	}
	if conn.stageInsertExecCount != 0 {
		t.Fatalf("stage INSERT exec count = %d; want 0 (no-op skip)", conn.stageInsertExecCount)
	}
	if len(emitter.Events) != 1 || emitter.Events[0].EventType != "route.config.updated" {
		t.Fatalf("expected route.config.updated event; got %v", emitter.Events)
	}
}

func TestRouteAdminDeactivate_ReplayBeforeReasonValidation(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, routeExists: true}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}).WithIdempStore(store)

	// Pre-seed a completed replay slot under the empty-reason hash.
	emptyReasonHash := computeDeactivateRoutePayloadHash("route-1", 2, "")
	k := memoryIdempKey("tenant-1", "user-1", "88888888-8888-4888-8888-888888888888-empty")
	store.deactivate[k] = &memoryRouteAdminSlot{
		hash:   emptyReasonHash,
		replay: &RouteAdminReplay{RouteID: "route-1"},
	}

	out, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		IdempotencyKey:  "88888888-8888-4888-8888-888888888888-empty",
		Reason:          "", // empty: would fail validation if checked before replay
		ExpectedVersion: 2,
	})
	if err != nil {
		t.Fatalf("expected replay to succeed despite empty reason; got %v", err)
	}
	if out.RouteID != "route-1" {
		t.Fatalf("RouteID = %q; want route-1 (from cached replay)", out.RouteID)
	}
}

func TestRouteAdminDeactivate_FirstCallEmptyReason_Returns400(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true, routeExists: true}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}).WithIdempStore(store)

	_, err := svc.Deactivate(context.Background(), newTxRunner(db), DeactivateRouteInput{
		TenantID:        "tenant-1",
		RouteID:         "route-1",
		ActorUserID:     "user-1",
		IdempotencyKey:  "idem-empty-first",
		Reason:          "   ",
		ExpectedVersion: 2,
	})
	if !errors.Is(err, ErrRouteDeactivateReasonRequired) {
		t.Fatalf("expected ErrRouteDeactivateReasonRequired; got %v", err)
	}
	// Slot must be released (Fail called → replay nil, pending false).
	k := memoryIdempKey("tenant-1", "user-1", "idem-empty-first")
	slot, ok := store.deactivate[k]
	if !ok {
		t.Fatalf("slot missing")
	}
	if slot.pending {
		t.Fatalf("slot still pending; expected Fail() to clear it")
	}
	if slot.replay != nil {
		t.Fatalf("slot replay should be nil after Fail; got %+v", slot.replay)
	}
}

func TestRouteAdminCreate_LogsCommitterFailError(t *testing.T) {
	conn := &routeAdminConn{authzGranted: false}
	db := newRouteAdminTestDB(t, conn)
	store := newMemoryRouteAdminIdempStore()

	// Pre-create the slot so we can wire failErr.
	failErr := errors.New("simulated fail-replay failure")
	k := memoryIdempKey("tenant-1", "user-1", "idem-fail-log")
	store.create[k] = &memoryRouteAdminSlot{
		hash:    computeCreateRoutePayloadHash("po", "PO Route", validRouteStages(), domain.Subject{Kind: domain.SubjectKindDocument, Key: "po"}),
		pending: false,
		failErr: failErr,
	}

	svc := (&RouteAdminService{
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}).WithIdempStore(store)

	// Swap slog default to capture.
	var buf bytes.Buffer
	oldDefault := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(oldDefault) })

	_, err := svc.Create(context.Background(), newTxRunner(db), CreateRouteInput{
		TenantID:       "tenant-1",
		ProfileCode:    "po",
		Name:           "PO Route",
		ActorUserID:    "user-1",
		IdempotencyKey: "idem-fail-log",
		Stages:         validRouteStages(),
	})
	if err == nil {
		t.Fatalf("Create expected to fail on cap denial")
	}

	out := buf.String()
	if !strings.Contains(out, "committer.Fail failed") {
		t.Fatalf("expected log entry; got: %s", out)
	}
	if !strings.Contains(out, "fail_err=") {
		t.Fatalf("expected fail_err attribute in log; got: %s", out)
	}
	if !strings.Contains(out, "op=create") {
		t.Fatalf("expected op=create in log; got: %s", out)
	}
}

func TestRouteAdminList_PassesThroughTotalFromRepo(t *testing.T) {
	conn := &routeAdminConn{authzGranted: true}
	db := newRouteAdminTestDB(t, conn)

	repo := &stubRouteListRepo{
		routes: []infrastructure.Route{
			{ID: "r1", TenantID: "tenant-list", Total: 3},
			{ID: "r2", TenantID: "tenant-list", Total: 3},
			{ID: "r3", TenantID: "tenant-list", Total: 3},
		},
	}
	svc := &RouteAdminService{
		repo:    repo,
		emitter: &MemoryEmitter{},
		clock:   fixedClock{t: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	out, err := svc.List(authzCtx("tenant-list", "actor-list"), newTxRunner(db), "tenant-list", "actor-list")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(out.Routes) != 3 {
		t.Fatalf("routes len = %d; want 3", len(out.Routes))
	}
	if out.Routes[0].Total != 3 {
		t.Fatalf("Total = %d; want 3", out.Routes[0].Total)
	}
}
