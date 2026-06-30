-- 0237_rls_all_tenant_tables.sql
-- RLS defense-in-depth extended to ALL remaining tenant-scoped tables.
-- Wave Z Z-2 (F-12 tail, RF-6, REQ-TEN-1) + Z-3 (F-09d idempotency tenant FK).
-- ADR 0027 sequencing EXECUTED IN FULL by this migration (operator override of
-- D-3 scope limits per the Wave Z spec, 2026-06-12).
--
-- Pattern is 0234's, verbatim: ENABLE + FORCE ROW LEVEL SECURITY plus ONE
-- NULL-permissive tenant_isolation policy per table, keyed on the
-- metaldocs.tenant_id GUC set by SeedTxIdentity (iam/authz/context.go).
--
--   - GUC set + matches    -> rows visible (authenticated user path)
--   - GUC set + mismatches -> rows hidden (cross-tenant bug caught: tripwire)
--   - GUC unset/empty      -> rows visible (legitimate GUC-less system paths:
--     workers, jobs, audit writer, observability counters, integration tests)
--
-- Census (re-run 2026-06-12 in-session): 30 tenant_id columns in public+
-- metaldocs minus the 2 covered by 0234 (public.controlled_documents,
-- metaldocs.audit_events) and minus metaldocs.user_process_areas, which is a
-- VIEW over public.user_process_areas (information_schema.columns lists view
-- columns; RLS cannot be enabled on views — SQLSTATE 42809; the base table
-- below is covered) = 27 tables. All 27 tenant_id columns are uuid, so every
-- policy uses the ::uuid cast form from 0234.
--
-- SUPERUSER CAVEAT (0234, verbatim): the dev Docker role metaldocs_app is a
-- superuser with BYPASSRLS — RLS does not apply to superusers regardless of
-- ENABLE/FORCE. In production the NOSUPERUSER deployment constraint (ADR 0022
-- Phase 5 §Item 7, ADR 0027) applies: the app role must be NOSUPERUSER and
-- NOBYPASSRLS, at which point ENABLE + FORCE makes these policies active for
-- all statements including the table owner's. This migration creates the
-- schema objects now; the deployment constraint is a runtime/infra concern.
--
-- FORCE ROW LEVEL SECURITY: required because table owner = app role. Without
-- FORCE, RLS is not applied to the owner's statements even when BYPASSRLS is
-- false.
--
-- Forward-only, idempotent.

BEGIN;
-- metaldocs.audit_export_jobs
ALTER TABLE metaldocs.audit_export_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.audit_export_jobs FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.audit_export_jobs;
CREATE POLICY tenant_isolation ON metaldocs.audit_export_jobs
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.auth_sessions
ALTER TABLE metaldocs.auth_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.auth_sessions FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.auth_sessions;
CREATE POLICY tenant_isolation ON metaldocs.auth_sessions
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.document_process_areas
ALTER TABLE metaldocs.document_process_areas ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.document_process_areas FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.document_process_areas;
CREATE POLICY tenant_isolation ON metaldocs.document_process_areas
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.document_profiles
ALTER TABLE metaldocs.document_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.document_profiles FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.document_profiles;
CREATE POLICY tenant_isolation ON metaldocs.document_profiles
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.iam_group_members
ALTER TABLE metaldocs.iam_group_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.iam_group_members FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.iam_group_members;
CREATE POLICY tenant_isolation ON metaldocs.iam_group_members
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.iam_groups
ALTER TABLE metaldocs.iam_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.iam_groups FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.iam_groups;
CREATE POLICY tenant_isolation ON metaldocs.iam_groups
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.iam_user_roles
ALTER TABLE metaldocs.iam_user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.iam_user_roles FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.iam_user_roles;
CREATE POLICY tenant_isolation ON metaldocs.iam_user_roles
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.iam_users
ALTER TABLE metaldocs.iam_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.iam_users FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.iam_users;
CREATE POLICY tenant_isolation ON metaldocs.iam_users
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.idempotency_keys
ALTER TABLE metaldocs.idempotency_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.idempotency_keys FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.idempotency_keys;
CREATE POLICY tenant_isolation ON metaldocs.idempotency_keys
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.materialize_dispatch_outbox
ALTER TABLE metaldocs.materialize_dispatch_outbox ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.materialize_dispatch_outbox FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.materialize_dispatch_outbox;
CREATE POLICY tenant_isolation ON metaldocs.materialize_dispatch_outbox
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.pdf_dispatch_outbox
ALTER TABLE metaldocs.pdf_dispatch_outbox ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.pdf_dispatch_outbox FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.pdf_dispatch_outbox;
CREATE POLICY tenant_isolation ON metaldocs.pdf_dispatch_outbox
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- metaldocs.tenant_plans
ALTER TABLE metaldocs.tenant_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.tenant_plans FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.tenant_plans;
CREATE POLICY tenant_isolation ON metaldocs.tenant_plans
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.approval_instances
ALTER TABLE public.approval_instances ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.approval_instances FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.approval_instances;
CREATE POLICY tenant_isolation ON public.approval_instances
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.approval_routes
ALTER TABLE public.approval_routes ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.approval_routes FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.approval_routes;
CREATE POLICY tenant_isolation ON public.approval_routes
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.cd_sequence_counters
ALTER TABLE public.cd_sequence_counters ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.cd_sequence_counters FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.cd_sequence_counters;
CREATE POLICY tenant_isolation ON public.cd_sequence_counters
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.controlled_document_area_grants
ALTER TABLE public.controlled_document_area_grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.controlled_document_area_grants FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.controlled_document_area_grants;
CREATE POLICY tenant_isolation ON public.controlled_document_area_grants
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.controlled_document_user_grants
ALTER TABLE public.controlled_document_user_grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.controlled_document_user_grants FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.controlled_document_user_grants;
CREATE POLICY tenant_isolation ON public.controlled_document_user_grants
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.document_comments
ALTER TABLE public.document_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.document_comments FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.document_comments;
CREATE POLICY tenant_isolation ON public.document_comments
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.document_placeholder_values
ALTER TABLE public.document_placeholder_values ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.document_placeholder_values FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.document_placeholder_values;
CREATE POLICY tenant_isolation ON public.document_placeholder_values
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.documents
ALTER TABLE public.documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.documents FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.documents;
CREATE POLICY tenant_isolation ON public.documents
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.editor_sessions
ALTER TABLE public.editor_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.editor_sessions FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.editor_sessions;
CREATE POLICY tenant_isolation ON public.editor_sessions
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.governance_events
ALTER TABLE public.governance_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.governance_events FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.governance_events;
CREATE POLICY tenant_isolation ON public.governance_events
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.template_audit_log
ALTER TABLE public.template_audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.template_audit_log FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.template_audit_log;
CREATE POLICY tenant_isolation ON public.template_audit_log
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.templates
ALTER TABLE public.templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.templates FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.templates;
CREATE POLICY tenant_isolation ON public.templates
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.templates_audit_log
ALTER TABLE public.templates_audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.templates_audit_log FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.templates_audit_log;
CREATE POLICY tenant_isolation ON public.templates_audit_log
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.templates_template
ALTER TABLE public.templates_template ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.templates_template FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.templates_template;
CREATE POLICY tenant_isolation ON public.templates_template
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- public.user_process_areas
ALTER TABLE public.user_process_areas ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_process_areas FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON public.user_process_areas;
CREATE POLICY tenant_isolation ON public.user_process_areas
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- ── Z-3: idempotency_keys tenant FK (F-09d) ─────────────────────────────────
-- Column exists (uuid NOT NULL); 0 orphan rows verified in-session against
-- metaldocs.tenants. ON DELETE CASCADE: a tenant hard-delete purges its
-- idempotency replay records with it — they are meaningless without the tenant.

ALTER TABLE metaldocs.idempotency_keys
  DROP CONSTRAINT IF EXISTS fk_idempotency_keys_tenant;
ALTER TABLE metaldocs.idempotency_keys
  ADD CONSTRAINT fk_idempotency_keys_tenant
  FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id) ON DELETE CASCADE;

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0237', 'RLS defense-in-depth on all 27 remaining tenant-scoped tables + idempotency_keys tenant FK (Wave Z Z-2/Z-3, F-12 tail, RF-6, REQ-TEN-1, F-09d, ADR 0027 executed in full)')
ON CONFLICT (version) DO NOTHING;

COMMIT;