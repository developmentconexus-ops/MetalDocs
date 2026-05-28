package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metaldocs/internal/modules/audit/application"
	"metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type Handler struct {
	service AuditQuerier
}

type AuditQuerier interface {
	ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error)
}

type EventResponse struct {
	ID           string         `json:"id"`
	OccurredAt   string         `json:"occurredAt"`
	ActorID      string         `json:"actorId"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resourceType"`
	ResourceID   string         `json:"resourceId"`
	Payload      map[string]any `json:"payload"`
	TraceID      string         `json:"traceId"`
}

func NewHandler(service AuditQuerier) *Handler {
	if service == nil {
		panic("audit: service is required")
	}
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/audit/events", h.handleEvents)
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := auditTenantFromRequest(w, r)
	if !ok {
		return
	}

	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			if writeErr := problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value")); writeErr != nil {
				slog.Warn("write problem response failed", "error", writeErr)
			}
			return
		}
		limit = parsed
	}

	items, err := h.service.ListEvents(r.Context(), domain.ListEventsQuery{
		ResourceType: strings.TrimSpace(r.URL.Query().Get("resourceType")),
		ResourceID:   strings.TrimSpace(r.URL.Query().Get("resourceId")),
		TenantID:     tenantID,
		Limit:        limit,
	})
	if err != nil {
		if errors.Is(err, application.ErrTenantRequired) {
			if writeErr := problem.Write(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Tenant claim required")); writeErr != nil {
				slog.Warn("write problem response failed", "error", writeErr)
			}
			return
		}
		if writeErr := problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit events")); writeErr != nil {
			slog.Warn("write problem response failed", "error", writeErr)
		}
		return
	}

	responseItems, err := buildEventResponses(items)
	if err != nil {
		slog.Error("audit payload decode failed", "error", err)
		if writeErr := problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Stored audit payload is invalid")); writeErr != nil {
			slog.Warn("write problem response failed", "error", writeErr)
		}
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"items": responseItems,
	})
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
		if writeErr := problem.Write(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required")); writeErr != nil {
			slog.Warn("write problem response failed", "error", writeErr)
		}
		return "", false
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		if writeErr := problem.Write(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Tenant claim required")); writeErr != nil {
			slog.Warn("write problem response failed", "error", writeErr)
		}
		return "", false
	}
	return tenantID, true
}
