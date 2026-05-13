---
name: metaldocs-module-doc-sync
description: "Use after an implementation, refactor, migration, route/API change, bug fix, or plan task touches an already documented MetalDocs module and the wiki needs a bounded update. Sync module docs, tech-debt registers, refactor backlog, route truth tables, runtime flows, public surface, persistence facts, dependencies, changelog, source artifacts, and sync log from the changed code only. This Codex-discoverable bridge points to `.claude/skills/metaldocs-module-doc-sync/SKILL.md`."
---

# MetalDocs Module Doc Sync

Read and follow `.claude/skills/metaldocs-module-doc-sync/SKILL.md`.

This bridge exists so Codex sessions that discover `.agents/skills` still load the canonical module doc sync workflow.

Use this only with a concrete change context:

- git range, commit, or branch diff
- completed plan/task with touched files
- explicit file list from the user
- current uncommitted diff after work in this thread

Use `.agents/skills/metaldocs-module-doc/SKILL.md` instead when the module lacks the full doc trio/artifacts or needs a full maturity rebuild.

Required canonical resources live under `.claude/skills/metaldocs-module-doc-sync/`:

- `templates/subagent-diff-scan.md`
- `templates/sync-log-entry.md`
- `scripts/wiki_sync_preflight.ps1`

Stop if the canonical `.claude/skills/metaldocs-module-doc-sync/SKILL.md` file is missing.
