# R10-B1 Relational Substrate — Fable Adjudication and Corrected Target

> **Status:** ADJUDICATED CORRECTED CANDIDATE — PENDING BOUNDED DELTA CHECK — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Candidate packet:** `docs/superpowers/analysis/2026-08-17-r10-b1-relational-substrate-fable-review-request.md` @ `a3bb4ac8`
> **Independent review:** `docs/superpowers/analysis/2026-08-17-r10-b1-independent-fable-review.md` @ `b38f598b`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Authority note:** this artifact records adjudication and the corrected R10-B1 candidate. It does not amend `wiki/architecture/r10-technical-architecture.md`, close R10-B1, open R10-B2, authorize schema changes, or authorize product implementation.
> **Implementation gate:** **CLOSED.**

---

## 1. Review result and adjudication

Independent verdict:

```text
APPROVE R10-B1 WITH MATERIAL FIXES
BLOCKER = 0
MAJOR   = F1, F2
LOW     = F3, F4, F5, F6
```

Adjudication:

| Finding | Decision | Corrected outcome |
|---|---|---|
| F1 — tenant-isolation scope / durable-intent ambiguity | **ACCEPT / ROOT-CAUSE RESTRUCTURE** | Replace the one-dimensional §6.14 taxonomy with orthogonal semantic-persistence and mutation classifications. Durable async intent is `DURABLE MECHANISM`, never mislabeled `EPHEMERAL`. Tenant-isolation law is bound to semantic class. |
| F2 — `SET NULL` / `SET DEFAULT` reachable cross-owner | **ACCEPT** | Cross-owner referential actions are a closed allowed set: `RESTRICT` / `NO ACTION` only. All other actions are forbidden. |
| F3 — FORCE RLS and platform maintenance posture | **ACCEPT / DEFER WITH NAMED SUCCESSOR** | Serving roles never bypass RLS. True cross-tenant migration/backfill/restore may use a separate non-serving maintenance principal or per-tenant execution; concrete mechanism belongs R10-F/implementation/ops. |
| F4 — due-work discovery source unnamed | **ACCEPT** | Exactly two lawful shapes: per-Tenant iteration or tenant-written platform routing/due intent. No RLS exemption on tenant semantic tables. |
| F5 — supporting-owner state missing from B2–B6 map | **ACCEPT** | Complete the R10-B decomposition map below. |
| F6 — target `metaldocs` name overlaps legacy namespace | **ACCEPT / ROUTE TO R10-F** | Final target schema remains `metaldocs`; namespace does not prove target provenance during transition. R10-F must carry explicit target-vs-legacy cutover identification/choreography. |

No finding reopens R9.5 or R10-A.

---

# 2. Corrected R10-B1 target

## 2.1 PostgreSQL topology

```text
one PostgreSQL database
one canonical MetalDocs product-state schema: metaldocs
```

Target product/business/support state does not use `public` as a product namespace. No schema-per-Tenant and no schema-per-bounded-context. PostgreSQL namespace is mechanism, never semantic authority.

The name `metaldocs` is the final target namespace even though legacy tables already occupy it today. R10-F owns the explicit migration-time distinction between legacy and target objects; namespace alone is never provenance evidence during cutover.

## 2.2 Tenant-owned identity law

For each durable tenant-owned entity:

```text
tenant_id UUID NOT NULL
id        UUID NOT NULL
PRIMARY KEY (tenant_id, id)
```

`id` is opaque technical identity. No second global `UNIQUE(id)` is required absent a real consumer.

Business/provider/external identity never becomes the technical PK:

```text
Document.code      != PK
REV label          != PK
Dossier stable key != PK
Artifact hash      != PK
ExternalReference != PK
provider key/URL   != PK
```

`Tenant` is the root and therefore is not tenant-owned by another Tenant. Genuinely global/product/credential facts may omit `tenant_id` only when their semantics require it; exact cases are derived in their owning later block.

## 2.3 Same-Tenant live reference law

When relational existence is required between tenant-owned rows:

```text
FOREIGN KEY (tenant_id, target_id)
REFERENCES target_table(tenant_id, id)
```

A cross-owner FK proves only existence, same-Tenant identity and target identity. It never transfers business/lifecycle authority.

## 2.4 Cross-owner referential-action law — closed allowed set

Across semantic owners, the only permitted FK actions are:

```text
ON DELETE RESTRICT
ON DELETE NO ACTION
ON UPDATE RESTRICT
ON UPDATE NO ACTION
```

Every other referential action is forbidden across owners:

```text
CASCADE
SET NULL
SET DEFAULT
```

No owner may delete **or mutate** another owner's durable state through an FK side effect. Required multi-owner changes are explicit coordinated use cases through the proper authorities.

Within one semantic owner, cascade remains non-default and may be used only for a strictly subordinate child with no independent historical meaning whose legal lifecycle always ends with the parent.

## 2.5 Historical snapshot law

When a fact means “preserve what was true at that moment”, store an immutable snapshot. A source reference may remain for correlation when justified, but later source mutation never rewrites the snapshot.

## 2.6 External/provenance reference law

External identifiers are typed provenance/correlation only. They are not internal technical identity and never database FKs to external systems.

## 2.7 No universal polymorphic business relation

Do not introduce generic domain persistence such as `resource_type/resource_id`, `subject_type/subject_id`, a generic Record/Object table or a universal attachment registry merely to reduce typed tables.

Audit may keep generic resource attribution because it is explicitly non-authoritative for resource state. Artifact no-confirmed-orphan representation remains deliberately deferred to B3/B5.

## 2.8 Primitive persistence defaults

```text
technical IDs      = UUID
business instants  = TIMESTAMPTZ
canonical SHA-256  = BYTEA with octet_length(hash)=32
frozen vocabulary  = TEXT + CHECK by default
real unknown/absence = NULL
```

Do not use zero UUIDs, empty strings, zero numeric sentinels or fabricated `UNKNOWN` values to erase legitimate uncertainty. Historical Migration preserves unknown as unknown.

`JSONB` is allowed for bounded whole snapshots or genuinely variable provider-neutral provenance where the atomic semantic is the document/payload itself; it is not a default escape hatch for unmodeled business state.

PostgreSQL ENUM is not the default.

---

# 3. Orthogonal persistence classification law

The original candidate mixed authority, durability and mutability in one taxonomy. The corrected law separates them.

Every persisted fact/table receives a **semantic persistence class**:

```text
SEMANTIC AUTHORITY
ATTRIBUTED SUPPORT
DURABLE MECHANISM
EPHEMERAL MECHANISM
REBUILDABLE PROJECTION
```

Meaning:

- `SEMANTIC AUTHORITY` — canonical business/supporting-semantic truth owned by one R10-A semantic owner.
- `ATTRIBUTED SUPPORT` — durable product-facing state with a named non-business owner, such as Notifications delivery/inbox/read state.
- `DURABLE MECHANISM` — durable machinery required to carry/process an already-owned semantic intent, such as an async dispatch intent or durable effect-attempt routing record.
- `EPHEMERAL MECHANISM` — transient machinery such as leases/claims/heartbeats whose loss can be recovered without rewriting canonical truth.
- `REBUILDABLE PROJECTION` — disposable derived state such as Search indexes.

Independently, every persisted fact/table receives a **mutation law** appropriate to its semantics:

```text
MUTABLE
IMMUTABLE / APPEND-ONLY
TERMINAL / TOMBSTONED
EXPLICIT STATE MACHINE
REBUILDABLE
```

Examples:

| State | Semantic class | Mutation law |
|---|---|---|
| RevisionSubmission | SEMANTIC AUTHORITY | IMMUTABLE |
| ApprovalDecision | SEMANTIC AUTHORITY | IMMUTABLE |
| AuditEvent | SEMANTIC AUTHORITY | APPEND-ONLY |
| Notification inbox/read state | ATTRIBUTED SUPPORT | MUTABLE / explicit state |
| required outbox/async intent | DURABLE MECHANISM | EXPLICIT STATE MACHINE |
| worker lease | EPHEMERAL MECHANISM | MUTABLE/expiring |
| Search index | REBUILDABLE PROJECTION | REBUILDABLE |

An immutable semantic fact may not rely only on application convention. Its owning block must choose a falsifiable enforcement mechanism proportional to the fact.

---

# 4. Tenant-isolation law bound to semantic class

## 4.1 Tenant-owned semantic authority

Tenant-owned `SEMANTIC AUTHORITY` tables require the full stack:

```text
application/repository tenant predicate
+ same-Tenant relational key/FK law
+ ENABLE ROW LEVEL SECURITY
+ FORCE ROW LEVEL SECURITY
+ explicit transaction-local Tenant context
+ missing Tenant context = FAIL CLOSED
```

## 4.2 Tenant-owned attributed support

Tenant-owned `ATTRIBUTED SUPPORT` state also receives the full tenant-isolation stack unless a later owning block proves a narrower representation that preserves the same isolation claim.

Non-business does not mean globally visible.

## 4.3 Durable / ephemeral mechanisms

`DURABLE MECHANISM` and `EPHEMERAL MECHANISM` tables do **not** inherit semantic-table RLS policy mechanically. Their owning block must state an explicit isolation posture appropriate to the mechanism.

A globally claimable mechanism surface may expose only routing/mechanism facts required for discovery/claim, such as:

```text
intent/claim id
tenant_id
intent kind
due/routing time
lease/claim state
opaque target references
```

It must not expose arbitrary Tenant business content merely to make global dispatch convenient.

## 4.4 Rebuildable projections

Projection storage receives an explicit isolation/access posture from R10-D/E. Projection visibility can never substitute for canonical Authorization, and a projection must not become a tenant-content bypass merely because it is rebuildable.

## 4.5 RLS scope

RLS remains Tenant isolation only. It must not encode Role, Area, Dossier, Approval participant, Document owner/responsibility or any other canonical Authorization predicate.

---

# 5. Durable async intent law

When a tenant business mutation requires future async work, the durable intent is a `DURABLE MECHANISM` fact written in the same local transaction as the business mutation.

Conceptually:

```text
BEGIN tenant-scoped transaction
  business facts
  required Audit append
  required durable routing/async intent
COMMIT
```

The intent carries explicit Tenant attribution and enough routing metadata to be claimable without reading arbitrary tenant business content.

After global/platform claim:

```text
claim routing fact
→ obtain tenant_id + opaque target/routing identity
→ BEGIN ordinary tenant-scoped transaction
→ seed Tenant/system execution context
→ load canonical tenant content under normal isolation/AuthZ rules
→ execute owner/application use case
→ COMMIT/ROLLBACK
```

R10-D owns claim/retry/DLQ/external-effect execution. It may not satisfy those requirements by granting an ordinary worker implicit all-Tenant content visibility.

If a future intent needs sensitive or immutable payload beyond globally claimable routing metadata, R10-D must preserve both properties — global mechanism claimability and tenant-content isolation — without weakening semantic-table isolation.

---

# 6. Database roles and sanctioned maintenance posture

## 6.1 Ordinary serving runtime

API, normal workers and normal jobs use DML roles that are:

```text
NOSUPERUSER
NOBYPASSRLS
not owners of tenant semantic/support tables
```

DDL/object ownership is separate from serving runtime DML.

No database role per bounded context is introduced in V1 absent a real trust/deployment boundary.

## 6.2 True cross-Tenant platform maintenance

Migration, controlled backfill and restore are a distinct non-serving trust surface.

Ordinary product operations remain tenant-by-tenant. Where a true cross-Tenant maintenance operation cannot reasonably be expressed tenant-by-tenant, R10-F/implementation/ops may use an explicit non-serving maintenance principal with the minimum required database bypass privilege.

Constraints:

```text
maintenance principal != API serving role
maintenance principal != ordinary worker/job serving role
maintenance principal not reachable from ordinary request path
BYPASSRLS never implies product Authorization
credentials/process/rotation/audit are explicit ops concerns
```

R10-B1 does not choose the concrete maintenance credential or workflow; it names the sanctioned home and preserves the trust boundary.

---

# 7. Global discovery law

A background/system actor touching tenant semantic/support state does not gain all-Tenant content access merely because it is automated.

There are exactly two lawful shapes for discovering due work whose business truth is tenant-owned:

### A. Per-Tenant iteration

```text
discover Tenant IDs from root/Organization platform surface
→ for each Tenant
   BEGIN tenant-scoped transaction
   seed tenant/system execution context
   query/execute due owner work
   COMMIT/ROLLBACK
```

### B. Tenant-written durable routing intent

```text
tenant business transition
→ same-commit durable platform routing/due intent
→ global dispatcher claims routing metadata only
→ re-enter ordinary tenant-scoped execution
```

A third path — disabling/failing-open tenant isolation on semantic tables for a scheduler — is not a legal target architecture.

---

# 8. Transaction and isolation law

Every ordinary tenant-owned mutation runs in an explicit local PostgreSQL transaction.

Single-owner mutation: the owner application service owns its transaction boundary.

Cross-owner atomic use case:

```text
composition opens one local PostgreSQL transaction
→ invokes owner-specific application seams
→ all required semantic/support/Audit/intent writes participate
→ one COMMIT or one ROLLBACK
```

No semantic owner hides an independent nested commit during a cross-owner atomic operation. Atomicity is not obtained by importing another owner's repository.

Default isolation remains `READ COMMITTED`. Later blocks use the narrowest sufficient mechanism (`UNIQUE`, partial `UNIQUE`, `CHECK`, FK, CAS/atomic UPDATE, `working_version`, `SELECT ... FOR UPDATE`). A stronger isolation level requires a demonstrated failure class.

---

# 9. R10-B decomposition map — corrected

```text
R10-B1
  relational substrate / tenancy / references / transaction law

R10-B2
  Authentication
  Organization
  Authorization

R10-B3
  Artifact relational core for the first real consumer
  Controlled Information
  WorkingContent
  RevisionSubmission

R10-B4
  Approval
  Rendition/Release relational state
  Distribution

R10-B5
  Documentary Context
  Records Governance
  Artifact second-consumer / no-confirmed-orphan closure

R10-B6
  Audit relational state
  Interchange batch/plan/outcome persistent state
  cross-owner transaction matrix
  imported-history closure
  global DB coherence

R10-D
  Notifications attributed-support persistence details
  Search projection persistence details
  durable/ephemeral async mechanism tables
  claim/retry/DLQ/external-effect execution
```

Artifact starts in B3 because Controlled Information is the first semantic consumer. B5 introduces Evidence as the second semantic consumer and must close the no-confirmed-orphan structural invariant without a generic attachment registry.

---

# 10. R10-F namespace/cutover constraint

Final target namespace remains `metaldocs`.

Because legacy current state already contains `metaldocs.*`, R10-F must use an explicit target-vs-legacy manifest/choreography during migration. A table being located in `metaldocs` during transition is not proof that it is already a target table.

Migration inconvenience or current namespace occupancy is not R10-B1 reopen evidence.

---

# 11. Corrected proof obligations

R10-B1 may be promoted only if the bounded delta check establishes:

1. every tenant-owned live reference can structurally prove same-Tenant identity where relational existence is required;
2. across semantic owners, only `RESTRICT` / `NO ACTION` FK actions are legal; no FK side effect can delete **or mutate** another owner's durable state;
3. target product namespace is one canonical `metaldocs` schema and does not preserve the `public`/`metaldocs` split as target architecture;
4. business/provider/external identifiers cannot become technical persistence identity by convenience;
5. semantic persistence class and mutation law are independent dimensions and every later durable fact must declare both when applicable;
6. missing Tenant context fails closed for tenant-owned semantic authority and attributed support;
7. RLS contains no canonical Authorization semantics;
8. ordinary serving runtime roles are non-owner, `NOSUPERUSER`, `NOBYPASSRLS` for the protected state;
9. durable async intent remains same-commit and globally claimable through routing/mechanism metadata without granting global tenant-content visibility;
10. system/background work has only the two lawful discovery shapes in §7;
11. true cross-Tenant maintenance is separated from serving runtime and routed to R10-F/implementation/ops without weakening ordinary runtime isolation;
12. cross-owner atomic use cases can share one local PostgreSQL transaction through application/composition seams without repository ownership inversion;
13. mandatory Audit and durable-intent writes can roll back with the business mutation;
14. supporting-owner persistent state has an explicit R10-B owner block and Notifications/Search/async persistence is explicitly routed to R10-D;
15. no generic polymorphic domain relationship/Record/Object platform is introduced;
16. R9.5 and R10-A reopen sets remain empty absent strict material counterevidence.

Later implementation proof must include at minimum:

```text
fail-closed negative RLS proof using a non-owner NOBYPASSRLS serving-class role
cross-owner FK-action census
same-commit business + Audit + required-intent rollback proof
serving-pool role posture proof
```

These are proof slots, not implementation authorized by this artifact.

---

# 12. Corrected candidate outcome

```text
OUTCOME: RESTRUCTURE TARGET SUBSTRATE

PostgreSQL DB                         = one
canonical target product schema      = metaldocs
schema-per-BC                         = no
schema-per-Tenant                     = no

tenant-owned entity PK               = (tenant_id, id)
technical id                          = UUID
global UNIQUE(id) default             = no
business/provider/external id as PK   = no
same-Tenant live reference            = composite FK when existence required
cross-owner FK                        = allowed, authority-neutral
cross-owner FK actions                = RESTRICT / NO ACTION only
universal polymorphic business FK     = forbidden

business timestamp                    = TIMESTAMPTZ
SHA-256 canonical storage             = BYTEA + 32-byte constraint
frozen vocabulary default             = TEXT + CHECK
real unknown                          = NULL
JSONB                                 = bounded snapshot/variable provenance only

semantic persistence classes:
  SEMANTIC AUTHORITY
  ATTRIBUTED SUPPORT
  DURABLE MECHANISM
  EPHEMERAL MECHANISM
  REBUILDABLE PROJECTION
mutation law                          = independent second dimension

semantic/support tenant isolation     = app predicate + composite keys/FKs + ENABLE/FORCE RLS
missing Tenant context                = fail closed on protected semantic/support state
RLS                                   = tenant isolation only
ordinary serving DB role              = non-owner / NOSUPERUSER / NOBYPASSRLS
true bulk maintenance                 = distinct non-serving trust surface; successor R10-F/ops

system/background work                = tenant-by-tenant or platform routing-intent discovery
implicit all-Tenant content access    = forbidden

default transaction isolation         = READ COMMITTED
cross-owner atomicity                 = one local PostgreSQL transaction
mandatory Audit append                = same commit where required
mandatory durable async intent        = same commit where required
durable intent class                  = DURABLE MECHANISM
claim/retry/DLQ/external execution    = R10-D

R9.5 reopen                           = EMPTY
R10-A reopen                          = EMPTY
R10-B2                                = BLOCKED UNTIL B1 PROMOTION
implementation                        = BLOCKED
```

---

# 13. Bounded delta-check request

The independent challenger should **not** re-run a broad substrate review unless the corrections introduce a new material contradiction.

Check only:

1. F1 is closed by the orthogonal semantic-class + mutation-law model and does not create a hidden new semantic owner;
2. durable async intent remains operationally claimable without global tenant-content access;
3. F2 is closed by the positive closed FK-action set;
4. F3/F4 preserve both fail-closed serving runtime and real maintenance/job operability;
5. F5 assignment covers all R10-A supporting owners without pulling R10-C/D/E/F mechanisms prematurely into R10-B;
6. F6 is correctly routed to R10-F without weakening the canonical `metaldocs` target namespace;
7. all corrected proof obligations are falsifiable and no R9.5/R10-A reopen is required.

Required verdict:

```text
APPROVE R10-B1 CORRECTED TARGET
APPROVE R10-B1 CORRECTED TARGET WITH MATERIAL FIXES
DO NOT APPROVE R10-B1 CORRECTED TARGET
```

Findings remain evidence. Promotion into `wiki/architecture/r10-technical-architecture.md` occurs only after operator adjudication of the bounded delta result.
