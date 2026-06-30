BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM public.schema_migrations
    WHERE version = '0233'
  ) THEN
    -- ADR 0013 (D-1): persist revision_number on templates_template_version,
    -- mirroring documents.revision_number. The Go layer (CreateVersion INSERT,
    -- Get/List SELECT projection, scanner, domain, api.gen) already expects this
    -- column; only the schema half of ADR 0013 was missing, which made every
    -- templates list/fetch fail with "column lv.revision_number does not exist".
    ALTER TABLE public.templates_template_version
      ADD COLUMN IF NOT EXISTS revision_number integer NOT NULL DEFAULT 0;

    -- Backfill is a row UPDATE on templates_template_version, which trips the
    -- enforce_capability_asserted() guard (it gates this table on the template.*
    -- capability family). Assert a template capability for the duration of this
    -- transaction, matching the pattern in 0205.
    PERFORM set_config(
      'metaldocs.asserted_caps',
      '[{"cap":"template.edit","area":"tenant"}]',
      true
    );

    -- One-shot backfill: revision_number is the 0-based regulated counter, and
    -- by ADR 0013 invariant revision_number = version_number - 1 for every row
    -- (no version_number gaps in the templates lifecycle today).
    UPDATE public.templates_template_version
       SET revision_number = version_number - 1;

    -- Per-template uniqueness, mirroring ux_documents_cd_revision.
    CREATE UNIQUE INDEX IF NOT EXISTS ux_templates_version_revision
      ON public.templates_template_version (template_id, revision_number);

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0233', 'add templates_template_version.revision_number (ADR 0013 schema half)');
  END IF;
END
$$;

COMMIT;
