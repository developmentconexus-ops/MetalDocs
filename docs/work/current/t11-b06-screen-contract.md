# T11 — B06 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B06 — Governance Case.  
> **Depends on:** B06 operator LOCK, Governance Case deadline read authority, Governance Review Layer future seam, current T6/T8-E/T8-F contracts.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the locked B06 functional wireframe is realizable by current authority without inventing frontend business truth, a screen-shaped API, a WorkingContent mutation path or dormant inline-review capability.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B06-01 case route | open the exact assigned governance case | op67 `getGovernanceAttempt` | route only | `governance_attempt_id` | 404/non-disclosable -> neutral unavailable state | no queue-carried case authority | READY |
| B06-02 orientation header | know exact Document / Revision / subject / case state | `GovernanceCaseView.subject + state` | return to B05 only | returned Document/Revision/attempt refs | stale navigation reconciles through op67 | no lifecycle reconstruction | READY |
| B06-03 subject summary | understand exactly what is being judged | `SubmissionGovernanceSubject` or `ObsolescenceGovernanceSubject` | none by virtue of display | `submission_id` or `request_id + target_revision` | target mismatch -> fail closed | no substitution with current DRAFT/newer Revision | READY |
| B06-04 active deadline | understand current Step time context | op67 `GovernanceStepView.due_at?` | none | exact active Step returned by case | stale time presentation refreshes; pending never receives due_at | no client due-date calculation/SLA state | READY |
| B06-05 ordered Steps | understand route progression and prior Decisions | op67 `steps[]`; op70 when targeted Decision reconciliation is required | none by display | returned `step_id` / Decision ids | conflict/stale state -> refetch authoritative case | no generic workflow engine | READY |
| B06-06 exact Submission viewer | judge immutable submitted bytes | op63 `getSubmission` + op64 `getSubmissionSource` | page/viewer presentation only | exact `subject.submission_id` | integrity/dependency failure -> no partial-success review | no WorkingContent/edit controls | READY |
| B06-07 exact obsolescence-target viewer | judge the exact current EFFECTIVE target | `subject.target_revision` + matching official Release/source reads (ops72–74 as applicable) | viewer presentation only | exact target Revision + returned matching release/rendition refs | if exact target bytes cannot be proven, render no substitute | no newer/different Revision fallback | READY |
| B06-08 feedback timeline | inspect immutable governance feedback | embedded first page in op67 + op68 `listGovernanceFeedback` | load older page | attempt id + opaque cursor | page failure leaves already-known feedback intact | no stable-Document Discussion substitution | READY |
| B06-09 add feedback | add separate case context | `allowed_actions` hint + current case | op69 `createGovernanceFeedback` | exact attempt id + one logical Idempotency-Key | ambiguous outcome -> retry same logical key/message | no Decision inferred from feedback | READY |
| B06-10 ACCEPT | record deliberate approval of active Step | `allowed_actions` + active Step | op71 `recordGovernanceStepDecision(ACCEPT)` after confirmation | exact attempt + active `step_id` | 403/409 -> authoritative reread | no optimistic lifecycle advance | READY |
| B06-11 RETURN_FOR_CHANGES | return exact Submission/request with explicit reason | same case truth | op71 `RETURN_FOR_CHANGES + reason` after confirmation | exact attempt + active `step_id` | blank reason blocked locally; 403/409 authoritative | no feedback-as-return-reason shortcut | READY |
| B06-12 no-actions disclosed case | understand visible case without current mutation | op67 `allowed_actions=[]` | none | current case | absence of actions stays understandable | no frontend AuthZ matrix | READY |
| B06-13 409 reconciliation | understand another participant won first | op71 409 + op67/op70 reread | no changed-outcome silent retry | returned winning Decision | announce and replace stale assumption | no client conflict resolver/verdict overwrite | READY |
| B06-14 403 reconciliation | understand action is no longer admitted | command 403 + op67 reread | none until truth returns new actions | exact case | preserve safe local form intent where useful | no permission inference from prior button visibility | READY |
| B06-15 exact-content failure | avoid deciding as if unavailable bytes were reviewed | exact content read failure + retained case context | retry exact-content read | same immutable subject identity | Decision surface unavailable in locked R1 until exact bytes return | presentation block is not Authorization authority | READY |
| B06-16 global notification chrome | preserve app-wide attention behavior | B01N Notifications owner/read family | Quick Inbox open/close only | Notification source identities | B06 does not interpret Notification truth | no B06 Notification store | READY |
| B06-17 B05 return boundary | return to assigned-work context | stable `/work` route | navigate only | no carried business state required | B05 rereads its own queue truth | no mutation inside queue handoff | READY |
| B06-18 responsive/accessibility | preserve same governance meaning on narrow/accessibility paths | same server truths | presentation-only reflow/focus/modal behavior | none new | material action/context remains accessible | no mobile-specific business semantics | READY |

## 3. Exact operation homes used by B06

```text
63  getSubmission
64  getSubmissionSource
67  getGovernanceAttempt
68  listGovernanceFeedback
69  createGovernanceFeedback
70  getGovernanceStepDecision
71  recordGovernanceStepDecision
72  getRelease
73  getReleaseSource
74  getOfficialRenditionContent
```

Ops72–74 are supporting exact-target reads only when the governed subject is obsolescence and the resolved official representation matches `subject.target_revision` exactly. They do not give B06 authority to substitute a different official/current Revision.

B06 adds no operation 87+.

## 4. Bidirectional trace

### Product/backend → frontend

```text
GovernanceCaseView
-> case route/header/subject/Steps/deadline/allowed-actions

exact Submission or exact obsolescence-target bytes
-> Exact Read-Only Content Viewer Shell

GovernanceFeedback pages
-> feedback timeline

GovernanceDecision singleton/current conflict truth
-> Decision zone + reconciliation

B01N Notification chrome
-> global bell / Quick Inbox only
```

### Frontend → Product/backend

```text
Load case
-> op67

Load older feedback
-> op68

Add feedback
-> op69 with one logical Idempotency-Key

Approve
-> confirmation -> op71 ACCEPT

Solicitar mudanças
-> mandatory reason -> confirmation -> op71 RETURN_FOR_CHANGES

409/403
-> op67 and op70 when exact Decision targeting is needed

Submission viewer
-> op63/64 exact subject

Obsolescence viewer
-> exact target -> matching ops72–74 as applicable
```

Unbound material controls: **0**.  
Invented application operations: **0**.  
Screen-shaped APIs: **0**.

## 5. Client state classes

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
  viewer page/presentation
  feedback composer disclosure
  confirmation modal
  content retry/failure presentation
  conflict/denial announcement
  Quick Inbox visibility inherited from B01N
```

No fifth durable/global frontend state class is introduced.

## 6. Material failure intent

```text
404 / non-disclosable case
  neutral unavailable state; no existence/denial detail leak

403 command denial
  server remains authority; refetch case; no local permission matrix

409 state.governance_step_already_decided
  show winning authoritative Decision after reread; never silently retry another outcome

ambiguous op69 result
  preserve same logical message + Idempotency-Key; never duplicate feedback

exact-content failure
  keep case/process context but do not claim bytes were reviewed; retry exact resource

feedback page failure
  preserve already-loaded immutable feedback; retry continuation
```

## 7. Access / authority proof

```text
allowed_actions != Authorization snapshot
button visibility != Authorization
case participation != WorkingContent/history grant
GovernanceFeedback != GovernanceDecision
RETURN reason != prior feedback
now >= due_at != SLA breach lifecycle
exact content viewer != mutation authority
B05 queue state != B06 current case authority
future Review Layer seam != current inline-review capability
```

Every material command rechecks current server truth.

## 8. P9 closure

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

P9 is complete for the locked B06 scope.
