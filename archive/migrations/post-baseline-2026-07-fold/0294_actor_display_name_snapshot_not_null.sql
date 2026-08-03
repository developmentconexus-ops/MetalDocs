-- 0294_actor_display_name_snapshot_not_null.sql
-- M2d F2d.2 (approval remediation, ADR 0079): make the actor display-name
-- snapshot a DB-enforced invariant on both signed-action history tables.
--
-- Root cause (no-fallback principle): actor_display_name_snapshot was created
-- NULLABLE on approval_signoffs and approval_review_verdicts (migration 0288),
-- both insert paths bound it CONDITIONALLY (NULL when empty), and both read
-- paths FELL BACK (coalesce / live re-resolution). A live fallback lets a later
-- rename rewrite signed history — forbidden in a regulated eQMS. The snapshot is
-- always written from iam_users.display_name, which is NOT NULL
-- (0001_current_schema.sql), so the fallback masks a hole the upstream invariant
-- already closes. This migration makes that invariant explicit at the storage
-- layer (DB-enforces-invariants), letting F2d.2 delete the write-conditional and
-- the read fallbacks.
--
-- Forward-only + idempotent: backfill is a no-op on already-non-empty rows;
-- SET NOT NULL and the CHECK are re-runnable (DROP CONSTRAINT IF EXISTS first,
-- 0288/0290 idiom). Baseline (db/baseline/0001_current_schema.sql) stays frozen
-- (ADR 0032 baseline-frozen precedent) — this migration is the change of record.

BEGIN;

-- Backfill any NULL/'' snapshot from the authoritative source (iam_users.
-- display_name is NOT NULL, so every matched row receives a non-empty value).
-- Join on (user_id, tenant_id) — actor_user_id is TEXT (iam_users.user_id PK,
-- ADR: tokens actor-id text contract) and actor_tenant_id is uuid.

UPDATE public.approval_review_verdicts v
   SET actor_display_name_snapshot = u.display_name
  FROM metaldocs.iam_users u
 WHERE u.user_id = v.actor_user_id
   AND u.tenant_id = v.actor_tenant_id
   AND (v.actor_display_name_snapshot IS NULL OR v.actor_display_name_snapshot = '');

UPDATE public.approval_signoffs s
   SET actor_display_name_snapshot = u.display_name
  FROM metaldocs.iam_users u
 WHERE u.user_id = s.actor_user_id
   AND u.tenant_id = s.actor_tenant_id
   AND (s.actor_display_name_snapshot IS NULL OR s.actor_display_name_snapshot = '');

-- Enforce the invariant: NOT NULL + non-empty on both tables.

ALTER TABLE public.approval_review_verdicts
    ALTER COLUMN actor_display_name_snapshot SET NOT NULL;
ALTER TABLE public.approval_review_verdicts
    DROP CONSTRAINT IF EXISTS approval_review_verdicts_display_name_nonempty;
ALTER TABLE public.approval_review_verdicts
    ADD CONSTRAINT approval_review_verdicts_display_name_nonempty
    CHECK (actor_display_name_snapshot <> '');

ALTER TABLE public.approval_signoffs
    ALTER COLUMN actor_display_name_snapshot SET NOT NULL;
ALTER TABLE public.approval_signoffs
    DROP CONSTRAINT IF EXISTS approval_signoffs_display_name_nonempty;
ALTER TABLE public.approval_signoffs
    ADD CONSTRAINT approval_signoffs_display_name_nonempty
    CHECK (actor_display_name_snapshot <> '');

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0294', 'M2d F2d.2 (ADR 0079): actor_display_name_snapshot DB-enforced invariant. Backfills NULL/empty from metaldocs.iam_users.display_name (NOT NULL source), then SET NOT NULL + CHECK (<> '''') on both approval_review_verdicts and approval_signoffs. Closes the eQMS audit-truth hole (nullable snapshot + read fallback let a rename rewrite signed history); lets F2d.2 delete the write-conditional bind and the read coalesce/live-lookup fallbacks. Forward-only, idempotent; baseline frozen (ADR 0032).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
