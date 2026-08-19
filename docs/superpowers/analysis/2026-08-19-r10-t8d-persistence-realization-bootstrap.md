# R10-T8D — Persistence Realization — Bootstrap

```text
ACTIVE STAGING
NON-AUTHORITATIVE
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-D ACTIVE  
> **Upstream authority:** `wiki/architecture/r10-t8c-internal-communication-contracts.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This is the active non-authoritative bootstrap for R10 **T8-D — Persistence Realization**. It routes persistence architecture work; it is not target authority.

---

## 1. Exact T8-D question

> **What is the smallest PostgreSQL persistence realization that makes T1→T8-C invariants and internal contracts structurally enforceable, assigns every persistent fact to its ratified semantic/mechanism owner, and maps required ACID/OCC/serialization behavior to explicit schema/constraint/query/lock rules without foreign SQL, duplicate truth, hidden shared write authority, wire leakage or speculative persistence?**

T8-D consumes the accepted semantic/contract architecture. It does not redesign Product semantics, T8-B topology or T8-C contract ownership by convenience.

---

## 2. Binding inputs

Read current authority in repository order, including:

```text
Product Contract REV001
Whole-Product GCR + 4+1 ownership
T1 semantic state/invariants
T2 governance/effectivity/transaction law
T3 Authorization/Audit + D4
T4 exact content/storage/restore
T5 durable async/search/external effects
T6 canonical API/frontend journeys
T7 migration truth
T8-A technical authority/legacy disposition
T8-B backend topology
T8-C internal communication contracts
Decision Registry + amendments through T8-C
post-T6 implementation-readiness program
TRRB
```

Current schema/migrations/SQL/repositories are **evidence only**. Existing table names, columns, RLS/GUC design, module schemas or repository patterns receive no survival entitlement.

---

## 3. Frozen upstream laws T8-D must preserve

### Semantic homes

Exactly:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

Mechanism persistence may exist for:

```text
managed content
AdmissionClaims / GC technical state
idempotency
River
technical rendering/provider state only when a current contract requires it
```

Mechanism tables do not become semantic owners.

### Persistence ownership

```text
owner-private SQL only
no application semantic SQL
no transport SQL
no platform semantic SQL
no cross-owner/foreign SQL as communication
no hidden shared semantic write repository
```

`platform/postgres` owns pool/driver/transaction mechanics, not business SQL.

### Transaction posture

From T2/T8-C:

```text
one native local ACID product-state transaction per business transition
PostgreSQL READ COMMITTED
+ narrow explicit serialization where required
+ OCC/CAS
+ structural constraints where required
```

Application owns `Runner.Within`; owners participate through `txscope.Scope`.

### Concrete transaction substrate

T8-C selects:

```text
database/sql transaction family
```

T8-D owns the exact PostgreSQL driver/pool/Runner wrapper realization behind that contract.

### Required same-Scope properties

At minimum:

```text
User + UserProfile + ProviderSubjectBinding + Audit creation atomicity
offboarding across Organization/AuthN/AuthZ/Audit atomicity
required Audit evidence with owning mutation
required OfficialRendition intent with Submission semantic commit
idempotency claim/completion with semantic mutation
AdmissionClaim consumption with semantic attachment
GROUP enabled-member snapshot with Step activation
command-time eligibility/AuthZ/domain truth where required
```

---

## 4. T8-D owns

Freeze the target persistent realization for correctness, including:

```text
PostgreSQL schema namespace strategy
tables and persistent state ownership
material columns/types
PK/FK/unique/check constraints
partial/exclusion constraints where materially required
immutable/history relational shapes
owner-private query/write boundaries
WorkingContent generation/OCC realization
owner VersionToken realization
Submission immutability
GovernanceAttempt/Step/member snapshot/Decision/Feedback persistence
Release/effectivity/replacement persistence
obsolescence persistence
Organization/AuthZ/ApplicationSession state
AuditEvent persistence/history attribution
exact-content semantic descriptor persistence
managed-content technical state
AdmissionClaims
GC_PENDING / reclaimability technical state
idempotency claim/replay state
River technical persistence boundary
canonical PostgreSQL Search query/view persistence realization where needed
transaction isolation mapping
serialization roots / lock ordering / row-lock mapping
same-key idempotency concurrency realization
backup/restore-relevant persistence properties where T4 requires them
```

Performance-only indexes may remain later implementation tuning unless they are load-bearing to a proven correctness/operability property.

---

## 5. T8-D does NOT own

```text
semantic lifecycle changes                         upstream T1→T7
backend package/dependency topology               T8-B
internal public contract ownership/signatures     T8-C
exact HTTP/OpenAPI fields/headers/status codes    T8-E
frontend state/query/cache realization            T8-F
runtime process/job schedule/deploy topology      T8-G
Golden Flow proof matrix                          T9
current→target migration/cutover sequencing       T10
implementation tranche decomposition              T11
```

T8-D may identify a real contradiction requiring a bounded upstream reopen, but may not silently resolve it by schema shape.

---

## 6. Persistent-state census T8-D must walk

The final candidate must explicitly disposition every persistent family below as:

```text
PERSIST — semantic owner
PERSIST — technical mechanism
STATIC / CODE AUTHORITY — no row/table
DERIVED / QUERY-ONLY — no persistent duplicate truth
DEFER / NOT LAUNCH
```

### Authentication

```text
ProviderSubjectBinding
ApplicationSession
session revocation/current validity facts
authentication assurance/fresh-auth persistence only if a current named consumer requires it
```

### Organization

```text
single Company root
User stable identity
separately erasable UserProfile
Area
Group
GroupMembership
User enabled/offboarding current truth
```

### Authorization

```text
Role vocabulary
Permission vocabulary
static Role→Permission bundle
RoleAssignment current truth
scope target representation
```

T8-D must decide deliberately whether role/permission vocabulary is code/static reference data vs persistent rows. Do not inherit the legacy catalog merely because tables exist.

### Controlled Documents

```text
DocumentType
numbering configuration/counter state
GovernanceRoute / Step configuration
Template role / eligible-template relationships
Controlled Document stable identity/code
Business Revision ordinal/title/state
WorkingContent current mutable source + generation
immutable Submission + exact frozen governed state
GovernanceAttempt
activated Step snapshot
GROUP enabled-member snapshot
Governance Decision
Governance Feedback
withdrawal/cancellation facts/reasons/provenance
Release/effectivity
OfficialRendition semantic facts
responsible-owner relationship
obsolescence request/governance/completion
exact-content descriptor attached to semantic records
history/provenance facts required by T1/T2/T6
```

Do not create an Artifact/TemplateVersion/Search/Workflow semantic table family by convenience.

### Audit

```text
immutable AuditEvent
trusted action time
actor attribution
operation/resource attribution
historical Company/Area visibility attribution
bounded facts
```

Audit does not store reconstructed current lifecycle state.

### Managed-content mechanism

```text
opaque handle technical metadata where required
create-once state/proof support
AdmissionClaim lifecycle
GC eligibility/GC_PENDING technical state
provider locator mapping internal to mechanism only
backup pin/exclusion technical state where required
```

No generic semantic `owner_type/owner_id` registry.

### Idempotency mechanism

```text
scoped key identity
semantic fingerprint
claim/completed state needed internally
versioned opaque PII-free ReplaySnapshot payload
bounded expiry/retention policy realization
```

No public/durable business IN_PROGRESS state.

### River

```text
River-owned technical tables only
named OfficialRendition job payload/identity mapping as required by River
```

River tables never become product workflow authority.

---

## 7. Material invariants requiring structural realization

T8-D must create an invariant matrix mapping each accepted property to the strongest reasonable PostgreSQL/type/query/lock enforcement.

At minimum cover:

### Identity / uniqueness

```text
stable IDs
normalized DocumentType.code uniqueness within Company
normalized Area.code uniqueness within Company
Document.code committed uniqueness and non-reuse
Revision ordinal uniqueness/non-reuse within Document
provider issuer+subject binding uniqueness/current-binding law
RoleAssignment duplicate prevention
GroupMembership duplicate prevention
```

### Revision/effectivity

```text
at most one current EFFECTIVE Revision per Document
successful replacement supersedes prior EFFECTIVE atomically
older EFFECTIVE remains official while newer Revision is DRAFT/SUBMITTED
cancelled ordinal never reused
Release is system-owned consequence, not mutable latest-file pointer
```

### WorkingContent OCC

```text
one monotonic generation protects Revision title + WorkingContent source mutation
expected generation mismatch -> zero mutation
successful mutation increments exactly once
```

### VersionToken replacements

For each T6 whole-replacement resource:

```text
stale expected token -> zero mutation
successful material replacement advances version exactly as defined
exact already-current repeat -> no version/Audit fabrication; return current token
```

T8-D must choose a persistence representation capable of deterministic strong ETag mapping later without making ETag itself persistence authority.

### Submission / governance immutability

```text
Submission exact governed attempt immutable after creation
return/resubmit creates a new Submission, never rewrites the old one
activated Step snapshot immutable
GROUP member snapshot immutable after activation
Decision immutable
feedback/history provenance truthful
```

### Authorization / offboarding

```text
User eligibility serialization with offboarding
GroupMembership add target eligibility serialization
new direct User RoleAssignment target eligibility serialization
session issue eligibility serialization
governance/user mutation actor eligibility serialization
responsible-owner target eligibility serialization
Group deletion dependency safety
```

### Audit

```text
required Audit append shares local commit
AuditEvent immutable
historical Company/Area attribution immutable/current relocation independent
```

### Exact content

```text
semantic exact-content descriptor = SHA-256 + size + closed format
provider handle/key never semantic identity
immutable Submission/Rendition content facts never silently repoint to different bytes
```

### Admission / GC

```text
live AdmissionClaim protects handle from reclaimability
claim consumption atomic with semantic attachment
safe rollback/release/expiry
phase-1 GC eligibility proof + GC_PENDING
full phase-2 immediate pre-delete re-proof
safe failure = leaked bytes, never deleted governed truth
```

### Idempotency

```text
scoped same key + same semantic fingerprint serializes
winner commit -> loser can replay without poisoned transaction under READ COMMITTED
winner rollback -> contender can become claimant
same key + different fingerprint -> conflict / zero business mutation
semantic mutation + Audit + required durable intent + completed ReplaySnapshot commit atomically where applicable
ReplaySnapshot remains PII-free and self-contained
```

### River intent

```text
required OfficialRendition Submission commit <=> required River intent commit
job cannot be worked before the transaction commits
rollback removes the intent with semantic rollback
```

---

## 8. Cross-owner FK / constraint question

T8-D must decide explicitly when a PostgreSQL FK/check is:

```text
owner-private structural integrity
```

versus when it would create:

```text
cross-owner persistent authority / forbidden communication coupling
```

Do not assume either:

```text
"all FKs are good"
"cross-owner FKs are always forbidden"
```

For every proposed cross-owner FK/constraint ask:

1. Does it enforce identity existence only, or another owner's lifecycle/business meaning?
2. Can the owning mutation satisfy it without foreign SQL?
3. Does it couple deletion/lifecycle behavior across authorities unintentionally?
4. Does it preserve offboarding/erasure semantics?
5. Would an application-routed contract be the correct semantic boundary instead?

The database may enforce referential impossibility while semantic decisions remain owner-owned, but only when ownership remains unambiguous.

---

## 9. Transaction and lock matrix

T8-D must produce a complete transaction/serialization matrix for every material mutation family.

For each operation record:

```text
application leaf
participating owners/mechanisms
rows/facts read for authority
rows/facts written
serialization root(s)
lock acquisition order
OCC/version predicate
required constraint(s)
Audit append participation
idempotency participation
River intent participation
external work before/after transaction
retryable PostgreSQL failure classes
```

At minimum include:

```text
session issuance
User create
offboarding
UserProfile/eligibility replacement
Area/Group replacement/lifecycle
GroupMembership add/remove
RoleAssignment create/revoke
Group delete
Document create
next Revision
DRAFT update
SUBMIT
feedback
ACCEPT / RETURN_FOR_CHANGES
withdraw
cancel
responsible-owner change
DocumentType/governance/template configuration replacement
obsolescence initiation/withdrawal/final completion
OfficialRendition completion
T5-J GC eligibility mark
idempotent creation/replay acquisition
```

Global table locking / SERIALIZABLE-by-default is not the baseline.

---

## 10. Query/read realization

T8-D must freeze only query structures that affect correctness/contract feasibility.

At minimum:

```text
Authorization canonical direct/group RoleAssignment evaluation
AuthorizedScopes query
ControlledDocs AccessFacts single/batch queries
Library canonical current-effective search/filter query
My Work authoring/governance queries
Document Official/Work/History queries
Audit historical visibility-before-pagination query
Group deletion dependency queries
responsible-owner/RoleAssignment eligibility reads
GC semantic reference proofs
idempotency same-key concurrency query path
```

No foreign owner query may be implemented by reaching into another owner's tables from the caller's repository. Cross-owner composition stays application-routed through T8-C contracts.

Views are allowed mechanisms only when their ownership/query semantics remain clear. A cross-owner view that becomes shared truth is suspect and must be challenged.

---

## 11. Current-schema selective reuse gate

T8-A's five-part gate applies to every current table/query/migration mechanism proposed for reuse:

```text
1. named current R10 consumer
2. public/semantic meaning free of legacy authority
3. dependency direction fits target
4. proof asserts target property
5. reuse smaller than rewrite after transition cost
```

Current evidence receives no survival entitlement from:

```text
existing table name
existing migration volume
existing repository tests
legacy API dependency
historical RLS/GUC assumptions
```

T8-D must explicitly disposition major current schema/query families as:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
```

Do not redesign the transition/cutover plan; concrete current→target moves remain T10.

---

## 12. Alternatives T8-D must compare

For material persistence choices, compare the smallest credible alternatives rather than assuming a familiar ORM/schema pattern.

Examples of likely comparison sets:

### PostgreSQL namespace

```text
one product schema with owner-owned tables
separate PostgreSQL schemas per semantic owner
hybrid technical schema + semantic schema strategy
```

### persistence implementation

```text
database/sql + explicit owner-private SQL
code generation/query tooling if a current proven consumer justifies it
ORM/repository framework only if evidence beats explicit SQL on total complexity
```

No generic repository framework by default.

### version/OCC realization

```text
explicit monotonic version column
hash/token derived from immutable current representation
provider/system MVCC metadata only if correctness/proof is sustainable
```

### effectivity uniqueness

```text
partial unique constraint
explicit pointer/current row
other relational shape that proves at-most-one current EFFECTIVE without duplicate authority
```

### immutable governed history

```text
append-only rows + mutation prohibition
relational lifecycle rows with structural immutability guards
other minimal form that makes historical rewrite demonstrably impossible
```

Choose by target invariant, ownership, proof, operational clarity and future retrofit cost — not legacy familiarity.

---

## 13. External reference / practice law

For material PostgreSQL/Go/database-library behavior:

```text
repo authority first
→ current repository/version evidence
→ PostgreSQL official current documentation
→ Go database/sql official documentation
→ River pinned source/current official docs where load-bearing
→ other primary tool/library documentation
→ reference implementations/patterns as falsification evidence only
```

Use Context7 when AGENTS.md requires it for current library/framework behavior.

Reference practice does not create Product requirements.

---

## 14. Required T8-D outputs

A promotable T8-D candidate must contain:

```text
complete persistent-state census + owner/disposition
exact target PostgreSQL namespace strategy
exact table/state ownership map
material table/column/type design
PK/FK/unique/check/partial/exclusion constraints required for correctness
immutable/history realization
OCC/VersionToken realization
transaction + serialization + lock matrix
lock ordering rules
same-transaction cross-owner/mechanism persistence mapping
Audit persistence/read mapping
Authorization query/grant persistence mapping
ControlledDocs lifecycle/effectivity persistence mapping
managed-content / admission / GC mechanism persistence
idempotency persistence + concurrency mapping
River persistence boundary
canonical Search/query/view realization where material
current-schema selective-reuse disposition
backup/restore/erasure persistence implications where material
failure/fail-closed behavior
proof/enforcement strategy
credible alternatives for disputed persistence families
Structural Inversion + subtractive/YAGNI pass
adversarial challenge
reopen triggers
operator-ratifiable candidate
```

No Writer may later have to invent a material target table/constraint/lock/ownership decision that T8-D should have frozen.

---

## 15. Required proof questions

T8-D must be able to answer falsifiably:

```text
Can two EFFECTIVE Revisions exist for one Document?
Can a cancelled Revision ordinal be reused?
Can immutable Submission/governance history be updated silently?
Can stale WorkingContent/VersionToken mutation commit?
Can required Audit be omitted while semantic mutation commits?
Can a required rendition intent be lost after Submission commit?
Can offboarding race allow a post-disable session/grant/governance action?
Can a foreign owner mutate/query another owner's semantic tables directly?
Can an AdmissionClaim disappear while attachment commits?
Can GC delete newly referenced governed bytes between phase 1 and phase 2?
Can concurrent same-key idempotency poison the loser Scope under READ COMMITTED?
Can same key with different semantic fingerprint mutate business state?
Can erasable UserProfile PII survive solely inside ReplaySnapshot?
Can Audit pagination leak events outside authorized historical attribution?
Can Search/read models become duplicate persistent truth?
```

If the target design cannot make a claim falsifiable at architecture/proof level, it is not ready for promotion.

---

## 16. Stage boundaries

```text
T8-C contracts/ownership                 CLOSED / INPUT
T8-D relational persistence/locks        ACTIVE
T8-E exact wire/OpenAPI                   NOT OPEN
T8-F frontend realization                NOT OPEN
T8-G runtime/process/deploy               NOT OPEN
T8-H whole-T8 coherence                   NOT OPEN
T9→T12                                    NOT OPEN
implementation                            BLOCKED
```

T8-D may name wire needs only as consequences; it must not decide exact JSON/header/status encoding.

T8-D may name runtime needs only as consequences; it must not decide process count, schedule or deploy topology.

---

## 17. Exact next action

```text
reconstruct complete persistent-state/invariant census from T1→T8-C
→ remeasure current schema/query evidence only where a concrete reuse claim needs it
→ map every persistent fact to semantic owner or technical mechanism
→ classify PERSIST / STATIC / DERIVED / DEFER
→ derive correctness constraints before table convenience
→ derive transaction/serialization/lock matrix
→ derive exact owner-private query/persistence boundaries
→ compare credible namespace/schema/version/history alternatives
→ apply T8-A selective-reuse gate
→ apply Method + Structural Inversion + subtractive pass
→ adversarial challenge
→ operator-ratifiable T8-D candidate
```

Implementation remains **BLOCKED**.
