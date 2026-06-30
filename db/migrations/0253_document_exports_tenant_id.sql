-- 0253_document_exports_tenant_id.sql
-- F-O4 [HIGH]: document_exports had no tenant_id column and was absent from RLS.
-- Violated REQ-TEN-1 (multi-tenant invariant: every tenant table carries tenant_id
-- and is covered by the metaldocs.tenant_id GUC policy).
--
-- FIX (forward-only, idempotent):
--   1. Add tenant_id uuid column (initially nullable for backfill safety).
--   2. Backfill from documents.tenant_id via JOIN; fail loudly if any orphan
--      document_exports row has no matching documents.tenant_id.
--   3. Set NOT NULL + FK to metaldocs.tenants(id) ON DELETE CASCADE.
--   4. Drop the old (document_id, composite_hash) unique index (column-pair is
--      not globally unique once tenant_id is added — two tenants could share a
--      composite_hash for different documents under cross-tenant use, but more
--      precisely the uniqueness domain is per-tenant).
--      Replace with (tenant_id, document_id, composite_hash) unique index.
--   5. Enable + Force RLS; add the same NULL-permissive tenant_isolation policy
--      used verbatim from 0234/0237 for every other tenant-scoped table.
--
-- SIBLING PATTERN: RLS block copied verbatim from 0237_rls_all_tenant_tables.sql.
-- Unique-index naming follows the uq_ prefix convention from 0250/0252.
--
-- SUPERUSER CAVEAT (0234 / 0237): the dev Docker role metaldocs_app is a
-- superuser with BYPASSRLS — RLS does not apply to superusers. In production
-- the app role must be NOSUPERUSER + NOBYPASSRLS (ADR 0022 Phase 5, ADR 0027).
--
-- FORCE ROW LEVEL SECURITY: required because table owner = app role.

BEGIN;

-- ── Step 1: add column (nullable for backfill) ───────────────────────────────

ALTER TABLE public.document_exports
  ADD COLUMN IF NOT EXISTS tenant_id uuid;

-- ── Step 2: backfill from documents.tenant_id ────────────────────────────────
-- Fail loudly if any document_exports row has no matching documents row or
-- the matched documents row has a NULL tenant_id (should never happen on a
-- healthy DB, but surfaces data corruption if it does).

DO $$
DECLARE
  orphan_count integer;
BEGIN
  SELECT COUNT(*) INTO orphan_count
  FROM public.document_exports de
  LEFT JOIN public.documents d ON d.id = de.document_id
  WHERE de.tenant_id IS NULL
    AND (d.id IS NULL OR d.tenant_id IS NULL);

  IF orphan_count > 0 THEN
    RAISE EXCEPTION
      'migration 0253: % document_exports row(s) have no matching documents.tenant_id — '
      'resolve data corruption before re-running this migration.',
      orphan_count;
  END IF;
END;
$$;

UPDATE public.document_exports de
SET    tenant_id = d.tenant_id
FROM   public.documents d
WHERE  d.id = de.document_id
  AND  de.tenant_id IS NULL;

-- ── Step 3: NOT NULL + FK ────────────────────────────────────────────────────

ALTER TABLE public.document_exports
  ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE public.document_exports
  DROP CONSTRAINT IF EXISTS fk_document_exports_tenant;
ALTER TABLE public.document_exports
  ADD CONSTRAINT fk_document_exports_tenant
  FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id) ON DELETE CASCADE;

-- ── Step 4: replace unique index ─────────────────────────────────────────────
-- Drop the old (document_id, composite_hash) unique index that the repo's
-- ON CONFLICT clause previously targeted. Replace with a tenant-scoped index.

DROP INDEX IF EXISTS public.document_exports_doc_hash_uidx;

CREATE UNIQUE INDEX IF NOT EXISTS uq_document_exports_tenant_doc_hash
  ON public.document_exports (tenant_id, document_id, composite_hash);

-- ── Step 5: RLS — verbatim from 0237 policy shape ───────────────────────────

ALTER TABLE public.document_exports ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.document_exports FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.document_exports;
CREATE POLICY tenant_isolation ON public.document_exports
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0253', 'document_exports: add tenant_id NOT NULL + FK + RLS + replace unique index (F-O4 HIGH, REQ-TEN-1)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
