-- MetalDocs product reference data.
-- Data in this file is required for every environment.

BEGIN;

INSERT INTO public.schema_migrations (version, description)
VALUES ('baseline-2026-05-14', 'curated current-state database baseline')
ON CONFLICT (version) DO NOTHING;

-- Canonical runtime capability matrix.
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'template.approve', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'template.review', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'membership.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'registry.obsolete', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'registry.supersede', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'registry.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'registry.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'template.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'registry.obsolete', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'registry.supersede', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'route.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'taxonomy.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.approve', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.publish', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.review', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('signer', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('signer', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'audit.read', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'membership.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'registry.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'registry.obsolete', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'registry.supersede', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'route.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'taxonomy.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.approve', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.publish', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.review', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'user.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- System blank template required for controlled document creation defaults.
SELECT set_config(
  'metaldocs.asserted_caps',
  '[{"cap":"template.create"},{"cap":"template.edit"},{"cap":"template.submit"},{"cap":"template.approve"},{"cap":"template.publish"}]',
  true
);

INSERT INTO public.templates_v2_template (
  id, tenant_id, doc_type_code, key, name, description, areas, visibility,
  specific_areas, latest_version, published_version_id, created_by, system_owned, archived_at
) VALUES (
  '00000000-0000-0000-0000-000000000101'::uuid,
  'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid,
  'system',
  '__system_blank__',
  'Em branco',
  'System blank template for controlled document creation.',
  ARRAY[]::text[],
  'internal',
  ARRAY[]::text[],
  1,
  NULL,
  'system',
  true,
  NULL
) ON CONFLICT (id) DO NOTHING;

INSERT INTO public.templates_v2_template_version (
  id, template_id, version_number, status, docx_storage_key, content_hash, metadata_schema,
  placeholder_schema, author_id, pending_reviewer_role, pending_approver_role, reviewer_id,
  approver_id, submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version
) VALUES (
  '00000000-0000-0000-0000-000000000102'::uuid,
  '00000000-0000-0000-0000-000000000101'::uuid,
  1,
  'published',
  'system/templates/blank.docx',
  'system-blank-template',
  '{}'::jsonb,
  '[]'::jsonb,
  'system',
  '',
  'system',
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  TIMESTAMPTZ '2026-05-14 20:40:29.86648+00',
  NULL,
  0
) ON CONFLICT (id) DO NOTHING;

UPDATE public.templates_v2_template
SET
  system_owned = true,
  latest_version = 1,
  published_version_id = '00000000-0000-0000-0000-000000000102'::uuid
WHERE id = '00000000-0000-0000-0000-000000000101'::uuid;

COMMIT;
