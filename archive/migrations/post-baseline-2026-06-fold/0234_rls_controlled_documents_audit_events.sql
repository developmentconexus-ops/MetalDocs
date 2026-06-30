-- 0234_rls_controlled_documents_audit_events.sql
-- RLS defense-in-depth on public.controlled_documents and metaldocs.audit_events.
-- Implements ADR 0027 Tier 1 (Wave 2.3), F-12, D-3, REQ-TEN-1, OWASP ASVS V4.1.3.
--
-- ── PRE-DESIGN AUDIT FINDINGS ────────────────────────────────────────────────
--
-- DB ROLE / OWNER ANALYSIS
--   The application connects as metaldocs_app (PGUSER in .env — see .env for
--   credentials). In the dev Docker setup, metaldocs_app is a superuser with
--   BYPASSRLS=true. PostgreSQL RLS does not apply to superusers regardless of
--   ENABLE or FORCE ROW LEVEL SECURITY. In production the NOSUPERUSER deployment
--   constraint (recorded in ADR 0022 Phase 5 §Item 7 and referenced in ADR 0027)
--   applies: the app role must be NOSUPERUSER and NOBYPASSRLS. When that constraint
--   is satisfied, ENABLE + FORCE makes the policies active for all statements
--   including those issued by the table owner. This migration creates the schema
--   objects now; the deployment constraint is a runtime/infra concern, not a
--   migration concern.
--
--   controlled_documents: schema=public,    owner=metaldocs_app, tenant_id type=uuid
--   audit_events:         schema=metaldocs, owner=metaldocs_app, tenant_id type=text
--   Policy for controlled_documents casts NULLIF(current_setting,...) to uuid.
--   Policy for audit_events uses plain text equality (no cast needed).
--
-- GUC-LESS PATH INVENTORY
--   metaldocs.tenant_id is set by SeedTxIdentity (context.go) on every
--   authenticated user transaction before writes. However, many legitimate
--   system paths touch both tables WITHOUT the GUC set:
--
--   controlled_documents (READ paths — no GUC):
--     internal/modules/controlleddocuments/infrastructure/repository.go
--       GetByID, GetByCode, CodeExists, List, ListWithDocuments — db.QueryRowContext
--     internal/modules/documents/delivery/http/handler.go:530
--       profile_code lookup — db.QueryRowContext in delivery layer
--     internal/modules/search/infrastructure/v2documents/reader.go
--       LEFT JOIN controlled_documents — db.QueryContext
--     internal/modules/documents/application/document_area.go
--       LEFT JOIN read — db.QueryContext
--     internal/modules/documents/approval/application/read_service.go
--       LEFT JOIN read — db.QueryContext
--     Integration tests — direct INSERT without tx GUC
--
--   audit_events (READ + WRITE paths — no GUC):
--     internal/modules/audit/infrastructure/postgres/writer.go Record()
--       Opens own tx via db.BeginTx — tenant_id set in row data but NOT as GUC
--     internal/modules/audit/infrastructure/postgres/writer.go ValidateIntegrity()
--       db.QueryContext — no GUC
--     internal/modules/iam/infrastructure/postgres/observability_repository.go
--       CountAuditEventsByActionPrefix, CountActiveUsersInWindow,
--       CountFailedLoginsInWindow — db.QueryRowContext, no GUC
--     internal/modules/audit/infrastructure/postgres/writer.go ListEvents()
--       db.QueryContext — no GUC (TenantID in WHERE clause but not in GUC)
--
-- CHOSEN POLICY: NULL-PERMISSIVE (tripwire-grade)
--   USING (
--     NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
--     OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
--   )
--
--   NULLIF(current_setting(..., true), '') collapses both the NULL (GUC unset)
--   and '' (empty string returned when GUC was never set) cases to a single NULL,
--   enabling a clean IS NULL short-circuit without attempting a text→uuid cast
--   on an empty string (which would error at runtime).
--
--   Rationale:
--   - When GUC is set AND matches: rows visible — authenticated user path works
--   - When GUC is set AND mismatches: rows NOT visible — cross-tenant bug caught
--   - When GUC is unset (NULL/''):  rows visible — all GUC-less system paths work
--   This is defense-in-depth against predicate omission in authenticated paths;
--   it does not block legitimate system paths that set tenant_id in row data
--   but not in the GUC.
--
--   The metaldocs.bypass_authz GUC (used by the scheduler tripwire) is not
--   referenced here — that GUC gates the enforce_capability_asserted trigger
--   on write paths, which is a separate mechanism. RLS policies apply regardless.
--   System writes that set bypass_authz but not tenant_id remain GUC-less for
--   RLS purposes and are therefore permitted by the NULL branch above.
--
--   FORCE ROW LEVEL SECURITY: required because table owner = app role. Without
--   FORCE, RLS is not applied to the owner's statements even when BYPASSRLS is
--   false (Postgres default: owner bypass RLS). FORCE ensures the policy fires
--   for NOSUPERUSER+NOBYPASSRLS production roles that still own the table.
--
-- Forward-only, idempotent.

BEGIN;

-- ── #1 public.controlled_documents ───────────────────────────────────────────

ALTER TABLE public.controlled_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.controlled_documents FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON public.controlled_documents;
CREATE POLICY tenant_isolation ON public.controlled_documents
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

-- ── #2 metaldocs.audit_events ────────────────────────────────────────────────

ALTER TABLE metaldocs.audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.audit_events FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON metaldocs.audit_events;
CREATE POLICY tenant_isolation ON metaldocs.audit_events
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')
  );
-- Note: metaldocs.audit_events.tenant_id is TEXT (not uuid), so no cast is needed.

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0234', 'RLS defense-in-depth: ENABLE+FORCE+tenant_isolation policy on public.controlled_documents and metaldocs.audit_events (ADR 0027 Tier 1, F-12, D-3, REQ-TEN-1)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
