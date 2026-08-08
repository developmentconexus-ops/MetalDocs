# metaldocs.tenant_keys

**Schema:** `metaldocs.tenant_keys`
**Owner:** security module (M7 F7.3 Task B — crypto-shred tenant erasure)
**Last verified:** 2026-08-07

## Purpose

Holds the per-tenant wrapped Data Encryption Key (DEK), wrapped under the
deployment's Key Encryption Key (KEK). Destroying a tenant's row here
("crypto-shredding" — `destroyed_at` set, `wrapped_dek` zeroed) is what makes
any payload encrypted under that DEK permanently unrecoverable; it is the
mechanism GDPR-style tenant erasure relies on, distinct from row-level deletes
elsewhere. Reads run off the shared pool with an explicit `tenant_id`
predicate rather than depending on the RLS session GUC (M6 F6.4 idiom) — see
`WrappedDEK`/`WrappedDEKTx`.

## Columns

| Column        | Type          | Notes |
|---------------|---------------|-------|
| `tenant_id`   | `uuid` PK     | FK → `metaldocs.tenants(id)`. |
| `wrapped_dek` | `bytea`       | Not null. The tenant's DEK, wrapped under the deployment KEK. Zeroed (`''::bytea`) on crypto-shred. |
| `created_at`  | `timestamptz` | Default `now()`, not null. |
| `destroyed_at`| `timestamptz` | Nullable. Non-null means the key has been crypto-shredded; `WrappedDEK`/`WrappedDEKTx` report `destroyed=true` and withhold the (already-zeroed) key material. |

## Migrations

Table is present in `db/baseline/0001_current_schema.sql` (folded baseline); no
post-baseline migration alters it as of 2026-08-07.

## Key callers

- `internal/modules/security/infrastructure/postgres/tenant_key_repository.go::TenantKeyRepository.InsertIfAbsentTx` — provisions the key row in-tx, idempotent via `ON CONFLICT (tenant_id) DO NOTHING`.
- `internal/modules/security/infrastructure/postgres/tenant_key_repository.go::TenantKeyRepository.WrappedDEK` — pool read, explicit `tenant_id` predicate.
- `internal/modules/security/infrastructure/postgres/tenant_key_repository.go::TenantKeyRepository.WrappedDEKTx` — tx-scoped read so a caller can see its own uncommitted insert in the same transaction.
- `internal/modules/security/infrastructure/postgres/tenant_key_repository.go::TenantKeyRepository.DestroyTx` — crypto-shred: sets `destroyed_at`, zeroes `wrapped_dek`. Idempotent (no-op if already destroyed).
- `internal/modules/security/application/tenant_crypto_service.go` — `TenantCryptoService` orchestrates provisioning and DEK resolution against this repository.
- `internal/modules/audit/infrastructure/postgres/writer.go` — reads the key (same-tx visibility) when sealing an audit event that depends on a just-provisioned tenant key.
- `internal/composition/tenantdata/registry/registry.go` — documents this table as the tenant lifecycle export/erase registry's crypto-key entry.

## Tenant scoping

PK is `tenant_id` (uuid), FK to `metaldocs.tenants(id)` — one row per tenant.
Every query carries an explicit `tenant_id = $1` predicate rather than relying
on RLS session state, because reads run off the shared pool with no GUC seeded
(M6 F6.4 idiom, called out directly in the file header). A cross-tenant lookup
by a different `tenant_id` simply returns no row (`sql.ErrNoRows` → `(nil,
false, nil)`), never another tenant's key.
