# Tech Debt Register — editor-chrome

> Companion to [`wiki/modules/editor-chrome.md`](editor-chrome.md). Lists known gaps as facts. **Debt only — no fix prescriptions.** Fixes live in [`wiki/backlog/editor-chrome-refactor.md`](../backlog/editor-chrome-refactor.md).

**Last verified:** 2026-07-02 (FE-06/FE-07 design-system hygiene wave)

## Severity scale

Rubric per `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Trigger list (concrete) used; no abstract "important / impactful" judgements.

## Items

### T-001 · Autosave 4-state visual cannot represent `dirty / stale / session_lost` — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx:3` (`AutosaveState` widened to 7-value union); `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:185` (ternary-collapse removed — direct passthrough).
- **Resolution (2026-05-11, commit `c2a43abd`):** `AutosaveState` union extended to `'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error'`. `DocumentEditorPage` ternary that mapped `dirty/stale/session_lost` → `idle` removed; `autosaveState` is now a direct assignment from `autosave.status`. All 7 states render with distinct pt-BR labels and icons in `AutosaveStatus`. RTL tests covering all 7 states added in commit `29994d7a` (`EditorChrome.test.tsx`).
- **Evidence:** `_artifacts/02-flow-autosave.md §2–3`; `_artifacts/01-surface.md §2`; `_artifacts/03-deps.md §2 (name collision)`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-001`](../backlog/editor-chrome-refactor.md) (closed)
- **Linked ADR:** missing-ADR

### T-002 · `AutosaveStatus` lacks `role="status"` / `aria-live="polite"` — **RESOLVED (verified 2026-07-02)**
- **Severity:** major → **resolved**
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx:40` (single wrapper `<span>`, all 7 states)
- **Resolution:** The wrapper `<span>` carries `role="status"` and `aria-live` — `"assertive"` for `error`/`session_lost`, `"polite"` for the other 5 states (`AutosaveStatus.tsx:30,37,40`). Already in place when this FE-06/FE-07 pass started; landed alongside the T-001 7-state widening (commit `c2a43abd` / `29994d7a`) rather than as a separate PR. `EditorChrome.test.tsx` asserts both the role and the correct `aria-live` value for every one of the 7 states (`AutosaveStatus` describe block, 14 assertions).
- **Evidence:** `AutosaveStatus.tsx:28-45`; `EditorChrome.test.tsx` `AutosaveStatus` describe block (all 7 `CASES` × role + aria-live checks); `npx vitest run src/features/shared/components/editor-chrome` → 29/29 passed (2026-07-02).
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-002`](../backlog/editor-chrome-refactor.md) (closed)
- **Linked ADR:** missing-ADR

### T-003 · Eigenpal `:global` overrides anchored on `data-testid` + hardcoded SVG hex; no version guard — **PARTIALLY RESOLVED 2026-07-02**
- **Severity:** major → **CI guard added; selector-DOM-match verification remains manual**
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css:169–254` (post-fix line range; was `160–225`)
- **Observation (original):** 17 override selectors target eigenpal-internal attributes... Eigenpal is pinned to 0.2.0 (vendored `.tgz`); coupling survival depends on the pin holding.
- **Correction found during resolution:** the "pinned to 0.2.0 vendored `.tgz`" premise was itself stale drift. The actual resolved dependency is `@eigenpal/docx-editor-core@1.9.0` + `@eigenpal/docx-editor-react@1.9.0` (npm-scoped packages, see `packages/eigenpal-adapter/package.json` and `packages/editor-ui/package.json`) — the `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` vendored tarball is not what's actually wired into the build. Actual selector count at verification time was 22 `:global(` occurrences (not 17) and 28 `!important` (not 31) — both prior counts were stale, not current facts.
- **Resolution:** Added `scripts/check-eigenpal-selector-pin.sh`, modeled on `scripts/check-css-token-discipline.sh`. It pins (a) the resolved `@eigenpal/docx-editor-core` version read from `packages/eigenpal-adapter/package.json`, (b) the `:global(` occurrence count in `EditorChrome.module.css` (22), and (c) the `!important` occurrence count (28). Any eigenpal version bump or edit to the override block fails the script until a human updates the pin and re-runs the manual smoke checklist in `wiki/references/eigenpal-controlled-package.md`. The coupling-pin comment in `EditorChrome.module.css:157-172` and the `EditorChrome.tsx` JSDoc were corrected to reference the real npm version instead of the stale tarball path.
- **Still open (not closed by this pass):** the guard is a count-drift tripwire, not a DOM-match assertion — it cannot detect an eigenpal release that keeps the same selector/`!important` counts but silently renames a `data-testid` or changes SVG fill hex. R-003's original sketch (b) — a build-time grep against `node_modules/@eigenpal/.../dist/*.js` for the literal `data-testid` strings — remains a legitimate follow-up if tighter guarantees are wanted.
- **Evidence:** `scripts/check-eigenpal-selector-pin.sh` (new); `bash scripts/check-eigenpal-selector-pin.sh` → `check-eigenpal-selector-pin: clean (eigenpal@1.9.0, 22 :global( selectors, 28 !important)` (2026-07-02).
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-003`](../backlog/editor-chrome-refactor.md) (partially closed — DOM-match sub-task remains open)
- **Linked ADR:** [`wiki/decisions/0001-eigenpal-adoption.md`](../decisions/0001-eigenpal-adoption.md) (parent decision; selector contract not in scope)

### T-004 · Zero test coverage on a primitive used by two editor pages — **RESOLVED (verified 2026-07-02, stale finding)**
- **Severity:** major → **resolved**
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.test.tsx` (29 tests)
- **Resolution:** This finding was already stale before the FE-06/FE-07 pass started — `EditorChrome.test.tsx` exists and covers slot truthy-collapse (9 tests), all 7 `AutosaveStatus` states × label + `aria-live` correctness (14 tests + 2 label/className tests + 1 icon-reuse test), and `VersionBadge` passthrough (2 tests). `npx vitest run src/features/shared/components/editor-chrome` → 29/29 passed (2026-07-02). No action taken this pass beyond verifying the claim against current code.
- **Evidence:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.test.tsx` (136 lines, full file); vitest run output (2026-07-02).
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-004`](../backlog/editor-chrome-refactor.md) (closed)
- **Linked ADR:** missing-ADR

### T-005 · "Fully token-driven" claim drifts from code (~15 hardcoded magic-value sites) — **RESOLVED 2026-07-02**
- **Severity:** minor → **resolved**
- **Surface:** `EditorChrome.module.css`, `parts/VersionBadge.module.css`, `parts/AutosaveStatus.module.css`, plus the related novo-documento wizard sites (`StepConfirm.module.css`, `StepProfile.module.css`) folded into this pass.
- **Resolution:** Densified the existing `--font-size-*` scale in `src/styles/tokens.css` (was 7 steps with gaps at 10.5/12/13/15/16px; now 11 steps: `2xs, 2xs-2, xs, xs-2, sm, sm-2, md, md-2, md-3, lg, xl, 2xl`) instead of introducing a parallel `--fs-*` naming scheme — extending the established, 60-file-adopted convention is the global-maximum choice over forking a second scale. Added `--sp-0-5: 2px` and `--sp-1-5: 6px` to densify the spacing scale for the badge padding and alert-banner padding sites. Added component-scoped custom properties (`--autosave-status-min-w`, `--autosave-dot-size`, `--autosave-pulse-duration`) on `AutosaveStatus.module.css`'s `.status` selector for the 3 single-use magic values (60px min-width, 8px dot, 1.2s pulse) that don't belong in the global token surface. Added `--ctl-h-sm: 26px` / `--ctl-h-md: 28px` control-height tokens for the chrome button primitives. All hardcoded eigenpal-section hex/`!important` values (`EditorChrome.module.css:169+`) were intentionally left as-is — those are DOM-coupling values governed by T-003, not design-token debt.
- **Evidence:** `git diff` on the 5 files above; `bash scripts/check-css-token-discipline.sh` → clean (2026-07-02); `npx tsc --noEmit` clean; `npx vitest run` 48/48 passed on touched dirs.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-005`](../backlog/editor-chrome-refactor.md) (closed)
- **Linked ADR:** missing-ADR

### T-006 · `editorChromeStyles` re-export is a weakly-typed flat record — **RESOLVED 2026-07-02**
- **Severity:** minor → **resolved**
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx:77-101`
- **Resolution:** Added an explicit `EditorChromeClass` string-literal union naming every class `EditorChrome.module.css` actually defines (`wrapper`, `overlayLeft/Center/Right/Alert`, `docTitle/Sep/Meta`, `iconBtn`, `ghostBtn`, `primaryBtn`). `editorChromeStyles` is now typed as `Record<EditorChromeClass, string>` (cast from the vite/client ambient CSS-module index-signature type — a type-only narrowing, not a new runtime object) so `editorChromeStyles.primarybtn` (typo) is a compile error instead of silently resolving to `undefined`. `noUncheckedIndexedAccess` remains off project-wide (unchanged — out of this pass's scope, would need a dedicated tsconfig-wide sweep).
- **Evidence:** `EditorChrome.tsx:77-101`; `npx tsc --noEmit` clean (2026-07-02).
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-006`](../backlog/editor-chrome-refactor.md) (closed)
- **Linked ADR:** missing-ADR

### T-007 · `.overlayCenter pointer-events:none` discipline not in TS contract — **RESOLVED (documented) 2026-07-02**
- **Severity:** minor → **resolved via documentation (option b from R-007)**
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx:7-17` (`center` prop JSDoc)
- **Resolution:** Chose R-007 option (b) over (a) — splitting `center` into `centerStatic`/`centerInteractive` would be a breaking API change across both consumers (`DocumentEditorPage.tsx`, `TemplateEditorPage.tsx`) for a contract that has zero current violations (per the original finding: "Latent: no current consumer hits this"). Moved the `pointer-events:none` contract from a class-level JSDoc paragraph (easy to miss) to a dedicated JSDoc block directly on the `center` prop declaration, explicit that interactive children (button/link/dropdown trigger) must self-opt-in with `pointer-events:auto`. No lint rule added (would need a custom ESLint rule scoped to JSX under this one prop — assessed as disproportionate to a latent, zero-instance risk).
- **Evidence:** `EditorChrome.tsx:7-17`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-007`](../backlog/editor-chrome-refactor.md) (closed — documentation path, not the TS-split path)
- **Linked ADR:** missing-ADR

### T-008 · Missing ADR for chrome-extraction rule and slot API — CLOSED 2026-07-02 (ADR 0063)
- **Severity:** minor (closed)
- **Surface:** [`wiki/decisions/`](../decisions/) (no `0xxx-editor-chrome*.md`); rule embedded in `wiki/architecture/frontend-structure.md` ("used by 2+ features ⇒ shared")
- **Observation (original):** Two load-bearing decisions are not captured as standalone ADRs: (a) extracting editor-chrome as a `features/shared/` primitive on the 2+-consumers rule; (b) the slot-based composition shape (`left/center/right/alert/children`) vs other valid shapes (compound components, render props, sub-components). Both decisions are visible only in code + the architecture overview doc.
- **Resolution:** ADR 0063 records both: (a) promote-on-second-caller placement, verified against the two actual consumers (`DocumentEditorPage.tsx`, `TemplateEditorPage.tsx`, per `EditorChrome.tsx:55`'s own doc comment) and the general rule at `frontend-structure.md:150-153`; (b) named slots chosen over compound components (unneeded context machinery for 4 fixed regions) and render props (no internal state slots need to react to) — the fixed, non-recursive, eigenpal-title-bar-overlay shape is the deciding factor.
- **Evidence:** `_artifacts/00-context.md §"ADR / concept anchors"`; §9 of [`editor-chrome.md`](editor-chrome.md); `wiki/decisions/0063-editor-chrome-extraction-and-slot-api.md`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-008`](../backlog/editor-chrome-refactor.md) (can be closed)
- **Linked ADR:** `wiki/decisions/0063-editor-chrome-extraction-and-slot-api.md`

### T-009 · Slot truthy-collapse semantics invisible in TS
- **Severity:** minor
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx:34–37`
- **Observation:** Each slot uses `{slot && <div>...</div>}`. A consumer passing `0`, `''`, `false`, `null`, or `undefined` all collapse identically — no slot div rendered. Intent is "render only when slot has content", but the TS type `ReactNode` already allows `false/null/undefined/string`. Empty-string and `0` (legal `ReactNode`) suppress the overlay rather than render an empty box; not documented in `EditorChromeProps` JSDoc.
- **Evidence:** `_artifacts/02-flow-mount.md §6`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-009`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 0 / 6 (every export named in §5.2 of the module doc)
- Operations missing C4 placement: 0 / 0 (no HTTP ops)
- Cross-deps missing in §5/§8: 0 / 7 OUT + 2 IN (per `_artifacts/03-deps.md`)
- State transitions missing in §6: 0 / 1 (autosave state machine is in §6.2)
- Decisions without ADR link: 2 open (T-009, and T-003's DOM-match sub-task); T-002/T-004/T-005/T-006/T-007 resolved 2026-07-02 (FE-06/FE-07); T-008 closed via ADR 0063; T-003 links ADR 0001 as parent (CI guard added, DOM-match verification sub-task remains); T-001 resolved
