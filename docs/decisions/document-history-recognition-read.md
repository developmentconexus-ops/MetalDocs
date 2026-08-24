---
id: document-history-recognition-read
kind: authority
owner: architecture
summary: Operator-ratified B07 precision making every Document History event independently attributable to its exact business Revision and governance context without browser-side historical graph reconstruction.
---

# Document History human-recognition read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified:** 2026-08-23  
> **Block:** B07 — Document History  
> **Method:** DevelopmentConexus Engineering Method + Frontend Product Experience Planning Method v2.2  
> **Impacts:** T6 Document History meaning, T8-E `DocumentHistoryItem`, T8-F B07 consumption.  
> **Implementation:** BLOCKED.

## 1. Decision outcome

DevelopmentConexus Decision Core outcome:

```text
CURRENT DOCUMENT HISTORY STRUCTURE CONFIRMED
+ BOUNDED HUMAN-RECOGNITION PROJECTION PRECISION
```

The existing route, owner, permission, operation and event family remain sufficient:

```text
GET /api/v1/documents/{document_id}/history
operation 53 getDocumentHistory
owner Controlled Documents
permission document.read_history
writes none
```

No operation 87+, new route, Permission, lifecycle state, semantic owner or Audit dependency is introduced.

## 2. Root cause

Current `DocumentHistoryItem` variants already contain truthful semantic facts, but several variants expose only technical identifiers for the human context that B07 must present.

Examples before this precision:

```text
governance_decision
  governance_attempt_id
  step_id
  actor / outcome / reason
  // no frozen Step label

feedback_added
  governance_attempt_id
  // no subject kind / Revision identity

revision_cancelled
  revision_id

release_completed
  revision_id
  predecessor_revision_id?

official_rendition_completed
  submission_id

obsolescence_withdrawn
  request_id
```

A browser could attempt to reconstruct maps such as:

```text
governance_attempt_id -> governed Submission / obsolescence target
submission_id         -> Revision
request_id            -> target Revision
revision_id           -> Revision ordinal
step_id               -> Step label
```

across previously loaded cursor pages.

That is rejected as the B07 correctness model.

> History recognition must not depend on client-side cross-event or cross-page graph reconstruction of owner truth.

## 3. Target invariant

> Every disclosed Document History event is independently attributable to the exact business Revision it belongs to. Governance decisions additionally carry the frozen human Step label, and governance feedback/decisions identify whether their subject was Submission governance or obsolescence governance.

The projection is derived read truth only. It does not create a new historical fact or snapshot.

## 4. Common Revision identity

Every `DocumentHistoryItem` variant carries:

```text
revision: RevisionIdentity
```

Source law:

```text
revision_created
  -> created Revision

submission_created | submission_withdrawn
  -> immutable Submission Revision

governance_decision | feedback_added
  submission subject    -> exact governed Submission Revision
  obsolescence subject  -> exact target Revision

revision_cancelled
  -> cancelled Revision

release_completed | official_rendition_completed
  -> exact released Submission Revision

obsolescence_requested | obsolescence_withdrawn | obsolescence_completed
  -> exact target Revision
```

`RevisionIdentity` remains:

```text
{ revision_id: Uuid, ordinal: RevisionOrdinal }
```

The wire does **not** fabricate a title-at-revision-creation snapshot. Exact historical submitted titles continue to come only from immutable Submission snapshots where already present.

## 5. Refined closed event union

The closed event family remains unchanged. Its human-recognition projection becomes:

```text
revision_created
  {
    kind,
    revision: RevisionIdentity,
    occurred_at
  }

submission_created
  {
    kind,
    submission_id,
    revision: RevisionIdentity,
    title,
    submitter,
    occurred_at,
    governance_attempt_id?
  }

governance_decision
  {
    kind,
    decision_id,
    governance_attempt_id,
    subject_kind: GovernanceSubjectKind,
    revision: RevisionIdentity,
    step_id,
    step_label: ShortText,
    actor,
    outcome,
    occurred_at,
    reason?
  }

feedback_added
  {
    kind,
    feedback_id,
    governance_attempt_id,
    subject_kind: GovernanceSubjectKind,
    revision: RevisionIdentity,
    actor,
    message,
    occurred_at
  }

submission_withdrawn
  {
    kind,
    submission_id,
    revision: RevisionIdentity,
    actor,
    occurred_at
  }

revision_cancelled
  {
    kind,
    revision: RevisionIdentity,
    actor,
    reason,
    occurred_at
  }

release_completed
  {
    kind,
    release_id,
    revision: RevisionIdentity,
    submission_id,
    occurred_at,
    predecessor_revision?: RevisionIdentity
  }

official_rendition_completed
  {
    kind,
    official_rendition_id,
    submission_id,
    revision: RevisionIdentity,
    occurred_at
  }

obsolescence_requested
  {
    kind,
    request_id,
    revision: RevisionIdentity,
    initiator,
    reason,
    occurred_at,
    governance_attempt_id?
  }

obsolescence_withdrawn
  {
    kind,
    request_id,
    revision: RevisionIdentity,
    actor,
    occurred_at
  }

obsolescence_completed
  {
    kind,
    request_id,
    revision: RevisionIdentity,
    occurred_at
  }
```

For `governance_decision`, existing presence law remains:

```text
reason present iff outcome = return_for_changes
```

## 6. Frozen Step label source

`governance_decision.step_label` is not a current governance-route lookup.

It comes from the immutable attempt Step snapshot:

```text
GovernanceAttemptStep.label_snapshot
```

Therefore a later governance-route label edit cannot rewrite historical Decision presentation.

B07 does not expose route configuration, candidate snapshots, grants or workflow administration data.

## 7. Governance subject recognition

`subject_kind` uses the existing closed `GovernanceSubjectKind`:

```text
submission
obsolescence
```

It exists only on the governance event variants that need that distinction:

```text
governance_decision
feedback_added
```

The browser must not infer subject kind from attempt ids, neighboring events or route position.

## 8. Predecessor Revision refinement

`release_completed` replaces the bare optional predecessor technical id with:

```text
predecessor_revision?: RevisionIdentity
```

This is exact release transition context only.

It does not create a generic Revision relation graph or authorize a compare/restore capability.

## 9. Ordering / pagination law unchanged

Operation 53 keeps its existing server-owned canonical order:

```text
occurred_at ASC,
kind,
semantic id
```

B07 may group already ordered events into visual Revision chapters using the returned `revision` identity, but it must not globally re-sort cursor pages.

Current pagination law remains:

```text
first page: optional limit
continuation: cursor + optional limit
current disclosure/AuthZ rechecked every page
no frozen multi-page snapshot
no total count
```

No History filter/search/sort DSL is added.

## 10. History versus Audit boundary

This precision does not merge B07 with Audit.

```text
Document History
  = Controlled Documents semantic lifecycle facts for one Document

Audit
  = meaningful action evidence across the product
```

B07 must not query Audit to fill missing history context or reconstruct current lifecycle truth.

## 11. Historical exact-content continuation

No content operation is added.

Existing identities remain sufficient:

```text
submission_created.submission_id
  -> op63 getSubmission
  -> op64 getSubmissionSource
  -> exact immutable submitted content

release_completed.release_id
  -> existing Release/source/rendition reads
  -> exact released official content
```

The already-graduated Exact Read-Only Content Viewer Shell may render these bytes inside B07. Historical viewing never mutates the source.

## 12. B07 presentation consequences

B07 may now safely present:

```text
REV000 / REV001 / REV002 chapter boundaries
Submission attempts inside their exact Revision cycle
Governance feedback and Decisions with human Step labels
RETURN reason inside the exact Revision cycle
Release predecessor/successor context without UUID interpretation
obsolescence events attached to the exact target Revision
```

B07 must not present current-Launch controls for:

```text
restore historical version
delete historical version
compare/diff versions
edit historical content
re-run governance
change Decision
```

Those capabilities require separate Product authority.

## 13. Census impact

```text
new application operations       0
new stable SPA routes            0
new PermissionCode values        0
new semantic owners              0
new lifecycle states             0
new ETag domains                 0
new Idempotency-Key creations    0
new exact-byte resources         0
new async workers                0
```

Current accepted census remains:

```text
86 operations / 11 routes / 16 PermissionCode values
```

## 14. Proof obligations

Later executable-contract/implementation proof must falsify at least:

```text
History event lacks exact Revision identity
Decision Step label is loaded from current route config instead of frozen attempt snapshot
browser infers governance subject kind from neighboring events
browser requires an earlier cursor page to interpret a later event
browser globally re-sorts independently fetched History pages
History queries Audit to recover Controlled Documents meaning
release predecessor is rendered from opaque UUID-only interpretation
revision_created fabricates a historical title snapshot
historical content viewer substitutes a newer/current version on failure
History exposes compare/restore/delete without separate authority
```

## 15. Reopen triggers

Reopen only the implicated decision if material evidence proves a need for:

```text
server-authored revision summary aggregates rather than event history
new History filter/sort/search requirements
content compare/diff
historical restore/revert
additional historical title/metadata snapshots not currently persisted
cross-Document history correlation
export/package evidence requiring a different projection
```

None is implied by B07-F1.
