# Tech Debt Register — editor-chrome

> Companion to [`wiki/modules/editor-chrome.md`](editor-chrome.md). Lists known gaps as facts. **Debt only — no fix prescriptions.** Fixes live in [`wiki/backlog/editor-chrome-refactor.md`](../backlog/editor-chrome-refactor.md).

**Last verified:** 2026-05-10

## Severity scale

Rubric per `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Trigger list (concrete) used; no abstract "important / impactful" judgements.

## Items

### T-001 · Autosave 4-state visual cannot represent `dirty / stale / session_lost`
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx:3` (`AutosaveState` is 4-value union) ↔ `frontend/apps/web/src/features/documents/hooks/v2/useDocumentAutosave.ts:5` (`AutosaveStatus` is 7-value union); collapse at `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:184`
- **Observation:** The component's exported state enum has 4 values. The documents autosave hook's state enum has 7. The consumer page's adapter ternary maps `'dirty' | 'stale' | 'session_lost'` to `'idle'`, which renders as the green `Salvo` dot — identical to a successful save. A user whose document session was force-released by the server sees the same visual state as a user whose last edit was committed.
- **Evidence:** `_artifacts/02-flow-autosave.md §2–3`; `_artifacts/01-surface.md §2`; `_artifacts/03-deps.md §2 (name collision)`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-001`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

### T-002 · `AutosaveStatus` lacks `role="status"` / `aria-live="polite"`
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx:35,43,51,58` (wrapper `<span>` on every state branch)
- **Observation:** Wrapper `<span>` carries no `role` and no `aria-live` region. Inner dots and the check SVG carry `aria-hidden="true"`. State changes (`saving` → `saved` / `error`) are not announced to assistive technology. On a regulated-document editor where the user must trust that edits are persisted, AT users have no feedback.
- **Evidence:** `_artifacts/01-surface.md §8`; `_artifacts/02-flow-autosave.md §6`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-002`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

### T-003 · Eigenpal `:global` overrides anchored on `data-testid` + hardcoded SVG hex; no version guard
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css:160–225`
- **Observation:** 17 override selectors target eigenpal-internal attributes (`[data-testid="title-bar"]`, `[data-testid="formatting-bar"]`, `[data-testid="font-size-display"]`, `[data-testid="font-size-input"]`, `.docx-advanced-color-picker-dropdown`) and 2 hardcoded SVG fill hex values (`#cbd5e1`, `#94a3b8`). Every line carries `!important` (31 occurrences). If eigenpal renames an attribute, changes the doc-icon palette, or alters the formatting-bar geometry, overrides silently no-op. No runtime check, no CI guard, no version-pin assertion. Eigenpal is pinned to 0.2.0 (vendored `.tgz`); coupling survival depends on the pin holding.
- **Evidence:** `_artifacts/01-surface.md §5.1, §7`; `_artifacts/02-flow-eigenpal-overrides.md §6`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-003`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** [`wiki/decisions/0001-eigenpal-adoption.md`](../decisions/0001-eigenpal-adoption.md) (parent decision; selector contract not in scope)

### T-004 · Zero test coverage on a primitive used by two editor pages
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/` (no co-located `*.test.tsx` / `*.spec.tsx`); whole-repo grep returned 0 tests touching `EditorChrome / VersionBadge / AutosaveStatus`
- **Observation:** Shared primitive consumed by `TemplateEditorPage` (template authoring) and `DocumentEditorPage` (document editing). No unit / RTL / visual-regression test exercises slot rendering, autosave state branching, or eigenpal-selector survival. Any silent breakage from eigenpal refresh or design-token rename is caught only by manual smoke per `references/eigenpal-controlled-package.md` validation checklist.
- **Evidence:** `_artifacts/03-deps.md §5`; `_artifacts/01-surface.md §1`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-004`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

### T-005 · "Fully token-driven" claim drifts from code (~15 hardcoded magic-value sites)
- **Severity:** minor
- **Surface:** `EditorChrome.module.css:26,36,45,56,65,70,78,84,95,104,121,128,141,147,175,176,229,234` and `parts/VersionBadge.module.css:5,8,11` and `parts/AutosaveStatus.module.css:18,30,39`
- **Observation:** The existing wiki stub stated "fully token-driven · no hardcoded hex colors · no magic pixel values outside `--sp-*` / `--r-*`". Hex-color claim holds. Magic-pixel claim does not: 40px overlay height, 26px button height, font-sizes (15/13/12.5/12/10.5 px), font-weights (600/500/400), badge `2px 6px` padding, animation duration 1.2s, dot 8px, scrollbar 10px/2px, and 2 inlined `rgba(107, 31, 42, …)` brand-RGB values bypass the token system. No `--editor-titlebar-h`, `--fs-*`, `--btn-h-sm` tokens exist today.
- **Evidence:** `_artifacts/01-surface.md §6 ("Hardcoded values present")`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-005`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

### T-006 · `editorChromeStyles` re-export is a weakly-typed flat record
- **Severity:** minor
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx:47`
- **Observation:** `editorChromeStyles` is `typeof styles` — the inferred CSS-Module record type. Consumers reach `.iconBtn / .primaryBtn / .docTitle / ...` by string property access. A typo (`editorChromeStyles.primarybtn`) returns `undefined` at runtime; TypeScript flags it only if `noUncheckedIndexedAccess` is enabled (verify in `tsconfig`). Class-name surface is not contractually narrowed by an exported union or branded type.
- **Evidence:** `_artifacts/01-surface.md §9`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-006`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

### T-007 · `.overlayCenter pointer-events:none` discipline not in TS contract
- **Severity:** minor
- **Surface:** `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css:43,49–50` (CSS-only rule + inline comment)
- **Observation:** The centered overlay sets `pointer-events:none` so clicks pass through to eigenpal's title bar. A CSS comment instructs consumers that interactive children must opt back in with `pointer-events:auto`. The contract is invisible to TypeScript — the `center?: ReactNode` slot accepts any JSX. A future consumer adding a `<button>` to the center slot will silently lose mouse activation while keyboard activation continues to work, producing inconsistent behavior. Latent: no current consumer hits this.
- **Evidence:** `_artifacts/02-flow-mount.md §6`; `_artifacts/01-surface.md §8`.
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-007`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

### T-008 · Missing ADR for chrome-extraction rule and slot API
- **Severity:** minor
- **Surface:** [`wiki/decisions/`](../decisions/) (no `0xxx-editor-chrome*.md`); rule embedded in `wiki/architecture/frontend-structure.md` ("used by 2+ features ⇒ shared")
- **Observation:** Two load-bearing decisions are not captured as standalone ADRs: (a) extracting editor-chrome as a `features/shared/` primitive on the 2+-consumers rule; (b) the slot-based composition shape (`left/center/right/alert/children`) vs other valid shapes (compound components, render props, sub-components). Both decisions are visible only in code + the architecture overview doc.
- **Evidence:** `_artifacts/00-context.md §"ADR / concept anchors"`; §9 of [`editor-chrome.md`](editor-chrome.md).
- **Linked backlog row:** [`backlog/editor-chrome-refactor.md#R-008`](../backlog/editor-chrome-refactor.md)
- **Linked ADR:** missing-ADR

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
- Decisions without ADR link: 8 (T-001, T-002, T-004, T-005, T-006, T-007, T-008, T-009 → all `missing-ADR`); T-003 links ADR 0001 as parent
