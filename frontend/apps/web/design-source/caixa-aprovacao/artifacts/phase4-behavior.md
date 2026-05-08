# Phase 4 — Behavior Verify
> Screen: Caixa de Aprovação (`caixa-aprovacao`)
> Date: 2026-05-08

---

## TSC Result

```
pnpm tsc --noEmit -p tsconfig.build.json
```

**Result: 0 errors in `features/approval/`.**

Pre-existing errors (unchanged from Phase 3a baseline) in unrelated features:
- `src/features/auth/__tests__/useAuthSession.returnTo.test.tsx` — TS2554
- `src/features/documents/components/LibrarySidebar.tsx` — TS2339, TS7006
- `src/features/documents/pages/NewDocumentWizardPage.tsx` — TS2339, TS7006, TS2740
- `src/features/documents/queries/__tests__/useAreasQuery.test.ts` — TS7053
- `src/features/documents/queries/useAreasQuery.ts` — TS2769
- `src/features/shell/components/Rail.tsx` — TS2322

---

## Test Result

```
pnpm test --run src/features/approval/
```

**Approval feature tests: all passing**

| File | Tests | Result |
|---|---|---|
| `approval/api/mutationClient.test.ts` | 7 | ✓ PASS |
| `approval/components/SupersedePublishDialog.test.tsx` | 5 | ✓ PASS |
| `approval/components/SignoffDialog.test.tsx` | 9 | ✓ PASS |
| `approval/pages/InboxPage.test.tsx` | 6 | ✓ PASS (rewritten for new UI) |
| `approval/pages/RouteAdminPage.test.tsx` | 4/5 | 1 pre-existing failure (unrelated to inbox) |

`RouteAdminPage` failure confirmed pre-existing (same failure before Phase 3c changes).

`InboxPage.test.tsx` was rewritten — old tests covered the legacy table-based UI which was replaced. New tests cover:
- Loading state
- Error state (role="alert")
- Empty API → MOCK_INBOX_ITEMS fallback (4 items, counter 01/04)
- API items rendered in queue rail
- View switcher persists to localStorage
- Next/prev navigation updates counter

---

## Manual Smoke Steps

Verified at `http://localhost:4174/approvals`:

| Step | Expected | Observed |
|---|---|---|
| Page loads | Toolbar + stack view | ✓ |
| Queue rail shows 4 items | POP-QUA-0148, IT-PROD-0072, POL-RH-0011, DC-TI-0203 | ✓ |
| Item 01 active (branded left border) | Yes | ✓ |
| Card shows POP-QUA-0148 data | Dark header, POP badge, summary, stats, 3 action buttons | ✓ |
| Counter shows "01 / 04" | Yes | ✓ |
| Click queue item 02 | Counter → "02 / 04", card updates | ✓ (wired in Phase 3c) |
| Click "próximo →" | Counter advances | ✓ |
| Click "← anterior" at 01 | Button disabled | ✓ |
| Click "Linha do tempo" switcher | Timeline view renders | ✓ |
| Refresh with timeline view | Persisted via localStorage | ✓ |
| Click "Foco" switcher | Returns to stack view | ✓ |
| Keyboard ←/→ | Navigates queue items | ✓ |
| Keyboard hints bar | "A Aprovar · D Devolver · ←/→ Navegar" with kbd tiles | ✓ |
| Ghost cards | Removed (per user request) | ✓ |

---

## Console Errors

None observed during smoke session.

---

## Notes

- Action buttons (Aprovar e assinar, Devolver, Abrir documento) have TODO-tagged onClick stubs → tracked in `wiki/backlog/caixa-aprovacao.md`
- Heatmap in timeline view is hardcoded → tracked in backlog
- `approvalApi.ts` uses raw fetch (pre-existing debt) → separate backlog item
