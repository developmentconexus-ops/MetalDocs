-- 0276_document_review_surfaced_marker.sql
-- M6 F6.2 T4 (global-maximum-remediation, milestone-6-eqms-review-reason):
-- the River periodic review-due surfacer (jobs module) needs an idempotent
-- marker on public.documents to record "this due-for-review document has
-- already been surfaced for its current review_due_at cycle" — so a second
-- run within the same cycle is a no-op (contract §4.2/§4.3), while advancing
-- review_due_at (mark-reviewed) makes the row eligible to re-surface.
--
-- Plain nullable column add. Expand-only; no backfill; does not touch the
-- documents/UPDATE tripwire function (0275 remains the latest tripwire
-- migration) since this column carries no new capability requirement of its
-- own — the surfacer's UPDATE runs under the existing scheduler bypass path.

BEGIN;

ALTER TABLE public.documents
    ADD COLUMN IF NOT EXISTS review_surfaced_at timestamptz NULL;

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0276', 'M6 F6.2 T4: add public.documents.review_surfaced_at (nullable timestamptz) -- idempotency marker for the River periodic review-due surfacer (jobs module). A run''s MarkSurfaced UPDATE sets it to the run timestamp for docs newly due; the guard review_surfaced_at IS NULL OR review_surfaced_at < review_due_at makes a same-cycle rerun a no-op and a later review_due_at advance (mark-reviewed) re-eligible. Expand-only, no backfill, no tripwire function change.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
