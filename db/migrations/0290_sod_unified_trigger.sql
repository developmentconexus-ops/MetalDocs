-- 0290_sod_unified_trigger.sql
-- M2b F7 (approval-kernel-backend, approval remediation design spec.md/plan.md
-- §5, W7): DB-tripwire-last-line SoD enforcement parity. Today exactly ONE
-- table (approval_signoffs) has a DB-level SoD trigger (enforce_signoff_sod /
-- trg_signoff_sod, db/baseline/0001_current_schema.sql:682-701). The F4
-- review-verdicts table (approval_review_verdicts, migration 0288) has NO
-- equivalent DB-level enforcement — the app-level check is the only line of
-- defense there today. This migration collapses the SoD rule into ONE shared
-- SQL function (rule text mirrors enforce_signoff_sod verbatim: an actor
-- cannot record a decision on an instance whose document they authored) and
-- attaches it to BOTH approval_signoffs (re-pointed, zero behavior change,
-- same error message/ERRCODE) and approval_review_verdicts (new — closes the
-- gap). App-checks-first (domain.CheckSoD, called identically from both
-- RecordSignoff and RecordVerdict as of this same feature) + DB-tripwire-last
-- (this migration) is now symmetric across both tables.
--
-- Both approval_signoffs and approval_review_verdicts already carry
-- approval_instance_id (FK to approval_instances), so the shared function's
-- lookup (approval_instances -> documents.created_by) is table-agnostic
-- as-is; no new columns are added by this migration.
--
-- Idempotent (CREATE OR REPLACE FUNCTION; DROP TRIGGER IF EXISTS before
-- CREATE TRIGGER — idiom copied from 0286/0288's DROP CONSTRAINT IF EXISTS
-- pattern, adapted for trigger objects).

BEGIN;

-- Shared function -------------------------------------------------------
-- Verbatim rule text and error shape from enforce_signoff_sod(), generalized
-- to any table with an approval_instance_id + actor_user_id column pair.

CREATE OR REPLACE FUNCTION public.enforce_approval_sod() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO 'pg_catalog', 'pg_temp'
    AS $$
DECLARE
  author_id TEXT;
BEGIN
  SELECT d.created_by INTO author_id
    FROM public.approval_instances i
    JOIN public.documents d ON d.id = i.document_id
   WHERE i.id = NEW.approval_instance_id;

  IF NEW.actor_user_id = author_id THEN
    RAISE EXCEPTION 'SoD: author cannot sign own revision'
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;

-- approval_signoffs: re-point the existing trigger at the shared function.
-- Same trigger name, same firing semantics (BEFORE INSERT), same error text —
-- zero behavior change for this table.

DROP TRIGGER IF EXISTS trg_signoff_sod ON public.approval_signoffs;
CREATE TRIGGER trg_signoff_sod BEFORE INSERT ON public.approval_signoffs
    FOR EACH ROW EXECUTE FUNCTION public.enforce_approval_sod();

-- approval_review_verdicts: new trigger — previously-missing DB-level SoD
-- enforcement for review-stage verdicts (F4 gap, closed here).

DROP TRIGGER IF EXISTS trg_review_verdict_sod ON public.approval_review_verdicts;
CREATE TRIGGER trg_review_verdict_sod BEFORE INSERT ON public.approval_review_verdicts
    FOR EACH ROW EXECUTE FUNCTION public.enforce_approval_sod();

-- The now-superseded single-table function is left in place (not dropped) so
-- any object referencing it by name (e.g. pg_dump snapshots, this migration
-- being re-applied) never errors on a missing function; it is simply no
-- longer referenced by any trigger after this migration runs.

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0290', 'M2b F7: unified SoD DB-tripwire. New enforce_approval_sod() function (rule text mirrors enforce_signoff_sod verbatim), attached to both approval_signoffs (trg_signoff_sod, re-pointed, zero behavior change) and approval_review_verdicts (trg_review_verdict_sod, new — closes the F4 DB-enforcement parity gap). App-level domain.CheckSoD is now called identically from both RecordSignoff and RecordVerdict.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
