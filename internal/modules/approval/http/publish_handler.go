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

var (
	publishApproved = func(h *Handler, ctx context.Context, runner db.TxRunner, req application.PublishRequest) (application.PublishResult, error) {
		if h.services == nil || h.services.Publish == nil {
			return application.PublishResult{}, errors.New("publish service not configured")
		}
		return h.services.Publish.PublishApproved(ctx, runner, req)
	}
	schedulePublish = func(h *Handler, ctx context.Context, runner db.TxRunner, req application.SchedulePublishRequest) (application.SchedulePublishResult, error) {
		if h.services == nil || h.services.Publish == nil {
			return application.SchedulePublishResult{}, errors.New("publish service not configured")
		}
		return h.services.Publish.SchedulePublish(ctx, runner, req)
	}
)

// PublishHandler publishes a document whose active approval instance reached
// InstanceApproved. Requires a valid If-Match header (OCC precondition).
func (h *Handler) PublishHandler(w http.ResponseWriter, r *http.Request) {
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
	if h.readSvc == nil {
		WriteError(w, errors.New("read service not configured"))
		return
	}

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.runner, tenantID, documentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	result, err := publishApproved(h, r.Context(), h.runner, application.PublishRequest{
		TenantID:                tenantID,
		InstanceID:              inst.ID,
		PublishedBy:             actorID,
		ExpectedRevisionVersion: expectedRevisionVersion,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.PublishResponse{
		DocumentID: result.DocumentID,
		NewStatus:  string(docsdomain.DocStatusPublished),
	})
}

// SchedulePublishHandler schedules an approved document for publish at a
// future effective date. Requires a valid If-Match header (OCC precondition).
func (h *Handler) SchedulePublishHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	documentID := r.PathValue("id")

	ifMatchVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.runner, tenantID, documentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.SchedulePublishRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	effectiveFrom, err := time.Parse(time.RFC3339, body.EffectiveFrom)
	if err != nil {
		WriteError(w, err)
		return
	}

	var effectiveTo *time.Time
	if body.EffectiveTo != nil {
		t := body.EffectiveTo.UTC()
		effectiveTo = &t
	}
	var reviewDueAt *time.Time
	if body.ReviewDueAt != nil {
		t := body.ReviewDueAt.UTC()
		reviewDueAt = &t
	}

	result, err := schedulePublish(h, r.Context(), h.runner, application.SchedulePublishRequest{
		TenantID:             tenantID,
		InstanceID:           inst.ID,
		ExpectedRevision:     ifMatchVersion,
		EffectiveDate:        effectiveFrom,
		ScheduledBy:          actorID,
		SupersededDocumentID: body.SupersededDocumentID,
		EffectiveTo:          effectiveTo,
		ReviewDueAt:          reviewDueAt,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.PublishResponse{
		DocumentID:    result.DocumentID,
		NewStatus:     string(docsdomain.DocStatusScheduled),
		EffectiveFrom: result.EffectiveDate.UTC().Format(time.RFC3339),
	})
}
