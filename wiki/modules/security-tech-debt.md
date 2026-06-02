# Sessions & Security — Tech Debt Register

**Last verified:** 2026-06-02 (PR-7).
**Scope:** features deliberately deferred from the v1 rebuild of the Admin
Center Sessions & Security tab.

Each row is a bounded defer with an explicit reason and the trigger that would
promote it to active work. New items appended in chronological order.

| # | Item | Reason for defer | Promote when |
|---|------|------------------|--------------|
| 1 | Real MFA enrollment flow (TOTP / WebAuthn) | PR-7 added `iam_users.mfa_enabled` + `mfa_enrolled_at` as stub columns so the coverage endpoint can render an honest 0% without blocking the tab. An actual enrollment product is a separate workstream that touches login + recovery + admin reset. | A customer requires MFA, or a compliance audit forces it. |
| 2 | Per-attempt failed-login log table | The `repeated-failed-login` signal currently uses `failed_login_attempts` + `last_failed_login_at` as a proxy because `auth_identities` does not record each attempt. False negatives occur when `RecordSuccessfulLogin` resets the counter mid-window. | False-negative rate becomes a complaint, or a SIEM integration demands raw attempt rows. |
| 3 | Per-tenant timezone for off-hours predicate | `off-hours-admin-action` uses `UTC 22:00–06:00` as a v1 approximation. `iam_tenants` does not carry a timezone column. | Multi-region rollout, or a tenant in a non-UTC-aligned timezone reports false positives. |
| 4 | ML / statistical anomaly detection | Tier-A platform feature; cross-tenant baselines + behavioural fingerprints + impossible-travel detection. Out of scope for tenant-admin v1. | Tier-A platform product picks this up. |
| 5 | Cross-tenant correlation | Tenant admins see only their tenant. Platform-owner correlation (e.g. same actor probing multiple tenants) belongs to a separate platform-owner scope. | Platform-owner control plane lands. |
| 6 | Geographic IP intelligence | Requires a third-party IP-geo provider (MaxMind etc.) and the legal review that comes with it. | A control-tower partner requires it, or a sec-eng program funds the provider. |
| 7 | Cursor pagination for `/auth/sessions` | OpenAPI exposes the `cursor` param but the v1 handler returns `has_more=false` and a hard `LIMIT 200`. Sufficient until any tenant has >200 concurrent sessions. | A tenant routinely exceeds the limit, OR the UI grows a server-paginated table. |
| 8 | `bulk-permission-change` signal kind | Listed in the PR-7 spec but not present in the PR-3 OpenAPI `SecuritySignalKind` enum. Adding it would force a codegen regen. | Contract change is scheduled with PR-12 (FE consumer). |
| 9 | Real-time signal push | FE polls `/security/signals` on tab focus (60s). A websocket / SSE push would shave the perceived latency. | Operators complain about lag, or push infra (mercure / SSE) lands generically. |
| 10 | Memory-mode parity for Sessions & Security | The handler returns `501 Not Implemented` when `deps.SQLDB == nil`. Memory mode is dev-only; the SQL JOINs that drive the tab don't translate cleanly to the in-memory maps. | Memory mode becomes the test path for E2E coverage, or someone wires an integration test that needs it. |

## Source pointers

- Service contract: [`internal/modules/security/application/service.go`](../../internal/modules/security/application/service.go)
- Signal rules: [`security-signals.md`](security-signals.md)
- Migration that added stub MFA columns: [`migrations/0210_iam_mfa_and_failed_login_metadata.sql`](../../migrations/0210_iam_mfa_and_failed_login_metadata.sql)
