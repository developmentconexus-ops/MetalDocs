# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions are binding; open items are explicit.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Canonical program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED. No product code, schema, OpenAPI, frontend or migration implementation is authorized yet.**

---

## 0. Fresh-session contract

Read in this order:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. this ledger
5. `wiki/references/current-agent-handoff.md`

Do not resume historical roadmaps, milestones, specs, deleted `docs/superpowers` artifacts or old implementation PRs by inertia.

Design sequence:

```text
product/domain semantics
→ invariants + lifecycle
→ organization/authz/workflow integration
→ build-vs-buy
→ bounded contexts / ownership
→ data model + DB constraints
→ event/transaction contracts
→ API + frontend journeys
→ migration/delete map
→ implementation specification
→ implementation plan
→ code
```

Goal: the smallest professional architecture that makes important invalid states unrepresentable and leaves explicit extension triggers instead of speculative engines.

---

# 1. Whole-platform reset

The redesign began with authorization drift, expanded into Approval, and then proved that `documents`, `controlleddocuments`, `templates`, taxonomy, IAM, Approval, rendering and release contained overlapping authorities.

Strongest counterexample: browser QA let a human review edited content while freeze rendered a blank template snapshot. The final PDF/hash therefore did not represent what the human reviewed.

**Target property:** every business fact has one authority; supporting concerns consume it instead of reinterpreting or mutating it independently. Current code/schema/API are migration evidence, not target-design authority.

---

# 2. Target responsibility map

| Current concern/module | Target disposition |
|---|---|
| `auth` | retain local V1 AuthN/session implementation behind stable identity/assurance boundary |
| `iam` | split conceptually into **Organization** + **Authorization**; tenant lifecycle leaves this catch-all |
| `approval` | small specialized Approval V1; never owns release/effectivity or periodic review |
| `documents` | becomes core of **Controlled Information** |
| `controlleddocuments` | retire as target context; identity/numbering move to Document/configuration |
| `templates` | retire parallel lifecycle; template becomes role of governed Document/Revision |
| `taxonomy` | dismantle: Area → Organization; Profile → DocumentType; Family → category; GovernanceClass deleted |
| `render` | supporting rendition infrastructure bound to exact RevisionSubmission |
| `distribution` | real obligation/acknowledgement domain over released Revisions |
| `tokens` | conceptually split into product System Value Catalog + tenant Dictionary |
| `audit` | append-only/tamper-evident cross-domain audit trail; never business-state authority |
| `notifications` | event consumer/inbox/delivery projection only |
| `search` | rebuildable projection/read model |
| `security` | split conceptually: sessions/lockouts → AuthN; tenant crypto → Platform Security; heuristic signals deferred/optional |
| `jobs` | orchestration infrastructure, not business bounded context |
| tenant lifecycle | platform/tenant administration concern with explicit PlatformOperator/System boundary |

---

# 3. LOCKED — Authentication + Organization + Authorization north star

## AUTHN-01 — Separate authorities

- Authentication: who is this actor/session?
- Authorization: what may this principal do in this tenant/scope?
- Approval: who participates in this concrete submission?
- Domain Governance: is this action legal now given lifecycle, SoD, reauth and immutable-content rules?

None substitutes for another.

## AUTHN-02 — No Keycloak now

Current MetalDocs AuthN is sufficient for V1. Preserve a stable authenticated-principal / assurance seam for future OIDC/SAML/enterprise IdP/MFA/passkeys when a real requirement appears.

## ORG-01 — Organization

```text
Tenant
Area
User
Group
GroupMembership
```

Area is organizational truth reused by Document ownership, RoleAssignment scope and Approval actor resolution.

## ORG-02 — Flat Groups

- User may belong to multiple Groups.
- No nested Groups V1.
- Group receives RoleAssignments, never raw permissions.
- Group is not structurally forced into one Area.

## AUTHZ-01 — Five tenant roles

```text
tenant_owner
area_manager
author
approver
viewer
```

Roles are bundles only. Runtime checks semantic Permissions, never role branches.

## AUTHZ-02 — One grant shape

```text
RoleAssignment
  subject: User | Group
  role
  scope: Tenant | Area
  granted_by / granted_at
  revoked_by? / revoked_at? / reason?
```

Additive grants + default deny. No deny engine, nested groups or temporal grant scheduler V1.

## AUTHZ-03 — Role meaning

- `viewer`: effective/released official content in scope.
- `author`: eligible working content + create/edit/comment/submit.
- `approver`: qualification only; no blanket draft visibility.
- `area_manager`: operational manager in Area; not RBAC administrator.
- `tenant_owner`: tenant product administrator through normal Permission grants; never bypasses Domain Governance.

## AUTHZ-04 — Decision composition

```text
current permission/qualification
+ workflow/domain relationship where applicable
+ Domain Governance constraints
= ALLOW
```

No tenant-owner bypass. No OpenFGA/SpiceDB V1; revisit only for material arbitrary resource-sharing graphs.

Final Permission Catalog and role bundles are intentionally deferred to **R9**, now that the product operations are known.

---

# 4. LOCKED — Approval V1

## APPR-01 — Specialized document approval, not BPM

No BPMN, generic branches/gateways/service tasks, Camunda/Flowable, CEL, generic delegation engine or M-of-N V1.

## APPR-02 — Versioned sequential ApprovalPolicy

```text
ApprovalPolicy
  id
  version
  ordered ApprovalStep[]
```

New policy version applies to new submissions; historical/in-flight instances stay pinned.

## APPR-03 — Step = human task

```text
order
name
purpose: review | approval
actor_rule
completion: ANY | ALL
requires_reauthentication
due_in_days?
```

Participant rules V1:

```text
NamedUser
Group
RoleInArea(role, fixed-area | subject-area)
```

Participants resolve when a Step activates and are snapshotted as evidence. Current qualification is rechecked when acting.

## APPR-04 — Human outcomes

```text
accept
return_for_changes
```

Separate operations:

```text
withdraw
cancel
reassign
```

No normal terminal `reject`.

## APPR-05 — Attempts are content-exact

ApprovalInstance binds one immutable `RevisionSubmission`.

`return_for_changes` terminates that attempt and returns the same DocumentRevision to DRAFT. Resubmission creates a new RevisionSubmission and, when approval is required, a new ApprovalInstance.

Audited reassignment covers unavailable actors. Optional per-Step reauthentication survives. Approval evidence pins actor, policy/version, Step, Submission/digest, outcome, reason/comment and reauth evidence when required.

---

# 5. LOCKED — Controlled Information configuration (R3)

## CI-01 — One core

The target split `documents + controlleddocuments + templates` is rejected.

```text
Document
DocumentRevision
```

are the core business concepts.

## CI-02 — DocumentType replaces DocumentProfile

```text
DocumentType
  id
  tenant_id
  code            // immutable
  name
  description?
  category_id?
  status: ACTIVE | INACTIVE
```

No own versioning V1. Inactive prevents new use but does not invalidate existing Documents. Document type is immutable after creation.

## CI-03 — DocumentTypeCategory

Optional classification/navigation only. No inherited Approval, numbering, metadata or permissions.

## CI-04 — GovernanceClass deleted

`controlado/simples/livre` is not a cross-domain authority. Each concern owns explicit configuration.

Approval configuration is explicit:

```text
NoHumanApproval
or
UsePolicy(ApprovalPolicyID)
```

No fake zero-stage route.

## CI-05 — Template is a role of governed Document

Template has no parallel lifecycle/version counter. Template changes use normal DocumentRevisions and official `REVxxx` labels.

## CI-06 — TemplateUse M:N

```text
TemplateUse
  template_document_id
  target_document_type_id
  is_default
```

At most one UX default per DocumentType. Blank creation stays valid V1. Creation resolves the template Document's current EFFECTIVE Revision once and pins exact source Document + Revision + digest permanently.

---

# 6. LOCKED — Document + Revision + Submission lifecycle (R4)

## REV-01 — Document is stable identity

Document owns tenant, business code, type, Area, origin/provenance, responsible operational owner where applicable, and pointers to effective/open revisions.

## REV-02 — Official revision labels

```text
REV001
REV002
REV003
...
```

`REVxxx` is the human/audit/business revision. Technical IDs, row versions and policy versions are separate namespaces.

## REV-03 — Revision = business change cycle

Autosaves/checkpoints/edit snapshots are technical history inside the open Revision and never consume REV numbers.

At most one open Revision per Document V1; one effective Revision may coexist with that open Revision.

## REV-04 — Revision states

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

No `APPROVED`, `SCHEDULED` or `PUBLISHED` Revision state. Those are Approval/Release facts projected in UI.

## REV-05 — REV allocation

Allocate next REV when change cycle starts. Never reuse a REV label even if cancelled. `REV002+` requires `reason_for_change` before submission.

## REV-06 — RevisionSubmission is immutable first-class evidence

```text
REV002
  ├── Submission #1 → digest A → returned
  └── Submission #2 → digest B → released
```

Exists even for `NoHumanApproval`. Approval, Rendition and Release bind Submission, not mutable Document/Revision state.

## REV-07 — Editing boundary

- DRAFT editable.
- SUBMITTED frozen.
- return/allowed withdraw closes old Submission and returns same REV to DRAFT.
- after completed Approval V1 does not reopen candidate for editing; cancel it and create a new REV if new content is needed.

## REV-08 — Effectivity/retirement

Release of REV002 atomically makes REV002 EFFECTIVE and prior REV001 SUPERSEDED.

`SUPERSEDED` = newer Revision of same Document replaced it.

`OBSOLETE` = Document explicitly retired without successor; terminal V1.

Cross-Document replacement is separate and deferred until a real requirement appears.

---

# 7. LOCKED — Numbering + TemplateSpec + metadata (R5)

## NUM-01 — Code identity

`Document.code` is immutable and unique tenant-wide. Document type, Area and code are immutable V1. `DocumentType.code` and `Area.code` are stable identifiers; display names may change.

## NUM-02 — Allocation

Code allocation + Document creation + initial REV001 creation are one atomic operation.

Successful creation permanently consumes the sequence number. Rollback before creation may leave it reusable. Preview is advisory only; Create response is authoritative.

## NUM-03 — Small numbering language

DocumentType owns numbering configuration.

Allowed V1 tokens:

```text
{TYPE}
{AREA}
{SEQ}
```

Sequence scopes: `TYPE | TYPE_AREA`. `sequence_width` is minimum zero-padding. No year/month/custom fields/formulas/scripts/resets V1.

Normal Create has no manual-code override. Legacy code preservation belongs to explicit import/migration/bootstrap authority.

## TPL-01 — TemplateSpec only owns authoring contract

Delete template `MetadataSchema` as authority for numbering/retention/distribution/generic metadata.

A template Revision contains ordinary governed source content (e.g. DOCX) + `TemplateSpec`.

## TPL-02 — TemplateField

```text
TemplateField
  key
  label
  value_type: text | date | number | choice | user | image
  source:
      user_input
    | system(key)
    | dictionary(key)
  typed constraints...
  visible_if?
```

`computed`/`dictionary` are sources, not types. Typed constraints survive; invalid cross-type combinations fail closed. `visible_if` uses a closed operator set such as `eq/ne/gt/gte/lt/lte`, not a generic expression language.

## TPL-03 — Parity and provenance

Before template submission, source content anchors/tokens and TemplateSpec must agree.

Creating from template copies/pins the exact effective source Revision content + applicable TemplateSpec into derived REV001 and records immutable `DocumentOrigin`:

```text
Blank
or
Template {
  source_template_document_id
  source_template_revision_id
  source_template_revision_label
  source_digest
}
```

Derived Documents never silently rebind to newer template Revisions.

## TPL-04 — RevisionContent + submission digest

```text
RevisionContent
  source/content artifact ref + hash
  authoring schema snapshot
  field values
  governed Revision metadata
```

`RevisionSubmission.submission_digest` covers the complete governed content identity needed to attest what was submitted.

`title` belongs to DocumentRevision. Stable identity/operational facts belong to Document. No generic tenant custom-metadata engine V1.

---

# 8. LOCKED — Periodic Review + Renditions + Release/Effectivity (R6)

## REVIEW-01 — Periodic Review belongs to Controlled Information

```text
PeriodicReviewPolicy =
    Disabled
  | Every(n months)
```

Cadence starts at actual Effectivity and restarts after completed review. Due/overdue does not invalidate EFFECTIVE content. Expiration is separate future policy.

## REVIEW-02 — Append-only review evidence

```text
PeriodicReviewRecord
  document_id
  reviewed_revision_id
  reviewed_by
  reviewed_at
  outcome: confirmed_current | change_required
  comment?
  previous_due_on
  next_due_on
  policy_snapshot
```

`confirmed_current` keeps same REV EFFECTIVE. `change_required` records finding but does not auto-create a new REV. Completion fails closed if the reviewed REV is no longer effective.

A Document subject to periodic review requires responsible owner relationship + relevant authorization. Owner grants responsibility, not access.

## REND-01 — Rendition

A Rendition is an immutable derived representation of one exact Submission with output hash and source provenance.

```text
Rendition
  submission_id
  kind
  storage_ref
  content_hash
  media_type
  size_bytes
  generated_at
  generator_component
  generator_version_or_build_digest
```

Only `OFFICIAL_PDF` is mandatory for Release V1. `FINAL_DOCX` may exist but does not block Effectivity.

Approval preview is not official Rendition but must derive from the same Submission/digest.

## REND-02 — Approval approves Submission, not PDF bytes

Official PDF is an attested derivation of the approved Submission and may manifest Approval/signature evidence. Persist official bytes + hash + source digest + renderer/build identity; no promise of future byte-identical rerender.

Render failure is operational state/projection, never Revision lifecycle state.

## RELEASE-01 — Submission is release candidate identity

`ReleaseGeneration` is no longer a required domain noun. Release Coordinator evaluates `submission_id`.

## RELEASE-02 — Automatic centralized release

No human publish button. Approval completion, rendition readiness, timer and reconciliation all invoke idempotent `EvaluateRelease(submission_id)`.

`ReleasePlan.not_before?` is a gate, not a Revision state. Actual fact is `effective_at = released_at` in the winning transaction; no silent retroactive effectivity.

## RELEASE-03 — Winning transaction + evidence

Atomically:

```text
candidate REV -> EFFECTIVE
prior effective REV -> SUPERSEDED
Document.effective_revision_id -> candidate
Document.open_revision_id -> NULL
insert immutable ReleaseRecord
emit/enqueue lifecycle events
```

Candidate may be CANCELLED after Approval but before Release; historical Approval/Rendition evidence remains.

Legacy `effective_date` is not a pre-release TemplateField V1 because actual Effectivity is born after the mandatory pre-release PDF.

---

# 9. LOCKED — Distribution + Values + Audit + Notifications + Search (R7)

## DIST-01 — Distribution = obligation/acknowledgement, not AuthZ or Training

Distribution answers who **must** take explicit notice of an EFFECTIVE Revision. It does not grant RBAC access and does not become an LMS/training engine.

## DIST-02 — Configuration on Document

```text
DistributionConfiguration =
    None
  | ReadAcknowledgement {
      targets: User | Group
      due_in_days?
      requires_reauthentication
    }
```

No Area target V1 without a real UserAreaMembership concept. Changing configuration affects future releases by default; current-REV assignment is explicit/audited.

## DIST-03 — Release snapshots concrete users

Release resolves User targets and Group memberships into concrete per-user `DistributionAssignment`s. Later Group membership changes never rewrite history. Post-release onboarding uses explicit assignment against the current effective REV.

One obligation per user/release; multiple source targets are provenance.

## DIST-04 — Assignment + acknowledgement

```text
pending
acknowledged
cancelled
superseded
```

Overdue is derived. A newer effective REV supersedes pending old-REV obligations and creates a fresh cohort.

Opening a notification, viewing or downloading does **not** complete the obligation. Explicit immutable `AcknowledgementRecord` does; optional fresh reauth reuses AuthN assurance.

Distribution never edits RoleAssignments. Task-specific visibility is finalized in R9.

## VALUE-01 — Product System Value Catalog vs tenant Dictionary

System keys V1:

```text
document_code
revision_label
revision_title
document_type_code
document_area_code
document_area_name
revision_created_by_name
```

Legacy `approval_date`/approvers move to Approval manifestation in official PDF; actual `effective_date` stays outside TemplateSpec V1.

## VALUE-02 — Mutable external values snapshot at new REV creation

Dictionary/external mutable values referenced by authoring schema resolve when a new DocumentRevision is created and are copied into RevisionContent/provenance. Same-REV return/resubmit does not silently re-resolve. Historical released content never depends on live Dictionary state.

## AUDIT-01 — Domain evidence vs Audit Trail

Business authorities stay their own records (`RevisionSubmission`, `ApprovalDecision`, `PeriodicReviewRecord`, `ReleaseRecord`, `DistributionAssignment`, `AcknowledgementRecord`, RoleAssignment history, etc.). `AuditEvent` is transversal compliance/investigation evidence and never substitutes for these authorities.

## AUDIT-02 — Durable audit intent

Critical governed mutation must not report success unless an audit intent/event is durably persisted in the same commit boundary (direct append or transactional outbox; exact mechanism R10). Usage telemetry such as view/download/search/notification-read may be asynchronous.

## AUDIT-03 — Integrity/export/system actors

Preserve Audit Trail as append-only, tamper-evident and exportable. User and explicit System actors are distinct; automatic release is attributed to System, not the last approver.

## NOTIF-01 — Notifications are projection/delivery only

Notification `READ` means the notification was read, never the document or an acknowledgement. “Minhas Pendências” queries business authorities, not unread notifications. Reminders derive from business due facts.

## SEARCH-01 — Search is rebuildable projection

Official Library indexes current EFFECTIVE Revision. Working Search may index open REV under current working-content authorization. Historical superseded/obsolete Revisions stay out of global search by default. Search is eventually consistent and never grants access; canonical endpoint rechecks current AuthZ.

No Elasticsearch/OpenSearch requirement yet.

---

# 10. LOCKED — Tenant lifecycle + Platform Security (R8)

## TENANT-01 — Platform authority is outside tenant RBAC

`PlatformOperator` and `SystemPrincipal` are platform identities, not a sixth/seventh tenant Role and not RoleAssignment subjects.

PlatformOperator may create/suspend/resume tenants and inspect lifecycle operations, but has **no implicit right to read/edit/approve tenant business content**. Support/break-glass access is outside V1 and would require explicit, time-bounded, audited design if ever added.

## TENANT-02 — Tenant lifecycle

```text
ACTIVE
SUSPENDED
ERASED
```

- `ACTIVE`: normal operation.
- `SUSPENDED`: reversible service stop; data remains intact.
- `ERASED`: terminal tombstone state; tenant is not usable/reactivatable V1.

Deletion request is a separate object/process, not another Tenant state:

```text
TenantDeletionRequest
  status: PENDING | CANCELLED | EXECUTED
  requested_by
  requested_at
  execute_after
```

Grace period is simple product/deployment policy. Tenant remains ACTIVE while request is pending and may cancel before execution.

## TENANT-03 — Tenant vs platform operations

- PlatformOperator: create tenant, suspend, resume.
- tenant_owner: manage own tenant, request own export, request/cancel own deletion.
- SystemPrincipal: execute async lifecycle jobs, including terminal erasure after grace.

Tenant owner never manages other tenants. PlatformOperator does not become tenant data owner.

## TENANT-04 — Onboarding

Onboarding creates:

```text
Tenant
+ initial User
+ tenant_owner @ Tenant
+ single-use time-limited activation credential
```

Do not create historical `system_admin` in the new tenant and do not ask a PlatformOperator to choose/know the tenant owner's password.

The same activation primitive may later serve ordinary new-user activation; no invitation platform is required now.

## TENANT-05 — Suspension and deactivation

Tenant suspension:

- revokes tenant sessions;
- denies new login and normal business mutations;
- preserves data;
- normal business jobs respect suspension;
- lifecycle/security jobs may continue where required.

User deactivation is independent from tenant suspension. It revokes that user's sessions and denies login while preserving immutable identity/history. Pending Approval/owner/responsibility relationships require explicit reassignment/attention; they never disappear silently.

## TENANT-06 — Export and deletion request require fresh auth

`tenant.export` and `tenant.deletion.request` are sensitive same-tenant operations for tenant_owner and require fresh authentication through the common AuthN assurance seam.

Platform operators do not receive implicit access to export artifacts.

## TENANT-07 — Erasure pipeline

Conceptually:

```text
DeletionRequest reaches execute_after
→ System suspends tenant
→ revoke sessions
→ erase live tenant-owned rows
→ delete tenant object-store blobs
→ destroy tenant DEK
→ preserve allowed non-PII audit/platform skeleton
→ Tenant -> ERASED
→ persist platform TenantErasureRecord
```

Each phase must be idempotent/reconcilable; exact job/transaction mechanics come in R10.

## TENANT-08 — Audit survives erasure as non-PII skeleton

Target rejects deletion of the Audit Trail itself.

Audit remains append-only/tamper-evident. Post-erasure retained skeleton may contain only opaque internal IDs and non-PII structural facts (event id, internal tenant/actor/resource IDs, action, timestamps, hash-chain fields, etc.). Sensitive payload/display data must be erasable/unreadable.

Tenant-scoped Audit payload is encrypted behind the tenant DEK when retention requires the skeleton to survive; destroying that DEK makes the payload unreadable while preserving structural evidence.

Do not claim cryptographic erasure for data that was never actually protected by the destroyed key. R10 migration must census legacy plaintext/PII surfaces honestly.

## TENANT-09 — Tenant erasure tombstone survives outside erased data

Platform-level `TenantErasureRecord` keeps only minimal non-PII facts such as opaque tenant id, request id and erasure timestamp/system execution fact.

Backup/restore procedure must reapply erasure tombstones before restored service can become available, preventing resurrection from an older backup.

## SECURITY-01 — Small tenant crypto primitive

V1 preserves the useful mechanism:

```text
Platform KEK
  ↓ wraps
Tenant DEK
```

No per-document key tree, tenant key-rotation engine, HSM workflow or key-escrow product without a concrete requirement. Exact KEK storage/KMS choice is R10 infrastructure.

## SECURITY-02 — AuthN assurance seam

Current local AuthN remains V1. Conceptually expose:

```text
AuthenticatedPrincipal
  user_id
  tenant_id
  authenticated_at
  auth_method
  assurance / fresh-auth evidence
```

Approval, acknowledgement, export and deletion request can demand fresh authentication without knowing whether the mechanism is password today or external MFA/IdP later.

## SECURITY-03 — Remove fake MFA from target V1

Current `mfa_enabled` / `mfa_enrolled_at` coverage is a stub without actual TOTP/WebAuthn enrollment. Stub fields/cards are **not** a V1 security control and have no entitlement to survive.

Real MFA/passkeys/SSO/SAML/per-tenant federation are formal triggers to re-evaluate Keycloak/external IdP **before** building more identity-provider capabilities inside MetalDocs.

Future federation maps `(issuer, subject) -> internal User`; MetalDocs remains authority for Organization, Roles, Permissions and workflow relationships.

## SECURITY-04 — Split current security catch-all

- sessions / credential lifecycle / lockouts / fresh-auth → AuthN;
- tenant envelope crypto → Platform Security;
- Active Sessions + Lockouts survive as useful tenant admin views/actions;
- current heuristic `Security Signals` (off-hours, new-device, spikes, etc.) are optional/deferred observability projections, not architectural V1 requirements.

Session absolute TTL, idle timeout and fresh-auth window are product/deployment policies V1, not a per-tenant policy engine.

---

# 11. Build-vs-buy rulings to date

| Technology/class | V1 ruling | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | do not deploy now | enterprise SSO/federation, real MFA/passkeys, tenant-specific IdP or credential externalization |
| OpenFGA / SpiceDB | do not deploy now | arbitrary resource sharing / large relationship graph / service split |
| Camunda / Flowable / BPMN | do not use for document Approval V1 | product genuinely becomes generic business-process engine |
| Temporal as Approval/Release engine | do not use | durable orchestration requirement current outbox/jobs cannot economically serve |
| CEL / expression language | do not use now | real conditional product/workflow policies cannot be represented cleanly by typed configuration |
| Elasticsearch/OpenSearch | no requirement yet | measured search scale/query needs exceed economical PostgreSQL projection |
| dedicated LMS/training engine | no requirement | competency/assessment/training-plan requirements appear |

Libraries/frameworks are selected only after exact responsibility closure.

---

# 12. Explicit target deletions/replacements

No entitlement to survive:

- target split `documents` / `controlleddocuments` / `templates`;
- separate `ControlledDocument` identity;
- `DocumentProfile`;
- behavioral `DocumentFamily` hierarchy;
- `GovernanceClass`;
- parallel TemplateVersion lifecycle/version counter;
- template `MetadataSchema` as numbering/retention/distribution authority;
- duplicate template `DocCodePattern`;
- normal-create manual code override;
- `CompositionJSON` without independent requirement;
- user-visible `v7` document revisions;
- autosaves as official DocumentRevisions;
- Document lifecycle carrying revision/workflow states;
- `StageKind`, configurable stage capability, M-of-N, drift policies, generic delegation engine;
- normal terminal human reject;
- ApprovalInstance reuse across edited content;
- generic BPMN/CEL/phase/branching machinery;
- role-based AuthZ bypasses and multiple semantic grant sources;
- `editable_by_role`;
- Approval owning periodic review or release/effectivity;
- Approval mutating Controlled Information tables directly;
- `ReleaseGeneration` as required business identity;
- mandatory final-DOCX gate V1;
- render/freeze path choosing bytes other than exact Submission;
- `SCHEDULED` Revision state;
- automatic expiration merely because periodic review is overdue;
- live Group-membership derivation as historical distribution denominator;
- notification-read as document-read/ack evidence;
- live Dictionary references from historical Revisions;
- Audit Trail as business-state database;
- Search projection as authorization/business-state authority;
- `system_admin` as tenant/platform super-role;
- PlatformOperator implemented as tenant RoleAssignment;
- tenant admin chosen initial passwords;
- fake/stub MFA coverage as if MFA existed;
- tenant erasure deleting the target Audit Trail itself;
- current heuristic Security Signals as mandatory V1 foundation;
- old roadmap/milestone/spec documents as live authority.

Prefer deletion over compatibility shims when no deployed/contractual compatibility requirement proves a shim necessary.

---

# 13. Remaining design queue before implementation

## R9 — Final Authorization Matrix — **NEXT**

Now that product operations are known, freeze:

1. final semantic Permission Catalog;
2. five tenant-role bundles;
3. PlatformOperator/System permissions remain a separate authority namespace;
4. Organization administration: users, groups, memberships, roles, Areas;
5. DocumentType/category/policy/dictionary/template administration;
6. Document/Revision operations by scope/state;
7. Approval act/oversee/cancel/reassign/policy management;
8. Periodic Review ownership/override/reassignment;
9. Distribution configure/assign/acknowledge/oversee;
10. Audit/export/security/session operations;
11. tenant export/deletion request/cancel;
12. collection filtering and effective/working/case-specific visibility;
13. relationship checks (Approval participant, Distribution assignee, responsible owner, submitter);
14. Domain Constraint/SoD matrix;
15. RLS/DB-constraint/tripwire backstops;
16. positive/negative Golden Matrix covering every role and sensitive operation.

## R10+ — Technical architecture to code-ready contract

1. final build-vs-buy pass;
2. bounded contexts/packages and dependency DAG;
3. data model/table ownership/DB constraints;
4. transaction boundaries and outbox/event contracts;
5. APIs/OpenAPI/DTOs/problem semantics;
6. frontend IA and complete journeys;
7. migration/delete/rename map from current code;
8. compatibility/import policy;
9. test/proof matrix;
10. final ADR/spec set;
11. adversarial review + operator approval;
12. implementation specification;
13. implementation plan;
14. only then code.

---

# 14. Documentation authority

- `wiki/architecture/cohesive-platform-redesign.md` is the current-program entrypoint.
- this ledger is the only active detailed WIP decision source under `docs/superpowers`.
- `wiki/references/current-agent-handoff.md` is a short recovery pointer.
- Git history is the archive for deleted historical staging documents and prior ledger forms.
- legacy wiki/module/ADR material may explain runtime/history but cannot override operator-approved target decisions here.
- final integrated decisions are promoted to durable wiki/ADR/spec authorities only after design closure.

---

# 15. Implementation gate

Implementation starts only when all are true:

- [ ] whole-product domain map closed;
- [x] Organization/AuthZ north star closed;
- [x] Approval V1 closed;
- [x] R3 Controlled Information configuration closed;
- [x] R4 Document/Revision/Submission lifecycle closed;
- [x] R5 Numbering/TemplateSpec/metadata boundaries closed;
- [x] R6 Periodic Review/Rendition/Release closed;
- [x] R7 Distribution/Values/Audit/Notifications/Search closed;
- [x] R8 Tenant lifecycle/Security closed;
- [ ] R9 final Permission + Role + Constraint matrix closed;
- [ ] whole-product domain map promoted as closed after R9 adversarial pass;
- [ ] build-vs-buy final for each responsibility;
- [ ] bounded contexts/data model/table + transaction ownership closed;
- [ ] event/API/frontend contracts closed;
- [ ] migration/deletion map closed;
- [ ] final ADR/spec set promoted;
- [ ] adversarial review finds no material ambiguity;
- [ ] operator approves integrated design;
- [ ] implementation plan authored from accepted target.

Until then: **design/documentation only.**

---

# 16. Exact next step

Continue **R9 — Final Authorization Matrix**.

Do not implement. For every operation determine independently:

1. semantic Permission required, if any;
2. legal scope (Tenant / Area / case-specific relationship);
3. which built-in roles bundle that Permission;
4. whether a workflow/domain relationship is additionally mandatory;
5. lifecycle/state preconditions;
6. SoD/fresh-auth/tenant-operability constraints;
7. collection/filter visibility;
8. DB/RLS/constraint backstop;
9. at least one positive and one negative Golden Matrix case.

Do not create a Permission merely because a current capability exists. Do not omit a Permission merely to minimize count. Permission boundaries follow independently delegable business authority.