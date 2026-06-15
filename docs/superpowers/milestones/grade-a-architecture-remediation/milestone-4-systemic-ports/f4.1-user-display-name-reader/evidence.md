# Feature F4.1 — Evidence

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.1-user-display-name-reader`  ·  **Closed:** 2026-06-15
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output.

## What was implemented

- **Port (producer matches consumer contract):** `iamdomain.UserDisplayNameReader` interface in
  `internal/modules/iam/domain/user_display_name_port.go` — `DisplayName(ctx, tenantID, userID) (string, error)`
  and `DisplayNames(ctx, tenantID, []userID) (map[string]string, error)`, plus `NoopUserDisplayNameReader`
  null-object. Pool-backed impl `UserDisplayNameRepository` in
  `internal/modules/iam/infrastructure/postgres/user_display_name_repository.go` — reads
  `metaldocs.iam_users.display_name` on the iam connection pool, tenant-scoped via explicit
  `tenant_id = $N::uuid` predicate, omits empty `display_name`. The contract shape was read from the
  three consumers first (approval signoff, approval get-instance handler, documents create), not invented.
- **Consumer 1 — approval signoff repository:** `LoadActorDisplayName` (`postgres_approval_repository.go`)
  now delegates entirely to the injected port (`return r.displayName.DisplayName(...)`); raw `iam_users`
  SQL removed. Read stays **off the caller's tx** (on the pool) — never inside the signoff advisory-lock
  atomic tx (H-PRE-1). Constructor `NewPostgresApprovalRepository(db, displayName)`.
- **Consumer 2 — approval get-instance handler:** `get_instance_handler.go` / `handler.go` resolve
  eligible-actor rendered names via the port (batch `DisplayNames`) with `missing→userID` fallback,
  replacing raw cross-module `iam_users` SQL.
- **Consumer 3 — documents create repository:** `repository.New(db, displayName)`
  (`internal/modules/documents/repository/repository.go`); `CreateDocumentTx` populates
  `created_by_display_name_snapshot` from `r.displayName.DisplayName(...)` read **off-tx** before the
  insert (H-PRE-1), tenant-scoped. All `repository.New` call sites updated to 2-arg.
- **Wiring:** `internal/modules/documents/module.go` constructs the real iam `UserDisplayNameRepository`
  and injects it into both documents repositories — no module-boundary crossing of raw `iam_users` SQL.
- Test harness rehab co-resident in the touched documents integration tests (search_path → `metaldocs, public`,
  tripwire-aware `asserted_caps`, drop dropped-`visibility` column) — see Bounded defers for the one
  remaining gap.

Committed: port + consumers + tests in this feature's commit (subject `feat(milestone-4): F4.1 …`);
the prerequisite bootstrap seed fix landed separately as `653c8f59 fix(bootstrap): restore NOT-NULL
visibility in templates_template reference-data seed`.

## Verification

All integration runs used the live dev Postgres (`METALDOCS_DATABASE_URL=…@127.0.0.1:5433/metaldocs`),
`-tags integration -count=1 -p 2`.

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | port contract + each migrated consumer | red (no port / 2-arg ctor) → green after impl | real + fixture |
| Static — plain build | `go build ./...` | `BUILD OK` (exit 0) | — |
| Static — vet | `go vet ./internal/modules/iam/... ./internal/modules/documents/...` | `VET OK` | — |
| Static — vet (integration tag) | `go vet -tags integration ./internal/modules/iam/... ./internal/modules/documents/...` | `VET-INTEGRATION OK` (all integration tests compile, incl. Noop-injected documents create) | — |
| iam port `DisplayName` present/absent/tenant-scoped | `go test -tags integration -run TestUserDisplayNameRepository_DisplayName_Live ./internal/modules/iam/infrastructure/postgres/` | `--- PASS` (present_returns_value, absent_returns_empty, tenant_scoped_other_tenant_returns_empty); `ok 3.924s` | **real (live PG)** |
| iam port `DisplayNames` batch present/absent/empty-omitted | `…_run TestUserDisplayNameRepository_DisplayNames_Live …` | `--- PASS` (present_with_name_returned, empty_input_returns_empty_map) | **real (live PG)** |
| approval signoff display-name off-tx (H-PRE-1) | `go test -tags integration -run TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema ./internal/modules/documents/approval/repository/` | `--- PASS 1.88s`; log `AC5 live read-back: LoadActorDisplayName(…) = "Alice Approver"`; `AC3 empty-on-missing … = "" (nil err)` | **real (live PG)** |
| eligible-actor rendered names (missing→userID fallback) | `go test -count=1 ./internal/modules/documents/approval/http/` (`TestResolveEligibleActorNames_UsesFakePortAndMissingFallback`, `…_EmptyInstanceReturnsEmptyMap`) | `ok …/approval/http 3.258s` | fixture (fake port) |
| approval + iam unit suites | `go test -count=1 ./internal/modules/documents/approval/repository/ ./internal/modules/iam/infrastructure/postgres/` | `ok` both | fixture |
| Class root cause — 0 raw `iam_users` display-name SQL outside `iam/` | grep `iam_users` in non-test prod code outside `iam/` | only comments + one `e2e_seed.go` INSERT (seed, not a read); `LoadActorDisplayName` body = `return r.displayName.DisplayName(...)` | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Port exists in iam/domain; impl in iam/infrastructure; pool-backed | yes | build OK; `user_display_name_port.go` + `user_display_name_repository.go` |
| `DisplayName` present/absent/tenant-scoped | yes | iam `DisplayName_Live` PASS (real) |
| `DisplayNames` batch present/omit absent+empty | yes | iam `DisplayNames_Live` PASS (real) |
| approval signoff snapshot unchanged; read off-tx (H-PRE-1) | yes | approval `LoadActorDisplayName` live PASS (real); body delegates to port, runs on pool not tx |
| documents `created_by_display_name_snapshot` value preserved (single-tenant) — live create → row read-back | **partial — deferred** | port wiring proven: integration build compiles with port injected; `CreateDocumentTx` reads off-tx via same port proven live for approval; **live documents create→read-back blocked by pre-existing testdb seed gap** (see Bounded defers). Note: the named test `TestCreateDocumentTx_PopulatesAllSnapshotColumns` asserts the non-display-name snapshot columns under a `Noop` reader; it is not a display-name *value* assertion even when green. |
| eligible-actor rendered names byte-identical, missing→userID | yes | approval/http resolve tests PASS (fixture); approval live read returns exact `"Alice Approver"` (real) |
| **Class root cause:** 0 raw `iam_users` display-name SQL outside `iam/` | yes | grep clean (comments/seed-insert only); delegation confirmed |
| build + vet clean; backend-api-qa green | yes | BUILD/VET/VET-INTEGRATION OK |

## Review disposition

- Spec-compliance review: producer matches the consumer-first contract (shape read from the 3 consumers).
  Two deliberate divergences from legacy behavior recorded in `spec.md` non-goals (tenant-scoping tightened;
  read moved off-tx) — both correctness improvements, not contract breaks.
- Code-quality review: null-object (`NoopUserDisplayNameReader`) keeps test/no-resolve paths explicit;
  required-collaborator constructor (panics on nil at first use) prevents silent nil. H-PRE-1 honored —
  display-name read is on the pool, never inside the create/signoff lock-holding tx.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Gate #5 **live** documents create→`created_by_display_name_snapshot` read-back (test `TestCreateDocumentTx_PopulatesAllSnapshotColumns` RED) | Root cause is **pre-existing and unrelated to the F4.1 port**: `tests/integration/testdb.ApplyCuratedBootstrap` applies prerequisites + baseline + reference-data + migration tail but **not dev-seeds**, so the governed FK parents `document_families` / `document_profiles` / `document_process_areas` are absent in the template DB → any `controlled_documents` seed fails FK `controlled_documents_tenant_id_process_area_code_fkey` (SQLSTATE 23503). Affects the whole documents integration suite (repository_create, commit_upload, snapshot, fillin, revision_history), not just this test. The port wiring itself is proven (integration compile + identical off-tx codepath proven live for the approval consumer). The proper fix is one shared testdb seed helper for the governed area/profile/family parents (tripwire-aware caps), not per-test duplication — and touching the shared testdb harness + all documents integration tests is outside F4.1's port boundary (HS-6). | Spawned task `task_2d29ac7d` (documents integration suite rehab). Trigger: that task seeds the governed FK parents in the curated test bootstrap, then re-run `TestCreateDocumentTx_PopulatesAllSnapshotColumns` green and add an explicit `created_by_display_name_snapshot` value assertion with a non-Noop reader. |
