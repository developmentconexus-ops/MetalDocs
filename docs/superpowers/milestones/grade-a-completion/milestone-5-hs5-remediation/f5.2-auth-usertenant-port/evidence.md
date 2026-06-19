# Evidence — F5.2 auth-usertenant-port (H-G site #1)

> **Status:** CLOSED 2026-06-16 · narrow iam read-port extraction; behavior-preserving SQL move.

## Change

| File | Change |
|------|--------|
| `internal/modules/iam/domain/user_tenant_reader_port.go` | **new** — `UserTenantReader` interface (`UserTenantIDs`) + `NoopUserTenantReader`. Inverse of `TenantUserReader`. |
| `internal/modules/iam/infrastructure/postgres/user_tenant_repository.go` | **new** — `UserTenantRepository` + `NewUserTenantRepository(db)`. Pool-backed (off-tx, H-PRE-1). Query character-identical to the replaced auth query (DISTINCT/WHERE/ORDER BY). |
| `internal/modules/iam/infrastructure/postgres/user_tenant_repository_test.go` | **new** — unit query-parity pin + Noop contract test (no DB). |
| `internal/modules/iam/infrastructure/postgres/user_tenant_repository_integration_test.go` | **new** — `//go:build integration` live parity test. |
| `internal/modules/auth/infrastructure/postgres/repository.go` | `GetUserTenants` now delegates to the iam port; removed direct `FROM metaldocs.iam_user_roles` query; struct gains `userTenants` field; `NewRepository(db, userTenants)` nil→Noop; added `iamdomain` import. |
| `internal/platform/bootstrap/api.go:80` | wires `iampg.NewUserTenantRepository(db)`. |
| `apps/api/cmd/metaldocs-api/main.go:245,376` | wire the real port into the two ad-hoc auth-repo constructions. |
| 6 auth-pg test callers (`service_test.go`, `repository_test.go`×4 incl. `repository_unit_test.go`×3, `sessions_admin_integration_test.go`) | `NewRepository(db, nil)` — none exercise postgres `GetUserTenants`; Noop is behavior-equivalent for them. |

## HS-2 disposition

Did **not** fire. The change is one narrow read port mirroring four existing iam ports — additive,
no auth↔iam contract redesign, no iam write-model change. Stayed inside the boundary.

## TDD record

1. **Red:** `user_tenant_repository_test.go` referenced `userTenantsQuery` / `NewUserTenantRepository`
   before they existed → compile failure (the red).
2. **Green:** added port + Postgres impl → unit test passes.
3. Consumer rewired; full module suites green.

The SQL move is parity-preserving **by construction** — the new `userTenantsQuery` is
character-identical to the removed auth query on the semantics that matter (`DISTINCT tenant_id`,
`FROM metaldocs.iam_user_roles`, `WHERE user_id = $1`, `ORDER BY tenant_id`). The unit test pins
exactly those four invariants so a future drift fails CI without a DB.

## Validation Gate results (real output)

1. **H-G grep → 0** —
   `grep -rn "FROM metaldocs\.iam_user_roles" --include="*.go" internal/modules/ | grep -v "internal/modules/iam/" | grep -v "_test\.go"`
   → **0 matches** (exit 1). Site #1 closed; combined with F5.1, **H-G = 0**.
2. **Build** — `go build ./...` → `BUILD OK` (clean; all 8 call sites updated: bootstrap, 2 in
   main.go, 6 tests).
3. **Unit tests** — `go test -count=1 ./internal/modules/auth/... ./internal/modules/iam/...` →
   all `ok` (auth application/delivery/domain/memory/postgres; iam application/authz/delivery/
   domain/memory/postgres/presence). Query-parity + Noop tests green.
4. **Integration compile** — `go vet -tags integration ./internal/modules/iam/infrastructure/postgres/`
   → `VET OK` (live test compiles under the build tag).

## Fixture-vs-real (honest labeling)

- **Live parity test: PROVIDED, NOT RUN here.**
  `go test -tags integration -run TestUserTenantRepository_UserTenantIDs_Live ./internal/modules/iam/infrastructure/postgres/`
  → **SKIP** (`no DATABASE_URL or METALDOCS_DATABASE_URL set`). A `metaldocs-postgres` container is
  up on host :5433, but its credentials are sourced from `.env` (compose interpolates
  `${POSTGRES_USER}/${POSTGRES_PASSWORD}`); per CLAUDE.md secrets are never read/printed, so no DSN
  was manufactured to drive the run. The test is committed and runs wherever `DATABASE_URL` is set
  (CI, or the milestone-validator with DB access).
- **No fixture stood in for the live SQL.** Parity here rests on (a) by-construction identical SQL
  and (b) the unit query-pin — NOT on a fake claiming to be the live path. The forbidden-list item
  ("no fixture-as-live-proof for F5.2") is respected: the live proof is labeled as deferred-to-DB,
  not asserted.

## Defers

- **Live parity run** → trigger: any environment with `DATABASE_URL`/`METALDOCS_DATABASE_URL` set
  (CI or validator). Owner: milestone-validator gate (C2 re-run from clean state). The test exists;
  only its execution is deferred to a DB-bearing environment.
