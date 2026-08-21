---
id: documentation-index
kind: authority
owner: engineering
summary: Canonical task/intention router to the smallest current MetalDocs authority pack.
---

# MetalDocs documentation

Start here after `AGENTS.md`, then read `roadmap.md`, then only the owning documents needed for the task.

| Task / intention | Start with | Do not read by default |
|---|---|---|
| Program progression / implementation permission / next action | [Roadmap](roadmap.md) | Git history, closed PRs, old implementation |
| Product scope and concepts | [Product contract](product/contract.md) | Technical realization docs |
| Whole-product alignment / ownership | [Product alignment](product/alignment.md) + [Ownership](architecture/ownership.md) | Historical review artifacts |
| User journeys / application API meaning | [Product journeys](product/journeys.md) + [API operation census](decisions/api-operation-census.md) | Removed OpenAPI/generated code |
| Domain state / invariants | [Domain model](architecture/domain-model.md) | Runtime archaeology |
| Governance / effectivity / transactions | [Lifecycle](architecture/lifecycle.md) | Old Approval implementation |
| Authorization / Audit | [Authorization and audit](architecture/authorization-and-audit.md) | Provider roles / old ACL machinery |
| Responsible-owner eligibility | [Responsible owner](architecture/responsible-owner.md) | Old grant/role assumptions |
| Exact content / restore | [Content integrity](architecture/content-integrity.md) | Removed storage-provider contracts |
| Durable jobs / Search | [Async and search](architecture/async-and-search.md) | Old jobs registry |
| Technical clean-slate / reuse decision | [Technical baseline](architecture/technical-baseline.md) | Legacy source unless the reuse gate requires it |
| Backend topology | [Backend](architecture/backend.md) | Removed package tree |
| Internal owner contracts | [Interfaces](architecture/interfaces.md) | Foreign-SQL legacy evidence unless named |
| Persistence realization | [Persistence](architecture/persistence.md) | Old migrations/schema |
| Transition / migration posture | [Transition](architecture/transition.md) | Old cutover plans |
| Current decision dispositions | [Decision register](decisions/index.md) | Review chronology |
| Preserved / reopen / deferred obligations | [Forward obligations](decisions/forward-obligations.md) | Old decision-registry corpus |
| Repository reset / provenance | [Repository reset](decisions/repository-reset.md) | Unmerged branches unless exact provenance is required |
| Documentation governance | [Documentation rules](development/documentation.md) | Product architecture |
| Repository-local engineering / CI / Git | [Engineering rules](development/engineering-rules.md) | Global Method text duplicated locally |
| Executable application wire | [Wire contract](architecture/wire-contract.md) | Temporary work/review history, precision provenance, old OpenAPI |
| Frontend realization | [Frontend](architecture/frontend.md) + [T8-F ratification](decisions/t8f-ratification.md) | Removed legacy frontend / speculative runtime or visual-framework choices |
| Runtime / process / deployment | [Runtime](architecture/runtime.md) + [T8-G ratification](decisions/t8g-ratification.md) | Removed legacy runtime/deploy topology / speculative scale infrastructure |

## Reading law

Default fresh-actor route:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ 1–2 task authorities
```

Normal work stays at five files or fewer. Exceed that only for a named material reason. `docs/work/**`, research/evidence/history, Git history, and removed implementation are never default-read.
