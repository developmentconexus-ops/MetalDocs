# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T8-C CLOSED / OPERATOR-RATIFIED; T8-D ACTIVE / PERSISTENCE REALIZATION; T8-E→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
11. current schema/migrations/SQL/code only when a concrete T8-D evidence/reuse claim needs them

Do not route target design through superseded/historical architecture, T8-C staging provenance or current table/repository existence.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-C                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-C
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-D                                     ACTIVE / PERSISTENCE REALIZATION
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

Do not reinvent a technical mechanism already solved by the selected stack unless a concrete MetalDocs invariant requires an additional boundary. Do not import an external best practice when it conflicts with MetalDocs authority.

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
semantic-owner public APIs are concrete; no default owner interfaces
real mechanism/resolver seams are narrow and consumer-owned
application routes cross-owner facts; owners never import one another
transaction substrate = database/sql family
T2 PostgreSQL READ COMMITTED posture remains binding
application owns txscope Runner.Within lifecycle
owner/application SQLTx/native-tx use forbidden; platform River adapter is named native-binding consumer
owner-authored Audit evidence appends in same Scope
Audit historical visibility filters before pagination
Authorization alone returns final ALLOW/default-DENY
AuthorizedScopes is prefilter only, never exact-resource authorization
ProtectedSecuritySubjectIn represents serialization with offboarding/disable
owner VersionToken/expected-version contract protects whole replacements
GROUP enabled-member snapshot freezes in same Scope; empty stays empty/no fallback
ManagedContent PresignCreate = create-once/no-overwrite
AdmissionClaims protect in-flight attachment from GC
malware CLEAN is bound to digest of the exact admitted bytes
required OfficialRendition River intent shares the semantic transaction
idempotency same-key concurrency must not poison Scope under T2 READ COMMITTED
ReplaySnapshot = versioned + self-contained + PII-free + snapshot-only reconstruction
T5-J GC host = internal/application/maintenance
T5-J requires full semantic/live-reference re-proof immediately before provider delete
```

T8-C independent-review convergence:

```text
Round 1  Global Maximum class confirmed
Round 2  BLOCKER 0 / surviving material contradiction 0
third Fable round not required
```

T8-C staging artifacts are historical provenance only and must not be used as current authority.

## T8-D — ACTIVE

Active bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-bootstrap.md`

T8-D exact question:

> **What is the smallest PostgreSQL persistence realization that makes T1→T8-C invariants and internal contracts structurally enforceable, assigns every persistent fact to its ratified semantic/mechanism owner, and maps required ACID/OCC/serialization behavior to explicit schema/constraint/query/lock rules without foreign SQL, duplicate truth, hidden shared write authority, wire leakage or speculative persistence?**

### T8-D owns

```text
PostgreSQL namespace/schema strategy
tables + persistent state ownership
material columns/types
PK/FK/unique/check/partial/exclusion constraints required for correctness
immutable/history relational shapes
owner-private SQL/query realization
WorkingContent OCC persistence
owner VersionToken persistence
Submission/governance/Release/effectivity/obsolescence persistence
Organization/AuthZ/ApplicationSession persistence
Audit persistence + historical-visibility query
managed-content technical state + AdmissionClaims + GC_PENDING
idempotency claim/replay persistence + same-key concurrency realization
River technical persistence boundary
canonical PostgreSQL Search/query/view realization where material
transaction/serialization/lock mapping + lock ordering
```

### T8-D does not own

```text
semantic/product changes                         T1→T7
package/dependency topology                     T8-B
internal contract ownership/signatures          T8-C
exact OpenAPI/wire/ETag encoding                T8-E
frontend realization                            T8-F
runtime/process/deploy                          T8-G
transition/cutover                              T10
implementation decomposition                    T11
```

### Persistence decision law

Every target persistent family must be explicitly classified:

```text
PERSIST — semantic owner
PERSIST — technical mechanism
STATIC / CODE AUTHORITY
DERIVED / QUERY-ONLY
DEFER / NOT LAUNCH
```

Current tables/queries/migrations survive only through T8-A's five-part reuse gate.

### Exact next action

```text
reconstruct complete persistent-state/invariant census from T1→T8-C
→ map every persistent fact to semantic owner or technical mechanism
→ classify PERSIST / STATIC / DERIVED / DEFER
→ derive correctness constraints before table convenience
→ derive complete transaction/serialization/lock matrix
→ derive owner-private query/persistence boundaries
→ compare credible PostgreSQL namespace/schema/version/history alternatives
→ inspect/re-measure current schema only for concrete reuse claims
→ apply T8-A reuse gate
→ apply Method + Structural Inversion + subtractive/YAGNI pass
→ adversarial challenge
→ operator-ratifiable T8-D candidate
```

Implementation remains **BLOCKED**.
