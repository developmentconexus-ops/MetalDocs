## Local dev startup

**Always use the PowerShell script — never bash/source .env:**
```powershell
.\scripts\start-api.ps1        # starts API on :8081
.\scripts\start-api.ps1 -Build # rebuild binary first
```
Login: `POST /api/v1/auth/login` with body `{"identifier":"admin","password":"AdminMetalDocs123!"}`.
Full details: `wiki/references/local-dev-startup.md`

## Wiki

Project knowledge lives in `wiki/`. **Always read `wiki/README.md` first** for any non-trivial task — it indexes every doc with file:line anchors. Skip re-grepping the codebase when the wiki already says where things are.

Critical entry points:
- `wiki/concepts/placeholders.md` — eigenpal native vs MetalDocs legacy gap
- `wiki/modules/editor-ui-eigenpal.md` — how MetalDocs wraps eigenpal
- `wiki/decisions/` — ADRs (zone purge, token migration, etc.)

When you change code referenced by a wiki doc, update its `Last verified:` stamp.

**After refactors / new implementations, dispatch the `wiki-curator` agent** (`.claude/agents/wiki-curator.md`). It owns wiki drift: refreshes Key files anchors, bumps `Last verified` stamps, updates `wiki/README.md` index, and creates new docs when a new module/concept/workflow is introduced. Invoke proactively — do not let wiki drift accumulate.

For full module documentation, maturity promotion, or rebuilding a module wiki trio/artifacts, use **`metaldocs-module-doc`** (`.agents/skills/metaldocs-module-doc/SKILL.md`). After implementation work touches an already documented module, use **`metaldocs-module-doc-sync`** (`.agents/skills/metaldocs-module-doc-sync/SKILL.md`) to update module docs, tech-debt, backlog, route truth, artifacts, and sync logs from the concrete change context.

## Frontend

For ANY work under `frontend/apps/web/src/` (new screens, components, refactors, design implementation, routing, state, API wiring), use the **`metaldocs-frontend`** skill (`.claude/skills/metaldocs-frontend/SKILL.md`). It enforces the canonical structure defined in `wiki/architecture/frontend-structure.md` — feature-sliced layout, `createBrowserRouter` + per-feature `routes.tsx`, TanStack Query for server state, OpenAPI-codegen types from `lib/api-types/`, CSS Modules + design tokens, no `HashRouter`, no string-pattern path dispatchers, no legacy `src/api/` or root flat files. **Never reintroduce legacy paths.** When you touch a file outside the canonical layout, migrate it in the same change (no shims, no re-exports).

For frontend API calls, TanStack Query hooks, query keys, cache invalidation, optimistic updates, generated frontend API types, or server-state performance, also use **`metaldocs-tanstack-query`** (`.agents/skills/metaldocs-tanstack-query/SKILL.md`). It is the focused workflow for making API/UI integration fast, correct, and maintainable.

Designed screens land in `frontend/apps/web/design-source/<slug>/` (HTML + screenshot + NOTES.md). For any task that says "implement screen X" or references a `design-source/<slug>/` directory, ALSO use the **`metaldocs-screen-implementation`** skill (`.claude/skills/metaldocs-screen-implementation/SKILL.md`) on top of `metaldocs-frontend`. It drives a 6-phase workflow (Audit → Map → Pre-flight → Page assembly → Verify → Document) with hard gates that captures lessons from the Library screen rollout.

Before real screen finalization, use **`metaldocs-screen-integration-audit`** (`.claude/skills/metaldocs-screen-integration-audit/SKILL.md`) when the screen/backlog may include mock-era widgets, missing backend capability, legacy API wrappers, deferred items, or uncertainty about what maps to real product behavior. It classifies what can be implemented now and what must become prerequisite or deferred work.

## Backend/API

For ANY work on MetalDocs backend HTTP routes, OpenAPI, oapi-codegen, handler wiring, API contracts, route migrations, or generated frontend API types, use the **`metaldocs-backend-api`** skill (`.claude/skills/metaldocs-backend-api/SKILL.md`). It enforces the canonical backend/API structure defined in `wiki/architecture/backend-api-structure.md` and the behavior contracts in `wiki/architecture/api-contract.md` and `wiki/architecture/api-design-system.md`.

Do not change public backend routes, generated `api.gen.go` wiring, or OpenAPI contract shape from memory. Build the route truth table first, compare runtime/spec/codegen/wiki, then implement from the canonical module pattern.

## Database

For ANY work on MetalDocs database migrations, bootstrap, curated baseline, reference data, dev seeds, schema ownership, Postgres extensions, grants, triggers, functions, `schema_migrations`, or database dictionary/wiki pages, use the **`metaldocs-database`** skill (`.claude/skills/metaldocs-database/SKILL.md`).

The database wiki under `wiki/database/` is the source of truth for schema ownership, dictionary entries, migration policy, reference data, and bootstrap rules. Do not duplicate those rules here.

## Mandatory Gates

Canonical design: `docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md`.
User/operator guide: `wiki/references/ai-operating-system.md`.

Workflow selection quick map:
- backend/API boundary -> `metaldocs-backend-api`
- frontend boundary -> `metaldocs-frontend`
- frontend API/query boundary -> `metaldocs-tanstack-query`
- designed screen boundary -> `metaldocs-frontend` + `metaldocs-screen-integration-audit` when real capability mapping is needed + `metaldocs-screen-implementation`
- module wiki rebuild -> `metaldocs-module-doc`
- module wiki sync after implementation -> `metaldocs-module-doc-sync`
- runtime/auth/route/contract drift -> `runtime-contract-prereq`
- database migrations / bootstrap / curated baseline / seeds / dictionary -> `metaldocs-database`

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

Startup uses script-truth policy:
- canonical scripts are the supported entrypoint
- ad hoc startup commands are not authoritative
- scripts must rebuild or explicitly prove freshness
- scripts must fail loudly on stale binary, blocked port, missing dependency, or broken prerequisite

Do not continue feature work through a failing prerequisite boundary.
If startup, auth/session, target route, or shared contract truth fails, switch to `runtime-contract-prereq`, repair the failing boundary, rerun the checkpoint, and only then return to the original task.

---

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

Tradeoff: These guidelines bias toward caution over speed. For trivial tasks, use judgment.

1. Think Before Coding
Don't assume. Don't hide confusion. Surface tradeoffs.

Before implementing:

State your assumptions explicitly. If uncertain, ask.
If multiple interpretations exist, present them - don't pick silently.
If a simpler approach exists, say so. Push back when warranted.
If something is unclear, stop. Name what's confusing. Ask.
2. Simplicity First
Minimum code that solves the problem. Nothing speculative.

No features beyond what was asked.
No abstractions for single-use code.
No "flexibility" or "configurability" that wasn't requested.
No error handling for impossible scenarios.
If you write 200 lines and it could be 50, rewrite it.
Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

3. Surgical Changes
Touch only what you must. Clean up only your own mess.

When editing existing code:

Don't "improve" adjacent code, comments, or formatting.
Don't refactor things that aren't broken.
Match existing style, even if you'd do it differently.
If you notice unrelated dead code, mention it - don't delete it.
When your changes create orphans:

Remove imports/variables/functions that YOUR changes made unused.
Don't remove pre-existing dead code unless asked.
The test: Every changed line should trace directly to the user's request.

4. Goal-Driven Execution
Define success criteria. Loop until verified.

Transform tasks into verifiable goals:

"Add validation" → "Write tests for invalid inputs, then make them pass"
"Fix the bug" → "Write a test that reproduces it, then make it pass"
"Refactor X" → "Ensure tests pass before and after"
For multi-step tasks, state a brief plan:

1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

These guidelines are working if: fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.
