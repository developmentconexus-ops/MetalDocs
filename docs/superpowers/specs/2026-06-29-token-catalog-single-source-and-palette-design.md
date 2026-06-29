# Design — Computed-token catalog single source of truth + unified editor palette

**Date:** 2026-06-29
**Status:** Draft (awaiting operator review)
**Author:** brainstorming skill (with operator)
**Supersedes/affects:** the hand-maintained `placeholderCatalog` slice in `templates/delivery/http/routes_catalog.go`
**Predecessor analysis:** `docs/superpowers/analysis/2026-06-29-dictionary-tokens-editor-palette-system-impact.md` (Green gate — superseded in scope by the foundation finding below)

---

## 1. Problem (evidenced)

Two questions were raised by the operator:

1. Tenant **dictionary tokens** should appear in the template editor palette (today they don't; a used dictionary token is even flagged "Tokens não reconhecidos").
2. Why are **dictionary tokens** (tokens module, tenant CRUD rows) and **computed/system tokens** (templates, hard-coded catalog) different modules/structures when "they are the same"?

Investigating (2) surfaced a real foundation defect that question (1) would have built on top of:

**Computed tokens have two sources of truth, in two modules, already drifted.**

- **Source A — resolver registry** (`internal/modules/render/resolvers/builtins.go`), the authority that fills values at render, registers **8**: `DocCode, DocTitle, RevisionNumber, EffectiveDate, ControlledByArea, Author, Approvers, ApprovalDate`.
- **Source B — catalog slice** (`internal/modules/templates/delivery/http/routes_catalog.go:16`), what the author sees in the palette, lists **7** — `approval_date` is **absent**.
- The guard `TestPlaceholderCatalog_Returns7Entries` (`routes_catalog_test.go:12`) hard-codes `want 7` + a literal key list and never compares against the registry — it **enforces** the drift instead of catching it.

Result: `approval_date` is renderable but undiscoverable to authors. The two lists match (minus that gap) only by manual discipline; nothing structural keeps them in sync.

This is the CLAUDE.md global-maximum / AS-2 case: do not optimize inside the patch. Fix the foundation first, then build the palette on it.

## 2. The real "they are the same" — symmetry, not a merge

The two token kinds are **not** the same at resolution (computed = a function over document/system context; dictionary = a stored tenant constant) and must **not** share storage (a per-document function has no row; a tenant constant is not a function). Merging into one module/table (rejected option) breaks resolution and the system-vs-tenant governance split.

They *are* the same in **shape**: each kind should be **one authoritative store that drives both resolution and discovery, published via its module's domain port, and composed uniformly at the palette.**

| | Computed | Dictionary |
|---|---|---|
| Single source of truth | resolver **registry + descriptor** (render) | dictionary **rows** (tokens) |
| Drives resolution | registry → render | rows → pinned at creation (SP-2) |
| Drives discovery (catalog) | **derived from the same source** | derived from the same rows (`/tokens`) |
| Published via | `render/domain` port | `tokens/domain` port (`DictionaryReader`) |
| Composed at palette | uniformly, by `kind` | uniformly, by `kind` |

Today only **dictionary** has this property. Computed violates it (catalog ≠ registry). This design gives computed the same property, making the two architecturally identical in structure and different only in store — the operator's "same structure, one fixed one editable."

## 3. Architecture & module boundaries

**Boundary law** (`scripts/check-module-boundaries.ps1:52`): a module may cross-import **only another module's `domain` layer**. Therefore the computed catalog must be **published from `render/domain`** and consumed by `templates`. (`templates → render/resolvers` would be illegal; `templates → render/domain` is legal.) Precedent: `tokens/domain/port.go` publishes `DictionaryReader` for SP-2.

```
render/domain   ──publishes──▶ computed-token catalog  ──┐
                                                         ├─▶ templates/delivery → palette (kind-tagged)
tokens/domain   ──publishes──▶ DictionaryReader (rows) ──┘
```

`render` is the rightful owner of the computed-token catalog: it already defines and resolves these tokens; their author-facing label/description are metadata of that same domain concept. `templates` is a consumer that surfaces them for authoring. This mirrors how `templates` already consumes `tokens` for dictionary values.

## 4. Units of work

### Phase 1 — Foundation (backend, Go)

**U1. `render/domain` computed-token catalog (new package).**
Single source of truth: a static, ordered list of descriptors.
```go
// internal/modules/render/domain/computed_catalog.go (illustrative)
package domain

type ComputedToken struct {
    Key           string
    Label         string // PT-BR, author-facing
    Description   string // PT-BR, author-facing
    AuthorVisible bool   // true → appears in the authoring palette
}

// ComputedCatalog is the single source of truth for computed tokens.
func ComputedCatalog() []ComputedToken { ... } // 8 entries

// CatalogProvider is the port templates consumes (read-only).
type CatalogProvider interface {
    ComputedCatalog() []ComputedToken
}
```
All 8 tokens declared here, `AuthorVisible` set per token. `approval_date` → `AuthorVisible: true` (operator decision §6). Label/description for `approval_date`: `"Data de aprovação"` / `"Data de aprovação final do documento publicado."`

**U2. Bind resolvers to descriptors + same-module parity guard (render).**
The resolver registry binds to descriptor keys. New test in `render` asserts both directions:
- every registered resolver key has a `ComputedCatalog()` descriptor;
- every `AuthorVisible` descriptor has a registered resolver.
Drift is killed at the source, same module, trivially. (A non-`AuthorVisible` descriptor must still have a resolver; an internal resolver with no descriptor is a failure — every resolver needs a descriptor, visibility is the only knob.)

**U3. `templates` derives the catalog (delete the slice).**
`listPlaceholderCatalog` consumes `render/domain.CatalogProvider` (injected at app wiring, dependency-inverted — `templates` imports `render/domain` only), filters `AuthorVisible`, maps to the **unchanged** `PlaceholderCatalogResponse` DTO. Route, contract, and `CapTemplateView` gate unchanged. Delete `placeholderCatalog` and `catalogEntry` from `routes_catalog.go`. Replace `TestPlaceholderCatalog_Returns7Entries` with a test asserting the endpoint mirrors `render/domain`'s `AuthorVisible` set (no magic number).

**U4. App wiring.** `apps/api` constructs the render catalog provider and injects it into the templates handler (same pattern as the dictionary reader wiring).

### Phase 2 — Unified palette (frontend, on the fixed foundation)

**U5. Unified FE `Token` model + adapter.** `features/templates/tokens/`: `Token { key, label, kind: 'computed' | 'dictionary', source }`; a `useTokenCatalog()` hook composes the now-derived computed catalog (`usePlaceholderCatalogQuery`) + dictionary (`QK.tokens` / `GET /tokens`) into one `Token[]` (+ loading/error). Logic-free composition: concatenate, tag each with its constant `kind`. No truth duplicated.

**U6. `AvailableTokensPanel` consumes `Token[]`.** Two labeled sections by `kind`: computed = "Preenchido pelo sistema (seguro)"; dictionary = "Definido pela sua organização". `onInsert(key)` unchanged — insert stays uniform `{name}`.

**U7. Editor validation union.** `usedKeys` / `unknownTokens` classification draws its known-set from the unified `Token[]`; a dictionary token in the body is **known**, not flagged.

## 5. Out of scope (explicit)

- **No one-table / one-module merge** (rejected — breaks resolution + governance).
- **No backend unified `kind`-tagged endpoint yet.** The merged view has one consumer (the palette) and the merge is logic-free, so it is a consumer-side view. **Promotion trigger** (recorded in the ADR): the moment a second consumer needs the merged list, or the merge acquires logic (cross-kind precedence, server-side filtering, collisions surfaced at read), lift the FE adapter into a backend `render`/`tokens`-composed contract. The FE `Token` shape already matches that contract, so promotion is mechanical.
- **No change to dictionary storage, `/tokens`, SP-2 pinning, or render transport.**
- **No new capability, no migration.** `CapTemplateView` / `CapTokenView` already exist and are held by all authoring roles.

## 6. Decisions

- **`approval_date` = AuthorVisible: true** (operator, 2026-06-29). The 7-vs-8 gap was an accidental omission; it is now fixed — the palette will show **8** computed tokens. Validates the single-source model (it surfaced a live defect the hand-copy hid).

## 7. Boundary-compliance proof

- `templates → render/domain` only (domain layer) — passes `check-module-boundaries.ps1`.
- `render/resolvers → render/domain` — same module, allowed.
- No `templates → render/resolvers`, no `render → templates/*` cross-import added.
- Dependency inversion keeps `templates` depending on a domain port, wired in `apps/api` (mirrors `DictionaryReader`).

## 8. Test & QA plan

- **render:** unit test for `ComputedCatalog()` (8 entries, `approval_date` visible); bidirectional parity guard (U2). Canonical Go test framework.
- **templates:** endpoint test asserts `/templates/placeholder-catalog` mirrors `render/domain` `AuthorVisible` set (replaces the magic-`7` test); existing `CapTemplateView` test retained.
- **frontend:** Vitest/RTL for `useTokenCatalog()` (composition, kinds), `AvailableTokensPanel` (two sections), editor classification (dictionary token = known). Note local vitest is blocked by node_modules junction drift — runtime preview is primary FE evidence.
- **runtime verify (Phase 2):** author sees 8 computed + dictionary section; a dictionary token in the body is not flagged; `GET /tokens` + `/placeholder-catalog` both 200; console clean. Screenshots as evidence.
- **boundary CI:** `check-module-boundaries.ps1` green.
- **QA gates:** `go build ./...`, `go test ./...`, FE typecheck/build.

## 9. Docs / ADR

- **ADR (new):** "Computed-token catalog single source of truth in `render/domain`." Records: render/domain ownership; `templates→render/domain` consume edge; `approval_date` visibility; supersedes the hand-maintained slice; the deferred backend-unified-endpoint promotion trigger (§5).
- **Wiki:** update `wiki/modules/render.md` (new domain catalog + parity guard), `wiki/modules/templates.md` (catalog now derived, palette sources two catalogs), cross-link `wiki/modules/tokens.md §1` (symmetry realized). Bump `Last verified`. REQ IDs: REQ-AUTHZ (gates), REQ-TEN (tenant-scoped dictionary read), module-boundary REQ.

## 10. Phasing & sequencing

1. **Phase 1 (backend foundation)** U1→U2→U3→U4, then ADR + render/templates wiki. Ships independently; no FE change required to be correct (it just fixes the drift + exposes `approval_date`).
2. **Phase 2 (palette)** U5→U6→U7 on the fixed catalog, then runtime verify + templates/tokens wiki palette note.

Each phase is independently verifiable and commit-worthy.

## 11. Risks / tradeoffs

- **New `render/domain` package** where render currently has none — small, follows the modular pattern; the `tokens/domain` precedent shows it's idiomatic.
- **`approval_date` exposure:** confirm the resolver tolerates draft/unpublished state (no approval instance) gracefully; if it can error pre-approval, the palette still lists it but the spec's implementation plan must verify render returns a safe placeholder (parity with `approvers` → "[aguardando aprovação]"). Carried into the plan as a verification item.
- **Scope grew** from "FE-only palette" to "backend foundation + palette." Accepted by the operator (global maximum over local patch).
