# MetalDocs Launch V1 — Product Contract

> **Status:** NON-AUTHORITATIVE PRODUCT CONTRACT CANDIDATE — OPERATOR WRITTEN REVIEW PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This contract defines **what Launch V1 must do and mean before technical architecture resumes**. It intentionally contains no SQL, table layout, Go package design, provider-key design or storage topology.

The current R10-C candidate is paused while this contract is reviewed. Earlier R9.5/R10 decisions remain evidence; they cannot force a Launch capability or abstraction that this contract does not justify.

---

## 1. North Star

> **MetalDocs is the company system for creating, governing, approving, publishing, finding and proving the history of official controlled documents.**

Launch V1 is not a generic ECM/file drive, BPM engine, QMS suite, LMS, records-management platform, eDiscovery system, PLM, ERP or generic integration platform.

The product must answer unambiguously:

```text
What is the official document?
Which revision is valid now?
What is being changed?
What exact content was submitted and governed?
Who acted and when?
What changed between revision cycles?
How does a reader find the current valid version without draft noise?
How does an effective document become obsolete?
How is history preserved without being rewritten?
```

---

## 2. Reference-platform test

Reference products are falsification evidence, not feature checklists.

- **SharePoint:** draft/pending content can be separated from approved/published versions; readers can continue seeing the prior approved version while changes are pending.  
  https://support.microsoft.com/en-US/SharePoint/lists/documents-and-library/how-versioning-works-in-lists-and-libraries
- **M-Files:** objects have independent version history; relationships are references rather than copies; ordinary objects can serve as templates.  
  https://userguide.m-files.com/user-guide/latest/eng/object_relationships.html  
  https://userguide.m-files.com/user-guide/latest/eng/using_template.html
- **Qualio:** focused controlled-document behavior includes draft, review, approval, make-effective, retire and audit, while periodic review/training are additional capability layers.  
  https://docs.qualio.com/en/articles/6526420-user-permissions  
  https://docs.qualio.com/en/articles/11122-audit-trail-overview
- **Veeva QualityDocs:** demonstrates mature change control, obsolescence, periodic review and Read & Understood/Training, but also provides the upper-bound warning against turning Launch into a configurable quality platform.  
  https://quality.veevavault.help/en/lr/15349/  
  https://quality.veevavault.help/en/lr/37406/  
  https://quality.veevavault.help/en/lr/72024/

Product conclusion: **draft/effective separation, immutable governed attempts, explicit effectivity and explicit obsolescence are essential; generic change-control/training/records platforms are not Launch prerequisites.**

---

## 3. Personas

- **Governance Admin:** configures users/areas/groups, document types, numbering, templates, access and the governance policy used by each document type.
- **Author / Document Owner:** creates, edits, submits, receives return feedback, changes and resubmits.
- **Reviewer / Approver:** evaluates the exact submitted candidate, collaborates where allowed and records a governance decision.
- **Reader:** finds and reads the current effective revision by default.
- **Auditor / Governance Viewer:** reconstructs lifecycle and action history without Audit becoming current-state authority.

Launch has no generic workflow designer, low-code rule engine or custom-form platform.

---

## 4. Core product concepts

### Controlled Document

Stable official identity across its lifetime, for example `PO-001 — Procedimento de Compras`. A Document is not a file and is not replaced when its content changes.

### Business Revision

One governed change cycle: `REV001`, `REV002`, ... Revision numbers never reuse.

Lifecycle:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

At most one revision is EFFECTIVE. A newer DRAFT/SUBMITTED revision may coexist with the prior EFFECTIVE revision.

### Working Content

Mutable work for the open DRAFT revision. Autosave/checkpoints preserve work but never consume a business revision number or create official history by themselves.

### Submission

Immutable exact governed attempt created by **Submit**. It freezes exact content plus all decision-relevant governed state/provenance. Same-Revision return/resubmit creates another immutable Submission.

### Governance Policy / Route

A Document Type chooses either:

```text
NoHumanApproval
or
UseGovernanceRoute
```

A governance route is one sequential Step concept with product labels such as `Revisão técnica`, `Gestor`, `Qualidade`. Launch does not create separate ReviewWorkflow and ApprovalWorkflow engines.

Normal participant outcomes:

```text
ACCEPT
RETURN_FOR_CHANGES
```

### Release / Effectivity

When the Submission satisfies every required gate, the system establishes effectivity. There is no separate user operation meaning “publish whichever file is latest”.

### Rendition / Official Representation

The submitted source remains meaningful. A Document Type may be source-only or require one derived official representation such as PDF. A representation cannot silently change the governed content.

### Template

A Template is an ordinary governed Document used to seed a new independent Document. There is no parallel TemplateVersion lifecycle.

---

## 5. End-to-end Launch journeys

### A. Create

```text
Choose Document Type
→ allocate stable code
→ create REV001 DRAFT
→ edit
```

Creation may start blank or from the exact current EFFECTIVE eligible Template. Later template changes never rewrite the derived Document.

### B. Draft/autosave

The author edits Working Content. Autosave is recoverable technical working history, not a new REV or official governed version.

### C. Submit

```text
DRAFT
→ validate submit requirements
→ freeze Submission S1
→ SUBMITTED
```

The submitted candidate is immutable.

### D. Review / approval

Each active participant sees the exact Submission. Feedback never mutates it. Participant chooses `ACCEPT` or `RETURN_FOR_CHANGES` according to the configured route.

For `NoHumanApproval`, no fake System approver is created; the human gate is simply absent.

### E. Return / resubmit

```text
S1 remains immutable
Revision → DRAFT
→ author changes Working Content
→ Submit
→ S2
```

S1, its feedback and its decisions remain history.

### F. Withdraw a Submission

An authorized author may withdraw an active governance attempt before release when the intent is to continue editing the same business Revision.

```text
SUBMITTED
→ terminate current governance attempt as WITHDRAWN
→ same Revision returns to DRAFT
```

Withdraw is **not** Revision cancellation and never fabricates a reject decision.

### G. Cancel an open Revision

When the business change cycle itself must stop permanently, an authorized actor cancels that open Revision with an explicit reason.

```text
DRAFT or eligible SUBMITTED open Revision
→ CANCELLED
```

If an older revision is EFFECTIVE, it remains EFFECTIVE. Cancellation does not delete historical identity or rewrite prior Submissions.

### H. Release / first effective version

When all required gates pass:

```text
REV001 SUBMITTED
→ system Release
→ REV001 EFFECTIVE
```

Readers can now discover it as official content.

### I. Revise an effective Document

```text
REV003 EFFECTIVE
→ REV004 DRAFT
```

Readers continue seeing REV003 while REV004 is drafted/governed. Successful Release of REV004 changes:

```text
REV003 → SUPERSEDED
REV004 → EFFECTIVE
```

### J. Governed obsolescence without replacement

MetalDocs must support intentionally withdrawing an EFFECTIVE Document from use without a successor.

Rules:

1. only the current EFFECTIVE revision can be the target;
2. reason is mandatory;
3. obsolescence is governed, never a raw status toggle;
4. until governance completes, the existing revision remains EFFECTIVE;
5. success changes it to OBSOLETE and leaves no EFFECTIVE revision;
6. ordinary current-document search stops presenting it as active;
7. authorized history remains available;
8. an existing open replacement Revision must be cancelled/withdrawn/resolved before obsolescence can complete;
9. reactivation of an OBSOLETE Document is not Launch scope.

Exact reuse/configuration of the governance engine for obsolescence is technical design work after this contract.

### K. Search / read / download

Normal discovery favors current EFFECTIVE documents. Core filters include code/title, Document Type, Area, responsible owner and status. Draft/submitted work appears in author/governance workspaces, not ordinary reader results as equivalent truth.

Opening a controlled Document exposes its stable identity, current status/revision, effective/release date, source or official representation and authorized revision history.

### L. Audit / history

Domain history owns Submissions, decisions, Releases, cancellations and obsolescence. Audit proves meaningful actions/timeline but is never queried to derive current lifecycle state.

### M. User offboarding

Offboarding stops future access/actions while preserving truthful historical attribution/evidence. Erasable profile enrichment can disappear without rewriting immutable governance facts.

### N. Historical migration

Migration is a go-live/cutover concern, not a generic integration platform. Preserve reliable source code/revision/state/provenance; unknown stays unknown; never fabricate native MetalDocs approvals, releases, actors or historical timestamps; never replay historical side effects as current events.

---

## 6. Scope tiers

### Launch Core

```text
single-company deployment
users / areas / groups
roles / scoped access
Document Types
numbering
Controlled Document + business Revision
DRAFT Working Content + autosave/concurrency
Templates
immutable Submission
NoHumanApproval or sequential governance route
feedback / return / withdraw / resubmit
Revision cancellation
Approval evidence
Release / EFFECTIVE / SUPERSEDED
optional required Rendition
explicit governed OBSOLETE flow
revision/history view
search/filter current effective content
source/official representation read/download
Audit trail
historical migration/cutover when required
backup/restore correctness
```

### Launch+ — recommended next, not first-use prerequisites

```text
Distribution / Read & Acknowledge
Periodic Review
```

`Read & Acknowledge` is not Training/LMS and must not imply competence, qualification, curricula or quizzes.

### Future

```text
Dossier / documentary case context
Evidence capture / quality records
Retention policies
Legal Hold
Governed disposition/destruction
Records-driven WORM/Object Lock
eDiscovery/custodian preservation
Governed Subject Export package
Generic External Repository IMPORT/PUBLISH connectors
Training/LMS
Generic/multi-document Change Control platform
pooled multi-customer tenancy
realtime coauthoring/CRDT
```

Future capabilities create no dormant Launch module/table/permission/job.

---

## 7. Whole-product adjudication of prior decisions

### KEEP

- Document ≠ Revision ≠ Working Draft ≠ Submission.
- Revision ordinals never reuse.
- One current EFFECTIVE revision; newer DRAFT/SUBMITTED may coexist.
- Immutable Submission; return/resubmit creates another Submission.
- One sequential governance Step model rather than Review/Approval engines.
- `NoHumanApproval` is explicit and never creates fake approver evidence.
- Automatic/system-owned Release is effectivity authority.
- Template is an ordinary governed Document role.
- Source remains primary; optional required Rendition is representation policy, not universal PDF.
- Search is discovery only and never grants access.
- Audit is timeline evidence, never lifecycle authority.
- Authentication provider may own credentials; MetalDocs owns product identity/authorization.
- Historical migration cannot fabricate native governance history.
- One company per Launch deployment.
- Confirmed governed history is not physically disposed in Launch.

### BOUNDED REOPEN / RESTRUCTURE

- **Standalone Artifact semantic owner:** remove from Launch target. Exact-content facts belong to the semantic record that freezes them; shared byte storage is mechanism only.
- **R10-C:** paused and must be redesigned after Product Contract acceptance.
- **Obsolescence:** becomes an explicit governed product journey.
- **B5 Launch scope:** Dossier/Evidence leave Launch Core unless a named rollout consumer is produced before final ratification.
- **B6 Governed Subject Export:** Future unless a concrete auditor/customer/portability obligation is named.
- **Generic External Repository import/publish:** Future; Historical Migration remains a distinct go-live concern.

### DEFER SAFELY

```text
Distribution/Acknowledgement → Launch+ recommended
Periodic Review              → Launch+
Dossier                      → Future
Evidence                     → Future
Retention/Hold/Disposition   → Future
Governed Export package      → Future
External repository sync     → Future
Training/LMS                 → Future
```

Reopen only on a named consumer, requirement or reachable production failure mode.

---

## 8. Product invariants

1. A Document has one stable official identity across revisions.
2. A business Revision is not an autosave/checkpoint.
3. Revision ordinals never reuse.
4. Ordinary readers see current effective truth, not moving drafts.
5. A Submission is immutable exact governed-candidate identity.
6. Governance/feedback always binds an exact Submission.
7. Return/withdraw never mutates prior Submission history.
8. Cancelling a Revision ends that business change cycle without disturbing an older EFFECTIVE revision.
9. Release is the only normal transition establishing a new EFFECTIVE revision.
10. Replacement Release supersedes the prior EFFECTIVE revision as one product transition.
11. Obsolescence without replacement is explicit, justified and governed.
12. EFFECTIVE/SUPERSEDED/OBSOLETE/CANCELLED are lifecycle facts, never physical delete commands.
13. Templates create independent Documents.
14. Search never grants access and never presents drafts as official to ordinary readers.
15. Audit proves actions but never becomes current-state authority.
16. Imported history never becomes fake native history.
17. Storage/provider identity never becomes Document/Revision/Submission identity.
18. Launch preserves governed history and has no governed physical disposition.
19. Later capabilities must attach without duplicating Document/Revision/Submission authority.

---

## 9. Whole-product scenario gate

Any renewed technical architecture must answer all scenarios without contradictory authorities:

| Scenario | Required truth |
|---|---|
| Create blank | stable Document + REV001 DRAFT, not official |
| Create template-based | current EFFECTIVE template seeds independent Document |
| Autosave | preserve work without consuming REV/official history |
| Submit | freeze exact candidate |
| Return | old Submission stays immutable; same REV returns DRAFT |
| Withdraw | terminate governance attempt, same REV returns DRAFT |
| Cancel Revision | end business change cycle; older EFFECTIVE remains |
| Resubmit | new Submission, same business REV |
| Governance | actor judges exact Submission |
| Release | establish one EFFECTIVE truth |
| New Revision | prior EFFECTIVE stays reader truth until successor Release |
| Obsolete | governed withdrawal without successor |
| Search | ordinary discovery favors current EFFECTIVE |
| View/download | content corresponds to intended governed version |
| Offboarding | access ends; history remains truthful |
| Migration | preserve source truth, fabricate nothing |
| Restore | never serve inconsistent governed content/history |
| Provider outage | business state remains truthful while bytes may be unavailable |
| Audit | reconstruct actions without deriving state from log events |

A proposed module/entity with no justification from these scenarios or another explicit invariant is presumptively YAGNI.

---

## 10. Launch success criterion

Launch is product-complete when a real company can:

```text
configure controlled-document governance
→ create from blank/template
→ author safely
→ submit exact content
→ govern it or return it
→ make it effective
→ let readers reliably find/read the current effective revision
→ revise it while the old revision stays effective
→ withdraw a Submission or cancel a change cycle correctly
→ obsolete an effective Document without replacement
→ inspect trustworthy lifecycle/audit history
→ migrate/restore without fabricating documentary truth
```

without Dossier, Evidence, Retention, Legal Hold, generic Change Control, governed export packages, repository sync or LMS machinery.

---

## 11. Promotion gate

This file is not yet authority.

```text
operator reviews this written Product Contract
→ bounded corrections if needed
→ operator accepts Product Contract
→ promote durable wiki/architecture Product Contract
→ Whole-Product Global Coherence Review against prior R10
→ rebuild ownership/technical stages only from accepted product semantics
```

Do not resume R10-C or author an implementation plan before this gate closes.
