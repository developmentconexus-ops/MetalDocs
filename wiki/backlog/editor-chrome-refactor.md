# Backlog — editor-chrome refactor

> One row = one PR. Companion to [`wiki/modules/editor-chrome-tech-debt.md`](../modules/editor-chrome-tech-debt.md). `debt_id` is `T-NNN` (register row) or `maint:<kind>` (maintenance with no debt origin).

**Last verified:** 2026-05-10

| R-id | Title | debt_id | Sketch (one line) | Effort | Risk |
|---|---|---|---|---|---|
| R-001 | Widen or align `AutosaveState` with documents autosave hook's 7-state enum | T-001 | Either extend `AutosaveState` to mirror `useDocumentAutosave`'s `AutosaveStatus` and add render branches for `dirty/stale/session_lost`, or centralize the adapter inside editor-chrome with explicit `stale`/`session_lost` visual differentiation. Today the mapping happens per-page (`DocumentEditorPage.tsx:184`) and loses signal. | M | M |
| R-002 | Add `role="status"` + `aria-live="polite"` to `AutosaveStatus` wrapper | T-002 | Wrapper `<span>` becomes `<span role="status" aria-live="polite">` (or `aria-live="assertive"` only when status='error'). Verify with screen-reader smoke (NVDA/VoiceOver). | S | L |
| R-003 | Add eigenpal selector contract guard | T-003 | Either (a) co-locate a Vitest snapshot that mounts `MetalDocsEditor` and asserts that each `:global` override selector matches at least one node in the rendered tree, or (b) ship a build-time grep against `node_modules/@eigenpal/.../dist/*.js` for the `data-testid` attribute literals. Either path catches eigenpal renames at CI time instead of in production. | M | M |
| R-004 | Establish baseline test coverage for editor-chrome primitives | T-004 | Add Vitest + RTL specs: (1) slot truthy-collapse — assert 4 overlays render only when slot truthy; (2) `AutosaveStatus` — 4 status branches render expected dot/check/label; (3) `VersionBadge` — children passthrough + class composition. Co-locate as `*.test.tsx`. | M | L |
| R-005 | Introduce missing tokens; remove hardcoded magic values | T-005 | Add `--editor-titlebar-h: 40px`, `--btn-h-sm: 26px`, `--fs-13/14/15`, `--fw-medium/semibold`, `--text-on-brand: #fff` to `styles/tokens.css`. Replace inlined `rgba(107,31,42,…)` with `color-mix(in srgb, var(--brand) 18%, transparent)` or a token alias. Drop "fully token-driven" overclaim from §5.4. | M | L |
| R-006 | Tighten `editorChromeStyles` typing or split into typed primitives | T-006 | Either narrow the re-export to a `Record<KnownClass, string>` keyed by a string-literal union, or replace re-export with `<IconBtn/> / <GhostBtn/> / <PrimaryBtn/> / <DocTitle/>` component primitives so consumers reach typed JSX instead of class strings. | M | M |
| R-007 | Lift the `pointer-events:none` discipline into the TS contract | T-007 | Either (a) split `center` slot into `centerStatic` (cosmetic, pointer-events none) vs `centerInteractive` (gets `pointer-events:auto` by default), or (b) keep `center` and document on the JSDoc that interactive children require `pointer-events:auto`, with a lint rule for `<button>` / `<a>` under `centerSlot`. | M | M |
| R-008 | Write ADR 0013 — editor-chrome extraction + slot API | T-008 | Capture (a) the `features/shared/` 2+-consumers extraction trigger applied to chrome, (b) the slot-composition shape chosen over compound components / render props, (c) the eigenpal CSS-only coupling rule (no `node_modules` patch). Cross-link from `wiki/decisions/0001-eigenpal-adoption.md` "Consequences". | S | L |
| R-009 | Document slot truthy-collapse rule on `EditorChromeProps` | T-009 | Per-slot JSDoc adds "Falsy values (`null/false/0/''/undefined`) suppress the overlay; pass any truthy `ReactNode` to render." Optionally narrow types from `ReactNode` to `ReactNode | null` to make intent visible. | S | L |

## Notes

- Effort: S ≤ ½ day, M ≤ 2 days, L > 2 days. Risk: L low, M moderate, H requires migration / external coordination.
- Order is severity-then-blast-radius (R-001/002/003/004 are the four Major items; R-005..R-009 are Minor). Pick R-002 first if you want the cheapest user-visible win.
- No `maint:` rows yet — every backlog row maps to a debt item.
