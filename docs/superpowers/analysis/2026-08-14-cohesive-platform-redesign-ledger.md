# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding for this redesign; open items are explicitly marked.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Governing method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Canonical program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED. No product code, schema, API, frontend or migration implementation is authorized yet.**

---

## 0. Fresh-session contract

A new session working on this program reads, in order:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. this ledger
5. `wiki/references/current-agent-handoff.md`

Do not resume an old roadmap, milestone, spec, deleted `docs/superpowers` artifact, PR #113, or historical ADR execution sequence by inertia.

The method is deliberately design-first:

```text
product/domain semantics
→ invariants + lifecycle
→ organization/authz/workflow integration
→ build-vs-buy
→ bounded contexts / ownership
→ data model
→ API/contracts/UX
→ migration/delete/rename map
→ implementation specification
→ implementation plan
→ code
```

Runtime/schema/OpenAPI answer **what exists today**. This ledger answers **what the target is becoming**. Existing implementation shape is evidence, not architectural entitlement.

---

# 1. Why the redesign expanded to the whole platform

The work began with authorization having multiple grant authorities. Approval then could not be finalized without understanding workflow semantics. That exposed duplicated authority across `documents`, `controlleddocuments`, `templates`, taxonomy, approval, rendering and release.

A real browser QA run proved the content model itself was contradictory: the approver reviewed editor-authored content while freeze rendered the blank template snapshot; the final signed PDF was blank and the signed hash did not represent what the human reviewed.

**Root cause:** locally reasonable modules/features accumulated without one coherent controlled-information model.

**Target property:** every business fact has one authority; supporting concerns consume that authority instead of reinterpreting or mutating it independently.

**Decision:** restructure at the semantic/design level first. Delete accidental complexity instead of preserving it with compatibility kernels.

---

# 2. Whole-product responsibility census

The current repository has 15 `internal/modules` directories. Their names are not target commitments; their legitimate responsibilities are:

| Current module | Legitimate responsibility | Target direction |
|---|---|---|
| `auth` | credential + session authentication | retain for V1 behind AuthN boundary |
| `iam` | people/groups/grants + tenant administration | split semantically into Organization + Authorization |
| `approval` | human policy, participation, decisions, evidence | redesign as small Approval V1 |
| `documents` | authoring/revision/comments/content/release-facing operations | becomes Controlled Information core after cleanup |
| `controlleddocuments` | stable business identity/code/type/area/sequence | retire as separate context; responsibilities move to Document + Numbering |
| `templates` | template content/schema/placeholders + parallel lifecycle | retire parallel lifecycle; template becomes a role of governed Document |
| `taxonomy` | areas/profile/family/governance class | dismantle: Area → Organization; Profile → DocumentType; Family → category; GovernanceClass deleted |
| `render` | materialization/PDF/rendition infrastructure | retain, aligned to exact Submission/Revision truth |
| `audit` | append-only/integrity evidence | retain as distinct evidence authority |
| `distribution` | released-revision coverage/read/ack domain | re-evaluate later |
| `notifications` | notification inbox/fanout | supporting event consumer only |
| `search` | read projection/query | supporting rebuildable read model |
| `tokens` | tenant dictionary/value provider | retain; snapshot timing still open |
| `security` | security signals/MFA/tenant-key concerns | re-evaluate after tenant/AuthN closure |
| `jobs` | River/background orchestration | infrastructure/composition, not business bounded context |

Still-required product concerns include periodic review, numbering, immutable submission evidence, rendition provenance, release/effectivity/supersession, distribution/read/acknowledgement, audit, notifications/search, tenant lifecycle/security, APIs and frontend journeys.

---

# 3. LOCKED — Authentication V1

## AUTHN-01 — AuthN, AuthZ, Approval and Domain Governance are distinct

```text
Authentication      → who is the actor/session?
Authorization       → what broad authority exists at tenant/area scope?
Approval            → who participates in this concrete human-work item?
Domain Governance   → is this action legal now?
```

No layer substitutes for another.

## AUTHN-02 — Do not deploy Keycloak now

Current MetalDocs AuthN is sufficient for V1: credentials, opaque/server-side sessions, revocation/lockout and tenant-bound identity context.

Keep an authentication seam so a future OIDC/Keycloak/external-IdP implementation can replace it when a concrete enterprise trigger appears: SAML/OIDC federation, tenant-specific Entra/Okta/Google SSO, broad MFA/passkeys, or a deliberate decision to stop operating credentials.

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

Area belongs to Organization, not document taxonomy. The same Area is referenced by Documents, RoleAssignments and Approval actor resolution.

## ORG-02 — Groups

Groups are first-class flat principals.

- User may belong to multiple Groups.
- No nested Groups in V1.
- Group receives RoleAssignments, not raw Permissions.
- Group does not have to structurally belong to one Area.

Example:

```text
Group Vendedores
  members: Beatriz, João, Jordana
  assignment: author @ Area(COMERCIAL)
```

## AUTHZ-01 — Exactly five built-in roles in V1

```text
tenant_owner
area_manager
author
approver
viewer
```

Historical `system_admin`, `editor`, `signer`, `qms_admin`, `area_admin` do not survive unless a later requirement proves an independent responsibility.

## AUTHZ-02 — One grant shape

```text
RoleAssignment
  subject: User | Group
  role
  scope: Tenant | Area
  granted_by / granted_at
  revoked_by? / revoked_at? / reason?
```

V1 does not require scheduled `valid_from/valid_until` grants. Default deny. Direct + Group grants compose additively. No explicit deny-policy system in V1.

Legal scopes:

| Role | Tenant | Area |
|---|---:|---:|
| tenant_owner | yes | no |
| area_manager | no | yes |
| author | yes | yes |
| approver | yes | yes |
| viewer | yes | yes |

## AUTHZ-03 — Role semantics

- `viewer`: effective/published official information in scope.
- `author`: published + working content in scope; create/edit/comment/submit eligible work. `created_by` is evidence, not authorization ownership.
- `approver`: qualification, not blanket case authority or broad draft visibility. Concrete Step participation opens only the case needed.
- `area_manager`: operational manager in assigned Area; Author-like work plus workflow oversight/cancel/reassign and eligible retirement operations. Not RBAC administration.
- `tenant_owner`: tenant product administrator through ordinary Role→Permission mappings. Never a bypass.

## AUTHZ-04 — Runtime checks Permissions, not role names

Candidate permissions remain provisional until the whole product is closed. Current semantic direction:

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
```

`document.supersede` is now reopened: ordinary REV-to-REV supersession is a mechanical release consequence, not necessarily an independent human authority. Cross-Document replacement is still to be designed.

## AUTHZ-05 — Human approval action composition

```text
base qualification/Permission at scope
+
workflow participation in StepInstance
+
Domain Governance constraints
=
ALLOW
```

No role, including tenant owner, bypasses SoD, immutable content, state legality, reauthentication or tenant isolation.

## AUTHZ-06 — Do not deploy OpenFGA/SpiceDB now

Scoped RBAC + flat Groups + natural domain relations such as ApprovalStepParticipant are sufficient for V1. Keep an engine-neutral Authorizer boundary (`Check` + collection/scoping support) for a future relationship engine only if arbitrary sharing/large relation graphs/service split prove the need.

---

# 5. LOCKED — Approval V1

## APPR-01 — Specialized approval, not BPM

No BPMN, gateways, loops, subprocesses, generic service tasks, Camunda/Flowable runtime, CEL rule language or generic parallel phase graph in V1.

## APPR-02 — Versioned sequential ApprovalPolicy

```text
ApprovalPolicy
  id
  version
  ordered ApprovalStep[]
```

New policy versions apply to new submissions; historical/in-flight ApprovalInstances remain pinned to their policy version.

`ApprovalPolicy.version` is a policy namespace, never a document `REVxxx` label.

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

`review` and `approval` are business meaning/UI/evidence, not separate engines.

## APPR-04 — Participant rules

V1:

```text
NamedUser
Group
RoleInArea(role, fixed-area | subject-area)
```

No submit-time arbitrary candidate choice until a real workflow proves it.

## APPR-05 — Completion

Only `ANY` and `ALL`. No M-of-N/weighted/percentage rules in V1.

## APPR-06 — Participant resolution

Resolve when a Step activates and snapshot selected participants. Snapshot is evidence, not permanent authorization. Current qualification is checked again when acting.

If an `ALL` participant becomes unavailable, do not silently reduce the denominator and do not call it a human rejection. Surface attention/blocked state and require audited reassignment.

Historical drift policies are deleted.

## APPR-07 — Human outcomes

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

No normal terminal human `reject` in V1.

## APPR-08 — Attempts are immutable

ApprovalInstance binds to one immutable `RevisionSubmission`. `return_for_changes` terminates that ApprovalInstance. Resubmission creates a new ApprovalInstance over a new RevisionSubmission; old human decisions never authorize changed bytes.

## APPR-09 — Deadline/reassignment simplicity

`due_in_days → due_at`, overdue surfacing/notification, and audited reassignment are enough for V1. Generic escalation chains and time-window delegation are deferred.

## APPR-10 — Reauthentication

Optional per Step. Current AuthN can satisfy it; future IdP/MFA can plug into the same seam.

## APPR-11 — Decision evidence

Must explain at least:

```text
actor / on_behalf_of if ever applicable
when
ApprovalPolicy + version
StepInstance + purpose
RevisionSubmission
DocumentRevision REVxxx
content/submission digest
outcome
comment/reason
reauth evidence when required
```

---

# 6. LOCKED — Controlled Information configuration (R3)

## CI-01 — One Controlled Information domain

Target core nouns:

```text
Document
DocumentRevision
RevisionSubmission
```

`documents + controlleddocuments + templates` do not survive as three target business contexts.

## CI-02 — Document is stable identity

Document answers “what controlled information is this?” and owns stable identity such as tenant, immutable business code, DocumentType, Area and pointers to its official/open revision state.

The separate public/domain noun `ControlledDocument` is retired.

## CI-03 — Official revision vocabulary is `REVxxx`

Human-facing, audit-facing and official revision identity uses:

```text
REV001
REV002
REV003
...
```

Never expose technical forms such as `v7` as the document revision.

Separate namespaces may exist internally:

```text
DocumentRevision.id        → technical immutable ID
DocumentRevision.number    → 1,2,3... internal numeric value
DocumentRevision.label     → REV001, REV002... official label
row_version / lock_version → OCC only
schema_version             → technical compatibility only
ApprovalPolicy.version     → approval-definition version only
```

## CI-04 — DocumentProfile → DocumentType

`DocumentProfile` is replaced by tenant-scoped `DocumentType`.

```text
DocumentType
  id
  tenant_id
  code          // immutable
  name
  description?
  category_id?
  status        // ACTIVE | INACTIVE
```

Rules:

- no own versioning in V1;
- ACTIVE/INACTIVE only;
- inactive types cannot create new Documents but existing Documents remain valid;
- Document type is immutable after Document creation in V1;
- type references/configures policies owned by their own authorities rather than implementing them;
- historical `editable_by_role` is deleted.

## CI-05 — Family becomes category only

`DocumentFamily` becomes optional `DocumentTypeCategory` for classification/navigation/filter/reporting only.

No policy inheritance, deep type hierarchy, independent permissions or behavior.

## CI-06 — GovernanceClass is deleted

`controlado/simples/livre` is rejected as a cross-domain authority. Each real concern expresses its own configuration only when that concern exists.

Approval configuration is explicit:

```text
ApprovalConfiguration =
    NoHumanApproval
  | UsePolicy(ApprovalPolicyID)
```

This prevents ambiguous `NULL` and replaces fake zero-stage routes.

## CI-07 — Template is a role of a governed Document

A template has no parallel TemplateVersion lifecycle. A Document designated for template use changes through ordinary `DocumentRevision REVxxx` lifecycle.

Changing layout, placeholder schema/type/default/options/validation/visibility/resolver semantics or other governed template rules means a new REV.

## CI-08 — TemplateUse is M:N

```text
TemplateUse
  template_document_id
  target_document_type_id
  is_default
```

Rules:

- one template Document may serve multiple DocumentTypes;
- one DocumentType may offer multiple template Documents;
- at most one default per DocumentType;
- default is UX/preselection, not mandatory governance;
- no `template_required` in V1 until a concrete requirement proves it;
- blank creation remains allowed.

When creating a Document, resolve the template Document's **current EFFECTIVE DocumentRevision** and pin immutable source provenance. Newer template revisions affect future creations only; existing Documents never rebind automatically.

---

# 7. LOCKED — Document + Revision + Submission lifecycle (R4)

## R4-01 — Document has stable identity, not workflow status

`Document` does not own `draft / under_review / approved / scheduled / published` as its primary lifecycle. Those labels historically conflated revision state, approval state and release state.

Conceptually Document owns stable identity plus pointers such as:

```text
Document
  code
  type_id
  area_id
  title
  effective_revision_id?
  open_revision_id?
  retired_at?
```

Exact fields wait for data-model design.

## R4-02 — One effective revision + at most one open revision

Normal change state:

```text
Document IT-LOG-001
  effective_revision → REV001
  open_revision      → REV002
```

V1 forbids parallel open revision branches. `REV002 DRAFT` + `REV003 DRAFT`, or `REV002 SUBMITTED` + `REV003 DRAFT`, are invalid.

## R4-03 — DocumentRevision is the business change cycle

`DocumentRevision` means `REV001`, `REV002`, ... — not each autosave or check-in.

A new REV is allocated when a new controlled change cycle begins (for example “Create new revision” on an effective Document), not on every save, submission attempt or approval correction.

Revision labels are monotonic and never reused. If REV002 is cancelled, the next revision is REV003.

## R4-04 — Revision states

Target V1 states:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

No persisted Revision states `APPROVED`, `SCHEDULED` or `PUBLISHED` are required.

Why:

- Approval owns whether human approval is complete.
- Release owns planned/effective-date/readiness state.
- Revision owns whether content is mutable, submitted candidate, official, replaced, retired or cancelled.

The UI may derive richer labels such as “Em aprovação” or “Aprovado · vigência em 01/09/2026” from Revision + Approval + Release projections without duplicating authorities.

## R4-05 — Draft/submitted mutability

`DRAFT` is mutable through authoring infrastructure.

Submitting transitions the REV to `SUBMITTED` and freezes that attempt's immutable bytes/metadata identity. While `SUBMITTED`, editing is illegal.

`return_for_changes` or valid `withdraw` closes the current attempt and returns the **same REV** to `DRAFT`.

## R4-06 — RevisionSubmission is first-class and immutable

A REV may have multiple immutable submission attempts:

```text
REV002
  Submission 1 → digest AAA → returned_for_changes
  Submission 2 → digest BBB → approved/released
```

Corrections do **not** create REV003 merely because Approval requested changes.

`RevisionSubmission` exists even when `ApprovalConfiguration = NoHumanApproval`; it is the canonical fact answering “which exact bytes/metadata were offered for release?”.

Conceptually it captures at least:

```text
id
document_revision_id
attempt_number
submitted_by
submitted_at
source_content_ref
source_content_hash
governed_metadata_snapshot
submission_digest
reason_for_change_snapshot
```

Exact schema waits for technical design.

## R4-07 — Approval binds Submission, not mutable Document/Revision

```text
ApprovalInstance
  → RevisionSubmission
      → DocumentRevision REVxxx
          → Document
```

The Approval engine never chooses content independently.

A new RevisionSubmission creates a new ApprovalInstance when human approval is configured.

## R4-08 — Rendition and Release bind the same Submission

Official DOCX/PDF and release readiness derive from the same RevisionSubmission that human approval evaluated (or that was explicitly submitted when no human approval exists).

This makes the historical “review one body, freeze another/template snapshot” state structurally invalid.

## R4-09 — No-human-approval path has no fake workflow

```text
REV001 DRAFT
→ RevisionSubmission
→ REV001 SUBMITTED
→ rendition/release gates
→ Release Coordinator
→ REV001 EFFECTIVE
```

No zero-stage ApprovalPolicy, auto-approval instance or submitter-as-approver evidence is created.

## R4-10 — Post-approval content cannot reopen in V1

Once the current Submission has completed required human approval, those bytes remain immutable.

If issuance must be stopped, cancel the candidate Revision with reason/audit. If the business still needs a changed candidate, create the next REV. A special “reopen approved but not effective” feature is deferred until a real need proves it.

## R4-11 — Reason for change belongs to the business Revision

`REV002+` requires a reason-for-change before first submission. The working reason may evolve while the REV is DRAFT, but every RevisionSubmission snapshots the reason that accompanied its exact bytes.

REV001 may use an initial-issue reason or the later product rule we choose during detailed metadata design.

## R4-12 — Autosave/checkpoints are authoring history, not REV history

Current implementation `Revision` rows that represent each saved content snapshot are not target `DocumentRevision` semantics.

Authoring may keep technical working snapshots/checkpoints/edit sequences inside the open REV, but:

```text
autosave ≠ new REV
checkpoint ≠ new REV
```

## R4-13 — EditorSession binds the open Revision

An editing session edits the current `DRAFT` Revision, not abstract Document state. A stale session never authorizes mutation after the REV transitions to `SUBMITTED`.

## R4-14 — Comments vs decision evidence

Editorial collaboration comments attach to the DocumentRevision/authoring context. Approval decision comments/reasons remain Approval evidence attached to the concrete RevisionSubmission/Step decision; they are not forced into one generic comment table/domain.

## R4-15 — EFFECTIVE and supersession

The prior effective revision remains official while a newer REV is DRAFT/SUBMITTED/approved-but-waiting-release.

When the new REV wins release, one atomic transition performs conceptually:

```text
REV001 EFFECTIVE  → SUPERSEDED
REV002 SUBMITTED  → EFFECTIVE
Document.effective_revision_id = REV002
Document.open_revision_id = null
```

There must never be two effective revisions for one Document.

Revision-to-revision `SUPERSEDED` is therefore a mechanical release consequence, not a separate manual “supersede revision” action.

## R4-16 — OBSOLETE means retirement without a successor

`SUPERSEDED` = replaced by a newer REV of the same Document.

`OBSOLETE` = the Document is intentionally removed from official use without a successor Revision.

Before obsoleting a Document, any open candidate REV must be resolved/cancelled. A Document retired/obsolete is terminal in V1; reactivation is deferred until a concrete requirement proves need.

## R4-17 — Cross-Document replacement remains separate/open

If `PO-COM-001` is replaced by another stable Document such as `PO-COM-017`, that is not ordinary Revision supersession. If the product needs it, design an explicit cross-Document replacement relation later in Release/effectivity work.

Do not preserve `document.supersede` merely because historical code used it for revision mechanics.

## R4-18 — Template-source provenance is immutable

A Document created from a template permanently records the exact source template Document + source REV label/ID + source content digest selected at creation. Placement (Document row, first Revision, or a dedicated provenance value object) remains a technical-data-model decision, but rebinding is never allowed.

---

# 8. LOCKED — Release and official truth

Human approval does not directly effectivate a revision.

Release evaluates explicit prerequisites such as:

- exact RevisionSubmission identity still current for the candidate;
- required human approval complete, when configured;
- required final rendition(s) ready;
- planned effective date reached;
- supersession/effectivity invariants hold.

Then a single release/effectivity transaction changes the official revision pointer/state.

The existing Release Coordinator concept survives because it solves a real asynchronous/effectivity boundary. Final package/context placement remains open.

---

# 9. Build-vs-buy rulings to date

| Technology / class | V1 ruling | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | do not deploy now | enterprise SSO/SAML/OIDC/MFA/passkeys or credential externalization |
| OpenFGA / SpiceDB | do not deploy now | arbitrary sharing/large relation graph/service split |
| Camunda / Flowable / BPMN | do not use for document approval V1 | genuinely generic business-process requirements |
| Temporal as Approval engine | do not use | durable orchestration need not economically served by current async model |
| CEL / expression language | do not use now | real conditional policies not cleanly expressible as typed configuration |

Library/framework decisions happen only after the exact domain responsibility is closed.

---

# 10. Explicit target deletions/replacements

Current concepts with no entitlement to survive:

- target contexts `documents` / `controlleddocuments` / `templates` as three business authorities;
- `ControlledDocument` as a third identity beside Document/Revision;
- `DocumentProfile` target concept;
- behavioral `DocumentFamily`; retained need is classification-only `DocumentTypeCategory`;
- `GovernanceClass {controlado, simples, livre}` and route-shape derivation;
- parallel Template lifecycle/version counter;
- historical roles `system_admin`, `editor`, `signer`, `qms_admin`, `area_admin`;
- role-based bypasses and magic `areaCode="tenant"`;
- multiple semantic grant sources;
- StageKind as two workflow engines;
- configurable approval-stage required capability;
- drift policies;
- terminal human reject as normal V1 outcome;
- reusing one ApprovalInstance after content edits;
- generic time-window delegation as V1 prerequisite;
- M-of-N, generic workflow phases/branching/BPMN/CEL;
- `editable_by_role` on type/profile;
- Approval code mutating Controlled Information tables directly;
- any freeze/rendition path able to select bytes independently of RevisionSubmission;
- autosave rows presented as official DocumentRevision;
- user-facing `v7`-style document revisions;
- manual revision-supersede authority for ordinary REV progression;
- old roadmap/milestone/spec material as forward authority.

Prefer deletion to compatibility shims when no deployed/contractual requirement proves compatibility necessary.

---

# 11. Remaining design program

## R0 — Authentication boundary — **CLOSED**

## R1 — Organization + Authorization V1 — **CLOSED at semantic level**

Final Permission Catalog waits for whole-product operation census.

## R2 — Approval V1 — **CLOSED at semantic level**

Final integration depends on R4/R6 release/rendition details.

## R3 — Controlled Information configuration — **CLOSED**

Locked:

- DocumentProfile → DocumentType;
- DocumentType code immutable, ACTIVE/INACTIVE, no own versioning;
- type immutable on Document in V1;
- Family → optional classification-only DocumentTypeCategory;
- GovernanceClass deleted;
- explicit `NoHumanApproval | UsePolicy(...)`;
- template as role of governed Document;
- TemplateUse M:N + at most one default per type;
- no `template_required` without requirement;
- exact source template REV/hash pinned at creation;
- official revision labels `REV001`, `REV002`, ... .

## R4 — Document / Revision / Submission lifecycle — **CLOSED**

Locked:

- stable Document identity;
- one effective + at most one open REV;
- DocumentRevision = business `REVxxx`, not autosave;
- Revision states: DRAFT/SUBMITTED/EFFECTIVE/SUPERSEDED/OBSOLETE/CANCELLED;
- new REV allocated when change cycle begins, never reused;
- RevisionSubmission is first-class immutable attempt;
- return-for-changes/withdraw return same REV to DRAFT;
- each resubmission = new RevisionSubmission and, when required, new ApprovalInstance;
- Approval/Rendition/Release bind the exact Submission;
- no fake ApprovalInstance for no-human-approval;
- reason-for-change belongs to REV and is snapshotted by Submission;
- authoring snapshots/checkpoints are technical history inside a REV;
- ordinary prior REV supersession is mechanical at release;
- obsolete = retire without successor;
- cross-Document replacement remains separate/open.

## R5 — Numbering + Template authoring payload — **NEXT**

Close:

1. Document business-code NumberSeries/format/scope;
2. relationship between DocumentType + Area and code generation;
3. code allocation timing and immutability;
4. whether any manual code override exists;
5. exact `TemplateSpec` revision payload;
6. source DOCX/body + placeholder/schema representation;
7. template provenance storage boundary;
8. behavior when template REV is superseded;
9. whether derived Documents surface “new template available” without rebinding;
10. metadata model and which metadata belongs to Document vs REV vs Submission.

## R6 — Periodic review + Renditions + Release/effectivity

Close periodic-review policy/evidence, final rendition provenance/reconstruction, renderer-version evidence, Release Coordinator contract, planned effective date, revision supersession and any cross-Document replacement semantics.

## R7 — Distribution + Tokens + Audit + Notifications + Search

Close read/ack obligation/evidence, token/computed-value snapshot timing, audit authority, domain events, notifications and rebuildable search projections.

## R8 — Tenant lifecycle + Security

Close tenant/control-plane authority, deletion request/grace/system erasure, security/MFA/tenant-key ownership and external-IdP migration trigger.

## R9 — Final Authorization Matrix

Only after all product operations are known: final Permission Catalog, five role bundles, group/admin operations, visibility/filter semantics, workflow relationships, Domain Constraints, RLS/tripwire backstops and Golden Matrix.

## R10 — Technical architecture

Build-vs-buy final pass, bounded contexts/packages, dependency DAG, table ownership, transaction boundaries, data model/constraints, event/outbox contracts.

## R11 — API + frontend journeys + migration map

OpenAPI operations/DTOs, frontend IA/journeys, migration/delete/rename map, compatibility policy and proof matrix.

## R12 — Final ADR/spec promotion + adversarial review

Promote final durable truth to `wiki/`, self-review for ambiguity/contradictions, operator review.

## R13 — Implementation plan

Only after integrated design approval.

## R14 — Product implementation

Blocked until all prior gates are closed.

---

# 12. Implementation gate

Implementation begins only when all are true:

- [ ] whole-product domain map closed;
- [x] R3 Controlled Information configuration closed;
- [x] R4 Document/Revision/Submission lifecycle closed;
- [ ] TemplateSpec/numbering/metadata closed;
- [ ] Approval integrated with final lifecycle/release contract;
- [ ] Organization/AuthZ integrated with every real operation;
- [ ] release/effectivity/rendition closed;
- [ ] periodic review closed;
- [ ] distribution/read/ack closed;
- [ ] audit/evidence boundaries closed;
- [ ] notifications/search/tokens/security/tenant lifecycle dispositioned;
- [ ] build-vs-buy final per responsibility;
- [ ] bounded contexts + table/transaction ownership closed;
- [ ] final Permission + Role matrix closed;
- [ ] API/UX contract closed;
- [ ] migration/delete/rename map closed;
- [ ] final ADR/spec set promoted to wiki;
- [ ] adversarial review finds no material ambiguity;
- [ ] operator approves integrated design;
- [ ] implementation plan then authored from accepted target.

Until then: **design/documentation only.**

---

# 13. Exact next design step

Continue with **R5 — Numbering + Template authoring payload + metadata placement**:

```text
business document code / NumberSeries
+ DocumentType/Area scope
+ code allocation timing/immutability
+ TemplateSpec exact governed payload
+ placeholder/schema/body representation
+ source-template provenance placement
+ newer-template availability semantics
+ Document vs REV vs Submission metadata ownership
```

Apply the same test to every proposed field/entity: does it represent an independent business fact with one authority, or is it merely historical encoding/duplication?

Do not implement.