# Authorization + Approval/Workflow Redesign — Decision Ledger

> **Status:** WIP decision ledger — approved decisions are binding for the redesign; open items are explicitly marked. This file is **not** implementation authorization and is not yet canonical wiki truth.
> **Date:** 2026-08-13
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch:** `docs/a8-authz-approval-redesign-ledger`
> **Primary program:** A8 / issue #89 — authorization grant-model redesign
> **Related architecture:** Approval / workflow kernel, Controlled Information (ADR 0093), release coordinator (ADR 0085), tenant lifecycle (ADR 0070)
> **Governing engineering method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Documentation rule:** staging work stays under `docs/`; accepted durable truth is promoted to the owning `wiki/` section after design approval.

## 1. Why this ledger exists

The authorization redesign started from A8's known root cause: MetalDocs currently has more than one authorization authority and can answer the same access question differently depending on which path is exercised. During redesign, the work exposed a second dependency: several proposed `approval.*` permissions were being derived from the current approval-route implementation even though that implementation has been repeatedly redesigned and is not yet trusted as the target domain model.

The operator explicitly required that decisions not live only in chat/session context. This ledger therefore records every approved architectural decision, the current candidate catalog, repository findings, and all open questions. It will be updated as the design proceeds. When the design converges, its durable conclusions will be promoted into canonical ADR/wiki material and the implementation plan will be written from that accepted target.

## 2. Engineering Decision Record

### Symptom

Authorization is fragmented across role/capability tables, area memberships, route-level capability mapping, in-transaction checks, bypasses, and DB tripwires. Approval routing also mixes authorization eligibility, actor selection, workflow configuration, stage execution, domain constraints, and evidence concerns.

### Root cause

Two related structural problems exist:

1. **Authorization has multiple semantic authorities.** Tier-1 and tier-2 derive grants from different relations; `system_admin` bypasses normal evaluation; `areaCode="tenant"` is a magic scope sentinel; role vocabulary and capability metadata are duplicated across layers.
2. **Approval/workflow semantics have evolved incrementally without a single greenfield synthesis.** Route stages, roles, capabilities, selectors, candidate resolution, snapshots, quorum, delegation, SoD, deadlines, cancellation, and release behavior have been repeatedly added or reshaped. Some individual pieces are sound, but the target whole has not yet been independently validated against mature workflow/eQMS patterns.

### Target property

MetalDocs must have:

- one coherent authorization semantics;
- explicit, typed scopes;
- professional role/permission vocabulary;
- workflow participation modeled separately from product authorization;
- domain constraints that cannot be bypassed by administrative roles;
- auditable, versioned approval policy and approval evidence;
- no duplicated authority between IAM, workflow routing, domain rules, and DB backstops;
- invalid or ambiguous states made structurally impossible where reasonably achievable.

### Authority and boundary

Target authorities are deliberately separated:

1. **Authorization** — roles, permissions, assignments, groups, scopes, resource relationships, `Check/Filter/Explain`.
2. **Approval Policy / Workflow Definition** — stages, actor-selection policy, ordering, completion policy, deadlines/escalation, applicability and versioning.
3. **Approval Instance / Human Work** — instantiated policy version, resolved candidates/assignees, stage instances, decisions, delegation usage, immutable evidence.
4. **Domain Governance** — SoD, reauthentication, signature rules, submission freeze, legal state transitions, withdrawal/cancellation legality, terminal-state invariants.
5. **RLS / DB constraints / tripwires** — defense in depth and invariant backstops, not a second product-authorization semantics.

### Local-maximum candidate

Rename `Capability` to `Permission`, re-seed the current 39 capabilities, rebase old A8.1 `capability_bindings`, and continue adapting the existing approval-route model.

### Global-maximum candidate

Redesign Authorization and Approval/Workflow together at the semantic level, while preserving separate authorities. Derive the final `approval.*` permission set only after the target human-workflow model is trusted. Replace dual-source grants, role bypasses, magic scope sentinels, legacy role vocabulary, and module-shape permissions. Preserve only approval concepts that survive independent industry/domain validation.

### Decision

**Restructure now at the design level.** Implementation remains blocked until the integrated design is complete and approved.

### Enforcement

Prefer structural/type/schema boundaries first, runtime fail-closed evaluation second, then tests/static guards only for properties that cannot reasonably be made unrepresentable.

### Proof

Final proof must include a Golden Matrix spanning direct and group grants, tenant/area inheritance, revoked/expired assignments, resource relationships, workflow assignment, SoD, owner/admin behavior, wrong tenant/scope, approval-policy versioning, participant visibility, and negative/counterfactual cases.

### Transitional exit

This ledger is staging material. Exit condition: integrated Authorization + Approval target design approved, self-reviewed, written as final design/ADR set, then implementation plan authored. This ledger is then either promoted or marked historical with pointers to canonical wiki truth.

---

# 3. Locked target architecture

The authorization north star is:

```text
Scoped RBAC
+ Group Principals
+ Resource Relationships
+ Domain Constraints
+ RLS
```

Logical flow:

```text
Permission Catalog
      ↓
Role Definitions
      ↓
Role Permissions
      ↓
Role Assignments ───── Group Memberships
      ↓                     ↓
      └──────── Subject effective grants
                          ↓
                Resource Relationships
                          ↓
                Domain Constraints
                          ↓
             Authorization Evaluator
                Check / Filter / Explain
                          ↓
                   RLS / DB backstop
```

`RLS` is tenant-isolation defense in depth. It is not the product policy engine.

---

# 4. Approved Authorization decisions

## D-AUTH-01 — Tenant owner is tenant-scoped, not platform-global

**APPROVED.** `tenant_owner` is the owner/admin of one tenant/company. A future platform operator/control plane is a different authority and must not receive implicit access to tenant content.

The historical `system_admin` role conflates tenant and platform authority and does not survive the target model in its current shape.

## D-AUTH-02 — Built-in roles V1

**APPROVED.** V1 exposes exactly five built-in roles:

1. `tenant_owner`
2. `area_manager`
3. `author`
4. `approver`
5. `viewer`

The architecture may support future custom roles, but V1 does not expose custom-role creation/editing UI or API.

Roles to eliminate unless a later functional census proves an independently necessary responsibility:

- `system_admin`
- `editor`
- `signer`
- `qms_admin`
- `area_admin`

The five role definitions must have one canonical authority rather than duplicated hard-coded enums across backend, DB, OpenAPI and frontend.

## D-AUTH-03 — Viewer semantics

**APPROVED.** `viewer @ scope` sees only released/published/controlled official information within that scope.

It does **not** gain visibility into drafts, mutable working revisions, or internal review artifacts simply because it is a viewer.

Examples:

```text
viewer @ tenant
→ published/controlled information across tenant

viewer @ area:QUALIDADE
→ published/controlled information only in QUALIDADE
```

## D-AUTH-04 — Author semantics

**APPROVED.** `author @ scope`:

- sees published information in scope;
- sees working/draft content in scope;
- creates controlled information;
- edits any eligible draft/working document in scope, regardless of creator;
- submits work into the workflow;
- follows workflow progress for its legitimate subjects.

`created_by` remains audit/metadata, not authorization ownership.

Per-document restricted editing is not V1. If a real requirement appears later, it should be modeled with Resource Relationships rather than changing base author semantics.

## D-AUTH-05 — Approver role means qualification, not blanket case authority

**APPROVED.** `approver @ scope` means the actor is qualified/eligible to participate in approval work in that scope. It does not mean the actor may approve every item in the scope.

A human approval action requires the composition:

```text
role assignment → base approval permission/qualification
+
workflow relationship → this actor is candidate/assigned/eligible for this case
+
domain constraints → current state, SoD, reauth, signature rules, etc.
=
ALLOW
```

The old `signer` role is presumed redundant unless the workflow census proves a distinct business qualification. A future signing credential/qualification should be a domain qualification/constraint, not automatically a new role.

## D-AUTH-06 — Area manager semantics

**APPROVED.** `area_manager` is an **operational manager**, not an RBAC administrator.

Within its assigned area it can broadly operate governed work: read published/working content, create/edit/submit, participate in approval when assigned, administratively cancel eligible workflows, obsolete/supersede where permitted, and oversee operational activity.

It does **not** administer:

- users;
- groups;
- roles;
- permissions;
- role assignments;
- RBAC policy;
- tenant-wide structural policy.

Structural configuration such as taxonomy management, approval-policy administration and template-policy administration remains with `tenant_owner` in V1 unless later design proves a narrower safe delegation model.

## D-AUTH-07 — Groups are first-class principals

**APPROVED.** Groups aggregate users and can receive Role Assignments. Groups do not own raw permissions directly.

```text
Group: Quality Team
members: João, Maria, Pedro
assignments:
  author @ area:QUALIDADE
  approver @ area:QUALIDADE
```

Only `tenant_owner` administers group structure/access administration in V1.

## D-AUTH-08 — Multiple roles compose additively

**APPROVED.** A user/group may hold multiple simultaneous role assignments at the same or different scopes.

```text
João
  viewer @ tenant
  author @ area:QUALIDADE
  approver @ area:QUALIDADE
```

Effective permission authority is the union of applicable active assignments, subject to resource relationships and domain constraints.

Do not create combinatorial roles such as `author_approver`.

## D-AUTH-09 — Typed scope hierarchy

**APPROVED.** Initial scope kinds:

- `TenantScope`
- `AreaScope`

Inheritance:

```text
tenant assignment → applies downward to all areas/resources in tenant
area assignment   → applies only to that area and descendants/resources
```

No upward inheritance and no sibling-area inheritance.

Assignable scopes by built-in role:

| Role | Tenant | Area |
|---|---:|---:|
| `tenant_owner` | yes | no |
| `area_manager` | no | yes |
| `author` | yes | yes |
| `approver` | yes | yes |
| `viewer` | yes | yes |

Invalid assignments such as `tenant_owner @ area` and `area_manager @ tenant` must be structurally rejected.

The target model removes the magic `areaCode="tenant"` sentinel.

## D-AUTH-10 — Tenant owner is a normal role, never a bypass

**APPROVED.** `tenant_owner` receives every **tenant product permission** through normal RolePermissions.

A catalog invariant must ensure every tenant-owned product Permission is included in the built-in owner bundle. Adding a new tenant permission without assigning it to `tenant_owner` must fail a mechanical proof/gate.

Future platform/control-plane permissions are a different catalog/domain and must never auto-enter `tenant_owner`.

Delete target patterns such as:

```text
if role == tenant_owner { return true }
```

and historical `system_admin → AllCapabilities()` bypass behavior.

`Explain()` must show ordinary evidence: assignment → role → permission → scope, not “admin bypass”.

## D-AUTH-11 — Domain Constraints outrank every role

**APPROVED.** No role, including `tenant_owner`, implicitly bypasses domain invariants.

Examples:

- SoD may block approving one's own work;
- published/frozen content may not be directly edited;
- mandatory stages may not be skipped;
- a user not assigned to the stage cannot sign merely because they are owner;
- terminal-state mutations remain illegal;
- cross-tenant access remains illegal;
- required reauthentication/signature ceremony remains mandatory.

Administrative operations, when needed, are explicit operations with explicit Permissions/rules/audit. A future break-glass path must be explicit, temporary, reason-required and audited; it is not `tenant_owner`.

## D-AUTH-12 — User provisioning requires initial access assignment

**APPROVED.** The normal product operation for creating an **active** user requires at least one valid initial Role Assignment.

The required input is not literally `area + role`; it is one or more valid `role + scope` assignments. Area is required only for an `AreaScope` assignment.

Role-definition metadata determines which scopes are legal. User creation + tenant membership + initial Role Assignment(s) occur atomically. If any assignment is invalid, the whole provisioning operation rolls back.

Authorization itself remains `default deny`.

## D-AUTH-13 — Existing active user may have zero assignments

**APPROVED.** An existing user may remain an active identity with zero effective Role Assignments.

```text
identity status ≠ authorization state
```

An active user with zero grants receives `DENY` for protected actions. Removing the final assignment does not require deactivating the identity.

## D-AUTH-14 — Role Assignment lifecycle is historical

**APPROVED.** Role Assignments are not normally hard-deleted. They preserve grant provenance and optional temporal validity.

Target attributes include conceptually:

```text
id
tenant_id
subject_type   (user | group)
subject_id
role_id
scope_type     (tenant | area)
scope_id
valid_from
valid_until?
granted_by
granted_at
revoked_by?
revoked_at?
revocation_reason?
```

Effective assignment predicate:

```text
revoked_at IS NULL
AND valid_from <= now
AND (valid_until IS NULL OR valid_until > now)
```

Revocation is auditable and reason-bearing. Temporary grants are supported by the model even if V1 UI exposes only a simple subset initially.

## D-AUTH-15 — Groups are flat in V1

**APPROVED.** No nested groups in V1.

A user may belong to multiple groups. Group nesting, recursion, cycle prevention and transitive explanation are deferred until a concrete requirement justifies them.

## D-AUTH-16 — Resource Relationships are domain-created, not manual ACLs

**APPROVED.** V1 has no general per-document sharing/ACL UI.

Resource Relationships exist only where the domain has a genuine case-specific relation, for example:

- approval-stage candidate/assignee;
- review assignment;
- distribution recipient;
- submitter relation when SoD/withdrawal logic depends on it.

RBAC supplies broad organizational authority. Relationships narrow or establish case participation. Domain constraints then decide whether the action is legal now.

## D-AUTH-17 — Permission Catalog is product-owned

**APPROVED.** Permissions are defined by the MetalDocs product, versioned and not tenant-created.

Tenants may eventually create custom roles by combining existing Permissions, but cannot invent arbitrary Permission identifiers.

Naming convention:

```text
<resource>.<action>
```

Permissions should be atomic authorities. Roles provide bundles; avoid vague `*.manage` only when it would hide independently grantable authorities.

Canonical metadata should include at least semantic code/description and supported scope classes. Exact representation remains a design detail.

## D-AUTH-18 — Capability count is not a target

**APPROVED.** The current 39 capabilities are historical input, not a target count.

For every existing capability and every protected operation:

- keep if it represents a real independent authority;
- rename when vocabulary is wrong;
- merge/delete when the distinction is only module/history shape;
- add a new Permission when a sensitive independent authority is currently hidden behind a broader one.

Criterion:

> **Would a professional administrator legitimately want to grant/deny this authority independently?**

## D-AUTH-19 — Tenant permanent deletion is not a normal tenant Permission

**APPROVED.** Delete `tenant.erase` from the tenant RBAC target.

`tenant_owner` may initiate:

```text
tenant.deletion.request
```

but permanent erasure is a lifecycle transition executed by the platform/system only after explicit ceremony, preflight, a cooling-off/grace period, and revalidation.

Target lifecycle concept:

```text
ACTIVE
  ↓ tenant_owner requests deletion + reauth + explicit confirmation + reason
PENDING_DELETION
  ↓ cancellation window / preflight / grace period
ELIGIBLE_FOR_ERASURE
  ↓ system lifecycle executor
PERMANENT_ERASURE / crypto-shred
```

Permanent erasure is not `tenant_owner` bypass and not a button mapped to a destructive Permission.

A future platform break-glass/manual execution path, if ever required, must be a separate control-plane mechanism with reason/audit and potentially dual control when staffing makes that meaningful.

## D-AUTH-20 — Area manager is operational, not structural

**APPROVED.** Chosen option A: `area_manager` operates work within an area but does not change taxonomy, approval policy/route definitions, template-use policy, or access-control configuration.

## D-AUTH-21 — Approver does not receive broad working-document visibility

**APPROVED.** `approver @ area` does **not** automatically receive `document.read_working` for every draft in that area.

Working-content visibility is established when a legitimate workflow/resource relation connects the actor to the case. Otherwise the approver has only the normal published visibility granted by its role/scope.

Example:

```text
approver @ QUALIDADE
DOC-A draft, unrelated            → no working read
DOC-B under review, actor eligible → working read for case
DOC-C published                    → normal published read
```

## D-AUTH-22 — Own withdrawal differs from administrative cancellation

**APPROVED.** An author may withdraw only their own submission when workflow/domain state permits it. This is derived from:

```text
document.submit
+ submitted_by relationship
+ domain constraints
```

It does not require giving authors a general `approval.cancel` permission.

Administrative cancellation is a distinct authority for `area_manager` in its area and `tenant_owner` in the tenant, always subject to domain constraints, required reason and audit.

## D-AUTH-23 — Approval permissions are frozen pending workflow redesign

**APPROVED PROCESS DECISION.** Do not finalize the `approval.*` Permission set yet.

Candidate names such as:

- `approval.review`
- `approval.signoff`
- `approval.oversee`
- `approval.cancel`
- `approval.extend_deadline`

are **HOLD**, not final. They may survive, merge, split or be renamed after the greenfield approval/workflow model is validated.

The Permission Catalog must be derived from real human-work authorities, not from current endpoints or historical capability strings.

## D-AUTH-24 — Authorization + Approval/Workflow are co-designed

**APPROVED.** The design scope is expanded before implementation.

We will jointly design:

```text
Authorization
+
Human Approval / Workflow
+
Domain Governance
```

They are designed together because their semantics interact, but they remain different authorities/modules/boundaries.

Do not create one giant “IAM workflow” subsystem.

---

# 5. Candidate Permission Catalog — current working state

This catalog is **not final**. Non-approval entries are stronger candidates; all approval/workflow-related entries remain subject to the Approval Architecture Census.

## Controlled Information / Documents

| Candidate | State | Rationale |
|---|---|---|
| `document.read_published` | strong candidate | Required to distinguish viewer/public controlled information from internal working content. |
| `document.read_working` | strong candidate | Broad working-content visibility for authors/managers; not granted broadly to approvers. |
| `document.create` | strong candidate | Independent authority. |
| `document.edit` | strong candidate | Independent from create/submit. |
| `document.comment` | strong candidate | Review/comment collaboration does not imply document-body editing. |
| `document.submit` | strong candidate | Governed lifecycle transition. |
| `document.review_periodic` | strong candidate | Periodic eQMS review is distinct from workflow review. |
| `document.obsolete` | strong candidate | Governed lifecycle authority. |
| `document.supersede` | strong candidate | Successor/supersession is independently meaningful. |

## Approval / Workflow — HOLD

| Candidate | State |
|---|---|
| `approval.review` | HOLD — derive from final human-task model |
| `approval.signoff` | HOLD — derive from final human-task model |
| `approval.oversee` | HOLD — oversight semantics need workflow/visibility model |
| `approval.cancel` | HOLD — administrative workflow authority likely real, exact model pending |
| `approval.extend_deadline` | HOLD — exact deadline/escalation authority pending |
| `approval_route.read` | HOLD — may become approval-policy vocabulary rather than “route” |
| `approval_route.manage` | HOLD — may become approval-policy vocabulary rather than “route” |

## Template-use policy

| Candidate | State | Rationale |
|---|---|---|
| `template_policy.manage` | HOLD/strong candidate | ADR 0093 says template is a relationship/policy on released Controlled Information, not a peer aggregate. Exact vocabulary should align with the A9 target context. |

## Taxonomy

- `taxonomy.read`
- `taxonomy.manage`

## Identity / Groups / RBAC

- `user.read`
- `user.manage`
- `group.read`
- `group.manage`
- `role.read`
- `role_assignment.read`
- `role_assignment.manage`

No `role.manage` in V1 because custom-role mutation is intentionally YAGNI for V1.

## Governance / Security / Sessions / Analytics

- `audit.read`
- `security.read`
- `session.read`
- `session.revoke`
- `analytics.read`

Operational Prometheus `/metrics` should not be forced through normal tenant RBAC; its eventual infrastructure protection is an ops/control-plane concern.

## Distribution / Notifications / Tokens

- `distribution.read`
- `token_dictionary.read`
- `token_dictionary.manage`

`notification.read` is not currently favored as a tenant RBAC Permission. Own-notification access is more naturally modeled as authenticated self/resource relationship (`recipient == subject`).

## Tenant lifecycle

- `tenant.export`
- `tenant.deletion.request`

`tenant.onboard` belongs to future platform/control-plane authority because no normal tenant actor can authorize creation of a tenant that does not yet exist.

`tenant.erase` is rejected from tenant RBAC.

---

# 6. Current 39-capability census — working disposition

The following is the current redesign mapping. Approval rows remain provisional until the approval/workflow synthesis finishes.

| Current capability | Target disposition |
|---|---|
| `document.view` | split → `document.read_published` + `document.read_working` |
| `document.create` | keep |
| `document.edit` | keep |
| `document.submit` | keep |
| `document.signoff` | move out of Document; exact Approval target on HOLD |
| `document.obsolete` | keep |
| `document.supersede` | keep |
| `document.review` | rename → `document.review_periodic` |
| `approval.review` | HOLD |
| `approval.oversee` | HOLD |
| `approval.sla_extend` | HOLD; likely deadline-management vocabulary |
| `template.view` | merge into Controlled Information read semantics |
| `template.create` | merge into Controlled Information create semantics |
| `template.edit` | merge into Controlled Information edit semantics |
| `template.submit` | merge into Controlled Information submit semantics |
| `template.approve` | merge into subject-generic approval semantics; exact Permission HOLD |
| `template.publish` | retire; release is system/domain driven |
| `template.archive` | merge into Controlled Information lifecycle semantics |
| `template.manage` | re-home to template-use/approval-policy governance; exact target pending |
| `controlled_documents.create` | merge → `document.create` |
| `controlled_documents.obsolete` | merge → `document.obsolete` |
| `controlled_documents.supersede` | merge → `document.supersede` |
| `taxonomy.view` | rename → `taxonomy.read` |
| `taxonomy.manage` | keep |
| `membership.view` | retire; replaced by explicit group/role-assignment reads |
| `membership.manage` | retire; replaced by group/role-assignment management |
| `route.manage` | HOLD; likely approval-policy read/manage split |
| `user.view` | split/rename → `user.read` plus security/session authorities |
| `user.manage` | keep |
| `metrics.view` | re-home → `analytics.read`; infra metrics separated |
| `audit.read` | keep |
| `session.manage` | split → `session.read` + `session.revoke` |
| `distribution.read` | keep |
| `notification.read` | retire from RBAC; self/resource relationship |
| `token.view` | rename → `token_dictionary.read` |
| `token_dictionary.manage` | keep |
| `tenant.onboard` | move to future platform/control-plane catalog |
| `tenant.export` | keep |
| `tenant.erase` | retire from tenant RBAC → `tenant.deletion.request` lifecycle initiation |

Additional candidate discovered by functional census:

- `document.comment` — collaboration authority independent of body editing.

---

# 7. Approval/Workflow redesign — approved scope expansion

## 7.1 Core design question

Do not ask first:

> Which Permission protects this approval endpoint?

Ask first:

> What is the correct human-workflow/approval model, and which independent authorities exist inside it?

Only after that derive Permissions.

## 7.2 Five conceptual authorities to keep separate

### A. Authorization

Answers:

```text
Can(subject, permission, scope/resource)?
```

Knows roles, permissions, assignments, groups and scopes. It must not need knowledge of stage numbers, quorum, next approver, SLA, workflow status or approval-route internals.

### B. Approval Policy / Workflow Definition

Answers:

```text
How must this kind of controlled information be reviewed/approved?
```

Candidate conceptual shape, not yet final:

```text
ApprovalPolicy
  id
  version
  applicability
  stages[]

Stage
  kind
  actor_selection_policy
  completion_policy
  deadline/escalation policy
  order
```

### C. Actor Selection / Human Work

Answers:

```text
Who should/can perform this specific human task?
```

Potential actor-selection sources to evaluate against real needs:

- specific user;
- group;
- role within scope;
- area manager;
- dynamic relationship;
- submit-time choice;
- future organizational/manager relation if a real requirement appears.

**Actor selection never grants product Permission.** A candidate must still satisfy authorization/domain requirements.

### D. Approval Instance / Evidence

At submission, a versioned policy is instantiated and the evidence-relevant facts are frozen/snapshotted as required.

Conceptual target:

```text
ApprovalInstance
  policy_definition_version
  immutable subject/submission reference
  stage instances
  resolved candidate/assignee facts as appropriate
  decisions / comments / signoffs
  delegation evidence
  timestamps / reasons / signatures / hashes
```

Changing the policy later must not rewrite the semantics/evidence of an already-started approval.

### E. Domain Governance

Owns legality that cannot be reduced to RBAC or workflow candidate lists:

- Separation of Duties;
- author/self-approval restrictions;
- reauthentication/signature requirements;
- frozen submission/hash binding;
- legal state transitions;
- withdrawal rules;
- administrative cancellation legality;
- deadline-extension constraints;
- delegation constraints;
- terminal-state invariants;
- release prerequisites.

## 7.3 Release remains downstream of approval evidence

Preserve until contrary material finding:

```text
Authoring
→ Submission freeze
→ Approval workflow / evidence
→ approval complete
→ release prerequisites
→ Release Coordinator
→ published/effective Controlled Information
```

The previously retired human `document.publish` capability should not be resurrected merely because AuthZ is being redesigned.

---

# 8. Approval Architecture Census — required before finalizing `approval.*`

The redesign must independently evaluate at least these axes:

| Axis | Required question |
|---|---|
| Definition | What exactly is an approval “route” / policy / workflow definition? |
| Versioning | How are definitions versioned and how do in-flight instances bind to a version? |
| Applicability | How is the correct policy selected for a subject/profile/governance class? |
| Stage semantics | Are `review` and `approval` genuinely different human-task types? |
| Ordering | Sequential only? Parallel stages? Branching? What is truly required now? |
| Completion | one-of, all-of, quorum, first-response, unanimous? |
| Actor selection | user, group, role-in-scope, relation, submit-time selection? |
| Resolution time | policy configuration, submission time, or stage activation? |
| Snapshot semantics | Which actor/policy facts freeze and which remain dynamic? |
| Authorization composition | Which base Permission/qualification must a candidate also possess? |
| Visibility | Participant, author, manager/overseer, viewer behavior. |
| SoD | author, prior approvers, same-person multi-stage behavior, delegation. |
| Delegation | who may delegate, window, revocation, on-behalf-of evidence. |
| Deadlines | due dates, extension, escalation, overdue surfacing. |
| Cancellation | self-withdraw vs administrative cancel. |
| Evidence | identity, reason, timestamps, decision, policy version, subject hash/snapshot. |
| Reauth/signature | which task types require ceremony and what is bound by the signature. |
| Policy mutation | whether active definitions are mutable, versioned, immutable after use, deactivate-only, etc. |
| Empty workflow | when zero stages is legitimate and what “automatic approval” means. |
| Release | exact relationship between approval completion and publication/effectivity. |
| Audit | what must be reconstructible years later without consulting mutable current config. |

The census must compare current MetalDocs behavior with mature workflow/eQMS patterns, not use the current schema/module shape as its own justification.

---

# 9. Current implementation facts that must not be mistaken for target truth

The current repository includes, among other things:

- `Capability` registry with 39 current capability constants;
- hard-coded role enum including `approver`, `area_admin`, `author`, `editor`, `qms_admin`, `signer`, `system_admin`, `viewer`;
- Tier-1 route resolution from user/group role relationships;
- Tier-2 authorization from area membership + role-capability relationships;
- explicit `system_admin` bypass in multiple paths;
- `CapabilityScope` classified as tenant-grade vs area-grade;
- magic `"tenant"` sentinel used to suppress area filtering;
- DB asserted-capability tripwires;
- approval policy/route definitions with stages;
- `approval_route_stage_selectors` introduced in migration 0303;
- selector snapshotting into stage instances introduced in migration 0304;
- old flat stage actor columns removed in migration 0305;
- approval subject generalization and subject-generic kernel work;
- route version/snapshot/evidence machinery;
- SoD, delegation, deadlines, SLA extension and release-coordinator behavior;
- ADR 0093's accepted design ruling that Documents, Controlled Documents and Templates target one Controlled Information bounded context.

These are evidence to inspect. They are **not automatically premises** of the target design.

---

# 10. OSS / build-vs-buy position already established

Current direction remains:

- **Keycloak** — future authentication/SSO/federation candidate; not canonical MetalDocs domain authorization.
- **OpenFGA** — preferred future externalization candidate if authorization becomes multi-service / relationship-heavy; do not introduce now solely for current monolith.
- **SpiceDB** — strong alternative when distributed consistency/platform scale justify it; not currently warranted.
- **Cerbos / Cedar / OPA / Casbin / Oso** — useful policy technologies, but none currently removes the need to model MetalDocs scopes, transactional authority, resource relationships, filtering and domain constraints correctly.

Target now: internal Postgres-backed authorizer behind an engine-neutral semantic boundary, conceptually:

```text
Authorizer.Check(...)
Authorizer.Filter(...)
Authorizer.Explain(...)
```

Do not over-generalize exact interface/data types until the semantics are fully approved.

---

# 11. Golden Matrix — evolving acceptance model

Minimum cases already required:

```text
user role @ tenant
user role @ area
group role @ tenant
group role @ area
multiple simultaneous roles
direct + group role union
future assignment
expired assignment
revoked assignment
duplicate/idempotent assignment
wrong tenant
wrong area
invalid role/scope combination
area deleted/inactive
published-only viewer access
working-content author access
approver without workflow relation cannot read unrelated working document
approver with legitimate workflow relation can read required working subject
resource relationship present
resource relationship absent
submitter withdrawal of own submission
non-submitter cannot use self-withdraw path
administrative cancellation by area_manager in area
administrative cancellation outside scope denied
SoD: author trying own approval
owner/admin still blocked by Domain Constraint
policy version changes do not rewrite in-flight/historical instance semantics
participant visibility vs oversight visibility
tenant deletion request vs permanent-erasure executor
platform operator without tenant grant
future break-glass support session
```

The last two are architectural reserved scenarios, not V1 features.

---

# 12. Known impact surfaces for eventual implementation map

At minimum the final design must account for:

## Backend / IAM / authorization

- `internal/modules/iam/domain/model.go`
- `internal/platform/iamtypes/role.go`
- `internal/modules/iam/domain/catalog.go`
- `internal/modules/iam/domain/capability_scope.go`
- `internal/modules/iam/application/capability_service.go`
- `internal/modules/iam/authz/authz.go`
- `internal/modules/iam/authz/bypass_audit.go`
- IAM persistence/repositories for role capabilities, user roles, group roles, group membership and area memberships
- composition-root HTTP permission resolution
- DB tripwire/asserted-capability integration

## Approval / workflow

- `internal/modules/approval/domain/route.go`
- actor selectors and selector resolution
- route/policy administration service
- submit services
- stage-instance creation and snapshots
- decision/review/signoff services
- delegation
- SoD
- cancellation / withdrawal
- deadlines/SLA/accountability
- approval read/inbox/visibility
- release-coordinator integration

## Contract / frontend

- OpenAPI `x-authz-*` metadata and IAM/approval contracts
- generated backend/frontend types
- IAM admin center and role vocabulary
- route-admin / approval-policy UI
- approval inbox/workspace visibility
- capability/permission hooks and presenters

## Persistence / reference data

- `role_capabilities`
- `iam_user_roles`
- `iam_group_roles`
- `iam_group_members`
- `user_process_areas`
- approval route/policy/stage/selectors/instance tables
- reference-data role/capability seeds
- RLS and tripwire functions

## Docs / guards / tests

- authz ADRs (0007, 0016, 0019, 0021, 0022, 0092 and related history)
- approval ADRs (0082, 0084, 0085, 0087 and others found by census)
- IAM/approval module docs
- `.claude/skills/developing-new-work/references/capability-wiring.md`
- API lint/authz surface guards
- role/capability registry tests
- integration tests for selectors, snapshots, SoD, visibility and tripwires

Historical migrations/ADRs are evidence. They should not be mass-renamed merely for terminology unless a live verifier treats historical text as current authority. Current canonical docs/code should converge on `Permission` once the redesign lands.

---

# 13. Open decisions — do not assume

The following remain intentionally unresolved:

1. Final approval/workflow vocabulary (`Route` vs `ApprovalPolicy` vs another domain term).
2. Exact human-task kinds: whether `review` and `signoff/approve` remain distinct primitives.
3. Exact completion modes needed in V1: first-response, all, quorum, unanimous, etc.
4. Actor-selector vocabulary and which selector types are genuinely required now.
5. Actor resolution timing and snapshot semantics.
6. Policy versioning/mutation rules.
7. Exact oversight/visibility model and whether `approval.oversee` survives as-is.
8. Exact approval-related Permission catalog after workflow synthesis.
9. Background/system/service-principal authorization semantics; current scheduler bypass must be redesigned independently from tenant owner.
10. Final authorization domain/module placement (`iam` vs a dedicated authorization module) after boundary census.
11. Exact `Authorizer.Check/Filter/Explain` types and transaction semantics.
12. Exact filtering/query-plan integration ensuring `Filter` and `Check` cannot disagree.
13. Final DB asserted-permission/tripwire posture after the single evaluator exists.
14. Custom-role persistence/lifecycle details beyond the V1 no-UI/no-API decision.
15. Exact tenant deletion grace-period duration and pending-deletion runtime behavior (read-only/suspended/etc.).

Any of these must be resolved explicitly, one decision at a time, before implementation.

---

# 14. Next design sequence

```text
1. Greenfield Approval/Workflow research + current implementation census
2. Define target Approval Policy / Human Task / Instance / Evidence model
3. Validate actor selection, completion, versioning, snapshot, SoD, delegation, deadlines, cancellation and release
4. Derive final approval-related Permissions from that target
5. Complete built-in Role → Permission matrix
6. Define Authorization evaluator semantics and data model
7. Define exact Authorization ↔ Workflow ↔ Domain integration contract
8. Complete code/schema/API/frontend/docs impact map
9. Present integrated design section-by-section for operator approval
10. Write final design spec / ADR promotion
11. Self-review for contradictions/placeholders/ambiguity
12. Operator reviews written spec
13. Only then write implementation plan
```

No implementation is authorized by this ledger.
