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

## Lead decision — ledger shape

Use **one closed operation ledger plus one closed component registry**. Do not create prose-per-operation mini-specifications and do not copy lifecycle/AuthZ predicates into the wire contract.

Each of the 78 operation rows must contain exactly the wire choices a Writer would otherwise have to invent:

```text
operation_id
method + path
request component(s)
request header profile
success status set
success body component or no-body
success header profile
allowed Problem.code set
pagination/filter/order profile when applicable
allowed_actions vocabulary reference when the response exposes it
request/body limit profile when materially applicable
```

The component registry owns exact reusable wire shapes once:

```text
shared scalar formats and bounded references
request objects
success projection objects
required / optional / nullable fields
closed enums / discriminators
pagination page shape
header profiles
Problem base + closed problem catalog
allowed_actions enums by projection
limit profiles
```

An operation row may reference a component/profile but may not rely on prose such as “appropriate fields”, “standard errors”, “usual headers”, “etc.”, or an implementation-defined enum. Conversely, a row does **not** restate business authorization, lifecycle transition law, transaction law, or owner eligibility when those are already owned by T1→T8-D authority.

This is the minimum sustainable closure because removing any ledger column leaves a material wire choice to the Writer, while copying semantic predicates into the ledger would create a second authority.

## Wire modeling laws

### Presence and nullability

```text
required   = member must be present on the wire
optional   = member may be absent
nullable   = explicit JSON null is a valid semantic value
```

Absence and `null` are never interchangeable by convention. `nullable: true` is permitted only where current Product/T1→T8-D semantics contain an explicit empty/unknown state that must be distinguishable on the wire. Otherwise use a required concrete value or optional absence according to the owning semantics.

Request update semantics must be explicit per request component. No PATCH/PUT field acquires implicit “null means clear” behavior. The accepted DRAFT PATCH remains the concrete precedent: omitted means unchanged and null has no implicit delete meaning.

### Closed vocabularies

Every semantic discriminator/action/state emitted or accepted by the application contract is a closed OpenAPI enum at Launch. No free-form string is used where current authority already defines a finite vocabulary.

Wire spelling is lower snake case even where an owning authority describes the same semantic value in uppercase prose. Wire normalization does not create a second semantic vocabulary.

### Response shape

Purpose-built success schemas remain the rule. There is no universal `{data, meta}` envelope and no generic action result.

Potentially unbounded collection responses use the accepted page shape:

```json
{
  "items": [],
  "page": {
    "next_cursor": null,
    "has_more": false
  }
}
```

`next_cursor` is the only baseline nullable pagination member: it is `null` when no next page exists. `items`, `page`, `next_cursor`, and `has_more` are present on every paginated success response.

### Header profiles

Define and reference profiles rather than duplicating header prose in 78 rows:

```text
JSON_NO_STORE
  Cache-Control: no-store

JSON_ETAG
  ETag: required strong opaque entity tag
  Cache-Control: no-store

JSON_ETAG_MUTATION
  ETag: required new strong opaque entity tag
  Cache-Control: no-store

EXACT_BYTES
  Content-Type
  Content-Length
  Content-Disposition: server-generated ASCII filename
  Content-Digest: SHA-256
  Cache-Control: private, no-store, no-transform
  Accept-Ranges: none
  X-Content-Type-Options: nosniff
```

`WWW-Authenticate` belongs to the 401 Problem response profile, not ordinary success responses. Unsafe-request CSRF and conditional/idempotency headers are request profiles, not response metadata.

### Problem closure

The accepted RFC 9457 catalog remains one global code → type/title/status authority. The operation ledger owns only the **allowed subset of catalog codes for that operation**.

No operation has a `default` response and no operation inherits an undocumented “common errors” bucket. Shared response components are allowed only when they preserve an explicit closed code set; tooling convenience may not widen the public contract.

`401`, `403`, `404`, `409`, `410`, `412`, `413`, `415`, `422`, `429`, `500`, and `503` therefore appear only on operations for which the corresponding failure is reachable under the accepted semantics. `405` remains HTTP method handling for the declared path surface and is not evidence of an extra application operation.

### Filters and ordering

There is no generic filter or sort DSL. Each list operation must name its exact accepted query members and one deterministic ordering. Cursor integrity binds the normalized filter/order tuple.

The already-ratified document catalog rule is preserved:

```text
q present:
exact code
→ code prefix
→ title prefix
→ title contains
→ code + stable id tie-break

q absent:
code + stable id
```

Other lists require their own smallest deterministic order from the owning semantic authority; do not copy document-search ranking by analogy.

### `allowed_actions`

`allowed_actions` is emitted only by projections whose owning journey names it as a UX hint. Each such projection references its own closed enum. There is no product-wide action vocabulary and no generic `/actions` operation.

The enum is a projection of canonical current authorization + Controlled Documents predicates; the ledger freezes only wire spellings. It must not encode roles, grants, candidate identities, or lifecycle rules as an independent policy matrix.

### Limits

Do not invent round-number limits to make the schema look complete.

- Pagination is already closed at default `20`, maximum `100`.
- Direct-upload capability lifetime is already bounded to at most `15 minutes`, with exact `expires_at` returned.
- Raw document bytes, structurally expanded DOCX bytes, ZIP entry count/depth, and any document-format-specific admission limits remain **Unknown pending the required measurement evidence**.
- JSON/string/list maxima are frozen only where Product/T1→T8-D semantics or concrete abuse/tooling evidence supplies a real bound. A runtime/framework default is not Product contract evidence.

A 413 catalog entry may exist before every operation uses it; operation rows may reference it only after a concrete application request-size ceiling is frozen for that request profile.

## Closed operation-census partition

The ledger must contain exactly the current authority's 78 operations and no more:

```text
Session/AuthN support               3
Organization                       26
Authorization                       4
Document Governance config         10
Controlled Documents / Work        34
Audit                               1
TOTAL                              78
```

`getUserProfile` and `getAreaLifecycle` are the two bounded read-symmetry operations already ratified in `api-operation-census.md`. Operation 79 is a material Product/T6 reopen.

## Ledger authoring order

Close the ledger in dependency order, not repository/module order:

```text
1. shared scalar/reference/problem/header/page components
2. exact read projections and their enums
3. exact command/request objects and presence/nullability
4. 78 operation rows: request → success → problems
5. list filters/order/cursor binding
6. projection-specific allowed_actions vocabularies
7. request/admission limits from evidence
8. generated Go + TypeScript feasibility probe
9. runtime conformance proof design
10. Structural Inversion / subtractive / global-coherence pass
```

This ordering lets command schemas reuse already-closed representations without making generated DTOs or handlers semantic authority.

## Generation feasibility target

The eventual OpenAPI must be generation-safe without weakening the contract:

```text
one full-spec generated Go wire/transport boundary
one generated TypeScript paths/components boundary
closed enums preserved as finite generated types
oneOf/discriminators used only for true closed semantic unions
no handwritten duplicate DTO authority
no generator-specific public fields or provider IDs
```

Feasibility is proven against the chosen generators with a disposable probe derived from the candidate contract. The probe may test generation/compile/type behavior; it is not Product implementation and does not authorize runtime work.

Failure to generate a valid accepted semantic union is evidence to adjust schema encoding, not evidence to widen the Product model.

## Runtime contract-conformance proof design

T8-E freezes the proof obligation, not the runtime implementation:

```text
request path:
raw HTTP request
→ central OpenAPI request validation
→ generated typed boundary
→ semantic handler

response path:
semantic result
→ generated typed response boundary
→ HTTP response
→ contract tests validate status + headers + body against OpenAPI
```

Required negative proof classes include at least:

```text
undeclared/malformed request member or header is rejected
missing required If-Match / Idempotency-Key / CSRF is rejected where applicable
weak/list/wildcard If-Match is rejected where forbidden
undeclared enum value is rejected
success cannot omit a required response member/header
operation cannot emit a Problem.code not declared for that operation
paginated response/cursor contract is exact
exact-byte response cannot silently become redirect/range/compressed/provider URL behavior
```

Do not add a generic production response-buffer validator merely to prove the contract; generated typed output plus contract tests is the accepted minimum.

## Still-open evidence / authority reads

The ledger structure itself is now decided. Closing its contents still requires bounded reads of implicated semantic authorities for exact fields/actions and one evidence activity for document limits.

Do **not** infer missing fields from legacy implementation. If a semantic authority does not distinguish a wire-relevant value required by an existing journey, classify that point as a bounded contradiction/reopen candidate rather than inventing it locally.

## Remaining closure work

1. Populate exact request/success components and presence/nullability.
2. Freeze all emitted/accepted closed enums.
3. Populate all 78 operation rows with exact success status/header and allowed Problem.code sets.
4. Freeze each list's exact filters and deterministic ordering.
5. Freeze exact projection-specific `allowed_actions` enums.
6. Measure and freeze raw/expanded document admission limits.
7. Run Go + TypeScript generation feasibility proof.
8. Freeze runtime request/response contract-conformance proof details.
9. Run final Structural Inversion / subtractive / global-coherence pass.

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