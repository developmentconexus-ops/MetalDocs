-- 0227_authz_p10_merge_redundant_caps.sql
-- ADR 0022 Phase 10 (F2) — merge the four redundant phantom caps Phase 8 minted
-- back into their canonical equivalents (registry 33 -> 29; area-grade 14 -> 11).
--
-- Phase 8 (migration 0226) registered + seeded four caps whose grant sets and HTTP
-- surfaces are IDENTICAL to existing caps — zero authorization differentiation:
--   doc.edit_draft           == document.edit  (tier-1 /session,/autosave = document.edit)
--   doc.reconstruct          == document.edit  (tier-1 /reconstruct       = document.edit)
--   workflow.instance.cancel == document.edit  (tier-1 /cancel            = document.edit)
--   doc.view_published       == document.view  (tier-1 GET /documents     = document.view)
-- The audit (2026-06-04) confirmed each phantom's seeded role set equals its
-- canonical sibling's, so collapsing them changes NO effective authorization.
-- Their tier-2 call sites now Require document.edit / document.view (see
-- internal/modules/documents/.../{fillin_authz,reconstruct_service,view_service}.go
-- and internal/modules/documents/approval/application/cancel_service.go).
--
-- This REVERSES Phase 8's registry over-modeling, NOT its security fix: the raw
-- phantom dead-grants are gone and the surfaces stay enforced by the canonical cap.
--
-- Forward-only per wiki/database/migration-policy.md (a reverse-of-data migration,
-- not a down migration). Idempotent: DELETE is naturally idempotent.
-- Mirrors the removal of the 22 grant rows from the curated baseline
-- db/reference-data/0001_product_reference_data.sql.
-- Does NOT modify permissions.go (tier-1 routing already mapped these surfaces to
-- document.view / document.edit; nothing to reroute there).

BEGIN;

DELETE FROM metaldocs.role_capabilities
 WHERE capability IN (
    'doc.edit_draft',
    'doc.reconstruct',
    'workflow.instance.cancel',
    'doc.view_published'
 );

INSERT INTO public.schema_migrations (version, description)
VALUES ('0227', 'ADR 0022 Phase 10: merge redundant phantom caps doc.edit_draft/doc.reconstruct/workflow.instance.cancel/doc.view_published into document.edit/document.view')
ON CONFLICT (version) DO NOTHING;

COMMIT;
