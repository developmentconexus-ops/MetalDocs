# Feature F4 — Evidence

> **Milestone:** 1 · **Feature:** `f4-delete-finalize` · **Closed:** 2026-07-06
> **Contract:** `spec.md` (delete finalize wrapper chain from Go; retain domain sentinels; one entrypoint).

## What was implemented
- OpenAPI `/finalize` path removed (absorbed dirty-tree pre-work); `/submit` description records the
  removal; `oapi-codegen` regenerated (absorbed).
- Go wrapper surface deleted: `finalizeDocument` handler, its route registration, finalize-only
  contract types, and the live `GetFinalizePrereqs` repository method (its query re-homed in F1).
- Four domain sentinels retained (now submit-owned, F1/F3). Provenance comments retained.
- Unrelated `FreezeFinalizer` / render `MarkFailed(finalize=true)` / legacy-state rejection untouched.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Grep gate — no live finalize symbol | `grep 'func.*finalize\|GetFinalizePrereqs\|/finalize\|finalizeDocument'` under `internal/modules/documents/approval` | 3 hits, **all comments** (`postgres_approval_repository.go:1222`, `submit_defaults.go:13`, `submit_service.go:375`); zero live symbols | **real** |
| OpenAPI — no finalize path | `grep -i finalize api/openapi/v1/openapi.yaml` | 1 hit: deprecation note in `/submit` description; **no `/finalize` path** | **real** |
| Static | `go build ./...` | `EXIT 0` (nothing references deleted symbols) | — |
| Regression — approval suite | `go test ./internal/modules/documents/approval/...` | all pkgs `ok` | **real** |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| grep finalize → comments only, zero live symbols | yes | grep-gate row |
| OpenAPI has no `/finalize` path | yes | openapi grep row |
| build + test green | yes | build EXIT 0; approval `ok` |
| four sentinels retained + exercised | yes | F1 integration + F3 mapping |

## Review disposition
- Spec-compliance: PASS — deletion is complete (no live symbol) yet the error contract survives via the
  retained sentinels; ADR 0073 §1 satisfied.
- Code-quality: PASS — safe because F1 re-homed the only real logic the wrapper held; the substring
  false-positives (`FreezeFinalizer`, render `finalize=true`, legacy-state guard) were correctly left
  intact (different concepts).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `document.finalize` capability-row DB cleanup | submit never checks it; harmless dormant row | migration hygiene sweep; owner iam/authz — out of M1 scope |
