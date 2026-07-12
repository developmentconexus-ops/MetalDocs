package signature

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned for a missing/wrong password or an unknown actor
	// (same error in both cases — disclosure-safe, never reveals which).
	ErrInvalidCredentials = errors.New("signature: invalid credentials")
	// ErrRateLimited is returned when the actor has exceeded maxFailures within windowDur,
	// or when the limiter itself fails (fail-closed).
	ErrRateLimited = errors.New("signature: too many failed attempts, try again later")
	// ErrRateLimiterConfig is returned when Sign is called without an AuthFailureRateLimiter
	// wired — password_reauth must never run unrate-limited.
	ErrRateLimiterConfig = errors.New("signature: auth-failure rate limiter not configured")
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
	maxFailures = 5
	windowDur   = 60 * time.Second
)

// PasswordReauthProvider implements Provider using bcrypt against iam_users.password_hash.
type PasswordReauthProvider struct {
	reader  IamUserReader
	emitter EventEmitterStub
	limiter AuthFailureRateLimiter
}

// NewPasswordReauthProvider creates the provider.
func NewPasswordReauthProvider(reader IamUserReader, emitter EventEmitterStub, limiter AuthFailureRateLimiter) *PasswordReauthProvider {
	return &PasswordReauthProvider{
		reader:  reader,
		emitter: emitter,
		limiter: limiter,
	}
}

// Method returns "password_reauth", identifying this Provider in the Registry.
func (p *PasswordReauthProvider) Method() string { return "password_reauth" }

// Sign verifies req.Credentials["password"] against the actor's bcrypt hash in
// iam_users, subject to the wired AuthFailureRateLimiter. It fails closed:
// ErrRateLimiterConfig if no limiter is wired, ErrRateLimited if the limiter
// denies or errors, and ErrInvalidCredentials for any lookup or compare
// failure (never distinguishing "user not found" from "wrong password").
// On success it resets the rate limiter and returns an opaque attestation —
// the raw password is never included in SignatureResult.Payload.
func (p *PasswordReauthProvider) Sign(ctx context.Context, req SignRequest) (SignatureResult, error) {
	password, ok := req.Credentials["password"]
	if !ok || password == "" {
		return SignatureResult{}, ErrInvalidCredentials
	}

	if p.limiter == nil {
		return SignatureResult{}, ErrRateLimiterConfig
	}

	allowed, err := p.limiter.Allow(ctx, req.ActorUserID)
	if err != nil {
		slog.ErrorContext(ctx, "signature: auth-failure limiter Allow failed; failing closed",
			"actor_user_id", req.ActorUserID, "err", err)
		return SignatureResult{}, ErrRateLimited
	}
	if !allowed {
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
