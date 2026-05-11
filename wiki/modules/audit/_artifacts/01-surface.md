# 1. File tree

- `application/.gitkeep` — type: marker
- `application/service.go` — type: service
- `delivery/http/.gitkeep` — type: marker
- `delivery/http/handler.go` — type: handler, router
- `domain/.gitkeep` — type: marker
- `domain/port.go` — type: model, dto, interface(port)
- `infrastructure/.gitkeep` — type: marker
- `infrastructure/memory/writer.go` — type: repo(writer)
- `infrastructure/postgres/writer.go` — type: repo(writer)

# 2. Public surface

| File:line | Kind | Name | Signature / receiver |
|---|---|---|---|
| `internal/modules/audit/application/service.go:10` | type | `Service` | `type Service struct { ... }` |
| `internal/modules/audit/application/service.go:14` | func | `NewService` | `func NewService(reader domain.Reader) *Service` |
| `internal/modules/audit/application/service.go:18` | method | `ListEvents` | `func (s *Service) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error)` |
| `internal/modules/audit/delivery/http/handler.go:15` | type | `Handler` | `type Handler struct { ... }` |
| `internal/modules/audit/delivery/http/handler.go:19` | type | `EventResponse` | `type EventResponse struct { ... }` |
| `internal/modules/audit/delivery/http/handler.go:30` | func | `NewHandler` | `func NewHandler(service *application.Service) *Handler` |
| `internal/modules/audit/delivery/http/handler.go:34` | method | `RegisterRoutes` | `func (h *Handler) RegisterRoutes(mux *http.ServeMux)` |
| `internal/modules/audit/domain/port.go:8` | type | `Event` | `type Event struct { ... }` |
| `internal/modules/audit/domain/port.go:19` | type | `ListEventsQuery` | `type ListEventsQuery struct { ... }` |
| `internal/modules/audit/domain/port.go:25` | type(interface) | `Writer` | `type Writer interface { Record(ctx context.Context, event Event) error }` |
| `internal/modules/audit/domain/port.go:29` | type(interface) | `Reader` | `type Reader interface { ListEvents(ctx context.Context, query ListEventsQuery) ([]Event, error) }` |
| `internal/modules/audit/infrastructure/memory/writer.go:11` | type | `Writer` | `type Writer struct { ... }` |
| `internal/modules/audit/infrastructure/memory/writer.go:16` | func | `NewWriter` | `func NewWriter() *Writer` |
| `internal/modules/audit/infrastructure/memory/writer.go:20` | method | `Record` | `func (w *Writer) Record(_ context.Context, event domain.Event) error` |
| `internal/modules/audit/infrastructure/memory/writer.go:27` | method | `ListEvents` | `func (w *Writer) ListEvents(_ context.Context, query domain.ListEventsQuery) ([]domain.Event, error)` |
| `internal/modules/audit/infrastructure/postgres/writer.go:12` | type | `Writer` | `type Writer struct { ... }` |
| `internal/modules/audit/infrastructure/postgres/writer.go:16` | func | `NewWriter` | `func NewWriter(db *sql.DB) *Writer` |
| `internal/modules/audit/infrastructure/postgres/writer.go:20` | method | `Record` | `func (w *Writer) Record(ctx context.Context, event domain.Event) error` |
| `internal/modules/audit/infrastructure/postgres/writer.go:44` | method | `ListEvents` | `func (w *Writer) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error)` |

# 3. HTTP operations

| Method | Path | Handler function | Auth middleware | Route registration |
|---|---|---|---|---|
| `GET` | `/api/v1/audit/events` | `h.handleEvents` | `(none)` | `internal/modules/audit/delivery/http/handler.go:35` |

# 4. Migration list

## migrations/0004_init_audit_events.sql
- filename: `migrations/0004_init_audit_events.sql`
- table name: `metaldocs.audit_events`
- DDL statements:
```sql
CREATE TABLE IF NOT EXISTS metaldocs.audit_events (
  id TEXT PRIMARY KEY,
  occurred_at TIMESTAMPTZ NOT NULL,
  actor_id TEXT NOT NULL,
  action TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  trace_id TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_events_occurred_at ON metaldocs.audit_events (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_actor_time ON metaldocs.audit_events (actor_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_resource_time ON metaldocs.audit_events (resource_type, resource_id, occurred_at DESC);
```

## migrations/0005_grant_workflow_audit_privileges.sql
- filename: `migrations/0005_grant_workflow_audit_privileges.sql`
- table name: `metaldocs.audit_events`
- DDL statements:
```sql
GRANT UPDATE ON TABLE metaldocs.documents TO metaldocs_app;
GRANT INSERT ON TABLE metaldocs.audit_events TO metaldocs_app;
```