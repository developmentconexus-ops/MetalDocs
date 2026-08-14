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

Do not restore/resume historical roadmaps, milestones, specs, deleted `docs/superpowers` material or stale implementation PRs by inertia.

## Current checkpoint

The product/domain redesign is closed enough to descend into technical architecture. Locked business model:

- local AuthN V1 behind a future external-IdP/assurance seam;
- Organization = Tenant / Area / User / flat Group / GroupMembership;
- five tenant roles only: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`;
- RoleAssignment subject User|Group, typed TenantScope|AreaScope; additive/default-deny; no tenant-owner bypass;
- final tenant Permission Catalog = 29 semantic permissions; old 8-role/38-capability registry is not target;
- authorization = Permission + required resource/case relationship + Domain Constraints;
- PlatformOperator/SystemPrincipal are outside tenant RBAC and gain no implicit tenant-content access;
- Approval V1 = versioned sequential policies, ANY/ALL, NamedUser/Group/RoleInArea, accept/return-for-changes, audited reassignment, optional fresh-auth, strict SoD;
- same user cannot accept two different Steps of one ApprovalInstance; author/submitter cannot accept their own Submission;
- stable Controlled Information core = `Document` + business `DocumentRevision`; official labels `REV001`, `REV002`, ...;
- Revision states: DRAFT / SUBMITTED / EFFECTIVE / SUPERSEDED / OBSOLETE / CANCELLED;
- `RevisionSubmission` is immutable first-class attempt identity; Approval/Rendition/Release always bind exact Submission/digest;
- `DocumentProfile`→`DocumentType`; Family→classification-only category; GovernanceClass deleted;
- Template = role of governed Document; no parallel TemplateVersion lifecycle; TemplateUse M:N; source effective REV resolved once at creation and pinned forever;
- Document code tenant-wide unique/immutable; numbering belongs to DocumentType and uses only literals + `{TYPE}/{AREA}/{SEQ}` with TYPE or TYPE_AREA sequence scope;
- TemplateSpec owns authoring/fill contract only; no generic metadata/policy bundle; mutable dictionary values snapshot at new REV creation;
- Periodic Review belongs to Controlled Information; overdue does not invalidate EFFECTIVE; append-only review evidence against exact REV;
- Rendition is immutable derived artifact of one Submission; only OFFICIAL_PDF mandatory for Release V1;
- Release is automatic/system-owned and idempotent; no publish button; actual `effective_at = released_at`; atomic candidate EFFECTIVE/prior SUPERSEDED/pointer swap/ReleaseRecord;
- Distribution = concrete per-user obligation/acknowledgement over released REV; group membership snapshots at release; Notification READ/view/download never equals acknowledgement;
- Audit Trail is append-only/tamper-evident transversal evidence, never source of business state; critical governed mutations require durable audit intent in commit boundary;
- Notifications = projection/delivery; Search = rebuildable/evenually-consistent projection with canonical AuthZ recheck;
- Tenant lifecycle = ACTIVE / SUSPENDED / ERASED; deletion request separate with grace/cancel; onboarding uses activation credential, not operator-chosen password;
- tenant erasure removes live tenant data/blobs, destroys Tenant DEK, preserves allowed non-PII audit/platform skeleton and terminal erasure tombstone;
- current MFA coverage is fake/stub and not target; real MFA/passkeys/SSO/SAML triggers re-evaluation of Keycloak/external IdP.

## Final R9 authorization anchors

### Tenant permissions (29)

```text
tenant.settings.manage
organization.manage
access.manage
document_type.manage
approval_policy.manage
template_use.manage
dictionary.manage

document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.comment
document.submit
document.cancel_revision
document.obsolete
document.review_periodic
document.owner.manage

approval.act
approval.oversee
approval.reassign
approval.cancel

distribution.manage
distribution.oversee

audit.read
audit.export
session.manage
tenant.export
tenant.deletion.request
```

### Role bundles

- `viewer`: `document.read_effective`.
- `author`: viewer + history/working/create/edit/comment/submit/periodic-review qualification.
- `approver`: viewer + `approval.act`; no blanket draft access.
- `area_manager`: author + revision cancel/obsolete/owner management + approval act/oversee/reassign/cancel + distribution manage/oversee in Area.
- `tenant_owner`: all 29 Permissions via normal Authorizer; still obeys relationships, SoD, lifecycle, fresh-auth and tenant-operability constraints.

Narrow case/self access is relation-driven rather than fake permissions: Approval participant exact Submission, Distribution assignee exact effective REV/acknowledgement, submitter withdrawal, own notifications/sessions/password, system release/rendition/erasure.

RLS is tenant-isolation defense-in-depth only. DB constraints enforce structural invariants. Current `system_admin` bypass, magic `"tenant"` area sentinel and asserted-capability GUC authorization model have no target entitlement.

## Exact next step — R10

**Design only — no product implementation.**

Create the integrated technical architecture from approved semantics, not current packages:

1. bounded contexts/modules and final names;
2. dependency DAG / legal imports / published ports;
3. aggregate ownership + application coordinators;
4. target table ownership and DB constraints;
5. transaction boundaries for governed mutations;
6. durable audit-intent + outbox/domain-event catalogue;
7. jobs/timers/reconciliation ownership;
8. object-storage artifact ownership/key strategy;
9. final build-vs-buy choices;
10. current module `KEEP / MOVE / REWRITE / DELETE` map;
11. current table `KEEP / TRANSFORM / DROP` map;
12. migration ordering/expand-contract/compatibility policy.

After R10: R11 API/frontend journeys; R12 proof matrix + final ADR/spec promotion/adversarial review; R13 implementation specification/plan. Implementation stays blocked until explicit integrated-design approval.

## Documentation rule

The working tree contains active truth, not an archive. Git history is the archive. The active ledger is the single detailed WIP authority.