package approvalhttp

import (
	"errors"
	"net/http"
	"strings"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/strictjson"
)

// ExtendSLAHandler postpones the deadline (due_at) of an instance's active
// stage (Task 9, approval accountability loop).
//
// Requires an Idempotency-Key header (the module's standing mutation
// contract), validated and failing closed BEFORE the service runs. No
// If-Match: the forward-only rule is enforced by a read and a write on the
// same row inside one transaction, so an OCC precondition would be contract
// surface solving a problem the transaction already solves (see
// contracts.ExtendSLARequest's doc comment).
//
// On success, the response is built by delegating to GetInstanceHandler —
// the same 200-with-the-current-instance shape a plain GET would return,
// just re-read after the write instead of hand-assembled here. Trade-off:
// the write already committed by the time GetInstanceHandler runs, so if
// THAT read fails (e.g. the instance vanishes from view for some other
// reason) the client sees an error for a mutation that already succeeded.
// A client retrying the extend-sla call on that error would then hit 422
// sla_extension_not_forward, since due_at has already moved forward. This
// is accepted, not fixed, here — re-litigating the response shape is out of
// this task's scope.
func (h *Handler) ExtendSLAHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	instanceID := r.PathValue("instance_id")

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if err := idempotency.ValidateKey(idempotencyKey); err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.ExtendSLARequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	// A blank reason surfaces as the registered 422
	// validation.sla_extension_reason_required (openapi.yaml), not the
	// generic 400 every other Validate() failure produces above — so it is
	// checked directly against the service's sentinel instead of folded into
	// contracts.ExtendSLARequest.Validate (see that method's doc comment).
	// The service re-checks this (trimmed) as the authoritative gate; this
	// is the friendly, fail-fast first line that also matches the
	// contract's advertised status/code before any transaction opens.
	if strings.TrimSpace(body.Reason) == "" {
		WriteError(w, application.ErrSLAExtensionReasonRequired)
		return
	}

	if h.slaExtensionSvc == nil {
		WriteError(w, errors.New("sla extension service not configured"))
		return
	}
	if err := h.slaExtensionSvc.Extend(r.Context(), h.runner, application.ExtendRequest{
		TenantID:   tenantID,
		InstanceID: instanceID,
		ActorID:    actorID,
		NewDueAt:   body.DueAtTime(),
		Reason:     body.Reason,
	}); err != nil {
		WriteError(w, err)
		return
	}

	h.GetInstanceHandler(w, r)
}
