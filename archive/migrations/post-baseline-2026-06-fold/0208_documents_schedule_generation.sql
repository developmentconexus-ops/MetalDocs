BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM public.schema_migrations
    WHERE version = '0208'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'documents'
        AND column_name = 'schedule_generation'
    ) THEN
      ALTER TABLE public.documents
        ADD COLUMN schedule_generation bigint NOT NULL DEFAULT 0;
    END IF;

    COMMENT ON COLUMN public.documents.schedule_generation IS
      'Monotonic generation used to invalidate stale scheduled publish jobs after reschedule or cancel.';

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0208', 'add schedule_generation discriminator for scheduled publish jobs');
  END IF;
END
$$;

COMMIT;
