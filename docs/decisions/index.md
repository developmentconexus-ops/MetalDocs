---
id: decision-registry
kind: authority
owner: architecture
summary: Compact index of current ratified MetalDocs decisions and their owning authority pages.
---

# Decision registry

This page is an index, not a duplicate architecture specification.

| Decision | Status | Authority |
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
| Historical migration truth | ratified | `../architecture/transition.md` |
| Clean-slate technical posture | ratified | `../architecture/technical-baseline.md` |
| Backend topology | ratified | `../architecture/backend.md` |
| Internal communication contracts | ratified | `../architecture/interfaces.md` |
| Persistence realization | ratified | `../architecture/persistence.md` |
| Repository clean-slate reset | approved | `repository-reset.md` |
| Executable API wire contract | active | `../work/current/proposal.md` |

## Provenance

Product/R10 authorities through persistence were ratified in PR #131. The repository reset preserves those authorities under semantic paths while removing the superseded implementation from the live tree.

A material contradiction reopens the owning decision; it is not resolved by silently editing this registry.