# Module #5a — `documents/{domain,application,repository}` — Review Findings

**Date:** 2026-05-22
**Reviewers:** go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer (all Sonnet 4.6)
**LoC:** ~9K across domain (459) + application (4926) + repository (3704)
**Raw:** 11C / 17H / 23M / 14L → deduped: **9C / 12H / 18M / 10L**

## Scope

`internal/modules/documents/domain/`, `application/`, `repository/`.
Migrations touching documents tables (0104, 0127, 0135, 0152, and others). Excludes `api/api.gen.go`.

---

## Critical

### C1 — `repository/repository.go:1381,1405` — `MarkArchived` + `Unarchive` discard `RowsAffected` → silent success on cross-tenant doc ID

```go
_, err = tx.ExecContext(ctx, `UPDATE public.documents SET archived_at = now()
    WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`, tenantID, docID)
if err != nil { return err }
return tx.Commit()
```

Both methods ignore `RowsAffected`. Archive/unarchive on a wrong `docID` (or cross-tenant `docID`) returns `nil` — HTTP 200, audit event written, no document actually changed. Every other state-transition in the file (`UpdateDocumentName`, `HeartbeatSession`, `ReleaseSession`) checks `n == 0`.

Recommend: `res, err := tx.ExecContext(...)` + `if n == 0 { return domain.ErrNotFound }` before `tx.Commit()`.

**Fix branch:** `fix/docs-5a-rows-affected-c1-c2` (land first)

---

### C2 — `repository/snapshot_repository.go:46,103,113,148,159` — 5 write methods discard `RowsAffected` → silent freeze/docx/pdf loss

`WriteSnapshot`, `WriteFreeze`, `WriteFinalDocx`, `WritePDF`, `AppendReconstruction` all use `_, err := exec.ExecContext(...)` with `WHERE tenant_id=... AND id=...` but never check `RowsAffected`. Worst cases:

- **`WriteFreeze`**: `values_frozen_at` never written; approval workflow proceeds on un-frozen document, no error surfaced.
- **`WriteFinalDocx`**: final DOCX S3 key lost; downstream export reads NULL key and fails with confusing `document not frozen` error.
- **`WritePDF`**: export record written, PDF pointer not persisted; cache returns stale data.

Recommend: `res.RowsAffected() == 0 → domain.ErrNotFound` on each method, matching `GrantAtomic` pattern used elsewhere.

**Fix branch:** `fix/docs-5a-rows-affected-c1-c2`

---

### C3 — `repository/repository.go:1054,1083` — `CreateCheckpoint` + `ListCheckpoints` missing `tenant_id` → IDOR

```go
SELECT current_revision_id::text FROM documents WHERE id=$1 FOR UPDATE   -- no tenant_id
SELECT ... FROM document_checkpoints WHERE document_id=$1                 -- no tenant_id
```

Any caller who learns a document UUID from another tenant can create checkpoints against it or enumerate all checkpoints on it.

Recommend: thread `tenantID` through both methods; add `AND tenant_id=$2` to both queries (or join through `documents`).

**Fix branch:** `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` (land third)

---

### C4 — `repository/repository.go:1176` — `RestoreCheckpoint` no `tenant_id` parameter → cross-tenant document restore

```go
SELECT coalesce(active_session_id::text,''), coalesce(current_revision_id::text,'')
FROM documents WHERE id=$1 FOR UPDATE   -- no tenant_id
```

Accepts `docID` + `actorUserID` only. Document lock and checkpoint resolution join both filter on `document_id` / `version_num` alone. Caller supplying a foreign-tenant `docID` can restore arbitrary checkpoints.

Recommend: add `tenantID` parameter; `WHERE id=$1 AND tenant_id=$2` on the initial lock; propagate to checkpoint join.

**Fix branch:** `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`

---

### C5 — `repository/repository.go:777` — `GetPendingForCommit` bare `pendingID` lookup, no tenant scope

```go
SELECT session_id, document_id, base_revision_id, content_hash, storage_key, expires_at, consumed_at
FROM autosave_pending_uploads WHERE id=$1   -- pendingID alone
```

No `tenantID`, no `userID` filter. Any authenticated caller with a valid `pendingID` UUID can read storage key, content hash, and session/document metadata of another tenant's pending upload. Downstream `CommitUpload` verifies session binding but doesn't protect the metadata read itself.

Recommend: add `tenant_id` to `autosave_pending_uploads` and filter; or join through `editor_sessions → documents.tenant_id`.

**Fix branch:** `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`

---

### C6 — `repository/repository.go:592` — `HeartbeatSession` no `tenant_id` filter → cross-tenant session keepalive

```go
UPDATE editor_sessions SET expires_at = now() + interval '5 minutes'
WHERE id=$1 AND user_id=$2 AND status='active'   -- no tenant_id
```

A user with a guessed `sessionID` from another tenant can extend that session's TTL indefinitely, preventing the legitimate holder from recovering the session lock. `editor_sessions` has no `tenant_id` column (see C7).

Recommend: add `tenant_id` column to `editor_sessions` (C7); once added, add `AND tenant_id=$3` to all session lookups including `HeartbeatSession`, `ForceReleaseSession`.

**Fix branch:** `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`

---

### C7 — `editor_sessions` schema missing `tenant_id` column — tenant isolation absent at row level

`editor_sessions` is scoped only via `document_id` FK. Every session operation (`HeartbeatSession`, `ForceReleaseSession`, session lookup in `AcquireSession`) provides no direct tenant guard. `editor_sessions.document_id` references `documents.tenant_id` indirectly but the DB cannot enforce isolation without the column.

Recommend: migration — `ALTER TABLE editor_sessions ADD COLUMN tenant_id UUID NOT NULL DEFAULT '...' REFERENCES documents(tenant_id)`; add index; then drop DEFAULT; update all session queries.

**Fix branch:** `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`

---

### C8 — `repository/repository.go:1147` — `GetRevision` no `tenant_id` scope → cross-tenant revision read

```sql
SELECT ... FROM document_revisions WHERE id=$1 AND document_id=$2
```

`document_id` prevents cross-document access but not cross-tenant: caller knowing `(revID, docID)` from another tenant can read that revision. Defence-in-depth boundary missing.

Recommend: join through `documents` on `tenant_id=$3` or denormalize `tenant_id` onto `document_revisions`.

**Fix branch:** `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`

---

### C9 — `domain/values_hash.go:18` — `json.Marshal` error silently discarded in hash computation

```go
v, _ := json.Marshal(values[k])
h.Write([]byte(k))
h.Write([]byte{0})
h.Write(v)   // v is nil when Marshal fails
```

When `Marshal` fails, `v = nil`, so zero bytes are written for that key's value. Two documents with different values for the same key can produce identical `values_hash` if one value is un-marshallable. Directly undermines freeze idempotency check and values-hash collision detection in `FreezeService.Freeze`.

Recommend:
```go
func ComputeValuesHash(values map[string]any) (string, error) {
    v, err := json.Marshal(values[k])
    if err != nil { return "", fmt.Errorf("marshal value for key %q: %w", k, err) }
    ...
}
```

**Fix branch:** `fix/docs-5a-values-hash-c9` (land second — standalone)

---

## High

### H1 — `application/fillin_service.go:252` — `err == sql.ErrNoRows` direct equality

```go
if err == sql.ErrNoRows {
```

Same pattern as 4 sites in auth, 2 in iam. `user_area_repository.go:192` in iam uses `errors.Is` correctly. This site is in the fill-in hot path; a wrapped error causes silent `nil, nil` return → validation bypass on placeholder schema load.

Recommend: `errors.Is(err, sql.ErrNoRows)`.

---

### H2 — `application/service.go:740` — discarded `DeleteObject` error on hash-mismatch cleanup

```go
_ = s.presigner.DeleteObject(ctx, meta.StorageKey)
return nil, domain.ErrContentHashMismatch
```

If delete fails, corrupted S3 object stays permanently. A retry with the same `pendingID` re-reads the bad object and may pass the TOCTOU check, potentially committing a corrupt revision. No log, no alert.

Recommend: log at WARN with storage key; don't return error (cleanup is best-effort), but silence is not.

---

### H3 — `application/service.go:331-334` — cleanup defer discards `DeleteObject` error

```go
defer func() {
    if cleanupKey != "" {
        _ = s.presigner.DeleteObject(context.Background(), cleanupKey)
    }
}()
```

Transient S3 failure → leaked temp object, no trace. Accumulated leaks cause storage waste and may cause content-addressed collision if key is reused.

Recommend: log the error with the storage key.

---

### H4 — `application/export_service.go:126` — `SignExportURL` passes caller-supplied `storageKey` directly to presigner

```go
func (s *ExportService) SignExportURL(ctx context.Context, storageKey string) (string, error) {
    return s.presigner.PresignObjectGET(ctx, storageKey)
}
```

If handler sources `storageKey` from request (not from DB-retrieved export row), attacker can presign arbitrary S3 keys including other tenants' objects. Zero validation inside the method.

Recommend: remove from public API or require `(tenantID, documentID, exportID)` — resolve key from DB inside the service, like `SignedDocxURL` does (lines 134-155).

---

### H5 — `repository/export_repository.go:37` — `InsertExport`/`GetExportByHash` no `tenant_id` filter

```sql
SELECT ... FROM document_exports WHERE document_id = $1 AND composite_hash = $2
```

`document_exports` has no `tenant_id`. Export metadata (including `storage_key`) readable cross-tenant if `document_id` is known.

Recommend: add `tenant_id` to `document_exports`; filter on it in both methods.

---

### H6 — `application/fillin_service.go:48` — `NewFillInServiceNoAuthz` is exported, silently bypasses all authz

```go
func NewFillInServiceNoAuthz(s SchemaReader, w FillInWriter) *FillInService {
    return &FillInService{schemas: s, writer: w}
}
```

Doc comment says "TEST-ONLY" but Go has no enforcement. Any future wiring mistake silently bypasses all `doc.edit_draft` capability checks.

Recommend: rename to `newFillInServiceNoAuthzForTest` (unexported) or move to `_test.go` / `testhelpers` package.

---

### H7 — `application/fillin_service.go:207` — `PHUser` validation silently skipped when IAM reader is nil

```go
if iam == nil {
    return nil  // skip user validation — any string accepted
}
```

Production wiring must call `WithIAMReader` but nothing enforces this at construction. Arbitrary strings accepted as PHUser placeholders.

Recommend: make `NewFillInService` require IAM reader as constructor argument, or panic/return error at `SetPlaceholderValue` time when type is `PHUser` and `iam == nil`.

---

### H8 — `repository/resolver_readers.go:18-59` — `RevisionReader` methods query `documents` table using `revisionID` parameter

`GetRevisionNumber`, `GetEffectiveFrom`, `GetAuthor`, `GetDocumentTitle` all: `WHERE tenant_id=$1::uuid AND id=$2::uuid` with second param named `revisionID`. The `documents` table has one row per document. Revision data lives in `document_revisions`. If callers pass actual revision UUIDs, these queries return empty/wrong results.

Recommend: clarify intent — if document-level reads, rename params to `documentID`; if per-revision, query `document_revisions JOIN documents`.

---

### H9 — `application/freeze_service.go:37,67` — anonymous interface field + `fanoutClient any` → silent nil panic

```go
valuesRead interface {   // anonymous — unreferenceable externally
    ListValues(ctx, tenantID, revisionID string) ([]repository.PlaceholderValue, error)
}
// ...
fanout, _ := fanoutClient.(FanoutClient)   // type assertion silently produces nil
```

Wrong concrete type passed to `NewFreezeService` → nil `FanoutClient` → panic at `s.fanout.Fanout(...)` line 207 at runtime, not at wiring time.

Recommend: promote anonymous interface to `PlaceholderValueReader` in `ports.go`; change `fanoutClient any` to `FanoutClient` directly (use `legacyFanout ...FanoutClient` variadic for backward compat).

---

### H10 — `repository/repository.go:584` — `AcquireSession` UPDATE on `documents` has no `tenant_id` guard

```sql
UPDATE documents SET active_session_id=$1, updated_at=now() WHERE id=$2
```

Prior `SELECT ... WHERE id=$1 AND tenant_id=$2` verified ownership. The subsequent `UPDATE documents` drops the `tenant_id` predicate. If a concurrent request races between the read and the update, tenant constraint can be violated.

Recommend: `WHERE id=$2 AND tenant_id=$3`.

---

### H11 — `repository/repository.go:403` — OFFSET pagination O(n) on large tenant datasets

`ListDocumentsPaginated` uses `OFFSET $n`. Postgres scans and discards all preceding rows on every page request. Degrades linearly with corpus size.

Recommend: cursor pagination with `WHERE (updated_at, id) < ($cursor_ts, $cursor_id)` and composite index `(tenant_id, updated_at DESC, id DESC)`.

---

### H12 — `migrations/0152_placeholder_fillin_columns.sql:83` — `document_placeholder_values` FK targets `documents(id)`, not `document_revisions(id)`

Column is named `revision_id` but FK goes to `documents(id)`. Go code at `repository.go:175` passes actual revision UUIDs from `document_revisions`. Schema inconsistency: FK allows any valid `document.id` as `revision_id`, meaning corrupt rows pass the constraint silently.

Recommend: change FK to `REFERENCES document_revisions(id) ON DELETE CASCADE` — or if the intent is document-level, rename the column to `document_id` and update all queries.

---

## Medium

| ID | Location | Finding |
|----|----------|---------|
| M1 | `repository.go:372` | `buildDocumentFilter` index arithmetic off-by-one risk with `len(args)` before limit/offset append; add unit test covering all filter combinations simultaneously |
| M2 | `fillin_service.go:277` | `GetPlaceholderValues`/`GetFillInSchema` return `nil, nil` when dependency unset — silent empty result vs misconfiguration |
| M3 | `fillin_service.go:73` | `LoadPlaceholderSchema`: `sql.ErrNoRows` not mapped to `domain.ErrNotFound` — raw sentinel leaks to handler |
| M4 | `freeze_service.go:79` | `FreezeService.Freeze` accepts `*sql.Tx` directly — breaks hexagonal port; use `repository.DBTX` interface |
| M5 | `fillin_service.go:144` | `log.Printf` in application layer — no tenant/revision context, not filterable in aggregation tooling |
| M6 | `fillin_authz.go:53` | `loadDocumentAreaCode` returns `"tenant"` fallback for missing doc → authz succeeds for non-existent resources |
| M7 | `repository.go:1312` | `UpdateComment` no authorship/role check — any tenant member can resolve/unresolve any other user's comment |
| M8 | `repository.go:1035` | `DeleteExpiredPending` exposed on `Repository` interface with no `authz.BypassSystem` guard |
| M9 | `snapshot_repository.go:98,112,158` | Parameter `tenant` instead of `tenantID` — inconsistency + argument-order confusion |
| M10 | `snapshot_repository.go:103` | `WriteFreeze` missing `::uuid` cast on `WHERE tenant_id=$3 AND id=$4` — PK index not used |
| M11 | `domain/model.go:12` | `DocStatusFinalized = "under_review"` — constant name contradicts wire value; migration artifact from 0142 |
| M12 | `domain/model.go:8`, `approval/domain/state.go:9`, `view_service.go:37` | Three parallel state representations (`DocumentStatus`, `DocState`, `viewableStatuses map[string]struct{}`) with no enforced bridge |
| M13 | `domain/export.go:6` | `Export` no constructor; `PaperSize string` unvalidated; `SizeBytes int64` unconstrained |
| M14 | `domain/comment.go:12` | `AuthorID string`, `ResolvedBy *string` — naked actor IDs; swap undetectable |
| M15 | `freeze_service.go:48` | `ApproverContext.UserID string`, `Capabilities []string` — should be `UserID` + `[]iamdomain.Capability` |
| M16 | `migrations/0127` | No `BEGIN/COMMIT` wrapper — trigger + function install not atomic |
| M17 | `migrations/0152:47` | `enforce_snapshot_on_submit` function missing `SET search_path = pg_catalog, pg_temp` |
| M18 | `fillin_repository.go:50` | `SeedDefaults` opens transaction even for empty placeholder slice — wasted round-trip; add early return |

---

## Low

| ID | Location | Finding |
|----|----------|---------|
| L1 | `repository.go:1369,1393` | `actorID` silently discarded in `MarkArchived`/`Unarchive` via `_ = actorID` — remove param or use it |
| L2 | `resolver_readers.go:95` | `GetFinalApprovalDate` returns `now()` via `COALESCE(max(signed_at), now())` when no approvals — fabricates timestamp; return `sql.NullTime` |
| L3 | `repository.go:528,610,641` | Bare `defer tx.Rollback()` inconsistent with rest of file using `defer func() { _ = tx.Rollback() }()` |
| L4 | `repository.go:290,317` | `ListDocuments` + `ListDocumentsForUser` unbounded (no LIMIT) — legacy paths; add LIMIT or deprecate |
| L5 | `domain/snapshot.go:19` | `SnapshotHashes` fields are `[]byte` — use `[32]byte` to enforce SHA-256 length at type level |
| L6 | `application/service.go` | Three constructors + two mutators for `Service` (`New`, `NewService`, `NewServiceWithSnapshot`, `WithDB`, `WithControlledDocumentDuplicator`) → 5 partial-init paths; `WithDB` mutates live service with no concurrency guard |
| L7 | `domain/model.go:25` | `Document` god struct with 30+ fields across 5 concerns; bridge fields (`ControlledDocumentID`, `ProfileCodeSnapshot`) should be grouped into `LegacyBridgeFields` |
| L8 | `domain/model.go:25`, `ports.go:13` | All domain IDs are naked `string`; `CapabilityChecker` port same — adopt `type DocumentID string`, `type TenantID string`, `type RevisionID string`, `type UserID string` once IAM ships them |
| L9 | `repository.go:1228` | `xmax::text::bigint <> 0` idiom to detect ON CONFLICT fire is undocumented Postgres internal; fragile under MVCC edge cases — prefer explicit boolean sentinel column |
| L10 | `export_repository.go` | `document_exports` has no index covering `(tenant_id, document_id)` — needed once tenant_id column is added |

---

## G3 Handoff — 9 Criticals

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5a-C1 | `repository/repository.go:1381,1405` MarkArchived/Unarchive no RowsAffected | Critical | leandrotca | TBC | `fix/docs-5a-rows-affected-c1-c2` | Backlog (land first) |
| 5a-C2 | `repository/snapshot_repository.go:46,103,113,148,159` 5 write methods no RowsAffected | Critical | leandrotca | TBC | `fix/docs-5a-rows-affected-c1-c2` | Backlog |
| 5a-C9 | `domain/values_hash.go:18` json.Marshal error discarded in hash computation | Critical | leandrotca | TBC | `fix/docs-5a-values-hash-c9` | Backlog (land second — standalone) |
| 5a-C3 | `repository/repository.go:1054,1083` CreateCheckpoint/ListCheckpoints no tenant_id | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog (land third) |
| 5a-C4 | `repository/repository.go:1176` RestoreCheckpoint no tenant_id parameter | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C5 | `repository/repository.go:777` GetPendingForCommit bare pendingID, no tenant scope | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C6 | `repository/repository.go:592` HeartbeatSession no tenant_id filter | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C7 | `editor_sessions` schema missing tenant_id column | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog (requires migration; verify C6 + H10 land same branch) |
| 5a-C8 | `repository/repository.go:1147` GetRevision no tenant_id scope | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |

**Fix branch land order:** `fix/docs-5a-rows-affected-c1-c2` → `fix/docs-5a-values-hash-c9` → `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`

### Cascade notes

- `fix/docs-5a-rows-affected-c1-c2`: cascades H2 (DeleteObject discard on hash mismatch, same service.go file), H3 (cleanup defer discard)
- `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8`: C7 (schema migration) must land before C3/C4/C5/C6/C8 query fixes; cascades H10 (AcquireSession UPDATE), H5 (export_repository tenant), M8 (DeleteExpiredPending no bypass guard)

---

## Module Assessment

This is the first sub-pass of the documents module (core lifecycle only). The predominant pattern is the same `RowsAffected` omission found in auth and IAM, but here it is more dangerous: `WriteFreeze` silence allows the approval workflow to proceed on un-frozen documents with no error surfaced anywhere in the call chain. The tenant isolation gaps (C3-C8) are structural — `editor_sessions` lacks `tenant_id` entirely, requiring a coordinated schema migration before runtime fixes can land. The `values_hash.go` silent Marshal discard (C9) is unique to this module and directly undermines the freeze idempotency guarantee that the entire approval flow depends on.
