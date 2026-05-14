## Backend/API

For ANY work on MetalDocs backend HTTP routes, OpenAPI, oapi-codegen, handler wiring, API contracts, or route migrations, use the `metaldocs-backend-api` skill at `.agents/skills/metaldocs-backend-api/SKILL.md` and read:

- `wiki/architecture/backend-api-structure.md`
- `wiki/architecture/api-contract.md`
- `wiki/architecture/api-design-system.md`

Do not duplicate the backend/API rules here. The wiki is the source of truth and the skill is the required workflow.

## Database

For ANY work on MetalDocs database migrations, bootstrap, curated baseline, reference data, dev seeds, schema ownership, Postgres extensions, grants, triggers, functions, `schema_migrations`, or database dictionary/wiki pages, use the `metaldocs-database` skill at `.agents/skills/metaldocs-database/SKILL.md`.

The database wiki under `wiki/database/` is the source of truth for schema ownership, dictionary entries, migration policy, reference data, and bootstrap rules. Do not duplicate those rules here.

## Frontend

For ANY work under `frontend/apps/web/`, use the `metaldocs-frontend` skill at `.agents/skills/metaldocs-frontend/SKILL.md`.

For designed screens under `frontend/apps/web/design-source/`, also use `metaldocs-screen-implementation` at `.agents/skills/metaldocs-screen-implementation/SKILL.md`.

Before real screen finalization, use `metaldocs-screen-integration-audit` at `.agents/skills/metaldocs-screen-integration-audit/SKILL.md` when the screen/backlog may include mock-era widgets, missing backend capability, legacy API wrappers, deferred items, or uncertainty about what can be wired to real product behavior.

## Frontend API / TanStack Query

For ANY work on MetalDocs frontend API calls, TanStack Query hooks, query keys, cache invalidation, optimistic updates, generated frontend API types, or server-state performance under `frontend/apps/web/src/`, use the `metaldocs-tanstack-query` skill at `.agents/skills/metaldocs-tanstack-query/SKILL.md`.

Use it alongside the broader frontend architecture workflow. The wiki remains the source of truth; do not duplicate the TanStack rules here.

## Module Wiki Memory

For full module documentation, maturity promotion, deep module mapping, or rebuilding a module wiki trio/artifacts, use `metaldocs-module-doc` at `.agents/skills/metaldocs-module-doc/SKILL.md`.

After implementation work touches an already documented module, use `metaldocs-module-doc-sync` at `.agents/skills/metaldocs-module-doc-sync/SKILL.md` to update module docs, tech-debt, backlog, route truth, artifacts, and sync logs from the concrete change context.

## Core Skill Map

Use the smallest skill set that matches the task boundary:

- backend route / OpenAPI / codegen / handler work -> `metaldocs-backend-api`
- frontend implementation under `frontend/apps/web/` -> `metaldocs-frontend`
- frontend API wrappers / query hooks / generated frontend API types -> `metaldocs-tanstack-query`
- designed screen from `frontend/apps/web/design-source/` -> `metaldocs-frontend` + `metaldocs-screen-integration-audit` when real capability mapping is needed + `metaldocs-screen-implementation`
- full module wiki build or maturity promotion -> `metaldocs-module-doc`
- post-implementation wiki update for documented modules -> `metaldocs-module-doc-sync`
- startup / auth / route / runtime-spec-generated-wrapper drift -> `runtime-contract-prereq`
- database migrations / bootstrap / curated baseline / seeds / dictionary -> `metaldocs-database`

If a task crosses boundaries, compose the skills. Do not force one skill to absorb unrelated prerequisite work.

## MetalDocs AI Operating System

Use the MetalDocs operating model for all non-trivial work.
Canonical design: `docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md`.

Truth hierarchy:
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

Default workflow order:
1. Choose the task skill.
2. Read the required wiki/architecture docs for that boundary.
3. Pass the relevant gate before implementation.
4. Implement only inside the current boundary.
5. Verify with scripts/tests.
6. Sync module wiki memory if code truth changed.

Escalate to `runtime-contract-prereq` when startup truth, auth/session truth, target route truth, or runtime/spec/generated/frontend-wrapper alignment is not trustworthy enough to continue local feature work.

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

<!-- context7 -->
Use Context7 MCP to fetch current documentation whenever the user asks about a library, framework, SDK, API, CLI tool, or cloud service -- even well-known ones like React, Next.js, Prisma, Express, Tailwind, Django, or Spring Boot. This includes API syntax, configuration, version migration, library-specific debugging, setup instructions, and CLI tool usage. Use even when you think you know the answer -- your training data may not reflect recent changes. Prefer this over web search for library docs.

Do not use for: refactoring, writing scripts from scratch, debugging business logic, code review, or general programming concepts.

## Steps

1. Always start with `resolve-library-id` using the library name and the user's question, unless the user provides an exact library ID in `/org/project` format
2. Pick the best match (ID format: `/org/project`) by: exact name match, description relevance, code snippet count, source reputation (High/Medium preferred), and benchmark score (higher is better). If results don't look right, try alternate names or queries (e.g., "next.js" not "nextjs", or rephrase the question). Use version-specific IDs when the user mentions a version
3. `query-docs` with the selected library ID and the user's full question (not single words)
4. Answer using the fetched docs
<!-- context7 -->
