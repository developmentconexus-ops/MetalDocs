# R10-T8D — Persistence Realization — Adjudicated Corrected Candidate

```text
ADJUDICATED CORRECTED CANDIDATE
NON-AUTHORITATIVE STAGING
BOUNDED ROUND-2 REVIEW INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Round-1 review HEAD:** `9f83085018643231bedc2d91314dc56bc2a85537`  
> **Original candidate:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`  
> **Independent review:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md`  
> **Stage:** T8-D ACTIVE  
> **Implementation:** BLOCKED

This file is the Lead-adjudicated corrected T8-D candidate for bounded Round 2.

It is intentionally a **correction overlay over the exact original candidate reviewed in Round 1** rather than a second 2,000+ line restatement. For Round-2 review, the effective corrected candidate is:

```text
original T8-D Global Maximum candidate
+
this adjudicated correction overlay
```

Where this file conflicts with the original candidate, **this file controls the corrected staging input**. Unchanged original sections remain unchanged. Neither file is target authority.

The Round-1 independent review confirmed the Global Maximum class and found no T8-C/T8-B/T1→T7 reopen and no T8-E/F/G/T10/T11 trespass. The operator explicitly approved the Lead adjudication below before this artifact was materialized.

---

## 1. Global Maximum — unchanged

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

Round-1 Fable independently confirmed this class. No replacement Global Maximum survived.

---

## 2. Round-1 Lead adjudication summary

```text
BLOCKER-1  River / PG16 / DB ownership             ACCEPT
BLOCKER-2  attach-vs-GC race                       ACCEPT

M1  Decision -> frozen candidate snapshot          ACCEPT
M2  mutable malware proof                          ACCEPT
M3  provider-subject historical uniqueness         ACCEPT WITH NARROWING
M4  actor/target User deadlock                      ACCEPT
M5  transaction census incomplete                  ACCEPT
M6  duplicate GovernanceAttempt                    ACCEPT
M7  singleton Company vs company_id                REJECT
M8  Go <-> DDL vocabulary drift                    ACCEPT
M9  idempotency retention ordering                 ACCEPT
M10 fingerprint derived PII                        ACCEPT
M11 DB trust classes / proof role                  ACCEPT
```

Accepted LOW corrections are folded into §§16-17. One LOW (`audit.events.resource_id`) was rejected because T3 requires stable resource identity for semantic Audit; restore/session operational reconciliation need not manufacture a semantic Audit event without a resource.

---

# PART I — BLOCKER corrections

## 3. B1 correction — River `river.*`, PG16 and DB trust ownership

The original candidate correctly selected:

```text
River v0.37.1
+ custom PostgreSQL schema `river`
+ database/sql-family InsertTx in the same semantic transaction
+ no first-party raw SQL against River tables
```

Round 1 verified those claims against the exact pinned dependency.

The original candidate was incomplete about ownership and River's default self-REINDEX behavior on the PostgreSQL 16 feature floor.

### 3.1 Target DB trust classes

T8-D now explicitly recognizes four database trust identities/classes:

```text
bootstrap / provisioner
  privileged installation-only identity
  creates roles/schemas when required
  transfers ownership
  never used for serving

metaldocs_owner
  NOLOGIN
  owns first-party schemas/tables and river schema/objects after provisioning
  migration / DDL owner
  never serving

metaldocs_runtime
  serving connection role
  receives only required USAGE/DML/sequence/function privileges
  never receives owner membership
  never SET ROLEs into metaldocs_owner

metaldocs_ci
  non-owner serving-equivalent proof role
  used to prove runtime privilege restrictions without false-green owner/superuser bypass
```

Exact credential provisioning/process placement remains T8-G/T10; the trust classes and privilege relationship are T8-D.

### 3.2 River schema provisioning

```text
bootstrap/provisioner
  creates schema river if absent
  assigns ownership to metaldocs_owner

River migrations
  run targeting schema `river`
  using an owner-capable migration path

metaldocs_runtime
  does NOT own river.*
  receives only privileges required by River runtime
```

`river.*` enters the closed DB-object catalog as:

```text
class = THIRD_PARTY_MANAGED
mechanism = River
schema = river
first-party raw SQL = forbidden
```

The catalog is complete at schema/object-class level without pretending MetalDocs owns River's individual internal table semantics.

### 3.3 PostgreSQL 16 consequence

PostgreSQL 16 remains the persistence feature floor.

The corrected candidate records the material consequence found by Round 1:

```text
PG16 does not provide the later grantable MAINTAIN privilege needed to let a non-owner
perform River's default index REINDEX maintenance without broadening ownership.
```

Therefore the selected Launch posture is:

```text
River self-REINDEX maintenance = DISABLED
```

This is a T8-D privilege/correctness decision. The exact River v0.37.1 configuration field/wiring that disables the reindexer is a T8-G implementation/runtime detail and must use the pinned-version supported mechanism.

No serving role is made owner merely to support optional background maintenance.

If later operating evidence proves River index maintenance is materially needed, activate the smallest explicit operations path appropriate to the then-current PostgreSQL/River versions; do not silently grant owner powers to serving.

### 3.4 D04/D05/D26/D30/D31 corrected meaning

```text
D04  PostgreSQL-16-compatible floor remains SELECTED; record no-MAINTAIN consequence.
D05  complete first-party/technical/third-party DB-object classification; river.* is THIRD_PARTY_MANAGED.
D26  River remains under river.*.
D30  serving runtime remains distinct from owner; provisioner/CI proof roles explicitly classified.
D31  immutable-history grant proofs must execute from a non-owner serving-equivalent identity.
```

BLOCKER-1 is closed by this selected correction without changing the Global Maximum class.

---

## 4. B2 correction — universal attach-side ManagedContent lock

The original candidate did not name the row lock held by semantic attachment transactions and therefore admitted a READ COMMITTED race in which GC could authorize physical delete before a semantic reference committed.

The corrected invariant is universal:

> **Every local semantic transaction that creates or replaces a semantic `managed_content_id` reference MUST acquire and hold a `FOR SHARE` lock on that exact `platform.managed_content` row as part of the final READY proof until the semantic transaction commits or rolls back.**

Conceptual order for every attachment path:

```text
BEGIN semantic Scope
→ SELECT exact platform.managed_content row FOR SHARE
→ prove state = READY
→ prove create-once/descriptor facts
→ prove applicable AdmissionClaim binding
→ prove applicable immutable malware inspection evidence
→ write/replace semantic managed_content_id reference
→ consume AdmissionClaim in the same Scope when that path uses one
→ required owner/Audit/River work
→ COMMIT
```

The rule applies to every current path that creates a semantic managed-content reference, including as applicable:

```text
Document / next-Revision seed final attachment
DRAFT source replacement
Submission admission
OfficialRendition admission
other T4 semantic-reference creation already ratified for Launch
```

AdmissionClaim is **not made universal** merely to close this race. It retains its T4/T8-C role for in-flight authorization/liveness where an admission binding exists.

GC remains:

```text
phase 1 managed_content FOR UPDATE
phase 2 managed_content FOR UPDATE
```

`FOR SHARE` attach and `FOR UPDATE` GC therefore serialize on the same technical root.

If GC wins first, the attach lock/re-read observes `GC_PENDING` and fails closed. If attach wins first, GC waits and then the repeated semantic-reference proof sees the committed reference and aborts deletion.

D21/D22/D37 are corrected accordingly.

---

# PART II — structural corrections

## 5. M1 — GovernanceDecision must reference the frozen candidate snapshot

Add the structural constraint:

```text
controlled_docs.governance_decisions
  (step_id, actor_user_id)

FOREIGN KEY
  -> controlled_docs.governance_step_candidates(step_id, user_id)
```

`governance_step_candidates` remains the exact frozen active-candidate authority for NAMED_USER and GROUP steps.

Consequences become structural:

```text
decider must have been in the exact Step activation snapshot
empty candidate snapshot -> no GovernanceDecision is committable
later GroupMembership drift cannot rewrite candidate eligibility
current T3 Authorization still rechecks whether the frozen candidate may act now
```

The FK proves frozen-candidate membership; Authorization still proves current permission/scope/domain authority. These meanings remain separate.

---

## 6. M2 — immutable malware inspection evidence

The mutable malware verdict columns in `platform.managed_content` are removed from the corrected target.

ManagedContent keeps only mechanism lifecycle and exact technical descriptor facts needed for OPEN/READY/GC:

```text
platform.managed_content
  id
  state OPEN | READY | GC_PENDING
  provider_locator
  trust_class
  sha256 / size_bytes / content_format when READY
  created_at / ready_at / gc_pending_at
```

Add one bounded mechanism evidence relation:

```text
platform.malware_inspections

id                  UUID PK
managed_content_id  UUID NOT NULL FK -> platform.managed_content(id)
digest              BYTEA NOT NULL CHECK(octet_length(digest)=32)
verdict             TEXT NOT NULL CHECK(verdict IN ('CLEAN','MALICIOUS'))
inspected_at         TIMESTAMPTZ NOT NULL

UNIQUE(managed_content_id, digest)
```

Runtime privilege:

```text
SELECT + INSERT
NO UPDATE
NO DELETE
NO TRUNCATE
```

There is no business MalwareScan aggregate, scan workflow or generic security-event owner.

For `UNTRUSTED_EXTERNAL` immutable governed admission:

```text
one immutable inspection row for exact handle/digest must exist
verdict = CLEAN
inspection.digest = semantic exact SHA-256
```

Scanner unavailable/incomplete persists no successful proof and fails closed. A MALICIOUS verdict on the exact create-once handle/digest makes that handle non-admissible; correcting an erroneous verdict requires a new managed-content handle rather than rewriting security evidence.

Trusted managed copy/internal derivation retain the T4 policy that may not require a scan.

---

## 7. M3 — ProviderSubjectBinding uniqueness narrowed without a new identity subsystem

Reject the Round-1 suggestion to add a second `provider_subject_identities` semantic/mechanism table. It would introduce an unnecessary persistent concept and could itself permanently preserve an erroneous first association.

Correct `authn.provider_subject_bindings` uniqueness to current truth:

```text
UNIQUE(issuer, subject)
WHERE replaced_at IS NULL

UNIQUE(user_id)
WHERE replaced_at IS NULL
```

Historical binding rows remain immutable except for the one-time transition that sets `replaced_at` when they stop being current.

The history can truthfully record:

```text
User A was bound to issuer/subject S during [t1,t2)
User B later became the current binding for S at t3
```

Current simultaneous double binding remains structurally impossible.

Replacement/no-op/version laws from the original candidate remain:

```text
expected VersionToken required
exact already-current repeat -> no version advance / no duplicate Audit
material replacement -> old current row ends; new current row starts; sessions invalidated; Audit
```

This correction closes ordinary admin mis-binding recovery without creating a subject-registry subsystem.

---

## 8. M4 — User lock ordering must consider actor and target as one lock class

The original per-operation wording that implied "actor first, target second" is replaced.

For any transaction requiring more than one `org.users` row lock:

```text
collect all required User IDs
→ deduplicate
→ sort stable UUID ascending
→ acquire each User row exactly once in that order
→ use the strongest lock mode required for that row in this transaction
```

Examples:

```text
actor only protected eligibility   -> FOR SHARE
eligible target only               -> FOR SHARE
User being offboarded/disabled     -> update-strength / FOR UPDATE as selected owner root
same User is actor and target      -> acquire once at strongest required mode
```

This rule controls offboarding, eligibility replacement, responsible-owner assignment, direct User RoleAssignment, GroupMembership add, session issue and every other T3/T8-C operation that requires protected User eligibility.

The corrected transaction matrix must not separately prescribe a contradictory actor-before-target order.

---

## 9. M5 — complete mutation/transaction census

The original §27 transaction matrix is extended with the missing T6/T8-C operations.

At minimum the corrected matrix includes these additional families:

```text
Area create
  IdempotencyKey
  protected actor where T3 requires
  Organization insert + Audit + Replay

Group create
  IdempotencyKey
  protected actor where T3 requires
  Organization insert + Audit + Replay

RoleAssignment revoke
  protected actor
  Authorization delete/revoke current grant
  Audit
  natural DELETE idempotency; no durable Idempotency-Key

Company replacement
  protected actor
  Company expected VersionToken CAS
  Audit

Draft upload allocate / OPEN
  authorize revision/upload journey
  create ManagedContent OPEN + AdmissionClaim/binding as required
  no governed semantic attachment yet

Draft upload complete OPEN -> READY
  server reads exact stored bytes outside semantic owner mutation as required by T4
  derive exact descriptor + actual format
  structural validation
  malware inspection where policy/phase requires it
  local mechanism transaction serializes ManagedContent state and persists READY exact facts
  insert immutable malware inspection evidence when an inspection reaches a terminal verdict

Session logout/revoke
  delete current ApplicationSession
  natural idempotency
```

The existing transaction families remain, with the B2 ManagedContent `FOR SHARE` rule added to every semantic attachment path.

---

## 10. M6 — one GovernanceAttempt per governed subject

Add:

```text
UNIQUE(submission_id)
UNIQUE(obsolescence_request_id)
```

on `controlled_docs.governance_attempts`.

Because the subject columns are nullable and the XOR check already requires exactly one, ordinary PostgreSQL NULL-distinct uniqueness provides one attempt per actual governed subject without extra partial predicates.

RETURN/WITHDRAW/CANCEL terminate an attempt; resubmission creates a new Submission and therefore a new governed subject and new attempt.

---

# PART III — Lead pushback retained

## 11. M7 — singleton Company + `company_id` propagation: finding REJECTED

The Round-1 claim that `company_id` is merely dormant pooled-tenancy substrate is rejected.

`Company` is a current Launch semantic concept and current scope/integrity dimension, not only a future tenancy seam. Ratified current laws use Company for at least:

```text
same-Company User/target eligibility
Company-scoped RoleAssignment
Company-vs-Area Authorization scope
DocumentType.code uniqueness within Company
Area.code uniqueness within Company
Document.code uniqueness within Company
Audit historical Company attribution
single Company root
```

Therefore `company_id` on rows whose integrity/scope meaning is currently Company-bound is not duplicate truth merely because Launch currently permits one Company row.

Corrected posture remains:

```text
org.companies structurally contains exactly one Launch Company root
company_id remains on facts where Company is a current semantic/scoping dimension
cross-owner same-Company composite constraints may use it
RLS / tenant GUC / pooled isolation substrate remains ABSENT
```

A future pooled-tenancy promotion must explicitly reopen the singleton/isolation/runtime assumptions; current `company_id` identities are only the stable semantic seam, not dormant pooled tenancy.

No M7 subtraction is applied.

---

# PART IV — proof and privacy corrections

## 12. M8 — static Go vocabulary vs DDL CHECK parity

The original choice of closed text vocabularies remains. PostgreSQL ENUM is not introduced.

However every DDL CHECK that mirrors a static product/owner Go vocabulary becomes part of a blocking parity proof.

Current classes include at least:

```text
Role codes
Revision states
Area states
numbering scope
governance mode
representation mode
selector kind
GovernanceAttempt / Step states
GovernanceDecision outcomes
Obsolescence states
ManagedContent state / trust class
malware inspection verdict
Audit actor kind
```

Target verification law:

```text
static owner/product vocabulary set
==
corresponding DDL-accepted vocabulary set

any addition/removal/mismatch -> verification FAIL
```

T8-D freezes the proof obligation. T11 may choose generation or parity-inspection implementation, but a hand-synced unchecked second enumeration is not acceptable.

---

## 13. M9 — idempotency retention deletion order

Keep the original asymmetric FK design:

```text
replay.key_id -> key.id                     immediate FK
key.id        -> replay.key_id              DEFERRABLE INITIALLY DEFERRED
```

Do not make both FKs deferred merely for order-independence.

Freeze the technical retention order:

```text
BEGIN
DELETE expired platform.idempotency_replays
DELETE matching platform.idempotency_keys
COMMIT
```

Deleting Key first is invalid by design. Deleting Replay without deleting its Key remains uncommittable because of the deferred completion invariant.

The Round-1 PostgreSQL 16 empirical proof of claim/winner/loser behavior remains accepted evidence.

---

## 14. M10 — semantic fingerprint becomes keyed HMAC

The fingerprint must continue to cover the complete validated semantic command. Excluding erasable fields such as `email` would make materially different commands compare equal under the same Idempotency-Key and is rejected.

Correct realization:

```text
HMAC-SHA-256(
  server-held idempotency fingerprint key,
  canonical operation identity + canonical validated semantic command
)
```

Persist:

```text
semantic_fingerprint   BYTEA CHECK(octet_length(...)=32)
fingerprint_key_version bounded identifier
```

Do not persist the HMAC key in the product database.

Platform still treats fingerprint bytes as opaque equality material per T8-C.

T8-G owns concrete secret provisioning/rotation. T8-D freezes one required compatibility law:

```text
key material needed to compare fingerprints for non-expired idempotency keys
must remain available until those keys expire or are safely retired.
```

ReplaySnapshot remains independently PII-free by construction and snapshot-only.

---

## 15. M11 — proof role and provisioning identity

DB-grant verification must prove serving restrictions using a non-owner role.

Binding proof law:

```text
A privilege proof executed as bootstrap superuser or object/schema owner
DOES NOT prove metaldocs_runtime restrictions.
```

Proofs such as:

```text
runtime cannot UPDATE/DELETE/TRUNCATE AuditEvent
runtime cannot rewrite Submission
runtime cannot rewrite GovernanceDecision
runtime cannot rewrite Release
runtime cannot rewrite OfficialRendition
runtime cannot rewrite immutable malware inspection evidence
```

must run as `metaldocs_ci` or an equivalent non-owner serving-equivalent role with the same intended grants.

No additional DB-role proliferation is introduced absent a named risk reduction.

---

# PART V — accepted LOW corrections

## 16. Constraint/precision corrections

### 16.1 ReplaySnapshot size

Remove `64 KiB` as a T8-D architectural invariant for ReplaySnapshot.

The current 64 KiB value was legacy raw-response storage evidence, not a proven target replay requirement. T8-D requires ReplaySnapshot storage to be bounded and PII-free; the exact maximum follows the T8-E success-representation/replay census unless a stronger target proof appears first.

The independently proof-backed Audit bounded-facts size rule remains separate.

### 16.2 RoleAssignment uniqueness

The two subject-specific duplicate-prevention indexes are explicitly partial:

```text
UNIQUE NULLS NOT DISTINCT (company_id, user_id, role_code, area_id)
WHERE user_id IS NOT NULL

UNIQUE NULLS NOT DISTINCT (company_id, group_id, role_code, area_id)
WHERE group_id IS NOT NULL
```

### 16.3 SUBMITTED/current Submission biconditional

Add a declared structural CHECK equivalent to:

```text
(state = 'SUBMITTED') = (current_submission_id IS NOT NULL)
```

The existing composite FK still proves that the pointed Submission belongs to the same Revision.

### 16.4 Revision state / Release proof obligation

No semantic lifecycle trigger is added.

Target proof must demonstrate:

```text
no Revision reaches EFFECTIVE without exactly one Release for that Revision/winning Submission
no Revision reaches SUPERSEDED without its own prior Release and a successful successor Release transition
no Revision reaches OBSOLETE unless it was the current EFFECTIVE released Revision targeted by successful obsolescence
```

The Document lifecycle lock + owner SQL + immutable Release + structural unique barriers remain the realization.

### 16.5 GroupMembership same-Company structural integrity

Remove "where practical" ambiguity.

Where `company_id` is persisted on the relation and stable referenced identities, composite uniqueness/FKs MUST structurally prove:

```text
membership.company_id = group.company_id = user.company_id
```

The same identity-only FK law applies to other same-Company relations where the candidate already carries all required keys.

### 16.6 One-open Revision derivation precision

Retain:

```text
UNIQUE(document_id)
WHERE state IN ('DRAFT','SUBMITTED')
```

and explicitly derive it from the ratified singular current-open Revision semantics in T2/T6 rather than presenting it as an independent new Product rule.

### 16.7 Protected actor locking narrowed

Remove the blanket statement that every authenticated semantic mutation takes actor `FOR SHARE`.

Instead:

```text
actor User FOR SHARE is required exactly for the T3/T8-C operation census whose correctness
requires current enabled-User eligibility to serialize with offboarding/disable.
```

The corrected transaction matrix names the requirement per operation. Ordinary mutations that do not depend on that protected eligibility guarantee do not acquire the lock merely for symmetry.

### 16.8 Audit resource identity

Round-1 LOW suggestion to make `audit.events.resource_id` nullable is REJECTED.

T3 semantic Audit evidence requires stable resource kind/id. Operational restore/session reconciliation that lacks a semantic product resource need not fabricate a semantic AuditEvent solely to create a row.

`resource_id` therefore remains required for semantic Audit.

### 16.9 Restore mechanism-state precision

T4's session invalidation and restore non-resurrection laws remain unchanged.

Corrected persistence notes make explicit that restored technical state also converges through its own canonical rules:

```text
River pending jobs -> revalidate/idempotently resume under T5
GC_PENDING -> repeated phase-2 proof before provider delete
AdmissionClaims -> bounded expiry / live-claim rules
Idempotency -> bounded expiry + paired completion invariant
```

No new restore-state table or generic recovery journal is introduced by T8-D.

### 16.10 Explicit SQL / ORM wording

D38/D39 are clarified as the Launch persistence baseline, not a forever implementation prohibition:

```text
explicit owner-private database/sql SQL = selected Launch baseline
no generic ORM/repository framework is required or authorized by current evidence
future named benefit may reopen this implementation mechanism choice without reopening semantic ownership
```

---

# PART VI — corrected decision ledger

## 17. T8D-D01→D40 after Round-1 adjudication

Unlisted wording remains as in the original candidate. Materially corrected decisions are restated below.

```text
D01  one PostgreSQL database                                      UNCHANGED
D02  owner/mechanism schemas                                     CORRECTED: river third-party provisioning/ownership explicit
D03  fully-qualified first-party SQL                             UNCHANGED
D04  PostgreSQL-16 feature floor                                 CORRECTED: no-MAINTAIN / River self-reindex consequence explicit
D05  closed bidirectional DB-object catalog                      CORRECTED: THIRD_PARTY_MANAGED river class

D06  persist RoleAssignment only                                 CORRECTED: partial subject-specific unique indexes
D07  no Launch RLS/tenant substrate                              UNCHANGED; M7 rejected
D08  no permission/status/Search cache                           UNCHANGED

D09  explicit BIGINT VersionToken                                UNCHANGED
D10  WorkingContent generation OCC                               UNCHANGED
D11  Revision.state canonical lifecycle                          UNCHANGED
D12  Release fact + EFFECTIVE barrier                            CORRECTED: explicit state/Release proof obligation
D13  one open Revision                                           CORRECTED: upstream derivation cited
D14  bounded current_submission_id                               CORRECTED: biconditional CHECK explicit

D15  closed relational governance model                          UNCHANGED
D16  one ACTIVE Step                                             UNCHANGED
D17  live GROUP dependency vs candidate snapshot                 CORRECTED: Decision composite FK to candidate snapshot
D18  no generic workflow persistence                             UNCHANGED

D19  semantic exact descriptors                                  UNCHANGED
D20  ManagedContent mechanism state only                         CORRECTED: malware proof removed to immutable relation
D21  row-existence AdmissionClaim                                CORRECTED: not universal; attach `FOR SHARE` is universal
D22  two-phase GC                                                CORRECTED: conflicts with universal attach `FOR SHARE`

D23  paired Key + Replay                                         CORRECTED: HMAC fingerprint + key version
D24  deferred completion FK                                      CORRECTED: retention order Replay -> Key explicit
D25  no durable IN_PROGRESS/FAILED                               UNCHANGED

D26  River under river.*                                         CORRECTED: owner/provisioner grants and self-reindex OFF
D27  no first-party raw River SQL                                UNCHANGED

D28  identity/existence-only cross-owner FK                      CORRECTED: candidate Decision FK is same-owner; same-Company composite FKs explicit
D29  no cross-owner semantic cascades                            UNCHANGED
D30  serving runtime role != DDL owner                           CORRECTED: provisioner/owner/runtime/CI trust classes explicit
D31  grants enforce immutable-history classes                    CORRECTED: immutable malware evidence + non-owner proof role
D32  zero semantic lifecycle triggers baseline                   UNCHANGED

D33  protected actor FOR SHARE                                   CORRECTED: narrowed to upstream-required operation census
D34  User offboarding/eligibility root                           UNCHANGED in mode; multi-User order corrected
D35  Document FOR UPDATE lifecycle root                          UNCHANGED
D36  DocumentType configuration serialization                    UNCHANGED
D37  deterministic lock ordering                                 CORRECTED: actor/target Users one sorted lock class; ManagedContent attach FOR SHARE

D38  explicit owner-private database/sql SQL                     CORRECTED: Launch baseline with reopen on named benefit
D39  reject generic ORM/repository baseline                      CORRECTED: current baseline, not eternal prohibition
D40  normal owner-private relational views; materialized Search OFF UNCHANGED
```

Add two bounded target facts without creating new architecture classes:

```text
platform.malware_inspections = immutable technical security evidence
static-Go <-> DDL closed-vocabulary parity = blocking verification obligation
```

---

# PART VII — corrected lock/order law

## 18. Global lock classes after correction

```text
1. Idempotency scoped Key claim when operation requires it
2. org.users rows required by protected actor/target semantics
   - collect/dedupe/sort UUID ascending
   - acquire each once at strongest required mode
3. other Organization mutable roots as required
4. DocumentType/configuration root
5. Document lifecycle root
6. owner-local child/current rows/counters
7. platform.managed_content
   - semantic attach/reference write: FOR SHARE
   - GC phase 1/2: FOR UPDATE
8. owner-local append-only evidence + Audit / River inserts
```

Within every repeated same-class set, stable identifier order is deterministic.

Do not acquire the same User first as actor and later as target through a second path.

---

# PART VIII — corrected transaction-census additions

## 19. Additional matrix rows Round 2 must verify

```text
Area create
Group create
RoleAssignment revoke
Company replacement
Draft upload allocate/open
Draft upload complete OPEN->READY
Session logout/revoke
```

Round 2 must also verify the B2 `managed_content FOR SHARE` rule is present on every semantic reference-creation/replacement path in the original transaction matrix, not only on Submission.

---

# PART IX — Round-2 attack surface

## 20. Bounded Fable Round 2

Round 2 must not re-review the 26 Round-1 decisions accepted unchanged unless a corrected delta creates a new contradiction.

It must attack exactly the material corrected delta:

```text
R2-1  River custom schema ownership/provisioning/grants on PG16
      + self-REINDEX OFF
      + no runtime owner membership
      + D05 third-party catalog law

R2-2  universal semantic attach ManagedContent FOR SHARE
      vs GC FOR UPDATE
      + AdmissionClaim remains non-universal
      + all semantic reference paths covered

R2-3  GovernanceDecision (step_id,actor_user_id) FK
      -> frozen GovernanceStepCandidate
      + empty snapshot impossibility

R2-4  immutable platform.malware_inspections relation
      + exact digest binding
      + one terminal verdict per exact handle/digest
      + no business scan lifecycle

R2-5  current-only provider issuer+subject uniqueness
      + current-only User binding uniqueness
      + historical rows truthful
      + no new subject registry

R2-6  actor/target User one sorted lock class
      + no remaining architecture-induced deadlock cycle

R2-7  added transaction-census operations
      especially upload OPEN->READY

R2-8  GovernanceAttempt subject uniques

R2-9  Lead rejection of M7
      Company current semantic scope vs dormant pooled tenancy
      confirm company_id is current meaning, not speculative isolation substrate

R2-10 static Go <-> DDL CHECK vocabulary parity obligation

R2-11 idempotency retention Replay->Key order
       + HMAC-SHA-256 semantic fingerprint
       + fingerprint-key-version / non-expired compatibility law

R2-12 provisioner / owner / runtime / CI DB trust classes
       + grant proofs run as non-owner

R2-13 accepted LOW corrections:
       subject-specific RoleAssignment partial uniques
       SUBMITTED/current_submission CHECK
       Revision state/Release proof
       same-Company composite FKs
       one-open citation
       protected-actor narrowing
       ReplaySnapshot size bound deferred to T8-E
       restore mechanism-state precision
       explicit SQL/ORM baseline wording
```

Round 2 must explicitly report whether any corrected item creates:

```text
new BLOCKER
surviving material contradiction
T8-C reopen
T8-B/T1→T7 reopen
T8-E/F/G/T10/T11 trespass
need for a full third review round
```

---

## 21. Corrected-candidate verdict before Round 2

```text
GLOBAL MAXIMUM CLASS                    CONFIRMED
ROUND-1 BLOCKERS                        RESOLVED IN CORRECTED STAGING
ROUND-1 MAJORS                          ADJUDICATED
UPSTREAM REOPEN                         NONE PROPOSED
T8-E                                    NOT OPEN
IMPLEMENTATION                          BLOCKED

EXACT NEXT GATE
= BOUNDED FABLE ROUND 2 ON THIS CORRECTED DELTA
```

No durable T8-D authority may be promoted until Round 2, final Lead adjudication and explicit operator ratification complete.

---

**End of adjudicated corrected T8-D candidate.**