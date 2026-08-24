---
id: my-work-governance-identification-read
kind: authority
owner: architecture
summary: Operator-ratified My Work governance projection authority for human recognition, active-Step deadline triage, bounded deadline filters, and one server-owned cursor order.
---

# My Work governance read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified F1/F2:** 2026-08-23  
> **F2 reopened by F3 / superseded by F4:** 2026-08-23  
> **Ratified F4:** 2026-08-23  
> **Impacts:** T6 canonical journeys/read-model meaning, T8-E executable wire shape/filter/order, T8-F frontend consumption.  
> **Implementation:** BLOCKED.

## 1. Current projection

```text
WorkGovernanceItem {
  governance_attempt_id:Uuid,
  subject_kind:GovernanceSubjectKind,
  document:DocumentReference,
  revision:RevisionReference,
  created_at:UtcInstant,
  due_at?:UtcInstant
}
```

Semantic source:

```text
submission
  revision = exact immutable Submission governed Revision/title snapshot

obsolescence
  revision = exact immutable target RevisionReference

due_at
  = exact persisted due_at of the current admitted GovernanceAttempt active Step
  = absent/null when that active Step has no configured deadline
```

`revision` and `due_at` are recognition/triage truth only. B06 remains Governance Case authority.

## 2. Decision outcome

DevelopmentConexus Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ BOUNDED DUE-AWARE QUEUE CORRECTION
```

F1 remains current. F3 promoted real Step deadline truth through `governance-step-deadline.md`, which deliberately fired F2's prior reopen trigger. F4 now supersedes only F2's former neutral `document.code`-first target order.

Lasting F2 negative laws remain binding:

```text
opaque governance_attempt_id alone is not a human ordering model
frontend sorting of loaded cursor pages cannot become a second global collection order
```

## 3. Target invariants

### Recognition

Every My Work governance row is sufficiently human-recognizable to distinguish the governed subject before owner-lens entry, without per-row Governance Case fan-out.

### Temporal triage

A governance work queue with real deadlines presents time-critical work first through one deterministic server-owned order. Deadline absence remains truthful and visible; it never becomes synthetic priority.

### Cursor truth

Deadline filtering and ordering remain globally coherent across cursor pages. Relative-time filters use one server-trusted anchor for the entire cursor traversal rather than reinterpreting the time window on every page.

My Work remains a read-only projection and owns no lifecycle, deadline mutation, Authorization or Governance Case decision.

## 4. Canonical server order — F4

`listGovernanceWork` default order is:

```text
due_at ASC NULLS LAST,
document.code ASC,
governance_attempt_id ASC
```

Meaning:

```text
oldest overdue deadline first
→ later overdue deadline
→ nearest future deadline
→ later future deadline
→ no-deadline work
```

No `now()` bucket participates in ordering; `due_at` is frozen Step truth. Therefore an item crossing from future to overdue does not require collection reordering merely because the clock advanced.

`document.code` and `governance_attempt_id` are deterministic tie-breaks only.

There is no Launch user-selectable sort control.

## 5. Bounded deadline filter

First-page `listGovernanceWork` may accept exactly one optional filter:

```text
deadline_filter = overdue | next_24h | next_7d | no_deadline
```

Omission means all admitted work.

For relative filters, let `A` be one server-trusted UTC instant captured for that first-page traversal:

```text
overdue
  due_at <= A

next_24h
  due_at > A
  AND due_at <= A + 24 elapsed hours

next_7d
  due_at > A
  AND due_at <= A + 168 elapsed hours

no_deadline
  due_at IS NULL
```

`next_7d` intentionally includes the `next_24h` subset. These are usability presets, not mutually exclusive lifecycle classes.

There is no baseline:

```text
arbitrary from/to date range
today / this-week calendar buckets
generic filter DSL
saved filter
manual High/Medium/Low priority
priority score
total-count KPI
```

## 6. Relative-time cursor anchoring

Current global cursor law remains:

```text
first page: operation filters + optional limit
next page: cursor + optional limit only
```

For `overdue | next_24h | next_7d`, the first page captures `A` from trusted server time. The opaque integrity-protected cursor binds at minimum:

```text
operationId
normalized deadline_filter
filter_anchor A
canonical F4 ordering
seek position
```

Continuation pages reuse the same `A`; they do not recalculate the relative filter window.

This freezes only filter interpretation, not authorization or collection membership:

```text
current disclosure/AuthZ rechecked every page
rows may legitimately disappear if no longer admitted
no frozen multi-page snapshot is created
no server cursor state is created
```

A fresh first-page request/refresh obtains a fresh anchor.

`no_deadline` and unfiltered traversal need no temporal anchor.

## 7. Disclosure / authority boundary

```text
listGovernanceWork canonical admission
→ determines whether a governance row exists for the actor
→ admitted row may disclose its exact subject recognition fields + active-Step due_at
→ server applies F4 filter/order
```

The row/filter/order:

```text
does not grant access
does not expose hidden Steps/decisions/feedback
does not mutate due_at
does not create manual priority
does not change lifecycle when overdue
does not replace getGovernanceAttempt authorization
```

Destination remains:

```text
governance_attempt_id
→ /work/governance/:attempt_id
→ getGovernanceAttempt
→ current truth/disclosure rechecked
```

## 8. Global Maximum analysis

Rejected:

```text
KEEP CODE-FIRST ORDER
  real deadline exists but queue ignores the exact fact that answers what is time-critical

CLIENT-SIDE DUE SORT
  creates false global order over cursor pages

MANUAL PRIORITY
  duplicates/competes with the proven deadline need

DYNAMIC overdue/today/tomorrow ORDER BUCKETS
  unnecessarily makes ordering depend on moving now/calendar semantics

GENERIC SORT/FILTER PLATFORM
  Camunda/Flowable/ProcessMaker breadth has no current MetalDocs consumer

TODAY / THIS WEEK PRESETS
  import timezone/calendar authority not present at Launch
```

Selected:

```text
persisted due_at
+ fixed due_at ASC NULLS LAST server order
+ four bounded deadline presets
+ one cursor-stable server time anchor for relative presets
```

Structural Inversion survives: even if the implementation were already code-first, priority-first or generic-query-driven, the current MetalDocs consumer would still require one deadline-owned queue order and bounded triage filters rather than duplicate priority or a generic task-query platform.

## 9. T6 / T8-E / T8-F binding

### T6

`Para aprovação` is actor work that may have current active-Step deadline truth. The Product default is deadline-first attention, with bounded deadline presets; `Em edição` has no artificial deadline symmetry.

### T8-E

Existing operation `listGovernanceWork` remains operation 55 and gains only:

```text
WorkGovernanceItem.due_at?
first-page optional deadline_filter
F4 canonical server order
cursor temporal-anchor binding for relative deadline_filter values
```

No operation 87 is created.

### T8-F

B05 may:

```text
show exact due_at
show relative labels derived for presentation
show OVERDUE as presentation when current time >= due_at
offer the four bounded filter presets
```

B05 must not:

```text
re-sort loaded cursor pages
invent/store manual urgency
construct arbitrary date ranges
interpret deadline breach as a lifecycle state
apply deadline filters to authoring work
```

## 10. Census / boundary impact

```text
new application operation       0
new stable SPA route            0
new Permission                  0
new semantic owner              0
new lifecycle state             0
new ETag domain                 0
new Idempotency-Key creation    0
new async worker                0
manual priority state           0
frontend per-row case fan-out   0
```

Current accepted census remains 86 operations / 11 routes / 16 PermissionCode values.

## 11. Proof obligations

Later executable-contract/implementation proof must falsify at least:

```text
row RevisionReference differs from exact governed subject
row due_at differs from exact active GovernanceAttemptStep due_at
unfiltered rows violate due_at ASC NULLS LAST + tie-break order
no-deadline item appears before timed work in default traversal
frontend reorders loaded cursor pages
relative filter continuation uses a different anchor than its first page
cursor allows repeated/changed first-page deadline_filter on continuation
fresh refresh incorrectly preserves an old filter anchor
passing due_at changes Step lifecycle automatically
frontend creates manual priority independent of due_at
```

P8 must verify that due-first ordering and the four filters materially improve triage without making the queue feel like a generic workflow product.

## 12. Reopen triggers

Reopen only the implicated F4 decision if evidence proves one of:

```text
real users need a different default queue order
a user-selectable sort becomes materially necessary
arbitrary deadline range is required beyond the four presets
business-calendar/timezone semantics are promoted
deadline-less work is operationally starved despite its explicit filter
queue scale/performance invalidates the sustainable indexed order/filter shape
parallel active governance introduces multiple candidate due_at values per work row
```

Deadline extension/escalation/breach effects belong to `governance-step-deadline.md` reopen triggers, not this queue projection authority.
