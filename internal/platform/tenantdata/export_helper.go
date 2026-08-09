package tenantdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// ExportTable runs a generic, explicit-tenant-predicate SELECT against table
// (fully-qualified, e.g. "metaldocs.iam_users") using row_to_json so the
// exported shape always matches the live column set, and returns a
// TableExport. tenantColumn is the column the predicate is written against
// ("tenant_id" for every table except public.approval_signoffs, which
// carries "actor_tenant_id" instead — §0.1).
//
// table and tenantColumn are always compile-time constant literals supplied
// by the calling port (never derived from user input), so building the
// query with fmt.Sprintf here carries no injection surface; the tenant VALUE
// itself is always bound as $1.
func ExportTable(ctx context.Context, db *sql.DB, table, tenantColumn, tenantID string) (TableExport, error) {
	query := fmt.Sprintf( // #nosec G201 -- table/tenantColumn are compile-time constant literals from each module's Port.Tables() or a fixed call-site string (never user input); the tenant VALUE is bound as $1. See the doc comment above.
		`SELECT row_to_json(t) FROM (SELECT * FROM %s WHERE %s = $1::uuid ORDER BY 1) t`,
		table, tenantColumn,
	)
	rows, err := db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return TableExport{}, fmt.Errorf("tenantdata: export %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	out := TableExport{Table: table, Rows: []json.RawMessage{}}
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return TableExport{}, fmt.Errorf("tenantdata: scan %s row: %w", table, err)
		}
		out.Rows = append(out.Rows, raw)
	}
	if err := rows.Err(); err != nil {
		return TableExport{}, fmt.Errorf("tenantdata: iterate %s rows: %w", table, err)
	}
	return out, nil
}

// ExportTableTextTenant is ExportTable for the one table whose tenant column
// is TEXT rather than uuid (metaldocs.audit_events.tenant_id — §0.1). Kept
// as a separate function rather than a cast flag so both call sites stay
// simple compile-time literals.
func ExportTableTextTenant(ctx context.Context, db *sql.DB, table, tenantColumn, tenantID string) (TableExport, error) {
	query := fmt.Sprintf( // #nosec G201 -- table/tenantColumn are compile-time constant literals from each module's Port.Tables() or a fixed call-site string (never user input); the tenant VALUE is bound as $1. See the doc comment above.
		`SELECT row_to_json(t) FROM (SELECT * FROM %s WHERE %s = $1 ORDER BY 1) t`,
		table, tenantColumn,
	)
	rows, err := db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return TableExport{}, fmt.Errorf("tenantdata: export %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	out := TableExport{Table: table, Rows: []json.RawMessage{}}
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return TableExport{}, fmt.Errorf("tenantdata: scan %s row: %w", table, err)
		}
		out.Rows = append(out.Rows, raw)
	}
	if err := rows.Err(); err != nil {
		return TableExport{}, fmt.Errorf("tenantdata: iterate %s rows: %w", table, err)
	}
	return out, nil
}

// EraseTable runs a generic, explicit-tenant-predicate DELETE against table
// inside tx and returns the number of rows removed. Same constant-literal
// safety note as ExportTable applies to table/tenantColumn.
func EraseTable(ctx context.Context, tx *sql.Tx, table, tenantColumn, tenantID string) (int64, error) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1::uuid`, table, tenantColumn) // #nosec G201 -- table/tenantColumn are compile-time constant literals from each module's Port.Tables() or a fixed call-site string (never user input); the tenant VALUE is bound as $1. See the doc comment above.
	res, err := tx.ExecContext(ctx, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("tenantdata: erase %s: %w", table, err)
	}
	return res.RowsAffected()
}

// EraseTableTextTenant is EraseTable for TEXT-typed tenant columns
// (metaldocs.audit_events — though audit's own port never calls this; kept
// for symmetry and any future text-tenant table).
func EraseTableTextTenant(ctx context.Context, tx *sql.Tx, table, tenantColumn, tenantID string) (int64, error) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, table, tenantColumn) // #nosec G201 -- table/tenantColumn are compile-time constant literals from each module's Port.Tables() or a fixed call-site string (never user input); the tenant VALUE is bound as $1. See the doc comment above.
	res, err := tx.ExecContext(ctx, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("tenantdata: erase %s: %w", table, err)
	}
	return res.RowsAffected()
}
