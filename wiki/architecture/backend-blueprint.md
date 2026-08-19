# Backend Blueprint — LEGACY CURRENT-STATE REFERENCE

> **Status:** CURRENT-STATE / HISTORICAL IMPLEMENTATION EVIDENCE ONLY  
> **Reclassified:** 2026-08-19 by operator-ratified T8-A  
> **Not R10 target authority.**

This page formerly described the current/previous MetalDocs backend as a canonical 15-module modular monolith with its platform packages, middleware, async stack and maturity grades.

That physical topology is no longer a target contract.

## Current authority

Use:

- `launch-v1-product-contract.md`
- `whole-product-alignment-review.md`
- `launch-v1-ownership-topology.md`
- `r10-t1-semantic-state-invariants.md` through the latest promoted R10 stage
- `r10-t8a-technical-authority-legacy-disposition.md`
- `rebaseline-decision-registry-t8a-amendment.md`
- `r10-technical-architecture.md` — sole current stage/status/next-action router

## T8-A disposition

The former blueprint's:

```text
15 business modules
37 platform-package map
legacy composition-root shape
pooled tenant/RLS substrate
old AuthN/AuthZ/module ownership
current async/process topology
current provider choices
```

are **current-state evidence only** and receive no preservation entitlement.

Properties such as PostgreSQL, River for named durable jobs, contract-first generated boundaries, runtime DB least privilege, observability principles and verification controls survive only where current R10 independently preserves them.

T8-B→T8-G derive the actual target physical architecture. T10 owns current→target transition/deletion.

The former detailed blueprint is preserved in Git history for archaeology.