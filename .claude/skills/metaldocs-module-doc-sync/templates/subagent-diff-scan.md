# Subagent - Targeted Diff Evidence Scan

You are a research subagent. Return facts only. Do not prescribe fixes. Do not edit files.

## Inputs

- Module name: `<m>`
- Doc trio:
  - `wiki/modules/<m>.md`
  - `wiki/modules/<m>-tech-debt.md`
  - `wiki/backlog/<m>-refactor.md`
- Artifact folder: `wiki/modules/<m>/_artifacts/`
- Change context: git range, task description plus touched files, or explicit file list

## Task

Compare the change context with the module docs and report every wiki fact that now needs a targeted update.

### 1. Anchors

- Every documented `path:line` whose target moved, was renamed, or was deleted.
- Format: `- <doc-file> L<doc-line>: <old-anchor> -> <new-anchor|DELETED>`

### 2. Public Surface

- Exported symbols, public interfaces/structs, route handlers, ports, or generated API methods added/removed/renamed by the change.
- Include the wiki section/artifact where each fact is currently documented or missing.

### 3. API And Routes

- Runtime routes added/removed/changed.
- OpenAPI paths/operationIds/schemas added/removed/changed.
- Generated codegen methods added/removed/changed.
- Permission/capability mapping changes.
- Format: `- <METHOD> <path>: runtime=<file:line> spec=<path|missing> operationId=<id|missing> codegen=<method|missing> status=<aligned|spec missing|runtime missing|changed>`

### 4. Runtime Behavior

- Documented flow steps, state transitions, authz layers, error envelopes, idempotency, audit emission, or transaction boundaries changed by the diff.
- Name the affected `wiki/modules/<m>.md` section and `_artifacts/02-flow-*.md` file when present.

### 5. Persistence

- Migrations, tables, constraints, triggers, GUC/tripwire behavior, tenant columns, or owned/read tables changed by the diff.
- Name affected doc sections and `_artifacts/04-persistence.md` when present.

### 6. Dependencies

- New or removed IN edges and OUT edges at package/module level.
- Include composition-root wiring changes.
- Name affected C4 relations, cross-dep sections, cross-links, and `_artifacts/03-deps.md`.

### 7. Debt And Backlog

- Existing T-NNN rows resolved, partially resolved, invalidated, or needing anchor/evidence updates.
- Existing R-NNN rows that moved to in-progress/merged or need evidence updates.
- Do not invent new debt IDs. Report possible new debt only as `unregistered evidence`.

### 8. Recommended Sync Mode

Return one of:
- `lite patch`
- `structural refresh`
- `full sweep`

Use `structural refresh` for bounded route/API/persistence/dependency/flow changes that can be verified from the diff.
Use `full sweep` only when the module boundary changed, the doc trio/artifacts are missing, or the diff is too broad to verify surgically.

## Output Format

Single markdown reply with sections 1 through 8. Bullets only. If a section is empty, write `- (none)`.

## Forbidden

- Do not suggest implementation fixes.
- Do not rewrite prose.
- Do not re-rate severity.
- Do not invent T-NNN or R-NNN rows.
- Do not assume a route/spec/codegen item exists without citing the file evidence.
