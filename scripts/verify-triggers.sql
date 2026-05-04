-- scripts/verify-triggers.sql
-- Confirms enforce_snapshot_on_submit_trg exists on metaldocs.documents.
-- Run: psql $DATABASE_URL -f scripts/verify-triggers.sql
SELECT
    tgname,
    tgrelid::regclass AS target,
    pg_get_triggerdef(oid) AS definition
FROM pg_trigger
WHERE tgname = 'enforce_snapshot_on_submit_trg';
