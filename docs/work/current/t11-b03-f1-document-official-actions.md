# T11 — B03-F1 Document Official action hints — candidate

> **Status:** CANDIDATE / OPERATOR ADJUDICATION REQUIRED.  
> **Scope:** bounded read-model precision for B03 only.  
> **Implementation:** BLOCKED.  
> **Method:** DevelopmentConexus Engineering Method + Frontend Product Experience Planning Method v2.2.

## 1. Trigger

B03 P8 R2 is operator-approved for content, interaction and layout, but the selected ficha contains management controls whose visibility must not be invented by the frontend.

Current accepted commands already exist:

```text
createDocumentRevision
replaceDocumentResponsibleOwner
createObsolescenceRequest
withdrawObsolescenceRequest
```

Current `DocumentOfficialView` exposes current Document/read truth but no complete server-derived statement of which of those command families are currently meaningful for this actor + Document.

The frontend must not reconstruct T3 Authorization or T2 lifecycle predicates.

## 2. Evidence

Accepted T8-F law already states:

```text
allowed_actions are UX hints only
commands recheck server truth
frontend owns no permission matrix
```

Accepted `GovernanceCaseView` already carries:

```text
allowed_actions: unique GovernanceCaseAction[]
```

filtered from the same current Authorization + Controlled Documents decisions used by commands.

The approved responsible-owner precision already proves `getDocument` may compose bounded current Authorization + Controlled Documents + Organization truth into purpose-built UX guidance without adding an operation or moving semantic ownership.

## 3. Root cause

Without a server-derived action projection, B03 has only bad choices:

```text
A. derive management eligibility from Permission/status/client state
   -> duplicate Authorization/lifecycle authority in frontend

B. add GET /documents/{id}/actions or similar
   -> screen-shaped API + unnecessary operation

C. render every management action and discover eligibility through 403/409
   -> error-as-capability-discovery, noisier UX, unnecessary requests

D. add bounded allowed_actions to existing DocumentOfficialView
   -> one current read lens, no new authority, commands still authoritative
```

## 4. Global Maximum decision candidate

Select **D**.

```text
DocumentOfficialAction =
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request

DocumentOfficialView.allowed_actions: unique DocumentOfficialAction[]
```

`allowed_actions` is required on every successfully disclosed `DocumentOfficialView` and may be empty.

Canonical order:

```text
create_revision
replace_responsible_owner
create_obsolescence_request
withdraw_obsolescence_request
```

The list is presentation guidance only. It is not an Authorization snapshot, reservation, capability token, concurrency guarantee, or command pre-approval.

## 5. Inclusion law

Each action is included only when the same current Authorization + Controlled Documents predicates used by the corresponding command say the command family is presently admissible before any still-unselected user input.

### create_revision

Include iff current command truth admits creating the next Revision for this actor + Document, including:

```text
document.edit in matching scope
+ authoring relationship predicate
+ current next-Revision eligibility
```

If current work already occupies the lifecycle slot or another current T2 predicate forbids creation, omit the hint. The actual command rechecks all current truth.

### replace_responsible_owner

Include iff current canonical Authorization admits `document.owner.manage` for the exact Document.

This aligns with the existing presence law for:

```text
responsible_owner_candidates?: UserReference[]
```

Target-specific D4 eligibility and current ResponsibleOwner ETag remain outside `allowed_actions` authority. The selected target and `If-Match` are rechecked by `replaceDocumentResponsibleOwner`.

### create_obsolescence_request

Include iff current command truth admits beginning obsolescence before reason input, including current:

```text
document.obsolete in matching scope
+ exact EFFECTIVE target eligibility
+ no open replacement conflict
+ no competing obsolescence
+ accepted governance-mode/route eligibility
```

The required human reason remains form input and is validated by the command.

### withdraw_obsolescence_request

Include iff current command truth admits withdrawal of the current active human-governed obsolescence request for this actor, including:

```text
document.obsolete in matching scope
+ active pre-completion request
+ actor is initiator OR actor has document.owner.manage
+ current T2 withdrawal eligibility
```

NoHumanApproval has no live withdrawal window and therefore cannot produce this hint.

## 6. Disclosure law

```text
getDocument not disclosable
-> no DocumentOfficialView at all

getDocument disclosable
-> allowed_actions always present
-> [] is valid
```

Absence of an action does not disclose *why* it is unavailable. No denial reason, missing Permission, lifecycle predicate, hidden request identity, or excluded target information is returned through this array.

`allowed_actions` does not expand source disclosure. It may only summarize action families already derived from current truth available to the application composition path.

## 7. Authority partition

Unchanged:

```text
Authorization       -> final ALLOW / default DENY
Controlled Documents -> Document relationship/lifecycle/governance predicates
Organization        -> User identity/eligibility where needed
application         -> purpose-built read composition/choreography
frontend            -> rendering only
```

Mechanism != authority:

```text
allowed_actions present
!= command will still succeed later
```

Every command rechecks current session, CSRF, Authorization, lifecycle, target eligibility, concurrency and all operation-specific preconditions at execution time.

## 8. Structural Inversion Test

Delete `allowed_actions` from the response while leaving commands intact:

```text
business correctness survives
Authorization survives
lifecycle correctness survives
```

Only UX discoverability degrades.

Therefore the projection is not new business authority.

## 9. YAGNI / effect census

```text
new Product capability             0
new semantic owner                 0
new Permission                     0
new stable SPA route               0
new application operation          0
new persistence object             0
new ETag domain                    0
new Idempotency-Key creation       0
new Problem code                   0
frontend Authorization evaluator   0
```

Current global census remains:

```text
application operations             86
Idempotency-Key creations          11
ETag read / mutation domains       13 / 13
exact-byte resources               4
```

## 10. Wire precision

Existing `getDocument` operation is unchanged in method/path/request/problems.

`DocumentOfficialView` gains:

```text
allowed_actions: unique DocumentOfficialAction[]
```

Recommended schema semantics:

```text
minItems: 0
uniqueItems: true
canonical server output order fixed as §4
```

No ETag is added to `DocumentOfficialView`; it remains a composed current read lens.

## 11. Proof strategy

Positive fixtures:

```text
actor eligible to create next Revision -> create_revision present
owner manager -> replace_responsible_owner present
eligible obsolescence initiator -> create_obsolescence_request present
eligible active-request initiator/manager -> withdraw_obsolescence_request present
multiple admitted actions -> unique canonical order
no admitted management action -> []
```

Causal negatives:

```text
Permission removed -> corresponding hint absent
responsible-owner relationship lost without owner.manage -> create_revision absent when relationship is required
open Revision blocks next Revision -> create_revision absent
open replacement / competing obsolescence -> create_obsolescence_request absent
non-initiator without owner.manage -> withdraw_obsolescence_request absent
NoHumanApproval synchronous completion -> withdraw hint absent
action returned earlier + state/AuthZ changes before command -> command still fails closed
responsible-owner target disabled after projection -> replacement still fails
```

No frontend test may prove Authorization by button visibility alone.

## 12. Adversarial challenge

Strongest counterarguments:

```text
"This duplicates command logic"
  -> reject if implemented as a second hand-written decision path.
  -> accepted only if both hint and command consume the same canonical predicates/shared equivalent.

"The hint can go stale"
  -> true and accepted; it is a non-authoritative read projection.
  -> commands recheck truth, and UI handles normal conflict/denial outcomes.

"A separate endpoint would be cleaner"
  -> no named independent consumer or semantic authority justifies it.
  -> existing Document Official read is the natural composition home.

"Return detailed reasons so UI can explain disabled buttons"
  -> YAGNI + disclosure expansion. Launch returns admitted actions only.
```

No surviving MATERIAL contradiction found in Lead review.

## 13. Review / promotion law

This candidate does not create or move an authority/trust boundary; it projects existing canonical decisions. A dedicated independent Fable round is therefore not structurally required solely for B03-F1. The already-required fresh independent challenge at T11/Whole-R10 closure remains unchanged.

If operator approves this candidate:

```text
promote bounded precision into durable T6/T8-E/T8-F current authority
-> reconcile executable read schema
-> mark B03-F1 CLOSED / OPERATOR-RATIFIED
-> B03 becomes eligible for operator-only final LOCK
```

Until operator approval, this file remains branch-only candidate evidence.
