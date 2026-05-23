# Module #6 Review — `internal/modules/controlleddocuments`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:database-reviewer, ecc:silent-failure-hunter, ecc:type-design-analyzer
**Severity totals:** 9 Critical / 13 High / 16 Medium / 9 Low
**Files reviewed:**
- `domain/controlled_document.go`, `domain/document_initializer.go`, `domain/port.go`
- `domain/resolution.go`, `domain/sequence.go`, `domain/visibility.go`
- `application/service.go`, `application/migration.go`
- `delivery/http/handler.go`, `delivery/http/routes.go`
- `infrastructure/repository.go`
- `module.go`

---

## Critical

### C1 — `delivery/http/routes.go:245` — GetActiveDocument no visibility/authz gate → IDOR

`GetActiveDocument` executes a raw repository query with no `CanRead` / visibility-scope check. Any authenticated user (any tenant, any role) can retrieve the full state of any controlled document by document ID. This is a full cross-tenant IDOR: `visibility_scope` (Internal, Restricted, Public) is stored on the document but never consulted before returning the payload.

**Recommend:** gate with `authz.Require(ctx, capability.ReadControlledDocument)` before the query. Add a `WHERE tenant_id = $tenantID AND visibility_scope IN (...)` predicate consistent with the user's visibility context. Pattern: `delivery/http/handler.go` in the documents module for the `CanRead` pattern.

**Fix branch:** `fix/cddocs-6-authz-idor-c1-c2-c3`

---

### C2 — `application/service.go:405-449` — `changeStatus` (Obsolete/Supersede) reads document without `CanRead` → restricted docs unprotected

`changeStatus` calls `repo.GetByID` to load the target document, then calls `authz.Require` to check the capability. The `GetByID` call has no visibility predicate — it succeeds for any document ID regardless of `visibility_scope`. An actor with the `ObsoleteDocument` capability can obsolete a Restricted document they have no read access to.

**Recommend:** apply visibility predicate inside `GetByID` (or add a separate `GetByIDForActor` variant that filters by visibility) before the capability check. Pattern: docs module `GetDocument` enforces visibility before returning.

**Fix branch:** `fix/cddocs-6-authz-idor-c1-c2-c3`

---

### C3 — `application/service.go:357-364` — `PreviewCode` / `PeekSeq` no authz check → sequence counters enumerable

`PreviewCode` and `PeekSeq` return the next sequence number for a document series. No `authz.Require` call precedes the read. Any authenticated user across any tenant can enumerate sequence counters for any series, leaking document throughput and series existence.

**Recommend:** add `authz.Require(ctx, capability.ReadControlledDocument)` and scope the series query to the caller's tenant. Quick fix, contained to `service.go`.

**Fix branch:** `fix/cddocs-6-authz-idor-c1-c2-c3`

---

### C4 — `application/service.go:420` — `changeStatus` calls `authz.Require` without prior `setAuthzGUC` → RLS context unset

The Postgres RLS GUC (`app.tenant_id`, `app.user_id`) is set by `setAuthzGUC` at the start of every transactional service call. In `changeStatus`, the first database operation (acquiring an advisory lock) opens a transaction, but `setAuthzGUC` is called *after* the `authz.Require` check. The authz check fires against an empty/stale GUC context, meaning RLS policies on the authz table read the wrong tenant, silently passing or failing for the wrong actor.

**Recommend:** move `setAuthzGUC(ctx, tx)` to immediately after the transaction opens and before any query or capability check. Pattern: all other service methods in this module.

**Fix branch:** `fix/cddocs-6-authz-guc-c4`

---

### C5 — `application/service.go:405-429` — TOCTOU race in `changeStatus`: `GetByID` outside transaction

`changeStatus` reads the document with `repo.GetByID` (non-transactional), validates the state transition, then opens a write transaction. Between the read and the write, a concurrent request can change the document status. Two concurrent `Obsolete` calls on the same document can both read `Active`, both pass state validation, and both write `Obsolete` — creating a double-obsolete with duplicate governance events and a corrupted audit trail.

**Recommend:** move `GetByID` inside the write transaction with `SELECT … FOR UPDATE`. Pattern: documents module `PublishDocument` uses `FOR UPDATE` on the document row before state transition.

**Fix branch:** `fix/cddocs-6-toctou-c5`

---

### C6 — `application/service.go:165,253,434` — `json.Marshal` errors discarded with `_` → nil governance event payloads

Three call sites discard the `json.Marshal` error when building governance event payloads:
- `service.go:165` (InitializeDocument event)
- `service.go:253` (UpdateDocument event)
- `service.go:434` (changeStatus event)

When Marshal fails (unreachable for well-typed structs but possible via interface values or future changes), the payload is `nil`. The event is still written to the outbox, creating a governance record with no payload — silent data loss in the audit trail.

**Recommend:** propagate the marshal error, fail the operation, or log + sentinel payload. Pattern: `platform/events.go` `MarshalPayload` helper used in other modules.

**Fix branch:** `fix/cddocs-6-audit-trail-c6`

---

### C7 — `infrastructure/repository.go:549` — `GetTemplateVersionState` no `tenant_id` predicate → cross-tenant leakage

```sql
SELECT ... FROM template_version_states WHERE template_version_id = $1
```

No `tenant_id` filter. A controlled document linked to a template version in tenant A can be read by a service call scoped to tenant B if it knows or guesses the template version ID. Template version IDs are UUIDs but are not secret.

**Recommend:** add `AND tenant_id = $tenantID` to the query. Pass tenant ID from the service layer (already in context). Requires confirming `template_version_states` schema has the column; if not, add it via migration.

**Fix branch:** `fix/cddocs-6-tenant-isolation-c7`

---

### C8 — `application/migration.go:30-35` — unbounded backfill SELECT → full-table scan, autovacuum contention

The one-shot migration job runs:
```go
rows, err := db.QueryContext(ctx, `SELECT id, document_id FROM controlled_documents WHERE legacy_ref IS NOT NULL`)
```

No `LIMIT` / pagination. On a large tenant this is a full-table scan that holds a snapshot for the entire migration duration, prevents autovacuum from reclaiming dead tuples, and can OOM the migration worker on millions of rows.

**Recommend:** paginate with keyset cursor (`WHERE id > $lastID LIMIT 1000`). Pattern: `platform/jobs/backfill_runner.go` pagination helper.

**Fix branch:** `fix/cddocs-6-migration-c8-c9`

---

### C9 — `application/migration.go:61-68` — `ON CONFLICT DO UPDATE RETURNING id` silently re-links already-linked documents; `skipped` counter logic inverted

```sql
INSERT INTO controlled_document_links (controlled_document_id, legacy_doc_id)
VALUES ($1, $2)
ON CONFLICT (controlled_document_id) DO UPDATE SET legacy_doc_id = EXCLUDED.legacy_doc_id
RETURNING id
```

The `ON CONFLICT DO UPDATE` always "succeeds" (always returns an `id`). The migration cannot distinguish new inserts from overwrites. The `skipped` counter is incremented when `RETURNING id` returns a row — which is every row including successful inserts — instead of being incremented only on true conflicts. Result: `skipped` counter equals `total` (all rows "skipped"), `processed` counter stays 0, making the migration silently appear to have done nothing while actually overwriting existing links.

**Recommend:** use `ON CONFLICT DO NOTHING` + check `RowsAffected()` to count actual inserts vs skips. Or use `INSERT … ON CONFLICT DO UPDATE … RETURNING (xmax = 0) AS inserted` to distinguish. Fix the counter logic.

**Fix branch:** `fix/cddocs-6-migration-c8-c9`

---

## High

### H1 — `domain/controlled_document.go:19` — `ControlledDocument` fully exported, no constructor → invariants bypassable

All invariants (valid status transitions, non-empty `SeriesID`, `TenantID`, sequence ordering) are documented in comments but enforced only in `domain/resolution.go` method calls. Callers can construct a zero-value `ControlledDocument{}` and assign fields directly, bypassing all invariants. This is already exploited in two test helpers.

**Recommend:** make `ControlledDocument` unexported or add a constructor `NewControlledDocument(...)` that validates required fields. Export a read-only interface or value type for callers that only need to inspect.

---

### H2 — `infrastructure/repository.go:113-145` — `InsertControlledDocument` no `RowsAffected` check

`InsertControlledDocument` calls `db.ExecContext` and does not check `RowsAffected()`. If the INSERT silently fails (conflict on a non-returning conflict clause, or a driver bug), the caller receives `nil` error and proceeds as if the document was created. Downstream operations (sequence allocation, link table inserts) then run against a non-existent document ID.

**Recommend:** add `if n, _ := result.RowsAffected(); n == 0 { return domain.ErrNotCreated }`. Pattern: documents module `InsertDocument` check.

---

### H3 — `application/service.go:95-120` — `InitializeDocument` does not validate `SeriesID` before sequence allocation

`InitializeDocument` calls `repo.AllocateSequence(ctx, req.SeriesID)` before checking that the series belongs to the caller's tenant. If the series UUID is guessed or leaked, an attacker can increment another tenant's sequence counter, causing a numbering gap in that tenant's document series.

**Recommend:** add a `SELECT id FROM series WHERE id = $1 AND tenant_id = $tenantID` guard before `AllocateSequence`. Alternatively, add `tenant_id` to the `AllocateSequence` query itself.

---

### H4 — `delivery/http/handler.go:78-115` — all handlers read `tenantID` from context without checking the IAM middleware result

Handlers call `ctx.Value(tenantKey)` and cast directly to `uuid.UUID` without nil-check. If the IAM middleware fails to inject the tenant (mis-configured route, future middleware removal), the cast panics or returns the zero UUID, silently scoping all subsequent queries to tenant `00000000-0000-0000-0000-000000000000`.

**Recommend:** use a typed `middleware.TenantIDFromContext(ctx)` helper that returns `(uuid.UUID, bool)` and return HTTP 401 if false. Pattern: iam module `TenantFromContext`.

---

### H5 — `application/service.go:200-240` — `UpdateDocument` re-computes content hash from caller-supplied body without verifying ETag

`UpdateDocument` accepts the new body, computes `sha256(body)`, and writes it without checking the request ETag against the current stored hash. Concurrent updates can silently overwrite each other — no OCC enforcement on `UpdateDocument`. Compare: `PublishDocument` does enforce ETag via `parseIfMatch`.

**Recommend:** add `If-Match` ETag check on `UpdateDocument` to be consistent with the pattern already established in the codebase.

---

### H6 — `infrastructure/repository.go:210-245` — `UpdateControlledDocument` no `RowsAffected` check

Same pattern as H2 for updates. Silent success on zero-row update means the caller believes the document was updated when it was not (cross-tenant ID, deleted document).

**Recommend:** same fix as H2.

---

### H7 — `application/service.go:330-355` — `LinkToLegacy` does not verify legacy document exists before creating link

`LinkToLegacy` inserts a `(controlled_document_id, legacy_doc_id)` row without verifying that `legacy_doc_id` exists in the `documents` table with matching `tenant_id`. This creates dangling foreign-key-like references (or would violate an actual FK if one is added later). Currently, reads that JOIN through the link return NULL for the legacy doc silently.

**Recommend:** add existence check (`SELECT 1 FROM documents WHERE id = $1 AND tenant_id = $2`) before insert or rely on a DB FK constraint.

---

### H8 — `domain/sequence.go:38-62` — `NextCode` panics on zero-value `Sequence`

`NextCode` calls `s.Pattern.Format(s.Counter)` where `Pattern` is a bare string that can be empty if the `Sequence` was constructed without a factory method (see H1). `Format` panics on empty pattern. No guard.

**Recommend:** validate in constructor or add nil/empty guard in `NextCode`.

---

### H9 — `application/service.go:450-490` — `ListControlledDocuments` accepts caller-supplied sort column without allowlist

```go
orderBy := req.OrderBy  // user-supplied
query := fmt.Sprintf("... ORDER BY %s %s", orderBy, dir)
```

`orderBy` is interpolated directly into SQL. While not a full injection risk (no `'` in column names), an adversary can probe internal column names, force index scans on unindexed columns, or cause a query error that leaks schema info.

**Recommend:** allowlist valid sort columns. Pattern: documents module `ValidSortColumns` map.

---

### H10 — `infrastructure/repository.go:380-410` — `GetSequenceForSeries` no `FOR UPDATE` inside transaction

`GetSequenceForSeries` reads the current counter inside the same transaction used to increment it, but does not use `SELECT … FOR UPDATE`. Concurrent transactions can read the same counter value and allocate the same sequence number to two documents.

**Recommend:** use `SELECT … FOR UPDATE` or rely on `UPDATE … RETURNING counter` atomically. The latter is already used in some paths but not this one.

---

### H11 — `delivery/http/routes.go:289-320` — `UploadAttachment` no file size limit

`UploadAttachment` calls `io.ReadAll(r.Body)` with no size cap. A client can stream an arbitrarily large body, exhausting service memory.

**Recommend:** wrap `r.Body` with `io.LimitReader(r.Body, maxAttachmentSize)` before `io.ReadAll`. Set `maxAttachmentSize` from config (default 50 MB, consistent with documents module).

---

### H12 — `application/migration.go:85-102` — migration job re-runs silently if `completed_at` column is missing

The idempotency check reads `completed_at IS NOT NULL` from `controlled_document_migration_runs`. If the column is absent (schema drift), the query returns an error that is swallowed with `err = nil` in the error path. The migration re-runs from scratch.

**Recommend:** propagate schema errors from the idempotency check instead of swallowing.

---

### H13 — `delivery/http/routes.go:195-230` — `DeleteControlledDocument` returns 200 without verifying row was deleted

`DeleteControlledDocument` calls the repository delete method and returns HTTP 200 regardless of `RowsAffected`. If the document ID belongs to another tenant (or was already deleted), the caller receives a false success.

**Recommend:** check `RowsAffected() == 0` and return HTTP 404.

---

## Medium

### M1 — `domain/port.go` — repository interface methods not grouped by read/write concern

All 22 methods are in one flat `Repository` interface. Handlers that only need reads take the full interface, making dependency boundaries unclear and mocking expensive in tests.

**Recommend:** split into `ReadRepository` and `WriteRepository` (or `QueryRepository` / `CommandRepository`). Pattern: iam module port split.

---

### M2 — `application/service.go:60-75` — service struct has 9 injected dependencies; no functional grouping

`Service` takes `repo`, `authz`, `sequencer`, `govEvents`, `migrator`, `objectStore`, `legacyLinker`, `templateSvc`, `clock` as constructor args. Constructor call is error-prone (positional mix-up).

**Recommend:** use an `Options` struct or functional options pattern. Pattern: auth module `NewService(repo Repository, opts ...Option)`.

---

### M3 — `domain/visibility.go:14-38` — `VisibilityScope` is a bare `string` type

`VisibilityScope` string type has no parse/validate constructor. Any string value (including invalid ones) compiles. Used directly in SQL predicates.

**Recommend:** define `const (VisibilityScopeInternal VisibilityScope = "internal" ...)` + `ParseVisibilityScope(s string) (VisibilityScope, error)`.

---

### M4 — `infrastructure/repository.go:550-590` — `ListByTemplateVersion` N+1 query pattern

`ListByTemplateVersion` returns a list of IDs and then the caller iterates calling `GetByID` per item.

**Recommend:** return full `ControlledDocument` slices from a single JOIN query.

---

### M5 — `application/service.go:155-175` — `InitializeDocument` success path logs at DEBUG but error paths not logged

Error paths in `InitializeDocument` return errors without logging context (doc ID, tenant ID, actor). Debug-only on success means failed initializations produce no server-side observability.

**Recommend:** add structured log on error paths (warn level, with non-sensitive context fields).

---

### M6 — `delivery/http/handler.go:301-340` — `UpdateDocument` handler does not forward `Content-Type` to storage

The handler stores attachments without preserving the `Content-Type` header from the upload. Download responses serve all attachments as `application/octet-stream`.

**Recommend:** capture and persist `Content-Type` alongside the attachment blob.

---

### M7 — `domain/resolution.go:18-55` — `Resolve` function has 5 nested if-else branches; cyclomatic complexity ~12

Hard to follow transition logic. Adding a new status requires reading all branches.

**Recommend:** model as a transition table `map[Status][]Status` (allowed targets). Pattern: approval module `AllowedTransitions`.

---

### M8 — `application/migration.go:110-135` — migration marks `completed_at` before committing link inserts

`migration.go` writes `completed_at` to the run record in the same DB call (but different statement) before the link batch completes. If the process dies mid-batch, the run is marked complete but links are partial.

**Recommend:** write `completed_at` in the same transaction as the final batch commit, or use a two-phase approach (mark `in_progress`, then `completed`).

---

### M9 — `infrastructure/repository.go:620-655` — `SearchControlledDocuments` ILIKE with leading wildcard prevents index use

```sql
WHERE title ILIKE '%' || $1 || '%'
```

Leading wildcard defeats the GIN/trgm index. On large tenants, full-table scans.

**Recommend:** use `pg_trgm` GIN index + `%` trigram search via `@@` operator, or accept trailing-wildcard-only search.

---

### M10 — `module.go:45-72` — module wire-up not validated at startup; nil service injected silently

`module.go` constructs the service with no validation that required dependencies are non-nil. If `templateSvc` is nil (dependency not registered), panics occur at runtime on first request.

**Recommend:** add nil-guard assertions in the `Wire()` function and return an error at startup. Pattern: auth module `Wire()` validator.

---

### M11 — `delivery/http/routes.go:350-380` — `GET /controlled-documents` returns 200 with empty list on auth failure instead of 401

The handler catches the authz error but falls through to return `[]ControlledDocument{}` with HTTP 200. API consumers cannot distinguish "no documents" from "not authorized".

**Recommend:** propagate the authz error as HTTP 403. Verify the `authz.Require` return path is checked.

---

### M12 — `domain/document_initializer.go:30-65` — `DocumentInitializer` exposes all fields as public

Same concern as H1 for the domain aggregate. `DocumentInitializer` has no constructor; partial initialization is valid to the compiler.

**Recommend:** constructor with required fields validated.

---

### M13 — `infrastructure/repository.go:700-735` — `GetRevisionHistory` no pagination

`GetRevisionHistory` selects all revisions for a document without LIMIT. Long-lived documents accumulate unbounded revision rows.

**Recommend:** add keyset pagination (`WHERE revision_number < $afterRevision LIMIT $pageSize`).

---

### M14 — `application/service.go:510-540` — `ArchiveControlledDocument` does not delete attachments

Documents are marked `Archived` but their object-store blobs and attachment DB rows remain. Storage cost grows unbounded; "archived" is misleading.

**Recommend:** schedule async attachment cleanup via River job on archive, or document the intentional retention policy in code.

---

### M15 — `delivery/http/handler.go:400-430` — response body not closed on error path

Several handlers call `resp, err := client.Do(req)` and return early on a non-nil error without `resp.Body.Close()` (when the response may still be non-nil on redirect errors).

**Recommend:** `defer resp.Body.Close()` immediately after the nil-check on `resp`.

---

### M16 — `application/service.go:290-320` — `RenameControlledDocument` does not re-compute sequence code

Renaming a document does not update the human-readable code derived from the series pattern. The stored `code` field drifts from the document's new title.

**Recommend:** either re-derive the code from the new title on rename, or document that `code` is immutable after assignment.

---

## Low

### L1 — `domain/sequence.go:10-18` — `Sequence` exported fields `Pattern` and `Counter` should be unexported

Both fields are read-only after allocation. Exported fields allow external mutation.

---

### L2 — `application/service.go:1` — package-level comment missing

No godoc comment on the `application` package.

---

### L3 — `infrastructure/repository.go:1` — `const` block uses raw SQL strings without named query constants

Long inline SQL in `const` blocks; queries named only by position. Consider named constants (`const queryGetByID = "..."`) for grep-ability.

---

### L4 — `delivery/http/routes.go:15-30` — route registration does not use `http.MethodGet` constants

Several routes use string literals `"GET"`, `"POST"` instead of `http.MethodGet`, `http.MethodPost`.

---

### L5 — `domain/port.go:8` — `Repository` interface comment says "DocumentRepository" (stale copy-paste)

---

### L6 — `application/migration.go:20` — migration job timeout hard-coded to 10 minutes

`ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)` — not configurable. Long-running migrations will be killed silently.

---

### L7 — `module.go:80` — `HealthCheck()` returns `nil` unconditionally without pinging the DB

The module registers a health endpoint that always reports healthy, masking DB connectivity issues.

---

### L8 — `delivery/http/handler.go:55-70` — `writeError` helper duplicated from documents module

Identical `writeError` function copy-pasted. Extract to `platform/httpresponse`.

---

### L9 — `infrastructure/repository.go:800-830` — dead code: `GetArchivedCount` function not called from any service

No callers. Safe to remove.

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/cddocs-6-authz-idor-c1-c2-c3` | C1 GetActiveDocument IDOR + C2 changeStatus visibility + C3 PreviewCode no authz | 1st (highest exposure, contained) |
| `fix/cddocs-6-authz-guc-c4` | C4 missing setAuthzGUC in changeStatus | 2nd (small diff, single call site) |
| `fix/cddocs-6-toctou-c5` | C5 TOCTOU race in changeStatus | 3rd |
| `fix/cddocs-6-audit-trail-c6` | C6 json.Marshal discards → nil governance payloads | 4th |
| `fix/cddocs-6-tenant-isolation-c7` | C7 GetTemplateVersionState no tenant_id | 5th (may need migration) |
| `fix/cddocs-6-migration-c8-c9` | C8 unbounded backfill + C9 ON CONFLICT inverted skipped counter | 6th |
