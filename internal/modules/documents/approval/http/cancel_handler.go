package approvalhttp

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
)

func (h *Handler) cancelInstance(ctx context.Context, db *sql.DB, req application.CancelInput) (application.CancelResult, error) {
	if h.cancelSvc == nil {
		return application.CancelResult{}, errors.New("cancel service not configured")
	}
	return h.cancelSvc.CancelInstance(ctx, db, req)
}

func (h *Handler) CancelHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	instanceID := r.PathValue("instance_id")

	idempKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}

	expectedRevisionVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.CancelRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	result, err := h.cancelInstance(r.Context(), h.db, application.CancelInput{
		TenantID:                tenantID,
		InstanceID:              instanceID,
		ExpectedRevisionVersion: expectedRevisionVersion,
		ActorUserID:             actorID,
		Reason:                  body.Reason,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.CancelResponse{
		DocumentID: result.DocumentID,
	})
}
