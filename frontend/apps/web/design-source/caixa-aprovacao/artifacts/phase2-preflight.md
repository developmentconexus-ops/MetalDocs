# Phase 2 — Pre-flight Artifact
> Screen: Caixa de Aprovação (`caixa-aprovacao`)
> Date: 2026-05-08

---

## 1. Codegen

**No** — `ListInboxResponse` and `InboxItem` are already hand-authored in
`src/features/approval/api/approvalTypes.ts`. No OpenAPI spec change needed for
this screen. Mock fields (`code`, `kind`, `deadline`, `urgent`, `summary`,
`changes`, `version`) are handled in `mockInboxData.ts` pending backend work
tracked in `wiki/backlog/caixa-aprovacao.md`.

---

## 2. Primitives modified

### Avatar — xs size added

File: `src/components/ui/Avatar.tsx` + `Avatar.module.css`

- Added `'xs'` to `AvatarSize` union type.
- Added `xs` entry to `sizeClass` record.
- Added `.xs` rule in `Avatar.module.css`: `width: var(--sp-5); height: var(--sp-5)` (20px),
  `font-family: var(--font-mono); font-size: 0.5rem`.

### CSS token violations fixed in same commit

| Location | Old value | Fixed to |
|---|---|---|
| `.avatar` rule — `color` | `#fff` | `var(--text-on-brand)` |

### Remaining raw-px notes (pre-existing, no matching token)

| Property | Value | Note |
|---|---|---|
| `.avatar` — `width/height` | `28px` | No `--sp-*` maps to 28px; nearest are 24/32. Not changed (pre-existing). |
| `.sm` — `width/height` | `22px` | No `--sp-*` maps to 22px. Not changed (pre-existing). |
| `.lg` — `width/height` | `36px` | No `--sp-*` maps to 36px; nearest are 32/40. Not changed (pre-existing). |
| Font sizes | `0.68rem`, `0.58rem`, `0.85rem` | rem units, not px — not subject to px audit rule. |

Pre-existing dimensions are not raw-px violations in the spirit of the rule
(no design-system size token exists for these specific values). Tracked here
for future token expansion.

---

## 3. Icon + CodeChip audit

### Icon.tsx — clean

`src/components/ui/Icon.tsx` has no CSS file. All styling is via inline SVG props
(`width`, `height`, `viewBox`) passed as component props. No CSS token violations.
Status: **CLEAN — no changes needed**.

### CodeChip.tsx — clean (with note)

`src/components/ui/CodeChip.tsx` has no CSS module. It applies global classes
`.code-chip` and `.mono`. Neither class is defined as a standalone rule in
`styles.css`; `.mono` is defined at line 1097 using `font-family: "DM Mono", Consolas, monospace`
(raw font string, not `var(--font-mono)` — pre-existing, in a global utility class).
The component itself contains no CSS to audit.
Status: **CLEAN — no changes needed**. (`.mono` global class pre-existing drift
is out of scope for this phase.)

---

## 4. Global CSS Leakage Map

Rules from `src/styles/base.css` and `src/styles.css` that target bare elements
(without a component-scoping ancestor class). Phase 3b must reset any of these
that could bleed into the approval page CSS Module.

### From `src/styles/base.css`

| Selector | Properties set | Relevance to approval page |
|---|---|---|
| `button, input, select, textarea` | `font: inherit` | Affects all button/input/select/textarea |
| `button` | `cursor: pointer` | Affects all buttons |
| `button:disabled` | `opacity: 0.5; cursor: not-allowed` | Affects disabled buttons |
| `input:not([type="checkbox"]):not([type="radio"]):not([type="file"]):not([type="range"]):not([type="color"]):not([type="image"]), select, textarea` | `width: 100%; border-radius: var(--r-2); border: 1px solid var(--border); background: var(--surface); color: var(--text-soft); padding: 0.5rem 0.75rem` | HIGH — any bare input/select/textarea inside the page layout will get full-width + border |
| `textarea` | `resize: vertical` | Affects any textarea |

### From `src/styles.css`

| Selector | Properties set | Relevance |
|---|---|---|
| `table` | `width: 100%; border-collapse: collapse` | Affects any bare `<table>` |
| `th, td` | `padding: 0.95rem 1rem; border-bottom: 1px solid var(--border); text-align: left; vertical-align: top` | Affects table cells |
| `tr` | `cursor: pointer` | Affects any table row |
| `tr:hover, .row-active` | `background: var(--brand-pale)` | Affects table row hover |

### Scoped rules (NOT leaking — included for completeness)

These target bare elements but are safely scoped behind a component ancestor class:

| Selector | Scoped under |
|---|---|
| `.hero h1` | `.hero` |
| `.create-doc-docx-dropzone input` | `.create-doc-docx-dropzone` |
| `.content-builder-field input, select, textarea` | `.content-builder-field` |
| `.content-builder-table-row select, input` | `.content-builder-table-row` |
| `.content-builder-checklist-row input[type="checkbox"]` | `.content-builder-checklist-row` |
| `.create-doc-content input, select, textarea` | `.create-doc-content` |
| `.metadata-row input` | `.metadata-row` |
| `.create-doc-page-title p` | `.create-doc-page-title` |
| `.create-doc-page-title h2` | `.create-doc-page-title` |
| `.panel-heading h2, .card h3` | `.panel-heading`, `.card` |
| `.create-doc-section-head h3` | `.create-doc-section-head` |

### Phase 3b action required

The approval page CSS Module must reset:
- `input` width: if any inline input is used (e.g. filter field), override `width` explicitly.
- `button` cursor: already `pointer` — no issue, but `opacity: 0.5` on disabled needs to match design.
- No table elements used in approval page layout — table leakage is not a concern.

---

## 5. Query hook path

`src/features/approval/queries/useInboxQuery.ts`

Wraps `listInbox(params: ListInboxParams)` with `QK.inbox(params)` as query key.
Type note: `QK.inbox` declares `InboxParams` (uses `areaFilter`, `page`, `onlyOverdue`);
`ListInboxParams` (uses `area_code`, `limit`, `offset`) is structurally compatible as a
variable assignment — TypeScript accepts it without error (confirmed by tsc run).

---

## 6. Mock lib path

`src/features/approval/lib/mockInboxData.ts`

Exports:
- `RichInboxItem` interface (extends `InboxItem` with 7 mock fields, all `TODO [BACKLOG]` tagged)
- `enrichInboxItem(item: InboxItem, idx: number): RichInboxItem`
- 5 mock extra entries in `MOCK_EXTRAS` (wraps around for lists > 5 items)

---

## 7. Route stub

**Confirmed existing** — `features/approval/routes.tsx` already has:

```
{ path: "approvals", handle: { workspaceView: "approvals" }, lazy: () => import("./pages/InboxPage")... }
```

No commit needed.

---

## 8. tsc result

**Pre-existing errors only — none introduced by Phase 2 changes.**

Errors present before and after Phase 2 (confirmed via `git stash` round-trip):

| File | Error |
|---|---|
| `src/features/auth/__tests__/useAuthSession.returnTo.test.tsx` | TS2554 Expected 0 arguments, but got 1 (×2) |
| `src/features/documents/components/LibrarySidebar.tsx` | TS2339/TS7006 area map/type issues (×2) |
| `src/features/documents/pages/NewDocumentWizardPage.tsx` | TS2339/TS7006/TS2740 area find/type issues (×3) |
| `src/features/documents/queries/__tests__/useAreasQuery.test.ts` | TS7053 index type issues (×3) |
| `src/features/documents/queries/useAreasQuery.ts` | TS2769 QueryFunction overload mismatch |
| `src/features/shell/components/Rail.tsx` | TS2322 string/undefined type |

None of these touch approval, Avatar, or any file modified in Phase 2.
Phase 2 files (`Avatar.tsx`, `Avatar.module.css`, `useInboxQuery.ts`, `mockInboxData.ts`) produce **zero new type errors**.

---

## 9. Skipped items

Nothing skipped. All 7 tasks completed.

- Task 1: Avatar xs size — done
- Task 2: Avatar CSS audit — done (fixed `#fff` → `var(--text-on-brand)`); pre-existing px dimensions noted
- Task 3: Global leakage map — done (see section 4)
- Task 4: useInboxQuery hook — done
- Task 5: mockInboxData lib — done
- Task 6: Route stub — confirmed existing, no commit needed
- Task 7: tsc — pre-existing errors only; Phase 2 changes are type-clean
