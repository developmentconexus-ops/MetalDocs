-- migrations/0171_drop_finalized_at.sql
-- DEFERRED: original Group C/C6 intent was to drop documents.finalized_at and
-- derive from metaldocs.document_state_history. That table was never created
-- by any earlier migration (Group C plan unfinished). Until document_state_history
-- ships, leaving finalized_at column in place. This migration only records the
-- ledger entry so subsequent migrations can run.
BEGIN;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0171', 'deferred: drop finalized_at + state-history view (document_state_history not yet built)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
