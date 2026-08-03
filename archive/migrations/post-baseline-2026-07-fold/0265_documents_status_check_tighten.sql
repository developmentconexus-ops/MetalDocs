-- 0265_documents_status_check_tighten.sql
-- Grade-A simplification: re-tighten public.documents_status_check, which has
-- never been re-tightened since the Spec-2 backfill (baseline
-- db/baseline/0001_current_schema.sql:1903 still allows the dead 'finalized'
-- literal from the pre-migration-0142 lifecycle).
--
-- Live-status-set derivation (code truth, both sides checked):
--   - Go: internal/modules/documents/domain/model.go DocumentStatus constants
--     (DocStatusDraft, DocStatusUnderReview, DocStatusApproved,
--     DocStatusRejected, DocStatusScheduled, DocStatusPublished,
--     DocStatusSuperseded, DocStatusObsolete, DocStatusArchived) — 9 values,
--     all actively referenced (repository.go, service.go, handler.go,
--     state.go). This is the enum backing public.documents.status.
--   - Generated wire enum: internal/modules/documents/api/api.gen.go
--     DocumentSummaryStatus — 8 values (mirrors the above minus 'archived',
--     which is not part of the list-summary projection but IS a valid detail
--     status per DocStatusArchived above and handler.go:1564 / repository.go:619).
--   - FE: frontend/apps/web/src/features/documents/lib/parseDocumentStatus.ts
--     CANONICAL_STATUSES has 10 entries — the 9 Go values above PLUS 'frozen'.
--     'frozen' traced to frontend/apps/web/src/components/ui/StatusPill.tsx
--     only (a label/pill config for a status that has no Go producer
--     anywhere — grep-confirmed zero hits for a "frozen" DocumentStatus/
--     DocState constant or literal write to documents.status in
--     internal/modules/documents). Go's ValuesFrozenAt / values_frozen_at is
--     an unrelated timestamp column (revision content-freeze marker, not a
--     document lifecycle status). 'frozen' is FE-only speculative/legacy UI
--     vocabulary and is NOT added to the DB constraint.
--   - 'finalized': confirmed dead. parseDocumentStatus.ts explicitly documents
--     it as a "pre-cutover alias" removed by migration 0142 (draft->finalized
--     transition retired); internal/modules/documents/approval/domain/state.go
--     StateFromString rejects it at the parse boundary via
--     ErrLegacyStateRejected. No Go DocStatus constant, no FE canonical
--     status, no write call site anywhere in internal/modules/documents.
--   - 'archived': KEPT. Explicitly live again per task instruction and
--     confirmed by code: DocStatusArchived is referenced in
--     internal/modules/documents/domain/state.go (CanTransitionDocument),
--     internal/modules/documents/delivery/http/handler.go:1564, and
--     internal/modules/documents/repository/repository.go:619. Note this is
--     a DIFFERENT concept from approval/domain/state.go's DocState, which
--     rejects "archived" as legacy in the separate 8-state Spec-2 approval
--     graph (a different column/state machine, not public.documents.status).
--
-- Net: live set = draft, under_review, approved, rejected, scheduled,
-- published, superseded, obsolete, archived (9 values). Dropped: finalized
-- (1 value). Not added: frozen (FE-only, no backend producer).
--
-- Pre-check: a DO block RAISEs a clear, named error BEFORE the constraint
-- swap if any existing row already violates the tightened set, so a bad
-- deploy fails loudly (migration aborts, transaction rolls back) instead of
-- either corrupting data or leaving the DB in a half-migrated state.
--
-- Safety class: the pre-check SELECT is a lightweight index-friendly scan
-- (no lock beyond the migration's own transaction's implicit read lock).
-- DROP CONSTRAINT + ADD CONSTRAINT CHECK on public.documents both take
-- ACCESS EXCLUSIVE on public.documents for the duration of the DDL. The new
-- CHECK is added WITHOUT NOT VALID (i.e. validated immediately) because the
-- pre-check DO block already proves no violating rows exist, so the
-- validation scan inside ADD CONSTRAINT is redundant-but-safe (still holds
-- ACCESS EXCLUSIVE while it runs; documents is an operationally sized table,
-- not reference data, so this is the one migration in this batch expected to
-- hold its lock for longer than milliseconds — plan the deploy window
-- accordingly, same caution as any CHECK constraint validation on a live
-- transactional table).
--
-- Rollback note: re-add the constraint with 'finalized' included:
--   ALTER TABLE public.documents DROP CONSTRAINT documents_status_check;
--   ALTER TABLE public.documents ADD CONSTRAINT documents_status_check
--     CHECK (status = ANY (ARRAY['draft','finalized','archived','under_review',
--       'approved','rejected','scheduled','published','superseded','obsolete']));
-- Safe to roll back at any time — widening a CHECK constraint never
-- invalidates existing rows.

BEGIN;

DO $$
DECLARE
  v_bad_count bigint;
  v_sample text;
BEGIN
  SELECT count(*), string_agg(DISTINCT status, ', ')
    INTO v_bad_count, v_sample
    FROM public.documents
   WHERE status NOT IN (
     'draft', 'under_review', 'approved', 'rejected', 'scheduled',
     'published', 'superseded', 'obsolete', 'archived'
   );

  IF v_bad_count > 0 THEN
    RAISE EXCEPTION 'migration 0265 aborted: % row(s) in public.documents have a status value outside the tightened live set (offending distinct values: %). Resolve these rows (backfill/correct status) before re-running this migration.',
      v_bad_count, v_sample
      USING ERRCODE = 'check_violation';
  END IF;
END $$;

ALTER TABLE public.documents
    DROP CONSTRAINT documents_status_check;

ALTER TABLE public.documents
    ADD CONSTRAINT documents_status_check
    CHECK ((status = ANY (ARRAY[
      'draft'::text, 'under_review'::text, 'approved'::text, 'rejected'::text,
      'scheduled'::text, 'published'::text, 'superseded'::text,
      'obsolete'::text, 'archived'::text
    ])));

INSERT INTO public.schema_migrations (version, description)
VALUES ('0265', 'Grade-A simplification: tighten public.documents_status_check from 10 to 9 live values, dropping dead legacy literal ''finalized'' (retired by migration 0142, confirmed zero code producers on both Go and FE sides). ''archived'' explicitly kept (live). ''frozen'' (FE StatusPill-only, no Go producer) explicitly NOT added. Pre-check DO block RAISEs before the swap if any existing row would violate.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
