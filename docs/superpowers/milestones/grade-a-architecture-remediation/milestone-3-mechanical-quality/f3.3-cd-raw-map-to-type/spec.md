# Feature F3.3 — CD AtomicCreate raw-map → generated response type

> **Milestone:** 3 · **Feature:** `f3.3-cd-raw-map-to-type`
> **Skill:** `metaldocs-backend-api` · **Class:** H-D-adjacent (producer not honoring declared optional contract)
> **Approved for code:** 2026-06-14 (self-approval recorded below; see "Contract decision" — the only
> wire delta is a drift-*fix* toward the already-declared OpenAPI contract, no OpenAPI/FE regen).

## Consumer contract (read from the consumer first)

The AtomicCreate `201` response is declared in OpenAPI and consumed by the frontend.

- **Declared producer contract (OpenAPI → generated Go):**
  `controlleddocuments/api/api.gen.go`
  - `AtomicCreateResponse { ControlledDocument ControlledDocument json:"controlled_document"; Document DocumentRef json:"document" }` (`:134`)
  - api `ControlledDocument` marks the optionals **`,omitempty`**: `DepartmentCode` (`:144`),
    `OverrideTemplateVersionId` (`:146`), `SequenceNum` (`:150`). `Id openapi_types.UUID json:"id"`.
  - api `DocumentRef { ContentHash string json:"content_hash"; Id openapi_types.UUID json:"id" }` (`:201`).
- **FE consumer:**
  `frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts:31-44`
  `createControlledDocumentAtomic(...)` returns `apiFetch<AtomicCreateResponse>(...)` where
  `AtomicCreateResponse = components["schemas"]["AtomicCreateResponse"]` — i.e. the FE **already**
  types the response as the generated contract, with the optional fields typed `field?: T`
  (present-or-absent), not `field: T | null`.

**Required shape (the contract the consumer already expects):** the handler emits
`controlleddocumentsapi.AtomicCreateResponse` — optionals **omitted when absent**, exactly as the
generated type and the FE type declare.

## Current producer (the drift)

`controlleddocuments/delivery/http/routes.go:123` emits a raw `map[string]any` of the **domain**
types:

```go
httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{
    "controlled_document": res.ControlledDocument,  // domain type — json tags have NO omitempty
    "document":           res.DocumentRef,
})
```

The domain `ControlledDocument` (`domain/controlled_document.go:18-33`) tags optionals **without**
`omitempty`, so an absent `department_code` / `override_template_version_id` / `sequence_num` is
serialized as `"...":null` — a value the declared/optional contract says should be **omitted**. The
domain `DocumentRef` already has `json:"-"` on its internal fields, so the `document` object is
already on-contract (`{id, content_hash}`) — no change there.

## What to implement

Replace the raw map at `routes.go:123` with `controlleddocumentsapi.AtomicCreateResponse`, built from
the **existing** domain→api mappers already used by Get/List:
`controlledDocumentResponse(*res.ControlledDocument)` and `documentRefResponse(res.DocumentRef)`.
On a mapper error, mirror the canonical Get pattern
(`httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")`).

## Contract decision (consumer-contract-first; the one nuance)

This is **not byte-identical** to the prior wire — and that is the point. The only delta:

| Absent optional | Before (raw map / domain) | After (typed / api) |
|---|---|---|
| `department_code` | `"department_code": null` | key omitted |
| `override_template_version_id` | `"...": null` | key omitted |
| `sequence_num` | `"sequence_num": null` | key omitted |

`null` → omitted **moves the producer onto the already-declared optional contract** (OpenAPI marks
these optional → generated `,omitempty`). No OpenAPI edit, no FE type regen: the FE already types
these as optional (`?`), so `undefined` (omitted) is *more* type-accurate than the current `null`.
Present optionals and all required fields are unchanged. This is therefore a drift-*fix*, not a
contract change — **no FE ripple, HS-2 does not trip.** (The milestone.md F3.3 row's "byte-identical"
wording predates this analysis and is corrected to "emits the declared `AtomicCreateResponse`; the
only wire delta is `null`→omitted on absent optionals, which the FE's optional-typed fields already
expect.")

## Non-goals

- No OpenAPI/spec edit; no FE type regen; no FE code change.
- No change to required fields, present-optional values, or the `document` object.
- No touching the request side, validation, error envelopes, or any other handler.
- No new mapper — reuse `controlledDocumentResponse` / `documentRefResponse` verbatim.

## Validation Gate

| # | Acceptance | Proof command / check |
|---|------------|------------------------|
| G1 | New TDD test asserts the 201 body unmarshals into `controlleddocumentsapi.AtomicCreateResponse` **and** that absent optionals are omitted (no `"department_code"` / `"override_template_version_id"` / `"sequence_num"` keys when nil). Red before impl, green after. | `go test ./internal/modules/controlleddocuments/delivery/http/ -run TestAtomicCreate_UsesGeneratedResponse -count=1` |
| G2 | No `map[string]any` at the AtomicCreate 201 site. | `grep -n "map\[string\]any" routes.go` → not at `:123` |
| G3 | Existing CD contract tests stay green (incl. `TestAtomicCreate_ForwardsGeneratedOnlyFields`, which now exercises the typed path via a UUID-valid fixture). | `go test ./internal/modules/controlleddocuments/... -count=1` |
| G4 | Build + codegen clean. | `go build ./...`; `go generate ./internal/modules/controlleddocuments/api/...` (no diff) |
| G5 | Spec unchanged (lint still clean — no OpenAPI edit). | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` |
| G6 | FE consumer unaffected (no regen needed; optional-typed fields tolerate omitted). | read `controlledDocuments.ts:31-44` — types already `AtomicCreateResponse`, optionals `?` |

## Interview record

No operator interview needed — the consumer contract is unambiguous (declared in OpenAPI + already
consumed by the FE as the generated type). The single nuance (`null`→omitted) is resolved by the
declared contract itself (optionals are `,omitempty`), recorded above for the validator and a brief
operator note at the HS-1 gate.
