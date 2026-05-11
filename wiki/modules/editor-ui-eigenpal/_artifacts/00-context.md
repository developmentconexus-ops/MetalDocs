# Phase 0 — Context Load

> Module: `editor-ui-eigenpal` (FE-only adapter package: `packages/editor-ui/`)
> Date: 2026-05-10

## Inputs read

- `wiki/README.md` — index entry for module + cross-deps to editor-chrome, placeholders, token-syntax, eigenpal-spike, eigenpal-controlled-package, ADR 0001.
- `wiki/modules/editor-ui-eigenpal.md` (existing stub — to be replaced by Arc42 doc).
- `wiki/modules/editor-chrome.md` (shipped; consumer-side primitive of this adapter).
- `wiki/concepts/placeholders.md` — fixed 7-token catalog, tokens literal until server freeze.
- `wiki/concepts/token-syntax.md` — `{name}` eigenpal-native; `{{uuid}}` legacy removed 2026-04-25.
- `wiki/decisions/0001-eigenpal-adoption.md` — accepted ADR; cites vendored `0.2.0.tgz`.
- `wiki/references/eigenpal-spike.md` — T1–T8 outcomes that seeded plugin selection.
- `wiki/references/eigenpal-controlled-package.md` — claims `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` artifact.

## Module shape

- Tiny FE adapter: ~3 source files + 3 plugin files.
- One public component (`MetalDocsEditor`) + 3 plugin exports (`buildSidebarModelPlugin`, `createOutlinePlugin`, `computeSidebarModel`).
- No HTTP. No DB. No state machine. No errors surfaced to API (RFC 9457 = n/a).
- Wraps external lib `@eigenpal/docx-js-editor` as the SEAM.

## Eigenpal version pin / fork status

- `package.json` (3 places) reference `file:.../vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- **FINDING — tarball missing from repo:** commit `0ee9160d` (2026-05-04, "chore(vendor): replace custom eigenpal vendor entry with go mod vendor output") deleted both `vendor/eigenpal/README.md` and `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. The three `package.json` `file:` URIs still resolve to the deleted path. Fresh `npm install` would fail. Lockfiles still carry the old integrity hashes so existing checkouts keep working off `node_modules/`.
- Fork status: controlled fork (per ADR 0001 + `vendor/eigenpal/README.md` reference, also deleted by the same commit). Upstream consolidation deferred. Lab-side source: `non_git/eigenpal-isolated-lab/analysis/eigenpal-upstream-source/`.

## IN-edges (consumers) — verified

- `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:241` — uses `MetalDocsEditor` with `mode='document-edit'|'readonly'`. **Sole real consumer of the wrapper.**
- `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:4` — **bypasses the wrapper**: imports `DocxEditor` directly from `@eigenpal/docx-js-editor/react`. Existing wiki claims this page is a `MetalDocsEditor` consumer. **STALE.**
- Type-only imports: `useDocumentComments.ts` etc. import `type { Comment }` from `@metaldocs/editor-ui` (re-exported pass-through).

## OUT-edges

- `@eigenpal/docx-js-editor` — sole runtime dependency (peer: react 18.2).
- `@metaldocs/shared-tokens` — feeds `computeSidebarModel`.

## Carry-forward gaps (debt seeds)

1. Vendor tarball missing on `main`; install break — **Critical** (supply-chain availability).
2. TemplateEditorPage bypasses the wrapper — adapter has effectively one consumer; doc/code drift — **Major**.
3. `templatePlugin.wiring.test.tsx` asserts `templatePlugin` is included for `mode='document-edit'`; current code gates it to `template-draft` only — **Major** (false-pass / stale spec).
4. `mergefieldPlugin.ts` exports `computeSidebarModel` but is never registered as a plugin; the file's only consumer is the sidebar-bridge (data shape). Wiki's "VERIFY" status still open — **Minor**.
5. `OutlinePlugin` (`createOutlinePlugin`) is exported but removed from the plugins array in `MetalDocsEditor.tsx` — dormant code — **Minor**.
6. `applyVariables` semantics: tokens stay literal until server freeze. Adapter contract is clear; no write-path emits substituted DOCX in writer mode. Confirmed by code inspection of `MetalDocsEditor.tsx` (no `applyVariables` call) — **no debt** (deferred preview-mode design per ADR 0008).

## Open questions for the user

None blocking. Proceeding to Phase 1–6 inline (module is small enough not to warrant 4 subagent dispatches; user mandate "push back on additions" applies). Artifacts will still be produced per gate.

## Skips recorded

- Phase 1/2/4 done inline by main agent (no Codex dispatch). Reason: ~7 source files, no HTTP, no DB, no migrations — subagent setup overhead exceeds value. Each phase still produces its artifact.
- Phase 3 done inline (grep already executed during Phase 0 IN-edge verification; recording in `03-deps.md`).
- RFC 9457 row: n/a — adapter surfaces no errors to API layer.
- §6 state-machine row: n/a — no state machine. Mode prop is a simple discriminator (template-draft / document-edit / readonly).
