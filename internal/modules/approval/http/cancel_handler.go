package approvalhttp

import (
	"context"
	"errors"
	"net/http"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/strictjson"
)

func (h *Handler) cancelInstance(ctx context.Context, runner db.TxRunner, req application.CancelInput) (application.CancelResult, error) {
	if h.cancelSvc == nil {
		return application.CancelResult{}, errors.New("cancel service not configured")
	}
	return h.cancelSvc.CancelInstance(ctx, runner, req)
}

// CancelHandler cancels an in-progress approval instance. Requires a valid
// If-Match header (OCC precondition against the document's revision_version).
func (h *Handler) CancelHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	actorID, err := actorIDFromRequest(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	instanceID := r.PathValue("instance_id")

	expectedRevisionVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var body contracts.CancelRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, r, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, r, NewValidationError(err.Error()))
		return
	}

	result, err := h.cancelInstance(r.Context(), h.runner, application.CancelInput{
		TenantID:                tenantID,
		InstanceID:              instanceID,
		ExpectedRevisionVersion: expectedRevisionVersion,
		ActorUserID:             actorID,
		Reason:                  body.Reason,
	})
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, r, http.StatusOK, contracts.CancelResponse{
		DocumentID: result.DocumentID,
	})
}
