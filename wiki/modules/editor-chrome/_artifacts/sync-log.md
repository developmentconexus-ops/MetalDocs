# Sync log - editor-chrome

## 2026-05-16 - documents editor hook path polish

- **Context:** uncommitted diff: documents editor hooks moved from `features/documents/hooks/editor`.
- **Mode:** lite patch
- **Anchors moved:** hook path references only
- **Public surface:** none
- **Routes/API:** none
- **Runtime flows:** autosave artifact path references updated
- **Persistence:** IndexedDB owner path reference updated
- **Dependencies:** documents hook dependency path updated
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=0 Major=4 Minor=5; missing-ADR=9
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/editor-chrome/_artifacts/01-surface.md`; `wiki/modules/editor-chrome/_artifacts/02-flow-autosave.md`; `wiki/modules/editor-chrome/_artifacts/03-deps.md`; `wiki/modules/editor-chrome/_artifacts/04-persistence.md`; `wiki/modules/editor-chrome/_artifacts/sync-log.md`

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.
