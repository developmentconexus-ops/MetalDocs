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
  created_at:UtcInstant
}
```

`WorkGovernanceItem.revision` is the operator-ratified bounded T11 precision in `../../decisions/my-work-governance-identification-read.md`:

```text
submission
  -> exact governed Submission Revision/title snapshot

obsolescence
  -> exact governed target RevisionReference
```

Both pages remain paginated projections. They are recognition/navigation lenses only and never mutation/current-lifecycle authority.

## 3. Legacy evidence disposition

Exact legacy `InboxStack` / `InboxApprovalCard` evidence showed useful ergonomics:

```text
queue rail
one selected card
previous / next navigation
keyboard left/right navigation
human code + title prominence
clear filtered/no-work empty states
open destination from selected work item
```

KEEP / ADAPT only when current authority supports the information.

Reject / do not restore:

```text
peer Approval owner/workspace
Approve / Return quick mutations from the queue
SLA / due date / overdue semantics
quorum progress
stage label unless current authority independently admits it
legacy overseer mode
legacy filters absent from current list operations
localStorage view preference baseline
timeline vs stack preference as durable Product state
Template peer lifecycle
```

B05 is read-only. Governance decisions belong B06.

## 4. Pre-v2.2 T11 evidence

The older T11 wireframe proposed simple Authoring/Governance lanes. Preserve only the semantic split:

```text
Authoring
  code / title / state / owner

Governance
  code / Revision/title / subject kind / created / open case
```

The old artifact is evidence, not current P8 authority.

## 5. Human needs

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

## 6. B05-F1 — Governance queue identification read symmetry — CLOSED / OPERATOR-RATIFIED

DevelopmentConexus Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ bounded read-projection correction
```

Target invariant:

> Every My Work row is sufficiently human-recognizable for the actor to distinguish the work and choose its destination without opening the owner lens merely to identify the subject, while the queue itself remains non-authoritative.

Selected precision:

```text
WorkGovernanceItem.revision: RevisionReference
```

Global Maximum result:

```text
NO CONTRACT CHANGE
  REJECTED — preserves recognition defect / local maximum

per-row getGovernanceAttempt fan-out
  REJECTED — N+1 + B05/B06 coupling + partial loading/failure complexity

rich Governance queue DTO
  REJECTED — pre-designs B06 / unsupported fields

generic unified WorkItem
  REJECTED — invents cross-lane order/state/pagination semantics

existing WorkGovernanceItem + existing RevisionReference
  SELECTED — fixes root cause at projection owner; additive seam; no duplicate authority
```

Structural Inversion survives: even if the current DTO were rich or generic, the governance queue would still need exact governed Revision/title identity but would still not need Steps/feedback/decision/content authority.

Binding authority:

```text
../../decisions/my-work-governance-identification-read.md
```

Impact:

```text
new operation               0
new route                   0
new Permission              0
new semantic owner          0
new schema family           0
new persistence authority   0
frontend join/read fan-out  0
```

Current census remains 86 operations / 11 routes / 16 permissions.

## 7. Bounded design constraints for P7

```text
B01 global shell/Minha Caixa IA is reused, not redesigned
B05 remains read-only
no queue quick approve/reject
no due dates / SLA / urgency inference
no total-count KPI absent an admitted total count
no generic filters/sort DSL
no B06 detail preview that requires getGovernanceAttempt inside B05
no B04 content preview inside queue
no frontend Authorization matrix
no merged cross-lane priority algorithm
```

Pagination remains owned independently by each current server page/cursor.

## 8. P7 design question

The material composition question is how the detailed `Minha Caixa` presentation should use the two already-LOCKED intents:

```text
A. focused lane / full-width queue per selected sidebar intent
B. two-lane overview in the detailed queue page
C. legacy-inspired master/selection queue with a bounded non-authoritative summary
```

Any selected structure must preserve direct navigation to B04/B06 and avoid turning B05 into a decision/content owner.

P7 comparison criteria:

```text
task completion
scanability / recognition
queue density
context preservation
pagination truth fit
accessibility / keyboard behavior
responsive viability
stale-row recovery
backend truth fit
cognitive load
```

## 9. Current gate

```text
B05 authority recovery                       COMPLETE
legacy queue ergonomics recovery             COMPLETE
B05-F1 governance identification precision   CLOSED / OPERATOR-RATIFIED
P7 composition                               NEXT / OPEN
P8 functional HTML                           NOT OPEN
B05 LOCK                                     NO
```

B06+ remain NOT OPEN.
