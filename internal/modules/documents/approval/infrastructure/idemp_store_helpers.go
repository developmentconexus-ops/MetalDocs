// Package infrastructure holds Postgres-backed idempotency stores for the
// approval subsystem's route-admin and signoff endpoints, wrapping the shared
// internal/platform/idempotency.Store with endpoint-specific replay envelopes.
package infrastructure

import (
	"context"
	"errors"

	"metaldocs/internal/platform/idempotency"
)

// beginReplayRaw performs the shared BeginReplay call and nil-guard logic
// shared by all Postgres idemp stores in this package. It returns the raw
// handle and replay without envelope unmarshalling, leaving callers free to
// decode into their own envelope type.
//
// Return semantics mirror idempotency.Store.BeginReplay:
//   - (nil, replay, nil)  — cache hit; caller MUST replay and skip handler.
//   - (handle, nil, nil)  — slot won; caller MUST run handler then Complete/Fail.
//   - (nil, nil, err)     — DB error or store not configured.
func beginReplayRaw(
	ctx context.Context,
	store *idempotency.Store,
	tenantID, actorID, idempKey, payloadHash string,
) (*idempotency.ReplayHandle, *idempotency.Replay, error) {
	if store == nil {
		return nil, nil, errors.New("idempotency store database not configured")
	}
	handle, replay, err := store.BeginReplay(ctx, tenantID, actorID, idempKey, payloadHash)
	if err != nil || replay == nil {
		if err != nil || handle == nil {
			return nil, nil, err
		}
		return handle, nil, nil
	}
	return handle, replay, nil
}
