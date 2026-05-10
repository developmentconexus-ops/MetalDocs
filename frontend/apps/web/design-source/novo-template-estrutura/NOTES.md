# NOTES — novo-template-estrutura

**Target route:** `/templates-v2/new?step=3`
**Owning feature:** `frontend/apps/web/src/features/templates/`
**Tier:** Light (per `metaldocs-screen-implementation` skill v1.3)

## Audit decisions

See `artifacts/phase0-audit.md`. Headline:

- **Cut:** placeholders block (7 chips, kind, ⚡ auto-fill flag, counters). No backend extract endpoint; `auto-fill` not in API.
- **Cut:** file metadata "147 KB · processado em 1.2s" — implies server processing.
- **Cut:** info banner "ℹ️ próximos passos refinados no editor" — editor handoff not implemented.
- **Mocked:** docx selection (filename + size echo only). Real upload deferred to editor handoff post-create.
- **Keep:** 2 starting-point cards + selected-file row (mock).

## Reused primitives

- `WizardShell`, `WizardFooter`.
- Globals: `card`, `kicker`, `h2`, `caption`, `mono`, `btn btn-sm btn-ghost`.

## State

`features/templates/state/templateWizard.reducer.ts` — extended with:
- `startingPoint: 'docx' | 'blank' | null`
- `selectedDocxName: string | null` (mock)
- `selectedDocxSize: number | null` (mock)
- Actions: `SET_STARTING_POINT`, `SET_SELECTED_DOCX`, `CLEAR_SELECTED_DOCX`.

## Backend gaps → Backlog

| Item | Backlog row |
|---|---|
| Real docx presigned upload | `step3-docx-upload` |
| Placeholder extraction endpoint | `step3-placeholder-extract` |
| Editor handoff post-create | `step3-editor-handoff` |

## Artifacts

- `artifacts/phase0-audit.md`
- `artifacts/phase1-map.md`
