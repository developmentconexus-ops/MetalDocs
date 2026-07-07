package approvalhttp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/strictjson"
)

// SubmitHandler submits a document revision for approval, opening a new
// approval instance. Requires both an Idempotency-Key header (persisted as
// approval_instances.idempotency_key, F-D4) and a valid If-Match header (OCC precondition).
func (h *Handler) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	documentID := r.PathValue("id")
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := iamdomain.UserIDFromContext(r.Context())

	// The platform idempotency middleware (router.go) already validated presence
	// and UUID shape; re-read the header to thread the client key into the service
	// so it becomes approval_instances.idempotency_key (DB UNIQUE backstop, F-D4).
	// The empty guard mirrors the route-admin sibling and fail-closes if this
	// handler is ever mounted without the middleware (e.g. in a unit test).
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}

	expectedRevisionVersion, err := parseSubmitIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var req contracts.SubmitRequest
	if err := strictjson.Decode(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	submitSvc := h.submitSvc
	if submitSvc == nil && h.services != nil {
		submitSvc = h.services.Submit
	}
	if submitSvc == nil {
		WriteError(w, errors.New("submit service not configured"))
		return
	}

	var reasonForChange, reasonCategory string
	if req.ReasonForChange != nil {
		reasonForChange = *req.ReasonForChange
	}
	if req.ReasonCategory != nil {
		reasonCategory = *req.ReasonCategory
	}

	result, err := submitSvc.SubmitRevisionForReview(r.Context(), h.runner, application.SubmitRequest{
		TenantID:        tenantID,
		DocumentID:      documentID,
		RouteID:         req.RouteID,
		SubmittedBy:     actorID,
		RevisionTitle:   req.RevisionTitle,
		ReasonForChange: reasonForChange,
		ReasonCategory:  reasonCategory,
		ContentFormData: map[string]any{"_content_hash": req.ContentHash},
		RevisionVersion: expectedRevisionVersion,
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	newETag := "\"v" + strconv.Itoa(expectedRevisionVersion+1) + "\""
	w.Header().Set("ETag", newETag)
	WriteJSON(w, http.StatusCreated, contracts.SubmitResponse{
		InstanceID: result.InstanceID,
		WasReplay:  false,
		ETag:       newETag,
	})
}
