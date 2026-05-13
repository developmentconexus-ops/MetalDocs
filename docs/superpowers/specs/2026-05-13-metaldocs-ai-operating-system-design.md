# MetalDocs AI Operating System

> Date: 2026-05-13
> Status: approved design, pre-implementation
> Scope: repo operating model for agentic development, runtime reliability, contract alignment, and wiki governance

## 1. Purpose

MetalDocs already has strong raw ingredients: a technical wiki, specialized skills, module docs, backlog structure, ADRs, and screen workflows. The problem is not the absence of guidance. The problem is that the current guidance is not enforced as a compact operating system.

Recent failures exposed the gap:

- runtime truth, binary truth, DB truth, OpenAPI truth, frontend-wrapper truth, and wiki truth can drift independently
- startup can appear successful while using stale binaries or stale migration ledger state
- screen work can begin before the local system is trustworthy end to end
- module wiki sync can update part of the technical memory while silently missing affected surfaces or modules
- repo instructions tell agents what to use, but not strongly enough when to stop, escalate, or classify drift

The goal of this operating system is to make the repo behave like a professional, senior-level, high-quality SaaS engineering environment for agentic work without becoming process-heavy or bureaucratic.

This design keeps the model intentionally small, strict, and executable.

## 2. Design Principles

### 2.1 Small but hard

The operating model must stay compact enough that agents will actually follow it every session.

It must not become a giant framework.

### 2.2 Layered truth

MetalDocs uses four truth layers:

- **Runtime truth**: what actually runs now
- **Contract truth**: what public consumers are supposed to rely on
- **Wiki truth**: what future agents and developers use as governed technical memory
- **Execution truth**: what must happen before work may be considered valid

### 2.3 Classify before continuing

When layers disagree, the default behavior is not “fix whatever looks easy.”

The agent must classify the issue first, then continue only if the issue is local to the current task boundary.

### 2.4 Executable over aspirational

Rules in prose are not enough.

Hard gates must live in:

- skills
- repo instructions (`AGENTS.md`, `CLAUDE.md`)
- executable scripts or preflight checks

### 2.5 Workflow hardening is part of fixing the bug

If a failure reveals a workflow gap, the job is not complete until the workflow is updated so the same class of mistake becomes harder to repeat.

## 3. Truth Hierarchy

### 3.1 Runtime truth

Runtime truth answers: “What exists and runs now?”

It includes:

- mounted routes
- actual handler wiring
- actual DB schema and constraints
- startup/build behavior
- auth/session behavior
- local service dependencies

Runtime truth is the source of truth for existence and live behavior.

### 3.2 Contract truth

Contract truth answers: “What should public consumers rely on?”

It includes:

- OpenAPI
- generated backend interfaces
- generated frontend API types

Contract truth is the target public interface, but it is only trustworthy when aligned with runtime truth.

### 3.3 Wiki truth

Wiki truth answers: “What technical memory should future agents and developers trust?”

It includes:

- module docs
- tech-debt registers
- refactor backlogs
- route truth tables
- artifacts
- ADRs
- roadmap and workflow docs

Wiki truth is governed memory, not the authority on live route existence.

### 3.4 Execution truth

Execution truth answers: “What must happen before we can say the work is real and valid?”

It includes:

- canonical startup scripts
- preflight checks
- verification commands
- mandatory gates inside skills

## 4. Standard Classification Model

Every meaningful drift or failure must be classified into one of these categories:

1. **Runtime prerequisite**
   - startup, build, binary freshness, env, migrations, DB ledger, auth bootstrap, local services

2. **Shared contract prerequisite**
   - runtime/OpenAPI/codegen/frontend-wrapper mismatch affecting more than the target task

3. **Module-local implementation**
   - safe to fix inside one module/task/PR

4. **Screen-local implementation**
   - safe to fix inside one screen PR

5. **Wiki-memory drift**
   - docs, debt, backlog, route truth, artifacts, roadmap, stale module memory

6. **Workflow/tooling gap**
   - weak skill, weak repo instruction, weak startup script, missing gate, bad verifier

7. **Defer**
   - missing product semantics, unsupported backend behavior, or a too-large redesign

This classification model is mandatory because it defines when an agent may continue and when the task boundary must stop.

## 5. Default Mismatch Rule

When truth layers disagree, the default agent behavior is:

1. detect mismatch
2. classify mismatch
3. continue only if the mismatch is truly local to the current task boundary
4. otherwise stop and create or prioritize the prerequisite work first

This rule prevents screen work from silently expanding into broad contract or startup repair work.

## 6. Hard Gates

Only five hard gates are required. This is the full operating core.

### Gate A: Startup Gate

Before feature or screen work:

- backend is built from current source
- startup happens via canonical script
- the script guarantees freshness or fails loudly
- auth/session bootstrap works
- the target route responds

If this gate fails, classify the issue as a runtime prerequisite.

### Gate B: Contract Gate

Before screen work or module API work:

- compare runtime route behavior
- compare OpenAPI
- compare generated backend surfaces
- compare generated frontend types
- compare feature API wrapper shape

If the mismatch affects more than the local task, classify it as a shared contract prerequisite and stop.

### Gate C: Screen Gate

Before any screen implementation:

- Startup Gate passes
- Contract Gate passes for the target module
- only then may screen audit and screen implementation begin

This gate becomes a mandatory pre-screen checkpoint.

### Gate D: Wiki Sync Gate

Before claiming module docs are synced:

- name the exact change context
- list all affected modules
- inspect all affected surfaces
- classify the sync as lite patch, structural refresh, or full rebuild
- run preflight/tally gates
- explicitly report any touched module that was not updated and why

This gate prevents partial documentation updates from pretending to be complete syncs.

### Gate E: Prerequisite Exit Gate

Before moving from prerequisite repair back into feature work:

- root cause is written
- fix scope is bounded
- the exact failed checkpoint is rerun and now passes
- no hidden drift remains inside that repaired boundary
- if the incident exposed a workflow gap, the relevant skill/runbook/instruction is updated before proceeding

This closes the loop between fixing code and hardening process.

## 7. Critical Contradiction Stop Rule

Agents do not need to stop for every stale sentence in the wiki.

They **must stop** when contradictions affect:

- route ownership or route prefix
- plan completion or prerequisite status
- startup/run instructions
- module ownership boundary
- API contract expectations
- verification expectations

These contradictions are considered unsafe because they change what work is allowed to proceed.

## 8. Script-Truth Startup Policy

Local development startup uses a strict policy:

- canonical scripts are the supported entrypoint
- ad hoc startup commands are not authoritative
- scripts must rebuild or explicitly prove freshness
- scripts must fail loudly on stale binary, blocked port, missing dependency, or broken prerequisite

The current startup model is too permissive because it allows stale binaries and ambiguous runtime evidence.

This policy must be enforced in both instructions and scripts.

## 9. Repo Surfaces To Harden

This operating model should be implemented by updating only the surfaces that matter.

### 9.1 Repo instructions

- `AGENTS.md`
- `CLAUDE.md`

These should define:

- truth hierarchy
- classification model
- critical contradiction stop rule
- mandatory gates
- script-truth startup policy

### 9.2 Core skills

Update only these skills:

- `metaldocs-screen-implementation`
- `metaldocs-backend-api`
- `metaldocs-frontend`
- `metaldocs-tanstack-query`
- `metaldocs-module-doc-sync`

Add at most one small new skill for shared prerequisite audit, such as:

- `post-rename-sync-audit`
- or `runtime-contract-prereq`

No skill explosion is allowed.

### 9.3 Scripts and checks

Improve only the essential executable surfaces:

- `scripts/start-api.ps1`
- one startup/auth/route preflight check
- one contract-sync check if necessary

### 9.4 Wiki/process docs

Add:

- one operating-model doc
- one runnable-checkpoint runbook

No large process library is needed.

## 10. Expected Skill Changes

### 10.1 `metaldocs-screen-implementation`

Add a mandatory pre-screen checkpoint before screen audit:

- fresh build truth
- runnable truth
- auth/session truth
- target route truth
- contract truth

If any checkpoint fails:

- classify the issue
- stop screen implementation unless the issue is strictly screen-local

### 10.2 `metaldocs-backend-api`

Strengthen route-truth and contract-truth enforcement:

- build route truth from runtime before editing
- compare runtime vs OpenAPI vs generated code vs frontend impact
- explicitly classify drift as local or shared prerequisite
- stop when contract drift exceeds the task boundary

### 10.3 `metaldocs-frontend`

Add stronger routing to prerequisite work:

- if UI/API work discovers startup, auth, or shared contract drift, stop and classify
- do not let feature work absorb general infrastructure repair by default

### 10.4 `metaldocs-tanstack-query`

Require contract ownership discipline:

- generated API types or canonical client must be authoritative for the module in scope
- if wrapper behavior diverges from runtime/spec and impacts more than the target task, classify as shared prerequisite

### 10.5 `metaldocs-module-doc-sync`

Turn it into a governed sync workflow rather than a best-effort updater.

Success requires:

- named change context
- explicit affected-module set
- explicit affected-surface scan
- explicit mode selection
- explicit skipped-module reporting
- tally/preflight result

No silent omissions.

## 11. New Missing Workflow

MetalDocs needs one dedicated prerequisite workflow for post-refactor drift:

- startup repair
- migration ledger repair
- runtime/contract/frontend-wrapper reconciliation
- wiki/process follow-through

This workflow should exist so screen work does not become the accidental place where shared repo drift is discovered and fixed ad hoc.

## 12. Rollout Plan

### Phase 1: Write the operating model

- publish this design
- agree on vocabulary and gates

### Phase 2: Harden repo instructions and skills

- update `AGENTS.md`
- update `CLAUDE.md`
- update the five core skills
- add at most one new prerequisite audit skill

### Phase 3: Harden startup and checks

- fix `start-api.ps1`
- add minimal preflight checks for startup/auth/route truth
- add minimal contract-sync checks where needed

### Phase 4: Apply the new workflow to current prerequisite issues

Use the hardened operating model on:

- migration/startup repair
- auth/runtime truth repair
- templates runtime/spec/frontend wrapper sync
- wiki drift exposed by these repairs

### Phase 5: Return to Plan 12 screen work

Only after the relevant gates pass:

- resume screen-by-screen work
- keep each screen PR narrow
- stop immediately when a shared prerequisite is encountered

## 13. Non-Goals

This operating model is intentionally not:

- a giant enterprise process framework
- a large CI governance program
- a release-management redesign
- a broad rewrite of every skill in the repo
- a mandate to fix all modules before any feature work forever

It is a compact professional operating system for keeping agentic development reliable.

## 14. Success Criteria

The operating model is successful when:

- screen work no longer starts on an unrunnable local system
- stale binaries stop producing misleading runtime evidence
- route/runtime/spec/codegen/frontend drift is classified before implementation
- module wiki sync cannot silently omit affected modules or surfaces
- workflow gaps are upgraded into repo rules instead of recurring surprises
- agents stop earlier, more correctly, and with better boundary discipline

## 15. Immediate Next Step

Do not jump back into Plan 12 directly.

Implement this operating model first, then use it to execute the current prerequisite repair work, then return to screen implementation under the new gates.
