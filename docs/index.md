---
id: documentation-index
kind: authority
owner: engineering
summary: Routes humans and agents to the smallest current MetalDocs authority set.
---

# MetalDocs documentation

Use this page as the documentation entrypoint.

| Need | Read |
|---|---|
| Current stage / whether implementation is allowed | [Status](status.md) |
| Product scope and concepts | [Product contract](product/contract.md) |
| Whole-product alignment | [Product alignment](product/alignment.md) |
| Canonical user/API journeys | [Product journeys](product/journeys.md) |
| Current `/api/v1` operation census | [API operation census](decisions/api-operation-census.md) |
| Semantic owners | [Ownership](architecture/ownership.md) |
| Domain concepts and invariants | [Domain model](architecture/domain-model.md) |
| Governance/effectivity transactions | [Lifecycle](architecture/lifecycle.md) |
| Authorization and Audit | [Authorization and audit](architecture/authorization-and-audit.md) |
| Responsible-owner eligibility | [Responsible owner](architecture/responsible-owner.md) |
| Exact content / integrity / restore | [Content integrity](architecture/content-integrity.md) |
| Durable jobs and Search | [Async and search](architecture/async-and-search.md) |
| Migration/compatibility posture | [Transition](architecture/transition.md) |
| What may be reused from removed implementation | [Technical baseline](architecture/technical-baseline.md) |
| Backend target topology | [Backend](architecture/backend.md) |
| Internal owner contracts | [Interfaces](architecture/interfaces.md) |
| Persistence realization | [Persistence](architecture/persistence.md) |
| Remaining T8-F→T12 stage ownership | [Stage program](decisions/stage-program.md) |
| Preserved future/reopen constraints | [Forward obligations](decisions/forward-obligations.md) |
| Decision routing | [Decision registry](decisions/index.md) |
| Why the legacy repository was removed | [Repository reset](decisions/repository-reset.md) |
| Repository documentation rules | [Documentation governance](development/documentation.md) |
| Engineering / PR / review rules | [Engineering rules](development/engineering-rules.md) |
| Paused accepted T8-E checkpoint | [T8-E checkpoint](reference/t8e-checkpoint.md) |

## Active Draft work

`docs/work/current/` is temporary and exists only while a governed Draft PR is under review. `AGENTS.md` routes to it conditionally when the directory exists; durable navigation never depends on it.

## Reading law

Routine work should normally require this index, current status, and 1–3 owning authorities. Any `wiki/...` path embedded inside a carried pre-reset authority is provenance only and MUST NOT be followed as current routing. Current routing is defined by this index, `status.md`, and `decisions/index.md`.

Git history, closed PRs, and removed implementation are provenance only unless a current authority explicitly requests them as evidence.