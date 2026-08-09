package application

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/auth/infrastructure/memory"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/passwordhash"
	"metaldocs/internal/platform/tenant"
)

// rehashFailingRepo wraps the in-memory Repository so RehashPassword always
// fails inside the login-lock transaction, without touching any other
// behavior. Used to prove a rehash write failure never fails the login
// itself (REQ-AUTHN-1 rehash-on-login: migration hiccups must not lock out a
// user who just authenticated correctly).
type rehashFailingRepo struct {
	*memory.Repository
}

func (r rehashFailingRepo) WithinLoginLock(ctx context.Context, userID string, fn func(authdomain.LoginTx) error) error {
	return r.Repository.WithinLoginLock(ctx, userID, func(tx authdomain.LoginTx) error {
		return fn(rehashFailingLoginTx{LoginTx: tx})
	})
}

type rehashFailingLoginTx struct {
	authdomain.LoginTx
}

func (rehashFailingLoginTx) RehashPassword(context.Context, string, authdomain.PasswordHash, string) error {
	return errors.New("simulated rehash write failure")
}

func seedIdentity(t *testing.T, repo authdomain.Repository, userID, hash, algo string) {
	t.Helper()
	if err := repo.CreateUser(context.Background(), authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        userID + "@example.com",
		DisplayName:  userID,
		PasswordHash: authdomain.PasswordHash(hash),
		PasswordAlgo: algo,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser(%s): %v", userID, err)
	}
}

// newArgon2TestService builds a Service with dev-tenant fallback enabled and
// seeds userID with a viewer role in the dev tenant, so Authenticate's
// post-credential-check tenant resolution succeeds (mirrors the existing
// timing/inactive tests' setup).
func newArgon2TestService(t *testing.T, repo authdomain.Repository, userID string) *Service {
	t.Helper()
	roleProvider := newMockRoleProvider()
	roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamtypes.Role{iamtypes.RoleViewer}
	return mustNewService(t, repo, roleProvider, newMockRoleAdminRepository(), Config{
		SessionCookieName:      "session",
		SessionTTL:             24 * time.Hour,
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 100,
		LoginLockDuration:      15 * time.Minute,
		AllowDevTenantFallback: true,
	})
}

// TestAuthenticate_RehashesBcryptToArgon2idOnLogin proves REQ-AUTHN-1's
// rehash-on-login migration: a user whose stored credential is still a
// bcrypt hash gets rehashed to Argon2id (memory-hard KDF) in the same
// locked transaction as a successful login, with both the algorithm and the
// hash bytes updated, and the new hash itself verifies.
func TestAuthenticate_RehashesBcryptToArgon2idOnLogin(t *testing.T) {
	const userID = "rehash-user"
	const password = "CorrectPassword123!"

	repo := memory.NewRepository()
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	seedIdentity(t, repo, userID, string(bcryptHash), passwordhash.AlgoBcrypt)

	svc := newArgon2TestService(t, repo, userID)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)

	if _, err := svc.Authenticate(context.Background(), userID, password, req); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	identity, err := repo.FindIdentityByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindIdentityByUserID: %v", err)
	}
	if identity.PasswordAlgo != passwordhash.AlgoArgon2id {
		t.Fatalf("PasswordAlgo = %q, want %q after rehash", identity.PasswordAlgo, passwordhash.AlgoArgon2id)
	}
	if string(identity.PasswordHash) == string(bcryptHash) {
		t.Fatal("PasswordHash unchanged after rehash, want a new argon2id hash")
	}
	ok, err := passwordhash.VerifyArgon2id(string(identity.PasswordHash), []byte(password))
	if err != nil {
		t.Fatalf("VerifyArgon2id(rehashed): %v", err)
	}
	if !ok {
		t.Fatal("rehashed argon2id hash does not verify the original password")
	}
}

// TestAuthenticate_Argon2idIdentityNotRehashed proves an already-migrated
// identity is left untouched: no gratuitous rehash work on every login once
// a user is on the target algorithm.
func TestAuthenticate_Argon2idIdentityNotRehashed(t *testing.T) {
	const userID = "argon2id-user"
	const password = "CorrectPassword123!"

	repo := memory.NewRepository()
	argonHash, err := passwordhash.HashArgon2id([]byte(password))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	seedIdentity(t, repo, userID, argonHash, passwordhash.AlgoArgon2id)

	svc := newArgon2TestService(t, repo, userID)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)

	if _, err := svc.Authenticate(context.Background(), userID, password, req); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	identity, err := repo.FindIdentityByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindIdentityByUserID: %v", err)
	}
	if identity.PasswordAlgo != passwordhash.AlgoArgon2id {
		t.Fatalf("PasswordAlgo = %q, want unchanged %q", identity.PasswordAlgo, passwordhash.AlgoArgon2id)
	}
	if string(identity.PasswordHash) != argonHash {
		t.Fatal("PasswordHash changed even though identity was already argon2id")
	}
}

// TestAuthenticate_UnrecognizedAlgoFailsClosed proves the no-fallback
// principle at the login boundary: an identity whose stored password_algo
// is neither "bcrypt" nor "argon2id" must fail the login (never match, never
// fall through to a different comparison), and the failure is
// indistinguishable in kind from an ordinary wrong-password rejection.
func TestAuthenticate_UnrecognizedAlgoFailsClosed(t *testing.T) {
	const userID = "unknown-algo-user"
	const password = "CorrectPassword123!"

	repo := memory.NewRepository()
	// The hash bytes are irrelevant here: Verify must fail closed on the
	// algorithm discriminator before ever attempting a comparison.
	seedIdentity(t, repo, userID, "irrelevant-hash-bytes", "scrypt")

	svc := newArgon2TestService(t, repo, userID)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)

	_, err := svc.Authenticate(context.Background(), userID, password, req)
	if !errors.Is(err, authdomain.ErrInvalidCredentials) {
		t.Fatalf("Authenticate error = %v, want ErrInvalidCredentials (fail closed)", err)
	}
}

// TestAuthenticate_RehashWriteFailureStillSucceedsLogin proves a rehash
// persistence failure is non-fatal to the login: the user already
// authenticated correctly against the credential on file, so a migration
// write hiccup must not turn a correct login into a failure.
func TestAuthenticate_RehashWriteFailureStillSucceedsLogin(t *testing.T) {
	const userID = "rehash-write-fail-user"
	const password = "CorrectPassword123!"

	inner := memory.NewRepository()
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	seedIdentity(t, inner, userID, string(bcryptHash), passwordhash.AlgoBcrypt)

	repo := rehashFailingRepo{Repository: inner}
	svc := newArgon2TestService(t, repo, userID)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)

	if _, err := svc.Authenticate(context.Background(), userID, password, req); err != nil {
		t.Fatalf("Authenticate must succeed despite rehash write failure, got: %v", err)
	}

	identity, err := inner.FindIdentityByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindIdentityByUserID: %v", err)
	}
	if identity.PasswordAlgo != passwordhash.AlgoBcrypt {
		t.Fatalf("PasswordAlgo = %q, want unchanged %q (rehash write failed)", identity.PasswordAlgo, passwordhash.AlgoBcrypt)
	}
}
