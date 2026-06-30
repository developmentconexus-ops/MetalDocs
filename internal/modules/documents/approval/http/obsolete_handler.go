package approvalhttp

import (
	"context"
	"errors"
	"net/http"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/platform/db"
)

func (h *Handler) markObsolete(ctx context.Context, runner db.TxRunner, req application.MarkObsoleteRequest) (application.MarkObsoleteResult, error) {
	if h.obsoleteSvc == nil {
		return application.MarkObsoleteResult{}, errors.New("obsolete service not configured")
	}
	return h.obsoleteSvc.MarkObsolete(ctx, runner, req)
}

func (h *Handler) ObsoleteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	documentID := r.PathValue("id")

	expectedRevisionVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.ObsoleteRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	_, err = h.markObsolete(r.Context(), h.runner, application.MarkObsoleteRequest{
		TenantID:        tenantID,
		DocumentID:      documentID,
		MarkedBy:        actorID,
		RevisionVersion: expectedRevisionVersion,
		Reason:          body.Reason,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.ObsoleteResponse{
		DocumentID: documentID,
	})
}
