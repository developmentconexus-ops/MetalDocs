# 0048 — Tenant Token Dictionary

- **Status:** Accepted
- **Last verified:** 2026-06-28
- **Supersedes:** ADR 0008 (token system — the definition of what constitutes a "token" and who owns the grammar)
- **Scope:** Design decisions for the SP-1 per-tenant author-defined `name → value` dictionary module (`internal/modules/tokens`).

---

## Context

SP-1 adds a new class of token: **tenant-defined dictionary entries**. These are runtime, tenant-governed `name → value` constants (e.g., `COMPANY_NAME = "Acme Metalurgia"`) that authors can reference in document templates without modifying the template's placeholder schema.

Prior to SP-1 the only tokens MetalDocs understood were **computed tokens** — system-defined placeholders (e.g., `{REVISION}`, `{AUTHOR}`) enumerated in the template's `placeholder_schema` and resolved at render time by the `render-fanout` module. ADR 0008 established the token syntax (`{name}` single-brace, eigenpal-native) but did not distinguish dictionary tokens from computed tokens, and did not address dictionary governance.

Two design decisions were required before implementing the tokens module:

**(A)** Who owns the token grammar — Go or Node?

**(B)** Who governs what names exist in the dictionary — system or tenant?

---

## Decision A: Node owns the token grammar; Go does anti-corruption storage hygiene only

### Choice

The canonical token grammar lives entirely in TypeScript/Node: `@metaldocs/shared-tokens` owns `scanText`, `detectTokens`, `parseDocxTokens`, the leading-char rule (`^[A-Za-z_]`), and the reserved-word list. Grammar enforcement is applied at the UI/editor edge (SP-3 will surface this in the token panel and creation form).

Go's validation (`nameRe = ^[A-Za-z0-9_]+$`, length 1–64) is **anti-corruption storage hygiene**: it mirrors the DB CHECK constraints and rejects characters the DB would reject, but it does not enforce the leading-char rule or any reserved-word list.

### Rationale

1. **Grammar is a UI/product concern.** What names are allowed, what the leading-char rule is, what words are reserved — these are editorial decisions driven by docxtemplater semantics and MetalDocs product rules. They belong in the layer where authors interact.

2. **Go must not implement a second grammar.** A Go token parser that re-implements the Node grammar would create two sources of truth that can drift. This is an explicit binding invariant documented in `wiki/concepts/token-syntax.md`: "Go never parses tokens."

3. **Permissive storage is safe at this level.** `^[A-Za-z0-9_]+$` is a subset of the Node grammar's valid identifiers. Any name that passes the Node grammar will pass Go's storage hygiene. Names that fail the leading-char rule at the UI edge never reach the API.

4. **Grammar can evolve independently.** If the Node grammar adds new reserved words or tightens rules, it does so without a Go migration. The store does not need to know about product rule changes.

### Consequences

- SP-3 UI must enforce the full Node grammar (leading-char rule, reserved words) before submitting to the API.
- Go's CHECK constraint (`name ~ '^[A-Za-z0-9_]+$'`) rejects pathological input (control characters, spaces, dots) but never substitutes for Node grammar enforcement.
- If a name that violates Node grammar somehow reaches the API (e.g., via direct API call), it will be stored. SP-3 will make this visible in the token panel as "invalid name". This is acceptable — the rendering path checks catalog membership, not grammar.

---

## Decision B: Dictionary entries are tenant-governed, not system-defined

### Choice

Tenant admins (holders of `token_dictionary.manage`) create, update, and delete their own dictionary entries. No system-level seeding; no cross-tenant entry. The dictionary is governed entirely by tenant operators at runtime.

### Rationale

1. **Tenant autonomy.** Different tenants have different company names, legal entities, and boilerplate phrases. A system-level dictionary cannot serve this without per-tenant overrides, which collapses to per-tenant governance anyway.

2. **Capability-gated, audited.** Entry mutations are guarded by `CapTokenDictionaryManage` (two-tier: path-prefix dispatcher + in-tx `authz.Require`) and every write is recorded as a `tokens.entry.created/updated/deleted` audit event inside the same transaction. This satisfies the same traceability requirement as other regulated mutations in MetalDocs.

3. **Alignment with MetalDocs authz model (ADR 0022).** "Admin/editor/author can manage tokens" reasoning is explicitly prohibited. The correct framing is "holder of `token_dictionary.manage` can create/update/delete entries." The capability is granted to the appropriate roles via IAM reference data, not hard-coded in the module.

### Consequences

- `token_dictionary_entries` table has `tenant_id UUID NOT NULL` with a NULL-permissive RLS policy (ADR 0247 pattern). No FK to the tenants table.
- `(tenant_id, name)` unique index enforces no duplicate names within a tenant.
- No system seeding of dictionary entries. Tenants start with an empty dictionary.
- `CapTokenDictionaryManage` and `CapTokenView` are registered in `internal/modules/iam/domain/capability.go` and granted via IAM reference data to the appropriate roles.

---

## SP-2..SP-5 Roadmap

| Sprint | Work | Module |
|--------|------|--------|
| SP-2 | Render substitution: `render-fanout` consumes `DictionaryReader.GetByName` / `List` at document generation time. Merge precedence vs. computed tokens must be specified (see tech-debt TD-1). | render-fanout |
| SP-3 | UI: Token panel in the document editor; full Node grammar enforcement on the creation form; reserved-word validation at the API call boundary. | Frontend, shared-tokens |
| SP-4 | Import/export: bulk CSV import for tenant token dictionary; export for backup/transfer. | tokens (new routes) |
| SP-5 | Computed+dictionary collision reconciliation: formal merge strategy; optionally surface collision warnings at dictionary-write time. | tokens, render-fanout |

---

## Supersession of ADR 0008

ADR 0008 established the `{name}` single-brace token syntax and the Node-owned grammar boundary. This ADR (0048) adds:

1. The distinction between **computed tokens** (system-defined, template-schema-driven) and **dictionary tokens** (tenant-defined, runtime CRUD).
2. The explicit governance model for dictionary tokens (capability-gated, audited, tenant-scoped).
3. The anti-corruption boundary: Go hygiene vs. Node grammar.

ADR 0008 remains the source of truth for the `{name}` syntax itself and the prohibition on Go token parsing. ADR 0048 governs the dictionary module's structure and governance.

---

## Related

- `wiki/modules/tokens.md` — implementation detail
- `wiki/modules/tokens-tech-debt.md` — TD-1 (collision), TD-2 (strictjson shim)
- `wiki/concepts/token-syntax.md` — grammar boundary reference
- `wiki/decisions/0007-two-tier-authz.md` — two-tier authz pattern
- `wiki/decisions/0022-capabilities-not-roles.md` — capability model
- `wiki/decisions/0247-rls-null-permissive.md` — RLS pattern used on `token_dictionary_entries`
- `internal/modules/tokens/domain/port.go` — `DictionaryReader` (SP-2 published port)
