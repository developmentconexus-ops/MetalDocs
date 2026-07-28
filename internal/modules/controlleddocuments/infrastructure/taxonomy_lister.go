package infrastructure

import (
	"context"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// taxonomyProfileListing is the narrow catalog-read slice of the taxonomy
// profile repository that controlled-documents consumes for the creation
// context. Kept separate from taxonomyProfileGetter (interface segregation) so
// the per-code lookup adapter and its fakes are untouched.
type taxonomyProfileListing interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]taxonomydomain.DocumentProfile, error)
}

// taxonomyAreaListing is the narrow catalog-read slice of the taxonomy area
// repository that controlled-documents consumes for the creation context.
type taxonomyAreaListing interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]taxonomydomain.ProcessArea, error)
}

// TaxonomyProfileLister adapts the canonical taxonomy profile repository to the
// controlled-documents ProfileLister port. Same rationale as
// TaxonomyProfileReader: controlled-documents reads the profile catalog through
// taxonomy's reader (authz GUC + CapTaxonomyView enforced there) rather than
// re-implementing the SQL.
type TaxonomyProfileLister struct {
	profiles taxonomyProfileListing
}

// NewTaxonomyProfileLister builds a TaxonomyProfileLister backed by profiles
// (the canonical taxonomy profile repository).
func NewTaxonomyProfileLister(profiles taxonomyProfileListing) *TaxonomyProfileLister {
	return &TaxonomyProfileLister{profiles: profiles}
}

// ListProfiles satisfies application.ProfileLister, returning the tenant's
// non-archived profiles.
func (r *TaxonomyProfileLister) ListProfiles(ctx context.Context, tenantID string) ([]taxonomydomain.DocumentProfile, error) {
	return r.profiles.List(ctx, tenantID, false)
}

// TaxonomyAreaLister adapts the canonical taxonomy area repository to the
// controlled-documents AreaLister port.
type TaxonomyAreaLister struct {
	areas taxonomyAreaListing
}

// NewTaxonomyAreaLister builds a TaxonomyAreaLister backed by areas (the
// canonical taxonomy area repository).
func NewTaxonomyAreaLister(areas taxonomyAreaListing) *TaxonomyAreaLister {
	return &TaxonomyAreaLister{areas: areas}
}

// ListAreas satisfies application.AreaLister, returning the tenant's
// non-archived process areas.
func (r *TaxonomyAreaLister) ListAreas(ctx context.Context, tenantID string) ([]taxonomydomain.ProcessArea, error) {
	return r.areas.List(ctx, tenantID, false)
}
