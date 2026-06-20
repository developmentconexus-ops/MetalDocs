# Feature F9.3 — Evidence

> **Milestone:** M9  ·  **Feature:** `f9.3-revision-url-contract`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` — `getDocumentRevisionUrl` returns `200 + RevisionUrlResponse {url}` (the FE
> consumer contract), resolving the prior `302`/no-body spec mismatch.

## What was implemented

- Consumer truth established: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
  reads `apiFetch<{ url?: string }>(...)` — a JSON `{url}` body. The `302` spec was the defect.
- OpenAPI: added `RevisionUrlResponse {url: string}` (required); `getDocumentRevisionUrl` `302`→`200`+`$ref`.
- Handler `signedRevisionURL` now emits `documentsapi.RevisionUrlResponse{Url: url}` via
  `httpresponse.WriteJSON(w, 200, ...)` (replaces `map[string]string`).
- BE + FE codegen regenerated. Commit `2e3c8a8b`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — lock | `go test ./internal/modules/documents/delivery/http/ -run WireContract` | `TestRevisionUrlResponse_WireContract` PASS — `{"url":"..."}`, key set `url` | real |
| Static (build) | `GOFLAGS=-mod=mod go build ./...` | exit 0 | — |
| FE codegen + types | `pnpm gen:api`; `npx tsc --noEmit -p tsconfig.build.json` | `RevisionUrlResponse` present at index.d.ts; tsc exit 0 | real |
| Gate | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` | exit 0 — no map literal at revision-url | real |
| Runtime proof | Handler returns `200 + {url}`; FE consumer reads `{url}` | spec now matches the live JSON body the FE already consumed | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Spec/runtime agree on `200 + {url}` | yes | OpenAPI diff + handler + FE consumer |
| Body typed (no map literal) | yes | lock + cilint |
| FE types regenerated & typecheck clean | yes | FE row |

## Review disposition

- Spec-compliance review: PASS — direction chosen by consumer contract, not by changing the handler to 302.
- Code-quality review: PASS — minimal, body-typing-only change.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
