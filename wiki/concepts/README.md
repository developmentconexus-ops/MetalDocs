# Concepts

> **Last verified:** 2026-05-04
> **Scope:** Cross-cutting ideas referenced from many places. Read these to understand WHY before HOW.

- [placeholders.md](placeholders.md) — eigenpal native vs MetalDocs legacy. **Read first.**
- [token-syntax.md](token-syntax.md) — `{name}` vs `{{uuid}}` deep dive
- controlled-documents.md — TBD
- iso-segregation.md — TBD
- freeze-and-hashing.md — TBD
- [error-ux.md](error-ux.md) — shared `apiFetch` / `ApiError` / auth-bus / `resolveErrorMessage` (E2/E3/E4)
- [design-workflow-audit.md](design-workflow-audit.md) — audit AI-generated `design-source/` mockups against real workflow / RBAC / personas before implementing
- [css-leakage-offenders.md](css-leakage-offenders.md) — global `.input { height: 32px }` and other rules that silently clobber component styles; override patterns for `<textarea>`
