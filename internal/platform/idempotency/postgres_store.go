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

// idempotencyTTL is the Postgres interval used for both the initial in-flight
// insert and the final completed-row update. Single source of truth for key TTL.
const idempotencyTTL = "24 hours"

// Replay holds the cached response for a previously completed request.
type Replay struct {
	Status int
	Body   []byte
}

// Validate checks that the cached response status is a well-formed HTTP status code.
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
	// Symmetric with the actorID guard above (#90/A3.5): a blank tenantID is
	// not a narrower key, it is a SHARED one every tenant-less caller would
	// collide into (same reasoning idempotency.Require's own docstring
	// already applies to actor absence). This is the persistence-boundary
	// half of the fix — TenantActorFromContext (identity.go) is the
	// application-boundary half. Both must hold: this guard exists so a
	// future caller that bypasses the shared resolver still fails closed
	// here, not because the resolver is expected to be bypassed.
	if tenantID == "" {
		return nil, nil, errors.New("idempotency: tenantID must not be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("idempotency: begin tx: %w", err)
	}

	won, err := s.tryClaimSlot(ctx, tx, tenantID, actorID, key, payloadHash)
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	if won {
		// Winner. Caller owns the handle until Complete/Fail.
		return &ReplayHandle{tx: tx, tenantID: tenantID, actorID: actorID, key: key}, nil, nil
	}

	return s.resolveExistingSlot(ctx, tx, tenantID, actorID, key, payloadHash)
}

// tryClaimSlot attempts to atomically claim the (tenant, actor, route, key)
// in_flight slot via INSERT ... ON CONFLICT DO NOTHING (only the
// row-inserting transaction observes RETURNING). won=true means this
// transaction owns the slot; won=false means the row already existed and the
// caller must fall through to resolveExistingSlot.
func (s *Store) tryClaimSlot(ctx context.Context, tx *sql.Tx, tenantID, actorID, key, payloadHash string) (bool, error) {
	var claimedTenant string
	err := tx.QueryRowContext(ctx, `
		-- TODO(phase11): event_id/idempotency_key are still TEXT-backed in the current schema; tighten column bounds in the pending DB migration.
		INSERT INTO metaldocs.idempotency_keys
			(tenant_id, actor_user_id, route_template, key, payload_hash, status, expires_at)
		VALUES
			($1, $2, $3, $4, $5, 'in_flight', now() + interval '`+idempotencyTTL+`')
		ON CONFLICT (tenant_id, actor_user_id, route_template, key) DO NOTHING
		RETURNING tenant_id::text`,
		tenantID, actorID, s.routeTemplate, key, payloadHash,
	).Scan(&claimedTenant)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, fmt.Errorf("idempotency: insert in_flight: %w", err)
}

// resolveExistingSlot handles the loser path from BeginReplay: the
// (tenant, actor, route, key) row already existed when tryClaimSlot ran. It
// blocks on the winner's row-level lock (FOR UPDATE) until the winner
// commits or rolls back, then reacts to the row's terminal state.
func (s *Store) resolveExistingSlot(ctx context.Context, tx *sql.Tx, tenantID, actorID, key, payloadHash string) (*ReplayHandle, *Replay, error) {
	var (
		status         string
		storedHash     string
		respStatus     sql.NullInt64
		respBody       []byte
		expiresExpired bool
	)
	err := tx.QueryRowContext(ctx, `
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
		return completeReplayFromRow(tx, respStatus, respBody)
	case "in_flight":
		return s.reclaimOrRefuseInFlight(ctx, tx, tenantID, actorID, key, payloadHash, expiresExpired)
	case "failed":
		return s.retryAfterFailedRow(ctx, tx, tenantID, actorID, key, payloadHash)
	default:
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("idempotency: unknown status %q", status)
	}
}

// completeReplayFromRow builds the Replay to hand back to the caller for a
// status='completed' row found by resolveExistingSlot.
func completeReplayFromRow(tx *sql.Tx, respStatus sql.NullInt64, respBody []byte) (*ReplayHandle, *Replay, error) {
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
}

// reclaimOrRefuseInFlight handles a status='in_flight' row found by
// resolveExistingSlot. Reached only if the prior writer crashed without
// rolling back and the connection-level lock has since been released. Treat
// an expired orphan as reclaimable; otherwise refuse and let the janitor sweep.
func (s *Store) reclaimOrRefuseInFlight(ctx context.Context, tx *sql.Tx, tenantID, actorID, key, payloadHash string, expiresExpired bool) (*ReplayHandle, *Replay, error) {
	if !expiresExpired {
		_ = tx.Commit()
		return nil, nil, fmt.Errorf("idempotency: in_flight orphan (key in use)")
	}
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

// retryAfterFailedRow handles a status='failed' row found by
// resolveExistingSlot: the prior attempt marked failed. Delete and retry so
// this caller claims the slot.
func (s *Store) retryAfterFailedRow(ctx context.Context, tx *sql.Tx, tenantID, actorID, key, payloadHash string) (*ReplayHandle, *Replay, error) {
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
		       expires_at      = now() + interval '`+idempotencyTTL+`'
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
