# Sessions & Security — Tech Debt Register

**Last verified:** 2026-06-10 — Stage-1 backend audit drift patch.
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
| 7 | Cursor pagination for `/auth/sessions` — `has_more` truthfulness fixed 2026-07-02 (CON-11/APP-03); opaque `next_cursor` still deferred | ~~OpenAPI exposes the `cursor` param but the v1 handler returns `has_more=false` and a hard `LIMIT 200`.~~ `internal/modules/iam/delivery/http/sessions_handler.go`'s `handleSessions` now requests `limit+1` from the `SessionAdmin` port (over-fetch-by-one), truncates the extra row before it reaches the wire, and sets `Page.HasMore` truthfully instead of the hardcoded `false`. The OpenAPI `cursor` param remains unused/dead — the handler still does not accept or emit an opaque `next_cursor`, so a caller can detect "there is more" but not yet request the next page without re-deriving an offset out of band. The underlying repo (`internal/modules/auth/infrastructure/postgres/sessions_admin.go`, outside this module's scope) still caps at `LIMIT 200` with `ORDER BY last_seen_at DESC, session_id` (already has a stable tiebreaker) and no `OFFSET` — flat top-N only, no true paging today regardless of `has_more`. Tests: `internal/modules/iam/delivery/http/sessions_handler_test.go` — `TestSessionsHandler_ListHasMoreTrueWhenExtraRowExists`, `TestSessionsHandler_ListHasMoreFalseWhenAllSessionsFit`, `TestSessionsHandler_ListDefaultLimitAppliedWhenOmitted`. | A tenant routinely exceeds the limit, OR the UI grows a server-paginated table needing real `next_cursor` forward paging (would require wiring the existing `internal/platform/pagination.EncodeCursor`/`DecodeCursor` primitive end to end: spec already declares `cursor`, so no contract change needed there — only handler + repo implementation). |
| 8 | `bulk-permission-change` signal kind | Listed in the PR-7 spec but not present in the PR-3 OpenAPI `SecuritySignalKind` enum. Adding it would force a codegen regen. | Contract change is scheduled with PR-12 (FE consumer). |
| 9 | Real-time signal push | FE refetches `/security/signals` when the cache is stale (staleTime = 5 min, `useSecuritySignalsQuery.ts:5,17`); no polling interval is set. A websocket / SSE push would shave perceived latency. | Operators complain about lag, or push infra (mercure / SSE) lands generically. |
| 10 | Memory-mode parity for Sessions & Security | When `deps.SQLDB == nil` (memory mode / no SQL DB), every handler returns `200 OK` with zero/empty data: MFA-coverage returns `{"total_users":0,"mfa_enabled":0,...}` (`handler.go:44–47, 166–173`) and lockouts/signals return `{"items":[]}` (`handler.go:66–68, 106–109`). No 501 is emitted. Memory mode is dev-only; the SQL JOINs that drive the tab don't translate cleanly to the in-memory maps. | Memory mode becomes the test path for E2E coverage, or someone wires an integration test that needs it. |

## Source pointers

- Service contract: [`internal/modules/security/application/service.go`](../../internal/modules/security/application/service.go)
- Signal rules: [`security-signals.md`](security-signals.md)
- Migration that added stub MFA columns: [`db/migrations/0222_iam_mfa_and_failed_login_metadata.sql`](../../db/migrations/0222_iam_mfa_and_failed_login_metadata.sql) (the `archive/migrations/0210_*` path is not replayed at runtime)
