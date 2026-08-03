-- 0296_approval_subject_generalization.sql
-- ROADMAP unit 3.1 (M3 — approval kernel extraction, P2.S1). Ratified spec:
-- ADR 0082 (approval subject generalization: document -> (subject_kind,
-- subject_key)).
--
-- EXPAND phase only, forward-only, idempotent (ADD COLUMN IF NOT EXISTS;
-- DROP CONSTRAINT/INDEX IF EXISTS before ADD — idiom from 0286/0287/0290/0295).
-- Behavior-preserving for every existing document row and for every EXISTING
-- INSERT call site (production repository code AND the shared integration
-- test factory), none of which is touched in this slice (P2.S2 does the Go
-- cutover).
--
-- Generalizes the approval kernel's keying from document-specific columns
-- (approval_routes.profile_code / approval_instances.document_id) to a
-- subject abstraction: subject_kind ('document' | 'template') + subject_key
-- (text). This lets a later phase (P3) point the SAME kernel at template
-- approvals without a parallel table or a parallel service.
--
--   approval_routes:
--     profile_code stays NOT NULL as the backward-compat alias for the
--     document case: (subject_kind='document', subject_key=profile_code).
--     KEPT: approval_routes_active_profile_uq / approval_routes_profile_version_uq
--     (tenant_id, profile_code [, version] — the 0287 route-versioning
--     replacements for the original baseline approval_routes_tenant_profile_key,
--     which 0287 already dropped; this migration does not touch either).
--     ADDED: ux_approval_routes_tenant_subject (tenant_id, subject_kind, subject_key).
--
--   approval_instances:
--     document_id stays NOT NULL THIS PHASE (only document rows exist yet;
--     template rows arrive in a later phase, which is the phase that relaxes
--     this NOT NULL — do not relax it here).
--     KEPT: approval_instances_document_id_idempotency_key_key,
--           ux_approval_instances_active_document_id (partial, status='in_progress'),
--           ix_approval_instances_tenant_document_id.
--     ADDED: ux_approval_instances_active_subject (partial, status='in_progress'),
--            ix_approval_instances_tenant_subject.
--
-- Compatibility trigger (the reason this expand phase is safe without a Go
-- cutover): every existing INSERT into either table omits subject_kind/
-- subject_key entirely (see internal/modules/approval/infrastructure/
-- postgres_approval_repository.go InsertInstance, internal/modules/approval/
-- application/route_admin_service.go route create + version-copy, and
-- tests/integration/testdb/factory.go NewApprovalRoute/NewApprovalInstance).
-- A BEFORE INSERT trigger (public.default_approval_subject) fills NEW.subject_kind
-- / NEW.subject_key from the legacy document columns whenever the caller left
-- them NULL, running before the NOT NULL constraint is checked. New code
-- (P2.S2 onward) that wants a template subject simply passes subject_kind/
-- subject_key explicitly, which the trigger leaves untouched.
--
-- Multi-tenant: new indexes are tenant-scoped (tenant_id leads every new
-- index). RLS FORCE is already active on both tables from the baseline; new
-- columns inherit the existing table-level tenant_isolation policy — no
-- per-column grants exist anywhere in the baseline for these tables, so none
-- are added here.
--
-- Out of scope / untouched: metaldocs.document_profiles (E-PROD-2, a known
-- open defect in a different schema — not touched by this migration).
--
-- DEBT — remove in the CONTRACT phase (tracked on P2.S2 in
-- docs/superpowers/specs/2026-07-12-m3-approval-kernel-extraction-plan.md):
-- public.default_approval_subject() and its two triggers
-- (trg_approval_routes_default_subject / trg_approval_instances_default_subject)
-- are a TEMPORARY expand-phase shim. They exist only so pre-cutover Go/test
-- INSERTs that omit subject_kind/subject_key keep working under the new NOT NULL.
-- Once every writer (production repositories in internal/modules/approval AND
-- the tests/integration/testdb factory) sets subject_kind/subject_key
-- explicitly, DROP both triggers and the function — otherwise a future write
-- that forgets to set the subject is silently defaulted to 'document' instead
-- of failing loudly, masking the exact bug the CHECK constraint exists to catch.

BEGIN;

-- 1. approval_routes -----------------------------------------------------

ALTER TABLE public.approval_routes
    ADD COLUMN IF NOT EXISTS subject_kind text,
    ADD COLUMN IF NOT EXISTS subject_key text;

UPDATE public.approval_routes
   SET subject_kind = 'document',
       subject_key  = profile_code
 WHERE subject_kind IS NULL
    OR subject_key IS NULL;

-- 2. approval_instances ---------------------------------------------------

ALTER TABLE public.approval_instances
    ADD COLUMN IF NOT EXISTS subject_kind text,
    ADD COLUMN IF NOT EXISTS subject_key text;

UPDATE public.approval_instances
   SET subject_kind = 'document',
       subject_key  = document_id::text
 WHERE subject_kind IS NULL
    OR subject_key IS NULL;

-- 3. Compatibility trigger: auto-populate subject_kind/subject_key on any
--    INSERT that omits them, from the legacy document-specific column on the
--    same row. Runs BEFORE the NOT NULL constraint (added in step 4) is
--    checked, so every pre-existing call site keeps working unmodified.

CREATE OR REPLACE FUNCTION public.default_approval_subject() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
    IF TG_TABLE_NAME = 'approval_routes' THEN
        IF NEW.subject_kind IS NULL THEN
            NEW.subject_kind := 'document';
        END IF;
        IF NEW.subject_key IS NULL THEN
            NEW.subject_key := NEW.profile_code;
        END IF;
    ELSIF TG_TABLE_NAME = 'approval_instances' THEN
        IF NEW.subject_kind IS NULL THEN
            NEW.subject_kind := 'document';
        END IF;
        IF NEW.subject_key IS NULL THEN
            NEW.subject_key := NEW.document_id::text;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_approval_routes_default_subject ON public.approval_routes;
CREATE TRIGGER trg_approval_routes_default_subject
    BEFORE INSERT ON public.approval_routes
    FOR EACH ROW EXECUTE FUNCTION public.default_approval_subject();

DROP TRIGGER IF EXISTS trg_approval_instances_default_subject ON public.approval_instances;
CREATE TRIGGER trg_approval_instances_default_subject
    BEFORE INSERT ON public.approval_instances
    FOR EACH ROW EXECUTE FUNCTION public.default_approval_subject();

-- 4. NOT NULL + CHECK ------------------------------------------------------

ALTER TABLE public.approval_routes
    ALTER COLUMN subject_kind SET NOT NULL,
    ALTER COLUMN subject_key SET NOT NULL;

ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_subject_kind_check;
ALTER TABLE public.approval_routes
    ADD CONSTRAINT approval_routes_subject_kind_check
    CHECK (subject_kind IN ('document', 'template'));

ALTER TABLE public.approval_instances
    ALTER COLUMN subject_kind SET NOT NULL,
    ALTER COLUMN subject_key SET NOT NULL;

ALTER TABLE public.approval_instances
    DROP CONSTRAINT IF EXISTS approval_instances_subject_kind_check;
ALTER TABLE public.approval_instances
    ADD CONSTRAINT approval_instances_subject_kind_check
    CHECK (subject_kind IN ('document', 'template'));

-- 5. New subject-keyed indexes, ALONGSIDE the existing document-keyed ones -

-- PARTIAL (WHERE active), mirroring approval_routes_active_profile_uq (0287
-- route versioning): a superseded/historical route row legitimately keeps the
-- same (tenant_id, subject_kind, subject_key) as its successor, so only the
-- live row per subject may be unique. Discovered via regression: the approval
-- integration suite's TestRouteVersioning_InUse_DefinitionUpdateBlocked_
-- ThenSupersede failed 23505 against an earlier non-partial draft of this
-- index when it inserted the superseding (inactive-predecessor) row.
DROP INDEX IF EXISTS public.ux_approval_routes_tenant_subject;
CREATE UNIQUE INDEX IF NOT EXISTS ux_approval_routes_tenant_subject
    ON public.approval_routes (tenant_id, subject_kind, subject_key)
    WHERE (active);

DROP INDEX IF EXISTS public.ux_approval_instances_active_subject;
CREATE UNIQUE INDEX IF NOT EXISTS ux_approval_instances_active_subject
    ON public.approval_instances (tenant_id, subject_kind, subject_key)
    WHERE (status = 'in_progress');

DROP INDEX IF EXISTS public.ix_approval_instances_tenant_subject;
CREATE INDEX IF NOT EXISTS ix_approval_instances_tenant_subject
    ON public.approval_instances (tenant_id, subject_kind, subject_key);

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0296', 'P2.S1 approval subject generalization (expand phase, ADR 0082): approval_routes/approval_instances gain subject_kind text NOT NULL CHECK (document|template) + subject_key text NOT NULL; backfill subject_kind=''document'', subject_key=profile_code (routes) / document_id::text (instances); public.default_approval_subject() BEFORE INSERT compatibility trigger backfills both columns from the legacy document column on any INSERT that omits them (no Go cutover required this phase); profile_code and document_id stay NOT NULL backward-compat aliases; new tenant-scoped indexes ux_approval_routes_tenant_subject, ux_approval_instances_active_subject (partial, in_progress), ix_approval_instances_tenant_subject added alongside the kept document-keyed constraints/indexes.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
