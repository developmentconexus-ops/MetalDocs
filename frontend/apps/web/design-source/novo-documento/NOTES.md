# Novo Documento (Wizard) — Design Notes

> **Slug:** novo-documento
> **Type:** Multi-step wizard (4 panels in one screen)
> **Owning feature:** `features/documents/`
> **Target route:** `/documents-v2/new` (replaces existing 1-step `DocumentCreatePage`)
> **Reference HTMLs (one per step):**
> - `../novo-perfil/novo-perfil.html` — Step 1 (profile)
> - `../novo-area-codigo-visibilidade/novo-area-codigo-visibilidade.html` — Step 2 (area + title + code preview + visibility)
> - `../novo-template/novo-template.html` — Step 3 (template choice)
> - `../novo-confirmacao/novo-confirmacao.html` — Step 4 (review + create)
> - `../selected-wizard.jsx` + `../selected-wizard-v2.jsx` — older AI mockups (treat as secondary reference; conflict with current HTMLs in spots)

## Purpose

Operator creates a new controlled document. Wizard collects: profile → area + title + visibility → template → confirm. Submit triggers two-call sequence (create slot, then create draft document). User lands on the editor for the new draft.

## Personas
- **Author** (`role=author`) — primary user. Has `registry.create` + `doc.create` + `doc.edit`.
- **Editor** (`role=editor`) — also has `registry.create` + `doc.create`. Same wizard UX.
- **Approver / Viewer** — no `registry.create` cap → wizard route hidden.

## Audit decision (per user, 2026-05-06)

User policy: **render every designed feature**, even features without backend support. Disabled-state controls + no-op submit + TODO trail + matching `wiki/backlog/<screen>.md` row. Same pattern as Library + EditorMetaSidebar mocks.

| Element | UI behavior | Submit behavior | Trail |
|---|---|---|---|
| Step 1 — Profile cards (real list from `/api/v2/profiles`) | Render + selectable | `profileCode` sent in `POST /controlled-documents` payload | — |
| Profile card "count" badge (mock "12 docs") | Render `—` placeholder | — | TODO — needs aggregate endpoint |
| Profile card "soon" / disabled badge | **DROP** — no model field, would mislead | — | — |
| Step 2 — Area `<select>` (real list from `/api/v2/process-areas`) | Render + selectable | `processAreaCode` sent | — |
| Step 2 — Title input | Render + required | `title` + `name` sent | — |
| Step 2 — Code preview "POP-QUA-120" | Render `≈ POP-QUA-???` with tooltip "Código final atribuído ao confirmar" | Server assigns | — |
| Step 2 — Visibility radios (area / people / company / external) | Render + selectable, default "Toda empresa" | **NOT submitted** (no backend field today) | TODO — needs visibility model |
| Step 2 — People-invite chips (when visibility=people) | Render + interactive | No-op | TODO — needs share model |
| Step 2 — External: password / watermark / expiry | Render + interactive | No-op | TODO — needs share model |
| Step 3 — Template cards per profile (real list from `/api/v2/templates?profile=X`) | Render + selectable, default = published version | `template_version_id` sent | — |
| Step 3 — Prior version radios | Render but **only published version selectable**; older versions grayed + tooltip "Em breve" | Default sent | TODO — needs `GET /api/v2/templates/:id/versions` |
| Step 3 — "Em branco" template option | Render + grayed disabled + tooltip "Em breve" | Not selectable | TODO — needs no-template POST path |
| Step 4 — Summary card (Perfil / Área / Código / Título / Visibilidade / Template) | Render derived from prior steps | Read-only | — |
| Step 4 — Author + Created at | Render (`useAuthStore` user + `new Date()`) | Set server-side | — |
| Step 4 — "Confirmo que entendi" checkbox | Render + required to enable submit | Frontend gate only | — |
| Step 4 — "Criar documento" button | Render + primary | Two-call sequence: `POST /controlled-documents` → `POST /documents` | — |
| Stepper header (4 dots / step labels) | Render — generic primitive `components/ui/Stepper.tsx` | URL `?step=N` | — |
| Cancel / back buttons | Render | URL nav back | — |

## State map (across steps)

```
formState = {
  profile: { code, familyCode } | null,                       // Step 1
  area:    { code, name } | null,                             // Step 2
  title:   string,                                            // Step 2
  visibility: 'area' | 'people' | 'company' | 'external',    // Step 2 (NOT submitted)
  invitees: { userId, displayName }[],                       // Step 2 visibility=people (no-op)
  external: { password?: string, watermark: bool, expiresAt?: Date }, // Step 2 visibility=external (no-op)
  templateChoice: { kind: 'default' | 'specific' | 'blank', templateID?, versionID? }, // Step 3
  consent: bool,                                              // Step 4
}
```

Step deps:
- Step 2 areas list = global (not filtered by profile today)
- Step 3 templates list = filtered by `profile.code` (filter param `doc_type`)
- Step 4 = pure read of accumulated state

## Routing

Single route `/documents-v2/new` with `?step=1..4` URL param. Internal step state via `useReducer`. Refresh-safe + back-button works. URL also accepts pre-fill query params (e.g. `?profile=POP`) for entry from profile detail screens later.

`features/documents/components/LibrarySidebar.tsx:59` currently navigates to broken `/documents/new` — **fix to `/documents-v2/new`** in this PR.

## Backend prereqs (for `wiki/backlog/novo-documento.md`)

- **Visibility model.** Define enum + storage on `controlled_documents` (or `documents`). Design uses 4 options (area/people/company/external) — different from earlier 3-option (`public/area/restricted`) attempted in 0164. Decide before backlog can be cleared.
- **Share controls** — invitee list + external password/watermark/expiry. Schema work + sharing module.
- **Sequence preview endpoint** — `GET /api/v2/controlled-documents/next-code?profile=X&area=Y` (optional UX polish; placeholder works fine in v1).
- **Template versions list** — `GET /api/v2/templates/:id/versions` returning `[{ID, VersionNum, IsPublished, CreatedAt, Label}]`.
- **No-template POST path** — `POST /documents` with `template_version_id: null` allowed; wizard surfaces "Em branco" option.
- **Slot rollback / atomic create** — if `POST /documents` fails after `POST /controlled-documents` succeeded, document the orphan-slot recovery path. Today: orphan stays + code consumed.
- **Profile doc-count aggregate** — for Step 1 card badge. Low priority.

## Open questions log (Phase 0)

- [Resolved 2026-05-06] Visibility default = "Toda empresa" (matches design pre-selection + standard QMS convention). Stored in form state, not submitted.
- [Resolved 2026-05-06] Code preview = `≈ POP-QUA-???` with tooltip.
- [Resolved 2026-05-06] Profile "soon" badge dropped (no consistent model).
- [Resolved 2026-05-06] Profile per-card count = `—` placeholder + TODO.
