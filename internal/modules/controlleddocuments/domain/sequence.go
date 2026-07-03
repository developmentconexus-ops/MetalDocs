package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// SequenceAllocator issues the per-(tenant, profile, area) numeric sequence
// used by AutoCode. Implementations must serialize concurrent allocation
// (e.g. row lock / atomic increment) so two concurrent creates never
// receive the same number.
type SequenceAllocator interface {
	// NextAndIncrement atomically returns the next sequence number and
	// persists the increment inside tx, so the allocation commits or rolls
	// back with the caller's controlled-document create.
	NextAndIncrement(ctx context.Context, tx db.Tx, tenantID, profileCode, areaCode string) (int, error)
	// Peek returns the current counter value without incrementing it
	// (used for preview endpoints).
	Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
	// EnsureCounter initializes the counter row for (tenantID, profileCode,
	// areaCode) if it does not already exist. Safe to call repeatedly.
	EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error
}
