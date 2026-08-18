# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + R10-A/B1/B2 PROMOTED + R10-B3/B4 ACCEPTED FOR R10 INTEGRATION / NON-FINAL + R10-B5 NEXT**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file — current checkpoint / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program/global-coherence authority
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority
6. `wiki/architecture/r10-technical-architecture.md` — promoted R10 authority through integrated B2
7. `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md` — accepted B3 working target
8. `docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md` — accepted corrected B4 working target
9. `docs/superpowers/analysis/2026-08-18-r10-b4-integration-acceptance.md` — operator adjudication / bounded refinement record
10. review artifacts only when auditing how an already-promoted or accepted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only.

---

## Current checkpoint

```text
R3–R9   = LOCKED
R9.5    = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED

R10-A   = CLOSED / APPROVED / GCR-REFINED / SINGLE-COMPANY-REFINED
R10-B   = IN PROGRESS / DESIGN ONLY
R10-B1  = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2  = CLOSED / APPROVED / INTEGRATED
R10-B3  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B4  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B5  = NEXT / DESIGN ONLY
R10-B6  = NOT STARTED
R10-C   = NOT STARTED
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

Promoted authority remains:

- R3–R9.5: frozen product/domain ledger;
- R10-A/B1/B2: `wiki/architecture/r10-technical-architecture.md` + program authority.

Accepted non-final integration inputs:

- B3 candidate above;
- B4 corrected candidate above;
- B4 operator acceptance record above.

B3/B4 remain challengeable by material counterexamples from B5–F. Whole-R10 Global Coherence Review + cold independent Fable review happens before final R10 ratification unless a truly exceptional material blocker requires earlier independent review.

---

## R10 working-mode rule

From B3 onward:

```text
repo authority
→ evidence + directed external research where useful
→ DevelopmentConexus Method
→ integrated candidate
→ internal adversarial challenge
→ cross-stage coherence
→ operator understanding/adjudication
→ ACCEPTED FOR R10 INTEGRATION / NON-FINAL
```

Do not return to per-table/per-field/microdecision review gates. Later blocks reopen only the materially implicated earlier decision when a real counterexample appears.

Implementation remains blocked until R10 is integrated, Whole-R10 coherence + independent review are closed, operator ratifies the target, and an implementation plan is authored from that target.

---

## R10-B3 accepted working outcome

Core:

```text
Document
→ business DocumentRevision
→ one mutable WorkingContent + working_version OCC
→ immutable RevisionSubmission + deterministic governed manifest/digest
→ exact provider-neutral Artifact
```

Key laws carried forward:

- downstream governance binds exact immutable `RevisionSubmission`, never floating current content;
- same-REV return/resubmit never mutates old Submission;
- SUBMIT consumes `working_version` generation so stale pre-submit writes cannot reappear after return-to-DRAFT;
- Artifact is exact-byte identity, not storage location;
- one-open/one-EFFECTIVE have structural backstops;
- template creation and PeriodicReview must serialize against B4 Release on the same Document root;
- B5 must finish global Artifact typed-owner/disposition closure.

---

## R10-B4 accepted working outcome

Current working target:

```text
RevisionSubmission
→ one sequential Approval/Governance Step model
→ detached SubmissionFeedback
→ immutable Rendition when produced
→ automatic system-owned Release
→ concrete Distribution obligations in winning Release transaction
→ explicit AcknowledgementRecord
```

Key laws:

- no `review|approval` Step type in the current R10 working target;
- every Step = one governance gate with `label`, `NamedUser|Group|RoleInArea`, `ANY|ALL`, optional fresh-auth, optional due date;
- active participants may view the exact Submission in-product, including exact PDF rendition where available, and may comment/annotate/suggest;
- feedback never mutates `RevisionSubmission`; applying returned suggestions is a later B3 WorkingContent OCC mutation;
- Approval requirement + exact PolicyVersion are snapshotted per Submission;
- representation requirement + `not_before` are snapshotted per Submission;
- strict SoD has application + structural backstops;
- required fresh-auth evidence is consumed and snapshotted into immutable decision evidence;
- `RETURN_FOR_CHANGES` requires reason and returns same Revision to DRAFT without resetting B3 generation;
- Step actor rule resolves to concrete Users at Step activation; action-time AuthZ stays live;
- qualified reassignment must satisfy the same frozen actor rule + User eligibility + `approval.act` + SoD;
- live Group FKs exist only while current/pending/active semantics need them; history does not make Group permanently undeletable;
- `OfficialRepresentationPolicy` != viewer capability; `SourceOnly` may still have auxiliary PDF viewer rendition;
- Rendition owns one exact output Artifact as immutable semantic success;
- Release is automatic/system-owned and is the only effectivity transition;
- winning Release atomically creates ReleaseRecord, flips EFFECTIVE/SUPERSEDED and freezes concrete Distribution obligations;
- Distribution never grants access; explicit acknowledgement by the obligated User is the only completion signal.

### Explicit bounded R9.5 refinement

Historical frozen wording contained:

```text
Step.purpose = review | approval
```

Operator accepted a bounded reopen on 2026-08-18. Current R10 working target removes this discriminator and replaces it with one governance Step semantic model. This is an explicit overlay for current R10 integration, not a silent rewrite of the frozen ledger.

Reason: new evidence + Structural Inversion Test showed the old discriminator encoded legacy editor/UI ceremony rather than an invariant. Collaboration and fresh-auth are orthogonal capabilities; the split also created unnecessary ambiguity with CI `PeriodicReview`.

No other Approval semantics were reopened by this correction.

---

## Exact next step — R10-B5 Documentary Context + Records Governance + Artifact closure

Open **R10-B5** in integrated batch/design-only mode from promoted B1/B2 plus accepted non-final B3/B4.

First deliverable: one integrated intake/decomposition and candidate covering at minimum:

```text
Documentary Context
  DossierType / Dossier
  Dossier↔Document contextual relation
  EvidenceType / Evidence
  primary Dossier + secondary links
  exact Artifact relation / capture immutability
  ExternalReference / imported provenance where owned here

Records Governance
  RetentionPolicy selection / RetentionBinding snapshot
  DocumentRevision retention unit
  Evidence retention unit
  RetentionExtension
  LegalHold scopes + materialized held subjects
  disposition eligibility / explicit DispositionRecord
  no automatic delete

Artifact closure
  typed owner/reference closure across DocumentRevision/Submission/Rendition/Evidence
  preservation when one Artifact participates in retained/held subjects
  no generic owner_type/id registry
  no confirmed orphan semantic Artifact

Cross-stage coherence
  B3 Submission/Artifact ownership
  B4 Approval/Rendition/Release evidence inside DocumentRevision retention unit
  Release/effectivity anchors for retention clocks
  Distribution/Acknowledgement interaction with retention where material
```

Separate clearly:

```text
Audit/Interchange/final cross-owner matrix               → B6
physical store/malware/restore                           → R10-C
async jobs/projections/notifications/provider effects    → R10-D
API/frontend/viewer/editor journeys                      → R10-E
historical migration/cutover/deletion                    → R10-F
```

Do not reopen B3/B4 for implementation convenience. Reopen only on a material B5 counterexample.

Current implementation is evidence only. Implementation remains **BLOCKED**.
