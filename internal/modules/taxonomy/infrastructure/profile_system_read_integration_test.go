//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/requesttrace"
	"metaldocs/tests/integration/testdb"
)

// bypassAuditSinkPG is the production-shaped authz.BypassAuditSink for this
// suite: it adapts the REAL audit writer, so a recorded bypass lands in
// metaldocs.audit_events with the production action/resource shape and hash
// chain. The composition root's own adapter (bypassAuditAdapter, constructed by
// wiring.NewBypassAuditSink and installed at
// apps/api/cmd/metaldocs-api/main.go:305) cannot be imported here — Go's
// internal rule limits apps/api/internal/... to apps/api/... — so its field
// mapping is mirrored verbatim while the writer and the target table stay
// production's own, which is what the commit-survival assertion needs.
type bypassAuditSinkPG struct {
	writer *auditpg.Writer
}

func (s bypassAuditSinkPG) RecordBypass(ctx context.Context, tx *sql.Tx, ev authz.BypassEvent) error {
	payload, err := json.Marshal(map[string]any{
		"kind":       string(ev.Kind),
		"capability": ev.Capability,
		"area_code":  ev.AreaCode,
	})
	if err != nil {
		payload = []byte("{}")
	}
	actor := ev.ActorID
	if actor == "" {
		actor = "system"
	}
	resourceID := ev.Capability
	if resourceID == "" {
		resourceID = string(ev.Kind)
	}
	return s.writer.RecordTx(ctx, tx, auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actor,
		Action:       "authz.bypass." + string(ev.Kind),
		ResourceType: "authz_bypass",
		ResourceID:   resourceID,
		PayloadJSON:  string(payload),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     ev.TenantID, // "" allowed for cross-tenant background sweeps
	})
}

// installBypassAuditSink wires the process-wide bypass audit sink to db for the
// duration of the test and restores the nil (auditing-off) default afterwards —
// the t.Cleanup pattern of internal/modules/iam/authz/authz_test.go:341-359, but
// with the real audit writer instead of a recording fake.
func installBypassAuditSink(t *testing.T, db *sql.DB) {
	t.Helper()
	authz.SetBypassAuditSink(bypassAuditSinkPG{writer: auditpg.NewWriter(db)})
	t.Cleanup(func() { authz.SetBypassAuditSink(nil) })
}

// countBypassAudit counts the background bypass-audit rows recorded for
// tenantID, reading on a SEPARATE connection (the non-owner metaldocs_ci role,
// so the read is RLS-truthful) with the tenant GUC seeded the way a real reader
// seeds it. A row is visible here only if the recording transaction COMMITTED:
// an insert that was rolled back is invisible to every session, which is exactly
// the property this guards.
func countBypassAudit(t *testing.T, ro *sql.DB, tenantID string) int {
	t.Helper()
	ctx := context.Background()
	tx, err := ro.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin bypass-audit read tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID); err != nil {
		t.Fatalf("seed tenant GUC for bypass-audit read: %v", err)
	}
	var n int
	if err := tx.QueryRowContext(ctx, `
SELECT count(*)
FROM metaldocs.audit_events
WHERE tenant_id = $1 AND action = 'authz.bypass.background' AND resource_type = 'authz_bypass'`,
		tenantID).Scan(&n); err != nil {
		t.Fatalf("count bypass audit rows: %v", err)
	}
	return n
}

// TestProfileRepository_GetByCodeSystem_BarrenBackgroundContext is the live
// regression guard for the release_evaluate job failure "review interval for
// profile: query profile: taxonomy: tenant context: tenant: not present in
// context": the River worker ctx carries neither tenant nor actor, so the
// identity-gated GetByCode can never serve the coordinator's preflight.
// GetByCodeSystem must read the row from a BARREN ctx — only the background
// bypass marker — with the tenant taken from the explicit parameter.
//
// It is ALSO the guard for the audited-bypass invariant (ADR 0022 Phase 11 F8):
// the sink writes the bypass event INSIDE GetByCodeSystem's own transaction, so
// a method that returned without committing would silently erase its own audit
// trail. With the production-shaped sink installed, both the found and the
// not-found path must leave a committed metaldocs.audit_events row behind.
func TestProfileRepository_GetByCodeSystem_BarrenBackgroundContext(t *testing.T) {
	db, dbName := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	shortID := strings.ReplaceAll(tenant.ID, "-", "")[:6]
	code := "psys-" + shortID
	fam := "fam-sys-" + shortID

	seedFamily(t, db, fam)
	seedProfile(t, db, code, tenant.ID, fam, false)

	installBypassAuditSink(t, db)
	readOnly := testdb.OpenAsCIRole(t, dbName)
	if n := countBypassAudit(t, readOnly, tenant.ID); n != 0 {
		t.Fatalf("bypass audit rows before any read = %d, want 0", n)
	}

	repo := NewProfileRepository(db)

	// Barren: no tenant.WithTenantID, no iamdomain.WithAuthContext — exactly the
	// shape of a River worker context.
	bgCtx := authz.WithBackgroundBypass(context.Background())

	got, err := repo.GetByCodeSystem(bgCtx, tenant.ID, domain.ProfileCode(code))
	if err != nil {
		t.Fatalf("GetByCodeSystem under barren background ctx: %v", err)
	}
	if string(got.Code) != code {
		t.Fatalf("GetByCodeSystem: code = %q, want %q", got.Code, code)
	}
	if got.ReviewIntervalDays != 365 {
		t.Fatalf("GetByCodeSystem: review_interval_days = %d, want 365", got.ReviewIntervalDays)
	}

	// The bypass audit row must have SURVIVED — i.e. the read tx committed. A
	// defer-rollback-only implementation leaves 0 rows here.
	if n := countBypassAudit(t, readOnly, tenant.ID); n != 1 {
		t.Fatalf("committed bypass audit rows after successful read = %d, want 1", n)
	}

	// A code that does not exist under this tenant still maps to the domain
	// sentinel, not a driver error.
	if _, err := repo.GetByCodeSystem(bgCtx, tenant.ID, domain.ProfileCode("pmissing-"+shortID)); !errors.Is(err, domain.ErrProfileNotFound) {
		t.Fatalf("GetByCodeSystem missing code: want ErrProfileNotFound, got %v", err)
	}

	// The bypass and the read attempt happened on the not-found path too, so its
	// audit record must survive the sentinel return exactly the same way.
	if n := countBypassAudit(t, readOnly, tenant.ID); n != 2 {
		t.Fatalf("committed bypass audit rows after not-found read = %d, want 2", n)
	}
}

// TestProfileRepository_GetByCodeSystem_FailsClosedOffBackground pins the
// fail-closed property that makes the tier-2 bypass safe: without the
// WithBackgroundBypass marker (i.e. on any HTTP-reachable path) the read is
// refused with authz.ErrBypassNotBackground before it can return a row — and,
// no bypass having happened, without recording a bypass audit event.
func TestProfileRepository_GetByCodeSystem_FailsClosedOffBackground(t *testing.T) {
	db, dbName := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	shortID := strings.ReplaceAll(tenant.ID, "-", "")[:6]
	code := "pfc-" + shortID
	fam := "fam-fc-" + shortID

	seedFamily(t, db, fam)
	seedProfile(t, db, code, tenant.ID, fam, false)

	installBypassAuditSink(t, db)
	readOnly := testdb.OpenAsCIRole(t, dbName)

	repo := NewProfileRepository(db)

	got, err := repo.GetByCodeSystem(context.Background(), tenant.ID, domain.ProfileCode(code))
	if !errors.Is(err, authz.ErrBypassNotBackground) {
		t.Fatalf("GetByCodeSystem off background ctx: want ErrBypassNotBackground, got %v", err)
	}
	if got != nil {
		t.Fatalf("GetByCodeSystem off background ctx returned a profile: %+v", got)
	}
	if n := countBypassAudit(t, readOnly, tenant.ID); n != 0 {
		t.Fatalf("bypass audit rows after refused read = %d, want 0", n)
	}
}
