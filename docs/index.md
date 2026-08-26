---
id: documentation-index
kind: authority
owner: engineering
summary: Task/intention router to the smallest current MetalDocs authority pack.
---

# MetalDocs documentation

> **Role:** routing only. Mutable program status, allowed work, blockers and exact next action live only in [`roadmap.md`](roadmap.md). Product/architecture meaning lives in the named owners below.

## Start

```text
AGENTS.md
→ roadmap.md
→ applicable local method(s)
→ this index
→ 1–2 task owners by default
→ relevant section / operation first
→ expand only for a named material reason
```

Large authority files are not default whole-file reads. Search the exact concept, invariant, heading or `operationId` first; add global sections or additional owners when they can materially change the conclusion.

## Task routes

| Task / intention | Start with | Add when | Do not read by default |
|---|---|---|---|
| Program progression / implementation permission / next action | [Roadmap](roadmap.md) | Current PR/CI state when execution depends on it | Git history, closed PRs, Evidence chronology |
| Engineering reasoning / Global Maximum / material decision | [Engineering Method](development/engineering-method.md) | Semantic owner(s) of the decision | All architecture merely for ceremony |
| Repository / Git / context / documentation / PR continuity | [Repository Operating Method](development/repository-method.md) + [Engineering rules](development/engineering-rules.md) | Engineering Method when the repository-governance choice is material | Product architecture, old repository standards, branch graveyard |
| Frontend Product Experience planning / wireframing | [Engineering Method](development/engineering-method.md) + [Frontend Product Experience Planning Method](development/frontend-product-experience-planning-method.md) | Current block's Product/architecture owners through the rows below; reference research when real UX ambiguity exists | Whole wire/persistence/backend corpus; prior block Evidence unless inherited semantics are implicated |
| Product scope and concepts | [Product contract](product/contract.md) | [Discussion/Notifications](decisions/discussion-notifications-launch.md) when that bounded capability is implicated | Technical realization docs |
| Whole-product alignment / ownership | [Product alignment](product/alignment.md) + [Ownership](architecture/ownership.md) | Current bounded decision that explicitly refines ownership | Historical review artifacts |
| Stable-Document Discussion / @Mention / Notifications / Inbox | [Discussion/Notifications](decisions/discussion-notifications-launch.md) | Async/authorization/wire owner only for the exact mechanism/question implicated | Temporary T11 work, legacy frontend |
| Notification Full Inbox recognition / `listNotifications` projection | [Notification Inbox recognition precision](decisions/notification-inbox-recognition-read.md) + [Discussion/Notifications](decisions/discussion-notifications-launch.md) | Exact wire operation/global pagination law if questioned | Per-row browser fan-out designs, B08 historical Evidence |
| User journeys / application API meaning | [Product journeys](product/journeys.md) + [API operation census](decisions/api-operation-census.md) | Exact bounded decision authority for the operation/journey implicated | Removed OpenAPI/generated implementation |
| Domain state / invariants | [Domain model](architecture/domain-model.md) | Lifecycle/authorization only when the invariant crosses those owners | Runtime archaeology |
| Governance / effectivity / transactions | [Lifecycle](architecture/lifecycle.md) | Deadline/review bounded decision when implicated | Old Approval implementation |
| Authorization / Audit | [Authorization and audit](architecture/authorization-and-audit.md) | [Discussion/Notifications](decisions/discussion-notifications-launch.md) for `document.discuss`; [Audit investigation](decisions/audit-investigation-read.md) for B09 query/recognition | Provider roles, old ACL machinery, whole persistence model |
| Access / GroupMembership / RoleAssignment / Roles | [Access Assignment read](decisions/access-assignment-read.md) + [Authorization and audit](architecture/authorization-and-audit.md) | [API operation census](decisions/api-operation-census.md) for operation ownership; search only implicated ops/global law in [Wire contract](architecture/wire-contract.md); use [Product journeys](product/journeys.md) when the human job itself is questioned | Whole Wire/Persistence files, provider roles, old ACL machinery, frontend effective-access inference |
| Template configuration / Modelos lens / `listTemplateConfigurations` | [Template Configuration read](decisions/template-configuration-read.md) | [API operation census](decisions/api-operation-census.md) for operation ownership; search op40/41/43/50/51 or the global pagination law in [Wire contract](architecture/wire-contract.md) when exact encoding matters | Whole wire contract, Search-owner machinery, temporary B12 work |
| Responsible-owner eligibility | [Responsible owner](architecture/responsible-owner.md) | Exact wire projection when frontend selection is implicated | Old grant/role assumptions |
| Document Official management affordances / `allowed_actions` | [Document Official action precision](decisions/document-official-actions-read.md) | Search `getDocument` in [Wire contract](architecture/wire-contract.md) when exact HTTP/read shape matters | Whole wire contract, temporary B03 work |
| My Work governance row recognition | [My Work governance read precision](decisions/my-work-governance-identification-read.md) | Search the implicated `WorkGovernanceItem`/operation in [Wire contract](architecture/wire-contract.md); [Frontend](architecture/frontend.md) only for realization law | Legacy Approval DTOs, whole wire contract |
| Document History recognition | [Document History recognition precision](decisions/document-history-recognition-read.md) | Search the exact history operation/schema in [Wire contract](architecture/wire-contract.md) | Audit reconstruction, whole wire contract, B07 work history |
| Audit investigation query / Query Assist | [Audit investigation read](decisions/audit-investigation-read.md) + [API operation census](decisions/api-operation-census.md) | Search op78/op87–op89 in Wire only when exact encoding matters | Temporary B09 work, generic search machinery |
| Content format scope / upload / in-app editing boundary | [Content format vocabulary](decisions/content-format-vocabulary.md) | [Content integrity](architecture/content-integrity.md) for admission/exactness mechanism; search the implicated format law in [Wire contract](architecture/wire-contract.md) when exact encoding matters | Whole wire contract, converter mechanism selection, legacy format archaeology |
| Exact content / restore | [Content integrity](architecture/content-integrity.md) | Exact Wire/runtime mechanism only when the claim reaches it | Removed storage-provider contracts |
| Durable jobs / Search / Notification async boundary | [Async and search](architecture/async-and-search.md) | [Discussion/Notifications](decisions/discussion-notifications-launch.md) for Notification-specific behavior | Old jobs registry |
| Technical clean-slate / reuse decision | [Technical baseline](architecture/technical-baseline.md) | [Repository reset](decisions/repository-reset.md) when provenance/reachability matters | Legacy source until the reuse gate names it |
| Backend topology | [Backend](architecture/backend.md) | Bounded owner decision when topology meaning was explicitly refined | Removed package tree |
| Internal owner contracts | [Interfaces](architecture/interfaces.md) | Search exact producer/consumer contract first; add bounded decision only when it refines that contract | Whole persistence/runtime corpus |
| Persistence realization | [Persistence](architecture/persistence.md) | Search exact entity/constraint first; add lifecycle/decision owner when semantics are questioned | Old migrations/schema |
| Transition / migration / cutover | [Transition](architecture/transition.md) + [T10 ratification](decisions/t10-ratification.md) | Repository reset only for source-tree provenance | Temporary T10 review Evidence, speculative tooling |
| Executable application wire | [API operation census](decisions/api-operation-census.md) + search the exact operation/schema in [Wire contract](architecture/wire-contract.md) | Read the relevant global wire law (pagination/idempotency/ETag/etc.) and exact bounded decision when implicated | Whole 90KB wire file, old OpenAPI/generated code |
| Frontend realization architecture | [Frontend](architecture/frontend.md) + [T8-F ratification](decisions/t8f-ratification.md) | Exact Product/decision owner for the surface being realized | Removed legacy frontend, visual-framework speculation |
| Runtime / process / deployment | [Runtime](architecture/runtime.md) + [T8-G ratification](decisions/t8g-ratification.md) | Exact backend/async owner when a process boundary is questioned | Removed legacy runtime/deploy tree |
| Golden Flows / validation proof baseline | [Validation baseline](architecture/validation-baseline.md) + [T9 ratification](decisions/t9-ratification.md) | Exact bounded decision whose new behavior needs proof | Temporary review Evidence, implementation test code |
| Current decision dispositions | [Decision register](decisions/index.md) | Open only the owning decision named by the relevant row | Review chronology |
| Preserved / reopen / deferred obligations | [Forward obligations](decisions/forward-obligations.md) | Owning semantic authority when closing/refining an obligation | Old decision-registry corpus |
| Repository reset / unmerged provenance | [Repository reset](decisions/repository-reset.md) | Exact archive/Evidence ref only when byte-level provenance is required | Unmerged branches broadly |
| Frontend LOCK/P8 Evidence recovery for later P11 | [B01–B09 locator](decisions/t11-b01-b09-lock-evidence.md) + [B10 locator](decisions/t11-b10-lock-evidence.md) + [B11 locator](decisions/t11-b11-lock-evidence.md) + [B12 locator](decisions/t11-b12-lock-evidence.md) + [P11 acceptance locator](decisions/t11-p11-acceptance-evidence.md) | Exact ref/blob named by the locator | Historical Git broadly, intermediate wireframe candidates |

## Navigation law

The table is an **entry map, not a fence**.

```text
start with the smallest credible authority pack
→ test the current claim
→ if another owner/source can materially falsify it, expand
→ otherwise stop
```

Do not read irrelevant material to satisfy ceremony. Do not omit material context merely to preserve a small context window.

Current repository authority owns Product meaning. The local methods govern reasoning/operation/frontend planning. Evidence and research may challenge accepted authority through those methods; they do not silently replace it.
