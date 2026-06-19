# Feature F5.1 — Templates published literal (H-G site #2)

> **Milestone:** 5 — HS-5 remediation  ·  **Feature:** `f5.1-templates-literal`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumer:** `TemplateVersionReader.IsPublished` in
`internal/modules/templates/infrastructure/template_version_reader.go`.

**What it needs:** to decide whether a scanned SQL `status` column equals the published
template-version state. It currently compares against a hardcoded string literal
(`status.String != "published"`), an H-G finding: it duplicates the domain-state constant owned
by the same module's `domain` package (`VersionStatusPublished`).

**Required shape after this feature:** the comparison still resolves identically at runtime, but
the published value is obtained via `string(templatesdomain.VersionStatusPublished)` — the
constant exported by `internal/modules/templates/domain/version.go:14`. Same module, so this is an
intra-module reference to the canonical source of truth, not a cross-module import. No behavior
change; only the source of truth changes. Mirrors the M4/F4.1 fix on
`documents/application/service.go:283`.

## Anchor (re-verified 2026-06-16)

- `internal/modules/templates/infrastructure/template_version_reader.go:44` —
  `if !status.Valid || status.String != "published" {`
- `internal/modules/templates/domain/version.go:14` —
  `VersionStatusPublished VersionStatus = "published"`
- The `infrastructure` package does **not** yet import the `templates/domain` package — this
  feature adds the import alias `templatesdomain`.

## Interview record

Contract is fully determined by `milestone.md` F5.1 row + the closed F4.1 precedent (identical
literal-to-constant swap, same wire value, same module family). No open questions — interview
waived under B1.5 (a contract that is unambiguous needs no discovery dialog). Recorded here so the
validator can confirm the waiver was deliberate, not skipped.

## Non-goals

- No new constant in `infrastructure` — reference the existing `domain` constant.
- No logic change in `IsPublished` or any other method.
- `GetTemplateVersionState` and other callsites untouched.
- No SQL change, no schema change, no HTTP/frontend change.

## Validation Gate

1. H-G literal grep:
   `grep -rn '"published"' --include="*.go" internal/modules/templates/infrastructure/ | grep -v "_test.go"`
   → 0 matches (or only comments, each named in `evidence.md`).
2. A test exercising `IsPublished` proves a `published` row → `(true, …)` and a non-published row
   → `(false, …)` — unchanged behavior, and a wire-value invariant test asserts
   `string(templatesdomain.VersionStatusPublished) == "published"`.
3. `go build ./...` clean.
4. `go test -count=1 ./internal/modules/templates/...` green.
