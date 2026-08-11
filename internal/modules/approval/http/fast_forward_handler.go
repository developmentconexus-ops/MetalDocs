package approvalhttp

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/strictjson"
)

// FastForwardHandler records the "Aprovar já" gesture (R5, unit 2.3 G3): one
// caller action records a `ready` review verdict AND an approve signoff as
// two ledger writes in one transaction. Mirrors ReviewVerdictHandler /
// SignoffHandler: requires both an Idempotency-Key header (self-managed
// replay via idempStore, under its OWN route template) and a valid If-Match
// header (OCC precondition); ceremony fields (password_token, content_hash)
// map into the signoff leg exactly as SignoffHandler maps them.
func (h *Handler) FastForwardHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, actorID, ok := requestIdentity(w, r)
	if !ok {
		return
	}
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
	if h.fastForwardSvc == nil {
		WriteError(w, r, errors.New("fast forward service not configured"))
		return
	}
	// Fail closed (no-fallback principle): if the wiring did not produce a
	// fastForwardIdempStore, the handler must not silently skip replay
	// protection on an idempotency guarantee. Refuse the request with a 500
	// rather than proceeding without a duplicate-submission guard.
	if h.fastForwardIdempStore == nil {
		WriteError(w, r, errors.New("fast forward idempotency store not configured"))
		return
	}

	var body contracts.FastForwardRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, r, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, r, NewValidationError(err.Error()))
		return
	}

	payloadHash := fastForwardPayloadHash(instanceID, stageID, body.Comment, body.ContentHash)
	var replayHandle application.SignoffReplayCommitter
	if h.fastForwardIdempStore != nil {
		handle, replay, err := h.fastForwardIdempStore.BeginFastForwardReplay(r.Context(), tenantID, actorID, idempKey, payloadHash)
		if err != nil {
			WriteError(w, r, err)
			return
		}
		if replay != nil {
			WriteJSON(w, r, http.StatusOK, contracts.FastForwardResponse{
				SignoffID: replay.SignoffID,
				WasReplay: true,
				Outcome:   replay.Outcome,
			})
			return
		}
		replayHandle = handle
	}

	result, err := h.fastForwardSvc.RecordFastForward(r.Context(), h.runner, application.FastForwardRequest{
		TenantID:                tenantID,
		InstanceID:              instanceID,
		StageInstanceID:         stageID,
		ActorUserID:             actorID,
		Comment:                 body.Comment,
		ExpectedRevisionVersion: expectedRevisionVersion,
		SignatureMethod:         "password_reauth",
		SignaturePayload:        map[string]any{"password_token": body.PasswordToken},
		ContentHash:             body.ContentHash,
	})
	if err != nil {
		if replayHandle != nil {
			_ = replayHandle.Fail(err)
		}
		WriteError(w, r, err)
		return
	}

	outcome := signoffOutcome(result.Signoff)
	if replayHandle != nil {
		if err := replayHandle.Complete(application.SignoffReplay{
			Outcome:   outcome,
			SignoffID: result.Signoff.SignoffID,
		}); err != nil {
			slog.Warn("fast forward idempotency record failed (non-fatal)", "err", err)
		}
	}

	WriteJSON(w, r, http.StatusOK, contracts.FastForwardResponse{
		SignoffID: result.Signoff.SignoffID,
		WasReplay: false,
		Outcome:   outcome,
	})
}

// fastForwardPayloadHash mirrors reviewVerdictPayloadHash's misuse-guard
// shape: derived only from client-stable request inputs. password_token is
// deliberately excluded (never hashed into the replay fingerprint, matching
// signoffPayloadHash's precedent).
func fastForwardPayloadHash(instanceID, stageID, comment, contentHash string) string {
	payload := strings.Join([]string{
		strings.TrimSpace(instanceID),
		strings.TrimSpace(stageID),
		strings.TrimSpace(comment),
		strings.TrimSpace(contentHash),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
