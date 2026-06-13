package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/platform/idempotency"
)

const (
	documentSignoffRouteTemplate = "POST /api/v1/documents/{id}/signoff"
	stageSignoffRouteTemplate    = "POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs"
)

type SignoffReplayHandle struct {
	store  *idempotency.Store
	handle *idempotency.ReplayHandle
}

func (h *SignoffReplayHandle) Complete(outcome string) error {
	if h == nil || h.store == nil || h.handle == nil {
		return errors.New("idempotency store not configured")
	}
	body, err := json.Marshal(application.SignoffReplay{Outcome: outcome})
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
		// Return a store with nil sub-stores; beginReplay returns an error for each call.
		return &PostgresSignoffIdempStore{}
	}
	return &PostgresSignoffIdempStore{
		document: idempotency.New(db, documentSignoffRouteTemplate),
		stage:    idempotency.New(db, stageSignoffRouteTemplate),
	}
}

func (s *PostgresSignoffIdempStore) BeginDocumentReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	return s.beginReplay(ctx, s.document, tenantID, actorID, idempKey, payloadHash)
}

func (s *PostgresSignoffIdempStore) BeginStageReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error) {
	return s.beginReplay(ctx, s.stage, tenantID, actorID, idempKey, payloadHash)
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
