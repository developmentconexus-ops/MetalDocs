-- 0278_tenant_lifecycle_jobs.sql
-- M7 F7.3 Task A (global-maximum-remediation, milestone-7-tenant-lifecycle,
-- ADR 0070): item 7 of the feature spec — a new tenant-scoped
-- metaldocs.tenant_lifecycle_jobs table backing the tenant export/erase
-- workflow (POST /tenants/{tenantId}/export and .../erase enqueue a row here;
-- the async worker fan-out is a separate task, not implemented by this
-- migration). Also adds metaldocs.tenants.erased_at, the erasure tombstone
-- marker column the F7.3 erase workflow will set (Task B/C).
--
-- Shape (per spec item 7): id, tenant_id FK -> metaldocs.tenants(id), kind
-- ('export'|'erase'), status ('pending'|'running'|'ready'|'failed'),
-- requested_by (actor id, TEXT — mirrors the actor-id-is-TEXT contract used
-- across the codebase, e.g. iam_users.user_id), object_key (nullable — the
-- blob key once an export artifact exists), error (nullable), created_at,
-- completed_at (nullable).
--
-- RLS: this is a tenant table (tenant_id column) so it MUST carry
-- FORCE ROW LEVEL SECURITY + a tenant_isolation policy per the pooled-tenancy
-- invariant. Idiom copied verbatim from the newest migration that adds RLS to
-- a tenant table, 0258_document_families_tenant_scope.sql (GUC
-- metaldocs.tenant_id, policy name tenant_isolation, NULLIF(...) IS NULL
-- bypass for GUC-unset sessions else tenant_id = GUC) — not invented here.
--
-- Tripwire: this migration does NOT attach trg_require_cap_asserted or extend
-- enforce_capability_asserted() — that is a separate, machine-generated step
-- (tripwire arm #20, internal/platform/tripwire + cmd/gen-tripwire ->
-- db/migrations/0279_*.sql), per the 0259/0277 attachment precedent of
-- keeping the generated trigger-function change in its own generated file.

BEGIN;

CREATE TABLE metaldocs.tenant_lifecycle_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES metaldocs.tenants(id),
    kind text NOT NULL CHECK (kind IN ('export', 'erase')),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'ready', 'failed')),
    requested_by text NOT NULL,
    object_key text,
    error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

CREATE INDEX ix_tenant_lifecycle_jobs_tenant_id ON metaldocs.tenant_lifecycle_jobs (tenant_id);

ALTER TABLE metaldocs.tenant_lifecycle_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY metaldocs.tenant_lifecycle_jobs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON metaldocs.tenant_lifecycle_jobs;
CREATE POLICY tenant_isolation ON metaldocs.tenant_lifecycle_jobs
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

-- ── tenants.erased_at (F7.3 erasure tombstone marker column) ────────────────

ALTER TABLE metaldocs.tenants ADD COLUMN erased_at timestamptz;

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0278', 'M7 F7.3 Task A: create metaldocs.tenant_lifecycle_jobs (export/erase job tracking, tenant-scoped, FORCE RLS + tenant_isolation policy per 0258 idiom) and add metaldocs.tenants.erased_at tombstone column.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
