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

type Handler struct {
	service  AuditQuerier
	exporter AuditExporter
}

type EventResponse struct {
	ID           string         `json:"id"`
	OccurredAt   string         `json:"occurred_at"`
	ActorID      string         `json:"actor_id"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Payload      map[string]any `json:"payload"`
	TraceID      string         `json:"trace_id"`
}

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

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/audit/events", h.handleEvents)
	mux.HandleFunc("/api/v1/audit/events/export", h.handleExport)
	mux.HandleFunc("/api/v1/audit/events/export/", h.handleExportSubresource)
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v1/audit/events/export") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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
			writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Tenant claim required"))
			return
		}
		slog.Error("audit list events failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit events"))
		return
	}

	responseItems, err := buildEventResponses(items)
	if err != nil {
		slog.Error("audit payload decode failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Stored audit payload is invalid"))
		return
	}

	// Canonical cursor envelope (matches ListAuditEventsResponse / CursorPage and
	// every other list op): {items, page:{next_cursor, has_more}}. Previously this
	// emitted next_cursor/has_more at the top level — a spec↔runtime drift the FE
	// bridged with a dual-shape adapter (closed: ADR 0028-audit-events-cursor-shape).
	// has_more now comes from the reader's limit+1 probe, so an exact-multiple last
	// page no longer falsely advertises a next page (AIP-158).
	page := map[string]any{"next_cursor": nil, "has_more": false}
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		page["next_cursor"] = encodeCursor(domain.Cursor{OccurredAt: last.OccurredAt, ID: last.ID})
		page["has_more"] = true
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"items": responseItems,
		"page":  page,
	})
}

func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	tenantID, ok := auditTenantFromRequest(w, r)
	if !ok {
		return
	}
	if h.exporter == nil {
		writeProblem(w, problem.New(http.StatusNotImplemented, "NOT_IMPLEMENTED", "Audit export not configured"))
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
		writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body"))
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
		writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "format must be csv or jsonl"))
		return
	}

	job, err := h.exporter.ExportEvents(r.Context(), actorID, format, filter)
	switch {
	case err == nil:
	case errors.Is(err, application.ErrExportTooLarge):
		writeProblem(w, problem.New(http.StatusNotImplemented, "NOT_IMPLEMENTED", "Result set exceeds synchronous export limit; async export not yet available"))
		return
	case errors.Is(err, application.ErrInvalidFormat):
		writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "format must be csv or jsonl"))
		return
	case errors.Is(err, application.ErrTenantRequired):
		writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Tenant claim required"))
		return
	case errors.Is(err, application.ErrActorRequired):
		writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return
	case errors.Is(err, application.ErrExportRepoMissing), errors.Is(err, application.ErrCounterMissing):
		writeProblem(w, problem.New(http.StatusNotImplemented, "NOT_IMPLEMENTED", "Audit export not configured"))
		return
	default:
		slog.Error("audit export failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to export audit events"))
		return
	}

	httpresponse.WriteJSON(w, http.StatusAccepted, map[string]any{
		"export_id":  job.ID,
		"status":    string(job.Status),
		"signed_url": h.exporter.BuildSignedURL(job),
		"expires_at": job.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handler) handleExportSubresource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.exporter == nil {
		writeProblem(w, problem.New(http.StatusNotImplemented, "NOT_IMPLEMENTED", "Audit export not configured"))
		return
	}
	tail := strings.TrimPrefix(r.URL.Path, "/api/v1/audit/events/export/")
	if tail == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.SplitN(tail, "/", 2)
	exportID := parts[0]
	if len(parts) == 2 && parts[1] == "download" {
		h.handleExportDownload(w, r, exportID)
		return
	}
	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}

	tenantID, ok := auditTenantFromRequest(w, r)
	if !ok {
		return
	}
	actorID, ok := authn.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(actorID) == "" {
		writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return
	}
	job, err := h.exporter.GetExportStatus(r.Context(), tenantID, actorID, exportID)
	if err != nil {
		if errors.Is(err, domain.ErrExportJobNotFound) {
			writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "Export job not found"))
			return
		}
		slog.Error("audit export status failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to read export status"))
		return
	}
	resp := map[string]any{
		"export_id":  job.ID,
		"status":    string(job.Status),
		"signed_url": h.exporter.BuildSignedURL(job),
	}
	if !job.ExpiresAt.IsZero() {
		resp["expires_at"] = job.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if job.ErrorMessage != "" {
		resp["error"] = job.ErrorMessage
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleExportDownload(w http.ResponseWriter, r *http.Request, exportID string) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "token query parameter required"))
		return
	}
	job, err := h.exporter.LoadExportPayload(r.Context(), exportID, token)
	if err != nil {
		if errors.Is(err, domain.ErrExportJobNotFound) {
			writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "Export not found or token expired"))
			return
		}
		slog.Error("audit export download failed", "error", err)
		writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load export payload"))
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
			return domain.ListEventsQuery{}, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value")
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
		return time.Time{}, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", fmt.Sprintf("%s must be an RFC3339 timestamp", field))
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
		return domain.Cursor{}, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid cursor")
	}
	if sortValue == "" {
		return domain.Cursor{}, nil
	}
	ts, err := time.Parse(time.RFC3339Nano, sortValue)
	if err != nil {
		return domain.Cursor{}, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid cursor")
	}
	return domain.Cursor{OccurredAt: ts, ID: id}, nil
}

func buildEventResponses(items []domain.Event) ([]EventResponse, error) {
	responseItems := make([]EventResponse, 0, len(items))
	for _, item := range items {
		payload := map[string]any{}
		if raw := strings.TrimSpace(item.PayloadJSON); raw != "" {
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				return nil, fmt.Errorf("decode payload for event %s: %w", item.ID, err)
			}
		}
		responseItems = append(responseItems, EventResponse{
			ID:           item.ID,
			OccurredAt:   item.OccurredAt.UTC().Format(time.RFC3339),
			ActorID:      item.ActorID,
			Action:       item.Action,
			ResourceType: item.ResourceType,
			ResourceID:   item.ResourceID,
			Payload:      payload,
			TraceID:      item.TraceID,
		})
	}
	return responseItems, nil
}

func auditTenantFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	if _, ok := authn.UserIDFromContext(r.Context()); !ok {
		writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return "", false
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Tenant claim required"))
		return "", false
	}
	return tenantID, true
}

func writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("write problem response failed", "error", err)
	}
}
