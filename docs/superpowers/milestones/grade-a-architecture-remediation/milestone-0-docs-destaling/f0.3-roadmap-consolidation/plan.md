# Feature F0.3 — Roadmap Consolidation

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Folder:** `f0.3-roadmap-consolidation`
> **Status:** Implementing

## Source

- Milestone spec row F0.3: *Implement* — "Mark `wiki/backlog/roadmap.md` (May) and
  `wiki/backend/roadmap.md` (June) **historical**; create **one** forward roadmap carrying
  this program + post-v1 progression." *Validate* — "Exactly **1** forward roadmap exists;
  the 2 old roadmaps are clearly labeled historical."
- Governing spec M0 / decision D5 (docs-first).

## Discovery (grounding, 2026-06-14)

Two competing roadmap surfaces today:

| File | What it is | State |
|------|-----------|-------|
| `wiki/backend/roadmap.md` | Backend Professionalization execution tracker (Waves 0–F/Z) | **COMPLETE + sealed** (2026-06-12/13); already self-describes "frozen as the historical record" |
| `wiki/backlog/roadmap.md` | Cross-module "Refactor Roadmap" (Plans 3–13; anchors locked 2026-05-11) | Mostly done; **Plan 12** (7-screen finalization) recorded `open`; Plans 10/13 recorded `done` |

No top-level `wiki/roadmap.md` exists. `wiki/backlog/index.md:8` links the refactor roadmap;
`wiki/README.md` indexes no roadmap. Forward work that must survive consolidation, **by
reference (not re-adjudicated here — HS-6)**:

1. **Grade-A Architecture Remediation (M0–M5)** — the live program. This *is* the forward
   surface; its `README.md` + governing spec already exist. (Note: the backend roadmap's open
   **H-G TRIGGER** at `backend/roadmap.md:211` — template-version-status reader — is already
   **M4** of this program.)
2. **Screen finalization** — refactor-roadmap **Plan 12** (library, novo-documento, templates,
   caixa-aprovacao, documento-publicado, template-editor, novo-template-wizard), recorded `open`.
3. **eigenpal post-v1 packaging / upstream-consolidation** — ADR 0001 + the eigenpal vendor-path
   spec (`docs/superpowers/specs/2026-06-14-eigenpal-vendor-path-design.md`).
4. **Wave-3 trigger-gated backend items** — `backend/roadmap.md:79-81` + `stage2-evaluation.md`.

## Approach

1. **Create** `wiki/roadmap.md` — the single canonical **forward** progression surface.
   Lean + truthful: primary = Grade-A program (link README + spec); a "post-v1 carried-forward"
   section enumerating items 2–4 **by reference**; a "superseded roadmaps" section linking both
   historical files. No status re-adjudication of Plan 10/12/13 — carried as the old roadmap recorded.
2. **Banner** `wiki/backend/roadmap.md` + `wiki/backlog/roadmap.md` at the very top as
   **HISTORICAL — superseded by `wiki/roadmap.md`**, keeping their bodies intact as the record.
3. **Repoint** `wiki/backlog/index.md`: mark the refactor-roadmap link historical, add the
   forward `wiki/roadmap.md` link, bump `Last verified`.

## Acceptance gates (run after edits)

- **Gate A (hard):** exactly **1** forward roadmap — `wiki/roadmap.md` exists and is the only
  roadmap not labeled historical.
- **Gate B (hard):** both `wiki/backend/roadmap.md` and `wiki/backlog/roadmap.md` carry a clear
  top-of-file HISTORICAL/superseded banner pointing at `wiki/roadmap.md`.
- **Gate C:** `wiki/backlog/index.md` points at the forward roadmap; no dangling roadmap link.
- **Gate D:** docs-only; no code; no status of carried-forward items re-adjudicated (HS-6 guard).

## Execution notes

Surgical doc edits + 1 new file. Direct edits (simplicity-first). Durable record → `evidence.md`.
