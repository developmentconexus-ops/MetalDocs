# Module #5b — documents/{delivery,http,jobs}

**Reviewed:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:silent-failure-hunter, ecc:type-design-analyzer, ecc:database-reviewer
**Scope:** `internal/modules/documents/delivery/http/`, `internal/modules/documents/http/`, `internal/modules/documents/jobs/`

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 6     |
| High     | 19    |
| Medium   | 22    |
| Low      | 12    |

---

## Critical

### 5b-C1 — Content-hash DB error silently discarded → corrupt approval submission

**File:** `internal/modules/documents/delivery/http/handler.go:559-564`
**Fix branch:** `fix/docs-5b-finalize-c1-c2`

```go
_ = h.db.QueryRowContext(r.Context(),
    `SELECT COALESCE(content_hash, '') FROM document_revisions
      WHERE document_id = $1
      ORDER BY created_at DESC LIMIT 1`,
    docID,
).Scan(&contentHash)
```

Entire result discarded with `_`. DB connectivity failure, permission error, or missing row silently proceeds with `contentHash = ""`. Approval submission records a blank hash — breaking freeze idempotency and the audit trail with no error surfaced.
Fix: distinguish `sql.ErrNoRows` (proceed empty) from real errors (return 500, abort finalize).

---

### 5b-C2 — `err.Error()` string-concatenated into JSON response → info leak + JSON injection

**File:** `internal/modules/documents/delivery/http/handler.go:618-622`
**Fix branch:** `fix/docs-5b-finalize-c1-c2`

```go
http.Error(w, `{"error":"`+msg+`","detail":"`+err.Error()+`"}`, status)
```

`err.Error()` can contain SQL table names, column names, DB driver internals, file paths, or `"` characters that break the JSON structure. This is the only handler that bypasses `httpErr`/`problem.Write`. Leaks internal implementation details to the caller.
Fix: use `httpErr(w, status, msg)` unconditionally; log the real error server-side.

---

### 5b-C3 — `X-User-Roles` header trusted from caller → privilege escalation across all document handlers

**File:** `internal/modules/documents/delivery/http/handler.go:1165,1187`
**Fix branch:** `fix/docs-5b-header-roles-c3` (land first)

`withAdminCtx` and `hasRole` both read `X-User-Roles` directly from the incoming HTTP request header as a second, independent role source. Auth middleware strips `X-Tenant-ID` (auth middleware line 89) and `X-User-ID` (IAM middleware line 68), but **neither strips `X-User-Roles`**. Any caller reaching the service can set `X-User-Roles: system_admin` and gain admin-level access across every handler that calls `withAdminCtx` or `hasRole` — which is every handler in the file.

Fix: delete the `r.Header.Get("X-User-Roles")` branch from both functions; derive roles exclusively from `iamdomain.RolesFromContext`. Add `Del("X-User-Roles")` to the IAM middleware alongside the existing `Del("X-User-ID")`.

---

### 5b-C4 — Webhook accepts `tenant_id` from attacker-controlled body → cross-tenant PDF overwrite

**File:** `internal/modules/documents/http/pdf_webhook_handler.go:83`
**Fix branch:** `fix/docs-5b-webhook-tenant-c4`

`pdfCompleteBody.TenantID` (from request body) is passed directly to `WritePDF`. An attacker who can forge or replay a valid HMAC can supply any `tenant_id` and overwrite PDF data for documents belonging to a different tenant. `{id}` in the path is the document ID; the canonical `tenant_id` can be derived server-side by looking up the document.

Fix: look up the document by `{id}` to retrieve its canonical `tenant_id`; reject the request if the body's `tenant_id` does not match. Alternatively, derive tenant entirely server-side and ignore the body field.

---

### 5b-C5 — `DeleteExpiredPending` unbounded cross-tenant DELETE → table-wide lock

**File:** `internal/modules/documents/repository/repository.go:1035`
**Fix branch:** `fix/docs-5b-sweeper-unbounded-c5-c6`

```sql
DELETE FROM autosave_pending_uploads WHERE expires_at < $1 AND consumed_at IS NULL
```

No `LIMIT`. On large datasets this holds an exclusive lock on every matching row for the full duration of the delete, blocking concurrent autosave commits and presign operations across all tenants.

Fix: batch with `LIMIT 500` using a CTE with `RETURNING`:
```sql
WITH to_delete AS (
  SELECT id FROM autosave_pending_uploads
  WHERE expires_at < $1 AND consumed_at IS NULL LIMIT 500
)
DELETE FROM autosave_pending_uploads WHERE id IN (SELECT id FROM to_delete)
```

---

### 5b-C6 — `ExpireStaleSessions` unbounded cross-tenant UPDATE → table-wide lock

**File:** `internal/modules/documents/repository/repository.go:667`
**Fix branch:** `fix/docs-5b-sweeper-unbounded-c5-c6`

CTE `UPDATE editor_sessions SET status='expired' WHERE status='active' AND expires_at < $1` updates every stale session across all tenants in one statement; cascaded `UPDATE documents SET active_session_id=NULL` hits all related documents. Both updates compete with every live `AcquireSession`, `HeartbeatSession`, and `CommitUpload` across all tenants.

Fix: same LIMIT 500 batch pattern — select candidate IDs first, update only those.
Additionally: add index `CREATE INDEX ON documents (active_session_id) WHERE active_session_id IS NOT NULL` — currently the cascade does a full table scan on `documents` every sweeper tick.

---

## High

### 5b-H1 — Authorization deferred past idempotency replay → replay without auth check

**File:** `internal/modules/documents/delivery/http/handler.go:496-501`

`BeginReplay` (line 469) runs before `authorizeDocumentScope` (line 498). A replay match returns the cached finalization result without verifying the current caller owns the document. An attacker who observes a valid `Idempotency-Key` can receive the result (including `instanceId`) without authorization.

Fix: call `authorizeDocumentScope` before `BeginReplay`, or validate `actorForReplay` matches the current `userID` before serving a replay response.

---

### 5b-H2 — `GetFillInSchema` and `ListPlaceholderValues` have no authorization check

**File:** `internal/modules/documents/http/fillin_handler.go:41-60`

Neither handler performs a role or ownership check. Any authenticated session can retrieve the fill-in schema and all placeholder values for any document ID.

Fix: apply `authorizeDocumentScope` or an equivalent ownership/role gate before accessing the service.

---

### 5b-H3 — `reconstruct` and `view` endpoints have no rate limiting

**Files:** `internal/modules/documents/http/reconstruct_handler.go:30-49`, `view_handler.go:33-55`

`POST /api/v1/documents/{id}/reconstruct` triggers an expensive render/fanout. `GET /api/v1/documents/{id}/view` generates signed URLs. Neither is covered by rate-limit middleware.

Fix: register both routes through `ratelimit.Middleware`.

---

### 5b-H4 — `RegisterRoutesWithRateLimit` only rate-limits 2 of ~20 mutating routes

**File:** `internal/modules/documents/delivery/http/handler.go:131-165`

Method name implies full rate-limit coverage but only `autosave/presign` and `autosave/commit` receive it. `finalize`, `archive`, `duplicate`, `createCheckpoint`, `restoreCheckpoint`, `forceReleaseSession` are unprotected.

Fix: apply appropriate rate-limit buckets to at minimum `finalize`, `duplicate`, `createCheckpoint`, `restoreCheckpoint`.

---

### 5b-H5 — `exportDocxURL` not rate-limited (while `exportPDF` is)

**File:** `internal/modules/documents/delivery/http/export_handler.go:49`

`exportPDF` is rate-limited; `exportDocxURL` is registered with plain `HandleFunc`. Signed URL generation can be abused for URL harvesting.

Fix: apply same rate limit to docx-url route.

---

### 5b-H6 — `profileCode` leaked in 409 error response

**File:** `internal/modules/documents/delivery/http/handler.go:549-555`

```go
httpErr(w, http.StatusConflict, "no active approval route for profile "+profileCode)
```

`profileCode` is a business-internal DB-sourced identifier. Embedding it verbatim leaks internal approval taxonomy.

Fix: use static code `"no_active_approval_route"`; log `profileCode` server-side only.

---

### 5b-H7 — Raw SQL queries embedded in HTTP handler (`finalizeDocument`)

**File:** `internal/modules/documents/delivery/http/handler.go:503-565`

Three `QueryRowContext` calls execute SQL directly in the HTTP handler, bypassing the repository layer. They are invisible to mocks, untestable in isolation, and bypass tenant-scoping conventions.

Fix: extract into repository/application-layer methods.

---

### 5b-H8 — `SignExportURL` error swallowed with no log

**File:** `internal/modules/documents/delivery/http/export_handler.go:84`

Export succeeded (billed/stored); URL signing failed; client gets 500 `"internal"` with no server-side log. Impossible to distinguish from a full export failure in prod.

Fix: log `err` with `docID`, `tenantID`, storage key before responding.

---

### 5b-H9 — `CompleteReplay` failure leaves idempotency record in limbo

**File:** `internal/modules/documents/delivery/http/handler.go:584-586`

`idempReleased = true` is set regardless of whether `CompleteReplay` succeeds. If it fails, `FailReplay` is suppressed; the record is never completed and never failed; retries with the same key hit `ErrConflict` forever even though the underlying operation succeeded.

Fix: only set `idempReleased = true` when `CompleteReplay` returns `nil`; let the deferred `FailReplay` run on error.

---

### 5b-H10 — `documents.active_session_id` FK has no index → full table scan on every sweeper tick

**File:** `internal/modules/documents/repository/repository.go:667`

The `UPDATE documents SET active_session_id=NULL WHERE active_session_id IN (...)` in `ExpireStaleSessions` does a sequential scan on `documents` because `active_session_id` is not indexed.

Fix: `CREATE INDEX ON documents (active_session_id) WHERE active_session_id IS NOT NULL;`

---

### 5b-H11 — Sweeper integration tests permanently skipped

**File:** `internal/modules/documents/jobs/jobs_test.go:7-13`

`t.Skip("integration placeholder")` means neither `DeleteExpiredPending` nor `ExpireStaleSessions` has any test coverage. The unbounded-batch criticals above would have been caught by seeding rows and asserting affected counts.

Fix: implement tests; seed N expired sessions + pending uploads, call each sweeper, assert `RowsAffected` matches, assert non-expired rows untouched.

---

### 5b-H12 — `finalizeDocument` is ~178 lines with two divergent code paths

**File:** `internal/modules/documents/delivery/http/handler.go:413-590`

Legacy path (no submit service) and full approval path share the same handler body with a nil-check branch. The nil-check is not expressed in the type system — callers of `NewHandler` silently get the legacy path.

Fix: split into `Handler` (legacy) and `HandlerWithApproval` (full), or enforce the full constructor and remove `NewHandler`.

---

### 5b-H13 — `RegisterRoutes`/`RegisterRoutesWithRateLimit` duplicate ~20 route registrations

**File:** `internal/modules/documents/delivery/http/handler.go:101-165`

Copy-paste with no shared route table. When routes diverge (e.g., `autosave/presign` gains rate-limiting in one but not the other), omissions are silent.

Fix: extract a shared route table; apply rate-limiting as a decoration step.

---

### 5b-H14 — `exportDocxURL` issues two sequential service calls with no transactional consistency

**File:** `internal/modules/documents/delivery/http/export_handler.go:98-125`

`SignedDocxURL` + `GetDocumentSummary` called sequentially. If the document is deleted or revised between the two calls, the returned `revisionID` belongs to a different state than the URL.

Fix: have the service return both signed URL and revision ID in a single call.

---

### 5b-H15 — String-prefix match used for error classification instead of typed error

**File:** `internal/modules/documents/delivery/http/handler.go:1273`

```go
case strings.HasPrefix(err.Error(), "form_data_invalid"):
```

Any error wrapping breaks the prefix match silently; a coincidentally prefixed unrelated error is mis-classified.

Fix: define `ErrFormDataInvalid` sentinel and use `errors.Is`/`errors.As`.

---

### 5b-H16 — `pdfCompleteBody.PDFHash` zero-length passes validation → zero-byte hash written to DB

**File:** `internal/modules/documents/http/pdf_webhook_handler.go`

`body.PDFHash == ""` guard passes; `hex.DecodeString("")` succeeds returning `[]byte{}`. `WritePDF` persists a zero-length hash.

Fix: require `len(body.PDFHash) == 64` (sha256 hex); validate hex charset before calling `hex.DecodeString`.

---

### 5b-H17 — Webhook replay window is unbounded — no timestamp validation

**File:** `internal/modules/documents/http/pdf_webhook_handler.go:94-105`

HMAC is verified in constant time (`hmac.Equal`) but no timestamp is checked. A captured valid webhook payload can be replayed indefinitely.

Fix: add `X-Docgen-Timestamp` (Unix seconds) to the HMAC payload; reject requests where `abs(now - timestamp) > 5 minutes`.

---

### 5b-H18 — Sweeper loops use `log.Printf` with no structure and no backoff

**Files:** `internal/modules/documents/jobs/orphan_pending_sweeper.go:24`, `session_sweeper.go:23`

On repeated DB failures, one unstructured log line per tick. At 1-minute intervals: 1,440 identical lines per outage day. No tenant, no job ID, no error type, no rate-limiting.

Fix: use structured logger with `err` field; add consecutive-failure counter and exponential backoff or skip-on-N-consecutive.

---

### 5b-H19 — `timingOracle`: `requestID` fallback uses `time.Now().UnixNano()`

**File:** `internal/modules/documents/http/fillin_handler.go:209-214`

Nano-second timestamp is not unique under concurrent load and leaks server clock information. Echoed in error responses.

Fix: use `crypto/rand` UUID as fallback; or mandate `X-Request-ID` from caller.

---

## Medium

### 5b-M1 — `withAdminCtx` called in every handler (22+ sites) instead of middleware

**File:** `internal/modules/documents/delivery/http/handler.go` (multiple lines)

Cross-cutting concern repeated per-handler. Should be applied at route registration time as middleware wrapping.

---

### 5b-M2 — `documentDetailResponse` JSON keys are PascalCase (`"ID"`, `"TenantID"`) inconsistent with rest of API

**File:** `internal/modules/documents/delivery/http/handler.go:320-344`

Fix: standardise on `snake_case`; add json tags; align with API design system.

---

### 5b-M3 — `status` filter not validated against `DocumentStatus` enum

**File:** `internal/modules/documents/delivery/http/handler.go:263-275`

Arbitrary strings reach the repository SQL layer.

Fix: validate each status string against known `DocumentStatus` constants; return 400 for unknowns.

---

### 5b-M4 — `PaperSize` not validated in export handler

**File:** `internal/modules/documents/delivery/http/export_handler.go:18-21`

Arbitrary string flows into `domain.RenderOptions.PaperSize` and downstream into the Gotenberg integration.

Fix: validate against allowlist `["A4", "Letter", ...]` at handler boundary.

---

### 5b-M5 — `renameDocument` and `createCheckpoint` accept empty/unbounded string inputs

**Files:** `internal/modules/documents/delivery/http/handler.go:393-411` (name), `:924-947` (label)

No length or content validation at the HTTP layer.

Fix: validate non-empty + max length (255 chars) before forwarding to service.

---

### 5b-M6 — `FinalPDFS3Key` accepted without path-traversal validation

**File:** `internal/modules/documents/http/pdf_webhook_handler.go:65`

Body field passed to `WritePDF` as S3 key without format validation. Malformed key (e.g. `../`, null bytes) could cause path traversal.

Fix: validate against expected pattern (UUID-prefixed `.pdf`, no traversal sequences, max 512 chars).

---

### 5b-M7 — `HandleGetOptions` exposes full tenant user list without authorization

**File:** `internal/modules/documents/http/placeholder_options_handler.go:39-78`

No role or ownership check; `PHUser` branch returns `(user_id, display_name)` pairs for the entire tenant user base.

Fix: add ownership/role check; consider whether returning internal user IDs is necessary.

---

### 5b-M8 — `decodeJSON` Content-Type check allows bypass via empty header

**File:** `internal/modules/documents/http/fillin_handler.go:183-191`

Guard only fires when `Content-Type` is non-empty; empty header passes through.

Fix: require `Content-Type: application/json` unconditionally via `mime.ParseMediaType`.

---

### 5b-M9 — `decodeJSON` Content-Type error mapped to 500 via unknown path

**File:** `internal/modules/documents/http/fillin_handler.go`

Content-Type validation error goes through `mapFillInError` and falls to `internal.unknown` / 500 for what is a client error.

Fix: introduce typed `ErrBadContentType` and map to 415 Unsupported Media Type.

---

### 5b-M10 — `tenantID` variable shadows package-level `tenantID` function

**File:** `internal/modules/documents/http/fillin_handler.go:64-65`

Fix: rename local variable to `tid` consistently.

---

### 5b-M11 — `pdf_url` and `signed_url` are duplicates in view response

**File:** `internal/modules/documents/http/view_handler.go:49-54`

Both keys point to the same value. Duplicate semantics confuse API consumers.

Fix: pick one canonical key; deprecate the other.

---

### 5b-M12 — Idempotency handle leaked when auth fails after `BeginReplay`

**File:** `internal/modules/documents/delivery/http/handler.go:459-465`

Deferred `FailReplay` fires with `cause = nil` whenever `authorizeDocumentScope` fails after replay handle acquired. Record left in ambiguous state.

Fix: pass a meaningful error or release the handle before the auth-fail return.

---

### 5b-M13 — `json.Marshal` error ignored when recording idempotency replay body

**File:** `internal/modules/documents/delivery/http/handler.go:580-589`

```go
body, _ := json.Marshal(respBody)
```

`CompleteReplay` called with nil/empty body on marshal failure → corrupt replay record.

Fix: handle the error explicitly.

---

### 5b-M14 — Session audit logs include `session_id` and `pending_upload_id` (bearer-token-equivalent)

**File:** `internal/modules/documents/delivery/http/handler.go:806-812`

These IDs act as bearer tokens for the autosave flow. Logging them in full exposes them to any log aggregation system.

Fix: log only first 8 chars as a debug prefix.

---

### 5b-M15 — Two handler_problem_test.go functions permanently skipped

**File:** `internal/modules/documents/delivery/http/handler_problem_test.go:7-13`

`t.Skip` bodies for `problem+json` content-type contract tests.

Fix: implement or delete.

---

### 5b-M16 — Sweeper cutoff hardcoded as 24h literal

**File:** `internal/modules/documents/jobs/orphan_pending_sweeper.go`

`time.Now().Add(-24 * time.Hour)` hardcoded inside goroutine — untestable and non-configurable without rebuild.

Fix: accept `maxAge time.Duration` parameter from caller.

---

### 5b-M17 — `autosave_pending_uploads` has no `tenant_id` — no per-tenant attribution for sweeper

**File:** `internal/modules/documents/repository/repository.go:1035`

Sweeper deletes cross-tenant with no per-tenant observability or rate limiting. A noisy tenant blocks all others.

Recommend: add `tenant_id` column (nullable, backfill via `session_id → editor_sessions → documents.tenant_id`), then index `(tenant_id, expires_at) WHERE consumed_at IS NULL`.

---

### 5b-M18 — `DeleteExpiredPending` runs outside a transaction — inconsistent with `ExpireStaleSessions`

**File:** `internal/modules/documents/repository/repository.go:1035`

`ExpireStaleSessions` uses `authz.BypassSystem` + transaction; `DeleteExpiredPending` uses `ExecContext` directly with no bypass and no tx. Asymmetric authz setup; may fail silently if RLS is ever tightened.

Fix: wrap in transaction with `authz.BypassSystem` or document explicit RLS exemption.

---

### 5b-M19 — Session sweeper passes tick delivery time instead of `time.Now()` as `now`

**File:** `internal/modules/documents/jobs/session_sweeper.go:21`

Under Go scheduler pressure, tick delivery can lag by seconds. Sessions expiring in that window are missed until next tick.

Fix: use `time.Now()` at query execution time, or let Postgres use `now()` in the CTE.

---

### 5b-M20 — `Service` interface has 23 methods — violates small-interface principle

**File:** `internal/modules/documents/delivery/http/handler.go:35-63`

Fix: split into focused sub-interfaces (`SessionService`, `CheckpointService`, `CommentService`); compose in handler struct.

---

### 5b-M21 — `reconstruct` and `view` default error branches log nothing

**Files:** `internal/modules/documents/http/reconstruct_handler.go:51-59`, `view_handler.go:57-66`

All non-sentinel errors produce 500 with no log line — silent for ops.

Fix: add structured log on default branch with tenant, doc ID, raw error.

---

### 5b-M22 — Content-hash query uses `created_at DESC` without secondary tie-breaker

**File:** `internal/modules/documents/delivery/http/handler.go:558-565`

Non-deterministic result if two revisions share the same `created_at` timestamp.

Fix: add `id DESC` (or sequence column) as secondary sort key.

---

## Low

### 5b-L1 — `roleTemplateAuthor` declared but never used in the file

**File:** `internal/modules/documents/delivery/http/handler.go:29-32`

Fix: remove or import from canonical roles package.

---

### 5b-L2 — Role constants defined locally in HTTP package; divergence from IAM domain undetectable at compile time

**File:** `internal/modules/documents/delivery/http/handler.go:29-32`

Fix: import canonical role constants from `iamdomain` or a shared `roles` package.

---

### 5b-L3 — `writeJSON` package-level alias is dead code

**File:** `internal/modules/documents/delivery/http/handler.go:83-84`

`var writeJSON = httpresponse.WriteJSON` defined but all call sites use `httpresponse.WriteJSON` directly.

Fix: remove the alias.

---

### 5b-L4 — `log` (stdlib) used throughout; inconsistent with structured platform logger

**File:** `internal/modules/documents/delivery/http/handler.go:8` (and multiple call sites)

Fix: inject `*slog.Logger` via constructor.

---

### 5b-L5 — `generated_adapter_test.go` in package `http` (internal) while other test files use `http_test` (external)

**File:** `internal/modules/documents/delivery/http/generated_adapter_test.go`

Fix: move to `http_test` package for consistency.

---

### 5b-L6 — `handler_test.go:663-688` parses source file text at test runtime

**File:** `internal/modules/documents/delivery/http/handler_test.go:663-688`

`os.ReadFile("handler.go")` + `strings.Index` to assert structural separation. Fragile to cosmetic refactors and CI working-directory assumptions.

Fix: enforce the separation via type-level or integration-level test.

---

### 5b-L7 — `withAuthHeaders` in tests dereferences `*http.Request` pointer to copy

**File:** `internal/modules/documents/delivery/http/handler_test.go:282-286`

`*req = *req.WithContext(...)` is unusual and fragile for structs containing interface fields.

Fix: return the modified `*http.Request` normally.

---

### 5b-L8 — `listDocuments` pagination envelope written as inline `map[string]any` literal

**File:** `internal/modules/documents/delivery/http/handler.go:182-198`

Same shape repeated across multiple list handlers with no shared envelope struct.

Fix: define `paginatedResponse[T]` struct or shared helper.

---

### 5b-L9 — `HandleGetOptions` uses `revisionID` for the `{id}` path param which is everywhere else a document ID

**File:** `internal/modules/documents/http/placeholder_options_handler.go:39`

Misleading local variable name.

Fix: rename to `docID`.

---

### 5b-L10 — `WritePDF` error (webhook) has no server-side log

**File:** `internal/modules/documents/http/pdf_webhook_handler.go:83`

500 `"persist_failed"` returned with no log. Retry storms are invisible.

Fix: log `err` with `docID`, `tenantID`, `s3Key` before returning 500.

---

### 5b-L11 — `generated_adapter.go` discards all typed OpenAPI parameters silently

**File:** `internal/modules/documents/delivery/http/generated_adapter.go:20-143`

Every method discards typed params and forwards to legacy mux. OpenAPI contract enforcement is completely bypassed; divergence between spec and legacy parsing is undetectable.

Fix: document explicitly as intentional ACL bridge; add integration test that routes through adapter to verify path extraction parity.

---

### 5b-L12 — `sweeper` functions accept concrete `*repository.Repository` instead of narrow interface

**Files:** `internal/modules/documents/jobs/*.go`

Violates "accept interfaces" principle; prevents testing without live repository.

Fix: define narrow sweeper interfaces (`DeleteExpiredPending(ctx, cutoff) (int64, error)` etc.).

---

## Critical Backlog — Fix Branches

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5b-C3 | `delivery/http/handler.go:1165,1187` X-User-Roles header trusted → privilege escalation | Critical | leandrotca | TBC | `fix/docs-5b-header-roles-c3` | Backlog (land first) |
| 5b-C4 | `http/pdf_webhook_handler.go:83` tenant_id from body → cross-tenant PDF overwrite | Critical | leandrotca | TBC | `fix/docs-5b-webhook-tenant-c4` | Backlog (land second) |
| 5b-C1 | `delivery/http/handler.go:559-564` content-hash DB error silently discarded | Critical | leandrotca | TBC | `fix/docs-5b-finalize-c1-c2` | Backlog |
| 5b-C2 | `delivery/http/handler.go:618-622` err.Error() in JSON response → info leak + JSON injection | Critical | leandrotca | TBC | `fix/docs-5b-finalize-c1-c2` | Backlog |
| 5b-C5 | `repository/repository.go:1035` unbounded DELETE in DeleteExpiredPending | Critical | leandrotca | TBC | `fix/docs-5b-sweeper-unbounded-c5-c6` | Backlog |
| 5b-C6 | `repository/repository.go:667` unbounded UPDATE in ExpireStaleSessions | Critical | leandrotca | TBC | `fix/docs-5b-sweeper-unbounded-c5-c6` | Backlog |
