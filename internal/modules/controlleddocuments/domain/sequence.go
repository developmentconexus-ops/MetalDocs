package domain

import (
	"context"
	"database/sql"
)

type DBTX interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, sql string, args ...any) *sql.Row
	ExecContext(ctx context.Context, sql string, args ...any) (sql.Result, error)
}

type SequenceAllocator interface {
	NextAndIncrement(ctx context.Context, tx DBTX, tenantID, profileCode, areaCode string) (int, error)
	Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
	EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error
}
