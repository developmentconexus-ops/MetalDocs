package infrastructure

import (
	"context"
	"database/sql"
	"errors"
)

const signoffRouteTemplate = "POST /api/v2/documents/{id}/signoff"

type PostgresSignoffIdempStore struct {
	db *sql.DB
}

func NewPostgresSignoffIdempStore(db *sql.DB) *PostgresSignoffIdempStore {
	return &PostgresSignoffIdempStore{db: db}
}

func (s *PostgresSignoffIdempStore) CheckReplay(ctx context.Context, tenantID, actorID, idempKey string) (bool, string, error) {
	if s.db == nil {
		return false, "", errors.New("idempotency store database not configured")
	}

	var outcome string
	err := s.db.QueryRowContext(ctx, `
		SELECT response_body->>'outcome'
		  FROM metaldocs.idempotency_keys
		 WHERE tenant_id = $1::uuid
		   AND actor_user_id = $2
		   AND route_template = $3
		   AND key = $4
		   AND status = 'completed'
		   AND expires_at > now()`,
		tenantID, actorID, signoffRouteTemplate, idempKey,
	).Scan(&outcome)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, outcome, nil
}

func (s *PostgresSignoffIdempStore) RecordReplay(ctx context.Context, tenantID, actorID, idempKey string, outcome string) error {
	if s.db == nil {
		return errors.New("idempotency store database not configured")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO metaldocs.idempotency_keys
			(tenant_id, actor_user_id, route_template, key, payload_hash, response_status, response_body, status, expires_at)
		VALUES
			($1::uuid, $2, $3, $4, '', 200, jsonb_build_object('outcome', $5::text), 'completed', now() + interval '24 hours')
		ON CONFLICT (tenant_id, actor_user_id, route_template, key)
		DO UPDATE SET
			response_status = 200,
			response_body = jsonb_build_object('outcome', $5::text),
			status = 'completed',
			expires_at = now() + interval '24 hours'`,
		tenantID, actorID, signoffRouteTemplate, idempKey, outcome,
	)
	return err
}
