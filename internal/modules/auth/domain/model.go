package domain

import (
	"encoding/json"
	"strings"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/iamtypes"
)

// PasswordHash must only be set via bcrypt.GenerateFromPassword; never assign raw strings.
type PasswordHash string

// PlainPassword holds an unhashed password in transit; never persisted or logged.
type PlainPassword string

// UserID identifies an auth identity.
type UserID string

// SessionID identifies a session.
type SessionID string

// TenantID identifies a tenant.
type TenantID string

// String returns the underlying user ID string.
func (id UserID) String() string { return string(id) }

// String returns the underlying tenant ID string.
func (id TenantID) String() string { return string(id) }

// Tenant is a minimal tenant projection used by the auth module.
type Tenant struct {
	ID   string
	Name string
	Slug string
}

// Identity is an authenticatable user record.
type Identity struct {
	UserID   string
	Username string
	Email    string
	// TODO: migration 0021 originally missed display_name/is_active in early
	// environments; keep later auth migrations applied before relying on them.
	DisplayName         string
	PasswordHash        PasswordHash
	PasswordAlgo        string
	MustChangePassword  bool
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	IsActive            bool
	Roles               []iamtypes.Role
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewIdentity constructs an active Identity with trimmed UserID and Username;
// all other fields are left zero for the caller to populate.
func NewIdentity(userID UserID, username string) Identity {
	return Identity{
		UserID:   strings.TrimSpace(userID.String()),
		Username: strings.TrimSpace(username),
		IsActive: true,
	}
}

// Session is a persisted login session.
type Session struct {
	SessionID  string
	UserID     string
	TenantID   string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	IPAddress  string
	UserAgent  string
	LastSeenAt time.Time
}

// OnlineUser is a projection of a user with a currently active (non-expired,
// non-revoked) session.
type OnlineUser struct {
	UserID      string
	Username    string
	DisplayName string
	LastSeenAt  time.Time
}

// ManagedUser is the admin-facing projection of an Identity for user-management
// listings (no PasswordHash).
type ManagedUser struct {
	UserID              string
	Username            string
	Email               string
	DisplayName         string
	IsActive            bool
	MustChangePassword  bool
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	Roles               []iamtypes.Role
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// CreateUserParams is the repository-layer input for persisting a new user;
// PasswordHash must already be hashed by the caller.
type CreateUserParams struct {
	UserID             string
	Username           string
	Email              string
	DisplayName        string
	PasswordHash       PasswordHash
	PasswordAlgo       string
	MustChangePassword bool
	IsActive           bool
	Roles              []iamtypes.Role
	CreatedBy          string
}

// CreateUserInput is the application-layer input for creating a user; Password
// is plaintext and must be hashed before persistence, never stored as-is.
type CreateUserInput struct {
	UserID      UserID
	Username    string
	Email       string
	DisplayName string
	Password    PlainPassword
	TenantID    TenantID
	Roles       []iamtypes.Role
	CreatedBy   string
}

// UpdateUserParams carries a partial update for an existing user; nil pointer
// fields are left unchanged. NewPasswordHash must already be hashed by the caller.
type UpdateUserParams struct {
	UserID             string
	DisplayName        *string
	Email              *string
	IsActive           *bool
	NewPasswordHash    *PasswordHash
	MustChangePassword *bool
	ResetLockState     bool
}

// BootstrapAdminParams seeds the first admin identity on system bootstrap;
// PasswordHash must already be hashed by the caller.
type BootstrapAdminParams struct {
	UserID             string
	Username           string
	Email              string
	DisplayName        string
	PasswordHash       PasswordHash
	PasswordAlgo       string
	MustChangePassword bool
}

// CurrentUser fields are PII; do not log this struct directly.
type CurrentUser struct {
	UserID             string          `json:"user_id"`
	TenantID           string          `json:"tenant_id"`
	TenantName         string          `json:"tenant_name"`
	Username           string          `json:"username"`
	Email              string          `json:"email,omitempty"`
	DisplayName        string          `json:"display_name"`
	MustChangePassword bool            `json:"must_change_password"`
	Roles              []iamtypes.Role `json:"roles"`
	// Capabilities is the union of capability codes the actor holds in the
	// active tenant. UX hint only — backend remains the sole trust boundary
	// (see wiki/concepts/authz-tiers.md). May be empty when no provider is
	// wired (in-memory dev modes / legacy tests).
	Capabilities []iamdomain.Capability `json:"capabilities"`
}

// AuthenticatedSession is the post-login result: raw token plus resolved
// CurrentUser. RawToken is redacted by both String and MarshalJSON — never
// log or serialize this struct through anything but the caller that issues
// the session cookie.
type AuthenticatedSession struct {
	RawToken    string
	CurrentUser CurrentUser
	ExpiresAt   time.Time
}

// String redacts RawToken so AuthenticatedSession is safe to pass to %v/%s
// formatting and loggers.
func (s AuthenticatedSession) String() string {
	return "***"
}

// MarshalJSON redacts RawToken so AuthenticatedSession is safe to serialize
// (e.g. in debug responses); the real token must be delivered via cookie only.
func (s AuthenticatedSession) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		RawToken    string      `json:"rawToken"`
		CurrentUser CurrentUser `json:"currentUser"`
		ExpiresAt   time.Time   `json:"expires_at"`
	}{
		RawToken:    "***",
		CurrentUser: s.CurrentUser,
		ExpiresAt:   s.ExpiresAt,
	})
}
