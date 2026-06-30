-- 0255_template_version_tenant_consistency_trigger.sql
--
-- F-DB5 [MEDIUM]: templates_template_version has no tenant_id column, so the
-- enforce_capability_asserted trigger sets v_tenant_id := NULL for it (baseline
-- ~:486).  Isolation currently depends 100% on the JOIN to templates_template in
-- every application query.  The concrete gap: any UPDATE that reaches the table
-- by template_id+status (e.g. ObsoletePreviousPublished) can cross tenants if
-- upstream validation is bypassed.
--
-- FIX (short-term — closes the concrete cross-tenant write gap):
--   Add a BEFORE INSERT OR UPDATE trigger function that reads the asserting
--   tenant from the tx-local GUC metaldocs.tenant_id (set by authz.SeedTxIdentity
--   before every guarded write) and verifies that the row's parent template
--   belongs to that same tenant.  The function is idempotent (CREATE OR REPLACE).
--   The trigger binding is idempotent (DROP TRIGGER IF EXISTS before CREATE).
--
-- FORWARD ITEM — medium-term (OUT OF SCOPE for this migration):
--   Add a tenant_id column to templates_template_version for full RLS parity
--   (matching all other tenant tables in the schema).  This enables RLS policies
--   to protect the table without a GUC-based workaround and closes the NULL
--   v_tenant_id gap in enforce_capability_asserted.  Deferred post-v1 because it
--   requires a backfill and a baseline regeneration pass.
--
-- SIBLING PATTERN:
--   Mirrors enforce_signoff_tenant_consistent() / trg_signoff_tenant_consistent
--   (baseline ~:728-747 / ~:3860-3863) and
--   enforce_placeholder_value_tenant_consistent() (migration 0241).
--   Those triggers JOIN to a parent table to derive and check the expected tenant;
--   this trigger does the same via templates_template.
--
-- Idempotent: CREATE OR REPLACE FUNCTION + DROP TRIGGER IF EXISTS before CREATE.
-- Forward-only: no rollback section, no edits to any previously applied migration.

BEGIN;

CREATE OR REPLACE FUNCTION public.enforce_template_version_tenant_consistent()
    RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
AS $$
DECLARE
  v_asserting_tenant TEXT;
  v_parent_tenant    UUID;
BEGIN
  -- Read the tx-local tenant GUC set by authz.SeedTxIdentity.
  -- The second argument (true) suppresses the error if the GUC is unset so we
  -- can emit a descriptive message rather than a generic "unrecognised parameter".
  v_asserting_tenant := pg_catalog.current_setting('metaldocs.tenant_id', true);

  IF v_asserting_tenant IS NULL OR v_asserting_tenant = '' THEN
    -- Fail closed: a write without a tenant context is a wiring error.
    RAISE EXCEPTION 'enforce_template_version_tenant_consistent: metaldocs.tenant_id GUC is not set; tenant context required for templates_template_version writes'
      USING ERRCODE = 'check_violation';
  END IF;

  -- Resolve the parent template's tenant.
  SELECT tenant_id INTO v_parent_tenant
    FROM public.templates_template
   WHERE id = NEW.template_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'enforce_template_version_tenant_consistent: parent template % does not exist',
      NEW.template_id
      USING ERRCODE = 'foreign_key_violation';
  END IF;

  IF v_parent_tenant::TEXT IS DISTINCT FROM v_asserting_tenant THEN
    RAISE EXCEPTION 'enforce_template_version_tenant_consistent: cross-tenant write rejected (asserting tenant %, parent template tenant %)',
      v_asserting_tenant, v_parent_tenant
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;

-- Drop the old binding if it exists (idempotent re-run safety).
DROP TRIGGER IF EXISTS trg_template_version_tenant_consistent
  ON public.templates_template_version;

CREATE TRIGGER trg_template_version_tenant_consistent
  BEFORE INSERT OR UPDATE
  ON public.templates_template_version
  FOR EACH ROW
  EXECUTE FUNCTION public.enforce_template_version_tenant_consistent();

INSERT INTO public.schema_migrations (version, description)
VALUES (
  '0255',
  'add enforce_template_version_tenant_consistent trigger on templates_template_version to close cross-tenant write gap (F-DB5)'
)
ON CONFLICT (version) DO NOTHING;

COMMIT;
