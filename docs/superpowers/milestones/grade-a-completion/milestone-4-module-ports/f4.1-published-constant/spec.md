# Feature F4.1 — Published constant (H-G fix)

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.1-published-constant`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumer:** `resolveTemplateVersionID` in `internal/modules/documents/application/service.go`.

**What it needs:** a `*string` value that equals the `"published"` wire string, set on
`controlleddocumentsdomain.TemplateVersionCandidate.Status`. It currently obtains this via a
hardcoded string literal (`overrideStatus := "published"`), which is an H-G finding because it
duplicates a domain-state constant owned by the `templates` module.

**Required shape after this feature:** the value still resolves to `"published"` at runtime, but
now obtained via `string(templatesdomain.VersionStatusPublished)` — the constant exported by the
owning module (`internal/modules/templates/domain/version.go:14`). The `templatesdomain` alias is
already present in `service.go` imports (line 17). No behavior change; only the source of truth
changes.

## Anchor (re-verified 2026-06-16)

- `internal/modules/documents/application/service.go:283` — `overrideStatus := "published"`
- `internal/modules/templates/domain/version.go:14` — `VersionStatusPublished VersionStatus = "published"`

## Non-goals

- No new constant in `documents` — must import from `templates`, not duplicate.
- No logic change in `resolveTemplateVersionID`.
- No other callsite touched.
- No new HTTP endpoint, no schema change, no frontend change.

## Validation Gate

1. `grep -n '"published"' internal/modules/documents/application/service.go` → returns 0 matches
   at the `overrideStatus` assignment (or only in comments/tests, each named in `evidence.md`).
2. A test asserting `string(templatesdomain.VersionStatusPublished) == "published"` passes
   (wire-value invariant — ensures the constant is the right one).
3. `go build ./...` clean.
4. `go test -count=1 ./internal/modules/documents/application/...` green.
