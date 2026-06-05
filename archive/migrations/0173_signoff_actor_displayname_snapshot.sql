-- 0173_signoff_actor_displayname_snapshot.sql
-- Adds actor_display_name_snapshot to approval_signoffs for QMS-grade attribution.
-- Backfills from current iam_users.display_name (best-effort historical).
BEGIN;

ALTER TABLE public.approval_signoffs
    ADD COLUMN IF NOT EXISTS actor_display_name_snapshot TEXT;

UPDATE public.approval_signoffs s
   SET actor_display_name_snapshot = u.display_name
  FROM metaldocs.iam_users u
 WHERE u.user_id = s.actor_user_id
   AND s.actor_display_name_snapshot IS NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0173', 'add actor_display_name_snapshot to approval_signoffs')
ON CONFLICT (version) DO NOTHING;

COMMIT;
