BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM public.schema_migrations
    WHERE version = '0207'
  ) THEN
    -- Normalize rows created before coupling enforcement.
    UPDATE public.document_revisions
       SET page_count_source = NULL
     WHERE page_count IS NULL
       AND page_count_source IS NOT NULL;

    UPDATE public.document_revisions
       SET page_count = NULL
     WHERE page_count_source IS NULL
       AND page_count IS NOT NULL;

    IF NOT EXISTS (
      SELECT 1
      FROM pg_constraint
      WHERE conname = 'document_revisions_page_count_provenance_coupling_check'
        AND conrelid = 'public.document_revisions'::regclass
    ) THEN
      ALTER TABLE public.document_revisions
        ADD CONSTRAINT document_revisions_page_count_provenance_coupling_check
        CHECK (
          (page_count IS NULL AND page_count_source IS NULL)
          OR (
            page_count IS NOT NULL
            AND page_count_source IN ('eigenpal_client', 'server_renderer')
          )
        );
    END IF;

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0207', 'enforce page_count provenance coupling on document_revisions');
  END IF;
END
$$;

COMMIT;
