package documentshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	v2domain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
)

type ViewResult struct {
	PDFStatus string // "pending" | "ready" | "failed"
	SignedURL string // populated only when PDFStatus == "ready"
}

type ViewService interface {
	GetViewURL(ctx context.Context, tenantID, actorID, docID string) (ViewResult, error)
}

type ViewHandler struct {
	svc ViewService
}

func NewViewHandler(svc ViewService) *ViewHandler {
	return &ViewHandler{svc: svc}
}

func (h *ViewHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/documents/{id}/view", h.HandleView)
}

func (h *ViewHandler) HandleView(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantID(r)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}

	result, err := h.svc.GetViewURL(r.Context(),
		tenantID,
		actorID(r),
		r.PathValue("id"),
	)
	if err != nil {
		writeViewError(r.Context(), w, tenantID, r.PathValue("id"), err)
		return
	}
	payload := map[string]any{"pdf_status": result.PDFStatus}
	if result.PDFStatus == "ready" && result.SignedURL != "" {
		payload["signed_url"] = result.SignedURL
		payload["pdf_url"] = result.SignedURL
	}
	writeFillInJSON(w, http.StatusOK, payload)
}

func writeViewError(ctx context.Context, w http.ResponseWriter, tenantID, documentID string, err error) {
	switch {
	case errors.As(err, &authz.ErrCapDenied{}):
		writeFillInJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
	case errors.Is(err, v2domain.ErrNotFound):
		writeFillInJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
	default:
		slog.ErrorContext(ctx, "view document failed", "tenant_id", tenantID, "document_id", documentID, "err", err)
		writeFillInJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
	}
}
