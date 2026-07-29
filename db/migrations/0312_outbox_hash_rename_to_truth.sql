-- 0312_outbox_hash_rename_to_truth.sql
-- Etapa 5 §5.1 (F-QA4-10) — the staging dispatch outboxes both called their
-- payload column `content_hash`, but the two tables carry DIFFERENT hashes.
-- The name was a misnomer on both, and a misnomer on an integrity column is a
-- defect: it invites a reader to compare two values that are not comparable.
--
-- ============================================================================
-- VERIFIED SEMANTICS (per-table, at the enqueue site)
-- ============================================================================
--   metaldocs.materialize_dispatch_outbox.content_hash
--     <- FreezeService.Pin / RepairMaterialization pass the resolved-placeholder
--        VALUES hash (domain.ComputeValuesHash, freeze_service.go:211-223 and
--        the documents.values_hash column read at :261-266) into
--        dispatchjobs.Enqueuer.EnqueueMaterializeTx (enqueuer.go:97).
--     => renamed to values_hash, matching documents.values_hash.
--
--   metaldocs.pdf_dispatch_outbox.content_hash
--     <- MaterializeJobRunner passes the MATERIALIZED FROZEN-DOCX hash
--        (the renderer's fanout response hash, also written to
--        documents.content_hash by WriteFinalDocx) into
--        dispatchjobs.Enqueuer.EnqueuePDFTx
--        (materialize_job_runner.go:141-151, snapshot_repository.go:157-165,
--        staging_outbox.go:98-114).
--     => renamed to frozen_docx_hash.
--
--   documents.content_hash, document_revisions.content_hash and
--   documents.content_hash_at_submit are DIFFERENT, correctly-named columns and
--   are deliberately untouched here.
--
-- ============================================================================
-- HARD CUTOVER (no dual naming, no compat alias)
-- ============================================================================
--   The River job args (dispatchjobs/args.go) serialize as JSON, so the renamed
--   Go fields change the on-the-wire arg key for in-flight jobs. This is a
--   dev-only stack and api+worker+jobs deploy together (single compose roll), so
--   the cutover is atomic in practice. A job enqueued by an OLD binary and
--   picked up by a NEW one decodes its hash field as empty; the affected kinds
--   (`pdf_dispatch`, `materialize_dispatch`) are re-driven from the outbox row,
--   which is the durable record. No production data carries the old arg shape.
--
--   RENAME COLUMN is atomic and preserves data, indexes, constraints and RLS
--   policies: no index or constraint in the baseline references either column by
--   name (only status/tenant_id/revision_id/next_retry_at do), so nothing else
--   needs rewriting.

BEGIN;

ALTER TABLE metaldocs.materialize_dispatch_outbox
    RENAME COLUMN content_hash TO values_hash;

ALTER TABLE metaldocs.pdf_dispatch_outbox
    RENAME COLUMN content_hash TO frozen_docx_hash;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0312', 'Etapa 5 §5.1 (F-QA4-10): rename each staging dispatch outbox hash column to its verified payload — metaldocs.materialize_dispatch_outbox.content_hash -> values_hash (carries the resolved-placeholder values hash written by FreezeService.Pin) and metaldocs.pdf_dispatch_outbox.content_hash -> frozen_docx_hash (carries the materialized frozen-docx hash produced by the renderer fanout). Hard cutover, no compat alias: the paired River job args JSON field names are renamed with them, and api+worker+jobs deploy together. documents.content_hash, documents.content_hash_at_submit and document_revisions.content_hash are different, correctly-named columns and are untouched.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
