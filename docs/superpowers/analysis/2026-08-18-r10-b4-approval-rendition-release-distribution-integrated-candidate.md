# R10-B4 — Approval + CI-owned Rendition/Release + Distribution — Integrated Candidate

> **Status:** NON-AUTHORITATIVE — SELF-REVIEWED CORRECTED CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B4 or silently rewrite the frozen R3–R9.5 ledger. It consumes the operator-accepted B3 working candidate, records one bounded R9.5 Approval refinement under explicit review, and remains challengeable by B5–F and Whole-R10 review.

> **Bounded reopen under operator review:** remove the frozen `ApprovalPolicy Step purpose = review | approval` discriminator. The replacement is one governance-Step semantic model with collaboration and fresh-auth as orthogonal capabilities. No other R9.5 Approval semantic is reopened by this correction.

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

Current implementation, schemas, OpenAPI, legacy ADRs and old editor modes are evidence only.

External products are comparison evidence only. The useful facts from current research are narrow:

- mature DMS/QMS products expose review, verdict, comments, redlines and stronger authentication/signature ceremony in different combinations;
- there is no universal vendor taxonomy proving that MetalDocs needs a structural `REVIEW | APPROVAL` Step type;
- participant/assignment resolution can occur at task activation rather than forcing all future task owners to freeze at workflow start;
- annotation/redline layers can be kept separate from original document content.

MetalDocs derives its own target from its frozen invariants rather than copying vendor lifecycle/workflow engines.

---

# 2. Operator requirements carried into B4

The operator requires:

1. supported governed documents are normally viewable **inside MetalDocs**; download + external desktop application is not the standard inspection journey;
2. approval-route participants may inspect the exact candidate through a source viewer and/or PDF rendition derived from that exact Submission;
3. route participants may comment, annotate and propose corrections/suggestions;
4. current editor/viewer products remain adapters, never product/domain identity;
5. the route does **not** expose `review` and `approval` as structurally different Step types.

Therefore:

```text
OfficialRepresentationPolicy != viewer capability
provider suggesting/viewing mode != Step identity
feedback/suggestion != mutation of RevisionSubmission
```

`SUBMITTED` remains the immutable review boundary. Suggestions are detached evidence/advice on one exact Submission. Applying a returned suggestion is a new B3 WorkingContent OCC mutation and can only affect a later Submission.

R10-E successor requirement:

> ordinary native V1 controlled-document formats admitted to approval/effective viewing must have a safe in-product inspection path for the exact governed candidate (source viewer and/or controlled rendition). Exact-source download remains a separate action/fallback, not the normal supported journey. Historical/imported unsupported content may retain explicitly bounded fallback behavior without becoming a native authoring/review entitlement.

---

# 3. Method — Evidence → Known / Inferred / Unknown / Deferred

## 3.1 Known — preserved authority

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
- the same User cannot ACCEPT two Steps of one ApprovalInstance;
- reassignment remains qualified and SoD-valid;
- no role, including `tenant_owner`, bypasses domain governance;
- ApprovalInstance binds one exact immutable `RevisionSubmission`;
- return/withdraw returns the same Revision to DRAFT without mutating old Submission;
- resubmission creates a new Submission and, when human approval applies, a new ApprovalInstance;
- fresh-auth consumes Authentication-owned bounded evidence; Approval never stores/challenges passwords;
- Approval evidence preserves actor, Step, policy version, decision, trusted server time and required fresh-auth evidence;
- `return_for_changes` requires a reason V1;
- Rendition is immutable derived representation of one exact Submission with output Artifact/hash + generator/build provenance;
- Approval judges Submission, never renderer output bytes;
- Release is automatic/system-owned; no publish button;
- optional `ReleasePlan.not_before`; winning Release instant is actual effectivity;
- winning Release atomically makes candidate EFFECTIVE and prior EFFECTIVE Revision SUPERSEDED;
- representation policy = `SourceOnly | RequireRendition(ContentFormat)`;
- Distribution is controlled obligation/acknowledgement, not AuthZ/LMS;
- Release snapshots concrete Users and later Group membership never rewrites historical denominator;
- only explicit immutable AcknowledgementRecord completes an obligation;
- B2 Group hard-delete fails while a live typed B4 Group dependency exists;
- B3 exact Submission identity, immutable bytes, one OCC generation, one-EFFECTIVE backstop and no-confirmed-orphan Artifact law remain integration inputs.

## 3.2 Bounded reopen — remove old Step purpose

Previously frozen:

```text
Step.purpose = REVIEW | APPROVAL
```

New evidence + operator review + Structural Inversion Test show that this discriminator primarily encoded legacy editor/UI ceremony:

```text
REVIEW   → suggesting-like editor behavior
APPROVAL → viewing/reauth ceremony
```

The actual invariant is independent:

```text
who owns this governance gate?
how many must ACCEPT?
is fresh-auth required?
what exact Submission are they judging?
what feedback did they leave?
what decision occurred?
```

Corrected semantic:

> **Every route Step is one ordered governance gate over one exact immutable Submission. Collaboration capabilities are available independently of Step identity. Fresh-auth is an explicit orthogonal Step requirement.**

## 3.3 Inferred corrected choices

1. `RevisionSubmission.id` is the shared exact-candidate identity across Approval, Rendition and Release; no duplicate `release_generation` identity.
2. ApprovalPolicy has stable identity + immutable numbered versions.
3. ApprovalPolicyStep has no purpose/stage-kind enum; `label` is business language only.
4. All active route participants may view, comment, annotate and suggest corrections on the exact Submission.
5. Fresh-auth is one-shot bounded evidence tied to a decision. If a Step requires it, **every Step decision** (`ACCEPT` or `RETURN_FOR_CHANGES`) must consume/snapshot it; no hidden outcome-specific exception exists.
6. Approval mode and selected PolicyVersion are immutably snapshotted **per Submission** so later type/policy edits cannot bypass or retroactively add human approval.
7. Representation policy is immutably snapshotted into the Submission's ReleasePlan so later DocumentType policy edits cannot change an in-flight Release gate.
8. Group/Role actor rules remain live selectors until a Step activates. At activation they resolve to concrete User assignments; later membership/grant drift does not silently rewrite that active denominator.
9. A GROUP Step keeps a typed live Group FK while it is PENDING/ACTIVE so Group hard-delete cannot destroy an in-flight actor rule that may still need activation/reassignment.
10. Reassignment may deliberately replace/fill an assignment only with a User who **currently satisfies the same frozen actor rule**, is enabled, currently has `approval.act`, and is SoD-valid. It is not limited to an obsolete frozen pool and it never becomes arbitrary delegation.
11. `SubmissionFeedback` is detached immutable state bound to exact Submission; feedback never has write authority over WorkingContent.
12. Auxiliary PDF Rendition may exist under `SourceOnly`; only the snapshotted representation requirement blocks Release.
13. V1 permits one semantic successful Rendition per `(Submission, ContentFormat)`; failed renderer attempts remain R10-D mechanism state. A real need to supersede a successful rendition is a bounded reopen trigger.
14. Distribution live audience V1 includes Group because B2 proves a real Group consumer; no generic audience polymorphism.
15. Distribution audience configuration is **late-bound at Release**, and its add/remove mutations serialize on the same Document root as Release so the snapshot is unambiguous.
16. Distribution obligations are inserted in the winning Release transaction, never discovered later by a worker.
17. Acknowledgement is explicit authenticated action by the obligated User, governed by current canonical document-read authorization; no new acknowledgement permission/eSignature/fresh-auth is invented.
18. `due_in_days` is attention/SLA data only; no auto-escalation/quorum mutation/lifecycle transition.
19. `approval.cancel` terminates the approval attempt and returns the Revision to DRAFT; cancelling the business Revision remains `document.cancel_revision`.

## 3.4 Unknown — kept unknown

- future direct User/Area/company Distribution audience types;
- mandatory audited APPLIED/DECLINED disposition for each suggestion;
- richer provider-neutral anchor schemas for future formats;
- post-release semantic re-render/supersession of a successful Rendition;
- future fresh-auth/eSignature requirement for Acknowledgement;
- exact viewer/editor adapter implementation;
- a genuine non-gating comment-only task class inside the same route;
- a future legally distinct decision meaning requiring more than Step label + decision evidence.

These are reopen triggers or later-stage concerns, not launch defaults.

## 3.5 Deferred

```text
Evidence/Dossier/Records + retention/hold/disposition        → B5
Audit skeleton/final same-commit cross-owner matrix          → B6
physical Artifact storage/malware/restore                    → R10-C
worker/outbox/timers/notification delivery/projections       → R10-D
API/frontend/viewer/editor/provider adapter journeys         → R10-E
historical migration/cutover/legacy deletion                 → R10-F
```

---

# 4. Root Cause

B4 addresses two structural defect classes.

## 4.1 Collapsed governance truth

Human decision, derived representation, effectivity and reader obligation are different authorities/times/failure modes. One floating document status or multiple paths that each infer “current content” can produce:

```text
approval candidate != rendition candidate != released candidate
```

or:

```text
release time audience != later async-resolved audience
```

## 4.2 Structural inversion from legacy interaction modes

A Step type derived from editor behavior makes provider/UI ceremony define domain taxonomy.

The redesign must instead make UI/provider modes derive from:

```text
exact Submission state
participant relationship
feedback capability
fresh-auth requirement
```

---

# 5. Target invariant

> **Every Approval route Step is one ordered governance gate over one exact immutable RevisionSubmission. At Step activation, the actor rule resolves to a concrete assignment denominator. A Step is satisfied only by configured ANY/ALL ACCEPT decisions under current Authorization, SoD and any required fresh-auth evidence. Only completion of all Steps satisfies the human gate. Every Rendition and ReleaseRecord binds the same Submission. Winning Release alone establishes effectivity and, in the same commit, freezes the concrete Distribution obligations selected from the serialized live audience configuration.**

Corollaries:

```text
Step label                 != Step type
feedback capability        != content mutation
fresh-auth                 != separate task class
ACCEPT on intermediate Step!= final ApprovalInstance acceptance
Approval complete          != Release complete
Rendition                   != source Submission
PDF preview                 != automatically official representation
Release                     != human publish action
Distribution                != Authorization
View/download               != Acknowledgement
Group membership            != historical obligation denominator
PeriodicReview              != Approval route Step
```

---

# 6. Credible alternatives

## A — keep `REVIEW | APPROVAL`

Legacy-compatible, but couples provider ceremony to domain type, duplicates orthogonal facts and collides semantically with PeriodicReview.

**Reject — Local Maximum.**

## B — generic workflow/task capability platform

Arbitrary `can_comment/can_sign/can_edit/custom verdict/branch/plugin` configuration.

**Reject — overengineered non-maximum.** No real V1 consumer requires a low-code BPM kernel.

## C — one governance Step + fixed collaboration surface + small orthogonal controls

```text
Step
  label
  actor_rule
  completion ANY|ALL
  requires_reauthentication
  due_in_days?
```

Universal participant interaction:

```text
view exact Submission
view exact PDF rendition when available
comment / annotate / suggest
ACCEPT | RETURN_FOR_CHANGES
```

**Recommended Global Maximum.** Smallest model preserving all currently evidenced governance invariants while keeping editor/viewer replaceable.

---

# 7. Approval configuration

## 7.1 DocumentTypeApprovalConfig — current configuration

```text
DocumentTypeApprovalConfig
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_HUMAN_APPROVAL|USE_POLICY
  approval_policy_id UUID NULL FK ApprovalPolicy(id) RESTRICT
```

```text
NO_HUMAN_APPROVAL → policy NULL
USE_POLICY        → policy NOT NULL
```

Current configuration is Tenant-wide `approval_policy.manage` state. It affects **future Submissions only** after each Submission snapshots its requirement.

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

  UNIQUE(approval_policy_id,version_no)
```

Immutable/append-only. A new version becomes current only through one atomic complete-version promotion; partially built versions never become selectable. “Rollback” creates a new version.

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

  UNIQUE(policy_version_id,step_order)
```

No `purpose`, `stage_kind`, `review_mode` or `approval_mode` exists.

Closed actor-union CHECK:

```text
NAMED_USER   → named_user_id only
GROUP        → group snapshot only
ROLE_IN_AREA → role_code + area_id only
```

New/current policy versions reject disabled NamedUser, retired Area and missing Group.

## 7.5 ApprovalPolicyLiveGroupRef — current configuration dependency

```text
ApprovalPolicyLiveGroupRef
  policy_step_id UUID PRIMARY KEY FK ApprovalPolicyStep(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT
```

Exists for GROUP Steps only while that PolicyVersion is current/live configuration. It prevents Group hard-delete while new Submissions can still select that rule.

Historical policy meaning stays in the immutable snapshot; old non-live versions do not permanently prevent Group deletion.

---

# 8. Per-Submission governance snapshots

## 8.1 SubmissionApprovalRequirement — Approval-owned immutable plan

Created in the same local transaction as B3 SUBMIT:

```text
SubmissionApprovalRequirement
  submission_id UUID PRIMARY KEY FK RevisionSubmission(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_HUMAN_APPROVAL|USE_POLICY
  policy_version_id UUID NULL FK ApprovalPolicyVersion(id) RESTRICT
  snapshotted_at TIMESTAMPTZ NOT NULL
```

```text
NO_HUMAN_APPROVAL → policy_version_id NULL
USE_POLICY        → policy_version_id NOT NULL
```

This is the historical answer to:

> what human governance did **this Submission** require?

Later edits to DocumentTypeApprovalConfig, ApprovalPolicy status or newer policy versions never alter this requirement.

NoHumanApproval therefore does not need a fake ApprovalInstance/System approver, while Release still has immutable proof that the human gate for that Submission is intentionally absent.

## 8.2 ReleasePlan — CI-owned immutable per-Submission plan

```text
ReleasePlan
  submission_id UUID PRIMARY KEY FK RevisionSubmission(id) RESTRICT
  representation_mode TEXT NOT NULL CHECK SOURCE_ONLY|REQUIRE_RENDITION
  required_format TEXT NULL CHECK closed ContentFormat vocabulary
  not_before TIMESTAMPTZ NULL
  created_at TIMESTAMPTZ NOT NULL
```

```text
SOURCE_ONLY       → required_format NULL
REQUIRE_RENDITION → required_format NOT NULL
```

`representation_mode/required_format` are system snapshots of the current DocumentTypeRepresentationPolicy at SUBMIT. `not_before` is the explicit Submission release plan value accepted by the authorized submit journey.

Later representation-policy edits affect future Submissions only.

---

# 9. Approval execution

## 9.1 ApprovalInstance

Created only when `SubmissionApprovalRequirement.mode=USE_POLICY`:

```text
ApprovalInstance
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL UNIQUE FK RevisionSubmission(id) RESTRICT
  policy_version_id UUID NOT NULL FK ApprovalPolicyVersion(id) RESTRICT
  status TEXT NOT NULL CHECK ACTIVE|ACCEPTED|RETURNED|WITHDRAWN|CANCELLED
  started_at TIMESTAMPTZ NOT NULL
  completed_at TIMESTAMPTZ NULL
```

Cross-row guard requires `policy_version_id` to equal the SubmissionApprovalRequirement snapshot.

No duplicate submitter/document/revision/hash authority is stored here; those are resolved through Submission.

## 9.2 ApprovalStepInstance

```text
ApprovalStepInstance
  id UUID PRIMARY KEY
  approval_instance_id UUID NOT NULL FK ApprovalInstance(id) RESTRICT
  policy_step_id UUID NOT NULL FK ApprovalPolicyStep(id) RESTRICT
  status TEXT NOT NULL CHECK PENDING|ACTIVE|COMPLETED|ABORTED
  activated_at TIMESTAMPTZ NULL
  participants_resolved_at TIMESTAMPTZ NULL
  due_at TIMESTAMPTZ NULL
  completed_at TIMESTAMPTZ NULL

  UNIQUE(approval_instance_id,policy_step_id)
```

Step label/completion/fresh-auth/actor semantics are read from the pinned immutable PolicyStep; they are not duplicated as second authority on the StepInstance.

## 9.3 ApprovalStepLiveGroupRef — in-flight dependency

For a GROUP PolicyStep:

```text
ApprovalStepLiveGroupRef
  step_instance_id UUID PRIMARY KEY FK ApprovalStepInstance(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT
```

Created when the ApprovalInstance is initialized and retained while the Step is PENDING or ACTIVE. It is deleted explicitly when the Step/Instance becomes terminal.

This preserves B2 Group hard-delete semantics:

```text
current policy needs Group     → delete blocked
pending/active instance needs Group → delete blocked
history only                   → Group may be deleted
```

## 9.4 Step activation — concrete participant snapshot

When one Step becomes ACTIVE, its frozen actor rule resolves **at activation time** against current Organization/Authorization data in one database snapshot. Then strict SoD exclusions are applied and concrete assignments are inserted.

```text
ApprovalParticipant
  id UUID PRIMARY KEY
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  assigned_at TIMESTAMPTZ NOT NULL
  ended_at TIMESTAMPTZ NULL
  assignment_kind TEXT NOT NULL CHECK ORIGINAL|REASSIGNED
  replaces_participant_id UUID NULL FK ApprovalParticipant(id) RESTRICT
  assigned_by_user_id UUID NULL FK User(id) RESTRICT
  assignment_reason TEXT NULL
```

Initial assignments:

```text
assignment_kind = ORIGINAL
replaces_participant_id = NULL
assigned_by_user_id = NULL
assignment_reason = NULL
```

Reassigned assignment:

```text
assignment_kind = REASSIGNED
replaces_participant_id = old assignment
assigned_by_user_id = overseer
assignment_reason = nonblank
```

Partial uniqueness prevents duplicate active assignment for the same User/Step.

Actor resolution rules:

- `NamedUser` → exact named User if currently enabled + `approval.act` authorized + SoD-valid;
- `Group` → current concrete Group members who are enabled + `approval.act` authorized + SoD-valid;
- `RoleInArea` → current concrete Users effectively holding that role in that Area and `approval.act`, after SoD.

Once resolved, later membership/grant changes do **not** silently add/remove active assignments. Action-time Authorization remains live, so later permission loss/offboarding blocks action without rewriting history.

If activation resolves zero actionable participants:

```text
Step remains ACTIVE but unsatisfied
ANY/ALL never vacuously succeeds
```

An authorized reassignment may later fill the Step only with a currently qualified User under the same actor rule. Otherwise the attempt must be returned/cancelled and resubmitted under suitable governance.

## 9.5 Reassignment

Reassignment is an explicit mutation of current assignment state, not a generic delegation platform.

Rules:

- only an undecided active assignment may be replaced;
- old assignment receives terminal `ended_at`;
- new assignment points to `replaces_participant_id`;
- new User must currently satisfy the frozen PolicyStep actor rule, be enabled, have live `approval.act`, and be SoD-valid;
- an unfilled zero-participant Step may receive a REASSIGNED assignment with no replacement only when a currently qualified User now exists;
- an ACCEPTed assignment cannot be reassigned;
- NamedUser policy does not silently permit substitution to a different User; if the named actor cannot act, governance must be changed through return/cancel + resubmission rather than policy bypass.

Audit B6 later records the transversal event; ApprovalParticipant rows remain domain assignment truth.

## 9.6 ApprovalStepDecision

```text
ApprovalStepDecision
  id UUID PRIMARY KEY
  approval_instance_id UUID NOT NULL FK ApprovalInstance(id) RESTRICT
  step_instance_id UUID NOT NULL FK ApprovalStepInstance(id) RESTRICT
  participant_id UUID NOT NULL FK ApprovalParticipant(id) RESTRICT
  actor_user_id UUID NOT NULL FK User(id) RESTRICT
  outcome TEXT NOT NULL CHECK ACCEPT|RETURN_FOR_CHANGES
  decision_reason TEXT NULL
  fresh_auth_schema TEXT NULL
  fresh_auth_snapshot JSONB NULL
  decided_at TIMESTAMPTZ NOT NULL
```

Immutable.

`actor_user_id` and `approval_instance_id` are intentionally denormalized **only for structural enforcement/evidence**, with DB guards requiring exact equality to the referenced Participant/StepInstance relationships.

Decision-time gates:

```text
ApprovalInstance ACTIVE
Step ACTIVE
participant active (ended_at NULL)
actor == participant User
User enabled
live canonical approval.act for this case
strict SoD for ACCEPT
fresh-auth evidence iff PolicyStep requires it
```

Outcome laws:

```text
RETURN_FOR_CHANGES → nonblank decision_reason REQUIRED
ACCEPT             → reason optional
```

Fresh-auth law:

```text
PolicyStep.requires_reauthentication = true
  → fresh_auth_schema + bounded fresh_auth_snapshot REQUIRED

requires_reauthentication = false
  → fresh_auth evidence absent
```

The bounded snapshot is Approval-owned historical decision evidence copied from Authentication `FreshAuthEvidence`, for example:

```text
session_id snapshot value (no FK dependency)
verified_at
provider_auth_time?
acr?
amr?
```

No credential, password, bearer token or arbitrary provider claim payload is stored.

Because `requires_reauthentication` is a Step property, both ACCEPT and RETURN_FOR_CHANGES decisions satisfy the same requirement. A future outcome-specific ceremony rule is a bounded reopen, not an implicit exception.

### Structural SoD

Strongest reasonable backstops:

```text
UNIQUE(approval_instance_id,actor_user_id)
  WHERE outcome='ACCEPT'
```

prevents the same User from ACCEPTing two Steps of one instance even if an application check is bypassed.

A DB cross-row guard/tripwire also rejects ACCEPT when actor equals the B3 Submission submitter or the owning DocumentRevision creator. Application checks provide friendly failure; DB enforcement covers all serving write paths.

No role bypasses SoD.

### ANY / ALL

`ANY`:

- first valid ACCEPT completes the Step;
- remaining undecided assignments become no-longer-required through Step terminal state, not fabricated decisions.

`ALL`:

- every currently active assignment must ACCEPT;
- empty assignment set never satisfies the Step.

### RETURN_FOR_CHANGES

Any active qualified participant may return the attempt. In one local transaction:

```text
insert immutable RETURN_FOR_CHANGES decision + reason + required fresh-auth snapshot
ApprovalInstance → RETURNED
remaining Steps → ABORTED
Revision SUBMITTED → DRAFT
B3 working_version never resets
```

Submission, prior decisions and SubmissionFeedback remain immutable.

## 9.7 Withdraw / cancel

```text
withdraw → instance WITHDRAWN → Revision DRAFT
cancel   → instance CANCELLED → Revision DRAFT
```

Neither sets business Revision `CANCELLED`; `document.cancel_revision` remains separate Controlled Information authority.

## 9.8 due_in_days

```text
due_at = activated_at + due_in_days
```

Overdue is projection/notification information only. No automatic transition, acceptance, quorum reduction or escalation workflow exists.

---

# 10. Unified approval workspace + SubmissionFeedback

## 10.1 SubmissionFeedback

The domain vocabulary deliberately avoids `ReviewFeedback` because MetalDocs already has `PeriodicReview`.

```text
SubmissionFeedback
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

Guards:

- feedback Submission = ApprovalInstance Submission;
- feedback Step = active Step of the instance;
- author = currently active participant with live `approval.act`;
- actor may append feedback only **before recording their Step decision**;
- anchor/proposal payloads use bounded versioned provider-neutral schemas;
- raw editor/viewer/library IDs never become business identity.

COMMENT and SUGGESTION are available on every active governance Step.

A suggestion is advice, never content authority. If mandatory, participant records feedback and chooses RETURN_FOR_CHANGES. ACCEPT never applies feedback.

After return:

```text
old Submission remains immutable
feedback remains bound to old Submission
Revision becomes DRAFT
```

Applying a suggestion requires:

```text
read SubmissionFeedback
→ provider-neutral adapter maps proposal
→ B3 WorkingContent CAS(expected_working_version)
→ content/Artifact/structured state changes as required
→ working_version++
→ later SUBMIT creates a new immutable Submission
```

No second feedback-resolution workflow is invented V1.

## 10.2 In-product viewing contract

```text
AUTHOR DRAFT
  editable source/provider view
  B3 WorkingContent OCC
  B3 EditorialComments

SUBMITTED / active Approval Step
  exact immutable Submission
  source viewer where supported
  exact PDF rendition tab when available
  SubmissionFeedback overlays
  ACCEPT / RETURN_FOR_CHANGES
  fresh-auth ceremony when Step requires it

EFFECTIVE
  exact Release-selected official rendition when required
  otherwise source/controlled auxiliary viewer
  explicit acknowledgement when obligation exists
```

Normal supported viewing does not require external application download.

No semantic mechanism permissions such as `pdf.view`, `editor.suggest` or `viewer.annotate` are added. Access derives from the owning Document/Approval relationship and frozen B2 Authorization.

---

# 11. Rendition / representation policy — Controlled Information

## 11.1 DocumentTypeRepresentationPolicy — current config

```text
DocumentTypeRepresentationPolicy
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK SOURCE_ONLY|REQUIRE_RENDITION
  required_format TEXT NULL CHECK closed ContentFormat vocabulary
```

This is future-Submission current configuration. Release never consults floating current policy for an existing Submission; it uses ReleasePlan snapshot.

Tenant-wide configuration uses existing `document_type.manage`; no renderer/viewer permission family is created.

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

  UNIQUE(submission_id,output_format)
```

Immutable semantic success record.

Cross-row guard requires `output_format == Artifact.content_format`.

### Semantic confirmation seam

External renderer success is not yet a confirmed MetalDocs Rendition. The local semantic confirmation transaction must atomically:

```text
insert/confirm exact immutable output Artifact
+
insert Rendition referencing that Artifact and exact Submission
+
required B6 Audit / R10-D intents when later specified
```

or roll back. A confirmed output Artifact cannot exist without its typed Rendition owner relation, preserving B3 no-confirmed-orphan law.

Provider bucket/key/version is never Rendition identity.

Failed/retried rendering attempts are R10-D mechanism state and create no Rendition row.

### PDF viewer without hidden mandatory PDF

`SourceOnly` ReleasePlan may still receive an auxiliary PDF Rendition for in-product viewing; Release does not wait for it.

`RequireRendition(PDF)` blocks Release until the exact Submission has its successful PDF Rendition. ReleaseRecord then pins that exact Rendition as official representation.

Approval workspace may show only a Rendition with `submission_id == ApprovalInstance.submission_id`.

---

# 12. Release

## 12.1 Gates use immutable per-Submission snapshots

```text
human_gate = SubmissionApprovalRequirement.mode=NO_HUMAN_APPROVAL
             OR matching ApprovalInstance is ACCEPTED

representation_gate = ReleasePlan.representation_mode=SOURCE_ONLY
                      OR exact required-format Rendition exists

time_gate = ReleasePlan.not_before IS NULL
            OR now >= not_before
```

Current DocumentType approval/representation configuration is irrelevant to this already-submitted attempt.

Candidate must still be the current SUBMITTED open Revision of its Document.

## 12.2 ReleaseRecord

```text
ReleaseRecord
  id UUID PRIMARY KEY
  submission_id UUID NOT NULL UNIQUE FK RevisionSubmission(id) RESTRICT
  official_rendition_id UUID NULL FK Rendition(id) RESTRICT
  prior_effective_revision_id UUID NULL FK DocumentRevision(id) RESTRICT
  released_at TIMESTAMPTZ NOT NULL
```

Immutable effectivity evidence.

```text
ReleasePlan SOURCE_ONLY
  → official_rendition_id NULL

ReleasePlan REQUIRE_RENDITION
  → official_rendition_id required
  → correct format
  → same exact Submission
```

`released_at` is effectivity instant; no second mutable effective-time authority is introduced.

## 12.3 Winning Release transaction

Distribution audience configuration mutations use the same `Document FOR UPDATE` serialization root as Release. Therefore after Release owns the Document root, the configured audience set cannot change underneath the snapshot.

Winning transaction:

```text
BEGIN
  lock Document serialization root
  lock/validate candidate Revision + exact Submission
  load immutable SubmissionApprovalRequirement + ReleasePlan
  prove Revision = SUBMITTED/current open
  prove human/representation/time gates
  identify prior EFFECTIVE Revision

  load stable DistributionAudienceGroup set under Document root
  lock referenced Group rows in UUID order
  lock existing GroupMembership rows deterministically
  resolve deduplicated concrete Users in this serialized membership cut

  insert unique ReleaseRecord(Submission)
  prior EFFECTIVE → SUPERSEDED
  candidate SUBMITTED → EFFECTIVE
  insert DistributionObligation rows for resolved Users

  // B6 composes required Audit
  // R10-D composes required durable notification/search/timer intents
COMMIT
```

Audience semantics are explicitly **late-bound at Release**. A config/membership change racing Release is either before or after the winning snapshot; it is never a post-commit guess.

Two Release retries:

```text
Document serialization
+ UNIQUE(ReleaseRecord.submission_id)
+ B3 one-EFFECTIVE partial uniqueness
→ one semantic winner
```

## 12.4 B3 seam closure

- create-from-template and Release on the template Document share the same Document serialization root, preventing stale template origin;
- PeriodicReviewRecord and Release on the reviewed Document share the same root, preventing stale periodic review of a Revision that ceased to be current EFFECTIVE.

---

# 13. Distribution

## 13.1 DistributionAudienceGroup — current config

```text
DistributionAudienceGroup
  document_id UUID NOT NULL FK Document(id) RESTRICT
  group_id UUID NOT NULL FK Group(id) RESTRICT

  PRIMARY KEY(document_id,group_id)
```

Current configuration only. Add/remove operations:

```text
require Area-targeted distribution.manage for Document.area_id
lock Document serialization root
validate live Group
mutate current config
```

Group FK participates in B2 hard-delete RESTRICT.

No generic `target_type/target_id`. Direct User/Area/company audience types require real product evidence.

## 13.2 DistributionObligation — immutable denominator

Created only in the winning Release transaction:

```text
DistributionObligation
  id UUID PRIMARY KEY
  release_id UUID NOT NULL FK ReleaseRecord(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(release_id,user_id)
```

Group rename/delete/membership change never rewrites historical obligations.

Distribution never grants document access; it records obligation only.

## 13.3 AcknowledgementRecord

```text
AcknowledgementRecord
  obligation_id UUID PRIMARY KEY FK DistributionObligation(id) RESTRICT
  acknowledged_at TIMESTAMPTZ NOT NULL
```

Immutable.

Acknowledgement write requires:

```text
current Session User == obligation.user_id
current canonical document.read_effective ALLOW on released Document
obligation still exists
no prior acknowledgement
```

No new semantic `document.acknowledge` permission is invented. The obligation relationship + existing document-read permission define the case-specific write authority.

Notification delivery/read, viewer open, PDF view, source download and search hit never acknowledge.

No proxy acknowledgement and no mandatory fresh-auth/eSignature V1.

Reminder/delivery work is R10-D effect state and may fail without rewriting obligation truth.

---

# 14. AuthZ / B2 coherence

Tenant-wide configuration:

```text
approval_policy.manage
  ApprovalPolicy / version promotion / DocumentTypeApprovalConfig

document_type.manage
  DocumentTypeRepresentationPolicy
```

Area-targeted through `Document.area_id` + exact domain relationship:

```text
approval.act
approval.oversee
approval.reassign
approval.cancel
distribution.manage
distribution.oversee
```

Case-specific approval participant access can expose the exact Submission required for that task even though an `approver` role has no blanket working/history access; the participant relationship is the domain relationship predicate already anticipated by B2.

Acknowledgement uses `document.read_effective` + ownership of the exact DistributionObligation.

No provider/mechanism permissions:

```text
artifact.read
pdf.view
editor.suggest
viewer.annotate
renderer.retry
```

No B4 fact grants arbitrary cross-Document access.

---

# 15. Persistence class × mutation law

| Family | Owner | Class / mutation law |
|---|---|---|
| DocumentTypeApprovalConfig | Approval | semantic current config / mutable |
| ApprovalPolicy | Approval | semantic stable identity/status |
| ApprovalPolicyVersion | Approval | semantic config history / immutable |
| ApprovalPolicyStep | Approval | semantic config snapshot / immutable |
| ApprovalPolicyLiveGroupRef | Approval | current live dependency / mutable add-remove |
| SubmissionApprovalRequirement | Approval | semantic per-Submission plan / immutable |
| ApprovalInstance | Approval | semantic execution / explicit state machine |
| ApprovalStepInstance | Approval | semantic execution / explicit state machine |
| ApprovalStepLiveGroupRef | Approval | current in-flight dependency / mutable add-remove |
| ApprovalParticipant | Approval | current + historical assignment state / terminal end |
| ApprovalStepDecision | Approval | semantic decision evidence / immutable |
| SubmissionFeedback | Approval | semantic feedback evidence / immutable |
| DocumentTypeRepresentationPolicy | Controlled Information | semantic current config / mutable |
| ReleasePlan | Controlled Information | semantic per-Submission plan / immutable |
| Rendition | Controlled Information | semantic derived representation / immutable |
| ReleaseRecord | Controlled Information | semantic effectivity evidence / immutable |
| DistributionAudienceGroup | Distribution | semantic current config / mutable |
| DistributionObligation | Distribution | semantic historical denominator / immutable |
| AcknowledgementRecord | Distribution | semantic acknowledgement / immutable |

No cross-owner CASCADE/SET NULL.

---

# 16. Structural constraint / enforcement envelope

```text
ApprovalPolicy.code                                      UNIQUE
ApprovalPolicyVersion(policy_id,version_no)              UNIQUE
ApprovalPolicyStep(version_id,step_order)                UNIQUE
ApprovalPolicyLiveGroupRef.policy_step_id                PK
SubmissionApprovalRequirement.submission_id              PK
ApprovalInstance.submission_id                           UNIQUE
ApprovalStepInstance(instance_id,policy_step_id)         UNIQUE
ApprovalStepLiveGroupRef.step_instance_id                PK
ApprovalParticipant active(step_id,user_id)              partial UNIQUE
ApprovalStepDecision(step_id,actor_user_id)              UNIQUE
ApprovalStepDecision ACCEPT(instance_id,actor_user_id)   partial UNIQUE
ReleasePlan.submission_id                                PK
Rendition(submission_id,output_format)                    UNIQUE
ReleaseRecord.submission_id                              UNIQUE
DistributionAudienceGroup(document_id,group_id)          PK
DistributionObligation(release_id,user_id)               UNIQUE
AcknowledgementRecord.obligation_id                      PK
```

DB/application guards prove:

- actor-rule closed union;
- live Group refs match Group snapshots/rules they protect;
- ApprovalInstance policy version = SubmissionApprovalRequirement snapshot;
- StepInstance belongs to the pinned policy version;
- Participant/Decision instance/step/User references are internally coherent;
- creator/submitter ACCEPT SoD;
- required fresh-auth snapshot presence/absence follows pinned PolicyStep;
- RETURN_FOR_CHANGES has nonblank reason;
- SubmissionFeedback Submission/Step/participant coherence;
- Rendition format = output Artifact format;
- Release official rendition = required format + same Submission;
- retired Area/disabled NamedUser/missing Group fail new live configuration;
- acknowledgement derives actor from obligation and cannot be duplicated.

Immutable evidence rows are non-updatable/non-deletable by ordinary serving trust. Mixed rows have DB-enforced immutable-column/terminal-state laws.

---

# 17. Core transaction contracts

## 17.1 SUBMIT + B4 governance initialization

B3 remains owner of coherent Submission freeze. B4 composes into the same local transaction:

```text
read/lock current DocumentTypeApprovalConfig
select/lock current complete ApprovalPolicyVersion when required
read current DocumentTypeRepresentationPolicy

B3 performs coherent SUBMIT freeze and inserts immutable RevisionSubmission

insert immutable SubmissionApprovalRequirement
insert immutable ReleasePlan including representation snapshot + not_before

if USE_POLICY:
  insert ApprovalInstance
  insert all ApprovalStepInstances
  insert ApprovalStepLiveGroupRef for every pending GROUP Step
  activate Step 1
  resolve Step 1 concrete participants from one current DB snapshot

transition Revision → SUBMITTED
consume B3 working_version generation
COMMIT
```

No successful SUBMITTED attempt exists without its immutable human/representation Release requirements.

Policy/version/config mutation uses compatible locks so SUBMIT linearizes before or after a configuration change; it never combines a floating old/new governance plan.

## 17.2 Step activation

```text
serialize ApprovalInstance/current Step transition
mark next Step ACTIVE
resolve frozen actor rule using one current DB snapshot
apply enabled/User + approval.act + SoD predicates
insert concrete ORIGINAL ApprovalParticipant rows
set participants_resolved_at / due_at
```

Membership/grant drift after resolution does not change assignments automatically; action-time AuthZ remains live.

## 17.3 SubmissionFeedback append

```text
prove instance/Step ACTIVE
prove active participant + live approval.act
prove actor has no Step decision yet
prove exact Submission
insert immutable COMMENT or SUGGESTION
```

No WorkingContent mutation.

## 17.4 Step decision

```text
serialize active Step
prove active participant + live AuthZ
consume/snapshot fresh-auth iff Step requires it

if RETURN_FOR_CHANGES:
  require nonblank reason
  insert immutable decision
  instance RETURNED
  remaining Steps ABORTED
  Revision SUBMITTED → DRAFT
  never reset working_version

if ACCEPT:
  prove creator/submitter + prior-Step SoD
  insert immutable ACCEPT
  evaluate ANY|ALL

  if Step completes and another Step exists:
    terminalize current Step/group dependency
    activate next Step + resolve its participants

  if final Step completes:
    ApprovalInstance → ACCEPTED
```

Approval success never asserts EFFECTIVE directly.

## 17.5 Reassign

```text
serialize active Step
prove from assignment undecided/active when present
resolve proposed User against same frozen actor rule NOW
prove enabled + approval.act + SoD
end old assignment if present
insert REASSIGNED participant with reason/actor/replacement link
```

No route/policy mutation occurs.

## 17.6 Rendition semantic confirmation

```text
provider/worker obtains derived bytes outside semantic DB atomicity

BEGIN
  prove exact source Submission
  confirm immutable output Artifact facts
  insert Rendition as first typed semantic owner
  B6/R10-D later compose required audit/intent
COMMIT
```

## 17.7 Release

Effectivity, predecessor supersession, ReleaseRecord and concrete Distribution denominator commit atomically as §12.3.

## 17.8 Acknowledgement

```text
prove authenticated User owns obligation
prove current document.read_effective
insert immutable AcknowledgementRecord once
```

---

# 18. Concurrency / lock law

B1 remains `READ COMMITTED`. B2 User/Area/Group lifecycle and offboarding laws remain authoritative.

B4 adds these serialization roots:

- Approval config/version promotion serializes on its stable ApprovalPolicy/config row;
- Step decisions/reassignment serialize the active ApprovalStepInstance;
- Release and DistributionAudienceGroup mutations serialize on `Document`;
- Group live FKs make Group deletion race fail closed without preserving historical Group rows forever;
- Release locks exact audience Groups and existing memberships deterministically before concrete obligation snapshot;
- B3 template creation/PeriodicReview and B4 Release share Document serialization where already required.

For Approval participant resolution, the concrete set is the database snapshot at Step activation; it is not a claim that Organization membership remains frozen before activation. Action-time AuthZ is always re-evaluated.

No global SERIALIZABLE, advisory-lock framework, distributed lock or generic workflow lock service is introduced.

The exact whole B1–B4 SQL lock-class order remains an implementation-spec proof obligation, but B4 may not require a reverse acquisition that violates promoted B2 lifecycle ordering. If implementation evidence shows the transaction contracts cannot be linearized under that law, B4 reopens before code rather than weakening B2 silently.

---

# 19. Integrated Method / global-coherence review

## 19.1 Authority coherence

**PASS after corrections.**

- Authentication still owns provider challenge/session/assurance evidence.
- Approval owns only consumed decision-evidence snapshot and approval lifecycle.
- Authorization remains live B2 authority; participant snapshots never become permission snapshots.
- Controlled Information owns representation/effectivity.
- Distribution owns obligations/acknowledgements and grants no access.
- Artifact remains exact-byte authority; Rendition owns only the semantic relation to output Artifact.
- Notifications/Search remain projections/effects.

No new provider identity or mechanism permission enters the kernel.

## 19.2 B1 relational coherence

**PASS after corrections.**

- UUID typed FKs;
- no universal tenant/company column;
- cross-owner RESTRICT/NO ACTION only;
- bounded JSON only for fresh-auth/feedback anchors;
- immutable/append-only evidence explicit;
- READ COMMITTED + narrow constraints/locks;
- same-local-commit authoritative facts where frozen atomicity requires it.

## 19.3 B2 Organization/AuthZ coherence

**PASS after corrections.**

- disabled Users cannot act;
- retired Areas cannot enter new policy configuration;
- Group hard-delete fails against current Policy, in-flight Step and Distribution current refs but is not blocked forever by historical snapshots;
- participant relationship opens exact case access but never snapshots grants;
- no role/provider bypass;
- fresh-auth snapshot follows promoted B2 consumer contract.

## 19.4 B3 Controlled Information coherence

**PASS after corrections.**

- every Approval/feedback/Rendition/Release binds one immutable Submission;
- suggestions never mutate SUBMITTED bytes;
- return-to-DRAFT preserves old Submission and B3 OCC generation;
- Rendition Artifact confirmation closes typed ownership atomically;
- Release uses B3 Document serialization root for stale-template/PeriodicReview races;
- no second review artifact/digest authority is introduced.

## 19.5 Frozen R9.5 coherence

**PASS WITH ONE EXPLICIT BOUNDED REFINEMENT.**

Preserved:

- sequential workflow;
- actor rules;
- ANY/ALL;
- SoD;
- fresh-auth;
- accept/return;
- versioned policy;
- exact Submission;
- automatic Release;
- immutable Rendition;
- concrete Distribution snapshot;
- explicit acknowledgement.

Refined only:

```text
REMOVE Step.purpose REVIEW|APPROVAL
ADD one governance-Step semantic with orthogonal collaboration/fresh-auth
```

This refinement must be carried explicitly into Whole-R10 authority before final ratification; the old frozen ledger wording must not silently coexist as equal authority.

## 19.6 North-star / provider independence

**PASS.**

- editor/viewer/renderer modes are adapters;
- PDF can be in-product viewing without becoming universally official;
- Keycloak remains AuthN mechanism/authority;
- storage locations never become Artifact/Rendition identity;
- no M-Files/Veeva/MasterControl workflow taxonomy is copied into domain authority.

## 19.7 Essential vs accidental complexity

Essential:

- exact Submission-bound governance;
- sequential Steps;
- actor resolution/snapshot;
- ANY/ALL;
- strict SoD;
- bounded fresh-auth;
- detached feedback;
- immutable Rendition;
- automatic Release;
- concrete obligations;
- explicit acknowledgement;
- in-product inspection contract.

Removed/deferred accidental complexity:

- REVIEW/APPROVAL Step type;
- separate route application modes;
- generic BPM/task capability engine;
- M-of-N / DAG routing;
- generic delegation;
- auto-escalation;
- provider-specific editor authority;
- universal mandatory PDF;
- duplicate release-generation identity;
- dynamic historical audience;
- read/view-as-ack;
- mandatory eSignature platform.

## 19.8 YAGNI / future-cost check

**PASS.**

Known invariants are preserved; unsupported future capabilities are seams/reopen triggers. The candidate does not construct generic workflow, annotation, identity, rendition or audience platforms.

---

# 20. Adversarial challenge

## F1 — policy changes after SUBMIT alter approval/release

Old candidate could let floating current config change a Submission's human or rendition gate.

**Closed:** immutable `SubmissionApprovalRequirement` + representation snapshot in `ReleasePlan`.

## F2 — NoHumanApproval has no historical proof

Without a snapshot, Release could not distinguish deliberate no-human governance from missing Approval state.

**Closed:** immutable per-Submission Approval requirement; no fake approver/instance.

## F3 — fresh-auth boolean loses evidence

B2 requires consumer-owned snapshot of consumed evidence.

**Closed:** bounded fresh-auth evidence snapshot in ApprovalStepDecision; no provider secret/token.

## F4 — RETURN_FOR_CHANGES has no required reason

Frozen R9.5 requires one.

**Closed:** DB/app law requires nonblank decision_reason.

## F5 — same User ACCEPTs two Steps through a bypass path

Application-only SoD is insufficient.

**Closed:** partial UNIQUE on ACCEPT actor per ApprovalInstance + cross-row submitter/creator DB guard.

## F6 — REVIEW/APPROVAL Step type reintroduces legacy provider coupling

**Closed:** one Step semantic; collaboration and reauth are orthogonal.

## F7 — Group disappears while old policy/instance needs it

Permanent history FK would block deletion forever; no live FK would break active actor resolution.

**Closed:** current Policy + PENDING/ACTIVE Step typed live refs; historical snapshots only after dependency ends.

## F8 — participant universe freezes too early

Freezing future-step members at SUBMIT makes long-running routes stale and overfits legacy behavior.

**Closed:** actor rule resolves when each Step activates; active assignments then freeze. Explicit reassignment can deliberately use a newly current qualified actor under the same rule.

## F9 — reassignment becomes arbitrary delegation

**Closed:** replacement must satisfy exact frozen actor rule + live AuthZ + SoD; NamedUser cannot silently become another person.

## F10 — feedback mutates submitted bytes

**Closed:** immutable SubmissionFeedback overlay; returned DRAFT uses B3 OCC for application.

## F11 — feedback appears after actor already decided

That creates confusing post-decision rationale.

**Closed:** actor cannot append Step feedback after recording their decision.

## F12 — viewer PDF silently becomes official

**Closed:** ReleasePlan representation snapshot alone decides requirement; ReleaseRecord pins official rendition only when required.

## F13 — renderer succeeds but Artifact is confirmed without owner

**Closed:** Artifact + Rendition semantic confirmation is one local transaction.

## F14 — Distribution audience changes during Release

**Closed:** audience config mutation and Release share Document serialization root.

## F15 — async worker resolves denominator after membership drift

**Closed:** obligations are inserted inside winning Release transaction; delivery remains async only.

## F16 — Acknowledgement duplicates user authority or grants access

**Closed:** Acknowledgement stores only obligation identity/time; actor derives from obligation and must still pass document.read_effective. Distribution never grants access.

## F17 — reauth only on ACCEPT despite Step-level requirement

That would create an undocumented outcome-specific policy.

**Closed:** if Step requires reauth, every governance decision on it consumes/snapshots fresh-auth. Outcome-specific ceremony is a future explicit decision.

---

# 21. Proof obligations

| Claim | Falsification/proof obligation |
|---|---|
| one Step semantic | target schema/API has no `purpose/stage_kind=review|approval` |
| policy pinning | config/policy edits cannot change existing SubmissionApprovalRequirement |
| representation pinning | DocumentType policy edits cannot change existing ReleasePlan gate |
| NoHuman truth | no-human Submission releases only because its immutable requirement says so |
| exact candidate | decision/feedback/rendition/release all resolve same Submission |
| Submission immutable | feedback/suggestion cannot change source bytes/digest |
| feedback application | returned suggestion changes only WorkingContent through OCC and later new Submission |
| fresh-auth proof | required decision cannot commit without bounded evidence snapshot; non-required decision cannot smuggle provider evidence |
| return reason | blank/missing RETURN_FOR_CHANGES reason rejected |
| submitter SoD | creator/submitter ACCEPT rejected even through direct DB-serving write path |
| multi-step SoD | same User second ACCEPT in same instance hits partial unique |
| ANY | one valid ACCEPT completes exactly one Step |
| ALL | all active assignments required; empty set never succeeds |
| step activation | actor membership/role changes before activation affect resolution; after activation do not silently rewrite denominator |
| live AuthZ | active assignment does not confer permission after offboarding/grant loss |
| Group deletion | live policy/in-flight group dependency blocks delete; history-only does not |
| reassign | replacement outside original assignment set allowed only if it currently satisfies same actor rule + AuthZ + SoD |
| viewer coherence | approval workspace cannot show PDF from another Submission |
| source-only viewer | auxiliary PDF absence cannot block SourceOnly Release |
| required rendition | RequireRendition cannot release without exact required Rendition |
| Artifact ownership | confirmed rendition output cannot commit without Rendition owner relation |
| one Release | concurrent retries create one ReleaseRecord/effectivity winner |
| one EFFECTIVE | B3 partial unique survives Release races |
| audience config race | config add/remove racing Release is fully before or after snapshot |
| membership race | Group membership race is fully before or after concrete Release denominator cut |
| historical denominator | later Group changes cannot rewrite obligations |
| acknowledgement | view/download/notification cannot create acknowledgement; non-obligated/wrong User cannot acknowledge |
| provider independence | editor/viewer/renderer/storage provider identity never enters business identity or permission catalog |

Architecture proof now = authority reconstruction + Method analysis + external comparison + counterexample challenge. Implementation proof later = PostgreSQL negative/constraint/concurrency tests, privilege tests, exact-Submission E2E, provider adapter fidelity tests, restart/retry tests and contract parity.

---

# 22. Decision — corrected candidate

Proposed Method outcome:

> **RESTRUCTURE NOW within B4 target design:** preserve the small specialized sequential Approval boundary, but remove the legacy-derived `REVIEW|APPROVAL` Step taxonomy. Use one governance Step over exact Submission, per-Submission immutable governance/release requirements, activation-time concrete actor snapshots, detached SubmissionFeedback, immutable Rendition, automatic Release and same-commit concrete Distribution obligations.

Candidate family set:

```text
Approval
  DocumentTypeApprovalConfig
  ApprovalPolicy
  ApprovalPolicyVersion
  ApprovalPolicyStep
  ApprovalPolicyLiveGroupRef
  SubmissionApprovalRequirement
  ApprovalInstance
  ApprovalStepInstance
  ApprovalStepLiveGroupRef
  ApprovalParticipant
  ApprovalStepDecision
  SubmissionFeedback

Controlled Information
  DocumentTypeRepresentationPolicy
  ReleasePlan
  Rendition
  ReleaseRecord

Distribution
  DistributionAudienceGroup
  DistributionObligation
  AcknowledgementRecord
```

This candidate is **not independently ratified** and is not yet accepted for R10 integration until operator adjudication.

If accepted:

```text
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R9.5 Step-purpose wording = bounded working refinement carried forward explicitly
implementation = BLOCKED
next = R10-B5 integrated candidate + B3↔B4↔B5 coherence challenge
```

---

# 23. Reopen triggers

Reopen only on material evidence:

- a genuinely non-gating comment-only task inside the same Approval route;
- legally distinct decision meaning requiring different persistence/enforcement, not merely a label;
- task that legitimately mutates submitted candidate without return-to-DRAFT (also reopens B3 immutability);
- real parallel/DAG route requirement;
- M-of-N requirement not representable by ANY/ALL;
- actor-rule semantics requiring live dynamic membership after Step activation;
- generic delegation requirement beyond qualified reassignment;
- outcome-specific fresh-auth policy requirement;
- successful Rendition supersession/re-render requirement;
- additional Distribution audience kinds with real consumers;
- acknowledgement requiring regulated eSignature/fresh-auth;
- provider-independent collaboration semantics that truly differ by Step and cannot use the common surface;
- implementation evidence that B1/B2 lock laws cannot preserve these transactions without a materially different serialization mechanism.

Legacy UI shape, current editor modes, current schema and vendor workflow taxonomy are not reopen triggers.

---

# 24. Whole-R10 posture

B4 remains **NON-AUTHORITATIVE CORRECTED CANDIDATE** pending operator acceptance.

The bounded `Step.purpose` refinement must be explicit in current routing once accepted; the old frozen ledger wording must not silently coexist as equal working authority. Final R10 ratification still requires Whole-R10 Global Coherence Review + cold independent review + operator adjudication under the accepted R10 working mode.
