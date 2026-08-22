---
id: responsible-owner-selection-read-precision
kind: authority
owner: architecture
summary: Records the operator-approved bounded T6/T8-E/T8-F precision that supplies least-privilege responsible-owner candidates on the Document Official lens without adding an application operation.
---

# Bounded precision — responsible-owner selection read

> **Operator approval:** 2026-08-22 during T11 Frontend Implementation Readiness F3 adjudication.

This records a bounded precision to the already-ratified T6/T8-E/T8-F authorities. It changes no Product capability, lifecycle, semantic owner, Permission, persistence authority, stable SPA route, or application-operation count.

Until the T11 closure candidate consolidates this member into the owning T6/T8-E/T8-F documents, this record is the more-specific authority for the responsible-owner candidate projection described below. The executable wire SSOT must contain the same precision before T11 integration; this file then remains provenance only.

## Trigger

T11 F3 Screen Contract `OFF-03 — Responsible owner management` proved that the accepted later owner-replacement journey was not safely implementable as a human interaction from the admitted wire alone.

Accepted replacement semantics already require:

```text
document.owner.manage
+ matching Document scope
+ target existing MetalDocs User
+ target same Company
+ target current eligibility = ENABLED
+ current ResponsibleOwner If-Match
→ replace responsible owner
```

The current wire exposes:

```text
getDocumentResponsibleOwner
  → current responsible owner + strong ETag

replaceDocumentResponsibleOwner
  ← target user_id
```

but no complete least-privilege candidate source for every valid existing-Document replacement state.

Rejected substitutes:

```text
Admin listUsers
  REJECTED — an area_manager may hold document.owner.manage without organization.manage;
             Product operation must not depend on privileged Organization administration UI.

DocumentCreationOptionsView.responsible_owner_candidates
  REJECTED as universal replacement source — creation options require creation-admitted active Area/DocumentType context,
             while later responsible-owner replacement has no accepted active-Area/active-DocumentType precondition.

manual opaque UUID entry
  REJECTED — transport-callable is not a coherent human selection journey and would leave accepted D4 eligibility undiscoverable.

new application operation
  REJECTED — no new Product capability exists; operation 79 remains absent.
```

## Precision

Existing operation 47 remains the sole Document Official read:

```text
GET /api/v1/documents/{document_id}
operationId: getDocument
```

`DocumentOfficialView` gains exactly one optional derived member:

```text
responsible_owner_candidates?: UserReference[]
```

Candidate semantics are exactly the ratified D4 eligibility law:

```text
existing MetalDocs User
+ same Company as the Document
+ current User eligibility = ENABLED
```

No target Role, Permission, provider claim/group, employment classification, or pre-existing document access is implied.

## Presence, disclosure and ordering law

```text
responsible_owner_candidates is present iff:
  getDocument is otherwise disclosable to the caller
  AND current canonical Authorization returns ALLOW for document.owner.manage on that exact Document

when present:
  contains the complete D4-eligible UserReference set
  ordered user_id ASC
  no silent truncation

when absent:
  caller must not infer whether candidates exist
  caller must not infer the reason for absence
```

`UserReference.display_name` retains its existing optionality; erased/missing profile presentation does not erase stable User identity or candidate eligibility.

The candidate list is UX/read guidance only. It grants no authority and does not reserve eligibility.

## Mutation and concurrency law

The authoritative mutation remains unchanged:

```text
GET responsible-owner
→ current ResponsibleOwnerView + strong ETag

PUT responsible-owner
If-Match: <current ETag>
body target user_id
```

`replaceDocumentResponsibleOwner` rechecks all current authority and D4 eligibility and preserves the accepted target-eligibility/offboarding serialization law.

Therefore:

```text
candidate present in an earlier getDocument response
≠ mutation eligibility guarantee
```

A target disabled/offboarded before the replacement linearizes is rejected fail-closed.

The new candidate list is **not** part of the ResponsibleOwner ETag concurrency domain. `getDocument` remains a non-ETag purpose-built composed read lens.

## Internal realization

No new semantic owner, persistence object, cross-owner SQL, or T8-C contract class is required.

Accepted T8-B/T8-C already allow application composition of:

```text
Controlled Documents current Document/access facts
+ Organization current User identity/eligibility truth
+ canonical Authorization decision
→ purpose-built Document Official read projection
```

Organization remains owner of User identity/eligibility; Controlled Documents remains owner of responsible-owner relationship; Authorization remains the only ALLOW/default-DENY authority.

## T6 / T8-E / T8-F effect

```text
T6
  later responsible-owner management on Document Official receives a bounded human-selectable candidate projection
  no new journey/capability/route family

T8-E
  DocumentOfficialView adds responsible_owner_candidates?: UserReference[]
  exact presence/disclosure/order/completeness fixtures are required
  operation 47 method/path/request/problems remain unchanged

T8-F
  OFF-03 Responsible Owner management consumes only this returned candidate set for selection
  current ResponsibleOwnerView/ETag still owns replacement concurrency
  no frontend Authorization engine or client eligibility authority is introduced
```

## Wire / census effect

```text
operation 47 getDocument response schema  → precision only
operations added                          → 0
operations removed                        → 0
application census                        → 78
operation 79                              → absent
new Problem code                          → 0
new request/header profile                → 0
new Permission                            → 0
new semantic owner                        → 0
new stable SPA route                      → 0
ETag read/mutation domains                → 13 / 13 unchanged
Idempotency-Key creations                 → exact 10 unchanged
exact-byte resources                      → exact 4 unchanged
```

## Proof obligation

Existing T9 evidence classes are sufficient. Future implementation proof must include at least:

```text
positive:
  caller with current document.owner.manage receives complete eligible candidates
  selected eligible target can replace under current ResponsibleOwner If-Match

causal negatives:
  caller without current document.owner.manage does not receive candidate existence
  disabled target is absent from a current candidate projection
  target disabled after projection but before mutation fails replacement
  stale ResponsibleOwner ETag cannot be bypassed by a fresh candidate projection
  Admin Organization privilege is not required for the area-manager replacement journey
```

A material scale finding showing complete same-Company ENABLED User candidates are not sustainable reopens the smallest T6/T8-E projection decision; it does not authorize silent truncation, generic search, or operation 79.
