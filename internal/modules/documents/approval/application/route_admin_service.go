package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
)

// RouteAdminService manages approval route configuration changes.
type RouteAdminService struct {
	repo       repository.ApprovalRepository
	emitter    EventEmitter
	clock      Clock
	idempStore RouteAdminIdempStore
}

// WithIdempStore returns a copy of the service wired with an idempotency store.
// Passing nil leaves replay disabled (useful for unit tests that exercise the
// service without the Postgres store).
func (s *RouteAdminService) WithIdempStore(store RouteAdminIdempStore) *RouteAdminService {
	if s == nil {
		return nil
	}
	cp := *s
	cp.idempStore = store
	return &cp
}

// ErrRouteNotFound is returned when a route does not exist for the tenant.
var ErrRouteNotFound = errors.New("route_admin: route not found")

// ErrRouteAlreadyInactive is returned when the route is already inactive.
var ErrRouteAlreadyInactive = errors.New("route_admin: route already inactive")

// ErrRouteDeactivateReasonRequired is returned when DeactivateRouteInput.Reason
// is empty or whitespace-only.
var ErrRouteDeactivateReasonRequired = errors.New("route_admin: deactivate reason must not be empty")

// CreateRouteInput carries all inputs for Create.
type CreateRouteInput struct {
	TenantID       string
	ProfileCode    string
	Name           string
	ActorUserID    string
	IdempotencyKey string
	Stages         []domain.Stage
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
	IdempotencyKey  string
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
	IdempotencyKey  string
	Reason          string
	ExpectedVersion int
}

// DeactivateRouteResult is returned on successful route deactivation.
type DeactivateRouteResult struct {
	RouteID string
}

// ListRoutesResult is returned by List.
type ListRoutesResult struct {
	Routes []repository.Route
}

// Create creates a new approval route and all route stages.
func (s *RouteAdminService) Create(ctx context.Context, db *sql.DB, in CreateRouteInput) (CreateRouteResult, error) {
	const op = "create"

	var (
		committer RouteAdminReplayCommitter
		replay    *RouteAdminReplay
	)
	if s.idempStore != nil && in.IdempotencyKey != "" {
		hash := computeCreateRoutePayloadHash(in.ProfileCode, in.Name, in.Stages)
		var err error
		committer, replay, err = s.idempStore.BeginCreateReplay(ctx, in.TenantID, in.ActorUserID, in.IdempotencyKey, hash)
		if err != nil {
			return CreateRouteResult{}, err
		}
		if replay != nil {
			return CreateRouteResult{RouteID: replay.RouteID}, nil
		}
	}

	result, err := s.createTx(ctx, db, in)
	if err != nil {
		if committer != nil {
			if ferr := committer.Fail(err); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", err,
					"fail_err", ferr,
				)
			}
		}
		return CreateRouteResult{}, wrapRouteAdminErr(op, err)
	}
	if committer != nil {
		if cerr := committer.Complete(result.RouteID, nil); cerr != nil {
			return result, wrapRouteAdminErr(op, fmt.Errorf("complete replay: %w", cerr))
		}
	}
	return result, nil
}

func (s *RouteAdminService) createTx(ctx context.Context, db *sql.DB, in CreateRouteInput) (CreateRouteResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return CreateRouteResult{}, fmt.Errorf("begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
		_ = tx.Rollback()
		return CreateRouteResult{}, err
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
		return CreateRouteResult{}, fmt.Errorf("insert route: %w", repository.MapPgError(err, repository.MapHints{}))
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
		return CreateRouteResult{}, fmt.Errorf("marshal event payload: %w", err)
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
		return CreateRouteResult{}, fmt.Errorf("emit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CreateRouteResult{}, fmt.Errorf("commit: %w", err)
	}
	return CreateRouteResult{RouteID: routeID}, nil
}

// Update updates route metadata and replaces all route stages atomically.
func (s *RouteAdminService) Update(ctx context.Context, db *sql.DB, in UpdateRouteInput) (UpdateRouteResult, error) {
	const op = "update"

	var (
		committer RouteAdminReplayCommitter
		replay    *RouteAdminReplay
	)
	if s.idempStore != nil && in.IdempotencyKey != "" {
		hash := computeUpdateRoutePayloadHash(in.RouteID, in.ExpectedVersion, in.Name, in.Stages)
		var err error
		committer, replay, err = s.idempStore.BeginUpdateReplay(ctx, in.TenantID, in.ActorUserID, in.IdempotencyKey, hash)
		if err != nil {
			return UpdateRouteResult{}, err
		}
		if replay != nil {
			res := UpdateRouteResult{RouteID: replay.RouteID}
			if replay.NewVersion != nil {
				res.NewVersion = *replay.NewVersion
			}
			return res, nil
		}
	}

	result, err := s.updateTx(ctx, db, in)
	if err != nil {
		if committer != nil {
			if ferr := committer.Fail(err); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", err,
					"fail_err", ferr,
				)
			}
		}
		return UpdateRouteResult{}, wrapRouteAdminErr(op, err)
	}
	if committer != nil {
		nv := result.NewVersion
		if cerr := committer.Complete(result.RouteID, &nv); cerr != nil {
			return result, wrapRouteAdminErr(op, fmt.Errorf("complete replay: %w", cerr))
		}
	}
	return result, nil
}

func (s *RouteAdminService) updateTx(ctx context.Context, db *sql.DB, in UpdateRouteInput) (UpdateRouteResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return UpdateRouteResult{}, fmt.Errorf("begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, err
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

	currentStages, err := loadRouteStagesTx(ctx, tx, in.RouteID)
	if err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, err
	}
	stagesChanged := !stagesEqual(currentStages, in.Stages)

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
		return UpdateRouteResult{}, fmt.Errorf("update route: %w", mapped)
	}

	if stagesChanged {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM approval_route_stages
			WHERE route_id = $1`,
			in.RouteID,
		); err != nil {
			_ = tx.Rollback()
			return UpdateRouteResult{}, fmt.Errorf("delete stages: %w", err)
		}

		if err := insertRouteStages(ctx, tx, in.RouteID, in.Stages); err != nil {
			_ = tx.Rollback()
			return UpdateRouteResult{}, err
		}
	}

	payload, err := json.Marshal(map[string]any{
		"route_id":    in.RouteID,
		"new_version": newVersion,
		"stage_count": len(in.Stages),
	})
	if err != nil {
		_ = tx.Rollback()
		return UpdateRouteResult{}, fmt.Errorf("marshal event payload: %w", err)
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
		return UpdateRouteResult{}, fmt.Errorf("emit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return UpdateRouteResult{}, fmt.Errorf("commit: %w", err)
	}
	return UpdateRouteResult{RouteID: in.RouteID, NewVersion: newVersion}, nil
}

// Deactivate marks a route inactive.
func (s *RouteAdminService) Deactivate(ctx context.Context, db *sql.DB, in DeactivateRouteInput) (DeactivateRouteResult, error) {
	const op = "deactivate"

	reason := strings.TrimSpace(in.Reason)
	in.Reason = reason

	var (
		committer RouteAdminReplayCommitter
		replay    *RouteAdminReplay
	)
	if s.idempStore != nil && in.IdempotencyKey != "" {
		hash := computeDeactivateRoutePayloadHash(in.RouteID, in.ExpectedVersion, in.Reason)
		var err error
		committer, replay, err = s.idempStore.BeginDeactivateReplay(ctx, in.TenantID, in.ActorUserID, in.IdempotencyKey, hash)
		if err != nil {
			return DeactivateRouteResult{}, err
		}
		if replay != nil {
			return DeactivateRouteResult{RouteID: replay.RouteID}, nil
		}
	}

	if reason == "" {
		if committer != nil {
			if ferr := committer.Fail(ErrRouteDeactivateReasonRequired); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", ErrRouteDeactivateReasonRequired,
					"fail_err", ferr,
				)
			}
		}
		return DeactivateRouteResult{}, ErrRouteDeactivateReasonRequired
	}

	result, err := s.deactivateTx(ctx, db, in)
	if err != nil {
		if committer != nil {
			if ferr := committer.Fail(err); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", err,
					"fail_err", ferr,
				)
			}
		}
		return DeactivateRouteResult{}, wrapRouteAdminErr(op, err)
	}
	if committer != nil {
		if cerr := committer.Complete(result.RouteID, nil); cerr != nil {
			return result, wrapRouteAdminErr(op, fmt.Errorf("complete replay: %w", cerr))
		}
	}
	return result, nil
}

func (s *RouteAdminService) deactivateTx(ctx context.Context, db *sql.DB, in DeactivateRouteInput) (DeactivateRouteResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return DeactivateRouteResult{}, fmt.Errorf("begin tx: %w", err)
	}

	if err := setAuthzGUC(ctx, tx, in.TenantID, in.ActorUserID); err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, err
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
		return DeactivateRouteResult{}, fmt.Errorf("deactivate route: %w", mapped)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, fmt.Errorf("deactivate route rows affected: %w", err)
	}
	if rows == 0 {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, repository.ErrStaleRevision
	}

	payload, err := json.Marshal(map[string]any{
		"route_id": in.RouteID,
		"active":   false,
		"reason":   in.Reason,
	})
	if err != nil {
		_ = tx.Rollback()
		return DeactivateRouteResult{}, fmt.Errorf("marshal event payload: %w", err)
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
		return DeactivateRouteResult{}, fmt.Errorf("emit event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return DeactivateRouteResult{}, fmt.Errorf("commit: %w", err)
	}
	return DeactivateRouteResult{RouteID: in.RouteID}, nil
}

// List loads tenant-scoped routes inside a single transaction that owns the
// tenant/actor GUCs and the route.admin authz check. This closes the TOCTOU
// gap where the prior handler ran authz in one tx and the list query in another.
func (s *RouteAdminService) List(ctx context.Context, db *sql.DB, tenantID, actorID string) (ListRoutesResult, error) {
	const op = "list"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ListRoutesResult{}, wrapRouteAdminErr(op, fmt.Errorf("begin tx: %w", err))
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
		return ListRoutesResult{}, wrapRouteAdminErr(op, err)
	}
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "route.admin", "tenant"); err != nil {
		return ListRoutesResult{}, err
	}

	routes, err := s.repo.ListRoutesTx(ctx, tx, tenantID)
	if err != nil {
		return ListRoutesResult{}, wrapRouteAdminErr(op, err)
	}
	if err := tx.Commit(); err != nil {
		return ListRoutesResult{}, wrapRouteAdminErr(op, fmt.Errorf("commit: %w", err))
	}
	return ListRoutesResult{Routes: routes}, nil
}

// wrapRouteAdminErr standardises the error envelope. Sentinels and mapped
// repository errors pass through unchanged so handler-level errors.Is checks
// keep working; everything else is annotated with the op name.
func wrapRouteAdminErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if isPassThroughRouteAdminErr(err) {
		return err
	}
	return fmt.Errorf("route_admin: %s: %w", op, err)
}

func isPassThroughRouteAdminErr(err error) bool {
	switch {
	case errors.Is(err, ErrRouteNotFound),
		errors.Is(err, ErrRouteAlreadyInactive),
		errors.Is(err, ErrRouteDeactivateReasonRequired),
		errors.Is(err, repository.ErrStaleRevision),
		errors.Is(err, repository.ErrRouteInUse),
		errors.Is(err, repository.ErrDuplicateRouteProfile):
		return true
	}
	var capDenied authz.ErrCapDenied
	return errors.As(err, &capDenied)
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
	if len(stages) == 0 {
		return nil
	}
	const colsPerRow = 9
	placeholders := make([]string, 0, len(stages))
	args := make([]any, 0, len(stages)*colsPerRow)
	for i, st := range stages {
		base := i*colsPerRow + 1
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base, base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8,
		))
		args = append(args,
			routeID,
			st.Order,
			st.Name,
			st.RequiredRole,
			st.RequiredCapability,
			st.AreaCode,
			st.Quorum,
			st.QuorumM,
			st.OnEligibilityDrift,
		)
	}
	query := `
		INSERT INTO approval_route_stages
			(route_id, stage_order, name, required_role, required_capability, area_code, quorum, quorum_m, on_eligibility_drift)
		VALUES ` + strings.Join(placeholders, ",")
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("route_admin: insert stages: %w", repository.MapPgError(err, repository.MapHints{}))
	}
	return nil
}

func loadRouteStagesTx(ctx context.Context, tx *sql.Tx, routeID string) ([]domain.Stage, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT stage_order, name, required_role, required_capability, area_code, quorum, quorum_m, on_eligibility_drift
		  FROM approval_route_stages
		 WHERE route_id = $1
		 ORDER BY stage_order ASC`,
		routeID,
	)
	if err != nil {
		return nil, fmt.Errorf("route_admin: load stages: %w", repository.MapPgError(err, repository.MapHints{}))
	}
	defer rows.Close()

	var out []domain.Stage
	for rows.Next() {
		var (
			st      domain.Stage
			quorumM sql.NullInt64
		)
		if err := rows.Scan(&st.Order, &st.Name, &st.RequiredRole, &st.RequiredCapability, &st.AreaCode, &st.Quorum, &quorumM, &st.OnEligibilityDrift); err != nil {
			return nil, fmt.Errorf("route_admin: scan stage: %w", err)
		}
		if quorumM.Valid {
			m := int(quorumM.Int64)
			st.QuorumM = &m
		}
		out = append(out, st)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("route_admin: iterate stages: %w", err)
	}
	return out, nil
}

func stagesEqual(a, b []domain.Stage) bool {
	return reflect.DeepEqual(a, b)
}
