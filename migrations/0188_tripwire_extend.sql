-- migrations/0188_tripwire_extend.sql
-- Plan 5: extend enforce_capability_asserted trigger to all regulated tables.
-- Also fixes doc.submit/doc.signoff → document.submit/document.signoff (missed by 0186).
-- Also adds document_families.code immutability trigger (mirrors migrations/0122, 0123).
--
-- Idempotent: DROP TRIGGER IF EXISTS before CREATE; CREATE OR REPLACE FUNCTION.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Extend the tripwire function.
--    Replaces the 0142b version: adds all new table/cap entries, updates
--    doc.* → document.* cap names for approval tables, and switches from
--    scalar v_required_cap to array v_required_caps for OR-logic (needed for
--    controlled_documents UPDATE which accepts either registry.obsolete or
--    registry.supersede).
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()
  RETURNS trigger
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  v_bypass        TEXT;
  v_asserted_raw  TEXT;
  v_asserted      JSONB;
  v_required_caps TEXT[];   -- one or more acceptable caps for this table/op
  v_tenant_id     UUID;
  v_cap_found     BOOLEAN := FALSE;
  v_element       JSONB;
BEGIN
  -- ---- Determine required capability set for this table/operation. --------
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
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      -- Either lifecycle cap is acceptable.
      v_required_caps := ARRAY['registry.obsolete', 'registry.supersede'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      -- document_families has no tenant_id column.
      v_tenant_id     := NULL;

    WHEN TG_TABLE_NAME = 'templates_v2_template' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'templates_v2_template_version' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      -- template_version has no tenant_id column.
      v_tenant_id     := NULL;

    ELSE
      -- Unknown table — conservative pass-through.
      RETURN NEW;
  END CASE;

  -- ---- Bypass path. -------------------------------------------------------
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
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %',
                      v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

  -- ---- Read asserted_caps GUC. --------------------------------------------
  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);
  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %',
                    v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  BEGIN
    v_asserted := v_asserted_raw::JSONB;
  EXCEPTION WHEN invalid_text_representation OR others THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)',
                    v_required_caps
      USING ERRCODE = 'P0001';
  END;

  IF jsonb_typeof(v_asserted) <> 'array' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)',
                    v_required_caps
      USING ERRCODE = 'P0001';
  END IF;

  -- ---- Scan for any required cap. ----------------------------------------
  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP
    IF (v_element->>'cap') = ANY(v_required_caps) THEN
      v_cap_found := TRUE;
      EXIT;
    END IF;
  END LOOP;

  IF NOT v_cap_found THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %',
                    v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  RETURN NEW;
END;
$$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_security_owner') THEN
    EXECUTE 'ALTER FUNCTION public.enforce_capability_asserted() OWNER TO metaldocs_security_owner';
  END IF;
END $$;
REVOKE EXECUTE ON FUNCTION public.enforce_capability_asserted() FROM PUBLIC;

-- ---------------------------------------------------------------------------
-- 2. Attach trigger to new tables.
--    approval_instances + approval_signoffs already have triggers (0142b).
-- ---------------------------------------------------------------------------

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.iam_user_roles;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.iam_user_roles
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.user_process_areas;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.user_process_areas
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.documents;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE ON public.documents
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.controlled_documents;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE ON public.controlled_documents
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.cd_sequence_counters;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE ON public.cd_sequence_counters
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.document_profiles;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.document_profiles
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.document_process_areas;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.document_process_areas
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.document_families;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.document_families
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.templates_v2_template;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.templates_v2_template
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.templates_v2_template_version;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.templates_v2_template_version
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

-- ---------------------------------------------------------------------------
-- 3. document_families.code immutability trigger (mirrors 0122/0123 pattern).
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.reject_families_code_update()
  RETURNS trigger
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  IF NEW.code IS DISTINCT FROM OLD.code THEN
    RAISE EXCEPTION 'document_families.code is immutable'
      USING ERRCODE = 'P0002';
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_reject_families_code_update ON metaldocs.document_families;
CREATE TRIGGER trg_reject_families_code_update
  BEFORE UPDATE ON metaldocs.document_families
  FOR EACH ROW EXECUTE FUNCTION public.reject_families_code_update();

INSERT INTO public.schema_migrations (version, description)
VALUES ('0188', 'Plan 5: extend tripwire to all regulated tables + document_families code immutability')
ON CONFLICT (version) DO NOTHING;

COMMIT;
