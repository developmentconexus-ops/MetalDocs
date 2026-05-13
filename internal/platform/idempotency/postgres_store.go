package idempotency

import (
	"context"
	"database/sql"
	"errors"
)

// ErrConflict is returned when the same idempotency key is reused with a
// different payload hash, indicating a caller bug or replay attack.
var ErrConflict = errors.New("idempotency: key reused with different payload")

// Replay holds the cached response for a previously completed request.
type Replay struct {
	Status int
	Body   []byte
}

// Store is a Postgres-backed idempotency store scoped to a single route
// template (e.g. "POST /api/v1/documents/{id}/submit").
type Store struct {
	db            *sql.DB
	routeTemplate string
}

// New returns a Store for the given route template.
func New(db *sql.DB, routeTemplate string) *Store {
	return &Store{db: db, routeTemplate: routeTemplate}
}

// CheckReplay looks up a completed idempotency record. It returns:
//   - (non-nil Replay, nil) — cache hit, caller should replay the response.
//   - (nil, nil)            — cache miss, caller should process the request.
//   - (nil, ErrConflict)   — same key with a different payload hash.
//   - (nil, err)            — unexpected DB error.
func (s *Store) CheckReplay(ctx context.Context, tenantID, actorID, key, payloadHash string) (*Replay, error) {
	var (
		status     int
		body       []byte
		storedHash string
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT response_status, response_body, payload_hash
		  FROM metaldocs.idempotency_keys
		 WHERE tenant_id      = $1
		   AND actor_user_id  = $2
		   AND route_template = $3
		   AND key            = $4
		   AND status         = 'completed'
		   AND expires_at     > now()`,
		tenantID, actorID, s.routeTemplate, key,
	).Scan(&status, &body, &storedHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if storedHash != payloadHash {
		return nil, ErrConflict
	}
	return &Replay{Status: status, Body: body}, nil
}

// RecordReplay stores a completed request/response pair. It upserts so that
// re-delivery of the same key (e.g. after a network retry) is idempotent.
func (s *Store) RecordReplay(ctx context.Context, tenantID, actorID, key, payloadHash string, status int, body []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO metaldocs.idempotency_keys
			(tenant_id, actor_user_id, route_template, key, payload_hash, response_status, response_body, status, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, 'completed', now() + interval '24 hours')
		ON CONFLICT (tenant_id, actor_user_id, route_template, key) DO UPDATE SET
			payload_hash    = EXCLUDED.payload_hash,
			response_status = EXCLUDED.response_status,
			response_body   = EXCLUDED.response_body,
			status          = 'completed',
			expires_at      = now() + interval '24 hours'`,
		tenantID, actorID, s.routeTemplate, key, payloadHash, status, body,
	)
	return err
}
