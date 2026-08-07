package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/tokens/application"
	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/platform/db"
)

type fakeRunner struct{}

func (fakeRunner) Do(ctx context.Context, fn func(tx *sql.Tx) error) error         { return fn(nil) }
func (fakeRunner) DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error { return fn(nil) }

type fakeAudit struct{ called bool }

func (f *fakeAudit) RecordTx(ctx context.Context, tx db.Tx, e auditdomain.Event) error {
	f.called = true
	return nil
}

type fakeRepo struct {
	entry  *domain.Entry
	getErr error
}

func (f *fakeRepo) Create(ctx context.Context, tx db.Tx, e *domain.Entry) (*domain.Entry, error) {
	return e, nil
}
func (f *fakeRepo) Update(ctx context.Context, tx db.Tx, e *domain.Entry) (*domain.Entry, error) {
	return e, nil
}
func (f *fakeRepo) Delete(ctx context.Context, tx db.Tx, tenantID, id string) error { return nil }
func (f *fakeRepo) GetByID(ctx context.Context, tx db.Tx, tenantID, id string) (*domain.Entry, error) {
	return f.entry, f.getErr
}
func (f *fakeRepo) GetByName(ctx context.Context, tx db.Tx, tenantID, name string) (*domain.Entry, error) {
	return f.entry, f.getErr
}
func (f *fakeRepo) List(ctx context.Context, tx db.Tx, tenantID string) ([]domain.Entry, error) {
	return nil, nil
}

func newSvc(repo domain.Repository, audit application.AuditRecorderForTest) *application.Service {
	noopRequire := func(ctx context.Context, tx *sql.Tx, cap, area string) error { return nil }
	noopSeed := func(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error { return nil }
	return application.NewServiceForTest(fakeRunner{}, repo, audit, noopRequire, noopSeed)
}

func TestUpdate_NotFound(t *testing.T) {
	svc := newSvc(&fakeRepo{getErr: domain.ErrNotFound}, &fakeAudit{})
	_, err := svc.Update(context.Background(), application.UpdateCommand{
		TenantID: "t", ActorID: "a", ID: "missing", Value: "v", Label: "l",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Update = %v, want ErrNotFound", err)
	}
}

func TestCreate_RecordsAudit(t *testing.T) {
	au := &fakeAudit{}
	svc := newSvc(&fakeRepo{}, au)
	_, err := svc.Create(context.Background(), application.CreateCommand{
		TenantID: "t", ActorID: "a", Name: "company_slogan", Value: "v", Label: "l",
	})
	if err != nil {
		t.Fatalf("Create = %v, want nil", err)
	}
	if !au.called {
		t.Fatal("audit.RecordTx not called on Create")
	}
}

func TestUpdate_NameChangeRejected(t *testing.T) {
	svc := newSvc(&fakeRepo{entry: &domain.Entry{ID: "id", Name: "old", Value: "v", Label: "l"}}, &fakeAudit{})
	_, err := svc.Update(context.Background(), application.UpdateCommand{
		TenantID: "t", ActorID: "a", ID: "id", Name: "new", Value: "v2", Label: "l2",
	})
	if !errors.Is(err, domain.ErrImmutableName) {
		t.Fatalf("Update = %v, want ErrImmutableName", err)
	}
}
