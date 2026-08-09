package application_test

// tenant_crypto_service_test.go — M7 F7.3 Task B TDD (RED-first): unit tests
// for TenantCryptoService against a fake tenantKeyRepository (no DB). The
// Postgres adapter itself is exercised by a separate integration test using
// the testdb factory (tests/integration/security).
//
//   go test ./internal/modules/security/application/...

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	securitydomain "metaldocs/internal/modules/security/domain"

	"metaldocs/internal/modules/security/application"
	"metaldocs/internal/platform/db"
)

// fakeTenantKeyRepository is an in-memory stand-in for the Postgres adapter,
// letting the service's business logic (idempotent provisioning, key-state
// error mapping) be proven without a database.
type fakeTenantKeyRepository struct {
	rows map[string]*fakeKeyRow

	// txRows, when non-nil, models a store visible ONLY via WrappedDEKTx —
	// i.e. a row inserted in an open tx that the pool-backed WrappedDEK
	// cannot see yet. Left nil by default so existing tests (which never
	// distinguish tx/pool visibility) are unaffected.
	txRows map[string]*fakeKeyRow

	insertErr error
}

type fakeKeyRow struct {
	wrappedDEK  []byte
	destroyedAt bool
}

func newFakeRepo() *fakeTenantKeyRepository {
	return &fakeTenantKeyRepository{rows: map[string]*fakeKeyRow{}}
}

func (f *fakeTenantKeyRepository) InsertIfAbsentTx(ctx context.Context, tx db.Tx, tenantID string, wrappedDEK []byte) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	if _, ok := f.rows[tenantID]; ok {
		return nil // idempotent: already provisioned
	}
	f.rows[tenantID] = &fakeKeyRow{wrappedDEK: wrappedDEK}
	return nil
}

func (f *fakeTenantKeyRepository) WrappedDEK(ctx context.Context, tenantID string) ([]byte, bool, error) {
	row, ok := f.rows[tenantID]
	if !ok {
		return nil, false, nil
	}
	if row.destroyedAt {
		return nil, true, nil
	}
	return row.wrappedDEK, false, nil
}

// WrappedDEKTx models the tx-visible store separately from the pool-visible
// one (txRows), so tests can prove EncryptForTenantTx actually reads via the
// tx path rather than silently delegating to the pool-backed WrappedDEK.
// When txRows is nil, it mirrors WrappedDEK exactly (no tx/pool visibility
// difference modeled) — sufficient for tests that only care that the Tx
// variant delegates correctly.
func (f *fakeTenantKeyRepository) WrappedDEKTx(ctx context.Context, tx db.Tx, tenantID string) ([]byte, bool, error) {
	if f.txRows != nil {
		row, ok := f.txRows[tenantID]
		if !ok {
			return nil, false, nil
		}
		if row.destroyedAt {
			return nil, true, nil
		}
		return row.wrappedDEK, false, nil
	}
	return f.WrappedDEK(ctx, tenantID)
}

func (f *fakeTenantKeyRepository) DestroyTx(ctx context.Context, tx db.Tx, tenantID string) error {
	row, ok := f.rows[tenantID]
	if !ok {
		return nil // idempotent no-op: nothing to destroy
	}
	row.destroyedAt = true
	row.wrappedDEK = []byte{}
	return nil
}

func mustKEK(t *testing.T) []byte {
	t.Helper()
	kek := make([]byte, 32)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("rand.Read kek: %v", err)
	}
	return kek
}

func TestNewTenantCryptoService_RejectsBadKEKLength(t *testing.T) {
	if _, err := application.NewTenantCryptoService(newFakeRepo(), []byte("too-short")); err == nil {
		t.Fatalf("NewTenantCryptoService() with a non-32-byte KEK = nil error, want failure")
	}
}

func TestTenantCryptoService_ProvisionThenEncryptDecrypt_RoundTrip(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()
	tenantID := "tenant-a"

	if err := svc.ProvisionTenantKeyTx(ctx, nil, tenantID); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	envelope, err := svc.EncryptForTenant(ctx, tenantID, []byte("classified audit payload"))
	if err != nil {
		t.Fatalf("EncryptForTenant() error = %v", err)
	}
	if envelope == "" {
		t.Fatalf("EncryptForTenant() returned empty envelope")
	}

	plaintext, err := svc.DecryptForTenant(ctx, tenantID, envelope)
	if err != nil {
		t.Fatalf("DecryptForTenant() error = %v", err)
	}
	if string(plaintext) != "classified audit payload" {
		t.Fatalf("DecryptForTenant() = %q, want %q", plaintext, "classified audit payload")
	}
}

func TestTenantCryptoService_ProvisionTenantKeyTx_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()
	tenantID := "tenant-b"

	if err := svc.ProvisionTenantKeyTx(ctx, nil, tenantID); err != nil {
		t.Fatalf("first ProvisionTenantKeyTx() error = %v", err)
	}
	envelope, err := svc.EncryptForTenant(ctx, tenantID, []byte("payload one"))
	if err != nil {
		t.Fatalf("EncryptForTenant() error = %v", err)
	}

	// Re-provisioning must NOT rotate the DEK — the earlier envelope must
	// still decrypt after a second ProvisionTenantKeyTx call.
	if err := svc.ProvisionTenantKeyTx(ctx, nil, tenantID); err != nil {
		t.Fatalf("second ProvisionTenantKeyTx() error = %v", err)
	}
	plaintext, err := svc.DecryptForTenant(ctx, tenantID, envelope)
	if err != nil {
		t.Fatalf("DecryptForTenant() after re-provision error = %v, want nil (DEK must not rotate)", err)
	}
	if string(plaintext) != "payload one" {
		t.Fatalf("DecryptForTenant() = %q, want %q", plaintext, "payload one")
	}
}

func TestTenantCryptoService_EncryptForTenant_NoKeyReturnsErrKeyNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()

	if _, err := svc.EncryptForTenant(ctx, "no-such-tenant", []byte("x")); !errors.Is(err, securitydomain.ErrKeyNotFound) {
		t.Fatalf("EncryptForTenant() error = %v, want ErrKeyNotFound", err)
	}
}

func TestTenantCryptoService_DecryptForTenant_NoKeyReturnsErrKeyNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()

	if _, err := svc.DecryptForTenant(ctx, "no-such-tenant", `{"enc":"aesgcm.v1","data":"AAAA"}`); !errors.Is(err, securitydomain.ErrKeyNotFound) {
		t.Fatalf("DecryptForTenant() error = %v, want ErrKeyNotFound", err)
	}
}

// TestTenantCryptoService_EncryptForTenantTx_ReadsThroughTxOnly proves the
// tx-aware seal-chain fix (F7.3 defect: tenant.onboarded audit payload landed
// plaintext because the same-tx key row, still uncommitted, was invisible to
// a pool read): a key visible ONLY via the fake's txRows (never via the
// pool-backed WrappedDEK) must still be resolvable by EncryptForTenantTx when
// a non-nil tx is passed, and the pool-only EncryptForTenant call for the
// same tenant must still fail with ErrKeyNotFound — proving the Tx variant
// actually delegates to WrappedDEKTx rather than silently falling back to
// the pool path.
func TestTenantCryptoService_EncryptForTenantTx_ReadsThroughTxOnly(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()
	tenantID := "tenant-tx-only"

	// Provision normally (writes into repo.rows, the pool-visible store),
	// then relocate the row into txRows and clear rows — this models a
	// key visible ONLY through the tx path (mirrors an
	// INSERT ... tenant_keys row that hasn't committed yet, so a pool
	// read cannot see it but a same-tx read can).
	if err := svc.ProvisionTenantKeyTx(ctx, nil, tenantID); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}
	repo.txRows = map[string]*fakeKeyRow{tenantID: repo.rows[tenantID]}
	delete(repo.rows, tenantID)

	// Pool-only read must still miss: the row is only in txRows.
	if _, err := svc.EncryptForTenant(ctx, tenantID, []byte("x")); !errors.Is(err, securitydomain.ErrKeyNotFound) {
		t.Fatalf("EncryptForTenant() (pool path) error = %v, want ErrKeyNotFound (proves txRows is tx-only)", err)
	}

	// A non-nil tx is required to take the tx-read path (the fake never
	// inspects tx's contents, but resolveDEKTx branches on tx == nil, so a
	// real, non-nil *sql.Tx is needed here — a sqlmock-backed one, since
	// the fake repo never issues SQL against it).
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer func() { _ = mockDB.Close() }()
	mock.ExpectBegin()
	tx, err := mockDB.Begin()
	if err != nil {
		t.Fatalf("mockDB.Begin() error = %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	envelope, err := svc.EncryptForTenantTx(ctx, tx, tenantID, []byte("classified payload"))
	if err != nil {
		t.Fatalf("EncryptForTenantTx() error = %v, want success reading through the tx-only key", err)
	}
	if envelope == "" {
		t.Fatalf("EncryptForTenantTx() returned empty envelope")
	}

	// Simulate the tx committing: the row becomes pool-visible again, so the
	// round-trip decrypt below (pool path — there is no DecryptForTenantTx,
	// matching the spec's Encrypt-only tx variant) can resolve it.
	repo.rows[tenantID] = repo.txRows[tenantID]

	plaintext, err := svc.DecryptForTenant(ctx, tenantID, envelope)
	if err != nil {
		t.Fatalf("DecryptForTenant() error = %v", err)
	}
	if string(plaintext) != "classified payload" {
		t.Fatalf("DecryptForTenant() = %q, want %q", plaintext, "classified payload")
	}
}

func TestTenantCryptoService_DestroyTenantKeyTx_ShredsDecryption(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()
	tenantID := "tenant-c"

	if err := svc.ProvisionTenantKeyTx(ctx, nil, tenantID); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}
	envelope, err := svc.EncryptForTenant(ctx, tenantID, []byte("shred me"))
	if err != nil {
		t.Fatalf("EncryptForTenant() error = %v", err)
	}

	if err := svc.DestroyTenantKeyTx(ctx, nil, tenantID); err != nil {
		t.Fatalf("DestroyTenantKeyTx() error = %v", err)
	}

	if _, err := svc.DecryptForTenant(ctx, tenantID, envelope); !errors.Is(err, securitydomain.ErrKeyDestroyed) {
		t.Fatalf("DecryptForTenant() after shred error = %v, want ErrKeyDestroyed", err)
	}
	if _, err := svc.EncryptForTenant(ctx, tenantID, []byte("new data")); !errors.Is(err, securitydomain.ErrKeyDestroyed) {
		t.Fatalf("EncryptForTenant() after shred error = %v, want ErrKeyDestroyed", err)
	}
}

func TestTenantCryptoService_DestroyTenantKeyTx_IdempotentOnMissingKey(t *testing.T) {
	repo := newFakeRepo()
	svc, err := application.NewTenantCryptoService(repo, mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	ctx := context.Background()

	if err := svc.DestroyTenantKeyTx(ctx, nil, "never-provisioned"); err != nil {
		t.Fatalf("DestroyTenantKeyTx() on missing key error = %v, want nil (idempotent)", err)
	}
}
