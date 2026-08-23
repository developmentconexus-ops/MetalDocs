# T11 — B06 Governance Case R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / P7 OPERATOR-APPROVED / P8 R1 OPERATOR-APPROVED / B06-F2 CLOSED / LOCK READY.  
> **Block:** B06 — Governance Case.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 LOCKED.  
> **Current bounded authority:** `../../decisions/governance-case-step-deadline-read.md` + `../../decisions/governance-step-deadline.md` + `../../decisions/governance-review-layer-seam.md`.  
> **B06-F2 record:** `t11-b06-f2-docx-review-layer.md`.  
> **Canonical P8 R1:** `t11-b06-governance-case-functional-wireframe.html`.  
> **Implementation:** BLOCKED.  
> **LOCK:** AWAITING EXPLICIT OPERATOR LOCK — do not infer from prior approvals.

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

B06-F1 corrected the entry gap around real Step deadline truth. B06-F2 later preserved a future provider-neutral inline-review seam without changing current Launch behavior.

## 2. Human jobs

```text
J1 identify the exact Document / Revision / governed subject and inspect immutable content
J2 understand ordered Steps, active Step, prior Decisions/feedback and real deadline when present
J3 ACCEPT or RETURN_FOR_CHANGES deliberately against the exact case
J4 add governance feedback without confusing communication with Decision
J5 reconcile stale/conflicting truth from the server
J6 understand a disclosed case even when no current action is offered
```

`RETURN_FOR_CHANGES` requires a reason. Governance feedback remains a separate immutable fact and never satisfies the RETURN reason by accident.

## 3. Exact current authority boundary

### Reads

```text
67 getGovernanceAttempt
68 listGovernanceFeedback
70 getGovernanceStepDecision
exact governed source only through already-admitted owner reads
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
   one Idempotency-Key per logical creation

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

Deadline law:

```text
PENDING Step
  no absolute due_at

ACTIVE timed Step
  exact persisted due_at

DECIDED timed Step
  exact persisted due_at preserved

now >= due_at
  presentation may say overdue
  Step/Attempt state, Authorization and allowed_actions do not change solely because time passed
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
  -> R1 makes Decision unavailable until exact content returns
```

The content-failure Decision treatment remains UX evidence, not a new Authorization/lifecycle rule.

## 6. P6 evidence disposition

Useful evidence retained:

```text
ordered process sequence supports orientation
real due date is useful when present
visible decision area reduces action hunting
explicit retry/error state is useful
change-request reason is materially different from generic feedback
```

Rejected generic/legacy semantics:

```text
SLA extension
cancel governance instance
fast-forward / "Aprovar já"
password/eSignature ceremony
legacy passed/failed/cancelled ontology
peer Approval product/workspace
reassignment / generic priority / task administration
```

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
H3 three-column generic workflow cockpit
```

The B06 governance rail remains B06-local; B04's Work operational rail is not generalized merely because both sit beside content.

## 8. P8 R1 — OPERATOR-APPROVED

Canonical file:

```text
docs/work/current/t11-b06-governance-case-functional-wireframe.html
```

The operator opened/operated R1 in-browser on 2026-08-23 and approved the current layout/experience.

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

R1 makes the Decision zone unavailable while exact governed bytes are unavailable:

```text
content read failure
-> keep case/process context visible
-> do not imply successful review of unavailable bytes
-> offer content retry
```

No Authorization/lifecycle authority is inferred from the presentation.

## 9. B06-F2 — CLOSED / OPERATOR-RATIFIED

Durable authority:

```text
../../decisions/governance-review-layer-seam.md
```

Forward obligation:

```text
GOV-12 in ../../decisions/forward-obligations.md
```

Ratified future-seam invariants:

```text
exact governed Submission remains immutable
stable Document Discussion != inline governance review
DRAFT EditorialComment remains separately deferred
future selected-range review binds to exact immutable reviewed snapshot
current unanchored GovernanceFeedback remains valid
tracked changes/suggestions require separate semantic promotion
vendor/editor ids never become MetalDocs semantic authority
RETURN leaves old review anchors with old immutable Submission
old anchors never blindly remap onto changed DRAFT bytes
future B04 remediation needs explicit server-authored review-context identity
```

### Current Launch disposition

B06-F2 promotes no present-tense inline-review capability.

```text
GovernanceFeedback wire     unchanged
allowed_actions             unchanged
current census              86 ops / 11 routes / 16 permissions
B04 contract                unchanged
B06 P8 R1 controls          unchanged
```

No disabled/coming-soon review controls are added.

## 10. R2 adjudication after B06-F2

The future seam does not change any current visible B06 behavior.

Therefore:

```text
P8 R1 operator approval      remains valid
P8 R2 inline review          NOT REQUIRED for current Launch
material visible findings    0
```

Creating R2 merely to show deferred controls would violate the no-dormant-capability rule.

## 11. Responsive / accessibility structure

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

Structural obligations represented in R1:

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

## 12. Current gate

```text
P6                                   COMPLETE
P7 H1                                OPERATOR-APPROVED
P8 R1 visual/functional experience   OPERATOR-APPROVED
B06-F1 deadline read                  CLOSED / OPERATOR-RATIFIED
B06-F2 review future seam             CLOSED / OPERATOR-RATIFIED
P8 R2 inline review                   NOT REQUIRED
material current P8 findings          0
B06 LOCK                              READY / AWAITING EXPLICIT OPERATOR LOCK
P9 / P10                              NOT OPEN
B07+                                  NOT OPEN
```

Next gate:

> Operator explicitly LOCKS B06 R1. Only after that explicit LOCK may P9 Screen Contract then P10 pattern consolidation execute. B07 remains closed until B06 P10 closure.
