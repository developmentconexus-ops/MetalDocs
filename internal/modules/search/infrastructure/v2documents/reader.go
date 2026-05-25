package v2documents

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	searchdomain "metaldocs/internal/modules/search/domain"
)

type Reader struct {
	db *sql.DB
}

func NewReader(db *sql.DB) *Reader {
	return &Reader{db: db}
}

func (r *Reader) ListDocuments(ctx context.Context, query searchdomain.Query, limit int) ([]searchdomain.Document, error) {
	const q = `
SELECT
	d.id,
	d.name,
	COALESCE(d.status, ''),
	COALESCE(d.profile_code_snapshot, ''),
	COALESCE(d.process_area_code_snapshot, ''),
	d.created_by::text,
	COALESCE(cd.code, ''),
	COALESCE(cd.sequence_num, d.revision_number, 0),
	d.created_at
FROM public.documents d
LEFT JOIN controlled_documents cd ON cd.id = d.controlled_document_id
WHERE d.tenant_id = $1
  AND d.archived_at IS NULL
  AND ($2 = '' OR UPPER(COALESCE(d.status, '')) = $2)
  AND ($3 = '' OR LOWER(COALESCE(d.profile_code_snapshot, '')) = $3)
  AND ($4 = '' OR LOWER(COALESCE(d.process_area_code_snapshot, '')) = $4)
  AND ($5 = '' OR d.created_by::text = $5)
ORDER BY d.created_at DESC
LIMIT $6
`
	// DocumentType and DocumentProfile both map to profile_code_snapshot; prefer whichever is set.
	profileFilter := strings.ToLower(strings.TrimSpace(query.DocumentType))
	if profileFilter == "" {
		profileFilter = strings.ToLower(strings.TrimSpace(query.DocumentProfile))
	}
	rows, err := r.db.QueryContext(
		ctx,
		q,
		query.TenantID,
		strings.ToUpper(strings.TrimSpace(string(query.Status))),
		profileFilter,
		strings.ToLower(strings.TrimSpace(query.ProcessArea)),
		strings.TrimSpace(query.OwnerID),
		limit,
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
		if err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&status,
			&profile,
			&doc.ProcessArea,
			&doc.OwnerID,
			&doc.DocumentCode,
			&doc.DocumentSequence,
			&doc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("v2 scan document: %w", err)
		}
		doc.Status = status
		doc.DocumentProfile = profile
		doc.DocumentType = doc.DocumentProfile
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
