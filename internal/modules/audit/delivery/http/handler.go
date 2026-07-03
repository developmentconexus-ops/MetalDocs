// Package httpdelivery adapts the audit module's application service onto
// HTTP: listing audit events and driving the export/status/download flow
// through the generated auditapi.ServerInterface router.
package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	auditapi "metaldocs/internal/modules/audit/api"
	"metaldocs/internal/modules/audit/application"
	"metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// AuditQuerier is the minimum the list handler needs. ListEvents returns the
// page plus hasMore so the handler emits next_cursor only on a genuine further
// page (limit+1 probe — see domain.Reader).
type AuditQuerier interface {
	ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, bool, error)
}

// AuditExporter captures the service surface for export/status/download.
type AuditExporter interface {
	ExportEvents(ctx context.Context, actorID string, format domain.ExportFormat, filter domain.ListEventsQuery) (domain.ExportJob, error)
	GetExportStatus(ctx context.Context, tenantID, actorID, exportID string) (domain.ExportJob, error)
	LoadExportPayload(ctx context.Context, exportID, token string) (domain.ExportJob, error)
	BuildSignedURL(job domain.ExportJob) string
}

// Handler implements auditapi.ServerInterface for the audit module's mounted
// routes. exporter is optional: when nil, export/status/download routes
// respond 501 NOT_IMPLEMENTED (see WithExporter).
type Handler struct {
	service  AuditQuerier
	exporter AuditExporter
}

// NewHandler constructs a Handler backed by service. Panics if service is nil.
func NewHandler(service AuditQuerier) *Handler {
	if service == nil {
		panic("audit: service is required")
	}
	return &Handler{service: service}
}

// WithExporter enables the export endpoints. When nil, export routes respond
// with 501 NOT_IMPLEMENTED.
func (h *Handler) WithExporter(exporter AuditExporter) *Handler {
	h.exporter = exporter
	return h
}

// RegisterRoutes mounts the audit module's routes onto mux via the generated
// auditapi.ServerInterface router (HandlerWithOptions), replacing the prior
// hand-written mux.HandleFunc registrations (T-008, residual of CON-09).
// Handler satisfies auditapi.ServerInterface directly (see the adapter
// methods below) — each one delegates straight through to the existing,
// already-tested private handler methods (handleEvents / handleExport /
// handleExportSubresource) unchanged. Mirrors the controlleddocuments
// pattern (internal/modules/controlleddocuments/delivery/http/handler.go)
// and the templates/approval CON-03 migrations
// (internal/modules/templates/delivery/http/handler.go,
// internal/modules/documents/approval/http/router.go).
//
// Audit has no per-route Idempotency-Key requirement (list is a GET, export
// creates a job but the spec does not mark it idempotent), so no Middlewares
// closure is needed here — unlike templates/approval. Tier-1 capability
// gating (CapAuditRead) lives in apps/api/cmd/metaldocs-api/permissions.go
// and keys off method + path, which the generated router preserves
// byte-for-byte (AD-1: BaseURL "/api/v1" + spec's relative paths), so it is
// unaffected by this mount-mechanism swap.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	auditapi.HandlerWithOptions(h, auditapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    "/api/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, err.Error()))
		},
	})
}

// ─── auditapi.ServerInterface adapter ──────────────────────────────────────
//
// Thin delegation only: each method below hands off to the pre-existing
// private handler (unchanged) rather than reimplementing parsing/response
// logic. The generated wrapper already validates/binds typed params before
// calling these, but the delegated-to methods keep parsing directly from `r`
// themselves (same as the iam Router precedent, e.g. RevokeSession) — the Go
// 1.22 mux guarantees r.URL.Path and the wrapper's bound params agree for the
// same request, so there is no behavior drift, only a redundant (harmless)
// re-parse.

// ListAuditEvents adapts GET /audit/events to the existing list handler.
func (h *Handler) ListAuditEvents(w http.ResponseWriter, r *http.Request, _ auditapi.ListAuditEventsParams) {
	h.handleEvents(w, r)
}

// ExportAuditEvents adapts POST /audit/events/export to the existing export handler.
func (h *Handler) ExportAuditEvents(w http.ResponseWriter, r *http.Request) {
	h.handleExport(w, r)
}

// GetAuditExportStatus adapts GET /audit/events/export/{export_id} to the
// existing subresource dispatcher, which already distinguishes status from
// download by inspecting the request path tail.
func (h *Handler) GetAuditExportStatus(w http.ResponseWriter, r *http.Request, _ string) {
	h.handleExportSubresource(w, r)
}

// DownloadAuditExport adapts GET /audit/events/export/{export_id}/download to
// the same existing subresource dispatcher as GetAuditExportStatus — it
// already routes to handleExportDownload once it sees the "/download" tail.
func (h *Handler) DownloadAuditExport(w http.ResponseWriter, r *http.Request, _ string, _ auditapi.DownloadAuditExportParams) {
	h.handleExportSubresource(w, r)
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v1/audit/events/export") {
		writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Not found"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeProblem(w, problem.New(http.StatusMethodNotAllowed, problem.CodeMethodNotAllowed, "Method not allowed"))
		return
	}
	tenantID, ok := auditTenantFromRequest(w, r)
	if !ok {
		return
	}

	query, perr := parseListQuery(r, tenantID)
	if perr != nil {
		writeProblem(w, perr)
		return
	}

	items, hasMore, err := h.service.ListEvents(r.Context(), query)
	if err != nil {
		if errors.Is(err, application.ErrTenantRequired) {
			writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthForbidden, "Tenant claim required"))
			return
		}
		slog.Error("audit list events failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to list audit events"))
		return
	}

	responseItems, err := buildEventResponses(items)
	if err != nil {
		slog.Error("audit payload decode failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Stored audit payload is invalid"))
		return
	}

	// Canonical cursor envelope (matches ListAuditEventsResponse / CursorPage and
	// every other list op): {items, page:{next_cursor, has_more}}. Previously this
	// emitted next_cursor/has_more at the top level — a spec↔runtime drift the FE
	// bridged with a dual-shape adapter (closed: ADR 0028-audit-events-cursor-shape).
	// has_more now comes from the reader's limit+1 probe, so an exact-multiple last
	// page no longer falsely advertises a next page (AIP-158).
	page := auditapi.CursorPage{NextCursor: nil, HasMore: false}
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		cursor := encodeCursor(domain.Cursor{OccurredAt: last.OccurredAt, ID: last.ID})
		page.NextCursor = &cursor
		page.HasMore = true
	}

	httpresponse.WriteJSON(w, http.StatusOK, auditapi.ListAuditEventsResponse{
		Items: responseItems,
		Page:  page,
	})
}

func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeProblem(w, problem.New(http.StatusMethodNotAllowed, problem.CodeMethodNotAllowed, "Method not allowed"))
		return
	}
	tenantID, ok := auditTenantFromRequest(w, r)
	if !ok {
		return
	}
	if h.exporter == nil {
		writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeNotImplemented, "Audit export not configured"))
		return
	}
	actorID, _ := authn.UserIDFromContext(r.Context())

	var body struct {
		Format string `json:"format"`
		Filter struct {
			ActorID        string `json:"actor_id"`
			Action         string `json:"action"`
			ResourceType   string `json:"resource_type"`
			ResourceID     string `json:"resource_id"`
			OccurredAfter  string `json:"occurred_after"`
			OccurredBefore string `json:"occurred_before"`
			Q              string `json:"q"`
		} `json:"filter"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "Invalid JSON body"))
		return
	}

	filter := domain.ListEventsQuery{
		TenantID:     tenantID,
		ActorID:      body.Filter.ActorID,
		Action:       body.Filter.Action,
		ResourceType: body.Filter.ResourceType,
		ResourceID:   body.Filter.ResourceID,
		Query:        body.Filter.Q,
	}
	if ts, perr := parseTime("occurredAfter", body.Filter.OccurredAfter); perr != nil {
		writeProblem(w, perr)
		return
	} else {
		filter.OccurredAfter = ts
	}
	if ts, perr := parseTime("occurredBefore", body.Filter.OccurredBefore); perr != nil {
		writeProblem(w, perr)
		return
	} else {
		filter.OccurredBefore = ts
	}

	format := domain.ExportFormat(strings.ToLower(strings.TrimSpace(body.Format)))
	if !format.Valid() {
		writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "format must be csv or jsonl"))
		return
	}

	job, err := h.exporter.ExportEvents(r.Context(), actorID, format, filter)
	switch {
	case err == nil:
	case errors.Is(err, application.ErrExportTooLarge):
		writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeNotImplemented, "Result set exceeds synchronous export limit; async export not yet available"))
		return
	case errors.Is(err, application.ErrInvalidFormat):
		writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "format must be csv or jsonl"))
		return
	case errors.Is(err, application.ErrTenantRequired):
		writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthForbidden, "Tenant claim required"))
		return
	case errors.Is(err, application.ErrActorRequired):
		writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
		return
	case errors.Is(err, application.ErrExportRepoMissing), errors.Is(err, application.ErrCounterMissing):
		writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeNotImplemented, "Audit export not configured"))
		return
	default:
		slog.Error("audit export failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to export audit events"))
		return
	}

	httpresponse.WriteJSON(w, http.StatusAccepted, auditapi.AuditExportResponse{
		ExportId:  job.ID,
		Status:    string(job.Status),
		SignedUrl: h.exporter.BuildSignedURL(job),
		ExpiresAt: job.ExpiresAt.UTC().Truncate(time.Second),
	})
}

func (h *Handler) handleExportSubresource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeProblem(w, problem.New(http.StatusMethodNotAllowed, problem.CodeMethodNotAllowed, "Method not allowed"))
		return
	}
	if h.exporter == nil {
		writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeNotImplemented, "Audit export not configured"))
		return
	}
	tail := strings.TrimPrefix(r.URL.Path, "/api/v1/audit/events/export/")
	if tail == "" {
		writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Not found"))
		return
	}
	parts := strings.SplitN(tail, "/", 2)
	exportID := parts[0]
	if len(parts) == 2 && parts[1] == "download" {
		h.handleExportDownload(w, r, exportID)
		return
	}
	if len(parts) != 1 {
		writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Not found"))
		return
	}

	tenantID, ok := auditTenantFromRequest(w, r)
	if !ok {
		return
	}
	actorID, ok := authn.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(actorID) == "" {
		writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
		return
	}
	job, err := h.exporter.GetExportStatus(r.Context(), tenantID, actorID, exportID)
	if err != nil {
		if errors.Is(err, domain.ErrExportJobNotFound) {
			writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Export job not found"))
			return
		}
		slog.Error("audit export status failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to read export status"))
		return
	}
	resp := auditapi.AuditExportStatusResponse{
		ExportId:  job.ID,
		Status:    string(job.Status),
		SignedUrl: h.exporter.BuildSignedURL(job),
	}
	if !job.ExpiresAt.IsZero() {
		expiresAt := job.ExpiresAt.UTC().Truncate(time.Second)
		resp.ExpiresAt = &expiresAt
	}
	if job.ErrorMessage != "" {
		msg := job.ErrorMessage
		resp.Error = &msg
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleExportDownload(w http.ResponseWriter, r *http.Request, exportID string) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "token query parameter required"))
		return
	}
	job, err := h.exporter.LoadExportPayload(r.Context(), exportID, token)
	if err != nil {
		if errors.Is(err, domain.ErrExportJobNotFound) {
			writeProblem(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Export not found or token expired"))
			return
		}
		slog.Error("audit export download failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to load export payload"))
		return
	}
	contentType := "application/octet-stream"
	filename := exportID
	switch job.Format {
	case domain.ExportFormatCSV:
		contentType = "text/csv; charset=utf-8"
		filename = "audit-export-" + exportID + ".csv"
	case domain.ExportFormatJSONL:
		contentType = "application/x-ndjson"
		filename = "audit-export-" + exportID + ".jsonl"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if _, werr := w.Write(job.Payload); werr != nil {
		slog.Warn("audit export write failed", "error", werr)
	}
}

func parseListQuery(r *http.Request, tenantID string) (domain.ListEventsQuery, *problem.Problem) {
	q := r.URL.Query()
	limit := 50
	if raw := strings.TrimSpace(q.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return domain.ListEventsQuery{}, problem.New(http.StatusBadRequest, problem.CodeValidationError, "Invalid limit value")
		}
		limit = parsed
	}
	// Clamp the upper bound to the design-system max (100), matching documents /
	// controlled-documents. The unset default (50) is below MaxLimit so it passes
	// through unchanged.
	limit = pagination.ClampLimit(limit)

	occurredAfter, perr := parseTime("occurred_after", q.Get("occurred_after"))
	if perr != nil {
		return domain.ListEventsQuery{}, perr
	}
	occurredBefore, perr := parseTime("occurred_before", q.Get("occurred_before"))
	if perr != nil {
		return domain.ListEventsQuery{}, perr
	}

	cursor, perr := decodeCursor(q.Get("cursor"))
	if perr != nil {
		return domain.ListEventsQuery{}, perr
	}

	return domain.ListEventsQuery{
		TenantID:       tenantID,
		ResourceType:   strings.TrimSpace(q.Get("resource_type")),
		ResourceID:     strings.TrimSpace(q.Get("resource_id")),
		ActorID:        strings.TrimSpace(q.Get("actor_id")),
		Action:         strings.TrimSpace(q.Get("action")),
		Query:          strings.TrimSpace(q.Get("q")),
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		Cursor:         cursor,
		Limit:          limit,
	}, nil
}

func parseTime(field, raw string) (time.Time, *problem.Problem) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, problem.New(http.StatusBadRequest, problem.CodeValidationError, fmt.Sprintf("%s must be an RFC3339 timestamp", field))
	}
	return t.UTC(), nil
}

// encodeCursor maps the audit (occurred_at, id) anchor onto the shared opaque
// keyset codec (pagination.EncodeCursor), so all three list endpoints share one
// base64 dialect and can no longer diverge. The sort value is the RFC3339Nano
// occurred_at; a zero cursor encodes to "".
func encodeCursor(c domain.Cursor) string {
	if c.IsZero() {
		return ""
	}
	return pagination.EncodeCursor(c.OccurredAt.UTC().Format(time.RFC3339Nano), c.ID)
}

// decodeCursor reverses encodeCursor via the shared codec, re-parsing the sort
// value back into occurred_at. A blank cursor is the first page; a malformed one
// (bad base64 / shape, or an unparseable timestamp) is a VALIDATION_ERROR.
func decodeCursor(raw string) (domain.Cursor, *problem.Problem) {
	sortValue, id, err := pagination.DecodeCursor(raw)
	if err != nil {
		return domain.Cursor{}, problem.New(http.StatusBadRequest, problem.CodeValidationError, "Invalid cursor")
	}
	if sortValue == "" {
		return domain.Cursor{}, nil
	}
	ts, err := time.Parse(time.RFC3339Nano, sortValue)
	if err != nil {
		return domain.Cursor{}, problem.New(http.StatusBadRequest, problem.CodeValidationError, "Invalid cursor")
	}
	return domain.Cursor{OccurredAt: ts, ID: id}, nil
}

// buildEventResponses maps stored events onto the generated auditapi.AuditEventItem.
// OccurredAt is truncated to the second (UTC) so the marshaled time.Time matches the
// historical RFC3339 second-precision wire output byte-for-byte. The payload decode
// buffer stays an untyped map — it feeds the allowlisted AuditEventItem.Payload
// domain-mirror field (arbitrary stored JSON), not a response literal.
func buildEventResponses(items []domain.Event) ([]auditapi.AuditEventItem, error) {
	responseItems := make([]auditapi.AuditEventItem, 0, len(items))
	for _, item := range items {
		payload := map[string]any{}
		if raw := strings.TrimSpace(item.PayloadJSON); raw != "" {
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				return nil, fmt.Errorf("decode payload for event %s: %w", item.ID, err)
			}
		}
		responseItems = append(responseItems, auditapi.AuditEventItem{
			Id:           item.ID,
			OccurredAt:   item.OccurredAt.UTC().Truncate(time.Second),
			ActorId:      item.ActorID,
			Action:       item.Action,
			ResourceType: item.ResourceType,
			ResourceId:   item.ResourceID,
			Payload:      payload,
			TraceId:      item.TraceID,
		})
	}
	return responseItems, nil
}

func auditTenantFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	if _, ok := authn.UserIDFromContext(r.Context()); !ok {
		writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required"))
		return "", false
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeProblem(w, problem.New(http.StatusForbidden, problem.CodeAuthForbidden, "Tenant claim required"))
		return "", false
	}
	return tenantID, true
}

func writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("write problem response failed", "error", err)
	}
}
