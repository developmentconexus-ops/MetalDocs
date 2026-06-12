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
| Backend HTTP routes / OpenAPI / oapi-codegen / handler wiring / contract shape | [`metaldocs-backend-api`](.agents/skills/metaldocs-backend-api/SKILL.md) |
| Frontend under `frontend/apps/web/src/` (screens, components, routing, state, API wiring) | [`metaldocs-frontend`](.agents/skills/metaldocs-frontend/SKILL.md) |
| FE API calls / TanStack Query hooks / query keys / cache invalidation / generated FE types | [`metaldocs-tanstack-query`](.agents/skills/metaldocs-tanstack-query/SKILL.md) |
| Designed screens under `frontend/apps/web/design-source/<slug>/` | `metaldocs-frontend` + [`metaldocs-screen-implementation`](.agents/skills/metaldocs-screen-implementation/SKILL.md) |
| Real-capability mapping for mock-era widgets, legacy wrappers, deferred items | [`metaldocs-screen-integration-audit`](.agents/skills/metaldocs-screen-integration-audit/SKILL.md) |
| DB migrations / bootstrap / curated baseline / seeds / dictionary / extensions / grants / triggers | [`metaldocs-database`](.agents/skills/metaldocs-database/SKILL.md) |
| Runtime/auth/route/contract drift, prerequisite repair | [`runtime-contract-prereq`](.agents/skills/runtime-contract-prereq/SKILL.md) |
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
Canonical QA/close-out policy: [`wiki/quality/qa-operating-system.md`](wiki/quality/qa-operating-system.md).
Path-stable operator bridge: [`wiki/references/ai-operating-system.md`](wiki/references/ai-operating-system.md).

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

**Hard-stop rule:** do not continue through redesign-grade issues. If the required fix implies shared API redesign, cross-module auth/authz model change, storage/provider architecture redesign, workflow semantic redesign outside the assigned boundary, or large coordinated rewrite, stop and report the architecture boundary and minimum prerequisite plan instead of symptom patching.

**Stop rule:** do not continue feature work through a failing prerequisite boundary. If startup, auth/session, target route, or shared contract truth fails → switch to `runtime-contract-prereq`, repair the boundary, rerun the checkpoint, return to original task.

**Default close-out loop for non-trivial work:**
1. Implement inside the bounded task
2. Run static and targeted verification for the touched slice
3. Run code review
4. Run product QA using the canonical checklist for the workflow class
5. Classify findings by root-cause family
6. Fix by family, not by scattered symptom patching
7. Rerun targeted review, QA, and regression
8. Rerun broader regression when the change crossed boundaries
9. Close only with evidence and explicit bounded defers

**Evidence rule:** `implemented`, `fixed`, `done`, `green`, or `looks good` are not sufficient closure by default. Record the verification commands, QA outcomes, review findings disposition, and any remaining bounded defer before claiming completion.

**Canonical QA checklists:**
- [`wiki/quality/screen-qa-checklist.md`](wiki/quality/screen-qa-checklist.md)
- [`wiki/quality/backend-api-qa-checklist.md`](wiki/quality/backend-api-qa-checklist.md)
- [`wiki/quality/workflow-async-qa-checklist.md`](wiki/quality/workflow-async-qa-checklist.md)
- [`wiki/quality/release-closeout-checklist.md`](wiki/quality/release-closeout-checklist.md)

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

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **MetalDocs** (30916 symbols, 68804 relationships, 300 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "main"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/MetalDocs/context` | Codebase overview, check index freshness |
| `gitnexus://repo/MetalDocs/clusters` | All functional areas |
| `gitnexus://repo/MetalDocs/processes` | All execution flows |
| `gitnexus://repo/MetalDocs/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
