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

## Frontend

For ANY work under `frontend/apps/web/src/` (new screens, components, refactors, design implementation, routing, state, API wiring), use the **`metaldocs-frontend`** skill (`.claude/skills/metaldocs-frontend/SKILL.md`). It enforces the canonical structure defined in `wiki/architecture/frontend-structure.md` — feature-sliced layout, `createBrowserRouter` + per-feature `routes.tsx`, TanStack Query for server state, OpenAPI-codegen types from `lib/api-types/`, CSS Modules + design tokens, no `HashRouter`, no string-pattern path dispatchers, no legacy `src/api/` or root flat files. **Never reintroduce legacy paths.** When you touch a file outside the canonical layout, migrate it in the same change (no shims, no re-exports).

Designed screens land in `frontend/apps/web/design-source/<slug>/` (HTML + screenshot + NOTES.md). For any task that says "implement screen X" or references a `design-source/<slug>/` directory, ALSO use the **`metaldocs-screen-implementation`** skill (`.claude/skills/metaldocs-screen-implementation/SKILL.md`) on top of `metaldocs-frontend`. It drives a 6-phase workflow (Audit → Map → Pre-flight → Page assembly → Verify → Document) with hard gates that captures lessons from the Library screen rollout.

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