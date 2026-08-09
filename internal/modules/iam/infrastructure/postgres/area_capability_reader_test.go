package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

var areaCapabilityNow = time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

// TestAreaCapabilityReaderPG_NewPanicsOnNilDB keeps the composition root honest:
// an unwired reader must fail loudly at startup, not return an empty (and
// therefore silently over-restrictive) area list at request time.
func TestAreaCapabilityReaderPG_NewPanicsOnNilDB(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil db")
		}
	}()
	NewAreaCapabilityReaderPG(nil)
}

// TestAreaCapabilityReaderPG_AreasWithCapability_ReturnsGrantedAreas pins the
// positional arg contract shared with authz.SystemAdminExistsSQL
// ($1=actor, $2=tenant, $3=capability, $4=now) and the pq.Array decode.
func TestAreaCapabilityReaderPG_AreasWithCapability_ReturnsGrantedAreas(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("FROM metaldocs.role_capabilities rc").
		WithArgs("actor-1", "tenant-a", string(iamdomain.CapControlledDocumentCreate), areaCapabilityNow).
		WillReturnRows(sqlmock.NewRows([]string{"tenant_wide", "area_codes"}).
			AddRow(false, []byte("{finance,quality}")))

	tenantWide, areas, err := NewAreaCapabilityReaderPG(db).AreasWithCapability(
		context.Background(), "tenant-a", "actor-1", string(iamdomain.CapControlledDocumentCreate), areaCapabilityNow,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenantWide {
		t.Error("tenant_wide = true, want false for a non-system_admin actor")
	}
	if len(areas) != 2 || areas[0] != "finance" || areas[1] != "quality" {
		t.Fatalf("areas = %+v, want [finance quality]", areas)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestAreaCapabilityReaderPG_AreasWithCapability_TenantWide covers the
// system_admin bypass flag consumers use to short-circuit narrowing, mirroring
// the tier-2 inheritance bypass authz.Require already honors.
func TestAreaCapabilityReaderPG_AreasWithCapability_TenantWide(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("iam_user_roles").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_wide", "area_codes"}).
			AddRow(true, []byte("{}")))

	tenantWide, areas, err := NewAreaCapabilityReaderPG(db).AreasWithCapability(
		context.Background(), "tenant-a", "admin-1", string(iamdomain.CapControlledDocumentCreate), areaCapabilityNow,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tenantWide {
		t.Fatal("tenant_wide = false, want true")
	}
	if areas == nil {
		t.Fatal("areas must be an empty slice, never nil")
	}
	if len(areas) != 0 {
		t.Fatalf("areas = %+v, want empty", areas)
	}
}

// TestAreaCapabilityReaderPG_AreasWithCapability_PropagatesError keeps an iam
// read failure loud: degrading to an empty list would silently hide areas the
// caller may legitimately create into.
func TestAreaCapabilityReaderPG_AreasWithCapability_PropagatesError(t *testing.T) {
	boom := errors.New("connection reset")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("role_capabilities").WillReturnError(boom)

	if _, _, err := NewAreaCapabilityReaderPG(db).AreasWithCapability(
		context.Background(), "tenant-a", "actor-1", "controlled_documents.create", areaCapabilityNow,
	); !errors.Is(err, boom) {
		t.Fatalf("expected the driver error to propagate, got %v", err)
	}
}

// TestNoopAreaCapabilityReader_IsFailClosed documents the null object's
// semantics: no tenant-wide bypass and no areas.
func TestNoopAreaCapabilityReader_IsFailClosed(t *testing.T) {
	tenantWide, areas, err := iamdomain.NoopAreaCapabilityReader{}.AreasWithCapability(
		context.Background(), "tenant-a", "actor-1", "controlled_documents.create", areaCapabilityNow,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenantWide {
		t.Error("noop must not report a tenant-wide bypass")
	}
	if len(areas) != 0 {
		t.Errorf("noop must grant no areas, got %+v", areas)
	}
}
