# Feature F2d.1 — Evidence

> **Milestone:** 2d — Workflow Coherence FE + Viewer Contract  ·  **Feature:** `f1-viewer-contract`  ·  **Closed:** 2026-07-09
> **Contract:** [`spec.md`](spec.md) (consumer contract + Validation Gate this proves against).

## What was implemented

By outcome (producer matches the `spec.md` consumer contract — the generated `viewer` DTO shape
`deriveWorkspaceMode` (F2d.3) consumes):

- **Server-derived `viewer` block** on `ApprovalInstanceByDocumentResponse`: `is_author`,
  `eligible_for_active_stage`, `has_signed_active_stage` (all required) + nullable
  `via_delegation_from{user_id, display_name}`. Always present whenever the instance is returned;
  terminal instance ⇒ all-false/null (interview rows 1–2).
- **Pure domain fn** `domain.ViewerEligibility` (`domain/viewer.go`) — composes the write-path
  primitives `ResolveEligibleIdentity` (snapshot pool ∪ active delegation) then `CheckSoD`
  (author-exclusion + cross-stage double-sign). **No second membership/SoD rule** — the anti-split-brain
  requirement.
- **App method** `ReadService.LoadInstanceByDocumentForViewWithViewer` (`application/read_service.go`) —
  loads in the view tx, reads viewer id from the tx GUC (`authz.MustActorID`), `LoadActiveDelegationsFor`
  (plain SELECT, H-PRE-1), calls the pure fn.
- **Handler + mapper** — delegator display name resolved **off-tx** via the existing `displayNameReader`
  port; by-id handler passes `nil` ⇒ `viewer` omitted.
- **OpenAPI + regen** — `viewer` schema in `api/openapi/v1/openapi.yaml`; `oapi-codegen` regenerated Go +
  TS; contracts structs added. No hand-written `body.data.viewer` reader (ADR 0035).
- **Emergent — visibility-gate convergence (see plan.md execution notes + ADR 0078 amendment):**
  `requireInstanceVisible` converged onto the single eligibility primitive (author fast-path →
  `CheckEligibility` → delegation fallback `ResolveEligibleIdentity`), deleting the hand-rolled
  membership loop. Closes the can-act-but-cannot-see delegate defect (delegate could sign but got 404
  loading). Gated by `/developing-new-work` → system-impact analysis 🟡 Yellow
  (`docs/superpowers/analysis/2026-07-08-delegation-aware-visibility-system-impact.md`, committed
  `890acd60`).

Code diff not yet committed (staged for review; milestone HS-1 gate is pre-push). Analysis + ADR
amendment committed on `main`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `TestViewerBlock_Delegate` (RED: `instance not visible to this actor` → GREEN after convergence) | red→green documented in plan.md execution notes | real |
| Static — build | `go build ./...` | `BUILD_EXIT=0` | — |
| Static — vet | `go vet ./internal/modules/documents/approval/...` | `VET_EXIT=0` | — |
| Unit suites | `go test ./internal/modules/documents/approval/...` | all packages `ok` (application 7.352s, domain/http/contracts/infrastructure ok) | real (in-proc) |
| Viewer scenarios (real DB) | `go test -tags integration ...application/... -run 'TestViewerBlock' -timeout 45m` | **7/7 PASS**, `GO_TEST_EXIT=0`, 567s | real |
| F8 visibility regression (real DB) | `go test -tags integration ...application/... -run 'TestLoadInstance_\|TestLoadActiveInstanceByDocument_' -timeout 45m` | **11/11 PASS** (10 named deny+grant + 1 bonus), `GO_TEST_EXIT=0`, 504s | real |

> Real-DB runs use the `testdb` factory, isolated DB per test, `-tags integration`. Drop-teardown
> stalls 60–480s/test on this Windows box (`db.go:103: drop isolated test database … context deadline
> exceeded` — teardown noise, all tests PASS); split into two runs with `-timeout 45m` to stay under
> Go's per-binary timeout. No fixture-only proof: every viewer/visibility claim is real Postgres.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| viewer **author** (`is_author=true`, `eligible=false` by author-rule) | yes | `TestViewerBlock_Author` PASS |
| viewer **snapshot actor** (`eligible=true`, `via_delegation_from=null`) | yes | `TestViewerBlock_SnapshotActor` PASS |
| viewer **delegate** (`eligible=true`, `via_delegation_from={id,name}`) | yes | `TestViewerBlock_Delegate` PASS |
| viewer **already-signed** (`eligible=true`, `has_signed_active_stage=true`) | yes | `TestViewerBlock_AlreadySigned` PASS |
| viewer **observer** (all false/null) | yes | `TestViewerBlock_Observer` PASS |
| viewer **author-who-is-also-approver** (excluded by author-rule) | yes | `TestViewerBlock_AuthorAlsoApprover` PASS |
| viewer present on **terminal instance** (all-false) | yes | `TestViewerBlock_NoActiveStage` PASS |
| no in-tx display-name lookup (H-PRE-1) | yes | mapper resolves delegator name off-tx via `displayNameReader`; code review confirmed |
| regen produces no hand-written DTO consumer; build clean | yes | `oapi-codegen` regen + `go build ./...` = 0; no `body.data.viewer` reader |
| **added gate (convergence):** F8 deny-regression stays green | yes | 11/11 `TestLoadInstance_*`/`TestLoadActiveInstanceByDocument_*` PASS — deny cases still deny |

All Validation-Gate criteria **met**.

## Review disposition

Independent diff review (spec-compliance + code-quality), 2026-07-09 — **VERDICT: APPROVE**. No
Critical, no Major.

- **Anti-split-brain (the whole point) — PASS.** `domain.ViewerEligibility` composes exclusively on
  `ResolveEligibleIdentity` + `CheckSoD`; converged `requireInstanceVisible` composes `CheckEligibility`
  (direct) then `ResolveEligibleIdentity` (delegation fallback). No re-implemented membership loop, no
  second SoD rule. Read-visibility is now a projection of act-eligibility.
- **Spec conformance — PASS.** `viewer` block shape matches spec (3 required bools + nullable
  `via_delegation_from{user_id, display_name}`); required on the by-document response; `via_delegation_from`
  non-null only when `onBehalf != ""` (pure-delegation); terminal ⇒ all-false/null (`active == nil`
  early return). Clean schema split: by-document → `ApprovalInstanceByDocumentResponse` (viewer required);
  by-id → new bare `ApprovalInstanceResponse` (viewer absent), matching `GetInstanceHandler` passing
  `nil`. Generated Go + TS match the OpenAPI edit.
- **H-PRE-1 — PASS.** Delegator display name resolved off-tx in `mapInstanceResponse` (after the view tx
  closed), via the existing `displayNameReader` port, id-fallback on miss. No in-tx name/recording read.
- **Invariants — PASS.** No new capability, no write-path change, no migration; `LoadActiveDelegationsFor`
  tenant-scoped with `asOf`; `ErrInstanceNotVisible` → 404 mapping preserved.
- **Pure-fn correctness — PASS.** author-rule via CheckSoD; `priorSignoffs` = OTHER stages only;
  `has_signed` independent of `eligible`.
- **Test quality — PASS.** real-DB via testdb factory; `ctxWithIdentity` harness fix is a correct
  root-cause fix (seeds the platform actor key the TxRunner reads), not a hack.

Minor (non-blocking, not fixed): (1) `time.Now().UTC()` called directly in
`LoadInstanceByDocumentForViewWithViewer` + `requireInstanceVisible` rather than an injected clock —
plan.md permitted this; delegation-window tests seed real timestamps. (2) `ApprovalInstanceResponse`
duplicates ByDocument's non-viewer fields in the yaml (no `allOf`) — cosmetic; oapi schema idiom.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Worklist/inbox SQL is a 4th, delegation-blind re-expression of pool membership | Not on this read path; not loosened here. Named risk in the system-impact analysis §10.4 + ADR 0078 amendment | **M3** (approval kernel extraction): derive worklist from the kernel OR pin with a Go≡SQL parity test (SoD-mirror precedent) |
