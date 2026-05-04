# Group C — Document Lifecycle Correctness Design

> **Status:** approved 2026-05-03
> **Scope:** Fix the 6 follow-up bugs C1–C6 from `wiki/bugs/audit-2026-05-03.md` (lines 162–184). Concerns the document creation/lifecycle pipeline: archive transitions, snapshot atomicity, revision allocation, and finalization timestamp semantics.
> **Out of scope:** Changes to template authoring, placeholder schema evolution, S3 lifecycle policies, frontend list views (only backend filter predicate changes).

---

## Why This Spec Exists

Six bugs surfaced during deep audit (C1–C6) covering the document creation and lifecycle pipeline. They share two architectural root causes that this spec formalises:

1. **Documents born incomplete.** The template snapshot is written to `documents` rows in a separate post-commit step. Failure modes leave half-born rows with `status='draft'` but NULL snapshot columns. Concurrent revision allocation races on `MAX(revision_number)+1` and surfaces raw `pq: duplicate key` errors to callers.
2. **Lifecycle truth scattered across columns.** `finalized_at` is stamped at submit (wrong moment), `archived_at` is stamped during a status transition the trigger forbids, and `Service.Archive` requires a `fromFinalized bool` parameter to model state that should live in `document_state_history`.

This spec collapses both into two invariants:

- **Documents born complete.** Snapshot columns populated in the same INSERT as row creation.
- **Lifecycle truth = state history.** Status field plus `document_state_history` are authoritative. Denormalized timestamps are kept only when used as query predicates.

---

## Architecture: Two-Pillar Invariant

| Pillar | Statement | Enforcement |
|---|---|---|
| 1. Atomic creation | Snapshot resolved before INSERT, written in same statement. | Single SQL `INSERT` writes status + snapshot columns. Trigger `enforce_snapshot_on_submit_trg` already blocks NULL snapshots on submit. |
| 2. Lifecycle truth | Status + state history are authoritative. Denormalized timestamps are filter predicates only. | Drop `finalized_at`. Keep `archived_at` (used in `WHERE archived_at IS NULL` filter). Soft-archive via timestamp, never status mutation. |

**Module boundaries unchanged.** Fixes localised to:
- `internal/modules/documents/application/service.go`
- `internal/modules/documents/repository/repository.go`
- `internal/modules/documents/snapshot/service.go`
- New migration `0171_drop_finalized_at.sql`
- Audit doc + new ADR `wiki/decisions/0008-soft-archive-via-timestamp.md`

---

## Per-Bug Fix Design

### C1 — Service.Archive blocked by transition trigger

**Files:**
- Modify: `internal/modules/documents/application/service.go:594` (Archive method)
- Modify: `internal/modules/documents/repository/repository.go` (new `MarkArchived`, drop archived_at write from `UpdateDocumentStatus`)
- Modify: list/search query builders (add default `archived_at IS NULL`)

**Problem:** Migration 0142 trigger `enforce_document_transition()` defines no transition into `archived`. `Service.Archive` calls `UpdateDocumentStatus(... → archived)`, which the trigger rejects. Archive never works.

**Fix:** Soft-archive via `archived_at` timestamp. Status remains terminal (`obsolete`/`superseded`/`published`/etc.). No trigger change.

```go
// service.go
func (s *Service) Archive(ctx context.Context, tenantID, docID, actorID string) error {
    return s.repo.MarkArchived(ctx, tenantID, docID, actorID)
}

// repository.go
func (r *Repo) MarkArchived(ctx context.Context, tenantID, docID, actorID string) error {
    _, err := r.db.ExecContext(ctx, `
        UPDATE metaldocs.documents
           SET archived_at = now(), updated_by = $3, updated_at = now()
         WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`,
        tenantID, docID, actorID)
    return err
}
```

`Unarchive` symmetric: sets `archived_at = NULL`.

List/search query builders gain default predicate `AND archived_at IS NULL`. Admin endpoints opt-in via explicit flag.

`Service.Archive(fromFinalized bool)` parameter dropped — no longer needed. Caller cleanup required (find via `grep -r "Archive("`).

ADR: `wiki/decisions/0008-soft-archive-via-timestamp.md` records the rationale (regulatory requirement that terminal status remain immutable; archive is orthogonal storage concern).

**Test:**
- Unit (sqlmock): `TestMarkArchived_SetsTimestampNoStatusChange`
- Integration: `TestArchive_PublishedDoc_TerminalStatusUnchanged`, `TestList_DefaultExcludesArchived`, `TestList_AdminFlagIncludesArchived`

---

### C2/C4 — Snapshot atomicity (post-commit seeder + missing wiring)

**Files:**
- Modify: `internal/modules/documents/snapshot/service.go` — split into pure `ResolveTemplateSnapshot`
- Modify: `internal/modules/documents/application/service.go:324–353` (Create flow)
- Modify: `internal/modules/documents/repository/repository.go` — `CreateDocument` accepts `domain.Snapshot`, single INSERT writes all snapshot columns

**Problem:** `SnapshotFromTemplate` runs after `CreateDocument` commits (service.go:324–327, 350–353). If snapshot resolution fails, the documents row exists in `draft` with NULL snapshot columns. Trigger `enforce_snapshot_on_submit_trg` blocks transition to `under_review` but the orphan row sits in draft indefinitely. Retried snapshot resolves against current (drifted) template/area data, breaking point-in-time evidence semantics.

**Fix:** Inline snapshot resolution into Create. Snapshot service becomes a pure resolver (reads only). Single INSERT writes status + snapshot columns. No half-born rows possible.

```go
// snapshot/service.go — pure resolver
func (s *Service) ResolveTemplateSnapshot(ctx context.Context, tenantID, tvID string) (domain.Snapshot, error) {
    tv, err := s.tvRepo.GetByID(ctx, tvID)
    if err != nil { return domain.Snapshot{}, err }
    schema, err := s.schemaRepo.GetForVersion(ctx, tv.SchemaID)
    if err != nil { return domain.Snapshot{}, err }
    area, err := s.areaRepo.GetByCode(ctx, tenantID, tv.AreaCode)
    if err != nil { return domain.Snapshot{}, err }
    return domain.Snapshot{
        SchemaJSON:   schema.JSON,
        SchemaHash:   schema.Hash,
        Composition:  tv.CompositionConfig,
        BodyDocxKey:  tv.BodyDocxS3Key,
        AreaCode:     area.Code,
        AreaName:     area.Name,
    }, nil
}

// documents/application/service.go
func (s *Service) Create(ctx context.Context, cmd CreateCmd) (*Doc, error) {
    snap, err := s.snapshotSvc.ResolveTemplateSnapshot(ctx, cmd.TenantID, cmd.TemplateVersionID)
    if err != nil { return nil, fmt.Errorf("resolve snapshot: %w", err) }
    doc := domain.Document{ /* ... */ Snapshot: snap }
    docID, revID, sessionID, err := s.repo.CreateDocument(ctx, &doc, contentHash)
    if err != nil { return nil, err }
    return &Doc{ID: docID, RevID: revID, SessionID: sessionID}, nil
}
```

INSERT writes `placeholder_schema_snapshot`, `placeholder_schema_hash`, `composition_config_snapshot`, `body_docx_snapshot_s3_key`, `area_code_snapshot`, `area_name_snapshot` in same statement.

**Verification gate (writing-plans phase):** confirm `SnapshotFromTemplate` and `snapshot/service.go` perform reads only. If any non-`documents` table writes are found, halt and revisit with tx-propagation approach.

**Test:**
- Unit: `TestResolveTemplateSnapshot_ReturnsFrozenPayload`
- Integration: `TestCreate_BadTemplateVersion_NoRowWritten`, `TestCreate_PopulatesAllSnapshotColumns`

---

### C3 — `enforce_snapshot_on_submit_trg` verification

**Files:**
- New: `scripts/verify-triggers.sql` (one-shot diagnostic)
- Modify: `wiki/bugs/audit-2026-05-03.md` (mark C3 false positive with evidence)

**Problem:** Audit claims trigger missing. Code inspection: trigger DEFINED in migration `0152_placeholder_fillin_columns.sql` lines 46–49. Likely false positive in audit; needs verification against deployed DB state.

**Fix:** Diagnostic SQL confirms trigger present in target environment.

```sql
-- scripts/verify-triggers.sql
SELECT tgname, tgrelid::regclass, pg_get_triggerdef(oid)
  FROM pg_trigger
 WHERE tgname = 'enforce_snapshot_on_submit_trg';
-- expect 1 row on metaldocs.documents
```

If 1 row returned → close C3 with evidence link in audit doc. If 0 rows → escalate, add idempotent re-creation migration.

**Test:** none. Verification step in plan, not code.

---

### C5 — `revision_number` race

**Files:**
- Modify: `internal/modules/documents/repository/repository.go:43–49` (CreateDocument INSERT path)

**Problem:** `INSERT … VALUES (… COALESCE((SELECT MAX(d2.revision_number) … WHERE controlled_document_id=$6), 0)+1)`. Concurrent submits for the same `controlled_document_id` both compute `MAX+1 = N`. Both INSERT. Unique index `ux_documents_v2_cd_revision` rejects the second with `pq: duplicate key`. Caller sees raw DB error, not a meaningful conflict signal.

**Fix:** PostgreSQL transaction-scoped advisory lock per `(tenant_id, controlled_document_id)`. Lock acquired before INSERT, auto-released at COMMIT/ROLLBACK. Serialises revision allocation per document family. Tenant prefix prevents cross-tenant hash collisions.

```go
// inside CreateDocument tx, before INSERT
if _, err := tx.ExecContext(ctx, `
    SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
    doc.TenantID + ":" + doc.ControlledDocumentID); err != nil {
    return "", "", "", fmt.Errorf("acquire revision lock: %w", err)
}
// existing INSERT with COALESCE(MAX)+1
```

QMS workload is human-paced; lock contention negligible. Deterministic — no retry budget, no surprise errors.

**Test:** integration `TestCreateDocument_ConcurrentSubmitsSerialised` — 10 goroutines create revisions for same `controlled_document_id`, assert monotonic `revision_number` 1..10, zero duplicate-key errors.

---

### C6 — `finalized_at` semantic mismatch

**Files:**
- New: `migrations/0171_drop_finalized_at.sql`
- Modify: `internal/modules/documents/repository/repository.go:196` (remove `finalized_at = now()` write at submit)
- Modify: `internal/modules/documents/application/service.go` — `Service.Archive` signature drops `fromFinalized bool`
- Modify: callers of `Archive` (find via grep)
- Modify: any reader of `documents.finalized_at` → switch to view `v_document_finalized`

**Problem:** repository.go:196 stamps `finalized_at = now()` on submit (status → `under_review`). Submit is not finalization. The plan `2026-04-21-foundation-doc-approval-state-machine.md` already targets this column for drop in Phase 12. State history is the authoritative source; the column is denormalized noise.

**Fix:** Drop column. Add view that derives finalization timestamp from earliest `to_status='approved'` event in `document_state_history`.

```sql
-- migrations/0171_drop_finalized_at.sql
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
```

`Service.Archive(fromFinalized bool)` becomes `Service.Archive(ctx, tenantID, docID, actorID)`. Archive uses `archived_at` (C1), independent of finalization state.

Asymmetry vs C1 (which keeps `archived_at`): `archived_at` is a hot-path filter predicate (`WHERE archived_at IS NULL`). `finalized_at` is a cold-path audit timestamp. Different roles → different storage decisions.

**Test:**
- Integration: submit doc, approve, query view, assert `finalized_at = approval changed_at`. Reject doc, view returns NULL. Approve, then re-approve after revision: view returns earliest approval (point-in-time stable).

---

## Rollout Plan

| Phase | Tasks | Parallelism | Model |
|---|---|---|---|
| 0 | Worktree, codex spec validate, wiki-curator verify, snapshot read-only audit | sequential | sonnet |
| 1 | C3 verify (`pg_trigger` query + audit doc) ‖ C6 migration + view ‖ C1 `MarkArchived` repo + service | parallel (no file overlap) | sonnet / codex / codex |
| 2 | C5 advisory lock in `CreateDocument` repo path | sequential (touches repository.go) | codex |
| 3 | C2/C4 snapshot resolver split + Create flow + INSERT shape | sequential (touches repository.go + service.go) | codex |
| 4 | List/search predicate sweep (`archived_at IS NULL`) + `fromFinalized` caller cleanup | parallel | sonnet |
| 5 | Verify: `go test ./...`, `-tags=integration`, codex audit, smoke (create→submit→approve→archive) | sequential | sonnet → codex audit |
| 6 | Merge via `finishing-a-development-branch`, update `audit-2026-05-03.md`, wiki-curator stamps | sequential | sonnet |

**Phase review after each phase:** Opus.

**Verification gate (between Phase 0 and Phase 1):** confirm `SnapshotFromTemplate` is read-only. Writes elsewhere → halt, revisit Q2 with tx-propagation alternative.

---

## Testing Strategy

**Per-bug:** see "Per-Bug Fix Design" sections above.

**Cross-cutting:**
- `go test -mod=mod ./...` — full pass
- `go test -tags=integration ./tests/integration/documents/...` — lifecycle + concurrency tests
- Smoke: bootstrap dev DB, create doc from template, assert all snapshot columns populated, submit→approve→archive, assert `archived_at` set with terminal status unchanged, default list excludes archived
- Concurrent submit test (C5)
- Codex independent audit per-bug PASS/FAIL with file:line evidence
- Wiki-curator: refresh stamps on `wiki/concepts/document-lifecycle.md`, `wiki/modules/documents-*.md`, new ADR `wiki/decisions/0008-soft-archive-via-timestamp.md`

**Coverage targets:** new code ≥80% line coverage. No new lint warnings.

---

## Acceptance Criteria

- [ ] C1: `Service.Archive(published_doc)` succeeds, `archived_at` populated, status unchanged, default list excludes
- [ ] C2/C4: Create with bad `template_version_id` fails before any row written; successful Create populates all 6 snapshot columns in a single INSERT
- [ ] C3: `pg_trigger` query returns trigger row in target environment; audit doc updated with evidence
- [ ] C5: 10 concurrent submits for same `controlled_document_id` produce monotonic revisions 1..10, zero duplicate-key errors
- [ ] C6: `finalized_at` column dropped; `v_document_finalized` returns approval timestamp; `Service.Archive` signature drops `fromFinalized`
- [ ] All Go tests pass with `-mod=mod`
- [ ] Integration tests pass with `-tags=integration`
- [ ] Codex audit returns 6/6 PASS (or 5/5 + 1 false-positive-with-evidence for C3)
- [ ] Smoke test: full create→submit→approve→archive flow completes
- [ ] Audit doc updated, all 6 entries closed with commit SHAs

---

## Open Questions

None.

---

## References

- Audit: `wiki/bugs/audit-2026-05-03.md` (lines 162–184)
- Migration 0142: `migrations/0142_disable_legacy_compat.sql` (transition trigger)
- Migration 0152: `migrations/0152_placeholder_fillin_columns.sql` (snapshot enforcement trigger)
- Group A spec: `docs/superpowers/specs/2026-05-03-group-a-blockers.md`
- Group B spec: `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`
- Group D spec: `docs/superpowers/specs/2026-05-03-group-d-freeze-quality-design.md`
- Prior plan: `docs/superpowers/plans/2026-04-21-foundation-doc-approval-state-machine.md` (Phase 12 column drop alignment)
