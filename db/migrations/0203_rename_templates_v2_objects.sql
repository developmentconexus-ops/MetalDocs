BEGIN;

DO $$
BEGIN
  IF to_regclass('public.templates_v2_template') IS NOT NULL
     AND to_regclass('public.templates_template') IS NULL THEN
    EXECUTE 'ALTER TABLE public.templates_v2_template RENAME TO templates_template';
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.templates_v2_template_version') IS NOT NULL
     AND to_regclass('public.templates_template_version') IS NULL THEN
    EXECUTE 'ALTER TABLE public.templates_v2_template_version RENAME TO templates_template_version';
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.templates_v2_approval_config') IS NOT NULL
     AND to_regclass('public.templates_approval_config') IS NULL THEN
    EXECUTE 'ALTER TABLE public.templates_v2_approval_config RENAME TO templates_approval_config';
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.templates_v2_audit_log') IS NOT NULL
     AND to_regclass('public.templates_audit_log') IS NULL THEN
    EXECUTE 'ALTER TABLE public.templates_v2_audit_log RENAME TO templates_audit_log';
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.templates_v2_audit_log_id_seq') IS NOT NULL
     AND to_regclass('public.templates_audit_log_id_seq') IS NULL THEN
    EXECUTE 'ALTER SEQUENCE public.templates_v2_audit_log_id_seq RENAME TO templates_audit_log_id_seq';
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.idx_templates_v2_audit_template_time') IS NOT NULL
     AND to_regclass('public.idx_templates_audit_template_time') IS NULL THEN
    EXECUTE 'ALTER INDEX public.idx_templates_v2_audit_template_time RENAME TO idx_templates_audit_template_time';
  END IF;

  IF to_regclass('public.idx_templates_v2_template_tenant_doctype') IS NOT NULL
     AND to_regclass('public.idx_templates_template_tenant_doctype') IS NULL THEN
    EXECUTE 'ALTER INDEX public.idx_templates_v2_template_tenant_doctype RENAME TO idx_templates_template_tenant_doctype';
  END IF;

  IF to_regclass('public.idx_templates_v2_version_template_status') IS NOT NULL
     AND to_regclass('public.idx_templates_version_template_status') IS NULL THEN
    EXECUTE 'ALTER INDEX public.idx_templates_v2_version_template_status RENAME TO idx_templates_version_template_status';
  END IF;

  IF to_regclass('public.ux_templates_v2_system_blank') IS NOT NULL
     AND to_regclass('public.ux_templates_system_blank') IS NULL THEN
    EXECUTE 'ALTER INDEX public.ux_templates_v2_system_blank RENAME TO ux_templates_system_blank';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_template'::regclass
      AND conname = 'templates_v2_template_pkey'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template RENAME CONSTRAINT templates_v2_template_pkey TO templates_template_pkey';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_template'::regclass
      AND conname = 'templates_v2_template_tenant_id_key_key'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template RENAME CONSTRAINT templates_v2_template_tenant_id_key_key TO templates_template_tenant_id_key_key';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_template'::regclass
      AND conname = 'fk_templates_v2_published_version'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template RENAME CONSTRAINT fk_templates_v2_published_version TO fk_templates_published_version';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_template_version'::regclass
      AND conname = 'templates_v2_template_version_pkey'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version RENAME CONSTRAINT templates_v2_template_version_pkey TO templates_template_version_pkey';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_template_version'::regclass
      AND conname = 'templates_v2_template_version_template_id_version_number_key'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version RENAME CONSTRAINT templates_v2_template_version_template_id_version_number_key TO templates_template_version_template_id_version_number_key';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_template_version'::regclass
      AND conname = 'templates_v2_template_version_template_id_fkey'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version RENAME CONSTRAINT templates_v2_template_version_template_id_fkey TO templates_template_version_template_id_fkey';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_approval_config'::regclass
      AND conname = 'templates_v2_approval_config_pkey'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_approval_config RENAME CONSTRAINT templates_v2_approval_config_pkey TO templates_approval_config_pkey';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_approval_config'::regclass
      AND conname = 'templates_v2_approval_config_template_id_fkey'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_approval_config RENAME CONSTRAINT templates_v2_approval_config_template_id_fkey TO templates_approval_config_template_id_fkey';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.templates_audit_log'::regclass
      AND conname = 'templates_v2_audit_log_pkey'
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_audit_log RENAME CONSTRAINT templates_v2_audit_log_pkey TO templates_audit_log_pkey';
  END IF;
END
$$;

CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
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
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
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
      v_tenant_id     := NULL;

    WHEN TG_TABLE_NAME = 'templates_template' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'templates_template_version' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      v_tenant_id     := NULL;

    ELSE
      RETURN NEW;
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
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %',
                      v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

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
  BEGIN
    EXECUTE 'ALTER FUNCTION public.enforce_capability_asserted() OWNER TO metaldocs_security_owner';
  EXCEPTION WHEN undefined_object OR insufficient_privilege THEN
    NULL;
  END;
END
$$;

REVOKE EXECUTE ON FUNCTION public.enforce_capability_asserted() FROM PUBLIC;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0203', 'rename templates v2 object names to current naming')
ON CONFLICT (version) DO NOTHING;

COMMIT;
