# Feature F7.2 — Spec: Tenant onboarding API

> **Milestone:** 7 — Tenant Lifecycle Kernel  ·  **Folder:** `f7.2-onboarding`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-05 / operator-locked via ADR 0070 (onboarding=API fork) + validation-contract §1 — no implementation before this spec.

> Contract, written before any code. Validator judges F7.2 against this file + validation-contract §1.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Onboarding path — API or runbook? | **API route (contract-first)** — operator-locked 2026-07-05 (ADR 0070 decision 1). |
| 2 | Remaining shape questions? | **None needed** — the consumer contract is fully enumerated in validation-contract §1 (binding, HS-7) + ADR 0070; runtime surface mapped 2026-07-05 (iam service/repo/router patterns, capability 10-touchpoint incl. M2 gen-tripwire registry, login prerequisites incl. `auth_identities`). |

## Consumer contract (FIRST)

- **Consumers:**
  1. **Operator/admin tooling** calling `POST /api/v1/tenants` (system_admin session) — expects 201 + `{tenant_id, admin_user_id}`; 409 on slug conflict; 403 without `tenant.onboard`; RFC 9457 problems.
  2. **The onboarded admin** — must be able to `POST /api/v1/auth/login` with the issued credential and then perform a capability-gated action for the new tenant (the F7.2 end-to-end bar).
  3. **F7.3** — consumes the tenant row + (later) per-tenant crypto key provisioning hook inside `OnboardTenant`.
- **Contract:**
  - Request `OnboardTenantRequest`: `name` (tenant display name, non-blank), `slug` (URL-safe, unique), `admin_user_id` (text — iam_users PK is TEXT, memory `tokens-actor-id-text-contract`), `admin_display_name`, `admin_password` (initial credential; returned never, logged never) — exact fields per openapi schema authored in this feature.
  - Response 201 `OnboardTenantResponse`: `tenant_id` (uuid), `admin_user_id` (text).
  - Errors: 400 validation / 401 unauthenticated / 403 missing cap / 409 slug or user conflict / 500 — all `application/problem+json`.
  - **End-state (binding, contract §1.3):** new tenant's admin can login (auth_identities + iam_users + iam_user_roles + role_capabilities all satisfied) and perform ≥1 capability-gated action; cross-tenant URL → 404.
- **Source of truth:** `api/openapi/v1/openapi.yaml` (authored here, operationId `onboardTenant`, tag `iam`) → regenerated `iam/api/api.gen.go`; validation-contract §1.

## What this feature implements

1. **Contract:** `POST /tenants` in openapi.yaml (schemas `OnboardTenantRequest`/`OnboardTenantResponse`, tag `iam`) + `go generate ./internal/modules/iam/api/`.
2. **Capability `tenant.onboard`** — full 10-touchpoint walk: const+registry (`model.go`), `ScopeTenant` classify (`capability_scope.go`), tier-1 rule `POST /api/v1/tenants` → `CapTenantOnboard` (`permissions.go`, `VisibilityPermissionGuarded`), tier-2 `authz.Require` in-tx, seed grant to `system_admin` (`0001_product_reference_data.sql`), **tripwire arm via M2 registry** (`internal/platform/tripwire/registry.go` + gen-tripwire migration: `tenants` INSERT requires `tenant.onboard`), guard tests, `TestCapabilityRegistrySize` 35→36 (F7.3 lifts to 38), REQ-AUTHZ-5 coherence, H-PRE-1 (audit record off any lock-holding tx).
3. **`iam` application service `OnboardTenantService`** (pattern: `AdminService`) — one `TxRunner.Do` tx: `authz.WithCapCache` + `SeedTxIdentity` + `Require(tenant.onboard,"tenant")`; INSERT `metaldocs.tenants`; INSERT `iam_users` (admin, new tenant_id); INSERT `iam_user_roles` (`system_admin` for that tenant); create `auth_identities` row (bcrypt hash of `admin_password` — reuse auth module's hashing, via its published interface, never inline crypto); `audit.RecordTx` `tenant.onboarded` event. Cache invalidation post-commit.
4. **Handler + wiring:** iam delivery handler for the generated op; composition-root wiring in `main.go` (pattern: `iamAdminService` site ~line 243).
5. **Key-provisioning seam:** `OnboardTenantService` exposes/accepts a `TenantKeyProvisioner` port (no-op in F7.2; F7.3 plugs the real envelope) — so contract §1.2(c) is satisfied at milestone close without F7.2 blocking on F7.3's crypto.

## Non-goals (mandatory)

- No per-tenant crypto key implementation (F7.3 — only the provisioner seam lands here).
- No export/erasure anything (F7.3), no RLS role changes (F7.4).
- No tenant update/offboard/suspend routes; no self-serve signup UI; no invite-email flow.
- No FE consumer work.
- No change to existing login/auth flow beyond creating the identity row via the existing mechanism.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Registry + scope guards green after cap add (36) | `go test ./internal/modules/iam/domain/ -run 'TestCapabilityRegistrySize|TestEveryCapabilityClassified'` | real |
| Contract regen clean, build green | `go generate ./internal/modules/iam/api/` + `go build ./...` (empty diff beyond intended) | real |
| Onboarding tx creates tenant+admin+role+identity+audit atomically | `TestOnboardTenant_CreatesTenantAdminAndAudit` (integration, testdb factory) | real (Postgres) |
| 409 on duplicate slug; 403 without cap (negative authz) | `TestOnboardTenant_DuplicateSlugConflict`, `TestOnboardTenant_RequiresCapability` | real |
| Tripwire arm: tenants INSERT without asserted cap → P0001 | `TestTenantsInsertTripwire` (negative proof) + M2 drift-check green | real |
| **End-to-end:** onboard → login as new admin → capability-gated action succeeds; cross-tenant → 404 | `TestOnboardTenant_EndToEnd_LoginAndAct` (integration) + **live QA drive** (start-api.ps1 -Build; curl POST /tenants → POST /auth/login → gated action; captured) | real |

TDD: failing test first, implement to green. Live-drive proof captured in evidence.md (§6.2 of contract).

## ADR needed?

- [x] Durable decision already recorded → **ADR 0070** (Accepted 2026-07-05). No new ADR.
