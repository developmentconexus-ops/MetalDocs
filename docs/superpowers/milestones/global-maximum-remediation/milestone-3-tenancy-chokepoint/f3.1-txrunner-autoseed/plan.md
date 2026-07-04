# F3.1 plan — TxRunner chokepoint auto-seed + census-drift lint

> Engine: `superpowers:subagent-driven-development` (TDD). Contract: `../validation-contract.md` §1.
> Failing test first → implement → green → review. No RLS policy change; Go-only; no new migration.

## Task list (ordered)

### T1 — Platform actor carrier (extend `internal/platform/tenant`)
- Add `WithActorID(ctx, actorID) context.Context` and `ActorFromContext(ctx) (string, error)` to
  `internal/platform/tenant/context.go` (mirror the existing tenant carrier; distinct ctx key;
  `ErrActorMissing` sentinel; empty/whitespace → missing).
- **TDD:** extend `context_test.go` — set/get round-trip; absent → error; empty → error.
- Rationale: `internal/platform/db` may import `internal/platform/tenant` (platform→platform, allowed by
  `check-module-boundaries.ps1`) but NOT `internal/modules/iam`. Actor must live in a platform package.

### T2 — Authn middleware sets the actor carrier
- In `internal/modules/auth/delivery/http/middleware.go` (~line 106, beside `WithTenantID`), add
  `ctx = platformtenant.WithActorID(ctx, currentUser.UserID)`.
- **TDD:** middleware test asserts both tenant + actor are on the outgoing ctx for an authed request.

### T3 — Chokepoint auto-seed (`internal/platform/db/runner.go`)
- In `do(ctx, opts, fn)`, **after** `BeginTx` and **before** `fn(tx)`: read
  `tenant.FromContext(ctx)` + `tenant.ActorFromContext(ctx)`. If **both** present & non-empty, execute the
  tx-local seed (same SQL as `authz.SeedTxIdentity`: `SELECT set_config('metaldocs.tenant_id',$1,true),
  set_config('metaldocs.actor_id',$2,true)`). If **either** absent → **no-op** (no error). Applies to
  BOTH `Do` and `DoReadOnly` (SET LOCAL is RO-safe).
- Seed occurs before any `authz.Require`/lock in `fn` (H-PRE-1 — seed is a config write, not a recording
  read; add NO `authz.Require`).
- **Do NOT** import `internal/modules/iam` — inline the two `set_config` calls in `platform/db`.
- **TDD (PG-1):** sqlmock test — identity-bearing ctx ⇒ the seed `ExecContext`/`QueryContext` is issued
  with the ctx tenant+actor; identity-less ctx ⇒ **no** seed statement. Both `Do` and `DoReadOnly`.

### T4 — Collapse the 62 manual `SeedTxIdentity` sites (behavior-preserving)
- For each `SeedTxIdentity(ctx, tx, X, Y)` **inside a `TxRunner.Do`/`DoReadOnly` callback**:
  - If `X` is provably the ctx tenant AND `Y` is provably the ctx actor (i.e. derived from the same
    request identity — e.g. `tenantID`/`actorID` params threaded from `tenantIDFromReq`/`userIDFromReq`,
    or `cmd.TenantID`/`cmd.ActorUserID`, or `req.TenantID`/`req.<Action>By` where `<Action>By` == the
    current actor) → **remove** the call (chokepoint now seeds it).
  - If `Y` is a **stored/distinct actor** not equal to the current ctx actor
    (`grantedByActor(GrantedBy)`, `assignedBy`, `d.CreatedBy`, `adminID` where it is a target not the
    caller) → **keep** the call and add it to the allowlist (category A) with a one-line reason. When in
    doubt, **allowlist rather than remove** (fail-closed: never change a seeded value).
  - Sites NOT inside a `TxRunner` callback (raw `BeginTx`, or infra repos that receive an external `tx`)
    → assess: if the enclosing tx flows through the chokepoint, remove; else allowlist (category B or a
    recorded reason).
- Produce `internal/platform/db` (or `scripts/api-lint`) **allowlist file** `seed-chokepoint-allowlist.txt`
  (mirror `scripts/api-lint/tripwire-allowlist.txt` format: `path:line  # reason`).
- **Verification per removal:** the enclosing handler/service still compiles and its existing test (if
  any) passes; the seeded identity is unchanged (ctx-derived).

### T5 — `SEED-CHOKEPOINT` blocking lint (`scripts/api-lint`)
- New rule (mirror `tripwire_arm_rules.go` structure; register in `RunCodeRules`/`RunRegistryRules` so it
  runs under the blocking `api-design-system-lint` job). AST-scan for `authz.SeedTxIdentity(` call
  expressions; violation if the call is **outside** the chokepoint file(s) AND outside
  `seed-chokepoint-allowlist.txt` AND outside `_test.go`. Emit file:line.
- **TDD (PG-3):** unit test with a positive fixture (allowlisted/chokepoint = clean) and a negative
  fixture (a stray `SeedTxIdentity` in a non-allowlisted app-service file = 1 violation). Mirror
  `tripwire_arm_rules_test.go` + `testdata`.

### T6 — Gate run + evidence
- `go build ./...`; `go test ./internal/platform/db/... ./internal/modules/auth/... ./scripts/api-lint/...`;
  `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 violations);
  census grep = 0 outside chokepoint+allowlist; capture negative-lint RED then GREEN.
- Targeted tenant-isolation integration (`-run 'Tenant|Isolation'`) or authored+deferred per box.

## Files expected to change
- `internal/platform/tenant/context.go` (+ test)
- `internal/modules/auth/delivery/http/middleware.go` (+ test)
- `internal/platform/db/runner.go` (+ test)
- ~15 module files across documents/controlleddocuments/templates/taxonomy/iam/tokens (site removals)
- `scripts/api-lint/` new rule + test + testdata; `seed-chokepoint-allowlist.txt`

## Risk / ordering notes
- T4 is the review-critical step. Bias to **allowlist over remove** on any ambiguity — a wrong removal
  changes the seeded tenant/actor (RLS/authz correctness). The lint (T5) enforces the census either way.
- H-PRE-1: never add `authz.Require` to a locked tx; the seed is `SET LOCAL` only.
