package v2documents

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	searchdomain "metaldocs/internal/modules/search/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/sqlescape"
)

type Reader struct {
	db       *sql.DB
	families taxonomydomain.FamilyCodeResolver
}

// NewReader builds the v2 documents reader. families resolves each document's
// family_code from taxonomy through a port (ADR 0038) instead of a raw
// cross-schema subquery against taxonomy's profile table (H-G remediation). A nil
// resolver falls back to the no-op (family projection/filter resolve to empty), so
// callers that never use the family dimension need not wire taxonomy.
func NewReader(db *sql.DB, families taxonomydomain.FamilyCodeResolver) *Reader {
	if families == nil {
		families = taxonomydomain.NoopFamilyCodeResolver{}
	}
	return &Reader{db: db, families: families}
}

func (r *Reader) ListDocuments(ctx context.Context, query searchdomain.Query, limit, offset int) ([]searchdomain.Document, error) {
	// This query reads ONLY published cross-module contracts (ADR-0039 D3a), never a
	// foreign module's base table: documents' metaldocs.v_document_search_facts
	// (projection of public.documents) and controlleddocuments' metaldocs.v_cd_search_facts
	// + metaldocs.v_cd_grantee (CD projection + visibility legs). The legacy governance
	// columns (subject, business_unit, classification, tags) are not part of v2 search.
	// Per-document visibility is enforced against the caller ($13) using the unified
	// model (AD-3); the active-membership semantics (revoked members excluded) are
	// encoded inside v_cd_grantee via iam's v_active_user_areas (effective_to IS NULL,
	// soft-delete model, ADR 0037) — see the visibility predicate below.
	// family_code is no longer resolved by a correlated subquery against taxonomy's
	// profile table (H-G cross-schema read, ADR 0038). Instead:
	//   - projection: column 5 selects the raw join key
	//     COALESCE(d.profile_code_snapshot, cd.profile_code); the FamilyCodeResolver
	//     batch-maps key → family in Go after the scan.
	//   - filter ($5): the caller's family is resolved to its profile-code set up
	//     front (ProfileCodesForFamily) and applied IN SQL via `key = ANY($14)`, so
	//     LIMIT/OFFSET pagination is unchanged. $5 stays only as the active-filter
	//     flag; an empty resolved set ($14) with $5 set excludes every row (no
	//     profile has that family) — same as the old family<>filter subquery.
	const q = `
SELECT
	d.id,
	d.name,
	COALESCE(d.status, ''),
	COALESCE(d.profile_code_snapshot, ''),
	COALESCE(d.profile_code_snapshot, cd.profile_code),
	COALESCE(d.process_area_code_snapshot, ''),
	COALESCE(cd.department_code, ''),
	d.created_by::text,
	COALESCE(cd.code, d.code, ''),
	COALESCE(cd.sequence_num, d.revision_number, 0),
	d.effective_from,
	d.effective_to,
	d.created_at
FROM metaldocs.v_document_search_facts d
LEFT JOIN metaldocs.v_cd_search_facts cd
  ON cd.controlled_document_id = d.controlled_document_id
 AND cd.tenant_id = d.tenant_id
WHERE d.tenant_id = $1
  AND d.archived_at IS NULL
  AND ($2 = '' OR LOWER(COALESCE(d.name, '')) LIKE '%' || $2 || '%' ESCAPE '\')
  AND ($3 = '' OR UPPER(COALESCE(d.status, '')) = $3)
  AND ($4 = '' OR LOWER(COALESCE(d.profile_code_snapshot, '')) = $4)
  AND ($5 = '' OR COALESCE(d.profile_code_snapshot, cd.profile_code) = ANY($14))
  AND ($6 = '' OR LOWER(COALESCE(d.process_area_code_snapshot, '')) = $6)
  AND ($7 = '' OR COALESCE(cd.department_code, '') = $7)
  AND ($8 = '' OR d.created_by::text = $8)
  AND ($9::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to <= $9::timestamptz))
  AND ($10::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to >= $10::timestamptz))
  -- Per-document visibility (unified model, AD-3): a document is visible only
  -- when the caller ($13) can see it. Documents linked to a controlled document
  -- inherit its company/owner/restricted+grant visibility; standalone documents
  -- fall back to creator ownership. No system_admin bypass.
  --
  -- The CD legs are composed from controlleddocuments' PUBLISHED contract
  -- (ADR-0039 D3a, M4/F4.1) instead of CD's base tables: v_cd_search_facts carries
  -- the unbounded scalar legs (is_company replaces the scope-enum literal;
  -- owner_user_id), and v_cd_grantee enumerates the bounded restricted-CD edges
  -- (active area-grant members via iam's v_active_user_areas ∪ direct user-grants).
  -- This set is identical to the former inline base-table predicate (proven by the
  -- F4.1 composed-decision parity test and F4.3's frozen-raw parity test).
  AND (
    (d.controlled_document_id IS NULL AND d.created_by = $13)
    OR (cd.controlled_document_id IS NOT NULL AND (
         cd.is_company
      OR cd.owner_user_id = $13
      OR EXISTS (
           SELECT 1
             FROM metaldocs.v_cd_grantee g
            WHERE g.tenant_id = cd.tenant_id
              AND g.controlled_document_id = cd.controlled_document_id
              AND g.grantee_user_id = $13
         )
    ))
  )
ORDER BY d.created_at DESC, d.id DESC
LIMIT $11 OFFSET $12
`
	// DocumentType and DocumentProfile both map to profile_code_snapshot; prefer whichever is set.
	profileFilter := strings.ToLower(strings.TrimSpace(query.DocumentType))
	if profileFilter == "" {
		profileFilter = strings.ToLower(strings.TrimSpace(query.DocumentProfile))
	}
	var expiryBefore any
	if query.ExpiryBefore != nil {
		expiryBefore = query.ExpiryBefore.UTC()
	}
	var expiryAfter any
	if query.ExpiryAfter != nil {
		expiryAfter = query.ExpiryAfter.UTC()
	}

	// Family filter ($5/$14): resolve the requested family to its profile-code set
	// via the taxonomy port (precedence-aware, case-insensitive) and push it into
	// SQL so pagination is computed on the filtered set. $5 is the lowercased family
	// (active-filter flag, matches the old `LOWER(family)=$5`); $14 carries the codes.
	familyFilter := strings.ToLower(strings.TrimSpace(query.DocumentFamily))
	filterCodeStrs := []string{}
	if familyFilter != "" {
		codes, err := r.families.ProfileCodesForFamily(ctx, strings.TrimSpace(query.TenantID), taxonomydomain.FamilyCode(query.DocumentFamily))
		if err != nil {
			return nil, fmt.Errorf("v2 resolve family filter: %w", err)
		}
		filterCodeStrs = make([]string, 0, len(codes))
		for _, c := range codes {
			filterCodeStrs = append(filterCodeStrs, string(c))
		}
	}

	rows, err := r.db.QueryContext(
		ctx,
		q,
		query.TenantID,
		sqlescape.LikeEscape(strings.ToLower(strings.TrimSpace(query.Text))),
		strings.ToUpper(strings.TrimSpace(string(query.Status))),
		profileFilter,
		familyFilter,
		strings.ToLower(strings.TrimSpace(query.ProcessArea)),
		strings.TrimSpace(query.Department),
		strings.TrimSpace(query.OwnerID),
		expiryBefore,
		expiryAfter,
		limit,
		offset,
		strings.TrimSpace(query.ActorUserID),
		pq.Array(filterCodeStrs),
	)
	if err != nil {
		return nil, fmt.Errorf("v2 list documents: %w", err)
	}
	defer rows.Close()

	var out []searchdomain.Document
	// familyKeys[i] is the profile-code join key for out[i]; resolved to a family in
	// a single batch after the scan (projection path of ADR 0038).
	var familyKeys []string
	for rows.Next() {
		var doc searchdomain.Document
		var status string
		var profile string
		var familyKey sql.NullString
		var effectiveAt sql.NullTime
		var expiryAt sql.NullTime
		if err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&status,
			&profile,
			&familyKey,
			&doc.ProcessArea,
			&doc.Department,
			&doc.OwnerID,
			&doc.DocumentCode,
			&doc.DocumentSequence,
			&effectiveAt,
			&expiryAt,
			&doc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("v2 scan document: %w", err)
		}
		doc.Status = searchdomain.Status(status)
		doc.DocumentProfile = profile
		doc.DocumentType = doc.DocumentProfile
		doc.EffectiveAt = cloneNullTime(effectiveAt)
		doc.ExpiryAt = cloneNullTime(expiryAt)
		out = append(out, doc)
		familyKeys = append(familyKeys, familyKey.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("v2 list documents rows: %w", err)
	}

	// Batch-resolve each row's family_code through the taxonomy port. Keys absent
	// from the map (no matching profile for tenant or sentinel) default to "" —
	// identical to the old `COALESCE(subquery, '')`.
	if len(out) > 0 {
		codes := make([]taxonomydomain.ProfileCode, 0, len(familyKeys))
		for _, k := range familyKeys {
			codes = append(codes, taxonomydomain.ProfileCode(k))
		}
		families, err := r.families.ResolveFamilyCodes(ctx, strings.TrimSpace(query.TenantID), codes)
		if err != nil {
			return nil, fmt.Errorf("v2 resolve family projection: %w", err)
		}
		for i := range out {
			out[i].DocumentFamily = string(families[taxonomydomain.ProfileCode(familyKeys[i])])
		}
	}
	return out, nil
}

func cloneNullTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}
