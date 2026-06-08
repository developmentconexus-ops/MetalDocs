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
