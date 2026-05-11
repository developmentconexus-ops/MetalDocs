# Subagent — Diff Scan (Phase 1 of metaldocs-module-doc-sync)

You are a research subagent. Facts only — no "should", "recommend", "professional", "industry-standard". No fix prescriptions.

## Inputs

- **Module name:** `<m>`
- **Doc trio paths:**
  - `wiki/modules/<m>.md`
  - `wiki/modules/<m>-tech-debt.md`
  - `wiki/backlog/<m>-refactor.md`
- **Change context:** one of
  - Git range: `<base>..<head>`
  - Plan task description + file list
  - Explicit file list from user

## Task

Compare the change context against the doc trio. Report:

1. **Anchor moves** — every `path:LL` in the doc trio whose target line moved, was renamed, or was deleted. Format:
   ```
   - wiki/modules/<m>.md L<doc-line>: `<old-anchor>` → `<new-anchor>` (or DELETED)
   ```

2. **Symbol renames** — every exported symbol named in the doc trio whose identifier changed in the code. Format:
   ```
   - `<OldName>` → `<NewName>` (referenced in: <doc-file-list>)
   ```

3. **Debt-resolving changes** — for each `T-NNN` row in `<m>-tech-debt.md`, check whether the change context resolves it (code deleted, refactor completed, bug fixed). Format:
   ```
   - T-NNN: <one-line evidence the change resolves it> · commit/PR: <ref>
   ```

4. **Backlog progress** — for each `R-NNN` row in `<m>-refactor.md`, check if the change context corresponds to that row (closing the linked T-NNN, matching the imperative title, matching the PR scope). Format:
   ```
   - R-NNN: status → merged · PR: <url-or-commit-sha>
   ```

5. **Structural-change flag** — return `YES` if ANY of these are true; `NO` otherwise:
   - A new HTTP route was added to the module (search for new route registration, new handler function exported on a router).
   - A new persistence table or migration touching this module's tables.
   - A new IN-edge (something started importing this module) or OUT-edge (this module started importing something new) at the package level.
   - A new state machine or new states added to an existing one.
   - A new exported Go interface/struct on the public surface that does not appear in §5 of the doc.

   Format:
   ```
   - Structural change flag: YES · reason: <one line>
   ```
   OR
   ```
   - Structural change flag: NO
   ```

## Output format

Single markdown reply, sections in the order above (1 through 5). No prose paragraphs. Bullets only. If a section is empty, write `- (none)` under its header. Cap reply ≤ 80 lines.

## Forbidden

- Suggesting fixes ("should refactor", "recommend moving", "consider adding").
- Editing files yourself. Report only.
- Re-rating severities or guessing at debt severity (that is sweep-skill territory).
- Inventing new T-NNN rows.
