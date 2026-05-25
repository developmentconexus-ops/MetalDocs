package domain

import (
	"strings"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

type PasswordHash string
type PlainPassword string
type UserID string
type SessionID string
type TenantID string

func (id UserID) String() string    { return string(id) }
func (id SessionID) String() string { return string(id) }
func (id TenantID) String() string  { return string(id) }

type Identity struct {
	UserID              string
	Username            string
	Email               string
	DisplayName         string
	PasswordHash        PasswordHash
	PasswordAlgo        string
	MustChangePassword  bool
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	IsActive            bool
	Roles               []iamdomain.Role
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewIdentity(userID UserID, username string) Identity {
	return Identity{
		UserID:   strings.TrimSpace(userID.String()),
		Username: strings.TrimSpace(username),
		IsActive: true,
	}
}

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

type OnlineUser struct {
	UserID      string
	Username    string
	DisplayName string
	LastSeenAt  time.Time
}

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
	Roles               []iamdomain.Role
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreateUserParams struct {
	UserID             string
	Username           string
	Email              string
	DisplayName        string
	PasswordHash       PasswordHash
	PasswordAlgo       string
	MustChangePassword bool
	IsActive           bool
	Roles              []iamdomain.Role
	CreatedBy          string
}

type CreateUserInput struct {
	UserID      UserID
	Username    string
	Email       string
	DisplayName string
	Password    PlainPassword
	TenantID    TenantID
	Roles       []iamdomain.Role
	CreatedBy   string
}

type UpdateUserParams struct {
	UserID             string
	DisplayName        *string
	Email              *string
	IsActive           *bool
	NewPasswordHash    *PasswordHash
	MustChangePassword *bool
	ResetLockState     bool
}

type BootstrapAdminParams struct {
	UserID             string
	Username           string
	Email              string
	DisplayName        string
	PasswordHash       PasswordHash
	PasswordAlgo       string
	MustChangePassword bool
}

type CurrentUser struct {
	UserID             string           `json:"userId"`
	TenantID           string           `json:"tenantId"`
	Username           string           `json:"username"`
	Email              string           `json:"email,omitempty"`
	DisplayName        string           `json:"displayName"`
	MustChangePassword bool             `json:"mustChangePassword"`
	Roles              []iamdomain.Role `json:"roles"`
}

type AuthenticatedSession struct {
	RawToken    string
	CurrentUser CurrentUser
	ExpiresAt   time.Time
}
