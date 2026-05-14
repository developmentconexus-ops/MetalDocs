-- MetalDocs curated current-state schema baseline.
-- This file contains product schema only. Product reference data belongs in
-- db/reference-data/0001_product_reference_data.sql. Local-only seed data belongs
-- in db/dev-seeds/0001_local_dev_seed.sql.

BEGIN;

--
-- PostgreSQL database dump
--




--
-- Name: metaldocs; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA IF NOT EXISTS metaldocs;


--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA IF NOT EXISTS public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: mddm_version_status; Type: TYPE; Schema: metaldocs; Owner: -
--

CREATE TYPE metaldocs.mddm_version_status AS ENUM (
    'draft',
    'pending_approval',
    'released',
    'archived'
);


--
-- Name: acquire_lease(text, text, interval); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.acquire_lease(_job text, _leader text, _ttl interval) RETURNS TABLE(acquired boolean, epoch bigint)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $$
DECLARE
    v_now TIMESTAMPTZ := now();
    v_lease metaldocs.job_leases%ROWTYPE;
BEGIN
    SELECT *
    INTO v_lease
    FROM metaldocs.job_leases
    WHERE job_name = _job
    FOR UPDATE SKIP LOCKED;

    IF NOT FOUND THEN
        IF EXISTS (
            SELECT 1
            FROM metaldocs.job_leases
            WHERE job_name = _job
        ) THEN
            RETURN QUERY SELECT false, -1::bigint;
            RETURN;
        END IF;

        BEGIN
            INSERT INTO metaldocs.job_leases (
                job_name,
                leader_id,
                lease_epoch,
                acquired_at,
                heartbeat_at,
                expires_at
            )
            VALUES (_job, _leader, 1, v_now, v_now, v_now + _ttl);

            RETURN QUERY SELECT true, 1::bigint;
            RETURN;
        EXCEPTION
            WHEN unique_violation THEN
                RETURN QUERY SELECT false, -1::bigint;
                RETURN;
        END;
    END IF;

    IF v_lease.expires_at < v_now THEN
        UPDATE metaldocs.job_leases
        SET
            leader_id = _leader,
            lease_epoch = v_lease.lease_epoch + 1,
            acquired_at = v_now,
            heartbeat_at = v_now,
            expires_at = v_now + _ttl
        WHERE job_name = _job;

        RETURN QUERY SELECT true, (v_lease.lease_epoch + 1)::bigint;
        RETURN;
    END IF;

    IF v_lease.leader_id = _leader THEN
        UPDATE metaldocs.job_leases
        SET
            heartbeat_at = v_now,
            expires_at = v_now + _ttl
        WHERE job_name = _job;

        RETURN QUERY SELECT true, v_lease.lease_epoch;
        RETURN;
    END IF;

    RETURN QUERY SELECT false, -1::bigint;
END;
$$;


--
-- Name: assert_lease_epoch(text, bigint); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.assert_lease_epoch(_job text, _epoch bigint) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM metaldocs.job_leases
        WHERE job_name = _job
          AND lease_epoch = _epoch
    ) THEN
        RAISE EXCEPTION 'ErrLeaseEpochStale: job % epoch % no longer current', _job, _epoch
            USING ERRCODE = 'P0001';
    END IF;
END;
$$;


--
-- Name: audit_event_row_hash(text, text, timestamp with time zone, text, text, text, text, jsonb, text, text); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.audit_event_row_hash(p_prev_hash text, p_id text, p_occurred_at timestamp with time zone, p_actor_id text, p_action text, p_resource_type text, p_resource_id text, p_payload jsonb, p_trace_id text, p_tenant_id text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT encode(
    digest(
      concat_ws(
        E'\x1f',
        COALESCE(p_prev_hash, ''),
        COALESCE(p_id, ''),
        to_char(p_occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
        COALESCE(p_actor_id, ''),
        COALESCE(p_action, ''),
        COALESCE(p_resource_type, ''),
        COALESCE(p_resource_id, ''),
        COALESCE(p_payload::text, '{}'),
        COALESCE(p_trace_id, ''),
        COALESCE(p_tenant_id, '')
      ),
      'sha256'
    ),
    'hex'
  )
$$;


--
-- Name: grant_area_membership(uuid, text, text, text, text); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.grant_area_membership(_tenant_id uuid, _user_id text, _area_code text, _role text, _granted_by text) RETURNS uuid
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $_$
DECLARE
  _existing_from  TIMESTAMPTZ;
  _now            TIMESTAMPTZ := pg_catalog.clock_timestamp();
  _correlation_id UUID        := pg_catalog.gen_random_uuid();
BEGIN
  -- -- Input validation (before any query) ----------------------------------
  IF _user_id !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid user_id: %', _user_id USING ERRCODE = '22023';
  END IF;
  IF _area_code !~ '^[A-Z0-9_]+$' THEN
    RAISE EXCEPTION 'invalid area_code: %', _area_code USING ERRCODE = '22023';
  END IF;
  -- Role values must match the CHECK constraint on user_process_areas (0125).
  IF _role NOT IN ('viewer', 'editor', 'reviewer', 'approver') THEN
    RAISE EXCEPTION 'invalid role: % (allowed: viewer, editor, reviewer, approver)', _role
      USING ERRCODE = '22023';
  END IF;
  IF _granted_by !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid granted_by: %', _granted_by USING ERRCODE = '22023';
  END IF;

  -- -- Idempotency check ----------------------------------------------------
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

  -- -- Insert membership row ------------------------------------------------
  INSERT INTO public.user_process_areas
    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
  VALUES
    (_user_id, _tenant_id, _area_code, _role, _now, NULL, _granted_by, NULL);

  -- -- Emit governance event ------------------------------------------------
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


--
-- Name: heartbeat_lease(text, text, bigint); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.heartbeat_lease(_job text, _leader text, _epoch bigint) RETURNS boolean
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $$
BEGIN
    UPDATE metaldocs.job_leases
    SET
        heartbeat_at = now(),
        expires_at = now() + interval '5 minutes'
    WHERE job_name = _job
      AND leader_id = _leader
      AND lease_epoch = _epoch;

    RETURN FOUND;
END;
$$;


--
-- Name: prevent_published_template_mutation(); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.prevent_published_template_mutation() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF OLD.is_published = true AND NEW.content_blocks IS DISTINCT FROM OLD.content_blocks THEN
    RAISE EXCEPTION 'Cannot modify content_blocks of a published template version (id=%)', OLD.id;
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: release_lease(text, text, bigint); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.release_lease(_job text, _leader text, _epoch bigint) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $$
BEGIN
    -- Mark the lease as immediately expired rather than deleting the row.
    -- The next acquire_lease call will hit the "expired" branch and increment
    -- the epoch, ensuring lease_epoch is strictly monotonic for the lifetime
    -- of the job_name key.
    UPDATE metaldocs.job_leases
    SET expires_at = now() - interval '1 second'
    WHERE job_name  = _job
      AND leader_id = _leader
      AND lease_epoch = _epoch;
END;
$$;


--
-- Name: revoke_area_membership(uuid, text, text, text, text); Type: FUNCTION; Schema: metaldocs; Owner: -
--

CREATE FUNCTION metaldocs.revoke_area_membership(_tenant_id uuid, _user_id text, _area_code text, _role text, _revoked_by text) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'metaldocs', 'pg_temp'
    AS $_$
DECLARE
  _rows_affected  INT;
  _now            TIMESTAMPTZ := pg_catalog.clock_timestamp();
BEGIN
  -- -- Input validation (before any query) ----------------------------------
  IF _user_id !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid user_id: %', _user_id USING ERRCODE = '22023';
  END IF;
  IF _area_code !~ '^[A-Z0-9_]+$' THEN
    RAISE EXCEPTION 'invalid area_code: %', _area_code USING ERRCODE = '22023';
  END IF;
  IF _role NOT IN ('viewer', 'editor', 'reviewer', 'approver') THEN
    RAISE EXCEPTION 'invalid role: % (allowed: viewer, editor, reviewer, approver)', _role
      USING ERRCODE = '22023';
  END IF;
  IF _revoked_by !~ '^[a-z0-9_.@-]+$' THEN
    RAISE EXCEPTION 'invalid revoked_by: %', _revoked_by USING ERRCODE = '22023';
  END IF;

  -- -- Soft-delete: set effective_to + revoked_by ---------------------------
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

  -- -- Emit governance event ------------------------------------------------
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


--
-- Name: check_document_tenant_consistency(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_document_tenant_consistency() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  cd_tenant UUID;
BEGIN
  IF TG_OP = 'UPDATE' AND OLD.controlled_document_id IS NOT DISTINCT FROM NEW.controlled_document_id THEN
    RETURN NEW;
  END IF;

  IF NEW.controlled_document_id IS NOT NULL THEN
    SELECT tenant_id INTO cd_tenant
      FROM controlled_documents WHERE id = NEW.controlled_document_id;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'controlled_document_id % does not exist in controlled_documents',
        NEW.controlled_document_id;
    END IF;
    IF cd_tenant IS DISTINCT FROM NEW.tenant_id THEN
      RAISE EXCEPTION 'tenant mismatch between document (%) and controlled_document (%)',
        NEW.tenant_id, cd_tenant;
    END IF;
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: enforce_capability_asserted(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_capability_asserted() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  v_bypass        TEXT;
  v_asserted_raw  TEXT;
  v_asserted      JSONB;
  v_required_caps TEXT[];   -- one or more acceptable caps for this table/op
  v_tenant_id     UUID;
  v_cap_found     BOOLEAN := FALSE;
  v_element       JSONB;
BEGIN
  -- ---- Determine required capability set for this table/operation. --------
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
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      -- Either lifecycle cap is acceptable.
      v_required_caps := ARRAY['registry.obsolete', 'registry.supersede'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      -- document_families has no tenant_id column.
      v_tenant_id     := NULL;

    WHEN TG_TABLE_NAME = 'templates_v2_template' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'templates_v2_template_version' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      -- template_version has no tenant_id column.
      v_tenant_id     := NULL;

    ELSE
      -- Unknown table — conservative pass-through.
      RETURN NEW;
  END CASE;

  -- ---- Bypass path. -------------------------------------------------------
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
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %',
                      v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

  -- ---- Read asserted_caps GUC. --------------------------------------------
  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);
  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %',
                    v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  BEGIN
    v_asserted := v_asserted_raw::JSONB;
  EXCEPTION WHEN invalid_text_representation OR others THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)',
                    v_required_caps
      USING ERRCODE = 'P0001';
  END;

  IF jsonb_typeof(v_asserted) <> 'array' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)',
                    v_required_caps
      USING ERRCODE = 'P0001';
  END IF;

  -- ---- Scan for any required cap. ----------------------------------------
  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP
    IF (v_element->>'cap') = ANY(v_required_caps) THEN
      v_cap_found := TRUE;
      EXIT;
    END IF;
  END LOOP;

  IF NOT v_cap_found THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %',
                    v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  RETURN NEW;
END;
$$;


--
-- Name: enforce_document_transition(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_document_transition() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  cancel_instance_id TEXT;
BEGIN
  IF OLD.status IS DISTINCT FROM NEW.status THEN
    -- Check cancel path: under_review?draft allowed if cancel GUC is set.
    IF OLD.status = 'under_review' AND NEW.status = 'draft' THEN
      cancel_instance_id := current_setting('metaldocs.cancel_in_progress', true);
      IF cancel_instance_id IS NULL OR cancel_instance_id = '' THEN
        RAISE EXCEPTION 'illegal status transition % -> % (set metaldocs.cancel_in_progress GUC to authorise cancel rollback)',
          OLD.status, NEW.status
          USING ERRCODE = 'check_violation';
      END IF;
      RETURN NEW;
    END IF;

    IF NOT (
      -- Spec 2 graph
      (OLD.status = 'draft'        AND NEW.status =  'under_review') OR
      (OLD.status = 'under_review' AND NEW.status IN ('approved','rejected')) OR
      (OLD.status = 'rejected'     AND NEW.status =  'draft') OR
      (OLD.status = 'approved'     AND NEW.status IN ('published','scheduled','draft')) OR
      (OLD.status = 'scheduled'    AND NEW.status IN ('published','draft')) OR
      (OLD.status = 'published'    AND NEW.status IN ('superseded','obsolete')) OR
      (OLD.status = 'superseded'   AND NEW.status =  'obsolete')
      -- Note: compat window (draft?finalized, finalized?archived) removed by 0142_disable_legacy_compat.sql
    ) THEN
      RAISE EXCEPTION 'illegal status transition % -> %', OLD.status, NEW.status
        USING ERRCODE = 'check_violation';
    END IF;
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: enforce_placeholder_value_tenant_consistent(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_placeholder_value_tenant_consistent() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE doc_tenant UUID;
BEGIN
    SELECT tenant_id INTO doc_tenant FROM documents WHERE id = NEW.revision_id;
    IF doc_tenant IS NULL THEN
        RAISE EXCEPTION 'document % not found', NEW.revision_id USING ERRCODE = 'foreign_key_violation';
    END IF;
    IF doc_tenant <> NEW.tenant_id THEN
        RAISE EXCEPTION 'tenant mismatch: document=% value=%', doc_tenant, NEW.tenant_id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: enforce_revision_version_monotonic(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_revision_version_monotonic() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  IF NEW.revision_version < OLD.revision_version THEN
    RAISE EXCEPTION 'revision_version cannot decrease (OLD=%, NEW=%)',
                    OLD.revision_version, NEW.revision_version
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: enforce_route_immutable(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_route_immutable() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  IF EXISTS (
    SELECT 1
      FROM public.approval_instances
     WHERE route_id = OLD.id
  ) THEN
    RAISE EXCEPTION 'ErrRouteInUse: route % is referenced by one or more approval instances and cannot be modified', OLD.id
      USING ERRCODE = 'P0001';
  END IF;
  RETURN NEW; -- for UPDATE; DELETE triggers ignore return value
END;
$$;


--
-- Name: enforce_signoff_eligibility(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_signoff_eligibility() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    eligible jsonb;
BEGIN
    SELECT eligible_actor_ids INTO eligible
    FROM   approval_stage_instances
    WHERE  id = NEW.stage_instance_id;

    IF eligible IS NULL OR NOT eligible @> to_jsonb(NEW.actor_user_id::text) THEN
        RAISE EXCEPTION 'signoff: actor % is not in eligible_actor_ids for stage %',
            NEW.actor_user_id, NEW.stage_instance_id
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: enforce_signoff_sod(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_signoff_sod() RETURNS trigger
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

  RETURN NEW;
END;
$$;


--
-- Name: enforce_signoff_tenant_consistent(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_signoff_tenant_consistent() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  instance_tenant UUID;
BEGIN
  SELECT tenant_id INTO instance_tenant
    FROM public.approval_instances
   WHERE id = NEW.approval_instance_id;

  IF instance_tenant IS DISTINCT FROM NEW.actor_tenant_id THEN
    RAISE EXCEPTION 'cross-tenant signoff rejected (instance tenant %, actor tenant %)',
                    instance_tenant, NEW.actor_tenant_id
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;


--
-- Name: enforce_snapshot_on_submit(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_snapshot_on_submit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.status IN ('under_review','approved','scheduled','published')
       AND (NEW.placeholder_schema_snapshot IS NULL
         OR NEW.placeholder_schema_hash IS NULL
         OR NEW.composition_config_snapshot IS NULL
         OR NEW.composition_config_hash IS NULL
         OR NEW.body_docx_snapshot_s3_key IS NULL
         OR NEW.body_docx_hash IS NULL) THEN
        RAISE EXCEPTION 'documents.% snapshot columns required for status=%',
            NEW.id, NEW.status
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: enforce_user_process_areas_update_contract(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_user_process_areas_update_contract() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  IF NEW.tenant_id      IS DISTINCT FROM OLD.tenant_id      OR
     NEW.user_id        IS DISTINCT FROM OLD.user_id        OR
     NEW.area_code      IS DISTINCT FROM OLD.area_code      OR
     NEW.role           IS DISTINCT FROM OLD.role           OR
     NEW.effective_from IS DISTINCT FROM OLD.effective_from OR
     NEW.granted_by     IS DISTINCT FROM OLD.granted_by     THEN
    RAISE EXCEPTION 'identity columns are immutable on user_process_areas'
      USING ERRCODE = 'check_violation';
  END IF;
  IF OLD.effective_to IS NOT NULL AND NEW.effective_to IS NULL THEN
    RAISE EXCEPTION 'cannot un-revoke membership (re-grant creates new row)'
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: grant_area_membership(uuid, text, text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.grant_area_membership(_tenant_id uuid, _user_id text, _area_code text, _role text, _granted_by text) RETURNS uuid
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  session_actor  TEXT := pg_catalog.current_setting('metaldocs.actor_id', true);
  session_cap    TEXT := pg_catalog.current_setting('metaldocs.verified_capability', true);
  actor_tenant   UUID;
BEGIN
  IF session_actor IS NULL OR session_actor = '' OR session_actor IS DISTINCT FROM _granted_by THEN
    RAISE EXCEPTION 'session actor context missing or mismatched'
      USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF session_cap IS NULL OR session_cap <> 'workflow.route.edit' THEN
    RAISE EXCEPTION 'session capability context missing or wrong'
      USING ERRCODE = 'insufficient_privilege';
  END IF;
  SELECT tenant_id INTO actor_tenant
    FROM metaldocs.iam_users WHERE user_id = _granted_by;
  IF actor_tenant IS DISTINCT FROM _tenant_id THEN
    RAISE EXCEPTION 'granted_by must belong to same tenant'
      USING ERRCODE = 'check_violation';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM metaldocs.iam_users
     WHERE user_id = _granted_by AND deactivated_at IS NULL
  ) THEN
    RAISE EXCEPTION 'granted_by must be active user'
      USING ERRCODE = 'check_violation';
  END IF;
  INSERT INTO public.user_process_areas
    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
    VALUES (_user_id, _tenant_id, _area_code, _role,
            pg_catalog.clock_timestamp(), NULL, _granted_by, NULL);
  RETURN pg_catalog.gen_random_uuid();
END;
$$;


--
-- Name: reject_code_update(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.reject_code_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF NEW.code IS DISTINCT FROM OLD.code THEN
    RAISE EXCEPTION 'code column is immutable (table=%, old=%, new=%)',
      TG_TABLE_NAME, OLD.code, NEW.code
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: reject_families_code_update(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.reject_families_code_update() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  IF NEW.code IS DISTINCT FROM OLD.code THEN
    RAISE EXCEPTION 'document_families.code is immutable'
      USING ERRCODE = 'P0002';
  END IF;
  RETURN NEW;
END;
$$;


--
-- Name: reject_signoff_update(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.reject_signoff_update() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  RAISE EXCEPTION 'approval_signoffs rows are immutable'
    USING ERRCODE = 'check_violation';
END;
$$;


--
-- Name: reject_user_process_areas_delete(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.reject_user_process_areas_delete() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
BEGIN
  RAISE EXCEPTION 'user_process_areas rows cannot be deleted (revoke via UPDATE effective_to)'
    USING ERRCODE = 'check_violation';
END;
$$;


--
-- Name: revoke_area_membership(uuid, text, text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.revoke_area_membership(_tenant_id uuid, _user_id text, _area_code text, _role text, _revoked_by text) RETURNS uuid
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  session_actor  TEXT := pg_catalog.current_setting('metaldocs.actor_id', true);
  session_cap    TEXT := pg_catalog.current_setting('metaldocs.verified_capability', true);
  actor_tenant   UUID;
  rows_affected  INT;
BEGIN
  IF session_actor IS NULL OR session_actor = '' OR session_actor IS DISTINCT FROM _revoked_by THEN
    RAISE EXCEPTION 'session actor context missing or mismatched'
      USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF session_cap IS NULL OR session_cap <> 'workflow.route.edit' THEN
    RAISE EXCEPTION 'session capability context missing or wrong'
      USING ERRCODE = 'insufficient_privilege';
  END IF;
  SELECT tenant_id INTO actor_tenant
    FROM metaldocs.iam_users WHERE user_id = _revoked_by;
  IF actor_tenant IS DISTINCT FROM _tenant_id THEN
    RAISE EXCEPTION 'revoked_by must belong to same tenant'
      USING ERRCODE = 'check_violation';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM metaldocs.iam_users
     WHERE user_id = _revoked_by AND deactivated_at IS NULL
  ) THEN
    RAISE EXCEPTION 'revoked_by must be active user'
      USING ERRCODE = 'check_violation';
  END IF;
  UPDATE public.user_process_areas
     SET effective_to = pg_catalog.clock_timestamp(),
         revoked_by   = _revoked_by
   WHERE tenant_id    = _tenant_id
     AND user_id      = _user_id
     AND area_code    = _area_code
     AND role         = _role
     AND effective_to IS NULL;
  GET DIAGNOSTICS rows_affected = ROW_COUNT;
  IF rows_affected = 0 THEN
    RAISE EXCEPTION 'no active membership to revoke'
      USING ERRCODE = 'no_data_found';
  END IF;
  RETURN pg_catalog.gen_random_uuid();
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_events; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.audit_events (
    id text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    actor_id text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    trace_id text NOT NULL,
    tenant_id text DEFAULT ''::text NOT NULL,
    audit_sequence bigint NOT NULL,
    prev_hash text DEFAULT ''::text NOT NULL,
    row_hash text DEFAULT ''::text NOT NULL
);


--
-- Name: audit_events_audit_sequence_seq; Type: SEQUENCE; Schema: metaldocs; Owner: -
--

CREATE SEQUENCE metaldocs.audit_events_audit_sequence_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_events_audit_sequence_seq; Type: SEQUENCE OWNED BY; Schema: metaldocs; Owner: -
--

ALTER SEQUENCE metaldocs.audit_events_audit_sequence_seq OWNED BY metaldocs.audit_events.audit_sequence;


--
-- Name: auth_identities; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.auth_identities (
    user_id text NOT NULL,
    username text NOT NULL,
    email text,
    password_hash text NOT NULL,
    password_algo text NOT NULL,
    must_change_password boolean DEFAULT false NOT NULL,
    last_login_at timestamp with time zone,
    failed_login_attempts integer DEFAULT 0 NOT NULL,
    locked_until timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: auth_sessions; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.auth_sessions (
    session_id text NOT NULL,
    user_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    ip_address text,
    user_agent text,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id text NOT NULL
);


--
-- Name: document_access_policies; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_access_policies (
    id bigint NOT NULL,
    subject_type text NOT NULL,
    subject_id text NOT NULL,
    resource_scope text NOT NULL,
    resource_id text NOT NULL,
    capability text NOT NULL,
    effect text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_document_access_policies_capability CHECK ((capability = ANY (ARRAY['document.create'::text, 'document.view'::text, 'document.edit'::text, 'document.upload_attachment'::text, 'document.change_workflow'::text, 'document.manage_permissions'::text]))),
    CONSTRAINT chk_document_access_policies_effect CHECK ((effect = ANY (ARRAY['allow'::text, 'deny'::text]))),
    CONSTRAINT chk_document_access_policies_resource_scope CHECK ((resource_scope = ANY (ARRAY['document'::text, 'document_type'::text, 'area'::text]))),
    CONSTRAINT chk_document_access_policies_subject_type CHECK ((subject_type = ANY (ARRAY['user'::text, 'role'::text, 'group'::text])))
);


--
-- Name: document_access_policies_id_seq; Type: SEQUENCE; Schema: metaldocs; Owner: -
--

CREATE SEQUENCE metaldocs.document_access_policies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: document_access_policies_id_seq; Type: SEQUENCE OWNED BY; Schema: metaldocs; Owner: -
--

ALTER SEQUENCE metaldocs.document_access_policies_id_seq OWNED BY metaldocs.document_access_policies.id;


--
-- Name: document_attachments; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_attachments (
    id text NOT NULL,
    document_id text NOT NULL,
    file_name text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_key text NOT NULL,
    uploaded_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_attachments_size_bytes_check CHECK ((size_bytes > 0))
);


--
-- Name: document_collaboration_presence; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_collaboration_presence (
    document_id text NOT NULL,
    user_id text NOT NULL,
    display_name text NOT NULL,
    last_seen_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_departments; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_departments (
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_edit_locks; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_edit_locks (
    document_id text NOT NULL,
    locked_by text NOT NULL,
    display_name text NOT NULL,
    lock_reason text DEFAULT ''::text NOT NULL,
    acquired_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_document_edit_locks_expiry CHECK ((expires_at > acquired_at))
);


--
-- Name: document_families; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_families (
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_images; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    sha256 text NOT NULL,
    mime_type text NOT NULL,
    byte_size integer NOT NULL,
    bytes bytea NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_images_byte_size_check CHECK ((byte_size > 0))
);


--
-- Name: document_process_areas; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_process_areas (
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    parent_code text,
    owner_user_id text,
    default_approver_role text,
    archived_at timestamp with time zone,
    CONSTRAINT area_code_format CHECK ((code ~ '^[a-z][a-z0-9_-]{1,63}$'::text))
);


--
-- Name: document_profile_governance; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_profile_governance (
    profile_code text NOT NULL,
    workflow_profile text DEFAULT 'standard_approval'::text NOT NULL,
    review_interval_days integer NOT NULL,
    approval_required boolean DEFAULT true NOT NULL,
    retention_days integer DEFAULT 0 NOT NULL,
    validity_days integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_profile_governance_retention_days_check CHECK ((retention_days >= 0)),
    CONSTRAINT document_profile_governance_review_interval_days_check CHECK ((review_interval_days > 0)),
    CONSTRAINT document_profile_governance_validity_days_check CHECK ((validity_days >= 0))
);


--
-- Name: document_profile_schema_versions; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_profile_schema_versions (
    profile_code text NOT NULL,
    version integer NOT NULL,
    metadata_rules_json jsonb DEFAULT '[]'::jsonb NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    content_schema_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT document_profile_schema_versions_version_check CHECK ((version > 0))
);


--
-- Name: document_profile_template_defaults; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_profile_template_defaults (
    profile_code text NOT NULL,
    template_key text NOT NULL,
    template_version integer NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_profiles; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_profiles (
    code text NOT NULL,
    family_code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    review_interval_days integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    alias text NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    default_template_version_id uuid,
    owner_user_id text,
    editable_by_role text DEFAULT 'admin'::text NOT NULL,
    archived_at timestamp with time zone,
    CONSTRAINT chk_document_profiles_alias_length CHECK (((char_length(alias) >= 1) AND (char_length(alias) <= 24))),
    CONSTRAINT document_profiles_review_interval_days_check CHECK ((review_interval_days > 0)),
    CONSTRAINT profile_code_format CHECK ((code ~ '^[a-z][a-z0-9_-]{1,63}$'::text))
);


--
-- Name: document_sequences; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_sequences (
    profile_code text NOT NULL,
    next_value integer NOT NULL,
    CONSTRAINT document_sequences_next_value_check CHECK ((next_value > 0))
);


--
-- Name: document_subjects; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_subjects (
    code text NOT NULL,
    process_area_code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_template_assignments; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_template_assignments (
    document_id text NOT NULL,
    template_key text NOT NULL,
    template_version integer NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_template_versions; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_template_versions (
    template_key text NOT NULL,
    version integer NOT NULL,
    profile_code text NOT NULL,
    schema_version integer NOT NULL,
    name text NOT NULL,
    definition_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    editor text DEFAULT 'ckeditor5'::text NOT NULL,
    content_format text DEFAULT 'html'::text NOT NULL,
    body_html text DEFAULT ''::text NOT NULL,
    export_config jsonb,
    status text DEFAULT 'published'::text NOT NULL
);


--
-- Name: document_template_versions_mddm; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_template_versions_mddm (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_id uuid NOT NULL,
    version integer NOT NULL,
    mddm_version integer NOT NULL,
    content_blocks jsonb NOT NULL,
    content_hash text NOT NULL,
    is_published boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_template_versions_mddm_mddm_version_check CHECK ((mddm_version >= 1)),
    CONSTRAINT document_template_versions_mddm_version_check CHECK ((version >= 1))
);


--
-- Name: document_type_schema_versions; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_type_schema_versions (
    type_key text NOT NULL,
    version integer NOT NULL,
    schema_json jsonb NOT NULL,
    governance_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_types; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_types (
    code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    review_interval_days integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    type_key text,
    family_key text DEFAULT ''::text NOT NULL,
    active_version integer DEFAULT 1 NOT NULL,
    CONSTRAINT document_types_review_interval_days_check CHECK ((review_interval_days > 0))
);


--
-- Name: document_version_images; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_version_images (
    document_version_id uuid NOT NULL,
    image_id uuid NOT NULL
);


--
-- Name: document_versions; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_versions (
    document_id text NOT NULL,
    version_number integer NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    content_hash text NOT NULL,
    change_summary text DEFAULT ''::text NOT NULL,
    content_source text DEFAULT 'native'::text NOT NULL,
    native_content jsonb,
    docx_storage_key text,
    pdf_storage_key text,
    text_content text,
    file_size_bytes bigint,
    original_filename text,
    page_count integer,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('portuguese'::regconfig, COALESCE(text_content, ''::text))) STORED,
    body_blocks jsonb DEFAULT '[]'::jsonb,
    values_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    template_key text,
    template_version integer,
    renderer_pin jsonb,
    release_artifact_key text,
    canonical_mddm_snapshot jsonb
);


--
-- Name: COLUMN document_versions.renderer_pin; Type: COMMENT; Schema: metaldocs; Owner: -
--

COMMENT ON COLUMN metaldocs.document_versions.renderer_pin IS 'Frozen renderer inputs captured at release time: {renderer_version, layout_ir_hash, template_key, template_version, pinned_at}. NULL for drafts.';


--
-- Name: COLUMN document_versions.release_artifact_key; Type: COMMENT; Schema: metaldocs; Owner: -
--

COMMENT ON COLUMN metaldocs.document_versions.release_artifact_key IS 'Storage key for the immutable DOCX artifact generated at release time';


--
-- Name: COLUMN document_versions.canonical_mddm_snapshot; Type: COMMENT; Schema: metaldocs; Owner: -
--

COMMENT ON COLUMN metaldocs.document_versions.canonical_mddm_snapshot IS 'Frozen MDDM envelope JSON captured at release time (post-migration, post-canonicalization)';


--
-- Name: document_versions_mddm; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.document_versions_mddm (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id text NOT NULL,
    version_number integer NOT NULL,
    revision_label text NOT NULL,
    status metaldocs.mddm_version_status NOT NULL,
    content_blocks jsonb,
    docx_bytes bytea,
    template_ref jsonb,
    content_hash text,
    revision_diff jsonb,
    change_summary text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    approved_at timestamp with time zone,
    approved_by text,
    CONSTRAINT document_versions_mddm_version_number_check CHECK ((version_number >= 1))
);


--
-- Name: documents; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.documents (
    id text NOT NULL,
    title text NOT NULL,
    owner_id text NOT NULL,
    classification text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    document_type_code text NOT NULL,
    business_unit text NOT NULL,
    department text NOT NULL,
    tags jsonb DEFAULT '[]'::jsonb NOT NULL,
    effective_at timestamp with time zone,
    expiry_at timestamp with time zone,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    document_profile_code text NOT NULL,
    document_family_code text NOT NULL,
    process_area_code text,
    subject_code text,
    profile_schema_version integer DEFAULT 1 NOT NULL,
    document_sequence integer NOT NULL,
    document_code text NOT NULL,
    document_type_key text DEFAULT ''::text NOT NULL,
    document_type_version integer DEFAULT 1 NOT NULL
);


--
-- Name: iam_group_members; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.iam_group_members (
    group_id uuid NOT NULL,
    user_id text NOT NULL,
    tenant_id uuid NOT NULL,
    granted_at timestamp with time zone DEFAULT now() NOT NULL,
    granted_by text
);


--
-- Name: iam_group_roles; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.iam_group_roles (
    group_id uuid NOT NULL,
    role text NOT NULL
);


--
-- Name: iam_groups; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.iam_groups (
    id uuid NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: iam_user_roles; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.iam_user_roles (
    user_id text NOT NULL,
    role_code text NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    assigned_by text,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    CONSTRAINT chk_iam_user_roles_role_code CHECK ((role_code = ANY (ARRAY['system_admin'::text, 'approver'::text, 'author'::text, 'editor'::text, 'viewer'::text])))
);


--
-- Name: iam_users; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.iam_users (
    user_id text NOT NULL,
    display_name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    deactivated_at timestamp with time zone,
    CONSTRAINT iam_users_deactivated_after_created CHECK (((deactivated_at IS NULL) OR (deactivated_at >= created_at)))
);


--
-- Name: idempotency_keys; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.idempotency_keys (
    tenant_id uuid NOT NULL,
    actor_user_id text NOT NULL,
    route_template text NOT NULL,
    key text NOT NULL,
    payload_hash text NOT NULL,
    response_status integer NOT NULL,
    response_body jsonb NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT idempotency_keys_status_check CHECK ((status = ANY (ARRAY['in_flight'::text, 'completed'::text, 'failed'::text])))
);


--
-- Name: job_leases; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.job_leases (
    job_name text NOT NULL,
    leader_id text NOT NULL,
    lease_epoch bigint DEFAULT 0 NOT NULL,
    acquired_at timestamp with time zone DEFAULT now() NOT NULL,
    heartbeat_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


--
-- Name: mddm_shadow_diff_events; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.mddm_shadow_diff_events (
    id bigint NOT NULL,
    document_id character varying(64) NOT NULL,
    version_number integer NOT NULL,
    user_id_hash character varying(64) NOT NULL,
    current_xml_hash character varying(64) NOT NULL,
    shadow_xml_hash character varying(64) NOT NULL,
    diff_summary jsonb DEFAULT '{}'::jsonb NOT NULL,
    current_duration_ms integer DEFAULT 0 NOT NULL,
    shadow_duration_ms integer DEFAULT 0 NOT NULL,
    shadow_error text,
    recorded_at timestamp with time zone DEFAULT now() NOT NULL,
    trace_id character varying(64)
);


--
-- Name: TABLE mddm_shadow_diff_events; Type: COMMENT; Schema: metaldocs; Owner: -
--

COMMENT ON TABLE metaldocs.mddm_shadow_diff_events IS 'Phase 1 shadow-test telemetry: compares docgen DOCX against the new client-side DOCX for browser_editor documents. Rows are append-only. user_id_hash is a salted SHA-256 so individual users cannot be identified from the raw table.';


--
-- Name: mddm_shadow_diff_events_id_seq; Type: SEQUENCE; Schema: metaldocs; Owner: -
--

CREATE SEQUENCE metaldocs.mddm_shadow_diff_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mddm_shadow_diff_events_id_seq; Type: SEQUENCE OWNED BY; Schema: metaldocs; Owner: -
--

ALTER SEQUENCE metaldocs.mddm_shadow_diff_events_id_seq OWNED BY metaldocs.mddm_shadow_diff_events.id;


--
-- Name: notifications; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.notifications (
    id text NOT NULL,
    recipient_user_id text NOT NULL,
    event_type text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    status text NOT NULL,
    idempotency_key text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    read_at timestamp with time zone,
    CONSTRAINT notifications_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'SENT'::text, 'READ'::text])))
);


--
-- Name: outbox_events; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.outbox_events (
    event_id text NOT NULL,
    event_type text NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    version integer NOT NULL,
    idempotency_key text NOT NULL,
    producer text NOT NULL,
    trace_id text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    published_at timestamp with time zone,
    attempt_count integer DEFAULT 0 NOT NULL,
    last_error text,
    last_attempt_at timestamp with time zone,
    next_attempt_at timestamp with time zone,
    dead_lettered_at timestamp with time zone
);


--
-- Name: pdf_dispatch_outbox; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.pdf_dispatch_outbox (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    content_hash bytea NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    last_error text,
    claimed_at timestamp with time zone,
    next_retry_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    dispatched_at timestamp with time zone,
    CONSTRAINT pdf_dispatch_outbox_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'processing'::text, 'dispatched'::text, 'failed'::text])))
);


--
-- Name: role_capabilities; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.role_capabilities (
    role text NOT NULL,
    capability text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_cap_format CHECK ((capability ~ '^[a-z][a-z0-9._]*[a-z0-9]$'::text)),
    CONSTRAINT ck_cap_not_legacy CHECK ((capability <> ALL (ARRAY['document.finalize'::text, 'document.archive'::text])))
);


--
-- Name: template_audit_log; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.template_audit_log (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_key text NOT NULL,
    version integer,
    action text NOT NULL,
    actor_id text NOT NULL,
    diff_summary text,
    trace_id text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: template_drafts; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.template_drafts (
    template_key text NOT NULL,
    profile_code text NOT NULL,
    base_version integer DEFAULT 0 NOT NULL,
    name text NOT NULL,
    theme_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    meta_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    blocks_json jsonb NOT NULL,
    lock_version integer DEFAULT 1 NOT NULL,
    has_stripped_fields boolean DEFAULT false NOT NULL,
    stripped_fields_json jsonb,
    created_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    published_html text,
    draft_status text DEFAULT 'draft'::text NOT NULL,
    CONSTRAINT template_drafts_draft_status_check CHECK ((draft_status = ANY (ARRAY['draft'::text, 'pending_review'::text, 'published'::text])))
);


--
-- Name: user_process_areas; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_process_areas (
    user_id text NOT NULL,
    tenant_id uuid NOT NULL,
    area_code text NOT NULL,
    role text NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    granted_by text,
    revoked_by text,
    CONSTRAINT effective_interval_valid CHECK (((effective_to IS NULL) OR (effective_to > effective_from))),
    CONSTRAINT revoked_by_required_when_revoked CHECK ((((effective_to IS NULL) AND (revoked_by IS NULL)) OR ((effective_to IS NOT NULL) AND (revoked_by IS NOT NULL)))),
    CONSTRAINT user_process_areas_role_check CHECK ((role = ANY (ARRAY['viewer'::text, 'editor'::text, 'reviewer'::text, 'approver'::text, 'author'::text, 'signer'::text, 'area_admin'::text, 'qms_admin'::text])))
);


--
-- Name: user_process_areas; Type: VIEW; Schema: metaldocs; Owner: -
--

CREATE VIEW metaldocs.user_process_areas AS
 SELECT user_id,
    tenant_id,
    area_code,
    role,
    effective_from,
    effective_to,
    granted_by
   FROM public.user_process_areas;


--
-- Name: workflow_approvals; Type: TABLE; Schema: metaldocs; Owner: -
--

CREATE TABLE metaldocs.workflow_approvals (
    id text NOT NULL,
    document_id text NOT NULL,
    requested_by text NOT NULL,
    assigned_reviewer text NOT NULL,
    decision_by text,
    status text NOT NULL,
    request_reason text DEFAULT ''::text NOT NULL,
    decision_reason text,
    requested_at timestamp with time zone NOT NULL,
    decided_at timestamp with time zone,
    CONSTRAINT workflow_approvals_status_check CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'REJECTED'::text])))
);


--
-- Name: approval_instances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.approval_instances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    route_id uuid NOT NULL,
    route_version_snapshot integer NOT NULL,
    status text NOT NULL,
    submitted_by text NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    content_hash_at_submit text NOT NULL,
    idempotency_key text NOT NULL,
    CONSTRAINT approval_instances_status_check CHECK ((status = ANY (ARRAY['in_progress'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text])))
);


--
-- Name: approval_route_stages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.approval_route_stages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    route_id uuid NOT NULL,
    stage_order integer NOT NULL,
    name text NOT NULL,
    required_role text NOT NULL,
    required_capability text NOT NULL,
    area_code text NOT NULL,
    quorum text NOT NULL,
    quorum_m integer,
    on_eligibility_drift text NOT NULL,
    CONSTRAINT approval_route_stages_on_eligibility_drift_check CHECK ((on_eligibility_drift = ANY (ARRAY['reduce_quorum'::text, 'fail_stage'::text, 'keep_snapshot'::text]))),
    CONSTRAINT approval_route_stages_quorum_check CHECK ((quorum = ANY (ARRAY['any_1_of'::text, 'all_of'::text, 'm_of_n'::text]))),
    CONSTRAINT approval_route_stages_quorum_m_consistent CHECK ((((quorum = 'm_of_n'::text) AND (quorum_m IS NOT NULL) AND (quorum_m >= 1)) OR ((quorum <> 'm_of_n'::text) AND (quorum_m IS NULL)))),
    CONSTRAINT approval_route_stages_stage_order_check CHECK ((stage_order >= 1))
);


--
-- Name: approval_routes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.approval_routes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    profile_code text NOT NULL,
    name text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    active boolean DEFAULT true NOT NULL
);


--
-- Name: approval_signoffs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.approval_signoffs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    approval_instance_id uuid NOT NULL,
    stage_instance_id uuid NOT NULL,
    actor_user_id text NOT NULL,
    actor_tenant_id uuid NOT NULL,
    decision text NOT NULL,
    comment text,
    signed_at timestamp with time zone DEFAULT now() NOT NULL,
    signature_method text NOT NULL,
    signature_payload jsonb NOT NULL,
    content_hash text NOT NULL,
    actor_display_name_snapshot text,
    CONSTRAINT approval_signoffs_decision_check CHECK ((decision = ANY (ARRAY['approve'::text, 'reject'::text])))
);


--
-- Name: approval_stage_instances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.approval_stage_instances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    approval_instance_id uuid NOT NULL,
    stage_order integer NOT NULL,
    name_snapshot text NOT NULL,
    required_role_snapshot text NOT NULL,
    required_capability_snapshot text NOT NULL,
    area_code_snapshot text NOT NULL,
    quorum_snapshot text NOT NULL,
    quorum_m_snapshot integer,
    on_eligibility_drift_snapshot text NOT NULL,
    eligible_actor_ids jsonb NOT NULL,
    effective_denominator integer,
    status text NOT NULL,
    opened_at timestamp with time zone,
    completed_at timestamp with time zone,
    CONSTRAINT approval_stage_instances_on_eligibility_drift_snapshot_check CHECK ((on_eligibility_drift_snapshot = ANY (ARRAY['reduce_quorum'::text, 'fail_stage'::text, 'keep_snapshot'::text]))),
    CONSTRAINT approval_stage_instances_quorum_snapshot_check CHECK ((quorum_snapshot = ANY (ARRAY['any_1_of'::text, 'all_of'::text, 'm_of_n'::text]))),
    CONSTRAINT approval_stage_instances_stage_order_check CHECK ((stage_order >= 1)),
    CONSTRAINT approval_stage_instances_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'active'::text, 'completed'::text, 'skipped'::text, 'rejected_here'::text, 'cancelled'::text])))
);


--
-- Name: autosave_pending_uploads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.autosave_pending_uploads (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    session_id uuid NOT NULL,
    document_id uuid NOT NULL,
    base_revision_id uuid NOT NULL,
    content_hash text NOT NULL,
    storage_key text NOT NULL,
    presigned_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);


--
-- Name: cd_sequence_counters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cd_sequence_counters (
    tenant_id uuid NOT NULL,
    profile_code text NOT NULL,
    process_area_code text NOT NULL,
    next_seq integer DEFAULT 1 NOT NULL
);


--
-- Name: controlled_document_area_grants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controlled_document_area_grants (
    tenant_id uuid NOT NULL,
    controlled_document_id uuid NOT NULL,
    area_code text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: controlled_document_user_grants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controlled_document_user_grants (
    tenant_id uuid NOT NULL,
    controlled_document_id uuid NOT NULL,
    user_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: controlled_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controlled_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    profile_code text NOT NULL,
    process_area_code text NOT NULL,
    department_code text,
    code text NOT NULL,
    sequence_num integer,
    title text NOT NULL,
    owner_user_id text NOT NULL,
    override_template_version_id uuid,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    visibility_scope text DEFAULT 'company'::text NOT NULL,
    CONSTRAINT controlled_document_code_format CHECK (((length(code) >= 2) AND (length(code) <= 100))),
    CONSTRAINT controlled_documents_status_check CHECK ((status = ANY (ARRAY['active'::text, 'obsolete'::text, 'superseded'::text]))),
    CONSTRAINT controlled_documents_visibility_scope_check CHECK ((visibility_scope = ANY (ARRAY['company'::text, 'restricted'::text])))
);


--
-- Name: document_checkpoints; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_checkpoints (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    version_num integer NOT NULL,
    label text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL
);


--
-- Name: document_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_comments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    library_comment_id integer NOT NULL,
    parent_library_id integer,
    author_id text NOT NULL,
    author_display text NOT NULL,
    content_json jsonb NOT NULL,
    resolved_at timestamp with time zone,
    resolved_by text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_exports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_exports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    composite_hash bytea NOT NULL,
    storage_key text NOT NULL,
    size_bytes bigint NOT NULL,
    paper_size text DEFAULT 'A4'::text NOT NULL,
    landscape boolean DEFAULT false NOT NULL,
    docgen_v2_ver text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_exports_composite_hash_check CHECK ((octet_length(composite_hash) = 32)),
    CONSTRAINT document_exports_size_bytes_check CHECK ((size_bytes > 0))
);


--
-- Name: document_placeholder_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_placeholder_values (
    tenant_id uuid NOT NULL,
    revision_id uuid NOT NULL,
    placeholder_id text NOT NULL,
    value_text text,
    value_typed jsonb,
    source text NOT NULL,
    computed_from text,
    resolver_version integer,
    inputs_hash bytea,
    validated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_placeholder_values_inputs_hash_check CHECK (((inputs_hash IS NULL) OR (octet_length(inputs_hash) = 32))),
    CONSTRAINT document_placeholder_values_source_check CHECK ((source = ANY (ARRAY['user'::text, 'computed'::text, 'default'::text])))
);


--
-- Name: document_revisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_revisions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    revision_num bigint NOT NULL,
    parent_revision_id uuid,
    session_id uuid NOT NULL,
    storage_key text NOT NULL,
    content_hash text NOT NULL,
    form_data_snapshot jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: document_revisions_revision_num_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.document_revisions_revision_num_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: document_revisions_revision_num_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.document_revisions_revision_num_seq OWNED BY public.document_revisions.revision_num;


--
-- Name: documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    template_version_id uuid NOT NULL,
    name text NOT NULL,
    status text NOT NULL,
    form_data_json jsonb NOT NULL,
    current_revision_id uuid,
    active_session_id uuid,
    finalized_at timestamp with time zone,
    archived_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    effective_from timestamp with time zone,
    effective_to timestamp with time zone,
    revision_number integer DEFAULT 1 NOT NULL,
    revision_version integer DEFAULT 0 NOT NULL,
    content_hash_at_submit text,
    placeholder_schema_snapshot jsonb,
    placeholder_schema_hash bytea,
    composition_config_snapshot jsonb,
    composition_config_hash bytea,
    body_docx_snapshot_s3_key text,
    body_docx_hash bytea,
    values_frozen_at timestamp with time zone,
    values_hash bytea,
    final_docx_s3_key text,
    content_hash bytea,
    final_pdf_s3_key text,
    pdf_hash bytea,
    pdf_generated_at timestamp with time zone,
    reconstruction_attempts jsonb DEFAULT '[]'::jsonb NOT NULL,
    controlled_document_id uuid,
    profile_code_snapshot text,
    process_area_code_snapshot text,
    code text,
    created_by_display_name_snapshot text,
    area_name_snapshot text,
    CONSTRAINT documents_body_docx_hash_len CHECK (((body_docx_hash IS NULL) OR (octet_length(body_docx_hash) = 32))),
    CONSTRAINT documents_composition_config_hash_len CHECK (((composition_config_hash IS NULL) OR (octet_length(composition_config_hash) = 32))),
    CONSTRAINT documents_content_hash_len CHECK (((content_hash IS NULL) OR (octet_length(content_hash) = 32))),
    CONSTRAINT documents_name_not_empty CHECK ((length(TRIM(BOTH FROM name)) > 0)),
    CONSTRAINT documents_pdf_hash_len CHECK (((pdf_hash IS NULL) OR (octet_length(pdf_hash) = 32))),
    CONSTRAINT documents_placeholder_schema_hash_len CHECK (((placeholder_schema_hash IS NULL) OR (octet_length(placeholder_schema_hash) = 32))),
    CONSTRAINT documents_reconstruction_attempts_is_array CHECK ((jsonb_typeof(reconstruction_attempts) = 'array'::text)),
    CONSTRAINT documents_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'finalized'::text, 'archived'::text, 'under_review'::text, 'approved'::text, 'rejected'::text, 'scheduled'::text, 'published'::text, 'superseded'::text, 'obsolete'::text]))),
    CONSTRAINT documents_values_hash_len CHECK (((values_hash IS NULL) OR (octet_length(values_hash) = 32)))
);


--
-- Name: editor_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.editor_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id uuid NOT NULL,
    user_id text NOT NULL,
    acquired_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    released_at timestamp with time zone,
    last_acknowledged_revision_id uuid NOT NULL,
    status text NOT NULL,
    CONSTRAINT editor_sessions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'expired'::text, 'released'::text, 'force_released'::text])))
);


--
-- Name: governance_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.governance_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    event_type text NOT NULL,
    actor_user_id text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    reason text,
    payload_json jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    dedupe_key text,
    correlation_id text
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text
);


--
-- Name: template_audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.template_audit_log (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    template_id uuid,
    template_version_id uuid,
    document_id uuid,
    action text NOT NULL,
    actor_user_id text NOT NULL,
    metadata_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: template_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.template_versions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    template_id uuid NOT NULL,
    version_num integer NOT NULL,
    status text NOT NULL,
    grammar_version integer DEFAULT 1 NOT NULL,
    docx_storage_key text NOT NULL,
    schema_storage_key text NOT NULL,
    docx_content_hash text NOT NULL,
    schema_content_hash text NOT NULL,
    published_at timestamp with time zone,
    published_by text,
    deprecated_at timestamp with time zone,
    lock_version integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    CONSTRAINT template_versions_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'published'::text, 'deprecated'::text])))
);


--
-- Name: templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text,
    current_published_version_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL
);


--
-- Name: templates_v2_approval_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.templates_v2_approval_config (
    template_id uuid NOT NULL,
    reviewer_role text,
    approver_role text NOT NULL
);


--
-- Name: templates_v2_audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.templates_v2_audit_log (
    id bigint NOT NULL,
    tenant_id text NOT NULL,
    template_id uuid NOT NULL,
    version_id uuid,
    actor_id text NOT NULL,
    action text NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: templates_v2_audit_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.templates_v2_audit_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: templates_v2_audit_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.templates_v2_audit_log_id_seq OWNED BY public.templates_v2_audit_log.id;


--
-- Name: templates_v2_template; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.templates_v2_template (
    id uuid NOT NULL,
    tenant_id text NOT NULL,
    doc_type_code text NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    areas text[] DEFAULT '{}'::text[] NOT NULL,
    visibility text NOT NULL,
    specific_areas text[] DEFAULT '{}'::text[] NOT NULL,
    latest_version integer DEFAULT 0 NOT NULL,
    published_version_id uuid,
    created_by text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone,
    system_owned boolean DEFAULT false NOT NULL
);


--
-- Name: templates_v2_template_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.templates_v2_template_version (
    id uuid NOT NULL,
    template_id uuid NOT NULL,
    version_number integer NOT NULL,
    status text NOT NULL,
    docx_storage_key text NOT NULL,
    content_hash text NOT NULL,
    metadata_schema jsonb NOT NULL,
    placeholder_schema jsonb NOT NULL,
    author_id text NOT NULL,
    pending_reviewer_role text,
    pending_approver_role text DEFAULT ''::text NOT NULL,
    reviewer_id text,
    approver_id text,
    submitted_at timestamp with time zone,
    reviewed_at timestamp with time zone,
    approved_at timestamp with time zone,
    published_at timestamp with time zone,
    obsoleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    lock_version integer DEFAULT 0 NOT NULL
);


--
-- Name: audit_events audit_sequence; Type: DEFAULT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.audit_events ALTER COLUMN audit_sequence SET DEFAULT nextval('metaldocs.audit_events_audit_sequence_seq'::regclass);


--
-- Name: document_access_policies id; Type: DEFAULT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_access_policies ALTER COLUMN id SET DEFAULT nextval('metaldocs.document_access_policies_id_seq'::regclass);


--
-- Name: mddm_shadow_diff_events id; Type: DEFAULT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.mddm_shadow_diff_events ALTER COLUMN id SET DEFAULT nextval('metaldocs.mddm_shadow_diff_events_id_seq'::regclass);


--
-- Name: document_revisions revision_num; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_revisions ALTER COLUMN revision_num SET DEFAULT nextval('public.document_revisions_revision_num_seq'::regclass);


--
-- Name: templates_v2_audit_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_audit_log ALTER COLUMN id SET DEFAULT nextval('public.templates_v2_audit_log_id_seq'::regclass);


--
-- Name: audit_events audit_events_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.audit_events
    ADD CONSTRAINT audit_events_pkey PRIMARY KEY (id);


--
-- Name: audit_events audit_events_prev_hash_shape; Type: CHECK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE metaldocs.audit_events
    ADD CONSTRAINT audit_events_prev_hash_shape CHECK (((prev_hash = ''::text) OR (prev_hash ~ '^[0-9a-f]{64}$'::text))) NOT VALID;


--
-- Name: audit_events audit_events_row_hash_shape; Type: CHECK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE metaldocs.audit_events
    ADD CONSTRAINT audit_events_row_hash_shape CHECK (((row_hash = ''::text) OR (row_hash ~ '^[0-9a-f]{64}$'::text))) NOT VALID;


--
-- Name: auth_identities auth_identities_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.auth_identities
    ADD CONSTRAINT auth_identities_pkey PRIMARY KEY (user_id);


--
-- Name: auth_sessions auth_sessions_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.auth_sessions
    ADD CONSTRAINT auth_sessions_pkey PRIMARY KEY (session_id);


--
-- Name: document_access_policies document_access_policies_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_access_policies
    ADD CONSTRAINT document_access_policies_pkey PRIMARY KEY (id);


--
-- Name: document_attachments document_attachments_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_attachments
    ADD CONSTRAINT document_attachments_pkey PRIMARY KEY (id);


--
-- Name: document_attachments document_attachments_storage_key_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_attachments
    ADD CONSTRAINT document_attachments_storage_key_key UNIQUE (storage_key);


--
-- Name: document_collaboration_presence document_collaboration_presence_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_collaboration_presence
    ADD CONSTRAINT document_collaboration_presence_pkey PRIMARY KEY (document_id, user_id);


--
-- Name: document_departments document_departments_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_departments
    ADD CONSTRAINT document_departments_pkey PRIMARY KEY (code);


--
-- Name: document_edit_locks document_edit_locks_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_edit_locks
    ADD CONSTRAINT document_edit_locks_pkey PRIMARY KEY (document_id);


--
-- Name: document_families document_families_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_families
    ADD CONSTRAINT document_families_pkey PRIMARY KEY (code);


--
-- Name: document_images document_images_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_images
    ADD CONSTRAINT document_images_pkey PRIMARY KEY (id);


--
-- Name: document_images document_images_sha256_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_images
    ADD CONSTRAINT document_images_sha256_key UNIQUE (sha256);


--
-- Name: document_process_areas document_process_areas_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_process_areas
    ADD CONSTRAINT document_process_areas_pkey PRIMARY KEY (code);


--
-- Name: document_profile_governance document_profile_governance_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profile_governance
    ADD CONSTRAINT document_profile_governance_pkey PRIMARY KEY (profile_code);


--
-- Name: document_profile_schema_versions document_profile_schema_versions_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profile_schema_versions
    ADD CONSTRAINT document_profile_schema_versions_pkey PRIMARY KEY (profile_code, version);


--
-- Name: document_profile_template_defaults document_profile_template_defaults_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profile_template_defaults
    ADD CONSTRAINT document_profile_template_defaults_pkey PRIMARY KEY (profile_code);


--
-- Name: document_profiles document_profiles_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profiles
    ADD CONSTRAINT document_profiles_pkey PRIMARY KEY (code);


--
-- Name: document_sequences document_sequences_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_sequences
    ADD CONSTRAINT document_sequences_pkey PRIMARY KEY (profile_code);


--
-- Name: document_subjects document_subjects_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_subjects
    ADD CONSTRAINT document_subjects_pkey PRIMARY KEY (code);


--
-- Name: document_template_assignments document_template_assignments_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_template_assignments
    ADD CONSTRAINT document_template_assignments_pkey PRIMARY KEY (document_id);


--
-- Name: document_template_versions_mddm document_template_versions_mddm_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_template_versions_mddm
    ADD CONSTRAINT document_template_versions_mddm_pkey PRIMARY KEY (id);


--
-- Name: document_template_versions_mddm document_template_versions_mddm_template_id_version_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_template_versions_mddm
    ADD CONSTRAINT document_template_versions_mddm_template_id_version_key UNIQUE (template_id, version);


--
-- Name: document_template_versions document_template_versions_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_template_versions
    ADD CONSTRAINT document_template_versions_pkey PRIMARY KEY (template_key, version);


--
-- Name: document_type_schema_versions document_type_schema_versions_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_type_schema_versions
    ADD CONSTRAINT document_type_schema_versions_pkey PRIMARY KEY (type_key, version);


--
-- Name: document_types document_types_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_types
    ADD CONSTRAINT document_types_pkey PRIMARY KEY (code);


--
-- Name: document_version_images document_version_images_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_version_images
    ADD CONSTRAINT document_version_images_pkey PRIMARY KEY (document_version_id, image_id);


--
-- Name: document_versions_mddm document_versions_mddm_document_id_version_number_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_versions_mddm
    ADD CONSTRAINT document_versions_mddm_document_id_version_number_key UNIQUE (document_id, version_number);


--
-- Name: document_versions_mddm document_versions_mddm_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_versions_mddm
    ADD CONSTRAINT document_versions_mddm_pkey PRIMARY KEY (id);


--
-- Name: document_versions document_versions_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_versions
    ADD CONSTRAINT document_versions_pkey PRIMARY KEY (document_id, version_number);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: iam_group_members iam_group_members_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_group_members
    ADD CONSTRAINT iam_group_members_pkey PRIMARY KEY (group_id, user_id);


--
-- Name: iam_group_roles iam_group_roles_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_group_roles
    ADD CONSTRAINT iam_group_roles_pkey PRIMARY KEY (group_id, role);


--
-- Name: iam_groups iam_groups_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_groups
    ADD CONSTRAINT iam_groups_pkey PRIMARY KEY (id);


--
-- Name: iam_groups iam_groups_tenant_id_name_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_groups
    ADD CONSTRAINT iam_groups_tenant_id_name_key UNIQUE (tenant_id, name);


--
-- Name: iam_user_roles iam_user_roles_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_user_roles
    ADD CONSTRAINT iam_user_roles_pkey PRIMARY KEY (user_id, role_code);


--
-- Name: iam_users iam_users_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_users
    ADD CONSTRAINT iam_users_pkey PRIMARY KEY (user_id);


--
-- Name: idempotency_keys idempotency_keys_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.idempotency_keys
    ADD CONSTRAINT idempotency_keys_pkey PRIMARY KEY (tenant_id, actor_user_id, route_template, key);


--
-- Name: job_leases job_leases_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.job_leases
    ADD CONSTRAINT job_leases_pkey PRIMARY KEY (job_name);


--
-- Name: mddm_shadow_diff_events mddm_shadow_diff_events_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.mddm_shadow_diff_events
    ADD CONSTRAINT mddm_shadow_diff_events_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_idempotency_key_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.notifications
    ADD CONSTRAINT notifications_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: outbox_events outbox_events_idempotency_key_key; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.outbox_events
    ADD CONSTRAINT outbox_events_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (event_id);


--
-- Name: pdf_dispatch_outbox pdf_dispatch_outbox_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.pdf_dispatch_outbox
    ADD CONSTRAINT pdf_dispatch_outbox_pkey PRIMARY KEY (id);


--
-- Name: role_capabilities role_capabilities_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.role_capabilities
    ADD CONSTRAINT role_capabilities_pkey PRIMARY KEY (role, capability);


--
-- Name: template_audit_log template_audit_log_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.template_audit_log
    ADD CONSTRAINT template_audit_log_pkey PRIMARY KEY (id);


--
-- Name: template_drafts template_drafts_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.template_drafts
    ADD CONSTRAINT template_drafts_pkey PRIMARY KEY (template_key);


--
-- Name: document_access_policies uq_document_access_policies_rule; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_access_policies
    ADD CONSTRAINT uq_document_access_policies_rule UNIQUE (subject_type, subject_id, resource_scope, resource_id, capability);


--
-- Name: iam_user_roles uq_iam_user_roles_user_tenant; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_user_roles
    ADD CONSTRAINT uq_iam_user_roles_user_tenant UNIQUE (tenant_id, user_id);


--
-- Name: pdf_dispatch_outbox ux_pdf_dispatch_outbox_revision; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.pdf_dispatch_outbox
    ADD CONSTRAINT ux_pdf_dispatch_outbox_revision UNIQUE (tenant_id, revision_id);


--
-- Name: workflow_approvals workflow_approvals_pkey; Type: CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.workflow_approvals
    ADD CONSTRAINT workflow_approvals_pkey PRIMARY KEY (id);


--
-- Name: approval_instances approval_instances_document_id_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_instances
    ADD CONSTRAINT approval_instances_document_id_idempotency_key_key UNIQUE (document_id, idempotency_key);


--
-- Name: approval_instances approval_instances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_instances
    ADD CONSTRAINT approval_instances_pkey PRIMARY KEY (id);


--
-- Name: approval_route_stages approval_route_stages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_route_stages
    ADD CONSTRAINT approval_route_stages_pkey PRIMARY KEY (id);


--
-- Name: approval_route_stages approval_route_stages_route_stage_order_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_route_stages
    ADD CONSTRAINT approval_route_stages_route_stage_order_key UNIQUE (route_id, stage_order);


--
-- Name: approval_routes approval_routes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_routes
    ADD CONSTRAINT approval_routes_pkey PRIMARY KEY (id);


--
-- Name: approval_routes approval_routes_tenant_profile_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_routes
    ADD CONSTRAINT approval_routes_tenant_profile_key UNIQUE (tenant_id, profile_code);


--
-- Name: approval_signoffs approval_signoffs_approval_instance_id_actor_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_approval_instance_id_actor_user_id_key UNIQUE (approval_instance_id, actor_user_id);


--
-- Name: approval_signoffs approval_signoffs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_pkey PRIMARY KEY (id);


--
-- Name: approval_signoffs approval_signoffs_stage_instance_id_actor_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_stage_instance_id_actor_user_id_key UNIQUE (stage_instance_id, actor_user_id);


--
-- Name: approval_stage_instances approval_stage_instances_approval_instance_id_stage_order_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_stage_instances
    ADD CONSTRAINT approval_stage_instances_approval_instance_id_stage_order_key UNIQUE (approval_instance_id, stage_order);


--
-- Name: approval_stage_instances approval_stage_instances_id_approval_instance_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_stage_instances
    ADD CONSTRAINT approval_stage_instances_id_approval_instance_id_key UNIQUE (id, approval_instance_id);


--
-- Name: approval_stage_instances approval_stage_instances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_stage_instances
    ADD CONSTRAINT approval_stage_instances_pkey PRIMARY KEY (id);


--
-- Name: autosave_pending_uploads autosave_pending_uploads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autosave_pending_uploads
    ADD CONSTRAINT autosave_pending_uploads_pkey PRIMARY KEY (id);


--
-- Name: autosave_pending_uploads autosave_pending_uploads_session_id_base_revision_id_conten_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autosave_pending_uploads
    ADD CONSTRAINT autosave_pending_uploads_session_id_base_revision_id_conten_key UNIQUE (session_id, base_revision_id, content_hash);


--
-- Name: cd_sequence_counters cd_sequence_counters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cd_sequence_counters
    ADD CONSTRAINT cd_sequence_counters_pkey PRIMARY KEY (tenant_id, profile_code, process_area_code);


--
-- Name: controlled_document_area_grants controlled_document_area_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_document_area_grants
    ADD CONSTRAINT controlled_document_area_grants_pkey PRIMARY KEY (tenant_id, controlled_document_id, area_code);


--
-- Name: controlled_document_user_grants controlled_document_user_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_document_user_grants
    ADD CONSTRAINT controlled_document_user_grants_pkey PRIMARY KEY (tenant_id, controlled_document_id, user_id);


--
-- Name: controlled_documents controlled_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_documents
    ADD CONSTRAINT controlled_documents_pkey PRIMARY KEY (id);


--
-- Name: controlled_documents controlled_documents_tenant_id_profile_code_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_documents
    ADD CONSTRAINT controlled_documents_tenant_id_profile_code_code_key UNIQUE (tenant_id, profile_code, code);


--
-- Name: document_checkpoints document_checkpoints_document_id_version_num_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_checkpoints
    ADD CONSTRAINT document_checkpoints_document_id_version_num_key UNIQUE (document_id, version_num);


--
-- Name: document_checkpoints document_checkpoints_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_checkpoints
    ADD CONSTRAINT document_checkpoints_pkey PRIMARY KEY (id);


--
-- Name: document_comments document_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comments
    ADD CONSTRAINT document_comments_pkey PRIMARY KEY (id);


--
-- Name: document_exports document_exports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_exports
    ADD CONSTRAINT document_exports_pkey PRIMARY KEY (id);


--
-- Name: document_placeholder_values document_placeholder_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_placeholder_values
    ADD CONSTRAINT document_placeholder_values_pkey PRIMARY KEY (tenant_id, revision_id, placeholder_id);


--
-- Name: document_revisions document_revisions_document_id_content_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_revisions
    ADD CONSTRAINT document_revisions_document_id_content_hash_key UNIQUE (document_id, content_hash);


--
-- Name: document_revisions document_revisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_revisions
    ADD CONSTRAINT document_revisions_pkey PRIMARY KEY (id);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: editor_sessions editor_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.editor_sessions
    ADD CONSTRAINT editor_sessions_pkey PRIMARY KEY (id);


--
-- Name: governance_events governance_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.governance_events
    ADD CONSTRAINT governance_events_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: template_audit_log template_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.template_audit_log
    ADD CONSTRAINT template_audit_log_pkey PRIMARY KEY (id);


--
-- Name: template_versions template_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.template_versions
    ADD CONSTRAINT template_versions_pkey PRIMARY KEY (id);


--
-- Name: template_versions template_versions_template_num_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.template_versions
    ADD CONSTRAINT template_versions_template_num_unique UNIQUE (template_id, version_num);


--
-- Name: templates templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates
    ADD CONSTRAINT templates_pkey PRIMARY KEY (id);


--
-- Name: templates templates_tenant_key_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates
    ADD CONSTRAINT templates_tenant_key_unique UNIQUE (tenant_id, key);


--
-- Name: templates_v2_approval_config templates_v2_approval_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_approval_config
    ADD CONSTRAINT templates_v2_approval_config_pkey PRIMARY KEY (template_id);


--
-- Name: templates_v2_audit_log templates_v2_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_audit_log
    ADD CONSTRAINT templates_v2_audit_log_pkey PRIMARY KEY (id);


--
-- Name: templates_v2_template templates_v2_template_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_template
    ADD CONSTRAINT templates_v2_template_pkey PRIMARY KEY (id);


--
-- Name: templates_v2_template templates_v2_template_tenant_id_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_template
    ADD CONSTRAINT templates_v2_template_tenant_id_key_key UNIQUE (tenant_id, key);


--
-- Name: templates_v2_template_version templates_v2_template_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_template_version
    ADD CONSTRAINT templates_v2_template_version_pkey PRIMARY KEY (id);


--
-- Name: templates_v2_template_version templates_v2_template_version_template_id_version_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_template_version
    ADD CONSTRAINT templates_v2_template_version_template_id_version_number_key UNIQUE (template_id, version_number);


--
-- Name: user_process_areas user_process_areas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_process_areas
    ADD CONSTRAINT user_process_areas_pkey PRIMARY KEY (user_id, area_code, effective_from);


--
-- Name: idx_audit_events_actor_time; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_audit_events_actor_time ON metaldocs.audit_events USING btree (actor_id, occurred_at DESC);


--
-- Name: idx_audit_events_audit_sequence; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX idx_audit_events_audit_sequence ON metaldocs.audit_events USING btree (audit_sequence);


--
-- Name: idx_audit_events_occurred_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_audit_events_occurred_at ON metaldocs.audit_events USING btree (occurred_at DESC);


--
-- Name: idx_audit_events_resource_time; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_audit_events_resource_time ON metaldocs.audit_events USING btree (resource_type, resource_id, occurred_at DESC);


--
-- Name: idx_audit_events_tenant_id; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_audit_events_tenant_id ON metaldocs.audit_events USING btree (tenant_id);


--
-- Name: idx_auth_sessions_active; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_auth_sessions_active ON metaldocs.auth_sessions USING btree (user_id, expires_at DESC) WHERE (revoked_at IS NULL);


--
-- Name: idx_auth_sessions_tenant_user; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_auth_sessions_tenant_user ON metaldocs.auth_sessions USING btree (tenant_id, user_id);


--
-- Name: idx_auth_sessions_user_id; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_auth_sessions_user_id ON metaldocs.auth_sessions USING btree (user_id);


--
-- Name: idx_doc_versions_search; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_doc_versions_search ON metaldocs.document_versions USING gin (search_vector);


--
-- Name: idx_document_access_policies_resource; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_access_policies_resource ON metaldocs.document_access_policies USING btree (resource_scope, resource_id);


--
-- Name: idx_document_access_policies_subject; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_access_policies_subject ON metaldocs.document_access_policies USING btree (subject_type, subject_id);


--
-- Name: idx_document_attachments_document_id_created_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_attachments_document_id_created_at ON metaldocs.document_attachments USING btree (document_id, created_at DESC);


--
-- Name: idx_document_collaboration_presence_last_seen; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_collaboration_presence_last_seen ON metaldocs.document_collaboration_presence USING btree (document_id, last_seen_at DESC);


--
-- Name: idx_document_departments_is_active; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_departments_is_active ON metaldocs.document_departments USING btree (is_active);


--
-- Name: idx_document_edit_locks_expires_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_edit_locks_expires_at ON metaldocs.document_edit_locks USING btree (expires_at);


--
-- Name: idx_document_images_sha256; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_images_sha256 ON metaldocs.document_images USING btree (sha256);


--
-- Name: idx_document_profile_schema_versions_active; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_profile_schema_versions_active ON metaldocs.document_profile_schema_versions USING btree (profile_code, is_active);


--
-- Name: idx_document_subjects_process_area_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_subjects_process_area_code ON metaldocs.document_subjects USING btree (process_area_code);


--
-- Name: idx_document_versions_created_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_document_versions_created_at ON metaldocs.document_versions USING btree (created_at DESC);


--
-- Name: idx_documents_business_unit; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_business_unit ON metaldocs.documents USING btree (business_unit);


--
-- Name: idx_documents_created_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_created_at ON metaldocs.documents USING btree (created_at DESC);


--
-- Name: idx_documents_department; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_department ON metaldocs.documents USING btree (department);


--
-- Name: idx_documents_document_code_unique; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX idx_documents_document_code_unique ON metaldocs.documents USING btree (document_code);


--
-- Name: idx_documents_document_family_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_document_family_code ON metaldocs.documents USING btree (document_family_code);


--
-- Name: idx_documents_document_profile_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_document_profile_code ON metaldocs.documents USING btree (document_profile_code);


--
-- Name: idx_documents_document_type_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_document_type_code ON metaldocs.documents USING btree (document_type_code);


--
-- Name: idx_documents_document_type_key; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_document_type_key ON metaldocs.documents USING btree (document_type_key);


--
-- Name: idx_documents_owner_id; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_owner_id ON metaldocs.documents USING btree (owner_id);


--
-- Name: idx_documents_process_area_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_process_area_code ON metaldocs.documents USING btree (process_area_code);


--
-- Name: idx_documents_profile_schema_version; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_profile_schema_version ON metaldocs.documents USING btree (document_profile_code, profile_schema_version);


--
-- Name: idx_documents_profile_sequence_unique; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX idx_documents_profile_sequence_unique ON metaldocs.documents USING btree (document_profile_code, document_sequence);


--
-- Name: idx_documents_review_reminder_window; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_review_reminder_window ON metaldocs.documents USING btree (status, expiry_at) WHERE ((expiry_at IS NOT NULL) AND (status = ANY (ARRAY['APPROVED'::text, 'PUBLISHED'::text])));


--
-- Name: idx_documents_subject_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_documents_subject_code ON metaldocs.documents USING btree (subject_code);


--
-- Name: idx_dvi_image; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_dvi_image ON metaldocs.document_version_images USING btree (image_id);


--
-- Name: idx_iam_user_roles_role_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_iam_user_roles_role_code ON metaldocs.iam_user_roles USING btree (role_code);


--
-- Name: idx_idempotency_keys_completed_expires; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_idempotency_keys_completed_expires ON metaldocs.idempotency_keys USING btree (expires_at) WHERE (status = 'completed'::text);


--
-- Name: idx_idempotency_keys_expires_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_idempotency_keys_expires_at ON metaldocs.idempotency_keys USING btree (expires_at);


--
-- Name: idx_notifications_recipient_created_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_notifications_recipient_created_at ON metaldocs.notifications USING btree (recipient_user_id, created_at DESC);


--
-- Name: idx_one_active_draft_per_doc; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX idx_one_active_draft_per_doc ON metaldocs.document_versions_mddm USING btree (document_id) WHERE (status = ANY (ARRAY['draft'::metaldocs.mddm_version_status, 'pending_approval'::metaldocs.mddm_version_status]));


--
-- Name: idx_one_released_per_doc; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX idx_one_released_per_doc ON metaldocs.document_versions_mddm USING btree (document_id) WHERE (status = 'released'::metaldocs.mddm_version_status);


--
-- Name: idx_outbox_claimable; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_outbox_claimable ON metaldocs.outbox_events USING btree (occurred_at) WHERE ((published_at IS NULL) AND (dead_lettered_at IS NULL));


--
-- Name: idx_outbox_dead_lettered; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_outbox_dead_lettered ON metaldocs.outbox_events USING btree (dead_lettered_at DESC) WHERE (dead_lettered_at IS NOT NULL);


--
-- Name: idx_template_audit_log_actor; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_template_audit_log_actor ON metaldocs.template_audit_log USING btree (actor_id);


--
-- Name: idx_template_audit_log_key; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_template_audit_log_key ON metaldocs.template_audit_log USING btree (template_key);


--
-- Name: idx_template_drafts_draft_status; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_template_drafts_draft_status ON metaldocs.template_drafts USING btree (draft_status);


--
-- Name: idx_template_drafts_profile; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_template_drafts_profile ON metaldocs.template_drafts USING btree (profile_code);


--
-- Name: idx_workflow_approvals_document_id_requested_at; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX idx_workflow_approvals_document_id_requested_at ON metaldocs.workflow_approvals USING btree (document_id, requested_at DESC);


--
-- Name: ix_pdf_dispatch_outbox_pending; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX ix_pdf_dispatch_outbox_pending ON metaldocs.pdf_dispatch_outbox USING btree (next_retry_at) WHERE (status = ANY (ARRAY['pending'::text, 'processing'::text]));


--
-- Name: mddm_shadow_diff_events_recorded_at_idx; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE INDEX mddm_shadow_diff_events_recorded_at_idx ON metaldocs.mddm_shadow_diff_events USING btree (recorded_at DESC);


--
-- Name: uq_auth_identities_email_ci; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX uq_auth_identities_email_ci ON metaldocs.auth_identities USING btree (lower(email)) WHERE (email IS NOT NULL);


--
-- Name: uq_auth_identities_username_ci; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX uq_auth_identities_username_ci ON metaldocs.auth_identities USING btree (lower(username));


--
-- Name: uq_document_types_type_key; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX uq_document_types_type_key ON metaldocs.document_types USING btree (type_key);


--
-- Name: ux_document_profiles_tenant_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX ux_document_profiles_tenant_code ON metaldocs.document_profiles USING btree (tenant_id, code);


--
-- Name: ux_iam_users_tenant_user; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX ux_iam_users_tenant_user ON metaldocs.iam_users USING btree (tenant_id, user_id);


--
-- Name: ux_iam_users_tenant_user_active; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX ux_iam_users_tenant_user_active ON metaldocs.iam_users USING btree (tenant_id, user_id) WHERE (deactivated_at IS NULL);


--
-- Name: ux_process_areas_tenant_code; Type: INDEX; Schema: metaldocs; Owner: -
--

CREATE UNIQUE INDEX ux_process_areas_tenant_code ON metaldocs.document_process_areas USING btree (tenant_id, code);


--
-- Name: document_exports_doc_hash_uidx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX document_exports_doc_hash_uidx ON public.document_exports USING btree (document_id, composite_hash);


--
-- Name: document_exports_document_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX document_exports_document_id_idx ON public.document_exports USING btree (document_id);


--
-- Name: idx_audit_tenant_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_tenant_created ON public.template_audit_log USING btree (tenant_id, created_at DESC);


--
-- Name: idx_document_comments_doc; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_document_comments_doc ON public.document_comments USING btree (document_id, created_at);


--
-- Name: idx_document_comments_doc_lib; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_document_comments_doc_lib ON public.document_comments USING btree (document_id, library_comment_id);


--
-- Name: idx_document_comments_parent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_document_comments_parent ON public.document_comments USING btree (document_id, parent_library_id) WHERE (parent_library_id IS NOT NULL);


--
-- Name: idx_documents_form_data_gin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_documents_form_data_gin ON public.documents USING gin (form_data_json jsonb_path_ops);


--
-- Name: idx_documents_template_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_documents_template_version ON public.documents USING btree (template_version_id);


--
-- Name: idx_documents_tenant_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_documents_tenant_status ON public.documents USING btree (tenant_id, status);


--
-- Name: idx_one_active_session_per_doc; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_one_active_session_per_doc ON public.editor_sessions USING btree (document_id) WHERE (status = 'active'::text);


--
-- Name: idx_one_draft_per_template; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_one_draft_per_template ON public.template_versions USING btree (template_id) WHERE (status = 'draft'::text);


--
-- Name: idx_pending_expired; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pending_expired ON public.autosave_pending_uploads USING btree (expires_at) WHERE (consumed_at IS NULL);


--
-- Name: idx_revisions_doc_num; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_revisions_doc_num ON public.document_revisions USING btree (document_id, revision_num DESC);


--
-- Name: idx_templates_tenant; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_templates_tenant ON public.templates USING btree (tenant_id);


--
-- Name: idx_templates_v2_audit_template_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_templates_v2_audit_template_time ON public.templates_v2_audit_log USING btree (template_id, occurred_at DESC);


--
-- Name: idx_templates_v2_template_tenant_doctype; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_templates_v2_template_tenant_doctype ON public.templates_v2_template USING btree (tenant_id, doc_type_code);


--
-- Name: idx_templates_v2_version_template_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_templates_v2_version_template_status ON public.templates_v2_template_version USING btree (template_id, status);


--
-- Name: ix_approval_instances_inbox; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_approval_instances_inbox ON public.approval_instances USING btree (tenant_id, submitted_by, status, submitted_at DESC);


--
-- Name: ix_approval_instances_tenant_document_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_approval_instances_tenant_document_id ON public.approval_instances USING btree (tenant_id, document_id);


--
-- Name: ix_cd_area_grants_tenant_area; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_cd_area_grants_tenant_area ON public.controlled_document_area_grants USING btree (tenant_id, area_code, controlled_document_id);


--
-- Name: ix_cd_user_grants_tenant_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_cd_user_grants_tenant_user ON public.controlled_document_user_grants USING btree (tenant_id, user_id, controlled_document_id);


--
-- Name: ix_controlled_documents_tenant_area; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_controlled_documents_tenant_area ON public.controlled_documents USING btree (tenant_id, process_area_code);


--
-- Name: ix_controlled_documents_tenant_profile; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_controlled_documents_tenant_profile ON public.controlled_documents USING btree (tenant_id, profile_code);


--
-- Name: ix_documents_controlled_doc; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_documents_controlled_doc ON public.documents USING btree (controlled_document_id) WHERE (controlled_document_id IS NOT NULL);


--
-- Name: ix_governance_events_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_governance_events_resource ON public.governance_events USING btree (resource_type, resource_id);


--
-- Name: ix_governance_events_tenant_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_governance_events_tenant_type ON public.governance_events USING btree (tenant_id, event_type, created_at DESC);


--
-- Name: ix_signoffs_stage; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_signoffs_stage ON public.approval_signoffs USING btree (stage_instance_id);


--
-- Name: ix_stage_instances_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_stage_instances_active ON public.approval_stage_instances USING btree (approval_instance_id, stage_order) WHERE (status = 'active'::text);


--
-- Name: ix_stage_instances_eligible_actors; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_stage_instances_eligible_actors ON public.approval_stage_instances USING gin (eligible_actor_ids);


--
-- Name: ix_user_process_areas_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_user_process_areas_active ON public.user_process_areas USING btree (user_id, area_code) WHERE (effective_to IS NULL);


--
-- Name: ux_approval_instances_active_document_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_approval_instances_active_document_id ON public.approval_instances USING btree (document_id) WHERE (status = 'in_progress'::text);


--
-- Name: ux_documents_v2_cd_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_documents_v2_cd_active ON public.documents USING btree (controlled_document_id) WHERE ((controlled_document_id IS NOT NULL) AND (status = ANY (ARRAY['draft'::text, 'under_review'::text, 'approved'::text, 'rejected'::text, 'scheduled'::text])));


--
-- Name: ux_documents_v2_cd_revision; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_documents_v2_cd_revision ON public.documents USING btree (controlled_document_id, revision_number) WHERE (controlled_document_id IS NOT NULL);


--
-- Name: ux_gov_events_dedupe; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_gov_events_dedupe ON public.governance_events USING btree (event_type, dedupe_key) WHERE (dedupe_key IS NOT NULL);


--
-- Name: ux_governance_events_caps_bump_spec_version; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_governance_events_caps_bump_spec_version ON public.governance_events USING btree (event_type, ((payload_json ->> 'spec'::text)), ((payload_json ->> 'to'::text))) WHERE (event_type = 'role_capabilities_version_bump'::text);


--
-- Name: ux_templates_v2_system_blank; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_templates_v2_system_blank ON public.templates_v2_template USING btree (tenant_id, key) WHERE ((system_owned = true) AND (key = '__system_blank__'::text));


--
-- Name: ux_user_process_areas_one_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_user_process_areas_one_active ON public.user_process_areas USING btree (user_id, tenant_id, area_code) WHERE (effective_to IS NULL);


--
-- Name: ux_user_process_areas_single_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_user_process_areas_single_active ON public.user_process_areas USING btree (tenant_id, user_id, area_code, role) WHERE (effective_to IS NULL);


--
-- Name: document_profiles trg_document_profiles_code_immutable; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_document_profiles_code_immutable BEFORE UPDATE ON metaldocs.document_profiles FOR EACH ROW EXECUTE FUNCTION public.reject_code_update();


--
-- Name: document_process_areas trg_process_areas_code_immutable; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_process_areas_code_immutable BEFORE UPDATE ON metaldocs.document_process_areas FOR EACH ROW EXECUTE FUNCTION public.reject_code_update();


--
-- Name: document_families trg_reject_families_code_update; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_reject_families_code_update BEFORE UPDATE ON metaldocs.document_families FOR EACH ROW EXECUTE FUNCTION public.reject_families_code_update();


--
-- Name: document_families trg_require_cap_asserted; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.document_families FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: document_process_areas trg_require_cap_asserted; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.document_process_areas FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: document_profiles trg_require_cap_asserted; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.document_profiles FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: iam_user_roles trg_require_cap_asserted; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON metaldocs.iam_user_roles FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: document_template_versions_mddm trg_template_immutable; Type: TRIGGER; Schema: metaldocs; Owner: -
--

CREATE TRIGGER trg_template_immutable BEFORE UPDATE ON metaldocs.document_template_versions_mddm FOR EACH ROW EXECUTE FUNCTION metaldocs.prevent_published_template_mutation();


--
-- Name: document_placeholder_values enforce_placeholder_value_tenant_trg; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER enforce_placeholder_value_tenant_trg BEFORE INSERT OR UPDATE ON public.document_placeholder_values FOR EACH ROW EXECUTE FUNCTION public.enforce_placeholder_value_tenant_consistent();


--
-- Name: approval_signoffs enforce_signoff_eligibility_trg; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER enforce_signoff_eligibility_trg BEFORE INSERT ON public.approval_signoffs FOR EACH ROW EXECUTE FUNCTION public.enforce_signoff_eligibility();


--
-- Name: documents enforce_snapshot_on_submit_trg; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER enforce_snapshot_on_submit_trg BEFORE INSERT OR UPDATE ON public.documents FOR EACH ROW EXECUTE FUNCTION public.enforce_snapshot_on_submit();


--
-- Name: controlled_documents trg_controlled_documents_code_immutable; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_controlled_documents_code_immutable BEFORE UPDATE ON public.controlled_documents FOR EACH ROW EXECUTE FUNCTION public.reject_code_update();


--
-- Name: documents trg_documents_v2_legal_transition; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_documents_v2_legal_transition BEFORE UPDATE ON public.documents FOR EACH ROW EXECUTE FUNCTION public.enforce_document_transition();


--
-- Name: documents trg_documents_v2_revision_version_monotonic; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_documents_v2_revision_version_monotonic BEFORE UPDATE ON public.documents FOR EACH ROW EXECUTE FUNCTION public.enforce_revision_version_monotonic();


--
-- Name: cd_sequence_counters trg_require_cap_asserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR UPDATE ON public.cd_sequence_counters FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: controlled_documents trg_require_cap_asserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR UPDATE ON public.controlled_documents FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: documents trg_require_cap_asserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR UPDATE ON public.documents FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: templates_v2_template trg_require_cap_asserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON public.templates_v2_template FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: templates_v2_template_version trg_require_cap_asserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON public.templates_v2_template_version FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: user_process_areas trg_require_cap_asserted; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE ON public.user_process_areas FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: approval_instances trg_require_cap_asserted_instances; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted_instances BEFORE INSERT ON public.approval_instances FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: approval_signoffs trg_require_cap_asserted_signoffs; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_require_cap_asserted_signoffs BEFORE INSERT ON public.approval_signoffs FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();


--
-- Name: approval_routes trg_route_config_immutable_del; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_route_config_immutable_del BEFORE DELETE ON public.approval_routes FOR EACH ROW EXECUTE FUNCTION public.enforce_route_immutable();


--
-- Name: approval_routes trg_route_config_immutable_upd; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_route_config_immutable_upd BEFORE UPDATE ON public.approval_routes FOR EACH ROW EXECUTE FUNCTION public.enforce_route_immutable();


--
-- Name: approval_signoffs trg_signoff_immutable; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_signoff_immutable BEFORE UPDATE ON public.approval_signoffs FOR EACH ROW EXECUTE FUNCTION public.reject_signoff_update();


--
-- Name: approval_signoffs trg_signoff_sod; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_signoff_sod BEFORE INSERT ON public.approval_signoffs FOR EACH ROW EXECUTE FUNCTION public.enforce_signoff_sod();


--
-- Name: approval_signoffs trg_signoff_tenant_consistent; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_signoff_tenant_consistent BEFORE INSERT ON public.approval_signoffs FOR EACH ROW EXECUTE FUNCTION public.enforce_signoff_tenant_consistent();


--
-- Name: user_process_areas trg_user_process_areas_no_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_user_process_areas_no_delete BEFORE DELETE ON public.user_process_areas FOR EACH ROW EXECUTE FUNCTION public.reject_user_process_areas_delete();


--
-- Name: user_process_areas trg_user_process_areas_update_contract; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_user_process_areas_update_contract BEFORE UPDATE ON public.user_process_areas FOR EACH ROW EXECUTE FUNCTION public.enforce_user_process_areas_update_contract();


--
-- Name: document_attachments document_attachments_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_attachments
    ADD CONSTRAINT document_attachments_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE RESTRICT;


--
-- Name: document_collaboration_presence document_collaboration_presence_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_collaboration_presence
    ADD CONSTRAINT document_collaboration_presence_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE CASCADE;


--
-- Name: document_edit_locks document_edit_locks_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_edit_locks
    ADD CONSTRAINT document_edit_locks_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE CASCADE;


--
-- Name: document_profile_governance document_profile_governance_profile_code_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profile_governance
    ADD CONSTRAINT document_profile_governance_profile_code_fkey FOREIGN KEY (profile_code) REFERENCES metaldocs.document_profiles(code);


--
-- Name: document_profile_schema_versions document_profile_schema_versions_profile_code_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profile_schema_versions
    ADD CONSTRAINT document_profile_schema_versions_profile_code_fkey FOREIGN KEY (profile_code) REFERENCES metaldocs.document_profiles(code);


--
-- Name: document_profiles document_profiles_default_template_version_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profiles
    ADD CONSTRAINT document_profiles_default_template_version_id_fkey FOREIGN KEY (default_template_version_id) REFERENCES public.templates_v2_template_version(id);


--
-- Name: document_profiles document_profiles_family_code_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_profiles
    ADD CONSTRAINT document_profiles_family_code_fkey FOREIGN KEY (family_code) REFERENCES metaldocs.document_families(code);


--
-- Name: document_sequences document_sequences_profile_code_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_sequences
    ADD CONSTRAINT document_sequences_profile_code_fkey FOREIGN KEY (profile_code) REFERENCES metaldocs.document_profiles(code);


--
-- Name: document_subjects document_subjects_process_area_code_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_subjects
    ADD CONSTRAINT document_subjects_process_area_code_fkey FOREIGN KEY (process_area_code) REFERENCES metaldocs.document_process_areas(code);


--
-- Name: document_template_assignments document_template_assignments_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_template_assignments
    ADD CONSTRAINT document_template_assignments_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE CASCADE;


--
-- Name: document_type_schema_versions document_type_schema_versions_type_key_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_type_schema_versions
    ADD CONSTRAINT document_type_schema_versions_type_key_fkey FOREIGN KEY (type_key) REFERENCES metaldocs.document_types(type_key);


--
-- Name: document_version_images document_version_images_document_version_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_version_images
    ADD CONSTRAINT document_version_images_document_version_id_fkey FOREIGN KEY (document_version_id) REFERENCES metaldocs.document_versions_mddm(id) ON DELETE CASCADE;


--
-- Name: document_version_images document_version_images_image_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_version_images
    ADD CONSTRAINT document_version_images_image_id_fkey FOREIGN KEY (image_id) REFERENCES metaldocs.document_images(id);


--
-- Name: document_versions document_versions_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_versions
    ADD CONSTRAINT document_versions_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE RESTRICT;


--
-- Name: document_versions_mddm document_versions_mddm_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_versions_mddm
    ADD CONSTRAINT document_versions_mddm_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE CASCADE;


--
-- Name: document_process_areas fk_area_parent_tenant; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.document_process_areas
    ADD CONSTRAINT fk_area_parent_tenant FOREIGN KEY (tenant_id, parent_code) REFERENCES metaldocs.document_process_areas(tenant_id, code);


--
-- Name: auth_sessions fk_auth_sessions_identity; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.auth_sessions
    ADD CONSTRAINT fk_auth_sessions_identity FOREIGN KEY (user_id) REFERENCES metaldocs.auth_identities(user_id) ON DELETE CASCADE;


--
-- Name: documents fk_documents_document_family_code; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.documents
    ADD CONSTRAINT fk_documents_document_family_code FOREIGN KEY (document_family_code) REFERENCES metaldocs.document_families(code);


--
-- Name: documents fk_documents_document_profile_code; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.documents
    ADD CONSTRAINT fk_documents_document_profile_code FOREIGN KEY (document_profile_code) REFERENCES metaldocs.document_profiles(code);


--
-- Name: documents fk_documents_document_type_code; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.documents
    ADD CONSTRAINT fk_documents_document_type_code FOREIGN KEY (document_type_code) REFERENCES metaldocs.document_types(code);


--
-- Name: documents fk_documents_process_area_code; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.documents
    ADD CONSTRAINT fk_documents_process_area_code FOREIGN KEY (process_area_code) REFERENCES metaldocs.document_process_areas(code);


--
-- Name: documents fk_documents_subject_code; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.documents
    ADD CONSTRAINT fk_documents_subject_code FOREIGN KEY (subject_code) REFERENCES metaldocs.document_subjects(code);


--
-- Name: iam_group_members iam_group_members_group_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_group_members
    ADD CONSTRAINT iam_group_members_group_id_fkey FOREIGN KEY (group_id) REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE;


--
-- Name: iam_group_roles iam_group_roles_group_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_group_roles
    ADD CONSTRAINT iam_group_roles_group_id_fkey FOREIGN KEY (group_id) REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE;


--
-- Name: iam_user_roles iam_user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.iam_user_roles
    ADD CONSTRAINT iam_user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES metaldocs.iam_users(user_id) ON DELETE CASCADE;


--
-- Name: template_drafts template_drafts_profile_code_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.template_drafts
    ADD CONSTRAINT template_drafts_profile_code_fkey FOREIGN KEY (profile_code) REFERENCES metaldocs.document_profiles(code);


--
-- Name: workflow_approvals workflow_approvals_document_id_fkey; Type: FK CONSTRAINT; Schema: metaldocs; Owner: -
--

ALTER TABLE ONLY metaldocs.workflow_approvals
    ADD CONSTRAINT workflow_approvals_document_id_fkey FOREIGN KEY (document_id) REFERENCES metaldocs.documents(id) ON DELETE RESTRICT;


--
-- Name: approval_instances approval_instances_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_instances
    ADD CONSTRAINT approval_instances_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id);


--
-- Name: approval_instances approval_instances_route_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_instances
    ADD CONSTRAINT approval_instances_route_id_fkey FOREIGN KEY (route_id) REFERENCES public.approval_routes(id);


--
-- Name: approval_instances approval_instances_submitted_by_tenant_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_instances
    ADD CONSTRAINT approval_instances_submitted_by_tenant_fkey FOREIGN KEY (tenant_id, submitted_by) REFERENCES metaldocs.iam_users(tenant_id, user_id);


--
-- Name: approval_route_stages approval_route_stages_route_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_route_stages
    ADD CONSTRAINT approval_route_stages_route_id_fkey FOREIGN KEY (route_id) REFERENCES public.approval_routes(id) ON DELETE CASCADE;


--
-- Name: approval_routes approval_routes_document_profile_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_routes
    ADD CONSTRAINT approval_routes_document_profile_fk FOREIGN KEY (tenant_id, profile_code) REFERENCES metaldocs.document_profiles(tenant_id, code);


--
-- Name: approval_signoffs approval_signoffs_actor_tenant_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_actor_tenant_fkey FOREIGN KEY (actor_tenant_id, actor_user_id) REFERENCES metaldocs.iam_users(tenant_id, user_id);


--
-- Name: approval_signoffs approval_signoffs_approval_instance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_approval_instance_id_fkey FOREIGN KEY (approval_instance_id) REFERENCES public.approval_instances(id);


--
-- Name: approval_signoffs approval_signoffs_stage_instance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_stage_instance_id_fkey FOREIGN KEY (stage_instance_id) REFERENCES public.approval_stage_instances(id);


--
-- Name: approval_signoffs approval_signoffs_stage_matches_instance; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_stage_matches_instance FOREIGN KEY (stage_instance_id, approval_instance_id) REFERENCES public.approval_stage_instances(id, approval_instance_id);


--
-- Name: approval_stage_instances approval_stage_instances_approval_instance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_stage_instances
    ADD CONSTRAINT approval_stage_instances_approval_instance_id_fkey FOREIGN KEY (approval_instance_id) REFERENCES public.approval_instances(id) ON DELETE CASCADE;


--
-- Name: autosave_pending_uploads autosave_pending_uploads_base_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autosave_pending_uploads
    ADD CONSTRAINT autosave_pending_uploads_base_revision_id_fkey FOREIGN KEY (base_revision_id) REFERENCES public.document_revisions(id);


--
-- Name: autosave_pending_uploads autosave_pending_uploads_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autosave_pending_uploads
    ADD CONSTRAINT autosave_pending_uploads_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: autosave_pending_uploads autosave_pending_uploads_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autosave_pending_uploads
    ADD CONSTRAINT autosave_pending_uploads_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.editor_sessions(id) ON DELETE CASCADE;


--
-- Name: cd_sequence_counters cd_sequence_counters_tenant_id_process_area_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cd_sequence_counters
    ADD CONSTRAINT cd_sequence_counters_tenant_id_process_area_code_fkey FOREIGN KEY (tenant_id, process_area_code) REFERENCES metaldocs.document_process_areas(tenant_id, code);


--
-- Name: cd_sequence_counters cd_sequence_counters_tenant_id_profile_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cd_sequence_counters
    ADD CONSTRAINT cd_sequence_counters_tenant_id_profile_code_fkey FOREIGN KEY (tenant_id, profile_code) REFERENCES metaldocs.document_profiles(tenant_id, code);


--
-- Name: controlled_document_area_grants controlled_document_area_grants_controlled_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_document_area_grants
    ADD CONSTRAINT controlled_document_area_grants_controlled_document_id_fkey FOREIGN KEY (controlled_document_id) REFERENCES public.controlled_documents(id) ON DELETE CASCADE;


--
-- Name: controlled_document_area_grants controlled_document_area_grants_tenant_id_area_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_document_area_grants
    ADD CONSTRAINT controlled_document_area_grants_tenant_id_area_code_fkey FOREIGN KEY (tenant_id, area_code) REFERENCES metaldocs.document_process_areas(tenant_id, code);


--
-- Name: controlled_document_user_grants controlled_document_user_grants_controlled_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_document_user_grants
    ADD CONSTRAINT controlled_document_user_grants_controlled_document_id_fkey FOREIGN KEY (controlled_document_id) REFERENCES public.controlled_documents(id) ON DELETE CASCADE;


--
-- Name: controlled_documents controlled_documents_override_template_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_documents
    ADD CONSTRAINT controlled_documents_override_template_version_id_fkey FOREIGN KEY (override_template_version_id) REFERENCES public.templates_v2_template_version(id);


--
-- Name: controlled_documents controlled_documents_tenant_id_process_area_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_documents
    ADD CONSTRAINT controlled_documents_tenant_id_process_area_code_fkey FOREIGN KEY (tenant_id, process_area_code) REFERENCES metaldocs.document_process_areas(tenant_id, code);


--
-- Name: controlled_documents controlled_documents_tenant_id_profile_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controlled_documents
    ADD CONSTRAINT controlled_documents_tenant_id_profile_code_fkey FOREIGN KEY (tenant_id, profile_code) REFERENCES metaldocs.document_profiles(tenant_id, code);


--
-- Name: document_checkpoints document_checkpoints_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_checkpoints
    ADD CONSTRAINT document_checkpoints_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: document_checkpoints document_checkpoints_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_checkpoints
    ADD CONSTRAINT document_checkpoints_revision_id_fkey FOREIGN KEY (revision_id) REFERENCES public.document_revisions(id);


--
-- Name: document_comments document_comments_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comments
    ADD CONSTRAINT document_comments_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: document_exports document_exports_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_exports
    ADD CONSTRAINT document_exports_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: document_placeholder_values document_placeholder_values_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_placeholder_values
    ADD CONSTRAINT document_placeholder_values_revision_id_fkey FOREIGN KEY (revision_id) REFERENCES public.document_revisions(id) ON DELETE CASCADE;


--
-- Name: document_revisions document_revisions_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_revisions
    ADD CONSTRAINT document_revisions_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: document_revisions document_revisions_parent_revision_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_revisions
    ADD CONSTRAINT document_revisions_parent_revision_id_fkey FOREIGN KEY (parent_revision_id) REFERENCES public.document_revisions(id);


--
-- Name: document_revisions document_revisions_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_revisions
    ADD CONSTRAINT document_revisions_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.editor_sessions(id);


--
-- Name: documents documents_controlled_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_controlled_document_id_fkey FOREIGN KEY (controlled_document_id) REFERENCES public.controlled_documents(id);


--
-- Name: editor_sessions editor_sessions_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.editor_sessions
    ADD CONSTRAINT editor_sessions_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: documents fk_active_session; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT fk_active_session FOREIGN KEY (active_session_id) REFERENCES public.editor_sessions(id) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: documents fk_current_revision; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT fk_current_revision FOREIGN KEY (current_revision_id) REFERENCES public.document_revisions(id) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: templates fk_templates_current_published; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates
    ADD CONSTRAINT fk_templates_current_published FOREIGN KEY (current_published_version_id) REFERENCES public.template_versions(id) DEFERRABLE;


--
-- Name: templates_v2_template fk_templates_v2_published_version; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_template
    ADD CONSTRAINT fk_templates_v2_published_version FOREIGN KEY (published_version_id) REFERENCES public.templates_v2_template_version(id);


--
-- Name: template_versions template_versions_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.template_versions
    ADD CONSTRAINT template_versions_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.templates(id) ON DELETE CASCADE;


--
-- Name: templates_v2_approval_config templates_v2_approval_config_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_approval_config
    ADD CONSTRAINT templates_v2_approval_config_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.templates_v2_template(id);


--
-- Name: templates_v2_template_version templates_v2_template_version_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.templates_v2_template_version
    ADD CONSTRAINT templates_v2_template_version_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.templates_v2_template(id);


--
-- Name: user_process_areas user_process_areas_granted_by_same_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_process_areas
    ADD CONSTRAINT user_process_areas_granted_by_same_tenant FOREIGN KEY (tenant_id, granted_by) REFERENCES metaldocs.iam_users(tenant_id, user_id);


--
-- Name: user_process_areas user_process_areas_revoked_by_same_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_process_areas
    ADD CONSTRAINT user_process_areas_revoked_by_same_tenant FOREIGN KEY (tenant_id, revoked_by) REFERENCES metaldocs.iam_users(tenant_id, user_id);


--
-- Name: user_process_areas user_process_areas_tenant_id_area_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_process_areas
    ADD CONSTRAINT user_process_areas_tenant_id_area_code_fkey FOREIGN KEY (tenant_id, area_code) REFERENCES metaldocs.document_process_areas(tenant_id, code);


--
-- PostgreSQL database dump complete
--




COMMIT;

