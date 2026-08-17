# R10-B1 Relational Substrate, Tenancy & Reference Law — Independent Fable Review Request

> **Status:** CANDIDATE / INDEPENDENT REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Candidate baseline:** `05a87fa4841ea71128c0538fe86f583075cb4643`
> **Stage:** R10-B — Transactional Domain State & DB Invariants / R10-B1
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this packet is review evidence only. It does not amend `wiki/architecture/r10-technical-architecture.md`, close R10-B1, open R10-B2, authorize schema changes, or authorize product implementation.

---

## 0. Reviewer bootstrap — cold start

Reconstruct the project exclusively from the repository. Do not use conversation memory or current implementation shape as requirement authority.

Read in the order required by `AGENTS.md`:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. this packet
8. current schema/code/runtime only when necessary to falsify or support a specific technical claim

R10-A is already promoted authority. Do not re-litigate its ownership topology merely because a different persistence layout would be convenient.

Apply the DevelopmentConexus Engineering Method directly:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Overengineering / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Apply the Structural Inversion Test aggressively: if the existing `public`/`metaldocs` schemas, current table names, current RLS implementation, or current transaction helpers were opposite, which candidate conclusions would still be required by the promoted product/ownership semantics?

A reviewer finding is evidence, not authority. Do not introduce a new product requirement, bounded context, external service, persistence framework, distributed transaction system, policy engine or compliance platform disguised as a review fix.

---

# 1. Verified stage gate

```text
R3–R9   = LOCKED
R9.5    = FROZEN
R9.5 reopen set = EMPTY

R10-A   = CLOSED / APPROVED
R10-B   = OPEN / DESIGN ONLY
R10-B1  = CANDIDATE
R10-B2..B6 = NOT STARTED

implementation = BLOCKED
```

R10-B decomposition currently is:

```text
R10-B1  Relational Substrate, Tenancy & Reference Law
R10-B2  Authentication / Organization / Authorization State
R10-B3  Controlled Information Authoring & Submission State
R10-B4  Approval / Release / Distribution State
R10-B5  Documentary Context / Records Governance State
R10-B6  Cross-owner Atomicity & Historical Truth Closure
```

Candidate closure order:

```text
B1 → B2 → B3 → B4 → B5 → B6
```

R10-B1 owns only the shared relational/transactional law required before aggregate design. It must not decide R10-C physical storage mechanics, R10-D worker/retry/external-effect execution, R10-E final API/frontend contracts, or R10-F migration/cutover sequencing.

---

# 2. Root cause

The promoted ownership topology now gives every durable/business fact one semantic owner, but the relational substrate can still recreate the old defect class if table layout, keys, FKs, cascade rules, RLS or transaction helpers acquire architectural authority independently.

Current-state evidence demonstrates the failure class, not the target:

- product state is split between `public` and `metaldocs`;
- multiple current `documents`/`controlled_documents` structures overlap stable identity, revision state, authoring and release concerns;
- current `public.document_revisions` is technical/autosave history rather than the frozen business `DocumentRevision` concept;
- current Approval tables bind legacy `document_id`/hash shapes and contain lifecycle vocabulary no longer accepted by R9.5;
- current Audit/outbox demonstrate useful append/durable-intent mechanisms but are not automatically the target table shapes.

Root cause statement:

> R10-A fixed semantic ownership, but without one explicit relational/transactional law, each owner can independently choose identity, tenancy, reference, deletion, immutability and transaction conventions that make cross-tenant references, duplicate authority, incomplete atomicity or historical corruption structurally reachable again.

---

# 3. Target invariant

For every valid target persistence realization:

> each durable fact is stored under its promoted semantic owner; every tenant-owned relationship is tenant-safe; immutable history cannot be silently rewritten; one owner cannot delete another owner's history through referential side effects; canonical business mutations that require multiple durable facts commit all of them or none of them; replaceable mechanisms remain mechanisms.

Derived substrate invariants:

1. one semantic owner per durable fact/table meaning;
2. tenant ownership is explicit in every tenant-owned row;
3. a tenant-owned FK cannot point to another Tenant;
4. business/provider/external identifiers do not become technical row identity;
5. cross-owner references may prove existence/tenant integrity without transferring lifecycle authority;
6. cross-owner deletion never cascades another owner's durable history;
7. snapshots preserve historical truth independently of mutable source rows;
8. frozen vocabularies receive structural checks where proportionate;
9. absence of tenant execution context fails closed for tenant-owned product data;
10. RLS enforces Tenant isolation only and never becomes canonical Authorization;
11. operations requiring cross-owner atomicity participate in one local PostgreSQL transaction;
12. critical governed mutation cannot report success without its required durable Audit append;
13. when a business mutation requires future async work, the required durable intent is inserted in that same transaction; execution/retry remains R10-D;
14. system/background execution does not gain implicit cross-tenant content access merely because it is automated.

---

# 4. Constraints inherited from authority

- R10-A ownership set remains exactly 8 business bounded contexts + 3 supporting semantic owners unless material new evidence reopens it.
- `composition` may coordinate but owns no durable semantic fact.
- shared PostgreSQL/local transaction participation is an accepted technical premise of the modular monolith unless evidence proves it cannot preserve a frozen invariant.
- RLS is tenant-isolation defense-in-depth, not a substitute for Authorization.
- PlatformOperator/SystemPrincipal remain outside tenant RBAC and have no implicit tenant-content access.
- unknown source truth remains unknown; no magic sentinels may fabricate historical meaning.
- provider keys/URLs/versions never become Artifact/Submission business identity.
- no generic `Record`, object platform, ReBAC graph, workflow engine or polymorphic domain registry may be introduced merely to simplify persistence.
- implementation is blocked.

---

# 5. Credible alternatives

## A. Keep current `public` + `metaldocs` product-state split

Local maximum: minimizes migration churn but preserves historical namespace accidents and lets old module cuts influence target state.

## B. One canonical product-state schema: `metaldocs`

Candidate. Semantic ownership remains module/authority metadata, not PostgreSQL-schema ownership. One database/schema keeps local transactions and FK enforcement simple without pretending each bounded context is independently deployable.

## C. One PostgreSQL schema per bounded context

Stronger visual namespace but no real operational isolation: all owners still share database, deployment, references and local transactions. Adds migrations/grants/search-path/cross-schema complexity without eliminating a failure class.

## D. No cross-owner FKs; opaque IDs only between bounded contexts

Reduces declared coupling but permits orphan and cross-tenant references structurally. Application discipline becomes the only integrity barrier for relationships that are already frozen as exact references.

## E. Global `id UUID PRIMARY KEY` on every tenant-owned table

Simpler single-column references, but same-tenant FK integrity then requires separate composite uniqueness/checking or application discipline. It also makes globally unique technical identity more authoritative than tenant ownership for no product requirement.

## F. Composite tenant identity `(tenant_id, id)` for tenant-owned entities

Candidate. Makes same-tenant FK the ordinary relational form and eliminates an entire cross-tenant-reference failure class without triggers or duplicate global-identity indexes.

The reviewer must compare these alternatives rather than judging the candidate in isolation.

---

# 6. Candidate R10-B1 decision

## 6.1 PostgreSQL topology

```text
one PostgreSQL database
one canonical MetalDocs product-state schema: metaldocs
```

Target MetalDocs business/support state does not live in `public`.

No schema-per-Tenant. No schema-per-bounded-context. PostgreSQL schemas are deployment/storage namespaces, not semantic authorities.

Platform-owned PostgreSQL objects may exist where they are genuinely platform mechanics, but target product tables have one canonical namespace and explicit owner attribution in architecture/schema documentation.

## 6.2 Tenant-owned identity law

For a durable tenant-owned entity:

```text
tenant_id UUID NOT NULL
id        UUID NOT NULL
PRIMARY KEY (tenant_id, id)
```

`id` is opaque technical identity. No requirement is introduced for a second `UNIQUE(id)` global index.

Business identity stays separate:

```text
Document.code      != PK
REV label          != PK
Dossier stable key != PK
Artifact hash      != PK
ExternalReference != PK
provider key/URL   != PK
```

Exceptions follow fact semantics rather than convenience:

- `Tenant` is the root and therefore is not tenant-owned by another Tenant;
- genuinely global/product/credential facts may lack `tenant_id` when their owning semantics require it; their exact design belongs to B2 or the owning later block.

Do not mechanically add `tenant_id` to every table merely for uniformity.

## 6.3 Same-tenant reference law

A relationship from one tenant-owned row to another uses tenant-qualified identity whenever relational existence is a required invariant:

```text
FOREIGN KEY (tenant_id, target_id)
REFERENCES target_table(tenant_id, id)
```

This FK proves only:

```text
existence
same Tenant
identity
```

It does not transfer business/lifecycle authority to the referencing owner.

Cross-owner FKs are allowed when the frozen model requires an exact live reference, for example Approval eventually binding an exact immutable `RevisionSubmission`.

## 6.4 Historical snapshot law

When a frozen fact means “preserve what was true at that moment”, persistence stores an immutable snapshot rather than re-deriving it later from mutable source state.

Examples include participant snapshots, distribution audience snapshots, attestation evidence, imported historical actor/source facts and other decision-relevant provenance.

A source reference may remain for correlation if semantically justified, but source mutation never rewrites the snapshot.

## 6.5 External/provenance reference law

External identities are provenance/correlation, not internal identity and never database FKs to external systems.

The owning domain stores typed connection/entity/external identifiers only to the degree required by the frozen contract.

## 6.6 No universal polymorphic business reference

R10-B1 rejects a generic domain table/column pattern such as:

```text
resource_type/resource_id
subject_type/subject_id
owner_type/owner_id
```

as the universal representation of business relationships.

Material relationships use typed tables/FKs owned by the owner of the relationship. This preserves referential proof and semantic meaning.

Audit may carry a generic resource attribution because it is explicitly a non-authoritative transversal timeline; that exception does not license generic domain persistence.

Artifact's no-confirmed-orphan relation is deliberately deferred to B3/B5 because its exact structural backstop depends on the two real semantic consumers; B1 does not create a generic attachment registry to solve it prematurely.

## 6.7 Delete/cascade law

Across semantic owners:

```text
ON DELETE CASCADE = forbidden
ON UPDATE CASCADE = forbidden/unnecessary for immutable technical identity
normal FK action  = RESTRICT / NO ACTION
```

Deletion/disposition/erasure across owners is explicit coordinated behavior, never an incidental FK cascade.

Within one owner, cascade is allowed only for a child with no independent historical meaning whose legal lifecycle is strictly subordinate to the parent. Later owner blocks must justify each cascade; cascade is not the default.

Tenant terminal erasure is coordinated lifecycle work ending in authoritative erasure/tombstone state, not a root `DELETE` whose cascades define product semantics.

## 6.8 Primitive persistence law

Candidate defaults:

```text
technical IDs      = UUID
business instants  = TIMESTAMPTZ
canonical SHA-256  = BYTEA with octet_length(hash)=32
frozen vocabulary  = TEXT + CHECK by default
unknown/absence    = NULL when semantically real
```

Do not use empty strings, zero UUIDs, zero numeric sentinels or fabricated `UNKNOWN` values to avoid legitimate nullability.

Historical Migration must preserve unknown as unknown.

`JSONB` is permitted for bounded whole snapshots or genuinely variable provider-neutral provenance whose atomic semantic is the document itself. It is not a default escape hatch for unmodeled business state.

PostgreSQL ENUM is not the default: `TEXT + CHECK` gives strong validation with lower migration coupling for the frozen but evolvable vocabularies.

## 6.9 Tenant isolation law

For tenant-owned product tables:

```text
application/repository tenant predicate
+ same-tenant relational keys/FKs
+ ENABLE ROW LEVEL SECURITY
+ FORCE ROW LEVEL SECURITY
```

RLS is Tenant isolation only. It must not encode role, Area, Dossier, Approval-participant, Document-owner or other canonical Authorization predicates.

Ordinary tenant-scoped execution must seed an explicit transaction-local Tenant context. Missing Tenant context for tenant-owned product rows must fail closed rather than widen visibility.

The exact SQL/GUC/policy expression is implementation work, but the protected behavior is fixed here.

## 6.10 Database role law

Ordinary runtime DML must execute through a role that is:

```text
NOSUPERUSER
NOBYPASSRLS
not owner of product tables
```

DDL/object ownership is separate from ordinary runtime DML authority.

R10-B1 does not create one database role per bounded context: the product has no evidenced per-context deployment/trust boundary requiring eleven PostgreSQL identities.

Existing current-state evidence where RLS tests falsely passed under owner/superuser connections is treated as evidence for this law, not as entitlement to retain the current role model.

## 6.11 Transaction law

Every ordinary tenant-owned mutation runs in an explicit local PostgreSQL transaction so transaction-local execution context and all mandatory durable facts share one commit boundary.

Single-owner mutation:

```text
owner application service
  owns the transaction boundary
```

Cross-owner atomic use case:

```text
composition
  opens one PostgreSQL transaction
  → calls owner-specific application seams
  → every participant uses the same transaction
  → one COMMIT or one ROLLBACK
```

No semantic owner may hide an independent nested commit inside a cross-owner atomic operation, and no owner may reach into another owner's repository merely to obtain atomicity.

The concrete UnitOfWork/Tx interface and package dependency inversion belong to later R10-B design, not B1.

## 6.12 Isolation-level law

Default transactional isolation is `READ COMMITTED`.

R10-B1 rejects global `SERIALIZABLE` as a substitute for modeling invariants.

Later blocks use the narrowest credible mechanism per invariant:

```text
UNIQUE / partial UNIQUE
CHECK
FK
atomic UPDATE / compare-and-swap
working_version
SELECT ... FOR UPDATE
```

A stronger isolation level requires a demonstrated failure class not sufficiently protected by a narrower invariant mechanism.

## 6.13 Same-commit Audit and durable intent law

Where frozen authority says a critical governed mutation requires Audit evidence, the transaction cannot report success unless the required Audit append is durable in the same commit.

Where a business mutation necessarily creates future async work, the durable intent required to cause that work is inserted in the same commit as the business mutation.

Conceptually:

```text
BEGIN
  business facts
  required Audit append
  required durable async intent
COMMIT
```

or all roll back.

This decision does not choose R10-D retry/worker/DLQ execution semantics and does not require the current outbox table shape to survive.

## 6.14 Immutability classification law

Every durable table/fact family in B2–B5 must be classified before closure as one of:

```text
MUTABLE AUTHORITY
APPEND-ONLY / IMMUTABLE HISTORY
TERMINAL / TOMBSTONED
EPHEMERAL MECHANISM
PROJECTION / ATTRIBUTED SUPPORT
```

Facts frozen as immutable history cannot rely solely on “application code normally does not call Update”. The later block must choose the strongest reasonable enforcement proportionate to the fact: privileges, schema design, trigger, constrained transitions or another falsifiable mechanism.

Known examples requiring immutable treatment include `RevisionSubmission`, `ApprovalDecision`, participant/audience snapshots, `AcknowledgementRecord`, `PeriodicReviewRecord`, retention snapshots, completed disposition evidence, AuditEvent and CAPTURED Evidence governed content.

---

# 7. Operational viability — MetalDocs must actually run

The target must support API traffic, tenant-scoped background work, release/review jobs, retention/disposition processing and tenant erasure without requiring implicit global content access.

## 7.1 No system-principal superuser shortcut

Automation does not acquire semantic access merely because it is a worker or `SystemPrincipal`.

A background operation touching tenant-owned rows must run with an explicit Tenant execution context and the operation's authorized system/application semantics.

## 7.2 Global discovery vs tenant-owned work

A global scheduler may discover which Tenants require work through an explicitly owned root/platform surface that does not expose arbitrary tenant content.

Once a Tenant is selected, work touching tenant-owned product rows proceeds tenant-by-tenant:

```text
discover eligible Tenant IDs
→ for each Tenant
   BEGIN tenant-scoped transaction
   seed Tenant/system execution context
   execute owner/application use case
   COMMIT/ROLLBACK
```

The candidate rejects an ordinary “unset Tenant context = see all tenants” path as an operational convenience.

## 7.3 Long-running cross-owner workflows

R10-B1 only requires atomicity where frozen semantics require one local atomic state transition. Tenant erasure, physical disposition and external publication can remain multi-step coordinated processes where authority already models verified completion rather than pretending one DB transaction includes external effects.

R10-C/D will own those physical/external mechanics.

## 7.4 Recovery and observability

Fail-closed Tenant context, FK rejection, uniqueness violation or audit/intent insertion failure must surface as an operation failure, not silently degrade to partial success. Exact problem codes/observability contracts are later-stage concerns.

---

# 8. Deliberate deferrals

R10-B1 does **not** decide:

```text
Authentication credential ↔ Organization User representation       → B2
Role/Permission/RoleAssignment exact tables                         → B2
Document owner/responsibility participant type/cardinality          → B3
Document/Revision/WorkingContent/Submission exact table set          → B3
Artifact no-confirmed-orphan exact relational backstop               → B3/B5
Approval exact table/state model                                     → B4
Evidence/Dossier/Records exact table/state model                      → B5
operation-by-operation cross-owner transaction matrix                 → B6
outbox execution/retry/DLQ                                            → R10-D
provider storage/relocation/physical delete                           → R10-C
final API/DTO/frontend                                                 → R10-E
migration/backfill/cutover execution                                  → R10-F
```

Deferral is safe because B1 fixes only the shared laws those decisions must obey.

---

# 9. Current-state evidence anchors

Current state is evidence only; the reviewer should inspect it when useful:

- `wiki/database/dictionary-index.md` — current dual-schema/table ownership inventory;
- `wiki/database/tables/documents.md` — overlapping current `metaldocs.documents` / `public.documents` state;
- `wiki/database/tables/controlled_documents.md` — legacy stable-identity/owner/Area facts;
- `wiki/database/tables/document_revisions.md` — current technical autosave revision table;
- `wiki/database/tables/editor_sessions.md` — current authoring lease;
- `wiki/database/tables/approval_instances.md`, `approval_stage_instances.md`, `approval_signoffs.md` — legacy Approval shapes/vocabularies;
- `wiki/database/tables/audit_events.md` — current append-only Audit enforcement evidence;
- `wiki/database/tables/outbox_events.md` — current durable-intent/retry evidence;
- `wiki/decisions/0027-rls-adoption-sequencing.md` — current RLS/role/GUC evidence, including false-green history under owner/bypass roles;
- `wiki/database/migration-policy.md` — current bootstrap/baseline/grants mechanics.

Do not infer target table names or retain legacy tables merely because they exist.

---

# 10. Structural Inversion / Global Maximum / YAGNI self-challenge

Candidate self-challenge before independent review:

- If the repo currently used one schema instead of two, one canonical product schema would still be the smallest sufficient structure; PASS.
- If current IDs were all globally unique single-column PKs, same-tenant FK enforcement would still be valuable because tenant ownership is frozen; candidate composite identity does not depend on legacy shape.
- If current code had no RLS, tenant-isolation defense-in-depth would still be frozen; PASS.
- If current code already used schema-per-context, no independent deployment/trust boundary would justify retaining eleven schemas; inversion still rejects it.
- Cross-owner FK is mechanism, not authority; semantic owner remains unchanged; PASS.
- One runtime DB role rather than one per bounded context avoids speculative infrastructure and still preserves FORCE-RLS enforcement; YAGNI PASS.
- `READ COMMITTED` plus invariant-specific locks/constraints avoids a global concurrency abstraction; YAGNI PASS.
- No ORM, event-sourcing framework, distributed transaction coordinator, generic relation registry or policy language is introduced.
- Tenant-by-tenant system work preserves operability without a global content bypass.

Self-review is not independent review. Fable must attempt to falsify these claims.

---

# 11. Mandatory adversarial questions

Attack at minimum:

1. Is one canonical `metaldocs` schema a true Global Maximum or just aesthetic cleanup?
2. Does composite `(tenant_id,id)` create accidental complexity larger than the cross-tenant-reference class it removes?
3. Is `UNIQUE(id)` globally required for any real caller/provider/integration, or would it be redundant?
4. Are there frozen cross-owner references where an FK would improperly couple lifecycle, migration order or erasure?
5. Is banning cross-owner CASCADE sufficient, or are other FK actions dangerous?
6. Can a valid tenant-owned relationship still cross tenants under any candidate path?
7. Are there legitimate product-global facts that the tenant-key law would wrongly force tenant-scoped?
8. Does fail-closed RLS make API/jobs/maintenance/restore operationally impossible or merely require explicit tenant execution context?
9. Is tenant-by-tenant system execution sufficient for release, periodic review, retention, erasure and reconciliation, or is a bounded global-content capability genuinely required?
10. Does the DB-role law actually guarantee RLS is exercised in ordinary runtime paths?
11. Could RLS accidentally become Authorization through future query convenience?
12. Can cross-owner local transaction participation be achieved without one owner importing another owner's repository?
13. Does requiring mandatory Audit/outbox intent same-commit overconstrain operations whose external effect is intentionally multi-step?
14. Are any frozen immutable facts impossible to protect under the proposed classification/enforcement approach?
15. Is `TEXT + CHECK` preferable to PostgreSQL ENUM for the real frozen vocabularies?
16. Is `BYTEA(32 semantics)` the correct canonical SHA-256 storage rule or does it harm real interoperability without a compensating invariant?
17. Does the candidate accidentally prejudge B2–B6 table design?
18. Does it leave any essential substrate decision deferred too late?
19. Does it introduce any abstraction only because another abstraction exists?
20. What is the strongest realistic path by which this substrate would make MetalDocs operationally fragile despite looking architecturally clean?

---

# 12. Proof obligations before promotion

R10-B1 may close only if independent review and adjudication establish:

1. every tenant-owned reference form can structurally prove same-Tenant identity where existence is required;
2. no cross-owner FK can cascade-delete another owner's durable/history state;
3. target product-state namespace no longer depends on the historical `public`/`metaldocs` split;
4. business/provider/external identifiers cannot become technical persistence identity by convenience;
5. missing tenant execution context fails closed on tenant-owned product data;
6. RLS contains no canonical Authorization semantics;
7. ordinary runtime role cannot bypass RLS through ownership/superuser/BYPASSRLS posture;
8. system/background jobs remain operational without implicit all-tenant content visibility;
9. cross-owner atomic use cases can share one local DB transaction through application/composition seams without repository ownership inversion;
10. mandatory Audit/durable-intent rows can participate in that same commit without making Audit/outbox domain authorities;
11. immutable fact families have a falsifiable enforcement strategy slot for B2–B5;
12. no generic polymorphic domain relationship table/Record/object platform has been introduced;
13. B2–B6 can derive their models under these laws without material contradiction;
14. R10-A reopen set remains empty unless strict material evidence proves otherwise.

---

# 13. Candidate outcome

```text
OUTCOME: RESTRUCTURE TARGET SUBSTRATE

one PostgreSQL DB                    = YES
canonical product-state schema       = metaldocs
schema-per-BC                         = NO
schema-per-Tenant                     = NO

tenant-owned PK                       = (tenant_id, id)
technical id                          = UUID
business/provider id as PK            = NO
same-tenant live reference            = composite FK when existence required
cross-owner FK                        = ALLOWED, authority-neutral
cross-owner ON DELETE CASCADE         = FORBIDDEN
generic polymorphic domain FK         = FORBIDDEN

business timestamp                    = TIMESTAMPTZ
SHA-256 canonical persistence         = BYTEA + 32-byte constraint
frozen vocabulary default             = TEXT + CHECK
real unknown/absence                  = NULL
JSONB                                 = bounded snapshot/variable provenance only

RLS                                   = tenant isolation only
FORCE RLS on tenant product tables    = YES
missing tenant context                = FAIL CLOSED
ordinary runtime DB role              = non-owner / NOSUPERUSER / NOBYPASSRLS
system/background content access      = explicit tenant-by-tenant context

default isolation                     = READ COMMITTED
cross-owner atomicity                 = one local PostgreSQL transaction
composition transaction               = coordination mechanism, no authority
mandatory Audit append                = same commit where frozen-required
mandatory durable async intent        = same commit where required
worker/retry/external execution        = deferred R10-D
```

R10-B1 remains unapproved until the independent review is adjudicated and promoted deliberately into `wiki/architecture/r10-technical-architecture.md`.

---

# 14. Required Fable output

Return one verdict:

```text
APPROVE R10-B1
APPROVE R10-B1 WITH MATERIAL FIXES
DO NOT APPROVE R10-B1
```

Then provide:

1. concise independent summary;
2. BLOCKER / MAJOR / LOW findings with evidence and affected invariant;
3. explicit adjudication recommendation for each finding;
4. better alternative substrate if rejecting a candidate decision;
5. Structural Inversion result;
6. Global Maximum vs local maximum analysis;
7. YAGNI/subtractive pass;
8. operational-viability assessment for real API/jobs/tenant lifecycle execution;
9. R10-A/R9.5 reopen assessment under their strict contracts;
10. exact promotion conditions.

Do not edit target authority as part of review output. Findings are evidence; the primary architect/operator adjudicates them under the Method.
