# T11 — B05 My Work / Work Queues R1 — Method v2.2 candidate

> **Status:** CURRENT FP1 BLOCK / CANDIDATE / NOT LOCKED.  
> **Block:** B05 — My Work / Work Queues.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 LOCKED.  
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
  created_at:UtcInstant
}
```

Both are paginated projections. They are navigation/attention lenses only and never mutation/current-lifecycle authority.

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
  code / subject kind / created / open case
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

## 6. B05-F1 — Governance queue identification read symmetry — OPEN

### Trigger

`WorkAuthoringItem` is human-identifiable without another read:

```text
Document code
+ title
+ Revision
+ state
+ responsible owner
+ updated_at
```

Current `WorkGovernanceItem` exposes only:

```text
Document code
+ subject kind
+ governance_attempt_id
+ created_at
```

The opaque attempt id is navigation identity, not useful human recognition. The row lacks the governed subject title and Revision even though both are already frozen/available in the exact `GovernanceCaseView` subject:

```text
Submission subject
  revision + title

Obsolescence subject
  target_revision = revision + title
```

### Candidate smallest precision

```text
WorkGovernanceItem {
  governance_attempt_id:Uuid,
  subject_kind:GovernanceSubjectKind,
  document:DocumentReference,
  revision:RevisionIdentity,
  title:LongText,
  created_at:UtcInstant
}
```

Semantics:

```text
subject_kind=submission
  revision/title = immutable Submission governed subject snapshot

subject_kind=obsolescence
  revision/title = immutable target RevisionReference snapshot
```

Properties:

```text
new operation               0
new route                   0
new Permission              0
new semantic owner          0
new persistence authority   0
frontend join/read fan-out  0
```

The projection remains non-authoritative navigation/recognition data. B06 re-reads the exact Governance Case and remains authority for participation/actions.

Without this precision, a professional queue would force users to identify governance work primarily by code + generic `Submission/Obsolescence`, which is materially weaker than the already-available governed subject truth.

**Status:** OPEN / operator adjudication required before final B05 LOCK. P7 may compare composition using the candidate fields, clearly marked as B05-F1.

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

The remaining material composition question is how the detailed `Minha Caixa` presentation should use the two already-LOCKED intents:

```text
A. focused lane / full-width queue per selected sidebar intent
B. two-lane overview in the detailed queue page
C. legacy-inspired master/selection queue with a bounded non-authoritative summary
```

Any selected structure must preserve direct navigation to B04/B06 and avoid turning B05 into a decision/content owner.

## 9. Current gate

```text
B05 authority recovery                       COMPLETE
legacy queue ergonomics recovery             COMPLETE
B05-F1 governance identification precision   OPEN
P7 composition                               OPEN
P8 functional HTML                           NOT OPEN
B05 LOCK                                     NO
```

B06+ remain NOT OPEN.
