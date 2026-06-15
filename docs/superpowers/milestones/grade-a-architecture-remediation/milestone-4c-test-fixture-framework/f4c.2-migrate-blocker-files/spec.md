# Feature F4c.2 — Spec

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.2-migrate-blocker-files`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / operator ("Yes" — start F4c.2) — *contract derived from the consumer tests, not invented.*

> This is the feature's **contract**, approved **before any code**. The factory API is read **from its
> consumers** (the four blocker test files), which this feature migrates onto `factory.go` (F4c.1).
> Producer additions to `factory.go` are allowed (it is the framework); **`db.go` stays at HEAD**.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Scope — which files? | The four M4-blocker files named in `milestone.md` F4c.2: `approval/repository/postgres_approval_repository_test.go`, `approval/jobs/scheduled_publish_job_test.go`, `documents/repository/repository_commit_upload_integration_test.go`, `documents/repository/fillin_repository_integration_test.go`. Remaining files are F4c.3. |
| 2 | The abandoned WIP in the tree? | **Discarded** (`git checkout`), not built on: `postgres_approval_repository_test.go` (+63), `scheduled_publish_job_test.go` (+19), `repository_revision_history_integration_test.go` (±2), `authz_bypass_test.go` (±5). The first two are then rewritten onto the factory; the latter two are restored to HEAD (their migration is F4c.3, out of F4c.2 scope). |
| 3 | approval-repo tests pass `tx` to repo reads (seed-in-tx, rollback). How to migrate without losing that? | Seed via the factory on `db` (commits to the **private template-clone** — no cross-test leak, DB destroyed after test), then `db.BeginTx` for the **read**, pass that `tx` to the repo, `defer tx.Rollback()`. Reads are SELECTs → no tripwire, no caps, H-PRE-1 N/A. Equivalent isolation, zero shared state. |
| 4 | `TestScheduleGenerationIncrementsOnScheduledWritePath` does a **raw guarded UPDATE** (`document.edit`) as its system-under-test. How, without inline `set_config` (F4c.4 forbids)? | Export `testdb.SeedWithCaps(t, db, capsJSON, fn)` (thin wrapper over the existing in-package `seedWithCaps`; tx-local `is_local=true`, pool-safe). The test seeds the approved doc via the factory, then runs its UPDATE inside `SeedWithCaps([{"cap":"document.edit"}])`. No new inline `set_config` in the test. |
| 5 | fillin — seed gap or HS-2 schema-shadow? | **Determined empirically in TDD** (run on the clean testdb baseline first). If it fails with a Family-A 42703 (column/relation the curated baseline does not have but a production read expects) → **HS-2 stop + report**. If it is a seed/FK gap (e.g. missing `tenants`/parent) → fix with the factory. The classification is recorded in `evidence.md` either way. |

## Consumer contract (the exact factory surface these four files require)

Read verbatim from the four consumers. **No new builder is invented** — F4c.1's surface covers it, plus
**one** exported helper (Q4) for raw guarded writes:

| Consumer test (acceptance) | Factory surface it consumes |
|----------------------------|-----------------------------|
| `TestValidateScheduledSupersedeTarget_RealRows` | `NewTenant`, `NewUser(WithRole)`, `NewTaxonomy`, two `NewControlledDoc(WithTenant,WithTaxonomy,WithOwner)` sharing one taxonomy, `NewDocument(WithControlledDoc,WithStatus,WithRevisionNumber,WithOwner)` ×3 (published rev0 / approved rev1 on CD1, published rev0 on CD2) |
| `TestLoadCurrentPublishedHeadForDocument_RealRows` | one CD, three `NewDocument` (published rev0, approved rev1, published rev2) — head = highest published `revision_number` |
| `TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion` | `NewDocument(WithStatus("approved"),WithRevisionVersion(7))`, `NewApprovalRoute(WithTenant,WithProfile)`, `NewApprovalInstance(WithDocument,WithRoute,WithStatus("approved"))` |
| `TestLoadInstance_LoadsDocumentRevisionVersion` | same shape, `WithRevisionVersion(9)` |
| `TestScheduleGenerationIncrementsOnScheduledWritePath` | `NewDocument(WithStatus("approved"),WithRevisionNumber(5),WithRevisionVersion(2),WithScheduleGen(9))` + `testdb.SeedWithCaps` for the raw `document.edit` UPDATE |
| `TestScheduledPublishWorker_{PublishesWhenTruthMatches,NoOpWhenGenerationIsStale,NoOpWhenDeliveredBeforeEffectiveTime}` | `Scenario{}.ScheduledRevision(t, db, gen, WithEffectiveFrom(at), WithRevisionVersion(rv))`; worker args/query use the returned `doc.TenantID` / `doc.ID` |
| `TestCommitUpload_*` (2) | `testdb.InsertDraftDocument` (editor lineage, unchanged) + `NewTenant(WithTenant(devTenant))` (Family-C: the missing `metaldocs.tenants` parent) + `NewUser(WithUserID(docCreator),WithTenant,WithRole("system_admin"))` replacing `seedTenantRole` |
| `TestFillInRepository_UpsertValueAndListValues` | `testdb.InsertDraftDocument` (unchanged); **plus** whatever the empirical run shows is missing (seed gap → factory parent; schema-shadow → HS-2) |

- **Source of truth:** the four consumer files (read directly) + `factory.go` (F4c.1, in-package builders) +
  `fixtures.go` (`seedWithCaps`, `InsertDraftDocument`, `randomSuffix`).
- **Producer additions allowed in this feature:** `testdb.SeedWithCaps` (exported wrapper) only, plus any
  fillin parent builder the empirical run proves necessary. **No `db.go` edit.**

## What this feature implements

1. Discard the abandoned WIP (Q2). 2. Rewrite the two `pgtest.OpenAndMigrate` files
(`postgres_approval_repository_test.go`, `scheduled_publish_job_test.go`) onto `testdb.Open` + the
factory; delete `seedGovernedParents`, `seedScheduledDocument`, `scheduledDocumentSeed`. 3. In
`repository_commit_upload_integration_test.go` delete `seedTenantRole`, seed the missing `tenants`
parent + role via the factory (Family-C). 4. Investigate and resolve `fillin_repository_integration_test.go`
(seed-gap fix **or** HS-2 stop). 5. Add `testdb.SeedWithCaps`. Keep the non-DB unit tests in the approval
repo file (`TestMapPgError`, OCC tests, etc.) untouched.

## Non-goals (mandatory)

- **No `db.go` edit** (empty diff is a milestone acceptance gate).
- **No migration of F4c.3 files** (`create_document_snapshot_*`, `template_version_reader_*`,
  `repository_revision_history_*`, `authz_bypass_*`, the ~35 hardcoded-tenant / inline-`set_config` sites).
  The latter two only get their **WIP discarded** here, not migrated.
- **No CI grep-guard** (F4c.4); **no wiki/ADR** (F4c.5).
- **No production-source change** — except an HS-2-escalated, operator-approved Family-A schema fix if the
  fillin investigation proves one. No tripwire weakening/disable/CASE edit. No `pgtest` change in this feature.

## Validation Gate (concrete — approved before code)

Proof env (both): `$env:METALDOCS_DATABASE_URL` and `$env:DATABASE_URL` = operator DSN. PowerShell/Windows.

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| approval/repository — 5 named tests green from clean baseline | `go test -tags integration -count=1 -run 'TestValidateScheduledSupersedeTarget_RealRows\|TestLoadCurrentPublishedHeadForDocument_RealRows\|TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion\|TestLoadInstance_LoadsDocumentRevisionVersion\|TestScheduleGenerationIncrementsOnScheduledWritePath' ./internal/modules/documents/approval/repository/...` | real (template-cloned DB) |
| approval/jobs — 3 worker tests green | `go test -tags integration -count=1 -run TestScheduledPublishWorker ./internal/modules/documents/approval/jobs/...` | real |
| commit_upload — 2 tests green; Family-C `tenants` seeded | `go test -tags integration -count=1 -run TestCommitUpload ./internal/modules/documents/repository/...` | real |
| fillin — green (seed-gap fixed) **or** HS-2 boundary reported with evidence | `go test -tags integration -count=1 -run TestFillInRepository ./internal/modules/documents/repository/...` | real |
| The 3 named local seed helpers are gone | `grep -rn 'func seedGovernedParents\|func seedScheduledDocument\|func seedTenantRole' internal/` → no matches | real |
| Abandoned WIP discarded | `git diff` shows the 4 WIP files restored / rewritten, no `+63/+19` symptom-patch remnants | real |
| Harness untouched | `git diff --exit-code tests/integration/testdb/db.go` → exit 0 | real |
| Package compiles / vets | `go vet -tags integration ./internal/modules/documents/...` exit 0 | real |

TDD: discard WIP → run the four suites RED on testdb (capture, incl. fillin classification) → migrate →
GREEN. All proof **real** (template-cloned Postgres), not fixture/mock.

## ADR needed?

- [x] No durable architecture decision in F4c.2 (the framework decision is the F4c.5 ADR). `SeedWithCaps`
  is a mechanical export of an existing helper. If the fillin investigation forces an HS-2 production-schema
  fix, **that** carries its own ADR under `wiki/decisions/` (recorded at escalation time).
