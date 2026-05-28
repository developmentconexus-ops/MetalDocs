# Docs Workspace

> **Last verified:** 2026-05-27
> **Purpose:** Staging workspace for non-canonical material.

`docs/` is not the durable source of truth for maintained project knowledge.

Use `wiki/` for canonical architecture, module, workflow, database, quality, and decision docs. The wiki landing page is [`../wiki/index.md`](../wiki/index.md).

## What belongs in `docs/`

- design specs and plans
- exploratory or imported research
- temporary runbooks not yet normalized into wiki truth
- prompts, templates, and execution helpers
- historical or legacy material kept for reference

## Promotion rule

When a `docs/` page becomes durable maintained truth, promote it into the owning wiki section and update the relevant indexes. Do not create a second canonical source here.

## Current high-signal areas

- `superpowers/specs/` and `superpowers/plans/` - design and implementation staging
- `runbooks/` - mixed operator material, some of which may later be promoted into `wiki/quality/` or other wiki domains
- `adr/` - legacy ADR history not yet reconciled with `wiki/decisions/`
- `db-research/`, `ck5-wiki/`, `hardening/` - reference and research material
