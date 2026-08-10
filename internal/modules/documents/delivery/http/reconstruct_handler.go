package http

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

// ReconstructService is the application boundary ReconstructHandler consumes
// to fetch a document's render-fanout reconstruction entry.
type ReconstructService interface {
	GetReconstruction(ctx context.Context, tenantID, actorID, docID string) (fanout.ReconstructionEntry, error)
}

// ReconstructHandler serves the document reconstruction route.
type ReconstructHandler struct {
	svc ReconstructService
}

// NewReconstructHandler constructs a ReconstructHandler backed by the given service.
func NewReconstructHandler(svc ReconstructService) *ReconstructHandler {
	return &ReconstructHandler{svc: svc}
}

// HandleReconstruct returns the render-fanout reconstruction entry for a document.
func (h *ReconstructHandler) HandleReconstruct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeReconstructError(r.Context(), w, r, "", r.PathValue("id"), requestID(r), err)
		return
	}

	entry, err := h.svc.GetReconstruction(
		r.Context(),
		tenantID,
		actorID(r),
		r.PathValue("id"),
	)
	if err != nil {
		writeReconstructError(r.Context(), w, r, tenantID, r.PathValue("id"), requestID(r), err)
		return
	}

	writeFillInJSON(w, r, http.StatusOK, entry)
}

func writeReconstructError(ctx context.Context, w http.ResponseWriter, r *http.Request, tenantID, documentID, reqID string, err error) {
	// Unified RFC 9457 writer (AD-2); mapFillInError classifies ErrCapDenied (403)
	// / ErrNotFound (404) and defaults to 500. Keep the log for the 500 case.
	if !errors.As(err, &authz.ErrCapDenied{}) && !errors.Is(err, v2dom.ErrNotFound) {
		slog.ErrorContext(ctx, "reconstruct document failed", "tenant_id", tenantID, "document_id", documentID, "err", err)
	}
	writeFillInError(w, r, reqID, err)
}
