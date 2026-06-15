# Feature F4.6 — Spec

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.6-security-display-name-port`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / operator (Option-2 full close; the 3 security display-name
> reaches named in the M4 census correction + the coupled `CountRecentLockouts` JOIN). Internal Go
> port migration; no public contract change. Consumes the F4.1 `UserDisplayNameReader` and the F4.5
> `TenantUserReader`.

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Enrich display names in the security **Service** (move names out of the repo structs) or inside the security **Repository** (keep structs, source names via the port)? | **Inside the Repository.** `security/infrastructure/postgres.Repository` is security's single owning read-adapter; "return tenant-scoped security rows *with display names*" is its existing contract (`Lockout.DisplayName`, `RecentFailureSummary.DisplayName`, `NewDeviceLogin.DisplayName`). It now sources names from the iam port instead of a JOIN. This keeps the `securitydomain.Repository` interface, the `Service`, the HTTP handler, and `service_test.go` (which mocks the interface) **byte-identical** — minimal blast, and the cross-module read stays in infrastructure where it belongs. |
| 2 | Tenant scope for the 3 `auth_identities`-coupled methods (`ListLockouts`, `CountRecentFailedLoginsByUser`, `CountRecentLockouts`) after dropping the `iam_users` JOIN? | **`WHERE i.user_id = ANY($memberIDs)`** where `$memberIDs = TenantUserReader.TenantUserIDs(tenant)` (F4.5). `auth_identities` has no `tenant_id` column (ADR 0027); the INNER JOIN's membership test was exactly "user has an `iam_users` row in this tenant". `= ANY(member-id set)` reproduces that set **byte-identical** (F4.5 returns all members, no `deactivated_at` filter — matching the un-filtered JOIN). Empty tenant → empty id set → no rows / count 0 (same as the JOIN). |
| 3 | Tenant scope for `ListNewDeviceLogins`? | **`auth_sessions.tenant_id` directly** (separable — `auth_sessions` *has* `tenant_id`). Mirrors F4.4's `sessions_admin`. The old `iam_users` JOIN did double duty (scope + name); the honest scope is the owned `s.tenant_id` column, names come from the port. |
| 4 | `CountRecentLockouts` carries **no** display-name read — the milestone census listed it as a *genuine defer*. Migrate it or leave it? | **Migrate it.** Its only `iam_users` use is the tenant-scope JOIN, removable with the *same* F4.5 `TenantUserReader` already wired for the sibling lockout query — no new dependency, no new behavior. Leaving it would keep a needless `iam_users` JOIN in the same code path we are decoupling. ADR 0031 already commits F4.6 to it. This **narrows** the defer (strictly more closure on the same port), not scope creep. The remaining genuine defer is **`MfaCoverage`** alone (a true iam-owned aggregate over `iam_users.mfa_enabled` + `iam_user_roles` — counting iam's own data, not reaching for a foreign display column). |
| 5 | `display_name` value semantics? | Old (all 3): `COALESCE(NULLIF(u.display_name,''), <id>)`. New: `UserDisplayNameReader.DisplayNames(tenant, ids)` returns present+non-empty names; the repo maps any missing `id → id`. **Byte-identical** rendered value, same tenant scoping (port scopes `iam_users.tenant_id`). |
| 6 | Membership-orphan edge for `ListNewDeviceLogins` (INNER JOIN dropped sessions whose user lacks an `iam_users` row)? | Same documented note as F4.4: a session in tenant T exists only after a login to T (⇒ membership), so no orphan rows occur on the real path. The lockout/failed-login queries keep the membership filter exactly (via `= ANY(memberIDs)`), so they have **zero** edge. |

## Consumer contract (FIRST — before any producer)

- **Consumers (unchanged):** `security/application.Service` (`ListLockouts`, `ListSignals`) reads
  `Lockout.DisplayName`, `RecentFailureSummary.DisplayName`, `NewDeviceLogin.DisplayName` and the
  `CountRecentLockouts` int. The `securitydomain.Repository` interface and all domain structs are
  **unchanged** — the contract the Service depends on is preserved exactly.
- **Producers (already exist):**
  - `iamdomain.UserDisplayNameReader.DisplayNames(ctx, tenantID, userIDs) (map[string]string, error)` (F4.1) — names.
  - `iamdomain.TenantUserReader.TenantUserIDs(ctx, tenantID) ([]string, error)` (F4.5) — tenant member id set.
- **Source of truth:** the four existing security queries (`repository.go:83/128/168/184`) — their
  row sets, ordering, limits, and `COALESCE` fallback are the contract to reproduce.

## What this feature implements

1. **security `infrastructure/postgres/repository.go`** — `Repository` gains two collaborators:
   `displayNames iamdomain.UserDisplayNameReader` and `members iamdomain.TenantUserReader`, via
   constructor injection `NewRepository(db, displayNames, members)` (the established F4.1
   `NewPostgresApprovalRepository(db, reader)` idiom; required collaborators, not optional setters).
   - `ListLockouts` — drop `JOIN iam_users`; scope `WHERE i.user_id = ANY($1)` from `TenantUserIDs`;
     drop the `display_name` SELECT column + scan; enrich names via `DisplayNames` + `missing→user_id`.
   - `CountRecentFailedLoginsByUser` — same JOIN→`ANY(memberIDs)` swap + port enrichment.
   - `CountRecentLockouts` — drop `JOIN iam_users`; scope `WHERE i.user_id = ANY($1)` (COUNT only).
   - `ListNewDeviceLogins` — drop `JOIN iam_users`; keep `s.tenant_id` scope; enrich names via port.
   - One private helper resolves a `[]userID → map[id]name` with the `COALESCE` fallback.
2. **main.go wiring** — `securitypg.NewRepository(sqlDB, iampg.NewUserDisplayNameRepository(sqlDB), iampg.NewTenantUserRepository(sqlDB))`.
3. Reads stay **live**, on the pool (off-tx). No lock-holding tx in these read paths — H-PRE-1 not in play.

## Non-goals (mandatory)

- **No** change to `MfaCoverage` (genuine bounded defer — iam-owned aggregate) or `ListOffHoursAdminActions`
  (`iam_user_roles` + `audit_events`, no `iam_users` display-name read — out of class).
- **No** change to the `securitydomain.Repository` interface, domain structs, `Service`, HTTP handler,
  OpenAPI, or response shape.
- **No** snapshot/denormalization (D4/Approach-3 — reads live).
- **No** adjacent refactor beyond the named file + main.go wiring (CLAUDE.md §5.3).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `security/` issues 0 `iam_users.display_name` reads | `grep -rn "display_name" internal/modules/security/ --include=*.go` → only doc/test refs, no SQL | real |
| `security/` `auth_identities`/lockout paths issue 0 `iam_users` JOINs (only `MfaCoverage`'s aggregate remains) | `grep -rn "iam_users" internal/modules/security/infrastructure/postgres/repository.go` → only `MfaCoverage` (lines ~31/44) | real |
| `ListLockouts` — members locked surface w/ name; other-tenant excluded; name byte-identical incl. `missing→user_id` fallback | new live-PG integration on `*Repository.ListLockouts` (real ports) | **real (live PG)** |
| `CountRecentFailedLoginsByUser` — tenant-scoped, threshold+window honored, name fallback | new live-PG integration | **real (live PG)** |
| `CountRecentLockouts` — tenant-scoped count, other-tenant excluded | new live-PG integration | **real (live PG)** |
| `ListNewDeviceLogins` — tenant-scoped via `auth_sessions.tenant_id`, new-device anti-join intact, name via port | new live-PG integration | **real (live PG)** |
| Existing security service tests green (interface unchanged) | `go test ./internal/modules/security/...` | fixture |
| `go build ./...` + `go vet` (incl. `-tags integration`) clean | `go build ./...`; `go vet ./internal/modules/security/... ./apps/api/...`; `go vet -tags integration ./internal/modules/security/...` | — |
| backend-api-qa-checklist green | checklist at close | — |

> TDD: failing live-PG integration (name via port + tenant scope via member-id set) first, then
> implement to green. The `securitydomain.Repository` interface is unchanged, so `service_test.go`
> stays green throughout (regression guard).

## ADR needed?

- [ ] No new durable decision — F4.6 **consumes** ADR 0029 (`UserDisplayNameReader`) and ADR 0031
  (`TenantUserReader`); both already record the boundary, the reads-live constraint, and name F4.6 as
  the consumer. This feature adds the security consumer to those ADRs' Key files (already listed in
  ADR 0031) and closes the H-G display-name class at true zero outside `iam/`.
