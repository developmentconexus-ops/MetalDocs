# R10 T8-D — Persistence Realization

> **Status:** CLOSED / OPERATOR-RATIFIED / PROMOTED  
> **Ratified:** 2026-08-20  
> **T8-E bounded correction:** 2026-08-20 — Governance Step label persistence + immutable attempt label snapshot
> **T8-E bounded correction:** 2026-08-21 — transaction/Audit/idempotency precision + same-PDF rendition + reconstructible CSRF session secret
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Upstream contract authority:** `wiki/architecture/r10-t8c-internal-communication-contracts.md`  
> **Implementation:** BLOCKED

This page is the durable target authority for R10 **T8-D — Persistence Realization**.

It consolidates the operator-ratified effective result of:

```text
original Global Maximum candidate
→ independent Fable Round 1
→ Lead adjudication
→ adjudicated corrected candidate
→ bounded Fable Round 2
→ final Lead adjudication
→ explicit operator ratification
```

The historical staging/reviewer artifacts are provenance only after promotion. They are not target authority.

T8-D freezes the smallest PostgreSQL persistence realization required to make T1→T8-C structurally executable without foreign SQL, duplicate current truth, hidden shared write authority, speculative persistence or legacy-shape inheritance. T8-D does not own exact HTTP/OpenAPI representation (T8-E), frontend realization (T8-F), runtime/deployment realization (T8-G), cutover (T10) or implementation decomposition (T11).

---

## 1. Ratified Global Maximum

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

No materially stronger persistence class with lower total complexity/proof cost survived the Method and independent challenge.

No T1→T7, T8-B or T8-C reopen is required.

---

## 2. Binding persistence laws

### 2.1 One product-state database

Launch uses one PostgreSQL product-state database so one native business transition can remain one local ACID transaction across current application choreography.

Database-per-owner/service splitting is rejected for Launch.

### 2.2 PostgreSQL isolation posture

Inherited unchanged from T2/T8-C:

```text
PostgreSQL READ COMMITTED
+ narrow explicit row serialization where required
+ explicit monotonic OCC/CAS
+ structural constraints where required
```

Global `SERIALIZABLE` is not the Launch default.

### 2.3 Owner-private SQL

```text
semantic SQL                  owner-private only
application semantic SQL      forbidden
transport SQL                 forbidden
platform semantic SQL         forbidden
cross-owner raw SQL           forbidden
shared semantic repository    forbidden
```

Application composes owner facts through T8-C contracts. Database FKs do not authorize foreign SQL.

### 2.4 Fully-qualified first-party SQL

First-party SQL names schema-qualified target relations. `search_path` is not an authority mechanism.

### 2.5 PostgreSQL feature floor

T8-D requires only PostgreSQL 16-compatible primitives. A later T8-G runtime may select a newer compatible version, but no T8-D invariant depends on post-16 features.

---

## 3. Namespace and object ownership

Target schemas:

```text
authn.*             Authentication
org.*               Organization
authz.*             Authorization
controlled_docs.*   Controlled Documents
audit.*             Audit
platform.*          MetalDocs technical mechanisms
river.*             River-owned third-party technical objects
```

PostgreSQL schemas are namespaces inside one database, not service/security boundaries.

### Closed DB-object catalog

Every target schema/table/view/sequence/type/function object class is classified.

Verification law:

```text
live target object absent from catalog       FAIL
catalog object absent from target schema     FAIL
first-party SQL against foreign owner schema FAIL
first-party raw SQL against river.*          FAIL
```

`river.*` is classified as `THIRD_PARTY_MANAGED`. MetalDocs does not pretend to own River's internal table semantics.

---

## 4. Physical type and closed-vocabulary policy

Baseline physical types:

```text
semantic/technical ids       UUID
versions/generations         BIGINT
revision ordinals/counters   BIGINT
trusted instants             TIMESTAMPTZ
exact hashes/digests         BYTEA
opaque replay snapshots      BYTEA
human/product codes          TEXT + bounded validation
closed lifecycle vocabularies TEXT + CHECK
Audit bounded facts          JSONB object + bounded size
```

PostgreSQL ENUM is not the default. Closed product/owner vocabularies remain code authority with DDL CHECK mirrors only where structural rejection is valuable.

For every DDL CHECK that mirrors a static code vocabulary:

```text
static code vocabulary set == DDL accepted set
mismatch => blocking verification failure
```

The executable falsifiable parity control belongs to T9's Validation Baseline; T8-D freezes the equality obligation.

---

## 5. Persistent-state census

### PERSIST — Authentication

```text
ProviderSubjectBinding
ApplicationSession
```

No password/local-auth/lockout/provider-role snapshot/session-history authority is persisted.

### PERSIST — Organization

```text
single Company root
User stable identity/current eligibility
separately erasable UserProfile
Area
Group
GroupMembership
```

### PERSIST — Authorization

```text
RoleAssignment current grant truth
```

### STATIC / CODE AUTHORITY

```text
Role vocabulary
Permission vocabulary
Role→Permission bundles
scope-compatibility vocabulary
```

### PERSIST — Controlled Documents

```text
DocumentType + independent config versions
GovernanceRoute step configuration
eligible Template relationships
number counters
Document stable identity/code/responsibility/Template role
DocumentOrigin
Revision
WorkingContent
Submission
SubmissionWithdrawal
RevisionCancellation
GovernanceAttempt
GovernanceAttemptStep
unactivated GROUP dependency
activated candidate snapshot
GovernanceDecision
SubmissionFeedback
Release
OfficialRendition
ObsolescenceRequest
semantic ExactContentDescriptor copies
```

### PERSIST — Audit

```text
immutable AuditEvent
```

### PERSIST — platform mechanisms

```text
ManagedContent mutable lifecycle/location state
immutable ManagedContent READY descriptor proof
immutable malware inspection evidence
AdmissionClaim
managed-content backup pin
IdempotencyKey claim identity
completed ReplaySnapshot
```

### PERSIST — River

River-owned third-party technical objects only.

### DERIVED / QUERY-ONLY

```text
Document catalog/current status
current EFFECTIVE lens
AuthorizedScopes
allowed_actions
Library
My Work
Governance work lists
Audit pages
Search baseline
```

### DEFER / NOT LAUNCH

```text
RLS / tenant GUC / pooled isolation substrate
generic Artifact
generic Workflow / policy engine
custom persistent Role/Permission catalog
expanded permission cache/materialized ACL
materialized Search / search_refresh
notifications / generic outbox / event bus
WorkingSnapshot business authority
EditorSession correctness authority
Distribution / Periodic Review / Dossier / Evidence / Records persistence
```

---

## 6. Authentication persistence

### `authn.provider_subject_bindings`

Material shape:

```text
id             UUID PRIMARY KEY
user_id        UUID NOT NULL REFERENCES org.users(id)
issuer         TEXT NOT NULL
subject        TEXT NOT NULL
version        BIGINT NOT NULL CHECK(version >= 1)
bound_at       TIMESTAMPTZ NOT NULL
replaced_at    TIMESTAMPTZ NULL
```

Current uniqueness:

```text
UNIQUE(issuer, subject) WHERE replaced_at IS NULL
UNIQUE(user_id)         WHERE replaced_at IS NULL
```

Historical rows remain immutable except the one-time `replaced_at` transition that ends currentness.

This deliberately permits truthful recovery from an administrative mis-binding and later legitimate reassignment/revert while making simultaneous current double binding impossible.

Provider-binding replacement:

```text
provider lookup external preflight
→ protected actor
→ lock current binding
→ expected VersionToken check
→ exact already-current = no-op/current token
→ terminate prior current row
→ insert new row version+1
→ delete current ApplicationSessions for User
→ required Audit
→ commit
```

Offboarding preserves provider-binding history/current binding; disabled User eligibility denies session issuance.

### `authn.application_sessions`

```text
id                  UUID PRIMARY KEY
user_id             UUID NOT NULL REFERENCES org.users(id)
token_digest        BYTEA NOT NULL UNIQUE
csrf_secret         BYTEA NOT NULL
created_at          TIMESTAMPTZ NOT NULL
expires_at          TIMESTAMPTZ NOT NULL
CHECK(expires_at > created_at)
```

Session state is current access state only:

```text
issue               INSERT
logout/revoke       DELETE
offboard            DELETE all for User
binding replacement DELETE all for User
restore readiness   purge all restored sessions before serving
```

Session issuance generates a cryptographically random per-session `csrf_secret`. Session resolution may return that opaque secret/token material only after the ApplicationSession cookie has authenticated through `token_digest`. Unsafe-request CSRF validation compares the supplied token to this server-side session secret in constant time. The CSRF secret is not an authentication bearer credential and grants nothing without the valid HttpOnly session cookie.

No raw **authentication** token/IP/User-Agent/device/tenant selector/permission snapshot is persisted as target authority.

---

## 7. Organization persistence

### `org.companies`

```text
id             UUID PRIMARY KEY
singleton_key  SMALLINT NOT NULL UNIQUE CHECK(singleton_key = 1)
display_name   TEXT NOT NULL
version        BIGINT NOT NULL CHECK(version >= 1)
created_at     TIMESTAMPTZ NOT NULL
```

The singleton CHECK is a fail-closed Launch interlock. `company_id` remains on rows where Company is a current semantic/scope dimension. This is not a pooled-tenancy isolation substrate.

A future pooled-tenancy reopen must change the singleton interlock together with the isolation/runtime substrate it then owes.

### `org.users`

```text
id                    UUID PRIMARY KEY
company_id            UUID NOT NULL REFERENCES org.companies(id)
enabled               BOOLEAN NOT NULL
eligibility_version   BIGINT NOT NULL CHECK(eligibility_version >= 1)
created_at             TIMESTAMPTZ NOT NULL
```

User is stable historical identity; serving runtime cannot DELETE it.

### `org.user_profiles`

Minimum correctness-bearing shape:

```text
user_id        UUID PRIMARY KEY REFERENCES org.users(id)
display_name   TEXT NOT NULL
email          TEXT NULL
version        BIGINT NOT NULL CHECK(version >= 1)
```

UserProfile is separately erasable without deleting User.

### `org.areas`

```text
id          UUID PRIMARY KEY
company_id  UUID NOT NULL REFERENCES org.companies(id)
code        TEXT NOT NULL
name        TEXT NOT NULL
state       TEXT NOT NULL CHECK(state IN ('ACTIVE','RETIRED'))
version     BIGINT NOT NULL CHECK(version >= 1)
created_at  TIMESTAMPTZ NOT NULL
UNIQUE(company_id, code)
```

Area code is immutable after creation.

Document creation protects Area eligibility with `FOR SHARE` through commit. Area lifecycle replacement/retirement uses `FOR UPDATE` plus the owner VersionToken/CAS and required Audit. Thus create-vs-retire linearizes without a global lock.

### `org.groups`

```text
id          UUID PRIMARY KEY
company_id  UUID NOT NULL REFERENCES org.companies(id)
name        TEXT NOT NULL
version     BIGINT NOT NULL CHECK(version >= 1)
created_at  TIMESTAMPTZ NOT NULL
```

Group is deletable only after its current dependencies disappear.

### `org.group_memberships`

```text
company_id  UUID NOT NULL
group_id    UUID NOT NULL
user_id     UUID NOT NULL
created_at  TIMESTAMPTZ NOT NULL
PRIMARY KEY(group_id, user_id)
```

Composite identity keys/FKs structurally preserve:

```text
membership.company_id = group.company_id = user.company_id
```

While Launch is structurally singleton this is seam structure, not a current multi-Company isolation proof. No membership-history table is introduced.

---

## 8. Authorization persistence

`authz.role_assignments` is the only persisted Authorization authority:

```text
id          UUID PRIMARY KEY
company_id  UUID NOT NULL
user_id     UUID NULL
group_id    UUID NULL
role_code   TEXT NOT NULL
area_id     UUID NULL
created_at  TIMESTAMPTZ NOT NULL
```

Checks:

```text
exactly one of user_id/group_id is non-null
role_code belongs to static Launch Role vocabulary
area_id NULL     = Company scope
area_id non-null = Area scope
governance_admin requires Company scope
area_manager requires Area scope
```

Duplicate prevention is subject-specific:

```text
UNIQUE NULLS NOT DISTINCT(company_id,user_id,role_code,area_id)
WHERE user_id IS NOT NULL

UNIQUE NULLS NOT DISTINCT(company_id,group_id,role_code,area_id)
WHERE group_id IS NOT NULL
```

No persistent Role/Permission/bundle/effective-permission tables exist.

---

## 9. DocumentType and numbering persistence

### `controlled_docs.document_types`

Material shape includes:

```text
id
company_id
code
name
active
numbering_scope = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA
governance_mode = NO_HUMAN_APPROVAL | USE_GOVERNANCE_ROUTE
representation_mode = SOURCE_ONLY | REQUIRE_OFFICIAL_RENDITION_PDF
base_version
governance_version
eligible_templates_version
created_at
UNIQUE(company_id, code)
```

The three versions are independent current concurrency authorities so unrelated whole-replacement resources do not create false conflicts.

`code` and `numbering_scope` become immutable after the first committed Document uses the type. This is owner-private SQL under a DocumentType lock, not a cross-table CHECK or trigger.

### `controlled_docs.document_type_governance_steps`

Each configured step stores:

```text
document_type_id
ordinal
label TEXT NOT NULL
selector_kind = NAMED_USER | GROUP
named_user_id NULL
group_id NULL
```

Exactly one selector is present. Current GROUP selectors carry a real FK to `org.groups` so current route dependency blocks Group deletion.

### `controlled_docs.document_type_eligible_templates`

Current eligible Template relationships are persisted as a bounded set under `eligible_templates_version`; no TemplateVersion authority is created.

### `controlled_docs.document_number_counters`

```text
id
 document_type_id
 area_id NULL
 next_value BIGINT CHECK(next_value >= 1)
UNIQUE NULLS NOT DISTINCT(document_type_id, area_id)
```

`area_id NULL` is the type-wide counter; non-null is type+Area. Preview is read-only/non-reserving. Final Document create allocates in its local transaction. Gaps are allowed.

---

## 10. Document / Revision / WorkingContent

### `controlled_docs.documents`

Material shape:

```text
id
company_id
document_type_id
area_id
code
responsible_user_id
responsible_owner_version
is_template
template_role_version
created_at
UNIQUE(company_id, code)
```

No `current_revision_id`, `current_release_id`, `current_status` or latest-submission pointer exists on Document.

The Document row is nevertheless the stable lifecycle serialization root.

### `controlled_docs.document_origins`

For create-from-Template provenance only:

```text
document_id PRIMARY KEY
source_template_document_id
source_template_revision_id
source_sha256
source_size_bytes
source_content_format
created_at
```

No provider locator/generic provenance graph.

### `controlled_docs.revisions`

Material shape:

```text
id
 document_id
 ordinal
 title
 state = DRAFT | SUBMITTED | EFFECTIVE | SUPERSEDED | CANCELLED | OBSOLETE
 current_submission_id NULL
 created_at
```

Constraints:

```text
UNIQUE(document_id, ordinal)
UNIQUE(document_id) WHERE state IN ('DRAFT','SUBMITTED')
UNIQUE(document_id) WHERE state = 'EFFECTIVE'
CHECK((state='SUBMITTED') = (current_submission_id IS NOT NULL))
```

One-open Revision is derived from T2/T6 singular current-open Revision semantics. Revision rows are not serving-deletable, so ordinals are not reused.

`Revision.state` is canonical lifecycle state. `Release` is the immutable effectivity-establishing fact. This is not duplicate authority; the forbidden duplicate remains `Document.current_status`.

Target proof must show no Revision reaches EFFECTIVE/SUPERSEDED/OBSOLETE without the corresponding Release/effectivity path.

### `controlled_docs.working_contents`

```text
revision_id      UUID PRIMARY KEY
generation       BIGINT NOT NULL CHECK(generation >= 0)
managed_content_id UUID NOT NULL
sha256           BYTEA NOT NULL CHECK(octet_length(sha256)=32)
size_bytes       BIGINT NOT NULL CHECK(size_bytes >= 0)
content_format   closed format vocabulary
updated_at       TIMESTAMPTZ NOT NULL
```

`generation` is the one DRAFT OCC authority for both Revision title and WorkingContent source.

Successful material DRAFT mutation increments exactly once. Stale expected generation causes zero mutation. Exact no-op does not fabricate a generation advance.

Every semantic write that creates/replaces a `managed_content_id` reference first locks the target `platform.managed_content` row `FOR SHARE` through commit and revalidates READY + immutable descriptor + applicable admission/malware proof.

---

## 11. Submission / governance / cancellation

### `controlled_docs.submissions`

Immutable governed attempt:

```text
id
revision_id
submitter_user_id
submitted_at
title_snapshot
managed_content_id
sha256
size_bytes
content_format
governance_mode_snapshot
representation_mode_snapshot
required_rendition_format NULL
```

Serving privileges are SELECT + INSERT only.

Revision/current Submission relationship is same-Revision constrained, and SUBMITTED iff current_submission_id is present.

### `controlled_docs.submission_withdrawals`

```text
submission_id PRIMARY KEY
actor_user_id
withdrawn_at
```

Insert-only.

### `controlled_docs.revision_cancellations`

```text
revision_id PRIMARY KEY
actor_user_id
reason
cancelled_at
```

Insert-only. Cancellation terminates the live attempt without fabricating a governance verdict.

### `controlled_docs.governance_attempts`

Exactly one governed subject:

```text
submission_id NULL
obsolescence_request_id NULL
state = ACTIVE | COMPLETED | RETURNED | WITHDRAWN | CANCELLED
created_at
ended_at NULL
CHECK exactly one subject
UNIQUE(submission_id)
UNIQUE(obsolescence_request_id)
```

No generic `subject_type/subject_id` workflow substrate.

### `controlled_docs.governance_attempt_steps`

Each immutable selector snapshot has bounded mutable activation/decision state:

```text
id
attempt_id
ordinal
label_snapshot TEXT NOT NULL
selector_kind
named_user_id NULL
group_id_snapshot NULL
state = PENDING | ACTIVE | DECIDED
activated_at NULL
UNIQUE(attempt_id, ordinal)
UNIQUE(attempt_id) WHERE state='ACTIVE'
```

Snapshot label/selector columns are not serving-updateable after creation.

### GROUP deletion / activation

`controlled_docs.governance_group_dependencies(step_id PK, group_id FK org.groups)` exists only while an unactivated GROUP Step still requires the live Group identity.

Activation in one Scope:

```text
NAMED_USER -> insert configured User into governance_step_candidates
GROUP      -> resolve current enabled members; insert all returned Users
empty GROUP -> insert zero candidates, no fallback
→ delete live GROUP dependency for activated Step
→ Step ACTIVE
```

This occurs on first-Step activation at SUBMIT and on next-Step activation after ACCEPT.

Historical `group_id_snapshot` has no permanent Group FK, so activated/completed history does not keep Group alive solely because Group once participated.

### `controlled_docs.governance_step_candidates`

```text
step_id
user_id
PRIMARY KEY(step_id,user_id)
```

Insert-only frozen candidate snapshot.

### `controlled_docs.governance_decisions`

```text
id
step_id UNIQUE
actor_user_id
outcome = ACCEPT | RETURN_FOR_CHANGES
reason NULL
decided_at
FOREIGN KEY(step_id,actor_user_id)
  REFERENCES governance_step_candidates(step_id,user_id)
```

Insert-only.

The FK proves membership in the frozen activation snapshot; current T3 Authorization/SoD is still rechecked when the actor acts. Empty candidate set structurally admits no decision.

### `controlled_docs.submission_feedback`

Append-only bounded feedback/history. Free-form feedback remains semantic history but is excluded from Audit/ReplaySnapshot unless an upstream bounded fact explicitly requires otherwise.

---

## 12. Release / rendition / obsolescence

### `controlled_docs.releases`

Immutable effectivity fact:

```text
id
 document_id
 revision_id
 submission_id
 predecessor_revision_id NULL
 released_at
UNIQUE(revision_id)
UNIQUE(submission_id)
UNIQUE(predecessor_revision_id) WHERE predecessor_revision_id IS NOT NULL
UNIQUE(document_id) WHERE predecessor_revision_id IS NULL
```

Composite constraints prove winning Revision belongs Document and winning Submission belongs winning Revision.

Replacement under Document `FOR UPDATE` atomically performs:

```text
prior EFFECTIVE -> SUPERSEDED
successor SUBMITTED -> EFFECTIVE
clear successor current_submission_id
INSERT Release
required Audit
```

The partial unique EFFECTIVE index is the final structural barrier against dual effectivity.

NoHumanApproval + SourceOnly may create Submission and Release in the same transaction when upstream gates are satisfied.

### `controlled_docs.official_renditions`

```text
id
submission_id
required_format
managed_content_id
sha256
size_bytes
content_format
created_at
UNIQUE(submission_id, required_format)
```

Insert-only semantic exact-content fact. Renderer/provider identity is not semantic authority.

If a Submission already contains admitted PDF and its frozen policy requires OfficialRendition(PDF), the OfficialRendition row may reference the **same** `managed_content_id` and copy the same exact descriptor as the Submission after canonical eligibility/descriptor revalidation. No provider copy, renderer output or River intent is created.

When transformation is required (current Launch: DOCX→PDF), rendition bytes are prepared outside the semantic transaction; final admission uses the universal ManagedContent `FOR SHARE` rule plus exact descriptor/malware evidence as applicable, then revalidates the governed state before insert/release consequence.

### `controlled_docs.obsolescence_requests`

```text
id
document_id
target_revision_id
initiator_user_id
reason
governance_mode_snapshot
state = ACTIVE | RETURNED | WITHDRAWN | COMPLETED
requested_at
ended_at NULL
UNIQUE(document_id) WHERE state='ACTIVE'
```

Final successful completion under Document serialization atomically transitions the target EFFECTIVE Revision to OBSOLETE and completes the attempt/request. No current-status pointer is added.

---

## 13. Audit persistence

`audit.events` material shape:

```text
id UUID PRIMARY KEY
occurred_at TIMESTAMPTZ NOT NULL
actor_kind
actor_user_id NULL
system_actor_code NULL
operation_code
resource_kind
resource_id UUID NOT NULL
company_id UUID NOT NULL
area_id UUID NULL
facts JSONB NOT NULL
correlation_id NULL
```

Checks include:

```text
USER actor <=> actor_user_id present
SYSTEM actor <=> system_actor_code present
facts is JSON object
facts bounded (64 KiB current proof-backed ceiling)
```

Serving privileges:

```text
SELECT + INSERT
NO UPDATE
NO DELETE
NO TRUNCATE
```

Audit does not copy UserProfile PII, governed content, request bodies or free-form governance reasons/feedback by default. Historical Company/Area attribution is immutable and filtered before pagination.

Audit `resource_id` stays required for semantic Audit. Operational restore/session reconciliation is not forced into semantic Audit merely to create a resource-less event.

No global Audit hash chain/sequence head is retained.

---

## 14. ManagedContent exact-byte mechanism

### `platform.managed_content`

Mutable mechanism lifecycle/location state only:

```text
id UUID PRIMARY KEY
state = OPEN | READY | GC_PENDING
provider_locator
trust_class = UNTRUSTED_EXTERNAL | TRUSTED_MANAGED_COPY | TRUSTED_INTERNAL_DERIVATION
created_at
ready_at NULL
gc_pending_at NULL
```

No semantic `owner_type/owner_id`, Document/Revision/Submission identity or provider ETag authority.

### `platform.managed_content_descriptors`

Immutable READY descriptor proof:

```text
managed_content_id UUID PRIMARY KEY REFERENCES platform.managed_content(id)
sha256             BYTEA NOT NULL CHECK(octet_length(sha256)=32)
size_bytes         BIGINT NOT NULL CHECK(size_bytes >= 0)
content_format     closed Launch format vocabulary NOT NULL
derived_at         TIMESTAMPTZ NOT NULL
```

Serving privileges: SELECT + INSERT only.

This is mechanism evidence used to establish semantic descriptor truth; semantic WorkingContent/Submission/OfficialRendition rows still copy and own their own descriptor.

### `platform.malware_inspections`

```text
id UUID PRIMARY KEY
managed_content_id UUID NOT NULL REFERENCES platform.managed_content(id)
digest BYTEA NOT NULL CHECK(octet_length(digest)=32)
verdict = CLEAN | MALICIOUS
inspected_at TIMESTAMPTZ NOT NULL
UNIQUE(managed_content_id,digest)
```

Serving privileges: SELECT + INSERT only.

For untrusted immutable governed admission:

```text
managed_content_descriptors.sha256 = malware_inspections.digest
AND verdict = CLEAN
```

`READY` alone never means CLEAN.

### OPEN → READY

External/provider exact-byte read, hash, format validation and scanner call remain outside a product semantic transaction. Local mechanism commit:

```text
ManagedContent FOR UPDATE
→ state OPEN
→ require same live claim when claim-bound
→ prove provider create-once/no-overwrite
→ INSERT immutable descriptor
→ INSERT terminal malware evidence if produced
→ state READY + ready_at
→ commit
```

---

## 15. AdmissionClaim and backup pins

### `platform.admission_claims`

Conceptual shape:

```text
id UUID PRIMARY KEY
managed_content_id UUID UNIQUE REFERENCES platform.managed_content(id)
opaque binding material/digest as required by T8-C
created_at
expires_at
```

For claim-bound preparation paths:

```text
allocation
→ create ManagedContent OPEN
→ create AdmissionClaim in or before same local commit
→ return handle only after claim exists

complete
→ require same live claim
→ OPEN -> READY

semantic attach
→ require same live claim
→ ManagedContent FOR SHARE + final READY proof
→ semantic reference write
→ ConsumeIn = DELETE claim in same Scope
```

A live claim spans OPEN and READY until consume/release/expiry. Rollback of ConsumeIn restores it.

Any live claim blocks GC eligibility independent of state.

### `platform.managed_content_backup_pins`

```text
backup_pin_id
managed_content_id
expires_at
PRIMARY KEY(backup_pin_id,managed_content_id)
```

Pin acquisition that creates new protection:

```text
ManagedContent FOR SHARE
→ prove eligible technical handle/current protection relation
→ INSERT pin
→ commit
```

Release/expiry removes the pin in a bounded technical transaction. Exact backup orchestration remains T8-G, but the persistence/locking law is T8-D.

---

## 16. GC realization

GC phase 1 and phase 2 serialize solely on the target ManagedContent row with `FOR UPDATE`.

### Phase 1

```text
ManagedContent FOR UPDATE
→ state must be reclaimable READY-class candidate
→ non-locking ControlledDocs semantic-reference proof
→ non-locking live AdmissionClaim proof
→ non-locking backup-pin proof
→ if all absent: READY -> GC_PENDING
→ commit
```

### Phase 2

Immediately before provider delete:

```text
ManagedContent FOR UPDATE
→ require still GC_PENDING
→ repeat full semantic-reference proof as NON-LOCKING current reads
→ repeat live-claim proof as NON-LOCKING current read
→ repeat backup-pin proof as NON-LOCKING current read
→ commit proof transaction
→ provider DeleteReclaimable outside product tx
```

GC acquires no lower-class semantic/claim/pin row locks after taking ManagedContent `FOR UPDATE`.

Safety argument under READ COMMITTED:

```text
new semantic reference requires ManagedContent FOR SHARE
new backup pin requires ManagedContent FOR SHARE
→ both block behind GC FOR UPDATE
claim-bound allocation creates claim before READY candidacy
semantic claim consumption is coupled to attach FOR SHARE
```

Therefore no protective semantic reference/pin can appear behind GC's back while the root lock is held.

Safe failure remains leaked bytes, never deleted governed truth.

---

## 17. Idempotency persistence

### `platform.idempotency_keys`

```text
id UUID PRIMARY KEY
actor_user_id UUID NOT NULL
operation_id TEXT NOT NULL
key UUID NOT NULL
semantic_fingerprint BYTEA NOT NULL CHECK(octet_length(semantic_fingerprint)=32)
fingerprint_key_version INTEGER NOT NULL CHECK(fingerprint_key_version > 0)
created_at TIMESTAMPTZ NOT NULL
expires_at TIMESTAMPTZ NOT NULL
UNIQUE(actor_user_id,operation_id,key)
```

### `platform.idempotency_replays`

```text
key_id UUID PRIMARY KEY
snapshot_version INTEGER NOT NULL
payload BYTEA NOT NULL CHECK(octet_length(payload) <= 2048)
completed_at TIMESTAMPTZ NOT NULL
```

ReplaySnapshot remains versioned, self-contained and PII-free by construction. T8-E freezes the Launch maximum at 2,048 bytes and T8-D mirrors that bound structurally. The client `Idempotency-Key` is stored as UUID identity so textual UUID case/form cannot create parallel scoped uniqueness identities.

### Commit invariant

```text
replay.key_id -> key.id        immediate FK
key.id        -> replay.key_id DEFERRABLE INITIALLY DEFERRED FK
```

Consequences:

```text
committed key without completed ReplaySnapshot -> COMMIT FAILS
semantic mutation + Audit + required River intent + ReplaySnapshot commit atomically
```

Winner path uses the immediate scoped unique with `INSERT ... ON CONFLICT DO NOTHING`; under READ COMMITTED a loser waits without poisoning Scope, then sees the winner's committed ReplaySnapshot on the next command. If winner rolls back, contender may become owner. Same key + different fingerprint is conflict/zero business mutation.

Completion chooses one trusted `completed_at` and, before commit, sets the paired Key `expires_at = completed_at + 24 hours`. This is the semantic replay boundary; cleanup timing cannot extend it.

A BeginIn conflict on an existing scoped key serializes on that row. If `now >= expires_at`, the transaction deletes the expired Replay then Key and retries claim establishment in the same Scope. If the row is still live, normal replay/fingerprint-conflict behavior applies. Concurrent post-expiry reuse still has one winner/loser path. The janitor remains cleanup only, never semantic expiry authority.

### Fingerprint privacy/equality

Application derives:

```text
HMAC-SHA-256(
  server-held fingerprint key,
  canonical operation identity + complete validated semantic command
)
```

The HMAC key is never persisted in the product database.

Exactly one fingerprint-key version is active for derivation at a time. Rotation is drain-before-activate:

```text
new version may become active only after
all idempotency keys produced under the old derivation version
have expired or been safely retired

old key material remains available until drain completes
```

Secret provisioning/rotation mechanics are T8-G; this equality law is T8-D.

### Retention

Mandatory order:

```text
BEGIN
DELETE expired idempotency_replays
DELETE matching idempotency_keys
COMMIT
```

The asymmetric FK design is retained deliberately.

---

## 18. River persistence boundary

River remains pinned by repository evidence at v0.37.1 for the current realization input and is configured to use:

```text
schema = river
```

River migrations run through an owner-capable migration path. `river.*` is owned by `metaldocs_owner`, never by serving runtime.

Runtime receives only required schema/type/function/sequence/DML privileges. Raw first-party SQL against River objects is forbidden; platform River integration uses River APIs and the T8-C `SQLTx(scope)` native binding.

### PostgreSQL 16 maintenance consequence

River self-REINDEX maintenance is **OFF for Launch**. On PostgreSQL 16 a non-owner cannot be granted the later MAINTAIN privilege needed to make self-REINDEX compatible with the target trust boundary. The serving role is not made owner merely for optional maintenance.

T8-G must wire the pinned supported disable mechanism and observe River index health/queue-fetch latency so materially adverse evidence can trigger a bounded operations reopen. No silent owner escalation.

---

## 19. Cross-owner FK law

Cross-owner FKs may prove **stable identity/existence only**.

Representative allowed references:

```text
authn bindings/sessions -> org.users
authz RoleAssignments   -> org.users/org.groups/org.areas/org.companies
ControlledDocs Document -> org.areas/org.users/org.companies
current route selector  -> org.users/org.groups
frozen governance candidates/actors -> org.users
Audit actor/visibility  -> org.users/org.companies/org.areas
semantic content handles -> platform.managed_content
idempotency actor -> org.users
```

Universal rules:

```text
NO cross-owner ON DELETE CASCADE
FK never decides enabled/authorized/EFFECTIVE/eligible/deletable semantics
foreign owner SQL remains forbidden
```

Group deletion is structurally blocked exactly by current dependencies:

```text
GroupMembership
Group RoleAssignment
current GovernanceRoute GROUP selector
unactivated live GROUP dependency
```

Activated/completed history no longer keeps Group alive solely because Group once participated.

---

## 20. DB trust and immutable-history grants

Exactly four target trust classes:

```text
bootstrap/provisioner
  provisioning-only
  creates roles/schemas/extensions
  ONLY class permitted to SET ROLE metaldocs_owner
  never serves product traffic

metaldocs_owner
  NOLOGIN
  owns first-party and river.* DDL objects
  migration/DDL authority

metaldocs_runtime
  LOGIN serving class
  no schema/table ownership
  no owner-role membership
  target serving privileges only

metaldocs_verifier
  non-owner proof/CI class
  no owner-role membership
  effective privileges MUST equal metaldocs_runtime across closed catalog
```

A fifth class requires a named risk reduction.

Blocking grant-parity obligation:

```text
effective privileges(metaldocs_verifier)
==
effective privileges(metaldocs_runtime)
for every catalogued schema/table/sequence/type/function/object class
```

Grant-restriction proofs must run as `metaldocs_verifier`, not owner/superuser.

### Immutable-history classes

Serving runtime receives SELECT + INSERT only, with no UPDATE/DELETE/TRUNCATE, for classes such as:

```text
submissions
submission_withdrawals
revision_cancellations
governance_step_candidates
governance_decisions
submission_feedback
releases
official_renditions
document_origins
audit.events
platform.managed_content_descriptors
platform.malware_inspections
```

Where a table contains both frozen snapshot columns and bounded lifecycle columns, column-level grants may allow UPDATE only on the bounded lifecycle fields.

No semantic lifecycle trigger is baseline authority.

---

## 21. VersionToken / OCC law

Whole-replacement resources use explicit monotonic `BIGINT` versions.

```text
initial = 1
expected != current        -> zero mutation
material replacement       -> version + 1
exact already-current      -> no mutation / no Audit / same version
```

DRAFT uses `WorkingContent.generation`, initial 0.

Do not use `xmin`, `ctid`, timestamps or representation hashes as persistence concurrency authority.

---

## 22. Serialization / lock law

### Protected actor

Every authenticated user-initiated semantic product mutation uses Organization `ProtectedSecuritySubjectIn`, physically realized as User `FOR SHARE`, held until the local transaction completes.

SYSTEM/background transitions do not invent a User lock.

### User lock ordering

If actor/targets require multiple User locks:

```text
collect User ids
→ dedupe
→ UUID ascending
→ acquire once per User at strongest required mode
```

Offboarding/eligibility root uses update-strength locking; ordinary protected eligibility uses `FOR SHARE`.

### Owner roots

```text
Document lifecycle/effectivity      Document FOR UPDATE
DRAFT relationship protection       Document FOR SHARE + generation CAS
DocumentType snapshot read          DocumentType FOR SHARE
DocumentType replacement            DocumentType FOR UPDATE
Area create-eligibility             Area FOR SHARE
Area lifecycle mutation             Area FOR UPDATE
ManagedContent semantic attach      ManagedContent FOR SHARE
ManagedContent OPEN->READY           ManagedContent FOR UPDATE
ManagedContent GC phase 1/2         ManagedContent FOR UPDATE
number allocation                   counter row FOR UPDATE
```

Rows within one lock class are acquired in stable deterministic order.

### GC exception to ordinary class ordering

GC takes ManagedContent `FOR UPDATE` first and then performs downstream proofs as non-locking reads. It does not acquire semantic/claim/pin row locks after the root lock.

---

## 23. Transaction census

Every authenticated semantic mutation inherits actor `FOR SHARE`. Natural idempotent DELETEs need no durable Idempotency-Key unless T6 explicitly says otherwise.

Each transaction appends **all and only** the T3-required semantic Audit evidence for the facts/effects it commits. One business transition may therefore append multiple Audit events; the singular `Audit` shorthand below never suppresses required companion evidence and never creates an event T3 does not require.

Material mutation families include:

```text
Session issue
  protected User -> insert ApplicationSession

Session logout/revoke
  delete current ApplicationSession

User create
  Idempotency claim -> protected actor -> User + UserProfile + ProviderSubjectBinding + Audit + Replay

Provider-binding replacement
  provider preflight -> actor/target ordering -> expected version -> replace binding -> purge sessions -> Audit

UserProfile replacement/erasure
  protected actor/target as applicable -> VersionToken/owner rule -> ordinary replacement has no mandatory semantic Audit; lawful erasure appends user_profile.erased evidence when T3 requires it

Offboarding
  ordered User locks -> target eligibility update root -> session purge + memberships/direct RoleAssignments removal + all required teardown/offboarding Audit

User re-enable
  protected actor/target -> eligibility VersionToken/CAS -> DISABLED->ENABLED only -> user.reenabled Audit; no prior session/membership/grant resurrection

Company replacement
  protected actor -> Company version CAS; no mandatory semantic Audit

Area create
  Idempotency -> protected actor -> Organization insert + Audit + Replay

Area replace
  protected actor -> version CAS -> Audit

Area lifecycle
  protected actor -> Area FOR UPDATE -> transition/version -> Audit

Group create
  Idempotency -> protected actor -> Organization insert + Audit + Replay

Group replace/delete
  protected actor -> Group current truth + FK/dependency rules -> Audit

GroupMembership add/remove
  protected actor + protected target User where applicable -> membership mutation -> Audit

RoleAssignment create
  Idempotency where T6 requires -> protected actor/target -> Authorization insert -> Audit -> Replay

RoleAssignment revoke
  protected actor -> delete current assignment -> Audit

DocumentType create/replace
  protected actor -> DocumentType config/version lock -> Audit

Governance config replace
  protected actor -> DocumentType FOR UPDATE -> governance version -> children replacement -> Audit

Eligible Templates replace
  protected actor -> DocumentType FOR UPDATE -> eligible_templates_version -> set replacement -> Audit

Document create
  Idempotency -> protected actor/owner target -> Area FOR SHARE ACTIVE -> DocumentType protection -> numbering counter -> semantic seed attach FOR SHARE -> Document + REV000 + WorkingContent + optional Origin + Audit + Replay

Next Revision
  Idempotency -> protected actor -> Document FOR UPDATE -> source current EFFECTIVE revalidation -> seed attach FOR SHARE -> Revision + WorkingContent + Audit + Replay

DRAFT PATCH
  protected actor -> Document FOR SHARE -> new handle FOR SHARE if replacing source -> generation CAS -> title/source update -> claim consume if applicable; no mandatory Launch semantic Audit

Draft upload allocate
  protected actor/journey authorization -> ManagedContent OPEN + AdmissionClaim in/before same commit

Draft upload complete
  external exact-byte read/scan -> ManagedContent FOR UPDATE -> same live claim -> create-once proof -> immutable descriptor + terminal malware evidence -> OPEN->READY

SUBMIT
  Idempotency -> protected actor -> DocumentType config snapshot protection -> Document FOR UPDATE -> source handle FOR SHARE -> immutable Submission -> route/Step snapshot including label_snapshot -> activate first candidate set
  -> SourceOnly: no rendition work
  -> RequireOfficialRendition(PDF) + source PDF: insert OfficialRendition over same handle/descriptor; no provider copy/renderer/River intent
  -> RequireOfficialRendition(PDF) + source DOCX: enqueue required River rendition intent
  -> append all T3-required evidence for Submission + any same-tx OfficialRendition/Release -> Replay

Feedback
  Idempotency where required -> protected actor -> exact case validation -> append immutable feedback -> Replay; no duplicate semantic Audit

Governance ACCEPT / RETURN
  protected actor -> Document FOR UPDATE -> exact active Step/candidate/Authorization/SoD -> insert Decision -> activate next candidates or RETURN/release consequence -> Audit

Withdraw
  protected actor -> Document FOR UPDATE -> withdrawal fact -> terminate attempt/dependencies -> Revision DRAFT -> Audit

Cancel
  protected actor -> Document FOR UPDATE -> cancellation fact -> terminate attempt/dependencies -> Revision CANCELLED -> Audit

Responsible-owner replacement
  ordered actor/target User protection -> Document FOR UPDATE -> owner VersionToken -> Audit

Template role replacement
  protected actor -> Document FOR UPDATE -> template_role_version -> Audit

Obsolescence initiation
  Idempotency -> protected actor -> Document FOR UPDATE -> current EFFECTIVE/no-open-replacement checks -> request/attempt or same-tx completion -> Audit/Replay

Obsolescence withdrawal/final completion
  protected actor -> Document FOR UPDATE -> request/attempt transition -> Revision OBSOLETE on success -> Audit

OfficialRendition finalization
  renderer/background transformation path -> Document FOR UPDATE -> rendition handle FOR SHARE -> exact descriptor/proof -> insert OfficialRendition -> official_rendition.completed Audit -> release consequence when gates satisfied -> release.completed Audit when Release is established in the same commit

Backup pin acquire/release
  handle FOR SHARE for new pin -> pin insert; bounded delete/expiry for release

Idempotency acquire/replay
  scoped unique Key + deferred completed Replay invariant

GC phase 1/2
  ManagedContent FOR UPDATE as sole serialization root + non-locking reference/claim/pin proofs
```

External provider/network work never joins the semantic product transaction.

---

## 24. Query ownership and Search

SQL query ownership:

```text
Authentication   binding/session
Organization     User/Profile/Area/Group/Membership and eligibility
Authorization    RoleAssignment/grant scope
Controlled Docs  lifecycle/access facts/Library/My Work/Governance/History/Search
Audit            historical event visibility
Platform         ManagedContent/claims/pins/idempotency
River            River API only
```

No owner joins foreign owner tables in its repository. Application gathers bounded cross-owner facts through T8-C contracts.

### Library / Search

Baseline Search remains a normal owner-private relational query over current EFFECTIVE truth:

```text
Document.code
current EFFECTIVE Revision.title
DocumentType
Area identity
responsible owner identity
lens-derived status
```

No materialized Search table/view, external engine, body/OCR/vector index or refresh job.

Organization display labels are composed after the owner page is fixed; free-text ranking/pagination remain entirely ControlledDocs-owned.

### Audit

Audit historical Company/Area visibility predicate is applied before pagination. Current relocation never rewrites attribution.

---

## 25. Restore persistence consequences

Restore remains fail-closed on missing/corrupt governed bytes and invalidates restored ApplicationSessions before serving.

Other restored mechanism state converges through its own invariant:

```text
River pending jobs       durable DB state revalidated/resumed by workers
GC_PENDING               phase-2 repeated proof
AdmissionClaim           expiry/release/consume law
backup pins              backup/recovery lifecycle
Idempotency Key/Replay   completion invariant + expiry/retention
```

No generic recovery journal/session-history subsystem is added.

---

## 26. T8-A selective-reuse disposition

```text
auth_identities/password/local-auth       DELETE / REWRITE
auth_sessions                            REWRITE; preserve durable-current-session property
iam_* / legacy access families           REHOME / REWRITE
role_capabilities                        DELETE; static code authority
RLS/tenant/GUC substrate                 DELETE from Launch target
controlled_documents                     REWRITE
technical document_revisions             DELETE / REWRITE
approval_*                               REWRITE into bounded governance relations
taxonomy/template platform tables        DELETE/fold into ratified owners
Audit table shape                        REWRITE / REHOME
Audit append-only serving grants         PRESERVE PROPERTY
Audit global hash chain                  DELETE
current idempotency status/raw HTTP      REWRITE
unique-key concurrency property          PRESERVE / REFINE
River mechanism                          PRESERVE / REHOME to river.*
runtime != DDL owner property            PRESERVE / REFINE to four trust classes
closed object-ownership completeness     PRESERVE PROPERTY / REWRITE catalog
outboxes/notifications                   DELETE absent named Launch consumer
```

Current existence/tests never create survival entitlement.

---

## 27. Trigger / ORM / framework posture

Enforcement priority:

```text
PK/FK/UNIQUE/CHECK/partial unique
→ grants/column grants
→ owner-private conditional SQL + locks/CAS
→ trigger only on later proof that the invariant cannot be sustained otherwise
```

Launch baseline contains no semantic lifecycle trigger.

Explicit owner-private `database/sql` SQL is the current baseline because it keeps transaction/lock/CAS behavior visible. A codegen/query tool may be adopted later on a concrete evidence-backed benefit if owner-private SQL and accepted lock semantics remain explicit. No generic ORM/Repository framework is default authority.

---

## 28. Required proof obligations

T9/T10/T11 implementation proof must make at least these properties falsifiable:

```text
closed DB-object catalog completeness in both directions
no foreign first-party SQL
no raw first-party SQL against river.*
Go/static vocabulary == DDL CHECK vocabulary
metaldocs_verifier effective grants == metaldocs_runtime grants
serving cannot mutate immutable-history columns/classes
one current provider binding per User and issuer/subject; historical recovery works
UserProfile erasure leaves historical User references valid
RoleAssignment duplicate/scope checks
one open Revision / at most one EFFECTIVE Revision
Revision ordinal and Document code non-reuse
SUBMITTED iff current_submission_id
Revision EFFECTIVE/SUPERSEDED/OBSOLETE only through Release/effectivity path
DRAFT stale generation => zero mutation; successful mutation increments once
GovernanceAttempt Step label_snapshot remains unchanged after current route relabel
frozen Governance candidate FK blocks non-candidate and empty-set decision
one GovernanceAttempt per governed subject
Group deletion four live blockers and no historical fifth blocker
semantic content attachment cannot race GC delete
READY descriptor/malware evidence immutable under serving grants
live AdmissionClaim blocks GC
backup pin acquisition serializes against GC
GC phase 2 repeats all proofs immediately before provider delete
Idempotency key without Replay cannot commit
same-key READ COMMITTED winner/loser path does not poison Scope
idempotency HMAC rotation drain prevents honest-retry false conflict
River self-REINDEX disabled on PG16 while runtime is non-owner
Audit historical visibility filters before pagination
ApplicationSession resolve can reproduce the per-session CSRF token from server-side session state; constant-time validation requires the authenticated session and token alone grants nothing
Idempotency UUID textual variants map to one scoped key identity
Idempotency ReplaySnapshot >2048 bytes is structurally rejected
Idempotency post-expiry reuse succeeds correctly even when janitor cleanup lags
already-PDF required PDF rendition reuses the exact Submission handle/descriptor and creates zero provider copy/renderer/River intent; DOCX→PDF still enqueues the required intent
```

Controls whose firing only becomes meaningful after a future pooled-tenancy reopen are documented as seam structure, not claimed as current multi-tenant proof.

---

## 29. Ratified decision ledger

```text
D01  one PostgreSQL product-state database
D02  schemas authn/org/authz/controlled_docs/audit/platform/river
D03  fully-qualified first-party SQL; no search_path authority
D04  PostgreSQL-16-compatible persistence feature floor
D05  closed bidirectional DB-object ownership catalog; river.* THIRD_PARTY_MANAGED

D06  Role/Permission/bundles static; persist RoleAssignment only
D07  no Launch RLS/tenant GUC/pooled isolation substrate
D08  no persisted permission expansion/current-status/Search cache

D09  explicit monotonic BIGINT owner VersionToken
D10  WorkingContent BIGINT generation OCC
D11  Revision.state canonical lifecycle state
D12  immutable Release effectivity fact + partial one-EFFECTIVE barrier
D13  partial one-open DRAFT/SUBMITTED Revision barrier
D14  bounded current_submission_id with SUBMITTED biconditional and same-Revision FK

D15  closed relational governance persistence with configured Step label + immutable attempt label snapshot; no generic workflow
D16  one ACTIVE Step structural uniqueness
D17  live GROUP dependency separated from activated candidate snapshot
D18  GovernanceDecision must FK to frozen candidate snapshot; candidates mandatory for NAMED_USER/GROUP activation

D19  semantic exact-content descriptor copied into semantic owner rows
D20  ManagedContent mutable lifecycle/location state only
D21  immutable managed_content_descriptors technical proof relation
D22  immutable malware_inspections technical evidence
D23  AdmissionClaim created at claim-bound allocation and spans OPEN/READY to consume/release/expiry
D24  every semantic handle attachment locks ManagedContent FOR SHARE; GC uses FOR UPDATE
D25  two-phase GC repeats semantic/claim/pin proofs; downstream proof reads are non-locking
D26  backup pins have explicit persistence/locking lifecycle

D27  paired idempotency Key + Replay tables
D28  deferred reverse FK makes incomplete committed replay impossible
D29  no durable target IN_PROGRESS/FAILED replay state
D30  idempotency semantic fingerprint = HMAC-SHA-256 + positive key version; drain-before-rotation
D31  idempotency retention deletes Replay before Key

D32  River uses river.* and transaction-coupled River API insertion
D33  River self-REINDEX OFF on PG16; serving runtime never owns river.*
D34  no first-party raw SQL against river.*

D35  cross-owner FKs only stable identity/existence; no cross-owner semantic cascade
D36  Company singleton CHECK is Launch fail-closed interlock; company_id current semantics retained
D37  four DB trust classes: provisioner / owner / runtime / verifier with grant parity
D38  immutable-history serving restrictions enforced by grants/column grants
D39  every authenticated user-initiated semantic mutation protects actor User with FOR SHARE
D40  multiple User locks dedupe/sort UUID/acquire once at strongest mode
D41  Document FOR UPDATE is lifecycle/effectivity serialization root
D42  Area FOR SHARE vs lifecycle FOR UPDATE closes create-vs-retire race
D43  DocumentType FOR SHARE/UPDATE protects coherent current configuration snapshots
D44  zero semantic lifecycle trigger baseline
D45  explicit owner-private database/sql SQL is Launch baseline; generic ORM/repository rejected by default
D46  normal owner-private relational Search/query only; materialized Search OFF
D47  static Go/DDL vocabulary equality is a blocking Validation-Baseline obligation
D48  restore mechanism states converge through their own existing invariants; no generic recovery journal
```

---

## 30. Independent review convergence

```text
Round 1
  verdict: APPROVE GLOBAL MAXIMUM WITH MATERIAL FIXES
  BLOCKER 2 / MAJOR 11 / LOW 10
  Global Maximum CONFIRMED

Lead Round-1 adjudication
  both blockers accepted and corrected
  M7 Company/company_id subtraction rejected

Round 2 bounded delta
  BLOCKER 0 / MAJOR 7 / LOW 6
  both Round-1 blockers CLOSED
  Global Maximum CONFIRMED
  upstream reopen NO
  third full review NOT REQUIRED

Final Lead adjudication
  7/7 Round-2 MAJOR closed
  6/6 Round-2 LOW closed
  surviving material contradiction 0

Operator ratification
  explicit 2026-08-20
```

Reviewer evidence remained non-authoritative until adjudicated and ratified.

---

## 31. Stage boundary and next stage

```text
T8-D Persistence Realization = CLOSED / OPERATOR-RATIFIED / PROMOTED
T8-E Executable Wire Contract = ACTIVE
T8-F→T8-H                   = NOT OPEN
T9→T12                       = NOT OPEN
implementation               = BLOCKED
```

T8-E consumes T6 semantic journeys plus this T8-D physical/concurrency authority and freezes exact executable `/api/v1` OpenAPI contracts. It may not silently change T8-D persistent meaning or upstream semantics to simplify wire encoding.

No product implementation, target migration or T11 execution work is authorized by T8-D closure.
