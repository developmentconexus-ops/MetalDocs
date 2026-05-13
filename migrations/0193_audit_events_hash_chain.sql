-- Audit T-004: tamper-evident hash chain for metaldocs.audit_events.

BEGIN;

DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM public.schema_migrations WHERE version = '0193') THEN
    RAISE NOTICE '0193 already applied - no-op';
    RETURN;
  END IF;
END $$;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE metaldocs.audit_events
  ADD COLUMN IF NOT EXISTS audit_sequence BIGSERIAL,
  ADD COLUMN IF NOT EXISTS prev_hash TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS row_hash TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_audit_events_audit_sequence
  ON metaldocs.audit_events (audit_sequence);

CREATE OR REPLACE FUNCTION metaldocs.audit_event_row_hash(
  p_prev_hash TEXT,
  p_id TEXT,
  p_occurred_at TIMESTAMPTZ,
  p_actor_id TEXT,
  p_action TEXT,
  p_resource_type TEXT,
  p_resource_id TEXT,
  p_payload JSONB,
  p_trace_id TEXT,
  p_tenant_id TEXT
) RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
AS $$
  SELECT encode(
    digest(
      concat_ws(
        E'\x1f',
        COALESCE(p_prev_hash, ''),
        COALESCE(p_id, ''),
        to_char(p_occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
        COALESCE(p_actor_id, ''),
        COALESCE(p_action, ''),
        COALESCE(p_resource_type, ''),
        COALESCE(p_resource_id, ''),
        COALESCE(p_payload::text, '{}'),
        COALESCE(p_trace_id, ''),
        COALESCE(p_tenant_id, '')
      ),
      'sha256'
    ),
    'hex'
  )
$$;

WITH RECURSIVE ordered AS (
  SELECT audit_sequence, id, occurred_at, actor_id, action, resource_type, resource_id,
         payload, trace_id, tenant_id,
         row_number() OVER (ORDER BY audit_sequence) AS rn
  FROM metaldocs.audit_events
),
chain AS (
  SELECT audit_sequence, rn, ''::TEXT AS prev_hash,
         metaldocs.audit_event_row_hash('', id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id) AS row_hash
  FROM ordered
  WHERE rn = 1

  UNION ALL

  SELECT ordered.audit_sequence, ordered.rn, chain.row_hash AS prev_hash,
         metaldocs.audit_event_row_hash(chain.row_hash, ordered.id, ordered.occurred_at, ordered.actor_id, ordered.action, ordered.resource_type, ordered.resource_id, ordered.payload, ordered.trace_id, ordered.tenant_id) AS row_hash
  FROM ordered
  JOIN chain ON ordered.rn = chain.rn + 1
)
UPDATE metaldocs.audit_events AS event
SET prev_hash = chain.prev_hash,
    row_hash = chain.row_hash
FROM chain
WHERE event.audit_sequence = chain.audit_sequence
  AND event.row_hash = '';

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'audit_events_prev_hash_shape'
      AND conrelid = 'metaldocs.audit_events'::regclass
  ) THEN
    ALTER TABLE metaldocs.audit_events
      ADD CONSTRAINT audit_events_prev_hash_shape
      CHECK (prev_hash = '' OR prev_hash ~ '^[0-9a-f]{64}$') NOT VALID;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'audit_events_row_hash_shape'
      AND conrelid = 'metaldocs.audit_events'::regclass
  ) THEN
    ALTER TABLE metaldocs.audit_events
      ADD CONSTRAINT audit_events_row_hash_shape
      CHECK (row_hash = '' OR row_hash ~ '^[0-9a-f]{64}$') NOT VALID;
  END IF;
END $$;

GRANT INSERT, SELECT ON TABLE metaldocs.audit_events TO metaldocs_app;
GRANT USAGE, SELECT ON SEQUENCE metaldocs.audit_events_audit_sequence_seq TO metaldocs_app;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0193', 'audit events hash chain and integrity validator support')
ON CONFLICT (version) DO NOTHING;

COMMIT;
