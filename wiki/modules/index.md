# Modules — Current Runtime Evidence Only

> **Last verified:** 2026-08-19  
> **Status:** CURRENT-STATE / HISTORICAL IMPLEMENTATION EVIDENCE — **NOT R10 TARGET OWNERSHIP**

The pages under `wiki/modules/` describe the implementation that exists or existed before the R10 physical rebaseline. They are useful for archaeology, transition analysis and locating current code, but they do **not** define the target bounded contexts, package topology, ownership model or Launch capability set.

## Current target authority

Read instead:

- `../architecture/launch-v1-product-contract.md`
- `../architecture/whole-product-alignment-review.md`
- `../architecture/launch-v1-ownership-topology.md`
- `../architecture/r10-t1-semantic-state-invariants.md` through the latest promoted R10 stage
- `../architecture/r10-t8a-technical-authority-legacy-disposition.md`
- `../architecture/r10-technical-architecture.md` — sole current stage/status/next-action router
- `../references/current-agent-handoff.md`

## Binding T8-A rule

```text
current module existence = evidence only
legacy module count/topology = no survival entitlement
PRESERVE must be proved
REWRITE / REHOME / DELETE are valid outcomes
```

T8-A specifically classifies the legacy semantic module/package topology and its cross-owner/foreign-SQL boundaries as **REWRITE / REHOME**, with replacement topology owned by T8-B/T8-C.

## Current module pages

The existing pages for `approval`, `auth`, `iam`, `controlled-documents`, `documents`, `taxonomy`, `templates`, `jobs`, `audit`, `distribution`, `notifications`, `render`, `search`, `security`, `tokens`, and related tech-debt/artifact pages are all **current-state or historical evidence by default**.

No individual module page may override the ratified 4+1 semantic ownership topology:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit (supporting)
```

That semantic ownership does not itself preselect exact Go packages; T8-B owns the physical backend/package topology.

## Frontend/module pages

Frontend/editor module pages are also current-state evidence. T6 defines the user-facing semantic lenses; T8-F will derive the physical frontend realization.

## Rule

When a module page uses words such as `canonical`, `target`, `MUST`, `owner` or `bounded context`, interpret them as describing the historical/current implementation unless a current R10 authority explicitly promotes the same property.

Git history is the archive for deeper legacy detail.