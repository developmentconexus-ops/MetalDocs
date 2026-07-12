package approvalhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/strictjson"
)

var markReviewed = func(h *Handler, ctx context.Context, runner db.TxRunner, req application.MarkReviewedRequest) (application.MarkReviewedResult, error) {
	if h.services == nil || h.services.MarkReviewed == nil {
		return application.MarkReviewedResult{}, errors.New("mark-reviewed service not configured")
	}
	return h.services.MarkReviewed.MarkReviewed(ctx, runner, req)
}

// MarkReviewedHandler records a periodic-review completion on a live
// published revision (F6.2). Requires a valid If-Match header (OCC
// precondition) and gates the write via document.review (tenant-grade,
// authz.Require inside the service tx) — not a status transition, the
// document stays "published".
func (h *Handler) MarkReviewedHandler(w http.ResponseWriter, r *http.Request) {
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

	var body contracts.MarkReviewedRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	reviewDueAt, err := body.ParsedReviewDueAt()
	if err != nil {
		WriteError(w, err)
		return
	}

	var effectiveTo *time.Time
	if body.EffectiveTo != nil {
		t := body.EffectiveTo.UTC()
		effectiveTo = &t
	}

	result, err := markReviewed(h, r.Context(), h.runner, application.MarkReviewedRequest{
		TenantID:                tenantID,
		DocumentID:              documentID,
		ReviewDueAt:             reviewDueAt,
		EffectiveTo:             effectiveTo,
		ExpectedRevisionVersion: expectedRevisionVersion,
		ReviewedBy:              actorID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("ETag", instanceETag(result.RevisionVersion))
	WriteJSON(w, http.StatusOK, contracts.MarkReviewedResponse{
		DocumentID: result.DocumentID,
		NewStatus:  string(docsdomain.DocStatusPublished),
	})
}
