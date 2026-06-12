package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/db"
)

func TestAreaServiceCreate_UsesDomainConstructorNormalizationAndValidation(t *testing.T) {
	repo := newFakeAreaRepository()
	service := NewAreaService(repo, &fakeGovernanceLogger{})

	in := &domain.ProcessArea{
		Code:     " AR-01 ",
		TenantID: " tenant-a ",
		Name:     " Finance ",
	}
	if err := service.Create(context.Background(), in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := repo.get("tenant-a", "AR-01")
	if got == nil {
		t.Fatal("expected created area in repository")
	}
	if got.Code != "AR-01" || got.TenantID != "tenant-a" || got.Name != "Finance" {
		t.Fatalf("unexpected normalized area: %#v", got)
	}
	if in.Code != " AR-01 " || in.TenantID != " tenant-a " || in.Name != " Finance " {
		t.Fatalf("caller pointer mutated: %#v", in)
	}
	if err := service.Create(context.Background(), &domain.ProcessArea{
		Code:     " ",
		TenantID: "tenant-a",
		Name:     "Finance",
	}); !errors.Is(err, domain.ErrAreaCodeRequired) {
		t.Fatalf("expected ErrAreaCodeRequired, got %v", err)
	}
}

func TestAreaServiceArchiveSetsArchivedAt(t *testing.T) {
	repo := newFakeAreaRepository()
	repo.put(&domain.ProcessArea{Code: "child", TenantID: "tenant-a"})

	service := NewAreaService(repo, &fakeGovernanceLogger{})
	fixedNow := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	err := service.Archive(context.Background(), "tenant-a", "child", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := repo.get("tenant-a", "child")
	if updated.ArchivedAt == nil || !updated.ArchivedAt.Equal(fixedNow) {
		t.Fatalf("expected archived_at %s, got %+v", fixedNow, updated.ArchivedAt)
	}
}

func TestAreaServiceArchiveFailsWhenAlreadyArchived(t *testing.T) {
	archivedAt := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	repo := newFakeAreaRepository()
	repo.put(&domain.ProcessArea{Code: "child", TenantID: "tenant-a", ArchivedAt: &archivedAt})

	service := NewAreaService(repo, &fakeGovernanceLogger{})
	err := service.Archive(context.Background(), "tenant-a", "child", "user-1")
	if !errors.Is(err, domain.ErrAreaArchived) {
		t.Fatalf("expected ErrAreaArchived, got %v", err)
	}
}

type fakeAreaRepository struct {
	byKey           map[string]*domain.ProcessArea
	ancestorsByCode map[string][]domain.AreaCode
}

func newFakeAreaRepository() *fakeAreaRepository {
	return &fakeAreaRepository{
		byKey:           map[string]*domain.ProcessArea{},
		ancestorsByCode: map[string][]domain.AreaCode{},
	}
}

func (r *fakeAreaRepository) GetByCode(_ context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	item, ok := r.byKey[tenantID+"|"+string(code)]
	if !ok {
		return nil, domain.ErrAreaNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeAreaRepository) List(_ context.Context, _ string, _ bool) ([]domain.ProcessArea, error) {
	return nil, nil
}

func (r *fakeAreaRepository) Create(_ context.Context, a *domain.ProcessArea) error {
	r.put(a)
	return nil
}

func (r *fakeAreaRepository) CreateTx(_ context.Context, _ domain.FamilyTx, a *domain.ProcessArea) error {
	r.put(a)
	return nil
}

func (r *fakeAreaRepository) Update(_ context.Context, a *domain.ProcessArea) error {
	r.put(a)
	return nil
}

func (r *fakeAreaRepository) ListAncestors(_ context.Context, _ string, code domain.AreaCode) ([]domain.AreaCode, error) {
	return r.ancestorsByCode[string(code)], nil
}

func (r *fakeAreaRepository) BeginTx(_ context.Context) (domain.FamilyTx, error) {
	return fakeTx{}, nil
}

func (r *fakeAreaRepository) GetByCodeForUpdate(ctx context.Context, _ domain.FamilyTx, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	return r.GetByCode(ctx, tenantID, code)
}

func (r *fakeAreaRepository) ListAncestorsTx(_ context.Context, _ domain.FamilyTx, _ string, code domain.AreaCode) ([]domain.AreaCode, error) {
	return r.ancestorsByCode[string(code)], nil
}

func (r *fakeAreaRepository) UpdateTx(_ context.Context, _ domain.FamilyTx, a *domain.ProcessArea) error {
	r.put(a)
	return nil
}

func (r *fakeAreaRepository) put(a *domain.ProcessArea) {
	copy := *a
	r.byKey[a.TenantID+"|"+string(a.Code)] = &copy
}

func (r *fakeAreaRepository) get(tenantID, code string) *domain.ProcessArea {
	return r.byKey[tenantID+"|"+code]
}

// commitTrackingTx records whether Commit was called.
type commitTrackingTx struct {
	committed bool
}

func (t *commitTrackingTx) Commit() error   { t.committed = true; return nil }
func (t *commitTrackingTx) Rollback() error { return nil }
func (t *commitTrackingTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}
func (t *commitTrackingTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, nil
}
func (t *commitTrackingTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}

// TestAreaServiceCreate_AtomicRollback_WhenLogTxFails asserts that when the
// governance LogTx call returns an error the mutation is not persisted (the
// transaction is rolled back and Commit is never called).
func TestAreaServiceCreate_AtomicRollback_WhenLogTxFails(t *testing.T) {
	tx := &commitTrackingTx{}
	repo := &atomicFakeAreaRepository{
		fakeAreaRepository: *newFakeAreaRepository(),
		tx:                 tx,
	}

	govErr := errors.New("governance write failed")
	logger := &fakeGovernanceLoggerWithError{err: govErr}

	service := NewAreaService(repo, logger)
	err := service.Create(context.Background(), &domain.ProcessArea{
		Code:     "AR-ATOMIC",
		TenantID: "tenant-a",
		Name:     "Atomic Test",
	})

	if err == nil {
		t.Fatal("expected error when LogTx fails, got nil")
	}
	if !errors.Is(err, govErr) {
		t.Fatalf("expected wrapped govErr, got: %v", err)
	}
	if tx.committed {
		t.Fatal("Commit must NOT be called when LogTx fails")
	}
	if repo.get("tenant-a", "AR-ATOMIC") != nil {
		t.Fatal("area must not be persisted when governance write fails and tx is rolled back")
	}
}

// atomicFakeAreaRepository wraps fakeAreaRepository with a commitTrackingTx
// and stages writes in a pending map; writes are only promoted on Commit.
type atomicFakeAreaRepository struct {
	fakeAreaRepository
	tx      *commitTrackingTx
	pending map[string]*domain.ProcessArea
}

func (r *atomicFakeAreaRepository) BeginTx(_ context.Context) (domain.FamilyTx, error) {
	r.pending = map[string]*domain.ProcessArea{}
	return r.tx, nil
}

func (r *atomicFakeAreaRepository) CreateTx(_ context.Context, _ domain.FamilyTx, a *domain.ProcessArea) error {
	cp := *a
	r.pending[a.TenantID+"|"+string(a.Code)] = &cp
	return nil
}

// fakeGovernanceLoggerWithError returns err from LogTx and nil from Log.
// It satisfies domain.GovernanceLogger.
type fakeGovernanceLoggerWithError struct {
	err error
}

func (f *fakeGovernanceLoggerWithError) Log(_ context.Context, _ domain.GovernanceEvent) error {
	return nil
}

func (f *fakeGovernanceLoggerWithError) LogTx(_ context.Context, _ db.Tx, _ domain.GovernanceEvent) error {
	return f.err
}
