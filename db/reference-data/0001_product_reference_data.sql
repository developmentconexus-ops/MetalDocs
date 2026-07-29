-- MetalDocs product reference data.
-- Data in this file is required for every environment.

BEGIN;

INSERT INTO public.schema_migrations (version, description)
VALUES ('baseline-2026-05-14', 'curated current-state database baseline')
ON CONFLICT (version) DO NOTHING;

-- Folded migration ledger (baseline squash 2026-06): migrations 0203..0256
-- were folded into db/baseline/0001_current_schema.sql and archived under
-- archive/migrations/post-baseline-2026-06-fold/. Seed their ledger rows so the
-- runner's ledger-skip treats them as already applied (the archived files are
-- never re-run on a fresh bootstrap).
INSERT INTO public.schema_migrations (version, description) VALUES
  ('0203', 'rename templates v2 object names to current naming'),
  ('0204', 'add documents.revision_title for governed revision metadata'),
  ('0205', 'make documents.revision_number zero based for governed REV labels'),
  ('0206', 'add artifact metadata to document_revisions'),
  ('0207', 'enforce page_count provenance coupling on document_revisions'),
  ('0208', 'add schedule_generation discriminator for scheduled publish jobs'),
  ('0209', 'add scheduled supersede target on documents'),
  ('0210', 'rename registry capabilities to controlled_documents namespace'),
  ('0211', 'add tenant_id to editor_sessions for document tenant isolation'),
  ('0212', 'idempotency keys two-phase schema alignment for curated baseline'),
  ('0213', 'convert templates tenant_id columns from text to uuid'),
  ('0214', 'add canonical tenants master table and convert auth_sessions tenant_id to uuid'),
  ('0215', 'add metaldocs.materialize_dispatch_outbox transactional outbox table'),
  ('0216', 'add approval_stage_instances.skip_reason column required by approval code'),
  ('0217', 'seed view-grade caps metrics.view/membership.view/user.view/taxonomy.view per ADR 0016'),
  ('0218', 'PR-2: broaden audit.read grants + add session.manage capability (ADR 0019)'),
  ('0219', 'PR-4: iam_users last-login context columns (ip/user_agent/device_label)'),
  ('0220', 'PR-9: iam_users.last_seen_at + tenant/last_seen_at index for WS presence stream'),
  ('0221', 'PR-8: tenant_plans table for Admin Center Usage card (read-only at Tier-B)'),
  ('0222', 'PR-7 retro-land: iam_users mfa metadata + auth_identities last_failed_login_* (originally orphan migrations/0210)'),
  ('0223', 'PR-7b retro-land: auth_identities partial index on last_failed_login_at (originally orphan migrations/0211)'),
  ('0225', 'ADR 0022 Phase 2: seed document-lifecycle write caps doc.publish/doc.obsolete/doc.supersede/template.archive'),
  ('0226', 'ADR 0022 Phase 8: seed registered phantom caps doc.edit_draft/doc.reconstruct/workflow.instance.cancel/doc.view_published'),
  ('0227', 'ADR 0022 Phase 10: merge redundant phantom caps doc.edit_draft/doc.reconstruct/workflow.instance.cancel/doc.view_published into document.edit/document.view'),
  ('0228', 'ADR 0022 Phase 11 (F7): reserve area_code=tenant — a real process area can no longer shadow the tier-2 area-filter sentinel'),
  ('0229', 'ADR 0022 Phase 12 (F3): rename document-lifecycle caps doc.publish/doc.obsolete/doc.supersede -> document.*'),
  ('0230', 'ADR 0022: decommission legacy reviewer area role (drop from user_process_areas CHECK + stored procs) and remove duplicated stored-proc role-list (table CHECK is the single source of truth)'),
  ('0231', 'DB hardening: fail-close enforce_capability_asserted ELSE branch; drop orphaned public.grant/revoke_area_membership duplicate stored procs'),
  ('0232', 'Drop document_access_policies: dead pre-unification per-document ACL table removed with its OpenAPI/FE/permissions slice (ADR 0022 unified model has no per-document ACL concept)'),
  ('0233', 'add templates_template_version.revision_number (ADR 0013 schema half)'),
  ('0234', 'RLS defense-in-depth: ENABLE+FORCE+tenant_isolation policy on public.controlled_documents and metaldocs.audit_events (ADR 0027 Tier 1, F-12, D-3, REQ-TEN-1)'),
  ('0235', 'auth_failure_counters: Postgres-backed auth-failure rate-limiter table (F-20e, REQ-REL-3, D-1)'),
  ('0236', 'dead schema drop (FE-5): templates.areas/visibility/specific_areas, document_profiles.is_active, document_subjects table'),
  ('0237', 'RLS defense-in-depth on all 27 remaining tenant-scoped tables + idempotency_keys tenant FK (Wave Z Z-2/Z-3, F-12 tail, RF-6, REQ-TEN-1, F-09d, ADR 0027 executed in full)'),
  ('0238', 'Drop orphan metaldocs.documents.subject_code column + index (Wave Z Z-25, CD T-010, 0236 residual)'),
  ('0239', 'pg_trgm GIN indexes on controlled_documents(code, title) for ILIKE search (Wave Z Z-24, F-20b)'),
  ('0240', 'drop legacy MDDM document cluster (metaldocs.documents + 8 satellites + metaldocs.template_audit_log); removes public.* shadow under metaldocs-first search_path'),
  ('0241', 'fix enforce_placeholder_value_tenant_consistent() to resolve revision_id via document_revisions (was looked up against documents.id; raised on every placeholder fill-in write)'),
  ('0242', 'iam publishes metaldocs.v_active_user_areas active-membership view (M3/F3.1, ADR-0039 D3a)'),
  ('0243', 'controlleddocuments publishes v_cd_search_facts + v_cd_grantee search visibility contract (M4/F4.1, ADR-0039 D3a)'),
  ('0244', 'documents publishes v_document_search_facts search projection contract (M4/F4.2, ADR-0039 D3a)'),
  ('0245', 'controlleddocuments publishes metaldocs.v_cd_obligated_readers obligated-reader view (M2/F2.1a, ADR-0040)'),
  ('0246', 'taxonomy publishes metaldocs.v_process_area_name per-area label view (M2/F2.1b, ADR-0041)'),
  ('0247', 'notifications module per-recipient inbox table (M3/F3.2, ADR-0043)'),
  ('0248', 'token dictionary per-tenant name->value constants table (SP-1)'),
  ('0249', 'widen document_placeholder_values source CHECK to add dictionary (SP-2 dictionary token substitution)'),
  ('0250', 'template version docx_storage_key de-share + UNIQUE (Phase 1)'),
  ('0251', 'template version status/content_hash/version_number CHECK constraints (F-DB1, F-DB4, F-DB7)'),
  ('0252', 'one published version per template partial-unique index (ADR 0013 / F-DB2)'),
  ('0253', 'document_exports: add tenant_id NOT NULL + FK + RLS + replace unique index (F-O4 HIGH, REQ-TEN-1)'),
  ('0254', NULL),  -- staging_outbox_dead_lettered_at registered version-only upstream; NULL description is faithful to a full migration replay

  ('0255', 'add enforce_template_version_tenant_consistent trigger on templates_template_version to close cross-tenant write gap (F-DB5)'),
  ('0256', 'templates_template_version: add tenant_id NOT NULL + FK + RLS + trigger branch resolves real tenant_id (F-DB5 MEDIUM, REQ-TEN-1)')
ON CONFLICT (version) DO NOTHING;

INSERT INTO metaldocs.tenants (id, name, slug)
VALUES ('ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid, 'System Tenant', 'system')
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  slug = EXCLUDED.slug,
  updated_at = now();

-- Canonical runtime capability matrix.
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'template.approve', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'membership.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'controlled_documents.obsolete', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'controlled_documents.supersede', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'controlled_documents.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'controlled_documents.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'template.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'controlled_documents.obsolete', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'controlled_documents.supersede', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'route.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'taxonomy.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.approve', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.publish', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('signer', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('signer', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'audit.read', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.signoff', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'membership.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'controlled_documents.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'controlled_documents.obsolete', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'controlled_documents.supersede', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'route.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'taxonomy.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.approve', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.create', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.edit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.publish', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.submit', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'user.manage', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'tenant.onboard', 'Provision and onboard a new tenant') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'tenant.export', 'Export a tenant''s data') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'tenant.erase', 'Erase (crypto-shred) a tenant''s data') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'document.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'template.view', '') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- View-grade capabilities per ADR 0016 (mirrors db/migrations/0217_view_grade_capabilities.sql).
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'membership.view', 'Read access policies and area memberships') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'taxonomy.view', 'Read taxonomy profiles, areas, families') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'membership.view', 'Read access policies and area memberships') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'taxonomy.view', 'Read taxonomy profiles, areas, families') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'user.view', 'Read user directory') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'membership.view', 'Read access policies and area memberships') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'taxonomy.view', 'Read taxonomy profiles, areas, families') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'membership.view', 'Read access policies and area memberships') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'taxonomy.view', 'Read taxonomy profiles, areas, families') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'membership.view', 'Read access policies and area memberships') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'metrics.view', 'Read /api/v1/metrics Prometheus scrape endpoint') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'taxonomy.view', 'Read taxonomy profiles, areas, families') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'user.view', 'Read user directory and IAM admin overview') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'membership.view', 'Read access policies and area memberships') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'taxonomy.view', 'Read taxonomy profiles, areas, families') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- PR-2 (ADR 0019): broaden audit.read; add session.manage (mirrors db/migrations/0218_iam_caps_audit_session_pr2.sql).
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'audit.read', 'Read governance / audit events') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'audit.read', 'Read governance / audit events') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'audit.read', 'Read governance / audit events') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'session.manage', 'Force-logout / revoke auth sessions') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- ADR 0022 Phase 2: document-lifecycle write caps (mirrors db/migrations/0225_authz_p2_document_lifecycle_grants.sql).
-- ADR 0022 Phase 12 (F3): cap values renamed doc.* -> document.* for single-prefix
-- coherence (resource.action). The historical 0225 migration still seeds the legacy
-- doc.* spelling; forward migration 0229 converges those rows to document.* on apply.
-- system_admin holds these via the tier-2 bypass (not seeded explicitly).
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.obsolete', 'Make a document obsolete') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.obsolete', 'Make a document obsolete') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.supersede', 'Supersede a document with a successor') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.supersede', 'Supersede a document with a successor') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
-- M6 F6.2 / ADR 0069: document.review — mark-reviewed governance action. Held by
-- the same governance actors as obsolete/supersede (area_admin, qms_admin);
-- system_admin reaches it via the tier-2 bypass (not seeded explicitly).
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'document.review', 'Record a periodic-review completion on a published document') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'document.review', 'Record a periodic-review completion on a published document') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
-- SoD boundary (by design): the terminal template transitions are held by
-- qms_admin (template.publish is also seeded to system_admin above; template.archive
-- is qms_admin-only, with system_admin reaching it via the tier-2 bypass). The
-- approver role deliberately holds template.approve/review/view but NOT
-- template.archive or template.publish: the capability that decides a template
-- (approve) is separated from the capability that executes its terminal lifecycle
-- transition (archive/publish). To broaden archive, grant the template.archive
-- capability to another role here — never re-reason in roles.
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.archive', 'Archive a template') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- SP-1 token dictionary capabilities (ADR superseding 0008).
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'token.view', 'Read tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('author', 'token.view', 'Read tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('editor', 'token.view', 'Read tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'token.view', 'Read tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'token.view', 'Read tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', 'token.view', 'Read tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'token_dictionary.manage', 'Create, update, delete tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'token_dictionary.manage', 'Create, update, delete tenant token dictionary entries') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- ADR 0051: template.manage — elevated approval-config governance capability.
-- Granted to qms_admin (was the former isOperator role-string check target).
-- system_admin accesses via tier-2 bypass (not seeded explicitly).
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'template.manage', 'Edit approval config of published templates') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- M2b F3 (approval-kernel-backend): approval.review mirrors the existing
-- document.signoff pool (same real actors who act on an approval stage);
-- approval.oversee (tenant-grade read oversight) goes to qms_admin, the
-- closest existing role to the design's "quality-manager profile".
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('approver', 'approval.review', 'Act on a review-kind approval stage (suggestions/parecer)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('area_admin', 'approval.review', 'Act on a review-kind approval stage (suggestions/parecer)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'approval.review', 'Act on a review-kind approval stage (suggestions/parecer)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('signer', 'approval.review', 'Act on a review-kind approval stage (suggestions/parecer)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'approval.review', 'Act on a review-kind approval stage (suggestions/parecer)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('qms_admin', 'approval.oversee', 'Oversee any approval instance in the tenant (worklist, cockpit observer mode)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;
INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('system_admin', 'approval.oversee', 'Oversee any approval instance in the tenant (worklist, cockpit observer mode)') ON CONFLICT (role, capability) DO UPDATE SET description = EXCLUDED.description;

-- ADR 0022 Phase 10 (F2): the four redundant phantom caps Phase 8 seeded here
-- (doc.edit_draft / doc.reconstruct / workflow.instance.cancel / doc.view_published)
-- were merged into the canonical document.edit / document.view caps — identical
-- grant sets, zero authorization differentiation. Their 22 grant rows are removed
-- (mirrors the reverse-only migration db/migrations/0227_authz_p10_merge_redundant_caps.sql).
-- This reverses Phase 8's registry over-modeling, NOT its security fix: the tier-2
-- surfaces stay enforced, now by document.edit / document.view.

-- System blank template required for controlled document creation defaults.
-- Invariant: this reference-data row is valid only when bootstrap/storage also seeds
-- `system/templates/blank.docx` into the configured attachments bucket.
-- Runtime gates must fail if the object is absent rather than allowing broken editor loads.
--
-- Capability assertion for the system-tenant template writes below. The
-- trg_require_cap_asserted BEFORE trigger rejects any templates_template[_version]
-- write unless metaldocs.asserted_caps lists the matching capability. The seed
-- asserts the same lifecycle caps the application would, rather than disabling the
-- tripwire (mirrors authz.SeedTxIdentity).
SELECT set_config(
  'metaldocs.asserted_caps',
  '[{"cap":"template.create"},{"cap":"template.edit"},{"cap":"template.submit"},{"cap":"template.approve"},{"cap":"template.publish"}]',
  true
);

-- Tenant context for the system tenant. Required by the templates_template_version
-- tenant-consistency trigger (folded migration 0255) and by the tenant_isolation RLS
-- policies (folded 0237/0256), both of which read the tx-local metaldocs.tenant_id GUC
-- exactly as authz.SeedTxIdentity sets it at runtime. The seed asserts the same context
-- the application would, rather than disabling the invariants.
SELECT set_config(
  'metaldocs.tenant_id',
  'ffffffff-ffff-ffff-ffff-ffffffffffff',
  true
);

-- Post baseline-squash (2026-06): the curated baseline is the FULL post-fold schema,
-- so this reference-data bundle now applies against the final shape. The old
-- `visibility` column was dropped by folded migration 0236 and is therefore absent
-- from the INSERT below (its former pre-tail NOT-NULL workaround is no longer needed).
INSERT INTO public.templates_template (
  id, tenant_id, doc_type_code, key, name, description,
  latest_version, published_version_id, created_by, system_owned, archived_at
) VALUES (
  '00000000-0000-0000-0000-000000000101'::uuid,
  'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid,
  'system',
  '__system_blank__',
  'Em branco',
  'System blank template for controlled document creation.',
  1,
  NULL,
  'system',
  true,
  NULL
) ON CONFLICT (id) DO NOTHING;

INSERT INTO public.templates_template_version (
  id, template_id, tenant_id, version_number, status, docx_storage_key, content_hash, metadata_schema,
  -- pending_reviewer_role/pending_approver_role deliberately omitted: the legacy
  -- per-version role-routing columns are retired (ADR 0082 phase c) and dropped
  -- by migration 0306, which runs after this seed in the curated bootstrap. This
  -- seed formerly wrote a vestigial pending_approver_role='system' here; that
  -- write is removed in the same change-set as the DROP (mirrors 0302's
  -- co-removal of the TenantDataPort DELETE) so 0306's pre-drop emptiness assert
  -- holds. The columns still exist at seed time (baseline schema), so omitting
  -- them just takes their DB defaults (reviewer NULL, approver '').
  placeholder_schema, author_id, reviewer_id,
  approver_id, submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version
) VALUES (
  '00000000-0000-0000-0000-000000000102'::uuid,
  '00000000-0000-0000-0000-000000000101'::uuid,
  'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid,
  1,
  'published',
  'system/templates/blank.docx',
  -- SHA-256 of the deterministic system blank.docx produced by
  -- scripts/seed-system-blank-template.ps1 (New-DeterministicBlankDocx: fixed
  -- 2026-05-17 timestamps, NoCompression, fixed OOXML parts, LF line endings).
  -- This is the TRUE content hash of the canonical published artifact (1289 bytes),
  -- verified by hashing the exact bytes the seed script's generator emits, and
  -- satisfies the 0251 integrity constraints (chk_template_version_content_hash =
  -- 64-hex, and chk_template_version_content_hash_non_draft requiring a real hash
  -- for a non-draft version). The former 'system-blank-template' placeholder
  -- (21 chars, published) violated both CHECKs and made a clean bootstrap fail at
  -- 0251.
  '5cdae1bb25103bbc121cdc696ed11eb09aa22041940f199164ebc302f6923d2e',
  '{}'::jsonb,
  '[]'::jsonb,
  'system',
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  TIMESTAMPTZ '2026-05-14 20:40:29.86648+00',
  NULL,
  0
) ON CONFLICT (id) DO NOTHING;

UPDATE public.templates_template
SET
  system_owned = true,
  latest_version = 1,
  published_version_id = '00000000-0000-0000-0000-000000000102'::uuid
WHERE id = '00000000-0000-0000-0000-000000000101'::uuid;

COMMIT;
