-- 0257_templates_version_status_under_review.sql
-- M1·T3: unify template version status vocabulary — rename "in_review" to
-- "under_review" so templates_template_version.status mirrors the documents
-- module's vocabulary (documents_status_check already uses "under_review").
--
-- Only the string value in chk_template_version_status changes.  All other
-- template-version status values (draft/approved/published/obsolete) are
-- unchanged.  The corresponding Go constant is VersionStatusUnderReview, value
-- "under_review".
--
-- FIX (forward-only, idempotent):
--   1. Drop the existing status CHECK constraint (which allows 'in_review').
--   2. Migrate any rows that carry status='in_review' to 'under_review'.
--   3. Re-add the constraint with the updated allowed-values array.

BEGIN;

-- ── Step 1: drop old constraint ──────────────────────────────────────────────

ALTER TABLE public.templates_template_version
  DROP CONSTRAINT IF EXISTS chk_template_version_status;

-- ── Step 2: migrate existing data ─────────────────────────────────────────────
-- The backfill touches version rows; both BEFORE-UPDATE triggers (cap-asserted
-- and tenant-consistent) must stand down for this system DDL data-fix, following
-- the established convention for trigger-guarded backfills (disable → data-fix →
-- re-enable in reverse order within the same transaction).

ALTER TABLE public.templates_template_version DISABLE TRIGGER trg_require_cap_asserted;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_template_version_tenant_consistent'
      AND tgrelid = 'public.templates_template_version'::regclass
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version DISABLE TRIGGER trg_template_version_tenant_consistent';
  END IF;
END
$$;

UPDATE public.templates_template_version
  SET status = 'under_review'
WHERE status = 'in_review';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_template_version_tenant_consistent'
      AND tgrelid = 'public.templates_template_version'::regclass
  ) THEN
    EXECUTE 'ALTER TABLE public.templates_template_version ENABLE TRIGGER trg_template_version_tenant_consistent';
  END IF;
END
$$;

ALTER TABLE public.templates_template_version ENABLE TRIGGER trg_require_cap_asserted;

-- ── Step 3: add updated constraint ───────────────────────────────────────────

ALTER TABLE public.templates_template_version
  ADD CONSTRAINT chk_template_version_status
  CHECK (status = ANY (ARRAY['draft'::text, 'under_review'::text, 'approved'::text, 'published'::text, 'obsolete'::text]));

-- ── schema_migrations ledger ─────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0257', 'templates_template_version: rename status in_review->under_review to match documents vocabulary (M1·T3)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
