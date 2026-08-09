package application

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/auth/infrastructure/memory"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/tenant"
)

// This file covers REQ-AUTHN-3 clauses nothing else in the package exercises
// directly: token opacity (no claims), CSPRNG-generated entropy, and the
// stored form being a hash rather than the raw token. See ADR 0094 for the
// design ruling these tests are cited from.

// TestSessionToken_OpaqueNoClaims_REQ_AUTHN_3 proves the session token is an
// undifferentiated random byte string, not a self-describing/decodable
// structure. A JWT (or any claims-bearing token) is three base64url segments
// separated by '.', with the middle segment decoding to a JSON object. This
// token is exactly one random segment plus one HMAC segment — decoding the
// first segment must NOT yield JSON, because there are no claims to carry.
func TestSessionToken_OpaqueNoClaims_REQ_AUTHN_3(t *testing.T) {
	svc := mustNewService(t, memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), Config{
		SessionSecret: testSessionSecret,
	})

	cookieValue, _, err := svc.newSessionToken()
	if err != nil {
		t.Fatalf("newSessionToken: %v", err)
	}

	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 {
		t.Fatalf("cookie value has %d dot-separated segments, want exactly 2 (token.signature) — a JWT would have 3", len(parts))
	}

	rawToken, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("token segment is not valid base64url: %v", err)
	}
	if len(rawToken) != 32 {
		t.Fatalf("decoded token is %d bytes, want 32 (raw CSPRNG bytes, not an encoded structure)", len(rawToken))
	}
	// A claims-bearing segment (JWT payload) decodes to a JSON object. Raw
	// random bytes must not.
	var probe map[string]any
	if err := json.Unmarshal(rawToken, &probe); err == nil {
		t.Fatalf("token segment decoded as JSON (%v) — it must be opaque random bytes, not claims", probe)
	}
}

// TestSessionToken_CSPRNGSufficientEntropy_REQ_AUTHN_3 proves token material
// is generated fresh from crypto/rand on every call (never derived,
// sequential, or reused) and carries the full 256 bits (32 bytes) the design
// claims. 64 draws with zero collisions and zero repeated byte-runs is not a
// statistical proof of randomness, but it is a real regression guard against
// the two failure modes that would actually happen by accident: a
// mis-wired/zeroed buffer, or a counter standing in for crypto/rand.
func TestSessionToken_CSPRNGSufficientEntropy_REQ_AUTHN_3(t *testing.T) {
	svc := mustNewService(t, memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), Config{
		SessionSecret: testSessionSecret,
	})

	const draws = 64
	seen := make(map[string]bool, draws)
	for i := 0; i < draws; i++ {
		cookieValue, hash, err := svc.newSessionToken()
		if err != nil {
			t.Fatalf("newSessionToken[%d]: %v", i, err)
		}
		parts := strings.Split(cookieValue, ".")
		if len(parts) != 2 {
			t.Fatalf("newSessionToken[%d]: malformed cookie value %q", i, cookieValue)
		}
		rawToken, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			t.Fatalf("newSessionToken[%d]: token segment not base64url: %v", i, err)
		}
		if len(rawToken) != 32 {
			t.Fatalf("newSessionToken[%d]: raw token is %d bytes, want 32 (256 bits)", i, len(rawToken))
		}
		allZero := true
		for _, b := range rawToken {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Fatalf("newSessionToken[%d]: raw token is all-zero — CSPRNG not wired", i)
		}
		if seen[parts[0]] {
			t.Fatalf("newSessionToken[%d]: token segment repeated across draws — not fresh CSPRNG output", i)
		}
		seen[parts[0]] = true
		if seen[hash] {
			t.Fatalf("newSessionToken[%d]: stored hash repeated across draws", i)
		}
		seen[hash] = true
	}
}

// TestSessionToken_StoredFormIsHashNeverRaw_REQ_AUTHN_3 proves the value
// newSessionToken returns for server-side persistence is sha256(rawToken),
// never the raw token itself or the signed cookie value — so a leaked DB row
// (backup, replica, log of a query) cannot be replayed as a session token,
// and cannot be reversed back to the bearer credential.
func TestSessionToken_StoredFormIsHashNeverRaw_REQ_AUTHN_3(t *testing.T) {
	svc := mustNewService(t, memory.NewRepository(), newMockRoleProvider(), newMockRoleAdminRepository(), Config{
		SessionSecret: testSessionSecret,
	})

	cookieValue, storedHash, err := svc.newSessionToken()
	if err != nil {
		t.Fatalf("newSessionToken: %v", err)
	}
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 {
		t.Fatalf("malformed cookie value %q", cookieValue)
	}
	rawTokenSegment := parts[0]

	if storedHash == rawTokenSegment {
		t.Fatal("stored form equals the raw token segment — must be a one-way hash")
	}
	if storedHash == cookieValue {
		t.Fatal("stored form equals the full cookie value — must be a one-way hash of the token alone")
	}

	// hashToken hashes the base64url token *string* itself (service.go:1235-1238),
	// not the decoded raw bytes — match that exactly.
	sum := sha256.Sum256([]byte(rawTokenSegment))
	want := hex.EncodeToString(sum[:])
	if storedHash != want {
		t.Fatalf("stored form = %q, want sha256(rawToken) = %q — stored form must be exactly the hash, not some other derivation", storedHash, want)
	}
}

// TestLogout_RevokesSession_REQ_AUTHN_3 proves a single session is revocable
// server-side on demand: a token that resolved successfully before Logout
// must fail to resolve afterward. Bulk revocation (password change, account
// deactivation) is already covered by TestChangePasswordForUser_RevokesSessions
// and TestUpdateUser_DeactivateRevokesSessions; this is the single-session path.
func TestLogout_RevokesSession_REQ_AUTHN_3(t *testing.T) {
	repo := memory.NewRepository()
	roleProvider := newMockRoleProvider()
	roleAdmin := newMockRoleAdminRepository()
	ctx := context.Background()

	userID := "logout-revoke-user"
	password := "TestPassword123!"
	if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       userID,
		Username:     userID,
		Email:        "logout-revoke@example.com",
		DisplayName:  "Logout Revoke User",
		PasswordHash: mustHashPassword(t, password),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	roleProvider.roles[userID+":"+tenant.DevTenantID] = []iamtypes.Role{iamtypes.RoleViewer}

	svc := mustNewService(t, repo, roleProvider, roleAdmin, Config{
		SessionCookieName:      "session",
		SessionSecret:          testSessionSecret,
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      15 * time.Minute,
		// SessionTTL was previously unset (zero value), which sets
		// ExpiresAt == CreatedAt. That is not "immediately expired" by
		// construction -- Before() is a strict less-than -- but it leaves
		// zero headroom: ExpiresAt and the ResolveSession-time now() are two
		// independent time.Now() reads with the monotonic component
		// stripped by .UTC() (see service.go), so they only need to land on
		// different wall-clock ticks (a certainty once anything schedules a
		// goroutine between them, e.g. contention from a concurrent `go
		// build`) for ExpiresAt.Before(now) to flip true and this test to
		// fail on ErrSessionExpired before it ever reaches the revocation
		// assertion it exists to prove. Giving the session a real TTL, like
		// every other session-bearing test in this package, removes that
		// race entirely and lets the test verify what it claims to verify.
		SessionTTL:             24 * time.Hour,
		AllowDevTenantFallback: true,
		CookieSecure:           false,
	})

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	authSession, err := svc.Authenticate(ctx, userID, password, req)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if _, err := svc.ResolveSession(ctx, authSession.RawToken); err != nil {
		t.Fatalf("ResolveSession before logout: %v", err)
	}

	if err := svc.Logout(ctx, authSession.RawToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if _, err := svc.ResolveSession(ctx, authSession.RawToken); !errors.Is(err, authdomain.ErrSessionRevoked) {
		t.Fatalf("ResolveSession after logout: err = %v, want ErrSessionRevoked", err)
	}
}
