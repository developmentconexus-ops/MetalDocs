-- 0297_approval_subject_relax_legacy_keys.sql
-- ROADMAP P3.S2b-0 (M3 — approval kernel extraction, Phase 3). Follows 0296
-- (P2.S1 approval subject generalization, ADR 0082), which deliberately left
-- approval_instances.document_id and approval_routes.profile_code NOT NULL
-- because "template rows arrive in a later phase, which is the phase that
-- relaxes this NOT NULL". This is that phase.
--
-- Without this migration a (subject_kind='template', subject_key=<template
-- version id>) approval_instances row, or a (subject_kind='template',
-- subject_key=<template id>) approval_routes row, cannot be INSERTed at all:
-- a template subject has no public.documents row and no
-- metaldocs.document_profiles row, so the legacy document-only NOT NULL
-- columns have nothing valid to point at.
--
-- EXPAND phase only, forward-only, idempotent (ALTER COLUMN DROP NOT NULL is
-- naturally idempotent/re-runnable; DROP CONSTRAINT/INDEX IF EXISTS before ADD
-- — idiom from 0286/0287/0290/0295/0296).
--
--   1. approval_instances.document_id and approval_routes.profile_code both
--      go NULL-able. Document rows continue to always populate both (no
--      existing writer changes — P2.S2 Go cutover territory, not this
--      migration); template rows will leave them NULL.
--
--   2. Both FKs (approval_instances_document_id_fkey ->
--      public.documents(id); approval_routes_document_profile_fk
--      (tenant_id, profile_code) -> metaldocs.document_profiles(tenant_id,
--      code)) are DELIBERATELY KEPT, not dropped. This is intentional and
--      correct, not an oversight:
--        - approval_instances_document_id_fkey is a single-column FK. Per
--          the SQL standard (and Postgres), a single-column FK is simply not
--          checked when the referencing column is NULL — MATCH SIMPLE (the
--          Postgres default) never fires the referential check for a NULL
--          value. So a template row (document_id NULL) skips the FK check
--          entirely, while every document row keeps full referential
--          integrity against public.documents(id) exactly as before.
--        - approval_routes_document_profile_fk is a composite
--          (tenant_id, profile_code) FK, also MATCH SIMPLE (the default;
--          this migration does not change the FK's MATCH mode). MATCH
--          SIMPLE only requires ALL referencing columns to be non-NULL
--          before the check fires — if ANY column in the FK is NULL (here,
--          profile_code for a template row), the whole FK is skipped, even
--          though tenant_id itself is NOT NULL and populated. Document rows
--          keep profile_code populated, so their FK is fully enforced
--          against metaldocs.document_profiles unchanged.
--      Net effect: dropping NOT NULL is sufficient by itself to make FKs
--      NULL-tolerant under MATCH SIMPLE — no FK DDL change is needed, and
--      keeping both FKs preserves 100% of existing document-path referential
--      integrity (E-PROD-2 / metaldocs.document_profiles is untouched, per
--      the 0296 out-of-scope note).
--
--   3. Projection CHECK constraints replace the referential integrity the
--      dropped NOT NULLs used to imply, scoped by subject_kind so the DB (not
--      app code) enforces the subject invariant as the last line:
--        - approval_instances_document_subject_projection_check:
--          subject_kind <> 'document' OR document_id IS NOT NULL
--          (a document-subject row MUST carry its document_id — the FK
--          alone no longer guarantees this once document_id is NULL-able).
--        - approval_instances_template_subject_projection_check:
--          subject_kind <> 'template' OR document_id IS NULL
--          (a template-subject row must NOT falsely carry a document_id —
--          guards against a future bug that half-migrates a row).
--        - approval_routes_document_subject_projection_check /
--          approval_routes_template_subject_projection_check: the same pair
--          on profile_code.
--
--   4. Idempotency investigation (submit-path dedup mechanism): SUBMIT
--      idempotency is enforced by the DB unique constraint
--      approval_instances_document_id_idempotency_key_key
--      (document_id, idempotency_key) — see
--      internal/modules/approval/application/submit_service.go step 3 comment
--      and internal/modules/approval/infrastructure/postgres_approval_repository.go
--      InsertInstance, whose 23505 on that exact constraint name is mapped to
--      infrastructure.ErrDuplicateSubmission by MapPgError
--      (internal/modules/approval/infrastructure/errors.go). It is NOT the
--      platform internal/platform/idempotency/postgres_store.go store — the
--      approval submit path never calls that store. Because Postgres unique
--      indexes treat NULL as distinct-from-everything (including itself),
--      once document_id is NULL-able (step 1) that constraint no longer dedups
--      two template-subject submissions carrying the same idempotency_key: two
--      rows with document_id NULL never collide. So this migration ADDS a
--      subject-scoped partial unique index, tenant-scoped and covering BOTH
--      subjects, alongside (not replacing) the existing document_id-keyed
--      constraint, which stays for document-path back-compat and is left
--      untouched:
--        ux_approval_instances_subject_idempotency
--        UNIQUE (tenant_id, subject_kind, subject_key, idempotency_key)
--        WHERE (idempotency_key IS NOT NULL)
--      idempotency_key is already NOT NULL on every row today, so the
--      WHERE clause is a no-op filter now and simply future-proofs the index
--      if that column is ever relaxed later; it does not change today's
--      behavior. This migration is schema-only: the Go submit path is not
--      cut over to use this index in this slice (that is P3.S2b-1 territory,
--      per the task boundary) — MapPgError's constraint-name switch is not
--      touched here.
--
-- Multi-tenant: the new index leads with tenant_id, matching every other new
-- index added in 0296. RLS FORCE is already active on both tables from the
-- baseline; the relaxed columns and new constraints/index inherit the
-- existing table-level tenant_isolation policy — no per-column grants exist
-- anywhere in the baseline for these tables, so none are added here.
--
-- Out of scope / untouched: metaldocs.document_profiles (E-PROD-2); the 0296
-- compat trigger public.default_approval_subject() and its two triggers
-- (still required — not dropped here; DEBT tracked in 0296's own header);
-- any Go application/HTTP code (schema-only slice).

BEGIN;

-- 1. Relax legacy document-only NOT NULL -----------------------------------

ALTER TABLE public.approval_instances
    ALTER COLUMN document_id DROP NOT NULL;

ALTER TABLE public.approval_routes
    ALTER COLUMN profile_code DROP NOT NULL;

-- 2. FKs are intentionally KEPT (see header) — no DDL needed; MATCH SIMPLE
--    already makes both NULL-tolerant once the referencing column(s) can be
--    NULL. No statement here is a deliberate no-op, not an omission.

-- 3. Projection CHECK constraints -------------------------------------------

ALTER TABLE public.approval_instances
    DROP CONSTRAINT IF EXISTS approval_instances_document_subject_projection_check;
ALTER TABLE public.approval_instances
    ADD CONSTRAINT approval_instances_document_subject_projection_check
    CHECK (subject_kind <> 'document' OR document_id IS NOT NULL);

ALTER TABLE public.approval_instances
    DROP CONSTRAINT IF EXISTS approval_instances_template_subject_projection_check;
ALTER TABLE public.approval_instances
    ADD CONSTRAINT approval_instances_template_subject_projection_check
    CHECK (subject_kind <> 'template' OR document_id IS NULL);

ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_document_subject_projection_check;
ALTER TABLE public.approval_routes
    ADD CONSTRAINT approval_routes_document_subject_projection_check
    CHECK (subject_kind <> 'document' OR profile_code IS NOT NULL);

ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_template_subject_projection_check;
ALTER TABLE public.approval_routes
    ADD CONSTRAINT approval_routes_template_subject_projection_check
    CHECK (subject_kind <> 'template' OR profile_code IS NULL);

-- 4. Subject-scoped submit-idempotency index (covers both subjects; the
--    legacy approval_instances_document_id_idempotency_key_key index stays
--    for document-path back-compat and is not touched) --------------------

DROP INDEX IF EXISTS public.ux_approval_instances_subject_idempotency;
CREATE UNIQUE INDEX IF NOT EXISTS ux_approval_instances_subject_idempotency
    ON public.approval_instances (tenant_id, subject_kind, subject_key, idempotency_key)
    WHERE (idempotency_key IS NOT NULL);

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0297', 'P3.S2b-0 relax legacy document-only NOT NULL for template approval subjects (M3 kernel extraction): approval_instances.document_id and approval_routes.profile_code go NULL-able; both existing FKs (approval_instances_document_id_fkey, approval_routes_document_profile_fk) are KEPT as-is and remain NULL-tolerant under MATCH SIMPLE, so document-row referential integrity is unchanged; projection CHECK constraints (approval_instances/_routes _document_subject_projection_check + _template_subject_projection_check) enforce subject_kind vs legacy-column presence as the new DB-level invariant; subject-scoped partial unique index ux_approval_instances_subject_idempotency (tenant_id, subject_kind, subject_key, idempotency_key) added alongside the kept document_id-keyed approval_instances_document_id_idempotency_key_key so template submissions get real dedup once document_id is NULL (Go submit-path cutover to use it is out of scope for this slice, tracked P3.S2b-1).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
