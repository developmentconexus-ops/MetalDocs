# Phase 2 — Data-flow trace: Eigenpal CSS overrides scope

**Operation:** `EditorChrome` wrapper styles reach into eigenpal DOM via `:global(.ep-root ...)` to enforce MetalDocs visual contract.
**Path traced:** purely CSS — no JS layer. `EditorChrome.module.css` ↔ eigenpal-rendered DOM nodes.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| Wrapper class | `.wrapper` | `EditorChrome.module.css:9` |
| Override scope prefix | `.wrapper :global(.ep-root ...)` | `EditorChrome.module.css:160–242` |
| Eigenpal DOM root | `.ep-root` (data attribute / class set by eigenpal `DocxEditor`) | `@eigenpal/docx-js-editor` (vendored at `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`) |
| Eigenpal anchors targeted | `[data-testid=*]` / `[role=*]` / SVG path fills | see §1.7 of `01-surface.md` |

### 2. Call chain

CSS does not call. Cascade resolution at paint:

```
1. CSS Modules transformer (Vite) scopes .wrapper / .overlayLeft / etc.
   to module-local class hashes (e.g. _EditorChrome_module__wrapper_abc123).

2. :global(...) bypasses scoping for the inner selector but keeps the
   outer .wrapper local. Effective selector at runtime:
     ._wrapper_abc123 .ep-root [data-testid="title-bar"] { ... !important }

3. Eigenpal renders its tree at runtime under <div class=ep-root>.
   No JS handshake with editor-chrome — pure DOM-tree containment.

4. Browser CSS engine matches override selectors against eigenpal DOM
   when descendant of .wrapper.

5. !important on every override defeats eigenpal's own inline styles
   and !important rules from eigenpal's stylesheet.
```

### 3. State changes

n/a — CSS is stateless.

### 4. SQL touched

n/a.

### 5. Response shape

n/a.

### 6. Cross-references

- **Coupling fragility:** every selector at `EditorChrome.module.css:160–225` is anchored to eigenpal-internal selectors (`[data-testid="title-bar"]`, `[data-testid="formatting-bar"]`, `[data-testid="font-size-display"]`, `[data-testid="font-size-input"]`, `.docx-advanced-color-picker-dropdown`) or to hardcoded SVG fill hex values (`svg path[fill="#cbd5e1"]`, `svg path[fill="#94a3b8"]`). If eigenpal renames an attribute, ships a different doc-icon palette, or changes the formatting-bar geometry, overrides silently no-op. No version-check, no CI guard, no runtime assertion.

- **Eigenpal version pin:** controlled fork is `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. Override compatibility is implicit in the pin. ADR 0001 + `references/eigenpal-controlled-package.md` cover the refresh process.

- **Bug-fix as comment, not code:** the `overflow: hidden !important` on `[data-testid="formatting-bar"]` at line 178 is paired with a code-comment explanation at lines 170–173 noting that `overflow-x:hidden` alone would force `overflow-y:auto` and create a scroll container that closes eigenpal's font-size dropdown on capture-phase scroll events. The fix is documented in CSS, not in a test or eigenpal-side patch.

- **`!important` density:** every line from 161 to 225 ends with `!important` (counted: 31 occurrences). Removing wrapper scoping or eigenpal stylesheet changes will not let cascade resolve naturally — every property is forced.

- **No JS coupling:** the module imports nothing from `@eigenpal/docx-js-editor`. Coupling is exclusively via CSS descendant selectors and DOM containment. Consumers (pages) import eigenpal; the chrome module does not.

- **Audit log emission:** n/a.

Tripwire pairing: **n/a — no DB writes.**
