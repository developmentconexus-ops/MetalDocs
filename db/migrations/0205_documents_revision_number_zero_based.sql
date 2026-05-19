BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM public.schema_migrations
    WHERE version = '0205'
  ) THEN
    ALTER TABLE public.documents
      ALTER COLUMN revision_number SET DEFAULT 0;

    -- One-time data normalization: the persisted governed revision number is
    -- now the REV suffix itself (0 => REV00, 1 => REV01), not a display offset.
    PERFORM set_config(
      'metaldocs.asserted_caps',
      '[{"cap":"document.edit","area":"tenant"}]',
      true
    );

    UPDATE public.documents
       SET revision_number = revision_number - 1
     WHERE revision_number > 0;

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0205', 'make documents.revision_number zero based for governed REV labels');
  END IF;
END
$$;

COMMIT;
