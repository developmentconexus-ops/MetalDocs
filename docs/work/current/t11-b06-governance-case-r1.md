# T11 — B06 Governance Case R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / ENTRY CLOSED / P7 CANDIDATE.  
> **Block:** B06 — Governance Case.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 LOCKED.  
> **Current bounded authority:** `../../decisions/governance-case-step-deadline-read.md` + `../../decisions/governance-step-deadline.md`.  
> **Implementation:** BLOCKED.  
> **P8:** NOT OPEN.

## 1. Entry recovery outcome

B06 entry recovery covered current Product/T2/T3/T6/T8-E/T8-F authority plus useful legacy governance ergonomics.

Entry result:

```text
Governance Case semantic owner     Controlled Documents lens
stable route                       /work/governance/:attempt_id
primary read                       getGovernanceAttempt
write family                       feedback + Step Decision only
frontend Authorization authority   none
case content                       exact immutable governed subject
legacy semantics inherited         none by sunk cost
material entry findings            1
material entry findings remaining  0
```

The single entry finding was B06-F1: current Step deadline truth existed in the domain but not in the Governance Case Step projection. It is now CLOSED / OPERATOR-RATIFIED through:

```text
../../decisions/governance-case-step-deadline-read.md
```

No Product operation, route, Permission, lifecycle state or semantic owner was added.

## 2. Human jobs

### J1 — understand exactly what is being governed

```text
When governance work opens,
I need to identify the exact Document / Revision / governed subject and inspect its immutable content,
so that I never decide against a different DRAFT, current revision or stale queue summary.
```

### J2 — understand where the case is now

```text
When I inspect the case,
I need to understand the ordered Steps, the active Step, prior decisions/feedback and the real deadline when one exists,
so that I can act with the right procedural context without learning a generic workflow engine.
```

### J3 — decide safely

```text
When I am currently allowed to act,
I need to ACCEPT or RETURN_FOR_CHANGES deliberately,
so that one immutable Step Decision is recorded against the exact case.
```

RETURN_FOR_CHANGES requires a reason. ACCEPT does not invent a reason requirement.

### J4 — communicate without confusing communication with decision

```text
When I need to add context or ask/supply information,
I need to add governance feedback separately from the Step Decision,
so that discussion does not become an implicit verdict.
```

### J5 — recover when current truth changed

```text
When another participant or server transition wins first,
I need the screen to replace my stale assumption with authoritative case truth and explain what happened,
so that I do not retry a conflicting decision or believe an uncommitted action succeeded.
```

### J6 — inspect a case with no current action

```text
When the case is disclosed but no action is currently offered,
I need to understand the governed subject and current/proven process state,
so that absence of controls is not mistaken for a broken UI or frontend permission decision.
```

## 3. Exact authority boundary

### Primary case truth

```text
67 getGovernanceAttempt
   -> GovernanceCaseView
```

`GovernanceCaseView` supplies:

```text
governance_attempt_id
state
subject = SubmissionGovernanceSubject | ObsolescenceGovernanceSubject
ordered steps
embedded first feedback page
allowed_actions
```

B06-F1 refines `steps` with exact persisted `due_at?` on activated/decided timed Steps.

### Supporting reads

```text
68 listGovernanceFeedback
   -> cursor continuation beyond embedded first20

70 getGovernanceStepDecision
   -> exact singleton Decision read when targeted reconciliation needs it

exact governed subject/source
   -> only through already-admitted owner reads
```

Submission source identity:

```text
GovernanceCaseView.subject.submission_id
-> exact immutable Submission source read
-> never current WorkingContent
```

Obsolescence content identity:

```text
GovernanceCaseView.subject.target_revision
-> exact target Revision identity must remain the rendered identity
-> supporting official/release reads may resolve bytes only when they match that exact target
-> never substitute a newer/current different Revision merely because it is easier to fetch
```

If a supporting read cannot lawfully resolve the exact target, B06 must fail closed rather than render a different Revision. A proven reachable case with no admitted exact-content resolution is an upstream authority finding; P7/P8 may not repair it with History leakage or client inference.

### Writes

```text
69 createGovernanceFeedback
   body.message
   one Idempotency-Key per logical feedback creation

71 recordGovernanceStepDecision
   ACCEPT
   RETURN_FOR_CHANGES + required reason
   natural singleton PUT; no Idempotency-Key row
```

### Local frontend state

```text
feedback draft
decision selection / return reason until accepted
confirmation disclosure
viewer UI state
feedback pagination/navigation state
```

No durable business truth is created in the browser.

## 4. Action and Authorization law

Current closed action vocabulary:

```text
accept
return_for_changes
add_feedback
```

`allowed_actions` is server-derived guidance only.

B06 may use it to suppress/offer affordances, but:

```text
hidden control != denied authority proof
shown control != granted authority proof
command always rechecks current server truth
```

No B06 control exists for:

```text
publish / release
withdraw Submission
cancel Revision
withdraw obsolescence request
cancel GovernanceAttempt
reassign / delegate
quorum / overseer
extend deadline / SLA
escalate
manual priority
edit governed content
```

Those are either owned by another lens/current command or absent from Launch.

## 5. Deadline behavior in B06

Current bounded authority:

```text
pending Step
  no due_at

active timed Step
  exact persisted due_at

active untimed Step
  no due_at

decided timed Step
  exact preserved due_at
```

Presentation may derive:

```text
absolute deadline
relative wording
overdue attention when now >= due_at
```

Presentation must not derive:

```text
new lifecycle state
SLA success/failure
permission
priority score
escalation
```

B06 reloads its own Step deadline from `getGovernanceAttempt`; B05 navigation state is never authority.

## 6. Material failure/recovery branches

### Case 404 / disclosure loss

```text
getGovernanceAttempt -> 404
-> do not reveal whether the attempt exists outside disclosure
-> terminal case-not-available presentation
-> offer safe return to Minha Caixa
```

### Case 403

```text
currently visible request/action but missing authority
-> explain action/context is not permitted now
-> server remains authority
```

### Decision conflict

```text
recordGovernanceStepDecision
-> 409 state.governance_step_already_decided
-> refetch current case / exact Decision
-> show the winning authoritative Decision
-> do not silently retry a different outcome
```

If the attempted logical PUT had an ambiguous transport outcome, an exact same outcome/reason retry remains compatible with singleton semantics; a changed outcome/reason is a new conflicting intent and must not reuse the old assumption.

### Feedback ambiguous outcome

```text
createGovernanceFeedback
-> preserve the same logical Idempotency-Key for retry
-> do not create a second message merely because the first response was lost
```

### Exact-content failure

```text
integrity/dependency failure
-> no successful partial exact-content claim
-> case metadata/process may remain understandable
-> decision surface must not imply the unavailable bytes were successfully reviewed
```

The exact interaction treatment is P8 material.

## 7. P6 — bounded reference study

Reference study is Evidence only; MetalDocs authority remains §§1–6.

### Current legacy archive

Useful observations:

```text
ordered process timeline supports orientation
real due date is useful when present
persistent/visible decision area reduces action hunting
explicit retry/error state is useful
reason entry for change request is materially different from generic comment
```

Rejected legacy semantics:

```text
ExtendSlaDialog
CancelInstanceDialog
fast-forward / "Aprovar já"
password/eSignature ceremony
legacy passed/failed/cancelled Step ontology
peer Approval product/workspace
```

### Camunda Tasklist — reference observation

Observed pattern:

```text
task queue -> selected task detail
context/due date kept near task detail
process view helps explain where a task sits in a broader flow
```

Mismatch:

```text
Camunda is generic task/workflow infrastructure
priority, follow-up, assignment and process tooling exceed MetalDocs authority
```

Reference:

```text
https://docs.camunda.io/docs/components/tasklist/userguide/using-tasklist/
```

### Veeva Vault document workflow — reference observation

Observed pattern:

```text
assigned document workflow task -> document workflow viewer
participant reviews bound document content before verdict
verdict and comments are distinct concepts
workflow binds/reviews a specific document version rather than silently following later content
```

Mismatch:

```text
multi-document envelopes, task acceptance, eSignature and broader workflow configuration are not MetalDocs Launch requirements
```

References:

```text
https://platform.veevavault.help/en/lr/50506
https://platform.veevavault.help/en/lr/50493
```

### P6 stop condition

Additional references stopped changing the decision space. Current evidence consistently favors keeping exact governed content and decision context together while avoiding generic workflow machinery.

P6 = COMPLETE.

## 8. P7 — credible layout hypotheses

The real ambiguity is how much space to give exact content versus case/process context while preserving deliberate decision ergonomics.

### H1 — Content-first Governance Workspace — LEADING CANDIDATE

```text
minimal Governance Case header
  return to Minha Caixa
  Document code + exact Revision/title
  subject kind
  case state
  active Step deadline when present

main workspace
  dominant content region
    Exact Read-Only Content Viewer Shell
    exact immutable Submission / exact obsolescence target

  B06-local governance rail
    A. what is being decided
    B. ordered Steps / active Step / prior Decisions
    C. governance feedback
    D. deliberate decision zone
```

Decision zone behavior:

```text
ACCEPT
  -> deliberate confirmation
  -> record singleton Decision

RETURN_FOR_CHANGES
  -> required reason form
  -> deliberate confirmation
  -> record singleton Decision

ADD FEEDBACK
  -> separate feedback composer
  -> never satisfies RETURN reason
```

Why it leads:

```text
task completion        strongest: content stays present while deciding
recognition            exact Revision identity can remain continuously visible
context preservation   strongest: no viewer round-trip before verdict
backend truth fit      direct mapping to exact governed source + case read
error recovery         decision conflict can reconcile beside same case context
accessibility          two semantic regions can reflow linearly
responsive viability   rail can reflow below content without inventing another route
product coherence      reuses exact read-only viewer semantics without copying B04 rail authority
```

Important constraint:

> The B06 governance rail is a new block-local semantic composition. B04's `Work operational rail` remains B04-local and is not generalized merely because both happen to sit beside content.

### H2 — Dossier-first Case + separate full viewer — REJECTED AS LEADING

```text
case summary / timeline / feedback / actions dominant
small content preview
-> open separate full viewer to inspect exact bytes
-> return to case to decide
```

Advantages:

```text
strong process scanability
close to B03 ficha mental model
```

Weaknesses:

```text
forces context switching at the exact moment content judgment matters
makes a thumbnail/preview too influential for a governed decision
increases risk of deciding from metadata/process rather than exact content
```

Disposition:

```text
REJECTED AS LEADING — task/content fit weaker than H1
```

### H3 — Three-column workflow cockpit — REJECTED

```text
left   Step/process navigator
center exact content
right  feedback/actions
```

Advantages:

```text
maximum simultaneous information on very wide desktop
```

Weaknesses:

```text
workflow topology becomes visually co-equal with the governed document
high horizontal density
poor narrow-screen transformation
encourages generic workflow-platform affordances
legacy cockpit resemblance creates sunk-cost risk
```

Disposition:

```text
REJECTED — density/product-shape mismatch
```

## 9. P7 leading hypothesis — material regions

H1 carries these regions into functional P8 if operator-approved:

```text
R1  Case orientation header
R2  Exact governed-content viewer
R3  Subject summary
R4  Active-Step deadline context
R5  Ordered Step progression / prior Decisions
R6  Feedback timeline + cursor continuation
R7  Add-feedback composer when hinted
R8  Decision zone when hinted
R9  ACCEPT confirmation
R10 RETURN_FOR_CHANGES reason + confirmation
R11 authoritative conflict/reconciliation state
R12 unavailable/non-disclosable case state
R13 exact-content unavailable/integrity state
```

No B07 History, B08 full Notifications Inbox or generic Approval administration is included.

## 10. P7 required truth check

| Need | Identity/source | Status |
|---|---|---|
| Case identity/state | `GovernanceCaseView.governance_attempt_id/state` | PRESENT-IN-AUTHORITY |
| Exact Document identity | `subject.document` | PRESENT-IN-AUTHORITY |
| Exact governed Revision/title | Submission `revision + title`; obsolescence `target_revision` | PRESENT-IN-AUTHORITY |
| Subject actor/time | submitter/submitted_at or initiator/requested_at | PRESENT-IN-AUTHORITY |
| Obsolescence reason | `subject.reason` | PRESENT-IN-AUTHORITY |
| Exact Submission bytes | `submission_id` -> admitted exact Submission source read | PRESENT-IN-AUTHORITY |
| Exact obsolescence target bytes | exact target identity -> admitted matching official/release content reads; never substitute another Revision | PRESENT-IN-AUTHORITY / FAIL-CLOSED MATCH REQUIRED |
| Step labels/order/state | `GovernanceCaseView.steps` route order | PRESENT-IN-AUTHORITY |
| Step deadline | B06-F1 `GovernanceStepView.due_at?` | PRESENT-IN-AUTHORITY |
| Prior Step Decision | decided Step embeds Decision; op70 supports targeted reconciliation | PRESENT-IN-AUTHORITY |
| Feedback first page | embedded `feedback` first20 | PRESENT-IN-AUTHORITY |
| Feedback continuation | op68 cursor | PRESENT-IN-AUTHORITY |
| Add feedback | `allowed_actions` hint + op69 | PRESENT-IN-AUTHORITY |
| ACCEPT | `allowed_actions` hint + op71 | PRESENT-IN-AUTHORITY |
| RETURN_FOR_CHANGES | hint + op71 + required reason | PRESENT-IN-AUTHORITY |
| Decision immutability/conflict | singleton + 409 current contract | PRESENT-IN-AUTHORITY |
| Pagination scale | feedback cursor only; no total-count requirement | PRESENT-IN-AUTHORITY |
| Case sorting/filtering | none inside one exact case | PRESENT-IN-AUTHORITY |
| Frontend AuthZ matrix | forbidden | PRESENT-IN-AUTHORITY — MUST NOT EXIST |

Material P7 authority findings: **0**.

## 11. Responsive / accessibility feasibility for H1

Desktop candidate:

```text
content remains dominant
rail remains independently understandable
Decision zone remains reachable without covering governed content
```

Narrow candidate:

```text
orientation header
-> subject/deadline summary
-> exact content
-> Steps
-> feedback
-> decision controls
```

A persistent compact action affordance may be explored in P8 only if it does not hide content, bypass the required reason or create accidental destructive activation.

Structural accessibility obligations:

```text
Step progression must not rely on color alone
active/current Step announced semantically
absolute deadline remains available even when relative label exists
overdue not color-only
viewer has explicit semantic identity outside rendered bytes
feedback and Decision are separately labeled regions
RETURN reason has explicit required semantics
confirmation focus/escape/return path must be testable
409 reconciliation announces authoritative change
```

## 12. P7 exit gate

Current status:

```text
H1 Content-first Governance Workspace   LEADING CANDIDATE
H2 Dossier-first + separate viewer      REJECTED AS LEADING
H3 Three-column workflow cockpit        REJECTED
blocking authority finding              0
P8 HTML                                 NOT OPEN
```

Next gate:

> Operator reviews/approves or changes H1. Only after that may B06 proceed to functional P8 HTML. No B07+ work is opened by this candidate.
