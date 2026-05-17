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

## Plan 12.4 integration summary (2026-05-15)

This directory is the canonical worksheet for the requested `novo-template-wizard` screen. The repo does not contain `design-source/novo-template-wizard/`; the committed wizard artifacts are split as `design-source/novo-template-*`.

**Implement now:**
- Keep profile and generic template creation.
- Pass `doc_type_code` when a profile-scoped template is selected.
- Use the shared API client + idempotency key for the active create path.
- Replace fake `TPL-*` code previews with the actual slug key currently submitted by the wizard.
- Keep blank-start editor handoff.

**Cut/defer for this PR:**
- Disable `.docx` import in the wizard until there is a real mid-flow upload or post-create handoff contract.
- Replace mocked Step 4 role/area/user-count permissions with an honest read-only public-visibility state.
- Preserve next-code, key UX, template counts, CHK enabled flag, placeholder extraction, permissions APIs, and visibility create-body support in `wiki/backlog/novo-template-wizard.md`.

**Final code truth:**
- Runtime smoke created a profile-scoped template using `doc_type_code = qa_seed`, `Idempotency-Key`, and the real `POST /api/v1/templates` route.
- The flow now redirects to `/templates/{id}/versions/1` after HTTP 201.
- A create-path backend prerequisite was fixed in `templates`: the transaction now sets tenant/actor authz GUCs before `authz.Require`.
- Screenshots and `smoke-report.json` live in `artifacts/screenshots/plan-12-4/`.
