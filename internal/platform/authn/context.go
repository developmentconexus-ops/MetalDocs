package authn

import (
	"context"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// UserIDFromContext resolves the authenticated user identity from the request
// context. The second return value is true when an IAM auth context was
// installed by middleware and yielded a non-empty (after trim) user id;
// false signals "no authenticated principal" so callers can fail-closed
// instead of silently treating an empty id as a permissive default.
//
// Deprecated: callers should prefer iamdomain.UserIDFromContext combined with
// their own presence handling. This wrapper is retained until M6 cleanup.
func UserIDFromContext(ctx context.Context) (string, bool) {
	raw := strings.TrimSpace(iamdomain.UserIDFromContext(ctx))
	if raw == "" {
		return "", false
	}
	return raw, true
}

// RolesFromContext resolves authenticated roles as normalized lowercase strings.
//
// Deprecated: see M6 — callers should consume iamdomain.RolesFromContext
// directly to preserve the iamdomain.Role enum.
func RolesFromContext(ctx context.Context) []string {
	roles := iamdomain.RolesFromContext(ctx)
	if len(roles) == 0 {
		return nil
	}
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		normalized := strings.ToLower(strings.TrimSpace(string(role)))
		if normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}
