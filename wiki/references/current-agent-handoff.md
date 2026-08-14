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

### Locked core

- local AuthN stays for V1 behind a future external-IdP seam;
- Organization = Tenant / Area / User / flat Group / GroupMembership;
- tenant roles = `tenant_owner`, `area_manager`, `author`, `approver`, `viewer` through scoped User/Group RoleAssignments;
- Approval V1 is sequential/versioned, ANY/ALL only, exact immutable Submission attempts, `accept | return_for_changes`, audited reassignment and optional reauth; no BPM engine;
- Controlled Information core = stable Document + business DocumentRevision;
- `DocumentProfile` → DocumentType; Family → classification-only category; GovernanceClass deleted;
- official Revisions = `REV001`, `REV002`, ...; autosaves are technical history;
- Revision states = DRAFT / SUBMITTED / EFFECTIVE / SUPERSEDED / OBSOLETE / CANCELLED;
- RevisionSubmission = immutable attempt identity; Approval/Rendition/Release bind exact Submission/digest;
- code/type/area identity is stable; numbering is small DocumentType config using `{TYPE}/{AREA}/{SEQ}` only;
- template is role of governed Document; TemplateUse M:N; exact effective source REV is pinned at creation; TemplateSpec only owns authoring contract;
- RevisionContent is composite governed truth; title is Revision metadata; no generic tenant custom-metadata engine V1;
- Periodic Review belongs to Controlled Information and records exact effective-REV review evidence;
- Rendition is immutable derived artifact; only OFFICIAL_PDF mandatory for Release V1;
- Release is automatic/system-owned/idempotent over `submission_id`; no publish button, no ReleaseGeneration domain noun, no SCHEDULED Revision state;
- release atomically flips candidate EFFECTIVE, prior REV SUPERSEDED, Document pointers, ReleaseRecord and lifecycle events.

### Locked R7 supporting services

- Distribution = explicit obligation/acknowledgement, not RBAC or Training;
- release expands User/Group targets into concrete per-user assignments; membership changes never rewrite history;
- acknowledgement is explicit immutable evidence; notification/view/download is not acknowledgement;
- System Value Catalog is product-owned; tenant Dictionary is mutable source whose values snapshot when a new REV is created;
- domain evidence stays domain authority; Audit Trail is transversal append-only/tamper-evident/exportable timeline only;
- critical governed mutations require durable audit intent in same commit boundary; usage telemetry may be async;
- Notifications are projection/delivery only; notification READ means notification read;
- Search is rebuildable/eventually-consistent projection; Official Library = effective REV, optional Working Search = authorized open REV; stale result never grants access.

### Locked R8 tenant/security boundary

- `PlatformOperator` and `SystemPrincipal` are platform identities outside tenant RoleAssignment; they are not extra tenant roles;
- PlatformOperator may create/suspend/resume tenants but has no implicit access to tenant business content;
- Tenant states V1: ACTIVE / SUSPENDED / ERASED;
- `TenantDeletionRequest` is separate PENDING/CANCELLED/EXECUTED process with grace period; tenant may remain ACTIVE until execution;
- onboarding creates Tenant + initial User + `tenant_owner @ Tenant` + single-use/time-limited activation credential; platform operator does not choose the owner's password and no `system_admin` is created;
- suspension revokes tenant sessions, blocks login/business mutations, preserves data; business jobs respect suspension; lifecycle/security jobs may continue;
- user deactivation revokes sessions but preserves identity/history; pending responsibilities require explicit reassignment/attention;
- tenant owner can export own tenant and request/cancel own deletion; export/deletion request require fresh authentication;
- terminal erasure: suspend/revoke sessions → delete live tenant rows → delete tenant blobs → destroy tenant DEK → preserve allowed non-PII audit/platform skeleton → ERASED + platform TenantErasureRecord;
- target Audit Trail is not deleted during tenant erasure; retained skeleton uses opaque IDs/non-PII, sensitive payload is encrypted/erasable and may become unreadable through tenant-key destruction;
- backup/restore must reapply erasure tombstones before restored service is available;
- V1 crypto remains small: platform KEK wraps per-tenant DEK; no per-document key hierarchy/rotation product;
- current local AuthN exposes/derives auth time/method/assurance/fresh-auth evidence;
- stub MFA coverage is not a real V1 control and is targeted for deletion;
- real MFA/passkeys/SSO/SAML/per-tenant federation are formal triggers to re-evaluate Keycloak/external IdP before rebuilding IdP features internally;
- future `(issuer, subject) -> internal User` mapping changes authentication only; MetalDocs remains authority for Organization/AuthZ/workflow;
- current `security` catch-all is conceptually split: sessions/lockouts → AuthN; tenant crypto → Platform Security; heuristic Security Signals are optional/deferred projection, not V1 foundation.

## Critical content invariant

```text
RevisionSubmission
    ├── Approval
    ├── Rendition(s)
    └── Release
```

All three refer to the same immutable submission identity/digest. Audit/Notifications/Search never compete with those facts.

## Exact next step — R9

Design only: **Final Authorization Matrix**.

Now that product operations are known, close:

- final semantic Permission Catalog;
- exact legal scopes for every permission;
- five built-in tenant-role bundles;
- Organization admin: Users/Groups/Memberships/Areas/RoleAssignments;
- DocumentType/category/ApprovalPolicy/Dictionary/template designation configuration permissions;
- Document/Revision authoring, submit, withdraw, cancel, obsolete and periodic-review operations;
- Approval act/oversee/cancel/reassign relationships;
- Distribution configure/assign/acknowledge/oversee and task-specific read access;
- Audit/export/session/security-admin reads/actions;
- tenant export/deletion-request/cancel operations;
- effective vs working vs case-specific visibility/filter semantics;
- relationships: Approval participant, Distribution assignee, responsible owner, submitter;
- Domain Constraints/SoD/fresh-auth/tenant-operability checks;
- RLS/DB constraints/tripwire backstops;
- positive/negative Golden Matrix for every sensitive operation.

PlatformOperator/System authority remains separate from the tenant Permission Catalog.

After R9: whole-product domain map can be considered closed subject to adversarial review, then R10+ technical architecture/data/event/API/frontend/migration specifications and implementation plan.

## Documentation rule

The working tree contains active truth, not an archive. Historical staging remains in Git history. The active detailed ledger is the sole current WIP decision source. Do not restore deleted historical plans into the live tree.