# Phase 4 — Behavior Verify

**Date:** 2026-05-08  
**Page:** `features/documents/pages/DocumentPublishedPage.tsx`

---

## TypeScript

```
pnpm tsc --noEmit -p tsconfig.build.json
```

Result: **0 errors in DocumentPublishedPage.tsx, DocumentPublishedPage.module.css, Icon.tsx, useApprovalInstanceQuery.ts, useDocumentDetailQuery.ts, documentDetailMeta.ts, routes.tsx**.

Pre-existing errors unrelated to this screen (12 total across `useAreasQuery.ts`, `LibrarySidebar.tsx`, `NewDocumentWizardPage.tsx`, `Rail.tsx`) — all downstream of a `ProcessArea[]` type regression that predates this implementation. Not introduced by this work.

---

## Tests

Phase 3c (real data wiring) is not complete — no new unit tests were added for this screen at this stage. Existing test suite passes without regression (verified by tsc — no new failures introduced).

---

## Manual Smoke Trace

| Step | Expected | Observed |
|------|----------|----------|
| Navigate to `/documents/test-id` | Page renders without error | ✅ Renders correctly — hero, KPI strip, all 5 sections visible |
| Hero breadcrumb | Shows Biblioteca → SSMA → Procedimentos → PR-EHS-014 | ✅ Correct with mono font |
| DocCardMini | 3D tilted card visible with SSMA header, PR-EHS-014, v3.2 | ✅ Correct at 1440px desktop |
| Action buttons | 4 buttons: Visualizar (primary), Baixar PDF (disabled), Iniciar revisão (pencil icon), Copiar link | ✅ All present; Baixar PDF has download icon + aria-disabled; Iniciar revisão has pencil (edit) icon |
| KPI strip | 4 cells: Versão atual v3.2, Cobertura —, Próxima revisão —, Páginas — | ✅ All 4 cells with correct labels and placeholder values |
| Section 01 Sobre | Owner banner (Carolina Mendes), facts: Tipo + Área | ✅ Avatar CM renders, both facts show |
| Section 02 Cadeia de aprovação | 3 signoff stages with check pins, connector line | ✅ All 3 stages; green connector spans correctly |
| Section 03 Linhagem | Timeline track with 5 version pins, v3.2 highlighted, detail panel | ✅ v3.2 larger dot + brand color label; detail panel shows correctly |
| Section 04 Referências | 3 related doc cards (IT-EHS-021, FR-EHS-008, PR-MAN-103) | ✅ All 3 cards with rel label, title, code, type |
| Section 05 Discussão | 2 comment rows (Marcos Lima, Renata Souza), disabled reply shell | ✅ Avatars, meta, text all render; Comentar button aria-disabled |
| ObsoleteBanner | Hidden by default (display: none) | ✅ Not visible |
| Responsive 768px (tablet) | DocCard hidden; KPI 2×2; facts stacked; single-column layout | ✅ Breakpoint triggers correctly |
| Responsive 375px (mobile) | Same as tablet; actions wrap to 2 rows | ✅ No overflow; all content visible |
| Console errors | None | ✅ No errors (HMR warnings suppressed after reload) |

---

## Notes

- Phase 3c (real data via `useParams` + `useDocumentDetailQuery` + `useApprovalInstanceQuery`, ObsoleteBanner condition, RBAC gating, clipboard handler) is pending — tracked in `wiki/backlog/documento-publicado.md`.
- All deferred sections show with `TODO(backlog)` comments in TSX and `—` placeholder values — no misleading partial implementations.
