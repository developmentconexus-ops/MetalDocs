-- 0317_template_version_content_hash_always.sql
-- ADR 0088 — a template version row can never exist without its content_hash.
--
-- ============================================================================
-- WHY
-- ============================================================================
--   `content_hash` carried TWO incompatible meanings:
--     (1) the verified hash of the object this version points at
--         (autosave CommitAutosave, CHECK chk_template_version_content_hash);
--     (2) "the user has actually edited this"
--         (the publish gate at application/lifecycle.go, and spawnNextDraft's
--          deliberate "leave ContentHash empty so the publish gate still forces
--          a real edit").
--
--   Meaning (2) was expressed by the ABSENCE of meaning (1). That is precisely
--   what made "a template version with no content" a state reachable BY
--   CONSTRUCTION -- and it is why creating a blank template and submitting it
--   untouched returned 409 UPLOAD_MISSING for a file the user never intended to
--   upload.
--
--   ADR 0088 collapses the field to meaning (1) only, and moves materialization
--   to creation time: every version-creation path copies its object PRE-TX to
--   the version's own canonical key and Confirms it, so the row is BORN with
--   the verified hash. Meaning (2) is not re-expressed anywhere -- whether an
--   unedited document deserves approval is the reviewer's judgment, not a shape
--   the system fabricates out of a null column.
--
--   The application change alone would leave the old shape legal at the DB
--   line. This migration makes the database the authoritative statement.
--
-- ============================================================================
-- WHAT CHANGES
-- ============================================================================
--   CONSTRAINT chk_template_version_content_hash_non_draft on
--   public.templates_template_version:
--     BEFORE: (status = 'draft' OR length(content_hash) = 64)
--             -- i.e. a DRAFT was allowed to carry '' forever.
--     AFTER:  (length(content_hash) = 64)
--             -- unconditional, in every status.
--
--   The constraint NAME is kept even though "_non_draft" no longer describes
--   it, deliberately: renaming would churn every error-message assertion and
--   every wiki/DB reference for zero semantic gain, and the definition (which
--   is what pg_get_constraintdef reports) states the rule unambiguously. The
--   sibling CHECK chk_template_version_content_hash (content_hash = '' OR
--   length = 64) is left untouched -- it is now strictly subsumed, harmless,
--   and dropping it is a separate concern from this ADR.
--
--   No trigger, index, policy or column is touched. docx_storage_key was
--   already NOT NULL; the missing half of the pointer was always the hash.
--
-- ============================================================================
-- DATA -- BACKFILL STRATEGY: PURGE, NOT REPAIR (with a referential assert)
-- ============================================================================
--   The honest options for a pre-existing draft row with content_hash = '':
--
--     (a) MATERIALIZE it -- copy the system blank asset to the row's canonical
--         key and record the hash. IMPOSSIBLE FROM SQL: it requires object-store
--         access (MinIO), which a migration does not and must not have. Faking
--         it by writing the pinned system-blank hash onto the row WITHOUT
--         copying the bytes would produce a row whose hash does not describe its
--         object -- the single meaning this ADR just established, violated by
--         the very migration that establishes it. Rejected.
--
--     (b) ASSERT the set is empty and fail loudly otherwise. Rejected as
--         dishonest here: these rows are EXPECTED to exist. Every blank template
--         ever created, and every draft ever spawned by CreateNextVersion (which
--         left the hash empty ON PURPOSE), is one. A migration that hard-fails
--         on the exact state the ADR was written to clean up is a migration that
--         cannot be applied.
--
--     (c) DELETE them.  <-- CHOSEN.
--         These rows are unusable by construction under the new rule and were
--         already half-broken under the old one: a content-less draft could not
--         be submitted (the approval kernel raised
--         ErrTemplateVersionNoContent) and could not be published (the gate this
--         ADR deletes). They are dead drafts. The product has NOT shipped and no
--         live tenant data depends on them (the dev database was wiped clean at
--         the 2026-07-29 baseline fold), so the cost of removal is zero and the
--         cost of carrying a permanently-invalid row is a permanently-weakened
--         invariant.
--
--   The purge is bounded by an EXPLICIT REFERENTIAL ASSERT that runs FIRST. If
--   any doomed row is pointed at by document_profiles.default_template_version_id,
--   controlled_documents.override_template_version_id,
--   templates_template.published_version_id, or an approval_instances row
--   (subject_kind='template'), the migration RAISES and changes nothing. All
--   four are structurally impossible for a content-less DRAFT that was never
--   submittable, so the assert should never fire -- and if it does, the database
--   is not the one ADR 0088 analyzed and a schema migration has no business
--   silently rewiring somebody's taxonomy or document configuration. Fail loud,
--   let a human decide.
--
--   Aggregate consistency after the purge:
--     * a template left with ZERO version rows is itself deleted -- a template
--       whose latest_version points at nothing is a new broken state, and
--       creating one while removing another is not a fix;
--     * a template that keeps other versions has latest_version rolled back to
--       MAX(version_number) of what remains, so the pointer stays truthful.
--
--   metaldocs.audit_events is deliberately NOT touched. It is the append-only,
--   hash-chained ledger: the history that a template once existed and was then
--   purged is TRUE, and deleting from that chain would both break its integrity
--   validation and destroy the record of this very migration's effect.
--
-- ============================================================================
-- TRIGGER / RLS NOTES (the ADR 0086 SQLSTATE 55006 class of bug)
-- ============================================================================
--   * templates_template and templates_template_version both carry
--     trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE. A DELETE
--     therefore requires metaldocs.asserted_caps to list one of the table's
--     mapped capabilities, exactly as db/reference-data does for its seed
--     writes. This migration asserts the same lifecycle caps the application
--     would rather than disabling the tripwire.
--   * That trigger's BEFORE-DELETE arm correctly RETURNs OLD (fixed by the
--     folded 0283) -- it does NOT return NEW, so it will not silently cancel
--     these deletions. This was checked, not assumed: returning NEW on DELETE
--     is the exact class of bug 0315 hit.
--   * enforce_template_version_tenant_consistent fires BEFORE INSERT OR UPDATE
--     only, so it is not involved in a DELETE, and metaldocs.tenant_id is
--     therefore deliberately left UNSET: with the GUC unset the tenant_isolation
--     RLS policy on both tables evaluates to TRUE for every row, which is what
--     lets one cross-tenant sweep see the whole purge set.
--   * NO deferred/constraint triggers exist on either templates table (the
--     DEFERRABLE INITIALLY DEFERRED policy triggers of ADR 0086 live on
--     approval_routes / approval_route_stages, which this migration never
--     touches), so the 55006 "cannot ALTER/disable constraint trigger mid-tx"
--     hazard does not arise here.
--
-- ============================================================================
-- REPLAY SAFETY
-- ============================================================================
--   Once the tightened CHECK is in place no content-less row can exist, so on a
--   second run the assert, the purges and the pointer rollback are all no-ops
--   over an empty set. The constraint swap is guarded by DROP ... IF EXISTS plus
--   a catalog check before ADD, so it is idempotent on its own.
--
-- Forward-only, idempotent, self-registering. Baseline (db/baseline) untouched.

BEGIN;

-- ── capability assertion for the purge writes ───────────────────────────────
-- Mirrors db/reference-data's idiom: assert the caps the application would hold
-- for these writes instead of disabling trg_require_cap_asserted. Tx-local.
SELECT set_config(
  'metaldocs.asserted_caps',
  '[{"cap":"template.create"},{"cap":"template.edit"},{"cap":"template.publish"},{"cap":"template.archive"}]',
  true
);

DO $migration$
DECLARE
    v_doomed         bigint;
    v_blocked        bigint;
    v_purged_ver     bigint;
    v_purged_tpl     bigint;
    v_rolled_back    bigint;
BEGIN
    -- ── 0. inventory ────────────────────────────────────────────────────────
    CREATE TEMP TABLE tmp_0317_contentless ON COMMIT DROP AS
    SELECT v.id, v.template_id, v.tenant_id, v.version_number, v.status
      FROM public.templates_template_version v
     WHERE v.content_hash IS NULL OR length(v.content_hash) <> 64;

    SELECT count(*) INTO v_doomed FROM tmp_0317_contentless;
    RAISE NOTICE '0317: % content-less template version row(s) found', v_doomed;

    IF v_doomed = 0 THEN
        RAISE NOTICE '0317: nothing to purge (already clean, or replay)';
    ELSE
        -- ── 1. referential assert: fail loud rather than rewire config ──────
        SELECT count(*) INTO v_blocked
          FROM tmp_0317_contentless d
         WHERE EXISTS (SELECT 1 FROM public.document_profiles p
                        WHERE p.default_template_version_id = d.id)
            OR EXISTS (SELECT 1 FROM public.controlled_documents c
                        WHERE c.override_template_version_id = d.id)
            OR EXISTS (SELECT 1 FROM public.templates_template t
                        WHERE t.published_version_id = d.id)
            OR EXISTS (SELECT 1 FROM public.approval_instances ai
                        WHERE ai.subject_kind = 'template'
                          AND ai.subject_key = d.id::text);

        IF v_blocked > 0 THEN
            RAISE EXCEPTION
                '0317: % content-less template version row(s) are referenced by a profile default, a controlled-document override, a published-version pointer, or an approval instance. ADR 0088 predicted zero such rows (a content-less draft was never submittable or publishable), so this database is not the one the ADR analyzed. Refusing to rewire configuration from a schema migration -- resolve these references by hand, then re-run.',
                v_blocked
                USING ERRCODE = 'P0001';
        END IF;

        -- ── 2. purge the dead draft rows ────────────────────────────────────
        DELETE FROM public.templates_template_version v
         USING tmp_0317_contentless d
         WHERE v.id = d.id;
        GET DIAGNOSTICS v_purged_ver = ROW_COUNT;

        -- ── 3. delete templates left with no version at all ─────────────────
        DELETE FROM public.templates_template t
         WHERE t.id IN (SELECT DISTINCT template_id FROM tmp_0317_contentless)
           AND NOT EXISTS (SELECT 1 FROM public.templates_template_version v
                            WHERE v.template_id = t.id);
        GET DIAGNOSTICS v_purged_tpl = ROW_COUNT;

        -- ── 4. keep latest_version truthful on the survivors ────────────────
        UPDATE public.templates_template t
           SET latest_version = sub.max_n
          FROM (SELECT v.template_id, max(v.version_number) AS max_n
                  FROM public.templates_template_version v
                 WHERE v.template_id IN (SELECT DISTINCT template_id FROM tmp_0317_contentless)
                 GROUP BY v.template_id) sub
         WHERE t.id = sub.template_id
           AND t.latest_version <> sub.max_n;
        GET DIAGNOSTICS v_rolled_back = ROW_COUNT;

        RAISE NOTICE '0317: purged % version row(s), % template row(s); rolled latest_version back on % template(s)',
            v_purged_ver, v_purged_tpl, v_rolled_back;
    END IF;
END
$migration$;

-- ── the authoritative line: content_hash is always a real sha256 ────────────

ALTER TABLE public.templates_template_version
    DROP CONSTRAINT IF EXISTS chk_template_version_content_hash_non_draft;

DO $constraint$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_constraint
         WHERE conname = 'chk_template_version_content_hash_non_draft'
           AND conrelid = 'public.templates_template_version'::regclass
    ) THEN
        ALTER TABLE public.templates_template_version
            ADD CONSTRAINT chk_template_version_content_hash_non_draft
            CHECK (length(content_hash) = 64);
    END IF;
END
$constraint$;

-- ── schema_migrations ledger ────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0317', 'ADR 0088: template version content is ALWAYS materialized -- content_hash collapses to ONE meaning (the verified sha256 of the object this version points at) and is present from the moment the row exists. Replaces CHECK chk_template_version_content_hash_non_draft on public.templates_template_version: the (status=''draft'' OR length(content_hash)=64) escape hatch becomes an unconditional length(content_hash)=64, so a content-less version row is illegal in EVERY status, not just after leaving draft. The escape hatch was how the field''s SECOND, incompatible meaning ("the user has actually edited this") was expressed -- by absence of the first -- which is what made "template version without content" reachable by construction and produced the operator-reported 409 UPLOAD_MISSING on submitting a freshly created blank template. Backfill strategy = PURGE, not repair, stated explicitly: materializing from SQL is impossible (it needs object-store access, and writing the pinned system-blank hash without copying the bytes would create the exact hash-does-not-describe-its-object lie this ADR abolishes), and asserting emptiness would hard-fail on the very rows the ADR exists to clean up (every blank template ever created, plus every draft CreateNextVersion deliberately spawned hash-less), so the content-less draft rows -- unusable by construction, never submittable and never publishable even under the old rules -- are DELETED. The purge is gated by a referential assert that RAISEs (changing nothing) if any doomed row is referenced by document_profiles.default_template_version_id, controlled_documents.override_template_version_id, templates_template.published_version_id or an approval_instances row with subject_kind=''template''; all four are structurally impossible for a never-submitted draft, so the assert is a tripwire, not an expected path, and a schema migration must not silently rewire tenant configuration. Aggregate consistency is preserved: templates left with zero versions are deleted, survivors have latest_version rolled back to MAX(version_number) of what remains. metaldocs.audit_events is deliberately untouched -- the append-only hash-chained ledger records history that remains true. Trigger discipline: templates_template[_version] carry trg_require_cap_asserted BEFORE INSERT OR DELETE OR UPDATE, so the migration asserts the same lifecycle caps the application would (metaldocs.asserted_caps, tx-local) rather than disabling the tripwire, and its BEFORE-DELETE arm was verified to RETURN OLD (folded 0283) so the deletions cannot be silently cancelled -- the failure mode 0315 hit. metaldocs.tenant_id is deliberately left unset so the tenant_isolation RLS policies on both tables admit every row for the one-pass sweep, and no DEFERRABLE constraint trigger exists on either templates table, so the ADR 0086 SQLSTATE 55006 hazard does not apply. Replay-safe: after the tightened CHECK no content-less row can exist, so the assert/purge/rollback are no-ops on re-run, and the constraint swap is DROP IF EXISTS + catalog-guarded ADD.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
