-- 0230_authz_decommission_reviewer_role.sql
-- ADR 0022 — decommission the legacy `reviewer` area role + remove the stored-proc
-- role-list drift.
--
-- `reviewer` was a phantom area role: present only in the user_process_areas role
-- CHECK and in the grant/revoke stored-proc validation lists, but ABSENT from the
-- Go role registry (iam/domain), `validRoles`, the role_capabilities seed, the
-- OpenAPI UserRole enum, and any issuance path. It grants ZERO capabilities, so a
-- user holding a `reviewer` membership could be resolved as an approval-stage
-- eligible actor yet be denied at tier-2 (holds no capability) — a silently broken
-- flow. The contract already disowns it ("no admin or reviewer phantoms").
--
-- Decommission it everywhere so all layers agree on the canonical role set:
--   1. Remove the non-functional `reviewer` memberships (they conferred nothing,
--      so this is behaviour-neutral). RAISE NOTICE the count for the operator.
--   2. Tighten user_process_areas_role_check to the 7 canonical AREA roles
--      (system_admin is tenant-wide, never an area membership).
--   3. Replace grant/revoke_area_membership: DROP the duplicated `_role NOT IN
--      (...)` validation. That hardcoded list was stale (it allowed only
--      viewer/editor/reviewer/approver — rejecting author/signer/area_admin/
--      qms_admin which the CHECK permits) AND it named the phantom. The table
--      CHECK is the SINGLE SOURCE OF TRUTH for the allowed role set; an invalid
--      role now fails on the INSERT against that one constraint, with no second
--      list to drift.
--
-- Forward-only, idempotent. Mirrored inline in the curated baseline
-- (db/baseline/0001_current_schema.sql).

BEGIN;

-- 1. Purge the non-functional 'reviewer' memberships (active AND historical) so
--    the tightened CHECK below validates against every row. user_process_areas is
--    normally append + soft-delete only — trg_user_process_areas_no_delete raises
--    on any DELETE, and trg_require_cap_asserted gates writes on a capability
--    assertion. Both are disabled ONLY for this one-time governance purge of a
--    decommissioned, zero-capability role value, then re-enabled in the SAME
--    transaction (transactional DDL: a rollback restores them). This does not
--    weaken the soft-delete contract for real memberships — it removes an
--    illegitimate role value that should never have been storable.
ALTER TABLE public.user_process_areas DISABLE TRIGGER trg_require_cap_asserted;
ALTER TABLE public.user_process_areas DISABLE TRIGGER trg_user_process_areas_no_delete;
DO $$
DECLARE
  _removed INT;
BEGIN
  DELETE FROM public.user_process_areas WHERE role = 'reviewer';
  GET DIAGNOSTICS _removed = ROW_COUNT;
  IF _removed > 0 THEN
    RAISE NOTICE '0230: purged % decommissioned reviewer area membership row(s)', _removed;
  END IF;
END$$;
ALTER TABLE public.user_process_areas ENABLE TRIGGER trg_user_process_areas_no_delete;
ALTER TABLE public.user_process_areas ENABLE TRIGGER trg_require_cap_asserted;

-- 2. Tighten the area-role CHECK to the 7 canonical area roles (drop 'reviewer').
ALTER TABLE public.user_process_areas
    DROP CONSTRAINT IF EXISTS user_process_areas_role_check;
ALTER TABLE public.user_process_areas
    ADD CONSTRAINT user_process_areas_role_check CHECK (role = ANY (ARRAY[
        'viewer'::text, 'editor'::text, 'approver'::text, 'author'::text,
        'signer'::text, 'area_admin'::text, 'qms_admin'::text
    ]));

-- 3a. grant_area_membership — drop the duplicated role-list validation.
CREATE OR REPLACE FUNCTION metaldocs.grant_area_membership(_tenant_id uuid, _user_id text, _area_code text, _role text, _granted_by text) RETURNS uuid
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $_$
DECLARE
  _existing_from  TIMESTAMPTZ;
  _now            TIMESTAMPTZ := pg_catalog.clock_timestamp();
  _correlation_id UUID        := pg_catalog.gen_random_uuid();
BEGIN
  -- ── Input validation (before any query) ──────────────────────────────────
  -- Role values are NOT re-validated here: the user_process_areas role CHECK
  -- constraint is the single source of truth and rejects an invalid role on the
  -- INSERT below (ADR 0022 — no duplicated role list to drift).
  IF _user_id !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid user_id: %', _user_id USING ERRCODE = '22023';
  END IF;
  IF _area_code !~ '^[A-Z0-9_]+$' THEN
    RAISE EXCEPTION 'invalid area_code: %', _area_code USING ERRCODE = '22023';
  END IF;
  IF _granted_by !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid granted_by: %', _granted_by USING ERRCODE = '22023';
  END IF;

  -- ── Idempotency check ────────────────────────────────────────────────────
  SELECT effective_from
    INTO _existing_from
    FROM public.user_process_areas
   WHERE tenant_id    = _tenant_id
     AND user_id      = _user_id
     AND area_code    = _area_code
     AND role         = _role
     AND effective_to IS NULL
   LIMIT 1;

  IF FOUND THEN
    -- Already active — return correlation id without side-effects.
    RETURN _correlation_id;
  END IF;

  -- ── Insert membership row ────────────────────────────────────────────────
  INSERT INTO public.user_process_areas
    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
  VALUES
    (_user_id, _tenant_id, _area_code, _role, _now, NULL, _granted_by, NULL);

  -- ── Emit governance event ────────────────────────────────────────────────
  INSERT INTO metaldocs.governance_events
    (tenant_id, event_type, actor_user_id, resource_type, resource_id, payload_json)
  VALUES
    (_tenant_id,
     'membership.granted',
     _granted_by,
     'user_process_area',
     _user_id || ':' || _area_code || ':' || _role,
     pg_catalog.to_jsonb(
       pg_catalog.json_build_object(
         'user_id',        _user_id,
         'area_code',      _area_code,
         'role',           _role,
         'granted_by',     _granted_by,
         'effective_from', _now
       )
     )
    );

  RETURN _correlation_id;
END;
$_$;

-- 3b. revoke_area_membership — drop the duplicated role-list validation.
CREATE OR REPLACE FUNCTION metaldocs.revoke_area_membership(_tenant_id uuid, _user_id text, _area_code text, _role text, _revoked_by text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $_$
DECLARE
  _rows_affected  INT;
  _now            TIMESTAMPTZ := pg_catalog.clock_timestamp();
BEGIN
  -- ── Input validation (before any query) ──────────────────────────────────
  -- Role not re-validated; an unknown role simply matches no active row below.
  IF _user_id !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid user_id: %', _user_id USING ERRCODE = '22023';
  END IF;
  IF _area_code !~ '^[A-Z0-9_]+$' THEN
    RAISE EXCEPTION 'invalid area_code: %', _area_code USING ERRCODE = '22023';
  END IF;
  IF _revoked_by !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid revoked_by: %', _revoked_by USING ERRCODE = '22023';
  END IF;

  -- ── Soft-delete: set effective_to + revoked_by ───────────────────────────
  UPDATE public.user_process_areas
     SET effective_to = _now,
         revoked_by   = _revoked_by
   WHERE tenant_id    = _tenant_id
     AND user_id      = _user_id
     AND area_code    = _area_code
     AND role         = _role
     AND effective_to IS NULL;

  GET DIAGNOSTICS _rows_affected = ROW_COUNT;

  IF _rows_affected = 0 THEN
    RAISE EXCEPTION 'membership not found'
      USING ERRCODE = 'P0002';
  END IF;

  -- ── Emit governance event ────────────────────────────────────────────────
  INSERT INTO metaldocs.governance_events
    (tenant_id, event_type, actor_user_id, resource_type, resource_id, payload_json)
  VALUES
    (_tenant_id,
     'membership.revoked',
     _revoked_by,
     'user_process_area',
     _user_id || ':' || _area_code || ':' || _role,
     pg_catalog.to_jsonb(
       pg_catalog.json_build_object(
         'user_id',      _user_id,
         'area_code',    _area_code,
         'role',         _role,
         'revoked_by',   _revoked_by,
         'effective_to', _now
       )
     )
    );
END;
$_$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0230', 'ADR 0022: decommission legacy reviewer area role (drop from user_process_areas CHECK + stored procs) and remove duplicated stored-proc role-list (table CHECK is the single source of truth)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
