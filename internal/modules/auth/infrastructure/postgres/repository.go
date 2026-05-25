package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	authdomain "metaldocs/internal/modules/auth/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

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

func (r *Repository) CreateSession(ctx context.Context, session authdomain.Session) error {
	const q = `
INSERT INTO metaldocs.auth_sessions (session_id, user_id, tenant_id, created_at, expires_at, revoked_at, ip_address, user_agent, last_seen_at)
VALUES ($1, $2, $3, $4, $5, NULL, $6, $7, $8)
`
	_, err := r.db.ExecContext(ctx, q, session.SessionID, session.UserID, session.TenantID, session.CreatedAt, session.ExpiresAt, session.IPAddress, session.UserAgent, session.LastSeenAt)
	if err != nil {
		return fmt.Errorf("insert auth session: %w", err)
	}
	return nil
}

func (r *Repository) FindSession(ctx context.Context, sessionID string) (authdomain.Session, error) {
	const q = `
SELECT session_id, user_id, COALESCE(tenant_id, ''), created_at, expires_at, revoked_at, COALESCE(ip_address, ''), COALESCE(user_agent, ''), last_seen_at
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

func (r *Repository) GetUserTenants(ctx context.Context, userID string) ([]string, error) {
	const q = `
SELECT DISTINCT tenant_id
FROM metaldocs.iam_user_roles
WHERE user_id = $1
ORDER BY tenant_id
`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get user tenants: %w", err)
	}
	defer rows.Close()
	var tenants []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, fmt.Errorf("scan user tenant: %w", err)
		}
		tenants = append(tenants, tid)
	}
	return tenants, rows.Err()
}

func (r *Repository) TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_sessions
SET last_seen_at = $2
WHERE session_id = $1
  AND last_seen_at < $2 - INTERVAL '30 seconds'
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

func (r *Repository) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_sessions
SET revoked_at = $2
WHERE session_id = $1
`
	res, err := r.db.ExecContext(ctx, q, sessionID, revokedAt)
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

func (r *Repository) RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_sessions
SET revoked_at = $2
WHERE user_id = $1
  AND revoked_at IS NULL
`
	_, err := r.db.ExecContext(ctx, q, userID, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke auth sessions by user: %w", err)
	}
	return nil
}

func (r *Repository) RecordSuccessfulLogin(ctx context.Context, userID string, loginAt time.Time) error {
	const q = `
UPDATE metaldocs.auth_identities
SET last_login_at = $2,
    failed_login_attempts = 0,
    locked_until = NULL,
    updated_at = $2
WHERE user_id = $1
`
	_, err := r.db.ExecContext(ctx, q, userID, loginAt)
	if err != nil {
		return fmt.Errorf("update successful login: %w", err)
	}
	return nil
}

func (r *Repository) RecordFailedLogin(ctx context.Context, userID string, maxAttempts int, lockDurationSeconds int) (int, *time.Time, error) {
	const q = `
UPDATE metaldocs.auth_identities
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE WHEN failed_login_attempts + 1 >= $2
                        THEN NOW() + ($3 * INTERVAL '1 second')
                        ELSE locked_until END,
    updated_at = NOW()
WHERE user_id = $1
RETURNING failed_login_attempts, locked_until
`
	var attempts int
	var lockedUntil sql.NullTime
	if err := r.db.QueryRowContext(ctx, q, userID, maxAttempts, lockDurationSeconds).Scan(&attempts, &lockedUntil); err != nil {
		return 0, nil, fmt.Errorf("update failed login: %w", err)
	}
	if !lockedUntil.Valid {
		return attempts, nil, nil
	}
	locked := lockedUntil.Time.UTC()
	return attempts, &locked, nil
}

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

func (r *Repository) ListOnlineUsers(ctx context.Context, activeSince time.Time) ([]authdomain.OnlineUser, error) {
	const q = `
SELECT s.user_id, i.username, i.display_name, MAX(s.last_seen_at)
FROM metaldocs.auth_sessions s
JOIN metaldocs.auth_identities i ON i.user_id = s.user_id
WHERE s.revoked_at IS NULL
  AND s.expires_at > NOW()
  AND s.last_seen_at >= $1
  AND i.is_active = true
GROUP BY s.user_id, i.username, i.display_name
ORDER BY MAX(s.last_seen_at) DESC
`
	rows, err := r.db.QueryContext(ctx, q, activeSince.UTC())
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

func (r *Repository) UpdateUser(ctx context.Context, params authdomain.UpdateUserParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update user tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if params.DisplayName != nil || params.IsActive != nil {
		if _, err := tx.ExecContext(ctx, `
UPDATE metaldocs.auth_identities
SET display_name = COALESCE($2, display_name),
    is_active = COALESCE($3, is_active),
    updated_at = NOW()
WHERE user_id = $1
`, params.UserID, nullableText(params.DisplayName), nullableBool(params.IsActive)); err != nil {
			return fmt.Errorf("update auth identity profile: %w", err)
		}
	}

	if params.Email != nil || params.NewPasswordHash != nil || params.MustChangePassword != nil || params.ResetLockState {
		if _, err := tx.ExecContext(ctx, `
UPDATE metaldocs.auth_identities
SET email = COALESCE($2, email),
    password_hash = COALESCE($3, password_hash),
    password_algo = CASE WHEN $3 IS NULL THEN password_algo ELSE 'bcrypt' END,
    must_change_password = COALESCE($4, must_change_password),
    failed_login_attempts = CASE WHEN $3 IS NOT NULL OR $5 THEN 0 ELSE failed_login_attempts END,
    locked_until = CASE WHEN $3 IS NOT NULL OR $5 THEN NULL ELSE locked_until END,
    updated_at = NOW()
WHERE user_id = $1
`, params.UserID, nullableText(params.Email), nullablePasswordHash(params.NewPasswordHash), nullableBool(params.MustChangePassword), params.ResetLockState); err != nil {
			if isUniqueViolation(err) {
				return authdomain.ErrUserAlreadyExists
			}
			return fmt.Errorf("update auth identity: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update user tx: %w", err)
	}
	return nil
}

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
	var pgErr *pq.Error
	return errors.As(err, &pgErr) && string(pgErr.Code) == "23505"
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
