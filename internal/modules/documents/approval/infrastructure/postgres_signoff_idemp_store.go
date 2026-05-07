package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"metaldocs/internal/platform/idempotency"
)

const signoffRouteTemplate = "POST /api/v2/documents/{id}/signoff"

type PostgresSignoffIdempStore struct {
	inner *idempotency.Store
}

func NewPostgresSignoffIdempStore(db *sql.DB) *PostgresSignoffIdempStore {
	if db == nil {
		return &PostgresSignoffIdempStore{inner: nil}
	}
	return &PostgresSignoffIdempStore{inner: idempotency.New(db, signoffRouteTemplate)}
}

func (s *PostgresSignoffIdempStore) CheckReplay(ctx context.Context, tenantID, actorID, idempKey string) (bool, string, error) {
	if s.inner == nil {
		return false, "", errors.New("idempotency store database not configured")
	}
	replay, err := s.inner.CheckReplay(ctx, tenantID, actorID, idempKey, "")
	if err != nil || replay == nil {
		return false, "", err
	}
	var envelope struct {
		Outcome string `json:"outcome"`
	}
	if err := json.Unmarshal(replay.Body, &envelope); err != nil {
		return false, "", err
	}
	return true, envelope.Outcome, nil
}

func (s *PostgresSignoffIdempStore) RecordReplay(ctx context.Context, tenantID, actorID, idempKey, outcome string) error {
	if s.inner == nil {
		return errors.New("idempotency store database not configured")
	}
	body, _ := json.Marshal(map[string]string{"outcome": outcome})
	return s.inner.RecordReplay(ctx, tenantID, actorID, idempKey, "", 200, body)
}
