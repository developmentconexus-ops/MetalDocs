# Evidence — F5.1 templates-literal (H-G site #2)

> **Status:** CLOSED 2026-06-16 · behavior-preserving literal→constant swap.

## Change

| File | Change |
|------|--------|
| `internal/modules/templates/infrastructure/template_version_reader.go:8` | added import alias `templatesdomain "metaldocs/internal/modules/templates/domain"` |
| `internal/modules/templates/infrastructure/template_version_reader.go:45` | `status.String != "published"` → `status.String != string(templatesdomain.VersionStatusPublished)` |
| `internal/modules/templates/infrastructure/template_version_reader_test.go` | added `TestIsPublishedComparesAgainstDomainConstant` wire-value guard |

## TDD record

Refactor-class (behavior-preserving), so no behavioral red→green. The added test is a **wire-value
invariant guard** (labeled as such, not asserted as a behavioral catch): it pins
`string(templatesdomain.VersionStatusPublished) == "published"` so a future rename of the constant
cannot silently desync from the SQL `status` comparison. Behavioral parity of `IsPublished` remains
covered by the pre-existing `template_version_reader_integration_test.go` (unchanged, still passing).

## Validation Gate results (real output)

1. **H-G literal grep** —
   `grep -rn '"published"' --include="*.go" internal/modules/templates/infrastructure/ | grep -v "_test.go"`
   → **0 matches** (shell exit 1 = no match). Gate met.
2. **Package tests** — `go test -count=1 ./internal/modules/templates/infrastructure/` →
   `ok  metaldocs/internal/modules/templates/infrastructure  1.300s`. Guard test green.
3. **Build** — `go build ./...` → clean (`BUILD OK`).

## Fixture-vs-real

Guard test is a pure constant assertion (no DB). Behavioral coverage is the existing integration
test (real DB). No fixture stood in for a live path here.

## Defers

None. Single-site surgical change, fully closed.
