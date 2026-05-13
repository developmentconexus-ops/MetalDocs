# NOTES — novo-template-permissoes

**Target route:** `/templates/new?step=4`
**Owning feature:** `frontend/apps/web/src/features/templates/`
**Tier:** Heavy (2-col roles grid + 3-col areas grid → `@media` collapse rules needed)

## Audit decisions

See `artifacts/phase0-audit.md`. Headline:

- **Keep:** 3-mode segmented control (roles | areas | all).
- **Keep (mocked):** Role cards 2-col grid — no personnel-roles API; data hardcoded with TODO.
- **Keep (mocked):** Area cards 3-col grid — area names from design, user counts mocked.
- **Keep (mocked):** "Todos" banner with ~340 company user count.
- **Keep (mocked):** Coverage summary block — count derived from selection; user counts per role/area are mocked.
- **No cuts** — all design elements map to supported patterns; user counts are mocked not fetched.

## Reused primitives

- `WizardShell`, `WizardFooter`.
- `Icon` (`home`, `taxonomy`, `users`, `check`).
- Globals: `card`, `kicker`, `h2`, `caption`, `mono`, `tiny`, `btn btn-sm btn-ghost`.

## State

`features/templates/state/templateWizard.reducer.ts` — extended with:
- `permissionsMode: 'roles' | 'areas' | 'all'` (default `'roles'`)
- `selectedRoleIds: string[]` (default `[]`)
- `selectedAreaIds: string[]` (default `[]`)
- Actions: `SET_PERMISSIONS_MODE`, `TOGGLE_ROLE_ID`, `TOGGLE_AREA_ID`

## Mock data

Hardcoded constants in `StepPermissions.tsx`:
- `MOCK_ROLES` — 6 role cards with user counts (no API)
- `MOCK_AREAS` — 6 area cards with user counts (no API)
- `COMPANY_USER_COUNT = 340` — company-wide count (no API)

All tagged `TODO(novo-template-wizard:permissions-api)`.

## Backend gaps → Backlog

| Item | Backlog row |
|---|---|
| Personnel roles list with user counts | `permissions-roles-api` |
| Area user counts | `permissions-area-counts` |
| Company-wide active user count | `permissions-user-count` |

## Artifacts

- `artifacts/phase0-audit.md`
- `artifacts/phase1-map.md`
- `artifacts/phase3a-structure.md`
- `artifacts/phase3b-style.md` (subagent)
- `artifacts/parity-diff.md` (subagent)
- `artifacts/phase4-behavior.md`
- `artifacts/phase4-review.md`
