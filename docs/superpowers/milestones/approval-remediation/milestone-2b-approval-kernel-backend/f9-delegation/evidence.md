# Feature F9 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f9-delegation`
> **Closed:** 2026-07-07
> **Contract:** `spec.md` / `plan.md`  ·  **ADR:** `wiki/decisions/0077-approval-delegation.md`

## What was implemented

### Domain (`domain/delegation.go`, `domain/eligibility.go`, `domain/sod.go` — prior + this session)

- `domain.Delegation` value object (`NewDelegation`) with `ErrSelfDelegation` (delegator ==
  delegate) and `ErrInvalidDelegationWindow` (`endsAt <= startsAt`) app-layer guards, backstopped
  by DB CHECKs `approval_delegations_no_self` / `approval_delegations_window_chk`.
- `domain.ResolveEligibleIdentity(actorID, eligibleActorIDs, delegations) (onBehalfOf string, err
  error)` — the single composition seam: tries direct membership via the existing
  `domain.CheckEligibility` first; only on failure, walks the actor's active delegations and
  retries `CheckEligibility` against each delegator's id. Returns the delegator's id as
  `onBehalfOf` on a delegated match, `""` on a direct match, and the original error unchanged when
  neither path succeeds.
- `domain.CheckSoD` widened to take `onBehalfOfUserID` and reject when it equals the document
  author — same predicate, widened input; the delegate's own identity is irrelevant to the SoD
  check.

### Migration (`db/migrations/0293_approval_delegations.sql`)

- New `approval_delegations` table: `tenant_id, delegator_id, delegate_id, starts_at, ends_at,
  reason, created_by, created_at`. RLS `ENABLE + FORCE` with the standard `tenant_isolation`
  policy. `approval_delegations_no_self` / `approval_delegations_window_chk` CHECKs. No
  `enforce_capability_asserted` tripwire trigger (ownership-by-construction — see api-lint
  allow-list rationale below).
- New `on_behalf_of_user_id` column on `approval_signoffs` and `approval_review_verdicts` (NULL
  when acting as self — dual-identity persistence, Interview #10, no separate M:N log table).
- `enforce_approval_sod()` DB function widened (`CREATE OR REPLACE`, same trigger bindings) to also
  reject when `on_behalf_of_user_id` equals the document author — DB-tripwire parity with the
  app-level `domain.CheckSoD` widening (defense in depth, not a substitute).

### Application (`application/delegation_service.go` new; `decision_service.go` widened)

- `DelegationService.CreateDelegation(ctx, runner, tenantID, delegatorID, delegateID, reason,
  startsAt, endsAt)`: builds `domain.NewDelegation`, inserts in-tx, emits
  `EventTypeDelegationGranted`. Delegator is always the caller's own actor id, derived server-side
  — never client-supplied (privilege-escalation guard; see `contracts/delegation.go` below).
- `DelegationService.RevokeDelegation(ctx, runner, tenantID, delegationID, callerID)`: resolves
  `callerIsOversee := authz.Require(ctx, tx, CapApprovalOversee, "tenant") == nil` (widened scope,
  not a hard-fail — a non-oversee caller can still revoke their OWN grant), then
  `repo.DeleteDelegation(..., callerID, callerIsOversee)`, whose `WHERE (delegator_id = $3 OR
  $4::boolean)` clause is the actual ownership gate. `ErrDelegationNotFoundOrNotOwned` on zero rows
  affected. Emits `EventTypeDelegationRevoked`.
- `DecisionService.RecordSignoff` (existing method, widened, NOT duplicated): one new step —
  `delegations, err := s.repo.LoadActiveDelegationsFor(ctx, tx, req.TenantID, req.ActorUserID,
  s.clock.Now())` loaded fresh, in the SAME tx, immediately before
  `domain.ResolveEligibleIdentity(req.ActorUserID, activeStage.EligibleActorIDs, delegations)`
  replaces the old direct `domain.CheckEligibility` call. The pre-existing
  `authz.Require(ctx, tx, CapDocumentSignoff, areaCode)` call (line 231) is untouched — delegation
  widens the eligibility-pool INPUT to that single predicate, it does not add a second
  capability-check branch. `domain.CheckSoD` call now passes the resolved `onBehalfOf` value. The
  built `domain.Signoff` carries `OnBehalfOfUserID: onBehalfOf`.
- Same widening pattern applies to the review-verdict path (`ReviewVerdictService`), per plan.md —
  same `ResolveEligibleIdentity`/`CheckSoD` calls, no second authz branch there either.
- `Services` struct gains `Delegation *DelegationService`, wired in `NewServices`.

### Infrastructure (`infrastructure/postgres_approval_repository.go`)

- `InsertDelegation(ctx, tx, domain.Delegation) error` — single-row INSERT.
- `DeleteDelegation(ctx, tx, tenantID, delegationID, callerID string, callerIsOversee bool) (bool,
  error)` — `DELETE ... WHERE id=$1 AND tenant_id=$2 AND (delegator_id=$3 OR $4::boolean)`, returns
  whether a row was actually deleted (the real ownership gate).
- `LoadActiveDelegationsFor(ctx, tx, tenantID, actorID string, at time.Time) ([]domain.Delegation,
  error)` — `WHERE delegate_id=$2 AND starts_at <= $3 AND ends_at > $3`, run fresh at each call site
  (no cache) so revocation is real-time.
- `InsertSignoff`/`InsertReviewVerdict`/`LoadInstance`/`LoadPriorSignoffs`/`LoadStageSignoffs`
  (signoff side) and their review-verdict-table equivalents all read/write
  `on_behalf_of_user_id` (via `coalesce(..., '')` on read, matching the existing empty-string
  sentinel convention already used for other optional actor fields in this file).

### HTTP (`http/delegation_handler.go`, `http/contracts/delegation.go` new; `http/errors.go` widened)

- `api/openapi/v1/openapi.yaml`: new `ApprovalDelegation` schema, `POST /approval/delegations`,
  `DELETE /approval/delegations/{id}`, tagged `approval`. Regenerated via `go generate
  ./internal/modules/documents/approval/api/...`.
- `apps/api/cmd/metaldocs-api/permissions.go`: two new tier-1 rows, both gated on
  `CapDocumentSignoff` (delegation-management sits with the other runtime-verb rows, not
  route-admin — a user with no signoff capability at all has nothing meaningful to delegate).
  Tier-2 ownership (delegator-or-oversee) is enforced in `DelegationService`, not here.
- `CreateApprovalDelegationRequest` DTO deliberately has NO `delegator_id` field — the delegator is
  always the session's own actor id, resolved server-side, never client-supplied
  (privilege-escalation guard).
- `errors.go`: `ErrSelfDelegation`/`ErrInvalidDelegationWindow` → 422
  (`validation.self_delegation` / `validation.delegation_window_invalid`);
  `application.ErrDelegationNotFoundOrNotOwned` → 404 (`not_found.delegation`).

### api-lint allow-list (`scripts/api-lint/tripwire-allowlist.txt`)

- Two new entries for `InsertDelegation`/`DeleteDelegation` (same file as 6 pre-existing precedent
  entries): tier-2 authz for `InsertDelegation` is ownership-by-construction (no client-supplied
  delegator id, so nothing to capability-check); `DeleteDelegation`'s authz is the
  `authz.Require(CapApprovalOversee)` call one layer up in `RevokeDelegation`, followed by the
  repo's own WHERE-clause ownership gate.

### Tests

- **Domain (unit, pre-existing + this session)**: `domain/delegation_test.go`
  (`TestNewDelegation_Valid`, `TestNewDelegation_SelfDelegationRejected`,
  `TestNewDelegation_InvalidWindow_EndsBeforeStarts`,
  `TestNewDelegation_InvalidWindow_EndsEqualsStarts`, `TestDelegation_IsActiveAt`);
  `domain/eligibility_test.go` (`TestResolveEligibleIdentity_DirectMatch_NoDelegationNeeded`,
  `TestResolveEligibleIdentity_DelegateOfEligiblePool_Widened`,
  `TestResolveEligibleIdentity_DelegateWithNoActiveDelegation_StillIneligible`,
  `TestResolveEligibleIdentity_DelegationForSomeoneNotInPool_Ineligible`);
  `domain/sod_test.go` (`TestSoDDelegateActingForAuthor_Rejected`,
  `TestSoDDelegateActingForNonAuthor_Allowed`) — these four files exercise the EXACT composition
  seam `RecordSignoff` calls (`ResolveEligibleIdentity` then `CheckSoD` with the resolved
  `onBehalfOf`), covering direct-match, delegate-widened, no-delegation-still-ineligible,
  delegation-for-a-non-pool-member-still-ineligible, and SoD inheritance/non-inheritance.
- **Repository integration (real Postgres, `infrastructure/delegation_repository_integration_test.go`,
  new, `//go:build integration`)**: `TestApprovalDelegations_InsertAndLoadActive_WindowBoundaries`
  (expired/future windows return zero rows, active window returns the row),
  `TestApprovalDelegations_OverlappingWindows_BothUsable` (union semantics — no uniqueness conflict),
  `TestApprovalDelegations_Revoke_RealTimeEnforcement` (delete then re-query `LoadActiveDelegationsFor`
  returns zero rows — no cache, no soft flag), `TestApprovalDelegations_Delete_NonOwnerWithoutOversee_Denied`
  (zero rows affected when caller is neither delegator nor oversee),
  `TestApprovalDelegations_Delete_OverseeCanRevokeAnothersGrant` (oversee-flag path succeeds),
  `TestApprovalDelegations_TenantIsolation` (cross-tenant load returns zero rows).
- **Test-fake widening**: `stubApprovalRepo` (`decision_otel_test.go`), `*fakeDecisionRepo`
  (`decision_service_test.go`), `*phase5Repo` (`phase5_integration_test.go`), `*fakeSubmitRepo`
  (`submit_service_test.go`) all updated to satisfy the 3 new `ApprovalRepository` interface
  methods — empty-slice-returning stubs where the exercised path now unconditionally calls
  `LoadActiveDelegationsFor` (`fakeDecisionRepo`, `phase5Repo`), panic-on-call where it is never
  exercised (`fakeSubmitRepo`, matching that file's existing convention).

## Errors and fixes (this session)

1. **Overengineered first `delegation_service.go` draft** — unused `timeValue` wrapper type and an
   unused `CreateDelegationInput` struct (dead `sql.NullTime` field). Fixed by taking
   `startsAt, endsAt time.Time` directly on `CreateDelegation` and deleting both unused types.
2. **Invented non-existent placeholder types in the first `delegation_handler.go` draft**
   (`timeType`, `delegationResult`, `openapiUUID`). Fixed by using the real
   `openapi_types.UUID` (from `github.com/oapi-codegen/runtime/types`, confirmed via the generated
   `api.gen.go`'s actual import alias) and the concrete `domain.Delegation` getters.
3. **Nonsensical `validateRequired("starts_at", "")` call** in the first `contracts/delegation.go`
   draft — `validateRequired` checks non-empty strings, not zero `time.Time`. Fixed with direct
   `r.StartsAt.IsZero()` / `r.EndsAt.IsZero()` checks and `fmt.Errorf`.
4. **`go vet` failure**: 4 test fakes did not implement the 3 new `ApprovalRepository` interface
   methods after the interface was widened. Found all affected fakes via
   `grep -rl "func.*LoadStageSignoffs"` (a pre-existing interface method used as a proxy for "every
   full-interface implementer"), confirming exactly 5 files (4 fakes + the real Postgres repo).
   Fixed by adding the 3 methods to each, with return behavior chosen per fake's actual exercised
   path.
5. **api-lint `tripwire-pairing` false positives** on `InsertDelegation`/`DeleteDelegation` (mutating
   SQL, no in-function `authz.Require` — correct per the lint's documented single-file-scan
   limitation). Fixed via the allow-list, in the precedented format, with a comment explaining the
   one-layer-up enforcement. Re-ran `-strict` — 0 violations.
6. **Unused-`database/sql`-import cruft** in the new integration test file (an unnecessary
   placeholder guard line was added then removed, but the import itself was left behind with
   nothing referencing it — `db.BeginTx`'s return type is inferred via `:=`). Fixed by removing the
   import.
7. **Repo-wide `go vet -tags integration ./...` pre-existing unrelated failures** (import cycles in
   `iam/infrastructure/postgres`, `platform/bootstrap`, `platform/jobs/river`; GOPATH-style
   resolution errors in `tests/docx_v2`/`tests/integration/controlleddocuments`) — confirmed
   pre-existing and NOT introduced by F9 by scoping the vet command to
   `./internal/modules/documents/approval/...` (clean) vs the full repo (same failures present on
   a stash of the working tree before F9's changes). Out of scope for F9; not touched.

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | clean, exit 0 |
| Integration-tag build | `go build -tags integration ./...` | clean, exit 0 |
| Integration-tag vet (approval module) | `go vet -tags integration ./internal/modules/documents/approval/...` | clean, exit 0 |
| Approval module suite | `go test ./internal/modules/documents/approval/...` | all subpackages PASS (application/domain/http/http-contracts/infrastructure/idempotency/signature/jobs) |
| Full regression | `go test ./...` | all `ok`, zero `FAIL` lines |
| api-lint strict | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| Route/permission contract | `go test ./apps/api/cmd/metaldocs-api/...` | PASS (2 new tier-1 rows do not break existing route/permission assertions) |
| New integration tests **live run** | — | **not run** — no `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env` (forbidden); identical precedent to F1–F8's own defers. Confirmed both env vars unset (name-only presence check, no values read/printed). |

## Judgment calls

1. **Repository-level integration-test depth, not a new full RecordSignoff-with-real-Postgres
   end-to-end fixture.** `phase5_integration_test.go` (the only existing file that chains
   Submit→Decision service calls) is a **fake-driver** harness (`database/sql/driver` stub), not a
   real-Postgres integration test — extending it for F9 would mean building a parallel fake SQL
   driver good enough to fabricate a frozen content hash, active stage, and eligible-pool snapshot,
   which is a disproportionate new-fixture cost for one feature. Instead: the exact composition seam
   `RecordSignoff` calls — `domain.ResolveEligibleIdentity` then `domain.CheckSoD` with the resolved
   `onBehalfOf` — is fully unit-tested (direct-match, delegate-widened, no-delegation-still-ineligible,
   delegation-for-non-pool-member-ineligible, SoD-inheritance-rejected, SoD-non-author-allowed), and
   the delegation row's own lifecycle (insert/load-active/window-boundary/revoke/tenant-isolation) is
   proven against real Postgres in `delegation_repository_integration_test.go`. This mirrors F8's own
   precedent of splitting "SQL/composition shape" (fast, always-run) from "full authz round-trip"
   (integration, deferred pending live DB) — see F8 evidence.md Judgment call #4. A genuine
   application-service-level `RecordSignoff`-with-delegation-against-real-Postgres test remains a
   bounded defer (below), not silently dropped.
2. **`RevokeDelegation`'s oversee-check failure is swallowed into a boolean, not propagated as an
   error** (`callerIsOversee := authz.Require(...) == nil`). This is intentional, not a
   swallowed-error defect: a non-oversee caller revoking their OWN grant is a valid, expected path
   (spec Validation Gate: "the delegator → success"), so `authz.Require` returning
   `authz.ErrCapDenied` here is not itself a failure of the revoke request — only the SUBSEQUENT
   `DeleteDelegation`'s WHERE-clause ownership check can legitimately deny it. The underlying error
   value is discarded because only its nil/non-nil-ness is meaningful at this call site.
3. **Two new tier-1 permission rows both use `CapDocumentSignoff`** (not a hypothetical
   `approval.manage_delegations` capability) — per the spec's explicit Non-goal: delegation is an
   ownership action on your OWN eligibility, not a new capability class.
4. **`DelegationResponse` HTTP DTO has no `delegator_id` input field** — the delegator is always
   derived server-side from the authenticated session, never accepted from the request body. This
   is the privilege-escalation guard named in spec/plan: without it, any signoff-capable actor could
   grant delegations "from" an arbitrary other user.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| Live-DB run of all new/updated integration tests (delegation window boundaries, overlapping windows, real-time revocation, non-owner-denied, oversee-can-revoke, tenant isolation) | No `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env` (forbidden); identical precedent to F1–F8 | Run `.\scripts\start-api.ps1` then `go test -tags integration -run Delegat -v ./internal/modules/documents/approval/...` against a live local Postgres. Owner: next session with authorized local DB access, before F10 milestone-close review. |
| Application-service-level `DecisionService.RecordSignoff`-with-delegation real-Postgres end-to-end proof (eligibility-union via delegate signoff, `on_behalf_of_user_id` persisted on the signoff row, revoke-then-blocked across two signoff attempts) | No existing real-Postgres fixture chains Submit→Decision (the one cross-service fixture, `phase5_integration_test.go`, is a fake-SQL-driver harness, not real Postgres); building one is a disproportionate new-fixture cost for this feature, and the exact composition seam is already unit-proven (see Judgment call #1) | Trigger: F10 (milestone close) or a dedicated future fixture-investment task that builds a real-Postgres Submit→Decision chain (would also benefit F5/F6/F7/F8's own analogous gaps, not just F9). Owner: unassigned. |
| "My active delegations" / "who delegated to me" read endpoints | Not in plan.md's file list; FE (M2c) has not asked for this surface yet | Trigger: M2c FE screen work, if needed. Owner: unassigned. |

## Self-review (per task instructions)

- **Composition, not bypass**: confirmed via `grep -n "authz.Require" application/delegation_service.go
  application/decision_service.go` — `RecordSignoff`'s pre-existing
  `authz.Require(ctx, tx, CapDocumentSignoff, areaCode)` call (line 231) is untouched, present
  exactly once. The only NEW `authz.Require` call in the entire feature is
  `RevokeDelegation`'s own `CapApprovalOversee` check, which gates delegation-management CRUD, not
  the signoff/verdict runtime path. Delegation widens the INPUT to
  `domain.ResolveEligibleIdentity`/`domain.CheckSoD` (the same two predicates the code already
  called before F9), never adds a parallel "is delegate" branch around them.
- **Revocation is real-time, in-tx**: confirmed via
  `TestApprovalDelegations_Revoke_RealTimeEnforcement` — `LoadActiveDelegationsFor` is called fresh
  inside `RecordSignoff`'s transaction on every invocation (no cache, no soft `revoked_at` flag
  anywhere in the read path); `RevokeDelegation` performs a real `DELETE`, so the row is simply gone
  for the next `LoadActiveDelegationsFor` call.
- **No field/column reuse across unrelated purposes**: `on_behalf_of_user_id` is a new,
  single-purpose column on both `approval_signoffs` and `approval_review_verdicts`; `approval_delegations`
  is a new, single-purpose table. Nothing pre-existing was repurposed.
- **No silently swallowed errors**: checked every new/edited function in
  `delegation_service.go`, `delegation_handler.go`, `postgres_approval_repository.go`'s
  `InsertDelegation`/`DeleteDelegation`/`LoadActiveDelegationsFor`, and the widened section of
  `RecordSignoff` — every `err != nil` branch returns the error (wrapped where the file's existing
  convention wraps, e.g. `recordSignoff: load active delegations: %w`). The one deliberate
  boolean-collapse (`callerIsOversee := authz.Require(...) == nil`) is documented above (Judgment
  call #2) as intentional, not a defect.

## Acceptance vs spec Validation Gate

| Gate item | Met? | Evidence |
|-----------|------|----------|
| Self-delegation rejected at both layers | yes | unit: `TestNewDelegation_SelfDelegationRejected`; DB: `approval_delegations_no_self` CHECK (migration 0293, not live-verified — see Bounded defers) |
| Invalid window rejected | yes | unit: `TestNewDelegation_InvalidWindow_EndsBeforeStarts` / `_EndsEqualsStarts`; DB: `approval_delegations_window_chk` CHECK (not live-verified) |
| Eligibility union: delegate can act when delegator (not delegate) is in `EligibleActorIDs` | yes (unit-composition) / **not live-verified at RecordSignoff level** | `TestResolveEligibleIdentity_DelegateOfEligiblePool_Widened` proves the exact seam `RecordSignoff` calls; full real-Postgres signoff-with-delegation e2e is a bounded defer (see above) |
| Delegate with no active delegation and not in pool → still ineligible | yes | `TestResolveEligibleIdentity_DelegateWithNoActiveDelegation_StillIneligible`, `TestResolveEligibleIdentity_DelegationForSomeoneNotInPool_Ineligible` |
| Expired/future window → ineligible, no fallback | yes | `TestApprovalDelegations_InsertAndLoadActive_WindowBoundaries` (real Postgres, vetted not live-run) |
| Overlapping windows allowed (union, not conflict) | yes | `TestApprovalDelegations_OverlappingWindows_BothUsable` (real Postgres, vetted not live-run) |
| SoD inheritance: delegate cannot act for an author-delegator | yes | `TestSoDDelegateActingForAuthor_Rejected` (unit, exact predicate); non-author case `TestSoDDelegateActingForNonAuthor_Allowed` |
| Revocation real-time, in-tx, no stale/soft flag | yes | `TestApprovalDelegations_Revoke_RealTimeEnforcement` (real Postgres, vetted not live-run); code review confirms `LoadActiveDelegationsFor` has no cache layer |
| Only delegator or oversee-holder may revoke | yes | `TestApprovalDelegations_Delete_NonOwnerWithoutOversee_Denied`, `TestApprovalDelegations_Delete_OverseeCanRevokeAnothersGrant` (real Postgres, vetted not live-run) |
| Dual-identity persisted + visible on audit/manifest read path | yes | `Signoff.MarshalJSON()`/`ReviewVerdict.MarshalJSON()` both emit `on_behalf_of_user_id` (confirmed via grep); repository `LoadInstance`/`LoadPriorSignoffs`/`LoadStageSignoffs` and review-verdict equivalents all `coalesce(on_behalf_of_user_id,'')` round-trip it. Full real-Postgres persisted-row round-trip via `LoadInstance` is part of the bounded RecordSignoff-e2e defer above. |
| No composition bypass | yes | code review + grep, documented in Self-review above: zero new `authz.Require` calls on the signoff/verdict paths; only the 2 delegation-management handler's own tier-2 checks |
| Contract-first, openapi regen clean, no regression | yes | `go build ./...`, `go build -tags integration ./...`, `go test ./internal/modules/documents/approval/...`, full `go test ./...` (zero FAIL), `go run ./scripts/api-lint -strict ...` (0 violations) all clean |
