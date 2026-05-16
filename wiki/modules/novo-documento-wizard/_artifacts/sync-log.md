## 2026-05-15 - D4 hard-cutover route wording sync

- **Context:** Worker E wiki/docs lane cutover refresh for Novo Documento wizard route and workflow references.
- **Mode:** lite patch
- **Anchors moved:** wizard route wording only
- **Public surface:** canonical wizard route is `/documents/new`
- **Routes/API:** no endpoint behavior changes; docs aligned to current route truth
- **Persistence:** none
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=0 Major=2 Minor=1; missing-ADR=n/a
- **Tally gate:** pending
- **Patched files:** `wiki/modules/novo-documento-wizard.md`; `wiki/modules/novo-documento-wizard/_artifacts/sync-log.md`
# Sync log - novo-documento-wizard

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-15 - Blank-template create path runtime sync

- **Context:** uncommitted Novo Documento blank-template runtime repair and browser smoke
- **Mode:** lite patch
- **Anchors moved:** `Last verified` stamp only
- **Public surface:** Step 3 blank-template card documented as selectable when profile templates are empty
- **Routes/API:** wizard documents `GET /api/v1/templates/system/blank` plus `POST /api/v1/controlled-documents`
- **Runtime flows:** system blank endpoint -> Registry atomic create -> Documents v2 draft/editor path documented
- **Persistence:** none in wizard module; persistence is owned by Registry/Documents v2
- **Dependencies:** Registry and Documents v2 dependency clarified
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=0 Major=2 Minor=1; missing-ADR=n/a
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/novo-documento-wizard.md`; `wiki/modules/novo-documento-wizard/_artifacts/00-context.md`; `wiki/modules/novo-documento-wizard/_artifacts/sync-log.md`
