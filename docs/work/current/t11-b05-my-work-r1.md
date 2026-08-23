# T11 — B05 My Work / Work Queues R1 — Method v2.2 candidate

> **Status:** CURRENT FP1 BLOCK / CANDIDATE / NOT LOCKED.  
> **Block:** B05 — My Work / Work Queues.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 LOCKED.  
> **Current bounded authorities:** `../../decisions/my-work-governance-identification-read.md` + `../../decisions/governance-step-deadline.md`.  
> **Implementation:** BLOCKED.

## 1. Current Product/architecture boundary

Stable route:

```text
/work
```

B01 LOCKED mental model remains:

```text
Início
  default /work operational-home presentation

Minha Caixa
  Para aprovação -> /work governance presentation
  Em edição      -> /work authoring presentation
```

B05 is read-only projection/navigation only:

```text
READ       listAuthoringWork + listGovernanceWork
WRITE      none
AUTHORITY  owner lenses re-read exact current truth
```

Destinations remain:

```text
Authoring row  -> B04 LOCKED
Governance row -> B06 NOT OPEN transition boundary only
```

## 2. Current read shapes

```text
WorkAuthoringItem {
  document:DocumentReference,
  revision:RevisionIdentity,
  title:LongText,
  state:OpenRevisionState,
  responsible_owner:UserReference,
  updated_at:UtcInstant
}

WorkGovernanceItem {
  governance_attempt_id:Uuid,
  subject_kind:GovernanceSubjectKind,
  document:DocumentReference,
  revision:RevisionReference,
  created_at:UtcInstant,
  due_at?:UtcInstant
}
```

Governance `revision` is exact governed-subject recognition. Governance `due_at` is exact persisted deadline of the currently active Step and may be absent when that Step has no configured deadline.

## 3. Legacy and external evidence disposition

Useful queue ergonomics retained:

```text
high-density work list
human code + title prominence
selected row / keyboard traversal
empty / error / stale recovery
owner-destination continuation
```

Legacy SLA machinery is not restored wholesale.

External evidence from Camunda, Flowable, ProcessMaker and ServiceNow confirmed a narrower common pattern:

```text
active human task/step owns a due date
sequential next-step deadline begins only when that step activates
queues may filter/order by due date
manual priority / escalation / business calendar are separate optional capabilities
```

MetalDocs adopts only the proven temporal core in `governance-step-deadline.md`.

## 4. Human needs

B05 must let the actor answer:

```text
What work is waiting for me?
Is it authoring or governance work?
Which exact Document / Revision is it?
For governance work, what is the active Step deadline when one exists?
Which governance work needs attention first?
Where should I continue it?
Did a stale row change/disappear?
Is there more work beyond the current cursor page?
```

It must not answer owner-lens detail questions belonging B04/B06/B07/B03.

## 5. B05-F1 — governance row recognition — CLOSED / OPERATOR-RATIFIED

Selected:

```text
WorkGovernanceItem.revision: RevisionReference
```

No per-row `getGovernanceAttempt` fan-out; no generic Work entity; no B06 summary moved into B05.

Binding authority:

```text
../../decisions/my-work-governance-identification-read.md
```

## 6. B05-F2 — neutral governance ordering — REOPENED BY F3

Former accepted order:

```text
document.code ASC,
governance_attempt_id ASC
```

was correct under the then-current assumption that no due/priority Product truth existed.

Operator P8 use promoted a real approval-deadline need. The exact F2 reopen trigger therefore fired.

Lasting F2 laws remain:

```text
opaque UUID ordering is not a human ordering model
frontend must never re-sort loaded cursor pages into a fake global order
```

Final server order is now owned by B05-F4.

## 7. B05-F3 — governance Step deadline — CLOSED / OPERATOR-RATIFIED

Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ BOUNDED TEMPORAL GOVERNANCE CORRECTION
```

Selected Product semantics:

```text
GovernanceRouteStep.due_in_days?
→ immutable GovernanceAttemptStep.due_in_days_snapshot?
→ Step activation freezes activated_at
→ timed Step activation freezes persisted due_at
→ 1 configured day = 24 elapsed hours
→ no hidden default
→ overdue is attention truth only
→ no automatic lifecycle effect
```

Rejected from Launch baseline:

```text
manual High/Medium/Low priority
SLA extension
escalation worker
reassignment/overseer
business-day/holiday calendar
breach notification
automatic governance consequence
```

Binding authority:

```text
../../decisions/governance-step-deadline.md
```

Current census remains 86 operations / 11 routes / 16 permissions.

## 8. P7 — focused queue per selected intent — OPERATOR-APPROVED

Selected structure remains:

```text
Minha Caixa
  [Para aprovação] [Em edição]

selected intent
  one full-width focused queue
  dense recognizable rows
  keyboard selection
  owner-lens continuation
  cursor load-more
  empty / error / stale recovery
```

Two-lane detailed view remains rejected as redundant with B01 Home. Legacy master/detail remains rejected because B05 owns no legitimate case-detail/decision surface.

## 9. P8 R1 — BASE STRUCTURE OPERATOR-APPROVED / MATERIAL POST-OPERATION FINDING

Operator approved the P8 R1 structure/ergonomics and then identified a material usability requirement:

```text
Para aprovação has real approval deadlines
→ actor needs filters / due-aware prioritization
Em edição has no equivalent deadline requirement
```

Disposition:

```text
P8 R1 structural direction        APPROVED
B05 overall LOCK                  NO
B05-F3 temporal domain semantics  CLOSED / OPERATOR-RATIFIED
B05-F4 queue filter/order UX      OPEN / NEXT DECISION
P8 R2                             BLOCKED ON F4
```

The lane asymmetry is intentional: governance may have deadline controls while authoring does not.

## 10. B05-F4 — due-aware governance queue behavior — OPEN / NEXT

F4 must apply the DevelopmentConexus Method to decide only:

```text
what deadline information is visible in each governance row
what the default server-owned order is
which bounded deadline filters exist
whether user-selectable ordering is justified
how null/no-deadline items interleave
how cursor continuation binds filters/order
what presentation labels are derived from due_at
```

Constraints already fixed:

```text
server owns global cursor order
frontend never re-sorts pages
manual business priority state absent
no generic filter/sort DSL
no SLA/escalation semantics hidden in the queue
no B06 per-row enrichment
```

External systems are evidence only; MetalDocs will not copy generic task-platform breadth by default.

## 11. IA tension still under observation

Current Authoring projection admits:

```text
DRAFT | SUBMITTED
```

under the B01 LOCKED label:

```text
Em edição
```

P8 R1 did not produce an operator-requested terminology reopen. Preserve the current B01 label unless later material evidence changes that.

## 12. Current gate

```text
B05 authority recovery                       COMPLETE
legacy/external queue evidence               COMPLETE
B05-F1 recognition                           CLOSED / OPERATOR-RATIFIED
B05-F2 neutral order                         REOPENED BY F3
B05-F3 Step deadline                         CLOSED / OPERATOR-RATIFIED
P7 focused queue A                           APPROVED
P8 R1 base structure                         OPERATOR-APPROVED
B05-F4 due-aware queue behavior              OPEN / NEXT DECISION
P8 R2                                        NOT OPEN
B05 LOCK                                     NO
```

B06+ remain NOT OPEN.
