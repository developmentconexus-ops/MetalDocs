-- Template version integrity constraints (F-DB1, F-DB4, F-DB7). Forward-only, idempotent.
--
-- (1) [F-DB1 HIGH] status has no CHECK constraint. The domain defines exactly
--     5 values (domain/version.go:10-16). Enforce the closed enum at DB level:
--     CHECK (status = ANY (ARRAY['draft','in_review','approved','published','obsolete']))
--     Sibling: public.template_versions at baseline line ~2120 (3-value legacy set),
--     public.approval_route_stages at baseline line ~1718 (ANY ARRAY pattern).
--
-- (2) [F-DB4 MEDIUM] content_hash text NOT NULL accepts ''. Two complementary CHECKs:
--     (a) format: CHECK (content_hash = '' OR length(content_hash) = 64)
--         A non-empty hash must be exactly 64 hex characters (SHA-256).
--         Sibling: documents_content_hash_len at baseline line ~2025 (octet_length
--         variant; our column is a hex string, not raw bytes, so length(x)=64
--         matches the SHA-256 hex representation used by the app).
--     (b) lifecycle: CHECK (status = 'draft' OR length(content_hash) = 64)
--         A non-draft version must carry a real hash (no empty placeholder).
--         The two CHECKs compose cleanly: format check allows '' only for drafts;
--         lifecycle check forbids '' once the row exits draft. Together they close
--         the gap without a trigger.
--
-- (3) [F-DB7 LOW] version_number has no positivity CHECK. Add:
--     CHECK (version_number >= 1)
--     Siblings: document_versions_mddm_version_number_check at baseline line ~1328,
--     approval_route_stages_stage_order_check at baseline line ~1721.
--
-- FAILURE CLASSES (all fail loudly — audit and resolve before re-running):
--   (a) F-DB1: existing row carries a status value outside the 5-value enum (data
--       corruption pre-dating this migration; MUST be resolved manually).
--   (b) F-DB4a: existing row has content_hash with length != 64 and != '' (corrupt
--       or truncated hash; MUST be resolved manually).
--   (c) F-DB4b: existing non-draft row has content_hash = '' (version was promoted
--       without a hash ever being written; MUST be resolved manually — the migration
--       surfaces the corruption correctly and must not be weakened to skip it).
--   (d) F-DB7: existing row has version_number < 1 (impossible under app invariants
--       but surfaced as corruption if present).

BEGIN;

DO $$
BEGIN
  -- F-DB1: status closed-enum check
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_template_version_status'
      AND conrelid = 'public.templates_template_version'::regclass
  ) THEN
    ALTER TABLE public.templates_template_version
      ADD CONSTRAINT chk_template_version_status
      CHECK (status = ANY (ARRAY['draft', 'in_review', 'approved', 'published', 'obsolete']));
  END IF;

  -- F-DB4a: content_hash format check (empty or exactly 64 chars)
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_template_version_content_hash'
      AND conrelid = 'public.templates_template_version'::regclass
  ) THEN
    ALTER TABLE public.templates_template_version
      ADD CONSTRAINT chk_template_version_content_hash
      CHECK (content_hash = '' OR length(content_hash) = 64);
  END IF;

  -- F-DB4b: content_hash lifecycle check (non-draft must have a real hash)
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_template_version_content_hash_non_draft'
      AND conrelid = 'public.templates_template_version'::regclass
  ) THEN
    ALTER TABLE public.templates_template_version
      ADD CONSTRAINT chk_template_version_content_hash_non_draft
      CHECK (status = 'draft' OR length(content_hash) = 64);
  END IF;

  -- F-DB7: version_number positivity check
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_template_version_number_positive'
      AND conrelid = 'public.templates_template_version'::regclass
  ) THEN
    ALTER TABLE public.templates_template_version
      ADD CONSTRAINT chk_template_version_number_positive
      CHECK (version_number >= 1);
  END IF;

  INSERT INTO public.schema_migrations (version, description)
  VALUES ('0251', 'template version status/content_hash/version_number CHECK constraints (F-DB1, F-DB4, F-DB7)')
  ON CONFLICT (version) DO NOTHING;
END
$$;

COMMIT;
