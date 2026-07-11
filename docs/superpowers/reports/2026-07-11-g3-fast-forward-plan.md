# Plan — Unit 2.3 G3 fast-forward "Aprovar já"

**Spec:** docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md §1 R5 + §4 G3.
**Gate:** docs/superpowers/analysis/2026-07-11-g3-fast-forward-system-impact.md (🟢 Green).
**Owning module:** documents/approval (nested kernel; extraction is later unit — build inside).
**Invariants touched:** authz-capabilities, contract-first, multi-tenant, DB-enforced, H-PRE-1, freeze boundary, G2 invariant.

## Design (locked)

Two deliverables:
1. **Eligibility surfacing** — `RecordVerdict` already knows, inside its tx, when a verdict completes the stage (`QuorumApprovedStage`) and which stage becomes active (`instance.Active()` post-`AdvanceStage`). If next active stage is approval-kind AND actor resolves eligible on its pool (same `domain.ResolveEligibleIdentity` mechanism the signoff path uses), set `ReviewVerdictResult.FastForwardEligible=true` + `NextStageID`. Contract: `ReviewVerdictResponse` += `fast_forward_eligible` (bool) + `next_stage_id` (optional string).
2. **Fast-forward endpoint** — `POST /approval/instances/{instance_id}/stages/{stage_id}/fast-forward`. Body: verdict comment + signature ceremony (`password_token`, `content_hash`, optional signature method), headers Idempotency-Key + If-Match. One `runner.Do` tx executes: verdict core (verdict=`ready`) → assert stage completed (else problem 409 `fast-forward-stage-not-completed`) → assert next active stage approval-kind ∧ actor eligible (else 409 `fast-forward-not-eligible`) → freeze fires at existing call site (crossing boundary) → signoff core. Fail closed: any leg fails ⇒ whole tx rolls back, NO partial verdict. Two ledger rows (`approval_review_verdicts`, `approval_signoffs`) + their governance events, one tx, never collapsed.

Prereq refactor: behavior-preserving extraction of tx-scoped cores from `RecordVerdict` and `RecordSignoff` so fast-forward composes them without duplication. Off-tx reads (display name, signature payload resolution) stay off-tx per H-PRE-1 discipline; `authz.Require` stays in the writable tx (module norm). Both capabilities required in-tx: `CapApprovalReview` (verdict leg) + signoff capability (signature leg).

## Slices (sequential; failing test first per slice; commit per green slice)

### S1 — Eligibility detection (vertical)
- Files: `api/openapi/v1/openapi.yaml` (ReviewVerdictResponse), regen `internal/modules/documents/approval/api`, `application/review_verdict_service.go`, `http/contracts/review_verdict.go`, `http/review_verdict_handler.go`, unit + integration tests.
- Failing tests: verdict completes review stage → next approval stage, actor in pool ⇒ `fast_forward_eligible=true` + `next_stage_id`; actor NOT in pool ⇒ false; quorum not satisfied ⇒ false; next stage review-kind ⇒ false; instance fully approved (no next stage) ⇒ false.
- Done: tests green, contract regen clean, api-lint clean, no behavior change elsewhere.

### S2 — Tx-core extraction (behavior-preserving)
- Files: `application/review_verdict_service.go`, `application/decision_service.go`.
- Extract `recordVerdictInTx(...)` and `recordSignoffInTx(...)` (tx-scoped cores) callable from a shared tx; public methods become thin `runner.Do` wrappers. Off-tx preloads passed in as params.
- Done: NO new behavior; all existing unit + integration tests green unchanged.

### S3 — Fast-forward service (contract + application + unit tests)
- Files: `api/openapi/v1/openapi.yaml` (new path + DTOs + problem responses), regen, new `application/fast_forward_service.go` (+ domain errors), unit tests.
- Failing tests: happy path composes both cores, two results returned; stage-not-completed error; not-eligible error; ready-on-approval-stage still blocked (G2 regression); rollback on signature failure leaves no verdict.
- Done: unit tests green, api-lint clean.

### S4 — Handler + wiring + integration tests
- Files: `http/fast_forward_handler.go` (new), `http/routes_generated.go` wiring per module pattern, tier-1 route→capability map (`apps/api/cmd/metaldocs-api/permissions.go` or module equivalent), idemp route-template constant in `infrastructure/idempotency/postgres_signoff_idemp_store.go`, `http/contracts/`, integration tests `tests/integration/approval/fast_forward_integration_test.go` (testdb, `//go:build integration`).
- Failing tests (testdb): happy path 200 → DB: 1 verdict row + 1 signoff row + 2 governance events, frozen hash set, instance advanced/approved; not-eligible → problem+json 409, zero new rows (atomicity); replay via Idempotency-Key → same outcome, no dup rows; missing Idempotency-Key → 400 problem; cross-tenant → 404; SoD author → blocked.
- Done: integration green under canonical runner (no NEW failures vs 9 accepted RED).

### S5 — Wiki + evidence
- wiki-curator: update `wiki/modules/documents.md` approval section (fast-forward), Last verified stamp.
- evidence.md at `docs/superpowers/reports/2026-07-11-g3-fast-forward-evidence.md` with dispatch ledger + gate outputs (L0: build, api-lint strict, module boundaries; L1: go test + test-integration.ps1).

## Verification ladder
L0 `go build ./...` · `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` · `.\scripts\check-module-boundaries.ps1` → L1 `go test ./...` + `.\scripts\test-integration.ps1` (bar: no NEW failures). L2/L3 out of chip scope (feature unit; UI panel is unit 2.4).
