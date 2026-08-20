---
id: api-operation-census
kind: authority
owner: architecture
summary: Owns the bounded T6 precision that establishes the current 78-operation /api/v1 application census.
---

# API operation census

The current Launch `/api/v1` application census contains **78 operations**.

`docs/product/journeys.md` §29 remains the semantic journey authority for the original 76-operation census. The following bounded read-symmetry precision was operator-approved during T8-E and is now recorded durably here:

```text
GET /api/v1/users/{user_id}/profile
operationId: getUserProfile

GET /api/v1/areas/{area_id}/lifecycle
operationId: getAreaLifecycle
```

## Why these two reads exist

They do not add a capability, semantic owner, product workspace, or new route family. They provide canonical resource representations for the already-ratified conditional whole-replacement surfaces:

```text
getUserProfile       → replaceUserProfile
getAreaLifecycle     → replaceAreaLifecycle
```

Each GET is the canonical strong-ETag source for its concurrency domain.

## Authority law

For the operation census only:

```text
original journeys census 76
+ bounded read precision   2
= current census          78
```

This page supersedes the **count and omission of these two GET operations only** in `docs/product/journeys.md` §29. All other journey meaning and operation-family restrictions remain owned by `docs/product/journeys.md`.

Any additional application operation requires either unchanged semantic normalization permitted by the journeys authority or an explicit bounded reopen of the owning product/API decision.

T8-E must use this 78-operation census.