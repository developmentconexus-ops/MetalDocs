package approvalhttp

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/http/contracts"
)

var (
	publishApproved = func(h *Handler, ctx context.Context, db *sql.DB, req application.PublishRequest) (application.PublishResult, error) {
		if h.services == nil || h.services.Publish == nil {
			return application.PublishResult{}, errors.New("publish service not configured")
		}
		return h.services.Publish.PublishApproved(ctx, db, req)
	}
	schedulePublish = func(h *Handler, ctx context.Context, db *sql.DB, req application.SchedulePublishRequest) (application.SchedulePublishResult, error) {
		if h.services == nil || h.services.Publish == nil {
			return application.SchedulePublishResult{}, errors.New("publish service not configured")
		}
		return h.services.Publish.SchedulePublish(ctx, db, req)
	}
)

func (h *Handler) PublishHandler(w http.ResponseWriter, r *http.Request) {
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

	if _, err := parseIfMatch(r.Header.Get("If-Match")); err != nil {
		WriteError(w, err)
		return
	}

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.db, tenantID, documentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	result, err := publishApproved(h, r.Context(), h.db, application.PublishRequest{
		TenantID:    tenantID,
		InstanceID:  inst.ID,
		PublishedBy: actorID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.PublishResponse{
		DocumentID: result.DocumentID,
		NewStatus:  "published",
	})
}

func (h *Handler) SchedulePublishHandler(w http.ResponseWriter, r *http.Request) {
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

	ifMatchVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.db, tenantID, documentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.SchedulePublishRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	effectiveFrom, err := time.Parse(time.RFC3339, body.EffectiveFrom)
	if err != nil {
		WriteError(w, err)
		return
	}

	result, err := schedulePublish(h, r.Context(), h.db, application.SchedulePublishRequest{
		TenantID:             tenantID,
		InstanceID:           inst.ID,
		ExpectedRevision:     ifMatchVersion,
		EffectiveDate:        effectiveFrom,
		ScheduledBy:          actorID,
		SupersededDocumentID: body.SupersededDocumentID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.PublishResponse{
		DocumentID:    result.DocumentID,
		NewStatus:     "scheduled",
		EffectiveFrom: result.EffectiveDate.UTC().Format(time.RFC3339),
	})
}
