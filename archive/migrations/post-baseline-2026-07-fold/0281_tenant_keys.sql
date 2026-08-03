-- 0281_tenant_keys.sql
-- M7 F7.3 Task B (global-maximum-remediation, milestone-7-tenant-lifecycle,
-- ADR 0070 decision 4): the tenant DEK/KEK crypto-shred framework needs a
-- durable, tenant-scoped home for each tenant's wrapped data-encryption key
-- (DEK). metaldocs.tenant_keys stores exactly one row per tenant: the DEK
-- wrapped under the deployment KEK (METALDOCS_TENANT_KEK env, never
-- persisted in this table) plus a destroyed_at marker. Destroying the key
-- (DestroyTenantKeyTx, F7.3 item 6) makes every envelope ciphertext sealed
-- under that DEK permanently unrecoverable — the crypto-shred erasure step.
--
-- Shape:
--   - tenant_id uuid PRIMARY KEY REFERENCES metaldocs.tenants(id) — one key
--     per tenant, FK-anchored to the tenant root row (never a broken
--     reference; matches the tenants(id) PK type exactly).
--   - wrapped_dek bytea NOT NULL — WrapDEK(kek, dek) output. Zeroed
--     ('' :: bytea) by DestroyTenantKeyTx rather than deleted, so the row
--     (and its tenant_id FK anchor) survives permanently — matches the
--     tenants row itself never being deleted (F7.3 item 6, tombstone
--     pattern).
--   - created_at timestamptz NOT NULL DEFAULT now().
--   - destroyed_at timestamptz NULL — set once, by DestroyTenantKeyTx; a
--     non-NULL destroyed_at is the crypto-shred marker DecryptForTenant
--     checks before attempting UnwrapDEK.
--
-- RLS: tenant_keys is tenant-scoped from birth (milestone constraint) —
-- ENABLE + FORCE ROW LEVEL SECURITY with the tenant_isolation policy idiom
-- copied verbatim from the newest tenant-scoped migration (0258
-- document_families_tenant_scope.sql): NULLIF(...) IS NULL bypass for
-- GUC-unset sessions (migration/bootstrap contexts), else
-- tenant_id = the tx-local metaldocs.tenant_id GUC.
--
-- No tripwire arm in this migration: TenantCrypto's Go-layer writes
-- (ProvisionTenantKeyTx / DestroyTenantKeyTx) are invoked from within
-- OnboardTenantService's / the erasure worker's own tx, which already
-- asserts tenant.onboard / tenant.erase before any INSERT in that tx --
-- Task A/D own the tenant_lifecycle_jobs tripwire arm (registry 19->20,
-- validation-contract.md ss5 touchpoint 6); tenant_keys itself carries no
-- separate capability-gated write surface of its own in this task's scope.

BEGIN;

CREATE TABLE metaldocs.tenant_keys (
    tenant_id    uuid PRIMARY KEY REFERENCES metaldocs.tenants(id),
    wrapped_dek  bytea NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    destroyed_at timestamptz
);

ALTER TABLE metaldocs.tenant_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY metaldocs.tenant_keys FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON metaldocs.tenant_keys;
CREATE POLICY tenant_isolation ON metaldocs.tenant_keys
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0281', 'M7 F7.3 Task B (ADR 0070 d4): create metaldocs.tenant_keys (tenant_id PK/FK, wrapped_dek, created_at, destroyed_at) with FORCE RLS + tenant_isolation policy -- backs the TenantCrypto DEK/KEK crypto-shred port.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
