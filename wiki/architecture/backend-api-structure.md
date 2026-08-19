# Backend/API Structure — LEGACY CURRENT-STATE REFERENCE

> **Status:** CURRENT-STATE / HISTORICAL IMPLEMENTATION EVIDENCE ONLY  
> **Reclassified:** 2026-08-19 by operator-ratified T8-A  
> **Not R10 target wire/package authority.**

This page formerly defined canonical module-tag HTTP ownership, generated package layout, `SurfacePublisher`, route mounting and migration rules for the previous backend topology.

Those physical rules are not inherited by R10.

## What survives

Current R10 independently preserves:

```text
OpenAPI 3.0.3 contract-first discipline
generated Go boundary
generated TypeScript boundary
contract/codegen drift proof
RFC 9457 behavior via T6
```

## What does not survive by inheritance

```text
one legacy module = one OpenAPI tag
one generated package per legacy module
current SurfacePublisher ownership shape
current route mounting topology
current OpenAPI paths/schemas
legacy auth/tenant/template/taxonomy/approval surfaces
```

T8-E owns the exact executable wire contract. T8-B/T8-C own the target backend/package and internal ownership boundaries that support it.

## Current authority

Use:

- `r10-t6-canonical-api-frontend-journeys.md`
- `r10-t8a-technical-authority-legacy-disposition.md`
- `rebaseline-decision-registry-t8a-amendment.md`
- `r10-technical-architecture.md`

The former detailed rules remain available in Git history as current-state archaeology and possible mechanism evidence.