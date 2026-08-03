-- 0309_pdf_dispatch_outbox_final_docx_key.sql
-- ROADMAP unit QR-C -- F-QA2-2 fix: PDF dispatch events never carried the
-- renderer-produced final_docx_s3_key, so every docgen_v2_pdf event dead-lettered
-- with "pdf job runner: missing final_docx_s3_key" (payload key was "").
--
-- ============================================================================
-- WHY (event-contract snapshot, not a derived value)
-- ============================================================================
--   The staging outbox row is the event-contract snapshot for its dispatch:
--   content_hash (bytea NOT NULL, baseline 0001:1363) already lives here for
--   exactly this reason. The PDFConvertPayload consumer
--   (internal/platform/worker/pdf_job_runner.go:113-118) REQUIRES
--   FinalDocxS3Key -- the renderer-produced key, which the worker must NEVER
--   guess -- and fail-closes without it. The producer path had no column to
--   carry it from the enqueuing tx to the dispatched event, so the field was
--   always empty on the wire. This adds the missing snapshot column.
--
-- ============================================================================
-- EXPAND-ONLY
-- ============================================================================
--   Nullable, no backfill. The two already-dead-lettered rows
--   (f6f4712a.../83db8465...) keep final_docx_s3_key NULL -- they are not
--   re-dispatched (a fresh journey is re-driven post-merge). Every NEW row is
--   written with a non-empty key: the producers read/hold it in-tx and the
--   dispatchjobs.Enqueuer rejects an empty key (fail-closed). Tightening this
--   column to NOT NULL is a follow-up contract migration once the dead rows are
--   purged by retention; it is intentionally NOT done here (expand-only).
--
--   Only pdf_dispatch_outbox gains the column. materialize_dispatch_outbox has
--   no docx key at its enqueue time (the docx is produced by the materialize
--   fanout the event triggers) and is untouched.

BEGIN;

ALTER TABLE metaldocs.pdf_dispatch_outbox
    ADD COLUMN final_docx_s3_key text;

COMMIT;
