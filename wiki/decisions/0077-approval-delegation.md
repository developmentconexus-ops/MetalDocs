# ADR 0077 — Approval Delegation (Eligibility-Widening, Not a New Authz Path)

> **Status:** Accepted
> **Date:** 2026-07-07
> **Scope:** Time-windowed delegation of eligibility for review/approval stages, capability-composed,
> real-time revocable.
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f9-delegation/`
> **Key files:**
> - `internal/modules/documents/approval/domain/delegation.go` — `Delegation` value object
> - `internal/modules/documents/approval/domain/eligibility.go` — `ResolveEligibleIdentity`
> - `internal/modules/documents/approval/domain/sod.go` — widened `CheckSoD`
> - `internal/modules/documents/approval/application/delegation_service.go` — grant/revoke
> - `internal/modules/documents/approval/application/decision_service.go` — signoff call site
> - `internal/modules/documents/approval/application/review_verdict_service.go` — verdict call site
> - `db/migrations/0293_approval_delegations.sql`

## Context

The design spec (`docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §4, "W5") asks for
delegation-of-authority: a user (delegator) authorizes another user (delegate) to act with the
delegator's eligibility for a scoped window, so approvals do not stall on a single unavailable person.
The spec is explicit about the shape this must NOT take: "Delegate acts *as themselves on behalf of* —
signoff records both identities. No credential sharing, no role impersonation."

MetalDocs' authz model (ADR 0022) is a two-tier PDP: tier-1 route→capability (middleware), tier-2
capability×area in-tx (`authz.Require`), DB tripwire last line. Neither tier reasons about roles. A
delegation feature that added a third "is this person a delegate, if so skip the real check" branch
alongside `authz.Require` would violate the single-source-of-truth invariant ADR 0022 exists to
protect — exactly the failure mode this ADR is written to rule out.

Separately, this module already has a narrower, per-instance predicate beneath the capability layer:
`domain.CheckEligibility(actorUserID, eligibleActorIDs []string) error` answers "is this specific actor
in this specific stage instance's frozen worklist pool" (`approval_stage_instances.eligible_actor_ids`,
snapshotted at submit time). It is called identically from `DecisionService.RecordSignoff` and
`ReviewVerdictService.RecordVerdict`, always immediately after `authz.Require` and always immediately
before `domain.CheckSoD` (Separation of Duties — F7 unified this into one shared predicate, called
identically from both services and mirrored by one DB trigger, `enforce_approval_sod()`, migration
0290).

## Decision

### 1. Delegation widens the INPUT to the existing eligibility predicate; it does not add a new predicate

`domain.CheckEligibility`'s signature is unchanged. A new pure function,
`domain.ResolveEligibleIdentity(actorUserID string, eligibleActorIDs []string, delegations []Delegation) (onBehalfOf string, err error)`,
tries `CheckEligibility(actorUserID, eligibleActorIDs)` first (the common, non-delegated case); on
`ErrActorNotEligible`, it tries `CheckEligibility(delegation.DelegatorID, eligibleActorIDs)` for each
currently-active delegation whose `DelegateID == actorUserID` — same function, same pool, a different
candidate identity. The first delegation whose delegator is in the pool wins; `onBehalfOf` is set to
that delegator's ID (empty string when acting as self). There is no code path where a delegate
proceeds without a real pool match being found for *someone* — delegation only ever supplies additional
candidate identities to the one real check.

The application services (`RecordSignoff`, `RecordVerdict`) load the actor's active delegations via a
new in-tx repository read, `LoadActiveDelegationsFor(ctx, tx, tenantID, delegateID, asOf)`, only as a
fallback after the direct membership check fails — a real-time, same-transaction read, not a cached or
off-tx lookup.

### 2. The delegate must independently hold the real capability; delegation grants no capability

`authz.Require(ctx, tx, CapDocumentSignoff|CapApprovalReview, areaCode)` — already present, unchanged —
still runs first in both services and still gates on the delegate's OWN capability grant in that area.
Delegation never substitutes for or is consulted by this call. What delegation supplies is narrower:
whether a specific stage instance's frozen worklist counts this specific delegate as a match. A
delegate who does not independently hold `document.signoff`/`approval.review` in the relevant area is
rejected at the capability gate exactly as anyone else would be, before eligibility is ever evaluated.
No new capability (`approval.delegate` or similar) is introduced — delegation-of-authority is an
ownership action (delegate your own eligibility slot), not a new capability class.

### 3. SoD inheritance: `CheckSoD` gains a 4th parameter, not a parallel rule

`domain.CheckSoD(authorUserID, actorUserID string, priorSignoffs []Signoff) error` only ever compared
the acting actor to the author. Delegation introduces a genuinely new identity the rule must reason
about: the delegator, whose eligibility slot is being filled. The signature widens to
`CheckSoD(authorUserID, actorUserID, onBehalfOfUserID string, priorSignoffs []Signoff) error`; the rule
becomes "reject if `actorUserID == authorUserID` (unchanged) **or** `onBehalfOfUserID != "" &&
onBehalfOfUserID == authorUserID` (new)". Both production call sites
(`decision_service.go`, `review_verdict_service.go`) pass the `onBehalfOf` value resolved in Decision
1 (empty string when acting as self, byte-identical to today's behavior). This is one shared predicate
gaining one more condition, not a second SoD rule living beside the first.

The DB tripwire (`enforce_approval_sod()`, migration 0290) is widened symmetrically in the same
migration that adds the `on_behalf_of_user_id` column: it already compares `NEW.actor_user_id` to the
document's author; it now also raises when `NEW.on_behalf_of_user_id` is non-null and equals the
author. Without this, a delegate acting on behalf of the author would have `actor_user_id != author_id`
and the trigger would silently pass despite the delegator being the author — closing that gap at the
DB layer keeps "DB enforces invariants... tripwire last line" symmetric with the app-level fix.

### 4. Dual identity is persisted on the existing signoff/verdict rows, not a new usage-log table

`approval_signoffs` and `approval_review_verdicts` each gain one new nullable column,
`on_behalf_of_user_id text` — NULL when the actor acted as themselves (the default, unchanged wire
shape), the delegator's user_id when eligibility was satisfied via delegation. This is additive to the
tables that are already the audit record for who acted and when; no new M:N delegation-usage table is
introduced. `approval_delegations` itself (the new grant/window table) is not a usage log — "which
signoffs used which delegation" is reconstructible from `(actor_user_id, on_behalf_of_user_id,
signed_at/verdict_at)` against the delegation's window.

### 5. Revocation is real-time and in-tx, not a soft flag

There is no `revoked_at` column read by one path and ignored by another. `DELETE /approval/delegations
/{id}` hard-deletes the row. `LoadActiveDelegationsFor`'s `WHERE starts_at <= $asOf AND ends_at > $asOf`
is evaluated fresh, in the same transaction as the signoff/verdict write, on every call — there is
exactly one read path for "is this delegation currently active," and it always runs at actual use-time
against current data, never a cache.

### 6. Grant/revoke ownership, not a new capability

Only the delegator, or a `CapApprovalOversee` holder, may revoke a delegation (`DeleteDelegation`'s
WHERE clause enforces this atomically). Overlapping windows for the same delegator are explicitly
allowed (union semantics — any one active delegation is sufficient); self-delegation
(`delegator_id == delegate_id`) is rejected at both the app layer (`domain.NewDelegation`) and the DB
layer (`approval_delegations_no_self` CHECK).

## Consequences

- **Positive:** delegation cannot diverge from the real eligibility/SoD rule by construction — it is
  additional input to the same two functions every non-delegated call already goes through, not a
  second implementation of the rule that could drift.
- **Positive:** revocation is provably real-time — the only read of "is this delegation active" happens
  in-tx at the moment of use.
- **Positive:** no new capability, no registry-size bump, no new tripwire arm — delegation rides
  entirely on the existing `CapDocumentSignoff`/`CapApprovalReview`/`CapApprovalOversee` capabilities.
- **Neutral:** the delegation lookup is a fallback query only (tried after the direct-membership check
  fails), so the common non-delegated path pays no additional round-trip.
- **Negative:** `CheckSoD`'s signature change touches its two production call sites and its existing
  test file — a controlled, in-scope breaking change to an internal package function, not a public API,
  and confirmed via grep to have exactly 2 production + 3 test call sites, all updated together.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| A parallel "is this actor a delegate of someone eligible" branch checked separately from `CheckEligibility` | Rejected | Exactly the divergence risk ADR 0022 exists to prevent — two implementations of "who may act" that could disagree. |
| A new `approval.delegate` capability | Rejected | Delegation is an ownership action on the delegator's own eligibility slot, not a new class of permission; spec.md's non-goals explicitly exclude a new capability. |
| Add `onBehalfOf` as a 4th positional parameter vs. a small struct/options type on `CheckSoD` | Positional param chosen | Only 2 production call sites; a struct type would be over-engineering for a single added bool-ish string, and keeps the diff minimal and reviewable. |
| Soft `revoked_at` flag on `approval_delegations`, read-and-check by the application | Rejected | Exactly the "soft flag a stale read could miss" failure class named in the task brief; a hard DELETE plus a fresh in-tx window query has no cache to go stale. |
| A separate delegation-usage log table (M:N, one row per use) | Rejected | The existing `approval_signoffs`/`approval_review_verdicts` rows are already the audit record for each use; a new table would duplicate data reconstructible from the existing rows plus the delegation window. |
| A distinct `EventTypeDelegationUsed` governance event, separate from the signoff/verdict event | Rejected | The existing `EventTypeSignoffRecorded`/`EventTypeReviewVerdictRecorded` events already carry the full audit trail for the action; adding `on_behalf_of` to that payload avoids a second event that could drift out of sync with the real action record (mirrors F7's precedent of extending an existing event's payload rather than inventing a new event type). |

## Rollback

Additive at the schema level: `approval_delegations` is a new table (drop it); `on_behalf_of_user_id`
is a new nullable column on both existing tables (drop the columns); the `enforce_approval_sod()`
trigger-function widening is a `CREATE OR REPLACE` (revert to the migration-0290 body). Reversible by
reverting the two eligibility/SoD call-site changes in `decision_service.go`/`review_verdict_service.go`
(both fall back to their pre-F9 unchanged behavior when no delegation rows exist, since
`ResolveEligibleIdentity` with an empty delegations slice behaves identically to a direct
`CheckEligibility` call), removing `DelegationService` and its HTTP routes, and dropping the new
migration — provided no code has come to depend on `on_behalf_of_user_id` being populated by then.

## References

- `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §4 ("Delegation (W5)"), §9 (ADR
  list), §11 (locked constraints)
- `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f9-delegation/spec.md`
- `wiki/decisions/0022-authz-capability-coherence.md` (single-source-of-truth invariant this ADR
  protects)
- `wiki/decisions/0075-approval-oversee-visibility.md` (structure template; F3/F8 milestone sibling)
- `wiki/decisions/0076-approval-freeze-boundary-and-choke-point-concurrency.md` (F5 milestone sibling;
  immediately preceding ADR in sequence)
- `db/migrations/0290_sod_unified_trigger.sql` (F7 — the DB tripwire this ADR extends)
