# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding for this redesign; unresolved items are explicit.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Governing method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Program entrypoint:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED. No product code/schema/OpenAPI/frontend/migration implementation is authorized yet.**

---

## 0. Fresh-session contract

Read, in order:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. this ledger
5. `wiki/references/current-agent-handoff.md`

Never resume deleted/historical roadmaps, milestones, specs, old `docs/superpowers` material, superseded ADR semantics or stale implementation PRs by inertia.

Target method:

```text
product/domain truth
→ invariants + lifecycle
→ Organization/AuthZ/Approval integration
→ build-vs-buy
→ bounded contexts/dependency DAG
→ data model + DB constraints
→ transaction/event contracts
→ API + frontend journeys
→ migration/delete map
→ implementation specification
→ implementation plan
→ code
```

Global maximum means **the smallest architecture that correctly models the domain, preserves invariants and exposes clean extension seams** — not maximum abstraction.

---

# 1. Why this is a whole-platform redesign

The reset began with authorization drift, expanded into Approval, and proved that `documents`, `controlleddocuments`, `templates`, taxonomy, IAM, Approval, rendering and release had overlapping authorities.

The strongest production-shape counterexample came from browser QA: an author edited one content body and an approver reviewed it, while freeze/render later selected the blank template source. The official artifact therefore did not represent the content humans reviewed.

**Root cause:** locally reasonable modules evolved without one coherent controlled-information model.

**Target property:** every business fact has exactly one authority; supporting concerns consume that fact rather than reinterpret/mutate it independently.

Existing code/schema/OpenAPI are **current-state and migration evidence**, not target-design authority.

---

# 2. Target responsibility map

| Current concern | Target disposition |
|---|---|
| `auth` | retain local AuthN V1 behind a stable principal/assurance seam |
| `iam` | conceptually split into **Organization** + **Authorization** |
| `approval` | small specialized human Approval engine; no document-state ownership |
| `documents` | evolves into core **Controlled Information** |
| `controlleddocuments` | retire as target context/concept |
| `templates` | retire parallel lifecycle; template is a role of governed Document/Revision |
| `taxonomy` | dismantle: Area→Organization; Profile→DocumentType; Family→category; GovernanceClass deleted |
| `render` | supporting Rendition infrastructure over exact RevisionSubmission |
| `distribution` | real obligation/acknowledgement domain over EFFECTIVE Revisions |
| `tokens` | split conceptually into product System Value Catalog + tenant Dictionary |
| `audit` | append-only/tamper-evident transversal trail; never business-state authority |
| `notifications` | projection/delivery only |
| `search` | rebuildable/evenually-consistent projection |
| `security` | split: AuthN security/admin views + platform tenant crypto; heuristic signals optional |
| `jobs` | infrastructure/orchestration, not business bounded context |
| tenant lifecycle in `iam` | move conceptually to Organization/Platform lifecycle boundary |

---

# 3. LOCKED — Authentication, Organization, Authorization north star

## AUTHN-01 — Separate authorities

- Authentication: **who is this actor/session?**
- Authorization: **what may this principal do in this tenant/scope?**
- Approval: **who participates in this concrete Submission?**
- Domain Governance: **is this action legal now given lifecycle, SoD, fresh-auth and immutable-content rules?**

None substitutes for another.

## AUTHN-02 — No external IdP now

Current local AuthN remains V1. Preserve a principal seam for future OIDC/SAML/MFA/passkeys/federation. Future external identity maps `(issuer, subject) → internal User`; MetalDocs remains authority for Organization/AuthZ/workflow relationships.

## ORG-01 — Organization

```text
Tenant
Area
User
Group
GroupMembership
```

Area is one organizational truth reused by Document ownership, RoleAssignment scope and Approval actor resolution.

Groups are flat V1; users may belong to multiple groups; no nested groups.

## AUTHZ-01 — Five tenant roles only

```text
tenant_owner
area_manager
author
approver
viewer
```

Roles are bundles; runtime asks for Permissions and constraints, never role equality.

## AUTHZ-02 — One assignment shape

```text
RoleAssignment
  subject: User | Group
  role
  scope: TenantScope | AreaScope
  granted_by / granted_at
  revoked_by? / revoked_at? / reason?
```

Additive grants + default deny. No explicit-deny engine, temporal scheduler or generic ACL/ReBAC graph V1.

Legal role scopes:

| Role | Tenant | Area |
|---|---:|---:|
| tenant_owner | yes | no |
| area_manager | no | yes |
| author | yes | yes |
| approver | yes | yes |
| viewer | yes | yes |

Effective roles = direct User assignments ∪ Group-inherited assignments.

## AUTHZ-03 — No bypass

`tenant_owner` satisfies a broad Permission bundle but **never bypasses** participant rules, SoD, state, fresh-auth, tenant operability, immutable-content or cross-tenant constraints.

No OpenFGA/SpiceDB V1. Revisit only if arbitrary per-resource relationship graphs become material.

---

# 4. LOCKED — Approval V1

## APPR-01 — Specialized human workflow, not BPM

No BPMN, generic gateways/service tasks, Camunda/Flowable, CEL, M-of-N, weighted voting, generic delegation or branching process engine V1.

## APPR-02 — Versioned sequential policy

```text
ApprovalPolicy
  id
  version
  ordered ApprovalStep[]

ApprovalStep
  order
  name
  purpose: review | approval
  actor_rule
  completion: ANY | ALL
  requires_reauthentication
  due_in_days?
```

Actor rules V1:

```text
NamedUser
Group
RoleInArea(role, fixed-area | subject-area)
```

Participants materialize when a Step activates and are snapshotted as evidence; current qualification is rechecked when acting.

## APPR-03 — Human outcomes

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

No normal terminal human `reject` V1.

## APPR-04 — Content-exact attempts

ApprovalInstance binds one immutable `RevisionSubmission`.

`return_for_changes` terminates that Approval attempt and returns the same business REV to DRAFT. Editing and resubmission create a **new RevisionSubmission** and, when approval is required, a **new ApprovalInstance**.

Deadlines only surface overdue work. Reassignment is explicit/audited repair. Fresh-auth may be required per Step.

---

# 5. LOCKED — Controlled Information configuration (R3)

## CI-01 — Core concepts

```text
Document
DocumentRevision
```

The target split `documents + controlleddocuments + templates` is rejected.

## CI-02 — DocumentType

`DocumentProfile` is replaced by tenant-scoped `DocumentType`:

```text
id
 tenant_id
 code          // immutable
 name
 description?
 category_id?
 status: ACTIVE | INACTIVE
```

No independent versioning V1. Inactive prevents new use; existing Documents remain valid. Document type is immutable after creation.

## CI-03 — Category only

`DocumentFamily` becomes optional `DocumentTypeCategory` for navigation/reporting only. No policy inheritance.

## CI-04 — GovernanceClass deleted

`controlado/simples/livre` is not an authority. Each domain owns explicit configuration.

Approval configuration:

```text
NoHumanApproval
or
UsePolicy(ApprovalPolicyID)
```

No nullable ambiguity and no fake zero-stage route.

## CI-05 — Template is a governed Document role

Templates have no parallel lifecycle/version counter. Template changes use ordinary DocumentRevisions and official `REVxxx` labels.

`TemplateUse` is M:N between template Document and target DocumentType; at most one UX default per type. Blank creation remains valid V1. Creation resolves the template Document's **current EFFECTIVE REV once** and pins exact origin forever.

---

# 6. LOCKED — Document / Revision / Submission lifecycle (R4)

## REV-01 — Stable Document identity

Document owns stable/business identity facts and pointers to current effective/open Revisions.

## REV-02 — Official revision labels

```text
REV001
REV002
REV003
...
```

`REVxxx` is the business/audit/user-visible revision. Technical IDs/OCC/schema/policy versions are distinct namespaces.

## REV-03 — Revision = business change cycle

Autosaves/checkpoints/editor snapshots are technical history inside an open Revision and never consume REV numbers.

At most one open Revision per Document V1; one EFFECTIVE Revision may coexist with that open Revision.

## REV-04 — States

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

No `APPROVED`, `SCHEDULED`, `PUBLISHED` Revision state.

`return_for_changes` / allowed withdraw: `SUBMITTED → DRAFT` on the **same REV**.

## REV-05 — RevisionSubmission

Immutable first-class attempt identity:

```text
REV002
  ├── Submission #1 → digest A → returned
  └── Submission #2 → digest B → released
```

Exists even for `NoHumanApproval`. Approval, Rendition and Release always bind the exact Submission/digest.

After completed Approval, V1 does not reopen candidate content; cancel candidate + create next REV if content must change.

## REV-06 — Superseded vs obsolete

`SUPERSEDED` = newer Revision of the same Document became effective.

`OBSOLETE` = the Document was intentionally retired without successor. Obsolete is terminal V1.

Cross-Document replacement is a future distinct concept, not `document.supersede` for ordinary revisioning.

---

# 7. LOCKED — Numbering / TemplateSpec / metadata (R5)

## NUM-01 — Document code

`Document.code` is immutable and unique tenant-wide; Document type, Area and code are immutable V1. `DocumentType.code` and `Area.code` are stable identifiers.

Code allocation + Document + REV001 creation are one atomic operation. Successful creation permanently consumes the sequence; rollback before creation need not.

## NUM-02 — Deliberately small numbering language

DocumentType numbering supports only literals +:

```text
{TYPE}
{AREA}
{SEQ}
```

Sequence scopes:

```text
TYPE
TYPE_AREA
```

`sequence_width` is minimum padding. No year/month/custom field/formula/script/reset engine V1.

Normal Create has no manual code override. Legacy-code preservation belongs to explicit import/migration.

## TPL-01 — TemplateSpec only owns authoring contract

Template does not own numbering, retention, distribution or generic metadata policy.

```text
TemplateField
  key
  label
  value_type: text | date | number | choice | user | image
  source: user_input | system(key) | dictionary(key)
  typed constraints...
  visible_if?
```

Typed constraints and a closed `visible_if` operator set survive; no generic expression language.

DOCX/source anchors and TemplateSpec must agree before template submission.

## TPL-02 — Seed then independent truth

Creating from template copies/pins exact source REV content + TemplateSpec into derived REV001 and records immutable `DocumentOrigin` provenance. Derived Documents never silently rebind.

## TPL-03 — RevisionContent

Governed content is one logical identity, potentially including source artifact/hash, authoring schema snapshot, field values and governed Revision metadata. `RevisionSubmission.submission_digest` covers the complete governed identity.

`title` belongs to DocumentRevision. Stable/operational facts live on Document. No generic tenant custom-metadata engine V1.

---

# 8. LOCKED — Periodic Review / Rendition / Release (R6)

## REVIEW-01 — Periodic Review

Owned by Controlled Information, not Approval.

```text
PeriodicReviewPolicy = Disabled | Every(n months)
```

Cadence starts at actual Effectivity and restarts after completed review. Due/overdue does **not** invalidate EFFECTIVE content.

Append-only `PeriodicReviewRecord` binds exact effective REV and records:

```text
confirmed_current
change_required
```

`change_required` does not auto-create a REV. Stale review completion fails if the reviewed REV is no longer effective.

Documents subject to review require a valid responsible owner. Ownership is responsibility, not authorization.

## REND-01 — Rendition

Immutable derived representation of exact RevisionSubmission with own hash and renderer/build provenance.

Only `OFFICIAL_PDF` is mandatory for Release V1. `FINAL_DOCX` may exist but does not block effectivity.

Approval approves Submission, not PDF bytes. Official PDF is an attested derivation and may manifest approval/signature information.

## RELEASE-01 — Automatic centralized effectivity

No human publish button. `RevisionSubmission` is the release-candidate identity; `ReleaseGeneration` is not a required domain noun.

Approval completion, rendition readiness, timer and reconciliation invoke idempotent `EvaluateRelease(submission_id)`.

Optional `ReleasePlan.not_before?` is a gate, not a Revision state. Actual business fact:

```text
effective_at = released_at
```

No silent retroactive effectivity.

Winning transaction atomically:

```text
candidate REV -> EFFECTIVE
prior effective REV -> SUPERSEDED
Document.effective_revision_id -> candidate
Document.open_revision_id -> NULL
insert ReleaseRecord
emit/enqueue events
```

Candidate can be CANCELLED after Approval but before release; historical evidence remains.

Legacy `effective_date` is not a pre-release TemplateField V1 because actual effectivity is born after mandatory pre-release PDF creation.

---

# 9. LOCKED — Distribution / Values / Audit / Notifications / Search (R7)

## DIST-01 — Distribution

Distribution is controlled obligation/read acknowledgement, not Authorization and not Training/LMS.

Configuration on Document:

```text
None
or
ReadAcknowledgement {
  targets: User | Group
  due_in_days?
  requires_reauthentication
}
```

No Area target V1 without a real UserAreaMembership concept.

At Release, targets/groups resolve to concrete per-user `DistributionAssignment`s. Later membership changes never rewrite historical denominator. One obligation per user/release; multiple target sources are provenance.

Assignment states:

```text
pending
acknowledged
cancelled
superseded
```

Overdue is derived. New effective REV supersedes pending old-REV obligations and creates a new cohort.

Opening notification/viewing/downloading does not acknowledge. Explicit immutable `AcknowledgementRecord` does. Fresh-auth is optional per configuration.

Distribution never edits RoleAssignments.

## VALUE-01 — System values vs Dictionary

Product-owned System Value Catalog V1:

```text
document_code
revision_label
revision_title
document_type_code
document_area_code
document_area_name
revision_created_by_name
```

`approval_date`/approvers are official-PDF manifestation, not TemplateFields. Actual `effective_date` remains outside TemplateSpec V1.

Tenant Dictionary is mutable source data. Mutable external values resolve/snapshot when a **new REV is created**; same-REV return/resubmit never silently re-resolves. Released/historical content never depends on live Dictionary state.

## AUDIT-01 — Business evidence vs Audit Trail

Domain records (`RevisionSubmission`, ApprovalDecision, PeriodicReviewRecord, ReleaseRecord, DistributionAssignment, AcknowledgementRecord, RoleAssignment history, etc.) remain business authorities.

AuditEvent is transversal compliance/investigation timeline and never substitutes for them.

Critical governed mutation must not report success without durable audit intent/event in the same commit boundary (direct append or transactional outbox; exact mechanism R10).

Usage telemetry (view/download/search/notification-read) may be async and must never equal acknowledgement.

Preserve Audit Trail property: append-only, tamper-evident, exportable, explicit User/System actors.

## NOTIF-01 — Notifications

Projection/delivery only. Notification `READ` means notification read, never document read/acknowledged. “Minhas Pendências” is composed from canonical Approval/Distribution/Review facts, not notification state. Reminders derive from business due facts.

## SEARCH-01 — Search

Rebuildable/eventually-consistent projection only.

```text
Official Library -> current EFFECTIVE REV
Working Search   -> open REV under current working-content access
```

Historical superseded/obsolete Revisions stay out of global search by default. Search result never grants access; canonical resource endpoint rechecks AuthZ.

No Elasticsearch/OpenSearch requirement yet.

---

# 10. LOCKED — Tenant lifecycle / Platform Security (R8)

## TENANT-01 — Platform authority outside tenant RBAC

`PlatformOperator` and `SystemPrincipal` are platform identities, not tenant Roles/RoleAssignments.

PlatformOperator can create/suspend/resume tenants and inspect lifecycle operations but has **no implicit tenant-content access**. Break-glass/support access is outside V1.

## TENANT-02 — Tenant lifecycle

```text
ACTIVE
SUSPENDED
ERASED
```

Deletion request is separate:

```text
TenantDeletionRequest
  PENDING | CANCELLED | EXECUTED
  requested_by
  requested_at
  execute_after
```

Grace period is product/deployment policy. Tenant remains ACTIVE while request is pending and may cancel.

## TENANT-03 — Onboarding

Platform creation yields:

```text
Tenant
+ initial User
+ tenant_owner @ Tenant
+ single-use time-limited activation credential
```

No historical `system_admin`; platform operator never chooses/knows tenant owner's password.

## TENANT-04 — Suspension/deactivation

Suspension revokes tenant sessions, blocks login/business mutations, preserves data and is reversible. Business jobs respect suspension; lifecycle/security jobs may continue.

User deactivation revokes that user's sessions and preserves immutable identity/evidence/history. Pending responsibilities require explicit repair/reassignment.

## TENANT-05 — Export/deletion

Tenant owner may export own tenant and request/cancel own deletion; both require fresh-auth. PlatformOperator has no implicit export artifact access.

## TENANT-06 — Erasure

Conceptually:

```text
request reaches execute_after
→ system suspends tenant
→ revoke sessions
→ erase live tenant-owned rows
→ delete tenant blobs
→ destroy Tenant DEK
→ preserve allowed non-PII audit/platform skeleton
→ Tenant ERASED
→ persist platform TenantErasureRecord
```

Audit Trail itself is not deleted in target. Retained skeleton uses opaque internal IDs/non-PII only; sensitive tenant payload is erasable/unreadable. Crypto-shred claims apply only to data actually protected by key scope.

Backup/restore must reapply erasure tombstones before service restoration to prevent tenant resurrection.

## SECURITY-01 — Small crypto primitive

```text
Platform KEK
  ↓ wraps
Tenant DEK
```

No per-document key hierarchy/rotation/HSM/escrow product V1.

## SECURITY-02 — Auth assurance seam

Conceptually:

```text
AuthenticatedPrincipal
  user_id
  tenant_id
  authenticated_at
  auth_method
  assurance / fresh-auth evidence
```

Approval, acknowledgement, tenant export/deletion can require fresh-auth without knowing whether today's mechanism is password or future IdP/MFA.

## SECURITY-03 — Remove fake MFA

Current MFA coverage is stub metadata without real enrollment and is not a V1 security control. Stub fields/cards have no target entitlement.

Real MFA/passkeys/SSO/SAML/per-tenant federation trigger re-evaluation of Keycloak/external IdP before building IdP features internally.

Sessions/credential lifecycle/lockouts/fresh-auth belong to AuthN; tenant envelope crypto belongs to Platform Security; heuristic security signals are optional observability projections, not core product requirements.

---

# 11. LOCKED — Final Authorization model (R9)

## R9-01 — Decision equation

Authorization is not one boolean source:

```text
BASE PERMISSION
+ required resource/case relationship (when applicable)
+ Domain Governance constraints
= ALLOW
```

Some operations are RBAC-only; some require RBAC + relationship; some narrow self/case operations are relationship-authorized without a broad tenant Permission.

Examples:

- `document.create`: Permission + scope + tenant ACTIVE.
- Approval accept: `approval.act` + active participant + SoD + optional fresh-auth.
- Periodic review: `document.review_periodic` + responsible-owner relation + exact REV still EFFECTIVE.
- Distribution acknowledgement: concrete assignee relation + active user + optional fresh-auth; no broad `distribution.acknowledge` Permission.

## R9-02 — Final tenant Permission Catalog (29)

### Tenant/configuration

```text
tenant.settings.manage
organization.manage
access.manage
document_type.manage
approval_policy.manage
template_use.manage
dictionary.manage
```

### Documents

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.comment
document.submit
document.cancel_revision
document.obsolete
document.review_periodic
document.owner.manage
```

### Approval

```text
approval.act
approval.oversee
approval.reassign
approval.cancel
```

### Distribution

```text
distribution.manage
distribution.oversee
```

### Compliance/security/lifecycle

```text
audit.read
audit.export
session.manage
tenant.export
tenant.deletion.request
```

No Permission exists merely because an old capability existed.

## R9-03 — Operations deliberately NOT modeled as tenant Permissions

Relationship/self/system authority covers:

```text
read/mark own notification
read own DistributionAssignment
acknowledge own DistributionAssignment
read exact Approval case as participant
read exact historical Submission acted upon
withdraw own active Submission (document.submit + submitted_by relation)
change own password
list/revoke own sessions
release/effectivity execution
rendition generation
tenant erasure execution
```

## R9-04 — Role bundles

### viewer

```text
document.read_effective
```

### author

Viewer +:

```text
document.read_history
document.read_working
document.create
document.edit
document.comment
document.submit
document.review_periodic
```

### approver

Viewer +:

```text
approval.act
```

Approver gets no blanket working/history access; Approval participation opens the exact case/Submission.

### area_manager

Author +:

```text
document.cancel_revision
document.obsolete
document.owner.manage
approval.act
approval.oversee
approval.reassign
approval.cancel
distribution.manage
distribution.oversee
```

No tenant IAM/config/audit/session/lifecycle administration.

### tenant_owner

All 29 tenant Permissions via ordinary Authorizer. **Still no bypass.**

## R9-05 — Templates use ordinary Document permissions

Template Documents use normal `document.*` lifecycle/authoring/Approval permissions. The historical `template.view/create/edit/submit/approve/publish/archive/manage` permission family is deleted. Only `template_use.manage` is template-specific administration.

## R9-06 — Organization vs access administration

`organization.manage` covers Users/Areas/Groups/basic organizational attributes and activation/deactivation flows.

`access.manage` covers GroupMembership and RoleAssignment because group membership can immediately grant inherited roles.

While Tenant is ACTIVE there must be at least one active tenant owner. Revoking/deactivating the last owner fails closed.

Deactivating a user who is still responsible owner of Documents requiring periodic review is blocked until reassignment. Approval participants may be deactivated, leaving the Step requiring explicit reassignment. Pending DistributionAssignments do not block deactivation and remain historical/repairable.

## R9-07 — Approval relationship + SoD

`approval.act` alone never authorizes a decision: actor must be a current participant for the active Step and satisfy Domain Governance.

SoD V1:

1. actor cannot `accept` if actor is the Revision creator **or** Submission submitter;
2. the same user cannot `accept` two different Steps of the same ApprovalInstance;
3. reassignment target must be active, currently qualified, not violate SoD and not already have satisfied a previous Step.

No role, including tenant_owner, bypasses SoD.

Area Manager can act on a Step that explicitly targets `RoleInArea(area_manager)` because the role bundle includes `approval.act`; no duplicate approver role is required.

## R9-08 — Case-specific visibility

Approval participant may access only the exact Submission/preview/evidence needed for that case and retains historical access to the exact Submission on which they acted; this does not open later submissions or unrelated drafts.

DistributionAssignment can grant the assignee narrow access to the exact effective Revision needed to fulfil acknowledgement, even without broad `document.read_effective`; it never opens unrelated Documents/open Revisions.

Responsible-owner relation does not by itself grant broad read/write; periodic-review completion also requires `document.review_periodic` in scope.

## R9-09 — Withdrawal/cancellation semantics

Submitter may withdraw own active Submission only when Domain Governance allows, using `document.submit + submitted_by` relation. Another Author cannot impersonate the submitter.

Area Manager/Tenant Owner may use `approval.cancel` administratively. `document.cancel_revision` abandons the entire open change cycle and may therefore close any active Approval attempt as a consequence.

## R9-10 — Typed scope; no magic tenant sentinel

Scopes are typed:

```text
TenantScope(tenant_id)
AreaScope(tenant_id, area_id)
```

Tenant-scoped grant covers all Areas in same tenant; AreaScope covers only exact Area. The historical string sentinel `areaCode = "tenant"` is not target architecture.

## R9-11 — RLS/DB boundaries

- **Authorizer:** RoleAssignment, Group inheritance, Permission and typed scope.
- **Domain:** participant/owner/assignee/submitter relationships; state; SoD; fresh-auth; tenant ACTIVE; immutable-submission checks.
- **RLS:** tenant isolation defense-in-depth only, not a second authorization system.
- **DB constraints:** structural invariants such as unique document code, one effective/open Revision, immutable evidence, legal role scopes, one default template/type, FKs/checks.

Current asserted-capability GUC/tripwire/system-admin-short-circuit mechanism has no right to survive. R10 may retain a mechanical proof-of-authorizer-check backstop if useful, but never a second RBAC implementation.

## R9-12 — Golden Matrix anchors

Mandatory future tests include at least:

- Viewer can read EFFECTIVE in scope, cannot read working or other Area.
- Author can edit another Author's DRAFT in same scope, never SUBMITTED or other Area.
- Approver role alone cannot browse drafts; active participant can read exact Submission but cannot edit.
- Participant who is author/submitter cannot accept.
- Same user cannot accept two Steps of same ApprovalInstance.
- Tenant Owner not participant cannot accept; participant + SoD valid can.
- Area Manager can reassign/cancel only in assigned Area.
- Submitter can withdraw own attempt; another Author cannot.
- Responsible owner + `document.review_periodic` can review exact effective REV; non-owner cannot.
- Distribution assignee without roles can read/ack exact assigned effective REV only.
- Notification READ does not alter DistributionAssignment.
- Last active tenant owner cannot be revoked/deactivated.
- Responsible owner cannot be deactivated while review-governed Documents still point to them.
- SUSPENDED tenant rejects normal business mutations.
- PlatformOperator cannot read tenant content merely by platform authority.
- Release rejects OFFICIAL_PDF from another Submission.
- stale Search result never grants canonical resource access.

Golden Matrix tests evaluate **Permission + scope + relationship + resource state + domain constraint**, not role-only happy paths.

---

# 12. Build-vs-buy rulings to date

| Technology/class | V1 ruling | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | no now | enterprise SSO/federation, real MFA/passkeys, tenant-specific IdP, credential externalization |
| OpenFGA / SpiceDB | no now | arbitrary resource sharing / large relationship graph / service split |
| Camunda / Flowable / BPMN | no for document Approval V1 | true generic process-engine product requirement |
| Temporal | no for current Approval/Release | durable orchestration requirements outgrow economical outbox/jobs |
| CEL/expression language | no now | typed configuration cannot represent real conditional requirements |
| Elasticsearch/OpenSearch | no requirement yet | measured search needs exceed economical PostgreSQL projection |
| LMS/training engine | no requirement | competency/assessment/training-plan requirements appear |

---

# 13. Explicit target deletions/replacements

No entitlement to survive:

- `documents`/`controlleddocuments`/`templates` as three target contexts;
- public/domain `ControlledDocument` identity;
- `DocumentProfile`, behavioral DocumentFamily, GovernanceClass;
- parallel TemplateVersion lifecycle/version counter;
- template MetadataSchema as numbering/retention/distribution authority;
- duplicate DocCodePattern; normal-create manual code override; CompositionJSON without need;
- user-visible `v7` revision semantics; autosaves as business Revisions;
- Document lifecycle carrying workflow/revision states;
- StageKind as separate engines, configurable required capability, M-of-N, drift policies, generic delegation, normal terminal reject;
- ApprovalInstance reuse after content edits;
- generic BPMN/CEL/branching process machinery;
- role-based authorization bypasses/multiple grant engines/editable_by_role;
- Approval ownership of document state, periodic review or release;
- ReleaseGeneration as required business identity; mandatory final-DOCX release gate; SCHEDULED Revision state;
- auto-expiration solely because periodic review is overdue;
- live Group membership as historical distribution denominator; notification read as acknowledgement;
- live Dictionary references from historical Revisions;
- Audit Trail as business-state database; Search as authorization/state authority;
- `system_admin` tenant/platform super-role and platform operations represented as tenant RoleAssignments;
- fake MFA coverage/stub security control; heuristic Security Signals as mandatory core;
- tenant erasure deleting Audit Trail itself;
- old 8-role / 38-capability registry as target;
- old template capability family;
- `document.supersede`, `approval.sla_extend`, `notification.read`, generic `distribution.read`, tenant onboarding/erase inside tenant RBAC, unless a future independent operation proves them;
- magic `"tenant"` scope sentinel;
- system_admin capability short-circuit;
- current asserted-capability GUC mechanism as authorization authority;
- old roadmap/milestone/spec documents as live authority.

Prefer deletion over compatibility shims unless a deployed/contractual compatibility requirement proves a shim necessary.

---

# 14. Remaining design queue before implementation

## R10 — Integrated technical architecture — **NEXT**

The whole-product **domain/authorization model is now closed enough to descend into technical design**. R10 must not alter approved business semantics merely to fit current packages/tables.

Close, in order:

1. final bounded-context/module map and names;
2. dependency DAG / allowed imports / published ports;
3. exact aggregate ownership and application coordinators;
4. target table/schema ownership and DB constraints;
5. transaction boundaries for every governed operation;
6. audit-intent + outbox/event mechanics and domain-event catalogue;
7. async jobs/timers/reconciliation ownership;
8. object-storage ownership and immutable artifact keys;
9. build-vs-buy final pass (search, render, auth, jobs, crypto);
10. explicit current-module **KEEP / MOVE / REWRITE / DELETE** map;
11. explicit current-table **KEEP / TRANSFORM / DROP** map;
12. migration ordering/expand-contract/compatibility policy.

After R10:

### R11 — API + frontend journeys

- resources/URLs/commands/queries;
- RFC 9457 error semantics;
- DTOs and optimistic concurrency;
- full UI IA and journeys for Library, Working Docs, Approval, Distribution, Review, Admin, Audit, tenant lifecycle;
- no frontend guessing of authorization/business truth.

### R12 — Proof + final durable spec set

- threat/invariant/Golden Matrix tests;
- integration/QA contracts;
- promote final ADRs/specs/wiki authorities;
- adversarial architecture review;
- complete migration/deletion map.

### R13 — implementation specification and plan

Only after explicit integrated-design approval:

- classes/types/interfaces/packages;
- tables/indexes/constraints;
- endpoints/events/jobs;
- test matrix;
- sequenced implementation plan.

Then and only then product implementation.

---

# 15. Implementation gate

- [x] Authentication boundary decided
- [x] Organization/AuthZ north star decided
- [x] Approval V1 decided
- [x] R3 Controlled Information configuration
- [x] R4 Document/Revision/Submission lifecycle
- [x] R5 Numbering/TemplateSpec/metadata boundaries
- [x] R6 Periodic Review/Rendition/Release
- [x] R7 Distribution/Values/Audit/Notifications/Search
- [x] R8 Tenant lifecycle/Security
- [x] R9 final Permission/Role/relationship/SoD model
- [x] whole-product **business-domain map is closed enough to begin technical architecture**
- [ ] R10 bounded contexts/data/table/transaction/event/migration architecture
- [ ] R11 API/frontend journeys
- [ ] R12 proof matrix + final durable ADR/spec promotion + adversarial review
- [ ] operator approval of integrated code-ready design
- [ ] R13 implementation specification + implementation plan

Until all remaining gates close: **design/documentation only.**

---

# 16. Exact next step

Continue **R10 — Integrated Technical Architecture**.

Do not implement. Start from approved target semantics, not current folder names. Produce:

```text
bounded contexts
→ dependency DAG
→ aggregates/application coordinators
→ table ownership + constraints
→ transaction/event/outbox boundaries
→ async jobs/reconciliation
→ current code/table disposition map
```

Every existing module/table must earn one of: **KEEP, MOVE, REWRITE, DELETE**.