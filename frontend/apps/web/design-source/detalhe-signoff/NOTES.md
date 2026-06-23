# Detalhe Signoff — Implementation Notes (F5.1)

Screen: `/approvals/:documentId` cockpit. Implemented as **pure consumer-side
assembly** (no new backend, no fork of existing tested approval components) per the
F5.1 plan. This file records the Keep / Cut / Defer audit of every element in the
design reference (`detalhe-signoff.html`) against what shipped, so omissions are
explicit and traceable rather than silent.

## Keep (implemented to parity)
- **Document header** — code chip (`CodeChip`), status pill (`StatusPill`), revision
  version, document title.
- **Tab strip** — Documento / Mudanças vs versão anterior / Comentários.
- **A4 document panel** — embeds the real rendered PDF via `useDocumentPdfStatus`
  (`GET /documents/{id}/view`) with honest pending / failed (retry) states.
- **Comentários tab** — live from `useDocumentCommentsQuery`
  (`GET /documents/{id}/comments`); loading / error / empty / list states.
- **Decision surface (right sidebar)** — the existing `ControlledDocumentDetailPanel`
  (Assinar / Cancelar + approval timeline + integrity metadata) is **mounted, not
  rebuilt**. This is the no-fork mandate: the design's bespoke right-panel markup is
  intentionally substituted by the canonical tested component. Visual deltas vs the
  design panel are accepted as the cost of the no-fork rule.

## Defer (design element present, shipped as honest placeholder, with trigger)
- **"Mudanças vs versão anterior" diff tab** — no document-diff backend exists
  (`/documents/{id}/view` returns a single rendered PDF; nothing compares the
  under-review revision to the published one). The tab renders an honest explanation
  and **no fabricated diff**. Unblock trigger + ownership recorded in
  `wiki/backlog/detalhe-signoff.md`.

## Cut (design element intentionally not shipped in F5.1; rationale below)
- **"Trilha" (audit-trail) tab** — the design exposes the approval trail as a 4th
  top tab. The same information (stage-by-stage approval timeline) is already rendered
  by `ApprovalTimelinePanel` **inside** the mounted `ControlledDocumentDetailPanel`
  (right sidebar). Duplicating it as a left-column tab would fork timeline rendering —
  forbidden by the no-fork mandate. Cut as redundant, not as missing. If a standalone
  full-width audit view is later required, add it as a tab that reuses
  `ApprovalTimelinePanel` (no new timeline component).
- **Rich header sub-elements** — the design header also carries a workflow kicker
  ("§ Aprovação · etapa X de Y · revisão final"), a deadline indicator ("vence hoje
  18:00"), and an author/submitter row (avatar, name, area, submitted-at, prior
  approver). These depend on instance/stage + submitter data that is not cleanly
  available from the queries this cockpit already consumes
  (`getDocument` + `getActiveDocumentContext` expose neither stage-count nor a
  due-date nor submitter identity). Cut from F5.1 to avoid fabricating values. Wiring
  the kicker + deadline becomes feasible when the active-instance payload (or a
  dedicated endpoint) surfaces stage index/total and a due date; tracked in
  `wiki/backlog/detalhe-signoff.md`.

## Token / styling notes
- Page styling uses the wine design tokens from `src/styles/tokens.css`
  (`--brand` for the active-tab indicator, `--surface` / `--border` / `--text` /
  `--text-muted` / `--success`), with same-family hex fallbacks. Layout is full-bleed
  (no outer page padding); each column owns its internal padding, matching the design's
  flush working-surface layout.
- `ControlledDocumentDetailPanel.module.css` (the reused panel) still carries
  pre-existing non-token hex values and other pre-existing UX debt (native
  `window.prompt` for the cancel reason, non-`resolveErrorMessage` error strings,
  >400 LOC). These predate F5.1 and were **not** modified here (additive props only);
  they are recorded as bounded defers in `wiki/backlog/detalhe-signoff.md` for a
  follow-up that owns that component.
