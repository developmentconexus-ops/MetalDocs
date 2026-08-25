---
id: decision-register
kind: authority
owner: architecture
summary: Compact current decision/disposition registry routing to the owning MetalDocs authorities.
---

# MetalDocs decision register

This file is a **router of current dispositions**, not a historical ledger. The named authority owns the full rationale, semantics, proof and reopen triggers.

| ID | Subject | Disposition | Current outcome | Authority |
|---|---|---|---|---|
| PRODUCT | Launch Product contract | CURRENT / APPROVED + BOUNDED T11 REOPEN | Controlled-document Launch plus accepted Discussion/Mention/Notification and governance-deadline additions | `../product/contract.md` + `discussion-notifications-launch.md` + `governance-step-deadline.md` |
| ALIGN | Whole-product alignment | CURRENT / APPROVED | Product alignment remains binding except explicit bounded refinements | `../product/alignment.md` |
| OWN | Semantic ownership | CURRENT / APPROVED + BOUNDED T11 REOPEN | Authentication, Organization, Authorization, Controlled Documents + supporting Audit/Notifications | `../architecture/ownership.md` + current bounded decisions |
| T1 | Domain state / invariants | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Current semantic families include accepted Discussion/Mention/Notification; deadlines add no lifecycle state | `../architecture/domain-model.md` + current bounded decisions |
| T2 | Governance / effectivity / transactions | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Sequential governance with accepted deadline snapshot/`due_at` precision | `../architecture/lifecycle.md` + `governance-step-deadline.md` |
| T3 | Authorization / Audit | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Live grant/scope/domain authorization, `document.discuss`, bounded same-commit Audit | `../architecture/authorization-and-audit.md` + `discussion-notifications-launch.md` |
| T3-D4 | Responsible-owner eligibility | REFINED / RATIFIED | New owner = existing same-Company ENABLED User | `../architecture/responsible-owner.md` |
| T4 | Exact content / restore | CURRENT / RATIFIED | Exact-content descriptor truth, managed-content seam, fail-closed restore | `../architecture/content-integrity.md` |
| T5 | Async / Search / effects | CURRENT / RATIFIED + BOUNDED T11 REOPEN | River named effects, PostgreSQL Search, persistent in-app Notification; deadline breach alone has no worker/effect | `../architecture/async-and-search.md` + current bounded decisions |
| T6 | Canonical journeys | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Accepted journeys include Notifications, governance recognition/deadline precision and structured Audit investigation | `../product/journeys.md` + current bounded decisions |
| T6-API | `/api/v1` census | REFINED / RATIFIED T11 BOUNDED REOPEN | **89** application operations; **11** Idempotency-Key creations; operation 90+ needs new lawful basis | `api-operation-census.md` |
| T7 | Historical migration truth | CURRENT / RATIFIED | No historical business migration required for Launch | `../architecture/transition.md` |
| T8-A | Clean-slate technical posture | CURRENT / RATIFIED | Legacy physical shape has no survival entitlement | `../architecture/technical-baseline.md` |
| T8-B | Backend topology | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Owner-first modular monolith; 4 business + 2 supporting semantic owners | `../architecture/backend.md` + current bounded decisions |
| T8-C | Internal contracts | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Authority-aligned owner/resolver contracts with bounded Notifications/Audit refinements | `../architecture/interfaces.md` + current bounded decisions |
| T8-D | Persistence | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Owner-namespaced PostgreSQL core plus accepted deadline state | `../architecture/persistence.md` + current bounded decisions |
| T8-E | Executable wire | CURRENT / RATIFIED + BOUNDED T11 REOPEN | Current 89-operation wire plus accepted frontend/read/Audit precisions | `../architecture/wire-contract.md` + `api-operation-census.md` + current bounded decisions |
| T8-E-FR | Document Official read symmetry | REFINED / OPERATOR-RATIFIED | op47 exposes disclosure-safe current routing references | `frontend-read-symmetry.md` + `../architecture/wire-contract.md` |
| T8-E-RO | Responsible-owner candidate read | REFINED / OPERATOR-APPROVED | op47 exposes complete D4-eligible owner candidates | `responsible-owner-selection-read.md` + `../architecture/wire-contract.md` |
| T8-E-DOA | Document Official action hints | REFINED / OPERATOR-RATIFIED | op47 `allowed_actions` guides UX; commands recheck truth | `document-official-actions-read.md` + `../architecture/wire-contract.md` |
| T8-E-WG | My Work governance projection | REFINED / OPERATOR-RATIFIED | Governed revision identity + bounded deadline/filter/order precision | `my-work-governance-identification-read.md` + `governance-step-deadline.md` |
| T8-E-GD | Governance Step deadline | REFINED / OPERATOR-RATIFIED | `due_in_days` frozen per attempt; Step activation freezes `due_at`; no automatic breach effect | `governance-step-deadline.md` |
| T8-E-GC | Governance Case deadline projection | REFINED / OPERATOR-RATIFIED | op67 projects exact active/decided deadline truth without queue inference | `governance-case-step-deadline-read.md` |
| T8-F | Frontend realization | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Semantic/lens realization for 89 operations; no frontend AuthZ authority or incomplete-page post-filter | `../architecture/frontend.md` + current bounded decisions + `t8f-ratification.md` |
| T8-G | Runtime / process / deployment | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Modular-monolith runtime with PostgreSQL/River + non-authoritative SSE wake-up | `../architecture/runtime.md` + `t8g-ratification.md` |
| T8-H | Whole-T8 coherence | REFINED BY T11 BOUNDED REOPEN | Prior closure remains for unchanged architecture; bounded T11 deltas require bounded downstream proof | `t8h-ratification.md` + current bounded decisions |
| T9 | Golden Flows / validation | CURRENT / OPERATOR-RATIFIED + BOUNDED T11 REOPEN | Existing validation baseline plus proof obligations for current bounded T11 changes | `../architecture/validation-baseline.md` + `t9-ratification.md` + current bounded decisions |
| T10 | Transition / cutover | CURRENT / OPERATOR-RATIFIED | One-way greenfield activation with five B0→B4 barriers | `../architecture/transition.md` + `t10-ratification.md` |
| T11-COLLAB | Discussion / Mention / Notifications | CURRENT / OPERATOR-RATIFIED / FABLE-CONVERGED | Stable-Document Discussion, Mention, Notifications owner, Inbox, ops79–86 and SSE invalidation | `discussion-notifications-launch.md` |
| T11-AUDIT | Audit investigation | CURRENT / OPERATOR-RATIFIED | op78 structured investigation + purpose-built op87–op89 Query Assist; Audit remains evidence | `audit-investigation-read.md` + `api-operation-census.md` |
| T11-ACCESS | B11 Access Assignment read precision | CURRENT / OPERATOR-RATIFIED | op31 remains the RoleAssignment traversal and gains exact server-side User/Group/Scope/Role filters plus human-recognizable read references; no operation 90 or effective-access engine | `access-assignment-read.md` + `api-operation-census.md` |
| T11-GOV-REVIEW | Governance Review Layer seam | CURRENT / OPERATOR-RATIFIED / FUTURE-SEAM ONLY | Future selected-range review binds to immutable reviewed snapshot; current Launch unchanged | `governance-review-layer-seam.md` + `forward-obligations.md` |
| RESET | Repository clean-slate reset | CURRENT / OPERATOR-RATIFIED | Superseded implementation remains absent; required unmerged provenance remains reachable | `repository-reset.md` |
| REPO-OPS | Repository operating model | CURRENT / OPERATOR-APPROVED | Local Engineering + Repository + Frontend methods; `AGENTS` bootstrap, selective `docs/index`, roadmap snapshot, Git/Evidence continuity | `../development/repository-method.md` + `../development/engineering-rules.md` |

## Consumption law

```text
start from current owning authority
+ apply only bounded decisions that explicitly refine it
+ consume PRESERVE obligations as proof-backed baseline
+ deliberately adjudicate stage-relevant REOPEN items
+ retain DEFERRED items only as future seams/counterexamples
- never inherit SUPERSEDED implementation/mechanics by default
```

Cross-stage obligations that still matter are registered in `forward-obligations.md`.

Reviewer findings and historical rows are Evidence, not requirement authority. A material contradiction reopens only the owner it actually implicates.
