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

The target is the smallest professional architecture that makes important invalid states unrepresentable and leaves explicit extension triggers instead of speculative engines.

---

# 1. Whole-platform reset

The redesign began with authorization drift, expanded into Approval, and then proved that `documents`, `controlleddocuments`, `templates`, taxonomy, IAM, Approval, rendering and release contain overlapping authorities.

The strongest product counterexample came from browser QA: a human reviewed edited content while freeze rendered a blank template snapshot. The final PDF and its signed hash did not represent what the human reviewed.

**Root cause:** MetalDocs evolved as locally reasonable modules/features instead of one coherent controlled-information model.

**Target property:** every business fact has one authority; supporting concerns consume it rather than reinterpret or mutate it independently.

Current code/schema/API are migration evidence, not target-design authority.

---

# 2. Target responsibility map

| Current concern/module | Target disposition |
|---|---|
| `auth` | retain V1 authentication/session implementation behind stable AuthN boundary |
| `iam` | conceptually split into **Organization** + **Authorization** |
| `approval` | small specialized Approval V1; never owner of release/effectivity or periodic review |
| `documents` | becomes core of **Controlled Information** after cleanup |
| `controlleddocuments` | retire as target context; stable identity/numbering move to Document/configuration |
| `templates` | retire parallel lifecycle; template becomes role of governed Document/Revision |
| `taxonomy` | dismantle: Area → Organization; Profile → DocumentType; Family → category; GovernanceClass deleted |
| `render` | supporting rendition infrastructure bound to exact RevisionSubmission |
| `audit` | distinct evidence/integrity authority; exact seam still open |
| `distribution` | supporting released-revision concern; R7 |
| `notifications` | event consumer, never workflow authority |
| `search` | rebuildable projection/read model |
| `tokens` | supporting value provider; R7 snapshot semantics |
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

Current candidate set, still not final until R7-R9 close:

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

organization/access-administration permissions — later
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

Audited reassignment covers unavailable actors. Optional per-Step reauthentication survives. Approval evidence always pins actor, policy/version, Step, Submission/digest, outcome, reason/comment and reauth evidence when required.

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

Replaces behavioral Family as optional classification/navigation only. No inherited Approval, numbering, metadata or permissions.

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

Allocate next REV when change cycle starts. Never reuse a REV label even if cancelled.

`REV002+` requires `reason_for_change` before submission.

## REV-06 — RevisionSubmission is first-class immutable evidence

```text
REV002
  ├── Submission #1 → digest A → returned
  └── Submission #2 → digest B → released
```

Exists even for `NoHumanApproval` because release still needs exact immutable candidate identity.

Approval, Rendition and Release bind Submission, not mutable Document/Revision state.

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

Sequence scopes:

```text
TYPE
TYPE_AREA
```

`sequence_width` is minimum zero-padding. No year/month/custom fields/formulas/scripts/resets V1.

Normal Create has no manual-code override. Legacy code preservation belongs to explicit import/migration/bootstrap authority.

## TPL-01 — TemplateSpec only owns authoring contract

Delete template `MetadataSchema` as authority for numbering/retention/distribution/generic metadata.

A template Revision contains ordinary governed source content (e.g. DOCX) + `TemplateSpec`.

No separate TemplateVersion lifecycle and no `CompositionJSON` V1 without independent need.

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

`computed`/`dictionary` are sources, not types.

Typed constraints survive; invalid cross-type combinations fail closed. `visible_if` uses a closed operator set such as `eq/ne/gt/gte/lt/lte`, not a generic expression language.

## TPL-03 — Parity and provenance

Before template submission, source content anchors/tokens and TemplateSpec must agree.

Creating from template copies/pins the exact effective source Revision content + applicable TemplateSpec into the derived REV001 and records immutable `DocumentOrigin`:

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

For structured templates, governed content is one logical identity:

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

Periodic Review answers whether the currently EFFECTIVE Revision is still suitable. It is not Approval and does not use the Approval workflow engine.

DocumentType configuration:

```text
PeriodicReviewPolicy =
    Disabled
  | Every(n months)
```

No cron/expression/rules engine V1.

## REVIEW-02 — Cadence and legal status

Cadence starts at actual Effectivity and restarts after a completed periodic review.

```text
next_review_due_on = base_date + policy interval
```

Due/overdue is a review obligation/projection. It does **not** automatically change Revision state or invalidate EFFECTIVE content.

Expiration is a separate future policy and is not part of V1.

Changing the DocumentType review policy does not silently rewrite existing schedules. New policy applies at the next explicit calculation point (new Effectivity or completed review) unless a future audited bulk-recalculation operation is deliberately introduced.

## REVIEW-03 — Append-only review evidence

```text
PeriodicReviewRecord
  document_id
  reviewed_revision_id
  reviewed_by
  reviewed_at
  outcome
  comment?
  previous_due_on
  next_due_on
  policy_snapshot
```

Outcomes V1:

```text
confirmed_current
change_required
```

`confirmed_current` leaves the same REV EFFECTIVE and schedules the next review.

`change_required` records the finding but does not auto-create a new REV. An authorized Author starts the change cycle.

Before completion, the system verifies that the reviewed Revision is still `Document.effective_revision_id`; stale review attempts fail closed.

## REVIEW-04 — Responsible owner

A Document with Periodic Review enabled requires a valid `responsible_owner` relationship.

Owner is operational responsibility, **not authorization**. Review completion requires the relevant relationship/assignment plus `document.review_periodic` and current-Revision checks. Area Manager may reassign the owner with audit evidence.

## REND-01 — Source and Rendition are different concepts

`RevisionContent`/`RevisionSubmission` are governed source truth. A Rendition is an immutable derived representation of one exact Submission.

Conceptually:

```text
Rendition
  id
  submission_id
  kind
  storage_ref
  content_hash
  media_type
  size_bytes
  generated_at
  generator_component
  generator_version_or_build_digest
  derived_metadata?
```

Rendition always has its own output hash and explicit source Submission provenance.

## REND-02 — Mandatory artifact set V1

Only:

```text
OFFICIAL_PDF
```

is a mandatory derived artifact for Release V1.

A `FINAL_DOCX` rendition may exist for download/export but does not block Effectivity unless a later concrete regulatory/business requirement says otherwise.

Approval preview is not an official Rendition, but any preview shown to an approver must be derived from the same `submission_id/submission_digest` being decided.

## REND-03 — Approval approves Submission, not PDF bytes

Human Approval binds `RevisionSubmission.submission_digest`.

The official PDF is generated afterward as an attested derivation:

```text
Submission digest A
  ↓ renderer
Official PDF hash B
```

The PDF may manifest signer/approval data for human readability, but ApprovalDecision remains the authority.

## REND-04 — Attestation, not eternal bit-reproducibility

Persist the official artifact bytes, artifact hash, source Submission digest and renderer/build identity.

V1 does not promise that a renderer changed years later can reproduce byte-identical PDF output. The stored official artifact + provenance is the durable evidence.

Render failure is operational state/projection, never a DocumentRevision lifecycle state.

## RELEASE-01 — RevisionSubmission replaces ReleaseGeneration as domain identity

The old `release_generation` concept existed to bundle Document + revision + approval instance + revision version + frozen hash. The new immutable `RevisionSubmission` already provides the exact candidate identity.

Release Coordinator evaluates by `submission_id`.

A separate persistence record may still exist for attempts/facts if the later data model proves it useful, but **`ReleaseGeneration` is not a required business/domain noun**.

## RELEASE-02 — Release is automatic and centralized

There is no human `publish` button V1.

Triggers such as approval completion, rendition readiness, timer firing or reconciliation all invoke the same idempotent conceptual operation:

```text
EvaluateRelease(submission_id)
```

Release checks at least:

- candidate Revision is still SUBMITTED;
- Submission is still the active candidate for the open Revision;
- Approval requirement is satisfied (`NoHumanApproval` or completed Approval evidence);
- mandatory OFFICIAL_PDF exists and attests this exact Submission;
- optional `release_not_before` gate has been reached;
- Document still points to this open Revision;
- one-effective-revision/supersession invariants hold.

## RELEASE-03 — Planned time vs actual Effectivity

A small operational plan may carry:

```text
ReleasePlan.not_before?
```

`null` = release as soon as all gates are satisfied.

A future timestamp means “not before this instant”. It does not create a `SCHEDULED` Revision state.

Actual business fact:

```text
effective_at = released_at
```

in the winning transaction V1.

No silent retroactive Effectivity when artifacts or other gates are late.

## RELEASE-04 — Atomic winning transaction

Conceptually:

```text
BEGIN
lock Document + candidate Revision
revalidate Submission/gates
candidate REV -> EFFECTIVE
prior effective REV -> SUPERSEDED (if any)
Document.effective_revision_id -> candidate
Document.open_revision_id -> NULL
insert immutable ReleaseRecord
emit/enqueue lifecycle events
COMMIT
```

Exactly one winner. Retries/reconciliation are idempotent.

## RELEASE-05 — ReleaseRecord

Immutable evidence of the automatic act:

```text
ReleaseRecord
  document_id
  revision_id
  submission_id
  official_pdf_rendition_id
  previous_effective_revision_id?
  released_at
  effective_at
  actor = system
```

It links/reaches Approval evidence or the explicit `NoHumanApproval` configuration rather than copying all human decisions.

## RELEASE-06 — Cancellation and races

A candidate may be CANCELLED after Approval completion but before Release. Approval/Rendition evidence remains historical; Release then no-ops because the predicate no longer holds.

Release-vs-cancel races resolve through the same locked transaction boundary: whichever valid transition wins first makes the other predicate fail.

## RELEASE-07 — Effective-date token re-opened for R7

A pre-release mandatory PDF cannot depend on a fact (`effective_at`) that only exists after the Release transaction.

Therefore the historical `effective_date` computed-token semantics are **not carried forward by inertia**. R7 must explicitly decide whether that token is removed from pre-release content, uses planned semantics, is manifested after release, or becomes viewer metadata.

## RELEASE-08 — Infrastructure class remains modest

The domain requires durable events/timers/retries/idempotent evaluation, but nothing here proves a need for Temporal/Camunda/BPMN. The existing outbox + job class remains conceptually sufficient pending later technical architecture.

Cross-Document replacement stays outside V1 until a concrete requirement appears.

---

# 9. Build-vs-buy rulings to date

| Technology/class | V1 ruling | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | do not deploy now | enterprise SSO/federation/MFA/passkeys or credential externalization |
| OpenFGA / SpiceDB | do not deploy now | arbitrary resource sharing / large relationship graph / service split |
| Camunda / Flowable / BPMN | do not use for document Approval V1 | product genuinely becomes generic business-process engine |
| Temporal as Approval/Release engine | do not use | durable orchestration requirement current outbox/jobs cannot economically serve |
| CEL / expression language | do not use now | real conditional product/workflow policies cannot be represented cleanly by typed configuration |

Libraries/frameworks are selected only after exact responsibility closure.

---

# 10. Explicit target deletions/replacements

No entitlement to survive:

- target split `documents` / `controlleddocuments` / `templates`;
- separate public/domain `ControlledDocument` identity;
- `DocumentProfile`;
- behavioral `DocumentFamily` hierarchy;
- `GovernanceClass {controlado, simples, livre}`;
- parallel TemplateVersion lifecycle/version counter;
- template `MetadataSchema` as numbering/retention/distribution authority;
- duplicate template `DocCodePattern`;
- normal-create manual code override;
- `CompositionJSON` without independent requirement;
- user-visible `v7`-style document revisions;
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
- `ReleaseGeneration` as a required business/domain identity;
- mandatory final-DOCX gate V1;
- any freeze/render path that can choose bytes other than the exact Submission;
- `SCHEDULED` as Revision state;
- automatic expiration merely because periodic review is overdue;
- old roadmap/milestone/spec documents as live authority.

Prefer deletion over compatibility shims when no deployed/contractual compatibility requirement proves a shim necessary.

---

# 11. Remaining design queue before implementation

## R7 — Distribution / Read/Acknowledgement + Tokens/Computed Values + Audit/Evidence + Notifications + Search — **NEXT**

Close as one supporting-services pass because they all consume the canonical released/submitted identities rather than owning lifecycle.

### Distribution / read / acknowledgement

- who/what defines distribution obligations;
- User/Group/Area targeting semantics;
- snapshot denominator at release vs live derivation;
- read event vs explicit acknowledgement;
- whether acknowledgement ever requires reauth/signature;
- reminders/deadlines/export;
- what happens when Group membership changes after release;
- permissions and historical evidence.

### Tokens / computed / dictionary values

- exact system-token catalogue after R3-R6 redesign;
- resolve/freeze timing per value meaning;
- dictionary-value pinning semantics;
- resolver identity/version/provenance;
- collision rules;
- explicit decision for `effective_date`, `approval_date`, `approvers`, `revision_number`, `doc_title`, `controlled_by_area` and legacy names.

### Audit / evidence

- distinction between domain evidence records and global audit trail;
- which mutations require same-transaction audit/outbox evidence;
- hash-chain/integrity/export boundaries;
- actor/system-principal semantics;
- tenant deletion/erasure behavior.

### Notifications / Search

- domain event catalogue emitted by canonical authorities;
- notifications as replaceable consumer/projection, not business state;
- search as rebuildable projection;
- working vs effective visibility;
- eventual-consistency expectations and rebuild strategy.

## R8 — Tenant lifecycle + security

- Tenant authority after Organization split;
- tenant owner vs platform operator;
- deletion request/grace/system erasure;
- MFA/security signals/crypto key ownership;
- external IdP migration trigger/seam.

## R9 — Final Authorization Matrix

After every product operation exists:

- final Permission Catalog;
- five role bundles;
- Organization/admin operations;
- collection filtering/visibility;
- workflow/domain relationship checks;
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

# 12. Documentation authority

- `wiki/architecture/cohesive-platform-redesign.md` is the current-program entrypoint.
- this ledger is the only active detailed WIP decision source under `docs/superpowers`.
- `wiki/references/current-agent-handoff.md` is a short recovery pointer.
- Git history is the archive for deleted historical staging documents and prior ledger forms.
- legacy wiki/module/ADR material may explain runtime/history but cannot override operator-approved target decisions here.
- final integrated decisions are promoted to durable wiki/ADR/spec authorities only after design closure.

---

# 13. Implementation gate

Implementation starts only when all are true:

- [ ] whole-product domain map closed;
- [x] Organization/AuthZ north star closed;
- [x] Approval V1 closed;
- [x] R3 Controlled Information configuration closed;
- [x] R4 Document/Revision/Submission lifecycle closed;
- [x] R5 Numbering/TemplateSpec/metadata boundaries closed;
- [x] R6 Periodic Review/Rendition/Release closed;
- [ ] R7 distribution/tokens/audit/notifications/search closed;
- [ ] R8 tenant lifecycle/security closed;
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

# 14. Exact next step

Continue **R7 — Distribution + Read/Acknowledgement + Tokens/Computed Values + Audit/Evidence + Notifications + Search**.

Do not implement. For every supporting concern answer:

1. which canonical authority emits/owns the source fact;
2. what is immutable evidence vs rebuildable projection;
3. what exact Document/REV/Submission/Release identity it refers to;
4. what must be snapshotted to preserve historical truth;
5. what can remain derived/eventually consistent;
6. whether any new engine/framework is actually justified.