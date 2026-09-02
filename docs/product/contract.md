# MetalDocs Launch V1 — Product Contract

> **Status:** ACTIVE / OPERATOR-APPROVED PRODUCT AUTHORITY  
> **Accepted:** 2026-08-18  
> **Product Contract revision:** **REV001** — 2026-08-18 — post-T5 independent-review bounded amendments  
> **Revision-numbering amendment:** 2026-08-18 — **REV000 is initial issuance; REV001 is the first revision**  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This contract defines **what Launch V1 must do and mean before technical architecture resumes**. It intentionally contains no SQL, table layout, Go package design, provider-key design or storage topology.

The operator accepted the written Product Contract on 2026-08-18, later clarified the business revision convention on the same date, and ratified two bounded product-completeness amendments after the post-T5 independent review: **document title is governed Revision metadata**, and an **active human-governed obsolescence request may be withdrawn before completion without changing the EFFECTIVE document**. This page is the durable product authority.

Earlier R9.5/R10 decisions remain evidence. They cannot force a Launch capability or abstraction that this contract does not justify.

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

The stable Document identity includes its stable code/identity. **The human-readable title is governed Revision metadata**, so a title change belongs to a business Revision rather than silently mutating the current official reader truth. While a newer Revision is DRAFT/SUBMITTED, ordinary readers continue seeing the title of the current EFFECTIVE Revision.

### Business Revision

One governed issuance/change cycle. Numbering starts at zero:

```text
REV000 = initial issuance
REV001 = first revision after the initial issuance
REV002 = second revision
...
```

Revision numbers never reuse. The ordinal therefore communicates revision count: `REV000` has not yet been revised; `REV001` has undergone one revision after initial issuance.

A Revision owns the governed human-readable title for that issuance/change cycle together with the content/state it presents as official when EFFECTIVE.

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

### Confidentiality Class

A Confidentiality Class is a Company-configured label answering "who may know this Document
exists and read it", carried by every Document — exactly one, always, with one distinguished
default class meaning "no restriction beyond the Area and Company grants".

Classes are **additive and non-hierarchical**: none implies, dominates or inherits another.
Reading a Document requires the ordinary read permission in its scope **and** a clearance for
its class; the two conditions are conjunctive. Clearances are granted centrally to Users or
Groups, never chosen document by document — there is no per-document people picker, and the
Area is never a secrecy mechanism.

A class is not part of a Document's identity. It never enters the code, never becomes a
numbering scope, and reclassification changes neither identity, code, revision ordinals,
effectivity nor history.

---

## 5. End-to-end Launch journeys

### A. Create

```text
Choose Document Type
→ allocate stable code
→ create REV000 DRAFT
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

If an older revision is EFFECTIVE, it remains EFFECTIVE. Cancellation does not delete historical identity or rewrite prior Submissions. A cancelled ordinal is never reused.

### H. Release / first effective version

When all required gates pass:

```text
REV000 SUBMITTED
→ system Release
→ REV000 EFFECTIVE
```

Readers can now discover the initial issuance as official content.

### I. Revise an effective Document

After the initial `REV000` is effective, the first business revision is `REV001`:

```text
REV000 EFFECTIVE
→ REV001 DRAFT
```

Readers continue seeing REV000 — including its governed title — while REV001 is drafted/governed. Successful Release of REV001 changes:

```text
REV000 → SUPERSEDED
REV001 → EFFECTIVE
```

The same law continues for later revisions (`REV001 → REV002`, etc.).

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
9. reactivation of an OBSOLETE Document is not Launch scope;
10. while a **human-governed** obsolescence request is still active and incomplete, an authorized initiator/manager may withdraw that request; the target remains EFFECTIVE, no fake `RETURN_FOR_CHANGES` is fabricated, and a later retry is a new request/attempt.

For a Document Type configured `NoHumanApproval`, governed obsolescence may complete with zero human Step after authorized initiation, mandatory reason and all eligibility/invariant checks; no fake System approver is created. Because completion occurs synchronously, there is no live human-governed request window to withdraw.

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
Controlled Document + business Revision starting at REV000
Revision-governed human-readable title
DRAFT Working Content + autosave/concurrency
Templates
Confidentiality Classes + clearance grants
immutable Submission
NoHumanApproval or sequential governance route
feedback / return / withdraw / resubmit
Revision cancellation
Approval evidence
Release / EFFECTIVE / SUPERSEDED
optional required Rendition
explicit governed OBSOLETE flow + bounded withdrawal of active human-governed request
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
- Revision ordinals start at `REV000` and never reuse; `REV001` is the first revision after initial issuance.
- Document identity remains stable; human-readable title is governed Revision metadata.
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
- **Obsolescence:** explicit governed product journey; `NoHumanApproval` may remove the human gate but never turns it into a raw status toggle. Active human-governed requests may be withdrawn before completion without changing the EFFECTIVE target.
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
3. Initial issuance is `REV000`; first revision is `REV001`; revision ordinals never reuse.
4. Ordinary readers see current effective truth, not moving drafts.
5. A Submission is immutable exact governed-candidate identity.
6. Governance/feedback always binds an exact Submission.
7. Return/withdraw never mutates prior Submission history.
8. Cancelling a Revision ends that business change cycle without disturbing an older EFFECTIVE revision.
9. Release is the only normal transition establishing a new EFFECTIVE revision.
10. Replacement Release supersedes the prior EFFECTIVE revision as one product transition.
11. Obsolescence without replacement is explicit, justified and governed; an active human-governed request may be withdrawn before completion without changing current effectivity.
12. EFFECTIVE/SUPERSEDED/OBSOLETE/CANCELLED are lifecycle facts, never physical delete commands.
13. Templates create independent Documents.
14. Search never grants access and never presents drafts as official to ordinary readers.
15. Audit proves actions but never becomes current-state authority.
16. Imported history never becomes fake native history.
17. Storage/provider identity never becomes Document/Revision/Submission identity.
18. Launch preserves governed history and has no governed physical disposition.
19. Later capabilities must attach without duplicating Document/Revision/Submission authority.
20. Human-readable title is Revision-governed metadata; a DRAFT/SUBMITTED retitle cannot rewrite the title ordinary readers see for the current EFFECTIVE revision.

---

## 9. Whole-product scenario gate

Any renewed technical architecture must answer all scenarios without contradictory authorities:

| Scenario | Required truth |
|---|---|
| Create blank | stable Document + REV000 DRAFT, not official |
| Create template-based | current EFFECTIVE template seeds independent Document as REV000 DRAFT |
| Autosave | preserve work without consuming REV/official history |
| Submit | freeze exact candidate |
| Return | old Submission stays immutable; same REV returns DRAFT |
| Withdraw | terminate governance attempt, same REV returns DRAFT |
| Cancel Revision | end business change cycle; older EFFECTIVE remains |
| Resubmit | new Submission, same business REV |
| Governance | actor judges exact Submission |
| First Release | establish REV000 as first EFFECTIVE truth |
| First Revision | after REV000, create REV001; prior EFFECTIVE stays reader truth until successor Release |
| Later Revision | increment ordinal monotonically without reuse |
| Retitle | title change belongs to the new Revision; readers keep current EFFECTIVE title until Release |
| Obsolete | governed withdrawal without successor |
| Withdraw obsolescence request | active human-governed request ends without fake RETURN; target remains EFFECTIVE |
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
→ create REV000 from blank/template
→ author safely, including governed title changes
→ submit exact content
→ govern it or return it
→ make REV000 effective as initial issuance
→ let readers reliably find/read the current effective revision
→ create REV001 as the first revision while REV000 stays effective
→ release successors without revision-number reuse
→ withdraw a Submission or cancel a change cycle correctly
→ obsolete an effective Document without replacement or withdraw an in-flight human-governed obsolescence request
→ inspect trustworthy lifecycle/audit history
→ migrate/restore without fabricating documentary truth
```

without Dossier, Evidence, Retention, Legal Hold, generic Change Control, governed export packages, repository sync or LMS machinery.

---

## 11. Authority and next gate

This file is the accepted Launch V1 product authority. The active technical architecture is routed by `wiki/architecture/r10-technical-architecture.md`.

```text
accepted Product Contract REV001
→ Whole-Product GCR / adjudication
→ approved 4+1 ownership
→ T1→T5 CLOSED / OPERATOR-RATIFIED
→ post-T5 Fable bounded amendments ratified
→ post-T5 delta review checkpoint
→ T6 only after checkpoint closure
→ T7
→ Whole-R10 Global Coherence Review
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

Implementation remains blocked until the R10 design/review gates close.