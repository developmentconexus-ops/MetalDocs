# Feature F4c.1 — Spec

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.1-factory-framework`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / operator ("Approve — start F4c.1") — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Scope of F4c.1 — full migration or factory build only? | Operator: "Approve — start F4c.1" (single feature). F4c.1 builds `factory.go` + its self-test only; consumer migration is F4c.2/.3. Stop at F4c.1 evidence. |
| 2 | One package or a new one? | `milestone.md` is binding: factories live **in** `tests/integration/testdb` (build on it, generalize `fixtures.go`), **not** a new package. |
| 3 | Fixed taxonomy codes (`'po'`/`'quality'`) or minted-unique? | Minted-unique per call (the abandoned WIP failed by reusing fixed `'po'` → `document_profiles_pkey` 23505). `WithProfileCode`/`WithProcessAreaCode` overrides for the rare test that asserts on a specific code. Tests assert on IDs/revision/status, never on code values — minting is transparent. |
| 4 | Does `NewDocument` need a real `templates_template` version row? | No. The CD-governed `public.documents` rows in the approval/jobs consumers carry a **free** `template_version_id` (no FK enforced — verified: approval tests pass a const UUID with no template-version seed). `NewDocument` mints a free UUID by default; `WithTemplateVersionID` overrides. The editor-era `templates`/`template_versions` lineage stays in the existing `InsertDraftDocument` (commit_upload/fillin), out of F4c.1 scope. |

## Consumer contract (FIRST — before any producer)

The factory API is read **from its consumers** — the integration tests F4c.2/.3 will migrate onto it.
The contract below is extracted from those tests verbatim (table usage, asserted fields, caps), so the
producer is built to match, never the reverse.

- **Consumers (drive the API surface):**
  - `internal/modules/documents/approval/repository/postgres_approval_repository_test.go` — 5 real-DB tests (`TestValidateScheduledSupersedeTarget_RealRows`, `TestLoadCurrentPublishedHeadForDocument_RealRows`, `TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion`, `TestLoadInstance_LoadsDocumentRevisionVersion`, `TestScheduleGenerationIncrementsOnScheduledWritePath`). Today seed via local `seedGovernedParents` + inline INSERTs on the shared `pgtest` DB.
  - `internal/modules/documents/approval/jobs/scheduled_publish_job_test.go` — 3 worker tests. Today seed via local `seedScheduledDocument` (commits to shared DB → the collision source).
  - `internal/modules/documents/repository/repository_commit_upload_integration_test.go` — 2 tests; today seed via `testdb.InsertDraftDocument` (editor lineage) + local `seedTenantRole` (iam user + system_admin role).
  - **F4c.1's own self-test** is the immediate in-repo consumer that proves every builder.

- **Contract (the exact shape the consumers require):** functional builders in package `testdb` (build tag `integration`), sane defaults + `WithX` options, each returning a struct carrying the IDs/columns the tests assert on. Each builder seeds its FK parents, asserts the **correct** tripwire cap via the existing `seedWithCaps` (tx-local `is_local=true`, pool-safe), and is collision-free across calls (minted UUIDs + per-call-unique taxonomy codes). Required surface:

  | Builder | Seeds (tables) | Cap asserted | Returns (fields consumers read) | Key options |
  |---------|----------------|--------------|---------------------------------|-------------|
  | `NewTenant(t, db, ...opt)` | `metaldocs.tenants` | none (no tripwire) | `Tenant{ID}` | `WithTenantID` |
  | `NewUser(t, db, ...opt)` | `metaldocs.iam_users` (+ `iam_user_roles` when role set) | `user.manage` (role write only) | `User{ID, DisplayName, TenantID}` | `WithTenant`, `WithUserID`, `WithDisplayName`, `WithRole("system_admin")` |
  | `NewTaxonomy(t, db, ...opt)` | `document_families`, `document_process_areas`, `document_profiles` | `taxonomy.manage` | `Taxonomy{TenantID, FamilyCode, ProcessAreaCode, ProfileCode}` | `WithTenant`, `WithFamilyCode`, `WithProcessAreaCode`, `WithProfileCode` |
  | `NewControlledDoc(t, db, ...opt)` | `public.controlled_documents` | `controlled_documents.create` | `ControlledDoc{ID, TenantID, Code, ProfileCode, ProcessAreaCode, OwnerUserID, Status}` | `WithTenant`, `WithTaxonomy`, `WithOwner`, `WithCode`, `WithStatus`, `WithTitle` |
  | `NewDocument(t, db, ...opt)` | `public.documents` | `document.create` | `Document{ID, TenantID, ControlledDocumentID, TemplateVersionID, Status, RevisionNumber, RevisionVersion, ScheduleGeneration}` | `WithTenant`, `WithControlledDoc`, `WithOwner`, `WithTemplateVersionID`, `WithName`, `WithStatus`, `WithRevisionNumber`, `WithRevisionVersion`, `WithScheduleGen`, `WithEffectiveFrom` |
  | `NewApprovalRoute(t, db, ...opt)` | `public.approval_routes` | none observed | `ApprovalRoute{ID, TenantID, ProfileCode}` | `WithTenant`, `WithProfile`, `WithName` |
  | `NewApprovalInstance(t, db, ...opt)` | `public.approval_instances` | `document.submit` | `ApprovalInstance{ID, TenantID, DocumentID, RouteID, Status}` | `WithDocument`, `WithRoute`, `WithStatus`, `WithIdempotencyKey`, `WithContentHash` |

  Composite (cuts the repeated multi-row shapes the consumers build by hand):
  - `Scenario` helpers returning the wired-together structs:
    - `PublishedDocument(t, db, ...opt) Document` — tenant + user + taxonomy + CD + one `published` document (the `LoadCurrentPublishedHead` / supersede shapes).
    - `ScheduledRevision(t, db, gen int64, ...opt) Document` — tenant + user + taxonomy + CD + one `scheduled` document at `schedule_generation = gen`, `revision_version` set (the `scheduled_publish_job` shape).

- **Source of truth for the contract:** the consumer test files above (read directly), plus the
  existing seed layer they generalize: `tests/integration/testdb/fixtures.go`
  (`seedWithCaps`, `SeedGovernedTaxonomy`, `SeedSystemAdmin`) and `db.go` (`Open`, `Qualified`,
  `DeterministicID`). Column lists, statuses, and caps are copied from the consumers — not invented.

## What this feature implements

Add `tests/integration/testdb/factory.go` (build tag `integration`, package `testdb`): the
functional-builder factories + `Scenario` composites above, generalizing `fixtures.go`. Each builder
mints fresh UUIDs and per-call-unique taxonomy codes (matching the `profile_code_format` CHECK
`^[a-z][a-z0-9_-]{1,63}$`), seeds FK parents, and asserts the real tripwire cap via `seedWithCaps`.
Ship a TDD self-test (`factory_test.go`) exercising every builder + both `Scenario` helpers against a
fresh template-cloned DB from `testdb.Open`, proving rows land, FKs/tripwire are satisfied, and two
calls in one suite do not collide.

## Non-goals (mandatory)

- **No consumer migration.** The approval/jobs/commit_upload/fillin/snapshot/template_version_reader
  test files are NOT touched in F4c.1 (that is F4c.2/.3). Their local seed helpers stay until then.
- **No edit to `tests/integration/testdb/db.go`** — empty diff is a milestone acceptance gate.
- **No `templates_template` / editor-`templates` builders.** `NewDocument` uses a free
  `template_version_id`; the editor lineage stays in `InsertDraftDocument`. Snapshot/template-version
  builders (if needed) are added when F4c.3 migrates their consumers.
- **No CI grep-guard** (F4c.4), **no wiki/ADR** (F4c.5).
- **No production-source change.** No tripwire weakening/disable/CASE edit. No `pgtest` change.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Each builder seeds its contracted rows + returns the asserted IDs/columns | `go test -tags integration -count=1 -run TestFactory ./tests/integration/testdb/...` (self-test asserts returned struct fields are non-empty and the rows exist) | real (clean template-cloned DB) |
| Every builder satisfies its FK parents and the tripwire (real cap asserted via `seedWithCaps`) | same self-test green (a wrong/absent cap → P0001; a missing parent → 23503) | real |
| Two factory calls in one suite do **not** collide (per-call-unique codes + clone isolation) | self-test subtest seeding two tenants/CDs in one DB without `document_profiles_pkey` 23505 | real |
| Generated taxonomy codes match the `profile_code_format` CHECK | self-test asserts each minted code matches `^[a-z][a-z0-9_-]{1,63}$` (and the INSERT does not raise 23514) | real |
| `Scenario.PublishedDocument` / `ScheduledRevision(gen)` produce the consumer shapes | self-test subtests assert status (`published`/`scheduled`), `schedule_generation`, `revision_version` | real |
| Harness untouched | `git diff --exit-code tests/integration/testdb/db.go` (empty) | real |

Proof env (both required): `$env:METALDOCS_DATABASE_URL` and `$env:DATABASE_URL` = operator DSN.
TDD: write `factory_test.go` failing first (builders undefined → compile fail), then implement
`factory.go` to green. All proof is **real** (template-cloned Postgres), not fixture/mock.

## ADR needed?

- [x] No durable architecture decision in F4c.1 — the framework/harness **decision** (IntegreSQL
  template-DB-per-test, factory pattern, discipline rules) is recorded by the **F4c.5** ADR under
  `wiki/decisions/`. F4c.1 is the producer build against an already-agreed contract.
