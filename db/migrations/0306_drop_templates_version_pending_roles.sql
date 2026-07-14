-- 0306_drop_templates_version_pending_roles.sql
-- ROADMAP unit 4.4 -- schema hygiene: drop public.templates_template_version's
-- pending_reviewer_role (text, nullable) and pending_approver_role (text NOT
-- NULL DEFAULT '') columns, the legacy per-version role-routing fields left
-- behind by the role-based template approval flow retired in unit 3.1a
-- (ADR 0082 phase c). The approval kernel's route model is the sole approval
-- configuration from this migration on.
--
-- ============================================================================
-- PREMISE VERIFICATION (runtime truth, not assumed from docs)
-- ============================================================================
--   - WRITE path: gone. Unit 3.1a S1 removed the create-time role seeding and
--     upsert path; S4 deleted the legacy approval routes/handlers and Service
--     methods that populated these columns. No code has written either column
--     since 3.1a landed.
--   - READ path: gone. grep across internal/ apps/ frontend/ finds ZERO
--     Go/TS non-test readers of pending_reviewer_role or pending_approver_role
--     on public.templates_template_version -- only wiki/docs/the baseline
--     schema dump mention them. The approval kernel route model (ADR 0082)
--     replaced this role-routing entirely.
--
-- ============================================================================
-- SAFETY
-- ============================================================================
-- Forward-only; safe to run exactly once and to no-op on re-run or in an
-- environment already lacking the columns (existence-guarded assert + DROP
-- COLUMN IF EXISTS + ledger ON CONFLICT DO NOTHING). The pre-drop assert fails
-- the migration loud (P0001) instead of silently destroying data: a populated
-- column means a still-live legacy writer or unmigrated historical routing
-- data that must be handled deliberately, never dropped by side effect
-- (no-fallback principle). Deliberately NO CASCADE: a column drop has no
-- dependents, but the intent is the same as 0302/0305 -- fail loud on any
-- surprise rather than silently cascade.

BEGIN;

-- ── Pre-drop emptiness assert ────────────────────────────────────────────────
DO $$
DECLARE
  v_rows bigint;
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'templates_template_version'
      AND column_name IN ('pending_reviewer_role', 'pending_approver_role')
  ) THEN
    SELECT count(*) INTO v_rows FROM public.templates_template_version
    WHERE pending_reviewer_role IS NOT NULL OR pending_approver_role <> '';
    IF v_rows > 0 THEN
      RAISE EXCEPTION '0306 pre-drop assert: public.templates_template_version still holds % row(s) with a non-empty pending_reviewer_role/pending_approver_role -- ADR 0082 phase c requires them empty (legacy per-version role-routing must be migrated to kernel approval routes deliberately, not dropped). Inspect/migrate the rows, then re-run.', v_rows;
    END IF;
  END IF;
END $$;

-- ── Drop the retired legacy role-routing columns ─────────────────────────────
ALTER TABLE public.templates_template_version DROP COLUMN IF EXISTS pending_reviewer_role;
ALTER TABLE public.templates_template_version DROP COLUMN IF EXISTS pending_approver_role;

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0306', 'ROADMAP unit 4.4 (schema hygiene): drop public.templates_template_version.pending_reviewer_role/pending_approver_role behind a pre-drop emptiness assert -- legacy per-version role-routing columns retired since 3.1a (ADR 0082 phase c). Write-never since 3.1a S1/S4; read-never, grep-confirmed zero Go/TS non-test readers. The approval kernel route model is the sole approval configuration.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
