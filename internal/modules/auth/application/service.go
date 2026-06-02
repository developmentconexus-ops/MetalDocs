package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/security"
	"metaldocs/internal/platform/tenant"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordAlgoBcrypt = "bcrypt"
	bcryptCost         = 12
)

type Secret string

func (s Secret) Value() string { return string(s) }

func (s Secret) String() string { return "***" }

type Config struct {
	SessionCookieName      string
	SessionTTL             time.Duration
	SessionSecret          Secret
	PasswordMinLength      int
	LoginMaxFailedAttempts int
	LoginLockDuration      time.Duration
	LegacyHeaderEnabled    bool
	OriginProtection       bool
	// AllowDevTenantFallback allows login to succeed for users with no IAM roles
	// by returning DevTenantID. Set true only in dev/test; defaults false (prod-safe).
	AllowDevTenantFallback bool
	TrustedOrigins         []string
	BootstrapAdminEnabled  bool
	BootstrapAdminUserID   string
	BootstrapAdminUsername string
	BootstrapAdminEmail    string
	BootstrapAdminPassword Secret
	BootstrapAdminName     string
	CookieSecure           bool
	// TrustedProxyCIDRs is the same allowlist consulted by the rate limiter
	// and origin-protection middleware. When the request's RemoteAddr falls
	// inside one of these prefixes, the leftmost X-Forwarded-For entry is
	// recorded as the session IP; otherwise RemoteAddr is used. Empty (default)
	// = upstream not trusted.
	TrustedProxyCIDRs []netip.Prefix
}

func (c Config) redacted() Config {
	if c.SessionSecret != "" {
		c.SessionSecret = Secret("***")
	}
	if c.BootstrapAdminPassword != "" {
		c.BootstrapAdminPassword = Secret("***")
	}
	return c
}

func (c Config) String() string {
	type configAlias Config
	redacted := configAlias(c.redacted())
	return fmt.Sprintf("%+v", redacted)
}

func (c Config) MarshalJSON() ([]byte, error) {
	type configAlias Config
	redacted := configAlias(c.redacted())
	return json.Marshal(redacted)
}

type Service struct {
	repo         authdomain.Repository
	roleProvider iamdomain.RoleProvider
	roleAdmin    iamdomain.RoleAdminRepository
	capProvider  authdomain.CapabilityProvider
	cfg          Config
	now          func() time.Time
}

// WithCapabilityProvider wires an optional CapabilityProvider so /auth/me and
// login responses can include a capabilities[] UX hint. Backend remains the
// sole authorization boundary.
func (s *Service) WithCapabilityProvider(p authdomain.CapabilityProvider) *Service {
	s.capProvider = p
	return s
}

type createUserTxRepository interface {
	CreateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.CreateUserParams) error
}

type replaceUserRolesTxRepository interface {
	ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error
}

type beginTxRepository interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func NewService(repo authdomain.Repository, roleProvider iamdomain.RoleProvider, roleAdmin iamdomain.RoleAdminRepository, cfg Config) (*Service, error) {
	if len(cfg.SessionSecret.Value()) < 32 {
		return nil, fmt.Errorf("new auth service: session secret must be at least 32 characters")
	}
	return &Service{
		repo:         repo,
		roleProvider: roleProvider,
		roleAdmin:    roleAdmin,
		cfg:          cfg,
		now:          time.Now,
	}, nil
}

func (s *Service) BootstrapLocalAdmin(ctx context.Context) error {
	if !s.cfg.BootstrapAdminEnabled || s.cfg.BootstrapAdminPassword == "" {
		return nil
	}
	if s.roleAdmin == nil {
		return fmt.Errorf("role admin repository is required")
	}

	// Bootstrap runs in single-tenant dev mode. Use the default tenant.
	hasAdmin, err := s.roleAdmin.HasAnyRole(ctx, iamdomain.RoleSystemAdmin, tenant.DevTenantID)
	if err != nil {
		return err
	}
	if hasAdmin {
		return nil
	}

	bootstrapPassword := []byte(s.cfg.BootstrapAdminPassword.Value())
	passwordHash, err := s.hashPasswordBytes(bootstrapPassword)
	zeroBytes(bootstrapPassword)
	s.cfg.BootstrapAdminPassword = ""
	if err != nil {
		return err
	}

	created, err := s.repo.BootstrapAdmin(ctx, authdomain.BootstrapAdminParams{
		UserID:             strings.TrimSpace(s.cfg.BootstrapAdminUserID),
		Username:           strings.TrimSpace(s.cfg.BootstrapAdminUsername),
		Email:              strings.TrimSpace(s.cfg.BootstrapAdminEmail),
		DisplayName:        strings.TrimSpace(s.cfg.BootstrapAdminName),
		PasswordHash:       passwordHash,
		PasswordAlgo:       passwordAlgoBcrypt,
		MustChangePassword: true,
	})
	if err != nil || !created {
		return err
	}
	return s.roleAdmin.UpsertUserAndAssignRole(
		ctx,
		strings.TrimSpace(s.cfg.BootstrapAdminUserID),
		strings.TrimSpace(s.cfg.BootstrapAdminName),
		tenant.DevTenantID,
		iamdomain.RoleSystemAdmin,
		"bootstrap",
	)
}

func (s *Service) Authenticate(ctx context.Context, identifier, password string, r *http.Request) (authdomain.AuthenticatedSession, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return authdomain.AuthenticatedSession{}, authdomain.ErrInvalidCredentials
	}

	identity, err := s.repo.FindIdentityByIdentifier(ctx, identifier)
	if err != nil {
		return authdomain.AuthenticatedSession{}, err
	}
	if identity.LockedUntil != nil && identity.LockedUntil.After(s.now().UTC()) {
		return authdomain.AuthenticatedSession{}, authdomain.ErrIdentityLocked
	}
	if !identity.IsActive {
		return authdomain.AuthenticatedSession{}, authdomain.ErrIdentityInactive
	}
	// TODO: use SELECT ... FOR UPDATE or advisory lock to make lockout atomic.
	if bcrypt.CompareHashAndPassword([]byte(identity.PasswordHash), []byte(password)) != nil {
		if _, _, err := s.repo.RecordFailedLogin(ctx, identity.UserID, s.cfg.LoginMaxFailedAttempts, int(s.cfg.LoginLockDuration.Seconds()), s.remoteIP(r)); err != nil {
			return authdomain.AuthenticatedSession{}, fmt.Errorf("record failed login: %w", err)
		}
		return authdomain.AuthenticatedSession{}, authdomain.ErrInvalidCredentials
	}

	now := s.now().UTC()
	claimedTenant := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	tenantID, err := s.resolveLoginTenant(ctx, identity.UserID, claimedTenant)
	if err != nil {
		return authdomain.AuthenticatedSession{}, err
	}
	if err := s.repo.RecordSuccessfulLogin(ctx, identity.UserID, now); err != nil {
		return authdomain.AuthenticatedSession{}, err
	}
	// Governance hint for the People-tab "Last login" drawer (PR-4). Best-effort:
	// failure to update iam_users must not block login — the credential row above
	// is the source of truth. PR-7 will populate deviceLabel via UA parsing.
	if err := s.repo.RecordLastLoginContext(ctx, identity.UserID, tenantID, s.remoteIP(r), truncate(strings.TrimSpace(r.UserAgent()), 512), ""); err != nil {
		// Swallow — already journalled by RecordSuccessfulLogin above.
		_ = err
	}

	rawToken, sessionID, err := s.newSessionToken()
	if err != nil {
		return authdomain.AuthenticatedSession{}, err
	}
	session := authdomain.Session{
		SessionID:  sessionID,
		UserID:     identity.UserID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.cfg.SessionTTL),
		IPAddress:  s.remoteIP(r),
		UserAgent:  truncate(strings.TrimSpace(r.UserAgent()), 512),
		LastSeenAt: now,
	}
	session.TenantID = tenantID

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return authdomain.AuthenticatedSession{}, err
	}

	user, err := s.buildCurrentUser(ctx, identity.UserID, tenantID)
	if err != nil {
		return authdomain.AuthenticatedSession{}, err
	}

	return authdomain.AuthenticatedSession{
		RawToken:    rawToken,
		CurrentUser: user,
		ExpiresAt:   session.ExpiresAt,
	}, nil
}

func (s *Service) resolveLoginTenant(ctx context.Context, userID, claimedTenantID string) (string, error) {
	tenants, err := s.repo.GetUserTenants(ctx, userID)
	if err != nil {
		return "", err
	}

	if claimedTenantID != "" {
		for _, t := range tenants {
			if t == claimedTenantID {
				return claimedTenantID, nil
			}
		}
		return "", authdomain.ErrTenantNotPermitted
	}

	if len(tenants) == 1 {
		return tenants[0], nil
	}
	if len(tenants) == 0 && s.cfg.AllowDevTenantFallback {
		return tenant.DevTenantID, nil
	}
	return "", authdomain.ErrTenantClaimRequired
}

func (s *Service) ResolveSession(ctx context.Context, rawToken string) (authdomain.CurrentUser, error) {
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return authdomain.CurrentUser{}, authdomain.ErrSessionNotFound
	}

	sessionID, err := s.tokenHashFromCookieValue(token)
	if err != nil {
		return authdomain.CurrentUser{}, authdomain.ErrSessionNotFound
	}

	session, err := s.repo.FindSession(ctx, sessionID)
	if err != nil {
		return authdomain.CurrentUser{}, err
	}
	if session.RevokedAt != nil {
		return authdomain.CurrentUser{}, authdomain.ErrSessionRevoked
	}
	if session.ExpiresAt.Before(s.now().UTC()) {
		return authdomain.CurrentUser{}, authdomain.ErrSessionExpired
	}
	now := s.now().UTC()
	if err := s.repo.TouchSession(ctx, sessionID, now); err != nil {
		return authdomain.CurrentUser{}, err
	}
	return s.buildCurrentUser(ctx, session.UserID, session.TenantID)
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return authdomain.ErrSessionNotFound
	}
	sessionID, err := s.tokenHashFromCookieValue(token)
	if err != nil {
		return err
	}
	return s.repo.RevokeSession(ctx, sessionID, s.now().UTC())
}

func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	return s.ChangePasswordForUser(ctx, authdomain.CurrentUser{UserID: userID}, currentPassword, newPassword)
}

func (s *Service) ChangePasswordForUser(ctx context.Context, currentUser authdomain.CurrentUser, currentPassword, newPassword string) error {
	userID := strings.TrimSpace(currentUser.UserID)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return authdomain.ErrInvalidCredentials
	}
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	identity, err := s.repo.FindIdentityByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if !currentUser.MustChangePassword {
		if currentPassword == "" {
			return authdomain.ErrInvalidCredentials
		}
		if bcrypt.CompareHashAndPassword([]byte(identity.PasswordHash), []byte(currentPassword)) != nil {
			return authdomain.ErrInvalidCredentials
		}
	}
	if currentUser.MustChangePassword && currentPassword != "" && bcrypt.CompareHashAndPassword([]byte(identity.PasswordHash), []byte(currentPassword)) != nil {
		return authdomain.ErrInvalidCredentials
	}

	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	required := false
	if err := s.repo.UpdateUser(ctx, authdomain.UpdateUserParams{
		UserID:             userID,
		NewPasswordHash:    &passwordHash,
		MustChangePassword: &required,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error) {
	items, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: batch role lookup with IN clause to avoid N+1 role queries.
	filtered := make([]authdomain.ManagedUser, 0, len(items))
	for i := range items {
		roles, roleErr := s.roleProvider.RolesByUserID(ctx, items[i].UserID, tenantID)
		if roleErr != nil {
			if errors.Is(roleErr, iamdomain.ErrUserNotFound) {
				continue
			}
			if errors.Is(roleErr, iamdomain.ErrNoRolesAssigned) {
				items[i].Roles = []iamdomain.Role{}
				filtered = append(filtered, items[i])
				continue
			}
			return nil, roleErr
		}
		items[i].Roles = roles
		filtered = append(filtered, items[i])
	}
	return filtered, nil
}

func (s *Service) ListOnlineUsers(ctx context.Context, tenantID string, activeSince time.Time) ([]authdomain.OnlineUser, error) {
	return s.repo.ListOnlineUsers(ctx, tenantID, activeSince)
}

func (s *Service) CreateUser(ctx context.Context, userID, username, email, displayName, password, tenantID string, roles []iamdomain.Role, createdBy string) error {
	return s.CreateUserWithInput(ctx, authdomain.CreateUserInput{
		UserID:      authdomain.UserID(userID),
		Username:    username,
		Email:       email,
		DisplayName: displayName,
		Password:    authdomain.PlainPassword(password),
		TenantID:    authdomain.TenantID(tenantID),
		Roles:       roles,
		CreatedBy:   createdBy,
	})
}

func (s *Service) CreateUserWithInput(ctx context.Context, input authdomain.CreateUserInput) error {
	fields, err := s.normalizeCreateUserInput(input)
	if err != nil {
		return err
	}
	params, err := s.buildCreateUserParams(fields)
	if err != nil {
		return err
	}
	if s.roleAdmin == nil {
		return fmt.Errorf("role admin repository is required")
	}
	if fields.tenantID == "" {
		fields.tenantID = tenant.DevTenantID
	}
	repoTx, repoTxOK := s.repo.(createUserTxRepository)
	roleTx, roleTxOK := s.roleAdmin.(replaceUserRolesTxRepository)
	beginner, beginOK := s.repo.(beginTxRepository)
	if repoTxOK && roleTxOK && beginOK {
		tx, err := beginner.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin create user tx: %w", err)
		}
		if err := repoTx.CreateUserTx(ctx, tx, params); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				return fmt.Errorf("create auth identity tx: %w (rollback failed: %v)", err, rollbackErr)
			}
			return err
		}
		if err := roleTx.ReplaceUserRolesTx(ctx, tx, fields.userID, fields.displayName, fields.tenantID, fields.role, fields.createdBy); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				return fmt.Errorf("replace user roles tx: %w (rollback failed: %v)", err, rollbackErr)
			}
			return err
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit create user tx: %w", err)
		}
		return nil
	}

	if err := s.repo.CreateUser(ctx, params); err != nil {
		return err
	}
	return s.roleAdmin.ReplaceUserRoles(ctx, fields.userID, fields.displayName, fields.tenantID, fields.role, fields.createdBy)
}

func (s *Service) UpdateUser(ctx context.Context, params authdomain.UpdateUserParams, newPassword string) error {
	newPassword = strings.TrimSpace(newPassword)
	if newPassword != "" {
		if err := s.validatePassword(newPassword); err != nil {
			return err
		}
		passwordHash, err := s.hashPassword(newPassword)
		if err != nil {
			return err
		}
		params.NewPasswordHash = &passwordHash
	}
	return s.repo.UpdateUser(ctx, params)
}

func (s *Service) AdminResetPassword(ctx context.Context, userID, newPassword string) error {
	newPassword = strings.TrimSpace(newPassword)
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}
	required := true
	if err := s.repo.UpdateUser(ctx, authdomain.UpdateUserParams{
		UserID:             strings.TrimSpace(userID),
		NewPasswordHash:    &passwordHash,
		MustChangePassword: &required,
		ResetLockState:     true,
	}); err != nil {
		return err
	}
	return s.repo.RevokeSessionsByUserID(ctx, strings.TrimSpace(userID), s.now().UTC())
}

func (s *Service) UnlockUser(ctx context.Context, userID string) error {
	return s.repo.UpdateUser(ctx, authdomain.UpdateUserParams{
		UserID:         strings.TrimSpace(userID),
		ResetLockState: true,
	})
}

func (s *Service) SessionCookie(rawToken string, expiresAt time.Time) *http.Cookie {
	seconds := int(expiresAt.UTC().Sub(s.now().UTC()).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		// SameSite=Strict keeps the session cookie off cross-site requests; revisit
		// this if future SSO or cross-site navigation flows require Lax semantics.
		SameSite: http.SameSiteStrictMode,
		Secure:   s.cfg.CookieSecure,
		Expires:  expiresAt.UTC(),
		MaxAge:   seconds,
	}
}

func (s *Service) SessionCookieName() string {
	return s.cfg.SessionCookieName
}

func (s *Service) ExpiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		// SameSite=Strict keeps the session cookie off cross-site requests; revisit
		// this if future SSO or cross-site navigation flows require Lax semantics.
		SameSite: http.SameSiteStrictMode,
		Secure:   s.cfg.CookieSecure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	}
}

func (s *Service) CurrentUser(ctx context.Context, userID, tenantID string) (authdomain.CurrentUser, error) {
	return s.buildCurrentUser(ctx, userID, tenantID)
}

func (s *Service) buildCurrentUser(ctx context.Context, userID, tenantID string) (authdomain.CurrentUser, error) {
	identity, err := s.repo.FindIdentityByUserID(ctx, userID)
	if err != nil {
		return authdomain.CurrentUser{}, err
	}
	tenantRow, err := s.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return authdomain.CurrentUser{}, err
	}
	roles, err := s.roleProvider.RolesByUserID(ctx, userID, tenantID)
	if err != nil {
		return authdomain.CurrentUser{}, err
	}
	// Always emit a slice (never nil) so the JSON response satisfies the
	// "required: [capabilities]" contract with an empty array when no provider
	// is wired or the user holds none.
	caps := []iamdomain.Capability{}
	if s.capProvider != nil {
		resolved, capErr := s.capProvider.CapsByUserID(ctx, userID, tenantID)
		if capErr != nil {
			return authdomain.CurrentUser{}, fmt.Errorf("resolve capabilities: %w", capErr)
		}
		if resolved != nil {
			caps = resolved
		}
	}
	return authdomain.CurrentUser{
		UserID:             identity.UserID,
		TenantID:           tenantID,
		TenantName:         tenantRow.Name,
		Username:           identity.Username,
		Email:              identity.Email,
		DisplayName:        identity.DisplayName,
		MustChangePassword: identity.MustChangePassword,
		Roles:              roles,
		Capabilities:       caps,
	}, nil
}

func (s *Service) validatePassword(password string) error {
	if password == "" || len(password) < s.cfg.PasswordMinLength {
		return fmt.Errorf("%w: password must contain at least %d characters", authdomain.ErrPasswordPolicy, s.cfg.PasswordMinLength)
	}
	return nil
}

func (s *Service) hashPassword(password string) (authdomain.PasswordHash, error) {
	return s.hashPasswordBytes([]byte(password))
}

func (s *Service) hashPasswordBytes(password []byte) (authdomain.PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword(password, bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return authdomain.PasswordHash(string(hash)), nil
}

type createUserFields struct {
	userID      string
	username    string
	email       string
	displayName string
	password    string
	tenantID    string
	role        iamdomain.Role
	createdBy   string
}

func (s *Service) normalizeCreateUserInput(input authdomain.CreateUserInput) (createUserFields, error) {
	fields := createUserFields{
		userID:      strings.TrimSpace(input.UserID.String()),
		username:    strings.TrimSpace(input.Username),
		email:       strings.TrimSpace(input.Email),
		displayName: strings.TrimSpace(input.DisplayName),
		password:    string(input.Password),
		tenantID:    strings.TrimSpace(input.TenantID.String()),
		createdBy:   strings.TrimSpace(input.CreatedBy),
	}
	if fields.userID == "" {
		fields.userID = fields.username
	}
	if fields.username == "" {
		return createUserFields{}, authdomain.ErrUserAlreadyExists
	}
	if fields.displayName == "" {
		fields.displayName = fields.username
	}
	if fields.createdBy == "" {
		fields.createdBy = "system"
	}
	if len(input.Roles) != 1 {
		return createUserFields{}, iamdomain.ErrInvalidRole
	}
	if !iamdomain.IsValidRole(input.Roles[0]) {
		return createUserFields{}, iamdomain.ErrInvalidRole
	}
	fields.role = input.Roles[0]
	if err := s.validatePassword(fields.password); err != nil {
		return createUserFields{}, err
	}
	return fields, nil
}

func (s *Service) buildCreateUserParams(fields createUserFields) (authdomain.CreateUserParams, error) {
	passwordHash, err := s.hashPassword(fields.password)
	if err != nil {
		return authdomain.CreateUserParams{}, err
	}
	return authdomain.CreateUserParams{
		UserID:             fields.userID,
		Username:           fields.username,
		Email:              fields.email,
		DisplayName:        fields.displayName,
		PasswordHash:       passwordHash,
		PasswordAlgo:       passwordAlgoBcrypt,
		MustChangePassword: true,
		IsActive:           true,
		Roles:              nil,
		CreatedBy:          fields.createdBy,
	}, nil
}

func (s *Service) newSessionToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	sig := s.signToken(token)
	cookieValue := token + "." + sig
	return cookieValue, hashToken(token), nil
}

func (s *Service) tokenHashFromCookieValue(raw string) (string, error) {
	parts := strings.Split(strings.TrimSpace(raw), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", authdomain.ErrSessionNotFound
	}
	if !hmac.Equal([]byte(parts[1]), []byte(s.signToken(parts[0]))) {
		return "", authdomain.ErrSessionNotFound
	}
	return hashToken(parts[0]), nil
}

func (s *Service) signToken(token string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.SessionSecret.Value()))
	_, _ = mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *Service) remoteIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if addr := security.ClientIP(r, s.cfg.TrustedProxyCIDRs); addr.IsValid() {
		return truncate(addr.String(), 128)
	}
	return truncate(strings.TrimSpace(r.RemoteAddr), 128)
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func zeroBytes(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
