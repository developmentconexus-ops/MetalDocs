// Package distributioninfra provides the PostgreSQL-backed repository for the
// distribution module. It reads the two published views (v_cd_obligated_readers,
// v_process_area_name) and resolves display names via the iam UserDisplayNameReader
// port (ADR-0029). All reads are non-transactional, pool-backed, with explicit
// WHERE tenant_id = $1 from the caller's auth context. No tx, no advisory locks
// (H-PRE-1). Compliant with ADR-0039 D3a/D4: reads only published views, never
// CD or IAM base tables directly.
package distributioninfra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	distributiondomain "metaldocs/internal/modules/distribution/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/pagination"
)

// nullAreaCursorSentinel is the sort-value token used to encode a NULL area_name
// into a keyset cursor. A NUL byte (\x00) never appears in a taxonomy area label
// (labels are human-readable UTF-8 strings), it survives base64 round-trip
// unchanged, and — critically — it is non-empty so pagination.DecodeCursor
// accepts it (DecodeCursor rejects "" as its first part).  The decode branch
// converts it back to the NULL-bucket SQL path.
const nullAreaCursorSentinel = "\x00"

// CoverageRepository reads distribution data from the published views.
type CoverageRepository struct {
	db    *sql.DB
	names iamdomain.UserDisplayNameReader
}

// NewCoverageRepository builds a ready repository. Both db and names are required.
func NewCoverageRepository(db *sql.DB, names iamdomain.UserDisplayNameReader) *CoverageRepository {
	if db == nil {
		panic("distributioninfra: db is required")
	}
	if names == nil {
		panic("distributioninfra: UserDisplayNameReader is required")
	}
	return &CoverageRepository{db: db, names: names}
}

// Summary returns the distinct obligated-reader count for a controlled document.
// Non-transactional pool read; explicit tenant_id filter.
func (r *CoverageRepository) Summary(ctx context.Context, tenantID, cdID string) (int, error) {
	const q = `
		SELECT COUNT(*)
		  FROM metaldocs.v_cd_obligated_readers
		 WHERE tenant_id = $1::uuid
		   AND controlled_document_id = $2::uuid
	`
	var total int
	if err := r.db.QueryRowContext(ctx, q, tenantID, cdID).Scan(&total); err != nil {
		return 0, fmt.Errorf("distribution.Summary: %w", err)
	}
	return total, nil
}

// Recipients returns a paginated page of obligated readers for a controlled document.
// Keyset cursor is on (area_name NULLS LAST, user_id) — a stable total order
// (user_id is unique per row in the view). Display names are resolved in one batch
// call to the iam port after row fetch (no N+1). An empty cursor string starts the
// first page. A malformed cursor returns pagination.ErrInvalidCursor.
func (r *CoverageRepository) Recipients(ctx context.Context, tenantID, cdID, cursor string, limit int) (distributiondomain.RecipientsPage, error) {
	limit = pagination.ClampLimit(limit)

	var (
		sortValue string // decoded cursor: area_name sentinel or actual value
		cursorID  string // decoded cursor: user_id tiebreaker
	)

	if cursor != "" {
		var err error
		sortValue, cursorID, err = pagination.DecodeCursor(cursor)
		if err != nil {
			return distributiondomain.RecipientsPage{}, pagination.ErrInvalidCursor
		}
	}

	var (
		rows *sql.Rows
		err  error
	)

	// We SELECT limit+1 rows to determine has_more without a separate COUNT.
	// Keyset filter: (area_name NULLS LAST, user_id) > (cursor_area_name, cursor_user_id)
	// using the tuple-comparison pattern that is NULL-safe.
	//
	// area_name NULLs LAST logic:
	//   - user_grant and company_scope rows have area_code=NULL → JOIN produces NULL area_name.
	//   - NULLS LAST means NULLs sort after all non-NULL values.
	//   - Tuple comparison: NULL > non-NULL is treated here via explicit IS NULL checks.
	//
	// Cursor encoding: NULL area_name is stored as nullAreaCursorSentinel ("\x00").
	// This sentinel is non-empty, so pagination.DecodeCursor accepts it (it rejects "").
	// On decode: cursorID=="" → first page; sortValue==nullAreaCursorSentinel → NULL bucket;
	// else → non-NULL area_name bucket.
	switch {
	case cursorID == "":
		// First page: no keyset filter.
		const q = `
			SELECT v.user_id,
			       v.area_code,
			       v.source,
			       pan.area_name
			  FROM metaldocs.v_cd_obligated_readers v
			  LEFT JOIN metaldocs.v_process_area_name pan
			    ON pan.tenant_id = v.tenant_id
			   AND pan.area_code = v.area_code
			 WHERE v.tenant_id = $1::uuid
			   AND v.controlled_document_id = $2::uuid
			 ORDER BY pan.area_name NULLS LAST, v.user_id
			 LIMIT $3
		`
		rows, err = r.db.QueryContext(ctx, q, tenantID, cdID, limit+1)
	case sortValue == nullAreaCursorSentinel:
		// Cursor points to a NULL area_name row: next page is rows with NULL area_name
		// and user_id > cursorID (since NULL NULLS LAST means all non-NULLs came first).
		// There are no non-NULL area_name rows after a NULL area_name row, so we only
		// need to filter within the NULL area_name bucket.
		// sortValue is the routing-only sentinel; it is NOT passed to SQL.
		const q = `
			SELECT v.user_id,
			       v.area_code,
			       v.source,
			       pan.area_name
			  FROM metaldocs.v_cd_obligated_readers v
			  LEFT JOIN metaldocs.v_process_area_name pan
			    ON pan.tenant_id = v.tenant_id
			   AND pan.area_code = v.area_code
			 WHERE v.tenant_id = $1::uuid
			   AND v.controlled_document_id = $2::uuid
			   AND pan.area_name IS NULL
			   AND v.user_id > $3
			 ORDER BY pan.area_name NULLS LAST, v.user_id
			 LIMIT $4
		`
		rows, err = r.db.QueryContext(ctx, q, tenantID, cdID, cursorID, limit+1)
	default:
		// Cursor points to a non-NULL area_name row.
		// Next page: rows where (area_name, user_id) > (sortValue, cursorID).
		// That means:
		//   - area_name > sortValue, OR
		//   - area_name = sortValue AND user_id > cursorID, OR
		//   - area_name IS NULL (NULLS LAST: all NULLs come after all non-NULLs)
		const q = `
			SELECT v.user_id,
			       v.area_code,
			       v.source,
			       pan.area_name
			  FROM metaldocs.v_cd_obligated_readers v
			  LEFT JOIN metaldocs.v_process_area_name pan
			    ON pan.tenant_id = v.tenant_id
			   AND pan.area_code = v.area_code
			 WHERE v.tenant_id = $1::uuid
			   AND v.controlled_document_id = $2::uuid
			   AND (
			     pan.area_name > $3
			     OR (pan.area_name = $3 AND v.user_id > $4)
			     OR pan.area_name IS NULL
			   )
			 ORDER BY pan.area_name NULLS LAST, v.user_id
			 LIMIT $5
		`
		rows, err = r.db.QueryContext(ctx, q, tenantID, cdID, sortValue, cursorID, limit+1)
	}
	if err != nil {
		return distributiondomain.RecipientsPage{}, fmt.Errorf("distribution.Recipients query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type rawRow struct {
		userID   string
		areaCode sql.NullString
		source   string
		areaName sql.NullString
	}

	var raw []rawRow
	for rows.Next() {
		var rr rawRow
		if err := rows.Scan(&rr.userID, &rr.areaCode, &rr.source, &rr.areaName); err != nil {
			return distributiondomain.RecipientsPage{}, fmt.Errorf("distribution.Recipients scan: %w", err)
		}
		if err := validateSource(rr.source); err != nil {
			return distributiondomain.RecipientsPage{}, err
		}
		raw = append(raw, rr)
	}
	if err := rows.Err(); err != nil {
		return distributiondomain.RecipientsPage{}, fmt.Errorf("distribution.Recipients rows: %w", err)
	}

	hasMore := len(raw) > limit
	if hasMore {
		raw = raw[:limit]
	}

	// Resolve display names in one batch call (no N+1).
	userIDs := make([]string, len(raw))
	for i, rr := range raw {
		userIDs[i] = rr.userID
	}
	nameMap, err := r.names.DisplayNames(ctx, tenantID, userIDs)
	if err != nil {
		return distributiondomain.RecipientsPage{}, fmt.Errorf("distribution.Recipients display names: %w", err)
	}

	items := make([]distributiondomain.RecipientRow, len(raw))
	for i, rr := range raw {
		name := nameMap[rr.userID]
		if name == "" {
			name = rr.userID // COALESCE fallback
		}
		var areaCode *string
		if rr.areaCode.Valid {
			v := rr.areaCode.String
			areaCode = &v
		}
		var areaName *string
		if rr.areaName.Valid {
			v := rr.areaName.String
			areaName = &v
		}
		items[i] = distributiondomain.RecipientRow{
			UserID:   rr.userID,
			Name:     name,
			AreaCode: areaCode,
			AreaName: areaName,
			Source:   rr.source,
		}
	}

	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		// Encode (area_name, user_id). NULL area_name → use nullAreaCursorSentinel so
		// DecodeCursor returns ("\x00", userID); the next-page query branches on the NULL
		// bucket.  We must NOT encode "" here — DecodeCursor rejects a token whose first
		// part is empty (cursor.go:64).
		sortVal := nullAreaCursorSentinel
		if last.AreaName != nil {
			sortVal = *last.AreaName
		}
		nextCursor = pagination.EncodeCursor(sortVal, last.UserID)
	}

	return distributiondomain.RecipientsPage{
		Items:      items,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// Coverage returns per-area counts of area_grant obligated readers for a controlled
// document. Returns an empty slice for company-scope documents (no area_grant rows).
// Results are ordered by area_name (ascending). Non-transactional pool read.
func (r *CoverageRepository) Coverage(ctx context.Context, tenantID, cdID string) ([]distributiondomain.AreaCoverageRow, error) {
	const q = `
		SELECT v.area_code,
		       pan.area_name,
		       COUNT(*) AS total
		  FROM metaldocs.v_cd_obligated_readers v
		  JOIN metaldocs.v_process_area_name pan
		    ON pan.tenant_id = v.tenant_id
		   AND pan.area_code = v.area_code
		 WHERE v.tenant_id = $1::uuid
		   AND v.controlled_document_id = $2::uuid
		   AND v.source = 'area_grant'
		 GROUP BY v.area_code, pan.area_name
		 ORDER BY pan.area_name
	`
	rows, err := r.db.QueryContext(ctx, q, tenantID, cdID)
	if err != nil {
		return nil, fmt.Errorf("distribution.Coverage query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []distributiondomain.AreaCoverageRow
	for rows.Next() {
		var row distributiondomain.AreaCoverageRow
		if err := rows.Scan(&row.AreaCode, &row.AreaName, &row.Total); err != nil {
			return nil, fmt.Errorf("distribution.Coverage scan: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("distribution.Coverage rows: %w", err)
	}
	if result == nil {
		result = []distributiondomain.AreaCoverageRow{}
	}
	return result, nil
}

// errUnknownSource is returned when a view row carries an unexpected source value.
var errUnknownSource = errors.New("distribution: unknown source value in obligated readers view")

// validateSource fails-closed on any source value not in the published enum.
func validateSource(s string) error {
	switch s {
	case "user_grant", "area_grant", "company_scope":
		return nil
	default:
		return fmt.Errorf("%w: %q", errUnknownSource, s)
	}
}
