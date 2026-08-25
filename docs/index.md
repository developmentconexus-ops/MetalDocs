---
id: documentation-index
kind: authority
owner: engineering
summary: Canonical task/intention router to the smallest current MetalDocs authority pack.
---

# MetalDocs documentation

Start here after `AGENTS.md`, then read `roadmap.md`, then follow the exact pinned methodology `ROUTER.md` named in `AGENTS.md` when method selection is required, then only the owning repository documents needed for the task.

| Task / intention | Start with | Do not read by default |
|---|---|---|
| Program progression / implementation permission / next action | [Roadmap](roadmap.md) | Git history, closed PRs, old implementation |
| Product scope and concepts | [Product contract](product/contract.md) + [Discussion/Notifications bounded authority](decisions/discussion-notifications-launch.md) when that capability is implicated | Technical realization docs |
| Whole-product alignment / ownership | [Product alignment](product/alignment.md) + [Ownership](architecture/ownership.md) | Historical review artifacts |
| Stable-Document Discussion / @Mention / Notifications / Inbox / realtime notification wake-up | [Discussion/Notifications bounded authority](decisions/discussion-notifications-launch.md) | Temporary T11 work, Fable dialogue, legacy frontend |
| Notification Full Inbox source recognition / `listNotifications` item projection | [Notification Inbox recognition precision](decisions/notification-inbox-recognition-read.md) + [Discussion/Notifications bounded authority](decisions/discussion-notifications-launch.md) | Browser per-row source fan-out, copied source ACL/content, temporary B08 work |
| User journeys / application API meaning | [Product journeys](product/journeys.md) + [API operation census](decisions/api-operation-census.md); add [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) for operations 79–86, [My Work governance read precision](decisions/my-work-governance-identification-read.md) when governance-work row identity is implicated, and [Access Assignment read](decisions/access-assignment-read.md) when op31 access inspection is implicated | Removed OpenAPI/generated code |
| Domain state / invariants | [Domain model](architecture/domain-model.md) | Runtime archaeology |
| Governance / effectivity / transactions | [Lifecycle](architecture/lifecycle.md) | Old Approval implementation |
| Authorization / Audit | [Authorization and audit](architecture/authorization-and-audit.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) when `document.discuss`/presentability is implicated + [Audit investigation read](decisions/audit-investigation-read.md) when B09 query/recognition is implicated | Provider roles / old ACL machinery |
| Access Administration / RoleAssignment inspection by User, Group, Area/Company scope or Role | [Access Assignment read](decisions/access-assignment-read.md) + [Authorization and audit](architecture/authorization-and-audit.md) + [API operation census](decisions/api-operation-census.md) | Temporary B11 wireframes, client-computed effective-access graphs, generic IAM machinery |
| Responsible-owner eligibility | [Responsible owner](architecture/responsible-owner.md) | Old grant/role assumptions |
| Document Official management affordances / `allowed_actions` | [Document Official action precision](decisions/document-official-actions-read.md) + [Wire contract](architecture/wire-contract.md) | Frontend permission inference, temporary B03 candidate files |
| My Work governance row recognition / `WorkGovernanceItem.revision` | [My Work governance read precision](decisions/my-work-governance-identification-read.md) + [Wire contract](architecture/wire-contract.md) + [Frontend](architecture/frontend.md) | Legacy Approval DTOs/actions, per-row Governance Case enrichment |
| Document History / human-recognizable historical projection | [Document History recognition precision](decisions/document-history-recognition-read.md) + [Wire contract](architecture/wire-contract.md) | Audit reconstruction, browser cross-page history graph, temporary B07 work |
| Audit investigation query / recognition / Query Assist | [Audit investigation read](decisions/audit-investigation-read.md) + [API operation census](decisions/api-operation-census.md) | Temporary B09 work, generic search/reference-data machinery |
| Exact content / restore | [Content integrity](architecture/content-integrity.md) | Removed storage-provider contracts |
| Durable jobs / Search / Notification async boundary | [Async and search](architecture/async-and-search.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) | Old jobs registry |
| Technical clean-slate / reuse decision | [Technical baseline](architecture/technical-baseline.md) | Legacy source unless the reuse gate requires it |
| Backend topology | [Backend](architecture/backend.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) for the sixth semantic owner | Removed package tree |
| Internal owner contracts | [Interfaces](architecture/interfaces.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) for same-Scope/realtime additions | Foreign-SQL legacy evidence unless named |
| Persistence realization | [Persistence](architecture/persistence.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) for new state/constraints | Old migrations/schema |
| Transition / migration / cutover | [Transition & cutover](architecture/transition.md) + [T10 ratification](decisions/t10-ratification.md) | Temporary T10 review Evidence, old cutover plans, speculative migration tooling |
| Current decision dispositions | [Decision register](decisions/index.md) | Review chronology |
| Preserved / reopen / deferred obligations | [Forward obligations](decisions/forward-obligations.md) | Old decision-registry corpus |
| Repository reset / provenance | [Repository reset](decisions/repository-reset.md) | Unmerged branches unless exact provenance is required |
| Frontend LOCK/P8 evidence recovery for later P11 | [T11 B01-B09 LOCK Evidence Locator](decisions/t11-b01-b09-lock-evidence.md) + [T11 B10 LOCK Evidence Locator](decisions/t11-b10-lock-evidence.md) | Historical Git broadly; use only exact Evidence refs/blobs named by the locators |
| Documentation governance | [Documentation rules](development/documentation.md) | Product architecture |
| Repository-local engineering / CI / Git | [Engineering rules](development/engineering-rules.md) | Global methodology text duplicated locally |
| Executable application wire | [Wire contract](architecture/wire-contract.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) + [API census](decisions/api-operation-census.md) + [Document Official action precision](decisions/document-official-actions-read.md) when `getDocument.allowed_actions` is implicated + [My Work governance read precision](decisions/my-work-governance-identification-read.md) when `WorkGovernanceItem` is implicated + [Audit investigation read](decisions/audit-investigation-read.md) when op78/op87-op89 is implicated + [Access Assignment read](decisions/access-assignment-read.md) when op31 is implicated | Temporary work/review history, old OpenAPI |
| Frontend realization | [Frontend](architecture/frontend.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) + [Document Official action precision](decisions/document-official-actions-read.md) + [My Work governance read precision](decisions/my-work-governance-identification-read.md) + [Audit investigation read](decisions/audit-investigation-read.md) + [Access Assignment read](decisions/access-assignment-read.md) + [T8-F ratification](decisions/t8f-ratification.md) | Removed legacy frontend / speculative visual-framework choices |
| Frontend Product Experience planning / wireframing | Pinned methodology `ROUTER.md` → `METHOD.md` + `FRONTEND-METHOD.md`, then the current block's smallest Product/architecture owner pack | Removed local reusable method copy, unrelated Product artifacts, speculative design-system choices |
| Runtime / process / deployment | [Runtime](architecture/runtime.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) + [T8-G ratification](decisions/t8g-ratification.md) | Removed legacy runtime/deploy topology / speculative scale infrastructure |
| Golden Flows / validation proof baseline | [Validation baseline](architecture/validation-baseline.md) + [Discussion/Notifications authority](decisions/discussion-notifications-launch.md) + [T9 ratification](decisions/t9-ratification.md) | Temporary T9 review Evidence, implementation test code, transition mechanics |

## Reading law

Default fresh-actor route:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ pinned methodology ROUTER.md when applicable
→ selected methodology profile
→ 1–2 repository-local task authorities
```

Normal repository-local work stays at five files or fewer. Selected methodology files are a separate method profile. Exceed the local budget only for a named material reason. `docs/work/**`, research/Evidence/history, Git history, and removed implementation are never default-read.
