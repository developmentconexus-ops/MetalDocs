---
id: my-work-governance-identification-read
kind: authority
owner: architecture
summary: Bounded operator-ratified T11 precision for human-recognizable and human-stable governance-work projection without moving Governance Case authority into My Work.
---

# My Work governance read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified:** 2026-08-23  
> **Impacts:** T6 canonical journeys/read-model meaning, T8-E executable wire shape/order, T8-F frontend consumption.  
> **Implementation:** BLOCKED.

## 1. Decision outcome

DevelopmentConexus Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ bounded read-projection corrections
```

The accepted My Work boundary remains sound:

```text
/work
→ actor-relevant authoring/governance projections
→ owner lens re-reads current/exact truth
```

Two bounded defects were found in the governance projection:

```text
F1  enough technical identity for routing, insufficient governed-subject identity for human recognition
F2  deterministic UUID ordering, but no human-legible ordering model for a cursor-paginated queue
```

The current wire is therefore refined to:

```text
WorkGovernanceItem {
  governance_attempt_id:Uuid,
  subject_kind:GovernanceSubjectKind,
  document:DocumentReference,
  revision:RevisionReference,
  created_at:UtcInstant
}

listGovernanceWork canonical order
  document.code ASC,
  governance_attempt_id ASC
```

Existing component reused unchanged:

```text
RevisionReference {
  revision:RevisionIdentity,
  title:LongText
}
```

No new schema family is introduced.

## 2. Target invariants

### Recognition

Every My Work row must be sufficiently human-recognizable for the actor to distinguish the work and choose its destination without opening the owner lens merely to identify the subject, while the queue itself remains a non-authoritative projection.

For governance rows the bounded recognition identity is:

```text
Document code
+ exact governed Revision ordinal
+ governed human title
+ governance subject kind
+ queue-entry time
+ governance_attempt_id for navigation
```

### Ordering

A paginated human work queue must have a deterministic ordering the user can understand without inventing unsupported urgency/priority semantics, and the frontend must not reorder cursor pages into a misleading partial order.

The fixed governance-work order is therefore:

```text
document.code ASC
governance_attempt_id ASC
```

The UUID remains a stable tie-break/navigation identity only; it is not presented as human priority.

Neither precision makes My Work a Governance Case summary or lifecycle authority.

## 3. Exact semantic source

`revision` never derives from current Document state and never performs a frontend join.

```text
subject_kind=submission
  WorkGovernanceItem.revision
  = exact immutable Submission governed Revision/title snapshot

subject_kind=obsolescence
  WorkGovernanceItem.revision
  = exact immutable target RevisionReference of the governed obsolescence subject
```

The projection is frozen/subject-aligned recognition data. B06 still loads `GovernanceCaseView` and remains the authority for exact case state, Steps, feedback, `allowed_actions`, governed content and decisions.

Ordering uses only fields already present in the admitted projection/query substrate. It introduces no queue priority fact.

## 4. Disclosure / authorization law

These precisions change **projection shape/order only**, not row admission.

```text
listGovernanceWork canonical admission/filtering
→ determines whether a WorkGovernanceItem exists for the actor
→ only then may its subject-aligned RevisionReference be returned
→ admitted rows are emitted in the canonical server order
```

The precision:

```text
does not add a row
does not grant access
does not permit content/history reads
does not reveal Step candidates/decisions/feedback
does not create SLA/due/priority semantics
does not replace Governance Case authorization
```

Destination navigation remains:

```text
WorkGovernanceItem.governance_attempt_id
→ /work/governance/:attempt_id
→ getGovernanceAttempt
→ current canonical disclosure/Authorization rechecked
```

A stale queue row is never case authority.

## 5. Global Maximum analysis — F1 recognition

### Root cause

The previous governance projection optimized for technical routing identity and omitted the already-owned governed-subject identity required for human recognition. That created an asymmetric UX:

```text
Authoring work   = recognizable before opening
Governance work  = code + generic kind + opaque attempt identity
```

### Rejected alternatives

```text
NO CONTRACT CHANGE
  rejected: preserves the recognition defect; local maximum inside current DTO

FRONTEND FAN-OUT TO getGovernanceAttempt PER ROW
  rejected: N+1 reads, partial loading/failure, B05→B06 coupling, duplicate cache/coherence work

RICH GOVERNANCE QUEUE SUMMARY
  rejected: Step/SLA/submitter/area/quorum/etc. lack a current B05 need and would pre-design B06

GENERIC UNIFIED WorkItem PLATFORM
  rejected: invents cross-lane state/order/pagination semantics and erases accepted Authoring/Governance distinctions
```

Selected structure:

```text
existing WorkGovernanceItem
+ existing RevisionReference
```

This is the smallest sustainable structure, not merely the smallest diff: it resolves the root cause at the owning projection boundary, removes frontend accidental complexity, preserves owner separation and leaves later evidence-backed enrichment additive.

## 6. Global Maximum analysis — F2 ordering

### Root cause

The previous order was:

```text
governance_attempt_id ASC
```

That is deterministic for the machine but opaque to the actor. Because the collection is cursor-paginated, the frontend cannot repair the ordering by sorting only loaded rows without presenting a false global order.

### Rejected alternatives

```text
KEEP governance_attempt_id ASC
  rejected: deterministic but human-arbitrary; preserves the UX defect

CLIENT-SIDE SORT
  rejected: cursor pages retain server ordering; local reorder creates a false partial/global order

created_at ASC/DESC
  rejected under current evidence: silently promotes FIFO/recency into queue-priority semantics

generic sort/filter control
  rejected: no Product need or API contract supports user-defined queue ordering
```

Selected structure:

```text
document.code ASC,
governance_attempt_id ASC
```

This is a neutral human-stable order, aligned with the existing authoring-work recognition posture, without asserting urgency, age or business priority.

## 7. Structural Inversion

If the current implementation exposed a rich governance work DTO, the sustainable target would still retain only subject identity needed for recognition/routing and remove Step/decision/content authority from the queue.

If the current implementation used a generic Work API, governance work would still need exact governed Revision/title identity rather than a per-row case fetch merely to identify it.

If the current implementation already sorted by recency, the sustainable choice would still require proof that recency is Product priority before preserving it; absent that proof, a neutral human-stable ordering remains preferable.

Therefore both decisions survive inversion of the accidental current DTO/order shape.

## 8. T6 / T8-E / T8-F binding

### T6 — Product/API/frontend journey meaning

My Work remains an actor-relevant projection and must support human recognition before handoff. Governance rows expose the exact governed `RevisionReference` and arrive in a fixed human-stable canonical order; no new journey, route, priority model or Product owner is created.

### T8-E — executable wire

Operation `listGovernanceWork` keeps the same method/path/pagination/problems. Its response item shape is refined as defined in §1, and its fixed canonical ordering is:

```text
document.code ASC,
governance_attempt_id ASC
```

The cursor binds that server order. The frontend does not reorder cursor pages.

### T8-F — frontend realization

B01/B05 may render returned:

```text
document.code
revision.revision.ordinal
revision.title
subject_kind
created_at
```

for recognition/navigation.

The frontend must not:

```text
call getGovernanceAttempt per row merely to enrich the queue
infer current Step/action state
normalize rows into a generic Work entity
re-sort cursor pages into a second ordering authority
infer urgency from created_at
```

## 9. Census / boundary impact

```text
new application operation       0
new stable SPA route            0
new Permission                  0
new semantic owner              0
new schema family               0
new persistence authority       0
new ETag domain                 0
new Idempotency-Key creation    0
new exact-byte resource         0
frontend read fan-out           0
new priority/SLA semantic       0
```

Current accepted census remains:

```text
application operations          86
stable SPA routes               11
PermissionCode values           16
Idempotency-Key creations       11
ETag read / mutation domains    13 / 13
exact-byte resources            4
```

## 10. Proof obligations

Later executable-contract/implementation proof must falsify at least:

```text
submission work row RevisionReference != exact governed Submission snapshot
obsolescence work row RevisionReference != exact governed target RevisionReference
row becomes visible solely because the new member exists
B05 performs per-row Governance Case enrichment
B05 uses revision/title as mutation/current-case authority
listGovernanceWork emits rows outside document.code ASC + governance_attempt_id ASC
cursor continuation is inconsistent with that canonical order
frontend reorders loaded cursor pages as if it owned global queue order
operation census changes from either precision
```

P8 must additionally verify that code + Revision + title + subject kind is sufficient for human queue recognition and that the fixed order is understandable without importing B06-only or unsupported urgency facts.

## 11. Reopen triggers

Reopen only the implicated read projection when material evidence proves one of:

```text
users still cannot reliably distinguish governance work items
users cannot use the neutral code order efficiently at real queue scale
active Step or another exact fact becomes demonstrably decision-critical before case entry
real SLA/due/priority Product semantics are promoted
an evidenced cross-lane unified-work journey requires different projection semantics
listGovernanceWork disclosure is not sufficient to disclose governed Revision/title identity safely
implementation evidence shows the subject-aligned reference/order cannot be projected without duplicating authority
```

Legacy Approval fields, visual preference, framework convenience and hypothetical future dashboards are not reopen triggers.
