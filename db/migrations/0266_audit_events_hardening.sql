-- 0266_audit_events_hardening.sql
-- DB-09 (P3, BOUNDED): audit hardening. Implements ONLY the two bounded
-- sub-items from the task; everything else (partitioning, pg-cron retention,
-- ULID/UUIDv7 ids) is deferred with rationale in
-- wiki/modules/audit-tech-debt.md (see T-013 appended there).
--
-- (a) INSERT-only grant posture for the audit_events writer role.
--
-- Role inventory performed before writing this migration (grep across
-- archive/migrations/ + db/migrations/ for "metaldocs_app", "CREATE ROLE",
-- "GRANT.*audit_events"):
--   - There is NO distinct "audit writer" role. All four binaries
--     (metaldocs-api, metaldocs-worker, metaldocs-jobs, docx-renderer)
--     connect as the single application role metaldocs_app
--     (scripts/export-schema-baseline.ps1:16 `pg_dump -U metaldocs_app`
--     confirms this is the live connection identity).
--   - metaldocs_app's grant on metaldocs.audit_events has only ever been
--     INSERT (archive/migrations/0005_grant_workflow_audit_privileges.sql:2
--     `GRANT INSERT ON TABLE metaldocs.audit_events TO metaldocs_app`) then
--     widened to INSERT+SELECT for the read API
--     (archive/migrations/0193_audit_events_hash_chain.sql:110
--     `GRANT INSERT, SELECT ON TABLE metaldocs.audit_events TO
--     metaldocs_app`). No migration in this repo's history has ever granted
--     UPDATE or DELETE on audit_events to metaldocs_app. The application
--     role's runtime posture is therefore ALREADY insert/select-only —
--     T-004 (closed, tamper-evidence) already documents this as the
--     append-only enforcement mechanism ("Append-only is enforced at the
--     application role (metaldocs_app has only INSERT)").
--   - Constraint (not invented): per task instructions, since the app uses a
--     single role rather than a distinct audit-writer role, this migration
--     documents and durably hardens that constraint instead of inventing a
--     new role. A dedicated single-purpose "audit_writer" role would need
--     its own connection-pooling wiring across 4 binaries plus a SET ROLE
--     or dblink-style hop from metaldocs_app — out of scope for a P3
--     bounded finding.
--   - Hardening applied below: an explicit, idempotent REVOKE of
--     UPDATE/DELETE/TRUNCATE from metaldocs_app on audit_events. This adds
--     no new behavior (those privileges were never granted) but makes the
--     append-only posture self-documenting and future-migration-proof: if
--     any later migration accidentally widens metaldocs_app's default
--     privileges (e.g. via a blanket `GRANT ALL ON ALL TABLES IN SCHEMA`),
--     this statement is the canonical place a reviewer looks to confirm
--     audit_events' privilege posture is deliberate, matching the
--     precedent set by archive/migrations/0108_docx_v2_template_audit_log.sql
--     (`REVOKE UPDATE, DELETE ON template_audit_log FROM %I`). Skipped if
--     the role does not exist in the target environment (DO block
--     conditional on pg_roles), so this is safe to run in any environment
--     ordering.
--
-- (b) Payload size cap CHECK on the unbounded JSONB payload
--     (wiki/modules/audit-tech-debt.md T-010).
--
-- Cap chosen: 64 KiB (65536 bytes) on octet_length(payload::text).
-- Justification: current producers (internal/modules/iam/delivery/http/
-- admin_handler.go, apps/api/cmd/metaldocs-api/main.go
-- documentsAuditAdapter, internal/modules/templates repository AppendAudit)
-- all emit small structured maps (ids, status transitions, field diffs) —
-- observed payloads are well under 1 KB. 64 KiB gives ~64x headroom over
-- any known real payload shape while still bounding worst case: at
-- audit_events' current growth pattern (one row per regulated mutation,
-- not user-uploaded content) a 64 KiB ceiling caps a single pathological
-- event at roughly the size of a large JSON diff or a small document
-- metadata dump, without allowing someone to smuggle a multi-MB blob (e.g.
-- an accidentally-included document body) into the tamper-evident hash
-- chain, which would both bloat the table and make hash-chain
-- integrity-validation scans (internal/modules/jobs/audit_integrity_
-- validator/job.go) far more expensive per row. This is a defensive
-- ceiling, not a tuned business limit — if a legitimate future payload
-- shape needs more, that is a deliberate schema change, not silent growth.
--
-- Pre-check: a DO block RAISEs a clear, named error BEFORE the constraint
-- add if any existing row already exceeds the cap, so deploy fails loudly
-- instead of silently being unable to add the constraint (ADD CONSTRAINT
-- would itself fail with a generic check_violation naming only the first
-- offending row; this pre-check gives an operator-actionable count + max
-- size instead).
--
-- Safety class:
--   (a) REVOKE on a role that was never granted the revoked privileges is a
--       catalog-only, near-instant operation — no table lock beyond a brief
--       catalog update.
--   (b) The pre-check SELECT does a full scan of audit_events computing
--       octet_length(payload::text) per row — audit_events grows
--       monotonically (T-003, open) so this scan cost grows over time; at
--       today's table size this is expected to be a read-only, non-locking
--       scan (no lock beyond default read consistency). ADD CONSTRAINT ...
--       CHECK is added as NOT VALID + a separate VALIDATE CONSTRAINT
--       statement: NOT VALID itself only takes a brief ACCESS EXCLUSIVE to
--       record the constraint (no full-table check), and VALIDATE CONSTRAINT
--       normally only needs SHARE UPDATE EXCLUSIVE (concurrent reads/writes
--       proceed) when run as its own transaction. CAVEAT: this migration
--       follows this repo's one-BEGIN/COMMIT-per-file convention
--       (db/migrations/README.md), so both statements run inside the same
--       transaction as the REVOKE and the pre-check — the NOT VALID lock is
--       therefore held for the combined duration, not released between
--       steps. The NOT VALID/VALIDATE split still mirrors the precedent set
--       by audit_events_prev_hash_shape / audit_events_row_hash_shape
--       (baseline lines ~2502-2514, also added NOT VALID) and remains the
--       correct pattern if this statement is ever re-run standalone outside
--       this migration's transaction; it does not buy true zero-downtime
--       within this specific deploy. Table size at time of writing is small
--       enough that this is not expected to matter in practice, but treat
--       this migration with the same care as documents_status_check (0265)
--       — plan the deploy window, do not assume sub-millisecond lock time.
--
-- Rollback note:
--   (a) Re-run `GRANT UPDATE, DELETE, TRUNCATE ON TABLE metaldocs.audit_events
--       TO metaldocs_app;` if the revoke ever needs reversing (it should
--       not — this would reopen the tamper path T-004 closed).
--   (b) `ALTER TABLE metaldocs.audit_events DROP CONSTRAINT
--       audit_events_payload_size_cap;` — safe at any time, only relaxes.

BEGIN;

-- ── (a) Durably harden the already-insert/select-only posture ──────────────
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_app') THEN
    EXECUTE 'REVOKE UPDATE, DELETE, TRUNCATE ON TABLE metaldocs.audit_events FROM metaldocs_app';
  END IF;
END $$;

-- ── (b) Payload size cap ────────────────────────────────────────────────────
DO $$
DECLARE
  v_bad_count bigint;
  v_max_bytes bigint;
BEGIN
  SELECT count(*), max(octet_length(payload::text))
    INTO v_bad_count, v_max_bytes
    FROM metaldocs.audit_events
   WHERE octet_length(payload::text) > 65536;

  IF v_bad_count > 0 THEN
    RAISE EXCEPTION 'migration 0266 aborted: % row(s) in metaldocs.audit_events have a payload exceeding the 64KiB cap (largest observed: % bytes). Resolve or explicitly widen the cap before re-running this migration.',
      v_bad_count, v_max_bytes
      USING ERRCODE = 'check_violation';
  END IF;
END $$;

ALTER TABLE metaldocs.audit_events
  ADD CONSTRAINT audit_events_payload_size_cap
  CHECK (octet_length(payload::text) <= 65536) NOT VALID;

ALTER TABLE metaldocs.audit_events
  VALIDATE CONSTRAINT audit_events_payload_size_cap;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0266', 'DB-09 (P3, bounded): (a) durably REVOKE UPDATE/DELETE/TRUNCATE on metaldocs.audit_events from metaldocs_app (privileges were never granted — this hardens the already-correct insert/select-only posture documented by closed T-004/T-009, no distinct audit-writer role exists so none was invented); (b) add audit_events_payload_size_cap CHECK (octet_length(payload::text) <= 65536), added NOT VALID + VALIDATE CONSTRAINT for low-downtime rollout, pre-checked by a RAISE-on-violation DO block (T-010). Partitioning, pg-cron retention, ULID/UUIDv7 ids deferred — see wiki/modules/audit-tech-debt.md.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
