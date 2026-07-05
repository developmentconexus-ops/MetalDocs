# Feature F7.2 — Plan

> **Spec:** `./spec.md` (approved pre-code) · **Binding contract:** `../validation-contract.md` §1, §5
> **Execution:** subagent-driven (fresh implementer per task, sonnet; main orchestrates+reviews+commits)

## Plan

### Task A — Capability `tenant.onboard` (10-touchpoint, mechanical + guarded)
Files: `internal/modules/iam/domain/model.go` (const `CapTenantOnboard = "tenant.onboard"` + `validCapabilities`), `capability_scope.go` (`ScopeTenant`), `internal/modules/iam/domain/model_test.go` (`want` 35→36), `apps/api/cmd/metaldocs-api/permissions.go` (`POST /api/v1/tenants` → `CapTenantOnboard`, `VisibilityPermissionGuarded`), `db/reference-data/0001_product_reference_data.sql` (grant to `system_admin`), `internal/platform/tripwire/registry.go` (+ gen-tripwire → new migration arming `metaldocs.tenants` INSERT with `tenant.onboard`).
TDD: bump registry-size test first (RED) → add cap (GREEN); tripwire negative test (`tenants` INSERT w/o asserted cap → P0001) RED → arm migration GREEN. M2 drift-check green.

### Task B — Contract: `POST /tenants` + regen
Files: `api/openapi/v1/openapi.yaml` (path `/tenants` POST, operationId `onboardTenant`, tag `iam`; schemas `OnboardTenantRequest{name,slug,admin_user_id,admin_display_name,admin_password}` required-all, `OnboardTenantResponse{tenant_id uuid,admin_user_id}`; 201/400/401/403/409/500 problem refs — copy `/iam/users` POST shape), then `go generate ./internal/modules/iam/api/`. Contract lints (oasdiff compat, nullable⇒required) green.

### Task C — `OnboardTenantService` + repo + handler + wiring (the core)
Files: new `internal/modules/iam/application/onboard_tenant_service.go` (pattern `admin_service.go`); new repo method(s) in `internal/modules/iam/infrastructure/postgres/` (pattern `role_admin_repository.go:70-107`): one tx — `WithCapCache`+`SeedTxIdentity`+`Require("tenant.onboard","tenant")` → INSERT `metaldocs.tenants` → INSERT `iam_users` (tenant_id=new) → INSERT `iam_user_roles` (`system_admin`) → create `auth_identities` via auth module's published creation path (bcrypt; never inline crypto; check auth module port — if none published, add one to auth's port surface, NOT reach into its tables) → `audit.RecordTx("tenant.onboarded", payload{name,slug,admin_user_id})`. `TenantKeyProvisioner` port param, no-op impl (F7.3 seam). Handler for generated op in `internal/modules/iam/delivery/http/` (router.go registration); wire in `apps/api/cmd/metaldocs-api/main.go` (~line 243 site).
TDD: integration tests (testdb factory, `//go:build integration`): `TestOnboardTenant_CreatesTenantAdminAndAudit`, `_DuplicateSlugConflict`, `_RequiresCapability` — RED first.

### Task D — End-to-end + live drive
`TestOnboardTenant_EndToEnd_LoginAndAct` (integration: onboard → authenticate via auth service → gated action for the new tenant; cross-tenant URL → 404). Then live QA drive: `.\scripts\start-api.ps1 -Build` → curl `POST /tenants` (admin session) → `POST /auth/login` (new admin) → capability-gated action → capture request/response into evidence.md.

## Test strategy
Integration-first (this is DB+authz+contract work); unit only where a pure function appears. All new tests on the canonical testdb framework. Targeted `-run` filters (no full suite locally).

## Ordering / dependencies
A → B → C → D. A and B are parallelizable (different files); C needs both; D needs C.

## Risks
- Auth-identity creation may lack a published port → HS-2 boundary check; expected: auth module exposes an identity-creation application path (invite/managed-user flow exists — `createManagedUser` op) to reuse.
- Tripwire arm on `tenants` INSERT must not break bootstrap/reference-data seeding (seed runs as superuser/migration context, not app role — verify; if seed path trips, arm scopes to app-role context like existing arms).
