---
id: document-official-actions-read-precision
kind: authority
owner: architecture
summary: Operator-ratified bounded T6/T8-E/T8-F precision that exposes current Document Official management action hints without creating frontend authorization authority or a new application operation.
---

# Bounded precision — Document Official action hints

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED PRECISION.  
> **Operator approval:** 2026-08-23 during T11 B03 closure.  
> **Method:** DevelopmentConexus Engineering Method v1.0.0 + Frontend Product Experience Planning Method v2.2.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Authority and scope

This page is the more-specific current authority for the bounded `DocumentOfficialView.allowed_actions` precision discovered during T11 B03 planning.

It refines only the current T6/T8-E/T8-F Document Official read/affordance clauses. Unrelated Product/API/frontend authority remains unchanged.

Until T11 merge-candidate consolidation absorbs this member into the owning executable wire/frontend authorities, consumers must read:

```text
base T6/T8-E/T8-F authority
+ this bounded precision
```

No new Product capability, semantic owner, Permission, route, operation, persistence object, ETag domain, Idempotency-Key creation or Problem code is created.

## 2. Trigger

B03 contains existing Document management commands:

```text
createDocumentRevision
replaceDocumentResponsibleOwner
createObsolescenceRequest
withdrawObsolescenceRequest
```

The frontend must not reconstruct current Authorization + Controlled Documents lifecycle/relationship predicates to decide which command families should be offered on the ficha.

Rejected substitutes:

```text
frontend permission/status evaluator
  REJECTED — duplicate Authorization/lifecycle authority

GET /documents/{id}/actions
  REJECTED — screen-shaped API + unnecessary operation

show every action and discover eligibility through 403/409
  REJECTED — error-as-capability-discovery and poorer UX
```

The smallest sustainable structure is a bounded action projection on the existing Document Official read.

## 3. Wire precision

Existing operation remains unchanged:

```text
GET /api/v1/documents/{document_id}
operationId: getDocument
```

Add the closed action vocabulary:

```text
DocumentOfficialAction =
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request
```

`DocumentOfficialView` gains exactly:

```text
allowed_actions: unique DocumentOfficialAction[]
```

Presence/order law:

```text
successfully disclosed DocumentOfficialView
-> allowed_actions always present
-> [] is valid

canonical output order:
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request
```

Recommended executable-schema constraints:

```text
minItems: 0
uniqueItems: true
```

No ETag is added to `DocumentOfficialView`; it remains a composed current read lens.

## 4. Meaning and authority law

`allowed_actions` is **UX guidance only**.

It is not:

```text
Authorization snapshot
capability token
reservation
concurrency guarantee
command pre-approval
denial-reason channel
```

Every corresponding command rechecks current session, CSRF, Authorization, lifecycle/relationship truth, target eligibility, concurrency and operation-specific preconditions at execution time.

Mechanism != authority:

```text
action present in prior read
!= command guaranteed to succeed later
```

## 5. Inclusion law

Each action is included only when the same canonical current Authorization + Controlled Documents predicates used by the corresponding command say that the command family is presently admissible **before any still-unselected user input**.

### `create_revision`

Include iff current command truth admits creating the next Revision for this actor + Document:

```text
document.edit in matching scope
+ authoring relationship predicate
+ current next-Revision eligibility
```

An existing current lifecycle occupant or any other canonical T2 predicate that prevents next-Revision creation omits the hint.

### `replace_responsible_owner`

Include iff current canonical Authorization returns ALLOW for:

```text
document.owner.manage
```

on the exact Document.

Target-specific D4 eligibility and ResponsibleOwner `If-Match` remain authoritative at mutation time.

This aligns with the current bounded `responsible_owner_candidates?: UserReference[]` presence law.

### `create_obsolescence_request`

Include iff current command truth admits initiating obsolescence before reason input:

```text
document.obsolete in matching scope
+ exact current EFFECTIVE target eligibility
+ no open replacement conflict
+ no competing obsolescence
+ accepted governance-mode/route eligibility
```

The mandatory reason remains user input and is validated by the command.

### `withdraw_obsolescence_request`

Include iff current command truth admits withdrawal of the current active human-governed request for this actor:

```text
document.obsolete in matching scope
+ active pre-completion obsolescence request
+ (actor is initiator OR actor has document.owner.manage in scope)
+ current T2 withdrawal eligibility
```

`NoHumanApproval` completes synchronously and has no live withdrawal window, therefore this hint is absent in that state.

## 6. Disclosure law

```text
getDocument not disclosable
-> no DocumentOfficialView

getDocument disclosable
-> allowed_actions always present
-> [] allowed
```

Absence of an action discloses no reason. The response does not expose:

```text
missing Permission
failed lifecycle predicate
hidden request identity
excluded target identity
internal denial explanation
```

The array may only summarize action families derived from current truth already available to the application composition path.

## 7. Authority partition

Unchanged:

```text
Authorization          -> final ALLOW / default DENY
Controlled Documents   -> Document relationship/lifecycle/governance predicates
Organization           -> User identity/eligibility where required
application            -> purpose-built read composition/choreography
frontend               -> rendering + interaction only
```

No second decision implementation is acceptable. Hint derivation and commands must consume the same canonical predicates or a provably shared equivalent.

## 8. Structural Inversion Test

Delete `allowed_actions` while leaving commands unchanged:

```text
business correctness survives
Authorization survives
lifecycle correctness survives
concurrency correctness survives
```

Only UX discoverability degrades.

Therefore the projection is not new business authority.

## 9. Census effect

Unchanged current totals:

```text
application operations             86
Idempotency-Key creations          11
ETag read / mutation domains       13 / 13
exact-byte resources               4
PermissionCode values              16
stable SPA routes                  11
semantic owners                    4 business + 2 supporting
```

Exact delta:

```text
operations added                   0
operations removed                 0
new Permission                     0
new route                          0
new owner                          0
new persistence state              0
new ETag domain                    0
new Problem code                   0
frontend Authorization evaluator   0
```

## 10. Proof obligations

Positive:

```text
eligible next-Revision actor          -> create_revision present
owner manager                         -> replace_responsible_owner present
eligible obsolescence initiator       -> create_obsolescence_request present
eligible request initiator/manager    -> withdraw_obsolescence_request present
multiple actions                      -> unique canonical order
no admitted management action         -> []
```

Causal negatives:

```text
Permission removed                    -> corresponding hint absent
authoring relationship removed        -> create_revision absent when required
open Revision blocks next Revision    -> create_revision absent
open replacement/competing request    -> create_obsolescence_request absent
non-initiator without owner.manage    -> withdraw hint absent
NoHumanApproval completion            -> withdraw hint absent
stale hint + later truth change       -> command still fails closed
target disabled after projection      -> owner replacement still fails closed
```

Button visibility never proves Authorization.

## 11. Reopen triggers

Reopen only on material evidence such as:

```text
a new real Document Official management command family
a need to explain disabled-action reasons that survives disclosure/YAGNI review
a consumer that needs an independently addressable action-capability resource
evidence that shared canonical predicates cannot safely serve both hint and command
```

Framework convenience, preference or desire for disabled buttons is not a reopen trigger.
