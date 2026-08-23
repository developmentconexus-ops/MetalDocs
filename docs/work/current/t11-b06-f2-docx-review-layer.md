# T11 — B06-F2 DOCX Review Layer seam

> **Status:** DESIGN APPROVED IN CHAT / WRITTEN OPERATOR RATIFICATION PENDING.  
> **Parent block:** B06 — Governance Case.  
> **Trigger:** operator review of the B06 P8 R1 surfaced a real future need for DOCX selected-range comments / review suggestions without mutating the governed Submission.  
> **Method:** DevelopmentConexus Engineering Method + Frontend Product Experience Planning Method v2.2.  
> **Implementation:** BLOCKED.  
> **Current Product/API census delta:** ZERO.  
> **Current P8 delta:** ZERO unless this candidate is later promoted into present-tense Launch capability.

## 1. Question being closed

MetalDocs needs a sustainable path for a reviewer to eventually do Word-like review against DOCX content:

```text
select exact text/range
→ attach a comment/thread
→ optionally propose a content change
→ preserve the exact governed Submission
→ let the author understand the review after RETURN_FOR_CHANGES
```

The immediate mechanism evidence includes EigenPal/docx-editor.dev and SuperDoc, but no editor framework is allowed to define MetalDocs Product semantics.

The current Launch baseline remains valid without active inline-review commands:

```text
DOCX editing / viewing        EigenPal Apache-2.0 core is a viable mechanism baseline
active comments               not a current free-core capability relied upon by MetalDocs
tracked changes / suggestions not a current free-core capability relied upon by MetalDocs
SuperDoc Community            AGPLv3 and therefore not an admitted proprietary-product baseline
```

Commercial or future mechanisms may later compete behind the same Product seam.

## 2. Current authority evidence

Current authority already keeps four concepts distinct:

```text
Document Discussion
  stable-Document conversation

DRAFT EditorialComment
  deferred editor/DRAFT concept (CNT-13)

GovernanceFeedback
  immutable feedback fact inside one exact GovernanceAttempt

GovernanceDecision
  immutable ACCEPT / RETURN_FOR_CHANGES verdict
```

Therefore selected-range governance review must not be disguised as stable Document Discussion or silently revive the deferred DRAFT EditorialComment platform.

Current GovernanceFeedback is deliberately simple:

```text
GovernanceFeedbackView
  feedback_id
  actor
  message
  created_at

CreateGovernanceFeedbackRequest
  message
```

Current op69 owns unanchored governance feedback only. No anchor, thread, suggestion or editor-vendor identifier exists in current authority.

## 3. Root cause / target invariant

Root cause:

> B06 correctly centers the exact immutable governed content, but current Launch feedback can describe content only in prose. A future richer review experience needs a seam that preserves exact-snapshot identity without turning the reviewer into a WorkingContent editor or tying Product truth to one DOCX engine.

Target invariant:

> Any future inline governance review binds to the exact immutable governed snapshot being reviewed. It never mutates that Submission, never silently rebinds to later DRAFT bytes, and never makes EigenPal/SuperDoc-native IDs the semantic authority of MetalDocs.

## 4. Credible alternatives

### A — promote inline comments/tracked changes into Launch now

Rejected now.

```text
+ immediately mirrors Word-like review UX
- current free EigenPal baseline does not provide active review commands
- would require new Product/wire semantics before a present-tense delivery decision exists
- would create dormant/unsupported capability in violation of repository hard stops
```

### B — reuse stable Document Discussion for selected-range review

Rejected.

```text
Document Discussion owns stable-Document conversation
Governance review owns an exact immutable attempt/subject
```

Reusing Discussion would erase the lifecycle/subject boundary and make a comment appear to follow the Document even after the reviewed snapshot is no longer current.

### C — reopen DRAFT EditorialComment now

Rejected.

The review target is the immutable Submission under governance, not mutable WorkingContent. CNT-13 remains deferred until a real DRAFT-editor-comment consumer independently requires it.

### D — create a generic ReviewAnnotation platform now

Rejected as over-scoped.

It would introduce a new abstraction/aggregate before there is evidence that PDF, DRAFT, Submission, official Revision and other media require one common annotation lifecycle.

### E — preserve a provider-neutral governance-review seam and defer active capability

**SELECTED CANDIDATE.**

Prepare the invariants and future extension boundary now; keep current Launch API/P8 truthful and unchanged until the capability has a real admitted delivery mechanism/requirement.

## 5. Future anchored governance feedback seam

If selected-range comments are promoted later, the preferred semantic direction is an **anchored form of governance feedback** bound to the exact governed subject, not a second generic comment owner.

Conceptually:

```text
GovernanceFeedback
  message
  actor / trusted time
  governed attempt / exact subject identity
  anchor?  // future bounded extension only
```

Anchor law:

```text
anchor absent
  = current ordinary GovernanceFeedback remains valid

anchor present
  = exact immutable reviewed snapshot + exact range/location within that snapshot
```

The exact wire encoding of the anchor is intentionally NOT frozen here. A future promotion must prove a format that survives DOCX round-trip and remains independent of:

```text
EigenPal comment ids
SuperDoc ids
ProseMirror/Lexical positions
DOM offsets
browser selection objects
```

A useful mechanism may import/export OOXML review metadata, but mechanism identifiers never become the sole MetalDocs semantic anchor.

## 6. Suggestions / tracked changes are a separate concept

A suggested content change is not merely a comment with special styling.

Future suggestion invariant:

```text
reviewer proposes delta against exact immutable snapshot
≠ reviewer edits Submission
≠ reviewer edits current WorkingContent
```

Therefore tracked-change/suggestion promotion requires its own bounded decision even if a selected vendor bundles comments + tracked changes in one SDK.

At minimum a future design must close:

```text
proposal identity
exact snapshot/range identity
insert/delete/replace semantics
acceptance/rejection ownership
relationship to RETURN_FOR_CHANGES
what, if anything, may be applied to a later DRAFT
```

No automatic patch/application behavior is authorized by this seam.

## 7. RETURN_FOR_CHANGES / author remediation law

Current lifecycle truth is preserved:

```text
review exact Submission S1
→ RETURN_FOR_CHANGES
→ S1 + Decision + GovernanceFeedback remain immutable
→ same Revision returns SUBMITTED -> DRAFT
→ author edits mutable WorkingContent again
```

Critical consequence:

> Review anchors that were valid against S1 must never be blindly overlaid onto the mutable DRAFT after RETURN. The DRAFT may already differ, making old ranges stale or misleading.

Preferred future author UX:

```text
B04 current DRAFT editor
  remains current mutable work authority

separate explicit review-context affordance
  opens/places the old reviewed Submission S1 + its anchored review context
  author compares/remediates deliberately
```

If future UX wants anchors mapped onto current DRAFT, that requires a separate proven mapping/reconciliation model. Client-side text matching or editor heuristics cannot become authoritative remapping.

## 8. B04 consequence when capability is promoted

B04 currently cannot lawfully reconstruct a returned attempt from current DRAFT truth:

```text
RevisionView.current_submission_id
  present iff Revision state = submitted

RETURN_FOR_CHANGES
  same Revision becomes DRAFT
  old Submission remains history/evidence
  current_submission_id is therefore absent
```

B04 also deliberately owns no Governance Case/History reconstruction.

Therefore a future author-visible review-context feature must receive an explicit server-authored identity/navigation projection or another bounded admitted read. It must not:

```text
scan History in the browser
infer the latest returned attempt
carry stale navigation state as authority
reuse current_submission_id after it is no longer current
```

The exact projection/operation is deferred until the capability is actually promoted.

## 9. B06 visual consequence

The operator-approved H1 remains structurally compatible with future inline review:

```text
exact content remains dominant
+ content viewport can later expose review markers
+ B06-local governance rail can later expose selected anchored review context
+ Decision remains a separate deliberate zone
```

No three-column workflow cockpit is required.

Current R1 remains truthful because Launch currently offers only:

```text
unanchored GovernanceFeedback
ACCEPT
RETURN_FOR_CHANGES + reason
```

Do NOT add disabled/coming-soon comment, reply, resolve, suggest or tracked-change controls merely to reserve visual space.

## 10. Technology boundary

Current mechanism posture:

```text
EigenPal Apache core
  acceptable current DOCX editing/viewing baseline

EigenPal Pro
  future commercial candidate for review commands

SuperDoc Community
  not admitted under AGPLv3 for the proprietary baseline

SuperDoc commercial
  future commercial candidate

other engines/custom implementation
  may compete when evidence justifies evaluation
```

Product law:

```text
vendor can implement rendering/editor/review mechanics
vendor cannot own Submission identity, GovernanceFeedback meaning,
Decision meaning, lifecycle, Authorization or future anchor authority
```

No current dependency manifest/runtime implementation is authorized by this planning decision.

## 11. Current census / authority disposition

While this remains a future seam:

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
current GovernanceFeedback wire  unchanged
current B06 allowed_actions       unchanged
current B04 contract             unchanged
```

Current accepted census therefore remains:

```text
86 operations / 11 routes / 16 PermissionCode values
```

This document is planning work, not current durable Product/API authority until explicitly promoted after written operator ratification and any required bounded reopen.

## 12. Proof obligations for future promotion

A later promotion must be able to falsify at least:

```text
inline comment mutates governed Submission bytes
comment silently follows a newer DRAFT/current Revision
stable Document Discussion is used as governance-review storage
DRAFT EditorialComment is silently reused for Submission review
vendor-specific selection/comment id is the sole semantic anchor
RETURN maps old anchor onto changed DRAFT without proof
reviewer gains WorkingContent mutation merely from governance participation
suggestion/track-change is treated as accepted DRAFT mutation
B04 reconstructs old attempt by client History scan
framework availability creates new Authorization/lifecycle authority
```

## 13. Reopen / promotion triggers

Reopen this seam when one of these becomes real:

```text
MetalDocs chooses a production-capable active review mechanism/contract;
inline selected-range review becomes a committed Launch/current release requirement;
a customer/regulatory process requires anchored review evidence;
tracked-change/suggestion becomes a separately committed Product requirement;
measured usability proves unanchored feedback is inadequate for governance review;
a proven cross-format requirement justifies a broader annotation model.
```

## 14. Written ratification gate

Current disposition:

```text
operator intent / high-level direction     APPROVED IN CHAT
written B06-F2 candidate                   EXISTS
current API/Product promotion              NOT AUTHORIZED
P8 R2 for inline review                     NOT OPEN
B06 LOCK                                    BLOCKED ON WRITTEN ADJUDICATION
B07+                                        NOT OPEN
```

Next gate:

> Operator reviews this written B06-F2 candidate. If ratified, promote only the durable seam/reopen obligations needed to prevent future architectural dead-end, keep current Launch census/API/P8 controls unchanged, and then re-evaluate whether B06 R1 may be LOCKED without an R2.
