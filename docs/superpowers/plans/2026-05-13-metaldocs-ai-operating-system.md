# MetalDocs AI Operating System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the MetalDocs AI operating system so agentic work follows hard truth layers, mandatory gates, startup discipline, contract classification, and governed wiki sync before feature work resumes.

**Architecture:** This rollout hardens four repo surfaces in parallel: repo instructions, core skills, executable startup/preflight checks, and user-facing wiki guidance. The plan keeps process strict but small: one new prerequisite-audit skill at most, one startup preflight script, one contract-sync preflight script, and one full wiki guide that explains how to work inside the new operating system.

**Tech Stack:** Markdown docs, PowerShell startup/preflight scripts, existing Codex/Claude skill markdown, Git, Go API startup flow, frontend OpenAPI/TanStack architecture.

---

## File Structure

### Existing files to modify

- `AGENTS.md`
  - repo-level hard rules for layered truth, classification, contradiction stop rules, and mandatory gates
- `CLAUDE.md`
  - local operating rules for startup, wiki sync, frontend/backend stop conditions, and prerequisite exit behavior
- `.claude/skills/metaldocs-screen-implementation/SKILL.md`
  - add pre-screen runnable checkpoint gate
- `.claude/skills/metaldocs-backend-api/SKILL.md`
  - add shared-contract prerequisite classification and stop rules
- `.claude/skills/metaldocs-frontend/SKILL.md`
  - add prerequisite classification/stop behavior for frontend discovery work
- `.agents/skills/metaldocs-tanstack-query/SKILL.md`
  - add generated-type authority and shared-contract stop rule
- `.claude/skills/metaldocs-module-doc-sync/SKILL.md`
  - strengthen sync-accountability gate and no-silent-omissions rule
- `.agents/skills/metaldocs-module-doc-sync/SKILL.md`
  - keep bridge text aligned with canonical sync workflow
- `scripts/start-api.ps1`
  - enforce script-truth startup behavior and binary freshness
- `wiki/README.md`
  - index the new operating-system user guide and related runbooks
- `wiki/references/README.md`
  - add the new operator guide to reference index
- `wiki/references/local-dev-startup.md`
  - align with script-truth startup policy and preflight expectations

### New files to create

- `.claude/skills/runtime-contract-prereq/SKILL.md`
  - one compact prerequisite-audit workflow for startup/contract/runtime drift
- `.agents/skills/runtime-contract-prereq/SKILL.md`
  - Codex-discoverable bridge to the canonical new skill
- `scripts/check-system-runnable.ps1`
  - startup/auth/target-route preflight check
- `scripts/check-module-contract-sync.ps1`
  - runtime/spec/generated/frontend-wrapper alignment check for a target module
- `wiki/references/ai-operating-system.md`
  - user-facing full guide for how to work in the new operating system

### Existing files to leave untouched in this plan

- `migrations/0189`-`0197`
  - these remain a separate runtime prerequisite repair stream and must not be mixed into the operating-system rollout commits

## Task 1: Create The Operating-System Implementation Plan Artifact

**Files:**
- Create: `docs/superpowers/plans/2026-05-13-metaldocs-ai-operating-system.md`

- [ ] **Step 1: Save this implementation plan**

Create `docs/superpowers/plans/2026-05-13-metaldocs-ai-operating-system.md` with the full contents of this plan.

- [ ] **Step 2: Review the plan file exists and is readable**

Run: `Get-Content -Raw docs\superpowers\plans\2026-05-13-metaldocs-ai-operating-system.md`
Expected: full plan text prints with no placeholder markers like `TODO`, `TBD`, or `fill later`.

- [ ] **Step 3: Commit the plan artifact**

Run:
```powershell
git add -- docs/superpowers/plans/2026-05-13-metaldocs-ai-operating-system.md
git commit -m "docs(process): add AI operating system implementation plan"
```
Expected: one commit containing only the plan file.

## Task 2: Harden Repo Instructions

**Files:**
- Modify: `AGENTS.md`
- Modify: `CLAUDE.md`
- Test: repo text review via `rg`

- [ ] **Step 1: Add operating-system rules to `AGENTS.md`**

Insert a compact section near the top-level behavioral rules with these exact concepts:

```md
## MetalDocs AI Operating System

Use the MetalDocs operating model for all non-trivial work.

Truth layers:
- Runtime truth: what actually runs now.
- Contract truth: OpenAPI, generated backend surfaces, generated frontend API types.
- Wiki truth: module docs, debt, backlog, roadmap, ADRs.
- Execution truth: scripts, preflight checks, verification steps, skill gates.

Required classification:
- runtime prerequisite
- shared contract prerequisite
- module-local implementation
- screen-local implementation
- wiki-memory drift
- workflow/tooling gap
- defer

Default mismatch rule:
- detect mismatch
- classify mismatch
- continue only if local to the current task boundary
- otherwise stop and surface the prerequisite work first

Critical contradiction stop rule:
Stop when contradictions affect route ownership/prefix, plan or prerequisite status, startup instructions, module ownership, API contract expectations, or verification expectations.
```

- [ ] **Step 2: Add operating-system enforcement to `CLAUDE.md`**

Append a small section under the startup/wiki/frontend/backend guidance with these exact gates:

```md
## Mandatory Gates

Before screen work:
1. Fresh build truth
2. Runnable truth
3. Auth/session truth
4. Target route truth
5. Contract truth

Before claiming module wiki sync success:
1. Named change context
2. Explicit affected modules
3. Explicit affected surfaces
4. Mode classification: lite patch / structural refresh / full rebuild
5. Preflight/tally result
6. Explicit skipped-module reporting

Before resuming feature work after prerequisite repair:
1. Root cause written
2. Fix scope bounded
3. Failed checkpoint rerun and passing
4. No hidden drift left in the repaired boundary
5. Skill/runbook/instruction updated if the failure exposed a workflow gap
```

- [ ] **Step 3: Add script-truth startup policy to `CLAUDE.md`**

Add this exact policy below the existing startup section:

```md
Startup uses script-truth policy:
- canonical scripts are the supported entrypoint
- ad hoc startup commands are not authoritative
- scripts must rebuild or explicitly prove freshness
- scripts must fail loudly on stale binary, blocked port, missing dependency, or broken prerequisite
```

- [ ] **Step 4: Verify the new instruction anchors exist**

Run:
```powershell
rg -n "MetalDocs AI Operating System|Required classification|Mandatory Gates|script-truth policy" AGENTS.md CLAUDE.md
```
Expected: all four phrases are found exactly once or in the expected new sections.

- [ ] **Step 5: Commit instruction hardening**

Run:
```powershell
git add -- AGENTS.md CLAUDE.md
git commit -m "docs(process): enforce operating-system rules in repo instructions"
```
Expected: commit contains only `AGENTS.md` and `CLAUDE.md`.

## Task 3: Harden Screen And API Skills

**Files:**
- Modify: `.claude/skills/metaldocs-screen-implementation/SKILL.md`
- Modify: `.claude/skills/metaldocs-backend-api/SKILL.md`
- Modify: `.claude/skills/metaldocs-frontend/SKILL.md`
- Modify: `.agents/skills/metaldocs-tanstack-query/SKILL.md`

- [ ] **Step 1: Add pre-screen checkpoint to screen skill**

In `.claude/skills/metaldocs-screen-implementation/SKILL.md`, add a new earliest gate before current audit/phase flow with this exact checklist:

```md
## Phase -1: Runnable Checkpoint

Before screen audit begins, verify:
1. Fresh build truth
2. Runnable truth
3. Auth/session truth
4. Target route truth
5. Contract truth

If any checkpoint fails:
- classify the issue
- stop screen work unless the issue is strictly screen-local
- route runtime or shared contract failures to the prerequisite workflow
```

- [ ] **Step 2: Add contract classification stop rule to backend skill**

In `.claude/skills/metaldocs-backend-api/SKILL.md`, add this rule in workflow or stop rules:

```md
When runtime, OpenAPI, generated code, frontend generated types, or frontend wrapper behavior disagree, classify the mismatch first.

Continue only if the mismatch is local to the current task boundary.
If it affects shared module contract behavior, stop and route it to a shared contract prerequisite.
```

- [ ] **Step 3: Add frontend prerequisite stop behavior**

In `.claude/skills/metaldocs-frontend/SKILL.md`, add this rule in the workflow section:

```md
If frontend work discovers startup drift, auth/session drift, or shared runtime/contract drift, do not absorb the repair silently into the feature task.
Classify it and stop unless it is strictly local to the current screen or module boundary.
```

- [ ] **Step 4: Add generated-type authority rule to TanStack skill**

In `.agents/skills/metaldocs-tanstack-query/SKILL.md`, add this rule near the API wrapper/stop rules:

```md
For the module in scope, generated API types and the canonical API client are authoritative unless the task is explicitly a prerequisite sync.
If wrapper behavior diverges from runtime/spec and affects more than the local task, classify it as a shared contract prerequisite and stop.
```

- [ ] **Step 5: Verify the new gate language is present**

Run:
```powershell
rg -n "Runnable Checkpoint|shared contract prerequisite|startup drift|generated API types and the canonical API client are authoritative" .claude/skills .agents/skills
```
Expected: each phrase is present in the intended skill files.

- [ ] **Step 6: Commit skill hardening A**

Run:
```powershell
git add -- .claude/skills/metaldocs-screen-implementation/SKILL.md .claude/skills/metaldocs-backend-api/SKILL.md .claude/skills/metaldocs-frontend/SKILL.md .agents/skills/metaldocs-tanstack-query/SKILL.md
git commit -m "docs(skills): add operating-system gates to screen and api workflows"
```
Expected: one commit containing only the four skill files.

## Task 4: Harden Module Wiki Sync Accountability

**Files:**
- Modify: `.claude/skills/metaldocs-module-doc-sync/SKILL.md`
- Modify: `.agents/skills/metaldocs-module-doc-sync/SKILL.md`

- [ ] **Step 1: Strengthen canonical sync success gate**

In `.claude/skills/metaldocs-module-doc-sync/SKILL.md`, add a dedicated section:

```md
## Sync Success Gate

A sync may claim success only when it includes:
1. Exact change context
2. Explicit affected-module list
3. Explicit affected-surface scan
4. Mode classification: lite patch, structural refresh, or full rebuild escalation
5. Preflight/tally result
6. Explicit explanation for every touched module that was not updated

No silent omissions are allowed.
```

- [ ] **Step 2: Strengthen cross-cutting module requirement**

In the workflow section, add:

```md
If a cross-cutting file changes, update every documented module whose route, contract, persistence, runtime flow, or dependency behavior changed.
Do not treat cross-cutting files as single-module edits by default.
```

- [ ] **Step 3: Align the `.agents` bridge copy**

Update `.agents/skills/metaldocs-module-doc-sync/SKILL.md` so the bridge summary explicitly says:

```md
Use this only when the sync can name the exact change context and every affected module. This workflow does not permit silent omissions.
```

- [ ] **Step 4: Verify sync-accountability phrases exist**

Run:
```powershell
rg -n "Sync Success Gate|No silent omissions|every documented module whose route, contract, persistence, runtime flow, or dependency behavior changed" .claude/skills/metaldocs-module-doc-sync .agents/skills/metaldocs-module-doc-sync
```
Expected: all phrases are present.

- [ ] **Step 5: Commit wiki-sync hardening**

Run:
```powershell
git add -- .claude/skills/metaldocs-module-doc-sync/SKILL.md .agents/skills/metaldocs-module-doc-sync/SKILL.md
git commit -m "docs(skills): harden module wiki sync accountability"
```
Expected: one commit containing only the sync skill files.

## Task 5: Add One Shared Prerequisite Audit Skill

**Files:**
- Create: `.claude/skills/runtime-contract-prereq/SKILL.md`
- Create: `.agents/skills/runtime-contract-prereq/SKILL.md`

- [ ] **Step 1: Create canonical prerequisite skill**

Create `.claude/skills/runtime-contract-prereq/SKILL.md` with this structure:

```md
---
name: runtime-contract-prereq
description: Use when startup, runtime wiring, migrations, route truth, OpenAPI truth, generated code truth, or frontend-wrapper truth may have drifted after a refactor or rename and feature work must stop until the system is trustworthy again.
---

# Runtime + Contract Prerequisite Audit

Use this before screen or feature work when local runtime truth is unreliable.

## Goal

Restore trust in:
- startup truth
- auth/session truth
- target route truth
- module contract truth
- workflow truth after the incident

## Required outputs

- issue classification
- route/runtime/spec/generated/frontend wrapper comparison
- prerequisite repair boundary
- exit-gate verification
- workflow gap updates required before resuming feature work

## Stop rules

Do not continue into feature work while the prerequisite boundary is still failing.
```

- [ ] **Step 2: Create Codex bridge for the new skill**

Create `.agents/skills/runtime-contract-prereq/SKILL.md` with this content:

```md
---
name: runtime-contract-prereq
description: Use when a MetalDocs task discovers startup drift, migration drift, auth/session failure, route mismatch, or runtime/OpenAPI/generated/frontend-wrapper mismatch that must be repaired before feature work continues.
---

# Runtime + Contract Prerequisite Audit

Read and follow `.claude/skills/runtime-contract-prereq/SKILL.md`.

This bridge exists so Codex sessions can discover the canonical prerequisite workflow.

Stop if feature work is trying to continue through a failing prerequisite boundary.
```

- [ ] **Step 3: Verify the new skill files are discoverable**

Run:
```powershell
Get-Content -Raw .claude\skills\runtime-contract-prereq\SKILL.md
Get-Content -Raw .agents\skills\runtime-contract-prereq\SKILL.md
```
Expected: both files load successfully.

- [ ] **Step 4: Commit the new prerequisite skill**

Run:
```powershell
git add -- .claude/skills/runtime-contract-prereq/SKILL.md .agents/skills/runtime-contract-prereq/SKILL.md
git commit -m "docs(skills): add runtime-contract prerequisite workflow"
```
Expected: one commit containing only the new skill files.

## Task 6: Harden Startup Script And Add Runnable Preflight

**Files:**
- Modify: `scripts/start-api.ps1`
- Create: `scripts/check-system-runnable.ps1`

- [ ] **Step 1: Make `start-api.ps1` enforce freshness**

Update `scripts/start-api.ps1` so it no longer trusts an old binary by default.
The script should:
- rebuild when `-Build` is passed
- otherwise compare source timestamps vs `metaldocs-api.exe`
- rebuild automatically if the binary is older than critical backend source/script files
- print a clear message when auto-rebuilding due to stale binary

Add a focused helper block like:

```powershell
$criticalPaths = @(
  'apps/api/cmd/metaldocs-api',
  'internal/modules',
  'internal/platform',
  'migrations',
  'scripts/start-api.ps1'
)
```

Then compute latest write time across those paths and rebuild if newer than the binary.

- [ ] **Step 2: Add blocked-port and startup-truth messaging**

When the script kills an existing process on `:8081`, make it print that the previous holder may have been stale and that a fresh process is being started from current source.

Expected message pattern:

```powershell
Write-Host "Detected existing process on :8081; replacing it with a fresh API process built from current source"
```

- [ ] **Step 3: Create `scripts/check-system-runnable.ps1`**

Create a preflight script that:
- optionally calls `start-api.ps1 -Build -NoWorker`
- verifies `/api/v1/auth/login` returns a non-network response
- verifies `/api/v1/auth/me` with session cookie after login
- verifies a caller-provided route, for example `/api/v1/templates`
- exits non-zero on failure and prints which checkpoint failed

Use parameters like:

```powershell
param(
  [string]$TargetRoute = '/api/v1/health/ready',
  [switch]$StartApi
)
```

- [ ] **Step 4: Verify runnable preflight script behavior**

Run:
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\check-system-runnable.ps1 -TargetRoute /api/v1/templates
```
Expected: explicit pass/fail lines for login, session, and target route.
If local auth is still broken, this script should fail loudly and identify that exact checkpoint.

- [ ] **Step 5: Commit startup hardening**

Run:
```powershell
git add -- scripts/start-api.ps1 scripts/check-system-runnable.ps1
git commit -m "chore(dev): harden startup truth and runnable preflight"
```
Expected: one commit containing only the startup and preflight scripts.

## Task 7: Add Contract-Sync Preflight Script

**Files:**
- Create: `scripts/check-module-contract-sync.ps1`

- [ ] **Step 1: Create module contract-sync checker**

Create a PowerShell script that accepts a module name and prints a compact reconciliation checklist for:
- runtime route ownership files
- OpenAPI path presence
- generated backend package presence
- generated frontend path/type presence
- feature API wrapper presence

Use parameters like:

```powershell
param(
  [Parameter(Mandatory = $true)]
  [string]$Module
)
```

For `templates`, it should inspect at minimum:
- `internal/modules/templates/delivery/http/handler.go`
- `api/openapi/v1/openapi.yaml`
- `internal/modules/templates/api/api.gen.go`
- `frontend/apps/web/src/lib/api-types/index.d.ts`
- `frontend/apps/web/src/features/templates/api/templatesV2.ts`

- [ ] **Step 2: Add explicit shared-prerequisite exit behavior**

If one or more surfaces are missing or obviously contradictory, the script must print:

```text
RESULT: shared contract prerequisite
```

If all required surfaces are present, print:

```text
RESULT: contract surfaces present; manual drift review still required
```

- [ ] **Step 3: Verify with templates module**

Run:
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\check-module-contract-sync.ps1 -Module templates
```
Expected: a readable report over the five target surfaces.

- [ ] **Step 4: Commit contract-sync preflight**

Run:
```powershell
git add -- scripts/check-module-contract-sync.ps1
git commit -m "chore(dev): add module contract-sync preflight"
```
Expected: one commit containing only the new preflight script.

## Task 8: Write The User-Facing Wiki Guide

**Files:**
- Create: `wiki/references/ai-operating-system.md`
- Modify: `wiki/README.md`
- Modify: `wiki/references/README.md`
- Modify: `wiki/references/local-dev-startup.md`

- [ ] **Step 1: Create the full wiki guide**

Create `wiki/references/ai-operating-system.md` with these exact sections:

```md
# MetalDocs AI Operating System

## What this is
- a compact workflow for keeping MetalDocs reliable during agentic development

## The four truths
- runtime truth
- contract truth
- wiki truth
- execution truth

## The seven classifications
- runtime prerequisite
- shared contract prerequisite
- module-local implementation
- screen-local implementation
- wiki-memory drift
- workflow/tooling gap
- defer

## The five hard gates
- Startup Gate
- Contract Gate
- Screen Gate
- Wiki Sync Gate
- Prerequisite Exit Gate

## How to start work safely
- read the relevant wiki docs
- use the required skills
- run the runnable checkpoint before screen work
- stop on critical contradictions

## How to know when to stop
- shared prerequisite vs local fix examples

## How wiki sync works now
- no silent omissions
- every affected module must be named or explicitly skipped with reason

## How to resume feature work after a prerequisite repair
- rerun the failed checkpoint
- update the workflow if the incident exposed a process gap
```

- [ ] **Step 2: Add concrete examples to the guide**

Add example scenarios in prose for:
- stale binary causing misleading route evidence
- runtime route exists but frontend wrapper/spec drift is shared prerequisite
- screen task discovering missing backend endpoint and deferring correctly
- module-doc sync needing to update more than one affected module

- [ ] **Step 3: Index the guide in `wiki/README.md` and `wiki/references/README.md`**

Add a new reference entry pointing to `wiki/references/ai-operating-system.md`.

Suggested entry text:

```md
- `references/ai-operating-system.md` - how to work inside the MetalDocs AI operating system: truth layers, gates, classifications, and stop rules.
```

- [ ] **Step 4: Align `local-dev-startup.md` with script-truth policy**

Add a short section to `wiki/references/local-dev-startup.md` stating:
- startup scripts are authoritative
- stale binaries are not trusted
- runnable preflight is required before screen work

- [ ] **Step 5: Verify wiki references**

Run:
```powershell
rg -n "ai-operating-system|script-truth|runnable preflight" wiki\README.md wiki\references\README.md wiki\references\local-dev-startup.md wiki\references\ai-operating-system.md
```
Expected: all new reference phrases are present.

- [ ] **Step 6: Commit the user-facing wiki guide**

Run:
```powershell
git add -- wiki/README.md wiki/references/README.md wiki/references/local-dev-startup.md wiki/references/ai-operating-system.md
git commit -m "docs(wiki): add AI operating system user guide"
```
Expected: one commit containing only the wiki guide and its index updates.

## Task 9: Validate The Operating System Surface End To End

**Files:**
- No new files required; uses all modified files above

- [ ] **Step 1: Run text-level validation over instructions, skills, scripts, and wiki**

Run:
```powershell
rg -n "runtime prerequisite|shared contract prerequisite|Screen Gate|Wiki Sync Gate|Prerequisite Exit Gate|No silent omissions|script-truth" AGENTS.md CLAUDE.md .claude/skills .agents/skills wiki scripts
```
Expected: the core operating-system vocabulary appears across the intended surfaces.

- [ ] **Step 2: Run the runnable preflight script once**

Run:
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\check-system-runnable.ps1 -TargetRoute /api/v1/templates
```
Expected: either a clean pass or a precise failure classification at auth/session/route checkpoint.

- [ ] **Step 3: Run the module contract-sync preflight once**

Run:
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts\check-module-contract-sync.ps1 -Module templates
```
Expected: report prints all five surfaces and a result classification.

- [ ] **Step 4: Review final diff for accidental mixing with migration repairs**

Run:
```powershell
git status --short
git diff -- AGENTS.md CLAUDE.md .claude/skills .agents/skills scripts wiki docs/superpowers/plans
```
Expected: only operating-system files are staged/committed in this stream; migration repair files remain separate.

- [ ] **Step 5: Commit validation cleanups if needed**

If validation required small wording or script fixes, commit them with:

```powershell
git add -- AGENTS.md CLAUDE.md .claude/skills .agents/skills scripts wiki docs/superpowers/plans
git commit -m "chore(process): finalize AI operating system rollout"
```

If no follow-up fix was needed, skip this commit.

## Task 10: Apply The New Workflow To The Current Prerequisite Stream

**Files:**
- No code change required in this task unless the new scripts or skills need tiny follow-up wording
- Uses existing uncommitted runtime prerequisite files as the first live consumer of the new operating system

- [ ] **Step 1: Classify the existing startup/migration/auth/templates issue set**

Record in the working notes or final handoff that the current blocked state includes at least:
- runtime prerequisite: migration ledger drift and startup trust
- shared contract prerequisite: templates runtime/spec/generated/frontend-wrapper drift
- workflow/tooling gap: startup script truth and module-doc-sync omission behavior

- [ ] **Step 2: State the next implementation boundary explicitly**

The next engineering task after operating-system rollout must be a prerequisite repair slice, not Plan 12 screen work.

Write this exact handoff note in the final response:

```text
Next execution boundary: use the new runtime-contract-prereq workflow to finish startup/auth/templates contract trust before resuming Plan 12 screen work.
```

- [ ] **Step 3: Do not mix the prerequisite repair commit stream yet**

Keep the runtime prerequisite files unstaged in this operating-system rollout unless the user explicitly asks to merge the streams.

## Task 11: Final Review And Handoff

**Files:**
- No new files required

- [ ] **Step 1: Run `git diff --check`**

Run:
```powershell
git diff --check
```
Expected: no whitespace errors in the operating-system changes.

- [ ] **Step 2: Summarize the rollout by surface**

Prepare final report grouped as:
- repo instructions
- skills
- scripts/checks
- wiki guide
- validation results
- remaining separate prerequisite stream

- [ ] **Step 3: Offer the next execution move**

The recommended next step after this rollout is:
1. run the new prerequisite workflow on the current startup/auth/templates issues
2. finish the prerequisite repair stream
3. only then return to Plan 12 screens