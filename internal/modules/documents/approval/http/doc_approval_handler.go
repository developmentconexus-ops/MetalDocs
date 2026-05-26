package approvalhttp

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
)

// GetInstanceByDocumentHandler handles GET /api/v1/documents/{id}/approval-instance.
// It looks up the active approval instance for the document and returns it.
func (h *Handler) GetInstanceByDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	docID := r.PathValue("id")

	if h.readSvc == nil {
		WriteError(w, errors.New("read service not configured"))
		return
	}

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.db, tenantID, docID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveInstance) {
			WriteError(w, repository.ErrNoActiveInstance)
			return
		}
		WriteError(w, err)
		return
	}

	resp, err := h.mapInstanceResponse(r.Context(), tenantID, inst)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("ETag", resp.ETag)
	WriteJSON(w, http.StatusOK, resp)
}

// docSignoffRequest accepts the frontend's field name ("password" instead of "password_token").
type docSignoffRequest struct {
	Decision    string `json:"decision"`
	Reason      string `json:"reason"`
	Password    string `json:"password"`
	ContentHash string `json:"content_hash"`
}

// SignoffByDocumentHandler handles POST /api/v1/documents/{id}/signoff.
// It finds the active instance+stage for the document and records the signoff.
func (h *Handler) SignoffByDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	docID := r.PathValue("id")
	idempKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))

	if idempKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}
	if h.decisionSvc == nil || h.readSvc == nil {
		WriteError(w, errors.New("services not configured"))
		return
	}

	var body docSignoffRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	decision := domain.Decision(strings.TrimSpace(body.Decision))
	if decision != domain.DecisionApprove && decision != domain.DecisionReject {
		WriteError(w, NewValidationError("decision must be one of: approve, reject"))
		return
	}
	if decision == domain.DecisionReject && strings.TrimSpace(body.Reason) == "" {
		WriteError(w, NewValidationError("reason is required for reject"))
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		WriteError(w, NewValidationError("password is required"))
		return
	}

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.db, tenantID, docID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveInstance) {
			WriteError(w, repository.ErrNoActiveInstance)
			return
		}
		WriteError(w, err)
		return
	}

	activeStage := inst.Active()
	if activeStage == nil {
		WriteError(w, repository.ErrInstanceCompleted)
		return
	}
	if err := domain.CheckEligibility(actorID, activeStage.EligibleActorIDs); err != nil {
		WriteError(w, err)
		return
	}

	if h.idempStore != nil {
		found, outcome, err := h.idempStore.CheckReplay(r.Context(), tenantID, actorID, idempKey)
		if err != nil {
			WriteError(w, err)
			return
		}
		if found {
			WriteJSON(w, http.StatusOK, contracts.SignoffResponse{WasReplay: true, Outcome: outcome})
			return
		}
	}

	result, err := h.decisionSvc.RecordSignoff(r.Context(), h.db, application.SignoffRequest{
		TenantID:         tenantID,
		InstanceID:       inst.ID,
		StageInstanceID:  activeStage.ID,
		ActorUserID:      actorID,
		Decision:         decision,
		Comment:          body.Reason,
		SignatureMethod:  "password_reauth",
		SignaturePayload: map[string]any{"password_token": body.Password},
		ContentFormData:  map[string]any{"_content_hash": body.ContentHash},
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	outcome := signoffOutcome(result)
	if h.idempStore != nil {
		if err := h.idempStore.RecordReplay(r.Context(), tenantID, actorID, idempKey, outcome); err != nil {
			log.Printf("WARN signoff idempotency record failed (non-fatal): %v", err)
		}
	}
	WriteJSON(w, http.StatusOK, contracts.SignoffResponse{
		SignoffID: "",
		WasReplay: false,
		Outcome:   outcome,
	})
}

// CancelByDocumentHandler handles POST /api/v1/documents/{id}/cancel.
// It finds the active instance for the document and cancels it.
func (h *Handler) CancelByDocumentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	docID := r.PathValue("id")
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
	if h.readSvc == nil {
		WriteError(w, errors.New("read service not configured"))
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

	inst, err := h.readSvc.LoadActiveInstanceByDocument(r.Context(), h.db, tenantID, docID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveInstance) {
			WriteError(w, repository.ErrNoActiveInstance)
			return
		}
		WriteError(w, err)
		return
	}

	result, err := h.cancelInstance(r.Context(), h.db, application.CancelInput{
		TenantID:                tenantID,
		InstanceID:              inst.ID,
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
