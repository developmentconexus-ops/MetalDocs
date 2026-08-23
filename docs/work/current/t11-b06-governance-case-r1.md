# T11 — B06 Governance Case R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / ENTRY CLOSED / P7 OPERATOR-APPROVED / P8 R1 OPERATOR-APPROVED / B06-F2 WRITTEN RATIFICATION PENDING.  
> **Block:** B06 — Governance Case.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 LOCKED.  
> **Current bounded authority:** `../../decisions/governance-case-step-deadline-read.md` + `../../decisions/governance-step-deadline.md`.  
> **B06-F2 candidate:** `t11-b06-f2-docx-review-layer.md`.  
> **Canonical P8 R1:** `t11-b06-governance-case-functional-wireframe.html`.  
> **Implementation:** BLOCKED.  
> **LOCK:** NOT YET — B06-F2 written adjudication must close first.

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
```

B06-F1 corrected the one entry gap: persisted Governance Step deadline truth is projected into the exact Governance Case. It is CLOSED / OPERATOR-RATIFIED through:

```text
../../decisions/governance-case-step-deadline-read.md
```

No Product operation, stable route, Permission, lifecycle state, semantic owner, SLA engine or deadline mutation was added.

## 2. Human jobs

### J1 — understand exactly what is being governed

```text
Identify the exact Document / Revision / governed subject and inspect its immutable content,
so the Decision is never made against a different DRAFT, current Revision or stale queue summary.
```

### J2 — understand where the case is now

```text
Understand ordered Steps, active Step, prior Decisions/feedback and the real deadline when one exists,
without learning a generic workflow engine.
```

### J3 — decide safely

```text
ACCEPT or RETURN_FOR_CHANGES deliberately against the exact case.
RETURN_FOR_CHANGES requires a reason.
```

### J4 — communicate without confusing communication with Decision

```text
Governance feedback remains a separate immutable fact.
It never becomes an implicit verdict and never satisfies the RETURN reason by accident.
```

### J5 — recover from stale current truth

```text
When another participant/server transition wins first,
replace stale assumptions with authoritative case truth rather than silently retrying a conflict.
```

### J6 — inspect a disclosed case with no current action

```text
Absence of controls must remain understandable as current case truth,
not be inferred as a frontend Authorization decision or broken UI.
```

## 3. Exact current authority boundary

### Reads

```text
67 getGovernanceAttempt
   -> GovernanceCaseView

68 listGovernanceFeedback
   -> cursor continuation beyond embedded first page

70 getGovernanceStepDecision
   -> exact singleton Decision when targeted reconciliation requires it

exact governed source
   -> only already-admitted owner reads
```

Submission subject:

```text
subject.submission_id
-> exact immutable Submission source
-> never current WorkingContent
```

Obsolescence subject:

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

Current closed action vocabulary:

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

B06 owns no current control for:

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
inline anchored comment / resolve / reply
tracked-change / suggestion mutation
```

Deadline presentation:

```text
PENDING Step
  no absolute due_at

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

```text
404 case/disclosure loss
  -> neutral unavailable state; no existence leak; safe return to Minha Caixa

403 action-time denial
  -> refetch current case; server remains authority

409 state.governance_step_already_decided
  -> refetch case/exact Decision; show winner; never silently retry a different outcome

ambiguous createGovernanceFeedback result
  -> preserve same logical message + Idempotency-Key for retry

exact-content integrity/dependency failure
  -> no partial-success exact-content claim
  -> keep understandable case/process context
  -> current P8 hypothesis makes Decision surface unavailable until exact content returns
```

The content-failure Decision treatment remains UX evidence, not new Authorization/lifecycle authority.

## 6. P6 evidence disposition

Useful evidence retained:

```text
ordered process sequence supports orientation
real due date is useful when present
visible decision area reduces action hunting
explicit retry/error state is useful
change-request reason is materially different from generic feedback
```

Explicitly rejected legacy/generic workflow semantics:

```text
SLA extension
cancel governance instance
fast-forward / "Aprovar já"
password/eSignature ceremony
legacy passed/failed/cancelled ontology
peer Approval product/workspace
reassignment / generic priority / task administration
```

External study converged on keeping exact governed content and verdict context together while rejecting generic workflow breadth.

P6 = COMPLETE.

## 7. P7 — OPERATOR-APPROVED

Selected hypothesis:

```text
H1 Content-first Governance Workspace
```

Composition:

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

Rejected:

```text
H2 dossier-first + separate viewer
  context switching weakens content judgment

H3 three-column workflow cockpit
  density/product-shape mismatch and generic workflow gravity
```

The B06 governance rail remains B06-local. B04's Work operational rail is not generalized merely because both sit beside content.

## 8. P8 R1 — OPERATOR-APPROVED VISUAL / FUNCTIONAL EVIDENCE

Canonical file:

```text
docs/work/current/t11-b06-governance-case-functional-wireframe.html
```

The operator opened/operated R1 in-browser on 2026-08-23 and explicitly approved the current layout/experience before raising B06-F2 as a future-capability planning concern.

R1 exercises:

```text
Submission / timed active Step
Obsolescence / overdue active Step
case with allowed_actions=[]
404-neutral unavailable case
clock crossing due_at without lifecycle change
exact-content failure + retry
409 Decision winner reconciliation
403 current-authority reconciliation
feedback + cursor + ambiguous replay
ACCEPT confirmation
RETURN_FOR_CHANGES required reason + confirmation
viewer page navigation
B01N Quick Inbox global chrome
B05 return boundary
responsive reflow
modal Escape/focus + aria-live reconciliation
```

### P8 content-failure hypothesis

R1 makes the Decision zone unavailable while exact governed bytes are unavailable. This is an operator-reviewable UX treatment only:

```text
content read failure
-> keep case/process context visible
-> do not imply successful review of unavailable bytes
-> offer content retry
```

No Authorization/lifecycle rule is inferred from it.

## 9. B06-F2 — DOCX Review Layer finding

During R1 review the operator identified a real product-evolution need: Word-like selected-range comments and possibly tracked changes/suggestions for DOCX governance review.

Written candidate:

```text
t11-b06-f2-docx-review-layer.md
```

Current candidate direction:

```text
1. exact governed Submission remains immutable;
2. stable Document Discussion is NOT inline governance review;
3. deferred DRAFT EditorialComment is NOT silently reused;
4. future selected-range comment should bind to the exact immutable governed snapshot;
5. current ordinary GovernanceFeedback remains valid and unchanged;
6. future tracked-change/suggestion is a separate semantic decision from comments;
7. editor/vendor ids never become MetalDocs semantic authority;
8. after RETURN_FOR_CHANGES, old review anchors stay with the old immutable Submission;
9. old anchors are never blindly overlaid onto changed DRAFT bytes;
10. future author remediation needs explicit server-authored review-context identity rather than browser History reconstruction.
```

Mechanism posture:

```text
EigenPal Apache core       current viable DOCX baseline
EigenPal Pro               future commercial review candidate
SuperDoc Community         not admitted as proprietary baseline under AGPLv3
SuperDoc commercial        future commercial candidate
other mechanisms           may compete behind the same seam
```

### Critical disposition

B06-F2 does **not** currently promote inline-review capability.

Therefore:

```text
current GovernanceFeedback wire  unchanged
current allowed_actions           unchanged
current census                    unchanged
current B04 contract              unchanged
P8 R2 inline-review controls      NOT OPEN
```

Do not add disabled/coming-soon comment, resolve, reply, suggest or tracked-change controls merely to reserve UI space.

## 10. Responsive / accessibility structure

Desktop:

```text
exact content remains dominant
Governance rail remains independently understandable
Decision remains reachable without covering content
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

Structural obligations already represented:

```text
Step state not color-only
active Step has semantic current indication
absolute deadline remains alongside relative wording
overdue has explicit text
exact Revision identity remains outside rendered bytes
feedback and Decision are separately labeled
RETURN reason explicitly required
modal Escape/cancel path
409/403 reconciliation announced
```

## 11. Current gate

```text
P7 H1                                OPERATOR-APPROVED
P8 R1 visual/functional experience   OPERATOR-APPROVED
B06-F2 high-level direction          APPROVED IN CHAT
B06-F2 written candidate             AWAITING WRITTEN OPERATOR RATIFICATION
P8 R2 inline review                  NOT OPEN
B06 LOCK                              NOT YET
P9 / P10                              NOT OPEN
B07+                                  NOT OPEN
```

Next gate:

> Operator reviews/ratifies `t11-b06-f2-docx-review-layer.md`. If ratified, promote only the durable future seam/reopen obligations needed to prevent architectural dead-end, while preserving the current 86-operation Launch API and R1 controls. Then re-evaluate B06 LOCK; do not create an R2 unless the ratified result materially changes current visible Launch behavior.
