# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / ROUND-1 REVIEW + LEAD ADJUDICATION COMPLETE / ADJUDICATED CORRECTED CANDIDATE MATERIALIZED / BOUNDED FABLE ROUND 2 NEXT; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/r10-technical-architecture.md` — sole status/next-action router
5. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
6. T1→T8-C durable authorities
7. Decision Registry + amendments through T8-C
8. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
9. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
10. `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-bootstrap.md`
11. original T8-D candidate
12. Round-1 independent Fable review
13. adjudicated corrected T8-D candidate — active Round-2 input
14. current schema/migrations/SQL/code only when a concrete T8-D evidence/reuse claim needs them

Exact staging paths:

```text
docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md
docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md
docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md
```

Do not route target design through superseded/historical architecture or current table/repository existence. Reviewer output and T8-D staging remain non-authoritative.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-C                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-C
Post-T6 Stage-Decomposition GCR          OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-D                                     ACTIVE / CORRECTED CANDIDATE / BOUNDED ROUND 2 NEXT
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

## Reference-backed decision law

```text
MetalDocs semantic/product authority
→ DevelopmentConexus Engineering Method
→ current repository evidence
→ primary/current standards and official tool/library documentation
→ relevant reference products/patterns as falsification evidence
→ credible alternatives + Global Maximum
→ proof/adversarial review
→ operator ratification
```

## Durable baseline through T8-C

T7:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

T8-A:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

T8-B:

```text
ONE GO MODULE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION ORCHESTRATION
+ TRANSPORT -> APPLICATION AS SOLE SEMANTIC INBOUND DOOR
+ NON-SEMANTIC PLATFORM MECHANISMS
+ CLOSED-WORLD DEFAULT-DENY DEPENDENCY GRAPH
```

T8-C:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

Key T8-C laws consumed by T8-D:

```text
database/sql family txscope
PostgreSQL READ COMMITTED posture
owner-private SQL only
same-Scope Audit
Authorization sole final ALLOW/default-DENY
protected User eligibility serialization
owner VersionToken / WorkingContent OCC
ManagedContent create-once + AdmissionClaims + two-phase GC
required River intent in same semantic commit
Idempotency same-key race must not poison Scope
ReplaySnapshot versioned/self-contained/PII-free/snapshot-only
```

## T8-D Round-1 result

Original candidate class:

```text
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+ DECLARATIVE CORRECTNESS
+ PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+ READ COMMITTED NARROW SERIALIZATION
+ EXPLICIT CAS
+ IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+ TRANSACTIONAL KEY↔REPLAY COMPLETION
+ THIRD-PARTY RIVER SCHEMA ISOLATION
+ SELECTIVE PROOF-BACKED PROPERTY REUSE
```

Independent Fable Round 1:

```text
APPROVE WITH MATERIAL FIXES
BLOCKER 2
MAJOR 11
LOW 10
Global Maximum CONFIRMED
upstream reopen NO
stage trespass NO
```

Publication delta from reviewed candidate HEAD was exactly one review artifact.

## Lead adjudication — operator approved

```text
B1 River PG16 / ownership         ACCEPT
B2 attach-vs-GC                   ACCEPT
M1 frozen candidate FK            ACCEPT
M2 immutable malware proof        ACCEPT
M3 provider subject uniqueness    ACCEPT WITH NARROWING
M4 User lock ordering             ACCEPT
M5 transaction census             ACCEPT
M6 GovernanceAttempt uniqueness   ACCEPT
M7 Company/company_id subtraction REJECT
M8 Go<->DDL vocabulary parity     ACCEPT
M9 idempotency cleanup order      ACCEPT
M10 HMAC fingerprint              ACCEPT
M11 DB trust classes/proof role   ACCEPT
```

No T1→T7, T8-B or T8-C reopen is proposed.

## Corrected T8-D staging — active Round-2 input

For bounded Round 2:

```text
effective corrected candidate
=
original candidate
+ adjudicated corrected candidate overlay

where conflict exists, overlay controls staging
```

Material corrections:

```text
River schema created/provisioned outside serving; owner != runtime
River self-REINDEX OFF under PG16 floor
river.* classified THIRD_PARTY_MANAGED

semantic managed-content reference write takes managed_content FOR SHARE
GC phase 1/2 take managed_content FOR UPDATE

GovernanceDecision composite FK -> frozen candidate snapshot
immutable insert-only platform.malware_inspections
provider issuer+subject unique only while current
actor/target User rows acquired as one sorted lock class
complete T6 mutation/transaction census
one GovernanceAttempt per governed subject
Company singleton + current semantic company_id retained; no RLS/tenant substrate
blocking static-Go <-> DDL vocabulary parity proof
idempotency retention DELETE Replay then Key
HMAC-SHA-256 semantic fingerprint + fingerprint key version
provisioner / owner / runtime / CI DB trust classes
accepted constraint/proof LOW corrections
```

## Bounded Round-2 scope

Round 2 must attack only the corrected material delta unless a correction exposes a new contradiction.

Required explicit verdicts:

```text
both Round-1 blockers actually closed?
new/surviving BLOCKER / MAJOR / LOW?
Global Maximum still confirmed?
M7 Lead rejection technically sustainable?
River v0.37.1 + PG16 privilege model coherent?
attach-vs-GC race closed on every semantic reference path?
HMAC fingerprint privacy/equality law coherent?
malware proof remains mechanism-only and immutable?
lock ordering free of architecture-induced cycles?
operation census complete?
T8-C reopen?
T8-B/T1→T7 reopen?
T8-E/F/G/T10/T11 trespass?
another review round materially required?
final Lead adjudication may proceed?
```

## Exact next action

```text
BOUNDED FABLE ROUND 2
→ final Lead adjudication
→ explicit operator ratification
→ only then durable T8-D promotion
→ only then T8-E opens
```

Do **not** start migrations, code, T8-E or implementation planning from this staging chain.

Implementation remains **BLOCKED**.
