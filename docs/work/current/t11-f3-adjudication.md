# T11 — F3 Screen Contract Adjudication

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** This records the operator adjudication that closes the sole F3 material blocker. It is not Product/API authority; the approved bounded precision is recorded durably in `../../decisions/responsible-owner-selection-read.md`.

## Result

```text
F3 Screen Contracts derived          36 / 36
READY before adjudication             35
BLOCKED before adjudication             1
finding                                F3-F01 Responsible Owner candidate read
operator adjudication                  APPROVED / 2026-08-22
approved precision                     T8-E-RO
operations added                       0
application operations                 78
operation 79                           absent
F3 final status                        36 / 36 READY
```

## F3-F01 closure

The `OFF-03 — Responsible owner management` Screen Contract is now READY under the operator-approved precision:

```text
getDocument → DocumentOfficialView
  responsible_owner_candidates?: UserReference[]

presence
  iff current canonical document.owner.manage = ALLOW for the exact Document

contents
  complete existing + same-Company + ENABLED UserReference set
  user_id ASC

replacement
  still uses getDocumentResponsibleOwner → ResponsibleOwnerView + ETag
  → replaceDocumentResponsibleOwner(target user_id, If-Match)
  → full current AuthZ + D4 eligibility recheck
```

The candidate projection is guidance only and is not part of the ResponsibleOwner ETag concurrency domain.

Rejected repairs remain rejected:

```text
Admin listUsers as required operational dependency
DocumentCreationOptions as universal later-owner selector
manual opaque UUID discovery
new application operation
operation 79
```

## Upstream consolidation obligation

Before T11 ratification/integration, the approved precision must be consolidated into the effective owning documents:

```text
docs/product/journeys.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
```

and the exact executable wire fixture/schema statements must agree with `docs/decisions/responsible-owner-selection-read.md`.

Until that consolidation, the precision decision is the more-specific authority for this single member.

## Next gate

F4 Navigation/Data Graph may proceed. Wireframes remain blocked until F4 is complete and any F4 material finding is adjudicated.
