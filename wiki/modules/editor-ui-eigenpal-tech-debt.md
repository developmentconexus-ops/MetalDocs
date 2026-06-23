# Tech Debt Register — editor-ui-eigenpal

> Companion to `wiki/modules/editor-ui-eigenpal.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/editor-ui-eigenpal-refactor.md`.

**Last verified:** 2026-06-23

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Triggers used here:

- **Critical** — schema/version drift the boot check should catch but does not; or data-loss / supply-chain path where fresh installs break.
- **Major** — defense-in-depth gap; documented contract not followed; duplicated write surfaces; cross-module dependency that blocks another module's refactor; false-pass test risk on a load-bearing branch.
- **Minor** — latent symbol, doc/code drift on a non-load-bearing path, missing standalone ADR for a rule already enforced by code + tests.

## Items

### T-001 · Vendored eigenpal tarball absent from `main` — **RESOLVED Plan 3**
- **Severity:** critical → **resolved**
- **Surface:** previously `packages/editor-ui/package.json`, `apps/docgen-v2/package.json`, `frontend/apps/web/package.json` — each referenced `file:../../[…]/third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- **Resolution (2026-05-11):** Tarball restored at `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` from git history. **2026-06-14:** tarball relocated to `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. **2026-06-23:** vendored tarball fully retired; `@eigenpal/docx-editor-react@1.9.0` now installed from npm registry. Tarball path `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` deleted; `third_party/eigenpal/NOTICE` present. All `package.json` `file:` refs replaced with npm registry refs.
- **Evidence:** `@eigenpal/docx-editor-react@1.9.0` in `package.json` dependencies; `third_party/eigenpal/NOTICE` present.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-001` (closed)
- **Linked ADR:** `wiki/decisions/0001-eigenpal-adoption.md`

### T-002 · TemplateEditorPage bypasses the `MetalDocsEditor` wrapper — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` — previously imported `DocxEditor` directly from `@eigenpal/docx-editor-react`.
- **Resolution (2026-05-11, commit `60fa5473`):** `TemplateEditorPage` migrated to `MetalDocsEditor`. Direct `@eigenpal/docx-editor-react` imports removed; `useRef<MetalDocsEditorRef>` now used. Repo-wide grep `@eigenpal/docx-editor-react` in `frontend/apps/web/src` returns zero outside type-only positions. Anti-Corruption Layer now holds for both consumer pages.
- **Evidence:** `_artifacts/03-deps.md` IN-edges table (updated).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-002` (closed)
- **Linked ADR:** missing-ADR (see T-008 — the wrapper-only rule still lacks its own ADR; the package rename from `@eigenpal/docx-js-editor` to `@eigenpal/docx-editor-react` 2026-06-23 is covered by the ADR 0001 amendment)

### T-003 · `templatePlugin.wiring.test.tsx` asserts pre-gating contract — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `packages/editor-ui/test/templatePlugin.wiring.test.tsx`
- **Resolution (2026-05-11, commit `ce6d809a`):** Test rewritten to 5 correct assertions gated on `template-draft` mode, aligned with the 2026-05-06 plugin-gating refactor. `document-edit` path asserts `data-plugins='0'` for `templatePlugin`. No stale contract survives.
- **Evidence:** `_artifacts/02-flow-plugin-registration.md` "Stale wiring spec" subsection (now stale-free).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-003` (closed)
- **Linked ADR:** missing-ADR (see T-007 for the gating rule itself)

### T-007 · No ADR for `templatePlugin` mode-gating rule
- **Severity:** minor
- **Surface:** `packages/editor-ui/src/MetalDocsEditor.tsx:55-56`; rationale lives only in source comments and `wiki/modules/editor-ui-eigenpal.md` "Plugin registration § templatePlugin mode gating".
- **Observation:** The rule "do not re-add `templatePlugin` unconditionally to document-edit; use CSS to hide chips instead" is enforced in code and prose. No ADR captures the decision or its rationale.
- **Evidence:** `_artifacts/02-flow-plugin-registration.md`; missing entry in `wiki/decisions/` index.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-007`
- **Linked ADR:** ADR 0047 — decision recorded; implementation in Phase 3

### T-008 · No ADR for Anti-Corruption Layer / wrapper-only consumption rule
- **Severity:** minor
- **Surface:** `packages/editor-ui/` as a whole; rule implied by ADR 0001 § Consequences ("All editor-related code consolidates in `packages/editor-ui/`"), `wiki/references/eigenpal-controlled-package.md` § "What belongs in MetalDocs docs".
- **Observation:** No ADR explicitly mandates that all `@eigenpal/docx-editor-react` access in `frontend/apps/web` goes through `@metaldocs/editor-ui`. T-002 (TemplateEditorPage bypass) was a consequence of this gap — now resolved; the rule still lacks a formal decision record.
- **Evidence:** `_artifacts/03-deps.md` direct-eigenpal IN-edges table.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-008`
- **Linked ADR:** ADR 0046 — decision recorded; implementation in Phase 3

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 0 / 9 (all exports cited in §5.2 of `editor-ui-eigenpal.md`)
- Operations missing C4 placement: 0 / 0 (no HTTP)
- Cross-deps missing in §5/§8: 0 / 5
- State transitions missing in §6: 0 / 0 (no state machine)
- Decisions without ADR link: 0 / 5 (T-007 → ADR 0047, T-008 → ADR 0046; T-004, T-005, T-006 removed as stale 2026-06-23; T-002 and T-003 resolved)
