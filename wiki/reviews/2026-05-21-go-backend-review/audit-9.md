# Module #9 Review — `internal/modules/audit`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:silent-failure-hunter
**Severity totals:** 3 Critical / 5 High / 7 Medium / 4 Low
**Files reviewed:**
- `domain/port.go`
- `application/service.go`
- `delivery/http/handler.go`
- `infrastructure/postgres/writer.go`
- `infrastructure/memory/writer.go`

---

## Critical

### C1 — `delivery/http/handler.go:39` — no authentication or authorization gate → unauthenticated read of full audit log

`handleEvents` has no session check and no `authz.Require` call. Any unauthenticated HTTP request to `/api/v1/audit/events` receives audit log data. The IAM module defines `CapAuditRead` for exactly this purpose.

**Recommend:** extract tenant from authenticated session context (same pattern as documents/iam modules). Call `authz.Require(ctx, tx, string(iamdomain.CapAuditRead), "tenant")` before delegating to the service. Return HTTP 401/403 on failure.

**Fix branch:** `fix/audit-9-authz-idor-c1-c2-c3` (land first)

---

### C2 — `delivery/http/handler.go:55` — `TenantID` never extracted from context → cross-tenant IDOR

The handler builds `domain.ListEventsQuery{}` without populating `TenantID`. The Postgres writer's WHERE clause is `($3 = '' OR tenant_id = $3)` — with an empty string it always evaluates true, returning audit events across all tenants to any caller.

**Recommend:** extract the authenticated tenant ID from context (e.g. `tenant.FromContext(r.Context())`), set it on the query, and reject the request with HTTP 403 if the claim is absent.

**Fix branch:** `fix/audit-9-authz-idor-c1-c2-c3`

---

### C3 — `application/service.go:22-27` — `TenantID` dropped during normalization → zeroed before reaching DB

Even if C2 is fixed independently, the service zeroes the tenant ID:
```go
normalized := domain.ListEventsQuery{
    ResourceType: query.ResourceType,
    ResourceID:   query.ResourceID,
    Limit:        query.Limit,
    // TenantID not copied
}
```

`TenantID` is silently omitted from the normalized struct. Any tenant scoping set by the handler is discarded before the repository call.

**Recommend:** add `TenantID: query.TenantID` to the normalized struct literal. Enforce `query.TenantID != ""` before proceeding — return an error if empty.

**Fix branch:** `fix/audit-9-authz-idor-c1-c2-c3`

---

## High

### H1 — `application/service.go:19` — nil-receiver guard + silent empty return on nil `reader` masks misconfiguration

`if s == nil` on a method is unreachable — a nil receiver panics before the method body. `if s.reader == nil` silently returns `(nil, nil)` — callers see zero results with no error, making a misconfigured DI graph invisible at runtime.

**Recommend:** remove the nil-receiver guard. In `NewService`, validate `reader != nil` and panic/return error immediately: `if reader == nil { panic("audit: reader is required") }`.

---

### H2 — `delivery/http/handler.go:69` — `json.Unmarshal` error discarded → silent payload corruption in response

```go
_ = json.Unmarshal([]byte(e.PayloadJSON), &payload)
```

On a malformed `PayloadJSON` (corrupt audit row, partial write, schema drift), `payload` stays as an empty map. The response silently loses payload data; callers cannot distinguish "no payload" from "corrupted payload."

**Recommend:** capture the error; log at warn level with the event ID; return a sentinel `"_parse_error": true` field or omit payload with an explicit null so callers know data is absent.

---

### H3 — `infrastructure/postgres/writer.go:62` — `RecordTx` does not check `RowsAffected`

The INSERT result is not inspected. If the row is silently suppressed (future `ON CONFLICT DO NOTHING`, trigger, or policy), the caller receives `nil` error and believes the event was recorded.

**Recommend:** `n, err := result.RowsAffected(); if err != nil { return err }; if n == 0 { return errors.New("audit: event not recorded") }`.

---

### H4 — `infrastructure/postgres/writer.go:71` — `ValidateIntegrity` unbounded cross-tenant table scan

`ValidateIntegrity` performs a full table scan with no tenant scope, no LIMIT, and no transaction — reads uncommitted rows from concurrent writers. On a large audit table this causes unbounded memory allocation and lock contention.

**Recommend:** add `LIMIT` (e.g. 10 000 rows per call); tenant-scope when context is available; wrap in `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ`.

---

### H5 — `infrastructure/memory/writer.go:28` — `RecordTx` ignores `*sql.Tx` → false transactional semantics in tests

`RecordTx` calls `Record` directly, ignoring the `tx` parameter. If the surrounding transaction rolls back, the event is already committed to the in-memory slice. Tests exercising rollback paths receive false positives — events appear recorded when they should not.

**Recommend:** either panic with `"memory writer does not support RecordTx"` to make misuse loud, or buffer the event against the tx and apply it only on commit notification.

---

## Medium

### M1 — `domain/port.go:45` — `Writer` interface exposes `*sql.Tx` → infrastructure type leaks into domain port

`RecordTx(ctx, tx *sql.Tx, event Event)` forces every non-Postgres `Writer` implementation to accept a `*sql.Tx` it cannot meaningfully use (confirmed by the memory writer ignoring it).

**Recommend:** move `RecordTx` to an infrastructure-only interface (`postgres.TransactionalWriter`). Keep the domain port to `Record` only.

---

### M2 — `domain/port.go:9` — `Event` exported struct, no constructor; all invariants bypassable

Callers can construct `domain.Event{}` with empty `ID`, blank `TenantID`, or zero `OccurredAt`, inserting a corrupt row that silently breaks the hash chain.

**Recommend:** add `NewEvent(tenantID, actorID, action, resourceType, resourceID string, now time.Time) (Event, error)` that validates required fields.

---

### M3 — `infrastructure/postgres/writer.go:27` — `defer tx.Rollback()` error discarded

A rollback error means the connection may be in an unknown state. Silently dropped.

**Recommend:** `if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) { slog.Error("audit: rollback failed", "err", rbErr) }`.

---

### M4 — `delivery/http/handler.go:49,61` — `problem.Write(...)` error discarded with `_ =`

On both the bad-request and internal-error paths, a failed response write is invisible. Client receives no status, server has no signal.

**Recommend:** log write errors at warn level: `if err := problem.Write(...); err != nil { slog.Warn("audit: write response failed", "err", err) }`.

---

### M5 — `delivery/http/handler.go:16` — `Handler` holds concrete `*application.Service`

Couples HTTP layer to concrete type; unit testing requires full service graph.

**Recommend:** define local `eventLister interface { ListEvents(ctx, query) ([]Event, error) }` and accept that in `NewHandler`.

---

### M6 — `infrastructure/memory/writer.go:63` — `ValidateIntegrity` always returns `nil, nil`

Any caller using the memory writer for integrity testing gets a permanently clean result, hiding bugs in integrity-check logic.

**Recommend:** return `nil, errors.New("integrity validation not supported by memory writer")` so callers fail fast.

---

### M7 — `infrastructure/postgres/writer.go:126` — `ListEvents` reachable only via concrete `*postgres.Writer`

`postgres.Writer` implements `domain.Reader` but the type declaration doesn't declare it. Callers that hold `domain.Writer` cannot call `ListEvents` without a type assertion — invisible coupling.

**Recommend:** split into `postgres.Writer` (implements `domain.Writer`) and `postgres.Reader` (implements `domain.Reader`), or add a compile-time assertion `var _ domain.Reader = (*Writer)(nil)`.

---

## Low

### L1 — `domain/port.go:9` — `ActorID`, `TenantID`, `Action`, `ResourceType`, `ResourceID` are bare `string`

Silently swappable at call sites; compiler provides no protection.

**Recommend:** `type ActorID string`, `type TenantID string` — at minimum for the most commonly confused pair.

---

### L2 — `infrastructure/memory/writer.go:46` — `ResourceType`/`ResourceID` use `EqualFold`; `TenantID` uses `==`

Inconsistent filter semantics between memory and Postgres implementations. Tests pass for case-insensitive `ResourceType` values that production Postgres would reject.

**Recommend:** use exact `==` for all three fields in both implementations.

---

### L3 — `infrastructure/postgres/writer.go:22` — `Record` convenience wrapper undocumented re: nested transactions

If called inside an existing transaction, `Record` starts a nested `BEGIN`, which is invalid in Postgres and produces an error.

**Recommend:** add doc comment: "Record must not be called inside an existing transaction; use RecordTx instead."

---

### L4 — No compile-time interface assertions anywhere in the module

Neither `postgres.Writer` nor `memory.Writer` has `var _ domain.Writer = (*Writer)(nil)`.

**Recommend:** add assertions in both infrastructure packages.

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/audit-9-authz-idor-c1-c2-c3` | C1 no authz gate + C2 no tenant in query + C3 tenant dropped in service | 1st (single chain, land together) |
