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
| PRODUCT | Launch Product contract | CURRENT / APPROVED + BOUNDED T11 REOPEN | Controlled-document Launch scope plus stable-Document Discussion / Mention / persistent in-app Notification Inbox; governance Step deadline is a bounded T11 temporal addition | `../product/contract.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Governance Step deadline](governance-step-deadline.md) |
| ALIGN | Whole-product alignment | CURRENT / APPROVED | Product-wide alignment baseline is binding except where bounded current T11 authorities explicitly refine it | `../product/alignment.md` + current bounded decisions |
| OWN | Semantic ownership | CURRENT / APPROVED + BOUNDED T11 REOPEN | Authentication, Organization, Authorization, Controlled Documents + supporting Audit + supporting Notifications; deadline remains Controlled Documents truth | `../architecture/ownership.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Governance Step deadline](governance-step-deadline.md) |
| T1 | Domain state/invariants | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Stable semantic families plus Discussion/Mention/Notification; no new lifecycle state for deadlines | `../architecture/domain-model.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Governance Step deadline](governance-step-deadline.md) |
| T2 | Governance/effectivity transactions | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Existing sequential GovernanceAttempt law plus optional per-Step elapsed-day deadline snapshot and activation-time frozen `due_at`; breach has no automatic lifecycle effect | `../architecture/lifecycle.md` + [Governance Step deadline](governance-step-deadline.md) |
| T3 | Authorization/Audit | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Live grant/scope/domain authorization; `document.discuss`; canonical Discussion disclosure; bounded same-commit Audit | `../architecture/authorization-and-audit.md` + [Discussion/Notifications](discussion-notifications-launch.md) |
| T3-D4 | Responsible-owner eligibility | REFINED / RATIFIED | New owner must be existing, same-Company, ENABLED User | `../architecture/responsible-owner.md` |
| T4 | Exact content / restore | CURRENT / RATIFIED | Exact-content descriptor truth, managed-content seam, fail-closed restore | `../architecture/content-integrity.md` |
| T5 | Async/Search/effects | CURRENT / RATIFIED + BOUNDED T11 REOPEN | River-backed named durable effects; PostgreSQL Search; persistent in-app Notification is local Product state; deadline breach alone creates no worker/effect | `../architecture/async-and-search.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Governance Step deadline](governance-step-deadline.md) |
| T6 | Canonical journeys | CURRENT / RATIFIED + BOUNDED T11 REOPEN | `/notifications`, stable-Document Discussion, human-recognizable governance work, optional active-Step deadline and deadline-first My Work triage; Governance Case re-reads exact persisted Step deadline truth independently of My Work navigation state | `../product/journeys.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [My Work read precision](my-work-governance-identification-read.md) + [Governance Step deadline](governance-step-deadline.md) + [Governance Case deadline read](governance-case-step-deadline-read.md) |
| T6-API | `/api/v1` application census | REFINED / RATIFIED T11 BOUNDED REOPEN | Current census = **86** operations; 11 Idempotency-Key creations; operation 87+ needs a new lawful basis | [API census](api-operation-census.md) + [Discussion/Notifications](discussion-notifications-launch.md) |
| T7 | Historical migration truth | CURRENT / RATIFIED | No historical business migration required for Launch | `../architecture/transition.md` |
| T8-A | Clean-slate technical posture | CURRENT / RATIFIED | Legacy physical shape has no survival entitlement | `../architecture/technical-baseline.md` |
| T8-B | Backend topology | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Owner-first modular monolith with 4 business + 2 supporting semantic owners | `../architecture/backend.md` + [Discussion/Notifications](discussion-notifications-launch.md) |
| T8-C | Internal contracts | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Authority-aligned owner/resolver model + same-Scope Discussion→Notification composition; deadline remains inside Controlled Documents contracts | `../architecture/interfaces.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Governance Step deadline](governance-step-deadline.md) |
| T8-D | Persistence | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Owner-namespaced PostgreSQL relational core plus deadline config/snapshot/frozen `due_at` on existing governance Step structures | `../architecture/persistence.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Governance Step deadline](governance-step-deadline.md) |
| T8-E | Executable wire contract | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Current 86-operation wire; `GovernanceRouteStep.due_in_days?`; `WorkGovernanceItem.due_at?`; Governance Case Step projection carries exact persisted `due_at?` after activation; `listGovernanceWork` deadline-first order + bounded `deadline_filter` + cursor-stable temporal anchor | `../architecture/wire-contract.md` + [API census](api-operation-census.md) + [My Work read precision](my-work-governance-identification-read.md) + [Governance Step deadline](governance-step-deadline.md) + [Governance Case deadline read](governance-case-step-deadline-read.md) |
| T8-E-FR | Document Official frontend read symmetry | REFINED / OPERATOR-RATIFIED PRECISION | Operation 47 exposes disclosure-safe current open-Revision and active-obsolescence routing references | `../architecture/wire-contract.md` + [precision](frontend-read-symmetry.md) |
| T8-E-RO | Responsible-owner candidate read symmetry | REFINED / OPERATOR-APPROVED PRECISION | Operation 47 adds disclosure-safe complete D4-eligible responsible-owner candidates | [precision](responsible-owner-selection-read.md) + `../architecture/wire-contract.md` |
| T8-E-DOA | Document Official action hints | REFINED / OPERATOR-RATIFIED PRECISION | Operation 47 returns required `allowed_actions` UX guidance; commands always recheck truth | [precision](document-official-actions-read.md) + `../architecture/wire-contract.md` |
| T8-E-WG | My Work governance read projection | REFINED / OPERATOR-RATIFIED PRECISION | Exact governed `RevisionReference` + optional active-Step `due_at`; default `due_at ASC NULLS LAST, document.code, governance_attempt_id`; four bounded deadline presets; relative-filter cursor anchor is server-owned | [My Work read precision](my-work-governance-identification-read.md) + [Governance Step deadline](governance-step-deadline.md) + `../architecture/wire-contract.md` |
| T8-E-GD | Governance Step deadline | REFINED / OPERATOR-RATIFIED PRECISION | Optional `due_in_days` is frozen per attempt; Step activation freezes persisted `due_at`; 1 day = 24 elapsed hours; breach has no automatic lifecycle/effect | [Governance Step deadline](governance-step-deadline.md) + `../architecture/lifecycle.md` + `../architecture/persistence.md` + `../architecture/wire-contract.md` |
| T8-E-GC | Governance Case Step deadline projection | REFINED / OPERATOR-RATIFIED PRECISION | Operation 67 remains sufficient; pending Step forbids `due_at`; active/decided timed Step exposes exact persisted `due_at`; untimed Step omits it; B06 never inherits queue deadline as authority | [Governance Case deadline read](governance-case-step-deadline-read.md) + [Governance Step deadline](governance-step-deadline.md) + `../architecture/wire-contract.md` |
| T8-F | Frontend Realization | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Semantic/lens realization covers 86 operations; My Work renders exact/relative deadline truth and four bounded governance filters; Governance Case independently re-reads exact Step deadline truth; no manual priority, generic filter DSL, client global re-sort or frontend Authorization matrix | `../architecture/frontend.md` + [Discussion/Notifications](discussion-notifications-launch.md) + [Document Official action precision](document-official-actions-read.md) + [My Work read precision](my-work-governance-identification-read.md) + [Governance Step deadline](governance-step-deadline.md) + [Governance Case deadline read](governance-case-step-deadline-read.md) + `t8f-ratification.md` |
| T8-G | Runtime / Process / Deployment | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | One modular-monolith runtime with PostgreSQL/River plus non-authoritative SSE wake-up; no deadline worker baseline | `../architecture/runtime.md` + current bounded decisions + `t8g-ratification.md` |
| T8-H | Whole-T8 Global Coherence | REFINED BY T11 BOUNDED REOPEN | Prior T8 closure remains current for unchanged architecture; bounded T11 deltas require later whole-program coherence before implementation authorization | current bounded decisions + [T8-H ratification](t8h-ratification.md) |
| T9 | Golden Flows & Validation Baseline | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Existing validation baseline plus current bounded T11 proof obligations; census remains 86 | `../architecture/validation-baseline.md` + current bounded decisions + [T9 ratification](t9-ratification.md) |
| T10 | Transition / Cutover | CURRENT / OPERATOR-RATIFIED | One-way greenfield activation with exactly five B0→B4 barriers | `../architecture/transition.md` + [T10 ratification](t10-ratification.md) |
| T11-COLLAB | Discussion / Mention / Notifications bounded reopen | CURRENT / OPERATOR-RATIFIED / FABLE-CONVERGED | Stable-Document Discussion, semantic Mention, Notifications owner, Inbox lifecycle, 11 routes, 86 operations and SSE invalidation | [Discussion/Notifications](discussion-notifications-launch.md) |
| RESET | Repository clean-slate reset | CURRENT / OPERATOR-RATIFIED | Superseded implementation remains absent; required provenance stays reachable | `repository-reset.md` |
| REPO-STD-V1 | Repository operating envelope | CURRENT ORGANIZATIONAL BASELINE / ALIGNED | Repository Standard v1 alignment remains binding | `../roadmap.md` + `../development/documentation.md` + `../development/engineering-rules.md` |

## Forward obligations

The old decision corpus is intentionally not restored wholesale. Cross-stage obligations that still matter are preserved in `forward-obligations.md`.

`CURRENT` semantics live in the owning authorities above. `SUPERSEDED` mechanics remain deleted unless material evidence explicitly reopens them.

## Consumption law

```text
consume current owning authorities
+ apply bounded current decisions where they explicitly supersede older clauses
+ consume PRESERVE as proof-backed baseline
+ deliberately decide only stage-relevant REOPEN items
+ retain DEFERRED as future seams/counterexamples
- never inherit SUPERSEDED implementation by default
```

Reviewer findings and historical registry rows are Evidence, not requirement authority. A material contradiction reopens only the owning decision it actually implicates.
