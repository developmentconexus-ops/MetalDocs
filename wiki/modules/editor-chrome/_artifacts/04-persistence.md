# Phase 4 — Persistence map

**n/a — frontend module, no persistence surface.**

editor-chrome is a React primitive (`features/shared/components/editor-chrome/`). It owns no Postgres tables, no migrations, no triggers, no GUCs, no indexes, no SQL of any kind.

### 1. Tables owned
None.

### 2. Tables read or written but NOT owned
None directly. Consumer pages (`TemplateEditorPage`, `DocumentEditorPage`) write to `templates` / `template_versions` and `documents` via REST → the templates / documents backend modules. Those writes are recorded in the respective module persistence maps (`wiki/modules/templates.md`, `wiki/modules/documents.md`). editor-chrome itself is a UI shell with no data-layer side effects.

### 3. Triggers, GUCs, functions
None.

### 4. Indexes
None.

### 5. Tripwire pairing audit
n/a — no repo methods.

### 6. Migration history
None.

### Browser-side storage (informational only)

The consumer hook `useDocumentAutosave` (`features/documents/hooks/v2/useDocumentAutosave.ts`) uses IndexedDB for pending-save crash recovery via `useIndexedDBRestore`. That storage is owned by the `documents` feature, not by editor-chrome. The chrome only renders the visual indicator of the hook's state via `AutosaveStatus`.

The `DocumentEditorPage` reads/writes `localStorage` key `editor-sidebar-open` (`DocumentEditorPage.tsx:181, 264`) for sidebar open/closed state; this is page state, not chrome state.

editor-chrome has **zero** direct persistence — Postgres, IndexedDB, or localStorage.
