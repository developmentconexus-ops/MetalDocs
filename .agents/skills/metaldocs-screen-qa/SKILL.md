---
name: metaldocs-screen-qa
description: "Use to run an autonomous, closed-loop QA pass on ONE MetalDocs frontend screen, driven through the built-in Preview browser as a real user. This Codex-discoverable bridge points to the canonical workflow in `.claude/skills/metaldocs-screen-qa/SKILL.md` and must be used for any `qa/<slug>` per-screen QA run from the frontend-screen-qa-roadmap."
---

# MetalDocs Screen QA

Read and follow `.claude/skills/metaldocs-screen-qa/SKILL.md`.

This bridge exists so Codex sessions that discover `.agents/skills` still load the canonical screen-QA workflow.

Always load these together:

1. `wiki/quality/qa-operating-system.md` (5 truths / 7 gates / closed loop / classification + severity / hard-stop)
2. `wiki/quality/screen-qa-checklist.md`
3. `.claude/PRPs/plans/frontend-screen-qa-roadmap.plan.md` (screen inventory, route/module/QA-class map, ordering, branch names)
4. The owning `wiki/modules/<module>.md` + `<module>-tech-debt.md` for the screen under QA
5. `wiki/quality/workflow-async-qa-checklist.md` when the screen is approval / distribution / render-fanout (async-owned)

Honor the Iron Law: no closure without evidence; drive the screen as a user via Preview (`preview_*`), never Chrome MCP or raw Playwright; root-cause by family; hard-stop on redesign-grade fixes and report instead of patching.

Stop if the canonical `.claude/skills/metaldocs-screen-qa/SKILL.md` file is missing.
