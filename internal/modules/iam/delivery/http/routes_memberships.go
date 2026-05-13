package httpdelivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type MembershipHandler struct {
	svc *iamapp.AreaMembershipService
}

type grantMembershipRequest struct {
	UserID   string `json:"userId"`
	AreaCode string `json:"areaCode"`
	Role     string `json:"role"`
}

func NewMembershipHandler(svc *iamapp.AreaMembershipService) *MembershipHandler {
	return &MembershipHandler{svc: svc}
}

func (h *MembershipHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/iam/area-memberships", h.listMemberships)
	mux.HandleFunc("POST /api/v1/iam/area-memberships", h.grantMembership)
	mux.HandleFunc("DELETE /api/v1/iam/area-memberships", h.revokeMembership)
}

func (h *MembershipHandler) listMemberships(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership service is not configured"))
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("userId"))
	if userID == "" {
		userID = strings.TrimSpace(authenticatedActor(r))
	}
	if userID == "" {
		_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId is required"))
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	items, err := h.svc.ListActive(r.Context(), userID, tenantID)
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list memberships"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *MembershipHandler) grantMembership(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership service is not configured"))
		return
	}

	var req grantMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	if strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.AreaCode) == "" || strings.TrimSpace(req.Role) == "" {
		_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId, areaCode and role are required"))
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	grantedBy := authn.UserIDFromContext(r.Context())
	err = h.svc.Grant(
		r.Context(),
		strings.TrimSpace(req.UserID),
		tenantID,
		strings.TrimSpace(req.AreaCode),
		iamdomain.Role(strings.ToLower(strings.TrimSpace(req.Role))),
		grantedBy,
	)
	if err != nil {
		h.writeMembershipError(w, err, "Failed to grant membership")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"userId":   strings.TrimSpace(req.UserID),
		"tenantId": tenantID,
		"areaCode": strings.TrimSpace(req.AreaCode),
		"role":     strings.ToLower(strings.TrimSpace(req.Role)),
	})
}

func (h *MembershipHandler) revokeMembership(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership service is not configured"))
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("userId"))
	areaCode := strings.TrimSpace(r.URL.Query().Get("areaCode"))
	if userID == "" || areaCode == "" {
		_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId and areaCode are required"))
		return
	}
	revokedBy := authn.UserIDFromContext(r.Context())

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	err = h.svc.Revoke(r.Context(), userID, tenantID, areaCode, revokedBy)
	if err != nil {
		h.writeMembershipError(w, err, "Failed to revoke membership")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MembershipHandler) writeMembershipError(w http.ResponseWriter, err error, defaultMessage string) {
	switch {
	case errors.Is(err, iamapp.ErrMembershipNotFound):
		_ = problem.Write(w, problem.New(http.StatusNotFound, "MEMBERSHIP_NOT_FOUND", "Membership not found"))
	case errors.Is(err, iamapp.ErrUnknownRole):
		_ = problem.Write(w, problem.New(http.StatusBadRequest, "UNKNOWN_ROLE", "Unknown role"))
	default:
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", defaultMessage))
	}
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}
