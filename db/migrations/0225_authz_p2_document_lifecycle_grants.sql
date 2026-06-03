-- 0225_authz_p2_document_lifecycle_grants.sql
-- ADR 0022 Phase 2 — seed the routed-but-unseeded document-lifecycle write caps.
--
-- doc.publish / doc.obsolete / doc.supersede / template.archive were promoted to
-- typed consts in Phase 1 and are routed in apps/api/cmd/metaldocs-api/permissions.go,
-- but seeded to NO tenant role — enforced today only by the system_admin tier-2
-- bypass. That left them as always-403 for every non-system_admin. This migration
-- grants them to the tenant roles that already hold the parallel authority.
--
-- Grant matrix (ADR 0022 Phase 2 — derived from existing seed parallels, not guessed):
--   doc.obsolete    -> area_admin, qms_admin   (mirror controlled_documents.obsolete holders)
--   doc.supersede   -> area_admin, qms_admin   (mirror controlled_documents.supersede holders)
--   doc.publish     -> area_admin, qms_admin   (document-lifecycle authority set, consistent
--                                               with obsolete/supersede; product-confirmed
--                                               over the broader signoff-holder parallel)
--   template.archive-> qms_admin               (template steward; area_admin holds no
--                                               template.* caps, so excluded)
--   system_admin holds all four via the tier-2 bypass (not seeded explicitly).
--
-- Idempotent: ON CONFLICT (role, capability) DO NOTHING.
-- Forward-only per wiki/database/migration-policy.md.
-- Mirrored into db/reference-data/0001_product_reference_data.sql (curated baseline).
-- Does NOT modify permissions.go (Tier-1 routing already present from Phase 1).

BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('area_admin', 'doc.obsolete',     'Make a document obsolete'),
    ('qms_admin',  'doc.obsolete',     'Make a document obsolete'),
    ('area_admin', 'doc.supersede',    'Supersede a document with a successor'),
    ('qms_admin',  'doc.supersede',    'Supersede a document with a successor'),
    ('area_admin', 'doc.publish',      'Publish an approved document'),
    ('qms_admin',  'doc.publish',      'Publish an approved document'),
    ('qms_admin',  'template.archive', 'Archive a template')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0225', 'ADR 0022 Phase 2: seed document-lifecycle write caps doc.publish/doc.obsolete/doc.supersede/template.archive')
ON CONFLICT (version) DO NOTHING;

COMMIT;
