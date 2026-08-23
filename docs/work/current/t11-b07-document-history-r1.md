# T11 — B07 Document History R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / ENTRY RECOVERY COMPLETE / B07-F1 OPERATOR-RATIFIED / P6 COMPLETE / P7 H1 OPERATOR-APPROVED / P8 R1 READY FOR OPERATOR USE.  
> **Block:** B07 — Document History.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 LOCKED.  
> **Bounded authority:** `../../decisions/document-history-recognition-read.md`.  
> **Canonical P8 R1:** `t11-b07-document-history-functional-wireframe.html`.  
> **Implementation:** BLOCKED.  
> **LOCK:** NOT YET — operator must operate/iterate P8 first.

## 1. Entry recovery outcome

Document History is a Controlled Documents lens, not Audit and not a lifecycle owner.

```text
stable route
  /documents/:document_id/history

orientation read
  47 getDocument

History read
  53 getDocumentHistory

access
  document.read_history

writes
  none
```

History answers the controlled story of one Document:

```text
business Revision cycles
immutable Submissions
governance feedback / Decisions
withdrawal / cancellation
Release / official rendition
obsolescence request / withdrawal / completion
```

Audit remains separate action evidence across the product and is never B07 lifecycle reconstruction authority.

## 2. B07-F1 — CLOSED / OPERATOR-RATIFIED

Durable authority:

```text
../../decisions/document-history-recognition-read.md
```

Current op53 remains operation 53 and gains only human-recognition projection precision.

Every `DocumentHistoryItem` now carries:

```text
revision: RevisionIdentity
```

Governance Decision additionally carries:

```text
subject_kind: GovernanceSubjectKind
step_label: ShortText
```

where `step_label` comes from immutable `GovernanceAttemptStep.label_snapshot`, never current route configuration.

Governance feedback additionally carries:

```text
subject_kind: GovernanceSubjectKind
```

Release predecessor is projected as:

```text
predecessor_revision?: RevisionIdentity
```

The refinement creates no:

```text
operation 87+
route
Permission
semantic owner
lifecycle state
Audit join
History sort/filter/search DSL
historical title-at-revision-creation snapshot
frontend historical relationship graph as authority
```

Current census remains 86 operations / 11 routes / 16 PermissionCode values.

## 3. Ordering / pagination law

Operation 53 keeps its server-owned order:

```text
occurred_at ASC,
kind,
semantic id
```

B07 does not globally re-sort independently fetched cursor pages.

Pagination stays:

```text
first page -> optional limit
next page  -> cursor + optional limit
no total count
no frozen multi-page snapshot
current disclosure/AuthZ rechecked every page
```

Already loaded historical facts remain valid if continuation loading fails.

## 4. Exact historical content

B07 adds no content endpoint.

Submission event:

```text
submission_id
-> op63 getSubmission
-> op64 getSubmissionSource
-> exact immutable submitted bytes
```

Release event:

```text
release_id
-> existing Release/source/rendition reads
-> exact released official bytes
```

Both use the already-graduated Exact Read-Only Content Viewer Shell.

Historical content failure must never substitute another/current version.

## 5. Human jobs

```text
J1 reconstruct the controlled Document story from origin to current state
J2 recognize which exact business Revision each event belongs to
J3 inspect exact immutable historical content when evidence-bearing identity exists
J4 understand governance feedback/Decision without entering a current Governance Case
J5 keep Document History conceptually distinct from Audit
```

## 6. P6 — COMPLETE

External evidence saturated after Veeva Vault Timeline, M-Files Version History, SharePoint and Qualio.

Accepted evidence:

```text
Document-specific lifecycle/version history is a dedicated lens
historical versions/content are read-only and deliberately inspectable
oldest -> newest history is a credible storytelling order
workflow/lifecycle history and broader Audit remain distinct
```

Rejected for current Launch:

```text
restore historical version
delete historical version
compare/diff versions
workflow/task administration
free-form History query platform
```

## 7. P7 — H1 OPERATOR-APPROVED

Selected structure:

```text
H1 Revision Chapters + chronological event spine
```

Composition:

```text
minimal History header
  return to Document Official
  Document code
  current official orientation from op47
  explicit History identity

chronological event spine
  Revision marker
  revision-created event
  Submission(s)
  governance feedback / Decisions
  withdrawal / cancellation
  Release / rendition
  obsolescence facts when that Revision is target

content-bearing event
  -> exact read-only historical viewer
```

Rejected as leading:

```text
flat audit-like event feed
revision summary table requiring browser aggregation
compare/restore/delete controls without Product authority
```

## 8. P8 R1 — functional low-fidelity evidence

Canonical file:

```text
docs/work/current/t11-b07-document-history-functional-wireframe.html
```

R1 is pure HTML/CSS/vanilla JavaScript with deterministic local fixtures only. It contains no production framework, API call, frontend Authorization evaluator or Product schema implementation.

### Material regions / behavior exercised

```text
R1  History orientation header + B03 return boundary
R2  current official orientation from op47
R3  chronological Revision markers
R4  revision-created event without fabricated title snapshot
R5  Submission event + exact historical Submission viewer
R6  governance feedback + subject kind
R7  Governance Decision + frozen human Step label
R8  RETURN_FOR_CHANGES reason
R9  Submission withdrawal
R10 Revision cancellation
R11 Release + predecessor Revision + exact official viewer
R12 official-rendition completion
R13 obsolescence request / withdrawal / completion
R14 cursor continuation
R15 continuation failure / retry preserving loaded facts
R16 History unavailable / disclosure-neutral 404
R17 exact historical-content failure / retry without substitution
R18 global Notification Quick Inbox inherited from B01N
R19 responsive chronological reflow
R20 viewer Escape/focus return + aria-live recovery announcements
```

### Deterministic timeline fixtures

R1 deliberately includes:

```text
REV000
  S1 -> feedback -> technical ACCEPT -> manager RETURN
  S2 -> ACCEPTs -> rendition -> Release

REV001
  S3 -> feedback -> Decisions -> Release

REV002
  S4 withdrawal -> S5 -> Revision cancellation

later chronology
  REV001 becomes the target again
  obsolescence request -> withdrawal
  new obsolescence request -> Decisions/feedback -> completion
```

This proves the UI cannot assume that all events belonging to a Revision are one permanently contiguous lifetime block.

## 9. P8-specific hypothesis under operator review

The server's chronological order remains authoritative.

When later chronological events target an older Revision after a newer Revision cycle has occurred, R1 does **not** move those events backward into the older visual block.

Instead it renders another Revision marker:

```text
REV001
  original cycle / Release

REV002
  later cycle / cancellation

REV001
  Revision retomada mais tarde na cronologia
  obsolescence events
```

This is a **P8 UX hypothesis**, not a new lifecycle/API rule.

Operator use should decide whether repeating the Revision marker preserves chronology clearly enough or whether another bounded visual treatment is needed.

## 10. Responsive / accessibility structure

Desktop:

```text
single dominant chronological reading column
Revision marker remains prominent/sticky while reading a segment
exact historical viewer opens as bounded read-only surface
```

Narrow:

```text
orientation
-> Revision marker
-> events in server chronological order
-> exact viewer only after deliberate open
```

Structural obligations represented:

```text
Revision boundaries are text headings, not color-only
Event kind is explicit text
actor / instant / outcome remain readable text
RETURN reason is visible text
Step label is human text, never UUID-only
content viewer has close/Escape + focus return
load-more failure preserves loaded reading position/truth
```

## 11. Current P8 gate

```text
entry recovery                         COMPLETE
B07-F1                                 CLOSED / OPERATOR-RATIFIED
P6                                     COMPLETE
P7 H1                                  OPERATOR-APPROVED
functional P8 R1                       EXISTS / READY FOR OPERATOR USE
P8 UX hypothesis                       AWAITING OPERATOR ADJUDICATION
B07 LOCK                               NOT YET
P9 / P10                               NOT OPEN
B08+                                   NOT OPEN
```

Next gate:

> Operator opens/operates B07 P8 R1, especially loading the second page where REV001 reappears after REV002 cancellation, opening historical Submission/Release content, forcing pagination/content failure and checking the 404 state. If friction exists, revise the same HTML. Only explicit B07 LOCK opens P9 then P10.
