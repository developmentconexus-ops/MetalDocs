# Feature F4.5 — Evidence

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.5-iam-tenant-membership-port`  ·  **Closed:** 2026-06-15
> **Contract:** [`spec.md`](spec.md). Producer-only feature under operator Option-2 full close — builds
> the previously-deferred iam tenant-membership port so F4.6 can decouple security's
> `auth_identities`-coupled reads. A feature closes only when every row below is real, honestly-labeled.

## What was implemented

- **iam/domain port** `internal/modules/iam/domain/tenant_user_reader_port.go` — `TenantUserReader`
  interface (`TenantUserIDs(ctx, tenantID) ([]string, error)`) + `NoopTenantUserReader` null-object.
  Doc comments record the membership semantics (an `iam_users` (user_id,tenant_id) row IS the
  membership; all members, **no** `deactivated_at` filter; pool/off-tx, H-PRE-1).
- **iam/infrastructure impl** `internal/modules/iam/infrastructure/postgres/tenant_user_repository.go`
  — pool-backed `TenantUserRepository`, `NewTenantUserRepository(db)`, interface assertion
  `var _ iamdomain.TenantUserReader = (*TenantUserRepository)(nil)`. `TenantUserIDs` =
  `SELECT user_id FROM metaldocs.iam_users WHERE tenant_id = $1::uuid`; empty (non-nil) slice when none.
- **ADR 0031** `wiki/decisions/0031-tenant-user-reader-port.md` — new iam-owned membership boundary;
  context (security `auth_identities` coupling, ADR 0027 global-PK), decision (id-set port, all
  members/no `deactivated_at`, pool/off-tx, reads-live), alternatives rejected (widen display-name
  port; keep JOIN; predicate port; snapshot). Registered in `wiki/decisions/index.md`; supersedes ADR
  0029's deferred-membership note (cross-linked both ways).
- **No consumer wiring** — F4.6 wires this into security. The producer-before-consumer order is sound
  because the contract was read from F4.6's three existing coupled queries, not invented.

## Verification

Live integration used the dev Postgres
(`METALDOCS_DATABASE_URL=postgres://…@127.0.0.1:5433/metaldocs?sslmode=disable&search_path=metaldocs,public`,
pgx driver via `openLiveIAMDB`), `-tags integration -count=1`.

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing unit test first | add `TestNoopTenantUserReader` before the type | `vet.exe: …undefined: NoopTenantUserReader` (RED) → green after impl | fixture |
| `NoopTenantUserReader` → empty non-nil slice, nil err | `go test -count=1 -run TestNoopTenantUserReader ./internal/modules/iam/domain/` | `ok …/iam/domain 0.859s` | fixture |
| `TenantUserIDs` returns ALL members incl. a `deactivated_at IS NOT NULL` member; excludes other tenant | `go test -tags integration -count=1 -v -run TestTenantUserRepository_TenantUserIDs_Live …/iam/infrastructure/postgres/` → `returns_all_members_incl_deactivated_excludes_other_tenant` | `--- PASS` (asserts exactly {userActive, userDeactivated}) | **real (live PG)** |
| Unknown tenant → empty non-nil slice, nil err | same test → `unknown_tenant_returns_empty` | `--- PASS` | **real (live PG)** |
| Interface satisfied; impl compiles | `var _ iamdomain.TenantUserReader = (*TenantUserRepository)(nil)` | `BUILD OK` | real |
| build + vet (plain) | `go build ./...`; `go vet ./internal/modules/iam/...` | `BUILD OK` / `VET OK` | — |
| vet (integration tag) | `go vet -tags integration ./internal/modules/iam/...` | `VET-INTEGRATION OK` | — |

Verbose live run (proves not skipped):
```
=== RUN   TestTenantUserRepository_TenantUserIDs_Live
    --- PASS: …/returns_all_members_incl_deactivated_excludes_other_tenant (0.00s)
    --- PASS: …/unknown_tenant_returns_empty (0.00s)
--- PASS: TestTenantUserRepository_TenantUserIDs_Live (0.94s)
ok  metaldocs/internal/modules/iam/infrastructure/postgres 1.279s
```

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Port in iam/domain; pool-backed impl in iam/infrastructure; assertion compiles | yes | BUILD OK; assertion present |
| `TenantUserIDs` returns all members incl. deactivated (no active-only filter) | yes | live subtest PASS (real) |
| Tenant-scoped — other tenant excluded | yes | same subtest asserts exactly the 2 tenant-A members (real) |
| Unknown/empty tenant → empty slice, nil error | yes | `unknown_tenant_returns_empty` PASS (real) |
| `NoopTenantUserReader` → empty slice, nil error | yes | `TestNoopTenantUserReader` PASS (fixture) |
| build + vet (incl. integration) clean | yes | BUILD/VET/VET-INTEGRATION OK |

## Review disposition

- Spec-compliance: contract read from F4.6's three coupled queries (membership id-set, no
  `deactivated_at` filter) — matches the INNER JOIN it will replace byte-for-byte. ISP honored — a
  second narrow port distinct from `UserDisplayNameReader`.
- Code-quality: mirrors the F4.1 port idiom (null-object, interface assertion, pool-backed off-tx,
  `iam: …` wrapped errors, empty-non-nil slice). H-PRE-1 preserved (pool read).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Active-only / role-filtered membership variants | Not needed by any current consumer; YAGNI (would be a separate method, not a behavior change to this one) | Add a sibling method only when a consumer needs it. Owner: backend (recorded in ADR 0031 out-of-scope) |
| Consumer wiring (security) | This is F4.5's explicit non-goal; done in F4.6 | F4.6 `f4.6-security-display-name-port` |
