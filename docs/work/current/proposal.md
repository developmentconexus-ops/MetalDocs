---
id: t8e-executable-wire-contract-proposal
kind: work
owner: architecture
summary: Active proposal for the remaining T8-E executable-wire contract decisions after the accepted checkpoint.
---

# T8-E executable wire contract — active proposal

> Temporary / non-authoritative. Implementation remains blocked.

## Accepted baseline — do not reinvent

The accepted T8-E design through layer 4 is preserved at:

`../../reference/t8e-checkpoint.md`

It already owns, among other accepted decisions:

```text
one OpenAPI 3.0.3 application SSOT
one generated Go wire boundary
one generated TypeScript boundary
strong ETag / If-Match semantics
Idempotency-Key rules
session / CSRF rules
stateless cursor pagination
closed RFC 9457 Problem profile/catalog
Conexus multi-product Problem URI namespace
create-only direct upload + server-authoritative completion
exact-byte authenticated delivery
Submission / Governance / Obsolescence / Release / OfficialRendition wire shapes
```

The current application operation census is durably owned by:

`../../decisions/api-operation-census.md`

Current count: **78 operations**.

The remaining-stage decision inputs also consume:

`../../decisions/forward-obligations.md`

## Current question

What is the smallest closed executable ledger for all 78 operations that leaves no material wire choice to an implementation Writer while avoiding duplicate lifecycle/AuthZ authority and speculative capability?

## Remaining decisions to freeze

1. Exact request schemas for every operation.
2. Exact success response schemas.
3. Required / optional / nullable field matrix.
4. Closed enum vocabulary.
5. Exact success status and response-header matrix.
6. Operation-specific allowed RFC 9457 problem codes.
7. List filters and deterministic ordering.
8. Exact `allowed_actions` vocabularies by projection.
9. Request-body limits where materially applicable.
10. Measured raw/expanded document limits required by the accepted upload design.
11. Exact generated Go + TypeScript boundary feasibility and conformance proof.
12. Runtime request/response contract-conformance proof design.
13. Final subtractive / Structural Inversion / whole-T8-E coherence pass.

## Laws

```text
Do not reopen accepted layers by preference.
Do not invent implementation code or schemas in this PR.
Do not restore legacy OpenAPI or generated code as target authority.
Do not introduce a generic response envelope, generic action API, generic error dialect, ACL, event bus, or provider-specific product wire.
Do not leave fields/enums/nullability/problem codes for Writers to choose later.
Unknown remains unknown; measurement obligations must close before promotion when they affect the executable contract.
```

## Completion gate

T8-E may be promoted only after:

```text
complete 78-operation success/error/header ledger
+ exact schema/component closure
+ generation/conformance feasibility proof
+ subtractive/global-coherence pass
+ one final independent Fable challenge
+ Lead adjudication
+ explicit operator ratification
```

Only after T8-E is ratified may T8-F open. Product implementation remains blocked.