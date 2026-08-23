# T11 — B05 My Work / Work Queues R1 — Method v2.2 candidate

> **Status:** LOCKED / OPERATOR-RATIFIED / P8 COMPLETE.  
> **Block:** B05 — My Work / Work Queues.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 LOCKED.  
> **Current bounded authorities:** `../../decisions/my-work-governance-identification-read.md` + `../../decisions/governance-step-deadline.md`.  
> **Canonical P8:** `t11-b05-my-work-functional-wireframe.html`.  
> **Implementation:** BLOCKED.

## 1. Product / architecture boundary

Stable route remains:

```text
/work
```

B01 LOCKED mental model remains:

```text
Início
  default operational-home presentation

Minha Caixa
  Para aprovação
  Em edição
```

B05 remains read-only projection/navigation:

```text
READ       listAuthoringWork + listGovernanceWork
WRITE      none
AUTHORITY  owner lenses re-read exact current truth
```

Destinations:

```text
Authoring row  -> B04 LOCKED
Governance row -> B06 transition boundary only
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

Governance `revision` is exact governed-subject recognition. `due_at` is exact persisted deadline of the currently active Governance Step and is absent when that Step has no configured deadline.

## 3. Evidence disposition

Legacy/current comparative work preserves only useful proven queue properties:

```text
high-density list
human code + title prominence
selected row / keyboard traversal
empty / error / stale recovery
owner-destination continuation
real active-task due date triage
```

External evidence from Camunda, Flowable, ProcessMaker and ServiceNow informed F3/F4 but is not Product authority.

Not restored:

```text
peer Approval owner/workspace
quick approve/reject inside queue
manual High/Medium/Low priority
SLA extension / escalation
reassignment / overseer
generic task query/filter platform
business-day calendar
saved searches
```

## 4. Human needs

B05 lets the actor answer:

```text
What work is waiting for me?
Is it authoring or governance work?
Which exact Document / Revision is it?
For governance work, what is the active Step deadline when one exists?
Which governance work needs attention first?
Can I narrow the queue to the relevant deadline horizon?
Where should I continue it?
Did a stale row change/disappear?
Is there more work beyond the current cursor page?
```

Owner-lens details remain outside B05.

## 5. B05-F1 — recognition — CLOSED / OPERATOR-RATIFIED

Selected:

```text
WorkGovernanceItem.revision: RevisionReference
```

No per-row `getGovernanceAttempt` fan-out; no generic Work entity; no B06 summary moved into B05.

## 6. B05-F2 — neutral order — SUPERSEDED BY F4

F2 correctly rejected opaque UUID order and frontend re-sort. Its former code-first target was valid only before real deadline semantics existed.

F3 fired F2's explicit reopen trigger. F4 now owns final governance-work order.

Lasting law:

```text
frontend never re-sorts loaded cursor pages into a second global order
```

## 7. B05-F3 — Governance Step deadline — CLOSED / OPERATOR-RATIFIED

Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ BOUNDED TEMPORAL GOVERNANCE CORRECTION
```

Selected:

```text
GovernanceRouteStep.due_in_days?
→ immutable GovernanceAttemptStep.due_in_days_snapshot?
→ activation freezes activated_at
→ timed activation freezes persisted due_at
→ 1 configured day = 24 elapsed hours
→ no hidden default
→ overdue = attention truth only
→ no automatic lifecycle effect
```

Binding authority:

```text
../../decisions/governance-step-deadline.md
```

## 8. B05-F4 — due-aware queue — CLOSED / OPERATOR-RATIFIED

Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ BOUNDED DUE-AWARE QUEUE CORRECTION
```

Default server order:

```text
due_at ASC NULLS LAST,
document.code ASC,
governance_attempt_id ASC
```

Bounded first-page filter:

```text
deadline_filter? =
  overdue
  | next_24h
  | next_7d
  | no_deadline

omitted = all
```

Relative filters use one server-trusted first-page anchor carried/authenticated by the opaque cursor. Continuation pages reuse that anchor. A fresh first-page request receives a fresh anchor.

No baseline:

```text
user-selectable sort
arbitrary date range
today / this-week calendar semantics
manual priority
saved filters
total-count KPI
```

Binding authority:

```text
../../decisions/my-work-governance-identification-read.md
```

Current census remains 86 operations / 11 routes / 16 permissions.

## 9. P7 — focused queue per selected intent — OPERATOR-APPROVED

Locked structure:

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

Lane asymmetry is intentional: only governance has deadline triage.

## 10. P8 R1 disposition

R1 base structure/ergonomics was operator-approved. Operation then exposed the missing deadline-triage Product need, which correctly blocked the overall B05 LOCK and produced F3/F4.

No B01 terminology reopen was requested from the explicit `SUBMITTED`-under-`Em edição` test.

## 11. P8 R2 — APPROVED / COMPLETE

Canonical functional evidence:

```text
docs/work/current/t11-b05-my-work-functional-wireframe.html
```

R2 exercised:

```text
Para aprovação <-> Em edição
F4 due_at ASC NULLS LAST ordering
exact + relative due presentation
overdue presentation without lifecycle state invention
Todos / Atrasados / Próximas 24h / Próximos 7 dias / Sem prazo
no deadline controls in Em edição
cursor/load-more
fixed relative-filter cursor anchor
fixture server clock advance without cursor-anchor drift
fresh first-page anchor renewal
keyboard selection + Enter
stale destination + refresh
list failure + retry
B04 handoff
explicit B06 unopened boundary
responsive reflow
```

Review-only fixture controls are Evidence only, not Product UI.

## 12. Operator LOCK

Operator operation on 2026-08-23 approved the R2 experience without a surviving material finding.

Locked Product-experience properties:

```text
one focused queue per selected Minha Caixa intent
governance rows are human-recognizable before case entry
governance default order is deadline-first, server-owned
no-deadline work remains visible and explicitly filterable
four bounded deadline presets are sufficient Launch triage
relative filter cursor anchor remains stable during continuation
fresh first-page request renews the relative-time anchor
overdue is presentation/attention truth, not lifecycle state
Em edição has no artificial deadline controls
row open performs owner-lens handoff; queue owns no decision mutation
keyboard, stale, error, empty and responsive behavior remain part of the experience
```

No B01 terminology reopen is required from the tested `SUBMITTED` case.

## 13. Post-LOCK proof gate

```text
B05 authority recovery                       COMPLETE
legacy/external queue evidence               COMPLETE
B05-F1 recognition                           CLOSED / OPERATOR-RATIFIED
B05-F2 neutral order                         SUPERSEDED BY F4
B05-F3 Step deadline                         CLOSED / OPERATOR-RATIFIED
B05-F4 due-aware queue                       CLOSED / OPERATOR-RATIFIED
P7 focused queue A                           APPROVED
P8 R1 base structure                         OPERATOR-APPROVED
P8 R2 due-aware experience                   APPROVED / COMPLETE
B05 LOCK                                     LOCKED / OPERATOR-RATIFIED
P9 Screen Contract                           NEXT
P10 pattern consolidation                    AFTER P9
```

B06 remains unopened while P9/P10 prove the locked B05 scope.