# F6.6 Spec — FE Codegen Regen + Wiki Stamp + H-D Grep A = 0 Proof

## Feature
F6.6 — Final feature of Milestone 6 (HS-5 contract sweep).

## Contract

### C1 — FE codegen regenerated
`npm run gen:api` (from `frontend/apps/web/`) must complete cleanly and produce a
non-empty diff in `frontend/apps/web/src/lib/api-types/index.d.ts`. The diff must
include the 8 new TypeScript types introduced by F6.1 and F6.2:
- `TemplateVersionEnvelope`
- `ArchiveTemplateResponse`
- `TemplateApprovalConfig`
- `UpsertTemplateApprovalConfigResponse`
- `GetTemplateResponse`
- `GetTemplateDocxUrlResponse`
- `TemplateAuditEvent`
- `ListTemplateAuditResponse`

### C2 — H-D Grep A = 0
```
grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go'
```
Must return 0 matches.

### C3 — Wiki stamp updated
`wiki/architecture/api-contract.md` Last-verified line must be updated to
`2026-06-19 (M6 / F6.1–F6.6 …)`.

### C4 — Regression clean
`go build ./...` and `go test -count=1 ./...` both exit 0.
