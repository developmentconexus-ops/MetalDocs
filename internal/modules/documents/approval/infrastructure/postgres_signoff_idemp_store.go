package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"metaldocs/internal/platform/idempotency"
)

const (
	documentSignoffRouteTemplate = "POST /api/v1/documents/{id}/signoff"
	stageSignoffRouteTemplate    = "POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs"
)

type SignoffReplay struct {
	Outcome string `json:"outcome"`
}

// SignoffReplayCommitter is the slot handle a winning caller must resolve with
// exactly one of Complete or Fail. It is returned as an interface so HTTP
// handlers depend on the behaviour, not the Postgres-bound concrete type, which
// keeps the replay seam unit-testable with an in-memory double.
type SignoffReplayCommitter interface {
	Complete(outcome string) error
	Fail(cause error) error
}

type SignoffReplayHandle struct {
	store  *idempotency.Store
	handle *idempotency.ReplayHandle
}

func (h *SignoffReplayHandle) Complete(outcome string) error {
	if h == nil || h.store == nil || h.handle == nil {
		return errors.New("idempotency store not configured")
	}
	body, err := json.Marshal(SignoffReplay{Outcome: outcome})
	if err != nil {
		return fmt.Errorf("signoff idempotency: marshal replay response: %w", err)
	}
	return h.store.CompleteReplay(h.handle, 200, body)
}

func (h *SignoffReplayHandle) Fail(cause error) error {
	if h == nil || h.store == nil || h.handle == nil {
		return nil
	}
	return h.store.FailReplay(h.handle, cause)
}

type PostgresSignoffIdempStore struct {
	document *idempotency.Store
	stage    *idempotency.Store
}

func NewPostgresSignoffIdempStore(db *sql.DB) *PostgresSignoffIdempStore {
	if db == nil {
		return &PostgresSignoffIdempStore{}
	}
	return &PostgresSignoffIdempStore{
		document: idempotency.New(db, documentSignoffRouteTemplate),
		stage:    idempotency.New(db, stageSignoffRouteTemplate),
	}
}

func (s *PostgresSignoffIdempStore) BeginDocumentReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (SignoffReplayCommitter, *SignoffReplay, error) {
	return s.beginReplay(ctx, s.document, tenantID, actorID, idempKey, payloadHash)
}

func (s *PostgresSignoffIdempStore) BeginStageReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (SignoffReplayCommitter, *SignoffReplay, error) {
	return s.beginReplay(ctx, s.stage, tenantID, actorID, idempKey, payloadHash)
}

func (s *PostgresSignoffIdempStore) beginReplay(ctx context.Context, store *idempotency.Store, tenantID, actorID, idempKey, payloadHash string) (SignoffReplayCommitter, *SignoffReplay, error) {
	if store == nil {
		return nil, nil, errors.New("idempotency store database not configured")
	}
	handle, replay, err := store.BeginReplay(ctx, tenantID, actorID, idempKey, payloadHash)
	if err != nil || replay == nil {
		if err != nil || handle == nil {
			return nil, nil, err
		}
		return &SignoffReplayHandle{store: store, handle: handle}, nil, nil
	}
	var envelope SignoffReplay
	if err := json.Unmarshal(replay.Body, &envelope); err != nil {
		return nil, nil, err
	}
	return nil, &envelope, nil
}
