BEGIN;

ALTER TABLE approval_instances
  VALIDATE CONSTRAINT approval_instances_submitted_by_tenant_fkey;

ALTER TABLE approval_signoffs
  VALIDATE CONSTRAINT approval_signoffs_actor_tenant_fkey;

COMMIT;
