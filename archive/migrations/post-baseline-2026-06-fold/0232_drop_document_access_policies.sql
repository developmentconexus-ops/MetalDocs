-- 0232_drop_document_access_policies.sql
-- Remove the document_access_policies table — the DB half of a pre-unification
-- per-document ACL slice that the ADR 0022 authz refactor superseded.
--
-- WHY (supersedes the "deliberately KEPT" note in migration 0231):
-- 0231 deferred this drop because ADR 0022 Phase 5 had tied the table's CHECK
-- to the OpenAPI AccessPolicyItem/AccessPolicyWriteItem vocabulary + FE types
-- and chose not to prune (a contract-shape change). A focused validation pass
-- then established the whole slice is a dead pre-unification vestige:
--   * Introduced 2026-03-17 (old migration 0011), 11 weeks BEFORE the ADR 0022
--     unification — part of the per-document ACL approach the refactor replaced.
--   * Zero Go reads/writes; the /access-policies route had NO handler (404);
--     the search ListAccessPolicies port is a nil (open-by-default) stub.
--   * Not even attached to the enforce_capability_asserted tripwire — off the
--     active write path entirely.
--   * The unified model (tier-1 route->capability, tier-2 authz.Require with
--     area scope, role_capabilities) has no per-document ACL concept. The live
--     per-resource mechanism is controlled_document_{area,user}_grants, which is
--     a different, narrower feature — it does not depend on this table.
--   * A published-but-unserved /access-policies endpoint is an OWASP API9:2023
--     (improper inventory) liability and a second, incoherent authz source of
--     truth — the exact thing the unification removed.
--
-- The whole slice is removed together in this change: the OpenAPI path + the
-- five AccessPolicy* schemas, the regenerated FE types, the two permissions.go
-- route rows + their tests, and now the table. If per-document ACL is ever a
-- real requirement it should be designed as part of the unified model, not
-- revived from this orphaned table.
--
-- CASCADE removes the owned sequence, the id DEFAULT, the primary key, the
-- uq_document_access_policies_rule unique constraint, and both indexes.
--
-- Forward-only, idempotent. Mirrored inline in the curated baseline
-- (db/baseline/0001_current_schema.sql).

BEGIN;

DROP TABLE IF EXISTS metaldocs.document_access_policies CASCADE;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0232', 'Drop document_access_policies: dead pre-unification per-document ACL table removed with its OpenAPI/FE/permissions slice (ADR 0022 unified model has no per-document ACL concept)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
