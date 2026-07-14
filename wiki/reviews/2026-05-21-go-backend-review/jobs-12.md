# Module #12 Review — `internal/modules/jobs`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer
**Severity totals:** 1 Critical / 4 High / 10 Medium / 6 Low
**Files reviewed:**
- `scheduler/scheduler.go`, `scheduler/lease_reaper.go`
- `audit_integrity_validator/job.go`
- `idempotency_janitor/job.go`
- `stuck_instance_watchdog/job.go`

---

## Critical

### C1 — `scheduler/lease_reaper.go:50` — `governance_events` INSERT uses hardcoded `tenant_id = 'system'`

The governance event for a reaped lease is inserted with `tenant_id = 'system'` as a string literal. If `governance_events.tenant_id` has a foreign key to a `tenants` table, this fails the FK constraint at runtime. If downstream consumers filter by `tenant_id`, system-level events are invisible to tenant-scoped queries. No schema comment or constant documents that `'system'` is a valid sentinel.

**Recommend:** verify the schema accepts `'system'` as a sentinel and document it explicitly with a `// system tenant sentinel — see migrations/...` comment. Or use a dedicated system tenant UUID constant defined in a shared domain package. Confirm FK constraint behavior.

**Fix branch:** `fix/jobs-12-reaper-c1`

---

## High

### H1 — `scheduler/lease_reaper.go:18` — `FOR UPDATE SKIP LOCKED` subquery race with outer `DELETE`

```sql
DELETE FROM metaldocs.job_leases
WHERE job_name IN (
    SELECT job_name FROM metaldocs.job_leases
    WHERE expires_at < now() - interval '10 minutes'
    FOR UPDATE SKIP LOCKED
)
```

In PostgreSQL the outer `DELETE` does not inherit the row-level lock from the subquery; concurrent reapers can delete overlapping sets, causing double-delete errors or missed rows.

**Recommend:** rewrite as a CTE:
```sql
WITH locked AS (
    SELECT job_name FROM metaldocs.job_leases
    WHERE expires_at < now() - interval '10 minutes'
    FOR UPDATE SKIP LOCKED
)
DELETE FROM metaldocs.job_leases
WHERE job_name IN (SELECT job_name FROM locked)
RETURNING job_name
```

**Fix branch:** `fix/jobs-12-reaper-c1` (same branch — related correctness cluster)

---

### H2 — `stuck_instance_watchdog/job.go:43` — `listStuckInstances` error swallowed; returns `nil` instead of error

When `listStuckInstances` fails, the function returns `nil` (success). The scheduler sees a successful tick; stuck instances are never processed; the failure is invisible.

**Recommend:** `return fmt.Errorf("stuck_instance_watchdog: list: %w", err)` so the scheduler can log it and apply backpressure.

**Fix branch:** `fix/jobs-12-watchdog-h2-h3-h4`

---

### H3 — `stuck_instance_watchdog/job.go:96` — `bypass_authz` GUC scope ambiguity → potential RLS bleed across pooled connections

`set_config('metaldocs.bypass_authz', 'watchdog', true)` — the third argument `true` must be `is_local=true` (transaction-scoped) for the GUC to reset on `COMMIT`. If `is_local=false` (session-scoped), the GUC persists after commit and bleeds to the next user of the pooled connection, silently bypassing RLS for subsequent unrelated requests.

**Recommend:** confirm the third argument maps to `is_local=true` in the driver binding. Add a comment: `// is_local=true: GUC resets on COMMIT, safe for connection pool`. If `false`, switch immediately.

**Fix branch:** `fix/jobs-12-watchdog-h2-h3-h4`

---

### H4 — `stuck_instance_watchdog/job.go:100` — stuck-instance query crosses all tenants with no documented authz justification

```sql
SELECT ... FROM approval_instances ai
JOIN approval_stage_instances asi ON ...
WHERE asi.stuck_since IS NOT NULL -- no tenant_id predicate
```

Cross-tenant read under `bypass_authz`. If per-tenant isolation is expected, this is an IDOR. If the global sweep is intentional, it is undocumented.

**Recommend:** add a code comment: `// Intentional global sweep under bypass_authz — watchdog owns all tenants.` Confirm in PR review. If per-tenant isolation is required, add `WHERE ai.tenant_id = $1` and drive from a tenant iterator.

**Fix branch:** `fix/jobs-12-watchdog-h2-h3-h4`

---

## Medium

### M1 — `scheduler/scheduler.go:104` — `Register()` not protected by mutex → data race if called concurrently or after `Start()`

`Register()` appends to `s.jobs` without acquiring `s.mu`. Concurrent registration or post-`Start` registration races with the ticker goroutine.

**Recommend:** acquire `s.mu.Lock()` in `Register`. Document that `Register` must not be called after `Start`.

---

### M2 — `scheduler/scheduler.go:90` — `New()` accepts empty `leaderID` silently

Empty `leaderID` produces broken lease queries with no error.

**Recommend:** return `(*Scheduler, error)` and reject empty `leaderID` at construction time.

---

### M3 — `scheduler/scheduler.go:214` — failed run counted in both `RunsTotal` and `ErrorsTotal`

`incRun` is called unconditionally; `incError` is also called on error. Semantics are unclear.

**Recommend:** only increment `RunsTotal` for successful executions (rename to `AttemptsTotal` if both success and failure should be counted).

---

### M4 — `scheduler/scheduler.go:227` — `releaseLease` uses `context.Background()` without timeout

Under DB unavailability, the release call blocks indefinitely.

**Recommend:** `context.WithTimeout(context.Background(), 10*time.Second)`.

---

### M5 — `scheduler/scheduler.go:262` — `probePressure` has no query timeout; blocks ticker goroutine under connection exhaustion

The probe queries `pg_stat_activity` during the exact condition being probed.

**Recommend:** `context.WithTimeout(ctx, 3*time.Second)` inside `probePressure`.

---

### M6 — `scheduler/scheduler.go:309` — two separate lock acquisitions in `probePressure` → non-atomic read/write window

`currentPressure` acquires and releases `s.mu`, then `probePressure` acquires it again. The two critical sections are not atomic.

**Recommend:** fold `currentPressure`'s read into `probePressure` under a single lock.

---

### M7 — `scheduler/lease_reaper.go:50` — governance events inserted per-row in a loop → N individual INSERTs in one transaction

For many expired leases, this creates N statements inside a single transaction.

**Recommend:** batch with a single multi-row `INSERT` or `unnest` approach after the rows loop.

---

### M8 — `scheduler/lease_reaper.go:16` — per-row errors abort the loop, leaving partial governance audit trail

An error in the loop triggers `tx.Rollback()` via defer. Rows scanned before the error have no governance event written (rolled back), with no indication of partial processing.

**Recommend:** collect scan/insert errors, continue the loop, return a combined error after — so partial audit logging is visible in the error message.

---

### M9 — `stuck_instance_watchdog/job.go:113` — `StuckAfter` constant not wired to SQL; hardcoded `interval '7 days'` in query

The Go constant `StuckAfter = 7 * 24 * time.Hour` exists but is never used in the query. A maintainer who changes `StuckAfter` will not affect the actual threshold.

**Recommend:** pass as a query parameter: `AND ai.submitted_at < now() - ($2 * interval '1 second')` with `StuckAfter.Seconds()` as the arg.

---

### M10 — `idempotency_janitor/job.go:45` — `RowsAffected()` error discarded

`n, _ := result.RowsAffected()` — driver error silently ignored.

**Recommend:** `n, raErr := result.RowsAffected(); if raErr != nil { slog.ErrorContext(...); break }`.

---

## Low

### L1 — `scheduler/scheduler.go:373` — `waitForInFlight` busy-polls every 10ms

**Recommend:** close a `done` channel when `inFlightCount` reaches zero to eliminate the polling loop.

---

### L2 — `scheduler/scheduler.go:100` — logger hardcoded to `slog.NewTextHandler(os.Stdout, nil)`

Callers cannot inject a production logger.

**Recommend:** add `WithLogger(*slog.Logger)` functional option.

---

### L3 — `scheduler/lease_reaper.go:10` — `RunLeaseReaper` uses package-level `slog` default, not scheduler's injected logger

**Recommend:** accept `*slog.Logger` parameter.

---

### L4 — `audit_integrity_validator/job.go:34` — only first integrity issue logged; rest silently discarded

**Recommend:** log full slice count and first N (≤5) issue identifiers.

---

### L5 — `stuck_instance_watchdog/job.go:22` — `StuckInstance` exported with no constructor; can be constructed with empty `TenantID`

Only used as an internal query result.

**Recommend:** make unexported (`stuckInstance`).

---

### L6 — `idempotency_janitor/job.go:11` — `BatchSize`/`MaxIterations` exported consts are internal tunables

**Recommend:** unexport (`batchSize`, `maxIterations`).

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/jobs-12-reaper-c1` | C1 hardcoded system tenant_id + H1 FOR UPDATE SKIP LOCKED race | 1st |
| `fix/jobs-12-watchdog-h2-h3-h4` | H2 error swallow + H3 GUC scope + H4 cross-tenant undocumented | 2nd |
