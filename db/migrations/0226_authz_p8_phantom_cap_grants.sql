-- 0226_authz_p8_phantom_cap_grants.sql
-- ADR 0022 Phase 8 — seed the phantom capabilities that Phase 8 registered.
--
-- doc.edit_draft / doc.reconstruct / workflow.instance.cancel / doc.view_published
-- were enforced at tier-2 by raw string but were absent from the registry
-- (validCapabilities) — seeded to NO role, so they passed only via the
-- system_admin tier-2 bypass (dead-grant checks for every other role). Phase 8
-- promotes them to typed consts; this migration seeds them so the tier-2 check
-- agrees with the tier-1 route authority each cap already shadows.
--
-- Grant matrix (derived from existing seed parallels, not guessed — each phantom
-- mirrors the role set of the tier-1 cap that routes its HTTP surface in
-- apps/api/cmd/metaldocs-api/permissions.go):
--   doc.edit_draft           -> document.edit holders  (tier-1 /session,/autosave = document.edit)
--   doc.reconstruct          -> document.edit holders  (tier-1 /reconstruct       = document.edit)
--   workflow.instance.cancel -> document.edit holders  (tier-1 /cancel            = document.edit)
--   doc.view_published       -> document.view holders  (tier-1 GET /documents     = document.view)
-- document.edit set: approver, area_admin, author, editor, qms_admin.
-- document.view set: the document.edit set + signer + viewer.
-- system_admin holds all four via the tier-2 bypass (not seeded explicitly,
-- consistent with migration 0225).
--
-- route.admin is NOT seeded here: it reconciles to the EXISTING route.manage cap
-- (already seeded to qms_admin + system_admin) — Phase 8 only retyped the call
-- sites; the grant matrix is unchanged.
--
-- Idempotent: ON CONFLICT (role, capability) DO NOTHING.
-- Forward-only per wiki/database/migration-policy.md.
-- Mirrored into db/reference-data/0001_product_reference_data.sql (curated baseline).
-- Does NOT modify permissions.go (tier-1 routing already present).

BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
    ('approver',   'doc.edit_draft',           'Edit a document draft'),
    ('area_admin', 'doc.edit_draft',           'Edit a document draft'),
    ('author',     'doc.edit_draft',           'Edit a document draft'),
    ('editor',     'doc.edit_draft',           'Edit a document draft'),
    ('qms_admin',  'doc.edit_draft',           'Edit a document draft'),
    ('approver',   'doc.reconstruct',          'Reconstruct a document from a revision'),
    ('area_admin', 'doc.reconstruct',          'Reconstruct a document from a revision'),
    ('author',     'doc.reconstruct',          'Reconstruct a document from a revision'),
    ('editor',     'doc.reconstruct',          'Reconstruct a document from a revision'),
    ('qms_admin',  'doc.reconstruct',          'Reconstruct a document from a revision'),
    ('approver',   'workflow.instance.cancel', 'Cancel a workflow instance'),
    ('area_admin', 'workflow.instance.cancel', 'Cancel a workflow instance'),
    ('author',     'workflow.instance.cancel', 'Cancel a workflow instance'),
    ('editor',     'workflow.instance.cancel', 'Cancel a workflow instance'),
    ('qms_admin',  'workflow.instance.cancel', 'Cancel a workflow instance'),
    ('approver',   'doc.view_published',       'View a published document'),
    ('area_admin', 'doc.view_published',       'View a published document'),
    ('author',     'doc.view_published',       'View a published document'),
    ('editor',     'doc.view_published',       'View a published document'),
    ('qms_admin',  'doc.view_published',       'View a published document'),
    ('signer',     'doc.view_published',       'View a published document'),
    ('viewer',     'doc.view_published',       'View a published document')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0226', 'ADR 0022 Phase 8: seed registered phantom caps doc.edit_draft/doc.reconstruct/workflow.instance.cancel/doc.view_published')
ON CONFLICT (version) DO NOTHING;

COMMIT;
