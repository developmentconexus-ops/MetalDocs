# Novo Template — Etapa 1 · Escopo

## Target route
`/templates/new` (Step 1 of 5)

## Owning feature
`features/templates/`

## Design source
`novo-template-escopo.html` (renders `TplStep1` from `template-wizard.jsx`)

## Context
Step 1 of the template creation wizard. User selects either:
- **Tab A** (default): pick a profile type (POP, IT, POL, DC, CHK, FOR) → blank template
- **Tab B**: clone from an existing published document → pre-fill content

Wizard shares same shell DNA as the document creation wizard (`documents-v2/new`).
`WizardShell` + `WizardFooter` must be promoted from `features/documents/` to `features/shared/`
so both wizards can reuse them.

## Reused primitives
- `components/ui/Stepper` — 5-step indicator
- `components/ui/SelectableCard` — profile card selection
- `components/ui/TabBar` — scope mode toggle
- `components/ui/SearchBar` — document search (Tab B)
- `features/shared/components/wizard/WizardShell` — scrollable layout + header (PROMOTE from documents)
- `features/shared/components/wizard/WizardFooter` — back/next footer (PROMOTE from documents)

## Decisions / cuts

### Cut (confirmed 2026-05-09)
- **"A partir de um documento" tab** — wrong concept (design misconception). Cut entirely. Replaced by `Escopo de aplicação` field in Step 2 (Identidade). Template scope is a filter hint: profile / profile+subtype / generic. Actual doc-to-template link happens at document creation (doc wizard Step 3), not here.

### Deferred
- **Template count per profile card** ("X templates publicados") — no summary endpoint. Show `—`. Same TODO pattern as doc wizard `StepProfile`. Backlog: `wiki/backlog/novo-template-wizard.md`.
- **Permissions step** (Step 4) — no template-audience API. Entire step deferred pending API.
- **"Escopo de aplicação" subtype + generic options** (Step 2) — ships locked to profile default. Subtype option deferred until Taxonomy adds subtypes.

### Kept
All other Step 1 elements backed by real data/state.

## Open backlog
See `wiki/backlog/novo-template-wizard.md` (created at Phase 5).
