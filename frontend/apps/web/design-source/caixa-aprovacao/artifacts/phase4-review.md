# Phase 4.5 — Visual Review
> Screen: Caixa de Aprovação (`caixa-aprovacao`)
> Reviewer: `frontend-screen-reviewer` agent
> Date: 2026-05-08

**Verdict: REQUEST CHANGES**

---

## Critical

None.

---

## Major

### M1 — Mobile layout collapse (no responsive breakpoint)

**File:** `src/features/approval/components/InboxStack.module.css` (no `@media` query present)

At 375px the root grid resolves to `grid-template-columns: 320px 0px`, making the card area, counter, prev/next buttons, approval card, action buttons, and keyboard hints completely inaccessible. Phase 3b artifact acknowledged the gap but Phase 3c was completed without implementing it.

Suggested fix:
```css
@media (max-width: 768px) {
  .root { grid-template-columns: 1fr; grid-template-rows: auto 1fr; }
  .queueRail { border-right: none; border-bottom: 1px solid var(--border); max-height: 220px; }
}
```

### M2 — `deadlineBlock` always rendered on non-urgent items

**File:** `src/features/approval/components/InboxApprovalCard.tsx:20-24`

The VENCE EM / countdown block is always rendered regardless of `item.urgent`. Design source (`inbox-alts.jsx:117`) renders it conditionally: `{item.urgent && (...)}`. Non-urgent cards (POL-RH-0011 confirmed via live DOM) show "VENCE EM / 1 dia" in the top-right header — violates the urgency-signal hierarchy.

Suggested fix:
```tsx
{item.urgent && (
  <div className={styles.deadlineBlock}>...</div>
)}
```

### M3 — Raw hex fallback in CSS Module

**File:** `src/features/approval/components/InboxStack.module.css:218`

`color: var(--danger, #e5534b)` — raw hex fallback violates token-only rule. `--danger` is defined in `tokens.css`, so the fallback never fires at runtime, but it must not exist in source.

Fix: `color: var(--danger);`

---

## Minor

**m1** — `InboxPage.tsx:10` — `view` state typed as `string`, should be `'stack' | 'timeline'` union with a validated localStorage initializer.

**m2** — `parity-diff.md` Region 5 — InboxTimeline parity verified via `document.styleSheets` enumeration (not live DOM snapshot) due to design-HTML Babel timeout blocker. Live spot-check in this review confirms correct values. Accept as-is, note in backlog.

**m3** — `InboxApprovalCard.tsx:56` — "Abrir documento" uses `<Icon name="docs" />`, design specifies `<Icon name="eye" />`. Phase 3a notes `eye` is not in `IconName` type — correct fix is to add `eye` to the Icon component rather than substitute `docs`.

---

## What's Good

- Architecture: feature fully contained in `features/approval/`, TanStack Query via `useInboxQuery`, `createBrowserRouter` integration correct.
- Error UX: `isError` → `<div role="alert">`, no raw `alert()`, loading + empty + error states all implemented.
- Mock data trail: every mock field has `// TODO [BACKLOG: caixa-aprovacao.md]` co-located.
- Timeline view: deadline-bucketed layout with pulsing urgent dot, heatmap sparkline, `role="button" tabIndex={0} onKeyDown` keyboard-accessible rows — semantically sound.
- CSS token compliance: near-perfect across both CSS Modules except M3.

---

## Required Before Merge

- [ ] M1: Add `@media (max-width: 768px)` breakpoint to `InboxStack.module.css`
- [ ] M2: Wrap `deadlineBlock` in `{item.urgent && (...)}`
- [ ] M3: Remove raw hex fallback — `color: var(--danger);`

Minor items → `wiki/backlog/caixa-aprovacao.md`.
