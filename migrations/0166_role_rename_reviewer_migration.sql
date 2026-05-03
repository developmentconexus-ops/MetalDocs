BEGIN;

ALTER TABLE metaldocs.iam_user_roles
  DROP CONSTRAINT IF EXISTS chk_iam_user_roles_role_code;

UPDATE metaldocs.iam_user_roles
   SET role_code = 'system_admin'
 WHERE role_code = 'admin';

UPDATE metaldocs.iam_user_roles
   SET role_code = 'approver'
 WHERE role_code = 'reviewer';

DELETE FROM metaldocs.iam_user_roles a
USING metaldocs.iam_user_roles b
WHERE a.ctid < b.ctid
  AND a.tenant_id = b.tenant_id
  AND a.user_id = b.user_id;

ALTER TABLE metaldocs.iam_user_roles
  ADD CONSTRAINT chk_iam_user_roles_role_code
  CHECK (role_code IN ('system_admin', 'approver', 'author', 'editor', 'viewer'));

ALTER TABLE metaldocs.iam_user_roles
  DROP CONSTRAINT IF EXISTS uq_iam_user_roles_user_tenant;

ALTER TABLE metaldocs.iam_user_roles
  ADD CONSTRAINT uq_iam_user_roles_user_tenant UNIQUE (tenant_id, user_id);

COMMIT;
