// Package application implements the auth module's Service: login/logout,
// session resolution, password lifecycle, and user administration. It holds
// the timing-safe credential checks (constant-time dummy hash, per-identity
// login lock) and the session-revoke-on-credential-change invariant
// (CWE-613); all SQL and HTTP concerns are delegated to authdomain.Repository
// and the delivery/http package respectively.
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
	"log/slog"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/passwordhash"
	"metaldocs/internal/platform/requesttrace"
	"metaldocs/internal/platform/security"
	"metaldocs/internal/platform/tenant"

	"github.com/google/uuid"
)

// passwordAlgoBcrypt and passwordAlgoArgon2id are re-exported locally (not
// just referenced as passwordhash.AlgoBcrypt/AlgoArgon2id at each call site)
// so the algorithm names read naturally against authdomain.CreateUserParams
// / UpdateUserParams field names in this file; both are the same string
// value as their passwordhash counterparts — one source, no drift.
const (
	passwordAlgoBcrypt   = passwordhash.AlgoBcrypt
	passwordAlgoArgon2id = passwordhash.AlgoArgon2id
)

// Secret holds a config value that must never be logged (session HMAC key,
// bootstrap admin password); String and MarshalJSON both redact it.
type Secret string

// Value returns the underlying secret string. Callers must not log the result.
func (s Secret) Value() string { return string(s) }

// String redacts the secret so Config is safe to pass to %v/%s formatting and loggers.
func (s Secret) String() string { return "***" }

// Config configures the auth Service: session cookie/TTL/secret, password
// policy, login lockout, origin protection, and bootstrap admin credentials.
type Config struct {
	SessionCookieName string
	SessionTTL        time.Duration
	// SessionIdleTimeout, when > 0, expires a session that has not been seen
	// within the window (sliding idle timeout, measured against the session's
	// pre-touch LastSeenAt). 0 disables idle expiry — the default in dev/test so
	// existing behaviour is unchanged; production sets it via env.
	SessionIdleTimeout     time.Duration
	SessionSecret          Secret
	PasswordMinLength      int
	LoginMaxFailedAttempts int
	LoginLockDuration      time.Duration
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

// redacted returns a copy of c with SessionSecret and BootstrapAdminPassword
// replaced by the "***" placeholder.
func (c Config) redacted() Config {
	if c.SessionSecret != "" {
		c.SessionSecret = Secret("***")
	}
	if c.BootstrapAdminPassword != "" {
		c.BootstrapAdminPassword = Secret("***")
	}
	return c
}

// String redacts secret fields so Config is safe to pass to %v/%s formatting and loggers.
func (c Config) String() string {
	type configAlias Config
	redacted := configAlias(c.redacted())
	return fmt.Sprintf("%+v", redacted)
}

// MarshalJSON redacts secret fields so Config is safe to serialize (e.g. in debug dumps).
func (c Config) MarshalJSON() ([]byte, error) {
	type configAlias Config
	redacted := configAlias(c.redacted())
	return json.Marshal(redacted)
}

// Service implements the auth module's application logic: login/logout,
// session resolution, password lifecycle, and user administration.
type Service struct {
	repo         authdomain.Repository
	roleProvider iamdomain.RoleProvider
	roleAdmin    iamdomain.RoleAdminRepository
	capProvider  authdomain.CapabilityProvider
	// loginCtxPort records governance metadata on iam_users after a successful
	// login. Required: must be non-nil at construction (F-06c, 2.13 step 1).
	loginCtxPort iamdomain.LoginContextPort
	// audit writes audit events. When non-nil and the repo supports tx variants,
	// mutations record their audit row inside the same transaction (H-3b, F-07).
	audit auditdomain.Writer
	cfg   Config
	now   func() time.Time
	// dummyHash is a fixed Argon2id PHC hash compared against on the
	// unknown-identifier path so that login spends Argon2id-equivalent time
	// whether or not the account exists. This removes the timing oracle that
	// let wall-clock latency reveal account existence (OWASP Authentication
	// Cheat Sheet — equalize work). It is the package-level
	// constantTimeDummyHash() value — computed once per process (see there),
	// never once per Service.
	dummyHash string
}

// constantTimeDummyHash and its guarding sync.Once make dummy-hash
// generation a once-per-process cost, not a once-per-Service cost. Before
// this fix, every NewService call ran a full KDF (bcrypt.GenerateFromPassword)
// to mint a value that is thrown away — no caller ever inspects it, it only
// exists to be compared against, spending KDF-equivalent wall-clock time on
// the unknown-identifier login path (see Authenticate). TestSurfaceConformance
// builds a fresh Service per route (147 routes) under -race, so a
// once-per-construction KDF run pushed apps/api/cmd/metaldocs-api's
// integration package past its timeout — and Argon2id is more expensive than
// bcrypt by design, so leaving the per-construction cost in place would have
// made that strictly worse. sync.Once bounds the real cost to one KDF run for
// the lifetime of the process, however many Services get built.
var (
	constantTimeDummyHashOnce sync.Once
	constantTimeDummyHashVal  string
	constantTimeDummyHashErr  error
)

func constantTimeDummyHash() (string, error) {
	constantTimeDummyHashOnce.Do(func() {
		constantTimeDummyHashVal, constantTimeDummyHashErr = passwordhash.HashArgon2id([]byte("metaldocs-constant-time-dummy"))
	})
	return constantTimeDummyHashVal, constantTimeDummyHashErr
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
	ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, role iamtypes.Role, assignedBy string) error
}

type beginTxRepository interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type updateUserTxRepository interface {
	UpdateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.UpdateUserParams) error
}

type revokeSessionsByUserIDTxRepository interface {
	RevokeSessionsByUserIDTx(ctx context.Context, tx *sql.Tx, userID string, revokedAt time.Time) error
}

// NewService constructs a Service. loginCtxPort is required (panics if nil).
// cfg.SessionSecret must be at least 32 characters, or construction fails.
// auditWriter is optional and variadic; when provided, mutations record their
// audit row in the same transaction where the repository supports it (H-3b, F-07).
func NewService(repo authdomain.Repository, roleProvider iamdomain.RoleProvider, roleAdmin iamdomain.RoleAdminRepository, loginCtxPort iamdomain.LoginContextPort, cfg Config, auditWriter ...auditdomain.Writer) (*Service, error) {
	if loginCtxPort == nil {
		panic("auth.NewService: loginCtxPort is required")
	}
	if len(cfg.SessionSecret.Value()) < 32 {
		return nil, fmt.Errorf("new auth service: session secret must be at least 32 characters")
	}
	// Computed once per process (constantTimeDummyHash, guarded by sync.Once),
	// not once per Service: the unknown-identifier path compares against this
	// so login latency is Argon2id-dominated regardless of account existence
	// (A1, timing oracle) without paying a fresh KDF run on every construction.
	dummyHash, err := constantTimeDummyHash()
	if err != nil {
		return nil, fmt.Errorf("new auth service: generate constant-time hash: %w", err)
	}
	var aw auditdomain.Writer
	if len(auditWriter) > 0 {
		aw = auditWriter[0]
	}
	return &Service{
		repo:         repo,
		roleProvider: roleProvider,
		roleAdmin:    roleAdmin,
		loginCtxPort: loginCtxPort,
		audit:        aw,
		cfg:          cfg,
		now:          time.Now,
		dummyHash:    dummyHash,
	}, nil
}

// BootstrapLocalAdmin seeds the first system-admin identity for single-tenant
// dev mode when cfg.BootstrapAdminEnabled is set and no admin role exists yet
// for the dev tenant; it is a no-op (nil error) once an admin is present.
func (s *Service) BootstrapLocalAdmin(ctx context.Context) error {
	if !s.cfg.BootstrapAdminEnabled || s.cfg.BootstrapAdminPassword == "" {
		return nil
	}
	if s.roleAdmin == nil {
		return fmt.Errorf("role admin repository is required")
	}

	// Bootstrap runs in single-tenant dev mode. Use the default tenant.
	hasAdmin, err := s.roleAdmin.HasAnyRole(ctx, iamtypes.RoleSystemAdmin, tenant.DevTenantID)
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
		PasswordAlgo:       passwordAlgoArgon2id,
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
		iamtypes.RoleSystemAdmin,
		"bootstrap",
	)
}

// loginOutcome is the result of the credential check performed inside the
// per-identity login lock, decided in the locked region and acted on after it.
type loginOutcome int

const (
	loginOK loginOutcome = iota
	loginInvalid
	loginLocked
	loginInactive
)

// Authenticate verifies identifier+password and, on success, creates a new
// session and returns an AuthenticatedSession. It runs the credential check
// inside a per-identity login lock so the lockout check and failed-attempt
// write are atomic (closes a TOCTOU window across concurrent attempts), and
// spends KDF-equivalent time on every failure path — unknown identifier,
// locked, inactive, or wrong password — so wall-clock latency never discloses
// account existence or state (OWASP Authentication Cheat Sheet). Returns
// ErrInvalidCredentials, ErrIdentityLocked, ErrIdentityInactive,
// ErrTenantNotPermitted, or ErrTenantClaimRequired depending on outcome.
//
// REQ-AUTHN-1 timing note: the unknown-identifier path (below) spends
// Argon2id-equivalent time via the package-level dummy hash — that is the
// algorithm every identity converges to, so once migration completes this
// path's cost matches every real comparison. During migration, a
// not-yet-rehashed bcrypt identity's wrong-password attempt still costs
// bcrypt-equivalent time (this function's known-identity branch dispatches
// on the identity's own stored algorithm), which is more expensive than
// Argon2id at this package's chosen parameters. That gap is a real, named
// residual: for the duration of the migration window, a wrong-password
// attempt against a legacy-bcrypt identity is wall-clock distinguishable
// from one against an unknown identifier or an already-migrated argon2id
// identity. It closes automatically as accounts rehash on login and
// disappears entirely once no bcrypt identity remains.
func (s *Service) Authenticate(ctx context.Context, identifier, password string, r *http.Request) (authdomain.AuthenticatedSession, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return authdomain.AuthenticatedSession{}, authdomain.ErrInvalidCredentials
	}

	identity, err := s.repo.FindIdentityByIdentifier(ctx, identifier)
	if err != nil {
		// Unknown identifier: spend Argon2id-equivalent time (this package's
		// target algorithm) before returning the generic credential error so
		// wall-clock latency does not reveal whether the account exists (A1).
		// Any other error is a real fault — return it.
		if errors.Is(err, authdomain.ErrIdentityNotFound) {
			_, _ = passwordhash.VerifyArgon2id(s.dummyHash, []byte(password))
			return authdomain.AuthenticatedSession{}, authdomain.ErrInvalidCredentials
		}
		return authdomain.AuthenticatedSession{}, err
	}

	// Verify credentials while holding a per-identity lock so the lockout check
	// and the failed-attempt write are atomic. Concurrent attempts on the same
	// identity are serialized: once the threshold lock is set, the remaining
	// in-flight attempts observe it and are rejected before the KDF runs —
	// closing the read-check-then-write TOCTOU window that previously let a
	// parallel burst exceed LoginMaxFailedAttempts guesses within a single
	// lock window.
	outcome := loginOK
	if lockErr := s.repo.WithinLoginLock(ctx, identity.UserID, func(tx authdomain.LoginTx) error {
		state, err := tx.LoadLoginState(ctx, identity.UserID)
		if err != nil {
			return err
		}
		// Run the credential comparison before branching so a locked or inactive
		// account spends the same time as an active one — the fast no-KDF path
		// would otherwise leak account state by wall-clock (OWASP Authentication
		// Cheat Sheet — equalize work on every failure path). The distinct
		// locked/inactive errors are kept deliberately (admin UX), only the
		// timing is equalized. Verify dispatches on state.PasswordAlgo
		// (REQ-AUTHN-1) and fails closed — verifyErr != nil (unrecognized
		// algorithm or a malformed stored hash) is treated as a failed
		// comparison, never as a match and never by falling through to a
		// different algorithm (no-fallback principle).
		passwordOK, verifyErr := passwordhash.Verify(state.PasswordAlgo, []byte(state.PasswordHash), []byte(password))
		if verifyErr != nil {
			passwordOK = false
		}
		if state.LockedUntil != nil && state.LockedUntil.After(s.now().UTC()) {
			outcome = loginLocked
			return nil
		}
		if !state.IsActive {
			outcome = loginInactive
			return nil
		}
		if !passwordOK {
			if _, _, err := tx.RecordFailedLogin(ctx, identity.UserID, s.cfg.LoginMaxFailedAttempts, int(s.cfg.LoginLockDuration.Seconds()), s.remoteIP(r)); err != nil {
				return fmt.Errorf("record failed login: %w", err)
			}
			outcome = loginInvalid
			return nil
		}
		// Rehash-on-login migration (REQ-AUTHN-1): a successful login against a
		// legacy bcrypt hash rehashes the verified plaintext to Argon2id and
		// persists it in the same locked transaction. A write failure here must
		// NOT fail the login — the user already authenticated correctly with the
		// credential on file; a migration hiccup must not lock them out. Log and
		// move on; the identity simply stays on bcrypt and gets another rehash
		// attempt on its next successful login.
		if state.PasswordAlgo == passwordAlgoBcrypt {
			newHash, hashErr := passwordhash.HashArgon2id([]byte(password))
			if hashErr != nil {
				slog.ErrorContext(ctx, "auth: rehash-on-login: hash password failed; login proceeds on legacy bcrypt hash",
					"user_id", identity.UserID, "err", hashErr)
			} else if rehashErr := tx.RehashPassword(ctx, identity.UserID, authdomain.PasswordHash(newHash), passwordAlgoArgon2id); rehashErr != nil {
				slog.ErrorContext(ctx, "auth: rehash-on-login: persist argon2id hash failed; login proceeds on legacy bcrypt hash",
					"user_id", identity.UserID, "err", rehashErr)
			}
		}
		return nil
	}); lockErr != nil {
		return authdomain.AuthenticatedSession{}, lockErr
	}
	switch outcome {
	case loginLocked:
		return authdomain.AuthenticatedSession{}, authdomain.ErrIdentityLocked
	case loginInactive:
		return authdomain.AuthenticatedSession{}, authdomain.ErrIdentityInactive
	case loginInvalid:
		return authdomain.AuthenticatedSession{}, authdomain.ErrInvalidCredentials
	case loginOK:
		// Credentials verified; fall through to session issuance below.
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
	// The write is delegated to the IAM module via LoginContextPort (F-06c) so
	// that auth does not hold a direct SQL dependency on iam_users.
	if err := s.loginCtxPort.RecordLoginContext(ctx, identity.UserID, tenantID, s.remoteIP(r), truncate(strings.TrimSpace(r.UserAgent()), 512), ""); err != nil {
		// Swallow — governance hint must not block login.
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

// ResolveSession validates rawToken's HMAC signature, loads the matching
// session, and returns its CurrentUser. Returns ErrSessionNotFound,
// ErrSessionRevoked, or ErrSessionExpired (including sliding idle-timeout
// expiry when cfg.SessionIdleTimeout > 0) on failure. On success it touches
// the session's LastSeenAt.
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
	now := s.now().UTC()
	if session.RevokedAt != nil {
		return authdomain.CurrentUser{}, authdomain.ErrSessionRevoked
	}
	if session.ExpiresAt.Before(now) {
		return authdomain.CurrentUser{}, authdomain.ErrSessionExpired
	}
	// Sliding idle timeout (A2): expire the session if it has not been seen within
	// the window. Measured against the pre-touch LastSeenAt — TouchSession below
	// would otherwise advance it and defeat the check. 0 = disabled.
	if s.cfg.SessionIdleTimeout > 0 && now.Sub(session.LastSeenAt) > s.cfg.SessionIdleTimeout {
		return authdomain.CurrentUser{}, authdomain.ErrSessionExpired
	}
	if err := s.repo.TouchSession(ctx, sessionID, now); err != nil {
		return authdomain.CurrentUser{}, err
	}
	return s.buildCurrentUser(ctx, session.UserID, session.TenantID)
}

// Logout revokes the session identified by rawToken. Returns
// ErrSessionNotFound if the token is empty or malformed.
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

// currentPasswordMatches dispatches on identity.PasswordAlgo (REQ-AUTHN-1)
// to compare candidate against the stored hash, failing closed (false) on
// any verification error — an unrecognized algorithm or a malformed stored
// hash is never treated as a match.
func (s *Service) currentPasswordMatches(identity authdomain.Identity, candidate string) bool {
	ok, err := passwordhash.Verify(identity.PasswordAlgo, []byte(identity.PasswordHash), []byte(candidate))
	if err != nil {
		return false
	}
	return ok
}

// ChangePasswordForUser changes currentUser's password to newPassword,
// verifying currentPassword first unless MustChangePassword is set. On
// success it revokes all of the user's other sessions (CWE-613: a stolen or
// stale session must not survive a password change) and, where the
// repository supports transactional variants, does so atomically with an
// audit row; otherwise the steps run as sequential best-effort autocommits.
func (s *Service) ChangePasswordForUser(ctx context.Context, currentUser authdomain.CurrentUser, currentPassword, newPassword string) error {
	userID := strings.TrimSpace(currentUser.UserID)
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
		if !s.currentPasswordMatches(identity, currentPassword) {
			return authdomain.ErrInvalidCredentials
		}
	}
	if currentUser.MustChangePassword && currentPassword != "" && !s.currentPasswordMatches(identity, currentPassword) {
		return authdomain.ErrInvalidCredentials
	}

	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	required := false
	newAlgo := passwordAlgoArgon2id
	updateParams := authdomain.UpdateUserParams{
		UserID:             userID,
		NewPasswordHash:    &passwordHash,
		NewPasswordAlgo:    &newAlgo,
		MustChangePassword: &required,
	}
	now := s.now().UTC()

	repoTx, repoTxOK := s.repo.(updateUserTxRepository)
	revokeTx, revokeTxOK := s.repo.(revokeSessionsByUserIDTxRepository)
	beginner, beginOK := s.repo.(beginTxRepository)
	if !repoTxOK || !revokeTxOK || !beginOK {
		// Fallback: repo does not support tx variants (e.g. in-memory test repo).
		// No transaction is available here, so the steps run sequentially as
		// separate autocommits; the audit Record is best-effort. (This branch
		// precedes the atomic path so the analyzer does not correlate this
		// non-tx Record with the atomic path's Commit below.)
		if err := s.repo.UpdateUser(ctx, updateParams); err != nil {
			return err
		}
		if err := s.repo.RevokeSessionsByUserID(ctx, userID, now); err != nil {
			return err
		}
		if s.audit != nil {
			raw, _ := json.Marshal(map[string]any{})
			_ = s.audit.Record(ctx, auditdomain.Event{
				ID:           uuid.NewString(),
				OccurredAt:   now,
				ActorID:      userID,
				Action:       "auth.password.changed",
				ResourceType: "user",
				ResourceID:   userID,
				PayloadJSON:  string(raw),
				TraceID:      requesttrace.Resolve(ctx),
				TenantID:     currentUser.TenantID,
			})
		}
		return nil
	}

	// Atomic path: mutation + session revocation + audit row in one tx.
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin change password tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := repoTx.UpdateUserTx(ctx, tx, updateParams); err != nil {
		return err
	}
	// Revoke all of the user's sessions so a stolen or stale session cannot
	// survive a password change (CWE-613) — mirrors AdminResetPassword.
	if err := revokeTx.RevokeSessionsByUserIDTx(ctx, tx, userID, now); err != nil {
		return err
	}
	if s.audit != nil {
		raw, _ := json.Marshal(map[string]any{})
		ev := auditdomain.Event{
			ID:           uuid.NewString(),
			OccurredAt:   now,
			ActorID:      userID,
			Action:       "auth.password.changed",
			ResourceType: "user",
			ResourceID:   userID,
			PayloadJSON:  string(raw),
			TraceID:      requesttrace.Resolve(ctx),
			TenantID:     currentUser.TenantID,
		}
		if err := s.audit.RecordTx(ctx, tx, ev); err != nil {
			return fmt.Errorf("audit change password: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit change password tx: %w", err)
	}
	return nil
}

// ListUsers returns all users that hold a role in tenantID, with roles
// populated from a single batched lookup (REQ-DATA-2 / F-10). Users absent
// from the tenant's role map are excluded; users present with no roles are
// included with an empty Roles slice.
func (s *Service) ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error) {
	items, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	// Batch-fetch all tenant roles in one round trip (REQ-DATA-2 / F-10).
	userIDs := make([]string, len(items))
	for i := range items {
		userIDs[i] = items[i].UserID
	}
	rolesMap, err := s.roleProvider.RolesByUserIDs(ctx, tenantID, userIDs)
	if err != nil {
		return nil, err
	}

	// Build result: users absent from the map are not in this tenant (skip them).
	// Users present with an empty slice are active but role-less (include with empty roles).
	filtered := make([]authdomain.ManagedUser, 0, len(items))
	for i := range items {
		roles, active := rolesMap[items[i].UserID]
		if !active {
			continue
		}
		items[i].Roles = roles
		filtered = append(filtered, items[i])
	}
	return filtered, nil
}

// ListOnlineUsers returns users in tenantID with a session seen at or after activeSince.
func (s *Service) ListOnlineUsers(ctx context.Context, tenantID string, activeSince time.Time) ([]authdomain.OnlineUser, error) {
	return s.repo.ListOnlineUsers(ctx, tenantID, activeSince)
}

// CreateUser is a convenience wrapper over CreateUserWithInput taking loose
// string/role arguments instead of a CreateUserInput.
func (s *Service) CreateUser(ctx context.Context, userID, username, email, displayName, password, tenantID string, roles []iamtypes.Role, createdBy string) error {
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

// CreateUserWithInput creates a new identity and assigns exactly one role
// (input.Roles must contain exactly one valid role). Where the repository and
// role-admin collaborator both support transactional variants, identity
// creation and role assignment commit atomically; otherwise they run as two
// sequential autocommits.
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
				return fmt.Errorf("create auth identity tx: %w (rollback failed: %w)", err, rollbackErr)
			}
			return err
		}
		if err := roleTx.ReplaceUserRolesTx(ctx, tx, fields.userID, fields.displayName, fields.tenantID, fields.role, fields.createdBy); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				return fmt.Errorf("replace user roles tx: %w (rollback failed: %w)", err, rollbackErr)
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

// UpdateUser applies params to a user, optionally setting newPassword (hashed
// here). When params.IsActive transitions to false, it also revokes all of
// the user's sessions in the same operation (F8.4, CWE-613: a session must
// not outlive the account it authenticates) — atomically where the
// repository supports transactional variants, otherwise as two autocommits.
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
		newAlgo := passwordAlgoArgon2id
		params.NewPasswordHash = &passwordHash
		params.NewPasswordAlgo = &newAlgo
	}

	// Deactivation revokes all of the user's sessions (F8.4, CWE-613): a session must
	// not outlive the account it authenticates. Mirrors the credential-change revoke
	// (ChangePasswordForUser / AdminResetPassword). Revoke is idempotent, so it fires
	// whenever IsActive is being set false — over-firing is harmless, under-firing is
	// a security hole. A non-deactivating update keeps the plain single-write path.
	deactivating := params.IsActive != nil && !*params.IsActive
	if !deactivating {
		return s.repo.UpdateUser(ctx, params)
	}

	now := s.now().UTC()
	repoTx, repoTxOK := s.repo.(updateUserTxRepository)
	revokeTx, revokeTxOK := s.repo.(revokeSessionsByUserIDTxRepository)
	beginner, beginOK := s.repo.(beginTxRepository)
	if !repoTxOK || !revokeTxOK || !beginOK {
		// Fallback: in-memory / test repo — no tx available, so the mutation and the
		// revoke run as separate autocommits.
		if err := s.repo.UpdateUser(ctx, params); err != nil {
			return err
		}
		return s.repo.RevokeSessionsByUserID(ctx, params.UserID, now)
	}

	// Atomic path: deactivation + session revocation in one tx. No authz-recording
	// read inside the tx (H-PRE-1) — same shape as AdminResetPassword.
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin deactivate user tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := repoTx.UpdateUserTx(ctx, tx, params); err != nil {
		return err
	}
	if err := revokeTx.RevokeSessionsByUserIDTx(ctx, tx, params.UserID, now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit deactivate user tx: %w", err)
	}
	return nil
}

// AdminResetPassword sets userID's password to newPassword, forces
// MustChangePassword, clears any lockout, and revokes all of the user's
// sessions (CWE-613) — atomically with an audit row where the repository
// supports transactional variants, otherwise as sequential best-effort
// autocommits. Actor/tenant for the audit row are resolved from ctx.
func (s *Service) AdminResetPassword(ctx context.Context, userID, newPassword string) error {
	userID = strings.TrimSpace(userID)
	newPassword = strings.TrimSpace(newPassword)
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}
	required := true
	newAlgo := passwordAlgoArgon2id
	updateParams := authdomain.UpdateUserParams{
		UserID:             userID,
		NewPasswordHash:    &passwordHash,
		NewPasswordAlgo:    &newAlgo,
		MustChangePassword: &required,
		ResetLockState:     true,
	}
	now := s.now().UTC()

	// Resolve actor and tenant from the request context (set by auth middleware).
	// Falls back to empty strings when no session is present (e.g. test context).
	actorID := ""
	tenantID := ""
	if principal, ok := authdomain.CurrentUserFromContext(ctx); ok {
		actorID = principal.UserID
		tenantID = principal.TenantID
	}

	repoTx, repoTxOK := s.repo.(updateUserTxRepository)
	revokeTx, revokeTxOK := s.repo.(revokeSessionsByUserIDTxRepository)
	beginner, beginOK := s.repo.(beginTxRepository)
	if !repoTxOK || !revokeTxOK || !beginOK {
		// Fallback: in-memory / test repo.
		// No transaction is available here, so the steps run sequentially as
		// separate autocommits; the audit Record is best-effort. (This branch
		// precedes the atomic path so the analyzer does not correlate this
		// non-tx Record with the atomic path's Commit below.)
		if err := s.repo.UpdateUser(ctx, updateParams); err != nil {
			return err
		}
		if err := s.repo.RevokeSessionsByUserID(ctx, userID, now); err != nil {
			return err
		}
		if s.audit != nil {
			raw, _ := json.Marshal(map[string]any{"must_change_password": true})
			_ = s.audit.Record(ctx, auditdomain.Event{
				ID:           uuid.NewString(),
				OccurredAt:   now,
				ActorID:      actorID,
				Action:       "auth.user.password_reset",
				ResourceType: "user",
				ResourceID:   userID,
				PayloadJSON:  string(raw),
				TraceID:      requesttrace.Resolve(ctx),
				TenantID:     tenantID,
			})
		}
		return nil
	}

	// Atomic path: mutation + session revocation + audit row in one tx.
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin admin reset password tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := repoTx.UpdateUserTx(ctx, tx, updateParams); err != nil {
		return err
	}
	if err := revokeTx.RevokeSessionsByUserIDTx(ctx, tx, userID, now); err != nil {
		return err
	}
	if s.audit != nil {
		raw, _ := json.Marshal(map[string]any{"must_change_password": true})
		ev := auditdomain.Event{
			ID:           uuid.NewString(),
			OccurredAt:   now,
			ActorID:      actorID,
			Action:       "auth.user.password_reset",
			ResourceType: "user",
			ResourceID:   userID,
			PayloadJSON:  string(raw),
			TraceID:      requesttrace.Resolve(ctx),
			TenantID:     tenantID,
		}
		if err := s.audit.RecordTx(ctx, tx, ev); err != nil {
			return fmt.Errorf("audit admin reset password: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit admin reset password tx: %w", err)
	}
	return nil
}

// UnlockUser clears userID's lockout state, recording an audit row where the
// repository supports transactional variants (atomic), otherwise best-effort.
func (s *Service) UnlockUser(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	updateParams := authdomain.UpdateUserParams{
		UserID:         userID,
		ResetLockState: true,
	}
	now := s.now().UTC()

	actorID := ""
	tenantID := ""
	if principal, ok := authdomain.CurrentUserFromContext(ctx); ok {
		actorID = principal.UserID
		tenantID = principal.TenantID
	}

	repoTx, repoTxOK := s.repo.(updateUserTxRepository)
	beginner, beginOK := s.repo.(beginTxRepository)
	if !repoTxOK || !beginOK {
		// Fallback: in-memory / test repo.
		// No transaction is available here, so the steps run sequentially as
		// separate autocommits; the audit Record is best-effort. (This branch
		// precedes the atomic path so the analyzer does not correlate this
		// non-tx Record with the atomic path's Commit below.)
		if err := s.repo.UpdateUser(ctx, updateParams); err != nil {
			return err
		}
		if s.audit != nil {
			raw, _ := json.Marshal(map[string]any{})
			_ = s.audit.Record(ctx, auditdomain.Event{
				ID:           uuid.NewString(),
				OccurredAt:   now,
				ActorID:      actorID,
				Action:       "auth.user.unlocked",
				ResourceType: "user",
				ResourceID:   userID,
				PayloadJSON:  string(raw),
				TraceID:      requesttrace.Resolve(ctx),
				TenantID:     tenantID,
			})
		}
		return nil
	}

	// Atomic path: mutation + audit row in one tx.
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin unlock user tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := repoTx.UpdateUserTx(ctx, tx, updateParams); err != nil {
		return err
	}
	if s.audit != nil {
		raw, _ := json.Marshal(map[string]any{})
		ev := auditdomain.Event{
			ID:           uuid.NewString(),
			OccurredAt:   now,
			ActorID:      actorID,
			Action:       "auth.user.unlocked",
			ResourceType: "user",
			ResourceID:   userID,
			PayloadJSON:  string(raw),
			TraceID:      requesttrace.Resolve(ctx),
			TenantID:     tenantID,
		}
		if err := s.audit.RecordTx(ctx, tx, ev); err != nil {
			return fmt.Errorf("audit unlock user: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit unlock user tx: %w", err)
	}
	return nil
}

// SessionCookie builds the HttpOnly, SameSite=Strict session cookie carrying
// rawToken, with MaxAge clamped to 0 if expiresAt is already in the past.
func (s *Service) SessionCookie(rawToken string, expiresAt time.Time) *http.Cookie {
	seconds := int(expiresAt.UTC().Sub(s.now().UTC()).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return &http.Cookie{ // #nosec G124 -- HttpOnly and SameSite=Strict are hard literals below; Secure is config-driven (internal/platform/authn/config.go: METALDOCS_AUTH_COOKIE_SECURE defaults true everywhere except appEnv=="local"), not omitted — gosec only recognizes a literal `Secure: true`, not a variable that resolves to true in every non-local environment.
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

// SessionCookieName returns the configured session cookie name.
func (s *Service) SessionCookieName() string {
	return s.cfg.SessionCookieName
}

// ExpiredSessionCookie returns a cookie that instructs the browser to delete
// the session cookie (empty value, MaxAge -1, Expires in the past).
func (s *Service) ExpiredSessionCookie() *http.Cookie {
	return &http.Cookie{ // #nosec G124 -- HttpOnly and SameSite=Strict are hard literals below; Secure is config-driven (internal/platform/authn/config.go: METALDOCS_AUTH_COOKIE_SECURE defaults true everywhere except appEnv=="local"), not omitted — gosec only recognizes a literal `Secure: true`, not a variable that resolves to true in every non-local environment.
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

// CurrentUser resolves the CurrentUser projection for userID in tenantID.
// Returns ErrIdentityInactive if the identity has been deactivated (fail-closed
// backstop for sessions that predate a deactivation, mirroring the
// revoke-on-deactivate write path in UpdateUser).
func (s *Service) CurrentUser(ctx context.Context, userID, tenantID string) (authdomain.CurrentUser, error) {
	return s.buildCurrentUser(ctx, userID, tenantID)
}

func (s *Service) buildCurrentUser(ctx context.Context, userID, tenantID string) (authdomain.CurrentUser, error) {
	identity, err := s.repo.FindIdentityByUserID(ctx, userID)
	if err != nil {
		return authdomain.CurrentUser{}, err
	}
	// Fail closed (F8.4, CWE-613): a deactivated identity never yields a CurrentUser,
	// even if a live session row still exists. This is the read-path backstop to the
	// revoke-on-deactivate write path in UpdateUser — resolve rejects the session
	// regardless of which path (or neither) revoked it. Login (Authenticate) already
	// rejects inactive accounts earlier, so this only ever fires on the resolve path.
	if !identity.IsActive {
		return authdomain.CurrentUser{}, authdomain.ErrIdentityInactive
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
	hash, err := passwordhash.HashArgon2id(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return authdomain.PasswordHash(hash), nil
}

// HashPassword is auth's published password-hashing seam for cross-module
// callers (M7 F7.2: iam's OnboardTenantService hashes the first admin's
// password inside its own provisioning tx). It uses the same Argon2id
// mechanism and the same passwordhash package as Service.hashPassword
// (REQ-AUTHN-1), so the KDF params have exactly one home and cannot drift
// across modules. It returns the algo discriminator ALONGSIDE the hash (the
// passwordhash.AlgoArgon2id constant, never a literal) so no caller can pair
// the hash with a stale or hand-typed algo stamp — the mis-pair that broke
// first login for every onboarded tenant admin when this seam migrated to
// Argon2id but iam kept hard-coding "bcrypt" is now unrepresentable.
func HashPassword(plain string) (authdomain.PasswordHash, string, error) {
	hash, err := passwordhash.HashArgon2id([]byte(plain))
	if err != nil {
		return "", "", fmt.Errorf("hash password: %w", err)
	}
	return authdomain.PasswordHash(hash), passwordAlgoArgon2id, nil
}

type createUserFields struct {
	userID      string
	username    string
	email       string
	displayName string
	password    string
	tenantID    string
	role        iamtypes.Role
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
	if !iamtypes.IsValidRole(input.Roles[0]) {
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
		PasswordAlgo:       passwordAlgoArgon2id,
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
