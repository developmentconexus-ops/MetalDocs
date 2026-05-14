# MetalDocs AI Operating System

> **Last verified:** 2026-05-13
> **Scope:** Compact operating model for safe agentic work in MetalDocs.
> **Out of scope:** Detailed module docs, API contract mechanics, and startup script implementation details.
> **Key files:**
> - `docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md` - canonical design for the operating model
> - `scripts/start-api.ps1` - startup script truth boundary
> - `scripts/check-system-runnable.ps1` - runnable preflight checkpoint

## What this is

This is a compact workflow for keeping MetalDocs reliable during agentic development. It tells us which truth to trust for each kind of question, how to classify drift before we absorb it into feature work, and where we must stop instead of guessing.

## The four truths

- `runtime truth` - what actually runs now: mounted routes, auth/session behavior, DB schema, and live handler wiring.
- `contract truth` - what shared consumers should rely on: OpenAPI, generated backend surfaces, and generated frontend API types.
- `wiki truth` - governed technical memory: module docs, debt registers, backlogs, ADRs, and route truth tables.
- `execution truth` - what must be run before work is trusted: scripts, preflight checks, verification commands, and skill gates.

A common failure mode is treating wiki truth as if it proves runtime truth. It does not. Wiki memory helps us work faster, but runnable evidence wins when they disagree.

## The seven classifications

- `runtime prerequisite` - startup, migration, auth/session, binary freshness, or dependency drift that makes local runtime untrustworthy.
- `shared contract prerequisite` - runtime, spec, generated code, or frontend wrapper drift that affects more than the task in front of us.
- `module-local implementation` - a bounded change that stays inside one module contract and does not change shared expectations.
- `screen-local implementation` - a bounded screen change that does not require shared runtime or contract repair.
- `wiki-memory drift` - docs or governed memory need updating after code truth changed.
- `workflow/tooling gap` - the failure exposed a missing or weak script, skill gate, or verification rule.
- `defer` - the task found a larger product or architecture gap that should be captured, not silently implemented.

## The five hard gates

- `Startup Gate` - start from canonical scripts, prove fresh build truth, and do not trust ad hoc startup commands.
- `Contract Gate` - compare runtime behavior, OpenAPI, generated backend surfaces, generated frontend types, and feature wrappers before changing shared behavior.
- `Screen Gate` - do not begin screen work until startup, auth/session, target route, and contract truth are all trustworthy.
- `Wiki Sync Gate` - no silent omissions; every affected module must be updated or explicitly skipped with a reason.
- `Prerequisite Exit Gate` - after a prerequisite repair, rerun the failed checkpoint and update workflow guidance if the incident exposed a gap.

## How to start work safely

1. Read the relevant wiki docs and the required skill guidance for the area you are touching.
2. Treat startup scripts as authoritative. This is the repo's `script-truth` policy: canonical scripts are supported, stale binaries are not trusted, and script output beats remembered commands.
3. Run the runnable preflight before screen work. A screen task starts only after login, session, and target route checks succeed.
4. Compare runtime truth and contract truth early if the task touches an HTTP surface or frontend API wrapper.
5. Stop on critical contradictions such as route ownership drift, startup instruction drift, or conflicting contract expectations.

Example: a stale binary can make route evidence lie. If an old `metaldocs-api.exe` still answers on `:8081`, you might think a route rename failed or a handler still exists. In reality, you may just be looking at yesterday's binary. Under the operating system, that is a `runtime prerequisite`: restart from the canonical script, rebuild or prove freshness, then re-check the route.

## How to know when to stop

Keep going only when the mismatch is local to the current task boundary. Stop when the mismatch changes shared runtime or contract behavior.

Example: runtime route exists but the frontend wrapper and generated types still reflect an older contract. Even if the backend endpoint responds, this is not a screen-local fix. The runtime/spec/frontend mismatch is a `shared contract prerequisite` because other callers could be wrong too. Stop the screen task, repair the shared contract surfaces, then resume feature work.

Example: during a screen task, you discover the backend endpoint the design expects does not exist at all. Do not stub around it and pretend the screen is done. Classify it as a prerequisite or `defer`, capture the missing backend work, and stop the screen implementation unless the assignment explicitly includes that backend slice.

## How wiki sync works now

Wiki sync is no longer best-effort. No silent omissions are allowed.

Every sync must name the exact change context, list every affected module, and either update each one or explicitly say why it was skipped. This matters most for cross-cutting changes.

Example: if a contract change touches `templates` but also changes a shared route prefix or a dependency flow consumed by `documents`, module-doc sync must update more than one affected module. A sync that edits only `templates.md` and leaves `documents.md` stale is incomplete even if the original code change started in one module.

## How to resume feature work after a prerequisite repair

1. Write the root cause in plain language.
2. Keep the repair boundary bounded. Do not quietly roll unrelated cleanup into the prerequisite fix.
3. Rerun the exact failed checkpoint. If startup drift blocked the task, rerun startup and runnable preflight. If contract drift blocked it, rerun the contract comparison.
4. Confirm there is no hidden drift left in that repaired boundary.
5. If the incident exposed a workflow gap, update the script, skill, or runbook before resuming feature work.

Once the failed checkpoint passes again, re-enter the original task at the same boundary instead of reopening discovery from scratch.
