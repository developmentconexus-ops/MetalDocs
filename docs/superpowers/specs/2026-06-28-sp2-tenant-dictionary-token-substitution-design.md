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
| D8 | Reproducibility | Value pinned at creation of **each revision** → immune to later dictionary edits; covered by existing `values_hash` (verified source-agnostic over placeholder values, `values_hash.go:11-29`). No render-path change. |
| D9 | Write-guard advisory on `/tokens` (TD-1 item 3) | **Deferred to SP-3** (UI edge). |
| **D10** | **Revision semantics** | **Re-resolve-on-revision.** `CreateRevision` re-runs the same `CloneTemplate` path, so each new revision pins the dictionary value current at *that revision's* creation (not cloned from the parent revision). Rationale: matches the existing clone path (no special-case branch), gives per-revision determinism, and each revision is itself immutable post-creation. The D8 immunity is therefore *per revision*, not "the value the document was first created with forever." **Must be tested** with an explicit revision case. |
| **D11** | **Author-overwrite guard (governance enforcement)** | A document author's fill-in write (`SetPlaceholderValue`) may target only placeholder rows whose **current** `source ∈ {user, default}`; writes to `source ∈ {computed, dictionary}` are **rejected**. Without this, an author can overwrite a tenant-governed dictionary value during the fill-in window (existing `SetPlaceholderValue` hardcodes `Source:"user"` and upserts `source=EXCLUDED.source` with no guard — `fillin_service.go:128-138`, `fillin_repository.go:46-79`). Enforced at **both** the app layer (friendly 409/422) and the DB (`... DO UPDATE ... WHERE source IN ('user','default')`), per the DB-enforces-invariants rule. |

## 3. Token taxonomy (the unified kernel)

The table is **`document_placeholder_values`** (keyed by placeholder ID; rendered into the body by eigenpal keyed by `name`). It already has **three** value **sources** today — `user`, `computed`, and `default` (the value seeded for an as-yet-unfilled required placeholder). SP-2 adds a fourth: `dictionary`. This is the global-maximum move: extend the existing kernel, do not fork a parallel path.

| Source | Declared where | `Name` rule | Resolved when | Resolved from |
|--------|----------------|-------------|---------------|---------------|
| `default` (existing) | template placeholder_schema, `Required=true`, before fill-in | — | document creation (seeded valueless) | placeholder seeding (`repository.go:247`, `source='default'`) |
| `user` | template placeholder_schema (no `Name`, ID-keyed) | — | document fill-in | author input (`fillin_service.go:134`) |
| `computed` | template placeholder_schema (`Name` ∈ native catalog, `Type=computed`, `ResolverKey=Name`) | must be a native name | freeze (`pinValidateAndHash`) | `render/resolvers` registry |
| **`dictionary` (NEW)** | template placeholder_schema (`Name` = dict entry name, `Type=dictionary`) | must **not** be a native name; must match dict-name format | **document creation** (re-resolved per revision, D10) | `tokens.DictionaryReader` |

All flow through the **one** existing freeze → materialize → eigenpal substitution path unchanged. The `source` column on `document_placeholder_values` is the discriminator (`fillin_repository.go:22`), constrained by a DB CHECK that **must be extended** (see §6). Honesty note: the "kernel" is not yet perfectly uniform — `user`/`default` key by ID, `computed`/`dictionary` carry a `name`; and the materialize vs. reconstruction paths key differently (§10 S2). SP-2 rides on those existing seams rather than unifying them (a larger boundary); the seams are recorded, not hidden.

## 4. Architecture & data flow

```
TEMPLATE AUTHORING                       DOC CREATION / REVISION (from template)   FREEZE / MATERIALIZE
──────────────────                       ───────────────────────────────────────  ────────────────────
templates.UpdateSchemas:                 controlleddocuments.Create / CreateRev:   (UNCHANGED)
  schema gains a PHDictionary             1. OFF-TX (before runner.Do @ :334):      pinned dictionary values
  reference {type:"dictionary",              resolve schema off-tx → extract        are already in
   name:"company_name"}                      PHDictionary phs → for each:           document_placeholder_values
        │                                     DictionaryReader.GetByName(tenant,     (source="dictionary"); the
        ▼                                     name) [own DoReadOnly tx, authz];      existing eigenpal pass
  ValidatePlaceholders (free fn):            ErrNotFound → 422; build {phID→val}    substitutes {company_name}
    branch on Type FIRST;       ◄──┐      2. ATOMIC TX (cloneIntoTx →               in the body docx
    PHDictionary: name ∉ native    │         CreateDocumentTx): seed value-bearing
    registry(8) + format (D5)      │         rows source="dictionary" (separate
                                   │         from valueless source='default' loop)
  tokens.CreateToken (D4): ────────┘                │
    reject name ∈ native registry                   ▼
    (primary guard)                        doc "comes rendered"; value pinned per
                                           revision, immune to later dict edits (D8/D10)
  fill-in SetPlaceholderValue: reject author writes to source∈{computed,dictionary} (D11)
```

**Value is the source of truth in the dictionary; the document captures a pinned copy at each revision's creation.** A later `PUT /tokens` never alters an already-created revision. The schema must be **resolved off-tx** before the atomic tx opens (`service.go:334`) — today the placeholder slice is resolved *in-tx* at `service.go:278`, so SP-2 hoists an off-tx schema read (the dictionary `GetByName` is authz-recording on its own tx and cannot run inside the creation tx — H-PRE-1).

## 5. Components (per module)

### 5.1 `tokens` — write-time reserved-name guard (D4)
- **New consumer port** in `internal/modules/tokens/application/ports.go`: `ReservedNames` (e.g. `IsReserved(name string) bool`). Keeps `tokens` a leaf — **no `tokens → render` import**.
- **Composition root** (`apps/api/cmd/metaldocs-api/main.go`): inject an impl backed by `render/resolvers.Registry.Known()` (the authoritative 8 native keys — `templates`' static `placeholderCatalogSet` is missing `approval_date`; we use the registry, not that set, and do **not** fix the pre-existing 7-vs-8 gap here).
- **`CreateToken`** (application `Service.Create` @ `service.go:73`): reject `name ∈ reserved` → new `domain.ErrReservedName` sentinel (alongside `ErrImmutableName` in `tokens/domain/entry.go`), mapped to **422 `reserved_name`** in the delivery `writeTokenError` switch (`delivery/http/handler.go:191-211`, same `errors.Is` pattern as `ErrImmutableName→422`). Name is immutable post-create, so Create is the only guard point (Update cannot change name).
- **No migration, no new route, no new capability.**

### 5.2 `templates` — `PHDictionary` schema support (D2, D5)
- **Domain** (`internal/modules/templates/domain/schemas.go:24-34`): add `PHDictionary PlaceholderType = "dictionary"` to the `const` block.
- **Validation** — ⚠️ **corrected against runtime.** `ValidatePlaceholders` (`schema.go:103`) is a **package-level free function** `func ValidatePlaceholders(phs []domain.Placeholder) error` — it has **no receiver and no access to `s.resolvers`** (the earlier draft's `schema.go:47-56`/`s.resolvers` reference was the `UpdateSchemas` *method*, a different site). Two structural changes are required:
  1. **Thread the native-name set in.** Give `ValidatePlaceholders` access to `resolvers.Known()` — either promote it to a method on the service or add a `knownResolvers map[string]…`/`func(string) bool` parameter. The native set is the **registry's 8 keys** (`resolvers.Known()`), **not** the static `placeholderCatalogSet` (7, missing `approval_date`) — otherwise a dictionary token named `approval_date` slips through (N1).
  2. **Type-dispatch BEFORE the existing name rule.** The current rule (`schema.go:114-127`) says *any* placeholder with a non-empty `Name` must be in the catalog **and** be `PHComputed` with `ResolverKey==Name` (else `ErrPlaceholderNotInCatalog`/`ErrPlaceholderNotComputed`). A `PHDictionary` carries a `Name` **by design** but is neither catalog nor computed, so under today's code **every dictionary placeholder is rejected.** Restructure to branch on `Type` first:
     - `Type == PHDictionary`: require non-empty `Name` matching `placeholderNameRe` (`^[a-z][a-z0-9_]{0,49}$`, `schema.go:96`); `Name` must **not** be a native name; `ResolverKey == nil`; `Computed == false`; participates in the existing `seenNames` duplicate check.
     - else (existing branch, unchanged): the `Name` ∈ catalog ⇒ computed rule.
- **No migration** — `placeholder_schema` is a JSON column and the OpenAPI DTO is a free-form object array (§6); `PHDictionary` is a new value within it.

### 5.3 `documents` + `controlleddocuments` — creation-time resolution & pinning (D1, D7, D8, D10)
- **Off-tx schema availability — ⚠️ corrected.** In `controlleddocuments.Create` the off-tx block (`service.go:322-332`) resolves only the **template version ID** (`ResolveTemplateVersionID`) + `ensureTemplateArtifact`; the **placeholder slice** is resolved *in-tx* by `snapshotSvc.ResolveTemplate` at `service.go:278` (inside `cloneIntoTx`), and the atomic tx opens at `service.go:334` (`s.runner.Do`). Because the dictionary `GetByName` is an authz-recording read on its **own** `DoReadOnly` tx (verified `tokens/application/service.go:186-207`), it **cannot** run inside the creation tx (H-PRE-1). SP-2 therefore **hoists a schema read off-tx** before `:334`: resolve the placeholder slice off-tx, extract the `PHDictionary` entries, resolve each value, then pass the `{placeholderID → value}` map into the creation tx. Either thread the already-resolved snapshot through `cloneIntoTx` (avoids a second read) **or** accept a second off-tx read of the immutable template version (lower churn); the plan picks, but the dictionary resolution **must** be off-tx.
- **Resolution** (off-tx): for each `PHDictionary` placeholder call `tokens.DictionaryReader.GetByName(ctx, tenantID, name)`. Actor flows via context (`iamdomain.UserIDFromContext(ctx)`), `token.view`-gated, tenant-predicated. Branch the error precisely: `errors.Is(err, domain.ErrNotFound)` → **422** fail-closed (D7); an authz denial or infra error → **403/5xx**, never mis-mapped to 422. (Empty values are impossible — `tokens/domain/entry.go` rejects `len<1`, DB CHECK `char_length BETWEEN 1 AND 4096` — so "missing vs empty" collapses to not-found.)
- **Seeding seam — ⚠️ corrected.** The existing in-tx seeding (`repository.go:247-257`, called from `CreateDocumentTx` @ `repository.go:125`) iterates `requiredPlaceholders` and inserts **valueless** rows with `source='default'`; and `parseRequiredPlaceholders` (`snapshot_service.go:43-54`) **filters the schema to `Required=true`** before it reaches `CreateDocumentTx(... phs)`, whose param carries **IDs only, no values**. So dictionary placeholders (a) may be filtered out by the `Required` filter, and (b) need a **value-bearing** insert distinct from the valueless `default` loop. SP-2 threads the resolved `{phID→value}` map through `cloneIntoTx` → `CreateDocumentTx` as a **separate parameter** and seeds those rows with `source='dictionary'` and `value_text` **inside** the creation tx, independent of the `Required` filter.
- **Revision path (D10).** `CreateRevision` (`controlleddocuments/application/service.go:696`) re-runs the same `CloneTemplate`/`cloneIntoTx` chain, so wiring the resolution there means each **revision re-resolves** the dictionary against the current value and pins it (it does **not** clone the parent revision's row). This is the chosen, stated semantics — the resolution+seeding code lives in the shared clone path so both fresh-create and revision get it.
- **Wiring** (`main.go`): inject `tokens.Module.Reader` (a `DictionaryReader`) into the creation service.
- **Multi-tenant:** `DictionaryReader` predicates on the creation tenant; no cross-tenant read.

### 5.5 `documents` — author-overwrite guard on governed sources (D11)
- **`fillin_service.SetPlaceholderValue`** (`fillin_service.go:128-138`) currently hardcodes `Source:"user"`, never reads the existing row, and the repo upsert (`fillin_repository.go:46-79`) sets `source = EXCLUDED.source` on the `(tenant_id, revision_id, placeholder_id)` PK — so an author write flips **any** row (computed or dictionary) to `user` with their input. Add a guard so author fill-in may write only rows whose **current** `source ∈ {user, default}`:
  - **App layer:** read the target row's current source first; if `∈ {computed, dictionary}`, reject with a clear problem (409 `placeholder_not_author_editable` or 422). Friendly first line.
  - **DB layer (enforcing):** scope the upsert's `DO UPDATE` with `WHERE document_placeholder_values.source IN ('user','default')` so the invariant holds even if the app check is bypassed (DB-enforces-invariants rule).
- This closes the governance hole SP-2 introduces (a tenant-governed value an author could silently overwrite). It also incidentally protects `computed`; that pre-existing exposure is in-scope only because SP-2 is the change that makes governed values author-reachable.

### 5.4 `render` — no change
Pinned `source="dictionary"` values are already in `placeholder_values` at freeze; the existing `freeze_service.go` materialize path maps them into `FanoutRequest.PlaceholderValues` by name and eigenpal substitutes them. **Zero render-module edits.**

## 6. Contract & data

- **OpenAPI — no change (definitive).** The template `placeholder_schema` is a **free-form object array** in both sites — write/update at `api/openapi/v1/openapi.yaml:1351-1353` (`items: {type: object, additionalProperties: true}`) and the `TemplateVersion` response DTO at `:5459-5462`. There is **no** `enum` for placeholder `type` anywhere in the spec. So `PHDictionary` requires **no OpenAPI edit and no oapi-codegen regeneration** — the type lives only in the Go `PlaceholderType` enum (`schemas.go:24-34`). `POST /tokens` shape is unchanged (the reserved-name rejection reuses the existing 422 problem envelope).
- **Migration — REQUIRED (not "none"). ⚠️ This was the spec's most consequential error.** The table is **`document_placeholder_values`** and its `source` column has a **closed** CHECK constraint that excludes `dictionary`:
  ```sql
  -- db/baseline/0001_current_schema.sql:1940  (mirror: migrations_baseline/0001_baseline_2026_05.sql:2062)
  CONSTRAINT document_placeholder_values_source_check
    CHECK ((source = ANY (ARRAY['user'::text, 'computed'::text, 'default'::text])))
  ```
  Inserting `source='dictionary'` will be **rejected by the DB on every dictionary-bearing creation**. SP-2 **must** ship a forward migration that drops and recreates the constraint with `'dictionary'` **added to the existing three** (preserve `'default'` — do not replace the set):
  ```sql
  ALTER TABLE public.document_placeholder_values
    DROP CONSTRAINT document_placeholder_values_source_check,
    ADD  CONSTRAINT document_placeholder_values_source_check
      CHECK (source = ANY (ARRAY['user'::text,'computed'::text,'default'::text,'dictionary'::text]));
  ```
  The baseline schema mirror must be updated to match (per the repo's baseline+migration convention — plan confirms the exact baseline file to touch). No new table; the placeholder-value `source` is a bare Go `string` (no typed enum to extend — `fillin_repository.go:22`).
- **Destructive change:** none — the constraint change is additive (widens the allowed set).

## 7. Error handling

| Condition | Where | Result |
|-----------|-------|--------|
| Dictionary token name = native name | `POST /tokens` | 422 `reserved_name` (D4) |
| `PHDictionary` reference name = native name | template schema-save | 422 (template validation, D5) |
| `PHDictionary` reference name bad format / empty | template schema-save | 422 (existing validation style) |
| Referenced dictionary token missing at creation/revision | doc creation | 422, blocks creation — **only** on `errors.Is(err, domain.ErrNotFound)` (D7) |
| Creating actor lacks `token.view` | doc creation | fail closed (403) — distinct branch, not 422 |
| Infra/DB error from `GetByName` | doc creation | 5xx — distinct branch, not 422 |
| Author writes a `source∈{computed,dictionary}` row | fill-in | 409/422 `placeholder_not_author_editable` (D11); DB guard backstops |
| Cross-tenant dictionary read attempt | doc creation | impossible — reader is tenant-predicated |

## 8. ADR & docs

- **New ADR 0049** — "Tenant dictionary token substitution & reserved-name guard." Fulfils ADR 0048's deferred SP-2 contract (closes TD-1), records: (a) creation-time pinning as the binding model, re-resolved per revision (D10); (b) the `PHDictionary` placeholder source extending the substitution kernel; (c) the backend reserved-name guard as a deliberate, bounded extension of ADR 0048 — explicitly within ADR 0048's blessed "catalog membership, not grammar" line (0048 line 46), **not** a grammar re-implementation; (d) the author-overwrite guard on governed sources (D11); (e) the known limitation that forensic reconstruction does not yet round-trip name-bearing sources (computed or dictionary — §10 S2). **ADR 0049 must explicitly state it supersedes ADR 0048's SP-2/SP-5 roadmap rows** (render-time merge/precedence → creation-time prevention + pinning), so the 0048 roadmap table is not left as contradicting truth. Cite REQ-AUTHZ-1/2 (authorized freeze + `token.view`-gated read), REQ-MT-1 (tenant-scoped read).
- **Wiki:** `wiki/modules/tokens.md` (flip SP-2 status; fix the §3 diagram's render-fanout framing to the documents-assembles/render-transports reality), `wiki/modules/tokens-tech-debt.md` (close TD-1), `wiki/modules/templates.md` (`PHDictionary`), `wiki/concepts/placeholders.md` (third value source). Refresh `Last verified`.
- **Code comment drift:** `internal/modules/tokens/domain/port.go:20-21` — the `DictionaryReader` doc-comment says "render reads dictionary values through it"; SP-2's reality is **`documents`/`controlleddocuments` reads it at creation**. Correct the comment when the consumer lands (not a contract change — the interface is unchanged).

## 9. Test & QA plan

Canonical `testdb` integration factory; `//go:build integration`; R1–R4 discipline. Unit tests for pure validation.

- **`tokens`** (integration): `POST /tokens` with a native name → 422 `reserved_name`; native name including `approval_date` (registry-only, not in `placeholderCatalogSet`) → also 422; non-native name → 201.
- **`templates`** (unit): `ValidatePlaceholders` accepts a valid `PHDictionary` reference; rejects native-name (incl. `approval_date`), bad-format, empty-name, a `PHDictionary` with a `ResolverKey`/`Computed=true`; and confirms an existing `computed` placeholder still validates (type-dispatch didn't regress the existing branch).
- **migration** (integration): the constraint migration permits `source='dictionary'` and still permits `user`/`computed`/`default`; a pre-migration insert of `'dictionary'` would fail (guard the migration's intent).
- **`documents`/`controlleddocuments`** (integration):
  - create-from-template pins the dictionary value (`source='dictionary'`, correct value, into `document_placeholder_values`);
  - **D8 (per-revision immunity):** edit the dictionary after a revision is created → that revision's pinned value is unchanged;
  - **D10 (revision re-resolve):** create doc (pins value V1), edit dictionary to V2, `CreateRevision` → the **new revision** pins V2 while the **old revision** still shows V1;
  - **D11 (governance):** an author `SetPlaceholderValue` against a `source='dictionary'` row is rejected (and the DB guard rejects it even if the app check is bypassed);
  - **D7:** missing referenced token → 422; authz-denied read → 403 (not 422);
  - two-tenant isolation (tenant A's creation never reads tenant B's dictionary).
- **`render`** (integration): a pinned `source='dictionary'` value substitutes into the body docx through the **normal freeze→materialize path** (not the reconstruction path — see §10 S2). Assert `values_hash` is stable across two freezes of the same pinned revision.

**Gates (per task + at close):** `go build ./...`; `go test ./...`; `go test -tags=integration ./internal/modules/{tokens,templates,documents,render}/...`; `.\scripts\check-system-runnable.ps1` (PowerShell only; never bash / `source .env`). Two-stage review per task (spec-compliance then code-quality).

## 10. Out of scope (bounded defers)

- The dedicated dictionary-management screen and the inline-create-from-template UX → **SP-3**.
- Proactive advisory warning on `/tokens` write vs active template schemas (TD-1 item 3) → **SP-3** (D9).
- Fixing the pre-existing `placeholderCatalogSet` (7) vs resolver registry (8, `approval_date`) inconsistency → left as-is; SP-2 reads the authoritative registry everywhere it guards.
- **S2 — forensic reconstruction does not round-trip name-bearing sources.** `ReadForReconstruction` (`resolver_readers.go:134-139`) keys `FanoutRequest.PlaceholderValues` by placeholder **ID**, while the normal materialize path re-keys ID→**Name** and eigenpal substitutes by **name** — so `computed` (today) and `dictionary` (after SP-2) reconstruct to empty and a forensic re-render diverges. SP-2 does **not** fix this keying drift (a larger boundary), but it is **no longer a silent defer**: ADR 0049 records it as a known limitation (§8e), and the §9 render test asserts the *normal* path, not reconstruction, so a green suite is not mistaken for reconstruction coverage. Unifying the ID→Name keying across materialize + reconstruct is the real global-maximum fix, deferred with a named owner (post-SP-2 token-kernel hardening).
- **Schema↔body-token reconciliation** (a declared `PHDictionary` with no matching `{name}` in the body docx pins a value that renders nowhere; and the inverse) → unvalidated for **all** sources today (no schema-to-body check exists); consistent with D6's decoupling. SP-3 UI owns live body-token reconciliation. Noted so "comes rendered" is not mistaken for a body-token guarantee.
- **Per-token capture audit event** → not added. The persisted `source='dictionary'` row (with its `value_text`) **is** the durable forensic record of what was captured at each revision's creation; `values_hash` anchors determinism. A separate audit event is a nice-to-have, not SP-2.

## 11. Module boundary summary

| Module | SP-2 change | Edge |
|--------|-------------|------|
| `tokens` | reserved-name guard at create; new `ReservedNames` consumer port; `ErrReservedName` sentinel→422 | consumes injected native-names (via port, not a `render` import) |
| `templates` | `PHDictionary` type; `ValidatePlaceholders` type-dispatch restructure + native-name set threaded in (free fn → method/param) | native-name set = `resolvers.Known()` (8), not `placeholderCatalogSet` (7) |
| `documents` / `controlleddocuments` | off-tx schema hoist + off-tx dictionary resolution; value-bearing in-tx pin (`source='dictionary'`); revision re-resolve (D10); author-overwrite guard (D11) | **NEW** `documents → tokens` via `DictionaryReader` (published port only) |
| `db` | forward migration: widen `document_placeholder_values_source_check` to include `'dictionary'` (+ baseline mirror) | — |
| `render` | none | unchanged |

Invariants honored: capabilities-not-roles (`token.view`-gated read; no new role/capability), contract-first (placeholder_schema is free-form — no route/DTO change), multi-tenant pooled (tenant-predicated read), **DB-enforced (source CHECK widened by migration; author-overwrite guard enforced in DB)**, cross-module via published port only (`DictionaryReader`), H-PRE-1 (dictionary read off the atomic tx, schema hoisted off-tx). ADR 0048 grammar boundary upheld (catalog-membership guard, not grammar — 0048 line 46).
