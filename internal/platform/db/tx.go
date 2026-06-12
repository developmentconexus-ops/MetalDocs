// Package db provides driver-agnostic SQL execution interfaces.
//
// Two interfaces, intentionally narrow:
//   - DB is for code that may run outside a transaction.
//   - Tx is for code that MUST run inside a transaction.
//
// Today both interfaces are structurally identical; the names document
// intent and let Tx grow independently (savepoints) without breaking DB
// consumers. *sql.DB satisfies DB; *sql.Tx satisfies both DB and Tx via
// Go structural typing — no adapters needed.
package db

import (
	"context"
	"database/sql"
)

// DB runs SQL outside a transaction. Implementations: *sql.DB, *sql.Tx.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Tx runs SQL inside a transaction. A nil Tx is never valid; constructors
// and ports MUST reject it explicitly. Implementations: *sql.Tx.
type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
