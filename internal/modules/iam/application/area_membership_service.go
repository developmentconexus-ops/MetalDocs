package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

var (
	// ErrMembershipNotFound is returned by Revoke when no active membership
	// exists for the given (user, tenant, area).
	ErrMembershipNotFound = errors.New("membership_not_found")
	// ErrUnknownRole is returned by Grant when role does not parse as a known
	// domain.Role.
	ErrUnknownRole = errors.New("unknown_role")
	// ErrMembershipExists is returned by Grant when an active membership for
	// (user, tenant, area) with the same role already exists. The admin
	// surface treats grant as non-idempotent: callers must explicitly revoke
	// before re-granting the same role; a role change still succeeds via
	// GrantAtomic (close-old + insert-new).
	ErrMembershipExists = errors.New("membership_exists")
)

// UserAreaWriteRepository is the persistence port for area-membership
// mutations and directory reads. All Tx-suffixed methods must run inside a
// transaction opened by BeginTx so the mutation and its governance log entry
// commit atomically (REQ-ASYNC-1).
type UserAreaWriteRepository interface {
	ListActive(ctx context.Context, userID, tenantID string, now time.Time) ([]domain.UserProcessArea, error)
	// ListActiveForUsers loads active memberships for a set of userIDs in a
	// single query, keyed by userID. Used by PeopleService.ListFiltered to
	// eliminate the per-user N+1 query (H-5.3 D2).
	ListActiveForUsers(ctx context.Context, tenantID string, userIDs []string, now time.Time) (map[string][]domain.UserProcessArea, error)
	ListByTenant(ctx context.Context, tenantID, userID, areaCode, role string, now time.Time) ([]domain.UserProcessArea, error)
	// MembershipDirectoryScope reports the actor's directory visibility (ADR 0022
	// Phase 4). tenantWide=true → full tenant directory (system_admin inheritance
	// bypass, R1). Otherwise hasManagedAreas=true → the actor holds the given
	// capability in at least one area (area_admin); both false → self-only.
	MembershipDirectoryScope(ctx context.Context, tenantID, actorID, capability string, now time.Time) (tenantWide bool, hasManagedAreas bool, err error)
	// ListByTenantInManagedAreas lists active memberships restricted to the areas
	// where the actor holds the given capability. The managed-area restriction is
	// applied IN SQL (ADR 0022 R3 — data-layer enforcement, not post-fetch).
	ListByTenantInManagedAreas(ctx context.Context, tenantID, userID, areaCode, role, actorID, capability string, now time.Time) ([]domain.UserProcessArea, error)
	GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*domain.UserProcessArea, error)
	// BeginTx opens a database transaction for in-tx mutation + governance writes
	// (REQ-ASYNC-1). The caller owns Commit/Rollback.
	BeginTx(ctx context.Context) (domain.MembershipTx, error)
	// InsertTx writes a new membership row inside an open transaction.
	InsertTx(ctx context.Context, tx domain.MembershipTx, membership domain.UserProcessArea) error
	// CloseActiveTx sets effective_to on the active row inside an open transaction.
	CloseActiveTx(ctx context.Context, tx domain.MembershipTx, userID, tenantID, areaCode string, effectiveTo time.Time, actorID string) error
	// GrantAtomicTx closes the old membership and inserts the new one inside an
	// open transaction (role-change path).
	GrantAtomicTx(ctx context.Context, tx domain.MembershipTx, oldMembership, newMembership domain.UserProcessArea) error
}

// MembershipGovernanceLogger records an area-membership grant/revoke event as
// part of the caller's transaction, so the audit trail can never diverge from
// the mutation it describes.
type MembershipGovernanceLogger interface {
	// LogTx writes the governance event inside an open transaction so the audit
	// record is atomically committed with the membership mutation (REQ-ASYNC-1,
	// T-007). A failure rolls back the mutation.
	LogTx(ctx context.Context, tx db.Tx, action string, membership domain.UserProcessArea) error
}

// AreaMembershipService owns grant/revoke of user↔area role memberships,
// pairing each mutation with an atomic governance-log write and (once wired)
// a role-cache invalidation so authorization reflects the change immediately.
type AreaMembershipService struct {
	repo        UserAreaWriteRepository
	logger      MembershipGovernanceLogger
	invalidator RoleCacheInvalidator
	nowFn       func() time.Time
}

// NewAreaMembershipService constructs an AreaMembershipService. repo and
// logger are required; it panics if either is nil since that is a wiring
// bug, not a runtime condition.
func NewAreaMembershipService(repo UserAreaWriteRepository, logger MembershipGovernanceLogger) *AreaMembershipService {
	if repo == nil {
		panic("iam.NewAreaMembershipService: repo is required")
	}
	if logger == nil {
		panic("iam.NewAreaMembershipService: logger is required")
	}
	return &AreaMembershipService{
		repo:   repo,
		logger: logger,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// WithRoleCacheInvalidator wires the role-cache invalidator so a grant or revoke
// flushes the actor's cached roles immediately, closing the window where a
// changed area membership keeps authorizing until the cache TTL expires (A3).
// Mirrors the builder style of authapp.Service.WithCapabilityProvider. nil is a
// no-op, so callers without a cache (e.g. memory dev mode) are unaffected.
func (s *AreaMembershipService) WithRoleCacheInvalidator(inv RoleCacheInvalidator) *AreaMembershipService {
	s.invalidator = inv
	return s
}

func (s *AreaMembershipService) invalidate(userID, tenantID string) {
	if s.invalidator != nil {
		s.invalidator.InvalidateUserTenant(userID, tenantID)
	}
}

// ListActive returns the caller's active area memberships as of now.
func (s *AreaMembershipService) ListActive(ctx context.Context, userID, tenantID string) ([]domain.UserProcessArea, error) {
	return s.repo.ListActive(ctx, userID, tenantID, s.nowFn())
}

// ListActiveBatch loads active memberships for a set of users in a single
// query. Returns a map of userID → []UserProcessArea. Used by
// PeopleService.ListFiltered to eliminate the per-user N+1 (H-5.3 D2).
func (s *AreaMembershipService) ListActiveBatch(ctx context.Context, tenantID string, userIDs []string) (map[string][]domain.UserProcessArea, error) {
	return s.repo.ListActiveForUsers(ctx, tenantID, userIDs, s.nowFn())
}

// ListByTenant returns active memberships across the tenant with optional
// exact-match filters (empty string = no filter). Tenant-wide directory mode
// for the admin surface; authz scoping (who may see the whole tenant vs only
// their own rows) is enforced by the HTTP handler.
func (s *AreaMembershipService) ListByTenant(ctx context.Context, tenantID, userID, areaCode, role string) ([]domain.UserProcessArea, error) {
	return s.repo.ListByTenant(ctx, tenantID, userID, areaCode, role, s.nowFn())
}

// DirectoryScope resolves the actor's membership-directory visibility (ADR 0022
// Phase 4). See UserAreaWriteRepository.MembershipDirectoryScope.
func (s *AreaMembershipService) DirectoryScope(ctx context.Context, tenantID, actorID, capability string) (tenantWide bool, hasManagedAreas bool, err error) {
	return s.repo.MembershipDirectoryScope(ctx, tenantID, actorID, capability, s.nowFn())
}

// ListByTenantInManagedAreas lists active memberships scoped to the actor's
// managed areas (areas where the actor holds capability), filtered in SQL.
func (s *AreaMembershipService) ListByTenantInManagedAreas(ctx context.Context, tenantID, userID, areaCode, role, actorID, capability string) ([]domain.UserProcessArea, error) {
	return s.repo.ListByTenantInManagedAreas(ctx, tenantID, userID, areaCode, role, actorID, capability, s.nowFn())
}

// Grant assigns role to userID in areaCode. If an active membership with a
// different role already exists it is atomically closed and replaced
// (role-change path); if the same role is already active it returns
// ErrMembershipExists (grant is non-idempotent — callers must Revoke first).
// The mutation and its governance-log entry commit in the same transaction,
// then the actor's cached roles are invalidated (A3).
func (s *AreaMembershipService) Grant(
	ctx context.Context,
	userID, tenantID, areaCode string,
	role domain.Role,
	grantedBy string,
) error {
	// F-QA4-2: an area membership is a public.user_process_areas row, and that
	// table's role CHECK accepts exactly the seven area-assignable roles. The
	// former ParseRole check admitted all EIGHT canonical roles, so a
	// `system_admin` grant passed the app layer and was then rejected by the DB
	// as 23514 — a 500 for what is a plain vocabulary violation. Binding the
	// check to IsAreaRole makes the app layer the friendly first line over the
	// same invariant the DB enforces, and matches the AreaRole enum the contract
	// now declares for GrantAreaMembershipRequest.role.
	if !domain.IsAreaRole(role) {
		return ErrUnknownRole
	}

	now := s.nowFn()
	existing, err := s.repo.GetActiveByUserAndArea(ctx, userID, tenantID, areaCode, now)
	if err != nil {
		return fmt.Errorf("get active membership: %w", err)
	}
	membership := buildMembership(userID, tenantID, areaCode, role, now, grantedBy)
	if existing != nil && existing.IsActive(now) {
		if existing.Role == role {
			return ErrMembershipExists
		}
		if err := s.commitMembershipMutation(ctx, "grant", membership, func(ctx context.Context, tx domain.MembershipTx) error {
			if err := s.repo.GrantAtomicTx(ctx, tx, *existing, membership); err != nil {
				return fmt.Errorf("grant membership atomically: %w", err)
			}
			return nil
		}); err != nil {
			return err
		}
		// Flush the actor's cached roles once the mutation and governance are
		// committed atomically (A3).
		s.invalidate(userID, tenantID)
		return nil
	}

	if err := s.commitMembershipMutation(ctx, "insert", membership, func(ctx context.Context, tx domain.MembershipTx) error {
		if err := s.repo.InsertTx(ctx, tx, membership); err != nil {
			return fmt.Errorf("insert membership: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	// Flush the actor's cached roles once the mutation and governance are
	// committed atomically (A3).
	s.invalidate(userID, tenantID)
	return nil
}

// commitMembershipMutation opens a transaction, runs mutate to perform the
// membership row change, writes the "role.grant" governance-log entry for the
// same mutation, and commits — the shared atomic-commit envelope
// (REQ-ASYNC-1) used by both of Grant's branches (role-change vs first
// grant). txLabel names the tx in the begin/commit wrap error only ("grant"
// or "insert"); the governance action logged is always role.grant, matching
// the pre-extraction behavior of both branches.
func (s *AreaMembershipService) commitMembershipMutation(ctx context.Context, txLabel string, membership domain.UserProcessArea, mutate func(ctx context.Context, tx domain.MembershipTx) error) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin %s tx: %w", txLabel, err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := mutate(ctx, tx); err != nil {
		return err
	}
	if err := s.logger.LogTx(ctx, tx, "role.grant", membership); err != nil {
		return fmt.Errorf("log grant governance: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit %s tx: %w", txLabel, err)
	}
	committed = true
	return nil
}

func buildMembership(userID, tenantID, areaCode string, role domain.Role, effectiveFrom time.Time, grantedBy string) domain.UserProcessArea {
	membership := domain.UserProcessArea{
		UserID:        userID,
		TenantID:      tenantID,
		AreaCode:      areaCode,
		Role:          role,
		EffectiveFrom: effectiveFrom,
	}
	if grantedBy != "" {
		membership.GrantedBy = &grantedBy
	}
	return membership
}

// Revoke closes the caller's active membership in areaCode as of now.
// Returns ErrMembershipNotFound if no active membership exists. The closure
// and its governance-log entry commit in the same transaction, then the
// actor's cached roles are invalidated (A3).
func (s *AreaMembershipService) Revoke(
	ctx context.Context,
	userID, tenantID, areaCode string,
	revokedBy string,
) error {
	now := s.nowFn()
	active, err := s.repo.GetActiveByUserAndArea(ctx, userID, tenantID, areaCode, now)
	if err != nil {
		return fmt.Errorf("get active membership: %w", err)
	}
	if active == nil || !active.IsActive(now) {
		return ErrMembershipNotFound
	}

	membership := *active
	if revokedBy != "" {
		membership.GrantedBy = &revokedBy
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin revoke tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := s.repo.CloseActiveTx(ctx, tx, userID, tenantID, areaCode, now, revokedBy); err != nil {
		return fmt.Errorf("close active membership: %w", err)
	}
	if err := s.logger.LogTx(ctx, tx, "role.revoke", membership); err != nil {
		return fmt.Errorf("log revoke governance: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoke tx: %w", err)
	}
	committed = true
	// Flush the actor's cached roles once the mutation and governance are
	// committed atomically (A3).
	s.invalidate(userID, tenantID)
	return nil
}
