# Group D — Freeze Quality Design

> **Status:** approved 2026-05-03
> **Scope:** Fix the 8 freeze-quality bugs D1-D8 from `wiki/bugs/audit-2026-05-03.md` lines 186-204.
> **Out of scope:** Group C (document lifecycle), Group E (UX/frontend) — separate plans.

---

## Why This Spec Exists

Freeze is the regulatory artifact moment in MetalDocs: the point where a controlled document becomes immutable evidence. Eight bugs erode that guarantee:

- Computed placeholder values written outside the approval transaction (D7) — divergent state on rollback.
- Resolver outputs are non-deterministic (D3 NULL `effective_date` → DB `now()`), wrong (D1 raw UUIDs as names; D2 area code where wiki says name), or stale (D8 approvers from prior cycles).
- PDF dispatch is post-commit best-effort with no durability (D4), wrong idempotency key (D5), and silently swallowed errors (D6).

A QMS frozen DOCX must be deterministic, attributable, and durable. This spec aligns the freeze path with that bar.

---

## Architecture: Three Sub-Areas

| # | Sub-area | Bugs | Pattern |
|---|---|---|---|
| 1 | Tx integrity | D7 | Variadic `repository.DBTX` propagation (matches existing pattern) |
| 2 | Resolver correctness | D1, D2, D3, D8 | Snapshot-at-action-time + strict NULL handling |
| 3 | Dispatch durability | D4, D5, D6 | Transactional outbox + worker |

**Invariant after this work:** every value written to the frozen DOCX is either (a) deterministically derived from immutable inputs or (b) a snapshot taken at the moment of the action it represents. No live IAM lookups at freeze time. No post-commit fire-and-forget side effects.

---

## Per-Bug Fix Design

### D1 — Author/approver names: snapshot at action time

**Files:**
- `migrations/0173_signoff_actor_displayname_snapshot.sql` (new)
- `migrations/0174_documents_created_by_displayname_snapshot.sql` (new)
- `internal/modules/documents/approval/application/decision_service.go` — write snapshot on signoff insert
- `internal/modules/documents/application/service.go` (Create) — write snapshot on document insert
- `internal/modules/render/resolvers/author.go` — read `documents.created_by_display_name_snapshot`
- `internal/modules/render/resolvers/approvers.go` — read `approval_signoffs.actor_display_name_snapshot`
- `internal/modules/render/resolvers/resolver.go` — `AuthorInfo`/`ApproverInfo` already carry `DisplayName`; producers change

**Problem:** Resolvers receive `AuthorInfo{DisplayName: a.UserID}` — raw UUID. Frozen DOCX shows UUIDs in author/approver fields.

**Fix (regulatory-grade):**
- Migration 0173: `ALTER TABLE approval_signoffs ADD COLUMN actor_display_name_snapshot TEXT`. Backfill existing rows from current `iam_users.display_name` JOIN on `actor_user_id` (best-effort historical reconstruction; future signoffs are accurate by construction).
- Migration 0174: `ALTER TABLE documents ADD COLUMN created_by_display_name_snapshot TEXT`. Backfill from `iam_users` JOIN on `created_by`.
- `decision_service.recordSignoff` reads `iam_users.display_name` once inside the approval tx, writes to new column on insert.
- `documents.Service.Create` reads `iam_users.display_name` once, writes to new column on insert.
- Resolvers SELECT the snapshot column directly. Fallback to `actor_user_id` / `created_by` if NULL (legacy or IAM lookup failed).

**Why snapshot, not live lookup:** ISO 9001 / 21 CFR Part 11 expect the name on a controlled document to be the person's name at the time they signed, not the current name. Admin rename years later must not retroactively alter frozen DOCX content.

**Test:**
- Unit: `TestRecordSignoff_PersistsActorDisplayNameSnapshot`, `TestCreateDocument_PersistsCreatedByDisplayNameSnapshot`.
- Integration: `TestApproversResolver_UsesSnapshotNotLiveIAM` — sign, rename user via IAM, freeze, assert old name in resolver output.

---

### D2 — `{controlled_by_area}` resolves to code not name

**Files:**
- `migrations/0175_documents_area_name_snapshot.sql` (new)
- `internal/modules/documents/application/service.go` (Create) — populate `area_name_snapshot` from `process_areas.name`
- `internal/modules/render/resolvers/resolver.go` — add `AreaNameSnapshot string` to `ResolveInput`
- `internal/modules/render/resolvers/controlled_by_area.go` — return `in.AreaNameSnapshot`; include name (not code) in `inputs_hash`
- `internal/modules/documents/application/context_builder.go` — pass `AreaNameSnapshot` from documents row

**Problem:** Wiki documents `{controlled_by_area}` as the area's display name. Resolver returns `in.AreaCodeSnapshot` (e.g., "QA-01" not "Quality Assurance — Area 1").

**Fix:**
- Migration 0175: `ALTER TABLE documents ADD COLUMN area_name_snapshot TEXT`. Backfill via `UPDATE documents d SET area_name_snapshot = pa.name FROM process_areas pa WHERE pa.tenant_id = d.tenant_id AND pa.code = d.area_code_snapshot`.
- `Service.Create` populates the column from `process_areas.name` at create time.
- Resolver returns the name. `inputs_hash` keyed on name (not code) so subsequent area renames don't break hash reproducibility (snapshot is immutable).
- Fallback to `area_code_snapshot` if `area_name_snapshot` NULL (legacy doc pre-backfill — log warn).

**Test:** Unit `TestControlledByAreaResolver_ReturnsAreaName`, integration `TestFreeze_AreaNameSnapshotInFrozenDOCX`.

---

### D3 — `effective_date` NULL fallback to DB `now()`

**File:** `internal/modules/render/resolvers/effective_date.go`, `internal/modules/documents/domain/errors.go`

**Problem:** `RevisionReader.GetEffectiveFrom` returns `now()` when column is NULL → non-deterministic across retries → breaks `values_hash` reproducibility.

**Fix:**
- New typed error `ErrEffectiveDateMissing` in `documents/domain/errors.go`.
- `EffectiveDateResolver.Resolve`: if `RevisionReader.GetEffectiveFrom` returns zero `time.Time` (sentinel for NULL), return `ErrEffectiveDateMissing`. Freeze fails 422 with clear message.
- `RevisionReader.GetEffectiveFrom` SQL changed: drop `COALESCE(effective_from, NOW())` if present; return zero `time.Time` for NULL row.

**Why strict:** approval flow requires `effective_from` set before submit per QMS workflow. NULL at freeze time is a real defect upstream — surface it.

**Test:** Unit `TestEffectiveDateResolver_NullReturnsTypedError`. Integration `TestFreeze_NullEffectiveFrom_Returns422`.

---

### D4 + D5 + D6 — PDF dispatch durability via transactional outbox

**Files:**
- `migrations/0176_pdf_dispatch_outbox.sql` (new)
- `internal/modules/render/fanout/pdf_outbox_repository.go` (new)
- `internal/modules/render/fanout/pdf_outbox_worker.go` (new)
- `internal/modules/render/fanout/pdf_outbox_repository_test.go` (new)
- `internal/modules/render/fanout/pdf_outbox_worker_test.go` (new)
- `internal/modules/documents/approval/application/decision_service.go` — replace post-commit `pdfDispatcher.Dispatch` with `outbox.Enqueue(tx, ...)`
- `internal/platform/bootstrap/api.go` — wire outbox repo + worker, start goroutine

**Problem (D4):** `pdfDispatcher.Dispatch` runs after `tx.Commit()`. Process crash between commit and dispatch leaves an approved doc with no PDF, no recovery path.

**Problem (D5):** Dispatcher input field labeled `RevisionID` actually receives `DocumentID`. Idempotency key collides on revision 2 — second freeze's PDF silently dropped by consumer dedup.

**Problem (D6):** `_ = s.pdfDispatcher.Dispatch(...)` discards error. No log, no metric, invisible failures.

**Fix:**

Migration 0176:
```sql
CREATE TABLE metaldocs.pdf_dispatch_outbox (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL,
  revision_id     UUID NOT NULL,
  content_hash    BYTEA NOT NULL,
  status          TEXT NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','processing','dispatched','failed')),
  attempts        INT NOT NULL DEFAULT 0,
  last_error      TEXT,
  claimed_at      TIMESTAMPTZ,
  next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  dispatched_at   TIMESTAMPTZ,
  CONSTRAINT ux_pdf_dispatch_outbox_revision UNIQUE (tenant_id, revision_id)
);
CREATE INDEX ix_pdf_dispatch_outbox_pending
  ON metaldocs.pdf_dispatch_outbox (next_retry_at)
  WHERE status IN ('pending','processing');
```

Repository surface:
```go
type PDFOutboxRepository interface {
    Enqueue(ctx context.Context, tx repository.DBTX, tenantID, revisionID string, contentHash []byte) error
    ClaimPending(ctx context.Context, limit int) ([]OutboxRow, error)         // FOR UPDATE SKIP LOCKED, sets status='processing', claimed_at=NOW()
    MarkDispatched(ctx context.Context, id string) error
    MarkFailed(ctx context.Context, id string, errStr string, nextRetryAt time.Time) error
    ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error) // claimed_at < NOW() - 5min back to pending
}
```

Worker:
- Goroutine started by bootstrap, polls every 5s.
- Per cycle: `ResetStaleClaims(5m)` → `ClaimPending(10)` → for each row: publish event with `(tenant_id, revision_id, content_hash)` → on success `MarkDispatched` → on failure `MarkFailed(err, NOW()+backoff)`.
- Backoff: `min(2^attempts * 30s, 30m)`. Max attempts 5, then `status='failed'`.
- Metric: `pdf_dispatch_outbox_pending_total` gauge, `pdf_dispatch_outbox_failed_total` counter.

Decision service change:
```go
// Inside recordSignoff tx, after WriteFinalDocx:
if shouldDispatchPDF {
    if err := s.pdfOutbox.Enqueue(ctx, tx, pdfTenantID, pdfRevisionID, contentHash); err != nil {
        _ = tx.Rollback()
        return SignoffResult{}, fmt.Errorf("recordSignoff: enqueue pdf outbox: %w", err)
    }
}
// Drop post-commit pdfDispatcher.Dispatch entirely.
```

**Why outbox:** atomic with approval (no lost dispatches), bounded retry (no infinite loops), observable (status + attempts + last_error per row), idempotent enqueue (UNIQUE constraint), idempotent consumer (content_hash dedup).

**D5 fix subsumed:** column is named `revision_id` and is typed correctly. No more "RevisionID actually DocumentID" confusion.

**D6 fix subsumed:** failures land in `last_error`, surface in metrics, never silently swallowed.

**Test:**
- Unit: `TestPDFOutbox_EnqueueIdempotent`, `TestPDFOutbox_ClaimPendingSkipLocks`, `TestPDFOutbox_MarkFailedAppliesBackoff`, `TestPDFOutbox_ResetStaleClaims`, `TestPDFOutboxWorker_PublishSuccessMarksDispatched`, `TestPDFOutboxWorker_PublishFailIncrementsAttempts`, `TestPDFOutboxWorker_MaxAttemptsMarksFailed`.
- Integration: `TestSignoffApproval_OutboxRowEnqueuedInSameTx`, `TestSignoffApproval_TxRollback_NoOutboxRow`, `TestPDFOutboxWorker_EndToEnd_PublishesEvent`.

---

### D7 — Computed placeholder values written outside approval tx

**Files:**
- `internal/modules/documents/repository/fillin_repository.go` — `UpsertValue(ctx, v, q ...DBTX)`
- `internal/modules/documents/application/freeze_service.go` — pass tx to `UpsertValue`; remove "intentionally outside" comment

**Problem:** `freeze_service.Freeze` calls `s.values.UpsertValue` which uses `r.db.ExecContext`, not the caller's `*sql.Tx`. If approval tx rolls back after computed values are upserted, the values persist. Next freeze attempt may not overwrite (resolver inputs may have changed → different `inputs_hash` → ON CONFLICT path takes new value, but the row exists with stale data window). Divergent state.

**Fix:**
- Change signature: `UpsertValue(ctx context.Context, v PlaceholderValue, q ...DBTX) error`. If `len(q) > 0`, use `q[0]` for ExecContext; else `r.db`.
- Freeze service passes its `*sql.Tx` (already in scope) to every `UpsertValue` call.
- Drop misleading comment block at line 145-148.

**Test:**
- Unit: `TestUpsertValue_UsesTxWhenProvided`, `TestUpsertValue_FallsBackToDB`.
- Integration: `TestFreeze_RollbackDiscardsComputedValues` — force tx error after `UpsertValue` calls, assert `document_placeholder_values` empty for revision.

---

### D8 — Approver list pulls from ALL instances, not current

**Files:**
- `internal/modules/render/resolvers/resolver.go` — `ResolveInput` adds `ApprovalInstanceID string`
- `internal/modules/render/resolvers/approvers.go` — SQL filter by `approval_instance_id`
- `internal/modules/documents/application/context_builder.go` — populate `ApprovalInstanceID` from active instance lookup

**Problem:** `ApproversResolver` queries `approval_signoffs` filtered only by `revision_id`. Documents on revision 2 see signoffs from revision 1's instance(s). Stale approver names in current cycle's frozen DOCX.

**Fix:**
- `ResolveInput` carries `ApprovalInstanceID` (already known by context builder — it's the instance triggering the freeze).
- `ApproversResolver` SQL: `WHERE approval_instance_id = $1 AND tenant_id = $2`.
- Context builder reads `approval_instances WHERE document_id=... AND status='approved' ORDER BY created_at DESC LIMIT 1` if instance ID not already in scope.

**Test:** Unit `TestApproversResolver_FiltersByInstanceID`. Integration `TestApproversResolver_StaleInstanceFiltered` — two instances on same doc, assert only current's signoffs returned.

---

## Cross-cutting Concerns

**Migration ordering:** 0173 → 0174 → 0175 → 0176. All idempotent (`IF NOT EXISTS` on columns, `ON CONFLICT DO NOTHING` on backfills).

**Backfill strategy:** Snapshot columns backfilled via JOIN to current source-of-truth tables (`iam_users`, `process_areas`). Acknowledge in audit doc that pre-existing rows reflect current state at migration time, not historical state. Going forward, snapshots are accurate by construction.

**Worker lifecycle:** `PDFOutboxWorker` started by `bootstrap.NewAPI` after DB connection ready. Stopped via context cancellation on shutdown. Single worker per process (no leader election needed — multiple replicas safe via `FOR UPDATE SKIP LOCKED`).

**Metrics:** new gauges `pdf_dispatch_outbox_pending_total`, `pdf_dispatch_outbox_processing_total`; counter `pdf_dispatch_outbox_failed_total`. Wire into existing Prometheus registry.

**Backwards compatibility:** `PDFDispatcher` (existing) kept temporarily for tests; production callers switched to outbox. Remove `PDFDispatcher` only after Phase 4 verified.

---

## Rollout Plan

| Phase | Tasks | Parallel | Model |
|---|---|---|---|
| 0 | Worktree, codex spec validate, wiki-curator verify | sequential | sonnet |
| 1 | D7 (tx propagation) ‖ D3 (typed error) ‖ D8 (instance filter) | parallel — no file overlap | codex / sonnet / sonnet |
| 2 | D1 (signoff + create snapshots, migrations 0173-0174, resolvers) | sequential | codex |
| 3 | D2 (migration 0175 + resolver + context builder) | sequential | codex |
| 4 | D4/D5/D6 (migration 0176 + repo + worker + decision_service swap + bootstrap) | sequential, largest | codex |
| 5 | Verify: `go test -mod=mod ./...`, integration, codex audit 8/8, smoke | sequential | sonnet → codex audit |
| 6 | Update audit doc, wiki-curator full pass, final Opus review, finishing-a-development-branch | sequential | sonnet → opus |

**Phase review after each:** Opus.

---

## Testing Strategy

**Per-bug tests:** see "Per-Bug Fix Design" sections above.

**Cross-cutting:**
- `go test -mod=mod ./...` — full suite passes
- `go test -tags=integration ./tests/integration/...` — freeze rollback, outbox end-to-end, snapshot retention across user rename
- Smoke: submit doc → approve → verify frozen DOCX shows display names not UUIDs, area name not code, deterministic `effective_date`; verify PDF dispatched within 10s of last signoff (poll outbox); verify revision 2 PDF dispatch not deduped against revision 1
- Codex independent audit (Group A / B pattern) — per-bug PASS/FAIL with file:line evidence
- Wiki-curator agent: refresh stamps on `wiki/concepts/freeze-pipeline.md`, `wiki/concepts/placeholders.md`, `wiki/modules/render-*.md`, new ADR if introduced

**Coverage targets:** new code ≥80% line coverage. No new lint warnings.

---

## Acceptance Criteria

- [ ] D1: frozen DOCX shows display names not UUIDs; admin rename after signoff does not change frozen content (snapshot integrity)
- [ ] D2: `{controlled_by_area}` returns area NAME; existing docs backfilled; `inputs_hash` stable across area rename
- [ ] D3: freeze with NULL `effective_from` returns 422 + `ErrEffectiveDateMissing` typed error
- [ ] D4: signoff completion + process kill before worker poll → PDF still dispatched on restart (durability)
- [ ] D5: revision 2 freeze publishes new event with distinct `(tenant_id, revision_id)` key — not deduped against revision 1
- [ ] D6: bus failure increments `attempts`, populates `last_error`, surfaces in metrics; eventual `status='failed'` after 5 retries
- [ ] D7: forced freeze tx rollback → `document_placeholder_values` empty for that revision
- [ ] D8: 2-instance test → resolver returns only current instance's approvers
- [ ] All Go tests pass with `-mod=mod`
- [ ] Integration tests pass with `-tags=integration`
- [ ] Codex audit returns 8/8 PASS with file:line evidence
- [ ] Smoke test: full freeze flow completes, PDF dispatched within 10s
- [ ] Audit doc updated, all 8 bugs marked fixed with commit SHAs
- [ ] Wiki stamps refreshed on freeze + render docs

---

## Open Questions

None.

---

## References

- Audit: `wiki/bugs/audit-2026-05-03.md` (lines 186-204)
- Group A spec: `docs/superpowers/specs/2026-05-03-group-a-blockers-design.md`
- Group B spec: `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`
- Existing `repository.DBTX` pattern: `internal/modules/documents/repository/freeze_writer.go`
- Existing governance event-in-tx pattern: `internal/modules/documents/approval/application/decision_service.go:336-348`
