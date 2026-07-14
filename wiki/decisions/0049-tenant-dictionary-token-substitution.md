# 0049 — Tenant Dictionary Token Substitution (SP-2)

- **Status:** Accepted
- **Last verified:** 2026-06-28
- **Supersedes (partial):** ADR 0048 SP-2/SP-5 roadmap rows (render-time merge / precedence → creation-time prevention + pinning; see §Supersession below)
- **Scope:** SP-2 design decisions: how tenant dictionary token values enter `document_placeholder_values`, the `PHDictionary` placeholder source, the backend reserved-name guard, author-overwrite protection on governed sources, and the known forensic reconstruction limitation.

---

## Context

ADR 0048 established the per-tenant dictionary module (`internal/modules/tokens`) and closed SP-1. Its roadmap table (SP-2 row) described the planned approach as: "`render-fanout` consumes `DictionaryReader.GetByName` / `List` at document generation time. Merge precedence vs. computed tokens must be specified (TD-1)." The SP-5 row described a "formal merge strategy."

During SP-2 design, the team evaluated three binding-time architectures:

- **Arch-A (creation-time pinning):** Resolve dictionary values off-tx at document creation; pin them into `document_placeholder_values` with `source='dictionary'` in the same creation tx. Render receives values, not names.
- **Arch-B (freeze-time merge):** Resolve dictionary values at finalize/freeze inside the render fanout pipeline. Requires render to call `DictionaryReader` and merge two catalogs at substitution time.
- **Arch-C (hybrid):** Record the dictionary name at creation; re-resolve at freeze.

The team chose **Arch-A**. The rationale is documented below. This choice made the SP-5 "formal merge strategy" row obsolete.

---

## Decision Summary Table

| # | Decision | Choice |
|---|----------|--------|
| D1 | Binding time | Creation-time pinning (Arch-A) |
| D2 | Template placeholder source | New `PHDictionary` ("dictionary") type in `placeholder_schema` |
| D4 | Reserved-name guard (primary) | Write-time in `POST /tokens`: reject names that equal a native/computed key |
| D5 | Defense-in-depth at schema-save | `ValidatePlaceholders` rejects `PHDictionary` references whose name is a native key |
| D6 | Dictionary-existence check at template-save | None — keeps `templates` decoupled from `tokens` |
| D7 | Missing dictionary token at creation | 422 `DICTIONARY_TOKEN_MISSING` |
| D8 | Reproducibility | `values_hash` deterministically captures the pinned dictionary value |
| D10 | Per-revision re-pin | Yes — each new revision re-resolves off-tx, immune to later dictionary edits |
| D11 | Author-overwrite guard on governed sources | App-layer 409 + DB `ON CONFLICT … WHERE` backstop on `source IN ('computed','dictionary')` |

---

## Decision A: Creation-Time Pinning (D1, D10)

### Choice

Dictionary token values are resolved **off-tx** (via the tokens module's published `domain.DictionaryReader`) and **seeded in-tx** into `document_placeholder_values` with `source='dictionary'` during:

1. **Document creation** — when `documents`/`controlleddocuments` initializes a new document from a template version.
2. **Each new revision** — when a new revision is created, the dictionary values are re-resolved from the current dictionary state and re-pinned, independent of the values that were pinned at creation or on any prior revision.

The render path receives pre-resolved `value_text` from `document_placeholder_values`. It does not call `DictionaryReader` and does not merge catalogs at substitution time.

### Rationale

1. **Reproducibility.** A pinned value is immutable after creation. The rendered DOCX for revision N always reflects the dictionary state at revision-N creation time, not at the time of re-render or PDF re-generation. This is the same guarantee provided by computed tokens (`source='computed'`).

2. **Eliminates the TD-1 merge problem.** With creation-time pinning there is no render-time catalog merge. A dictionary name `{REVISION}` and a computed placeholder `{REVISION}` cannot both reach render with different values — the reserved-name guard (D4/D5) prevents a `PHDictionary` reference from shadowing a native/computed name before any document is ever created.

3. **Render stays a transport.** The render-fanout pipeline's job is to apply `{name: value}` substitutions. Keeping catalog resolution out of the render path limits render's responsibility and avoids coupling render to the tokens module.

4. **Per-revision re-pin (D10) is correct.** A new revision reflects the author's intent at revision-creation time. Re-pinning at revision N uses the dictionary as it stands then. If a dictionary entry changed between revision N−1 and N, revision N picks up the new value; revision N−1 is unchanged (its values are already committed).

5. **Satisfies REQ-TEN-1** (pooled tenant model — every tenant-owned table carries `tenant_id`; `document_placeholder_values` rows are tenant-scoped by the document's `tenant_id`).

### Consequences

- The `documents` and `controlleddocuments` creation paths are responsible for calling `DictionaryReader` off-tx and passing resolved values into the in-tx creation call.
- A composition-root adapter bridges `tokens.domain.DictionaryReader` → the `documents`/`controlleddocuments` port (`DictionaryValueReader`). The `documents` and `controlleddocuments` modules import **no tokens types** — the adapter lives at the composition root (§11 invariant #6; see §Composition Root below).
- Missing dictionary entry at creation time → 422 `DICTIONARY_TOKEN_MISSING` (D7). Templates may reference dictionary names that don't yet exist; the existence check is deferred to creation.

---

## Decision B: PHDictionary — The Fourth Placeholder Source (D2, D5)

### Choice

Templates may declare a **dictionary reference** in their `placeholder_schema` as a new placeholder type `PHDictionary` (string value `"dictionary"`). A `PHDictionary` entry carries a `Name` (the dictionary entry name) and a human-readable `Label`. It carries **no value** — it is a declared reference, not a stored constant.

The four `source` values in `document_placeholder_values` are now: `default`, `user`, `computed`, `dictionary`.

### Type-dispatch validation rules (D5, D6)

At template schema-save (`ValidatePlaceholders`):

- `PHDictionary` entries are validated for format (`^[A-Za-z0-9_]+$`, 1–64 chars — same hygiene rule as the token store).
- A `PHDictionary` name must **not** equal any native/computed key (the resolver registry key set — 8 keys including `approval_date`). This is the defense-in-depth guard complementing D4.
- `PHDictionary` entries must **not** carry a `resolver_key` field or `computed: true` — these fields belong to `PHComputed`.
- **No dictionary-existence check at template-save** (D6): `templates` remains decoupled from `tokens`. SP-3 UI performs a live `GET /tokens` lookup to surface broken references to authors.

Migration `0249_document_placeholder_values_dictionary_source.sql` widened the `document_placeholder_values_source_check` CHECK constraint to include `'dictionary'` (four allowed values: `default`, `user`, `computed`, `dictionary`).

---

## Decision C: Backend Reserved-Name Guard (D4) — Within ADR 0048's "Catalog Membership, Not Grammar" Line

### Choice

A write-time guard in `tokens.application.Service.Create` rejects any dictionary entry whose `name` equals a native/computed key (the render resolver registry's key set). The guard returns HTTP 422 with code `RESERVED_NAME`.

### Relation to ADR 0048

ADR 0048 Decision A, Consequences §3 (line 46) states: "the rendering path checks **catalog membership**, not grammar." The reserved-name guard is a **catalog membership extension**, not a grammar re-implementation. It does not re-implement the Node token grammar (leading-char rule, reserved words in `@metaldocs/shared-tokens`). It enforces that a tenant cannot create a dictionary entry that would shadow a name the compute layer owns.

### Bounded extension

- The `tokens` module defines a `ReservedNames` port (`application.ReservedNames`): `interface { IsReserved(name string) bool }`.
- The composition root supplies the concrete adapter (the render resolver registry's key set).
- `tokens` has **no import** of the render module. The dependency flows inward: composition root → tokens, not tokens → render.

This structure satisfies REQ-AUTHZ-1 (code checks capabilities, never role names) for the authz dimension and preserves ADR 0048's "Go never parses tokens" invariant for the grammar dimension.

---

## Decision D: Author-Overwrite Guard on Governed Sources (D11)

### Choice

`document_placeholder_values` rows with `source IN ('computed', 'dictionary')` are governed — they were seeded by the system at creation time and must not be overwritten by author input. Two enforcement layers:

1. **App-layer check:** Any attempt to update a governed row returns HTTP 409 before touching the DB.
2. **DB backstop:** The `ON CONFLICT (revision_id, placeholder_id) WHERE source IN ('computed','dictionary') DO NOTHING` clause (or equivalent PK conflict with conditional guard) ensures that even a buggy app path cannot overwrite a governed row.

### Rationale

Computed and dictionary values are pinned by the system at known points (creation, per-revision re-pin). Author mutations would break reproducibility (D8) and undermine the audit guarantee.

This satisfies REQ-AUTHZ-2 (every capability reference is a typed registry const — the guard is a data-integrity invariant, not an authz capability check, but it enforces the same "governed writes are system-only" principle).

---

## Known Limitation: Forensic Reconstruction (§10 S2)

The `ReadForReconstruction` path in the documents module does **not** round-trip name-bearing sources. Specifically:

- `source='computed'` rows carry a `computed_from` key and a `resolver_version`. Re-constructing the human-readable name from these fields is not implemented.
- `source='dictionary'` rows carry the **pinned value** in `value_text` but do **not** carry the dictionary entry name that was used to derive it. Forensic readers see the value but cannot determine which dictionary entry it came from.

**This is out of scope for SP-2.** A post-SP-2 owner must be named before shipping SP-3. The fix requires either:
- Storing the dictionary entry name in a `computed_from`-equivalent column for `source='dictionary'` rows, or
- Adding a forensic reconstruction index table.

**Named post-SP-2 owner:** Decoupled from SP-3 (operator decision 2026-06-29). SP-3 is the
dictionary management CRUD UI and does not read or reconstruct provenance, so it does not
depend on this fix. Forensic reconstruction (storing the dictionary entry name on
`source='dictionary'` pinned rows) is owned by the future forensic-audit epic and must be
resolved before any forensic-audit feature is certified.

---

## Supersession of ADR 0048 Roadmap Rows

The ADR 0048 SP-2/SP-5 roadmap rows described:
- **SP-2:** "`render-fanout` consumes `DictionaryReader.GetByName` / `List` at document generation time. Merge precedence vs. computed tokens must be specified (TD-1)."
- **SP-5:** "Computed+dictionary collision reconciliation: formal merge strategy; optionally surface collision warnings at dictionary-write time."

**ADR 0049 supersedes these rows.** The actual SP-2 delivery is creation-time pinning (Arch-A), not render-time merge. TD-1 (collision problem) is resolved by the reserved-name guard (D4/D5) — prevention, not a runtime merge strategy. SP-5 as described is obsolete.

The 0048 roadmap table retains the SP-3/SP-4 rows unchanged.

---

## Composition Root Adapter (§11 Invariant #6)

The `documents` and `controlleddocuments` modules define their own port for reading dictionary values (`DictionaryValueReader`). They do **not** import `internal/modules/tokens` types. The composition root (`apps/api/cmd/metaldocs-api/`) supplies an adapter that implements the modules' `DictionaryValueReader` interface by delegating to `tokens.Module.Reader` (`domain.DictionaryReader`).

This maintains the MetalDocs cross-module access rule: cross-module access goes through a module's application service or published Go interface — never through another module's repository, SQL, or domain internals.

---

## Related

- `wiki/decisions/0048-tenant-token-dictionary.md` — SP-1 module ADR (partially superseded, §Supersession)
- `wiki/decisions/0007-two-tier-authz.md` — two-tier authz pattern
- `wiki/decisions/0022-capabilities-not-roles.md` — capability model (REQ-AUTHZ-1/2)
- `wiki/architecture/backend-target-architecture.md` — REQ-AUTHZ-1, REQ-AUTHZ-2, REQ-TEN-1
- `wiki/modules/tokens.md` — tokens module implementation
- `wiki/modules/tokens-tech-debt.md` — TD-1 (resolved by this ADR)
- `wiki/modules/templates.md` — `PHDictionary` placeholder type, `ValidatePlaceholders`
- `wiki/concepts/placeholders.md` — full placeholder concept
- `wiki/database/tables/document_placeholder_values.md` — `source` CHECK constraint
- `internal/modules/tokens/domain/port.go` — `DictionaryReader` published port
- `db/migrations/0249_document_placeholder_values_dictionary_source.sql` — widened `source` CHECK
