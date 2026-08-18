# MetalDocs Rebaseline Decision Reconciliation — Candidate Baseline

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — OPERATOR REVIEW PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED  
> **Current technical stage:** T3 PAUSED ON DECISION RECONCILIATION

## 0. Purpose

The Whole-Product rebaseline changed Launch scope and semantic ownership after a large body of R3–R9.5 and R10-A→C decisions had already been reviewed. The correct response is neither:

```text
copy every old decision forward
```

nor:

```text
forget every old decision and redesign from zero
```

The operator-approved working rule is:

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

This candidate reconciles the prior decision corpus into a current baseline before T3 continues.

It is intentionally a **disposition registry**, not a new competing source of detailed product truth. After operator approval it should be promoted to a durable `wiki/architecture/` registry that points to the canonical authority for each active decision. Product Contract, GCR, ownership topology, ratified T1/T2 and later ratified T-stages remain the detailed authorities.

---

# 1. Authority and evidence order

Current authority:

1. `AGENTS.md`;
2. DevelopmentConexus Engineering Method v1.0.0;
3. `wiki/references/current-agent-handoff.md`;
4. `wiki/architecture/launch-v1-product-contract.md`;
5. `wiki/architecture/whole-product-alignment-review.md`;
6. `wiki/architecture/launch-v1-ownership-topology.md`;
7. `wiki/architecture/r10-technical-architecture.md` — ratified T1;
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md` — ratified T2.

Reconciliation evidence:

- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — consolidated R3–R9.5 prior decision ledger;
- R10-B1 corrected relational-substrate reviews;
- R10-B2 AuthN/Organization/AuthZ corrected target;
- R10-B3 Controlled Information candidate;
- R10-B4 Approval/Rendition/Release/Distribution candidate;
- R10-B5 Documentary Context/Records candidate;
- R10-B6 Audit/Interchange candidate;
- paused R10-C physical-integrity candidate only as safety/mechanism evidence.

Old files never override current authority.

---

# 2. Disposition vocabulary

```text
CURRENT
  Already re-proven and ratified by current Product/GCR/Topology/T1/T2 authority.
  Remaining T-stages consume it; they do not casually re-open it.

PRESERVE
  Earlier decision is still coherent with current authority and has no current counterexample.
  Treat as revalidated baseline input when this registry is ratified, while its owning later stage still chooses realization detail.

REFINED
  The essential decision survives, but its old wording/shape depended on superseded scope/topology.
  Only the corrected statement in this registry/current authority survives.

REOPEN
  Prior decision is useful evidence but current authority has not re-proven it or changed conditions materially.
  The named T-stage must decide it deliberately.

DEFERRED
  Capability/decision is expected later but is not Launch authority/state now.
  Preserve attachment seam and old design as evidence only; no dormant Launch implementation.

SUPERSEDED
  Prior decision conflicts with current authority or encoded a removed owner/capability.
  Do not carry it into a T-stage except as a counterexample/provenance reference.
```

Rule for future T-stages:

```text
CURRENT/PRESERVE/REFINED → baseline; reopen only with explicit counterexample
REOPEN                   → deliberate T-stage decision required
DEFERRED                 → future seam only
SUPERSEDED               → forbidden as target inheritance
```

---

# 3. Cross-cutting current laws

These are already CURRENT and control every later classification:

```text
one semantic authority per meaning
mechanism != authority
single company per Launch deployment
4 business owners + Audit supporting owner
Document != Revision != WorkingContent != Submission
REV000 = initial issuance; REV001 = first revision
Submission immutable exact governed attempt
one sequential governance Step semantic
NoHumanApproval creates no fake approver
Release is system-owned effectivity authority
one EFFECTIVE Revision at most per Document
replacement Release is atomic SUPERSEDED + EFFECTIVE
explicit governed obsolescence without replacement
Search never grants access or establishes effectivity
Audit proves actions and never reconstructs current state
storage/provider identity never becomes semantic identity
imported history never becomes fake native history
future capability attaches by reference; no dormant platform machinery
one local ACID transaction per native business transition
external/provider calls never join local semantic atomicity
Document is lifecycle serialization root
WorkingContent uses OCC/CAS for DRAFT races
READ COMMITTED + narrow explicit serialization/CAS is the accepted default posture
```

---

# 4. Authentication reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| AUTH-01 | Authentication provider owns credentials/password/MFA/federation | **CURRENT** | Provider credential truth remains outside MetalDocs semantic identity. | T6 may expose journeys only. |
| AUTH-02 | Keycloak chosen as V1 provider | **PRESERVE — mechanism selection** | Keep Keycloak as the existing selected AuthN provider unless concrete deployment evidence reopens it. It remains replaceable behind the provider anti-corruption seam and owns no MetalDocs roles/groups. | Implementation/ops realization, not T3 semantic authority. |
| AUTH-03 | Stable provider identity = issuer + subject | **CURRENT** | `ProviderSubjectBinding` binds provider subject to stable MetalDocs User. Email/username/display name are attributes. | Closed by T1. |
| AUTH-04 | Provider roles/groups/org/claims do not feed canonical AuthZ | **CURRENT** | No provider-role mapping/claim bridge. | T3 consumes. |
| AUTH-05 | Application Session is MetalDocs-owned and independently revocable | **CURRENT** | Session lifecycle can terminate product access independently of provider session. Session is not Role/Permission authority. | T3 check/offboarding; T6 journey. |
| AUTH-06 | No atomic MetalDocs↔Keycloak transaction | **PRESERVE / T2-compatible** | Provider changes are post-commit/reconciled effects; semantic local transaction remains authoritative. | T5 execution mechanism. |
| AUTH-07 | Fresh-auth/reauth evidence available for approval | **DEFERRED** | No named Launch consumer currently requires fresh-auth/eSignature. Authentication remains the owner if later promoted. | Reopen T3/T6 only on named requirement. |

---

# 5. Organization reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| ORG-01 | `Tenant` singleton semantic company root | **REFINED / CURRENT** | Current product term is stable single `Company` root per deployment. It is not a universal partition key. | T1 closed. |
| ORG-02 | User is stable organizational participant identity | **CURRENT** | Offboarding never rewrites historical actor identity. | T1/T3. |
| ORG-03 | UserProfile separated/erasable | **CURRENT** | Human-readable/contact enrichment may be erased without rewriting immutable governance/Audit references. | T1/T4/T7 privacy proof. |
| ORG-04 | Area is stable flat organizational scope | **CURRENT** | Stable Area identity/code/name/eligibility; no hierarchy/default approver platform without consumer. | T1. |
| ORG-05 | Group is flat company grouping | **CURRENT** | Group + current GroupMembership remain first-class Organization facts. No nested/dynamic/provider-mirrored groups. | T1/T3. |
| ORG-06 | GroupMembership is current truth only | **CURRENT** | Row/relation exists = current membership; historical transition evidence belongs in Audit when required. | T1/T3. |
| ORG-07 | User↔Area membership/home-area relation | **SUPERSEDED AS BASELINE** | No independent `UserAreaMembership` unless a current journey proves it. Area-scoped grants do not require such a relation. | Reopen only on named consumer. |
| ORG-08 | Area retirement blocks new references while preserving old history | **PRESERVE** | Retirement/inactivity is future-facing; existing Document/history references remain truthful. | T3/T6 admin details. |
| ORG-09 | Group hard delete only when no current/live references remain | **PRESERVE** | A Group must not disappear while current memberships, RoleAssignments or live governance configuration needs it. Historical activated governance uses concrete User snapshots, so history need not keep Group alive solely for attribution. | T3/T6 exact administration law. |
| ORG-10 | User offboarding destructively removes future access; re-enable does not restore old authority | **PRESERVE** | Disable User; revoke application Sessions; remove direct grants and current memberships as required; history survives. Re-enable restores identity eligibility only, never deleted grants/memberships/sessions. | T3 must integrate exact same-commit/AuthZ/Audit law. |

---

# 6. Authorization reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| AZ-01 | Product-owned static Role/Permission catalogs | **CURRENT** | No customer-editable custom-role/custom-permission platform in Launch. | T1/T3. |
| AZ-02 | RoleAssignment is sole persisted current grant family | **CURRENT** | Current grant subject is `User | Group`. | T1/T3. |
| AZ-03 | Scope = whole company or Area | **REFINED / CURRENT** | Old `TenantScope` becomes `CompanyScope`; AreaScope remains. No magic scope sentinel. | T1/T3. |
| AZ-04 | Additive grants + default deny | **CURRENT** | No deny engine/ACL graph required. | T1/T3. |
| AZ-05 | Direct User grants + Group-mediated grants evaluated live | **PRESERVE** | Canonical evaluation uses current RoleAssignments + current GroupMembership, not Session/JWT durable permission snapshots. | T3. |
| AZ-06 | Domain relationship/lifecycle constraints are not RBAC permissions | **CURRENT** | Grant+scope is necessary but Controlled Documents predicates still decide eligibility/state. | Ownership/T1/T2/T3. |
| AZ-07 | No role, including admin, bypasses domain governance | **CURRENT** | Admin is a grant bundle, never `allow everything`. | GCR/T1/T3. |
| AZ-08 | Exact five-role set (`tenant_owner`, `area_manager`, `author`, `approver`, `viewer`) | **REOPEN / PARTIAL PRESERVE** | Role concepts remain useful evidence, especially `area_manager`, author, approver, viewer and whole-company admin. Names/bundles must be rederived for current Launch and GCR requires least-privilege Governance Viewer/Auditor. | T3. |
| AZ-09 | Exact 5×43 permission matrix | **SUPERSEDED** | It encoded Distribution, Periodic Review, Dossier, Evidence, Records, Interchange and old Approval richness. Do not preserve by subtraction. | T3 regenerates Launch catalog. |
| AZ-10 | Area Manager is operational, not RBAC administrator | **PRESERVE CANDIDATE** | Area Manager concept remains valuable; it should not silently gain whole-company access administration. Exact Launch bundle still needs T3 derivation. | T3. |
| AZ-11 | `access.manage` protects both RoleAssignment and GroupMembership mutation | **PRESERVE CANDIDATE** | Membership can change inherited authority, so access administration must guard it even though Organization owns membership semantics. | T3. |
| AZ-12 | RLS as Role/Area/permission policy engine | **SUPERSEDED / REJECTED** | Canonical Authorization is application/domain authority; single-company Launch does not require Tenant RLS substrate. DB remains structural backstop only. | Implementation security proof later. |
| AZ-13 | Universal `tenant_id` composite PK/FK and tenant RLS | **SUPERSEDED** | Single-company rebaseline + Company root replaced pooled-tenant substrate. | Do not revive before pooled-tenancy trigger. |

---

# 7. Governance reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| GOV-01 | Approval specialized sequential human workflow, not BPM | **REFINED / CURRENT** | Governance is internal Controlled Documents semantics, not separate Approval owner; still one sequential Step model. | GCR/T1/T2. |
| GOV-02 | Separate Approval semantic owner/context | **SUPERSEDED** | Governance belongs inside Controlled Documents for the two Launch consumers. | GCR A3/A4. |
| GOV-03 | Versioned ApprovalPolicy aggregate mandatory | **SUPERSEDED AS REQUIREMENT** | Current route configuration + immutable per-attempt snapshot is sufficient. A first-class browsable versioned policy is added only if a real consumer proves it. | T1. |
| GOV-04 | Structural Step purpose `REVIEW | APPROVAL` | **SUPERSEDED** | One Step semantic; label is business language only. | GCR/T1. |
| GOV-05 | Actor selector `NamedUser | Group | RoleInArea` | **REFINED / CURRENT** | Launch T2 accepts `NAMED_USER | GROUP` only. `ROLE_IN_AREA` is not baseline. | T2. |
| GOV-06 | Configurable `ANY | ALL`/quorum | **REFINED / CURRENT** | Sequential route; Group Step = ANY-one from concrete enabled membership snapshot at activation. ALL/N-of-M is deferred. | T2. |
| GOV-07 | Group/actor pool remains live throughout attempt | **REFINED / CURRENT** | Membership resolves to concrete candidate snapshot at Step activation; later drift does not rewrite active Step. Current AuthZ is still rechecked when acting. | T2. |
| GOV-08 | Strict creator/submitter SoD + cross-Step same-user SoD | **REFINED / CURRENT** | Only bounded independence is Launch baseline: Submission submitter / obsolescence initiator cannot satisfy a human Step on that same attempt. Cross-Step prohibition is deferred. | T2. |
| GOV-09 | Fresh-auth per configured Step | **DEFERRED** | No current Launch consumer proves it. | Reopen on regulation/customer contract. |
| GOV-10 | Due dates/SLA/escalation | **DEFERRED** | No Launch lifecycle effect. | Future governance enhancement. |
| GOV-11 | Reassign/overseer/delegation machinery | **DEFERRED** | Launch escape: withdraw → fix current route → resubmit. Reassignment reopens only if operationally insufficient. | T2. |
| GOV-12 | Approval cancel as separate workflow operation | **REFINED** | Withdraw Submission attempt and cancel Revision have explicit separate meanings; no generic approval-cancel platform. | T2. |
| GOV-13 | NoHumanApproval creates no fake approver | **CURRENT** | Human gate is absent; same applies to accepted no-human obsolescence. | Product/T1/T2. |
| GOV-14 | GovernanceAttempt arbitrary generic subject | **SUPERSEDED / REJECTED** | Closed subject universe `SUBMISSION | OBSOLESCENCE`; adding a third subject is a material reopen. | T1. |

---

# 8. Controlled Document / revision reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| DOC-01 | Stable Document identity across revisions | **CURRENT** | Document is not file/content/version. | Product/T1. |
| DOC-02 | Revision is business change cycle, autosave is not Revision | **CURRENT** | `REV000` initial issuance; later ordinals monotonic/non-reused. | Product/T1/T2. |
| DOC-03 | Initial numbering started at REV001 | **SUPERSEDED** | Binding correction: `REV000` initial; `REV001` first revision. | Product/T1/T2. |
| DOC-04 | Revision lifecycle states DRAFT/SUBMITTED/EFFECTIVE/SUPERSEDED/OBSOLETE/CANCELLED | **CURRENT** | Same vocabulary. | Product/T1. |
| DOC-05 | At most one open and one EFFECTIVE Revision | **CURRENT** | Prior EFFECTIVE may coexist with newer DRAFT/SUBMITTED. | T1/T2. |
| DOC-06 | `REV002+` mandatory reason-for-change | **SUPERSEDED AS LAUNCH REQUIREMENT** | No accepted Product Contract consumer currently requires this rule. May reopen deliberately later. | Not baseline. |
| DOC-07 | Document code/type/Area are stable | **CURRENT** | Stable official code, DocumentType and Area context; transfer is not silently invented. | T1. |
| DOC-08 | Responsible owner is current mutable Document relationship | **CURRENT** | Historical actions remain bound to actual actors; responsibility changes do not rewrite history. | T1/T3/T6. |
| DOC-09 | Document duplicates `currentRevision/currentStatus` pointer authority | **SUPERSEDED / REJECTED** | Current-effective truth derives from unique Revision lifecycle established by Release/Obsolescence. Cache/pointer may only be structurally synchronized projection. | T1/T2. |
| DOC-10 | DocumentType category/taxonomy | **DEFERRED** | No classification platform in Launch without named consumer. | Future/reopen. |
| DOC-11 | Tenant/company editable dictionary / System Value platform | **DEFERRED** | No Launch dictionary machinery absent consumer. | Future/reopen. |
| DOC-12 | Exact numbering grammar `{TYPE}/{AREA}/{SEQ}`, TYPE/TYPE_AREA and padding | **REOPEN / PRIOR EVIDENCE** | Numbering capability and stable non-reused allocation are CURRENT; exact configurable grammar remains to be validated against Launch admin/API journeys before implementation. | T6/implementation spec. |
| DOC-13 | Number allocation is monotonic; committed codes never reuse; preview reserves nothing | **CURRENT/PRESERVE** | T2 accepted code allocation as atomic create and no cosmetic gap-free guarantee. | T2/T6. |

---

# 9. Working Content / Submission / Template reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| CNT-01 | WorkingContent = sole mutable DRAFT authority | **CURRENT** | Browser/editor/provider state is never authority. | T1/T2. |
| CNT-02 | One monotonic OCC generation; no silent last-write-wins | **CURRENT** | Caller-observed generation/CAS guards DRAFT mutation and SUBMIT. | T2. |
| CNT-03 | EditorSession/one-active-writer lease required | **REOPEN / MECHANISM** | OCC is correctness authority; editor lease is optional UX/recovery mechanism only if T4/T6 proves need. | T4/T6. |
| CNT-04 | WorkingSnapshot as business history | **SUPERSEDED / REJECTED** | Recovery checkpoint may exist only as technical mechanism; never official history. | T4/T5 if needed. |
| CNT-05 | Submission immutable and exact | **CURRENT** | Same Revision may have multiple immutable attempts; return/withdraw never mutate old Submission. | T1/T2. |
| CNT-06 | Submission must freeze coherent content + governed state/config | **CURRENT** | Exact WorkingContent generation + decision-relevant governance/representation snapshot. | T1/T2. |
| CNT-07 | Mandatory RFC8785/JCS manifest + SHA-256 Submission digest shape | **REOPEN** | Exact-content proof is required, but canonical encoding/hash realization belongs T4. | T4. |
| CNT-08 | Standalone `Artifact` owns exact bytes | **SUPERSEDED** | No Artifact semantic owner. Exact-content identity facts live on WorkingContent/Submission/Rendition/imported target facts; storage is mechanism. | GCR/T1/T4. |
| CNT-09 | Template is ordinary governed Document role | **CURRENT** | No parallel TemplateVersion lifecycle. | Product/T1. |
| CNT-10 | Template eligibility M:N to target DocumentTypes | **CURRENT** | Current eligibility relation remains small current config. | T1/T6. |
| CNT-11 | Derived Document origin strongly references native source Submission | **REFINED / CURRENT** | Origin pins source Template Document + exact EFFECTIVE Revision/content provenance, without requiring native Submission because imported truth may lack one. | T1/T2/T7. |
| CNT-12 | Structured TemplateSpec platform | **DEFERRED** | Not Launch baseline. | Future named structured-authoring consumer. |
| CNT-13 | DRAFT EditorialComment platform / unresolved comments block submit | **DEFERRED** | Submission feedback remains product-required; DRAFT collaboration/comment platform is not Launch baseline. | Future/reopen. |
| CNT-14 | EigenPal as DOCX authoring provider | **PRESERVE — mechanism selection** | Keep as selected adapter/provider evidence if still used; it never owns WorkingContent/Revision/Submission meaning. | T6/provider conformance. |
| CNT-15 | Realtime Yjs/CRDT coauthoring | **DEFERRED** | Known future seam attaches at WorkingContent concurrency boundary. | Future. |

---

# 10. Rendition / Release / effectivity reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| REL-01 | SourceOnly vs one required official representation | **CURRENT** | DocumentType/Submission freezes representation requirement. Preview is not semantic Rendition. | Product/T1/T2. |
| REL-02 | Rendition binds exact Submission, not mutable latest content | **CURRENT** | Required official rendition is immutable exact derived representation/provenance. | T1/T2/T4/T5. |
| REL-03 | Universal mandatory PDF | **SUPERSEDED** | SourceOnly remains valid. | Product/GCR. |
| REL-04 | Scheduled/future-dated `ReleasePlan.not_before` | **DEFERRED** | No Launch consumer. | Reopen on real scheduled-effectivity requirement. |
| REL-05 | Human publish button | **SUPERSEDED / REJECTED** | Release is system-owned when all gates are satisfied. | Product/T1/T2. |
| REL-06 | First Release and replacement Release atomicity | **CURRENT** | First establishes EFFECTIVE; replacement atomically SUPERSEDED predecessor + EFFECTIVE successor. | T2. |
| REL-07 | Release waits on human gate + required rendition gate only | **CURRENT** | Gates are orthogonal; truthful SUBMITTED may persist while rendition remains missing. | T2. |
| REL-08 | Distribution obligations are created inside Release transaction | **SUPERSEDED FOR LAUNCH** | Distribution moved to Launch+ and is explicitly outside Launch-Core Release atomicity. | Future Launch+ design. |

---

# 11. Obsolescence reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| OBS-01 | Obsolescence as explicit governed withdrawal without successor | **CURRENT** | Raw status toggle is invalid. | Product/T1/T2. |
| OBS-02 | Target remains EFFECTIVE while human governance is active | **CURRENT** | Only successful completion transitions it to OBSOLETE. | T2. |
| OBS-03 | NoHumanApproval obsolescence still needs one human Step | **SUPERSEDED** | Accepted T1-J: zero human Step after authorization, reason and invariant checks; no fake approver. | T1/T2. |
| OBS-04 | Open replacement Revision may coexist with active obsolescence intent | **REFINED / CURRENT** | T2 chose stronger Launch mutual exclusion: obsolescence starts only with no open replacement, and active obsolescence blocks new Revision. | T2. |
| OBS-05 | Separate obsolescence governance route | **DEFERRED** | Reuse same DocumentType governance route in Launch; separate route is reopen trigger. | T2. |
| OBS-06 | Reactivation of OBSOLETE | **DEFERRED / OUT OF LAUNCH** | No reactivation journey. | Future explicit requirement. |

---

# 12. Search / notifications / async reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| ASY-01 | Search is rebuildable/eventually-consistent discovery projection | **CURRENT** | Search never grants access or establishes current effectivity. Canonical state/AuthZ is re-resolved before serve. | GCR/T1; T5/T6 realize. |
| ASY-02 | Notifications are delivery projection, not domain authority | **PRESERVE** | Do not make notification delivery part of lifecycle truth. Launch notification requirement still must be named by T5/T6. | T5/T6. |
| ASY-03 | Required future external/async work gets durable intent in same local transaction | **PRESERVE / T2-compatible** | Only when a real async/external effect is required; outbox/intent is mechanism, never state authority. | T5. |
| ASY-04 | River/job framework from current implementation is target authority | **SUPERSEDED AS AUTHORITY** | Existing worker tech is evidence only. T5 selects the smallest mechanism from current requirements. | T5. |
| ASY-05 | Global SERIALIZABLE or global worker lock framework | **SUPERSEDED/REJECTED** | Use narrow serialization/CAS unless later proof disproves it. | T2/T5. |

---

# 13. Audit reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| AUD-01 | AuditEvent is transversal evidence, not domain state/event sourcing | **CURRENT** | Current state never derives by latest Audit event. | GCR/T1. |
| AUD-02 | Critical governed/security mutations require same-local-commit Audit | **CURRENT AT PRINCIPLE / REOPEN CENSUS** | Principle is ratified; exact operation census and bounded facts belong T3. | T3. |
| AUD-03 | Audit append-only and PII-minimized | **CURRENT** | Profile/credentials/tokens/request bodies/free-form domain reasons are not copied by convenience. | GCR/T1/T3. |
| AUD-04 | Deployment-wide AuditChainHead/hash-chain/global terminal lock | **SUPERSEDED / DEFERRED** | No Launch assurance requirement justifies global chain/serialization. | Reopen only on concrete tamper-evidence/non-repudiation requirement. |
| AUD-05 | Audit export permission/capability | **DEFERRED** | Launch requires trustworthy history/read path, not generic export. | Future/named auditor requirement. |
| AUD-06 | Audit retained indefinitely | **REOPEN / PRIOR EVIDENCE** | No current accepted retention period. Avoid inventing a statutory claim. | Future compliance/records decision. |
| AUD-07 | DB CRUD triggers infer semantic Audit | **SUPERSEDED / REJECTED** | Audit must record semantic operation meaning through explicit use-case composition. | T3. |

---

# 14. Storage / content integrity reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| STO-01 | Storage/provider identity is not semantic identity | **CURRENT** | Provider key/bucket/version/location never becomes Document/Revision/Submission/Rendition identity. | GCR/T1. |
| STO-02 | Provider-neutral shared storage/integrity seam | **PRESERVE / REQUIRED SEAM** | Shared mechanism may serve current/future semantic owners without becoming an owner itself. | T4. |
| STO-03 | `ManagedArtifactStore` as Artifact-owned semantic surface | **SUPERSEDED IN NAME/OWNERSHIP** | T4 may preserve a provider-neutral managed-content mechanism, but no Artifact semantic owner/API entitlement follows. | T4. |
| STO-04 | SHA-256 exact-byte identity | **PRESERVE CANDIDATE** | Cryptographic exact-content digest is required; SHA-256 remains strong prior choice absent counterexample. Exact descriptor/canonicalization belongs T4. | T4. |
| STO-05 | Existing governed bytes never overwrite in place | **PRESERVE CANDIDATE** | Immutable Submission/Rendition content requires no-overwrite semantics; mutable WorkingContent may point to replacement content under OCC. | T4. |
| STO-06 | Local provider for dev/test + AWS S3 reference production provider | **REOPEN / PRIOR MECHANISM CHOICE** | Preserve as evidence; T4 must verify current deployment requirements/provider conformance before freezing. | T4. |
| STO-07 | Production malware inspection before admitting untrusted governed bytes | **PRESERVE CANDIDATE** | Safety requirement remains coherent and has no scope/topology contradiction, but concrete admission/scan semantics belong T4. | T4. |
| STO-08 | Object-store versioning/Object Lock/WORM are lifecycle authority | **SUPERSEDED / REJECTED** | They may be defense/enforcement only. Records-driven WORM is Future. | T4/Future Records. |
| STO-09 | Application-layer Company/Tenant DEK/crypto-erasure mandatory | **DEFERRED / REJECTED AS DEFAULT** | No named immutable encrypted Target Data requires it. | Reopen on concrete privacy/assurance need. |
| STO-10 | External repository IDs become product content IDs | **SUPERSEDED / REJECTED** | Repository connectors are Future copy/import/publish boundaries. | Future. |
| STO-11 | Backup/restore must prove semantic facts + exact required bytes coherent | **CURRENT PRODUCT OBLIGATION / T4 REALIZATION** | Restore cannot claim success with missing/corrupt governed content. | Product Contract/T4. |

---

# 15. Distribution / Periodic Review reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| LP-01 | Distribution / Read & Acknowledge | **DEFERRED — LAUNCH+** | Expected next capability; attach to Release + User/Group. Must not become AuthZ/effectivity authority. | Launch+ redesign when promoted. |
| LP-02 | Group audience snapshots concrete Users so membership drift does not rewrite historical denominator | **PRESERVE FUTURE EVIDENCE** | Strong prior semantic for future Distribution; not Launch state now. | Launch+ revalidation. |
| LP-03 | Explicit AcknowledgementRecord, not view/download, completes obligation | **PRESERVE FUTURE EVIDENCE** | Read/view/download must not silently equal acknowledgement. | Launch+ revalidation. |
| LP-04 | Periodic Review | **DEFERRED — LAUNCH+** | Attach to stable Document + exact current EFFECTIVE Revision; due/overdue must not silently alter effectivity. | Launch+ redesign. |
| LP-05 | Detailed PeriodicReviewPolicy/Record schema and outcome vocabulary | **DEFERRED EVIDENCE** | Do not instantiate Launch tables/permissions/jobs now. | Launch+ revalidation. |

---

# 16. Dossier / Evidence / Records reconciliation

These capabilities were intentionally removed from Launch by GCR A6. Their prior design is **not legacy garbage**; it is future design evidence. It must not create Launch state/backward pressure.

| ID | Prior decision | Disposition | Reconciled current meaning | Future reopen guidance |
|---|---|---|---|---|
| FUT-01 | Dossier as documentary context, never content owner/access grant | **DEFERRED FUTURE / PRESERVE EVIDENCE** | Stable Document identity is the attachment seam. | Revalidate when Dossier has named consumer. |
| FUT-02 | Evidence as independent captured-record lifecycle, not forced through Document Revision | **DEFERRED FUTURE / PRESERVE EVIDENCE** | Future Evidence may become independent owner using Organization/AuthZ + shared exact-content mechanism. | Revalidate when promoted. |
| FUT-03 | Retention/Legal Hold/Disposition separate from document lifecycle/provider storage | **DEFERRED FUTURE / PRESERVE EVIDENCE** | Attach to stable governed identities/history; expiry must not imply delete by accident. | Revalidate under future Records product contract. |
| FUT-04 | Dossier link never grants access | **PRESERVE FUTURE INVARIANT** | Context relationship must not become AuthZ authority. | Future Dossier. |
| FUT-05 | LegalHold/ObjectLock distinction | **PRESERVE FUTURE INVARIANT** | Business preservation policy != provider physical enforcement. | Future Records/T4 seam. |
| FUT-06 | No generic Record declaration/BPM/object platform | **PRESERVE FUTURE CONSTRAINT** | Future capability should still model real lifecycle rather than generic platform unless requirements change. | Future review. |
| FUT-07 | Artifact retention-root/one-root model | **SUPERSEDED** | Artifact semantic owner no longer exists; future Records must attach directly to governed semantic subjects/content facts. | Do not restore B5 Artifact-root design. |

---

# 17. Historical Migration / Interchange reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| MIG-01 | Historical Migration distinct from normal import/integration | **CURRENT** | It is go-live/cutover capability that writes truthful target-owner state. | Product/T1/T7. |
| MIG-02 | Unknown remains unknown; never fabricate native Submission/Approval/Release/User actions | **CURRENT** | Imported/native distinction is mandatory. | Product/T1/T7. |
| MIG-03 | Reliable legacy revision ordinals preserved; next native ordinal above real history | **CURRENT AT SEMANTIC LEVEL** | Exact persistence shape waits T7. `REV000` current native convention does not justify rewriting source history. | T1/T7. |
| MIG-04 | Imported exact content may exist without native Submission | **CURRENT** | Template/provenance/history must not assume native Submission exists. | T1/T7. |
| MIG-05 | Plan/dry-run/deterministic outcomes/reconciliation/idempotency/atomic semantic import unit | **PRESERVE CANDIDATE** | These are strong prior cutover mechanics consistent with current truthfulness requirement. | T7. |
| MIG-06 | `CURRENT_STATE | FULL_HISTORY` modes | **PRESERVE EVIDENCE / REOPEN T7** | Useful prior product shape, but actual legacy completeness must determine final modes. | T7. |
| MIG-07 | Generic `Interchange` semantic owner | **SUPERSEDED** | No generic integration domain. | GCR A7. |
| MIG-08 | Governed Subject Export in Launch | **DEFERRED FUTURE** | Stable semantic IDs/content descriptors preserve seam. | Future. |
| MIG-09 | Generic External Repository IMPORT/PUBLISH in Launch | **DEFERRED FUTURE** | Provider IDs remain external; future copy/import/publish boundaries. | Future. |
| MIG-10 | Detailed imported families (`RevisionOrdinalReservation`, imported governance/content tables) fixed now | **REOPEN** | T7 chooses smallest truthful form from actual source evidence. | T7. |

---

# 18. Relational / transaction substrate reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning | Next owner/gate |
|---|---|---|---|---|
| DB-01 | One PostgreSQL product-state database | **PRESERVE TECHNICAL DEFAULT** | No current requirement proves distributed semantic databases. | Later implementation review may reopen only on real trust/scale boundary. |
| DB-02 | One schema `metaldocs`, not schema-per-context | **PRESERVE TECHNICAL DEFAULT** | PostgreSQL namespace is mechanism, not semantic ownership. | Implementation spec. |
| DB-03 | UUID opaque technical IDs; business/provider identity not PK | **PRESERVE TECHNICAL DEFAULT** | Stable semantic identity remains independent of codes/provider IDs. | Implementation spec/T4/T7. |
| DB-04 | `TIMESTAMPTZ` for business instants | **PRESERVE TECHNICAL DEFAULT** | Trusted instants remain timezone-aware. | Implementation spec. |
| DB-05 | Typed FKs; no universal polymorphic business relation | **PRESERVE** | Prefer closed typed relations/unions; generic `resource_type/id` only where semantics genuinely generic such as Audit attribution. | Remaining T stages. |
| DB-06 | Cross-owner FK mutation via CASCADE/SET NULL | **PRESERVE REJECTION** | Cross-owner reference must not mutate another owner's state by FK side effect. `RESTRICT/NO ACTION` is the prior safe default. | Implementation spec. |
| DB-07 | JSONB as generic unmodeled state escape hatch | **PRESERVE REJECTION** | Bounded snapshots/provenance only where atomic semantic payload is genuinely variable. | Remaining T stages. |
| DB-08 | Global SERIALIZABLE baseline | **SUPERSEDED / REJECTED** | Current T2: READ COMMITTED + narrow serialization/CAS. | T2. |
| DB-09 | Lower layers/repositories independently commit nested semantic changes | **SUPERSEDED / REJECTED** | Composed use case owns transaction boundary; no nested semantic commits. | T2. |
| DB-10 | One local transaction may compose multiple semantic owners for one invariant | **PRESERVE/CURRENT PRINCIPLE** | Ownership separation does not forbid one local atomic transaction when one business transition needs it. | T2/T3/T5. |

---

# 19. Product/security/operations constraints reconciliation

| ID | Prior decision | Disposition | Reconciled current meaning |
|---|---|---|---|
| SEC-01 | Platform/operator/system principal gets implicit company-content access | **PRESERVE REJECTION** | Operational/platform identity is not automatic product Authorization. Any maintenance trust surface must be explicit and non-serving. |
| SEC-02 | Current runtime/provider implementation is target authority | **SUPERSEDED AS AUTHORITY** | Runtime is evidence only. |
| SEC-03 | Generic BPM/ReBAC/low-code/object/ECM platform | **CURRENT REJECTION** | No generic framework without proven consumer. |
| SEC-04 | Pooled/shared multi-customer tenancy in Launch | **DEFERRED FUTURE** | Stable Company root preserves seam; no universal partition machinery now. |
| SEC-05 | Tenant/customer lifecycle, portability deletion/export as Launch product | **DEFERRED FUTURE** | No customer-company lifecycle machinery in single-company Launch. |
| SEC-06 | Generic eDiscovery/PKI/TSA/HSM/signature/quarantine platform | **DEFERRED** | Add only on concrete assurance/security requirement. |
| SEC-07 | Restore may silently resurrect lawfully erased user-profile PII | **PRESERVE REJECTION / PROOF OBLIGATION** | Backup/restore/privacy design must reconcile lawful erasure with retained stable actor identity. | T4/T7/ops. |

---

# 20. Known future horizon — preserved, not implemented

The following are intentionally remembered and must be checked as counterexamples in every relevant T-stage:

```text
Launch+:
  Distribution / Read & Acknowledge
  Periodic Review

Future:
  Dossier
  Evidence
  Retention / Legal Hold / Disposition
  Governed Export
  External Repository IMPORT/PUBLISH
  Training/LMS
  generic/multi-document Change Control
  pooled multi-customer tenancy
  realtime coauthoring / CRDT
```

Binding law remains:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

---

# 21. Decisions that MUST NOT leak into T3–T7

This is the explicit anti-legacy list:

```text
standalone Artifact semantic owner
old 8+3 semantic ownership topology
separate Approval semantic owner
exact old 5×43 permission catalog
Tenant universal partition/RLS substrate
REV001 as initial issuance
RoleInArea Launch actor selector
configurable ANY|ALL quorum baseline
strict cross-Step same-user SoD baseline
fresh-auth/SLA/reassign/overseer as Launch defaults
DocumentType category/taxonomy platform
Tenant Dictionary/System Value platform
TemplateSpec platform
DRAFT EditorialComment platform
Periodic Review Launch state
Distribution Launch-Core state/Release atomicity
Dossier/Evidence/Records Launch owners/state
generic Interchange owner
Governed Export Launch capability
External Repository generic Launch capability
global AuditChainHead/hash-chain serialization
scheduled Release
universal PDF official representation
Artifact-rooted retention model
provider/storage identity as semantic identity
```

If a later T-stage proposes one of these, it must explicitly state the new evidence that reopens it.

---

# 22. Reopen set routed to remaining T-stages

## T3 — Authorization & Audit Enforcement

```text
exact Launch role vocabulary
exact permission vocabulary/bundles
whether/how area_manager survives
whole-company admin role naming/bundle
role↔scope matrix
access administration law for GroupMembership/RoleAssignment
Group administration/deletion exact law
offboarding exact access teardown transaction
least-privilege Governance Viewer/Auditor
canonical check sites
authorization-sensitive in-flight/offboarding races where material
same-local-commit Audit operation census + minimum bounded facts
Audit read visibility/scoping
```

T3 starts from CURRENT/PRESERVE decisions above rather than from zero.

## T4 — Exact Content, Storage Integrity & Restore

```text
exact content descriptor/digest algorithm/canonicalization
provider-neutral managed-content mechanism
provider choice/profile/conformance
staging/confirmation/admission
malware policy/scan ordering
immutable byte/no-overwrite enforcement
mutable WorkingContent recovery
backup/restore completeness + privacy non-resurrection
```

## T5 — Durable Async, Search & External Effects

```text
which effects actually require durable intent/outbox
worker/lease/retry/DLQ mechanism
renderer execution
notifications if a Launch consumer remains
Search projection/rebuild/freshness/reconciliation
provider effect receipts where needed
```

## T6 — Canonical API / Frontend Journeys

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
```

## T7 — Historical Migration & Cutover

```text
actual source evidence census
CURRENT_STATE/FULL_HISTORY or smaller real modes
imported target-owned fact shapes
ordinal/content/governance provenance
plan/dry-run/idempotency/reconciliation
semantic-unit atomicity
cutover/readiness/rollback/deletion map
```

---

# 23. Proposed governance for this registry

If operator-ratified, promote this candidate as:

`wiki/architecture/rebaseline-decision-registry.md`

Authority role:

> **Current disposition authority for prior R3–R9.5 / old R10 decisions. It tells later stages whether a prior decision is CURRENT, PRESERVE, REFINED, REOPEN, DEFERRED or SUPERSEDED. It does not duplicate the detailed semantics of Product Contract/T1/T2/later accepted T-stage authorities.**

Mandatory future-stage entry rule:

```text
before Tn design:
  read registry
  consume CURRENT/PRESERVE/REFINED
  design only REOPEN decisions owned by Tn
  keep DEFERRED as counterexample/seam
  reject SUPERSEDED inheritance
```

Registry update rule:

```text
when Tn closes:
  update dispositions affected by Tn
  point each closed decision to its new durable authority
  do not copy entire Tn narrative into registry
```

This keeps one navigable current decision baseline without creating a second giant architecture specification.

---

# 24. Operator adjudication packet

Recommended dispositions:

```text
DR-1 ACCEPT — create/promote a durable Rebaseline Decision Registry before T3 resumes.
DR-2 ACCEPT — use CURRENT / PRESERVE / REFINED / REOPEN / DEFERRED / SUPERSEDED vocabulary.
DR-3 ACCEPT — later T-stages consume CURRENT/PRESERVE/REFINED and only deliberately redesign their REOPEN set.
DR-4 ACCEPT — old `2026-08-14-cohesive-platform-redesign-ledger.md` becomes historical inventory/evidence, never current decision authority after registry promotion.
DR-5 ACCEPT — detailed future Dossier/Evidence/Records/Distribution/PeriodicReview designs remain preserved as future evidence, not deleted conceptually and not instantiated in Launch.
DR-6 ACCEPT — explicit anti-legacy list in §21 is binding unless a later material counterexample reopens one item.
DR-7 ACCEPT — T3 remains paused until this reconciliation is operator-ratified/promoted; premature T3 staging is removed/rebuilt from the accepted registry.
DR-8 ACCEPT — update the registry after every T-stage closure so it remains the current cross-stage decision baseline rather than another stale document.
```

Implementation remains **BLOCKED**.