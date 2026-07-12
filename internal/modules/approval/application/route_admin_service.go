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

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

// routeProfileFKConstraint is the FK on approval_routes (tenant_id, profile_code)
// → metaldocs.document_profiles. A violation of this specific constraint means
// the profile is not registered for the tenant: report it as a validation error.
// Any other FK violation must fall through to a 500 with diagnostic context.
const routeProfileFKConstraint = "approval_routes_document_profile_fk"

// RouteAdminService manages approval route configuration changes.
type RouteAdminService struct {
	repo         infrastructure.ApprovalRepository
	emitter      EventEmitter
	clock        Clock
	idempStore   RouteAdminIdempStore
	policyReader ProfilePolicyReader // G1: resolves the profile route-signature policy off-tx; nil ⇒ friendly check skipped (DB trigger is the last line)
}

// WithPolicyReader returns a copy of the service wired with the G1
// profile-policy reader. Passing nil leaves the friendly route-shape check
// disabled (the DB deferrable trigger remains the authoritative last line).
func (s *RouteAdminService) WithPolicyReader(reader ProfilePolicyReader) *RouteAdminService {
	if s == nil {
		return nil
	}
	cp := *s
	cp.policyReader = reader
	return &cp
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

// ErrRouteProfileUnknown is returned when the route's profile_code has no
// matching document profile for the tenant (FK violation on create). Surfaced
// as an actionable 4xx instead of an opaque 500.
var ErrRouteProfileUnknown = errors.New("route_admin: profile_code not registered for tenant")

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
	Routes []infrastructure.Route
}

// Create creates a new approval route and all route stages.
func (s *RouteAdminService) Create(ctx context.Context, runner db.TxRunner, in CreateRouteInput) (CreateRouteResult, error) {
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

	result, err := s.createTx(ctx, runner, in)
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

// resolvePolicy returns the profile's route-signature policy via the G1 reader,
// resolved OFF any write tx (H-PRE-1: the reader records CapTaxonomyView in
// taxonomy's own short tx, so it must never run inside a lock-holding approval
// tx). A nil reader yields "" (friendly check skipped; the DB deferrable trigger
// stays authoritative). A read error propagates so a genuine lookup failure
// surfaces rather than being silently downgraded (no-fallback principle).
func (s *RouteAdminService) resolvePolicy(ctx context.Context, tenantID, profileCode string) (taxonomydomain.RoutePolicy, error) {
	if s.policyReader == nil {
		return "", nil
	}
	return s.policyReader.RoutePolicy(ctx, tenantID, profileCode)
}

func (s *RouteAdminService) createTx(ctx context.Context, runner db.TxRunner, in CreateRouteInput) (CreateRouteResult, error) {
	var result CreateRouteResult
	// G1: resolve the per-profile route-signature policy OFF-TX (H-PRE-1) so the
	// friendly route-shape check runs before the write tx opens.
	policy, err := s.resolvePolicy(ctx, in.TenantID, in.ProfileCode)
	if err != nil {
		return result, err
	}
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		route := domain.Route{
			TenantID:    in.TenantID,
			ProfileCode: in.ProfileCode,
			Subject:     domain.NewDocumentSubject(in.ProfileCode),
			Version:     1,
			Stages:      in.Stages,
		}
		if err := route.Validate(policy); err != nil {
			return err
		}

		var routeID string
		err := tx.QueryRowContext(ctx, `
			INSERT INTO approval_routes
				(tenant_id, profile_code, name, version, created_by, active, subject_kind, subject_key)
			VALUES ($1, $2, $3, 1, $4, TRUE, $5, $6)
			RETURNING id`,
			in.TenantID, in.ProfileCode, in.Name, in.ActorUserID,
			string(route.Subject.Kind), route.Subject.Key,
		).Scan(&routeID)
		if err != nil {
			mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
			// Only a 23503 violation of the profile FK means the profile is not
			// registered for the tenant — map that to a validation error. Any other
			// FK violation falls through to a wrapped error so WriteError logs it at 500.
			if errors.Is(mapped, infrastructure.ErrFKViolation) {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.ConstraintName == routeProfileFKConstraint {
					return ErrRouteProfileUnknown
				}
			}
			return fmt.Errorf("insert route: %w", mapped)
		}

		if err := insertRouteStages(ctx, tx, routeID, in.Stages); err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"route_id":      routeID,
			"profile_code":  in.ProfileCode,
			"stage_count":   len(in.Stages),
			"initial_state": "active",
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeRouteConfigCreated,
			ActorUserID:  in.ActorUserID,
			ResourceType: "approval_route",
			ResourceID:   routeID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}

		result = CreateRouteResult{RouteID: routeID}
		return nil
	})
	if err != nil {
		return CreateRouteResult{}, err
	}
	return result, nil
}

// Update updates route metadata and replaces all route stages atomically.
func (s *RouteAdminService) Update(ctx context.Context, runner db.TxRunner, in UpdateRouteInput) (UpdateRouteResult, error) {
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

	result, err := s.updateTx(ctx, runner, in)
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

// resolveUpdatePolicy resolves the route's profile route-signature policy OFF-TX
// for an update (H-PRE-1). UpdateRouteInput carries no profile_code, so it first
// reads the route's (immutable) profile_code in a short non-recording tx, then
// resolves the policy via the G1 reader. The profile_code read is scoped to
// ACTIVE routes to mirror the DB direction-A trigger, which only enforces on
// active routes: an inactive/absent route yields "" (friendly check skipped),
// keeping the friendly and authoritative layers identical in scope while the
// write tx below surfaces any real not-found/stale error. Reading profile_code
// off-tx is TOCTOU-safe because a route's profile_code is immutable across its
// life (supersede preserves it).
func (s *RouteAdminService) resolveUpdatePolicy(ctx context.Context, runner db.TxRunner, tenantID, routeID string) (taxonomydomain.RoutePolicy, error) {
	if s.policyReader == nil {
		return "", nil
	}
	var (
		profileCode string
		found       bool
	)
	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		code, ok, err := loadActiveRouteProfileCode(ctx, tx, tenantID, routeID)
		profileCode, found = code, ok
		return err
	}); err != nil {
		return "", err
	}
	if !found {
		return "", nil
	}
	return s.policyReader.RoutePolicy(ctx, tenantID, profileCode)
}

func (s *RouteAdminService) updateTx(ctx context.Context, runner db.TxRunner, in UpdateRouteInput) (UpdateRouteResult, error) {
	var result UpdateRouteResult
	// G1: resolve the route's profile policy OFF-TX (H-PRE-1) before the write tx.
	policy, err := s.resolveUpdatePolicy(ctx, runner, in.TenantID, in.RouteID)
	if err != nil {
		return result, err
	}
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		locked, err := lockRouteForUpdate(ctx, tx, in.TenantID, in.RouteID)
		if err != nil {
			return err
		}

		currentStages, err := loadRouteStagesTx(ctx, tx, in.RouteID)
		if err != nil {
			return err
		}
		stagesChanged := !stagesEqual(currentStages, in.Stages)

		route := domain.Route{
			ID:       in.RouteID,
			TenantID: in.TenantID,
			Stages:   in.Stages,
		}
		if err := route.Validate(policy); err != nil {
			return err
		}

		if in.ExpectedVersion != 0 && locked.Version != in.ExpectedVersion {
			return infrastructure.ErrStaleRevision
		}

		newRouteID, newVersion, err := s.updateInPlaceOrSupersede(ctx, tx, in, locked, stagesChanged)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"route_id":    newRouteID,
			"new_version": newVersion,
			"stage_count": len(in.Stages),
			"superseded":  newRouteID != in.RouteID,
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeRouteConfigUpdated,
			ActorUserID:  in.ActorUserID,
			ResourceType: "approval_route",
			ResourceID:   newRouteID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}

		result = UpdateRouteResult{RouteID: newRouteID, NewVersion: newVersion}
		return nil
	})
	if err != nil {
		return UpdateRouteResult{}, err
	}
	return result, nil
}

// updateInPlaceOrSupersede tries the cheap in-place mutation first (matches
// pre-F2 behavior for a route never referenced by any instance). If the
// enforce_route_immutable trigger (migration 0287) rejects it with
// ErrRouteInUse, it falls back to the versioned-supersede path: the current
// row is retired (active=false, superseded_at=now() — the ONLY columns the
// trigger still allows to change on an in-use row) and a brand-new row is
// inserted carrying the incremented version and the new definition. Both
// branches run inside the caller's single transaction.
func (s *RouteAdminService) updateInPlaceOrSupersede(ctx context.Context, tx *sql.Tx, in UpdateRouteInput, locked lockedRouteState, stagesChanged bool) (routeID string, newVersion int, err error) {
	// The in-place UPDATE below is speculative: enforce_route_immutable() may
	// reject it with ErrRouteInUse (P0001), which — per Postgres transaction
	// semantics — aborts the entire enclosing transaction, not just the failed
	// statement. Without a SAVEPOINT here, every subsequent statement in this
	// tx (including the supersede fallback three lines below) fails closed
	// with SQLSTATE 25P02 ("current transaction is aborted"), masking the real
	// ErrRouteInUse behind an opaque 500. F10 live-QA caught this: a fresh
	// in-use route's first-ever versioned update 500'd instead of superseding.
	const savepoint = "route_update_attempt"
	if _, spErr := tx.ExecContext(ctx, "SAVEPOINT "+savepoint); spErr != nil {
		return "", 0, fmt.Errorf("savepoint before speculative update: %w", spErr)
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE approval_routes
		   SET name = $1,
		       version = version + 1
		 WHERE id = $2
		   AND tenant_id = $3
		RETURNING version`,
		in.Name, in.RouteID, in.TenantID,
	).Scan(&newVersion)
	if err == nil {
		if _, relErr := tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepoint); relErr != nil {
			return "", 0, fmt.Errorf("release savepoint after in-place update: %w", relErr)
		}
		if stagesChanged {
			if _, err := tx.ExecContext(ctx, `DELETE FROM approval_route_stages WHERE route_id = $1`, in.RouteID); err != nil {
				return "", 0, fmt.Errorf("delete stages: %w", err)
			}
			if err := insertRouteStages(ctx, tx, in.RouteID, in.Stages); err != nil {
				return "", 0, err
			}
		}
		return in.RouteID, newVersion, nil
	}

	mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
	if !errors.Is(mapped, infrastructure.ErrRouteInUse) {
		return "", 0, fmt.Errorf("update route: %w", mapped)
	}

	// Roll back to the savepoint to un-abort the transaction before issuing
	// any further statements — the failed in-place UPDATE must not poison the
	// supersede fallback that follows.
	if _, rbErr := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint); rbErr != nil {
		return "", 0, fmt.Errorf("rollback to savepoint after ErrRouteInUse: %w", rbErr)
	}

	// In use: retire the old row (trigger permits an active/superseded_at-only
	// UPDATE on an in-use row), then insert the new version as a new row. Order
	// matters — the partial unique index on (tenant_id, profile_code) WHERE
	// active would reject the new row's INSERT if the old row were still
	// active at that point.
	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_routes
		   SET active = FALSE,
		       superseded_at = $1
		 WHERE id = $2
		   AND tenant_id = $3`,
		s.clock.Now(), in.RouteID, in.TenantID,
	); err != nil {
		return "", 0, fmt.Errorf("supersede route: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}

	newVersion = locked.Version + 1
	err = tx.QueryRowContext(ctx, `
		INSERT INTO approval_routes
			(tenant_id, profile_code, name, version, created_by, active, subject_kind, subject_key)
		SELECT tenant_id, profile_code, $1, $2, $3, TRUE, subject_kind, subject_key
		  FROM approval_routes
		 WHERE id = $4
		RETURNING id`,
		in.Name, newVersion, in.ActorUserID, in.RouteID,
	).Scan(&routeID)
	if err != nil {
		return "", 0, fmt.Errorf("insert superseding route version: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}

	if err := insertRouteStages(ctx, tx, routeID, in.Stages); err != nil {
		return "", 0, err
	}

	return routeID, newVersion, nil
}

// Deactivate marks a route inactive.
func (s *RouteAdminService) Deactivate(ctx context.Context, runner db.TxRunner, in DeactivateRouteInput) (DeactivateRouteResult, error) {
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

	result, err := s.deactivateTx(ctx, runner, in)
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

func (s *RouteAdminService) deactivateTx(ctx context.Context, runner db.TxRunner, in DeactivateRouteInput) (DeactivateRouteResult, error) {
	var result DeactivateRouteResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		lockedRoute, err := lockRouteForUpdate(ctx, tx, in.TenantID, in.RouteID)
		if err != nil {
			return err
		}
		if !lockedRoute.Active {
			if in.ExpectedVersion > 0 && lockedRoute.Version != in.ExpectedVersion {
				return infrastructure.ErrStaleRevision
			}
			return ErrRouteAlreadyInactive
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
			mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
			if errors.Is(mapped, infrastructure.ErrRouteInUse) {
				return mapped
			}
			return fmt.Errorf("deactivate route: %w", mapped)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("deactivate route rows affected: %w", err)
		}
		if rows == 0 {
			return infrastructure.ErrStaleRevision
		}

		payload, err := json.Marshal(map[string]any{
			"route_id": in.RouteID,
			"active":   false,
			"reason":   in.Reason,
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeRouteConfigDeactivated,
			ActorUserID:  in.ActorUserID,
			ResourceType: "approval_route",
			ResourceID:   in.RouteID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}

		result = DeactivateRouteResult{RouteID: in.RouteID}
		return nil
	})
	if err != nil {
		return DeactivateRouteResult{}, err
	}
	return result, nil
}

// List loads tenant-scoped routes inside a single transaction that owns the
// tenant/actor GUCs and the route.admin authz check. This closes the TOCTOU
// gap where the prior handler ran authz in one tx and the list query in another.
func (s *RouteAdminService) List(ctx context.Context, runner db.TxRunner, tenantID, actorID string) (ListRoutesResult, error) {
	const op = "list"

	var result ListRoutesResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		routes, err := s.repo.ListRoutesTx(ctx, tx, tenantID)
		if err != nil {
			return err
		}

		result = ListRoutesResult{Routes: routes}
		return nil
	})
	if err != nil {
		return ListRoutesResult{}, wrapRouteAdminErr(op, err)
	}
	return result, nil
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
		errors.Is(err, ErrRouteProfileUnknown),
		errors.Is(err, infrastructure.ErrStaleRevision),
		errors.Is(err, infrastructure.ErrRouteInUse),
		errors.Is(err, infrastructure.ErrDuplicateRouteProfile):
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

// loadActiveRouteProfileCode reads the immutable profile_code of an ACTIVE route
// with a plain non-recording SELECT (safe in any tx; no authz recording, so it
// never trips H-PRE-1). Returns found=false when the route is inactive or absent
// so the G1 friendly check can be skipped in lockstep with the DB trigger's
// active-only scope.
func loadActiveRouteProfileCode(ctx context.Context, tx *sql.Tx, tenantID, routeID string) (string, bool, error) {
	var code string
	err := tx.QueryRowContext(ctx, `
		SELECT profile_code
		  FROM approval_routes
		 WHERE id = $1
		   AND tenant_id = $2
		   AND active = TRUE`,
		routeID, tenantID,
	).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("route_admin: load route profile_code: %w", err)
	}
	return code, true, nil
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
	const colsPerRow = 11
	placeholders := make([]string, 0, len(stages))
	args := make([]any, 0, len(stages)*colsPerRow)
	for i, st := range stages {
		base := i*colsPerRow + 1
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base, base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10,
		))
		// stage_kind is NOT NULL DEFAULT 'approval'; an empty domain zero-value
		// must not clobber the DB default with an empty string (F1 migration 0286).
		kind := st.Kind
		if kind == "" {
			kind = domain.StageKindApproval
		}
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
			string(kind),
			st.DueInDays,
		)
	}
	query := `
		INSERT INTO approval_route_stages
			(route_id, stage_order, name, required_role, required_capability, area_code, quorum, quorum_m, on_eligibility_drift, stage_kind, due_in_days)
		VALUES ` + strings.Join(placeholders, ",")
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("route_admin: insert stages: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}
	return nil
}

func loadRouteStagesTx(ctx context.Context, tx *sql.Tx, routeID string) ([]domain.Stage, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT stage_order, name, required_role, required_capability, area_code, quorum, quorum_m, on_eligibility_drift, stage_kind, due_in_days
		  FROM approval_route_stages
		 WHERE route_id = $1
		 ORDER BY stage_order ASC`,
		routeID,
	)
	if err != nil {
		return nil, fmt.Errorf("route_admin: load stages: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}
	defer rows.Close()

	var out []domain.Stage
	for rows.Next() {
		var (
			st        domain.Stage
			quorumM   sql.NullInt64
			kindStr   string
			dueInDays sql.NullInt64
		)
		if err := rows.Scan(&st.Order, &st.Name, &st.RequiredRole, &st.RequiredCapability, &st.AreaCode, &st.Quorum, &quorumM, &st.OnEligibilityDrift, &kindStr, &dueInDays); err != nil {
			return nil, fmt.Errorf("route_admin: scan stage: %w", err)
		}
		if quorumM.Valid {
			m := int(quorumM.Int64)
			st.QuorumM = &m
		}
		st.Kind = domain.StageKind(kindStr)
		if dueInDays.Valid {
			d := int(dueInDays.Int64)
			st.DueInDays = &d
		}
		out = append(out, st)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("route_admin: iterate stages: %w", err)
	}
	return out, nil
}

// stagesEqual compares stage sets for the update-diff decision. Kind is
// normalized to its DB default (StageKindApproval) before comparing so a
// caller that omits Kind (its Go zero value, "") is never spuriously
// considered "changed" against a stage loaded back from the DB, which always
// reads a concrete value thanks to the NOT NULL DEFAULT (migration 0286). No
// other field changed shape, so DeepEqual remains the comparison for those.
func stagesEqual(a, b []domain.Stage) bool {
	norm := func(in []domain.Stage) []domain.Stage {
		out := make([]domain.Stage, len(in))
		copy(out, in)
		for i := range out {
			if out[i].Kind == "" {
				out[i].Kind = domain.StageKindApproval
			}
		}
		return out
	}
	return reflect.DeepEqual(norm(a), norm(b))
}
