# Rebaseline Decision Registry — T8-D Closure Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-20  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **T8-D authority:** `wiki/architecture/r10-t8d-persistence-realization.md`

This bounded amendment reconciles the Decision Registry after T8-D closure. It changes persistence realization only. Product Contract REV001 and T1→T8-C semantic/topology/contract authority remain unchanged.

Registry authority chain is now:

```text
rebaseline-decision-registry.md
→ rebaseline-decision-registry-d4-amendment.md
→ rebaseline-decision-registry-t6-amendment.md
→ rebaseline-decision-registry-post-t6-amendment.md
→ rebaseline-decision-registry-t7-amendment.md
→ rebaseline-decision-registry-t8a-amendment.md
→ rebaseline-decision-registry-t8b-amendment.md
→ rebaseline-decision-registry-t8c-amendment.md
→ rebaseline-decision-registry-t8d-amendment.md
```

## 1. Stage disposition

```text
T8-D Persistence Realization = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Ratified Global Maximum:

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

## 2. Namespace / ownership

```text
one PostgreSQL product-state database
schemas authn / org / authz / controlled_docs / audit / platform / river
first-party SQL fully qualified
closed bidirectional DB-object ownership catalog
river.* = THIRD_PARTY_MANAGED
raw first-party SQL against river.* = forbidden
PostgreSQL 16-compatible feature floor
```

## 3. Persist/static/derived disposition

```text
Authentication      ProviderSubjectBinding + ApplicationSession persist
Organization        Company/User/UserProfile/Area/Group/GroupMembership persist
Authorization       RoleAssignment persists; Role/Permission/bundles are static code authority
Controlled Docs     bounded native document/governance/release/obsolescence facts persist
Audit               immutable AuditEvent persists
Platform            ManagedContent/admission/GC/idempotency technical facts persist
River               River-owned technical objects only
Search/status/authz expansion = derived/query-only
RLS/tenant GUC/generic Workflow/Artifact/outbox/materialized Search = absent/deferred
```

## 4. Lifecycle / effectivity / governance

```text
Revision.state remains canonical lifecycle state
Release remains immutable effectivity-establishing fact
partial unique one EFFECTIVE Revision per Document
partial unique one open DRAFT/SUBMITTED Revision per Document
SUBMITTED iff current_submission_id present
Document has no current-status/current-revision/current-release pointer
Document row is lifecycle serialization root
GovernanceAttempt has exactly one subject and one attempt per subject
one ACTIVE Step per attempt
NAMED_USER/GROUP activation always materializes frozen candidates
GovernanceDecision FK proves decider was in frozen candidate snapshot
GROUP live dependency exists only until activation
historical activated GROUP snapshot does not keep Group identity alive
```

## 5. OCC / serialization

```text
owner whole-replacement VersionToken = explicit monotonic BIGINT
DRAFT title/source OCC = WorkingContent generation
READ COMMITTED inherited
all authenticated user-initiated semantic mutations protect actor User via FOR SHARE
multiple User locks dedupe/sort UUID/acquire once at strongest mode
Document lifecycle/effectivity = FOR UPDATE
DocumentType current config snapshot = FOR SHARE / replacement FOR UPDATE
Area create eligibility = FOR SHARE / lifecycle FOR UPDATE
```

## 6. Exact-content / GC

```text
semantic exact descriptor belongs to semantic row
ManagedContent mutable row owns mechanism lifecycle/location only
immutable platform.managed_content_descriptors proves READY exact bytes
immutable platform.malware_inspections records terminal CLEAN/MALICIOUS evidence
AdmissionClaim is created at claim-bound allocation and spans OPEN/READY until consume/release/expiry
every semantic managed_content reference write takes ManagedContent FOR SHARE through commit
backup-pin acquisition also takes ManagedContent FOR SHARE
GC phase 1/2 take ManagedContent FOR UPDATE as sole GC serialization lock
GC downstream semantic/claim/pin proofs are non-locking current reads
phase 2 repeats all proofs immediately before provider delete
safe failure = leaked bytes, never deleted governed truth
```

## 7. Idempotency

```text
paired platform.idempotency_keys + platform.idempotency_replays
scoped unique actor+operation+key
deferred reverse FK makes committed incomplete replay impossible
no durable target IN_PROGRESS/FAILED state
same-key loser uses READ COMMITTED without poisoned Scope
semantic fingerprint = HMAC-SHA-256 over complete validated semantic command
positive fingerprint key version persisted
one derivation version active at a time; rotation drains prior idempotency keys first
retention deletes Replay before Key
ReplaySnapshot remains bounded, versioned, self-contained and PII-free; exact wire-size maximum is T8-E
```

## 8. River / DB trust

```text
river.* owned by metaldocs_owner
serving runtime never owns river.* and never receives owner membership
River self-REINDEX OFF on PostgreSQL 16
River runtime receives only required schema/type/function/sequence/DML privileges
four trust classes:
  bootstrap/provisioner
  metaldocs_owner NOLOGIN
  metaldocs_runtime serving
  metaldocs_verifier non-owner proof identity
only provisioner may SET ROLE metaldocs_owner
verifier effective grants must equal runtime across closed object catalog
immutable-history security tests run as verifier, not owner/superuser
```

## 9. Query / framework posture

```text
owner-private SQL only
no foreign owner joins
application composes through T8-C contracts
Library/Search remains owner-private ordinary relational query over current EFFECTIVE truth
materialized Search OFF
zero semantic lifecycle trigger baseline
explicit owner-private database/sql SQL is Launch baseline
no generic ORM/repository framework by default
```

## 10. T8-A reuse disposition

```text
password/local-auth tables                    DELETE / REWRITE
current session shape                         REWRITE; preserve durable-current-session property
IAM/access legacy families                    REHOME / REWRITE
role_capabilities                             DELETE
RLS/tenant/GUC                               DELETE from Launch target
controlled_documents                         REWRITE
technical document_revisions                 DELETE / REWRITE
approval_*                                   REWRITE into bounded governance relations
taxonomy/template platform tables             DELETE/fold
Audit table shape                            REWRITE / REHOME
Audit append-only runtime restriction         PRESERVE PROPERTY
Audit global hash chain                      DELETE
current idempotency status/raw HTTP           REWRITE
unique-key contention property                PRESERVE / REFINE
River mechanism                              PRESERVE / REHOME to river.*
runtime != DDL-owner property                 PRESERVE / REFINE
closed ownership-catalog completeness         PRESERVE PROPERTY / REWRITE catalog
outboxes/notifications                        DELETE absent current Launch consumer
```

## 11. Independent-review convergence

```text
Round 1: BLOCKER 2 / MAJOR 11 / LOW 10; Global Maximum CONFIRMED
Round-1 Lead adjudication: both blockers corrected; M7 subtraction rejected
Round 2: BLOCKER 0 / MAJOR 7 / LOW 6; both Round-1 blockers CLOSED
Round 2: Global Maximum CONFIRMED; upstream reopen NO; third full review NOT REQUIRED
Final Lead adjudication: 7/7 MAJOR + 6/6 LOW closed; surviving material contradiction 0
Operator ratification: explicit 2026-08-20
```

## 12. Stage boundary

```text
T8-D = CLOSED / PROMOTED
T8-E = ACTIVE / Executable Wire Contract
T8-F→T8-H = NOT OPEN
T9→T12 = NOT OPEN
implementation = BLOCKED
```

T8-E consumes T6 semantic journeys and T8-D persistence/concurrency authority. It may not change persistent meaning or upstream semantics by wire convenience.
