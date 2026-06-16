# Feature F3.1 — Spec

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Folder:** `f3.1-wire-iam-user-options`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-16 / leandrotca.work — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Engine used: `superpowers:brainstorming` (invoked 2026-06-16, seed = F3.1 row of `milestone.md`).
Pre-interview context-read covered: the consumer port at `internal/modules/documents/application/iam_user_options.go`, the existing nil-safe adapter at `internal/modules/documents/module.go:139`, the missing wiring at `apps/api/cmd/metaldocs-api/main.go:414` (re-anchored — mission §5 path drifted from `cmd/metaldocs-api/main.go:413`), and the existing `auth.Service.ListUsers(ctx, tenantID)` at `internal/modules/auth/application/service.go:521` (already tenant-scoped via role-provider join).

| # | Question | Answer |
|---|----------|--------|
| 1 | What filter should the adapter apply when returning the tenant's user options to the fillin placeholder picker? | **Active only** — `mu.IsActive == true`. Deactivated users are almost always wrong as a "who" pick. |
| 2 | In what order should the list be returned? | **DisplayName ASC, case-insensitive; tie-break by UserID ASC.** Predictable for picker UX, deterministic for tests. |
| 3 | Where does the new adapter live (port-vs-adapter seam)? | **`apps/api/internal/wiring/`** — Composition Root pattern; sibling of `wiring.NewCapabilityChecker` / `NewDocumentsAuditSink` / `NewProfileDefaults`. Documents owns the consumer port; wiring owns the cross-module adapter. Honors milestone constraint *"no new IAM port, no IAM API change"* (HS-2 boundary); M4's IAM-owned port swap will be a one-line change at the same seam. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the documents-module placeholder catalog. Concrete path: `documents.PlaceholderOptionsHandler` (constructed at `internal/modules/documents/module.go:94`) calls `placeholderOptionsIAMAdapter.ListUserOptions(ctx, tenantID)` ([module.go:143](../../../../../internal/modules/documents/module.go:143)), which delegates to the injected `application.IAMUserOptionsReader` ([module.go:96](../../../../../internal/modules/documents/module.go:96)). End-consumer = the fillin placeholder picker asking for the user-type option list when the author binds a `${user}` placeholder.
- **Contract:** `IAMUserOptionsReader.ListUserOptions(ctx context.Context, tenantID string) ([]UserOption, error)` where:
  - `UserOption{UserID string, DisplayName string}` ([iam_user_options.go:5](../../../../../internal/modules/documents/application/iam_user_options.go:5)). Both fields non-empty for every returned element.
  - Returned slice contains **only** users belonging to `tenantID` AND with `IsActive == true`. Inactive or cross-tenant users MUST NOT appear.
  - Returned slice is **sorted** by `strings.ToLower(DisplayName)` ASC; ties broken by `UserID` ASC.
  - On empty result, returns a **non-nil empty slice** (`[]UserOption{}`) plus `nil` error — never `nil, nil`. Mirrors the existing adapter's empty-list semantics ([module.go:145](../../../../../internal/modules/documents/module.go:145)).
  - On underlying error, returns `nil, err` — error is propagated, not swallowed.
  - `ctx` cancellation propagates to the underlying auth call.
- **Source of truth for the contract:** consumer-defined port at `internal/modules/documents/application/iam_user_options.go:11` (port unchanged by this feature). Filter/order/empty semantics are NEW commitments above the port — recorded here as the binding contract the producer satisfies.

## What this feature implements

A new production adapter `wiring.DocumentsIAMUserOptions` (in package `apps/api/internal/wiring/`) that satisfies `documents.application.IAMUserOptionsReader` by wrapping the existing `auth.Service.ListUsers(ctx, tenantID)` ([auth/application/service.go:521](../../../../../internal/modules/auth/application/service.go:521)). Implementation responsibilities:

1. Call `authService.ListUsers(ctx, tenantID)` — already tenant-scoped via role-provider join.
2. Filter `mu.IsActive == true`.
3. Map `authdomain.ManagedUser{UserID, DisplayName}` → `application.UserOption{UserID, DisplayName}`.
4. Sort by `strings.ToLower(DisplayName)` ASC; tie-break by `UserID` ASC.
5. Return non-nil empty slice when no users qualify.

Composition root wires it in: a single line added to the `docDeps` literal at `apps/api/cmd/metaldocs-api/main.go:414` — `IAMUserOptions: wiring.NewDocumentsIAMUserOptions(authService)`. The nil-safe branch at `module.go:144-146` stays as defense-in-depth.

## Non-goals (mandatory)

- **No change to `IAMUserOptionsReader` shape.** Consumer port stays as defined at `iam_user_options.go:11`.
- **No new IAM port and no IAM API change.** F3.1 wires the existing dep; M4/F4.2+F4.3 will introduce IAM-owned ports.
- **No change to `auth.Service.ListUsers` or `auth.domain.Port`.** Producer of the underlying call stays as-is.
- **No modification to `placeholderOptionsIAMAdapter` in `documents/module.go`.** Its nil-guard stays.
- **No FE codegen, no OpenAPI change.** The placeholder-options endpoint contract is unchanged — only its returned data goes from "empty list" to "real list".
- **No DB migration, no schema change, no new SQL.**
- **No new IAM-owned `tenantID`-scoped query.** Reuses what `auth.Service.ListUsers` already provides.
- **No exhaustive role-based filtering.** "Active in tenant" is the bar; role-based UX filtering is out of scope (would be its own feature with its own consumer).

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|-----------------|
| 1 | Adapter filters `!IsActive` — deactivated users excluded | `TestDocumentsIAMUserOptions_FiltersInactive` in `internal/wiring/iam_user_options_test.go` — table-driven, in-memory fake `authListUsers` returning mixed active/inactive; assert only active in output | fixture (unit) |
| 2 | Adapter sorts by `strings.ToLower(DisplayName)` ASC, tie-break `UserID` ASC | `TestDocumentsIAMUserOptions_SortOrder` in same file — input out-of-order incl. case variants + same-display-name tie; assert deterministic output | fixture (unit) |
| 3 | Empty result returns non-nil empty slice | `TestDocumentsIAMUserOptions_EmptyReturnsNonNil` — assert `len == 0` AND `result != nil` | fixture (unit) |
| 4 | Underlying error propagates unchanged | `TestDocumentsIAMUserOptions_PropagatesError` — fake returns sentinel err; assert `errors.Is(out, sentinel)` and `result == nil` | fixture (unit) |
| 5 | Tenant isolation — users of tenant B not returned for tenant A | `TestDocumentsIAMUserOptions_TenantIsolation` — fake `ListUsers(ctx, "A")` returns only tenant-A users (the auth port already tenant-scopes); assert tenantID is forwarded verbatim and no cross-tenant leakage | fixture (unit) |
| 6 | End-to-end: placeholder-options endpoint returns the real user list against live IAM/auth on a seeded tenant | Manual runtime smoke captured in `evidence.md`: `.\scripts\start-api.ps1`, login dev, `curl <placeholder-options route>?type=user` — assert non-empty JSON `[{user_id, display_name}, ...]`, sorted, active-only. Body snapshot pasted into `evidence.md` and explicitly **labeled `real`** per mission §8. | **real** (live auth + DB) |
| 7 | Wiring proof — `IAMUserOptions` field is set in `docDeps` | `grep -RIn 'IAMUserOptions' apps/api/cmd/metaldocs-api/` returns the wire line + evidence comment shows resolved variable `authService`. | command |
| 8 | Whole-repo regression | `go test ./...` green; M0/M1/M2 sentinel re-checks: `go test ./internal/iam/... ./internal/modules/documents/...`; report §6 H-D grep set returns 0. | command |
| 9 | Authz-scope check (F3.2 R1 analog applied defensively) | N/A — F3.1 adds no parameter and removes none. Recorded as N/A in `evidence.md`. | recorded N/A |

**Test-framework gate (CLAUDE.md §4):**
- Acceptance 1–5 are **application/wiring unit tests** → table-driven Go tests with shared in-memory fakes. The fake auth port (`type fakeAuthListUsers struct{...}` or function-typed `type authListUsersFunc func(ctx, tenantID) ([]authdomain.ManagedUser, error)`) lives in the same `_test.go` file; if an adjacent test file already exports a canonical fake from a `testing` subpackage, reuse it. **No ad-hoc fixture strings** — UUIDs or `uuid.New().String()` for `UserID`; deterministic seeded values for `DisplayName`.
- Acceptance 6 is runtime smoke (real provider), captured as evidence — not a test in the framework sense.
- Acceptance 8 invokes the existing test suites under their existing frameworks (testdb for DB integration, handler-test for HTTP). F3.1 does not add a new handler test; the endpoint contract is unchanged.

> TDD: write failing tests 1–5 first against the not-yet-existing `wiring.NewDocumentsIAMUserOptions`; implement adapter to green; then wire in `main.go` and capture acceptance 6 + 7 + 8 evidence.

## ADR needed?

- [x] No durable decision — skip. F3.1 is a Composition Root wiring + adapter following the established `apps/api/internal/wiring/` pattern (no new architectural choice). The cross-module-vs-IAM-port decision IS durable but is deferred to M4/F4.2 + F4.3, where it gets its own ADR. Recorded here for the validator's C1 read.
- [ ] Durable decision made → record an ADR under `wiki/decisions/` and link it here: —
