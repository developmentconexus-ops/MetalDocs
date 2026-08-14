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

# 1. Why the redesign is whole-platform

The redesign began with authorization drift, expanded into Approval, and then proved that `documents`, `controlleddocuments`, `templates`, taxonomy, IAM, Approval, rendering and release contain overlapping authorities.

The strongest product counterexample came from browser QA: the user edited and the approver reviewed one content revision, while freeze rendered the blank template snapshot. The official PDF was blank and its signed hash did not represent the content reviewed.

**Root cause:** MetalDocs evolved as locally reasonable modules/features instead of one coherent controlled-information model.

**Target property:** every business fact has one authority; supporting concerns consume it rather than reinterpret or mutate it independently.

Existing code/schema/API remain evidence of requirements and migration impact. Their existence does **not** grant target-design legitimacy.

---

# 2. Whole-product responsibility census

| Current concern/module | Target disposition |
|---|---|
| `auth` | retain V1 authentication/session implementation behind a stable AuthN boundary |
| `iam` | conceptually split into **Organization** + **Authorization**; exact package layout later |
| `approval` | redesign as small specialized Approval V1; never owner of document release/effectivity |
| `documents` | becomes the core of **Controlled Information** after responsibility cleanup |
| `controlleddocuments` | retire as target context; stable identity/numbering responsibilities move to Document/configuration |
| `templates` | retire parallel lifecycle; template becomes a role of a governed Document/DocumentRevision |
| `taxonomy` | dismantle: Area → Organization; Profile → DocumentType; Family → classification-only category; GovernanceClass deleted |
| `render` | supporting rendition/rendering infrastructure bound to exact RevisionSubmission truth |
| `audit` | retain as distinct evidence/integrity authority; exact transaction seam later |
| `distribution` | supporting released-revision concern; read/ack semantics still open |
| `notifications` | supporting event consumer, never workflow authority |
| `search` | rebuildable read model/projection consumer |
| `tokens` | supporting configuration/value provider; snapshot semantics still open |
| `security` | retain/re-evaluate after tenant lifecycle/AuthN seam closes |
| `jobs` | infrastructure/orchestration, not a business bounded context |

Responsibilities that must be closed before code even if current module names disappear: numbering, revision lifecycle, immutable submission evidence, template payload, periodic review, rendition provenance, release/effectivity, distribution/read/ack, tokens/computed values, audit, notifications/search, tenant lifecycle/security, final Permission catalog, APIs and frontend journeys.

---

# 3. LOCKED — Authentication boundary

## AUTHN-01 — AuthN, AuthZ, Approval and Domain Governance are separate authorities

- Authentication: **who is this actor/session?**
- Authorization: **what may this principal do in this tenant/scope?**
- Approval: **who participates in this concrete submission?**
- Domain Governance: **is this action legal now given lifecycle, SoD, reauth and immutable-content rules?**

None substitutes for another.

## AUTHN-02 — No Keycloak now

Current MetalDocs AuthN is sufficient for V1: credential verification, opaque/server-side sessions, revocation/lockout and tenant-bound identity context.

Future external IdP seam is preserved. Revisit Keycloak/OIDC/SAML when enterprise SSO/federation, broad MFA/passkeys, tenant-specific IdPs or deliberate credential externalization becomes a real requirement.

Other modules depend on a stable authenticated-principal abstraction, not on password-session implementation details.

---

# 4. LOCKED — Organization + Authorization V1

## ORG-01 — Organization

```text
Tenant
Area
User
Group
GroupMembership
```

Area is organizational truth used consistently by Document ownership, RoleAssignment scope and Approval actor resolution.

## ORG-02 — Groups are flat first-class principals

- User may belong to multiple Groups.
- No nested Groups V1.
- Groups receive ordinary RoleAssignments, never raw permissions.
- Group is not forced to belong structurally to one Area.

Example:

```text
Group Vendedores
  members: Beatriz, João, Jordana
  assignment: author @ Area(COMERCIAL)
```

## AUTHZ-01 — Five built-in roles

```text
tenant_owner
area_manager
author
approver
viewer
```

Historical `system_admin`, `editor`, `signer`, `qms_admin`, `area_admin` are not target roles.

Roles are bundles; runtime checks semantic Permissions, never `if role == ...` branches.

## AUTHZ-02 — One grant shape

```text
RoleAssignment
  subject: User | Group
  role
  scope: Tenant | Area
  granted_by / granted_at
  revoked_by? / revoked_at? / reason?
```

Legal built-in scopes:

| Role | Tenant | Area |
|---|---:|---:|
| tenant_owner | yes | no |
| area_manager | no | yes |
| author | yes | yes |
| approver | yes | yes |
| viewer | yes | yes |

Effective access = direct User assignments + Group assignments. Additive grants + default deny. No explicit deny engine and no temporal grant scheduler in V1.

## AUTHZ-03 — Role semantics

- **viewer:** released/effective official information in scope; no broad working-content access.
- **author:** published + eligible working content; create/edit/comment/submit within scope. `created_by` is evidence, not authorization ownership.
- **approver:** qualification only; not blanket draft visibility. Concrete Approval participation grants case access.
- **area_manager:** operational management inside assigned Area: author-like work plus workflow oversight/cancel/reassign and eligible retirement operations. Not IAM/RBAC administrator.
- **tenant_owner:** tenant product administrator through ordinary Permission grants; never an AuthZ or Domain Governance bypass.

## AUTHZ-04 — Candidate semantic permissions

Current candidate set (not final until whole product closes):

```text
document.read_published
document.read_working
document.create
document.edit
document.comment
document.submit
document.obsolete

approval.act
approval.oversee
approval.cancel
approval.reassign
approval_policy.manage

organization/access-administration permissions — later
```

`document.supersede` is no longer presumed necessary: ordinary Revision supersession is a mechanical release consequence, not a separate human authority. Cross-document replacement will be evaluated separately.

## AUTHZ-05 — Decision composition

```text
current permission/qualification
+ workflow participation where applicable
+ Domain Governance constraints
= ALLOW
```

No tenant owner bypass.

## AUTHZ-06 — No OpenFGA/SpiceDB now

Scoped RBAC + flat Groups + natural domain relationships (such as Step participants) solve the demonstrated V1. Revisit an external ReBAC engine only if arbitrary per-resource sharing/relationship graphs become material.

---

# 5. LOCKED — Approval V1

## APPR-01 — Specialized document approval, not BPM

No BPMN, generic gateways, arbitrary service tasks, Camunda/Flowable, CEL, parallel workflow graph or generic process engine in V1.

## APPR-02 — Versioned sequential ApprovalPolicy

```text
ApprovalPolicy
  id
  version
  ordered ApprovalStep[]
```

A changed active policy creates a new policy version for future submissions. In-flight/historical ApprovalInstances remain pinned to the version used at creation.

`ApprovalPolicy.version` is its own namespace; it is not a document `REVxxx`.

## APPR-03 — ApprovalStep is the human task

```text
order
name
purpose: review | approval
actor_rule
completion: ANY | ALL
requires_reauthentication
due_in_days?
```

`review`/`approval` describe business meaning and UI/evidence, not separate engines.

## APPR-04 — Participant rules

V1:

```text
NamedUser
Group
RoleInArea(role, fixed-area | subject-area)
```

No arbitrary submitter-selected actors unless a real requirement later proves it.

## APPR-05 — Completion

Only `ANY` and `ALL`. No M-of-N/weighted voting V1.

Participants resolve when a Step activates and are snapshotted as evidence. Current qualification is checked again at action time.

If an `ALL` participant becomes unavailable, do not shrink silently and do not call it rejection. Step requires attention; authorized manager performs audited reassignment.

## APPR-06 — Human outcomes

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

No normal terminal human `reject` concept V1.

## APPR-07 — Attempts are content-exact

ApprovalInstance binds one immutable `RevisionSubmission`.

`return_for_changes` terminates that approval attempt and returns the same DocumentRevision to DRAFT. Editing and resubmission produce a **new RevisionSubmission and new ApprovalInstance**. Old decisions never approve new bytes.

## APPR-08 — Deadlines, reassignment, reauth

- `due_in_days` → derived `due_at`; overdue surfacing/notification only.
- audited explicit reassignment covers absence/error/termination.
- no generic time-window delegation engine V1.
- a Step may require fresh reauthentication before `accept`.

## APPR-09 — Evidence

Every decision remains explainable against actor, timestamp, policy/version, Step, Submission, Document/REVxxx, content digest, outcome, comment/reason and reauth evidence when required.

---

# 6. LOCKED — Controlled Information configuration (R3)

## CI-01 — One Controlled Information context

The target split `documents + controlleddocuments + templates` is rejected.

Core semantic model:

```text
Document
DocumentRevision
```

Exact Go/package boundary is deferred until domain closure.

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

Rules:

- tenant-scoped;
- no independent versioning V1;
- inactive means unavailable for new Documents, not invalidation of existing Documents;
- a Document's type is immutable after creation V1;
- Authorization owns edit authority; `editable_by_role` is deleted;
- DocumentType references/configures domain-specific policies rather than reimplementing them.

## CI-03 — DocumentTypeCategory replaces behavioral Family

Optional tenant-scoped classification/navigation grouping only.

No inherited Approval, numbering, metadata or permissions; no `Type → Subtype → Classification` hierarchy V1.

## CI-04 — GovernanceClass deleted

`controlado/simples/livre` currently duplicates Approval shape and risks becoming a cross-domain god enum.

Each concern owns explicit configuration when it exists.

Approval is explicitly:

```text
ApprovalConfiguration =
    NoHumanApproval
  | UsePolicy(ApprovalPolicyID)
```

This avoids ambiguous `NULL` and removes fake zero-stage routes.

## CI-05 — Template is a role of a governed Document

A template has no parallel lifecycle/version counter. A template Document uses the same `DocumentRevision` lifecycle and official `REVxxx` labels as every other governed Document.

Changing DOCX/layout/placeholders/schema/constraints/visibility/resolver semantics means a new DocumentRevision.

## CI-06 — TemplateUse is M:N

```text
TemplateUse
  template_document_id
  target_document_type_id
  is_default
```

- one template Document may serve multiple DocumentTypes;
- one DocumentType may offer multiple templates;
- at most one default template per type;
- default is UX convenience only;
- no `template_required` V1 until a real requirement exists;
- blank creation remains valid.

At creation MetalDocs resolves the template Document's **current EFFECTIVE DocumentRevision**, then permanently pins the exact source Document + source REV + source digest on the created Document's origin. Newer template revisions affect only future creations.

---

# 7. LOCKED — Document + DocumentRevision lifecycle (R4)

## REV-01 — Document is stable identity

Document owns stable identity facts such as tenant, business code, DocumentType, Area, origin/provenance and pointers to current effective/open revisions.

`Document.status = draft/under_review/approved/published...` is rejected as the principal lifecycle model.

## REV-02 — Official revision labels

Human/audit/official revision identity:

```text
REV001
REV002
REV003
...
```

`REVxxx` is the business revision. Technical IDs, row versions, schema versions and policy versions are separate namespaces and never replace it in UI/evidence.

## REV-03 — DocumentRevision is a business change cycle, not autosave

Autosaves/checkpoints/edit snapshots belong to authoring infrastructure inside the open Revision. They do not increment official REV numbers.

One Document may have:

```text
effective_revision = REV001
open_revision      = REV002
```

At most one open Revision per Document V1; no editorial branching.

## REV-04 — Revision states

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

No `APPROVED`, `SCHEDULED` or `PUBLISHED` state on Revision: those truths belong to Approval/Release and can be projected in UI.

Typical flow:

```text
DRAFT
  ↓ submit
SUBMITTED
  ↓ release
EFFECTIVE
  ↓ newer REV releases
SUPERSEDED
```

`return_for_changes` / allowed withdrawal: `SUBMITTED → DRAFT` on the **same REV**.

A candidate may become `CANCELLED`. An effective Document may become `OBSOLETE` when retired without a successor.

## REV-05 — REV allocation

The next `REVxxx` is allocated when a new change cycle is created. It is never reused, even if that Revision is later cancelled.

`REV002+` requires `reason_for_change` before first submission. While DRAFT the reason may change; each immutable Submission snapshots the reason that accompanied those exact bytes.

## REV-06 — RevisionSubmission is first-class immutable evidence

```text
DocumentRevision REV002
  ├── Submission #1 → hash/digest A → returned
  └── Submission #2 → hash/digest B → accepted/released
```

RevisionSubmission exists even for `NoHumanApproval` because release still needs an exact immutable candidate identity.

Approval, Rendition and Release bind the **Submission**, not a mutable Revision or Document chosen independently.

## REV-07 — Editing/approval boundary

- DRAFT is editable.
- SUBMITTED is frozen from content/governed-metadata mutation.
- return/withdraw creates no new REV; it returns the same REV to DRAFT and closes the old Submission attempt.
- after final Approval completion, V1 does not reopen that candidate for editing. If emission must be stopped, cancel the candidate and start a new REV.

## REV-08 — Working history

Autosaves/checkpoints/editor sessions are technical authoring history tied to the open Revision. An old EditorSession never authorizes mutation of a SUBMITTED Revision.

Collaboration comments may attach to the Revision/authoring context. Approval decision comments belong to Approval evidence, not a generic document-comments table.

## REV-09 — Effectivity and retirement semantics

When REV002 becomes EFFECTIVE, the same atomic release transition makes prior REV001 SUPERSEDED and changes the Document effective pointer.

`SUPERSEDED` = replaced by a newer Revision of the same Document.

`OBSOLETE` = the Document is explicitly retired without a successor Revision. Obsolete is terminal V1; no reactivation until a proven requirement appears.

Cross-Document replacement, if required, is a separate future relation and must not be conflated with normal revision supersession.

---

# 8. LOCKED — Numbering + Template payload + metadata boundaries (R5)

## NUM-01 — Document code is immutable tenant-wide business identity

`Document.code` is unique per tenant and identifies the stable Document, not one Revision.

Example:

```text
IT-LOG-01
├── REV001
├── REV002
└── REV003
```

A Document's `type_id`, `area_id` and `code` are immutable V1. `DocumentType.code` and `Area.code` are also stable codes; names may change without changing identity.

## NUM-02 — Code allocated at Document creation

Code allocation and creation of the Document + initial `REV001` are one atomic operation.

A successfully created Document permanently consumes its sequence number, even if later cancelled/retired. A transaction that rolls back before Document creation need not consume it.

Preview is advisory only; Create response is the authoritative assigned code.

## NUM-03 — Deliberately small numbering configuration

Numbering is a DocumentType configuration/value object, not a generic rules engine.

V1 format supports literals plus only:

```text
{TYPE}
{AREA}
{SEQ}
```

V1 sequence scopes:

```text
TYPE
TYPE_AREA
```

`sequence_width` is minimum zero-padding, not a numeric limit.

Examples:

```text
{TYPE}-{AREA}-{SEQ}, width=2 → IT-LOG-01
DOC-{TYPE}-{AREA}-{SEQ}, width=3 → DOC-IT-LOG-001
```

No year/month/user/custom-field tokens, formulas/scripts, resets or expression language V1. Add only when a concrete tenant requirement proves need.

## NUM-04 — Manual codes are not normal authoring

Normal Create does not expose arbitrary manual-code override.

Preserving pre-existing codes is a separate import/migration/bootstrap authority with uniqueness, provenance and audit requirements. Do not contaminate ordinary creation with migration semantics.

## TPL-01 — Delete template MetadataSchema as policy authority

Template does **not** own:

```text
DocCodePattern
RetentionDays
DistributionDefault
RequiredMetadata generic framework
```

Numbering belongs to DocumentType. Retention/distribution get their own authorities only if those requirements are approved. Generic custom metadata is not hidden inside TemplateSpec.

## TPL-02 — Template Revision = ordinary Revision content + TemplateSpec

A Revision used as a template contains its ordinary governed content (e.g. DOCX/source content) plus a `TemplateSpec` describing the filling/authoring contract.

No duplicate TemplateVersion lifecycle, no `CompositionJSON` V1 without an independent requirement.

## TPL-03 — TemplateField separates data type from value source

Conceptual field:

```text
TemplateField
  key
  label
  value_type: text | date | number | choice | user | image
  source:
      user_input
    | system(key)
    | dictionary(key)
  constraints...
  visible_if?
```

`computed` and `dictionary` are value **sources**, not value types.

Typed constraints may include required/default, choices, regex/max-length, numeric/date ranges. Invalid cross-type combinations are rejected.

A limited typed `visible_if` survives using closed operators such as `eq/ne/gt/gte/lt/lte`; no CEL/JavaScript/generic expression language.

## TPL-04 — DOCX/layout and TemplateSpec must agree

Before a template Revision can be submitted, content anchors/tokens and TemplateSpec must pass parity validation. Missing/extra field definitions fail closed unless a future metadata-only field concept is explicitly designed.

## TPL-05 — Template seeds; derived Revision becomes independent truth

Creating from template copies/pins the exact effective source Revision content + applicable TemplateSpec into the new Document's initial Revision and records immutable `DocumentOrigin` provenance:

```text
Blank
or
Template {
  source_template_document_id
  source_template_revision_id
  source_template_revision_label   // REVxxx
  source_digest
}
```

The derived Document no longer consults the template to know what its content is. New template Revisions never silently rebind it.

A derived Document does not redefine inherited TemplateSpec in V1. Adopting a newer template schema, if later required, will be an explicit migration/rebase operation.

A possible future “newer template available” indicator is a derived advisory/read-model fact only, never automatic state/rebinding.

## TPL-06 — RevisionContent is one logical governed content identity

For structured templates, content truth may include more than raw DOCX bytes.

Conceptually:

```text
RevisionContent
  source/content artifact ref + hash
  authoring schema snapshot
  field values
  governed Revision metadata
```

The immutable `RevisionSubmission.submission_digest` covers the complete governed content identity needed to reproduce/attest what was submitted. Raw DOCX hash or form-data hash alone is insufficient when both contribute to rendered meaning.

## META-01 — Stable vs governed vs operational metadata

- **Document:** stable identity and operational facts that should not inherently create a new REV (e.g. code, type, area, origin; potentially responsible owner).
- **DocumentRevision:** governed facts whose change is part of the controlled content lifecycle (including title and reason-for-change).
- **RevisionSubmission:** immutable snapshot of the exact governed and decision-relevant facts for that attempt.

`title` belongs to DocumentRevision so historical REV titles remain truthful.

A `responsible_owner` may exist as audited operational metadata on Document and **does not grant authorization**. Exact semantics will be revisited with periodic review/ownership responsibilities.

## META-02 — No generic tenant custom-metadata engine V1

Do not create `CustomFieldDefinition`, arbitrary type inheritance, metadata expressions or metadata-specific permission machinery until a real product requirement appears.

If tenant-defined metadata later becomes necessary, `DocumentType` is the likely authority for field definitions; TemplateSpec remains authoring-layout contract, not a second metadata platform.

---

# 9. LOCKED NORTH STAR — Release and official truth

Human Approval never directly effectivates a Revision.

Release evaluates independent prerequisites such as:

- immutable RevisionSubmission exists;
- Approval requirement satisfied (`NoHumanApproval` or completed Approval evidence);
- same Submission digest remains the candidate;
- required official Renditions/artifacts exist and attest that Submission;
- effective date is reached;
- supersession/effectivity invariants hold.

The single winning release transaction sets the candidate Revision EFFECTIVE, updates `Document.effective_revision_id`, makes the prior effective Revision SUPERSEDED when present, and closes the open-revision pointer.

The previous effective Revision remains official until that transaction wins. Never two effective Revisions for one Document.

The Release Coordinator concept survives because this asynchronous/effectivity boundary is real. Its final package/context placement is open.

---

# 10. Build-vs-buy rulings to date

| Technology/class | V1 ruling | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | do not deploy now | enterprise SSO/federation/MFA/passkeys or credential externalization |
| OpenFGA / SpiceDB | do not deploy now | arbitrary resource sharing / large relationship graph / service split |
| Camunda / Flowable / BPMN | do not use for document Approval V1 | product genuinely becomes generic business-process engine |
| Temporal as Approval engine | do not use | separate durable orchestration requirement current outbox/River cannot economically serve |
| CEL / expression language | do not use now | real conditional product/workflow policies cannot be represented cleanly by typed configuration |

Libraries/frameworks are selected only after the exact domain responsibility is closed.

---

# 11. Explicit target deletions/replacements

No entitlement to survive:

- target split `documents` / `controlleddocuments` / `templates`;
- separate public/domain `ControlledDocument` identity;
- `DocumentProfile`;
- behavioral `DocumentFamily` hierarchy;
- `GovernanceClass {controlado, simples, livre}` and route-shape derivation;
- parallel TemplateVersion lifecycle/version counter;
- template `MetadataSchema` as numbering/retention/distribution authority;
- duplicate `DocCodePattern` in templates;
- normal-create manual code override;
- user-visible document `v7`-style revision labels;
- autosave rows presented as official DocumentRevisions;
- Document lifecycle carrying `draft/under_review/approved/published` as the revision truth;
- `StageKind` as two workflow engines;
- configurable approval-stage required capability;
- drift policies that silently change completion requirements;
- M-of-N and generic delegation as V1 requirements;
- normal terminal human reject semantics;
- reuse of a single ApprovalInstance across edited content;
- generic BPMN/CEL/phase/branching workflow machinery;
- role-based authorization bypasses;
- magic tenant-scope area sentinel;
- multiple semantic grant sources;
- `editable_by_role`;
- Approval code mutating Controlled Information tables directly;
- any freeze/render path capable of choosing bytes other than the exact submitted content identity;
- old roadmap/milestone/spec documents as live authority.

Prefer deletion over compatibility shims when no deployed/contractual compatibility requirement proves a shim necessary.

---

# 12. Remaining design queue before implementation

## R6 — Periodic Review + Renditions/Rendering + Release/Effectivity — **NEXT**

Close together because they determine the complete official-information path:

1. PeriodicReviewPolicy ownership/configuration.
2. due-date derivation and whether due/overdue changes legal effectivity.
3. “reviewed, no change” evidence vs “change required → new REV”.
4. who performs periodic review and how AuthZ/Approval compose.
5. `Rendition` exact responsibility and statuses.
6. source DOCX vs final DOCX vs official PDF semantics.
7. immutable provenance tuple tying a Rendition to RevisionSubmission/digest.
8. renderer version/configuration evidence needed for attestation/reconstruction.
9. failure/retry/idempotency expectations.
10. what artifacts are mandatory before Release.
11. planned effective date semantics.
12. atomic effectivity/supersession transaction.
13. cancellation after Approval but before Release.
14. cross-Document replacement, if the product genuinely needs it.

## R7 — Distribution / read / acknowledgement + Tokens + Audit + Notifications + Search

- obligation denominator snapshot vs live derivation;
- read vs acknowledgement evidence;
- deadlines/reminders/export;
- token/dictionary pinning timing;
- computed resolver catalogue/version/provenance;
- global audit/evidence seam and hash-chain/export behavior;
- domain event catalogue;
- notifications as consumer;
- search projection visibility/rebuildability/eventual consistency.

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
- workflow relation checks;
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
- this ledger is the only active detailed decision source under `docs/superpowers`.
- `wiki/references/current-agent-handoff.md` is a short recovery pointer.
- Git history is the archive for deleted historical staging documents.
- legacy wiki/module/ADR material may explain current runtime/history but cannot override operator-approved target decisions here.
- final integrated decisions will be promoted to durable wiki/ADR/spec authorities only after design closure.

---

# 14. Implementation gate

Implementation starts only when all are true:

- [ ] whole-product domain map closed;
- [x] Organization/AuthZ north star closed;
- [x] Approval V1 closed;
- [x] R3 Controlled Information configuration closed;
- [x] R4 Document/Revision/Submission lifecycle closed;
- [x] R5 Numbering/TemplateSpec/metadata boundaries closed;
- [ ] R6 periodic review/rendition/release closed;
- [ ] R7 distribution/tokens/audit/notifications/search closed;
- [ ] R8 tenant lifecycle/security closed;
- [ ] final Permission + Role matrix closed;
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

Continue **R6 — Periodic Review + Renditions/Rendering + Release/Effectivity**.

Do not implement. For every proposed state/fact answer:

1. which authority owns it;
2. whether it is durable business truth, immutable evidence or derived projection;
3. what exact RevisionSubmission/digest it refers to;
4. whether an external library/service is actually necessary;
5. what simpler model would fail the business requirement.
