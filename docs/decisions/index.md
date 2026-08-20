---
id: decision-registry
kind: authority
owner: architecture
summary: Routes current ratified decisions, forward obligations, stage ownership, and the active T8-E checkpoint.
---

# Decision registry

This page is an index. Detailed current truth remains in the owning Product/architecture authority.

| Decision | State | Authority |
|---|---|---|
| Product contract | approved | `../product/contract.md` |
| Whole-product alignment | approved | `../product/alignment.md` |
| Semantic ownership | approved | `../architecture/ownership.md` |
| Domain state and invariants | ratified | `../architecture/domain-model.md` |
| Governance/effectivity transactions | ratified | `../architecture/lifecycle.md` |
| Authorization and Audit | ratified | `../architecture/authorization-and-audit.md` |
| Responsible-owner eligibility | ratified | `../architecture/responsible-owner.md` |
| Exact content / integrity / restore | ratified | `../architecture/content-integrity.md` |
| Durable async / Search / effects | ratified | `../architecture/async-and-search.md` |
| Canonical journeys | ratified | `../product/journeys.md` |
| Current `/api/v1` operation census | ratified precision | `api-operation-census.md` |
| Historical migration truth | ratified | `../architecture/transition.md` |
| Clean-slate technical posture | ratified | `../architecture/technical-baseline.md` |
| Backend topology | ratified | `../architecture/backend.md` |
| Internal communication contracts | ratified | `../architecture/interfaces.md` |
| Persistence realization | ratified | `../architecture/persistence.md` |
| Remaining T8-F→T12 program | ratified | `stage-program.md` |
| Cross-stage forward obligations | ratified baseline | `forward-obligations.md` |
| Repository clean-slate reset | operator-ratified | `repository-reset.md` |
| T8-E executable wire contract | active accepted checkpoint | `../reference/t8e-checkpoint.md` |

## Consumption law

Remaining stages consume current semantic authorities plus `forward-obligations.md`. A stage deliberately resolves only the `REOPEN` obligations it owns, treats `PRESERVE` as proof-backed baseline evidence, and keeps `DEFERRED` as future seams without dormant implementation.

A material contradiction reopens the owning decision; it is never reconciled by silently changing this index.

## Provenance

Product/R10 authorities through persistence were ratified in PR #131. Their pre-reset branch remains a reachable provenance ref until equivalent immutable archival tags exist. The live tree contains only current semantic authority plus the compact forward obligations that still matter.