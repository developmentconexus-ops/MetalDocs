# SP-0 — Available-tokens palette over the computed catalog

**Date:** 2026-06-27
**Parent:** `2026-06-27-template-tokens-north-star.md`
**Scope:** Frontend (`features/templates`) + adapter (`@metaldocs/editor-ui`) only.
**No** backend change, **no** new capabilities, **no** render change, **no** content-controls.

---

## 1. Goal

Turn the "Placeholders disponíveis" panel from a dead reference list into a working
**insert palette** for the existing computed catalog, backed by Eigenpal's native template
primitive through the adapter. Author clicks a token → exact `{key}` lands at the cursor
(zero typos) → resolved server-side at freeze (unchanged pipeline).

## 2. Boundaries (non-goals)

- No tenant dictionary (SP-1+).
- No per-document fill-in schema authoring.
- No new IAM capabilities; insertion is editing the template body, already covered by
  `template.edit`.
- No backend / `render/*` / fanout change.
- No content controls. The stored buffer keeps literal `{key}` text.
- No draft-time value preview (SP-5).

## 3. Architecture & layering

Binding invariant (north-star §2): **MetalDocs → adapter → vendor.** The page speaks token
keys; the adapter owns all docxtemplater/Eigenpal translation.

```
TemplateEditorPage / AvailableTokensPanel   (MetalDocs catalog keys only)
        │  insertToken(key) · getUsedTokens(): string[]
        ▼
@metaldocs/editor-ui  (adapter)             key ⇄ "{key}"; strips TagType; hides vendor
        │  inner DocxEditorRef: getAgent() / getEditorRef() (typed)
        ▼
@eigenpal/docx-editor-react                 (vendor: templatePlugin, insertTemplateVariable)
```

No `as any`. No `TemplateTag`, no `{}` syntax in app code.

## 4. Components

### 4.1 Adapter capability (extends existing `MetalDocsEditorRef`)

Add two methods to the **existing** ref in `packages/editor-ui/src/types.ts` +
`MetalDocsEditor.tsx` (do not create a new package or ref):

```ts
interface MetalDocsEditorRef {
  // ...existing: getDocumentBuffer, saveNow, getPageCount, focus
  /** Insert the MetalDocs token `key` at the cursor. Adapter renders vendor `{key}`. */
  insertToken(key: string): void;
  /** Token keys currently present in the document body (delimiters stripped,
   *  variable-type tags only). MetalDocs language — no vendor tag objects. */
  getUsedTokens(): string[];
}
```

- `insertToken(key)` — implemented over the typed inner `DocxEditorRef` using the vendor's
  native insertion (`getAgent()` / `insertTemplateVariable` command / `getEditorRef()`
  insert-at-selection). Exact native call finalized in the plan; fallback is inserting the
  literal `{key}` text at the current selection. Renders `{key}` — the adapter owns the
  delimiter shape, callers never pass braces.
- `getUsedTokens()` — reads the native `templatePlugin` tags
  (`getTemplateTags(state)`), keeps only `type === 'variable'`, returns `tag.name`
  (no braces). The adapter is the **only** place docxtemplater concepts appear.

### 4.2 Authoring panel (`AvailableTokensPanel`, replaces `PlaceholderCatalogPanel`)

- Renders catalog entries (`PlaceholderCatalogEntry`) under a labelled group
  **"Preenchido pelo sistema (seguro)"**, each row: `{key}` · `label` · one-line
  description, plus an **Insert** affordance (click row / button).
- Click → `onInsert(key)` → page calls `editorRef.current?.insertToken(key)`.
- **Used-state**: a `usedKeys: Set<string>` prop (from the page) marks which catalog tokens
  are already in the document (visual "in use" state — replaces the dead green highlight,
  now driven by real data).
- **Styling**: CSS module + shared design tokens (`@metaldocs/shared-tokens`). No inline
  `style={{}}`.
- **No data fetching inside the panel** — catalog is passed in as a prop.

### 4.3 Page wiring (`TemplateEditorPage.tsx`)

- Catalog fetched **once** at page level (already is, L77) and passed to the panel; the
  panel's own `fetchPlaceholderCatalog` (duplicate) is removed.
- New debounced reader replaces `syncPlaceholdersFromDocument`:
  ```
  on editor change (debounced):
    used = editorRef.current?.getUsedTokens() ?? []
    setUsedKeys(new Set(used.filter(k => catalogByKey.has(k))))
    setUnknownTokens(used.filter(k => !catalogByKey.has(k)))   // §4.4
  ```
  This reader **never writes `localSchemas`** — it is display-only.

### 4.4 Unknown-token validation (kept; see SP-0 approval)

- `unknownTokens = getUsedTokens() ∖ catalog` — tokens typed into the body that match no
  catalog key. Surfaced inline as a non-blocking warning ("`{nope}` não será preenchido —
  verifique o nome"), catching typos before freeze instead of as a silent
  `UnreplacedVar`. Validation lives in the page (catalog is MetalDocs-domain); the adapter
  only supplies `getUsedTokens()`.
- Native canvas chips already show *where* tags are; this adds the catalog-correctness
  check the vendor cannot do (it does not know MetalDocs' catalog).

## 5. Deletions (gambiarra removal)

Remove from `TemplateEditorPage.tsx` and siblings:

- `syncPlaceholdersFromDocument` (incl. the `detected → localSchemas.placeholders → 400 ms
  autosave` corruption path).
- The `(editorRef.current as any).getAgent().getVariables()` reach-through.
- `syncOutline` + `lib/readHeadings.ts` custom outline (native `showOutlineButton` already
  provides outline; the custom path is dead `getAgent` code).
- `PlaceholderCatalogPanel`'s duplicate catalog fetch and inline styles (replaced by 4.2).

After removal, `localSchemas` equals the loaded schema for the whole session → no spurious
schema autosave. The schema-write path is untouched except that the corrupting writer is
gone.

## 6. Data flow

```
load:   draft.version → deriveTemplateSchemas → localSchemas (unchanged, authoritative)
        catalog fetch (once) → catalogByKey
insert: user clicks token → editorRef.insertToken(key) → adapter inserts "{key}"
        → editor change → debounced getUsedTokens() → usedKeys / unknownTokens (display only)
freeze: unchanged — fanout substitutes {key} via resolvers
```

The schema-derivation/save path (`useTemplateSchemas`, `deriveTemplateSchemas`,
`putTemplateSchemas`) is **unchanged**; SP-0 only stops the document-scan writer from
corrupting it.

## 7. Testing

- **Adapter unit** (`packages/editor-ui`): `insertToken(key)` calls the inner ref's native
  insertion with `{key}`; `getUsedTokens()` maps vendor variable tags → bare keys and
  drops non-variable tag types. Mock the inner `DocxEditorRef`.
- **Panel unit**: renders catalog rows; click → `onInsert(key)`; `usedKeys` drives the
  in-use state; `unknownTokens` renders the warning.
- **Regression guard**: editing the document does **not** mutate `localSchemas`
  (lock the data-loss fix so it cannot regress).
- Follow the canonical frontend test framework (no bespoke harness).

## 8. Verification (runtime, per the verify skill)

Drive the running app (Go API :8081, Vite :4173, login admin):

1. Open a draft template editor → panel lists computed catalog tokens with Insert.
2. Click `doc_code` → `{doc_code}` appears at the cursor; native chip decorates it.
3. Type a bogus `{nope}` → inline unknown-token warning appears; `doc_code` does not warn.
4. Network capture: editing the body issues **no** schema PUT that zeroes
   `placeholder_schema` (the prior corruption path is gone).
5. Capture screenshot + network as evidence.

## 9. Risks

- **Native insertion API shape.** `insertToken` depends on the vendor's insert path; if the
  `insertTemplateVariable` command is not directly reachable through the react ref, fall
  back to `agent.insertText(selectionPosition, "{key}")`. Resolved in the plan; either way
  it stays inside the adapter.
- **`getUsedTokens` source.** Uses `getTemplateTags(editorState)`; requires reaching the
  live ProseMirror state via the typed inner ref. Confined to the adapter.
