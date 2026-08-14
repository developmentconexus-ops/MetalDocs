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

### Locked platform core

- current AuthN retained for V1 behind a future external-IdP seam;
- Organization = Tenant / Area / User / flat Group / GroupMembership;
- scoped RoleAssignments for User or Group;
- roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`;
- scoped RBAC + Groups sufficient V1; no OpenFGA/SpiceDB now;
- Approval V1 = versioned sequential policy, ordered Steps, NamedUser/Group/RoleInArea participants, ANY/ALL, `accept` / `return_for_changes`, audited reassignment and optional reauth;
- no BPMN/CEL/M-of-N/generic delegation engine V1;
- Controlled Information core = stable `Document` + business `DocumentRevision`; `controlleddocuments` and parallel `templates` lifecycle retire;
- `DocumentProfile` → `DocumentType`; `DocumentFamily` → classification-only `DocumentTypeCategory`; `GovernanceClass` deleted;
- official Revisions are `REV001`, `REV002`, ...; autosaves/checkpoints are technical history;
- one effective + at most one open Revision per Document; states: DRAFT / SUBMITTED / EFFECTIVE / SUPERSEDED / OBSOLETE / CANCELLED;
- `RevisionSubmission` = immutable first-class attempt identity; Approval/Rendition/Release always bind exact Submission/digest;
- return-for-changes/allowed withdrawal returns the same REV to DRAFT; resubmission creates a new Submission;
- Document code is immutable tenant-wide identity; type/area/code immutable V1;
- numbering belongs to DocumentType: literals + `{TYPE}` / `{AREA}` / `{SEQ}`, scope TYPE or TYPE_AREA, padding width only;
- Template is a role of governed Document; TemplateUse M:N; exact effective source REV is resolved only at creation and provenance pinned forever;
- TemplateSpec owns authoring/fill contract only; value type is separate from source (`user_input/system/dictionary`); typed constraints + limited `visible_if`; DOCX/schema parity required;
- `RevisionContent` is the logical governed content identity; Submission digest covers all governed content/metadata needed to attest the attempt;
- `title` is Revision metadata; responsible owner is operational Document metadata and grants no authorization;
- no generic tenant custom-metadata engine V1.

### Locked review / rendition / release

- Periodic Review belongs to Controlled Information: `Disabled | Every(n months)` on DocumentType;
- cadence starts at Effectivity and restarts after completed review; overdue does not automatically invalidate EFFECTIVE content;
- append-only `PeriodicReviewRecord` against exact effective REV with `confirmed_current | change_required`; change-required does not auto-create a REV;
- `Rendition` = immutable derived artifact of exact RevisionSubmission with output hash + renderer/build provenance;
- only `OFFICIAL_PDF` mandatory for Release V1; final DOCX optional/non-blocking;
- Approval approves Submission, not PDF bytes; official PDF may manifest Approval evidence;
- Release is automatic/system-owned; no human publish button;
- `RevisionSubmission` replaces `ReleaseGeneration` as required domain candidate identity;
- optional `ReleasePlan.not_before`; no SCHEDULED Revision state;
- actual `effective_at = released_at`; no silent retroactive Effectivity;
- winning release atomically makes candidate EFFECTIVE, prior REV SUPERSEDED, swaps pointers, records immutable `ReleaseRecord` and emits lifecycle events;
- candidate may be CANCELLED after Approval but before Release; evidence remains historical;
- legacy `effective_date` TemplateField is removed V1 because actual Effectivity is born after the mandatory pre-release PDF.

### Locked R7 — supporting services

- Distribution = controlled obligation/read acknowledgement, not Authorization and not Training/LMS;
- DistributionConfiguration lives on Document and targets `User | Group`; no Area target without a real UserAreaMembership concept;
- Release expands Groups to concrete per-user `DistributionAssignment`s; later group membership changes never rewrite history;
- post-release assignment to current effective REV is explicit/audited; one obligation per user/release; pending old-REV assignments become `superseded` when a new REV is effective;
- opening notification/viewing/downloading does not complete obligation; explicit immutable `AcknowledgementRecord` does; optional fresh reauth reuses AuthN assurance seam;
- Distribution never edits RoleAssignments; task-specific visibility is finalized in R9;
- System values are product-owned closed contracts; tenant Dictionary is mutable tenant-owned source data;
- System keys V1: `document_code`, `revision_label`, `revision_title`, `document_type_code`, `document_area_code`, `document_area_name`, `revision_created_by_name`;
- `approval_date`/approvers move to Approval manifestation in official PDF; actual `effective_date` stays outside TemplateSpec V1;
- mutable Dictionary/external values resolve and snapshot when a new REV is created; same-REV return/resubmit does not silently re-resolve;
- domain evidence (`ApprovalDecision`, `ReleaseRecord`, `AcknowledgementRecord`, etc.) remains authority; Audit Trail never substitutes for it;
- critical governed mutations require durable audit intent in the same commit boundary; exact same-tx vs transactional-outbox implementation is R10;
- Audit Trail remains append-only, tamper-evident, exportable, with User/System actors; view/download telemetry may be async and never equals acknowledgement;
- Notifications are projection/delivery only; notification `READ` means the notification was read; “Minhas Pendências” comes from business authorities, not unread notifications;
- Search is rebuildable/eventually-consistent projection: Official Library = effective REV, optional Working Search = open REV under current AuthZ; historical superseded/obsolete Revisions not in global search by default;
- stale Search result never grants access; canonical endpoint rechecks current AuthZ; no Elasticsearch/OpenSearch requirement yet.

## Critical invariant

```text
RevisionSubmission
    ├── Approval
    ├── Rendition(s)
    └── Release
```

All three refer to the same immutable submission identity/digest. Notifications/Search/Audit never become a competing source of these facts.

## Exact next step — R8

Design only:

```text
Tenant Lifecycle
+ Platform Operator boundary
+ Security / MFA / crypto / external IdP seam
```

Close:

- Tenant authority after Organization split;
- tenant owner vs MetalDocs/platform operator;
- tenant creation/bootstrap/onboarding;
- suspension/deactivation and effects on sessions/users/jobs;
- tenant export;
- deletion request, grace period, cancellation, terminal erasure/crypto-shred/anonymization;
- which audit/business evidence survives vs becomes unreadable;
- credential/session/reset/security-signal ownership;
- whether V1 needs tenant-wide MFA policy vs only fresh reauth for sensitive actions;
- encryption/key ownership + rotation level actually required;
- exact trigger and identity-mapping seam for future Keycloak/external IdP.

After R8: R9 final Permission Catalog + Role Matrix + Domain Constraint/visibility Golden Matrix; then R10+ technical architecture/data/API/frontend/migration specifications and implementation plan.

## Documentation rule

The working tree contains active truth, not an archive. Historical staging remains in Git history. The active detailed ledger is the sole current WIP decision source.