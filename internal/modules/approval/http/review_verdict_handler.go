package approvalhttp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/strictjson"
)

// decodeReviewVerdictRequest reads and validates the ReviewVerdictRequest
// wire body, including the verdict enum constraint that Validate itself
// does not cover.
func decodeReviewVerdictRequest(r *http.Request) (contracts.ReviewVerdictRequest, domain.Verdict, error) {
	var body contracts.ReviewVerdictRequest
	if err := strictjson.Decode(r, &body); err != nil {
		return contracts.ReviewVerdictRequest{}, "", err
	}
	if err := body.Validate(); err != nil {
		return contracts.ReviewVerdictRequest{}, "", NewValidationError(err.Error())
	}
	verdict := domain.Verdict(body.Verdict)
	switch verdict {
	case domain.VerdictReady, domain.VerdictRequestChanges:
	default:
		return contracts.ReviewVerdictRequest{}, "", NewValidationError("verdict must be one of: ready, request_changes")
	}
	return body, verdict, nil
}

// resolveReviewVerdictReplay begins (or replays) a review-verdict idempotency
// slot. handled is true when the caller must return immediately: either the
// begin call failed (error already written) or a prior attempt is being
// replayed (response already written).
func (h *Handler) resolveReviewVerdictReplay(w http.ResponseWriter, r *http.Request, tenantID, actorID, idempKey, payloadHash string) (replayHandle application.SignoffReplayCommitter, handled bool) {
	if h.idempStore == nil {
		return nil, false
	}
	handle, replay, err := h.idempStore.BeginStageReplay(r.Context(), tenantID, actorID, idempKey, payloadHash)
	if err != nil {
		WriteError(w, r, err)
		return nil, true
	}
	if replay != nil {
		WriteJSON(w, r, http.StatusOK, contracts.ReviewVerdictResponse{
			VerdictID: replay.SignoffID,
			WasReplay: true,
			Outcome:   replay.Outcome,
		})
		return nil, true
	}
	return handle, false
}

// ReviewVerdictHandler records a ready/request_changes verdict for the
// requesting actor on a review-kind stage. Mirrors SignoffHandler: requires
// both an Idempotency-Key header (self-managed replay via idempStore, like
// signoffs — NOT the router-level idempotentHandler middleware) and a valid
// If-Match header (OCC precondition).
func (h *Handler) ReviewVerdictHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	actorID := actorIDFromRequest(r)
	instanceID := r.PathValue("instance_id")
	stageID := r.PathValue("stage_id")
	idempKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))

	// F-QA4-6: self-managed replay slot (no idempotency.Require wrapper), so the
	// shared UUID wire rule is enforced here.
	if err := idempotency.ValidateKey(idempKey); err != nil {
		WriteError(w, r, err)
		return
	}
	expectedRevisionVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if h.reviewVerdictSvc == nil {
		WriteError(w, r, errors.New("review verdict service not configured"))
		return
	}

	body, verdict, err := decodeReviewVerdictRequest(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	payloadHash := reviewVerdictPayloadHash(instanceID, stageID, verdict, body.Comment)
	replayHandle, handled := h.resolveReviewVerdictReplay(w, r, tenantID, actorID, idempKey, payloadHash)
	if handled {
		return
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
		WriteError(w, r, err)
		return
	}

	outcome := reviewVerdictOutcome(result)
	if replayHandle != nil {
		if err := replayHandle.Complete(application.SignoffReplay{
			Outcome:   outcome,
			SignoffID: result.VerdictID,
		}); err != nil {
			slog.Warn("review verdict idempotency record failed (non-fatal)", "err", err)
		}
	}

	resp := contracts.ReviewVerdictResponse{
		VerdictID:           result.VerdictID,
		WasReplay:           false,
		Outcome:             outcome,
		FastForwardEligible: result.FastForwardEligible,
	}
	if result.NextStageID != nil {
		resp.NextStageID = *result.NextStageID
	}
	WriteJSON(w, r, http.StatusOK, resp)
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
