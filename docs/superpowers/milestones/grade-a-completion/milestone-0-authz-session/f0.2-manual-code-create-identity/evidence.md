# Feature F0.2 — Evidence

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Feature:** `f0.2-manual-code-create-identity`  ·  **Closed:** 2026-06-15
> **Contract:** `spec.md` (Path A — symmetric refactor of the manual-code branch).

## What was implemented

- `internal/modules/controlleddocuments/application/service.go` — manual-code branch
  (lines 173-282 post-edit) now wraps the persist step in `s.runner.Do`:
  `authz.SeedTxIdentity` → `authz.Require(controlled_documents.create, cmd.ProcessAreaCode)`
  → `s.docs.CreateTx`. The "DELIBERATELY-PRESERVED asymmetry" comment block is removed —
  the asymmetry is closed; this matches the auto branch's discipline byte-for-byte.
- Producer ↔ consumer contract: every caller of `ControlledDocumentService.Create` with a
  non-nil `ManualCode` now goes through tier-2 area-scoped authz (ADR 0022 Phase 7), so a
  non-system-admin actor with `controlled_documents.create` in the area **succeeds**, an
  actor without it is **denied** with `authz.ErrCapDenied`, and a system_admin still
  bypasses (existing `Require` short-circuit). Symmetric with the auto branch — same
  cap, same area, same TxRunner.
- All OFF-tx preflight reads stay OFF-tx (H-PRE-1 honored): `CodeExists`,
  `ensureTemplateArtifact`, `GetTemplateVersionState`, `Resolve`, governance-event
  marshalling, `NewVisibility`, `NewControlledDocument`. Inside the new tx: only the
  three required calls.
- `internal/modules/controlleddocuments/application/manual_code_create_integration_test.go`
  — new `//go:build integration`, three real-Postgres tests (the spec's Validation Gate).
- `internal/modules/controlleddocuments/application/service_test.go` — `TestCreate_ManualCode`
  switched from `nil` runner to `newTxRunner(newPermissiveMockDB(t))` (harness alignment;
  the manual path now owns a tx). One-line change, mirrors the pattern used by other
  service tests that traverse the tx path.

Commits: pending close-out commit on `main`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — RED first, then GREEN | `go test -tags integration -run 'TestCreate_ManualCode_(NonAdmin\|SystemAdmin)' ./internal/modules/controlleddocuments/application/ -v` | RED on pre-refactor code: all 3 tests FAIL with `registry: authz check Create: read actor_id GUC: sql: Scan error on column index 0, name "current_setting": converting NULL to string is unsupported` — exactly the B2 root cause (no `SeedTxIdentity` → `MustActorID` reads NULL GUC and fails). GREEN on post-refactor code: `--- PASS: TestCreate_ManualCode_NonAdminWithCap_Succeeds (1.31s) --- PASS: TestCreate_ManualCode_NonAdminWithoutCap_Denied (0.12s) --- PASS: TestCreate_ManualCode_SystemAdmin_Succeeds (0.14s)` | **real** (live Postgres via `testdb.Open`, curated bootstrap, real `PostgresControlledDocumentRepository`, real `platformdb.NewTxRunner`) |
| Static (build) | `go build ./...` | clean (no output) | — |
| Targeted unit | `go test ./internal/modules/controlleddocuments/...` | all `ok` (controlleddocuments/application/delivery/http/domain/infrastructure) | fixture (sqlmock for `TestCreate_ManualCode` after harness alignment) |
| Whole-repo unit | `go test ./...` (FAIL grep) | empty — no FAIL line | mixed |
| CD integration sweep | `go test -tags integration -count=1 ./internal/modules/controlleddocuments/...` | application + delivery/http + infrastructure all `ok`. Pre-existing failure: `domain/sequence_test.go:50 TestSequenceAllocatorNextAndIncrement_Concurrent` — reproduced on clean `main` HEAD (stash + rerun), unrelated to F0.2 (raw `INSERT INTO document_profiles` without `taxonomy.manage` assertion). | **real** |
| Runtime proof of the consumer-contract change | The three integration tests ARE the runtime proof: a real `*PostgresControlledDocumentRepository` is constructed; the manual-code branch goes through `runner.Do` → `SeedTxIdentity` → `authz.Require` against curated `role_capabilities` (`author` → `controlled_documents.create`) and real `user_process_areas`; non-admin-with-cap commits a real row, non-admin-without-cap is denied with `authz.ErrCapDenied`, system_admin bypasses. | see TDD row | **real** |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Non-admin with `controlled_documents.create` in area → success | yes | `TestCreate_ManualCode_NonAdminWithCap_Succeeds` PASS |
| Non-admin without the cap → `authz.ErrCapDenied` | yes | `TestCreate_ManualCode_NonAdminWithoutCap_Denied` PASS |
| System-admin → success (bypass unchanged) | yes | `TestCreate_ManualCode_SystemAdmin_Succeeds` PASS |
| No regression in the auto-code branch | yes | CD integration sweep — auto-path tests unchanged and `ok` |
| Whole-repo green | yes | Whole-repo unit row + build row |

## Review disposition

- Spec-compliance review (self): Path A applied verbatim — `s.runner.Do` wraps
  `SeedTxIdentity` + `Require(CapControlledDocumentCreate, cmd.ProcessAreaCode)` +
  `s.docs.CreateTx`; OFF-tx preflight untouched; comment block removed; no auto-branch
  change; no schema change; no new repo method. PASS.
- Code-quality review (self): the new manual block is byte-for-byte the auto-branch
  pattern (same cap, same area, same TxRunner port) — closes the asymmetry instead of
  symptom-patching it (ADR 0022 root-cause discipline). One unit-test harness alignment
  (nil runner → permissive sqlmock) is mechanical, not a symptom patch. PASS.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Pre-existing `TestSequenceAllocatorNextAndIncrement_Concurrent` failure (raw INSERT without `taxonomy.manage` assertion) | Reproduced on clean F0.1-committed `main` HEAD (`git stash` + rerun) — not introduced by F0.2. Unrelated to authz/session correctness. | Trigger: M0 milestone-validator may flag; if confirmed not a F0.2 regression, stays as a pre-existing integration-suite issue for a future fix feature (not in this mission's §5 inventory). Owner: grade-a-completion operator. |
| `controlleddocuments.Repository.Create` (non-tx variant) still exists and still calls `authz.Require` internally — now dead within the manual branch (the manual branch uses `CreateTx`). | Removing it is a wider repo-interface cleanup that would touch unrelated call sites. Out of F0.2 scope (non-goal: "**Not** changing `s.docs.Create` or `s.docs.CreateTx` repo signatures or bodies"). | Trigger: future repo-port cleanup feature or M1 contract-tightening. Owner: grade-a-completion operator. |
