# Security Signals (rule-based, v1)

**Last verified:** 2026-06-02 (PR-7 backend implementation).
**Owner:** Admin Center Sessions & Security tab.
**Source of truth:**
- Service: [`internal/modules/security/application/service.go`](../../internal/modules/security/application/service.go)
- Repository: [`internal/modules/security/infrastructure/postgres/repository.go`](../../internal/modules/security/infrastructure/postgres/repository.go)
- Handler: [`internal/modules/security/delivery/http/handler.go`](../../internal/modules/security/delivery/http/handler.go)
- HTTP endpoint: `GET /api/v1/security/signals`

## Scope

Tenant-scoped only. Every signal query filters on `tenant_id` from the
authenticated session (`tenant.FromContext`). Tenant admin sees ONLY their
tenant's signals. Cross-tenant correlation, ML/statistical anomaly detection,
geographic intelligence, and impossible-travel detection are explicitly
out-of-scope for v1 — see [`security-tech-debt.md`](security-tech-debt.md).

Signals are computed on-demand per request, not pre-aggregated. Each rule is a
single SQL query against `audit_events` + `auth_sessions` + `iam_users` +
`auth_identities`, filtered to the calling tenant and a small time window
(≤ 24h). If the workload grows, swap on-demand for a materialised view + cron
without changing the HTTP shape.

## Rules

| Kind                       | Severity | Trigger                                                                       | Window |
|----------------------------|----------|-------------------------------------------------------------------------------|--------|
| `repeated-failed-login`    | `warn`   | A user has `failed_login_attempts ≥ 5` AND `last_failed_login_at` within window | 15 min |
| `lockout-spike`            | `warn`   | `≥ 3` accounts in the tenant have `locked_until` set within window           | 1 h    |
| `new-device-login`         | `info`   | A successful session was created from a `(user_id, user_agent)` pair never seen in the lookback window | 24 h window; 90 d lookback |
| `off-hours-admin-action`   | `info`   | A mutating audit event by `system_admin`/`area_admin`/`qms_admin` between `22:00–06:00 UTC` | 24 h |

Severity vocabulary matches the OpenAPI `SecuritySignalSeverity` enum:
`info | warn | critical`.

### Why these rules

- **`repeated-failed-login`** uses the `failed_login_attempts` counter on
  `auth_identities` — the only signal that survives a `RecordSuccessfulLogin`
  reset. Without a per-attempt log table (deferred), this is the closest
  expression of "N failures within window" we can write in SQL.
- **`lockout-spike`** counts `locked_until` set within the window — a single
  lockout is noise, three is a pattern worth a warn card.
- **`new-device-login`** anti-joins `auth_sessions` against itself in the
  90-day lookback. The 90-day window is the trade-off between false-positive
  rate (smaller window = "new device" on every long absence) and signal
  freshness.
- **`off-hours-admin-action`** uses the audit log because admin mutations are
  the highest-impact actions a compromised credential can take. Off-hours
  (22:00–06:00 UTC) is a v1 approximation; per-tenant timezone is tracked in
  tech-debt.

## Signal lifecycle

- **Computed on every request.** Endpoint refresh cadence is set by the FE
  (currently 60s on tab focus).
- **`signalId` is content-addressed** (SHA-1 over `kind + tenantId +
  evidence`). The same evidence-tuple returns the same `signalId` across
  requests so the UI can dedupe and persist per-card dismiss state without a
  server round-trip.
- **No persistence.** Signals are not stored — they re-derive from the
  underlying tables.

## Tenant isolation invariants

- Every SQL query takes `tenantID` as the FIRST positional parameter.
- Joins between `auth_sessions` / `iam_user_roles` and `iam_users` are bound
  on **both** `(user_id, tenant_id)` — a `user_id` collision across tenants
  cannot leak.
- Unit tests in
  [`internal/modules/security/application/service_test.go`](../../internal/modules/security/application/service_test.go)
  verify "tenant A's signals are invisible to tenant B".

## Bounded defers

See [`security-tech-debt.md`](security-tech-debt.md) for the full list. Key
items:
- ML / statistical baselines (out of scope — Tier-A platform feature)
- Cross-tenant correlation (Tier-A)
- Geographic IP intelligence (third-party integration)
- Per-attempt failed-login log (would unlock real `count(*) within window`)
- Per-tenant timezone for off-hours predicate
- `bulk-permission-change` rule (waiting on OpenAPI enum extension)
