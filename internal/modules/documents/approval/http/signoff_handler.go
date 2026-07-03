package approvalhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/platform/strictjson"
)

var (
	// ErrIdempotencyRequired is returned when a mutating request is missing the Idempotency-Key header.
	ErrIdempotencyRequired = errors.New("idempotency: Idempotency-Key header required on mutating requests")

	// ErrContentHashMismatch re-exports application.ErrContentHashMismatch for HTTP-layer error mapping.
	ErrContentHashMismatch = application.ErrContentHashMismatch
)

// SignoffHandler records an approve/reject decision for the requesting actor on
// a stage. Requires both an Idempotency-Key header (replay-safe via idempStore)
// and a valid If-Match header (OCC precondition).
func (h *Handler) SignoffHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	instanceID := r.PathValue("instance_id")
	stageID := r.PathValue("stage_id")
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
	if h.decisionSvc == nil {
		WriteError(w, errors.New("decision service not configured"))
		return
	}

	var body contracts.SignoffRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}
	decision := domain.Decision(body.Decision)
	switch decision {
	case domain.DecisionApprove, domain.DecisionReject:
	default:
		WriteError(w, NewValidationError("decision must be one of: approve, reject"))
		return
	}

	payloadHash := signoffPayloadHash("", instanceID, stageID, decision, body.Reason, body.ContentHash)
	var replayHandle interface {
		Complete(outcome string) error
		Fail(cause error) error
	}
	if h.idempStore != nil {
		handle, replay, err := h.idempStore.BeginStageReplay(r.Context(), tenantID, actorID, idempKey, payloadHash)
		if err != nil {
			WriteError(w, err)
			return
		}
		if replay != nil {
			WriteJSON(w, http.StatusOK, contracts.SignoffResponse{
				WasReplay: true,
				Outcome:   replay.Outcome,
			})
			return
		}
		replayHandle = handle
	}

	result, err := h.decisionSvc.RecordSignoff(r.Context(), h.runner, application.SignoffRequest{
		TenantID:                tenantID,
		InstanceID:              instanceID,
		StageInstanceID:         stageID,
		ActorUserID:             actorID,
		Decision:                decision,
		Comment:                 body.Reason,
		SignatureMethod:         "password_reauth",
		SignaturePayload:        map[string]any{"password_token": body.PasswordToken},
		ContentFormData:         map[string]any{"_content_hash": body.ContentHash},
		ExpectedRevisionVersion: expectedRevisionVersion,
	})
	if err != nil {
		if replayHandle != nil {
			_ = replayHandle.Fail(err)
		}
		WriteError(w, err)
		return
	}

	outcome := signoffOutcome(result)
	if replayHandle != nil {
		if err := replayHandle.Complete(outcome); err != nil {
			slog.Warn("signoff idempotency record failed (non-fatal)", "err", err)
		}
	}

	WriteJSON(w, http.StatusOK, contracts.SignoffResponse{
		SignoffID: "",
		WasReplay: false,
		Outcome:   outcome,
	})
}

func signoffOutcome(result application.SignoffResult) string {
	switch {
	case result.InstanceApproved:
		return "approved"
	case result.InstanceRejected:
		return "rejected"
	case result.StageCompleted:
		return "stage_completed"
	default:
		return "pending"
	}
}
