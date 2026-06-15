# Feature F4.1a — Evidence

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.1a-documents-testdb-rehab`  ·  **Closed:** 2026-06-15
> **Contract:** `spec.md` (consumer contract = the documents integration tests + the Validation Gate).
> Test-harness-only fix-feature. No production code / schema / migration change.

## What was implemented

All changes are in test/harness code:

- **`tests/integration/testdb/fixtures.go`** (+176):
  - `seedWithCaps(t, db, capsJSON, fn)` — runs `fn` in its own tx with `metaldocs.asserted_caps`
    asserted **transaction-locally** (`set_config(..., true)`, mirroring production
    `authz.appendAssertedCap`). Pool-safe: the tripwire reads the caps on the same connection that
    performs the write, and the assertion is discarded on commit (never leaks into the caller's session).
  - `SeedGovernedTaxonomy(...)` rewritten to seed `document_families` / `document_process_areas` /
    `document_profiles` through `seedWithCaps([taxonomy.manage])` instead of the previous **session-level**
    set/restore (which was not pool-safe with a multi-connection pool). Same rows, same idempotency.
  - `SeedSystemAdmin(t, db, tenantID, userID, displayName)` — new. Upserts `metaldocs.tenants` (FK
    parent of `iam_users`; no tripwire), upserts `iam_users` (carrying `display_name` for the F4.1
    port), and grants tenant-scoped `system_admin` in `iam_user_roles` under a `user.manage` tx-local
    assertion. Satisfies the create-path `authz.Require(document.create)` via the ADR 0022 tier-2
    system_admin inheritance bypass — seeding legitimately-required membership, **not** symptom-patching
    authz.
  - `SupersedeActiveDocumentForCD(t, db, controlledDocumentID)` — new. Walks the single active document
    of a CD through the legal status graph `draft → under_review → approved → published → superseded`
    (stubbing the six snapshot columns `enforce_snapshot_on_submit` requires; 32-byte hashes satisfy the
    `*_hash_len` checks), under a `document.edit` tx-local assertion. `trg_documents_legal_transition`
    permits no shortcut out of `draft`, so this is the only legal way to vacate the per-CD active set
    (`ux_documents_cd_active`) — mirrors production (a prior revision is published then superseded
    before its successor is created).

- **`internal/modules/documents/application/create_document_snapshot_integration_test.go`** (Gate #5):
  - `search_path` fixed via `ALTER DATABASE "<dbName>" SET search_path TO public, metaldocs` (bare
    `documents` now resolves to the **real** `public.documents`, not the dead `metaldocs.documents`);
    pre-ALTER pooled connection evicted (`SetMaxIdleConns(0)`), pool raised to 4 connections so the
    off-tx port read + templates `GetPublishedVersion` read do not deadlock against the open create tx
    (H-PRE-1). `SetMaxOpenConns(1)` removed (was the 600 s hang).
  - Manual `iam_users` insert + session-level cap block replaced by `SeedSystemAdmin(..., "Snapshot
    Author")`. `seedCreateDocumentSnapshotRows` rewritten to assert its caps tx-locally (pool-safe).
  - The display-name **value** assertion (`createdByDisplayNameSnap == "Snapshot Author"`) now runs
    under the **real** `iampg.NewUserDisplayNameRepository(db)` (was a `Noop`-reader, non-value test).

- **`internal/modules/documents/repository/repository_create_integration_test.go`** (4 tests):
  - `search_path` → `public, metaldocs` (bare `documents` → real `public.documents`).
  - `SeedSystemAdmin(...)` added after `SeedGovernedTaxonomy(...)` in each test (fixes the
    `document.create` authz denial).
  - The two intermediate "publish first doc" workarounds (which were silent no-ops against the dead
    `metaldocs.documents`) replaced by `SupersedeActiveDocumentForCD(...)`.

## Verification

All integration runs: live dev Postgres (`METALDOCS_DATABASE_URL=…@127.0.0.1:5433/metaldocs`),
`-tags integration -count=1 -p 2`.

| # | Check | Command | Result |
|---|-------|---------|--------|
| 1 | **Gate #5 real-reader value proof** | `go test -tags integration -count=1 -p 2 -run TestCreateDocumentTx_PopulatesAllSnapshotColumns ./internal/modules/documents/application/` | **`ok … 161.952s`** (was 600 s deadlock). Asserts `created_by_display_name_snapshot == "Snapshot Author"` under real `iampg.NewUserDisplayNameRepository`. |
| 2 | `repository_create` suite | `go test -tags integration -count=1 -p 2 -run 'TestCreateDocumentTx_StorageKeyInvariant\|…RevisionNumber…\|…RejectsEmptyName\|TestGetDocument_ReturnsSnapshotMetadata' -v ./internal/modules/documents/repository/` | **PASS** — `StorageKeyInvariant` (Non/Empty both PASS), `RevisionNumberIncrementsForSameCD` PASS, `RejectsEmptyName` PASS, `GetDocument_ReturnsSnapshotMetadata` PASS; `ok … 269.889s` |
| 3 | No production change | `git diff --stat` (touched paths) | only `fixtures.go` (+176) + the two integration test files (+77, +44); `269 insertions(+), 28 deletions(-)`; **zero** non-test files |
| 4 | Static clean | `go vet -tags integration ./internal/modules/documents/repository/ ./internal/modules/documents/application/ ./tests/integration/testdb/` | **VET OK** (exit 0) |

## Acceptance vs spec Validation Gate

| Acceptance (spec.md) | Met? | Evidence |
|----------------------|------|----------|
| Gate #5 green, real reader, value assertion | yes | check #1 PASS; `== "Snapshot Author"` |
| `repository_create` suite green on real `public.documents` | yes | check #2 PASS (all 4 tests / 5 cases) |
| No production change | yes | check #3 — test/harness files only |
| Static clean | yes | check #4 VET OK |

## Review disposition

- **Root-cause, not symptom:** the authz denial is fixed by seeding the *required* tenant-scoped
  `system_admin` membership (tier-2 bypass per ADR 0022), not by relaxing/asserting around the authz
  layer. The `search_path` fix points the tests at the real table production already uses. The deadlock
  fix raises the pool so the off-tx read (H-PRE-1) can proceed — it does **not** move the read into the
  lock-holding tx.
- **Pool-safety:** all governed seeds now assert caps tx-locally (mirroring `authz.appendAssertedCap`),
  so the multi-connection Gate #5 pool is safe and no cap leaks into the caller's session.
- **H-PRE-1 intact:** the F4.1 display-name port read stays off the create tx (on the pool); raising the
  pool size is what *allows* that off-tx read to run, confirming the production placement is correct.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **Rest of the documents integration suite still RED** — `TestFillInRepository_UpsertValueAndListValues`, `TestCommitUpload_PersistsRevisionAndFormDataSnapshot` + `…IdempotentReplay…`, `TestSnapshotRepository_*` (4), `TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows`. | **Same pre-existing harness-rot class** as the F4.1 defer documented ("affects the whole documents integration suite") — `column "tenant_id"/"active_session_id" does not exist` (bare names resolving to dead legacy tables, same `search_path`/dual-table root cause) and a stale `registry.create` cap in `revision_history` (superseded by `controlled_documents.create` per migration 0210/0231). **Not regressions** from this fix-feature: `fixtures.go` changes are additive + signature-compatible (`SeedGovernedTaxonomy` leaves identical session-cap state), and none of these test files were touched. Rehabbing them is materially larger than the operator-approved Gate-#5 slice (per-target-table investigation, not mechanical) → **HS-6: surfaced, not silently absorbed.** | Spawned `task_2d29ac7d` remains open as the owner. *Trigger:* a dedicated documents-integration-suite rehab pass (apply the same `search_path` + tripwire-aware seed-helper pattern to fillin/commit_upload/snapshot_repository, and refresh `revision_history`'s stale cap), gated on operator approval to expand scope beyond F4.1's Gate #5. |
