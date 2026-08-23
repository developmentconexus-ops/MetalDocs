---
id: my-work-governance-identification-read
kind: authority
owner: architecture
summary: Operator-ratified My Work governance recognition precision; its former neutral queue ordering is explicitly reopened after governance Step deadline semantics became current.
---

# My Work governance read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified F1/F2:** 2026-08-23  
> **F2 bounded reopen:** 2026-08-23 by `governance-step-deadline.md`  
> **Impacts:** T6 canonical journeys/read-model meaning, T8-E executable wire shape/order, T8-F frontend consumption.  
> **Implementation:** BLOCKED.

## 1. Current disposition

F1 remains fully current:

```text
WorkGovernanceItem {
  governance_attempt_id:Uuid,
  subject_kind:GovernanceSubjectKind,
  document:DocumentReference,
  revision:RevisionReference,
  created_at:UtcInstant,
  due_at?:UtcInstant              // promoted by governance-step-deadline.md
}
```

`revision` gives exact governed-subject identity for human recognition without a per-row Governance Case fetch.

F2's former fixed order:

```text
document.code ASC,
governance_attempt_id ASC
```

was operator-ratified under the then-current invariant that no real urgency/deadline semantics existed. That premise is no longer true.

The F2 reopen trigger explicitly named:

```text
real SLA/due/priority Product semantics are promoted
```

That trigger fired when B05 operator use proved the need for approval deadlines and `governance-step-deadline.md` promoted real Step `due_at` truth.

Therefore:

```text
F1 recognition                  CURRENT / CLOSED
F2 neutral code ordering        REOPENED / NOT FINAL TARGET ORDER
B05-F4 due-aware queue behavior OPEN / owns final filter/order decision
```

The old F2 order remains provenance for why client-side sorting and opaque UUID ordering were rejected; it is not the final target order after F3.

## 2. Recognition invariant — unchanged

Every My Work row must be sufficiently human-recognizable for the actor to distinguish the work and choose its destination without opening the owner lens merely to identify the subject, while the queue itself remains a non-authoritative projection.

Governance recognition identity is:

```text
Document code
+ exact governed Revision ordinal
+ governed human title
+ governance subject kind
+ queue-entry time
+ optional current active-Step due_at
+ governance_attempt_id for navigation
```

Deadline truth does not make My Work a Governance Case summary or lifecycle authority.

## 3. Exact semantic sources

`revision` never derives from current Document state and never performs a frontend join.

```text
subject_kind=submission
  revision = exact immutable Submission governed Revision/title snapshot

subject_kind=obsolescence
  revision = exact immutable target RevisionReference
```

`due_at` is likewise not computed from current route configuration by the frontend:

```text
WorkGovernanceItem.due_at
= current admitted GovernanceAttempt's active Step persisted due_at
= null when the active Step has no configured deadline
```

Deadline lifecycle/provenance is owned by `governance-step-deadline.md`.

B06 will continue to load `GovernanceCaseView` as exact case authority.

## 4. Disclosure / authorization law

Read-shape enrichment never changes row admission.

```text
listGovernanceWork canonical admission/filtering
→ determines whether a row exists for the actor
→ only then may subject-aligned RevisionReference and active-Step due_at be returned
```

The row:

```text
does not grant access
does not reveal hidden Steps/decisions/feedback
does not own deadline mutation
does not replace Governance Case authorization
```

Destination navigation remains:

```text
governance_attempt_id
→ /work/governance/:attempt_id
→ getGovernanceAttempt
→ current disclosure/Authorization rechecked
```

A stale queue row is never case authority.

## 5. F1 Global Maximum — still confirmed

Rejected alternatives remain rejected:

```text
NO CONTRACT CHANGE
  preserves recognition defect

PER-ROW getGovernanceAttempt FAN-OUT
  N+1 reads + B05/B06 coupling + partial loading/failure

RICH GOVERNANCE CASE SUMMARY IN QUEUE
  pre-designs B06 and moves owner truth into My Work

GENERIC UNIFIED WorkItem PLATFORM
  invents cross-lane lifecycle/state semantics
```

Selected structure remains the existing governance projection plus existing semantic references and now the exact active-Step deadline required by a real consumer.

## 6. F2 provenance and reopen

F2 correctly established two lasting negative laws:

```text
opaque governance_attempt_id alone is not a human ordering model
frontend sorting of loaded cursor pages cannot pretend to own global collection order
```

Its previous `document.code ASC` selection was intentionally neutral because no priority fact existed. Deadline promotion invalidates only that premise/selection, not the two negative laws.

B05-F4 must now compare due-aware server ordering and bounded deadline filters while preserving cursor truth.

Until F4 is closed, no new final governance-work order may be inferred from this record.

## 7. Boundary impact

```text
new application operation       0
new stable SPA route            0
new Permission                  0
new semantic owner              0
frontend per-row case fan-out   0
frontend global re-sort         forbidden
manual priority state           absent
```

Current census remains 86 operations / 11 routes / 16 PermissionCode values.

## 8. Proof obligations

F1/F3/F4 combined later proof must falsify at least:

```text
row RevisionReference differs from exact governed subject
row due_at differs from exact active GovernanceAttemptStep due_at
row becomes visible solely because enrichment exists
B05 performs per-row Governance Case enrichment
frontend reorders cursor pages into a second global order
frontend creates manual urgency independent of due_at
```

Final queue ordering/filter proof belongs B05-F4.

## 9. Reopen triggers

F1 recognition reopens only if users still cannot reliably distinguish work or exact subject identity cannot be projected safely.

F4, once ratified, will own later queue-order/filter reopen triggers. `governance-step-deadline.md` separately owns temporal-governance reopen triggers such as business calendars, extensions or breach effects.
