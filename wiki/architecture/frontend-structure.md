# Frontend Structure — LEGACY CURRENT-STATE REFERENCE

> **Status:** CURRENT-STATE / HISTORICAL IMPLEMENTATION EVIDENCE ONLY  
> **Reclassified:** 2026-08-19 by operator-ratified T8-A  
> **Not R10 target frontend authority.**

This page formerly defined the canonical feature-sliced frontend tree, route ownership, TanStack Query usage, Zustand rules, `ArtifactViewModel`, legacy feature domains and shell conventions.

Those physical choices are not inherited by R10.

## Current target authority

T6 defines the user-facing semantic lenses and wire behavior. T8-F will derive the target frontend realization.

Use:

- `r10-t6-canonical-api-frontend-journeys.md`
- `r10-t8a-technical-authority-legacy-disposition.md`
- `rebaseline-decision-registry-t8a-amendment.md`
- `r10-technical-architecture.md`

## Current-state evidence only

The present implementation may still provide useful evidence about:

```text
React/Vite mechanics
routing behavior
TanStack Query patterns
generated TypeScript transport
error/loading UX
editor/viewer integration
CSS/design-system assets
```

But T8-A explicitly leaves undecided whether React Router, TanStack Query, Zustand, the current folder tree or any legacy feature-domain decomposition survives.

The current feature topology built around `approval`, `documents`, `templates`, `tokens`, `taxonomy`, `iam`, password-change and similar legacy domains is **REWRITE / REHOME** candidate, not target authority.

The former detailed structure remains available in Git history for archaeology.