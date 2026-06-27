# Token Syntax — `{name}` (docxtemplater native)

> **Last verified:** 2026-06-27
> **Scope:** Why `{name}` was chosen, what the format implies, comparison with the legacy `{{uuid}}` approach.
> **Out of scope:** Migration mechanics (see `decisions/0003-token-syntax-migration.md`), placeholder concept overall (see `placeholders.md`).
> **Key files:**
> - `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` — current `` `{${placeholder.name}}` `` insertion (renamed from `TemplateAuthorPage` 2026-05-10)
> - `packages/editor-ui/src/MetalDocsEditor.tsx:54` — `templatePlugin` (detects `{name}`)

---

## The two formats

### `{name}` (single brace) — eigenpal / docxtemplater native

```
Hello {customer_name}, your order #{order_id} ships on {ship_date}.
```

- **Origin:** Mustache → Handlebars → docxtemplater
- **Identifier type:** Semantic name (e.g., `customer_name`)
- **Detection:** Auto by docxtemplater regex
- **Rendering:** Library-native — eigenpal's `templatePlugin` highlights live
- **Tooling:** Universal — Word's "merge field" preview, docxtemplater CLI, eigenpal, etc. all understand it

### `{{uuid}}` (double brace) — MetalDocs legacy (removed 2026-04-25)

```
Hello {{a3f1b2c0-...}}, your order #{{8d2e9c44-...}} ships on {{02b5e8a1-...}}.
```

- **Origin:** Custom MetalDocs convention (looks like Mustache double-brace, but UUIDs not names)
- **Identifier type:** UUID (opaque)
- **Detection:** Custom regex in MetalDocs server fanout
- **Rendering:** No editor highlighting (eigenpal regex doesn't match `{{...}}`)
- **Tooling:** MetalDocs only
- **Status:** Removed. See `decisions/0003-token-syntax-migration.md`.

## Why `{name}` wins

| Factor | `{name}` | `{{uuid}}` |
|--------|----------|------------|
| Author can read raw DOCX | ✅ semantic | ❌ opaque hex |
| Editor highlighting | ✅ free via eigenpal | ❌ would need custom plugin |
| Industry tooling | ✅ docxtemplater, Carbone, others | ❌ MetalDocs-only |
| Audit trail readability | ✅ "value of {revision} = 1.2" | ❌ "value of {{a3f1b2c0...}} = 1.2" |
| Template authoring in Word desktop | ✅ paste tokens directly, validate later | ❌ must use MetalDocs UI to insert UUIDs |
| Stable across renames | ⚠️ rename = breakage if not refactored | ✅ rename label, ID stays |
| Collision risk | ⚠️ duplicate names within template | ✅ UUID always unique |

The "stable across renames" + "no collision" advantages of UUIDs are the only real wins, both addressable:
- **Renames:** Provide a "rename placeholder" action that refactors the DOCX in place (find/replace `{old_name}` → `{new_name}`).
- **Collisions:** Validate uniqueness on schema save (server returns 422 if duplicate).

## Key charset and the Node-owned-grammar boundary

### Canonical charset

MetalDocs token identifiers follow snake_case: `IDENT_RE = /^[A-Za-z_][A-Za-z0-9_]*$/` (from `packages/shared-tokens/src/grammar.ts`). Dotted names (e.g., `{a.b}`) and hyphenated names are **rejected** — they collide with docxtemplater's nested-property path semantics (the vendor traverses `a.b` as `obj.a.obj.b`). Authors who type `{a.b}` see an "invalid token" warning in the panel; freeze may still act on the literal text, but the MetalDocs schema never accepts it.

### Node-owned-grammar boundary

Token parsing lives entirely in TypeScript/Node: `@metaldocs/shared-tokens` owns `scanText` (the single scan regex), `detectTokens` (browser adapter), and `parseDocxTokens` (server-side DOCX parser). Both paths share the same grammar core.

**Go never parses tokens.** The Go backend validates tokens by catalog/dictionary membership over the `DetectedToken` output surfaced by the editor adapter, and by checking `unreplacedVars` from the renderer. Token syntax is never re-implemented in Go — this is a binding invariant (any Go token parser would be a second grammar that could drift).

For SP-2 (tenant dictionary), the validation flow will be: editor → `detectTokens` → `DetectedToken[]` (valid/invalid) → API carries token names → Go checks names against tenant dictionary → 422 if name not in dictionary. The charset gate happens client-side; the catalog membership gate happens server-side.

## Sections, conditionals, raw — eigenpal extras we miss today

Docxtemplater syntax includes more than simple substitution:

```
{#items}                 <-- begin loop
  - {name}: ${price}
{/items}                 <-- end loop

{^empty_section}         <-- if NOT empty
  Default content.
{/empty_section}

{@raw_html}              <-- inject raw OOXML
```

Legacy `{{uuid}}` covered only simple substitution. `{name}` syntax unlocks sections, conditionals, and raw injection natively.

## MetalDocs schema (current)

```json
{
  "id": "a3f1b2c0-...",          // internal PK only — never in DOCX
  "name": "customer_name",       // slug — used as the token, unique per template version
  "label": "Customer Name",      // human display
  "type": "text",
  "required": true
}
```

Token in DOCX: `{customer_name}`

Frontend insert: `` insertTokenAtCursor(`{${placeholder.name}}`) ``

Backend substitution map: `{ customer_name: "Acme", revision: "1.2", ... }` — semantic keys, not UUIDs.

## Cross-refs

- [placeholders.md](placeholders.md) — full concept
- [decisions/0003-token-syntax-migration.md](../decisions/0003-token-syntax-migration.md) — migration ADR
- [references/eigenpal-spike.md](../references/eigenpal-spike.md) — T4 used `{name}` fixture, validated approach
- [modules/templates.md §2](../modules/templates.md) — architecture constraint: `{name}` single-brace eigenpal-native syntax required by the templates backend (oapi-codegen + `ValidatePlaceholders`)
