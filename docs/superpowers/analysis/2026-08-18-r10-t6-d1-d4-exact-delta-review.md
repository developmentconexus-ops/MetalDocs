# R10-T6 — D1→D4 Exact Bounded Delta Review

> **Status:** ACTIVE STAGING / DELTA REVIEW — **APPROVE / SUMMARY RATIFICATION MAY RESUME**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Reviewed target:** `docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary-rev2.md`  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This is an exact bounded review of D1→D4 only. It does not restart T6 and does not re-adjudicate frozen Product Contract/T1→T5 decisions.

Authority checked:

```text
Product Contract REV001
Whole-Product GCR A1–A10
4+1 ownership
T1
T2
T3
T3 D4 amendment
T4
T5
Decision Registry
Registry D4 amendment
```

## Verdict

```text
D1 = CLOSED
D2 = CLOSED
D3 = CLOSED
D4 = CLOSED

NEW MATERIAL FINDINGS = 0
FORMAL REOPEN T1/T2/T4/T5 = EMPTY
T3 REOPEN = CLOSED BY OPERATOR-RATIFIED D4 AMENDMENT
DELTA VERDICT = APPROVE
PLATFORM SUMMARY REV2 = MAY RETURN TO OPERATOR RATIFICATION
T7 = NOT OPEN
IMPLEMENTATION = BLOCKED
```

## D1 — CLOSED

`Platform Summary REV2` makes completed idempotency replay subject to the current request trust boundary:

```text
current ApplicationSession
+ current CSRF
+ current T3 permission/scope
+ minimum current result/resource disclosure predicate
```

Only response disclosure is re-authorized. The historical mutation is not re-executed and its former lifecycle eligibility is not re-proved.

This preserves both:

```text
T3 current/live Authorization authority
AND
idempotency = transport replay, not a second business transition
```

The replay store therefore cannot grant access by possession of an old key.

## D2 — CLOSED

The new GroupMembership read surface:

```text
GET /api/v1/groups/{group_id}/members
```

is current, cursor-paginated, bounded and protected by `access.manage` at Company scope.

It does not create a history owner, general User directory or new permission. GroupMembership remains Organization-owned semantic state while access-sensitive administration remains protected by Authorization exactly as T3 requires.

## D3 — CLOSED

`GET /api/v1/document-creation/options` is a purpose-built journey projection, not a reference-data platform.

It uses existing authorities:

```text
Area/DocumentType options
→ current document.create scopes + current active/eligibility predicates

Template options
→ document.read_effective + current Template role + target-type eligibility + current EFFECTIVE Revision

responsible-owner candidates
→ document.owner.manage + D4 eligible-target law
```

Ordinary author creation still defaults actor as responsible owner.

The projection exposes bounded references only, never admin configuration or a general UserProfile directory, and the real create command revalidates every selected reference.

No new Role/Permission or semantic owner is introduced.

## D4 — CLOSED

The operator-ratified T3 amendment defines exactly:

```text
eligible responsible-owner target
=
existing MetalDocs User
+ same Company
+ current eligibility = ENABLED
```

It explicitly rejects owner-as-grant semantics and provider-role inference.

The amendment also extends the existing T3 offboarding serialization law so owner assignment cannot commit after the target has already linearized to DISABLED.

This is a precision closure of the previous phrase, not a new AuthZ model.

## Cross-authority checks

### Product Contract

Still preserves:

```text
stable Document identity
responsible-owner current relationship
current-effective reader truth
immutable Submission governance
truthful offboarding/history
```

### T1 / ownership

No new semantic owner. Creation options, replay, membership list and HTTP selectors remain read/transport mechanisms over existing owners.

### T2

No new lifecycle transition or external provider call enters semantic atomicity. D4 owner replacement remains a local current-relation transition with concurrency preconditions.

### T3

Current grants/scopes remain sole access authority. Responsible owner remains a domain relationship predicate, not a grant. D1 specifically prevents replay from becoming stale access authority.

### T4

No content-identity rule changes.

### T5

No async/Search/effect rule changes.

## Non-material implementation proof note — idempotency and erasable PII

The exact public response DTOs remain implementation-contract design. The idempotency mechanism must not become an independent retention root for lawfully erasable `UserProfile` PII.

Therefore later OpenAPI/storage design must prove one of the equivalent bounded shapes:

```text
replayed creation responses are PII-minimized to stable semantic identifiers
OR
idempotency response persistence participates correctly in required profile-erasure cleanup/reconciliation
```

This is not a new product capability or T3/T4 reopen. It follows existing UserProfile erasure/privacy semantics and the rule that mechanism state cannot acquire independent semantic retention authority.

## Final delta gate

```text
C1→C8 = CLOSED
L1→L5 = CLOSED
D1→D4 = CLOSED
NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
```

The corrected `Platform Summary REV2` may now be presented for the required explicit operator summary ratification.

That ratification still does **not** authorize implementation. After ratification T6 must be promoted/reconciled/cleaned before T7 opens.
