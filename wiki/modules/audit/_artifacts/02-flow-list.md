# Phase 2 Data-flow Trace — GET /api/v1/audit/events

## §1 Entry point / OpenAPI op
- OpenAPI path present: `api/openapi/v1/openapi.yaml:1058` ? `  /audit/events:`
- OpenAPI method present: `api/openapi/v1/openapi.yaml:1059` ? `    get:`
- OpenAPI operationId: `(unclear: no operationId field under /audit/events get in api/openapi/v1/openapi.yaml:1058-1103)`
- Handler registration: `internal/modules/audit/delivery/http/handler.go:34-35`
  - `func (h *Handler) RegisterRoutes(mux *http.ServeMux) {`
  - `	mux.HandleFunc("/api/v1/audit/events", h.handleEvents)`
- Handler entry: `internal/modules/audit/delivery/http/handler.go:38` ? `func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {`

## §2 Call chain
1. `internal/modules/audit/delivery/http/handler.go:38` `func (h *Handler) handleEvents...`
   - calls `internal/modules/audit/application/service.go:18` via `internal/modules/audit/delivery/http/handler.go:54`
   - exact call site:
     - `items, err := h.service.ListEvents(r.Context(), domain.ListEventsQuery{`
2. `internal/modules/audit/application/service.go:18` `func (s *Service) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {`
   - limit clamp logic (exact lines):
     - `internal/modules/audit/application/service.go:28` `	if normalized.Limit <= 0 {`
     - `internal/modules/audit/application/service.go:29` `		normalized.Limit = 50`
     - `internal/modules/audit/application/service.go:30` `	}`
     - `internal/modules/audit/application/service.go:31` `	if normalized.Limit > 200 {`
     - `internal/modules/audit/application/service.go:32` `		normalized.Limit = 200`
     - `internal/modules/audit/application/service.go:33` `	}`
   - calls `internal/modules/audit/infrastructure/postgres/writer.go:44` via `internal/modules/audit/application/service.go:35`
     - `return s.reader.ListEvents(ctx, normalized)`
3. `internal/modules/audit/infrastructure/postgres/writer.go:44` `func (w *Writer) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {`
   - SQL execution call:
     - `internal/modules/audit/infrastructure/postgres/writer.go:59` `rows, err := w.db.QueryContext(ctx, q, strings.TrimSpace(query.ResourceType), strings.TrimSpace(query.ResourceID), limit)`

## §3 State changes
- none

## §4 SQL touched
- Query definition in `internal/modules/audit/infrastructure/postgres/writer.go:50-57`:
```sql
const q = `
SELECT id, occurred_at, actor_id, action, resource_type, resource_id, payload::text, trace_id
FROM metaldocs.audit_events
WHERE ($1 = '' OR resource_type = $1)
  AND ($2 = '' OR resource_id = $2)
ORDER BY occurred_at DESC, id DESC
LIMIT $3
`
```
- Table touched: `metaldocs.audit_events` (`internal/modules/audit/infrastructure/postgres/writer.go:52`)
- Read verb: `SELECT` (`internal/modules/audit/infrastructure/postgres/writer.go:51`)
- Auth-area arg in SQL: none visible in this query text.

## §5 Response shape
- Success payload (`internal/modules/audit/delivery/http/handler.go:82-84`):
```json
{"items": [...]} 
```
  - exact code:
    - `httpresponse.WriteJSON(w, http.StatusOK, map[string]any{`
    - `	"items": responseItems,`
    - `})`
- Error envelope (`internal/modules/audit/delivery/http/handler.go:97-105`):
```json
{"error":{"code":"...","message":"...","details":{},"trace_id":"..."}}
```
  - exact code keys:
    - `"error": map[string]any{`
    - `"code":     code,`
    - `"message":  message,`
    - `"details":  map[string]any{},`
    - `"trace_id": traceID,`
- Status codes in handler behavior:
  - `405` method not allowed plain (no JSON body write): `internal/modules/audit/delivery/http/handler.go:39-41`
    - `if r.Method != http.MethodGet {`
    - `	w.WriteHeader(http.StatusMethodNotAllowed)`
    - `	return`
  - `400 VALIDATION_ERROR`: `internal/modules/audit/delivery/http/handler.go:48`
    - `writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value", requestTraceID(r))`
  - `500 INTERNAL_ERROR`: `internal/modules/audit/delivery/http/handler.go:60`
    - `writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit events", requestTraceID(r))`
- RFC 9457 / `application/problem+json`:
  - Handler uses `httpresponse.WriteJSON(... map[string]any{"error": ...})` at `internal/modules/audit/delivery/http/handler.go:98-105`.
  - `(unclear: content-type constant for this path not shown in inspected files)`

## §6 Cross-references
- Idempotency: none found in this flow (`handleEvents` ? `Service.ListEvents` ? `Writer.ListEvents`).
- Pagination: none (limit-only clamp, no cursor field).
- Audit-log emission: self (reads from audit events table via `internal/modules/audit/infrastructure/postgres/writer.go:52`).
- Auth middleware around route registration (verbatim findings):
  - Route registration: `apps/api/cmd/metaldocs-api/main.go:193` `	auditHandler.RegisterRoutes(mux)`
  - Middleware wiring: `apps/api/cmd/metaldocs-api/main.go:171-174`
    - `authMiddleware := authdelivery.NewMiddleware(authService, authCfg, authn.Enabled()).`
    - `		WithPublicPathChecker(newPublicPathChecker(permResolver))`
    - `iamMiddleware := iamdelivery.NewMiddleware(capabilityService, cachedProvider, authn.Enabled(), authCfg.LegacyHeaderEnabled).`
    - `		WithPermissionResolver(permResolver)`
  - Middleware wraps mux: `apps/api/cmd/metaldocs-api/main.go:386`
    - `handler := cors.Wrap(originProtection.Wrap(authMiddleware.Wrap(iamMiddleware.Wrap(httpObs.Wrap(rateLimiter.Wrap(mux))))))`
  - Permission resolver default and public checker behavior:
    - `apps/api/cmd/metaldocs-api/permissions.go:211` `		return "", false`
    - `apps/api/cmd/metaldocs-api/permissions.go:220-221`
      - `		_, guarded := resolver(method, path)`
      - `		return !guarded`
  - Grep evidence for explicit `/api/v1/audit/events` rule in resolver scope:
    - `rg -n "/api/v1/audit/events|audit/events" apps/api/cmd/metaldocs-api/permissions.go` ? no matches.
  - `(no authz middleware on this route — verified by grep)`