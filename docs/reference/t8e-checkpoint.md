---
id: t8e-executable-wire-contract
kind: checkpoint
owner: architecture
summary: Durable paused checkpoint preserving the accepted T8-E executable-wire design until active work resumes.
---

# T8-E executable wire contract checkpoint

> **PAUSED ACCEPTED CHECKPOINT.** Non-authoritative until the final T8-E candidate is ratified. Implementation remains blocked.

Current operation-census authority: `../decisions/api-operation-census.md` — **78 operations**.

## Question

What is the smallest exact OpenAPI 3.0.3 wire contract that realizes the ratified product journeys and persistence/concurrency laws without inventing new state, leaking mechanism shape, leaving Writers to choose missing fields/enums/errors, or turning the wire into a second lifecycle/AuthZ authority?

## Approved layer 1 — SSOT and generated boundaries

```text
api/openapi/v1/openapi.yaml is the sole /api/v1 application SSOT
browser OIDC routes remain outside the application SSOT
one generated Go transport/wire package from the full spec
one generated TypeScript paths/components boundary
no handwritten duplicate DTO authority
purpose-built response schemas; no universal data envelope
stable semantic operationIds are canonical wire operation identity
tags are documentation-only
central request-contract validation property is preserved
generated typed output path + contract-test response validation
no generic production response-buffer validation baseline
```

The bounded read-symmetry precision adding `getUserProfile` and `getAreaLifecycle` is now durably owned by `../decisions/api-operation-census.md`; T8-E consumes that 78-operation census and does not invent another route family.

## Approved layer 2 — concurrency, idempotency, CSRF, pagination

- strong opaque ETags; no raw database version/generation leakage;
- one concurrency domain per ETag-protected representation;
- `If-Match` accepts exactly one strong tag; no `*`, weak tags, or lists;
- missing/malformed required conditional headers are request errors; stale valid tags return 412;
- exact-current whole-replacement PUT retry may return success when already applied is provable and mutation/Audit are zero;
- stale DRAFT PATCH always returns 412;
- GET is the canonical ETag source;
- DRAFT PATCH returns a new ETag;
- UserProfile recreation uses `If-None-Match: *` when no current profile representation exists;
- authenticated application JSON uses `Cache-Control: no-store`.

Ten semantic creation POSTs use durable `Idempotency-Key` with a client-generated UUID and minimal replay-safe success bodies. No replay-indicator response header is exposed. Replay bodies exclude erasable UserProfile PII and free text.

`createSubmission` requires both:

```text
Idempotency-Key
If-Match of the exact DRAFT generation being frozen
```

Completed replay recognition happens before re-executing historical submit lifecycle preconditions.

Every unsafe `/api/v1` request requires `X-CSRF-Token`; `GET /api/v1/session` bootstraps the token and does not expose a role/permission snapshot.

Pagination uses a stateless opaque integrity-protected seek cursor bound to operation, filters, and ordering. Authorization is never snapshotted into the cursor; current session/AuthZ is rechecked on every page. Default limit 20, max 100. No offset, total count, generic sort DSL, server-side cursor state, or frozen multi-page snapshot baseline.

## Approved layer 3 — RFC 9457 problems

Problem responses use `application/problem+json` with required:

```text
type
title
status
detail
instance
code
trace_id
```

`errors[]`, when present, uses RFC 6901 pointers over `/path`, `/query`, `/header`, or `/body` and never echoes sensitive rejected values.

Problem type URI convention:

```text
https://errors.conexus.fun/{product_namespace}/{code}
```

For MetalDocs:

```text
product_namespace = metaldocs
```

Conexus owns the namespace convention; each product owns its own closed catalog. Runtime host/port is not part of durable problem identity.

One code maps to exactly one type, title, and HTTP status. The OpenAPI specification owns the catalog. No `default` response is allowed.

Selected status semantics:

```text
400 malformed/structurally invalid request
401 unauthenticated + WWW-Authenticate MetalDocsSession challenge
403 visible request permission denied or unsafe-request trust failure
404 absent or non-disclosable item resource
405 method not allowed
409 valid request conflicts with current lifecycle/state
410 known upload handle expired
412 supplied precondition evaluated false
413 MetalDocs HTTP request content too large
415 unsupported request media type
422 syntactically valid semantic instruction/content invalid
429 rate limit
500 unexpected internal/integrity failure
503 temporary dependency unavailability
```

No raw provider/storage/scanner/database errors escape.

## Approved layer 4 — upload, exact bytes, governance/effectivity views

Draft upload baseline:

```text
allocate temporary direct-upload capability
→ provider PUT is create-only (`If-None-Match: *` or equivalent)
→ bodyless completion endpoint
→ server independently reads bytes and derives exact descriptor
→ OPEN → READY
```

Client does not author SHA-256, provider identity, bucket/key, storage ETag, or semantic content identity. Allocation returns only opaque upload capability data and exact required headers. Capability lifetime is bounded to at most 15 minutes and exact `expires_at` is returned.

Completion is naturally idempotent and uses no durable Idempotency-Key. Server derives actual size and format. Malware proof remains mandatory at the immutable governed admission boundary, not on every mutable DRAFT debounce.

Exact-byte semantic resources return authenticated application-origin `200` bytes, not external redirects. Launch baseline rejects Range/206, 304, HEAD, compression/transformation, and provider URL exposure.

Exact-byte responses carry:

```text
Content-Type
Content-Length
Content-Disposition with server-generated ASCII filename
Content-Digest using SHA-256
Cache-Control: private, no-store, no-transform
Accept-Ranges: none
X-Content-Type-Options: nosniff
```

Semantic bytes missing/corrupt for an existing visible record fail closed as an internal integrity failure; provider outage is a dependency 503.

Shared wire references are bounded and purpose-built (`UserReference`, `DocumentReference`, `RevisionReference`, `ContentSummary`). Provider/mechanism IDs are not product truth.

Submission views expose orthogonal human and representation gates rather than queue/job lifecycle. A derived representation `attention_required` hint may surface terminal renderer attention without becoming Release authority.

Governance subject is a discriminated union of exactly:

```text
submission
obsolescence
```

Governance cases expose the exact subject, ordered steps, bounded feedback page, and canonical `allowed_actions`; they do not expose candidate user lists, group membership, or grant internals.

Governance decision request is a discriminated union:

```text
accept                → no reason
return_for_changes    → required nonblank reason
```

Release is an immutable system-owned effectivity view. Its representation is a closed union of source-only or required official PDF with an established OfficialRendition. Public renderer/River job state is not exposed.

## Accepted creation result rules

Semantic Idempotency-Key POST creations return `201` with minimal purpose-built results. Natural singleton PUT fact creation uses HTTP create/update semantics (`201` first creation; `200`/`204` exact repeat as appropriate).

Submission creation result uses:

```text
governance_pending
rendition_pending
released
```

and only navigation-safe IDs needed to continue the journey.

## Open measurement obligation

Do not guess maximum document sizes. Before final T8-E promotion, measure representative DOCX/PDF corpus and tooling constraints, then freeze:

```text
maximum raw bytes
maximum structurally expanded DOCX bytes
ZIP entry/depth limits where required
```

## Next design layer

Freeze the remaining executable ledger for all 78 operations:

1. exact request and success schemas;
2. required/nullable field matrix;
3. enum vocabulary;
4. success status and response-header matrix;
5. operation-specific allowed problem codes;
6. list filters and deterministic ordering;
7. exact `allowed_actions` enums by projection;
8. request-body size limits where applicable;
9. Go + TypeScript generation/conformance proof;
10. final subtractive/global-coherence pass.

Then run one final independent Fable challenge and obtain explicit operator ratification before opening T8-F.

## Resume law

When the repository reset is merged, copy this checkpoint into a fresh `docs/work/current/proposal.md` for active T8-E work. This checkpoint remains durable provenance until T8-E is ratified; it is never edited as if it were active work.