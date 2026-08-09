// Package postgres implements authdomain.Repository against Postgres:
// identities, sessions, and login-lock serialization via
// pg_advisory_xact_lock. It resolves tenant membership through the
// iam-owned UserTenantReader port rather than reading iam_user_roles
// directly (module-boundary remediation, ADR 0031).
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// Compile-time assertion that the postgres adapter satisfies the auth port.
var _ authdomain.Repository = (*Repository)(nil)

// Repository is the Postgres-backed authdomain.Repository implementation.
type Repository struct {
	db *sql.DB
	// userTenants is the iam-owned read port for a user's tenant memberships.
	// auth does NOT own metaldocs.iam_user_roles; it resolves a user's tenant set
	// through this port instead of reading that table directly (H-G remediation,
	// M5/F5.2; see ADR 0031). It reads the pool (off-tx, H-PRE-1).
	userTenants iamdomain.UserTenantReader
}

// NewRepository constructs the postgres auth adapter. userTenants is the iam port
// backing GetUserTenants; a nil value installs the Noop null-object (callers that
// never resolve tenant membership — e.g. unit tests of other methods).
func NewRepository(db *sql.DB, userTenants iamdomain.UserTenantReader) *Repository {
	if userTenants == nil {
		userTenants = iamdomain.NoopUserTenantReader{}
	}
	return &Repository{db: db, userTenants: userTenants}
}

// BeginTx starts a transaction on the underlying pool; callers own its lifecycle.
func (r *Repository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// FindIdentityByIdentifier implements authdomain.Repository, matching
// identifier case-insensitively against username first, then email.
func (r *Repository) FindIdentityByIdentifier(ctx context.Context, identifier string) (authdomain.Identity, error) {
	needle := strings.ToLower(strings.TrimSpace(identifier))
	const q = `
SELECT i.user_id, i.username, COALESCE(i.email, ''), i.display_name, i.password_hash, i.password_algo,
       i.must_change_password, i.last_login_at, i.failed_login_attempts, i.locked_until, i.is_active,
       i.created_at, i.updated_at
FROM metaldocs.auth_identities i
WHERE lower(i.username) = $1
UNION ALL
SELECT i.user_id, i.username, COALESCE(i.email, ''), i.display_name, i.password_hash, i.password_algo,
       i.must_change_password, i.last_login_at, i.failed_login_attempts, i.locked_until, i.is_active,
       i.created_at, i.updated_at
FROM metaldocs.auth_identities i
WHERE i.email IS NOT NULL
  AND lower(i.email) = $1
  AND lower(i.username) <> $1
LIMIT 1
`
	return r.loadIdentity(ctx, q, needle)
}

// FindIdentityByUserID implements authdomain.Repository.
func (r *Repository) FindIdentityByUserID(ctx context.Context, userID string) (authdomain.Identity, error) {
	const q = `
SELECT i.user_id, i.username, COALESCE(i.email, ''), i.display_name, i.password_hash, i.password_algo,
       i.must_change_password, i.last_login_at, i.failed_login_attempts, i.locked_until, i.is_active,
       i.created_at, i.updated_at
FROM metaldocs.auth_identities i
WHERE i.user_id = $1
`
	return r.loadIdentity(ctx, q, userID)
}

// CreateSession implements authdomain.Repository.
func (r *Repository) CreateSession(ctx context.Context, session authdomain.Session) error {
	const q = `
INSERT INTO metaldocs.auth_sessions (session_id, user_id, tenant_id, created_at, expires_at, revoked_at, ip_address, user_agent, last_seen_at)
VALUES ($1, $2, $3::uuid, $4, $5, NULL, $6, $7, $8)
`
	_, err := r.db.ExecContext(ctx, q, session.SessionID, session.UserID, session.TenantID, session.CreatedAt, session.ExpiresAt, session.IPAddress, session.UserAgent, session.LastSeenAt)
	if err != nil {
		return fmt.Errorf("insert auth session: %w", err)
	}
	return nil
}

// FindSession implements authdomain.Repository. Returns ErrSessionNotFound if
// no row matches sessionID.
func (r *Repository) FindSession(ctx context.Context, sessionID string) (authdomain.Session, error) {
	const q = `
SELECT session_id, user_id, tenant_id::text, created_at, expires_at, revoked_at, COALESCE(ip_address, ''), COALESCE(user_agent, ''), last_seen_at
FROM metaldocs.auth_sessions
WHERE session_id = $1
`
	var session authdomain.Session
	if err := r.db.QueryRowContext(ctx, q, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.TenantID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.IPAddress,
		&session.UserAgent,
		&session.LastSeenAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authdomain.Session{}, authdomain.ErrSessionNotFound
		}
		return authdomain.Session{}, fmt.Errorf("select auth session: %w", err)
	}
	return session, nil
}

// GetUserTenants resolves the user's tenant memberships through the iam-owned
// UserTenantReader port rather than reading metaldocs.iam_user_roles directly
// (module-boundary remediation, M5/F5.2). The port reproduces the prior query
// (distinct, sorted) on the connection pool (off-tx, H-PRE-1).
func (r *Repository) GetUserTenants(ctx context.Context, userID string) ([]string, error) {
	return r.userTenants.UserTenantIDs(ctx, userID)
}

// GetTenantByID implements authdomain.Repository. Returns ErrTenantNotFound
// if no row matches tenantID.
func (r *Repository) GetTenantByID(ctx context.Context, tenantID string) (authdomain.Tenant, error) {
	const q = `
SELECT id::text, name, slug
FROM metaldocs.tenants
WHERE id = $1::uuid
`
	var tenant authdomain.Tenant
	if err := r.db.QueryRowContext(ctx, q, tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authdomain.Tenant{}, authdomain.ErrTenantNotFound
		}
		return authdomain.Tenant{}, fmt.Errorf("select tenant: %w", err)
	}
	return tenant, nil
}

// TouchSession implements authdomain.Repository. The write is skipped
// (no-op) unless at least 30s have elapsed since the session's stored
// last_seen_at, bounding write amplification on hot sessions.
func (r *Repository) TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error {
	// TODO: consider updating only expired/grace-window-stale sessions to reduce write amplification further.
	const q = `
UPDATE metaldocs.auth_sessions
SET last_seen_at = $2
WHERE session_id = $1
  AND last_seen_at < (($2)::timestamptz - INTERVAL '30 seconds')
`
	res, err := r.db.ExecContext(ctx, q, sessionID, seenAt)
	if err != nil {
		return fmt.Errorf("touch auth session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("touch auth session rows affected: %w", err)
	}
	if n == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM metaldocs.auth_sessions WHERE session_id = $1)`, sessionID).Scan(&exists); err != nil {
			return fmt.Errorf("touch auth session exists: %w", err)
		}
		if !exists {
			return authdomain.ErrSessionNotFound
		}
	}
	return nil
}

// RevokeSessionTx is the tx-scoped variant backing RevokeSession; the caller
// owns the transaction's commit/rollback. Returns ErrSessionNotFound if
// sessionID does not exist.
func (r *Repository) RevokeSessionTx(ctx context.Context, tx *sql.Tx, sessionID string, revokedAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_sessions
SET revoked_at = $2
WHERE session_id = $1
`
	res, err := tx.ExecContext(ctx, q, sessionID, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke auth session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke auth session rows affected: %w", err)
	}
	if n == 0 {
		return authdomain.ErrSessionNotFound
	}
	return nil
}

// RevokeSession implements authdomain.Repository by wrapping RevokeSessionTx
// in its own transaction.
func (r *Repository) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin revoke session tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.RevokeSessionTx(ctx, tx, sessionID, revokedAt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoke session tx: %w", err)
	}
	return nil
}

// RevokeSessionsByUserIDTx is the tx-scoped variant backing
// RevokeSessionsByUserID; already-revoked sessions are left untouched
// (idempotent). The caller owns the transaction's commit/rollback.
func (r *Repository) RevokeSessionsByUserIDTx(ctx context.Context, tx *sql.Tx, userID string, revokedAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_sessions
SET revoked_at = $2
WHERE user_id = $1
  AND revoked_at IS NULL
`
	_, err := tx.ExecContext(ctx, q, userID, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke auth sessions by user: %w", err)
	}
	return nil
}

// RevokeSessionsByUserID implements authdomain.Repository by wrapping
// RevokeSessionsByUserIDTx in its own transaction.
func (r *Repository) RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin revoke sessions tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.RevokeSessionsByUserIDTx(ctx, tx, userID, revokedAt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoke sessions tx: %w", err)
	}
	return nil
}

// sqlExecer / sqlRowQuerier are the minimal surfaces shared by *sql.DB and
// *sql.Tx, letting the login-write SQL run either standalone or inside the
// advisory-locked transaction opened by WithinLoginLock — one SQL source, no
// drift between the locked and unlocked code paths.
type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type sqlRowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func recordSuccessfulLogin(ctx context.Context, e sqlExecer, userID string, loginAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_identities
SET last_login_at = $2,
    failed_login_attempts = 0,
    locked_until = NULL,
    updated_at = $2
WHERE user_id = $1
`
	if _, err := e.ExecContext(ctx, q, userID, loginAt); err != nil {
		return fmt.Errorf("update successful login: %w", err)
	}
	return nil
}

func recordFailedLogin(ctx context.Context, q sqlRowQuerier, userID string, maxAttempts int, lockDurationSeconds int, ip string) (int, *time.Time, error) {
	const stmt = `
UPDATE metaldocs.auth_identities
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE WHEN failed_login_attempts + 1 >= $2
                        THEN NOW() + ($3 * INTERVAL '1 second')
                        ELSE locked_until END,
    last_failed_login_at = NOW(),
    last_failed_login_ip = NULLIF($4, ''),
    updated_at = NOW()
WHERE user_id = $1
RETURNING failed_login_attempts, locked_until
`
	var attempts int
	var lockedUntil sql.NullTime
	if err := q.QueryRowContext(ctx, stmt, userID, maxAttempts, lockDurationSeconds, ip).Scan(&attempts, &lockedUntil); err != nil {
		return 0, nil, fmt.Errorf("update failed login: %w", err)
	}
	if !lockedUntil.Valid {
		return attempts, nil, nil
	}
	locked := lockedUntil.Time.UTC()
	return attempts, &locked, nil
}

// RecordSuccessfulLogin implements authdomain.Repository: clears failed-attempt
// count and lockout, and stamps last_login_at.
func (r *Repository) RecordSuccessfulLogin(ctx context.Context, userID string, loginAt time.Time) error {
	return recordSuccessfulLogin(ctx, r.db, userID, loginAt)
}

// loginTx binds the LoginTx critical-section ops to the single transaction that
// holds the advisory lock, so every read/write fn performs is serialized for the
// identity and commits atomically with the lock release.
type loginTx struct {
	tx *sql.Tx
}

func (t *loginTx) LoadLoginState(ctx context.Context, userID string) (authdomain.LoginState, error) {
	const q = `
SELECT password_hash, password_algo, is_active, locked_until
FROM metaldocs.auth_identities
WHERE user_id = $1
`
	var state authdomain.LoginState
	var hash string
	var lockedUntil sql.NullTime
	if err := t.tx.QueryRowContext(ctx, q, userID).Scan(&hash, &state.PasswordAlgo, &state.IsActive, &lockedUntil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authdomain.LoginState{}, authdomain.ErrIdentityNotFound
		}
		return authdomain.LoginState{}, fmt.Errorf("load login state: %w", err)
	}
	state.PasswordHash = authdomain.PasswordHash(hash)
	if lockedUntil.Valid {
		lu := lockedUntil.Time.UTC()
		state.LockedUntil = &lu
	}
	return state, nil
}

func (t *loginTx) RecordFailedLogin(ctx context.Context, userID string, maxAttempts int, lockDurationSeconds int, ip string) (int, *time.Time, error) {
	return recordFailedLogin(ctx, t.tx, userID, maxAttempts, lockDurationSeconds, ip)
}

// RehashPassword implements authdomain.LoginTx: it persists newHash/newAlgo
// for userID inside the login-lock transaction (rehash-on-login migration,
// REQ-AUTHN-1). The caller (Service.Authenticate) treats any returned error
// as non-fatal to the login — see the RehashPassword doc on the LoginTx
// interface.
func (t *loginTx) RehashPassword(ctx context.Context, userID string, newHash authdomain.PasswordHash, newAlgo string) error {
	const q = `
UPDATE metaldocs.auth_identities
SET password_hash = $2,
    password_algo = $3,
    updated_at = NOW()
WHERE user_id = $1
`
	if _, err := t.tx.ExecContext(ctx, q, userID, string(newHash), newAlgo); err != nil {
		return fmt.Errorf("rehash password: %w", err)
	}
	return nil
}

// WithinLoginLock serializes concurrent login attempts for userID via a
// transaction-scoped advisory lock. hashtextextended derives a stable 64-bit
// key from the user_id; pg_advisory_xact_lock holds it until the tx
// commits/rolls back, covering the lockout check and the failed-attempt write
// fn performs.
func (r *Repository) WithinLoginLock(ctx context.Context, userID string, fn func(authdomain.LoginTx) error) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return authdomain.ErrIdentityNotFound
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin login lock tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, userID); err != nil {
		return fmt.Errorf("acquire login lock: %w", err)
	}

	if err := fn(&loginTx{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login lock tx: %w", err)
	}
	return nil
}

// RecordFailedLogin atomically increments failed_login_attempts and, when the
// threshold is reached, sets locked_until. It also stamps last_failed_login_at
// + last_failed_login_ip (PR-7) so /security/lockouts and /security/signals
// can surface the most recent failure context per user.
//
// It is no longer part of authdomain.Repository (removed on this branch — the
// production failed-attempt write goes through LoginTx.RecordFailedLogin inside
// the login lock). It is retained as a concrete helper: the shared
// recordFailedLogin helper backs loginTx.RecordFailedLogin, and tests call this
// method directly via the concrete *Repository type.
func (r *Repository) RecordFailedLogin(ctx context.Context, userID string, maxAttempts int, lockDurationSeconds int, ip string) (int, *time.Time, error) {
	return recordFailedLogin(ctx, r.db, userID, maxAttempts, lockDurationSeconds, ip)
}

// CreateUser implements authdomain.Repository by wrapping CreateUserTx in its
// own transaction.
func (r *Repository) CreateUser(ctx context.Context, params authdomain.CreateUserParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create user tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.CreateUserTx(ctx, tx, params); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create user tx: %w", err)
	}
	return nil
}

// CreateUserTx is the tx-scoped variant backing CreateUser; the caller owns
// the transaction's commit/rollback. Returns ErrUserAlreadyExists on a unique
// constraint violation (username/email/user_id collision).
func (r *Repository) CreateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.CreateUserParams) error {
	const insertIdentity = `
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, NULL, 0, NULL, NOW(), NOW())
`
	if _, err := tx.ExecContext(ctx, insertIdentity, params.UserID, params.Username, params.Email, params.DisplayName, params.IsActive, params.PasswordHash, params.PasswordAlgo, params.MustChangePassword); err != nil {
		if isUniqueViolation(err) {
			return authdomain.ErrUserAlreadyExists
		}
		return fmt.Errorf("insert auth identity: %w", err)
	}
	return nil
}

// ListUsers implements authdomain.Repository, ordered by creation time descending.
func (r *Repository) ListUsers(ctx context.Context) ([]authdomain.ManagedUser, error) {
	const q = `
SELECT i.user_id, i.username, COALESCE(i.email, ''), i.display_name, i.is_active, i.must_change_password,
       i.last_login_at, i.failed_login_attempts, i.locked_until, i.created_at, i.updated_at
FROM metaldocs.auth_identities i
ORDER BY i.created_at DESC
`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list managed users: %w", err)
	}
	defer rows.Close()

	items := make([]authdomain.ManagedUser, 0)
	for rows.Next() {
		var item authdomain.ManagedUser
		if err := rows.Scan(
			&item.UserID,
			&item.Username,
			&item.Email,
			&item.DisplayName,
			&item.IsActive,
			&item.MustChangePassword,
			&item.LastLoginAt,
			&item.FailedLoginAttempts,
			&item.LockedUntil,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan managed user: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate managed users: %w", err)
	}
	return items, nil
}

// ListOnlineUsers implements authdomain.Repository: active, non-revoked,
// non-expired sessions in tenantID seen at or after activeSince, one row per user.
func (r *Repository) ListOnlineUsers(ctx context.Context, tenantID string, activeSince time.Time) ([]authdomain.OnlineUser, error) {
	// TODO: use monotonic heartbeat or session token expiry instead of wall-clock comparison.
	const q = `
SELECT s.user_id, i.username, i.display_name, MAX(s.last_seen_at)
FROM metaldocs.auth_sessions s
JOIN metaldocs.auth_identities i ON i.user_id = s.user_id
WHERE s.tenant_id = $1::uuid
  AND s.revoked_at IS NULL
  AND s.expires_at > NOW()
  AND s.last_seen_at >= $2
  AND i.is_active = true
GROUP BY s.user_id, i.username, i.display_name
ORDER BY MAX(s.last_seen_at) DESC
`
	rows, err := r.db.QueryContext(ctx, q, tenantID, activeSince.UTC())
	if err != nil {
		return nil, fmt.Errorf("list online users: %w", err)
	}
	defer rows.Close()

	items := make([]authdomain.OnlineUser, 0)
	for rows.Next() {
		var item authdomain.OnlineUser
		if err := rows.Scan(
			&item.UserID,
			&item.Username,
			&item.DisplayName,
			&item.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("scan online user: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate online users: %w", err)
	}
	return items, nil
}

// UpdateUserTx is the tx-scoped variant backing UpdateUser; the caller owns
// the transaction's commit/rollback. nil fields in params are left
// unchanged; a no-op params still requires the row to exist. Returns
// ErrIdentityNotFound if no row matches, or ErrUserAlreadyExists on an email
// unique-constraint violation.
func (r *Repository) UpdateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.UpdateUserParams) error {
	if !hasMutableUserUpdates(params) {
		// No writes to do; existence is guaranteed by the surrounding service tx.
		return nil
	}
	updated := false

	if params.DisplayName != nil || params.IsActive != nil {
		res, err := tx.ExecContext(ctx, `
UPDATE metaldocs.auth_identities
SET display_name = COALESCE($2, display_name),
    is_active = COALESCE($3, is_active),
    updated_at = NOW()
WHERE user_id = $1
`, params.UserID, nullableText(params.DisplayName), nullableBool(params.IsActive))
		if err != nil {
			return fmt.Errorf("update auth identity profile: %w", err)
		}
		if err := expectRowsAffected(res, authdomain.ErrIdentityNotFound); err != nil {
			return err
		}
		updated = true
	}

	if params.Email != nil || params.NewPasswordHash != nil || params.MustChangePassword != nil || params.ResetLockState {
		res, err := tx.ExecContext(ctx, `
UPDATE metaldocs.auth_identities
SET email = COALESCE($2, email),
    password_hash = COALESCE($3, password_hash),
    password_algo = CASE WHEN $3 IS NULL THEN password_algo ELSE COALESCE($6, password_algo) END,
    must_change_password = COALESCE($4, must_change_password),
    failed_login_attempts = CASE WHEN $3 IS NOT NULL OR $5 THEN 0 ELSE failed_login_attempts END,
    locked_until = CASE WHEN $3 IS NOT NULL OR $5 THEN NULL ELSE locked_until END,
    updated_at = NOW()
WHERE user_id = $1
`, params.UserID, nullableText(params.Email), nullablePasswordHash(params.NewPasswordHash), nullableBool(params.MustChangePassword), params.ResetLockState, nullableText(params.NewPasswordAlgo))
		if err != nil {
			if isUniqueViolation(err) {
				return authdomain.ErrUserAlreadyExists
			}
			return fmt.Errorf("update auth identity: %w", err)
		}
		if err := expectRowsAffected(res, authdomain.ErrIdentityNotFound); err != nil {
			return err
		}
		updated = true
	}

	if !updated {
		return authdomain.ErrIdentityNotFound
	}
	return nil
}

// UpdateUser implements authdomain.Repository by wrapping UpdateUserTx in its
// own transaction (profile and credential updates commit atomically together).
func (r *Repository) UpdateUser(ctx context.Context, params authdomain.UpdateUserParams) error {
	if !hasMutableUserUpdates(params) {
		return r.ensureIdentityExists(ctx, params.UserID)
	}

	// Keep profile and credential updates in one transaction so mixed update
	// requests stay atomic even though each branch may issue its own statement.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update user tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.UpdateUserTx(ctx, tx, params); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update user tx: %w", err)
	}
	return nil
}

func (r *Repository) ensureIdentityExists(ctx context.Context, userID string) error {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM metaldocs.auth_identities
	WHERE user_id = $1
)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check auth identity exists: %w", err)
	}
	if !exists {
		return authdomain.ErrIdentityNotFound
	}
	return nil
}

// BootstrapAdmin implements authdomain.Repository. Uses ON CONFLICT DO
// NOTHING on user_id, so it is idempotent — created=false (no error) when
// the row already exists.
func (r *Repository) BootstrapAdmin(ctx context.Context, params authdomain.BootstrapAdminParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin bootstrap admin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULLIF($3, ''), $4, TRUE, $5, $6, $7, NULL, 0, NULL, NOW(), NOW())
ON CONFLICT (user_id)
DO NOTHING
`, params.UserID, params.Username, params.Email, params.DisplayName, params.PasswordHash, params.PasswordAlgo, params.MustChangePassword)
	if err != nil {
		if isUniqueViolation(err) {
			return false, authdomain.ErrUserAlreadyExists
		}
		return false, fmt.Errorf("bootstrap auth identity: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("bootstrap auth identity rows affected: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit bootstrap admin tx: %w", err)
	}
	return rows > 0, nil
}

func (r *Repository) loadIdentity(ctx context.Context, query string, arg string) (authdomain.Identity, error) {
	var identity authdomain.Identity
	if err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&identity.UserID,
		&identity.Username,
		&identity.Email,
		&identity.DisplayName,
		&identity.PasswordHash,
		&identity.PasswordAlgo,
		&identity.MustChangePassword,
		&identity.LastLoginAt,
		&identity.FailedLoginAttempts,
		&identity.LockedUntil,
		&identity.IsActive,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authdomain.Identity{}, authdomain.ErrIdentityNotFound
		}
		return authdomain.Identity{}, fmt.Errorf("load auth identity: %w", err)
	}
	return identity, nil
}

func isUniqueViolation(err error) bool {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
		return true
	}
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == "23505"
}

func nullableText(value *string) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(*value)
}

func nullablePasswordHash(value *authdomain.PasswordHash) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(string(*value))
}

func nullableBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func expectRowsAffected(res sql.Result, notFound error) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}
	if rows == 0 {
		return notFound
	}
	return nil
}

func hasMutableUserUpdates(params authdomain.UpdateUserParams) bool {
	return params.DisplayName != nil ||
		params.Email != nil ||
		params.IsActive != nil ||
		params.NewPasswordHash != nil ||
		params.MustChangePassword != nil ||
		params.ResetLockState
}
