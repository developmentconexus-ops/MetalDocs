# T11 — B05 My Work / Work Queues R1 — Method v2.2 candidate

> **Status:** CURRENT FP1 BLOCK / CANDIDATE / NOT LOCKED.  
> **Block:** B05 — My Work / Work Queues.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 LOCKED.  
> **Current bounded authority:** `../../decisions/my-work-governance-identification-read.md`.  
> **Implementation:** BLOCKED.

## 1. Current Product/architecture boundary

Stable route:

```text
/work
```

B01 already LOCKED the human model:

```text
Início
  default /work operational-home presentation

Minha Caixa
  Para aprovação -> /work governance presentation
  Em edição      -> /work authoring presentation
```

Lane/tab/query state is browser presentation only. B05 does not redesign the B01 default Home.

Current My Work authority:

```text
READ       listAuthoringWork + listGovernanceWork
WRITE      none
STATE      WorkAuthoringPage + WorkGovernancePage
AUTHORITY  projections route to owner lenses; My Work owns no lifecycle truth
```

Destinations:

```text
Authoring row
-> /documents/:document_id/work
-> B04 LOCKED
-> destination resolves current truth again

Governance row
-> /work/governance/:attempt_id
-> B06 NOT OPEN
-> B05 may expose only the transition boundary
```

## 2. Current read shapes and canonical order

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
  created_at:UtcInstant
}
```

Governance revision semantics:

```text
submission
  -> exact governed Submission Revision/title snapshot

obsolescence
  -> exact governed target RevisionReference
```

Canonical server order:

```text
listAuthoringWork
  document.code ASC, document_id ASC

listGovernanceWork
  document.code ASC, governance_attempt_id ASC
```

Both pages remain cursor-paginated recognition/navigation projections, never mutation/current-lifecycle authority.

## 3. Legacy evidence disposition

Useful ergonomics retained as Evidence:

```text
high-density work list
human code + title prominence
selected row / clear focus
keyboard-friendly traversal
clear no-work / error / stale-row recovery
owner-destination continuation
```

Rejected from current baseline:

```text
peer Approval owner/workspace
Approve / Return quick mutations
SLA / due / overdue / priority semantics
quorum progress
stage label absent from current authority
legacy overseer mode
filters absent from current list operations
localStorage view preference baseline
Template peer lifecycle
```

B05 remains read-only. Governance decisions belong B06.

## 4. Human needs

B05 must let the actor answer:

```text
What work is waiting for me?
Is it authoring work or governance work?
Which exact Document / Revision subject is it?
Where should I continue it?
Did a stale row change/disappear before destination entry?
Is there more work beyond the current cursor page?
```

B05 must not answer:

```text
exact governance Step / decision -> B06
exact DRAFT/source               -> B04
full lifecycle/history           -> B07
current official truth           -> B03
```

## 5. B05-F1 — governance row recognition — CLOSED / OPERATOR-RATIFIED

DevelopmentConexus Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ bounded read-projection correction
```

Selected precision:

```text
WorkGovernanceItem.revision: RevisionReference
```

Rejected alternatives:

```text
no contract change
  -> preserves recognition defect / local maximum

per-row getGovernanceAttempt fan-out
  -> N+1 + B05/B06 coupling + partial failure complexity

rich Governance queue DTO
  -> pre-designs B06 / unsupported fields

generic unified WorkItem
  -> invents cross-lane order/state/pagination semantics
```

## 6. B05-F2 — governance queue ordering — CLOSED / OPERATOR-RATIFIED

Target invariant:

> A cursor-paginated human work queue has one deterministic server-owned order the actor can understand, without inventing unsupported urgency/priority semantics and without frontend page reordering.

Selected:

```text
document.code ASC,
governance_attempt_id ASC
```

Rejected:

```text
governance_attempt_id ASC
  -> deterministic but human-arbitrary

client-side sort
  -> false global order over cursor pages

created_at ASC/DESC
  -> silently promotes FIFO/recency priority semantics

generic sort/filter DSL
  -> unsupported capability
```

F1/F2 binding authority:

```text
../../decisions/my-work-governance-identification-read.md
```

Current census remains 86 operations / 11 routes / 16 permissions.

## 7. P7 — focused queue per selected intent — OPERATOR-APPROVED

The operator approved **A — focused queue per selected Minha Caixa intent**.

Locked-for-P8 hypothesis:

```text
GLOBAL SHELL — inherited B01

Minha Caixa
  intent switch
    Para aprovação | Em edição

selected intent
  one full-width focused queue
  server cursor order preserved
  dense human-recognizable rows
  local row selection / keyboard navigation
  direct continuation to owner-lens boundary
  cursor load-more

material states
  empty
  load failure + retry
  stale destination + refresh
```

Governance row:

```text
Document code
Revision
Title
Submission | Obsolescência
Created at
Abrir caso -> B06 boundary
```

Authoring row:

```text
Document code
Revision
Title
DRAFT | SUBMITTED
Responsible owner
Updated at
Continuar/Abrir trabalho -> B04
```

Alternative B was rejected because a two-lane detailed view duplicates the B01 Home attention composition and pressures cross-lane comparison without a priority model.

Alternative C was rejected because legacy master/detail depended on approval context/actions that current read-only B05 intentionally does not own. Useful legacy density/selection/keyboard properties are retained inside A.

No content/detail preview pane is baseline.

## 8. P8 R1 — RENDERED / OPERATOR OPERATION+REVIEW

Functional low-fidelity R1 is browser-operable HTML/CSS/vanilla JavaScript with deterministic local fixtures.

It exercises:

```text
Para aprovação <-> Em edição presentation switch
server-order-preserving governance/authoring lists
dense row recognition
row selection
ArrowUp / ArrowDown navigation
Enter to open selected row
cursor-style Carregar mais
Governance handoff terminates at explicit B06 boundary
Authoring handoff terminates at B04 LOCKED boundary
stale destination -> row projection is not authority -> refresh
list failure -> retry
empty lane
B01N Quick Inbox reuse
responsive list reflow
```

Review-only controls force stale/error/empty states and are Evidence only, not Product UI.

### IA tension under explicit P8 test

Current accepted Authoring projection admits:

```text
state = DRAFT | SUBMITTED
```

while B01's global human label is:

```text
Em edição
```

P8 R1 deliberately contains visible SUBMITTED items under `Em edição`.

Disposition:

```text
DO NOT reopen B01 from theory.
If operator operation shows the label materially misleads, raise a smallest-scope B01 terminology finding.
If it remains understandable in context, preserve the existing LOCK.
```

## 9. Hard constraints preserved

```text
B05 read-only
no queue quick approve/reject
no due/SLA/urgency/priority inference
no total-count KPI absent total count
no generic filters/sort control
no B06 per-row enrichment
no B04 content preview
no frontend Authorization matrix
no merged cross-lane priority algorithm
server cursor order remains presentation order
```

## 10. Current gate

```text
B05 authority recovery                       COMPLETE
legacy queue ergonomics recovery             COMPLETE
B05-F1 governance identification precision   CLOSED / OPERATOR-RATIFIED
B05-F2 governance ordering                   CLOSED / OPERATOR-RATIFIED
P7 focused queue A                           APPROVED
P8 functional R1                             RENDERED / OPERATOR OPERATION+REVIEW
B05 LOCK                                     NO
```

Next:

```text
operator operates P8 R1
-> iterate only material B05 findings
-> operator-only B05 LOCK
-> P9 exact Screen Contract
-> P10 bounded pattern consolidation
```

B06+ remain NOT OPEN.
