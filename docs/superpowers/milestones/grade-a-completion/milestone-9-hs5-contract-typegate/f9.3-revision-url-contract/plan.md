# Feature F9.3 — revision-url contract resolution

> **Milestone:** M9 — HS-5 contract type-gate  ·  **Folder:** `f9.3-revision-url-contract`
> **Status:** Done

## Source

- Milestone spec row: resolve the `signedRevisionURL` contract mismatch — runtime emits `200 + {url}`,
  spec declared `302`/no-body. Pick truth via the FE consumer (consumer-contract-first), then align.
- Governing-spec reference: mission.md §8 contract-api dimension.

## Plan

1. Read the FE consumer: `DocumentEditorPage.tsx` calls `apiFetch<{ url?: string }>(signedRevisionURL(...))`
   — it reads a JSON `{url}` body, not a redirect. **Consumer truth = 200 + {url}.**
2. Align the spec to runtime (not the reverse): add `RevisionUrlResponse {url: string, required}`; change
   `getDocumentRevisionUrl` from `302`/no-body to `200` + `$ref RevisionUrlResponse`.
3. Regen BE + FE codegen; route the handler through `documentsapi.RevisionUrlResponse` (replaces the
   `map[string]string` literal).
4. TDD: wire-contract lock `{url}`. Gates: build, tests, cilint, FE tsc.

## Execution notes

- Direction chosen by consumer contract, not convenience: the FE already depends on the JSON body, so the
  302 spec was the defect. No handler behavior change beyond typing the body.
- `wiki/modules/documents.md` revision-url row annotated to record the 302→200 alignment.
