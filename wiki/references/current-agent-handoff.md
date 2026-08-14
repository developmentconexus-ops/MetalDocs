# Current Agent Handoff

> **Last verified:** 2026-08-14
> **Status:** ACTIVE — Cohesive Platform Redesign
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Read order

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. this file

Do not resume old roadmaps/milestones/specs, deleted `docs/superpowers` material or historical implementation PRs by inertia.

## Where we are

MetalDocs is being redesigned as one cohesive controlled-information platform before the next implementation wave.

Locked:

- current AuthN retained for V1 behind a future external-IdP seam;
- Organization = Tenant / Area / User / flat Group / GroupMembership;
- scoped RoleAssignments for User or Group;
- roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`;
- scoped RBAC + Groups is enough V1; no OpenFGA/SpiceDB now;
- Approval V1 = versioned sequential policy, ordered human Steps, NamedUser/Group/RoleInArea participants, ANY/ALL only, `accept` / `return_for_changes`, audited reassignment and optional reauth;
- no BPMN/CEL/M-of-N/generic delegation engine V1;
- `documents` + `controlleddocuments` + `templates` do not survive as three target contexts;
- stable core = `Document` + business `DocumentRevision`;
- `DocumentProfile` → `DocumentType`;
- `DocumentFamily` → classification-only `DocumentTypeCategory`;
- `GovernanceClass` deleted;
- Approval configuration explicitly distinguishes `NoHumanApproval` from `UsePolicy(...)`;
- Area belongs to Organization;
- template is a role of a governed Document; its changes use ordinary DocumentRevisions;
- TemplateUse is M:N across template Documents and DocumentTypes, with at most one UX default/type;
- source template's current EFFECTIVE Revision is resolved only at creation and exact origin is pinned forever;
- official revisions are `REV001`, `REV002`, ...; autosaves/checkpoints are technical history, not business Revisions;
- at most one open Revision + one effective Revision per Document V1;
- Revision states: DRAFT / SUBMITTED / EFFECTIVE / SUPERSEDED / OBSOLETE / CANCELLED;
- `RevisionSubmission` is immutable first-class evidence; every resubmission creates a new Submission and, when applicable, new ApprovalInstance;
- return-for-changes/withdraw can return the **same REV** to DRAFT;
- Approval/Rendition/Release always bind exact Submission/digest;
- Document code is immutable tenant-wide business identity; type/area/code immutable V1;
- numbering belongs to DocumentType and is deliberately small: literals + `{TYPE}` / `{AREA}` / `{SEQ}`, scope TYPE or TYPE_AREA, padding width; no formulas/scripts/date resets V1;
- normal Create has no manual code override; legacy-code preservation belongs to explicit import/migration;
- template `MetadataSchema` policy bundle is deleted; TemplateSpec only owns authoring/fill contract;
- TemplateField separates value type (`text/date/number/choice/user/image`) from source (`user_input/system/dictionary`), with typed constraints and limited `visible_if`;
- no `CompositionJSON` V1 without proven need;
- DOCX/content anchors and TemplateSpec must pass parity before template submission;
- `RevisionContent` is the logical governed content identity; Submission digest covers all content/governed metadata needed to reproduce what was submitted;
- `title` is Revision metadata; stable identity lives on Document; operational owner may exist on Document but grants no authorization;
- no generic tenant custom-metadata engine V1;
- Release/effectivity remains downstream of human Approval.

## Critical product invariant

A real QA failure proved the previous architecture could let a human review one content body while freeze rendered another (the blank template snapshot). The target must make this impossible by construction:

```text
RevisionSubmission
    ├── Approval
    ├── Renditions
    └── Release
```

all refer to the same immutable submission identity/digest.

## Exact next step — R6

Design only:

```text
Periodic Review
+ Renditions / Rendering
+ Release / Effectivity
```

Close:

- review policy/configuration and due-date semantics;
- reviewed-no-change evidence vs new-REV path;
- whether overdue review affects legal effectivity;
- review actor/permission semantics;
- source DOCX vs final DOCX vs official PDF;
- `Rendition` ownership/status/provenance;
- renderer/version evidence and reconstruction guarantees;
- retry/idempotency behavior;
- which artifacts are mandatory for Release;
- planned effective date;
- atomic effectivity + prior-REV supersession;
- cancellation after Approval but before Release;
- whether cross-Document replacement is a real independent requirement.

After R6: R7 Distribution/Tokens/Audit/Notifications/Search; R8 Tenant lifecycle/Security; R9 final Authorization matrix; only then technical architecture/data/API/frontend/migration specs and implementation plan.

## Documentation rule

The working tree contains active truth, not an archive. Historical staging remains in Git history. The active detailed ledger is the sole current WIP decision source. Do not restore deleted historical plans into the live tree.
