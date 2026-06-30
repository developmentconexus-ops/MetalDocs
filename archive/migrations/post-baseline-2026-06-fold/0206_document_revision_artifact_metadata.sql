BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM public.schema_migrations
    WHERE version = '0206'
  ) THEN
    ALTER TABLE public.document_revisions
      ADD COLUMN file_size_bytes bigint,
      ADD COLUMN page_count integer,
      ADD COLUMN page_count_source text;

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_file_size_bytes_nonnegative
      CHECK (file_size_bytes IS NULL OR file_size_bytes >= 0);

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_page_count_positive
      CHECK (page_count IS NULL OR page_count > 0);

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_page_count_source_check
      CHECK (
        page_count_source IS NULL
        OR page_count_source IN ('eigenpal_client', 'server_renderer')
      );

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_page_count_provenance_coupling_check
      CHECK (
        (page_count IS NULL AND page_count_source IS NULL)
        OR (
          page_count IS NOT NULL
          AND page_count_source IN ('eigenpal_client', 'server_renderer')
        )
      );

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0206', 'add artifact metadata to document_revisions');
  END IF;
END
$$;

COMMIT;
