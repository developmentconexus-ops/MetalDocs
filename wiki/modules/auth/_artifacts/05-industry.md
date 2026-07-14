# Phase 5 — Industry comparison (auth)

Patterns drawn from `references/industry-patterns-index.md` only. No fresh additions.

## Admissible patterns

### IP-001 — RFC 9457 Problem Details
- **Source:** https://www.rfc-editor.org/rfc/rfc9457.html (RFC 9457, 2023-07)
- **Quote:** "A problem details object can be extended with additional members."
- **Applies to:** `internal/modules/auth/delivery/http/handler.go:166` `writeAPIError`; `internal/modules/auth/delivery/http/middleware.go:65,76,79,83`.
- **Fact:** auth emits legacy envelope `{error:{code,message,details,trace_id}}`, not `application/problem+json`. Mirrors documents T-001 + iam T-006. Drives auth T-003.

### IP-004 — Defense-in-depth (NIST SP 800-95 §4.3)
- **Source:** NIST SP 800-95 (2007) — "Multiple layers of access control reduce single-point bypass risk."
- **Applies to:**
  - `internal/modules/auth/delivery/http/middleware.go:58` `LegacyHeaderEnabled` X-User-Id bypass — single-flag compromise. Drives auth T-001.
  - `internal/modules/auth/application/service.go:117-126` per-account bcrypt + lockout = single layer at identity level; no IP-based rate limit upstream. Drives auth T-005.

### IP-008 — Row-level tenant_id on multi-tenant tables
- **Source:** Crunchy Data multi-tenant Postgres design (accessed 2026-05-10) — "Add tenant_id to every multi-tenant table and index it first."
- **Applies to:** `migrations/0021_init_auth_identities_and_sessions.sql`, `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql`.
- **Fact:** `auth_identities` and `auth_sessions` have NO `tenant_id` column (artifact 04 §3). IAM tables added `tenant_id` in 0130/0162. Identity remains tenant-global. Drives auth T-008 (latent).

## Not applicable to auth

- **IP-002 (idempotency)** — auth ops are not idempotency-key-enabled by spec; login is intentionally non-idempotent (each login mints a new session row). Not applicable.
- **IP-003 (cursor pagination)** — auth has no list endpoints exposed via its own HTTP surface (`ListUsers` is invoked via `iam` admin handler). Not applicable to auth's surface.
- **IP-005 (OpenAPI as source-of-truth)** — auth handler registers via stdlib `mux.HandleFunc` (`handler.go:35-39`); no oapi-codegen stub for `/api/v1/auth/*`. Drift fact captured in §2 of module doc; not promoted to debt because auth was excluded from ADR 0012 partial rollout (consistent with iam scope).
- **IP-006 (forward-only migrations)** — auth migrations 0021/0022/0036/0159 are append-only; observed compliant. No debt.
- **IP-007 (observability)** — auth uses `log.Printf` only (`handler.go:56,112`); no structured logger or trace-id propagation beyond opportunistic `X-Trace-Id` header echo. Index row IP-007 itself is not yet wired anywhere in MetalDocs; per index notes, "flag as missing-ADR if a module assumes it" — auth does not assume it. No debt.

## Summary

3 admissible patterns drive 3 debt items (T-001, T-005, T-008 via IP-004 & IP-008; T-003 via IP-001). Severity rationale captured per-row in `auth-tech-debt.md`.
