# Architecture: Data Model — LEGACY CURRENT-STATE REFERENCE

> **Status:** CURRENT DATABASE EVIDENCE ONLY — **NOT R10 TARGET DATA MODEL**  
> **Reclassified:** 2026-08-19 by operator-ratified T8-A

The current PostgreSQL schema is authoritative only for **what the current DEV/test implementation stores today**. It is not target persistence authority.

## Current target authority

Use:

- `launch-v1-product-contract.md`
- `r10-t1-semantic-state-invariants.md`
- `r10-t2-governance-effectivity-transactions.md`
- `r10-t3-authorization-audit-enforcement.md`
- `r10-t4-exact-content-storage-integrity-restore.md`
- `r10-t5-durable-async-search-external-effects.md`
- `r10-t7-historical-migration-truth-semantic-mapping.md`
- `r10-t8a-technical-authority-legacy-disposition.md`
- `r10-technical-architecture.md`

T8-D owns the exact target persistence realization.

## T8-A disposition

Do not preserve current tables, schemas, functions, triggers, RLS policies, GUCs, provider-key columns or parallel Document/ControlledDocument/Template/Approval families merely because they exist.

Current persistent shapes are **REWRITE** candidates and many legacy/non-Launch objects are **DELETE** candidates.

PostgreSQL itself remains preserved as the product-state substrate because upstream R10 decisions independently require it. That does not preserve current table layout or RLS posture.

## Current-state evidence

For archaeology of the existing implementation use:

- `wiki/database/index.md`
- `db/baseline/`
- current migrations/grants/reference data
- current repositories/SQL
- Git history

T7 established that current MetalDocs business data/history is DEV/test/throwaway and creates no business-data compatibility entitlement.

The former detailed data-model narrative remains in Git history.