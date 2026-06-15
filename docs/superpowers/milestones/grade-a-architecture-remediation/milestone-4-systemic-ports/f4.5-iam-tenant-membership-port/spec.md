# Feature F4.5 — Spec

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.5-iam-tenant-membership-port`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / operator (Option-2 full close — the previously-deferred iam
> tenant-scope/membership port, now built so F4.6 can decouple security's `auth_identities`-coupled
> reads). New iam-owned port; no public contract change. Engineering decisions in the interview record.

> The producer (this port) is built **before** its consumer's migration (F4.6), but its **contract is
> read from that consumer** — the three coupled security queries — not invented. The validator judges
> F4.5 against this file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Why does security need a new port at all (F4.1's display-name port exists)? | Security's `ListLockouts` / `CountRecentFailedLoginsByUser` / `CountRecentLockouts` query `auth_identities` (a **global-PK table with no `tenant_id` column**, ADR 0027). Their only tenant scoping is `JOIN iam_users u ON u.user_id=i.user_id WHERE u.tenant_id=$1`. To drop that cross-module JOIN, security needs an iam-owned way to learn **which user_ids belong to a tenant** — a different concern from display names. |
| 2 | Shape: a membership *predicate*, a *set of ids*, or per-row enrichment? | **Set of ids.** `TenantUserIDs(ctx, tenantID) ([]string, error)`. Security then scopes `auth_identities` with `WHERE i.user_id = ANY($ids)`. A predicate can't cross the module boundary in one SQL statement without the JSON/JOIN it's meant to remove; the id-set is the minimal decoupling that reproduces the INNER-JOIN membership exactly. |
| 3 | `deactivated_at` filter? | **No filter.** The three coupled queries JOIN `iam_users` with **no** `deactivated_at` predicate (unlike `MfaCoverage`, which filters `deactivated_at IS NULL`). To stay byte-identical, `TenantUserIDs` returns **all** member user_ids regardless of `deactivated_at`. (A future active-only variant is a separate method if ever needed — not this feature.) |
| 4 | Membership source of truth? | `metaldocs.iam_users` — a `(user_id, tenant_id)` row **is** the membership (that is exactly what the security INNER JOIN tests). So `SELECT user_id FROM metaldocs.iam_users WHERE tenant_id = $1::uuid`. iam owns this table. |
| 5 | Pool or tx? | **Pool (off-tx).** Same H-PRE-1 discipline as `UserDisplayNameReader` — these are list/aggregate reads, never inside a lock-holding tx. |
| 6 | Empty tenant / unknown tenant? | Returns `([]string{}, nil)` (empty slice, nil error). Security's `ANY('{}')` then matches no rows — identical to the INNER JOIN producing zero rows for a tenant with no members. |

## Consumer contract (FIRST — read from F4.6's three coupled queries)

- **Consumer (F4.6):** `security/infrastructure/postgres/repository.go` — `ListLockouts`,
  `CountRecentFailedLoginsByUser`, `CountRecentLockouts`. Each must scope `auth_identities` to a
  tenant **without** joining `iam_users`.
- **Required surface:**
  `iamdomain.TenantUserReader.TenantUserIDs(ctx context.Context, tenantID string) ([]string, error)`
  — returns every `metaldocs.iam_users.user_id` for `tenant_id`, no `deactivated_at` filter, empty
  slice when none. Tenant-scoped via `tenant_id = $1::uuid`.
- **Null-object:** `NoopTenantUserReader` returning `([]string{}, nil)` — for ctor sites/tests that
  don't resolve membership (mirrors `NoopUserDisplayNameReader`).

## What this feature implements

1. **iam/domain** — `tenant_user_reader_port.go`: `TenantUserReader` interface + `NoopTenantUserReader`
   null-object, with doc comments stating the membership semantics (all members, no `deactivated_at`
   filter, pool/off-tx, H-PRE-1).
2. **iam/infrastructure/postgres** — `tenant_user_repository.go`: `TenantUserRepository` (pool-backed)
   implementing `TenantUserIDs` with `SELECT user_id FROM metaldocs.iam_users WHERE tenant_id = $1::uuid`;
   `var _ iamdomain.TenantUserReader = (*TenantUserRepository)(nil)` assertion; `NewTenantUserRepository(db)`.
3. **No consumer wiring in this feature** — F4.6 wires it into security. F4.5 ships the producer +
   tests + ADR only (the producer-before-consumer order is acceptable here because the contract is
   read from F4.6's existing queries, not invented).

## Non-goals (mandatory)

- **No** change to security yet (that is F4.6).
- **No** `deactivated_at` / active-only filtering (Q3) — would change behavior.
- **No** snapshot/denormalization (reads live).
- **No** widening of `UserDisplayNameReader` to carry membership (separate concern — ISP).
- **No** adjacent refactor beyond the two new files + ADR (CLAUDE.md §5.3).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Port in iam/domain; pool-backed impl in iam/infrastructure; interface assertion compiles | `go build ./...`; `var _ ... = (*TenantUserRepository)(nil)` | real |
| `TenantUserIDs` returns all member user_ids for a tenant (incl. a `deactivated_at IS NOT NULL` member — proves no active-only filter) | new live-PG integration test `TestTenantUserRepository_TenantUserIDs_Live` | **real (live PG)** |
| Tenant-scoped — a member of another tenant is not returned | same live test (second tenant seeded) | **real (live PG)** |
| Unknown/empty tenant → empty slice, nil error | same live test case | **real (live PG)** |
| `NoopTenantUserReader` returns empty slice, nil error | unit test `TestNoopTenantUserReader` | fixture |
| build + vet (incl. integration) clean | `go build ./...`; `go vet [-tags integration] ./internal/modules/iam/...` | — |

> TDD: failing unit test (Noop + interface satisfaction) first; live-PG integration proves the real
> membership read against `metaldocs.iam_users`.

## ADR needed?

- [x] **Yes — new durable decision.** A new iam-owned **TenantUserReader** boundary (tenant→member-ids)
  for cross-module consumers that must scope a `tenant_id`-less table (`auth_identities`). Record as a
  new ADR (0031) — context (security's `auth_identities` coupling; ADR 0027 global-PK), decision
  (owning-module id-set port, all members/no `deactivated_at`, pool/off-tx), consequences (reads live,
  H-PRE-1, ISP vs the display-name port), alternatives rejected (widen `UserDisplayNameReader`;
  keep the JOIN; predicate port). Supersedes the F4.3/ADR-0029 "membership port deferred" note.
