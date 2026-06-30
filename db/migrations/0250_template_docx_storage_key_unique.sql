-- Template docx storage-key integrity (Phase 1). Forward-only, idempotent.
--
-- (1) Re-key any NON-published version row whose docx_storage_key collides with
--     another row, to its canonical positional key
--     tenants/{tenant_id}/templates/{template_id}/versions/{version_number}.docx.
--     Published rows are never re-keyed -- they own their immutable object.
--     tenant_id lives on the parent public.templates_template (the version table
--     has no tenant_id), so the re-key joins to it.
-- (2) Add UNIQUE(docx_storage_key) so a shared key can never recur
--     ("DB enforces invariants"). Friendly first line is the app de-sharing in
--     internal/modules/templates/application; this is the last line.
--
-- OBJECT-STORE NOTE (operational, NOT run here): a re-keyed draft row now points
-- at a key with no object yet; opening it before first autosave 404s (D2-class:
-- the FE renders "open empty"). If a draft's object had DIVERGED from the source
-- (corruption already occurred), copying/repairing bytes is a manual operator
-- step -- SQL cannot move MinIO objects. The affected set is expected to be
-- empty/small (feature is pre-v1). If two PUBLISHED rows share a key (a publish-
-- injection artifact), the UNIQUE index creation below fails loudly -- audit and
-- resolve those before re-running.

BEGIN;

UPDATE public.templates_template_version v
SET docx_storage_key =
        'tenants/' || tt.tenant_id
        || '/templates/' || v.template_id::text
        || '/versions/' || v.version_number || '.docx'
FROM public.templates_template tt
WHERE tt.id = v.template_id
  AND v.status <> 'published'
  AND EXISTS (
        SELECT 1
        FROM public.templates_template_version o
        WHERE o.docx_storage_key = v.docx_storage_key
          AND o.id <> v.id
      );

CREATE UNIQUE INDEX IF NOT EXISTS uq_templates_template_version_docx_storage_key
    ON public.templates_template_version (docx_storage_key);

INSERT INTO public.schema_migrations (version, description)
VALUES ('0250', 'template version docx_storage_key de-share + UNIQUE (Phase 1)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
