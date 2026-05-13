BEGIN;

ALTER TABLE templates_v2_template_version
    DROP COLUMN IF EXISTS editable_zones_schema_snapshot,
    DROP COLUMN IF EXISTS editable_zones_schema,
    DROP COLUMN IF EXISTS editable_zones;

ALTER TABLE documents
    DROP COLUMN IF EXISTS editable_zones_schema_snapshot,
    DROP COLUMN IF EXISTS editable_zones_schema;

DROP TABLE IF EXISTS document_editable_zone_content;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0196', 'drop editable zone columns and content table')
ON CONFLICT (version) DO NOTHING;

COMMIT;
