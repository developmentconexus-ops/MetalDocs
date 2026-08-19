# R10-T7 — No Historical Business Migration Required — Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T7 CANDIDATE / OPERATOR SUMMARY RATIFICATION NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

## 1. Binding facts

The operator established two source-corpus facts:

```text
current MetalDocs DB/content/history = DEV / TEST / THROWAWAY
Launch does NOT require pre-existing business documents to be imported
```

Therefore:

```text
current MetalDocs as business source = EXCLUDED
external/manual source corpus         = NOT REQUIRED FOR LAUNCH
historical business migration         = NOT REQUIRED
```

No source evidence census beyond this necessity gate is required because there is no business corpus whose history must be carried forward.

## 2. Considered postures

### A — No historical business migration

Launch begins with R10-native governed documents only. No legacy business Document/Revision/Submission/Release history is imported.

**Benefits**
- preserves truthful provenance;
- eliminates accidental inheritance of DEV/test data;
- avoids dormant import/ETL abstractions;
- minimizes migration risk and implementation scope;
- keeps T10 focused on technical replacement/cutover only.

**Cost**
- any future desire to import a real external corpus becomes a new, explicitly scoped capability/reopen trigger.

### B — Build a generic historical-import seam anyway

Create generic source adapters, migration modes or provenance infrastructure despite no Launch corpus.

**Verdict:** REJECT.

Reason: speculative capability, violates YAGNI and `prepare seam, not dormant implementation`. T1 already preserves a native/imported provenance evolution seam if a real source appears later.

### C — Preserve/import current MetalDocs DEV data

Treat current DB/MinIO/Audit/Approval history as launch history for continuity.

**Verdict:** REJECT.

Reason: operator proved the data is DEV/test/throwaway. Importing it would fabricate business history and violate the Product Contract's `native history != imported history` and `unknown/source truth must not be invented` laws.

## 3. Global Maximum candidate

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Consequences:

1. R10 Launch creates only native post-cutover governed history.
2. No current MetalDocs product row/object/event is imported as business truth.
3. No historical Document/Revision/Submission/Decision/Release/Obsolescence records are synthesized.
4. No historical actors/timestamps are fabricated.
5. No generic importer, ETL framework, repository connector or migration mode is part of Launch.
6. T1's provenance/evolution seam remains sufficient for a future named source corpus.
7. T10 still owns technical transition from the current DEV implementation to R10: code/schema/API/frontend/runtime replacement, environment reset/cutover, and removal of disposable DEV state.
8. T8 must not preserve legacy persistence/API/module structures merely for historical-data compatibility, because no such compatibility consumer exists.

## 4. T10 boundary after this decision

T10 remains required, but its business-history-import branch is empty for Launch.

T10 owns:

```text
current-code → target-code transition
current-schema → target-schema transition
current API/frontend/runtime replacement
DEV/test data disposal/reset semantics
deployment cutover/rollback/readiness
legacy technical deletion map
```

T10 does **not** own:

```text
historical business document import
historical approval reconstruction
historical release reconstruction
legacy business actor/time backfill
```

unless a future explicit T7 reopen introduces a real business source corpus.

## 5. Reopen triggers

T7 reopens only if, before implementation/final ratification or in a later release, a concrete consumer appears such as:

```text
a named external/manual corpus that must be present in MetalDocs
a contractual/regulatory requirement to preserve pre-R10 controlled-document history
a production MetalDocs dataset created before R10 cutover that becomes business-authoritative
a named merger/import/onboarding requirement requiring historical provenance
```

A hypothetical future import is not a trigger.

## 6. Candidate verdict

```text
business corpus required at Launch              NO
current MetalDocs business history              NONE
historical business migration                   NOT REQUIRED
generic migration/import platform               REJECT
DEV/test history preservation                   REJECT
T10 technical transition                        STILL REQUIRED
T8→T12                                           STILL REQUIRED
implementation                                  BLOCKED
```

This candidate is ready for platform-facing summary ratification. It is not durable T7 authority until the operator explicitly ratifies the summary.
