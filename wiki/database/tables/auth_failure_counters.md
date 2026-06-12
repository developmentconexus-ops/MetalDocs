# public.auth_failure_counters

> **Source:** `db/migrations/0235_auth_failure_counters.sql`
> **Schema:** `public`
> **Owner:** approval
> **Last verified:** 2026-06-12 (created — F-20e, REQ-REL-3, OWASP ASVS §2.2.1)

## Purpose

Shared, Postgres-backed storage for the authentication-failure rate limiter used
by the approval e-signature (`password_reauth`) path. Replaces the process-local
`InMemoryAuthFailureRateLimiter` in production so lockout state survives API
restarts and is consistent across replicas (decision D-1: no Redis).

One row per `actor_id` (iam user UUID as text). The window is fixed: `fail_count`
resets when `window_start + 60s` elapses. After five failures within a window the
actor is blocked until the window expires or a successful authentication resets
the row via DELETE.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `actor_id` | `text` | no (PK) | IAM user UUID as text — matches `iam_users.id`. |
| `fail_count` | `integer` | no | Number of failures recorded in the current window. |
| `window_start` | `timestamptz` | no | Timestamp of the first failure in the current window. Reset to NOW() when the previous window expires. |

## Schema

```sql
CREATE TABLE public.auth_failure_counters (
    actor_id     TEXT        NOT NULL,
    fail_count   INTEGER     NOT NULL DEFAULT 0,
    window_start TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (actor_id)
);
```

## Important Indexes

- `auth_failure_counters_pkey` on `(actor_id)` — primary key, used by all three
  operations (Allow SELECT, RecordFailure UPSERT, Reset DELETE). EXPLAIN confirms
  Index Scan (cost=0.15..8.17) on the primary key for single-actor lookups.

## Runtime Usage

- **Reader:** `Allow` — `SELECT fail_count, window_start WHERE actor_id = $1`
- **Writer:** `RecordFailure` — UPSERT with window-reset CASE expression
- **Deleter:** `Reset` — `DELETE WHERE actor_id = $1` on successful authentication

Implementation: `internal/modules/documents/approval/infrastructure/signature/postgres_auth_failure_rate_limiter.go`  
Interface: `AuthFailureRateLimiter` in `internal/modules/documents/approval/infrastructure/signature/password_reauth.go`  
Wiring: `apps/api/cmd/metaldocs-api/reauth.go` — Postgres impl when `db != nil`, in-memory for dev/test.

## Seed or Reference Data

None. Rows are transient (created on first failure, deleted on success or window expiry).

## Notes and Debt

- Stale rows are pruned inline by the `RecordFailure` UPSERT CASE expression — no
  janitor job is needed for this table.
- The `actor_id` key is the user UUID only (no IP component), matching the
  in-memory implementation semantics exactly.
- Threshold `maxFailures = 5`, window `windowDur = 60s` — same constants as the
  in-memory predecessor (defined in `password_reauth.go`).
