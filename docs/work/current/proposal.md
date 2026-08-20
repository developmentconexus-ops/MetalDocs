---
id: t8e-executable-wire-contract-proposal
kind: work
owner: architecture
summary: Active T8-E proposal containing only the unresolved executable-wire decisions after the accepted checkpoint.
---

# T8-E executable wire contract — active proposal

> Temporary / non-authoritative. Implementation remains blocked.

## Accepted baseline

Do not duplicate or reinvent the accepted layers. Read:

- `../../reference/t8e-checkpoint.md`
- `../../product/journeys.md`
- `../../decisions/api-operation-census.md`
- `../../decisions/forward-obligations.md`

Current application census: **78 operations**.

## Current question

What is the smallest closed executable ledger for all 78 operations that leaves no material wire choice to an implementation Writer while avoiding duplicate lifecycle/AuthZ authority and speculative capability?

## Remaining closure work

1. Exact request schemas for every operation.
2. Exact success response schemas.
3. Required / optional / nullable field matrix.
4. Closed enum vocabularies.
5. Exact success status and response-header matrix.
6. Operation-specific allowed RFC 9457 problem codes.
7. List filters and deterministic ordering.
8. Exact `allowed_actions` vocabularies by projection.
9. Request/body limits where materially applicable.
10. Measured raw/expanded document limits required by the accepted upload design.
11. Generated Go + TypeScript boundary feasibility.
12. Runtime request/response contract-conformance proof design.
13. Final Structural Inversion / subtractive / global-coherence pass.

## Laws

```text
accepted checkpoint decisions are not reopened by preference
no implementation code/schema/OpenAPI authoring in T8-E
no restored legacy wire becomes target authority
no generic response envelope/action API/error dialect/ACL/event bus
no field/enum/nullability/problem-code choice is deferred to Writers
unknown remains unknown until evidence closes it
prepare seams; do not build dormant capability
```

## Completion gate

```text
closed 78-operation executable ledger
+ exact schema/component closure
+ generation/conformance feasibility proof
+ subtractive/global-coherence pass
+ isolated final Fable review branch
+ Lead adjudication
+ explicit operator ratification
```

Only after T8-E ratification may T8-F open.