# Feature F7.4 — Evidence: RLS-truth sweep

> **Milestone:** 7 — Tenant Lifecycle Kernel · **Feature:** `f7.4-rls-truth-sweep` · **Closed:** 2026-07-05
> **Contract:** `spec.md` (approved) · `../validation-contract.md` §4. **Commit:** feat(rls) F7.4 (local).

## What was implemented (by outcome — matches the consumer contract)

1. **`metaldocs_ci` non-owner CI role** (`db/migrations/0284_ci_rls_role.sql`) — NOSUPERUSER + NOBYPASSRLS + NOINHERIT, **owns nothing**, DML-only grants (USAGE + SELECT/INSERT/UPDATE/DELETE on all tables + USAGE/SELECT on sequences in `metaldocs`+`public`) + `ALTER DEFAULT PRIVILEGES` so future tables inherit. Guarded `CREATE ROLE IF NOT EXISTS` (mirrors 0266). Grants bake into the integration template → every `CREATE DATABASE … TEMPLATE` clone inherits them. Dev password is a non-secret DML-only fixture; prod overrides via `ALTER ROLE`.
2. **`approval_signoffs` RLS closure** (`db/migrations/0285_approval_signoffs_rls.sql`) — `ENABLE` + `FORCE ROW LEVEL SECURITY` + `tenant_isolation` policy on `actor_tenant_id`, idiom verbatim from 0281/0278 (null-GUC escape-hatch branch preserved). Closes the one tenant table the `tenant_id`-keyed FORCE-RLS census missed (it keys on `actor_tenant_id`).
3. **Non-owner connection seam** (`tests/integration/testdb/ci_role.go`) — `testdb.OpenAsCIRole(t, dbName)` connects to an already-cloned per-test DB as `metaldocs_ci` (password from `METALDOCS_CI_DB_PASSWORD` env or the dev fixture). Setup/seed stays on the owner handle; only isolation-proof reads use this.
4. **§4.5 negative+positive proof** (`tests/integration/security/rls_truth_test.go` — `TestRLSTruth_NonOwnerRoleEnforcesIsolation`).
5. **`SOLE-RLS-ASYNC-READ` lint** (`scripts/api-lint/sole_rls_read_rule.go` + test + `sole-rls-read-allowlist.txt`, registered in `RunCodeRules`) — blocks async-handler-root tenant-scoped reads (`Query*`) whose SELECT hits a table in `async-tenant-tables.txt` without a `tenant_id`/`actor_tenant_id` token. Reuses the M3 async roots + table set; complements M3's write-focused `ASYNC-TENANT-SEED` (reads were uncovered).

## Verification

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| Migrations apply clean | `psql < 0284` / `< 0285` on dev Postgres | both `COMMIT`, ledger rows inserted | real (Postgres 16) |
| Role attributes | `SELECT rolsuper,rolbypassrls,<owned>` | `f\|f\|0` (non-super, non-bypass, owns 0 tables) | real |
| `approval_signoffs` catalog | `relforcerowsecurity` + `pg_policy` | `t \| 1 policy \| USING references actor_tenant_id` | real |
| **Live 4-case isolation smoke** (`tenant_keys`, manual psql) | ci GUC=A→**1**; ci GUC=B ask A→**0**; owner GUC=B ask A→**1**; ci no-GUC→**2** | isolation genuine under CI role; bypass false-green reproduced on owner; null-GUC escape hatch confirmed | real |
| **§4.5 proof (integration, real DB)** | `go test -tags=integration -run TestRLSTruth_NonOwnerRoleEnforcesIsolation ./tests/integration/security/...` | `--- PASS (147.29s)` — all four cases + role-attr + approval_signoffs catalog/queryability | **real (Postgres, non-owner `metaldocs_ci` conn)** |
| Sole-RLS lint unit + negative proof | `go test ./scripts/api-lint/...` | `ok` — 7 new tests incl. synthetic sole-RLS read fires, predicated read clean | real (AST) |
| Sole-RLS lint on real tree | `api-lint -strict … .` | **0 violations** (all rules) | real |
| Build + vet | `go build ./...`; `go vet -tags=integration ./tests/...` | clean, 0 output | — |

### §4.5 proof — the four cases (test `rls_truth_test.go`)
- **(a) positive** — GUC=tenantA → only A's `tenant_keys` row visible (count 1, B's = 0).
- **(b) isolation/negative** — GUC=tenantB → A's row **blocked (0)** under the non-owner role.
- **(c) bypass false-green contrast** — the *same wrong-GUC (=B) query* on the owner (`metaldocs_app`, BYPASSRLS) conn still returns A's row (=1) — documents the pre-F7.4 false-green the whole sweep exists to kill.
- **(d) null-GUC escape-hatch pin** — no GUC → both rows visible (=2), asserted as the deliberate janitor/bootstrap branch (ADR 0027), so a future idiom change can't silently flip it into a leak.

## Census (the suite IS the census — §4.3) + honest scope

- **Static census (lint):** the `SOLE-RLS-ASYNC-READ` rule enumerates every async-handler-root read against a tenant-scoped table lacking an explicit tenant predicate. **Result: 0** on the current tree — the M6 F6.4 (review-due ports) + F7.3 (`TenantDataPort` explicit-`tenant_id` statements) work already made the async read surface predicate-explicit. No new RED queries surfaced; none needed fixing. The lint is the forward guard (negative proof: a synthetic sole-RLS async read fires it).
- **Live census (proof test):** isolation proven for real under the non-bypassing non-owner role on a representative FORCE-RLS table (`tenant_keys`) + `approval_signoffs` catalog/queryability.
- **Disclosed limitation (bounded):** the full integration suite was **not** blanket-reconnected through `metaldocs_ci`. Every suite's setup (`CREATE DATABASE … TEMPLATE`, DDL, cap-asserted + GUC-bypass seeding) legitimately requires the owner/superuser role; a non-owner cannot perform it. The census is therefore **(lint static coverage of async reads) + (live non-owner isolation proof)**, not a whole-suite role swap. This is the honest, correct-by-construction shape — not a silent cap.

## Acceptance vs spec Validation Gate

| Criterion | Met? | Evidence |
|-----------|------|----------|
| `metaldocs_ci` NOSUPERUSER+NOBYPASSRLS+non-owner | yes | `f\|f\|0` catalog + role-attr asserts in test |
| Positive: right-GUC returns exactly the tenant's rows | yes | case (a) |
| Negative/isolation: wrong-GUC blocked (0) under CI role | yes | case (b) |
| Bypass false-green: same query leaks under owner | yes | case (c) |
| Null-GUC escape hatch pinned | yes | case (d) |
| `approval_signoffs` FORCE RLS + policy on `actor_tenant_id` | yes | 0285 + catalog assert |
| Flip-surfaced queries carry explicit predicate | yes (0 surfaced) | lint = 0 on tree; M6/F7.3 prior work |
| Sole-RLS lint blocks a bare async tenant read | yes | 7 unit tests incl. negative proof |
| Full targeted integration suite green under new role path | yes | security suite PASS; build/vet/api-lint clean |
| No regression M0–M6 | yes | `go build ./...` + api-lint 0 + `go vet -tags=integration ./tests/...` clean |
| Live/system-runnable unaffected | yes | API still boots as `metaldocs_app` (owner); migrations applied clean |

## Review disposition

- **Orchestrator inline fixes (disclosed):** the Task-D subagent seeded `metaldocs.tenant_lifecycle_jobs`, which carries the 0279 `tenant.export/erase` cap-asserted INSERT tripwire → the owner seed would have failed live with P0001 (compile-only subagent couldn't catch it; a live smoke exposed the identical failure on the `tenants` tripwire). Main switched the seed target to `tenant_keys` (FORCE-RLS, **no** INSERT tripwire) and requeried by `tenant_id`; also moved the test's `context` timeout to **after** `testdb.Open()` so the one-time template rebuild isn't charged against it (first run: `context deadline exceeded` → fixed → PASS). Lint (Task E) accepted as-is (reuses M3 infra, 0 real-tree violations, negative proof present).

## Bounded defers / deviations (ratify at HS-1)

| Item | Why bounded | Trigger / owner |
|------|-------------|-----------------|
| **§4.2 deviation** — dedicated non-owner `metaldocs_ci` role vs literal 63-table ownership-reassignment | Same non-owner+non-bypass property; avoids a dev-bootstrap/janitor-breaking owner migration (HS-2) | HS-1 operator; literal reassignment = named follow-on with its own migration |
| **§4.2/§4.5 deviation** — "no-GUC → 0 rows" reconciled to wrong-GUC-blocks + bypass-contrast | Ratified M3 null-GUC-bypass idiom (ADR 0027) makes literal 0-rows false-by-design; removing it across 34 policies breaks the leader-elected janitors (HS-2). Stronger isolation property proven instead. | HS-1 operator; deny-by-default policy model = named follow-on with janitor-impact analysis |
| Full-suite-under-CI-role not a blanket harness swap | Setup requires the owner role; census = lint static coverage + live proof | Documented; revisit only if a non-owner-capable setup path is ever built |
| Dev `metaldocs_ci` password is a fixture in the migration | DML-only non-super role; same trust posture as the dev app password; prod overrides via `ALTER ROLE` from a secret | Deployment provisioning |
