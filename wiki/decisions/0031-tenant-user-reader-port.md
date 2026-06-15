# ADR 0031 — `TenantUserReader`: cross-module tenant-membership reads go through an iam-owned port

> **Status:** Accepted 2026-06-15
> **Last verified:** 2026-06-15
> **Scope:** How modules other than iam learn which users belong to a tenant, so they can tenant-scope a table that has no `tenant_id` column of its own (notably `metaldocs.auth_identities`) without JOINing `metaldocs.iam_users`. The owning module (iam), the port shape (tenant → member user_id set), the membership semantics (all members, no `deactivated_at` filter), and the reads-live / off-tx (H-PRE-1) constraint.
> **Out of scope:** display-name reads (ADR 0029's `UserDisplayNameReader`); iam's own intra-module membership reads; active-only / role-filtered membership variants (not built until a consumer needs them); the two-tier authz model (ADR 0022).
> **Key files:**
> - `internal/modules/iam/domain/tenant_user_reader_port.go` — the owned port (`TenantUserIDs`, `NoopTenantUserReader` null-object)
> - `internal/modules/iam/infrastructure/postgres/tenant_user_repository.go` — pool-backed impl (`SELECT user_id FROM metaldocs.iam_users WHERE tenant_id = $1::uuid`)
> - `internal/modules/security/infrastructure/postgres/repository.go` — consumer (M4/F4.6): `ListLockouts` / `CountRecentFailedLoginsByUser` / `CountRecentLockouts` scope `auth_identities` via `= ANY(ids)` instead of the `iam_users` JOIN

## Context

The Sessions & Security module reports on `metaldocs.auth_identities` (lockouts, recent failed logins).
`auth_identities` is a **global-PK table with no `tenant_id` column** (ADR 0027) — its only tenant
scoping was `JOIN metaldocs.iam_users u ON u.user_id = i.user_id WHERE u.tenant_id = $1`. That JOIN is a
cross-module reach: a consumer module issuing SQL against iam's owned table, coupling to its physical
schema (H-G class, M4). ADR 0029 closed the *display-name* reaches but explicitly deferred this
*tenant-scope* JOIN as "a different concern". The M4 census correction (2026-06-15) found the same
security methods also read `display_name` through that JOIN, so under operator Option-2 (full close) the
JOIN must go — which requires an iam-owned way to express tenant membership.

## Decision

Introduce a second narrow iam-owned port, **`TenantUserReader`**, distinct from `UserDisplayNameReader`
(Interface Segregation — membership and display-name are different concerns):

```go
TenantUserIDs(ctx, tenantID string) ([]string, error)
```

A `(user_id, tenant_id)` row in `metaldocs.iam_users` **is** the tenant membership (exactly what the
security INNER JOIN tested), so the impl is `SELECT user_id FROM metaldocs.iam_users WHERE tenant_id =
$1::uuid`. Consumers fetch the id set and filter their own `tenant_id`-less table with `WHERE col =
ANY(ids)`.

Membership semantics are **all members regardless of `deactivated_at`** — the consumers it replaces do
not filter on `deactivated_at` (only `MfaCoverage` does, and that stays an intra-security aggregate),
so returning all members keeps behavior byte-identical. The read is on the connection **pool**, never a
caller's lock-holding transaction (H-PRE-1). Empty/unknown tenant → empty slice, nil error. A
`NoopTenantUserReader` null-object mirrors `NoopUserDisplayNameReader` for ctor/test sites.

Reads stay **live** — no snapshot/denormalization of membership into consumer tables (design
constraint D4 / Approach-3).

## Consequences

- Security (M4/F4.6) drops every `iam_users` JOIN: tenant scope comes from `TenantUserIDs`, display
  names from `UserDisplayNameReader`. No cross-module `iam_users` SQL remains in security's
  display-name/lockout paths.
- Consumers couple to a stable Go interface; iam can change `iam_users` physical layout without breaking
  them.
- Two small ports instead of one wide one (ISP) — a consumer that needs only membership does not depend
  on the display-name surface, and vice-versa.
- `= ANY(ids)` ships the member-id set to the consumer; for very large tenants this is a larger payload
  than a JOIN, but it is the minimal decoupling and matches existing list-query scales (LIMIT 100/50).

### Alternatives rejected

- **Widen `UserDisplayNameReader` to also return membership** — conflates two concerns; violates ISP;
  forces display-name-only consumers to carry membership semantics.
- **Keep the `iam_users` JOIN in security** — the H-G reach this program exists to eliminate.
- **A membership *predicate* (`IsMember(tenant, user)`)** — can't be applied set-wise in one SQL
  statement across the boundary without re-introducing a JOIN; N+1 per row otherwise.
- **Snapshot tenant membership into `auth_identities`** — denormalization rejected program-wide (D4).

## References
- Feature F4.5 — `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4-systemic-ports/f4.5-iam-tenant-membership-port/spec.md`
- Consumer F4.6 — `…/milestone-4-systemic-ports/f4.6-security-display-name-port/spec.md`
- Sibling port ADR [`0029-user-display-name-reader-port.md`](0029-user-display-name-reader-port.md) (display names; this ADR supersedes its "membership port deferred" note)
- ADR 0027 — `auth_identities` global-PK (no `tenant_id`)
- H-PRE-1 — advisory-lock deadlock constraint (never an `iam_users` read inside a lock-holding tx)
