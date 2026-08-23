---
id: governance-review-layer-seam
kind: authority
owner: architecture
summary: Operator-ratified future seam for provider-neutral anchored governance review over exact immutable governed content, without promoting current Launch capability.
---

# Governance Review Layer future seam

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED FUTURE-SEAM AUTHORITY  
> **Ratified:** 2026-08-23  
> **Origin:** T11 B06-F2 — Governance Case DOCX Review Layer finding  
> **Implementation:** BLOCKED.  
> **Current Launch capability delta:** ZERO.

## 1. Decision outcome

DevelopmentConexus Decision Core outcome:

```text
CURRENT LAUNCH STRUCTURE CONFIRMED
+ BOUNDED FUTURE GOVERNANCE-REVIEW SEAM PRESERVED
```

Current Launch remains intentionally limited to:

```text
exact immutable governed content
GovernanceFeedback message
ACCEPT
RETURN_FOR_CHANGES + mandatory reason
```

No inline selected-range comment, review thread, resolve/reply action, tracked change, suggestion or automatic remediation operation is promoted by this authority.

## 2. Semantic separation

The following concepts remain distinct:

```text
Document Discussion
  stable-Document conversation

DRAFT EditorialComment
  separately deferred editor / mutable-WorkingContent concept

GovernanceFeedback
  immutable feedback inside one exact GovernanceAttempt

GovernanceDecision
  immutable ACCEPT / RETURN_FOR_CHANGES verdict
```

Future selected-range governance review must not be stored as stable Document Discussion and must not silently revive the deferred DRAFT EditorialComment platform.

## 3. Exact-snapshot binding law

If selected-range governance comments are promoted later, the preferred semantic direction is an anchored extension of GovernanceFeedback bound to the exact immutable governed subject.

```text
anchor absent
  -> current ordinary GovernanceFeedback remains valid

anchor present in a future promoted design
  -> identifies the exact immutable reviewed snapshot
  -> identifies an exact range/location within that snapshot
```

The governed Submission remains immutable. Creating review feedback never edits Submission bytes and never grants WorkingContent mutation authority.

A review anchor never silently follows:

```text
a later DRAFT
another Submission
a newer/current Revision
another rendition with unproven correspondence
```

## 4. Provider-neutral anchor boundary

The exact anchor wire/encoding is deliberately deferred until a real promotion.

A future design must prove a stable semantic anchor that does not make any editor/runtime identifier the sole authority, including:

```text
vendor comment ids
editor document-node ids
DOM offsets
browser selection objects
framework-specific positions
```

An editor may implement selection, rendering, OOXML import/export or review mechanics. It may not own MetalDocs Submission identity, GovernanceFeedback meaning, Decision meaning, lifecycle, Authorization or anchor authority.

## 5. Suggestions / tracked changes remain separate

A proposed content change is not merely a comment with different presentation.

Future invariant:

```text
reviewer proposes a delta against an exact immutable snapshot
!= reviewer edits the governed Submission
!= reviewer edits current WorkingContent
```

Tracked-change / suggestion promotion therefore requires a separate bounded decision that closes at least:

```text
proposal identity
exact snapshot/range identity
insert/delete/replace semantics
accept/reject ownership
relationship to RETURN_FOR_CHANGES
whether and how anything may be applied to a later DRAFT
```

This seam authorizes no automatic patch/application behavior.

## 6. RETURN_FOR_CHANGES law

Current lifecycle truth remains unchanged:

```text
review exact Submission S1
-> RETURN_FOR_CHANGES
-> S1 + Decision + GovernanceFeedback remain immutable
-> same Revision returns SUBMITTED -> DRAFT
-> author resumes mutable WorkingContent
```

Any future review anchor created against S1 remains attached to S1.

It must not be blindly overlaid onto the returned DRAFT because the mutable bytes may already differ.

If future UX wants to map old review anchors onto a changed DRAFT, that requires a separate proven reconciliation/mapping model. Client-side text matching or editor heuristics cannot become authoritative remapping.

## 7. Future B04 remediation boundary

After RETURN_FOR_CHANGES, current DRAFT truth does not expose the old Submission as a current submission identity. B04 also owns no browser-side History reconstruction authority.

Therefore a future author-visible review-context feature must receive an explicit server-authored review-context identity/projection or another bounded admitted read.

It must not:

```text
scan History in the browser to guess the returned attempt
infer the latest returned Submission
carry stale navigation state as business authority
reuse current_submission_id after it is no longer current
```

The exact read/projection is deferred until the review capability is actually promoted.

## 8. Current B06 consequence

The operator-approved B06 Content-first Governance Workspace remains structurally compatible with the future seam:

```text
exact governed content remains dominant
B06-local governance context remains beside it
Decision remains a separate deliberate zone
```

Because this authority promotes no present-tense inline-review capability:

```text
current GovernanceFeedback wire   unchanged
current allowed_actions            unchanged
current B06 P8 R1 controls         unchanged
P8 R2 inline-review controls       not required
```

Do not add disabled or “coming soon” review controls merely to reserve visual space.

## 9. Current census impact

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

Operation 87+ still requires its own lawful current Product/T6 basis.

## 10. Proof obligations for any future promotion

A later promotion must be able to falsify at least:

```text
inline review mutates governed Submission bytes
review silently follows a newer DRAFT/current Revision
stable Document Discussion is used as governance-review storage
DRAFT EditorialComment is silently reused for Submission review
vendor-specific selection/comment identity becomes sole semantic anchor
RETURN remaps an old anchor onto changed DRAFT without proof
review participation grants WorkingContent mutation
suggestion is treated as accepted DRAFT mutation
B04 reconstructs old attempt through client-side History scan
framework availability creates lifecycle or Authorization authority
```

## 11. Reopen / promotion triggers

Promote/reopen only when material evidence establishes one of:

```text
inline selected-range review becomes a committed current-release requirement
a production-capable review mechanism/contract is selected and Product semantics must be realized
a customer/regulatory process requires anchored review evidence
tracked-change/suggestion becomes a committed separate Product requirement
measured usage proves unanchored governance feedback inadequate
a proven cross-format requirement justifies a broader annotation model
```

Until then, preserve the seam and do not implement dormant capability.
