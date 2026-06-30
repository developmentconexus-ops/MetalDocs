-- db/migrations/0240_drop_legacy_mddm_document_cluster.sql
-- Drops the dead legacy MetalDocs Document Model (MDDM) cluster from the
-- metaldocs schema. These early editor-era tables are unused (zero runtime Go
-- references) and the metaldocs.documents / metaldocs.template_audit_log copies
-- SHADOW the live public.documents / public.template_audit_log governance tables
-- whenever the connection search_path is metaldocs-first. Removing them makes
-- bare unqualified `documents` / `template_audit_log` resolve to the real
-- public.* table under ANY search_path ordering.
--
-- Verified dead by F4b.1 census (10 objects, all inbound FKs internal to the
-- cluster, no view/trigger/RLS/function/sequence/seed dependent):
--   milestone-4b-legacy-schema-teardown/f4b.1-legacy-cluster-census/evidence.md
-- Decision + data-loss / rollback / maintenance-window posture:
--   wiki/decisions/0032-drop-legacy-mddm-document-cluster.md
-- Precedent: 0231_db_hardening_tripwire_and_dead_schema.sql, 0236_dead_schema_drop.sql.
--
-- DESTRUCTIVE: any historical rows in these tables are abandoned. Believed
-- empty/dead in production; dump before applying to prod if forensic value is
-- suspected. Apply in a maintenance window (HS-1 operator gate).
--
-- Drop order: satellites before anchor. CASCADE removes attached indexes, FK
-- constraints, and owned sequences; every inbound FK originates inside this set,
-- so no kept table is affected. IF EXISTS makes the migration idempotent.

BEGIN;

-- 2nd-level satellite (child of document_versions_mddm) — drop first.
DROP TABLE IF EXISTS metaldocs.document_version_images CASCADE;

-- 1st-level satellites of metaldocs.documents.
DROP TABLE IF EXISTS metaldocs.document_attachments CASCADE;
DROP TABLE IF EXISTS metaldocs.document_collaboration_presence CASCADE;
DROP TABLE IF EXISTS metaldocs.document_edit_locks CASCADE;
DROP TABLE IF EXISTS metaldocs.document_template_assignments CASCADE;
DROP TABLE IF EXISTS metaldocs.document_versions CASCADE;
DROP TABLE IF EXISTS metaldocs.document_versions_mddm CASCADE;
DROP TABLE IF EXISTS metaldocs.workflow_approvals CASCADE;

-- Anchor of the legacy MDDM document model.
DROP TABLE IF EXISTS metaldocs.documents CASCADE;

-- Independent dead duplicate of public.template_audit_log.
DROP TABLE IF EXISTS metaldocs.template_audit_log CASCADE;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0240', 'drop legacy MDDM document cluster (metaldocs.documents + 8 satellites + metaldocs.template_audit_log); removes public.* shadow under metaldocs-first search_path')
ON CONFLICT (version) DO NOTHING;

COMMIT;
