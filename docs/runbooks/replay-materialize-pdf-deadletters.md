# Runbook — Replay tripwire-era dead-lettered materialize/PDF outbox events

**Script:** `scripts/replay-materialize-pdf-deadletters.sql`
**Owner:** ops (hub-executed, one-time)
**Context:** QA-1 F9 — before commits `d1a3d757`/`e20a22c8`, the async
materialize and PDF consumers never asserted the capability GUC, so every
guarded write failed with P0001 `ErrCapabilityNotAsserted`, exhausted the
5-attempt budget, and dead-lettered the outbox row. The consumers are fixed;
this replay revives ONLY the rows killed by that defect so the pipeline
re-processes them.

## When to run

Exactly once, after deploying a build containing the F9 consumer fixes
(`unit/qr-b-pdf-pipeline` merged). Running it earlier just re-dead-letters
the rows after 5 more failed attempts.

## Preconditions

- Fixed `metaldocs-worker` image deployed and consumers running.
- Direct psql access to the application database.

## Procedure

```
psql "$DATABASE_URL" -f scripts/replay-materialize-pdf-deadletters.sql
```

The script is transaction-wrapped and idempotent. It resets
`dead_lettered_at`, `attempt_count`, `next_attempt_at`, and `last_error`
ONLY for rows where `event_type IN ('docx_materialize','docgen_v2_pdf')`,
`dead_lettered_at IS NOT NULL`, and `last_error LIKE
'%ErrCapabilityNotAsserted%'` — dead letters from any other failure cause
are untouched. It RETURNs the revived rows as the receipt.

## Verification

- Script output lists the revived `(id, event_type, aggregate_id)` rows.
- Within a few consumer poll cycles:
  `SELECT count(*) FROM metaldocs.outbox_events WHERE event_type IN
  ('docx_materialize','docgen_v2_pdf') AND dead_lettered_at IS NOT NULL;`
  should not grow back with `ErrCapabilityNotAsserted` errors.
- Affected documents' `pdf_status` transitions pending → ready.

## Rollback

None needed: a second run is a no-op (predicate no longer matches), and if
a revived row fails again for a *different* cause it dead-letters normally
and surfaces via `pdf_status='failed'` (QA-1 F13 fix).
