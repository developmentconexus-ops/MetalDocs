package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
)

// RouteAdminService manages approval route configuration changes.
type RouteAdminService struct {
	repo    repository.ApprovalRepository
	emitter EventEmitter
	clock   Clock
}

// ErrRouteNotFound is returned when a route does not exist for the tenant.
var ErrRouteNotFound = errors.New("route_admin: route not found")

// ErrRouteAlreadyInactive is returned when the route is already inactive.
var ErrRouteAlreadyInactive = errors.New("route_admin: route already inactive")

// CreateRouteInput carries all inputs for Create.
type CreateRouteInput struct {
	TenantID    string
	ProfileCode string
	Name        string
	ActorUserID string
	Stages      []domain.Stage
}

// CreateRouteResult is returned on successful route creation.
type CreateRouteResult struct {
	RouteID string
}

// UpdateRouteInput carries all inputs for Update.
type UpdateRouteInput struct {
	TenantID        string
	RouteID         string
	Name            string
	ActorUserID     string
	ExpectedVersion int
	Stages          []domain.Stage
}

// UpdateRouteResult is returned on successful route update.
type UpdateRouteResult struct {
	RouteID    string
	NewVersion int
}

// DeactivateRouteInput carries all inputs for Deactivate.
type DeactivateRouteInput struct {
	TenantID        string
	RouteID         string
	ActorUserID     string
	ExpectedVersion int
}

// DeactivateRouteResult is returned on successful route deactivation.
type DeactivateRouteResult struct {
	RouteID string
}

// Create creates a new approval route and all route stages.
func (s *RouteAdminService) Create(ctx context.Context, db *sql.DB, in CreateRouteInput) (CreateRouteResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return CreateRouteResult{}, fmt.Errorf("route_admin: begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, fmt.Errorf("route_admin create: %w", err)
	}
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "route.admin", "tenant"); err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, err
	}

	route := domain.Route{
		TenantID:    in.TenantID,
		ProfileCode: in.ProfileCode,
		Version:     1,
		Stages:      in.Stages,
	}
	if err := route.Validate(); err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, err
	}

	var routeID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO approval_routes
			(tenant_id, profile_code, name, version, created_by, active)
		VALUES ($1, $2, $3, 1, $4, TRUE)
		RETURNING id`,
		in.TenantID, in.ProfileCode, in.Name, in.ActorUserID,
	).Scan(&routeID)
	if err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, fmt.Errorf("route_admin: insert route: %w", repository.MapPgError(err, repository.MapHints{}))
	}

	if err := insertRouteStages(ctx, tx, routeID, in.Stages); err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, err
	}

	payload, err := json.Marshal(map[string]any{
		"route_id":      routeID,
		"profile_code":  in.ProfileCode,
		"stage_count":   len(in.Stages),
		"initial_state": "active",
	})
	if err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, fmt.Errorf("route_admin: marshal event payload: %w", err)
	}
	if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
		TenantID:     in.TenantID,
		EventType:    "route.config.created",
		ActorUserID:  in.ActorUserID,
		ResourceType: "approval_route",
		ResourceID:   routeID,
		PayloadJSON:  payload,
		OccurredAt:   s.clock.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, fmt.Errorf("route_admin: emit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CreateRouteResult{}, fmt.Errorf("route_admin: commit: %w", err)
	}
	return CreateRouteResult{RouteID: routeID}, nil
}

// Update updates route metadata and replaces all route stages atomically.
func (s *RouteAdminService) Update(ctx context.Context, db *sql.DB, in UpdateRouteInput) (UpdateRouteResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return UpdateRouteResult{}, fmt.Errorf("route_admin: begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, fmt.Errorf("route_admin update: %w", err)
	}
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "route.admin", "tenant"); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, err
	}

	if _, err := lockRouteForUpdate(ctx, tx, in.TenantID, in.RouteID); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, err
	}

	route := domain.Route{
		ID:       in.RouteID,
		TenantID: in.TenantID,
		Stages:   in.Stages,
	}
	if err := route.Validate(); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, err
	}

	var newVersion int
	err = tx.QueryRowContext(ctx, `
		UPDATE approval_routes
		   SET name = $1,
		       version = version + 1
		 WHERE id = $2
		   AND tenant_id = $3
		   AND ($4 = 0 OR version = $4)
		RETURNING version`,
		in.Name, in.RouteID, in.TenantID, in.ExpectedVersion,
	).Scan(&newVersion)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return UpdateRouteResult{}, repository.ErrStaleRevision
		}
		mapped := repository.MapPgError(err, repository.MapHints{})
		if errors.Is(mapped, repository.ErrRouteInUse) {
			return UpdateRouteResult{}, mapped
		}
		return UpdateRouteResult{}, fmt.Errorf("route_admin: update route: %w", mapped)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM approval_route_stages
		WHERE route_id = $1`,
		in.RouteID,
	); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, fmt.Errorf("route_admin: delete stages: %w", err)
	}

	if err := insertRouteStages(ctx, tx, in.RouteID, in.Stages); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, err
	}

	payload, err := json.Marshal(map[string]any{
		"route_id":    in.RouteID,
		"new_version": newVersion,
		"stage_count": len(in.Stages),
	})
	if err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, fmt.Errorf("route_admin: marshal event payload: %w", err)
	}
	if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
		TenantID:     in.TenantID,
		EventType:    "route.config.updated",
		ActorUserID:  in.ActorUserID,
		ResourceType: "approval_route",
		ResourceID:   in.RouteID,
		PayloadJSON:  payload,
		OccurredAt:   s.clock.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, fmt.Errorf("route_admin: emit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return UpdateRouteResult{}, fmt.Errorf("route_admin: commit: %w", err)
	}
	return UpdateRouteResult{RouteID: in.RouteID, NewVersion: newVersion}, nil
}

// Deactivate marks a route inactive.
func (s *RouteAdminService) Deactivate(ctx context.Context, db *sql.DB, in DeactivateRouteInput) (DeactivateRouteResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return DeactivateRouteResult{}, fmt.Errorf("route_admin: begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, fmt.Errorf("route_admin deactivate: %w", err)
	}
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "route.admin", "tenant"); err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, err
	}

	lockedRoute, err := lockRouteForUpdate(ctx, tx, in.TenantID, in.RouteID)
	if err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, err
	}
	if !lockedRoute.Active {
		if in.ExpectedVersion > 0 && lockedRoute.Version != in.ExpectedVersion {
			_ = tx.Rollback()
			return DeactivateRouteResult{}, repository.ErrStaleRevision
		}
		_ = tx.Rollback()
		return DeactivateRouteResult{}, ErrRouteAlreadyInactive
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE approval_routes
		   SET active = FALSE,
		       version = version + 1
		 WHERE id = $1
		   AND tenant_id = $2
		   AND active = TRUE
		   AND ($3 = 0 OR version = $3)`,
		in.RouteID, in.TenantID, in.ExpectedVersion,
	)
	if err != nil {
		_ = tx.Rollback()
		mapped := repository.MapPgError(err, repository.MapHints{})
		if errors.Is(mapped, repository.ErrRouteInUse) {
			return DeactivateRouteResult{}, mapped
		}
		return DeactivateRouteResult{}, fmt.Errorf("route_admin: deactivate route: %w", mapped)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, fmt.Errorf("route_admin: deactivate route rows affected: %w", err)
	}
	if rows == 0 {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, repository.ErrStaleRevision
	}

	payload, err := json.Marshal(map[string]any{
		"route_id": in.RouteID,
		"active":   false,
	})
	if err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, fmt.Errorf("route_admin: marshal event payload: %w", err)
	}
	if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
		TenantID:     in.TenantID,
		EventType:    "route.config.deactivated",
		ActorUserID:  in.ActorUserID,
		ResourceType: "approval_route",
		ResourceID:   in.RouteID,
		PayloadJSON:  payload,
		OccurredAt:   s.clock.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, fmt.Errorf("route_admin: emit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return DeactivateRouteResult{}, fmt.Errorf("route_admin: commit: %w", err)
	}
	return DeactivateRouteResult{RouteID: in.RouteID}, nil
}

type lockedRouteState struct {
	ID      string
	Version int
	Active  bool
}

func lockRouteForUpdate(ctx context.Context, tx *sql.Tx, tenantID, routeID string) (lockedRouteState, error) {
	var state lockedRouteState
	err := tx.QueryRowContext(ctx, `
		SELECT id, version, active
		  FROM approval_routes
		 WHERE id = $1
		   AND tenant_id = $2
		 FOR UPDATE`,
		routeID, tenantID,
	).Scan(&state.ID, &state.Version, &state.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return lockedRouteState{}, ErrRouteNotFound
		}
		return lockedRouteState{}, fmt.Errorf("route_admin: lock route: %w", err)
	}
	return state, nil
}

func insertRouteStages(ctx context.Context, tx *sql.Tx, routeID string, stages []domain.Stage) error {
	for _, st := range stages {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO approval_route_stages
				(route_id, stage_order, name, required_role, required_capability, area_code, quorum, quorum_m, on_eligibility_drift)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			routeID,
			st.Order,
			st.Name,
			st.RequiredRole,
			st.RequiredCapability,
			st.AreaCode,
			st.Quorum,
			st.QuorumM,
			st.OnEligibilityDrift,
		); err != nil {
			return fmt.Errorf("route_admin: insert stage %d: %w", st.Order, repository.MapPgError(err, repository.MapHints{}))
		}
	}
	return nil
}
