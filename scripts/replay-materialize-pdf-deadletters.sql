-- scripts/replay-materialize-pdf-deadletters.sql
--
-- Fixes: revives metaldocs.outbox_events rows that were dead-lettered by the
-- now-fixed F9 tripwire bug (QA-1 era, event_type IN ('docx_materialize',
-- 'docgen_v2_pdf'), last_error containing 'ErrCapabilityNotAsserted'). The
-- underlying capability-assertion bug is fixed in the consumers; these rows
-- were exhausted (attempt_count = 5) and permanently dead-lettered before the
-- fix landed, so they need a one-shot reset to re-enter the claim queue.
--
-- When to run: ONCE, AFTER the fixed metaldocs-worker binaries (and any
-- fixed metaldocs-jobs binaries, if applicable) have been deployed to the
-- target environment. Running it before the fix is deployed will just
-- dead-letter the same rows again on next attempt.
--
-- How to run: hub/ops-executed only. This script does not run itself against
-- any live stack. Typical invocation against the target environment:
--   docker exec -i <postgres-container> \
--     psql -U <app-role> -d metaldocs -f - < scripts/replay-materialize-pdf-deadletters.sql
-- or, with a direct connection string:
--   psql "$DATABASE_URL" -f scripts/replay-materialize-pdf-deadletters.sql
--
-- Idempotent: the WHERE clause only matches rows that are still
-- dead-lettered with the tripwire error signature. After the first
-- successful run, the reset rows have dead_lettered_at IS NULL and
-- last_error IS NULL, so a second run affects 0 rows.
--
-- Scope: revives ONLY tripwire corpses (last_error LIKE the tripwire
-- signature). Genuinely poisoned events (any other last_error) are left
-- dead-lettered and must be triaged separately.
--
-- Reset targets the exact predicate ClaimUnpublished uses to select claimable
-- rows (internal/platform/messaging/outbox/postgres/consumer.go):
--   published_at IS NULL AND dead_lettered_at IS NULL
--   AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
-- so the reset rows are immediately claimable by the fixed consumers.

BEGIN;

UPDATE metaldocs.outbox_events
SET dead_lettered_at = NULL,
    attempt_count = 0,
    next_attempt_at = now(),
    last_error = NULL
WHERE event_type IN ('docx_materialize', 'docgen_v2_pdf')
  AND dead_lettered_at IS NOT NULL
  AND last_error LIKE '%ErrCapabilityNotAsserted%'
RETURNING event_id, event_type;

COMMIT;
