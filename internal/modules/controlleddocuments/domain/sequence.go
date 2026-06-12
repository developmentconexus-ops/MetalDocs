package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

type SequenceAllocator interface {
	NextAndIncrement(ctx context.Context, tx db.Tx, tenantID, profileCode, areaCode string) (int, error)
	Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
	EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error
}
