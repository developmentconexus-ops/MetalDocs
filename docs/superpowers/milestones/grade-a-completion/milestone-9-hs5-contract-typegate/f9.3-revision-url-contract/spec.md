# F9.3 — signedRevisionURL contract alignment

> **Milestone:** M9  ·  **Status:** approved (operator Option-A proceed, 2026-06-20) — code may begin.

## Consumer contract (read from the consumer, before the producer)

`GET /api/v1/documents/{id}/revisions/{rid}/url`.

**Consumer truth = 200 + `{ "url": "<string>" }`.** The live FE consumer
(`frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:95`) does:

```ts
const signedPayload = await apiFetch<{ url?: string }>(signedRevisionURL(documentID, revisionID));
```

It **reads a JSON body**, it does **not** follow a 302 redirect. The OpenAPI declaration
(`openapi.yaml:2904`) says `302: redirect to signed URL` with no body — that is **stale and contradicts
both the runtime and the consumer** (runtime truth beats docs, CLAUDE.md). The handler (`handler.go:1105`)
emits `200 + map[string]string{"url": url}` — correct status/shape, wrong (untyped) Go type, and the spec
is wrong.

## What to implement (contract-first regen order)

1. **OpenAPI:** change `getDocumentRevisionUrl` responses from `302/no-body` to `200` with
   `$ref: '#/components/schemas/RevisionUrlResponse'`; add component
   `RevisionUrlResponse: { type: object, required: [url], properties: { url: { type: string } } }`.
2. **Regen BE codegen** (`go generate ./internal/modules/documents/...`) → `documentsapi.RevisionUrlResponse`.
3. **Handler:** emit `documentsapi.RevisionUrlResponse{Url: url}` at 200 (no `map[string]string`).
4. **Regen FE codegen** (`pnpm gen:api` in `frontend/apps/web`) → committed.
5. Correct the stale `wiki/modules/documents.md:258` "Aligned" note to reflect 200+typed body.

## Non-goals

- No change to the redirect-vs-body decision — consumer already expects a body; we are aligning the spec
  to runtime, not redesigning the auth/redirect flow.
- No change to how the signed URL is generated.

## Validation Gate

- OpenAPI declares 200 + `RevisionUrlResponse`; `go generate` produces the type with no other drift.
- Handler emits the generated type; no `map[string]string` at `handler.go:1105`.
- FE codegen regenerated & committed; `{ url?: string }` consumer still type-checks.
- Handler test: 200 + `{"url": "..."}` body (status + shape unchanged vs live).
- `go build ./...` + `go test ./internal/modules/documents/...` green.
- The widened `noresponsemap` (F9.4) reports this site clean.
