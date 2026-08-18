# MetalDocs Launch V1 — Product Contract

> **Status:** NON-AUTHORITATIVE PRODUCT CONTRACT CANDIDATE — OPERATOR WRITTEN REVIEW PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This contract freezes **what the Launch V1 product must do and mean before technical architecture is allowed to continue**. It deliberately avoids SQL, table design, package boundaries, provider keys, object-store topology and implementation mechanics.

The current R10-C candidate is paused while this contract is under review. Earlier R9.5/R10 decisions remain evidence and historical design inputs; they are not allowed to force a Launch capability or abstraction that this product contract does not justify.

---

## 1. North Star

> **MetalDocs is the company system for creating, governing, approving, publishing, finding and proving the history of official controlled documents.**

Launch V1 is **not** a generic ECM, file drive, BPM engine, QMS suite, LMS, eDiscovery system, records-management platform, PLM, ERP or generic integration platform.

The product must make these questions easy and unambiguous:

```text
What is the official document?
Which revision is currently valid?
What is being changed right now?
What exact content was submitted and approved?
Who reviewed/approved it and when?
What changed between revision cycles?
Can an ordinary reader find the current valid content without seeing draft noise?
How did an effective document become obsolete?
What happened historically without rewriting history?
```

---

## 2. Reference-platform test

Reference products are used as **falsification evidence**, never as feature checklists.

### SharePoint — lower bound for versioning/publishing

Official Microsoft guidance distinguishes draft/pending content from approved/published versions, controls who can see drafts and lets ordinary readers continue seeing the last approved version while changes are pending.

Evidence:
- https://support.microsoft.com/en-US/SharePoint/lists/documents-and-library/how-versioning-works-in-lists-and-libraries
- https://support.microsoft.com/en-US/SharePoint/libraries/plan-document-versioning-content-approval-and-check-out-controls-in-sharepoint

**MetalDocs implication:** draft work and effective content must remain distinct; a new revision in progress must not make the prior effective revision disappear from readers.

### M-Files — metadata/relationship/template evidence

M-Files treats relationships as references between independently versioned objects rather than copies, and allows ordinary objects to serve as templates.

Evidence:
- https://userguide.m-files.com/user-guide/latest/eng/object_relationships.html
- https://userguide.m-files.com/user-guide/latest/eng/using_template.html

**MetalDocs implication:** template-as-document is sound; contextual relationships can be added later without turning the controlled-document core into a folder/graph platform.

### Qualio — focused controlled-document evidence

Qualio exposes draft, review, approval, make-effective, retire and audit behaviors, and distinguishes broader periodic-review/training capabilities from the core document path.

Evidence:
- https://docs.qualio.com/en/articles/6526420-user-permissions
- https://docs.qualio.com/en/articles/6597163-first-document-review
- https://docs.qualio.com/en/articles/11122-audit-trail-overview

**MetalDocs implication:** the core create → review/approve → effective → retire/obsolete path is real product complexity; periodic review/training are additional capability layers, not prerequisites for first controlled-document use.

### Veeva QualityDocs — upper bound / overengineering guard

Veeva supports rich Document Change Control, periodic review, obsolescence, Read & Understood and Training. It demonstrates mature regulated behavior but also shows how quickly controlled documents can expand into a large configurable quality platform.

Evidence:
- https://quality.veevavault.help/en/lr/15349/
- https://quality.veevavault.help/en/lr/37406/
- https://quality.veevavault.help/en/lr/72024/
- https://quality.veevavault.help/en/gr/71995/

**MetalDocs implication:** explicit obsolescence is important; generic change-control/workflow/training machinery is not Launch entitlement.

---

## 3. Launch personas

### Administrator / Governance Admin

Configures the company structure and the controlled-document rules required for operation:

- users, areas and groups;
- document types;
- numbering rules;
- templates;
- who may create/edit/read/review/approve by product role/scope;
- the governance route applied to a document type.

Launch does not provide a generic workflow designer, low-code rules engine or custom metadata/form platform.

### Author / Document Owner

Creates a document, edits the current draft revision, collaborates, submits exact content for governance, receives return feedback, revises and resubmits.

### Reviewer / Approver

Sees the **exact submitted candidate**, leaves comments/suggestions where allowed and records a governance decision. A decision never mutates the submitted candidate.

### Reader

Finds and reads the **current effective revision** by default. Reader journeys do not surface draft/submitted work as if it were official content.

### Auditor / Governance Viewer

Can reconstruct the document's lifecycle and action history from domain history plus audit evidence without Audit becoming the current-state authority.

---

## 4. Core product concepts

### Controlled Document

Stable official identity across its lifetime.

Examples:

```text
PO-001 — Procedimento de Compras
PL-003 — Política de Entregas
IT-012 — Instrução de Conferência
```

A Document is **not a file** and is not replaced when its content changes.

### Business Revision

One governed change cycle of a Document:

```text
REV001
REV002
REV003
```

Revision numbers are monotonic and never reused.

Core lifecycle:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

At most one current effective revision exists. A newer DRAFT/SUBMITTED revision may coexist with the prior EFFECTIVE revision.

### Draft Working Content

The mutable current work for the open revision. Autosave/checkpoints are technical working history and do not consume business revision numbers.

### Submission

An immutable exact attempt created by **Submit for governance**.

A Submission freezes the decision-relevant state, including exact content, governed metadata and relevant provenance/digest. Reviewers approve the Submission, not a moving draft.

Same-revision return/resubmit creates a new immutable Submission.

### Governance Route

One sequential governance concept with product-language step labels such as:

```text
Revisão técnica
Gestor da área
Qualidade
Diretoria
```

Launch does not create separate ReviewWorkflow and ApprovalWorkflow engines. Step behavior may differ by participants/quorum/fresh-auth, but collaboration and decision ceremony are orthogonal to step identity.

Normal decision vocabulary:

```text
ACCEPT
RETURN_FOR_CHANGES
```

Withdraw/cancel are separate lifecycle operations, not fabricated approval outcomes.

### Release / Effectivity

When every required gate for a submitted candidate is satisfied, the system performs the release/effectivity transition. There is no independent user “publish whichever file is latest” concept.

### Rendition / Official Representation

The primary submitted source always remains meaningful. A document type may use source-only representation or require a specific derived representation such as PDF. A representation is never allowed to silently change what humans approved.

### Template

A template is an ordinary governed Document used as a source for new documents. It does not have a parallel TemplateVersion lifecycle.

---

## 5. End-to-end Launch journeys

### Journey A — Create from blank

```text
Choose Document Type
→ allocate stable document code
→ create REV001 DRAFT
→ edit
```

The new document is not official until released.

### Journey B — Create from template

```text
Choose Document Type
→ choose eligible current EFFECTIVE template
→ seed new independent document/revision
→ edit independently
```

The derived document does not remain coupled to later template changes.

### Journey C — Draft/autosave

The author edits current Working Content. Autosave keeps work recoverable and concurrency-safe but does not manufacture official revisions, submissions or permanent governed history on every keystroke.

### Journey D — Submit

```text
DRAFT
→ validate submit requirements
→ freeze exact Submission S1
→ SUBMITTED
```

After submission the candidate is immutable.

### Journey E — Review/approval

Each active governance participant sees the exact Submission under decision.

They may add bounded feedback, then choose:

```text
ACCEPT
RETURN_FOR_CHANGES
```

Feedback never edits the Submission in place.

### Journey F — Return and resubmit

```text
Submission S1 remains immutable
REV returns to DRAFT
→ author changes Working Content
→ submit again
→ Submission S2
```

S1 and its decisions/feedback remain historical truth.

### Journey G — Release / first effective revision

When required approval, representation and timing gates pass:

```text
REV001 SUBMITTED
→ system release
→ REV001 EFFECTIVE
```

Ordinary readers can now discover and read REV001 as the official version.

### Journey H — Revise an effective document

```text
REV003 EFFECTIVE
→ create REV004 DRAFT
```

While REV004 is drafted/reviewed, normal readers continue seeing REV003 as effective. When REV004 is successfully released:

```text
REV003 → SUPERSEDED
REV004 → EFFECTIVE
```

### Journey I — Governed obsolescence without replacement

MetalDocs must support the case where an effective document is intentionally withdrawn from use without a successor revision.

Product contract:

1. only the current EFFECTIVE document state can enter an obsolescence request;
2. a nonblank reason is mandatory;
3. obsolescence is governed, not a raw status toggle;
4. while the request is unresolved, the existing EFFECTIVE revision remains official;
5. successful completion changes that revision to OBSOLETE and leaves no current EFFECTIVE revision;
6. normal search/current-document journeys stop presenting it as active content;
7. history remains accessible to authorized users;
8. if an open replacement revision already exists, obsolescence cannot complete until that competing open revision is cancelled/withdrawn or resolved;
9. reactivation of an obsolete document is not a Launch capability and requires an explicit future product decision.

Exact reuse/configuration of the existing governance route is a later architecture decision; the product invariant is that obsolescence is explicit, justified and governed.

### Journey J — Search/read/download

Normal reader discovery prioritizes **current EFFECTIVE documents**.

Core search/filter dimensions:

```text
code
name/title
Document Type
Area
responsible owner
status
```

Draft/submitted work is surfaced in appropriate author/governance workspaces, not mixed into ordinary current-document search as if equally authoritative.

Opening a controlled document shows at minimum:

```text
stable document identity
current effective revision/status
relevant effective/release date
approved source or official representation
change/revision history available by permission
```

### Journey K — Audit/history

A user can answer who performed meaningful governance actions and when. Domain records remain the authority for submissions, decisions, release, obsolescence and acknowledgements; transversal Audit proves actions/timeline and must not be queried as a substitute for current business state.

### Journey L — User offboarding

Disabling/offboarding a user must stop future access/actions while preserving historical attribution/evidence already created. Historical evidence is not rewritten merely because a profile becomes unavailable or erasable.

### Journey M — Historical migration before/around go-live

Migration is a deployment necessity when legacy controlled documents exist, not a generic everyday integration platform.

Rules:

- preserve reliable document codes/revision ordinals/states when truthfully known;
- preserve source provenance;
- unknown remains unknown;
- never fabricate native MetalDocs Approval, Release, User action or historical timestamps;
- migration is not allowed to replay old notifications or side effects as if they happened now.

---

## 6. Launch scope tiers

### Launch Core — required before first production use

```text
single-company deployment identity
users / areas / groups
roles / permissions / scoped access
Document Types
numbering
Controlled Document identity
business Revision lifecycle
DRAFT Working Content + autosave/concurrency
Templates
immutable Submission
sequential governance route
feedback + RETURN_FOR_CHANGES + resubmit
Approval evidence
Release / EFFECTIVE / SUPERSEDED
optional required Rendition by Document Type
explicit governed OBSOLETE flow
revision/history viewing
search/filter current effective content
source/official representation read/download
Audit trail
historical migration/cutover tooling when needed
backup/restore correctness required for go-live
```

### Launch+ — valuable next capabilities, deliberately not prerequisites for first controlled-document operation

Recommended placement:

```text
Distribution / acknowledgement (Read & Acknowledge, not LMS)
Periodic Review
```

Rationale: both are legitimate controlled-document capabilities, but the core create/govern/read lifecycle remains correct without them and they can be added later without replacing Document/Revision/Submission authority.

### Future product capabilities

```text
Dossier / documentary case context
Evidence capture / quality records
Retention policies
Legal Hold
Governed disposition/destruction
Records-driven WORM/Object Lock
eDiscovery/custodian preservation
Governed Subject Export packages
Generic External Repository IMPORT_COPY / PUBLISH_COPY connectors
Training/LMS/curricula/quizzes/qualification
Generic Change Control platform / multi-document change-control engine
pooled multi-customer tenancy
realtime coauthoring/CRDT
```

These ideas remain valid future options but create **no dormant Launch module/table/permission/job**.

---

## 7. Decisions from prior R10 work — whole-product adjudication candidate

### KEEP

1. **Document ≠ Revision ≠ Working Draft ≠ Submission.** This separation survives Structural Inversion and reference-platform comparison.
2. **REV numbers never reuse.**
3. **One current EFFECTIVE revision, with a newer DRAFT/SUBMITTED revision allowed concurrently.**
4. **Immutable Submission.** Governance binds an exact attempt, not mutable editor state.
5. **Return/resubmit on same business Revision creates a new Submission.**
6. **One sequential governance Step model.** Review-like and approval-like language does not justify two workflow engines.
7. **Automatic/system-owned Release as effectivity authority.**
8. **Template as role of ordinary governed Document.**
9. **Source content remains primary; optional required Rendition is representation policy, not universal PDF.**
10. **Search is projection/discovery only and never grants access.**
11. **Audit is timeline/forensic evidence, never current-state authority.**
12. **Authentication provider may own credential mechanics; MetalDocs owns product identity/authorization.**
13. **Historical migration cannot fabricate native governance history.**
14. **One company per Launch deployment.**
15. **Governed history is not physically disposed in Launch V1.**

### RESTRUCTURE / BOUNDED REOPEN

1. **Standalone Artifact semantic owner:** remove from Launch target. Exact content facts belong to the semantic event/record that freezes them (Submission, Rendition, imported content; future EvidenceCapture). Shared byte storage remains mechanism only.
2. **R10-C current candidate:** paused; storage design must be rebuilt only after this Product Contract is accepted.
3. **Obsolescence:** promote from an under-specified state value to an explicit governed product journey.
4. **B5 Launch scope:** Dossier/Evidence move out of Launch Core unless a concrete rollout consumer is named before ratification.
5. **B6 Governed Subject Export:** move to Future unless an auditor/customer/portability obligation is concretely named.
6. **Generic External Repository IMPORT/PUBLISH:** move to Future; Historical Migration remains a distinct go-live/cutover concern.

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

Reopen any defer only with a named consumer, contractual/regulatory requirement or production failure mode.

---

## 8. Product invariants

The following must remain true regardless of implementation:

1. A Document has one stable official identity across revisions.
2. A business Revision is not an autosave/checkpoint.
3. Revision ordinals never reuse.
4. Ordinary readers see the current effective truth, not a moving draft.
5. A Submission is immutable exact governed candidate identity.
6. Approval/feedback always binds one exact Submission.
7. Return-for-changes never mutates the old Submission.
8. Release is the only normal transition that establishes a new effective revision.
9. Releasing a replacement supersedes the prior effective revision atomically from product perspective.
10. Obsolescence without replacement is explicit, reasoned and governed.
11. EFFECTIVE/SUPERSEDED/OBSOLETE/CANCELLED are product lifecycle facts, not physical-file delete commands.
12. Templates create independent documents; future template changes do not rewrite derived documents.
13. Search/discovery never grants access and must not present draft content as official to ordinary readers.
14. Audit records that an action occurred but never becomes the canonical lifecycle authority.
15. Imported history never becomes fake native history.
16. Provider/file-storage identity never becomes Document/Revision/Submission identity.
17. Launch preserves confirmed governed history; no governed physical disposition exists.
18. Optional later capabilities must attach to this core without duplicating Document/Revision/Submission authority.

---

## 9. Whole-product scenario gate before renewed technical architecture

Every later architecture must give one simple, non-contradictory answer for each scenario:

| Scenario | Required product truth |
|---|---|
| Create blank | stable Document + REV001 DRAFT, not official yet |
| Create from template | exact current template source seeds independent document |
| Autosave | recover work without consuming REV or creating official history |
| Submit | freeze exact candidate |
| Return | old Submission remains immutable; same REV returns to DRAFT |
| Resubmit | new Submission, same business REV |
| Approval | participant judges exact Submission |
| Release | candidate becomes EFFECTIVE; current replacement truth changes once |
| New Revision | old EFFECTIVE stays reader truth until successor release |
| Obsolete | governed withdrawal with reason; no successor required |
| Search | normal search favors current EFFECTIVE and hides draft noise |
| Download/view | bytes/representation correspond to exact intended governed version |
| Offboarding | future access stops; history remains truthful |
| Migration | preserve source truth; fabricate nothing |
| Backup/restore | restored product cannot serve content whose governed bytes/history are inconsistent |
| Provider outage | business state remains truthful even if content mechanism is temporarily unavailable |
| Audit | reconstruct actions without deriving current state from logs |

If a proposed entity/module cannot be justified by one of these scenarios or another explicit invariant/requirement, it is presumptively YAGNI.

---

## 10. Success criterion for Launch V1

Launch V1 is product-complete when a real company can:

```text
configure document governance
→ create a policy/procedure/instruction
→ author it safely
→ submit the exact candidate
→ review/approve or return it
→ make it effective
→ let ordinary users find/read the current effective revision
→ create and approve a successor revision while the old one stays active
→ intentionally obsolete an effective document without replacement
→ inspect trustworthy history/audit
→ recover/migrate the system without fabricating documentary truth
```

without requiring Dossier, Evidence, Retention, Legal Hold, generic Change Control, governed export packages, repository sync or LMS machinery.

---

## 11. Review / promotion gate

This file is **not yet authority**. Required next step:

```text
operator written-contract review
→ bounded edits if needed
→ operator accepts Product Contract
→ promote a durable wiki/architecture Product Contract
→ run Whole-Product Global Coherence Review against prior R10 decisions
→ rebuild ownership/technical architecture only from accepted product semantics
```

Do not resume R10-C or write an implementation plan before this gate closes.
