# Feature F5.4 — templates routes typed responses (Major #3)

> **Milestone:** 5 — HS-5 remediation  ·  **Feature:** `f5.4-templates-routes-typed`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumers:** every public templates HTTP route that currently materializes a template or version
into the response body through one of two `map[string]any` helpers in
`internal/modules/templates/delivery/http/routes_create.go`:

- `toTemplateResponse(*domain.Template) map[string]any` (line 44) — used at
  `routes_query.go:64,131`, `routes_lifecycle.go:178`.
- `toVersionResponse(*domain.TemplateVersion) map[string]any` (line 67) — used at
  `routes_query.go:132`, `routes_schema.go:65`, `routes_lifecycle.go:43,92,140`.

**What they need:** to materialize the DTO value in the wire body through the **strict-server
generated** type — `templatesapi.TemplateDTO` / `templatesapi.VersionDTO` — exactly as the canonical
M1/F1.3 typed mappers already do for the flat-shaped routes:

- `toAPITemplateDTO(*domain.Template) (templatesapi.TemplateDTO, error)` (`routes_mapping.go:122`)
- `toAPIVersionDTO(*domain.TemplateVersion) (templatesapi.VersionDTO, error)` (`routes_mapping.go:18`)

The two `map[string]any` helpers are **post-M1 leftover** (Minor #22 of the same re-audit) — the
typed equivalents are the canonical post-M1 path, already used by the flat routes
(`routes_create.go:36,41`, `routes_generated.go:64,69`, `routes_autosave.go:88`, `routes_query.go:160`).
Major #3 closes when no public-templates route materializes a template/version as an untyped map.

**Required shape after this feature:**

1. `toTemplateResponse` and `toVersionResponse` are **removed**.
2. Each of the 8 callsites uses the strict-typed equivalent and handles its error path with the
   existing `writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")`
   pattern (same fallback the flat routes already use on mapper failure).
3. The outer **envelope wrappers** (`map[string]any{"data": {...}, "meta": {...}}` etc.) on
   lifecycle/query endpoints **stay as they are**. The envelope is not a Major-#3 finding — the
   re-audit names only the helpers; the envelope is consistent with the existing pre-flat
   API surface (e.g., `routes_query.go:67`'s `{data:{templates:[...]}, meta:{...}}` shape).
   Typing the envelopes themselves would require declaring new strict-server response types for
   each of these legacy operations — explicit out-of-scope per M5 appetite (surgical only) and
   would risk HS-6.

## Interview record (B1.5)

| Question | Resolution | Source |
|----------|-----------|--------|
| Two helpers vs broader untyped-map sweep? | Re-audit Major #3 names only the two helpers (and their usage in query+lifecycle); envelopes flagged as Minor #22 leftover, not the Major. | re-audit lines 50, 81, 192 |
| Typed equivalents exist? | Yes — `toAPITemplateDTO`, `toAPIVersionDTO` (M1/F1.3). | `routes_mapping.go:14,122` |
| Behavioral parity? | Typed mappers emit the spec-declared shape (M1/F1.3). The map helpers emitted similar keys but as untyped Go map values (`v.Status` as `domain.VersionStatus` → string in JSON, etc.). Differences: typed mapper uses pointer-`omitempty` for nullable fields where the OpenAPI declares optionality; the map helpers used raw values. Per OpenAPI declaration the pointer-omitempty path IS correct. | `routes_mapping.go`, `api.gen.go` |
| Error path? | Typed mappers can fail on uuid parse. Use the same fallback `writeErr(...codeTplInternalError, "internal server error")` already used by flat routes. | `routes_create.go:36-39` |
| 8 callsites? | `routes_query.go:64,131,132` (list + with-latest), `routes_schema.go:65`, `routes_lifecycle.go:43,92,140,178` | grep |
| Tests that pin wire keys? | M1/F1.3 typed-response tests already pin the flat DTO shape (`routes_typed_response_test.go`). Existing lifecycle tests assert nested fields via map decoding — those keep passing because the wire keys are unchanged. | grep |

## Non-goals

- No envelope typing (out of M5 surgical scope; not a Major-#3 finding).
- No OpenAPI/spec change.
- No service-layer signature change (mappers consume `*domain.Template` / `*domain.TemplateVersion`
  exactly as the helpers did).
- No frontend change (wire keys unchanged).

## Validation Gate

1. **Helpers removed:** `grep -n "func toTemplateResponse\|func toVersionResponse"
   internal/modules/templates/delivery/http/routes_create.go` → 0 matches.
2. **No remaining callers:** `grep -rn "toTemplateResponse\|toVersionResponse" --include="*.go"
   internal/modules/templates/` → 0 matches (across both source and tests).
3. **No new public-route `map[string]any{...:`** value materialization for template/version DTOs in
   `routes_query.go` / `routes_schema.go` / `routes_lifecycle.go` — each former call now uses the
   typed DTO.
4. **Build:** `go build ./...` clean.
5. **Tests:** `go test -count=1 ./internal/modules/templates/...` all green (including the existing
   lifecycle/query route tests — these decode body via `map[string]any` and read the same keys, so
   they remain valid against the spec-declared shape).
