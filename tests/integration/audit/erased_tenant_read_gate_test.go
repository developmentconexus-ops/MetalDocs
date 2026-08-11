//go:build integration

// erased_tenant_read_gate_test.go — DEC-8 / PR #121: drives the audit
// erased-tenant read gate against a real Postgres instance (testdb factory).
// A unit test with a mocked repository proves the function does what it was
// written to do; this proves the gate actually holds against a real
// tombstoned tenant, for both the realistic plaintext case (dev DB today:
// 8923/8923 audit rows plaintext, 0 tenant_keys rows) and the crypto-envelope
// case, without over-blocking an active tenant. It also proves the P1
// export-persistence fix (review round 1 on PR #121): a tenant erased
// mid-export leaves zero new audit_export_jobs rows.
//
//	go test -tags=integration ./tests/integration/audit/... -run ErasedTenantReadGate
package audit_test

import (
	"context"
	"database/sql"
	"testing"

	auditapp "metaldocs/internal/modules/audit/application"
	auditdomain "metaldocs/internal/modules/audit/domain"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	platcrypto "metaldocs/internal/platform/crypto"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// markTenantErased tombstones tenantID by setting tenants.erased_at directly.
// This gate (DEC-8) only ever consults that one column — iamdomain's
// TenantErasureReader / TenantErasureRepository.IsErased is a plain
// `SELECT erased_at ... WHERE id = $1` — so driving it directly, rather than
// running the full crypto-shred tenant-lifecycle service, keeps this test
// scoped to what the read gate actually depends on.
func markTenantErased(t *testing.T, sqlDB *sql.DB, tenantID string) {
	t.Helper()
	if _, err := sqlDB.ExecContext(context.Background(),
		`UPDATE metaldocs.tenants SET erased_at = now() WHERE id = $1::uuid`, tenantID,
	); err != nil {
		t.Fatalf("markTenantErased(%s): %v", tenantID, err)
	}
}

func newErasureGatedWriter(sqlDB *sql.DB) *auditpg.Writer {
	return auditpg.NewWriter(sqlDB).WithTenantErasureCheck(iampg.NewTenantErasureRepository(sqlDB))
}

// (a) plaintext row, erased tenant: ListEvents withholds it — the realistic
// case on this program's dev DB (8923/8923 rows plaintext today).
func TestErasedTenantReadGate_PlaintextRow_Withheld_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	writer := newErasureGatedWriter(sqlDB)

	event := newEvent(tenant.ID, "audit.gate.plaintext", `{"username":"real.person"}`)
	if err := writer.Record(context.Background(), event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	markTenantErased(t, sqlDB, tenant.ID)

	items, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: event.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListEvents() returned %d items, want 1", len(items))
	}
	if items[0].PayloadJSON != `{"redacted":"tenant-erased"}` {
		t.Fatalf("payload for erased tenant = %q, want the tenant-erased redaction marker (plaintext leaked otherwise)", items[0].PayloadJSON)
	}
}

// (b) envelope row, erased tenant: ListEvents withholds it too, and without
// attempting to decrypt — the gate must fire before the envelope branch.
func TestErasedTenantReadGate_EnvelopeRow_WithheldWithoutDecrypting_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	cryptoWriter, svc := newCryptoWriter(t, sqlDB)
	writer := cryptoWriter.WithTenantErasureCheck(iampg.NewTenantErasureRepository(sqlDB))
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	event := newEvent(tenant.ID, "audit.gate.envelope", `{"secret":"classified"}`)
	if err := writer.Record(context.Background(), event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	stored := rawPayload(t, sqlDB, event.ID)
	if !platcrypto.IsEnvelope(stored) {
		t.Fatalf("precondition failed: stored payload is not an envelope: %s", stored)
	}

	markTenantErased(t, sqlDB, tenant.ID)

	items, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: event.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListEvents() returned %d items, want 1", len(items))
	}
	if items[0].PayloadJSON != `{"redacted":"tenant-erased"}` {
		t.Fatalf("payload for erased tenant = %q, want the tenant-erased redaction marker", items[0].PayloadJSON)
	}
}

// (c) active tenant: both plaintext and envelope rows still read normally —
// a fail-closed gate that over-blocks an active tenant is a different
// defect, not a safer one.
func TestErasedTenantReadGate_ActiveTenant_ReadsNormally_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	cryptoWriter, svc := newCryptoWriter(t, sqlDB)
	writer := cryptoWriter.WithTenantErasureCheck(iampg.NewTenantErasureRepository(sqlDB))
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	plainEvent := newEvent(tenant.ID, "audit.gate.active_plain", `{"n":1}`)
	if err := auditpg.NewWriter(sqlDB).Record(context.Background(), plainEvent); err != nil {
		t.Fatalf("Record() plaintext error = %v", err)
	}
	envEvent := newEvent(tenant.ID, "audit.gate.active_envelope", `{"n":2}`)
	if err := writer.Record(context.Background(), envEvent); err != nil {
		t.Fatalf("Record() envelope error = %v", err)
	}

	// tenant is never marked erased in this test.

	// PayloadJSON is stored in a jsonb column: Postgres canonicalizes
	// whitespace/key-order on write, so the contract for the plaintext row is
	// semantic identity, not byte identity (same rationale as
	// jsonSemanticallyEqual's other use in payload_crypto_test.go, same
	// package).
	plainItems, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: plainEvent.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() plaintext error = %v", err)
	}
	if len(plainItems) != 1 || !jsonSemanticallyEqual(t, plainItems[0].PayloadJSON, plainEvent.PayloadJSON) {
		t.Fatalf("ListEvents() plaintext = %+v, want semantically %q", plainItems, plainEvent.PayloadJSON)
	}

	envItems, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: envEvent.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() envelope error = %v", err)
	}
	if len(envItems) != 1 || !jsonSemanticallyEqual(t, envItems[0].PayloadJSON, envEvent.PayloadJSON) {
		t.Fatalf("ListEvents() envelope = %+v, want semantically decrypted %q", envItems, envEvent.PayloadJSON)
	}
}

// (d) P1 (PR #121 review round 1): a tenant erased mid-export — after
// renderExportPayload has already fetched every row, but before ExportEvents
// persists the export job — must leave zero new audit_export_jobs rows.
// eraseAfterFirstRead simulates the race by tombstoning the tenant as a side
// effect of the first (and, for this small fixture, only) ListEvents page
// fetchAll performs, then lets ExportEvents run its real, unmodified
// pre-persist erasure re-check (application.Service.refuseIfTenantErased,
// wired automatically by WithExports from writer's embedded
// domain.ErasureChecker — no separate setter call) against the
// now-genuinely-erased row.
func TestErasedTenantReadGate_ExportPersistence_MidExportErasure_LeavesZeroRows_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	writer := newErasureGatedWriter(sqlDB)

	event := newEvent(tenant.ID, "audit.gate.export_race", `{"n":1}`)
	if err := writer.Record(context.Background(), event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	exportRepo := auditpg.NewExportJobRepository(sqlDB)
	reader := &eraseAfterFirstRead{Reader: writer, sqlDB: sqlDB, tenantID: tenant.ID}
	svc := auditapp.NewService(reader).
		WithExports(writer, exportRepo, writer, func(auditdomain.ExportJob) string { return "" })

	_, err := svc.ExportEvents(context.Background(), "test-actor", auditdomain.ExportFormatCSV, auditdomain.ListEventsQuery{TenantID: tenant.ID})
	if err == nil {
		t.Fatal("ExportEvents() error = nil, want ErrExportTenantErased (tenant was erased mid-export)")
	}

	got := countExportJobs(t, sqlDB, tenant.ID)
	if got != 0 {
		t.Fatalf("audit_export_jobs rows for tenant %s = %d, want 0 — erasure committed before persist, nothing should have been written", tenant.ID, got)
	}
}

// eraseAfterFirstRead wraps a real domain.Reader and, the first time
// ListEvents is called, tombstones tenantID directly in Postgres before
// returning — reproducing "erasure commits while an export is in flight"
// against the actual DB row the subsequent erasure re-check queries, rather
// than a mock standing in for that timing.
type eraseAfterFirstRead struct {
	auditdomain.Reader
	sqlDB    *sql.DB
	tenantID string
	fired    bool
}

func (r *eraseAfterFirstRead) ListEvents(ctx context.Context, q auditdomain.ListEventsQuery) ([]auditdomain.Event, bool, error) {
	items, hasMore, err := r.Reader.ListEvents(ctx, q)
	if err != nil {
		return items, hasMore, err
	}
	if !r.fired {
		r.fired = true
		if _, execErr := r.sqlDB.ExecContext(context.Background(),
			`UPDATE metaldocs.tenants SET erased_at = now() WHERE id = $1::uuid`, r.tenantID,
		); execErr != nil {
			return items, hasMore, execErr
		}
	}
	return items, hasMore, err
}

func countExportJobs(t *testing.T, sqlDB *sql.DB, tenantID string) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.audit_export_jobs WHERE tenant_id = $1::uuid`, tenantID,
	).Scan(&n); err != nil {
		t.Fatalf("countExportJobs(%s): %v", tenantID, err)
	}
	return n
}
