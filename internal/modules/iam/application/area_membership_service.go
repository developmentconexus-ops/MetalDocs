package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"metaldocs/internal/modules/iam/domain"
)

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
	Insert(ctx context.Context, membership domain.UserProcessArea) error
	CloseActive(ctx context.Context, userID, tenantID, areaCode string, effectiveTo time.Time, actorID string) error
	GrantAtomic(ctx context.Context, oldMembership, newMembership domain.UserProcessArea) error
	GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*domain.UserProcessArea, error)
}

type MembershipGovernanceLogger interface {
	Log(ctx context.Context, action string, membership domain.UserProcessArea) error
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
		if err := s.repo.GrantAtomic(ctx, *existing, membership); err != nil {
			return fmt.Errorf("grant membership atomically: %w", err)
		}
		// Flush the actor's cached roles unconditionally once the mutation commits:
		// the cache flush is a best-effort safety net and must not be gated on the
		// governance-logger outcome (A3).
		s.invalidate(userID, tenantID)
		if err := s.logger.Log(ctx, "role.grant", membership); err != nil {
			// Best-effort governance (A3): the membership mutation is already
			// committed; a governance-sink failure must not produce a torn outcome.
			// Atomic in-tx governance via RecordTx is the eventual target (T-007 /
			// next-touch refactor of the membership repo to share a tx).
			slog.WarnContext(ctx, "membership governance log failed (best-effort)", "action", "role.grant", "tenant_id", membership.TenantID, "user_id", membership.UserID, "err", err)
		}
		return nil
	}

	if err := s.repo.Insert(ctx, membership); err != nil {
		return fmt.Errorf("insert membership: %w", err)
	}
	// Flush the actor's cached roles unconditionally once the mutation commits
	// (see above): not gated on the governance-logger outcome (A3).
	s.invalidate(userID, tenantID)
	if err := s.logger.Log(ctx, "role.grant", membership); err != nil {
		// Best-effort governance (A3): the membership mutation is already
		// committed; a governance-sink failure must not produce a torn outcome.
		// Atomic in-tx governance via RecordTx is the eventual target (T-007 /
		// next-touch refactor of the membership repo to share a tx).
		slog.WarnContext(ctx, "membership governance log failed (best-effort)", "action", "role.grant", "tenant_id", membership.TenantID, "user_id", membership.UserID, "err", err)
	}
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

	if err := s.repo.CloseActive(ctx, userID, tenantID, areaCode, now, revokedBy); err != nil {
		return fmt.Errorf("close active membership: %w", err)
	}

	// Flush the actor's cached roles unconditionally once the mutation commits:
	// the cache flush is a best-effort safety net and must not be gated on the
	// governance-logger outcome (A3).
	s.invalidate(userID, tenantID)
	membership := *active
	if revokedBy != "" {
		membership.GrantedBy = &revokedBy
	}
	if err := s.logger.Log(ctx, "role.revoke", membership); err != nil {
		// Best-effort governance (A3): the membership mutation is already
		// committed; a governance-sink failure must not produce a torn outcome.
		// Atomic in-tx governance via RecordTx is the eventual target (T-007 /
		// next-touch refactor of the membership repo to share a tx).
		slog.WarnContext(ctx, "membership governance log failed (best-effort)", "action", "role.revoke", "tenant_id", membership.TenantID, "user_id", membership.UserID, "err", err)
	}
	return nil
}
