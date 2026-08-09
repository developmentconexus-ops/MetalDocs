package signature

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"metaldocs/internal/platform/passwordhash"
)

// bcryptCost reports the cost factor of a bcrypt hash, for the legacy
// "bcrypt_cost" attestation field only (never for verification).
func bcryptCost(hash []byte) (int, error) {
	return bcrypt.Cost(hash)
}

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

// IamUserReader abstracts password-hash lookup for testability. It returns
// the stored hash bytes plus the password_algo discriminator ("bcrypt" or
// "argon2id") the hash was minted with (REQ-AUTHN-1) — Sign dispatches its
// comparison on the algo, never assumes a single algorithm.
type IamUserReader interface {
	GetPasswordHash(ctx context.Context, userID string) ([]byte, string, error)
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

// Sign verifies req.Credentials["password"] against the actor's stored
// password hash in iam_users, dispatching the comparison on the stored
// password_algo (REQ-AUTHN-1: bcrypt or argon2id — never guessed, never
// falls through to a different algorithm on an unrecognized value), subject
// to the wired AuthFailureRateLimiter. It fails closed: ErrRateLimiterConfig
// if no limiter is wired, ErrRateLimited if the limiter denies or errors,
// and ErrInvalidCredentials for any lookup, unrecognized-algorithm, or
// compare failure (never distinguishing "user not found" from "wrong
// password" from "unrecognized algorithm"). On success it resets the rate
// limiter and returns an opaque attestation — the raw password is never
// included in Result.Payload.
func (p *PasswordReauthProvider) Sign(ctx context.Context, req SignRequest) (Result, error) {
	password, ok := req.Credentials["password"]
	if !ok || password == "" {
		return Result{}, ErrInvalidCredentials
	}

	if p.limiter == nil {
		return Result{}, ErrRateLimiterConfig
	}
	if err := p.checkAuthFailureLimiter(ctx, req.ActorUserID); err != nil {
		return Result{}, err
	}

	hash, algo, err := p.verifyPassword(ctx, req.ActorUserID, password)
	if err != nil {
		return Result{}, err
	}

	// Success — clear failure state.
	if err := p.limiter.Reset(ctx, req.ActorUserID); err != nil {
		return Result{}, ErrRateLimited
	}

	now := time.Now().UTC()
	attestation := signAttestation(algo, hash, now)
	payload, _ := json.Marshal(attestation)
	return Result{Method: "password_reauth", Payload: payload, SignedAt: now}, nil
}

// checkAuthFailureLimiter reports whether the actor is currently allowed to
// attempt a re-auth, failing closed (ErrRateLimited) on any limiter error.
func (p *PasswordReauthProvider) checkAuthFailureLimiter(ctx context.Context, actorUserID string) error {
	allowed, err := p.limiter.Allow(ctx, actorUserID)
	if err != nil {
		slog.ErrorContext(ctx, "signature: auth-failure limiter Allow failed; failing closed",
			"actor_user_id", actorUserID, "err", err)
		return ErrRateLimited
	}
	if !allowed {
		return ErrRateLimited
	}
	return nil
}

// verifyPassword looks up the actor's stored password hash and verifies
// password against it, recording a rate-limiter failure and emitting an
// audit event on any lookup or verification failure. Both collapse to
// ErrInvalidCredentials (disclosure-safe — never reveals which case fired).
// On success it returns the raw stored hash and algo; the caller resets the
// rate limiter itself once every subsequent step also succeeds.
func (p *PasswordReauthProvider) verifyPassword(ctx context.Context, actorUserID, password string) ([]byte, string, error) {
	hash, algo, err := p.reader.GetPasswordHash(ctx, actorUserID)
	if err != nil {
		// User missing → same error as wrong password (disclosure-safe).
		if err := p.limiter.RecordFailure(ctx, actorUserID); err != nil {
			return nil, "", ErrRateLimited
		}
		if p.emitter != nil {
			p.emitter.EmitAuthFailed(ctx, actorUserID, "user_not_found")
		}
		return nil, "", ErrInvalidCredentials
	}

	// passwordOK stays false (and verifyErr is swallowed into a generic
	// failure) for both a wrong password and an unrecognized/malformed
	// stored algorithm — no-fallback principle: an unrecognized password_algo
	// never matches and never tries a different comparison.
	passwordOK, verifyErr := passwordhash.Verify(algo, hash, []byte(password))
	if verifyErr != nil {
		passwordOK = false
	}
	if !passwordOK {
		if err := p.limiter.RecordFailure(ctx, actorUserID); err != nil {
			return nil, "", ErrRateLimited
		}
		if p.emitter != nil {
			reason := "wrong_password"
			if verifyErr != nil {
				reason = "unverifiable_credential"
			}
			p.emitter.EmitAuthFailed(ctx, actorUserID, reason)
		}
		return nil, "", ErrInvalidCredentials
	}
	return hash, algo, nil
}

// signAttestation builds the opaque attestation payload recorded for a
// successful Sign, including the KDF cost actually spent, per algorithm —
// "bcrypt_cost" is kept unchanged (backward-compatible field name/shape) for
// actors still on the legacy algorithm; "argon2_params" is the analogous
// record for the target algorithm (REQ-AUTHN-1). Only one of the two is ever
// present, matching the algo that verified.
func signAttestation(algo string, hash []byte, now time.Time) map[string]any {
	attestation := map[string]any{
		"method":      "password_reauth",
		"algo":        algo,
		"verified_at": now.Format(time.RFC3339),
	}
	switch algo {
	case passwordhash.AlgoBcrypt:
		if cost, costErr := bcryptCost(hash); costErr == nil {
			attestation["bcrypt_cost"] = cost
		}
	case passwordhash.AlgoArgon2id:
		if summary, summaryErr := passwordhash.Argon2ParamsSummary(string(hash)); summaryErr == nil {
			attestation["argon2_params"] = summary
		}
	}
	return attestation
}
