# Current Agent Handoff

> **Last verified:** 2026-08-14
> **Status:** ACTIVE — Cohesive Platform Redesign
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Read this first

A fresh session MUST read:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. this file

Do not resume old roadmaps/milestones/specs, deleted `docs/superpowers` material, PR #113 or historical implementation shapes by inertia.

## Where we are

MetalDocs is being redesigned as one cohesive controlled-information platform before the next implementation wave.

Locked so far:

### Authentication / Organization / Authorization

- current MetalDocs AuthN remains sufficient for V1; Keycloak/external IdP is future-triggered;
- Organization owns Tenant/Area/User/Group/GroupMembership;
- Groups are flat and receive ordinary RoleAssignments;
- built-in roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`;
- scoped RBAC + Groups is enough for V1; no OpenFGA/SpiceDB now;
- tenant owner is never an authorization/domain-governance bypass.

### Approval V1

- specialized sequential ApprovalPolicy, not BPM;
- ApprovalSteps use NamedUser / Group / RoleInArea and ANY/ALL completion;
- human outcomes `accept` / `return_for_changes`;
- reassignment is explicit/audited; generic delegation/escalation/M-of-N/CEL/BPMN are deferred;
- ApprovalInstance binds one immutable RevisionSubmission;
- changed bytes/resubmission create a new RevisionSubmission and new ApprovalInstance.

### Controlled Information configuration (R3)

- `documents + controlleddocuments + templates` do not survive as three target business contexts;
- core nouns are `Document`, `DocumentRevision`, `RevisionSubmission`;
- `ControlledDocument` is retired as separate identity;
- `DocumentProfile` → tenant-scoped `DocumentType`, immutable code, ACTIVE/INACTIVE, no own versioning;
- Document type is immutable after creation in V1;
- `DocumentFamily` → optional classification-only `DocumentTypeCategory`;
- `GovernanceClass {controlado, simples, livre}` is deleted;
- Approval configuration explicitly distinguishes `NoHumanApproval` from `UsePolicy(...)`;
- Area belongs to Organization;
- template is a role of a governed Document, not a parallel lifecycle;
- TemplateUse is M:N across template Documents and DocumentTypes; at most one default per type; default is UX only;
- blank creation remains allowed until a real `template_required` requirement exists;
- template creation resolves current effective source REV and permanently pins source Document + REV + digest;
- newer template revisions affect future creations only; existing Documents never rebind;
- changing template layout/placeholders/schema/constraints/visibility/resolver semantics means a new ordinary REV.

### Document / Revision / Submission lifecycle (R4)

- official business revisions are `REV001`, `REV002`, `REV003`, ... — never user-facing `v7`;
- technical row/OCC/schema/policy versions are separate namespaces;
- `Document` is stable identity, not `draft/approved/published` workflow state;
- one Document may have one effective REV and at most one open REV;
- `DocumentRevision` is the business change cycle, not an autosave/check-in;
- REV states: `DRAFT`, `SUBMITTED`, `EFFECTIVE`, `SUPERSEDED`, `OBSOLETE`, `CANCELLED`;
- a new REV is allocated when the change cycle begins and labels are never reused;
- return-for-changes/withdraw returns the **same REV** to DRAFT;
- each submit/resubmit creates an immutable `RevisionSubmission` attempt;
- `RevisionSubmission` exists even with `NoHumanApproval`;
- Approval, official Rendition and Release must bind the same exact Submission;
- no fake zero-stage ApprovalInstance for no-human-approval;
- reason-for-change belongs to the REV and each Submission snapshots the reason accompanying its bytes;
- autosaves/checkpoints/edit history are technical authoring history inside the open REV;
- EditorSession edits the open DRAFT REV and cannot mutate it after SUBMITTED;
- prior effective REV becomes SUPERSEDED mechanically when the new REV becomes EFFECTIVE;
- OBSOLETE = retire the Document without a successor; retired Document is terminal in V1;
- cross-Document replacement is a separate future design question, not ordinary revision supersession.

### Release

Human approval does not directly effectivate. Release Coordinator remains downstream and will later close approval/rendition/effective-date/supersession gates atomically.

## Important product evidence

A real browser QA run proved the old model could show one editor-authored body to the approver while freeze rendered a blank template snapshot. The final signed PDF/hash did not represent what was reviewed. The target therefore requires Approval, Rendition and Release to share one immutable RevisionSubmission identity.

## Remaining whole-product coverage

Before code we still must close:

- **NEXT R5:** numbering/NumberSeries + TemplateSpec exact revision payload + metadata ownership;
- periodic review/reason-for-change operational policy;
- rendition/rendering/reconstruction evidence;
- release/effectivity + possible cross-Document replacement;
- distribution/read/acknowledgement;
- token/computed-value snapshot timing;
- audit/evidence boundary;
- notifications/search projections;
- tenant lifecycle/security/external IdP trigger;
- final Permission catalog + role bundles;
- bounded contexts, table/transaction ownership, event contracts, data model, OpenAPI and frontend journeys;
- explicit migration/delete/rename map;
- final ADR/spec set and implementation plan.

## Exact next step

Continue **R5 — Numbering + Template authoring payload + metadata placement**:

```text
Document business code / NumberSeries
+ DocumentType/Area relationship to numbering
+ allocation timing and code immutability
+ manual-override question
+ TemplateSpec exact governed payload
+ DOCX/body + placeholder/schema representation
+ template provenance storage boundary
+ newer-template availability without rebinding
+ which metadata belongs to Document vs REV vs Submission
```

Apply Global Maximum/YAGNI and compare with mature DMS/eQMS products where useful. Do not implement.

## Documentation reset

`docs/superpowers` is intentionally limited to active redesign staging material. Git history is the archive for previous plans/specs/milestones/reports. Legacy wiki/module pages are current-state evidence only and cannot override the active redesign.