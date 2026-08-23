# T11 — B07 Document History R1 — Method v2.2 locked

> **Status:** LOCKED / OPERATOR-RATIFIED / P9-P10 COMPLETE.  
> **Block:** B07 — Document History.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 LOCKED.  
> **Bounded authority:** `../../decisions/document-history-recognition-read.md`.  
> **Canonical P8:** `t11-b07-document-history-functional-wireframe.html`.  
> **Canonical HTML blob:** `20ec64d34085fbc9075b136a61e69c48c0cad981`.  
> **Post-LOCK proof:** `t11-b07-screen-contract.md` + `t11-b07-pattern-consolidation.md`.  
> **Implementation:** BLOCKED.

## 1. Lock basis

The operator explicitly approved:

```text
B07-F1 human-recognizable History projection
P7 H1 Revision Chapters + chronological event spine
functional P8 R1 after browser operation
final B07 LOCK authorization
```

No material current-Launch B07 finding remains open.

The P8 hypothesis was accepted with the prototype: when later chronological events target an older Revision, authoritative server chronology remains primary and the Revision marker may repeat later rather than moving the event backward in time.

## 2. Current Product boundary

Document History is a Controlled Documents lens, not Audit and not lifecycle mutation authority.

```text
route
  /documents/:document_id/history

orientation
  47 getDocument

History
  53 getDocumentHistory

access
  document.read_history

writes
  none
```

History answers the controlled story of one Document across:

```text
business Revision cycles
immutable Submissions
governance feedback / Decisions
withdrawal / cancellation
Release / official rendition
obsolescence request / withdrawal / completion
```

Audit remains separate cross-product action evidence and is never queried to reconstruct B07 lifecycle meaning.

## 3. B07-F1 — CLOSED / OPERATOR-RATIFIED

Durable authority:

```text
../../decisions/document-history-recognition-read.md
```

Operation 53 remains the sole History list read and the census stays 86 operations / 11 routes / 16 PermissionCode values.

Every History event carries:

```text
revision: RevisionIdentity
```

Governance Decision additionally carries:

```text
subject_kind
step_label  // frozen GovernanceAttemptStep.label_snapshot
```

Governance feedback carries:

```text
subject_kind
```

Release predecessor is:

```text
predecessor_revision?: RevisionIdentity
```

No browser cross-page relationship graph, Audit join, fabricated title snapshot, new route/operation/Permission/owner/lifecycle state or History query DSL was introduced.

## 4. Ordering / pagination law

Operation 53 keeps the server-owned order:

```text
occurred_at ASC,
kind,
semantic id
```

B07 does not globally re-sort independently fetched cursor pages.

A later chronological event may target an older Revision. The locked UX preserves order and repeats the Revision marker when useful for recognition.

Pagination remains cursor-based with current disclosure/AuthZ rechecked on every page, no total count and no frozen multi-page snapshot.

## 5. Exact historical content

B07 adds no content endpoint.

```text
Submission event
  submission_id
  -> op63 getSubmission
  -> op64 getSubmissionSource

Release event
  release_id
  -> op72 getRelease
  -> op73 getReleaseSource or op74 getOfficialRenditionContent as applicable
```

Both use the shared Exact Read-Only Content Viewer Shell.

Historical content failure never substitutes another/current version.

## 6. P6 — COMPLETE

Reference evidence from Veeva Vault Timeline, M-Files Version History, SharePoint and Qualio supported:

```text
dedicated Document lifecycle/version History
historical exact content as read-only evidence
oldest -> newest as credible storytelling order
History distinct from broader Audit
```

Current Launch still excludes:

```text
restore historical version
delete historical version
compare/diff versions
workflow/task administration
free-form History query platform
```

## 7. P7 / P8 — LOCKED

Selected structure:

```text
H1 Revision Chapters + chronological event spine
```

Canonical P8:

```text
t11-b07-document-history-functional-wireframe.html
```

Material evidence exercised:

```text
REV000 return/resubmit/release
REV001 release
REV002 withdrawal + cancellation
later REV001 obsolescence request/withdrawal/retry/completion
cursor continuation + retry
exact Submission / Release viewer
exact-content failure without substitution
neutral 404
responsive reflow
B01N Quick Inbox inheritance
viewer Escape/focus recovery
```

The operator operated R1 and explicitly approved it on 2026-08-23.

## 8. P9 — COMPLETE

Post-LOCK proof:

```text
t11-b07-screen-contract.md
```

Closure:

```text
material regions/controls traced        20 / 20
unbound material controls               0
invented operations                     0
operation 87+                           0
screen-shaped APIs                      0
frontend historical graph authority     0
frontend Authorization evaluator        0
History mutations                       0
Audit reconstruction dependencies       0
material findings                       0
```

## 9. P10 — COMPLETE

Post-LOCK proof:

```text
t11-b07-pattern-consolidation.md
```

Shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
Exact Read-Only Content Viewer Shell
```

New shared patterns graduated:

```text
none
```

B07-local patterns remain local, including Revision Chapters, repeated Revision markers for later activity, human-recognizable History rows and History-specific chronological continuation.

No generic Timeline/Event/History abstraction was created merely from geometry.

## 10. Closure

```text
B07-F1              CLOSED / OPERATOR-RATIFIED
P6                  COMPLETE
P7 H1               OPERATOR-APPROVED
P8 R1               LOCKED / OPERATOR-RATIFIED
P9                  COMPLETE
P10                 COMPLETE
B07                  LOCKED / OPERATOR-RATIFIED
B08+                 NOT OPEN
```

B08 may become the next eligible FP1 block only after this closure is recorded in the repository roadmap. Implementation remains blocked.
