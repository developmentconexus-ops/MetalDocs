package signature

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"metaldocs/internal/platform/passwordhash"
)

// fakeUserReader implements IamUserReader.
type fakeUserReaderEntry struct {
	hash []byte
	algo string
}

type fakeUserReader struct {
	users map[string]fakeUserReaderEntry
}

// newFakeReader seeds every user with a bcrypt hash (the historical default
// for this fake, kept so pre-existing tests are unaffected).
func newFakeReader(users map[string]string) *fakeUserReader {
	entries := make(map[string]fakeUserReaderEntry, len(users))
	for id, pw := range users {
		h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
		entries[id] = fakeUserReaderEntry{hash: h, algo: passwordhash.AlgoBcrypt}
	}
	return &fakeUserReader{users: entries}
}

// newFakeReaderArgon2id seeds every user with an Argon2id hash, for tests
// that exercise the target algorithm (REQ-AUTHN-1).
func newFakeReaderArgon2id(users map[string]string) *fakeUserReader {
	entries := make(map[string]fakeUserReaderEntry, len(users))
	for id, pw := range users {
		h, _ := passwordhash.HashArgon2id([]byte(pw))
		entries[id] = fakeUserReaderEntry{hash: []byte(h), algo: passwordhash.AlgoArgon2id}
	}
	return &fakeUserReader{users: entries}
}

// setAlgo overrides the stored algo discriminator for userID without
// touching the hash bytes — used to simulate an unrecognized/corrupt
// password_algo value.
func (f *fakeUserReader) setAlgo(userID, algo string) {
	e := f.users[userID]
	e.algo = algo
	f.users[userID] = e
}

func (f *fakeUserReader) GetPasswordHash(_ context.Context, userID string) ([]byte, string, error) {
	e, ok := f.users[userID]
	if !ok {
		return nil, "", errors.New("not found")
	}
	return e.hash, e.algo, nil
}

// fakeEmitter records auth failures.
type fakeEmitter struct {
	failures []string
}

func (e *fakeEmitter) EmitAuthFailed(_ context.Context, actorID, _ string) {
	e.failures = append(e.failures, actorID)
}

// testInMemoryLimiter is a test-only process-local rate limiter.
// Not for production use — InMemoryAuthFailureRateLimiter was deleted (2.13 step 1).
type testInMemoryLimiter struct {
	mu      sync.Mutex
	entries map[string]*struct {
		count  int
		oldest time.Time
	}
}

func newTestLimiter() *testInMemoryLimiter {
	return &testInMemoryLimiter{entries: make(map[string]*struct {
		count  int
		oldest time.Time
	})}
}

func (l *testInMemoryLimiter) Allow(_ context.Context, actorID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[actorID]
	if !ok {
		return true, nil
	}
	if time.Since(e.oldest) >= windowDur {
		delete(l.entries, actorID)
		return true, nil
	}
	return e.count < maxFailures, nil
}

func (l *testInMemoryLimiter) RecordFailure(_ context.Context, actorID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e, ok := l.entries[actorID]
	if !ok {
		l.entries[actorID] = &struct {
			count  int
			oldest time.Time
		}{count: 1, oldest: now}
		return nil
	}
	if now.Sub(e.oldest) >= windowDur {
		e.count = 1
		e.oldest = now
	} else {
		e.count++
	}
	return nil
}

func (l *testInMemoryLimiter) Reset(_ context.Context, actorID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, actorID)
	return nil
}

func newProvider(users map[string]string) (*PasswordReauthProvider, *fakeEmitter) {
	em := &fakeEmitter{}
	p := NewPasswordReauthProvider(newFakeReader(users), em, newTestLimiter())
	return p, em
}

func TestPasswordReauthHappy(t *testing.T) {
	p, _ := newProvider(map[string]string{"u1": "secret123"})
	res, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "u1", ActorTenantID: "t1",
		ContentHash: "abc", Credentials: map[string]string{"password": "secret123"},
	})
	if err != nil {
		t.Fatalf("happy path: %v", err)
	}
	if res.Method != "password_reauth" {
		t.Error("method mismatch")
	}
}

func TestPasswordReauthWrongPassword(t *testing.T) {
	p, em := newProvider(map[string]string{"u1": "correct"})
	_, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "u1", Credentials: map[string]string{"password": "wrong"},
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("want ErrInvalidCredentials; got %v", err)
	}
	if len(em.failures) == 0 {
		t.Error("failure event should be emitted")
	}
}

func TestPasswordReauthMissingUser(t *testing.T) {
	p, _ := newProvider(map[string]string{})
	_, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "ghost", Credentials: map[string]string{"password": "pw"},
	})
	// Must return same error as wrong password — no disclosure.
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("missing user: want ErrInvalidCredentials; got %v", err)
	}
}

func TestPasswordReauthRateLimitTrip(t *testing.T) {
	p, _ := newProvider(map[string]string{"u1": "correct"})
	req := SignRequest{ActorUserID: "u1", Credentials: map[string]string{"password": "wrong"}}

	for i := 0; i < maxFailures; i++ {
		p.Sign(context.Background(), req) //nolint
	}
	// 6th attempt should be rate-limited.
	_, err := p.Sign(context.Background(), req)
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("want ErrRateLimited after %d failures; got %v", maxFailures, err)
	}
}

func TestPasswordReauthFailsClosedWhenLimiterMissing(t *testing.T) {
	em := &fakeEmitter{}
	p := NewPasswordReauthProvider(newFakeReader(map[string]string{"u1": "secret"}), em, nil)
	_, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "u1",
		Credentials: map[string]string{"password": "secret"},
	})
	if !errors.Is(err, ErrRateLimiterConfig) {
		t.Fatalf("error = %v, want %v", err, ErrRateLimiterConfig)
	}
}

// TestPasswordReauthArgon2idHappy proves the reauth path verifies an
// Argon2id-hashed actor (REQ-AUTHN-1) and records "argon2_params" (not
// "bcrypt_cost") in the attestation payload.
func TestPasswordReauthArgon2idHappy(t *testing.T) {
	em := &fakeEmitter{}
	p := NewPasswordReauthProvider(newFakeReaderArgon2id(map[string]string{"u1": "secret123"}), em, newTestLimiter())
	res, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "u1", ActorTenantID: "t1",
		ContentHash: "abc", Credentials: map[string]string{"password": "secret123"},
	})
	if err != nil {
		t.Fatalf("argon2id happy path: %v", err)
	}
	payload := string(res.Payload)
	if !strings.Contains(payload, `"algo":"argon2id"`) {
		t.Errorf("payload = %s, want algo=argon2id", payload)
	}
	if !strings.Contains(payload, "argon2_params") {
		t.Errorf("payload = %s, want argon2_params present", payload)
	}
	if strings.Contains(payload, "bcrypt_cost") {
		t.Errorf("payload = %s, want no bcrypt_cost for an argon2id actor", payload)
	}
}

// TestPasswordReauthBcryptAttestationShape proves the legacy "bcrypt_cost"
// attestation field is preserved unchanged in shape for actors still on
// bcrypt, alongside the new "algo" field.
func TestPasswordReauthBcryptAttestationShape(t *testing.T) {
	p, _ := newProvider(map[string]string{"u1": "secret123"})
	res, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "u1", Credentials: map[string]string{"password": "secret123"},
	})
	if err != nil {
		t.Fatalf("bcrypt happy path: %v", err)
	}
	payload := string(res.Payload)
	if !strings.Contains(payload, `"algo":"bcrypt"`) {
		t.Errorf("payload = %s, want algo=bcrypt", payload)
	}
	if !strings.Contains(payload, "bcrypt_cost") {
		t.Errorf("payload = %s, want bcrypt_cost present", payload)
	}
}

// TestPasswordReauthUnrecognizedAlgoFailsClosed proves the no-fallback
// principle at the reauth boundary: an actor whose stored password_algo is
// neither "bcrypt" nor "argon2id" must fail sign-off, never match, and never
// fall through to a different comparison.
func TestPasswordReauthUnrecognizedAlgoFailsClosed(t *testing.T) {
	reader := newFakeReader(map[string]string{"u1": "secret123"})
	reader.setAlgo("u1", "scrypt")
	em := &fakeEmitter{}
	p := NewPasswordReauthProvider(reader, em, newTestLimiter())
	_, err := p.Sign(context.Background(), SignRequest{
		ActorUserID: "u1", Credentials: map[string]string{"password": "secret123"},
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials (fail closed)", err)
	}
	if len(em.failures) == 0 {
		t.Error("failure event should be emitted for an unrecognized algorithm")
	}
}
