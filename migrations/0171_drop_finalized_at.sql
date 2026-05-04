-- migrations/0171_drop_finalized_at.sql
-- Group C / C6: drop denormalized finalized_at, derive from state history.
BEGIN;

CREATE OR REPLACE VIEW metaldocs.v_document_finalized AS
SELECT
    d.id AS document_id,
    (SELECT h.changed_at
       FROM metaldocs.document_state_history h
      WHERE h.document_id = d.id
        AND h.to_status = 'approved'
      ORDER BY h.changed_at ASC
      LIMIT 1) AS finalized_at
FROM metaldocs.documents d;

ALTER TABLE metaldocs.documents DROP COLUMN IF EXISTS finalized_at;

COMMIT;
