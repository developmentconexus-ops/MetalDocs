BEGIN;

DO $$
BEGIN
  IF to_regclass('public.ux_documents_v2_cd_active') IS NOT NULL
     AND to_regclass('public.ux_documents_cd_active') IS NULL THEN
    EXECUTE 'ALTER INDEX public.ux_documents_v2_cd_active RENAME TO ux_documents_cd_active';
  END IF;
END
$$;

DO $$
BEGIN
  IF to_regclass('public.ux_documents_v2_cd_revision') IS NOT NULL
     AND to_regclass('public.ux_documents_cd_revision') IS NULL THEN
    EXECUTE 'ALTER INDEX public.ux_documents_v2_cd_revision RENAME TO ux_documents_cd_revision';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_trigger t
    JOIN pg_class c ON c.oid = t.tgrelid
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relname = 'documents'
      AND t.tgname = 'trg_documents_v2_legal_transition'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_trigger t
    JOIN pg_class c ON c.oid = t.tgrelid
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relname = 'documents'
      AND t.tgname = 'trg_documents_legal_transition'
  ) THEN
    EXECUTE 'ALTER TRIGGER trg_documents_v2_legal_transition ON public.documents RENAME TO trg_documents_legal_transition';
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_trigger t
    JOIN pg_class c ON c.oid = t.tgrelid
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relname = 'documents'
      AND t.tgname = 'trg_documents_v2_revision_version_monotonic'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM pg_trigger t
    JOIN pg_class c ON c.oid = t.tgrelid
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relname = 'documents'
      AND t.tgname = 'trg_documents_revision_version_monotonic'
  ) THEN
    EXECUTE 'ALTER TRIGGER trg_documents_v2_revision_version_monotonic ON public.documents RENAME TO trg_documents_revision_version_monotonic';
  END IF;
END
$$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0202', 'rename documents object names to current naming')
ON CONFLICT (version) DO NOTHING;

COMMIT;
