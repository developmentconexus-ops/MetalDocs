package infrastructure

import (
	"context"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// taxonomyProfileGetter is the narrow read slice of the taxonomy profile
// repository that controlled-documents consumes. The canonical taxonomy
// repository satisfies it; depending on the slice keeps controlled-documents
// from coupling to taxonomy's write surface (interface segregation).
type taxonomyProfileGetter interface {
	GetByCode(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error)
}

// taxonomyAreaGetter is the narrow read slice of the taxonomy area repository
// that controlled-documents consumes.
type taxonomyAreaGetter interface {
	GetByCode(ctx context.Context, tenantID string, code taxonomydomain.AreaCode) (*taxonomydomain.ProcessArea, error)
}

// TaxonomyProfileReader adapts the canonical taxonomy profile repository to the
// controlled-documents ProfileReader port. It exists so controlled-documents
// reads profiles through the taxonomy reader that enforces the authz GUC and
// the CapTaxonomyView capability (and returns the full row, including alias),
// instead of re-implementing the SQL and silently skipping authorization.
type TaxonomyProfileReader struct {
	profiles taxonomyProfileGetter
}

// NewTaxonomyProfileReader builds a TaxonomyProfileReader backed by
// profiles (the canonical taxonomy profile repository).
func NewTaxonomyProfileReader(profiles taxonomyProfileGetter) *TaxonomyProfileReader {
	return &TaxonomyProfileReader{profiles: profiles}
}

// GetByCode satisfies application.ProfileReader by delegating to the
// wrapped taxonomy profile repository.
func (r *TaxonomyProfileReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	return r.profiles.GetByCode(ctx, tenantID, taxonomydomain.ProfileCode(code))
}

// TaxonomyAreaReader adapts the canonical taxonomy area repository to the
// controlled-documents AreaReader port. See TaxonomyProfileReader for the
// rationale.
type TaxonomyAreaReader struct {
	areas taxonomyAreaGetter
}

// NewTaxonomyAreaReader builds a TaxonomyAreaReader backed by areas (the
// canonical taxonomy area repository).
func NewTaxonomyAreaReader(areas taxonomyAreaGetter) *TaxonomyAreaReader {
	return &TaxonomyAreaReader{areas: areas}
}

// GetByCode satisfies application.AreaReader by delegating to the
// wrapped taxonomy area repository.
func (r *TaxonomyAreaReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error) {
	return r.areas.GetByCode(ctx, tenantID, taxonomydomain.AreaCode(code))
}
