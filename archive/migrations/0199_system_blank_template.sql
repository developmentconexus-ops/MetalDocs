BEGIN;

-- Historical evidence migration:
-- this migration seeds the immutable blank template metadata row, but the runtime
-- invariant also requires the MinIO object `system/templates/blank.docx` to exist.
-- Curated bootstrap/runtime scripts own that storage seed path; do not treat this
-- SQL row alone as sufficient for controlled document creation.

-- Migration-time writes on templates_* tables pass through authz tripwire checks.
-- Seed with an asserted capability set that satisfies template mutation guards.
SELECT set_config(
  'metaldocs.asserted_caps',
  '[{"cap":"template.create"},{"cap":"template.edit"},{"cap":"template.submit"},{"cap":"template.approve"},{"cap":"template.publish"}]',
  true
);

ALTER TABLE templates_v2_template
  ADD COLUMN IF NOT EXISTS system_owned boolean NOT NULL DEFAULT false;

CREATE UNIQUE INDEX IF NOT EXISTS ux_templates_v2_system_blank
  ON templates_v2_template (tenant_id, key)
  WHERE system_owned = true AND key = '__system_blank__';

INSERT INTO templates_v2_template (
  id, tenant_id, doc_type_code, key, name, description, areas, visibility,
  specific_areas, latest_version, published_version_id, created_by, system_owned
) VALUES (
  '00000000-0000-0000-0000-000000000101',
  'ffffffff-ffff-ffff-ffff-ffffffffffff',
  'system',
  '__system_blank__',
  'Em branco',
  'System blank template for controlled document creation.',
  '{}'::text[],
  'internal',
  '{}'::text[],
  1,
  NULL,
  'system',
  true
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO templates_v2_template_version (
  id, template_id, version_number, status, docx_storage_key, content_hash,
  metadata_schema, placeholder_schema, author_id,
  pending_reviewer_role, pending_approver_role,
  reviewer_id, approver_id,
  submitted_at, reviewed_at, approved_at, published_at, obsoleted_at,
  created_at
) VALUES (
  '00000000-0000-0000-0000-000000000102',
  '00000000-0000-0000-0000-000000000101',
  1,
  'published',
  'system/templates/blank.docx',
  'system-blank-template',
  '{}'::jsonb,
  '[]'::jsonb,
  'system',
  NULL,
  'system',
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  now(),
  NULL,
  now()
)
ON CONFLICT (id) DO NOTHING;

UPDATE templates_v2_template
SET
  system_owned = true,
  latest_version = 1,
  published_version_id = '00000000-0000-0000-0000-000000000102'
WHERE id = '00000000-0000-0000-0000-000000000101';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0199', 'seed immutable system blank template/version')
ON CONFLICT (version) DO NOTHING;

COMMIT;
