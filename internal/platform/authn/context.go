package authn

import (
	"context"
	"errors"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// ErrMissingActor is the explicit "no authenticated principal" failure. It
// exists so a consumer below the delivery layer — an application service, a
// ctx-resolving infrastructure seam — can refuse to proceed WITHOUT importing
// net/http: absence becomes a propagated error rather than an empty string
// that keeps travelling (A3.3).
//
// Delivery layers map it to the same 401 + problem.CodeAuthUnauthenticated an
// authenticated route already answers with when the session is absent, so the
// wire contract does not fork.
//
// The message leads with "authentication required" because several modules'
// error mappers use err.Error() as the problem `title` for a 4xx: the sentinel
// therefore has to read as a client-facing reason, not only as a log line. The
// clause after the colon is the log half — it says which of the two ways the
// actor can be absent (no auth context at all, or a blank/whitespace one)
// without the reader having to guess.
var ErrMissingActor = errors.New("authentication required: request context carries no authenticated actor")

// UserIDFromContext resolves the authenticated user identity from the request
// context. The second return value is true when an IAM auth context was
// installed by middleware and yielded a non-empty (after trim) user id;
// false signals "no authenticated principal" so callers can fail-closed
// instead of silently treating an empty id as a permissive default.
//
// This is the canonical presence-aware accessor: it centralises the trim +
// empty-check that every delivery handler needs, so a missing principal is one
// uniform fail-closed decision rather than 27 hand-rolled copies. (The bare
// iamdomain.UserIDFromContext returns only the string and pushes that check
// onto each caller.)
func UserIDFromContext(ctx context.Context) (string, bool) {
	raw := strings.TrimSpace(iamdomain.UserIDFromContext(ctx))
	if raw == "" {
		return "", false
	}
	return raw, true
}

// RequireUserID is UserIDFromContext for the actor-REQUIRED consumer: it turns
// absence into ErrMissingActor instead of a boolean the caller may forget to
// read. `id, _ := UserIDFromContext(ctx)` compiles and yields ""; `id, err :=
// RequireUserID(ctx)` does not compile into a usable id without disposing of
// err, and the repository's errcheck settings refuse a silently dropped one.
//
// Use it wherever the actor is an input to authorization, audit attribution,
// a domain mutation, or persistence. Use UserIDFromContext directly only where
// absence is a documented, deliberate alternative policy (A3.3 class B).
func RequireUserID(ctx context.Context) (string, error) {
	id, ok := UserIDFromContext(ctx)
	if !ok {
		return "", ErrMissingActor
	}
	return id, nil
}
