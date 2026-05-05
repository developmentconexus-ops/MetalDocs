-- 0174_documents_created_by_displayname_snapshot.sql
BEGIN;

ALTER TABLE public.documents
    ADD COLUMN IF NOT EXISTS created_by_display_name_snapshot TEXT;

UPDATE public.documents d
   SET created_by_display_name_snapshot = u.display_name
  FROM metaldocs.iam_users u
 WHERE u.user_id = d.created_by
   AND d.created_by_display_name_snapshot IS NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0174', 'add created_by_display_name_snapshot to documents')
ON CONFLICT (version) DO NOTHING;

COMMIT;
