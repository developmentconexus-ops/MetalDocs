# Module #13 — Shared Infra (`internal/test`, `internal/testsupport`)

**Reviewed:** 2026-05-22
**Reviewer:** direct read (small scope, no agents needed)
**Severity summary:** 0 Critical · 0 High · 4 Medium · 5 Low

## Files reviewed

| File | LoC | Notes |
|------|-----|-------|
| `internal/test/e2e_seed.go` | ~360 | E2E test harness routes + seed helpers |
| `internal/test/e2e_clock_offset_nonprod.go` | ~30 | Clock-override API (non-prod build) |
| `internal/test/e2e_clock_offset_production.go` | ~20 | No-op stubs (prod build) |
| `internal/testsupport/http/auth_request.go` | ~30 | Test auth helper |
| `internal/testsupport/pgtest/pgtest.go` | ~40 | Postgres test helper |

## Summary

This is test/test-support code only. No production security surface — `e2e_seed.go` requires `METALDOCS_E2E=1` at both registration time and inside each handler. Clock offset files are build-tagged correctly. `testsupport/` helpers are clean. Findings are quality and robustness observations, not security risks.

---

## Medium

### M1 — `e2e_seed.go` has no build tag; compiled into every binary

**File:** `internal/test/e2e_seed.go`
**Line:** top of file (no `//go:build` directive)

`RegisterE2EHandlers` is compiled unconditionally into production, staging, and CI binaries. The only gate is an env-var check:

```go
if os.Getenv("METALDOCS_E2E") != "1" {
    return
}
```

The check appears at registration time AND inside each handler, so the risk is low in practice. But if `METALDOCS_E2E=1` is accidentally set in a production environment (e.g., via a misconfigured secret or copied `.env`), five destructive endpoints become live:

- `POST /internal/test/reset` — runs 14 DELETE statements against the live DB
- `POST /internal/test/seed` — upserts users and tenant rows
- `POST /internal/test/advance-clock` — shifts the application clock
- `GET /internal/test/governance-events` — exposes governance event data
- `POST /internal/test/trigger-scheduler-tick` — runs the scheduler

**Recommend:** add `//go:build !production` to `e2e_seed.go`, mirroring `e2e_clock_offset_nonprod.go`. The clock files already demonstrate the correct pattern. This eliminates the entire risk class without changing runtime behavior.

---

### M2 — `writeJSON` discards encode error silently

**File:** `internal/test/e2e_seed.go`
**Line:** `writeJSON` helper (search for `_ = json.NewEncoder`)

```go
func writeJSON(w http.ResponseWriter, status int, payload any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}
```

Encode errors are discarded. In E2E context this is a test reliability concern: a serialization failure produces a truncated/empty response body, the test gets `200 OK` with no body, and debugging becomes painful.

**Recommend:** log the error with `log.Printf("writeJSON encode: %v", err)` or return it to callers so test failures surface clearly.

---

### M3 — `isUndefinedTable`/`isUndefinedColumn` use string-matching instead of pgcode constants

**File:** `internal/test/e2e_seed.go`
**Lines:** `isUndefinedTable`, `isUndefinedColumn` helpers

```go
func isUndefinedTable(err error) bool {
    return strings.Contains(err.Error(), "sqlstate 42p01")
}
func isUndefinedColumn(err error) bool {
    return strings.Contains(err.Error(), "sqlstate 42703")
}
```

String-matching on driver error messages is fragile across `pgx` versions. The `github.com/jackc/pgx/v5/pgconn` package exposes `(*pgconn.PgError).Code` with typed constants from `github.com/jackc/pgx/v5/pgxconn` or the `pgerrcode` package.

**Recommend:**
```go
import "github.com/jackc/pgx/v5/pgconn"

func isUndefinedTable(err error) bool {
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}
```

---

### M4 — `reset` handler surfaces raw DB error message to HTTP caller

**File:** `internal/test/e2e_seed.go`
**Lines:** `reset` handler, error-reporting path

When a non-schema error occurs mid-sequence in the reset loop, the handler returns `500` with `execErr.Error()` in the response body. In E2E the caller is a test harness (low concern), but the same pattern leaks Postgres error details (table names, constraint names, schema structure) to anyone who can reach the endpoint.

**Recommend:** In E2E context, detailed errors are acceptable for debuggability. Add a comment: `// Error details intentional — E2E-only endpoint.` If M1 is addressed (build tag), this surfaces only in non-prod, which removes the concern entirely.

---

## Low

### L1 — `e2ePassword` hardcoded without explanatory comment

**File:** `internal/test/e2e_seed.go`
**Line:** `const e2ePassword = "test1234"`

Not a risk (E2E only), but the constant has no comment explaining why it's intentional and not a secret management gap. Future reviewers may flag it.

**Recommend:** add `// intentional static value — test harness only, never used in production`.

---

### L2 — `mapRoleToMembership` has no explicit "admin" case

**File:** `internal/test/e2e_seed.go`
**Lines:** `mapRoleToMembership` function

```go
func mapRoleToMembership(role string) string {
    switch role {
    case "viewer":
        return "viewer"
    case "editor":
        return "editor"
    default:
        return "approver"
    }
}
```

The `admin` role (used in `upsertSeedUser` calls) silently falls through to `approver`. E2E admin users are seeded with `approver` membership, which may or may not match test expectations.

**Recommend:** add an explicit `case "admin": return "admin"` and default to a panic or error return so unknown roles don't silently produce incorrect seed data.

---

### L3 & L4 — Deterministic user ID truncation can collide on UUID tenant IDs

**File:** `internal/test/e2e_seed.go`
**Lines:** `upsertSeedUser`, `sanitizeSlug`

```go
id := fmt.Sprintf("e2e-%s-%s", role, sanitizeSlug(tenantID))
// sanitizeSlug strips [-_] then truncates to 12 chars
```

UUID tenantIDs like `550e8400-e29b-41d4-a716-446655440000` all produce the same 12-char prefix after stripping `-`. Two different tenants → same truncated slug → same user ID → upsert silently overwrites the wrong tenant's seed user.

**Recommend:** use a hash (e.g. `fmt.Sprintf("%x", sha256sum[:4])`) of the full tenantID instead of a truncated slug, or include the full tenantID without truncation if IDs are short enough.

---

### L5 — `triggerSchedulerTick` fallback sleep is undocumented

**File:** `internal/test/e2e_seed.go`
**Lines:** `triggerSchedulerTick` handler

```go
if runSchedulerTick == nil {
    time.Sleep(6 * time.Second)
    // responds 200 OK
}
```

When `runSchedulerTick` is not wired, the handler sleeps 6 seconds and returns success. E2E tests waiting on scheduler effects will see a 6-second delay with no indication the tick didn't actually run.

**Recommend:** return a `503` with body `{"error":"scheduler not wired"}` instead of sleeping. If the sleep is intentional (polling wait), add a comment explaining the rationale and the expected caller contract.

---

## Skipped / Clean

| File | Verdict |
|------|---------|
| `internal/test/e2e_clock_offset_nonprod.go` | Clean — atomic ops, no side effects, build-tagged correctly |
| `internal/test/e2e_clock_offset_production.go` | Clean — no-op stubs, build-tagged `production` |
| `internal/testsupport/http/auth_request.go` | Clean — mirrors production IAM middleware correctly |
| `internal/testsupport/pgtest/pgtest.go` | Clean — env-skipped, cleanup registered, timeout handled |
