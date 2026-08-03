-- 0315_template_routes_profile_keyed.sql
-- ADR 0086 — template approval routes become config-first and profile-keyed.
--
-- ============================================================================
-- WHY
-- ============================================================================
--   Until now a template ROUTE was keyed by the template INSTANCE it governed
--   (subject_kind='template', subject_key=templates_template.id, profile_code
--   NULL by 0297's approval_routes_template_subject_projection_check). That
--   inverted the kernel's own contract: a route is CONFIGURATION (ADR 0081/0082
--   — one route governs every subject of a profile), never a per-instance
--   artifact. It also made template creation un-gateable: there is no route to
--   look up before the template row exists.
--
--   0315 re-keys template routes onto the profile (doc_type_code), exactly as
--   document routes are keyed, and makes the resulting gate enforceable:
--     * template routes carry profile_code, and profile_code = subject_key;
--     * generic templates (doc_type_code = '') stop existing, so every template
--       has a profile whose template route can be required at creation time.
--
--   Hard cutover, no fallback (ADR 0086 "Hard cutover"): existing template
--   routes are keyed by template id and cannot be re-pointed at a profile
--   without inventing data, so they are deleted along with the approval
--   instances that reference them. This is dev/pre-release data only.
--
-- ============================================================================
-- SCOPE / ORDERING FACTS (baseline 0001 + every later migration; the baseline
-- snapshot is STALE — e.g. it still shows templates_approval_config, which
-- 0302 dropped — so it is never authoritative on its own)
-- ============================================================================
--   * approval_instances_route_id_fkey has NO ON DELETE CASCADE — template
--     instances must be deleted before their routes.
--   * approval_signoffs -> approval_instances / approval_stage_instances FKs
--     have NO cascade either; approval_stage_instances -> approval_instances
--     DOES cascade, approval_route_stages -> approval_routes DOES cascade.
--   * approval_review_verdicts (0288) carries approval_instance_id with NO FK
--     — it is swept explicitly so no orphans survive.
--   * release_generations (0310) has document_id NOT NULL, so template-subject
--     rows are impossible; asserted rather than deleted.
--   * enforce_route_immutable() (0287) ends in RETURN NEW. On a BEFORE DELETE
--     row trigger NEW is NULL, so the trigger SILENTLY CANCELS every route
--     delete. trg_route_config_immutable_del is therefore disabled around the
--     purge and re-enabled after, with a hard post-condition assertion.
--   * trg_route_profile_policy / trg_route_stage_profile_policy (0295) are
--     DEFERRABLE INITIALLY DEFERRED constraint triggers armed for DELETE;
--     left enabled they queue deferred events during the purge and the next
--     ALTER TABLE on approval_routes fails with SQLSTATE 55006 (pending
--     trigger events). Disabled around the purge, re-enabled after — no
--     events queue while disabled, so the later ALTERs run clean.
--   * templates_template and templates_template_version carry
--     trg_require_cap_asserted armed for DELETE (0283 returns OLD but still
--     demands metaldocs.asserted_caps). Disabled around the generic-template
--     purge and re-enabled after — same idiom as 0257.
--   * RLS is FORCE on these tables but every tenant_isolation policy carries
--     the ADR 0027 null-GUC escape-hatch branch, so a migration session with no
--     metaldocs.tenant_id GUC sees all rows.
--
--   ROLLBACK: not reversible. The deleted routes/instances/generic templates
--   are gone; restoring the pre-0315 shape means dropping the two new CHECKs
--   (approval_routes_template_subject_key_check,
--   templates_template_doc_type_code_required_check), restoring
--   approval_routes_template_subject_projection_check, and rebuilding
--   approval_routes_active_profile_uq / approval_routes_profile_version_uq
--   without subject_kind.

BEGIN;

-- ── 1. Purge template-subject approval instances (FK order) ─────────────────

DELETE FROM public.approval_review_verdicts v
      USING public.approval_instances i
      WHERE v.approval_instance_id = i.id
        AND i.subject_kind = 'template';

DELETE FROM public.approval_signoffs s
      USING public.approval_instances i
      WHERE s.approval_instance_id = i.id
        AND i.subject_kind = 'template';

-- approval_stage_instances cascade from approval_instances.
DELETE FROM public.approval_instances
      WHERE subject_kind = 'template';

DO $$
DECLARE
    orphans bigint;
BEGIN
    SELECT count(*) INTO orphans FROM public.release_generations WHERE subject_kind = 'template';
    IF orphans > 0 THEN
        RAISE EXCEPTION '0315: % release_generations rows have subject_kind=template; the purge would orphan them', orphans;
    END IF;
END $$;

-- ── 2. Purge template-subject routes (immutability trigger disabled) ────────

ALTER TABLE public.approval_routes DISABLE TRIGGER trg_route_config_immutable_del;
ALTER TABLE public.approval_routes DISABLE TRIGGER trg_route_profile_policy;
ALTER TABLE public.approval_route_stages DISABLE TRIGGER trg_route_stage_profile_policy;

-- approval_route_stages (and their selectors) cascade from approval_routes.
DELETE FROM public.approval_routes
      WHERE subject_kind = 'template';

ALTER TABLE public.approval_routes ENABLE TRIGGER trg_route_config_immutable_del;
ALTER TABLE public.approval_routes ENABLE TRIGGER trg_route_profile_policy;
ALTER TABLE public.approval_route_stages ENABLE TRIGGER trg_route_stage_profile_policy;

DO $$
DECLARE
    remaining bigint;
BEGIN
    SELECT count(*) INTO remaining FROM public.approval_routes WHERE subject_kind = 'template';
    IF remaining > 0 THEN
        RAISE EXCEPTION '0315: % template-subject approval_routes survived the purge (BEFORE DELETE trigger still armed?)', remaining;
    END IF;
END $$;

-- ── 3. Projection CHECK: template routes are profile-keyed ─────────────────
-- Replaces 0297's "template routes have NO profile" rule with its inverse.
-- This CHECK is the DB backstop for ADR 0086: there is no tripwire arm for
-- approval_routes to regenerate.

ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_template_subject_projection_check;

ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_template_subject_key_check;
ALTER TABLE public.approval_routes
    ADD CONSTRAINT approval_routes_template_subject_key_check
    CHECK (subject_kind <> 'template' OR (profile_code IS NOT NULL AND profile_code = subject_key));

-- approval_routes_document_profile_fk is table-wide FOREIGN KEY
-- (tenant_id, profile_code) REFERENCES metaldocs.document_profiles(tenant_id, code),
-- so template rows inherit profile referential integrity with no new FK.

-- ── 4. Unique indexes gain subject_kind ────────────────────────────────────
-- 0287 created both indexes on (tenant_id, profile_code[, version]) with no
-- subject_kind; template routes dodged them only because profile_code was NULL.
-- Post-cutover an active template route and an active document route on the
-- same profile are both legal and must not collide (23505).

DROP INDEX IF EXISTS public.approval_routes_active_profile_uq;
CREATE UNIQUE INDEX approval_routes_active_profile_uq
    ON public.approval_routes (tenant_id, subject_kind, profile_code)
    WHERE active;

DROP INDEX IF EXISTS public.approval_routes_profile_version_uq;
CREATE UNIQUE INDEX approval_routes_profile_version_uq
    ON public.approval_routes (tenant_id, subject_kind, profile_code, version);

-- Index NAMES are load-bearing: internal/modules/approval/infrastructure/errors.go
-- string-matches "approval_routes_active_profile_uq" to raise
-- ErrDuplicateRouteProfile. They are rebuilt, never renamed.

-- ── 5. Exterminate generic templates (doc_type_code = '') ──────────────────
-- Order mirrors internal/modules/templates/infrastructure/tenant_data_port.go:
-- clear the non-cascading self-reference first, assert no cross-module pointers
-- into the doomed versions, then versions, then templates. templates_approval_
-- config is NOT swept — 0302 dropped that table (ADR 0082 phase c).

ALTER TABLE public.templates_template DISABLE TRIGGER trg_require_cap_asserted;
ALTER TABLE public.templates_template_version DISABLE TRIGGER trg_require_cap_asserted;

UPDATE public.templates_template
   SET published_version_id = NULL
 WHERE doc_type_code = ''
   AND system_owned = false
   AND published_version_id IS NOT NULL;

DO $$
DECLARE
    dangling bigint;
BEGIN
    SELECT count(*) INTO dangling
      FROM metaldocs.document_profiles p
      JOIN public.templates_template_version v ON v.id = p.default_template_version_id
      JOIN public.templates_template t ON t.id = v.template_id
     WHERE t.doc_type_code = '' AND t.system_owned = false;
    IF dangling > 0 THEN
        RAISE EXCEPTION '0315: % document_profiles point default_template_version_id at a generic template; generic templates were never publishable to a profile', dangling;
    END IF;

    SELECT count(*) INTO dangling
      FROM public.controlled_documents c
      JOIN public.templates_template_version v ON v.id = c.override_template_version_id
      JOIN public.templates_template t ON t.id = v.template_id
     WHERE t.doc_type_code = '' AND t.system_owned = false;
    IF dangling > 0 THEN
        RAISE EXCEPTION '0315: % controlled_documents override_template_version_id point at a generic template', dangling;
    END IF;
END $$;

DELETE FROM public.templates_template_version v
      USING public.templates_template t
      WHERE v.template_id = t.id
        AND t.doc_type_code = ''
        AND t.system_owned = false;

DELETE FROM public.templates_template
      WHERE doc_type_code = ''
        AND system_owned = false;

ALTER TABLE public.templates_template ENABLE TRIGGER trg_require_cap_asserted;
ALTER TABLE public.templates_template_version ENABLE TRIGGER trg_require_cap_asserted;

-- ── 6. doc_type_code is mandatory for author-created templates ─────────────
-- Added only after the generic rows are gone. system_owned rows (the seeded
-- blank template) never pass through CreateTemplate and keep their empty code.

ALTER TABLE public.templates_template
    DROP CONSTRAINT IF EXISTS templates_template_doc_type_code_required_check;
ALTER TABLE public.templates_template
    ADD CONSTRAINT templates_template_doc_type_code_required_check
    CHECK (system_owned = true OR (doc_type_code IS NOT NULL AND doc_type_code <> ''));

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0315', 'ADR 0086: template approval routes become config-first and profile-keyed. Deletes every subject_kind=template approval_route (they were keyed by template id and cannot be re-pointed at a profile) together with its dependent chain -- approval_review_verdicts (no FK), approval_signoffs, approval_instances (stage_instances cascade), route_stages cascade -- with trg_route_config_immutable_del disabled around the purge because enforce_route_immutable() RETURNs NEW and therefore silently cancels every BEFORE DELETE. Replaces approval_routes_template_subject_projection_check (profile_code IS NULL for template rows) with approval_routes_template_subject_key_check (profile_code IS NOT NULL AND profile_code = subject_key). Rebuilds approval_routes_active_profile_uq and approval_routes_profile_version_uq with subject_kind so a template route and a document route may share a profile; index names preserved because approval/infrastructure/errors.go string-matches them. Exterminates generic templates (doc_type_code = '''' AND system_owned = false) in tenant_data_port order -- NULL published_version_id, assert no document_profiles/controlled_documents pointers, delete versions, templates -- with trg_require_cap_asserted disabled on both templates tables (0257 idiom), then adds templates_template_doc_type_code_required_check. No tripwire arm exists for approval_routes; the new CHECK is the DB backstop.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
