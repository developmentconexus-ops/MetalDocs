-- migrations/0208_templates_revision_number.sql
-- ADR 0013 — Template Revision Labels (backend-canonical).
-- Adds a 0-based regulated-revision counter to templates_template_version,
-- mirroring documents.revision_number (migrations/0131_documents_v2_state_columns.sql).
-- Backfill: revision_number = version_number - 1 (templates lifecycle has no
-- version_number gaps today, so this preserves the existing display mapping).
-- INSERT-side allocation handled in repository (mirror documents pattern).

BEGIN;

ALTER TABLE templates_template_version
  ADD COLUMN IF NOT EXISTS revision_number INT NOT NULL DEFAULT 0;

-- One-shot backfill. Only touches rows that were inserted before this migration
-- with the column at its default value 0. Subsequent INSERTs supply the value
-- explicitly via COALESCE(MAX(rn)+1, 0).
UPDATE templates_template_version
   SET revision_number = version_number - 1
 WHERE revision_number = 0
   AND version_number > 1;

CREATE UNIQUE INDEX IF NOT EXISTS ux_templates_version_revision
  ON templates_template_version (template_id, revision_number);

COMMIT;
