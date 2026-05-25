package approvalhttp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
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
		WriteError(w, err)
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Create(r.Context(), h.db, application.CreateRouteInput{
		TenantID:    tenantID,
		ProfileCode: req.ProfileCode,
		Name:        req.Name,
		ActorUserID: actorID,
		Stages:      mapStageRequests(req.Stages),
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
	if _, err := parseIfMatch(r.Header.Get("If-Match")); err != nil {
		WriteError(w, err)
		return
	}

	var req contracts.UpdateRouteRequest
	if err := contracts.Decode(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteError(w, err)
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Update(r.Context(), h.db, application.UpdateRouteInput{
		TenantID:    tenantID,
		RouteID:     routeID,
		Name:        req.Name,
		ActorUserID: actorID,
		Stages:      mapStageRequests(req.Stages),
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.RouteResponse{
		RouteID:    result.RouteID,
		NewVersion: result.NewVersion,
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
	if _, err := parseIfMatch(r.Header.Get("If-Match")); err != nil {
		WriteError(w, err)
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Deactivate(r.Context(), h.db, application.DeactivateRouteInput{
		TenantID:    tenantID,
		RouteID:     routeID,
		ActorUserID: actorID,
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

	if h.db == nil {
		WriteError(w, fmt.Errorf("database not configured"))
		return
	}

	if err := h.requireRouteManagement(r.Context(), tenantID, actorID); err != nil {
		WriteError(w, err)
		return
	}

	routeRepo := h.routeRepo
	if routeRepo == nil {
		routeRepo = repository.NewPostgresApprovalRepository(h.db)
	}
	routeRows, err := routeRepo.ListRoutes(r.Context(), tenantID)
	if err != nil {
		WriteError(w, err)
		return
	}

	routes := make([]contracts.ListRouteItem, 0, len(routeRows))
	for _, route := range routeRows {
		routes = append(routes, mapListRoute(route))
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"routes": routes,
		"total":  len(routes),
	})
}

func (h *Handler) requireRouteManagement(ctx context.Context, tenantID, actorID string) error {
	return withAuthzCheck(ctx, h.db, tenantID, actorID, CapManageRoutes, "tenant")
}

func withAuthzCheck(ctx context.Context, db *sql.DB, tenantID, actorID, capability, areaCode string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("authz check: begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return fmt.Errorf("authz check: set tenant GUC: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", actorID); err != nil {
		return fmt.Errorf("authz check: set actor GUC: %w", err)
	}
	return authz.Require(authz.WithCapCache(ctx), tx, capability, areaCode)
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
			Members:            []string{stage.RequiredRole},
			RequiredRole:       stage.RequiredRole,
			RequiredCapability: stage.RequiredCapability,
			AreaCode:           stage.AreaCode,
			QuorumKind:         stage.Quorum,
			DriftPolicy:        stage.DriftPolicy,
		})
	}

	return contracts.ListRouteItem{
		ID:          route.ID,
		Name:        route.Name,
		TenantID:    route.TenantID,
		ProfileCode: route.ProfileCode,
		Active:      route.Active,
		Stages:      stages,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func mapStageRequests(stages []contracts.StageRequest) []domain.Stage {
	out := make([]domain.Stage, 0, len(stages))
	for _, s := range stages {
		out = append(out, domain.Stage{
			Order:              s.Order,
			Name:               s.Name,
			RequiredRole:       s.RequiredRole,
			RequiredCapability: s.RequiredCapability,
			AreaCode:           strings.ToLower(strings.TrimSpace(s.AreaCode)),
			Quorum:             domain.QuorumPolicy(s.Quorum),
			QuorumM:            s.QuorumM,
			OnEligibilityDrift: domain.DriftPolicy(s.DriftPolicy),
		})
	}
	return out
}
