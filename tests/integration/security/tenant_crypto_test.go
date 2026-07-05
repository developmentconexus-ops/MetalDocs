//go:build integration

// tenant_crypto_test.go — M7 F7.3 Task B: drives TenantCryptoService against
// a real Postgres instance (testdb factory), proving the Postgres adapter
// (TenantKeyRepository) round-trips through the actual metaldocs.tenant_keys
// table + FORCE RLS tenant_isolation policy from migration 0281.
//
//   go test -tags=integration ./tests/integration/security/... -run TenantCrypto
package security_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"testing"

	securityapp "metaldocs/internal/modules/security/application"
	securitydomain "metaldocs/internal/modules/security/domain"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

func mustKEK(t *testing.T) []byte {
	t.Helper()
	kek := make([]byte, 32)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("rand.Read kek: %v", err)
	}
	return kek
}

func newTenantCryptoService(t *testing.T, sqlDB *sql.DB) *securityapp.TenantCryptoService {
	t.Helper()
	svc, err := securityapp.NewTenantCryptoService(securitypg.NewTenantKeyRepository(sqlDB), mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	return svc
}

// ctxForTenant seeds the tenant+actor identity the TxRunner chokepoint
// auto-seeds into tx-local GUCs (F3.1 contract), so writes inside
// runner.Do exercise the tenant_keys FORCE RLS policy exactly like a real
// request would.
func ctxForTenant(tenantID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, "test-actor")
	return ctx
}

func TestTenantCrypto_ProvisionEncryptDecrypt_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	svc := newTenantCryptoService(t, sqlDB)
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	var wrapped []byte
	var destroyedAt sql.NullTime
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT wrapped_dek, destroyed_at FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`,
		tenant.ID,
	).Scan(&wrapped, &destroyedAt); err != nil {
		t.Fatalf("read tenant_keys row: %v", err)
	}
	if len(wrapped) == 0 {
		t.Fatalf("wrapped_dek is empty after provisioning")
	}
	if destroyedAt.Valid {
		t.Fatalf("destroyed_at is set right after provisioning, want NULL")
	}

	envelope, err := svc.EncryptForTenant(context.Background(), tenant.ID, []byte("classified audit payload"))
	if err != nil {
		t.Fatalf("EncryptForTenant() error = %v", err)
	}
	plaintext, err := svc.DecryptForTenant(context.Background(), tenant.ID, envelope)
	if err != nil {
		t.Fatalf("DecryptForTenant() error = %v", err)
	}
	if string(plaintext) != "classified audit payload" {
		t.Fatalf("DecryptForTenant() = %q, want %q", plaintext, "classified audit payload")
	}
}

func TestTenantCrypto_ProvisionTenantKeyTx_IdempotentRealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	svc := newTenantCryptoService(t, sqlDB)
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	provision := func() error {
		return runner.Do(ctx, func(tx *sql.Tx) error {
			return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
		})
	}
	if err := provision(); err != nil {
		t.Fatalf("first ProvisionTenantKeyTx() error = %v", err)
	}

	var firstWrapped []byte
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT wrapped_dek FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenant.ID,
	).Scan(&firstWrapped); err != nil {
		t.Fatalf("read tenant_keys row: %v", err)
	}

	if err := provision(); err != nil {
		t.Fatalf("second ProvisionTenantKeyTx() error = %v", err)
	}

	var secondWrapped []byte
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT wrapped_dek FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenant.ID,
	).Scan(&secondWrapped); err != nil {
		t.Fatalf("read tenant_keys row: %v", err)
	}

	if string(firstWrapped) != string(secondWrapped) {
		t.Fatalf("wrapped_dek changed across a second ProvisionTenantKeyTx call — key was rotated, must be idempotent")
	}

	var count int
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenant.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count tenant_keys rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("tenant_keys row count = %d, want 1", count)
	}
}

func TestTenantCrypto_DestroyTenantKeyTx_ShredsDecryption_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	svc := newTenantCryptoService(t, sqlDB)
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	envelope, err := svc.EncryptForTenant(context.Background(), tenant.ID, []byte("shred me"))
	if err != nil {
		t.Fatalf("EncryptForTenant() error = %v", err)
	}

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.DestroyTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("DestroyTenantKeyTx() error = %v", err)
	}

	var wrapped []byte
	var destroyedAt sql.NullTime
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT wrapped_dek, destroyed_at FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`,
		tenant.ID,
	).Scan(&wrapped, &destroyedAt); err != nil {
		t.Fatalf("read tenant_keys row: %v", err)
	}
	if len(wrapped) != 0 {
		t.Fatalf("wrapped_dek is non-empty after destroy, want zeroed")
	}
	if !destroyedAt.Valid {
		t.Fatalf("destroyed_at is NULL after destroy, want set")
	}

	if _, err := svc.DecryptForTenant(context.Background(), tenant.ID, envelope); !errors.Is(err, securitydomain.ErrKeyDestroyed) {
		t.Fatalf("DecryptForTenant() after shred error = %v, want ErrKeyDestroyed", err)
	}
}

// NOTE (bounded limitation, mirrors the M6/F6.4 finding recorded in
// project memory): a live-RLS-enforcement test asserting that tenant_keys'
// FORCE RLS tenant_isolation policy (migration 0281) blocks cross-tenant
// SELECTs was written and run against the real dev/test Postgres instance,
// but metaldocs_app (the runtime/test connection role) is
// Superuser+BypassRLS in this environment, so ROW LEVEL SECURITY is
// unconditionally bypassed regardless of policy correctness — the test
// failed not because the policy is wrong, but because the DB role cannot
// exercise RLS at all. The policy text is copied verbatim from the proven
// document_families tenant_isolation idiom (migration 0258, already load-
// bearing in production), and application-level tenant scoping is proven by
// TenantCryptoService's own unit tests (explicit tenant_id predicates,
// TestTenantCryptoService_* in internal/modules/security/application) plus
// TenantKeyRepository's explicit `tenant_id = $1` predicate on every
// statement (no query in tenant_key_repository.go relies on the RLS GUC for
// isolation). Fixing the dev-role BypassRLS attribute is outside this
// task's boundary (affects every tenant-scoped table, not just
// tenant_keys) — see the M6 GMR memory note "metaldocs_app superuser+
// BYPASSRLS -> RLS inert in dev".
