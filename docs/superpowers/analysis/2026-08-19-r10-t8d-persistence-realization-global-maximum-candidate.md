# R10-T8D — Persistence Realization — Global Maximum Candidate

```text
GLOBAL MAXIMUM CANDIDATE
NON-AUTHORITATIVE STAGING
INDEPENDENT-REVIEW INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Revalidated baseline HEAD:** `b1d6e36b315b09a2eadee20a9089e1f9978b2a51`  
> **Stage:** T8-D ACTIVE  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Upstream contract authority:** `wiki/architecture/r10-t8c-internal-communication-contracts.md`  
> **Implementation:** BLOCKED

This is the operator-approved design candidate for R10 **T8-D — Persistence Realization**. It is staging evidence only until independent challenge, Lead adjudication and explicit operator ratification promote one durable T8-D authority.

It consumes Product Contract REV001, Whole-Product GCR/4+1 ownership and T1→T8-C. It must not redesign their semantic owners, lifecycle, Authorization, exact-content, async, API or internal-contract meaning by persistence convenience.

---

## 1. Exact T8-D question

> **What is the smallest PostgreSQL persistence realization that makes T1→T8-C invariants structurally enforceable, gives every persistent fact one semantic/mechanism owner, and realizes required ACID/OCC/serialization behavior without foreign SQL, duplicate truth, hidden shared write authority, wire leakage or speculative persistence?**

The candidate answer is:

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

No materially stronger design with lower total authority/complexity/proof cost survived the current Method pass.

---

## 2. Binding upstream laws

### Semantic homes remain exactly

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

Technical persistence may exist only for named current mechanisms:

```text
ManagedContent / AdmissionClaims / GC / backup pins
Idempotency
River
```

No Artifact, Approval, Template, Search, Workflow, Retention, Notifications or generic Records semantic owner is created.

### Transaction posture

Inherited from T2/T8-C:

```text
one native business transition
=
one local PostgreSQL ACID product-state transaction

PostgreSQL READ COMMITTED
+ narrow explicit serialization
+ OCC/CAS
+ structural constraints where required
```

Application owns `Runner.Within`; owners execute only owner-private SQL through the caller-provided `txscope.Scope`.

### Persistence ownership

```text
semantic SQL        = owner-private only
application SQL     = forbidden
transport SQL       = forbidden
platform semantic SQL = forbidden
cross-owner SQL as communication = forbidden
shared semantic repository = forbidden
```

Cross-owner composition stays application-routed through T8-C contracts even though the database may use narrow identity/existence FKs.

---

## 3. Reference-evidence baseline

The design intentionally uses PostgreSQL 16-compatible primitives only. Current repo runtime evidence uses PostgreSQL 16; T8-D does not require a database-version upgrade.

Load-bearing PostgreSQL properties were rechecked against official PostgreSQL 16 documentation before materialization:

```text
schemas/search_path:
  https://www.postgresql.org/docs/16/runtime-config-client.html

CREATE TABLE / NULLS NOT DISTINCT / DEFERRABLE FK:
  https://www.postgresql.org/docs/16/sql-createtable.html

unique/partial uniqueness:
  https://www.postgresql.org/docs/16/ddl-constraints.html

row locks:
  https://www.postgresql.org/docs/16/sql-select.html
  https://www.postgresql.org/docs/16/applevel-consistency.html

grants/column privileges:
  https://www.postgresql.org/docs/16/sql-grant.html

READ COMMITTED command snapshots:
  https://www.postgresql.org/docs/16/sql-set-transaction.html
```

River remains pinned by the repository at v0.37.1. Current River primary evidence confirms a configurable schema and transaction-coupled insertion capability; River physical tables remain third-party mechanism state, not MetalDocs authority.

External reference behavior informs mechanism feasibility only. Repository authority controls Product meaning.

---

## 4. Namespace strategy

### T8D-D01 — one PostgreSQL database

Select:

```text
ONE PostgreSQL product-state database
```

Reject database-per-owner/service splitting because T2 requires one local ACID transaction across current cross-owner application choreography.

### T8D-D02 — owner/mechanism schemas

Select:

```text
authn.*             Authentication
org.*               Organization
authz.*             Authorization
controlled_docs.*   Controlled Documents
audit.*             Audit
platform.*          MetalDocs technical mechanisms
river.*             River-owned third-party mechanism objects
```

The short physical names are deliberate. They keep fully-qualified SQL readable while the table-ownership catalog maps them unambiguously to the ratified semantic owners.

PostgreSQL schemas are namespaces inside one database, not microservice boundaries.

### T8D-D03 — fully-qualified first-party SQL

All first-party runtime SQL names target objects by explicit schema-qualified name:

```text
org.users
authz.role_assignments
controlled_docs.documents
audit.events
platform.managed_content
```

`search_path` is not an authority mechanism and is not relied on to choose a first-party relation.

### T8D-D04 — PostgreSQL feature floor

Target DDL/queries remain compatible with PostgreSQL 16 primitives. A later T8-G runtime version may be newer, but no T8-D invariant requires a post-16 feature.

### T8D-D05 — closed table ownership catalog

Every first-party base table/view is classified by exact schema/name and owner/mechanism class.

Verification law:

```text
live target object absent from catalog = FAIL
catalog object absent from target migration/schema set = FAIL
foreign first-party raw SQL ownership = FAIL
raw first-party SQL against river.* = FAIL
```

This preserves/refines the strongest useful property of the current table-ownership analyzer: completeness in both directions. The target catalog is rewritten around R10 owners rather than legacy modules.

---

## 5. Physical type policy

Use ordinary PostgreSQL primitives unless a stronger invariant requires otherwise:

```text
semantic/technical ids       UUID
versions/generations         BIGINT
revision ordinals/counters   BIGINT
trusted times                TIMESTAMPTZ
hashes/digests               BYTEA
bounded opaque replay        BYTEA
human/product codes          TEXT + owner/DB validation
closed states/enums          TEXT + CHECK, not PostgreSQL enum types by default
Audit bounded facts          JSONB + bounded object check
```

Avoid PostgreSQL enum types as the default because they add DDL coupling without improving current semantic authority. Closed vocabularies are still structurally checked.

All semantic UUIDs are application/server generated or database-generated through one accepted UUID primitive; UUID generation location is not business authority.

---

## 6. Persistent-state census

### PERSIST — Authentication

```text
ProviderSubjectBinding
ApplicationSession
```

No passwords, provider-role/group snapshots, local lockout policy, fresh-auth/e-sign state or session history are baseline persistence.

### PERSIST — Organization

```text
single Company root
User stable identity/current enabled eligibility
separately erasable UserProfile
Area
Group
GroupMembership
```

### PERSIST — Authorization

```text
RoleAssignment current grant truth
```

### STATIC / CODE AUTHORITY — Authorization

```text
Role vocabulary
Permission vocabulary
Role→Permission bundles
scope-compatibility vocabulary
```

No Role/Permission/reference tables are required for Launch authority.

### PERSIST — Controlled Documents

```text
DocumentType + independent mutable config versions
Governance route step configuration
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
exact-content descriptors on owning semantic rows
```

### PERSIST — Audit

```text
immutable AuditEvent
```

### PERSIST — platform mechanism

```text
ManagedContent technical state
AdmissionClaim
backup pin/exclusion
Idempotency key claim identity
completed ReplaySnapshot
```

### PERSIST — River

River-owned technical tables only.

### DERIVED / QUERY-ONLY — no duplicate rows

```text
Document official/catalog status
current EFFECTIVE lens
AuthorizedScopes
allowed_actions
Library
My Work
Governance work lists
Audit read pages
Search baseline
```

### DEFER / NOT LAUNCH

```text
RLS/tenant substrate
generic Artifact
WorkingSnapshot business history
EditorSession correctness state
custom Role/Permission tables
expanded permissions/materialized ACL
materialized Search
Search refresh jobs
generic Workflow/PolicyVersion
generic outbox/event log
notifications
Distribution/Periodic Review/Dossier/Evidence/Records persistence
```

---

## 7. Authentication tables

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

Structural rules:

```text
UNIQUE(issuer, subject)
unique current binding per User:
  UNIQUE partial index on user_id WHERE replaced_at IS NULL

replaced_at is the only mutable historical-binding field
issuer/subject/user/version/bound_at never change after INSERT
```

Provider-binding replacement:

```text
resolve provider subject outside semantic tx
→ protected actor
→ lock current binding
→ expected current version check
→ exact already-current = no-op/current token
→ otherwise terminate old binding with replaced_at
→ insert new binding with version+1
→ delete all current ApplicationSessions for User
→ required Audit
→ commit
```

Offboarding preserves the current ProviderSubjectBinding; User disabled state denies session issuance.

No standalone binding-disable API/state is introduced.

### `authn.application_sessions`

Material shape:

```text
id                  UUID PRIMARY KEY
user_id             UUID NOT NULL REFERENCES org.users(id)
token_digest        BYTEA NOT NULL UNIQUE
csrf_secret_digest  BYTEA NOT NULL
created_at          TIMESTAMPTZ NOT NULL
expires_at          TIMESTAMPTZ NOT NULL

CHECK(expires_at > created_at)
```

Session is current access state only:

```text
issue               INSERT
logout              DELETE
offboard            DELETE all for User
binding replacement DELETE all for User
restore readiness   remove all restored sessions before ordinary serving
```

No raw token, IP address, User-Agent, device label, tenant selector, `last_seen`, `revoked_at` history or permission snapshot is baseline authority.

---

## 8. Organization tables

### `org.companies`

```text
id             UUID PRIMARY KEY
singleton_key  SMALLINT NOT NULL UNIQUE CHECK(singleton_key = 1)
display_name   TEXT NOT NULL
version        BIGINT NOT NULL CHECK(version >= 1)
created_at     TIMESTAMPTZ NOT NULL
```

The singleton key makes the Launch single-Company root structural without reintroducing tenancy.

### `org.users`

```text
id                    UUID PRIMARY KEY
company_id            UUID NOT NULL REFERENCES org.companies(id)
enabled               BOOLEAN NOT NULL
eligibility_version   BIGINT NOT NULL CHECK(eligibility_version >= 1)
created_at             TIMESTAMPTZ NOT NULL
```

User rows are stable participant identity and receive no serving-runtime DELETE privilege.

### `org.user_profiles`

Minimum correctness-bearing shape:

```text
user_id        UUID PRIMARY KEY REFERENCES org.users(id)
display_name   TEXT NOT NULL
email          TEXT NULL
version        BIGINT NOT NULL CHECK(version >= 1)
```

UserProfile is separately erasable. Additional bounded profile presentation fields may be named by T8-E only if already authorized by Product/T6; they remain columns under this erasable owner row and do not change T8-D ownership/erasure semantics.

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

Area code is immutable after creation; runtime receives no UPDATE privilege for `code`.

### `org.groups`

```text
id          UUID PRIMARY KEY
company_id  UUID NOT NULL REFERENCES org.companies(id)
name        TEXT NOT NULL
version     BIGINT NOT NULL CHECK(version >= 1)
created_at  TIMESTAMPTZ NOT NULL
```

Group may be deleted only after semantic dependency resolution and successful no-action/restrict FK checks.

### `org.group_memberships`

```text
company_id  UUID NOT NULL
 group_id   UUID NOT NULL REFERENCES org.groups(id)
 user_id    UUID NOT NULL REFERENCES org.users(id)
created_at  TIMESTAMPTZ NOT NULL

PRIMARY KEY(group_id, user_id)
```

Same-Company eligibility is revalidated by Organization; persistent `company_id` keeps the scope explicit and is checked against the referenced identities through target composite constraints where practical. No membership-history table is introduced.

---

## 9. Authorization persistence

### T8D-D06 — persist RoleAssignment only

`authz.role_assignments`:

```text
id          UUID PRIMARY KEY
company_id  UUID NOT NULL REFERENCES org.companies(id)
user_id     UUID NULL REFERENCES org.users(id)
group_id    UUID NULL REFERENCES org.groups(id)
role_code   TEXT NOT NULL
area_id     UUID NULL REFERENCES org.areas(id)
created_at  TIMESTAMPTZ NOT NULL
```

Checks:

```text
exactly one of user_id / group_id is non-null
role_code in static Launch Role vocabulary
area_id NULL     = Company scope
area_id non-null = Area scope

governance_admin requires Company scope
area_manager requires Area scope
```

Duplicate prevention uses two unique indexes, one for User subjects and one for Group subjects. Company scope NULL is treated as a real duplicate scope using PostgreSQL `NULLS NOT DISTINCT` semantics.

No database rows exist for Role, Permission or Role→Permission expansion.

### T8D-D07 — no Launch RLS/tenant substrate

The current tenant/RLS/GUC design is not preserved in the target. Stable Company identity remains; pooled tenancy is a future reopen seam.

### T8D-D08 — no persisted permission/search/status cache

Do not persist effective permissions, `Document.current_status`, `allowed_actions`, Search result state or materialized discovery authority.

---

## 10. DocumentType/configuration persistence

### `controlled_docs.document_types`

```text
id                          UUID PRIMARY KEY
company_id                  UUID NOT NULL REFERENCES org.companies(id)
code                        TEXT NOT NULL
name                        TEXT NOT NULL
active                      BOOLEAN NOT NULL
numbering_scope             TEXT NOT NULL CHECK(numbering_scope IN ('DOCUMENT_TYPE','DOCUMENT_TYPE_AREA'))
governance_mode             TEXT NOT NULL CHECK(governance_mode IN ('NO_HUMAN_APPROVAL','USE_GOVERNANCE_ROUTE'))
representation_mode         TEXT NOT NULL CHECK(representation_mode IN ('SOURCE_ONLY','REQUIRE_OFFICIAL_RENDITION_PDF'))
base_version                BIGINT NOT NULL CHECK(base_version >= 1)
governance_version          BIGINT NOT NULL CHECK(governance_version >= 1)
eligible_templates_version  BIGINT NOT NULL CHECK(eligible_templates_version >= 1)
created_at                  TIMESTAMPTZ NOT NULL

UNIQUE(company_id, code)
```

Normalized code is server-produced and DB-checked for the ratified uppercase ASCII-alphanumeric/no-`-` law. Exact wire maximum length remains T8-E because T6 explicitly leaves that bound there.

`code` and `numbering_scope` may change only before the first committed Document uses the type. This conditional semantic rule is enforced by owner-private SQL under a DocumentType lock, not a trigger.

Three version columns deliberately prevent false ETag conflicts between base configuration, governance/representation configuration and eligible-template replacement.

### `controlled_docs.document_type_governance_steps`

```text
document_type_id  UUID NOT NULL REFERENCES controlled_docs.document_types(id)
ordinal           BIGINT NOT NULL CHECK(ordinal >= 0)
selector_kind     TEXT NOT NULL CHECK(selector_kind IN ('NAMED_USER','GROUP'))
named_user_id     UUID NULL REFERENCES org.users(id)
group_id          UUID NULL REFERENCES org.groups(id)

PRIMARY KEY(document_type_id, ordinal)
```

CHECK enforces exactly the selector column corresponding to `selector_kind`.

Current route GROUP references are real FKs so Group deletion fails closed while current configuration depends on the Group.

### `controlled_docs.document_type_eligible_templates`

```text
document_type_id     UUID NOT NULL REFERENCES controlled_docs.document_types(id)
template_document_id UUID NOT NULL REFERENCES controlled_docs.documents(id)

PRIMARY KEY(document_type_id, template_document_id)
```

FK proves identity only. Template role/current-effectivity/eligibility remain Controlled Documents semantics and are validated in owner-private SQL.

### `controlled_docs.document_number_counters`

```text
id                UUID PRIMARY KEY
document_type_id  UUID NOT NULL REFERENCES controlled_docs.document_types(id)
area_id           UUID NULL REFERENCES org.areas(id)
next_value        BIGINT NOT NULL CHECK(next_value >= 1)
```

Unique logical counter:

```text
UNIQUE NULLS NOT DISTINCT(document_type_id, area_id)
```

`area_id=NULL` is DocumentType-wide numbering. Area-id row is DocumentType+Area numbering.

Preview never updates this table. Document create allocates and advances the counter inside the business transaction. PostgreSQL sequences are deliberately not used because the current product does not require a separate non-transactional allocation primitive; gaps remain allowed through rollback/operational realities without making sequence state authority.

---

## 11. Document root and origin

### `controlled_docs.documents`

```text
id                         UUID PRIMARY KEY
company_id                 UUID NOT NULL REFERENCES org.companies(id)
document_type_id           UUID NOT NULL REFERENCES controlled_docs.document_types(id)
area_id                    UUID NOT NULL REFERENCES org.areas(id)
code                       TEXT NOT NULL
responsible_user_id        UUID NOT NULL REFERENCES org.users(id)
responsible_owner_version  BIGINT NOT NULL CHECK(responsible_owner_version >= 1)
is_template                BOOLEAN NOT NULL
template_role_version      BIGINT NOT NULL CHECK(template_role_version >= 1)
created_at                 TIMESTAMPTZ NOT NULL

UNIQUE(company_id, code)
```

No `current_revision_id`, `current_release_id`, `current_status` or `latest_submission_id` is persisted.

Document row is still the T2 lifecycle serialization root. Its stable identity/type/Area/code are not rewritten by lifecycle transitions.

Responsible-owner and Template-role replacements use independent version counters.

Document rows have no serving-runtime DELETE path in Launch; committed codes therefore cannot rebind through deletion/reuse.

### `controlled_docs.document_origins`

Only template-derived Documents have a row:

```text
document_id                  UUID PRIMARY KEY REFERENCES controlled_docs.documents(id)
source_template_document_id  UUID NOT NULL REFERENCES controlled_docs.documents(id)
source_template_revision_id  UUID NOT NULL REFERENCES controlled_docs.revisions(id)
source_sha256                BYTEA NOT NULL CHECK(octet_length(source_sha256)=32)
source_size_bytes            BIGINT NOT NULL CHECK(source_size_bytes >= 0)
source_content_format        TEXT NOT NULL
created_at                   TIMESTAMPTZ NOT NULL
```

Insert-only. No source Submission is required; no provider locator is semantic provenance.

---

## 12. Revision lifecycle

### T8D-D09 — explicit owner VersionTokens

Whole-replacement mutable resources use explicit monotonic `BIGINT` versions. Do not use `xmin`, `ctid`, timestamps or JSON hashes as concurrency authority.

Law:

```text
initial version = 1
expected != current -> stale / zero mutation
material replacement -> version+1 exactly once
exact already-current repeat -> no mutation/no Audit/same current version
```

### T8D-D10 — WorkingContent generation

DRAFT title+source share explicit monotonic `BIGINT generation`, initial zero.

### T8D-D11 — Revision.state is canonical lifecycle state

`controlled_docs.revisions`:

```text
id                    UUID PRIMARY KEY
document_id           UUID NOT NULL REFERENCES controlled_docs.documents(id)
ordinal               BIGINT NOT NULL CHECK(ordinal >= 0)
title                 TEXT NOT NULL
state                 TEXT NOT NULL CHECK(state IN ('DRAFT','SUBMITTED','EFFECTIVE','SUPERSEDED','CANCELLED','OBSOLETE'))
current_submission_id UUID NULL
created_at            TIMESTAMPTZ NOT NULL

UNIQUE(document_id, ordinal)
```

Partial unique structural barriers:

```text
one open business Revision per Document:
  UNIQUE(document_id) WHERE state IN ('DRAFT','SUBMITTED')

at most one EFFECTIVE Revision per Document:
  UNIQUE(document_id) WHERE state='EFFECTIVE'
```

`SUBMITTED` iff `current_submission_id` is present. A composite FK binds the current Submission to the same Revision.

Revision rows are never serving-runtime deleted, so ordinal uniqueness plus no delete makes ordinal reuse impossible.

This intentionally rejects the explored release-chain-only state model. T2 already owns explicit Revision transitions (`EFFECTIVE→SUPERSEDED`, `SUBMITTED→EFFECTIVE`, `EFFECTIVE→OBSOLETE`); persisting `Revision.state` realizes that semantic authority. What remains forbidden is a second `Document.current_status` cache.

### T8D-D12 — Release fact plus structural effectivity barrier

Release remains a distinct immutable fact; `Revision.state` expresses lifecycle, while Release proves exactly which Submission established effectivity.

### T8D-D13 — one open Revision

The partial unique open-revision index structurally enforces the Launch singular current DRAFT/SUBMITTED cycle.

### T8D-D14 — current Submission is bounded Revision relation

`current_submission_id` exists only while state is `SUBMITTED`; RETURN/withdraw/release/cancel clear it. Historical Submission rows remain immutable.

---

## 13. WorkingContent

`controlled_docs.working_contents`:

```text
revision_id         UUID PRIMARY KEY REFERENCES controlled_docs.revisions(id)
generation          BIGINT NOT NULL CHECK(generation >= 0)
managed_content_id  UUID NOT NULL REFERENCES platform.managed_content(id)
sha256              BYTEA NOT NULL CHECK(octet_length(sha256)=32)
size_bytes          BIGINT NOT NULL CHECK(size_bytes >= 0)
content_format      TEXT NOT NULL
updated_at          TIMESTAMPTZ NOT NULL
```

DRAFT update performs owner-private conditional SQL equivalent to:

```text
WHERE revision_id = target AND generation = expected
→ atomically change accepted title/source facts
→ generation = generation+1 exactly once
```

If the CAS affects zero rows, the entire semantic mutation returns stale and exposes zero mutation.

Title remains stored on Revision, but any title mutation participates in the same transaction and generation CAS. There is no second title version.

WorkingContent remains through SUBMITTED so RETURN/withdraw can reopen the same Revision. It may be removed when Revision reaches a terminal noneditable state such as Release or cancellation; immutable Submission/Release history remains authoritative.

---

## 14. Submission, withdrawal and cancellation

### `controlled_docs.submissions`

```text
id                        UUID PRIMARY KEY
revision_id               UUID NOT NULL REFERENCES controlled_docs.revisions(id)
submitter_user_id         UUID NOT NULL REFERENCES org.users(id)
submitted_at              TIMESTAMPTZ NOT NULL
title_snapshot            TEXT NOT NULL
managed_content_id        UUID NOT NULL REFERENCES platform.managed_content(id)
sha256                    BYTEA NOT NULL CHECK(octet_length(sha256)=32)
size_bytes                BIGINT NOT NULL CHECK(size_bytes >= 0)
content_format            TEXT NOT NULL
governance_mode_snapshot  TEXT NOT NULL
representation_snapshot   TEXT NOT NULL
required_rendition_format TEXT NULL

UNIQUE(revision_id, id)
```

Serving runtime privilege:

```text
SELECT + INSERT only
NO UPDATE / DELETE / TRUNCATE
```

Submission is the immutable exact governed attempt. Re-submit creates another row.

`Revision.current_submission_id` references `(revision_id,id)` from the same Revision.

### `controlled_docs.submission_withdrawals`

```text
submission_id   UUID PRIMARY KEY REFERENCES controlled_docs.submissions(id)
actor_user_id   UUID NOT NULL REFERENCES org.users(id)
withdrawn_at    TIMESTAMPTZ NOT NULL
```

Insert-only. Withdrawal does not mutate the old Submission.

### `controlled_docs.revision_cancellations`

```text
revision_id      UUID PRIMARY KEY REFERENCES controlled_docs.revisions(id)
actor_user_id    UUID NOT NULL REFERENCES org.users(id)
reason           TEXT NOT NULL
cancelled_at     TIMESTAMPTZ NOT NULL
```

Insert-only. Revision transition to `CANCELLED` is current lifecycle state; cancellation reason/actor/time remain immutable evidence.

---

## 15. Governance persistence

### T8D-D15 — closed relational governance model

Select a MetalDocs-specific relational shape. Reject JSON route blobs as the sole relational truth and reject generic workflow/state-machine schemas.

Tables:

```text
controlled_docs.governance_attempts
controlled_docs.governance_attempt_steps
controlled_docs.governance_group_dependencies
controlled_docs.governance_step_candidates
controlled_docs.governance_decisions
controlled_docs.submission_feedback
```

### `controlled_docs.governance_attempts`

```text
id                       UUID PRIMARY KEY
submission_id            UUID NULL REFERENCES controlled_docs.submissions(id)
obsolescence_request_id  UUID NULL REFERENCES controlled_docs.obsolescence_requests(id)
state                    TEXT NOT NULL CHECK(state IN ('ACTIVE','COMPLETED','RETURNED','WITHDRAWN','CANCELLED'))
created_at               TIMESTAMPTZ NOT NULL
ended_at                 TIMESTAMPTZ NULL
```

Exactly one governed subject is non-null. No generic `subject_type/subject_id` pair exists.

`CANCELLED` represents termination caused by Revision cancellation without fabricating a participant verdict.

### `controlled_docs.governance_attempt_steps`

```text
id                 UUID PRIMARY KEY
attempt_id         UUID NOT NULL REFERENCES controlled_docs.governance_attempts(id)
ordinal            BIGINT NOT NULL CHECK(ordinal >= 0)
selector_kind      TEXT NOT NULL CHECK(selector_kind IN ('NAMED_USER','GROUP'))
named_user_id      UUID NULL REFERENCES org.users(id)
group_id_snapshot  UUID NULL
state              TEXT NOT NULL CHECK(state IN ('PENDING','ACTIVE','DECIDED'))
activated_at       TIMESTAMPTZ NULL

UNIQUE(attempt_id, ordinal)
```

Selector snapshot columns are immutable after INSERT. State/activation fields are the bounded mutable execution facts.

### T8D-D16 — one ACTIVE Step

Structural partial uniqueness:

```text
UNIQUE(attempt_id) WHERE state='ACTIVE'
```

Two active human Steps in one attempt are uncommittable.

### T8D-D17 — live GROUP dependency vs historical candidate snapshot

`controlled_docs.governance_group_dependencies`:

```text
step_id   UUID PRIMARY KEY REFERENCES controlled_docs.governance_attempt_steps(id)
group_id  UUID NOT NULL REFERENCES org.groups(id)
```

Row exists only while a GROUP-selector Step remains unactivated in a live attempt. This FK is one of the structural Group-delete blockers required by T3.

On activation:

```text
resolve enabled Group members in same Scope
→ insert frozen candidates
→ delete live group dependency
→ mark Step ACTIVE
```

`group_id_snapshot` remains historical identifier after activation but deliberately has no permanent FK to Group, so activated/completed history alone does not keep Group identity alive.

`controlled_docs.governance_step_candidates`:

```text
step_id   UUID NOT NULL REFERENCES controlled_docs.governance_attempt_steps(id)
user_id   UUID NOT NULL REFERENCES org.users(id)

PRIMARY KEY(step_id,user_id)
```

Candidates are insert-only after activation. The same table may hold the resolved candidate for NAMED_USER Steps as well, giving one exact active-candidate authority.

### `controlled_docs.governance_decisions`

```text
id             UUID PRIMARY KEY
step_id        UUID NOT NULL UNIQUE REFERENCES controlled_docs.governance_attempt_steps(id)
actor_user_id  UUID NOT NULL REFERENCES org.users(id)
outcome        TEXT NOT NULL CHECK(outcome IN ('ACCEPT','RETURN_FOR_CHANGES'))
reason         TEXT NULL
decided_at     TIMESTAMPTZ NOT NULL
```

RETURN_FOR_CHANGES requires a nonblank reason. Decision rows are insert-only.

### `controlled_docs.submission_feedback`

```text
id             UUID PRIMARY KEY
attempt_id     UUID NOT NULL REFERENCES controlled_docs.governance_attempts(id)
actor_user_id  UUID NOT NULL REFERENCES org.users(id)
body           TEXT NOT NULL
created_at     TIMESTAMPTZ NOT NULL
```

Insert-only. Feedback remains semantic evidence and is not copied into Audit by convenience.

### T8D-D18 — no generic workflow persistence

No generic transition table, workflow definition engine, quorum engine, polymorphic domain-resource registry or event-sourced workflow log is created.

---

## 16. Release/effectivity

`controlled_docs.releases`:

```text
id                       UUID PRIMARY KEY
document_id              UUID NOT NULL REFERENCES controlled_docs.documents(id)
revision_id              UUID NOT NULL REFERENCES controlled_docs.revisions(id)
submission_id            UUID NOT NULL REFERENCES controlled_docs.submissions(id)
predecessor_revision_id  UUID NULL REFERENCES controlled_docs.revisions(id)
released_at              TIMESTAMPTZ NOT NULL

UNIQUE(revision_id)
UNIQUE(submission_id)
```

A partial unique index ensures at most one first Release row per Document (`predecessor_revision_id IS NULL`). Another unique index on non-null predecessor ensures one successor branch from a released predecessor.

Owner-private composite checks/FKs prove winning Revision belongs the Document, Submission belongs the winning Revision, and predecessor belongs the same Document and has prior Release truth.

Release rows are insert-only.

Replacement transaction, under `Document FOR UPDATE`:

```text
load exact current EFFECTIVE predecessor
→ validate successor current SUBMITTED Revision/current Submission
→ predecessor Revision EFFECTIVE→SUPERSEDED
→ successor Revision SUBMITTED→EFFECTIVE
→ successor current_submission_id=NULL
→ INSERT Release(predecessor_revision_id)
→ required Audit
→ commit
```

The partial unique `state='EFFECTIVE'` index is the final database barrier against dual effectivity.

No current Release/Revision pointer is added to Document.

---

## 17. OfficialRendition

`controlled_docs.official_renditions`:

```text
id                  UUID PRIMARY KEY
submission_id       UUID NOT NULL REFERENCES controlled_docs.submissions(id)
required_format     TEXT NOT NULL
managed_content_id  UUID NOT NULL REFERENCES platform.managed_content(id)
sha256              BYTEA NOT NULL CHECK(octet_length(sha256)=32)
size_bytes          BIGINT NOT NULL CHECK(size_bytes >= 0)
content_format      TEXT NOT NULL
created_at          TIMESTAMPTZ NOT NULL

UNIQUE(submission_id, required_format)
```

Insert-only. No renderer/provider job id or provider locator becomes semantic identity.

Finalization occurs only after exact READY/admission/malware/eligibility reproof. Duplicate at-least-once executions cannot create two semantic renditions for the same Submission+format.

---

## 18. Obsolescence

`controlled_docs.obsolescence_requests`:

```text
id                        UUID PRIMARY KEY
document_id               UUID NOT NULL REFERENCES controlled_docs.documents(id)
target_revision_id        UUID NOT NULL REFERENCES controlled_docs.revisions(id)
initiator_user_id         UUID NOT NULL REFERENCES org.users(id)
reason                    TEXT NOT NULL
governance_mode_snapshot  TEXT NOT NULL
state                     TEXT NOT NULL CHECK(state IN ('ACTIVE','RETURNED','WITHDRAWN','COMPLETED'))
requested_at              TIMESTAMPTZ NOT NULL
ended_at                  TIMESTAMPTZ NULL
```

One active obsolescence intent:

```text
UNIQUE(document_id) WHERE state='ACTIVE'
```

Request target/initiator/reason/snapshot/requested time are immutable after creation. Only bounded state/ended-at progression is mutable.

NoHumanApproval may insert+complete inside the same business transaction. Human-governed completion occurs through final Governance Step decision.

Completion under Document serialization:

```text
request ACTIVE→COMPLETED
exact target Revision EFFECTIVE→OBSOLETE
attempt ACTIVE→COMPLETED when human-governed
required Audit
commit
```

RETURN/WITHDRAW leave target EFFECTIVE.

No generic second obsolescence workflow family is created.

---

## 19. Exact content and ManagedContent

### T8D-D19 — semantic descriptor persistence

WorkingContent, Submission, OfficialRendition and DocumentOrigin own their exact descriptors directly:

```text
sha256 BYTEA(semantic check length=32)
size_bytes BIGINT >= 0
content_format closed TEXT vocabulary
managed_content_id UUID mechanism handle
```

Provider identity never replaces semantic descriptor truth.

### `platform.managed_content`

```text
id                  UUID PRIMARY KEY
state               TEXT NOT NULL CHECK(state IN ('OPEN','READY','GC_PENDING'))
provider_locator    TEXT NOT NULL
trust_class         TEXT NOT NULL CHECK(trust_class IN ('UNTRUSTED_EXTERNAL','TRUSTED_MANAGED_COPY','TRUSTED_INTERNAL_DERIVATION'))
sha256              BYTEA NULL
size_bytes          BIGINT NULL
content_format      TEXT NULL
malware_verdict     TEXT NULL CHECK(malware_verdict IN ('CLEAN','MALICIOUS'))
malware_digest      BYTEA NULL
malware_checked_at  TIMESTAMPTZ NULL
created_at          TIMESTAMPTZ NOT NULL
ready_at            TIMESTAMPTZ NULL
gc_pending_at       TIMESTAMPTZ NULL
```

READY requires complete server-derived descriptor. Malware digest, when present, is exactly 32 bytes. Provider migration may change provider locator without changing handle or semantic descriptors.

There is no `owner_type/owner_id`, Document/Revision/Submission owner field, provider ETag semantic identity or generic retention root.

Create-once/no-overwrite is enforced by provider primitive/conformance proof plus target write path; no serving operation can overwrite a handle.

### T8D-D20 — mechanism state only

`platform.managed_content` remains technical state; references from semantic rows determine governed preservation.

---

## 20. Admission claims and backup pins

### T8D-D21 — row-existence AdmissionClaim

`platform.admission_claims`:

```text
id                  UUID PRIMARY KEY
managed_content_id  UUID NOT NULL UNIQUE REFERENCES platform.managed_content(id)
created_at          TIMESTAMPTZ NOT NULL
expires_at          TIMESTAMPTZ NOT NULL

CHECK(expires_at > created_at)
```

Semantics:

```text
row exists = live opaque claim
Reserve     = INSERT after proving handle READY
ProveLive   = SELECT
ConsumeIn   = DELETE in the semantic attachment transaction
Release     = DELETE
expiry      = bounded mechanism cleanup
```

If attachment transaction rolls back, claim DELETE rolls back and the claim stays live.

### `platform.managed_content_backup_pins`

```text
backup_pin_id      UUID NOT NULL
managed_content_id UUID NOT NULL REFERENCES platform.managed_content(id)
expires_at         TIMESTAMPTZ NOT NULL

PRIMARY KEY(backup_pin_id, managed_content_id)
```

This is recovery/GC exclusion state only, not a Backup semantic aggregate.

### T8D-D22 — two-phase GC with repeated proof

GC serialization root is `platform.managed_content` row.

Phase 1:

```text
ManagedContent FOR UPDATE
→ require state READY
→ ControlledDocs proves no current WorkingContent ref
→ proves no immutable Submission/Rendition/imported governed ref
→ prove no live AdmissionClaim
→ prove no backup pin
→ state=GC_PENDING
→ commit
```

New AdmissionClaim reservation only succeeds while handle is READY, so no new in-flight attachment can begin after GC_PENDING commits.

Phase 2 immediately before provider deletion:

```text
ManagedContent FOR UPDATE
→ still GC_PENDING
→ repeat FULL ControlledDocs semantic-reference proof
→ repeat live-claim proof
→ repeat backup-pin proof
→ commit
→ provider DeleteReclaimable outside semantic tx
→ after confirmed provider absence, finalize/remove technical row
```

Safe failure remains leaked storage, never deleted governed truth.

---

## 21. Idempotency persistence

### T8D-D23 — paired Key + Replay

`platform.idempotency_keys`:

```text
id                    UUID PRIMARY KEY
actor_user_id         UUID NOT NULL REFERENCES org.users(id)
operation_id          TEXT NOT NULL
key                   TEXT NOT NULL
semantic_fingerprint  BYTEA NOT NULL CHECK(octet_length(semantic_fingerprint)=32)
created_at            TIMESTAMPTZ NOT NULL
expires_at            TIMESTAMPTZ NOT NULL

UNIQUE(actor_user_id, operation_id, key)
CHECK(expires_at > created_at)
```

Application derives the 32-byte fingerprint deterministically from canonical validated semantic command fields; platform treats it as opaque equality material. It is not a copy of raw HTTP bytes or erasable profile data.

`platform.idempotency_replays`:

```text
key_id            UUID PRIMARY KEY
snapshot_version  INTEGER NOT NULL CHECK(snapshot_version >= 1)
payload           BYTEA NOT NULL CHECK(octet_length(payload) <= 65536)
completed_at      TIMESTAMPTZ NOT NULL
```

Payload is opaque, versioned, self-contained and PII-free per T8-C. It is not raw HTTP response bytes.

### T8D-D24 — incomplete replay is structurally uncommittable

Two referential constraints form one transactional completion invariant:

```text
idempotency_replays.key_id
  → idempotency_keys.id

idempotency_keys.id
  → idempotency_replays.key_id
  DEFERRABLE INITIALLY DEFERRED
```

The winner may insert the key first and satisfy the deferred reverse FK by inserting completed ReplaySnapshot before commit.

If `CompleteIn` is omitted or fails:

```text
key row exists at COMMIT
+ no Replay row
→ COMMIT fails
→ business mutation/Audit/intents roll back with the same Scope
```

Thus database structure reinforces the T6/T8-C law:

```text
successful idempotent semantic fact commit
<=>
completed ReplaySnapshot commit
```

The scoped unique key used for `ON CONFLICT DO NOTHING` remains immediate/non-deferrable; the completion FK is the deferred constraint.

Retention cleanup removes key+replay together inside one technical transaction.

### T8D-D25 — no durable target IN_PROGRESS/FAILED state

No `status`, public/durable IN_PROGRESS business state or target `FailReplay` row exists. A winning key row is uncommitted while the business transition is in progress.

Concurrent same-key realization under ratified READ COMMITTED:

```text
INSERT key ... ON CONFLICT DO NOTHING
→ winner continues in its Scope
→ loser waits on unique-key outcome
winner commit:
  loser receives DO NOTHING result
  later command sees completed key+Replay
winner rollback:
  contender may become key owner
same key/different fingerprint:
  conflict / zero business mutation
```

A realization that aborts an expected same-key loser transaction does not satisfy T8-C.

---

## 22. River persistence

### T8D-D26 — `river.*` is third-party managed

River-owned technical objects live under configured schema:

```text
river.*
```

They are not enumerated as MetalDocs semantic tables.

### T8D-D27 — no first-party raw River SQL

Only `platform/river` communicates through River's supported Go API/driver/migrator. First-party direct SQL against `river.*` is a verification failure.

Required OfficialRendition job is inserted through River's transaction-coupled API using the same native `*sql.Tx` obtained via T8-C `txscope.SQLTx`.

River job state remains execution mechanism only; it never establishes Submission, OfficialRendition, Release or governance truth.

---

## 23. Audit persistence

`audit.events`:

```text
id                 UUID PRIMARY KEY
occurred_at        TIMESTAMPTZ NOT NULL
actor_kind         TEXT NOT NULL CHECK(actor_kind IN ('USER','SYSTEM'))
actor_user_id      UUID NULL REFERENCES org.users(id)
system_actor_code  TEXT NULL
operation_code     TEXT NOT NULL
resource_kind      TEXT NOT NULL
resource_id        UUID NOT NULL
company_id         UUID NOT NULL REFERENCES org.companies(id)
area_id            UUID NULL REFERENCES org.areas(id)
facts              JSONB NOT NULL DEFAULT '{}'
correlation_id     UUID NULL
```

Checks:

```text
USER actor   -> actor_user_id present / system_actor_code absent
SYSTEM actor -> system_actor_code present / actor_user_id absent
jsonb_typeof(facts)='object'
octet_length(facts::text) <= 65536
```

Visibility meaning:

```text
area_id NULL     = Company historical visibility attribution
area_id non-null = exact historical Area attribution
```

Current resource relocation never rewrites this row.

Audit runtime grants:

```text
SELECT + INSERT
NO UPDATE / DELETE / TRUNCATE
```

The current implementation's append-only DB-grant property and bounded Audit payload property are selectively preserved/refined. Current global sequence/hash-chain columns are deleted because T1 explicitly rejects a global AuditChainHead/hash-chain Launch requirement.

---

## 24. Cross-owner FK policy

### T8D-D28 — identity/existence only

A cross-owner FK is allowed only when it proves stable identity/existence, not another owner's current business permission/lifecycle state.

Accepted classes include:

```text
authn session/binding              → org User
authz RoleAssignment               → org User/Group/Area/Company
ControlledDocs area/responsibility → org Area/User
current Governance selectors       → org User/Group
activated governance actors        → org User
Audit actor/visibility             → org User/Company/Area
semantic exact-content handles     → platform ManagedContent
Idempotency actor                  → org User
```

Database FK never decides:

```text
User enabled?
actor authorized?
Document current EFFECTIVE?
Template currently eligible?
Group deletion semantically allowed?
```

Those remain owner/app contract decisions.

### T8D-D29 — no cross-owner cascades

No cross-owner FK uses semantic `ON DELETE CASCADE` to mutate another owner. Delete either fails/restricts until dependencies are explicitly resolved or the referenced stable identity is nondeletable in Launch.

Same-mechanism technical cleanup may delete its own rows explicitly; cascade is not an architectural communication channel.

---

## 25. Runtime DB identity and immutable-history privileges

### T8D-D30 — serving runtime role != DDL owner

Select two trust classes:

```text
DDL/migration owner
serving runtime DB role
```

Serving runtime receives no schema/table DDL power and no blanket object ownership.

### T8D-D31 — DB grants enforce append-only classes

Examples:

```text
controlled_docs.submissions            SELECT, INSERT
controlled_docs.submission_withdrawals SELECT, INSERT
controlled_docs.revision_cancellations SELECT, INSERT
controlled_docs.governance_decisions   SELECT, INSERT
controlled_docs.submission_feedback    SELECT, INSERT
controlled_docs.releases               SELECT, INSERT
controlled_docs.official_renditions    SELECT, INSERT
controlled_docs.document_origins       SELECT, INSERT
audit.events                           SELECT, INSERT
```

No serving-runtime UPDATE/DELETE/TRUNCATE on immutable tables.

Mutable tables receive only required DML; where practical, column-level UPDATE grants prevent mutation of immutable snapshot columns while allowing bounded lifecycle/version columns.

Example:

```text
governance_attempt_steps
  selector snapshot columns = no UPDATE
  state/activated_at         = bounded UPDATE
```

This is structural defense, not a substitute for owner semantic validation.

### T8D-D32 — zero semantic lifecycle triggers baseline

Use enforcement order:

```text
PK/FK/NOT NULL/UNIQUE/CHECK/partial unique
→ DB privileges
→ owner-private conditional SQL + row locks/CAS
→ trigger only if later material evidence proves the first three cannot sustain an invariant
```

No Launch lifecycle business rule is hidden in PL/pgSQL triggers by default.

---

## 26. Concurrency/version realization

### Protected actor lock

T8-C `ProtectedSecuritySubjectIn` maps to a current `org.users` row lock that prevents concurrent offboarding/eligibility mutation for the Scope lifetime without serializing independent protected actions against each other.

Selected primitive:

```text
SELECT ... FROM org.users ... FOR SHARE
```

### T8D-D33 — protected actor for authenticated mutations

Every authenticated semantic mutation acquires protected current actor User `FOR SHARE` before relying on current User eligibility.

This slightly strengthens the minimum T3 census while using the same low-contention shared row lock. SYSTEM transitions have no actor lock.

### T8D-D34 — User offboarding/eligibility root

Offboarding/eligibility mutation serializes on target User with update-strength row lock/update. If protected action locks first, action may commit before offboarding; if offboarding locks/commits first, later protected action observes disabled User and fails closed.

### T8D-D35 — Document lifecycle root

Every operation capable of changing one Document's lifecycle/effectivity truth obtains the stable `controlled_docs.documents` row `FOR UPDATE` before lifecycle mutation.

Includes:

```text
next Revision
SUBMIT
RETURN_FOR_CHANGES
withdraw
cancel
Step decision/advance/final completion
Release
obsolescence initiation/completion/withdrawal
OfficialRendition finalization when Release eligibility may change
```

DRAFT editing need not take exclusive Document lifecycle lock; it uses Document shared relationship protection plus WorkingContent CAS.

### T8D-D36 — DocumentType config serialization

Consumers freezing current DocumentType configuration (e.g. SUBMIT/obsolescence snapshot) take DocumentType `FOR SHARE`. Whole configuration replacement takes update/exclusive row lock and version CAS.

This resolves concurrent snapshot/edit to coherent old-or-new configuration, never a mixed route.

### T8D-D37 — deterministic lock ordering

When one transaction needs multiple rows from the same lock class, acquire them in ascending stable UUID order.

Cross-class default order for current operations:

```text
1 idempotency scoped key claim where T6 requires it
2 actor/target User eligibility locks
3 Organization mutable roots needed for current eligibility/config
4 DocumentType root
5 Document lifecycle root(s)
6 owner-local child/current-state rows/counters
7 ManagedContent technical row/claim where participating
8 append-only Audit / River insertions
```

An operation may omit classes it does not need. FK internal key-share locking is not treated as cross-owner semantic communication.

Deadlock `40P01` remains a real PostgreSQL retryable failure class; architecture reduces cycles but does not claim mathematical elimination.

---

## 27. Material transaction matrix

The exact owner-private statements may vary while preserving this matrix.

### Session issuance

```text
external OIDC preflight complete
→ User protected FOR SHARE
→ verify enabled + current ProviderSubjectBinding
→ INSERT ApplicationSession
→ commit
```

### User create — Idempotency-Key

```text
BeginIn key claim
→ protected current actor
→ current AuthZ
→ INSERT org.User
→ INSERT org.UserProfile
→ INSERT authn.ProviderSubjectBinding
→ required Audit
→ CompleteIn ReplaySnapshot
→ commit
```

### Offboarding

```text
protected current admin actor
→ target User FOR UPDATE
→ set enabled=false + eligibility_version++
→ DELETE target ApplicationSessions
→ DELETE target GroupMemberships
→ DELETE target direct User RoleAssignments
→ required teardown/offboarding Audit events
→ commit
```

ProviderSubjectBinding remains.

### UserProfile replacement / erasure

```text
protected actor
→ current AuthZ
→ version CAS for replace OR DELETE profile for lawful erase
→ required Audit where T3 requires
→ commit
```

### User eligibility replacement

```text
protected actor
→ target User update lock/version
→ exact already-current = no-op
→ otherwise update eligibility/version
→ required Audit
→ commit
```

Re-enable never recreates sessions, grants or memberships.

### Area/Group replacement

```text
protected actor
→ owner row update lock/version
→ no-op or material version++
→ required Audit
→ commit
```

Area code is not mutable.

### GroupMembership add/remove

```text
protected actor
→ target User FOR SHARE on add
→ Authorization current access check
→ INSERT/DELETE membership
→ required Audit
→ commit
```

FK races with concurrent Group delete fail closed.

### RoleAssignment POST — Idempotency-Key

```text
BeginIn
→ protected actor
→ target User FOR SHARE when direct User assignment
→ application-routed Organization target facts
→ Authorization validates static role/scope
→ INSERT RoleAssignment
→ required Audit
→ CompleteIn
→ commit
```

Group target existence/deletion safety is also protected by FK semantics.

### Group delete

```text
protected actor
→ application gathers Authorization + ControlledDocs dependency facts
→ Organization validates dependencies resolved/absent
→ DELETE org.groups row
→ FK backstops any concurrent membership/grant/current-route/unactivated-step dependency race
→ required Audit
→ commit
```

### DocumentType POST — Idempotency-Key

```text
BeginIn
→ protected actor
→ AuthZ
→ INSERT DocumentType
→ required Audit
→ CompleteIn
→ commit
```

### DocumentType base/governance/template-set replacement

```text
protected actor
→ DocumentType update lock
→ expected resource-specific version check
→ for code/numbering change prove no committed Document yet
→ replace complete child set/config where applicable
→ material version++ or exact no-op
→ required Audit
→ commit
```

### Document create — Idempotency-Key

```text
external blank/template content seed prepared first
→ BeginIn
→ protected actor and deliberate responsible-owner target User when needed
→ protect/revalidate active Area
→ protect/revalidate active DocumentType
→ if Template seed, protect source Template Document against concurrent lifecycle/role change and revalidate exact EFFECTIVE source
→ lock/advance correct number counter
→ INSERT Document + REV000 DRAFT + WorkingContent + optional DocumentOrigin
→ consume AdmissionClaim if applicable
→ required Audit
→ CompleteIn
→ commit
```

### Next Revision — Idempotency-Key

```text
external copy from exact candidate EFFECTIVE source prepared first
→ BeginIn
→ protected actor
→ Document FOR UPDATE
→ revalidate exact source still current EFFECTIVE and no conflict/active obsolescence
→ allocate next ordinal = max+1 under Document serialization
→ INSERT next Revision DRAFT + WorkingContent
→ consume claim
→ required Audit
→ CompleteIn
→ commit
```

### DRAFT PATCH

```text
prepared READY upload optional
→ protected actor
→ Document shared relationship protection
→ exact current DRAFT + WorkingContent generation CAS
→ if source replaced, prove READY/claim/descriptor and consume claim
→ mutate title/source atomically
→ generation++ exactly once
→ commit
```

No semantic Audit for ordinary autosave per T3.

### SUBMIT — Idempotency-Key

```text
required malware preflight completed outside semantic tx
→ BeginIn
→ protected actor
→ DocumentType FOR SHARE coherent config snapshot
→ Document FOR UPDATE
→ prove exact current DRAFT + expected generation
→ prove exact handle READY/create-once/claim/descriptor/malware
→ INSERT immutable Submission
→ freeze route/steps when UseGovernanceRoute
→ create live GROUP dependency rows for unactivated GROUP steps
→ Revision DRAFT→SUBMITTED + current_submission_id
→ if NoHumanApproval + SourceOnly, system Release in same tx
→ if OfficialRendition required, River intent InsertTx in same Scope
→ required Audit
→ CompleteIn
→ consume applicable AdmissionClaim
→ commit
```

### Feedback POST — Idempotency-Key

```text
BeginIn
→ protected actor
→ exact live case authorization
→ INSERT Feedback
→ CompleteIn
→ commit
```

No duplicate Audit text copy.

### Governance decision

```text
protected actor
→ Document FOR UPDATE
→ reload exact active Attempt/Step/candidate snapshot
→ current AuthZ + SoD
→ INSERT immutable Decision
→ Step DECIDED
→ ACCEPT with more steps: activate next Step, freezing candidates and deleting GROUP live dependency if applicable
→ final ACCEPT: attempt COMPLETED; Release in same tx when representation gate satisfied
→ RETURN: attempt RETURNED; Revision SUBMITTED→DRAFT/current_submission_id=NULL
→ required Audit
→ commit
```

Governance Step decision surface is naturally idempotent and uses no Idempotency-Key POST machinery.

### Submission withdrawal

```text
protected actor
→ Document FOR UPDATE
→ exact current pre-Release Submission eligibility
→ INSERT Withdrawal
→ attempt WITHDRAWN where present; remove unactivated GROUP dependency rows
→ Revision SUBMITTED→DRAFT/current_submission_id=NULL
→ required Audit
→ commit
```

### Revision cancellation

```text
protected actor
→ Document FOR UPDATE
→ exact current DRAFT/pre-Release SUBMITTED eligibility
→ INSERT RevisionCancellation
→ attempt CANCELLED where live; remove live GROUP dependencies
→ Revision→CANCELLED/current_submission_id=NULL
→ required Audit
→ commit
```

### Responsible-owner replacement

```text
protected actor + target User FOR SHARE
→ Document FOR UPDATE
→ expected responsible_owner_version
→ validate target current enabled/same Company
→ exact no-op or change + version++
→ required Audit
→ commit
```

### Template-role replacement

```text
protected actor
→ Document FOR UPDATE
→ expected template_role_version
→ exact no-op or change + version++
→ maintain any owner-intrinsic eligibility invariants
→ required Audit
→ commit
```

### Obsolescence POST — Idempotency-Key

```text
BeginIn
→ protected actor
→ DocumentType FOR SHARE coherent governance snapshot
→ Document FOR UPDATE
→ prove exact current EFFECTIVE target / no open replacement / no active obsolescence
→ INSERT ObsolescenceRequest
→ NoHumanApproval: complete + target EFFECTIVE→OBSOLETE in same tx
→ human route: create GovernanceAttempt/steps/GROUP dependencies
→ required Audit
→ CompleteIn
→ commit
```

### Obsolescence withdrawal

```text
protected actor
→ Document FOR UPDATE
→ exact active human request/initiator-or-manager predicate
→ request ACTIVE→WITHDRAWN
→ attempt WITHDRAWN + remove unresolved GROUP dependencies
→ target remains EFFECTIVE
→ required Audit
→ commit
```

### OfficialRendition finalization

```text
render/provider work outside semantic tx
→ admitted READY candidate prepared
→ Document FOR UPDATE
→ reload exact Submission/current eligibility/representation requirement
→ prove READY/descriptor/claim
→ INSERT immutable OfficialRendition
→ required Audit
→ if human gate satisfied, system Release in same tx
→ consume claim
→ commit
```

Dead candidate = semantic no-op/fail-closed as upstream requires; produced content remains reclaimable.

### GC phase 1/phase 2

As §20; application/maintenance coordinates platform mechanism state with owner canonical semantic-reference proofs. Platform never queries ControlledDocs tables itself.

---

## 28. Query/read realization

### T8D-D38 — explicit owner-private `database/sql` SQL

Default persistence implementation is explicit owner-private SQL over T8-C `txscope.Scope` / ordinary DB executor paths.

No ORM/generic repository abstraction is introduced merely for consistency.

A query generator may be considered later only if it has a named current proof-backed benefit and preserves owner-local SQL visibility; it is not required by T8-D.

### Authorization canonical query

Authorization alone queries `authz.role_assignments`. It receives Organization's User/GroupMembership facts through T8-C, expands Role→Permission in static Go authority and computes scope/grant truth.

No Authorization SQL joins `org.*` to bypass Organization.

### AuthorizedScopes

Derived from the same RoleAssignment query/evaluator. It remains a prefilter and never evaluates ControlledDocs domain predicates.

### Controlled Documents reads

ControlledDocs owner-private SQL realizes:

```text
Library candidate query
Document Official
Document Work
Document History
AuthoringWork
GovernanceWork/Case
AccessFacts single/batch
GroupGovernanceUsage
RenditionWorkCandidate
semantic reference proofs for GC
```

Cross-owner display/reference enrichment remains application composition; ControlledDocs SQL does not join `org.*`.

### Search baseline

No Search table/materialized view exists.

Canonical Search/Library query reads current canonical ControlledDocs facts, including Revision state `EFFECTIVE`, code/title/type/Area/responsible-owner identifiers and T6 deterministic filters/ranking. Organization display names are composed outside ControlledDocs SQL.

### Audit read

Audit queries only `audit.events` and applies historical Company/Area visibility predicate before cursor pagination. Application must not post-filter a page.

### Idempotency read path

Platform idempotency queries only `platform.idempotency_keys/replays`; it never reprojects current semantic state to reconstruct historical replay.

### T8D-D39 — reject generic ORM/repository framework

A generic ORM/repository layer would obscure owner-specific lock/CAS/constraint semantics without a current named consumer benefit. Reject as baseline.

### T8D-D40 — normal views allowed, materialized Search rejected

Owner-private non-materialized views are allowed only when they reduce repeated query expression without creating shared cross-owner authority. Search materialization remains OFF.

---

## 29. PostgreSQL failure semantics

Target error mapping must preserve fail-closed business semantics. Exact wire errors remain T8-E.

Important PostgreSQL classes:

```text
23505 unique_violation
  expected only where natural conflict/duplicate semantics apply;
  idempotency expected same-key races use conflict-safe acquisition rather than poisoning Scope

23503 foreign_key_violation
  concurrent/dependency integrity failure -> fail closed

23514 check_violation
  invariant/programming-invalid state -> no blind retry

40P01 deadlock_detected
  whole local transaction is retryable in principle; no partial commit

40001 serialization_failure
  unexpected as routine READ COMMITTED path;
  retryable in principle, recurring evidence is architecture signal

57014 query_canceled/deadline
  rollback/fail loud
```

Global SERIALIZABLE, table-wide semantic locks and broad advisory-lock frameworks are not baseline.

---

## 30. T8-A selective-reuse disposition

Current implementation is evidence only. Reuse is property-level unless the current exact physical artifact passes all five T8-A proofs.

### Authentication current families

```text
auth_identities      DELETE / REWRITE
  password/local-auth/lockout/profile coupling conflicts with OIDC-only T6

auth_sessions        REWRITE
  preserve only durable current-session+expiry property
```

### Organization/AuthZ current families

```text
iam_users/groups/memberships/etc.   REHOME / REWRITE around Organization
role_capabilities                    DELETE; static Role→Permission authority
legacy tenant/RLS/GUC substrate      DELETE from Launch target
```

### Controlled Documents current families

```text
current controlled_documents shape  REWRITE
legacy technical document_revisions DELETE / REWRITE; not Business Revision authority
approval_*                            REWRITE into bounded governance relations
taxonomy families                     DELETE/fold into Area + DocumentType as ratified
template-version families             DELETE/fold into governed Document Template role
legacy current-status/pointers        DELETE unless explicitly selected above
```

### Audit

```text
current table shape       REWRITE / REHOME
audit append-only grants  PRESERVE PROPERTY
audit 64KiB bounded facts REFINE/PRESERVE PROPERTY
global hash-chain fields  DELETE
```

### Idempotency

```text
current table shape/status/raw response  REWRITE
scoped unique-key concurrency property   PRESERVE/REFINE
```

### Async/mechanism

```text
River mechanism                  PRESERVE / REHOME to configured river.*
legacy custom outboxes           DELETE
notifications persistence        DELETE
```

### Verification/deployment properties

```text
runtime DB identity != DDL owner                  PRESERVE
complete bidirectional table-ownership property   PRESERVE / REWRITE target catalog
```

Concrete migration/cutover/deletion order remains T10.

---

## 31. Alternatives challenged

### Namespace

Rejected:

```text
single catch-all schema only
```

It works technically but weakens owner visibility/verification while giving no transaction benefit over owner namespaces.

Rejected:

```text
database/service per owner
```

Contradicts T2 local one-transaction business transitions.

Selected owner semantic schemas + technical schemas preserve one DB/transaction while making ownership explicit.

### Role/Permission persistence

Rejected reference tables because product vocabulary is static authority and no admin customization exists.

### VersionToken

Rejected `xmin`, `ctid`, update timestamps and representation hashes as mutable-resource authority. Select explicit `BIGINT`.

### Document effectivity

Explored release-chain-only derivation and rejected it because T2 already defines explicit Revision lifecycle states. Selected Revision.state + immutable Release fact + partial unique EFFECTIVE barrier; rejected a duplicated `Document.current_status/current_revision_id` pointer.

### Governance

Rejected JSON-only route/attempt snapshot because structural one-active-step/candidate/dependency proof becomes weaker. Rejected generic workflow engine as speculative false owner. Selected closed relational MetalDocs tables.

### Idempotency

Rejected committed `IN_PROGRESS/FAILED` state and raw stored HTTP body/status. Rejected a constraint-trigger solution as more hidden than the relational paired Key↔Replay FK invariant. Selected uncommitted key claim + completed replay with deferred completion constraint.

### Immutability

Rejected trigger-heavy event-sourcing/lifecycle guards. Selected append-only tables + DML grants + owner transition SQL.

### ORM/repository framework

Rejected absent a current consumer proving lower total complexity while preserving exact SQL/lock ownership.

### Search

Rejected materialized Search and external search engine; T5/T6 current baseline remains canonical PostgreSQL query only.

---

## 32. Structural Inversion

Question:

> If MetalDocs were built new today with Product/T1→T8-C already ratified, would we choose the current physical database shape?

Answer:

```text
NO
```

We would not independently choose:

```text
password auth
tenant/RLS substrate
Approval semantic package/table family
taxonomy platform
TemplateVersion platform
permission expansion tables
generic outbox/notifications
Audit global hash chain
technical autosave revision table as business history
materialized current status
```

Preserving those would be sunk-cost architecture.

But greenfield purity is also rejected where current evidence proves a target property:

```text
PostgreSQL
River
runtime/DDL role separation
Audit append-only grants
complete ownership-catalog parity
conflict-safe idempotency unique-key concurrency
```

The candidate preserves/refines those properties while discarding legacy authority shape.

---

## 33. Proof obligations carried forward

A later accepted implementation/validation program must make at least these falsifiable:

### Namespace/ownership

```text
every first-party DB object classified both directions
all first-party SQL schema-qualified
foreign semantic SQL rejected
raw first-party River SQL rejected
runtime role has no DDL
```

### Organization/AuthN/AuthZ

```text
unknown/disabled User cannot issue Session
binding replacement invalidates sessions atomically
profile erasure leaves stable User
re-enable does not restore sessions/grants/memberships
membership/RoleAssignment races with offboarding fail closed
Group delete fails on every required live dependency
AuthorizedScopes never substitutes for exact Decide
```

### Document lifecycle

```text
code/ordinal uniqueness and non-reuse
no two open DRAFT/SUBMITTED Revisions
no two EFFECTIVE Revisions
DRAFT stale generation => zero mutation
SUBMIT cannot freeze stale generation
Submission immutable after commit
RETURN/withdraw/cancel never rewrite old Submission/Decision/Feedback
one active governance Step
GROUP activation snapshot immutable and empty stays empty
activated Group history does not keep Group alive
Release binds exact winning Submission
replacement Release state transition atomic
obsolescence cannot complete against stale/noncurrent target
```

### Exact content

```text
semantic descriptor stored on owning record
provider handle never proves exactness alone
create-once/no-overwrite provider conformance
untrusted immutable admission requires exact matching CLEAN digest
AdmissionClaim rollback safety
phase-1 + immediate phase-2 GC semantic proof
```

### Audit

```text
required Audit append failure rolls back governed/security mutation
Audit rows cannot be UPDATE/DELETE/TRUNCATE by runtime
historical visibility attribution is immutable and filter-before-pagination
```

### Idempotency

```text
same-key same-fingerprint loser does not poison caller Scope
winner rollback allows contender ownership
same key/different fingerprint => zero business mutation
commit without ReplaySnapshot is structurally rejected
ReplaySnapshot exact reconstruction is self-contained/PII-free
replay current disclosure authorization is rechecked without re-executing history
```

### River

```text
required OfficialRendition Submission commit <=> River intent commit
job cannot be worked before enclosing tx commit
tx rollback removes job intent
duplicate worker execution cannot duplicate semantic OfficialRendition/Release
```

---

## 34. Non-decisions / stage boundary

T8-D does **not** freeze:

```text
exact OpenAPI schemas/field maximums/headers/status mappings  T8-E
frontend package/query/cache realization                     T8-F
runtime process count/worker schedule/deploy topology        T8-G
Golden Flow proof matrix                                     T9
current→target schema migration/cutover/deletion order       T10
implementation tranche/file decomposition                    T11
```

T8-D also does not activate:

```text
pooled tenancy
RLS
materialized Search
body/OCR/vector search
Records retention/disposition
notifications
fresh-auth/eSignature
generic quorum/reassignment
future Artifact owner
```

Performance-only indexes remain implementation tuning unless independent review proves one is load-bearing for correctness/operability. Correctness unique/partial indexes in this candidate are architecture.

Exact query text is T11-local so long as it satisfies the owner/query/lock/constraint contracts frozen here.

---

## 35. Reopen triggers

Reopen only the implicated decision on material evidence, e.g.:

```text
PostgreSQL 16-compatible primitives cannot sustain a ratified invariant
one-database transaction model cannot sustainably realize T2
cross-owner identity FK creates unavoidable semantic coupling
explicit SQL becomes less provable/sustainable than a named query-generation alternative
BIGINT CAS cannot represent a required owner precondition safely
closed governance tables cannot represent a promoted Launch requirement
paired idempotency completion constraint proves operationally unsustainable
River custom schema/transaction integration fails pinned-version proof
a real Search workload activates T5 materialization seam
pooled tenancy becomes Product authority
```

Preference, legacy familiarity, ORM fashion and hypothetical scale are not reopen triggers.

---

## 36. Decision ledger

```text
T8D-D01  SELECT one PostgreSQL product-state database.
T8D-D02  SELECT authn/org/authz/controlled_docs/audit/platform/river schemas.
T8D-D03  SELECT fully-qualified first-party SQL; search_path never selects authority.
T8D-D04  SELECT PostgreSQL-16-compatible persistence feature floor.
T8D-D05  SELECT complete bidirectional DB-object ownership catalog; unknown=FAIL.

T8D-D06  SELECT static Role/Permission/bundles; persist RoleAssignment only.
T8D-D07  REJECT RLS/tenant substrate for Launch target.
T8D-D08  REJECT persisted permission/status/Search/current-view duplicate truth.

T8D-D09  SELECT explicit monotonic BIGINT VersionToken.
T8D-D10  SELECT WorkingContent BIGINT generation OCC.
T8D-D11  SELECT Revision.state as canonical lifecycle state.
T8D-D12  SELECT immutable Release + partial unique EFFECTIVE structural barrier.
T8D-D13  SELECT at most one open DRAFT/SUBMITTED Revision per Document.
T8D-D14  SELECT current_submission_id only as bounded current SUBMITTED relation.

T8D-D15  SELECT closed relational governance model.
T8D-D16  SELECT structural one-ACTIVE-Step uniqueness.
T8D-D17  SELECT separate live GROUP dependency and immutable activated candidates.
T8D-D18  REJECT generic Workflow/polymorphic subject persistence.

T8D-D19  SELECT semantic exact descriptors on owning semantic rows.
T8D-D20  SELECT ManagedContent technical state only.
T8D-D21  SELECT row-existence AdmissionClaim lifecycle.
T8D-D22  SELECT two-phase GC with repeated full semantic/claim/pin proof.

T8D-D23  SELECT paired idempotency Key + Replay tables.
T8D-D24  SELECT deferred reverse FK making incomplete replay uncommittable.
T8D-D25  REJECT durable target IN_PROGRESS/FAILED replay state.

T8D-D26  SELECT River third-party persistence under river.*.
T8D-D27  REJECT first-party raw SQL against River tables.

T8D-D28  SELECT cross-owner FKs only for stable identity/existence.
T8D-D29  REJECT cross-owner semantic cascades.
T8D-D30  SELECT serving runtime DB role separate from DDL/migration owner.
T8D-D31  SELECT DB grants/column grants as immutable-history enforcement.
T8D-D32  SELECT zero semantic lifecycle triggers baseline.

T8D-D33  SELECT protected actor User FOR SHARE for authenticated semantic mutations.
T8D-D34  SELECT target User update-strength lock for offboarding/eligibility root.
T8D-D35  SELECT stable Document FOR UPDATE as lifecycle serialization root.
T8D-D36  SELECT DocumentType shared/update locking for coherent configuration snapshots.
T8D-D37  SELECT deterministic same-class row-lock ordering.

T8D-D38  SELECT explicit owner-private database/sql SQL as persistence baseline.
T8D-D39  REJECT generic ORM/repository framework baseline.
T8D-D40  SELECT owner-private normal relational queries/views only; materialized Search OFF.
```

---

## 37. Independent-review attack surface

Independent Fable review must attack at least:

```text
1. Is owner-namespaced schema strategy actually stronger than one schema after accounting for migration/privilege complexity?
2. Does any cross-owner FK violate T8-B/T8-C communication ownership or create hidden lifecycle coupling?
3. Is the RoleAssignment one-table shape complete for T3 scope/subject semantics without a persistent role catalog?
4. Can ProviderSubjectBinding history/current uniqueness survive legitimate replacement without blocking required security recovery?
5. Are VersionToken resources independently versioned where T6 requires, without false conflicts or stale holes?
6. Does Revision.state + Release create duplicate effectivity truth, or is the split exactly T2 lifecycle-state vs immutable effectivity fact?
7. Are the partial unique open/EFFECTIVE constraints sufficient under READ COMMITTED replacement races?
8. Does `current_submission_id` add only bounded current SUBMITTED relation rather than hidden latest-pointer authority?
9. Does governance relational shape preserve current route snapshots, Group deletion law, empty GROUP snapshot and bounded recovery?
10. Are immutable-history grants sufficient where one-time terminal-state updates remain necessary?
11. Does ManagedContent/AdmissionClaim/GC locking eliminate attach-vs-delete races without platform semantic SQL?
12. Is storing malware proof on ManagedContent mechanism the smallest complete exact-byte realization?
13. Is paired Key↔Replay with a deferred reverse FK valid PostgreSQL 16 DDL and operationally safer than committed status/savepoint/constraint-trigger alternatives?
14. Can loser idempotency read completed replay under the ratified READ COMMITTED path without transaction poisoning?
15. Does River v0.37.1 custom schema + database/sql InsertTx integration hold for the exact pinned dependency?
16. Is one runtime DB role + column grants strong enough without exploding grants or requiring owner-specific DB identities?
17. Does the lock-order matrix cover create-from-template/config/offboarding/GC cases without hidden deadlock cycles?
18. Is zero semantic trigger baseline actually sustainable for every accepted invariant?
19. Does any current-schema reuse claim preserve legacy authority by accident?
20. Is any Launch/future capability persisted without a named current consumer?
```

A reviewer must not promote implementation preferences into Product requirements or require compatibility with current throwaway DEV/TEST business data.

---

## 38. Candidate verdict

Current Lead/Method verdict before independent challenge:

```text
T8-D GLOBAL MAXIMUM CANDIDATE = READY FOR INDEPENDENT FABLE REVIEW

selected class:
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+ DECLARATIVE CORRECTNESS
+ PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+ READ COMMITTED NARROW SERIALIZATION
+ EXPLICIT CAS
+ IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+ TRANSACTIONAL KEY↔REPLAY COMPLETION
+ THIRD-PARTY RIVER SCHEMA ISOLATION
+ SELECTIVE PROOF-BACKED PROPERTY REUSE
```

No T8-C/upstream reopen is currently proposed.

T8-E remains NOT OPEN.

Implementation remains BLOCKED.

### Exact next gate

```text
independent Fable review of this exact candidate
→ Lead confrontation/adjudication
→ bounded correction/re-review only if materially required
→ explicit operator ratification
→ only then promote durable T8-D authority and open T8-E
```

---

**End of non-authoritative T8-D Global Maximum candidate.**
