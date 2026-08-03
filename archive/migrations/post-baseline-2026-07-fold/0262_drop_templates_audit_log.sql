-- 0262_drop_templates_audit_log.sql
-- DB-06 / T-013 (wiki/modules/templates-tech-debt.md:121): retire the
-- module-local public.templates_audit_log table now that BOTH sides of the
-- templates audit sink are canonical.
--
-- Runtime-truth verification performed before writing this migration:
--   - WRITE path: internal/modules/templates/repository/postgres.go
--     AppendAudit/AppendAuditTx (:604-624) delegate unconditionally to
--     r.audit (auditdomain.Writer, wired via WithAudit) which inserts into
--     metaldocs.audit_events through the hash-chained
--     internal/modules/audit/infrastructure/postgres/writer.go RecordTx.
--     grep-confirmed: no call site in internal/modules/templates writes
--     public.templates_audit_log directly (closed in Wave 1.8, F-07-sub-split,
--     well before this migration).
--   - READ path: internal/modules/templates/repository/postgres.go ListAudit
--     (:660-706) already SELECTs from metaldocs.audit_events filtered to
--     resource_type = 'template' AND resource_id = $1 AND tenant_id = $2,
--     ORDER BY occurred_at DESC — also closed in Wave 1.8/F-07-sub-split.
--     wiki/modules/templates-tech-debt.md T-013 and
--     wiki/architecture/data-model.md's prior audit-sink description were
--     stale against this runtime truth; both are corrected alongside this
--     migration.
--   Net effect: templates_audit_log has had zero readers and zero writers
--   in application code for the entirety of this migration's authoring.
--
-- Continuity check (migration-history evidence, not a live DB query):
--   grep across db/migrations/ and archive/migrations/ finds no prior
--   INSERT-SELECT / copy migration from templates_audit_log into
--   audit_events. The only migration ever touching templates_audit_log is
--   archived 0213_templates_tenant_id_uuid.sql, which only converted its
--   tenant_id column text -> uuid (no data movement to audit_events).
--   Historical rows written before the Wave 1.8 write-sink cutover would
--   therefore vanish from GET /api/v1/templates/{id}/audit if the table
--   were simply dropped. This migration backfills them first.
--
-- Backfill mechanics (append-at-tail, hash-chain-honest):
--   metaldocs.audit_events is a tamper-evident hash chain: row_hash for
--   sequence N is metaldocs.audit_event_row_hash(prev_hash_of_N-1, ...);
--   chain order is audit_sequence (insertion order), not occurred_at.
--   internal/modules/audit/infrastructure/postgres/writer.go RecordTx
--   establishes this precedent: it never assumes occurred_at correlates
--   with audit_sequence, only that each new row's prev_hash equals the
--   immediately-preceding row's row_hash. This migration follows the same
--   contract: historical rows are appended at the current chain tail, in
--   ascending (occurred_at, id) order for a stable within-batch sequence,
--   using the same canonical metaldocs.audit_event_row_hash(...) SQL
--   function the Go writer calls (no reimplementation, no drift risk). The
--   listing route orders by occurred_at DESC, so backfilled rows still sort
--   correctly for readers despite landing at the tail of audit_sequence.
--   The chain is inherently sequential (row_hash[N] depends on
--   row_hash[N-1]), so the backfill uses WITH RECURSIVE to walk the batch
--   one row at a time computing each row_hash off the previous one — a
--   window function cannot express this dependency.
--   Same advisory lock the Go writer takes (auditHashChainLockID,
--   internal/modules/audit/infrastructure/postgres/writer.go:20) is taken
--   here too, so this migration cannot race a concurrent RecordTx writer.
--   version_id (templates_audit_log has a dedicated column; audit_events
--   does not) is folded into the payload under "version_id", matching
--   internal/modules/templates/repository/postgres.go:626-637 (auditEvent),
--   which does the same for live writes — ListAudit already knows to lift
--   it back out (:697-702).
--   trace_id has no templates_audit_log equivalent; backfilled rows use the
--   literal 'migration:0262' so they are identifiable as backfilled if ever
--   inspected directly (audit_events.trace_id has no format constraint).
--   Idempotent via a deterministic synthetic id ('tpl-audit-bf-' || source
--   bigint id) + a NOT EXISTS guard against audit_events, so reruns (or a
--   partial prior run) do not duplicate or re-chain rows.
--
-- DESTRUCTIVE: drops public.templates_audit_log (and its owned sequence)
-- after the backfill above. IF EXISTS guards make both the backfill and the
-- drop safe to run exactly once, and safe to no-op if 0262 already applied
-- or if the table was already removed out-of-band.

BEGIN;

-- ── Backfill: copy any templates_audit_log rows not yet in audit_events ────
DO $$
DECLARE
  seed_hash text;
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'templates_audit_log'
  ) THEN
    PERFORM pg_advisory_xact_lock(90120260513004);

    SELECT COALESCE(
      (SELECT row_hash FROM metaldocs.audit_events WHERE row_hash <> '' ORDER BY audit_sequence DESC LIMIT 1),
      ''
    ) INTO seed_hash;

    CREATE TEMP TABLE tpl_audit_backfill_src ON COMMIT DROP AS
    SELECT
      'tpl-audit-bf-' || t.id::text AS id,
      t.occurred_at,
      t.actor_id,
      t.action,
      'template'::text AS resource_type,
      t.template_id::text AS resource_id,
      CASE
        WHEN t.version_id IS NOT NULL THEN
          COALESCE(t.details, '{}'::jsonb) || jsonb_build_object('version_id', t.version_id::text)
        ELSE COALESCE(t.details, '{}'::jsonb)
      END AS payload,
      'migration:0262'::text AS trace_id,
      t.tenant_id::text AS tenant_id,
      row_number() OVER (ORDER BY t.occurred_at, t.id) AS rn
    FROM public.templates_audit_log t
    WHERE NOT EXISTS (
      SELECT 1 FROM metaldocs.audit_events ae
      WHERE ae.id = 'tpl-audit-bf-' || t.id::text
    );

    -- Walk the batch in rn order, chaining each row's prev_hash to the
    -- previous row's freshly-computed row_hash (recursive: a set-based
    -- window function cannot express this self-referential dependency).
    WITH RECURSIVE chain AS (
      SELECT
        s.rn, s.id, s.occurred_at, s.actor_id, s.action, s.resource_type,
        s.resource_id, s.payload, s.trace_id, s.tenant_id,
        seed_hash AS prev_hash,
        metaldocs.audit_event_row_hash(
          seed_hash, s.id, s.occurred_at, s.actor_id, s.action, s.resource_type,
          s.resource_id, s.payload, s.trace_id, s.tenant_id
        ) AS row_hash
      FROM tpl_audit_backfill_src s
      WHERE s.rn = 1

      UNION ALL

      SELECT
        s.rn, s.id, s.occurred_at, s.actor_id, s.action, s.resource_type,
        s.resource_id, s.payload, s.trace_id, s.tenant_id,
        c.row_hash AS prev_hash,
        metaldocs.audit_event_row_hash(
          c.row_hash, s.id, s.occurred_at, s.actor_id, s.action, s.resource_type,
          s.resource_id, s.payload, s.trace_id, s.tenant_id
        ) AS row_hash
      FROM tpl_audit_backfill_src s
      JOIN chain c ON s.rn = c.rn + 1
    )
    INSERT INTO metaldocs.audit_events (
      id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id, prev_hash, row_hash
    )
    SELECT id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id, prev_hash, row_hash
    FROM chain
    ORDER BY rn;

    DROP TABLE tpl_audit_backfill_src;
  END IF;
END $$;

-- ── Drop the retired local sink ─────────────────────────────────────────────
-- No triggers, views, or FKs reference templates_audit_log (grep-confirmed
-- against db/baseline/0001_current_schema.sql and internal/); IF EXISTS
-- covers the "already dropped out-of-band" and "table never existed in this
-- environment" cases. CASCADE removes the PK and the owned id_seq sequence
-- along with the table (owned-sequence semantics: templates_audit_log_id_seq
-- is OWNED BY templates_audit_log.id, so DROP TABLE ... CASCADE reclaims it).
DROP TABLE IF EXISTS public.templates_audit_log CASCADE;

-- ── schema_migrations ledger ─────────────────────────────────────────────────
INSERT INTO public.schema_migrations (version, description)
VALUES ('0262', 'DB-06/T-013: backfill any pre-Wave-1.8 templates_audit_log rows into the canonical metaldocs.audit_events hash chain, then drop public.templates_audit_log (module-local audit sink; write path already canonical since F-07-sub-split, read path already canonical since same wave)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
