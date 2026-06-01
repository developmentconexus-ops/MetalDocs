-- 0216_approval_stage_instances_skip_reason.sql
-- Restore the skip_reason column the approval code requires.
--
-- The curated baseline (2026-05-14) created approval_stage_instances WITHOUT
-- skip_reason, but the status CHECK already permits 'skipped' and the approval
-- code (domain.StageInstance.SkipReason, loadStageInstances SELECT) reads/writes
-- it. Without this column every signoff fails at stage load:
--   column "skip_reason" does not exist (SQLSTATE 42703)
BEGIN;

ALTER TABLE public.approval_stage_instances
    ADD COLUMN IF NOT EXISTS skip_reason TEXT;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0216', 'add approval_stage_instances.skip_reason column required by approval code')
ON CONFLICT (version) DO NOTHING;

COMMIT;
