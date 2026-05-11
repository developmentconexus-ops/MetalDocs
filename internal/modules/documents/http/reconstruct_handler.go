package documentshttp

import (
	"context"
	"errors"
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
	mux.HandleFunc("POST /api/v2/documents/{id}/reconstruct", h.HandleReconstruct)
}

func (h *ReconstructHandler) HandleReconstruct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeReconstructError(w, err)
		return
	}

	entry, err := h.svc.GetReconstruction(
		r.Context(),
		tenantID,
		actorID(r),
		r.PathValue("id"),
	)
	if err != nil {
		writeReconstructError(w, err)
		return
	}

	writeFillInJSON(w, http.StatusOK, entry)
}

func writeReconstructError(w http.ResponseWriter, err error) {
	switch {
	case errors.As(err, &authz.ErrCapDenied{}):
		writeFillInJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
	case errors.Is(err, v2dom.ErrNotFound):
		writeFillInJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
	default:
		writeFillInJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
	}
}
