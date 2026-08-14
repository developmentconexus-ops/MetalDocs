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

Locked so far:

- current AuthN retained for V1 behind a future external-IdP seam;
- Organization = Tenant / Area / User / flat Group / GroupMembership;
- scoped RoleAssignments for User or Group;
- built-in roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`;
- scoped RBAC + Groups is sufficient V1; no OpenFGA/SpiceDB now;
- Approval V1 = versioned sequential policy, ordered human Steps, NamedUser/Group/RoleInArea participants, ANY/ALL only, `accept` / `return_for_changes`, audited reassignment and optional reauth;
- no BPMN/CEL/M-of-N/generic delegation engine V1;
- `documents` + `controlleddocuments` + `templates` do not survive as three target contexts;
- Controlled Information core = stable `Document` + business `DocumentRevision`;
- `DocumentProfile` → `DocumentType`; `DocumentFamily` → classification-only `DocumentTypeCategory`; `GovernanceClass` deleted;
- Approval configuration explicitly distinguishes `NoHumanApproval` from `UsePolicy(...)`;
- template is a role of a governed Document; TemplateUse is M:N with at most one UX default/type;
- source template current EFFECTIVE Revision is resolved at creation only and exact origin is pinned permanently;
- official business revisions are `REV001`, `REV002`, ...; autosaves/checkpoints are technical history;
- one effective + at most one open Revision per Document V1;
- Revision states: DRAFT / SUBMITTED / EFFECTIVE / SUPERSEDED / OBSOLETE / CANCELLED;
- `RevisionSubmission` is immutable first-class evidence; resubmission creates a new Submission and, when required, new ApprovalInstance;
- return-for-changes/allowed withdrawal returns the same REV to DRAFT;
- Approval/Rendition/Release always bind exact Submission/digest;
- Document code is immutable tenant-wide identity; type/area/code immutable V1;
- numbering belongs to DocumentType and supports only literals + `{TYPE}` / `{AREA}` / `{SEQ}`, sequence scope TYPE or TYPE_AREA and padding width;
- normal Create has no manual-code override; legacy preservation belongs to import/migration;
- TemplateSpec owns authoring/fill contract only; no template MetadataSchema policy bundle and no CompositionJSON V1 without proven need;
- TemplateField separates value type from source; typed constraints + limited `visible_if` survive;
- DOCX/content anchors and TemplateSpec must pass parity before template submission;
- `RevisionContent` is the logical governed content identity; Submission digest covers all governed content/metadata needed to attest what was submitted;
- `title` is Revision metadata; responsible owner may be Document operational metadata and grants no authorization;
- no generic tenant custom-metadata engine V1;
- Periodic Review belongs to Controlled Information, configured as `Disabled | Every(n months)` on DocumentType;
- review cadence starts at Effectivity and restarts after completed review; overdue review does not automatically invalidate EFFECTIVE content;
- PeriodicReviewRecord is append-only against exact effective REV, with `confirmed_current | change_required`; change-required does not auto-create a REV;
- Documents subject to periodic review require a responsible owner relationship plus authorization to complete review;
- `Rendition` is immutable derived artifact of one RevisionSubmission with its own hash and generator/build provenance;
- only `OFFICIAL_PDF` is mandatory for Release V1; final DOCX may exist but does not block Effectivity;
- Approval approves Submission, not PDF bytes; official PDF attests derivation from the approved Submission and may manifest signature data;
- stored official bytes + hash + source digest + generator identity are durable evidence; no promise of future bit-identical rerender;
- `RevisionSubmission` replaces `ReleaseGeneration` as the required domain identity for release;
- Release is automatic/system-owned; no human publish button;
- release uses optional `ReleasePlan.not_before`, not a SCHEDULED Revision state;
- actual `effective_at = released_at` in the winning transaction; no silent retroactive effectivity;
- winning Release atomically makes candidate EFFECTIVE, prior REV SUPERSEDED, swaps Document pointers, records ReleaseRecord and emits lifecycle events;
- a candidate can be CANCELLED after Approval but before Release; evidence remains historical;
- historical `effective_date` token semantics are reopened because mandatory pre-release PDF cannot depend on a fact created only at Release;
- current outbox/jobs infrastructure class remains conceptually sufficient; no Temporal/Camunda requirement proven.

## Critical invariant

The target must make the historical content mismatch impossible by construction:

```text
RevisionSubmission
    ├── Approval
    ├── Rendition(s)
    └── Release
```

All three refer to the same immutable submission identity/digest.

## Exact next step — R7

Design only:

```text
Distribution + Read/Acknowledgement
+ Tokens / Computed / Dictionary Values
+ Audit / Evidence
+ Notifications
+ Search
```

Close:

### Distribution
- who defines distribution obligations;
- User/Group/Area targeting;
- denominator snapshot at Release vs live membership;
- read vs explicit acknowledgement;
- deadlines/reminders/export;
- reauth/signature requirement, if any;
- historical behavior when Group membership changes.

### Tokens
- final system-token catalogue;
- when each value resolves and freezes;
- dictionary pinning;
- resolver identity/version/provenance;
- collision rules;
- explicit fate of legacy `effective_date`, `approval_date`, `approvers`, `revision_number`, `doc_title`, `controlled_by_area` and related names.

### Audit / Notifications / Search
- domain evidence vs global audit trail;
- same-transaction evidence/outbox requirements;
- hash-chain/export/erasure semantics;
- domain event catalogue;
- notifications as consumer/projection only;
- search as rebuildable projection with correct working/effective visibility and eventual consistency.

After R7: R8 Tenant lifecycle/Security; R9 final Authorization matrix; then R10+ technical architecture/data/API/frontend/migration specifications and implementation plan.

## Documentation rule

The working tree contains active truth, not an archive. Historical staging remains in Git history. The active detailed ledger is the sole current WIP decision source. Do not restore deleted historical plans into the live tree.