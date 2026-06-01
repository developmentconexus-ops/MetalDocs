package approvalhttp

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	approvalinfra "metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/documents/approval/repository"
)

type mutationByDocumentReadService interface {
	LoadActiveInstanceByDocumentForMutation(ctx context.Context, db *sql.DB, tenantID, documentID string) (*domain.Instance, error)
}

func (h *Handler) loadActiveInstanceByDocumentForMutation(r *http.Request, tenantID, docID string) (*domain.Instance, error) {
	mutationLookup, ok := h.readSvc.(mutationByDocumentReadService)
	if !ok {
		return nil, errors.New("mutation lookup service not configured")
	}
	return mutationLookup.LoadActiveInstanceByDocumentForMutation(r.Context(), h.db, tenantID, docID)
}

// failReplaySlot releases an in-flight idempotency slot on an error exit so it
// is not orphaned until expiry. A non-nil release error is logged (not fatal):
// the primary error is already being returned to the client, and the platform
// store reclaims expired orphans on its own.
func failReplaySlot(handle approvalinfra.SignoffReplayCommitter, cause error) {
	if handle == nil {
		return
	}
	if err := handle.Fail(cause); err != nil {
		log.Printf("WARN signoff idempotency slot release failed (non-fatal): %v", err)
	}
}

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
	expectedRevisionVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
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
	contractReq := contracts.SignoffRequest{
		Decision:      contracts.Decision(strings.TrimSpace(body.Decision)),
		Reason:        strings.TrimSpace(body.Reason),
		PasswordToken: strings.TrimSpace(body.Password),
		ContentHash:   strings.TrimSpace(body.ContentHash),
	}
	if err := contractReq.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}
	decision := domain.Decision(contractReq.Decision)

	// The replay fingerprint must depend only on client-stable request inputs
	// (the path-supplied document ID plus the request body), never on
	// server-resolved instance/stage state. The active instance and stage rotate
	// or go terminal between attempts; hashing them would make an identical retry
	// produce a different fingerprint and trip the misuse guard with a spurious
	// conflict instead of replaying the recorded outcome. The route template
	// carries no document ID, so docID is also what isolates distinct documents
	// under a reused key.
	payloadHash := signoffPayloadHash(docID, "", "", decision, contractReq.Reason, contractReq.ContentHash)
	var replayHandle approvalinfra.SignoffReplayCommitter
	if h.idempStore != nil {
		handle, replay, err := h.idempStore.BeginDocumentReplay(r.Context(), tenantID, actorID, idempKey, payloadHash)
		if err != nil {
			WriteError(w, err)
			return
		}
		if replay != nil {
			WriteJSON(w, http.StatusOK, contracts.SignoffResponse{WasReplay: true, Outcome: replay.Outcome})
			return
		}
		replayHandle = handle
	}

	inst, err := h.loadActiveInstanceByDocumentForMutation(r, tenantID, docID)
	if err != nil {
		failReplaySlot(replayHandle, err)
		if errors.Is(err, repository.ErrNoActiveInstance) {
			WriteError(w, repository.ErrNoActiveInstance)
			return
		}
		WriteError(w, err)
		return
	}

	activeStage := inst.Active()
	if activeStage == nil {
		failReplaySlot(replayHandle, repository.ErrInstanceCompleted)
		WriteError(w, repository.ErrInstanceCompleted)
		return
	}
	if err := domain.CheckEligibility(actorID, activeStage.EligibleActorIDs); err != nil {
		failReplaySlot(replayHandle, err)
		WriteError(w, err)
		return
	}

	result, err := h.decisionSvc.RecordSignoff(r.Context(), h.db, application.SignoffRequest{
		TenantID:                tenantID,
		InstanceID:              inst.ID,
		StageInstanceID:         activeStage.ID,
		ActorUserID:             actorID,
		Decision:                decision,
		Comment:                 contractReq.Reason,
		SignatureMethod:         "password_reauth",
		SignaturePayload:        map[string]any{"password_token": contractReq.PasswordToken},
		ContentFormData:         map[string]any{"_content_hash": contractReq.ContentHash},
		ExpectedRevisionVersion: expectedRevisionVersion,
	})
	if err != nil {
		failReplaySlot(replayHandle, err)
		WriteError(w, err)
		return
	}

	outcome := signoffOutcome(result)
	if replayHandle != nil {
		if err := replayHandle.Complete(outcome); err != nil {
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

	inst, err := h.loadActiveInstanceByDocumentForMutation(r, tenantID, docID)
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
