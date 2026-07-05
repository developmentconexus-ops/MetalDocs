-- 0274_document_review_and_reason.sql
-- M6 F6.2 + F6.3 (global-maximum-remediation, milestone-6-eqms-review-reason):
-- the shared forward migration for the eQMS periodic-review / expiry model and
-- the structured reason-for-change capture. Authored against the M6 validation
-- contract §2 (review/expiry column + CHECK table) and §5.1 (reason columns).
--
-- Expand-only, forward-only. Every new column is nullable; legacy rows keep
-- NULL (= "no review cycle set" / "no structured reason recorded"). No backfill
-- (intent on legacy rows cannot be reconstructed — mission §8 bounded defer).
--
-- New columns on public.documents:
--   review_due_at    timestamptz NULL -- when the doc is next due for periodic review
--   last_reviewed_at timestamptz NULL -- set by the mark-reviewed workflow (F6.2 §4.4)
--   reason_for_change text        NULL -- structured 21 CFR Part 11 change reason (F6.3 §5.1)
--   reason_category   text        NULL -- optional enum classifying the reason (F6.3 §5.1)
--
-- effective_from / effective_to ALREADY EXIST on public.documents
-- (db/baseline/0001_current_schema.sql:1868-1869, both nullable) and are NOT
-- added here. Per contract §0.1/§2 the review REUSES effective_from as the
-- effective date and WIRES the previously-unwritten effective_to as expiry —
-- this migration only adds the DB CHECK that guards the pair. Verified at
-- authoring time: the only pre-existing CHECKs on public.documents are the hash
-- length checks, documents_name_not_empty, documents_reconstruction_attempts_is_array,
-- and documents_status_check (baseline :1896-1904) — NONE covers the
-- effective_from/effective_to window, so ck_documents_effective_window is new,
-- not a duplicate.
--
-- All three CHECKs are NULL-tolerant (col IS NULL OR ...) so they never
-- invalidate the expand-only legacy rows:
--   ck_documents_effective_window  -- expiry strictly after the effective date
--   ck_documents_review_due_sane   -- review-due not before the effective date
--   ck_documents_reason_category   -- reason_category confined to the enum
--
-- Named CHECK constraints (ck_* convention) so the DB-CHECK integration proof
-- can assert on the constraint name
-- (internal/modules/documents/repository/document_review_columns_integration_test.go).

BEGIN;

ALTER TABLE public.documents
    ADD COLUMN IF NOT EXISTS review_due_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS last_reviewed_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS reason_for_change text,
    ADD COLUMN IF NOT EXISTS reason_category text;

-- Expiry (effective_to) must be strictly after the effective date
-- (effective_from). NULL-tolerant: either endpoint NULL = no window to check.
ALTER TABLE public.documents
    ADD CONSTRAINT ck_documents_effective_window
    CHECK (effective_to IS NULL OR effective_from IS NULL OR effective_to > effective_from);

-- A review cannot be due before the document is even effective. NULL-tolerant.
ALTER TABLE public.documents
    ADD CONSTRAINT ck_documents_review_due_sane
    CHECK (review_due_at IS NULL OR effective_from IS NULL OR review_due_at >= effective_from);

-- reason_category, when present, is confined to the F6.3 §5.1 enum. NULL-tolerant.
ALTER TABLE public.documents
    ADD CONSTRAINT ck_documents_reason_category
    CHECK (reason_category IS NULL OR reason_category IN ('content', 'corrective', 'regulatory', 'periodic_review', 'administrative'));

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0274', 'M6 F6.2/F6.3: add eQMS review/expiry + structured reason columns to public.documents (review_due_at, last_reviewed_at, reason_for_change, reason_category), all nullable/expand-only. Add NULL-tolerant CHECKs ck_documents_effective_window (effective_to > effective_from), ck_documents_review_due_sane (review_due_at >= effective_from) and ck_documents_reason_category (enum). Reuses existing effective_from/effective_to; no new duplicate column family; no backfill.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
