# SP-2 — Tenant Dictionary Token Substitution (creation-time)

**Date:** 2026-06-28
**Program:** MetalDocs tenant-token program (SP-1 backend dictionary → **SP-2 substitution** → SP-3 UI)
**Status:** Design approved (operator), pending spec review
**Upstream gate:** [SP-2 system-impact analysis](../analysis/2026-06-28-sp2-render-token-substitution-system-impact.md) — verdict 🟡 Yellow
**Supersedes assumption:** the SP-2 brief framed this as a render/freeze-time *merge* in the `render` module. Runtime truth + operator intent relocate it to **creation-time pinning** spanning `tokens` + `templates` + `documents`/`controlleddocuments`.

---

## 1. Problem & intent

SP-1 delivered the `tokens` module: a per-tenant, capability-governed `name → value` dictionary (CRUD only; no substitution). SP-2 makes those tokens usable in documents.

**Operator intent (locked in brainstorming):**
- A dictionary token (name→value) is created/managed on its **own dedicated screen** = the `tokens` module CRUD (SP-1 backend exists; SP-3 builds the screen).
- A **template references** a dictionary token by name ("choose whatever"). The template never owns the value.
- When a **document is created from that template, it already comes with the token rendered** — the value is bound at creation, not deferred to freeze.
- Dictionary tokens are a **template-construction** concept; a document author cannot introduce new ones.
- Creating a token inline from the template screen is a **UX convenience (SP-3 UI)** — it reuses `POST /tokens`; no new SP-2 backend.

## 2. Locked decisions

| # | Decision | Choice |
|---|----------|--------|
| D1 | Binding time | **Pinned at document creation** (Arch-A: declared reference + creation-time resolution). Not freeze-time merge. |
| D2 | How a template references a token | **Declared in `placeholder_schema`** as a new placeholder source `PHDictionary` whose `Name` = the dictionary entry name. A *reference*, carrying no value. |
| D3 | Collision precedence | **Computed/native wins** — but enforced by *prevention*, not render-time precedence. |
| D4 | Collision prevention (primary) | **At dictionary-token creation** (`POST /tokens`): reject any name equal to an already-registered native/computed name. Operator's chosen mechanism. |
| D5 | Collision prevention (defense-in-depth) | **At template schema-save** (`ValidatePlaceholders`): a `PHDictionary` reference name must not be a native name. |
| D6 | Template-save existence validation | **No** dictionary-existence lookup at save (keeps `templates` decoupled from `tokens`). Format + not-native only. Existence enforced at creation; SP-3 UI does a live `GET /tokens` check. |
| D7 | Missing referenced token at creation | **Block creation (422)** — a controlled doc cannot "start rendered" with an unknown token. |
| D8 | Reproducibility | Value pinned at creation → immune to later dictionary edits; covered by existing `values_hash`. No render-path change. |
| D9 | Write-guard advisory on `/tokens` (TD-1 item 3) | **Deferred to SP-3** (UI edge). |

## 3. Token taxonomy (the unified kernel)

A document's substitution map (`placeholder_values`, keyed by placeholder ID; rendered into the body by eigenpal keyed by `name`) already has two value **sources**. SP-2 adds a third. This is the global-maximum move: extend the existing kernel, do not fork a parallel path.

| Source | Declared where | `Name` rule | Resolved when | Resolved from |
|--------|----------------|-------------|---------------|---------------|
| `user` | template placeholder_schema (no `Name`, ID-keyed) | — | document fill-in | author input |
| `computed` | template placeholder_schema (`Name` ∈ native catalog, `Type=computed`, `ResolverKey=Name`) | must be a native name | freeze (`pinValidateAndHash`) | `render/resolvers` registry |
| **`dictionary` (NEW)** | template placeholder_schema (`Name` = dict entry name, `Type=dictionary`) | must **not** be a native name; must match dict-name format | **document creation** | `tokens.DictionaryReader` |

All three flow through the **one** existing freeze → materialize → eigenpal substitution path unchanged. The `source` column on `placeholder_values` is the discriminator (`fillin_repository.go:22`).

## 4. Architecture & data flow

```
TEMPLATE AUTHORING                       DOC CREATION (from template)              FREEZE / MATERIALIZE
──────────────────                       ────────────────────────────             ────────────────────
templates.UpdateSchemas:                 controlleddocuments.Create:               (UNCHANGED)
  schema gains a PHDictionary             1. OFF-TX (H-PRE-1): resolve template     pinned dictionary values
  reference {type:"dictionary",              snapshot + for each PHDictionary ph:   are already in placeholder_values
   name:"company_name"}                      DictionaryReader.GetByName(tenant,     (source="dictionary"); the existing
        │                                     name) → value; missing → 422          eigenpal pass substitutes
        ▼                                  2. ATOMIC TX (cloneIntoTx →               {company_name} in the body docx
  ValidatePlaceholders:                       CreateDocumentTx): seed
    name ∉ native registry  ◄──┐              placeholder_values(source=
    (D5)                       │              "dictionary", value) alongside
                               │              existing required-value seeding
  tokens.CreateToken (D4): ────┘                   │
    reject name ∈ native registry                  ▼
    (primary guard)                        doc "comes rendered" — value frozen at
                                           creation, immune to later dict edits (D8)
```

**Value is the source of truth in the dictionary; the document captures a pinned copy at creation.** A later `PUT /tokens` never alters an already-created document.

## 5. Components (per module)

### 5.1 `tokens` — write-time reserved-name guard (D4)
- **New consumer port** in `internal/modules/tokens/application/ports.go`: `ReservedNames` (e.g. `IsReserved(name string) bool`). Keeps `tokens` a leaf — **no `tokens → render` import**.
- **Composition root** (`apps/api/cmd/metaldocs-api/main.go`): inject an impl backed by `render/resolvers.Registry.Known()` (the authoritative 8 native keys — `templates`' static `placeholderCatalogSet` is missing `approval_date`; we use the registry, not that set, and do **not** fix the pre-existing 7-vs-8 gap here).
- **`CreateToken`** (`internal/modules/tokens/delivery/http/handler.go` / application `Service.Create`): reject `name ∈ reserved` → 422 with a new `domain.ErrReservedName` / problem code `reserved_name`. Name is immutable post-create, so Create is the only guard point (Update cannot change name).
- **No migration, no new route, no new capability.**

### 5.2 `templates` — `PHDictionary` schema support (D2, D5)
- **Domain** (`internal/modules/templates/domain/schemas.go:33`): add `PHDictionary PlaceholderType = "dictionary"`.
- **Validation** (`internal/modules/templates/application/schema.go:103` `ValidatePlaceholders`): for `Type == PHDictionary` —
  - require non-empty `Name`, matching the existing `placeholderNameRe` format;
  - `Name` must **not** be a native name (check against `s.resolvers.Known()` — already injected at `schema.go:47-56`);
  - `ResolverKey` must be nil; `Computed` must be false;
  - participates in the existing duplicate-name check (`seenNames`).
  - The existing computed-name rule (`Name` ∈ catalog ⇒ computed) is untouched for non-dictionary placeholders.
- **No migration** — `placeholder_schema` is a JSON column; `PHDictionary` is a new enum value within it.

### 5.3 `documents` + `controlleddocuments` — creation-time resolution & pinning (D1, D7, D8)
- **`controlleddocuments.Create`** (`internal/modules/controlleddocuments/application/service.go`, near the existing OFF-TX template resolution at `:327`): after the template snapshot/schema is resolved off-tx, iterate `PHDictionary` placeholders and resolve each via the injected `tokens.DictionaryReader.GetByName(ctx, tenantID, name)` — **off-tx (H-PRE-1: never an authz-recording read inside the atomic creation tx)**, under the creating actor's identity. Build `{placeholderID → value}`. A missing token → fail closed **422** (D7).
- **`documents.cloneIntoTx` / `Repository.CreateDocumentTx`** (`internal/modules/documents/application/service.go:262`, `repository/repository.go:246`): accept the resolved dictionary values and seed them into `placeholder_values` with `Source = "dictionary"` (keyed by placeholder ID) **inside** the creation tx, joining the existing required-value seeding.
- **Wiring** (`main.go`): inject `tokens.Module.Reader` (a `DictionaryReader`) into the creation service.
- **Authz:** the dictionary read is `token.view`-gated inside `DictionaryReader`; the creating actor must hold `token.view` (broadly granted in SP-1). Fail closed if not.
- **Multi-tenant:** `DictionaryReader` predicates on the creation tenant; no cross-tenant read.

### 5.4 `render` — no change
Pinned `source="dictionary"` values are already in `placeholder_values` at freeze; the existing `freeze_service.go` materialize path maps them into `FanoutRequest.PlaceholderValues` by name and eigenpal substitutes them. **Zero render-module edits.**

## 6. Contract & data

- **OpenAPI:** no route added/changed. `PHDictionary` is a value inside the existing template-schema DTO; if the generated template-schema DTO enumerates placeholder types, regenerate after adding the enum value (contract-first: edit `api/openapi/v1/openapi.yaml` schema enum first, then `go generate`). `POST /tokens` shape is unchanged (the reserved-name rejection reuses the existing 422 problem envelope).
- **Migration:** **none expected.** `PHDictionary` and `source="dictionary"` are JSON / free-string values. **Plan must verify** whether `placeholder_values.source` carries a CHECK constraint enumerating sources; if so, a forward migration adds `'dictionary'` to it. No new table.
- **Destructive change:** none. All additive.

## 7. Error handling

| Condition | Where | Result |
|-----------|-------|--------|
| Dictionary token name = native name | `POST /tokens` | 422 `reserved_name` (D4) |
| `PHDictionary` reference name = native name | template schema-save | 422 (template validation, D5) |
| `PHDictionary` reference name bad format / empty | template schema-save | 422 (existing validation style) |
| Referenced dictionary token missing at creation | doc creation | 422, blocks creation (D7) |
| Creating actor lacks `token.view` | doc creation | fail closed (403/500, clear error) |
| Cross-tenant dictionary read attempt | doc creation | impossible — reader is tenant-predicated |

## 8. ADR & docs

- **New ADR 0049** — "Tenant dictionary token substitution & reserved-name guard." Fulfils ADR 0048's deferred SP-2 contract (closes TD-1), records: (a) creation-time pinning as the binding model, (b) the `PHDictionary` placeholder source extending the substitution kernel, (c) the backend reserved-name guard (a deliberate extension of ADR 0048's "Go does hygiene only" — the native catalog, not token grammar). Cite REQ-AUTHZ-1/2 (authorized freeze + `token.view`-gated read), REQ-MT-1 (tenant-scoped read).
- **Wiki:** `wiki/modules/tokens.md` (flip SP-2 status; fix the §3 diagram's render-fanout framing to the documents-assembles/render-transports reality), `wiki/modules/tokens-tech-debt.md` (close TD-1), `wiki/modules/templates.md` (`PHDictionary`), `wiki/concepts/placeholders.md` (third value source). Refresh `Last verified`.

## 9. Test & QA plan

Canonical `testdb` integration factory; `//go:build integration`; R1–R4 discipline. Unit tests for pure validation.

- **`tokens`** (integration): `POST /tokens` with a native name → 422 `reserved_name`; non-native name → 201.
- **`templates`** (unit): `ValidatePlaceholders` accepts a valid `PHDictionary` reference; rejects native-name, bad-format, empty-name, and a `PHDictionary` with a `ResolverKey`.
- **`documents`/`controlleddocuments`** (integration): create-from-template pins the dictionary value (`source="dictionary"`, correct value); **a dictionary edit after creation does not change the created document's pinned value** (D8 reproducibility); missing referenced token → 422 (D7); two-tenant isolation (tenant A's creation never reads tenant B's dictionary).
- **`render`** (integration): a pinned `source="dictionary"` value substitutes into the body docx at freeze (end-to-end through the unchanged path).

**Gates (per task + at close):** `go build ./...`; `go test ./...`; `go test -tags=integration ./internal/modules/{tokens,templates,documents,render}/...`; `.\scripts\check-system-runnable.ps1` (PowerShell only; never bash / `source .env`). Two-stage review per task (spec-compliance then code-quality).

## 10. Out of scope (bounded defers)

- The dedicated dictionary-management screen and the inline-create-from-template UX → **SP-3**.
- Proactive advisory warning on `/tokens` write vs active template schemas (TD-1 item 3) → **SP-3** (D9).
- Fixing the pre-existing `placeholderCatalogSet` (7) vs resolver registry (8, `approval_date`) inconsistency → left as-is; SP-2 reads the authoritative registry.
- Fixing the pre-existing name-vs-ID keying drift between `Materialize` and `ReadForReconstruction` → out of scope (drive-by rule; repair only contract/invariant guards).

## 11. Module boundary summary

| Module | SP-2 change | Edge |
|--------|-------------|------|
| `tokens` | reserved-name guard at create; new `ReservedNames` consumer port | consumes injected native-names (via port, not a `render` import) |
| `templates` | `PHDictionary` type + validation | uses already-injected `resolvers.Known()` |
| `documents` / `controlleddocuments` | off-tx dictionary resolution + in-tx pin (`source="dictionary"`) | **NEW** `documents → tokens` via `DictionaryReader` (published port only) |
| `render` | none | unchanged |

Invariants honored: capabilities-not-roles (`token.view`-gated read; no new role), contract-first (no hand-added routes), multi-tenant pooled (tenant-predicated read), DB-enforced (constraints unchanged; verify source CHECK), cross-module via published port only (`DictionaryReader`), H-PRE-1 (dictionary read off the atomic tx).
