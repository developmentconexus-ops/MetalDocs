package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrConflict is returned when the same idempotency key is reused with a
// different payload hash, indicating a caller bug or replay attack.
var ErrConflict = errors.New("idempotency: key reused with different payload")

// maxBodyBytes is the maximum response body size that the store will persist.
// Matches the CHECK constraint added in migration 0204. Responses larger than
// this are refused by CompleteReplay and RecordReplay with an explicit error;
// callers should log the rejection and proceed — idempotency is lost for that
// request, but the handler result was still delivered to the client.
const maxBodyBytes = 64 * 1024

// Replay holds the cached response for a previously completed request.
type Replay struct {
	Status int
	Body   []byte
}

func (r Replay) Validate() error {
	if r.Status < 100 || r.Status > 599 {
		return fmt.Errorf("idempotency: corrupt response_status %d", r.Status)
	}
	return nil
}

// ReplayHandle is the transaction-scoped token returned by BeginReplay when
// the calling goroutine wins the race to insert an `in_flight` row for the
// (tenant, actor, route, key) tuple. The transaction holds a row-level lock
// for the lifetime of the handler; callers MUST call exactly one of
// CompleteReplay or FailReplay to commit or roll back.
type ReplayHandle struct {
	tx       *sql.Tx
	tenantID string
	actorID  string
	key      string
	closed   bool
}

// Store is a Postgres-backed idempotency store scoped to a single route
// template (e.g. "POST /api/v1/documents/{id}/submit").
type Store struct {
	db            *sql.DB
	routeTemplate string
}

// TODO(phase11): add a tenant_id FK in the idempotency_keys schema once the DB migration is scheduled; current tenant isolation is enforced by application scoping and the composite PK.

// New returns a Store for the given route template.
func New(db *sql.DB, routeTemplate string) *Store {
	return &Store{db: db, routeTemplate: routeTemplate}
}

// BeginReplay atomically claims the (tenant, actor, route, key) slot for the
// caller. Outcomes:
//   - (nil, replay, nil) — cache hit on a completed prior request; caller MUST
//     write `replay` to the client and skip the handler.
//   - (handle, nil, nil) — caller owns the slot; MUST run the handler and then
//     call CompleteReplay or FailReplay on the returned handle.
//   - (nil, nil, ErrConflict) — same key reused with a different payload hash.
//   - (nil, nil, err) — unexpected DB error.
//
// Concurrency: a concurrent same-key request is serialized via the row-level
// `SELECT ... FOR UPDATE` against the in-flight row. The loser blocks until
// the winner's transaction commits (CompleteReplay → cache hit) or rolls back
// (FailReplay → loser re-tries the claim).
func (s *Store) BeginReplay(ctx context.Context, tenantID, actorID, key, payloadHash string) (*ReplayHandle, *Replay, error) {
	if actorID == "" {
		return nil, nil, errors.New("idempotency: actorID must not be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("idempotency: begin tx: %w", err)
	}

	// Step 1: try to claim the slot. ON CONFLICT DO NOTHING means only the
	// row-inserting transaction observes RETURNING.
	var claimedTenant string
	err = tx.QueryRowContext(ctx, `
		-- TODO(phase11): event_id/idempotency_key are still TEXT-backed in the current schema; tighten column bounds in the pending DB migration.
		INSERT INTO metaldocs.idempotency_keys
			(tenant_id, actor_user_id, route_template, key, payload_hash, status, expires_at)
		VALUES
			($1, $2, $3, $4, $5, 'in_flight', now() + interval '24 hours')
		ON CONFLICT (tenant_id, actor_user_id, route_template, key) DO NOTHING
		RETURNING tenant_id::text`,
		tenantID, actorID, s.routeTemplate, key, payloadHash,
	).Scan(&claimedTenant)
	if err == nil {
		// Winner. Caller owns the handle until Complete/Fail.
		return &ReplayHandle{tx: tx, tenantID: tenantID, actorID: actorID, key: key}, nil, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("idempotency: insert in_flight: %w", err)
	}

	// Step 2: loser. Row exists; block on the writer's row-level lock until
	// the winner commits or rolls back, then re-read state.
	var (
		status         string
		storedHash     string
		respStatus     sql.NullInt64
		respBody       []byte
		expiresExpired bool
	)
	err = tx.QueryRowContext(ctx, `
		SELECT status, payload_hash, response_status, response_body, expires_at <= now()
		  FROM metaldocs.idempotency_keys
		 WHERE tenant_id      = $1
		   AND actor_user_id  = $2
		   AND route_template = $3
		   AND key            = $4
		 FOR UPDATE`,
		tenantID, actorID, s.routeTemplate, key,
	).Scan(&status, &storedHash, &respStatus, &respBody, &expiresExpired)
	if errors.Is(err, sql.ErrNoRows) {
		// Winner rolled back / deleted the row. Loser becomes the new winner.
		_ = tx.Commit()
		return s.BeginReplay(ctx, tenantID, actorID, key, payloadHash)
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("idempotency: select for update: %w", err)
	}

	if storedHash != payloadHash {
		_ = tx.Commit()
		return nil, nil, ErrConflict
	}

	switch status {
	case "completed":
		if !respStatus.Valid {
			_ = tx.Rollback()
			return nil, nil, fmt.Errorf("idempotency: completed row missing response_status")
		}
		replay := Replay{Status: int(respStatus.Int64), Body: respBody}
		if err := replay.Validate(); err != nil {
			_ = tx.Rollback()
			return nil, nil, err
		}
		_ = tx.Commit()
		return nil, &replay, nil

	case "in_flight":
		// Reached only if the prior writer crashed without rolling back and
		// the connection-level lock has since been released. Treat an expired
		// orphan as reclaimable; otherwise refuse and let the janitor sweep.
		if expiresExpired {
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM metaldocs.idempotency_keys
				 WHERE tenant_id      = $1
				   AND actor_user_id  = $2
				   AND route_template = $3
				   AND key            = $4
				   AND status         = 'in_flight'`,
				tenantID, actorID, s.routeTemplate, key,
			); err != nil {
				_ = tx.Rollback()
				return nil, nil, fmt.Errorf("idempotency: reclaim expired in_flight: %w", err)
			}
			if err := tx.Commit(); err != nil {
				return nil, nil, fmt.Errorf("idempotency: commit reclaim: %w", err)
			}
			return s.BeginReplay(ctx, tenantID, actorID, key, payloadHash)
		}
		_ = tx.Commit()
		return nil, nil, fmt.Errorf("idempotency: in_flight orphan (key in use)")

	case "failed":
		// Prior attempt marked failed. Delete and retry so this caller claims.
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM metaldocs.idempotency_keys
			 WHERE tenant_id      = $1
			   AND actor_user_id  = $2
			   AND route_template = $3
			   AND key            = $4
			   AND status         = 'failed'`,
			tenantID, actorID, s.routeTemplate, key,
		); err != nil {
			_ = tx.Rollback()
			return nil, nil, fmt.Errorf("idempotency: clear failed row: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, nil, fmt.Errorf("idempotency: commit clear failed: %w", err)
		}
		return s.BeginReplay(ctx, tenantID, actorID, key, payloadHash)

	default:
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("idempotency: unknown status %q", status)
	}
}

// CompleteReplay finalizes the slot owned by `handle` with the given response
// and commits the transaction. The UPDATE is gated on `status = 'in_flight'`
// so a completed row can never be silently overwritten (M2 belt-and-suspenders).
func (s *Store) CompleteReplay(handle *ReplayHandle, status int, body []byte) error {
	if handle == nil {
		return errors.New("idempotency: nil handle")
	}
	if handle.closed {
		return errors.New("idempotency: handle already closed")
	}
	if status < 100 || status > 599 {
		_ = handle.tx.Rollback()
		handle.closed = true
		return fmt.Errorf("idempotency: refusing to record invalid status %d", status)
	}
	if len(body) > maxBodyBytes {
		_ = handle.tx.Rollback()
		handle.closed = true
		return fmt.Errorf("idempotency: response body %d bytes exceeds %d-byte cap; idempotency not recorded", len(body), maxBodyBytes)
	}
	handle.closed = true
	res, err := handle.tx.Exec(`
		UPDATE metaldocs.idempotency_keys
		   SET status          = 'completed',
		       response_status = $1,
		       response_body   = $2,
		       expires_at      = now() + interval '24 hours'
		 WHERE tenant_id      = $3
		   AND actor_user_id  = $4
		   AND route_template = $5
		   AND key            = $6
		   AND status         = 'in_flight'`,
		status, body, handle.tenantID, handle.actorID, s.routeTemplate, handle.key,
	)
	if err != nil {
		_ = handle.tx.Rollback()
		return fmt.Errorf("idempotency: complete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		_ = handle.tx.Rollback()
		return fmt.Errorf("idempotency: complete: expected 1 row, got %d", n)
	}
	if err := handle.tx.Commit(); err != nil {
		return fmt.Errorf("idempotency: complete commit: %w", err)
	}
	return nil
}

// FailReplay rolls back the slot owned by `handle` by deleting the in_flight
// row, freeing the key so the next retry can re-execute. Call from `defer`
// on panic or non-2xx response.
//
// `cause` is logged-context only — the row is removed unconditionally so a
// crashed handler does not wedge the key for the full TTL.
func (s *Store) FailReplay(handle *ReplayHandle, _ error) error {
	if handle == nil {
		return errors.New("idempotency: nil handle")
	}
	if handle.closed {
		return nil
	}
	handle.closed = true
	if _, err := handle.tx.Exec(`
		DELETE FROM metaldocs.idempotency_keys
		 WHERE tenant_id      = $1
		   AND actor_user_id  = $2
		   AND route_template = $3
		   AND key            = $4
		   AND status         = 'in_flight'`,
		handle.tenantID, handle.actorID, s.routeTemplate, handle.key,
	); err != nil {
		_ = handle.tx.Rollback()
		return fmt.Errorf("idempotency: fail: %w", err)
	}
	if err := handle.tx.Commit(); err != nil {
		return fmt.Errorf("idempotency: fail commit: %w", err)
	}
	return nil
}

// CheckReplay is the legacy single-statement read kept only for callers that
// have not yet migrated to BeginReplay. It is race-prone (two concurrent
// same-key requests both miss) and SHOULD NOT be used for new code.
//
// Deprecated: use BeginReplay/CompleteReplay/FailReplay for race-free behavior.
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

// RecordReplay is the legacy completed-row write kept for callers that have
// not yet migrated to BeginReplay/CompleteReplay. The ON CONFLICT clause now
// refuses to overwrite a row already marked completed (M2 belt-and-suspenders)
// so a stale concurrent writer cannot stomp a good payload_hash.
// Rewriting the payload for rows still transitioning from `in_flight`/`failed`
// to `completed` is safe because the idempotency contract is keyed by the same
// tenant/actor/route/key/payload-hash tuple.
//
// Deprecated: use BeginReplay/CompleteReplay/FailReplay for race-free behavior.
func (s *Store) RecordReplay(ctx context.Context, tenantID, actorID, key, payloadHash string, status int, body []byte) error {
	if status < 100 || status > 599 {
		return fmt.Errorf("idempotency: refusing to record invalid status %d", status)
	}
	if len(body) > maxBodyBytes {
		return fmt.Errorf("idempotency: response body %d bytes exceeds %d-byte cap; idempotency not recorded", len(body), maxBodyBytes)
	}
	_, err := s.db.ExecContext(ctx, `
		-- TODO(phase11): review idempotency janitor/index ordering together when the DB migration is scheduled; current cleanup still relies on expires_at pruning.
		INSERT INTO metaldocs.idempotency_keys
			(tenant_id, actor_user_id, route_template, key, payload_hash, response_status, response_body, status, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, 'completed', now() + interval '24 hours')
		ON CONFLICT (tenant_id, actor_user_id, route_template, key) DO UPDATE SET
			payload_hash    = EXCLUDED.payload_hash,
			response_status = EXCLUDED.response_status,
			response_body   = EXCLUDED.response_body,
			status          = 'completed',
			expires_at      = now() + interval '24 hours'
		 WHERE metaldocs.idempotency_keys.status <> 'completed'`,
		tenantID, actorID, s.routeTemplate, key, payloadHash, status, body,
	)
	return err
}
