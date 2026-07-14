# ADR 0059 — IAM tenant-scoping rule: every IAM-owned table carries `tenant_id`, every repo method filters by it

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Records the existing convention that every table owned by the `iam` module carries a `tenant_id` column and every repository method that reads or writes it filters by that column. Closes tech-debt T-011 (`wiki/modules/iam-tech-debt.md`). Specific to IAM-owned tables; does not restate the pooled multi-tenant invariant for the whole system (already stated in `CLAUDE.md`) or the RLS rollout sequencing (ADR 0027).
- **Depends on:** the pooled multi-tenant invariant (every tenant table carries `tenant_id`; cross-tenant URL → 404) already binding across the system.

---

## Context

IAM manages `iam_users`, `iam_user_roles`, and related tables. The tenant-scoping convention — every table carries `tenant_id`, every repository method filters by it — was enforced by a specific historical fix (Group B, audit 2026-05-03, items B5/B6) but was never captured as a standalone decision. ADR 0007 references migration 0162 in passing (its "Key files" list) but does not author the tenancy rule itself.

### Verified runtime facts

- **`iam_user_roles` carries `tenant_id`.** `archive/migrations/0162_iam_user_roles_tenant_id.sql` adds the column (the tech-debt row's cited fix for B5/B6).
- **`iam_users` carries `tenant_id` (+ deactivation scoping).** `archive/migrations/0130_iam_users_tenant_deactivated.sql`.
- **`iam_groups` (and related group tables) carry `tenant_id`.** `archive/migrations/0163_iam_groups.sql`.
- **Every repository read filters by `tenant_id`.** `internal/modules/iam/infrastructure/postgres/role_provider.go:28-38` (`RolesByUserID`) — `WHERE u.user_id = $1 AND u.tenant_id = $2::uuid AND u.deactivated_at IS NULL`, joining `iam_user_roles` on `r.tenant_id = u.tenant_id` in addition to `r.user_id = u.user_id` (line 34) — the join itself is tenant-scoped, not just the outer predicate.
- **Every repository write filters or sets `tenant_id` explicitly.** `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:24-31` (`HasAnyRole` — `WHERE role_code = $1 AND tenant_id = $2::uuid`); `:37-50,52-80` (`UpsertUserAndAssignRoleTx` — upserts `iam_users` with an explicit `tenant_id` bind, code comment at line 61-64 explains why: "`tenant_id` is NOT NULL on `iam_users` with a sentinel default; it must be set explicitly so new rows carry the caller's real tenant ... `EXCLUDED.tenant_id` on conflict also repairs any pre-existing sentinel row"; role replacement at line 75-79 deletes `WHERE tenant_id = $1::uuid AND user_id = $2` before re-inserting — never a bare `user_id` predicate).
- **Tier-2 authz is layered on top, not a substitute.** The same `UpsertUserAndAssignRoleTx` calls `authz.SeedTxIdentity` + `authz.Require(ctx, tx, CapUserManage, "tenant")` (`role_admin_repository.go:54-59`) before the tenant-scoped writes — the tenancy filter and the capability check are two independent layers, matching the two-tier model (ADR 0007/0022).

## Decision

**Every table owned by the `iam` module MUST carry a `tenant_id` column, and every repository method (read or write) on an IAM-owned table MUST filter or set `tenant_id` explicitly — never rely on a bare primary-key or `user_id`-only predicate, even when the primary key is globally unique.** This is binding because IAM tables sit directly behind the tier-1/tier-2 authz decision path (`RolesByUserID` feeds capability resolution); a tenant-scoping gap here is a privilege-escalation-shaped bug, not a data-hygiene nicety.

Concretely:
1. New IAM-owned tables MUST add `tenant_id NOT NULL` at creation (no sentinel-default-then-backfill pattern for new tables — the sentinel-default handling in `UpsertUserAndAssignRoleTx` is a historical-data repair mechanism, not a template for new schema).
2. Every `WHERE` clause touching an IAM-owned table in the `iam` module's infrastructure layer MUST include a `tenant_id = $N::uuid` predicate (or an equivalent join-carried tenant predicate, as `RolesByUserID` does).
3. INSERT/UPSERT statements MUST bind `tenant_id` from the verified caller context, never trust a client-supplied value without validating it against the caller's actual tenant membership.
4. This is a repository-layer rule (defense-in-depth alongside RLS where applicable, per ADR 0027's sequencing) — it does not replace tier-1/tier-2 capability checks, which remain mandatory on the same code paths.

## Consequences

- T-011 (`wiki/modules/iam-tech-debt.md`) is closed by this ADR.
- Reviewers checking new IAM repository code should reject any query against an IAM-owned table that lacks an explicit `tenant_id` predicate, citing this ADR.
- No migration, schema change, or code change is required by this ADR — it documents and binds existing, verified runtime behavior (migrations 0130, 0162, 0163 already did the schema work; this ADR records the rule those migrations exist to serve).

## References

- `archive/migrations/0130_iam_users_tenant_deactivated.sql`, `0162_iam_user_roles_tenant_id.sql`, `0163_iam_groups.sql` — schema migrations adding `tenant_id` to IAM-owned tables.
- `internal/modules/iam/infrastructure/postgres/role_provider.go:19-40` — tenant-scoped read example.
- `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:12-80` — tenant-scoped write example, including the sentinel-repair comment.
- `wiki/modules/iam-tech-debt.md` T-011 — tech-debt row closed by this ADR.
- `wiki/bugs/audit-2026-05-03.md` B5/B6 — original fix this ADR retroactively documents.
- ADR [`0007-two-tier-authz.md`](0007-two-tier-authz.md) — tier-2 capability layer that sits alongside (not instead of) this tenancy rule.
