package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// noopMembershipTx satisfies MembershipTx for unit tests; Commit is a no-op.
type noopMembershipTx struct{}

func (noopMembershipTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}
func (noopMembershipTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, nil
}
func (noopMembershipTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}
func (noopMembershipTx) Commit() error   { return nil }
func (noopMembershipTx) Rollback() error { return nil }

// commitTrackingMembershipTx records whether Commit was called so rollback
// atomicity tests can assert the mutation was not committed when governance fails.
type commitTrackingMembershipTx struct {
	committed bool
}

func (t *commitTrackingMembershipTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}
func (t *commitTrackingMembershipTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, nil
}
func (t *commitTrackingMembershipTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}
func (t *commitTrackingMembershipTx) Commit() error   { t.committed = true; return nil }
func (t *commitTrackingMembershipTx) Rollback() error { return nil }

type userAreaWriteRepoStub struct {
	activeList       []domain.UserProcessArea
	active           *domain.UserProcessArea
	closeCalls       int
	insertCalls      int
	grantAtomicCalls int
	closedAt         time.Time
	closedActorID    string
	inserted         domain.UserProcessArea
	atomicOld        domain.UserProcessArea
	atomicNew        domain.UserProcessArea
	getActiveErr     error
	closeActiveErr   error
	insertErr        error
	grantAtomicErr   error
	tenantFilter     [4]string
	// tx returned by BeginTx; defaults to noopMembershipTx{} when nil.
	tx domain.MembershipTx
}

func (s *userAreaWriteRepoStub) beginTx() domain.MembershipTx {
	if s.tx != nil {
		return s.tx
	}
	return noopMembershipTx{}
}

func (s *userAreaWriteRepoStub) ListActive(ctx context.Context, userID, tenantID string, now time.Time) ([]domain.UserProcessArea, error) {
	return append([]domain.UserProcessArea(nil), s.activeList...), nil
}

func (s *userAreaWriteRepoStub) ListActiveForUsers(_ context.Context, _ string, _ []string, _ time.Time) (map[string][]domain.UserProcessArea, error) {
	return map[string][]domain.UserProcessArea{}, nil
}

func (s *userAreaWriteRepoStub) ListByTenant(ctx context.Context, tenantID, userID, areaCode, role string, now time.Time) ([]domain.UserProcessArea, error) {
	s.tenantFilter = [4]string{tenantID, userID, areaCode, role}
	return append([]domain.UserProcessArea(nil), s.activeList...), nil
}

func (s *userAreaWriteRepoStub) MembershipDirectoryScope(ctx context.Context, tenantID, actorID, capability string, now time.Time) (bool, bool, error) {
	return false, false, nil
}

func (s *userAreaWriteRepoStub) ListByTenantInManagedAreas(ctx context.Context, tenantID, userID, areaCode, role, actorID, capability string, now time.Time) ([]domain.UserProcessArea, error) {
	return append([]domain.UserProcessArea(nil), s.activeList...), nil
}

func (s *userAreaWriteRepoStub) GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*domain.UserProcessArea, error) {
	if s.getActiveErr != nil {
		return nil, s.getActiveErr
	}
	return s.active, nil
}

func (s *userAreaWriteRepoStub) BeginTx(_ context.Context) (domain.MembershipTx, error) {
	return s.beginTx(), nil
}

func (s *userAreaWriteRepoStub) InsertTx(_ context.Context, _ domain.MembershipTx, membership domain.UserProcessArea) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	s.insertCalls++
	s.inserted = membership
	return nil
}

func (s *userAreaWriteRepoStub) CloseActiveTx(_ context.Context, _ domain.MembershipTx, userID, tenantID, areaCode string, effectiveTo time.Time, actorID string) error {
	if s.closeActiveErr != nil {
		return s.closeActiveErr
	}
	s.closeCalls++
	s.closedAt = effectiveTo
	s.closedActorID = actorID
	return nil
}

func (s *userAreaWriteRepoStub) GrantAtomicTx(_ context.Context, _ domain.MembershipTx, oldMembership, newMembership domain.UserProcessArea) error {
	if s.grantAtomicErr != nil {
		return s.grantAtomicErr
	}
	s.grantAtomicCalls++
	s.atomicOld = oldMembership
	s.atomicNew = newMembership
	return nil
}

type membershipLoggerStub struct {
	actions []string
	logTxErr error
}

func (s *membershipLoggerStub) LogTx(_ context.Context, _ db.Tx, action string, _ domain.UserProcessArea) error {
	if s.logTxErr != nil {
		return s.logTxErr
	}
	s.actions = append(s.actions, action)
	return nil
}

// invalidatorSpy records (userID, tenantID) pairs passed to InvalidateUserTenant
// so the cache-flush guard (A3) can assert grant/revoke flush the role cache.
type invalidatorSpy struct {
	calls [][2]string
}

func (s *invalidatorSpy) InvalidateUserTenant(userID, tenantID string) {
	s.calls = append(s.calls, [2]string{userID, tenantID})
}

func TestGrant_New(t *testing.T) {
	repo := &userAreaWriteRepoStub{}
	logger := &membershipLoggerStub{}
	service := NewAreaMembershipService(repo, logger)
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	service.nowFn = func() time.Time { return now }

	err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleEditor, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.closeCalls != 0 {
		t.Fatalf("expected no close call, got %d", repo.closeCalls)
	}
	if repo.grantAtomicCalls != 0 {
		t.Fatalf("expected no atomic grant call, got %d", repo.grantAtomicCalls)
	}
	if repo.insertCalls != 1 {
		t.Fatalf("expected one insert call, got %d", repo.insertCalls)
	}
	if repo.inserted.Role != domain.RoleEditor {
		t.Fatalf("expected inserted role editor, got %q", repo.inserted.Role)
	}
	if repo.inserted.EffectiveFrom != now {
		t.Fatalf("expected effective_from %v, got %v", now, repo.inserted.EffectiveFrom)
	}
	if len(logger.actions) != 1 || logger.actions[0] != "role.grant" {
		t.Fatalf("expected role.grant log, got %v", logger.actions)
	}
}

func TestGrant_Overlap_Merge(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	repo := &userAreaWriteRepoStub{
		active: &domain.UserProcessArea{
			UserID:        "u1",
			TenantID:      "t1",
			AreaCode:      "A1",
			Role:          domain.RoleViewer,
			EffectiveFrom: now.Add(-2 * time.Hour),
		},
	}
	logger := &membershipLoggerStub{}
	service := NewAreaMembershipService(repo, logger)
	service.nowFn = func() time.Time { return now }

	err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleApprover, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.closeCalls != 0 {
		t.Fatalf("expected no close call (atomic path), got %d", repo.closeCalls)
	}
	if repo.insertCalls != 0 {
		t.Fatalf("expected no direct insert call (atomic path), got %d", repo.insertCalls)
	}
	if repo.grantAtomicCalls != 1 {
		t.Fatalf("expected one atomic grant call, got %d", repo.grantAtomicCalls)
	}
	if repo.atomicNew.Role != domain.RoleApprover {
		t.Fatalf("expected atomic new role approver, got %q", repo.atomicNew.Role)
	}
}

func TestListByTenant_PassesFilters(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	repo := &userAreaWriteRepoStub{
		activeList: []domain.UserProcessArea{
			{UserID: "u1", TenantID: "t1", AreaCode: "A1", Role: domain.RoleEditor, EffectiveFrom: now},
		},
	}
	service := NewAreaMembershipService(repo, &membershipLoggerStub{})
	service.nowFn = func() time.Time { return now }

	items, err := service.ListByTenant(context.Background(), "t1", "u1", "A1", "editor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if got, want := repo.tenantFilter, ([4]string{"t1", "u1", "A1", "editor"}); got != want {
		t.Fatalf("filters not threaded: got %v want %v", got, want)
	}
}

func TestRevoke_Active(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	repo := &userAreaWriteRepoStub{
		active: &domain.UserProcessArea{
			UserID:        "u1",
			TenantID:      "t1",
			AreaCode:      "A1",
			Role:          domain.RoleApprover,
			EffectiveFrom: now.Add(-time.Hour),
		},
	}
	logger := &membershipLoggerStub{}
	service := NewAreaMembershipService(repo, logger)
	service.nowFn = func() time.Time { return now }

	err := service.Revoke(context.Background(), "u1", "t1", "A1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.closeCalls != 1 {
		t.Fatalf("expected one close call, got %d", repo.closeCalls)
	}
	if len(logger.actions) != 1 || logger.actions[0] != "role.revoke" {
		t.Fatalf("expected role.revoke log, got %v", logger.actions)
	}
}

func TestRevoke_NonExistent(t *testing.T) {
	repo := &userAreaWriteRepoStub{}
	service := NewAreaMembershipService(repo, &membershipLoggerStub{})

	err := service.Revoke(context.Background(), "u1", "t1", "A1", "admin")
	if !errors.Is(err, ErrMembershipNotFound) {
		t.Fatalf("expected ErrMembershipNotFound, got %v", err)
	}
}

func TestGrant_UnknownRole(t *testing.T) {
	repo := &userAreaWriteRepoStub{}
	service := NewAreaMembershipService(repo, &membershipLoggerStub{})

	err := service.Grant(context.Background(), "u1", "t1", "A1", domain.Role("ghost"), "admin")
	if !errors.Is(err, ErrUnknownRole) {
		t.Fatalf("expected ErrUnknownRole, got %v", err)
	}
}

// TestAreaMembershipService_InvalidatesCache asserts every committed grant/revoke
// flushes the role cache for (userID, tenantID) — closing the window where a
// changed area membership keeps authorizing until the TTL expires (A3).
func TestAreaMembershipService_InvalidatesCache(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)

	t.Run("grant new (insert path)", func(t *testing.T) {
		repo := &userAreaWriteRepoStub{}
		spy := &invalidatorSpy{}
		service := NewAreaMembershipService(repo, &membershipLoggerStub{}).WithRoleCacheInvalidator(spy)
		service.nowFn = func() time.Time { return now }

		if err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleEditor, "admin"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := [][2]string{{"u1", "t1"}}; !equalCalls(spy.calls, want) {
			t.Fatalf("invalidator calls = %v, want %v", spy.calls, want)
		}
	})

	t.Run("grant role change (atomic path)", func(t *testing.T) {
		repo := &userAreaWriteRepoStub{
			active: &domain.UserProcessArea{
				UserID:        "u1",
				TenantID:      "t1",
				AreaCode:      "A1",
				Role:          domain.RoleViewer,
				EffectiveFrom: now.Add(-time.Hour),
			},
		}
		spy := &invalidatorSpy{}
		service := NewAreaMembershipService(repo, &membershipLoggerStub{}).WithRoleCacheInvalidator(spy)
		service.nowFn = func() time.Time { return now }

		if err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleApprover, "admin"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := [][2]string{{"u1", "t1"}}; !equalCalls(spy.calls, want) {
			t.Fatalf("invalidator calls = %v, want %v", spy.calls, want)
		}
	})

	t.Run("revoke", func(t *testing.T) {
		repo := &userAreaWriteRepoStub{
			active: &domain.UserProcessArea{
				UserID:        "u1",
				TenantID:      "t1",
				AreaCode:      "A1",
				Role:          domain.RoleApprover,
				EffectiveFrom: now.Add(-time.Hour),
			},
		}
		spy := &invalidatorSpy{}
		service := NewAreaMembershipService(repo, &membershipLoggerStub{}).WithRoleCacheInvalidator(spy)
		service.nowFn = func() time.Time { return now }

		if err := service.Revoke(context.Background(), "u1", "t1", "A1", "admin"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := [][2]string{{"u1", "t1"}}; !equalCalls(spy.calls, want) {
			t.Fatalf("invalidator calls = %v, want %v", spy.calls, want)
		}
	})

	t.Run("duplicate grant does not flush", func(t *testing.T) {
		repo := &userAreaWriteRepoStub{
			active: &domain.UserProcessArea{
				UserID:        "u1",
				TenantID:      "t1",
				AreaCode:      "A1",
				Role:          domain.RoleApprover,
				EffectiveFrom: now.Add(-time.Hour),
			},
		}
		spy := &invalidatorSpy{}
		service := NewAreaMembershipService(repo, &membershipLoggerStub{}).WithRoleCacheInvalidator(spy)
		service.nowFn = func() time.Time { return now }

		// Same role on an active row → ErrMembershipExists, no mutation, no flush.
		if err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleApprover, "admin"); !errors.Is(err, ErrMembershipExists) {
			t.Fatalf("expected ErrMembershipExists, got %v", err)
		}
		if len(spy.calls) != 0 {
			t.Fatalf("expected no invalidation on no-op grant, got %v", spy.calls)
		}
	})
}

func equalCalls(got, want [][2]string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestGrant_GovernanceError_RollsBack asserts that when LogTx returns an error
// Commit is never called (the mutation is rolled back atomically — T-007 /
// REQ-ASYNC-1).
func TestGrant_GovernanceError_RollsBack(t *testing.T) {
	govErr := errors.New("governance write failed")

	t.Run("insert path", func(t *testing.T) {
		trackedTx := &commitTrackingMembershipTx{}
		repo := &userAreaWriteRepoStub{tx: trackedTx}
		logger := &membershipLoggerStub{logTxErr: govErr}
		service := NewAreaMembershipService(repo, logger)

		err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleEditor, "admin")
		if err == nil {
			t.Fatal("expected error when LogTx fails, got nil")
		}
		if !errors.Is(err, govErr) {
			t.Fatalf("expected wrapped govErr, got: %v", err)
		}
		if trackedTx.committed {
			t.Fatal("Commit must NOT be called when LogTx fails")
		}
	})

	t.Run("atomic (role change) path", func(t *testing.T) {
		now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
		trackedTx := &commitTrackingMembershipTx{}
		repo := &userAreaWriteRepoStub{
			tx: trackedTx,
			active: &domain.UserProcessArea{
				UserID:        "u1",
				TenantID:      "t1",
				AreaCode:      "A1",
				Role:          domain.RoleViewer,
				EffectiveFrom: now.Add(-time.Hour),
			},
		}
		logger := &membershipLoggerStub{logTxErr: govErr}
		service := NewAreaMembershipService(repo, logger)
		service.nowFn = func() time.Time { return now }

		err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleApprover, "admin")
		if err == nil {
			t.Fatal("expected error when LogTx fails, got nil")
		}
		if !errors.Is(err, govErr) {
			t.Fatalf("expected wrapped govErr, got: %v", err)
		}
		if trackedTx.committed {
			t.Fatal("Commit must NOT be called when LogTx fails")
		}
	})
}

// TestRevoke_GovernanceError_RollsBack asserts that when LogTx returns an error
// during revoke, Commit is never called (REQ-ASYNC-1).
func TestRevoke_GovernanceError_RollsBack(t *testing.T) {
	govErr := errors.New("governance write failed")
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	trackedTx := &commitTrackingMembershipTx{}
	repo := &userAreaWriteRepoStub{
		tx: trackedTx,
		active: &domain.UserProcessArea{
			UserID:        "u1",
			TenantID:      "t1",
			AreaCode:      "A1",
			Role:          domain.RoleApprover,
			EffectiveFrom: now.Add(-time.Hour),
		},
	}
	logger := &membershipLoggerStub{logTxErr: govErr}
	service := NewAreaMembershipService(repo, logger)
	service.nowFn = func() time.Time { return now }

	err := service.Revoke(context.Background(), "u1", "t1", "A1", "admin")
	if err == nil {
		t.Fatal("expected error when LogTx fails, got nil")
	}
	if !errors.Is(err, govErr) {
		t.Fatalf("expected wrapped govErr, got: %v", err)
	}
	if trackedTx.committed {
		t.Fatal("Commit must NOT be called when LogTx fails")
	}
}

func TestTemporalQuery_EffectiveTo_Past(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Minute)
	repo := &userAreaWriteRepoStub{
		active: &domain.UserProcessArea{
			UserID:        "u1",
			TenantID:      "t1",
			AreaCode:      "A1",
			Role:          domain.RoleEditor,
			EffectiveFrom: now.Add(-time.Hour),
			EffectiveTo:   &past,
		},
	}
	service := NewAreaMembershipService(repo, &membershipLoggerStub{})
	service.nowFn = func() time.Time { return now }

	err := service.Revoke(context.Background(), "u1", "t1", "A1", "admin")
	if !errors.Is(err, ErrMembershipNotFound) {
		t.Fatalf("expected ErrMembershipNotFound for expired active query result, got %v", err)
	}
}
