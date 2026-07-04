# F3.1 evidence — TxRunner chokepoint auto-seed + census-drift lint

> Contract: `../validation-contract.md` §1. Executed via subagents (TDD); main session reviewed the
> aggregate diff, found + fixed one over-removal, and commits. Fixture proof labeled; RLS integration
> proof is F3.2's negative drive (this feature adds seeding coverage on the sync/API chokepoint).

## What shipped

**Mechanism (T1–T3):**
- `internal/platform/tenant/context.go` — `WithActorID`/`ActorFromContext`/`ErrActorMissing` (actor now
  carried in a platform package so `internal/platform/db` can read it without importing `internal/modules/iam`).
- `internal/modules/auth/delivery/http/middleware.go:107` — `platformtenant.WithActorID(ctx, currentUser.UserID)`
  set beside the existing `WithTenantID` (single authn injection point).
- `internal/platform/db/runner.go` — `seedTxIdentityFromContext` in the shared `do()`: reads tenant+actor
  from the platform carrier; **both present** → tx-local `set_config('metaldocs.tenant_id'/'.actor_id', …, true)`
  (SQL inlined, no iam import); **either absent** → no-op (system/janitor paths stay NULL-permissive).
  Runs before `fn`, before any `authz.Require`/lock (H-PRE-1); applies to both `Do` and `DoReadOnly`.

**Collapse (T4):** 61 non-test `SeedTxIdentity` call sites → **41 removed** (provably ctx-identity:
`cmd.TenantID/cmd.ActorUserID`, threaded `tenantID/actorID` params, `req.<Action>By` == current caller)
across controlleddocuments/documents(+approval)/templates/documents-repository. **21 allowlisted**
(`scripts/api-lint/seed-chokepoint-allowlist.txt`): category A distinct-actor (`d.CreatedBy`,
`assignedBy`, `grantedByActor(GrantedBy)`), raw-`BeginTx` paths not reached by the chokepoint
(documents repository, user_area, taxonomy shared `setAuthzGUC` helper), and one category-B system path
(see review finding). **Census outside chokepoint+allowlist = 0.**

**Lint (T5):** `scripts/api-lint/seed_chokepoint_rule.go` — `checkSeedChokepoint` AST-scans for
`authz.SeedTxIdentity(` calls, exempting the chokepoint/definition files, `_test.go`, and allowlisted
`path:line`; also emits `SEED-CHOKEPOINT-ALLOWLIST-STALE` for dead allowlist lines. Registered blocking in
`RunCodeRules` (`code_rules.go`).

## Main-session review finding (fixed before commit)

The collapse **over-removed one seed**: `cancel_service.go` `cancelInstance` is shared by
`SystemCancelInstance`, which `internal/modules/jobs/stuck_instance_watchdog/job.go` calls on a
**carrier-less background ctx** (`BypassSystem`, `ActorUserID: SystemActor`). The chokepoint no-ops
without a ctx carrier, so removing that seed left the watchdog auto-cancel tx **unseeded** — a regression
vs pre-M3 (this path *did* seed), and it is **not** among F3.2's five seed sites. **Fix:** restored
`SeedTxIdentity(ctx, tx, in.TenantID, in.ActorUserID)` **inside the `if system` block** (the sync
`CancelInstance` path stays chokepoint-covered and needs no manual seed) and allowlisted the line
(category B, §2.4 stuck-watchdog per-instance action). Full carrier-less-reachability audit confirmed
this was the **only** over-removal: `SystemCancelInstance` is the sole dual sync/system service among the
collapsed set, and `stuck_instance_watchdog` is the only worker/jobs package importing a collapsed
application package. `RunScheduledPublishJob` uses its own private tx (a F3.2 seed site, never seeded).

## Gates (captured)

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go test ./internal/platform/tenant/... ./internal/platform/db/... ./internal/modules/auth/...` | ok (PG-1 seed behavior: both `Do`/`DoReadOnly` seed when tenant+actor present; no-op when either absent) |
| collapsed-module tests (controlleddocuments/documents/approval/templates/iam/taxonomy application) | ok (8 test fixtures updated to inject the platform tenant+actor carrier the chokepoint now reads — mechanical, behavior-preserving) |
| `go test ./scripts/api-lint/` | ok (SEED-CHOKEPOINT 7/7: clean-tree 0, stray fires named file:line, removal→0, chokepoint/`_test.go`/allowlisted exempt, stale-allowlist fires) |
| `go run ./scripts/api-lint -strict …` (live) | **0 violations** |
| Census (live non-test calls excl. definition + lint doc-comment) | **21 sites = 21 allowlist entries**; 0 outside allowlist |
| PG-3 negative | stray `SeedTxIdentity` in non-allowlisted file → lint **RED** (exit 1, names `zz_pg3_probe.go:11`); removed → **GREEN** (0, exit 0) |

## Fixture-vs-real
- PG-1 chokepoint proof uses a **fake `driver.Conn`** recording `set_config` execs (sqlmock-class) — proves
  the chokepoint *issues* the seed with the right args, **not** that Postgres RLS then blocks a cross-tenant
  row. The **real-DB RLS backstop proof is F3.2's `//go:build integration` negative drive** (contract §2.5).

## Bounded defers
- 2 pre-existing, unrelated wire-contract test failures (`controlleddocuments/delivery/http`
  `department_code` omitempty; `documents/delivery/http` comment wire-shape) — **not touched** (verified
  zero diff in those files this session); out of F3.1 scope.

## Contract conformance (§1.6 exit criteria)
Platform carrier ✓ · chokepoint seeds both `Do`/`DoReadOnly`, present→seed/absent→no-op, before Require &
locks ✓ · PG-1 positive proof ✓ · 61 sites → 0 outside chokepoint+allowlist ✓ · allowlist enumerated,
categories A/B + recorded raw-tx reason ✓ · `SEED-CHOKEPOINT` registered/blocking, GREEN clean / RED on
synthetic ✓ · module tests green, `go build` green ✓ · no RLS policy change ✓.
