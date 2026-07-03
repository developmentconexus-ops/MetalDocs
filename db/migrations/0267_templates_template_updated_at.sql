-- 0267_templates_template_updated_at.sql
-- FE-08 (P3): add updated_at to templates_template.
--
-- Runtime-truth check performed before writing this migration: grepped
-- internal/modules/templates/repository/postgres.go and
-- db/baseline/0001_current_schema.sql for templates_template's columns —
-- there is NO updated_at column and NO auto-touch trigger convention
-- anywhere in this schema (grepped for trigger_set_timestamp / moddatetime /
-- "BEFORE UPDATE" across the baseline; every updated_at column in this
-- schema, e.g. metaldocs.template_drafts, is set explicitly by the
-- application on each UPDATE, not by a DB trigger). The wiki backlog row
-- (wiki/backlog/templates.md) assumed the column already existed and the API
-- just needed to surface it; that assumption was false. This migration adds
-- the column so the FE-08 fix is a real, additive capability instead of a
-- surfaced no-op.
--
-- Column: updated_at timestamptz NOT NULL DEFAULT now(). Backfilled to
-- created_at for existing rows (best-known-truth: a row that has never been
-- updated was last "updated" at creation). New INSERTs get DEFAULT now()
-- (repository INSERT does not set it explicitly, matching created_at's
-- pattern of relying on DEFAULT). Every application UPDATE path
-- (UpdateTemplate / UpdateTemplateTx in
-- internal/modules/templates/repository/postgres.go) is updated in the same
-- change to SET updated_at = now(), so the column reflects real mutations
-- going forward — this migration only adds the column; the app-side SET is
-- shipped in the same commit, not deferred.
--
-- Safety class: DEFAULT now() is VOLATILE, so this ADD COLUMN does NOT
-- qualify for the Postgres 11+ metadata-only fast path (that applies to
-- constant-foldable defaults only) — it takes ACCESS EXCLUSIVE and rewrites
-- the table, and the backfill UPDATE below rewrites every row again anyway.
-- Acceptable here because templates_template is small (one row per template,
-- not per document/version): a full rewrite is cheap and the lock is brief.
--
-- Rollback note: `ALTER TABLE public.templates_template DROP COLUMN
-- updated_at;` — safe at any time, no dependents created by this migration.

BEGIN;

ALTER TABLE public.templates_template
  ADD COLUMN updated_at timestamp with time zone NOT NULL DEFAULT now();

-- Backfill existing rows to created_at (best-known "last touched" value).
UPDATE public.templates_template
   SET updated_at = created_at;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0267', 'FE-08 (P3): add templates_template.updated_at (timestamptz NOT NULL DEFAULT now(), backfilled to created_at). No DB-level auto-touch trigger convention exists in this schema, so application UPDATE paths in internal/modules/templates/repository/postgres.go explicitly SET updated_at = now() (shipped in the same change). Closes the wiki/backlog/templates.md TemplateDTO updated_at gap.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
