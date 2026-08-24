---
id: governance-case-step-deadline-read
kind: authority
owner: architecture
summary: Operator-ratified B06 precision projecting exact persisted GovernanceAttempt Step due_at into Governance Case without creating lifecycle, SLA, priority or frontend authority.
---

# Governance Case Step deadline read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified:** 2026-08-23  
> **Block:** B06 — Governance Case  
> **Method:** DevelopmentConexus Engineering Method + Frontend Product Experience Planning Method v2.2  
> **Impacts:** T6 Governance Case meaning, T8-E `GovernanceStepView`, T8-F Governance Case consumption.  
> **Implementation:** BLOCKED.

## 1. Decision outcome

DevelopmentConexus Decision Core outcome:

```text
CURRENT STRUCTURE CONFIRMED
+ BOUNDED GOVERNANCE CASE DEADLINE READ PRECISION
```

B05-F3 established real optional per-Step deadline truth and deliberately left exact Governance Case Step deadline presentation for B06 entry. B06 recovery proved that the current `GovernanceCaseView.steps` shape cannot expose that already-owned truth without a bounded projection correction.

No new semantic owner, lifecycle, permission, route or operation is required.

## 2. Evidence and root cause

Current Controlled Documents authority already owns:

```text
GovernanceRouteStep.due_in_days?
→ GovernanceAttemptStep.due_in_days_snapshot?
→ Step activation freezes activated_at
→ timed activation freezes persisted due_at
```

Current deadline law already requires:

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

The pre-B06 T8-E `GovernanceStepView` projected only Step identity/label/state and Decision. Therefore a direct Governance Case read could not present the exact Step deadline without either depending on B05 queue state, recomputing domain truth in the browser or inventing another read.

Root cause:

> Domain deadline truth was promoted during B05, while the exact Governance Case projection was intentionally deferred until B06 opened.

## 3. Target invariant

> A disclosed Governance Case receives the exact persisted `due_at` of every Step that has activated when that Step was configured with a deadline. Pending Steps never receive a deadline. The browser may format this truth for attention, but it never calculates deadline authority or changes lifecycle/Authorization because time passed.

## 4. Exact bounded wire refinement

Existing operation remains:

```text
67 getGovernanceAttempt
GET /api/v1/governance-attempts/{attempt_id}
→ GovernanceCaseView
```

`GovernanceStepView` is refined to:

```text
PENDING
{
  step_id: Uuid,
  label: ShortText,
  state: pending
}

ACTIVE
{
  step_id: Uuid,
  label: ShortText,
  state: active,
  due_at?: UtcInstant
}

DECIDED
{
  step_id: Uuid,
  label: ShortText,
  state: decided,
  due_at?: UtcInstant,
  decision: GovernanceDecisionView
}
```

Presence/source law:

```text
pending Step
  due_at forbidden

active | decided timed Step
  due_at = exact persisted GovernanceAttemptStep.due_at
  due_at present

active | decided untimed Step
  due_at absent
```

`due_at` uses normal wire absence semantics; no `null` placeholder or hidden/default deadline is introduced.

The projection does not expose `due_in_days_snapshot` or `activated_at` because B06 has no current human need to recalculate or administer the deadline rule. Those remain owner truth/provenance.

## 5. Governance Case presentation semantics

For the ACTIVE Step:

```text
due_at present
→ show the exact deadline in human-readable form
→ relative presentation may be derived for usability

trusted presentation now >= due_at
→ UI may say overdue / atrasado
→ GovernanceStep.state remains active
→ GovernanceAttempt.state unchanged solely because of time
→ allowed_actions unchanged solely because of time
→ Authorization unchanged solely because of time
```

For a DECIDED timed Step, B06 may show its preserved deadline as historical context of that Step. It must not invent:

```text
SLA passed / failed
breach lifecycle state
compliance score
escalation state
penalty
manual priority
```

For PENDING Steps, B06 never predicts or fabricates a future absolute deadline.

## 6. B05 → B06 authority boundary

B05 remains recognition/triage/navigation only:

```text
WorkGovernanceItem.due_at?
→ exact current active-Step deadline for queue attention
→ navigate by governance_attempt_id
```

B06 remains case authority:

```text
/work/governance/:attempt_id
→ getGovernanceAttempt
→ current disclosure/AuthZ rechecked
→ GovernanceCaseView.steps[*].due_at? is read again from case truth
```

Consequences:

```text
direct deep-link works without B05 state
B06 never trusts a queue-carried due_at as current case authority
refresh re-reads case truth
queue and case may format the same persisted fact differently for their different human jobs
```

## 7. Authorization / disclosure boundary

Deadline visibility follows disclosure of the Governance Case and Step projection. It grants nothing.

```text
allowed_actions
  = server-derived UX hints only

recordGovernanceStepDecision / createGovernanceFeedback
  = server rechecks current canonical Authorization + lifecycle truth

due_at
  = attention/context truth only
```

The frontend must not create a permission matrix from Step state, deadline state or `allowed_actions`.

## 8. Global Maximum analysis

Rejected:

```text
COPY B05 due_at INTO B06 NAVIGATION STATE
  breaks direct deep-link and creates stale client authority

ACTIVE-ONLY due_at
  drops a deadline fact the domain deliberately preserves after Decision

GovernanceCaseView.active_step_due_at TOP-LEVEL FIELD
  duplicates Step-owned truth and cannot represent prior Step deadline context cleanly

EXPOSE activated_at + due_in_days_snapshot AND RECOMPUTE
  moves domain arithmetic to the browser and weakens future deadline-algorithm evolution

NEW DEADLINE ENDPOINT
  adds an operation for truth already naturally owned by getGovernanceAttempt

FULL SLA/BREACH MODEL
  no current MetalDocs consumer; conflicts with the bounded deadline decision
```

Selected:

```text
optional exact due_at on ACTIVE + DECIDED GovernanceStepView variants
+ no due_at on PENDING
+ no lifecycle/AuthZ consequence
```

Structural Inversion survives: even if the legacy approval UI had no deadline display, the current case still needs its already-owned active responsibility deadline; if legacy had a rich SLA platform, the unsupported mechanisms would still be removed.

## 9. T6 / T8-E / T8-F binding

### T6

Governance Case continues to mean:

```text
exact GovernanceAttempt
+ exact immutable governed subject
+ exact Step sequence/current Step
+ bounded feedback/decisions
+ live allowed_actions
```

A Step deadline is case context, not a new Product action or state.

### T8-E

Operation 67 remains sufficient. Refine only `GovernanceStepView` as §4. No operation 87 is created.

### T8-F

B06 may:

```text
show exact active-Step deadline
show relative deadline wording
show overdue presentation when now >= due_at
show preserved decided-Step deadline as Step context
```

B06 must not:

```text
calculate authoritative due_at
mutate due_at
inherit due_at from B05 as authority
convert overdue into lifecycle state
create SLA extension/escalation/reassignment controls
create frontend priority
```

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

## 11. Proof obligations

Later executable-contract/implementation proof must falsify at least:

```text
direct B06 deep-link requires prior B05 navigation state
B06 active timed Step due_at differs from persisted GovernanceAttemptStep.due_at
B06 decided timed Step loses or changes its frozen due_at
pending Step exposes a synthetic due_at
untimed active/decided Step receives a default due_at
route config edit rewrites an existing attempt Step deadline
browser recomputes authoritative due_at from duration/config
overdue presentation changes GovernanceStep/GovernanceAttempt lifecycle
passing due_at suppresses or grants an action without server authority
frontend introduces SLA extension/escalation/manual priority
```

## 12. Reopen triggers

Reopen only the implicated decision when evidence proves a need for:

```text
business-calendar / timezone-owned deadline semantics
sub-day deadline configuration
absolute per-attempt deadline override
exceptional deadline extension
breach-triggered durable effect or escalation
breach-triggered lifecycle consequence
parallel active Steps with multiple simultaneous deadline attention needs
an explicit user need for activated_at/duration provenance in the case UI
```

Those capabilities are not implied by B06-F1.
