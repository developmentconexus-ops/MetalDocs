package idempotency

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/platform/idempotency"
)

const (
	documentSignoffRouteTemplate = "POST /api/v1/documents/{id}/signoff"
	stageSignoffRouteTemplate    = "POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs"
	// fastForwardRouteTemplate (R5, unit 2.3 G3) is deliberately its OWN
	// template, distinct from stageSignoffRouteTemplate. Fast-forward and
	// stage-signoff are different endpoints with different payload shapes
	// (fast-forward has no `decision` field); sharing a template would let a
	// replayed stage-signoff Idempotency-Key collide with an unrelated
	// fast-forward call for the same (tenant, actor, key). Note review-verdict
	// reuses the stage-signoff template today (a pre-existing quirk) — that is
	// NOT repeated here; fast-forward gets its own.
	fastForwardRouteTemplate = "POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/fast-forward"
)

// SignoffReplayHandle commits or fails an in-flight signoff idempotency slot
// claimed by PostgresSignoffIdempStore. It implements application.SignoffReplayCommitter.
type SignoffReplayHandle struct {
	store  *idempotency.Store
	handle *idempotency.ReplayHandle
}

// Complete persists replay as the replay response for this slot, so a retried
// request with the same Idempotency-Key returns the same outcome AND the same
// approval_signoffs id (F-QA4-7).
func (h *SignoffReplayHandle) Complete(replay application.SignoffReplay) error {
	if h == nil || h.store == nil || h.handle == nil {
		return errors.New("idempotency store not configured")
	}
	body, err := json.Marshal(replay)
	if err != nil {
		return fmt.Errorf("signoff idempotency: marshal replay response: %w", err)
	}
	return h.store.CompleteReplay(h.handle, 200, body)
}

// Fail releases the slot after cause, allowing a subsequent retry with the
// same Idempotency-Key to attempt the handler again rather than replay a failure.
func (h *SignoffReplayHandle) Fail(cause error) error {
	if h == nil || h.store == nil || h.handle == nil {
		return nil
	}
	return h.store.FailReplay(h.handle, cause)
}

// PostgresSignoffIdempStore implements application.SignoffIdempStore for the
// document-level and stage-level signoff endpoints.
type PostgresSignoffIdempStore struct {
	document    *idempotency.Store
	stage       *idempotency.Store
	fastForward *idempotency.Store
}

// NewPostgresSignoffIdempStore constructs a PostgresSignoffIdempStore backed by
// db. A nil db yields a store whose sub-stores are unset; every Begin*Replay
// call then fails closed with an error instead of panicking.
func NewPostgresSignoffIdempStore(db *sql.DB) *PostgresSignoffIdempStore {
	if db == nil {
		// Return a store with nil sub-stores; beginReplay returns an error for each call.
		return &PostgresSignoffIdempStore{}
	}
	return &PostgresSignoffIdempStore{
		document:    idempotency.New(db, documentSignoffRouteTemplate),
		stage:       idempotency.New(db, stageSignoffRouteTemplate),
		fastForward: idempotency.New(db, fastForwardRouteTemplate),
	}
}

// BeginDocumentReplay claims or replays an idempotency slot for the
// document-level signoff endpoint. See beginReplayRaw for the return contract.
func (s *PostgresSignoffIdempStore) BeginDocumentReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	return s.beginReplay(ctx, s.document, tenantID, actorID, idempKey, payloadHash)
}

// BeginStageReplay claims or replays an idempotency slot for the stage-level
// signoff endpoint. See beginReplayRaw for the return contract.
func (s *PostgresSignoffIdempStore) BeginStageReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	return s.beginReplay(ctx, s.stage, tenantID, actorID, idempKey, payloadHash)
}

// BeginFastForwardReplay claims or replays an idempotency slot for the
// fast-forward endpoint, under its own fastForwardRouteTemplate (not the
// stage-signoff template). See beginReplayRaw for the return contract.
func (s *PostgresSignoffIdempStore) BeginFastForwardReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	return s.beginReplay(ctx, s.fastForward, tenantID, actorID, idempKey, payloadHash)
}

func (s *PostgresSignoffIdempStore) beginReplay(ctx context.Context, store *idempotency.Store, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	handle, replay, err := beginReplayRaw(ctx, store, tenantID, actorID, idempKey, payloadHash)
	if err != nil || replay == nil {
		if handle == nil {
			return nil, nil, err
		}
		return &SignoffReplayHandle{store: store, handle: handle}, nil, nil
	}
	var envelope application.SignoffReplay
	if err := json.Unmarshal(replay.Body, &envelope); err != nil {
		return nil, nil, err
	}
	return nil, &envelope, nil
}
