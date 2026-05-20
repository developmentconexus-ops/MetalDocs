# Documents + Approval Product Plus QA System Design

Date: 2026-05-20
Status: proposed
Scope: modern `documents + approval` flow, including product behavior validation and the engineering workflow needed to run deep QA without manual improvisation

## 1. Goal

Close product + engineering workflow: validate system behavior, and also design the QA, fixtures, and tools process so future sessions do not depend on manual improvisation.

This design treats success as both:

- the runtime product behaving correctly across the modern `documents + approval` flow
- the repository containing enough execution truth that a later deep-QA session can reproduce, validate, classify, and close scenarios without rebuilding the method from scratch

## 2. Problem Statement

The current codebase already contains strong runtime and contract truth for many parts of the modern flow, but deep QA still depends too much on session memory and operator improvisation.

Observed pain points:

- scenario coverage exists, but execution guidance is spread across plans, wiki pages, and session memory
- fixture state is often known only implicitly, which makes re-entry expensive
- some scenario classes require runtime proof while others reasonably rely on contract or integration proof, but that rule is not operationalized in one place
- async ownership changed over time, especially for `scheduled` publish, which can cause stale assumptions about what process owns what
- future sessions risk spending time rediscovering startup truth, worker topology, evidence capture methods, and blocker boundaries instead of validating product behavior

The design goal is to keep the modern product flow canonical while making QA repeatable, bounded, and classification-driven.

## 3. Non-Goals

- redesign the underlying approval or document domain model
- replace the existing module wiki with a new documentation system
- force every fault-injection scenario to become a live browser scenario
- build speculative fixture automation before a recurring need is proven
- reintroduce any legacy fallback path

## 4. Operating Model

The intended operating model has three permanent truths:

1. Product truth

- the canonical surface remains `/documents/:id`
- governed lineage remains in `documents` by `controlled_document_id`
- `document_revisions` remains technical/autosave history only
- `revision_title` remains a governed `documents` field born at finalize / submit-for-review
- `scheduled` is part of QA scope and no longer treated as excluded, but its validation must reflect the new jobs-worker ownership model

2. QA truth

- scenario closure uses `runtime + contract`
- canonical user journeys require runtime browser/API proof
- fault injection, OCC, and other hard-to-stage scenarios may use contract/integration proof when runtime setup is impractical
- a scenario is only considered proved when its required evidence standard has actually been met

3. Workflow truth

- every future session starts from maintained artifacts, not memory
- every scenario ends as either `proved`, `blocked`, or `deferred`
- every non-proved scenario must carry formal classification and named boundary
- runtime, contract, wiki, and execution truth must be updated together when code truth changes

## 5. Primary Artifacts

The system uses two primary artifacts and one support layer.

### 5.1 Deep QA Matrix

The deep QA matrix remains the authoritative execution surface for scenario coverage.

Its purpose is to answer:

- what scenario exists
- why it matters
- what boundary owns it
- what must be proved
- what evidence is required
- whether it is currently proved, blocked, or deferred

The matrix stays scenario-shaped and release-shaped. It must not grow into a mixed debugging manual.

### 5.2 QA Runtime Runbook

The runbook is the operational companion for future deep-QA sessions.

Its purpose is to answer:

- how to start the correct local processes now
- which runtime currently owns each relevant behavior
- how to gather browser, API, worker, and log evidence
- how to distinguish product bugs from fixture gaps, tooling gaps, and architecture prerequisites
- how to close a session cleanly

The runbook stays procedural and execution-oriented.

### 5.3 Fixture and Tooling Catalog

The support layer records the reusable state needed by the matrix and runbook.

Its purpose is to answer:

- which canonical fixtures exist
- what state each one represents
- whether each fixture is reusable or consumable
- how to advance it safely
- which scenarios are runtime-testable versus test-only
- which known tooling limitations still exist

This support layer can live as a bounded section inside the runbook or as a sibling artifact, but it must be maintained as first-class execution truth.

## 6. Matrix Content Model

Each matrix scenario row should include:

- `ID`
- `category`
- `scenario`
- `current owner boundary`
- `preconditions`
- `action`
- `expected result`
- `evidence standard`
- `status`: `proved | blocked | deferred`
- `classification`: only when not proved
- `artifact links`: tests, screenshots, API captures, logs, commits, and wiki sync references

This keeps the matrix optimized for answering: what is required, what is true now, and what is still open.

## 7. Runbook Content Model

The runbook should contain short operational sections:

- `Current runtime topology`
- `Canonical startup paths`
- `Surface-to-owner map`
- `Evidence collection recipes`
- `Worker and async validation`
- `Fault-injection map`
- `Known tooling gaps`
- `Stop rules and classification rules`
- `Session close-out checklist`

Each section should stay short, operational, and current. The runbook is not a second architecture document.

## 8. Fixture and Tooling Catalog Model

Each fixture entry should include:

- `fixture name`
- `controlled document ID`
- `document ID` or related approval instance ID when relevant
- `state represented`
- `how it was created`
- `safe reuse?`
- `how to advance it to next state`
- `when to discard and mint fresh`
- `known caveats`

The fixture catalog must clearly distinguish:

- reusable reference fixtures
- consumable scenario fixtures
- special fixtures for OCC, authz, worker, or cross-tenant validation

## 9. Evidence Standard by Scenario Class

### 9.1 Canonical runtime UX flows

Required evidence:

- runtime browser proof
- backing API proof for the relevant state

Tests are supporting evidence only.

### 9.2 Route and contract truth

Preferred evidence:

- runtime API proof

Acceptable fallback:

- focused contract or integration test

### 9.3 OCC and concurrency

Default acceptable evidence:

- focused automated proof

Reason:

- many race conditions are difficult or unstable to reproduce manually

Minimum evidence:

- exact invariant being protected
- exact focused test coverage or runtime proof if available

### 9.4 Fault injection

Acceptable evidence:

- contract, integration, or fault-injection test proof unless a live runtime switch exists

Requirement:

- it must be explicitly labeled as injected evidence, not live-user evidence

### 9.5 Worker-owned async flows

Required proof model:

- enqueue proof
- worker execution proof
- resulting canonical UI/API truth after execution

If full end-to-end proof is not feasible, each part must be split and classified separately rather than being silently collapsed into one conclusion.

### 9.6 Auth and authz

Preferred evidence:

- runtime denial proof for user-visible authorization behavior

Acceptable fallback:

- contract or integration proof for hard-to-stage cross-tenant or synthetic-principal scenarios

## 10. Session Workflow

Every deep-QA session follows five phases.

### 10.1 Re-grounding

Before testing:

- read the active matrix
- read relevant module wiki pages
- verify startup and runtime ownership truth
- verify worker and scheduler ownership if async behavior is in scope

Output:

- what is already proved
- what is still open
- what changed since the previous session

### 10.2 Fixture Selection

Before interacting with runtime, the operator explicitly chooses:

- reuse known fixture
- advance known fixture
- mint fresh fixture

Output:

- chosen fixture IDs
- target scenario set
- reason that fixture is valid

This prevents state contamination and rediscovery by accident.

### 10.3 Proof Loop

For each scenario:

- reproduce
- collect runtime evidence if required
- compare against contract truth
- classify any mismatch
- stop immediately when the mismatch is prerequisite-grade or cross-boundary
- patch only when the local boundary is clear

### 10.4 Blocker Handling

When a scenario cannot be proved, the session must classify it explicitly, such as:

- missing fixture
- missing tooling
- missing runtime control
- architecture transition
- contract ambiguity
- wiki drift

No scenario may be left in an informal "later" state.

### 10.5 Close-Out

A session closes only after updating:

- matrix state
- runbook notes if execution truth changed
- module wiki if code truth changed
- blocker ledger for any unclosed scenarios

This makes the next session artifact-driven instead of memory-driven.

## 11. Scheduled Publish in the New Ownership Model

`scheduled` remains part of QA scope and must be included in deep validation.

Because runtime ownership has changed, `scheduled` QA must be split into distinct proof layers:

1. schedule request and UI behavior

- approved document exposes correct schedule path
- preconditions, OCC, and authz are enforced correctly

2. enqueue ownership

- API path proves transactional schedule persistence and job enqueue

3. jobs-host execution

- dedicated jobs worker proves cutover execution at effective time

4. resulting product truth

- canonical `/documents/:id`
- `active-document`
- lineage and supersede state
- published head semantics after cutover

This split prevents false conclusions such as blaming UI for worker failures or blaming enqueue logic for post-cutover rendering drift.

## 12. Repo Improvements Required

The recommended implementation approach is `docs + fixture registry`.

That means the repository should gain or promote:

### 12.1 QA Deep-Run Runbook

A maintained runbook documenting:

- startup truth
- browser-first validation loop
- API evidence recipes
- jobs-worker validation
- mismatch classification
- close-out checklist

### 12.2 Fixture Catalog

A maintained catalog of canonical fixture IDs and states, including:

- published baseline
- active draft
- under_review
- approved
- scheduled
- superseded
- rejected and resubmittable
- OCC target
- authz target
- cross-tenant target when available

### 12.3 Scenario-to-Proof Mapping

A compact reference saying:

- scenario class
- required proof type
- acceptable fallback proof
- non-acceptable shortcuts

### 12.4 Worker-Aware Async Validation Notes

A canonical note describing:

- API ownership of enqueue
- jobs-host ownership of delayed execution
- evidence needed for each side
- how to distinguish enqueue failure, worker failure, and post-cutover product drift

### 12.5 Known Tooling Gaps Ledger

A short maintained list of:

- browser limitations
- approved manual-step exceptions
- native input limitations
- scenarios that are test-only
- missing multi-tenant or corruption fixtures

### 12.6 QA Blocker Ledger

A bounded list of obstacles that prevent full proof, such as:

- absent cross-tenant live fixtures
- absent corruption fixtures
- missing async observability
- missing test-only hooks for certain failures

This ledger distinguishes QA infrastructure debt from product defects.

## 13. Why the Recommended Approach is Docs + Fixture Registry

Three approaches were considered:

1. docs only
2. docs + fixture registry
3. docs + fixture registry + helper tooling

The recommended immediate path is `2`.

Why:

- it solves the most expensive current problem, which is rediscovery and fixture ambiguity
- it improves repeatability without overbuilding new tooling
- it keeps runtime truth central instead of hiding it behind synthetic helpers too early

Helper tooling may still be added later for recurring blockers, but should be justified by repeated operational pain rather than created speculatively.

## 14. Success Criteria

This design is successful when:

- future deep-QA sessions can start from known runtime and fixture truth without relying on memory
- the matrix cleanly distinguishes proved, blocked, and deferred scenarios
- evidence standards are applied consistently by scenario class
- `scheduled` is validated as part of normal QA under the new worker ownership model
- code truth changes trigger bounded updates to execution truth and module wiki truth
- hidden regressions become easier to detect because runtime, contract, wiki, and execution artifacts stay aligned

## 15. Implementation Direction

This design does not prescribe product behavior changes by default.

It prescribes documentation and process hardening around the existing modern flow, plus any bounded support changes needed to keep QA execution reproducible. Any later code changes should be justified by explicit gaps discovered while implementing the artifact system described here.
