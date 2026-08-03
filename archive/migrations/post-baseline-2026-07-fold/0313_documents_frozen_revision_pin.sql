-- 0313_documents_frozen_revision_pin.sql
-- Etapa 5 §5.2 (F-QA3-1 option (a) remainder) — pin WHICH revision was frozen.
--
-- ============================================================================
-- WHY
-- ============================================================================
--   Freeze is a two-phase split (ADR 0015): Pin writes values_hash /
--   values_frozen_at in the approver's transaction, and the async Materialize
--   half renders the body some time later. Until now Materialize re-read
--   documents.current_revision_id at render time, so nothing recorded which
--   revision the approver actually signed, and any movement of the head between
--   the two phases silently changed the frozen content.
--
--   frozen_revision_id closes that drift window: it is written in the SAME
--   UPDATE as values_hash (SnapshotRepository.WriteFreeze) and read back by
--   FreezeService.Materialize, which then verifies the fetched bytes against
--   that revision's own document_revisions.content_hash before any artifact is
--   produced (fail closed on mismatch).
--
-- ============================================================================
-- EXPAND-ONLY, NO BACKFILL (deliberate)
-- ============================================================================
--   Nullable with no backfill. Rows frozen BEFORE this migration have no
--   recorded pin, and inventing one (e.g. current_revision_id at migration time)
--   would fabricate lineage: for a document whose head moved after the freeze,
--   the fabricated pin would name a revision that was never frozen while the
--   hash verification would still pass. Under the no-fallback principle a
--   missing pin must fail closed, not be guessed.
--
--   Consequence, surfaced deliberately: re-materializing a PRE-migration frozen
--   document (Etapa 5 §5.3 expurgo/backfill) fails closed with "no pinned frozen
--   revision" until a pin is recorded for it. That is an operations ruling, not
--   a schema default.
--
--   SANCTIONED PATH for those legacy NULL-pin rows: the operator-only repair,
--   FreezeService.RepairMaterialization (scripts/release-backfill). It takes a
--   FRESH pin in the repair transaction — re-reading the document's current
--   revision and writing frozen_revision_id through the same WriteFreeze seam
--   Pin uses — then enqueues materialize. That is an explicit, operator-
--   sanctioned RE-FREEZE from the current revision body with true lineage (the
--   pinned revision really is the body about to be rendered, and Materialize
--   still verifies its bytes against that revision's content_hash), which is
--   precisely what a blind migration-time backfill could not have been. Repair
--   re-pins documents that already carry a pin as well: choosing to repair IS
--   the decision to re-freeze, so the old pin is superseded. values_hash and
--   values_frozen_at are never recomputed — the approver signed those.
--
--   REFERENCES document_revisions(id) with no ON DELETE action: revisions are
--   append-only and are never deleted while their document lives, so a pin can
--   never dangle. Tenancy is enforced through documents (document_revisions
--   carries no tenant_id of its own) exactly as current_revision_id does.

BEGIN;

ALTER TABLE public.documents
    ADD COLUMN frozen_revision_id uuid REFERENCES public.document_revisions(id);

COMMENT ON COLUMN public.documents.frozen_revision_id IS
    'document_revisions.id pinned at freeze time (ADR 0015 Pin phase). Materialize renders THIS revision, not current_revision_id, and verifies the fetched bytes against its document_revisions.content_hash. NULL means no pin was recorded (pre-0313 freezes) and materialization fails closed.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0313', 'Etapa 5 §5.2 (F-QA3-1 option (a) remainder): add public.documents.frozen_revision_id uuid NULL REFERENCES document_revisions(id) — the freeze-time lineage pin. Written in the same UPDATE as values_hash by SnapshotRepository.WriteFreeze and read back by FreezeService.Materialize, which fetches the pinned revision body and fails closed unless its sha256 equals that revision''s recorded content_hash. Expand-only with NO backfill: pre-migration freezes keep a NULL pin and fail closed rather than carry fabricated lineage. documents.content_hash keeps its meaning (hash of the OUTPUT frozen docx).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
