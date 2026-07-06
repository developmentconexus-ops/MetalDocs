package domain

import (
	"context"
	"database/sql"
	"errors"
)

// TenantCrypto is the published port for per-tenant envelope encryption (M7
// F7.3, ADR 0070 decision 4). It backs the crypto-shred erasure model: each
// tenant gets exactly one data-encryption key (DEK), wrapped under the
// deployment KEK and stored in metaldocs.tenant_keys. Destroying that row's
// key material (DestroyTenantKeyTx) makes every envelope ciphertext sealed
// under it permanently unrecoverable — no per-row rewrite needed.
//
// Other modules depend on this interface, never on the security module's
// infrastructure/postgres adapter or its SQL. Composition roots wire the
// concrete *PostgresTenantCrypto (or a nil/no-op path when
// METALDOCS_TENANT_KEK is unset, mirroring F7.2's TenantKeyProvisioner
// no-op seam).
type TenantCrypto interface {
	// ProvisionTenantKeyTx generates a fresh 32-byte DEK, wraps it under the
	// service's KEK, and inserts the metaldocs.tenant_keys row for tenantID
	// inside the caller's tx. Idempotent: calling it again for a tenant that
	// already has a key is a no-op (does not rotate or overwrite).
	ProvisionTenantKeyTx(ctx context.Context, tx *sql.Tx, tenantID string) error

	// EncryptForTenant seals plaintext under tenantID's active DEK and
	// returns the envelope JSON string ({"enc":"aesgcm.v1","data":"..."}).
	// Returns ErrKeyNotFound if the tenant has no key row, ErrKeyDestroyed
	// if the tenant's key has been crypto-shredded.
	EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (string, error)

	// EncryptForTenantTx is EncryptForTenant's tx-aware variant: when tx is
	// non-nil, the DEK is resolved by reading tenant_keys through tx rather
	// than the pool. This closes a same-transaction visibility gap: a
	// caller that provisions a tenant's key (ProvisionTenantKeyTx) and then
	// seals a payload for that same tenant in the SAME still-open
	// transaction (e.g. OnboardTenantService sealing its tenant.onboarded
	// audit event) would otherwise have its pool-backed EncryptForTenant
	// read race the uncommitted insert and always miss it, silently
	// falling through to plaintext storage. Falls back to the ordinary
	// pool-backed resolution when tx is nil (byte-identical to
	// EncryptForTenant).
	EncryptForTenantTx(ctx context.Context, tx *sql.Tx, tenantID string, plaintext []byte) (string, error)

	// DecryptForTenant reverses EncryptForTenant. Returns ErrKeyNotFound if
	// the tenant has no key row, ErrKeyDestroyed if the tenant's key has
	// been crypto-shredded (the intended, permanent failure mode of
	// erasure — this is not a bug to work around).
	DecryptForTenant(ctx context.Context, tenantID string, envelope string) ([]byte, error)

	// DestroyTenantKeyTx crypto-shreds tenantID's key inside the caller's
	// tx: sets destroyed_at = now() and zeroes wrapped_dek. Idempotent
	// (destroying an already-destroyed or missing key is a no-op, not an
	// error) so erasure retries are safe.
	DestroyTenantKeyTx(ctx context.Context, tx *sql.Tx, tenantID string) error
}

// ErrKeyNotFound is returned when a tenant has no metaldocs.tenant_keys row
// at all (never provisioned).
var ErrKeyNotFound = errors.New("security: tenant key not found")

// ErrKeyDestroyed is returned when a tenant's key has been crypto-shredded
// (destroyed_at is set) — the designed, permanent failure mode for
// decrypting post-erasure ciphertext.
var ErrKeyDestroyed = errors.New("security: tenant key destroyed")
