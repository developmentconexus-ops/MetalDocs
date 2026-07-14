# Module #10 Review — `internal/modules/render`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:silent-failure-hunter
**Severity totals:** 0 Critical / 7 High / 12 Medium / 8 Low
**Files reviewed:**
- `fanout/{client,pdf_dispatcher,pdf_dispatch_adapter,reconstruction,pdf_outbox_repository,pdf_outbox_worker}.go`
- `resolvers/{resolver,registry,builtins,hash,approval_date,approvers,author,doc_code,doc_title,effective_date,revision_number,controlled_by_area}.go`

---

## Critical

None.

---

## High

### H1 — `fanout/pdf_outbox_repository.go:72` — `MarkDispatched` discards `RowsAffected` → silent lost dispatch acknowledgement

`MarkDispatched` calls `ExecContext` and ignores the result. If the UPDATE matches zero rows (row deleted, ID wrong), the function returns `nil` — the worker believes dispatch was recorded when it was not. The outbox row stays in `processing` state indefinitely.

**Recommend:** `res, err := r.db.ExecContext(...); if err != nil { return err }; if n, _ := res.RowsAffected(); n == 0 { return fmt.Errorf("mark dispatched: row %s not found", id) }`.

---

### H2 — `fanout/pdf_outbox_repository.go:80` — `MarkFailed` discards `RowsAffected` → same silent no-op

Both branches of `MarkFailed` discard the `sql.Result`. Same failure mode as H1.

**Recommend:** same fix — check `RowsAffected() == 0` and return a descriptive error.

---

### H3 — `fanout/pdf_outbox_worker.go:88` — backoff duration overflow on high `Attempts`

```go
time.Duration(1 << r.Attempts) * 30 * time.Second
```

Left-shift on `int` overflows when `r.Attempts >= 33` (64-bit) or even lower edge cases, producing a negative duration. `math.Min` on the resulting negative `float64` may clamp correctly by accident but produces undefined behaviour.

**Recommend:** clamp before shifting: `shift := r.Attempts; if shift > 20 { shift = 20 }; backoff := time.Duration(1<<shift) * 30 * time.Second; if backoff > 30*time.Minute { backoff = 30*time.Minute }`. Eliminate the `float64` round-trip.

---

### H4 — `resolvers/registry.go:3` — `Registry` has no mutex → data race on concurrent `Register`/`Get`

`Register` and `Get` write/read the internal map with no synchronization. Concurrent calls from multiple goroutines during resolver registration at startup produce a data race.

**Recommend:** embed `sync.RWMutex`; use `RLock` in `Get`/`Known`, `Lock` in `Register`.

---

### H5 — `resolvers/resolver.go:8` — `ResolveInput` no constructor; nil readers cause runtime panics

All resolver `Resolve` implementations dereference `in.WorkflowReader`, `in.RevisionReader`, etc. without nil checks. A nil reader causes an opaque panic at runtime. No constructor enforces non-nil.

**Recommend:** add `NewResolveInput(tenantID, revisionID string, ...) (ResolveInput, error)` that validates required fields non-nil and non-empty. Add per-resolver guard: `if in.WorkflowReader == nil { return ResolvedValue{}, fmt.Errorf("approval_date: WorkflowReader is nil") }`.

---

### H6 — `resolvers/approval_date.go:15` — zero `time.Time` formatted as `"0001-01-01"` → silent wrong render value

When `GetFinalApprovalDate` returns a zero `time.Time` (approval not yet completed), the resolver formats it as `"0001-01-01"` and returns it as a valid resolved value. Document consumers receive a nonsensical date with no error.

**Recommend:** mirror `effective_date.go`'s zero-check: `if approvalDate.IsZero() { value = "" }`.

---

### H7 — `resolvers/author.go:32` — `Value` set to `AuthorInfo` struct, not `string` → type assertion fails at render time

All other resolvers place a `string` in `ResolvedValue.Value`. `author.go` places the full `AuthorInfo` struct. Callers expecting a string will receive a struct inside `any`; type assertion silently fails at render time producing a blank author field.

**Recommend:** change to `Value: author.DisplayName` (or a formatted string). Update `ResolvedValue` contract if structured data is intentional.

---

## Medium

### M1 — `resolvers/` (all files) — no `TenantID`/`RevisionID` validation before passing to readers

No resolver validates `in.TenantID != ""` or `in.RevisionID != ""` before calling DB readers. An empty `TenantID` produces a cross-tenant or unbounded query depending on the reader implementation.

**Recommend:** validate in `NewResolveInput` (see H5) or add a check at the top of each `Resolve`: `if in.TenantID == "" { return ResolvedValue{}, fmt.Errorf("%s: TenantID is required", r.Key()) }`.

---

### M2 — `resolvers/registry.go:13` — `Register` silently overwrites existing key

Duplicate registration produces no error, no log. A mis-wired builtin at startup is invisible.

**Recommend:** return an error or panic on duplicate key: `if _, exists := r.m[key]; exists { panic("resolvers: duplicate registration: " + key) }`.

---

### M3 — `resolvers/resolver.go:9` — `TenantID`, `RevisionID`, `ControlledDocumentID` are bare `string`

Silently swappable at call sites.

**Recommend:** `type TenantID string`, `type RevisionID string`, `type DocumentID string`.

---

### M4 — `resolvers/approvers.go:33` — hash input omits `ApprovalInstanceID` → cache collision across approval rounds

Two calls with different `ApprovalInstanceID` values produce identical hashes, causing incorrect cache hits when multiple approval rounds exist.

**Recommend:** add `ApprovalInstanceID string \`json:"approval_instance_id"\`` to the hash struct.

---

### M5 — `resolvers/author.go` — `ResolvedValue.Value` type contract is implicit

No documented type contract on `Value any`. Every resolver places a `string`; `author.go` breaks this (see H7). Without a typed contract, future resolvers will diverge silently.

**Recommend:** define `type ResolvedValue struct { Value string }` (typed) or add a `// Value is always a string` doc contract enforced by tests.

---

### M6 — `fanout/pdf_outbox_repository.go:43` — `ClaimPending` has no tenant filter (cross-tenant ownership undocumented)

The worker claims rows across ALL tenants. This is likely intentional for a global background worker, but the absence of any comment makes it indistinguishable from a missing predicate.

**Recommend:** add a comment: `// ClaimPending is intentionally cross-tenant — the outbox worker owns all tenants.`

---

### M7 — `fanout/reconstruction.go:78` — `resp.ContentHash` compared as raw string without hex decoding

Comparison relies on the fanout service always returning lowercase hex. Uppercase or alternate encoding silently returns `MatchesOriginal() == false` with no error.

**Recommend:** decode `resp.ContentHash` from hex into `[]byte` and use `bytes.Equal`, or document and enforce the format contract.

---

### M8 — `fanout/pdf_outbox_repository.go:119` — interval injected as formatted millisecond string → fragile SQL

`fmt.Sprintf("%d milliseconds", ...)` string-formats the interval into the query. Non-standard and fragile.

**Recommend:** pass `olderThan.Milliseconds()` as a plain `int64` arg and use `($1 * interval '1 millisecond')` in SQL, or use `pgtype` interval encoding.

---

### M9 — `resolvers/controlled_by_area.go:14` — empty snapshots return empty string silently

When both `AreaNameSnapshot` and `AreaCodeSnapshot` are empty, resolver returns `""` with nil error. Render pipeline cannot distinguish "area not populated" from "area genuinely unnamed."

**Recommend:** return a sentinel error or a distinct `ResolvedValue{Missing: true}` field when both snapshots are empty.

---

### M10 — `fanout/client.go:59` — error response body not size-capped before `io.ReadAll`

A misbehaving or malicious fanout service can send an unbounded error body, exhausting heap.

**Recommend:** `io.LimitReader(resp.Body, 64*1024)` before `io.ReadAll`.

---

### M11 — `fanout/pdf_dispatcher.go:35` + `pdf_outbox_worker.go:74` — idempotency key not tenant-scoped

Key is `"docgen_v2_pdf:" + revisionID` without `tenantID`. Shared key space across tenants (negligible UUID collision risk but poor hygiene).

**Recommend:** `"docgen_v2_pdf:" + tenantID + ":" + revisionID`.

---

### M12 — `resolvers/resolver.go:48` — `GetFinalApprovalDate` lacks `approvalInstanceID` parameter

`approval_date.go` calls `GetFinalApprovalDate` without an instance ID. With multiple approval rounds, the wrong date may be silently returned.

**Recommend:** add `approvalInstanceID` to the `WorkflowReader.GetFinalApprovalDate` signature, or document the deliberate single-instance assumption.

---

## Low

### L1 — `fanout/client.go:33` — `http.DefaultClient` fallback has no timeout

Callers passing `nil` inherit a client with no deadline — violates Go HTTP security practice.

**Recommend:** default to `&http.Client{Timeout: 30 * time.Second}`.

---

### L2 — `fanout/client.go:45` — URL assembled by raw string concatenation

Double-slash if `baseURL` has trailing slash.

**Recommend:** `strings.TrimRight(baseURL, "/")` in constructor, or `url.JoinPath`.

---

### L3 — `fanout/pdf_outbox_worker.go:55` — `ResetStaleClaims` count discarded with `_`

Operational visibility gap — stale-claim resets are silent.

**Recommend:** `if n > 0 { slog.Info("reset stale outbox claims", "count", n) }`.

---

### L4 — `resolvers/approval_date.go:31` + 4 other resolver files — `ResolverKey`/`ResolverVer` literals duplicate `Key()`/`Version()` return values

A version bump in one place silently diverges from stored metadata.

**Recommend:** use the method return value or a package-level const.

---

### L5 — `resolvers/approvers.go:27` — empty `DisplayName` silently dropped with no log

All approvers with empty names degrade to placeholder string silently.

**Recommend:** `slog.Warn("approver has empty DisplayName", "revision_id", in.RevisionID, "count", emptyCount)`.

---

### L6 — `resolvers/controlled_by_area.go:22` — hash field tagged `area_code_fallback` but source field is `AreaCodeSnapshot`

Asymmetric naming causes confusion and silent hash mismatch risk on rename.

**Recommend:** align JSON tag: `` `json:"area_code_snapshot"` ``.

---

### L7 — `resolvers/builtins.go:4` — `RegisterBuiltins(nil)` panics with opaque nil-deref

**Recommend:** `if r == nil { panic("resolvers: RegisterBuiltins called with nil registry") }`.

---

### L8 — `fanout/pdf_outbox_repository.go:114` — `ResetStaleClaims` returns raw error without wrapping

All other methods in the file wrap errors; this one does not.

**Recommend:** `return 0, fmt.Errorf("reset stale claims: %w", err)`.

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/render-10-outbox-h1-h2` | H1 MarkDispatched + H2 MarkFailed RowsAffected gaps | 1st |
| `fix/render-10-backoff-h3` | H3 backoff overflow | 2nd |
| `fix/render-10-registry-h4` | H4 Registry mutex | 3rd |
| `fix/render-10-resolvers-h5-h6-h7` | H5 ResolveInput constructor + H6 approval_date zero time + H7 author struct in Value | 4th |
