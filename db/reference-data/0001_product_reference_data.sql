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

-- Folded migration ledger (baseline squash 2026-07-29): migrations 0257..0315
-- (55 files; 0261, 0280, 0289 and 0291 never existed; 0309 self-registered no
-- ledger row upstream so only 54 rows are seeded) were folded into
-- db/baseline/0001_current_schema.sql and archived under
-- archive/migrations/post-baseline-2026-07-fold/. Seed their ledger rows so the
-- runner's ledger-skip treats them as already applied (the archived files are
-- never re-run on a fresh bootstrap). Descriptions are verbatim from the
-- reference database that replayed the full chain, so a folded bootstrap and a
-- full replay produce byte-identical schema_migrations content.
INSERT INTO public.schema_migrations (version, description) VALUES
  ('0257', 'templates_template_version: rename status in_review->under_review to match documents vocabulary (M1·T3)'),
  ('0258', 'SEC-01: tenant-scope metaldocs.document_families (tenant_id + RLS + tripwire attribution), mirrors document_profiles/document_process_areas sentinel pattern'),
  ('0259', 'SEC-05: attach enforce_capability_asserted tripwire to metaldocs.iam_users (column-scoped UPDATE excluding login/presence telemetry cols) and iam_group* (user.manage); documents/iam_user_roles/user_process_areas already covered (0188), not touched'),
  ('0260', 'DB-02/DB-04(c)/DB-05: drop dead tables — metaldocs.documents (no-op safety net, already gone via archived 0240), public.template_audit_log (zero code refs, RLS-hardened but unread/unwritten), metaldocs.document_template_versions_mddm + metaldocs.template_drafts (MDDM shadow residue, zero code refs); metaldocs.document_template_versions explicitly NOT dropped (live: metaldocs-e2e-seed reads it); DB-04(a) subject_code and DB-04(b) editable_zones already closed by archived 0238/0196, no action'),
  ('0262', 'DB-06/T-013: backfill any pre-Wave-1.8 templates_audit_log rows into the canonical metaldocs.audit_events hash chain, then drop public.templates_audit_log (module-local audit sink; write path already canonical since F-07-sub-split, read path already canonical since same wave)'),
  ('0263', 'DB-07: validate the last NOT VALID constraints — audit_events prev_hash/row_hash shape CHECKs (approval iam_users FKs from the original finding were already superseded by validated composite tenant FKs; register drift)'),
  ('0264', 'DB-08 (T-015/R-015) half A: promote metaldocs.document_process_areas PK from (code) to (tenant_id, code) by converting the ux_process_areas_tenant_code unique index into the PK''s backing index (PRIMARY KEY USING INDEX — keeps the index OID so the 5 inbound composite FKs bound to it stay valid; index renamed to document_process_areas_pkey). document_profiles half deferred (no DDL), see wiki/modules/taxonomy-tech-debt.md T-015 design note.'),
  ('0265', 'Grade-A simplification: tighten public.documents_status_check from 10 to 9 live values, dropping dead legacy literal ''finalized'' (retired by migration 0142, confirmed zero code producers on both Go and FE sides). ''archived'' explicitly kept (live). ''frozen'' (FE StatusPill-only, no Go producer) explicitly NOT added. Pre-check DO block RAISEs before the swap if any existing row would violate.'),
  ('0266', 'DB-09 (P3, bounded): (a) durably REVOKE UPDATE/DELETE/TRUNCATE on metaldocs.audit_events from metaldocs_app (privileges were never granted — this hardens the already-correct insert/select-only posture documented by closed T-004/T-009, no distinct audit-writer role exists so none was invented); (b) add audit_events_payload_size_cap CHECK (octet_length(payload::text) <= 65536), added NOT VALID + VALIDATE CONSTRAINT for low-downtime rollout, pre-checked by a RAISE-on-violation DO block (T-010). Partitioning, pg-cron retention, ULID/UUIDv7 ids deferred — see wiki/modules/audit-tech-debt.md.'),
  ('0267', 'FE-08 (P3): add templates_template.updated_at (timestamptz NOT NULL DEFAULT now(), backfilled to created_at). No DB-level auto-touch trigger convention exists in this schema, so application UPDATE paths in internal/modules/templates/repository/postgres.go explicitly SET updated_at = now() (shipped in the same change). Closes the wiki/backlog/templates.md TemplateDTO updated_at gap.'),
  ('0268', 'DB-01/DB-10: drop legacy template family (public.templates, public.template_versions — canonical templates_template/templates_template_version is the live path since ARC-01 4160018e; docgenv2 legacy fallback reader deleted in the same change-set as this migration''s promotion out of _pending/) and legacy public.signoffs (safety-net no-op; table absent from canonical baseline, confirmed zero runtime writers/readers). GATED: apply only after docgenv2.LegacyTemplateReadCount() is observed at zero across a full QA run window.'),
  ('0269', 'Goal-3 QA: add template.review to the templates_template_version tripwire cap list — Service.Review writes the version row under CapTemplateReview and the reviewer stage was fail-closed into a 500 (P0001) on every accept/reject. Function-only swap; all other branches preserved from 0259.'),
  ('0270', 'Grade-A review APP-09: add template.archive to the templates_template tripwire cap list — Service.ArchiveTemplate updates the template row under CapTemplateArchive and archive was fail-closed into a 500 (P0001) for every actor. Function-only swap; all other branches preserved from 0269.'),
  ('0271', 'M2 F2.1: widen documents/UPDATE tripwire arm to {document.edit, document.obsolete, membership.manage} -- ForceReleaseSession/ForceReleaseSessionTx (repository.go:798,828) and MarkObsolete (obsolete_service.go:88->93) each assert a single non-document.edit capability in the same function as the mutating UPDATE and were fail-closed P0001 for every actor, same defect class as 0269/0270. Additive-only; all other branches preserved from 0270; machine-generated from internal/platform/tripwire.'),
  ('0272', 'Remove dead ''rejected'' document status: tighten documents_status_check to 8 live values and rewrite enforce_document_transition() to drop the under_review->rejected and rejected->draft arcs, restoring app<->DB parity with documents/domain.CanTransitionDocumentStatus.'),
  ('0273', 'Drop metaldocs.job_leases and its acquire_lease/heartbeat_lease/release_lease/assert_lease_epoch functions — custom lease scheduler retired by M5 async-River consolidation (ADR 0067); River owns leader/lease lifecycle in river_leader.'),
  ('0274', 'M6 F6.2/F6.3: add eQMS review/expiry + structured reason columns to public.documents (review_due_at, last_reviewed_at, reason_for_change, reason_category), all nullable/expand-only. Add NULL-tolerant CHECKs ck_documents_effective_window (effective_to > effective_from), ck_documents_review_due_sane (review_due_at >= effective_from) and ck_documents_reason_category (enum). Reuses existing effective_from/effective_to; no new duplicate column family; no backfill.'),
  ('0275', 'M6 F6.2: widen documents/UPDATE tripwire arm to {document.edit, document.obsolete, membership.manage, document.review} -- the mark-reviewed workflow asserts only document.review (ADR 0069) then UPDATEs documents (last_reviewed_at, review_due_at) and would be fail-closed P0001 for every actor without the arm, same defect class as 0269/0270/0271. Additive-only; all other branches preserved from 0271; machine-generated from internal/platform/tripwire.'),
  ('0276', 'M6 F6.2 T4: add public.documents.review_surfaced_at (nullable timestamptz) -- idempotency marker for the River periodic review-due surfacer (jobs module). A run''s MarkSurfaced UPDATE sets it to the run timestamp for docs newly due; the guard review_surfaced_at IS NULL OR review_surfaced_at < review_due_at makes a same-cycle rerun a no-op and a later review_due_at advance (mark-reviewed) re-eligible. Expand-only, no backfill, no tripwire function change.'),
  ('0277', 'M7 F7.2 (ADR 0070): add tenants/INSERT tripwire arm requiring tenant.onboard and attach trg_require_cap_asserted to metaldocs.tenants (BEFORE INSERT) -- the onboarding workflow asserts only tenant.onboard then INSERTs metaldocs.tenants, previously the one privileged provisioning write with no DB tripwire backstop. Additive-only; all other branches preserved from 0275; machine-generated from internal/platform/tripwire.'),
  ('0278', 'M7 F7.3 Task A: create metaldocs.tenant_lifecycle_jobs (export/erase job tracking, tenant-scoped, FORCE RLS + tenant_isolation policy per 0258 idiom) and add metaldocs.tenants.erased_at tombstone column.'),
  ('0279', 'M7 F7.3 Task A (ADR 0070): add tenant_lifecycle_jobs/INSERT tripwire arm requiring {tenant.export, tenant.erase} (match-one) and attach trg_require_cap_asserted to metaldocs.tenant_lifecycle_jobs (BEFORE INSERT) -- the export/erase enqueue workflows assert only one of tenant.export/tenant.erase then INSERT tenant_lifecycle_jobs, previously with no DB tripwire backstop. Additive-only; all other branches preserved from 0277; machine-generated from internal/platform/tripwire.'),
  ('0281', 'M7 F7.3 Task B (ADR 0070 d4): create metaldocs.tenant_keys (tenant_id PK/FK, wrapped_dek, created_at, destroyed_at) with FORCE RLS + tenant_isolation policy -- backs the TenantCrypto DEK/KEK crypto-shred port.'),
  ('0282', 'M7 F7.3 Task E: amend public.reject_user_process_areas_delete() to allow DELETE on public.user_process_areas when the tx-local metaldocs.tenant_erasure GUC is set (tenant erasure lifecycle context); all other callers still P0001 with the identical message.'),
  ('0283', 'M7 F7.3 Task F fix: enforce_capability_asserted() returned NEW on every path -- NULL for DELETE, silently cancelling every DELETE on a trigger-carrying table. Now RETURN OLD when TG_OP is DELETE (both bypass and cap-asserted exits), v_tenant_id = COALESCE(NEW.tenant_id, OLD.tenant_id), governance resource_id falls back to OLD.id. No arm/cap change; branches byte-for-byte from 0279; machine-generated from internal/platform/tripwire.'),
  ('0284', 'M7 F7.4 (validation-contract §4.1/§4.2): create dedicated non-owner NOSUPERUSER+NOBYPASSRLS metaldocs_ci role with DML-only grants (USAGE + SELECT/INSERT/UPDATE/DELETE on all tables + sequences in metaldocs/public, ALTER DEFAULT PRIVILEGES for future objects) so the integration suite exercises RLS for real. DEVIATION (HS-1): dedicated non-owner role instead of literal 63-table ownership-reassignment -- same non-owner+non-bypass property, avoids a dev-bootstrap-breaking owner migration. Dev password is a non-secret DML-only fixture; prod overrides via ALTER ROLE.'),
  ('0285', 'M7 F7.4 (validation-contract §4.4): add FORCE ROW LEVEL SECURITY + tenant_isolation policy (on actor_tenant_id) to public.approval_signoffs -- the one tenant table the tenant_id-keyed FORCE-RLS census missed (it keys on actor_tenant_id). Idiom verbatim from 0281/0278 incl. null-GUC escape-hatch branch (ADR 0027). No cross-tenant co-sign, so strict policy is correct.'),
  ('0286', 'M2b F1: purely additive StageKind (review/approval) substrate. Adds approval_route_stages.stage_kind/due_in_days, approval_stage_instances.stage_kind/due_at, approval_instances.frozen_content_hash/cancel_reason, approval_signoffs.signature_meaning, all nullable-or-defaulted so no existing row or code path changes behavior. Named CHECK constraints for later constraint-name assertions. No service wiring (F2-F8).'),
  ('0287', 'M2b F2: route versioning (W1) + empty-pool 422 substrate (W6). Adds approval_routes.superseded_at; replaces single tenant_id+profile_code unique with partial-active-unique + full version-history-unique; enforce_route_immutable() gains a column-scoped exception permitting active/superseded_at-only updates on in-use rows while continuing to block definition-column edits on in-use or inactive rows. No behavior change for never-superseded active routes.'),
  ('0288', 'M2b F4: review-stage runtime verdicts substrate. Widens approval_instances_status_check to admit changes_requested (5th, non-terminal status). Adds approval_review_verdicts (ready/request_changes, one verdict per actor per stage via UNIQUE(stage_instance_id, actor_user_id)), RLS ENABLE+FORCE + tenant_isolation policy on actor_tenant_id copied verbatim from 0285. No reuse of approval_signoffs row shape (only EvaluateQuorum counting fn reused).'),
  ('0290', 'M2b F7: unified SoD DB-tripwire. New enforce_approval_sod() function (rule text mirrors enforce_signoff_sod verbatim), attached to both approval_signoffs (trg_signoff_sod, re-pointed, zero behavior change) and approval_review_verdicts (trg_review_verdict_sod, new — closes the F4 DB-enforcement parity gap). App-level domain.CheckSoD is now called identically from both RecordSignoff and RecordVerdict.'),
  ('0292', 'M2b F8: SLA snapshot + surfacer substrate. Adds approval_stage_instances.due_in_days_snapshot (per-instance snapshot of the route stage''s due_in_days config, read at stage-activation time to compute due_at without a second read against a possibly-superseded route version) and approval_stage_instances.sla_surfaced_at (idempotency marker for the new approval_sla_surfacer periodic job, distinct from and NOT reusing documents.review_surfaced_at). No existing row or code path changes behavior until F8''s service/repository/job wiring populates these columns.'),
  ('0293', 'M2b F9: approval delegation (ADR 0077). New approval_delegations table (tenant/delegator/delegate/window/reason/audit, RLS ENABLE+FORCE + tenant_isolation, no-self + window CHECKs, union semantics for overlapping windows). New on_behalf_of_user_id column on approval_signoffs and approval_review_verdicts (dual-identity persistence, NULL when acting as self). enforce_approval_sod() widened (CREATE OR REPLACE, same triggers) to also reject when on_behalf_of_user_id equals the document author — DB-tripwire parity with the app-level domain.CheckSoD widening.'),
  ('0294', 'M2d F2d.2 (ADR 0079): actor_display_name_snapshot DB-enforced invariant. Backfills NULL/empty from metaldocs.iam_users.display_name (NOT NULL source), then SET NOT NULL + CHECK (<> '''') on both approval_review_verdicts and approval_signoffs. Closes the eQMS audit-truth hole (nullable snapshot + read fallback let a rename rewrite signed history); lets F2d.2 delete the write-conditional bind and the read coalesce/live-lookup fallbacks. Forward-only, idempotent; baseline frozen (ADR 0032).'),
  ('0295', 'G1 per-profile signature policy: metaldocs.document_profiles.governance_class text NOT NULL DEFAULT ''controlado'' + CHECK (controlado/simples/livre); shared assert_route_shape() helper + enforce_profile_route_policy() DEFERRABLE INITIALLY DEFERRED constraint triggers on approval_route_stages (stage-shape), approval_routes (route-row insert/activation — closes zero-stage + activate-later paths), and document_profiles (reclassify guard). controlado routes need >=1 approval stage, livre routes forbidden, simples review-only allowed; only active routes enforced (superseded versions and in-flight snapshots untouched).'),
  ('0296', 'P2.S1 approval subject generalization (expand phase, ADR 0082): approval_routes/approval_instances gain subject_kind text NOT NULL CHECK (document|template) + subject_key text NOT NULL; backfill subject_kind=''document'', subject_key=profile_code (routes) / document_id::text (instances); public.default_approval_subject() BEFORE INSERT compatibility trigger backfills both columns from the legacy document column on any INSERT that omits them (no Go cutover required this phase); profile_code and document_id stay NOT NULL backward-compat aliases; new tenant-scoped indexes ux_approval_routes_tenant_subject, ux_approval_instances_active_subject (partial, in_progress), ix_approval_instances_tenant_subject added alongside the kept document-keyed constraints/indexes.'),
  ('0297', 'P3.S2b-0 relax legacy document-only NOT NULL for template approval subjects (M3 kernel extraction): approval_instances.document_id and approval_routes.profile_code go NULL-able; both existing FKs (approval_instances_document_id_fkey, approval_routes_document_profile_fk) are KEPT as-is and remain NULL-tolerant under MATCH SIMPLE, so document-row referential integrity is unchanged; projection CHECK constraints (approval_instances/_routes _document_subject_projection_check + _template_subject_projection_check) enforce subject_kind vs legacy-column presence as the new DB-level invariant; subject-scoped partial unique index ux_approval_instances_subject_idempotency (tenant_id, subject_kind, subject_key, idempotency_key) added alongside the kept document_id-keyed approval_instances_document_id_idempotency_key_key so template submissions get real dedup once document_id is NULL (Go submit-path cutover to use it is out of scope for this slice, tracked P3.S2b-1).'),
  ('0298', 'M3 P3.S2b-2 regression fix: normalize pre-existing divergent approval_routes dev rows (subject_key := profile_code where subject_kind=document) and add approval_routes_document_subject_key_alias_check CHECK (subject_kind <> ''document'' OR subject_key = profile_code) so the DB enforces the document-route profile_code alias invariant as the last line, matching the new app-level guard in resolveCreateRouteSubject (ErrDocumentSubjectKeyMismatch).'),
  ('0299', 'M3 P3.S2b-3b-0 (ADR 0083): approval_instances/INSERT arm split into two subject-discriminated entries (subject_kind=document -> document.submit; subject_kind=template -> template.submit, never unioned) via a nested CASE NEW.subject_kind with a fail-closed ELSE. No other arm/cap change; every other branch byte-for-byte from 0283; machine-generated from internal/platform/tripwire.'),
  ('0300', 'M3 P3.S2b-3b-iii-a (ADR 0083 follow-on): approval_signoffs/INSERT arm split into two parent-lookup-discriminated entries (parent approval_instances.subject_kind=document -> document.signoff; subject_kind=template -> template.approve, never unioned) via a SELECT of the parent approval_instances row subject_kind + a nested CASE with a fail-closed ELSE. No other arm/cap change; every other branch byte-for-byte from 0299; machine-generated from internal/platform/tripwire.'),
  ('0301', 'Unit 3.1a S5 (ADR 0082 phase c): template.review removed from the templates_template_version tripwire arm -- the legacy reviewer stage (Service.Review, deleted in S4) was its only asserting writer and the capability itself is retired from the IAM registry in this change-set; reverses 0269 for the deleted stage. No other arm/cap change; every other branch byte-for-byte from 0300; machine-generated from internal/platform/tripwire.'),
  ('0302', 'Unit 3.1a S5 (ADR 0082 phase c): drop public.templates_approval_config behind a pre-drop emptiness assert -- legacy role-based template-approval config retired. S1 removed the create-time writer, S4 deleted the 4 legacy routes and every remaining reader/writer, and the templates TenantDataPort erase DELETE is removed in the same change-set. The approval kernel route model is the sole approval configuration.'),
  ('0303', 'Unit 3.2 slice 2 (M4 ActorSelector, B2 expand): new child table public.approval_route_stage_selectors (tenant_id, route_stage_id FK CASCADE, selector_order, kind, user_id/role/area_code with per-kind CHECK mirroring domain ActorSelector.Validate) + tenant-scoped indexes; backfill one role_in_fixed_area(required_role, area_code) selector per existing approval_route_stages row; end-of-migration verification asserts every stage has >=1 selector. Flat required_role/area_code columns retained this slice -- dropped in migration 0304 (contract step, slice 6) only after resolver/route-admin/submit cut over.'),
  ('0304', 'Unit 3.2 slice 6a (M4 ActorSelector, Option C prime drift-source cutover, additive): new column public.approval_stage_instances.selectors_snapshot jsonb NOT NULL DEFAULT [] freezing the stage''s actor selectors onto the instance at submit, materializing submit_choice into named_user at write time; backfill synthesizes one role_in_fixed_area(required_role_snapshot, area_code_snapshot) selector per existing in-flight row (not an empty array, to preserve quorum denominators under re-evaluate drift policies) -- this is the parity mapping to the pre-cutover flat resolver. Flat route-side and instance-side snapshot columns retained -- contract step (drop) is a later slice.'),
  ('0305', 'Unit 3.2 slice 6b (M4 ActorSelector, contract step): drop flat public.approval_route_stages.required_role/area_code columns now that approval_route_stage_selectors (0303) is the sole source of truth for a route stage''s actor pool; flat<->selector synthesis relocates to the HTTP boundary (route_admin_handler.go). Instance-side required_role_snapshot/area_code_snapshot audit columns and ResolveEligibleActors unchanged.'),
  ('0306', 'ROADMAP unit 4.4 (schema hygiene): drop public.templates_template_version.pending_reviewer_role/pending_approver_role behind a pre-drop emptiness assert -- legacy per-version role-routing columns retired since 3.1a (ADR 0082 phase c). Write-never by Go code since 3.1a S1/S4; the curated bootstrap seed''s vestigial pending_approver_role=''system'' is co-removed in this change-set (db/reference-data/0001_product_reference_data.sql). Read-never, grep-confirmed zero Go/TS non-test readers. The approval kernel route model is the sole approval configuration.'),
  ('0307', 'Unit 4.4 (schema hygiene, hub-added item): drop dead metaldocs.grant_area_membership(uuid,text,text,text,text). Zero product-runtime callers; the 3 repo references are all test-only infra: internal/test/e2e_seed.go:579 (integration seed helper, never run by Go tests), tests/integration/scenarios/membership_fn_test.go (integration test, RED today, skip-graceful on drop), tests/integration/testdb/db.go:422 (passive metaldocsOwnedObjects name->schema resolver entry, harmless when stale) -- Track B (unit 4.3) prunes all three. Also unsatisfiable by construction: the function requires an UPPERCASE area_code but the FK-referenced metaldocs.document_process_areas.code CHECK area_code_format requires lowercase, so no INSERT could ever succeed. Behind a pre-drop pg_depend assert (normal/auto catalogued-edge coverage only; PL/pgSQL bodies opaque), no CASCADE.'),
  ('0308', 'ROADMAP unit 4.5 (T-015/R-015 half B): drop 4 dead document_profiles-adjacent tables (document_profile_governance, document_profile_schema_versions, document_sequences -- zero Go refs, archive-era 0027/0040s; document_profile_template_defaults -- zero readers, sole writer being deleted in sibling slice, superseded by document_profiles.default_template_version_id), then promote metaldocs.document_profiles PK from (code) to (tenant_id, code) via PRIMARY KEY USING INDEX ux_document_profiles_tenant_code (0264 technique -- keeps the index OID so the 3 inbound composite FKs bound to it, approval_routes_document_profile_fk / cd_sequence_counters_tenant_id_profile_code_fkey / controlled_documents_tenant_id_profile_code_fkey, stay valid untouched; index renamed to document_profiles_pkey, no leftover redundant index).'),
  ('0310', 'ADR 0085 Stage A: add public.documents.planned_effective_from (plan) and re-home effective_from as the ACTUAL release timestamp; move scheduled rows plan -> planned_effective_from and backfill legacy published rows; add ck_documents_published_effective_from and ck_documents_scheduled_planned_effective_from; demote extra published rows per controlled document and add ux_documents_published_head (one published head per (tenant_id, controlled_document_id), the DB enforcement of ADR 0085 supersession-head); create public.release_generations (single-table generation identity + approval/artifact facts + readiness-hold state, FORCE RLS + tenant_isolation); widen materialize/pdf dispatch outbox dedupe from (tenant_id, revision_id) to a generation-aware unique index.'),
  ('0311', 'ADR 0085 Stage B: retire the document.publish capability — delete every metaldocs.role_capabilities grant for it, matching its removal from the Go capability registry and the curated reference-data seed. Publication is no longer a human-invoked action (the release coordinator reacts to durable readiness facts), so no route or service asserts the capability. document.supersede is retained and re-homed to submit-time cross-document plan authorization. Also strand-safe: deletes every NON-TERMINAL public.river_job row of the deleted kind scheduled_publish_cutover (available/scheduled/retryable/pending), guarded by to_regclass because River provisions its own schema at binary startup rather than through this migration series; completed/cancelled/discarded history and leased running rows are left untouched.'),
  ('0312', 'Etapa 5 §5.1 (F-QA4-10): rename each staging dispatch outbox hash column to its verified payload — metaldocs.materialize_dispatch_outbox.content_hash -> values_hash (carries the resolved-placeholder values hash written by FreezeService.Pin) and metaldocs.pdf_dispatch_outbox.content_hash -> frozen_docx_hash (carries the materialized frozen-docx hash produced by the renderer fanout). Hard cutover, no compat alias: the paired River job args JSON field names are renamed with them, and api+worker+jobs deploy together. documents.content_hash, documents.content_hash_at_submit and document_revisions.content_hash are different, correctly-named columns and are untouched.'),
  ('0313', 'Etapa 5 §5.2 (F-QA3-1 option (a) remainder): add public.documents.frozen_revision_id uuid NULL REFERENCES document_revisions(id) — the freeze-time lineage pin. Written in the same UPDATE as values_hash by SnapshotRepository.WriteFreeze and read back by FreezeService.Materialize, which fetches the pinned revision body and fails closed unless its sha256 equals that revision''s recorded content_hash. Expand-only with NO backfill: pre-migration freezes keep a NULL pin and fail closed rather than carry fabricated lineage. documents.content_hash keeps its meaning (hash of the OUTPUT frozen docx).'),
  ('0314', 'Grant DELETE on metaldocs.outbox_events to metaldocs_app so the new outbox-events-retention maintenance job (internal/modules/jobs/outbox_retention, executed by metaldocs-jobs per ADR 0067) can purge terminal rows. The table previously had NO retention at all, so published/dead-lettered rows accumulated forever and each one pinned its idempotency_key against the publisher''s ON CONFLICT (idempotency_key) DO NOTHING dedup, silently swallowing later legitimate re-publishes on the same key. Prior grants were INSERT (0008) and SELECT, UPDATE (0019) only; DELETE was missing, which would have failed the purge closed in any environment where metaldocs_app is not the table owner.'),
  ('0315', 'ADR 0086: template approval routes become config-first and profile-keyed. Deletes every subject_kind=template approval_route (they were keyed by template id and cannot be re-pointed at a profile) together with its dependent chain -- approval_review_verdicts (no FK), approval_signoffs, approval_instances (stage_instances cascade), route_stages cascade -- with trg_route_config_immutable_del disabled around the purge because enforce_route_immutable() RETURNs NEW and therefore silently cancels every BEFORE DELETE. Replaces approval_routes_template_subject_projection_check (profile_code IS NULL for template rows) with approval_routes_template_subject_key_check (profile_code IS NOT NULL AND profile_code = subject_key). Rebuilds approval_routes_active_profile_uq and approval_routes_profile_version_uq with subject_kind so a template route and a document route may share a profile; index names preserved because approval/infrastructure/errors.go string-matches them. Exterminates generic templates (doc_type_code = '''' AND system_owned = false) in tenant_data_port order -- NULL published_version_id, assert no document_profiles/controlled_documents pointers, delete versions, templates -- with trg_require_cap_asserted disabled on both templates tables (0257 idiom), then adds templates_template_doc_type_code_required_check. No tripwire arm exists for approval_routes; the new CHECK is the DB backstop.')
ON CONFLICT (version) DO NOTHING;

-- The system-tenant seed is a privileged provisioning write. Folded migration
-- 0277 attached trg_require_cap_asserted to metaldocs.tenants (BEFORE INSERT,
-- arm {tenant.onboard}, ADR 0070); before the 2026-07-29 fold that trigger did
-- not exist yet when this file ran, so the INSERT below silently pre-dated it.
-- With the trigger now in the baseline, the bootstrap must assert the
-- capability like any other writer or fail closed with P0001
-- (ErrCapabilityNotAsserted). set_config(..., true) is transaction-local and
-- this file is one explicit BEGIN/COMMIT -- same idiom as folded 0310/0267.
SELECT set_config('metaldocs.asserted_caps', '[{"cap":"tenant.onboard"}]', true);

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
