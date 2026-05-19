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

    -- Use a two-phase shift so existing lineages like REV01/REV02 do not
    -- transiently collide with the active unique index on
    -- (controlled_document_id, revision_number).
    UPDATE public.documents
       SET revision_number = revision_number + 1000000000
     WHERE controlled_document_id IS NOT NULL
       AND revision_number > 0;

    UPDATE public.documents
       SET revision_number = revision_number - 1000000001
     WHERE controlled_document_id IS NOT NULL
       AND revision_number >= 1000000001;

    UPDATE public.documents
       SET revision_number = revision_number - 1
     WHERE controlled_document_id IS NULL
       AND revision_number > 0;

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0205', 'make documents.revision_number zero based for governed REV labels');
  END IF;
END
$$;

COMMIT;
