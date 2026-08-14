# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding for this redesign; open items are explicitly marked.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Governing method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Canonical program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED. No product code, schema, API or migration implementation is authorized yet.**

---

## 0. Fresh-session contract

A new session working on this program reads, in order:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. this ledger
5. `wiki/references/current-agent-handoff.md`

Do not resume an old roadmap, milestone, spec, PR implementation or ADR execution sequence by inertia. Historical code/docs are evidence about requirements and failures, not target authority.

The working method is deliberately **design-first**:

```text
product/domain semantics
→ invariants + lifecycle
→ organization/authz/workflow integration
→ build-vs-buy
→ bounded contexts / ownership
→ data model
→ contracts / APIs / UX
→ migration/deletion map
→ implementation spec
→ implementation plan
→ code
```

The goal is that implementation contains no unresolved product or architecture judgment.

---

# 1. Why the redesign expanded beyond AuthZ

The work began from a real authorization defect class: MetalDocs had more than one grant authority and could answer the same authorization question differently depending on the enforcement path.

While redesigning AuthZ, approval permissions could not be finalized without trusting the Approval/Workflow model. Inspecting Approval then exposed that workflow, Documents, Controlled Documents, Templates, taxonomy and release shared or duplicated authority. The architecture audit had already measured the same shape: package-level layering was relatively disciplined, but the module graph contained a large strongly-connected cluster and foreign-table access, including Approval mutating Documents state directly.

A browser QA run supplied the strongest product-level counterexample. The editor-authored revision was reviewed by the approver, but the freeze pipeline rendered the blank template snapshot instead of the reviewed revision. The final signed PDF was blank and the signed hash did not correspond to the content the human reviewed. That is not an endpoint defect; it proves the system had two competing authorities for “what content is the document”.

### Root cause

MetalDocs evolved as locally reasonable modules/features instead of one coherent controlled-information model. Several nouns are implementation history rather than business truth, and several policies are duplicated across modules.

### Target property

Each business fact has exactly one authority. Supporting modules consume that authority; they do not reinterpret or mutate it independently. Essential complexity remains; accidental complexity is deleted.

### Decision

**Restructure now at the semantic/design level.** Existing code has no right to survive merely because it exists. No implementation begins until the integrated target is closed.

---

# 2. Whole-product responsibility census

The current repository contains 15 `internal/modules` directories. The redesign must account for all their legitimate responsibilities even where their current module boundary will not survive.

| Current module | Legitimate responsibility | Target disposition at this stage |
|---|---|---|
| `approval` | human approval policy, tasks/steps, decisions, evidence | **Redesign** as the simpler Approval V1 below; release no longer treated as human approval responsibility |
| `audit` | immutable/hash-chained regulatory evidence and export | **Retain as distinct evidence authority**; later reconcile transaction/contract seam |
| `auth` | login, credential, session lifecycle | **Retain for V1** behind an authentication boundary; future external IdP seam |
| `controlleddocuments` | stable numbered document identity, profile/area binding, sequence | **Retire as a separate target context**; legitimate responsibilities move into Controlled Information + numbering |
| `distribution` | released-revision recipient/coverage read model; future read/ack | **Retain/re-evaluate as a supporting released-revision concern**; future evidence write-path remains a separate design decision |
| `documents` | editor/draft/revision/comments/freeze/view/export | **Become the Controlled Information core after responsibility cleanup** |
| `iam` | tenant/people/groups/grants/authz plus tenant lifecycle administration | **Split conceptually into Organization + Authorization**; exact package boundary still open |
| `jobs` | periodic River orchestration over other domains | **Not a bounded context**; re-home as composition/orchestration infrastructure |
| `notifications` | user notification inbox + async fanout | **Retain as a supporting consumer of domain events**, not a workflow authority |
| `render` | materialization/PDF dispatch, renderer seam, computed resolvers | **Retain as supporting rendering/rendition infrastructure**, aligned to Revision truth |
| `search` | cross-entity query/projection | **Retain as read model/projection consumer** |
| `security` | security signals/MFA coverage/lockouts/tenant-key security | **Retain/re-evaluate** after AuthN/Organization/Tenant lifecycle target settles |
| `taxonomy` | areas, document profiles/types, families, governance class | **Dismantle as current context**: Area → Organization; Profile → DocumentType candidate; Family/classification re-evaluated |
| `templates` | template upload/schema/placeholders/versioning/lifecycle | **Retire parallel lifecycle**; template becomes a role/designation of an exact DocumentRevision |
| `tokens` | tenant dictionary values + published read port | **Retain as supporting configuration/value provider**, later reconcile with Controlled Information snapshot semantics |

Additional product responsibilities that are not reducible to those module names and must be covered before implementation:

- periodic review and reason-for-change;
- number series / business-code allocation;
- immutable submission snapshots/evidence;
- final rendition provenance and reconstruction/attestation;
- release/effectivity/supersession;
- distribution obligations/read/acknowledgement;
- notifications;
- search/read projections;
- tenant lifecycle/deletion;
- audit evidence;
- async outbox/jobs/worker behavior;
- API/contract and frontend journeys.

This is the “nothing missing” checklist for the redesign. A module may disappear while its real responsibility survives elsewhere.

---

# 3. LOCKED — Authentication boundary

## AUTHN-01 — AuthN and AuthZ are separate authorities

Authentication answers **who is this actor/session?** Authorization answers **what may this principal do in this tenant/scope/resource?** Approval answers **who participates in this case?** Domain Governance answers **is the action legal now?**

None substitutes for another.

## AUTHN-02 — Do not introduce Keycloak now

The current MetalDocs authentication/session implementation already provides the V1 properties required to continue the product redesign: credential verification, opaque/server-side session semantics, revocation/lockout and tenant-bound identity context.

Keycloak/OIDC/SAML/MFA/passkeys/enterprise federation solve legitimate future requirements, but none is currently necessary to make the target product correct.

Target seam:

```text
Authentication boundary
    ├── current MetalDocs implementation (V1)
    └── future OIDC/Keycloak/external IdP adapter when a concrete enterprise requirement fires
```

Trigger examples: tenant-specific Entra/Google/Okta SSO, SAML, enterprise federation, broad MFA/passkey policy, or a deliberate decision to stop operating credentials.

---

# 4. LOCKED — Organization + Authorization V1

## ORG-01 — Area is Organization, not document taxonomy

V1 organizational model:

```text
Tenant
Area
User
Group
GroupMembership
```

Area is reused by Document ownership, scoped permissions and Approval participant resolution. One Area definition, one meaning.

## ORG-02 — Groups are first-class flat principals

Groups solve a demonstrated requirement: a team such as `Vendedores` can receive one assignment (`author @ COMERCIAL`) and membership administration becomes the only recurring operation.

Rules:

- User may belong to multiple Groups.
- No nested Groups in V1.
- Group does not receive raw Permissions.
- Group receives Role Assignments through the same path as User.
- Group need not structurally “belong” to one Area; scope is expressed by RoleAssignment.

Example:

```text
Group Vendedores
  members: Beatriz, João, Jordana
  assignment: author @ Area(COMERCIAL)
```

## AUTHZ-01 — Five built-in V1 roles

```text
tenant_owner
area_manager
author
approver
viewer
```

Historical `system_admin`, `editor`, `signer`, `qms_admin`, `area_admin` do not survive as built-in product roles unless a later requirement proves an independent responsibility.

Roles are named bundles. Runtime authorization checks Permissions, never hard-coded role branches.

## AUTHZ-02 — Scoped RoleAssignment is the single grant shape

Conceptually:

```text
RoleAssignment
  subject: User | Group
  role
  scope: Tenant | Area
  granted_by / granted_at
  revoked_by? / revoked_at? / reason?
```

V1 does **not** require temporal `valid_from/valid_until` grant scheduling. Historical grant/revoke evidence does survive. Temporary grants may be added later on demonstrated need.

Legal built-in scopes:

| Role | Tenant | Area |
|---|---:|---:|
| tenant_owner | yes | no |
| area_manager | no | yes |
| author | yes | yes |
| approver | yes | yes |
| viewer | yes | yes |

Effective grants are the union of applicable direct User assignments plus Group assignments. Default is deny. No explicit deny rule system in V1.

## AUTHZ-03 — Role semantics

### `viewer`
Published/effective official information in scope. No draft/working access by base role.

### `author`
Published + working content in scope; create/edit eligible drafts; comment; submit. `created_by` is evidence, not authorization ownership.

### `approver`
Qualification/eligibility, **not blanket authority over every draft in an Area**. Base role does not grant broad working-content visibility. A concrete workflow relationship opens only the working content needed for that case.

### `area_manager`
Operational manager in assigned Area: Author-like operations plus workflow oversight/cancel/reassign and eligible obsolete/supersede operations. It is **not** IAM/RBAC administration.

### `tenant_owner`
Tenant product administrator through ordinary Role→Permission grants. Never an authorization bypass and never a Domain Governance bypass.

## AUTHZ-04 — Permissions, not role checks

Current candidate semantic permissions include:

```text
document.read_published
document.read_working
document.create
document.edit
document.comment
document.submit
document.obsolete
document.supersede

approval.act
approval.oversee
approval.cancel
approval.reassign
approval_policy.manage

organization / access-administration permissions (to be finalized with the complete domain)
```

This list is **not final** until the Controlled Information design is closed. Permission count is never a goal. A permission survives only if a professional administrator would legitimately want to grant/deny the authority independently.

## AUTHZ-05 — Authorization + workflow + Domain Governance compose

For a concrete human decision:

```text
base qualification/Permission at scope
+
workflow participation in this StepInstance
+
Domain Governance constraints (state, SoD, reauth, revision/hash, etc.)
=
ALLOW
```

No role, including tenant owner, skips Domain Governance.

## AUTHZ-06 — Do not introduce OpenFGA/SpiceDB now

The demonstrated V1 model is scoped RBAC + flat groups plus natural domain relationships such as ApprovalStepParticipant. A generic ReBAC tuple engine would add a second consistency domain/datastore without a current requirement for arbitrary per-resource sharing graphs.

Design the Authorizer as a boundary (`Check` plus collection/scoping support) so a future external engine remains possible if the product later gains large relationship graphs or arbitrary resource sharing.

---

# 5. LOCKED — Approval V1

## APPR-01 — Specialized governed-document approval, not BPM

MetalDocs does not build a generic business-process/workflow platform in V1.

Explicitly absent unless a future requirement proves need:

- BPMN;
- arbitrary gateways/branches/loops/subprocesses;
- generic service tasks;
- Camunda/Flowable runtime;
- CEL/expression-language policy system;
- general parallel phase graph.

## APPR-02 — Versioned sequential ApprovalPolicy

```text
ApprovalPolicy
  id
  version
  ordered ApprovalSteps[]
```

A new version applies to new submissions. An in-flight/historical ApprovalInstance remains bound to the version from which it was created.

## APPR-03 — ApprovalStep is the human task

No `Phase → HumanTask` hierarchy in V1.

Conceptually a Step owns:

```text
order
name
purpose: review | approval
actor_rule
completion: ANY | ALL
requires_reauthentication
due_in_days?
```

`review` and `approval` describe business meaning/UI/evidence. They do not create two execution engines.

## APPR-04 — Initial participant rules

V1 needs only:

```text
NamedUser
Group
RoleInArea(role, fixed-area | subject-area)
```

Historical submit-time arbitrary actor choice is not part of the V1 target until a real workflow proves it necessary.

## APPR-05 — Completion rules ANY / ALL only

No `M-of-N`, weighted voting or percentage quorum in V1. `AT_LEAST(n)` remains an easy future extension if a concrete business requirement appears.

## APPR-06 — Participant resolution timing

Participants are resolved when their Step activates and are snapshotted into the StepInstance.

The snapshot is evidence of who was selected, **not an irrevocable authorization grant**. When a user acts, current qualification/permission is checked again.

If an `ALL` participant becomes unavailable/unqualified, do not silently shrink quorum and do not label the business decision “rejected”. The Step enters an explicit attention/blocked condition and an authorized manager performs audited reassignment.

Historical drift policies such as `reduce_quorum`, `fail_stage`, `keep_snapshot` are targeted for deletion.

## APPR-07 — Decisions

Human decision outcomes:

```text
accept
return_for_changes
```

UI wording may be “Concluir revisão”, “Aprovar”, “Solicitar correções”, etc., based on Step purpose.

No normal terminal human `reject` concept in V1.

Separate lifecycle/administrative operations:

```text
withdraw
cancel
reassign
```

## APPR-08 — Return-for-changes creates a clean new attempt

An ApprovalInstance is bound to an immutable submission/revision/hash.

`return_for_changes` terminates that attempt. Editing creates new content/revision state; resubmission creates a **new ApprovalInstance**. Decisions over old bytes never carry forward as approval of new bytes.

## APPR-09 — Deadlines and reassignment are deliberately simple

A Step may have `due_in_days` → `due_at`, with overdue surfacing/notification.

V1 does not require a generic escalation-chain engine.

Audited `reassign(old_actor, new_actor, reason, performed_by, timestamp)` covers absence/error/termination. The current sophisticated time-window delegation engine is deferred; bring it back only if recurring substitution (for example vacations) proves the need.

## APPR-10 — Reauthentication is a Step requirement

A Step may require fresh credential verification before `accept`. This is a Domain Governance/evidence requirement and does not require Keycloak. Future external IdP/MFA can satisfy the same seam.

## APPR-11 — Evidence stays strong

Every decision must remain explainable against at least:

```text
actor
when
ApprovalPolicy + version
StepInstance + purpose
Document / Revision / immutable submission
content digest/hash
outcome
comment/reason
reauth evidence when required
```

---

# 6. LOCKED — Controlled Information north star

## CI-01 — One Controlled Information context

The existing target split `documents` + `controlleddocuments` + `templates` is rejected.

The accepted semantic direction is one Controlled Information context centered on:

```text
Document
DocumentRevision
```

The exact Go package/module layout is deliberately deferred until the domain design finishes.

## CI-02 — Document is stable governed identity

`Document` answers “what controlled information is this?” and owns stable business identity such as tenant, code, type, owning area, lifecycle-level identity and the effective revision pointer.

The current public/domain noun `ControlledDocument` is targeted for retirement as a separate third object. Its legitimate identity/numbering/effectivity responsibilities move to `Document` or explicit policy/supporting concepts.

## CI-03 — DocumentRevision is versioned governed content

`DocumentRevision` is the content candidate/issuance. A draft may be mutable through authoring infrastructure, but submission freezes an immutable content identity/hash. Effective/released revisions are immutable; changes require a new revision.

The editor/autosave implementation may use working state, but “AuthoringWorkspace” is not automatically promoted to a large business aggregate. Preserve the technical property (high-frequency working writes do not contend on effectivity) without inventing a new noun unless the domain proves it.

## CI-04 — Template is a role/designation of an exact DocumentRevision

Template does **not** have a parallel lifecycle/version counter.

A template is an exact governed revision designated as usable to seed new documents.

Changing any governed template content means a new DocumentRevision, including:

- DOCX/body layout;
- placeholder schema;
- placeholder type;
- required/default/options;
- validation constraint;
- visibility condition;
- resolver binding/contract;
- any other material authoring rule embedded in that template revision.

This is operator-approved and eliminates the “TemplateVersion vs DocumentRevision” double-versioning problem.

## CI-05 — Template seeds; Revision becomes content truth

Creating a document from a template binds provenance to the exact source revision/hash. The source template may seed initial editor content and schema, but after creation the new document's own Revision is the truth being edited/reviewed/approved.

A later template revision never silently rebinds existing documents.

## CI-06 — Freeze/approval/rendition bind the reviewed revision

The only legal source for immutable submission evidence, Approval and official Rendition is the exact revision/content digest the human reviewed.

The historical QA failure where the editor revision was reviewed but a blank template snapshot was frozen must become structurally unrepresentable.

## CI-07 — Profile → DocumentType is the current approved direction

The current `DocumentProfile` already acts as a tenant document type but mixes classification, authorization, review, template and governance settings.

Target noun: `DocumentType` (examples IT, PO, RG, PL).

A DocumentType references policies/configuration owned by the appropriate authority rather than reimplementing them.

Example direction:

```text
DocumentType IT
  numbering policy
  approval policy?
  periodic-review policy?
  template designation/default?
```

`editable_by_role`-style configuration does not survive; Authorization owns authority.

## CI-08 — Area leaves taxonomy

Area belongs to Organization and is referenced by Document, RoleAssignment and Approval actor resolution.

## CI-09 — Family and GovernanceClass remain OPEN for independent-value testing

`DocumentFamily` may survive as a cheap classification/navigation grouping if users genuinely need it. It must not become a second behavior authority without proof.

`GovernanceClass {controlado, simples, livre}` survives only if it has independent business meaning beyond deriving Approval shape — for example if it drives retention, distribution, training, periodic review, reporting or other coherent policy. If it merely duplicates ApprovalPolicy configuration, delete it.

## CI-10 — TemplateDesignation cardinality remains OPEN

The current likely direction is a relationship between an effective template Revision and one or more DocumentTypes rather than a permanent `is_template` subtype, but exact cardinality/default/eligibility semantics still need to be closed.

---

# 7. LOCKED — Release and official truth

Human approval does not directly publish/effectivate a document.

Approval produces decision evidence. The release boundary evaluates mechanical/domain prerequisites such as:

- approval receipt complete;
- same immutable revision/hash;
- final required renditions/artifacts ready;
- effective date reached;
- supersession/effectivity invariants hold.

Then a single release/effectivity transition updates the official revision pointer/state.

The previous effective revision remains official until the new one is legally released. There must never be two effective revisions for the same Document.

The existing Release Coordinator concept survives because it solves a real asynchronous/effectivity boundary. Its final package/context placement is still open.

---

# 8. Build-vs-buy decisions to date

These are architectural rulings, not claims that the technologies are poor.

| Technology / class | V1 decision | Revisit trigger |
|---|---|---|
| Keycloak / external IdP | **Do not deploy now** | enterprise SSO/SAML/OIDC federation/MFA/passkeys or decision to externalize credentials |
| OpenFGA / SpiceDB | **Do not deploy now** | arbitrary resource sharing / large relationship graph / service split makes centralized ReBAC valuable |
| Camunda / Flowable / BPMN | **Do not use for document approval V1** | product genuinely becomes a generic business-process engine with branches/service tasks/subprocesses |
| Temporal as Approval engine | **Do not use** | separate durable orchestration requirement appears that the current outbox/River model cannot economically serve |
| CEL / expression language | **Do not use now** | real conditional workflow/product-policy rules appear that cannot be represented cleanly by typed configuration |

Libraries/frameworks will be selected only after each exact domain problem is closed. “Can outsource” is evaluated per responsibility rather than as a preference for more infrastructure.

---

# 9. Concepts explicitly targeted for deletion/replacement

The implementation migration later starts from this subtractive list; it is not implementation authorization.

Current concepts that have **no entitlement to survive**:

- three target contexts `documents` / `controlleddocuments` / `templates`;
- public/domain `ControlledDocument` as a separate identity beside Document;
- independent Template lifecycle/versioning parallel to DocumentRevision;
- old built-in roles `system_admin`, `editor`, `signer`, `qms_admin`, `area_admin`;
- role-based authorization branches/bypasses;
- magic `areaCode="tenant"` scope sentinel;
- multiple grant sources (`iam_user_roles`, `user_process_areas`, group-role paths as separate semantic authorities);
- `StageKind` as two workflow engines;
- configurable `required_capability` on approval stages;
- drift policies that silently mutate completion requirements;
- normal terminal human `reject` semantics;
- old `changes_requested` instance reuse across edited content;
- generic time-window delegation as a V1 prerequisite;
- `M-of-N` as V1 requirement;
- generic workflow phases/branching/BPMN/CEL;
- `editable_by_role` on document type/profile;
- GovernanceClass as an Approval-shape duplicate if no independent domain meaning survives;
- Approval code that writes Controlled Information tables directly;
- any freeze path that can select content other than the submitted/reviewed Revision;
- old roadmap/milestone/spec documents as live authority.

Deletion is preferred to compatibility shims where no deployed/contractual compatibility requirement proves a shim necessary.

---

# 10. Supporting concerns that MUST be closed before implementation

AuthZ + Approval + Documents are not sufficient to claim a whole-platform design. The following sections must be explicitly ruled before the implementation gate opens.

## DESIGN QUEUE A — Controlled Information configuration

1. `DocumentType` exact responsibility and lifecycle.
2. `DocumentFamily`: retain as classification vs delete/simplify.
3. `GovernanceClass`: independent policy value vs delete.
4. template designation/default/eligibility and cardinality.
5. metadata model and which facts snapshot on revision/submission.

## DESIGN QUEUE B — Document / Revision lifecycle

1. exact Document lifecycle states;
2. exact DocumentRevision lifecycle states;
3. when a new revision number/label is allocated;
4. draft/working-content persistence vs immutable submission snapshot;
5. comment/checkpoint/editor-session ownership;
6. reason-for-change semantics;
7. withdrawal/cancel/obsolete/supersede relationships;
8. one-effective-revision invariant and atomic boundary.

## DESIGN QUEUE C — Numbering

1. NumberSeries/numbering policy scope;
2. area/type relationship to code generation;
3. allocation timing and immutability;
4. manual override requirements, if any.

## DESIGN QUEUE D — Template authoring semantics

1. exact `TemplateSpec` contents as revision payload;
2. how template source DOCX and placeholder/schema data are represented;
3. derived-document provenance snapshot;
4. what happens when a template revision is superseded;
5. whether documents ever track “recommended newer template” without automatic rebinding.

## DESIGN QUEUE E — Periodic review

1. review policy ownership (likely DocumentType/policy reference);
2. review due calculation and effect on released revision/document;
3. “reviewed, no change” evidence vs “new revision required” path;
4. permissions and notifications.

## DESIGN QUEUE F — Renditions / rendering / evidence

1. Revision vs Submission vs Rendition ownership;
2. immutable DOCX/PDF provenance tuple;
3. reconstruction/attestation semantics;
4. renderer version/hash evidence;
5. final-artifact readiness relation to Release Coordinator.

## DESIGN QUEUE G — Distribution / read / acknowledgement

1. obligation denominator: snapshot at release vs live derivation;
2. read event semantics;
3. explicit acknowledgement semantics;
4. whether acknowledgement requires reauth/signature;
5. reminders/deadlines/export;
6. permissions and relationship to Groups/Areas;
7. emitted events and notification integration.

## DESIGN QUEUE H — Tokens / computed values / metadata

1. tenant dictionary values and pinning semantics;
2. computed resolver catalogue ownership;
3. whether values are frozen at document creation, submission or rendition depending on meaning;
4. collision/version/provenance rules.

## DESIGN QUEUE I — Audit / Evidence

1. single audit authority and event ownership;
2. same-transaction business mutation evidence requirements;
3. immutable Approval evidence vs global audit event relationship;
4. export/integrity/tenant-erasure behavior.

## DESIGN QUEUE J — Notifications + Search

1. notifications consume events and do not own workflow/business state;
2. exact domain events published by Controlled Information/Approval/Distribution;
3. search indexes/projections derive from canonical released/working visibility rules;
4. rebuildability and eventual-consistency expectations.

## DESIGN QUEUE K — Tenant lifecycle / security

1. Tenant authority boundary after Organization/IAM split;
2. tenant owner vs platform operator;
3. deletion request/grace/system erasure lifecycle;
4. security signals, MFA state and crypto-key ownership;
5. external IdP trigger and migration seam.

## DESIGN QUEUE L — Final Authorization matrix

Only after all product operations above are known:

1. final Permission Catalog;
2. five built-in Role bundles;
3. Group/admin operations;
4. visibility/filtering semantics;
5. workflow relationship checks;
6. Domain Constraint matrix;
7. DB/RLS/tripwire backstops;
8. Golden Matrix of positive/negative cases.

## DESIGN QUEUE M — Technical architecture and implementation contract

After domain closure:

1. build-vs-buy final pass;
2. bounded contexts/packages and dependency DAG;
3. table ownership and transaction boundaries;
4. data model + DB constraints;
5. event/outbox contracts;
6. OpenAPI operations and DTOs;
7. frontend information architecture/journeys;
8. migration/delete/rename map from current code;
9. compatibility policy;
10. test/proof matrix;
11. implementation specification;
12. implementation plan.

---

# 11. Documentation authority reset

The repository had multiple competing forward roadmaps, milestone trees, specs, reports and living module pages. The current program deliberately collapses that authority.

Rules from 2026-08-14 forward:

- `wiki/architecture/cohesive-platform-redesign.md` is the canonical current-program entrypoint.
- this ledger is the only active detailed WIP decision source.
- `wiki/references/current-agent-handoff.md` is the recovery pointer only; it must stay short.
- old `docs/superpowers/{ROADMAP,milestones,plans,reports,specs,analysis}` material is removed from the live tree; Git history is the archive.
- affected wiki module pages are marked/replaced as LEGACY current-state references and must not be used as target architecture.
- old roadmap/backlog surfaces are frozen/repointed to this program.
- historical ADRs remain useful evidence, but where their target semantics conflict with this ledger they do not control the redesign. Final accepted decisions will be written as new/amending/superseding ADRs at design closure.

Do not resurrect deleted documents by copying them from Git history into the live tree. Recover history only to answer a specific evidence question.

---

# 12. Implementation gate

Implementation begins only after all of the following are true:

- [ ] whole-product domain map closed;
- [ ] Controlled Information configuration/lifecycle closed;
- [ ] Template-as-revision-role semantics closed;
- [ ] Approval V1 integrated with the final document lifecycle;
- [ ] Organization/AuthZ integrated with every real product operation;
- [ ] Release/effectivity/rendition model closed;
- [ ] periodic review closed;
- [ ] distribution/read/ack decision closed;
- [ ] audit/evidence boundaries closed;
- [ ] notifications/search/tokens/security/tenant lifecycle dispositioned;
- [ ] build-vs-buy decisions final for each responsibility;
- [ ] bounded contexts and table/transaction ownership closed;
- [ ] final Permission + Role matrix closed;
- [ ] API/UX lifecycle and operation contract closed;
- [ ] migration/deletion map from current code closed;
- [ ] final ADR/spec set promoted to `wiki/`;
- [ ] adversarial self-review finds no material ambiguity;
- [ ] operator approves the integrated design;
- [ ] implementation plan is then authored from the accepted target.

Until then: **design/documentation only.**

---

# 13. Exact next design step

Continue with **Controlled Information configuration** before touching code:

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ Area ownership (already Organization)
+ TemplateDesignation/default policy
```

For each concept apply the same test:

> Does this noun/policy represent independent business meaning in the target product, or is it only a historical way the current implementation encoded another authority?

Only after this section is operator-approved proceed to the Document/Revision lifecycle section.
