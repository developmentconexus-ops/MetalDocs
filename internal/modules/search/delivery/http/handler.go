// Routes are mounted through the generated searchapi.ServerInterface router
// (HandlerWithOptions) rather than bare mux.HandleFunc — see Mount.
// Handler satisfies searchapi.ServerInterface directly (not the strict
// variant): handleSearchDocuments already writes typed JSON/problem+json
// bodies straight to http.ResponseWriter via the existing httpresponse
// helpers, which the strict (ctx, RequestObject) -> (ResponseObject, error)
// shape would force to be re-expressed as XxxResponseObject wrappers for no
// behavioral gain. Mirrors the security module's precedent and Task 6's auth
// migration.
package httpdelivery

import (
	"context"
	"metaldocs/internal/platform/httprouter"
	"net/http"
	"strconv"
	"strings"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	searchapi "metaldocs/internal/modules/search/api"
	searchdomain "metaldocs/internal/modules/search/domain"
	"metaldocs/internal/platform/apibase"
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

// searchDocumentsResponse is the typed envelope for the search-documents 200
// body, mirroring the OpenAPI SearchDocumentsResponse schema (required:
// [items]). Search is pre-codegen; per ADR 0012 hand-rolled typed structs are
// the sanctioned posture. It replaces the prior untyped-map response literal
// (M7 F7.3 typed-body parity) with byte-identical wire output: the same
// already-initialized slice is emitted under the same "items" key.
type searchDocumentsResponse struct {
	Items []SearchDocumentResponse `json:"items"`
}

func NewHandler(service Searcher) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Name() string { return "search" }
func (h *Handler) Tag() string  { return "search" }

func (h *Handler) Mount(mux httprouter.Muxer) {
	searchapi.HandlerWithOptions(h, searchapi.StdHTTPServerOptions{
		BaseURL:    apibase.BaseURL,
		BaseRouter: mux,
	})
}

// SearchDocuments adapts GET /search/documents to the existing handler. The
// generated SearchDocumentsParams struct is intentionally ignored: the
// existing handler parses the raw query itself (including subject,
// classification and tag, which the spec's typed params do not declare), and
// this migration must not change parsing behavior.
func (h *Handler) SearchDocuments(w http.ResponseWriter, r *http.Request, _ searchapi.SearchDocumentsParams) {
	h.handleSearchDocuments(w, r)
}

func (h *Handler) handleSearchDocuments(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(iamdomain.UserIDFromContext(r.Context())) == "" {
		httpresponse.WriteError(w, http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required")
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required")
		return
	}

	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid limit value")
			return
		}
		limit = n
	}

	expiryBefore, err := parseOptionalDateTimeQuery(r, "expiry_before")
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid expiry_before value")
		return
	}
	expiryAfter, err := parseOptionalDateTimeQuery(r, "expiry_after")
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid expiry_after value")
		return
	}

	items, err := h.service.SearchDocuments(r.Context(), searchdomain.Query{
		TenantID:        tenantID,
		Text:            r.URL.Query().Get("q"),
		DocumentType:    r.URL.Query().Get("document_type"),
		DocumentProfile: r.URL.Query().Get("document_profile"),
		DocumentFamily:  r.URL.Query().Get("document_family"),
		ProcessArea:     r.URL.Query().Get("process_area"),
		OwnerID:         r.URL.Query().Get("owner_id"),
		Department:      r.URL.Query().Get("department"),
		Status:          searchdomain.Status(r.URL.Query().Get("status")),
		ExpiryBefore:    expiryBefore,
		ExpiryAfter:     expiryAfter,
		Limit:           limit,
	})
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "Internal server error")
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

	httpresponse.WriteJSON(w, http.StatusOK, searchDocumentsResponse{Items: out})
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
