# Feature F7.3 — Search typed response (hand-rolled envelope; pre-codegen)

> **Milestone:** 7 — HS-2 Contract Completion  ·  **Folder:** `f7.3-search-typed`
> **Status:** Approved 2026-06-20 — code change may begin.
> **Approved before code:** 2026-06-20 / leandrotca.work — inherited from the M7 Phase-2 operator
> approval (commit `45a03fa6`). No new consumer contract beyond `milestone.md` F7.3.

## Interview record (B1.5)

| # | Question | Answer / source |
|---|----------|----------------|
| 1 | Is search codegen or pre-codegen? | **Pre-codegen** — no `internal/modules/search/api/` dir, no `api.gen.go`. ADR 0012 legacy posture: hand-rolled typed Go structs are the sanctioned pattern. A `SearchDocumentResponse` item mirror struct already exists at `handler.go:30`. |
| 2 | Which sites are response literals? | 1: `handler.go:134` — `WriteJSON(w, 200, map[string]any{"items": out})` where `out` is `[]SearchDocumentResponse`. |
| 3 | Which `map[string]any` are kept? | None. After the swap, search `handler.go` has 0 `map[string]any` of any kind. |
| 4 | What is the wire shape? | `{items: [SearchDocumentItem...]}` — OpenAPI `SearchDocumentsResponse` (`openapi.yaml:4921`, required `[items]`, `items` array of `SearchDocumentItem`). The item shape is already governed by the existing `SearchDocumentResponse` mirror struct. |
| 5 | Byte-identical risk? | `out` is `make([]SearchDocumentResponse, 0, len(items))` — always non-nil → marshals to `[]`, never `null`. Wrapping the same slice in a struct field tagged `json:"items"` yields the identical bytes `{"items":[...]}`. |
| 6 | HS-2 risk — codegen standup? | **No.** Hand-rolled struct only; no `cfg.yaml`, no `go generate`, no routing rewire. Within ADR 0012 legacy posture (operator's explicit M7 scope). |

## Consumer contract (FIRST — before any producer)

**Consumers:** FE search caller + the OpenAPI `SearchDocumentsResponse` schema (source of truth).

| Op | Path / method | Status | Body type emitted | Wire keys (unchanged) |
|----|---------------|--------|-------------------|------------------------|
| search documents | `GET /api/v1/search/documents` | 200 | `searchDocumentsResponse{Items []SearchDocumentResponse}` | `{items}` |

`items` serializes to the existing `SearchDocumentResponse` array JSON (`document_id, title, …, created_at`) —
unchanged, since the same `[]SearchDocumentResponse` value is emitted.

## What this feature implements

1. Define one unexported typed envelope struct in `search/delivery/http/handler.go` (after
   `SearchDocumentResponse`):
   - `searchDocumentsResponse{ Items []SearchDocumentResponse \`json:"items"\` }`
2. Swap `:134` emit → `searchDocumentsResponse{Items: out}`.

## Non-goals (mandatory)

- No change to wire keys/values (byte-identical, incl. empty → `{"items":[]}`).
- No change to the `SearchDocumentResponse` item shape, the SQL reader, or query mapping.
- No codegen pipeline standup, no `go generate`, no routing rewire (HS-2 boundary).
- No FE codegen regen (no OpenAPI change — the schema already declares this shape).

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|------------------|
| 1 | Zero `map[string]any` in `search/.../handler.go` | `grep -nE 'map\[string\]any' internal/modules/search/delivery/http/handler.go` → 0 hits (exit 1) | real (grep) — **red→green: 1 response literal → 0** |
| 2 | Envelope wire keys == OpenAPI `SearchDocumentsResponse` | NEW `TestSearchDocumentsResponse_WireContract` — marshal `searchDocumentsResponse`, assert key set `{items}` + item `document_id` round-trip | real |
| 3 | Empty result marshals to `{"items":[]}` not `null` | NEW `TestSearchDocumentsResponse_EmptyIsArrayNotNull` | real |
| 4 | Build + existing search tests green | `go build ./...`; `go test -count=1 ./internal/modules/search/...` → 0 FAIL | real |

## ADR needed?

- [x] No durable decision — F7.3 follows the existing ADR 0012 legacy posture (hand-rolled typed structs
  for pre-codegen modules); no new design.
