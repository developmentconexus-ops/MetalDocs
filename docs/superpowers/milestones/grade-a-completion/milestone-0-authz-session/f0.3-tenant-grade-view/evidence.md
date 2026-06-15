# Feature F0.3 — Evidence

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Feature:** `f0.3-tenant-grade-view`  ·  **Closed:** 2026-06-15
> **Contract:** `spec.md` (root-cause family fix — both sibling sites aligned to declared `ScopeTenant`).

## What was implemented

- `internal/modules/documents/approval/application/read_service.go` — two call-site alignments to the canonical `documents/application/view_service.go:71` idiom:
  - **`LoadInstance`** (lines 44-88 post-edit): kept `loadInstanceAreaCode` for its `found` existence probe; dropped the `if areaCode == "" { areaCode = "tenant" }` coalesce; `authz.Require(... CapDocumentView, areaCode)` → `authz.Require(... CapDocumentView, "tenant")`; replaced the stale "COALESCE / Phase 11 F7" comment block with a one-line reference to `capability_scope.go:51` and the canonical sibling. `areaCode` return is now intentionally discarded with `_, found, err := …`.
  - **`LoadActiveInstanceByDocument`** (lines 91-128 post-edit): removed the `docapp.LoadDocumentAreaCode` call entirely (no other consumer at this site); dropped the coalesce; `authz.Require(... "tenant")` with the same one-line comment. The repo lookup below still returns `ErrNoActiveInstance` for a missing document/instance, so defense-in-depth is preserved.
  - Removed unused `docapp` import.
  - `loadInstanceAreaCode` helper definition (~360-384): unchanged.
- Producer ↔ consumer contract: every caller of `ReadService.LoadInstance` / `ReadService.LoadActiveInstanceByDocument` (instance-keyed inbox detail; document-keyed approval lookup) now obeys the declared tenant-grade scope of `CapDocumentView` (`iam/domain/capability_scope.go:51`). A viewer holding `document.view` anywhere in the tenant reads any document; an actor with no `document.view` is denied with `authz.ErrCapDenied{AreaCode:"tenant"}`; `system_admin` still bypasses.
- `internal/modules/documents/approval/application/read_service_tenant_grade_view_integration_test.go` — new `//go:build integration`, six real-Postgres tests covering the spec's Validation Gate (granted-cross-area × 2 methods, denied × 2 methods, system_admin × 2 methods).
- `internal/modules/documents/approval/application/read_service_test.go` — `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad`: dropped the now-stale `SELECT COALESCE(d.process_area_code_snapshot ...)` mock expectation (the call is gone); added an explicit `denied.AreaCode == "tenant"` assertion so the test now pins the tenant-grade envelope.

Commits: pending close-out commit on `main`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — RED first, then GREEN | `METALDOCS_DATABASE_URL=… go test -tags integration -run 'TestLoad(Instance\|ActiveInstanceByDocument)_TenantGradeViewer_DocWithAreaCode_Granted' ./internal/modules/documents/approval/application/ -v -count=1` | RED on pre-fix: `LoadInstance for tenant-grade viewer = authz: capability "document.view" denied for actor "ab1e8aa3-…" in area "pa-54b876fdfa"` and `LoadActiveInstanceByDocument … denied … in area "pa-0d1a6c8bd3"` — exactly the B3 root cause (tenant-grade view narrowed to area-grade). GREEN on post-fix: `--- PASS: TestLoadInstance_TenantGradeViewer_DocWithAreaCode_Granted (1.43s)`, `--- PASS: TestLoadActiveInstanceByDocument_TenantGradeViewer_DocWithAreaCode_Granted (0.16s)`. | **real** (live Postgres via `testdb.Open`, curated bootstrap, real `PostgresApprovalRepository`, real `platformdb.NewTxRunner`; viewer role grant seeded in a DIFFERENT area than the document's — proves the tenant-wide grant) |
| Other four new integration cases | same command, broader run | `--- PASS: TestLoadInstance_NoViewGrant_Denied (0.14s)`, `--- PASS: TestLoadActiveInstanceByDocument_NoViewGrant_Denied (0.13s)`, `--- PASS: TestLoadInstance_SystemAdmin_Granted (0.15s)`, `--- PASS: TestLoadActiveInstanceByDocument_SystemAdmin_Granted (0.15s)` — deny envelope is `AreaCode:"tenant"` on both deny tests (asserted in-test); system_admin bypass works on both methods. | **real** |
| Static (build) | `go build ./...` | clean (no output) | — |
| Targeted unit (module) | `go test ./internal/modules/documents/approval/application/...` | `ok  metaldocs/internal/modules/documents/approval/application  2.002s` — includes the retargeted `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad` (now asserts `denied.AreaCode == "tenant"`). | fixture (sqlmock) |
| Module integration sweep | `METALDOCS_DATABASE_URL=… go test -tags integration -count=1 ./internal/modules/documents/approval/...` | `ok` for application, domain, http, http/contracts, infrastructure, infrastructure/signature, jobs, repository. **Note:** the pre-existing `TestSequenceAllocatorNextAndIncrement_Concurrent` flake recorded in F0.2's defer did **not** trigger this run (`domain` passed). Not a F0.3 effect — F0.3 touches no domain/sequence code. | **real** |
| Whole-repo build + unit | `go build ./...` then `go test ./...` | build clean; `go test ./...` no `FAIL` line (tail: `ok  metaldocs/tests/unit/iam_people (cached)` … `ok  metaldocs/tools/cilint/internal/analyzers (cached)`). | mixed |
| Runtime proof of the consumer-contract change | The six new integration tests ARE the runtime proof: real `PostgresApprovalRepository`, real tx, real curated `role_capabilities` (`viewer → document.view`), real `user_process_areas` row seeded in a **different** area than the document's. Granted-path tests prove the tenant-wide grant lands across areas; denied-path tests prove the deny envelope is the tenant sentinel; system_admin tests prove the bypass short-circuit remains. | see TDD rows | **real** |

## Acceptance vs spec Validation Gate

| # | Acceptance criterion (from `spec.md`) | Met? | Evidence |
|---|---------------------------------------|------|----------|
| 1 | `LoadInstance` — tenant-grade viewer, doc with real area code → success | yes | `TestLoadInstance_TenantGradeViewer_DocWithAreaCode_Granted` PASS (RED pre-fix proven first) |
| 2 | `LoadActiveInstanceByDocument` — same → success | yes | `TestLoadActiveInstanceByDocument_TenantGradeViewer_DocWithAreaCode_Granted` PASS (RED pre-fix proven first) |
| 3 | `LoadInstance` no-grant → `ErrCapDenied{AreaCode:"tenant"}` | yes | `TestLoadInstance_NoViewGrant_Denied` PASS (asserts envelope) |
| 4 | `LoadActiveInstanceByDocument` no-grant → `ErrCapDenied{AreaCode:"tenant"}` | yes | `TestLoadActiveInstanceByDocument_NoViewGrant_Denied` PASS (asserts envelope) |
| 5 | system_admin still bypasses on both methods | yes | `TestLoad*_SystemAdmin_Granted` PASS |
| 6 | Existing `read_service_test.go` green (sqlmock case retargeted to tenant envelope) | yes | Targeted-unit row above; `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad` PASS with new `denied.AreaCode == "tenant"` assertion |
| 7 | No authz regression across approval module | yes | Module integration sweep row above (all 8 sub-packages `ok`) |
| 8 | Whole-repo green | yes | Whole-repo build + unit row above |

## Review disposition

- Spec-compliance review (self): scope is exactly the two sibling sites named in spec.md §"What this feature implements"; `loadInstanceAreaCode` helper unchanged (non-goal honored); no area-grade approval call site touched (`decision_service.go`, `publish_service.go`, etc. — non-goals); `authz.Require` body and `capability_scope.go` untouched (HS-2 respected); no schema change; no repo or HTTP change. PASS.
- Code-quality review (self): both sites now byte-for-byte mirror the canonical `view_service.go:70-71` idiom — same comment shape, same `"tenant"` literal, same control flow. The unused `docapp` import was removed, not silently kept. The `_, found, err :=` discard in `LoadInstance` is the honest signal that the resolver remains for its existence probe only — recorded in the comment and in the deferred-rename note in `spec.md`. The unit test was retargeted at the **contract** (tenant envelope on deny), not at incidental sql mocks, so future moves of the mock surface won't false-positive. PASS.
- Industry-standard root-cause discipline: fix targets the declared scope property at the divergent call sites, not the symptom at the cited line only; aligns with the canonical sibling; no shared-API redesign required.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `authz.Require` does not assert `ScopeOf(cap) == ScopeTenant ⇒ areaCode == "tenant"`. A static or dynamic assert at the shared layer would catch this class of bug program-wide. | Shared authz-API surface change → HS-2. Deserves its own ADR + a program-wide audit of every `Require` call site (and a discussion of whether the assert should be a hard `panic`/error or a structured warn). Out of M0 scope (M0 is the 4 named correctness defects, not authz-API hardening). | Trigger: M1 (contract-tail) or a dedicated authz-hardening micro-milestone. Owner: grade-a-completion operator. |
| Rename `loadInstanceAreaCode` → `instanceExists` (its only remaining use is the `found` probe). | Cosmetic refactor, no behavior change. CLAUDE.md §5.3 ("touch only what you must"). Today the helper still resolves a real `areaCode` — harmless because the caller discards it; a rename is a clean follow-up but not required for the contract. | Trigger: next planned edit of `read_service.go`. Owner: grade-a-completion operator. |
| F0.2-carried defer: pre-existing `TestSequenceAllocatorNextAndIncrement_Concurrent` flake. | Did not trigger in this F0.3 run, but the underlying gap (raw `INSERT INTO document_profiles` without a `taxonomy.manage` assertion) is unrelated to F0.3 and untouched. | Unchanged from F0.2 evidence: M0 milestone-validator may flag; if confirmed not a F0.3 regression, stays as a pre-existing integration-suite issue. Owner: grade-a-completion operator. |
