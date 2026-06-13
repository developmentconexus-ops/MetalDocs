package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// MembershipTx is the transaction handle services receive from the repository
// for in-tx mutation + governance writes (REQ-ASYNC-1). It embeds db.Tx so the
// governance logger can write audit rows inside the same database transaction.
// *sql.Tx satisfies this interface without an adapter.
type MembershipTx interface {
	db.Tx
	Commit() error
	Rollback() error
}

var (
	ErrMembershipNotFound = errors.New("membership_not_found")
	ErrUnknownRole        = errors.New("unknown_role")
	// ErrMembershipExists is returned by Grant when an active membership for
	// (user, tenant, area) with the same role already exists. The admin
	// surface treats grant as non-idempotent: callers must explicitly revoke
	// before re-granting the same role; a role change still succeeds via
	// GrantAtomic (close-old + insert-new).
	ErrMembershipExists = errors.New("membership_exists")
)

type UserAreaWriteRepository interface {
	ListActive(ctx context.Context, userID, tenantID string, now time.Time) ([]domain.UserProcessArea, error)
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
	BeginTx(ctx context.Context) (MembershipTx, error)
	// InsertTx writes a new membership row inside an open transaction.
	InsertTx(ctx context.Context, tx MembershipTx, membership domain.UserProcessArea) error
	// CloseActiveTx sets effective_to on the active row inside an open transaction.
	CloseActiveTx(ctx context.Context, tx MembershipTx, userID, tenantID, areaCode string, effectiveTo time.Time, actorID string) error
	// GrantAtomicTx closes the old membership and inserts the new one inside an
	// open transaction (role-change path).
	GrantAtomicTx(ctx context.Context, tx MembershipTx, oldMembership, newMembership domain.UserProcessArea) error
}

type MembershipGovernanceLogger interface {
	// LogTx writes the governance event inside an open transaction so the audit
	// record is atomically committed with the membership mutation (REQ-ASYNC-1,
	// T-007). A failure rolls back the mutation.
	LogTx(ctx context.Context, tx db.Tx, action string, membership domain.UserProcessArea) error
}

type AreaMembershipService struct {
	repo        UserAreaWriteRepository
	logger      MembershipGovernanceLogger
	invalidator RoleCacheInvalidator
	nowFn       func() time.Time
}

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

func (s *AreaMembershipService) ListActive(ctx context.Context, userID, tenantID string) ([]domain.UserProcessArea, error) {
	return s.repo.ListActive(ctx, userID, tenantID, s.nowFn())
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

func (s *AreaMembershipService) Grant(
	ctx context.Context,
	userID, tenantID, areaCode string,
	role domain.Role,
	grantedBy string,
) error {
	if _, err := domain.ParseRole(string(role)); err != nil {
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
		tx, err := s.repo.BeginTx(ctx)
		if err != nil {
			return fmt.Errorf("begin grant tx: %w", err)
		}
		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback()
			}
		}()
		if err := s.repo.GrantAtomicTx(ctx, tx, *existing, membership); err != nil {
			return fmt.Errorf("grant membership atomically: %w", err)
		}
		if err := s.logger.LogTx(ctx, tx, "role.grant", membership); err != nil {
			return fmt.Errorf("log grant governance: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit grant tx: %w", err)
		}
		committed = true
		// Flush the actor's cached roles once the mutation and governance are
		// committed atomically (A3).
		s.invalidate(userID, tenantID)
		return nil
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin insert tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := s.repo.InsertTx(ctx, tx, membership); err != nil {
		return fmt.Errorf("insert membership: %w", err)
	}
	if err := s.logger.LogTx(ctx, tx, "role.grant", membership); err != nil {
		return fmt.Errorf("log grant governance: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit insert tx: %w", err)
	}
	committed = true
	// Flush the actor's cached roles once the mutation and governance are
	// committed atomically (A3).
	s.invalidate(userID, tenantID)
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
