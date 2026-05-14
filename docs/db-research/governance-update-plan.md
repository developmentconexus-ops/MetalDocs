# Governance Update Plan

## Commands

```powershell
rg -n "governance|database|db|dictionary|baseline|policy|canonical|source of truth|workflow" AGENTS.md CLAUDE.md wiki .agents/skills .claude/skills -S
```

## Evidence Highlights

- `AGENTS.md` and `CLAUDE.md` currently lack explicit curated DB baseline/dictionary workflow pointers.
- `wiki/architecture/data-model.md` indicates migrations-folder schema truth, which conflicts with curated baseline+tail direction from the approved 2026-05-14 design.
- `wiki/database/` does not yet exist though required by plan/spec for dictionary and migration policy centralization.

## Required Governance Updates

1. Introduce canonical DB workflow skill pointer (`metaldocs-database`) in `AGENTS.md` and `CLAUDE.md`.
2. Establish `wiki/database/` as DB operational source of truth:
   - overview
   - migration policy
   - dictionary index
   - table pages.
3. Align architecture wording (`wiki/architecture/data-model.md`) with curated baseline+tail policy.
4. Add verification gates to DB runbook surface:
   - curated bootstrap check
   - dictionary coverage check
   - legacy replay explicit path.

## Contradictions/Gaps

- Source-of-truth contradiction: historical migrations-only phrasing vs curated baseline model.
- Missing discoverable DB-specific skill bridge/canonical workflow.
- Missing wiki DB policy/dictionary tree.

## Open Questions

- Final ratification home for baseline marker policy (`wiki/database/migration-policy.md` and/or ADR).
