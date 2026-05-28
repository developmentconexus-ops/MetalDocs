-- 0213_templates_tenant_id_uuid.sql
BEGIN;

DO $$
DECLARE
    invalid_templates text;
    invalid_template_audit text;
BEGIN
    SELECT string_agg(tenant_id, ', ' ORDER BY tenant_id)
      INTO invalid_templates
      FROM (
        SELECT DISTINCT tenant_id
          FROM public.templates_template
         WHERE tenant_id IS NULL
            OR tenant_id !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
      ) invalid_values;

    IF invalid_templates IS NOT NULL THEN
        RAISE EXCEPTION
            '0213_templates_tenant_id_uuid: backfill required for public.templates_template.tenant_id values: %',
            invalid_templates;
    END IF;

    SELECT string_agg(tenant_id, ', ' ORDER BY tenant_id)
      INTO invalid_template_audit
      FROM (
        SELECT DISTINCT tenant_id
          FROM public.templates_audit_log
         WHERE tenant_id IS NULL
            OR tenant_id !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
      ) invalid_values;

    IF invalid_template_audit IS NOT NULL THEN
        RAISE EXCEPTION
            '0213_templates_tenant_id_uuid: backfill required for public.templates_audit_log.tenant_id values: %',
            invalid_template_audit;
    END IF;
END $$;

ALTER TABLE public.templates_template
    ALTER COLUMN tenant_id TYPE UUID
    USING tenant_id::uuid;

ALTER TABLE public.templates_audit_log
    ALTER COLUMN tenant_id TYPE UUID
    USING tenant_id::uuid;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0213', 'convert templates tenant_id columns from text to uuid')
ON CONFLICT (version) DO NOTHING;

COMMIT;
