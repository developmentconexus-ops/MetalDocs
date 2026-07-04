# F3.1 — TxRunner chokepoint auto-seed + census-drift lint

> **Milestone:** M3 · **Binding contract:** `../validation-contract.md` §1 + §4 (api row). This spec
> distills the consumer contract; the contract governs on any conflict (HS-7).
> **Approval:** approved for implementation 2026-07-03 (main session, against runtime-truth census).

## Consumer contract (who consumes what)

The **consumer** is (a) the `FORCE ROW LEVEL SECURITY` backstop on 33 tenant tables, which needs
`metaldocs.tenant_id` seeded on every API business tx to engage; and (b) the `SEED-CHOKEPOINT` CI lint,
which needs manual `SeedTxIdentity` calls to have been collapsed to the chokepoint + a bounded allowlist.

**Required end-state:**
1. A **platform-layer ctx identity carrier** (tenant + actor), set by the API authn middleware
   (`internal/modules/auth/delivery/http/middleware.go`), readable by `internal/platform/db` **without
   importing `internal/modules/iam`** (module-boundary rule).
2. `TxRunner.Do` **and** `TxRunner.DoReadOnly` auto-seed `metaldocs.tenant_id` + `metaldocs.actor_id`
   (tx-local `set_config`) at tx begin **when identity is present in ctx**, **no-op when absent**
   (system paths stay NULL-permissive). Seed happens **before** `fn`, before any `authz.Require`, before
   any `FOR UPDATE`/advisory lock (H-PRE-1).
3. The **62** manual `SeedTxIdentity` call sites inside `Do`/`DoReadOnly` callbacks are **collapsed to
   0** outside the chokepoint, except an **explicit reasoned allowlist** (categories A actor≠ctx-actor,
   B cross-tenant platform-admin path — contract §1.4).
4. A **blocking** `scripts/api-lint` rule (`SEED-CHOKEPOINT`) fails on any `SeedTxIdentity` outside the
   chokepoint + allowlist (+ non-test).

## Non-goals (mandatory)

- **No RLS policy change** — do not touch `ENABLE`/`FORCE`/NULL-permissive/`WITH CHECK`/`FOR`. Seeding
  coverage only.
- **No async change** — worker/jobs seeding is F3.2 (`SeedTxTenant`), not this feature.
- **No new migration** — Go-only.
- **No actor-semantics rewrite** — sites seeding a genuinely distinct actor are allowlisted, not
  "fixed" to ctx actor.
- **No call-graph lint** — the census lint is a call-site AST scan (handler-local), same as M2 lints.

## Validation gate (positive + negative, captured output)

- **PG-1 (positive behavior):** chokepoint test — `Do` with identity-bearing ctx ⇒
  `current_setting('metaldocs.tenant_id')` == ctx tenant inside `fn`; identity-less ctx ⇒ empty GUC.
- **PG-2 (census):** `grep -rn "SeedTxIdentity(" --include=*.go internal apps | grep -v _test.go`
  outside chokepoint+allowlist = **0**; allowlist enumerated with file:line + reason.
- **PG-3 (lint negative):** add a throwaway `SeedTxIdentity` in an application service ⇒ `api-lint`
  **RED** naming file:line; remove ⇒ **GREEN**.
- **PG-4 (regression):** existing tenant-isolation suites green; `go build ./...` green; 5 authz lints +
  M2 tripwire lints green; cross-tenant 404 behavior unchanged.

## Named tests / proof commands

- `go test ./internal/platform/db/...` (chokepoint seed behavior)
- `go test ./scripts/api-lint/...` (new rule unit test, incl. negative fixture)
- `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 live violations)
- `go build ./...`
- targeted: `go test -tags integration -run 'Tenant|Isolation' ./...` (or authored + deferred per box)

## Interview record

| Q | A (from runtime truth / contract) |
|---|---|
| Does a chokepoint exist? | Yes — `internal/platform/db/runner.go` `Do`/`DoReadOnly`; seeds nothing today. |
| Where does identity live in ctx? | `platform/tenant` (tenant) + `iam/domain` (actor). Chokepoint may not import iam → need platform-level actor carrier. |
| Can autoseed be unconditional? | No — `SeedTxIdentity` requires non-empty actor+tenant; system paths have none → must no-op when absent (preserve NULL-permissive). |
| Why an allowlist, not census 0? | A few sites seed a stored actor (GrantedBy/assignedBy) or run cross-tenant (platform-admin). Contract §1.4 categories A/B. |
| New primitive needed? | No for F3.1 (reuse `SeedTxIdentity` SQL at chokepoint). `SeedTxTenant` is F3.2. |
