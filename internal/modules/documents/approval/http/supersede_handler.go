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

func (h *Handler) publishSuperseding(ctx context.Context, db *sql.DB, req application.SupersedeRequest) (application.SupersedeResult, error) {
	if h.supersedeSvc == nil {
		return application.SupersedeResult{}, errors.New("supersede service not configured")
	}
	return h.supersedeSvc.PublishSuperseding(ctx, db, req)
}

func (h *Handler) SupersedeHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	documentID := r.PathValue("id")

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
	if h.db == nil {
		WriteError(w, errors.New("database not configured"))
		return
	}

	var body contracts.SupersedeRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	_, err = h.publishSuperseding(r.Context(), h.db, application.SupersedeRequest{
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
