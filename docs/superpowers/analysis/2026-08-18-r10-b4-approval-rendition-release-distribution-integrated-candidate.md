# R10-B4 — Approval + CI-owned Rendition/Release + Distribution — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input:** B1/B2 promoted authority + B3 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This is staging analysis. It does not independently ratify B4 or amend frozen R3–R9.5 authority. B4 remains challengeable by B5–F and Whole-R10 review.

---

# 1. Authority / evidence boundary

Authority path:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. accepted non-final B3 candidate

Current code/schema/OpenAPI/legacy ADRs are evidence only. External products are comparison evidence only.

---

# 2. Operator refinement carried into B4

The operator requires:

- supported governed documents are viewable **inside MetalDocs**; normal viewing must not require download + external desktop application;
- review and approval can visualize the exact candidate as PDF when a PDF rendition is available/generated;
- collaborative review supports comments and a suggesting/redline-like interaction;
- current editor/viewer technology remains an adapter, never product/domain identity.

Therefore:

```text
OfficialRepresentationPolicy
!= viewer capability
```

`SourceOnly` may still have an auxiliary PDF Rendition for in-product viewing. That PDF is not official merely because the viewer uses it.

And:

```text
suggesting UX
!= mutation of RevisionSubmission
```

`SUBMITTED` remains immutable. Review comments/proposed changes are detached, exact-Submission-bound feedback. Applying a suggestion after `return_for_changes` is a normal B3 WorkingContent OCC mutation.

---

# 3. Known / Inferred / Unknown / Deferred

## Known

B4 preserves:

- specialized sequential Approval, not BPM;
- versioned ApprovalPolicy with ordered Steps;
- Step `purpose=review|approval`;
- actor rule `NamedUser|Group|RoleInArea`;
- completion `ANY|ALL`;
- optional reauthentication and `due_in_days`;
- human outcomes `accept|return_for_changes`;
- separate withdraw/cancel/reassign;
- no normal terminal reject V1;
- strict SoD: creator/submitter cannot accept own Submission; same User cannot accept two Steps of one ApprovalInstance; reassignment remains qualified + SoD-valid;
- ApprovalInstance binds one exact immutable RevisionSubmission;
- return/withdraw returns same Revision to DRAFT without mutating Submission;
- fresh-auth is Authentication-owned evidence consumption;
- Rendition derives exact Submission and retains output hash/Artifact + generator/build provenance;
- Approval approves Submission, never renderer bytes;
- Release is automatic/system-owned; no publish button;
- optional `ReleasePlan.not_before`; `effective_at=released_at`;
- winning Release makes candidate EFFECTIVE and predecessor SUPERSEDED atomically;
- representation policy `SourceOnly|RequireRendition(ContentFormat)`;
- Distribution snapshots concrete Users at Release;
- later Group membership never rewrites historical denominator;
- explicit immutable AcknowledgementRecord is the only completion signal;
- B2 Group live-reference RESTRICT law;
- B3 exact Submission identity, OCC generation consumption and one-EFFECTIVE backstop.

## Inferred candidate choices

1. `RevisionSubmission.id` is the shared identity for Approval, Rendition and Release; no duplicate `release_generation` identity.
2. ApprovalPolicy has stable identity + immutable numbered versions.
3. At ApprovalInstance creation, every Step actor rule resolves to a **concrete User pool snapshot**. In-flight execution never depends on later Group/Role membership.
4. At step activation, the active participant denominator is selected from that frozen pool after strict SoD filtering; action-time AuthZ remains live.
5. ReviewFeedback is detached immutable state bound to exact Submission.
6. `SUGGESTION` feedback is REVIEW-only; COMMENT is permitted on REVIEW or APPROVAL.
7. Auxiliary PDF Rendition may exist under `SourceOnly`; only representation policy decides Release blocking.
8. V1 admits at most one semantic Rendition per `(Submission, ContentFormat)`; failed renderer attempts are R10-D mechanism state and never Rendition rows.
9. ReleaseRecord selects the exact official Rendition when one is required.
10. Distribution live audience V1 includes Group because B2 proves a real Group consumer; no generic polymorphic audience engine.
11. Distribution obligations are inserted in the **winning Release transaction**, not resolved later by a worker.
12. Acknowledgement is explicit authenticated action but not fresh-auth/e-signature V1.
13. `due_in_days` is attention/SLA data only; no auto-escalation or lifecycle mutation.
14. `approval.cancel` returns the Approval attempt to DRAFT; cancelling the Revision remains `document.cancel_revision`.

## Unknown

- future direct User/Area/company Distribution audience types;
- mandatory per-feedback APPLIED/DECLINED resolution before resubmit;
- richer review-anchor schemas for future formats;
- post-release re-rendering/replacement of a semantic Rendition;
- future e-signature/fresh-auth for Acknowledgement;
- exact viewer/editor adapter implementation.

## Deferred

```text
Evidence/Dossier/Records/Retention/Hold/Disposition → B5
Audit final skeleton / same-commit matrix           → B6
physical storage/malware/restore                    → R10-C
workers/outbox/timers/notifications/projections     → R10-D
API/frontend/viewer/editor/provider journeys        → R10-E
migration/cutover/deletion                          → R10-F
```

R10-E has a hard successor obligation: supported content must be viewable in-product, including exact PDF viewing where applicable, without making a provider brand part of the domain surface.

---

# 4. Root Cause

The structural failure class is **collapsed governance stages**. Human decision, derived representation, effectivity and distribution obligation can be wrongly collapsed into one floating “document status” or multiple pipelines that each infer “current content”.

B4 must separate these truths while causally binding all of them to one immutable Submission.

---

# 5. Target invariant

> Every Approval decision, every Rendition selected for official use and every ReleaseRecord refers to the same immutable RevisionSubmission. Winning Release is the only effectivity transition; it supersedes the prior EFFECTIVE Revision and freezes concrete Distribution obligations in the same local transaction. Viewer/review interaction may add detached feedback or derived preview representations but can never mutate the submitted candidate.

```text
ApprovalDecision    != content mutation
ReviewFeedback      != WorkingContent
Rendition           != source Submission
PDF preview         != automatically official
Approval complete   != Release complete
Release             != human publish action
Distribution        != Authorization
View/download       != Acknowledgement
Group membership    != historical denominator
```

---

# 6. Alternatives

## A — status-heavy Revision workflow

**Reject / Local Maximum.** Mixes incompatible mutation laws and cannot prove exact-candidate coherence.

## B — generic BPM/workflow platform

**Reject / overengineered.** Creates generic nodes/conditions/delegation machinery without real consumers.

## C — specialized Approval + detached feedback + immutable Rendition + automatic Release + concrete Distribution snapshot

**Recommended Global Maximum.** Minimum structure preserving the frozen invariants and operator viewer/review requirements.

---

# 7. Approval configuration

## 7.1 DocumentTypeApprovalConfig

```text
DocumentTypeApprovalConfig
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_HUMAN_APPROVAL|USE_POLICY
  approval_policy_id UUID NULL FK ApprovalPolicy(id) RESTRICT
```

CHECK:

```text
NO_HUMAN_APPROVAL → policy NULL
USE_POLICY        → policy NOT NULL
```

Tenant-wide `approval_policy.manage`.

## 7.2 ApprovalPolicy

```text
ApprovalPolicy
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  created_at TIMESTAMPTZ NOT NULL
```

Stable identity/current eligibility only.

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

Immutable. Highest committed version of ACTIVE policy is current. “Rollback” creates a new version.

## 7.4 ApprovalPolicyStep — immutable semantic snapshot

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
  group_id_snapshot UUID NULL          // historical identity value, NOT a live FK
  group_name_snapshot TEXT NULL
  role_code TEXT NULL
  area_id UUID NULL FK Area(id) RESTRICT

  UNIQUE(policy_version_id, step_order)
```

Closed-union CHECK prevents mixed actor kinds.

User identity is durable; Area is retired rather than historically deleted. Group is different because B2 permits hard delete after live references disappear. Therefore historical Group rule is a snapshot, not permanent FK.

## 7.5 ApprovalPolicyLiveGroupRef — current configuration only

```text
ApprovalPolicyLiveGroupRef
  policy_id UUID NOT NULL FK ApprovalPolicy(id) RESTRICT
  policy_version_id UUID NOT NULL FK ApprovalPolicyVersion(id) RESTRICT
  policy_step_id UUID NOT NULL FK ApprovalPolicyStep(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT

  PRIMARY KEY(policy_step_id)
```

Exists only for Group steps in the **current live PolicyVersion**. New version promotion atomically replaces the live refs after validating exact snapshot UUID/name coherence. Group hard delete fails while this live config exists.

Crucially, ApprovalInstance initialization resolves every Step to concrete Users, so in-flight instances never require Group survival. Once a version stops being current, its live Group refs can be removed without damaging running/history truth.

New policy version rejects disabled NamedUser and retired Area; Group live ref proves Group exists. Runtime actors still need live `approval.act`.

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

No duplicate document/revision/hash identity.

`NoHumanApproval` creates no fake human instance; it simply satisfies the human Release gate for the existing B3 Submission.

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

## 8.3 ApprovalParticipantPool — frozen actor resolution

At ApprovalInstance creation, **all** Step rules are resolved to concrete Users using current Organization/AuthZ state:

```text
ApprovalParticipantPool
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  resolved_at TIMESTAMPTZ NOT NULL

  PRIMARY KEY(step_instance_id, user_id)
```

This is the immutable pool snapshot for that attempt.

Consequences:

- later Group membership/RoleAssignment change does not add/remove pool members;
- old Group can later be deleted without rewriting instance history;
- policy version and actor pool are deterministic for the attempt;
- live Authorization still gates actual action.

Pool initialization fails closed if a Step resolves to zero concrete Users before SoD filtering.

## 8.4 ApprovalParticipant — active denominator

At each Step activation, materialize the actionable denominator from its frozen pool after applying current strict SoD exclusions:

```text
ApprovalParticipant
  id UUID PRIMARY KEY
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  assignment_kind TEXT NOT NULL CHECK ORIGINAL|REASSIGNED
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(step_instance_id,user_id)
```

A User who ACCEPTED a prior Step is excluded from later active denominator. If filtering leaves zero participants, the Step remains fail-closed/uncompletable and requires qualified reassignment or cancellation; **zero participants never satisfies ANY/ALL**.

Later offboarding/permission loss does not rewrite participant history. It blocks action live and may require reassignment.

Reassignment may select only a User from the frozen Step pool who is currently enabled, currently `approval.act`-authorized and SoD-valid. If no such User exists, the attempt must be cancelled/returned and resubmitted under a suitable policy; V1 does not expand the actor rule mid-instance.

## 8.5 ApprovalDecision

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

  UNIQUE(step_instance_id,actor_user_id)
```

Immutable.

Decision-time gates:

```text
active instance + active step
current participant
User enabled
live canonical approval.act
strict SoD
fresh-auth one-shot consumed when required
```

Approval stores no password/provider token/claim payload. It persists only the durable product fact that required fresh-auth was satisfied for that decision.

### ANY

First valid ACCEPT completes Step. No fabricated decisions for other participants.

### ALL

Every active participant must ACCEPT. Empty denominator is invalid, not vacuous success.

### RETURN_FOR_CHANGES

Any current authorized participant may return the attempt:

```text
instance → RETURNED
remaining steps → ABORTED
Revision SUBMITTED → DRAFT
Submission/feedback/decisions remain immutable
B3 working_version never resets
```

## 8.6 Strict SoD

Before ACCEPT:

```text
actor != Submission creator/submitter under frozen rule
actor has not ACCEPTED another Step in this ApprovalInstance
```

RETURN_FOR_CHANGES is not ACCEPT and does not consume the “one accepted Step” constraint.

No role bypasses SoD.

## 8.7 Withdraw / cancel / reassign

```text
withdraw → instance WITHDRAWN → Revision DRAFT
cancel   → instance CANCELLED → Revision DRAFT
```

Neither sets business Revision `CANCELLED`; that remains Controlled Information `document.cancel_revision`.

Reassignment is immutable evidence:

```text
ApprovalReassignment
  id UUID PRIMARY KEY
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  from_participant_id UUID NULL FK ApprovalParticipant(id) RESTRICT
  to_participant_id UUID NOT NULL FK ApprovalParticipant(id) RESTRICT
  performed_by_user_id UUID NOT NULL FK User(id) RESTRICT
  reason TEXT NOT NULL
  reassigned_at TIMESTAMPTZ NOT NULL
```

`from_participant_id` may be NULL only when a Step activation produced zero actionable participants and an overseer activates a qualified pool User. No generic delegation platform.

## 8.8 due_in_days

```text
due_at = activated_at + due_in_days
```

Overdue is projection/notification information only. No automatic transition/quorum change/escalation.

---

# 9. Review feedback / suggesting / annotations

## 9.1 ReviewFeedback

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

Laws:

- feedback Submission must equal instance Submission;
- COMMENT allowed in REVIEW or APPROVAL;
- SUGGESTION allowed only in REVIEW;
- anchors/proposals use bounded versioned provider-neutral contracts;
- raw editor-library IDs never become business identity;
- provider may visually show tracked insertions/deletions, highlights and comments, but submitted bytes never change;
- B3 EditorialComment remains distinct DRAFT collaboration/submit-gate state.

## 9.2 Applying a suggestion

After RETURN_FOR_CHANGES:

```text
S7 remains immutable
feedback remains on S7
Revision becomes DRAFT
```

“Apply suggestion” is:

```text
read ReviewFeedback
→ adapter maps proposal to concrete candidate mutation
→ B3 WorkingContent CAS(expected_working_version)
→ Artifact/structured state changes as required
→ working_version++
```

Feedback itself has no write authority over WorkingContent.

V1 does not invent a second feedback-resolution workflow. A future requirement for mandatory APPLIED/DECLINED disposition is a reopen trigger.

---

# 10. In-product viewer semantics

B4 supplies the semantic contract; R10-E supplies UI/API/provider implementation.

```text
AUTHOR DRAFT
  editable source/provider view
  WorkingContent OCC
  B3 EditorialComments

REVIEW SUBMITTED
  exact immutable Submission
  read-only source view
  optional PDF view derived from same Submission
  COMMENT + SUGGESTION overlays
  accept / return-for-changes

APPROVAL SUBMITTED
  exact immutable Submission
  read-only source view
  optional PDF view derived from same Submission
  COMMENT/annotation feedback
  accept / return-for-changes
  optional fresh-auth before decision

EFFECTIVE
  in-product official representation view
  exact Release-selected rendition when one exists
  otherwise source/auxiliary viewer through provider-neutral adapter
  explicit acknowledgement when obligation exists
```

Normal supported viewing must not require download + external app. Download/export remains separate optional action.

Viewer/editor brands never enter Document/Revision/Submission identity or semantic permissions.

---

# 11. Rendition / representation policy — Controlled Information

## 11.1 DocumentTypeRepresentationPolicy

```text
DocumentTypeRepresentationPolicy
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK SOURCE_ONLY|REQUIRE_RENDITION
  required_format TEXT NULL CHECK closed ContentFormat
```

```text
SOURCE_ONLY        → format NULL
REQUIRE_RENDITION → format NOT NULL
```

Policy controls Release requirements only, not viewer capability.

## 11.2 Rendition

```text
Rendition
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL FK RevisionSubmission(id) RESTRICT
  output_artifact_id UUID NOT NULL UNIQUE FK Artifact(id) RESTRICT
  output_format TEXT NOT NULL
  generator_name TEXT NOT NULL
  generator_version TEXT NULL
  generator_build TEXT NULL
  generated_at TIMESTAMPTZ NOT NULL

  UNIQUE(submission_id, output_format)
```

Immutable semantic success record.

Failed/retried renderer attempts live in R10-D job/effect state and do not create Rendition. First confirmed semantic success for `(Submission,Format)` wins V1. Replacing a confirmed Rendition is not supported silently; a real re-render correction requirement reopens this bounded rule.

Laws:

- exact input = Submission;
- exact output = Artifact;
- generator/build provenance retained;
- provider/storage location absent;
- renderer output never changes Submission digest.

### PDF viewer without hidden mandatory PDF

`SourceOnly` may generate PDF for viewing. Release does not wait for it.

`RequireRendition(PDF)` blocks Release until the exact Submission has its semantic PDF Rendition. ReleaseRecord then pins that Rendition as official.

Review/approval PDF tab may only display a Rendition whose `submission_id` equals the ApprovalInstance Submission.

---

# 12. Release

## 12.1 ReleasePlan

```text
ReleasePlan
  submission_id UUID PRIMARY KEY FK RevisionSubmission(id) RESTRICT
  not_before TIMESTAMPTZ NULL
  created_at TIMESTAMPTZ NOT NULL
```

Immutable per Submission. No generic publication-plan engine.

## 12.2 Gates

```text
human_gate = NoHumanApproval OR ACCEPTED ApprovalInstance for same Submission
representation_gate = SourceOnly OR exact required-format Rendition exists
time_gate = no not_before OR now >= not_before
```

Candidate must still be the current SUBMITTED open Revision of its Document.

## 12.3 ReleaseRecord

```text
ReleaseRecord
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL UNIQUE FK RevisionSubmission(id) RESTRICT
  official_rendition_id UUID NULL FK Rendition(id) RESTRICT
  prior_effective_revision_id UUID NULL FK DocumentRevision(id) RESTRICT
  released_at TIMESTAMPTZ NOT NULL
```

Immutable.

Cross-row guard:

```text
SourceOnly        → official_rendition_id NULL
RequireRendition → rendition required, correct format, same Submission
```

`effective_at` semantics come from `released_at`; no second mutable effective timestamp authority is necessary.

## 12.4 Winning Release transaction

```text
BEGIN
  lock Document serialization root
  lock/validate candidate Revision + exact Submission
  prove Revision = SUBMITTED/current open
  prove human/representation/time gates
  identify prior EFFECTIVE Revision

  load DistributionAudienceGroup rows
  lock referenced Group rows FOR UPDATE in UUID order
    // conflicts with concurrent FK membership insert / Group deletion
  lock existing GroupMembership rows for those Groups FOR UPDATE in deterministic order
    // conflicts with membership removal/offboarding deletion
  resolve deduplicated concrete Users visible in this serialized membership cut

  insert unique ReleaseRecord(Submission)
  prior EFFECTIVE → SUPERSEDED
  candidate SUBMITTED → EFFECTIVE
  insert DistributionObligation rows for resolved Users

  // B6: required Audit composition
  // R10-D: post-commit notification/search/timer intents
COMMIT
```

This uses B2’s existing FK/group-deletion semantics rather than a new membership-version subsystem. Concurrent membership change is serialized before or after the Release snapshot; audience truth is never “best effort after commit”.

No User row lock is acquired after Group/member locks. Offboarding that already owns a User lock and tries to delete a membership serializes on the membership row; Release may therefore logically precede or follow offboarding without introducing a Group→User reverse lock cycle.

Two release retries are harmless: Document serialization + `UNIQUE(ReleaseRecord.submission_id)` + B3 one-EFFECTIVE partial uniqueness yields one winner.

## 12.5 B3 seam closure

Create-from-template and Release on the template Document use the same Document serialization root, preventing stale template origin.

PeriodicReviewRecord and Release on the reviewed Document share the same root, preventing stale review of a Revision that ceased to be current EFFECTIVE.

---

# 13. Distribution

## 13.1 Live audience configuration

Minimum evidenced V1:

```text
DistributionAudienceGroup
  document_id UUID NOT NULL FK Document(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT

  PRIMARY KEY(document_id,group_id)
```

Current configuration only; Group FK participates in B2 hard-delete RESTRICT.

No generic `target_type/target_id`. Direct User/Area/company audience types require real product evidence.

## 13.2 DistributionObligation

Created in winning Release transaction:

```text
DistributionObligation
  id UUID PRIMARY KEY
  release_id UUID NOT NULL FK ReleaseRecord(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(release_id,user_id)
```

Concrete historical denominator. No live Group FK.

Group rename/delete/membership change never rewrites obligations.

## 13.3 AcknowledgementRecord

```text
AcknowledgementRecord
  id UUID PRIMARY KEY
  obligation_id UUID NOT NULL UNIQUE FK DistributionObligation(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  acknowledged_at TIMESTAMPTZ NOT NULL
```

Immutable.

- User must equal obligation User;
- explicit authenticated action only;
- notification delivery/read, view, PDF view, source download, search hit do not acknowledge;
- no proxy acknowledgement;
- no mandatory fresh-auth/e-signature V1.

Reminder/delivery jobs are R10-D effects. They may fail without changing obligation truth.

---

# 14. AuthZ / B2 coherence

Tenant-wide configuration:

```text
approval_policy.manage
```

Area-targeted execution through Document.area_id + exact domain relationship:

```text
approval.act
approval.oversee
approval.reassign
approval.cancel
distribution.manage
distribution.oversee
```

Permission alone never grants arbitrary case access. Participant/ownership/state/SoD predicates still apply.

No provider/mechanism permissions:

```text
artifact.read
pdf.view
renderer.retry
editor.suggest
viewer.annotate
```

Viewing/feedback are authorized through Document/Approval relationships.

---

# 15. Persistence class × mutation law

| Family | Owner | Mutation law |
|---|---|---|
| DocumentTypeApprovalConfig | Approval | mutable current config |
| ApprovalPolicy | Approval | stable identity/status/display |
| ApprovalPolicyVersion | Approval | immutable |
| ApprovalPolicyStep | Approval | immutable snapshot |
| ApprovalPolicyLiveGroupRef | Approval | current live config only |
| ApprovalInstance | Approval | explicit lifecycle |
| ApprovalStepInstance | Approval | explicit lifecycle |
| ApprovalParticipantPool | Approval | immutable pool snapshot |
| ApprovalParticipant | Approval | active assignment history |
| ApprovalDecision | Approval | immutable |
| ApprovalReassignment | Approval | immutable |
| ReviewFeedback | Approval | immutable detached feedback |
| DocumentTypeRepresentationPolicy | CI | mutable current config |
| Rendition | CI | immutable semantic success |
| ReleasePlan | CI | immutable per Submission |
| ReleaseRecord | CI | immutable effectivity evidence |
| DistributionAudienceGroup | Distribution | mutable current config |
| DistributionObligation | Distribution | immutable denominator |
| AcknowledgementRecord | Distribution | immutable acknowledgement |

---

# 16. Structural constraints

```text
ApprovalPolicy.code                                  UNIQUE
ApprovalPolicyVersion(policy_id,version_no)          UNIQUE
ApprovalPolicyStep(version_id,step_order)            UNIQUE
ApprovalPolicyLiveGroupRef.policy_step_id            PK
ApprovalInstance.submission_id                       UNIQUE
ApprovalStepInstance(instance_id,step_order)         UNIQUE
ApprovalParticipantPool(step_id,user_id)             PK
ApprovalParticipant(step_id,user_id)                 UNIQUE
ApprovalDecision(step_id,actor_user_id)              UNIQUE
Rendition(submission_id,output_format)                UNIQUE
ReleasePlan.submission_id                            PK
ReleaseRecord.submission_id                          UNIQUE
DistributionAudienceGroup(document_id,group_id)      PK
DistributionObligation(release_id,user_id)           UNIQUE
AcknowledgementRecord.obligation_id                  UNIQUE
```

Cross-row DB/application guards must prove:

- actor union validity;
- live Group ref matches current Group snapshot in step;
- instance Submission/state coherence;
- feedback Submission = instance Submission;
- official Rendition = same Submission + required format;
- acknowledgement User = obligation User;
- retired Area/disabled NamedUser/new Group live refs rejected where applicable.

No cross-owner CASCADE/SET NULL.

---

# 17. Core transaction contracts

## 17.1 SUBMIT + Approval initialization

B3 freezes Submission. B4 composition:

```text
read DocumentTypeApprovalConfig
if NO_HUMAN_APPROVAL:
  no fake ApprovalInstance
else:
  lock ACTIVE ApprovalPolicy/current version
  validate live actor refs
  insert ApprovalInstance + all StepInstances
  resolve every Step actor rule to concrete User pool snapshot
  fail if any pool is empty
  activate Step 1 from pool after SoD filter
```

Group resolution during initialization must take a deterministic membership cut (Group row + membership rows) so its pool snapshot is real at transaction time; exact lock order must be composed with B3 Document locks in the implementation spec without reverse acquisition.

## 17.2 ReviewFeedback append

```text
prove ACTIVE instance/step/current participant/live approval.act
prove exact Submission
if SUGGESTION prove REVIEW purpose
insert immutable feedback
```

No WorkingContent mutation.

## 17.3 ACCEPT

```text
serialize active Step
prove participant + live AuthZ + SoD
consume fresh-auth if required
insert immutable ACCEPT
apply ANY|ALL
if complete activate next Step from its frozen pool minus SoD-ineligible prior acceptors
if no next Step → instance ACCEPTED
```

Approval success never directly asserts EFFECTIVE.

## 17.4 RETURN_FOR_CHANGES

```text
insert immutable decision
instance RETURNED
remaining steps ABORTED
Revision SUBMITTED → DRAFT
never reset working_version
```

## 17.5 Release

See §12.4. Effectivity + predecessor supersession + Distribution denominator are one local commit.

## 17.6 Acknowledgement

```text
load obligation/release/document
prove caller == obligation User
prove authenticated eligible user + effective-view authorization
insert once
```

Duplicate request is idempotent/already-acknowledged.

---

# 18. B3↔B4 adversarial challenge

## F1 — suggesting mode mutates Submission

**Closed:** detached ReviewFeedback; applying feedback only in DRAFT via B3 OCC.

## F2 — PDF becomes approval authority

**Closed:** decisions bind Submission; Rendition only derives it.

## F3 — in-product viewer reintroduces universal mandatory PDF

**Closed:** viewer capability and official representation policy are separate.

## F4 — wrong PDF shown during approval

**Closed:** UI may display only Rendition whose Submission matches the instance exactly.

## F5 — stale pre-SUBMIT writer after return

**Closed by B3:** SUBMIT consumes working_version N→N+1; return never resets.

## F6 — policy edit changes in-flight instance

**Closed:** exact immutable PolicyVersion + concrete pools pinned.

## F7 — Group membership drift changes approval pool

**Closed:** pools snapshot at instance start.

## F8 — Group historical FK makes hard-delete impossible

**Closed:** only current live policy config has Group FK; immutable versions/pools survive without Group row.

## F9 — same user accepts two steps

**Closed:** next active denominator excludes prior acceptors; decision path rechecks SoD.

## F10 — zero pool/denominator vacuously completes ALL

**Closed:** empty pool fails initialization; zero active denominator is explicitly uncompletable.

## F11 — renderer failure makes Revision falsely EFFECTIVE

**Closed:** Approval complete != Release complete; required Rendition is separate gate.

## F12 — duplicate renderer jobs create competing semantic PDFs

**Closed V1:** one semantic Rendition per `(Submission,Format)`; job attempts remain R10-D state.

## F13 — duplicate release

**Closed:** Document serialization + unique ReleaseRecord + B3 EFFECTIVE unique backstop.

## F14 — async Distribution snapshots membership too late

**Closed:** obligations inserted in winning Release transaction.

## F15 — membership insert/delete races Release snapshot

**Closed candidate law:** lock Group parent rows FOR UPDATE + current membership rows FOR UPDATE; FK insertion conflicts at Group parent, deletion/offboarding conflicts at membership row. No post-lock User acquisition.

## F16 — Group delete destroys history

**Closed:** current config uses RESTRICT; history is concrete Users/snapshots.

## F17 — view/download counts as acknowledgement

**Closed:** only explicit AcknowledgementRecord.

## F18 — fresh-auth duplicates passwords/provider state

**Closed:** Authentication owns evidence/challenge; Approval stores only satisfied decision fact.

## F19 — B3 EditorialComment and B4 feedback become duplicate authority

**Closed:** DRAFT collaboration vs immutable feedback on submitted attempt; different lifecycle and gate.

---

# 19. Proof obligations

Later implementation proof must show at minimum:

1. policy version edit never mutates in-flight instance;
2. Group hard delete fails only while live current config ref exists, not due historical version/pool;
3. pool snapshot cannot drift after instance start;
4. live AuthZ/offboarding still blocks participant action;
5. strict SoD under concurrent accepts;
6. ANY/ALL cannot complete from empty active denominator;
7. return-to-DRAFT preserves exact old Submission and B3 OCC generation advance;
8. feedback cannot mutate Submission/WorkingContent while SUBMITTED;
9. PDF shown in review is exact-Submission-derived;
10. SourceOnly releases without PDF;
11. RequireRendition blocks until exact format Rendition exists;
12. only one semantic Rendition per Submission/format;
13. Release has one winner and one EFFECTIVE Revision;
14. predecessor supersession + effectivity + obligations are atomic;
15. concurrent GroupMembership add/remove has deterministic serial order versus Release snapshot;
16. later membership changes do not alter obligations;
17. acknowledgement cannot be produced by view/download/notification;
18. in-product viewer path exists for supported effective/review content without requiring external application;
19. provider/editor/viewer identifiers never become business identity or semantic permission.

---

# 20. Essential vs accidental complexity / YAGNI

## Essential

- immutable policy versions;
- live Group ref + historical User pool separation;
- sequential Steps/ANY/ALL;
- strict SoD;
- exact Submission decisions;
- detached comments/suggestions;
- bounded fresh-auth consumption;
- immutable semantic Rendition;
- automatic one-winner Release;
- exact official representation selection;
- Release-time concrete obligations;
- immutable explicit acknowledgement;
- in-product viewer successor contract.

## Remove/defer

- BPMN/generic workflow engine;
- parallel DAG stages;
- M-of-N;
- generic delegation;
- auto-escalation workflow;
- publish button;
- approval of renderer bytes;
- universal mandatory PDF;
- editor/viewer-specific business permissions;
- mutable review of submitted bytes;
- generic annotation platform for every format;
- eventual post-release audience as authority;
- read analytics as acknowledgement;
- LMS/training platform;
- e-signature acknowledgement without requirement;
- duplicate release-generation identity.

---

# 21. Candidate decision

> **RESTRUCTURE NOW** — B4 should be a small specialized Approval kernel over exact RevisionSubmission, with frozen concrete actor pools, detached provider-neutral review feedback, immutable derived Rendition, CI-owned automatic Release as sole effectivity authority, and concrete Distribution obligations frozen in the winning Release transaction. R10-E must expose this through an in-product viewer/review workspace without making the current editor/viewer provider part of product identity.

Candidate families:

```text
Approval
  DocumentTypeApprovalConfig
  ApprovalPolicy
  ApprovalPolicyVersion
  ApprovalPolicyStep
  ApprovalPolicyLiveGroupRef
  ApprovalInstance
  ApprovalStepInstance
  ApprovalParticipantPool
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

No Fable/microreview now. Whole-R10 independent cold review remains the default gate.

---

# 22. Reopen triggers

- real parallel approval stages;
- legitimate M-of-N;
- actor qualification beyond NamedUser/Group/RoleInArea;
- real delegation domain;
- compliance requires Approval to target a rendition/signature package rather than Submission;
- official multi-rendition package;
- more complex effectivity plan than optional `not_before`;
- proven direct User/Area/company Distribution config;
- acknowledgement e-signature/fresh-auth/training semantics;
- real second content format requires richer provider-neutral suggestion patch model;
- Group/membership lock proof reveals a wait-for cycle or false snapshot under B1/B2 laws;
- B3 exact Submission proves insufficient as shared identity;
- viewer/provider cannot serve exact Submission/rendition without leaking provider identity.

Implementation inconvenience or current EigenPal/Gotenberg/MinIO/legacy schema shape is not a reopen trigger.
