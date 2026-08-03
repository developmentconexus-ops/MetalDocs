-- 0288_review_verdicts_changes_requested.sql
-- M2b F4 (approval-kernel-backend, approval remediation design spec.md/plan.md
-- §2, W11/W13): the review-stage runtime-verdict substrate. Adds the
-- InstanceChangesRequested (5th, non-terminal) approval_instances.status value
-- and a new approval_review_verdicts table for `ready`/`request_changes`
-- verdicts recorded against review-kind stages (StageKindReview, migration
-- 0286). Verdicts are their own row shape, NOT a widened reuse of
-- approval_signoffs (only the EvaluateQuorum counting function is reused, not
-- the storage row) — see spec.md Non-goals.
--
-- approval_instances.cancel_reason already exists (migration 0286); this
-- migration adds no new column there, only the widened status CHECK.
--
-- Idempotent (CREATE TABLE IF NOT EXISTS; DROP CONSTRAINT IF EXISTS before ADD
-- CONSTRAINT — idiom copied from 0286_approval_stage_kinds_expand.sql).

BEGIN;

-- approval_instances.status -------------------------------------------------
-- Widen the CHECK to admit 'changes_requested' alongside the existing four
-- values. No existing row's status changes; this is purely additive.

ALTER TABLE public.approval_instances
    DROP CONSTRAINT IF EXISTS approval_instances_status_check;
ALTER TABLE public.approval_instances
    ADD CONSTRAINT approval_instances_status_check
    CHECK (status = ANY (ARRAY['in_progress'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text, 'changes_requested'::text]));

-- approval_review_verdicts ---------------------------------------------------
-- Mirrors approval_signoffs' column shape (id, tenant_id via actor_tenant_id,
-- approval_instance_id, stage_instance_id, actor_user_id, decision-equivalent,
-- comment, timestamp, actor_display_name_snapshot) — adapted: verdicts carry
-- no signature_method/signature_payload/content_hash (no e-signature
-- semantics for review-stage feedback, unlike binding approval signoffs).

CREATE TABLE IF NOT EXISTS public.approval_review_verdicts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    approval_instance_id uuid NOT NULL,
    stage_instance_id uuid NOT NULL,
    actor_user_id text NOT NULL,
    actor_tenant_id uuid NOT NULL,
    verdict text NOT NULL,
    comment text,
    verdict_at timestamp with time zone DEFAULT now() NOT NULL,
    actor_display_name_snapshot text,
    CONSTRAINT approval_review_verdicts_pkey PRIMARY KEY (id),
    CONSTRAINT approval_review_verdicts_verdict_check
        CHECK (verdict = ANY (ARRAY['ready'::text, 'request_changes'::text])),
    CONSTRAINT approval_review_verdicts_stage_actor_uq
        UNIQUE (stage_instance_id, actor_user_id)
);

-- RLS: idiom copied VERBATIM from 0285_approval_signoffs_rls.sql (same
-- actor_tenant_id-keyed shape, same null-GUC escape-hatch branch, ADR 0027).

ALTER TABLE public.approval_review_verdicts ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY public.approval_review_verdicts FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON public.approval_review_verdicts;
CREATE POLICY tenant_isolation ON public.approval_review_verdicts
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (actor_tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0288', 'M2b F4: review-stage runtime verdicts substrate. Widens approval_instances_status_check to admit changes_requested (5th, non-terminal status). Adds approval_review_verdicts (ready/request_changes, one verdict per actor per stage via UNIQUE(stage_instance_id, actor_user_id)), RLS ENABLE+FORCE + tenant_isolation policy on actor_tenant_id copied verbatim from 0285. No reuse of approval_signoffs row shape (only EvaluateQuorum counting fn reused).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
