-- 0292_approval_stage_sla.sql
-- M2b F8 (approval-kernel-backend, approval remediation design spec.md §4/§6.3):
-- purely additive schema expansion wiring the F1-reserved SLA substrate
-- (approval_stage_instances.due_at) to an actual write path. F1 (migration
-- 0286) added approval_route_stages.due_in_days and
-- approval_stage_instances.due_at, both unpopulated/unused until this
-- feature. F8 needs one more column: a per-stage-instance SNAPSHOT of the
-- route stage's due_in_days config, captured at submit time — mirroring the
-- *_snapshot convention already used by this table (name_snapshot,
-- required_role_snapshot, stage_kind, etc.) — so the stage-activation UPDATE
-- that computes due_at never needs a second read against approval_route_stages
-- (which may have moved to a superseded route version by the time a later
-- stage activates, per F2 route versioning; the snapshot is what stays pinned
-- for the life of the instance, exactly like every other *_snapshot field).
--
-- Also adds approval_stage_instances.sla_surfaced_at: the idempotency marker
-- for the new approval_sla_surfacer periodic job (F8), mirroring
-- documents.review_surfaced_at's shape but on a DIFFERENT table for a
-- DIFFERENT concept (per-stage-instance approval SLA vs per-document periodic
-- eQMS review) — deliberately NOT reusing documents.review_surfaced_at, which
-- would conflate two unrelated due-date mechanisms on one column (see F8
-- spec.md Interview #2).
--
-- Expand-only, forward-only, idempotent (ADD COLUMN IF NOT EXISTS; DROP
-- CONSTRAINT IF EXISTS before ADD CONSTRAINT — idiom copied from 0286).
--
-- New columns:
--   public.approval_stage_instances.due_in_days_snapshot integer NULL
--     CHECK (due_in_days_snapshot IS NULL OR due_in_days_snapshot > 0)
--     -- snapshot of approval_route_stages.due_in_days at submit time.
--   public.approval_stage_instances.sla_surfaced_at timestamp with time zone NULL
--     -- idempotency marker: the approval_sla_surfacer job stamps this the
--     -- first time it surfaces an overdue stage instance for its CURRENT
--     -- due_at cycle; NULL means not (yet) surfaced. Alert-only (ADR 0068) --
--     -- no state mutation on status/due_at itself.

BEGIN;

-- approval_stage_instances ----------------------------------------------------

ALTER TABLE public.approval_stage_instances
    ADD COLUMN IF NOT EXISTS due_in_days_snapshot integer,
    ADD COLUMN IF NOT EXISTS sla_surfaced_at timestamp with time zone;

ALTER TABLE public.approval_stage_instances
    DROP CONSTRAINT IF EXISTS approval_stage_instances_due_in_days_snapshot_check;
ALTER TABLE public.approval_stage_instances
    ADD CONSTRAINT approval_stage_instances_due_in_days_snapshot_check
    CHECK (due_in_days_snapshot IS NULL OR due_in_days_snapshot > 0);

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0292', 'M2b F8: SLA snapshot + surfacer substrate. Adds approval_stage_instances.due_in_days_snapshot (per-instance snapshot of the route stage''s due_in_days config, read at stage-activation time to compute due_at without a second read against a possibly-superseded route version) and approval_stage_instances.sla_surfaced_at (idempotency marker for the new approval_sla_surfacer periodic job, distinct from and NOT reusing documents.review_surfaced_at). No existing row or code path changes behavior until F8''s service/repository/job wiring populates these columns.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
