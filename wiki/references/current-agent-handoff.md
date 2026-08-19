# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T8-A CLOSED / OPERATOR-RATIFIED; T8-B ACTIVE / TARGET BACKEND TOPOLOGY DERIVATION NEXT; T8-C→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T8-A durable authorities
6. Decision Registry + D4/T6/post-T6/T7/T8-A amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. `wiki/architecture/r10-technical-architecture.md`
10. `docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-bootstrap.md`
11. current package/import/composition evidence only when a concrete T8-B claim needs it

Do not route target design through superseded/historical architecture or current module/package existence.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-A                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + D4 + T6 + post-T6 + T7 + T8-A amendments
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-B                                     ACTIVE / TARGET TOPOLOGY DERIVATION NEXT
T8-C                                     NOT OPEN
T8-D                                     NOT OPEN
T8-E                                     NOT OPEN
T8-F                                     NOT OPEN
T8-G                                     NOT OPEN
T8-H                                     NOT OPEN
T9                                       NOT OPEN
T10                                      NOT OPEN
T11                                      NOT OPEN
T12                                      NOT OPEN
implementation                           BLOCKED
```

## Binding execution law

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

## T7 — CLOSED

Ratified Launch decision:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs DB/content/history is DEV/test/throwaway and gives no business-data compatibility entitlement.

## T8-A — CLOSED

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8a-amendment.md`

Ratified Global Maximum:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Binding consequences:

```text
current implementation = evidence only
PRESERVE must be proved
sunk cost / tests / migration ease = no survival entitlement
legacy module/package topology = REWRITE / REHOME
current persistence/API/frontend/local-auth/provider-key shapes = no target inheritance
non-Launch implementation = DELETE / DEFER without named Launch consumer
```

Preserved properties/mechanisms include PostgreSQL product-state, River for named T5 jobs, contract-first generated Go/TS boundaries, verification registry/local-CI SSOT, runtime DB identity separated from DDL/schema ownership, deterministic DB bootstrap/proof and exact-content SHA/size/fail-closed principles.

Selective reuse requires all five T8-A proofs:

```text
named current R10 consumer
+ public contract free of legacy semantic authority
+ dependency direction fits target
+ proof asserts target property rather than legacy shape
+ reuse remains smaller than rewrite after transition cost
```

Completed T8-A staging has been removed from the live tree after promotion; Git history is provenance.

## T8-B — ACTIVE

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-bootstrap.md`

T8-B derives the target backend package/module topology from the accepted semantic ownership:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence authority
```

It freezes:

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

A semantic owner does not imply exactly one Go package. Supporting mechanisms do not become semantic owners merely because they have packages.

### Do not decide by stealth

```text
exact inter-owner contracts       → T8-C
exact persistence                 → T8-D
exact OpenAPI                     → T8-E
frontend realization             → T8-F
runtime/deployment               → T8-G
transition/deletion              → T10
```

### Exact next action

```text
derive target backend responsibilities from the 4+1 owners + T1→T8-A
→ identify cohesive vs isolated responsibilities
→ compare 2–3 materially distinct package-topology approaches
→ choose Global Maximum candidate
→ map allowed/forbidden dependencies
→ adversarial challenge
→ platform-facing summary
→ operator ratification
```

Do **not** start by mapping the legacy 15 modules one-for-one.

Implementation remains **BLOCKED**.