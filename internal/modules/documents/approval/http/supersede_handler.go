package approvalhttp

import (
	"context"
	"errors"
	"net/http"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/strictjson"
)

func (h *Handler) publishSuperseding(ctx context.Context, runner db.TxRunner, req application.SupersedeRequest) (application.SupersedeResult, error) {
	if h.supersedeSvc == nil {
		return application.SupersedeResult{}, errors.New("supersede service not configured")
	}
	return h.supersedeSvc.PublishSuperseding(ctx, runner, req)
}

func (h *Handler) SupersedeHandler(w http.ResponseWriter, r *http.Request) {
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
	var body contracts.SupersedeRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	_, err = h.publishSuperseding(r.Context(), h.runner, application.SupersedeRequest{
		TenantID:           tenantID,
		NewDocumentID:      documentID,
		PriorDocumentID:    body.SupersededDocumentID,
		SupersededBy:       actorID,
		NewRevisionVersion: expectedRevisionVersion,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.SupersedeResponse{
		DocumentID:   documentID,
		SupersededID: body.SupersededDocumentID,
	})
}
