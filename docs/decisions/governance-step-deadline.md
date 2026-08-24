---
id: governance-step-deadline
kind: authority
owner: architecture
summary: Operator-ratified bounded T11 reopen adding optional per-governance-Step elapsed-day deadlines, immutable attempt snapshots, activation-time due_at truth, and no automatic lifecycle consequence.
---

# Governance Step deadline precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified:** 2026-08-23  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Impacts:** T2 lifecycle/governance, T6 My Work meaning, T8-D persistence, T8-E wire read/config shapes, T8-F frontend consumption.  
> **Implementation:** BLOCKED.

## 1. Decision outcome

DevelopmentConexus Method outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ BOUNDED TEMPORAL GOVERNANCE CORRECTION
```

The current bounded sequential GovernanceAttempt model remains correct. The proven defect is narrower: a human Step owns assignee/candidate and activation truth but previously carried no expected completion deadline, leaving My Work unable to distinguish time-critical approval work without inventing frontend priority.

Accepted structure:

```text
GovernanceRouteStep
  label
  selector
  due_in_days?                  // optional positive whole elapsed days

GovernanceAttemptStep
  existing selector/label snapshot
  due_in_days_snapshot?         // immutable attempt snapshot
  existing activated_at?
  due_at?                       // frozen trusted deadline when a timed Step activates
```

`due_at` is Product/domain truth for the activated Step instance. It is not a frontend calculation authority and is not a mutable manual priority field.

## 2. Evidence

### Current MetalDocs authority before this reopen

T2 explicitly deferred `SLA/escalation` absent a named requirement. Current T8-D already persists Step configuration plus immutable attempt Step snapshots and `activated_at`; current T8-E exposed no due field. B05 operator use then supplied the named consumer:

```text
Para aprovação
→ actor needs to know which assigned governance work is time-critical
→ queue must be able to present/filter/order by a real deadline
```

The existing My Work decision had already declared `real SLA/due/priority Product semantics are promoted` as a reopen trigger. That trigger is now satisfied.

### External comparative evidence

The external pass was evidence, not authority. It confirmed the same structural pattern across mature workflow/task systems:

```text
Camunda 8
  user-task dueDate expression evaluated when the user task activates
  Tasklist exposes due date and supports due-date ordering/filtering
  manual priority is a separate generic-workflow feature

Flowable
  task due-date query supports due/on-before/on-after/no-due
  query supports orderByTaskDueDate

ProcessMaker
  task Due In starts when the task becomes active
  due_at is surfaced in task lists with overdue presentation
  richer defaults/escalations exist but are not required by MetalDocs

ServiceNow sequential approvals
  each sequential approver receives its due date only after the prior approval completes
  richer schedules/breach jobs exist but are separate capabilities
```

Primary references consulted:

```text
https://docs.camunda.io/docs/components/modeler/bpmn/user-tasks/
https://docs.camunda.io/docs/components/tasklist/userguide/using-tasklist/
https://github.com/camunda/camunda/blob/672b5a7a0add9f5be80078f191f988c5bb9fc841/webapp/client/apps/orchestration-cluster-webapp/src/tasklist/modules/available-tasks/searchSchema.ts
https://github.com/flowable/flowable-engine/blob/53e93b6681e86dccea720efaa3c0fc2a96f57366/modules/flowable-task-service-api/src/main/java/org/flowable/task/api/TaskInfoQuery.java
https://docs.processmaker.com/v1/docs/form-task
https://github.com/ProcessMaker/processmaker/blob/58a8d9b6b484942205c2ff79b90a2b06bcb94e1d/resources/js/processes/modeler/components/inspector/TaskDueIn.vue
https://downloads.docs.servicenow.com/pdf/enus/servicenow-washingtondc-source-to-pay-operations-enus.pdf
```

External products are intentionally broader than MetalDocs; their generic manual priority, follow-up date, escalation, reassignment, saved-search and business-calendar machinery is not inherited by analogy.

## 3. Root cause and target invariant

Root cause:

> Governance knew who must act and when the Step activated, but not the accepted completion deadline of that active responsibility.

Target invariant:

> A human Governance Step may have an optional route-configured elapsed-day deadline. Each GovernanceAttempt freezes that Step deadline configuration. When the Step activates, the system freezes one trusted `activated_at` and, when configured, one corresponding `due_at`. Passing `due_at` changes attention/overdue presentation only; it never changes lifecycle or authorization by itself.

## 4. Exact temporal semantics

### 4.1 Configuration

```text
due_in_days absent
  = no deadline for that configured Step

due_in_days present
  = positive whole elapsed days
```

There is no hidden/default deadline. Exact executable numeric ceiling is a transport/representability concern; Product establishes only positive whole elapsed days and requires overflow-safe conversion to a representable trusted UTC instant.

Launch time basis:

```text
1 due day = 24 elapsed hours
```

No business-day/calendar/holiday/work-hours semantics exists at Launch.

### 4.2 Attempt snapshot

SUBMIT or human-governed obsolescence attempt creation already freezes one coherent Governance route snapshot. It now freezes `due_in_days_snapshot` with each Step.

```text
route config edited after attempt creation
→ existing attempt deadline snapshot unchanged
→ later attempt uses later accepted config
```

### 4.3 Activation

Exactly one Step is active at a time under existing T2 law.

The same transaction that activates a Step establishes:

```text
activated_at = trusted current instant

if due_in_days_snapshot absent
  due_at = null

if due_in_days_snapshot = N
  due_at = activated_at + (N × 24 elapsed hours)
```

`due_at` is persisted/frozen when the Step activates. It is not continuously recomputed from current route configuration.

Reason for materializing `due_at`:

```text
historical deadline remains exact
future deadline algorithms may change without reinterpreting old Steps
restart cannot change deadline
future business-calendar support can affect later activations without versioning old elapsed-day arithmetic into every read
```

`due_in_days_snapshot` remains provenance of the rule applied; `activated_at` remains activation truth; `due_at` remains the resulting deadline fact. These are differentiated meanings, not duplicate authority.

### 4.4 State correlation

```text
PENDING
  activated_at = null
  due_at = null

ACTIVE
  activated_at required
  due_at present iff due_in_days_snapshot present

DECIDED
  activated_at preserved
  due_at preserved exactly as activation established it
```

A Step without configured deadline remains truthful as `due_at = null`; the system never substitutes a default.

## 5. Deadline breach has no automatic lifecycle effect

```text
trusted now >= due_at
→ Step is overdue for attention semantics
```

It does not by itself:

```text
accept
return_for_changes
cancel
withdraw
advance Step
change assignee/candidate snapshot
change Authorization
create escalation
create notification
```

The Step remains ACTIVE until an existing lawful governance transition changes it.

No timer/worker is required merely to make overdue truth exist; overdue is a read/presentation predicate over persisted `due_at` and trusted current time.

## 6. Credible alternatives / Global Maximum

```text
MANUAL HIGH/MEDIUM/LOW PRIORITY WITHOUT DEADLINE
  rejected: creates unrelated second truth and does not answer actual deadline requirement

ABSOLUTE DEADLINE CHOSEN PER SUBMISSION
  rejected: deadline belongs to the sequential responsibility/Step, not generic Submission identity

RECOMPUTE due_at ON EVERY READ
  rejected after external challenge: couples historical truth to calculation rules and makes future calendar semantics harder to evolve safely

FULL LEGACY/GENERIC SLA ENGINE
  rejected: extension, escalation, reassignment, overseer, timers, business calendar and breach jobs have no current MetalDocs consumer

STEP due_in_days + IMMUTABLE SNAPSHOT + ACTIVATION due_at
  selected: resolves real consumer at existing owner boundary, preserves history, leaves future enrichment additive, adds no generic workflow authority
```

Structural Inversion survives: even if legacy had never contained SLA machinery, a sequential human responsibility with a configured duration and known activation instant still requires an instance deadline. If current implementation instead contained a rich SLA platform, MetalDocs would still remove unsupported mechanisms while preserving this temporal core.

## 7. Authority / boundary

```text
Controlled Documents
  owns deadline configuration and GovernanceAttemptStep temporal truth

DocumentType governance configuration
  owns current due_in_days

GovernanceAttemptStep
  owns frozen due_in_days_snapshot + activated_at + due_at

My Work / Governance Case reads
  may project deadline truth
  own no deadline lifecycle

Frontend
  may format relative/absolute time and derive presentation labels such as overdue
  must not invent a second priority field or mutate deadline locally
```

No new semantic owner is created.

## 8. Persistence realization binding

T8-D target shape is refined as follows:

```text
controlled_docs.document_type_governance_steps
  + due_in_days BIGINT NULL
    CHECK(due_in_days IS NULL OR due_in_days >= 1)

controlled_docs.governance_attempt_steps
  + due_in_days_snapshot BIGINT NULL
  + due_at TIMESTAMPTZ NULL
```

`activated_at` already exists.

Structural laws:

```text
due_in_days_snapshot immutable after attempt creation
due_at immutable after activation
pending Step cannot have due_at
timed active/decided Step must preserve due_at
untimed Step must not invent due_at
```

The implementation proof may use checks/service transaction enforcement as appropriate; this decision does not prescribe trigger-heavy SQL.

## 9. T6 / T8-E / T8-F binding

### T6

Governance work may carry a real optional deadline for the actor's currently active assigned Step. This is attention/triage truth, not lifecycle consequence.

### T8-E

Existing operations remain sufficient. No operation 87 is created.

Required bounded wire refinements before implementation:

```text
GovernanceRouteStep
  + due_in_days? positive integer

WorkGovernanceItem
  + due_at? UtcInstant
```

Exact governance-case Step deadline presentation may be closed when B06 opens; B05 does not pre-design B06 layout.

`listGovernanceWork` filter/order semantics are deliberately NOT closed here; B05-F4 owns that separate decision because the previous neutral code order's reopen trigger has fired.

### T8-F

B05 may present server-projected `due_at` and derive labels relative to trusted client `now` for usability. Filtering and cursor ordering remain server-owned and are pending B05-F4. No manual priority state is introduced.

## 10. Census impact

```text
new application operation       0
new stable SPA route            0
new Permission                  0
new semantic owner              0
new lifecycle state             0
new ETag domain                 0
new Idempotency-Key creation    0
new async worker                0
new manual priority state       0
```

Current accepted census remains 86 operations / 11 routes / 16 PermissionCode values.

## 11. Proof strategy

Later executable-contract/implementation proof must falsify at least:

```text
configured due_in_days does not snapshot into a new attempt
a route edit mutates an existing attempt's deadline snapshot
PENDING Step receives due_at before activation
activation produces due_at from a different instant than its activated_at transaction
untimed Step receives a default due_at
restart changes an existing due_at
passing due_at changes Step lifecycle automatically
frontend invents/stores manual urgency as deadline authority
same attempt deadline changes because current config changed
```

Temporal arithmetic must be overflow-safe and reject an unrepresentable configured duration rather than wrap/truncate a deadline.

## 12. Reopen triggers

Reopen only the implicated temporal decision when evidence proves a need for one of:

```text
business days / holidays / working-hours calendar
sub-day configurable duration becomes materially necessary
absolute per-attempt override becomes a Product requirement
exceptional deadline extension becomes required
breach must trigger Notification/escalation or another durable effect
deadline breach must have a lifecycle consequence
parallel governance introduces multiple simultaneous active Step deadlines
```

Those capabilities are not implied by this decision.
