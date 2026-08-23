# T11 — B07 Document History R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / ENTRY RECOVERY COMPLETE / P6 COMPLETE / B07-F1 CANDIDATE / P7 AWAITING OPERATOR ADJUDICATION.  
> **Block:** B07 — Document History.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 LOCKED.  
> **Implementation:** BLOCKED.  
> **P8:** NOT OPEN.

## 1. Entry recovery

Document History is a Controlled Documents lens, not Audit and not a lifecycle owner.

```text
stable route
  /documents/:document_id/history

primary history read
  53 getDocumentHistory
  GET /api/v1/documents/{document_id}/history

orientation read
  47 getDocument

access
  document.read_history

writes
  none
```

History answers:

```text
what happened to this one controlled Document?
which business Revision did an event belong to?
what immutable Submission was governed?
who decided / returned / withdrew / cancelled?
when did a Release become effective?
what happened during governed obsolescence?
```

Audit remains separate:

```text
Document History
  Controlled Documents semantic facts for one Document

Audit
  meaningful action evidence across the product
```

Audit must never be queried to reconstruct current Document lifecycle or used as B07's event source.

## 2. Current History wire

Operation 53 returns:

```text
DocumentHistoryPage {
  items: DocumentHistoryItem[]
  page: Page
}
```

Canonical order is server-owned:

```text
occurred_at ASC,
kind,
semantic id
```

The frontend does not globally re-sort cursor pages. There is no current sort/filter/search DSL, total count or frozen multi-page snapshot.

Current closed event family:

```text
revision_created
submission_created
governance_decision
feedback_added
submission_withdrawn
revision_cancelled
release_completed
official_rendition_completed
obsolescence_requested
obsolescence_withdrawn
obsolescence_completed
```

Autosave is correctly absent: WorkingContent autosaves/checkpoints do not create business Revision/Submission/history.

### Existing exact-content continuation

B07 needs no new content endpoint.

```text
submission_created.submission_id
  -> op63 getSubmission
  -> op64 getSubmissionSource
  -> exact immutable submitted content

release_completed.release_id
  -> ops72–74 existing Release/source/official-rendition reads as applicable
  -> exact released official content
```

These may open the already-graduated Exact Read-Only Content Viewer Shell inside the History lens. B07 does not mutate historical bytes.

## 3. Human jobs

### J1 — reconstruct the Document's controlled story

```text
When I open History,
I need to follow the Document from REV000 through submissions, governance, releases and obsolescence,
so that I can explain how official truth arrived at its current state.
```

### J2 — recognize exact revision cycles

```text
When several business revisions and resubmissions exist,
I need every event to remain recognizably tied to the exact business Revision/context,
so that I do not have to interpret UUIDs or manually infer which cycle an approval belongs to.
```

### J3 — inspect historical exact content

```text
When an event refers to a Submission or Release,
I need to open the exact immutable historical bytes,
so that the timeline is evidence-backed rather than a list of labels only.
```

### J4 — understand governance evidence without entering B06

```text
When History shows feedback or a Decision,
I need enough human context to understand what happened,
without requiring current Governance Case participation/access or reconstructing a case in the browser.
```

### J5 — distinguish History from Audit

```text
When I need the business lifecycle story of one Document,
I stay in History;
when I need broader action evidence across the system,
I use Audit.
```

## 4. B07-F1 — human-recognizable History projection — CANDIDATE

### Evidence / root cause

Current `DocumentHistoryItem` is semantically truthful but inconsistent in human recognition.

Examples:

```text
revision_created
  revision: RevisionIdentity

submission_created
  revision: RevisionIdentity
  title

governance_decision
  governance_attempt_id
  step_id                 // no Step label
  actor / outcome / reason

feedback_added
  governance_attempt_id   // no subject kind / Revision identity

revision_cancelled
  revision_id             // opaque id only

release_completed
  revision_id             // opaque id only
  predecessor_revision_id?

official_rendition_completed
  submission_id           // Revision must be reconstructed indirectly

obsolescence_withdrawn
  request_id              // target Revision must be reconstructed indirectly
```

A browser can theoretically build maps across previously loaded events:

```text
attempt_id -> Submission/obsolescence subject
submission_id -> Revision
request_id -> target Revision
revision_id -> ordinal
```

That is rejected as the baseline B07 contract. It makes human history recognition dependent on client-side cross-page graph reconstruction and means an independently rendered event may expose opaque identifiers instead of owner-authored context.

### Target invariant

> Every History event must be independently attributable to its exact business Revision and, for governance decisions, to a human-recognizable frozen Step label. Governance event context must distinguish Submission governance from obsolescence governance without a browser reconstructing the attempt subject.

### Smallest sustainable refinement

Keep operation 53, route, permission, owner and pagination unchanged. Refine only `DocumentHistoryItem` projection:

```text
COMMON
  revision: RevisionIdentity
    required on every History variant
    exact business Revision to which the event belongs

GOVERNANCE EVENTS
  governance_decision
    + subject_kind: GovernanceSubjectKind
    + step_label: ShortText
      from immutable GovernanceAttemptStep.label_snapshot

  feedback_added
    + subject_kind: GovernanceSubjectKind

RELEASE
  predecessor_revision_id?
    -> predecessor_revision?: RevisionIdentity
```

Where a variant currently exposes a bare `revision_id` / `target_revision_id`, the common `revision:RevisionIdentity` becomes the human-usable exact identity rather than duplicating two competing fields.

Source law:

```text
Submission events        -> Submission Revision
Governance Submission    -> exact Submission Revision
Governance Obsolescence  -> exact target Revision
Release/rendition        -> exact released Submission Revision
Obsolescence events      -> exact target Revision
Cancellation             -> exact cancelled Revision
```

This is derived owner read truth only. It creates no event, snapshot, lifecycle fact or historical title that does not exist.

### Explicit non-goals

```text
no new operation 87+
no new route
no new Permission
no new semantic owner
no Audit join
no generic History query DSL
no frontend relationship graph as business authority
no fabricated title-at-revision-creation snapshot
```

## 5. P6 — bounded reference study — COMPLETE

### Veeva Vault — Timeline View

Useful evidence:

```text
Document Timeline keeps versions together with workflow/lifecycle events.
Version numbers are clickable into exact historical document context.
Document Audit History is a separate broader action trail.
```

Accepted lesson:

> Document-specific lifecycle history and system/action Audit should remain distinct; historical content should be reachable from the lifecycle story.

Rejected Veeva breadth:

```text
workflow administration actions
reassignment/task administration
generic lifecycle configuration inside History
100-task workflow mechanics
```

### M-Files — Version History

Useful evidence:

```text
history is recorded oldest -> newest
prior versions are read-only
selection can expose historical content/metadata
```

Accepted lesson:

> Chronological origin-to-current storytelling is credible; no B07 reorder finding is justified merely because another product might show newest first.

Rejected current-scope capabilities:

```text
rollback
version labels
compare selected documents
```

### SharePoint / Qualio

Useful evidence:

```text
version history is a dedicated historical lens
historical content/versions can be inspected
comparison is a separate deliberate capability
```

Rejected for current MetalDocs Launch:

```text
restore previous version
delete historical version
content diff/compare control
revision-history free-text platform
```

Current Product contract does not admit those mutations/capabilities.

### P6 saturation

Additional products stopped changing the decision space.

```text
P6 = COMPLETE
```

## 6. P7 hypotheses

### H1 — Revision Chapters + chronological event spine — LEADING

```text
minimal History header
  back to Document Official
  Document code
  current official Revision/title/status from op47
  explicit "Histórico do documento"

main history
  REV000 chapter
    revision created
    Submission S1
    feedback / Step Decisions
    return / resubmit if present
    Release

  REV001 chapter
    revision created
    Submission(s)
    governance
    cancellation OR Release

  ...

  final/current chapter

select content-bearing event
  -> exact read-only historical viewer
```

Why leading:

```text
matches Product's business Revision mental model
keeps resubmissions inside the same revision cycle
makes approval/return/release sequences scannable
preserves server chronological order
supports long documents through cursor continuation
keeps exact content reachable without turning History into an editor
```

B07-F1 is what makes these chapters server-truthful without client graph reconstruction.

### H2 — flat event timeline

```text
one event after another
no revision chapters
```

Strength:

```text
closest visual mapping to op53
```

Rejected as leading because:

```text
weak recognition of revision cycles
resubmissions/approval loops become visually noisy
approaches an audit-feed mental model
```

### H3 — revision table + event detail drawer

```text
REV | title | outcome | date
select row -> lifecycle events/detail
```

Strength:

```text
compact for many revisions
```

Rejected as leading because current History authority is event-oriented, not one version-summary aggregate. Producing a truthful row outcome/title summary would either hide relevant governance/obsolescence facts or encourage frontend aggregation that is not currently authored as a server read model.

## 7. Leading P7 region model

```text
R1  History orientation header + B03 return
R2  current official context (orientation only; never History authority)
R3  chronological Revision chapter spine
R4  revision-created event
R5  Submission event + exact Submission open
R6  Governance feedback event
R7  Governance Step Decision event + frozen Step label
R8  withdrawal / return context
R9  Revision cancellation event
R10 Release event + exact official open
R11 official-rendition completion event
R12 obsolescence request / withdrawal / completion
R13 cursor continuation / loading state
R14 history unavailable/non-disclosable state
R15 exact historical-content failure/viewer state
```

There is no B07 write zone.

## 8. Interaction / failure law

```text
route deep-link
  -> op47 orientation + op53 History

op53 404 / non-disclosable
  -> neutral History unavailable; safe return to B03

pagination failure
  -> preserve already loaded truthful events; retry continuation

historical content open
  -> exact existing owner read
  -> integrity/dependency failure never substitutes another Submission/Release

current Document drift while History is open
  -> header orientation may refresh
  -> already accepted historical events remain historical facts
```

No event click silently mutates current state.

## 9. Responsive / accessibility direction

Desktop:

```text
single dominant chronological column
revision chapter marker remains sticky/recognizable where useful
historical viewer may use bounded overlay/adjacent read-only surface
```

Narrow:

```text
orientation
-> revision marker
-> events in chronological order
-> selected historical content after deliberate open
```

Accessibility obligations:

```text
revision boundaries use semantic headings, not color only
event kind has visible text, not icon only
actor/time/outcome remain text
RETURN reason remains readable
Step label remains human-readable
viewer focus/close path explicit
load-more preserves reading/focus position
```

## 10. Current gate

```text
B07 entry recovery                       COMPLETE
B07-F1 human-recognizable projection     CANDIDATE / AWAITING OPERATOR RATIFICATION
P6                                       COMPLETE
P7 H1 Revision Chapters                  LEADING / AWAITING OPERATOR APPROVAL
P8                                       NOT OPEN
B07 LOCK / P9 / P10                      NOT OPEN
B08+                                     NOT OPEN
```

Next gate:

> Operator adjudicates B07-F1 and P7 H1. If approved, promote only the bounded History projection precision, then create the B07 functional P8 HTML. If rejected/revised, remain in B07 and do not open B08.
