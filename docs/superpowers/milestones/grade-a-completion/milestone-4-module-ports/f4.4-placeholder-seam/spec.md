# Feature F4.4 — Placeholder seam boundary decision

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.4-placeholder-seam`
> **Status:** Approved 2026-06-16 — read-pass performed; decision recorded below.

## Consumer

`documents/repository.CreateDocumentTx` — seeds `document_placeholder_values` rows using
placeholder IDs from the template schema.

## Anchor (re-verified 2026-06-16)

- `internal/modules/documents/repository/repository.go:16` — `templatesdomain` import
- `internal/modules/documents/repository/repository.go:112` — `requiredPlaceholders []templatesdomain.Placeholder`

## Read-pass findings

`templatesdomain.Placeholder` is used throughout the **full `documents` stack**:

| Layer | File | Use |
|-------|------|-----|
| Delivery | `delivery/http/fillin_handler.go` | `GetFillInSchema` interface returns `[]templatesdomain.Placeholder` |
| Delivery | `delivery/http/placeholder_options_handler.go` | `LoadPlaceholderSchema` interface; type-switches on `PHSelect`, `PHUser` |
| Application | `application/fillin_service.go` | `LoadPlaceholderSchema`, `validateValue`, `findPlaceholder` — uses type extensively |
| Application | `application/service.go` | `var phs []templatesdomain.Placeholder`; `Repository` interface param |
| Application | `application/snapshot_service.go` | `ResolveTemplate` returns `[]templatesdomain.Placeholder` |
| Repository | `repository/repository.go` | `CreateDocumentTx` param; uses only `p.ID` to seed DB rows |

`templatesdomain.Placeholder` is defined in `internal/modules/templates/domain/schemas.go:61` —
an **exported type** in the `domain` package (not internal, not infrastructure-layer). The `PHType`
constants (`PHDate`, `PHNumber`, `PHSelect`, `PHUser`) are also published domain values.

## BOUNDARY DECISION: Legitimate port-typed dependency

**Classification:** (a) Legitimate port-typed dependency — **no code change required.**

**Rule applied:** `templates/domain.Placeholder` is a **published domain type** owned by the
`templates` module. The `documents` module is the canonical downstream consumer of template
schemas: it reads placeholder definitions to seed values, validate inputs, and render fill-in
forms. This is the standard "published language / shared kernel" pattern in module-boundary
architecture — the producing module (`templates`) publishes a domain type; the consuming module
(`documents`) depends on it through the owning module's `domain` package.

**Why this is not an H-G site:**

H-G class = "hardcoded foreign domain-state literal OR cross-module reach without a port".
`templatesdomain.Placeholder` is not a hardcoded literal (it's a type, not a magic string) and
it's not a cross-module reach without a port — the import goes through the owning module's
published `domain` package, which IS the port surface for value types. A `documents`-local
`Placeholder` struct that mirrors the templates one would be a split-brain (same shape, two
sources of truth), which is worse than the cross-import.

**Why introducing a port would be over-engineering:**
1. The type is consumed at every layer of `documents` — rewriting all six sites to use a
   `documents`-owned mirror type would add significant complexity with no module-safety benefit.
2. The `documents` module already legitimately imports `templatesdomain` for
   `VersionStatusPublished` (now the owning constant post-F4.1) — this is the same pattern.
3. The only field used in `repository.go` is `p.ID` (a primitive string). If the repository
   boundary were the only concern, the caller could pass `[]string` instead of
   `[]templatesdomain.Placeholder` — but that would mean diverging from the upstream application
   contract unnecessarily (the application already computed `[]Placeholder`).
4. The risk from this dependency is low: if `templates` renames or restructures `Placeholder`,
   `documents` compilation breaks immediately — no silent drift.

**Operator-confirmable:** the re-audit auditor may confirm this is not an H-G site by checking:
(a) `templatesdomain.Placeholder` is exported from `templates/domain/` (public API), and
(b) `documents` uses the type as a value object (not reaching into templates' private state).

## Non-goals (confirmed)

- No `documents`-local `Placeholder` mirror type.
- No `PlaceholderSeeder` port in iamdomain or templatesdomain.
- No restructuring of the templates ↔ documents relationship.

## Validation Gate

1. Written BOUNDARY DECISION (this document) with explicit rationale and the rule applied.
2. `go build ./...` clean (no code change — compilation already passes).
3. `go test -count=1 ./internal/modules/documents/...` green (unchanged).
