-- 0272_documents_remove_rejected.sql
-- Removes the dead 'rejected' public.documents.status literal.
--
-- 'rejected' is confirmed dead: no app service ever writes it. The reject
-- path collapses under_review -> draft directly (never through 'rejected').
-- Task A (this program) shipped the unified documents/domain transition
-- function CanTransitionDocumentStatus(cur,next) with NO rejected arcs
-- (internal/modules/documents/domain/state.go). This migration removes
-- 'rejected' from the DB side (CHECK constraint + trigger arcs) so the DB
-- trigger and the app function agree (parity) — see the accompanying Go
-- unit test internal/modules/documents/domain/state_parity_test.go.
--
-- Pre-check: a DO block RAISEs a clear, named error BEFORE the constraint
-- swap if any existing row already has status='rejected', so a bad deploy
-- fails loudly instead of corrupting data or leaving the DB half-migrated.
-- Same idiom as migration 0265_documents_status_check_tighten.sql.
--
-- Net: live set = draft, under_review, approved, scheduled, published,
-- superseded, obsolete, archived (8 values). Dropped: rejected (1 value).
-- 'archived' explicitly KEPT (ADR 0010, unrelated to this removal).
--
-- Trigger rewrite: public.enforce_document_transition() is replaced with the
-- identical body from db/baseline/0001_current_schema.sql (lines ~549-585)
-- EXCEPT for exactly two edits:
--   1. (OLD.status = 'under_review' AND NEW.status IN ('approved','rejected'))
--      becomes
--      (OLD.status = 'under_review' AND NEW.status = 'approved')
--   2. The line (OLD.status = 'rejected' AND NEW.status = 'draft') is deleted.
-- The under_review->draft cancel-GUC special case, SET search_path, DECLARE,
-- and all other arcs are preserved verbatim.
--
-- Rollback note: re-add 'rejected' to the CHECK constraint and restore the
-- two trigger arcs (see migration 0265 for the historical 9-value list).
-- Safe to roll back at any time before new rows accumulate outside the
-- tightened set.

BEGIN;

DO $$
DECLARE
  v_bad_count bigint;
BEGIN
  SELECT count(*)
    INTO v_bad_count
    FROM public.documents
   WHERE status = 'rejected';

  IF v_bad_count > 0 THEN
    RAISE EXCEPTION 'migration 0272 aborted: % row(s) in public.documents have status=''rejected'', which this migration removes from the live set. Resolve these rows (backfill/correct status to ''draft'' or another live status) before re-running this migration.',
      v_bad_count
      USING ERRCODE = 'check_violation';
  END IF;
END $$;

ALTER TABLE public.documents
    DROP CONSTRAINT documents_status_check;

ALTER TABLE public.documents
    ADD CONSTRAINT documents_status_check
    CHECK ((status = ANY (ARRAY[
      'draft'::text, 'under_review'::text, 'approved'::text,
      'scheduled'::text, 'published'::text, 'superseded'::text,
      'obsolete'::text, 'archived'::text
    ])));

CREATE OR REPLACE FUNCTION public.enforce_document_transition() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  cancel_instance_id TEXT;
BEGIN
  IF OLD.status IS DISTINCT FROM NEW.status THEN
    -- Check cancel path: under_review→draft allowed if cancel GUC is set.
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
      (OLD.status = 'under_review' AND NEW.status =  'approved') OR
      (OLD.status = 'approved'     AND NEW.status IN ('published','scheduled','draft')) OR
      (OLD.status = 'scheduled'    AND NEW.status IN ('published','draft')) OR
      (OLD.status = 'published'    AND NEW.status IN ('superseded','obsolete')) OR
      (OLD.status = 'superseded'   AND NEW.status =  'obsolete')
      -- Note: compat window (draft→finalized, finalized→archived) removed by 0142_disable_legacy_compat.sql
      -- Note: 'rejected' arcs (under_review→rejected, rejected→draft) removed by 0272_documents_remove_rejected.sql
    ) THEN
      RAISE EXCEPTION 'illegal status transition % -> %', OLD.status, NEW.status
        USING ERRCODE = 'check_violation';
    END IF;
  END IF;
  RETURN NEW;
END;
$$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0272', 'Remove dead ''rejected'' document status: tighten documents_status_check to 8 live values and rewrite enforce_document_transition() to drop the under_review->rejected and rejected->draft arcs, restoring app<->DB parity with documents/domain.CanTransitionDocumentStatus.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
