# 0046 — Eigenpal Anti-Corruption Layer

- **Status:** Proposed
- **Last verified:** 2026-06-23
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

Introduce a single Anti-Corruption Layer package, `@metaldocs/eigenpal-adapter`, as the
**only** place in the repo permitted to depend on `@eigenpal/*`. The rest of the system
speaks MetalDocs vocabulary:

1. **Capabilities, not plugins.** Callers declare intent (`{ placeholders, outline, comments }`); the adapter privately realizes them (for eigenpal, by assembling `templatePlugin` + the sidebar plugin). No eigenpal plugin type crosses the boundary.
2. **Opaque document model.** Document content crosses the seam as `ArrayBuffer` (DOCX bytes), never as structured eigenpal nodes (`Paragraph`/`Table`/`BlockContent`), which are deep recursive OOXML no MetalDocs code manipulates.
3. **Narrow domain DTOs.** Comments cross as a MetalDocs `EditorComment` DTO; the row↔vendor mapping lives inside the adapter.
4. **Classified failures.** The adapter translates eigenpal throws into a MetalDocs `RenderError` taxonomy; the render sidecar returns a classified error contract; the Go worker classifies by status.
5. **Two source entrypoints.** A framework-free `.` (server-safe, no React) and a `./react` door, so the backend never bundles React.
6. **Enforcement.** A repo-wide ESLint `no-restricted-imports` rule bans `@eigenpal/*` everywhere except the adapter; only the adapter's `package.json` declares the vendor.

## Consequences

- A vendor-breaking upgrade or swap is contained to one package's two entry files.
- The phantom dependency is eliminated; the vendor name appears in exactly one `package.json`.
- New cost: the adapter must reproduce today's exact plugin wiring and own the comment row↔vendor mapping and the error taxonomy. Covered by porting existing tests.
- Supersedes the previous "editor-ui is the wrapper" arrangement; closes tech-debt T-008.

## Alternatives considered

- **Keep editor-ui as the wrapper (status quo).** Rejected: it re-exported raw vendor types, so it was a pass-through, not a boundary.
- **Reimplement DOCX fill in Go to drop the JS vendor server-side.** Rejected: creates two DOCX engines that must agree byte-for-byte forever, breaking the reconstruction hash audit; eigenpal is JS-only.
- **Full OOXML domain model mapped to MetalDocs types.** Rejected as YAGNI: the recursive node types are never manipulated by live code; an opaque buffer suffices.

## Related

- ADR 0001 (eigenpal adoption) — amended by this boundary.
- ADR 0025 (RFC 9457 problem+json) — the public-edge error envelope; the internal render contract is classified JSON, not problem+json.
- Spec: `docs/superpowers/specs/2026-06-23-eigenpal-adapter-acl-design.md`.
- ADR 0047 (templatePlugin mode-gating).
