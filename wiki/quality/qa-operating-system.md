# QA Operating System

> **Last verified:** 2026-05-27
> **Scope:** Canonical operating model for review, QA, root-cause remediation, and feature close-out in MetalDocs.
> **Out of scope:** Module-specific scenario matrices, fixture IDs, or one-off release evidence blobs.

## Purpose

MetalDocs does not treat "implemented" as "done".

A feature is done only when:

- code truth is correct
- runtime behavior is correct
- contract truth is aligned
- user-facing workflows behave correctly
- findings were fixed by root cause rather than symptom patching
- evidence exists so later sessions can reproduce the conclusion without memory

This document defines the closed-loop operating model for that standard.

## Core principles

### Runtime wins

If wiki truth, contract truth, and runtime truth disagree, runnable evidence wins until the mismatch is resolved.

### Root cause over symptom

Do not stack local patches on top of recurring failures. Cluster failures by family and repair the owning boundary.

### Hard-stop honesty

When a fix requires a broader redesign, stop and classify it. Do not continue through a risky partial implementation.

### Separation of roles

Implementation, review, QA, and gatekeeping are distinct responsibilities even when one agent performs them sequentially.

### Evidence before closure

No feature, fix, or QA scenario is complete without the evidence standard required for that scenario class.

## Operating roles

### Implementer

- changes code within the bounded task
- does not self-certify product quality
- must surface assumptions, prerequisites, and affected boundaries

### Code Reviewer

- reviews correctness, architecture, contract alignment, security, maintainability, and test adequacy
- reports findings first, ordered by severity
- treats hidden coupling and drift as first-class defects

### Product QA

- validates the product as a user would
- checks workflows, frontend behavior, interaction states, error handling, API outcomes, async effects, and authz behavior
- does not stop at "request succeeded" when the user-visible result is wrong

### Root-Cause Owner

- clusters findings into bounded failure families
- rejects symptom-only fixes
- decides whether the issue is local, shared, prerequisite-grade, or architecture-grade

### Gatekeeper

- decides whether the work passes, passes with explicit defer, or hits a hard stop
- requires evidence, not confidence

## The five truths

- `runtime truth` - what actually runs now
- `contract truth` - OpenAPI, generated backend surfaces, generated frontend API types
- `wiki truth` - maintained technical memory
- `execution truth` - scripts, checks, and runnable validation procedures
- `review truth` - current findings and closure state for the task

## Classification model

Every issue found during review or QA must be classified before fixing.

- `runtime prerequisite`
- `shared contract prerequisite`
- `module-local implementation`
- `screen-local implementation`
- `wiki-memory drift`
- `workflow/tooling gap`
- `architecture contradiction`
- `defer`

## Severity model

- `critical` - unsafe data, security, authz, integrity, destructive workflow break, or misleading release verdict
- `high` - feature cannot be trusted in normal use, contract/runtime mismatch, major workflow corruption, or major regression risk
- `medium` - bounded correctness, UX, observability, or maintainability issue that should be fixed before considering the area finished
- `low` - polish, local simplification, or non-blocking cleanup

## Closed-loop workflow

Every non-trivial change follows this loop.

1. Establish scope truth.
2. Implement inside the local boundary.
3. Run static and targeted verification.
4. Perform code review.
5. Perform product QA.
6. Classify findings by root-cause family.
7. Fix by root cause.
8. Re-run targeted review and QA.
9. Re-run affected regression scope.
10. Close only with evidence and explicit defer records.

## Delivery gates

### Gate 0 - Scope truth

Required:

- owning boundary is known
- route/runtime/contract ownership is clear
- no unresolved prerequisite-grade contradiction exists

Stop when:

- startup truth is unreliable
- public route ownership is ambiguous
- frontend expects behavior the backend does not own

### Gate 1 - Implementation truth

Required:

- code compiles
- tests/types/lint relevant to the slice pass
- no knowingly broken path is left hidden behind a partial patch

### Gate 2 - Review truth

Required:

- code review completed
- all critical and high findings either fixed or explicitly classified as hard stop / defer with rationale

### Gate 3 - Product QA truth

Required:

- happy path verified
- relevant edge/error/authz paths verified
- user-visible behavior matches intended workflow

### Gate 4 - Root-cause truth

Required:

- findings grouped by family
- fixes address the owning boundary
- no recurring defect is left as scattered local patches

### Gate 5 - Regression truth

Required:

- targeted regression passes for touched surfaces
- broader affected-system regression passes when the change crosses boundaries

### Gate 6 - Evidence truth

Required:

- validation commands, artifacts, and closure state are recorded
- remaining defers are explicit, bounded, and linked to the right wiki location

## Review workflow

Code review should answer:

- is the behavior correct
- is the boundary correct
- is the contract aligned
- is the fix local or hiding a shared problem
- are tests proving the actual invariant

Default review output:

- findings first
- severity order
- exact boundary
- exact risk
- open questions or assumptions
- only then brief change summary

## Product QA workflow

Product QA should behave like a disciplined senior engineer, not a click-through demo.

Default QA pass checks:

- initial load state
- happy path
- empty state
- validation errors
- network/server failure behavior
- auth and authz behavior
- stale state / refresh / cache behavior
- async workflow completion where relevant
- navigation and return-path correctness
- copy and UI states that can mislead the user

When async or workflow-owned behavior exists, QA must split proof into:

- request accepted
- state persisted
- async owner executed
- user-visible final truth updated

## Evidence standards

### Canonical user workflows

Required:

- runtime user-facing proof
- backing API or persisted-state proof

### Contract and route truth

Preferred:

- runtime API proof

Acceptable fallback:

- focused contract or integration proof

### OCC and concurrency

Default:

- focused automated proof

### Fault injection and synthetic failures

Allowed:

- integration or injected proof

Requirement:

- label it explicitly as injected rather than live-runtime evidence

### Auth and authz

Preferred:

- runtime denial or allow proof for user-visible behavior

Fallback:

- contract or integration proof when runtime setup is disproportionate

## Root-cause remediation rules

When findings appear:

1. identify the owning boundary
2. group related failures
3. decide whether the issue is local, shared, prerequisite-grade, or architecture-grade
4. fix the family, not the first visible symptom
5. rerun proof on the full affected family

Do not:

- patch one handler while leaving the contract lying
- patch one screen while the shared query or API wrapper remains wrong
- hide prerequisite failures behind mocks or no-ops
- reopen design debt as silent implementation

## Hard-stop rules

Stop immediately when a fix requires:

- a public API redesign that affects multiple consumers
- a cross-module auth/authz model change
- a storage or provider architecture redesign
- worker/outbox semantic redesign outside the local boundary
- migration framework or policy redesign
- a large cross-screen/frontend-backend coordinated rewrite not included in the task

When stopped, report:

- why it is a hard stop
- which architecture boundary is wrong
- what is locally fixable versus redesign-grade
- the minimum redesign or prerequisite plan needed before resuming

## Feature close-out contract

A feature can close only when:

- implementation gates passed
- review gates passed
- QA gates passed
- root-cause remediation completed for discovered issues
- residual defers are explicit and linked
- wiki truth is updated when code truth changed

## Current execution artifacts

These are the concrete execution layer for the modern documents/approval flow:

- [deep-qa/index.md](deep-qa/index.md)
- [deep-qa/runbook.md](deep-qa/runbook.md)
- [deep-qa/matrix.md](deep-qa/matrix.md)
- [deep-qa/fixtures.md](deep-qa/fixtures.md)

These are the reusable QA checklists for default close-out across feature classes:

- [screen-qa-checklist.md](screen-qa-checklist.md)
- [backend-api-qa-checklist.md](backend-api-qa-checklist.md)
- [workflow-async-qa-checklist.md](workflow-async-qa-checklist.md)
- [release-closeout-checklist.md](release-closeout-checklist.md)

Compatibility breadcrumbs remain under `wiki/references/documents-approval-deep-qa/` for existing path consumers.

## Rollout plan

The QA operating system is only partially promoted until the following work is complete.

### Phase 1 - Canonical policy complete

- keep this page as the canonical QA operating model
- keep [index.md](index.md) as the entrypoint for the quality domain

### Phase 2 - Promote release-quality rules

- normalize `docs/runbooks/release-readiness.md`
- promote the durable release gate into `wiki/quality/release-readiness.md`
- link it from `wiki/tests/` and `wiki/quality/`

### Phase 3 - Promote deep-QA execution artifacts

- keep `wiki/quality/deep-qa/` as the canonical home for the promoted artifact set
- leave compatibility breadcrumbs under `wiki/references/documents-approval-deep-qa/`

### Phase 4 - Add reusable QA checklists complete

- canonical reusable checklists now live under `wiki/quality/`
- quality index exposes them as the default QA entrypoint by workflow class

### Phase 5 - Connect autonomous execution complete

- agent-facing workflow docs must default to implement -> code review -> QA -> root-cause classification -> fix by family -> targeted regression -> broader regression when boundary-crossing -> close only with evidence
- repo instruction layers should treat [../references/ai-operating-system.md](../references/ai-operating-system.md) as the path-stable bridge to this QA operating model until references are fully normalized
