package approvalhttp

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
)

const CapManageRoutes = "route.admin"

func (h *Handler) CreateRouteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}

	var req contracts.CreateRouteRequest
	if err := contracts.Decode(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Create(r.Context(), h.db, application.CreateRouteInput{
		TenantID:       tenantID,
		ProfileCode:    req.ProfileCode,
		Name:           req.Name,
		ActorUserID:    actorID,
		IdempotencyKey: idempotencyKey,
		Stages:         mapStageRequests(req.Stages),
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, contracts.RouteResponse{
		RouteID: result.RouteID,
	})
}

func (h *Handler) UpdateRouteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	routeID := r.PathValue("id")
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}
	expectedVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var req contracts.UpdateRouteRequest
	if err := contracts.Decode(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Update(r.Context(), h.db, application.UpdateRouteInput{
		TenantID:        tenantID,
		RouteID:         routeID,
		Name:            req.Name,
		ActorUserID:     actorID,
		IdempotencyKey:  idempotencyKey,
		ExpectedVersion: expectedVersion,
		Stages:          mapStageRequests(req.Stages),
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.RouteResponse{
		RouteID:    result.RouteID,
		NewVersion: intPtr(result.NewVersion),
	})
}

func (h *Handler) DeactivateRouteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	routeID := r.PathValue("id")
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}
	expectedVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.DeactivateRouteRequest
	if err := contracts.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}
	reason := strings.TrimSpace(body.Reason)

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Deactivate(r.Context(), h.db, application.DeactivateRouteInput{
		TenantID:        tenantID,
		RouteID:         routeID,
		ActorUserID:     actorID,
		IdempotencyKey:  idempotencyKey,
		Reason:          reason,
		ExpectedVersion: expectedVersion,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.RouteResponse{
		RouteID: result.RouteID,
	})
}

func (h *Handler) ListRoutesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	out, err := routeAdminSvc.List(r.Context(), h.db, tenantID, actorID)
	if err != nil {
		WriteError(w, err)
		return
	}

	routes := make([]contracts.ListRouteItem, 0, len(out.Routes))
	for _, route := range out.Routes {
		routes = append(routes, mapListRoute(route))
	}

	WriteJSON(w, http.StatusOK, contracts.ListRoutesResponse{
		Routes: routes,
		Total:  len(routes),
	})
}

func mapListRoute(route repository.Route) contracts.ListRouteItem {
	createdAt := route.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := route.UpdatedAt.UTC().Format(time.RFC3339)
	if route.UpdatedAt.IsZero() {
		updatedAt = createdAt
	}

	stages := make([]contracts.ListStageItem, 0, len(route.Stages))
	for _, stage := range route.Stages {
		stages = append(stages, contracts.ListStageItem{
			Label:              stage.Name,
			RequiredRole:       stage.RequiredRole,
			RequiredCapability: stage.RequiredCapability,
			AreaCode:           stage.AreaCode,
			QuorumKind:         contracts.QuorumKind(stage.Quorum),
			QuorumM:            stage.QuorumM,
			DriftPolicy:        contracts.DriftPolicyKind(stage.DriftPolicy),
		})
	}

	return contracts.ListRouteItem{
		ID:          route.ID,
		Name:        route.Name,
		TenantID:    route.TenantID,
		ProfileCode: route.ProfileCode,
		Active:      route.Active,
		Version:     route.Version,
		Stages:      stages,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func mapStageRequests(stages []contracts.StageRequest) []domain.Stage {
	out := make([]domain.Stage, 0, len(stages))
	for _, s := range stages {
		cap := strings.TrimSpace(s.RequiredCapability)
		if cap == "" {
			cap = "workflow.sign"
		}
		out = append(out, domain.Stage{
			Order:              s.Order,
			Name:               s.Name,
			RequiredRole:       strings.ToLower(strings.TrimSpace(s.RequiredRole)),
			RequiredCapability: cap,
			AreaCode:           strings.ToLower(strings.TrimSpace(s.AreaCode)),
			Quorum:             domain.QuorumPolicy(s.Quorum),
			QuorumM:            s.QuorumM,
			OnEligibilityDrift: domain.DriftPolicy(s.DriftPolicy),
		})
	}
	return out
}

func intPtr(v int) *int {
	return &v
}
