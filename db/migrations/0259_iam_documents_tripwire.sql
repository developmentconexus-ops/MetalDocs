-- 0259_iam_documents_tripwire.sql
-- SEC-05 (P1): IAM mutating tables have tier-1 route→capability gates only —
-- residual DB-layer gap identified in wiki/modules/iam-tech-debt.md:106 (T-004,
-- "PARTIALLY CLOSED") and wiki/reviews/grade-a-simplification-report-2026-07-01.md
-- SEC-05. T-004's Go-layer remediation (Plan 5) already added tier-2
-- authz.Require(...) to every write path in
-- internal/modules/iam/infrastructure/postgres/role_admin_repository.go
-- (CapUserManage, "tenant") and user_area_repository.go (CapMembershipManage,
-- <areaCode>) — verified present in the current tree, no Go changes needed here.
-- migration 0188 (archived) attached trg_require_cap_asserted to
-- metaldocs.iam_user_roles and public.user_process_areas, and to public.documents
-- (SEC-06's DB half — confirmed already attached, baseline :3886 — not touched by
-- this migration to avoid duplicating the other agent's Go-side SEC-06 work).
--
-- Residual DB-layer gap this migration closes:
--   - metaldocs.iam_users: NOT NULL tenant_id column (baseline :1283) but no
--     trg_require_cap_asserted trigger and no enforce_capability_asserted() CASE
--     branch — a direct INSERT/UPDATE/DELETE against iam_users bypasses tier-2
--     entirely (defense-in-depth gap; the Go repository layer happens to always
--     call authz.Require first, but the DB does not enforce it).
--
--   IMPORTANT SCOPING (found during investigation, not in the original ticket):
--   iam_users is NOT a purely-privileged-admin table like iam_user_roles. Two
--   hot, non-admin, per-request runtime paths issue raw UPDATEs against it with
--   NO authz context at all, on behalf of the currently authenticated user (not
--   an admin managing someone else) —
--     internal/modules/iam/infrastructure/postgres/login_context_repository.go
--       RecordLoginContext: UPDATE ... SET last_login_ip, last_login_user_agent,
--       last_login_device_label, updated_at — fires on every successful login.
--     internal/modules/iam/presence/repository.go
--       PostgresRepository.BumpLastSeen: UPDATE ... SET last_seen_at — fires on
--       every debounced authenticated request (presence heartbeat).
--   Gating ALL of iam_users' UPDATE with user.manage would break login and the
--   presence heartbeat in production (neither path asserts user.manage, nor
--   should it — self-service telemetry is not "manage users"). This is why
--   0188/0258 never extended the tripwire to iam_users despite iam_user_roles
--   and user_process_areas being covered: the table mixes privileged admin
--   columns (display_name, is_active, tenant_id, deactivated_at — written only
--   by role_admin_repository.go, which already asserts user.manage) with
--   self-service telemetry columns (last_login_*, last_seen_at, mfa_*) in one
--   row. A blanket UPDATE trigger cannot distinguish the two.
--
--   Fix: use a Postgres COLUMN-LEVEL trigger (`UPDATE OF <cols>`) so the
--   tripwire fires only when a privileged column changes, leaving the
--   telemetry-only UPDATE statements (which never touch those columns) outside
--   the trigger's fan-in entirely — not merely inside a CASE branch that would
--   still require metaldocs.asserted_caps to be set. INSERT and DELETE are
--   fully gated (grep-confirmed: the only INSERT/DELETE callers are
--   role_admin_repository.go, already asserting user.manage, and
--   internal/test/e2e_seed.go, which already pre-asserts user.manage in its
--   e2eAssertedCaps superset. Every test-fixture call site that raw-inserts into
--   iam_users/iam_user_roles (tests/integration/testdb/{factory,fixtures}.go's
--   NewUser/SeedSystemAdmin, tests/integration/fixtures/seed.go's SeedUser, and
--   the in-scope tests/integration/iam/*.go + internal/modules/iam/infrastructure
--   /postgres/*_integration_test.go raw-SQL seeders) was updated in this change
--   to assert user.manage tx-locally (seedWithCaps) or via the scheduler bypass
--   GUC (withBypass), matching the production authz.Require convention.
--   - metaldocs.iam_groups / iam_group_members / iam_group_roles: zero
--     Go write path exists today (grep-confirmed: these tables are read-only in
--     application code, joined only for system_admin/capability derivation —
--     internal/modules/iam/authz/authz.go SystemAdminExistsSQL,
--     internal/modules/iam/application/capability_service.go). They carry NEITHER
--     a tier-1 route gate (no /groups route registered) NOR a tier-2 check NOR a
--     tripwire — genuinely zero-layer. SEC-05 explicitly names "groups" in its
--     finding. Attaching the tripwire now is safe (no live writer to break) and
--     closes the gap before any future group-admin feature is built on top of an
--     unguarded table. Required cap is user.manage, matching iam_user_roles
--     (group role-grant is the same privileged operation class — see
--     SystemAdminExistsSQL, which treats iam_group_roles.role = 'system_admin'
--     as equivalent to a direct iam_user_roles system_admin grant).
--   - iam_group_roles has no tenant_id column (group_id → iam_groups is its only
--     tenant linkage) — mirrors the document_families / templates_v2_template_version
--     precedent (migration 0188) for tables lacking tenant_id: v_tenant_id := NULL.
--
-- Convention followed exactly (0257/0258/archived 0188): CREATE OR REPLACE
-- preserving every existing CASE branch byte-for-byte (including 0258's
-- document_families branch with NEW.tenant_id and the current fail-closed ELSE
-- RAISE EXCEPTION — NOT the older 0188 RETURN NEW pass-through), new branches
-- inserted before ELSE, DROP TRIGGER IF EXISTS + CREATE TRIGGER
-- trg_require_cap_asserted per new table, ledger insert, BEGIN/COMMIT wrapper.
-- No backfill/data-fix needed (iam_users/iam_group* rows already carry tenant_id
-- or need none), so no DISABLE/ENABLE TRIGGER dance is required.

BEGIN;

-- ── enforce_capability_asserted(): add iam_users + iam_group* branches ──────
-- Preserves every branch from 0258 (db/migrations/0258_document_families_tenant_scope.sql)
-- byte-for-byte; only the four new WHEN arms below (iam_users, iam_groups,
-- iam_group_members, iam_group_roles) are added, before ELSE.

CREATE OR REPLACE FUNCTION public.enforce_capability_asserted() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  v_bypass        TEXT;
  v_asserted_raw  TEXT;
  v_asserted      JSONB;
  v_required_caps TEXT[];
  v_tenant_id     UUID;
  v_cap_found     BOOLEAN := FALSE;
  v_element       JSONB;
BEGIN
  CASE
    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.submit'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.signoff'];
      v_tenant_id     := NEW.actor_tenant_id;
    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'user_process_areas' THEN
      v_required_caps := ARRAY['membership.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['document.edit'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['controlled_documents.obsolete', 'controlled_documents.supersede'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['controlled_documents.create'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'templates_template' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'templates_template_version' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit', 'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;
    -- ── New branches (SEC-05 / T-004 residual) ──────────────────────────────
    WHEN TG_TABLE_NAME = 'iam_users' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'iam_groups' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'iam_group_members' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;
    WHEN TG_TABLE_NAME = 'iam_group_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      -- iam_group_roles has no tenant_id column (scoped transitively via
      -- group_id -> iam_groups.tenant_id); mirrors the document_families
      -- (pre-0258) / templates_v2_template_version (0188) precedent for
      -- tenant_id-less tables.
      v_tenant_id     := NULL;
    ELSE
      -- Fail-closed: a table carrying this trigger with no capability mapping is a
      -- wiring error, not a license to pass through. Refuse the write loudly.
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: no capability mapping for table % (op %); trg_require_cap_asserted attached without an enforce_capability_asserted CASE branch', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'P0001';
  END CASE;

  v_bypass := pg_catalog.current_setting('metaldocs.bypass_authz', true);
  IF v_bypass IS NOT NULL AND v_bypass <> '' THEN
    IF v_bypass = 'scheduler' THEN
      BEGIN
        INSERT INTO public.governance_events
          (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
        VALUES (
          v_tenant_id,
          'authz.bypass_used',
          'system:scheduler',
          TG_TABLE_NAME,
          COALESCE(NEW.id::TEXT, 'unknown'),
          'scheduler bypass for ' || v_required_caps[1],
          pg_catalog.to_jsonb(jsonb_build_object(
            'required_caps', to_jsonb(v_required_caps),
            'bypass_token',  v_bypass,
            'table',         TG_TABLE_NAME,
            'op',            TG_OP
          ))
        );
      EXCEPTION WHEN others THEN
        RAISE NOTICE 'enforce_capability_asserted: governance_events insert failed: %', SQLERRM;
      END;
      RETURN NEW;
    ELSE
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %', v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);
  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %', v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  BEGIN
    v_asserted := v_asserted_raw::JSONB;
  EXCEPTION WHEN invalid_text_representation OR others THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)', v_required_caps
      USING ERRCODE = 'P0001';
  END;

  IF jsonb_typeof(v_asserted) <> 'array' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)', v_required_caps
      USING ERRCODE = 'P0001';
  END IF;

  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP
    IF (v_element->>'cap') = ANY(v_required_caps) THEN
      v_cap_found := TRUE;
      EXIT;
    END IF;
  END LOOP;

  IF NOT v_cap_found THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %', v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  RETURN NEW;
END;
$$;

-- ── Attach trg_require_cap_asserted to the newly-covered tables ─────────────
-- Naming/shape convention exact match to archived migration 0188 and baseline
-- attachments (e.g. iam_user_roles, baseline :3816): DROP TRIGGER IF EXISTS +
-- CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ...
-- FOR EACH ROW. iam_users/iam_group* are row-level admin tables with occasional
-- corrective DELETEs (unlike append-mostly documents/controlled_documents which
-- only got INSERT OR UPDATE) — DELETE included to match the sibling admin tables
-- iam_user_roles and user_process_areas.
--
-- iam_users is COLUMN-SCOPED on UPDATE (Postgres `UPDATE OF <cols>` trigger
-- form): the trigger fires for every INSERT and DELETE, but for UPDATE only
-- when one of the privileged columns actually changes. This is the fix for the
-- login-context / presence-heartbeat hazard documented in the header comment —
-- last_login_ip, last_login_user_agent, last_login_device_label, last_seen_at,
-- mfa_enabled, mfa_enrolled_at are deliberately EXCLUDED from the trigger's
-- column list so RecordLoginContext and BumpLastSeen keep working unguarded
-- (self-service telemetry, not a "manage users" action). display_name,
-- is_active, tenant_id, deactivated_at are the columns role_admin_repository.go
-- actually writes under its existing user.manage assertion — those are gated.
DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.iam_users;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR DELETE OR UPDATE OF display_name, is_active, tenant_id, deactivated_at
  ON metaldocs.iam_users
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.iam_groups;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.iam_groups
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.iam_group_members;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.iam_group_members
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.iam_group_roles;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.iam_group_roles
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0259', 'SEC-05: attach enforce_capability_asserted tripwire to metaldocs.iam_users (column-scoped UPDATE excluding login/presence telemetry cols) and iam_group* (user.manage); documents/iam_user_roles/user_process_areas already covered (0188), not touched')
ON CONFLICT (version) DO NOTHING;

COMMIT;
