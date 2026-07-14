# ADR 0050 — Computed-token catalog single source of truth in `render/domain`

> **Status:** Accepted
> **Last verified:** 2026-06-29
> **Date:** 2026-06-29
> **Scope:** Moving the computed/system-token catalog from a hand-maintained slice in the `templates` module to a single source of truth published from `render/domain`; surfacing `approval_date` to authors; the legal `templates → render/domain` cross-module edge; and the deferred backend-unified-endpoint promotion trigger.

---

## Context

### Two independently drifting hand-copies

The render module registered **8** built-in resolvers (in `internal/modules/render/resolvers/builtins.go`):
`doc_code`, `doc_title`, `revision_number`, `author`, `effective_date`, `approvers`,
`controlled_by_area`, **`approval_date``.

The templates module maintained a separate hand-written 7-key slice (previously in
`internal/modules/templates/delivery/http/routes_catalog.go`) that fed `GET /placeholder-catalog`
(the author-facing token palette).  `approval_date` was absent from that slice.

A second independent 7-key set existed in `internal/modules/templates/application/schema.go`
(`placeholderCatalogSet`, used by `ValidatePlaceholders`).  This set also missed `approval_date`
and was maintained entirely by hand.  An old test (`TestPlaceholderCatalog_Returns7Entries`)
hard-coded the count `want 7` and a literal key list without cross-checking the resolver registry —
enforcing the gap instead of catching it.

The net effect was that `approval_date` was fully resolvable at render time but structurally
invisible to authors: the palette never listed it, and `ValidatePlaceholders` would reject any
template schema that declared it.

### Root cause

No structural binding existed between the resolver registry (authoritative for what can be
rendered) and the two catalog copies (authoritative for what authors could discover and declare).
Keeping them in sync required manual discipline across two separate modules.

This is the CLAUDE.md AS-2 / global-maximum case: optimising inside the two-copy model would
lock in a local maximum.  The correct fix is to establish one source and derive all consumers
from it.

### Design spec

The full analysis and unit-of-work breakdown are in:
`docs/superpowers/specs/2026-06-29-token-catalog-single-source-and-palette-design.md`

---

## Decision

### D1 — `render/domain.ComputedCatalog()` is the single source of truth

A new `internal/modules/render/domain/computed_catalog.go` package publishes:

```go
type ComputedToken struct {
    Key           string
    Label         string // PT-BR, author-facing
    Description   string // PT-BR, author-facing
    AuthorVisible bool   // true → appears in the authoring palette
}

func ComputedCatalog() []ComputedToken { ... }
```

The function returns all **8** canonical computed tokens, all with `AuthorVisible: true`.  This is
the only place where the set of computed tokens and their author-facing metadata are declared.

### D2 — `templates` derives `/placeholder-catalog` from `render/domain`

The hand-maintained 7-key slice in `routes_catalog.go` is deleted.
`Handler.listPlaceholderCatalog` iterates `renderdomain.ComputedCatalog()`, filters for
`AuthorVisible == true`, and maps to `PlaceholderCatalogEntry{Key, Label, Description}`.
The route, contract, and `CapTemplateView` gate are unchanged.

### D3 — `ValidatePlaceholders` derives its computed set from `render/domain`

The hand-maintained `placeholderCatalogSet` in `schema.go` is replaced by a `sync.Once`-cached
accessor (`computedCatalogSet()`) that calls `renderdomain.ComputedCatalog()`.  Both the
palette-derivation in D2 and the validation gate in D3 now track the same source; drift between
them is structurally impossible.

### D4 — Bidirectional parity guard at the source

`internal/modules/render/resolvers/catalog_parity_test.go` asserts:
- every key registered in `resolvers.NewRegistry()` + `RegisterBuiltins` has a matching descriptor
  in `render/domain.ComputedCatalog()`, AND
- every descriptor in `ComputedCatalog()` has a matching registered resolver.

Drift between the resolver registry and the catalog is now impossible without a red test in the
owning module.

### D5 — `approval_date` is author-visible; degrades to the pending sentinel

`approval_date` is declared `AuthorVisible: true` in `ComputedCatalog()`.  The resolver
(`internal/modules/render/resolvers/approval_date.go`) returns the sentinel string
`"[aguardando aprovação]"` when the final approval date is zero (draft or pre-approval state),
mirroring the `approvers` resolver.  Authors can safely place `{approval_date}` in draft templates
without triggering a resolver error.

### D6 — `templates → render/domain` is the only new cross-module import edge

`templates` imports `render/domain` (not `render/resolvers`, not `render/fanout`).  The boundary
law enforced by `scripts/check-module-boundaries.ps1` permits a module to cross-import only
another module's `domain` layer.  This edge is therefore legal.  No `render → templates` reverse
edge is added.  This mirrors the existing `tokens/domain.DictionaryReader` symmetry: a module
publishes a catalog from its own domain; consumers import only that domain layer.

---

## Amends ADR 0008

ADR 0008 (`wiki/decisions/0008-placeholder-fixed-catalog.md`) established the fixed-catalog
model: all template placeholders are computed; the set is closed; validation enforces catalog
membership.  **ADR 0050 does not open or change the catalog contract.**  It relocates the
single source of truth from a hand-maintained slice in `templates` to a published Go package in
`render/domain`.  The catalog remains fixed and closed; adding a new computed token still requires
an explicit code change (to `ComputedCatalog()` and a new resolver).

ADR 0008's 7-token table is now historical: the catalog has 8 entries (`approval_date` added).

---

## Deferred promotion trigger (YAGNI)

The frontend composes the computed catalog (`GET /placeholder-catalog`) and the dictionary catalog
(`GET /tokens`) client-side, tagged by `kind`, in the `AvailableTokensPanel`.  Today that panel
is the only consumer of the merged view, and the merge is logic-free (concatenate + tag).

**Promotion trigger:** when a second backend consumer needs the merged list, OR when the merge
acquires non-trivial logic (cross-kind precedence, server-side filtering, collision surfacing at
read time), lift the client adapter into a backend composed contract — a single endpoint that
returns both catalogs tagged by `kind`.  The `Token { key, label, kind }` FE shape already matches
what that endpoint would return, so promotion is mechanical.  Until that trigger fires, the
client-side composition is YAGNI-correct.

---

## Consequences

**Positive:**
- Drift between the resolver registry and the author-facing catalog is structurally impossible.
  No manual synchronisation is needed when adding or changing a computed token.
- A single place to add a new computed token: add the descriptor to `ComputedCatalog()` and
  register a resolver.  The palette, the validator, and the parity guard all update automatically.
- `approval_date` is now discoverable and declarable by template authors.  The live defect (a
  renderable token invisible at authoring time) is structurally fixed, not patched.
- The new `render/domain` published surface mirrors `tokens/domain.DictionaryReader` in structure:
  each token kind (computed, dictionary) publishes its catalog from its owning module's domain
  layer, composed uniformly at the palette.

**Costs / trade-offs:**
- `templates` now imports `render/domain`.  This is a new legal cross-module edge that must be
  maintained.  The boundary guard (`check-module-boundaries.ps1`) makes an illegal widening
  (e.g., importing `render/resolvers`) a build failure.
- The `render` module now has a `domain` package where none existed before.  This is idiomatic
  (see `tokens/domain`), but it is new surface that must be kept stable.

---

## Related

- `wiki/decisions/0008-placeholder-fixed-catalog.md` — **amended** by this ADR (source of truth
  for the catalog relocates from a hand-maintained slice to `render/domain`; token count 7 → 8)
- `wiki/decisions/0029-user-display-name-reader-port.md` — reader-port pattern (iam/domain)
- `wiki/decisions/0030-template-version-state-port.md` — reader-port pattern (templates/domain)
- `wiki/decisions/0031-tenant-user-reader-port.md` — reader-port pattern (iam/domain)
- `wiki/decisions/0038-family-code-resolver-port.md` — reader-port pattern (taxonomy/domain)
- `wiki/decisions/0039-cross-module-base-table-read-boundary.md` — module boundary law;
  `templates → render/domain` is the legal form (domain layer only)
- `wiki/decisions/0048-tenant-token-dictionary.md` — SP-1 token dictionary module; symmetry:
  `tokens/domain` publishes `DictionaryReader`; `render/domain` publishes `ComputedCatalog`
- `wiki/decisions/0049-tenant-dictionary-token-substitution.md` — SP-2; `ValidatePlaceholders`
  D5 guard (reject dictionary names that equal a native/computed key) now uses the full 8-key set
  from `render/domain.ComputedCatalog()`, which includes `approval_date`
- `wiki/decisions/0022-authz-capability-coherence.md` — capability model; `CapTemplateView` gate
  on `/placeholder-catalog` is unchanged
- `docs/superpowers/specs/2026-06-29-token-catalog-single-source-and-palette-design.md` —
  design spec (authoritative intent)

---

## Key files

- `internal/modules/render/domain/computed_catalog.go` — the new single source of truth
- `internal/modules/render/resolvers/catalog_parity_test.go` — bidirectional parity guard
- `internal/modules/render/resolvers/approval_date.go` — `approval_date` resolver with sentinel
- `internal/modules/templates/delivery/http/routes_catalog.go` — derives catalog from `render/domain`
- `internal/modules/templates/application/schema.go` — `computedCatalogSet()` derived via `sync.Once`
- `scripts/check-module-boundaries.ps1:52` — boundary rule: `targetLayer` must be `"domain"`
