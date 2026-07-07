# Feature F9 — `delegation` — plan.md

Ref: `spec.md` (this folder), plan.md `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F9 section (stale migration number `0291` corrected to real `0293`), design spec.md §4/§9/§11.

## Task 0 — ADR (4th and final due ADR for M2b)

- [ ] `wiki/decisions/0077-approval-delegation.md` — mirror ADR 0075/0076 structure (Context/Decision/Consequences/Alternatives/Rollback/References). Decision: delegation widens the eligibility-set INPUT to the existing `domain.CheckEligibility`/`domain.CheckSoD` predicates; never a parallel authz branch; no new capability; real-time in-tx revocation via DELETE + live window query.
- [ ] Add entry to `wiki/decisions/index.md`.

## Task 1 — Contract-first: OpenAPI + regen

- [ ] Edit `api/openapi/v1/openapi.yaml`: add `ApprovalDelegation` schema; `POST /approval/delegations` (create); `DELETE /approval/delegations/{id}` (revoke). Both tagged `approval`.
- [ ] Run the module's oapi-codegen regen target (`go generate ./internal/modules/documents/approval/api/...` or repo-root `go generate ./...` scoped). `go build ./...` — expect compile errors only from the NEW unimplemented `ServerInterface` methods (expected, TDD red).

## Task 2 — Migration (verify against real baseline first)

- [ ] Grep-confirm `db/migrations/` real max version (expect 0292 is latest; confirm 0289/0291 are unused gaps) before naming the file.
- [ ] `db/migrations/0293_approval_delegations.sql`: new `approval_delegations` table (tenant_id, delegator_id, delegate_id, starts_at, ends_at, reason, created_by, created_at; window + no-self CHECKs; RLS ENABLE+FORCE + tenant_isolation policy copied from 0288's shape); `ADD COLUMN IF NOT EXISTS on_behalf_of_user_id text` on `approval_signoffs` and `approval_review_verdicts`; extend `enforce_approval_sod()` (from migration 0290) to ALSO raise when `NEW.on_behalf_of_user_id IS NOT NULL AND NEW.on_behalf_of_user_id = author_id` (`CREATE OR REPLACE FUNCTION`, idempotent); index `(tenant_id, delegate_id, starts_at, ends_at)`. Ledger insert `ON CONFLICT DO NOTHING`.
- [ ] Grep-confirm no existing `approval_delegations` table / `on_behalf_of_user_id` column before writing (expect zero hits — new).

## Task 3 — Domain (TDD: failing test first)

- [ ] Failing unit tests: `domain/delegation_test.go` — `NewDelegation` self-delegation → `ErrSelfDelegation`; invalid window (`endsAt<=startsAt`) → `ErrInvalidDelegationWindow`; valid params → success + getters round-trip.
- [ ] Implement `domain/delegation.go`: `Delegation` struct + `DelegationParams` + `NewDelegation`.
- [ ] Failing unit test: `domain/eligibility_test.go` — new `ResolveEligibleIdentity(actorID, eligibleActorIDs, delegations)` — direct match returns `("", nil)`; no direct match but active delegation's delegator is in pool returns `(delegatorID, nil)`; no match anywhere returns `("", ErrActorNotEligible)`.
- [ ] Implement `ResolveEligibleIdentity` in `domain/eligibility.go` (calls the SAME unchanged `CheckEligibility` internally for both the direct try and each delegation try — no duplicated membership-loop logic).
- [ ] Failing unit test: `domain/sod_test.go` — extend existing table test file, add cases for the 4th `onBehalfOfUserID` param: acting-as-self unchanged (empty string, existing behavior byte-identical); `onBehalfOfUserID == authorUserID` (delegate acting for author) → `ErrAuthorCannotSign`.
- [ ] Implement: widen `CheckSoD` signature to `(authorUserID, actorUserID, onBehalfOfUserID string, priorSignoffs []Signoff) error`; update its 2 production call sites (Task 5) — this is a controlled, in-scope breaking change to an internal package function, not a public API.
- [ ] Failing unit tests: `domain/signoff_test.go` / `domain/review_verdict_test.go` — extend `NewSignoff`/`NewVerdict`/`MarshalJSON` assertions for the new `OnBehalfOfUserID` field (empty-string default, populated value round-trips, JSON key present).
- [ ] Implement: add `onBehalfOfUserID`/`OnBehalfOf()`/`OnBehalfOfUserID` to `domain/signoff.go` and `domain/review_verdict.go` (struct field, params field, constructor wiring, `MarshalJSON`).
- [ ] Run `go test ./internal/modules/documents/approval/domain/...` — PASS.

## Task 4 — Repository (interface + Postgres impl)

- [ ] `infrastructure/approval_repository.go` interface: add `InsertDelegation`, `DeleteDelegation`, `LoadActiveDelegationsFor`.
- [ ] `infrastructure/postgres_approval_repository.go`: implement all three (plain INSERT; `DELETE ... WHERE id=$1 AND tenant_id=$2 AND (delegator_id=$3 OR $4) RETURNING delegator_id` for ownership-scoped delete with an oversee-bypass boolean param; `SELECT ... WHERE tenant_id=$1 AND delegate_id=$2 AND starts_at<=$3 AND ends_at>$3`).
- [ ] Extend existing SELECT lists (`LoadInstance`, `LoadInstancesByIDs`, `LoadPriorSignoffs`, `LoadStageSignoffs`, `LoadStageVerdicts`) to include `on_behalf_of_user_id`, mapping into the domain objects' new field.
- [ ] Extend `InsertSignoff`/`InsertVerdict` fixed column lists to include `on_behalf_of_user_id`.
- [ ] Repository-level integration test (testdb factory, deferred live-run per Bounded Defers — write the test now, run later): `delegation_repository_integration_test.go` covering insert/delete/active-window-query semantics directly against Postgres (RLS, CHECK constraints).

## Task 5 — Application service composition (the core of F9)

- [ ] `application/delegation_service.go` (new): `DelegationService` with `CreateDelegation`, `RevokeDelegation`; wires `EventTypeDelegationGranted`/`EventTypeDelegationRevoked` governance events.
- [ ] `application/events.go`: add the two new `EventType` consts.
- [ ] `application/services.go`: add `Delegation *DelegationService` field; wire in `NewServices`.
- [ ] Failing unit/integration tests first (`delegation_service_test.go` with `MemoryEmitter`/fake repo, or integration test in Task 7) for `CreateDelegation`/`RevokeDelegation` happy path + `ErrDelegationNotFoundOrNotOwned` on foreign revoke attempt.
- [ ] `decision_service.go` (`RecordSignoff`): before the existing `domain.CheckEligibility` call, call `domain.ResolveEligibleIdentity(req.ActorUserID, activeStage.EligibleActorIDs, delegations)` where `delegations` comes from a new in-tx `s.repo.LoadActiveDelegationsFor(ctx, tx, req.TenantID, req.ActorUserID, s.clock.Now())` call gated behind "only query if direct membership fails" (avoid the extra round-trip on the common path — direct `CheckEligibility` first, delegation lookup only as a fallback). Wire the resolved `onBehalfOf` into the widened `domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, onBehalfOf, priorSignoffs)` call and into `SignoffParams.OnBehalfOfUserID`. Widen the governance event payload with `on_behalf_of`.
- [ ] `review_verdict_service.go` (`RecordVerdict`): identical mirrored change (same helper, same call shape) at its `domain.CheckEligibility`/`domain.CheckSoD` call sites.
- [ ] Run `go test ./internal/modules/documents/approval/application/...` — PASS (existing tests still green with `onBehalfOf=""` default; new tests for delegated path).

## Task 6 — HTTP layer

- [ ] `http/contracts/delegation.go` (new, or extend nearest existing contracts file): request/response DTOs matching the OpenAPI schema.
- [ ] `http/delegation_handler.go` (new): implements the 2 new generated `ServerInterface` methods; derives `delegator_id` from the auth context (`authz.MustActorID`), NEVER from the request body (privilege-escalation guard per spec.md Interview #11); maps `ErrSelfDelegation`/`ErrInvalidDelegationWindow` → 422, `ErrDelegationNotFoundOrNotOwned` → 404 in `http/errors.go`.
- [ ] `apps/api/cmd/metaldocs-api/permissions.go`: add 2 explicit tier-1 rows (`CapDocumentSignoff`, mirroring nearby runtime-verb rows).
- [ ] `go build ./...` clean — generated interface fully implemented.

## Task 7 — Integration tests (testdb factory)

- [ ] `tests/integration/approval/delegation_integration_test.go` (or module-local `application/delegation_integration_test.go` matching sibling F-features' placement convention — check F7/F8's actual test file location before picking): grant → delegate signs off successfully (`on_behalf_of_user_id` persisted) → revoke → SAME delegate blocked on a subsequent stage/instance; expired window ineligible; future window ineligible; overlapping windows both usable (union); SoD inheritance (delegate blocked when delegator is the author); non-owner revoke attempt → 404-equivalent error; oversee-holder can revoke someone else's delegation.
- [ ] Run — PASS (or deferred to live-DB session per Bounded Defers if no DB reachable in this session; attempt first, defer only if genuinely blocked).

## Task 8 — Full verification sweep

- [ ] `go build ./...`
- [ ] `go build -tags integration ./...`
- [ ] `go test -count=1 ./internal/modules/documents/approval/...`
- [ ] `go test -count=1 ./...` (grep zero FAIL)
- [ ] `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 violations)
- [ ] `go test ./scripts/api-lint/...`

## Task 9 — Self-review pass (explicit, per task brief)

- [ ] Confirm delegation composes with the SAME `authz.Require`/`CheckEligibility`/`CheckSoD` path — grep diff for zero new/parallel authz branches in the signoff/verdict paths.
- [ ] Confirm revocation is real-time (DELETE + fresh in-tx query at use-time), not a soft/cached flag.
- [ ] Confirm no column/field reused across unrelated semantic purposes (`on_behalf_of_user_id` is new on both tables, not repurposing an existing column).
- [ ] Confirm no silently swallowed error path in the new service/repository/handler code.

## Task 10 — Evidence + commit

- [ ] Write `evidence.md` (implementation summary, verification table, judgment calls, bounded defers, self-review confirmations, exact commit hash placeholder filled after commit).
- [ ] `git add` explicit touched paths only (never `-A` — repo has unrelated pending deletions/untracked files).
- [ ] Commit: `feat(approval): F9 delegation — capability-composed eligibility widening + SoD inheritance + ADR 0077 (W5)`.
