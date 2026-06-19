# Plan — F5.4 templates-routes-typed

## Files

**Modified:**
- `internal/modules/templates/delivery/http/routes_query.go` — swap 3 callsites:
  - `:64` list iteration: `toTemplateResponse(tpl)` → `toAPITemplateDTO(tpl)`; out slice becomes `[]templatesapi.TemplateDTO`; error path writes `codeTplInternalError`.
  - `:131` get-template: `toTemplateResponse(tpl)` → typed.
  - `:132` get-template: `toVersionResponse(latest)` → `toAPIVersionDTO(latest)`; same error path.
- `internal/modules/templates/delivery/http/routes_schema.go:65` — swap `toVersionResponse(v)` → typed.
- `internal/modules/templates/delivery/http/routes_lifecycle.go` — swap 4 callsites:
  - `:43` submit-for-review → typed.
  - `:92` review → typed.
  - `:140` approve → typed (`res.Version`).
  - `:178` archive → `toAPITemplateDTO(tpl)`.
- `internal/modules/templates/delivery/http/routes_create.go:44-95` — **delete** `toTemplateResponse` and `toVersionResponse`. Helper `timePtrRFC3339` becomes dead if no other caller; check before deleting.

## Steps

1. Swap each of the 8 callsites; preserve every outer envelope key (`data`, `version`, `template`, `next_draft`, etc.) and shape.
2. After all swaps, delete the two helpers. Re-run `grep "toTemplateResponse\|toVersionResponse\|timePtrRFC3339"` and prune anything unreferenced.
3. `go build ./...` clean.
4. `go test -count=1 ./internal/modules/templates/...` — every existing test still green. The wire keys are unchanged so map-decoding lifecycle/query tests keep passing.
5. Gate grep — Validation Gate steps 1–3.

## Test strategy

No new tests; the M1/F1.3 typed-shape pins (`routes_typed_response_test.go`) plus the existing
`routes_lifecycle_test.go` / `routes_query_test.go` map-key assertions already cover the
spec-declared wire shape. F5.4 is a structural refactor: same JSON keys, typed DTO values, no
behavior change.
