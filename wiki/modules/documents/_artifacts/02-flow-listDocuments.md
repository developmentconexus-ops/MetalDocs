### 1. Entry point
| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `operationId: listDocuments` | (unclear: not found — `operationId: listDocumentsV2` exists at `api/openapi/v1/openapi.yaml:3156`) |
| Generated server stub | `ServerInterface.ListDocuments` | (unclear: not found — generated symbol is `ServerInterface.ListDocumentsV2` at `internal/modules/documents/api/api.gen.go:721`) |
| Handler | `Handler.listDocuments` | `internal/modules/documents/delivery/http/handler.go:145` (route registration at `:111` and `:82`) |

### 2. Call chain
1. `internal/modules/documents/delivery/http/handler.go:111` route registration — binds `GET /api/v1/documents` to handler.
   -> calls: `internal/modules/documents/delivery/http/handler.go:145` `(*Handler).listDocuments`
2. `internal/modules/documents/delivery/http/handler.go:145` `(*Handler).listDocuments` — enforces role gate (`admin` or `document_filler`), parses query/pagination, invokes app service, returns paginated JSON.
   -> calls: `internal/modules/documents/delivery/http/handler.go:200` `parseListOptions`
3. `internal/modules/documents/delivery/http/handler.go:200` `parseListOptions` — parses `page`, `pageSize`, `status`, `areaCode`, `profileCode`, `q`, `includeArchived`; enforces `page>=1`, `1<=pageSize<=50`.
   -> calls: local derivation at `internal/modules/documents/delivery/http/handler.go:259-262` (`effectiveUserID`/`opts.CreatedBy` for non-admin caller)
4. `internal/modules/documents/delivery/http/handler.go:160` `h.svc.ListDocumentsPaginated(...)` — application entry for list query.
   -> calls: `internal/modules/documents/application/service.go:525` `(*Service).ListDocumentsPaginated`
5. `internal/modules/documents/application/service.go:525` `(*Service).ListDocumentsPaginated` — reapplies user scoping (`opts.CreatedBy=userID` when `userID != ""`), queries list and total.
   -> calls: `internal/modules/documents/repository/repository.go:343` `(*Repository).ListDocumentsPaginated`
6. `internal/modules/documents/repository/repository.go:343` `(*Repository).ListDocumentsPaginated` — builds SQL filter + executes paginated `SELECT ... FROM documents ... ORDER BY updated_at DESC LIMIT/OFFSET`.
   -> returns rows to service
7. `internal/modules/documents/application/service.go:535` continues total count.
   -> calls: `internal/modules/documents/repository/repository.go:376` `(*Repository).CountDocuments`
8. `internal/modules/documents/repository/repository.go:376` `(*Repository).CountDocuments` — executes `SELECT COUNT(*) FROM documents WHERE ...` using same filter.
   -> returns total to service -> handler writes response at `internal/modules/documents/delivery/http/handler.go:167`

Tier-1 capability check in this flow:
- (unclear: no direct `Caps.CanDo`/`CapabilityChecker` call found in `listDocuments` handler/service path; authorization here is role gate + creator scoping)

Transaction boundary:
- None in this read path (`QueryContext`/`QueryRowContext` on `r.db`; no `BeginTx` in list/count path).

### 3. State changes
- None (read-only GET operation).

### 4. SQL touched
| DB call | File:line | Verb | Table(s) | auth-area arg |
|---|---|---|---|---|
| `ListDocumentsPaginated` | `internal/modules/documents/repository/repository.go:343` | `SELECT` | `documents` | Tenant scope via `tenant_id = $1`; optional creator scope via `created_by = $N` when `effectiveUserID`/`opts.CreatedBy` set |
| `CountDocuments` | `internal/modules/documents/repository/repository.go:376` | `SELECT COUNT(*)` | `documents` | Same filter path as above (`buildDocumentFilter` at `:310`) |

Tripwire pairing:
- `N/A` for these `SELECT`/`COUNT` reads (tripwire applies to mutations only).

### 5. Response shape
2xx body:
- Handler emits inline map at `internal/modules/documents/delivery/http/handler.go:167`:
  - `items`
  - `page`
  - `pageSize`
  - `total`
- OpenAPI/generated model counterpart: `DocumentListResponse` at `internal/modules/documents/api/api.gen.go:266`:
  - `Items []DocumentSummary \`json:"items"\``
  - `Page int \`json:"page"\``
  - `PageSize int \`json:"pageSize"\``
  - `Total int64 \`json:"total"\``

Error responses in `listDocuments`:
- `403 forbidden` when caller lacks role (`hasAnyRole` check) at `internal/modules/documents/delivery/http/handler.go:146-147` via `httpErr`.
- `400 bad request` for query parse/validation failures from `parseListOptions` at `internal/modules/documents/delivery/http/handler.go:154-157` via `httpErr`.
- Service/repo errors mapped by `mapErr` at `internal/modules/documents/delivery/http/handler.go:162-163`:
  - Could map to `403/404/400/410/422/409/500` depending on error type (`mapErr` table at `handler.go:958-1009`), then written via `httpErr` (`:1013`).

### 6. Cross-references
- Idempotency: yes in HTTP semantics for GET (read-only, no state change).
- Pagination: yes.
  - Inputs: `page`, `pageSize` in query.
  - Defaults: `page=1`, `pageSize=20` (`parseListOptions` at `handler.go:203-206`).
  - Cap: `pageSize <= 50` (`handler.go:225-227`; repo also caps defensively in `ListOptions.Limit()`).
  - Response metadata: `page`, `pageSize`, `total` in body (`handler.go:167-172`).
  - Cursor support: (unclear: not found; offset/limit pagination only).
- Audit log emission: none in this list path (no `audit.Write` call in `listDocuments`, service list methods, or repository list/count methods).
