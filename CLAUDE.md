# MetalDocs — Agent Operating Instructions

> Read order for any non-trivial task: this file → `wiki/README.md` → `wiki/references/current-agent-handoff.md` → relevant skill SKILL.md.

## 1. Local Dev Startup (script-truth policy)

**Always use the PowerShell script — never bash, never `source .env`:**

```powershell
.\scripts\start-api.ps1          # start API on :8081
.\scripts\start-api.ps1 -Build   # rebuild binary first
```

- Canonical scripts are the only supported entrypoint.
- Scripts must rebuild or explicitly prove freshness.
- Scripts must fail loudly on stale binary, blocked port, missing dependency, broken prerequisite.
- Ad hoc startup commands are not authoritative.

Login (dev): `POST /api/v1/auth/login` body `{"identifier":"admin","password":"AdminMetalDocs123!"}`.

Full details: [`wiki/references/local-dev-startup.md`](wiki/references/local-dev-startup.md).

## 2. Wiki Is Source of Truth

Project knowledge lives in `wiki/`. **Always read `wiki/README.md` first** — it indexes every doc with `file:line` anchors. Do not re-grep the codebase when the wiki already says where things are.

Critical entry points:
- [`wiki/concepts/placeholders.md`](wiki/concepts/placeholders.md) — eigenpal native vs MetalDocs legacy gap (fixed 7-token catalog)
- [`wiki/concepts/authz-tiers.md`](wiki/concepts/authz-tiers.md) — two-tier authz + Postgres tripwire
- [`wiki/modules/editor-ui-eigenpal.md`](wiki/modules/editor-ui-eigenpal.md) — eigenpal Anti-Corruption Layer
- [`wiki/decisions/`](wiki/decisions/) — ADRs (token migration, atomic CD create, contract-first API, etc.)

**Drift policy:** when you change code referenced by a wiki doc, update its `Last verified:` stamp same change.

**After refactors / new implementations** → dispatch the `wiki-curator` agent ([`.claude/agents/wiki-curator.md`](.claude/agents/wiki-curator.md)). It refreshes Key files anchors, bumps stamps, updates `wiki/README.md` index, creates new docs when a new module/concept/workflow appears. Invoke proactively — do not let drift accumulate.

For full module documentation, maturity promotion, or rebuilding a module wiki trio/artifacts → [`metaldocs-module-doc`](.agents/skills/metaldocs-module-doc/SKILL.md). After implementation work touches an already-documented module → [`metaldocs-module-doc-sync`](.agents/skills/metaldocs-module-doc-sync/SKILL.md).

## 3. Skill Routing (mandatory)

| Boundary you touch | Skill |
|--------------------|-------|
| Backend HTTP routes / OpenAPI / oapi-codegen / handler wiring / contract shape | [`metaldocs-backend-api`](.claude/skills/metaldocs-backend-api/SKILL.md) |
| Frontend under `frontend/apps/web/src/` (screens, components, routing, state, API wiring) | [`metaldocs-frontend`](.claude/skills/metaldocs-frontend/SKILL.md) |
| FE API calls / TanStack Query hooks / query keys / cache invalidation / generated FE types | [`metaldocs-tanstack-query`](.agents/skills/metaldocs-tanstack-query/SKILL.md) |
| Designed screens under `frontend/apps/web/design-source/<slug>/` | `metaldocs-frontend` + [`metaldocs-screen-implementation`](.claude/skills/metaldocs-screen-implementation/SKILL.md) |
| Real-capability mapping for mock-era widgets, legacy wrappers, deferred items | [`metaldocs-screen-integration-audit`](.claude/skills/metaldocs-screen-integration-audit/SKILL.md) |
| DB migrations / bootstrap / curated baseline / seeds / dictionary / extensions / grants / triggers | [`metaldocs-database`](.claude/skills/metaldocs-database/SKILL.md) |
| Runtime/auth/route/contract drift, prerequisite repair | [`runtime-contract-prereq`](.claude/skills/runtime-contract-prereq/SKILL.md) |
| Module wiki rebuild / sync | `metaldocs-module-doc` / `-sync` |

### Frontend rules (non-negotiable)
Canonical structure: [`wiki/architecture/frontend-structure.md`](wiki/architecture/frontend-structure.md).
- Feature-sliced layout, `createBrowserRouter` + per-feature `routes.tsx`
- TanStack Query for server state, OpenAPI-codegen types from `lib/api-types/`
- CSS Modules + design tokens
- **No** `HashRouter`, string-pattern path dispatchers, legacy `src/api/`, root flat files
- **Never reintroduce legacy paths.** If you touch a file outside the canonical layout, migrate it in the same change. No shims. No re-exports.

### Backend/API rules
Canonical structure: [`wiki/architecture/backend-api-structure.md`](wiki/architecture/backend-api-structure.md). Contracts: [`api-contract.md`](wiki/architecture/api-contract.md), [`api-design-system.md`](wiki/architecture/api-design-system.md).
- Do not change public routes, generated `api.gen.go` wiring, or OpenAPI shape from memory.
- Build the route truth table first → compare runtime / spec / codegen / wiki → implement from the canonical module pattern.

### Database rules
Source of truth: [`wiki/database/`](wiki/database/) (schema ownership, dictionary, migration policy, reference data, bootstrap rules). Do not duplicate those rules here.

## 4. Mandatory Gates

Canonical design: [`docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md`](docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md).
Operator guide: [`wiki/references/ai-operating-system.md`](wiki/references/ai-operating-system.md).

**Before screen work:**
1. Fresh build truth
2. Runnable truth
3. Auth/session truth
4. Target route truth
5. Contract truth

**Before claiming module wiki sync success:**
1. Named change context
2. Explicit affected modules
3. Explicit affected surfaces
4. Mode classification: lite patch / structural refresh / full rebuild
5. Preflight/tally result
6. Explicit skipped-module reporting

**Before resuming feature work after prerequisite repair:**
1. Root cause written
2. Fix scope bounded
3. Failed checkpoint rerun and passing
4. No hidden drift left in the repaired boundary
5. Skill/runbook/instruction updated if the failure exposed a workflow gap

**Stop rule:** do not continue feature work through a failing prerequisite boundary. If startup, auth/session, target route, or shared contract truth fails → switch to `runtime-contract-prereq`, repair the boundary, rerun the checkpoint, return to original task.

## 5. Behavioral Guidelines

> Bias toward caution over speed. For trivial tasks, use judgment.

### 5.1 Think before coding
Don't assume. Don't hide confusion. Surface tradeoffs.
- State assumptions explicitly. If uncertain, ask.
- Multiple interpretations? Present them — don't pick silently.
- Simpler approach exists? Say so. Push back when warranted.
- Unclear? Stop. Name what's confusing. Ask.

### 5.2 Simplicity first
Minimum code that solves the problem. Nothing speculative.
- No features beyond what was asked.
- No abstractions for single-use code.
- No flexibility / configurability that wasn't requested.
- No error handling for impossible scenarios.
- If you wrote 200 lines and it could be 50, rewrite it.
- Senior engineer test: would they call this overcomplicated? If yes, simplify.

### 5.3 Surgical changes
Touch only what you must. Clean up only your own mess.
- Don't improve adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style even if you'd do it differently.
- Notice unrelated dead code? Mention it — don't delete it.
- Orphans from your changes: remove them. Pre-existing dead code: leave it unless asked.
- Test: every changed line traces directly to the user's request.

### 5.4 Goal-driven execution
Define success criteria. Loop until verified.
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**Working signal:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, clarifying questions arrive before implementation rather than after mistakes.
