# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / GLOBAL MAXIMUM CANDIDATE MATERIALIZED / INDEPENDENT FABLE REVIEW NEXT; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
11. `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`
12. current schema/migrations/SQL/code only when a concrete T8-D evidence/reuse claim needs them

Do not route target design through superseded/historical architecture, T8-C staging provenance or current table/repository existence.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-C                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-C
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-D                                     ACTIVE / CANDIDATE MATERIALIZED / INDEPENDENT FABLE REVIEW NEXT
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

For material technical decisions:

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

Do not reinvent a technical mechanism already solved by the selected stack unless a concrete MetalDocs invariant requires an additional boundary. Do not import external practice when it conflicts with MetalDocs authority.

## T7 — CLOSED

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs business DB/content/history is DEV/test/throwaway and creates no historical-business compatibility entitlement.

## T8-A — CLOSED

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Ratified law:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Current implementation/schema remains evidence only. PRESERVE requires all five T8-A proofs.

## T8-B — CLOSED

Durable authority:

`wiki/architecture/r10-t8b-backend-module-package-topology.md`

Ratified Global Maximum:

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

Semantic homes remain exactly:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

No direct owner→owner imports, foreign SQL, mechanism-as-authority or second Authorization evaluator.

## T8-C — CLOSED / PROMOTED

Durable authority:

`wiki/architecture/r10-t8c-internal-communication-contracts.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8c-amendment.md`

Ratified Global Maximum:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

Key binding laws consumed by T8-D:

```text
transaction substrate = database/sql family
T2 PostgreSQL READ COMMITTED posture remains binding
application owns txscope Runner.Within lifecycle
owner-private SQL only; no owner imports/cross-owner SQL
owner-authored Audit evidence appends in same Scope
Authorization alone returns final ALLOW/default-DENY
AuthorizedScopes is prefilter only
ProtectedSecuritySubjectIn serializes eligibility with offboarding/disable
owner VersionToken/expected-version protects whole replacements
GROUP enabled-member snapshot freezes in same Scope; empty stays empty
ManagedContent create-once + AdmissionClaims + two-phase GC
malware CLEAN digest must match exact admitted bytes
required OfficialRendition River intent shares semantic transaction
idempotency same-key race must not poison Scope under READ COMMITTED
ReplaySnapshot = versioned + self-contained + PII-free + snapshot-only reconstruction
T5-J GC host = internal/application/maintenance
```

T8-C staging artifacts are historical provenance only.

## T8-D — ACTIVE / INDEPENDENT REVIEW GATE

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-bootstrap.md`

Current candidate:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`

The candidate is **NON-AUTHORITATIVE**. The operator approved the design for materialization; this is not final T8-D ratification.

### Candidate Global Maximum

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

### Principal candidate decisions

```text
one PostgreSQL database
schemas authn/org/authz/controlled_docs/audit/platform/river
fully-qualified first-party SQL
PostgreSQL-16-compatible feature floor
complete bidirectional DB-object ownership catalog
Role/Permission/bundles static; RoleAssignment persisted
no Launch RLS/tenant substrate
explicit BIGINT VersionToken + WorkingContent generation
Revision.state canonical lifecycle + immutable Release fact
partial unique one-open/one-EFFECTIVE barriers
closed relational governance model
live GROUP dependency separate from activated candidate snapshot
semantic exact descriptors + technical ManagedContent
AdmissionClaim row existence + repeated two-phase GC proof
paired Idempotency Key+Replay + deferred completion FK
no durable IN_PROGRESS/FAILED target replay state
River under third-party river.* and no raw first-party River SQL
cross-owner FKs only for stable identity/existence; no semantic cascades
serving runtime DB role != DDL owner
DB/column grants protect immutable-history classes
zero semantic lifecycle triggers baseline
protected actor FOR SHARE + target User update lock + Document FOR UPDATE lifecycle root
owner-private explicit database/sql SQL; no generic ORM/repository framework
materialized Search OFF
```

### Current-schema selective-reuse posture

```text
auth/password tables          DELETE / REWRITE
current sessions              REWRITE; preserve durable-current-session property
IAM/Role current families     REHOME / REWRITE
role_capabilities             DELETE
RLS/tenant substrate          DELETE from Launch target
controlled_documents          REWRITE
technical document_revisions  DELETE / REWRITE
approval_*                    REWRITE into bounded governance relations
taxonomy/template platforms   DELETE/fold into ratified owners
Audit table shape             REWRITE / REHOME
Audit append-only grants      PRESERVE PROPERTY
current idempotency shape     REWRITE
unique-key concurrency        PRESERVE / REFINE
River                         PRESERVE mechanism / REHOME to river.*
runtime!=DDL role             PRESERVE
ownership-catalog completeness PRESERVE PROPERTY / REWRITE catalog
```

### Independent-review attack surface

Reviewer must independently attack, at minimum:

```text
owner-namespaced schemas vs simpler alternatives
cross-owner FK authority boundary
ProviderSubjectBinding current/history uniqueness
static Role/Permission persistence choice
VersionToken/OCC completeness
Revision.state + Release effectivity split
partial unique concurrency barriers
current_submission_id boundedness
governance relational completeness / GROUP deletion law
runtime DB grants for immutable history
ManagedContent/AdmissionClaim/GC attach-delete races
malware proof persistence
paired Key↔Replay deferred-FK validity/operability
READ COMMITTED same-key loser behavior
River v0.37.1 custom schema/InsertTx support
runtime-role vs DDL-owner complexity
lock ordering/deadlock matrix
zero-trigger baseline
T8-A reuse dispositions
future-capability subtraction
T8-D/T8-E/T8-G/T10 boundary discipline
```

### Exact next action

```text
independent Fable review of the exact T8-D candidate
→ BLOCKER / MAJOR / LOW findings
→ Global Maximum confirmed yes/no
→ T8-C/upstream reopen yes/no
→ T8-E/T8-G/T10 trespass yes/no
→ Lead confrontation/adjudication
→ bounded correction/re-review only if materially required
→ explicit operator ratification before durable promotion
```

Do **not** start migrations, code, T8-E or implementation planning from the candidate.

Implementation remains **BLOCKED**.
