package domain

import "context"

// FamilyCodeResolver is the narrow read surface cross-module consumers use to
// resolve a document profile code to its family code — and the inverse — without
// reaching across the module boundary into metaldocs.document_profiles. Taxonomy
// owns document_profiles; this port is the only sanctioned cross-module path to
// the profile→family mapping (H-G class remediation, sibling of ADR 0031's
// TenantUserReader; see ADR 0038).
//
// Both methods honor the sentinel-tenant global fallback
// ('ffffffff-ffff-ffff-ffff-ffffffffffff'): a code owned by the sentinel resolves
// for any querying tenant, and a code owned by the querying tenant resolves for
// that tenant (tenant_id IN (tenant, sentinel)). The tenant-own "precedence"
// tie-break in the impl is live behavior: document_profiles' primary key is
// (tenant_id, code) (migration 0308), so the same code can legally exist under
// both the querying tenant and the sentinel tenant simultaneously, and the
// ORDER BY tenant-over-sentinel tie-break picks the tenant-owned row. Neither
// method filters archived_at — they reproduce the archived-agnostic semantics of
// the raw subqueries they replace (search's v2 documents reader), NOT the
// active-only semantics of FamilyRepository.HasActiveProfiles.
//
// The implementation reads from the connection pool — never from a caller's
// lock-holding transaction (H-PRE-1).
type FamilyCodeResolver interface {
	// ResolveFamilyCodes returns the precedence-resolved family for each requested
	// profile code. Codes with no matching profile (for the tenant or the sentinel)
	// are absent from the returned map — callers default such codes to "". The
	// returned map is non-nil. Code matching is exact-case.
	ResolveFamilyCodes(ctx context.Context, tenantID string, codes []ProfileCode) (map[ProfileCode]FamilyCode, error)

	// ProfileCodesForFamily returns every profile code whose precedence-resolved
	// family equals family (case-insensitive). An empty family, or no match,
	// returns an empty (non-nil) slice. This is the inverse of ResolveFamilyCodes,
	// used to keep a family filter inside SQL pagination via `code = ANY(...)`.
	ProfileCodesForFamily(ctx context.Context, tenantID string, family FamilyCode) ([]ProfileCode, error)
}

// NoopFamilyCodeResolver is the explicit null-object for callers that do not
// resolve family codes (tests, or paths where it is irrelevant). It satisfies
// FamilyCodeResolver and always returns empty results with a nil error. Mirrors
// NoopTenantUserReader.
type NoopFamilyCodeResolver struct{}

// ResolveFamilyCodes always returns an empty, non-nil map with a nil error.
func (NoopFamilyCodeResolver) ResolveFamilyCodes(_ context.Context, _ string, _ []ProfileCode) (map[ProfileCode]FamilyCode, error) {
	return map[ProfileCode]FamilyCode{}, nil
}

// ProfileCodesForFamily always returns an empty, non-nil slice with a nil error.
func (NoopFamilyCodeResolver) ProfileCodesForFamily(_ context.Context, _ string, _ FamilyCode) ([]ProfileCode, error) {
	return []ProfileCode{}, nil
}
