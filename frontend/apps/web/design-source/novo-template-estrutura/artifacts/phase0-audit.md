# Phase 0 — Audit · novo-template-estrutura

**Confirmed by user:** 2026-05-09. Path A-lite — wizard captures starting-point choice + selected file (all mocked client-side); real upload + extraction live in editor post-create.

## Keep / Cut / Defer

| Element | Map → | Decision | Reason |
|---|---|---|---|
| 2 starting-point cards (`.docx` / `Em branco`) | `startingPoint: 'docx' \| 'blank'` reducer field | **Keep** | Drives next-step UX + post-create redirect. |
| Card visual: thumbnail mock + check-icon on selected | presentation only | **Keep** | Helps user confirm choice. |
| File-selected row (`inspecao-mp-base.docx · 147 KB`) | mock state — `selectedDocxName: string \| null` | **Keep (mocked)** | UX needs feedback after picking docx; real upload deferred. |
| "Substituir" button on file row | reset `selectedDocxName` → re-pick | **Keep** | Cheap. |
| Placeholders block (7 chips, kind, auto-flag, ⚡) | — | **Cut** | No `extract` endpoint; only `publish` returns token analysis. Backend has no auto-fill flag. |
| "2 auto-fill · 5 manuais" counter | — | **Cut** | Same reason — no data. |
| Info banner "ℹ️ próximos passos refinados no editor" | — | **Cut** | UX assumption — editor flow not yet implemented for templates wizard handoff. |
| File metadata "147 KB · processado em 1.2s" | — | **Cut** | Implies server processing; mock would lie. Show only filename. |

## Backend gaps → Backlog rows

| Item | Backlog row |
|---|---|
| Real docx upload via presigned URL | `wiki/backlog/novo-template-wizard.md` → `step3-docx-upload` |
| Placeholder extraction | `step3-placeholder-extract` |
| Editor handoff for blank/docx start | `step3-editor-handoff` |

## Mock surface

- Native `<input type="file" accept=".docx">` for picker.
- On change: store `file.name` + `file.size` (formatted KB) in local component state. No upload, no parsing.
- "Substituir" clears state.
- TODO comments + backlog refs at every mock site.

## Out of scope

- Drag-drop area (design's `phase === 'empty'` state). Native picker is enough for the click-to-upload UX. Add later if needed.
