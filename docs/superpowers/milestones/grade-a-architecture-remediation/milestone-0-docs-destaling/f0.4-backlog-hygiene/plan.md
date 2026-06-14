# Feature F0.4 — Backlog Hygiene

> **Milestone:** 0 — Docs Progression De-Staling  ·  **Folder:** `f0.4-backlog-hygiene`
> **Status:** Implementing

## Source

- Milestone spec row F0.4: *Implement* — "Close/archive completed items in `wiki/backlog/*`;
  keep only active deferred work." *Validate* — "`wiki/backlog/*` contains active deferred work
  only; closed items archived, not deleted-without-trace."
- **Operator ruling (2026-06-14):** archive rule = **only fully-closed files** (a file with
  **zero remaining active defers**). Any file with an `open` row or a live deferred-stub item
  stays. (Per-item splitting and "all work-done files" both rejected.)

## Discovery (evidence-based census, 2026-06-14)

Read every `wiki/backlog/*.md` (Explore sweep + per-row verification). Classification under the
operator's fully-closed rule:

| File | Active `open` rows / live defers? | Disposition |
|------|-----------------------------------|-------------|
| `api-contract-hardening.md` | **none** — "Phase F shipped, PROGRAM CLOSED"; closing 4-dim re-audit 0 CRITICAL/0 HIGH; all ledger findings `closed`. (The lone `open` grep hit is description text `open-by-default` in a row whose status = `closed 2026-06-05`.) | **CLOSED → archive** (F0.5 relocates) |
| `contract-first-followups.md` | none — but **superseded**: "folded into api-contract-hardening Phase C/E" | **Superseded → F0.5** (superseded-historical class, not "completed") |
| `approval-refactor.md` | R-002/004/005/007/009/010/011/012 `open` | stays (active) |
| `audit-refactor.md`, `auth-refactor.md`, `controlled-documents-refactor.md`, `documents-refactor.md`, `editor-chrome-refactor.md`, `editor-ui-eigenpal-refactor.md`, `frontend-primitives-refactor.md`, `iam-refactor.md`, `novo-documento-wizard-refactor.md`, `render-fanout-refactor.md`, `search-refactor.md`, `taxonomy-refactor.md`, `templates-refactor.md` | one or more `open` rows | stay (active) |
| `caixa-aprovacao.md`, `distribuicao.md`, `documento-publicado.md`, `editor.md`, `library-screen.md`, `novo-documento.md`, `novo-template-wizard.md`, `template-editor.md`, `templates.md` | implemented but carry live **deferred** items (intentional stubs / missing-backend capability) | stay (active deferred work) |
| `planned-endpoints.md` | forward sketches (notifications, workflow transitions/approvals) — active | stays |
| `roadmap.md`, `index.md` | handled in F0.3 / this feature | n/a |

**Net:** exactly **one** fully-closed backlog file — `api-contract-hardening.md`. Everything else
holds active deferred work (rule A). `contract-first-followups.md` is *superseded*, routed to F0.5.

## F0.4 / F0.5 boundary (coupling, explicit)

`wiki/_archive/` does not exist yet — **F0.5** owns creating it and the physical relocation +
governance-map rows. So F0.4 does the **closure judgment** and marks closed-in-place; F0.5 does the
**move**. "Not deleted-without-trace" is guaranteed by the CLOSED banner (F0.4) + the governance
migration-map row (F0.5). No file is moved in F0.4.

## Approach

1. **Banner** `api-contract-hardening.md` at top: ⚠️ CLOSED — completed program; to be relocated to
   `wiki/_archive/` in F0.5. Body retained intact.
2. **Hygiene-annotate** `wiki/backlog/index.md`: split the program-level list into **active** vs
   **closed/superseded**; mark `api-contract-hardening` closed and `contract-first-followups`
   superseded (both → `_archive/` in F0.5); bump `Last verified`.
3. Hand the closed/superseded set to F0.5 for physical relocation (recorded as bounded defer).

## Acceptance gates (run after edits)

- **Gate A:** every fully-closed backlog file is identified and marked CLOSED — census above is
  complete and each row is evidence-backed.
- **Gate B:** `wiki/backlog/index.md` unambiguously separates active deferred work from
  closed/superseded items; no closed item presented as active.
- **Gate C:** no backlog file deleted; docs-only; no code. Relocation deferred to F0.5 with trigger.

## Execution notes

Surgical: 1 banner + 1 index edit. Direct edits. Durable record → `evidence.md`.
