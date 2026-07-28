package controlleddocuments_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	controlleddocuments "metaldocs/internal/modules/controlleddocuments"
	"metaldocs/internal/modules/controlleddocuments/application"
	documentsdomain "metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/db"
)

// TestModule_New_WithInjectedFakes proves F-CD4: the CD module constructor
// accepts interface-typed collaborators (no real sibling schema required) and
// does not instantiate any sibling-module infrastructure internally.
func TestModule_New_WithInjectedFakes(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	// sqlmock does not need any expectations — the constructor only stores the DB.
	mock.MatchExpectationsInOrder(false)

	m := controlleddocuments.New(controlleddocuments.Dependencies{
		DB:                     mockDB,
		ActiveInstanceReader:   fakeActiveInstanceReader{},
		ProfileReader:          fakeProfileReader{},
		AreaReader:             fakeAreaReader{},
		GovernanceLogger:       fakeGovernanceLogger{},
		TemplateVersionChecker: fakeTemplateVersionChecker{},
		RouteReadinessReader:   approvaldomain.NoopRouteReadinessReader{},
		ProfileLister:          fakeProfileLister{},
		AreaLister:             fakeAreaLister{},
		AreaCapabilityReader:   iamdomain.NoopAreaCapabilityReader{},
	})

	if m == nil {
		t.Fatal("New returned nil")
	}
	if m.Handler == nil {
		t.Fatal("Module.Handler is nil")
	}
	if m.Service() == nil {
		t.Fatal("Module.Service() is nil")
	}
	if m.Repo() == nil {
		t.Fatal("Module.Repo() is nil")
	}
}

// TestModule_New_NilGuards verifies the constructor panics when a required
// interface port is nil (mirrors the taxonomy module nil-guard pattern).
func TestModule_New_NilGuards(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	mock.MatchExpectationsInOrder(false)

	makeDeps := func() controlleddocuments.Dependencies {
		return controlleddocuments.Dependencies{
			DB:                     mockDB,
			ActiveInstanceReader:   fakeActiveInstanceReader{},
			ProfileReader:          fakeProfileReader{},
			AreaReader:             fakeAreaReader{},
			GovernanceLogger:       fakeGovernanceLogger{},
			TemplateVersionChecker: fakeTemplateVersionChecker{},
			RouteReadinessReader:   approvaldomain.NoopRouteReadinessReader{},
			ProfileLister:          fakeProfileLister{},
			AreaLister:             fakeAreaLister{},
			AreaCapabilityReader:   iamdomain.NoopAreaCapabilityReader{},
		}
	}
	assertPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatalf("%s: expected panic on nil port, got none", name)
			}
		}()
		fn()
	}

	assertPanic("nil ProfileReader", func() {
		d := makeDeps()
		d.ProfileReader = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil AreaReader", func() {
		d := makeDeps()
		d.AreaReader = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil GovernanceLogger", func() {
		d := makeDeps()
		d.GovernanceLogger = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil TemplateVersionChecker", func() {
		d := makeDeps()
		d.TemplateVersionChecker = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil RouteReadinessReader", func() {
		d := makeDeps()
		d.RouteReadinessReader = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil ProfileLister", func() {
		d := makeDeps()
		d.ProfileLister = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil AreaLister", func() {
		d := makeDeps()
		d.AreaLister = nil
		controlleddocuments.New(d)
	})
	assertPanic("nil AreaCapabilityReader", func() {
		d := makeDeps()
		d.AreaCapabilityReader = nil
		controlleddocuments.New(d)
	})
}

// TestModule_New_NilActiveInstanceReader_UsesNoop verifies that a nil
// ActiveInstanceReader is accepted and silently replaced with the Noop default
// (mirrors documents.New pattern — the port has a well-defined Noop).
func TestModule_New_NilActiveInstanceReader_UsesNoop(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	mock.MatchExpectationsInOrder(false)

	// Should not panic.
	m := controlleddocuments.New(controlleddocuments.Dependencies{
		DB:                     mockDB,
		ActiveInstanceReader:   nil, // expects Noop default
		ProfileReader:          fakeProfileReader{},
		AreaReader:             fakeAreaReader{},
		GovernanceLogger:       fakeGovernanceLogger{},
		TemplateVersionChecker: fakeTemplateVersionChecker{},
		RouteReadinessReader:   approvaldomain.NoopRouteReadinessReader{},
		ProfileLister:          fakeProfileLister{},
		AreaLister:             fakeAreaLister{},
		AreaCapabilityReader:   iamdomain.NoopAreaCapabilityReader{},
	})
	if m == nil {
		t.Fatal("New returned nil with nil ActiveInstanceReader")
	}
}

// --- minimal fakes satisfying the injected interface ports ---

type fakeActiveInstanceReader struct{}

func (fakeActiveInstanceReader) ActiveInstanceForControlledDocument(_ context.Context, _, _ string) (*documentsdomain.ActiveInstanceView, error) {
	return nil, nil
}

type fakeProfileReader struct{}

func (fakeProfileReader) GetByCode(_ context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	return &taxonomydomain.DocumentProfile{TenantID: tenantID, Code: taxonomydomain.ProfileCode(code)}, nil
}

type fakeAreaReader struct{}

func (fakeAreaReader) GetByCode(_ context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error) {
	return &taxonomydomain.ProcessArea{TenantID: tenantID, Code: taxonomydomain.AreaCode(code)}, nil
}

type fakeProfileLister struct{}

func (fakeProfileLister) ListProfiles(_ context.Context, _ string) ([]taxonomydomain.DocumentProfile, error) {
	return nil, nil
}

type fakeAreaLister struct{}

func (fakeAreaLister) ListAreas(_ context.Context, _ string) ([]taxonomydomain.ProcessArea, error) {
	return nil, nil
}

type fakeGovernanceLogger struct{}

func (fakeGovernanceLogger) Log(_ context.Context, _ taxonomydomain.GovernanceEvent) error {
	return nil
}
func (fakeGovernanceLogger) LogTx(_ context.Context, _ db.Tx, _ taxonomydomain.GovernanceEvent) error {
	return nil
}

type fakeTemplateVersionChecker struct{}

func (fakeTemplateVersionChecker) GetTemplateVersionState(_ context.Context, _, _ string) (*string, string, error) {
	return nil, "", nil
}

// compile-time interface satisfaction checks
var _ application.ProfileReader = fakeProfileReader{}
var _ application.AreaReader = fakeAreaReader{}
var _ application.ProfileLister = fakeProfileLister{}
var _ application.AreaLister = fakeAreaLister{}
var _ taxonomydomain.GovernanceLogger = fakeGovernanceLogger{}
var _ application.TemplateVersionChecker = fakeTemplateVersionChecker{}
var _ documentsdomain.ActiveInstanceReader = fakeActiveInstanceReader{}

// Ensure the *sql.DB from sqlmock compiles as the DB field type.
var _ *sql.DB = (*sql.DB)(nil)
