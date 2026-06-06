package v2documents

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	searchdomain "metaldocs/internal/modules/search/domain"
)

type Reader struct {
	db *sql.DB
}

func NewReader(db *sql.DB) *Reader {
	return &Reader{db: db}
}

func (r *Reader) ListDocuments(ctx context.Context, query searchdomain.Query, limit, offset int) ([]searchdomain.Document, error) {
	// Only columns that exist on public.documents / public.controlled_documents
	// are referenced. The legacy governance columns (subject, business_unit,
	// classification, tags) live on the decommissioned metaldocs.documents schema,
	// not on public.documents — selecting/filtering them errored at runtime, so
	// they are not part of v2 search. Per-document visibility is enforced against
	// the caller ($13) using the unified model (AD-3).
	const q = `
SELECT
	d.id,
	d.name,
	COALESCE(d.status, ''),
	COALESCE(d.profile_code_snapshot, ''),
	COALESCE((
		SELECT dp.family_code
		FROM metaldocs.document_profiles dp
		WHERE dp.code = COALESCE(d.profile_code_snapshot, cd.profile_code)
		  AND dp.tenant_id IN (d.tenant_id, 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid)
		ORDER BY CASE WHEN dp.tenant_id = d.tenant_id THEN 0 ELSE 1 END
		LIMIT 1
	), ''),
	COALESCE(d.process_area_code_snapshot, ''),
	COALESCE(cd.department_code, ''),
	d.created_by::text,
	COALESCE(cd.code, d.code, ''),
	COALESCE(cd.sequence_num, d.revision_number, 0),
	d.effective_from,
	d.effective_to,
	d.created_at
FROM public.documents d
LEFT JOIN public.controlled_documents cd
  ON cd.id = d.controlled_document_id
 AND cd.tenant_id = d.tenant_id
WHERE d.tenant_id = $1
  AND d.archived_at IS NULL
  AND ($2 = '' OR LOWER(COALESCE(d.name, '')) LIKE '%' || $2 || '%')
  AND ($3 = '' OR UPPER(COALESCE(d.status, '')) = $3)
  AND ($4 = '' OR LOWER(COALESCE(d.profile_code_snapshot, '')) = $4)
  AND ($5 = '' OR LOWER(COALESCE((
		SELECT dp.family_code
		FROM metaldocs.document_profiles dp
		WHERE dp.code = COALESCE(d.profile_code_snapshot, cd.profile_code)
		  AND dp.tenant_id IN (d.tenant_id, 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid)
		ORDER BY CASE WHEN dp.tenant_id = d.tenant_id THEN 0 ELSE 1 END
		LIMIT 1
	), '')) = $5)
  AND ($6 = '' OR LOWER(COALESCE(d.process_area_code_snapshot, '')) = $6)
  AND ($7 = '' OR COALESCE(cd.department_code, '') = $7)
  AND ($8 = '' OR d.created_by::text = $8)
  AND ($9::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to <= $9::timestamptz))
  AND ($10::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to >= $10::timestamptz))
  -- Per-document visibility (unified model, AD-3): a document is visible only
  -- when the caller ($13) can see it. Documents linked to a controlled document
  -- inherit its company/owner/restricted+grant visibility (the same predicate
  -- the controlled-documents list enforces); standalone documents fall back to
  -- creator ownership. No system_admin bypass — matches the CD list semantics.
  AND (
    (d.controlled_document_id IS NULL AND d.created_by::text = $13)
    OR (cd.id IS NOT NULL AND (
         cd.visibility_scope = 'company'
      OR cd.owner_user_id = $13
      OR (cd.visibility_scope = 'restricted' AND (
           EXISTS (
             SELECT 1
               FROM public.controlled_document_area_grants cdag
              WHERE cdag.tenant_id = cd.tenant_id
                AND cdag.controlled_document_id = cd.id
                AND EXISTS (
                  SELECT 1
                    FROM public.user_process_areas upa
                   WHERE upa.tenant_id = cd.tenant_id
                     AND upa.user_id = $13
                     AND upa.area_code = cdag.area_code
                     AND upa.effective_to IS NULL
                )
           )
           OR EXISTS (
             SELECT 1
               FROM public.controlled_document_user_grants cdug
              WHERE cdug.tenant_id = cd.tenant_id
                AND cdug.controlled_document_id = cd.id
                AND cdug.user_id = $13
           )
      ))
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
	rows, err := r.db.QueryContext(
		ctx,
		q,
		query.TenantID,
		strings.ToLower(strings.TrimSpace(query.Text)),
		strings.ToUpper(strings.TrimSpace(string(query.Status))),
		profileFilter,
		strings.ToLower(strings.TrimSpace(query.DocumentFamily)),
		strings.ToLower(strings.TrimSpace(query.ProcessArea)),
		strings.TrimSpace(query.Department),
		strings.TrimSpace(query.OwnerID),
		expiryBefore,
		expiryAfter,
		limit,
		offset,
		strings.TrimSpace(query.ActorUserID),
	)
	if err != nil {
		return nil, fmt.Errorf("v2 list documents: %w", err)
	}
	defer rows.Close()

	var out []searchdomain.Document
	for rows.Next() {
		var doc searchdomain.Document
		var status string
		var profile string
		var family string
		var effectiveAt sql.NullTime
		var expiryAt sql.NullTime
		if err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&status,
			&profile,
			&family,
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
		doc.DocumentFamily = family
		doc.EffectiveAt = cloneNullTime(effectiveAt)
		doc.ExpiryAt = cloneNullTime(expiryAt)
		out = append(out, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("v2 list documents rows: %w", err)
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

