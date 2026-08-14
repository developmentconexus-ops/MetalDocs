# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **NO PRODUCT IMPLEMENTATION AUTHORIZED**
> **Established:** 2026-08-14
> **Repository baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** [`../standards/root-cause-global-maximum-method.md`](../standards/root-cause-global-maximum-method.md)
> **Detailed working ledger:** [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md)

## 1. Purpose

MetalDocs is being redesigned as one coherent product before the next large implementation wave.

Authentication, IAM, areas, approval routes, Documents, Controlled Documents, Templates, taxonomy, rendering and release evolved incrementally and now overlap in authority. The current code remains valuable evidence about real requirements, failures and operational constraints, but **current implementation shape is not admissible as proof that the target should keep the same nouns, modules or boundaries**.

The target must be the smallest professional architecture that:

- represents the real controlled-information domain correctly;
- gives every business fact one authority;
- preserves multi-tenancy, auditability, immutable evidence and fail-closed authorization;
- deletes duplicated lifecycle/policy implementations;
- avoids speculative BPM, ReBAC, policy languages or external identity infrastructure;
- has explicit extension triggers for future enterprise requirements;
- is specified end-to-end before product implementation begins.

## 2. Fresh-session reading order

Any new session working on product/domain architecture MUST read:

1. [`../../AGENTS.md`](../../AGENTS.md)
2. [`../standards/root-cause-global-maximum-method.md`](../standards/root-cause-global-maximum-method.md)
3. **this file**
4. [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md)
5. [`../references/current-agent-handoff.md`](../references/current-agent-handoff.md)

Do not start an old roadmap unit, milestone, migration, implementation PR or deleted `docs/superpowers` plan by inertia.

## 3. Authority during the reset

For **target design** questions, authority is:

1. operator-approved decisions in the active ledger;
2. this page for program state/scope;
3. canonical cross-cutting standards;
4. final ADRs explicitly retained by this program;
5. runtime/schema/OpenAPI/module docs as evidence only;
6. historical plans/specs/ADRs as evidence only.

For **what runs today**, runtime/code/database and OpenAPI remain authoritative.

This distinction is deliberate: runtime truth answers what exists; it does not grant architectural legitimacy to the existing shape.

## 4. Approved north star

```text
Authentication
     │
     ▼
Organization ───────────── Authorization
Tenant                    Roles / Permissions
Areas                     Role Assignments
Users                     User + Group principals
Groups                    Tenant / Area scopes
     │                         │
     └──────────────┬──────────┘
                    ▼
          Controlled Information
          Document / Revision
          DocumentType
          Template-as-revision-role
          Numbering / lifecycle
                    │
                    ▼
                Approval
         versioned sequential policy
         human steps + participants
         decisions + evidence
                    │
                    ▼
           Domain Governance
       freeze / SoD / lifecycle rules
                    │
                    ▼
           Release Coordinator
                    │
                    ▼
             Effective Revision
```

Supporting concerns consume those authorities instead of redefining them:

```text
Audit / Evidence
Rendering / Renditions
Periodic Review
Distribution / Read-Acknowledge
Notifications
Search / Projections
Token Dictionary / Computed Values
Tenant Lifecycle / Security
Async orchestration / outbox / jobs
```

## 5. Decisions already locked

### Authentication

- AuthN and product AuthZ stay separate.
- Current MetalDocs authentication/session implementation is acceptable for V1.
- Keycloak/external IdP is a future adapter, not a current dependency.

### Organization + Authorization

- `Area` belongs to Organization, not document taxonomy.
- V1 organization: Tenant, Area, User, Group, GroupMembership.
- Groups are flat and may receive ordinary RoleAssignments.
- Built-in roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`.
- Roles are bundles; authorization checks semantic Permissions.
- RoleAssignment subject is User or Group; scope is typed Tenant or Area.
- Grants compose additively; default deny; no explicit deny engine in V1.
- `tenant_owner` is never a bypass.
- OpenFGA/SpiceDB are not required for V1.

### Approval V1

- Specialized governed-document approval, not generic BPM.
- Versioned `ApprovalPolicy` with ordered sequential `ApprovalStep`s.
- A Step is the human task; `review`/`approval` are purpose/meaning, not two engines.
- Initial actor rules: named user, group, role-in-area.
- Completion: ANY or ALL only.
- No BPMN, generic branching, phase graph, CEL, Camunda/Flowable, M-of-N or generic escalation engine in V1.
- Participants resolve on Step activation and are snapshotted; current authorization is rechecked when acting.
- Human outcomes: `accept`, `return_for_changes`.
- `return_for_changes` terminates the attempt; edited content is resubmitted as a new ApprovalInstance.
- `withdraw`, `cancel`, `reassign` are separate operations.
- Reauthentication may be required by a Step.
- Audited reassignment is V1; generic time-window delegation is deferred.

### Controlled Information

- `documents`, `controlleddocuments`, `templates` are not three target bounded contexts.
- Target core is `Document` + `DocumentRevision` inside one Controlled Information context.
- `ControlledDocument` as a separate third public/domain object is targeted for retirement.
- `DocumentProfile` converges toward `DocumentType` rather than remaining a bag of cross-domain configuration.
- Template has no independent lifecycle/version counter. It is a designation/role of an exact governed DocumentRevision.
- Changing DOCX/template body, placeholder schema, validation, visibility, resolver binding or other governed template semantics requires a new DocumentRevision.
- A derived document is permanently bound to the exact source template revision/hash used to seed it; later template revisions never rebind it.
- After creation, the new document's own Revision is the edited/reviewed content truth.
- Freeze, Approval evidence and official Rendition must bind the exact submitted/reviewed Revision/hash.
- The historical blank-PDF defect — approver reviews editor content while freeze renders another template snapshot — must become structurally impossible.
- Area moves to Organization.
- Family and GovernanceClass are still under independent-value review; neither survives merely because it exists today.

### Release

- Approval produces human decision evidence; it does not directly publish.
- Release/effectivity remains downstream and verifies mechanical/domain gates such as approval receipt, same revision/hash, required artifacts, effective date and supersession legality.
- Exactly one effective revision per Document is a core invariant.

## 6. Whole-product coverage

The redesign is not complete when AuthZ + Approval + Documents are complete. Before implementation, it must explicitly disposition:

- DocumentType / Family / GovernanceClass;
- Document and Revision lifecycles;
- numbering / NumberSeries;
- template designation and template payload semantics;
- periodic review and reason-for-change;
- renditions, rendering provenance, reconstruction/attestation;
- release/effectivity/supersession;
- distribution obligations/read/acknowledgement/reminders;
- tokens and computed values / snapshot timing;
- audit/evidence;
- notifications;
- search/projections;
- tenant lifecycle/security/external IdP trigger;
- final Permission Catalog + role bundles;
- bounded contexts/package DAG/table ownership/transactions;
- data model, events, APIs, frontend journeys and migration/deletion map.

The detailed checklist and exact next step live in the active ledger.

## 7. Current module disposition

Current module docs are **current-state evidence**, not target architecture. The core pages affected by the redesign are explicitly marked LEGACY/HISTORICAL in `wiki/modules/`.

High-level target disposition:

- `approval` → redesign to Approval V1;
- `auth` → retain V1 behind AuthN boundary;
- `iam` → conceptually split Organization + Authorization;
- `taxonomy` → dismantle: Area → Organization, Profile → DocumentType candidate, remaining classification re-evaluated;
- `controlleddocuments` → absorb legitimate identity/numbering responsibilities into Controlled Information;
- `documents` → become the Controlled Information core after responsibility cleanup;
- `templates` → retire parallel lifecycle; template becomes revision role/designation;
- `jobs` → orchestration/composition, not bounded context;
- `audit`, `render`, `search`, `notifications`, `distribution`, `tokens`, `security` → retained/re-evaluated supporting concerns; do not rewrite healthy boundaries without a material reason.

## 8. Build-vs-buy posture

Do not choose infrastructure before the domain proves the requirement.

Current rulings:

- Keycloak/OIDC: defer until enterprise identity requirement fires.
- OpenFGA/SpiceDB: defer until arbitrary resource-sharing/relationship graph justifies it.
- Camunda/Flowable/BPMN: not for document Approval V1.
- Temporal: not an Approval prerequisite.
- CEL/expression language: defer until typed configuration cannot represent a real policy requirement.

External libraries/frameworks will still be evaluated later for precise responsibilities after semantics close.

## 9. Documentation reset

The old accumulation under `docs/superpowers/` has been removed from the live tree. Git history is the archive.

Live planning authority is intentionally narrow:

```text
AGENTS.md
→ this architecture page
→ active redesign ledger
→ current-agent-handoff
```

Do not restore deleted historical planning artifacts into the live tree. Recover a historical version only for a specific evidence question.

## 10. Implementation gate

**Closed.** No product implementation until the active ledger's integrated-design checklist is complete, final durable ADR/spec material is promoted to `wiki/`, adversarial review has no material ambiguity, the operator approves the integrated design, and an implementation plan is authored from that accepted target.

## 11. Exact next step

Continue design with:

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ TemplateDesignation/default policy
```

For every noun/policy ask:

> Does it represent independent business meaning in the target product, or is it only historical encoding of another authority?

Only then proceed to the full Document/Revision lifecycle.
