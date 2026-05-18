# Sync Log — editor-ui-eigenpal

## 2026-05-17 - active v2 reference memory sync

- **Context:** post-merge scan found the editor-ui C4 context still pointing at `/api/v2/documents` even though production route truth is `/api/v1/documents`.
- **Mode:** lite patch.
- **Affected-surface scan:** editor-ui module doc C4 context only.
- **Routes/API:** no adapter behavior or backend route change; relationship label corrected to `/api/v1/documents`.
- **Runtime flows:** none.
- **Persistence:** none.
- **Debt/backlog:** no T/R rows opened or closed.
- **Tally gate:** preflight PASS before edits; final tally recorded in session output.
- **Patched files:** `wiki/modules/editor-ui-eigenpal.md`; `wiki/modules/editor-ui-eigenpal/_artifacts/sync-log.md`.

## 2026-05-17 - blank editable Eigenpal mount

- **Context:** uncommitted diff: packages/editor-ui blank no-buffer mount behavior plus tests.
- **Mode:** structural refresh
- **Anchors moved:** none
- **Public surface:** no exported prop/type changed; adapter now passes internal `document={createEmptyDocument()}` for editable no-buffer mounts
- **Routes/API:** none
- **Runtime flows:** plugin/mount flow updated for blank document seed
- **Persistence:** none; no blank DOCX is written until user autosaves/imports
- **Dependencies:** OUT-edge updated for `@eigenpal/docx-js-editor/core:createEmptyDocument`; IN-edge notes updated for template runtime consumer
- **T-NNN touched:** none
- **R-NNN touched:** R-003 backlog status reconciled to closed 2026-05-11 (stale backlog fact)
- **Counts after:** Critical=1 Major=2 Minor=5; missing-ADR=6
- **Tally gate:** PASS preflight; final PASS after sync
- **Patched files:** `wiki/modules/editor-ui-eigenpal.md`; `wiki/modules/editor-ui-eigenpal/_artifacts/02-flow-plugin-registration.md`; `wiki/modules/editor-ui-eigenpal/_artifacts/03-deps.md`; `wiki/backlog/editor-ui-eigenpal-refactor.md`; `wiki/modules/editor-ui-eigenpal/_artifacts/sync-log.md`

## 2026-05-16 - documents comments hook path polish

- **Context:** uncommitted diff: documents editor hooks moved from `features/documents/hooks/editor`.
- **Mode:** lite patch
- **Anchors moved:** hook path references only
- **Public surface:** none
- **Routes/API:** none
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** editor-ui-eigenpal type-only dependency paths updated
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=2 Minor=5; missing-ADR=6
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/editor-ui-eigenpal/_artifacts/03-deps.md`; `wiki/modules/editor-ui-eigenpal/_artifacts/sync-log.md`

| Date | Change context | Files patched | T-NNN / R-NNN affected | Tally |
|---|---|---|---|---|
| 2026-05-11 | Plan 3 — tarball restored (T-001 resolved) | `wiki/modules/editor-ui-eigenpal.md` | T-001 closure reflected in §2 vendor note + §11 Top 3; R-001 already closed in backlog | PASS |
