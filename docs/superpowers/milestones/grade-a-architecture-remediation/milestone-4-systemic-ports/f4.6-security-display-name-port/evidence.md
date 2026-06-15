# Feature F4.6 — Evidence

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.6-security-display-name-port`  ·  **Closed:** 2026-06-15
> **Contract:** [`spec.md`](spec.md). Closes the 3 security `iam_users.display_name` reaches (census
> correction) + the coupled `CountRecentLockouts` tenant-scope JOIN — the last H-G display-name sites
> outside `iam/`. Consumes ADR 0029 (`UserDisplayNameReader`) + ADR 0031 (`TenantUserReader`).

## What was implemented

- **`security/infrastructure/postgres/repository.go`** — `Repository` gained two iam-owned
  collaborators (`displayNames iamdomain.UserDisplayNameReader`, `members iamdomain.TenantUserReader`),
  injected via `NewRepository(db, displayNames, members)` (nil-guarded to Noop). New private
  `resolveNames(ctx, tenantID, ids)` helper = port `DisplayNames` + `missing→user_id` fallback
  (reproduces `COALESCE(NULLIF(display_name,''), user_id)` consumer-side). Four queries migrated:
  - `ListLockouts` — dropped `JOIN iam_users` + the `display_name` SELECT/scan; tenant scope is now
    `WHERE i.user_id = ANY($memberIDs)` (`TenantUserIDs`); names enriched post-scan.
  - `CountRecentFailedLoginsByUser` — same JOIN→`ANY(memberIDs)` swap + port enrichment.
  - `CountRecentLockouts` — dropped `JOIN iam_users`; `WHERE i.user_id = ANY($memberIDs)` (COUNT only).
  - `ListNewDeviceLogins` — dropped `JOIN iam_users`; kept `s.tenant_id` scope (auth_sessions owns it);
    names enriched via port. Orphan-session note documented (mirrors F4.4).
- **`apps/api/cmd/metaldocs-api/main.go`** — `securitypg.NewRepository(sqlDB, iampg.NewUserDisplayNameRepository(sqlDB), iampg.NewTenantUserRepository(sqlDB))`.
- **Unchanged (by design):** `securitydomain.Repository` interface, all domain structs, `application.Service`,
  the HTTP handler, OpenAPI — so `service_test.go` (mocks the interface) stays green as a regression guard.
- **ADR 0029** Key files += security consumer; consequences updated to record true-zero closure +
  `MfaCoverage` as the sole remaining (out-of-class) cross-module `iam_users` read.

## Verification

Live integration: dev Postgres `metaldocs-postgres` (`127.0.0.1:5433`, db `metaldocs`, user
`metaldocs_app`), pgx driver via `openLiveSecurityDB`, real iam ports (`NewUserDisplayNameRepository` +
`NewTenantUserRepository`), `-tags integration -count=1`.

| Check | Command / action | Result | Real vs fixture |
|-------|------------------|--------|-----------------|
| TDD — failing test first | `go vet -tags integration ./internal/modules/security/infrastructure/postgres/` before impl | `vet.exe: …too many arguments in call to securitypg.NewRepository have (*sql.DB, …Reader, …Reader) want (*sql.DB)` (RED) | fixture |
| `ListLockouts` — members only, deactivated member still surfaces, other-tenant + non-member excluded, names via port | live `…/ListLockouts_members_only_with_port_names` | `--- PASS` | **real (live PG)** |
| `CountRecentFailedLoginsByUser` — threshold+window honored, tenant-scoped, name enriched, below-threshold excluded | live `…/CountRecentFailedLoginsByUser_tenant_scoped` | `--- PASS` | **real (live PG)** |
| `CountRecentLockouts` — tenant-scoped count = 2 (members; tenant-B + non-member excluded) | live `…/CountRecentLockouts_tenant_scoped_count` | `--- PASS` | **real (live PG)** |
| `ListNewDeviceLogins` — scoped via `auth_sessions.tenant_id`, anti-join intact, name via port, `missing→user_id` fallback for a session whose user has no `iam_users` row (proves JOIN gone) | live `…/ListNewDeviceLogins_scoped_via_session_tenant_with_fallback` | `--- PASS` | **real (live PG)** |
| `security/` issues 0 `iam_users.display_name` SQL | `grep -rn "display_name" internal/modules/security/ --include=*.go` (non-test) | only api.gen.go DTO field, service.go/handler.go render maps, repository.go comments — **no SQL read** | real |
| Only `MfaCoverage`'s aggregate `iam_users` JOIN remains | `grep -n "iam_users" repository.go` → SQL only at lines 67/80, both inside `MfaCoverage` (63–119); all other methods clean | confirmed | real |
| Interface unchanged → service tests green (regression guard) | `go test ./internal/modules/security/...` | `ok …/security/application` | fixture |
| Regression — iam + auth unaffected | `go test ./internal/modules/security/... ./internal/modules/iam/... ./internal/modules/auth/...` | all `ok` | fixture |
| build + vet (plain + integration) | `go build ./...`; `go vet ./internal/modules/security/... ./apps/api/...`; `go vet -tags integration ./internal/modules/security/...` | `BUILD OK` / `VET OK` / `VET-INT OK` | — |

Verbose live run:
```
--- PASS: TestSecurityRepository_NoIamUsersJoin_Live (3.59s)
    --- PASS: …/ListLockouts_members_only_with_port_names (0.01s)
    --- PASS: …/CountRecentFailedLoginsByUser_tenant_scoped (0.00s)
    --- PASS: …/CountRecentLockouts_tenant_scoped_count (0.17s)
    --- PASS: …/ListNewDeviceLogins_scoped_via_session_tenant_with_fallback (0.00s)
ok  metaldocs/internal/modules/security/infrastructure/postgres 3.818s
```

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|--------------------------------|------|----------|
| `security/` issues 0 `iam_users.display_name` reads | yes | grep (real) |
| Only `MfaCoverage` aggregate JOIN remains | yes | grep lines 67/80 in MfaCoverage (real) |
| `ListLockouts` members-only, names byte-identical incl. fallback | yes | live subtest PASS |
| `CountRecentFailedLoginsByUser` tenant-scoped, threshold/window, fallback | yes | live subtest PASS |
| `CountRecentLockouts` tenant-scoped count | yes | live subtest PASS |
| `ListNewDeviceLogins` scoped via `auth_sessions.tenant_id`, anti-join, name via port | yes | live subtest PASS |
| Existing security service tests green (interface unchanged) | yes | `go test` ok |
| build + vet (incl. integration) clean | yes | BUILD/VET/VET-INT OK |

## Review disposition

- **Spec-compliance:** consumer contract honored exactly — `securitydomain.Repository`, domain structs,
  Service, handler, OpenAPI all unchanged; names + row sets reproduced byte-identical (membership via
  `= ANY(memberIDs)` matches the un-`deactivated_at`-filtered INNER JOIN; new-device scope via owned
  `s.tenant_id`). ISP honored — two narrow ports, not one wide one.
- **Code-quality:** constructor injection matches the F4.1 `NewPostgresApprovalRepository(db, reader)`
  idiom; nil-guard to Noop; single shared `resolveNames` helper; member-empty short-circuit avoids a
  pointless round-trip; reads on the pool (off-tx, H-PRE-1 not in play — no lock-holding tx here).
- **Scope:** `CountRecentLockouts` migration is **narrowing a named defer** with the already-wired
  F4.5 port (more closure, same dependency), not new scope — recorded in spec Q4 + ADR 0031.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `security.MfaCoverage` `iam_users` aggregate JOIN | Genuine iam-owned *aggregate over iam's own data* (`mfa_enabled` + `iam_user_roles`), not a display-name reach; counting iam_users is a different concern from reading a foreign display column | *Trigger:* M5 re-audit flags it / next structural touch of `MfaCoverage`. Owner: backend (ADR 0029 consequences) |
| `security.ListOffHoursAdminActions` `iam_user_roles` JOIN | `iam_user_roles` + `audit_events`, no `iam_users.display_name`; out of the H-G display-name class | Out of class — not an M4 target (recorded in spec non-goals) |
