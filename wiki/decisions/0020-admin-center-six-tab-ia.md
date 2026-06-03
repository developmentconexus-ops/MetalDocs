# ADR 0020 — Admin Center 6-tab information architecture

> **Status:** Accepted — shipped at PR-12 (`feat/admin-center-rebuild-pr12`, 2026-06-03).
> **Last verified:** 2026-06-03

## Context

Pre-PR-12 the Admin Center was a single composed view (`AdminCenterView`) backed by a god hook (`useAdminCenter`) plus `useManagedUsers` and `state/admin.store.ts`. It conflated lifecycle (users, roles) with observability (audit, sessions, usage), forced every visit to pay for every fetch, and made capability gating coarse-grained (`requiresAdmin` boolean).

The hardening pass at PR-7b uncovered two cross-cutting needs that the monolith made expensive: granular tier-1 capabilities (`CapAuditRead`, `CapSessionManage`, `CapMembershipView`, `CapMetricsView`) and an evidence-grade audit/sessions surface. Splitting the surface was the prerequisite for landing those capabilities without redesigning the parent page each time.

## Decision

Adopt a 6-tab IA, one route per tab, one capability per route handle:

| Tab | Route | Capability |
|---|---|---|
| Visão geral | `/admin/overview` | any of `user.view`, `membership.view`, `metrics.view` |
| Pessoas | `/admin/people` (+ `/:userId` drawer) | `user.view` |
| Funções e capacidades | `/admin/roles` | `membership.view` |
| Auditoria | `/admin/audit` | `audit.read` |
| Sessões & Segurança | `/admin/sessions` | `user.view` |
| Consumo | `/admin/usage` | `metrics.view` |

Each tab owns its `*.route.tsx` lazy entry + `*.tsx` page component, its TanStack Query hooks under `features/iam/queries/`, and its mutations under `features/iam/mutations/`. All API calls go through typed openapi-fetch (`api.GET/POST/PATCH/DELETE`), generated from the OpenAPI spec.

The pre-existing `AdminCenterView`, `useAdminCenter`, `ManagedUsersPanel*`, `useManagedUsers`, and `state/admin.store.ts` are deleted (PR-12 Phase 1). The legacy IAM fields are removed from `ui.store.ts`. The membership area admin route (`/admin/memberships`) remains under `apiFetch` for now (deferred).

## Consequences

**Wins.**

- Lazy loading per tab — visiting Overview no longer parses Audit/Sessions/Usage code.
- Capability gating moves to route handles, enabling per-tab redirect (gated by `AppShell`).
- Each tab's mutations/queries live next to its UI — no central god hook to thread.
- New tabs (e.g. "Notifications", "API keys") are additive: one route, one cap, one hook tree.

**Costs.**

- One more level of nesting in the router config; tab additions need a 3-file change (route, page, query/mutation hooks).
- Two stale literals (`admin`, `reviewer`) remain in `UserRole` because Phase 1 chose not to touch unrelated call sites; cleanup deferred.
- (Fixed at PR-12 closeout) `AppShell` capability gate now collects every `required*Capability` along the match chain and requires every one to pass — parent + child constraints both enforced.

## References

- PR-12 spec (handed to operator at kickoff)
- [`wiki/modules/frontend/iam.md`](../modules/frontend/iam.md)
- [`wiki/concepts/authz-tiers.md`](../concepts/authz-tiers.md)
- ADR [`0019-cap-audit-read-and-session-manage.md`](0019-cap-audit-read-and-session-manage.md)
- ADR [`0021-tenant-vs-platform-admin-separation.md`](0021-tenant-vs-platform-admin-separation.md)
