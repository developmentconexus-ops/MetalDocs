# Feature F1 — Evidence

> **Milestone:** 1 · **Feature:** `f1-intx-resolution` · **Closed:** 2026-07-06
> **Contract:** `spec.md` (in-tx resolution of route_id/profile/content_hash; REV0 zero-gov; REV≥1 gates).

## What was implemented
- `SubmitDefaultsResolver` narrow port (`submit_defaults.go`) + Postgres impl on the approval repo
  (`postgres_approval_repository.go:1222` route query mirrors wrapper `repository.go:1801-1817`).
- `SubmitService.resolveActiveRoute` / `resolveContentFormData` run inside the submit tx before CAS
  (`submit_service.go`); nil-resolver → `ErrApprovalRouteMissing` (fail-closed); caller map never mutated.
- Governed rev gates: REV0 empty title → default "Criacao do documento"; REV≥1 empty → typed sentinels.
- Wiring `services.go` via `repo.(SubmitDefaultsResolver)` + existing `cdRead`.
- Producer matches consumer contract: an author POSTing `/submit` with no route_id/content_hash
  succeeds (verified live below).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Static | `go build ./...` | `EXIT 0` | — |
| Integration — resolve active route | `go test -tags integration ./internal/modules/documents/approval/application/...` → `TestSubmitResolvesActiveRoute_RealDB` | PASS (instance.route_id = seeded route) | **real** (testdb pg16) |
| Integration — bind head hash | `TestSubmitBindsHeadContentHash_RealDB` | PASS (head hash, not empty-form) | **real** |
| Integration — REV0 zero-gov | `TestSubmitRev0NoGovernanceData_RealDB` | PASS | **real** |
| Integration — REV≥1 reason gate | `TestSubmitRev1RequiresReason_RealDB` | PASS | **real** |
| Integration — explicit route+hash | `TestSubmitExplicitRouteAndHash_RealDB` | PASS | **real** |
| Integration — replay | `TestSubmitReplayDuplicate_RealDB` | PASS (same instance, no dup) | **real** |
| Runtime — REV0 zero-gov submit | docker curl `POST /submit` If-Match `"v0"` body `{}` | **201** `{instance_id, etag:"v1"}`; doc→`under_review`, revision_version 0→1, title auto "Criacao do documento" | **real** (docker) |
| Runtime — no active route | same, profile w/o route | **409** `state.approval_route_missing` (not 500) | **real** |

> Integration suite = 6/6 PASS on postgres:16 (pg17 wedged at initdb on this host; testdb bootstrap
> needs a connected superuser `metaldocs_app`). Live QA transcript: `scratch_qa/M1-live-qa.md`.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| REV0 empty-body → 201, ETag "v1", instance, under_review | yes | integration REV0 + live curl row |
| REV≥1 empty title/reason → typed 422 | yes | `TestSubmitRev1RequiresReason_RealDB` + F3 |
| Explicit route_id/content_hash honored | yes | `TestSubmitExplicitRouteAndHash_RealDB` |
| No active route → 409 ErrApprovalRouteMissing | yes | integration + live 409 row |
| Replay → same instance, no dup | yes | `TestSubmitReplayDuplicate_RealDB` |
| resolver-nil + explicit path unaffected | yes | existing approval unit tests green |
| build + test green | yes | build EXIT 0; approval pkgs `ok` |

## Review disposition
- Spec-compliance: PASS — resolution reads are non-recording SELECTs on the caller's tx (HS-PRE-1
  respected); ADR 0072 cross-module read via `CDFieldReader` only; `content_hash_at_submit` semantics
  unchanged (HS-2 boundary held).
- Code-quality: PASS — ISP port keeps the service off the fat repo; fail-closed on nil resolver.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| REV≥1 live path not driven via docker (long journey) | covered by real-DB integration `TestSubmitRev1RequiresReason_RealDB` | drive live when a revision-of-published fixture exists; owner documents/approval |
