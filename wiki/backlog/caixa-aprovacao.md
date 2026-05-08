# Backlog: Caixa de Aprovação screen (`/approvals`)

> Last updated: 2026-05-08
> Implementation artifacts: `frontend/apps/web/design-source/caixa-aprovacao/artifacts/`
> Phase 4 review: `artifacts/phase4-review.md`

---

## Deferred features

### Action buttons — signoff / return flows

**Files:** `src/features/approval/components/InboxApprovalCard.tsx:56,63,70`

All three action buttons ("Aprovar e assinar →", "Devolver", "Abrir documento") have `onClick` stubs with TODO comments. They need to be wired to real flows:

- **Aprovar e assinar**: open `SignoffDialog` (already exists at `components/SignoffDialog.tsx`)
- **Devolver**: open rejection dialog or inline comment flow (not yet designed)
- **Abrir documento**: navigate to document viewer route (route exists but needs param wiring)

---

### Heatmap — real data

**File:** `src/features/approval/components/InboxTimeline.tsx:84`

The heatmap sparkline (14-day decision history) is hardcoded with mock values `[3,5,2,7,4,6,8,5,3,9,4,7,5,2]`. Real data requires a new backend endpoint returning per-day signoff counts for the actor over the last 14 days.

Backend: `GET /api/v1/approvals/inbox/activity?days=14` → `{ date: string, count: number }[]`

---

### Timeline item click — open document

**File:** `src/features/approval/components/InboxTimeline.tsx:151-153`

`handleClick` in timeline item rows is a no-op stub. Should navigate to the document viewer or open `InboxApprovalCard` modal for the selected item.

---

### Open queue item from timeline "Revisar →" button

**File:** `src/features/approval/components/InboxTimeline.tsx:197-200`

"Revisar →" button `onClick` is a stub. Should set `selectedIdx` in `InboxPage` and switch view to `stack` so the approval card opens.

---

## Minor issues (from Phase 4.5 review)

### `view` state type narrowing

**File:** `src/features/approval/pages/InboxPage.tsx:10`

`view` state is typed as `string`. Should be `'stack' | 'timeline'` with a validated localStorage initializer to guard against stale/invalid stored values.

```ts
type ViewType = 'stack' | 'timeline';
const [view, setView] = useState<ViewType>(() => {
  const raw = localStorage.getItem('md.inbox.v');
  return raw === 'stack' || raw === 'timeline' ? raw : 'stack';
});
```

---

### Icon: "Abrir documento" should use `eye` not `docs`

**File:** `src/features/approval/components/InboxApprovalCard.tsx:58`

Design source specifies `<Icon name="eye" />`. `eye` is not currently in `IconName` type in `components/ui/Icon.tsx`. Fix: add `eye` SVG to Icon component, then revert to `name="eye"`.

---

### `approvalApi.ts` uses raw `fetch` / plain `Error`

**File:** `src/features/approval/api/approvalApi.ts`

Pre-existing debt. `getJSON` throws `new Error(...)` instead of `ApiError`, so query errors cannot be resolved via `resolveErrorMessage`. Should be migrated to use `lib/api/client.ts` (`openapi-fetch` instance).

---

### `parity-diff.md` Region 5 — InboxTimeline parity methodology gap

The Phase 3b `parity-diff.md` Region 5 (InboxTimeline) was verified via `document.styleSheets` enumeration rather than a live DOM computed-style snapshot, because the design HTML preview (Babel renderer) timed out on animation-heavy content. Live DOM spot-check at Phase 4.5 confirmed correct values. Re-verify when design-source HTML is ported to a non-Babel renderer.
