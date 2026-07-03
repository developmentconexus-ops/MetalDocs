-- 0263_validate_audit_hash_shape_checks.sql
-- DB-07 (P2) closure — with a scope correction found during investigation.
--
-- The original DB-07 finding (approval-tech-debt.md T-009) cites the archived
-- migration 0135's NOT VALID FKs on approval_instances.submitted_by and
-- approval_signoffs.actor_user_id. That claim is STALE register drift: both
-- constraints were superseded by composite tenant-scoped FKs
-- (approval_instances_submitted_by_tenant_fkey, approval_signoffs_actor_tenant_fkey
-- → metaldocs.iam_users(tenant_id, user_id)) which are fully validated today —
-- verified on the live DB (pg_constraint.convalidated = true for both) and in the
-- canonical baseline (db/baseline/0001_current_schema.sql:4145,4169 carries them
-- WITHOUT the NOT VALID marker pg_dump prints for unvalidated constraints).
-- No approval work needed.
--
-- The REAL remaining members of the "declared NOT VALID, never validated" class —
-- the only two in the entire schema (SELECT ... FROM pg_constraint WHERE NOT
-- convalidated) — are the audit hash-shape CHECKs added by the 0237-era audit
-- hardening:
--   metaldocs.audit_events.audit_events_prev_hash_shape
--   metaldocs.audit_events.audit_events_row_hash_shape
-- Both assert hash columns are '' or 64-hex. Local verification: 0 of 1047 rows
-- violate. VALIDATE CONSTRAINT takes only SHARE UPDATE EXCLUSIVE (does not block
-- reads/writes) and fails loudly if any environment has a malformed hash —
-- which would also indicate hash-chain corruption worth failing a deploy over.
-- Idempotent: validating an already-validated constraint is a no-op.

BEGIN;

ALTER TABLE metaldocs.audit_events VALIDATE CONSTRAINT audit_events_prev_hash_shape;
ALTER TABLE metaldocs.audit_events VALIDATE CONSTRAINT audit_events_row_hash_shape;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0263', 'DB-07: validate the last NOT VALID constraints — audit_events prev_hash/row_hash shape CHECKs (approval iam_users FKs from the original finding were already superseded by validated composite tenant FKs; register drift)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
