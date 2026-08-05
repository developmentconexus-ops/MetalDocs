package approvalhttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/strictjson"
)

var extendSLA = func(h *Handler, ctx context.Context, runner db.TxRunner, req application.ExtendRequest) error {
	if h.slaExtensionSvc == nil {
		return errors.New("sla extension service not configured")
	}
	return h.slaExtensionSvc.Extend(ctx, runner, req)
}

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
// just re-read after the write instead of hand-assembled here.
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

	newDueAt, err := body.DueAtTime()
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := extendSLA(h, r.Context(), h.runner, application.ExtendRequest{
		TenantID:   tenantID,
		InstanceID: instanceID,
		ActorID:    actorID,
		NewDueAt:   newDueAt,
		Reason:     body.Reason,
	}); err != nil {
		WriteError(w, err)
		return
	}

	h.GetInstanceHandler(w, r)
}
