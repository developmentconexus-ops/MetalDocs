package main

import (
	"context"
	"database/sql"
	"log/slog"

	"metaldocs/internal/modules/approval/infrastructure/signature"
	authdomain "metaldocs/internal/modules/auth/domain"
)

// authPasswordHashReader adapts the auth module's identity repository to the
// approval signature port. The acting user's password hash AND its algorithm
// discriminator (REQ-AUTHN-1: "bcrypt" or "argon2id") are resolved through the
// auth port (auth_identities) — the approval module never reaches across the
// boundary with raw SQL, and never assumes a single hashing algorithm.
type authPasswordHashReader struct {
	repo authdomain.Repository
}

func (r authPasswordHashReader) GetPasswordHash(ctx context.Context, userID string) ([]byte, string, error) {
	identity, err := r.repo.FindIdentityByUserID(ctx, userID)
	if err != nil {
		// User missing / lookup failure → the provider maps this to the same
		// disclosure-safe invalid-credentials error as a wrong password.
		return nil, "", err
	}
	return []byte(identity.PasswordHash), identity.PasswordAlgo, nil
}

// slogReauthEmitter records failed sign-off re-authentication attempts. The
// throttle + the rejected request already carry the security signal; this adds
// structured observability without widening the audit-event schema.
type slogReauthEmitter struct{}

func (slogReauthEmitter) EmitAuthFailed(ctx context.Context, actorUserID, reason string) {
	slog.WarnContext(ctx, "signoff re-authentication failed",
		slog.String("actor_user_id", actorUserID),
		slog.String("reason", reason),
	)
}

// newSignoffReauthRegistry builds the e-signature verifier registry wired into
// the approval DecisionService. The Postgres-backed rate limiter ensures lockout
// state survives API restarts and is shared across replicas (F-20e, REQ-REL-3, D-1).
func newSignoffReauthRegistry(repo authdomain.Repository, db *sql.DB) *signature.Registry {
	registry := signature.NewRegistry()
	registry.Register(signature.NewPasswordReauthProvider(
		authPasswordHashReader{repo: repo},
		slogReauthEmitter{},
		signature.NewPostgresAuthFailureRateLimiter(db),
	))
	return registry
}
