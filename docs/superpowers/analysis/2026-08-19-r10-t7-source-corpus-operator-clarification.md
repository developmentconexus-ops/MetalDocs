# R10-T7 — Source Corpus Operator Clarification

> **Status:** ACTIVE STAGING / OPERATOR-PROVIDED SOURCE FACT  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

## Operator clarification

The operator explicitly clarified on 2026-08-19:

```text
the current MetalDocs database/history/content is all development/test data
there is no current real business history in MetalDocs
all current MetalDocs product data is disposable for migration purposes
```

This is binding T7 source-corpus evidence because deployment/business reality cannot be derived from repository schema shape alone.

## Consequence

```text
current MetalDocs DB rows                 = DEV/TEST ONLY
current MetalDocs object-store contents   = DEV/TEST ONLY
current MetalDocs Approval history        = DEV/TEST ONLY
current MetalDocs Audit history           = DEV/TEST ONLY
current MetalDocs Release history         = DEV/TEST ONLY
current MetalDocs Template history        = DEV/TEST ONLY

migration-source status                   = EXCLUDED
business-history authority                = NONE
```

No current MetalDocs row, object, audit event, approval decision, release generation, user action or technical revision may be promoted into R10 imported business history merely because the current implementation can represent it.

The current schema/code/data remain useful only as **technical legacy evidence** for T8/T10 when designing/refactoring the implementation.

## T7 Structural Inversion

The previous census direction implicitly treated the current MetalDocs application as the historical source system. That premise is now disproved.

The T7 question becomes:

> **Is there any real pre-R10 business corpus that must be present at Launch, and if so, what is its actual source and what truth can that source prove?**

Until an actual business source is identified:

```text
ACTUAL BUSINESS SOURCE CORPUS = NOT IDENTIFIED
CURRENT_STATE migration mode  = NOT DECIDABLE
FULL_HISTORY migration mode   = NOT DECIDABLE
historical mapping            = NOT DECIDABLE
```

Unknown remains unknown.

## Two legitimate outcomes

### Outcome A — no historical business corpus is required at Launch

If the business confirms that no pre-existing controlled documents need to enter the new MetalDocs at go-live:

```text
Historical Migration required = NO
```

T7 should close with a bounded **NO HISTORICAL BUSINESS MIGRATION REQUIRED** decision. T10 would then own only the technical current-code/schema replacement/cutover, not business-history import.

### Outcome B — an external/manual business corpus must be imported

If real controlled documents exist outside current MetalDocs (for example files, folders, another system or another authoritative business repository), T7 must inspect that actual source directly before selecting any migration mode or historical mapping.

No generic connector/ETL platform is justified by this possibility alone.

## Hard prohibition

Do not resume source-history analysis from:

```text
public.documents
public.controlled_documents
public.document_revisions
public.approval_*
public.release_generations
public.templates_*
metaldocs.audit_events
current MinIO objects
current dev users/groups
```

for business migration semantics.

Those surfaces describe/dev-test the current implementation only.

## Exact next gate

```text
identify whether Launch requires any pre-existing business documents
→ if NO: derive bounded T7 no-migration candidate
→ if YES: identify actual source corpus/location/authority
→ inspect that corpus directly
→ classify its facts PROVEN / INFERABLE / UNKNOWN
→ only then compare migration-truth approaches
```

T8→T12 and implementation remain blocked.
