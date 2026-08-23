# T11 — B06 Governance Case R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / ENTRY CLOSED / P7 OPERATOR-APPROVED / P8 R1 READY FOR OPERATOR USE.  
> **Block:** B06 — Governance Case.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 LOCKED.  
> **Current bounded authority:** `../../decisions/governance-case-step-deadline-read.md` + `../../decisions/governance-step-deadline.md`.  
> **Canonical P8 R1:** `t11-b06-governance-case-functional-wireframe.html`.  
> **Implementation:** BLOCKED.  
> **LOCK:** NOT YET — operator must operate/iterate P8 first.

## 1. Entry recovery outcome

B06 entry recovery covered current Product/T2/T3/T6/T8-E/T8-F authority plus useful legacy governance ergonomics.

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

The one material entry finding was B06-F1: real Step deadline truth existed in Controlled Documents but the Governance Case Step projection did not expose it. It is CLOSED / OPERATOR-RATIFIED through:

```text
../../decisions/governance-case-step-deadline-read.md
```

No Product operation, stable route, Permission, lifecycle state, semantic owner, SLA engine or deadline mutation was added.

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

`RETURN_FOR_CHANGES` requires a reason. `ACCEPT` does not invent a reason requirement.

### J4 — communicate without confusing communication with decision

```text
When I need to add context,
I need to add governance feedback separately from the Step Decision,
so that communication never becomes an implicit verdict.
```

### J5 — recover from stale current truth

```text
When another participant or server transition wins first,
I need authoritative case truth to replace my stale assumption,
so that a conflicting Decision is not silently retried or fabricated as success.
```

### J6 — inspect a disclosed case with no current action

```text
When the case is visible but no action is currently offered,
I still need its exact subject and process context,
so that absence of controls is not mistaken for a broken UI or frontend Authorization decision.
```

## 3. Exact authority boundary

### Primary case truth

```text
67 getGovernanceAttempt
   -> GovernanceCaseView
```

Supplies:

```text
governance_attempt_id
state
subject = SubmissionGovernanceSubject | ObsolescenceGovernanceSubject
ordered steps
embedded first feedback page
allowed_actions
```

B06-F1 refines `GovernanceStepView` with exact persisted `due_at?` on timed ACTIVE/DECIDED Steps; PENDING never receives `due_at`.

### Supporting reads

```text
68 listGovernanceFeedback
   -> cursor continuation beyond embedded first20

70 getGovernanceStepDecision
   -> exact singleton Decision when targeted reconciliation requires it

exact governed subject/source
   -> only through already-admitted owner reads
```

Submission:

```text
subject.submission_id
-> exact immutable Submission source
-> never current WorkingContent
```

Obsolescence:

```text
subject.target_revision
-> exact target Revision identity
-> supporting official/release reads must match that exact target
-> never substitute a newer/different Revision
```

If exact target bytes cannot be lawfully resolved, B06 fails closed rather than leaking History or rendering another Revision.

### Writes

```text
69 createGovernanceFeedback
   message
   one Idempotency-Key per logical feedback creation

71 recordGovernanceStepDecision
   ACCEPT
   RETURN_FOR_CHANGES + mandatory reason
   natural singleton PUT
```

### Local frontend state

```text
feedback draft
decision intent / return reason until accepted
confirmation disclosure
viewer page/presentation state
feedback cursor/pagination state
```

No durable business truth is created in the browser.

## 4. Action / Authorization / deadline law

Closed action vocabulary:

```text
accept
return_for_changes
add_feedback
```

`allowed_actions` is server-derived UX guidance only.

```text
hidden control != denied authority proof
shown control  != granted authority proof
command always rechecks current server truth
```

B06 owns no control for:

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

Deadline presentation:

```text
PENDING timed/untimed Step
  no absolute due_at yet

ACTIVE timed Step
  exact persisted due_at

DECIDED timed Step
  exact persisted due_at preserved

now >= due_at
  may present overdue attention
  does not change Step/Attempt state, Authorization or allowed_actions by itself
```

B06 re-reads deadline truth from `getGovernanceAttempt`; B05 navigation state is never case authority.

## 5. Material failure / recovery branches

### Case 404 / disclosure loss

```text
getGovernanceAttempt -> 404
-> reveal neither existence nor denial reason
-> terminal case-not-available presentation
-> safe return to Minha Caixa
```

### Action 403

```text
visible action becomes disallowed at command time
-> server returns 403
-> re-read current case
-> no frontend permission matrix is inferred
```

### Decision conflict

```text
recordGovernanceStepDecision
-> 409 state.governance_step_already_decided
-> re-read current case / exact Decision
-> display winning authoritative Decision
-> never silently retry a different outcome
```

### Feedback ambiguous outcome

```text
createGovernanceFeedback
-> ambiguous/lost response
-> preserve same logical message + same Idempotency-Key
-> retry must not manufacture a duplicate feedback record
```

### Exact-content failure

```text
integrity/dependency failure
-> no successful partial exact-content claim
-> case metadata/process may remain understandable
-> UX must not suggest the unavailable bytes were successfully reviewed
```

The exact control treatment is intentionally exercised by P8 and remains operator-adjudicable UX evidence, not new Product authority.

## 6. P6 — bounded evidence

Useful legacy evidence only:

```text
ordered process sequence helps orientation
real due date is useful when present
visible/persistent decision area reduces action hunting
explicit retry/error state is useful
change-request reason is materially distinct from generic feedback
```

Explicitly rejected legacy semantics:

```text
ExtendSlaDialog
CancelInstanceDialog
fast-forward / "Aprovar já"
password/eSignature ceremony
legacy passed/failed/cancelled Step ontology
peer Approval product/workspace
```

External comparative evidence stopped changing the decision space after Camunda Tasklist + Veeva Vault document workflow. It supports keeping exact governed content and verdict context together while rejecting generic workflow breadth, priority, reassignment, task acceptance, eSignature and multi-document envelope concepts.

P6 = COMPLETE.

## 7. P7 — OPERATOR-APPROVED

Operator adjudication on 2026-08-23 approved:

```text
H1 Content-first Governance Workspace
```

Selected composition:

```text
minimal Governance Case header
  return to Minha Caixa
  Document code + exact Revision/title
  subject kind
  case state
  active Step deadline when present

main workspace
  dominant exact governed-content region
    Exact Read-Only Content Viewer Shell
    immutable Submission or exact obsolescence target

  B06-local governance rail
    A. what is being decided
    B. ordered Steps / active deadline / prior Decisions
    C. governance feedback
    D. deliberate Decision zone
```

Decision behavior:

```text
ACCEPT
  -> explicit confirmation
  -> singleton Decision command

RETURN_FOR_CHANGES
  -> required reason
  -> explicit confirmation
  -> singleton Decision command

ADD FEEDBACK
  -> separate composer
  -> never satisfies RETURN reason
```

Rejected structures:

```text
H2 dossier-first + separate viewer
  REJECTED AS LEADING — forces context switch during content judgment

H3 three-column workflow cockpit
  REJECTED — density/product-shape mismatch and generic workflow gravity
```

Important boundary:

> The B06 governance rail is B06-local. B04's Work operational rail remains B04-local; geometric similarity does not graduate a shared semantic sidebar.

P7 authority findings = 0.

## 8. P8 R1 — functional low-fidelity evidence

Canonical file:

```text
docs/work/current/t11-b06-governance-case-functional-wireframe.html
```

P8 R1 is browser-functional low-fi with deterministic local fixtures/state simulation only. It contains no production React, API integration, client Authorization evaluator or product schema implementation.

### Material regions exercised

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

### Functional fixtures

```text
Submission / timed active Step
Obsolescence / already-overdue active Step
Disclosed case with allowed_actions=[]
Non-disclosable/absent case -> 404-neutral terminal state
presentation clock +20h crossing due_at without lifecycle change
exact-content failure + retry
next Decision -> 409 winner reconciliation
next Decision -> 403 current-authority reconciliation
feedback ambiguous response -> same-logical-key retry
feedback cursor load-more
viewer page navigation
Notification Quick Inbox inherited as global chrome only
B05 return boundary
responsive reflow
Escape / modal focus path + aria-live reconciliation announcements
```

### P8-specific hypothesis under operator review

When exact governed bytes are unavailable while decision actions would otherwise be offered, R1 currently makes the Decision zone unavailable and explains that the exact content is not confirmed.

This is intentionally a **P8 UX hypothesis**, not a new Authorization or lifecycle rule:

```text
content read failure
-> keep case/process context visible
-> prevent the prototype from suggesting completed content review
-> offer content retry
```

Operator use may accept, revise or reject this treatment without reopening Product unless new evidence proves missing semantic authority.

## 9. Responsive / accessibility structure exercised

Desktop:

```text
exact content remains dominant
Governance rail remains independently understandable
Decision zone stays visible without covering content
```

Narrow:

```text
orientation header
-> compact subject/deadline context
-> exact content
-> Steps
-> feedback
-> Decision
```

Structural obligations represented in R1:

```text
Step state not color-only
active Step uses semantic aria-current
absolute deadline retained alongside relative label
overdue includes explicit text
exact Revision identity remains outside rendered bytes
feedback and Decision are separate labeled regions
RETURN reason is explicitly required
modal Escape/cancel path exists
409/403 reconciliation is announced
```

## 10. Current P8 gate

```text
P7 H1                              OPERATOR-APPROVED
functional P8 HTML                 EXISTS
material local interactions        IMPLEMENTED IN R1 FIXTURES
operator browser operation         REQUIRED NEXT
operator LOCK                      NOT YET
P9                                 NOT OPEN
P10                                NOT OPEN
B07+                               NOT OPEN
```

Next gate:

> Operator operates P8 R1, reports friction/findings, and either requests an R2 iteration or explicitly LOCKS B06. Only after LOCK may P9 then P10 execute.
