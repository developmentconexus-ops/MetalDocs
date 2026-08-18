# R10-B4 — Approval + CI-owned Rendition/Release + Distribution — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B4 or amend the frozen R3–R9.5 ledger. It consumes the operator-accepted non-final B3 candidate and must remain challengeable by B5–F and Whole-R10 review.

---

# 1. Authority and evidence boundary

Authority path:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md` — accepted non-final integration input

Current code/schema/OpenAPI/legacy ADRs are evidence only.

External references are comparison evidence only. The useful pattern is narrow: mature DMS/QMS products separate document/version identity, review/approval work, viewer/annotations, and publication/effectivity. MetalDocs does not import a generic workflow/BPM engine.

---

# 2. Operator refinement carried into B4

The operator explicitly requires:

1. governed documents must be viewable **inside MetalDocs**; supported viewing must not require downloading a file and opening another application;
2. approval and review journeys must be able to view the exact candidate as PDF when a PDF rendition exists/is generated;
3. collaborative review must support comments and a suggesting/redline-like interaction;
4. this capability must be provider-neutral — current authoring/editor technology is an adapter, never product/domain identity.

Consequent architectural distinction:

```text
OfficialRepresentationPolicy
!=
what the UI is capable of previewing
```

A DocumentType may be `SourceOnly` for official Release and still have an auxiliary PDF `Rendition` for in-product viewing. Such a rendition does not become official merely because the viewer uses it.

Likewise:

```text
review suggesting UX
!=
mutation of RevisionSubmission
```

While the Revision is `SUBMITTED`, the exact Submission remains immutable. Review comments/proposed changes are a **detached review-feedback layer bound to that Submission**. If the attempt is returned to DRAFT, applying a suggestion becomes an ordinary B3 `WorkingContent` OCC mutation and therefore produces new working state before a new Submission.

---

# 3. Evidence → Known / Inferred / Unknown / Deferred

## 3.1 Known — frozen/promoted inputs

B4 must preserve:

- Approval is specialized sequential human workflow, not generic BPM;
- `ApprovalPolicy(version)` has ordered Steps;
- Step fields: `purpose = review|approval`, `actor_rule = NamedUser|Group|RoleInArea`, `completion = ANY|ALL`, `requires_reauthentication`, optional `due_in_days`;
- human decision outcomes are `accept | return_for_changes`;
- `withdraw | cancel | reassign` are separate operations;
- there is no normal terminal reject V1;
- strict SoD: creator/submitter cannot accept own Submission; same User cannot accept two Steps of one ApprovalInstance; reassignment must stay qualified and SoD-valid;
- `tenant_owner` is never a bypass;
- ApprovalInstance binds exactly one immutable `RevisionSubmission`;
- return/withdraw terminates the attempt and returns the same Revision to DRAFT;
- resubmission creates a new Submission and new ApprovalInstance when human approval applies;
- fresh-auth consumes Authentication-owned assurance evidence; Approval never stores/challenges passwords;
- Rendition is immutable derived representation of exact Submission with output hash + generator/build provenance;
- Approval approves Submission, never renderer output bytes;
- Release is automatic/system-owned; no publish button;
- optional `ReleasePlan.not_before`; actual `effective_at = released_at`;
- winning Release atomically makes candidate EFFECTIVE and prior Revision SUPERSEDED;
- representation policy = `SourceOnly | RequireRendition(ContentFormat)`;
- at most one required derived rendition V1; source Artifact always remains exact Submission source;
- Distribution is controlled obligation/acknowledgement, not AuthZ/LMS;
- Release snapshots concrete Users; later Group membership never rewrites historical denominator;
- explicit immutable AcknowledgementRecord completes obligation; notification read/view/download never completes it;
- B2 Group live-reference law applies to ApprovalPolicy and Distribution configuration;
- B3 exact Submission identity, OCC generation consumption, one-open/one-effective constraints and exact Artifact semantics remain integration inputs.

## 3.2 Inferred — candidate technical choices

1. `RevisionSubmission.id` is the shared exact-candidate identity across Approval, Rendition and Release. No second `release_generation` identity is introduced.
2. ApprovalPolicy has stable identity plus immutable numbered versions; in-flight instances pin one exact version.
3. Step participants are resolved to concrete Users when a step becomes active, not kept as a live Group/Role expression for historical completion.
4. Authorization stays live at action time; participant snapshot never becomes an AuthZ snapshot.
5. Review feedback is detached immutable/submission-bound state. It may contain comments or proposed changes but never mutates Submission bytes.
6. `SUGGESTION` feedback is allowed only on `purpose=review`; ordinary COMMENT feedback may be added during review or approval.
7. A PDF Rendition may exist for preview even when `OfficialRepresentationPolicy=SourceOnly`; release only requires one when the policy says so.
8. ReleaseRecord points to the exact Rendition used as official representation when one is required; post-release effective-document viewing must use that exact rendition when present.
9. Distribution V1 live audience configuration includes Group because B2 proves it as a real consumer; unproven generic User/Area/company audience unions are not added yet.
10. Distribution obligations are created in the winning Release transaction, not asynchronously after Release, so the denominator is historical truth rather than a later membership projection.
11. Acknowledgement V1 is an explicit authenticated product action but does not require fresh-auth/e-signature absent a new requirement.
12. `due_in_days` is deadline/attention metadata only; it does not auto-escalate, auto-accept, auto-return or mutate quorum.
13. Approval `cancel` terminates the attempt and returns the Revision to DRAFT; cancelling the business Revision itself remains `document.cancel_revision`.

## 3.3 Unknown — bounded

- whether future Distribution needs direct User, Area or whole-company live audience types in addition to Group;
- whether review feedback needs mandatory per-item resolution/disposition before resubmit;
- whether approval-stage comments should support geometric/text anchors beyond general Submission comments;
- whether post-release re-rendering of a SourceOnly preview PDF needs a retained history policy;
- whether future compliance requires fresh-auth/e-signature for Acknowledgement;
- exact provider-neutral review-anchor schemas for every future content format;
- exact viewer implementation/provider selection.

None currently requires B2/B3 reopen.

## 3.4 Deferred

```text
Evidence/Dossier/Records + retention/hold/disposition        → B5
Audit skeleton/final cross-owner same-commit matrix          → B6
physical Artifact storage/malware/restore                    → R10-C
worker/outbox/timers/notification delivery/projections       → R10-D
API/frontend/viewer/editor/provider adapter journeys         → R10-E
historical migration/cutover/legacy deletion                 → R10-F
```

R10-E has a **hard successor requirement**: for supported formats/representations, users can view governed content in-product; approval/review does not require download-and-open-external-app as the normal viewing journey.

---

# 4. Root Cause

The failure class is not “missing approval tables”. It is **collapsed governance stages**:

```text
human decision
+ derived representation
+ effectivity
+ distribution obligation
```

can be incorrectly represented by one floating document status or by multiple paths that each infer “current content”. That permits approval, renderer, publication and reader obligations to refer to different candidates.

B4 must make these truths separate but causally linked through one immutable Submission.

---

# 5. Target invariant

> Every human Approval decision, every derived Rendition selected for official use, and every ReleaseRecord refers to the same immutable RevisionSubmission. A winning Release is the sole transition that makes its Revision EFFECTIVE, supersedes the prior EFFECTIVE Revision, and freezes the concrete Distribution obligations created by that effectivity event. Viewer/review interaction may add detached feedback or derived preview representations but can never mutate or replace the submitted candidate.

Corollaries:

```text
ApprovalDecision   != content mutation
ReviewFeedback     != WorkingContent
Rendition          != source Submission
PDF preview        != automatically official representation
Approval complete  != Release complete
Release            != human publish action
Distribution       != Authorization
View/download      != Acknowledgement
Group membership   != historical obligation denominator
```

---

# 6. Credible alternatives

## A — one status-heavy DocumentRevision workflow

Encode `under_review/approved/published/read` and related fields directly on Revision.

**Reject / Local Maximum.** It cannot prove exact-candidate coherence and mixes incompatible mutation laws.

## B — generic BPM/workflow platform

Model Approval, review, distribution, periodic review and possibly retention as generic nodes/transitions/conditions.

**Reject / overengineered non-maximum.** Solves hypothetical workflow classes and creates a new low-code authority.

## C — specialized Approval + detached review feedback + immutable Rendition + automatic Release + concrete Distribution snapshot

**Recommended Global Maximum.** Each necessary fact gets one authority while all four stages share the same exact Submission identity.

---

# 7. Approval configuration

## 7.1 DocumentTypeApprovalConfig

Approval-owned live binding from Controlled Information type to Approval behavior:

```text
DocumentTypeApprovalConfig
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_HUMAN_APPROVAL|USE_POLICY
  approval_policy_id UUID NULL FK ApprovalPolicy(id) RESTRICT
```

CHECK:

```text
NO_HUMAN_APPROVAL → approval_policy_id IS NULL
USE_POLICY        → approval_policy_id IS NOT NULL
```

Tenant-wide `approval_policy.manage` governs configuration.

## 7.2 ApprovalPolicy

```text
ApprovalPolicy
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  created_at TIMESTAMPTZ NOT NULL
```

Stable identity/current eligibility only. Policy change never rewrites historical versions.

## 7.3 ApprovalPolicyVersion

```text
ApprovalPolicyVersion
  id UUID PRIMARY KEY
  approval_policy_id UUID NOT NULL FK ApprovalPolicy(id) RESTRICT
  version_no INTEGER NOT NULL CHECK version_no >= 1
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(approval_policy_id, version_no)
```

Immutable/append-only. Current version = highest committed version of an ACTIVE policy. Reverting means creating a new version, never reviving/mutating an older row.

## 7.4 ApprovalPolicyStep

```text
ApprovalPolicyStep
  id UUID PRIMARY KEY
  policy_version_id UUID NOT NULL FK ApprovalPolicyVersion(id) RESTRICT
  step_order INTEGER NOT NULL CHECK step_order >= 1
  purpose TEXT NOT NULL CHECK REVIEW|APPROVAL
  completion TEXT NOT NULL CHECK ANY|ALL
  requires_reauthentication BOOLEAN NOT NULL
  due_in_days INTEGER NULL CHECK due_in_days > 0

  actor_kind TEXT NOT NULL CHECK NAMED_USER|GROUP|ROLE_IN_AREA
  named_user_id UUID NULL FK User(id) RESTRICT
  group_id UUID NULL FK Group(id) RESTRICT           // live/current version relation only
  role_code TEXT NULL
  area_id UUID NULL FK Area(id) RESTRICT

  UNIQUE(policy_version_id, step_order)
```

Closed-union CHECK:

```text
NAMED_USER   → only named_user_id
GROUP        → only group_id
ROLE_IN_AREA → only role_code + area_id
```

New/current policy configuration rejects disabled NamedUser, deleted Group and retired Area. Role vocabulary comes from B2 static catalog; runtime action still requires `approval.act`.

### Historical Group-reference lifecycle

B2 requires Group hard-delete to fail only while a **live typed reference** exists. Therefore B4 implementation specification must distinguish current policy live references from immutable historical policy/instance evidence. Superseding/deactivating the live version removes its Group FK dependency only after all in-flight instances have concrete participant snapshots; historical interpretation must not require Group row survival.

The exact normalized live-reference table/transition mechanism is implementation-spec work, but a permanent FK from every historical PolicyVersion to Group is **not** acceptable because it would silently convert Group hard-delete into never-delete.

---

# 8. Approval execution

## 8.1 ApprovalInstance

```text
ApprovalInstance
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL UNIQUE FK RevisionSubmission(id) RESTRICT
  policy_version_id UUID NOT NULL FK ApprovalPolicyVersion(id) RESTRICT
  status TEXT NOT NULL CHECK ACTIVE|ACCEPTED|RETURNED|WITHDRAWN|CANCELLED
  submitted_by_user_id UUID NOT NULL FK User(id) RESTRICT
  started_at TIMESTAMPTZ NOT NULL
  completed_at TIMESTAMPTZ NULL
```

No duplicated document/revision/hash identity. They are derived through Submission.

`NoHumanApproval` creates no fake human instance. Release eligibility treats the human gate as satisfied by configuration while still requiring the B3 Submission.

## 8.2 ApprovalStepInstance

```text
ApprovalStepInstance
  id UUID PRIMARY KEY
  approval_instance_id UUID NOT NULL FK ApprovalInstance(id) RESTRICT
  policy_step_id UUID NOT NULL FK ApprovalPolicyStep(id) RESTRICT
  step_order INTEGER NOT NULL
  purpose TEXT NOT NULL CHECK REVIEW|APPROVAL
  completion TEXT NOT NULL CHECK ANY|ALL
  requires_reauthentication BOOLEAN NOT NULL
  status TEXT NOT NULL CHECK PENDING|ACTIVE|COMPLETED|ABORTED
  activated_at TIMESTAMPTZ NULL
  due_at TIMESTAMPTZ NULL
  completed_at TIMESTAMPTZ NULL

  UNIQUE(approval_instance_id, step_order)
```

Purpose/completion/reauth are frozen snapshots of the policy step semantics relevant to the instance; the instance does not consult floating current policy after start.

## 8.3 ApprovalParticipant

When a step becomes ACTIVE, its actor rule is resolved against current Organization/AuthZ facts and strict SoD, then concrete Users are frozen:

```text
ApprovalParticipant
  id UUID PRIMARY KEY
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  assignment_kind TEXT NOT NULL CHECK ORIGINAL|REASSIGNED
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(step_instance_id, user_id)
```

Participant snapshot is historical workflow assignment, **not Authorization authority**.

At decision time User must still be enabled, still possess `approval.act` under canonical B2 evaluation, be the current participant and satisfy SoD.

Participant resolution occurs at **step activation**, not for all steps at initial submit. This allows prior accepted Users to be excluded from later steps and avoids freezing future-step membership too early. Once a step activates, its denominator does not drift when Group/Role membership later changes.

If an active participant later becomes ineligible/offboarded, the row remains historical assignment but action fails live. `approval.reassign`/cancel handles the blocked workflow; quorum never silently shrinks.

## 8.4 ApprovalDecision

```text
ApprovalDecision
  id UUID PRIMARY KEY
  approval_instance_id UUID NOT NULL FK ApprovalInstance(id) RESTRICT
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  participant_id UUID NOT NULL FK ApprovalParticipant(id) RESTRICT
  actor_user_id UUID NOT NULL FK User(id) RESTRICT
  outcome TEXT NOT NULL CHECK ACCEPT|RETURN_FOR_CHANGES
  comment TEXT NULL
  fresh_auth_satisfied BOOLEAN NOT NULL
  decided_at TIMESTAMPTZ NOT NULL

  UNIQUE(step_instance_id, actor_user_id)
```

Immutable/append-only.

For `requires_reauthentication=true`, the decision transaction must consume valid Authentication-owned one-shot fresh-auth evidence and persist only the durable product fact that the requirement was satisfied; no password, Keycloak token or provider claim payload is copied into Approval.

### ANY

First valid ACCEPT completes the Step; other undecided participants become no-longer-required through Step status, not fabricated decisions.

### ALL

All current concrete participants must ACCEPT.

### RETURN_FOR_CHANGES

Any authorized current participant may return the attempt. The ApprovalInstance becomes terminal `RETURNED`; the same Revision returns to DRAFT; Submission and all feedback/decisions remain immutable. B3 `working_version` is never reset.

## 8.5 Strict SoD

Before ACCEPT:

```text
actor_user_id != Submission submitter/creator under frozen SoD rule
actor_user_id has not ACCEPTED another Step of this ApprovalInstance
```

Returning for changes is not an ACCEPT and does not consume the “one accepted step per user” constraint.

No role, including tenant_owner, bypasses SoD.

## 8.6 Withdraw / cancel / reassign

### Withdraw

Submitter-authorized termination:

```text
ACTIVE instance → WITHDRAWN
Revision SUBMITTED → DRAFT
```

### Cancel

`approval.cancel` overseer operation:

```text
ACTIVE instance → CANCELLED
Revision SUBMITTED → DRAFT
```

It does **not** set Revision `CANCELLED`; that is the separate Controlled Information `document.cancel_revision` operation.

### Reassign

Reassignment is an explicit immutable event/relationship change, never an UPDATE that erases the prior assignment:

```text
ApprovalReassignment
  id UUID PRIMARY KEY
  step_instance_id UUID NOT NULL
  from_participant_id UUID NOT NULL
  to_participant_id UUID NOT NULL
  performed_by_user_id UUID NOT NULL
  reason TEXT NOT NULL
  reassigned_at TIMESTAMPTZ NOT NULL
```

Replacement must be enabled, `approval.act`-authorized, compatible with the step qualification semantics and SoD-valid. No broad delegation platform V1.

## 8.7 due_in_days

On step activation:

```text
due_at = activated_at + due_in_days
```

Overdue is query/projection/notification information only. It never auto-accepts, auto-returns, changes ANY/ALL, bypasses SoD or mutates Revision state.

---

# 9. Review feedback / annotations / suggesting mode

## 9.1 ReviewFeedback

Product-owned detached feedback bound to exact immutable candidate:

```text
ReviewFeedback
  id UUID PRIMARY KEY
  approval_instance_id UUID NOT NULL FK ApprovalInstance(id) RESTRICT
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  submission_id UUID NOT NULL FK RevisionSubmission(id) RESTRICT
  author_user_id UUID NOT NULL FK User(id) RESTRICT
  kind TEXT NOT NULL CHECK COMMENT|SUGGESTION
  body TEXT NOT NULL
  anchor_schema TEXT NULL
  anchor_payload JSONB NULL
  suggestion_schema TEXT NULL
  suggestion_payload JSONB NULL
  created_at TIMESTAMPTZ NOT NULL
```

Immutable/append-only V1.

Rules:

- every feedback row must bind the same Submission as its ApprovalInstance;
- COMMENT may be added on REVIEW or APPROVAL steps;
- SUGGESTION may be added only on REVIEW steps;
- `anchor_payload` and `suggestion_payload` are bounded/versioned provider-neutral contracts; raw editor-library IDs are not business identity;
- a provider may visually present tracked insertions/deletions, comments, highlights and redlines, but persisted business feedback is detached from immutable Submission bytes;
- review feedback is not B3 `EditorialComment`: EditorialComment is mutable-DRAFT collaboration and affects pre-SUBMIT eligibility; ReviewFeedback exists after SUBMIT and explains a human review of that exact attempt.

## 9.2 Return to DRAFT / applying a suggestion

On `RETURN_FOR_CHANGES`:

```text
Submission S7 remains immutable
ReviewFeedback remains bound to S7
Revision returns to DRAFT
```

The authoring journey may offer “apply suggestion”. Successful application is:

```text
read S7 ReviewFeedback
→ provider/adapter proposes concrete mutation
→ ordinary B3 WorkingContent CAS at expected working_version
→ new Artifact/structured state as needed
→ working_version++
```

No feedback item has direct write authority over WorkingContent.

V1 does not require a second feedback-resolution workflow before resubmit. If a real compliance/UX requirement later demands mandatory APPLIED/DECLINED disposition per suggestion, add that as a typed fact rather than overloading comment state.

## 9.3 In-product viewer requirement

B4 semantic requirement supplied to R10-E:

```text
Approval/review participant
→ opens exact Submission inside MetalDocs
→ may switch among supported in-product representations
→ may view PDF Rendition when available
→ never must download and open an external application as the normal supported path
```

Viewer technology is replaceable. It may use browser-native PDF, PDF.js, an authoring-provider read-only surface, or another adapter later; none becomes Document/Submission identity.

Review mode may show detached suggesting/redline overlays. Approval mode is content-read-only but may still add COMMENT/annotation feedback and return-for-changes.

---

# 10. Rendition / representation policy — Controlled Information-owned

## 10.1 DocumentTypeRepresentationPolicy

```text
DocumentTypeRepresentationPolicy
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK SOURCE_ONLY|REQUIRE_RENDITION
  required_format TEXT NULL CHECK closed ContentFormat
```

CHECK:

```text
SOURCE_ONLY       → required_format IS NULL
REQUIRE_RENDITION→ required_format IS NOT NULL
```

Policy answers only what must exist for Release. It does not prohibit auxiliary view renditions.

## 10.2 Rendition

```text
Rendition
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL FK RevisionSubmission(id) RESTRICT
  output_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  output_format TEXT NOT NULL
  generator_name TEXT NOT NULL
  generator_version TEXT NULL
  generator_build TEXT NULL
  generated_at TIMESTAMPTZ NOT NULL
```

Immutable.

Laws:

- exact input identity is Submission, not floating Revision/WorkingContent;
- output Artifact proves exact derived bytes;
- generator/build provenance is retained;
- storage location/provider absent;
- renderer output never changes Submission digest;
- renderer retries may create more than one immutable Rendition attempt; only ReleaseRecord selects an official one when policy requires it.

### PDF viewing without mandatory PDF authority

For `SourceOnly`, a PDF Rendition may be generated for viewer convenience before/during review or after Release. It is auxiliary. No PDF availability is a Release blocker.

For `RequireRendition(PDF)`, Release is blocked until a valid PDF Rendition for the exact Submission is selected. After Release, in-product effective-document viewing uses that selected official rendition where applicable.

This prevents `SourceOnly` from becoming “must download DOCX externally” while also preventing universal PDF from re-entering as a hidden product invariant.

---

# 11. Release

## 11.1 ReleasePlan

```text
ReleasePlan
  submission_id UUID PRIMARY KEY FK RevisionSubmission(id) RESTRICT
  not_before TIMESTAMPTZ NULL
  created_at TIMESTAMPTZ NOT NULL
```

Immutable for the Submission attempt. Absence of `not_before` = effective as soon as all required gates hold.

No generic publication-plan engine or cross-document supersession graph is introduced absent frozen requirement.

## 11.2 Release eligibility

For Submission S:

```text
human_gate(S) =
  NoHumanApproval
  OR terminal ACCEPTED ApprovalInstance for S

representation_gate(S) =
  SourceOnly
  OR exact required-format Rendition selected for S

time_gate(S) =
  ReleasePlan absent/not_before NULL
  OR now >= not_before
```

Release may proceed iff all three hold and S still belongs to the current SUBMITTED open Revision of its Document.

## 11.3 ReleaseRecord

```text
ReleaseRecord
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL UNIQUE FK RevisionSubmission(id) RESTRICT
  official_rendition_id UUID NULL FK Rendition(id) RESTRICT
  prior_effective_revision_id UUID NULL FK DocumentRevision(id) RESTRICT
  released_at TIMESTAMPTZ NOT NULL
```

Immutable.

`official_rendition_id`:

```text
SourceOnly        → NULL
RequireRendition  → exact matching Rendition for same Submission
```

Release itself is attributed to a system principal in B6 Audit causality; it is not falsely attributed as a human publication action.

## 11.4 Winning Release transaction

```text
BEGIN
  lock target Document serialization root
  lock/validate candidate Revision + exact Submission
  prove candidate Revision = SUBMITTED and still current open Revision
  prove human gate
  prove representation gate
  prove time gate

  identify prior EFFECTIVE Revision, if any
  lock it under same Document root

  resolve Distribution live Groups to concrete current User membership
  freeze concrete obligations (see §12)

  insert unique ReleaseRecord(S)
  prior EFFECTIVE → SUPERSEDED
  candidate SUBMITTED → EFFECTIVE
  effective_at = released_at semantic fact supplied by ReleaseRecord

  insert DistributionObligation rows for exact Release denominator

  // B6 composes required Audit evidence
  // R10-D composes post-commit notification/search/timer intents
COMMIT
```

Structural B3 partial uniqueness on EFFECTIVE is the final backstop.

Atomicity claim:

```text
ReleaseRecord
+ Revision winner
+ predecessor supersession
+ Distribution denominator
```

commit together or none do.

Two retries/workers on the same Submission are harmless: unique `ReleaseRecord.submission_id` + Document serialization yields one winner. No second “release generation” key exists.

## 11.5 B3 seam closures

### Template-source race

Create-from-template and Release on a Template Document use the same per-Document serialization root. A derived Document cannot commit provenance to a Submission that ceased to be current EFFECTIVE during creation.

### Periodic-review race

PeriodicReviewRecord and Release on the reviewed Document share the same serialization root. Review cannot commit as current-EFFECTIVE evidence after a new Revision wins Release.

---

# 12. Distribution

## 12.1 Live configuration — Group only where evidenced

Minimum V1 live audience relation:

```text
DistributionAudienceGroup
  document_id UUID NOT NULL FK Document(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT

  PRIMARY KEY(document_id, group_id)
```

Current configuration only. Group FK is live and therefore participates in B2 `RESTRICT` hard-delete law.

No generic polymorphic audience target. Direct User/Area/whole-company audience types are reopen additions only when a real consumer is proven.

## 12.2 Concrete Release snapshot

At winning Release, all configured Groups are resolved to concrete membership and deduplicated by User.

Historical obligation never carries a live Group FK:

```text
DistributionObligation
  id UUID PRIMARY KEY
  release_id UUID NOT NULL FK ReleaseRecord(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(release_id, user_id)
```

This is the historical denominator.

Later:

```text
Group membership add/remove
Group rename/delete
User group reassignment
```

never rewrites existing obligations.

### Membership concurrency

Release must not resolve audience in a post-commit worker. The denominator is frozen in the local Release transaction.

Implementation-spec lock law must coordinate with B2 Group/GroupMembership mutation such that concurrent membership change has a serial order relative to the Release snapshot. Use the existing B2 Group hard-delete/membership coordination seams and deterministic Group/member locking; do not introduce an eventual “best effort audience” projection as authority.

If implementation proof shows the promoted B2 locking contract cannot provide this serial boundary without a cycle, that is a material cross-stage counterexample and narrowly reopens the implicated lock law before coding.

## 12.3 AcknowledgementRecord

```text
AcknowledgementRecord
  id UUID PRIMARY KEY
  obligation_id UUID NOT NULL UNIQUE FK DistributionObligation(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  acknowledged_at TIMESTAMPTZ NOT NULL
```

Immutable.

Laws:

- user_id must equal obligation.user_id;
- caller must be authenticated/eligible and authorized to view the effective Document through normal domain rules;
- explicit acknowledgement action is required;
- notification delivered/read, document viewed, PDF viewed, source downloaded, search hit or email-open never creates acknowledgement;
- no manager/proxy acknowledgement V1;
- no mandatory fresh-auth/e-signature V1.

Distribution reminders/notifications are R10-D effects over these canonical obligation facts; delivery failure never erases obligation.

---

# 13. Permission target classification / B2 coherence

Tenant-wide configuration:

```text
approval_policy.manage
```

Area-targeted execution by owning Document.area_id and exact relationship:

```text
approval.act
approval.oversee
approval.reassign
approval.cancel

distribution.manage
distribution.oversee
```

Whole-company configuration must not be silently downgraded because a referenced Document/Area exists.

Participant/action relationship rules supplement permissions; permission alone never lets a user act on an arbitrary ApprovalInstance.

No direct `artifact.read`, `pdf.view`, `renderer.retry`, `eigenpal.review`, `viewer.annotate` or provider-specific permissions are introduced. Viewing and review feedback are authorized through Document/Approval relationships.

---

# 14. Persistence class × mutation law

| Family | Owner | Mutation law |
|---|---|---|
| DocumentTypeApprovalConfig | Approval | mutable current config |
| ApprovalPolicy | Approval | stable identity/status/display |
| ApprovalPolicyVersion | Approval | immutable append-only |
| ApprovalPolicyStep | Approval | immutable within version |
| ApprovalInstance | Approval | explicit lifecycle only |
| ApprovalStepInstance | Approval | explicit lifecycle only |
| ApprovalParticipant | Approval | append/current assignment semantics; history retained |
| ApprovalDecision | Approval | immutable append-only |
| ApprovalReassignment | Approval | immutable append-only |
| ReviewFeedback | Approval | immutable append-only detached feedback |
| DocumentTypeRepresentationPolicy | Controlled Information | mutable current config |
| Rendition | Controlled Information | immutable derived representation |
| ReleasePlan | Controlled Information | immutable per Submission attempt |
| ReleaseRecord | Controlled Information | immutable effectivity evidence |
| DistributionAudienceGroup | Distribution | mutable current config |
| DistributionObligation | Distribution | immutable Release denominator |
| AcknowledgementRecord | Distribution | immutable one-time explicit acknowledgement |

---

# 15. Structural constraint envelope

```text
ApprovalPolicy.code                                  UNIQUE
ApprovalPolicyVersion(policy_id, version_no)         UNIQUE
ApprovalPolicyStep(policy_version_id, step_order)    UNIQUE

ApprovalInstance.submission_id                       UNIQUE
ApprovalStepInstance(instance_id, step_order)        UNIQUE
ApprovalParticipant(step_id,user_id)                 UNIQUE
ApprovalDecision(step_id,actor_user_id)              UNIQUE

ReleasePlan.submission_id                            PK/UNIQUE
ReleaseRecord.submission_id                          UNIQUE

DistributionAudienceGroup(document_id,group_id)      PK
DistributionObligation(release_id,user_id)           UNIQUE
AcknowledgementRecord.obligation_id                  UNIQUE
```

Cross-row guards must prove:

- ApprovalInstance Submission belongs to expected Document/Revision state;
- ReviewFeedback Submission = instance Submission;
- ReleaseRecord official Rendition, if present, derives from same Submission;
- required representation format matches policy;
- Acknowledgement User = obligation User;
- live Group/Area refs obey B2 lifecycle eligibility.

No cross-owner CASCADE/SET NULL.

---

# 16. Transaction contracts

## 16.1 SUBMIT → Approval initialization

B3 performs exact Submission freeze. B4 composition in the same product command/transaction where appropriate:

```text
read DocumentTypeApprovalConfig
if NoHumanApproval:
  no fake human instance
else:
  lock/read ACTIVE ApprovalPolicy
  pin latest committed PolicyVersion
  insert ApprovalInstance(S)
  insert StepInstances from exact version
  activate Step 1
  resolve concrete participants
  fail closed if active step has no valid participant after SoD
```

No policy version change can affect the running instance.

## 16.2 Review feedback append

```text
BEGIN
  prove instance ACTIVE
  prove step ACTIVE and actor is current participant
  prove live approval.act + relationship
  prove feedback Submission = instance Submission
  if SUGGESTION prove step purpose=REVIEW
  insert immutable ReviewFeedback
COMMIT
```

Does not touch WorkingContent/Artifact/Submission.

## 16.3 ACCEPT decision

```text
BEGIN
  serialize active StepInstance
  prove actor participant + enabled + live approval.act
  prove SoD
  if requires_reauthentication consume one-shot Authentication evidence
  insert immutable ACCEPT decision
  evaluate ANY|ALL

  if step complete:
    complete current Step
    if next step:
      activate next
      resolve concrete participant snapshot excluding SoD-invalid prior acceptors
      fail/hold closed if empty; never relax rule
    else:
      ApprovalInstance → ACCEPTED
COMMIT
```

Release is evaluated separately/idempotently; Approval transaction never lies that content is EFFECTIVE.

## 16.4 RETURN_FOR_CHANGES

```text
BEGIN
  serialize instance/step
  prove actor participant + live authorization
  insert immutable RETURN_FOR_CHANGES decision
  instance → RETURNED
  abort remaining steps
  Revision SUBMITTED → DRAFT
  keep Submission/decisions/feedback immutable
  never reset B3 working_version
COMMIT
```

## 16.5 Release

See §11.4. Distribution denominator is part of the winning local Release transaction.

## 16.6 Acknowledgement

```text
BEGIN
  load obligation + Release/Document relation
  prove caller == obligation User
  prove current authenticated eligible User
  prove effective Document remains view-authorized
  insert AcknowledgementRecord once
COMMIT
```

Duplicate request is idempotent/already-acknowledged; no second record.

---

# 17. Viewer / representation semantics supplied to R10-E

R10-E must build one coherent in-product document workspace with mode-specific capabilities rather than provider-branded product pages:

```text
AUTHOR DRAFT
  source/editor view
  B3 EditorialComments
  WorkingContent mutation via OCC

REVIEW SUBMITTED
  immutable exact Submission
  read-only source view
  optional PDF view of same Submission
  COMMENT + SUGGESTION detached review overlays
  accept / return-for-changes

APPROVAL SUBMITTED
  immutable exact Submission
  read-only source view
  optional PDF view of same Submission
  COMMENT/annotation feedback
  accept / return-for-changes
  optional fresh-auth before decision

EFFECTIVE
  in-product official representation view
  if official_rendition_id exists use that exact rendition
  otherwise source/approved auxiliary viewer through provider-neutral adapter
  acknowledgement action when obligation exists
```

Normal supported viewing must not require “download file → open another desktop application”. Download/export may exist as a separate authorized action, not the viewer implementation.

---

# 18. Adversarial challenge — B3 ↔ B4

## F1 — review suggestions mutate frozen Submission

**Attack:** provider suggesting mode writes tracked changes into submitted bytes.

**Closed:** ReviewFeedback is detached; Submission is immutable. Applying feedback happens only after return to DRAFT through B3 OCC.

## F2 — PDF preview accidentally becomes approval authority

**Attack:** approver views PDF and implementation starts binding decision to PDF hash rather than Submission.

**Closed:** ApprovalDecision binds instance→Submission only. PDF is a Rendition of that same Submission; approval never targets output bytes.

## F3 — SourceOnly secretly becomes mandatory PDF

**Attack:** in-product viewer requirement forces every release to wait for PDF.

**Closed:** representation policy and viewer capability are separate. SourceOnly releases without PDF; auxiliary PDF may be created solely for viewing.

## F4 — reviewer sees PDF from another Submission

**Attack:** floating “latest PDF” is shown during S8 approval while it was rendered from S7/S9.

**Closed:** every viewer Rendition is submission-bound; review/approval UI may show only a rendition whose `submission_id` equals the instance Submission.

## F5 — return-to-DRAFT stale write

**Attack:** pre-SUBMIT writer becomes valid after return.

**Already closed by B3:** SUBMIT consumes N→N+1; B4 return never resets generation.

## F6 — policy changes mid-instance

**Attack:** current policy edit silently alters in-flight steps.

**Closed:** instance pins immutable PolicyVersion.

## F7 — Group membership drift rewrites approval denominator

**Attack:** active review step gains/loses participants as Group changes.

**Closed:** concrete Users frozen when step activates; live AuthZ still gates action.

## F8 — same approver signs two steps

**Attack:** a broad role resolves same User again later.

**Closed:** later participant resolution excludes Users who ACCEPTED earlier Steps; decision path also enforces SoD.

## F9 — renderer failure makes document falsely effective

**Attack:** terminal approval immediately flips EFFECTIVE despite required PDF missing.

**Closed:** Approval complete != Release complete. Required Rendition is independent Release gate; prior Revision remains EFFECTIVE.

## F10 — duplicate release retry creates two effective transitions

**Closed:** per-Document serialization + unique ReleaseRecord(Submission) + B3 one-EFFECTIVE backstop.

## F11 — async Distribution resolves membership too late

**Attack:** Release at 12:00, user leaves Group at 12:01, worker snapshots at 12:05 and historical denominator is false.

**Closed:** concrete obligations are inserted in the winning Release transaction. Notification can be async; obligation cannot.

## F12 — Group delete makes history uninterpretable

**Closed:** live config uses Group FK RESTRICT. Historical approval participants/distribution obligations point to concrete Users, not Group.

## F13 — view/download marks acknowledgement

**Closed:** only explicit AcknowledgementRecord action completes obligation.

## F14 — fresh-auth becomes password duplication

**Closed:** Authentication owns the challenge/provider evidence; Approval records only successful bounded consumption.

## F15 — comment systems duplicate B3 EditorialComment

**Closed:** B3 EditorialComment = DRAFT collaboration/submit gate. B4 ReviewFeedback = immutable feedback on an already-submitted exact attempt. Different owner/time/mutation law.

---

# 19. Essential vs accidental complexity / YAGNI

## Essential

- immutable policy versions;
- ordered specialized steps;
- concrete participant snapshots;
- strict SoD;
- exact-Submission decisions;
- detached review comments/suggestions;
- optional fresh-auth consumption;
- immutable rendition provenance;
- automatic one-winner Release;
- exact official-rendition selection when required;
- concrete Release-time distribution obligations;
- immutable explicit acknowledgement;
- in-product viewer successor contract.

## Remove/defer

- BPMN/generic workflow engine;
- parallel DAG stages;
- M-of-N quorum;
- generic delegation platform;
- auto-escalation workflow;
- publish button/capability;
- approval of PDF/renderer bytes;
- universal mandatory PDF;
- provider/editor-specific business permissions;
- mutable review of submitted bytes;
- generic annotation platform covering every format up front;
- post-commit eventual audience as historical authority;
- read-event analytics as acknowledgement truth;
- LMS/training platform;
- e-signature for acknowledgement without requirement;
- duplicate release-generation identity.

---

# 20. Global Maximum decision — candidate only

Recommended Method outcome:

> **RESTRUCTURE NOW** — converge B4 on a small specialized Approval kernel over exact RevisionSubmission, detached provider-neutral review feedback, immutable derived Renditions, CI-owned automatic Release as the sole effectivity transition, and concrete-user Distribution obligations frozen in that Release transaction. R10-E must expose those semantics through an in-product viewer/review workspace without making current editor/viewer technology part of product identity.

Candidate family set:

```text
Approval
  DocumentTypeApprovalConfig
  ApprovalPolicy
  ApprovalPolicyVersion
  ApprovalPolicyStep
  ApprovalInstance
  ApprovalStepInstance
  ApprovalParticipant
  ApprovalDecision
  ApprovalReassignment
  ReviewFeedback

Controlled Information
  DocumentTypeRepresentationPolicy
  Rendition
  ReleasePlan
  ReleaseRecord

Distribution
  DistributionAudienceGroup
  DistributionObligation
  AcknowledgementRecord
```

On operator acceptance:

```text
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
implementation = BLOCKED
next = R10-B5 integrated candidate
```

No Fable/microreview is requested here. Whole-R10 independent review remains the default gate.

---

# 21. Reopen triggers

Reopen B4 only on material evidence such as:

- real parallel/concurrent approval stages required;
- legitimate M-of-N quorum;
- actor qualification cannot be represented by NamedUser/Group/RoleInArea;
- real delegated authority requires a durable delegation domain;
- a compliance requirement makes Approval target a rendition/signature package rather than Submission;
- official multi-rendition package becomes mandatory;
- explicit effective-date semantics require more than one optional `not_before` gate;
- Distribution proves direct User/Area/company audience types are first-class V1 configuration;
- acknowledgement needs e-signature/fresh-auth/training semantics;
- review suggestions require a richer provider-neutral patch model with a real second format consumer;
- B2 Group/membership lock law cannot serialize release-time audience snapshot without a material deadlock/invariant failure;
- B3 exact-Submission identity proves insufficient for Release correlation;
- viewer/provider cannot expose exact Submission/rendition safely without leaking provider identity into product contracts.

Implementation inconvenience, existing EigenPal/Gotenberg/MinIO shape or old approval tables are not reopen triggers.
