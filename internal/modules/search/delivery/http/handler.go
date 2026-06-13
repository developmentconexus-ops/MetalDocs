package httpdelivery

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	searchdomain "metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type Handler struct {
	service Searcher
}

type Searcher interface {
	SearchDocuments(ctx context.Context, q searchdomain.Query) ([]searchdomain.Document, error)
}

// SearchDocumentResponse mirrors the SearchDocumentItem schema in
// api/openapi/v1/openapi.yaml. The legacy subject/business_unit/
// classification/tags fields were removed in Wave 1.5 (F-13a): the SQL
// reader never populated them, so they were always zero-valued on the wire
// and absent from the spec.
type SearchDocumentResponse struct {
	DocumentID       string `json:"document_id"`
	Title            string `json:"title"`
	DocumentType     string `json:"document_type"`
	DocumentProfile  string `json:"document_profile"`
	DocumentFamily   string `json:"document_family"`
	DocumentSequence int    `json:"document_sequence"`
	DocumentCode     string `json:"document_code"`
	ProcessArea      string `json:"process_area,omitempty"`
	OwnerID          string `json:"owner_id"`
	Department       string `json:"department,omitempty"`
	Status           string `json:"status"`
	EffectiveAt      string `json:"effective_at,omitempty"`
	ExpiryAt         string `json:"expiry_at,omitempty"`
	CreatedAt        string `json:"created_at"`
}

func NewHandler(service Searcher) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/search/documents", h.handleSearchDocuments)
}

func (h *Handler) handleSearchDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpresponse.WriteMethodNotAllowed(w, http.MethodGet)
		return
	}

	if strings.TrimSpace(iamdomain.UserIDFromContext(r.Context())) == "" {
		httpresponse.WriteError(w, http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required")
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusUnauthorized, problem.CodeAuthUnauthorized, "Authentication required")
		return
	}

	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "Invalid limit value")
			return
		}
		limit = n
	}

	expiryBefore, err := parseOptionalDateTimeQuery(r, "expiry_before")
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "Invalid expiry_before value")
		return
	}
	expiryAfter, err := parseOptionalDateTimeQuery(r, "expiry_after")
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "Invalid expiry_after value")
		return
	}

	items, err := h.service.SearchDocuments(r.Context(), searchdomain.Query{
		TenantID:        tenantID,
		Text:            r.URL.Query().Get("q"),
		DocumentType:    r.URL.Query().Get("document_type"),
		DocumentProfile: r.URL.Query().Get("document_profile"),
		DocumentFamily:  r.URL.Query().Get("document_family"),
		ProcessArea:     r.URL.Query().Get("process_area"),
		Subject:         r.URL.Query().Get("subject"),
		OwnerID:         r.URL.Query().Get("owner_id"),
		Department:      r.URL.Query().Get("department"),
		Classification:  searchdomain.Classification(r.URL.Query().Get("classification")),
		Status:          searchdomain.Status(r.URL.Query().Get("status")),
		Tag:             r.URL.Query().Get("tag"),
		ExpiryBefore:    expiryBefore,
		ExpiryAfter:     expiryAfter,
		Limit:           limit,
	})
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")
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
			OwnerID:          item.OwnerID,
			Department:       item.Department,
			Status:           string(item.Status),
			EffectiveAt:      formatOptionalTime(item.EffectiveAt),
			ExpiryAt:         formatOptionalTime(item.ExpiryAt),
			CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		})
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"items": out})
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
