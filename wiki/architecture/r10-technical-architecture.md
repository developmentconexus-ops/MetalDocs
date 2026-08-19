# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T8-B CLOSED / OPERATOR-RATIFIED; T8-C ACTIVE / INTERNAL COMMUNICATION CONTRACTS; T8-D→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
5. T1→T8-B durable R10 authorities
6. Decision Registry + D4/T6/post-T6/T7/T8-A/T8-B amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. this router
10. active T8-C staging listed in §6
11. current code/interfaces only for a concrete T8-C evidence claim

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

Ratified T8-B topology law:

```text
ONE GO MODULE FOR BACKEND GO CODE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
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
T8-B — Backend Module & Package Topology         CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                                CURRENT + amendments through T8-B
TRRB                                             CLOSED / OPERATOR-RATIFIED / PROMOTED

T8 — Technical Realization Architecture          ACTIVE
  T8-C Internal Communication Contracts          ACTIVE / CONTRACT DERIVATION NEXT
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

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8a-amendment.md`

Binding consequences remain:

```text
current implementation = evidence only
legacy module/package topology = REWRITE / REHOME
current persistence/API/frontend/auth/storage shapes = no inheritance entitlement
non-Launch implementation = DELETE / DEFER absent named Launch consumer
selective reuse requires all five T8-A proofs
```

## 5. T8-B closure

Durable authority:

`wiki/architecture/r10-t8b-backend-module-package-topology.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8b-amendment.md`

Binding consequences:

```text
semantic owner public roots = authentication / organization / authorization /
                              controlleddocs / audit
one importable public surface per owner
owner-private decomposition ungated
transport → application = only semantic inbound door
application leaves = stateless choreography
owner → owner direct imports forbidden
platform = mechanism only
composition = wiring only
transaction/Audit/AuthZ seam classes named, exact contracts deferred T8-C
first-party package classification = closed-world
first-party dependency edges = default-deny
```

T8-B staging/reviewer artifacts are completed provenance and are removed from the live staging tree after promotion; Git history preserves them.

## 6. T8-C — ACTIVE

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-bootstrap.md`

T8-C answers:

> **What is the smallest complete set of internal contracts that lets the ratified owners and non-semantic application/mechanism layers realize T1→T8-B semantics without direct owner imports, foreign SQL, duplicate authority, hidden write ownership or unnecessary interface ceremony?**

T8-C freezes:

```text
owner queries
owner capabilities
read projections
same-process/local calls
transaction-coupled intents
River/durable job seams where already justified
consumer/producer contract ownership
```

T8-C must resolve at minimum the exact contracts for the T8-B seam classes:

```text
provider-neutral transaction participation
same-transaction owner evidence → Audit handoff
owner-authored domain predicate facts → Authorization decision
transaction-coupled durable intent
material consumer-owned mechanism ports
```

### Exact next action

```text
derive complete interaction census from T2/T3/T4/T5/T6 + T8-B
→ classify sync query / mutation / same-tx / durable-intent / mechanism / read-projection
→ freeze contract ownership and direction
→ resolve txscope / Audit evidence / Authorization predicate contracts first
→ derive remaining owner/mechanism contracts
→ compare credible contract-placement alternatives
→ apply Method + subtractive pass
→ adversarial challenge
→ operator-ratifiable T8-C candidate
```

Current implementation interfaces may be inspected as evidence only and receive no survival entitlement from existence.

## 7. Stage boundaries

```text
T8-C = internal communication contracts
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

T8-C may identify persistence needs but must not design schema/locks by stealth. It may not reopen T8-B topology without a concrete contract contradiction.

## 8. Final implementation gate

Implementation remains blocked until:

```text
T8→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 GCR PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted equal-or-stronger target realization.