# Current Agent Handoff

> **Last verified:** 2026-08-20  
> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / BOTH FABLE PASSES COMPLETE / FINAL LEAD ADJUDICATION MATERIALIZED / FINAL OPERATOR RATIFICATION NEXT; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
10. T8-D staging chain below
11. current schema/migrations/SQL/code only for a concrete evidence/reuse claim

Current implementation proves what exists, not what survives. Staging/reviewer artifacts are not authority.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-C                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-C
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-D                                     ACTIVE / FINAL OPERATOR RATIFICATION NEXT
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

## Binding execution/reference law

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

For material technical decisions:

```text
MetalDocs semantic/product authority
→ DevelopmentConexus Engineering Method
→ current repository evidence
→ primary/current standards + exact pinned library evidence
→ credible alternatives + Global Maximum
→ proof/adversarial review
→ operator ratification
```

Reference evidence informs mechanism feasibility; it never creates Product requirements.

## Durable baseline through T8-C

T8-A:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
```

T8-B:

```text
ONE GO MODULE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

Semantic homes remain exactly:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

T8-C:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

Key T8-C laws consumed by final T8-D staging:

```text
database/sql transaction family
PostgreSQL READ COMMITTED posture
application owns txscope lifecycle
owner-private SQL only; no owner imports/cross-owner SQL
same-Scope required Audit
Authorization final ALLOW/default-DENY
ProtectedSecuritySubjectIn serializes enabled User against offboarding
owner VersionToken + WorkingContent generation OCC
GROUP snapshot freezes in same Scope; empty stays empty
ManagedContent create-once + AdmissionClaims + two-phase GC
required OfficialRendition River intent shares semantic commit
same-key idempotency race must not poison Scope
ReplaySnapshot self-contained / PII-free / snapshot-only
```

## T8-D staging chain

```text
original candidate
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md

Round-1 independent Fable review
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md

adjudicated corrected candidate
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md

bounded Round-2 Fable delta review
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-corrected-candidate-fable-delta-review.md

FINAL LEAD ADJUDICATION / ACTIVE RATIFICATION INPUT
  docs/superpowers/analysis/2026-08-20-r10-t8d-persistence-realization-final-lead-adjudication.md
```

Effective final staging candidate:

```text
original candidate
+ adjudicated corrected overlay
+ final Lead adjudication overlay
```

## Review convergence

```text
Round 1:
  APPROVE WITH MATERIAL FIXES
  BLOCKER 2 / MAJOR 11 / LOW 10
  Global Maximum CONFIRMED

Round 2:
  APPROVE CORRECTED DELTA WITH MATERIAL FIXES
  BLOCKER 0 / MAJOR 7 / LOW 6
  BOTH Round-1 blockers CLOSED
  Global Maximum CONFIRMED
  upstream reopen NO
  stage trespass NO
  third review round NOT REQUIRED

Final Lead:
  7/7 Round-2 MAJORs adjudicated/closed in staging
  6/6 Round-2 LOWs adjudicated/closed in staging
  surviving material contradiction 0
```

No T1→T7, T8-B or T8-C reopen is proposed.

## Final selected T8-D posture

Global Maximum:

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

Final adjudicated precision includes:

```text
schemas authn / org / authz / controlled_docs / audit / platform / river
fully-qualified owner-private SQL
PostgreSQL-16-compatible persistence floor
static Role/Permission; persisted RoleAssignment
no Launch RLS/tenant substrate
explicit BIGINT VersionToken + WorkingContent generation
Revision.state canonical lifecycle + immutable Release evidence
closed relational governance + structural frozen-candidate FK
immutable managed-content descriptor + immutable malware evidence
semantic attach FOR SHARE vs GC FOR UPDATE
GC downstream proofs non-locking
AdmissionClaim reserve at claim-bound OPEN allocation
Area lifecycle serialization
paired Idempotency Key↔Replay + HMAC fingerprint drain law
River self-REINDEX OFF; runtime never owner
provisioner / owner / runtime / verifier trust classes + grant parity proof
Company singleton fail-closed no-isolation interlock; semantic company_id retained
backup-pin acquire/release locking contract
closed vocabulary ↔ DDL CHECK equality obligation; execution in T9
zero semantic lifecycle triggers baseline
```

## Exact next action

```text
EXPLICIT OPERATOR RATIFICATION OF FINAL T8-D EFFECTIVE STAGING CANDIDATE
```

Do not perform durable promotion until the operator explicitly ratifies.

After ratification only:

```text
consolidate one durable T8-D authority
append Decision Registry T8-D amendment
tombstone/retire T8-D staging artifacts with Git history preserved
mark T8-D CLOSED / OPERATOR-RATIFIED / PROMOTED
open T8-E Executable Wire Contract
```

Implementation remains **BLOCKED**.