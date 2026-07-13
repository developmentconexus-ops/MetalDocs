-- 0302_drop_templates_approval_config.sql
-- Unit 3.1a S5 (ADR 0082 phase c -- legacy template-approval retirement):
-- drop public.templates_approval_config, the legacy role-based approval
-- config table (template_id PK -> templates_template.id, reviewer_role,
-- approver_role). The approval kernel's route model (ADR 0082) is the sole
-- approval configuration from this migration on.
--
-- ============================================================================
-- PREMISE VERIFICATION (runtime truth, not assumed from docs)
-- ============================================================================
--   - WRITE path: gone. 3.1a S1 (62a415f9) removed the create-time role
--     seeding + config upsert (CreateTemplate no longer writes this table);
--     3.1a S4 (e3eb9884) deleted the 4 legacy approval routes/handlers, the
--     legacy Service methods, domain role symbols, and the repository config
--     methods. grep across internal/ apps/ tests/ finds zero remaining
--     readers or writers of this table.
--   - GDPR erasure: the templates TenantDataPort's EraseTenantData carried a
--     raw DELETE against this table (in-module child cleanup before the
--     templates_template erase); that DELETE is removed in the SAME
--     change-set as this migration
--     (internal/modules/templates/infrastructure/tenant_data_port.go) --
--     landing the DROP alone would have broken tenant erasure at runtime.
--   - Data: M3 P3.S1 (R4) counted 0 rows in every observed environment
--     (SELECT count(*) FROM public.templates_approval_config -> 0; see
--     docs/superpowers/reports/2026-07-12-m3-kernel-extraction-evidence.md),
--     and S1 removed the only INSERT path, so rows cannot have reappeared on
--     any environment running S1+ code. Environments still carrying pre-S1
--     rows are exactly what the pre-drop assert below exists to catch.
--
-- ============================================================================
-- INVENTORY: objects that die with this table (verified against
-- db/baseline/0001_current_schema.sql)
-- ============================================================================
--   - PK: templates_approval_config_pkey (template_id)
--   - Outbound FK: templates_approval_config_template_id_fkey ->
--     public.templates_template(id) (plain, non-cascading)
--   - No RLS/policies, no triggers (no tripwire arm or attachment), no
--     indexes beyond the PK, no owned sequences (uuid PK, no serial), and NO
--     inbound FK from any table (grep-confirmed against the baseline).
--
-- ============================================================================
-- SAFETY
-- ============================================================================
-- Forward-only; safe to run exactly once and to no-op on re-run or in an
-- environment already lacking the table (existence-guarded assert + DROP IF
-- EXISTS + ledger ON CONFLICT DO NOTHING). The pre-drop emptiness assert
-- fails the migration loud (P0001) instead of silently destroying rows: a
-- populated table means a still-governed legacy approval flow that must be
-- migrated to kernel approval routes deliberately, never dropped by side
-- effect. Deliberately NO CASCADE: zero dependents are known (see inventory),
-- so an unknown dependent in a divergent environment should fail the DROP
-- loudly rather than be silently destroyed along with it.

BEGIN;

-- ── Pre-drop emptiness assert ────────────────────────────────────────────────
DO $$
DECLARE
  v_rows bigint;
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'templates_approval_config'
  ) THEN
    SELECT count(*) INTO v_rows FROM public.templates_approval_config;
    IF v_rows > 0 THEN
      RAISE EXCEPTION '0302 pre-drop assert: public.templates_approval_config still holds % row(s) -- ADR 0082 phase c requires it empty (legacy template-approval config must be migrated to kernel approval routes deliberately, not dropped). Inspect/migrate the rows, then re-run.', v_rows;
    END IF;
  END IF;
END $$;

-- ── Drop the retired legacy config table ────────────────────────────────────
DROP TABLE IF EXISTS public.templates_approval_config;

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0302', 'Unit 3.1a S5 (ADR 0082 phase c): drop public.templates_approval_config behind a pre-drop emptiness assert -- legacy role-based template-approval config retired. S1 removed the create-time writer, S4 deleted the 4 legacy routes and every remaining reader/writer, and the templates TenantDataPort erase DELETE is removed in the same change-set. The approval kernel route model is the sole approval configuration.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
