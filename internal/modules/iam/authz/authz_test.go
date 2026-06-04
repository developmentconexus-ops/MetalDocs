package authz

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

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeBypassSink struct {
	events []BypassEvent
	err    error
}

func (f *fakeBypassSink) RecordBypass(_ context.Context, _ *sql.Tx, ev BypassEvent) error {
	f.events = append(f.events, ev)
	return f.err
}

type authzTestState struct {
	granted         bool
	isAdmin         bool
	actorID         string
	tenantID        string
	assertedCaps    string
	requireQueries  int
	executedQueries []string
}

type authzTestDriver struct {
	state *authzTestState
}

func (d *authzTestDriver) Open(_ string) (driver.Conn, error) {
	return &authzTestConn{state: d.state}, nil
}

type authzTestConn struct {
	state *authzTestState
}

func (c *authzTestConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("Prepare not used")
}
func (c *authzTestConn) Close() error              { return nil }
func (c *authzTestConn) Begin() (driver.Tx, error) { return authzTestTx{}, nil }
func (c *authzTestConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return authzTestTx{}, nil
}

func (c *authzTestConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.executedQueries = append(c.state.executedQueries, query)

	if strings.Contains(query, "set_config('metaldocs.asserted_caps'") {
		if len(args) != 1 {
			return nil, fmt.Errorf("set_config asserted caps: got %d args, want 1", len(args))
		}
		val, ok := args[0].Value.(string)
		if !ok {
			return nil, fmt.Errorf("set_config asserted caps arg type %T", args[0].Value)
		}
		c.state.assertedCaps = val
	}

	return driver.RowsAffected(1), nil
}

func (c *authzTestConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "SELECT EXISTS") && strings.Contains(query, "role_code = 'system_admin'"):
		return &authzTestRows{
			columns: []string{"exists"},
			rows:    [][]driver.Value{{c.state.isAdmin}},
		}, nil
	case strings.Contains(query, "SELECT EXISTS") && strings.Contains(query, "role_capabilities"):
		c.state.requireQueries++
		return &authzTestRows{
			columns: []string{"exists"},
			rows:    [][]driver.Value{{c.state.granted}},
		}, nil
	case strings.Contains(query, "current_setting('metaldocs.asserted_caps', true)"):
		return &authzTestRows{
			columns: []string{"current_setting"},
			rows:    [][]driver.Value{{c.state.assertedCaps}},
		}, nil
	case strings.Contains(query, "current_setting('metaldocs.actor_id', true)"):
		return &authzTestRows{
			columns: []string{"current_setting"},
			rows:    [][]driver.Value{{c.state.actorID}},
		}, nil
	case strings.Contains(query, "current_setting('metaldocs.tenant_id', true)"):
		return &authzTestRows{
			columns: []string{"current_setting"},
			rows:    [][]driver.Value{{c.state.tenantID}},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
}

type authzTestTx struct{}

func (authzTestTx) Commit() error   { return nil }
func (authzTestTx) Rollback() error { return nil }

type authzTestRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (r *authzTestRows) Columns() []string { return r.columns }
func (r *authzTestRows) Close() error      { return nil }
func (r *authzTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func openAuthzTestDB(t *testing.T, state *authzTestState) (*sql.DB, *sql.Tx) {
	t.Helper()

	name := fmt.Sprintf("authz_test_%p", state)
	sql.Register(name, &authzTestDriver{state: state})

	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	return db, tx
}

func TestRequire_CapGranted(t *testing.T) {
	state := &authzTestState{
		granted:      true,
		actorID:      "actor-1",
		tenantID:     tenant.DevTenantID,
		assertedCaps: `[]`,
	}
	_, tx := openAuthzTestDB(t, state)

	err := Require(WithCapCache(context.Background()), tx, "doc.submit", "AREA1")
	if err != nil {
		t.Fatalf("Require returned error: %v", err)
	}

	var got []map[string]string
	if err := json.Unmarshal([]byte(state.assertedCaps), &got); err != nil {
		t.Fatalf("Unmarshal asserted caps: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("asserted caps len = %d, want 1", len(got))
	}
	if got[0]["cap"] != "doc.submit" || got[0]["area"] != "AREA1" {
		t.Fatalf("asserted cap = %#v, want doc.submit/AREA1", got[0])
	}
}

func TestRequire_CapDenied(t *testing.T) {
	state := &authzTestState{
		granted:  false,
		actorID:  "actor-2",
		tenantID: tenant.DevTenantID,
	}
	_, tx := openAuthzTestDB(t, state)

	err := Require(WithCapCache(context.Background()), tx, "document.publish", "AREA2")
	if err == nil {
		t.Fatal("Require returned nil, want ErrCapDenied")
	}

	var denied ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("Require error = %T %v, want ErrCapDenied", err, err)
	}
	if denied.Capability != "document.publish" || denied.AreaCode != "AREA2" || denied.ActorID != "actor-2" {
		t.Fatalf("denied error = %#v", denied)
	}
}

func TestRequire_CacheHit(t *testing.T) {
	state := &authzTestState{
		granted:      true,
		actorID:      "actor-3",
		tenantID:     tenant.DevTenantID,
		assertedCaps: `[]`,
	}
	_, tx := openAuthzTestDB(t, state)
	ctx := WithCapCache(context.Background())

	if err := Require(ctx, tx, "doc.signoff", "AREA3"); err != nil {
		t.Fatalf("first Require: %v", err)
	}
	if err := Require(ctx, tx, "doc.signoff", "AREA3"); err != nil {
		t.Fatalf("second Require: %v", err)
	}
	if state.requireQueries != 1 {
		t.Fatalf("require query count = %d, want 1", state.requireQueries)
	}
}

func TestRequire_CacheScopedPerTxActorAndTenant(t *testing.T) {
	state := &authzTestState{
		granted:      true,
		actorID:      "actor-1",
		tenantID:     tenant.DevTenantID,
		assertedCaps: `[]`,
	}
	db, tx1 := openAuthzTestDB(t, state)
	ctx := WithCapCache(context.Background())

	if err := Require(ctx, tx1, "doc.signoff", "AREA3"); err != nil {
		t.Fatalf("first Require: %v", err)
	}

	state.granted = false
	state.actorID = "actor-2"
	state.tenantID = "00000000-0000-0000-0000-000000000042"

	tx2, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx tx2: %v", err)
	}
	t.Cleanup(func() { _ = tx2.Rollback() })

	err = Require(ctx, tx2, "doc.signoff", "AREA3")
	if err == nil {
		t.Fatal("second Require returned nil, want ErrCapDenied")
	}
	var denied ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("second Require error = %T %v, want ErrCapDenied", err, err)
	}
	if denied.ActorID != "actor-2" {
		t.Fatalf("denied actor = %q, want actor-2", denied.ActorID)
	}
	if state.requireQueries != 2 {
		t.Fatalf("require query count = %d, want 2", state.requireQueries)
	}
}

func TestRequire_GroupSystemAdminBypasses(t *testing.T) {
	state := &authzTestState{
		granted:      false,
		isAdmin:      true,
		actorID:      "actor-admin",
		tenantID:     tenant.DevTenantID,
		assertedCaps: `[]`,
	}
	_, tx := openAuthzTestDB(t, state)

	if err := Require(WithCapCache(context.Background()), tx, "does.not.exist", "AREA9"); err != nil {
		t.Fatalf("group system_admin bypass returned error: %v", err)
	}
	if state.requireQueries != 0 {
		t.Fatalf("require query count = %d, want 0 when bypassed before capability query", state.requireQueries)
	}
}

func TestBypassSystem(t *testing.T) {
	state := &authzTestState{}
	_, tx := openAuthzTestDB(t, state)

	// ADR 0022 Phase 7: BypassSystem requires a background-marked context.
	ctx := WithBackgroundBypass(context.Background())
	if err := BypassSystem(ctx, tx); err != nil {
		t.Fatalf("BypassSystem returned error: %v", err)
	}

	found := false
	for _, query := range state.executedQueries {
		if query == "SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("executed queries = %#v, want bypass set_config", state.executedQueries)
	}
}

// TestRequire_SystemAdminBypassEmitsAudit (ADR 0022 Phase 11 F8): every tier-2
// system_admin short-circuit emits a full-fidelity bypass audit event.
func TestRequire_SystemAdminBypassEmitsAudit(t *testing.T) {
	sink := &fakeBypassSink{}
	SetBypassAuditSink(sink)
	t.Cleanup(func() { SetBypassAuditSink(nil) })

	state := &authzTestState{isAdmin: true, actorID: "admin-1", tenantID: tenant.DevTenantID, assertedCaps: `[]`}
	_, tx := openAuthzTestDB(t, state)

	if err := Require(WithCapCache(context.Background()), tx, "document.edit", "AREA1"); err != nil {
		t.Fatalf("Require (admin) returned error: %v", err)
	}
	if len(sink.events) != 1 {
		t.Fatalf("bypass events = %d, want 1", len(sink.events))
	}
	ev := sink.events[0]
	if ev.Kind != BypassKindSystemAdmin || ev.ActorID != "admin-1" || ev.TenantID != tenant.DevTenantID ||
		ev.Capability != "document.edit" || ev.AreaCode != "AREA1" {
		t.Fatalf("bypass event = %#v", ev)
	}
}

// TestRequire_SystemAdminBypassAuditFailClosed: if the bypass audit write fails,
// Require fails too — a system_admin bypass cannot commit unaudited.
func TestRequire_SystemAdminBypassAuditFailClosed(t *testing.T) {
	sink := &fakeBypassSink{err: errors.New("audit pipe down")}
	SetBypassAuditSink(sink)
	t.Cleanup(func() { SetBypassAuditSink(nil) })

	state := &authzTestState{isAdmin: true, actorID: "admin-1", tenantID: tenant.DevTenantID, assertedCaps: `[]`}
	_, tx := openAuthzTestDB(t, state)

	if err := Require(WithCapCache(context.Background()), tx, "document.edit", "AREA1"); err == nil {
		t.Fatal("Require returned nil, want error when bypass audit fails (fail-closed)")
	}
}

// TestBypassSystem_EmitsBackgroundAudit: the background bridge emits a bypass audit
// event attributed to the GUC identity (or system/"" when unseeded).
func TestBypassSystem_EmitsBackgroundAudit(t *testing.T) {
	sink := &fakeBypassSink{}
	SetBypassAuditSink(sink)
	t.Cleanup(func() { SetBypassAuditSink(nil) })

	state := &authzTestState{actorID: "system", tenantID: ""}
	_, tx := openAuthzTestDB(t, state)

	ctx := WithBackgroundBypass(context.Background())
	if err := BypassSystem(ctx, tx); err != nil {
		t.Fatalf("BypassSystem: %v", err)
	}
	if len(sink.events) != 1 {
		t.Fatalf("bypass events = %d, want 1", len(sink.events))
	}
	if ev := sink.events[0]; ev.Kind != BypassKindBackground {
		t.Fatalf("bypass event = %#v, want background", ev)
	}
}

// TestDenyByDefault_EveryCapabilityDeniesUnprivilegedActor (ADR 0022 Phase 11 F8):
// the standing deny-by-default matrix. Every registry capability must deny an actor
// who holds no role and is not system_admin — no capability may be special-cased to
// allow. Locks the enforcement model against a future Require change that leaks a
// grant. (Seed-level grant drift is locked separately by the api-lint
// seed-registry-parity rule + the deny-default seed test in apps/api.)
func TestDenyByDefault_EveryCapabilityDeniesUnprivilegedActor(t *testing.T) {
	caps := iamdomain.AllCapabilities()
	if len(caps) == 0 {
		t.Fatal("registry returned no capabilities")
	}
	for _, c := range caps {
		c := c
		t.Run(string(c), func(t *testing.T) {
			state := &authzTestState{granted: false, isAdmin: false, actorID: "nobody", tenantID: tenant.DevTenantID}
			_, tx := openAuthzTestDB(t, state)

			area := "tenant"
			if iamdomain.IsAreaGrade(c) {
				area = "AREA1"
			}
			err := Require(WithCapCache(context.Background()), tx, string(c), area)
			var denied ErrCapDenied
			if !errors.As(err, &denied) {
				t.Fatalf("Require(%s) = %v, want ErrCapDenied (deny-by-default)", c, err)
			}
		})
	}
}

// TestBypassSystem_FailsClosedWithoutBackgroundContext locks the CWE-269 guard:
// BypassSystem must refuse (and emit no GUC) on a non-background context, so the
// tier-2 bypass can never be reached from an HTTP request path.
func TestBypassSystem_FailsClosedWithoutBackgroundContext(t *testing.T) {
	state := &authzTestState{}
	_, tx := openAuthzTestDB(t, state)

	if err := BypassSystem(context.Background(), tx); !errors.Is(err, ErrBypassNotBackground) {
		t.Fatalf("BypassSystem on non-background ctx = %v, want ErrBypassNotBackground", err)
	}
	for _, query := range state.executedQueries {
		if query == "SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)" {
			t.Fatalf("bypass GUC was set despite non-background context: %#v", state.executedQueries)
		}
	}
}
