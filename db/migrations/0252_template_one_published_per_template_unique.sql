-- One-published-per-template invariant (ADR 0013). Forward-only, idempotent.
--
-- F-DB2 (HIGH): there is no DB-level guard preventing two concurrent publishes
-- from both racing past ObsoletePreviousPublishedTx and writing status='published'
-- for the same template_id.  Under READ COMMITTED isolation both can read
-- zero existing 'published' rows and both proceed, corrupting published_version_id
-- on the parent templates_template row.  The CAS lock does not cover this
-- cross-row invariant.
--
-- FIX: a partial-unique index on (template_id) WHERE status='published' forces
-- the DB to reject the second concurrent INSERT/UPDATE with a unique-violation,
-- closing the window regardless of isolation level.
--
-- SIBLING PATTERN: mirrors ux_approval_instances_active_document_id
-- (db/baseline/0001_current_schema.sql ~line 3625) and
-- idx_one_active_draft_per_doc (~line 3289) -- partial-unique active-row
-- enforcement; also mirrors ux_documents_cd_revision (~line 3639).
--
-- NOTE: the per-template revision uniqueness index (template_id, revision_number)
-- was added by migration 0233 and is NOT recreated here.
--
-- UNIQUE-INDEX FAILURE CLASS: if this migration fails with a unique-violation,
-- existing data already has two 'published' rows for the same template_id --
-- pre-existing corruption the index surfaces.  Audit and resolve before
-- re-running; re-running does NOT auto-fix this.

BEGIN;

CREATE UNIQUE INDEX IF NOT EXISTS uq_one_published_per_template
    ON public.templates_template_version (template_id)
    WHERE status = 'published';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0252', 'one published version per template partial-unique index (ADR 0013 / F-DB2)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
