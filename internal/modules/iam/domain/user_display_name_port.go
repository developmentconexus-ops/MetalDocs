package domain

import "context"

// UserDisplayNameReader is the narrow read surface cross-module consumers use to
// fetch metaldocs.iam_users.display_name without reaching across module boundaries.
// IAM owns metaldocs.iam_users; callers that need a display name depend on this
// port instead of issuing raw SQL (H-G class defect remediation, M4/F4.1).
//
// The implementation lives in iam/infrastructure/postgres and reads from the
// connection pool — never from a caller's lock-holding transaction (H-PRE-1).
type UserDisplayNameReader interface {
	// DisplayName returns iam_users.display_name for (tenantID, userID).
	// Returns ("", nil) when the user is absent in that tenant — best-effort
	// snapshot that matches all current single-read call sites.
	DisplayName(ctx context.Context, tenantID, userID string) (string, error)

	// DisplayNames returns userID -> display_name for the given users that are
	// present in the tenant AND have a non-empty display_name. Users absent or
	// whose display_name is null/empty are omitted; the caller applies its own
	// fallback (e.g. mapping missing id -> the id itself), which reproduces the
	// COALESCE(NULLIF(display_name,''), user_id) behavior consumer-side.
	DisplayNames(ctx context.Context, tenantID string, userIDs []string) (map[string]string, error)
}

// NoopUserDisplayNameReader is the explicit null-object for callers that want no
// display-name resolution (tests, or paths where names are irrelevant). It
// satisfies UserDisplayNameReader and always returns empty results with nil errors.
// Use it in constructor injection sites where display-name resolution is not needed,
// instead of wiring the real pool-backed repository.
type NoopUserDisplayNameReader struct{}

// DisplayName always returns ("", nil).
func (NoopUserDisplayNameReader) DisplayName(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

// DisplayNames always returns an empty, non-nil map and a nil error.
func (NoopUserDisplayNameReader) DisplayNames(_ context.Context, _ string, _ []string) (map[string]string, error) {
	return map[string]string{}, nil
}
