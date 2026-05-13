# Phase 2 Pre-flight — documento-publicado

Completed: 2026-05-08

---

## Summary

| Task | Status |
|---|---|
| Codegen | No change needed — api-types already current |
| Primitives audited | StatusPill, CodeChip, Avatar, Icon |
| Primitive changes | eye icon added to Icon.tsx |
| Icon.tsx: eye icon added | ✅ |
| Status-meta file | `src/features/documents/lib/documentDetailMeta.ts` ✅ |
| New API types | SignoffRecord, StageInstance, ApprovalInstanceResponse in `documentsV2.ts` ✅ |
| New API function | `getApprovalInstance` in `documentsV2.ts` ✅ |
| New query hooks | `useDocumentDetailQuery`, `useApprovalInstanceQuery` ✅ |
| Route stub | `documents/:documentId` → `DocumentPublishedPage` (stub) ✅ |
| Global leakage map | `artifacts/leakage-map.md` ✅ |
| Backlog file | `wiki/backlog/documento-publicado.md` ✅ |
| tsc | Pre-existing errors only — zero new errors ✅ |

---

## Step 1 — Codegen

`pnpm gen:api` ran and produced no diff to `src/lib/api-types/index.d.ts`. The OpenAPI
spec has not changed since the last codegen run. No commit needed.

---

## Step 2 — Primitive audit

### StatusPill.tsx + StatusPill.module.css

All CSS properties use `var(--token)` — clean. No raw hex or raw px violations
(small spacing values like `5px gap`, `6px dot`, `3px 8px padding` are structural
sizing, not palette/color values — acceptable). **No changes needed.**

### CodeChip.tsx (no CSS module)

Uses global classes `code-chip` and `mono`. The `code-chip` class is defined in
`frontend/apps/web/design-source/styles.css` (design reference, not src). The `.mono`
global class at styles.css:1097 uses raw font-family string (`"DM Mono", Consolas, monospace`)
not `var(--font-mono)` — pre-existing drift in a global utility class. Out of scope.
The component itself has no CSS to audit. **No changes needed.**

### Avatar.tsx + Avatar.module.css

All color properties use `var(--token)`. The `color` property was already fixed to
`var(--text-on-brand)` in the caixa-aprovacao Phase 2 run. Pre-existing pixel
dimensions (`28px`, `22px`, `36px`) have no matching `--sp-*` token — documented
as pre-existing, not changed. **No changes needed.**

### Icon.tsx

No CSS file. All styling via SVG props (width, height, viewBox). **No token violations.**

**Eye icon added:** Added `eye` to `IconName` union type and `PATHS` record.

---

## Step 2b — Global CSS leakage map

Full map written to `artifacts/leakage-map.md`.

Key findings for this screen:
- **No HIGH-risk leakage** — documento-publicado is a read-only display screen with no bare inputs or tables.
- `button:disabled { opacity: 0.5 }` from base.css is overridden by the more-specific `.btn:disabled { opacity: 0.45 }` — no issue.
- `input:not(...)` width/border rules irrelevant (no form inputs in this screen).
- `table/tr/td` rules irrelevant (SignoffPipeline uses divs, not tables).

---

## Step 3 — Status-meta SSOT

Created `src/features/documents/lib/documentDetailMeta.ts` with:
- `resolveProfileLabel(code, profiles[])` — lookup util
- `resolveAreaLabel(code, areas[])` — lookup util
- `SignoffStatus` type
- `SIGNOFF_STATUS_META` record (4 statuses: pending, approved, rejected, abstained)

---

## Step 4 — New API types and function

Added to `src/features/documents/api/documentsV2.ts`:
- `SignoffRecord` type
- `StageInstance` type
- `ApprovalInstanceResponse` type
- `getApprovalInstance(documentId: string)` function → `GET /api/v1/documents/:id/approval-instance`

---

## Step 5 — New query hooks

Created `src/features/documents/queries/useDocumentDetailQuery.ts`:
- Wraps `getDocument(id)` with `QK.documents.detail(id)` key
- `enabled: Boolean(id)`

Created `src/features/documents/queries/useApprovalInstanceQuery.ts`:
- Wraps `getApprovalInstance(documentId)` with `QK.approval.instance(documentId)` key
- 404 retry short-circuit: returns false on 404 (no approval instance = null state, not error)
- `enabled: Boolean(documentId)`

---

## Step 6 — Route stub

Added to `src/features/documents/routes.tsx` (before `documents-v2/new` route):
```ts
{
  path: 'documents/:documentId',
  handle: { workspaceView: 'library' },
  lazy: () => import('./pages/DocumentPublishedPage').then(m => ({ Component: m.DocumentPublishedPage })),
},
```

Created stub page `src/features/documents/pages/DocumentPublishedPage.tsx`:
- Returns `<div style={{ padding: '2rem', color: 'var(--text-muted)' }}>Documento publicado — em construção</div>`
- No page content — stub only, per hard rules.

---

## Step 7 — Backlog file

Created `wiki/backlog/documento-publicado.md` with deferred items:
- AuditCard / ISO seal (`values_hash` not in API)
- CommentsCard (architecture brainstorm needed)
- PDF download (no endpoint)
- Coverage card (no fanout API)
- VersionTimeline (no revision list endpoint)
- RelatedGrid (no relationship model)

---

## tsc result

```
Exit code 2 — pre-existing errors only
```

Errors are identical to those documented in the caixa-aprovacao Phase 2 preflight:
- `src/features/auth/__tests__/useAuthSession.returnTo.test.tsx` (×2)
- `src/features/documents/components/LibrarySidebar.tsx` (×2)
- `src/features/documents/pages/NewDocumentWizardPage.tsx` (×3)
- `src/features/documents/queries/__tests__/useAreasQuery.test.ts` (×3)
- `src/features/documents/queries/useAreasQuery.ts` (×1)
- `src/features/shell/components/Rail.tsx` (×1)

None of these files were touched in Phase 2. Zero new type errors introduced.

---

## Skipped

Nothing skipped. All 7 steps completed.
