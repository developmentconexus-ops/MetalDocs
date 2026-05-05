-- 0170_dev_approver_role_correction.sql
-- Rolls back migration 0166's blanket admin->system_admin rename for the dev
-- approver seed user (introduced by 0159). Restores SoD by ensuring approver
-- and admin-local are distinct roles in dev environments.
--
-- Idempotent: only flips rows that are still incorrectly system_admin.

BEGIN;

UPDATE metaldocs.iam_user_roles
   SET role_code = 'approver'
 WHERE user_id = 'approver'
   AND role_code = 'system_admin';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0170', 'restore dev approver role to approver after 0166 blanket rename')
ON CONFLICT (version) DO NOTHING;

COMMIT;
