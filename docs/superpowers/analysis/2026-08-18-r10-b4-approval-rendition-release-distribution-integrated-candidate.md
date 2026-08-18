# R10-B4 — Approval + CI-owned Rendition/Release + Distribution — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B4 or amend the frozen R3–R9.5 ledger. It consumes the operator-accepted non-final B3 candidate and remains challengeable by B5–F and Whole-R10 review.

> **Bounded reopen under evaluation:** the frozen `ApprovalPolicy Step purpose = review | approval` distinction is challenged by new operator review + external evidence + Method analysis. This candidate replaces it with one governance-step semantic model. No other R9.5 Approval semantic is reopened by this file.

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

Current code/schema/OpenAPI/legacy ADRs remain current-state evidence only.

External products are comparison evidence only, never MetalDocs authority.

Primary external references used for the bounded step-model challenge:

- Veeva Vault document workflow task configuration: <https://quality.veevavault.help/en/gr/21922/>
- Veeva Vault modern document workflow example: <https://quality.veevavault.help/en/lr/50493/>
- Veeva Vault workflow verdict configuration: <https://quality.veevavault.help/en/gr/50498/>
- MasterControl document workflow configuration: <https://currentcloud.onlinehelp.mastercontrol.com/2024.1/en_us/Content/Documents/Manage_a_Document_Workflow.htm>
- MasterControl periodic review policy: <https://currentcloud.onlinehelp.mastercontrol.com/2024.1/en_us/Content/InfoCards/Manage_a_Review_Policy.htm>
- M-Files assignments: <https://userguide.m-files.com/user-guide/latest/eng/assignments.html>
- M-Files annotations/redlining: <https://userguide.m-files.com/user-guide/latest/eng/annotations_and_redlining.html>

Evidence conclusion:

- mature systems do distinguish collaborative review, formal approval, annotations and signatures **at the process/task capability level**;
- they do **not** establish one universal semantic taxonomy that MetalDocs should copy;
- Veeva configures verdicts, comments and eSignature as orthogonal task options;
- MasterControl has its own `Approval | Collaboration | Review` taxonomy, where collaboration owns redlining and `Review` also serves post-approval/periodic-review semantics;
- M-Files distinguishes ordinary assignment and approval assignment while annotations/redlining are a detachable document layer;
- therefore the existence of “review” and “approval” concepts in mature products does not prove that `purpose=REVIEW|APPROVAL` belongs in the MetalDocs kernel.

---

# 2. Operator refinement carried into B4

The operator requires:

1. supported governed documents are viewable **inside MetalDocs**; the normal experience must not require download + external desktop application;
2. approval-route participants may view the exact candidate using the source viewer and/or a PDF rendition generated from that exact Submission;
3. all route participants may comment, annotate and propose corrections/suggestions;
4. provider-specific editor concepts such as a current “suggesting mode” remain adapter behavior, never product/domain identity;
5. **the route must not expose `review` and `approval` as structurally different step types.**

Consequent architecture:

```text
one Approval/Governance Step semantic
+
viewer / feedback / suggestion capability
+
optional fresh-auth
+
ANY | ALL completion
+
actor rule
```

not:

```text
ReviewStep
ApprovalStep
```

Human-facing step names may still be:

```text
Revisão técnica
Qualidade
Diretoria
Validação regulatória
```

but those labels carry no hidden workflow behavior.

---

# 3. Method — Evidence → Known / Inferred / Unknown / Deferred

## 3.1 Known — preserved frozen/product inputs

B4 preserves:

- Approval is a specialized sequential human governance workflow, not generic BPM;
- ApprovalPolicy is versioned and contains ordered Steps;
- actor rule = `NamedUser | Group | RoleInArea`;
- completion = `ANY | ALL`;
- optional `requires_reauthentication`;
- optional `due_in_days`;
- human outcomes = `accept | return_for_changes`;
- `withdraw | cancel | reassign` are separate operations;
- no normal terminal reject V1;
- strict SoD: creator/submitter cannot ACCEPT own Submission;
- same User cannot ACCEPT two Steps of one ApprovalInstance;
- reassignment remains qualified + SoD-valid;
- no role, including `tenant_owner`, is a bypass;
- ApprovalInstance binds one exact immutable RevisionSubmission;
- return/withdraw returns the same Revision to DRAFT without mutating old Submission;
- resubmission creates a new Submission and, where human approval applies, a new ApprovalInstance;
- fresh-auth consumes Authentication-owned evidence; Approval never stores/challenges passwords;
- Rendition derives one exact Submission and retains output Artifact/hash + generator/build provenance;
- Approval decides on Submission, never renderer bytes;
- Release is automatic/system-owned; no publish button;
- optional `ReleasePlan.not_before`; actual effectivity is the winning Release instant;
- winning Release makes candidate EFFECTIVE and prior EFFECTIVE Revision SUPERSEDED atomically;
- representation policy = `SourceOnly | RequireRendition(ContentFormat)`;
- Distribution is controlled obligation/acknowledgement, not AuthZ/LMS;
- Release snapshots concrete Users; later Group membership never rewrites the historical denominator;
- only explicit immutable AcknowledgementRecord completes an obligation;
- B2 Group live-reference RESTRICT law remains;
- B3 exact Submission identity, OCC generation consumption, immutable Submission and one-EFFECTIVE backstop remain integration inputs.

## 3.2 Bounded reopen — old Step purpose

Previously frozen:

```text
Step.purpose = REVIEW | APPROVAL
```

This candidate classifies that distinction as **accidental complexity / legacy-influenced product taxonomy** rather than a necessary invariant.

Why it reopens:

1. route participants ultimately perform the same governance outcome pair: `ACCEPT | RETURN_FOR_CHANGES`;
2. comments/suggestions are useful to any participant, including the final approver;
3. fresh-auth is already an orthogonal Step property and does not require an `APPROVAL` type;
4. exact Submission immutability is the same for every route participant;
5. SoD is the same for every ACCEPT;
6. a `REVIEW` step name collides conceptually with the separate product feature `PeriodicReview`;
7. external products use different taxonomies, demonstrating there is no universal domain type to import;
8. current legacy implementation previously used `review` vs `approval` to drive editor-mode differences — precisely the provider/UI coupling the redesign is removing.

Corrected semantic:

> **Every route Step is one ordered governance gate over the same immutable Submission. The Step says who must act, how many must accept, whether fresh-auth is required, and when it is due. Collaboration capabilities are available independently of Step identity.**

## 3.3 Inferred candidate choices

1. `RevisionSubmission.id` is the shared exact-candidate identity across Approval, Rendition and Release; no duplicate `release_generation` identity.
2. ApprovalPolicy has stable identity + immutable numbered versions.
3. ApprovalPolicyStep has no `purpose` enum.
4. Step `label` is presentation/business language only, not a behavior discriminator.
5. All active route participants may view, comment, annotate and suggest corrections on the exact Submission.
6. `requires_reauthentication` is orthogonal and one-shot per ACCEPT where required.
7. At ApprovalInstance creation, every Step actor expression resolves to a concrete User-pool snapshot so in-flight execution does not depend on later Group/Role membership.
8. At Step activation, the actionable denominator is selected from that frozen pool after SoD filtering; action-time Authorization remains live.
9. ReviewFeedback is detached immutable state bound to exact Submission and never mutates Submission bytes.
10. Any participant who considers a suggested change mandatory uses `RETURN_FOR_CHANGES`; an ACCEPT never applies a suggestion.
11. Auxiliary PDF Rendition may exist under `SourceOnly`; only representation policy determines Release blocking.
12. V1 admits at most one semantic Rendition per `(Submission, ContentFormat)`; failed renderer attempts remain R10-D mechanism state.
13. ReleaseRecord pins the exact official Rendition only when official policy requires one.
14. Distribution live audience V1 includes Group because B2 proves a real Group consumer; no generic polymorphic audience engine.
15. Distribution obligations are inserted in the winning Release transaction, not resolved later by a worker.
16. Acknowledgement is explicit authenticated action but not fresh-auth/e-signature V1 absent a new requirement.
17. `due_in_days` is attention/SLA data only; no auto-escalation or automatic lifecycle mutation.
18. `approval.cancel` terminates the approval attempt and returns the Revision to DRAFT; cancelling the business Revision remains `document.cancel_revision`.

## 3.4 Unknown — bounded

- future direct User/Area/company Distribution audience types;
- mandatory per-feedback APPLIED/DECLINED disposition before resubmit;
- richer review-anchor schemas for future content formats;
- post-release re-rendering/replacement rules for a semantic Rendition;
- future e-signature/fresh-auth requirement for Acknowledgement;
- exact viewer/editor adapter implementation;
- whether a future customer genuinely needs a non-gating “comment-only task” inside the approval route.

The final item is a reopen trigger, not launch machinery.

## 3.5 Deferred

```text
Evidence/Dossier/Records + retention/hold/disposition        → B5
Audit skeleton/final cross-owner same-commit matrix          → B6
physical Artifact storage/malware/restore                    → R10-C
worker/outbox/timers/notification delivery/projections       → R10-D
API/frontend/viewer/editor/provider adapter journeys         → R10-E
historical migration/cutover/legacy deletion                 → R10-F
```

R10-E hard successor requirement:

> supported governed content is viewable in-product; the approval route uses one consistent governance workspace rather than branching into separate Review vs Approval application modes.

---

# 4. Root Cause

B4 has two distinct root-cause classes.

## 4.1 Collapsed governance truth

Human decision, derived representation, effectivity and distribution obligation can be incorrectly represented as one floating document status or through separate paths that each infer “current content”.

That permits:

```text
Approval candidate != rendition candidate != released candidate != distribution denominator
```

## 4.2 Accidental route taxonomy

The legacy-inspired `REVIEW | APPROVAL` Step type embeds interaction/ceremony decisions into Step identity:

```text
REVIEW   → suggesting/editor behavior
APPROVAL → read-only/signature behavior
```

That is a structural inversion.

The target domain fact is not “which UI mode is this Step?” but:

```text
who is responsible for this gate?
how many must accept?
is fresh-auth required?
what exact Submission are they judging?
what feedback did they leave?
what was their decision?
```

The UI/provider mode must derive from capabilities and state, not own the business type.

---

# 5. Target invariant

> **Every Approval route Step is one ordered governance gate over one exact immutable RevisionSubmission. Every active participant may inspect and provide detached feedback on that Submission. A Step is satisfied only by its configured ANY/ALL ACCEPT decisions under live Authorization, SoD and optional fresh-auth. Only completion of all Steps makes the ApprovalInstance ACCEPTED. Rendition and Release continue to bind the same Submission, and winning Release atomically establishes effectivity and the concrete Distribution denominator.**

Corollaries:

```text
Step label           != Step type
feedback capability  != content mutation
fresh-auth           != separate ApprovalStep class
ACCEPT on Step 1     != final document approval
ApprovalInstance ACCEPTED = all configured governance gates satisfied
ReviewFeedback       != WorkingContent
Rendition            != source Submission
PDF preview          != automatically official representation
Approval complete    != Release complete
Release              != human publish action
View/download        != Acknowledgement
Group membership     != historical denominator
PeriodicReview       != Approval route “review step”
```

---

# 6. Credible alternatives

## A — keep `purpose=REVIEW|APPROVAL`

Pros:

- easy mapping to current legacy UI;
- simple mental model when one team “reviews” and another “approves”.

Cons:

- couples editor mode/feedback rules to domain Step type;
- duplicates capability facts already expressible independently;
- ambiguous beside PeriodicReview;
- final approvers may still need comments/suggestions;
- external products disagree on what “review” means;
- forces future providers to implement our legacy mode taxonomy.

**Decision:** reject as Local Maximum.

## B — generic task capability engine

Model each Step with arbitrary pluggable verbs/capabilities such as:

```text
can_comment
can_suggest
can_edit
can_sign
can_choose_verdict
can_set_fields
...
```

This resembles the configurability of broad workflow platforms.

**Decision:** reject as overengineered non-maximum. MetalDocs has no consumer for a low-code workflow task platform.

## C — one governance Step + fixed collaboration surface + small orthogonal controls

```text
ApprovalPolicyStep
  label
  order
  actor_rule
  completion ANY|ALL
  requires_reauthentication
  due_in_days?
```

Universal route interaction:

```text
view exact Submission
view exact PDF rendition when available
comment / annotate / suggest
ACCEPT | RETURN_FOR_CHANGES
```

**Decision:** recommended Global Maximum.

It keeps only capabilities that every real route participant can use, while preserving the one meaningful variable ceremony requirement (`requires_reauthentication`).

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

Immutable. “Rollback” creates a new version.

## 7.4 ApprovalPolicyStep — one semantic type

```text
ApprovalPolicyStep
  id UUID PRIMARY KEY
  policy_version_id UUID NOT NULL FK ApprovalPolicyVersion(id) RESTRICT
  step_order INTEGER NOT NULL CHECK step_order >= 1
  label TEXT NOT NULL
  completion TEXT NOT NULL CHECK ANY|ALL
  requires_reauthentication BOOLEAN NOT NULL
  due_in_days INTEGER NULL CHECK due_in_days > 0

  actor_kind TEXT NOT NULL CHECK NAMED_USER|GROUP|ROLE_IN_AREA
  named_user_id UUID NULL FK User(id) RESTRICT
  group_id_snapshot UUID NULL
  group_name_snapshot TEXT NULL
  role_code TEXT NULL
  area_id UUID NULL FK Area(id) RESTRICT

  UNIQUE(policy_version_id, step_order)
```

There is intentionally **no**:

```text
purpose
stage_kind
review_mode
approval_mode
```

Closed actor-union CHECK prevents mixed actor kinds.

`label` is human language only. For example:

```text
1. Revisão técnica
2. Qualidade
3. Diretor responsável
```

All three Steps have identical kernel semantics.

User identity is durable; Area is retired rather than historically deleted. Group is different because B2 permits hard delete once live references disappear. Therefore immutable historical Group actor expression is snapshot data, not permanent Group FK.

## 7.5 ApprovalPolicyLiveGroupRef — live configuration only

```text
ApprovalPolicyLiveGroupRef
  policy_step_id UUID PRIMARY KEY FK ApprovalPolicyStep(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT
```

Exists only for GROUP actor steps in the **current live PolicyVersion**.

Policy-version promotion atomically:

1. validates live Group existence;
2. snapshots Group UUID/name into immutable Step semantics;
3. establishes live Group refs for the newly current version;
4. removes old live refs only when they are no longer current configuration dependencies.

In-flight instances do not depend on Group survival because actor pools are resolved to concrete Users at instance initialization.

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

No duplicated document/revision/hash identity.

`NoHumanApproval` creates no fake human instance. The human Release gate is satisfied by configuration over the existing B3 Submission.

## 8.2 ApprovalStepInstance

```text
ApprovalStepInstance
  id UUID PRIMARY KEY
  approval_instance_id UUID NOT NULL FK ApprovalInstance(id) RESTRICT
  policy_step_id UUID NOT NULL FK ApprovalPolicyStep(id) RESTRICT
  step_order INTEGER NOT NULL
  label TEXT NOT NULL
  completion TEXT NOT NULL CHECK ANY|ALL
  requires_reauthentication BOOLEAN NOT NULL
  status TEXT NOT NULL CHECK PENDING|ACTIVE|COMPLETED|ABORTED
  activated_at TIMESTAMPTZ NULL
  due_at TIMESTAMPTZ NULL
  completed_at TIMESTAMPTZ NULL

  UNIQUE(approval_instance_id, step_order)
```

No purpose discriminator is copied because none exists.

## 8.3 ApprovalParticipantPool — frozen actor resolution

At ApprovalInstance creation, every Step actor rule resolves to concrete Users using the then-current Organization/AuthZ state:

```text
ApprovalParticipantPool
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  resolved_at TIMESTAMPTZ NOT NULL

  PRIMARY KEY(step_instance_id,user_id)
```

Consequences:

- later Group membership/RoleAssignment change does not add/remove pool members;
- Group may later be deleted after live config references disappear;
- policy version + participant universe remain deterministic for that attempt;
- live Authorization still gates actual action.

Pool initialization fails closed if any configured Step resolves to zero concrete Users before SoD filtering.

## 8.4 ApprovalParticipant — active denominator

At each Step activation, materialize the actionable denominator from its frozen pool after strict SoD exclusions:

```text
ApprovalParticipant
  id UUID PRIMARY KEY
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  assignment_kind TEXT NOT NULL CHECK ORIGINAL|REASSIGNED
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(step_instance_id,user_id)
```

A User who ACCEPTED a prior Step is excluded from later active denominator.

If filtering leaves zero participants:

```text
Step does NOT complete
ANY/ALL does NOT vacuously succeed
```

The attempt requires qualified reassignment from the frozen pool or cancellation/return and resubmission under a suitable policy.

Later offboarding/permission loss does not rewrite participant history. It blocks action live.

## 8.5 ApprovalStepDecision

```text
ApprovalStepDecision
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
active instance
active step
current participant
User enabled
live canonical approval.act
strict SoD
fresh-auth one-shot consumed iff required
```

Approval stores no password/provider token/provider claim payload.

### ACCEPT semantics

`ACCEPT` means:

> this actor accepts the exact Submission for the responsibility represented by this Step.

It does **not** imply final document approval unless it completes the last required Step.

### ANY

First valid ACCEPT completes Step.

### ALL

Every active participant must ACCEPT.

### RETURN_FOR_CHANGES

Any current authorized participant may return the attempt:

```text
ApprovalInstance → RETURNED
remaining StepInstances → ABORTED
Revision SUBMITTED → DRAFT
Submission + feedback + decisions remain immutable
B3 working_version never resets
```

The participant may attach a decision comment and/or existing ReviewFeedback explaining required changes.

## 8.6 Strict SoD

Before ACCEPT:

```text
actor != Submission creator/submitter under frozen SoD rule
actor has not ACCEPTED another Step of this ApprovalInstance
```

RETURN_FOR_CHANGES is not ACCEPT and does not consume the “one accepted Step” constraint.

No role bypasses SoD.

## 8.7 Withdraw / cancel / reassign

```text
withdraw → instance WITHDRAWN → Revision DRAFT
cancel   → instance CANCELLED → Revision DRAFT
```

Neither operation sets the business Revision to `CANCELLED`; `document.cancel_revision` remains separate Controlled Information authority.

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

V1 reassignment may select only a User from the frozen Step pool who is currently enabled, currently `approval.act`-authorized and SoD-valid.

No generic delegation platform.

## 8.8 due_in_days

```text
due_at = activated_at + due_in_days
```

Overdue is projection/notification information only. No auto-transition, auto-accept, quorum reduction or escalation engine.

---

# 9. Unified approval-route workspace

All active Steps use one product-semantic interaction surface:

```text
exact Submission viewer
source and/or exact PDF rendition
comments / annotations
suggestions / redlines
Step context
ACCEPT
RETURN_FOR_CHANGES
fresh-auth only if Step requires it
```

R10-E chooses actual UI composition and provider adapters.

There is no product route mode switch:

```text
review mode
approval mode
```

A current editor provider may expose `editing`, `suggesting`, `viewing` internally, but those are adapter implementation states.

### Example

Policy:

```text
Step 1 — "Revisão técnica"
  Group Engenharia
  ANY
  fresh-auth = false

Step 2 — "Qualidade"
  Role approver in Area QUAL
  ANY
  fresh-auth = true
```

Both Steps can:

```text
view
comment
annotate
suggest
ACCEPT
RETURN_FOR_CHANGES
```

The only semantic difference is that Step 2 requires fresh-auth before ACCEPT.

---

# 10. ReviewFeedback / annotations / suggestions

## 10.1 ReviewFeedback

The name denotes feedback on the submitted candidate; it does **not** recreate a Review Step type.

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

- feedback Submission must equal ApprovalInstance Submission;
- COMMENT and SUGGESTION are available to any current active participant;
- anchors/proposals use bounded versioned provider-neutral contracts;
- raw editor/viewer/library IDs never become business identity;
- provider may visually display tracked insertions/deletions, highlights and annotations;
- submitted bytes never change;
- B3 EditorialComment remains distinct DRAFT collaboration/submit-gate state.

### Suggestion semantics

A suggestion is **advice**, not content authority.

If a participant considers it mandatory:

```text
create suggestion/comment
→ RETURN_FOR_CHANGES
```

If the participant ACCEPTs, any attached feedback remains advisory historical context; it is not silently applied.

## 10.2 Applying a suggestion after return

```text
S7 remains immutable
feedback remains bound to S7
Revision returns to DRAFT
```

Applying a suggestion is:

```text
read ReviewFeedback
→ provider-neutral adapter maps proposal
→ B3 WorkingContent CAS(expected_working_version)
→ governed content/Artifact changes
→ working_version++
```

Feedback has no write authority over WorkingContent.

V1 does not create a second mandatory feedback-disposition workflow. A future audited APPLIED/DECLINED requirement is a reopen trigger.

---

# 11. In-product viewer semantics

B4 owns the semantic contract; R10-E owns API/UI/provider implementation.

```text
AUTHOR DRAFT
  editable source/provider view
  WorkingContent OCC
  B3 EditorialComments

SUBMITTED / any active Approval Step
  exact immutable Submission
  source viewer where supported
  optional PDF view derived from same Submission
  COMMENT + SUGGESTION overlays
  ACCEPT / RETURN_FOR_CHANGES
  optional one-shot fresh-auth before ACCEPT

EFFECTIVE
  in-product official representation view
  exact Release-selected rendition when one exists
  otherwise source/auxiliary viewer through provider-neutral adapter
  explicit acknowledgement when obligation exists
```

Normal supported viewing must not require download + external app.

Download/export remains separate optional action.

Viewer/editor brands never enter Document/Revision/Submission identity or semantic permissions.

---

# 12. Rendition / representation policy — Controlled Information

## 12.1 DocumentTypeRepresentationPolicy

```text
DocumentTypeRepresentationPolicy
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK SOURCE_ONLY|REQUIRE_RENDITION
  required_format TEXT NULL CHECK closed ContentFormat
```

```text
SOURCE_ONLY       → format NULL
REQUIRE_RENDITION → format NOT NULL
```

Policy controls Release requirement only, not viewer capability.

## 12.2 Rendition

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

  UNIQUE(submission_id,output_format)
```

Immutable semantic success record.

Failed/retried renderer attempts remain R10-D job/effect state and do not create Rendition.

Laws:

- exact input = Submission;
- exact output = Artifact;
- generator/build provenance retained;
- provider/storage location absent;
- renderer output never changes Submission digest.

### PDF viewer without hidden mandatory PDF

`SourceOnly` may still generate an auxiliary PDF for in-product viewing. Release does not wait for it.

`RequireRendition(PDF)` blocks Release until the exact Submission has its semantic PDF Rendition. ReleaseRecord then pins that Rendition as official representation.

Any PDF shown in the approval workspace must have `Rendition.submission_id == ApprovalInstance.submission_id`.

---

# 13. Release

## 13.1 ReleasePlan

```text
ReleasePlan
  submission_id UUID PRIMARY KEY FK RevisionSubmission(id) RESTRICT
  not_before TIMESTAMPTZ NULL
  created_at TIMESTAMPTZ NOT NULL
```

Immutable per Submission. No generic publication-plan engine.

## 13.2 Release gates

```text
human_gate = NoHumanApproval
             OR ACCEPTED ApprovalInstance for same Submission

representation_gate = SourceOnly
                      OR exact required-format Rendition exists

time_gate = no not_before
            OR now >= not_before
```

Candidate must still be the current SUBMITTED open Revision of its Document.

## 13.3 ReleaseRecord

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

`released_at` is the effectivity instant; no second mutable effective-time authority is needed.

## 13.4 Winning Release transaction

```text
BEGIN
  lock Document serialization root
  lock/validate candidate Revision + exact Submission
  prove Revision = SUBMITTED/current open
  prove human/representation/time gates
  identify prior EFFECTIVE Revision

  load DistributionAudienceGroup rows
  lock referenced Group rows in UUID order
  lock existing GroupMembership rows deterministically
  resolve deduplicated concrete Users in this serialized membership cut

  insert unique ReleaseRecord(Submission)
  prior EFFECTIVE → SUPERSEDED
  candidate SUBMITTED → EFFECTIVE
  insert DistributionObligation rows for resolved Users

  // B6 later composes required Audit
  // R10-D later composes notification/search/timer intents
COMMIT
```

Distribution denominator is semantic release truth, not a post-commit best-effort projection.

Concurrent membership mutation is serialized before or after the snapshot. No membership-version subsystem is introduced.

Two Release retries are harmless:

```text
Document serialization
+ UNIQUE(ReleaseRecord.submission_id)
+ B3 one-EFFECTIVE partial uniqueness
→ exactly one semantic winner
```

## 13.5 B3 seam closure

Create-from-template and Release on the template Document share the Document serialization root, preventing stale template origin.

PeriodicReviewRecord and Release on the reviewed Document share the same root, preventing stale review of a Revision that ceased to be current EFFECTIVE.

---

# 14. Distribution

## 14.1 Live audience configuration

Minimum evidenced V1:

```text
DistributionAudienceGroup
  document_id UUID NOT NULL FK Document(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT

  PRIMARY KEY(document_id,group_id)
```

Current configuration only; Group FK participates in B2 hard-delete RESTRICT.

No generic `target_type/target_id`. Direct User/Area/company audience types require real product evidence.

## 14.2 DistributionObligation

Created in the winning Release transaction:

```text
DistributionObligation
  id UUID PRIMARY KEY
  release_id UUID NOT NULL FK ReleaseRecord(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(release_id,user_id)
```

Concrete immutable historical denominator.

Group rename/delete/membership change never rewrites obligation history.

## 14.3 AcknowledgementRecord

```text
AcknowledgementRecord
  id UUID PRIMARY KEY
  obligation_id UUID NOT NULL UNIQUE FK DistributionObligation(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  acknowledged_at TIMESTAMPTZ NOT NULL
```

Immutable.

Rules:

- User must equal obligation User;
- explicit authenticated product action only;
- notification delivery/read, document view, PDF view, source download, search hit do not acknowledge;
- no proxy acknowledgement;
- no mandatory fresh-auth/e-signature V1.

Reminder/delivery jobs are R10-D effects. They may fail without rewriting obligation truth.

---

# 15. AuthZ / B2 coherence

Tenant-wide configuration:

```text
approval_policy.manage
```

Area-targeted execution through `Document.area_id` + exact domain relationship:

```text
approval.act
approval.oversee
approval.reassign
approval.cancel
distribution.manage
distribution.oversee
```

Permission alone never grants arbitrary case access. Participant/ownership/state/SoD predicates still apply.

No mechanism/provider permissions:

```text
pdf.view
editor.suggest
viewer.annotate
renderer.retry
artifact.replace
```

Viewing/feedback are authorized through the owning Document/Approval relationship.

---

# 16. Persistence class × mutation law

| Family | Owner | Mutation law |
|---|---|---|
| DocumentTypeApprovalConfig | Approval | mutable current config |
| ApprovalPolicy | Approval | stable identity/status/display |
| ApprovalPolicyVersion | Approval | immutable |
| ApprovalPolicyStep | Approval | immutable one-type governance-step snapshot |
| ApprovalPolicyLiveGroupRef | Approval | current live config only |
| ApprovalInstance | Approval | explicit lifecycle |
| ApprovalStepInstance | Approval | explicit lifecycle |
| ApprovalParticipantPool | Approval | immutable actor-pool snapshot |
| ApprovalParticipant | Approval | active assignment history |
| ApprovalStepDecision | Approval | immutable |
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

# 17. Structural constraint envelope

```text
ApprovalPolicy.code                                  UNIQUE
ApprovalPolicyVersion(policy_id,version_no)          UNIQUE
ApprovalPolicyStep(version_id,step_order)            UNIQUE
ApprovalPolicyLiveGroupRef.policy_step_id            PK
ApprovalInstance.submission_id                       UNIQUE
ApprovalStepInstance(instance_id,step_order)         UNIQUE
ApprovalParticipantPool(step_id,user_id)             PK
ApprovalParticipant(step_id,user_id)                 UNIQUE
ApprovalStepDecision(step_id,actor_user_id)          UNIQUE
Rendition(submission_id,output_format)                UNIQUE
ReleasePlan.submission_id                            PK
ReleaseRecord.submission_id                          UNIQUE
DistributionAudienceGroup(document_id,group_id)      PK
DistributionObligation(release_id,user_id)           UNIQUE
AcknowledgementRecord.obligation_id                  UNIQUE
```

Cross-row guards must prove:

- actor union validity;
- current live Group ref matches immutable Group snapshot on current Step;
- instance Submission/state coherence;
- feedback Submission = instance Submission;
- official Rendition = same Submission + required format;
- acknowledgement User = obligation User;
- retired Area / disabled NamedUser / missing Group are rejected for new live policy configuration.

No cross-owner CASCADE/SET NULL.

---

# 18. Core transaction contracts

## 18.1 SUBMIT + Approval initialization

B3 freezes Submission. B4 composition:

```text
read DocumentTypeApprovalConfig

if NO_HUMAN_APPROVAL:
  no fake ApprovalInstance
else:
  lock ACTIVE ApprovalPolicy/current version
  validate current live actor refs
  insert ApprovalInstance + all StepInstances
  resolve every Step actor rule to concrete User pool snapshot
  fail if any pool is empty
  activate Step 1 from pool after SoD filter
```

Group resolution must use a deterministic membership cut and compose with B2/B3 lock order without reverse acquisition.

## 18.2 ReviewFeedback append

```text
prove ACTIVE instance / active Step
prove current participant + live approval.act
prove exact Submission
insert immutable COMMENT or SUGGESTION
```

No WorkingContent mutation.

## 18.3 ACCEPT

```text
serialize active Step
prove participant + live AuthZ + SoD
consume fresh-auth iff required
insert immutable ACCEPT
apply ANY|ALL

if Step completes and another Step exists:
  activate next Step from frozen pool minus SoD-ineligible prior acceptors

if final Step completes:
  ApprovalInstance → ACCEPTED
```

No Step type branch exists.

Approval success never directly asserts EFFECTIVE.

## 18.4 RETURN_FOR_CHANGES

```text
insert immutable RETURN_FOR_CHANGES decision
ApprovalInstance → RETURNED
remaining Steps → ABORTED
Revision SUBMITTED → DRAFT
never reset B3 working_version
```

## 18.5 Release

Effectivity + predecessor supersession + Distribution denominator commit atomically as §13.4.

---

# 19. Concurrency / lock laws

B1 remains `READ COMMITTED`.

B2 lock law remains authoritative for User/Area/Group eligibility mutations.

B4 rules:

- ApprovalInstance initialization takes deterministic actor-resolution locks and never turns provider/AuthZ calls into cross-system atomicity;
- active Step decisions serialize the Step row before inserting/evaluating decisions;
- Group audience Release snapshots conflict with concurrent membership add/remove through the same Group/membership rows rather than a new versioning subsystem;
- Release uses Document as the per-document serialization root;
- no global SERIALIZABLE, distributed lock or generic workflow lock service.

Detailed whole-B1–B4 lock-class ordering is an implementation-spec proof obligation before code.

---

# 20. Proof obligations

| Claim | Falsification/proof |
|---|---|
| one semantic Step type | schema/API contains no `purpose=review|approval` discriminator |
| capability orthogonality | comments/suggestions work on every active Step regardless of fresh-auth |
| final approval semantics | intermediate ACCEPT cannot make ApprovalInstance ACCEPTED before final configured Step |
| exact candidate | every decision/feedback/rendition/release resolves same Submission |
| Submission immutability | feedback/suggestion never changes Submission bytes/digest |
| suggestion application | applying returned feedback requires B3 WorkingContent OCC and new Submission later |
| SoD submitter | submitter cannot ACCEPT own Submission |
| SoD multi-step | same User cannot ACCEPT two Steps of same instance |
| ANY | first valid ACCEPT completes exactly one Step |
| ALL | every active participant required; empty denominator never succeeds |
| live AuthZ | frozen participant pool never grants action after permission loss/offboarding |
| frozen pool | later Group/Role membership does not add/remove in-flight candidate actors |
| Group deletion | historical Approval remains intelligible after Group live ref is gone/deleted |
| fresh-auth | Step can require one-shot reauth without a separate Step class |
| due date | overdue cannot mutate Step/Revision lifecycle automatically |
| PDF viewer coherence | approval workspace cannot display PDF rendition from another Submission |
| auxiliary PDF | SourceOnly release does not wait for viewer PDF |
| required rendition | RequireRendition release cannot win without exact required rendition |
| one release | concurrent retries produce one ReleaseRecord/effectivity winner |
| one effective | B3 partial unique backstop survives B4 Release race |
| distribution snapshot | membership change racing Release is fully before or after denominator cut |
| historical denominator | later Group change never rewrites obligation rows |
| acknowledgement truth | view/download/notification never creates acknowledgement |

---

# 21. Adversarial challenge

## F1 — `REVIEW|APPROVAL` is required because signatures differ

Counterexample to the old model:

Veeva allows verdict, comment and eSignature settings to vary orthogonally within workflow tasks. The stronger ceremony is a property of the decision, not proof of a separate universal task type.

**Closure:** retain `requires_reauthentication`; remove Step purpose.

## F2 — reviewers need suggestions but approvers should be read-only

A final approver can discover a defect and needs the same ability to explain/propose correction before `RETURN_FOR_CHANGES`. Preventing suggestions because of Step type removes useful information without protecting Submission immutability.

**Closure:** all feedback is detached; every Step may comment/suggest; only returned DRAFT can apply changes.

## F3 — allowing suggestions on an ACCEPT creates ambiguity

A participant may leave advisory feedback and still accept.

**Closure:** ACCEPT never applies feedback. If feedback is mandatory for correctness, participant uses `RETURN_FOR_CHANGES`. This rule is explicit and provider-independent.

## F4 — removing purpose loses “who reviewed vs who approved” audit meaning

Historical meaning remains recoverable from:

```text
PolicyVersion
Step order
Step label
actor rule/pool
fresh-auth requirement
Step decisions
```

If regulated signature-meaning semantics become a real requirement, add a small explicit `decision_meaning` fact — do not reintroduce a UI-driven Step type.

## F5 — one Step type prevents separate visual experiences

R10-E can show contextual labels, instructions and fresh-auth ceremony without changing domain Step type. Provider/view modes are adapter concerns.

**Closure:** one semantic workspace, contextual presentation.

## F6 — “review” is a real domain concept

Yes, but MetalDocs already has a distinct domain concept named **PeriodicReview**, which reviews an EFFECTIVE Revision on a schedule. Reusing the same noun as an Approval route type introduces ambiguity rather than clarity.

**Closure:** route Steps are governance gates; PeriodicReview remains its own CI concept.

## F7 — generic configurable task capabilities would be more future-proof

It would create a low-code workflow platform without a proven consumer.

**Closure:** fixed collaboration surface + only the evidenced orthogonal knobs (`actor`, `ANY|ALL`, `fresh-auth`, `due`).

## F8 — Group actor history blocks Group deletion

Permanent historical Group FKs would contradict B2.

**Closure:** live Group FK only for current configuration; immutable Step snapshots + concrete User pools preserve history.

## F9 — post-commit distribution target resolution drifts

Group membership may change between Release and async worker execution.

**Closure:** concrete obligations are inserted inside winning Release transaction; notifications remain asynchronous.

## F10 — viewer PDF silently becomes official

A convenient PDF tab can be mistaken for official representation.

**Closure:** representation policy alone decides Release requirement; ReleaseRecord pins official rendition only when required.

---

# 22. Essential vs accidental complexity / YAGNI

## Essential

- exact Submission-bound human governance;
- ordered Steps;
- actor rule resolution;
- ANY/ALL;
- strict SoD;
- optional fresh-auth;
- detached feedback/suggestions;
- immutable decisions;
- immutable Rendition;
- automatic Release;
- concrete Distribution denominator;
- explicit acknowledgement;
- in-product viewer contract.

## Accidental — remove/defer

- `REVIEW|APPROVAL` Step type;
- separate review/approval application modes;
- generic BPM task engine;
- arbitrary capability switches per Step;
- M-of-N;
- branching/DAG routing;
- generic delegation;
- auto-escalation workflow;
- provider-specific editor mode authority;
- mandatory PDF for every DocumentType;
- release-generation duplicate identity;
- dynamic historical distribution denominator;
- read/view as acknowledgement;
- mandatory eSignature for every Step or acknowledgement.

---

# 23. Local vs Global Maximum

### Legacy-compatible Local Maximum

```text
Step purpose = REVIEW|APPROVAL
→ UI/provider mode branch
→ separate feedback rules
→ separate ceremony semantics
```

It makes the redesigned kernel conform to a legacy interaction taxonomy.

### Overengineered non-maximum

```text
generic workflow task platform
→ arbitrary capabilities
→ branching
→ plugins
→ custom verdicts
```

It maximizes flexibility rather than minimizing sustainable complexity.

### Current Global Maximum candidate

```text
one governance Step
+ exact Submission
+ fixed collaborative feedback surface
+ ANY|ALL
+ actor rule
+ optional fresh-auth
+ sequential completion
```

This is the smallest model that preserves every currently evidenced governance invariant while keeping editor/viewer technology replaceable.

---

# 24. Candidate decision

Proposed bounded correction:

```text
REOPEN ONLY:
  ApprovalPolicy Step.purpose = review|approval

REPLACE WITH:
  one Approval/Governance Step semantic
```

Corrected route model:

```text
ApprovalPolicy(version)
  ordered Steps

Step
  label
  actor_rule: NamedUser | Group | RoleInArea
  completion: ANY | ALL
  requires_reauthentication
  due_in_days?

All active Step participants may:
  view exact Submission
  view exact PDF rendition when available
  comment / annotate / suggest
  ACCEPT | RETURN_FOR_CHANGES
```

ApprovalInstance becomes `ACCEPTED` only when all configured Steps complete.

This candidate does **not** reopen:

- exact Submission binding;
- SoD;
- actor rules;
- ANY/ALL;
- optional fresh-auth;
- return/resubmit model;
- Rendition/Release semantics;
- Distribution snapshot/ack semantics;
- B1/B2/B3 substrate.

---

# 25. Reopen triggers

Reopen the one-Step model only if real evidence demonstrates a materially different task class, for example:

- a required non-gating comment-only task inside the same route;
- a legally distinct decision type whose persistence/enforcement differs from ACCEPT rather than merely its label/fresh-auth ceremony;
- a route task that legitimately mutates the submitted candidate without return-to-DRAFT, which would also reopen B3 Submission immutability;
- required parallel/DAG execution not representable by ordered Steps;
- a real regulated signature-meaning model that cannot be expressed as a small decision fact;
- provider-independent collaboration semantics that materially differ by Step and cannot be represented without reintroducing a task class.

Current legacy UI shape, current editor modes or vendor workflow taxonomies are **not** reopen triggers.

---

# 26. Whole-R10 posture

B4 remains **NON-AUTHORITATIVE CANDIDATE** until operator acceptance.

If the operator accepts this correction:

```text
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R9.5 Approval Step purpose distinction = bounded working refinement for Whole-R10 integration
implementation = BLOCKED
next = R10-B5 integrated candidate + B3/B4/B5 coherence challenge
```

The frozen ledger is not silently rewritten in this candidate. Whole-R10 promotion must explicitly carry the bounded correction so fresh sessions cannot accidentally resurrect `purpose=review|approval` from the earlier ledger wording.
