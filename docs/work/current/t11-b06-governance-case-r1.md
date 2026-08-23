# T11 — B06 Governance Case R1 — Method v2.2 locked

> **Status:** LOCKED / OPERATOR-RATIFIED / P9-P10 COMPLETE.  
> **Block:** B06 — Governance Case.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 LOCKED.  
> **Bounded authorities:** `../../decisions/governance-case-step-deadline-read.md` + `../../decisions/governance-step-deadline.md` + `../../decisions/governance-review-layer-seam.md`.  
> **Canonical P8:** `t11-b06-governance-case-functional-wireframe.html`.  
> **Post-LOCK proof:** `t11-b06-screen-contract.md` + `t11-b06-pattern-consolidation.md`.  
> **Implementation:** BLOCKED.

## 1. Lock basis

The operator explicitly approved, in sequence:

```text
P7 H1 Content-first Governance Workspace
functional P8 R1 after browser operation
B06-F2 future DOCX Review Layer seam
final B06 LOCK authorization
```

No material current-Launch B06 finding remains open.

B06-F2 did not require a P8 R2 because it promotes no present-tense inline-review behavior; adding dormant controls would violate the method/repository hard stops.

## 2. Locked Product/architecture boundary

Stable route:

```text
/work/governance/:attempt_id
```

Primary case truth:

```text
67 getGovernanceAttempt
68 listGovernanceFeedback
70 getGovernanceStepDecision
```

Writes:

```text
69 createGovernanceFeedback
71 recordGovernanceStepDecision
```

Current action vocabulary:

```text
accept
return_for_changes
add_feedback
```

`allowed_actions` is server-derived UX guidance only. Every command rechecks current canonical server truth.

B06 owns no current control for:

```text
release/publish
Submission withdrawal
Revision cancellation
obsolescence withdrawal
GovernanceAttempt cancellation
reassignment/delegation/quorum
SLA extension/escalation/manual priority
WorkingContent mutation
inline anchored comment/reply/resolve
tracked-change/suggestion mutation
```

## 3. Locked composition

```text
minimal Governance Case header
  return to Minha Caixa
  Document code + exact Revision/title
  governed subject kind
  case state
  active Step deadline when present

CONTENT-FIRST GOVERNANCE WORKSPACE
  dominant exact-content region
    Exact Read-Only Content Viewer Shell
    exact immutable Submission
    OR exact obsolescence target

  B06-local governance rail
    1. governed subject summary
    2. ordered Steps / active deadline / prior Decisions
    3. governance feedback
    4. deliberate Decision zone
```

Narrow order remains:

```text
orientation
-> compact subject/deadline
-> exact content
-> Steps
-> feedback
-> Decision
```

## 4. Exact-content / subject law

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
-> supporting official/release reads must prove that exact target
-> never substitute another current/newer Revision
```

If exact governed bytes cannot be proven/lawfully resolved, B06 fails closed rather than rendering substitute content.

Locked R1 content-failure treatment:

```text
exact bytes unavailable
-> keep understandable case/process context
-> do not claim successful content review
-> Decision surface unavailable until exact content returns
-> explicit retry
```

This is UX structure, not Authorization/lifecycle authority.

## 5. Governance Step deadline law

B06-F1 is CLOSED / OPERATOR-RATIFIED.

```text
PENDING
  due_at forbidden

ACTIVE / DECIDED timed Step
  due_at = exact persisted GovernanceAttemptStep.due_at

ACTIVE / DECIDED untimed Step
  due_at absent
```

Presentation may derive relative wording and explicit overdue attention.

```text
now >= due_at
!= Step lifecycle transition
!= GovernanceAttempt transition
!= Authorization change
!= allowed_actions change
!= SLA breach state
```

B06 rereads this truth from op67; B05 queue-carried deadline state is never current case authority.

## 6. Feedback / Decision separation

Governance feedback:

```text
immutable multi-record fact
separate from Decision
op69 uses one Idempotency-Key per logical creation
ambiguous response -> retry same logical message/key
```

Decision:

```text
ACCEPT
  -> deliberate confirmation
  -> singleton op71

RETURN_FOR_CHANGES
  -> mandatory reason
  -> deliberate confirmation
  -> singleton op71
```

Decision singleton conflict:

```text
first accepted Decision wins
later different outcome/reason -> 409 state.governance_step_already_decided
-> reread authoritative case/Decision
-> never silently retry a changed outcome
```

Command-time 403 likewise causes authoritative reread; the frontend never builds a permission matrix.

## 7. B06-F2 — future Governance Review Layer seam

B06-F2 is CLOSED / OPERATOR-RATIFIED through:

```text
../../decisions/governance-review-layer-seam.md
```

Forward obligation:

```text
GOV-12 in ../../decisions/forward-obligations.md
```

Future invariants preserved without current capability:

```text
future selected-range review binds to exact immutable reviewed snapshot
stable Document Discussion != inline governance review
DRAFT EditorialComment remains separately deferred
current ordinary GovernanceFeedback remains valid
tracked changes/suggestions require separate semantic promotion
vendor/editor ids never become MetalDocs semantic authority
RETURN leaves old anchors with old immutable Submission
old anchors never blindly remap onto changed DRAFT bytes
future B04 remediation needs explicit server-authored review-context identity
```

Current Launch remains unchanged:

```text
GovernanceFeedback wire      unchanged
allowed_actions              unchanged
B04 contract                 unchanged
B06 P8 controls              unchanged
operations/routes/permissions 86 / 11 / 16
```

## 8. Locked P8 evidence

Canonical browser-functional evidence:

```text
docs/work/current/t11-b06-governance-case-functional-wireframe.html
```

R1 exercises:

```text
Submission / timed active Step
obsolescence / overdue Step
allowed_actions=[] case
404-neutral unavailable case
deadline crossing without lifecycle mutation
exact-content failure/retry
feedback pagination + ambiguous replay
ACCEPT confirmation
RETURN required reason + confirmation
409 winner reconciliation
403 authority reconciliation
B01N Quick Inbox global chrome
B05 return boundary
responsive reflow
Escape/focus + aria-live behavior
```

The operator operated and approved the R1 experience, then explicitly authorized final B06 LOCK after B06-F2 closure.

## 9. State authority

```text
SERVER STATE
  GovernanceCaseView
  GovernanceFeedback pages
  GovernanceDecisionView
  exact content/source responses

NAVIGATION / URL
  governance_attempt_id + stable B06 route

FORM DRAFT
  feedback message before acceptance
  RETURN reason before Decision acceptance

EPHEMERAL UI
  viewer presentation/page
  composer disclosure
  confirmation modal
  content failure/retry
  conflict/denial announcements
  inherited Quick Inbox visibility
```

No fifth durable/global frontend state class exists.

## 10. P9 Screen Contract — COMPLETE

Canonical proof:

```text
docs/work/current/t11-b06-screen-contract.md
```

Closure:

```text
material B06 regions/controls traced        18 / 18
unbound material controls                   0
invented operations                         0
operation 87+                               absent
screen-shaped APIs                          0
frontend Authorization evaluator            0
WorkingContent mutation path                 0
current inline-review controls               0
navigation identities unsourced              0
material B06 Screen Contract findings        0
```

## 11. P10 Pattern Consolidation — COMPLETE

Canonical proof:

```text
docs/work/current/t11-b06-pattern-consolidation.md
```

Shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
Exact Read-Only Content Viewer Shell
```

No new shared pattern graduates from B06.

B06-local semantics remain local, including Content-first Governance Workspace, Governance rail, Step/deadline context, GovernanceFeedback, Decision zone and authoritative Decision reconciliation.

Closure:

```text
existing locked shared patterns reused          3
new shared semantic patterns graduated           0
B06-local semantic patterns retained             8
false abstractions introduced                    0
unexplained duplicate locked semantic patterns  0
```

## 12. Block closure

```text
B06 authority recovery                    COMPLETE
B06-F1 deadline projection                CLOSED / OPERATOR-RATIFIED
P6                                       COMPLETE
P7 H1                                    OPERATOR-APPROVED
P8 R1                                    OPERATOR-APPROVED
B06-F2 Governance Review Layer seam       CLOSED / OPERATOR-RATIFIED
P8 R2 inline review                       NOT REQUIRED
B06 LOCK                                  LOCKED / OPERATOR-RATIFIED
P9 Screen Contract                        COMPLETE
P10 pattern consolidation                 COMPLETE
```

B06 is closed for FP1. B07 is now eligible as the next block but remains NOT OPEN until explicitly begun through the roadmap.
