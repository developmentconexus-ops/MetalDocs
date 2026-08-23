---
id: my-work-governance-identification-read
kind: authority
owner: architecture
summary: Bounded operator-ratified T11 precision making governance-work rows human-recognizable by projecting the exact governed RevisionReference without moving Governance Case authority into My Work.
---

# My Work governance identification read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified:** 2026-08-23  
> **Impacts:** T6 canonical journeys/read-model meaning, T8-E executable wire shape, T8-F frontend consumption.  
> **Implementation:** BLOCKED.

## 1. Decision outcome

DevelopmentConexus Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ bounded read-projection correction
```

The accepted My Work boundary remains sound:

```text
/work
→ actor-relevant authoring/governance projections
→ owner lens re-reads current/exact truth
```

The bounded defect was narrower: `WorkGovernanceItem` carried enough identity for routing but not enough governed-subject identity for reliable human recognition.

The current wire is therefore refined to:

```text
WorkGovernanceItem {
  governance_attempt_id:Uuid,
  subject_kind:GovernanceSubjectKind,
  document:DocumentReference,
  revision:RevisionReference,
  created_at:UtcInstant
}
```

Existing component reused unchanged:

```text
RevisionReference {
  revision:RevisionIdentity,
  title:LongText
}
```

No new schema family is introduced.

## 2. Target invariant

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

This does **not** make My Work a Governance Case summary or lifecycle authority.

## 3. Exact semantic source

`revision` never derives from current Document state and never performs a frontend join.

```text
subject_kind=submission
  WorkGovernanceItem.revision
  = exact immutable Submission governed Revision/title snapshot

subject_kind=obsolescence
  WorkGovernanceItem.revision
  = exact immutable target `RevisionReference` of the governed obsolescence subject
```

The projection is frozen/subject-aligned recognition data. B06 still loads `GovernanceCaseView` and remains the authority for exact case state, Steps, feedback, `allowed_actions`, governed content and decisions.

## 4. Disclosure / authorization law

This precision changes **projection shape only**, not row admission.

```text
listGovernanceWork canonical admission/filtering
→ determines whether a WorkGovernanceItem exists for the actor
→ only then may its subject-aligned RevisionReference be returned
```

The new member:

```text
does not add a row
does not grant access
does not permit content/history reads
does not reveal Step candidates/decisions/feedback
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

## 5. Global Maximum analysis

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

This is the smallest **sustainable** structure, not merely the smallest diff: it resolves the root cause at the owning projection boundary, removes frontend accidental complexity, preserves owner separation and leaves later evidence-backed enrichment additive.

## 6. Structural Inversion

If the current implementation had instead exposed a very rich governance work DTO, the sustainable target would still retain only the subject identity needed for recognition and routing while removing Step/decision/content authority from the queue.

If the current implementation had a generic Work API, the governance row would still need the exact governed Revision/title rather than a frontend fetch of the case merely to identify it.

Therefore this decision is independent of the accidental current DTO shape.

## 7. T6 / T8-E / T8-F binding

### T6 — Product/API/frontend journey meaning

My Work remains an actor-relevant projection and must support human recognition before handoff. Governance rows expose the exact governed `RevisionReference`; no new journey, route or Product owner is created.

### T8-E — executable wire

Operation `listGovernanceWork` keeps the same method/path/pagination/problems. Only the response item shape is refined as defined in §1.

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

The frontend must **not** call `getGovernanceAttempt` per row merely to enrich the queue, infer current Step/action state, or normalize the row into a generic Work entity.

## 8. Census / boundary impact

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

## 9. Proof obligations

Later executable-contract/implementation proof must falsify at least:

```text
submission work row RevisionReference != exact governed Submission snapshot
obsolescence work row RevisionReference != exact governed target RevisionReference
row becomes visible solely because the new member exists
B05 performs per-row Governance Case enrichment
B05 uses revision/title as mutation/current-case authority
operation census changes from this precision
```

P8 must additionally verify that code + Revision + title + subject kind is sufficient for human queue recognition without importing B06-only facts.

## 10. Reopen triggers

Reopen only the implicated read projection when material evidence proves one of:

```text
users still cannot reliably distinguish governance work items
active Step or another exact fact becomes demonstrably decision-critical before case entry
real SLA/due/priority Product semantics are promoted
an evidenced cross-lane unified-work journey requires different projection semantics
listGovernanceWork disclosure is not sufficient to disclose governed Revision/title identity safely
implementation evidence shows the subject-aligned reference cannot be projected without duplicating authority
```

Legacy Approval fields, visual preference, framework convenience and hypothetical future dashboards are not reopen triggers.
