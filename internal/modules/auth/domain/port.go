package domain

import (
	"context"
	"time"
)

type Repository interface {
	FindIdentityByIdentifier(ctx context.Context, identifier string) (Identity, error)
	FindIdentityByUserID(ctx context.Context, userID string) (Identity, error)
	CreateSession(ctx context.Context, session Session) error
	FindSession(ctx context.Context, sessionID string) (Session, error)
	TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
	RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error
	RecordSuccessfulLogin(ctx context.Context, userID string, loginAt time.Time) error
	RecordFailedLogin(ctx context.Context, userID string, maxAttempts int, lockDurationSeconds int) (attempts int, lockedUntil *time.Time, err error)
	CreateUser(ctx context.Context, params CreateUserParams) error
	ListUsers(ctx context.Context) ([]ManagedUser, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) error
	ListOnlineUsers(ctx context.Context, activeSince time.Time) ([]OnlineUser, error)
	BootstrapAdmin(ctx context.Context, params BootstrapAdminParams) (bool, error)
	// GetUserTenants returns the distinct tenant IDs from iam_user_roles for the
	// given user. Used by Login to verify the X-Tenant-ID claim and bind it to
	// the session row.
	GetUserTenants(ctx context.Context, userID string) ([]string, error)
}
