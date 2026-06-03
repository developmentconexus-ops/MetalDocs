# ADR 0021 — Tenant admin vs. platform admin separation

> **Status:** Accepted — shipped at PR-12 (`feat/admin-center-rebuild-pr12`, 2026-06-03).
> **Last verified:** 2026-06-03

## Context

The pre-PR-12 Admin Center collapsed two distinct operator personas onto one `system_admin` role:

1. **Tenant admin** — manages users, roles, audit, sessions, usage *inside one tenant*. Scope = the operator's `tenantId`.
2. **Platform admin** — provisions tenants, manages cross-tenant grants, runs maintenance jobs. Scope = all tenants.

Because both shared `system_admin`, the FE could not gate by persona without leaking platform-only operations to tenant operators. PR-7b's cross-tenant guard hardened the backend but exposed the FE conflation: the route gate could not distinguish a tenant operator from a platform operator.

## Decision

`/admin/*` (the 6-tab IA from ADR [`0020`](0020-admin-center-six-tab-ia.md)) is the **tenant operator** surface. Every tab is gated by a tier-1 *tenant-scoped* capability (`user.view`, `membership.view`, `audit.read`, `metrics.view`, `user.manage`, `session.manage`, …). The current operator's `tenantId` is implicit — the backend resolves it from the session and applies the cross-tenant guard from PR-7b.

Platform admin surfaces (tenant provisioning, system-wide telemetry, maintenance jobs) are **out of scope** for `/admin/*`. When they are built they will live under a separate route family (e.g. `/platform/*`) gated by platform-scoped capabilities. They will not reuse the tenant Admin Center components.

The `system_admin` role is preserved in the canonical role enum because it still seeds the tenant capabilities, but the FE no longer pivots on the role name — it pivots on capabilities, which is the boundary the backend enforces.

## Consequences

**Wins.**

- Tenant operators see only their tenant's data — no accidental cross-tenant visibility from a stale FE filter.
- Platform admin can ship independently without re-litigating the tenant Admin Center IA.
- Backend cross-tenant guard (PR-7b) and FE capability gate are aligned on the same scope.

**Costs.**

- The role enum still ships `system_admin` for compatibility with existing seeds; renaming is deferred.
- Two stale `UserRole` literals (`admin`, `reviewer`) remain in `lib/types/index.ts`; see ADR [`0020`](0020-admin-center-six-tab-ia.md).
- No platform admin UI exists yet — when platform-scoped tooling is needed today it ships as a CLI or as ad-hoc SQL.

## References

- PR-7b hardening pass (cross-tenant guard + LIKE escape + permissions hardening)
- ADR [`0019-cap-audit-read-and-session-manage.md`](0019-cap-audit-read-and-session-manage.md)
- ADR [`0020-admin-center-six-tab-ia.md`](0020-admin-center-six-tab-ia.md)
- [`wiki/concepts/authz-tiers.md`](../concepts/authz-tiers.md)
- [`wiki/modules/frontend/iam.md`](../modules/frontend/iam.md)
