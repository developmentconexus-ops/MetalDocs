# 02 - IAM admin recent-audit-events READ flow

## §1. Entry point

- HTTP route registration: `mux.HandleFunc("/api/v1/iam/admin/overview", h.handleAdminOverview)` in `AdminHandler.RegisterRoutes` at `internal/modules/iam/delivery/http/admin_handler.go:82-85`.
- Handler entry for this operation: `func (h *AdminHandler) handleAdminOverview(w http.ResponseWriter, r *http.Request)` at `internal/modules/iam/delivery/http/admin_handler.go:99`.
- Internal reader injection point: `func (h *AdminHandler) WithAuditReader(reader auditdomain.Reader) *AdminHandler` at `internal/modules/iam/delivery/http/admin_handler.go:77`, assignment `h.auditReader = reader` at `internal/modules/iam/delivery/http/admin_handler.go:78`.

## §2. Call chain (IAM -> audit module)

1. IAM HTTP route dispatch
   - `internal/modules/iam/delivery/http/admin_handler.go:85`
   - `/api/v1/iam/admin/overview` is bound to `h.handleAdminOverview`.

2. IAM admin overview handler executes audit read
   - `internal/modules/iam/delivery/http/admin_handler.go:99`
   - `handleAdminOverview` checks `h.auditReader != nil` (`:127`) and calls:
   - `h.auditReader.ListEvents(r.Context(), auditdomain.ListEventsQuery{Limit: 25})` (`:128`).

3. Cross-module port boundary (IN-edge into audit module)
   - `internal/modules/audit/domain/port.go:29-30`
   - `type Reader interface { ListEvents(ctx context.Context, query ListEventsQuery) ([]Event, error) }`.

4. Concrete adapter implementation in audit module
   - `internal/modules/audit/infrastructure/postgres/writer.go:44`
   - `func (w *Writer) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error)`.

5. DB query execution
   - SQL statement declared at `internal/modules/audit/infrastructure/postgres/writer.go:50-57`.
   - `w.db.QueryContext(...)` call at `internal/modules/audit/infrastructure/postgres/writer.go:59`.
   - Row mapping via `rows.Scan(...)` at `internal/modules/audit/infrastructure/postgres/writer.go:68`.

6. Bootstrap wiring for Reader dependency
   - Reader provisioned in platform dependencies as `AuditReader auditdomain.Reader` (`internal/platform/bootstrap/api.go:38`), set to:
   - `AuditReader: auditpg.NewWriter(db)` (`internal/platform/bootstrap/api.go:101`) in postgres mode, or `AuditReader: auditStore` (`internal/platform/bootstrap/api.go:130`) in memory mode.
   - IAM handler receives it via:
   - `iamdelivery.NewAdminHandler(..., deps.AuditWriter).WithAuditReader(deps.AuditReader)` at `apps/api/cmd/metaldocs-api/main.go:182-183`.

## §3. State changes

none (read-only path)

## §4. SQL touched

- File: `internal/modules/audit/infrastructure/postgres/writer.go`
- Method: `(*Writer).ListEvents` (`:44`)
- Query text: `SELECT ... FROM metaldocs.audit_events ... ORDER BY occurred_at DESC, id DESC LIMIT $3` (`:50-57`)
- Driver call: `w.db.QueryContext(...)` (`:59`)
- Tables read: `metaldocs.audit_events`
- Writes: none

## §5. Response shape

n/a - IAM owns the response shape (cross-module IN-edge; audit module does not own the HTTP response)

## §6. Audit emission

none - this is a READ path; no event emitted

## Key observation

This flow is a cross-module IN-edge into `internal/modules/audit` via the `auditdomain.Reader` port (`internal/modules/audit/domain/port.go:29-30`). IAM receives the dependency by setter injection (`h.auditReader = reader` in `WithAuditReader`) and bootstrap wiring connects it from platform deps (`deps.AuditReader`) into IAM handler construction (`apps/api/cmd/metaldocs-api/main.go:182-183`).
