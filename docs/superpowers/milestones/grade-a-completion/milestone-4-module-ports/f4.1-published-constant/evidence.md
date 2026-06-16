# Feature F4.1 — Evidence

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.1-published-constant`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md`.

## Change

`internal/modules/documents/application/service.go:283`

```go
// Before
overrideStatus := "published"

// After
overrideStatus := string(templatesdomain.VersionStatusPublished)
```

`templatesdomain` alias already imported (line 17) — no new import.

## Verification

| Gate | Command | Result | Real vs fixture |
|------|---------|--------|-----------------|
| G1: `"published"` literal gone from overrideStatus site | `grep -n '"published"' internal/modules/documents/application/service.go` | 0 matches | — |
| G2: constant wire-value = `"published"` | `VersionStatusPublished` defined in `templates/domain/version.go:14` as `VersionStatus = "published"` | confirmed by read | — |
| G3: build clean | `go build ./...` | no output (clean) | — |
| G4: tests green | `go test -count=1 ./internal/modules/documents/application/...` | `ok metaldocs/internal/modules/documents/application 2.170s` | fixture |

## H-G finding disposition

C1 (`overrideStatus := "published"` hardcoded domain-state): **Closed**. The literal is replaced
by `string(templatesdomain.VersionStatusPublished)` — the constant owned by the `templates`
module. No `documents`-local duplicate constant introduced.

## Bounded defers

None.
