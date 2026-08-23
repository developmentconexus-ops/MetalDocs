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

B01 already LOCKED the human mental model and route presentation:

```text
Início
  default /work operational-home presentation

Minha Caixa
  Para aprovação -> /work governance presentation
  Em edição      -> /work authoring presentation
```

Any lane/tab/query key is browser presentation state only and never `/api/v1` semantics.

B05 does **not** redesign the B01 default Home. It owns the detailed work-queue presentations reached from `Minha Caixa`.

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
-> B05 may expose transition semantics only; it does not design B06
```

## 2. Current read shapes and ordering

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

Both pages remain cursor-paginated recognition/navigation projections. They are never mutation/current-lifecycle authority.

## 3. Legacy evidence disposition

Useful ergonomics recovered from the exact legacy `InboxStack` / `InboxApprovalCard`:

```text
high-density work list
human code + title prominence
selected row / clear focus
previous / next keyboard-friendly traversal where useful
clear no-work and stale-row states
open the owner destination from the selected work item
```

Reject / do not restore:

```text
peer Approval owner/workspace
Approve / Return quick mutations from the queue
SLA / due date / overdue semantics
quorum progress
stage label absent from current authority
legacy overseer mode
filters absent from current list operations
localStorage view preference baseline
Template peer lifecycle
```

B05 is read-only. Governance decisions belong B06.

## 4. Human needs

B05 must let the actor answer:

```text
What work is waiting for me?
Is it work I am authoring or governance work I need to act on?
Which exact controlled Document/work subject is it?
Where should I continue it?
Did a stale queue row disappear/change before I entered the owner lens?
Do I have more work beyond the current page?
```

B05 must not answer owner-lens questions such as:

```text
What exact governance Step may I decide?       -> B06
What is the exact DRAFT/source?                -> B04
What is the full lifecycle/history?            -> B07
What is currently official?                    -> B03
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

Global Maximum comparison:

```text
NO CONTRACT CHANGE
  REJECTED — preserves recognition defect / local maximum

per-row getGovernanceAttempt fan-out
  REJECTED — N+1 + B05/B06 coupling + partial loading/failure complexity

rich Governance queue DTO
  REJECTED — pre-designs B06 / unsupported fields

generic unified WorkItem
  REJECTED — invents cross-lane state/order/pagination semantics

existing WorkGovernanceItem + existing RevisionReference
  SELECTED
```

## 6. B05-F2 — governance queue ordering — CLOSED / OPERATOR-RATIFIED

Target invariant:

> A cursor-paginated human work queue has one deterministic server-owned order the actor can understand, without inventing unsupported urgency/priority semantics and without frontend page reordering.

Selected fixed order:

```text
document.code ASC,
governance_attempt_id ASC
```

Rejected:

```text
governance_attempt_id ASC
  deterministic but human-arbitrary

client-side sort
  false global order over cursor pages

created_at ASC/DESC
  would silently promote FIFO/recency priority semantics

generic sort/filter DSL
  unsupported Product capability
```

Impact remains zero for operations, fields, routes, Permissions, owners and priority/SLA semantics. The cursor simply binds the revised canonical server order.

Binding F1/F2 authority:

```text
../../decisions/my-work-governance-identification-read.md
```

Current census remains 86 operations / 11 routes / 16 permissions.

## 7. Bounded design constraints for P7

```text
B01 global shell/Minha Caixa IA is reused, not redesigned
B05 remains read-only
no queue quick approve/reject
no due dates / SLA / urgency inference
no total-count KPI absent an admitted total count
no generic filters/sort control
no B06 detail preview requiring getGovernanceAttempt inside B05
no B04 content preview inside queue
no frontend Authorization matrix
no merged cross-lane priority algorithm
server cursor order remains presentation order
```

Potential IA tension to **test**, not silently reopen:

```text
B01 label "Em edição"
vs
WorkAuthoringItem.state = DRAFT | SUBMITTED
```

B05 P8 must render an explicit SUBMITTED item. Reopen B01 terminology only if operation evidence shows the locked label materially misleads users.

## 8. P7 composition comparison

### A — Focused queue per selected intent — LEADING HYPOTHESIS

```text
Minha Caixa / Para aprovação
→ one full-width governance queue
→ rows optimized for scanability
→ row click/open -> B06 boundary

Minha Caixa / Em edição
→ one full-width authoring queue
→ rows optimized for scanability
→ row click/continue -> B04
```

Strengths:

```text
matches locked sidebar intent directly
maximizes scan density
fits independent cursor pagination naturally
no duplicate Home overview
no invented cross-lane priority
responsive transformation is straightforward
stale row recovery stays local to the selected queue
```

### B — Two lanes together — REJECTED AS LEADING BASELINE

Would reproduce the B01 Home attention composition rather than add a distinct detailed queue experience. It also creates cross-lane comparison pressure without an admitted cross-lane order/priority model.

### C — Legacy-inspired master/detail — REJECTED AS LEADING BASELINE

The legacy stack made sense because the selected card also carried approval context/actions. Current B05 is read-only and intentionally lacks Step/content/decision truth. A large detail pane would either repeat the selected row or pressure B05 to import B04/B06 facts.

The useful legacy properties survive inside A as dense selectable rows, keyboard/focus behavior and clear continuation actions; the old master/detail shell does not.

## 9. Leading P7 structure

```text
GLOBAL SHELL — inherited B01

page header
  Minha Caixa
  concise intent explanation

intent switch / route-presentation state
  Para aprovação | Em edição

selected intent
  focused full-width queue
  server order preserved
  dense rows
  row identity + bounded metadata
  direct owner-lens continuation
  load-more cursor continuation

empty / load failure / stale destination
  explicit bounded state
```

Governance row candidate fields:

```text
Document code
Revision
Title
Submission | Obsolescência
Created at
Abrir caso
```

Authoring row candidate fields:

```text
Document code
Revision
Title
DRAFT | SUBMITTED
Responsible owner
Updated at
Continuar trabalho
```

No preview/detail pane is baseline.

## 10. Current gate

```text
B05 authority recovery                       COMPLETE
legacy queue ergonomics recovery             COMPLETE
B05-F1 governance identification precision   CLOSED / OPERATOR-RATIFIED
B05-F2 governance ordering                   CLOSED / OPERATOR-RATIFIED
P7 composition                               LEADING A / OPERATOR REVIEW
P8 functional HTML                           NOT OPEN
B05 LOCK                                     NO
```

Next:

```text
operator adjudicates P7 A
→ build one functional low-fi P8
→ operator operates / iterates / LOCKs
→ P9 exact Screen Contract
→ P10 bounded pattern consolidation
```

B06+ remain NOT OPEN.
