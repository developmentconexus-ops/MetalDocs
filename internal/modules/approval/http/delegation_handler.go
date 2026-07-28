package approvalhttp

import (
	"errors"
	"net/http"
	"strings"

	approvalapi "metaldocs/internal/modules/approval/api"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/strictjson"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CreateApprovalDelegation implements approvalapi.ServerInterface. Requires
// an Idempotency-Key header (contract-mandated, mirrors route-admin writes).
// delegatorID is ALWAYS derived from the authenticated session — never from
// the request body — so a caller can never grant a delegation "from" another
// user's identity (F9/ADR 0077 privilege-escalation guard).
func (h *Handler) CreateApprovalDelegation(w http.ResponseWriter, r *http.Request, params approvalapi.CreateApprovalDelegationParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)

	// F-QA4-6: the spec now declares Idempotency-Key as format: uuid, so the
	// generated wrapper binds it as a UUID and has already rejected a malformed
	// value before this handler runs. ValidateKey is the fail-closed backstop
	// for any mount that bypasses the generated wrapper (e.g. a unit-test mux).
	idempotencyKey := strings.TrimSpace(params.IdempotencyKey.String())
	if err := idempotency.ValidateKey(idempotencyKey); err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.CreateDelegationRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	if h.services == nil || h.services.Delegation == nil {
		WriteError(w, errors.New("delegation service not configured"))
		return
	}

	d, err := h.services.Delegation.CreateDelegation(r.Context(), h.runner, tenantID, actorID, body.DelegateID, body.Reason, body.StartsAt, body.EndsAt)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, contracts.DelegationResponse{
		ID:          d.ID(),
		TenantID:    d.TenantID(),
		DelegatorID: d.DelegatorID(),
		DelegateID:  d.DelegateID(),
		StartsAt:    d.StartsAt(),
		EndsAt:      d.EndsAt(),
		Reason:      d.Reason(),
		CreatedBy:   d.CreatedBy(),
		CreatedAt:   d.CreatedAt(),
	})
}

// RevokeApprovalDelegation implements approvalapi.ServerInterface. Hard-
// deletes the delegation (F9/ADR 0077 §5 — real-time revocation, no soft
// flag). Only the delegator or a CapApprovalOversee holder may revoke;
// enforced inside application.DelegationService.RevokeDelegation.
func (h *Handler) RevokeApprovalDelegation(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)

	if h.services == nil || h.services.Delegation == nil {
		WriteError(w, errors.New("delegation service not configured"))
		return
	}

	if err := h.services.Delegation.RevokeDelegation(r.Context(), h.runner, tenantID, id.String(), actorID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
