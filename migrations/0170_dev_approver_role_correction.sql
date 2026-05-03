-- 0170_dev_approver_role_correction.sql
-- Rolls back migration 0166's blanket admin->system_admin rename for the dev
-- approver seed user (introduced by 0159). Restores SoD by ensuring approver
-- and admin-local are distinct roles in dev environments.
--
-- Idempotent: only flips rows that are still incorrectly system_admin.

UPDATE metaldocs.iam_user_roles
   SET role_code = 'approver'
 WHERE user_id = 'approver'
   AND role_code = 'system_admin';
