---
name: metaldocs-module-doc-sync
description: "Use after an implementation, refactor, migration, route/API change, bug fix, or plan task touches an already documented MetalDocs module and the wiki needs a bounded update. Use this only when the sync can name the exact change context and every affected module. This workflow does not permit silent omissions. This Codex-discoverable bridge points to `.claude/skills/metaldocs-module-doc-sync/SKILL.md`."
---

# MetalDocs Module Doc Sync

Read and follow `.claude/skills/metaldocs-module-doc-sync/SKILL.md`.

This bridge exists so Codex sessions that discover `.agents/skills` still load the canonical module doc sync workflow.

Use this only when the sync can name the exact change context and every affected module. This workflow does not permit silent omissions.
Require an explicit affected-surface scan before claiming the sync is complete.

Use `.agents/skills/metaldocs-module-doc/SKILL.md` instead when the module lacks the full doc trio/artifacts or needs a full maturity rebuild.

Required canonical resources live under `.claude/skills/metaldocs-module-doc-sync/`:

- `templates/subagent-diff-scan.md`
- `templates/sync-log-entry.md`
- `scripts/wiki_sync_preflight.ps1`

Stop if the canonical `.claude/skills/metaldocs-module-doc-sync/SKILL.md` file is missing.
