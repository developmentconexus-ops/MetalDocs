# Plan — F5.1 templates-literal

## Nature of change

Behavior-preserving literal→constant swap (refactor-class). The runtime comparison value is
identical before and after; there is no behavioral red→green. Honest TDD here is a **wire-value
invariant guard**: a test that pins `string(templatesdomain.VersionStatusPublished) == "published"`
so a future rename of the domain constant cannot silently desync from the SQL `status` column this
reader compares against. Labeled as a guard, not a behavioral red-green — per the evidence rules.

## Files touched

- `internal/modules/templates/infrastructure/template_version_reader.go` — add import alias
  `templatesdomain "metaldocs/internal/modules/templates/domain"`; replace `"published"` at line 44
  with `string(templatesdomain.VersionStatusPublished)`.
- `internal/modules/templates/infrastructure/template_version_reader_test.go` — add
  `TestIsPublishedComparesAgainstDomainConstant` invariant guard.

## Steps

1. **Red(ish):** add the invariant guard test asserting
   `string(templatesdomain.VersionStatusPublished) == "published"`. (Passes immediately — it pins
   the constant; its purpose is regression, documented as such.) Confirm it compiles and runs.
2. **Refactor:** add the import alias; swap the literal for the constant at line 44.
3. **Green:** `go test -count=1 ./internal/modules/templates/...`.
4. **Gate:** run the H-G literal grep → expect 0 non-test matches in `infrastructure/`.
5. **Build:** `go build ./...`.

## Test strategy

White-box test in `package infrastructure` (matches existing `_test.go`). No DB needed for the
invariant guard. Behavioral parity of `IsPublished` is already covered by the existing integration
test, which continues to pass unchanged.
