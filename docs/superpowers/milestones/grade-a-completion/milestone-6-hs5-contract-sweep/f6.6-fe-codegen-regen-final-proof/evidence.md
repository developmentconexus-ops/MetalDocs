# F6.6 Evidence

## Date
2026-06-19

## C1 — FE codegen regenerated

### Command
```
cd frontend/apps/web && npm run gen:api
```

### Output
```
> @metaldocs/web@0.1.0 gen:api
> openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts

✨ openapi-typescript 7.13.0
🚀 ../../../api/openapi/v1/openapi.yaml → src/lib/api-types/index.d.ts [122.4ms]
```

### Diff stat
```
frontend/apps/web/src/lib/api-types/index.d.ts | 91 +++++++++++++++++++++++---
 1 file changed, 83 insertions(+), 8 deletions(-)
```

### New types confirmed in diff
- `TemplateVersionEnvelope` ✓
- `ArchiveTemplateResponse` ✓
- `TemplateApprovalConfig` ✓
- `UpsertTemplateApprovalConfigResponse` ✓
- `GetTemplateResponse` ✓
- `GetTemplateDocxUrlResponse` ✓
- `TemplateAuditEvent` ✓
- `ListTemplateAuditResponse` ✓

All 8 new schemas from F6.1 and F6.2 are present as typed TypeScript interfaces.

## C2 — H-D Grep A = 0

### Command
```
grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go'
```

### Output
```
(no output — 0 matches)
```

**PASS — H-D Grep A = 0.**

## C3 — Wiki stamp updated

File: `wiki/architecture/api-contract.md`

Old line:
```
> **Last verified:** 2026-06-15 (F1.2 / ADR 0035 — …)
```

New line:
```
> **Last verified:** 2026-06-19 (M6 / F6.1–F6.6 — templates lifecycle + query 200 schemas declared; IAM admin/sessions/observability/memberships typed; security/taxonomy/catalog/schema typed; FE openapi-typescript regenerated; H-D Grep A = 0. Prior: 2026-06-15)
```

## C4 — Regression

### `go build ./...`
Exit 0. No output (clean).

### `go test -count=1 ./...`
All packages passed. Representative output:
```
ok  metaldocs/apps/api/cmd/metaldocs-api           29.945s
ok  metaldocs/apps/api/internal/wiring              4.032s
ok  metaldocs/apps/worker/cmd/metaldocs-worker     31.264s
ok  metaldocs/internal/modules/templates/delivery/http   6.254s
ok  metaldocs/internal/modules/iam/delivery/http        5.868s
ok  metaldocs/internal/modules/security/application     2.095s
ok  metaldocs/internal/modules/taxonomy/delivery/http   6.702s
... (all packages ok or [no test files])
```

**PASS — build and test regression clean.**
