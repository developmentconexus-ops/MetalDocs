# ADR 0007 — Two-Tier Authorization

> **Status:** accepted 2026-05-03; amended 2026-05-05 (J2 wiring); amended 2026-05-10 (codegen rejected); amended 2026-05-11 Plan 5 (tier-2 + tier-3 tripwire extended to all regulated modules: IAM, documents, registry, taxonomy, templates)
> **Last verified:** 2026-05-11
> **Scope:** Authorization boundary between HTTP middleware (tier 1) and in-transaction area checks (tier 2).
> **Out of scope:** Authentication; Role/capability table definitions — see `wiki/modules/iam.md`.
> **Key files:**
> - `internal/modules/iam/application/capability_service.go:31` — tier-1 `CanDo` implementation
> - `internal/modules/iam/authz/authz.go:44` — tier-2 `Require` implementation; system_admin bypass at :58
> - `internal/modules/iam/authz/context.go:13` — typed errors `ErrActorContextMissing` / `ErrTenantContextMissing`; GUC helpers `MustActorID` at :21, `MustTenantID` at :34
> - `internal/modules/iam/infrastructure/postgres/role_provider.go:19` — `RolesByUserID` filters by `tenant_id` (Group B B5/B6 fix)
> - `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:20` — `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` all tenant-scoped (Group B fix)
> - `internal/platform/tenant/const.go:4` — `DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` sentinel UUID for single-tenant dev/test mode (extracted from bootstrap/api.go by c4a7d9a9)

## Context

MetalDocs has two authz services that each consult `role_capabilities`:

1. `CapabilityService.CanDo(ctx, userID, tenantID, capability)` — reads `iam_user_roles`. Used by HTTP middleware to gate route access.
2. `authz.Require(ctx, tx, capability, areaCode)` — reads `user_process_areas`. Used inside service-layer transactions to gate area-scoped writes (signoff, area-bound state changes).

The 2026-05-02 IAM unification plan attempted to consolidate both. It unified the middleware path (`StaticAuthorizer` → `CapabilityService`) but left `authz.Require` reading `user_process_areas`. This produced apparent dual systems and confused engineers.

## Decision

Treat the two services as **distinct tiers** with explicit responsibilities, not as a unification gap.

| Tier | Service | Table | Used by | Question answered |
|---|---|---|---|---|
| 1 — Tenant | `CapabilityService.CanDo` | `iam_user_roles` JOIN `role_capabilities` | HTTP middleware | "Can user X do `doc.create` in tenant T?" |
| 2 — Area | `authz.Require` | `user_process_areas` JOIN `role_capabilities` | Service layer (in-tx) | "Can user X sign for area QA-01?" |

`role_capabilities` is shared: both tiers map a role code to a capability set. Roles in `iam_user_roles` are tenant-scoped; roles in `user_process_areas` are area-scoped.

`system_admin` bypasses both tiers.

`authz.Require` accepts `areaCode = "tenant"` as a degenerate form: skips the area filter, behaves like a tier-1 check inside the transaction.

## Consequences

- **Positive:** matches the QMS domain (ISO 9001 segregation requires per-area approver grants); clear contract; no schema migration needed.
- **Negative:** IAM admin UI must distinguish "tenant role" assignment from "area membership" assignment — separate flows.
- **Open:** future area-membership service consolidation if area grants diverge from `user_process_areas`.

## Amendment — J2: `document.create` wiring (2026-05-05)

**Bug J2** (`wiki/bugs/audit-2026-05-04.md`) found that `permissiveAuthzChecker` in `main.go` returned `nil` for every capability check, bypassing tier-1 enforcement on `document.create`.

**Resolution (commits 563d650e, c0fa0485, 7b26f3cd, 1cebea64):**

- `AuthorizationChecker` interface removed from the documents module.
- New consumer port `CapabilityChecker` declared in `internal/modules/documents/application/ports.go` — one-method interface matching `CapabilityService.CanDo`.
- `apps/api/internal/wiring/documents.go:24` — `NewCapabilityChecker` adapter bridges `*iamapp.CapabilityService` to `docsapp.CapabilityChecker` (string capability conversion).
- `apps/api/cmd/metaldocs-api/main.go:275` — `Caps: wiring.NewCapabilityChecker(capabilityService)` replaces the former `permissiveAuthzChecker`. `permissiveAuthzChecker` struct deleted.

Documents module wiring was also lifted out of `main.go` into `apps/api/internal/wiring/documents.go` as part of this change (god-file reduction).

## Amendment — Codegen rejected (2026-05-10)

A `cmd/authzgen` spike was evaluated as a way to wrap `authz.Require` calls per operation from the OpenAPI spec. Rejected because:

1. `authz.Require` is tx-coupled (`internal/modules/iam/authz/authz.go:44` takes `*sql.Tx`). The proposed wrapper sits at the StrictServerInterface boundary — before the transaction opens — so it cannot supply tx.
2. The Postgres tripwire trigger (`migrations/0142b_role_capabilities_v2_enforce.sql:138-172`) reads the `metaldocs.asserted_caps` GUC on every mutation and rejects if the cap is absent. The static enforcement guarantee already exists at the database layer; codegen would duplicate the check with zero additional safety.
3. Industry pattern (Stripe, GitHub, AWS IAM) is lint + runtime, not generated per-op authz wrappers.

Replaced with two CI lint rules: `authz-call-present` (every op with `x-authz-area` has a matching `authz.Require` call in the handler) and `tripwire-pairing` (every mutating SQL statement in repositories pairs with an `authz.Require` call). Implemented in `scripts/api-lint/code_rules.go`.

Full spike notes: `docs/superpowers/notes/2026-05-10-authz-codegen-feasibility.md`.

## References

- `internal/modules/iam/application/capability_service.go` — tier 1
- `internal/modules/iam/authz/authz.go` — tier 2
- `internal/modules/iam/authz/context.go` — typed GUC errors (Group B fix)
- `internal/modules/iam/infrastructure/postgres/role_provider.go` — tenant-scoped `RolesByUserID` (Group B fix)
- `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` — tenant-scoped admin ops, DELETE-then-INSERT (Group B fix)
- `apps/api/internal/wiring/documents.go:24` — `NewCapabilityChecker` adapter (J2 fix)
- `scripts/api-lint/code_rules.go` — `authz-call-present` + `tripwire-pairing` lint rules (codegen-rejection amendment)
- Migration 0142b — Postgres tripwire trigger (`enforce_capability_asserted`) on `approval_instances` + `approval_signoffs`
- Migration 0162 — added `tenant_id` to `iam_user_roles`
- Migration 0165 — reseeded `role_capabilities`
- Migration 0170 (`migrations/0170_dev_approver_role_correction.sql`) — corrects dev approver role after 0166 over-promotion
- Audit 2026-05-03 — bugs B1-B6 (`wiki/bugs/audit-2026-05-03.md` lines 111-136)
- Audit 2026-05-04 — bug J2 (`wiki/bugs/audit-2026-05-04.md`)
- Tests: `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go`, `tests/integration/iam/tenant_isolation_test.go`

## See also

- [`wiki/modules/iam.md`](../modules/iam.md) — full Arc42 + C4 living architecture doc for `internal/modules/iam`; two live tiers (AuthorizationService deleted Plan 4 — T-003 closed)
- [`wiki/modules/documents.md §8.1`](../modules/documents.md#81-authentication--authorization) — documents consumer: tier-1 role gate, tier-2 `authz.Require` in all 5 `documents` mutations, tripwire on `documents` + approval tables (T-003 closed Plan 5)
- [`wiki/modules/auth.md`](../modules/auth.md) — canonical auth module doc; §8.1 covers how auth's middleware injects `iamdomain.WithAuthContext` so tier-1 and tier-2 checks have an actor; tier-0 session enforcement sits here, upstream of both authz tiers
- [`wiki/modules/registry.md §8.1`](../modules/registry.md#81-authentication--authorization) — registry now conforms: tier-2 `authz.Require` wired for Create/CreateTx + changeStatus (Obsolete/Supersede); tier-3 tripwire on `controlled_documents` + `cd_sequence_counters` (T-001/T-004 closed Plan 5)
- [`wiki/modules/taxonomy.md §8.1`](../modules/taxonomy.md#81-authentication--authorization) — taxonomy partially conformant Plan 5: PATCH dispatcher fixed (T-003 closed); `authz.Require` wired in Create/Update methods + tripwire on all 3 tables; archive/deactivate paths still tier-1 only (T-006 partial)
