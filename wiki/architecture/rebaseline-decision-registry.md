# MetalDocs Rebaseline Decision Registry

> **Status:** ACTIVE / OPERATOR-RATIFIED DECISION DISPOSITION AUTHORITY  
> **Ratified:** 2026-08-18  
> **Last stage update:** T5 — Durable Async, Search & External Effects — OPERATOR-RATIFIED  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

## 0. Role of this registry

The Whole-Product rebaseline changed Launch scope and semantic ownership after R3–R9.5 and old R10-A→C had already produced a large decision corpus. The operator ratified the following law:

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

This registry is the current disposition authority for those prior decisions. It is **not** a competing detailed product specification. Detailed current truth remains in the Product Contract, Whole-Product GCR, ownership topology and ratified T-stage authorities.

Before every remaining T-stage:

```text
read this registry
→ consume CURRENT / PRESERVE / REFINED as baseline
→ deliberately design only that T-stage's REOPEN set
→ keep DEFERRED as future counterexample/seam
→ reject SUPERSEDED inheritance
```

When a T-stage closes, this registry is updated to point the affected decisions to their new durable authority.

## 1. Disposition vocabulary

```text
CURRENT
  Already re-proven and ratified by current authority.

PRESERVE
  Earlier decision remains coherent and has no current counterexample.
  It is baseline unless later material evidence explicitly reopens it.

REFINED
  Essential decision survives, but old wording/shape depended on superseded scope/topology.
  Only the corrected current meaning survives.

REOPEN
  Prior decision remains useful evidence but current authority has not settled its current form.
  The named T-stage must decide it deliberately.

DEFERRED
  Expected later capability/decision; preserve its seam/evidence, create no dormant Launch state.

SUPERSEDED
  Conflicts with current authority or encoded removed ownership/scope.
  Must not be inherited without a new explicit material reopen.
```

## 2. Current authority order

```text
AGENTS.md
→ DevelopmentConexus Engineering Method v1.0.0
→ wiki/references/current-agent-handoff.md
→ wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/r10-t3-authorization-audit-enforcement.md
→ wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
→ wiki/architecture/r10-t5-durable-async-search-external-effects.md
→ this registry for prior-decision disposition
→ active T-stage candidate only for its REOPEN set
```

Old R3–R9.5/R10 files are evidence/provenance only and never override this chain.

## 3. Cross-cutting CURRENT laws

```text
one semantic authority per meaning
mechanism != authority
single company per Launch deployment
4 business owners + Audit supporting owner
Document != Revision != WorkingContent != Submission
REV000 = initial issuance; REV001 = first revision
Submission = immutable exact governed attempt
one sequential governance Step semantic
NoHumanApproval creates no fake approver
Release = system-owned effectivity authority
at most one EFFECTIVE Revision per Document
replacement Release = atomic SUPERSEDED + EFFECTIVE
explicit governed obsolescence without replacement
Search never grants access or establishes effectivity
Audit proves actions and never reconstructs current state
storage/provider identity never becomes semantic identity
imported history never becomes fake native history
future capability attaches by reference; no dormant platform machinery
one local ACID transaction per native business transition
external/provider calls never join local semantic atomicity
Document = lifecycle serialization root
WorkingContent = OCC/CAS for DRAFT races
READ COMMITTED + narrow explicit serialization/CAS = accepted default posture
current Authorization = live User/Group grants + scope + domain predicates
critical governed/security mutations cannot commit without required same-local-commit Audit
exact governed content = semantic descriptor truth + opaque replaceable storage mechanism
required future effects use transaction-coupled durable intent; queue/job state is never business authority
```

# 4. Authentication

| ID | Disposition | Current meaning |
|---|---|---|
| AUTH-01 | **CURRENT** | Authentication provider owns credentials/password/MFA/federation; MetalDocs does not. |
| AUTH-02 | **PRESERVE — mechanism selection** | Keycloak remains selected V1 AuthN provider unless concrete deployment evidence reopens it; replaceable behind anti-corruption seam. |
| AUTH-03 | **CURRENT** | Stable provider identity is `issuer + subject`; `ProviderSubjectBinding` binds it to stable MetalDocs User. |
| AUTH-04 | **CURRENT** | Provider roles/groups/org/claims never feed canonical MetalDocs AuthZ. |
| AUTH-05 | **CURRENT** | `ApplicationSession` is MetalDocs-owned and independently revocable; Session is not Role/Permission authority. |
| AUTH-06 | **PRESERVE / T2-compatible** | No atomic MetalDocs↔Keycloak transaction; provider effects are post-commit/reconciled. |
| AUTH-07 | **DEFERRED** | Fresh-auth/eSignature evidence has no named Launch consumer; Authentication remains future owner if promoted. |

# 5. Organization

| ID | Disposition | Current meaning |
|---|---|---|
| ORG-01 | **REFINED / CURRENT** | Old singleton Tenant root is now stable single `Company` root per deployment; never universal partition key. |
| ORG-02 | **CURRENT** | User is stable organizational participant identity; offboarding never rewrites history. |
| ORG-03 | **CURRENT** | `UserProfile` is separately erasable human-readable/contact enrichment. |
| ORG-04 | **CURRENT** | Area is stable flat organizational scope; no hierarchy/default-approver platform without consumer. |
| ORG-05 | **CURRENT** | Group + current GroupMembership remain first-class; no nested/dynamic/provider-mirrored groups. |
| ORG-06 | **CURRENT** | GroupMembership is current truth only; historical transition evidence belongs Audit when required. |
| ORG-07 | **SUPERSEDED AS BASELINE** | No independent `UserAreaMembership`/home-area relation without named consumer. |
| ORG-08 | **PRESERVE** | Area retirement/inactivity blocks future use as appropriate but preserves existing references/history. |
| ORG-09 | **CURRENT — T3** | Group deletion fails closed while current memberships, Group grants, current governance routes or live unactivated GROUP-step snapshots depend on it; resolved/completed Steps use concrete User snapshots. |
| ORG-10 | **CURRENT — T3** | Offboarding disables User, revokes Sessions, removes current memberships/direct grants and preserves history; re-enable restores eligibility only and never silently restores prior authority. |

# 6. Authorization

| ID | Disposition | Current meaning |
|---|---|---|
| AZ-01 | **CURRENT** | Static product-owned Role/Permission catalogs; no customer custom-role/custom-permission platform Launch. |
| AZ-02 | **CURRENT** | RoleAssignment is sole persisted current grant family; subject `User | Group`. |
| AZ-03 | **REFINED / CURRENT** | Scope is `Company | Area`; old TenantScope terminology is replaced. |
| AZ-04 | **CURRENT** | Additive grants + default deny; no deny engine/ACL graph. |
| AZ-05 | **CURRENT — T3** | Canonical evaluation uses current direct User grants + current GroupMembership→Group grants; Session/JWT/provider claims never become durable permission authority. |
| AZ-06 | **CURRENT** | Domain relationship/lifecycle/governance constraints remain with Controlled Documents, not RBAC. |
| AZ-07 | **CURRENT** | No role, including admin, bypasses domain governance. |
| AZ-08 | **REFINED / CURRENT — T3** | Launch roles are `governance_admin`, `area_manager`, `author`, `approver`, `viewer`, `governance_viewer`; exact old five-role set/bundles are superseded. |
| AZ-09 | **SUPERSEDED** | Exact old 5×43 permission matrix must not be preserved by subtraction. |
| AZ-10 | **CURRENT — T3** | `area_manager` is AreaScope-only operational management, not RBAC administration; it gains Area-wide document management through `document.owner.manage`. |
| AZ-11 | **CURRENT — T3** | `access.manage` protects GroupMembership + RoleAssignment mutations because either can change effective authority. |
| AZ-12 | **SUPERSEDED / REJECTED** | RLS is not Role/Area/permission policy engine; canonical AuthZ is application/domain authority. |
| AZ-13 | **SUPERSEDED** | Universal `tenant_id` composite PK/FK + tenant RLS substrate is incompatible with single-company Launch. |
| AZ-14 | **CURRENT — T3** | `governance_admin` = CompanyScope only; `area_manager` = AreaScope only; `author/approver/viewer/governance_viewer` = CompanyScope or AreaScope; every role may be assigned to User or Group. |
| AZ-15 | **CURRENT — T3** | Accepted Launch permission vocabulary is the 15-permission catalog owned by `wiki/architecture/r10-t3-authorization-audit-enforcement.md`. |
| AZ-16 | **CURRENT — T3** | Ordinary author working authority is bounded by current responsible-owner relationship unless actor also has `document.owner.manage`; governance verdict requires `governance.act` plus exact active-Step participation and T2 predicates. |

# 7. Governance

| ID | Disposition | Current meaning |
|---|---|---|
| GOV-01 | **REFINED / CURRENT** | Specialized sequential governance survives inside Controlled Documents; not generic BPM/separate Approval owner. |
| GOV-02 | **SUPERSEDED** | Separate Approval semantic owner/context removed. |
| GOV-03 | **SUPERSEDED AS REQUIREMENT** | Mandatory first-class versioned ApprovalPolicy aggregate removed; current route config + immutable per-attempt snapshot is enough absent new consumer. |
| GOV-04 | **SUPERSEDED** | Structural `REVIEW | APPROVAL` Step purpose removed; one Step semantic + business label. |
| GOV-05 | **REFINED / CURRENT** | Launch actor selector is `NAMED_USER | GROUP`; `ROLE_IN_AREA` not baseline. |
| GOV-06 | **REFINED / CURRENT** | Group Step = ANY-one from enabled membership snapshot at activation; ALL/N-of-M deferred. |
| GOV-07 | **REFINED / CURRENT** | Group membership resolves to concrete active-Step candidates; later drift does not rewrite active denominator; current AuthZ still rechecked at action. |
| GOV-08 | **REFINED / CURRENT** | Submitter/obsolescence initiator cannot satisfy human Step on same attempt; cross-Step same-user prohibition deferred. |
| GOV-09 | **DEFERRED** | Fresh-auth per Step. |
| GOV-10 | **DEFERRED** | Due dates/SLA/escalation. |
| GOV-11 | **DEFERRED** | Reassign/overseer/delegation; Launch escape is withdraw→fix route→resubmit. |
| GOV-12 | **REFINED** | Withdraw Submission and cancel Revision are separate explicit meanings; no generic approval-cancel platform. |
| GOV-13 | **CURRENT** | NoHumanApproval creates no fake approver, including accepted no-human obsolescence. |
| GOV-14 | **SUPERSEDED / REJECTED** | GovernanceAttempt has closed subject universe `SUBMISSION | OBSOLESCENCE`; no arbitrary subject registry. |

# 8. Controlled Document / Revision

| ID | Disposition | Current meaning |
|---|---|---|
| DOC-01 | **CURRENT** | Stable Document identity across Revisions. |
| DOC-02 | **CURRENT** | Revision = business change cycle; autosave != Revision; `REV000` initial. |
| DOC-03 | **SUPERSEDED** | `REV001` as initial issuance is invalid; `REV000` is binding initial issuance. |
| DOC-04 | **CURRENT** | Revision states remain `DRAFT | SUBMITTED | EFFECTIVE | SUPERSEDED | OBSOLETE | CANCELLED`. |
| DOC-05 | **CURRENT** | At most one open and one EFFECTIVE Revision; prior EFFECTIVE may coexist with newer DRAFT/SUBMITTED. |
| DOC-06 | **SUPERSEDED AS LAUNCH REQUIREMENT** | Mandatory reason-for-change at `REV002+` is not current Product Contract requirement. |
| DOC-07 | **CURRENT** | Document code/type/Area stable; transfer not invented. |
| DOC-08 | **CURRENT** | Responsible owner is current mutable Document relationship; historical actions remain actor-bound. |
| DOC-09 | **SUPERSEDED / REJECTED** | No second Document `currentRevision/currentStatus` authority; current-effective truth comes from Revision lifecycle/Release/Obsolescence. |
| DOC-10 | **DEFERRED** | DocumentType category/taxonomy platform. |
| DOC-11 | **DEFERRED** | Editable Dictionary/System Value platform. |
| DOC-12 | **REOPEN / PRIOR EVIDENCE** | Numbering capability/non-reuse is current; exact `{TYPE}/{AREA}/{SEQ}`, TYPE/TYPE_AREA/padding grammar must be revalidated in T6/admin/API design. |
| DOC-13 | **CURRENT / PRESERVE** | Number allocation monotonic; committed codes never reuse; preview reserves nothing. |

# 9. Working Content / Submission / Template

| ID | Disposition | Current meaning |
|---|---|---|
| CNT-01 | **CURRENT** | WorkingContent = sole mutable DRAFT authority; editor/provider never authority. |
| CNT-02 | **CURRENT** | One monotonic OCC generation; no silent last-write-wins. |
| CNT-03 | **REOPEN / T6 UX MECHANISM** | T4 ratified that EditorSession is not a correctness dependency; T6 may still add a bounded editor lease only if UX/integration evidence requires it. |
| CNT-04 | **SUPERSEDED / REJECTED** | WorkingSnapshot is never business history; T4 ratified current WorkingContent as the Launch DRAFT recovery baseline. |
| CNT-05 | **CURRENT** | Submission immutable/exact; same Revision may have multiple attempts. |
| CNT-06 | **CURRENT** | Submission freezes coherent exact WorkingContent generation + decision-relevant governed config/state. |
| CNT-07 | **CURRENT — T4** | Exact-content descriptor is `SHA-256 + exact size + closed ContentFormat`; no mandatory whole-Submission JCS/composite digest Launch. |
| CNT-08 | **SUPERSEDED** | Standalone Artifact exact-byte owner removed; T4 confirms no Artifact/ArtifactStage semantic ownership. |
| CNT-09 | **CURRENT** | Template = ordinary governed Document role; no parallel lifecycle. |
| CNT-10 | **CURRENT** | Template eligibility remains small M:N current relation to target DocumentTypes. |
| CNT-11 | **REFINED / CURRENT** | Derived origin pins source Template Document + exact EFFECTIVE Revision/content provenance without mandatory native Submission. |
| CNT-12 | **DEFERRED** | Structured TemplateSpec platform. |
| CNT-13 | **DEFERRED** | DRAFT EditorialComment platform; SubmissionFeedback remains required. |
| CNT-14 | **PRESERVE — mechanism selection** | EigenPal may remain selected DOCX adapter/provider evidence; never owns semantic truth. |
| CNT-15 | **DEFERRED** | Realtime Yjs/CRDT; seam remains WorkingContent concurrency boundary. |

# 10. Rendition / Release / Effectivity

| ID | Disposition | Current meaning |
|---|---|---|
| REL-01 | **CURRENT / T5-CLARIFIED** | `SourceOnly` vs one required official representation remains current; ordinary preview/viewing is mechanism only and never semantic Rendition. PDF source is viewed directly by default; SourceOnly DOCX does not require persisted PDF. |
| REL-02 | **CURRENT / T4+T5-CLARIFIED** | OfficialRendition binds exact Submission, freezes its own ExactContentDescriptor + managed-content handle, and is rendered durably only when the frozen Submission representation policy requires it. |
| REL-03 | **SUPERSEDED** | Universal mandatory PDF removed; universal DOCX→PDF persistence for viewing is also not Launch baseline. |
| REL-04 | **DEFERRED** | Scheduled/future-dated Release. |
| REL-05 | **SUPERSEDED / REJECTED** | No human publish button; system Release establishes effectivity. |
| REL-06 | **CURRENT** | First/replacement Release atomicity. |
| REL-07 | **CURRENT** | Release waits only on human gate + required rendition gate. |
| REL-08 | **SUPERSEDED FOR LAUNCH** | Distribution obligations are outside Launch-Core Release atomicity. |

# 11. Obsolescence

| ID | Disposition | Current meaning |
|---|---|---|
| OBS-01 | **CURRENT** | Explicit governed withdrawal without successor; raw status toggle invalid. |
| OBS-02 | **CURRENT** | Target remains EFFECTIVE during human governance; only successful completion makes OBSOLETE. |
| OBS-03 | **SUPERSEDED** | NoHumanApproval obsolescence may complete with zero human Step after authorization/reason/invariant checks. |
| OBS-04 | **REFINED / CURRENT** | No open replacement at initiation; active obsolescence blocks new Revision. |
| OBS-05 | **DEFERRED** | Separate obsolescence route; Launch reuses same DocumentType route. |
| OBS-06 | **DEFERRED / OUT OF LAUNCH** | Reactivation of OBSOLETE. |

# 12. Search / Notifications / Async

| ID | Disposition | Current meaning |
|---|---|---|
| ASY-01 | **CURRENT — T5** | Search = one PostgreSQL-backed rebuildable/eventually-consistent discovery projection keyed by Document; `search_refresh(document_id)` reloads latest canonical state, and current lifecycle/AuthZ are always re-resolved before serve. |
| ASY-02 | **DEFERRED / PRESERVE FUTURE** | Notifications remain a delivery projection concept but have no mandatory Launch inbox/email/push consumer; add only a named delivery job/projection when a concrete capability requires it. |
| ASY-03 | **CURRENT — T5** | If a semantic transaction creates a required future effect, durable job/intent is inserted in the same local transaction; provider/network execution stays outside and intent is mechanism only. |
| ASY-04 | **PRESERVE — T5 MECHANISM SELECTION** | Launch uses one PostgreSQL-backed durable-job mechanism; River remains selected/reference implementation, replaceable without changing semantic meaning. |
| ASY-05 | **SUPERSEDED / REJECTED** | No global SERIALIZABLE/global worker-lock framework baseline and no parallel hand-rolled scheduler/lease/outbox runtime beside the selected job mechanism. |
| ASY-06 | **CURRENT — T5** | Managed-content cleanup is periodic reconciliation over durable T4 `GC_PENDING`; immediately re-prove no live/governed/backup-protected reference before physical delete; no per-handle durable outbox baseline. |
| ASY-07 | **CURRENT — T5** | `official_rendition_render` is a conditional durable job only when frozen representation policy requires OfficialRendition; ordinary PDF/DOCX viewing creates no durable render requirement. |
| ASY-08 | **CURRENT — T5** | Durable jobs are at-least-once, idempotent/revalidating, bounded-retry, terminal-visible/redrivable and carry bounded stable IDs/minimum immutable routing facts rather than business/AuthZ snapshots. |
| ASY-09 | **CURRENT — T5** | Full Search rebuild/reconciliation is mandatory proof that Search is derivative; an always-on global reconciliation crawler is not Launch baseline. |
| ASY-10 | **CURRENT — T5** | No generic domain-event bus or `ExternalEffectReceipt` family Launch; required async mechanisms expose minimum health/backlog/retry/terminal-failure observability. |

# 13. Audit

| ID | Disposition | Current meaning |
|---|---|---|
| AUD-01 | **CURRENT** | AuditEvent = transversal evidence, never current state/event sourcing. |
| AUD-02 | **CURRENT — T3** | Exact same-local-commit Audit census and bounded minimum facts are ratified in `wiki/architecture/r10-t3-authorization-audit-enforcement.md`. |
| AUD-03 | **CURRENT** | Audit append-only and PII-minimized; no credential/token/request-body/free-form reason copying by convenience. |
| AUD-04 | **SUPERSEDED / DEFERRED** | Deployment-wide AuditChainHead/hash-chain/global terminal lock removed absent concrete assurance requirement. |
| AUD-05 | **DEFERRED** | Generic Audit export permission/capability. |
| AUD-06 | **REOPEN / FUTURE COMPLIANCE** | T3 intentionally does not claim indefinite/statutory retention; future Records/compliance requirement defines retention/pruning/checkpoint semantics. |
| AUD-07 | **SUPERSEDED / REJECTED** | DB CRUD triggers do not infer semantic Audit; explicit use-case composition records operation meaning. |
| AUD-08 | **CURRENT — T3** | Audit visibility is immutable `Company` or `Area` attribution captured at event time; `audit.read @ Company` sees company+areas, `audit.read @ Area` sees only that Area's events. |
| AUD-09 | **CURRENT — T3** | Ordinary autosave/search/read/download/login/logout/notification/preview/deny events are telemetry, not mandatory semantic Audit absent a named compliance requirement. |

# 14. Storage / Exact Content / Restore

| ID | Disposition | Current meaning |
|---|---|---|
| STO-01 | **CURRENT** | Storage/provider identity never semantic identity. |
| STO-02 | **CURRENT — T4** | Provider-neutral shared `ManagedContentStore` mechanism serves semantic owners without becoming an owner; Launch uses one active store per deployment. |
| STO-03 | **SUPERSEDED / T4-REFINED** | `ManagedArtifactStore`/Artifact-owned surface is removed; T4 preserves only provider-neutral managed-content mechanism with opaque handle. |
| STO-04 | **CURRENT — T4** | ExactContentDescriptor = SHA-256 + exact size + closed ContentFormat; client/provider metadata is not semantic exactness. |
| STO-05 | **CURRENT — T4** | Managed handles are create-once/no-overwrite; DRAFT replacement creates a new handle. |
| STO-06 | **CURRENT — T4** | Local = dev/test/conformance provider; AWS S3 = reference production provider; alternate providers must pass the same contract. |
| STO-07 | **CURRENT — T4** | UNTRUSTED_EXTERNAL exact bytes require CLEAN malware proof before immutable governed admission in production; autosaves are not scanned on every debounce; ClamAV/clamd is reference MalwareInspector. |
| STO-08 | **SUPERSEDED / REJECTED** | Object-store Versioning/Object Lock/WORM are defense/enforcement only, never lifecycle authority. |
| STO-09 | **DEFERRED / REJECTED AS DEFAULT** | Application-layer Company DEK/crypto-erasure is not mandatory absent named Target Data/assurance requirement. |
| STO-10 | **SUPERSEDED / REJECTED** | External repository IDs never become MetalDocs content identity. |
| STO-11 | **CURRENT — T4** | Restore is non-serving until every required semantic content reference resolves and exact size/SHA-256/format match; missing/corrupt required content fails closed. |
| STO-12 | **CURRENT — T4** | `managed_content_id` is an opaque retrieval handle only; OPEN→READY is durable mechanism state, not semantic Artifact confirmation. |
| STO-13 | **CURRENT — T4** | Server independently derives descriptor from stored bytes; provider PUT/client claims never equal semantic admission. |
| STO-14 | **CURRENT — T4** | Opaque server-owned admission binding prevents arbitrary/cross-root handle reuse without introducing a generic owner registry. |
| STO-15 | **CURRENT — T4** | SUBMIT/Rendition semantic transactions make zero provider/scanner calls; they revalidate READY/descriptor/binding/malware proof and freeze exact handle+descriptor atomically with semantic truth. |
| STO-16 | **CURRENT — T4** | Only unreferenced/non-governed mechanism content may be reclaimed; immutable governed content has no ordinary Launch delete/disposition path. |
| STO-17 | **CURRENT — T4** | Backup = consistent DB recovery point + exact required-content manifest/copy, including current DRAFT WorkingContent; capture must exclude GC races through bounded pin/equivalent. |
| STO-18 | **CURRENT — T4** | Historical restore must reconcile post-snapshot lawful UserProfile erasures through independently retained erasure barrier/journal or equivalent proof before serving human-readable profile data. |
| STO-19 | **DEFERRED — T4** | Whole-Submission canonical JCS/composite digest remains deferred until a named signing/export/non-repudiation consumer exists. |

# 15. Launch+ — Distribution / Periodic Review

| ID | Disposition | Future meaning |
|---|---|---|
| LP-01 | **DEFERRED — LAUNCH+** | Distribution / Read & Acknowledge attaches to Release + User/Group; never AuthZ/effectivity authority. |
| LP-02 | **PRESERVE FUTURE EVIDENCE** | Group audience resolves concrete Users so membership drift never rewrites historical denominator. |
| LP-03 | **PRESERVE FUTURE EVIDENCE** | Explicit AcknowledgementRecord completes obligation; view/download never silently equals acknowledgement. |
| LP-04 | **DEFERRED — LAUNCH+** | Periodic Review attaches to stable Document + exact current EFFECTIVE Revision; due/overdue never silently changes effectivity. |
| LP-05 | **DEFERRED EVIDENCE** | Detailed PeriodicReviewPolicy/Record schema/outcomes are future evidence only. |

# 16. Future — Dossier / Evidence / Records

These designs are **not legacy garbage**. They are future evidence and must create no Launch backward pressure.

| ID | Disposition | Future meaning |
|---|---|---|
| FUT-01 | **DEFERRED FUTURE / PRESERVE EVIDENCE** | Dossier = documentary context, never content owner/access grant; stable Document is seam. |
| FUT-02 | **DEFERRED FUTURE / PRESERVE EVIDENCE / T4 SEAM** | Evidence may own its own exact descriptor + managed-content handle while retaining independent lifecycle; no Artifact owner required. |
| FUT-03 | **DEFERRED FUTURE / PRESERVE EVIDENCE** | Retention/Hold/Disposition remain separate from document lifecycle/provider storage; expiry must not imply delete. |
| FUT-04 | **PRESERVE FUTURE INVARIANT** | Dossier link never grants access. |
| FUT-05 | **PRESERVE FUTURE INVARIANT** | LegalHold business preservation != ObjectLock/provider physical enforcement. |
| FUT-06 | **PRESERVE FUTURE CONSTRAINT** | Do not build generic Record/BPM/object platform without changed requirements. |
| FUT-07 | **SUPERSEDED** | Artifact retention-root/one-root design does not survive removal of Artifact semantic owner. |

# 17. Historical Migration / Interchange

| ID | Disposition | Current/future meaning |
|---|---|---|
| MIG-01 | **CURRENT** | Historical Migration is distinct go-live/cutover capability, not normal integration. |
| MIG-02 | **CURRENT** | Unknown remains unknown; never fabricate native Submission/decision/Release/User actions. |
| MIG-03 | **CURRENT AT SEMANTIC LEVEL** | Preserve reliable legacy ordinals; next native ordinal above real history; current `REV000` convention never rewrites source history. |
| MIG-04 | **CURRENT / T4-CLARIFIED** | Imported exact content may exist without native Submission and uses the same exact descriptor + managed-content admission seam; T7 owns imported fact shapes. |
| MIG-05 | **PRESERVE CANDIDATE** | Plan/dry-run/deterministic outcomes/reconciliation/idempotency/atomic semantic import unit remain strong cutover evidence. |
| MIG-06 | **PRESERVE EVIDENCE / REOPEN T7** | `CURRENT_STATE | FULL_HISTORY` modes are useful evidence but actual source completeness decides final modes. |
| MIG-07 | **SUPERSEDED** | Generic Interchange semantic owner removed. |
| MIG-08 | **DEFERRED FUTURE / T4 SEAM** | Governed Subject Export may derive future manifests/digests from stable semantic descriptors without provider identity becoming product truth. |
| MIG-09 | **DEFERRED FUTURE / T4 SEAM** | Generic External Repository IMPORT/PUBLISH remains Future and reuses exact descriptor/copy mechanisms without making external object IDs MetalDocs identity. |
| MIG-10 | **REOPEN** | Detailed imported target families are not frozen; T7 derives smallest truthful shape from actual source evidence. |

# 18. Relational / Transaction substrate

| ID | Disposition | Current meaning |
|---|---|---|
| DB-01 | **PRESERVE TECHNICAL DEFAULT** | One PostgreSQL product-state DB absent real distributed trust/scale boundary. |
| DB-02 | **PRESERVE TECHNICAL DEFAULT** | One `metaldocs` schema; DB namespace is mechanism, not semantic ownership. |
| DB-03 | **PRESERVE TECHNICAL DEFAULT** | UUID opaque technical IDs; business/provider identities never PK authority. |
| DB-04 | **PRESERVE TECHNICAL DEFAULT** | `TIMESTAMPTZ` for trusted business instants. |
| DB-05 | **PRESERVE** | Typed FKs/closed unions; avoid universal polymorphic business relations except genuinely generic semantics such as Audit attribution. |
| DB-06 | **PRESERVE REJECTION** | Cross-owner FKs do not mutate another owner's state via CASCADE/SET NULL; `RESTRICT/NO ACTION` safe default. |
| DB-07 | **PRESERVE REJECTION** | JSONB is not unmodeled-state escape hatch; only bounded snapshots/provenance when semantics justify variability. |
| DB-08 | **SUPERSEDED / REJECTED** | Global SERIALIZABLE baseline removed; T2 narrow serialization/CAS controls. |
| DB-09 | **SUPERSEDED / REJECTED** | Lower repositories/layers do not independently commit nested semantic changes; composed use case owns transaction. |
| DB-10 | **PRESERVE / CURRENT PRINCIPLE** | One local transaction may compose multiple semantic owners when one invariant/business transition requires it. |

# 19. Product / Security / Operations constraints

| ID | Disposition | Current meaning |
|---|---|---|
| SEC-01 | **PRESERVE REJECTION** | Platform/operator/system principal gets no implicit company-content access; maintenance trust surface must be explicit/non-serving. |
| SEC-02 | **SUPERSEDED AS AUTHORITY** | Current runtime/provider implementation is evidence only. |
| SEC-03 | **CURRENT REJECTION** | No generic BPM/ReBAC/low-code/object/ECM platform without proven consumer. |
| SEC-04 | **DEFERRED FUTURE** | Pooled/shared multi-customer tenancy not Launch; stable Company root preserves seam. |
| SEC-05 | **DEFERRED FUTURE** | Customer-company lifecycle/portability deletion/export not Launch. |
| SEC-06 | **DEFERRED** | Generic eDiscovery/PKI/TSA/HSM/signature/quarantine platform absent concrete requirement. |
| SEC-07 | **CURRENT — T4 RESTORE LAW** | Restore may not silently resurrect lawfully erased UserProfile PII; post-snapshot erasures must be reconciled before serving restored profile data. |

# 20. Known future horizon

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

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

# 21. Anti-legacy list — MUST NOT leak into T6–T7

```text
standalone Artifact semantic owner
ArtifactStage business model
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
universal DOCX→PDF persistence merely for viewing
Artifact-rooted retention model
provider/storage identity as semantic identity
whole-Submission JCS digest without a named consumer
multi-cloud/BYOS/active-active content platform
Object Lock/WORM as lifecycle authority
generic domain-event/integration platform without a named consumer
generic ExternalEffectReceipt family
parallel custom async schedulers/lease frameworks beside the selected job runtime
```

A later stage may propose one only by naming new material evidence that explicitly reopens it.

# 22. Closed T-stages and official REOPEN set for remaining stages

## T3 — Authorization & Audit Enforcement — CLOSED / OPERATOR-RATIFIED

Detailed authority:

`wiki/architecture/r10-t3-authorization-audit-enforcement.md`

## T4 — Exact Content, Storage Integrity & Restore — CLOSED / OPERATOR-RATIFIED

Detailed authority:

`wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`

## T5 — Durable Async, Search & External Effects — CLOSED / OPERATOR-RATIFIED

Detailed authority:

`wiki/architecture/r10-t5-durable-async-search-external-effects.md`

Closed decisions include:

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
search_refresh(document_id) always-required durable projection job
official_rendition_render conditional only for policy-required OfficialRendition
ordinary PDF/DOCX viewing creates no durable-rendition requirement
required durable intent transaction-coupled with the semantic fact that creates it
provider/renderer execution outside semantic transaction
OfficialRendition finalization revalidates T4/T2/T3 state and is idempotent
Search = PostgreSQL rebuildable projection keyed by Document
latest-state refresh makes duplicate/out-of-order jobs converge safely
Search may lag by omission but never grants stale authority/effectivity
full Search rebuild/reconciliation required; always-on crawler not baseline
managed-content GC = periodic reconciliation over GC_PENDING with immediate recheck
no mandatory Launch notifications/inbox/fanout/domain-event bus
no mandatory durable external IdP-disable job
jobs at-least-once/idempotent/revalidating/bounded-retry/terminal-visible with bounded-ID payloads
no generic ExternalEffectReceipt family
minimum async operational observability required
future capabilities add only named jobs/effects/receipts for proven consumers
```

## T6 — Canonical API / Frontend Journeys — NEXT REOPEN SET

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
EditorSession/UX lease only if a real editor-integration consumer requires it
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
concrete restore/erasure reconciliation choreography where cutover/recovery requires it
```

# 23. Registry governance — OPERATOR-RATIFIED DR-1→DR-8

```text
DR-1 ACCEPT — durable Rebaseline Decision Registry exists before T3 resumes.
DR-2 ACCEPT — disposition vocabulary = CURRENT / PRESERVE / REFINED / REOPEN / DEFERRED / SUPERSEDED.
DR-3 ACCEPT — later T-stages consume CURRENT/PRESERVE/REFINED and deliberately redesign only their REOPEN set.
DR-4 ACCEPT — old 2026-08-14 cohesive redesign ledger is historical inventory/evidence, never current decision authority.
DR-5 ACCEPT — future Dossier/Evidence/Records/Distribution/PeriodicReview designs remain preserved as future evidence, not conceptually discarded and not instantiated in Launch.
DR-6 ACCEPT — anti-legacy list is binding absent later explicit material reopen.
DR-7 ACCEPT — premature pre-registry T3 candidate was discarded; ratified T3 was rebuilt from this registry.
DR-8 ACCEPT — update this registry after every T-stage closure so it remains the current cross-stage baseline.
```

Implementation remains **BLOCKED**.