# R10-T7 — Historical Migration Truth & Semantic Mapping — Stage Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T7 OPEN / BUSINESS SOURCE CORPUS IDENTIFICATION NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

T7 is the first stage opened under the operator-ratified post-T6 Implementation Readiness Program.

T7 owns **historical/source truth and semantic migration mapping only**. It does not own target package/database/API/frontend/runtime realization and it does not own concrete migration/cutover implementation.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. Product Contract + Whole-Product GCR + 4+1 ownership
5. T1→T6 durable authorities
6. Decision Registry + D4/T6/post-T6 amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. `wiki/architecture/r10-technical-architecture.md`
10. `docs/superpowers/analysis/2026-08-19-r10-t7-source-corpus-operator-clarification.md`
11. this bootstrap
12. actual business-source evidence only after that source is identified

Legacy code/schema/content is evidence, never target authority by existence.

## 2. Operator-provided source-corpus fact

The operator explicitly established:

```text
current MetalDocs product data/history = DEV / TEST / THROWAWAY
no real business history currently lives in MetalDocs
```

Therefore:

```text
current MetalDocs DB rows                 NOT A MIGRATION SOURCE
current MetalDocs object-store contents   NOT A MIGRATION SOURCE
current MetalDocs Approval/Audit history  NOT A MIGRATION SOURCE
current MetalDocs Release/Template data   NOT A MIGRATION SOURCE
```

These surfaces remain technical legacy evidence for T8/T10 only.

The prior census direction that inspected current MetalDocs tables as if they might contain business history is superseded by this operator clarification.

## 3. T7 official scope

T7 decides only:

```text
whether Launch requires any pre-existing business-document corpus
actual source corpus/location/authority if one exists
PROVEN / INFERABLE WITH EXPLICIT RULE / UNKNOWN classification
CURRENT_STATE / FULL_HISTORY or a smaller real migration-mode set when applicable
which source facts become target-owned imported facts
which source facts remain provenance-only evidence
source document/revision identity quality
source revision/ordinal mapping
exact-content/file provenance quality
source actor/owner/governance provenance quality
semantic migration unit definition
truthful representation of partial/ambiguous/unknown historical evidence
```

T7 answers:

> **What real pre-R10 business truth, if any, must MetalDocs carry into Launch, and how may that truth map into the ratified semantic target without fabrication?**

## 4. Two legitimate T7 outcome families

### A. No historical business migration required

If no pre-existing controlled-document corpus must be present at Launch:

```text
Historical Migration required = NO
```

T7 may close with a bounded no-migration decision. This does not remove T10: T10 still owns technical current-code/schema replacement and cutover.

### B. External/manual business corpus required

If real controlled documents exist outside the current MetalDocs application, T7 must identify and inspect that actual source directly before choosing migration modes or semantic mapping.

Possible source mechanisms are not assumed. A file share, another system, a folder tree, paper/scanned records or another repository are possibilities only after evidence identifies one.

No generic ETL/interchange/repository connector platform is justified by possibility alone.

## 5. Explicitly out of T7

T7 does **not** decide:

```text
backend package/module topology                    → T8-B
internal package/owner communication realization  → T8-C
target relational schema/tables/constraints        → T8-D
exact executable OpenAPI schemas/codegen           → T8-E
frontend route/feature/query/cache topology        → T8-F
runtime binaries/processes/deploy/observability    → T8-G
whole physical realization coherence               → T8-H
Golden Flow proof architecture                      → T9
migration tooling/runtime implementation            → T10
dry-run implementation                              → T10
migration idempotency/reconciliation implementation → T10
production cutover/readiness/rollback/deletion      → T10
restore/offboarding cutover choreography            → T10
implementation decomposition                        → T11
product code                                        → BLOCKED
```

## 6. Inherited non-negotiable laws

```text
unknown source truth remains unknown
inference is labeled and governed by an explicit rule
imported history never becomes fake native MetalDocs history
migration convenience never changes Document/Revision/Submission/Release meaning
native governance evidence is never fabricated
provider/storage identity never becomes semantic identity
exact imported bytes remain governed by T4 exact-content law
current T3 security/offboarding truth remains authoritative for serving
historical side effects/jobs/notifications are never replayed as new current effects
Historical Migration is not a generic Interchange/ETL domain
prepare migration seams only where the actual migration requires them
```

## 7. Current source-evidence state

```text
CURRENT METALDOCS AS BUSINESS SOURCE = DISPROVEN / EXCLUDED
ACTUAL BUSINESS SOURCE CORPUS        = NOT IDENTIFIED
MIGRATION REQUIRED?                  = UNKNOWN UNTIL BUSINESS-CORPUS GATE
CURRENT_STATE mode                   = NOT DECIDABLE YET
FULL_HISTORY mode                    = NOT DECIDABLE YET
```

Do not analyze current MetalDocs dev rows to answer business-history questions.

## 8. First required work

Before any migration-semantic approach is proposed:

```text
identify whether Launch needs pre-existing business documents
```

If **NO**:

```text
derive bounded no-historical-migration T7 candidate
```

If **YES**:

```text
identify actual source corpus/location/authority
→ inspect that source directly
→ examine stable identifiers/codes
→ revision/version evidence
→ exact source bytes/formats
→ title/metadata history
→ area/type/template meaning
→ owner/author identities
→ governance/approval evidence where it actually exists
→ release/effectivity evidence where it actually exists
→ gaps/duplicates/contradictions
→ classify PROVEN / INFERABLE / UNKNOWN
```

No default may convert `UNKNOWN` into plausible history.

## 9. Global Maximum test

Reject:

```text
migrate current MetalDocs dev/test history because the tables already exist
```

Reject:

```text
build a generic import/ETL framework before a real source corpus is identified
```

Target:

```text
no migration when no business corpus is required
OR
actual business source evidence
+ explicit truth classification
+ smallest justified migration mode
+ target-owner semantic mapping
+ explicit provenance for uncertainty
→ truthful historical migration contract
```

## 10. T7 process

```text
business-corpus necessity gate
→ actual source identification when required
→ source evidence census
→ PROVEN / INFERABLE / UNKNOWN
→ compare 2–3 truthful migration-semantic approaches only if migration is required
→ Global Maximum candidate
→ material adjudication
→ platform-facing T7 summary
→ explicit operator summary ratification
→ durable T7 promotion + Registry reconciliation + staging cleanup
```

Only after T7 closes may T8 open.

T8→T12 and implementation remain blocked.
