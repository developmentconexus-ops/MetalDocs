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

func (h *Handler) markObsolete(ctx context.Context, db *sql.DB, req application.MarkObsoleteRequest) (application.MarkObsoleteResult, error) {
	if h.obsoleteSvc == nil {
		return application.MarkObsoleteResult{}, errors.New("obsolete service not configured")
	}
	return h.obsoleteSvc.MarkObsolete(ctx, db, req)
}

func (h *Handler) ObsoleteHandler(w http.ResponseWriter, r *http.Request) {
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

	var body contracts.ObsoleteRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	_, err = h.markObsolete(r.Context(), h.db, application.MarkObsoleteRequest{
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
