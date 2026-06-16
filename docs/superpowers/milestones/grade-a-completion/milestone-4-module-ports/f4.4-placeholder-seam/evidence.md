# Feature F4.4 — Evidence

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.4-placeholder-seam`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md`.

## BOUNDARY DECISION (summary)

**Classification: (a) Legitimate port-typed dependency — no code change.**

`templatesdomain.Placeholder` (`templates/domain/schemas.go:61`) is a published, exported domain
type owned by the `templates` module. The `documents` module is the canonical consumer: it uses
the type across delivery, application, and repository layers to fill, validate, and seed
placeholder data. Introducing a `documents`-local mirror type would create split-brain (two
structs of the same shape, no single source of truth). The type is not an H-G site because H-G
requires either a hardcoded magic string OR a reach into another module's non-published internals
— this is neither.

Full rationale and rule in `spec.md § BOUNDARY DECISION`.

## Verification

| Gate | Command | Result | Real vs fixture |
|------|---------|--------|-----------------|
| G1: written boundary decision | `spec.md` BOUNDARY DECISION section | present — rule applied, rationale, auditor confirmation path | — |
| G2: build clean | `go build ./...` | no output (clean) | — |
| G3: tests green | `go test -count=1 ./internal/modules/documents/...` | all `ok` | fixture |

## C4 finding disposition

The re-audit C4 finding (`documents/repository` imports `templatesdomain.Placeholder`) is
**closed as a non-issue** after read-pass: the import is a legitimate published-type dependency,
not an H-G boundary breach. Recorded in `spec.md`; no grep command returns 0 (the import
remains and is correct).

## Bounded defers

None.
