// tenant_handler.go — M7 F7.2 Task C: POST /tenants (operationId
// onboardTenant). Replaces the Task B 501 stub (Router.OnboardTenant) with
// real delegation to iamapp.OnboardTenantService.
package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	iamapi "metaldocs/internal/modules/iam/api"
	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// TenantOnboarder is the narrow application-layer surface TenantHandler
// depends on. Implemented by *iamapp.OnboardTenantService; kept as an
// interface so the handler depends on a capability, not the concrete service.
type TenantOnboarder interface {
	OnboardTenant(ctx context.Context, in iamapp.OnboardTenantInput, actorID, actorTenantID string) (iamapp.OnboardTenantResult, error)
}

// TenantHandler serves POST /tenants (tenant onboarding, M7 F7.2).
type TenantHandler struct {
	service TenantOnboarder
}

// NewTenantHandler constructs the handler. service may be nil, in which case
// handleOnboardTenant answers 501 (matching the pre-Task-C stub behavior for
// boot paths without a configured SQLDB).
func NewTenantHandler(service TenantOnboarder) *TenantHandler {
	return &TenantHandler{service: service}
}

func (h *TenantHandler) handleOnboardTenant(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeIAMNotImplemented(w, "Tenant onboarding service is not configured")
		return
	}

	var req iamapi.OnboardTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, "Invalid JSON payload"))
		return
	}

	// actorID is the authenticated principal performing the onboarding and
	// becomes the audit ActorID / tier-2 authz.Require actor; actorTenantID is
	// that principal's OWN tenant (grants are tenant-scoped, so authz.Require
	// resolves under it, not the new tenant). Both come from the request
	// context, never the client payload (mirrors admin_handler.go).
	actorID := authenticatedActor(r)
	actorTenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "internal server error"))
		return
	}

	result, err := h.service.OnboardTenant(r.Context(), iamapp.OnboardTenantInput{
		Name:             req.Name,
		Slug:             req.Slug,
		AdminUserID:      req.AdminUserId,
		AdminDisplayName: req.AdminDisplayName,
		AdminPassword:    req.AdminPassword,
	}, actorID, actorTenantID)
	if err != nil {
		h.writeOnboardError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, iamapi.OnboardTenantResponse{
		TenantId:    mustParseTenantUUID(result.TenantID),
		AdminUserId: result.AdminUserID,
	})
}

// mustParseTenantUUID parses the tenant id produced by OnboardTenantService
// (always a freshly minted uuid.NewString()); a parse failure here would be a
// service-layer defect, not a client input error, so it panics (caught by
// platform/middleware.Recovery) rather than silently returning a zero UUID.
func mustParseTenantUUID(id string) openapi_types.UUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		panic("iam: OnboardTenantService returned a non-UUID tenant id: " + err.Error())
	}
	return openapi_types.UUID(parsed)
}

func (h *TenantHandler) writeOnboardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, iamdomain.ErrTenantOnboardValidation):
		h.writeProblem(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, err.Error()))
	case errors.Is(err, iamdomain.ErrTenantSlugConflict):
		h.writeProblem(w, problem.New(http.StatusConflict, problem.CodeConflict, "Tenant slug already exists"))
	case errors.Is(err, iamdomain.ErrTenantAdminUserConflict):
		h.writeProblem(w, problem.New(http.StatusConflict, problem.CodeConflict, "Admin user already exists"))
	case isCapabilityDenied(err):
		h.writeProblem(w, problem.New(http.StatusForbidden, problem.CodeForbiddenCapability, "tenant.onboard capability required"))
	default:
		slog.Error("iam: onboard tenant failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Failed to onboard tenant"))
	}
}

func (h *TenantHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if werr := problem.Write(w, p); werr != nil {
		slog.Warn("iam: write response failed", "err", werr)
	}
}

// isCapabilityDenied reports whether err is (or wraps) authz.ErrCapDenied.
func isCapabilityDenied(err error) bool {
	var denied authz.ErrCapDenied
	return errors.As(err, &denied)
}
