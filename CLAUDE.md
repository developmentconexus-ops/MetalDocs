# MetalDocs — Agent Operating Instructions

> Read order any non-trivial task: this file → `wiki/README.md` → `wiki/references/current-agent-handoff.md` → relevant skill SKILL.md.

## 1. Local Dev Startup (script-truth policy)

**Always use PowerShell script — never bash, never `source .env`:**

```powershell
.\scripts\start-api.ps1          # start API on :8081
.\scripts\start-api.ps1 -Build   # rebuild binary first
```

- Canonical scripts only supported entrypoint.
- Scripts must rebuild or prove freshness.
- Scripts must fail loud on stale binary, blocked port, missing dependency, broken prerequisite.
- Ad hoc startup commands not authoritative.

Login (dev): `POST /api/v1/auth/login` body `{"identifier":"admin","password":"AdminMetalDocs123!"}`.

Full details: [`wiki/references/local-dev-startup.md`](wiki/references/local-dev-startup.md).

## 2. Wiki Is Source of Truth

Project knowledge lives in `wiki/`. **Always read `wiki/README.md` first** — indexes every doc with `file:line` anchors. No re-grep codebase when wiki already says where things are.

Critical entry points:
- [`wiki/concepts/placeholders.md`](wiki/concepts/placeholders.md) — eigenpal native vs MetalDocs legacy gap (fixed 7-token catalog)
- [`wiki/concepts/authz-tiers.md`](wiki/concepts/authz-tiers.md) — two-tier authz + Postgres tripwire
- [`wiki/modules/editor-ui-eigenpal.md`](wiki/modules/editor-ui-eigenpal.md) — eigenpal Anti-Corruption Layer
- [`wiki/decisions/`](wiki/decisions/) — ADRs (token migration, atomic CD create, contract-first API, etc.)

**Drift policy:** change code referenced by wiki doc → update its `Last verified:` stamp same change.

**After refactors / new implementations** → dispatch `wiki-curator` agent ([`.claude/agents/wiki-curator.md`](.claude/agents/wiki-curator.md)). Refreshes Key files anchors, bumps stamps, updates `wiki/README.md` index, creates new docs when new module/concept/workflow appears. Invoke proactive — no drift accumulation.

Full module documentation, maturity promotion, or rebuild module wiki trio/artifacts → [`metaldocs-module-doc`](.agents/skills/metaldocs-module-doc/SKILL.md). Implementation work touches already-documented module → [`metaldocs-module-doc-sync`](.agents/skills/metaldocs-module-doc-sync/SKILL.md).

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
| Organize a large program into Milestones → Features with consumer-contract-first per-feature spec gates, an independent `milestone-validator` close gate (separation of powers, binding C1–C7), human-in-loop hard-stops, spec-up-front per milestone, evidence-based close-out | [`milestone`](.claude/skills/milestone/SKILL.md) |

### Frontend rules (non-negotiable)
Canonical structure: [`wiki/architecture/frontend-structure.md`](wiki/architecture/frontend-structure.md).
- Feature-sliced layout, `createBrowserRouter` + per-feature `routes.tsx`
- TanStack Query for server state, OpenAPI-codegen types from `lib/api-types/`
- CSS Modules + design tokens
- **No** `HashRouter`, string-pattern path dispatchers, legacy `src/api/`, root flat files
- **Never reintroduce legacy paths.** Touch file outside canonical layout → migrate same change. No shims. No re-exports.

### Backend/API rules
Canonical structure: [`wiki/architecture/backend-api-structure.md`](wiki/architecture/backend-api-structure.md). Contracts: [`api-contract.md`](wiki/architecture/api-contract.md), [`api-design-system.md`](wiki/architecture/api-design-system.md).
- No change public routes, generated `api.gen.go` wiring, or OpenAPI shape from memory.
- Build route truth table first → compare runtime / spec / codegen / wiki → implement from canonical module pattern.

### Database rules
Source of truth: [`wiki/database/`](wiki/database/) (schema ownership, dictionary, migration policy, reference data, bootstrap rules). No duplicate those rules here.

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
4. No hidden drift left in repaired boundary
5. Skill/runbook/instruction updated if failure exposed workflow gap

**Hard-stop rule:** no continue through redesign-grade issues. If required fix implies shared API redesign, cross-module auth/authz model change, storage/provider architecture redesign, workflow semantic redesign outside assigned boundary, or large coordinated rewrite → stop, report architecture boundary and minimum prerequisite plan instead of symptom patching.

**Stop rule:** no continue feature work through failing prerequisite boundary. If startup, auth/session, target route, or shared contract truth fails → switch to `runtime-contract-prereq`, repair boundary, rerun checkpoint, return to original task.

**Default close-out loop for non-trivial work:**
1. Implement inside bounded task
2. Run static and targeted verification for touched slice
3. Run code review
4. Run product QA using canonical checklist for workflow class
5. Classify findings by root-cause family
6. Fix by family, not scattered symptom patching
7. Rerun targeted review, QA, regression
8. Rerun broader regression when change crossed boundaries
9. Close only with evidence and explicit bounded defers

**Evidence rule:** `implemented`, `fixed`, `done`, `green`, or `looks good` not sufficient closure by default. Record verification commands, QA outcomes, review findings disposition, and any remaining bounded defer before claiming completion.

**Canonical QA checklists:**
- [`wiki/quality/screen-qa-checklist.md`](wiki/quality/screen-qa-checklist.md)
- [`wiki/quality/backend-api-qa-checklist.md`](wiki/quality/backend-api-qa-checklist.md)
- [`wiki/quality/workflow-async-qa-checklist.md`](wiki/quality/workflow-async-qa-checklist.md)
- [`wiki/quality/release-closeout-checklist.md`](wiki/quality/release-closeout-checklist.md)

### Test framework hard gate

**All tests must use the canonical test framework for their class. No exceptions for new tests.**

- **DB integration tests** → unified `testdb` factory framework (M4c). No ad-hoc `sqlmock` / per-package fixture builders for new tests.
- **HTTP handler / delivery tests** → canonical handler-test framework with typed fixtures (UUIDs for UUID-typed contract fields, deterministic identity, shared fake builders). No new ad-hoc `fakeSvc` literals with sloppy fixture strings (`"sess_1"`, `"doc_1"`, `"tenant-a"`, `"rev_2"`).
- **Application / domain unit tests** → table-driven Go tests with shared in-memory fakes from the module's `testing` subpackage. No new bespoke per-test stubs duplicating service contracts.

**No new test outside the framework.** A new test file or new test function that bypasses the canonical framework for its class is a review block. Reviewers reject; authors migrate before merge.

**Drive-by repair (not big-bang).** Pre-existing non-framework tests are not a mass-migration sweep. Policy: when a feature touches a test file with non-framework patterns, migrate the touched tests to the framework as part of that feature. Adjacent untouched tests stay until their own feature touches them. Same surgical rule as §5.3.

**Trigger smell:** a typed contract change (`uuid.Parse`, typed struct) breaks tests because fixtures used sloppy strings → root cause = non-framework fixture, not the contract fix. Migrate the fixture, do not weaken the contract.

**Framework references:**
- M4c testdb framework: see memory `m4c-test-fixture-framework.md` + related ADRs in `wiki/decisions/`
- Handler-test framework (formalization pending — track via wiki + ADR when scaffolded)

## 5. Behavioral Guidelines

> Bias toward caution over speed. Trivial tasks → use judgment.

### 5.0 Commit authorization (standing)
**You may commit without asking.** Overrides the harness default ("commit or push only when the user asks") — operator grants standing authorization to commit work as part of normal task close-out. Still: keep commit messages honest, never `--no-verify`/skip hooks, and **never push** without explicit ask (commit ≠ push).

### 5.1 Think before coding
No assume. No hide confusion. Surface tradeoffs.
- State assumptions explicit. Uncertain → ask.
- Multiple interpretations? Present them — no silent pick.
- Simpler approach exists? Say so. Push back when warranted.
- Unclear? Stop. Name what confusing. Ask.

### 5.2 Simplicity first
Minimum code that solves problem. Nothing speculative.
- No features beyond ask.
- No abstractions for single-use code.
- No flexibility / configurability not requested.
- No error handling for impossible scenarios.
- Wrote 200 lines could be 50 → rewrite.
- Senior engineer test: would they call this overcomplicated? Yes → simplify.

### 5.3 Surgical changes
Touch only what you must. Clean up only own mess.
- No improve adjacent code, comments, or formatting.
- No refactor things not broken.
- Match existing style even if you'd do different.
- Notice unrelated dead code? Mention it — no delete.
- Orphans from your changes: remove. Pre-existing dead code: leave unless asked.
- Test: every changed line traces direct to user request.

### 5.4 Goal-driven execution
Define success criteria. Loop until verified.
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

Multi-step tasks → state brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
```

Strong success criteria let you loop independent. Weak criteria ("make it work") require constant clarification.

---

**Working signal:** fewer unnecessary changes in diffs, fewer rewrites from overcomplication, clarifying questions arrive before implementation not after mistakes.

---

## GitNexus usage policy (OVERRIDES the managed block below)

The auto-generated `GitNexus — Code Intelligence` block below is **advisory, not mandatory**. Its "MUST run impact before every edit" / "NEVER commit without detect_changes" framing is **downgraded to opt-in** to save tokens.

- GitNexus is **opt-in**, for **high-risk or large-blast changes only**: cross-module refactors, renames touching many callers, edits to shared/public symbols, unfamiliar subsystems.
- **Routine edits** (single function, local change, well-understood code) → just use `Grep`/`Glob`/`Read`. **Do NOT** run `impact`/`detect_changes` per edit.
- Run `impact({direction: "upstream"})` **only when blast radius is genuinely unknown** and the change is risky. Then warn on HIGH/CRITICAL.
- `rename` (call-graph aware) over find-and-replace **still applies** for multi-file renames.
- This override wins on regeneration: re-running `analyze` rewrites the block below but not this section.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **MetalDocs** (31446 symbols, 70239 relationships, 300 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

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
