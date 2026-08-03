-- 0293_approval_delegations.sql
-- M2b F9 (approval-kernel-backend, approval remediation design spec.md §4,
-- W5; ADR 0077): approval delegation. A user (delegator) authorizes another
-- user (delegate) to act with the delegator's eligibility for a time-windowed
-- period. Delegation widens the INPUT to the existing eligibility/SoD
-- predicates (domain.CheckEligibility / domain.CheckSoD) — it grants no
-- capability and creates no parallel authz path (ADR 0022).
--
-- New table: approval_delegations (tenant, delegator, delegate, window,
-- reason, audit). Overlapping windows for the same delegator are allowed
-- (union semantics); self-delegation is rejected at both the app layer
-- (domain.NewDelegation) and the DB layer (CHECK below).
--
-- New column on BOTH approval_signoffs and approval_review_verdicts:
-- on_behalf_of_user_id text NULL — dual-identity persistence. NULL when the
-- actor acted as themselves (unchanged default); the delegator's user_id when
-- eligibility was satisfied via an active delegation. Genuinely new column,
-- not a reuse of any existing column for a second purpose.
--
-- DB tripwire parity: enforce_approval_sod() (migration 0290) only compares
-- NEW.actor_user_id to the document's author. Once on_behalf_of_user_id
-- exists, a delegate acting for the author would have actor_user_id != author
-- and the trigger would wrongly pass. This migration widens the shared
-- function (CREATE OR REPLACE, same trigger attachments) to also reject when
-- NEW.on_behalf_of_user_id equals the author — symmetric with the app-level
-- domain.CheckSoD widening (this same feature).
--
-- Idempotent (CREATE TABLE IF NOT EXISTS; ADD COLUMN IF NOT EXISTS; DROP
-- CONSTRAINT/TRIGGER IF EXISTS before CREATE — idiom copied from
-- 0288/0290/0292).

BEGIN;

-- approval_delegations --------------------------------------------------

CREATE TABLE IF NOT EXISTS public.approval_delegations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    delegator_id text NOT NULL,
    delegate_id text NOT NULL,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    reason text NOT NULL,
    created_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT approval_delegations_pkey PRIMARY KEY (id),
    CONSTRAINT approval_delegations_window_chk CHECK (ends_at > starts_at),
    CONSTRAINT approval_delegations_no_self CHECK (delegator_id <> delegate_id)
);

CREATE INDEX IF NOT EXISTS approval_delegations_active_lookup_idx
    ON public.approval_delegations (tenant_id, delegate_id, starts_at, ends_at);

-- RLS: idiom copied from 0288 (NULL-GUC-permissive branch, ADR 0027), keyed
-- directly on tenant_id (delegator and delegate are always same-tenant by the
-- iam_users tenant-scoping invariant — no actor_tenant_id ambiguity here).

ALTER TABLE public.approval_delegations ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY public.approval_delegations FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON public.approval_delegations;
CREATE POLICY tenant_isolation ON public.approval_delegations
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

-- approval_signoffs / approval_review_verdicts: dual-identity column -------

ALTER TABLE public.approval_signoffs
    ADD COLUMN IF NOT EXISTS on_behalf_of_user_id text;
ALTER TABLE public.approval_review_verdicts
    ADD COLUMN IF NOT EXISTS on_behalf_of_user_id text;

-- SoD DB-tripwire parity: widen the shared function to also check
-- on_behalf_of_user_id. Same trigger names/attachments as 0290; only the
-- function body changes.

CREATE OR REPLACE FUNCTION public.enforce_approval_sod() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  author_id TEXT;
BEGIN
  SELECT d.created_by INTO author_id
    FROM public.approval_instances i
    JOIN public.documents d ON d.id = i.document_id
   WHERE i.id = NEW.approval_instance_id;

  IF NEW.actor_user_id = author_id THEN
    RAISE EXCEPTION 'SoD: author cannot sign own revision'
      USING ERRCODE = 'check_violation';
  END IF;

  IF NEW.on_behalf_of_user_id IS NOT NULL AND NEW.on_behalf_of_user_id = author_id THEN
    RAISE EXCEPTION 'SoD: delegate cannot act on behalf of the document author'
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;

-- Trigger attachments unchanged (same names, same tables) — CREATE OR REPLACE
-- FUNCTION above is sufficient; re-stating DROP/CREATE TRIGGER defensively so
-- this migration is fully self-contained and idempotent on its own.

DROP TRIGGER IF EXISTS trg_signoff_sod ON public.approval_signoffs;
CREATE TRIGGER trg_signoff_sod BEFORE INSERT ON public.approval_signoffs
    FOR EACH ROW EXECUTE FUNCTION public.enforce_approval_sod();

DROP TRIGGER IF EXISTS trg_review_verdict_sod ON public.approval_review_verdicts;
CREATE TRIGGER trg_review_verdict_sod BEFORE INSERT ON public.approval_review_verdicts
    FOR EACH ROW EXECUTE FUNCTION public.enforce_approval_sod();

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0293', 'M2b F9: approval delegation (ADR 0077). New approval_delegations table (tenant/delegator/delegate/window/reason/audit, RLS ENABLE+FORCE + tenant_isolation, no-self + window CHECKs, union semantics for overlapping windows). New on_behalf_of_user_id column on approval_signoffs and approval_review_verdicts (dual-identity persistence, NULL when acting as self). enforce_approval_sod() widened (CREATE OR REPLACE, same triggers) to also reject when on_behalf_of_user_id equals the document author — DB-tripwire parity with the app-level domain.CheckSoD widening.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
