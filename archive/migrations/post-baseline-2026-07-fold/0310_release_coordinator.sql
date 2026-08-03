-- 0310_release_coordinator.sql
-- ADR 0085 Stage A — release coordinator substrate: approval-driven publication.
--
-- Forward-only, expand-first. Four concerns, in dependency order:
--
--  (1) public.documents.planned_effective_from  — the PLAN half of the
--      planned/actual split that ADR 0085 introduces (amending ADR 0069's
--      "effective_from is the single effective/target date" ruling).
--      `effective_from` keeps its column name and becomes the ACTUAL release
--      timestamp, written by the winning release transaction and NEVER cleared
--      (this is what closes F-QA4-13: the M6 review predicate
--      documents/infrastructure/review_due_reader.go:40 requires
--      effective_from IS NOT NULL).
--
--  (2) Data move (ADR 0085 "In-flight disposition" step 2): existing
--      status='scheduled' rows carry their PLANNED date in effective_from
--      (written by approval/application/publish_service.go SchedulePublish).
--      Move it to planned_effective_from and null the actual. Dev-stage
--      dataset only (migration-policy.md dev-seed separation) — no production
--      tenants exist.
--      Symmetrically, legacy `published` rows have NO effective_from at all
--      (publish_service.go PublishApproved never wrote one; scheduler_service.go
--      actively CLEARED it) — they are backfilled to a safe in-window instant so
--      the new invariant CHECK can be added without a violation.
--
--  (3) The two lifecycle invariants ADR 0085 §"DB invariants" demands, as
--      NAMED CHECK constraints in the ck_* convention of migration 0274:
--        published  => effective_from IS NOT NULL
--        scheduled  => planned_effective_from IS NOT NULL AND effective_from IS NULL
--      Both are written status-scoped (`status <> 'x' OR ...`) so every other
--      status stays untouched and legacy rows keep passing.
--
--  (3b) The SUPERSESSION-HEAD invariant, also from ADR 0085 §"DB invariants":
--      at most ONE published row per controlled document. The baseline's
--      ux_documents_cd_active only covers draft/under_review/approved/rejected/
--      scheduled, so nothing stopped the still-live legacy publish path
--      (approval/application/publish_service.go PublishApproved, which does NOT
--      supersede the existing head) from producing two published rows for one
--      CD — after which the coordinator's re-decision silently no-ops. This is
--      the release coordinator's core invariant, so it is enforced HERE (Stage
--      A), not deferred to the Stage B retirement of that path. Idempotent
--      pre-repair first (dev-stage dataset), then the partial unique index.
--
--  (4) public.release_generations — the ONE durable table the coordinator
--      keys everything on. Shape decision (delegated by the Stage A brief):
--      a SINGLE table carrying the generation identity + fact timestamps +
--      coordinator state, rather than separate fact tables. Rationale: the
--      release predicate is a conjunction over facts of the SAME identity, the
--      release CAS needs one row to lock, and the readiness-hold projection is
--      one reason per generation — three separate tables would need a 3-way
--      join and a synthetic "which row is the projection" rule to answer any
--      of them. Fact production stays fail-closed either way (a fact is a
--      non-null timestamp written in the producing transaction).
--
--  (5) Generation-aware materialization dedupe: metaldocs.materialize_dispatch_outbox
--      and metaldocs.pdf_dispatch_outbox gain release_generation_id and their
--      revision-only UNIQUE (tenant_id, revision_id) is replaced by a
--      generation-aware unique index. ADR 0085: "Materialization
--      idempotency/dedupe keys widen from revision-only to the generation key."
--      Without this, the re-enqueue the backfill (step 3 of the in-flight
--      disposition) and every re-approval depend on is silently swallowed by
--      the old ON CONFLICT DO NOTHING.
--      The index uses COALESCE(release_generation_id, <nil uuid>) rather than
--      UNIQUE NULLS NOT DISTINCT so it is version-portable and so legacy
--      (generation-less) rows still dedupe on revision alone.
--
-- RLS: public.release_generations is a tenant table and gets the verbatim
-- ENABLE + FORCE + tenant_isolation idiom (GUC metaldocs.tenant_id, null-GUC
-- escape-hatch branch preserved per the ADR 0027 amendment) copied from
-- 0285_approval_signoffs_rls.sql. It is registered on the approval module's
-- TenantDataPort in the same change (internal/modules/approval/infrastructure/
-- tenant_data_port.go) so tenant export/erasure covers it.
--
-- No tripwire is added on release_generations: it is written only by the
-- approval kernel's own in-transaction fact producers and by the coordinator
-- job (background-bypass root), never by a raw HTTP path — the same posture as
-- metaldocs.materialize_dispatch_outbox.

BEGIN;

-- ── (1) planned_effective_from ───────────────────────────────────────────────

ALTER TABLE public.documents
    ADD COLUMN IF NOT EXISTS planned_effective_from timestamp with time zone;

COMMENT ON COLUMN public.documents.planned_effective_from IS
    'ADR 0085: immutable PLAN date declared in the publication plan at submission (NULL = effective on release). Never mutated by release; the actual release timestamp is effective_from.';

COMMENT ON COLUMN public.documents.effective_from IS
    'ADR 0085: ACTUAL effective timestamp, written by the winning release transaction at coordinator-evaluation time and never cleared. NULL while the document is not yet released (scheduled rows carry their plan in planned_effective_from).';

-- ── (2) data move + legacy backfill ─────────────────────────────────────────

-- The capability tripwire (trg_require_cap_asserted, BEFORE INSERT OR UPDATE on
-- public.documents) fires on every documents UPDATE below — the data move, the
-- legacy backfill and the (3b) pre-repair — and raises P0001
-- ErrCapabilityNotAsserted unless metaldocs.asserted_caps names one of
-- {document.edit, document.obsolete, membership.manage, document.review}
-- (latest arm: db/migrations/0301_tripwire_template_review_retired.sql:93-104).
-- The migration runner (internal/platform/migrate/migrate.go) never sets that
-- GUC, so any pre-existing matching row fails the migration. Assert tx-locally
-- — same defect class and same remedy as
-- db/migrations/0267_templates_template_updated_at.sql:49 (proven 2026-07-03,
-- SQLSTATE P0001). set_config(..., true) is transaction-local and this file is
-- one explicit BEGIN/COMMIT, so it covers every documents write that follows.
SELECT set_config('metaldocs.asserted_caps', '[{"cap":"document.edit"}]', true);

-- scheduled: plan lived in effective_from; move it, null the actual.
UPDATE public.documents
   SET planned_effective_from = effective_from,
       effective_from         = NULL
 WHERE status = 'scheduled'
   AND effective_from IS NOT NULL
   AND planned_effective_from IS NULL;

-- published with no actual date: choose an instant that is (a) not in the
-- future, (b) not after any review_due_at (ck_documents_review_due_sane), and
-- (c) strictly before any effective_to (ck_documents_effective_window). LEAST
-- over the candidates gives exactly that; 'infinity' neutralises absent bounds.
UPDATE public.documents
   SET effective_from = LEAST(
           COALESCE(created_at, now()),
           COALESCE(review_due_at, 'infinity'::timestamptz),
           COALESCE(effective_to - interval '1 microsecond', 'infinity'::timestamptz)
       )
 WHERE status = 'published'
   AND effective_from IS NULL;

-- ── (3) invariant CHECKs ────────────────────────────────────────────────────

ALTER TABLE public.documents
    ADD CONSTRAINT ck_documents_published_effective_from
    CHECK (status <> 'published' OR effective_from IS NOT NULL);

ALTER TABLE public.documents
    ADD CONSTRAINT ck_documents_scheduled_planned_effective_from
    CHECK (status <> 'scheduled'
           OR (planned_effective_from IS NOT NULL AND effective_from IS NULL));

-- ── (3b) one-published-head invariant ───────────────────────────────────────

-- Pre-repair. Keep the highest revision_number as the head and demote every
-- other published row of the same controlled document to 'superseded' (a legal
-- transition; published -> superseded is what the coordinator itself performs).
-- revision_version is bumped on the demoted rows so any in-flight OCC holder
-- loses its CAS instead of writing against a row whose status changed
-- underneath it. Re-running is a no-op: once at most one published row per
-- (tenant, controlled document) remains, the ranked subquery selects nothing.
WITH ranked AS (
    SELECT id,
           row_number() OVER (
               PARTITION BY tenant_id, controlled_document_id
               ORDER BY revision_number DESC, created_at DESC, id
           ) AS rn
      FROM public.documents
     WHERE status = 'published'
       AND controlled_document_id IS NOT NULL
)
UPDATE public.documents d
   SET status           = 'superseded',
       revision_version = d.revision_version + 1
  FROM ranked
 WHERE ranked.id = d.id
   AND ranked.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS ux_documents_published_head
    ON public.documents (tenant_id, controlled_document_id)
 WHERE status = 'published' AND controlled_document_id IS NOT NULL;

-- ── (4) release_generations ─────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS public.release_generations (
    id                     uuid DEFAULT gen_random_uuid() NOT NULL,
    -- generation_seq gives a total, monotonic order per (tenant, document) so
    -- the "G is still D's freeze head" conjunct of the release predicate is a
    -- plain NOT EXISTS (... seq > G.seq) rather than a timestamp comparison.
    generation_seq         bigserial NOT NULL,
    tenant_id              uuid NOT NULL,
    subject_kind           text NOT NULL DEFAULT 'document',
    document_id            uuid NOT NULL,
    approval_instance_id   uuid NOT NULL,
    revision_id            uuid NOT NULL,
    revision_version       integer NOT NULL,
    frozen_content_hash    text NOT NULL,

    -- approval fact (written in the terminal-approval transaction; fail-closed)
    approval_fact_at       timestamp with time zone,
    final_approver_id      text,
    submitted_by           text,

    -- artifact fact (written in the artifact-persistence transaction)
    final_docx_s3_key      text,
    final_pdf_s3_key       text,
    artifact_fact_at       timestamp with time zone,

    -- coordinator state / readiness-hold projection
    hold_reason            text,
    hold_detail            text,
    released_at            timestamp with time zone,
    last_evaluated_at      timestamp with time zone,

    created_at             timestamp with time zone DEFAULT now() NOT NULL,
    updated_at             timestamp with time zone DEFAULT now() NOT NULL,

    CONSTRAINT release_generations_pkey PRIMARY KEY (id),

    -- The coordinator is document-only: template subjects keep their existing
    -- skip and never produce release facts (ADR 0085 §"Release generation").
    CONSTRAINT ck_release_generations_subject_kind
        CHECK (subject_kind = 'document'),

    CONSTRAINT ck_release_generations_frozen_hash
        CHECK (frozen_content_hash ~ '^[0-9a-f]{64}$'),

    CONSTRAINT ck_release_generations_revision_version
        CHECK (revision_version >= 0),

    -- The readiness-hold reason vocabulary, verbatim from ADR 0085
    -- §"Observability, attribution, reconciliation".
    CONSTRAINT ck_release_generations_hold_reason
        CHECK (hold_reason IS NULL OR hold_reason IN (
            'awaiting_approval_fact',
            'materializing',
            'awaiting_effective_date',
            'supersede_conflict',
            'plan_invalid',
            'failed'
        )),

    -- A released generation must carry both facts — the predicate can never
    -- have held otherwise. DB-enforced so a coding regression in the
    -- coordinator surfaces as a constraint violation, not a silent release.
    CONSTRAINT ck_release_generations_released_implies_facts
        CHECK (released_at IS NULL
               OR (approval_fact_at IS NOT NULL AND artifact_fact_at IS NOT NULL)),

    -- The artifact fact means the FULL set exists (ADR 0085 "artifact-ready(G)"
    -- = final DOCX **and** final PDF). Column presence is never readiness, but
    -- the fact timestamp may not be written without both pointers.
    CONSTRAINT ck_release_generations_artifact_fact_full_set
        CHECK (artifact_fact_at IS NULL
               OR (final_docx_s3_key IS NOT NULL AND final_pdf_s3_key IS NOT NULL)),

    -- The generation identity from ADR 0085 §"Release generation".
    CONSTRAINT ux_release_generations_identity UNIQUE
        (tenant_id, subject_kind, document_id, approval_instance_id,
         revision_id, revision_version, frozen_content_hash)
);

-- Head lookup ("is G still D's freeze head") and the per-document projection.
CREATE INDEX IF NOT EXISTS ix_release_generations_document
    ON public.release_generations (tenant_id, document_id, generation_seq DESC);

-- Reconciliation sweep: generations still on hold.
CREATE INDEX IF NOT EXISTS ix_release_generations_open
    ON public.release_generations (tenant_id, last_evaluated_at)
 WHERE released_at IS NULL;

ALTER TABLE public.release_generations ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY public.release_generations FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON public.release_generations;
CREATE POLICY tenant_isolation ON public.release_generations
  USING (
    (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
    OR (tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid)
  );

-- ── (5) generation-aware materialization dedupe ─────────────────────────────

ALTER TABLE metaldocs.materialize_dispatch_outbox
    ADD COLUMN IF NOT EXISTS release_generation_id uuid;

ALTER TABLE metaldocs.pdf_dispatch_outbox
    ADD COLUMN IF NOT EXISTS release_generation_id uuid;

ALTER TABLE metaldocs.materialize_dispatch_outbox
    DROP CONSTRAINT IF EXISTS ux_materialize_dispatch_outbox_revision;

ALTER TABLE metaldocs.pdf_dispatch_outbox
    DROP CONSTRAINT IF EXISTS ux_pdf_dispatch_outbox_revision;

CREATE UNIQUE INDEX IF NOT EXISTS ux_materialize_dispatch_outbox_generation
    ON metaldocs.materialize_dispatch_outbox
       (tenant_id, revision_id,
        COALESCE(release_generation_id, '00000000-0000-0000-0000-000000000000'::uuid));

CREATE UNIQUE INDEX IF NOT EXISTS ux_pdf_dispatch_outbox_generation
    ON metaldocs.pdf_dispatch_outbox
       (tenant_id, revision_id,
        COALESCE(release_generation_id, '00000000-0000-0000-0000-000000000000'::uuid));

-- ── schema_migrations ledger ────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0310', 'ADR 0085 Stage A: add public.documents.planned_effective_from (plan) and re-home effective_from as the ACTUAL release timestamp; move scheduled rows plan -> planned_effective_from and backfill legacy published rows; add ck_documents_published_effective_from and ck_documents_scheduled_planned_effective_from; demote extra published rows per controlled document and add ux_documents_published_head (one published head per (tenant_id, controlled_document_id), the DB enforcement of ADR 0085 supersession-head); create public.release_generations (single-table generation identity + approval/artifact facts + readiness-hold state, FORCE RLS + tenant_isolation); widen materialize/pdf dispatch outbox dedupe from (tenant_id, revision_id) to a generation-aware unique index.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
