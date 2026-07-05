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
	"database/sql"
	"errors"
	"testing"

	securitydomain "metaldocs/internal/modules/security/domain"

	"metaldocs/internal/modules/security/application"
)

// fakeTenantKeyRepository is an in-memory stand-in for the Postgres adapter,
// letting the service's business logic (idempotent provisioning, key-state
// error mapping) be proven without a database.
type fakeTenantKeyRepository struct {
	rows map[string]*fakeKeyRow

	insertErr error
}

type fakeKeyRow struct {
	wrappedDEK  []byte
	destroyedAt bool
}

func newFakeRepo() *fakeTenantKeyRepository {
	return &fakeTenantKeyRepository{rows: map[string]*fakeKeyRow{}}
}

func (f *fakeTenantKeyRepository) InsertIfAbsentTx(ctx context.Context, tx *sql.Tx, tenantID string, wrappedDEK []byte) error {
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

func (f *fakeTenantKeyRepository) DestroyTx(ctx context.Context, tx *sql.Tx, tenantID string) error {
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
