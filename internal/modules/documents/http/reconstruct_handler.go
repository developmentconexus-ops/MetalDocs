package documentshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/platform/tenant"
)

type ReconstructService interface {
	GetReconstruction(ctx context.Context, tenantID, actorID, docID string) (fanout.ReconstructionEntry, error)
}

type ReconstructHandler struct {
	svc ReconstructService
}

func NewReconstructHandler(svc ReconstructService) *ReconstructHandler {
	return &ReconstructHandler{svc: svc}
}

func (h *ReconstructHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/documents/{id}/reconstruct", h.HandleReconstruct)
}

func (h *ReconstructHandler) HandleReconstruct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeReconstructError(r.Context(), w, "", r.PathValue("id"), requestID(r), err)
		return
	}

	entry, err := h.svc.GetReconstruction(
		r.Context(),
		tenantID,
		actorID(r),
		r.PathValue("id"),
	)
	if err != nil {
		writeReconstructError(r.Context(), w, tenantID, r.PathValue("id"), requestID(r), err)
		return
	}

	writeFillInJSON(w, http.StatusOK, entry)
}

func writeReconstructError(ctx context.Context, w http.ResponseWriter, tenantID, documentID, reqID string, err error) {
	// Unified RFC 9457 writer (AD-2); mapFillInError classifies ErrCapDenied (403)
	// / ErrNotFound (404) and defaults to 500. Keep the log for the 500 case.
	if !errors.As(err, &authz.ErrCapDenied{}) && !errors.Is(err, v2dom.ErrNotFound) {
		slog.ErrorContext(ctx, "reconstruct document failed", "tenant_id", tenantID, "document_id", documentID, "err", err)
	}
	writeFillInError(w, reqID, err)
}
