package infrastructure

import (
	"context"
	"errors"
	"testing"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

type fakeProfileGetter struct {
	gotTenantID string
	gotCode     taxonomydomain.ProfileCode
	profile     *taxonomydomain.DocumentProfile
	err         error
}

func (f *fakeProfileGetter) GetByCode(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error) {
	f.gotTenantID = tenantID
	f.gotCode = code
	return f.profile, f.err
}

type fakeAreaGetter struct {
	gotTenantID string
	gotCode     taxonomydomain.AreaCode
	area        *taxonomydomain.ProcessArea
	err         error
}

func (f *fakeAreaGetter) GetByCode(ctx context.Context, tenantID string, code taxonomydomain.AreaCode) (*taxonomydomain.ProcessArea, error) {
	f.gotTenantID = tenantID
	f.gotCode = code
	return f.area, f.err
}

func TestTaxonomyProfileReaderDelegatesAndConvertsCode(t *testing.T) {
	want := &taxonomydomain.DocumentProfile{Code: "PROC-DESIGN"}
	getter := &fakeProfileGetter{profile: want}
	reader := NewTaxonomyProfileReader(getter)

	got, err := reader.GetByCode(context.Background(), "tenant-1", "PROC-DESIGN")
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got != want {
		t.Fatalf("expected delegated profile pointer, got %#v", got)
	}
	if getter.gotTenantID != "tenant-1" {
		t.Fatalf("tenant not forwarded: %q", getter.gotTenantID)
	}
	if getter.gotCode != taxonomydomain.ProfileCode("PROC-DESIGN") {
		t.Fatalf("code not converted to ProfileCode: %q", getter.gotCode)
	}
}

func TestTaxonomyProfileReaderPropagatesError(t *testing.T) {
	sentinel := errors.New("authz: capability denied")
	reader := NewTaxonomyProfileReader(&fakeProfileGetter{err: sentinel})

	if _, err := reader.GetByCode(context.Background(), "tenant-1", "X"); !errors.Is(err, sentinel) {
		t.Fatalf("expected propagated error, got %v", err)
	}
}

func TestTaxonomyAreaReaderDelegatesAndConvertsCode(t *testing.T) {
	want := &taxonomydomain.ProcessArea{Code: "AREA-OPS"}
	getter := &fakeAreaGetter{area: want}
	reader := NewTaxonomyAreaReader(getter)

	got, err := reader.GetByCode(context.Background(), "tenant-9", "AREA-OPS")
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got != want {
		t.Fatalf("expected delegated area pointer, got %#v", got)
	}
	if getter.gotTenantID != "tenant-9" {
		t.Fatalf("tenant not forwarded: %q", getter.gotTenantID)
	}
	if getter.gotCode != taxonomydomain.AreaCode("AREA-OPS") {
		t.Fatalf("code not converted to AreaCode: %q", getter.gotCode)
	}
}

func TestTaxonomyAreaReaderPropagatesError(t *testing.T) {
	sentinel := errors.New("authz: capability denied")
	reader := NewTaxonomyAreaReader(&fakeAreaGetter{err: sentinel})

	if _, err := reader.GetByCode(context.Background(), "tenant-1", "X"); !errors.Is(err, sentinel) {
		t.Fatalf("expected propagated error, got %v", err)
	}
}
