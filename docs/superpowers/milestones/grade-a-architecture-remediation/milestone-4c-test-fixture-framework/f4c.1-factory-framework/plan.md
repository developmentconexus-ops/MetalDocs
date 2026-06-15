# Feature F4c.1 — Plan (the "how")

> Spec (contract) is `spec.md`, approved pre-code. This is the build plan only.

## Files

- **New:** `tests/integration/testdb/factory.go` (build tag `integration`, package `testdb`).
- **New:** `tests/integration/testdb/factory_test.go` (build tag `integration`, package `testdb` — in-package so it can reuse `Open`, `randomSuffix`, `Qualified`, `seedWithCaps`).
- **Untouched:** `db.go` (empty-diff gate), `fixtures.go` (kept; factories generalize, do not delete it in F4c.1), `pgtest`, all consumer tests.

## Design

Functional-options builders, all in-package so they reuse `seedWithCaps` (tx-local cap assert,
pool-safe) and `Qualified`. Each builder:
1. starts from a defaults struct (config),
2. applies `WithX` options,
3. mints any unset identity (`uuid.NewString()` via a fresh helper; per-call-unique taxonomy codes from `randomSuffix`),
4. seeds FK parents it owns (only what it must — `NewControlledDoc` does NOT auto-seed taxonomy unless none supplied; prefer explicit wiring via options to keep each builder single-purpose),
5. executes the guarded INSERT inside `seedWithCaps(... correct cap ...)`,
6. returns a struct carrying the IDs/columns the consumers assert on.

Minted taxonomy codes: `fam-<suffix>` / `pa-<suffix>` / `prof-<suffix>` (suffix = `randomSuffix`,
lowercase hex) — all satisfy `^[a-z][a-z0-9_-]{1,63}$`. CD `code` (no format CHECK) minted as
`cd-<suffix>`; tests that assert a specific code use `WithCode`.

`NewDocument` default `template_version_id` = minted free UUID (no FK); `WithTemplateVersionID`
overrides. Default status `draft`; `WithStatus` for `published`/`approved`/`scheduled`. Numeric
columns default 0; `WithRevisionVersion`/`WithScheduleGen`/`WithRevisionNumber` override.

`Scenario.PublishedDocument` / `ScheduledRevision(gen)` compose `NewTenant → NewUser(WithRole system_admin) → NewTaxonomy → NewControlledDoc → NewDocument(...)`, returning the final `Document`
(and exposing parent IDs through its fields). Options on the composite thread through to the inner
builders where needed (tenant, owner).

## Builder → cap → table map (from consumers)

| Builder | INSERT target | cap (tx-local) |
|---------|---------------|----------------|
| `NewTenant` | `metaldocs.tenants` | none |
| `NewUser` (role) | `metaldocs.iam_users` (no cap) + `iam_user_roles` | `user.manage` |
| `NewTaxonomy` | families/process_areas/profiles | `taxonomy.manage` |
| `NewControlledDoc` | `public.controlled_documents` | `controlled_documents.create` |
| `NewDocument` | `public.documents` | `document.create` |
| `NewApprovalRoute` | `public.approval_routes` | none observed |
| `NewApprovalInstance` | `public.approval_instances` | `document.submit` |

## TDD order

1. Write `factory_test.go` first — `TestFactory_*` subtests for each builder + both `Scenario`
   helpers + the two-calls-no-collision + code-format subtests. Compile fails (builders undefined) =
   the failing test.
2. Implement `factory.go` to green.
3. Run the gate commands (spec Validation Gate), capture real output → `evidence.md`.

## Verify

- `go test -tags integration -count=1 -run TestFactory ./tests/integration/testdb/...` green.
- `git diff --exit-code tests/integration/testdb/db.go` empty.
- `go vet` / build clean for the package.
