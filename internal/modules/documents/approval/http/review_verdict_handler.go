package approvalhttp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/platform/strictjson"
)

// ReviewVerdictHandler records a ready/request_changes verdict for the
// requesting actor on a review-kind stage. Mirrors SignoffHandler: requires
// both an Idempotency-Key header (self-managed replay via idempStore, like
// signoffs — NOT the router-level idempotentHandler middleware) and a valid
// If-Match header (OCC precondition).
func (h *Handler) ReviewVerdictHandler(w http.ResponseWriter, r *http.Request) {
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
	if h.reviewVerdictSvc == nil {
		WriteError(w, errors.New("review verdict service not configured"))
		return
	}

	var body contracts.ReviewVerdictRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}
	verdict := domain.Verdict(body.Verdict)
	switch verdict {
	case domain.VerdictReady, domain.VerdictRequestChanges:
	default:
		WriteError(w, NewValidationError("verdict must be one of: ready, request_changes"))
		return
	}

	payloadHash := reviewVerdictPayloadHash(instanceID, stageID, verdict, body.Comment)
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
			WriteJSON(w, http.StatusOK, contracts.ReviewVerdictResponse{
				WasReplay: true,
				Outcome:   replay.Outcome,
			})
			return
		}
		replayHandle = handle
	}

	result, err := h.reviewVerdictSvc.RecordVerdict(r.Context(), h.runner, application.ReviewVerdictRequest{
		TenantID:                tenantID,
		InstanceID:              instanceID,
		StageInstanceID:         stageID,
		ActorUserID:             actorID,
		Verdict:                 verdict,
		Comment:                 body.Comment,
		ExpectedRevisionVersion: expectedRevisionVersion,
	})
	if err != nil {
		if replayHandle != nil {
			_ = replayHandle.Fail(err)
		}
		WriteError(w, err)
		return
	}

	outcome := reviewVerdictOutcome(result)
	if replayHandle != nil {
		if err := replayHandle.Complete(outcome); err != nil {
			slog.Warn("review verdict idempotency record failed (non-fatal)", "err", err)
		}
	}

	resp := contracts.ReviewVerdictResponse{
		VerdictID:           "",
		WasReplay:           false,
		Outcome:             outcome,
		FastForwardEligible: result.FastForwardEligible,
	}
	if result.NextStageID != nil {
		resp.NextStageID = *result.NextStageID
	}
	WriteJSON(w, http.StatusOK, resp)
}

func reviewVerdictOutcome(result application.ReviewVerdictResult) string {
	switch {
	case result.InstanceApproved:
		return "approved"
	case result.ChangesRequested:
		return "changes_requested"
	case result.StageCompleted:
		return "stage_completed"
	default:
		return "pending"
	}
}

// reviewVerdictPayloadHash mirrors signoffPayloadHash's misuse-guard shape
// (derived only from client-stable request inputs).
func reviewVerdictPayloadHash(instanceID, stageID string, verdict domain.Verdict, comment string) string {
	payload := strings.Join([]string{
		strings.TrimSpace(instanceID),
		strings.TrimSpace(stageID),
		string(verdict),
		strings.TrimSpace(comment),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
