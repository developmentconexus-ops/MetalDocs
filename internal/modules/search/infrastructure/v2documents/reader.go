package v2documents

import (
	"context"
	"database/sql"
	"encoding/json"
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
	COALESCE(d.subject_code, ''),
	COALESCE(d.business_unit, ''),
	COALESCE(cd.department_code, ''),
	COALESCE(d.classification, ''),
	COALESCE(d.tags, '[]'::jsonb)::text,
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
  AND ($7 = '' OR LOWER(COALESCE(d.subject_code, '')) = $7)
  AND ($8 = '' OR COALESCE(d.business_unit, '') = $8)
  AND ($9 = '' OR COALESCE(cd.department_code, '') = $9)
  AND ($10 = '' OR UPPER(COALESCE(d.classification, '')) = $10)
  AND ($11 = '' OR d.created_by::text = $11)
  AND ($12 = '' OR EXISTS (
		SELECT 1
		FROM jsonb_array_elements_text(COALESCE(d.tags, '[]'::jsonb)) AS tag(value)
		WHERE LOWER(tag.value) = $12
  ))
  AND ($13::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to <= $13::timestamptz))
  AND ($14::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to >= $14::timestamptz))
ORDER BY d.created_at DESC, d.id DESC
LIMIT $15 OFFSET $16
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
		strings.ToLower(strings.TrimSpace(query.Subject)),
		strings.TrimSpace(query.BusinessUnit),
		strings.TrimSpace(query.Department),
		strings.ToUpper(strings.TrimSpace(string(query.Classification))),
		strings.TrimSpace(query.OwnerID),
		strings.ToLower(strings.TrimSpace(query.Tag)),
		expiryBefore,
		expiryAfter,
		limit,
		offset,
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
		var classification string
		var tagsJSON string
		var effectiveAt sql.NullTime
		var expiryAt sql.NullTime
		if err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&status,
			&profile,
			&family,
			&doc.ProcessArea,
			&doc.Subject,
			&doc.BusinessUnit,
			&doc.Department,
			&classification,
			&tagsJSON,
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
		doc.Classification = searchdomain.Classification(classification)
		doc.Tags = decodeTags(tagsJSON)
		doc.EffectiveAt = cloneNullTime(effectiveAt)
		doc.ExpiryAt = cloneNullTime(expiryAt)
		out = append(out, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("v2 list documents rows: %w", err)
	}
	return out, nil
}

// ListAccessPolicies returns no policies — v2 documents use open-by-default access.
// The search service treats empty policy list as allow.
func (r *Reader) ListAccessPolicies(_ context.Context, _, _ string) ([]searchdomain.AccessPolicy, error) {
	return nil, nil
}

func cloneNullTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func decodeTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil
	}
	return append([]string(nil), tags...)
}
