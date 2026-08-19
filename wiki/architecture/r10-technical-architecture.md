# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T8-A CLOSED / OPERATOR-RATIFIED; T8-B ACTIVE / BACKEND MODULE & PACKAGE TOPOLOGY; T8-C→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the **sole R10 current stage/status/next-action router**. Detailed meaning lives in the durable authorities it routes to.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T8-A durable R10 authorities
6. Decision Registry + D4/T6/post-T6/T7/T8-A amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. this router
10. active T8-B staging listed in §5
11. current code/import/package evidence only for a concrete T8-B claim

Legacy implementation proves what exists, not what survives.

## 2. Binding Method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
unknown remains unknown
revalidation does not mean reinvention
prepare the seam, not dormant future capability
```

Program law:

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

Ratified T8-A physical-realization law:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

## 3. Current descent

```text
Product Contract REV001                          CLOSED / OPERATOR-APPROVED
Whole-Product GCR A1→A10                         CLOSED / OPERATOR-APPROVED
Launch ownership topology                        CLOSED / OPERATOR-APPROVED / 4+1
T1 — Semantic State & Invariants                 CLOSED / OPERATOR-RATIFIED
T2 — Governance / Effectivity / Tx               CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit                       CLOSED / OPERATOR-RATIFIED + D4
T4 — Exact Content / Storage / Restore           CLOSED / OPERATOR-RATIFIED
T5 — Durable Async / Search / Effects            CLOSED / OPERATOR-RATIFIED
T6 — Canonical API / Frontend Journeys           CLOSED / OPERATOR-RATIFIED / PROMOTED
T7 — Historical Migration Truth & Mapping        CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-A — Technical Authority & Legacy Census       CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                                CURRENT + D4 + T6 + post-T6 + T7 + T8-A amendments
TRRB                                             CLOSED / OPERATOR-RATIFIED / PROMOTED

T8 — Technical Realization Architecture          ACTIVE
  T8-B Backend Module & Package Topology         ACTIVE / TARGET TOPOLOGY DERIVATION NEXT
  T8-C Internal Communication Contracts          NOT OPEN
  T8-D Persistence Realization                   NOT OPEN
  T8-E Executable Wire Contract                  NOT OPEN
  T8-F Frontend Realization                      NOT OPEN
  T8-G Runtime / Process / Deployment            NOT OPEN
  T8-H Whole-T8 Global Coherence Review          NOT OPEN

T9 — Golden Flows & Validation Baseline          NOT OPEN
T10 — Transition / Refactor / Migration/Cutover  NOT OPEN
T11 — Implementation Program & Execution Graph   NOT OPEN
T12 — Adversarial Implementation-Readiness       NOT OPEN

implementation                                    BLOCKED
```

## 4. T8-A closure

Durable T8-A authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Registry reconciliation:

`wiki/architecture/rebaseline-decision-registry-t8a-amendment.md`

Binding consequences:

```text
current implementation = evidence only
legacy module/package topology = REWRITE / REHOME
current persistence/API/frontend/auth/storage shapes = no inheritance entitlement
non-Launch implementation = DELETE / DEFER absent named Launch consumer
PostgreSQL/River/contract-first/verifier/DB least-privilege properties = preserved where ratified/proven
selective reuse requires all five T8-A proofs
```

Completed T8-A staging is removed after promotion. Git history is provenance.

## 5. T8-B — ACTIVE

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-bootstrap.md`

T8-B answers:

> **What is the smallest backend package/module topology that gives each ratified semantic authority one clear home, keeps supporting mechanisms non-semantic, exposes only intentional public surfaces, and makes forbidden dependencies mechanically understandable?**

T8-B freezes only:

```text
target repository/package layout
semantic-owner realization boundaries
layering within owners
public/internal Go package surfaces
allowed dependency graph
forbidden dependency graph
composition root / dependency injection
location of shared mechanisms
```

Semantic ownership baseline:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence authority
```

Supporting mechanisms are not semantic owners merely because they have packages.

### Exact next action

```text
derive target backend responsibilities from the ratified 4+1 owners + T1→T8-A
→ identify cohesive vs isolated responsibilities
→ compare 2–3 materially distinct package-topology approaches
→ apply Global Maximum / essential-vs-accidental complexity
→ produce allowed/forbidden dependency candidate
→ adversarial challenge
→ T8-B platform-facing summary for operator ratification
```

Do **not** begin by mapping the legacy 15 modules one-for-one.

## 6. Stage boundaries

```text
T8-B = backend/package topology
T8-C = detailed inter-owner communication contracts
T8-D = persistence realization
T8-E = exact executable OpenAPI/wire contract
T8-F = frontend realization
T8-G = runtime/process/deployment realization
T8-H = Whole-T8 coherence
T9   = Golden Flows + falsifiable Validation Baseline
T10  = current→target technical transition/cutover/rollback/deletion
T11  = implementation Execution Graph
T12  = adversarial implementation-readiness
```

T8-B may identify that a seam is required to justify dependency direction; it must not invent the detailed T8-C contract by stealth.

## 7. Final implementation gate

Implementation remains blocked until:

```text
T8→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 GCR PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted equal-or-stronger target realization.