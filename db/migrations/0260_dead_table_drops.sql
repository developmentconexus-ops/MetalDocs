-- 0260_dead_table_drops.sql
-- DB-02 (P1) / DB-04 (P2) / DB-05 (P2): drop tables confirmed dead by grep
-- census against the CURRENT canonical baseline (db/baseline/0001_current_schema.sql,
-- last regenerated 2026-06-30 by scripts/export-schema-baseline.ps1) — see
-- wiki/reviews/grade-a-simplification-report-2026-07-01.md.
--
-- IMPORTANT premise correction (found during investigation, not in the original
-- ticket): the ticket's evidence pointer (create_document_snapshot_integration_
-- test.go:24-25, "metaldocs.documents is a dead legacy duplicate") is accurate,
-- but metaldocs.documents itself was ALREADY DROPPED by archived migration 0240
-- (db/reference-data/0001_product_reference_data.sql ledger row '0240', applied
-- June 2026) along with its 8 satellites and metaldocs.template_audit_log.
-- Zero-hit grep against db/baseline/0001_current_schema.sql confirms this table
-- and metaldocs.controlled_documents (which never existed — the real controlled-
-- documents table is public.controlled_documents, owned by the controlleddocuments
-- module) are both absent from the live schema today. DB-02's DROP is therefore a
-- no-op safety net (IF EXISTS, harmless); DB-02's real remaining work is repairing
-- the dead test scaffolding still hardcoding `metaldocs.documents` in literal SQL
-- (done separately in this commit: tests/integration/scenarios/concurrency_test.go,
-- obsolete_cascade_test.go, outbox_same_tx_test.go, schema_lockdown_test.go).
--
-- A STALE snapshot at migrations_baseline/0001_baseline_2026_05.sql (dated
-- 2026-05-14, predates 0236/0238/0240) still shows metaldocs.documents alive —
-- that file is not the canonical baseline (db/baseline/0001_current_schema.sql
-- is, per scripts/export-schema-baseline.ps1 and 0249's own dual-file update
-- convention) and was not used to justify anything in this migration.
--
-- DB-04 sub-items (a) subject_code and (b) editable_zones are ALREADY CLOSED:
--   (a) metaldocs.documents.subject_code + its index + fk_documents_subject_code
--       were dropped by archived 0238 (which then became moot when 0240 dropped
--       the whole table); zero hits in current baseline. No action needed.
--   (b) templates_template_version.editable_zones was dropped by archived 0196
--       (db/migrations/README lineage: 0196_drop_templates_editable_zones.sql);
--       zero hits in current baseline. No action needed.
-- Only DB-04(c) public.template_audit_log is live-dead and dropped below:
-- created by archived 0108 (docx_v2 era), RLS-hardened by 0237 (June 2026,
-- deliberate tenant-isolation investment on a table nobody reads/writes), but
-- zero Go code references anywhere in the tree (grep-confirmed: no INSERT/SELECT
-- call site). No inbound FK from any other table. Distinct from the already-gone
-- metaldocs.template_audit_log (dropped by 0240) — this is the public.* copy.
--
-- DB-05 MDDM shadow-table residue — confirmed refless by grep:
--   - metaldocs.document_template_versions_mddm: zero Go/TS/SQL code references
--     outside db/baseline DDL. No inbound FK. Has its own trg_template_immutable
--     trigger, dropped automatically with the table.
--   - metaldocs.template_drafts: zero Go/TS/SQL code references outside db/baseline
--     DDL. Outbound FK to metaldocs.document_profiles(code) only (no inbound FK);
--     DROP TABLE handles the outbound FK automatically.
--   NOT dropped: metaldocs.document_template_versions (the ADR 0032/0240
--   "MDDM cluster" name is superficially similar but this is a DIFFERENT, LIVE
--   table — apps/api/cmd/metaldocs-e2e-seed/main.go:123 SELECTs it directly to
--   gate a real INSERT into metaldocs.document_profile_template_defaults. Dropping
--   it would break the e2e-seed binary. Left untouched.
--
-- Drop order: no cross-table FK ordering issues among the three targets (each is
-- either already absent or a childless/single-outbound-FK leaf), so a flat list
-- is safe. IF EXISTS makes this idempotent and safe to re-run.

BEGIN;

-- DB-02: safety-net no-op — already dropped by archived 0240. Kept here so the
-- migration is self-documenting and re-runnable against any environment that
-- somehow missed 0240.
DROP TABLE IF EXISTS metaldocs.documents CASCADE;

-- DB-04(c): public.template_audit_log — zero Go code references, no inbound FK.
DROP TABLE IF EXISTS public.template_audit_log CASCADE;

-- DB-05: MDDM shadow-table residue — zero code references, confirmed above.
DROP TABLE IF EXISTS metaldocs.document_template_versions_mddm CASCADE;
DROP TABLE IF EXISTS metaldocs.template_drafts CASCADE;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0260', 'DB-02/DB-04(c)/DB-05: drop dead tables — metaldocs.documents (no-op safety net, already gone via archived 0240), public.template_audit_log (zero code refs, RLS-hardened but unread/unwritten), metaldocs.document_template_versions_mddm + metaldocs.template_drafts (MDDM shadow residue, zero code refs); metaldocs.document_template_versions explicitly NOT dropped (live: metaldocs-e2e-seed reads it); DB-04(a) subject_code and DB-04(b) editable_zones already closed by archived 0238/0196, no action')
ON CONFLICT (version) DO NOTHING;

COMMIT;
