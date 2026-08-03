-- 0258_document_families_tenant_scope.sql
-- SEC-01 (P0): metaldocs.document_families carries no tenant_id column
-- (baseline 0001_current_schema.sql:1021-1027) — the DB tripwire CASE branch
-- for it sets v_tenant_id := NULL (baseline :464-466), and every
-- FamilyRepository query is unscoped (family_repository.go). A qms_admin (or
-- any actor holding taxonomy.manage) in tenant A can mutate a family and the
-- write is visible/effective for every tenant sharing the deployment —
-- violates the pooled-tenancy invariant (every tenant table carries
-- tenant_id; wiki/modules/taxonomy-tech-debt.md T-002).
--
-- DEC-02: tenant-scope document_families, matching the EXACT pattern already
-- used by its sibling taxonomy tables document_profiles (0122) and
-- document_process_areas (0123):
--   - tenant_id UUID NOT NULL DEFAULT sentinel ('ffffffff-ffff-ffff-ffff-ffffffffffff')
--   - PRIMARY KEY stays on (code) alone — code remains a GLOBAL key. This is
--     load-bearing: document_profiles.family_code_fkey references
--     document_families(code) (baseline :4008-4009), not a composite key, and
--     ADR 0038 / FamilyCodeResolver depend on document_profiles.code being a
--     global PK with sentinel-tenant fallback precedence
--     (tenant_id IN (tenant, sentinel)). Widening document_families' PK to
--     (tenant_id, code) would be consistent for this table alone but would
--     break nothing structurally either way since the FK target key doesn't
--     change shape — kept as (code) to mirror the sibling pattern exactly, as
--     directed, and to avoid touching the FK definition at all.
--   - UNIQUE(tenant_id, code) index added for tenant-scoped lookups (mirrors
--     ux_document_profiles_tenant_code / ux_process_areas_tenant_code).
--   - RLS enabled with the identical tenant_isolation policy shape used by
--     document_profiles / document_process_areas (NULLIF(...) IS NULL bypass
--     for GUC-unset sessions, else tenant_id = GUC).
--   - Existing rows backfill to the sentinel tenant (preserves current
--     cross-deployment visibility for pre-existing global families; new
--     tenant-owned families become possible going forward exactly as
--     document_profiles/document_process_areas already allow).
--   - enforce_capability_asserted() document_families CASE branch changes
--     v_tenant_id from NULL to NEW.tenant_id — governance events for family
--     mutations are now tenant-attributed instead of always NULL.
--
-- No FK from other tables needs to become composite: document_profiles and
-- controlled_document-adjacent tables reference document_families(code) only
-- (a plain, still-globally-unique key) — untouched by this migration.

BEGIN;

-- ── Step 1: add tenant_id with sentinel default (mirrors 0122/0123) ─────────

ALTER TABLE metaldocs.document_families
  ADD COLUMN IF NOT EXISTS tenant_id UUID
    DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff';

-- Backfill: every pre-existing row is a pre-tenant-scope global family. Point
-- it at the sentinel tenant so current cross-tenant visibility is preserved
-- exactly (no behavior change for existing data) — only new writes going
-- forward can be tenant-owned.
--
-- The backfill UPDATE fires trg_require_cap_asserted (BEFORE ... UPDATE on
-- this table), which after Step 4 below would require metaldocs.asserted_caps
-- to already contain taxonomy.manage — not guaranteed for a bare migration
-- session. Stand the tripwire down for this system DDL data-fix only,
-- mirroring the established disable → data-fix → re-enable convention used by
-- migration 0257 for the same class of conflict.
ALTER TABLE metaldocs.document_families DISABLE TRIGGER trg_require_cap_asserted;

UPDATE metaldocs.document_families
   SET tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
 WHERE tenant_id IS NULL;

ALTER TABLE metaldocs.document_families ENABLE TRIGGER trg_require_cap_asserted;

ALTER TABLE metaldocs.document_families
  ALTER COLUMN tenant_id SET NOT NULL;

-- ── Step 2: tenant-scoped unique index (mirrors ux_document_profiles_tenant_code) ──

CREATE UNIQUE INDEX IF NOT EXISTS ux_document_families_tenant_code
  ON metaldocs.document_families (tenant_id, code);

-- ── Step 3: RLS (mirrors document_profiles / document_process_areas tenant_isolation) ──

ALTER TABLE metaldocs.document_families ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY metaldocs.document_families FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON metaldocs.document_families;
CREATE POLICY tenant_isolation ON metaldocs.document_families
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

-- ── Step 4: enforce_capability_asserted() — attribute family mutations to the ──
-- ── acting tenant instead of NULL. CREATE OR REPLACE preserves every other  ──
-- ── CASE branch byte-for-byte from baseline 0001_current_schema.sql:417-542. ──

CREATE OR REPLACE FUNCTION public.enforce_capability_asserted() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
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
      v_tenant_id     := NEW.tenant_id;
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
$$;

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0258', 'SEC-01: tenant-scope metaldocs.document_families (tenant_id + RLS + tripwire attribution), mirrors document_profiles/document_process_areas sentinel pattern')
ON CONFLICT (version) DO NOTHING;

COMMIT;
