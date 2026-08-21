---
id: decision-register
kind: authority
owner: architecture
summary: Compact current decision/disposition register for Product and R10.
---

# MetalDocs decision register

This is the compact current register. It points to the owning semantic authority instead of recreating the old historical ledger.

## Current decisions

| ID | Subject | Disposition | Current decision | Authority |
|---|---|---|---|---|
| PRODUCT | Launch Product contract | CURRENT / APPROVED | Controlled-document Launch scope is binding | `../product/contract.md` |
| ALIGN | Whole-product alignment | CURRENT / APPROVED | Product-wide alignment baseline is binding | `../product/alignment.md` |
| OWN | Semantic ownership | CURRENT / APPROVED | Authentication, Organization, Authorization, Controlled Documents + supporting Audit | `../architecture/ownership.md` |
| T1 | Domain state/invariants | CURRENT / RATIFIED | Stable semantic state/lifecycle vocabulary | `../architecture/domain-model.md` |
| T2 | Governance/effectivity transactions | CURRENT / RATIFIED | Transaction, serialization, Release/effectivity laws | `../architecture/lifecycle.md` |
| T3 | Authorization/Audit | CURRENT / RATIFIED | Live grant/scope/domain authorization + bounded same-commit Audit | `../architecture/authorization-and-audit.md` |
| T3-D4 | Responsible-owner eligibility | REFINED / RATIFIED | New owner must be existing, same-Company, ENABLED User | `../architecture/responsible-owner.md` |
| T4 | Exact content / restore | CURRENT / RATIFIED | Exact-content descriptor truth, managed-content seam, fail-closed restore | `../architecture/content-integrity.md` |
| T5 | Async/Search/effects | CURRENT / RATIFIED | River-backed named durable effects; canonical PostgreSQL Search baseline | `../architecture/async-and-search.md` |
| T6 | Canonical journeys | CURRENT / RATIFIED | Product/API/frontend journey semantics | `../product/journeys.md` |
| T6-API | `/api/v1` application census | REFINED / RATIFIED PRECISION | Current census = 78 operations; no operation 79 without material reopen | `api-operation-census.md` + `../product/journeys.md` |
| T7 | Historical migration truth | CURRENT / RATIFIED | No historical business migration required for Launch | `../architecture/transition.md` |
| T8-A | Clean-slate technical posture | CURRENT / RATIFIED | Legacy physical shape has no survival entitlement | `../architecture/technical-baseline.md` |
| T8-B | Backend topology | CURRENT / RATIFIED | Owner-first modular monolith realization baseline | `../architecture/backend.md` |
| T8-C | Internal contracts | CURRENT / RATIFIED | Authority-aligned owner/resolver contract model | `../architecture/interfaces.md` |
| T8-D | Persistence | CURRENT / RATIFIED | Owner-namespaced PostgreSQL relational core and concurrency realization | `../architecture/persistence.md` |
| T8-E | Executable wire contract | CURRENT / RATIFIED | Exact 78-operation OAS 3.0.3 application wire is binding; no operation 79 without material Product/T6 reopen | `../architecture/wire-contract.md` |
| RESET | Repository clean-slate reset | CURRENT / OPERATOR-RATIFIED | Superseded implementation remains absent; required provenance stays reachable | `repository-reset.md` |
| REPO-STD-V1 | Repository operating envelope | CURRENT ORGANIZATIONAL BASELINE / ALIGNED | Repository Standard v1 alignment is merged; local repository controls preserve that operating envelope | `../roadmap.md` + `../development/documentation.md` + `../development/engineering-rules.md` |

## Forward obligations

The old decision corpus is intentionally not restored wholesale. Cross-stage obligations that still matter are preserved in `forward-obligations.md`:

```text
PRESERVE  21
REOPEN     3
DEFERRED  27
TOTAL     51
```

`CURRENT` semantics already live in the owning authorities above. `SUPERSEDED` mechanics remain deleted unless new material evidence explicitly reopens them.

## Consumption law

Remaining stages:

```text
consume current owning authorities
+ consume PRESERVE as proof-backed baseline
+ deliberately decide only stage-relevant REOPEN items
+ retain DEFERRED as future seams/counterexamples
- never inherit SUPERSEDED implementation by default
```

Reviewer findings and historical registry rows are Evidence, not independent requirement authority. A material contradiction reopens only the owning decision it actually implicates.
