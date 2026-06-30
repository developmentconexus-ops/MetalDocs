-- 0256_templates_template_version_tenant_id.sql
-- F-DB5 [MEDIUM]: templates_template_version had no tenant_id column. Migration
-- 0237 (RLS rollout) skipped it, and the enforce_capability_asserted trigger set
-- v_tenant_id := NULL for it (db/baseline/0001_current_schema.sql:482-486). Tenant
-- isolation for this table depended 100% on the JOIN to templates_template — a
-- single point of failure. The short-term guards (0255 tenant-consistency trigger
-- + ObsoletePreviousPublished tenant predicate) are already in place; this is the
-- medium-term RLS-parity fix: give the table its own real tenant_id column so it
-- satisfies REQ-TEN-1 (every tenant table carries tenant_id and is covered by the
-- metaldocs.tenant_id GUC policy) directly, not by proxy.
--
-- FIX (forward-only, idempotent):
--   1. Add tenant_id uuid column (initially nullable for backfill safety).
--   2. Backfill from templates_template.tenant_id via JOIN on template_id; fail
--      loudly if any version row has no matching templates_template.tenant_id.
--   3. Set NOT NULL + FK to metaldocs.tenants(id) ON DELETE CASCADE.
--   4. (intentionally SKIPPED — see note below)
--   5. Enable + Force RLS; add the same NULL-permissive tenant_isolation policy
--      used verbatim from 0234/0237 for every other tenant-scoped table.
--   6. CREATE OR REPLACE the enforce_capability_asserted trigger function so the
--      templates_template_version CASE branch resolves the real NEW.tenant_id
--      (matching the templates_template branch) instead of NULL.
--
-- NO INDEX CHANGE (the 0253 step-4 analogue is deliberately omitted): the existing
-- UNIQUE constraint on templates_template_version is (template_id, version_number).
-- template_id is a globally-unique uuid FK to templates_template(id), so the
-- uniqueness domain is already effectively per-tenant — adding tenant_id widens no
-- key and could create no cross-tenant collision. Touching the unique index would
-- be a no-op risk. Existing indexes/constraints are left untouched on purpose.
--
-- SIBLING PATTERN: structure + RLS block mirror 0253_document_exports_tenant_id.sql
-- (which itself copies the RLS policy verbatim from 0237_rls_all_tenant_tables.sql).
--
-- SUPERUSER CAVEAT (0234 / 0237): the dev Docker role metaldocs_app is a superuser
-- with BYPASSRLS — RLS does not apply to superusers. In production the app role
-- must be NOSUPERUSER + NOBYPASSRLS (ADR 0022 Phase 5, ADR 0027).
--
-- FORCE ROW LEVEL SECURITY: required because table owner = app role.

BEGIN;

-- ── Step 1: add column (nullable for backfill) ───────────────────────────────

ALTER TABLE public.templates_template_version
  ADD COLUMN IF NOT EXISTS tenant_id uuid;

-- ── Step 2: backfill from templates_template.tenant_id ───────────────────────
-- Fail loudly if any version row has no matching templates_template row or the
-- matched parent row has a NULL tenant_id (should never happen on a healthy DB,
-- but surfaces data corruption if it does).

DO $$
DECLARE
  orphan_count integer;
BEGIN
  SELECT COUNT(*) INTO orphan_count
  FROM public.templates_template_version v
  LEFT JOIN public.templates_template t ON t.id = v.template_id
  WHERE v.tenant_id IS NULL
    AND (t.id IS NULL OR t.tenant_id IS NULL);

  IF orphan_count > 0 THEN
    RAISE EXCEPTION
      'migration 0256: % templates_template_version row(s) have no matching templates_template.tenant_id — '
      'resolve data corruption before re-running this migration.',
      orphan_count;
  END IF;
END;
$$;

-- The backfill is a system DDL data-fix, not a user-driven write, but the table
-- carries TWO BEFORE INSERT/UPDATE triggers that would reject it:
--   (1) trg_require_cap_asserted (baseline) — rejects writes without an
--       asserted-caps GUC.
--   (2) trg_template_version_tenant_consistent (migration 0255) — rejects writes
--       without the metaldocs.tenant_id GUC, and checks the row's parent-template
--       tenant against it. The per-row backfill cannot set a single tx-local
--       tenant GUC that matches every row, so this trigger must also stand down
--       for the duration of the system data-fix.
-- Disable both around the backfill and re-enable immediately after — the
-- canonical migration-time pattern from 0230_authz_decommission_reviewer_role.sql.
-- Both ENABLEs run inside this migration's BEGIN/COMMIT, so any failure rolls the
-- whole tx back and can never leave a security trigger disabled. The 0255 trigger
-- is created by an earlier migration in the same lineage, but it is guarded with
-- an existence check so this migration stays robust if applied to a DB that never
-- received 0255 (e.g. a pre-fold environment).
ALTER TABLE public.templates_template_version DISABLE TRIGGER trg_require_cap_asserted;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_template_version_tenant_consistent'
      AND tgrelid = 'public.templates_template_version'::regclass
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version DISABLE TRIGGER trg_template_version_tenant_consistent';
  END IF;
END
$$;

UPDATE public.templates_template_version v
SET    tenant_id = t.tenant_id
FROM   public.templates_template t
WHERE  t.id = v.template_id
  AND  v.tenant_id IS NULL;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_template_version_tenant_consistent'
      AND tgrelid = 'public.templates_template_version'::regclass
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version ENABLE TRIGGER trg_template_version_tenant_consistent';
  END IF;
END
$$;

ALTER TABLE public.templates_template_version ENABLE TRIGGER trg_require_cap_asserted;

-- ── Step 3: NOT NULL + FK ────────────────────────────────────────────────────

ALTER TABLE public.templates_template_version
  ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE public.templates_template_version
  DROP CONSTRAINT IF EXISTS fk_templates_template_version_tenant;
ALTER TABLE public.templates_template_version
  ADD CONSTRAINT fk_templates_template_version_tenant
  FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id) ON DELETE CASCADE;

-- ── Step 4: (skipped) no unique-index change ─────────────────────────────────
-- See the header note: the existing UNIQUE (template_id, version_number) already
-- has a per-tenant uniqueness domain because template_id is a globally-unique uuid
-- FK. Adding tenant_id does not widen any key. No index is dropped or replaced.

-- ── Step 5: RLS — verbatim from 0237 policy shape ───────────────────────────

ALTER TABLE public.templates_template_version ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.templates_template_version FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.templates_template_version;
CREATE POLICY tenant_isolation ON public.templates_template_version
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- ── Step 6: trigger function — resolve real tenant_id for version writes ──────
-- The table now carries tenant_id, so the enforce_capability_asserted CASE branch
-- for templates_template_version must resolve NEW.tenant_id (so the scheduler
-- bypass governance_events row gets the correct tenant) instead of NULL. This is
-- a byte-for-byte reproduction of the live function (pg_get_functiondef) with
-- EXACTLY ONE line changed: the templates_template_version branch assignment.
-- SECURITY-CRITICAL: every other branch is preserved verbatim (ADR 0022).

CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()
 RETURNS trigger
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path TO 'pg_catalog', 'pg_temp'
AS $function$
DECLARE
  v_bypass        TEXT;
  v_asserted_raw  TEXT;
  v_asserted      JSONB;
  v_required_caps TEXT[];
  v_tenant_id     UUID;
  v_cap_found     BOOLEAN := FALSE;
  v_element       JSONB;
BEGIN
  CASE
    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.submit'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.signoff'];
      v_tenant_id     := NEW.actor_tenant_id;
    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'user_process_areas' THEN
      v_required_caps := ARRAY['membership.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['document.edit'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['controlled_documents.obsolete', 'controlled_documents.supersede'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NULL;
    WHEN TG_TABLE_NAME = 'templates_template' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'templates_template_version' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;
    ELSE
      -- Fail-closed: a table carrying this trigger with no capability mapping is a
      -- wiring error, not a license to pass through. Refuse the write loudly.
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for table % (op %); trg_require_cap_asserted attached without an enforce_capability_asserted CASE branch', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'P0001';
  END CASE;

  v_bypass := pg_catalog.current_setting('metaldocs.bypass_authz', true);
  IF v_bypass IS NOT NULL AND v_bypass <> '' THEN
    IF v_bypass = 'scheduler' THEN
      BEGIN
        INSERT INTO public.governance_events
          (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
        VALUES (
          v_tenant_id,
          'authz.bypass_used',
          'system:scheduler',
          TG_TABLE_NAME,
          COALESCE(NEW.id::TEXT, 'unknown'),
          'scheduler bypass for ' || v_required_caps[1],
          pg_catalog.to_jsonb(jsonb_build_object(
            'required_caps', to_jsonb(v_required_caps),
            'bypass_token',  v_bypass,
            'table',         TG_TABLE_NAME,
            'op',            TG_OP
          ))
        );
      EXCEPTION WHEN others THEN
        RAISE NOTICE 'enforce_capability_asserted: governance_events insert failed: %', SQLERRM;
      END;
      RETURN NEW;
    ELSE
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %', v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);
  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %', v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  BEGIN
    v_asserted := v_asserted_raw::JSONB;
  EXCEPTION WHEN invalid_text_representation OR others THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)', v_required_caps
      USING ERRCODE = 'P0001';
  END;

  IF jsonb_typeof(v_asserted) <> 'array' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)', v_required_caps
      USING ERRCODE = 'P0001';
  END IF;

  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP
    IF (v_element->>'cap') = ANY(v_required_caps) THEN
      v_cap_found := TRUE;
      EXIT;
    END IF;
  END LOOP;

  IF NOT v_cap_found THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %', v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  RETURN NEW;
END;
$function$;

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0256', 'templates_template_version: add tenant_id NOT NULL + FK + RLS + trigger branch resolves real tenant_id (F-DB5 MEDIUM, REQ-TEN-1)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
