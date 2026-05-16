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

INSERT INTO public.schema_migrations (version, description)
VALUES ('0203', 'rename templates v2 object names to current naming')
ON CONFLICT (version) DO NOTHING;

COMMIT;
