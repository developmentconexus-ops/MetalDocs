# Phase 2 — Data-flow trace: Mount + slot composition

**Operation:** Page mounts `<EditorChrome>` and renders eigenpal canvas inside, with all three overlay slots populated.
**Path traced:** `TemplateEditorPage` (the more elaborate of the two consumers).
**Method:** manual trace via Read/Grep; no AST. Codex blocked.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| Route | `/templates/:id/versions/:n` | (verify in `features/templates/routes.tsx`) |
| Page component | `TemplateEditorPage` | `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` |
| Editor primitive | `EditorChrome` | `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx:31` |
| Editor canvas | `DocxEditor` (eigenpal native) | `@eigenpal/docx-js-editor` |

### 2. Call chain

```
1. TemplateEditorPage.tsx:273   <main className={styles.canvas}>
   → renders <EditorChrome center={...} right={...} alert={...}>...</EditorChrome>

2. EditorChrome.tsx:31          function EditorChrome({left,center,right,alert,children,className})
   → returns <div className={styles.wrapper + className}>
       {left   && <div className={styles.overlayLeft}>{left}</div>}      // line :34
       {center && <div className={styles.overlayCenter}>{center}</div>}  // line :35
       {right  && <div className={styles.overlayRight}>{right}</div>}    // line :36
       {alert  && <div className={styles.overlayAlert}>{alert}</div>}    // line :37
       {children}                                                        // line :38

3. center slot (TemplateEditorPage.tsx:275-285):
   <span class=docTitle>{template.name}</span>
   <span class=docSep>·</span>
   <span class=docMeta>Template</span>
   <VersionBadge>v{versionNum}</VersionBadge>           → parts/VersionBadge.tsx:13
   {versionStatus && <StatusPill status=... />}         → components/ui/StatusPill (out-of-module)

4. right slot (TemplateEditorPage.tsx:286-317):
   <AutosaveStatus status={autosaveState} labels={AUTOSAVE_LABELS_PT} />  → parts/AutosaveStatus.tsx:28
   {isDraft && <input type=file ... />}                  // hidden file input — pointer-events stays auto
   {isDraft && <button class=primaryBtn>Importar .docx</button>}
   {isDraft && <button class=primaryBtn>Submeter para revisão</button>}

5. alert slot (TemplateEditorPage.tsx:318-329):
   submitMsg → <div role={'alert'|'status'} class=alertError|alertSuccess>{text}</div>
   importErr → <div role="alert" class=alertError>{importErr}</div>

6. children (TemplateEditorPage.tsx:331-340):
   <DocxEditor ref={editorRef} documentBuffer={...} readOnly={!isDraft} ... />
   → eigenpal mounts .ep-root inside .wrapper; :global() overrides apply
```

### 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|

n/a — mount is a UI render. No state machine transition is initiated by mount itself. Page-level state (autosave, version) is read, not written.

### 4. SQL touched

n/a — frontend operation, no DB calls in this trace.

### 5. Response shape

n/a — JSX render, no HTTP.

### 6. Cross-references

- **Slot rendering rule:** all four slots use truthy short-circuit (`{left && ...}`). Empty arrays / `0` / empty strings collapse to nothing — consumers must pass `null`/`false` to suppress, or any truthy fragment to render.
- **`pointer-events` discipline:** `.overlayCenter` sets `pointer-events:none` (`EditorChrome.module.css:43`). Interactive children inside center slot must opt back in with `pointer-events:auto`. Comment at lines 49–50 documents this; not enforced by type system or runtime.
- **Keyboard a11y:** even with `pointer-events:none`, focusable elements in center slot can still receive focus via Tab and activate via Enter/Space. No automated guard.
- **Right-slot z-index war:** `.overlayRight` has `z-index:100`, `.overlayLeft`/`.overlayCenter` use `z-index:60`. `.overlayAlert` shares `z-index:100`. Eigenpal's own toolbar runs under `z-index:50` (per its CSS). The chrome overlays consistently render above eigenpal's own toolbar.
- **`className` passthrough:** consumer `className` is appended to `.wrapper` via string concat at `:33`. No `clsx`/`classnames` helper used.

### 7. Symmetry note

`DocumentEditorPage.tsx:217-257` follows the identical pattern with:
- No `left` slot (back button is in a separate `<aside class=rail>` outside `EditorChrome`).
- Same `center` shape: `CodeChip` + title + `VersionBadge` + `StatusPill`.
- Smaller `right` slot: only `AutosaveStatus` + single primary button.
- No `alert` slot.

Tripwire pairing: **n/a — no DB calls in flow.**
