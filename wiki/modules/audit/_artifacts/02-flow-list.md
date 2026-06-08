# Phase 2 Data-flow Trace — GET /api/v1/audit/events

## 1 Entry point / OpenAPI op
- OpenAPI path present: `api/openapi/v1/openapi.yaml:819` → `  /audit/events:`
- OpenAPI method present: `api/openapi/v1/openapi.yaml:820` → `    get:`
- OpenAPI operationId: `(unclear: no operationId field under /audit/events get in api/openapi/v1/openapi.yaml:819)`
- Handler registration: `internal/modules/audit/delivery/http/handler.go:67-68`
  - `func (h *Handler) RegisterRoutes(mux *http.ServeMux) {`
  - `	mux.HandleFunc("/api/v1/audit/events", h.handleEvents)`
- Handler entry: `internal/modules/audit/delivery/http/handler.go:73` `func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {`

## 2 Call chain
1. `internal/modules/audit/delivery/http/handler.go:73` `func (h *Handler) handleEvents...`
   - calls `internal/modules/audit/application/service.go:18` via `internal/modules/audit/delivery/http/handler.go:93`
   - exact call site:
     - `items, err := h.service.ListEvents(r.Context(), query)`
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
     - `internal/modules/audit/infrastructure/postgres/writer.go:59` `rows, err := w.db.QueryContext(ctx, q, ...)`

## 3 State changes
- none

## 4 SQL touched
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

## 5 Response shape
- Success payload (`internal/modules/audit/delivery/http/handler.go:122-125`):
```json
{"items": [...], "page": {"next_cursor": "...", "has_more": true}}
```
  - exact code:
    - `httpresponse.WriteJSON(w, http.StatusOK, map[string]any{`
    - `	"items": responseItems,`
    - `	"page":  page,`
    - `})`
  - Note: cursor envelope shape (`{items, page:{next_cursor, has_more}}`) matches every other list op (closed: ADR 2026-06-03-audit-events-cursor-shape). Previously emitted flat `next_cursor`/`has_more`.
- Error envelope: RFC 9457 `application/problem+json` via `writeProblem` / `problem.Write` at `internal/modules/audit/delivery/http/handler.go:427-430`.
- Status codes in handler behavior:
  - `405` method not allowed plain (no JSON body write): `internal/modules/audit/delivery/http/handler.go:78-81`
    - `if r.Method != http.MethodGet {`
    - `	w.WriteHeader(http.StatusMethodNotAllowed)`
    - `	return`
  - `400 VALIDATION_ERROR`: `internal/modules/audit/delivery/http/handler.go:88-90`
    - `writeProblem(w, perr)` (from `parseListQuery`)
  - `500 INTERNAL_ERROR`: `internal/modules/audit/delivery/http/handler.go:99-101`
    - `writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", ...))`

## 6 Cross-references
- Idempotency: none found in this flow (`handleEvents` → `Service.ListEvents` → `Writer.ListEvents`).
- Pagination: opaque keyset cursor (`occurred_at|id` base64) + `limit`; response includes `page.next_cursor`/`page.has_more`.
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
  - Permission resolver — T-001 CLOSED: explicit `CapAuditRead` rows now present at `apps/api/cmd/metaldocs-api/permissions.go:229-231` (GET /audit/events, POST /audit/events/export, GET /audit/events/export/*). The old resolver-default fall-through (":211", ":220-221") no longer applies to this route.
