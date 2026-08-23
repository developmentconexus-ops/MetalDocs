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
| PRODUCT | Launch Product contract | CURRENT / APPROVED + BOUNDED T11 REOPEN | Controlled-document Launch scope plus stable-Document Discussion / explicit Mention / persistent in-app Notification Inbox is binding | `../product/contract.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| ALIGN | Whole-product alignment | CURRENT / APPROVED | Product-wide alignment baseline is binding except where the bounded T11 current authority explicitly supersedes pre-consumer Notification assumptions | `../product/alignment.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| OWN | Semantic ownership | CURRENT / APPROVED + BOUNDED T11 REOPEN | Authentication, Organization, Authorization, Controlled Documents + supporting Audit + supporting Notifications | `../architecture/ownership.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T1 | Domain state/invariants | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Stable semantic state/lifecycle vocabulary plus DocumentDiscussionMessage/Mention/Notification families | `../architecture/domain-model.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T2 | Governance/effectivity transactions | CURRENT / RATIFIED | Transaction, serialization, Release/effectivity laws | `../architecture/lifecycle.md` |
| T3 | Authorization/Audit | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Live grant/scope/domain authorization; `document.discuss`; canonical Discussion disclosure; bounded same-commit Audit | `../architecture/authorization-and-audit.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T3-D4 | Responsible-owner eligibility | REFINED / RATIFIED | New owner must be existing, same-Company, ENABLED User | `../architecture/responsible-owner.md` |
| T4 | Exact content / restore | CURRENT / RATIFIED | Exact-content descriptor truth, managed-content seam, fail-closed restore | `../architecture/content-integrity.md` |
| T5 | Async/Search/effects | CURRENT / RATIFIED + BOUNDED T11 REOPEN | River-backed named durable effects; canonical PostgreSQL Search baseline; persistent in-app Notification is local Product state; SSE wake-up is ephemeral mechanism | `../architecture/async-and-search.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T6 | Canonical journeys | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Product/API/frontend journey semantics including `/notifications` and stable-Document Discussion | `../product/journeys.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T6-API | `/api/v1` application census | REFINED / RATIFIED T11 BOUNDED REOPEN | Current census = **86** operations; 11 Idempotency-Key creations; operation 87+ needs a new lawful basis | [API census](api-operation-census.md) + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T7 | Historical migration truth | CURRENT / RATIFIED | No historical business migration required for Launch | `../architecture/transition.md` |
| T8-A | Clean-slate technical posture | CURRENT / RATIFIED | Legacy physical shape has no survival entitlement | `../architecture/technical-baseline.md` |
| T8-B | Backend topology | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Owner-first modular monolith with 4 business + 2 supporting semantic owners | `../architecture/backend.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T8-C | Internal contracts | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Authority-aligned owner/resolver model + same-Scope Discussion→Notification composition + narrow realtime port | `../architecture/interfaces.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T8-D | Persistence | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Owner-namespaced PostgreSQL relational core expanded for Discussion/Mention + `notifications.*` | `../architecture/persistence.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) |
| T8-E | Executable wire contract | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Base OAS 3.0.3 wire + exact eight-operation Discussion/Notifications extension = **86** application operations | `../architecture/wire-contract.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) + [API census](api-operation-census.md) |
| T8-E-FR | Document Official frontend read symmetry | REFINED / OPERATOR-RATIFIED PRECISION | Operation 47 exposes disclosure-safe current open-Revision and active-obsolescence routing references | `../architecture/wire-contract.md` + [precision record](frontend-read-symmetry.md) |
| T8-E-RO | Responsible-owner candidate read symmetry | REFINED / OPERATOR-APPROVED PRECISION | Operation 47 adds disclosure-safe complete D4-eligible responsible-owner candidates; must be absorbed before T11 merge-candidate cleanup | [precision record](responsible-owner-selection-read.md) + `../architecture/wire-contract.md` |
| T8-F | Frontend Realization | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Semantic/lens frontend realization now covers 86 operations and adds global Notifications utility route/feature without client authority duplication | `../architecture/frontend.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) + `t8f-ratification.md` |
| T8-G | Runtime / Process / Deployment | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | One modular-monolith runtime with PostgreSQL/River plus non-authoritative SSE/in-process Notification wake-up | `../architecture/runtime.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) + `t8g-ratification.md` |
| T8-H | Whole-T8 Global Coherence | REFINED BY T11 BOUNDED REOPEN | Prior T8 closure remains current for unchanged architecture; Discussion/Notifications delta independently converged under Lead GCR + Fable | [Discussion/Notifications authority](discussion-notifications-launch.md) + [T8-H ratification](t8h-ratification.md) |
| T9 | Golden Flows & Validation Baseline | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Existing validation baseline plus Discussion/Mention/Notification/disclosure/SSE proof obligations; current census = 86 | `../architecture/validation-baseline.md` + [Discussion/Notifications authority](discussion-notifications-launch.md) + [T9 ratification](t9-ratification.md) |
| T10 | Transition / Cutover | CURRENT / OPERATOR-RATIFIED | One-way greenfield activation with exactly five B0→B4 barriers; verified clean seal precedes first authoritative Product mutation | `../architecture/transition.md` + [T10 ratification](t10-ratification.md) |
| T11-COLLAB | Discussion / @Mention / Notifications bounded reopen | CURRENT / OPERATOR-RATIFIED / FABLE-CONVERGED | Stable-Document Discussion, semantic Mention, supporting Notifications owner, Inbox lifecycle, 11 routes, 86 operations, SSE invalidation and Lexical mechanism boundaries | [Discussion/Notifications authority](discussion-notifications-launch.md) |
| RESET | Repository clean-slate reset | CURRENT / OPERATOR-RATIFIED | Superseded implementation remains absent; required provenance stays reachable | `repository-reset.md` |
| REPO-STD-V1 | Repository operating envelope | CURRENT ORGANIZATIONAL BASELINE / ALIGNED | Repository Standard v1 alignment is merged; local repository controls preserve that operating envelope | `../roadmap.md` + `../development/documentation.md` + `../development/engineering-rules.md` |

## Forward obligations

The old decision corpus is intentionally not restored wholesale. Cross-stage obligations that still matter are preserved in `forward-obligations.md`:

```text
PRESERVE  21
REOPEN     3
DEFERRED  26
TOTAL     50
```

`ASY-02` was consumed by `T11-COLLAB` and is no longer DEFERRED. Email/push/preferences remain explicitly outside current Launch scope in the bounded current authority.

`CURRENT` semantics live in the owning authorities above. `SUPERSEDED` mechanics remain deleted unless new material evidence explicitly reopens them.

## Consumption law

Remaining stages:

```text
consume current owning authorities
+ apply bounded current decisions where they explicitly supersede older clauses
+ consume PRESERVE as proof-backed baseline
+ deliberately decide only stage-relevant REOPEN items
+ retain DEFERRED as future seams/counterexamples
- never inherit SUPERSEDED implementation by default
```

Reviewer findings and historical registry rows are Evidence, not independent requirement authority. A material contradiction reopens only the owning decision it actually implicates.