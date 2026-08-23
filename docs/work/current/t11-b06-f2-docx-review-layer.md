# T11 — B06-F2 DOCX Review Layer seam — ratification record

> **Status:** CLOSED / OPERATOR-RATIFIED / PROMOTED AS FUTURE-SEAM AUTHORITY.  
> **Parent block:** B06 — Governance Case.  
> **Ratified:** 2026-08-23.  
> **Durable authority:** `../../decisions/governance-review-layer-seam.md`.  
> **Forward obligation:** `GOV-12` in `../../decisions/forward-obligations.md`.  
> **Implementation:** BLOCKED.  
> **Current Product/API/P8 delta:** ZERO.

## 1. Trigger

Operator review of the approved B06 P8 R1 surfaced a real future product-evolution need for Word-like DOCX selected-range review without mutating the governed Submission.

The finding was treated as a cross-layer architecture question rather than as a request to add speculative editor buttons.

## 2. Ratified direction

```text
exact governed Submission remains immutable
stable Document Discussion != inline governance review
DRAFT EditorialComment remains separately deferred
future selected-range review binds to the exact immutable reviewed snapshot
current unanchored GovernanceFeedback remains valid
tracked changes / suggestions require a separate semantic promotion
vendor/editor ids never become MetalDocs semantic authority
RETURN leaves review anchors with the old immutable Submission
old anchors never blindly remap onto changed DRAFT bytes
future B04 remediation needs explicit server-authored review-context identity
```

The durable decision intentionally does not freeze the future anchor encoding or create a generic annotation platform.

## 3. Provider / editor boundary

Editor/provider evidence may inform future mechanism selection, but no framework defines Product truth.

```text
vendor may implement selection/rendering/review mechanics
vendor may not own Submission identity
vendor may not own GovernanceFeedback meaning
vendor may not own GovernanceDecision meaning
vendor may not own lifecycle or Authorization
vendor-specific comment/selection ids may not be the sole semantic anchor
```

The previously studied editor/license options remain evidence only and are not promoted by this ratification.

## 4. RETURN_FOR_CHANGES consequence

```text
review immutable Submission S1
-> RETURN_FOR_CHANGES
-> S1 + Decision + feedback remain immutable
-> same Revision returns to DRAFT
-> author edits mutable WorkingContent
```

Any future anchor remains attached to S1. It is never overlaid onto the changed DRAFT merely through text matching, editor heuristics or stale browser state.

Future cross-snapshot mapping, if ever required, needs its own proven reconciliation model.

## 5. Current Launch consequence

B06-F2 does **not** promote active inline review.

Therefore current truth remains:

```text
GovernanceFeedback wire         unchanged
allowed_actions                 accept | return_for_changes | add_feedback
application operations          86
stable SPA routes               11
PermissionCode values           16
B04 contract                    unchanged
B06 P8 R1 controls              unchanged
inline-review P8 R2             not required
```

No disabled/coming-soon comment, reply, resolve, suggest or tracked-change controls are added.

## 6. Re-evaluation of B06 P8

The operator had already operated and approved B06 P8 R1 before raising this future-capability finding.

Ratification changes no present-tense visible behavior. Therefore:

```text
P8 R1                     remains valid / operator-approved
P8 R2 inline review       NOT REQUIRED for current Launch
material visible finding  0
```

The next legitimate frontend-method gate is explicit operator `LOCK` of B06 R1, not another prototype iteration.

## 7. Closure

```text
B06-F2 written candidate                 RATIFIED
future seam authority                    PROMOTED
durable current Product/API capability   NOT PROMOTED
GOV-12 forward obligation                RECORDED
current census                           UNCHANGED
B06                                      LOCK READY / awaiting explicit operator LOCK
B07+                                     NOT OPEN
```
