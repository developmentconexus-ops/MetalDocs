# NOTES — novo-template-identidade

**Target route:** `/templates/new?step=2`
**Owning feature:** `frontend/apps/web/src/features/templates/`
**Tier:** Light (per `metaldocs-screen-implementation` skill v1.3)

## Audit decisions

See `artifacts/phase0-audit.md` for the full Keep/Cut/Defer table. Headline:

- **Cut:** tags field (no taxonomy yet; not in backend contract).
- **Defer (mocked):** code preview — backlog row `novo-template-wizard:next-code-preview`.
- **Defer:** `key` field — auto-stub from name slug at submit (Step 5). Backlog row `novo-template-wizard:key-generation`.

## Reused primitives

- `features/shared/components/wizard/WizardShell` — page shell + stepper.
- `features/shared/components/wizard/WizardFooter` — back/advance bar.
- Globals: `card`, `kicker`, `h2`, `caption`, `mono`, `code-chip`, `input`, `btn btn-sm btn-ghost`.

## State

`features/templates/state/templateWizard.reducer.ts` — extended with `name`, `description` + `SET_NAME` / `SET_DESCRIPTION`.

## Backend contract gaps

| Field | Status |
|---|---|
| `name` | exists — `POST /api/v1/templates` body |
| `description` | exists — same body |
| code preview | mocked client-side; needs `GET /api/v1/templates/next-code` |
| `key` | deferred — auto-slug from name at Step 5 submit |
| tags | cut — no backend taxonomy |

## Artifacts

- `artifacts/phase0-audit.md`
- `artifacts/phase1-map.md`
- `artifacts/phase3-combined.md`
- `artifacts/parity-diff.md`
- `artifacts/leakage-probe.md`
- `artifacts/phase4-behavior.md`
- `artifacts/phase4-review.md`
