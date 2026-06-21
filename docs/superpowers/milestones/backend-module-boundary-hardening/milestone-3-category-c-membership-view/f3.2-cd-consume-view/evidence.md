# Feature F3.2 — Evidence

> **Milestone:** 3  ·  **Feature:** `f3.2-cd-consume-view` (C1 + C2)  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` — CD's `List`/`CanRead` restricted-visibility membership leg reads iam's
> published `metaldocs.v_active_user_areas` (ADR-0039 D3a), not the `user_process_areas` base table.

## What was implemented

- **EDIT** `internal/modules/controlleddocuments/infrastructure/repository.go` — both restricted-visibility
  membership EXISTS legs repointed:
  - `List` (`:150`): `FROM user_process_areas upa … AND upa.effective_to IS NULL` → `FROM metaldocs.v_active_user_areas upa …` (the `effective_to IS NULL` clause dropped — the view encodes it).
  - `CanRead` (`:492`): identical repoint.
  - The two `upa.effective_to IS NULL` anchor comments rewritten to record the view-read (ADR-0039 D3a / ADR 0037 D1). CD-owned legs (`controlled_document_area_grants`, `controlled_document_user_grants`, base `controlled_documents` scan, pagination/filters) **unchanged**; Go signatures/scanning unchanged; set-based EXISTS preserved (no N+1).
- **NEW** `internal/modules/controlleddocuments/infrastructure/membership_view_parity_integration_test.go` —
  per-site parity gate (`TestCanRead_ViewParityWithRaw`, `TestList_ViewParityWithRaw`) comparing the repo
  methods against **verbatim inline copies of the deleted raw SQL**, across company / restricted-area /
  restricted-area-**revoked** / restricted-user / owner / no-access scopes + explicit drift discriminators.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule.go` — drained the
  `{controlleddocuments/infrastructure/repository.go, user_process_areas}` (C1+C2) ledger entry, replaced
  with a "ported (M3/F3.2)" note.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule_test.go` — realigned
  `TestHGCrossModule_Negative_PendingBaseline` from the now-drained CD site to the still-pending C4d
  search site (`search/infrastructure/v2documents/reader.go × user_process_areas`, scheduled M4).

## Verification

| Check | Command | Result (evidence) | Real vs fixture |
|-------|---------|-------------------|-----------------|
| TDD — parity green **before** raw deleted (D6 sanity: raw repo == raw baseline) | `go test -tags integration -run 'TestCanRead_ViewParityWithRaw\|TestList_ViewParityWithRaw' ./…/controlleddocuments/infrastructure/` (pre-repoint) | `ok …/infrastructure 3.396s` | real (PG :5434) |
| Repoint correct — parity green **after** raw deleted (view repo == raw baseline, zero authz drift) | same command (post-repoint) | `ok …/infrastructure 3.583s` | real (PG :5434) |
| Build | `go build ./...` | `BUILD-DONE` (exit 0) | — |
| Guard exit 0, C1+C2 drained | `go run ./tools/cilint ./...` | `cilint-exit=0` | real |
| Cilint unit suite green after fixture realign | `go test ./tools/cilint/...` | `ok …/tools/cilint/internal/analyzers 2.803s` | real |
| `user_process_areas` gone from CD **production** code | `git grep -n user_process_areas -- internal/modules/controlleddocuments/` | only matches are in `application/manual_code_create_integration_test.go` (a membership-seed **write** in a test; not a C1/C2 read) + the new parity test's verbatim baseline; `repository.go` has **zero** | real |
| CD package regression | `go test -tags integration ./internal/modules/controlleddocuments/...` | `application`, `delivery/http`, `infrastructure` **ok**; only `domain` `TestSequenceAllocatorNextAndIncrement_Concurrent` FAILs — pre-existing bounded defer (runs on raw base DSN, `metaldocs.document_profiles` absent), NOT F3.2 | real (PG :5434) |

> **Authz-drift discriminator (the load-bearing proof):** the seeded **revoked** membership row (past
> `effective_to`, `revoked_by` set) is excluded by both the raw `effective_to IS NULL` baseline AND the
> view — `repo.CanRead(restricted, revokedMem)=false`, `repo.List(revokedMem)` = {company CD only}; the
> active member sees both CDs. View-form == raw-form for every (actor, cd). No visibility change.

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|--------------------------------|------|----------|
| `CanRead` view-form == raw-form, all scopes incl. revoked | yes | `TestCanRead_ViewParityWithRaw` GREEN + discriminators |
| `List` visible-id set view-form == raw-form, all scopes incl. revoked | yes | `TestList_ViewParityWithRaw` GREEN + discriminators |
| Parity green before raw deleted (D6) | yes | pre-repoint GREEN row, then repoint, then post-repoint GREEN |
| Build + CD package | yes | `go build` exit 0; CD packages ok (defer noted) |
| Guard exit 0, C1+C2 entry drained | yes | `cilint-exit=0`; ledger note |
| `git grep` CD production → 0 raw reads | yes | only test-file seeds + parity baseline remain |
| Cilint unit suite green after fixture realign | yes | `ok …/analyzers 2.803s` |

## Review disposition

- Spec-compliance review: **PASS** — only the membership leg changed; CD-owned legs, signatures, scanning,
  pagination untouched; set-based EXISTS preserved; no temporal predicate re-derived in CD SQL.
- Code-quality review: **PASS** — parity test asserts set-equality against verbatim deleted SQL (not a
  count), includes the revoked-membership drift discriminator the search test lacks; ledger drain + cilint
  fixture realign + suite re-green done in-feature (the discipline M2 skipped). Repo `nil` reader resolves to
  `NoopActiveInstanceReader{}` — visibility path does not touch the active-instance port.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `TestSequenceAllocatorNextAndIncrement_Concurrent` FAIL on raw base DSN | Pre-existing, not introduced by F3.2; needs full bootstrap schema (`metaldocs.document_profiles`); orthogonal to the membership-leg seam | Trigger: sequence-allocator test migrated onto the testdb bootstrap harness; owner: mission backlog (not M3 scope) |
