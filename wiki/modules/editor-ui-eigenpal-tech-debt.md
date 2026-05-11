# Tech Debt Register — editor-ui-eigenpal

> Companion to `wiki/modules/editor-ui-eigenpal.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/editor-ui-eigenpal-refactor.md`.

**Last verified:** 2026-05-10

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Triggers used here:

- **Critical** — schema/version drift the boot check should catch but does not; or data-loss / supply-chain path where fresh installs break.
- **Major** — defense-in-depth gap; documented contract not followed; duplicated write surfaces; cross-module dependency that blocks another module's refactor; false-pass test risk on a load-bearing branch.
- **Minor** — latent symbol, doc/code drift on a non-load-bearing path, missing standalone ADR for a rule already enforced by code + tests.

## Items

### T-001 · Vendored eigenpal tarball absent from `main`
- **Severity:** critical
- **Surface:** `packages/editor-ui/package.json:29`, `apps/docgen-v2/package.json:15`, `frontend/apps/web/package.json:17` — each references `file:../../[…]/vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. Directory deleted by commit `0ee9160d` (2026-05-04).
- **Observation:** Three `package.json` files declare `@eigenpal/docx-js-editor` via a `file:` URI that resolves to a path that no longer exists on `main`. Fresh `npm install` (or `pnpm install --frozen-lockfile=false`) from a clean checkout fails to resolve the dep. Existing checkouts keep working off cached `node_modules/`. Lockfiles (`pnpm-lock.yaml`, `package-lock.json`) still carry the prior integrity hashes.
- **Evidence:** `_artifacts/00-context.md` "Eigenpal version pin / fork status"; `git show --stat 0ee9160d` confirms the tarball and README were deleted in the go-mod-vendor commit; `git ls-files vendor/eigenpal` returns empty.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-001`
- **Linked ADR:** `wiki/decisions/0001-eigenpal-adoption.md` (claim is now outdated; see also R-009)

### T-002 · TemplateEditorPage bypasses the `MetalDocsEditor` wrapper
- **Severity:** major
- **Surface:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:4-5,61,331` — imports `DocxEditor` directly from `@eigenpal/docx-js-editor/react` and `createEmptyDocument` from `@eigenpal/docx-js-editor/core`. Holds `useRef<DocxEditorRef>` rather than `MetalDocsEditorRef`.
- **Observation:** The adapter package's stated purpose (ADR 0001 + `wiki/modules/editor-ui-eigenpal.md`) is to centralize the seam between MetalDocs and `@eigenpal/docx-js-editor`. One of the two pages that should consume the seam reaches past it. Consequence: any wrapper-level concern (autosave debounce, plugin gating, ref shape, future telemetry) ships only for `DocumentEditorPage`. The existing wiki names `TemplateEditorPage` as a consumer; current code contradicts that.
- **Evidence:** Grep results in `_artifacts/03-deps.md` IN-edges table; `MetalDocsEditor` is referenced in 6 files under `features/documents/` and 0 under `features/templates/`.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-002`
- **Linked ADR:** missing-ADR (the rule "Anti-Corruption Layer between MetalDocs and eigenpal" is implied by ADR 0001 but never given its own decision record; see T-008).

### T-003 · `templatePlugin.wiring.test.tsx` asserts pre-gating contract
- **Severity:** major
- **Surface:** `packages/editor-ui/test/templatePlugin.wiring.test.tsx:29-34,36-55`
- **Observation:** Test expects `templatePlugin` to be included when `<MetalDocsEditor mode="document-edit" />`. Current production code (`MetalDocsEditor.tsx:56`) gates the plugin to `mode === 'template-draft'` and would yield `data-plugins='0'` for `document-edit` with no sidebar model. The test is out of sync with the 2026-05-06 plugin-gating refactor (per `editor-ui-eigenpal.md` changelog). Either the test currently fails or has been silently neutralized (no green/red signal recorded in the wiki).
- **Evidence:** `_artifacts/02-flow-plugin-registration.md` "Stale wiring spec" subsection; source comparison `MetalDocsEditor.tsx:55-59` vs `templatePlugin.wiring.test.tsx:30-34`.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-003`
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
- **Observation:** No ADR explicitly mandates that all `@eigenpal/docx-js-editor` access in `frontend/apps/web` goes through `@metaldocs/editor-ui`. T-002 (TemplateEditorPage bypass) is the live consequence of this gap.
- **Evidence:** `_artifacts/03-deps.md` direct-eigenpal IN-edges table.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-008`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 0 / 9 (all exports cited in §5.2 of `editor-ui-eigenpal.md`)
- Operations missing C4 placement: 0 / 0 (no HTTP)
- Cross-deps missing in §5/§8: 0 / 5
- State transitions missing in §6: 0 / 0 (no state machine)
- Decisions without ADR link: 4 / 8 (T-002, T-003, T-007, T-008 each carry a `missing-ADR` row; two ADRs would cover all four — one for `templatePlugin` mode gating, one for the wrapper-only consumption boundary)
