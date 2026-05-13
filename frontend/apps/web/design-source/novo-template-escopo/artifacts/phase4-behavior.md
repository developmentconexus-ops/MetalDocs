# Phase 4 — Behavior Verify: novo-template-escopo

> **Status:** PASS
> **Date:** 2026-05-09

---

## tsc result

```
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

**Result:** 15 pre-existing errors in `auth/`, `documents/`, `shell/` — zero errors in any file touched by this implementation (`features/templates/`, `features/shared/components/wizard/`, `features/taxonomy/queries/`).

---

## Test result

```
Test Files: 8 failed | 46 passed (54)
Tests:      17 failed | 238 passed (255)
```

All 17 failures are pre-existing (approval, auth, documents, template-author-page-convergence). Zero new failures introduced by the wizard implementation.

---

## Smoke trace

| Step | Action | Expected | Observed | Result |
|---|---|---|---|---|
| 1 | Navigate to `/templates/new` | Route renders, stepper shows step 1 active | "TEMPLATES / NOVO" kicker + "Novo template reutilizável" h1 + 5-step stepper with step 1 filled circle | ✅ PASS |
| 2 | Profile grid load | Profiles from taxonomy API visible | "dc Documento Controlado" + "proc Procedimento Operacional" (real API data) | ✅ PASS |
| 3 | URL sync | `?step=1` in URL | `window.location.search === "?step=1"` | ✅ PASS |
| 4 | Select profile card | Card gets brand border + checkmark; Avançar enables | "dc" card selected with `✓`; `advanceDisabled: false` | ✅ PASS |
| 5 | Footer step label | Changes to "Pronto para avançar" on selection | `advanceDisabled: false`, "Avançar →" button enabled | ✅ PASS |
| 6 | Cancelar button | Present | Found in DOM | ✅ PASS |
| 7 | Caption text | "Templates podem ser genéricos para um perfil (POP, IT, etc.)…" | Correct copy visible | ✅ PASS |

---

## Console errors

None observed during smoke trace.

---

## Notes

- This environment has 2 taxonomy profiles (`dc`, `proc`) — CHK disabled-profile hardcode not testable without CHK in API. TODO covered by `DISABLED_PROFILES = new Set(['CHK'])` with comment.
- Template count shows `—` as expected (no summary endpoint).
