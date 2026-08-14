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

The redesign began with authorization drift, expanded into Approval, and then proved that `documents`, `controlleddocuments`, `templates`, taxonomy, IAM, Approval, rendering and release contain overlapping authorities.

Strongest counterexample: browser QA let a human review edited content while freeze rendered a blank template snapshot. The final PDF/hash therefore did not represent what the human reviewed.

**Target property:** every business fact has one authority; supporting concerns consume it instead of reinterpreting or mutating it independently. Current code/schema/API are migration evidence, not target-design authority.

---

# 2. Target responsibility map

| Current concern/module | Target disposition |
|---|---|
| `auth` | retain V1 authentication/session implementation behind stable AuthN boundary |
| `iam` | conceptually split into **Organization** + **Authorization** |
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
| `security` | R8 after tenant/AuthN seam closes |
| `jobs` | orchestration infrastructure, not business bounded context |

---

# 3. LOCKED — Authentication + Organization + Authorization

## AUTHN-01 — Separate authorities

- Authentication: who is this actor/session?
- Authorization: what may this principal do in this tenant/scope?
- Approval: who participates in this concrete submission?
- Domain Governance: is this action legal now given lifecycle, SoD, reauth and immutable-content rules?

None substitutes for another.

## AUTHN-02 — No Keycloak now

Current MetalDocs AuthN is sufficient for V1. Preserve a stable authenticated-principal seam for future OIDC/SAML/enterprise IdP/MFA/passkeys when a real requirement appears.

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

## AUTHZ-01 — Five roles

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

## AUTHZ-04 — Candidate permissions

Current candidate set; final freeze happens in R9:

```text
document.read_published
document.read_working
document.create
document.edit
document.comment
document.submit
document.obsolete
document.review_periodic

approval.act
approval.oversee
approval.cancel
approval.reassign
approval_policy.manage

organization/access-administration permissions — R9

distribution/acknowledgement permissions — R9
```

`document.supersede` is not presumed necessary: same-Document Revision supersession is mechanical release behavior.

## AUTHZ-05 — Decision composition

```text
current permission/qualification
+ workflow/domain relationship where applicable
+ Domain Governance constraints
= ALLOW
```

No tenant-owner bypass. No OpenFGA/SpiceDB V1; revisit only for material arbitrary resource-sharing graphs.

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

## CI-05 — Template is a role of a governed Document

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

## REV-06 — RevisionSubmission is first-class immutable evidence

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

No cron/expression/rules engine V1.

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

## RELEASE-01 — RevisionSubmission is the release candidate identity

`ReleaseGeneration` is no longer a required domain noun. Release Coordinator evaluates `submission_id`.

## RELEASE-02 — Automatic centralized release

No human publish button. Approval completion, rendition readiness, timer and reconciliation all invoke idempotent `EvaluateRelease(submission_id)`.

Release checks candidate/open Revision, active Submission, Approval requirement, mandatory OFFICIAL_PDF provenance, optional `ReleasePlan.not_before`, and one-effective-revision invariants.

## RELEASE-03 — Actual Effectivity

`ReleasePlan.not_before?` is a gate, not a `SCHEDULED` state.

```text
effective_at = released_at
```

in the winning transaction. No silent retroactive effectivity.

## RELEASE-04 — Winning transaction + evidence

Atomically:

```text
candidate REV -> EFFECTIVE
prior effective REV -> SUPERSEDED
Document.effective_revision_id -> candidate
Document.open_revision_id -> NULL
insert immutable ReleaseRecord
emit/enqueue lifecycle events
```

`ReleaseRecord` binds Document, Revision, Submission, OFFICIAL_PDF, previous Revision and actual release/effective timestamps with system actor.

Candidate may be CANCELLED after Approval but before Release; Approval/Rendition evidence remains historical.

## RELEASE-05 — effective_date token removed from pre-release TemplateSpec V1

Actual `effective_at` is born at Release, after the mandatory pre-release PDF must exist. Therefore legacy `effective_date` cannot survive as a mandatory pre-release system field by inertia.

---

# 9. LOCKED — Distribution + Values + Audit + Notifications + Search (R7)

## DIST-01 — Distribution = obligation/acknowledgement, not AuthZ or Training

Authorization answers who **may** read; Distribution answers who **must** take explicit notice of an EFFECTIVE Revision.

No quiz, assessment, learning path, training-plan or LMS engine V1. If competency/training becomes a real requirement, create a separate Training context later.

Distribution never edits RoleAssignments to manufacture access. Exact task-specific visibility interaction is finalized in R9.

## DIST-02 — Configuration on Document

Conceptually:

```text
DistributionConfiguration =
    None
  | ReadAcknowledgement {
      targets: User | Group
      due_in_days?
      requires_reauthentication
    }
```

No Area target V1 because Organization has no independent UserAreaMembership concept; use explicit Groups instead.

Changing configuration affects future releases by default. Applying to the currently effective Revision is an explicit audited assignment operation, never silent retroactivity.

## DIST-03 — Release snapshots concrete users

At Release, User targets and Group memberships are resolved into concrete per-user `DistributionAssignment`s.

Group membership changes later never rewrite historical assignments. Post-release onboarding/reassignment uses explicit assignment against the current effective Revision.

One obligation per user per release/Revision; multiple source targets are provenance, not duplicate assignments.

## DIST-04 — Assignment states and revision rollover

```text
pending
acknowledged
cancelled
superseded
```

Overdue is derived from `pending + due_at`, not another persisted state.

When a newer REV becomes EFFECTIVE, pending obligations for the prior REV become `superseded`; acknowledged history stays immutable. New assignments are materialized for the new release using current configuration.

## DIST-05 — Explicit acknowledgement

Opening a notification, viewing a document or downloading a PDF does **not** complete a distribution obligation.

Explicit action creates immutable:

```text
AcknowledgementRecord
  assignment_id
  actor_user_id
  acknowledged_at
  meaning = read_and_acknowledge
  reauth_evidence?    // when configured
```

Fresh reauth reuses the same AuthN assurance seam used by Approval; no separate signature engine.

## VALUE-01 — System Value Catalog vs tenant Dictionary

System/computed values are product-owned closed contracts. Tenant Dictionary is tenant-owned mutable name/value configuration.

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

Legacy mapping/disposition:

```text
doc_code            -> document_code
doc_title           -> revision_title
revision_number     -> revision_label      // e.g. REV004
author              -> revision_created_by_name
controlled_by_area  -> document_area_code / document_area_name
approval_date       -> not a TemplateField; Approval manifestation in official PDF
author/approvers    -> no live approval-list TemplateField
effective_date      -> not a pre-release TemplateField V1
```

System keys do not become tenant-provided SQL/scripts/custom resolvers.

## VALUE-02 — Mutable external values are snapshotted at REV creation

Dictionary and other mutable external values referenced by authoring schema are resolved when a new DocumentRevision is created and copied into its RevisionContent/provenance.

Return-for-changes/resubmission on the same REV does **not** silently re-resolve them. A new REV resolves current values again.

Historical Revision never depends on live Dictionary state. Provenance retains key/id, resolved value and resolution context sufficient to explain the historical result.

System key meaning is a stable contract; incompatible semantic change uses a new key or explicit migration rather than a hidden resolver-version engine.

## AUDIT-01 — Domain evidence vs global Audit Trail

Business authorities remain their own records:

```text
RevisionSubmission
ApprovalDecision
PeriodicReviewRecord
ReleaseRecord
DistributionAssignment
AcknowledgementRecord
RoleAssignment history
...
```

`AuditEvent` is a transversal compliance/investigation timeline and never substitutes for these records. Business logic must not query Audit Trail to discover whether Approval, acknowledgement or release happened.

## AUDIT-02 — Durable audit intent for governed mutations

A critical governed mutation must not report success unless an audit intent/event is durably persisted in the same commit boundary (direct same-transaction append or transactional outbox; exact mechanism R10).

Examples include role grant/revoke, submit, approval decision, return/cancel, release, obsoletion, acknowledgement, periodic review, responsible-owner change, distribution config change and policy administration.

Usage telemetry such as view/download/search/notification-read may be asynchronous and must not block the user-facing read path merely because analytics delivery is temporarily unavailable.

## AUDIT-03 — Integrity/export/system actors

Preserve the property: Audit Trail is append-only, tamper-evident and exportable. Current hash-chain implementation is evidence of a valid mechanism but final technical shape is R10.

Actors may be User or explicit System principal (for example release coordinator). Never attribute automatic release to the last approver.

## NOTIF-01 — Notifications are projection/delivery only

Notifications consume canonical domain events and create in-app/e-mail delivery artifacts. Notification state never becomes workflow/distribution/review authority.

`Notification.READ` means the **notification** was read, not the document and not an acknowledgement.

“Minhas Pendências” must query actual authorities (active Approval participation, pending DistributionAssignments, due Periodic Review work), not unread notifications.

Reminders derive from business due facts; notifications only deliver them.

## SEARCH-01 — Search is rebuildable projection

Search is discovery, never canonical Document state or authorization truth. It may be eventually consistent and must be replay/rebuild capable.

Two conceptual surfaces:

```text
Official Library -> current EFFECTIVE Revision
Working Search   -> open Revision only for users with working-content access
```

Historical SUPERSEDED/OBSOLETE Revisions stay out of global search by default and remain accessible through version history/audit-specific surfaces.

Search results are filtered against current AuthZ; stale result never grants access. Canonical resource endpoint always rechecks authorization.

No Elasticsearch/OpenSearch requirement is assumed. R10 chooses PostgreSQL FTS vs dedicated engine from measured query/scale requirements.

---

# 10. Build-vs-buy rulings to date

| Technology/class | V1 ruling | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | do not deploy now | enterprise SSO/federation/MFA/passkeys or credential externalization |
| OpenFGA / SpiceDB | do not deploy now | arbitrary resource sharing / large relationship graph / service split |
| Camunda / Flowable / BPMN | do not use for document Approval V1 | product genuinely becomes generic business-process engine |
| Temporal as Approval/Release engine | do not use | durable orchestration requirement current outbox/jobs cannot economically serve |
| CEL / expression language | do not use now | real conditional product/workflow policies cannot be represented cleanly by typed configuration |
| Elasticsearch/OpenSearch | no requirement yet | measured search scale/query needs exceed economical PostgreSQL projection |

Libraries/frameworks are selected only after exact responsibility closure.

---

# 11. Explicit target deletions/replacements

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
- any render/freeze path choosing bytes other than exact Submission;
- `SCHEDULED` Revision state;
- automatic expiration merely because periodic review is overdue;
- live Group-membership derivation as historical distribution denominator;
- notification-read as document-read/ack evidence;
- live Dictionary references from historical released Revisions;
- Audit Trail as a business-state database;
- Search projection as authorization/business-state authority;
- old roadmap/milestone/spec documents as live authority.

Prefer deletion over compatibility shims when no deployed/contractual compatibility requirement proves a shim necessary.

---

# 12. Remaining design queue before implementation

## R8 — Tenant lifecycle + security — **NEXT**

Close:

- Tenant as Organization authority after IAM split;
- tenant owner vs platform/system operator;
- tenant creation/bootstrap/onboarding;
- tenant suspension/deactivation semantics;
- export and deletion request/grace/cancel/erasure lifecycle;
- legal/audit preservation vs crypto-shred/anonymization boundaries;
- session/user effects during suspension/erasure;
- password/reset/session/security-signal ownership;
- MFA/reauth assurance model and whether V1 needs tenant MFA policy;
- tenant encryption/key ownership and rotation requirements;
- exact trigger/seam for future external IdP/Keycloak migration.

## R9 — Final Authorization Matrix

After every product operation exists:

- final Permission Catalog;
- five role bundles;
- Organization/admin operations;
- Distribution and Periodic Review operations;
- collection filtering/visibility;
- Approval/Distribution/domain relationship checks;
- Domain Constraint matrix;
- RLS/DB tripwire backstops;
- positive/negative Golden Matrix.

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

# 13. Documentation authority

- `wiki/architecture/cohesive-platform-redesign.md` is the current-program entrypoint.
- this ledger is the only active detailed WIP decision source under `docs/superpowers`.
- `wiki/references/current-agent-handoff.md` is a short recovery pointer.
- Git history is the archive for deleted historical staging documents and prior ledger forms.
- legacy wiki/module/ADR material may explain runtime/history but cannot override operator-approved target decisions here.
- final integrated decisions are promoted to durable wiki/ADR/spec authorities only after design closure.

---

# 14. Implementation gate

Implementation starts only when all are true:

- [ ] whole-product domain map closed;
- [x] Organization/AuthZ north star closed;
- [x] Approval V1 closed;
- [x] R3 Controlled Information configuration closed;
- [x] R4 Document/Revision/Submission lifecycle closed;
- [x] R5 Numbering/TemplateSpec/metadata boundaries closed;
- [x] R6 Periodic Review/Rendition/Release closed;
- [x] R7 Distribution/Values/Audit/Notifications/Search closed;
- [ ] R8 Tenant lifecycle/Security closed;
- [ ] R9 final Permission + Role matrix closed;
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

# 15. Exact next step

Continue **R8 — Tenant Lifecycle + Security**.

Do not implement. For every lifecycle/security proposal answer:

1. who is the authority — tenant actor vs platform operator vs system;
2. whether the action is reversible, grace-period based or terminal;
3. what must survive for compliance/audit and what must become unreadable/erased;
4. what happens to sessions, users, jobs and object storage;
5. whether a feature belongs in current AuthN or is a trigger for an external IdP/security product;
6. what is the minimum V1 mechanism that preserves the future seam.