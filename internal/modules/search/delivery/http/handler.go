package httpdelivery

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	iamdomain "metaldocs/internal/modules/iam/domain"
	searchdomain "metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/tenant"
)

type Handler struct {
	service Searcher
}

type Searcher interface {
	SearchDocuments(ctx context.Context, q searchdomain.Query) ([]searchdomain.Document, error)
}

type SearchDocumentResponse struct {
	DocumentID       string   `json:"documentId"`
	Title            string   `json:"title"`
	DocumentType     string   `json:"documentType"`
	DocumentProfile  string   `json:"documentProfile"`
	DocumentFamily   string   `json:"documentFamily"`
	DocumentSequence int      `json:"documentSequence"`
	DocumentCode     string   `json:"documentCode"`
	ProcessArea      string   `json:"processArea,omitempty"`
	Subject          string   `json:"subject,omitempty"`
	OwnerID          string   `json:"ownerId"`
	BusinessUnit     string   `json:"businessUnit,omitempty"`
	Department       string   `json:"department,omitempty"`
	Classification   string   `json:"classification,omitempty"`
	Status           string   `json:"status"`
	Tags             []string `json:"tags,omitempty"`
	EffectiveAt      string   `json:"effectiveAt,omitempty"`
	ExpiryAt         string   `json:"expiryAt,omitempty"`
	CreatedAt        string   `json:"createdAt"`
}

func NewHandler(service Searcher) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/search/documents", h.handleSearchDocuments)
}

func (h *Handler) handleSearchDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if strings.TrimSpace(iamdomain.UserIDFromContext(r.Context())) == "" {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", requestTraceID(r))
		return
	}

	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value", requestTraceID(r))
			return
		}
		limit = n
	}

	expiryBefore, err := parseOptionalDateTimeQuery(r, "expiryBefore")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid expiryBefore value", requestTraceID(r))
		return
	}
	expiryAfter, err := parseOptionalDateTimeQuery(r, "expiryAfter")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid expiryAfter value", requestTraceID(r))
		return
	}

	items, err := h.service.SearchDocuments(r.Context(), searchdomain.Query{
		TenantID:        tenantID,
		Text:            r.URL.Query().Get("q"),
		DocumentType:    r.URL.Query().Get("documentType"),
		DocumentProfile: r.URL.Query().Get("documentProfile"),
		DocumentFamily:  r.URL.Query().Get("documentFamily"),
		ProcessArea:     r.URL.Query().Get("processArea"),
		Subject:         r.URL.Query().Get("subject"),
		OwnerID:         r.URL.Query().Get("ownerId"),
		BusinessUnit:    r.URL.Query().Get("businessUnit"),
		Department:      r.URL.Query().Get("department"),
		Classification:  searchdomain.Classification(r.URL.Query().Get("classification")),
		Status:          searchdomain.Status(r.URL.Query().Get("status")),
		Tag:             r.URL.Query().Get("tag"),
		ExpiryBefore:    expiryBefore,
		ExpiryAfter:     expiryAfter,
		Limit:           limit,
	})
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", requestTraceID(r))
		return
	}

	out := make([]SearchDocumentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, SearchDocumentResponse{
			DocumentID:       item.ID,
			Title:            item.Title,
			DocumentType:     item.DocumentType,
			DocumentProfile:  item.DocumentProfile,
			DocumentFamily:   item.DocumentFamily,
			DocumentSequence: item.DocumentSequence,
			DocumentCode:     item.DocumentCode,
			ProcessArea:      item.ProcessArea,
			Subject:          item.Subject,
			OwnerID:          item.OwnerID,
			BusinessUnit:     item.BusinessUnit,
			Department:       item.Department,
			Classification:   string(item.Classification),
			Status:           string(item.Status),
			Tags:             append([]string(nil), item.Tags...),
			EffectiveAt:      formatOptionalTime(item.EffectiveAt),
			ExpiryAt:         formatOptionalTime(item.ExpiryAt),
			CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		})
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func requestTraceID(_ *http.Request) string {
	return uuid.NewString()
}

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	httpresponse.WriteJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
			TraceID: traceID,
		},
	})
}

func parseOptionalDateTimeQuery(r *http.Request, key string) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
