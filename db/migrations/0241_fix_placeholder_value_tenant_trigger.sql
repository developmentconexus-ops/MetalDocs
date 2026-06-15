-- db/migrations/0241_fix_placeholder_value_tenant_trigger.sql
-- Fixes the tenant-isolation trigger enforce_placeholder_value_tenant_consistent()
-- on public.document_placeholder_values.
--
-- BUG: the function resolves the row's tenant from
--   `SELECT tenant_id FROM documents WHERE id = NEW.revision_id`
-- but document_placeholder_values.revision_id is a document_revisions.id (the FK
-- document_placeholder_values_revision_id_fkey references document_revisions,
-- re-pointed there by archived migration 0191; the FillInRepository writes
-- documents.current_revision_id). A revision id never equals a documents.id, so
-- doc_tenant is always NULL and EVERY insert/update raises
--   'document % not found' (SQLSTATE 23503)
-- making the placeholder fill-in write path unusable on the canonical schema.
-- The check fails CLOSED (raise-on-every-write), so there is no cross-tenant
-- leak — but the tenant-consistency guarantee it is meant to enforce never runs.
--
-- The trigger (added in archived 0153_placeholder_values_tenant_consistency.sql)
-- was never updated when 0191 re-pointed the FK. Correct it to resolve the
-- revision's owning document via document_revisions, restoring the real
-- cross-tenant consistency check.
--
-- Decision + rollback posture: wiki/decisions/0033-fix-placeholder-value-tenant-trigger.md
-- Surfaced by: milestone-4c-test-fixture-framework/f4c.2-migrate-blocker-files (HS-2 escalation).
--
-- Idempotent (CREATE OR REPLACE), forward-only. The trigger binding is unchanged
-- (still BEFORE INSERT OR UPDATE on public.document_placeholder_values); only the
-- function body is corrected, so no trigger DDL is needed.

BEGIN;

CREATE OR REPLACE FUNCTION public.enforce_placeholder_value_tenant_consistent() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE doc_tenant UUID;
BEGIN
    SELECT d.tenant_id INTO doc_tenant
      FROM document_revisions r
      JOIN documents d ON d.id = r.document_id
     WHERE r.id = NEW.revision_id;
    IF doc_tenant IS NULL THEN
        RAISE EXCEPTION 'revision % not found', NEW.revision_id USING ERRCODE = 'foreign_key_violation';
    END IF;
    IF doc_tenant <> NEW.tenant_id THEN
        RAISE EXCEPTION 'tenant mismatch: document=% value=%', doc_tenant, NEW.tenant_id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0241', 'fix enforce_placeholder_value_tenant_consistent() to resolve revision_id via document_revisions (was looked up against documents.id; raised on every placeholder fill-in write)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
