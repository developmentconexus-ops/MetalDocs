-- 0229_authz_p12_rename_document_lifecycle_caps.sql
-- ADR 0022 Phase 12 (F3) — rename the document-lifecycle capability VALUES from the
-- legacy doc.* spelling to the document.* prefix used by every sibling document cap
-- (document.view / .create / .edit / .submit / .signoff). This gives the document
-- aggregate a single resource.action prefix; the Go consts (CapDocumentPublish/...) and
-- their string VALUES, the curated seed (db/reference-data/0001), and the registry
-- (validCapabilities) all move together this phase.
--
--   doc.publish   -> document.publish
--   doc.obsolete  -> document.obsolete
--   doc.supersede -> document.supersede
--
-- Collision-safe + idempotent by construction. A bare
--   UPDATE role_capabilities SET capability='document.publish' WHERE capability='doc.publish'
-- would violate the (role, capability) primary key on a freshly bootstrapped DB: the
-- curated reference-data already seeds document.publish, while the replayed historical
-- migration 0225 still seeds doc.publish, so BOTH spellings coexist before this tail
-- migration runs. The INSERT ... SELECT ... ON CONFLICT DO NOTHING + DELETE pattern
-- converges to the renamed value whether the target row pre-exists (fresh bootstrap)
-- or not (an older DB holding only doc.*), and is a clean no-op on re-run.
--
-- Forward-only per wiki/database/migration-policy.md. Does NOT patch the historical
-- 0225 seed (policy: no historical-migration backfill); this tail converges its rows.
-- system_admin is unaffected (holds these via the tier-2 bypass, never seeded).

BEGIN;

-- doc.publish -> document.publish
INSERT INTO metaldocs.role_capabilities (role, capability, description)
SELECT role, 'document.publish', description
  FROM metaldocs.role_capabilities
 WHERE capability = 'doc.publish'
ON CONFLICT (role, capability) DO NOTHING;
DELETE FROM metaldocs.role_capabilities WHERE capability = 'doc.publish';

-- doc.obsolete -> document.obsolete
INSERT INTO metaldocs.role_capabilities (role, capability, description)
SELECT role, 'document.obsolete', description
  FROM metaldocs.role_capabilities
 WHERE capability = 'doc.obsolete'
ON CONFLICT (role, capability) DO NOTHING;
DELETE FROM metaldocs.role_capabilities WHERE capability = 'doc.obsolete';

-- doc.supersede -> document.supersede
INSERT INTO metaldocs.role_capabilities (role, capability, description)
SELECT role, 'document.supersede', description
  FROM metaldocs.role_capabilities
 WHERE capability = 'doc.supersede'
ON CONFLICT (role, capability) DO NOTHING;
DELETE FROM metaldocs.role_capabilities WHERE capability = 'doc.supersede';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0229', 'ADR 0022 Phase 12 (F3): rename document-lifecycle caps doc.publish/doc.obsolete/doc.supersede -> document.*')
ON CONFLICT (version) DO NOTHING;

COMMIT;
