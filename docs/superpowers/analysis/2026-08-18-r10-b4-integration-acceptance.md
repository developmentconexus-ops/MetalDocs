# R10-B4 — Integration Acceptance

> **Status:** OPERATOR-ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

This artifact records operator adjudication of the self-reviewed corrected B4 candidate:

`docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md`

It does not independently ratify R10-B4. Under the accepted R10 working mode, B4 is now a non-final integration input that remains challengeable by B5–F, Whole-R10 Global Coherence Review, and the final cold independent review.

## Accepted B4 target

B4 is accepted as the current Global Maximum working target:

```text
exact immutable RevisionSubmission
→ one sequential Approval/Governance Step model
→ detached SubmissionFeedback
→ immutable Rendition when produced
→ automatic system-owned Release
→ concrete Distribution obligations in the winning Release transaction
→ explicit immutable AcknowledgementRecord
```

Key accepted laws:

- `RevisionSubmission.id` is the single exact-candidate identity for Approval, Rendition and Release; no duplicate release-generation authority.
- ApprovalPolicy is versioned; each Submission snapshots its Approval requirement and exact PolicyVersion.
- `ApprovalPolicyStep` has one semantic type. There is no `purpose=review|approval`, `stage_kind`, `review_mode` or `approval_mode` in the target kernel.
- Step labels such as “Revisão técnica”, “Qualidade” or “Diretoria” are business language only.
- Every active Step participant may inspect the exact Submission, use the in-product viewer/PDF rendition where available, comment, annotate, suggest corrections, and decide `ACCEPT | RETURN_FOR_CHANGES`.
- Fresh-auth is an orthogonal Step requirement. Required bounded `FreshAuthEvidence` is consumed and snapshotted in the immutable decision authority; Approval stores no password/provider token.
- `RETURN_FOR_CHANGES` requires a nonblank reason, returns the same Revision to DRAFT, never mutates the old Submission, and never resets B3 `working_version`.
- Strict SoD remains: creator/submitter cannot ACCEPT own Submission; the same User cannot ACCEPT two Steps of one ApprovalInstance; no role bypasses SoD.
- Step actor rule `NamedUser | Group | RoleInArea` resolves to concrete Users at Step activation; action-time Authorization remains live.
- Reassignment must currently satisfy the same frozen actor rule, User eligibility, `approval.act` and SoD; it is not generic delegation.
- Live Group FKs exist only while current policy or pending/active Step semantics still need that Group; historical evidence does not make Group permanently undeletable.
- `SubmissionFeedback` is detached immutable feedback on the exact Submission; applying a suggestion after return-to-DRAFT is a normal B3 WorkingContent OCC mutation.
- In-product viewing is a product requirement for supported governed formats; download/external-app is not the normal supported inspection journey.
- `OfficialRepresentationPolicy` is distinct from viewer capability. `SourceOnly` may still have an auxiliary PDF for viewing; only `RequireRendition(format)` blocks Release on that rendition.
- A successful Rendition confirms its output Artifact and establishes the Rendition as the Artifact's typed semantic owner in one local transaction.
- Release is automatic/system-owned. No publish button.
- Each Submission snapshots representation requirement and optional `not_before`; later configuration edits do not alter an in-flight Release gate.
- Winning Release is the only effectivity transition: it atomically creates ReleaseRecord, makes candidate EFFECTIVE, supersedes prior EFFECTIVE Revision, and freezes concrete Distribution obligations.
- Distribution audience configuration and Release serialize on the same Document root; Group membership changes are fully before or after the release-time denominator cut.
- Acknowledgement is an explicit authenticated action by the obligated User; view/download/notification/search do not acknowledge and Distribution does not grant access.

## Bounded R9.5 refinement — operator accepted

The previously frozen Approval Step discriminator:

```text
Step.purpose = review | approval
```

is explicitly reopened and removed for the current R10 working target.

Replacement:

```text
ApprovalPolicy(version)
  ordered Steps

Step
  label
  actor_rule: NamedUser | Group | RoleInArea
  completion: ANY | ALL
  requires_reauthentication
  due_in_days?
```

Rationale under DevelopmentConexus Engineering Method v1.0.0:

- new evidence showed the old distinction was primarily a legacy/editor interaction taxonomy rather than a necessary invariant;
- collaboration, verdict and stronger authentication are orthogonal capabilities in comparable mature systems rather than a universal `review|approval` domain taxonomy;
- the old discriminator created structural inversion by allowing UI/provider mode to define domain Step identity;
- removing it preserves all real governance invariants while reducing accidental complexity and ambiguity with Controlled Information `PeriodicReview`.

This bounded refinement does not reopen exact-Submission binding, actor rules, ANY/ALL, SoD, fresh-auth, return/resubmit, Rendition, Release, Distribution, B1, B2 or B3.

## Self-review result before acceptance

The integrated Method/coherence review returned **PASS after bounded corrections**. Material findings closed before operator acceptance included:

1. per-Submission Approval/representation policy drift;
2. insufficient fresh-auth evidence persistence;
3. missing mandatory reason for `RETURN_FOR_CHANGES`;
4. insufficient structural SoD backstops;
5. prematurely frozen all-Step actor pools / incomplete reassignment semantics;
6. historical Group FK conflict with B2 hard-delete law;
7. ambiguous `ReviewFeedback` vocabulary, replaced by `SubmissionFeedback`;
8. confirmed Rendition Artifact ownership seam;
9. Distribution audience mutation vs Release race;
10. duplicated User identity on AcknowledgementRecord.

No remaining material contradiction was found across Method, R9.5 (with the bounded refinement above), R10-A, B1, B2 and accepted B3.

## Residual proof obligations

These remain architecture/implementation proof obligations, not unresolved product ambiguity:

- whole B1→B4 lock ordering must be shown cycle-free/linearizable under `READ COMMITTED`;
- DB guards/privileges must prove immutable evidence and SoD controls fire on all admitted serving write paths;
- exact viewer/rendition/API/provider adapters are designed in R10-E;
- async workers/timers/notification/search effects are designed in R10-D;
- Audit cross-owner same-commit composition is finalized in B6.

If implementation evidence invalidates these assumptions, reopen only the implicated decision before code.

## Next

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B5 = NEXT / DESIGN ONLY
implementation = BLOCKED
```

B5 must explicitly challenge B3↔B4↔B5 coherence, especially Artifact ownership/disposition closure, DocumentRevision retention units containing Submission/Approval/Rendition/Release evidence, LegalHold, Evidence/Dossier relations and disposition atomicity.
