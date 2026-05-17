# Phase 4 — Persistence Map

> Scope: tables, FKs, triggers, GUC reads, tripwire pairing.

## Result: n/a

`editor-ui-eigenpal` is an FE adapter package. Zero database surface:

- No SQL queries.
- No migrations (`packages/editor-ui/` contains no `*.sql`).
- No GUC reads (`SET LOCAL ...`, `current_setting(...)` — none).
- No tripwire pairing (n/a — no tx writer).
- No connection to `internal/platform/db/...` or any backend persistence module.

## Indirect persistence touch

The adapter emits `ArrayBuffer` (DOCX bytes) to the parent's `onAutoSave` callback. The parent (`DocumentEditorPage`) is responsible for uploading via `templates`/`documents` backend modules, which own their own tables. Persistence concerns belong in those module docs:

- `wiki/modules/documents.md` — `public.documents`, snapshot columns, finalize tripwire.
- `wiki/modules/templates.md` — `public.template_versions`, DOCX storage in MinIO.

## In-memory state held by the adapter

| Ref | Type | Lifetime |
|---|---|---|
| `inner.current` (`DocxEditorRef`) | mutable ref | component instance |
| `onAutoSaveRef.current` | callback ref | component instance |
| `timerRef.current` | `setTimeout` handle | component instance; cleared on unmount |
| `inFlightRef.current` | boolean | component instance |

No persistence guarantees, no replay. Lost on tab close before `cb(buf)` resolves — accepted, parent handles 409/etag conflicts.
