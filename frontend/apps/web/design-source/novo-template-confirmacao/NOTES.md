# NOTES — novo-template-confirmacao

**Target route:** `/templates/new?step=5`
**Owning feature:** `frontend/apps/web/src/features/templates/`
**Tier:** Light (no @media, ≤100 lines CSS, no new shared primitives; 1 checkbox → leakage-probe required)

## Audit decisions

See `artifacts/phase0-audit.md`. Headline:

- **Keep:** paper-mock thumbnail (decorative static), code chip (mocked next-code), StatusPill "draft", version "v1.0", template name (real: `state.name`), metadata grid (5 rows).
- **Keep (real):** Perfil row (`state.profileCode` + profile name), Família row (from selectedProfile), Autor row (`useAuthStore` displayName).
- **Keep (computed):** Origem row (`state.startingPoint` + `state.selectedDocxName`), Permissões row (derived from permissionsMode + selections).
- **Cut:** Auto-fill row ("2 campos automáticos") — no placeholder extraction endpoint.
- **Keep (adapted):** "Ao confirmar" ol item 4 — removed hardcoded `QUA-COORD`; generic text.
- **Keep (mocked):** CTA submit → navigate to `/templates` (no API call). Backlog: `confirmacao-backend-submit`.
- **Checkbox gate:** local `useState(false)` — no reducer needed.

## Reused primitives

- `WizardShell`, `WizardFooter`.
- `StatusPill` (status='draft').
- `useAuthStore` (displayName).
- Globals: `card`, `kicker`, `h2`, `caption`, `mono`, `tiny`, `pill`, `code-chip`.

## State

- `confirmed: boolean` → local `useState(false)` — gates submit. Not persisted (no back-nav from Step 5).
- `user.displayName` → `useAuthStore(s => s.user)`.
- All wizard state (name, profileCode, startingPoint, permissionsMode, etc.) → props from TemplateWizardPage.

## Backend gaps → Backlog

| Item | Backlog row |
|---|---|
| Real `POST /api/v1/templates` + permissions/structure wiring | `confirmacao-backend-submit` |

## Leakage probe

Required — `<label>` + `<input type="checkbox">` rendered.
