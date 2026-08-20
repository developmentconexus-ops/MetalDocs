# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / ROUND-1 + ROUND-2 REVIEW COMPLETE / FINAL LEAD ADJUDICATION MATERIALIZED / FINAL OPERATOR RATIFICATION NEXT; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-20  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the **sole R10 current stage/status/next-action router**. Detailed target meaning lives only in durable authorities already promoted. T8-D artifacts remain non-authoritative staging until operator ratification and promotion.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T8-C durable R10 authorities
6. Decision Registry + amendments through T8-C
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. this router
10. active T8-D staging chain in §6
11. current schema/SQL/code only for concrete T8-A reuse/feasibility evidence

Legacy implementation proves what exists, not what survives. Reviewer output and staging candidates are evidence/input only, never authority.

## 2. Binding Method / realization laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
revalidation != reinvention
prepare the seam, not dormant future capability
```

Binding program law:

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

Already-ratified realization baseline:

```text
T8-A:
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE

T8-B:
ONE GO MODULE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH

T8-C:
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
+ database/sql-family shared txscope
+ application-routed cross-owner facts
+ owner-authored same-tx Audit
+ Authorization sole final ALLOW/default-DENY
+ named transaction-coupled durable intents
+ self-contained PII-free ReplaySnapshot
- shared semantic contracts / generic UnitOfWork / EventBus / policy language
```

T2 PostgreSQL posture remains:

```text
READ COMMITTED
+ narrow explicit serialization
+ OCC/CAS
+ structural constraints where required
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
T8-C — Internal Communication Contracts          CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                                CURRENT + amendments through T8-C
TRRB                                             CLOSED / OPERATOR-RATIFIED / PROMOTED

T8 — Technical Realization Architecture          ACTIVE
  T8-D Persistence Realization                   ACTIVE / FINAL RATIFICATION NEXT
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

## 4. T8-D Global Maximum class

Independently confirmed in both review passes:

```text
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+
DECLARATIVE CORRECTNESS
+
PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+
READ COMMITTED NARROW SERIALIZATION
+
EXPLICIT CAS
+
IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+
TRANSACTIONAL KEY↔REPLAY COMPLETION
+
THIRD-PARTY RIVER SCHEMA ISOLATION
+
PROOF-BACKED SELECTIVE LEGACY PROPERTY REUSE
-
LEGACY PHYSICAL SHAPE INHERITANCE
-
GENERIC PERSISTENCE FRAMEWORKS
-
DUPLICATE CURRENT TRUTH
```

## 5. Independent-review convergence

Round 1:

```text
APPROVE T8-D GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
BLOCKER 2 / MAJOR 11 / LOW 10
Global Maximum CONFIRMED
T1→T7/T8-B/T8-C reopen NO
stage trespass NO
```

After Lead adjudication and operator-approved corrected-candidate materialization, bounded Round 2 found:

```text
APPROVE CORRECTED T8-D DELTA WITH MATERIAL FIXES
BLOCKER 0 / MAJOR 7 / LOW 6
BOTH Round-1 blockers CLOSED
Global Maximum CONFIRMED
upstream reopen NO
stage trespass NO
third review round NOT REQUIRED
final Lead adjudication MAY PROCEED
```

Final Lead adjudication then closes all remaining Round-2 findings in staging:

```text
7 / 7 MAJOR closed
6 / 6 LOW closed
surviving material contradiction 0
Global Maximum CONFIRMED
upstream reopen NO
stage trespass NO
third Fable round NOT MATERIAL
```

## 6. Active T8-D staging chain

```text
original candidate
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md

Round-1 independent review
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md

adjudicated corrected candidate
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md

bounded Round-2 delta review
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-corrected-candidate-fable-delta-review.md

FINAL LEAD ADJUDICATION — ACTIVE RATIFICATION INPUT
  docs/superpowers/analysis/2026-08-20-r10-t8d-persistence-realization-final-lead-adjudication.md
```

For final operator ratification:

```text
effective final staging candidate
=
original candidate
+ adjudicated corrected overlay
+ final Lead adjudication overlay
```

Later staging controls where text conflicts. Reviewer artifacts remain evidence only.

Key final adjudicated corrections include:

```text
immutable managed_content_descriptors + immutable malware evidence
universal semantic attach ManagedContent FOR SHARE vs GC FOR UPDATE
GC downstream proofs are non-locking after its root
blanket protected actor FOR SHARE + deterministic User lock ordering
governance decider FK to frozen candidate + candidate materialization on activation
AdmissionClaim reserved at claim-bound OPEN allocation
Area lifecycle serialization against create
Company singleton fail-closed no-isolation interlock; semantic company_id retained
paired Idempotency Key↔Replay + Replay→Key cleanup order
HMAC fingerprint + drain-before-rotation derivation law
River river.* third-party isolation + self-REINDEX OFF on PG16
provisioner / owner / runtime / verifier DB trust classes + runtime/verifier grant parity
backup-pin persistence/locking family
Go/static vocabulary ↔ DDL CHECK equality; execution routed to T9 validation baseline
```

## 7. Exact next action

```text
EXPLICIT OPERATOR RATIFICATION OF FINAL T8-D EFFECTIVE STAGING CANDIDATE
```

If and only if explicitly ratified, normal repository governance may:

```text
1. consolidate one durable T8-D authority under wiki/architecture/
2. append the T8-D Decision Registry amendment
3. tombstone/retire live T8-D staging artifacts while preserving Git history
4. mark T8-D CLOSED / OPERATOR-RATIFIED / PROMOTED
5. open T8-E Executable Wire Contract
```

No product implementation is authorized by T8-D closure.

## 8. Stage boundaries

```text
T8-D = relational persistence / constraints / queries / locks
T8-E = exact executable OpenAPI/wire contract
T8-F = frontend realization
T8-G = runtime/process/deployment realization
T8-H = Whole-T8 coherence
T9   = Golden Flows + falsifiable Validation Baseline
T10  = current→target transition/cutover/rollback/deletion
T11  = implementation Execution Graph
T12  = adversarial implementation-readiness
```

Until explicit ratification/promotion:

```text
T8-D ACTIVE
T8-E→T12 NOT OPEN
implementation BLOCKED
```
