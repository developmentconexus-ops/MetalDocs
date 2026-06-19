# Evidence — F5.4 templates-routes-typed (Major #3)

> **Status:** CLOSED 2026-06-16 · 8 callsites swapped from `map[string]any` helpers to
> strict-server typed DTOs; both legacy helpers deleted.

## Change

| File | Change |
|------|--------|
| `internal/modules/templates/delivery/http/routes_query.go` | List iteration (`:63-71`) builds `[]templatesapi.TemplateDTO`; `getTemplate` (`:135-149`) writes `tplDTO`+`latestDTO` from `toAPITemplateDTO`/`toAPIVersionDTO`; error path returns `codeTplInternalError` 500. |
| `internal/modules/templates/delivery/http/routes_schema.go` | `updateSchemas` (`:63-72`) writes `toAPIVersionDTO(v)` instead of `toVersionResponse(v)`. |
| `internal/modules/templates/delivery/http/routes_lifecycle.go` | 4 callsites swapped: `submitForReview` (`:41-50`), `review` (`:95-104`), `approve` (`:139-152`, on `res.Version`), `archiveTemplate` (`:181-190`, via `toAPITemplateDTO`). Each adds the standard error fallback. |
| `internal/modules/templates/delivery/http/routes_create.go` | **Deleted** `toTemplateResponse`, `toVersionResponse`, and the now-orphan `timePtrRFC3339` helper. Imports trimmed (`time`, `domain` removed — unused after deletion). |
| `internal/modules/templates/delivery/http/routes_lifecycle_test.go` · `routes_autosave_test.go` | Drive-by repair on pre-existing fixtures: version ID field + fake-repo map keys upgraded from `"ver-1"` to a valid UUID (`22222222-…`) so `toAPIVersionDTO`'s `uuid.Parse` succeeds. No behavior assertions changed. |

The outer envelope wrappers (`map[string]any{"data": {...}}`, `meta` etc.) are intentionally
preserved — not in F5.4 scope per spec §Non-goals (envelope typing flagged Minor #22, not the
Major-#3 finding; would risk HS-6 if widened).

## TDD record

**Refactor-class — structural lift, behavior parity.** The wire keys exposed by the typed mappers
(`toAPITemplateDTO`/`toAPIVersionDTO`) are the M1/F1.3 strict-server shape, already pinned by
`routes_typed_response_test.go` and `routes_typed_response_f53_test.go`. Existing lifecycle/query
tests decode the response body through `map[string]any` and assert on the canonical keys
(`version.status`, `data.template.id`, etc.) — those keys remain unchanged so the existing
assertions are the durable behavioral guard. No new "red→green" test added; the contract is now
compile-time enforced by the strict-server struct.

Honest labeling: the **fixture repair** in the two test files (UUID-shaped IDs) is NOT a behavior
change — the fake's `r.versions[v.ID]` lookup pattern required the map key and ID to match. Pre-F5.4
the helpers accepted any string; post-F5.4 the typed mapper validates `uuid.Parse(v.ID)`. Repair was
the minimal change to keep the existing assertions valid against the new compile-time contract.

## Validation Gate results (real output)

1. **Helpers removed** — `grep -n "func toTemplateResponse\|func toVersionResponse"
   internal/modules/templates/delivery/http/routes_create.go` → **0 matches**.
2. **No remaining callers** — `grep -rn "toTemplateResponse\|toVersionResponse\|timePtrRFC3339"
   --include="*.go" internal/modules/templates/` → **0 matches** (source AND tests).
3. **No `map[string]any{...}` materializing template/version DTOs** on the public routes — each
   former call now uses the typed DTO. The remaining `map[string]any` literals in `routes_query.go`,
   `routes_schema.go`, `routes_lifecycle.go` are envelope wrappers (`{"data": {...}, "meta": {...}}`)
   only — out of F5.4 scope per spec.
4. **Build** — `go build ./...` → `BUILD OK`.
5. **Module suite** — `go test -count=1 ./internal/modules/templates/...` → all packages `ok`
   (api/application/delivery/http/domain/infrastructure/repository).
6. **Whole-repo regression** — `go test -count=1 ./...` → 0 FAIL.

## Fixture-vs-real

Pure unit + route-level `httptest` against the existing fake repo + real `application.Service`.
F5.4 is wire-shape only — no live SQL involved. No fixture stood in for a live path.

## Defers

None. All 8 cited callsites closed; both helpers deleted. **Mention-don't-fix:** the outer
`map[string]any{"data": ...}` envelopes on lifecycle/query/audit routes remain (Minor #22 of the
re-audit) — out of M5 surgical scope and explicitly carved out by F5.4 spec §Non-goals.
