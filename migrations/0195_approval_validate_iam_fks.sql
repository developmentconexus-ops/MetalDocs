BEGIN;

ALTER TABLE approval_instances
  VALIDATE CONSTRAINT approval_instances_submitted_by_tenant_fkey;

ALTER TABLE approval_signoffs
  VALIDATE CONSTRAINT approval_signoffs_actor_tenant_fkey;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0195', 'validate approval IAM foreign keys')
ON CONFLICT (version) DO NOTHING;

COMMIT;
