# 0046 — Eigenpal Anti-Corruption Layer

- **Status:** Accepted
- **Last verified:** 2026-06-26
- **Scope:** The boundary between MetalDocs and the `@eigenpal/*` DOCX editor/engine vendor, across the browser editor (`packages/editor-ui`) and the server render sidecar (`apps/docx-renderer`).

## Context

`@eigenpal/*` is a third-party vendor. The integration leaked the vendor across the
codebase: raw eigenpal types were re-exported through `@metaldocs/editor-ui`
(`BlockContent`, `Paragraph`, `Table`, `Comment`), `apps/docx-renderer` imported the
vendor's `headless` entrypoint directly, and `@eigenpal/docx-editor-core` was a phantom
dependency (declared by no `package.json`, resolved only transitively via
`@eigenpal/docx-editor-react`). The defining failure test — "eigenpal ships a breaking
2.0, or we swap vendors; how many files change?" — answered "many, scattered."

ADR 0001 adopted eigenpal but did not record a containment boundary.

## Decision

Contain the vendor behind an Anti-Corruption Layer realized as **two walls**, split by
runtime so the backend never bundles React. Each wall imports `@eigenpal/*` privately
but exposes a vendor-free public surface; the rest of the system speaks MetalDocs
vocabulary.

- **Server wall — `@metaldocs/eigenpal-adapter`.** A framework-free (`.`) entrypoint,
  safe to bundle in the Node render sidecar. Imports `@eigenpal/docx-editor-core`
  internally; exposes the render capability, the `RenderError` taxonomy, and the
  template processor. `apps/docx-renderer` depends only on this adapter, never on the
  vendor.
- **Browser wall — `@metaldocs/editor-ui`.** The React editor wall. Imports
  `@eigenpal/docx-editor-react` (+ core types) internally; exposes `MetalDocsEditor`
  and the `EditorComment` DTO. `frontend/apps/web` depends only on this package.

Within that two-wall structure:

1. **Capabilities, not plugins.** Callers declare intent (`{ placeholders, outline, comments }`); each wall privately realizes them (for eigenpal, by assembling `templatePlugin` + the sidebar plugin). No eigenpal plugin type crosses a wall.
2. **Opaque document model.** Document content crosses each seam as `ArrayBuffer` (DOCX bytes), never as structured eigenpal nodes (`Paragraph`/`Table`/`BlockContent`), which are deep recursive OOXML no MetalDocs code manipulates.
3. **Narrow domain DTOs.** Comments cross as a MetalDocs `EditorComment` DTO; the row↔vendor mapping lives inside the browser wall (`packages/editor-ui/src/comment-mapping.ts`).
4. **Classified failures.** The server wall translates eigenpal throws into a MetalDocs `RenderError` taxonomy; the render sidecar returns a classified error contract; the Go worker classifies by status.
5. **Runtime split is the seam.** The server wall is React-free so the backend never bundles React; the browser wall owns all React/editor concerns. There is one entrypoint per wall — the adapter exposes `.` only (no `./react`; the React door is the separate editor-ui package).
6. **Enforcement.** A repo-wide ESLint `no-restricted-imports` rule (`eslint.config.mjs`, gated by `.github/workflows/lint.yml`) bans `@eigenpal/*` everywhere except the two walls. Guard tests back it: `packages/editor-ui/test/public-surface.test.ts` keeps the browser wall's public surface vendor-free, and `apps/docx-renderer/src/render/__tests__/bundle-guard.test.ts` keeps the vendor out of the sidecar bundle. The vendor name appears in exactly two `package.json` files (the two walls).

## Consequences

- A vendor-breaking upgrade or swap is contained to the two wall packages' internals.
- The phantom dependency is eliminated; the vendor name appears in exactly two `package.json` files (`eigenpal-adapter`, `editor-ui`) — both walls, nowhere else.
- New cost: each wall must reproduce today's exact plugin wiring; the browser wall owns the comment row↔vendor mapping and the server wall owns the error taxonomy. Covered by porting existing tests.
- Supersedes the previous "editor-ui is the wrapper" pass-through arrangement; closes tech-debt T-008.

## Alternatives considered

- **Keep editor-ui as the wrapper (status quo).** Rejected: it re-exported raw vendor types, so it was a pass-through, not a boundary.
- **Single adapter package with `.` + `./react` doors.** The original framing. Rejected in implementation: the React editor is large and browser-only, so the React door is its own package (`editor-ui`) rather than a second entrypoint of the server adapter — keeping the server wall React-free is cleaner than gating React behind a subpath export.
- **Reimplement DOCX fill in Go to drop the JS vendor server-side.** Rejected: creates two DOCX engines that must agree byte-for-byte forever, breaking the reconstruction hash audit; eigenpal is JS-only.
- **Full OOXML domain model mapped to MetalDocs types.** Rejected as YAGNI: the recursive node types are never manipulated by live code; an opaque buffer suffices.

## Related

- ADR 0001 (eigenpal adoption) — amended by this boundary.
- ADR 0025 (RFC 9457 problem+json) — the public-edge error envelope; the internal render contract is classified JSON, not problem+json.
- Spec: `docs/superpowers/specs/2026-06-23-eigenpal-adapter-acl-design.md`.
- ADR 0047 (templatePlugin mode-gating).
