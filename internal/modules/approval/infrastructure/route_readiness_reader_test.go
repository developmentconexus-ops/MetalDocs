package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/approval/domain"
)

// newRouteReadinessMock returns a sqlmock-backed *sql.DB plus its controller.
// The reader under test issues plain SELECTs, so the default regexp matcher is
// enough to pin the predicates that make the read tenant-safe.
func newRouteReadinessMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database, mock
}

func TestNewRouteReadinessReaderPG_PanicsOnNilPool(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil pool")
		}
	}()
	NewRouteReadinessReaderPG(nil)
}

// TestRouteReadinessReaderPG_ActiveRouteSubjectKeys pins the bulk read's shape:
// tenant + subject_kind + active=true, returning a set of subject keys.
func TestRouteReadinessReaderPG_ActiveRouteSubjectKeys(t *testing.T) {
	database, mock := newRouteReadinessMock(t)
	mock.ExpectQuery(`SELECT DISTINCT subject_key\s+FROM approval_routes\s+WHERE tenant_id\s+= \$1\s+AND subject_kind = \$2\s+AND active\s+= true`).
		WithArgs("tenant-a", string(domain.SubjectKindDocument)).
		WillReturnRows(sqlmock.NewRows([]string{"subject_key"}).AddRow("po").AddRow("it"))

	reader := NewRouteReadinessReaderPG(database)
	got, err := reader.ActiveRouteSubjectKeys(context.Background(), "tenant-a", string(domain.SubjectKindDocument))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d keys, want 2 (%+v)", len(got), got)
	}
	for _, want := range []string{"po", "it"} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing subject key %q", want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestRouteReadinessReaderPG_ActiveRouteSubjectKeys_EmptyIsNotNil keeps the
// consumer contract simple: "no routes" is an empty set, never a nil map.
func TestRouteReadinessReaderPG_ActiveRouteSubjectKeys_EmptyIsNotNil(t *testing.T) {
	database, mock := newRouteReadinessMock(t)
	mock.ExpectQuery("SELECT DISTINCT subject_key").
		WithArgs("tenant-a", string(domain.SubjectKindDocument)).
		WillReturnRows(sqlmock.NewRows([]string{"subject_key"}))

	reader := NewRouteReadinessReaderPG(database)
	got, err := reader.ActiveRouteSubjectKeys(context.Background(), "tenant-a", string(domain.SubjectKindDocument))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected an empty set, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected an empty set, got %+v", got)
	}
}

func TestRouteReadinessReaderPG_ActiveRouteSubjectKeys_PropagatesError(t *testing.T) {
	boom := errors.New("connection reset")
	database, mock := newRouteReadinessMock(t)
	mock.ExpectQuery("SELECT DISTINCT subject_key").WillReturnError(boom)

	reader := NewRouteReadinessReaderPG(database)
	if _, err := reader.ActiveRouteSubjectKeys(context.Background(), "tenant-a", "document"); !errors.Is(err, boom) {
		t.Fatalf("expected the driver error to propagate, got %v", err)
	}
}

// TestRouteReadinessReaderPG_HasActiveRoute_True pins the single-subject
// predicate: tenant + subject_kind + subject_key + active=true.
func TestRouteReadinessReaderPG_HasActiveRoute_True(t *testing.T) {
	database, mock := newRouteReadinessMock(t)
	mock.ExpectQuery(`SELECT 1\s+FROM approval_routes\s+WHERE tenant_id\s+= \$1\s+AND subject_kind = \$2\s+AND subject_key\s+= \$3\s+AND active\s+= true`).
		WithArgs("tenant-a", string(domain.SubjectKindDocument), "po").
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))

	reader := NewRouteReadinessReaderPG(database)
	ready, err := reader.HasActiveRoute(context.Background(), database, "tenant-a", string(domain.SubjectKindDocument), "po")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ready {
		t.Fatal("expected ready=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestRouteReadinessReaderPG_HasActiveRoute_NoRowsIsNotAnError is the gate's
// load-bearing case: "no active route" is a clean false, so the caller can map
// it to the 409 rather than a 500.
func TestRouteReadinessReaderPG_HasActiveRoute_NoRowsIsNotAnError(t *testing.T) {
	database, mock := newRouteReadinessMock(t)
	mock.ExpectQuery("SELECT 1").
		WithArgs("tenant-a", string(domain.SubjectKindDocument), "ghost").
		WillReturnError(sql.ErrNoRows)

	reader := NewRouteReadinessReaderPG(database)
	ready, err := reader.HasActiveRoute(context.Background(), database, "tenant-a", string(domain.SubjectKindDocument), "ghost")
	if err != nil {
		t.Fatalf("no-rows must not be an error, got %v", err)
	}
	if ready {
		t.Fatal("expected ready=false")
	}
}

func TestRouteReadinessReaderPG_HasActiveRoute_PropagatesError(t *testing.T) {
	boom := errors.New("deadlock detected")
	database, mock := newRouteReadinessMock(t)
	mock.ExpectQuery("SELECT 1").WillReturnError(boom)

	reader := NewRouteReadinessReaderPG(database)
	ready, err := reader.HasActiveRoute(context.Background(), database, "tenant-a", "document", "po")
	if !errors.Is(err, boom) {
		t.Fatalf("expected the driver error to propagate, got %v", err)
	}
	if ready {
		t.Fatal("a failed read must not report ready=true")
	}
}

// TestRouteReadinessReaderPG_HasActiveRoute_RunsOnCallerExecutor is the
// in-tx contract: the creation gate must observe the same snapshot as the
// write it guards, so the query goes to the caller's executor and NOT to the
// reader's own pool. Two independent mocks make a regression visible: the pool
// mock would report an unmet/unexpected call if the executor were ignored.
func TestRouteReadinessReaderPG_HasActiveRoute_RunsOnCallerExecutor(t *testing.T) {
	poolDB, poolMock := newRouteReadinessMock(t)
	execDB, execMock := newRouteReadinessMock(t)
	execMock.ExpectQuery("SELECT 1").
		WithArgs("tenant-a", string(domain.SubjectKindDocument), "po").
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))

	reader := NewRouteReadinessReaderPG(poolDB)
	if _, err := reader.HasActiveRoute(context.Background(), execDB, "tenant-a", string(domain.SubjectKindDocument), "po"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := execMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("caller executor expectations: %v", err)
	}
	if err := poolMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("pool must not be touched by HasActiveRoute: %v", err)
	}
}
