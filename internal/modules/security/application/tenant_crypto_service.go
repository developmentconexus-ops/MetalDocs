package application

// tenant_crypto_service.go — TenantCryptoService (M7 F7.3 Task B, ADR 0070
// decision 4) implements securitydomain.TenantCrypto: per-tenant DEK
// provisioning, envelope encrypt/decrypt, and crypto-shred destruction.
//
// Key material handling: the KEK is held only in memory (never logged,
// never printed, never included in any error string). DEKs are generated
// with crypto/rand and immediately wrapped under the KEK before ever
// leaving this service — the unwrapped DEK exists only for the duration of
// a single Encrypt/Decrypt call's stack frame.

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	platcrypto "metaldocs/internal/platform/crypto"
	"metaldocs/internal/platform/db"

	securitydomain "metaldocs/internal/modules/security/domain"
)

// tenantKeyRepository is the persistence port TenantCryptoService depends on
// — implemented by the security module's Postgres adapter
// (infrastructure/postgres/tenant_key_repository.go) and, for unit tests,
// an in-memory fake. Every method takes tenantID and is expected to carry an
// explicit tenant_id predicate in its SQL (no reliance on RLS GUCs for
// read paths off a shared pool).
type tenantKeyRepository interface {
	// InsertIfAbsentTx inserts the tenant_keys row for tenantID with the
	// given already-wrapped DEK, inside tx, UNLESS a row already exists —
	// in which case it is a no-op (idempotent; never rotates an existing
	// key).
	InsertIfAbsentTx(ctx context.Context, tx db.Tx, tenantID string, wrappedDEK []byte) error

	// WrappedDEK returns the tenant's wrapped DEK bytes and whether the key
	// has been destroyed. found=false means no row exists at all.
	WrappedDEK(ctx context.Context, tenantID string) (wrappedDEK []byte, destroyed bool, err error)

	// WrappedDEKTx is WrappedDEK's tx-scoped variant: it reads through tx
	// rather than the pool, so a caller that is inserting the tenant_keys
	// row and reading it back in the same still-open transaction (e.g. the
	// audit writer sealing a tenant.onboarded event in the same tx as
	// ProvisionTenantKeyTx) actually sees the just-inserted row instead of
	// racing an uncommitted write via the pool (which always misses,
	// yielding a false ErrKeyNotFound).
	WrappedDEKTx(ctx context.Context, tx db.Tx, tenantID string) (wrappedDEK []byte, destroyed bool, err error)

	// DestroyTx crypto-shreds tenantID's key inside tx: sets
	// destroyed_at = now() and zeroes wrapped_dek. No-op (not an error) if
	// the tenant has no key row, or the key is already destroyed.
	DestroyTx(ctx context.Context, tx db.Tx, tenantID string) error
}

// TenantCryptoService implements securitydomain.TenantCrypto.
type TenantCryptoService struct {
	repo tenantKeyRepository
	kek  []byte
}

var _ securitydomain.TenantCrypto = (*TenantCryptoService)(nil)

// NewTenantCryptoService constructs the service. kek must be exactly
// platcrypto.KeySize (32) bytes — callers derive it from
// METALDOCS_TENANT_KEK (base64-decoded) at the composition root. Returns an
// error (never panics) on a malformed KEK so boot can log-and-degrade to the
// nil/no-op path rather than crash.
func NewTenantCryptoService(repo tenantKeyRepository, kek []byte) (*TenantCryptoService, error) {
	if len(kek) != platcrypto.KeySize {
		return nil, fmt.Errorf("security: KEK must be %d bytes, got %d", platcrypto.KeySize, len(kek))
	}
	if repo == nil {
		return nil, errors.New("security: TenantCryptoService requires a non-nil repository")
	}
	kekCopy := make([]byte, len(kek))
	copy(kekCopy, kek)
	return &TenantCryptoService{repo: repo, kek: kekCopy}, nil
}

// ProvisionTenantKeyTx generates a fresh 32-byte DEK, wraps it under the
// service's KEK, and inserts the tenant_keys row inside tx. Idempotent.
func (s *TenantCryptoService) ProvisionTenantKeyTx(ctx context.Context, tx db.Tx, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return errors.New("security: tenant id required")
	}

	dek, err := generateDEK()
	if err != nil {
		return fmt.Errorf("security: generate tenant DEK: %w", err)
	}
	wrapped, err := platcrypto.WrapDEK(s.kek, dek)
	if err != nil {
		return fmt.Errorf("security: wrap tenant DEK: %w", err)
	}
	zeroBytes(dek)

	if err := s.repo.InsertIfAbsentTx(ctx, tx, tenantID, wrapped); err != nil {
		return fmt.Errorf("security: provision tenant key: %w", err)
	}
	return nil
}

// EncryptForTenant seals plaintext under tenantID's active DEK.
func (s *TenantCryptoService) EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (string, error) {
	dek, err := s.resolveDEK(ctx, tenantID)
	if err != nil {
		return "", err
	}
	defer zeroBytes(dek)

	sealed, err := platcrypto.Seal(dek, plaintext)
	if err != nil {
		return "", fmt.Errorf("security: encrypt for tenant: %w", err)
	}
	return platcrypto.EncodeEnvelope(sealed), nil
}

// EncryptForTenantTx is EncryptForTenant's tx-aware variant: when tx is
// non-nil, the DEK is resolved via WrappedDEKTx (a read through tx) instead
// of the pool, so a caller sealing a payload in the SAME transaction that
// just provisioned the tenant's key (e.g. OnboardTenantService's
// tenant.onboarded audit event, written in the same tx as
// ProvisionTenantKeyTx) sees the just-inserted, still-uncommitted key row
// rather than racing a pool read that always misses it. When tx is nil, this
// falls back to the ordinary pool-backed resolution (byte-identical to
// EncryptForTenant).
func (s *TenantCryptoService) EncryptForTenantTx(ctx context.Context, tx db.Tx, tenantID string, plaintext []byte) (string, error) {
	dek, err := s.resolveDEKTx(ctx, tx, tenantID)
	if err != nil {
		return "", err
	}
	defer zeroBytes(dek)

	sealed, err := platcrypto.Seal(dek, plaintext)
	if err != nil {
		return "", fmt.Errorf("security: encrypt for tenant: %w", err)
	}
	return platcrypto.EncodeEnvelope(sealed), nil
}

// DecryptForTenant reverses EncryptForTenant.
func (s *TenantCryptoService) DecryptForTenant(ctx context.Context, tenantID string, envelope string) ([]byte, error) {
	dek, err := s.resolveDEK(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer zeroBytes(dek)

	sealed, err := platcrypto.DecodeEnvelope(envelope)
	if err != nil {
		return nil, fmt.Errorf("security: decrypt for tenant: %w", err)
	}
	plaintext, err := platcrypto.Open(dek, sealed)
	if err != nil {
		return nil, fmt.Errorf("security: decrypt for tenant: %w", err)
	}
	return plaintext, nil
}

// DestroyTenantKeyTx crypto-shreds tenantID's key inside tx.
func (s *TenantCryptoService) DestroyTenantKeyTx(ctx context.Context, tx db.Tx, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return errors.New("security: tenant id required")
	}
	if err := s.repo.DestroyTx(ctx, tx, tenantID); err != nil {
		return fmt.Errorf("security: destroy tenant key: %w", err)
	}
	return nil
}

// resolveDEK fetches and unwraps tenantID's DEK, mapping repository
// key-state to the published typed errors.
func (s *TenantCryptoService) resolveDEK(ctx context.Context, tenantID string) ([]byte, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, errors.New("security: tenant id required")
	}
	wrapped, destroyed, err := s.repo.WrappedDEK(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("security: resolve tenant key: %w", err)
	}
	return s.unwrapAndCheck(wrapped, destroyed)
}

// resolveDEKTx is resolveDEK's tx-aware variant: when tx is non-nil, the
// wrapped DEK is read via WrappedDEKTx (through tx) instead of the pool.
func (s *TenantCryptoService) resolveDEKTx(ctx context.Context, tx db.Tx, tenantID string) ([]byte, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, errors.New("security: tenant id required")
	}
	if tx == nil {
		wrapped, destroyed, err := s.repo.WrappedDEK(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("security: resolve tenant key: %w", err)
		}
		return s.unwrapAndCheck(wrapped, destroyed)
	}
	wrapped, destroyed, err := s.repo.WrappedDEKTx(ctx, tx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("security: resolve tenant key: %w", err)
	}
	return s.unwrapAndCheck(wrapped, destroyed)
}

// unwrapAndCheck maps repository key-state to the published typed errors and
// unwraps the DEK on success. Shared by resolveDEK and resolveDEKTx.
func (s *TenantCryptoService) unwrapAndCheck(wrapped []byte, destroyed bool) ([]byte, error) {
	if destroyed {
		return nil, securitydomain.ErrKeyDestroyed
	}
	if wrapped == nil {
		return nil, securitydomain.ErrKeyNotFound
	}
	dek, err := platcrypto.UnwrapDEK(s.kek, wrapped)
	if err != nil {
		return nil, fmt.Errorf("security: unwrap tenant key: %w", err)
	}
	return dek, nil
}

// generateDEK returns a fresh, random 32-byte data-encryption key.
func generateDEK() ([]byte, error) {
	dek := make([]byte, platcrypto.KeySize)
	if _, err := rand.Read(dek); err != nil {
		return nil, err
	}
	return dek, nil
}

// zeroBytes best-effort scrubs key material from memory once no longer
// needed. Not a cryptographic guarantee (Go's GC/compiler may retain copies)
// but is the established minimum-diligence idiom for this codebase's key
// handling.
func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
