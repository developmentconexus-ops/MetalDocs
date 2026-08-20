# R10-T8D — Persistence Realization — Bootstrap

```text
ACTIVE STAGING ROUTER
NON-AUTHORITATIVE
ROUND-1 INDEPENDENT REVIEW COMPLETE
BOUNDED ROUND-2 REVIEW COMPLETE
FINAL LEAD ADJUDICATION MATERIALIZED
FINAL OPERATOR RATIFICATION = NEXT
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-20  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-D ACTIVE  
> **Upstream authority:** T1→T8-C durable R10 authority chain  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This file is only the active T8-D staging router. It is not target authority.

## Authority order

Read in repository-mandated order:

```text
AGENTS.md
→ docs/engineering/standards/root-cause-global-maximum-method.md
→ wiki/references/current-agent-handoff.md
→ wiki/architecture/r10-technical-architecture.md (sole stage/status/next-action router)
→ Product Contract REV001 + Whole-Product GCR + 4+1 ownership
→ T1→T8-C durable authorities
→ Decision Registry + amendments through T8-C
→ r10-post-t6-implementation-readiness-program.md
→ r10-technical-realization-reconciliation-baseline.md
→ this staging chain
```

Current schema/migrations/SQL/code are evidence only for concrete T8-A reuse/feasibility claims.

## T8-D staging chain

```text
1. Original Global Maximum candidate
   docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md

2. Round-1 independent Fable review
   docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md

3. Lead-adjudicated corrected candidate
   docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md

4. Bounded Round-2 Fable delta review
   docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-corrected-candidate-fable-delta-review.md

5. Final Lead adjudication — ACTIVE RATIFICATION INPUT
   docs/superpowers/analysis/2026-08-20-r10-t8d-persistence-realization-final-lead-adjudication.md
```

For final operator ratification, the effective staging target is:

```text
original candidate
+ adjudicated corrected overlay
+ final Lead adjudication overlay
```

Later overlays control where staging text conflicts. Reviewer artifacts are evidence only.

## Global Maximum class

Still selected and independently confirmed:

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

## Review convergence

```text
Round 1:
  APPROVE WITH MATERIAL FIXES
  BLOCKER 2 / MAJOR 11 / LOW 10
  Global Maximum CONFIRMED

Lead adjudication:
  both blockers accepted and corrected
  material findings adjudicated
  operator approved corrected-candidate materialization

Round 2:
  APPROVE CORRECTED DELTA WITH MATERIAL FIXES
  BLOCKER 0 / MAJOR 7 / LOW 6
  BOTH Round-1 blockers CLOSED
  Global Maximum CONFIRMED
  upstream reopen NO
  stage trespass NO
  third review round NOT REQUIRED

Final Lead adjudication:
  7/7 Round-2 MAJORs closed in staging
  6/6 Round-2 LOWs closed in staging
  surviving material contradiction 0
```

## Final adjudicated deltas

The final Lead adjudication additionally freezes, without changing the Global Maximum class:

```text
immutable platform.managed_content_descriptors
blanket protected actor FOR SHARE for authenticated user-initiated semantic mutations
target metaldocs_verifier role + runtime/verifier effective-grant parity proof
HMAC fingerprint drain-before-rotation equality law
AdmissionClaim reserve at claim-bound allocation/OPEN
Area lifecycle serialization against Document create
GC ManagedContent FOR UPDATE + non-locking downstream proofs
governance candidate materialization for NAMED_USER and GROUP activation
backup-pin acquire/release persistence + ManagedContent FOR SHARE
explicit River TYPE/object grant class + provisioner-only SET ROLE owner
observable River index-maintenance reopen condition classes
Company singleton fail-closed no-isolation interlock precision
T9 ownership of closed-vocabulary parity validation execution
```

## Exact next action

```text
EXPLICIT OPERATOR RATIFICATION OF FINAL T8-D EFFECTIVE STAGING CANDIDATE
```

Only after an explicit operator ratification may the normal repository process:

```text
→ consolidate one durable T8-D authority under wiki/architecture/
→ append the T8-D Decision Registry amendment
→ tombstone/retire live T8-D staging artifacts while preserving Git history
→ mark T8-D CLOSED / OPERATOR-RATIFIED / PROMOTED
→ open T8-E Executable Wire Contract
```

Until then:

```text
T8-D ACTIVE
T8-E→T12 NOT OPEN
implementation BLOCKED
```

Do not create target migrations, code or OpenAPI from this staging router.