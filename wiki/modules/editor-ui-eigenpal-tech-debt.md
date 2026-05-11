# Tech Debt Register — editor-ui-eigenpal

> Companion to `wiki/modules/editor-ui-eigenpal.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/editor-ui-eigenpal-refactor.md`.

**Last verified:** 2026-05-11

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Triggers used here:

- **Critical** — schema/version drift the boot check should catch but does not; or data-loss / supply-chain path where fresh installs break.
- **Major** — defense-in-depth gap; documented contract not followed; duplicated write surfaces; cross-module dependency that blocks another module's refactor; false-pass test risk on a load-bearing branch.
- **Minor** — latent symbol, doc/code drift on a non-load-bearing path, missing standalone ADR for a rule already enforced by code + tests.

## Items

### T-001 · Vendored eigenpal tarball absent from `main` — **RESOLVED Plan 3**
- **Severity:** critical → **resolved**
- **Surface:** `packages/editor-ui/package.json:29`, `apps/docgen-v2/package.json:15`, `frontend/apps/web/package.json:17` — each references `file:../../[…]/vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- **Resolution (2026-05-11):** Tarball restored at `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` (blob `0e35c089`) and `vendor/eigenpal/README.md` (blob `4ec632f0`) from git history. Fresh `pnpm install` resolves the dep. ADR 0001 pin is intact. R-009 (wiki refresh for ADR 0001 + eigenpal-controlled-package) remains open as a separate docs PR.
- **Evidence:** `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` present; `vendor/eigenpal/README.md` present.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-001` (closed)
- **Linked ADR:** `wiki/decisions/0001-eigenpal-adoption.md`

### T-002 · TemplateEditorPage bypasses the `MetalDocsEditor` wrapper — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` — previously imported `DocxEditor` directly from `@eigenpal/docx-js-editor/react`.
- **Resolution (2026-05-11, commit `60fa5473`):** `TemplateEditorPage` migrated to `MetalDocsEditor`. Direct `@eigenpal/docx-js-editor` imports removed; `useRef<MetalDocsEditorRef>` now used. Repo-wide grep `@eigenpal/docx-js-editor` in `frontend/apps/web/src` returns zero outside type-only positions. Anti-Corruption Layer now holds for both consumer pages.
- **Evidence:** `_artifacts/03-deps.md` IN-edges table (updated).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-002` (closed)
- **Linked ADR:** missing-ADR (see T-008 — the rule still lacks its own ADR)

### T-003 · `templatePlugin.wiring.test.tsx` asserts pre-gating contract — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `packages/editor-ui/test/templatePlugin.wiring.test.tsx`
- **Resolution (2026-05-11, commit `ce6d809a`):** Test rewritten to 5 correct assertions gated on `template-draft` mode, aligned with the 2026-05-06 plugin-gating refactor. `document-edit` path asserts `data-plugins='0'` for `templatePlugin`. No stale contract survives.
- **Evidence:** `_artifacts/02-flow-plugin-registration.md` "Stale wiring spec" subsection (now stale-free).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-003` (closed)
- **Linked ADR:** missing-ADR (see T-007 for the gating rule itself)

### T-004 · `createOutlinePlugin` exported but not registered
- **Severity:** minor
- **Surface:** `packages/editor-ui/src/index.ts:6` exports `createOutlinePlugin`; `packages/editor-ui/src/MetalDocsEditor.tsx:55-59` plugin list does not include it. Source file `plugins/OutlinePlugin.tsx` is still maintained (147 LOC) but dormant.
- **Observation:** Public surface advertises a plugin that the wrapper does not register. Removed from the array in the 2026-05-06 refactor; eigenpal's own `docx-outline-nav` overlay still ships, so user-facing behavior is unchanged. The export persists with no current consumer (`grep createOutlinePlugin frontend/` is empty for non-test code).
- **Evidence:** `_artifacts/01-surface.md` public-exports table; `_artifacts/03-deps.md` IN-edges (no production import of `createOutlinePlugin`).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-004`
- **Linked ADR:** none required

### T-005 · `mergefieldPlugin.ts` filename misnames its contents
- **Severity:** minor
- **Surface:** `packages/editor-ui/src/plugins/mergefieldPlugin.ts:1-30`
- **Observation:** The file does not export an `EditorPlugin`. It exports a data-only function `computeSidebarModel` and a `SidebarModel` type. The plugin that consumes that data lives in `sidebarModelBridge.ts`. The `Plugin` suffix in the filename is misleading; the existing wiki carries a status `VERIFY` flag on it.
- **Evidence:** `_artifacts/01-surface.md` public exports; `mergefieldPlugin.ts` source (no `EditorPlugin` return type anywhere).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-005`
- **Linked ADR:** none required

### T-006 · `onLockLost` prop declared but never wired
- **Severity:** minor
- **Surface:** `packages/editor-ui/src/types.ts:25` declares `onLockLost?: () => void`; `packages/editor-ui/src/MetalDocsEditor.tsx` never reads or forwards it.
- **Observation:** Type surface advertises a lock-loss callback. The wrapper never invokes it, and `DocxEditor` is not configured with any lock-listener. Consumers that pass `onLockLost` get no callback ever. Latent — no current MetalDocs page passes the prop.
- **Evidence:** `_artifacts/01-surface.md` props surface; source grep `onLockLost` in `MetalDocsEditor.tsx` returns nothing.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-006`
- **Linked ADR:** none required

### T-007 · No ADR for `templatePlugin` mode-gating rule
- **Severity:** minor
- **Surface:** `packages/editor-ui/src/MetalDocsEditor.tsx:55-56`; rationale lives only in source comments and `wiki/modules/editor-ui-eigenpal.md` "Plugin registration § templatePlugin mode gating".
- **Observation:** The rule "do not re-add `templatePlugin` unconditionally to document-edit; use CSS to hide chips instead" is enforced in code and prose. No ADR captures the decision or its rationale.
- **Evidence:** `_artifacts/02-flow-plugin-registration.md`; missing entry in `wiki/decisions/` index.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · No ADR for Anti-Corruption Layer / wrapper-only consumption rule
- **Severity:** minor
- **Surface:** `packages/editor-ui/` as a whole; rule implied by ADR 0001 § Consequences ("All editor-related code consolidates in `packages/editor-ui/`"), `wiki/references/eigenpal-controlled-package.md` § "What belongs in MetalDocs docs".
- **Observation:** No ADR explicitly mandates that all `@eigenpal/docx-js-editor` access in `frontend/apps/web` goes through `@metaldocs/editor-ui`. T-002 (TemplateEditorPage bypass) was a consequence of this gap — now resolved; the rule still lacks a formal decision record.
- **Evidence:** `_artifacts/03-deps.md` direct-eigenpal IN-edges table.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-008`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 0 / 9 (all exports cited in §5.2 of `editor-ui-eigenpal.md`)
- Operations missing C4 placement: 0 / 0 (no HTTP)
- Cross-deps missing in §5/§8: 0 / 5
- State transitions missing in §6: 0 / 0 (no state machine)
- Decisions without ADR link: 2 / 8 (T-007, T-008 each carry a `missing-ADR` row; two ADRs would cover both — one for `templatePlugin` mode gating, one for the wrapper-only consumption boundary; T-002 and T-003 resolved)
