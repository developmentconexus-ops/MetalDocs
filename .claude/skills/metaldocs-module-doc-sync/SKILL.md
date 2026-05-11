---
name: metaldocs-module-doc-sync
description: Incremental diff-driven updater for module docs produced by metaldocs-module-doc. Use this skill whenever code changes touch a module that already has a wiki/modules/<m>.md + <m>-tech-debt.md + backlog/<m>-refactor.md trio — at end of a refactor task, after merging a PR, after running a migration, or whenever the user says "sync the wiki for module X", "update X docs after this refactor", "the docs are stale, refresh them", "wiki-curator pass on X", or similar. Reads the in-progress execution (changed files, commits, plan task just done) and surgically patches affected sections of the live doc + tech-debt register + refactor backlog. Bumps Last verified stamps and re-validates with tally_check.sh. Iron rule: this skill patches existing sections only. Adding a new §, changing C4 boxes, or onboarding a never-documented module requires the full metaldocs-module-doc skill instead.
---

# MetalDocs Module Doc Sync

Keep canonical module docs live during refactor execution. The sister skill `metaldocs-module-doc` builds the doc from scratch (8 phases, heavy). This skill keeps it true between rebuilds (3 phases, light).

## Why this skill exists separately

Module docs that drift become more dangerous than no docs — readers trust them, get burned, then trust no doc. The sweep skill is too heavy to re-run after every refactor PR. Without a lighter sync workflow, the realistic outcome is "we'll update the wiki later" — and later never comes. This skill is the in-between: cheap enough to run after each task, strict enough to catch drift before it accumulates.

The split is also about scope discipline. The sweep skill can change anything — new diagram boxes, new sections, re-rated severities. This skill can only **patch what already exists**. That keeps both safe: the sweep skill stays infrequent and deliberate; the sync skill stays mechanical and safe to run often.

## When to use which

| Situation | Skill |
|---|---|
| Module has no doc yet | `metaldocs-module-doc` (full sweep) |
| Doc exists; refactor changed file:line anchors, surface symbols, or one debt row's status | **this skill** |
| Doc exists; refactor introduced a new HTTP route, a new persistence table, a new IN/OUT cross-dep, or a new state machine | `metaldocs-module-doc` (the §3/§5 diagram shape changed — sweep) |
| Doc exists; debt fully retired, backlog row merged | **this skill** |
| Last verified stamp >30 days old AND no recent code change record | `metaldocs-module-doc` (re-validate from scratch — anchors likely rotted silently) |

When in doubt: try this skill first. If Phase 1 diff scan reports "structural change detected" (new route / table / dep / state machine), it will tell you to escalate to the sweep skill instead of forcing a patch.

## Output contract

For module `M`, this skill produces:
- Patches to `wiki/modules/M.md` (§5 Key Files anchors, §6 flow tables, §8 cross-deps, §11 stats — surgical edits only)
- Patches to `wiki/modules/M-tech-debt.md` (close resolved T-NNN rows, update file:line anchors, never escalate/de-escalate severity here — that's a judgment call requiring sweep)
- Patches to `wiki/backlog/M-refactor.md` (mark R-NNN status `merged` w/ PR link)
- Bumped `Last verified:` stamp on all three (use today's date — `2026-MM-DD`)
- Append one line to `wiki/modules/M/_artifacts/sync-log.md` per sync run
- Mechanical gate: `tally_check.sh M` must PASS after edits

If the doc set doesn't exist yet, refuse and point at `metaldocs-module-doc`.

## 3-phase workflow

| Phase | Owner | Inputs | Artifact |
|---|---|---|---|
| 1 — Diff scan | Codex subagent | module name + change context (commits, plan task, or explicit file list) | scan result inline (no separate file — embed in subagent reply) |
| 2 — Surgical patch | Main agent | scan result + existing doc trio | patched docs + `sync-log.md` entry |
| 3 — Gate | Main agent | `scripts/tally_check.sh` from sibling skill | `[tally] PASS` printed |

## Run sequence

1. **Establish change context.** Ask the user (or read from the calling context) which of these applies:
   - Plan task just completed — get the task description + list of files touched in this task.
   - Commits — get a git range (`HEAD~3..HEAD`, branch name, etc.).
   - Explicit file list — the user names files.
   Without a change context, refuse: this skill is diff-driven; running it blind = re-do a sweep badly.

2. **Verify doc trio exists** for the target module: `wiki/modules/M.md`, `wiki/modules/M-tech-debt.md`, `wiki/backlog/M-refactor.md`. If any is missing, escalate to `metaldocs-module-doc`.

3. **Phase 1 — Diff scan.** Dispatch one `codex:codex-rescue` subagent w/ `templates/subagent-diff-scan.md`. Pass the change context and the doc trio paths. Subagent returns a structured report: list of changed file:line anchors that the docs cite, list of debt rows that reference changed code, structural-change flag (new route / new table / new dep / new state machine = YES → escalate).

4. **Decision branch.**
   - Structural-change flag is YES → STOP. Tell the user "structural change detected: <thing>. Patching would create drift. Re-run `metaldocs-module-doc` for this module instead."
   - Structural-change flag is NO → continue to Phase 2.

5. **Phase 2 — Surgical patch.** For each change in the scan:
   - File:line anchor moved → update §5 Key Files line in `M.md` and any tech-debt row citing that file:line.
   - Symbol renamed → update every occurrence in the doc trio.
   - Debt resolved by the change (deleted code, fixed bug, refactor completed) → close the `T-NNN` row in `M-tech-debt.md` (move to a "Resolved" subsection w/ resolution date + commit ref, OR delete if you also mark the backlog row merged in the same patch); mark the corresponding `R-NNN` row `status: merged` w/ PR link in `M-refactor.md`.
   - Behavior change inside a documented flow → update only the affected step in the §6 flow table.
   - Last verified date — bump to today on all three files.

6. **Recompute counts.** If T-NNN rows were closed or new ones explicitly added by the user in the same patch, recompute §11 stats (Critical / Major / Minor counts; "Decisions without ADR link" total).

7. **Append sync-log entry.** One line in `wiki/modules/M/_artifacts/sync-log.md` per the template at `templates/sync-log-entry.md`. Captures date, change context, files patched, T-NNN/R-NNN affected.

8. **Phase 3 — Gate.** Run `bash .claude/skills/metaldocs-module-doc/scripts/tally_check.sh M`. On FAIL: fix patches, re-run until PASS. Do NOT publish on FAIL.

9. **No commit from this skill.** The calling context (the refactor task) owns the commit. Sync edits land in the same commit as the code change OR in an immediately-following doc commit. The user decides which.

## Patches the skill is allowed to make

- File:line anchor updates
- Symbol-name renames (mechanical replace across doc trio)
- Last verified stamp bumps
- T-NNN status closures (resolved by the change in this patch)
- R-NNN status updates (open → in-progress, in-progress → merged)
- §6 flow-table cell edits (one cell at a time, when behavior in a step changed but the step itself still exists)
- §11 count recomputation when severity counts shift due to closures
- Cross-link target updates (when a wiki/* file the doc references was renamed)

## Patches the skill is NOT allowed to make

These all indicate a structural change that needs the sweep skill — stop and tell the user.

- Adding a new T-NNN row (new debt = new severity call = main-agent judgment + ADR check)
- Re-rating a T-NNN severity (Major ↔ Critical = rubric application = sweep territory)
- Adding a new step to a §6 flow table
- Adding/removing a box from a §3 or §5 Mermaid diagram
- Adding a new HTTP route or persistence table reference
- Onboarding a new cross-dep (new IN-edge or OUT-edge in §5 / §8)
- Rewriting prose paragraphs (sync is about facts, not narrative)
- Industry-comparison edits (admissions tracking in §5 = sweep)

## Subagent purity

Same rule as the sibling skill. Phase 1 diff scan subagent returns facts only — file:line anchors that moved, symbol renames, debt rows that reference changed code. No "should", "recommend", "professional", "industry-standard". If the subagent prose contains those, reject and re-dispatch with stricter framing.

## Anti-patterns

- Running sync without a change context (= silent sweep, drift in disguise).
- Patching a doc trio that's missing one of the three files (escalate to sweep).
- Adding a new section under §5 because "the refactor introduced a new component" — that's a sweep trigger.
- Closing a T-NNN row without recomputing §11 counts.
- Skipping tally_check.sh because "the edit was small."
- Letting the subagent rewrite English prose.
- Sync-log entry that says "updated docs" with no detail — log must list every T-NNN/R-NNN/file:line touched so the next reader can audit.
- Bumping Last verified without actually verifying anchors (`grep -n <symbol> <file>` proves the anchor still resolves).
- Committing sync edits + code changes mixed when the user prefers separate doc commits — ask once if unclear.

## Red flags — STOP

| Thought | Reality |
|---|---|
| "Diff scan says new route added — I'll just add the row to §5" | New route = §5 Container box change = sweep. STOP. |
| "Severity should be Critical now after this refactor" | Sync skill can't re-rate. Sweep. STOP. |
| "Doc trio doesn't fully exist — I'll patch what's there" | Escalate to sweep. STOP. |
| "tally_check.sh failed, but my patch is right — bypass the gate" | Gate is the only thing between sync and silent drift. Fix patch. |
| "I'll skip the sync-log entry, the commit message is enough" | Sync log lives next to artifacts so future Phase 1 scans can read it. Keep. |

## Output expectations

End-of-run report: module name · change context (one line) · scan summary (N anchors moved, N symbols renamed, N debt rows touched) · patched files (list) · tally gate result · sync-log entry written (yes/no) · escalation needed (yes/no, with reason).

## Changelog

- 1.0 (2026-05-11) — initial release. Pairs with metaldocs-module-doc v1.1. Reuses scripts/tally_check.sh from sibling. 3-phase workflow (diff scan → surgical patch → gate). Strict scope: patches existing sections only; structural change escalates to sweep.
