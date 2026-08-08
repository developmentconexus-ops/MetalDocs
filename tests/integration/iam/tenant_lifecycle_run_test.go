//go:build integration

// tenant_lifecycle_run_test.go — M7 F7.3 Task F's named integration-test
// deliverable: drives TenantLifecycleService's RUN SIDE (export fan-out +
// erase fan-out) directly against a real Postgres instance, mirroring
// tenant_lifecycle_request_test.go's (enqueue side) idiom and
// apps/jobs/cmd/metaldocs-jobs/main.go's buildTenantLifecycleWorker wiring —
// but calling svc.RunJob(ctx, jobID) directly (an exported service method)
// rather than enqueuing through River, per the task's "no River polling, no
// sleeps" instruction.
//
// Deviations from the prescribed construction, both narrow and justified:
//
//   - blobs is nil (task-sanctioned "nil-safe paths" instruction). runExport
//     hard-requires a configured object store (`iam: export requires a
//     configured object store`), so TestTenantExport_CompleteArtifact does
//     NOT call RunJob/runExport at all — it verifies the export contract at
//     the PORT layer directly (every port's ExportTenantData), which is the
//     task's own prescribed fallback ("read runExport first: if nil blobs
//     makes export fail/skip, instead verify export at the port layer").
//     Artifact-write-to-MinIO is proven by the live QA drive instead (per
//     the task instructions), not by this suite.
//
//   - RunJob is exported on TenantLifecycleService (found while reading
//     tenant_lifecycle_service.go), so no unexported-method workaround was
//     needed for the erase tests — they call svc.RunJob(ctx, jobID) directly,
//     synchronously, no River client involved.
//
//     go test -tags=integration ./tests/integration/iam/... -run TestTenant(Export|Erasure)
package iam_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/composition/tenantdata/registry"
	auditdomain "metaldocs/internal/modules/audit/domain"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	iamapp "metaldocs/internal/modules/iam/application"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	securityapp "metaldocs/internal/modules/security/application"
	securitydomain "metaldocs/internal/modules/security/domain"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	platcrypto "metaldocs/internal/platform/crypto"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/internal/platform/tenantdata"
	"metaldocs/tests/integration/testdb"
)

// mustRunKEK mints a fresh 32-byte KEK for TenantCryptoService, mirroring
// tests/integration/security/tenant_crypto_test.go's mustKEK.
func mustRunKEK(t *testing.T) []byte {
	t.Helper()
	kek := make([]byte, platcrypto.KeySize)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("rand.Read kek: %v", err)
	}
	return kek
}

// newRunSideTenantLifecycleService constructs the service exactly as
// apps/jobs/cmd/metaldocs-jobs/main.go's buildTenantLifecycleWorker does —
// the four enqueue-side-only args (tenantLookupTx, lifecycleJobInserter,
// TenantLifecycleEnqueuer, db.TxRunner) are nil, blobs is nil (task
// instruction), crypto is a real TenantCryptoService over a test KEK, and
// ports is the full production registry.AllTenantDataPorts(db) slice.
func newRunSideTenantLifecycleService(t *testing.T, sqlDB *sql.DB) (*iamapp.TenantLifecycleService, *securityapp.TenantCryptoService) {
	t.Helper()
	cryptoSvc, err := securityapp.NewTenantCryptoService(securitypg.NewTenantKeyRepository(sqlDB), mustRunKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}

	lifecycleRepo := iampg.NewTenantLifecycleRepository(sqlDB)
	svc := iamapp.NewTenantLifecycleService(
		nil, // tenantLookupTx: enqueue-side only
		nil, // lifecycleJobInserter: enqueue-side only
		lifecycleRepo,
		nil,                   // iamdomain.TenantLifecycleEnqueuer: enqueue-side only
		db.NewTxRunner(sqlDB), // runErase phase-1/phase-3 txs run through it
		auditpg.NewWriter(sqlDB),
		sqlDB,
		registry.AllTenantDataPorts(sqlDB),
		nil, // blobs: nil-safe path per task instruction; export artifact
		// write-to-MinIO is proven by the live QA drive, not this suite.
		cryptoSvc,
	)
	return svc, cryptoSvc
}

// insertPendingLifecycleJob inserts a tenant_lifecycle_jobs row directly
// (bypassing the enqueue-side authz/tripwire path entirely — this suite
// drives the RUN side only) so RunJob has a row to load. Mirrors
// TenantLifecycleRepository.InsertLifecycleJobTx's column set, run under the
// scheduler bypass token (same hatch withBypass/withBypassErr use elsewhere
// in this package) since tenant_lifecycle_jobs carries the tripwire arm from
// migration 0279.
func insertPendingLifecycleJob(t *testing.T, sqlDB *sql.DB, tenantID, kind, requestedBy string) string {
	t.Helper()
	jobID := uuid.NewString()
	if err := withBypassErr(sqlDB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
INSERT INTO metaldocs.tenant_lifecycle_jobs (id, tenant_id, kind, status, requested_by)
VALUES ($1::uuid, $2::uuid, $3, 'pending', $4)
`, jobID, tenantID, kind, requestedBy)
		return err
	}); err != nil {
		t.Fatalf("insert pending tenant_lifecycle_jobs row (%s): %v", kind, err)
	}
	return jobID
}

// seedTenantWithSpread seeds a fresh tenant with a cross-module data spread:
// a system_admin actor (also usable as the lifecycle job's requested_by), an
// EXTRA iam user beyond the admin, and 2+ sealed audit events recorded under
// a provisioned tenant crypto key — the minimum spread the task calls for
// ("an extra iam user, an audit event or two"). Returns the tenant, the
// extra user id, and the admin (actor) id.
func seedTenantWithSpread(t *testing.T, sqlDB *sql.DB, cryptoSvc *securityapp.TenantCryptoService, namePrefix string) (tenant testdb.Tenant, extraUserID, adminUserID string) {
	t.Helper()
	ctx := context.Background()

	tenant = testdb.NewTenant(t, sqlDB)
	// Unique per run — the dev DB is shared, so a deterministic id would
	// collide with leftovers from any earlier failed run.
	adminUserID = namePrefix + "-admin-" + uuid.NewString()[:8]
	testdb.SeedSystemAdmin(t, sqlDB, tenant.ID, adminUserID, namePrefix+" Admin")

	extraUser := testdb.NewUser(t, sqlDB, testdb.WithTenant(tenant.ID), testdb.WithDisplayName(namePrefix+" Extra User"))
	extraUserID = extraUser.ID

	// Seed a real metaldocs.auth_identities credential row for the admin
	// user (bypass_authz='scheduler', same idiom insertPendingLifecycleJob
	// uses for direct-SQL setup rows outside the enqueue-side authz path) —
	// this is the F7.3 Defect 1 regression fixture: without it, the
	// erasure assertion below (auth_identities count = 0 post-erase) would
	// trivially pass on an empty table and prove nothing.
	seedAuthIdentity(t, sqlDB, adminUserID)

	runner := db.NewTxRunner(sqlDB)
	provisionCtx := runContextForTenant(tenant.ID)
	if err := runner.Do(provisionCtx, func(tx *sql.Tx) error {
		return cryptoSvc.ProvisionTenantKeyTx(provisionCtx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx(%s) error = %v", tenant.ID, err)
	}

	writer := auditpg.NewWriter(sqlDB).WithPayloadCrypto(runAuditCryptoAdapter{crypto: cryptoSvc})
	for i := 0; i < 2; i++ {
		ev := auditdomain.Event{
			ID:           uuid.NewString(),
			OccurredAt:   time.Now().UTC(),
			ActorID:      adminUserID,
			Action:       "tenant.run_test.seed",
			ResourceType: "tenant",
			ResourceID:   tenant.ID,
			PayloadJSON:  `{"seed":true,"n":` + itoa(i) + `}`,
			TraceID:      "trace-" + uuid.NewString(),
			TenantID:     tenant.ID,
		}
		if err := writer.Record(ctx, ev); err != nil {
			t.Fatalf("Record() seed audit event %d error = %v", i, err)
		}
	}

	return tenant, extraUserID, adminUserID
}

// seedAuthIdentity inserts a minimal metaldocs.auth_identities credential row
// for userID directly via SQL under the scheduler bypass token, mirroring
// insertPendingLifecycleJob's direct-SQL-setup idiom in this same file.
func seedAuthIdentity(t *testing.T, sqlDB *sql.DB, userID string) {
	t.Helper()
	if err := withBypassErr(sqlDB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
INSERT INTO metaldocs.auth_identities
	(user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, failed_login_attempts, created_at, updated_at)
VALUES ($1, $1, NULL, $1, TRUE, 'x', 'bcrypt', FALSE, 0, NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING
`, userID)
		return err
	}); err != nil {
		t.Fatalf("seed auth_identities row for %s: %v", userID, err)
	}
}

func itoa(n int) string {
	digits := "0123456789"
	if n == 0 {
		return "0"
	}
	return string(digits[n])
}

// runContextForTenant seeds tenant+actor identity into context, mirroring
// tests/integration/security/tenant_crypto_test.go's ctxForTenant (this
// package already declares its own ctxForTenant-shaped helpers under other
// names, so this is named distinctly to avoid collisions across the iam_test
// package's many sibling files) — the TxRunner chokepoint (F3.1 contract)
// auto-seeds these into tx-local GUCs so ProvisionTenantKeyTx's write goes
// through tenant_keys' FORCE RLS policy exactly like a real request would.
func runContextForTenant(tenantID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, "test-actor")
	return ctx
}

// runAuditCryptoAdapter mirrors tests/integration/audit/payload_crypto_test.go's
// auditPayloadCryptoAdapter (the exact composition-root seam shape) — kept
// here, distinctly named, because this test lives in package iam_test, not
// audit_test, and cannot import that unexported type.
type runAuditCryptoAdapter struct {
	crypto *securityapp.TenantCryptoService
}

func (a runAuditCryptoAdapter) EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenant(ctx, tenantID, plaintext)
	if err != nil {
		if errors.Is(err, securitydomain.ErrKeyNotFound) || errors.Is(err, securitydomain.ErrKeyDestroyed) {
			return "", false, nil // no key / crypto-shredded → plaintext fall-through
		}
		return "", false, err // genuine failure propagates, mirroring composition-root adapter
	}
	return envelope, true, nil
}

// EncryptForTenantTx mirrors the composition-root adapter's tx-aware variant
// (same-tx key-visibility fix): delegates to TenantCrypto.EncryptForTenantTx.
func (a runAuditCryptoAdapter) EncryptForTenantTx(ctx context.Context, tx *sql.Tx, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenantTx(ctx, tx, tenantID, plaintext)
	if err != nil {
		if errors.Is(err, securitydomain.ErrKeyNotFound) || errors.Is(err, securitydomain.ErrKeyDestroyed) {
			return "", false, nil // no key / crypto-shredded → plaintext fall-through
		}
		return "", false, err // genuine failure propagates, mirroring composition-root adapter
	}
	return envelope, true, nil
}

func (a runAuditCryptoAdapter) DecryptForTenant(ctx context.Context, tenantID, envelope string) ([]byte, error) {
	return a.crypto.DecryptForTenant(ctx, tenantID, envelope)
}

// TestTenantExport_CompleteArtifact proves the export fan-out's data
// completeness at the PORT layer (blobs=nil makes runExport itself fail with
// "export requires a configured object store" — the task's own documented
// fallback for this case): the union of every registered port's
// ExportTenantData tables equals the union of their own Tables()
// declarations, audit_events export rows carry no "payload" key (the
// skeleton/never-decrypt-in-export contract), and the seeded extra iam user
// shows up in the metaldocs.iam_users export rows.
func TestTenantExport_CompleteArtifact(t *testing.T) {
	sqlDB := openDB(t)
	svc, cryptoSvc := newRunSideTenantLifecycleService(t, sqlDB)
	_ = svc // constructed to prove full wiring compiles/links; RunJob not called here (see file doc comment).

	tenant, extraUserID, _ := seedTenantWithSpread(t, sqlDB, cryptoSvc, "export-artifact")

	ports := registry.AllTenantDataPorts(sqlDB)

	declaredTables := map[string]bool{}
	exportedTables := map[string]bool{}
	var auditRows []tenantdataTableExportAlias
	var iamUsersRows []tenantdataTableExportAlias

	for _, port := range ports {
		for _, tbl := range port.Tables() {
			declaredTables[tbl] = true
		}
		exports, err := port.ExportTenantData(context.Background(), sqlDB, tenant.ID)
		if err != nil {
			t.Fatalf("ExportTenantData(%s) error = %v", port.Module(), err)
		}
		for _, te := range exports {
			exportedTables[te.Table] = true
			if te.Table == "metaldocs.audit_events" {
				auditRows = append(auditRows, tenantdataTableExportAlias{Table: te.Table, Rows: te.Rows})
			}
			if te.Table == "metaldocs.iam_users" {
				iamUsersRows = append(iamUsersRows, tenantdataTableExportAlias{Table: te.Table, Rows: te.Rows})
			}
		}
	}

	// (a) census completeness: union of returned tables == union of declared
	// Tables().
	for tbl := range declaredTables {
		if !exportedTables[tbl] {
			t.Errorf("declared table %q never appeared in any port's ExportTenantData output", tbl)
		}
	}
	for tbl := range exportedTables {
		if !declaredTables[tbl] {
			t.Errorf("exported table %q was not declared by any port's Tables()", tbl)
		}
	}

	// (b) audit_events export rows contain NO "payload" key (skeleton
	// contract — audit_events is deliberately never decrypted/exported with
	// its payload; see the iam TenantDataPort erase-order comment and the
	// audit port's own EraseTenantData "does NOT delete" carve-out for the
	// symmetric, export-side skeleton rule).
	if len(auditRows) == 0 {
		t.Fatalf("no metaldocs.audit_events export rows found for tenant %s (seeded 2)", tenant.ID)
	}
	sawSeededAuditRow := false
	for _, group := range auditRows {
		for _, row := range group.Rows {
			if containsJSONKey(t, row, "payload") {
				t.Errorf("audit_events export row contains a %q key, want skeleton (no payload): %s", "payload", string(row))
			}
			if containsJSONValue(t, row, "resource_id", tenant.ID) {
				sawSeededAuditRow = true
			}
		}
	}
	if !sawSeededAuditRow {
		t.Errorf("no exported audit_events row references tenant %s (want at least the 2 seeded rows)", tenant.ID)
	}

	// (c) seeded rows appear: the extra iam user shows in the
	// metaldocs.iam_users export.
	if len(iamUsersRows) == 0 {
		t.Fatalf("no metaldocs.iam_users export rows found for tenant %s", tenant.ID)
	}
	sawExtraUser := false
	for _, group := range iamUsersRows {
		for _, row := range group.Rows {
			if containsJSONValue(t, row, "user_id", extraUserID) {
				sawExtraUser = true
			}
		}
	}
	if !sawExtraUser {
		t.Errorf("extra seeded iam user %s not found in metaldocs.iam_users export rows", extraUserID)
	}
}

// tenantdataTableExportAlias is a local copy of tenantdata.TableExport's
// shape so this file does not need to import the concrete type name twice
// under two aliases; assignment-compatible field-for-field.
type tenantdataTableExportAlias = tenantdata.TableExport

// TestTenantErasure_ChainStaysGreen drives a real erase job to completion via
// svc.RunJob and asserts the full erasure contract: audit_events rows for
// the tenant still exist (immutable skeleton — never deleted) with an intact
// hash chain (prev_hash of row n == row_hash of row n-1, by audit_sequence,
// restricted to the tenant's own sequence slice), the tenant_keys row is
// crypto-shredded (wrapped_dek zeroed, destroyed_at set), the tenants row is
// tombstoned (name='erased', erased_at set), and governance_events / iam_users
// counts for the tenant are both 0.
func TestTenantErasure_ChainStaysGreen(t *testing.T) {
	sqlDB := openDB(t)
	svc, cryptoSvc := newRunSideTenantLifecycleService(t, sqlDB)

	tenant, extraUserID, adminUserID := seedTenantWithSpread(t, sqlDB, cryptoSvc, "erase-chain-green")
	// Capture the tenant's user ids BEFORE erasure runs — iam_users rows
	// (and, by extension, the join-key this test checks against) are
	// erased by the iam port, so this must be read pre-erase.
	formerUserIDs := []string{adminUserID, extraUserID}

	jobID := insertPendingLifecycleJob(t, sqlDB, tenant.ID, "erase", adminUserID)

	if err := svc.RunJob(context.Background(), jobID); err != nil {
		t.Fatalf("RunJob(erase) error = %v", err)
	}

	assertJobReady(t, sqlDB, jobID)
	assertAuditEventsSurviveWithIntactChain(t, sqlDB, tenant.ID, 2)
	assertTenantKeyDestroyed(t, sqlDB, tenant.ID)
	assertTenantTombstoned(t, sqlDB, tenant.ID)
	assertGovernanceEventsCount(t, sqlDB, tenant.ID, 0)
	assertAuthIdentitiesErased(t, sqlDB, formerUserIDs)
	assertIAMUsersCount(t, sqlDB, tenant.ID, 0)
}

// TestTenantErasure_DoesNotTouchOtherTenant seeds two tenants (A, B) with
// parallel data spreads, erases A only, and asserts B is fully intact
// (iam_users count unchanged, audit_events count unchanged, tenants row
// untouched, tenant_keys not destroyed) while A is erased per the same
// spot-check assertions as TestTenantErasure_ChainStaysGreen (iam_users=0 +
// tombstone).
func TestTenantErasure_DoesNotTouchOtherTenant(t *testing.T) {
	sqlDB := openDB(t)
	svc, cryptoSvc := newRunSideTenantLifecycleService(t, sqlDB)

	tenantA, _, adminA := seedTenantWithSpread(t, sqlDB, cryptoSvc, "erase-cross-a")
	tenantB, _, adminB := seedTenantWithSpread(t, sqlDB, cryptoSvc, "erase-cross-b")

	iamUsersBeforeB := countIAMUsers(t, sqlDB, tenantB.ID)
	auditEventsBeforeB := countAuditEvents(t, sqlDB, tenantB.ID)
	if iamUsersBeforeB == 0 || auditEventsBeforeB == 0 {
		t.Fatalf("tenant B seed spread looks empty before erasure: iam_users=%d audit_events=%d", iamUsersBeforeB, auditEventsBeforeB)
	}
	_ = adminB

	jobID := insertPendingLifecycleJob(t, sqlDB, tenantA.ID, "erase", adminA)
	if err := svc.RunJob(context.Background(), jobID); err != nil {
		t.Fatalf("RunJob(erase tenant A) error = %v", err)
	}

	// Tenant A: erased (spot-check per test 2's assertions).
	assertIAMUsersCount(t, sqlDB, tenantA.ID, 0)
	assertTenantTombstoned(t, sqlDB, tenantA.ID)

	// Tenant B: fully intact.
	iamUsersAfterB := countIAMUsers(t, sqlDB, tenantB.ID)
	if iamUsersAfterB != iamUsersBeforeB {
		t.Errorf("tenant B iam_users count changed: before=%d after=%d, want unchanged", iamUsersBeforeB, iamUsersAfterB)
	}
	auditEventsAfterB := countAuditEvents(t, sqlDB, tenantB.ID)
	if auditEventsAfterB != auditEventsBeforeB {
		t.Errorf("tenant B audit_events count changed: before=%d after=%d, want unchanged", auditEventsBeforeB, auditEventsAfterB)
	}

	var nameB, slugB string
	var erasedAtB sql.NullTime
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT name, slug, erased_at FROM metaldocs.tenants WHERE id = $1::uuid`, tenantB.ID,
	).Scan(&nameB, &slugB, &erasedAtB); err != nil {
		t.Fatalf("read tenant B tenants row: %v", err)
	}
	if erasedAtB.Valid {
		t.Errorf("tenant B tenants.erased_at is set, want NULL (untouched)")
	}
	if nameB == "erased" {
		t.Errorf("tenant B tenants.name = %q, want untouched (not 'erased')", nameB)
	}

	var destroyedAtB sql.NullTime
	var wrappedB []byte
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT wrapped_dek, destroyed_at FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenantB.ID,
	).Scan(&wrappedB, &destroyedAtB); err != nil {
		t.Fatalf("read tenant B tenant_keys row: %v", err)
	}
	if destroyedAtB.Valid {
		t.Errorf("tenant B tenant_keys.destroyed_at is set, want NULL (not destroyed)")
	}
	if len(wrappedB) == 0 {
		t.Errorf("tenant B tenant_keys.wrapped_dek is empty, want intact (not shredded)")
	}
}

// ── Shared assertion/count helpers ──────────────────────────────────────────

func assertJobReady(t *testing.T, sqlDB *sql.DB, jobID string) {
	t.Helper()
	var status string
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT status FROM metaldocs.tenant_lifecycle_jobs WHERE id = $1::uuid`, jobID,
	).Scan(&status); err != nil {
		t.Fatalf("read tenant_lifecycle_jobs status: %v", err)
	}
	if status != "ready" {
		t.Fatalf("tenant_lifecycle_jobs.status = %q, want %q", status, "ready")
	}
}

// assertAuditEventsSurviveWithIntactChain asserts at least wantMin
// audit_events rows for tenantID still exist (immutable skeleton — erasure
// never deletes audit_events) and that the hash chain is intact across the
// tenant's own sequence slice: prev_hash of row n == row_hash of row n-1,
// ordered by audit_sequence.
func assertAuditEventsSurviveWithIntactChain(t *testing.T, sqlDB *sql.DB, tenantID string, wantMin int) {
	t.Helper()
	rows, err := sqlDB.QueryContext(context.Background(),
		// audit_events.tenant_id is TEXT (§0.1), not uuid — no cast.
		`SELECT audit_sequence, prev_hash, row_hash
		   FROM metaldocs.audit_events
		  WHERE tenant_id = $1
		  ORDER BY audit_sequence ASC`, tenantID,
	)
	if err != nil {
		t.Fatalf("query audit_events for tenant %s: %v", tenantID, err)
	}
	defer rows.Close()

	type chainRow struct {
		sequence int64
		prev     string
		row      string
	}
	var chain []chainRow
	for rows.Next() {
		var r chainRow
		if err := rows.Scan(&r.sequence, &r.prev, &r.row); err != nil {
			t.Fatalf("scan audit_events chain row: %v", err)
		}
		chain = append(chain, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate audit_events chain rows: %v", err)
	}

	if len(chain) < wantMin {
		t.Fatalf("audit_events rows surviving for tenant %s = %d, want >= %d (erasure must NOT delete audit_events)", tenantID, len(chain), wantMin)
	}

	for i := 1; i < len(chain); i++ {
		if chain[i].prev != chain[i-1].row {
			t.Errorf("audit hash chain broken at tenant %s sequence %d: prev_hash=%q, want previous row's row_hash=%q",
				tenantID, chain[i].sequence, chain[i].prev, chain[i-1].row)
		}
	}
}

func assertTenantKeyDestroyed(t *testing.T, sqlDB *sql.DB, tenantID string) {
	t.Helper()
	var wrapped []byte
	var destroyedAt sql.NullTime
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT wrapped_dek, destroyed_at FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenantID,
	).Scan(&wrapped, &destroyedAt); err != nil {
		t.Fatalf("read tenant_keys row for %s: %v", tenantID, err)
	}
	if len(wrapped) != 0 {
		t.Errorf("tenant_keys.wrapped_dek for %s is non-empty after erasure, want zeroed", tenantID)
	}
	if !destroyedAt.Valid {
		t.Errorf("tenant_keys.destroyed_at for %s is NULL after erasure, want set", tenantID)
	}
}

func assertTenantTombstoned(t *testing.T, sqlDB *sql.DB, tenantID string) {
	t.Helper()
	var name, slug string
	var erasedAt sql.NullTime
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT name, slug, erased_at FROM metaldocs.tenants WHERE id = $1::uuid`, tenantID,
	).Scan(&name, &slug, &erasedAt); err != nil {
		t.Fatalf("read tenants row for %s: %v", tenantID, err)
	}
	if name != "erased" {
		t.Errorf("tenants.name for %s = %q, want %q", tenantID, name, "erased")
	}
	if !erasedAt.Valid {
		t.Errorf("tenants.erased_at for %s is NULL, want set", tenantID)
	}
}

func assertGovernanceEventsCount(t *testing.T, sqlDB *sql.DB, tenantID string, want int) {
	t.Helper()
	var n int
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.governance_events WHERE tenant_id = $1::uuid`, tenantID,
	).Scan(&n); err != nil {
		t.Fatalf("count governance_events for %s: %v", tenantID, err)
	}
	if n != want {
		t.Errorf("governance_events count for %s = %d, want %d", tenantID, n, want)
	}
}

// assertAuthIdentitiesErased is the F7.3 Defect 1 regression guard: proves
// metaldocs.auth_identities holds ZERO rows for any of the erased tenant's
// FORMER user ids (captured before erasure ran, since iam_users itself is
// erased by the iam port and can no longer resolve the join afterward).
func assertAuthIdentitiesErased(t *testing.T, sqlDB *sql.DB, formerUserIDs []string) {
	t.Helper()
	if len(formerUserIDs) == 0 {
		t.Fatalf("assertAuthIdentitiesErased: no former user ids captured — test setup bug")
	}
	placeholders := make([]string, len(formerUserIDs))
	args := make([]any, len(formerUserIDs))
	for i, id := range formerUserIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := "SELECT count(*) FROM metaldocs.auth_identities WHERE user_id IN (" + strings.Join(placeholders, ",") + ")"
	var n int
	if err := sqlDB.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count auth_identities for former user ids %v: %v", formerUserIDs, err)
	}
	if n != 0 {
		t.Errorf("auth_identities rows surviving for erased tenant's former user ids %v = %d, want 0", formerUserIDs, n)
	}
}

func assertIAMUsersCount(t *testing.T, sqlDB *sql.DB, tenantID string, want int) {
	t.Helper()
	n := countIAMUsers(t, sqlDB, tenantID)
	if n != want {
		t.Errorf("iam_users count for %s = %d, want %d", tenantID, n, want)
	}
}

func countIAMUsers(t *testing.T, sqlDB *sql.DB, tenantID string) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.iam_users WHERE tenant_id = $1::uuid`, tenantID,
	).Scan(&n); err != nil {
		t.Fatalf("count iam_users for %s: %v", tenantID, err)
	}
	return n
}

func countAuditEvents(t *testing.T, sqlDB *sql.DB, tenantID string) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.audit_events WHERE tenant_id = $1`, tenantID,
	).Scan(&n); err != nil {
		t.Fatalf("count audit_events for %s: %v", tenantID, err)
	}
	return n
}

// containsJSONKey/containsJSONValue do a minimal, dependency-free scan of a
// row_to_json-produced JSON object for a top-level key (absence) or a
// top-level string field's exact value (presence), avoiding a full
// encoding/json round-trip per row across two lightweight helpers — accurate
// enough for this file's assertions since every export row here is a flat
// row_to_json object with no nested objects sharing these exact keys.
func containsJSONKey(t *testing.T, raw []byte, key string) bool {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal export row JSON: %v (row=%s)", err, string(raw))
	}
	_, ok := m[key]
	return ok
}

func containsJSONValue(t *testing.T, raw []byte, key, want string) bool {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal export row JSON: %v (row=%s)", err, string(raw))
	}
	v, ok := m[key]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s == want
}
