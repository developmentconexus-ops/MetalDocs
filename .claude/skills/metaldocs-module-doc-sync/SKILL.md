---
name: metaldocs-module-doc-sync
description: Targeted, change-context-driven updater for MetalDocs technical wiki module memory. Use after an implementation, refactor, migration, route/API change, bug fix, or plan task touches an already documented module and the user asks to sync, refresh, update, verify, or keep the wiki current. Updates the affected module doc, tech-debt register, refactor backlog, route truth table, runtime flows, public surface, persistence facts, cross-dependencies, source artifacts, changelog, and sync log from the changed code only. Do not use for onboarding a never-documented module, blind full-wiki refreshes, or broad architecture rewrites without a concrete change context; use metaldocs-module-doc for those.
---

# MetalDocs Module Doc Sync

Keep the MetalDocs wiki useful as technical memory for future LLM and developer work.

This skill is not a cosmetic documentation pass. It refreshes the facts an agent needs to understand a module without rereading the whole codebase: API routes, runtime behavior, authz, persistence, public surface, cross-module dependencies, state transitions, known debt, and implementation evidence.

## Core Rule

Sync from a concrete change context only.

Accepted contexts:
- A git range, commit, or branch diff.
- A completed plan/task with touched files.
- An explicit file list from the user.
- The current uncommitted diff when the user just asked to sync after the work in this thread.

If there is no change context, do not improvise a sweep. Ask for the context or use `metaldocs-module-doc` for a deliberate full module rebuild.

## Scope Choice

Use the smallest mode that keeps the module true.

| Mode | Use when | Update |
|---|---|---|
| Lite patch | Anchors moved, symbols renamed, debt closed, verified date changed | Existing lines, counts, backlog status, sync log |
| Structural refresh | The changed files add/remove/change routes, OpenAPI contract, public structs/interfaces, DB tables/migrations, dependencies, state transitions, authz, or runtime flow behavior | The affected sections and matching `_artifacts/*` files for that module |
| Full rebuild | The module has no complete doc trio/artifacts, the change spans too many unrelated areas, the docs are stale beyond recovery, or evidence cannot be bounded to the diff | Use `metaldocs-module-doc` |

Structural refresh is allowed. Do not escalate merely because a new route/table/dep exists; update the affected module fully from the diff. Escalate only when the changed scope cannot be verified without redoing the full module sweep.

## Sync Success Gate

A sync may claim success only when it includes:
1. Exact change context
2. Explicit affected-module list
3. Explicit affected-surface scan
4. Mode classification: lite patch, structural refresh, or full rebuild escalation
5. Preflight/tally result
6. Explicit explanation for every touched module that was not updated

No silent omissions are allowed.

## Required Wiki Shape

A synced backend module should preserve these memory surfaces when applicable:
- `wiki/modules/M.md`: Arc42/C4 module doc, key files, C4 diagrams, public surface, HTTP operations, API Route Truth Table, runtime flows, state transitions, authz, persistence summary, cross-deps, decisions, quality requirements, risk summary, glossary, cross-links, changelog.
- `wiki/modules/M-tech-debt.md`: evidence-backed debt rows, severity counts, ADR coverage stats, closed/resolved markers when applicable.
- `wiki/backlog/M-refactor.md`: executable refactor rows tied to debt IDs or maintenance IDs.
- `wiki/modules/M/_artifacts/*.md`: source scans for context, surface, flows, deps, persistence, industry/self-review where they already exist.
- `wiki/modules/M/_artifacts/sync-log.md`: append-only audit of every sync run.

Frontend-only or legacy wiki pages may not have the full trio. If the trio is missing, report that the module is not ready for this skill and use/prepare `metaldocs-module-doc` instead.

## Workflow

1. Determine changed modules.
   - Read the diff or touched file list.
   - Map `internal/modules/<name>/...`, `frontend/.../features/<name>/...`, `api/openapi/...`, `migrations/...`, and composition-root changes to affected wiki modules.
   - If a cross-cutting file changes, update every documented module whose route, contract, persistence, runtime flow, or dependency behavior changed.
   - Do not treat cross-cutting files as single-module edits by default.

2. Run preflight.
   - Prefer `scripts/wiki_sync_preflight.ps1 -Module M` from this skill.
   - Confirm the doc trio exists.
   - Confirm `_artifacts/` exists for full module docs.
   - Locate Git Bash for the tally gate on Windows.
   - Note existing tally failures before editing; do not hide pre-existing drift.

3. Scan evidence.
   - Use `templates/subagent-diff-scan.md` for a focused evidence scan when useful.
   - Verify facts directly from code, OpenAPI, migrations, tests, and existing artifacts.
   - Use `rg`, `git diff`, `git grep`, and targeted file reads. Avoid rereading the whole repo.

4. Patch the wiki.
   - Update every affected fact in the module doc, not only line anchors.
   - If routes changed, update both `HTTP operations` and `API Route Truth Table`.
   - If OpenAPI/codegen changed, update spec path, operationId, generated method, contract status, and related debt/backlog rows.
   - If persistence changed, update persistence summaries, C4 DB labels, tripwire/authz notes, and `_artifacts/04-persistence.md` if present.
   - If dependencies changed, update C4 relationships, inbound/outbound dependency lists, cross-links, and `_artifacts/03-deps.md` if present.
   - If behavior changed inside a documented flow, update the flow section and matching `_artifacts/02-flow-*.md`.
   - If public exported surface changed, update the public surface section and `_artifacts/01-surface.md`.
   - If debt was resolved, mark/close the T-row, update the R-row, recompute counts, and record evidence.
   - Bump `Last verified:` where present. Use the actual current date in `YYYY-MM-DD`.
   - Add a changelog entry to the module doc for meaningful structural or behavioral updates.

5. Preserve discipline.
   - Do not invent debt. New debt requires concrete evidence and the existing severity rubric.
   - Do not re-rate severity unless the changed evidence clearly removes or changes the original trigger; otherwise defer to a full sweep.
   - Do not rewrite narrative prose unless the fact changed.
   - Do not update unrelated modules just because they are nearby.

6. Write the sync log.
   - Use `templates/sync-log-entry.md`.
   - Include changed context, mode, structural facts updated, files patched, T/R rows touched, and gate result.

7. Gate.
   - Run the sibling tally check after edits.
   - On Windows, prefer:
     `& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh M`
   - If plain `bash` works, this is also acceptable:
     `bash .claude/skills/metaldocs-module-doc/scripts/tally_check.sh M`
   - If the gate fails because of your edits, fix and rerun.
   - If the gate failed before your edits, report the pre-existing failure and keep your changes internally consistent.

## Allowed Updates

- File and line anchors.
- Symbol names and public surface summaries.
- API route truth tables and HTTP operation tables.
- OpenAPI/codegen status for affected routes.
- Runtime flow steps and sequence diagrams for affected operations.
- C4 boxes/relationships when the diff changed module structure.
- Persistence tables, migrations, constraints, triggers, tripwire/authz notes.
- Cross-dependency lists and cross-links.
- Tech-debt row status, evidence, counts, and ADR coverage stats.
- Backlog row status and links.
- Last verified stamps, module changelog, and sync log.

## Escalate To Full Sweep

Use `metaldocs-module-doc` instead when:
- The module lacks `M.md`, `M-tech-debt.md`, or `M-refactor.md`.
- `_artifacts/` is missing for a module that is supposed to be an Arc42/C4 living doc.
- The implementation changed the module's purpose or ownership boundary.
- More than a small set of modules changed and the affected facts cannot be isolated.
- Existing docs are contradictory enough that patching would be guesswork.
- The user asks for a full wiki/module rebuild.

## Output

End with:
- Module(s) synced.
- Change context.
- Explicit affected-surface scan.
- Mode used: lite patch, structural refresh, or escalated full sweep.
- Facts updated: routes, OpenAPI/codegen, flows, persistence, deps, public surface, debt/backlog.
- Files patched.
- Tally result, including whether failures were pre-existing.
- Every touched module not updated, with reason.
- Any modules skipped because their doc trio is incomplete.

## Bundled Resources

- `templates/subagent-diff-scan.md`: focused evidence scan prompt.
- `templates/sync-log-entry.md`: sync log block format.
- `scripts/wiki_sync_preflight.ps1`: doc trio/artifact/Git Bash/tally preflight helper.
