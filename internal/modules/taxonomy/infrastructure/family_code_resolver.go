package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"metaldocs/internal/modules/taxonomy/domain"
)

// sentinelTenantID is the global-fallback tenant for document profiles and the
// DEFAULT for document_profiles.tenant_id (baseline 0001_current_schema.sql:1130):
// a code owned by this tenant applies to every tenant. Because document_profiles.code
// is a GLOBAL primary key (baseline 0001:2403), a code lives under exactly one
// tenant, so the tenant-vs-sentinel ORDER BY below never actually tie-breaks — it
// is retained verbatim for parity with the search subquery this replaces. Mirrors
// the literal the search v2 reader hard-coded.
const sentinelTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

// FamilyCodeResolverRepository is the taxonomy-owned Postgres implementation of
// domain.FamilyCodeResolver (ADR 0038). It reads metaldocs.document_profiles from
// the connection pool — never from a caller's lock-holding transaction (H-PRE-1)
// — and applies sentinel-tenant precedence without an authz GUC (matching the raw
// subqueries it replaces, which read the table pool-side with no GUC).
type FamilyCodeResolverRepository struct {
	db *sql.DB
}

func NewFamilyCodeResolverRepository(db *sql.DB) *FamilyCodeResolverRepository {
	return &FamilyCodeResolverRepository{db: db}
}

var _ domain.FamilyCodeResolver = (*FamilyCodeResolverRepository)(nil)

// ResolveFamilyCodes resolves each requested code to its precedence-winning
// family (tenant row over sentinel). archived_at is ignored (parity). Codes with
// no matching profile are omitted from the map.
func (r *FamilyCodeResolverRepository) ResolveFamilyCodes(ctx context.Context, tenantID string, codes []domain.ProfileCode) (map[domain.ProfileCode]domain.FamilyCode, error) {
	out := make(map[domain.ProfileCode]domain.FamilyCode, len(codes))
	if len(codes) == 0 {
		return out, nil
	}
	codeStrs := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		s := string(c)
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		codeStrs = append(codeStrs, s)
	}

	const q = `
SELECT DISTINCT ON (code) code, family_code
FROM metaldocs.document_profiles
WHERE tenant_id IN ($1::uuid, $2::uuid)
  AND code = ANY($3)
ORDER BY code, CASE WHEN tenant_id = $1::uuid THEN 0 ELSE 1 END`

	rows, err := r.db.QueryContext(ctx, q, tenantID, sentinelTenantID, pq.Array(codeStrs))
	if err != nil {
		return nil, fmt.Errorf("resolve family codes: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var code, family string
		if err := rows.Scan(&code, &family); err != nil {
			return nil, fmt.Errorf("scan family code: %w", err)
		}
		out[domain.ProfileCode(code)] = domain.FamilyCode(family)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolve family codes rows: %w", err)
	}
	return out, nil
}

// ProfileCodesForFamily returns the profile codes whose precedence-resolved
// family equals family (case-insensitive). Empty family → empty slice.
func (r *FamilyCodeResolverRepository) ProfileCodesForFamily(ctx context.Context, tenantID string, family domain.FamilyCode) ([]domain.ProfileCode, error) {
	if family == "" {
		return []domain.ProfileCode{}, nil
	}
	const q = `
SELECT code FROM (
  SELECT DISTINCT ON (code) code, family_code
  FROM metaldocs.document_profiles
  WHERE tenant_id IN ($1::uuid, $2::uuid)
  ORDER BY code, CASE WHEN tenant_id = $1::uuid THEN 0 ELSE 1 END
) resolved
WHERE LOWER(family_code) = LOWER($3)
ORDER BY code`

	rows, err := r.db.QueryContext(ctx, q, tenantID, sentinelTenantID, string(family))
	if err != nil {
		return nil, fmt.Errorf("profile codes for family: %w", err)
	}
	defer rows.Close()
	out := []domain.ProfileCode{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan profile code: %w", err)
		}
		out = append(out, domain.ProfileCode(code))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("profile codes for family rows: %w", err)
	}
	return out, nil
}
