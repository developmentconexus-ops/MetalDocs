# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / ROUND-1 REVIEW + LEAD ADJUDICATION COMPLETE / ADJUDICATED CORRECTED CANDIDATE MATERIALIZED / BOUNDED FABLE ROUND 2 NEXT; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file is the **sole R10 current stage/status/next-action router**. Detailed target meaning lives only in durable authorities already promoted. T8-D artifacts remain non-authoritative staging.

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
10. active T8-D staging chain in §7
11. current schema/SQL/code only for a concrete T8-D evidence/reuse claim

Legacy implementation proves what exists, not what survives. Reviewer output and staging candidates are evidence/input only, never authority.

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

Ratified T8-A law:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Ratified T8-B law:

```text
ONE GO MODULE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

Ratified T8-C law:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL

concrete semantic-owner APIs
+ real consumer-owned mechanism/resolver ports
+ application-routed cross-owner facts
+ database/sql-family shared txscope
+ owner-authored same-tx Audit evidence
+ Authorization sole final ALLOW/default-DENY
+ named transaction-coupled durable intents
+ self-contained PII-free ReplaySnapshot
+ bounded owner facts + application read composition
-
shared/common semantic contracts
-
generic UnitOfWork/EventBus/policy language/ServiceLocator
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
  T8-D Persistence Realization                   ACTIVE / CORRECTED CANDIDATE / BOUNDED ROUND 2 NEXT
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

## 4. Durable T8-A/T8-B/T8-C closure

Durable authorities:

```text
wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md
wiki/architecture/r10-t8b-backend-module-package-topology.md
wiki/architecture/r10-t8c-internal-communication-contracts.md
```

Their registry amendments remain binding. T8-D may not redesign their owner topology, internal contract ownership, READ COMMITTED posture, Authorization authority, exact-content ownership or River-intent semantics by persistence convenience.

Key T8-C consequences consumed by T8-D:

```text
database/sql-family txscope
T2 READ COMMITTED posture
protected eligibility serialization
owner VersionToken / expected-version contract
same-Scope Audit append
Authorization Decide/DecideMany + AuthorizedScopes
owner-private facts/queries only; no foreign SQL
ManagedContent create-once + AdmissionClaims + two-phase GC
named OfficialRendition River intent shares semantic commit
idempotency same-key concurrency must not poison Scope
ReplaySnapshot = versioned / self-contained / PII-free / snapshot-only reconstruction
```

## 5. T8-D Round-1 review result

Round-1 independent artifact:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md`

Reviewed original candidate:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`

Review verdict:

```text
APPROVE T8-D GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
BLOCKER 2
MAJOR 11
LOW 10
Global Maximum class CONFIRMED
T8-C reopen NO
T8-B/T1→T7 reopen NO
stage trespass NO
```

Reviewer evidence is not authority.

## 6. Lead adjudication result

The Lead confronted Round-1 findings technically and selected bounded T8-D-local corrections.

```text
B1 River PG16/ownership                ACCEPT
B2 attach-vs-GC race                   ACCEPT
M1 frozen candidate FK                 ACCEPT
M2 immutable malware proof             ACCEPT
M3 provider-subject uniqueness         ACCEPT WITH NARROWING
M4 User lock ordering                  ACCEPT
M5 transaction census                  ACCEPT
M6 GovernanceAttempt uniqueness        ACCEPT
M7 singleton Company/company_id        REJECT
M8 Go<->DDL vocabulary parity          ACCEPT
M9 idempotency cleanup order           ACCEPT
M10 HMAC fingerprint                   ACCEPT
M11 DB trust classes/proof role        ACCEPT
```

The operator explicitly approved this adjudication before the corrected staging candidate was materialized.

No upstream reopen is proposed.

## 7. T8-D active corrected staging

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-bootstrap.md`

Original reviewed candidate:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`

Round-1 review evidence:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md`

**Active bounded-Round-2 corrected input:**

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md`

The corrected file is an explicit adjudicated overlay. For Round 2:

```text
effective corrected candidate
=
original candidate
+ adjudicated corrected overlay

where conflict exists, corrected overlay controls staging
```

The Global Maximum class remains:

```text
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+ DECLARATIVE CORRECTNESS
+ PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+ READ COMMITTED NARROW SERIALIZATION
+ EXPLICIT CAS
+ IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+ TRANSACTIONAL KEY↔REPLAY COMPLETION
+ THIRD-PARTY RIVER SCHEMA ISOLATION
+ PROOF-BACKED SELECTIVE LEGACY PROPERTY REUSE
- LEGACY PHYSICAL SHAPE INHERITANCE
- GENERIC PERSISTENCE FRAMEWORKS
- DUPLICATE CURRENT TRUTH
```

Material corrected delta includes:

```text
River river.* owner/provisioning/runtime grants + self-REINDEX OFF on PG16
semantic ManagedContent attachment FOR SHARE vs GC FOR UPDATE
GovernanceDecision composite FK to frozen candidate snapshot
immutable platform.malware_inspections evidence
current-only provider issuer+subject uniqueness
actor/target User one sorted lock class
complete transaction census
GovernanceAttempt subject uniqueness
M7 Lead rejection: Company remains current semantic scope; company_id retained
static Go <-> DDL closed-vocabulary parity proof
idempotency retention Replay->Key
HMAC-SHA-256 semantic fingerprint + key version
provisioner/owner/runtime/CI DB trust classes
accepted LOW constraint/proof precision corrections
```

## 8. Exact next action

```text
BOUNDED FABLE ROUND 2

review only the corrected material delta unless it creates a new contradiction

→ prove both Round-1 blockers are actually closed
→ attack accepted/narrowed corrections and M7 Lead rejection
→ report new/surviving BLOCKER / MAJOR / LOW
→ report Global Maximum still confirmed yes/no
→ report T8-C/T8-B/T1→T7 reopen yes/no
→ report T8-E/F/G/T10/T11 trespass yes/no
→ state whether another review round is materially required
→ final Lead adjudication
→ explicit operator ratification
→ only then promote one consolidated durable T8-D authority and open T8-E
```

No full Round-1 re-review is required unless the corrected delta exposes a material contradiction outside its bounded scope.

## 9. Stage boundaries

```text
T8-D = relational persistence / constraints / queries / locks
T8-E = exact executable OpenAPI/wire contract
T8-F = frontend realization
T8-G = runtime/process/deployment realization
T8-H = Whole-T8 coherence
T9   = Golden Flows + falsifiable Validation Baseline
T10  = current→target technical transition/cutover/rollback/deletion
T11  = implementation Execution Graph
T12  = adversarial implementation-readiness
```

T8-E remains closed until T8-D is independently converged, finally adjudicated, explicitly operator-ratified and promoted.

## 10. Final implementation gate

Implementation remains blocked until:

```text
T8→T12 CLOSED / OPERATOR-RATIFIED
→ Integrated Whole-R10 GCR PASS
→ fresh independent/cold review converged
→ operator explicitly authorizes implementation
```

Existing runtime safety controls remain binding until deliberately replaced by accepted equal-or-stronger target realization.
