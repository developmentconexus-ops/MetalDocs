---
id: frontend-read-symmetry-precision
kind: authority
owner: architecture
summary: Operator-approved bounded T8-E read-model precision discovered by T8-F; adds disclosure-safe current routing references to DocumentOfficialView without changing the 78-operation census.
---

# T8-E bounded precision — frontend read symmetry

> **Operator approval:** 2026-08-21 during T8-F Fable Round-1 adjudication.

This is a bounded precision to the ratified T8-E wire authority. It changes no Product capability, route family, lifecycle, semantic owner, Permission, persistence authority, or operation count.

## Trigger

T8-F proved that two accepted stable routes/actions could not be resolved safely from the admitted Document Official read model after a fresh navigation:

```text
/documents/:document_id/work
  needs the identity of the current open Revision

Document Official obsolescence management
  needs the identity of the current ACTIVE ObsolescenceRequest when one exists and is disclosable
```

Using My Work or ascending Document History as a resolver would make projection/history scans part of current-resource identity resolution and would leave implementation to rediscover architecture.

## Precision

The effective `DocumentOfficialView` is the ratified T8-E shape plus exactly two optional derived routing references:

```text
OpenRevisionRoutingReference {
  revision: RevisionIdentity,
  state: OpenRevisionState
}

DocumentOfficialView {
  document: DocumentReference,
  document_type: DocumentTypeReference,
  area: AreaReference,
  responsible_owner: UserReference,
  status: DocumentOfficialStatus,
  official?: ReleasedRevisionView,
  open_revision?: OpenRevisionRoutingReference,
  active_obsolescence_request_id?: Uuid
}
```

The existing `official` presence laws remain unchanged.

## Disclosure and presence laws

These members are current derived read truth and never persisted pointers.

```text
open_revision
  source truth:
    the unique current Revision whose state is DRAFT or SUBMITTED

  present iff:
    such Revision exists
    AND current disclosure/Authorization permits the caller to receive working-context existence

  absent otherwise

active_obsolescence_request_id
  source truth:
    the unique current ObsolescenceRequest whose state is ACTIVE

  present iff:
    such request exists
    AND current disclosure/Authorization permits the caller to receive that request context

  absent otherwise
```

Absence never proves semantic non-existence to a caller that lacks disclosure authority.

These optional members do not grant access. Follow-up reads and every mutation still perform current canonical Authorization and lifecycle checks.

## Executability

No persistence reopen is required. Ratified T8-D already establishes:

```text
UNIQUE(document_id) WHERE Revision.state IN (DRAFT, SUBMITTED)
UNIQUE(document_id) WHERE ObsolescenceRequest.state = ACTIVE
```

No new durable current pointer is added to Document.

No T8-C contract class is added: `DocumentOfficial` is already an admitted Controlled Documents read projection. The application may compose these owner-derived references into operation 47 under the same disclosure rules used for the resulting view.

## Wire/census effect

```text
operation 47 getDocument response schema  -> precision only
operations added                          -> 0
operations removed                        -> 0
application census                        -> 78
operation 79                              -> absent
new Problem code                          -> 0
new header/profile                        -> 0
new Permission                            -> 0
```

This document supersedes only the `DocumentOfficialView` member set in `../architecture/wire-contract.md`; every other T8-E law remains ratified and unchanged.
