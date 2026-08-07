package signature

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// fakeUserReader implements IamUserReader.
type fakeUserReader struct {
	users map[string][]byte // userID → bcrypt hash
}

func newFakeReader(users map[string]string) *fakeUserReader {
	hashes := make(map[string][]byte)
	for id, pw := range users {
		h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
		hashes[id] = h
	}
	return &fakeUserReader{users: hashes}
}

func (f *fakeUserReader) GetPasswordHash(_ context.Context, userID string) ([]byte, error) {
	h, ok := f.users[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return h, nil
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
