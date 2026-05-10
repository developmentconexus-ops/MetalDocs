# Screen Review: distribuicao

**Implementation:** `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx`
**Components:** `frontend/apps/web/src/features/documents/components/distribution/`
**Hero (shared):** `frontend/apps/web/src/features/documents/components/DocumentHero.tsx`
**Layout:** `frontend/apps/web/src/features/documents/pages/DocumentDetailLayout.tsx`
**Design source:** `frontend/apps/web/design-source/distribuicao/`
**Visual comparison:** Numerical parity only — `preview_screenshot` timed out on both servers (design HTML likely blocked on SVG animation loop); all evidence below is computed-style from `preview_eval`. App server: 4174; Design server: 4181.
**Verdict:** APPROVE WITH NITS

---

## Critical

None.

---

## Major

### Visual / numerical parity

- [ ] `KPIStrip.module.css:13 · cell horizontal padding` — design `pl/pr = 22px` (measured: `style="padding: 18px 22px"`), impl `pl/pr = 20px` (`padding: 18px var(--sp-5)`), delta `+2px`. *Why:* `tokens.css:74` defines `--sp-5: 20px`; the design uses `22px` which has no mapped token. *Evidence:* design eval `pt:18px pl:22px`; impl eval `pt:18px pl:20px`. *Fix:* Introduce `--sp-5h: 22px` or accept `--sp-5` and add a note in parity-diff — do NOT silently call it a match.

- [ ] `RecipientsCard.module.css:40 · headerRow horizontal padding` — design `padding: 10px 22px` (measured), impl `10px var(--sp-5)` = `10px 20px`, delta `+2px`. *Why:* same token gap as KPI cell. *Evidence:* design eval `pl:22px pt:10px`; impl eval `pl:20px pt:10px`. *Fix:* same token fix as KPI cell, or document as accepted delta.

- [ ] `RecipientsCard.module.css:76 · recipientRow horizontal padding` — design `padding: 12px 22px` (measured `style="...padding: 12px 22px..."`), impl `var(--sp-3) var(--sp-5)` = `12px 20px`, delta `+2px`. *Why:* same token gap. *Evidence:* design eval `pt:12px pl:22px`; impl eval `pt:12px pl:20px`. *Fix:* same as above.

- [ ] `CoverageByArea.module.css:62 · rows horizontal padding` — design `padding: 20px 22px` (measured `style="padding: 20px 22px"`), impl `padding: var(--sp-5)` = `20px 20px`, delta `+2px`. *Why:* same token gap. *Evidence:* design eval `pl:22px pt:20px`; impl eval `pl:20px pt:20px`. *Fix:* same.

**Summary of the 22px pattern:** The design consistently uses `22px` for horizontal padding inside card-level regions (KPI cells, RecipientsCard header/data rows, CoverageByArea rows). The token set only has `--sp-5: 20px` and `--sp-6: 24px`. The implementer rounded to `--sp-5`. The drift is uniform at 2px and visually subtle, but the `parity-diff.md` artifact claims "Matches design (sp-5=20px used for horizontal)" for the KPI cell — this is incorrect and constitutes a false claim in the Phase 3b artifact (see Iron-Law cross-check below).

### Iron-Law artifacts

- [ ] `artifacts/parity-diff.md · KPIStrip cell padding` — artifact claims `18px 20px` "Matches design exactly". Actual design measurement (via computed style + inline style inspection) is `18px 22px`. The artifact states a match where there is a 2px horizontal delta. *Why:* Iron-Law §3b requires parity-diff numbers to match reality. *Fix:* Correct the parity-diff entry to show design:`18px 22px`, impl:`18px 20px`, delta:`pl/pr +2px`; add a note that `--sp-5=20px` was used as the nearest token.

### A11y

- [ ] `RecipientsCard.tsx:122-127 · icon-only row action buttons` — 4 buttons (mail + more icons, 2 per visible row) have `aria-disabled="true"` + `title="Em breve"` but no `aria-label`. Screenreaders announce an unnamed button with role "button" and no accessible name. *Why:* `wiki/architecture/frontend-structure.md §5.E` — icon-only buttons must have `aria-label`. *Fix:* Add `aria-label="Enviar lembrete"` / `aria-label="Mais opções"` to each `actionBtn`.

- [ ] `RecipientsCard.tsx:139-146 · pagination prev/next buttons` — both `._paginationBtn_*` are icon-only with no `aria-label`, no `title`, no visible text. Screenreader has no accessible name. *Why:* `wiki/architecture/frontend-structure.md §5.E` — icon-only interactive elements require accessible name. *Fix:* Add `aria-label="Página anterior"` and `aria-label="Próxima página"` to the two buttons.

### Tokens / primitive drift

- [ ] `CoverageByArea.module.css:116-125 · barRead + barAck transitions` — `transition: width 1.4s cubic-bezier(...)` and `transition: width 1.4s cubic-bezier(...) 0.1s` have no `@media (prefers-reduced-motion: reduce)` fallback. *Why:* `wiki/architecture/frontend-structure.md §10` + SKILL.md Phase 5E — every animated element must have a reduced-motion block. *Fix:*
  ```css
  @media (prefers-reduced-motion: reduce) {
    .barRead, .barAck { transition: none; }
  }
  ```

- [ ] `KPIStrip.module.css:50 · progressFill transition` — `transition: width 0.3s ease-out` has no `@media (prefers-reduced-motion: reduce)` fallback. *Why:* same rule. *Fix:*
  ```css
  @media (prefers-reduced-motion: reduce) {
    .progressFill { transition: none; }
  }
  ```

### Wiki hygiene

- [ ] `wiki/modules/documents.md` does not reference `DocumentDistributionPage`, `DocumentDetailLayout`, or the `distribuicao` route. The `Last verified` stamp is `2026-05-08` but the screen was fully implemented on the same date and the wiki was not updated. *Why:* `CLAUDE.md` — after implementations, update `wiki/modules/<feature>.md` anchor + `Last verified`. *Fix:* Dispatch `wiki-curator` to add `DocumentDistributionPage` + `DocumentDetailLayout` anchors, update route table, bump stamp.

---

## Minor

- `DocumentDistributionPage.tsx:99 · sectionAside anchor` — `<a href="#" className={styles.sectionAside}>Lembrar todas as áreas críticas →</a>` uses `href="#"` for a deferred action but is not `aria-disabled`. Consistent with other deferred CTAs which use `<button aria-disabled="true">`. The `<a href="#">` is mildly misleading. *Fix:* Either `<button type="button" aria-disabled="true">` or `<span role="link" aria-disabled="true">` for consistency with the page's other "Em breve" pattern.

- `DonutCard.tsx:6 · module-level computation` — `const C = 2 * Math.PI * 42` and derived arc values run at module parse time (not in component body). No functional issue since the data is mock, but when the component is wired to real data via props, the derivations will not react to prop changes. Tag with a comment or move into component body as a reminder. (This is a code-reviewer concern; flagged as Minor here for awareness.)

- `IMPLEMENTATION.md:63-100 · Phases 1–5 stubs` — Phases 1–5 in the worksheet all read "_To be filled after Phase X_" even though those phases are complete. The artifact files (`phase1-map.md`, `phase2-preflight.md`, `phase3a-structure.md`, `phase3b-style.md`, `phase4-behavior.md`) exist and are filled in. The worksheet itself was never updated to reference them. Not a blocking issue but creates confusion for future readers. *Fix:* Update IMPLEMENTATION.md to link to or summarize each completed phase artifact.

---

## What's good

- Routing architecture is solid: `documents/:documentId` with `DocumentDetailLayout` as parent and child routes (`index: DocumentPublishedPage`, `distribution: DocumentDistributionPage`) using `lazy()` throughout — correct React Router v6 nested-route pattern with lazy loading.
- Checkbox leakage completely handled: `RecipientsCard.module.css:65` resets `width: auto; border: none; background: transparent; padding: 0` on `.checkbox`, and the global `base.css` `input:not([type="checkbox"])` selector correctly excludes checkboxes. Measured computed width `13px` (natural). The known `SearchBar` input leakage is pre-existing and correctly documented.
- All deferred CTAs use `aria-disabled="true"` + `title="Em breve"` + `pointer-events: none`. No misleading enabled state. All 5 deferred CTAs (4 hero + 1 "Editar política") verified.
- Heading outline is clean: `H1` → `H2` × 4. No skipped levels.
- Mock data TODO comments in `distributionMeta.ts` (lines 68-69, 112-113) are correctly keyed to `wiki/backlog/distribuicao.md` with specific endpoint references. Backlog file exists and is comprehensive.

---

## Iron-Law cross-check

- Phase 0 audit signed: ✅ (confirmed by user 2026-05-08, full Keep/Cut/Defer table in `artifacts/phase0-audit.md`)
- Phase 1 worksheet complete: ✅ (`artifacts/phase1-map.md` — decomposition, state design, backend contract, new token list)
- Phase 2 primitive audit verified: ✅ (`artifacts/phase2-preflight.md` exists)
- Phase 3a DOM diff approved: ✅ (`artifacts/phase3a-structure.md` exists)
- Phase 3b: parity-diff covers all regions ✅; leakage-probe covers form elements ✅; token-coverage implicitly empty (no bypasses found in code) ✅; **numbers match reality ❌** — `parity-diff.md` KPIStrip entry claims `18px 20px` "Matches design exactly" but actual design value is `18px 22px` (2px horizontal delta, verified by computed style probe)
- Phase 4 behavior trace present: ✅ (`artifacts/phase4-behavior.md` — tsc, test, manual smoke all documented)
- Open Questions Log resolved: ✅ (all 3 OQs resolved in Phase 0)
