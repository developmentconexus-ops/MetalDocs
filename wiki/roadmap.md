# MetalDocs Roadmap

> **Last verified:** 2026-06-14
> **Scope:** The single canonical **forward** progression surface for MetalDocs. What is
> being built next, and where the work is governed. Historical execution trackers are
> linked at the bottom — they are frozen records, not live plans.
> **One-roadmap rule:** this is the only forward roadmap. Per-program execution detail lives
> in each program's governing spec / milestone tree, not here. This page points; it does not
> re-adjudicate status.

## Active program — Grade-A Architecture Remediation

The live program. Takes the backend's three formerly-C audit dimensions to Grade A−/A across
Milestones **M0–M5**, docs-first (decision D5).

- **Program index:** [`docs/superpowers/milestones/grade-a-architecture-remediation/README.md`](../docs/superpowers/milestones/grade-a-architecture-remediation/README.md)
- **Governing spec:** [`docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`](../docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md)

| Milestone | Objective | Status |
|-----------|-----------|--------|
| M0 — Docs De-Staling | Make the wiki/progression docs truthful post v1 re-baseline | in progress |
| M1–M5 | Backend grade-A remediation (see program index for the live slice list) | planned |

> Carried in from the sealed backend tracker: the **H-G template-version-status reader**
> trigger item is **M4** of this program, not a separate backlog line.

## Post-v1 carried-forward

Real, still-open work that survives the roadmap consolidation. Carried **by reference** —
each item's status is owned by its source doc, not re-adjudicated here.

- **Screen finalization** — the 7-screen set (library, novo-documento, templates,
  caixa-aprovacao, documento-publicado, template-editor, novo-template-wizard). Recorded
  `open` as **Plan 12** in the historical refactor roadmap. Owner doc:
  [`wiki/backlog/roadmap.md`](backlog/roadmap.md) (Plan 12).
- **eigenpal post-v1 packaging / upstream consolidation** — vendor tarball now canonical at
  `third_party/eigenpal/`; packaging follow-ups deferred post-v1. Governed by
  [`wiki/decisions/0001-eigenpal-adoption.md`](decisions/0001-eigenpal-adoption.md) +
  [`docs/superpowers/specs/2026-06-14-eigenpal-vendor-path-design.md`](../docs/superpowers/specs/2026-06-14-eigenpal-vendor-path-design.md).
- **Wave-3 trigger-gated backend items** — deferred backend items that fire only on their
  recorded triggers. Owner doc: [`wiki/backend/roadmap.md`](backend/roadmap.md) (Wave 3) +
  [`wiki/backend/stage2-evaluation.md`](backend/stage2-evaluation.md).

## Superseded roadmaps (historical record — do not plan from these)

- [`wiki/backend/roadmap.md`](backend/roadmap.md) — Backend Professionalization execution
  tracker. **COMPLETE + Wave Z sealed** (2026-06-12/13). Frozen historical record.
- [`wiki/backlog/roadmap.md`](backlog/roadmap.md) — cross-module "Refactor Roadmap"
  (Plans 3–13, anchors locked 2026-05-11). Historical; its one still-open item (Plan 12)
  is carried forward above.
