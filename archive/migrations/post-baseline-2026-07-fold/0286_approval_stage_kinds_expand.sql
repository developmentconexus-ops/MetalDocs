-- 0286_approval_stage_kinds_expand.sql
-- M2b F1 (approval-kernel-backend, approval remediation design spec.md/plan.md):
-- purely additive schema expansion introducing the StageKind concept
-- ("review" vs "approval") plus a handful of nullable columns the later F2-F8
-- features will populate. NO behavior change anywhere in this migration —
-- stage_kind always defaults to 'approval', so every existing route/instance
-- row (and every existing code path reading these tables) continues to
-- behave exactly as it did before this migration. Wiring stage-kind
-- semantics into services (review vs. binding-approval gating, due-date
-- surfacing, freeze/hash-chain, cancel-reason UX) is explicitly OUT OF SCOPE
-- here — see F2-F8.
--
-- Expand-only, forward-only, idempotent (ADD COLUMN IF NOT EXISTS; DROP
-- CONSTRAINT IF EXISTS before ADD CONSTRAINT — idiom copied from
-- 0274_document_review_and_reason.sql / 0257_templates_version_status_under_review.sql).
--
-- New columns:
--   public.approval_route_stages.stage_kind    text NOT NULL DEFAULT 'approval'
--     CHECK IN ('review','approval') -- route-level stage-kind configuration.
--   public.approval_route_stages.due_in_days   integer NULL
--     CHECK (due_in_days IS NULL OR due_in_days > 0) -- optional per-stage SLA,
--     unused until F5/F8 wire due-date derivation.
--   public.approval_stage_instances.stage_kind text NOT NULL DEFAULT 'approval'
--     CHECK IN ('review','approval') -- snapshot of the route stage's kind at
--     submit time, mirroring the *_snapshot convention already used by this
--     table (name_snapshot, required_role_snapshot, etc).
--   public.approval_stage_instances.due_at     timestamptz NULL -- derived due
--     date for this stage instance; NULL until F5/F8 populate it.
--   public.approval_instances.frozen_content_hash text NULL
--     CHECK (frozen_content_hash IS NULL OR frozen_content_hash ~ '^[0-9a-f]{64}$')
--     -- reserved for the future content-freeze/hash-chain feature (OUT OF
--     SCOPE here: no freeze/hash-chain logic is added by this migration).
--   public.approval_instances.cancel_reason  text NULL -- reserved for a
--     future cancel-with-reason UX; unpopulated until that feature lands.
--   public.approval_signoffs.signature_meaning text NOT NULL DEFAULT 'approval'
--     CHECK IN ('approval','rejection') -- meaning of the recorded signature;
--     existing rows and existing INSERT call sites are unaffected by the
--     DEFAULT.
--
-- Named CHECK constraints (explicit names below) so a later integration test
-- can assert on the constraint name in a Postgres 23514 error, mirroring the
-- ck_* / *_check convention already used across this schema.

BEGIN;

-- approval_route_stages ------------------------------------------------------

ALTER TABLE public.approval_route_stages
    ADD COLUMN IF NOT EXISTS stage_kind text NOT NULL DEFAULT 'approval',
    ADD COLUMN IF NOT EXISTS due_in_days integer;

ALTER TABLE public.approval_route_stages
    DROP CONSTRAINT IF EXISTS approval_route_stages_stage_kind_check;
ALTER TABLE public.approval_route_stages
    ADD CONSTRAINT approval_route_stages_stage_kind_check
    CHECK (stage_kind IN ('review', 'approval'));

ALTER TABLE public.approval_route_stages
    DROP CONSTRAINT IF EXISTS approval_route_stages_due_in_days_check;
ALTER TABLE public.approval_route_stages
    ADD CONSTRAINT approval_route_stages_due_in_days_check
    CHECK (due_in_days IS NULL OR due_in_days > 0);

-- approval_stage_instances ----------------------------------------------------

ALTER TABLE public.approval_stage_instances
    ADD COLUMN IF NOT EXISTS stage_kind text NOT NULL DEFAULT 'approval',
    ADD COLUMN IF NOT EXISTS due_at timestamp with time zone;

ALTER TABLE public.approval_stage_instances
    DROP CONSTRAINT IF EXISTS approval_stage_instances_stage_kind_check;
ALTER TABLE public.approval_stage_instances
    ADD CONSTRAINT approval_stage_instances_stage_kind_check
    CHECK (stage_kind IN ('review', 'approval'));

-- approval_instances -----------------------------------------------------------

ALTER TABLE public.approval_instances
    ADD COLUMN IF NOT EXISTS frozen_content_hash text,
    ADD COLUMN IF NOT EXISTS cancel_reason text;

ALTER TABLE public.approval_instances
    DROP CONSTRAINT IF EXISTS approval_instances_frozen_content_hash_check;
ALTER TABLE public.approval_instances
    ADD CONSTRAINT approval_instances_frozen_content_hash_check
    CHECK (frozen_content_hash IS NULL OR frozen_content_hash ~ '^[0-9a-f]{64}$');

-- approval_signoffs -------------------------------------------------------------

ALTER TABLE public.approval_signoffs
    ADD COLUMN IF NOT EXISTS signature_meaning text NOT NULL DEFAULT 'approval';

ALTER TABLE public.approval_signoffs
    DROP CONSTRAINT IF EXISTS approval_signoffs_signature_meaning_check;
ALTER TABLE public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_signature_meaning_check
    CHECK (signature_meaning IN ('approval', 'rejection'));

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0286', 'M2b F1: purely additive StageKind (review/approval) substrate. Adds approval_route_stages.stage_kind/due_in_days, approval_stage_instances.stage_kind/due_at, approval_instances.frozen_content_hash/cancel_reason, approval_signoffs.signature_meaning, all nullable-or-defaulted so no existing row or code path changes behavior. Named CHECK constraints for later constraint-name assertions. No service wiring (F2-F8).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
