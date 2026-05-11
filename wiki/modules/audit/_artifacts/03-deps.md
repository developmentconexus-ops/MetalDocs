# Audit module — Cross-deps (Phase 3)

_Module path: `internal/modules/audit`_
_Last verified: 2026-05-10_

---

## §1 Imports OUT

Internal MetalDocs packages imported by the audit module itself. (Self-imports and stdlib/vendor excluded.)

| Imported package | First seen in (file:line) | Symbols used | Purpose |
|---|---|---|---|
| `internal/modules/audit/domain` | `application/service.go:7` | `Reader`, `ListEventsQuery`, `Event` | Service reads via port interface |
| `internal/modules/audit/domain` | `infrastructure/postgres/writer.go:9` | `Event`, `ListEventsQuery` | Postgres impl satisfies port |
| `internal/modules/audit/domain` | `infrastructure/memory/writer.go:8` | `Event`, `ListEventsQuery` | Memory impl satisfies port |
| `internal/modules/audit/application` | `delivery/http/handler.go:10` | `Service` | HTTP handler wraps app service |
| `internal/modules/audit/domain` | `delivery/http/handler.go:11` | `ListEventsQuery` | Query struct for handler |
| `internal/platform/httpresponse` | `delivery/http/handler.go:12` | `WriteJSON` | Shared JSON response helper |

**OUT count: 3 distinct external internal packages** (`audit/domain` counted once; `platform/httpresponse`; `audit/application` from delivery layer counts as intra-module cross-layer).

---

## §2 Imports IN

Other internal packages that import the audit module. Listed by consumer with all symbols used.

### 2a. `internal/platform/bootstrap`

**File:** `internal/platform/bootstrap/api.go`

| Line | Import alias | Symbol used | Context |
|---|---|---|---|
| 13 | `auditdomain` | `auditdomain.Writer`, `auditdomain.Reader` | `APIDependencies` struct field types |
| 14 | `auditmemory` | `auditmemory.NewWriter()` | dev-mode wire (memory store) |
| 15 | `auditpg` | `auditpg.NewWriter(db)` | production wire (postgres store) |

`AuditWriter` and `AuditReader` are fields on the `APIDependencies` struct (`api.go:37-38`).

Production path (`api.go:100-101`):
```
AuditWriter: auditpg.NewWriter(db)
AuditReader: auditpg.NewWriter(db)   // NOTE: same *Writer satisfies both interfaces
```
Dev path (`api.go:124-130`):
```
auditStore := auditmemory.NewWriter()
AuditWriter: auditStore
AuditReader: auditStore
```

### 2b. `internal/modules/iam/delivery/http` — `AdminHandler`

**File:** `internal/modules/iam/delivery/http/admin_handler.go`

| Line | Alias | Symbol | Usage |
|---|---|---|---|
| 12 | `auditdomain` | `auditdomain.Writer` | field type `audit auditdomain.Writer` (line 32) |
| — | — | `auditdomain.Reader` | field type `auditReader auditdomain.Reader` (line 33) |
| — | — | `auditdomain.Event` | constructed in `recordAudit` helper (line 457) |
| — | — | `auditdomain.ListEventsQuery` | passed to `h.auditReader.ListEvents(...)` (line 128) |

Constructor signature (`admin_handler.go:69`):
```go
func NewAdminHandler(service *iamapp.AdminService, authService UserAdminService, auditWriter ...auditdomain.Writer) *AdminHandler
```
`WithAuditReader` method (`admin_handler.go:77`):
```go
func (h *AdminHandler) WithAuditReader(reader auditdomain.Reader) *AdminHandler
```

`.Record(` call sites in `admin_handler.go`:
- `admin_handler.go:457` — `_ = h.audit.Record(r.Context(), auditdomain.Event{...})` — **fire-and-forget** (error discarded)

`.ListEvents(` call sites:
- `admin_handler.go:128` — `events, err := h.auditReader.ListEvents(r.Context(), auditdomain.ListEventsQuery{Limit: 25})` — error-checked

### 2c. `apps/api/cmd/metaldocs-api` — main / documentsV2AuditAdapter

**File:** `apps/api/cmd/metaldocs-api/main.go`

| Line | Alias | Symbol | Usage |
|---|---|---|---|
| 21 | `auditdomain` | `auditdomain.Writer`, `auditdomain.Event` | adapter struct field + Event construction |
| 38 | `auditapp` | `auditapp.NewService(deps.AuditReader)` | constructs `*application.Service` |
| 39 | `auditdelivery` | `auditdelivery.NewHandler(auditService)` | constructs `*delivery/http.Handler` |

`documentsV2AuditAdapter` (`main.go:445-479`) wraps `auditdomain.Writer` behind `documents/application.Audit` interface:
```go
type documentsV2AuditAdapter struct { writer auditdomain.Writer }
func (a *documentsV2AuditAdapter) Write(ctx, tenantID, actorID, action, docID string, meta any)
```
`.Record(` call in adapter (`main.go:467`):
```go
if err := a.writer.Record(ctx, auditdomain.Event{...}); err != nil { log.Printf(...) }
```
**Error-checked** (logged, not propagated).

---

## §3 DI / wiring touchpoints

### `internal/platform/bootstrap/api.go`

| Line | What is wired |
|---|---|
| 100 | `AuditWriter: auditpg.NewWriter(db)` — production postgres writer constructed |
| 101 | `AuditReader: auditpg.NewWriter(db)` — same `*postgres.Writer` satisfies Reader interface |
| 124 | `auditStore := auditmemory.NewWriter()` — dev-mode memory writer |
| 129 | `AuditWriter: auditStore` |
| 130 | `AuditReader: auditStore` |

### `apps/api/cmd/metaldocs-api/main.go`

| Line | What is wired |
|---|---|
| 153 | `auditService := auditapp.NewService(deps.AuditReader)` |
| 155 | `auditHandler := auditdelivery.NewHandler(auditService)` |
| 182 | `iamAdminHandler := iamdelivery.NewAdminHandler(iamAdminService, authService, deps.AuditWriter)` |
| 183 | `.WithAuditReader(deps.AuditReader)` |
| 193 | `auditHandler.RegisterRoutes(mux)` — registers `GET /api/v1/audit/events` |
| 275 | `Audit: newDocumentsV2AuditAdapter(deps.AuditWriter)` — adapter injected into documents service |

---

## §4 Configuration surface

No env vars, config struct fields, or feature flags referencing audit (retention, max payload size, etc.) found in `internal/` or `apps/`.

**none**

---

## §5 Test surface

| Test file | Subject (file under test) | Kind |
|---|---|---|
| `tests/unit/audit_http_handler_test.go` | `delivery/http/handler.go` + `application/service.go` | unit |

No other `_test.go` files in the main worktree import the audit module directly. The `documents/application` test files use a local `noopAudit` / `fakeAudit` struct that satisfies `documents/application.Audit` (not `auditdomain`) — they do not import `modules/audit`.

---

## §6 Emitted action string catalogue

All `Action:` / action-string literals passed through the audit path, by consumer:

### Consumer: `internal/modules/iam/delivery/http/admin_handler.go`

| Call site | Action string | ResourceType | Fire-and-forget? |
|---|---|---|---|
| `admin_handler.go:307` via `recordAudit` | `"iam.user.updated"` | `"user"` | yes (`_ = h.audit.Record(...)`) |
| `admin_handler.go:398` via `recordAudit` | `"iam.user.roles.replaced"` | `"user"` | yes |
| `admin_handler.go:417` via `recordAudit` | `"auth.user.password_reset"` | `"user"` | yes |

### Consumer: `internal/modules/documents/application/service.go` (via `documentsV2AuditAdapter`)

| Call site (service.go) | Action string | ResourceType | Fire-and-forget? |
|---|---|---|---|
| `service.go:341` | `"document.created"` | `"document"` | yes (Write returns void) |
| `service.go:362` | `"document.created"` | `"document"` | yes |
| `service.go:579` | `"document.renamed"` | `"document"` | yes |
| `service.go:688` | `"document.autosaved"` | `"document"` | yes |
| `service.go:701` | `"session.acquired"` | `"document"` | yes |
| `service.go:713` | `"session.released"` | `"document"` | yes |
| `service.go:721` | `"session.force_released"` | `"document"` | yes |
| `service.go:730` | `"document.checkpoint_created"` | `"document"` | yes |
| `service.go:743` | `"document.checkpoint_restored"` | `"document"` | yes |
| `service.go:757` | `"document.finalized"` | `"document"` | yes |
| `service.go:765` | `"document.archived"` | `"document"` | yes |

### Consumer: `internal/modules/documents/application/export_service.go` (via `documentsV2AuditAdapter`)

| Call site (export_service.go) | Action string | ResourceType | Fire-and-forget? |
|---|---|---|---|
| `export_service.go:78` | `"export.pdf_generated"` (cached=true) | `"document"` | yes |
| `export_service.go:122` | `"export.pdf_generated"` (cached=false) | `"document"` | yes |
| `export_service.go:153` | `"export.docx_downloaded"` | `"document"` | yes |

### Consumer: test fixture (`tests/unit/audit_http_handler_test.go`)

| Call site | Action string | Fire-and-forget? |
|---|---|---|
| `audit_http_handler_test.go:25` | `"user.created"` | yes (`_ = store.Record(...)`) |
| `audit_http_handler_test.go:35` | `"user.deleted"` | yes |

---

**Total unique action strings: 15**
(`iam.user.updated`, `iam.user.roles.replaced`, `auth.user.password_reset`, `document.created`, `document.renamed`, `document.autosaved`, `session.acquired`, `session.released`, `session.force_released`, `document.checkpoint_created`, `document.checkpoint_restored`, `document.finalized`, `document.archived`, `export.pdf_generated`, `export.docx_downloaded`)
Test-only: `user.created`, `user.deleted` (not emitted by production code).
