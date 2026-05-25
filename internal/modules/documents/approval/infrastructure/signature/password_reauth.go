package signature

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("signature: invalid credentials")
	ErrRateLimited        = errors.New("signature: too many failed attempts, try again later")
	ErrRateLimiterConfig  = errors.New("signature: auth-failure rate limiter not configured")
)

// IamUserReader abstracts password-hash lookup for testability.
type IamUserReader interface {
	GetPasswordHash(ctx context.Context, userID string) ([]byte, error)
}

// EventEmitterStub abstracts audit-event emission for Phase 3 (Phase 5 wires real one).
type EventEmitterStub interface {
	EmitAuthFailed(ctx context.Context, actorUserID, reason string)
}

// AuthFailureRateLimiter controls password failure attempts.
// Implementations may use shared state (redis/postgres) for cross-replica limits.
type AuthFailureRateLimiter interface {
	Allow(ctx context.Context, actorID string) (bool, error)
	RecordFailure(ctx context.Context, actorID string) error
	Reset(ctx context.Context, actorID string) error
}

const (
	maxFailures   = 5
	windowDur     = 60 * time.Second
)

type failEntry struct {
	count  int
	oldest time.Time // oldest failure in window
	lastAt time.Time
}

// PasswordReauthProvider implements Provider using bcrypt against iam_users.password_hash.
type PasswordReauthProvider struct {
	reader  IamUserReader
	emitter EventEmitterStub
	limiter AuthFailureRateLimiter
}

// NewPasswordReauthProvider creates the provider.
func NewPasswordReauthProvider(_ context.Context, reader IamUserReader, emitter EventEmitterStub, limiter AuthFailureRateLimiter) *PasswordReauthProvider {
	return &PasswordReauthProvider{
		reader:  reader,
		emitter: emitter,
		limiter: limiter,
	}
}

func (p *PasswordReauthProvider) Method() string { return "password_reauth" }

func (p *PasswordReauthProvider) Sign(ctx context.Context, req SignRequest) (SignatureResult, error) {
	password, ok := req.Credentials["password"]
	if !ok || password == "" {
		return SignatureResult{}, ErrInvalidCredentials
	}

	if p.limiter == nil {
		return SignatureResult{}, ErrRateLimiterConfig
	}

	allowed, err := p.limiter.Allow(ctx, req.ActorUserID)
	if err != nil || !allowed {
		return SignatureResult{}, ErrRateLimited
	}

	hash, err := p.reader.GetPasswordHash(ctx, req.ActorUserID)
	if err != nil {
		// User missing → same error as wrong password (disclosure-safe).
		if err := p.limiter.RecordFailure(ctx, req.ActorUserID); err != nil {
			return SignatureResult{}, ErrRateLimited
		}
		if p.emitter != nil {
			p.emitter.EmitAuthFailed(ctx, req.ActorUserID, "user_not_found")
		}
		return SignatureResult{}, ErrInvalidCredentials
	}

	cost, _ := bcrypt.Cost(hash)
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		if err := p.limiter.RecordFailure(ctx, req.ActorUserID); err != nil {
			return SignatureResult{}, ErrRateLimited
		}
		if p.emitter != nil {
			p.emitter.EmitAuthFailed(ctx, req.ActorUserID, "wrong_password")
		}
		return SignatureResult{}, ErrInvalidCredentials
	}

	// Success — clear failure state.
	if err := p.limiter.Reset(ctx, req.ActorUserID); err != nil {
		return SignatureResult{}, ErrRateLimited
	}

	now := time.Now().UTC()
	payload, _ := json.Marshal(map[string]any{
		"method":      "password_reauth",
		"bcrypt_cost": cost,
		"verified_at": now.Format(time.RFC3339),
	})
	return SignatureResult{Method: "password_reauth", Payload: payload, SignedAt: now}, nil
}

// InMemoryAuthFailureRateLimiter is process-local and intended for tests/dev only.
// Production should provide a shared implementation.
type InMemoryAuthFailureRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*failEntry
}

func NewInMemoryAuthFailureRateLimiter() *InMemoryAuthFailureRateLimiter {
	return &InMemoryAuthFailureRateLimiter{entries: make(map[string]*failEntry)}
}

func (l *InMemoryAuthFailureRateLimiter) Allow(_ context.Context, actorID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[actorID]
	if !ok {
		return true, nil
	}
	now := time.Now()
	if now.Sub(e.oldest) >= windowDur {
		delete(l.entries, actorID)
		return true, nil
	}
	return e.count < maxFailures, nil
}

func (l *InMemoryAuthFailureRateLimiter) RecordFailure(_ context.Context, actorID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.entries[actorID]
	if !ok {
		l.entries[actorID] = &failEntry{count: 1, oldest: now, lastAt: now}
		return nil
	}
	if now.Sub(e.oldest) >= windowDur {
		e.count = 1
		e.oldest = now
	} else {
		e.count++
	}
	e.lastAt = now
	return nil
}

func (l *InMemoryAuthFailureRateLimiter) Reset(_ context.Context, actorID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, actorID)
	return nil
}
