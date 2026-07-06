# F7.2 — onboarding · evidence

> **Feature:** F7.2 (tenant onboarding API) · **Status:** CLOSED 2026-07-05
> **Spec:** `./spec.md` (approved pre-code) · **Binding contract:** `../validation-contract.md` §1, §5
> **Commits:** `098618c6` (Tasks A+B), `50ce3bc8` (Task C), this commit (Task D e2e test + pgx conflict-mapping fix + evidence)

## Deliverables

| Piece | Where |
|---|---|
| Capability `tenant.onboard` (registry 35→36, ScopeTenant, catalog, tier-1 rule `POST /api/v1/tenants`, system_admin seed grant) | `iam/domain/model.go`, `capability_scope.go`, `catalog.go`, `permissions.go`, `0001_product_reference_data.sql` |
| Tripwire arm #19 `tenants`/INSERT ← `tenant.onboard`, generated via M2 registry (`gen-tripwire`) | `internal/platform/tripwire/arms.go`+`render.go`, migration `db/migrations/0277_tenants_insert_tripwire_onboard_cap.sql` (applied to dev DB) |
| Contract `POST /tenants` (operationId `onboardTenant`, tag iam, `OnboardTenantRequest/Response`, 201/400/401/403/409/500 problem refs) + regen | `api/openapi/v1/openapi.yaml`, `iam/api/api.gen.go` |
| `OnboardTenantService` — one tx: `SeedTxIdentity(actor tenant)` → `Require(tenant.onboard)` → `Require(user.manage)` → `SeedTxTenant(new tenant)` → INSERT tenants → `TenantKeyProvisioner` seam (no-op; F7.3 plugs envelope) → auth identity (`authpg.CreateUserTx`, hash via exported `authapp.HashPassword`) → iam_users + iam_user_roles(system_admin) → `audit.RecordTx tenant.onboarded` | `iam/application/onboard_tenant_service.go`, `iam/infrastructure/postgres/tenant_repository.go`, `iam/delivery/http/tenant_handler.go`, `main.go` wiring |
| Integration tests (canonical testdb framework) | `tests/integration/iam/tenants_tripwire_test.go`, `onboard_tenant_test.go`, `onboard_tenant_e2e_test.go` |

## Gate results (spec Validation Gate table)

| Criterion | Proof | Result |
|---|---|---|
| Registry + scope guards green (36) | `go test ./internal/modules/iam/domain/` (RED at 35≠36 first, then GREEN — TDD order kept) | PASS |
| Contract regen clean, build green | `go generate ./internal/modules/iam/api/` + `go build ./...`; oasdiff no-breaking PASS; redocly PASS; `SHAPE-NULLABLE-NOT-REQUIRED` 0; `TRIPWIRE-ARM-PARITY`/`-DRIFT` 0 | PASS |
| Onboarding tx atomic (tenant+admin+role+identity+audit) | **Live drive** (below): all 5 rows present after 201; failed attempts (dup slug / dup user) left **0 orphans** (`orphan_user|0`, `orphan_tenant|0`) | PASS (live) |
| 409 duplicate slug; 403 without cap | Live: dup slug → **409** `CONFLICT_ERROR`; dup admin user → **409**; unauthenticated → **401**. `TestOnboardTenant_DuplicateSlugConflict` / `_RequiresCapability` compile-verified, run deferred (below) | PASS (live) |
| Tripwire negative proof | Live SQL (dev DB, migration 0277): tenants INSERT without asserted caps → **P0001** `ErrCapabilityNotAsserted: one of {tenant.onboard} required`; with `tenant.onboard` asserted tx-local → INSERT succeeds (both rolled back). M2 drift/parity lints 0 violations | PASS (live) |
| End-to-end onboard → login → gated action; cross-tenant 404 | **Live QA drive** below + `TestOnboardTenant_EndToEnd_LoginAndAct` (compile-verified, run deferred) | PASS (live) |

## Live QA drive (mission D4 — runtime proof, 2026-07-05)

`.\scripts\start-api.ps1 -Build` (PowerShell; `.env` never read/printed). API `:8081`. Seed creds from wiki (`local-dev-startup.md`), not `.env`.

1. `POST /api/v1/auth/login {identifier:"admin"}` → 200; capability list **contains `tenant.onboard`** (seed grant live).
2. `POST /api/v1/tenants` without session → **401** `AUTH_UNAUTHORIZED`.
3. `POST /api/v1/tenants` (admin session + same-origin header; CSRF origin guard enforced — missing Origin → 403 `FORBIDDEN_ORIGIN`, observed) → **201** `{"admin_user_id":"acme-admin","tenant_id":"1d01f02f-3d04-4095-97f8-3b3a845c3832"}`.
4. `POST /auth/login {identifier:"acme-admin"}` (initial password) → **200**, `tenant_id` = new tenant, `tenant_name":"Acme Metals QA"`, `must_change_password:true`, `roles:["system_admin"]`, full capability set.
5. Gated action before password change → **403** `AUTH_PASSWORD_CHANGE_REQUIRED` (fail-closed first-access gate, correct).
6. `POST /auth/change-password` (first-access flow) → 200 `changed:true`; old session rotated (next call 401, correct); re-login with new password → 200.
7. **Capability-gated action:** `GET /api/v1/iam/users` as acme-admin → **200**, items = exactly 1 (acme-admin itself) — tenant-scoped listing, no foreign users visible.
8. **Cross-tenant:** `GET /api/v1/documents/ec200048-…` (System-Tenant document) from acme session → **404** `NOT_FOUND`.
9. Negatives after fix: duplicate slug → **409** "Tenant slug already exists"; duplicate admin user (fresh slug) → **409** "Admin user already exists".
10. DB proof (container psql): `tenants` row (id/name/slug), `iam_users` acme-admin under new tenant, `iam_user_roles` system_admin, `audit_events` `tenant.onboarded` actor=admin tenant=new payload `{name,slug,admin_user_id}` **only** (no password material). Orphan counts after both failed onboards: 0/0.

## Review/QA disposition (subagent-driven; main orchestrated, reviewed, committed)

- Task A implementer initially stopped claiming `metaldocs.tenants` doesn't exist — orchestrator refuted (baseline `0001_current_schema.sql:1601`, agent had searched only `db/migrations/`), resumed, completed.
- Task C orchestrator review caught **3 defects pre-commit** (all masked by not-yet-runnable integration tests):
  1. `SeedTxIdentity` seeded the NEW tenant id — grants are tenant-scoped → every call would 403. Fixed: actor tenant via `tenant.FromContext` (admin_handler precedent) + new `actorTenantID` param.
  2. Only `tenant.onboard` asserted, but `iam_users`/`iam_user_roles` arms require `user.manage` → P0001 mid-tx. Fixed: second `Require(user.manage)` before any INSERT.
  3. Tenant GUC left on actor tenant while provisioning rows carry the new tenant (breaks under F7.4 real RLS) + bcrypt cost const duplicated in iam. Fixed: `SeedTxTenant(new tenant)` after authz; exported `authapp.HashPassword` wrapper (single crypto policy).
- Live drive caught **1 more defect**: duplicate slug → 500 not 409. Root cause **class defect**: unique-violation detection via `*pq.Error` while the runtime driver is pgx (`otelsql.Open("pgx", …)`) — `errors.As` never matched. Fixed in `tenant_repository.go` **and** the same latent bug in `auth/infrastructure/postgres/repository.go` (dual pgconn/pq check, canonical `MapPgError` idiom). Re-driven: both conflict paths 409.
- Task C TDD-order deviation disclosed by implementer (impl before tests, mid-task compaction); tests were then review-corrected and the live drive supplied the runtime RED→GREEN equivalent (500→409 observed and fixed).

## Bounded defers

| Defer | Why | Trigger |
|---|---|---|
| ~~Run `go test -tags=integration ./tests/integration/iam/... -run 'TestTenantsInsertTripwire|TestOnboardTenant'` (4 tests)~~ **CLOSED 2026-07-05 (F7.3 Task F session) — with correction**: the earlier close note here claimed "all 5 tests PASS in 7.8s"; that run was a **false green** (DATABASE_URL unset → `testdb.DSN` t.Skip → `ok` was skips, discovered during F7.3 Task F). The suite was then run for REAL (`.env`-loaded DATABASE_URL, PowerShell script pattern): first real run exposed 7 latent test defects across the iam suite (stale `doc.*` capability names, hardcoded-tenant FK violations, pgx multi-command seeds, template-DB missing dev-seed user, slug-leak cleanup gap) — all repaired in commit `49a9fed0`; full `./tests/integration/iam/...` now **PASS live (156s)**, incl. all onboarding/tripwire tests. Prerequisite repair commit d9d21719 stands | — | closed |
| `admin_user_id` naming: login identifier = user_id (no separate username field) | Matches iam_users TEXT PK contract; revisit only if a distinct username surface ships | Product decision |
